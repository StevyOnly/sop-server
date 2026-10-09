package model

// Equipment 设备机型字典表，对应 SPAREPART 库 dbo.equipment
// 该表为备件老库的设备/机型字典（ID 为自增主键，无 create_time/update_time 等审计字段），
// 因此未使用 BaseModel，而是按库表实际结构定义。
type Equipment struct {
	ID       int    `gorm:"column:ID;primaryKey" json:"id"`
	EModel   string `gorm:"column:e_model" json:"eModel"`
	EBrand   string `gorm:"column:e_brand" json:"eBrand"`
	RProcess string `gorm:"column:r_process" json:"rProcess"`
	EShort   string `gorm:"column:e_short" json:"eShort"`
}

// TableName 返回 equipment 表名
func (Equipment) TableName() string { return "equipment" }
