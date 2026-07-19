// Package handler 同城车辆买卖 HTTP 处理层 - 车险配置
// 依据 v3.2.1 架构方案：对标平安车险/太平洋车险/人保车险
// Insurance 为全局配置数据（无 region_id），管理后台维护
package handler

import (
	"net/http"
	"strconv"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/car/dto"
	"wuchang-tongcheng/internal/modules/car/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// InsuranceHandler 车险配置 HTTP 处理器
type InsuranceHandler struct {
	service service.InsuranceService
}

// NewInsuranceHandler 创建 InsuranceHandler 实例
func NewInsuranceHandler(svc service.InsuranceService) *InsuranceHandler {
	return &InsuranceHandler{service: svc}
}

// ===== C 端 =====

// ListPublished 已发布的车险列表（C 端浏览）
// GET /api/v1/car/insurances  （公开）
func (h *InsuranceHandler) ListPublished(ctx plugin.Context) {
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListPublished(page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeInsuranceError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListHot 热门车险
// GET /api/v1/car/insurances/hot  （公开）
func (h *InsuranceHandler) ListHot(ctx plugin.Context) {
	limitStr := ctx.Query("limit")
	limit := 10
	if n, err := strconv.Atoi(limitStr); err == nil && n > 0 && n <= 100 {
		limit = n
	}

	list, err := h.service.ListHot(limit)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeInsuranceError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// GetByID 车险详情
// GET /api/v1/car/insurances/:id  （公开）
func (h *InsuranceHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeInsuranceNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// Quote 车险报价（C 端）
// POST /api/v1/car/insurances/quote  （需登录）
func (h *InsuranceHandler) Quote(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.InsuranceQuoteRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	resp, err := h.service.Quote(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeInsuranceError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// ===== M 端管理 =====

// AdminList 管理后台车险列表
// GET /api/v1/car/admin/insurances  （需 car:audit 权限）
func (h *InsuranceHandler) AdminList(ctx plugin.Context) {
	var req dto.InsuranceListRequest
	_ = ctx.Bind(&req)

	pagination, list, err := h.service.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeInsuranceError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdminGetByID 管理后台车险详情
// GET /api/v1/car/admin/insurances/:id  （需 car:audit 权限）
func (h *InsuranceHandler) AdminGetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeInsuranceNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// Create 创建车险
// POST /api/v1/car/admin/insurances  （需 car:audit 权限）
func (h *InsuranceHandler) Create(ctx plugin.Context) {
	var req dto.CreateInsuranceRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	info, err := h.service.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeInsuranceError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("车险创建成功", info))
}

// Update 更新车险
// PUT /api/v1/car/admin/insurances/:id  （需 car:audit 权限）
func (h *InsuranceHandler) Update(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	var req dto.UpdateInsuranceRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.service.Update(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeInsuranceError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除车险
// DELETE /api/v1/car/admin/insurances/:id  （需 car:audit 权限）
func (h *InsuranceHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	if err := h.service.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeInsuranceError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// AdminUpdateStatus 管理后台更新车险状态
// PUT /api/v1/car/admin/insurances/:id/status  （需 car:audit 权限）
func (h *InsuranceHandler) AdminUpdateStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	var req dto.AdminUpdateStatusRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.service.UpdateStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeInsuranceError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态更新成功", nil))
}
