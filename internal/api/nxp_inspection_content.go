package api

import (
	"github.com/gin-gonic/gin"

	"sop/internal/service"
	"sop/pkg/response"
)

// InspectionContentApi 维修内容接口处理器
type InspectionContentApi struct{}

// ListInspectionContentRequest 维修内容列表请求
type ListInspectionContentRequest struct {
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
	PositionID int    `json:"positionId"`
	Keyword    string `json:"keyword"`
	// ProcessID 按工艺 ID 过滤（process_id 列）：键缺失或传 null → 不过滤；
	// 传 0 → 仅查 process_id 为空的记录；传 >0 → 精确过滤
	ProcessID *int `json:"processId"`
}

// GetInspectionContentList 分页获取维修内容列表
func (a InspectionContentApi) GetInspectionContentList(c *gin.Context) {
	var req ListInspectionContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}
	response.NormalizePage(&req.Page, &req.Limit, 100)

	list, total, err := service.GetInspectionContentList(req.Page, req.Limit, req.PositionID, req.Keyword, req.ProcessID)
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithPage(c, list, total, req.Page, req.Limit)
}
