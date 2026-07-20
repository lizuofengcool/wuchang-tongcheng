// Package handler love 相亲交友 HTTP 处理层 - 礼物
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/love/dto"
	"wuchang-tongcheng/internal/modules/love/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// GiftHandler 礼物 HTTP 处理器
type GiftHandler struct {
	service service.LoveGiftService
}

// NewGiftHandler 创建 GiftHandler 实例
func NewGiftHandler(svc service.LoveGiftService) *GiftHandler {
	return &GiftHandler{service: svc}
}

// ===== M 端 =====

// Create 创建礼物
// POST /api/v1/love/admin/gifts  （需 love:audit 权限）
func (h *GiftHandler) Create(ctx plugin.Context) {
	var req dto.CreateLoveGiftRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.service.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// Update 更新礼物
// PUT /api/v1/love/admin/gifts/:id  （需 love:audit 权限）
func (h *GiftHandler) Update(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateLoveGiftRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.Update(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除礼物
// DELETE /api/v1/love/admin/gifts/:id  （需 love:audit 权限）
func (h *GiftHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.service.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 礼物详情
// GET /api/v1/love/gifts/:id  （公开）
func (h *GiftHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 礼物列表
// GET /api/v1/love/gifts  （公开）
func (h *GiftHandler) List(ctx plugin.Context) {
	var req dto.LoveGiftListRequest
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// UpdateStatus 更新礼物状态
// PUT /api/v1/love/admin/gifts/:id/status  （需 love:audit 权限）
func (h *GiftHandler) UpdateStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req struct {
		Status int `json:"status" binding:"required,oneof=0 1"`
	}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.UpdateStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// BatchUpdateStatus 批量更新状态
// PUT /api/v1/love/admin/gifts/batch-status  （需 love:audit 权限）
func (h *GiftHandler) BatchUpdateStatus(ctx plugin.Context) {
	var req struct {
		IDs    []uint `json:"ids" binding:"required,min=1"`
		Status int    `json:"status" binding:"required,oneof=0 1"`
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

// ===== C 端 =====

// ListAvailable 当前会员等级可用的礼物
// GET /api/v1/love/gifts/available  （公开）
func (h *GiftHandler) ListAvailable(ctx plugin.Context) {
	memberLevelStr := ctx.Query("member_level")
	memberLevel, _ := parseUint(memberLevelStr)
	list, err := h.service.ListAvailable(int(memberLevel))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// Send 送礼
// POST /api/v1/love/gifts/send  （需登录）
func (h *GiftHandler) Send(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	var req dto.SendLoveGiftRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	loveIDStr := ctx.Query("love_id")
	loveID, _ := parseUint(loveIDStr)
	info, err := h.service.Send(userID, loveID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("送礼成功", info))
}
