// Package handler 地区HTTP处理层
package handler

import (
	"net/http"
	"strconv"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/region/dto"
	"wuchang-tongcheng/internal/modules/region/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// Handler 地区HTTP处理器
type Handler struct {
	service service.RegionService
}

// NewHandler 创建地区处理器
func NewHandler(svc service.RegionService) *Handler {
	return &Handler{service: svc}
}

// Create 创建地区
func (h *Handler) Create(ctx plugin.Context) {
	userID, _ := ctx.Get(middleware.ContextUserID)
	if userID == uint(0) {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.CreateRegionRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}

	info, err := h.service.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeRegionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// Update 更新地区
func (h *Handler) Update(ctx plugin.Context) {
	userID, _ := ctx.Get(middleware.ContextUserID)
	if userID == uint(0) {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的地区ID"))
		return
	}

	var req dto.UpdateRegionRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}

	if err := h.service.Update(uint(id), &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeRegionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除地区
func (h *Handler) Delete(ctx plugin.Context) {
	userID, _ := ctx.Get(middleware.ContextUserID)
	if userID == uint(0) {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的地区ID"))
		return
	}

	if err := h.service.Delete(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeRegionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 根据ID获取地区
func (h *Handler) GetByID(ctx plugin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的地区ID"))
		return
	}

	info, err := h.service.GetByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeRegionNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetByParentID 根据父级ID获取子地区
func (h *Handler) GetByParentID(ctx plugin.Context) {
	parentIDStr := ctx.Query("parent_id")
	parentID, err := strconv.ParseUint(parentIDStr, 10, 32)
	if err != nil {
		parentID = 0
	}

	list, err := h.service.GetByParentID(uint(parentID))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeRegionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// GetTree 获取地区树形结构
func (h *Handler) GetTree(ctx plugin.Context) {
	tree, err := h.service.GetTree()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeRegionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(tree))
}
