package service

import (
	"sop/internal/database"
	"sop/internal/model"
)

// GetInspectionQuestionTypeList 分页查询问题类型列表（Onebe 老库只读），支持按名称模糊搜索；
// processId 为三态：nil（不传）→ 不过滤；*0 → 查 process_id 为空的记录；*>0 → 按 process_id 精确过滤。
func GetInspectionQuestionTypeList(page, pageSize int, keyword string, processId *int) ([]model.NxpInspectionQuestionType, int64, error) {
	db := database.Onebe().Model(&model.NxpInspectionQuestionType{})
	if keyword != "" {
		db = db.Where("name LIKE ?", "%"+keyword+"%")
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

	var list []model.NxpInspectionQuestionType
	if err := db.Order("id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	if list == nil {
		list = []model.NxpInspectionQuestionType{}
	}
	return list, total, nil
}
