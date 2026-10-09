package model

// NxpSopExecutionOption 执行说明选项表 nxp_sop_execution_option
// 维护 PDA「执行与反馈记录」界面「操作记录 / 执行说明」下拉菜单的候选选项，
// 后台端通过「执行说明选项管理」模块增删改查，前端字段与 ExecutionOption 对齐。
type NxpSopExecutionOption struct {
	BaseModel
	Label     string `gorm:"column:label" json:"label"`         // 选项文案（下拉展示，如"已完成维修，设备恢复正常"）
	SortOrder int    `gorm:"column:sort_order" json:"sortOrder"` // 排序值，越小越靠前（PDA 下拉按此升序展示）
	Active    bool   `gorm:"column:active" json:"active"`        // 是否启用（停用后 PDA 下拉不再展示）
}

// TableName 返回 nxp_sop_execution_option 表名
func (NxpSopExecutionOption) TableName() string { return "nxp_sop_execution_option" }
