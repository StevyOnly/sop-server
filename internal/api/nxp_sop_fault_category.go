package api

import (
	"github.com/gin-gonic/gin"

	"sop/internal/model"
	"sop/internal/service"
	"sop/pkg/response"
)

// FaultCategoryApi 故障分类接口处理器
type FaultCategoryApi struct{}

// ListFaultCategoryRequest 故障分类列表请求
type ListFaultCategoryRequest struct {
	Page    int    `json:"page"`
	Limit   int    `json:"limit"`
	Keyword string `json:"keyword"`
}

// SaveFaultCategoryRequest 保存故障分类请求
// 独立 DTO：仅暴露 id 与业务字段，label 必填。不沿用 model.X 类型别名。
type SaveFaultCategoryRequest struct {
	ID    uint   `json:"id"`
	Label string `json:"label" binding:"required"`
}

// FaultCategoryIDRequest 故障分类 ID 请求
type FaultCategoryIDRequest struct {
	ID uint `json:"id" binding:"required"`
}

// FaultCategoryByIDsRequest 按 ID 列表查询故障分类请求
// max=1000：对齐 SQL Server 单条语句参数上限约 2100，避免超限报错（与 GetFacilitiesByIDs 分片约定一致）
type FaultCategoryByIDsRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1,max=1000"`
}

// GetFaultCategoryList 分页获取故障分类列表
func (a FaultCategoryApi) GetFaultCategoryList(c *gin.Context) {
	var req ListFaultCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}
	response.NormalizePage(&req.Page, &req.Limit, 100)

	list, total, err := service.GetFaultCategoryList(req.Page, req.Limit, req.Keyword)
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithPage(c, list, total, req.Page, req.Limit)
}

// GetFaultCategoryByIDs 按 ID 列表批量获取故障分类
func (a FaultCategoryApi) GetFaultCategoryByIDs(c *gin.Context) {
	var req FaultCategoryByIDsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}
	list, err := service.GetFaultCategoryByIDs(req.IDs)
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithData(c, list)
}

// FaultCategoriesByModelIdRequest 按机型 ID 分页查询关联故障分类请求
type FaultCategoriesByModelIdRequest struct {
	Page           int  `json:"page"`
	Limit          int  `json:"limit"`
	MachineModelID uint `json:"machineModelId" binding:"required"`
}

// GetFaultCategoriesByModelId 按机型 ID 分页查询关联故障分类及报警明细
func (a FaultCategoryApi) GetFaultCategoriesByModelId(c *gin.Context) {
	var req FaultCategoriesByModelIdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}
	response.NormalizePage(&req.Page, &req.Limit, 100)

	list, total, err := service.GetFaultCategoriesByModelId(req.Page, req.Limit, req.MachineModelID)
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithPage(c, list, total, req.Page, req.Limit)
}

// CreateFaultCategory 新增故障分类
func (a FaultCategoryApi) CreateFaultCategory(c *gin.Context) {
	var req SaveFaultCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}

	cat := &model.NxpSopFaultCategory{Label: req.Label}
	if err := service.CreateFaultCategory(cat); err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithData(c, cat)
}

// UpdateFaultCategory 更新故障分类
func (a FaultCategoryApi) UpdateFaultCategory(c *gin.Context) {
	var req SaveFaultCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}

	cat := &model.NxpSopFaultCategory{Label: req.Label}
	cat.ID = req.ID
	if err := service.UpdateFaultCategory(cat); err != nil {
		response.FailError(c, err)
		return
	}
	response.OK(c)
}

// DeleteFaultCategory 删除故障分类
func (a FaultCategoryApi) DeleteFaultCategory(c *gin.Context) {
	var req FaultCategoryIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}

	if err := service.DeleteFaultCategory(req.ID); err != nil {
		response.FailError(c, err)
		return
	}
	response.OK(c)
}
