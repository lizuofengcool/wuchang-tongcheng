// Package handler 拍卖 HTTP 处理层
// 依据 v3.2.1 架构方案：状态机 pending → active → ended → sold / failed
package handler

import (
	"net/http"
	"strconv"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/ershou/dto"
	"wuchang-tongcheng/internal/modules/ershou/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// AuctionHandler 拍卖 HTTP 处理器
type AuctionHandler struct {
	service service.AuctionService
}

// NewAuctionHandler 创建拍卖 Handler 实例
func NewAuctionHandler(svc service.AuctionService) *AuctionHandler {
	return &AuctionHandler{service: svc}
}

// Create 创建拍卖（卖家为商品挂拍卖）
// POST /api/v1/ershou/:id/auction  （需登录 + 仅发布者本人）
func (h *AuctionHandler) Create(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	ershouID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.AuctionCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.service.Create(ershouID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("拍卖创建成功", resp))
}

// GetByErshouID 查询商品的拍卖信息
// GET /api/v1/ershou/:id/auction  （公开）
func (h *AuctionHandler) GetByErshouID(ctx plugin.Context) {
	ershouID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	resp, err := h.service.GetByErshouID(ershouID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// Bid 出价（含自动延拍逻辑）
// POST /api/v1/ershou/:id/auction/bid  （需登录 + 不能竞拍自己的商品）
func (h *AuctionHandler) Bid(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	ershouID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.AuctionBidRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	ip := ctx.GetHeader("X-Real-IP")
	if ip == "" {
		ip = ctx.GetHeader("X-Forwarded-For")
	}
	ua := ctx.GetHeader("User-Agent")
	resp, err := h.service.Bid(ershouID, userID, &req, ip, ua)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("出价成功", resp))
}

// EndManually 手动截拍（仅卖家）
// POST /api/v1/ershou/:id/auction/end  （需登录 + 仅发布者本人）
func (h *AuctionHandler) EndManually(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	ershouID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	resp, err := h.service.EndManually(ershouID, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("拍卖已截拍", resp))
}

// List 拍卖列表（按状态筛选）
// GET /api/v1/ershou/auctions  （公开）
func (h *AuctionHandler) List(ctx plugin.Context) {
	regionID := getRegionID(ctx)
	page, pageSize := parsePagination(ctx)
	statusStr := ctx.Query("status")
	var statusPtr *int
	if statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			statusPtr = &s
		}
	}
	pagination := utils.NewPagination(page, pageSize)
	pagination, list, err := h.service.List(regionID, statusPtr, pagination)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}
