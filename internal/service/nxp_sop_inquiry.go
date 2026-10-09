package service

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"sop/internal/database"
	"sop/internal/model"
	bizerr "sop/pkg/errors"
)

// 提问状态枚举（与 api.md、前端约定一致）
const (
	InquiryStatusPending  = "pending"  // 待回复
	InquiryStatusResolved = "resolved" // 已处理
)

// CreateInquiry 创建提问
func CreateInquiry(inquiry *model.NxpSopInquiry) error {
	if inquiry.Status == "" {
		inquiry.Status = InquiryStatusPending
	}
	// 校验执行说明选项真实存在，避免写入无效外键
	if inquiry.ExecutionOptionID != 0 {
		if _, err := GetExecutionOptionByID(inquiry.ExecutionOptionID); err != nil {
			return err
		}
	}
	return database.SOPDB().Create(inquiry).Error
}

// IncrementStepRepairCount 将指定步骤的历史维修次数 +1（现场提问时统计）
func IncrementStepRepairCount(stepID uint) error {
	return database.SOPDB().Model(&model.NxpSopGuideStep{}).
		Where("id = ?", stepID).
		UpdateColumn("history_repair_count", gorm.Expr("history_repair_count + 1")).
		Error
}

// GetInquiryList 分页查询提问列表，可按故障码、状态筛选；均空时分页返回全部。
func GetInquiryList(page, pageSize int, faultCode, status string) ([]model.NxpSopInquiry, int64, error) {
	db := database.SOPDB().Model(&model.NxpSopInquiry{})

	if faultCode != "" {
		like := "%" + faultCode + "%"
		db = db.Where("fault_code LIKE ?", like)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []model.NxpSopInquiry
	if err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	if list == nil {
		list = []model.NxpSopInquiry{}
	}
	return list, total, nil
}

// GetNxpSopInquiryByIds 按 ID 列表批量查询提问，结果按 id 升序返回。
// SQL Server 单条语句参数上限约 2100，按 1000 个 ID 一批分片查询，避免超限报错
// （与 GetFacilitiesByIDs / GetFaultCategoryByIDs 分片约定一致）。
func GetNxpSopInquiryByIds(ids []uint) ([]model.NxpSopInquiry, error) {
	if len(ids) == 0 {
		return []model.NxpSopInquiry{}, nil
	}
	const chunkSize = 1000
	result := make([]model.NxpSopInquiry, 0, len(ids))
	for start := 0; start < len(ids); start += chunkSize {
		end := start + chunkSize
		if end > len(ids) {
			end = len(ids)
		}
		var chunk []model.NxpSopInquiry
		if err := database.SOPDB().Where("id IN ?", ids[start:end]).Order("id ASC").Find(&chunk).Error; err != nil {
			return nil, err
		}
		result = append(result, chunk...)
	}
	return result, nil
}

// UpdateInquiryReply 管理员回复提问：更新回复内容、回复人与回复时间，并可按需变更状态。
// 状态仅接受 pending / resolved，非法值返回 400；answered_at 仅在真正更新
// 回复内容或回复人时刷新，避免「只改状态」误污染最后回复时间。
func UpdateInquiryReply(id uint, answer string, status string, repliedBy string) error {
	var inquiry model.NxpSopInquiry
	if err := database.SOPDB().First(&inquiry, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return bizerr.ErrInquiryNotFound
		}
		return err
	}

	updates := map[string]interface{}{}
	if answer != "" {
		updates["answer"] = answer
	}
	if repliedBy != "" {
		updates["answered_by"] = repliedBy
	}
	if status != "" {
		if status != InquiryStatusPending && status != InquiryStatusResolved {
			return bizerr.NewParam("status 仅支持 pending / resolved")
		}
		if status != inquiry.Status {
			updates["status"] = status
		}
	}
	// 仅本次真正写入了回复内容或回复人才刷新回复时间；
	// 单纯变更状态（标记已处理）不应覆盖 answered_at。
	if answer != "" || repliedBy != "" {
		updates["answered_at"] = timeNow()
	}

	return database.SOPDB().Model(&model.NxpSopInquiry{}).Where("id = ?", id).Updates(updates).Error
}

// timeNow 返回当前时间，供回复时间写库
func timeNow() *time.Time {
	now := time.Now()
	return &now
}
