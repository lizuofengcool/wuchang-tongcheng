// Package handler 同城零工兼职 HTTP 处理层 - 技能标签
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/linggong/dto"
	"wuchang-tongcheng/internal/modules/linggong/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// SkillHandler 技能标签 HTTP 处理器
type SkillHandler struct {
	service service.SkillService
}

// NewSkillHandler 创建 SkillHandler 实例
func NewSkillHandler(svc service.SkillService) *SkillHandler {
	return &SkillHandler{service: svc}
}

// ===== C 端只读 =====

// GetByID 技能详情
// GET /api/v1/linggong/skills/:id  （公开）
func (h *SkillHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongSkillNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 技能列表
// GET /api/v1/linggong/skills  （公开）
func (h *SkillHandler) List(ctx plugin.Context) {
	var req dto.SkillListRequest
	_ = ctx.Bind(&req)

	pagination, list, err := h.service.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByCategory 按分类查询
// GET /api/v1/linggong/skills/category/:category  （公开）
func (h *SkillHandler) ListByCategory(ctx plugin.Context) {
	category := ctx.Param("category")
	if category == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的分类"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListByCategory(category, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByParent 按父级查询
// GET /api/v1/linggong/skills/parent/:id  （公开）
func (h *SkillHandler) ListByParent(ctx plugin.Context) {
	parentID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的父级ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListByParent(parentID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListHot 热门技能
// GET /api/v1/linggong/skills/hot  （公开）
func (h *SkillHandler) ListHot(ctx plugin.Context) {
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListHot(page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ===== M 端管理 =====

// AdminList 管理后台技能列表
// GET /api/v1/linggong/admin/skills  （需 linggong:audit 权限）
func (h *SkillHandler) AdminList(ctx plugin.Context) {
	var req dto.SkillAdminListRequest
	_ = ctx.Bind(&req)

	listReq := dto.SkillListRequest{
		Category: req.Category,
		Status:   req.Status,
		Keyword:  req.Keyword,
	}
	listReq.Page = req.Page
	listReq.PageSize = req.PageSize
	pagination, list, err := h.service.List(&listReq)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Create 创建技能
// POST /api/v1/linggong/admin/skills  （需 linggong:audit 权限）
func (h *SkillHandler) Create(ctx plugin.Context) {
	var req dto.CreateSkillRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.service.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("技能创建成功", info))
}

// Update 更新技能
// PUT /api/v1/linggong/admin/skills/:id  （需 linggong:audit 权限）
func (h *SkillHandler) Update(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateSkillRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.Update(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongSkillNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除技能
// DELETE /api/v1/linggong/admin/skills/:id  （需 linggong:audit 权限）
func (h *SkillHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.service.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongSkillNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// AdminUpdateStatus 更新技能状态
// PUT /api/v1/linggong/admin/skills/:id/status  （需 linggong:audit 权限）
func (h *SkillHandler) AdminUpdateStatus(ctx plugin.Context) {
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
