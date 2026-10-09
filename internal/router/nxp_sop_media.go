package router

import (
	"github.com/gin-gonic/gin"

	"sop/internal/api"
	"sop/internal/middleware"
)

// initMediaRouter 注册媒体资源相关路由
func initMediaRouter(group *gin.RouterGroup) {
	mediaApi := api.MediaApi{}
	mediaRouter := group.Group("/media")
	mediaRouter.Use(middleware.JWTAuth())
	{
		// 上传接口豁免服务器级读写超时（大文件上传耗时不确定），置于链首优先生效
		// 【web端使用】（后台 SOP 编辑器 / 媒体库上传）
		mediaRouter.POST("uploadMedia", middleware.NoTimeout(), middleware.RequireActiveUser(), mediaApi.UploadMedia)           // 上传文件
		mediaRouter.POST("uploadScenePhoto", middleware.NoTimeout(), middleware.RequireActiveUser(), mediaApi.UploadScenePhoto) // 上传现场照片 —— 【APP端使用】
		// 【web端使用】（后台媒体库管理）
		mediaRouter.POST("createMedia", middleware.RequireActiveUser(), mediaApi.CreateMedia) // 创建媒体资源信息
		mediaRouter.POST("updateMedia", middleware.RequireActiveUser(), mediaApi.UpdateMedia) // 更新媒体资源信息
		mediaRouter.POST("deleteMedia", middleware.RequireActiveUser(), mediaApi.DeleteMedia) // 删除媒体资源
		// 【web端使用】（后台媒体库 / SOP 编辑器媒体选择）
		mediaRouter.POST("getMediaList", mediaApi.GetMediaList) // 分页获取媒体资源列表
	}
}
