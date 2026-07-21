// Package handler LBS地图中台 HTTP 处理层 - 区域分站
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/lbs/dto"
	"wuchang-tongcheng/internal/modules/lbs/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// RegionHandler 区域分站处理器
type RegionHandler struct {
	svc service.RegionService
}

// NewRegionHandler 创建 Region Handler 实例
func NewRegionHandler(svc service.RegionService) *RegionHandler {
	return &RegionHandler{svc: svc}
}

// Create 创建区域
// POST /api/v1/lbs/regions  （需登录）
func (h *RegionHandler) Create(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.CreateRegionRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.svc.Create(&req)
	if err != nil {
		failByError(ctx, CodeLBSRegionCreateError, err)
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// Update 更新区域
// PUT /api/v1/lbs/regions/:id  （需登录）
func (h *RegionHandler) Update(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateRegionRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Update(id, &req); err != nil {
		failByError(ctx, CodeLBSRegionCreateError, err)
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除区域
// DELETE /api/v1/lbs/regions/:id  （需登录）
func (h *RegionHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Delete(id); err != nil {
		failByError(ctx, CodeLBSRegionNotFound, err)
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 获取区域详情
// GET /api/v1/lbs/regions/:id  （公开）
func (h *RegionHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		failByError(ctx, CodeLBSRegionNotFound, err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 区域列表
// GET /api/v1/lbs/regions  （公开）
func (h *RegionHandler) List(ctx plugin.Context) {
	var req dto.RegionListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		failByError(ctx, CodeLBSRegionListError, err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByParent 按父级列出
// GET /api/v1/lbs/regions/by-parent/:parent_id  （公开）
func (h *RegionHandler) ListByParent(ctx plugin.Context) {
	parentID, err := parseSubID(ctx, "parent_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的父级ID"))
		return
	}
	list, err := h.svc.ListByParent(parentID)
	if err != nil {
		failByError(ctx, CodeLBSRegionListError, err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// FindByCityCode 根据城市编码查找
// GET /api/v1/lbs/regions/by-city-code?city_code=  （公开）
func (h *RegionHandler) FindByCityCode(ctx plugin.Context) {
	cityCode := ctx.Query("city_code")
	if cityCode == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("city_code 不能为空"))
		return
	}
	info, err := h.svc.FindByCityCode(cityCode)
	if err != nil {
		failByError(ctx, CodeLBSRegionNotFound, err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// FindByLocation 根据经纬度判断分站
// GET /api/v1/lbs/regions/by-location?latitude=&longitude=  （公开）
func (h *RegionHandler) FindByLocation(ctx plugin.Context) {
	var req dto.RegionQueryRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if req.Latitude == 0 || req.Longitude == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("经纬度不能为空"))
		return
	}
	resp, err := h.svc.FindByLocation(req.Latitude, req.Longitude)
	if err != nil {
		failByError(ctx, CodeLBSRegionByLocation, err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// ===== M 端管理 =====

// AdminList 管理后台区域列表
// GET /api/v1/lbs/admin/regions  （需 lbs:manage 权限）
func (h *RegionHandler) AdminList(ctx plugin.Context) {
	var req dto.RegionListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		failByError(ctx, CodeLBSRegionListError, err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdminUpdateStatus 管理后台更新区域状态
// PUT /api/v1/lbs/admin/regions/:id/status  （需 lbs:manage 权限）
func (h *RegionHandler) AdminUpdateStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req struct {
		Status int `json:"status"`
	}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.AdminUpdateStatus(id, req.Status); err != nil {
		failByError(ctx, CodeLBSStatusInvalid, err)
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态已更新", nil))
}
