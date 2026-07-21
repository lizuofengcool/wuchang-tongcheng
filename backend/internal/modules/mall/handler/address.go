// Package handler 同城商城 HTTP 处理层 - 收货地址
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/mall/dto"
	"wuchang-tongcheng/internal/modules/mall/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// AddressHandler 收货地址 HTTP 处理器
type AddressHandler struct {
	svc service.AddressService
}

// NewAddressHandler 创建收货地址 Handler 实例
func NewAddressHandler(svc service.AddressService) *AddressHandler {
	return &AddressHandler{svc: svc}
}

// Create 创建收货地址
// POST /api/v1/mall/addresses  （需登录）
func (h *AddressHandler) Create(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	var req dto.CreateAddressRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.svc.Create(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallAddressError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// Update 更新收货地址
// PUT /api/v1/mall/addresses/:id  （需登录）
func (h *AddressHandler) Update(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateAddressRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallAddressError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除收货地址
// DELETE /api/v1/mall/addresses/:id  （需登录）
func (h *AddressHandler) Delete(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Delete(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallAddressError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 收货地址详情
// GET /api/v1/mall/addresses/:id  （需登录）
func (h *AddressHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallAddressNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetDefault 获取默认地址
// GET /api/v1/mall/addresses/default  （需登录）
func (h *AddressHandler) GetDefault(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	info, err := h.svc.GetDefault(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallAddressNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListByUser 用户的收货地址列表
// GET /api/v1/mall/addresses  （需登录）
func (h *AddressHandler) ListByUser(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	list, err := h.svc.ListByUser(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallAddressError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// SetDefault 设默认地址
// PUT /api/v1/mall/addresses/:id/default  （需登录）
func (h *AddressHandler) SetDefault(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.SetDefault(userID, id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallAddressError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("设置成功", nil))
}

// List 管理后台地址列表
// GET /api/v1/mall/admin/addresses  （需 mall:audit 权限）
func (h *AddressHandler) List(ctx plugin.Context) {
	var req dto.AddressListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallAddressError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}
