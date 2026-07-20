// Package handler 同城拼车出行 HTTP 处理层 - 顺风车保险
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

// InsuranceHandler 顺风车保险 HTTP 处理器
type InsuranceHandler struct {
	service service.InsuranceService
}

// NewInsuranceHandler 创建 InsuranceHandler 实例
func NewInsuranceHandler(svc service.InsuranceService) *InsuranceHandler {
	return &InsuranceHandler{service: svc}
}

// ===== C 端 =====

// Create 投保
// POST /api/v1/pinche/insurances  （需登录）
func (h *InsuranceHandler) Create(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.CreateInsuranceRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	regionID := getRegionID(ctx)
	info, err := h.service.Create(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheInsuranceFailed, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("投保成功", info))
}

// Claim 保险理赔
// POST /api/v1/pinche/insurances/:id/claim  （需登录）
func (h *InsuranceHandler) Claim(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.InsuranceClaimRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.service.Claim(id, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheInsuranceFailed, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("理赔申请已提交", info))
}

// GetByID 保险详情
// GET /api/v1/pinche/insurances/:id  （公开）
func (h *InsuranceHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheInsuranceFailed, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetByPolicyNo 按保单号查询
// GET /api/v1/pinche/insurances/no/:policy_no  （公开）
func (h *InsuranceHandler) GetByPolicyNo(ctx plugin.Context) {
	policyNo := ctx.Param("policy_no")
	if policyNo == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的保单号"))
		return
	}
	info, err := h.service.GetByPolicyNo(policyNo)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheInsuranceFailed, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListByPinche 按拼车行程查询保险
// GET /api/v1/pinche/pinches/:id/insurances  （公开）
func (h *InsuranceHandler) ListByPinche(ctx plugin.Context) {
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

// ListByBooking 按预订查询保险
// GET /api/v1/pinche/bookings/:id/insurances  （公开）
func (h *InsuranceHandler) ListByBooking(ctx plugin.Context) {
	bookingID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的预订ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListByBooking(bookingID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Quote 保险报价
// POST /api/v1/pinche/insurances/quote  （需登录）
func (h *InsuranceHandler) Quote(ctx plugin.Context) {
	var req dto.InsuranceQuoteRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.service.Quote(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheInsuranceFailed, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ===== M 端管理 =====

// AdminList 管理后台保险列表
// GET /api/v1/pinche/admin/insurances  （需 pinche:audit 权限）
func (h *InsuranceHandler) AdminList(ctx plugin.Context) {
	var req dto.InsuranceListRequest
	_ = ctx.Bind(&req)

	pagination, list, err := h.service.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// UpdateStatus 管理后台更新保险状态
// PUT /api/v1/pinche/admin/insurances/:id/status  （需 pinche:audit 权限）
func (h *InsuranceHandler) UpdateStatus(ctx plugin.Context) {
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
