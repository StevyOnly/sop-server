package model

// NxpSopFaultCategory 故障分类表 nxp_sop_fault_category
// 维护 SOP「故障分类」与 PDA「报警内容」下拉菜单的候选分类，
// 后台端通过「故障分类管理」模块增删改查。
type NxpSopFaultCategory struct {
	BaseModel
	Label string `gorm:"column:label" json:"label"` // 分类文案（下拉展示，如"传感器污染"）
}

// TableName 返回 nxp_sop_fault_category 表名
func (NxpSopFaultCategory) TableName() string { return "nxp_sop_fault_category" }
