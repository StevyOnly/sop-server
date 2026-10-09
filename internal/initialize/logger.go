package initialize

import (
	"sop/global"
	"sop/pkg/logger"
)

// InitLogger 根据配置初始化日志
func InitLogger() error {
	cfg := global.GetConfig()
	return logger.Init(logger.Options{
		Level:      cfg.Logger.Level,
		Dir:        cfg.Logger.Dir,
		Filename:   cfg.Logger.Filename,
		MaxSize:    cfg.Logger.MaxSize,
		MaxAge:     cfg.Logger.MaxAge,
		MaxBackups: cfg.Logger.MaxBackups,
		Compress:   cfg.Logger.Compress,
		Cut:        logger.Cut(cfg.Logger.Cut),
		Console:    cfg.Logger.Console,
	})
}
