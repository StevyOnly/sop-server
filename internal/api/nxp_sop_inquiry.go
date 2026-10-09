package api

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"sop/global"
	"sop/internal/middleware"
	"sop/internal/model"
	"sop/internal/service"
	"sop/pkg/response"
)

// InquiryApi 提问接口处理器
type InquiryApi struct{}

// ListInquiryRequest 提问列表请求
type ListInquiryRequest struct {
	Page      int    `json:"page"`
	Limit     int    `json:"limit"`
	FaultCode string `json:"faultCode"` // 故障码筛选，为空不筛
	Status    string `json:"status"`    // 状态筛选，为空不筛
}

// CreateInquiryRequest 创建提问请求
type CreateInquiryRequest struct {
	EngineerID uint     `json:"engineerId"`
	GuideID    uint     `json:"guideId"`
	StepID     uint     `json:"stepId"`
	FacilityID uint     `json:"facilityId"`
	FaultCode  string   `json:"faultCode"`
	Question   string   `json:"question"`
	Photos     []string `json:"photos"`
	IsNewIssue bool     `json:"isNewIssue"`
	// 执行说明选项 ID（关联 nxp_sop_execution_option.id），
	// 0 表示该次反馈未选择执行说明
	ExecutionOptionID uint `json:"executionOptionId"`
}

// GetInquiryList 分页获取提问列表
func (a InquiryApi) GetInquiryList(c *gin.Context) {
	var req ListInquiryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}
	response.NormalizePage(&req.Page, &req.Limit, 100)

	list, total, err := service.GetInquiryList(req.Page, req.Limit, req.FaultCode, req.Status)
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithPage(c, list, total, req.Page, req.Limit)
}

// CreateInquiry 创建提问
// 用户启用状态校验由路由上的 middleware.RequireActiveUser() 完成
func (a InquiryApi) CreateInquiry(c *gin.Context) {
	var req CreateInquiryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}

	inquiry := model.NxpSopInquiry{
		EngineerID:        req.EngineerID,
		GuideID:           req.GuideID,
		StepID:            req.StepID,
		FacilityID:        req.FacilityID,
		FaultCode:         req.FaultCode,
		Question:          req.Question,
		Photos:            req.Photos,
		IsNewIssue:        req.IsNewIssue,
		ExecutionOptionID: req.ExecutionOptionID,
	}
	inquiry.Status = service.InquiryStatusPending

	if err := service.CreateInquiry(&inquiry); err != nil {
		response.FailError(c, err)
		return
	}

	// 现场提问成功后，将该步骤的历史维修次数 +1
	if inquiry.StepID != 0 {
		if err := service.IncrementStepRepairCount(inquiry.StepID); err != nil {
			// 次数统计失败不影响提问主流程
			global.Logger.Warn("increment step repair count failed",
				zap.Uint("stepID", inquiry.StepID),
				zap.Error(err))
		}
	}

	response.OKWithData(c, inquiry)
}

// InquiryByIdsRequest 按 ID 列表批量查询提问请求
// max=1000：对齐 SQL Server 单条语句参数上限约 2100，避免超限报错（与 GetFacilitiesByIDs 分片约定一致）
type InquiryByIdsRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1,max=1000"`
}

// GetNxpSopInquiryByIds 按 ID 列表批量获取提问
func (a InquiryApi) GetNxpSopInquiryByIds(c *gin.Context) {
	var req InquiryByIdsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}
	list, err := service.GetNxpSopInquiryByIds(req.IDs)
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithData(c, list)
}

// ReplyInquiryRequest 管理员回复提问请求
type ReplyInquiryRequest struct {
	ID         uint   `json:"id" binding:"required"`
	Answer     string `json:"answer"`
	Status     string `json:"status"` // 可选：resolved / pending
	AnsweredBy string `json:"answeredBy"`
}

// ReplyInquiry 管理员回复提问
// 用户启用状态校验由路由上的 middleware.RequireActiveUser() 完成
func (a InquiryApi) ReplyInquiry(c *gin.Context) {
	var req ReplyInquiryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}
	if req.Answer == "" {
		response.Fail(c, "回复内容不能为空")
		return
	}

	// 回复人优先取登录态用户名（RequireActiveUser 已反查）；反查不到时回退前端传入值
	repliedBy := req.AnsweredBy
	if user := middleware.GetActiveUser(c); user != nil && user.User.Name != "" {
		repliedBy = user.User.Name
	}

	if err := service.UpdateInquiryReply(req.ID, req.Answer, req.Status, repliedBy); err != nil {
		response.FailError(c, err)
		return
	}
	response.OK(c)
}
