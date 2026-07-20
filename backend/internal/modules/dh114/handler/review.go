// Package handler 同城114 HTTP 处理层 - 评价 + 商家回复
// 依据 v3.2.1 架构方案：对标大众点评 5 星评价体系
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/dh114/dto"
	"wuchang-tongcheng/internal/modules/dh114/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ReviewHandler 评价 HTTP 处理器
type ReviewHandler struct {
	svc service.ReviewService
}

// NewReviewHandler 创建评价 Handler 实例
func NewReviewHandler(svc service.ReviewService) *ReviewHandler {
	return &ReviewHandler{svc: svc}
}

// Create 发布评价（C 端）
// POST /api/v1/dh114/reviews  （需登录）
func (h *ReviewHandler) Create(ctx plugin.Context) {
	userID, username, phone, avatar := getUserProfile(ctx)
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
	info, err := h.svc.Create(regionID, userID, username, phone, avatar, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114ReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("评价发布成功", info))
}

// Update 更新评价
// PUT /api/v1/dh114/reviews/:id  （需登录）
func (h *ReviewHandler) Update(ctx plugin.Context) {
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
	var req dto.UpdateReviewRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114ReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除评价
// DELETE /api/v1/dh114/reviews/:id  （需登录）
func (h *ReviewHandler) Delete(ctx plugin.Context) {
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
	if err := h.svc.Delete(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114ReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 评价详情
// GET /api/v1/dh114/reviews/:id  （公开）
func (h *ReviewHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114ReviewNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 评价列表
// GET /api/v1/dh114/reviews  （公开）
func (h *ReviewHandler) List(ctx plugin.Context) {
	var req dto.ReviewListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	regionID := getRegionID(ctx)
	pagination, list, err := h.svc.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114ReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByDh114 按商户列出评价
// GET /api/v1/dh114/dh114/:id/reviews  （公开）
func (h *ReviewHandler) ListByDh114(ctx plugin.Context) {
	dh114ID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的商户ID"))
		return
	}
	regionID := getRegionID(ctx)
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListByDh114(regionID, dh114ID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114ReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListMine 我的评价
// GET /api/v1/dh114/reviews/mine  （需登录）
func (h *ReviewHandler) ListMine(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListByReviewer(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114ReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Stats 评价统计
// GET /api/v1/dh114/reviews/stats  （公开）
func (h *ReviewHandler) Stats(ctx plugin.Context) {
	dh114IDStr := ctx.Query("id")
	dh114ID, err := parseUintStr(dh114IDStr)
	if err != nil || dh114ID == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的商户ID"))
		return
	}
	stats, err := h.svc.StatsByDh114(dh114ID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114ReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(stats))
}

// Reply 商家回复评价
// POST /api/v1/dh114/reviews/:id/reply  （需登录）
func (h *ReviewHandler) Reply(ctx plugin.Context) {
	userID, username, _, avatar := getUserProfile(ctx)
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
	if err := h.svc.Reply(id, userID, username, avatar, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114ReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("回复成功", nil))
}

// ListReplies 评价回复列表
// GET /api/v1/dh114/reviews/:id/replies  （公开）
func (h *ReviewHandler) ListReplies(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	list, err := h.svc.ListReplies(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114ReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// DeleteReply 删除回复
// DELETE /api/v1/dh114/reviews/replies/:reply_id  （需登录）
func (h *ReviewHandler) DeleteReply(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	replyID, err := parseSubID(ctx, "reply_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的回复ID"))
		return
	}
	if err := h.svc.DeleteReply(replyID, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114ReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// Like 点赞评价
// POST /api/v1/dh114/reviews/:id/like  （公开）
func (h *ReviewHandler) Like(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.IncrLikeCount(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114ReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("点赞成功", nil))
}

// Audit 审核评价（M 端）
// PUT /api/v1/dh114/admin/reviews/:id/audit  （需 dh114:audit 权限）
func (h *ReviewHandler) Audit(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.AuditRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.AuditReview(id, req.AuditStatus, req.AuditReason); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114ReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核完成", nil))
}

// BatchAudit 批量审核评价
// POST /api/v1/dh114/admin/reviews/batch-audit  （需 dh114:audit 权限）
func (h *ReviewHandler) BatchAudit(ctx plugin.Context) {
	var req dto.BatchAuditRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	result, err := h.svc.BatchAuditReviews(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114ReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(result))
}
