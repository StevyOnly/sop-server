package model

import "github.com/shopspring/decimal"

// Sparepart 备件表，对应 SPAREPART 库 sparepart
// 该表主键 s_id 为 char 类型，且无 create_time/update_time 等审计字段，
// 因此未使用 BaseModel，而是按库表实际结构定义。
type Sparepart struct {
	SID          string          `gorm:"column:s_id;primaryKey" json:"sId"`
	SBrand       string          `gorm:"column:s_brand" json:"sBrand"`
	SModel       string          `gorm:"column:s_model" json:"sModel"`
	SDescription string          `gorm:"column:s_description" json:"sDescription"`
	KKind        string          `gorm:"column:k_kind" json:"kKind"`
	SVendor      string          `gorm:"column:s_vendor" json:"sVendor"`
	SPrice       decimal.Decimal `gorm:"column:s_price" json:"sPrice"`
	SPicture     string          `gorm:"column:s_picture" json:"sPicture"`
	SSafeAmount  int             `gorm:"column:s_safeamount" json:"sSafeAmount"`
	SAmount      int             `gorm:"column:s_amount" json:"sAmount"`
	WID          string          `gorm:"column:w_id" json:"wId"`
	WarehouseID  string          `gorm:"column:warehouse_id" json:"warehouseId"`
}

// TableName 返回 sparepart 表名
func (Sparepart) TableName() string { return "sparepart" }
