package model

// NxpMachineModel 机型表，对应 nxp_machine_model
// 注意：该表 createdt/updatedt 为 nvarchar 字符串类型，按库表实际结构定义。
type NxpMachineModel struct {
	ID                 int    `gorm:"column:id;primaryKey;comment:主键" json:"id"`
	Name               string `gorm:"column:name;comment:机型" json:"name"`
	Status             int    `gorm:"column:status;comment:状态：1启用，0禁用" json:"status"`
	CreateBy           int    `gorm:"column:createby;comment:发布人" json:"createBy"`
	Createdt           string `gorm:"column:createdt;comment:发布时间" json:"createdt"`
	Updatedt           string `gorm:"column:updatedt;comment:更新时间" json:"updatedt"`
	InspectionPosition string `gorm:"column:inspection_position;comment:维修位置" json:"inspectionPosition"`
	ProcessID          int    `gorm:"column:process_id;comment:工序ID" json:"processId"`
	Owner              int    `gorm:"column:owner;comment:负责人" json:"owner"`
}

// TableName 返回 nxp_machine_model 表名
func (NxpMachineModel) TableName() string { return "nxp_machine_model" }
