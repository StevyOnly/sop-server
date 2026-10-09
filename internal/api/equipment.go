package api

import (
	"github.com/gin-gonic/gin"

	"sop/internal/service"
	"sop/pkg/response"
)

// EquipmentApi 设备机型接口处理器
type EquipmentApi struct{}

// GetEquipmentList 获取全部设备机型（无分页，查全部）
func (a EquipmentApi) GetEquipmentList(c *gin.Context) {
	list, err := service.GetAllEquipments()
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithData(c, gin.H{"list": list})
}
