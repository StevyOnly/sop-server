package router

import (
	"github.com/gin-gonic/gin"

	"sop/internal/api"
	"sop/internal/middleware"
)

// initMenuRouter 注册菜单与职务相关路由
func initMenuRouter(group *gin.RouterGroup) {
	menuApi := api.MenuApi{}
	menuRouter := group.Group("/menu")
	menuRouter.Use(middleware.JWTAuth())
	{
		menuRouter.POST("getMenuList", menuApi.GetMenuList)                                              // 分页获取菜单列表
		menuRouter.POST("getMenuAll", menuApi.GetMenuAll)                                                // 获取全部菜单
		menuRouter.POST("saveMenu", middleware.RequireActiveUser(), menuApi.SaveMenu)                    // 新增/更新菜单
		menuRouter.POST("deleteMenu", middleware.RequireActiveUser(), menuApi.DeleteMenu)                // 删除菜单
		menuRouter.POST("getUserMenus", menuApi.GetUserMenus)                                            // 获取当前用户可见菜单
		menuRouter.POST("getPositionList", menuApi.GetPositionList)                                      // 获取全部职务
		menuRouter.POST("getPositionMenus", menuApi.GetPositionMenus)                                    // 按职务获取可见菜单
		menuRouter.POST("getPositionsWithMenus", menuApi.GetPositionsWithMenus)                          // 获取全部职务及其菜单
		menuRouter.POST("savePositionMenus", middleware.RequireActiveUser(), menuApi.SavePositionMenus)  // 保存职务-菜单关联
	}
}
