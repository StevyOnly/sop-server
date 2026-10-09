package service

import (
	"sop/internal/database"
	"sop/internal/model"
)

// GetInspectionContentList 分页查询维修内容列表（Onebe 老库只读）。
// 支持按维修位置 ID 精确过滤、按内容模糊搜索；
// processID 为三态：nil（不传）→ 不过滤；*0 → 查 process_id 为空的记录；*>0 → 按 process_id 精确过滤。
// 按 position_id 排序便于按位置归类。
func GetInspectionContentList(page, pageSize int, positionID int, keyword string, processID *int) ([]model.NxpInspectionContent, int64, error) {
	db := database.Onebe().Model(&model.NxpInspectionContent{})
	if positionID > 0 {
		db = db.Where("position_id = ?", positionID)
	}
	if keyword != "" {
		db = db.Where("content LIKE ?", "%"+keyword+"%")
	}
	if processID != nil {
		if *processID == 0 {
			// 传 0 仅查未绑定工序的记录
			db = db.Where("process_id IS NULL")
		} else {
			db = db.Where("process_id = ?", *processID)
		}
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []model.NxpInspectionContent
	if err := db.Order("position_id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	if list == nil {
		list = []model.NxpInspectionContent{}
	}
	return list, total, nil
}
