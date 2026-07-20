// Package handler love 相亲交友 HTTP 处理层 - 拉黑
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/love/dto"
	"wuchang-tongcheng/internal/modules/love/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// BlockHandler 拉黑 HTTP 处理器
type BlockHandler struct {
	service service.LoveBlockService
}

// NewBlockHandler 创建 BlockHandler 实例
func NewBlockHandler(svc service.LoveBlockService) *BlockHandler {
	return &BlockHandler{service: svc}
}

// Block 拉黑用户
// POST /api/v1/love/blocks  （需登录）
func (h *BlockHandler) Block(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	var req dto.CreateLoveBlockRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	// loveID 由前端从 /love/me 获取后传入
	loveIDStr := ctx.Query("love_id")
	loveID, _ := parseUint(loveIDStr)
	info, err := h.service.Block(loveID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("拉黑成功", info))
}

// Unblock 取消拉黑
// DELETE /api/v1/love/blocks/:id  （需登录）
func (h *BlockHandler) Unblock(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.service.Unblock(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("取消拉黑成功", nil))
}

// GetByID 查询拉黑记录详情
// GET /api/v1/love/blocks/:id  （需登录）
func (h *BlockHandler) GetByID(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.service.GetByID(id, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 我的拉黑列表
// GET /api/v1/love/blocks  （需登录）
func (h *BlockHandler) List(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	var req dto.LoveBlockListRequest
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.List(&req, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// HasBlocked 查询是否已拉黑
// GET /api/v1/love/blocks/check  （需登录）
func (h *BlockHandler) HasBlocked(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if !ok {
		return
	}
	targetUserIDStr := ctx.Query("target_user_id")
	targetUserID, err := parseUint(targetUserIDStr)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("target_user_id 无效"))
		return
	}
	blocked, err := h.service.HasBlocked(userID, targetUserID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLoveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(map[string]bool{"blocked": blocked}))
}
