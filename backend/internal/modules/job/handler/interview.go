// Package handler 面试邀约 HTTP 处理层
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/job/dto"
	"wuchang-tongcheng/internal/modules/job/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// InterviewHandler 面试邀约 HTTP 处理器
type InterviewHandler struct {
	svc service.InterviewService
}

// NewInterviewHandler 创建面试邀约 Handler 实例
func NewInterviewHandler(svc service.InterviewService) *InterviewHandler {
	return &InterviewHandler{svc: svc}
}

// Create 创建面试邀约
// POST /api/v1/job/interviews
func (h *InterviewHandler) Create(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.InterviewCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.svc.Create(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeInterviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建面试邀约成功", info))
}

// Update 更新面试
// PUT /api/v1/job/interviews/:id
func (h *InterviewHandler) Update(ctx plugin.Context) {
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
	var req dto.InterviewUpdateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeInterviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// GetByID 面试详情
// GET /api/v1/job/interviews/:id
func (h *InterviewHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	userID, _, _, _ := getUserProfile(ctx)
	info, err := h.svc.GetByID(id, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeInterviewNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 面试列表
// GET /api/v1/job/interviews
func (h *InterviewHandler) List(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.InterviewListQuery
	_ = ctx.Bind(&req)
	pagination, list, err := h.svc.List(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeInterviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByApplication 投递记录的面试列表
// GET /api/v1/job/applications/:id/interviews
func (h *InterviewHandler) ListByApplication(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	appID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	list, err := h.svc.ListByApplication(appID, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeInterviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// Action 面试状态变更
// PUT /api/v1/job/interviews/:id/action
func (h *InterviewHandler) Action(ctx plugin.Context) {
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
	var req dto.InterviewActionRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.svc.Action(id, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeInterviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// Feedback 面试反馈
// PUT /api/v1/job/interviews/:id/feedback
func (h *InterviewHandler) Feedback(ctx plugin.Context) {
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
	var req dto.InterviewFeedbackRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.svc.Feedback(id, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeInterviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// Stats 面试统计
// GET /api/v1/job/interviews/stats
func (h *InterviewHandler) Stats(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	role := ctx.DefaultQuery("role", "applicant")
	stats, err := h.svc.Stats(userID, role)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeInterviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(stats))
}
