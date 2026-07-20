// Package handler love 相亲交友 HTTP 处理层 - 详细资料
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

// ProfileHandler 详细资料 HTTP 处理器
type ProfileHandler struct {
	service service.LoveProfileService
}

// NewProfileHandler 创建 ProfileHandler 实例
func NewProfileHandler(svc service.LoveProfileService) *ProfileHandler {
	return &ProfileHandler{service: svc}
}

// Create 创建详细资料
// POST /api/v1/love/profiles  （需登录）
func (h *ProfileHandler) Create(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	var req dto.CreateLoveProfileRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	loveIDStr := ctx.Query("love_id")
	loveID, _ := parseUint(loveIDStr)
	info, err := h.service.Create(loveID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// Update 更新详细资料
// PUT /api/v1/love/profiles/:id  （需登录）
func (h *ProfileHandler) Update(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateLoveProfileRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// UpdateByLoveID 按 loveID 更新详细资料（不存在则创建）
// PUT /api/v1/love/loves/:id/profile  （需登录）
func (h *ProfileHandler) UpdateByLoveID(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	loveID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateLoveProfileRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.UpdateByLoveID(loveID, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// GetByID 详细资料详情
// GET /api/v1/love/profiles/:id  （公开）
func (h *ProfileHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetByLoveID 按 loveID 查询详细资料
// GET /api/v1/love/loves/:id/profile  （公开）
func (h *ProfileHandler) GetByLoveID(ctx plugin.Context) {
	loveID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.service.GetByLoveID(loveID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetByUserID 查询当前登录用户的详细资料
// GET /api/v1/love/profiles/me  （需登录）
func (h *ProfileHandler) GetByUserID(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	info, err := h.service.GetByUserID(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 详细资料列表
// GET /api/v1/love/profiles  （公开）
func (h *ProfileHandler) List(ctx plugin.Context) {
	var req dto.LoveProfileListRequest
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// UpdateStep 更新资料填写步骤
// PUT /api/v1/love/loves/:id/profile/step  （需登录）
func (h *ProfileHandler) UpdateStep(ctx plugin.Context) {
	_, ok := requireLogin(ctx)
	if !ok {
		return
	}
	loveID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateProfileStepRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.UpdateStep(loveID, req.Step); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}
