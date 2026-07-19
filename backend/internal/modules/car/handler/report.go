// Package handler 同城车辆买卖 HTTP 处理层 - 举报 + 评价
// 依据 v3.2.1 架构方案：对标瓜子风控中心/人人车评价/懂车帝评论
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package handler

import (
	"net/http"
	"strconv"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/car/dto"
	"wuchang-tongcheng/internal/modules/car/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// strToUint 将字符串转换为 uint（report.go 内部使用）
func strToUint(s string) (uint, error) {
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}

// ReportHandler 风控聚合 HTTP 处理器（聚合 Report + Review 2 个 service）
type ReportHandler struct {
	reportSvc service.ReportService
	reviewSvc service.ReviewService
}

// NewReportHandler 创建 ReportHandler 实例
func NewReportHandler(reportSvc service.ReportService, reviewSvc service.ReviewService) *ReportHandler {
	return &ReportHandler{
		reportSvc: reportSvc,
		reviewSvc: reviewSvc,
	}
}

// ===== 举报 =====

// CreateReport C 端用户举报
// POST /api/v1/car/reports  （需登录）
func (h *ReportHandler) CreateReport(ctx plugin.Context) {
	userID, username, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.CreateReportRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	info, err := h.reportSvc.Create(userID, username, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReportErrorCar, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("举报已提交", info))
}

// GetReport 举报详情（M 端）
// GET /api/v1/car/admin/reports/:id  （需 car:audit 权限）
func (h *ReportHandler) GetReport(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的举报ID"))
		return
	}

	info, err := h.reportSvc.AdminGetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReportErrorCar, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListReports 举报列表（M 端）
// GET /api/v1/car/admin/reports  （需 car:audit 权限）
func (h *ReportHandler) ListReports(ctx plugin.Context) {
	var req dto.ReportListRequest
	_ = ctx.Bind(&req)

	pagination, list, err := h.reportSvc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReportErrorCar, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListMyReports 我的举报列表
// GET /api/v1/car/reports/mine  （需登录）
func (h *ReportHandler) ListMyReports(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.reportSvc.ListByReporter(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReportErrorCar, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListReportsByTarget 按目标对象查询举报
// GET /api/v1/car/reports/by-target  （公开）
func (h *ReportHandler) ListReportsByTarget(ctx plugin.Context) {
	targetType := ctx.Query("target_type")
	targetIDStr := ctx.Query("target_id")
	if targetType == "" || targetIDStr == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("target_type 和 target_id 不能为空"))
		return
	}

	targetID, err := strToUint(targetIDStr)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的 target_id"))
		return
	}

	list, err := h.reportSvc.ListByTarget(targetType, targetID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReportErrorCar, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// AppealReport 举报申诉
// POST /api/v1/car/reports/:id/appeal  （需登录）
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

	var req dto.AppealReportRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.reportSvc.Appeal(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReportErrorCar, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("申诉已提交", nil))
}

// ProcessReport 处理举报（M 端）
// PUT /api/v1/car/admin/reports/:id/process  （需 car:audit 权限）
func (h *ReportHandler) ProcessReport(ctx plugin.Context) {
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

	var req dto.ProcessReportRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.reportSvc.Process(id, userID, username, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReportErrorCar, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("举报处理成功", nil))
}

// ProcessAppeal 处理申诉（M 端）
// PUT /api/v1/car/admin/reports/:id/appeal  （需 car:audit 权限）
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

	var req dto.ProcessAppealRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.reportSvc.ProcessAppeal(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReportErrorCar, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("申诉处理成功", nil))
}

// UpdateReportStatus 更新举报状态（M 端）
// PUT /api/v1/car/admin/reports/:id/status  （需 car:audit 权限）
func (h *ReportHandler) UpdateReportStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的举报ID"))
		return
	}

	var req dto.AdminUpdateStatusRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.reportSvc.UpdateStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReportErrorCar, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态更新成功", nil))
}

// ===== 评价 =====

// CreateReview C 端用户创建评价
// POST /api/v1/car/reviews  （需登录）
func (h *ReportHandler) CreateReview(ctx plugin.Context) {
	userID, username, _, avatar := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.CreateReviewRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	regionID := getRegionID(ctx)
	info, err := h.reviewSvc.Create(regionID, userID, username, avatar, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReviewErrorCar, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("评价成功", info))
}

// UpdateReview 更新评价（仅评价人）
// PUT /api/v1/car/reviews/:id  （需登录）
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

	var req dto.UpdateReviewRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.reviewSvc.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReviewErrorCar, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// DeleteReview 删除评价（仅评价人）
// DELETE /api/v1/car/reviews/:id  （需登录）
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

	if err := h.reviewSvc.Delete(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReviewErrorCar, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetReview 评价详情
// GET /api/v1/car/reviews/:id  （公开）
func (h *ReportHandler) GetReview(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的评价ID"))
		return
	}

	info, err := h.reviewSvc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReviewErrorCar, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListReviews 评价列表
// GET /api/v1/car/reviews  （公开）
func (h *ReportHandler) ListReviews(ctx plugin.Context) {
	var req dto.ReviewListRequest
	_ = ctx.Bind(&req)

	regionID := getRegionID(ctx)
	pagination, list, err := h.reviewSvc.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReviewErrorCar, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListReviewsByTarget 按目标对象查询评价
// GET /api/v1/car/reviews/by-target  （公开）
func (h *ReportHandler) ListReviewsByTarget(ctx plugin.Context) {
	targetType := ctx.Query("target_type")
	targetIDStr := ctx.Query("target_id")
	if targetType == "" || targetIDStr == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("target_type 和 target_id 不能为空"))
		return
	}

	targetID, err := strToUint(targetIDStr)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的 target_id"))
		return
	}

	regionID := getRegionID(ctx)
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.reviewSvc.ListByTarget(regionID, targetType, targetID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReviewErrorCar, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListMyReviews 我的评价
// GET /api/v1/car/reviews/mine  （需登录）
func (h *ReportHandler) ListMyReviews(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.reviewSvc.ListByReviewer(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReviewErrorCar, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ReplyReview 回复评价
// POST /api/v1/car/reviews/:id/reply  （需登录）
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

	if err := h.reviewSvc.Reply(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReviewErrorCar, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("回复成功", nil))
}

// AppendReview 追加评价
// POST /api/v1/car/reviews/:id/append  （需登录）
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

	if err := h.reviewSvc.Append(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReviewErrorCar, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("追加成功", nil))
}

// ReviewStats 评价统计
// GET /api/v1/car/reviews/stats  （公开）
func (h *ReportHandler) ReviewStats(ctx plugin.Context) {
	targetType := ctx.Query("target_type")
	targetIDStr := ctx.Query("target_id")
	if targetType == "" || targetIDStr == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("target_type 和 target_id 不能为空"))
		return
	}

	targetID, err := strToUint(targetIDStr)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的 target_id"))
		return
	}

	resp, err := h.reviewSvc.Stats(targetType, targetID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReviewErrorCar, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// LikeReview 点赞评价
// POST /api/v1/car/reviews/:id/like  （需登录）
func (h *ReportHandler) LikeReview(ctx plugin.Context) {
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

	if err := h.reviewSvc.Like(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReviewErrorCar, err.Error()))
		return
	}
	_ = userID
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("点赞成功", nil))
}

// ===== M 端评价管理 =====

// AdminListReviews 管理后台评价列表
// GET /api/v1/car/admin/reviews  （需 car:audit 权限）
func (h *ReportHandler) AdminListReviews(ctx plugin.Context) {
	var req dto.ReviewListRequest
	_ = ctx.Bind(&req)

	pagination, list, err := h.reviewSvc.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReviewErrorCar, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// UpdateReviewStatus 更新评价状态（M 端）
// PUT /api/v1/car/admin/reviews/:id/status  （需 car:audit 权限）
func (h *ReportHandler) UpdateReviewStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的评价ID"))
		return
	}

	var req dto.AdminUpdateStatusRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.reviewSvc.UpdateStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeReviewErrorCar, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态更新成功", nil))
}
