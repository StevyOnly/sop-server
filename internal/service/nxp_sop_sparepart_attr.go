package service

import (
	"errors"

	"gorm.io/gorm"

	"sop/internal/database"
	"sop/internal/model"
	bizerr "sop/pkg/errors"
)

// SparepartAttrListItem 备件属性列表项：attr 字段 + 老库备件关键信息（跨库内存拼装）
type SparepartAttrListItem struct {
	model.NxpSopSparepartAttr
	SBrand       string `json:"sBrand"`
	SModel       string `json:"sModel"`
	SDescription string `json:"sDescription"`
}

// GetSparepartAttrList 分页查询备件属性列表（以 SOPHub 库 attr 表为主体）。
// sparepart 在 SPAREPART 只读老库，跨库无法 JOIN：keyword 非空时先在老库
// 按 s_id/品牌/型号/描述模糊匹配出 s_id 集合过滤 attr；每页再按 s_id 批量
// 回查老库，在内存中拼装备件信息（同 GetSparepartListWithAttr 模式）。
func GetSparepartAttrList(page, pageSize int, keyword string) ([]SparepartAttrListItem, int64, error) {
	db := database.SOPDB().Model(&model.NxpSopSparepartAttr{})

	if keyword != "" {
		like := "%" + keyword + "%"
		var sids []string
		if err := database.Sparepart().Model(&model.Sparepart{}).
			Where("s_id LIKE ? OR s_brand LIKE ? OR s_model LIKE ? OR s_description LIKE ?",
				like, like, like, like).
			Pluck("s_id", &sids).Error; err != nil {
			return nil, 0, err
		}
		if len(sids) == 0 {
			return []SparepartAttrListItem{}, 0, nil
		}
		db = db.Where("s_id IN ?", sids)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []model.NxpSopSparepartAttr
	if err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	result := make([]SparepartAttrListItem, 0, len(list))
	if len(list) == 0 {
		return result, total, nil
	}

	ids := make([]string, 0, len(list))
	for _, attr := range list {
		ids = append(ids, attr.SID)
	}

	var sps []model.Sparepart
	if err := database.Sparepart().Where("s_id IN ?", ids).Find(&sps).Error; err != nil {
		return nil, 0, err
	}
	spBySID := make(map[string]model.Sparepart, len(sps))
	for _, sp := range sps {
		spBySID[sp.SID] = sp
	}

	for _, attr := range list {
		item := SparepartAttrListItem{NxpSopSparepartAttr: attr}
		if sp, ok := spBySID[attr.SID]; ok {
			item.SBrand = sp.SBrand
			item.SModel = sp.SModel
			item.SDescription = sp.SDescription
		}
		result = append(result, item)
	}
	return result, total, nil
}

// GetSparepartAttrBySID 按备件 s_id 查询属性；未配置返回 ErrSparepartAttrNotFound
func GetSparepartAttrBySID(sid string) (*model.NxpSopSparepartAttr, error) {
	var attr model.NxpSopSparepartAttr
	if err := database.SOPDB().Where("s_id = ?", sid).First(&attr).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizerr.ErrSparepartAttrNotFound
		}
		return nil, err
	}
	return &attr, nil
}

// validateSparepartAttrValues 校验属性取值：标准周期为正整数，安全系数 0~1.5
func validateSparepartAttrValues(standardCycle int, safetyFactor float64) error {
	if standardCycle <= 0 {
		return bizerr.NewParam("标准更换周期必须为正整数")
	}
	if safetyFactor <= 0 || safetyFactor > 1.5 {
		return bizerr.NewParam("安全系数须在 0~1.5 之间")
	}
	return nil
}

// CreateSparepartAttr 新增备件属性：校验取值合法、备件存在于老库、s_id 未被配置过
func CreateSparepartAttr(attr *model.NxpSopSparepartAttr) error {
	if attr.SID == "" {
		return bizerr.NewParam("缺少备件 ID")
	}
	if err := validateSparepartAttrValues(attr.StandardCycle, attr.SafetyFactor); err != nil {
		return err
	}

	var spCount int64
	if err := database.Sparepart().Model(&model.Sparepart{}).
		Where("s_id = ?", attr.SID).Count(&spCount).Error; err != nil {
		return err
	}
	if spCount == 0 {
		return bizerr.ErrSparepartNotFound
	}

	var attrCount int64
	if err := database.SOPDB().Model(&model.NxpSopSparepartAttr{}).
		Where("s_id = ?", attr.SID).Count(&attrCount).Error; err != nil {
		return err
	}
	if attrCount > 0 {
		return bizerr.ErrSparepartAttrExists
	}

	return database.SOPDB().Create(attr).Error
}

// UpdateSparepartAttr 更新备件属性：sId 创建后不可变，仅更新周期与安全系数两列
func UpdateSparepartAttr(id uint, standardCycle int, safetyFactor float64) error {
	if id == 0 {
		return bizerr.NewParam("缺少属性记录 ID")
	}
	if err := validateSparepartAttrValues(standardCycle, safetyFactor); err != nil {
		return err
	}

	var existing model.NxpSopSparepartAttr
	if err := database.SOPDB().First(&existing, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return bizerr.ErrSparepartAttrNotFound
		}
		return err
	}

	return database.SOPDB().Model(&model.NxpSopSparepartAttr{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"standard_cycle": standardCycle,
			"safety_factor":  safetyFactor,
		}).Error
}

// DeleteSparepartAttr 删除备件属性（仅移除本库属性配置，老库备件不受影响）
func DeleteSparepartAttr(id uint) error {
	result := database.SOPDB().Delete(&model.NxpSopSparepartAttr{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return bizerr.ErrSparepartAttrNotFound
	}
	return nil
}
