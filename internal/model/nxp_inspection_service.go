package model

// NxpInspectionService 维修服务（报修维修）表，对应 nxp_inspection_service（Onebe 老库只读）
// 注意：该表 created_at/updated_at 为 int 类型（unix 秒），service_time/end_time 为 datetime 类型，按库表实际结构定义。
type NxpInspectionService struct {
	ID                    int       `gorm:"column:id;primaryKey;comment:主键" json:"id"`
	InspectionID          int       `gorm:"column:inspection_id;comment:报修表id" json:"inspectionId"`
	ContentID             int       `gorm:"column:content_id;comment:维修内容表id" json:"contentId"`
	PositionID            int       `gorm:"column:position_id;comment:维修位置表id" json:"positionId"`
	UserInfoID            int       `gorm:"column:user_info_id;comment:维修人id" json:"userInfoId"`
	Description           string    `gorm:"column:description;comment:维修情况说明" json:"description"`
	CreatedAt             int       `gorm:"column:created_at;comment:创建时间" json:"createdAt"`
	UpdatedAt             int       `gorm:"column:updated_at;comment:修改时间" json:"updatedAt"`
	ServiceTime           LocalTime `gorm:"column:service_time;comment:维修时间" json:"serviceTime"`
	FileID                string    `gorm:"column:file_id;comment:维修拍照id" json:"fileId"`
	Status                int8      `gorm:"column:status;comment:维修状态 1已完成 2中止维修" json:"status"`
	FacilityID            int       `gorm:"column:facility_id;comment:设备ID" json:"facilityId"`
	EndTime               LocalTime `gorm:"column:end_time;comment:维修结束时间" json:"endTime"`
	AbortReasonID         int       `gorm:"column:abort_reason_id;comment:中止原因id" json:"abortReasonId"`
	AbortReasonRemark     string    `gorm:"column:abort_reason_remark;comment:其他原因" json:"abortReasonRemark"`
	MaintenanceReasonID   int64     `gorm:"column:maintenance_reason_id;comment:无需维修原因id" json:"maintenanceReasonId"`
	MaintenanceReasonName string    `gorm:"column:maintenance_reason_name;comment:无需维修原因其他" json:"maintenanceReasonName"`
}

// TableName 返回 nxp_inspection_service 表名
func (NxpInspectionService) TableName() string { return "nxp_inspection_service" }
