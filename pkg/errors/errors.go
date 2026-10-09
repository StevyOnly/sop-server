package errors

import stderrors "errors"

// 通用业务错误码（HTTP 与业务码合并使用，方便前端按 code 统一拦截）
const (
	CodeSuccess      = 0
	CodeFail         = 7   // 通用失败
	CodeParamError   = 400 // 参数错误
	CodeUnauthorized = 401 // 未认证 / 认证失败
	CodeForbidden    = 403 // 无权限
	CodeNotFound     = 404 // 资源不存在
	CodeConflict     = 409 // 冲突（如用户名/邮箱已存在）
	CodeServerError  = 500
)

// BizError 带业务码与 HTTP 状态码的业务错误
type BizError struct {
	Code int    // 业务码
	HTTP int    // HTTP 状态码
	Msg  string // 错误提示
}

// Error 返回错误消息（实现 error 接口）
func (e *BizError) Error() string {
	return e.Msg
}

// New 创建带业务码的错误
func New(code, http int, msg string) *BizError {
	return &BizError{Code: code, HTTP: http, Msg: msg}
}

// NewParam 参数错误（400）
func NewParam(msg string) *BizError {
	return New(CodeParamError, 400, msg)
}

// NewNotFound 资源不存在（404）
func NewNotFound(msg string) *BizError {
	return New(CodeNotFound, 404, msg)
}

// NewConflict 冲突（409）
func NewConflict(msg string) *BizError {
	return New(CodeConflict, 409, msg)
}

// NewUnauthorized 未认证 / 认证失败（401）
func NewUnauthorized(msg string) *BizError {
	return New(CodeUnauthorized, 401, msg)
}

// IsBiz 判定错误是否为业务错误（*BizError）。
// 业务错误的 Msg 是预定义的用户可读文案，可安全返回客户端；
// 非业务错误（驱动、文件系统、网络等原始错误）可能含表名、SQL 片段、服务器路径等内部信息，禁止透出。
func IsBiz(err error) bool {
	var be *BizError
	return stderrors.As(err, &be)
}

// StatusOf 从错误中提取 HTTP 状态码与业务码；未知错误回退 500/通用码
func StatusOf(err error) (http, code int) {
	if be, ok := err.(*BizError); ok {
		if be.HTTP > 0 {
			return be.HTTP, be.Code
		}
		return CodeServerError, be.Code
	}
	return CodeServerError, CodeFail
}

// 用户业务错误
var (
	ErrUserNotFound   = NewNotFound("用户不存在")
	ErrUsernameExists = NewConflict("用户名已存在")
	ErrWrongPassword  = New(CodeFail, 400, "原密码错误")
	ErrEmailExists    = NewConflict("邮箱已存在")
	ErrPasswordOrUser = NewUnauthorized("用户名或密码错误")
	ErrUserDisabled   = New(CodeForbidden, 403, "用户已被禁用")
)

// 业务错误
var (
	ErrGuideNotFound        = NewNotFound("指南不存在")
	ErrGuideInUse           = New(CodeConflict, 409, "该指南已被提问或维修记录引用，不能删除")
	ErrDeviceNotFound       = NewNotFound("设备不存在")
	ErrMediaNotFound        = NewNotFound("媒体资源不存在")
	ErrInquiryNotFound      = NewNotFound("提问不存在")
	ErrRepairRecordNotFound = NewNotFound("维修记录不存在")
	ErrMenuNotFound         = NewNotFound("菜单不存在")
	ErrPositionNotFound     = NewNotFound("职务不存在")
)

// 执行说明选项业务错误
var (
	ErrExecutionOptionNotFound = NewNotFound("执行说明选项不存在")
	ErrExecutionOptionInUse    = New(CodeConflict, 409, "该执行说明选项已被提问引用，不能删除")
)

// 故障分类业务错误
var (
	ErrFaultCategoryNotFound = NewNotFound("故障分类不存在")
)

// 备件 SOP 属性业务错误
var (
	ErrSparepartAttrNotFound = NewNotFound("备件属性不存在")
	ErrSparepartAttrExists   = NewConflict("该备件已配置属性")
	ErrSparepartNotFound     = NewNotFound("备件不存在")
)
