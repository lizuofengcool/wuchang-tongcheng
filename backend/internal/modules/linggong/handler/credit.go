// Package handler 同城零工兼职 HTTP 处理层 - 信用分
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/linggong/dto"
	"wuchang-tongcheng/internal/modules/linggong/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// CreditHandler 信用分 HTTP 处理器
type CreditHandler struct {
	service service.CreditService
}

// NewCreditHandler 创建 CreditHandler 实例
func NewCreditHandler(svc service.CreditService) *CreditHandler {
	return &CreditHandler{service: svc}
}

// ===== C 端 =====

// GetByID 信用变更记录详情
// GET /api/v1/linggong/credits/:id  （公开）
func (h *CreditHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongCreditNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 信用变更记录列表
// GET /api/v1/linggong/credits  （公开）
func (h *CreditHandler) List(ctx plugin.Context) {
	var req dto.CreditListRequest
	_ = ctx.Bind(&req)

	regionID := getRegionID(ctx)
	pagination, list, err := h.service.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByUser 按用户查询信用变更记录
// GET /api/v1/linggong/credits/user/:id  （公开）
func (h *CreditHandler) ListByUser(ctx plugin.Context) {
	userID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的用户ID"))
		return
	}
	userType := ctx.Query("user_type")
	if userType == "" {
		userType = "worker"
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListByUser(userID, userType, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// GetScore 查询信用分
// GET /api/v1/linggong/credits/score  （需登录）
func (h *CreditHandler) GetScore(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	userType := ctx.Query("user_type")
	if userType == "" {
		userType = "worker"
	}
	resp, err := h.service.GetScore(userID, userType)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongCreditNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// ===== M 端管理 =====

// Adjust 信用分调整（M 端）
// POST /api/v1/linggong/admin/credits/adjust  （需 linggong:audit 权限）
func (h *CreditHandler) Adjust(ctx plugin.Context) {
	operatorID, operatorName, _, _ := getUserProfile(ctx)
	var req dto.CreditAdjustRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.service.Adjust(regionID, operatorID, operatorName, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("信用分调整成功", info))
}

// Delete 删除信用变更记录
// DELETE /api/v1/linggong/admin/credits/:id  （需 linggong:audit 权限）
func (h *CreditHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.service.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongCreditNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}
