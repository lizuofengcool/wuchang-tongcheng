// Package handler 数据统计 HTTP 处理层
// 依据 v3.2.1 架构方案：M 端运营数据 + C 端招聘者/求职者数据
package handler

import (
	"net/http"
	"strconv"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/job/dto"
	"wuchang-tongcheng/internal/modules/job/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// StatisticsHandler 数据统计 HTTP 处理器
type StatisticsHandler struct {
	svc service.StatisticsService
}

// NewStatisticsHandler 创建 StatisticsHandler 实例
func NewStatisticsHandler(svc service.StatisticsService) *StatisticsHandler {
	return &StatisticsHandler{svc: svc}
}

// GetOverview M 端运营总览
// GET /api/v1/job/admin/statistics/overview  （需登录 + job:audit）
func (h *StatisticsHandler) GetOverview(ctx plugin.Context) {
	regionID := getRegionID(ctx)
	resp, err := h.svc.Overview(regionID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeJobError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// GetJobStats 职位趋势统计
// GET /api/v1/job/statistics/job-trend  （公开）
func (h *StatisticsHandler) GetJobStats(ctx plugin.Context) {
	regionID := getRegionID(ctx)
	daysStr := ctx.DefaultQuery("days", "30")
	days, _ := strconv.Atoi(daysStr)
	resp, err := h.svc.JobTrend(regionID, days)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeJobError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// GetCompanyStats 公司分类统计
// GET /api/v1/job/statistics/category  （公开）
func (h *StatisticsHandler) GetCompanyStats(ctx plugin.Context) {
	regionID := getRegionID(ctx)
	list, err := h.svc.CategoryStats(regionID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeJobError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// GetUserStats C 端招聘者数据总览
// GET /api/v1/job/statistics/recruiter  （需登录）
func (h *StatisticsHandler) GetUserStats(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	resp, err := h.svc.RecruiterOverview(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeJobError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// GetApplicantStats C 端求职者数据总览
// GET /api/v1/job/statistics/applicant  （需登录）
func (h *StatisticsHandler) GetApplicantStats(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	resp, err := h.svc.ApplicantOverview(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeJobError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// GetTradeStats 薪资趋势统计
// GET /api/v1/job/statistics/salary-trend  （公开）
func (h *StatisticsHandler) GetTradeStats(ctx plugin.Context) {
	categoryIDStr := ctx.Query("category_id")
	categoryID, _ := strconv.ParseUint(categoryIDStr, 10, 32)
	daysStr := ctx.DefaultQuery("days", "30")
	days, _ := strconv.Atoi(daysStr)
	resp, err := h.svc.SalaryTrend(uint(categoryID), days)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeJobError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// GetFunnelStats 转化漏斗统计
// GET /api/v1/job/statistics/conversion  （需登录 + job:audit）
func (h *StatisticsHandler) GetFunnelStats(ctx plugin.Context) {
	regionID := getRegionID(ctx)
	var req dto.DateTrendRequest
	_ = ctx.Bind(&req)
	req.RegionID = regionID
	resp, err := h.svc.ConversionStats(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeJobError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// GetTrendStats 地区统计
// GET /api/v1/job/statistics/region  （公开）
func (h *StatisticsHandler) GetTrendStats(ctx plugin.Context) {
	list, err := h.svc.RegionStats()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeJobError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ExportReport 热门职位导出（MVP 返回 Top N 列表，前端自行转 Excel/CSV）
// GET /api/v1/job/statistics/hot-jobs  （公开）
func (h *StatisticsHandler) ExportReport(ctx plugin.Context) {
	regionID := getRegionID(ctx)
	limitStr := ctx.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)
	list, err := h.svc.HotJobs(regionID, limit)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeJobError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// GetDashboard C 端卖家/招聘者仪表盘（聚合：招聘者总览 + 热门职位）
// GET /api/v1/job/statistics/dashboard  （需登录）
func (h *StatisticsHandler) GetDashboard(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	overview, err := h.svc.RecruiterOverview(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeJobError, err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	hotJobs, _ := h.svc.HotJobs(regionID, 5)
	ctx.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"overview": overview,
		"hot_jobs": hotJobs,
	}))
}
