package model

// NxpSopSparepartAttr 备件 SOP 属性表，对应 SOPHub 库 nxp_sop_sparepart_attr
// 备件 SOP 属性表，记录备件的 SOP 属性，如标准周期、安全因子等。
// SID 为跨库逻辑外键，指向 SPAREPART 库 sparepart.s_id（老库只读，属性在本库维护）。
type NxpSopSparepartAttr struct {
	BaseModel
	SID           string  `gorm:"column:s_id" json:"sId"`
	StandardCycle int     `gorm:"column:standard_cycle" json:"standardCycle"`
	SafetyFactor  float64 `gorm:"column:safety_factor" json:"safetyFactor"`
}

// TableName 返回 nxp_sop_sparepart_attr 表名
func (NxpSopSparepartAttr) TableName() string {
	return "nxp_sop_sparepart_attr"
}
