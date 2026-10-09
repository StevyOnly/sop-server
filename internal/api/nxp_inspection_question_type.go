package api

import (
	"github.com/gin-gonic/gin"

	"sop/internal/service"
	"sop/pkg/response"
)

// InspectionQuestionTypeApi 问题类型接口处理器
type InspectionQuestionTypeApi struct{}

// ListInspectionQuestionTypeRequest 问题类型列表请求
type ListInspectionQuestionTypeRequest struct {
	Page    int    `json:"page"`
	Limit   int    `json:"limit"`
	Keyword string `json:"keyword"`
	// ProcessID 按工艺 ID 过滤（process_id 列）：键缺失或传 null → 不过滤；
	// 传 0 → 仅查 process_id 为空的记录；传 >0 → 精确过滤
	ProcessID *int `json:"processId"`
}

// GetInspectionQuestionTypeList 分页获取问题类型列表
func (a InspectionQuestionTypeApi) GetInspectionQuestionTypeList(c *gin.Context) {
	var req ListInspectionQuestionTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}
	response.NormalizePage(&req.Page, &req.Limit, 100)

	list, total, err := service.GetInspectionQuestionTypeList(req.Page, req.Limit, req.Keyword, req.ProcessID)
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithPage(c, list, total, req.Page, req.Limit)
}
