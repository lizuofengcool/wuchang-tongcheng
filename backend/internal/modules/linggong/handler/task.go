// Package handler 同城零工兼职 HTTP 处理层 - 任务包
// 对标斗米任务制 + 猪八戒威客
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/linggong/dto"
	"wuchang-tongcheng/internal/modules/linggong/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// TaskHandler 任务包 HTTP 处理器
type TaskHandler struct {
	service service.TaskService
}

// NewTaskHandler 创建 TaskHandler 实例
func NewTaskHandler(svc service.TaskService) *TaskHandler {
	return &TaskHandler{service: svc}
}

// ===== C 端 =====

// Create 创建任务包
// POST /api/v1/linggong/tasks  （需登录）
func (h *TaskHandler) Create(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.CreateTaskRequest
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
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("任务包创建成功", info))
}

// Update 更新任务包（仅发布者本人）
// PUT /api/v1/linggong/tasks/:id  （需登录）
func (h *TaskHandler) Update(ctx plugin.Context) {
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

	var req dto.UpdateTaskRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.service.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongTaskNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除任务包（仅发布者本人）
// DELETE /api/v1/linggong/tasks/:id  （需登录）
func (h *TaskHandler) Delete(ctx plugin.Context) {
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongTaskNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 任务包详情
// GET /api/v1/linggong/tasks/:id  （公开）
func (h *TaskHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongTaskNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 任务包列表
// GET /api/v1/linggong/tasks  （公开）
func (h *TaskHandler) List(ctx plugin.Context) {
	var req dto.TaskListRequest
	_ = ctx.Bind(&req)

	pagination, list, err := h.service.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByLinggong 按岗位查询任务包
// GET /api/v1/linggong/:id/tasks  （公开）
func (h *TaskHandler) ListByLinggong(ctx plugin.Context) {
	linggongID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的岗位ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListByLinggong(linggongID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByEmployer 按雇主查询任务包
// GET /api/v1/linggong/tasks/employer/:id  （公开）
func (h *TaskHandler) ListByEmployer(ctx plugin.Context) {
	employerID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的雇主ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListByEmployer(employerID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Claim 领取任务
// POST /api/v1/linggong/tasks/:id/claim  （需登录）
func (h *TaskHandler) Claim(ctx plugin.Context) {
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

	var req dto.TaskClaimRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.service.Claim(id, userID, username, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongTaskClaimError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("任务领取成功", nil))
}

// Submit 提交任务交付
// POST /api/v1/linggong/tasks/:id/submit  （需登录）
func (h *TaskHandler) Submit(ctx plugin.Context) {
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

	var req dto.TaskSubmitRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.service.Submit(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongTaskClaimError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("任务交付成功", nil))
}

// Verify 任务验收
// POST /api/v1/linggong/tasks/:id/verify  （需登录）
func (h *TaskHandler) Verify(ctx plugin.Context) {
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

	var req dto.TaskVerifyRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.service.Verify(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongTaskClaimError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("任务验收完成", nil))
}

// ===== M 端管理 =====

// AdminList 管理后台任务包列表
// GET /api/v1/linggong/admin/tasks  （需 linggong:audit 权限）
func (h *TaskHandler) AdminList(ctx plugin.Context) {
	var req dto.TaskAdminListRequest
	_ = ctx.Bind(&req)

	pagination, list, err := h.service.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdminGetByID 管理后台任务包详情
// GET /api/v1/linggong/admin/tasks/:id  （需 linggong:audit 权限）
func (h *TaskHandler) AdminGetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.service.AdminGetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongTaskNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// AdminUpdateStatus 管理后台任务状态变更
// PUT /api/v1/linggong/admin/tasks/:id/status  （需 linggong:audit 权限）
func (h *TaskHandler) AdminUpdateStatus(ctx plugin.Context) {
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
