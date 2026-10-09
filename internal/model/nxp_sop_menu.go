package model

// NxpSopMenu 菜单表
type NxpSopMenu struct {
	BaseModel
	MenuName string `gorm:"column:menu_name" json:"menuName"`
}

// TableName 返回 nxp_sop_menu 表名
func (NxpSopMenu) TableName() string { return "nxp_sop_menu" }

// NxpSopMenuNames 预置菜单名称
var NxpSopMenuNames = []string{
	"统计看板",
	"标准 SOP 库",
	"现场提问记录",
	"预防性维护管理",
	"SOP 库的使用记录",
	"故障分类管理",
	"执行说明选项管理",
	"多媒体资料库",
	"用户权限管理",
}
