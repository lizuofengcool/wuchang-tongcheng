// Package handler 简历 HTTP 处理层
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/job/dto"
	"wuchang-tongcheng/internal/modules/job/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ResumeHandler 简历 HTTP 处理器
type ResumeHandler struct {
	svc service.ResumeService
}

// NewResumeHandler 创建简历 Handler 实例
func NewResumeHandler(svc service.ResumeService) *ResumeHandler {
	return &ResumeHandler{svc: svc}
}

// Create 创建简历
// POST /api/v1/job/resumes
func (h *ResumeHandler) Create(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.ResumeCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.svc.Create(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeResumeError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// Update 更新简历
// PUT /api/v1/job/resumes/:id
func (h *ResumeHandler) Update(ctx plugin.Context) {
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
	var req dto.ResumeUpdateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeResumeError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除简历
// DELETE /api/v1/job/resumes/:id
func (h *ResumeHandler) Delete(ctx plugin.Context) {
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
	if err := h.svc.Delete(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeResumeError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 简历详情
// GET /api/v1/job/resumes/:id
func (h *ResumeHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	userID, _, _, _ := getUserProfile(ctx)
	info, err := h.svc.GetByID(id, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeResumeNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 简历列表
// GET /api/v1/job/resumes
func (h *ResumeHandler) List(ctx plugin.Context) {
	var req dto.ResumeListQuery
	_ = ctx.Bind(&req)
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeResumeError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListMine 我的简历
// GET /api/v1/job/resumes/mine
func (h *ResumeHandler) ListMine(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListMine(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeResumeError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// GetDefault 默认简历
// GET /api/v1/job/resumes/default
func (h *ResumeHandler) GetDefault(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	info, err := h.svc.GetDefault(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeResumeNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// SetDefault 设为默认简历
// PUT /api/v1/job/resumes/:id/default
func (h *ResumeHandler) SetDefault(ctx plugin.Context) {
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
	if err := h.svc.SetDefault(userID, id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeResumeError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("设置成功", nil))
}

// UpdateStatus 修改简历状态
// PUT /api/v1/job/resumes/:id/status
func (h *ResumeHandler) UpdateStatus(ctx plugin.Context) {
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
	var req dto.AdminUpdateStatusRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.UpdateStatus(id, userID, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeResumeError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("操作成功", nil))
}
