// Package handler LBS地图中台 HTTP 处理层 - 公共能力（距离计算 + 路线规划）
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/lbs/dto"
	"wuchang-tongcheng/internal/modules/lbs/service"
)

// LBSHandler LBS 公共处理器（距离/路线规划）
type LBSHandler struct {
	routeSvc service.RouteService
}

// NewLBSHandler 创建 LBS 公共 Handler 实例
func NewLBSHandler(routeSvc service.RouteService) *LBSHandler {
	return &LBSHandler{routeSvc: routeSvc}
}

// CalculateDistance 计算两点间直线距离
// GET /api/v1/lbs/distance?from_lat=&from_lng=&to_lat=&to_lng=  （公开）
func (h *LBSHandler) CalculateDistance(ctx plugin.Context) {
	var req dto.DistanceRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if req.FromLat == 0 || req.FromLng == 0 || req.ToLat == 0 || req.ToLng == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("经纬度不能为空"))
		return
	}
	resp, err := h.routeSvc.CalculateDistance(&req)
	if err != nil {
		failByError(ctx, CodeLBSDistanceError, err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// PlanRoute 路线规划
// GET /api/v1/lbs/route?from_lat=&from_lng=&to_lat=&to_lng=&mode=  （公开）
func (h *LBSHandler) PlanRoute(ctx plugin.Context) {
	var req dto.RouteRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if req.FromLat == 0 || req.FromLng == 0 || req.ToLat == 0 || req.ToLng == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("经纬度不能为空"))
		return
	}
	resp, err := h.routeSvc.PlanRoute(&req)
	if err != nil {
		failByError(ctx, CodeLBSRouteError, err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}
