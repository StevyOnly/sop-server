package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sop/global"
	"sop/internal/database"
	"sop/internal/initialize"
	"sop/internal/router"

	"go.uber.org/zap"
)

func main() {
	// 初始化配置、日志与数据库
	if err := initialize.InitAll(); err != nil {
		log.Fatalf("init failed: %v", err)
	}

	// 注册路由并启动服务
	r := router.Setup()
	addr := fmt.Sprintf(":%d", global.GetConfig().Server.Port)

	server := &http.Server{
		Addr:    addr,
		Handler: r,
		// 超时防护：缺省为不限时，Slowloris 慢连接会长期占用 goroutine 与连接句柄。
		// ReadHeaderTimeout 防护慢速请求头；此处读写超时为全局默认兜底，
		// 媒体上传与 /uploads 静态大文件传输由 middleware.NoTimeout() 按路由豁免（不限时）。
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       90 * time.Second, // keep-alive 空闲连接回收，须 ≥ ReadHeaderTimeout
	}

	// 启动 HTTP 服务
	go func() {
		global.Logger.Info("server starting", zap.String("addr", addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			global.Logger.Fatal("server listen failed", zap.Error(err))
		}
	}()

	// 等待退出信号，实现优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	global.Logger.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		global.Logger.Error("server forced to shutdown", zap.Error(err))
	}
	// 在途请求已排空（或已超时），再关闭数据库连接池
	if err := database.Close(); err != nil {
		global.Logger.Error("close database pools failed", zap.Error(err))
	}
	global.Logger.Info("server exited")
}
