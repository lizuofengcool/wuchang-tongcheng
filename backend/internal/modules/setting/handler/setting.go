// Package handler 系统设置HTTP处理层
package handler

import (
	"net/http"
	"strconv"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/setting/dto"
	"wuchang-tongcheng/internal/modules/setting/service"
)

// Handler 系统设置HTTP处理器
type Handler struct {
	service service.SettingService
}

// NewHandler 创建系统设置处理器
func NewHandler(svc service.SettingService) *Handler {
	return &Handler{service: svc}
}

func getUserID(ctx plugin.Context) uint {
	if v, ok := ctx.Get(middleware.ContextUserID); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}

func getRegionID(ctx plugin.Context) uint {
	if v, ok := ctx.Get(middleware.RegionIDKey); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return middleware.DefaultRegionID
}

func parseID(ctx plugin.Context) (uint, error) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// Create 创建配置
func (h *Handler) Create(ctx plugin.Context) {
	if getUserID(ctx) == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.CreateSettingRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}
	info, err := h.service.Create(getRegionID(ctx), &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(1007, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// Update 更新配置
func (h *Handler) Update(ctx plugin.Context) {
	if getUserID(ctx) == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的配置ID"))
		return
	}
	var req dto.UpdateSettingRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}
	if err := h.service.Update(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(1001, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除配置
func (h *Handler) Delete(ctx plugin.Context) {
	if getUserID(ctx) == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的配置ID"))
		return
	}
	if err := h.service.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(1001, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 获取配置
func (h *Handler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的配置ID"))
		return
	}
	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(1006, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetByGroup 按分组获取配置
func (h *Handler) GetByGroup(ctx plugin.Context) {
	group := ctx.Param("group")
	if group == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("缺少分组参数"))
		return
	}
	list, err := h.service.GetByGroup(group, getRegionID(ctx))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(1001, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// GetAll 获取所有配置（按group分组）
func (h *Handler) GetAll(ctx plugin.Context) {
	result, err := h.service.GetAll(getRegionID(ctx))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(1001, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(result))
}

// BatchUpdate 批量更新
func (h *Handler) BatchUpdate(ctx plugin.Context) {
	if getUserID(ctx) == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.BatchUpdateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}
	if err := h.service.BatchUpdate(getRegionID(ctx), &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(1001, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("批量更新成功", nil))
}
