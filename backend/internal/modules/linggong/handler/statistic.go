// Package handler 同城零工兼职 HTTP 处理层 - 总览统计（M 端）
// 依据需求：GET /api/v1/linggong/admin/statistics/overview
// 跨表聚合 linggongs/employers/workers/applications/contracts/payments/skills
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/linggong/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// StatisticHandler 总览统计 HTTP 处理器（M 端，需 linggong:audit 权限）
type StatisticHandler struct {
	service service.OverviewService
}

// NewStatisticHandler 创建 StatisticHandler 实例
func NewStatisticHandler(svc service.OverviewService) *StatisticHandler {
	return &StatisticHandler{service: svc}
}

// Overview 总览统计
// GET /api/v1/linggong/admin/statistics/overview  （需 linggong:audit 权限）
// 返回字段：total_linggongs/total_employers/total_workers/total_applications/
//
//	total_contracts/total_payments/pending_audits/today_new/week_new/month_new/
//	hot_skills[]/top_employers[]
func (h *StatisticHandler) Overview(ctx plugin.Context) {
	stats, err := h.service.Overview()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(stats))
}
