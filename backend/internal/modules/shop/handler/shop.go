// Package handler 商家模块HTTP处理层
package handler

import (
	"net/http"
	"strconv"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/shop/dto"
	"wuchang-tongcheng/internal/modules/shop/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// Handler 商家HTTP处理器
type Handler struct {
	service service.ShopService
}

// NewHandler 创建商家处理器
func NewHandler(svc service.ShopService) *Handler {
	return &Handler{service: svc}
}

// getUserID 从上下文获取用户ID和用户名
func getUserID(ctx plugin.Context) (uint, string) {
	userID, _ := ctx.Get(middleware.ContextUserID)
	username, _ := ctx.Get(middleware.ContextUsername)
	id, _ := userID.(uint)
	name, _ := username.(string)
	return id, name
}

// getRegionID 从上下文获取地区ID
func getRegionID(ctx plugin.Context) uint {
	if v, ok := ctx.Get(middleware.RegionIDKey); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return middleware.DefaultRegionID
}

// parseID 解析URL中的ID参数
func parseID(ctx plugin.Context) (uint, error) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// ==================== 公开接口 ====================

// List 店铺列表
func (h *Handler) List(ctx plugin.Context) {
	var req dto.ShopListRequest
	_ = ctx.Bind(&req)

	regionID := getRegionID(ctx)
	pagination, list, err := h.service.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeShopError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// GetByID 店铺详情
func (h *Handler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的店铺ID"))
		return
	}
	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeShopNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetImages 店铺相册
func (h *Handler) GetImages(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的店铺ID"))
		return
	}
	list, err := h.service.GetImages(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeShopError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// GetReviews 店铺评价列表
func (h *Handler) GetReviews(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的店铺ID"))
		return
	}
	var req dto.ReviewListRequest
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.GetReviews(id, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeShopReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ==================== 用户接口 ====================

// Apply 商家入驻申请
func (h *Handler) Apply(ctx plugin.Context) {
	userID, _ := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.ApplyShopRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.service.Apply(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeShopApplyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("申请成功，请等待审核", info))
}

// GetMyShop 我的店铺
func (h *Handler) GetMyShop(ctx plugin.Context) {
	userID, _ := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.service.GetMyShop(userID, regionID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeShopNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// UpdateMyShop 编辑我的店铺
func (h *Handler) UpdateMyShop(ctx plugin.Context) {
	userID, _ := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.UpdateShopRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}
	regionID := getRegionID(ctx)
	if err := h.service.UpdateMyShop(userID, regionID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeShopError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// AddImage 上传店铺图片
func (h *Handler) AddImage(ctx plugin.Context) {
	userID, _ := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.AddShopImageRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.service.AddImage(userID, regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeShopError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("上传成功", info))
}

// DeleteImage 删除店铺图片
func (h *Handler) DeleteImage(ctx plugin.Context) {
	userID, _ := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	imageID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的图片ID"))
		return
	}
	if err := h.service.DeleteImage(userID, imageID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeShopError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// CreateReview 发表评价
func (h *Handler) CreateReview(ctx plugin.Context) {
	userID, _ := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	shopID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的店铺ID"))
		return
	}
	var req dto.CreateReviewRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.service.CreateReview(regionID, userID, shopID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeShopReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("评价成功，请等待审核", info))
}

// ==================== 管理接口 ====================

// AdminList 管理端店铺列表
func (h *Handler) AdminList(ctx plugin.Context) {
	var req dto.AdminShopListRequest
	_ = ctx.Bind(&req)
	regionID := getRegionID(ctx)
	pagination, list, err := h.service.AdminList(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeShopError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AuditShop 审核店铺
func (h *Handler) AuditShop(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的店铺ID"))
		return
	}
	var req dto.AuditShopRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}
	if err := h.service.AuditShop(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeShopAuditError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核成功", nil))
}

// UpdateShopStatus 修改营业状态
func (h *Handler) UpdateShopStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的店铺ID"))
		return
	}
	var req dto.UpdateShopStatusRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}
	if err := h.service.UpdateShopStatus(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeShopError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态更新成功", nil))
}

// SetRecommend 设置推荐
func (h *Handler) SetRecommend(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的店铺ID"))
		return
	}
	var req dto.SetRecommendRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}
	if err := h.service.SetRecommend(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeShopError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("设置成功", nil))
}

// DeleteShop 删除店铺
func (h *Handler) DeleteShop(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的店铺ID"))
		return
	}
	if err := h.service.DeleteShop(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeShopError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// AdminReviewList 评价管理列表
func (h *Handler) AdminReviewList(ctx plugin.Context) {
	var req dto.AdminReviewListRequest
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.AdminReviewList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeShopReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AuditReview 审核评价
func (h *Handler) AuditReview(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的评价ID"))
		return
	}
	var req dto.AuditReviewRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}
	if err := h.service.AuditReview(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeShopReviewError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核成功", nil))
}
