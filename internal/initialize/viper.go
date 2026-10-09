package initialize

import (
	"fmt"

	"sop/global"
	"sop/internal/config"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// v viper 全局实例（用于支持配置文件热重载）
var v = viper.New()

// Init 读取根目录 config.yaml 并加载配置
func Init() error {
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	return Reload()
}

// Reload 将 viper 中的最新配置解析为新的快照并原子发布到全局配置
func Reload() error {
	var cfg config.Config
	if err := v.Unmarshal(&cfg); err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}
	// 热重载由 fsnotify 回调 goroutine 触发，与请求 goroutine 并发读构成竞争，
	// 必须通过 SetConfig 原子替换整份快照，禁止裸指针赋值
	global.SetConfig(&cfg)
	return nil
}

// Watch 监听配置文件变化，变化时自动热重载全局配置
func Watch() {
	v.WatchConfig()
	v.OnConfigChange(func(in fsnotify.Event) {
		old := global.GetConfig()
		if err := Reload(); err != nil {
			global.Logger.Error("配置文件热重载失败", zap.String("file", in.Name), zap.Error(err))
			return
		}
		global.Logger.Info("配置文件已热重载", zap.String("file", in.Name))
		// 数据库连接池与 JWT 密钥在启动时一次性初始化，热重载不会重新初始化，
		// 这些段发生变化必须重启服务才生效，显式告警避免"改了配置却静默无效"的陷阱
		if old != nil {
			warnIfRestartNeeded(old, global.GetConfig())
		}
	})
}

// warnIfRestartNeeded 对比热重载前后的配置，对不会热生效的配置段输出重启提示
func warnIfRestartNeeded(old, newer *config.Config) {
	dbChanged := old.DatabaseOnebe != newer.DatabaseOnebe ||
		old.DatabaseSPAREPART != newer.DatabaseSPAREPART ||
		old.DatabaseSOPHub != newer.DatabaseSOPHub
	jwtChanged := old.JWT != newer.JWT
	portChanged := old.Server.Port != newer.Server.Port
	if dbChanged {
		global.Logger.Warn("databaseOnebe/databaseSPAREPART/databaseSOPHub 配置已变更，但数据库连接池不会热重建，需重启服务生效")
	}
	if jwtChanged {
		global.Logger.Warn("jwt 配置已变更，但 JWT 不会热重建，需重启服务生效")
	}
	if portChanged {
		global.Logger.Warn("server.port 配置已变更，需重启服务生效")
	}
}
