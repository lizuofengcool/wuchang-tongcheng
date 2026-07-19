// Package handler 同城车辆买卖 HTTP 处理层 - 车源主表
// 依据 v3.2.1 架构方案：对标瓜子/人人车/懂车帝/毛豆新车/易鑫车贷
// 依据需求文档 1.10：4 维数据隔离（region_id 由 Region 中间件注入，user_id 由 JWT 解析）
// 依据需求文档 1.5：内容审核必须做（MVP 简化：发布即通过，M 端可手动审核/下架）
package handler

import (
	"net/http"
	"strconv"
	"strings"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/car/dto"
	"wuchang-tongcheng/internal/modules/car/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ===== 辅助函数（car 模块通用，供所有 handler 文件复用） =====

// getUserProfile 从上下文获取登录用户的完整信息（id/name/phone/avatar）
// phone/avatar 来自 JWT 冗余字段（见 jwt.GenerateTokenWithProfile），
// 用于发布时冗余存储到 cars 表，避免每次发布都查 users 表。
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

// parseSubID 解析 URL 中的子资源 ID（如 :inspection_id / :test_drive_id）
func parseSubID(ctx plugin.Context, param string) (uint, error) {
	idStr := ctx.Param(param)
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

// getClientIP 从上下文获取客户端 IP（用于浏览记录）
func getClientIP(ctx plugin.Context) string {
	// 优先从 X-Forwarded-For 取
	if ip := ctx.GetHeader("X-Forwarded-For"); ip != "" {
		// 取第一个
		if idx := strings.Index(ip, ","); idx > 0 {
			return strings.TrimSpace(ip[:idx])
		}
		return strings.TrimSpace(ip)
	}
	if ip := ctx.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}
	// plugin.Context 接口未暴露 RemoteAddr，兜底返回空串
	return ""
}

// ===== CarHandler 车源主表 HTTP 处理器 =====

// CarHandler 车源主表 HTTP 处理器
type CarHandler struct {
	service service.CarService
}

// NewCarHandler 创建 CarHandler 实例
func NewCarHandler(svc service.CarService) *CarHandler {
	return &CarHandler{service: svc}
}

// ===== C 端 =====

// Create 发布车源
// POST /api/v1/car  （需登录 + car:publish）
func (h *CarHandler) Create(ctx plugin.Context) {
	userID, username, phone, avatar := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.CreateCarRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	regionID := getRegionID(ctx)
	info, err := h.service.Create(regionID, userID, username, phone, avatar, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarPublishError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("发布成功", info))
}

// Update 更新车源（仅发布者本人）
// PUT /api/v1/car/:id  （需登录 + car:publish）
func (h *CarHandler) Update(ctx plugin.Context) {
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

	var req dto.UpdateCarRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.service.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除车源（仅发布者本人）
// DELETE /api/v1/car/:id  （需登录 + car:publish）
func (h *CarHandler) Delete(ctx plugin.Context) {
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 获取车源详情（同时增加浏览量）
// GET /api/v1/car/:id  （公开）
func (h *CarHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	userID, _, _, _ := getUserProfile(ctx)
	info, err := h.service.GetByID(id, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 车源列表（C 端，地区隔离，默认仅返回已发布）
// GET /api/v1/car  （公开）
func (h *CarHandler) List(ctx plugin.Context) {
	var req dto.CarListRequest
	_ = ctx.Bind(&req)

	regionID := getRegionID(ctx)
	pagination, list, err := h.service.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Nearby 附近车源（基于 PostGIS 空间查询，扩展不可用降级 Haversine）
// GET /api/v1/car/nearby  （公开）
func (h *CarHandler) Nearby(ctx plugin.Context) {
	var req dto.CarNearbyRequest
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Search 关键词搜索（MVP 走 keyword LIKE，后续接 Elasticsearch）
// GET /api/v1/car/search  （公开）
func (h *CarHandler) Search(ctx plugin.Context) {
	var req dto.CarSearchRequest
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdvancedSearch 高级搜索（多条件组合 + 地理位置）
// GET /api/v1/car/advanced-search  （公开）
func (h *CarHandler) AdvancedSearch(ctx plugin.Context) {
	var req dto.AdvancedSearchRequest
	_ = ctx.Bind(&req)

	regionID := getRegionID(ctx)
	pagination, list, err := h.service.AdvancedSearch(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListMine 我的车源列表
// GET /api/v1/car/mine  （需登录）
func (h *CarHandler) ListMine(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListMine(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ===== 收藏 =====

// Fav 收藏/取消收藏（toggle 语义）
// POST /api/v1/car/:id/fav  （需登录）
func (h *CarHandler) Fav(ctx plugin.Context) {
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	if res.HasFaved {
		ctx.JSON(http.StatusOK, response.SuccessWithMessage("收藏成功", res))
	} else {
		ctx.JSON(http.StatusOK, response.SuccessWithMessage("已取消收藏", res))
	}
}

// FavStatus 查询收藏状态
// GET /api/v1/car/:id/fav  （公开，未登录返回 hasFaved=false）
func (h *CarHandler) FavStatus(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	res, err := h.service.FavStatus(userID, id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(res))
}

// ListFavs 我的收藏列表
// GET /api/v1/car/favorites  （需登录）
func (h *CarHandler) ListFavs(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListFavs(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ===== 互动 =====

// IncrContact 联系次数 +1（C 端点击"联系卖家"时调用）
// POST /api/v1/car/:id/contact  （公开）
func (h *CarHandler) IncrContact(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.service.IncrContact(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已记录联系", nil))
}

// IncrShare 分享次数 +1
// POST /api/v1/car/:id/share  （公开）
func (h *CarHandler) IncrShare(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.service.IncrShare(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已记录分享", nil))
}

// RecordView 记录浏览（用于行为分析）
// POST /api/v1/car/views  （公开，登录用户带 userID）
func (h *CarHandler) RecordView(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	ip := getClientIP(ctx)

	var req dto.CarViewRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.RecordView(userID, ip, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已记录浏览", nil))
}

// ===== M 端管理 =====

// AdminList 管理后台车源列表（可跨地区）
// GET /api/v1/car/admin/list  （需 car:audit 权限）
func (h *CarHandler) AdminList(ctx plugin.Context) {
	var req dto.CarAdminListRequest
	_ = ctx.Bind(&req)

	pagination, list, err := h.service.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdminGetByID 管理后台车源详情
// GET /api/v1/car/admin/:id  （需 car:audit 权限）
func (h *CarHandler) AdminGetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.service.AdminGetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// Audit 审核车源
// PUT /api/v1/car/admin/:id/audit  （需 car:audit 权限）
func (h *CarHandler) Audit(ctx plugin.Context) {
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarAuditError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核完成", nil))
}

// AdminUpdateStatus 强制下架/恢复
// PUT /api/v1/car/admin/:id/status  （需 car:audit 权限）
func (h *CarHandler) AdminUpdateStatus(ctx plugin.Context) {
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态更新成功", nil))
}

// RealCarVerify 真车认证
// PUT /api/v1/car/admin/:id/real-car-verify  （需 car:audit 权限）
func (h *CarHandler) RealCarVerify(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	var req dto.RealCarVerifyRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.service.RealCarVerify(id, req.Verified, req.Reason); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarAuditError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("认证状态更新成功", nil))
}

// UpdatePromotion 配置推广（精选/甄选/认证/权重等）
// PUT /api/v1/car/admin/:id/promotion  （需 car:manage 权限）
func (h *CarHandler) UpdatePromotion(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	var req dto.PromotionRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.service.UpdatePromotion(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("推广配置更新成功", nil))
}
