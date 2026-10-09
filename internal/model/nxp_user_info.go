package model

// 用户状态常量（nxp_user_info.status）
// 1=启用，2=禁用；values 由各端约定
const (
	UserStatusDisabled int8 = 2 // 禁用
	UserStatusEnabled  int8 = 1 // 启用
)

// NxpUserInfo 主用户信息表
type NxpUserInfo struct {
	BaseModel
	RoleID           *int   `gorm:"column:role_id" json:"roleId"`
	Status           *int8  `gorm:"column:status" json:"status"`
	LeaderUID        *int   `gorm:"column:leader_uid" json:"leaderUid"`
	PID              string `gorm:"column:pid" json:"pid"`
	Tel              string `gorm:"column:tel" json:"tel"`
	ChineseName      string `gorm:"column:chinese_name" json:"chineseName"`
	EnglishName      string `gorm:"column:english_name" json:"englishName"`
	Sex              int8   `gorm:"column:sex" json:"sex"`
	Email            string `gorm:"column:email" json:"email"`
	Position         *int   `gorm:"column:position" json:"position"`
	Shift            *int   `gorm:"column:shift" json:"shift"`
	Area             string `gorm:"column:area" json:"area"`
	StationID        *int   `gorm:"column:station_id" json:"stationId"`
	SFCNumber        string `gorm:"column:sfc_number" json:"sfcNumber"`
	FSLNumber        string `gorm:"column:fsl_number" json:"fslNumber"`
	MealCardCode     string `gorm:"column:meal_card_code" json:"mealCardCode"`
	WBICode          string `gorm:"column:wbi_code" json:"wbiCode"`
	BadgeCode        string `gorm:"column:badge_code" json:"badgeCode"`
	Skill            *int   `gorm:"column:skill" json:"skill"`
	DepartmentNumber string `gorm:"column:department_number" json:"departmentNumber"`
	Section          string `gorm:"column:section" json:"section"`
	GembaRole        *int   `gorm:"column:gemba_role" json:"gembaRole"`
	Team             *int   `gorm:"column:team" json:"team"`
	ShoeType         *int   `gorm:"column:shoe_type" json:"shoeType"`
	NewArea          string `gorm:"column:new_area" json:"newArea"`
	SectionIR        *int   `gorm:"column:section_ir" json:"sectionIr"`
}

// TableName 返回 nxp_user_info 表名
func (NxpUserInfo) TableName() string { return "nxp_user_info" }
