// Package handler love 相亲交友 HTTP 处理层 - 隐私设置
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/love/dto"
	"wuchang-tongcheng/internal/modules/love/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// PrivacyHandler 隐私设置 HTTP 处理器
type PrivacyHandler struct {
	service service.LovePrivacyService
}

// NewPrivacyHandler 创建 PrivacyHandler 实例
func NewPrivacyHandler(svc service.LovePrivacyService) *PrivacyHandler {
	return &PrivacyHandler{service: svc}
}

// Get 获取当前用户隐私设置
// GET /api/v1/love/privacy  （需登录）
func (h *PrivacyHandler) Get(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	info, err := h.service.Get(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLovePrivacyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetByLoveID 按 loveID 查询隐私设置
// GET /api/v1/love/loves/:id/privacy  （公开）
func (h *PrivacyHandler) GetByLoveID(ctx plugin.Context) {
	loveID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.service.GetByLoveID(loveID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLovePrivacyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// Update 更新隐私设置（不存在则创建）
// PUT /api/v1/love/privacy  （需登录）
func (h *PrivacyHandler) Update(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	var req dto.UpdateLovePrivacySettingRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	loveIDStr := ctx.Query("love_id")
	loveID, _ := parseUint(loveIDStr)
	info, err := h.service.Update(userID, loveID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLovePrivacyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", info))
}

// Reset 重置为默认
// POST /api/v1/love/privacy/reset  （需登录）
func (h *PrivacyHandler) Reset(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	if err := h.service.Reset(userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLovePrivacyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("重置成功", nil))
}

// IsVisible 查询字段可见性
// GET /api/v1/love/privacy/visible  （公开）
// query: field=online|location|age|...&user_id=123
func (h *PrivacyHandler) IsVisible(ctx plugin.Context) {
	field := ctx.Query("field")
	if field == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("field 不能为空"))
		return
	}
	userIDStr := ctx.Query("user_id")
	userID, err := parseUint(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("user_id 无效"))
		return
	}
	visible, err := h.service.IsVisible(field, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLovePrivacyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(map[string]bool{"visible": visible}))
}
