package service

import (
	"strconv"

	"sop/internal/database"
	"sop/internal/model"
)

// GetInspectionPositionList 分页查询维修位置列表（Onebe 老库只读），支持按位置名称模糊搜索
// 及按机型过滤：model_id 为逗号分隔的机型编号（如 "44,45,60"）；
// modelId>0 时仅返回 model_id 为空（通用位置，所有机型适用）或包含该机型编号的位置。
// processId 为三态：nil（不传）→ 不过滤；*0 → 查 process_id 为空的记录；*>0 → 按 process_id 精确过滤。
func GetInspectionPositionList(page, pageSize int, keyword string, modelId int, processId *int) ([]model.NxpInspectionPosition, int64, error) {
	db := database.Onebe().Model(&model.NxpInspectionPosition{})
	if keyword != "" {
		db = db.Where("position LIKE ?", "%"+keyword+"%")
	}
	if modelId > 0 {
		idStr := strconv.Itoa(modelId)
		// model_id 为空 → 通用位置（始终匹配）；否则逗号分段精确匹配该机型编号
		db = db.Where(
			"(model_id IS NULL OR model_id = '' OR model_id = ? OR "+
				"model_id LIKE ? OR model_id LIKE ? OR model_id LIKE ?)",
			idStr, idStr+",%", "%,"+idStr+",%", "%,"+idStr,
		)
	}
	if processId != nil {
		if *processId == 0 {
			// 传 0 仅查未绑定工序的记录
			db = db.Where("process_id IS NULL")
		} else {
			db = db.Where("process_id = ?", *processId)
		}
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []model.NxpInspectionPosition
	if err := db.Order("id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	if list == nil {
		list = []model.NxpInspectionPosition{}
	}
	return list, total, nil
}
