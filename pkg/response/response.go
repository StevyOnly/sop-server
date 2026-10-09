package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"sop/global"
	bizerr "sop/pkg/errors"
)

// 错误码
const (
	CodeSuccess = 0 // 成功
	CodeFail    = 7 // 失败
)

// Body 统一响应结构
type Body struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

// Result 返回成功响应（HTTP 200）
func Result(c *gin.Context, data interface{}, msg string) {
	c.JSON(http.StatusOK, Body{Code: CodeSuccess, Data: data, Msg: msg})
}

// OK 返回成功响应，无数据
func OK(c *gin.Context) {
	Result(c, gin.H{}, "操作成功")
}

// OKWithData 返回成功响应并携带数据
func OKWithData(c *gin.Context, data interface{}) {
	Result(c, data, "操作成功")
}

// OKWithMessage 返回成功响应并携带自定义消息
func OKWithMessage(c *gin.Context, msg string) {
	Result(c, gin.H{}, msg)
}

// Fail 返回失败响应（HTTP 200，通用业务码）
func Fail(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Body{Code: CodeFail, Data: nil, Msg: msg})
}

// FailError 根据错误返回对应的业务码与 HTTP 状态码
// 未知错误回退 500；业务错误使用 errors 包中登记的码值
//
// 安全约定：仅 BizError 的 Msg（预定义的用户可读文案）可返回客户端；
// 非 BizError 的原始错误（驱动报文、SQL 片段、文件系统路径等）一律脱敏为通用提示，
// 原文只写入服务端日志。
func FailError(c *gin.Context, err error) {
	httpStatus, code := bizerr.StatusOf(err)
	msg := "服务器内部错误"
	if bizerr.IsBiz(err) {
		msg = err.Error()
	} else {
		global.Logger.Error("internal error",
			zap.String("path", c.Request.URL.Path),
			zap.Error(err))
	}
	c.JSON(httpStatus, Body{Code: code, Data: nil, Msg: msg})
}

// PageData 分页数据结构
type PageData struct {
	List       interface{} `json:"list"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	Pagination struct {
		Page  int   `json:"page"`
		Limit int   `json:"limit"`
		Total int64 `json:"total"`
	} `json:"pagination"`
}

// OKWithPage 返回分页成功响应（data.list 与 data.pagination）
func OKWithPage(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	p := PageData{List: list, Total: total, Page: page, PageSize: pageSize}
	p.Pagination.Page = page
	p.Pagination.Limit = pageSize
	p.Pagination.Total = total
	Result(c, p, "操作成功")
}

// PageQuery 分页查询基础参数
type PageQuery struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Keyword  string `json:"keyword"`
	OrderKey string `json:"orderKey"`
	Desc     bool   `json:"desc"`
}

// Normalize 校验并归一化分页参数
func (p *PageQuery) Normalize() {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 10
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
}

// NormalizePage 归一化分页参数并限制分页大小上限，防止一次拉取全表
func NormalizePage(page, pageSize *int, maxPageSize int) {
	if page == nil || pageSize == nil {
		return
	}
	if *page <= 0 {
		*page = 1
	}
	if *pageSize <= 0 {
		*pageSize = 10
	}
	if maxPageSize > 0 && *pageSize > maxPageSize {
		*pageSize = maxPageSize
	}
}
