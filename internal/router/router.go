package router

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"sop/global"
	"sop/internal/middleware"
)

// Setup 注册所有路由并返回 gin 引擎
func Setup() *gin.Engine {
	srvCfg := global.GetConfig().Server
	// gin 运行模式由配置 server.mode 控制；未配置时保持 gin 默认（debug）并告警
	switch mode := strings.ToLower(srvCfg.Mode); mode {
	case gin.DebugMode, gin.ReleaseMode, gin.TestMode:
		gin.SetMode(mode)
	default:
		global.Logger.Warn("server.mode 未配置或不合法，gin 使用默认 debug 模式（生产环境必须配置 release）",
			zap.String("mode", srvCfg.Mode))
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()
	r.Use(middleware.Cors(), middleware.Logger(), middleware.Recovery())

	// 健康检查：对两个数据库做 Ping，任一不可用返回 503，便于负载均衡/探活及时摘除实例
	r.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		services := gin.H{}
		httpStatus := http.StatusOK
		for name, db := range map[string]*gorm.DB{"sophub": global.SOPhubDB, "onebe": global.OnebeDB} {
			if db == nil {
				services[name] = "down"
				httpStatus = http.StatusServiceUnavailable
				continue
			}
			sqlDB, err := db.DB()
			if err != nil || sqlDB.PingContext(ctx) != nil {
				services[name] = "down"
				httpStatus = http.StatusServiceUnavailable
				continue
			}
			services[name] = "ok"
		}
		status := "ok"
		if httpStatus != http.StatusOK {
			status = "degraded"
		}
		c.JSON(httpStatus, gin.H{"status": status, "services": services})
	})

	// 静态资源：上传文件（视频/PDF 等大文件传输，豁免服务器级读写超时）
	// 空前缀 Group 仅用于挂载 NoTimeout，Static 注册的路径与行为与原 r.Static 完全一致
	uploads := r.Group("", middleware.NoTimeout())
	uploads.Static("/uploads", srvCfg.UploadDir)
	// API 路由组
	apiGroup := r.Group("/api/v1")
	initBaseRouter(apiGroup)
	initUserRouter(apiGroup)
	initGuideRouter(apiGroup)
	initFacilityRouter(apiGroup)
	initInquiryRouter(apiGroup)
	initMediaRouter(apiGroup)
	initRepairRecordRouter(apiGroup)
	initMenuRouter(apiGroup)
	initExecutionOptionRouter(apiGroup)
	initFaultCategoryRouter(apiGroup)
	initMachineModelRouter(apiGroup)
	initInspectionQuestionTypeRouter(apiGroup)
	initInspectionPositionRouter(apiGroup)
	initInspectionContentRouter(apiGroup)
	initSparepartRouter(apiGroup)
	initSparepartAttrRouter(apiGroup)
	initEquipmentRouter(apiGroup)
	initInspectionServiceRouter(apiGroup)

	return r
}
