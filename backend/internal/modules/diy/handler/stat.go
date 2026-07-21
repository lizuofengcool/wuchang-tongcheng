// Package handler DIY 前端页面中台 HTTP 处理层 - 统计（stat 子域）
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/diy/dto"
	"wuchang-tongcheng/internal/modules/diy/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// StatHandler 统计 HTTP 处理器
type StatHandler struct {
	svc service.StatService
}

// NewStatHandler 创建统计 Handler 实例
func NewStatHandler(svc service.StatService) *StatHandler {
	return &StatHandler{svc: svc}
}

// RecordView 记录浏览（公开）
// POST /api/v1/diy/stats/view
func (h *StatHandler) RecordView(ctx plugin.Context) {
	var req dto.RecordViewRequest
	if err := ctx.Bind(&req); err != nil {
		badRequest(ctx, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.RecordView(req.PageID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDiyStatError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已记录浏览", nil))
}

// RecordClick 记录点击（公开）
// POST /api/v1/diy/stats/click
func (h *StatHandler) RecordClick(ctx plugin.Context) {
	var req dto.RecordClickRequest
	if err := ctx.Bind(&req); err != nil {
		badRequest(ctx, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.RecordClick(req.PageID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDiyStatError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已记录点击", nil))
}

// RecordConversion 记录转化（公开）
// POST /api/v1/diy/stats/conversion
func (h *StatHandler) RecordConversion(ctx plugin.Context) {
	var req dto.RecordConversionRequest
	if err := ctx.Bind(&req); err != nil {
		badRequest(ctx, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.RecordConversion(req.PageID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDiyStatError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已记录转化", nil))
}

// ListByPageID 按页面 ID 列出统计（admin 权限）
// GET /api/v1/diy/admin/stats/page/:id
func (h *StatHandler) ListByPageID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		badRequest(ctx, "无效的ID")
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListByPageID(id, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDiyStatError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByDateRange 按日期范围列出统计（admin 权限）
// GET /api/v1/diy/admin/stats/date-range
func (h *StatHandler) ListByDateRange(ctx plugin.Context) {
	var req dto.StatSummaryRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.ListByDateRange(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDiyStatError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// SumByPageID 按页面 ID 汇总（admin 权限）
// GET /api/v1/diy/admin/stats/summary/page/:id
func (h *StatHandler) SumByPageID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		badRequest(ctx, "无效的ID")
		return
	}
	summary, err := h.svc.SumByPageID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDiyStatNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(summary))
}

// SumByDateRange 按日期范围汇总（admin 权限）
// GET /api/v1/diy/admin/stats/summary
func (h *StatHandler) SumByDateRange(ctx plugin.Context) {
	var req dto.StatSummaryRequest
	_ = ctx.Bind(&req)
	summary, err := h.svc.SumByDateRange(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDiyStatNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(summary))
}
