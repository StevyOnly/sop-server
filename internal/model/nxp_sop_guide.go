package model

// NxpSopGuideStep SOP 步骤（字段与前端 GuideStep 类型对齐）
type NxpSopGuideStep struct {
	BaseModel
	GuideID            uint     `json:"guideID" gorm:"index;comment:关联Guide ID"`
	Number             int      `json:"number" gorm:"not null;default:0;comment:步骤序号"`
	Stage              string   `json:"stage" gorm:"comment:阶段"`
	Title              string   `json:"title" gorm:"comment:步骤标题"`
	Description        string   `json:"description" gorm:"comment:步骤描述"`
	Instruction        string   `json:"instruction" gorm:"comment:操作说明"`
	ImageUrls          []string `json:"imageUrls" gorm:"serializer:json;comment:多图片"`
	VideoUrls          []string `json:"videoUrls" gorm:"serializer:json;comment:多视频"`
	PdfUrls            []string `json:"pdfUrls" gorm:"serializer:json;comment:多PDF"`
	JudgmentMethod     string   `json:"judgmentMethod" gorm:"comment:判断方法"`
	ExpertAdvice       string   `json:"expertAdvice" gorm:"column:expert_advice;comment:专家建议"`
	SafetyWarning      string   `json:"safetyWarning" gorm:"comment:安全提示"`
	Enabled            *bool    `json:"enabled" gorm:"comment:是否启用"`
	HistoryRepairCount int      `json:"historyRepairCount" gorm:"comment:历史维修次数"`
}

// TableName 返回 SOP 步骤表名
func (NxpSopGuideStep) TableName() string {
	return "nxp_sop_guide_step"
}

// NxpSopGuide SOP 指南（字段与前端 MaintenanceGuide 类型对齐）
type NxpSopGuide struct {
	BaseModel
	MachineModelId       uint              `json:"machineModelId" gorm:"column:machine_model_id;comment:关联机型"`
	FaultCode            string            `json:"faultCode" gorm:"index;comment:故障码"`
	FaultCategoryID      uint              `json:"faultCategoryId" gorm:"column:fault_category_id;comment:关联故障分类(NxpSopFaultCategory)"`
	QuestionTypeID       uint              `json:"questionTypeId" gorm:"column:question_type_id;comment:关联问题类型(NxpInspectionQuestionType)"`
	RepairPositionID     uint              `json:"repairPositionId" gorm:"column:repair_position_id;comment:关联维修位置(NxpInspectionPosition)"`
	RepairContentID      uint              `json:"repairContentId" gorm:"column:repair_content_id;comment:关联维修内容(NxpInspectionContent)"`
	FaultCodeDescription string            `json:"faultCodeDescription" gorm:"column:fault_code_description;comment:报警代码描述"`
	Enabled              bool              `json:"enabled" gorm:"comment:是否启用"`
	TotalOccurrenceCount int64             `json:"totalOccurrenceCount" gorm:"comment:总发生次数"`
	Steps                []NxpSopGuideStep `json:"steps" gorm:"foreignKey:GuideID;references:ID"`
}

// TableName 返回 Guide 表名
func (NxpSopGuide) TableName() string {
	return "nxp_sop_guide"
}
