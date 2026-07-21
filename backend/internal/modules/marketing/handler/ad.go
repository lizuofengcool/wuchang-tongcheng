// Package handler 营销活动中台 HTTP 处理层 - 广告位（ad 子域）
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/marketing/dto"
	"wuchang-tongcheng/internal/modules/marketing/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// 本模块错误码常量位于 handler/codes.go（5801-5830 区间）
// 通过 plugin.go 的 init() 注册到 utils.RegisterCode

// AdHandler 广告位 HTTP 处理器
type AdHandler struct {
	svc service.AdService
}

// NewAdHandler 创建广告位 Handler 实例
func NewAdHandler(svc service.AdService) *AdHandler {
	return &AdHandler{svc: svc}
}

// Create 创建广告位（M 端）
// POST /api/v1/marketing/ads  （需 marketing:manage 权限）
func (h *AdHandler) Create(ctx plugin.Context) {
	var req dto.CreateAdPositionRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.svc.Create(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingAdError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// Update 更新广告位
// PUT /api/v1/marketing/ads/:id  （需 marketing:manage 权限）
func (h *AdHandler) Update(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateAdPositionRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Update(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingAdError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除广告位
// DELETE /api/v1/marketing/ads/:id  （需 marketing:manage 权限）
func (h *AdHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingAdError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 广告位详情
// GET /api/v1/marketing/ads/:id  （公开）
func (h *AdHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingAdNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 广告位列表（M 端）
// GET /api/v1/marketing/ads  （需 marketing:manage 权限）
func (h *AdHandler) List(ctx plugin.Context) {
	var req dto.AdPositionListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	regionID := getRegionID(ctx)
	pagination, list, err := h.svc.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingAdError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByPositionCode 按位置编码获取广告（C 端）
// GET /api/v1/marketing/positions/:code/ads  （公开）
func (h *AdHandler) ListByPositionCode(ctx plugin.Context) {
	code := ctx.Param("code")
	if code == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("位置编码不能为空"))
		return
	}
	page, pageSize := parsePagination(ctx)
	regionID := getRegionID(ctx)
	pagination, list, err := h.svc.ListByPositionCode(regionID, code, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeMarketingAdError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}
