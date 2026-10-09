package router

import (
	"github.com/gin-gonic/gin"

	"sop/internal/api"
)

// initBaseRouter 注册基础路由（登录等）
func initBaseRouter(group *gin.RouterGroup) {
	baseApi := api.BaseApi{}
	baseRouter := group.Group("base")

	// 【web端使用】用户登录（用户名密码）
	baseRouter.POST("login", baseApi.Login)
	// 【APP端使用】APP 端免密码登录（userId + appKey 换取永久 AppToken），POST /api/v1/base/appLogin
	baseRouter.POST("appLogin", baseApi.AppLogin)
}
