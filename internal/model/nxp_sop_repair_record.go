package model

// NxpSopRepairRecord 维修记录（字段与前端 RepairRecord 类型对齐）
type NxpSopRepairRecord struct {
	BaseModel
	GuideID              uint       `json:"guideId" gorm:"comment:关联Guide"`
	EngineerID           uint       `json:"engineerId" gorm:"comment:工程师"`
	FacilityID           uint       `json:"facilityId" gorm:"column:facility_id;comment:关联设备"`
	MachineModelID       uint       `json:"machineModelId" gorm:"column:machine_model_id;comment:关联机型"`
	StepID               uint       `json:"stepId" gorm:"column:step_id;comment:关联步骤"`
	RepairPositionID     uint       `json:"repairPositionId" gorm:"column:repair_position_id;comment:关联维修位置"`
	RepairContentID      uint       `json:"repairContentId" gorm:"column:repair_content_id;comment:关联维修内容"`
	SituationDescription string     `json:"situationDescription" gorm:"column:situation_description;comment:现状描述"`
	RequestNumber        string     `json:"requestNumber" gorm:"column:request_number;comment:报修单号"`
	RepairNumber         string     `json:"repairNumber" gorm:"column:repair_number;comment:维修单号"`
	FaultCode            string     `json:"faultCode" gorm:"column:fault_code;comment:故障码"`
	RepairTime           *LocalTime `json:"repairTime" gorm:"column:repair_time;comment:维修时间"`
	Photos               []string   `json:"photos" gorm:"serializer:json;comment:现场照片"`
	Status               string     `json:"status" gorm:"comment:状态"`
	SparepartSIDs        []string   `json:"sparepartSIDs" gorm:"column:sparepart_s_ids;serializer:json;comment:关联备件ID列表"`
}

// TableName 返回维修记录表名
func (NxpSopRepairRecord) TableName() string {
	return "nxp_sop_repair_record"
}
