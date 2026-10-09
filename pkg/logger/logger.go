package logger

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sop/global"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Cut 日志切割粒度
type Cut string

const (
	// CutHour 按小时切割
	CutHour Cut = "hour"
	// CutDay 按天切割
	CutDay Cut = "day"
)

// Options 日志配置
type Options struct {
	Level      string // 日志等级：debug/info/warn/error
	Dir        string // 日志目录，如 logs
	Filename   string // 日志文件名，如 app.log（最终文件名会带上日期，如 app-2026-08-13.log）
	MaxSize    int    // 单个日志文件大小上限（MB）
	MaxAge     int    // 保留的旧日志文件最大天数
	MaxBackups int    // 保留的旧日志文件最大个数
	Compress   bool   // 是否压缩旧的日志文件
	Cut        Cut    // 切割粒度：day / hour
	Console    bool   // 是否同时输出日志到终端控制台
}

// timeWriter 按日期时间切割的日志写入器
type timeWriter struct {
	mu       sync.Mutex
	opts     Options
	logger   *lumberjack.Logger
	curKey   string
	basePath string // 去掉日期后的基础文件名（含路径）
}

// Write 每次写入前检测时间窗口是否变化，若变化则切换日志文件
func (w *timeWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	key := w.windowKey(time.Now())
	if key != w.curKey {
		if err := w.switchFile(key); err != nil {
			return 0, err
		}
	}
	return w.logger.Write(p)
}

// windowKey 根据切割粒度生成时间窗口标识
func (w *timeWriter) windowKey(t time.Time) string {
	if w.opts.Cut == CutHour {
		return t.Format("2006-01-02-15")
	}
	return t.Format("2006-01-02")
}

// switchFile 关闭旧文件并打开新日期文件
func (w *timeWriter) switchFile(key string) error {
	if w.logger != nil {
		w.logger.Close()
	}

	filename := w.filenameForKey(key)
	w.logger = &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    w.opts.MaxSize,
		MaxAge:     w.opts.MaxAge,
		MaxBackups: w.opts.MaxBackups,
		LocalTime:  true,
		Compress:   w.opts.Compress,
	}
	w.curKey = key
	return nil
}

// filenameForKey 生成带日期后缀的文件名
func (w *timeWriter) filenameForKey(key string) string {
	dir := filepath.Dir(w.basePath)
	ext := filepath.Ext(w.basePath)
	base := filepath.Base(w.basePath)
	base = base[:len(base)-len(ext)]
	return filepath.Join(dir, fmt.Sprintf("%s-%s%s", base, key, ext))
}

// Init 初始化全局日志
func Init(opts Options) error {
	if opts.Dir == "" {
		return errors.New("日志目录不能为空")
	}
	if opts.Filename == "" {
		return errors.New("日志文件名不能为空")
	}
	if opts.Cut == "" {
		opts.Cut = CutDay
	}
	if opts.MaxSize <= 0 {
		opts.MaxSize = 100
	}

	if err := os.MkdirAll(opts.Dir, 0o755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	basePath := filepath.Join(opts.Dir, opts.Filename)
	tw := &timeWriter{
		opts:     opts,
		basePath: basePath,
	}
	key := tw.windowKey(time.Now())
	if err := tw.switchFile(key); err != nil {
		return fmt.Errorf("初始化日志文件失败: %w", err)
	}

	// 将文件写入器与（可选）终端输出合并，实现日志同时落盘与显示
	writers := []zapcore.WriteSyncer{zapcore.AddSync(tw)}
	if opts.Console {
		writers = append(writers, zapcore.AddSync(os.Stdout))
	}
	writerSync := zapcore.NewMultiWriteSyncer(writers...)

	level := zapcore.InfoLevel
	if opts.Level != "" {
		if l, err := zapcore.ParseLevel(opts.Level); err == nil {
			level = l
		}
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		writerSync,
		level,
	)

	global.Logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(0))
	return nil
}
