package model

// NxpInspectionPosition 维修位置表，对应 nxp_inspection_position（Onebe 老库只读）
// 注意：该表 created_at/updated_at 为 datetime2 类型，按库表实际结构定义。
type NxpInspectionPosition struct {
	ID        int       `gorm:"column:id;primaryKey;comment:主键" json:"id"`
	Position  string    `gorm:"column:position;comment:维修位置" json:"position"`
	Status    int       `gorm:"column:status;comment:状态" json:"status"`
	CreatedAt LocalTime `gorm:"column:created_at;comment:创建时间" json:"createdAt"`
	UpdatedAt LocalTime `gorm:"column:updated_at;comment:更新时间" json:"updatedAt"`
	ModelID   string    `gorm:"column:model_id;comment:机型标识" json:"modelId"`
	ProcessID *int      `gorm:"column:process_id;comment:工序ID" json:"processId"`
}

// TableName 返回 nxp_inspection_position 表名
func (NxpInspectionPosition) TableName() string { return "nxp_inspection_position" }
