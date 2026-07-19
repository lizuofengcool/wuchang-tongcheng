// Package handler 同城招聘求职HTTP处理层 - 职位主表
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 依据需求文档 1.5：内容审核必须做（MVP 简化：发布即通过，M 端可手动审核/下架）
// 依据 v3.2.1 架构方案第四章：对标 BOSS直聘/拉勾/58招聘
package handler

import (
	"net/http"
	"strconv"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/job/dto"
	"wuchang-tongcheng/internal/modules/job/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ===== 辅助函数（包内共享，其他 handler 依赖） =====

// getUserProfile 从上下文获取登录用户的完整信息（id/name/phone/avatar）
// phone/avatar 来自 JWT 冗余字段，用于发布时冗余存储到 job 表，避免每次发布都查 users 表。
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

// parseSubID 解析 URL 中的指定 key 参数（如 :report_id :review_id :user_id 等）
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

// ===== JobHandler =====

// JobHandler 职位 HTTP 处理器
type JobHandler struct {
	service service.JobService
}

// NewJobHandler 创建 JobHandler 实例
func NewJobHandler(svc service.JobService) *JobHandler {
	return &JobHandler{service: svc}
}

// ===== C 端 =====

// Create 发布职位
// POST /api/v1/job  （需登录 + job:publish）
func (h *JobHandler) Create(ctx plugin.Context) {
	userID, username, phone, avatar := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.CreateJobRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	regionID := getRegionID(ctx)
	info, err := h.service.Create(regionID, userID, username, phone, avatar, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeJobPublishError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("发布成功", info))
}

// Update 更新职位（仅发布者本人）
// PUT /api/v1/job/:id  （需登录 + job:publish）
func (h *JobHandler) Update(ctx plugin.Context) {
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

	var req dto.UpdateJobRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.service.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeJobError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除职位（仅发布者本人）
// DELETE /api/v1/job/:id  （需登录 + job:publish）
func (h *JobHandler) Delete(ctx plugin.Context) {
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeJobError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 获取职位详情（同时增加浏览量，登录用户会标记收藏状态）
// GET /api/v1/job/:id  （公开）
func (h *JobHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	userID, _, _, _ := getUserProfile(ctx)
	info, err := h.service.GetByID(id, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeJobNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 职位列表（C 端，仅返回已发布）
// GET /api/v1/job  （公开）
func (h *JobHandler) List(ctx plugin.Context) {
	var req dto.JobListRequest
	_ = ctx.Bind(&req)

	regionID := getRegionID(ctx)
	pagination, list, err := h.service.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeJobError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListNearby 附近职位查询（基于 PostGIS 空间查询，扩展不可用降级 Haversine）
// GET /api/v1/job/nearby  （公开）
func (h *JobHandler) ListNearby(ctx plugin.Context) {
	var req dto.JobNearbyRequest
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeJobError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Search 全文检索（MVP 走 keyword LIKE，后续接 Elasticsearch）
// GET /api/v1/job/search  （公开）
func (h *JobHandler) Search(ctx plugin.Context) {
	var req dto.JobSearchRequest
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeJobError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdvancedSearch 高级搜索（多条件组合 + 技能/福利筛选 + 距离排序）
// GET /api/v1/job/advanced-search  （公开）
func (h *JobHandler) AdvancedSearch(ctx plugin.Context) {
	var req dto.AdvancedSearchRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	pagination, list, err := h.service.AdvancedSearch(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeJobError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListMine 我的发布列表
// GET /api/v1/job/mine  （需登录）
func (h *JobHandler) ListMine(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListMine(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeJobError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListSimilar 相似职位推荐
// GET /api/v1/job/:id/similar  （公开）
func (h *JobHandler) ListSimilar(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	limitStr := ctx.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)
	list, err := h.service.ListSimilar(id, limit)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeJobError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// UpdateStatus C 端上下架职位（仅发布者本人）
// PUT /api/v1/job/:id/status  （需登录 + job:publish）
func (h *JobHandler) UpdateStatus(ctx plugin.Context) {
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
	var req dto.UpdateJobStatusRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.UpdateStatus(id, userID, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeJobError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态更新成功", nil))
}

// ===== 收藏 =====

// Fav 收藏职位
// POST /api/v1/job/:id/fav  （需登录）
func (h *JobHandler) Fav(ctx plugin.Context) {
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeJobError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("收藏成功", res))
}

// Unfav 取消收藏职位
// DELETE /api/v1/job/:id/fav  （需登录）
func (h *JobHandler) Unfav(ctx plugin.Context) {
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
	res, err := h.service.Unfav(userID, id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeJobError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已取消收藏", res))
}

// FavStatus 查询收藏状态
// GET /api/v1/job/:id/fav  （公开，未登录返回 hasFaved=false）
func (h *JobHandler) FavStatus(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	res, err := h.service.FavStatus(userID, id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeJobNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(res))
}

// ListFavs 我的收藏列表（分页，按收藏时间倒序）
// GET /api/v1/job/favorites  （需登录）
func (h *JobHandler) ListFavs(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListFavs(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeJobError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ===== 推广 =====

// Promotion 职位推广（置顶/紧急/精选/认证/流量加权）
// POST /api/v1/job/:id/promotion  （需登录 + job:publish）
func (h *JobHandler) Promotion(ctx plugin.Context) {
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
	var req dto.JobPromotionRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.Promotion(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeJobError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("推广设置成功", nil))
}

// ===== M 端管理 =====

// AdminList 管理后台职位列表
// GET /api/v1/job/admin/list  （需 job:audit 权限）
func (h *JobHandler) AdminList(ctx plugin.Context) {
	var req dto.JobAdminListRequest
	_ = ctx.Bind(&req)

	pagination, list, err := h.service.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeJobError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdminGetByID 管理后台职位详情
// GET /api/v1/job/admin/:id  （需 job:audit 权限）
func (h *JobHandler) AdminGetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.service.AdminGetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeJobNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// Audit 审核职位（通过/拒绝）
// PUT /api/v1/job/admin/:id/audit  （需 job:audit 权限）
func (h *JobHandler) Audit(ctx plugin.Context) {
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeJobAuditError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核完成", nil))
}

// AdminUpdateStatus 强制下架/恢复
// PUT /api/v1/job/admin/:id/status  （需 job:audit 权限）
func (h *JobHandler) AdminUpdateStatus(ctx plugin.Context) {
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeJobAuditError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态更新成功", nil))
}
