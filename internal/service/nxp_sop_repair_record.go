package service

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"sop/internal/database"
	"sop/internal/model"
	bizerr "sop/pkg/errors"
)

// SaveRepairRecord 保存/完成维修记录（创建新记录或更新已有记录的结束信息）
func SaveRepairRecord(rec *model.NxpSopRepairRecord) error {
	if rec.Status == "" {
		rec.Status = "completed"
	}
	now := time.Now()
	if rec.RepairTime == nil {
		lt := model.LocalTime(now)
		rec.RepairTime = &lt
	}
	return database.SOPDB().Create(rec).Error
}

// GetRepairRecordList 分页查询维修记录，支持按设施/故障码/工程师/状态筛选
func GetRepairRecordList(page, pageSize int, facilityID uint, faultCode string, engineerID uint, status string) ([]model.NxpSopRepairRecord, int64, error) {
	db := database.SOPDB().Model(&model.NxpSopRepairRecord{})

	if facilityID != 0 {
		db = db.Where("facility_id = ?", facilityID)
	}
	if faultCode != "" {
		db = db.Where("fault_code = ?", faultCode)
	}
	if engineerID != 0 {
		db = db.Where("engineer_id = ?", engineerID)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []model.NxpSopRepairRecord
	if err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	if list == nil {
		list = []model.NxpSopRepairRecord{}
	}
	return list, total, nil
}

// GetRepairRecordByID 根据 ID 查询维修记录
func GetRepairRecordByID(id uint) (*model.NxpSopRepairRecord, error) {
	var rec model.NxpSopRepairRecord
	if err := database.SOPDB().First(&rec, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizerr.ErrRepairRecordNotFound
		}
		return nil, err
	}
	return &rec, nil
}

// GetNxpRepairRecordByIds 按 ID 列表批量查询维修记录，结果按 id 升序返回。
// SQL Server 单条语句参数上限约 2100，按 1000 个 ID 一批分片查询，避免超限报错
// （与 GetFacilitiesByIDs / GetNxpSopInquiryByIds 分片约定一致）。
func GetNxpRepairRecordByIds(ids []uint) ([]model.NxpSopRepairRecord, error) {
	if len(ids) == 0 {
		return []model.NxpSopRepairRecord{}, nil
	}
	const chunkSize = 1000
	result := make([]model.NxpSopRepairRecord, 0, len(ids))
	for start := 0; start < len(ids); start += chunkSize {
		end := start + chunkSize
		if end > len(ids) {
			end = len(ids)
		}
		var chunk []model.NxpSopRepairRecord
		if err := database.SOPDB().Where("id IN ?", ids[start:end]).Order("id ASC").Find(&chunk).Error; err != nil {
			return nil, err
		}
		result = append(result, chunk...)
	}
	return result, nil
}
