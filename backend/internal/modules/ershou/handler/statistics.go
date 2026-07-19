// Package handler 数据统计 HTTP 处理层
// 依据 v3.2.1 架构方案：M 端运营数据 + C 端卖家数据
package handler

import (
	"net/http"
	"strconv"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/ershou/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// StatisticsHandler 数据统计 HTTP 处理器
type StatisticsHandler struct {
	service service.StatisticsService
}

// NewStatisticsHandler 创建统计 Handler 实例
func NewStatisticsHandler(svc service.StatisticsService) *StatisticsHandler {
	return &StatisticsHandler{service: svc}
}

// Overview 平台总览（M 端运营数据）
// GET /api/v1/ershou/statistics/overview  （需登录 + content:audit）
func (h *StatisticsHandler) Overview(ctx plugin.Context) {
	regionID := getRegionID(ctx)
	resp, err := h.service.Overview(regionID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// SellerOverview 卖家数据总览（C 端）
// GET /api/v1/ershou/statistics/seller  （需登录）
func (h *StatisticsHandler) SellerOverview(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	resp, err := h.service.SellerOverview(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// HotItems 热门商品 Top N
// GET /api/v1/ershou/statistics/hot-items  （公开）
func (h *StatisticsHandler) HotItems(ctx plugin.Context) {
	limitStr := ctx.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)
	list, err := h.service.HotItems(limit)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// PriceTrend 价格趋势
// GET /api/v1/ershou/statistics/price-trend  （公开）
func (h *StatisticsHandler) PriceTrend(ctx plugin.Context) {
	brandIDStr := ctx.Query("brand_id")
	brandID, _ := strconv.ParseUint(brandIDStr, 10, 32)
	daysStr := ctx.DefaultQuery("days", "30")
	days, _ := strconv.Atoi(daysStr)
	resp, err := h.service.PriceTrend(uint(brandID), days)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}
