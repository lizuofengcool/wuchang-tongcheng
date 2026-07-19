// Package module 模块注册表 HTTP 处理层
package module

import (
	"net/http"
	"strings"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/pkg/utils"
)

// Handler 模块管理 HTTP 处理器
type Handler struct {
	service Service
}

// NewHandler 创建模块处理器
func NewHandler(svc Service) *Handler {
	return &Handler{service: svc}
}

// List 模块列表
// GET /api/v1/modules
func (h *Handler) List(ctx plugin.Context) {
	modules, err := h.service.ListModules()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDBError, "查询模块列表失败"))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(modules))
}

// GetByName 模块详情
// GET /api/v1/modules/:name
func (h *Handler) GetByName(ctx plugin.Context) {
	name := strings.TrimSpace(ctx.Param("name"))
	if name == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("模块名称不能为空"))
		return
	}
	info, err := h.service.GetModule(name)
	if err != nil {
		ctx.JSON(http.StatusOK, response.NotFound(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// Enable 启用模块
// POST /api/v1/modules/:name/enable
func (h *Handler) Enable(ctx plugin.Context) {
	name := strings.TrimSpace(ctx.Param("name"))
	if name == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("模块名称不能为空"))
		return
	}
	if err := h.service.EnableModule(name); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDBError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("模块已启用", nil))
}

// Disable 禁用模块
// POST /api/v1/modules/:name/disable
func (h *Handler) Disable(ctx plugin.Context) {
	name := strings.TrimSpace(ctx.Param("name"))
	if name == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("模块名称不能为空"))
		return
	}
	if err := h.service.DisableModule(name); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDBError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("模块已禁用", nil))
}

// Update 更新模块元信息
// PUT /api/v1/modules/:name
func (h *Handler) Update(ctx plugin.Context) {
	name := strings.TrimSpace(ctx.Param("name"))
	if name == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("模块名称不能为空"))
		return
	}
	var req UpdateModuleRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}
	if err := h.service.UpdateModule(name, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDBError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}
