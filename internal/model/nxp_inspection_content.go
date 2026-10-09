package model

// NxpInspectionContent 维修内容表，对应 nxp_inspection_content（Onebe 老库只读）
// 注意：该表 created_at/updated_at 为 int 类型（unix 秒），按库表实际结构定义。
type NxpInspectionContent struct {
	ID             int     `gorm:"column:id;primaryKey;comment:主键" json:"id"`
	PositionID     int     `gorm:"column:position_id;comment:关联维修位置ID" json:"positionId"`
	Content        string  `gorm:"column:content;comment:维修内容" json:"content"`
	Status         int     `gorm:"column:status;comment:状态" json:"status"`
	CreatedAt      int     `gorm:"column:created_at;comment:创建时间" json:"createdAt"`
	UpdatedAt      int     `gorm:"column:updated_at;comment:更新时间" json:"updatedAt"`
	PicType        int64   `gorm:"column:pic_type;comment:图片类型" json:"picType"`
	CheckPerson    string  `gorm:"column:check_person;comment:检查人" json:"checkPerson"`
	ProcessID      *int    `gorm:"column:process_id;comment:工序ID" json:"processId"`
	IsCheck        int     `gorm:"column:is_check;comment:是否检查" json:"isCheck"`
	CheckGroupID   int64   `gorm:"column:check_group_id;comment:检查组ID" json:"checkGroupId"`
	CheckProcessID int64   `gorm:"column:check_process_id;comment:检查工序ID" json:"checkProcessId"`
	IsReplaceMold  *int8   `gorm:"column:is_replace_mold;comment:是否换模" json:"isReplaceMold"`
}

// TableName 返回 nxp_inspection_content 表名
func (NxpInspectionContent) TableName() string { return "nxp_inspection_content" }
