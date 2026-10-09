package service

import (
	"sop/internal/database"
	"sop/internal/model"
)

// GetAllEquipments 查询全部设备机型（无分页），数据源为 SPAREPART 库 equipment 表（只读）
func GetAllEquipments() ([]model.Equipment, error) {
	var list []model.Equipment
	if err := database.Sparepart().Order("ID ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	if list == nil {
		list = []model.Equipment{}
	}
	return list, nil
}
