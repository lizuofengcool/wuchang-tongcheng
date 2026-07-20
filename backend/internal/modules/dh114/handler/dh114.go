// Package handler 同城114 HTTP 处理层 - 商户主表
// 依据 v3.2.1 架构方案：对标大众点评/美团/58同城
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 依据需求文档 1.5：内容审核必须做（MVP 简化：发布即通过，M 端可手动审核/下架）
package handler

import (
	"net/http"
	"strconv"
	"strings"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/dh114/dto"
	"wuchang-tongcheng/internal/modules/dh114/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// Handler 商户主表 HTTP 处理器
type Handler struct {
	svc service.Dh114Service
}

// NewHandler 创建 Handler 实例
func NewHandler(svc service.Dh114Service) *Handler {
	return &Handler{svc: svc}
}

// getUserProfile 从上下文获取登录用户信息（id/name/phone/avatar）
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

// getRegionID 从上下文获取地区 ID（由 Region 中间件注入）
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

// parseSubID 解析 URL 中的指定参数（如 :group_id/:menu_id 等）
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
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")
	page, _ = strconv.Atoi(pageStr)
	pageSize, _ = strconv.Atoi(pageSizeStr)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return
}

// parseBoolPtr 解析 *bool 类型 query 参数
func parseBoolPtr(ctx plugin.Context, key string) *bool {
	v := ctx.Query(key)
	if v == "" {
		return nil
	}
	b := strings.ToLower(v) == "true" || v == "1"
	return &b
}

// parseIntPtr 解析 *int 类型 query 参数
func parseIntPtr(ctx plugin.Context, key string) *int {
	v := ctx.Query(key)
	if v == "" {
		return nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return nil
	}
	return &n
}

// getClientIP 获取客户端 IP
func getClientIP(ctx plugin.Context) string {
	// 优先取 X-Forwarded-For
	if ip := ctx.GetHeader("X-Forwarded-For"); ip != "" {
		if idx := strings.Index(ip, ","); idx > 0 {
			return strings.TrimSpace(ip[:idx])
		}
		return ip
	}
	if ip := ctx.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}
	// plugin.Context 接口未暴露 RemoteAddr，兜底返回空串
	return ""
}

// ===== C 端 =====

// Create 发布商户
// POST /api/v1/dh114  （需登录）
func (h *Handler) Create(ctx plugin.Context) {
	userID, username, phone, avatar := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.CreateDh114Request
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.svc.Create(regionID, userID, username, phone, avatar, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114PublishError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("发布成功", info))
}

// Update 更新商户
// PUT /api/v1/dh114/:id  （需登录）
func (h *Handler) Update(ctx plugin.Context) {
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
	var req dto.UpdateDh114Request
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114Error, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除商户
// DELETE /api/v1/dh114/:id  （需登录）
func (h *Handler) Delete(ctx plugin.Context) {
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114Error, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 获取商户详情
// GET /api/v1/dh114/:id  （公开）
func (h *Handler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	userID, _, _, _ := getUserProfile(ctx)
	info, err := h.svc.GetByID(id, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114NotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 商户列表
// GET /api/v1/dh114  （公开）
func (h *Handler) List(ctx plugin.Context) {
	var req dto.Dh114ListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	regionID := getRegionID(ctx)
	pagination, list, err := h.svc.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114Error, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListNearby 附近商户
// GET /api/v1/dh114/nearby  （公开）
func (h *Handler) ListNearby(ctx plugin.Context) {
	var req dto.Dh114NearbyRequest
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114Error, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Search 搜索商户
// GET /api/v1/dh114/search  （公开）
func (h *Handler) Search(ctx plugin.Context) {
	var req dto.Dh114SearchRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	regionID := getRegionID(ctx)
	pagination, list, err := h.svc.Search(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114Error, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdvancedSearch 高级搜索
// GET /api/v1/dh114/advanced-search  （公开）
func (h *Handler) AdvancedSearch(ctx plugin.Context) {
	var req dto.AdvancedSearchRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	regionID := getRegionID(ctx)
	pagination, list, err := h.svc.AdvancedSearch(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114Error, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListMine 我的商户
// GET /api/v1/dh114/mine  （需登录）
func (h *Handler) ListMine(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListMine(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114Error, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ===== 收藏 =====

// Fav 收藏商户
// POST /api/v1/dh114/:id/fav  （需登录）
func (h *Handler) Fav(ctx plugin.Context) {
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
	resp, err := h.svc.Fav(userID, id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114FavoriteError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("收藏成功", resp))
}

// Unfav 取消收藏
// DELETE /api/v1/dh114/:id/fav  （需登录）
func (h *Handler) Unfav(ctx plugin.Context) {
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
	resp, err := h.svc.Unfav(userID, id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114FavoriteError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已取消收藏", resp))
}

// FavStatus 收藏状态
// GET /api/v1/dh114/:id/fav  （公开，登录后返回真实状态）
func (h *Handler) FavStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	userID, _, _, _ := getUserProfile(ctx)
	resp, err := h.svc.FavStatus(userID, id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114NotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// ListFavs 我的收藏列表
// GET /api/v1/dh114/favorites  （需登录）
func (h *Handler) ListFavs(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.FavoriteListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.ListFavs(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114Error, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ===== 互动 =====

// IncrShare 增加分享数
// POST /api/v1/dh114/:id/share  （公开）
func (h *Handler) IncrShare(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.IncrShare(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114Error, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已记录分享", nil))
}

// IncrContact 增加联系数
// POST /api/v1/dh114/:id/contact  （公开）
func (h *Handler) IncrContact(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.IncrContact(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114Error, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已记录联系", nil))
}

// RecordCall 记录电话拨打（一键拨号核心）
// POST /api/v1/dh114/:id/call  （公开，登录后记录主叫）
func (h *Handler) RecordCall(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.PhoneCallRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	req.Dh114ID = id
	userID, _, _, _ := getUserProfile(ctx)
	ip := getClientIP(ctx)
	userAgent := ctx.GetHeader("User-Agent")
	resp, err := h.svc.RecordCall(userID, id, "", &req, ip, userAgent)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114PhoneCallError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// RecordView 记录浏览
// POST /api/v1/dh114/:id/view  （公开）
func (h *Handler) RecordView(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	userID, _, _, _ := getUserProfile(ctx)
	ip := getClientIP(ctx)
	req := &dto.Dh114ViewRequest{Dh114ID: id, VisitType: "business"}
	_ = ctx.Bind(req)
	req.Dh114ID = id
	if err := h.svc.RecordView(userID, ip, req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114Error, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已记录浏览", nil))
}

// ===== M 端管理 =====

// AdminList 管理后台商户列表
// GET /api/v1/dh114/admin/list  （需 content:audit 权限）
func (h *Handler) AdminList(ctx plugin.Context) {
	var req dto.Dh114AdminListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114Error, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdminGetByID 管理后台获取详情
// GET /api/v1/dh114/admin/:id  （需 content:audit 权限）
func (h *Handler) AdminGetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.AdminGetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114NotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// Audit 审核商户
// PUT /api/v1/dh114/admin/:id/audit  （需 content:audit 权限）
func (h *Handler) Audit(ctx plugin.Context) {
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
	if err := h.svc.Audit(id, req.AuditStatus, req.AuditReason); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114AuditError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核完成", nil))
}

// BatchAudit 批量审核
// POST /api/v1/dh114/admin/batch-audit  （需 content:audit 权限）
func (h *Handler) BatchAudit(ctx plugin.Context) {
	var req dto.BatchAuditRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	result, err := h.svc.BatchAudit(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114AuditError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(result))
}

// AdminUpdateStatus 管理后台更新状态（强制下架/恢复）
// PUT /api/v1/dh114/admin/:id/status  （需 content:audit 权限）
func (h *Handler) AdminUpdateStatus(ctx plugin.Context) {
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
	if err := h.svc.AdminUpdateStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114StatusInvalid, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态已更新", nil))
}

// UpdatePromotion 更新推广配置
// PUT /api/v1/dh114/admin/:id/promotion  （需 content:audit 权限）
func (h *Handler) UpdatePromotion(ctx plugin.Context) {
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
	if err := h.svc.UpdatePromotion(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114Error, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("推广配置已更新", nil))
}
