// Package handler 同城拼车出行 HTTP 处理层 - 统计
// 依据 v3.2.1 架构方案：日统计/周统计/月统计/总统计
// 对标哈啰出行/嘀嗒出行 数据分析
package handler

import (
	"net/http"
	"time"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/pinche/dto"
	"wuchang-tongcheng/internal/modules/pinche/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// StatisticsHandler 统计 HTTP 处理器
type StatisticsHandler struct {
	service service.StatisticService
}

// NewStatisticsHandler 创建 StatisticsHandler 实例
func NewStatisticsHandler(svc service.StatisticService) *StatisticsHandler {
	return &StatisticsHandler{service: svc}
}

// ===== C 端 =====

// GetByID 统计详情
// GET /api/v1/pinche/statistics/:id  （公开）
func (h *StatisticsHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 统计列表
// GET /api/v1/pinche/statistics  （公开）
func (h *StatisticsHandler) List(ctx plugin.Context) {
	var req dto.StatisticListRequest
	_ = ctx.Bind(&req)

	regionID := getRegionID(ctx)
	pagination, list, err := h.service.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByUser 我的统计
// GET /api/v1/pinche/statistics/mine  （需登录）
func (h *StatisticsHandler) ListByUser(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListByUser(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByDateRange 按日期区间查询
// GET /api/v1/pinche/statistics/date-range  （公开）
func (h *StatisticsHandler) ListByDateRange(ctx plugin.Context) {
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")
	statType := ctx.Query("stat_type")
	if startDateStr == "" || endDateStr == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("start_date 和 end_date 必填"))
		return
	}
	startDate, err := time.ParseInLocation("2006-01-02", startDateStr, time.Local)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("start_date 格式无效"))
		return
	}
	endDate, err := time.ParseInLocation("2006-01-02", endDateStr, time.Local)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("end_date 格式无效"))
		return
	}
	regionID := getRegionID(ctx)
	list, err := h.service.ListByDateRange(regionID, startDate, endDate, statType)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// Overview 平台总览统计
// GET /api/v1/pinche/statistics/overview  （公开）
func (h *StatisticsHandler) Overview(ctx plugin.Context) {
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")
	if startDateStr == "" || endDateStr == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("start_date 和 end_date 必填"))
		return
	}
	startDate, err := time.ParseInLocation("2006-01-02", startDateStr, time.Local)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("start_date 格式无效"))
		return
	}
	endDate, err := time.ParseInLocation("2006-01-02", endDateStr, time.Local)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("end_date 格式无效"))
		return
	}
	regionID := getRegionID(ctx)
	resp, err := h.service.Overview(regionID, startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// ===== M 端管理 =====

// AdminList 管理后台统计列表（跨地区）
// GET /api/v1/pinche/admin/statistics  （需 pinche:audit 权限）
func (h *StatisticsHandler) AdminList(ctx plugin.Context) {
	var req dto.StatisticListRequest
	_ = ctx.Bind(&req)

	// 跨地区：regionID=0
	pagination, list, err := h.service.List(0, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Upsert 创建/更新统计（M 端定时任务调用）
// POST /api/v1/pinche/admin/statistics  （需 pinche:audit 权限）
func (h *StatisticsHandler) Upsert(ctx plugin.Context) {
	var req dto.StatisticUpsertRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.service.Upsert(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("统计已更新", info))
}

// Delete 删除统计记录
// DELETE /api/v1/pinche/admin/statistics/:id  （需 pinche:audit 权限）
func (h *StatisticsHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.service.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}
