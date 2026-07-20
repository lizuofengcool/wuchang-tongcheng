// Package handler 同城零工兼职 HTTP 处理层 - 岗位主表
// 依据 v3.2.1 架构方案：对标斗米/青团兼职/兼职猫/猪八戒
// 依据需求文档 1.10：4 维数据隔离（region_id 由 Region 中间件注入，user_id 由 JWT 解析）
// 依据需求文档 1.5：内容审核必须做（MVP 简化：发布即通过，M 端可手动审核/下架）
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/linggong/dto"
	"wuchang-tongcheng/internal/modules/linggong/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// LinggongHandler 岗位主表 HTTP 处理器
type LinggongHandler struct {
	service service.LinggongService
}

// NewLinggongHandler 创建 LinggongHandler 实例
func NewLinggongHandler(svc service.LinggongService) *LinggongHandler {
	return &LinggongHandler{service: svc}
}

// ===== C 端 =====

// Create 发布岗位
// POST /api/v1/linggong  （需登录）
func (h *LinggongHandler) Create(ctx plugin.Context) {
	userID, username, phone, avatar := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.CreateLinggongRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	regionID := getRegionID(ctx)
	info, err := h.service.Create(regionID, userID, username, avatar, phone, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongPublishError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("发布成功", info))
}

// Update 更新岗位（仅发布者本人）
// PUT /api/v1/linggong/:id  （需登录）
func (h *LinggongHandler) Update(ctx plugin.Context) {
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

	var req dto.UpdateLinggongRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.service.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除岗位（仅发布者本人）
// DELETE /api/v1/linggong/:id  （需登录）
func (h *LinggongHandler) Delete(ctx plugin.Context) {
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 获取岗位详情（同时增加浏览量）
// GET /api/v1/linggong/:id  （公开）
func (h *LinggongHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 岗位列表（C 端，地区隔离）
// GET /api/v1/linggong  （公开）
func (h *LinggongHandler) List(ctx plugin.Context) {
	var req dto.LinggongListRequest
	_ = ctx.Bind(&req)

	regionID := getRegionID(ctx)
	pagination, list, err := h.service.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Search 关键词搜索
// GET /api/v1/linggong/search  （公开）
func (h *LinggongHandler) Search(ctx plugin.Context) {
	keyword := ctx.Query("keyword")
	if keyword == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("keyword 必填"))
		return
	}
	page, pageSize := parsePagination(ctx)
	regionID := getRegionID(ctx)
	pagination, list, err := h.service.Search(regionID, keyword, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Nearby 附近岗位
// GET /api/v1/linggong/nearby  （公开）
func (h *LinggongHandler) Nearby(ctx plugin.Context) {
	var req dto.LinggongNearbyRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if req.Latitude < -90 || req.Latitude > 90 || req.Longitude < -180 || req.Longitude > 180 {
		ctx.JSON(http.StatusOK, response.BadRequest("经纬度参数无效"))
		return
	}

	regionID := getRegionID(ctx)
	pagination, list, err := h.service.Nearby(regionID, req.Latitude, req.Longitude, req.RadiusKm, req.Page, req.PageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListMine 我的发布
// GET /api/v1/linggong/mine  （需登录）
func (h *LinggongHandler) ListMine(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListMine(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByEmployer 按雇主反查岗位
// GET /api/v1/linggong/employers/:id/linggongs  （公开）
func (h *LinggongHandler) ListByEmployer(ctx plugin.Context) {
	employerID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的雇主ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListByEmployer(employerID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// IncrContact 联系次数 +1
// POST /api/v1/linggong/:id/contact  （公开）
func (h *LinggongHandler) IncrContact(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.service.IncrContactCount(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已记录联系", nil))
}

// IncrShare 分享次数 +1
// POST /api/v1/linggong/:id/share  （公开）
func (h *LinggongHandler) IncrShare(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.service.IncrShareCount(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已记录分享", nil))
}

// IncrView 浏览量 +1
// POST /api/v1/linggong/:id/view  （公开）
func (h *LinggongHandler) IncrView(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.service.IncrViewCount(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已记录浏览", nil))
}

// ===== M 端管理 =====

// AdminList 管理后台岗位列表（可跨地区）
// GET /api/v1/linggong/admin/linggongs  （需 linggong:audit 权限）
func (h *LinggongHandler) AdminList(ctx plugin.Context) {
	var req dto.LinggongAdminListRequest
	_ = ctx.Bind(&req)

	pagination, list, err := h.service.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdminGetByID 管理后台岗位详情
// GET /api/v1/linggong/admin/linggongs/:id  （需 linggong:audit 权限）
func (h *LinggongHandler) AdminGetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.service.AdminGetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// Audit 审核岗位
// PUT /api/v1/linggong/admin/linggongs/:id/audit  （需 linggong:audit 权限）
func (h *LinggongHandler) Audit(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	var req dto.LinggongAuditRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.service.Audit(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongAuditError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核完成", nil))
}

// AdminUpdateStatus 强制下架/恢复
// PUT /api/v1/linggong/admin/linggongs/:id/status  （需 linggong:audit 权限）
func (h *LinggongHandler) AdminUpdateStatus(ctx plugin.Context) {
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

	if err := h.service.UpdateStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongStatusInvalid, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态更新成功", nil))
}

// BatchUpdateStatus 批量状态变更
// POST /api/v1/linggong/admin/linggongs/batch-status  （需 linggong:audit 权限）
func (h *LinggongHandler) BatchUpdateStatus(ctx plugin.Context) {
	var req dto.BatchStatusUpdateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.BatchUpdateStatus(req.IDs, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongStatusInvalid, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("批量状态更新完成", nil))
}
