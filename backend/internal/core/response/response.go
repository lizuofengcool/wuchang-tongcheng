// Package response 统一响应封装
// 提供标准的API响应格式：{code, message, data}
package response

import (
	"net/http"
)

// Response 统一响应结构体
type Response struct {
	Code    int         `json:"code"`    // 状态码，0表示成功，非0表示失败
	Message string      `json:"message"` // 响应消息
	Data    interface{} `json:"data"`    // 响应数据
}

// 常用状态码定义
const (
	CodeSuccess      = 0   // 成功
	CodeBadRequest   = 400 // 请求参数错误
	CodeUnauthorized = 401 // 未授权
	CodeForbidden    = 403 // 禁止访问
	CodeNotFound     = 404 // 资源不存在
	CodeServerError  = 500 // 服务器内部错误
)

// Success 成功响应
func Success(data interface{}) *Response {
	return &Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	}
}

// SuccessWithMessage 成功响应（带自定义消息）
func SuccessWithMessage(message string, data interface{}) *Response {
	return &Response{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	}
}

// Fail 失败响应
func Fail(code int, message string) *Response {
	return &Response{
		Code:    code,
		Message: message,
		Data:    nil,
	}
}

// FailWithData 失败响应（带数据）
func FailWithData(code int, message string, data interface{}) *Response {
	return &Response{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// BadRequest 参数错误响应
func BadRequest(message string) *Response {
	return Fail(CodeBadRequest, message)
}

// Unauthorized 未授权响应
func Unauthorized(message string) *Response {
	if message == "" {
		message = "未授权访问"
	}
	return Fail(CodeUnauthorized, message)
}

// Forbidden 禁止访问响应
func Forbidden(message string) *Response {
	if message == "" {
		message = "禁止访问"
	}
	return Fail(CodeForbidden, message)
}

// NotFound 资源不存在响应
func NotFound(message string) *Response {
	if message == "" {
		message = "资源不存在"
	}
	return Fail(CodeNotFound, message)
}

// ServerError 服务器错误响应
func ServerError(message string) *Response {
	if message == "" {
		message = "服务器内部错误"
	}
	return Fail(CodeServerError, message)
}

// HTTPStatus 获取对应的HTTP状态码
func HTTPStatus(code int) int {
	switch code {
	case CodeSuccess:
		return http.StatusOK
	case CodeBadRequest:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeServerError:
		return http.StatusInternalServerError
	default:
		// 自定义业务错误码统一返回200，由前端根据code判断
		return http.StatusOK
	}
}

// PageResult 分页结果
type PageResult struct {
	List     interface{} `json:"list"`     // 数据列表
	Total    int64       `json:"total"`    // 总记录数
	Page     int         `json:"page"`     // 当前页码
	PageSize int         `json:"pageSize"` // 每页大小
}

// NewPageResult 创建分页结果
func NewPageResult(list interface{}, total int64, page, pageSize int) *PageResult {
	return &PageResult{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}
