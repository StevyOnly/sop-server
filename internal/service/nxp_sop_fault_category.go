package service

import (
	"errors"

	"gorm.io/gorm"

	"sop/internal/database"
	"sop/internal/model"
	bizerr "sop/pkg/errors"
)

// GetFaultCategoryList 分页查询故障分类，支持按分类名称模糊搜索
func GetFaultCategoryList(page, pageSize int, keyword string) ([]model.NxpSopFaultCategory, int64, error) {
	db := database.SOPDB().Model(&model.NxpSopFaultCategory{})
	if keyword != "" {
		db = db.Where("label LIKE ?", "%"+keyword+"%")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []model.NxpSopFaultCategory
	if err := db.Order("id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	if list == nil {
		list = []model.NxpSopFaultCategory{}
	}
	return list, total, nil
}

// GetFaultCategoryByIDs 按 ID 列表批量查询故障分类，结果按 id 升序返回
// SQL Server 单条语句参数上限约 2100，按 1000 个 ID 一批分片查询，避免超限报错
func GetFaultCategoryByIDs(ids []uint) ([]model.NxpSopFaultCategory, error) {
	if len(ids) == 0 {
		return []model.NxpSopFaultCategory{}, nil
	}
	const chunkSize = 1000
	result := make([]model.NxpSopFaultCategory, 0, len(ids))
	for start := 0; start < len(ids); start += chunkSize {
		end := start + chunkSize
		if end > len(ids) {
			end = len(ids)
		}
		var chunk []model.NxpSopFaultCategory
		if err := database.SOPDB().Where("id IN ?", ids[start:end]).Order("id ASC").Find(&chunk).Error; err != nil {
			return nil, err
		}
		result = append(result, chunk...)
	}
	return result, nil
}

// FaultCategoryFault 分类下的报警明细（来自启用指南）
type FaultCategoryFault struct {
	FaultCode            string `json:"faultCode"`
	FaultCodeDescription string `json:"faultCodeDescription"`
	GuideID              uint   `json:"guideId"`
}

// FaultCategoryGroup 故障分类分组（含该机型下的报警明细）
type FaultCategoryGroup struct {
	FaultCategoryID uint                 `json:"faultCategoryId"`
	Label           string               `json:"label"`
	Faults          []FaultCategoryFault `json:"faults"`
}

// GetFaultCategoriesByModelId 按机型 ID 分页查询关联故障分类及报警明细。
// 机型与分类通过 nxp_sop_guide 间接关联（guide 同时持有 machine_model_id 与 fault_category_id）：
// 仅取启用（enabled = 1）且 fault_category_id > 0 的指南；分页粒度为「分类组」，按 fault_category_id ASC。
func GetFaultCategoriesByModelId(page, pageSize int, machineModelID uint) ([]FaultCategoryGroup, int64, error) {
	// 每次链式调用前重新构造条件，避免 GORM 复用 session 导致条件污染
	newGuideQuery := func() *gorm.DB {
		return database.SOPDB().Model(&model.NxpSopGuide{}).
			Where("machine_model_id = ? AND enabled = 1 AND fault_category_id > 0", machineModelID)
	}

	var total int64
	if err := newGuideQuery().Distinct("fault_category_id").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var catIDs []uint
	if err := newGuideQuery().Order("fault_category_id ASC").
		Offset((page-1)*pageSize).Limit(pageSize).
		Pluck("DISTINCT fault_category_id", &catIDs).Error; err != nil {
		return nil, 0, err
	}
	if len(catIDs) == 0 {
		return []FaultCategoryGroup{}, total, nil
	}

	// 本页分类的文案
	var cats []model.NxpSopFaultCategory
	if err := database.SOPDB().Where("id IN ?", catIDs).Find(&cats).Error; err != nil {
		return nil, 0, err
	}
	labelMap := make(map[uint]string, len(cats))
	for _, c := range cats {
		labelMap[c.ID] = c.Label
	}

	// 本页分类在该机型下的报警明细
	type guideFaultRow struct {
		FaultCategoryID      uint
		ID                   uint
		FaultCode            string
		FaultCodeDescription string
	}
	var rows []guideFaultRow
	if err := database.SOPDB().Model(&model.NxpSopGuide{}).
		Select("fault_category_id", "id", "fault_code", "fault_code_description").
		Where("machine_model_id = ? AND enabled = 1 AND fault_category_id IN ?", machineModelID, catIDs).
		Order("fault_category_id ASC, id ASC").
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	faultMap := make(map[uint][]FaultCategoryFault, len(catIDs))
	for _, r := range rows {
		faultMap[r.FaultCategoryID] = append(faultMap[r.FaultCategoryID], FaultCategoryFault{
			FaultCode:            r.FaultCode,
			FaultCodeDescription: r.FaultCodeDescription,
			GuideID:              r.ID,
		})
	}

	// 按分页取到的 catIDs 顺序组装，保证组顺序与分页一致
	groups := make([]FaultCategoryGroup, 0, len(catIDs))
	for _, id := range catIDs {
		fs := faultMap[id]
		if fs == nil {
			fs = []FaultCategoryFault{}
		}
		groups = append(groups, FaultCategoryGroup{FaultCategoryID: id, Label: labelMap[id], Faults: fs})
	}
	return groups, total, nil
}

// CreateFaultCategory 新增故障分类，校验名称非空且唯一
func CreateFaultCategory(cat *model.NxpSopFaultCategory) error {
	if cat.Label == "" {
		return bizerr.NewParam("分类名称不能为空")
	}

	var count int64
	if err := database.SOPDB().Model(&model.NxpSopFaultCategory{}).
		Where("label = ?", cat.Label).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return bizerr.NewConflict("分类名称已存在")
	}

	return database.SOPDB().Create(cat).Error
}

// UpdateFaultCategory 更新故障分类；校验记录存在且名称唯一（排除自身）
func UpdateFaultCategory(cat *model.NxpSopFaultCategory) error {
	if cat.ID == 0 {
		return bizerr.NewParam("缺少分类 ID")
	}
	if cat.Label == "" {
		return bizerr.NewParam("分类名称不能为空")
	}

	var existing model.NxpSopFaultCategory
	if err := database.SOPDB().First(&existing, cat.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return bizerr.ErrFaultCategoryNotFound
		}
		return err
	}

	var count int64
	if err := database.SOPDB().Model(&model.NxpSopFaultCategory{}).
		Where("label = ? AND id <> ?", cat.Label, cat.ID).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return bizerr.NewConflict("分类名称已存在")
	}

	return database.SOPDB().Model(&model.NxpSopFaultCategory{}).
		Where("id = ?", cat.ID).
		Update("label", cat.Label).Error
}

// DeleteFaultCategory 删除故障分类
func DeleteFaultCategory(id uint) error {
	result := database.SOPDB().Delete(&model.NxpSopFaultCategory{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return bizerr.ErrFaultCategoryNotFound
	}
	return nil
}
