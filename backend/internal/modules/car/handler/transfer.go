// Package handler 同城车辆买卖 HTTP 处理层 - 过户办理
// 依据 v3.2.1 架构方案：对标瓜子过户服务/人人车代办/优信过户
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 状态机：0待提交 1已提交 2审核中 3已审核 4已完成 5已取消 6退回 7归档
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/car/dto"
	"wuchang-tongcheng/internal/modules/car/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// TransferHandler 过户办理 HTTP 处理器
type TransferHandler struct {
	service service.TransferService
}

// NewTransferHandler 创建 TransferHandler 实例
func NewTransferHandler(svc service.TransferService) *TransferHandler {
	return &TransferHandler{service: svc}
}

// ===== C 端 =====

// Create 创建过户办理
// POST /api/v1/car/transfers  （需登录）
func (h *TransferHandler) Create(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.CreateTransferRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	regionID := getRegionID(ctx)
	info, err := h.service.Create(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeTransferError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("过户办理创建成功", info))
}

// Update 更新过户办理（仅创建者）
// PUT /api/v1/car/transfers/:id  （需登录）
func (h *TransferHandler) Update(ctx plugin.Context) {
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

	var req dto.UpdateTransferRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.service.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeTransferError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除过户办理（仅创建者）
// DELETE /api/v1/car/transfers/:id  （需登录）
func (h *TransferHandler) Delete(ctx plugin.Context) {
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

	if err := h.service.Delete(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeTransferError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 获取过户办理详情
// GET /api/v1/car/transfers/:id  （需登录）
func (h *TransferHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeTransferNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetByCarID 按车源获取过户办理
// GET /api/v1/car/cars/:id/transfer  （需登录）
func (h *TransferHandler) GetByCarID(ctx plugin.Context) {
	carID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的车源ID"))
		return
	}

	info, err := h.service.GetByCarID(carID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeTransferNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 过户办理列表
// GET /api/v1/car/transfers  （需登录）
func (h *TransferHandler) List(ctx plugin.Context) {
	var req dto.TransferListRequest
	_ = ctx.Bind(&req)

	regionID := getRegionID(ctx)
	pagination, list, err := h.service.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeTransferError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListBySeller 我卖出的过户
// GET /api/v1/car/transfers/sold  （需登录）
func (h *TransferHandler) ListBySeller(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListBySeller(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeTransferError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByBuyer 我买到的过户
// GET /api/v1/car/transfers/bought  （需登录）
func (h *TransferHandler) ListByBuyer(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListByBuyer(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeTransferError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ===== M 端管理 =====

// AdminList 管理后台过户办理列表
// GET /api/v1/car/admin/transfers  （需 car:audit 权限）
func (h *TransferHandler) AdminList(ctx plugin.Context) {
	var req dto.TransferListRequest
	_ = ctx.Bind(&req)

	pagination, list, err := h.service.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeTransferError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdminGetByID 管理后台过户办理详情
// GET /api/v1/car/admin/transfers/:id  （需 car:audit 权限）
func (h *TransferHandler) AdminGetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	info, err := h.service.AdminGetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeTransferNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// UpdateStatus 更新过户办理状态
// PUT /api/v1/car/admin/transfers/:id/status  （需 car:audit 权限）
func (h *TransferHandler) UpdateStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	var req dto.TransferStatusUpdateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.service.UpdateStatus(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeTransferError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态更新成功", nil))
}
