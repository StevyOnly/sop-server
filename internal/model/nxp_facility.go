package model

// NxpFacility 设备（设施）表，对应 nxp_facility
// 注意：该表 create_time/update_time/delete_time 为 int 类型（unix 秒），
// 因此未使用 BaseModel，而是按库表实际结构定义。
type NxpFacility struct {
	ID               uint   `gorm:"column:id;primaryKey" json:"id"`
	Name             string `gorm:"column:name" json:"name"`
	Number           string `gorm:"column:number" json:"number"`
	CreateUser       int    `gorm:"column:create_user" json:"createUser"`
	Model            string `gorm:"column:model" json:"model"`
	Area             int    `gorm:"column:area" json:"area"`
	Shift            int    `gorm:"column:shift" json:"shift"`
	SerialNo         string `gorm:"column:serial_no" json:"serialNo"`
	VendorID         int    `gorm:"column:vendor_id" json:"vendorId"`
	AssetNo          string `gorm:"column:ASSET_NO" json:"assetNo"`
	BuyOffDate       int    `gorm:"column:BUY_OFF_DATE" json:"buyOffDate"`
	CBAPassDate      int    `gorm:"column:CBA_pass_date" json:"cbaPassDate"`
	HeadUser         int    `gorm:"column:head_user" json:"headUser"`
	Status           int    `gorm:"column:status" json:"status"`
	IsOk             int    `gorm:"column:is_ok" json:"isOk"`
	File             string `gorm:"column:file" json:"file"`
	CreateTime       int    `gorm:"column:create_time" json:"createTime"`
	UpdateTime       int    `gorm:"column:update_time" json:"updateTime"`
	DeleteTime       int    `gorm:"column:delete_time" json:"deleteTime"`
	Process          string `gorm:"column:process" json:"process"`
	RelevantPeople   string `gorm:"column:relevant_people" json:"relevantPeople"`
	InspectionStatus int    `gorm:"column:inspection_status" json:"inspectionStatus"`
	ReviewUser       int    `gorm:"column:review_user" json:"reviewUser"`
	Floor            int    `gorm:"column:floor" json:"floor"`
	Remark           string `gorm:"column:remark" json:"remark"`
	IP               string `gorm:"column:ip" json:"ip"`
	IsReplace        int    `gorm:"column:is_replace" json:"isReplace"`
}

// TableName 返回 nxp_facility 表名
func (NxpFacility) TableName() string { return "nxp_facility" }
