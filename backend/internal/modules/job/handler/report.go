// Package handler 举报 + 评价 HTTP 处理层
// 依据 v3.2.1 架构方案第四章：对标 BOSS直聘/看准
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/job/dto"
	"wuchang-tongcheng/internal/modules/job/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ReportHandler 举报 + 评价 HTTP 处理器
type ReportHandler struct {
	svc service.ReportService
}

// NewReportHandler 创建 ReportHandler 实例
func NewReportHandler(svc service.ReportService) *ReportHandler {
	return &ReportHandler{svc: svc}
}

// ===== 举报 =====

// CreateReport C 端用户举报（职位/公司/简历/招聘者/评价）
// POST /api/v1/job/reports  （需登录）
func (h *ReportHandler) CreateReport(ctx plugin.Context) {
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
	regionID := getRegionID(ctx)
	resp, err := h.svc.CreateReport(regionID, userID, username, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("举报已提交", resp))
}

// GetReport 举报详情
// GET /api/v1/job/reports/:id  （需登录）
func (h *ReportHandler) GetReport(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的举报ID"))
		return
	}
	resp, err := h.svc.GetReport(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// ListReports 举报列表（M 端 / C 端按目标筛选）
// GET /api/v1/job/reports  （需登录 + job:audit）
func (h *ReportHandler) ListReports(ctx plugin.Context) {
	var query dto.ReportListQuery
	_ = ctx.Bind(&query)
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 10
	}
	pagination, list, err := h.svc.ListReports(&query)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListMyReports 我的举报列表
// GET /api/v1/job/reports/mine  （需登录）
func (h *ReportHandler) ListMyReports(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var query dto.ReportListQuery
	_ = ctx.Bind(&query)
	query.ReporterID = userID
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 10
	}
	pagination, list, err := h.svc.ListReports(&query)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListReportsByTarget 目标关联的举报列表
// GET /api/v1/job/reports/by-target  （公开）
func (h *ReportHandler) ListReportsByTarget(ctx plugin.Context) {
	targetType := ctx.Query("target_type")
	if targetType == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("target_type 不能为空"))
		return
	}
	targetID, err := parseSubID(ctx, "target_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的 target_id"))
		return
	}
	list, err := h.svc.ListReportsByTarget(targetType, targetID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// HandleReport 处理举报（M 端审核员）
// PUT /api/v1/job/reports/:id/process  （需登录 + job:audit）
func (h *ReportHandler) HandleReport(ctx plugin.Context) {
	userID, username, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的举报ID"))
		return
	}
	var req dto.ReportProcessRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.svc.ProcessReport(id, userID, username, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("举报处理成功", resp))
}

// AppealReport 举报申诉（举报人发起）
// POST /api/v1/job/reports/:id/appeal  （需登录 + 举报人本人）
func (h *ReportHandler) AppealReport(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的举报ID"))
		return
	}
	var req dto.ReportAppealRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.svc.AppealReport(id, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("申诉已提交", resp))
}

// ProcessAppeal 处理申诉（M 端审核员）
// PUT /api/v1/job/reports/:id/appeal  （需登录 + job:audit）
func (h *ReportHandler) ProcessAppeal(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的举报ID"))
		return
	}
	var req dto.ReportAppealProcessRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.svc.ProcessAppeal(id, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("申诉处理成功", resp))
}

// ===== 评价 =====

// CreateReview 创建公司评价
// POST /api/v1/job/reviews  （需登录）
func (h *ReportHandler) CreateReview(ctx plugin.Context) {
	userID, username, _, avatar := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.ReviewCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	resp, err := h.svc.CreateReview(regionID, userID, username, avatar, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("评价成功", resp))
}

// GetReview 评价详情
// GET /api/v1/job/reviews/:id  （公开）
func (h *ReportHandler) GetReview(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的评价ID"))
		return
	}
	resp, err := h.svc.GetReview(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// ListReviews 评价列表（M 端 / C 端按公司/用户筛选）
// GET /api/v1/job/reviews  （公开）
func (h *ReportHandler) ListReviews(ctx plugin.Context) {
	var query dto.ReviewListQuery
	_ = ctx.Bind(&query)
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 10
	}
	pagination, list, err := h.svc.ListReviews(&query)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListMyReviews 我的评价列表
// GET /api/v1/job/reviews/mine  （需登录）
func (h *ReportHandler) ListMyReviews(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var query dto.ReviewListQuery
	_ = ctx.Bind(&query)
	query.ReviewerID = userID
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 10
	}
	pagination, list, err := h.svc.ListReviews(&query)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListReviewsByUser 指定用户的评价列表
// GET /api/v1/job/users/:user_id/reviews  （公开）
func (h *ReportHandler) ListReviewsByUser(ctx plugin.Context) {
	targetUserID, err := parseSubID(ctx, "user_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的用户ID"))
		return
	}
	var query dto.ReviewListQuery
	_ = ctx.Bind(&query)
	query.ReviewerID = targetUserID
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 10
	}
	pagination, list, err := h.svc.ListReviews(&query)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListReviewsByCompany 公司评价列表
// GET /api/v1/job/companies/:id/reviews  （公开）
func (h *ReportHandler) ListReviewsByCompany(ctx plugin.Context) {
	companyID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的公司ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListReviewsByCompany(companyID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ReviewStats 公司评价统计
// GET /api/v1/job/companies/:id/reviews/stats  （公开）
func (h *ReportHandler) ReviewStats(ctx plugin.Context) {
	companyID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的公司ID"))
		return
	}
	resp, err := h.svc.ReviewStats(companyID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// UpdateReview 更新评价（仅评价人本人）
// PUT /api/v1/job/reviews/:id  （需登录 + 评价人本人）
func (h *ReportHandler) UpdateReview(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的评价ID"))
		return
	}
	var req dto.ReviewUpdateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.UpdateReview(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// ReplyReview 公司回复评价
// POST /api/v1/job/reviews/:id/reply  （需登录）
func (h *ReportHandler) ReplyReview(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的评价ID"))
		return
	}
	var req dto.ReviewReplyRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.ReplyReview(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("回复成功", nil))
}

// AppendReview 追评（仅评价人本人）
// POST /api/v1/job/reviews/:id/append  （需登录 + 评价人本人）
func (h *ReportHandler) AppendReview(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的评价ID"))
		return
	}
	var req dto.ReviewAppendRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.AppendReview(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("追评成功", nil))
}

// DeleteReview 删除评价（仅评价人本人）
// DELETE /api/v1/job/reviews/:id  （需登录 + 评价人本人）
func (h *ReportHandler) DeleteReview(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的评价ID"))
		return
	}
	if err := h.svc.DeleteReview(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// LikeReview 点赞评价
// POST /api/v1/job/reviews/:id/like  （公开）
func (h *ReportHandler) LikeReview(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的评价ID"))
		return
	}
	if err := h.svc.LikeReview(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("点赞成功", nil))
}

// AuditReview M 端审核评价
// PUT /api/v1/job/reviews/:id/audit  （需登录 + job:audit）
func (h *ReportHandler) AuditReview(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的评价ID"))
		return
	}
	var req dto.AuditRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.AuditReview(id, req.AuditStatus, req.AuditReason); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核完成", nil))
}
