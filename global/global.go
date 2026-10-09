package global

import (
	"sop/internal/config"
	"sync/atomic"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// configValue 保存全局配置的不可变快照。
// 热重载在 fsnotify 回调 goroutine 中整包替换配置，请求 goroutine 并发读取，
// 裸指针赋值构成数据竞争（-race 必报）。指针的载入/存储走 atomic，
// 每个快照发布后只读不再修改，因此读写双方均安全。
var configValue atomic.Pointer[config.Config]

// SetConfig 原子发布一份新的配置快照（仅在启动加载与热重载时调用）
func SetConfig(cfg *config.Config) { configValue.Store(cfg) }

// GetConfig 返回当前配置快照；整个生命周期只取一次即可保证读到同一份一致视图
func GetConfig() *config.Config { return configValue.Load() }

// 全局变量集中存放，供项目各层统一访问
var (
	// SOPhubDB SOPHub 新库连接（读写），SOP 业务表全部在此
	SOPhubDB *gorm.DB
	// OnebeDB Onebe 老库连接（只读），仅供查询，禁止任何写操作
	OnebeDB *gorm.DB
	// SparepartDB SPAREPART 备件老库连接（只读），仅供查询，禁止任何写操作
	SparepartDB *gorm.DB
	// Logger 全局日志实例
	Logger *zap.Logger
)
