// Package handler 同城114 HTTP 处理层 - 分类
package handler

import (
	"net/http"
	"strconv"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/dh114/dto"
	"wuchang-tongcheng/internal/modules/dh114/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// CategoryHandler 分类 HTTP 处理器
type CategoryHandler struct {
	svc service.CategoryService
}

// NewCategoryHandler 创建分类 Handler 实例
func NewCategoryHandler(svc service.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

// Create 创建分类（M 端）
// POST /api/v1/dh114/categories  （需 content:audit 权限）
func (h *CategoryHandler) Create(ctx plugin.Context) {
	var req dto.CreateCategoryRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.svc.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114CategoryError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("分类创建成功", info))
}

// Update 更新分类
// PUT /api/v1/dh114/categories/:id  （需 content:audit 权限）
func (h *CategoryHandler) Update(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateCategoryRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Update(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114CategoryError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除分类
// DELETE /api/v1/dh114/categories/:id  （需 content:audit 权限）
func (h *CategoryHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114CategoryError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 分类详情
// GET /api/v1/dh114/categories/:id  （公开）
func (h *CategoryHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114CategoryNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 分类列表
// GET /api/v1/dh114/categories  （公开）
func (h *CategoryHandler) List(ctx plugin.Context) {
	var req dto.CategoryListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114CategoryError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByParent 按父级查询子分类
// GET /api/v1/dh114/categories/by-parent/:parent_id  （公开）
func (h *CategoryHandler) ListByParent(ctx plugin.Context) {
	parentID, err := parseSubID(ctx, "parent_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的父分类ID"))
		return
	}
	list, err := h.svc.ListByParent(parentID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114CategoryError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ListByLevel 按层级查询分类
// GET /api/v1/dh114/categories/by-level/:level  （公开）
func (h *CategoryHandler) ListByLevel(ctx plugin.Context) {
	levelStr := ctx.Param("level")
	level, err := strconv.Atoi(levelStr)
	if err != nil || level <= 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的层级"))
		return
	}
	list, err := h.svc.ListByLevel(level)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114CategoryError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ListByBusinessType 按业务类型查询分类
// GET /api/v1/dh114/categories/by-business-type  （公开）
func (h *CategoryHandler) ListByBusinessType(ctx plugin.Context) {
	businessType := ctx.Query("business_type")
	if businessType == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("business_type 参数不能为空"))
		return
	}
	list, err := h.svc.ListByBusinessType(businessType)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114CategoryError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// UpdateStatus 更新分类状态
// PUT /api/v1/dh114/categories/:id/status  （需 content:audit 权限）
func (h *CategoryHandler) UpdateStatus(ctx plugin.Context) {
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
	if err := h.svc.UpdateStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114CategoryError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态已更新", nil))
}
