package model

// NxpInspectionServiceSparepart 维修服务备件关联表，对应 nxp_inspection_service_sparepart（Onebe 老库只读）
type NxpInspectionServiceSparepart struct {
	ID                  int    `gorm:"column:id;primaryKey;comment:主键" json:"id"`
	InspectionServiceID int    `gorm:"column:inspection_service_id;comment:维修ID" json:"inspectionServiceId"`
	SID                 string `gorm:"column:s_id;comment:备件ID" json:"sId"`
	EID                 string `gorm:"column:e_id;comment:设备编号" json:"eId"`
	OAmount             int    `gorm:"column:o_amount;comment:备件领取数量" json:"oAmount"`
}

// TableName 返回 nxp_inspection_service_sparepart 表名
func (NxpInspectionServiceSparepart) TableName() string { return "nxp_inspection_service_sparepart" }
