package middleware

import (
	"time"

	"sop/global"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger 返回一个使用 zap 记录请求日志的 gin 中间件
// 记录请求方法、路径、查询参数、状态码、耗时、客户端 IP 等信息
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		fields := []zap.Field{
			zap.Int("status", status),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("ip", c.ClientIP()),
			zap.Duration("latency", latency),
			zap.Int("size", c.Writer.Size()),
		}
		if query != "" {
			fields = append(fields, zap.String("query", query))
		}
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()))
		}

		if status >= 500 {
			global.Logger.Error("request failed", fields...)
		} else {
			global.Logger.Info("request", fields...)
		}
	}
}
