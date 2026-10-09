package api

import (
	"github.com/gin-gonic/gin"

	"sop/internal/middleware"
	"sop/internal/model"
	"sop/internal/service"
	"sop/pkg/response"
)

// MenuApi 菜单接口处理器
type MenuApi struct{}

// ListMenuRequest 菜单列表请求
type ListMenuRequest struct {
	Page    int    `json:"page"`
	Limit   int    `json:"limit"`
	Keyword string `json:"keyword"`
}

// SaveMenuRequest 保存菜单请求
// 独立 DTO 仅暴露 id 与业务字段；menuName 必填。不沿用 model.X 类型别名，
// 避免客户端可直写 BaseModel 的 createTime/updateTime 等内部字段。
// id 为 0 表示新增，非 0 表示重命名对应菜单（由 service 区分）。
type SaveMenuRequest struct {
	ID       uint   `json:"id"`
	MenuName string `json:"menuName" binding:"required"`
}

// MenuIDRequest 菜单 ID 请求
type MenuIDRequest struct {
	ID uint `json:"id" binding:"required"`
}

// PositionMenusRequest 职务-菜单关联请求
type PositionMenusRequest struct {
	PositionID int   `json:"positionId" binding:"required"`
	MenuIDs    []int `json:"menuIds"`
}

// GetMenuList 分页获取菜单列表
func (a MenuApi) GetMenuList(c *gin.Context) {
	var req ListMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}
	response.NormalizePage(&req.Page, &req.Limit, 100)

	list, total, err := service.GetMenuList(req.Page, req.Limit, req.Keyword)
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithPage(c, list, total, req.Page, req.Limit)
}

// GetMenuAll 获取全部菜单（供职务-菜单配置整表展示）
func (a MenuApi) GetMenuAll(c *gin.Context) {
	list, err := service.GetAllMenus()
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithData(c, gin.H{"list": list})
}

// SaveMenu 新增或更新菜单
func (a MenuApi) SaveMenu(c *gin.Context) {
	var req SaveMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}

	menu := model.NxpSopMenu{MenuName: req.MenuName}
	menu.ID = req.ID
	if err := service.SaveMenu(&menu); err != nil {
		response.FailError(c, err)
		return
	}
	response.OK(c)
}

// DeleteMenu 删除菜单
func (a MenuApi) DeleteMenu(c *gin.Context) {
	var req MenuIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}

	if err := service.DeleteMenu(req.ID); err != nil {
		response.FailError(c, err)
		return
	}
	response.OK(c)
}

// GetPositionList 获取全部职务
func (a MenuApi) GetPositionList(c *gin.Context) {
	list, err := service.GetPositionList()
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithData(c, gin.H{"list": list})
}

// GetPositionMenus 按职务获取可见菜单
func (a MenuApi) GetPositionMenus(c *gin.Context) {
	var req PositionMenusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}

	list, err := service.GetPositionMenus(req.PositionID)
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithData(c, gin.H{"list": list})
}

// GetPositionsWithMenus 获取全部职务及其可见菜单
func (a MenuApi) GetPositionsWithMenus(c *gin.Context) {
	list, err := service.GetPositionsWithMenus()
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithData(c, gin.H{"list": list})
}

// SavePositionMenus 保存职务-菜单关联（整体覆盖）
func (a MenuApi) SavePositionMenus(c *gin.Context) {
	var req PositionMenusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}

	if err := service.SavePositionMenus(req.PositionID, req.MenuIDs); err != nil {
		response.FailError(c, err)
		return
	}
	response.OK(c)
}

// GetUserMenus 获取当前登录用户可见的菜单
func (a MenuApi) GetUserMenus(c *gin.Context) {
	claims, ok := middleware.GetJWTClaims(c)
	if !ok {
		response.Fail(c, "未获取到用户信息")
		return
	}

	list, err := service.GetUserMenus(claims.UserID)
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithData(c, gin.H{"list": list})
}
