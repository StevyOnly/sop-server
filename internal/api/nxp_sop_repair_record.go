package api

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"sop/internal/model"
	"sop/internal/service"
	"sop/pkg/response"
)

// RepairRecordApi 维修记录接口处理器
type RepairRecordApi struct{}

// CompleteRepairRequest 完成维修记录请求
// 前端 EngineerApp 提交字段：id, faultCode, photos[], isNewIssue
// RepairTime 可选（"2006-01-02 15:04:05" 或 RFC3339），表示维修实际开始时间；
// 缺省时后端按「本次提交即完工」处理，开始时间取当前时间。
type SaveRepairRecordRequest struct {
	ID                   uint            `json:"id"`
	GuideID              uint            `json:"guideId"`
	FacilityID           uint            `json:"facilityId"`
	EngineerID           uint            `json:"engineerId"`
	FaultCode            string          `json:"faultCode"`
	MachineModelID       uint            `json:"machineModelId"`
	StepID               uint            `json:"stepId"`
	RepairPositionID     uint            `json:"repairPositionId"`
	RepairContentID      uint            `json:"repairContentId"`
	SituationDescription string          `json:"situationDescription"`
	RequestNumber        string          `json:"requestNumber"`
	RepairNumber         string          `json:"repairNumber"`
	RepairTime           model.LocalTime `json:"repairTime"`
	Photos               []string        `json:"photos"`
	SparepartSIDs        []string        `json:"sparepartSIDs"`
	IsNewIssue           bool            `json:"isNewIssue"`
	Status               string          `json:"status"` // PDA 提交的维修状态（如：维修中/已完成/待料中）
}

// SaveRepairRecordWithDetails 保存维修执行记录（PDA 端完工/结束提交）
// 用户启用状态校验由路由上的 middleware.RequireActiveUser() 完成
func (a RepairRecordApi) SaveRepairRecordWithDetails(c *gin.Context) {
	var req SaveRepairRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		println(err.Error())
		bindFail(c, err)
		return
	}

	rec := model.NxpSopRepairRecord{
		GuideID:              req.GuideID,
		FacilityID:           req.FacilityID,
		EngineerID:           req.EngineerID,
		FaultCode:            req.FaultCode,
		MachineModelID:       req.MachineModelID,
		StepID:               req.StepID,
		RepairPositionID:     req.RepairPositionID,
		RepairContentID:      req.RepairContentID,
		SituationDescription: req.SituationDescription,
		RequestNumber:        req.RequestNumber,
		RepairNumber:         genRepairNumber(time.Now()),
		Photos:               req.Photos,
		SparepartSIDs:        req.SparepartSIDs,
		Status:               req.Status,
	}
	// 维修状态：存 PDA 提交的状态字符串；未传时兜底为 completed
	if rec.Status == "" {
		rec.Status = "completed"
	}
	// 开始时间：前端传了就用前端的真实开始时间；未传（零值）时保持原行为，由 service 兜底为当前时间
	if !req.RepairTime.IsZero() {
		start := req.RepairTime
		rec.RepairTime = &start
	}

	if err := service.SaveRepairRecord(&rec); err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithData(c, rec)
}

// NxpRepairRecordByIdsRequest 按 ID 列表批量查询维修记录请求
// max=1000：对齐 SQL Server 单条语句参数上限约 2100，避免超限报错（与 GetFacilitiesByIDs 分片约定一致）
type NxpRepairRecordByIdsRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1,max=1000"`
}

// GetNxpRepairRecordByIds 按 ID 列表批量获取维修记录
func (a RepairRecordApi) GetNxpRepairRecordByIds(c *gin.Context) {
	var req NxpRepairRecordByIdsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}
	list, err := service.GetNxpRepairRecordByIds(req.IDs)
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithData(c, list)
}

// genRepairNumber 生成维修单号：wx-yyyyMMddHHmmss + 3 位毫秒，例 wx-20260906120145123
func genRepairNumber(t time.Time) string {
	return fmt.Sprintf("wx-%s%03d", t.Format("20060102150405"), t.Nanosecond()/int(time.Millisecond))
}

// ListRepairRecordsRequest 维修记录列表请求
type ListRepairRecordsRequest struct {
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
	FacilityID uint   `json:"facilityId"`
	FaultCode  string `json:"faultCode"`
	EngineerID uint   `json:"engineerId"`
	Status     string `json:"status"`
}

// ListRepairRecords 分页获取维修记录列表（管理后台展示使用记录）
func (a RepairRecordApi) ListRepairRecords(c *gin.Context) {
	var req ListRepairRecordsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	list, total, err := service.GetRepairRecordList(req.Page, req.Limit, req.FacilityID, req.FaultCode, req.EngineerID, req.Status)
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithData(c, gin.H{
		"list":       list,
		"total":      total,
		"page":       req.Page,
		"pageSize":   req.Limit,
		"pagination": gin.H{"page": req.Page, "limit": req.Limit, "total": total},
	})
}
