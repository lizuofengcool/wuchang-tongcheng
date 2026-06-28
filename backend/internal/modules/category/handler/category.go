// Package handler 分类信息HTTP处理层
package handler

import (
	"net/http"
	"strconv"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/category/dto"
	"wuchang-tongcheng/internal/modules/category/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// Handler 分类HTTP处理器
type Handler struct {
	service service.CategoryService
}

// NewHandler 创建分类处理器
func NewHandler(svc service.CategoryService) *Handler {
	return &Handler{service: svc}
}

// Create 创建分类
func (h *Handler) Create(ctx plugin.Context) {
	userID, _ := ctx.Get(middleware.ContextUserID)
	if userID == uint(0) {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.CreateCategoryRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}

	// 从上下文获取地区ID（由Region中间件注入）
	regionID := getRegionID(ctx)

	info, err := h.service.Create(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCategoryError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// Update 更新分类
func (h *Handler) Update(ctx plugin.Context) {
	userID, _ := ctx.Get(middleware.ContextUserID)
	if userID == uint(0) {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的分类ID"))
		return
	}

	var req dto.UpdateCategoryRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}

	if err := h.service.Update(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCategoryError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除分类
func (h *Handler) Delete(ctx plugin.Context) {
	userID, _ := ctx.Get(middleware.ContextUserID)
	if userID == uint(0) {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的分类ID"))
		return
	}

	if err := h.service.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCategoryError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 根据ID获取分类
func (h *Handler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的分类ID"))
		return
	}

	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCategoryNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetByParentID 根据父级ID获取子分类
func (h *Handler) GetByParentID(ctx plugin.Context) {
	parentIDStr := ctx.Query("parent_id")
	parentID, err := strconv.ParseUint(parentIDStr, 10, 32)
	if err != nil {
		parentID = 0
	}
	regionID := getRegionID(ctx)

	list, err := h.service.GetByParentID(uint(parentID), regionID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCategoryError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// GetTree 获取分类树形结构
func (h *Handler) GetTree(ctx plugin.Context) {
	regionID := getRegionID(ctx)
	tree, err := h.service.GetTree(regionID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCategoryError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(tree))
}

// GetAll 获取全部分类平铺列表（供 PC/小程序门户使用）
func (h *Handler) GetAll(ctx plugin.Context) {
	regionID := getRegionID(ctx)
	list, err := h.service.GetAll(regionID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCategoryError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// parseID 解析URL中的ID参数
func parseID(ctx plugin.Context) (uint, error) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// getRegionID 从上下文获取地区ID
func getRegionID(ctx plugin.Context) uint {
	if v, ok := ctx.Get(middleware.RegionIDKey); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return middleware.DefaultRegionID
}
