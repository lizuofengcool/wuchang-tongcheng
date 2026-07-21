// Package handler LBS地图中台 HTTP 处理层 - 地理围栏
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/lbs/dto"
	"wuchang-tongcheng/internal/modules/lbs/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// GeofenceHandler 地理围栏处理器
type GeofenceHandler struct {
	svc service.GeofenceService
}

// NewGeofenceHandler 创建 Geofence Handler 实例
func NewGeofenceHandler(svc service.GeofenceService) *GeofenceHandler {
	return &GeofenceHandler{svc: svc}
}

// Create 创建围栏
// POST /api/v1/lbs/geofences  （需登录）
func (h *GeofenceHandler) Create(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.CreateGeofenceRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.svc.Create(&req)
	if err != nil {
		failByError(ctx, CodeLBSGeofenceCreateError, err)
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// Update 更新围栏
// PUT /api/v1/lbs/geofences/:id  （需登录）
func (h *GeofenceHandler) Update(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateGeofenceRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Update(id, &req); err != nil {
		failByError(ctx, CodeLBSGeofenceCreateError, err)
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除围栏
// DELETE /api/v1/lbs/geofences/:id  （需登录）
func (h *GeofenceHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Delete(id); err != nil {
		failByError(ctx, CodeLBSGeofenceNotFound, err)
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 获取围栏详情
// GET /api/v1/lbs/geofences/:id  （公开）
func (h *GeofenceHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		failByError(ctx, CodeLBSGeofenceNotFound, err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 围栏列表
// GET /api/v1/lbs/geofences  （公开）
func (h *GeofenceHandler) List(ctx plugin.Context) {
	var req dto.GeofenceListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		failByError(ctx, CodeLBSGeofenceListError, err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByRegion 按区域列出
// GET /api/v1/lbs/geofences/by-region/:region_id  （公开）
func (h *GeofenceHandler) ListByRegion(ctx plugin.Context) {
	regionID, err := parseSubID(ctx, "region_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的区域ID"))
		return
	}
	list, err := h.svc.ListByRegion(regionID)
	if err != nil {
		failByError(ctx, CodeLBSGeofenceListError, err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ListByOwner 按所有者列出
// GET /api/v1/lbs/geofences/by-owner/:owner_id?owner_type=  （公开）
func (h *GeofenceHandler) ListByOwner(ctx plugin.Context) {
	ownerID, err := parseSubID(ctx, "owner_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的所有者ID"))
		return
	}
	ownerType := ctx.Query("owner_type")
	list, err := h.svc.ListByOwner(ownerID, ownerType)
	if err != nil {
		failByError(ctx, CodeLBSGeofenceListError, err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// CheckPoint 检查点是否在指定围栏内
// POST /api/v1/lbs/geofences/:id/check-point  （公开）
func (h *GeofenceHandler) CheckPoint(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.CheckPointRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.svc.CheckPoint(id, req.Latitude, req.Longitude)
	if err != nil {
		failByError(ctx, CodeLBSCheckPointError, err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// CheckPointInRegion 检查点是否在区域内任一围栏内
// POST /api/v1/lbs/geofences/check-point-in-region?region_id=  （公开）
// 未传 region_id 时使用 Region 中间件注入的默认值
func (h *GeofenceHandler) CheckPointInRegion(ctx plugin.Context) {
	var req dto.CheckPointRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if req.Latitude == 0 || req.Longitude == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("经纬度不能为空"))
		return
	}
	// 优先使用 query 参数 region_id，未传则取 Region 中间件注入的默认值
	regionID := uint(parseIntQuery(ctx, "region_id"))
	if regionID == 0 {
		regionID = getRegionID(ctx)
	}
	resp, err := h.svc.CheckPointInRegion(regionID, req.Latitude, req.Longitude)
	if err != nil {
		failByError(ctx, CodeLBSCheckPointError, err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// ===== M 端管理 =====

// AdminList 管理后台围栏列表
// GET /api/v1/lbs/admin/geofences  （需 lbs:manage 权限）
func (h *GeofenceHandler) AdminList(ctx plugin.Context) {
	var req dto.GeofenceListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		failByError(ctx, CodeLBSGeofenceListError, err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdminUpdateStatus 管理后台更新围栏状态
// PUT /api/v1/lbs/admin/geofences/:id/status  （需 lbs:manage 权限）
func (h *GeofenceHandler) AdminUpdateStatus(ctx plugin.Context) {
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
