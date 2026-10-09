package api

import (
	"github.com/gin-gonic/gin"

	"sop/internal/model"
	"sop/internal/service"
	"sop/pkg/response"
)

// SparepartAttrApi 备件 SOP 属性接口处理器
type SparepartAttrApi struct{}

// ListSparepartAttrRequest 备件属性列表请求
type ListSparepartAttrRequest struct {
	Page    int    `json:"page"`
	Limit   int    `json:"limit"`
	Keyword string `json:"keyword"`
}

// CreateSparepartAttrRequest 新增备件属性请求
// 独立 DTO：仅暴露业务字段，sId 必填（须为老库已存在的备件）。
type CreateSparepartAttrRequest struct {
	SID           string  `json:"sId" binding:"required"`
	StandardCycle int     `json:"standardCycle"`
	SafetyFactor  float64 `json:"safetyFactor"`
}

// UpdateSparepartAttrRequest 更新备件属性请求（sId 创建后不可变，不接收）
type UpdateSparepartAttrRequest struct {
	ID            uint    `json:"id" binding:"required"`
	StandardCycle int     `json:"standardCycle"`
	SafetyFactor  float64 `json:"safetyFactor"`
}

// SparepartAttrIDRequest 备件属性 ID 请求
type SparepartAttrIDRequest struct {
	ID uint `json:"id" binding:"required"`
}

// SparepartAttrSIDRequest 备件属性 sId 请求
type SparepartAttrSIDRequest struct {
	SID string `json:"sId" binding:"required"`
}

// GetSparepartAttrList 分页获取备件属性列表
func (a SparepartAttrApi) GetSparepartAttrList(c *gin.Context) {
	var req ListSparepartAttrRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}
	response.NormalizePage(&req.Page, &req.Limit, 100)

	list, total, err := service.GetSparepartAttrList(req.Page, req.Limit, req.Keyword)
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithPage(c, list, total, req.Page, req.Limit)
}

// GetSparepartAttrBySID 按备件 sId 获取属性
func (a SparepartAttrApi) GetSparepartAttrBySID(c *gin.Context) {
	var req SparepartAttrSIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}

	attr, err := service.GetSparepartAttrBySID(req.SID)
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithData(c, attr)
}

// CreateSparepartAttr 新增备件属性
func (a SparepartAttrApi) CreateSparepartAttr(c *gin.Context) {
	var req CreateSparepartAttrRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}

	attr := &model.NxpSopSparepartAttr{
		SID:           req.SID,
		StandardCycle: req.StandardCycle,
		SafetyFactor:  req.SafetyFactor,
	}
	if err := service.CreateSparepartAttr(attr); err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithData(c, attr)
}

// UpdateSparepartAttr 更新备件属性
func (a SparepartAttrApi) UpdateSparepartAttr(c *gin.Context) {
	var req UpdateSparepartAttrRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}

	if err := service.UpdateSparepartAttr(req.ID, req.StandardCycle, req.SafetyFactor); err != nil {
		response.FailError(c, err)
		return
	}
	response.OK(c)
}

// DeleteSparepartAttr 删除备件属性
func (a SparepartAttrApi) DeleteSparepartAttr(c *gin.Context) {
	var req SparepartAttrIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}

	if err := service.DeleteSparepartAttr(req.ID); err != nil {
		response.FailError(c, err)
		return
	}
	response.OK(c)
}
