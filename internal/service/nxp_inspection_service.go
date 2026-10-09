package service

import (
	"sop/internal/database"
	"sop/internal/model"
	"time"
)

// SparepartReplaceItem 备件更换记录响应结构体。
// 为前端定制的 DTO，字段与 model 对齐但独立于库模型，避免直接暴露 model。
type SparepartReplaceItem struct {
	ID                  int                      `json:"id"`
	InspectionServiceID int                      `json:"inspectionServiceId"`
	SID                 string                   `json:"sId"`
	EID                 string                   `json:"eId"`
	OAmount             int                      `json:"oAmount"`
	InspectionService   *InspectionServiceDetail `json:"inspectionService"` // 维修服务详情（含反查用户名），查不到为 nil
	Sparepart           *SparepartDetail         `json:"sparepart"`         // 备件字典（跨库），查不到为 nil
}

// InspectionServiceDetail 维修服务详情：嵌入 NxpInspectionService 全字段，
// 并附加由 user_info_id 反查 nxp_user_info 得到的维修人姓名（chinese_name）。
type InspectionServiceDetail struct {
	model.NxpInspectionService
	UserName string `json:"userName"` // 维修人姓名（nxp_user_info.chinese_name）
}

// SparepartDetail 备件字典详情：嵌入 Sparepart 全字段（SPAREPART 库）。
type SparepartDetail struct {
	model.Sparepart
}

// GetSparepartReplaceList 分页查询备件更换记录。
// 数据源为 Onebe 老库 nxp_inspection_service_sparepart 表（只读）。
// inspectionServiceID>0 时精确匹配维修ID；sID / eID 非空时按模糊匹配；
// sModel 非空时先在 SPAREPART 库按 sparepart.s_model 预查 s_id 集合再过滤（跨库无法 JOIN）；
// startDate / endDate 非空时按维修时间 service_time 范围过滤（仅年月日，endDate 含当天）。
//
// 无外键约束 + 跨库（sparepart 在 SPAREPART 库）无法 JOIN，故采用：
// 分页查主表 → 批量查维修服务(nxp_inspection_service) → 批量反查用户(nxp_user_info)
// → 批量查备件字典(sparepart) → 内存拼装。任一关联查不到，对应子对象置 nil。
func GetSparepartReplaceList(page, pageSize int, inspectionServiceID int, sID, eID, sModel, startDate, endDate string) ([]SparepartReplaceItem, int64, error) {
	db := database.Onebe().Model(&model.NxpInspectionServiceSparepart{}).
		Joins("JOIN nxp_inspection_service s ON s.id = nxp_inspection_service_sparepart.inspection_service_id")
	if inspectionServiceID > 0 {
		db = db.Where("nxp_inspection_service_sparepart.inspection_service_id = ?", inspectionServiceID)
	}
	if sID != "" {
		db = db.Where("nxp_inspection_service_sparepart.s_id LIKE ?", "%"+sID+"%")
	}
	if eID != "" {
		db = db.Where("nxp_inspection_service_sparepart.e_id LIKE ?", "%"+eID+"%")
	}
	// 备件型号过滤（跨库预查）：先在 SPAREPART 库按 s_model 查出 s_id 集合，再 IN 过滤主查询
	if sModel != "" {
		modelIDs, err := querySparepartIDsByModel(sModel)
		if err != nil {
			return nil, 0, err
		}
		if len(modelIDs) == 0 {
			// 没有该型号的备件，直接短路返回空
			return []SparepartReplaceItem{}, 0, nil
		}
		db = db.Where("nxp_inspection_service_sparepart.s_id IN ?", modelIDs)
	}
	// 维修时间范围过滤：按开始/结束日期（含当天）
	start, hasStart := parseDateStart(startDate)
	end, hasEnd := parseDateEnd(endDate)
	if hasStart {
		db = db.Where("s.service_time >= ?", start)
	}
	if hasEnd {
		db = db.Where("s.service_time <= ?", end)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []model.NxpInspectionServiceSparepart
	if err := db.Order("nxp_inspection_service_sparepart.id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	if len(rows) == 0 {
		return []SparepartReplaceItem{}, total, nil
	}

	// 1. 收集关联键（去重）
	serviceIDs := make(map[int]struct{}, len(rows))
	sIDs := make(map[string]struct{}, len(rows))
	for _, r := range rows {
		if r.InspectionServiceID > 0 {
			serviceIDs[r.InspectionServiceID] = struct{}{}
		}
		if r.SID != "" {
			sIDs[r.SID] = struct{}{}
		}
	}

	// 2. 批量查维修服务（Onebe），收集维修人 id
	serviceMap := make(map[int]*model.NxpInspectionService, len(serviceIDs))
	var userInfoIDs []int
	if len(serviceIDs) > 0 {
		var services []model.NxpInspectionService
		if err := database.Onebe().Model(&model.NxpInspectionService{}).
			Where("id IN ?", keysOfIntSet(serviceIDs)).Find(&services).Error; err != nil {
			return nil, 0, err
		}
		for i := range services {
			sv := services[i]
			serviceMap[sv.ID] = &sv
			if sv.UserInfoID > 0 {
				userInfoIDs = append(userInfoIDs, sv.UserInfoID)
			}
		}
	}

	// 3. 批量反查用户名（nxp_user_info.id = user_info_id）
	userNameMap := make(map[int]string, len(userInfoIDs))
	if len(userInfoIDs) > 0 {
		var infos []model.NxpUserInfo
		if err := database.Onebe().Model(&model.NxpUserInfo{}).
			Where("id IN ?", userInfoIDs).Find(&infos).Error; err != nil {
			return nil, 0, err
		}
		for i := range infos {
			userNameMap[int(infos[i].ID)] = infos[i].ChineseName
		}
	}

	// 4. 批量查备件字典（SPAREPART 库，跨库）
	sparepartMap := make(map[string]*model.Sparepart, len(sIDs))
	if len(sIDs) > 0 {
		var spares []model.Sparepart
		if err := database.Sparepart().Model(&model.Sparepart{}).
			Where("s_id IN ?", keysOfStrSet(sIDs)).Find(&spares).Error; err != nil {
			return nil, 0, err
		}
		for i := range spares {
			sp := spares[i]
			sparepartMap[sp.SID] = &sp
		}
	}

	// 5. 内存拼装
	list := make([]SparepartReplaceItem, 0, len(rows))
	for _, r := range rows {
		item := SparepartReplaceItem{
			ID:                  r.ID,
			InspectionServiceID: r.InspectionServiceID,
			SID:                 r.SID,
			EID:                 r.EID,
			OAmount:             r.OAmount,
		}

		if sv, ok := serviceMap[r.InspectionServiceID]; ok {
			detail := &InspectionServiceDetail{NxpInspectionService: *sv}
			if name, ok := userNameMap[sv.UserInfoID]; ok {
				detail.UserName = name
			}
			item.InspectionService = detail
		}
		// 关联查不到的备件字典置 nil（默认）
		if sp, ok := sparepartMap[r.SID]; ok {
			item.Sparepart = &SparepartDetail{Sparepart: *sp}
		}

		list = append(list, item)
	}
	return list, total, nil
}

// keysOfIntSet 返回 int 集合的键切片
func keysOfIntSet(m map[int]struct{}) []int {
	out := make([]int, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// keysOfStrSet 返回 string 集合的键切片
func keysOfStrSet(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// querySparepartIDsByModel 在 SPAREPART 库按备件型号 s_model 模糊查询去重后的 s_id 集合（跨库预查）。
// 为防止 LIKE 全表扫 + IN 集合过大，限制最多返回 maxModelIDLimit 个。
const maxModelIDLimit = 5000

func querySparepartIDsByModel(keyword string) ([]string, error) {
	var ids []string
	err := database.Sparepart().Model(&model.Sparepart{}).
		Where("s_model LIKE ?", "%"+keyword+"%").
		Distinct("s_id").
		Limit(maxModelIDLimit).
		Pluck("s_id", &ids).Error
	return ids, err
}

const dateOnlyLayout = "2006-01-02"

// parseDateStart 解析开始日期（仅年月日），解析失败返回 has=false。
// 成功时返回该日 00:00:00。
func parseDateStart(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}
	t, err := time.ParseInLocation(dateOnlyLayout, s, time.Local)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

// parseDateEnd 解析结束日期（仅年月日），解析失败返回 has=false。
// 成功时返回该日 23:59:59（含当天）。
func parseDateEnd(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}
	t, err := time.ParseInLocation(dateOnlyLayout, s, time.Local)
	if err != nil {
		return time.Time{}, false
	}
	return t.Add(24*time.Hour - time.Second), true
}

// GetSparepartReplaceListExport 获取备件更换记录全量（Excel 导出用）。
// 直接复用列表的完整筛选逻辑，pageSize 放大取全量（导出不走 api 层 NormalizePage 上限）。
func GetSparepartReplaceListExport(inspectionServiceID int, sID, eID, sModel, startDate, endDate string) ([]SparepartReplaceItem, error) {
	list, _, err := GetSparepartReplaceList(1, 100000, inspectionServiceID, sID, eID, sModel, startDate, endDate)
	return list, err
}

// SparepartModelOption 备件型号下拉项：型号 + 示例 s_id（仅更换记录中用过的备件）。
type SparepartModelOption struct {
	SModel string `json:"sModel"` // 备件型号（sparepart.s_model，去重）
	SID    string `json:"sId"`    // 该型号的一个示例 s_id（去重后取最小）
}

// GetSparepartModelOptions 返回更换记录中实际用过的去重备件型号（含示例 s_id）。
// 数据源：Onebe 库 nxp_inspection_service_sparepart 去重 s_id → SPAREPART 库 sparepart 反查 s_model。
// 保持数据库原值，不做 TRIM 去空格处理。
func GetSparepartModelOptions() ([]SparepartModelOption, error) {
	// 第1步：Onebe 库查更换记录实际用过的 s_id（去重）
	var usedSIDs []string
	if err := database.Onebe().Model(&model.NxpInspectionServiceSparepart{}).
		Where("s_id IS NOT NULL AND s_id <> ''").
		Distinct("s_id").
		Pluck("s_id", &usedSIDs).Error; err != nil {
		return nil, err
	}
	if len(usedSIDs) == 0 {
		return []SparepartModelOption{}, nil
	}

	// 第2步：SPAREPART 库按 s_id 反查型号（型号去重，取最小 s_id 作示例）
	var rows []struct {
		SModel string
		SID    string
	}
	if err := database.Sparepart().Model(&model.Sparepart{}).
		Select("s_model, MIN(s_id) AS s_id").
		Where("s_id IN ? AND s_model <> ''", usedSIDs).
		Group("s_model").
		Order("s_model").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]SparepartModelOption, 0, len(rows))
	for _, r := range rows {
		if r.SModel == "" {
			continue
		}
		out = append(out, SparepartModelOption{SModel: r.SModel, SID: r.SID})
	}
	return out, nil
}
