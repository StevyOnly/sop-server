package api

import (
	"github.com/gin-gonic/gin"

	"sop/internal/service"
	"sop/pkg/response"
)

// InspectionPositionApi 维修位置接口处理器
type InspectionPositionApi struct{}

// ListInspectionPositionRequest 维修位置列表请求
type ListInspectionPositionRequest struct {
	Page    int    `json:"page"`
	Limit   int    `json:"limit"`
	Keyword string `json:"keyword"`
	// ModelId 按机型过滤维修位置：position.model_id 为逗号分隔的机型编号（如 "44,45,60"），
	// 传 0 不过滤；>0 时仅返回包含该机型编号的位置（model_id 为空=通用位置，始终返回）
	ModelId int `json:"modelId"`
	// ProcessID 按工艺 ID 过滤（process_id 列）：键缺失或传 null → 不过滤；
	// 传 0 → 仅查 process_id 为空的记录；传 >0 → 精确过滤
	ProcessID *int `json:"processId"`
}

// GetInspectionPositionList 分页获取维修位置列表
func (a InspectionPositionApi) GetInspectionPositionList(c *gin.Context) {
	var req ListInspectionPositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}
	response.NormalizePage(&req.Page, &req.Limit, 100)

	list, total, err := service.GetInspectionPositionList(req.Page, req.Limit, req.Keyword, req.ModelId, req.ProcessID)
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithPage(c, list, total, req.Page, req.Limit)
}
