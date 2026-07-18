// Package handler 团购优惠券HTTP处理层
package handler

import (
	"net/http"
	"strconv"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/groupbuy/dto"
	"wuchang-tongcheng/internal/modules/groupbuy/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// Handler 团购HTTP处理器
type Handler struct {
	service service.Service
}

// NewHandler 创建团购处理器
func NewHandler(svc service.Service) *Handler {
	return &Handler{service: svc}
}

// getUserID 从上下文获取用户ID和用户名
func getUserID(ctx plugin.Context) (uint, string) {
	userID, _ := ctx.Get(middleware.ContextUserID)
	username, _ := ctx.Get(middleware.ContextUsername)
	id, _ := userID.(uint)
	name, _ := username.(string)
	return id, name
}

// getRegionID 从上下文获取地区ID
func getRegionID(ctx plugin.Context) uint {
	if v, ok := ctx.Get(middleware.RegionIDKey); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return middleware.DefaultRegionID
}

// parseID 解析URL中的ID参数
func parseID(ctx plugin.Context) (uint, error) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// ===== 公开接口 =====

// List 团购商品列表
func (h *Handler) List(ctx plugin.Context) {
	var req dto.GroupBuyListRequest
	_ = ctx.Bind(&req)
	// 公开列表默认只查上架
	if req.Status < 0 || req.Status > 2 {
		req.Status = 1
	}

	regionID := getRegionID(ctx)
	pagination, list, err := h.service.ListGroupBuy(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeGroupBuyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// GetByID 团购详情
func (h *Handler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的团购ID"))
		return
	}

	info, err := h.service.GetGroupBuy(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeGroupBuyNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListCoupons 可领取的优惠券列表
func (h *Handler) ListCoupons(ctx plugin.Context) {
	var req dto.CouponListRequest
	_ = ctx.Bind(&req)

	regionID := getRegionID(ctx)
	pagination, list, err := h.service.AvailableCoupons(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCouponError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ===== 用户接口 =====

// CreateOrder 创建团购订单
func (h *Handler) CreateOrder(ctx plugin.Context) {
	userID, _ := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.CreateOrderRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}

	regionID := getRegionID(ctx)
	info, err := h.service.CreateOrder(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeOrderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("下单成功", info))
}

// MyOrders 我的订单列表
func (h *Handler) MyOrders(ctx plugin.Context) {
	userID, _ := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.OrderListRequest
	_ = ctx.Bind(&req)

	pagination, list, err := h.service.MyOrders(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeOrderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// GetOrder 订单详情
func (h *Handler) GetOrder(ctx plugin.Context) {
	userID, _ := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的订单ID"))
		return
	}

	info, err := h.service.GetOrder(id, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeOrderNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// CancelOrder 取消订单
func (h *Handler) CancelOrder(ctx plugin.Context) {
	userID, _ := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的订单ID"))
		return
	}

	if err := h.service.CancelOrder(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeOrderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("取消成功", nil))
}

// ReceiveCoupon 领取优惠券
func (h *Handler) ReceiveCoupon(ctx plugin.Context) {
	userID, _ := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的优惠券ID"))
		return
	}

	regionID := getRegionID(ctx)
	info, err := h.service.ReceiveCoupon(regionID, userID, id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCouponError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("领取成功", info))
}

// MyCoupons 我的优惠券
func (h *Handler) MyCoupons(ctx plugin.Context) {
	userID, _ := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.CouponListRequest
	_ = ctx.Bind(&req)

	pagination, list, err := h.service.MyCoupons(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCouponError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// VerifyOrder 核销订单（商家核销）
func (h *Handler) VerifyOrder(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的订单ID"))
		return
	}

	var req dto.VerifyOrderRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}

	if err := h.service.VerifyOrder(id, req.VerifyCode); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeOrderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("核销成功", nil))
}

// ===== 管理接口 =====

// AdminCreate 创建团购商品
func (h *Handler) AdminCreate(ctx plugin.Context) {
	userID, _ := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.CreateGroupBuyRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}

	regionID := getRegionID(ctx)
	info, err := h.service.CreateGroupBuy(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeGroupBuyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// AdminUpdate 编辑团购
func (h *Handler) AdminUpdate(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的团购ID"))
		return
	}

	var req dto.UpdateGroupBuyRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}

	if err := h.service.UpdateGroupBuy(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeGroupBuyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// AdminDelete 删除团购
func (h *Handler) AdminDelete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的团购ID"))
		return
	}

	if err := h.service.DeleteGroupBuy(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeGroupBuyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// AdminUpdateStatus 上架/下架
func (h *Handler) AdminUpdateStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的团购ID"))
		return
	}

	var req dto.UpdateGroupBuyStatusRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}

	if err := h.service.UpdateGroupBuyStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeGroupBuyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("操作成功", nil))
}

// AdminAudit 审核团购
func (h *Handler) AdminAudit(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的团购ID"))
		return
	}

	var req dto.AuditGroupBuyRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}

	if err := h.service.AuditGroupBuy(id, req.AuditStatus); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeGroupBuyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核成功", nil))
}

// AdminList 团购管理列表（所有状态）
func (h *Handler) AdminList(ctx plugin.Context) {
	var req dto.GroupBuyListRequest
	_ = ctx.Bind(&req)
	// 管理端默认查全部状态
	if req.Status < 0 || req.Status > 2 {
		req.Status = -1
	}

	regionID := getRegionID(ctx)
	pagination, list, err := h.service.ListGroupBuy(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeGroupBuyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdminOrderList 订单管理列表
func (h *Handler) AdminOrderList(ctx plugin.Context) {
	var req dto.OrderListRequest
	_ = ctx.Bind(&req)

	regionID := getRegionID(ctx)
	pagination, list, err := h.service.AdminOrderList(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeOrderError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdminCreateCoupon 创建优惠券
func (h *Handler) AdminCreateCoupon(ctx plugin.Context) {
	var req dto.CreateCouponRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}

	regionID := getRegionID(ctx)
	info, err := h.service.CreateCoupon(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCouponError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// AdminUpdateCoupon 编辑优惠券
func (h *Handler) AdminUpdateCoupon(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的优惠券ID"))
		return
	}

	var req dto.UpdateCouponRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}

	if err := h.service.UpdateCoupon(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCouponError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// AdminDeleteCoupon 删除优惠券
func (h *Handler) AdminDeleteCoupon(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的优惠券ID"))
		return
	}

	if err := h.service.DeleteCoupon(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCouponError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// AdminListCoupons 优惠券列表
func (h *Handler) AdminListCoupons(ctx plugin.Context) {
	var req dto.CouponListRequest
	_ = ctx.Bind(&req)
	if req.Status < 0 || req.Status > 1 {
		req.Status = -1
	}

	regionID := getRegionID(ctx)
	pagination, list, err := h.service.ListCoupon(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCouponError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}
