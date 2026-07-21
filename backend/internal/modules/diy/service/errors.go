// Package service DIY 前端页面中台业务逻辑层 - 错误定义
package service

import "errors"

// ===== 各子域错误 =====
var (
	// page 子域
	ErrPageNotFound     = errors.New("页面不存在")
	ErrPageNoPermission = errors.New("无权操作此页面")
	ErrPageStatusInvalid = errors.New("页面状态不允许此操作")
	ErrPageSlugConflict = errors.New("页面 slug 已被占用")
	ErrPageSlugEmpty    = errors.New("已发布页面必须设置 slug")

	// component 子域
	ErrComponentNotFound  = errors.New("组件不存在")
	ErrComponentCodeConflict = errors.New("组件编码已存在")

	// template 子域
	ErrTemplateNotFound = errors.New("模板不存在")

	// stat 子域
	ErrStatNotFound = errors.New("统计记录不存在")
)
