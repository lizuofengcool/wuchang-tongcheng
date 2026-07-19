// Package handler 同城房屋租售 HTTP 处理层
// 依据 v3.2.1 架构方案第五章：对标贝壳/链家
// 依据需求文档 1.5：内容审核必须做（MVP 简化：发布即通过，M 端可手动审核/下架）
// 依据需求文档 1.10：4 维数据隔离（region_id 由 Region 中间件注入，user_id 由 JWT 解析）
package handler

import (
	"net/http"
	"strconv"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/house/dto"
	"wuchang-tongcheng/internal/modules/house/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ===== 辅助函数（其他 handler 共享） =====

// getUserProfile 从上下文获取登录用户的完整信息（id/name/phone/avatar）
// phone/avatar 来自 JWT 冗余字段，用于发布时冗余存储到 house 表，避免每次发布都查 users 表。
func getUserProfile(ctx plugin.Context) (userID uint, username, phone, avatar string) {
	if v, ok := ctx.Get(middleware.ContextUserID); ok {
		if id, ok := v.(uint); ok {
			userID = id
		}
	}
	if v, ok := ctx.Get(middleware.ContextUsername); ok {
		if name, ok := v.(string); ok {
			username = name
		}
	}
	if v, ok := ctx.Get(middleware.ContextUserPhone); ok {
		if p, ok := v.(string); ok {
			phone = p
		}
	}
	if v, ok := ctx.Get(middleware.ContextUserAvatar); ok {
		if a, ok := v.(string); ok {
			avatar = a
		}
	}
	return
}

// getRegionID 从上下文获取地区ID（由 Region 中间件注入）
func getRegionID(ctx plugin.Context) uint {
	if v, ok := ctx.Get(middleware.RegionIDKey); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return middleware.DefaultRegionID
}

// parseID 解析 URL 中的 :id 参数
func parseID(ctx plugin.Context) (uint, error) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// parseIDStr 解析字符串形式的 ID（如 query 参数）
func parseIDStr(s string) (uint, error) {
	id, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// parseSubID 解析 URL 中指定的路径参数（如 :sku_id / :agent_id 等）
func parseSubID(ctx plugin.Context, key string) (uint, error) {
	idStr := ctx.Param(key)
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// parsePagination 从 query 解析分页参数
func parsePagination(ctx plugin.Context) (page, pageSize int) {
	pageStr := ctx.Query("page")
	if pageStr == "" {
		pageStr = "1"
	}
	pageSizeStr := ctx.Query("page_size")
	if pageSizeStr == "" {
		pageSizeStr = "10"
	}
	page, _ = strconv.Atoi(pageStr)
	pageSize, _ = strconv.Atoi(pageSizeStr)
	return
}

// ===== HouseHandler =====

// HouseHandler 房源主表 HTTP 处理器
type HouseHandler struct {
	service service.HouseService
}

// NewHouseHandler 创建 HouseHandler 实例
func NewHouseHandler(svc service.HouseService) *HouseHandler {
	return &HouseHandler{service: svc}
}

// ===== C 端 =====

// Create 发布房源
// POST /api/v1/house  （需登录）
func (h *HouseHandler) Create(ctx plugin.Context) {
	userID, username, phone, avatar := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.CreateHouseRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.service.Create(regionID, userID, username, phone, avatar, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHousePublishError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("发布成功", info))
}

// Update 更新房源（仅发布者本人）
// PUT /api/v1/house/:id  （需登录）
func (h *HouseHandler) Update(ctx plugin.Context) {
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
	var req dto.UpdateHouseRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除房源（仅发布者本人）
// DELETE /api/v1/house/:id  （需登录）
func (h *HouseHandler) Delete(ctx plugin.Context) {
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
	if err := h.service.Delete(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 获取房源详情（同时增加浏览量）
// GET /api/v1/house/:id  （公开）
func (h *HouseHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	userID, _, _, _ := getUserProfile(ctx)
	info, err := h.service.GetByID(id, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 房源列表（C 端，仅返回已发布）
// GET /api/v1/house  （公开）
func (h *HouseHandler) List(ctx plugin.Context) {
	var req dto.HouseListRequest
	_ = ctx.Bind(&req)
	regionID := getRegionID(ctx)
	pagination, list, err := h.service.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListNearby 附近房源查询
// GET /api/v1/house/nearby  （公开）
func (h *HouseHandler) ListNearby(ctx plugin.Context) {
	var req dto.HouseNearbyRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if req.Latitude < -90 || req.Latitude > 90 || req.Longitude < -180 || req.Longitude > 180 {
		ctx.JSON(http.StatusOK, response.BadRequest("经纬度参数无效"))
		return
	}
	regionID := getRegionID(ctx)
	pagination, list, err := h.service.ListNearby(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Search 关键词搜索
// GET /api/v1/house/search  （公开）
func (h *HouseHandler) Search(ctx plugin.Context) {
	var req dto.HouseSearchRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if req.Keyword == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("关键词不能为空"))
		return
	}
	regionID := getRegionID(ctx)
	pagination, list, err := h.service.Search(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdvancedSearch 高级搜索
// GET /api/v1/house/advanced-search  （公开）
func (h *HouseHandler) AdvancedSearch(ctx plugin.Context) {
	var req dto.HouseAdvancedSearchRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	pagination, list, err := h.service.AdvancedSearch(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListMine 我的发布列表
// GET /api/v1/house/mine  （需登录）
func (h *HouseHandler) ListMine(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListMine(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ===== 收藏 =====

// Fav 收藏/取消收藏（toggle 语义）
// POST /api/v1/house/:id/fav  （需登录）
func (h *HouseHandler) Fav(ctx plugin.Context) {
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
	res, err := h.service.Fav(userID, id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	if res.HasFaved {
		ctx.JSON(http.StatusOK, response.SuccessWithMessage("收藏成功", res))
	} else {
		ctx.JSON(http.StatusOK, response.SuccessWithMessage("已取消收藏", res))
	}
}

// FavStatus 查询收藏状态
// GET /api/v1/house/:id/fav  （公开，未登录返回 has_faved=false）
func (h *HouseHandler) FavStatus(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	res, err := h.service.FavStatus(userID, id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(res))
}

// ListFavs 我的收藏列表
// GET /api/v1/house/favorites  （需登录）
func (h *HouseHandler) ListFavs(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListFavs(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ===== 互动 =====

// IncrContactCount 联系次数 +1（C 端点击"联系经纪人"时调用）
// POST /api/v1/house/:id/contact  （公开）
func (h *HouseHandler) IncrContactCount(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.service.IncrContactCount(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(nil))
}

// IncrShareCount 分享次数 +1
// POST /api/v1/house/:id/share  （公开）
func (h *HouseHandler) IncrShareCount(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.service.IncrShareCount(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(nil))
}

// ===== 相似推荐 =====

// ListSimilar 相似房源推荐
// GET /api/v1/house/:id/similar  （公开）
func (h *HouseHandler) ListSimilar(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	limitStr := ctx.DefaultQuery("limit", "5")
	limit, _ := strconv.Atoi(limitStr)
	list, err := h.service.ListSimilar(id, limit)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ===== 推广 =====

// UpdatePromotion 更新房源推广配置（M 端）
// PUT /api/v1/admin/house/:id/promotion  （需 house:audit 权限）
func (h *HouseHandler) UpdatePromotion(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.HousePromotionRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.UpdatePromotion(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("推广配置更新成功", nil))
}

// ===== M 端管理 =====

// AdminList 管理后台房源列表
// GET /api/v1/admin/house/list  （需 house:audit 权限）
func (h *HouseHandler) AdminList(ctx plugin.Context) {
	var req dto.HouseAdminListRequest
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdminGetByID 管理后台房源详情
// GET /api/v1/admin/house/:id  （需 house:audit 权限）
func (h *HouseHandler) AdminGetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.service.AdminGetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// Audit 审核房源（通过/拒绝）
// PUT /api/v1/admin/house/:id/audit  （需 house:audit 权限）
func (h *HouseHandler) Audit(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.AuditRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.Audit(id, req.AuditStatus, req.AuditReason); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseAuditError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核完成", nil))
}

// AdminUpdateStatus 强制下架/恢复
// PUT /api/v1/admin/house/:id/status  （需 house:audit 权限）
func (h *HouseHandler) AdminUpdateStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.AdminUpdateStatusRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.AdminUpdateStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseAuditError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态更新成功", nil))
}

// BatchAudit 批量审核
// POST /api/v1/admin/house/batch/audit  （需 house:audit 权限）
func (h *HouseHandler) BatchAudit(ctx plugin.Context) {
	var req dto.BatchAuditRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	res, err := h.service.BatchAudit(req.IDs, req.AuditStatus, req.AuditReason)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseAuditError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("批量审核完成", res))
}

// BatchUpdateStatus 批量状态变更
// POST /api/v1/admin/house/batch/status  （需 house:audit 权限）
func (h *HouseHandler) BatchUpdateStatus(ctx plugin.Context) {
	var req dto.BatchStatusUpdateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	res, err := h.service.BatchUpdateStatus(req.IDs, req.Status)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseAuditError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("批量状态更新完成", res))
}

// BatchDelete 批量删除
// POST /api/v1/admin/house/batch/delete  （需 house:audit 权限）
func (h *HouseHandler) BatchDelete(ctx plugin.Context) {
	var req dto.BatchDeleteRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	res, err := h.service.BatchDelete(req.IDs)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("批量删除完成", res))
}
