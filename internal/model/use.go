package model

// Use 备件-机型使用关系表（SPAREPART 库 dbo.[use]，只读）。
// 通过 s_id 与 sparepart 关联（一个备件可被多台机型共用，多对多），
// e_model 直接存机型代号。老库只读表，无审计字段，按库表实际结构定义。
type Use struct {
	ID     int    `gorm:"column:ID;primaryKey" json:"id"`
	SID    string `gorm:"column:s_id" json:"sId"`
	EModel string `gorm:"column:e_model" json:"eModel"`
}

// TableName 返回 use 表名
func (Use) TableName() string { return "use" }
