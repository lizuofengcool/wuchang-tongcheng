// Package handler 同城拼车出行 HTTP 处理层 - 完成行程（含行程分享）
// 依据 v3.2.1 架构方案：对标哈啰出行/嘀嗒出行/滴滴顺风车
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/pinche/dto"
	"wuchang-tongcheng/internal/modules/pinche/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// TripHandler 完成行程 HTTP 处理器
type TripHandler struct {
	service service.TripService
}

// NewTripHandler 创建 TripHandler 实例
func NewTripHandler(svc service.TripService) *TripHandler {
	return &TripHandler{service: svc}
}

// ===== C 端 =====

// Start 启动行程
// POST /api/v1/pinche/trips  （需登录）
func (h *TripHandler) Start(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.StartTripRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	regionID := getRegionID(ctx)
	info, err := h.service.Start(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("行程已启动", info))
}

// Complete 完成行程（仅车主可调用）
// PUT /api/v1/pinche/trips/:id/complete  （需登录）
func (h *TripHandler) Complete(ctx plugin.Context) {
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

	var req dto.CompleteTripRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	info, err := h.service.Complete(id, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheTripNotStarted, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("行程已完成", info))
}

// Confirm 确认行程（车主/乘客双方确认）
// POST /api/v1/pinche/trips/:id/confirm  （需登录）
func (h *TripHandler) Confirm(ctx plugin.Context) {
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

	var req dto.ConfirmTripRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	info, err := h.service.Confirm(id, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("行程已确认", info))
}

// GetByID 行程详情
// GET /api/v1/pinche/trips/:id  （公开）
func (h *TripHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheTripNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetByTripNo 按行程单号查询
// GET /api/v1/pinche/trips/no/:trip_no  （公开）
func (h *TripHandler) GetByTripNo(ctx plugin.Context) {
	tripNo := ctx.Param("trip_no")
	if tripNo == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的行程单号"))
		return
	}

	info, err := h.service.GetByTripNo(tripNo)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheTripNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListByUser 我的行程（车主+乘客）
// GET /api/v1/pinche/trips/mine  （需登录）
func (h *TripHandler) ListByUser(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListByUser(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByDriver 按车主查询行程
// GET /api/v1/pinche/trips/driver/:id  （需登录）
func (h *TripHandler) ListByDriver(ctx plugin.Context) {
	driverID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的车主ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListByDriver(driverID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByPassenger 按乘客查询行程
// GET /api/v1/pinche/trips/passenger/:id  （需登录）
func (h *TripHandler) ListByPassenger(ctx plugin.Context) {
	passengerID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的乘客ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListByPassenger(passengerID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByPinche 按拼车行程查询
// GET /api/v1/pinche/pinches/:id/trips  （公开）
func (h *TripHandler) ListByPinche(ctx plugin.Context) {
	pincheID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的拼车ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListByPinche(pincheID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Share 生成行程分享
// POST /api/v1/pinche/trips/:id/share  （需登录）
func (h *TripHandler) Share(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.ShareTripRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.service.Share(id, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheTripNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("分享链接已生成", resp))
}

// GetByShareToken 通过分享 token 查询行程
// GET /api/v1/pinche/trips/share/:token  （公开）
func (h *TripHandler) GetByShareToken(ctx plugin.Context) {
	token := ctx.Param("token")
	if token == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的分享 token"))
		return
	}
	info, err := h.service.GetByShareToken(token)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheTripNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ===== M 端管理 =====

// AdminList 管理后台行程列表
// GET /api/v1/pinche/admin/trips  （需 pinche:audit 权限）
func (h *TripHandler) AdminList(ctx plugin.Context) {
	var req dto.TripListRequest
	_ = ctx.Bind(&req)

	pagination, list, err := h.service.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// UpdateStatus 管理后台更新行程状态
// PUT /api/v1/pinche/admin/trips/:id/status  （需 pinche:audit 权限）
func (h *TripHandler) UpdateStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateStatusRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.UpdateStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheStatusInvalid, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态更新成功", nil))
}
