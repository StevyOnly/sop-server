package service

import (
	"sop/internal/database"
	"sop/internal/model"
)

// GetMachineModelList 分页查询机型列表，支持按机型名称模糊搜索
// 数据源为 Onebe 老库 nxp_machine_model 表（只读）。
func GetMachineModelList(page, pageSize int, keyword string) ([]model.NxpMachineModel, int64, error) {
	db := database.Onebe().Model(&model.NxpMachineModel{})
	if keyword != "" {
		db = db.Where("name LIKE ?", "%"+keyword+"%")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []model.NxpMachineModel
	if err := db.Order("id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	if list == nil {
		list = []model.NxpMachineModel{}
	}
	return list, total, nil
}

// GetMachineModelsByIDs 按 ID 列表批量获取机型，结果按 id 升序返回。
// 数据源为 Onebe 老库 nxp_machine_model 表（只读）。
func GetMachineModelsByIDs(ids []uint) ([]model.NxpMachineModel, error) {
	if len(ids) == 0 {
		return []model.NxpMachineModel{}, nil
	}
	const chunkSize = 1000
	result := make([]model.NxpMachineModel, 0, len(ids))
	for start := 0; start < len(ids); start += chunkSize {
		end := start + chunkSize
		if end > len(ids) {
			end = len(ids)
		}
		var chunk []model.NxpMachineModel
		if err := database.Onebe().Where("id IN ?", ids[start:end]).Order("id ASC").Find(&chunk).Error; err != nil {
			return nil, err
		}
		result = append(result, chunk...)
	}
	if result == nil {
		result = []model.NxpMachineModel{}
	}
	return result, nil
}
