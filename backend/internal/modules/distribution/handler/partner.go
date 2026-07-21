// Package handler 分销合伙人中台 HTTP 处理层 - 合伙人
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/distribution/dto"
	"wuchang-tongcheng/internal/modules/distribution/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// PartnerHandler 合伙人处理器
type PartnerHandler struct {
	svc service.PartnerService
}

// NewPartnerHandler 创建 PartnerHandler 实例
func NewPartnerHandler(svc service.PartnerService) *PartnerHandler {
	return &PartnerHandler{svc: svc}
}

// Apply 申请加入合伙人
// POST /api/v1/distribution/partners/apply  （需登录）
func (h *PartnerHandler) Apply(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, failUnauthorized())
		return
	}
	var req dto.PartnerApplyRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.svc.Apply(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("申请已提交", info))
}

// GetMine 我的合伙人信息
// GET /api/v1/distribution/partners/mine  （需登录）
func (h *PartnerHandler) GetMine(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, failUnauthorized())
		return
	}
	info, err := h.svc.GetByUserID(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionPartnerNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// Tree 上下级树
// GET /api/v1/distribution/partners/tree  （需登录）
func (h *PartnerHandler) Tree(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, failUnauthorized())
		return
	}
	// 默认以当前用户的合伙人 ID 为根
	me, err := h.svc.GetByUserID(userID)
	var parentID uint
	if err == nil && me != nil {
		parentID = me.ID
	}
	var req dto.PartnerTreeRequest
	_ = ctx.Bind(&req)
	if req.ParentID > 0 {
		parentID = req.ParentID
	}
	list, err := h.svc.Tree(&dto.PartnerTreeRequest{ParentID: parentID, Depth: req.Depth})
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// GetByID 详情（公开只读）
// GET /api/v1/distribution/partners/:id
func (h *PartnerHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, failParam("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionPartnerNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 列表（公开只读）
// GET /api/v1/distribution/partners
func (h *PartnerHandler) List(ctx plugin.Context) {
	var req dto.PartnerListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ===== 管理后台 =====

// AdminList 管理后台列表
// GET /api/v1/distribution/admin/partners
func (h *PartnerHandler) AdminList(ctx plugin.Context) {
	var req dto.PartnerListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdminGetByID 管理后台详情
// GET /api/v1/distribution/admin/partners/:id
func (h *PartnerHandler) AdminGetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, failParam("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionPartnerNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// AdminUpdate 管理后台更新（等级/佣金率/上级/状态）
// PUT /api/v1/distribution/admin/partners/:id
func (h *PartnerHandler) AdminUpdate(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, failParam("无效的ID"))
		return
	}
	var req dto.PartnerUpdateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Update(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// AdminUpdateStatus 管理后台更新状态（审核通过/冻结等）
// PUT /api/v1/distribution/admin/partners/:id/status
func (h *PartnerHandler) AdminUpdateStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, failParam("无效的ID"))
		return
	}
	var body struct {
		Status int `json:"status" binding:"required,oneof=0 1 2 3 4"`
	}
	if err := ctx.Bind(&body); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.UpdateStatus(id, body.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionPartnerStatusInvalid, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态已更新", nil))
}

// AdminUpgrade 管理后台升级等级
// PUT /api/v1/distribution/admin/partners/:id/upgrade
func (h *PartnerHandler) AdminUpgrade(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, failParam("无效的ID"))
		return
	}
	var body struct {
		Level int `json:"level" binding:"required,oneof=1 2 3"`
	}
	if err := ctx.Bind(&body); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Upgrade(id, body.Level); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionPartnerLevelInvalid, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("等级已调整", nil))
}

// AdminAdjustRate 管理后台调整佣金率
// PUT /api/v1/distribution/admin/partners/:id/commission-rate
func (h *PartnerHandler) AdminAdjustRate(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, failParam("无效的ID"))
		return
	}
	var body struct {
		Rate float64 `json:"rate" binding:"required,min=0,max=1"`
	}
	if err := ctx.Bind(&body); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.AdjustCommissionRate(id, body.Rate); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionPartnerRateInvalid, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("佣金率已调整", nil))
}
