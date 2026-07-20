// Package handler love 相亲交友 HTTP 处理层 - 动态广场
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

// StoryHandler 动态 HTTP 处理器
type StoryHandler struct {
	service service.LoveStoryService
}

// NewStoryHandler 创建 StoryHandler 实例
func NewStoryHandler(svc service.LoveStoryService) *StoryHandler {
	return &StoryHandler{service: svc}
}

// Create 发布动态
// POST /api/v1/love/stories  （需登录）
func (h *StoryHandler) Create(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	var req dto.CreateLoveStoryRequest
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
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("发布成功", info))
}

// Update 更新动态
// PUT /api/v1/love/stories/:id  （需登录）
func (h *StoryHandler) Update(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateLoveStoryRequest
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

// Delete 删除动态
// DELETE /api/v1/love/stories/:id  （需登录）
func (h *StoryHandler) Delete(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.service.Delete(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 动态详情
// GET /api/v1/love/stories/:id  （公开）
func (h *StoryHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	userID, _, _ := getUserProfile(ctx)
	info, err := h.service.GetByID(id, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 动态列表
// GET /api/v1/love/stories  （公开）
func (h *StoryHandler) List(ctx plugin.Context) {
	var req dto.LoveStoryListRequest
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByLoveID 指定用户的动态列表
// GET /api/v1/love/loves/:id/stories  （公开）
func (h *StoryHandler) ListByLoveID(ctx plugin.Context) {
	loveID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListByLoveID(loveID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByUser 当前用户的动态列表
// GET /api/v1/love/stories/mine  （需登录）
func (h *StoryHandler) ListByUser(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListByUser(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListFeatured 精选动态
// GET /api/v1/love/stories/featured  （公开）
func (h *StoryHandler) ListFeatured(ctx plugin.Context) {
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListFeatured(page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByTopic 按话题查询动态
// GET /api/v1/love/stories/topic/:topic  （公开）
func (h *StoryHandler) ListByTopic(ctx plugin.Context) {
	topic := ctx.Param("topic")
	if topic == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("topic 不能为空"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListByTopic(topic, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// IncrView 增加浏览数
// POST /api/v1/love/stories/:id/view  （公开）
func (h *StoryHandler) IncrView(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	_ = h.service.IncrView(id)
	ctx.JSON(http.StatusOK, response.Success(nil))
}

// IncrLike 点赞
// POST /api/v1/love/stories/:id/like  （需登录）
func (h *StoryHandler) IncrLike(ctx plugin.Context) {
	_, ok := requireLogin(ctx)
	if !ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	_ = h.service.IncrLike(id)
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("点赞成功", nil))
}

// DecrLike 取消点赞
// POST /api/v1/love/stories/:id/unlike  （需登录）
func (h *StoryHandler) DecrLike(ctx plugin.Context) {
	_, ok := requireLogin(ctx)
	if !ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	_ = h.service.DecrLike(id)
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("取消点赞成功", nil))
}

// IncrComment 增加评论数
// POST /api/v1/love/stories/:id/comment  （需登录）
func (h *StoryHandler) IncrComment(ctx plugin.Context) {
	_, ok := requireLogin(ctx)
	if !ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	_ = h.service.IncrComment(id)
	ctx.JSON(http.StatusOK, response.Success(nil))
}

// IncrShare 增加分享数
// POST /api/v1/love/stories/:id/share  （公开）
func (h *StoryHandler) IncrShare(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	_ = h.service.IncrShare(id)
	ctx.JSON(http.StatusOK, response.Success(nil))
}

// ===== M 端 =====

// AdminList 管理后台：动态列表
// GET /api/v1/love/admin/stories  （需 love:audit 权限）
func (h *StoryHandler) AdminList(ctx plugin.Context) {
	var req dto.LoveStoryListRequest
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Audit 管理后台：审核动态
// PUT /api/v1/love/admin/stories/:id/audit  （需 love:audit 权限）
func (h *StoryHandler) Audit(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req struct {
		AuditStatus int    `json:"audit_status" binding:"required,oneof=1 2 3"`
		AuditReason string `json:"audit_reason"`
	}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.Audit(id, req.AuditStatus, req.AuditReason); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核成功", nil))
}

// UpdateStatus 管理后台：更新动态状态
// PUT /api/v1/love/admin/stories/:id/status  （需 love:audit 权限）
func (h *StoryHandler) UpdateStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req struct {
		Status int `json:"status" binding:"required"`
	}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.UpdateStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// SetFeatured 设置精选
// PUT /api/v1/love/admin/stories/:id/featured  （需 love:audit 权限）
func (h *StoryHandler) SetFeatured(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req struct {
		Featured bool `json:"featured"`
	}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.SetFeatured(id, req.Featured); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("设置成功", nil))
}

// BatchAudit 批量审核
// PUT /api/v1/love/admin/stories/batch-audit  （需 love:audit 权限）
func (h *StoryHandler) BatchAudit(ctx plugin.Context) {
	var req struct {
		IDs         []uint `json:"ids" binding:"required,min=1"`
		AuditStatus int    `json:"audit_status" binding:"required,oneof=1 2 3"`
		AuditReason string `json:"audit_reason"`
	}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.BatchAudit(req.IDs, req.AuditStatus, req.AuditReason); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("批量审核成功", nil))
}
