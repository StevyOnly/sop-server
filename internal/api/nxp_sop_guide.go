package api

import (
	"github.com/gin-gonic/gin"

	"sop/internal/model"
	"sop/internal/service"
	"sop/pkg/response"
)

// GuideApi 指南接口处理器
type GuideApi struct{}

// ListGuideRequest 指南列表请求
type ListGuideRequest struct {
	Page            int    `json:"page"`
	Limit           int    `json:"limit"`
	FaultCode       string `json:"faultCode"`
	FaultCategoryID uint   `json:"faultCategoryId"`
	Keyword         string `json:"keyword"`
}

// SaveGuideRequest 保存指南请求
type SaveGuideRequest = model.NxpSopGuide

// GuideStatusRequest 更新指南状态请求
type GuideStatusRequest struct {
	ID      uint `json:"id" binding:"required"`
	Enabled bool `json:"enabled"`
}

// GuideIDRequest 指南 ID 请求
type GuideIDRequest struct {
	ID uint `json:"id" binding:"required"`
}

// GuideSampleByMachineModelIdsRequest 按设备（Facility ID）列表分页获取指南精简视图请求
// ids 为 nxp_facility.id，通过 nxp_facility.model（机器型号/machine_model_id）中转后查询指南。
// ids 上限 1000：对齐 SQL Server 单条语句参数上限约 2100
type GuideSampleByMachineModelIdsRequest struct {
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
	Ids   []uint `json:"ids" binding:"required,min=1,max=1000"`
}

// GetGuideList 分页获取指南列表
func (a GuideApi) GetGuideList(c *gin.Context) {
	var req ListGuideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}
	response.NormalizePage(&req.Page, &req.Limit, 100)

	list, total, err := service.GetGuideList(req.Page, req.Limit, req.FaultCode, req.FaultCategoryID, req.Keyword)
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithPage(c, list, total, req.Page, req.Limit)
}

// GetGuideSampleByMachineModelIds 按设备（Facility ID）列表分页获取指南精简视图（不含步骤）
func (a GuideApi) GetGuideSampleByMachineModelIds(c *gin.Context) {
	var req GuideSampleByMachineModelIdsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}
	response.NormalizePage(&req.Page, &req.Limit, 100)

	list, total, err := service.GetGuideSamplesByMachineModelIds(req.Page, req.Limit, req.Ids)
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithPage(c, list, total, req.Page, req.Limit)
}

// GuideByIdsRequest 按指南 ID 列表分页获取指南请求
// ids 上限 1000：对齐 SQL Server 单条语句参数上限约 2100
type GuideByIdsRequest struct {
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
	Ids   []uint `json:"ids" binding:"required,min=1,max=1000"`
}

// GetGuideByIds 按指南 ID 列表分页获取指南（含步骤）
func (a GuideApi) GetGuideByIds(c *gin.Context) {
	var req GuideByIdsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}
	response.NormalizePage(&req.Page, &req.Limit, 100)

	list, total, err := service.GetGuidesByIds(req.Page, req.Limit, req.Ids)
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithPage(c, list, total, req.Page, req.Limit)
}

// StepsByFaultCodeRequest 按故障码获取对应步骤请求
type StepsByFaultCodeRequest struct {
	FaultCode string `json:"faultCode" binding:"required"`
}

// GetStepsByFaultCode 按故障码获取对应步骤（data 直接为扁平步骤数组，不包 list）
func (a GuideApi) GetStepsByFaultCode(c *gin.Context) {
	var req StepsByFaultCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}

	list, err := service.GetStepsByFaultCode(req.FaultCode)
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithData(c, list)
}

// SaveGuide 保存指南（创建或更新）
func (a GuideApi) SaveGuide(c *gin.Context) {
	var req SaveGuideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}

	if err := service.SaveGuide(&req); err != nil {
		response.FailError(c, err)
		return
	}
	response.OK(c)
}

// UpdateGuideStatus 更新指南发布状态
func (a GuideApi) UpdateGuideStatus(c *gin.Context) {
	var req GuideStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}

	if err := service.UpdateGuideStatus(req.ID, req.Enabled); err != nil {
		response.FailError(c, err)
		return
	}
	response.OK(c)
}

// DeleteGuide 删除指南
func (a GuideApi) DeleteGuide(c *gin.Context) {
	var req GuideIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}

	if err := service.DeleteGuide(req.ID); err != nil {
		response.FailError(c, err)
		return
	}
	response.OK(c)
}
