package middleware

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"sop/global"
)

// NoTimeout 清除当前连接的读/写截止时间，使该请求不受服务器级
// ReadTimeout / WriteTimeout 约束（大文件上传与静态大文件传输用）。
//
// 说明：
//  1. 通过 http.ResponseController 穿透 gin 的 responseWriter 包装器
//     （gin >= 1.10 实现了 Unwrap），直达底层 net.Conn；
//  2. 只清读/写 deadline，不影响 ReadHeaderTimeout——Slowloris 慢请求头
//     防护对这些路径依然生效；IdleTimeout 作用于 keep-alive 空闲期，同样不受影响；
//  3. 仅应挂在确有长传输需求的路由上，其余接口保留服务器级超时兜底。
func NoTimeout() gin.HandlerFunc {
	return func(c *gin.Context) {
		rc := http.NewResponseController(c.Writer)
		noDeadline := time.Time{}
		// 清除失败仅可能因底层 ResponseWriter 不支持（标准 net/http 服务端不会出现），
		// 记日志后放行，不因设置豁免而中断业务请求。
		if err := rc.SetReadDeadline(noDeadline); err != nil && !errors.Is(err, http.ErrNotSupported) {
			global.Logger.Warn("NoTimeout: 清除读超时失败", zap.String("path", c.Request.URL.Path), zap.Error(err))
		}
		if err := rc.SetWriteDeadline(noDeadline); err != nil && !errors.Is(err, http.ErrNotSupported) {
			global.Logger.Warn("NoTimeout: 清除写超时失败", zap.String("path", c.Request.URL.Path), zap.Error(err))
		}
		c.Next()
	}
}
