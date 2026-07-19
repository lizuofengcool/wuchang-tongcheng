// Package handler 同城车辆买卖 HTTP 处理层 - 分期付款方案
// 依据 v3.2.1 架构方案：对标毛豆新车/易鑫车贷/瓜子金融
// Financing 为全局配置数据（无 region_id），管理后台维护
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

// FinancingHandler 分期方案 HTTP 处理器
type FinancingHandler struct {
	service service.FinancingService
}

// NewFinancingHandler 创建 FinancingHandler 实例
func NewFinancingHandler(svc service.FinancingService) *FinancingHandler {
	return &FinancingHandler{service: svc}
}

// ===== C 端 =====

// ListPublished 已发布的分期方案列表（C 端浏览）
// GET /api/v1/car/financings  （公开）
func (h *FinancingHandler) ListPublished(ctx plugin.Context) {
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListPublished(page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeFinancingError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListHot 热门分期方案
// GET /api/v1/car/financings/hot  （公开）
func (h *FinancingHandler) ListHot(ctx plugin.Context) {
	limitStr := ctx.Query("limit")
	limit := 10
	if n, err := strconv.Atoi(limitStr); err == nil && n > 0 && n <= 100 {
		limit = n
	}

	list, err := h.service.ListHot(limit)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeFinancingError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// GetByID 分期方案详情
// GET /api/v1/car/financings/:id  （公开）
func (h *FinancingHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeFinancingNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// Calculate 分期计算（C 端）
// POST /api/v1/car/financings/calculate  （需登录）
func (h *FinancingHandler) Calculate(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.FinancingCalculateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	resp, err := h.service.Calculate(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeFinancingError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// ===== M 端管理 =====

// AdminList 管理后台分期方案列表
// GET /api/v1/car/admin/financings  （需 car:audit 权限）
func (h *FinancingHandler) AdminList(ctx plugin.Context) {
	var req dto.FinancingListRequest
	_ = ctx.Bind(&req)

	pagination, list, err := h.service.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeFinancingError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdminGetByID 管理后台分期方案详情
// GET /api/v1/car/admin/financings/:id  （需 car:audit 权限）
func (h *FinancingHandler) AdminGetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeFinancingNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// Create 创建分期方案
// POST /api/v1/car/admin/financings  （需 car:audit 权限）
func (h *FinancingHandler) Create(ctx plugin.Context) {
	var req dto.CreateFinancingRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	info, err := h.service.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeFinancingError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("分期方案创建成功", info))
}

// Update 更新分期方案
// PUT /api/v1/car/admin/financings/:id  （需 car:audit 权限）
func (h *FinancingHandler) Update(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	var req dto.UpdateFinancingRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.service.Update(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeFinancingError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除分期方案
// DELETE /api/v1/car/admin/financings/:id  （需 car:audit 权限）
func (h *FinancingHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	if err := h.service.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeFinancingError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// AdminUpdateStatus 管理后台更新分期方案状态
// PUT /api/v1/car/admin/financings/:id/status  （需 car:audit 权限）
func (h *FinancingHandler) AdminUpdateStatus(ctx plugin.Context) {
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeFinancingError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态更新成功", nil))
}
