// Package handler 同城商城 HTTP 处理层 - 评价
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/mall/dto"
	"wuchang-tongcheng/internal/modules/mall/service"
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

// Create 创建评价
// POST /api/v1/mall/reviews  （需登录）
func (h *ReviewHandler) Create(ctx plugin.Context) {
	userID, username, _, _ := getUserProfile(ctx)
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
	info, err := h.svc.Create(regionID, userID, username, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("评价已提交", info))
}

// Reply 商家回复评价
// POST /api/v1/mall/reviews/:id/reply  （需登录，卖家操作）
func (h *ReviewHandler) Reply(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.ReplyReviewRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Reply(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("回复成功", nil))
}

// Append 追加评价
// PUT /api/v1/mall/reviews/:id/append  （需登录）
func (h *ReviewHandler) Append(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.AppendReviewRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Append(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("追加成功", nil))
}

// Delete 删除评价
// DELETE /api/v1/mall/admin/reviews/:id  （需 mall:audit 权限）
func (h *ReviewHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallReviewNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 评价详情
// GET /api/v1/mall/reviews/:id  （公开）
func (h *ReviewHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallReviewNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 评价列表（C 端）
// GET /api/v1/mall/reviews  （公开）
func (h *ReviewHandler) List(ctx plugin.Context) {
	var req dto.ReviewListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdminList 管理后台评价列表
// GET /api/v1/mall/admin/reviews  （需 mall:audit 权限）
func (h *ReviewHandler) AdminList(ctx plugin.Context) {
	var req dto.AdminReviewListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByProduct 按商品列出评价
// GET /api/v1/mall/reviews/by-product/:product_id  （公开）
func (h *ReviewHandler) ListByProduct(ctx plugin.Context) {
	productID, err := parseSubID(ctx, "product_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的商品ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListByProduct(productID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByShop 按店铺列出评价
// GET /api/v1/mall/reviews/by-shop/:shop_id  （公开）
func (h *ReviewHandler) ListByShop(ctx plugin.Context) {
	shopID, err := parseSubID(ctx, "shop_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的店铺ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListByShop(shopID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByUser 按用户列出评价
// GET /api/v1/mall/reviews/mine  （需登录）
func (h *ReviewHandler) ListByUser(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListByUser(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByOrder 按订单列出评价
// GET /api/v1/mall/reviews/by-order/:order_id  （需登录）
func (h *ReviewHandler) ListByOrder(ctx plugin.Context) {
	orderID, err := parseSubID(ctx, "order_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的订单ID"))
		return
	}
	list, err := h.svc.ListByOrder(orderID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// UpdateStatus 更新评价状态（管理后台审核）
// PUT /api/v1/mall/admin/reviews/:id/status  （需 mall:audit 权限）
func (h *ReviewHandler) UpdateStatus(ctx plugin.Context) {
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
	if err := h.svc.UpdateStatus(id, req.AuditStatus, req.AuditReason); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态已更新", nil))
}

// Stats 评价统计
// GET /api/v1/mall/reviews/stats  （公开）
func (h *ReviewHandler) Stats(ctx plugin.Context) {
	var req dto.ReviewStatsRequest
	_ = ctx.Bind(&req)
	stats, err := h.svc.Stats(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(stats))
}

// Like 点赞 +1
// POST /api/v1/mall/reviews/:id/like  （需登录）
func (h *ReviewHandler) Like(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.IncrLikeCount(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已点赞", nil))
}

// Dislike 踩 +1
// POST /api/v1/mall/reviews/:id/dislike  （需登录）
func (h *ReviewHandler) Dislike(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.IncrDislikeCount(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已踩", nil))
}
