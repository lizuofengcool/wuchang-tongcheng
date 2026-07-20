// Package handler 同城零工兼职 HTTP 处理层 - 求职者档案
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/linggong/dto"
	"wuchang-tongcheng/internal/modules/linggong/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// WorkerHandler 求职者档案 HTTP 处理器
type WorkerHandler struct {
	service service.WorkerService
}

// NewWorkerHandler 创建 WorkerHandler 实例
func NewWorkerHandler(svc service.WorkerService) *WorkerHandler {
	return &WorkerHandler{service: svc}
}

// ===== C 端 =====

// Create 创建求职者档案
// POST /api/v1/linggong/workers  （需登录）
func (h *WorkerHandler) Create(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.CreateWorkerRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	regionID := getRegionID(ctx)
	info, err := h.service.Create(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("档案创建成功", info))
}

// Update 更新求职者档案（仅本人）
// PUT /api/v1/linggong/workers/:id  （需登录）
func (h *WorkerHandler) Update(ctx plugin.Context) {
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

	var req dto.UpdateWorkerRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.service.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongWorkerNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除求职者档案（仅本人）
// DELETE /api/v1/linggong/workers/:id  （需登录）
func (h *WorkerHandler) Delete(ctx plugin.Context) {
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongWorkerNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 求职者档案详情
// GET /api/v1/linggong/workers/:id  （公开）
func (h *WorkerHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongWorkerNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetByUserID 按用户ID查询本人档案
// GET /api/v1/linggong/workers/me  （需登录）
func (h *WorkerHandler) GetByUserID(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	info, err := h.service.GetByUserID(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongWorkerNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 求职者档案列表
// GET /api/v1/linggong/workers  （公开）
func (h *WorkerHandler) List(ctx plugin.Context) {
	var req dto.WorkerListRequest
	_ = ctx.Bind(&req)

	regionID := getRegionID(ctx)
	pagination, list, err := h.service.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ===== M 端管理 =====

// AdminList 管理后台求职者档案列表
// GET /api/v1/linggong/admin/workers  （需 linggong:audit 权限）
func (h *WorkerHandler) AdminList(ctx plugin.Context) {
	var req dto.WorkerAdminListRequest
	_ = ctx.Bind(&req)

	pagination, list, err := h.service.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Audit 审核求职者档案
// PUT /api/v1/linggong/admin/workers/:id/audit  （需 linggong:audit 权限）
func (h *WorkerHandler) Audit(ctx plugin.Context) {
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
	if err := h.service.Audit(id, req.Status, ""); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongAuditError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核完成", nil))
}

// AdminUpdateStatus 管理后台更新求职者档案状态
// PUT /api/v1/linggong/admin/workers/:id/status  （需 linggong:audit 权限）
func (h *WorkerHandler) AdminUpdateStatus(ctx plugin.Context) {
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongStatusInvalid, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态更新成功", nil))
}
