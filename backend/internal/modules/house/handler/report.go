// Package handler 举报 + 评价 HTTP 处理层
// 依据 v3.2.1 架构方案第五章：对标贝壳/链家
// 聚合 ReportService + ReviewService 两个 service
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/house/dto"
	"wuchang-tongcheng/internal/modules/house/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ReportHandler 举报 + 评价聚合 HTTP 处理器
type ReportHandler struct {
	reportSvc service.ReportService
	reviewSvc service.ReviewService
}

// NewReportHandler 创建 Report Handler 实例
func NewReportHandler(reportSvc service.ReportService, reviewSvc service.ReviewService) *ReportHandler {
	return &ReportHandler{
		reportSvc: reportSvc,
		reviewSvc: reviewSvc,
	}
}

// ===== 举报 C 端 =====

// CreateReport C 端用户举报房源/房源发布/经纪人/小区/评价
// POST /api/v1/house/reports  （需登录）
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
	info, err := h.reportSvc.Create(regionID, userID, username, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("举报已提交", info))
}

// GetReport 举报详情（举报人本人或 M 端可查）
// GET /api/v1/house/reports/:id  （需登录）
func (h *ReportHandler) GetReport(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.reportSvc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetReportByNo 按举报单号查询
// GET /api/v1/house/reports/no/:no  （需登录）
func (h *ReportHandler) GetReportByNo(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	no := ctx.Param("no")
	if no == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("举报单号不能为空"))
		return
	}
	info, err := h.reportSvc.GetByNo(no)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListReports 举报列表（按 targetType/targetID 过滤）
// GET /api/v1/house/reports  （需登录）
func (h *ReportHandler) ListReports(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.ReportListQuery
	_ = ctx.Bind(&req)
	pagination, list, err := h.reportSvc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListMyReports 我的举报列表
// GET /api/v1/house/reports/mine  （需登录）
func (h *ReportHandler) ListMyReports(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.ReportListQuery
	_ = ctx.Bind(&req)
	pagination, list, err := h.reportSvc.ListMine(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListReportsByTarget 按目标查询举报列表（房源/经纪人/小区/评价关联举报）
// GET /api/v1/house/:target_type/:target_id/reports  （公开）
func (h *ReportHandler) ListReportsByTarget(ctx plugin.Context) {
	targetType := ctx.Param("target_type")
	targetID, err := parseSubID(ctx, "target_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的目标ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.reportSvc.ListByTarget(targetType, targetID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AppealReport 举报人申诉
// POST /api/v1/house/reports/:id/appeal  （需登录 + 举报人本人）
func (h *ReportHandler) AppealReport(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.ReportAppealRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.reportSvc.Appeal(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("申诉已提交", nil))
}

// ===== 举报 M 端 =====

// AdminListReports 管理后台举报列表
// GET /api/v1/admin/house/reports  （需 house:audit 权限）
func (h *ReportHandler) AdminListReports(ctx plugin.Context) {
	var req dto.ReportAdminListQuery
	_ = ctx.Bind(&req)
	pagination, list, err := h.reportSvc.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ProcessReport 处理举报
// PUT /api/v1/admin/house/reports/:id/process  （需 house:audit 权限）
func (h *ReportHandler) ProcessReport(ctx plugin.Context) {
	handlerID, handlerName, _, _ := getUserProfile(ctx)
	if handlerID == 0 {
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
	if err := h.reportSvc.Process(id, handlerID, handlerName, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("处理成功", nil))
}

// AppealHandleReport 申诉处理
// PUT /api/v1/admin/house/reports/:id/appeal  （需 house:audit 权限）
func (h *ReportHandler) AppealHandleReport(ctx plugin.Context) {
	handlerID, _, _, _ := getUserProfile(ctx)
	if handlerID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.ReportAppealHandleRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.reportSvc.AppealHandle(id, handlerID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("申诉已处理", nil))
}

// BatchUpdateReportStatus 批量更新举报状态
// POST /api/v1/admin/house/reports/batch  （需 house:audit 权限）
func (h *ReportHandler) BatchUpdateReportStatus(ctx plugin.Context) {
	var req dto.BatchStatusUpdateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	result, err := h.reportSvc.BatchUpdateStatus(req.IDs, req.Status)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(result))
}

// CountPendingReports 待处理举报数（M 端仪表盘）
// GET /api/v1/admin/house/reports/pending-count  （需 house:audit 权限）
func (h *ReportHandler) CountPendingReports(ctx plugin.Context) {
	count, err := h.reportSvc.CountPending()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseReportError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(map[string]int64{"count": count}))
}

// ===== 评价 C 端 =====

// CreateReview 创建评价（评价经纪人/小区/房源）
// POST /api/v1/house/reviews  （需登录）
func (h *ReportHandler) CreateReview(ctx plugin.Context) {
	userID, username, avatar, _ := getUserProfile(ctx)
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
	info, err := h.reviewSvc.Create(regionID, userID, username, avatar, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("评价成功", info))
}

// GetReview 评价详情
// GET /api/v1/house/reviews/:id  （公开）
func (h *ReportHandler) GetReview(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.reviewSvc.GetByID(id, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListReviews 评价列表
// GET /api/v1/house/reviews  （公开）
func (h *ReportHandler) ListReviews(ctx plugin.Context) {
	var req dto.ReviewListQuery
	_ = ctx.Bind(&req)
	pagination, list, err := h.reviewSvc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListReviewsByTarget 按目标查询评价列表
// GET /api/v1/house/:target_type/:target_id/reviews  （公开）
func (h *ReportHandler) ListReviewsByTarget(ctx plugin.Context) {
	targetType := ctx.Param("target_type")
	targetID, err := parseSubID(ctx, "target_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的目标ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.reviewSvc.ListByTarget(targetType, targetID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListMyReviews 我的评价列表
// GET /api/v1/house/reviews/mine  （需登录）
func (h *ReportHandler) ListMyReviews(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.ReviewListQuery
	_ = ctx.Bind(&req)
	pagination, list, err := h.reviewSvc.ListMine(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ReplyReview 回复评价（评价目标所有者可回复）
// POST /api/v1/house/reviews/:id/reply  （需登录）
func (h *ReportHandler) ReplyReview(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.ReviewReplyRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.reviewSvc.Reply(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("回复成功", nil))
}

// AppendReview 追评（评价人本人可追评）
// POST /api/v1/house/reviews/:id/append  （需登录 + 评价人本人）
func (h *ReportHandler) AppendReview(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.ReviewAppendRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.reviewSvc.Append(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("追评成功", nil))
}

// LikeReview 点赞评价
// POST /api/v1/house/reviews/:id/like  （需登录）
func (h *ReportHandler) LikeReview(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.reviewSvc.Like(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("点赞成功", nil))
}

// ReviewStats 评价统计（按目标统计评分分布）
// GET /api/v1/house/:target_type/:target_id/reviews/stats  （公开）
func (h *ReportHandler) ReviewStats(ctx plugin.Context) {
	targetType := ctx.Param("target_type")
	targetID, err := parseSubID(ctx, "target_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的目标ID"))
		return
	}
	stats, err := h.reviewSvc.GetStats(targetType, targetID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(stats))
}

// ===== 评价 M 端 =====

// AdminListReviews 管理后台评价列表
// GET /api/v1/admin/house/reviews  （需 house:audit 权限）
func (h *ReportHandler) AdminListReviews(ctx plugin.Context) {
	var req dto.ReviewAdminListQuery
	_ = ctx.Bind(&req)
	pagination, list, err := h.reviewSvc.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// UpdateReviewStatus 更新评价状态（显示/隐藏/被举报）
// PUT /api/v1/admin/house/reviews/:id/status  （需 house:audit 权限）
func (h *ReportHandler) UpdateReviewStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateHouseStatusRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.reviewSvc.UpdateStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态更新成功", nil))
}

// BatchUpdateReviewStatus 批量更新评价状态
// POST /api/v1/admin/house/reviews/batch  （需 house:audit 权限）
func (h *ReportHandler) BatchUpdateReviewStatus(ctx plugin.Context) {
	var req dto.BatchStatusUpdateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	result, err := h.reviewSvc.BatchUpdateStatus(req.IDs, req.Status)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(result))
}

// DeleteReview 删除评价
// DELETE /api/v1/admin/house/reviews/:id  （需 house:audit 权限）
func (h *ReportHandler) DeleteReview(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.reviewSvc.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}
