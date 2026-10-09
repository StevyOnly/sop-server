package model

type BaseModelWithoutUpdateTime struct {
	ID         uint      `json:"id" gorm:"primaryKey;column:id"`
	CreateTime LocalTime `json:"createTime" gorm:"column:create_time;autoCreateTime"`
}

// BaseModel 基础模型，包含通用的主键和时间字段
type BaseModel struct {
	BaseModelWithoutUpdateTime
	UpdateTime LocalTime `json:"updateTime" gorm:"column:update_time;autoUpdateTime"`
}
