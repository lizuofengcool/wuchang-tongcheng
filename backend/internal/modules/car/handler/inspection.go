// Package handler 同城车辆买卖 HTTP 处理层 - 车况检测
// 依据 v3.2.1 架构方案：对标瓜子严选检测/人人车检测报告
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/car/dto"
	"wuchang-tongcheng/internal/modules/car/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// InspectionHandler 车况检测 HTTP 处理器
type InspectionHandler struct {
	service service.InspectionService
}

// NewInspectionHandler 创建 InspectionHandler 实例
func NewInspectionHandler(svc service.InspectionService) *InspectionHandler {
	return &InspectionHandler{service: svc}
}

// ===== C 端 =====

// Create 创建检测
// POST /api/v1/car/inspections  （需登录）
func (h *InspectionHandler) Create(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.CreateInspectionRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	regionID := getRegionID(ctx)
	info, err := h.service.Create(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeInspectionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("检测创建成功", info))
}

// Update 更新检测（仅创建者）
// PUT /api/v1/car/inspections/:id  （需登录）
func (h *InspectionHandler) Update(ctx plugin.Context) {
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

	var req dto.UpdateInspectionRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.service.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeInspectionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除检测（仅创建者）
// DELETE /api/v1/car/inspections/:id  （需登录）
func (h *InspectionHandler) Delete(ctx plugin.Context) {
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeInspectionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 获取检测详情
// GET /api/v1/car/inspections/:id  （公开）
func (h *InspectionHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeInspectionNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetByCarID 按车源获取检测
// GET /api/v1/car/cars/:id/inspection  （公开）
func (h *InspectionHandler) GetByCarID(ctx plugin.Context) {
	carID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的车源ID"))
		return
	}

	info, err := h.service.GetByCarID(carID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeInspectionNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 检测列表
// GET /api/v1/car/inspections  （公开）
func (h *InspectionHandler) List(ctx plugin.Context) {
	var req dto.InspectionListRequest
	_ = ctx.Bind(&req)

	regionID := getRegionID(ctx)
	pagination, list, err := h.service.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeInspectionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByInspector 我的检测任务
// GET /api/v1/car/inspections/mine  （需登录）
func (h *InspectionHandler) ListByInspector(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListByInspector(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeInspectionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ===== M 端管理 =====

// AdminList 管理后台检测列表
// GET /api/v1/car/admin/inspections  （需 car:audit 权限）
func (h *InspectionHandler) AdminList(ctx plugin.Context) {
	var req dto.InspectionListRequest
	_ = ctx.Bind(&req)

	pagination, list, err := h.service.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeInspectionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdminGetByID 管理后台检测详情
// GET /api/v1/car/admin/inspections/:id  （需 car:audit 权限）
func (h *InspectionHandler) AdminGetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	info, err := h.service.AdminGetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeInspectionNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// Review 检测复核
// PUT /api/v1/car/admin/inspections/:id/review  （需 car:audit 权限）
func (h *InspectionHandler) Review(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	var req dto.InspectionReviewRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.service.Review(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeInspectionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("复核完成", nil))
}

// AdminUpdateStatus 管理后台更新检测状态
// PUT /api/v1/car/admin/inspections/:id/status  （需 car:audit 权限）
func (h *InspectionHandler) AdminUpdateStatus(ctx plugin.Context) {
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeInspectionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态更新成功", nil))
}
