// Package handler 同城114 HTTP 处理层 - 统计
// 依据 v3.2.1 架构方案：日统计/商户统计/分类统计
package handler

import (
	"net/http"
	"time"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/dh114/model"
	"wuchang-tongcheng/internal/modules/dh114/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// StatisticHandler 统计 HTTP 处理器
type StatisticHandler struct {
	svc service.StatisticService
}

// NewStatisticHandler 创建统计 Handler 实例
func NewStatisticHandler(svc service.StatisticService) *StatisticHandler {
	return &StatisticHandler{svc: svc}
}

// ListByDateRange 按日期区间列出统计
// GET /api/v1/dh114/statistics/date-range  （公开）
func (h *StatisticHandler) ListByDateRange(ctx plugin.Context) {
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")
	statType := ctx.DefaultQuery("stat_type", model.StatTypeDaily)

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("start_date 格式应为 YYYY-MM-DD"))
		return
	}
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("end_date 格式应为 YYYY-MM-DD"))
		return
	}
	list, err := h.svc.ListByDateRange(startDate, endDate, statType)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114StatisticError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ListByDh114 按商户列出统计
// GET /api/v1/dh114/dh114/:id/statistics  （公开）
func (h *StatisticHandler) ListByDh114(ctx plugin.Context) {
	dh114ID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的商户ID"))
		return
	}
	startDate, endDate, ok := parseDateRange(ctx)
	if !ok {
		return
	}
	list, err := h.svc.ListByDh114(dh114ID, startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114StatisticError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ListByCategory 按分类列出统计
// GET /api/v1/dh114/statistics/by-category  （公开）
func (h *StatisticHandler) ListByCategory(ctx plugin.Context) {
	categoryID, err := parseUintStr(ctx.Query("category_id"))
	if err != nil || categoryID == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的分类ID"))
		return
	}
	startDate, endDate, ok := parseDateRange(ctx)
	if !ok {
		return
	}
	list, err := h.svc.ListByCategory(categoryID, startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114StatisticError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// SumByDh114 商户统计汇总
// GET /api/v1/dh114/dh114/:id/statistics/summary  （公开）
func (h *StatisticHandler) SumByDh114(ctx plugin.Context) {
	dh114ID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的商户ID"))
		return
	}
	startDate, endDate, ok := parseDateRange(ctx)
	if !ok {
		return
	}
	info, err := h.svc.SumByDh114(dh114ID, startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114StatisticError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// HotBusiness 热门商户
// GET /api/v1/dh114/statistics/hot-business  （公开）
func (h *StatisticHandler) HotBusiness(ctx plugin.Context) {
	regionID := getRegionID(ctx)
	limitStr := ctx.Query("limit")
	limit := 10
	if n, err := parseUintStr(limitStr); err == nil && n > 0 && n <= 100 {
		limit = int(n)
	}
	list, err := h.svc.HotBusiness(regionID, limit)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114StatisticError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// HotCategories 热门分类
// GET /api/v1/dh114/statistics/hot-categories  （公开）
func (h *StatisticHandler) HotCategories(ctx plugin.Context) {
	regionID := getRegionID(ctx)
	startDate, endDate, ok := parseDateRange(ctx)
	if !ok {
		return
	}
	list, err := h.svc.HotCategories(regionID, startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114StatisticError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// Overview 平台总览统计
// GET /api/v1/dh114/admin/statistics/overview  （需 dh114:audit 权限）
func (h *StatisticHandler) Overview(ctx plugin.Context) {
	regionID := getRegionID(ctx)
	startDate, endDate, ok := parseDateRange(ctx)
	if !ok {
		return
	}
	info, err := h.svc.Overview(regionID, startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114StatisticError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// AdminUpsert 写入/更新统计（M 端定时任务调用）
// POST /api/v1/dh114/admin/statistics  （需 dh114:audit 权限）
func (h *StatisticHandler) AdminUpsert(ctx plugin.Context) {
	regionID := getRegionID(ctx)
	var req model.Dh114Statistic
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Upsert(regionID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114StatisticError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("统计已写入", nil))
}

// parseDateRange 解析日期区间 query 参数（start_date/end_date，YYYY-MM-DD）
// 解析失败时写入响应并返回 ok=false
func parseDateRange(ctx plugin.Context) (time.Time, time.Time, bool) {
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("start_date 格式应为 YYYY-MM-DD"))
		return time.Time{}, time.Time{}, false
	}
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("end_date 格式应为 YYYY-MM-DD"))
		return time.Time{}, time.Time{}, false
	}
	return startDate, endDate, true
}
