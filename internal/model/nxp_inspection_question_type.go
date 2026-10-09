package model

// NxpInspectionQuestionType 问题类型表，对应 nxp_inspection_question_type（Onebe 老库只读）
// 注意：该表 create_time 为 datetime2 类型，按库表实际结构定义。
type NxpInspectionQuestionType struct {
	ID         int       `gorm:"column:id;primaryKey;comment:主键" json:"id"`
	Name       string    `gorm:"column:name;comment:问题类型名称" json:"name"`
	Status     int       `gorm:"column:status;comment:状态" json:"status"`
	CreateTime LocalTime `gorm:"column:create_time;comment:创建时间" json:"createTime"`
	ProcessID  *int      `gorm:"column:process_id;comment:工序ID" json:"processId"`
	Count      int       `gorm:"column:count;comment:数量" json:"count"`
	Type       int       `gorm:"column:type;comment:类型" json:"type"`
}

// TableName 返回 nxp_inspection_question_type 表名
func (NxpInspectionQuestionType) TableName() string { return "nxp_inspection_question_type" }
