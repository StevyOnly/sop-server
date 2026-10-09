package api

import (
	"github.com/gin-gonic/gin"

	"sop/internal/service"
	"sop/pkg/response"
)

// MachineModelApi 机型接口处理器
type MachineModelApi struct{}

// ListMachineModelRequest 机型列表请求
type ListMachineModelRequest struct {
	Page    int    `json:"page"`
	Limit   int    `json:"limit"`
	Keyword string `json:"keyword"`
}

// MachineModelByIDsRequest 按 ID 列表查询机型请求
// max=1000：对齐 SQL Server 单条语句参数上限约 2100，避免超限报错（与 GetFacilitiesByIDs 分片约定一致）
type MachineModelByIDsRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1,max=1000"`
}

// GetMachineModelList 分页获取机型列表
func (a MachineModelApi) GetMachineModelList(c *gin.Context) {
	var req ListMachineModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}
	response.NormalizePage(&req.Page, &req.Limit, 100)

	list, total, err := service.GetMachineModelList(req.Page, req.Limit, req.Keyword)
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithPage(c, list, total, req.Page, req.Limit)
}

// GetMachineModelsByIDs 按 ID 列表批量获取机型
func (a MachineModelApi) GetMachineModelsByIDs(c *gin.Context) {
	var req MachineModelByIDsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}
	list, err := service.GetMachineModelsByIDs(req.IDs)
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithData(c, list)
}
