package api

import (
	"github.com/gin-gonic/gin"

	"sop/internal/service"
	"sop/pkg/response"
)

// SparepartApi 备件接口处理器
type SparepartApi struct{}

// SparepartListRequest 备件列表请求
type SparepartListRequest struct {
	Page    int    `json:"page"`
	Limit   int    `json:"limit"`
	Keyword string `json:"keyword"`
}

// GetSparepartList 分页获取备件列表
func (a SparepartApi) GetSparepartList(c *gin.Context) {
	var req SparepartListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}
	response.NormalizePage(&req.Page, &req.Limit, 100)

	list, total, err := service.GetSparepartList(req.Page, req.Limit, req.Keyword)
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithPage(c, list, total, req.Page, req.Limit)
}

// GetSparepartListWithAttr 分页获取备件列表（含 SOP 属性 attr）
func (a SparepartApi) GetSparepartListWithAttr(c *gin.Context) {
	var req SparepartListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}
	response.NormalizePage(&req.Page, &req.Limit, 100)

	list, total, err := service.GetSparepartListWithAttr(req.Page, req.Limit, req.Keyword)
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithPage(c, list, total, req.Page, req.Limit)
}
