package service

import (
	"sop/internal/database"
	"sop/internal/model"
)

// GetFacilityList 分页查询设备（设施）列表，数据源为 nxp_facility 表
func GetFacilityList(page, pageSize int, keyword string) ([]model.NxpFacility, int64, error) {
	db := database.Onebe().Model(&model.NxpFacility{})

	if keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("name LIKE ? OR number LIKE ? OR model LIKE ? OR serial_no LIKE ?",
			like, like, like, like)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var facilities []model.NxpFacility
	if err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&facilities).Error; err != nil {
		return nil, 0, err
	}
	if facilities == nil {
		facilities = []model.NxpFacility{}
	}
	return facilities, total, nil
}

// GetFacilitiesByIDs 按 ID 列表批量查询设施（用于提问/反馈等记录反查设备编号与 SN）
// SQL Server 单条语句参数上限约 2100，按 1000 个 ID 一批分片查询，避免超限报错
// （与 GetUsersByIDs 的分片约定一致，见 nxp_user.go）。
func GetFacilitiesByIDs(ids []uint) ([]model.NxpFacility, error) {
	if len(ids) == 0 {
		return []model.NxpFacility{}, nil
	}
	const chunkSize = 1000
	result := make([]model.NxpFacility, 0, len(ids))
	for start := 0; start < len(ids); start += chunkSize {
		end := start + chunkSize
		if end > len(ids) {
			end = len(ids)
		}
		var chunk []model.NxpFacility
		if err := database.Onebe().Where("id IN ?", ids[start:end]).Find(&chunk).Error; err != nil {
			return nil, err
		}
		result = append(result, chunk...)
	}
	if result == nil {
		result = []model.NxpFacility{}
	}
	return result, nil
}
