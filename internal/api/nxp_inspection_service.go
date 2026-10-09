package api

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"

	"sop/internal/service"
	"sop/pkg/response"
)

// InspectionServiceApi 维修服务接口处理器
type InspectionServiceApi struct{}

// SparepartReplaceListRequest 备件更换列表请求
type SparepartReplaceListRequest struct {
	response.PageQuery
	InspectionServiceID int    `json:"inspectionServiceId"` // 维修ID（精确）
	SID                 string `json:"sId"`                 // 备件ID（模糊）
	EID                 string `json:"eId"`                 // 设备编号（模糊）
	SModel              string `json:"sModel"`              // 备件型号（sparepart.s_model，模糊）
	StartDate           string `json:"startDate"`           // 维修开始日期（年月日，可选，如 2026-09-01）
	EndDate             string `json:"endDate"`             // 维修结束日期（年月日，可选，如 2026-09-30，含当天）
}

// SparepartReplaceList 备件更换列表（web端使用）
func (a InspectionServiceApi) SparepartReplaceList(c *gin.Context) {
	var req SparepartReplaceListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}
	response.NormalizePage(&req.Page, &req.PageSize, 100)

	list, total, err := service.GetSparepartReplaceList(req.Page, req.PageSize, req.InspectionServiceID, req.SID, req.EID, req.SModel, req.StartDate, req.EndDate)
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithPage(c, list, total, req.Page, req.PageSize)
}

// SparepartReplaceExcelExportRequest 备件更换台账导出请求（无分页）
type SparepartReplaceExcelExportRequest struct {
	InspectionServiceID int    `json:"inspectionServiceId"` // 维修ID（精确）
	SID                 string `json:"sId"`                 // 备件ID（模糊）
	EID                 string `json:"eId"`                 // 设备编号（模糊）
	SModel              string `json:"sModel"`              // 备件型号（sparepart.s_model，模糊）
	StartDate           string `json:"startDate"`           // 维修开始日期（年月日，可选，如 2026-09-01）
	EndDate             string `json:"endDate"`             // 维修结束日期（年月日，可选，如 2026-09-30，含当天）
}

// SparepartReplaceExcelExport 备件更换台账导出 Excel（web端使用，响应当前筛选全部结果）
func (a InspectionServiceApi) SparepartReplaceExcelExport(c *gin.Context) {
	var req SparepartReplaceExcelExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}

	list, err := service.GetSparepartReplaceListExport(req.InspectionServiceID, req.SID, req.EID, req.SModel, req.StartDate, req.EndDate)
	if err != nil {
		response.FailError(c, err)
		return
	}

	var buf bytes.Buffer
	if err := buildSparepartReplaceExcel(&buf, list); err != nil {
		response.FailError(c, err)
		return
	}

	// RFC5987 编码中文文件名，避免下载乱码
	filename := fmt.Sprintf("备件更换台账_%s.xlsx", time.Now().Format("20060102"))
	disposition := fmt.Sprintf("attachment; filename*=UTF-8''%s", url.PathEscape(filename))
	c.Header("Content-Disposition", disposition)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

// SparepartModelOptions 备件型号下拉（web端使用）：返回更换记录中实际用过的去重型号（含示例 s_id）
func (a InspectionServiceApi) SparepartModelOptions(c *gin.Context) {
	list, err := service.GetSparepartModelOptions()
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithData(c, list)
}

// sparepartReplaceExcelHeaders 备件更换台账导出列（与列表回显一致，8 列）
var sparepartReplaceExcelHeaders = []string{"更换单号", "设备编号", "备件编号", "备件型号", "备件描述", "维修日期", "维修原因", "操作人"}

// buildSparepartReplaceExcel 将备件更换记录写入 xlsx 工作簿并输出到 buf。
// 字段缺失时输出空串，保证行列对齐。关联查不到的子对象为 nil，取不到的值留空。
func buildSparepartReplaceExcel(buf *bytes.Buffer, list []service.SparepartReplaceItem) error {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	sheet := f.GetSheetName(0)

	// 表头行：加粗 + 行高 + 居中
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return err
	}
	for i, h := range sparepartReplaceExcelHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheet, cell, h); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheet, cell, cell, headerStyle); err != nil {
			return err
		}
	}
	if err := f.SetRowHeight(sheet, 1, 22); err != nil {
		return err
	}

	// 数据行
	for i, item := range list {
		row := make([]interface{}, 0, len(sparepartReplaceExcelHeaders))
		// 更换单号（最外层 inspectionServiceId）
		if item.InspectionServiceID > 0 {
			row = append(row, item.InspectionServiceID)
		} else {
			row = append(row, "")
		}
		row = append(row, item.EID, item.SID)

		// 备件型号 / 备件描述（跨库备件字典）
		sModel, sDesc := "", ""
		if item.Sparepart != nil {
			sModel, sDesc = item.Sparepart.SModel, item.Sparepart.SDescription
		}
		row = append(row, sModel, sDesc)

		// 维修日期 / 维修原因 / 操作人（维修服务；userName 反查，查不到为空）
		svcTime, desc, userName := "", "", ""
		if item.InspectionService != nil {
			svc := item.InspectionService
			if !svc.ServiceTime.IsZero() {
				svcTime = svc.ServiceTime.Format("2006-01-02 15:04:05")
			}
			desc, userName = svc.Description, svc.UserName
		}
		row = append(row, svcTime, desc, userName)

		if err := f.SetSheetRow(sheet, fmt.Sprintf("A%d", i+2), &row); err != nil {
			return err
		}
	}

	// 冻结首行 + 自适应列宽
	if err := f.SetPanes(sheet, &excelize.Panes{Freeze: true, YSplit: 1}); err != nil {
		return err
	}
	widths := []float64{12, 16, 16, 20, 30, 20, 30, 12}
	for i, w := range widths {
		colLetter, _ := excelize.ColumnNumberToName(i + 1)
		if err := f.SetColWidth(sheet, colLetter, colLetter, w); err != nil {
			return err
		}
	}

	if _, err := f.WriteTo(buf); err != nil {
		return err
	}
	return nil
}
