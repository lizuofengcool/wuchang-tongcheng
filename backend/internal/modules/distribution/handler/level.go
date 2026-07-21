// Package handler 分销合伙人中台 HTTP 处理层 - 合伙人等级
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/distribution/dto"
	"wuchang-tongcheng/internal/modules/distribution/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// LevelHandler 等级处理器
type LevelHandler struct {
	svc service.LevelService
}

// NewLevelHandler 创建 LevelHandler 实例
func NewLevelHandler(svc service.LevelService) *LevelHandler {
	return &LevelHandler{svc: svc}
}

// List 等级列表（公开只读）
// GET /api/v1/distribution/levels
func (h *LevelHandler) List(ctx plugin.Context) {
	var req dto.LevelListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListAll 全部启用等级（公开，用于下拉选择）
// GET /api/v1/distribution/levels/all
func (h *LevelHandler) ListAll(ctx plugin.Context) {
	list, err := h.svc.ListAll()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// GetByID 详情
// GET /api/v1/distribution/levels/:id
func (h *LevelHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, failParam("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionLevelNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ===== 管理后台 =====

// AdminList 管理后台列表
// GET /api/v1/distribution/admin/levels
func (h *LevelHandler) AdminList(ctx plugin.Context) {
	var req dto.LevelListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdminCreate 创建等级
// POST /api/v1/distribution/admin/levels
func (h *LevelHandler) AdminCreate(ctx plugin.Context) {
	var req dto.LevelCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	info, err := h.svc.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("等级已创建", info))
}

// AdminUpdate 更新等级
// PUT /api/v1/distribution/admin/levels/:id
func (h *LevelHandler) AdminUpdate(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, failParam("无效的ID"))
		return
	}
	var req dto.LevelUpdateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Update(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// AdminDelete 删除等级
// DELETE /api/v1/distribution/admin/levels/:id
func (h *LevelHandler) AdminDelete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, failParam("无效的ID"))
		return
	}
	if err := h.svc.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionLevelHasPartners, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// AdminCheckUpgrade 检查自动升级
// POST /api/v1/distribution/admin/levels/check-upgrade
func (h *LevelHandler) AdminCheckUpgrade(ctx plugin.Context) {
	var req dto.LevelCheckUpgradeRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	info, err := h.svc.CheckUpgrade(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}
