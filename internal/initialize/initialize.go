package initialize

import (
	"sop/global"
	"sop/pkg/jwt"
)

// InitAll 依次执行全部初始化流程
func InitAll() error {
	if err := InitConfig(); err != nil {
		return err
	}
	if err := InitLogger(); err != nil {
		return err
	}
	if err := InitJWT(); err != nil {
		return err
	}
	return InitDatabase()
}

// InitJWT 根据配置初始化 JWT 签名密钥与有效期
func InitJWT() error {
	cfg := global.GetConfig()
	jwt.Init(cfg.JWT.Secret, cfg.JWT.ExpiresDays)
	return nil
}
