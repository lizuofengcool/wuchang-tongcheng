// Package handler 举报/评价/审核规则/用户信用 HTTP 处理层
// 依据 v3.2.1 架构方案：对标转转风控中心
package handler

import (
	"net/http"
	"strconv"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/ershou/dto"
	"wuchang-tongcheng/internal/modules/ershou/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// RiskHandler 风控聚合 HTTP 处理器（聚合 Report/Review/AuditRule/UserCredit 4 个 service）
type RiskHandler struct {
	reportSvc    service.ReportService
	reviewSvc    service.ReviewService
	auditRuleSvc service.AuditRuleService
	creditSvc    service.UserCreditService
}

// NewRiskHandler 创建 Risk Handler 实例
func NewRiskHandler(
	reportSvc service.ReportService,
	reviewSvc service.ReviewService,
	auditRuleSvc service.AuditRuleService,
	creditSvc service.UserCreditService,
) *RiskHandler {
	return &RiskHandler{
		reportSvc:    reportSvc,
		reviewSvc:    reviewSvc,
		auditRuleSvc: auditRuleSvc,
		creditSvc:    creditSvc,
	}
}

// ===== 举报 =====

// CreateReport C 端用户举报物品
// POST /api/v1/ershou/reports  （需登录）
func (h *RiskHandler) CreateReport(ctx plugin.Context) {
	userID, username, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.ReportCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.reportSvc.Create(userID, username, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("举报已提交", resp))
}

// GetReport 举报详情（M 端）
// GET /api/v1/ershou/reports/:id  （需登录 + content:audit）
func (h *RiskHandler) GetReport(ctx plugin.Context) {
	reportID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的举报ID"))
		return
	}
	resp, err := h.reportSvc.GetByID(reportID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// ListReports 举报列表（M 端）
// GET /api/v1/ershou/reports  （需登录 + content:audit）
func (h *RiskHandler) ListReports(ctx plugin.Context) {
	var query dto.ReportListQuery
	_ = ctx.Bind(&query)
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 10
	}
	pagination, list, err := h.reportSvc.List(query)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListReportsByErshouID 物品关联的举报列表
// GET /api/v1/ershou/:id/reports  （公开）
func (h *RiskHandler) ListReportsByErshouID(ctx plugin.Context) {
	ershouID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	list, err := h.reportSvc.ListByErshouID(ershouID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ProcessReport 处理举报（M 端审核员）
// PUT /api/v1/ershou/reports/:id/process  （需登录 + content:audit）
func (h *RiskHandler) ProcessReport(ctx plugin.Context) {
	userID, username, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	reportID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的举报ID"))
		return
	}
	var req dto.ReportProcessRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.reportSvc.Process(reportID, userID, username, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("举报处理成功", resp))
}

// ===== 评价 =====

// CreateReview 买家评价订单
// POST /api/v1/ershou/orders/:id/reviews  （需登录 + 买家）
func (h *RiskHandler) CreateReview(ctx plugin.Context) {
	userID, username, _, avatar := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	orderID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的订单ID"))
		return
	}
	var req dto.ReviewCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.reviewSvc.Create(orderID, userID, username, avatar, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("评价成功", resp))
}

// GetReview 评价详情
// GET /api/v1/ershou/reviews/:id  （公开）
func (h *RiskHandler) GetReview(ctx plugin.Context) {
	reviewID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的评价ID"))
		return
	}
	resp, err := h.reviewSvc.GetByID(reviewID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// ListReviews 评价列表（M 端 / C 端按商品/用户筛选）
// GET /api/v1/ershou/reviews  （公开）
func (h *RiskHandler) ListReviews(ctx plugin.Context) {
	var query dto.ReviewListQuery
	_ = ctx.Bind(&query)
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 10
	}
	pagination, list, err := h.reviewSvc.List(query)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListReviewsByErshouID 商品评价列表
// GET /api/v1/ershou/:id/reviews  （公开）
func (h *RiskHandler) ListReviewsByErshouID(ctx plugin.Context) {
	ershouID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination := utils.NewPagination(page, pageSize)
	pagination, list, err := h.reviewSvc.ListByErshouID(ershouID, pagination)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ReplyReview 卖家回复评价
// POST /api/v1/ershou/reviews/:id/reply  （需登录 + 被评价人）
func (h *RiskHandler) ReplyReview(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	reviewID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的评价ID"))
		return
	}
	var req dto.ReviewReplyRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.reviewSvc.Reply(reviewID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("回复成功", resp))
}

// ReviewStats 商品评价统计
// GET /api/v1/ershou/:id/reviews/stats  （公开）
func (h *RiskHandler) ReviewStats(ctx plugin.Context) {
	ershouID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	resp, err := h.reviewSvc.StatsByErshouID(ershouID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// ===== 审核规则 =====

// CreateAuditRule 创建审核规则（M 端）
// POST /api/v1/ershou/audit-rules  （需登录 + content:audit）
func (h *RiskHandler) CreateAuditRule(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.AuditRuleCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.auditRuleSvc.Create(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouAuditError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核规则创建成功", resp))
}

// GetAuditRule 审核规则详情
// GET /api/v1/ershou/audit-rules/:id  （需登录 + content:audit）
func (h *RiskHandler) GetAuditRule(ctx plugin.Context) {
	id, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的规则ID"))
		return
	}
	resp, err := h.auditRuleSvc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// UpdateAuditRule 更新审核规则
// PUT /api/v1/ershou/audit-rules/:id  （需登录 + content:audit）
func (h *RiskHandler) UpdateAuditRule(ctx plugin.Context) {
	id, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的规则ID"))
		return
	}
	var req dto.AuditRuleCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.auditRuleSvc.Update(id, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouAuditError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核规则更新成功", resp))
}

// DeleteAuditRule 删除审核规则
// DELETE /api/v1/ershou/audit-rules/:id  （需登录 + content:audit）
func (h *RiskHandler) DeleteAuditRule(ctx plugin.Context) {
	id, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的规则ID"))
		return
	}
	if err := h.auditRuleSvc.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouAuditError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核规则已删除", nil))
}

// ListAuditRules 审核规则列表
// GET /api/v1/ershou/audit-rules  （需登录 + content:audit）
func (h *RiskHandler) ListAuditRules(ctx plugin.Context) {
	ruleType := ctx.Query("rule_type")
	page, pageSize := parsePagination(ctx)
	statusStr := ctx.Query("status")
	var statusPtr *int
	if statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			statusPtr = &s
		}
	}
	pagination, list, err := h.auditRuleSvc.List(ruleType, statusPtr, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListEnabledAuditRules 启用中的审核规则列表（C 端发布时使用）
// GET /api/v1/ershou/audit-rules/enabled  （公开）
func (h *RiskHandler) ListEnabledAuditRules(ctx plugin.Context) {
	list, err := h.auditRuleSvc.ListEnabled()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ===== 用户信用 =====

// GetUserCredit 查询用户信用（C 端 / M 端）
// GET /api/v1/ershou/credit  （需登录）
func (h *RiskHandler) GetUserCredit(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	resp, err := h.creditSvc.GetByUserID(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// GetUserCreditByID 查询指定用户的信用（M 端）
// GET /api/v1/ershou/credit/:user_id  （需登录 + content:audit）
func (h *RiskHandler) GetUserCreditByID(ctx plugin.Context) {
	targetUserID, err := parseSubID(ctx, "user_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的用户ID"))
		return
	}
	resp, err := h.creditSvc.GetByUserID(targetUserID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// UpdateUserCredit 更新用户信用（M 端，手动调整）
// PUT /api/v1/ershou/credit/:user_id  （需登录 + content:audit）
func (h *RiskHandler) UpdateUserCredit(ctx plugin.Context) {
	targetUserID, err := parseSubID(ctx, "user_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的用户ID"))
		return
	}
	var req dto.UserCreditUpdateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.creditSvc.Update(targetUserID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("信用更新成功", resp))
}
