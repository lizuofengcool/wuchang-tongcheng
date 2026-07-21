// Package handler 同城商城 HTTP 处理层 - 数据统计
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/mall/dto"
	"wuchang-tongcheng/internal/modules/mall/repository"
	"wuchang-tongcheng/internal/modules/mall/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// StatisticHandler 数据统计 HTTP 处理器
type StatisticHandler struct {
	svc service.StatisticService
}

// NewStatisticHandler 创建统计 Handler 实例
func NewStatisticHandler(svc service.StatisticService) *StatisticHandler {
	return &StatisticHandler{svc: svc}
}

// Upsert 写入/更新统计（M 端定时任务调用）
// POST /api/v1/mall/admin/statistics  （需 mall:audit 权限）
func (h *StatisticHandler) Upsert(ctx plugin.Context) {
	var req dto.UpsertStatisticRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	if err := h.svc.Upsert(regionID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallStatisticError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("写入成功", nil))
}

// List 统计列表
// GET /api/v1/mall/admin/statistics  （需 mall:audit 权限）
func (h *StatisticHandler) List(ctx plugin.Context) {
	var req dto.StatisticListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallStatisticError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Summary 统计汇总
// GET /api/v1/mall/statistics/summary  （公开）
func (h *StatisticHandler) Summary(ctx plugin.Context) {
	var req dto.StatisticSummaryRequest
	_ = ctx.Bind(&req)
	regionID := getRegionID(ctx)
	summary, err := h.svc.Summary(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallStatisticError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(summary))
}

// HotProducts 热销商品榜
// GET /api/v1/mall/statistics/hot-products  （公开）
func (h *StatisticHandler) HotProducts(ctx plugin.Context) {
	regionID := getRegionID(ctx)
	limit := parseQueryInt(ctx, "limit", 10)
	list, err := h.svc.HotProducts(regionID, limit)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallStatisticError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// HotShops 热门店铺榜
// GET /api/v1/mall/statistics/hot-shops  （公开）
func (h *StatisticHandler) HotShops(ctx plugin.Context) {
	regionID := getRegionID(ctx)
	limit := parseQueryInt(ctx, "limit", 10)
	list, err := h.svc.HotShops(regionID, limit)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallStatisticError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// HotCategories 热门分类榜
// GET /api/v1/mall/statistics/hot-categories  （公开）
func (h *StatisticHandler) HotCategories(ctx plugin.Context) {
	regionID := getRegionID(ctx)
	startDate := ctx.Query("start_date")
	endDate := ctx.Query("end_date")
	list, err := h.svc.HotCategories(regionID, startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallStatisticError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// Overview 平台总览
// GET /api/v1/mall/admin/statistics/overview  （需 mall:audit 权限）
func (h *StatisticHandler) Overview(ctx plugin.Context) {
	regionID := getRegionID(ctx)
	startDate := ctx.Query("start_date")
	endDate := ctx.Query("end_date")
	overview, err := h.svc.Overview(regionID, startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallStatisticError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(overview))
}

// 确保 repository 包被引用（service 接口返回 repository.HotProductStat 等类型，handler 透传无需显式调用）
var _ repository.HotProductStat
