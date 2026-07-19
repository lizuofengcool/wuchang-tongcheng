// Package handler 批量操作 HTTP 处理层
// 依据 v3.2.1 架构方案：M 端批量审核/状态变更/删除/导出
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/ershou/dto"
	"wuchang-tongcheng/internal/modules/ershou/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// BatchHandler 批量操作 HTTP 处理器
type BatchHandler struct {
	service service.BatchService
}

// NewBatchHandler 创建批量操作 Handler 实例
func NewBatchHandler(svc service.BatchService) *BatchHandler {
	return &BatchHandler{service: svc}
}

// Audit 批量审核（M 端）
// POST /api/v1/ershou/batch/audit  （需登录 + content:audit）
func (h *BatchHandler) Audit(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.BatchAuditRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.service.Audit(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouAuditError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("批量审核完成", resp))
}

// UpdateStatus 批量状态变更（M 端）
// POST /api/v1/ershou/batch/status  （需登录 + content:audit）
func (h *BatchHandler) UpdateStatus(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.BatchStatusUpdateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.service.UpdateStatus(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("批量状态变更完成", resp))
}

// Delete 批量删除（M 端软删除）
// POST /api/v1/ershou/batch/delete  （需登录 + content:audit）
func (h *BatchHandler) Delete(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.BatchDeleteRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.service.Delete(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("批量删除完成", resp))
}

// Export 导出 Excel/CSV（返回行数据，由前端转为文件）
// POST /api/v1/ershou/batch/export  （需登录 + content:audit）
func (h *BatchHandler) Export(ctx plugin.Context) {
	var req dto.ExportRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	rows, err := h.service.Export(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(rows))
}
