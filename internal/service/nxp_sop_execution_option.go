package service

import (
	"errors"

	"gorm.io/gorm"

	"sop/internal/database"
	"sop/internal/model"
	bizerr "sop/pkg/errors"
)

// GetExecutionOptionByID 按 ID 查询执行说明选项；不存在返回 ErrExecutionOptionNotFound
func GetExecutionOptionByID(id uint) (*model.NxpSopExecutionOption, error) {
	var opt model.NxpSopExecutionOption
	if err := database.SOPDB().First(&opt, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizerr.ErrExecutionOptionNotFound
		}
		return nil, err
	}
	return &opt, nil
}

// GetExecutionOptionList 分页查询执行说明选项，支持按文案模糊搜索
func GetExecutionOptionList(page, pageSize int, keyword string) ([]model.NxpSopExecutionOption, int64, error) {
	db := database.SOPDB().Model(&model.NxpSopExecutionOption{})
	if keyword != "" {
		db = db.Where("label LIKE ?", "%"+keyword+"%")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []model.NxpSopExecutionOption
	if err := db.Order("sort_order ASC, id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	if list == nil {
		list = []model.NxpSopExecutionOption{}
	}
	return list, total, nil
}

// GetActiveExecutionOptions 返回启用中的选项（供 PDA「执行说明」下拉，按 sortOrder 升序）
func GetActiveExecutionOptions() ([]model.NxpSopExecutionOption, error) {
	var list []model.NxpSopExecutionOption
	if err := database.SOPDB().Model(&model.NxpSopExecutionOption{}).
		Where("active = ?", true).
		Order("sort_order ASC, id ASC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	if list == nil {
		list = []model.NxpSopExecutionOption{}
	}
	return list, nil
}

// CreateExecutionOption 新增执行说明选项，校验文案非空且唯一
func CreateExecutionOption(opt *model.NxpSopExecutionOption) error {
	if opt.Label == "" {
		return bizerr.NewParam("选项文案不能为空")
	}

	var count int64
	if err := database.SOPDB().Model(&model.NxpSopExecutionOption{}).
		Where("label = ?", opt.Label).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return bizerr.NewConflict("选项文案已存在")
	}

	return database.SOPDB().Create(opt).Error
}

// UpdateExecutionOption 更新执行说明选项；校验记录存在且文案唯一（排除自身）
func UpdateExecutionOption(opt *model.NxpSopExecutionOption) error {
	if opt.ID == 0 {
		return bizerr.NewParam("缺少选项 ID")
	}
	if opt.Label == "" {
		return bizerr.NewParam("选项文案不能为空")
	}

	var existing model.NxpSopExecutionOption
	if err := database.SOPDB().First(&existing, opt.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return bizerr.ErrExecutionOptionNotFound
		}
		return err
	}

	if existing.Label != opt.Label {
		// 已被提问引用的选项禁止修改文案，避免历史记录文案悬空
		if err := checkExecutionOptionInUse(opt.ID); err != nil {
			return err
		}
	}

	var count int64
	if err := database.SOPDB().Model(&model.NxpSopExecutionOption{}).
		Where("label = ? AND id <> ?", opt.Label, opt.ID).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return bizerr.NewConflict("选项文案已存在")
	}

	return database.SOPDB().Model(&model.NxpSopExecutionOption{}).
		Where("id = ?", opt.ID).
		Updates(map[string]interface{}{
			"label":      opt.Label,
			"sort_order": opt.SortOrder,
			"active":     opt.Active,
		}).Error
}

// DeleteExecutionOption 删除执行说明选项
func DeleteExecutionOption(id uint) error {
	if err := checkExecutionOptionInUse(id); err != nil {
		return err
	}
	result := database.SOPDB().Delete(&model.NxpSopExecutionOption{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return bizerr.ErrExecutionOptionNotFound
	}
	return nil
}

// checkExecutionOptionInUse 校验执行说明选项是否已被提问引用（nxp_sop_inquiry.execution_option_id）
func checkExecutionOptionInUse(id uint) error {
	var count int64
	if err := database.SOPDB().Model(&model.NxpSopInquiry{}).
		Where("execution_option_id = ?", id).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return bizerr.ErrExecutionOptionInUse
	}
	return nil
}
