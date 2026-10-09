package api

import (
	"github.com/gin-gonic/gin"

	"sop/internal/model"
	"sop/internal/service"
	"sop/pkg/response"
)

// ExecutionOptionApi 执行说明选项接口处理器
type ExecutionOptionApi struct{}

// ListExecutionOptionRequest 执行说明选项列表请求
type ListExecutionOptionRequest struct {
	Page    int    `json:"page"`
	Limit   int    `json:"limit"`
	Keyword string `json:"keyword"`
}

// SaveExecutionOptionRequest 保存执行说明选项请求
// 独立 DTO：仅暴露业务字段，label 必填。不沿用 model.X 类型别名（不直写内部字段）。
type SaveExecutionOptionRequest struct {
	ID        uint   `json:"id"`
	Label     string `json:"label" binding:"required"`
	SortOrder int    `json:"sortOrder"`
	Active    bool   `json:"active"`
}

// ExecutionOptionIDRequest 执行说明选项 ID 请求
type ExecutionOptionIDRequest struct {
	ID uint `json:"id" binding:"required"`
}

// GetExecutionOptionList 分页获取执行说明选项列表
func (a ExecutionOptionApi) GetExecutionOptionList(c *gin.Context) {
	var req ListExecutionOptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}
	response.NormalizePage(&req.Page, &req.Limit, 100)

	list, total, err := service.GetExecutionOptionList(req.Page, req.Limit, req.Keyword)
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithPage(c, list, total, req.Page, req.Limit)
}

// GetActiveExecutionOptions 获取启用中的执行说明选项（供 PDA 下拉）
func (a ExecutionOptionApi) GetActiveExecutionOptions(c *gin.Context) {
	list, err := service.GetActiveExecutionOptions()
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithData(c, gin.H{"list": list})
}

// CreateExecutionOption 新增执行说明选项
func (a ExecutionOptionApi) CreateExecutionOption(c *gin.Context) {
	var req SaveExecutionOptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}

	opt := &model.NxpSopExecutionOption{
		Label:     req.Label,
		SortOrder: req.SortOrder,
		Active:    req.Active,
	}
	if err := service.CreateExecutionOption(opt); err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithData(c, opt)
}

// UpdateExecutionOption 更新执行说明选项
func (a ExecutionOptionApi) UpdateExecutionOption(c *gin.Context) {
	var req SaveExecutionOptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}

	opt := &model.NxpSopExecutionOption{
		Label:     req.Label,
		SortOrder: req.SortOrder,
		Active:    req.Active,
	}
	opt.ID = req.ID
	if err := service.UpdateExecutionOption(opt); err != nil {
		response.FailError(c, err)
		return
	}
	response.OK(c)
}

// DeleteExecutionOption 删除执行说明选项
func (a ExecutionOptionApi) DeleteExecutionOption(c *gin.Context) {
	var req ExecutionOptionIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}

	if err := service.DeleteExecutionOption(req.ID); err != nil {
		response.FailError(c, err)
		return
	}
	response.OK(c)
}
