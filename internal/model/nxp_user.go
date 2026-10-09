package model

// NxpUser 用户登陆表，主键 id 即用户唯一标识（对应 JWT 中的 UserID）
type NxpUser struct {
	BaseModelWithoutUpdateTime
	UserID   int    `gorm:"column:user_id" json:"userId"`
	Name     string `gorm:"column:name" json:"name"`
	Password string `gorm:"column:password" json:"-"`
}

// TableName 返回 nxp_user 表名
func (NxpUser) TableName() string { return "nxp_user" }
