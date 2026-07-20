// Package handler 同城拼车出行 HTTP 处理层 - 紧急联系人/一键报警
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

// EmergencyHandler 紧急联系人/一键报警 HTTP 处理器
type EmergencyHandler struct {
	service service.EmergencyService
}

// NewEmergencyHandler 创建 EmergencyHandler 实例
func NewEmergencyHandler(svc service.EmergencyService) *EmergencyHandler {
	return &EmergencyHandler{service: svc}
}

// ===== C 端 - 紧急联系人 =====

// CreateContact 创建紧急联系人
// POST /api/v1/pinche/emergency-contacts  （需登录）
func (h *EmergencyHandler) CreateContact(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.CreateEmergencyContactRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	regionID := getRegionID(ctx)
	info, err := h.service.CreateContact(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("紧急联系人已添加", info))
}

// UpdateContact 更新紧急联系人
// PUT /api/v1/pinche/emergency-contacts/:id  （需登录）
func (h *EmergencyHandler) UpdateContact(ctx plugin.Context) {
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

	var req dto.UpdateEmergencyContactRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.service.UpdateContact(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheEmergencyNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// DeleteContact 删除紧急联系人
// DELETE /api/v1/pinche/emergency-contacts/:id  （需登录）
func (h *EmergencyHandler) DeleteContact(ctx plugin.Context) {
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

	if err := h.service.DeleteContact(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheEmergencyNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// ListContacts 紧急联系人列表
// GET /api/v1/pinche/emergency-contacts  （需登录）
func (h *EmergencyHandler) ListContacts(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListContacts(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ===== C 端 - 一键报警 =====

// SOS 一键报警
// POST /api/v1/pinche/emergencies/sos  （需登录）
func (h *EmergencyHandler) SOS(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.SOSAlertRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	regionID := getRegionID(ctx)
	info, err := h.service.SOS(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("报警已发送", info))
}

// GetAlert 报警详情
// GET /api/v1/pinche/emergencies/:id  （需登录）
func (h *EmergencyHandler) GetAlert(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.service.GetAlert(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheEmergencyNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListMyAlerts 我的报警记录
// GET /api/v1/pinche/emergencies/mine  （需登录）
func (h *EmergencyHandler) ListMyAlerts(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListMyAlerts(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListAlertsByPinche 按拼车行程查询报警
// GET /api/v1/pinche/pinches/:id/emergencies  （需登录）
func (h *EmergencyHandler) ListAlertsByPinche(ctx plugin.Context) {
	pincheID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的拼车ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListAlertsByPinche(pincheID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListAlertsByTrip 按行程查询报警
// GET /api/v1/pinche/trips/:id/emergencies  （需登录）
func (h *EmergencyHandler) ListAlertsByTrip(ctx plugin.Context) {
	tripID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的行程ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListAlertsByTrip(tripID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ===== M 端管理 =====

// AdminListAlerts 管理后台报警列表
// GET /api/v1/pinche/admin/emergencies  （需 pinche:audit 权限）
func (h *EmergencyHandler) AdminListAlerts(ctx plugin.Context) {
	var req dto.EmergencyListRequest
	_ = ctx.Bind(&req)

	pagination, list, err := h.service.AdminListAlerts(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// HandleAlert 处理报警
// PUT /api/v1/pinche/admin/emergencies/:id/handle  （需 pinche:audit 权限）
func (h *EmergencyHandler) HandleAlert(ctx plugin.Context) {
	handlerID, _, _, _ := getUserProfile(ctx)
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	var req dto.HandleAlertRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.service.HandleAlert(id, handlerID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheEmergencyNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("报警处理完成", nil))
}

// UpdateAlertStatus 管理后台更新报警状态
// PUT /api/v1/pinche/admin/emergencies/:id/status  （需 pinche:audit 权限）
func (h *EmergencyHandler) UpdateAlertStatus(ctx plugin.Context) {
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
	if err := h.service.UpdateAlertStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheStatusInvalid, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态更新成功", nil))
}
