// Package handler 推广/物流/担保/退款 HTTP 处理层
// 依据 v3.2.1 架构方案：对标闲鱼/转转/瓜子
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/ershou/dto"
	"wuchang-tongcheng/internal/modules/ershou/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// TradeHandler 推广/物流/担保/退款 HTTP 处理器（聚合 4 个 service）
type TradeHandler struct {
	promotionSvc service.PromotionService
	logisticsSvc service.LogisticsService
	escrowSvc    service.EscrowService
	refundSvc    service.RefundService
}

// NewTradeHandler 创建 Trade Handler 实例
func NewTradeHandler(
	promotionSvc service.PromotionService,
	logisticsSvc service.LogisticsService,
	escrowSvc service.EscrowService,
	refundSvc service.RefundService,
) *TradeHandler {
	return &TradeHandler{
		promotionSvc: promotionSvc,
		logisticsSvc: logisticsSvc,
		escrowSvc:    escrowSvc,
		refundSvc:    refundSvc,
	}
}

// ===== 推广 =====

// CreatePromotion 创建推广
// POST /api/v1/ershou/:id/promotions  （需登录 + 仅发布者本人）
func (h *TradeHandler) CreatePromotion(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	ershouID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.PromotionCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.promotionSvc.Create(ershouID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("推广创建成功", resp))
}

// ListPromotions 商品推广记录列表
// GET /api/v1/ershou/:id/promotions  （公开）
func (h *TradeHandler) ListPromotions(ctx plugin.Context) {
	ershouID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	list, err := h.promotionSvc.ListByErshouID(ershouID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// PromotionStats 推广效果统计
// GET /api/v1/ershou/:id/promotions/stats  （公开）
func (h *TradeHandler) PromotionStats(ctx plugin.Context) {
	ershouID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	resp, err := h.promotionSvc.Stats(ershouID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// ===== 物流 =====

// CreateLogistics 创建物流记录（卖家发货时调用）
// POST /api/v1/ershou/orders/:id/logistics  （需登录 + 卖家）
func (h *TradeHandler) CreateLogistics(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	orderID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的订单ID"))
		return
	}
	var req dto.LogisticsCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.logisticsSvc.Create(orderID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("物流记录创建成功", resp))
}

// GetLogistics 查询订单物流
// GET /api/v1/ershou/orders/:id/logistics  （需登录 + 买卖双方）
func (h *TradeHandler) GetLogistics(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	orderID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的订单ID"))
		return
	}
	resp, err := h.logisticsSvc.GetByOrderID(orderID, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// UpdateLogistics 更新物流状态/轨迹
// PUT /api/v1/ershou/orders/:id/logistics  （需登录 + 卖家）
func (h *TradeHandler) UpdateLogistics(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	orderID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的订单ID"))
		return
	}
	var req dto.LogisticsUpdateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.logisticsSvc.Update(orderID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("物流更新成功", resp))
}

// ===== 担保 =====

// CreateEscrow 创建担保记录
// POST /api/v1/ershou/orders/:id/escrow  （需登录 + 买卖双方）
func (h *TradeHandler) CreateEscrow(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	orderID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的订单ID"))
		return
	}
	var req dto.EscrowCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.escrowSvc.Create(orderID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("担保记录创建成功", resp))
}

// GetEscrow 查询订单担保
// GET /api/v1/ershou/orders/:id/escrow  （需登录 + 买卖双方）
func (h *TradeHandler) GetEscrow(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	orderID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的订单ID"))
		return
	}
	resp, err := h.escrowSvc.GetByOrderID(orderID, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// ReleaseEscrow 放款（T+1 自动放款在调度任务中实现）
// POST /api/v1/ershou/orders/:id/escrow/release  （需登录）
func (h *TradeHandler) ReleaseEscrow(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	orderID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的订单ID"))
		return
	}
	var req dto.EscrowReleaseRequest
	_ = ctx.Bind(&req)
	resp, err := h.escrowSvc.Release(orderID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("放款成功", resp))
}

// ===== 退款 =====

// CreateRefund 买家申请退款
// POST /api/v1/ershou/orders/:id/refund  （需登录 + 买家）
func (h *TradeHandler) CreateRefund(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	orderID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的订单ID"))
		return
	}
	var req dto.RefundCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.refundSvc.Create(orderID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("退款申请已提交", resp))
}

// GetRefund 查询订单退款详情
// GET /api/v1/ershou/orders/:id/refund  （需登录 + 买卖双方）
func (h *TradeHandler) GetRefund(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	orderID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的订单ID"))
		return
	}
	resp, err := h.refundSvc.GetByOrderID(orderID, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// ProcessRefund 处理退款（卖家 approve/reject，平台 arbitrate）
// PUT /api/v1/ershou/refunds/:id/process  （需登录）
func (h *TradeHandler) ProcessRefund(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	refundID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的退款ID"))
		return
	}
	var req dto.RefundProcessRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.refundSvc.Process(refundID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("退款处理成功", resp))
}

// ListRefunds 退款列表（按 buyer/seller/all 角色查询）
// GET /api/v1/ershou/refunds  （需登录）
func (h *TradeHandler) ListRefunds(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.OrderQuery
	_ = ctx.Bind(&req)
	pagination, list, err := h.refundSvc.List(userID, req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}
