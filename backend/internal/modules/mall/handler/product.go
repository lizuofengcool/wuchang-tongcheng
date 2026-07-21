// Package handler 同城商城 HTTP 处理层 - 商品 SPU
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/mall/dto"
	"wuchang-tongcheng/internal/modules/mall/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ProductHandler 商品 HTTP 处理器
type ProductHandler struct {
	svc service.ProductService
}

// NewProductHandler 创建商品 Handler 实例
func NewProductHandler(svc service.ProductService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

// Create 创建商品
// POST /api/v1/mall/products  （需登录，卖家操作）
func (h *ProductHandler) Create(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	var req dto.CreateProductRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	shopID := uint(parseQueryInt(ctx, "shop_id", 0))
	if shopID == 0 {
		shopID = req.ShopID
	}
	info, err := h.svc.Create(regionID, userID, shopID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallProductError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// Update 更新商品
// PUT /api/v1/mall/products/:id  （需登录）
func (h *ProductHandler) Update(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateProductRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallProductError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除商品
// DELETE /api/v1/mall/products/:id  （需登录）
func (h *ProductHandler) Delete(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.Delete(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallProductError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 商品详情
// GET /api/v1/mall/products/:id  （公开）
func (h *ProductHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallProductNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 商品列表
// GET /api/v1/mall/products  （公开）
func (h *ProductHandler) List(ctx plugin.Context) {
	var req dto.ProductListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	regionID := getRegionID(ctx)
	pagination, list, err := h.svc.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallProductError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdminList 管理后台商品列表
// GET /api/v1/mall/admin/products  （需 mall:audit 权限）
func (h *ProductHandler) AdminList(ctx plugin.Context) {
	var req dto.ProductAdminListRequest
	_ = ctx.Bind(&req)
	if req.Page == 0 {
		req.Page, req.PageSize = parsePagination(ctx)
	}
	if req.RegionID == 0 {
		req.RegionID = getRegionID(ctx)
	}
	pagination, list, err := h.svc.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallProductError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Search 搜索商品
// GET /api/v1/mall/products/search  （公开）
func (h *ProductHandler) Search(ctx plugin.Context) {
	keyword := ctx.Query("keyword")
	page, pageSize := parsePagination(ctx)
	regionID := getRegionID(ctx)
	pagination, list, err := h.svc.Search(regionID, keyword, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallProductError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByShop 按店铺列出商品
// GET /api/v1/mall/products/by-shop/:shop_id  （公开）
func (h *ProductHandler) ListByShop(ctx plugin.Context) {
	shopID, err := parseSubID(ctx, "shop_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的店铺ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListByShop(shopID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallProductError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByCategory 按分类列出商品
// GET /api/v1/mall/products/by-category/:category_id  （公开）
func (h *ProductHandler) ListByCategory(ctx plugin.Context) {
	categoryID, err := parseSubID(ctx, "category_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的分类ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	regionID := getRegionID(ctx)
	pagination, list, err := h.svc.ListByCategory(regionID, categoryID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallProductError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByUser 按用户列出商品
// GET /api/v1/mall/products/mine  （需登录）
func (h *ProductHandler) ListByUser(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListByUser(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallProductError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListFeatured 精选商品
// GET /api/v1/mall/products/featured  （公开）
func (h *ProductHandler) ListFeatured(ctx plugin.Context) {
	regionID := getRegionID(ctx)
	limit := parseQueryInt(ctx, "limit", 10)
	list, err := h.svc.ListFeatured(regionID, limit)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallProductError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ListHot 热销商品
// GET /api/v1/mall/products/hot  （公开）
func (h *ProductHandler) ListHot(ctx plugin.Context) {
	regionID := getRegionID(ctx)
	limit := parseQueryInt(ctx, "limit", 10)
	list, err := h.svc.ListHot(regionID, limit)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallProductError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ListNew 新品
// GET /api/v1/mall/products/new  （公开）
func (h *ProductHandler) ListNew(ctx plugin.Context) {
	regionID := getRegionID(ctx)
	limit := parseQueryInt(ctx, "limit", 10)
	list, err := h.svc.ListNew(regionID, limit)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallProductError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// UpdateStatus 更新商品状态
// PUT /api/v1/mall/products/:id/status  （需登录）
func (h *ProductHandler) UpdateStatus(ctx plugin.Context) {
	userID, ok := requireLogin(ctx)
	if ok {
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateStatusRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.UpdateStatus(id, userID, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallProductError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态已更新", nil))
}

// Audit 审核商品
// PUT /api/v1/mall/admin/products/:id/audit  （需 mall:audit 权限）
func (h *ProductHandler) Audit(ctx plugin.Context) {
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallProductAuditError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核完成", nil))
}

// UpdatePromotion 更新商品推广配置
// PUT /api/v1/mall/admin/products/:id/promotion  （需 mall:audit 权限）
func (h *ProductHandler) UpdatePromotion(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.PromotionRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.UpdatePromotion(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallProductError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("推广配置已更新", nil))
}

// IncrViewCount 增加浏览数
// POST /api/v1/mall/products/:id/view  （公开）
func (h *ProductHandler) IncrViewCount(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.svc.IncrViewCount(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMallProductError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已记录浏览", nil))
}
