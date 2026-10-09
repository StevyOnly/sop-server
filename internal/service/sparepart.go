package service

import (
	"sop/internal/database"
	"sop/internal/model"
)

// GetSparepartList 分页查询备件列表，数据源为 SPAREPART 库 sparepart 表（只读）
func GetSparepartList(page, pageSize int, keyword string) ([]model.Sparepart, int64, error) {
	db := database.Sparepart().Model(&model.Sparepart{})

	if keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("s_brand LIKE ? OR s_model LIKE ? OR s_description LIKE ?",
			like, like, like)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []model.Sparepart
	if err := db.Order("s_id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	if list == nil {
		list = []model.Sparepart{}
	}
	return list, total, nil
}

// SparepartWithAttr 备件信息 + SOP 属性（standardCycle/safetyFactor 平铺在上层；
// 未配置属性时为 null）
type SparepartWithAttr struct {
	model.Sparepart
	StandardCycle *int     `json:"standardCycle"`
	SafetyFactor  *float64 `json:"safetyFactor"`
	EModel        []string `json:"eModel"`
}

// GetSparepartListWithAttr 分页查询备件列表并附加 SOP 属性。
// sparepart 在 SPAREPART 只读库、attr 在 SOPHub 主库，跨库无法 JOIN，
// 故先分页查备件，再按 s_id 批量查属性后在内存中拼装。
func GetSparepartListWithAttr(page, pageSize int, keyword string) ([]SparepartWithAttr, int64, error) {
	list, total, err := GetSparepartList(page, pageSize, keyword)
	if err != nil {
		return nil, 0, err
	}

	result := make([]SparepartWithAttr, 0, len(list))
	if len(list) == 0 {
		return result, total, nil
	}

	ids := make([]string, 0, len(list))
	for _, sp := range list {
		ids = append(ids, sp.SID)
	}

	var attrs []model.NxpSopSparepartAttr
	if err := database.SOPDB().Where("s_id IN ?", ids).Find(&attrs).Error; err != nil {
		return nil, 0, err
	}
	attrBySID := make(map[string]model.NxpSopSparepartAttr, len(attrs))
	for _, attr := range attrs {
		attrBySID[attr.SID] = attr
	}

	// 批量查询 use（同库内联，多对多：一个备件可挂多台机型），按 s_id 聚合机型列表并去重
	var uses []model.Use
	if err := database.Sparepart().Where("s_id IN ?", ids).Find(&uses).Error; err != nil {
		return nil, 0, err
	}
	eModelBySID := make(map[string][]string, len(list))
	seen := make(map[string]map[string]struct{}, len(list))
	for _, u := range uses {
		if seen[u.SID] == nil {
			seen[u.SID] = make(map[string]struct{})
		}
		if _, dup := seen[u.SID][u.EModel]; dup {
			continue
		}
		seen[u.SID][u.EModel] = struct{}{}
		eModelBySID[u.SID] = append(eModelBySID[u.SID], u.EModel)
	}

	for _, sp := range list {
		item := SparepartWithAttr{Sparepart: sp}
		if attr, ok := attrBySID[sp.SID]; ok {
			item.StandardCycle = &attr.StandardCycle
			item.SafetyFactor = &attr.SafetyFactor
		}
		if eModels, ok := eModelBySID[sp.SID]; ok {
			item.EModel = eModels
		}
		result = append(result, item)
	}
	return result, total, nil
}
