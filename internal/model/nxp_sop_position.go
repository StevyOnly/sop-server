package model

import (
	"errors"

	"gorm.io/gorm"
)

// NxpSopPosition 职务表
// 该表为预置只读表：应用层禁止对其做增/改/删，仅由初始化流程写入预置职务，
// 以保证 nxp_sop_position 恒为初始化时的样子（见 initialize.seedPresetData）。
// Menus 通过 many2many 连接表 nxp_sop_position_menu 关联可见菜单（多对多）。
type NxpSopPosition struct {
	BaseModel
	PositionName string       `gorm:"column:position_name" json:"positionName"`
	Menus        []NxpSopMenu `json:"menus" gorm:"many2many:nxp_sop_position_menu;joinForeignKey:PositionID;joinReferences:MenuID"`
}

// TableName 返回 nxp_sop_position 表名
func (NxpSopPosition) TableName() string { return "nxp_sop_position" }

// 只读校验错误：任何对职务表的非预置写入都应被拦截
var ErrPositionReadOnly = errors.New("nxp_sop_position 为预置只读表，禁止新增/修改/删除职务")

// BeforeCreate 阻断新增
func (NxpSopPosition) BeforeCreate(*gorm.DB) error { return ErrPositionReadOnly }

// BeforeUpdate 阻断修改
func (NxpSopPosition) BeforeUpdate(*gorm.DB) error { return ErrPositionReadOnly }

// BeforeDelete 阻断删除
func (NxpSopPosition) BeforeDelete(*gorm.DB) error { return ErrPositionReadOnly }

// NxpSopPositionNames 预置职务名称
var NxpSopPositionNames = []string{
	"操作员",
	"带班",
	"产品工程师",
	"设备工程师",
	"PM技术员",
	"PM leader",
	"生产 leader",
	"分部经理",
	"部门经理",
	"化学分析",
	"Inline QA",
	"化学工程师",
	"培训员",
	"文员",
	"工艺工程师",
}
