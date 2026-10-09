package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"sop/global"
)

// Recovery 捕获请求处理过程中的 panic，将堆栈写入统一 lumberjack 日志体系，
// 并对客户端返回脱敏的通用 500 报文（不泄露堆栈与内部信息）。
//
// 背景：gin.Recovery() 的 panic 堆栈只写到 os.Stderr（控制台），
// 绕过全局 global.Logger 的落盘/切割，生产环境会丢失崩溃现场；
// 且其默认 500 body 直接携带堆栈，与响应脱敏约定相悖。
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				global.Logger.Error("panic recovered",
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("ip", c.ClientIP()),
					zap.String("panic", fmt.Sprintf("%v", r)),
					zap.ByteString("stack", debug.Stack()),
				)
				// 挂到 c.Errors，使 middleware.Logger 的 errors 字段也能带出
				c.Error(fmt.Errorf("panic: %v", r))
				c.AbortWithStatusJSON(http.StatusInternalServerError,
					gin.H{"code": 7, "msg": "服务器内部错误", "data": nil})
			}
		}()
		c.Next()
	}
}
