// Package handler 同城车辆买卖 HTTP 处理层 - 车辆评估
// 依据 v3.2.1 架构方案：对标瓜子估值/人人车在线估价/懂车帝二手车估值
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/car/dto"
	"wuchang-tongcheng/internal/modules/car/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// EvaluationHandler 车辆评估 HTTP 处理器
type EvaluationHandler struct {
	service service.EvaluationService
}

// NewEvaluationHandler 创建 EvaluationHandler 实例
func NewEvaluationHandler(svc service.EvaluationService) *EvaluationHandler {
	return &EvaluationHandler{service: svc}
}

// ===== C 端 =====

// Create 创建评估
// POST /api/v1/car/evaluations  （需登录）
func (h *EvaluationHandler) Create(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.CreateEvaluationRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	regionID := getRegionID(ctx)
	info, err := h.service.Create(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeEvaluationError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("评估创建成功", info))
}

// Update 更新评估（仅创建者）
// PUT /api/v1/car/evaluations/:id  （需登录）
func (h *EvaluationHandler) Update(ctx plugin.Context) {
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

	var req dto.UpdateEvaluationRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.service.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeEvaluationError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除评估（仅创建者）
// DELETE /api/v1/car/evaluations/:id  （需登录）
func (h *EvaluationHandler) Delete(ctx plugin.Context) {
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeEvaluationError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 获取评估详情
// GET /api/v1/car/evaluations/:id  （公开）
func (h *EvaluationHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeEvaluationNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetByCarID 按车源获取评估
// GET /api/v1/car/cars/:id/evaluation  （公开）
func (h *EvaluationHandler) GetByCarID(ctx plugin.Context) {
	carID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的车源ID"))
		return
	}

	info, err := h.service.GetByCarID(carID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeEvaluationNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 评估列表
// GET /api/v1/car/evaluations  （公开）
func (h *EvaluationHandler) List(ctx plugin.Context) {
	var req dto.EvaluationListRequest
	_ = ctx.Bind(&req)

	regionID := getRegionID(ctx)
	pagination, list, err := h.service.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeEvaluationError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByEvaluator 我的评估任务
// GET /api/v1/car/evaluations/mine  （需登录）
func (h *EvaluationHandler) ListByEvaluator(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListByEvaluator(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeEvaluationError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByCarID 车源的所有评估历史
// GET /api/v1/car/cars/:id/evaluations  （公开）
func (h *EvaluationHandler) ListByCarID(ctx plugin.Context) {
	carID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的车源ID"))
		return
	}

	list, err := h.service.ListByCarID(carID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeEvaluationError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// OnlineEvaluate 在线估值（C 端免费估值工具）
// POST /api/v1/car/evaluations/online  （需登录）
func (h *EvaluationHandler) OnlineEvaluate(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.OnlineEvaluationRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	regionID := getRegionID(ctx)
	resp, err := h.service.OnlineEvaluate(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeEvaluationError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// ===== M 端管理 =====

// AdminList 管理后台评估列表
// GET /api/v1/car/admin/evaluations  （需 car:audit 权限）
func (h *EvaluationHandler) AdminList(ctx plugin.Context) {
	var req dto.EvaluationListRequest
	_ = ctx.Bind(&req)

	pagination, list, err := h.service.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeEvaluationError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdminGetByID 管理后台评估详情
// GET /api/v1/car/admin/evaluations/:id  （需 car:audit 权限）
func (h *EvaluationHandler) AdminGetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	info, err := h.service.AdminGetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeEvaluationNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// AdminUpdateStatus 管理后台更新评估状态
// PUT /api/v1/car/admin/evaluations/:id/status  （需 car:audit 权限）
func (h *EvaluationHandler) AdminUpdateStatus(ctx plugin.Context) {
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeEvaluationError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态更新成功", nil))
}
