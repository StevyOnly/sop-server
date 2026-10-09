package model

// NxpSopInquiry 现场提问（字段与前端 StepInquiry 类型对齐）
type NxpSopInquiry struct {
	BaseModel
	EngineerID        uint       `json:"engineerId" gorm:"comment:工程师"`
	GuideID           uint       `json:"guideId" gorm:"comment:关联Guide"`
	StepID            uint       `json:"stepId" gorm:"comment:关联步骤"`
	FacilityID        uint       `json:"facilityId" gorm:"comment:关联设施"`
	FaultCode         string     `json:"faultCode" gorm:"comment:故障码"`
	Question          string     `json:"question" gorm:"comment:问题内容"`
	Photos            []string   `json:"photos" gorm:"serializer:json;comment:现场图片列表"`
	Answer            string     `json:"answer" gorm:"comment:专家回复"`
	AnsweredBy        string     `json:"answeredBy" gorm:"comment:回复人"`
	AnsweredAt        *LocalTime `json:"answeredAt" gorm:"comment:回复时间"`
	Status            string     `json:"status" gorm:"comment:状态"`
	IsNewIssue        bool       `json:"isNewIssue" gorm:"comment:是否新问题"`
	ExecutionOptionID uint       `json:"executionOptionId" gorm:"comment:执行说明选项ID"`
}

// TableName 返回提问表名
func (NxpSopInquiry) TableName() string {
	return "nxp_sop_inquiry"
}
