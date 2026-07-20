// Package handler 同城零工兼职 HTTP 处理层 - 举报管理（M 端）
// 依据需求：复用 linggong_disputes 表实现举报管理
// 前端字段对齐 reports.vue：report_no/reporter/reported_user/target/report_type/reason/handle_result
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/linggong/dto"
	"wuchang-tongcheng/internal/modules/linggong/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ReportHandler 举报管理 HTTP 处理器（M 端，需 linggong:audit 权限）
type ReportHandler struct {
	service service.DisputeService
}

// NewReportHandler 创建 ReportHandler 实例
func NewReportHandler(svc service.DisputeService) *ReportHandler {
	return &ReportHandler{service: svc}
}

// reportStatusText 举报状态文本（对齐前端 statusMap）
// 0待处理 1已核实警告 2已下架 3已封号 4已驳回 5已转交
func reportStatusText(s int) string {
	switch s {
	case 0:
		return "待处理"
	case 1:
		return "已核实警告"
	case 2:
		return "已下架"
	case 3:
		return "已封号"
	case 4:
		return "已驳回"
	case 5:
		return "已转交"
	}
	return ""
}

// reportTypeText 举报类型文本（对齐前端 typeMap）
// 复用 disputes.dispute_type 字段，但语义采用举报类型
func reportTypeText(t string) string {
	switch t {
	case "fraud":
		return "欺诈"
	case "fake":
		return "虚假信息"
	case "porn":
		return "色情低俗"
	case "spam":
		return "垃圾信息"
	case "infringement":
		return "侵权"
	case "prohibited":
		return "违禁"
	case "harassment":
		return "骚扰"
	case "default_payment":
		return "拖欠薪资"
	case "breach":
		return "违约"
	case "other":
		return "其他"
	}
	return ""
}

// toReportInfo 将 DisputeInfo 转为 ReportInfo（前端字段对齐）
func toReportInfo(d *dto.DisputeInfo) dto.ReportInfo {
	return dto.ReportInfo{
		ID:               d.ID,
		ReportNo:         d.DisputeNo,
		LinggongID:       d.LinggongID,
		TaskID:           d.TaskID,
		ApplicationID:    d.ApplicationID,
		ContractID:       d.ContractID,
		PaymentID:        d.PaymentID,
		TargetType:       "linggong", // 举报目标类型默认 linggong（disputes 表无此字段）
		TargetID:         d.LinggongID,
		ReportType:       d.DisputeType,
		ReportTypeText:   reportTypeText(d.DisputeType),
		ReporterID:       d.ApplicantID,
		ReporterName:     d.ApplicantName,
		ReporterPhone:    "", // disputes 表无举报人电话字段
		ReportedUserID:   d.RespondentID,
		ReportedUserName: d.RespondentName,
		ReportedPhone:    "", // disputes 表无被举报人电话字段
		Reason:           d.Title,
		Description:      d.Description,
		EvidenceImages:   d.EvidenceImages,
		EvidenceVideos:   d.EvidenceVideos,
		EvidenceDocs:     d.EvidenceDocs,
		ClaimAmount:      d.ClaimAmount,
		Status:           d.Status,
		StatusText:       reportStatusText(d.Status),
		HandlerID:        d.HandlerID,
		HandlerName:      d.HandlerName,
		HandleResult:     d.MediationResult,
		SLADeadline:      d.SLADeadline,
		HandledAt:        d.HandledAt,
		ResolvedAt:       d.ResolvedAt,
		ClosedAt:         d.ClosedAt,
		RegionID:         d.RegionID,
		CreatedAt:        d.CreatedAt,
		UpdatedAt:        d.UpdatedAt,
	}
}

// ===== M 端管理 =====

// AdminList 举报列表
// GET /api/v1/linggong/admin/reports  （需 linggong:audit 权限）
// 返回 {list, total, stats:{total, pending, processed}, page, page_size}
func (h *ReportHandler) AdminList(ctx plugin.Context) {
	var req dto.ReportListRequest
	_ = ctx.Bind(&req)

	// 转换为 DisputeListRequest（复用 dispute service）
	disputeReq := dto.DisputeListRequest{
		DisputeType: req.ReportType,
		Status:      req.Status,
		Keyword:     req.Keyword,
		Pagination: utils.Pagination{
			Page:     req.Page,
			PageSize: req.PageSize,
		},
	}
	// target_type 过滤 disputes 表不支持，由前端过滤（disputes 表无 target_type 字段）

	pagination, list, err := h.service.AdminList(&disputeReq)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}

	// 转换为 ReportInfo（前端字段名）
	reports := make([]dto.ReportInfo, 0, len(list))
	for i := range list {
		reports = append(reports, toReportInfo(&list[i]))
	}

	// 举报统计（全局，不限当前过滤条件）
	stats, err := h.service.GetReportStats()
	if err != nil {
		// 统计失败不影响列表展示，降级返回零值
		stats = &dto.ReportStats{}
	}

	ctx.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"list":      reports,
		"total":     pagination.Total,
		"stats":     stats,
		"page":      pagination.Page,
		"page_size": pagination.PageSize,
	}))
}

// Process 处理举报
// PUT /api/v1/linggong/admin/reports/:id/process  （需 linggong:audit 权限）
// body: {status, handle_result, handle_note, result, penalty_type, credit_change}
func (h *ReportHandler) Process(ctx plugin.Context) {
	// 处理人信息（从 JWT 解析）
	userID, username, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	var req dto.ReportProcessRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	// 处理人姓名优先用 username，回退 "管理员"
	handlerName := username
	if handlerName == "" {
		handlerName = "管理员"
	}

	if err := h.service.ProcessReport(id, userID, handlerName, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("处理成功", nil))
}
