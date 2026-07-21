// Package handler 分销合伙人中台 HTTP 处理层 - 推广渠道
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/distribution/dto"
	"wuchang-tongcheng/internal/modules/distribution/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ChannelHandler 渠道处理器
type ChannelHandler struct {
	svc service.ChannelService
}

// NewChannelHandler 创建 ChannelHandler 实例
func NewChannelHandler(svc service.ChannelService) *ChannelHandler {
	return &ChannelHandler{svc: svc}
}

// Create 创建渠道（C 端，需登录，需为合伙人）
// POST /api/v1/distribution/channels
func (h *ChannelHandler) Create(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, failUnauthorized())
		return
	}
	var req dto.ChannelCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	// 通过 userID 解析 partnerID（简化：直接查合伙人）
	// 此处依赖前端在 body 中传 partner_id，若未传则需 partner 服务按 user_id 查询
	// 为保持 service 接口简洁，handler 层从 body 扩展字段
	type createReq struct {
		dto.ChannelCreateRequest
		PartnerID uint `json:"partner_id"`
	}
	var cr createReq
	_ = ctx.Bind(&cr)
	if cr.PartnerID == 0 {
		ctx.JSON(http.StatusOK, failParam("partner_id 不能为空"))
		return
	}
	info, err := h.svc.Create(cr.PartnerID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("渠道已创建", info))
}

// Update 更新渠道
// PUT /api/v1/distribution/channels/:id
func (h *ChannelHandler) Update(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, failUnauthorized())
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, failParam("无效的ID"))
		return
	}
	var req dto.ChannelUpdateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	// operatorPartnerID=0 表示管理员操作，跳过权限校验
	if err := h.svc.Update(id, 0, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除渠道
// DELETE /api/v1/distribution/channels/:id
func (h *ChannelHandler) Delete(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, failUnauthorized())
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, failParam("无效的ID"))
		return
	}
	if err := h.svc.Delete(id, 0); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 详情
// GET /api/v1/distribution/channels/:id
func (h *ChannelHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, failParam("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionChannelNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 列表
// GET /api/v1/distribution/channels
func (h *ChannelHandler) List(ctx plugin.Context) {
	var req dto.ChannelListRequest
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

// ListMine 我的渠道（需登录）
// GET /api/v1/distribution/channels/mine
func (h *ChannelHandler) ListMine(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, failUnauthorized())
		return
	}
	partnerID := parseUintPtr(ctx, "partner_id")
	if partnerID == nil || *partnerID == 0 {
		ctx.JSON(http.StatusOK, failParam("partner_id 不能为空"))
		return
	}
	list, err := h.svc.ListByPartner(*partnerID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// Stats 渠道统计
// GET /api/v1/distribution/channels/stats
func (h *ChannelHandler) Stats(ctx plugin.Context) {
	var req dto.ChannelStatsRequest
	_ = ctx.Bind(&req)
	if req.PartnerID == 0 {
		ctx.JSON(http.StatusOK, failParam("partner_id 不能为空"))
		return
	}
	info, err := h.svc.Stats(req.PartnerID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// Track 渠道追踪（公开，记录点击/注册/下单）
// POST /api/v1/distribution/channels/track
func (h *ChannelHandler) Track(ctx plugin.Context) {
	var req dto.ChannelTrackRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, failParam("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Track(&req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(CodeDistributionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已记录", nil))
}

// ===== 管理后台 =====

// AdminList 管理后台列表
// GET /api/v1/distribution/admin/channels
func (h *ChannelHandler) AdminList(ctx plugin.Context) {
	var req dto.ChannelListRequest
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
