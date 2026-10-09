package router

import (
	"github.com/gin-gonic/gin"

	"sop/internal/api"
	"sop/internal/middleware"
)

// initUserRouter 注册用户相关路由（除 login 外均需 JWT 认证）
func initUserRouter(group *gin.RouterGroup) {
	userApi := api.UserApi{}
	userRouter := group.Group("user")
	// 需 JWT 认证接口
	auth := userRouter.Group("")
	auth.Use(middleware.JWTAuth())
	{
		// 【web端使用】（后台用户管理 / 会话恢复）
		auth.POST("getUserList", userApi.GetUserList) // 分页获取用户列表
		auth.GET("getUserInfo", userApi.GetUserInfo)  // 获取自身信息
		// 【app+web使用】按 ID 列表批量获取用户
		auth.POST("getUsersByIDs", userApi.GetUsersByIDs)
	}
}
