// Package handler love 相亲交友 HTTP 处理层 - 会员（会员等级 + 会员订阅）
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/love/dto"
	"wuchang-tongcheng/internal/modules/love/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// MemberLevelHandler 会员等级 HTTP 处理器（M 端配置）
type MemberLevelHandler struct {
	service service.LoveMemberLevelService
}

// NewMemberLevelHandler 创建 MemberLevelHandler 实例
func NewMemberLevelHandler(svc service.LoveMemberLevelService) *MemberLevelHandler {
	return &MemberLevelHandler{service: svc}
}

// Create 创建会员等级
// POST /api/v1/love/admin/member-levels  （需 love:audit 权限）
func (h *MemberLevelHandler) Create(ctx plugin.Context) {
	var req dto.CreateLoveMemberLevelRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.service.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// Update 更新会员等级
// PUT /api/v1/love/admin/member-levels/:id  （需 love:audit 权限）
func (h *MemberLevelHandler) Update(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateLoveMemberLevelRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.Update(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除会员等级
// DELETE /api/v1/love/admin/member-levels/:id  （需 love:audit 权限）
func (h *MemberLevelHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.service.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 会员等级详情
// GET /api/v1/love/member-levels/:id  （公开）
func (h *MemberLevelHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetByLevelCode 按等级编码查询
// GET /api/v1/love/member-levels/code/:code  （公开）
func (h *MemberLevelHandler) GetByLevelCode(ctx plugin.Context) {
	code := ctx.Param("code")
	if code == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("code 不能为空"))
		return
	}
	info, err := h.service.GetByLevelCode(code)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 会员等级列表
// GET /api/v1/love/member-levels  （公开）
func (h *MemberLevelHandler) List(ctx plugin.Context) {
	var req dto.LoveMembershipListRequest
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListAll 全部启用会员等级
// GET /api/v1/love/member-levels/all  （公开）
func (h *MemberLevelHandler) ListAll(ctx plugin.Context) {
	list, err := h.service.ListAll()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ===== 会员订阅 =====

// MembershipHandler 会员订阅 HTTP 处理器
type MembershipHandler struct {
	service service.LoveMembershipService
}

// NewMembershipHandler 创建 MembershipHandler 实例
func NewMembershipHandler(svc service.LoveMembershipService) *MembershipHandler {
	return &MembershipHandler{service: svc}
}

// Open 开通会员
// POST /api/v1/love/memberships  （需登录）
func (h *MembershipHandler) Open(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	var req dto.CreateLoveMembershipRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	loveIDStr := ctx.Query("love_id")
	loveID, _ := parseUint(loveIDStr)
	info, err := h.service.Open(loveID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("开通成功", info))
}

// GetByID 会员订阅详情
// GET /api/v1/love/memberships/:id  （需登录）
func (h *MembershipHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetMyActive 我的当前有效订阅
// GET /api/v1/love/memberships/me  （需登录）
func (h *MembershipHandler) GetMyActive(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	info, err := h.service.GetMyActive(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 会员订阅列表（M 端）
// GET /api/v1/love/admin/memberships  （需 love:audit 权限）
func (h *MembershipHandler) List(ctx plugin.Context) {
	var req dto.LoveMembershipListRequest
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByUser 我的订阅列表
// GET /api/v1/love/memberships  （需登录）
func (h *MembershipHandler) ListByUser(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	var req dto.LoveMembershipListRequest
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.ListByUser(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Cancel 取消订阅
// POST /api/v1/love/memberships/:id/cancel  （需登录）
func (h *MembershipHandler) Cancel(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.CancelLoveMembershipRequest
	_ = ctx.Bind(&req)
	if err := h.service.Cancel(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("取消成功", nil))
}

// Refund 退款
// POST /api/v1/love/memberships/:id/refund  （需登录）
func (h *MembershipHandler) Refund(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.RefundLoveMembershipRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.Refund(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("退款成功", nil))
}

// MarkPaid 标记已支付（支付回调调用）
// POST /api/v1/love/memberships/:id/paid  （需登录）
func (h *MembershipHandler) MarkPaid(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req struct {
		PayMethod  string `json:"pay_method" binding:"required,oneof=wechat alipay credits"`
		PayOrderNo string `json:"pay_order_no" binding:"required,max=64"`
	}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.MarkPaid(id, req.PayMethod, req.PayOrderNo); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("支付成功", nil))
}

// UpdateAutoRenew 更新自动续费
// PUT /api/v1/love/memberships/:id/auto-renew  （需登录）
func (h *MembershipHandler) UpdateAutoRenew(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req struct {
		AutoRenew bool `json:"auto_renew"`
	}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.UpdateAutoRenew(id, userID, req.AutoRenew); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}
