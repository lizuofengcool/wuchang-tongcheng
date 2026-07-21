// Package handler LBS地图中台 HTTP 处理层 - POI 兴趣点
package handler

import (
	"net/http"
	"strconv"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/lbs/dto"
	"wuchang-tongcheng/internal/modules/lbs/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// POIHandler POI 处理器
type POIHandler struct {
	svc service.POIService
}

// NewPOIHandler 创建 POI Handler 实例
func NewPOIHandler(svc service.POIService) *POIHandler {
	return &POIHandler{svc: svc}
}

// Create 创建 POI
// POST /api/v1/lbs/pois  （需登录）
func (h *POIHandler) Create(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.CreatePOIRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.svc.Create(regionID, userID, &req)
	if err != nil {
		failByError(ctx, CodeLBSPOICreateError, err)
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// Update 更新 POI
// PUT /api/v1/lbs/pois/:id  （需登录）
func (h *POIHandler) Update(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdatePOIRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Update(id, userID, &req); err != nil {
		failByError(ctx, CodeLBSPOIUpdateError, err)
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除 POI
// DELETE /api/v1/lbs/pois/:id  （需登录）
func (h *POIHandler) Delete(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Delete(id, userID); err != nil {
		failByError(ctx, CodeLBSPOIDeleteError, err)
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 获取 POI 详情
// GET /api/v1/lbs/pois/:id  （公开）
func (h *POIHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		failByError(ctx, CodeLBSPOINotFound, err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List POI 列表
// GET /api/v1/lbs/pois  （公开）
func (h *POIHandler) List(ctx plugin.Context) {
	var req dto.POIListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	regionID := getRegionID(ctx)
	pagination, list, err := h.svc.List(regionID, &req)
	if err != nil {
		failByError(ctx, CodeLBSPOIListError, err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListNearby 附近 POI 检索
// GET /api/v1/lbs/pois/nearby?latitude=&longitude=&radius_km=  （公开）
func (h *POIHandler) ListNearby(ctx plugin.Context) {
	var req dto.NearbyRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if req.Latitude == 0 || req.Longitude == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("经纬度不能为空"))
		return
	}
	if req.RadiusKm == 0 {
		req.RadiusKm = 5
	}
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	regionID := getRegionID(ctx)
	pagination, list, err := h.svc.ListNearby(regionID, &req)
	if err != nil {
		failByError(ctx, CodeLBSPOINearbyError, err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListMine 我的 POI 列表
// GET /api/v1/lbs/pois/mine  （需登录）
func (h *POIHandler) ListMine(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.POIListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.ListMine(userID, &req)
	if err != nil {
		failByError(ctx, CodeLBSPOIListError, err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ===== M 端管理 =====

// AdminList 管理后台 POI 列表
// GET /api/v1/lbs/admin/pois  （需 lbs:manage 权限）
func (h *POIHandler) AdminList(ctx plugin.Context) {
	var req dto.POIAdminListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.AdminList(&req)
	if err != nil {
		failByError(ctx, CodeLBSPOIListError, err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdminGetByID 管理后台获取 POI 详情
// GET /api/v1/lbs/admin/pois/:id  （需 lbs:manage 权限）
func (h *POIHandler) AdminGetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		failByError(ctx, CodeLBSPOINotFound, err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// AdminUpdateStatus 管理后台更新 POI 状态
// PUT /api/v1/lbs/admin/pois/:id/status  （需 lbs:manage 权限）
func (h *POIHandler) AdminUpdateStatus(ctx plugin.Context) {
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

// AdminDelete 管理后台删除 POI（强制下架）
// DELETE /api/v1/lbs/admin/pois/:id  （需 lbs:manage 权限）
func (h *POIHandler) AdminDelete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.AdminUpdateStatus(id, 4); err != nil {
		failByError(ctx, CodeLBSPOIDeleteError, err)
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已删除", nil))
}

// GetPOICategoriesHelper 提供给前端的 POI 分类聚合（占位）
// 实际可由 lbs_categories 表提供，MVP 阶段返回简单分类列表
func GetPOICategoriesHelper() []string {
	return []string{
		"restaurant", "landmark", "bus_station", "hotel", "hospital",
		"school", "shopping", "scenic", "gas_station", "atm",
		"parking", "other",
	}
}

// AdminListCategories 返回 POI 分类列表
// GET /api/v1/lbs/admin/pois/categories  （需 lbs:manage 权限）
func (h *POIHandler) AdminListCategories(ctx plugin.Context) {
	categories := GetPOICategoriesHelper()
	ctx.JSON(http.StatusOK, response.Success(categories))
}

// parseIntQuery 兼容方法：解析 query 中的整型参数
func parseIntQuery(ctx plugin.Context, key string) int {
	v := ctx.Query(key)
	if v == "" {
		return 0
	}
	n, _ := strconv.Atoi(v)
	return n
}
