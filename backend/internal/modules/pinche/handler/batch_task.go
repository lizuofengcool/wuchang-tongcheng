// Package handler 同城拼车出行 HTTP 处理层 - 批量任务
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/pinche/dto"
	"wuchang-tongcheng/internal/modules/pinche/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// BatchTaskHandler 批量任务 HTTP 处理器
type BatchTaskHandler struct {
	service service.BatchTaskService
}

// NewBatchTaskHandler 创建 BatchTaskHandler 实例
func NewBatchTaskHandler(svc service.BatchTaskService) *BatchTaskHandler {
	return &BatchTaskHandler{service: svc}
}

// AdminList 批量任务列表
// GET /api/v1/pinche/admin/batch-tasks  （需 pinche:audit 权限）
func (h *BatchTaskHandler) AdminList(ctx plugin.Context) {
	var req dto.BatchTaskListRequest
	_ = ctx.Bind(&req)

	result, err := h.service.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(result))
}

// AdminGetByID 批量任务详情
// GET /api/v1/pinche/admin/batch-tasks/:id  （需 pinche:audit 权限）
func (h *BatchTaskHandler) AdminGetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.service.AdminGetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// Create 创建批量任务
// POST /api/v1/pinche/admin/batch-tasks  （需 pinche:audit 权限）
func (h *BatchTaskHandler) Create(ctx plugin.Context) {
	userID, username, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.CreateBatchTaskRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	regionID := getRegionID(ctx)
	info, err := h.service.Create(regionID, userID, username, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("批量任务已创建", info))
}

// Cancel 取消批量任务
// PUT /api/v1/pinche/admin/batch-tasks/:id/cancel  （需 pinche:audit 权限）
func (h *BatchTaskHandler) Cancel(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.service.Cancel(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("任务已取消", nil))
}

// Retry 重试批量任务
// POST /api/v1/pinche/admin/batch-tasks/:id/retry  （需 pinche:audit 权限）
func (h *BatchTaskHandler) Retry(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.service.Retry(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("任务已重新加入队列", nil))
}

// Delete 删除批量任务
// DELETE /api/v1/pinche/admin/batch-tasks/:id  （需 pinche:audit 权限）
func (h *BatchTaskHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.service.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("任务已删除", nil))
}

// PreviewIDs 预览目标 ID 列表
// GET /api/v1/pinche/admin/batch-tasks/preview-ids  （需 pinche:audit 权限）
func (h *BatchTaskHandler) PreviewIDs(ctx plugin.Context) {
	var req dto.PreviewIDsRequest
	_ = ctx.Bind(&req)

	resp, err := h.service.PreviewIDs(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}
