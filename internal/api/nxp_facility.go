package api

import (
	"github.com/gin-gonic/gin"

	"sop/internal/service"
	"sop/pkg/response"
)

// FacilityApi 设备（设施）接口处理器
type FacilityApi struct{}

// FacilityListRequest 设备列表请求
type FacilityListRequest struct {
	Page    int    `json:"page"`
	Limit   int    `json:"limit"`
	Keyword string `json:"keyword"`
}

// FacilityByIDsRequest 按 ID 列表查询设备请求
// max=1000：对齐 SQL Server 单条语句参数上限约 2100，避免超限报错（与 GetUsersByIDs 分片约定一致）
type FacilityByIDsRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1,max=1000"`
}

// GetFacilityList 分页获取设备（设施）列表
func (a FacilityApi) GetFacilityList(c *gin.Context) {
	var req FacilityListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}
	response.NormalizePage(&req.Page, &req.Limit, 100)

	list, total, err := service.GetFacilityList(req.Page, req.Limit, req.Keyword)
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithPage(c, list, total, req.Page, req.Limit)
}

// GetFacilitiesByIDs 按 ID 列表批量获取设备（设施）
func (a FacilityApi) GetFacilitiesByIDs(c *gin.Context) {
	var req FacilityByIDsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}
	list, err := service.GetFacilitiesByIDs(req.IDs)
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithData(c, list)
}
