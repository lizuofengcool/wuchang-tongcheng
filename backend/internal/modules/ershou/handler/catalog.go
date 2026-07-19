// Package handler 标签/品牌/型号/分类属性 HTTP 处理层
// 依据 v3.2.1 架构方案：对标转转
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

// CatalogHandler 标签/品牌/型号/分类属性聚合 HTTP 处理器
type CatalogHandler struct {
	tagSvc          service.TagService
	brandSvc        service.BrandService
	modelSvc        service.ModelService
	categoryAttrSvc service.CategoryAttrService
}

// NewCatalogHandler 创建 Catalog Handler 实例
func NewCatalogHandler(
	tagSvc service.TagService,
	brandSvc service.BrandService,
	modelSvc service.ModelService,
	categoryAttrSvc service.CategoryAttrService,
) *CatalogHandler {
	return &CatalogHandler{
		tagSvc:          tagSvc,
		brandSvc:        brandSvc,
		modelSvc:        modelSvc,
		categoryAttrSvc: categoryAttrSvc,
	}
}

// ===== 标签 =====

// CreateTag 创建标签（M 端）
// POST /api/v1/ershou/tags  （需登录 + content:audit）
func (h *CatalogHandler) CreateTag(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.TagCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.tagSvc.Create(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("标签创建成功", resp))
}

// GetTag 标签详情
// GET /api/v1/ershou/tags/:id  （公开）
func (h *CatalogHandler) GetTag(ctx plugin.Context) {
	id, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的标签ID"))
		return
	}
	resp, err := h.tagSvc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// UpdateTag 更新标签
// PUT /api/v1/ershou/tags/:id  （需登录 + content:audit）
func (h *CatalogHandler) UpdateTag(ctx plugin.Context) {
	id, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的标签ID"))
		return
	}
	var req dto.TagCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.tagSvc.Update(id, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("标签更新成功", resp))
}

// DeleteTag 删除标签
// DELETE /api/v1/ershou/tags/:id  （需登录 + content:audit）
func (h *CatalogHandler) DeleteTag(ctx plugin.Context) {
	id, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的标签ID"))
		return
	}
	if err := h.tagSvc.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("标签已删除", nil))
}

// ListTags 标签列表
// GET /api/v1/ershou/tags  （公开）
func (h *CatalogHandler) ListTags(ctx plugin.Context) {
	tagType := ctx.Query("type")
	page, pageSize := parsePagination(ctx)
	statusStr := ctx.Query("status")
	isHotStr := ctx.Query("is_hot")
	var statusPtr *int
	if statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			statusPtr = &s
		}
	}
	var isHotPtr *bool
	if isHotStr != "" {
		b := isHotStr == "true" || isHotStr == "1"
		isHotPtr = &b
	}
	pagination, list, err := h.tagSvc.List(tagType, statusPtr, isHotPtr, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListHotTags 热门标签
// GET /api/v1/ershou/tags/hot  （公开）
func (h *CatalogHandler) ListHotTags(ctx plugin.Context) {
	limitStr := ctx.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)
	list, err := h.tagSvc.ListHot(limit)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ===== 品牌 =====

// CreateBrand 创建品牌（M 端）
// POST /api/v1/ershou/brands  （需登录 + content:audit）
func (h *CatalogHandler) CreateBrand(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.BrandCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.brandSvc.Create(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("品牌创建成功", resp))
}

// GetBrand 品牌详情
// GET /api/v1/ershou/brands/:id  （公开）
func (h *CatalogHandler) GetBrand(ctx plugin.Context) {
	id, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的品牌ID"))
		return
	}
	resp, err := h.brandSvc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// UpdateBrand 更新品牌
// PUT /api/v1/ershou/brands/:id  （需登录 + content:audit）
func (h *CatalogHandler) UpdateBrand(ctx plugin.Context) {
	id, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的品牌ID"))
		return
	}
	var req dto.BrandCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.brandSvc.Update(id, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("品牌更新成功", resp))
}

// DeleteBrand 删除品牌
// DELETE /api/v1/ershou/brands/:id  （需登录 + content:audit）
func (h *CatalogHandler) DeleteBrand(ctx plugin.Context) {
	id, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的品牌ID"))
		return
	}
	if err := h.brandSvc.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("品牌已删除", nil))
}

// ListBrands 品牌列表
// GET /api/v1/ershou/brands  （公开）
func (h *CatalogHandler) ListBrands(ctx plugin.Context) {
	keyword := ctx.Query("keyword")
	page, pageSize := parsePagination(ctx)
	statusStr := ctx.Query("status")
	var statusPtr *int
	if statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			statusPtr = &s
		}
	}
	pagination, list, err := h.brandSvc.List(keyword, statusPtr, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ===== 型号 =====

// CreateModel 创建型号（M 端）
// POST /api/v1/ershou/brands/:id/models  （需登录 + content:audit）
func (h *CatalogHandler) CreateModel(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	brandID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的品牌ID"))
		return
	}
	var req dto.ModelCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.modelSvc.Create(brandID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("型号创建成功", resp))
}

// GetModel 型号详情
// GET /api/v1/ershou/models/:id  （公开）
func (h *CatalogHandler) GetModel(ctx plugin.Context) {
	id, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的型号ID"))
		return
	}
	resp, err := h.modelSvc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// UpdateModel 更新型号
// PUT /api/v1/ershou/models/:id  （需登录 + content:audit）
func (h *CatalogHandler) UpdateModel(ctx plugin.Context) {
	id, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的型号ID"))
		return
	}
	var req dto.ModelCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.modelSvc.Update(id, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("型号更新成功", resp))
}

// DeleteModel 删除型号
// DELETE /api/v1/ershou/models/:id  （需登录 + content:audit）
func (h *CatalogHandler) DeleteModel(ctx plugin.Context) {
	id, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的型号ID"))
		return
	}
	if err := h.modelSvc.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("型号已删除", nil))
}

// ListModelsByBrandID 品牌下的型号列表
// GET /api/v1/ershou/brands/:id/models  （公开）
func (h *CatalogHandler) ListModelsByBrandID(ctx plugin.Context) {
	brandID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的品牌ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.modelSvc.ListByBrandID(brandID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListModels 型号列表（支持品牌筛选）
// GET /api/v1/ershou/models  （公开）
func (h *CatalogHandler) ListModels(ctx plugin.Context) {
	brandIDStr := ctx.Query("brand_id")
	brandID, _ := strconv.ParseUint(brandIDStr, 10, 32)
	keyword := ctx.Query("keyword")
	page, pageSize := parsePagination(ctx)
	statusStr := ctx.Query("status")
	var statusPtr *int
	if statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			statusPtr = &s
		}
	}
	pagination, list, err := h.modelSvc.List(uint(brandID), keyword, statusPtr, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ===== 分类属性 =====

// CreateCategoryAttr 创建分类属性（M 端）
// POST /api/v1/ershou/category-attrs  （需登录 + content:audit）
func (h *CatalogHandler) CreateCategoryAttr(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.CategoryAttrCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.categoryAttrSvc.Create(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("分类属性创建成功", resp))
}

// GetCategoryAttr 分类属性详情
// GET /api/v1/ershou/category-attrs/:id  （公开）
func (h *CatalogHandler) GetCategoryAttr(ctx plugin.Context) {
	id, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的属性ID"))
		return
	}
	resp, err := h.categoryAttrSvc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// UpdateCategoryAttr 更新分类属性
// PUT /api/v1/ershou/category-attrs/:id  （需登录 + content:audit）
func (h *CatalogHandler) UpdateCategoryAttr(ctx plugin.Context) {
	id, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的属性ID"))
		return
	}
	var req dto.CategoryAttrCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.categoryAttrSvc.Update(id, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("分类属性更新成功", resp))
}

// DeleteCategoryAttr 删除分类属性
// DELETE /api/v1/ershou/category-attrs/:id  （需登录 + content:audit）
func (h *CatalogHandler) DeleteCategoryAttr(ctx plugin.Context) {
	id, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的属性ID"))
		return
	}
	if err := h.categoryAttrSvc.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("分类属性已删除", nil))
}

// ListCategoryAttrsByCategoryID 分类下的属性列表
// GET /api/v1/ershou/categories/:id/attrs  （公开）
func (h *CatalogHandler) ListCategoryAttrsByCategoryID(ctx plugin.Context) {
	categoryID, err := parseSubID(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的分类ID"))
		return
	}
	list, err := h.categoryAttrSvc.ListByCategoryID(categoryID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ListCategoryAttrs 分类属性列表
// GET /api/v1/ershou/category-attrs  （公开）
func (h *CatalogHandler) ListCategoryAttrs(ctx plugin.Context) {
	categoryIDStr := ctx.Query("category_id")
	categoryID, _ := strconv.ParseUint(categoryIDStr, 10, 32)
	page, pageSize := parsePagination(ctx)
	statusStr := ctx.Query("status")
	var statusPtr *int
	if statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			statusPtr = &s
		}
	}
	pagination, list, err := h.categoryAttrSvc.List(uint(categoryID), statusPtr, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeErshouError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}
