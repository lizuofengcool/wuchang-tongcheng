// Package handler 多租户分站 HTTP 处理层 - 配置
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/tenant/dto"
	"wuchang-tongcheng/internal/modules/tenant/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ConfigHandler 配置 HTTP 处理器
type ConfigHandler struct {
	svc service.ConfigService
}

// NewConfigHandler 创建配置 Handler 实例
func NewConfigHandler(svc service.ConfigService) *ConfigHandler {
	return &ConfigHandler{svc: svc}
}

// List 配置列表（admin）
// GET /api/v1/tenant/admin/configs
func (h *ConfigHandler) List(ctx plugin.Context) {
	var req dto.ConfigListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeTenantConfigError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// GetByID 配置详情（admin）
// GET /api/v1/tenant/admin/configs/:id
func (h *ConfigHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		invalidID(ctx)
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeTenantConfigNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// Upsert 新增/更新配置（admin）
// POST /api/v1/tenant/admin/configs
func (h *ConfigHandler) Upsert(ctx plugin.Context) {
	var req dto.UpsertConfigRequest
	if err := ctx.Bind(&req); err != nil {
		badRequest(ctx, "参数错误: "+err.Error())
		return
	}
	info, err := h.svc.Upsert(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeTenantConfigError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("配置保存成功", info))
}

// Update 更新配置值（admin）
// PUT /api/v1/tenant/admin/configs/:id
func (h *ConfigHandler) Update(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		invalidID(ctx)
		return
	}
	var req dto.UpdateConfigRequest
	if err := ctx.Bind(&req); err != nil {
		badRequest(ctx, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.Update(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeTenantConfigError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除配置（admin）
// DELETE /api/v1/tenant/admin/configs/:id
func (h *ConfigHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		invalidID(ctx)
		return
	}
	if err := h.svc.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeTenantConfigError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// ListByStationAndModule 按分站+模块查询配置（admin）
// GET /api/v1/tenant/admin/configs/by-station-module
func (h *ConfigHandler) ListByStationAndModule(ctx plugin.Context) {
	stationIDStr := ctx.Query("station_id")
	bizModule := ctx.Query("biz_module")
	if stationIDStr == "" || bizModule == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("station_id 和 biz_module 参数不能为空"))
		return
	}
	stationID, err := parseUint(stationIDStr)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的 station_id"))
		return
	}
	list, err := h.svc.ListByStationAndModule(stationID, bizModule)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeTenantConfigError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// BatchGet 批量获取配置（admin）
// POST /api/v1/tenant/admin/configs/batch-get
func (h *ConfigHandler) BatchGet(ctx plugin.Context) {
	var req dto.BatchGetConfigRequest
	if err := ctx.Bind(&req); err != nil {
		badRequest(ctx, "参数错误: "+err.Error())
		return
	}
	list, err := h.svc.BatchGet(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeTenantConfigError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// parseUint 解析字符串为 uint（内部辅助）
func parseUint(s string) (uint, error) {
	var n uint64
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, errInvalidUint
		}
		n = n*10 + uint64(c-'0')
	}
	return uint(n), nil
}

var errInvalidUint = &invalidUintError{}

type invalidUintError struct{}

func (e *invalidUintError) Error() string { return "invalid uint" }
