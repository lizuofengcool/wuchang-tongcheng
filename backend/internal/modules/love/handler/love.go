// Package handler love 相亲交友 HTTP 处理层 - 主表 Love
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 依据需求文档 1.5：内容审核必须做（MVP 简化：发布即通过，M 端可手动审核/下架）
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/love/dto"
	"wuchang-tongcheng/internal/modules/love/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// LoveHandler 主表 HTTP 处理器
type LoveHandler struct {
	service service.LoveService
}

// NewLoveHandler 创建 LoveHandler 实例
func NewLoveHandler(svc service.LoveService) *LoveHandler {
	return &LoveHandler{service: svc}
}

// ===== C 端 =====

// Create 创建用户资料
// POST /api/v1/love  （需登录）
func (h *LoveHandler) Create(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	var req dto.CreateLoveRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.service.Create(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// Update 更新用户资料
// PUT /api/v1/love/:id  （需登录）
func (h *LoveHandler) Update(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateLoveRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// GetByID 获取用户资料详情（同时增加浏览量）
// GET /api/v1/love/:id  （公开）
func (h *LoveHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	userID, _, _ := getUserProfile(ctx)
	info, err := h.service.GetByID(id, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetByUserID 获取当前登录用户资料
// GET /api/v1/love/me  （需登录）
func (h *LoveHandler) GetByUserID(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	info, err := h.service.GetByUserID(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 用户资料列表
// GET /api/v1/love  （公开）
func (h *LoveHandler) List(ctx plugin.Context) {
	var req dto.LoveListRequest
	_ = ctx.Bind(&req)
	regionID := getRegionID(ctx)
	pagination, list, err := h.service.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Nearby 附近的人（基于 PostGIS 空间查询）
// GET /api/v1/love/nearby  （公开）
func (h *LoveHandler) Nearby(ctx plugin.Context) {
	var req dto.LoveNearbyRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	pagination, list, err := h.service.ListNearby(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Search 关键词搜索
// GET /api/v1/love/search  （公开）
func (h *LoveHandler) Search(ctx plugin.Context) {
	var req dto.LoveAdvancedSearchRequest
	_ = ctx.Bind(&req)
	regionID := getRegionID(ctx)
	pagination, list, err := h.service.Search(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdvancedSearch 高级搜索（多条件组合）
// GET /api/v1/love/advanced-search  （公开）
func (h *LoveHandler) AdvancedSearch(ctx plugin.Context) {
	var req dto.LoveAdvancedSearchRequest
	_ = ctx.Bind(&req)
	regionID := getRegionID(ctx)
	pagination, list, err := h.service.AdvancedSearch(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// UpdateLocation 更新位置
// PUT /api/v1/love/:id/location  （需登录）
func (h *LoveHandler) UpdateLocation(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateLocationRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.UpdateLocation(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// UpdateVoiceIntro 更新语音介绍
// PUT /api/v1/love/:id/voice-intro  （需登录）
func (h *LoveHandler) UpdateVoiceIntro(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateVoiceIntroRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.UpdateVoiceIntro(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// MatchScore 灵魂匹配评分
// GET /api/v1/love/match-score  （需登录）
func (h *LoveHandler) MatchScore(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	targetUserIDStr := ctx.Query("target_user_id")
	if targetUserIDStr == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("target_user_id 不能为空"))
		return
	}
	targetUserID, err := parseUint(targetUserIDStr)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("target_user_id 无效"))
		return
	}
	resp, err := h.service.MatchScore(userID, targetUserID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// ===== M 端 =====

// AdminList 管理后台：用户资料列表
// GET /api/v1/love/admin  （需 love:audit 权限）
func (h *LoveHandler) AdminList(ctx plugin.Context) {
	var req dto.LoveListRequest
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdminGetByID 管理后台：获取用户资料详情
// GET /api/v1/love/admin/:id  （需 love:audit 权限）
func (h *LoveHandler) AdminGetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.service.AdminGetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// Audit 管理后台：审核用户资料
// PUT /api/v1/love/admin/:id/audit  （需 love:audit 权限）
func (h *LoveHandler) Audit(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req struct {
		AuditStatus int    `json:"audit_status" binding:"required,oneof=1 2 3"`
		AuditReason string `json:"audit_reason"`
	}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.Audit(id, req.AuditStatus, req.AuditReason); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核成功", nil))
}

// AdminUpdateStatus 管理后台：更新资料状态
// PUT /api/v1/love/admin/:id/status  （需 love:audit 权限）
func (h *LoveHandler) AdminUpdateStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req struct {
		Status int `json:"status" binding:"required"`
	}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.AdminUpdateStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// SetFeatured 设置精选
// PUT /api/v1/love/admin/:id/featured  （需 love:audit 权限）
func (h *LoveHandler) SetFeatured(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req struct {
		Featured bool `json:"featured"`
	}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.SetFeatured(id, req.Featured); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("设置成功", nil))
}

// SetPicked 设置推荐
// PUT /api/v1/love/admin/:id/picked  （需 love:audit 权限）
func (h *LoveHandler) SetPicked(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req struct {
		Picked bool `json:"picked"`
	}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.SetPicked(id, req.Picked); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("设置成功", nil))
}

// BatchAudit 批量审核
// PUT /api/v1/love/admin/batch-audit  （需 love:audit 权限）
func (h *LoveHandler) BatchAudit(ctx plugin.Context) {
	var req struct {
		IDs         []uint `json:"ids" binding:"required,min=1"`
		AuditStatus int    `json:"audit_status" binding:"required,oneof=1 2 3"`
		AuditReason string `json:"audit_reason"`
	}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.BatchAudit(req.IDs, req.AuditStatus, req.AuditReason); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("批量审核成功", nil))
}

// BatchUpdateStatus 批量更新状态
// PUT /api/v1/love/admin/batch-status  （需 love:audit 权限）
func (h *LoveHandler) BatchUpdateStatus(ctx plugin.Context) {
	var req struct {
		IDs    []uint `json:"ids" binding:"required,min=1"`
		Status int    `json:"status" binding:"required"`
	}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.BatchUpdateStatus(req.IDs, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("批量更新成功", nil))
}

