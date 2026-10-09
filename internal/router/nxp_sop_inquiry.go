package router

import (
	"github.com/gin-gonic/gin"

	"sop/internal/api"
	"sop/internal/middleware"
)

// initInquiryRouter 注册提问相关路由
func initInquiryRouter(group *gin.RouterGroup) {
	inquiryApi := api.InquiryApi{}
	inquiryRouter := group.Group("/inquiries")
	inquiryRouter.Use(middleware.JWTAuth())
	{
		// 【web端使用】（后台提问管理）
		inquiryRouter.POST("list", inquiryApi.GetInquiryList)                                // 分页获取提问列表
		inquiryRouter.POST("reply", middleware.RequireActiveUser(), inquiryApi.ReplyInquiry) // 回复提问
		// 【APP端使用】APP 端提问与批量查询
		inquiryRouter.POST("getNxpSopInquiryByIds", inquiryApi.GetNxpSopInquiryByIds)          // 按 ID 列表批量获取提问
		inquiryRouter.POST("create", middleware.RequireActiveUser(), inquiryApi.CreateInquiry) // 创建提问
	}
}
