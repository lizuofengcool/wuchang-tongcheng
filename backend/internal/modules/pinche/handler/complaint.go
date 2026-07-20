// Package handler 同城拼车出行 HTTP 处理层 - 举报（基于 complaints 表）
// 前端 URL 为 /admin/reports，handler 内部使用 complaint 语义
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/pinche/dto"
	"wuchang-tongcheng/internal/modules/pinche/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ComplaintHandler 举报 HTTP 处理器
type ComplaintHandler struct {
	service service.ComplaintService
}

// NewComplaintHandler 创建 ComplaintHandler 实例
func NewComplaintHandler(svc service.ComplaintService) *ComplaintHandler {
	return &ComplaintHandler{service: svc}
}

// AdminList 举报列表
// GET /api/v1/pinche/admin/reports  （需 pinche:audit 权限）
func (h *ComplaintHandler) AdminList(ctx plugin.Context) {
	var req dto.ComplaintListRequest
	_ = ctx.Bind(&req)

	result, err := h.service.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(result))
}

// Create 创建举报
// POST /api/v1/pinche/admin/reports  （需 pinche:audit 权限）
func (h *ComplaintHandler) Create(ctx plugin.Context) {
	userID, username, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.CreateComplaintRequest
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
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("举报已创建", info))
}

// Process 处理举报
// PUT /api/v1/pinche/admin/reports/:id/process  （需 pinche:audit 权限）
func (h *ComplaintHandler) Process(ctx plugin.Context) {
	userID, username, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	var req dto.ProcessComplaintRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.service.Process(id, userID, username, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("处理成功", nil))
}

// Stats 举报统计
// GET /api/v1/pinche/admin/reports/stats  （需 pinche:audit 权限）
func (h *ComplaintHandler) Stats(ctx plugin.Context) {
	stats, err := h.service.Stats()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(stats))
}
