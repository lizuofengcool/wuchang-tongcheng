// Package handler 同城114 HTTP 处理层 - 团购
// 依据 v3.2.1 架构方案：限时抢购/数量限制/使用规则/有效期
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/dh114/dto"
	"wuchang-tongcheng/internal/modules/dh114/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// GroupbuyHandler 团购 HTTP 处理器
type GroupbuyHandler struct {
	svc service.GroupbuyService
}

// NewGroupbuyHandler 创建团购 Handler 实例
func NewGroupbuyHandler(svc service.GroupbuyService) *GroupbuyHandler {
	return &GroupbuyHandler{svc: svc}
}

// Create 创建团购（C 端商户发布）
// POST /api/v1/dh114/groupbuys  （需登录）
func (h *GroupbuyHandler) Create(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.CreateGroupbuyRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.svc.Create(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114GroupbuyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("团购创建成功", info))
}

// Update 更新团购
// PUT /api/v1/dh114/groupbuys/:id  （需登录）
func (h *GroupbuyHandler) Update(ctx plugin.Context) {
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
	var req dto.UpdateGroupbuyRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114GroupbuyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除团购
// DELETE /api/v1/dh114/groupbuys/:id  （需登录）
func (h *GroupbuyHandler) Delete(ctx plugin.Context) {
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114GroupbuyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 团购详情
// GET /api/v1/dh114/groupbuys/:id  （公开）
func (h *GroupbuyHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	// 公开浏览 + 浏览计数
	_ = h.svc.IncrViewCount(id)
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114GroupbuyNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 团购列表
// GET /api/v1/dh114/groupbuys  （公开）
func (h *GroupbuyHandler) List(ctx plugin.Context) {
	var req dto.GroupbuyListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	regionID := getRegionID(ctx)
	pagination, list, err := h.svc.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114GroupbuyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByDh114 按商户列出团购
// GET /api/v1/dh114/dh114/:id/groupbuys  （公开）
func (h *GroupbuyHandler) ListByDh114(ctx plugin.Context) {
	dh114ID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的商户ID"))
		return
	}
	regionID := getRegionID(ctx)
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListByDh114(regionID, dh114ID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114GroupbuyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListHot 热门团购
// GET /api/v1/dh114/groupbuys/hot  （公开）
func (h *GroupbuyHandler) ListHot(ctx plugin.Context) {
	regionID := getRegionID(ctx)
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListHot(regionID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114GroupbuyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// IncrSold 增加团购销量
// POST /api/v1/dh114/groupbuys/:id/sold  （需登录）
func (h *GroupbuyHandler) IncrSold(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req struct {
		Count int `json:"count" binding:"min=1"`
	}
	if err := ctx.Bind(&req); err != nil {
		req.Count = 1
	}
	if err := h.svc.IncrSoldCount(id, req.Count); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114GroupbuyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("销量已更新", nil))
}

// AdminList 团购列表（M 端）
// GET /api/v1/dh114/admin/groupbuys  （需 dh114:audit 权限）
func (h *GroupbuyHandler) AdminList(ctx plugin.Context) {
	var req dto.GroupbuyAdminListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114GroupbuyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Audit 审核团购
// PUT /api/v1/dh114/admin/groupbuys/:id/audit  （需 dh114:audit 权限）
func (h *GroupbuyHandler) Audit(ctx plugin.Context) {
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114GroupbuyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核完成", nil))
}

// BatchAudit 批量审核团购
// POST /api/v1/dh114/admin/groupbuys/batch-audit  （需 dh114:audit 权限）
func (h *GroupbuyHandler) BatchAudit(ctx plugin.Context) {
	var req dto.BatchAuditRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	result, err := h.svc.BatchAudit(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114GroupbuyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(result))
}

// AdminUpdateStatus 更新团购状态（M 端）
// PUT /api/v1/dh114/admin/groupbuys/:id/status  （需 dh114:audit 权限）
func (h *GroupbuyHandler) AdminUpdateStatus(ctx plugin.Context) {
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeDh114GroupbuyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态已更新", nil))
}
