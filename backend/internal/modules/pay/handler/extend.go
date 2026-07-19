// Package handler 支付中台扩展 HTTP 处理层
// 依据 012_pay_full.sql：交易流水/渠道/商户/回调/担保争议/统计
package handler

import (
	"net/http"
	"strconv"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/pay/dto"
	"wuchang-tongcheng/internal/modules/pay/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ExtendHandler 支付中台扩展处理器
type ExtendHandler struct {
	extSvc service.PayExtendService
}

// NewExtendHandler 创建扩展 Handler 实例
func NewExtendHandler(extSvc service.PayExtendService) *ExtendHandler {
	return &ExtendHandler{extSvc: extSvc}
}

// ===== 交易流水 =====

// ListTransactions 交易流水列表
// GET /api/v1/pay/transactions （需登录，M 端可查所有）
func (h *ExtendHandler) ListTransactions(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	var req dto.TransactionListRequest
	_ = ctx.Bind(&req)
	req.Page = page
	req.PageSize = pageSize
	// 普通用户只能看自己的流水；M 端可传 user_id=0 查所有
	if !hasFinancePermission(ctx) {
		req.UserID = userID
	}
	list, total, err := h.extSvc.ListTransactions(getRegionID(ctx), &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePayError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// GetTransaction 查询交易流水
// GET /api/v1/pay/transactions/:txn_no （需登录）
func (h *ExtendHandler) GetTransaction(ctx plugin.Context) {
	txnNo := ctx.Param("txn_no")
	if txnNo == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("流水号不能为空"))
		return
	}
	info, err := h.extSvc.GetTransaction(txnNo)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePayError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ===== 渠道 =====

// CreateChannel 创建支付渠道（M 端）
// POST /api/v1/pay/admin/channels
func (h *ExtendHandler) CreateChannel(ctx plugin.Context) {
	var req dto.CreateChannelRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.extSvc.CreateChannel(getRegionID(ctx), &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePayError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建渠道成功", info))
}

// UpdateChannel 更新支付渠道（M 端）
// POST /api/v1/pay/admin/channels/:id
func (h *ExtendHandler) UpdateChannel(ctx plugin.Context) {
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if id == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("渠道ID无效"))
		return
	}
	var req dto.UpdateChannelRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.extSvc.UpdateChannel(uint(id), &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePayChannelNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// DeleteChannel 删除支付渠道（M 端）
// DELETE /api/v1/pay/admin/channels/:id
func (h *ExtendHandler) DeleteChannel(ctx plugin.Context) {
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if id == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("渠道ID无效"))
		return
	}
	if err := h.extSvc.DeleteChannel(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePayChannelNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// ListChannels 渠道列表
// GET /api/v1/pay/channels （需登录）
func (h *ExtendHandler) ListChannels(ctx plugin.Context) {
	code := ctx.Query("channel_code")
	page, pageSize := parsePagination(ctx)
	list, total, err := h.extSvc.ListChannels(code, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePayError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// GetChannel 查询渠道详情
// GET /api/v1/pay/channels/:id （需登录）
func (h *ExtendHandler) GetChannel(ctx plugin.Context) {
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if id == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("渠道ID无效"))
		return
	}
	info, err := h.extSvc.GetChannel(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePayChannelNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ===== 商户 =====

// CreateMerchant 创建商户
// POST /api/v1/pay/merchants （需登录）
func (h *ExtendHandler) CreateMerchant(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.CreateMerchantRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if req.UserID == 0 {
		req.UserID = userID
	}
	info, err := h.extSvc.CreateMerchant(getRegionID(ctx), &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePayError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("商户申请已提交", info))
}

// AuditMerchant 审核商户（M 端）
// POST /api/v1/pay/admin/merchants/audit
func (h *ExtendHandler) AuditMerchant(ctx plugin.Context) {
	var req dto.AuditMerchantRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.extSvc.AuditMerchant(&req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePayError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核完成", nil))
}

// ListMerchants 商户列表
// GET /api/v1/pay/merchants
func (h *ExtendHandler) ListMerchants(ctx plugin.Context) {
	status, _ := strconv.Atoi(ctx.DefaultQuery("status", "-1"))
	page, pageSize := parsePagination(ctx)
	list, total, err := h.extSvc.ListMerchants(status, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePayError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// GetMerchant 查询商户详情
// GET /api/v1/pay/merchants/:id
func (h *ExtendHandler) GetMerchant(ctx plugin.Context) {
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if id == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("商户ID无效"))
		return
	}
	info, err := h.extSvc.GetMerchant(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePayError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ===== 回调 =====

// RecordCallback 记录三方支付回调
// POST /api/v1/pay/callbacks （不强制登录，第三方调用）
func (h *ExtendHandler) RecordCallback(ctx plugin.Context) {
	var req dto.RecordCallbackRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.extSvc.RecordCallback(getRegionID(ctx), &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePayError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListCallbacks 回调记录列表（M 端）
// GET /api/v1/pay/admin/callbacks
func (h *ExtendHandler) ListCallbacks(ctx plugin.Context) {
	orderNo := ctx.Query("order_no")
	channel := ctx.Query("channel")
	status, _ := strconv.Atoi(ctx.DefaultQuery("status", "-1"))
	page, pageSize := parsePagination(ctx)
	list, total, err := h.extSvc.ListCallbacks(orderNo, channel, status, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePayError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// GetCallback 查询回调详情
// GET /api/v1/pay/callbacks/:id
func (h *ExtendHandler) GetCallback(ctx plugin.Context) {
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if id == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("回调ID无效"))
		return
	}
	info, err := h.extSvc.GetCallback(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePayError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ===== 担保争议 =====

// CreateDispute 买家发起担保争议
// POST /api/v1/pay/escrow/disputes （需登录）
func (h *ExtendHandler) CreateDispute(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.EscrowDisputeRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.extSvc.CreateDispute(getRegionID(ctx), userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePayEscrowNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("争议已发起，等待平台仲裁", nil))
}

// ListDisputes 争议列表（M 端）
// GET /api/v1/pay/admin/disputes
func (h *ExtendHandler) ListDisputes(ctx plugin.Context) {
	page, pageSize := parsePagination(ctx)
	list, total, err := h.extSvc.ListDisputes(page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePayError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// ArbitrateDispute 仲裁争议（M 端）
// POST /api/v1/pay/admin/disputes/arbitrate
func (h *ExtendHandler) ArbitrateDispute(ctx plugin.Context) {
	arbitratorID := getUserID(ctx)
	var req dto.EscrowArbitrateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.extSvc.ArbitrateDispute(arbitratorID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePayEscrowNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("仲裁完成", nil))
}

// ===== 统计 =====

// Statistics 支付总览统计（M 端）
// GET /api/v1/pay/admin/statistics
func (h *ExtendHandler) Statistics(ctx plugin.Context) {
	resp, err := h.extSvc.Statistics()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePayError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// ===== 辅助函数 =====

// hasFinancePermission 检查当前用户是否具备 finance 权限
func hasFinancePermission(ctx plugin.Context) bool {
	if v, ok := ctx.Get("user_permissions"); ok {
		if perms, ok := v.([]string); ok {
			for _, p := range perms {
				if p == "finance:reconcile" || p == "finance:*" || p == "*" {
					return true
				}
			}
		}
	}
	return false
}
