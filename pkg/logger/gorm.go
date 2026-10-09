package logger

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

// GormOptions GORM 日志配置
type GormOptions struct {
	SlowThreshold time.Duration // 慢 SQL 阈值，<=0 时使用默认 200ms
	LogLevel      gormlogger.LogLevel
}

// GormLogger 将 GORM 日志接入 zap 的适配器，替换 GORM 默认输出到 stdout 的日志
type GormLogger struct {
	zapLogger     *zap.Logger
	slowThreshold time.Duration
	logLevel      gormlogger.LogLevel
}

// NewGormLogger 创建 GORM -> zap 日志适配器
func NewGormLogger(zapLogger *zap.Logger, opts GormOptions) gormlogger.Interface {
	if opts.SlowThreshold <= 0 {
		opts.SlowThreshold = 200 * time.Millisecond
	}
	return &GormLogger{
		zapLogger:     zapLogger,
		slowThreshold: opts.SlowThreshold,
		logLevel:      opts.LogLevel,
	}
}

// LogMode 返回一份仅日志等级不同的副本。
// GORM 契约要求 LogMode 不得修改原实例：db.Debug() 会在请求 goroutine 中调用
// db.Logger.LogMode(Info)，若原地修改共享实例，既与并发读取构成数据竞争，
// 又会把共享实例的等级永久提升（Warn 配置被泄漏为 Info）。
func (gl *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *gl
	newLogger.logLevel = level
	return &newLogger
}

// Info Info 等级日志
func (gl *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if gl.logLevel < gormlogger.Info {
		return
	}
	gl.zapLogger.Info(gl.msg(msg, data...))
}

// Warn Warn 等级日志
func (gl *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if gl.logLevel < gormlogger.Warn {
		return
	}
	gl.zapLogger.Warn(gl.msg(msg, data...))
}

// Error Error 等级日志
func (gl *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if gl.logLevel < gormlogger.Error {
		return
	}
	gl.zapLogger.Error(gl.msg(msg, data...))
}

// Trace 记录每条 SQL，包含调用位置、耗时、慢 SQL 阈值与错误
func (gl *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if gl.logLevel <= gormlogger.Silent {
		return
	}
	elapsed := time.Since(begin)
	sql, rows := fc()
	if rows == -1 {
		rows = 0
	}
	fields := []zap.Field{
		zap.String("line", utils.FileWithLineNum()),
		zap.String("sql", sql),
		zap.Int64("rows", rows),
		zap.Duration("cost", elapsed),
	}
	switch {
	case err != nil && gl.logLevel >= gormlogger.Error && !errors.Is(err, gorm.ErrRecordNotFound):
		fields = append(fields, zap.Error(err))
		gl.zapLogger.Error("GORM SQL ERROR", fields...)
	case elapsed > gl.slowThreshold && gl.logLevel >= gormlogger.Warn:
		fields = append(fields, zap.Duration("slow", gl.slowThreshold))
		gl.zapLogger.Warn("SLOW SQL", fields...)
	case gl.logLevel >= gormlogger.Info:
		gl.zapLogger.Info("GORM SQL", fields...)
	}
}

// msg 按 GORM 惯例格式化带参数的日志消息
func (gl *GormLogger) msg(msg string, data ...interface{}) string {
	if len(data) > 0 {
		return fmt.Sprintf(msg, data...)
	}
	return msg
}
