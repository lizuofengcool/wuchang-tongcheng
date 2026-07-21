// Package handler DIY 前端页面中台错误码常量定义
// 由于 error_code.go 不允许修改，本模块错误码在 plugin.go 的 init() 中注册到 utils
// 错误码区间：6001-6030（diy 模块独占）
package handler

// ===== diy 错误码常量（6001-6030） =====
const (
	// page 子域 6001-6010
	CodeDiyPageError         = 6001 // 页面通用错误
	CodeDiyPageNotFound      = 6002 // 页面不存在
	CodeDiyPageNoPermission  = 6003 // 无权操作页面
	CodeDiyPageStatusInvalid = 6004 // 页面状态不允许此操作
	CodeDiyPageSlugConflict  = 6005 // 页面 slug 已被占用
	CodeDiyPageSlugEmpty     = 6006 // 已发布页面必须设置 slug

	// component 子域 6011-6020
	CodeDiyComponentError       = 6011 // 组件通用错误
	CodeDiyComponentNotFound    = 6012 // 组件不存在
	CodeDiyComponentCodeConflict = 6013 // 组件编码已存在

	// template 子域 6021-6025
	CodeDiyTemplateError    = 6021 // 模板通用错误
	CodeDiyTemplateNotFound = 6022 // 模板不存在

	// stat 子域 6026-6030
	CodeDiyStatError   = 6026 // 统计通用错误
	CodeDiyStatNotFound = 6027 // 统计记录不存在
)
