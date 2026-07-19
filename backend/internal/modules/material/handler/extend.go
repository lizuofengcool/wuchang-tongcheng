// Package handler 素材中台扩展 HTTP 处理层
// 依据 014_material_full.sql：分类/标签/图片标签/搜索历史/相似结果/OCR/统计
package handler

import (
	"net/http"
	"strconv"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/material/dto"
	"wuchang-tongcheng/internal/modules/material/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ExtendHandler 素材中台扩展处理器
type ExtendHandler struct {
	extSvc service.MaterialExtendService
}

// NewExtendHandler 创建扩展 Handler 实例
func NewExtendHandler(extSvc service.MaterialExtendService) *ExtendHandler {
	return &ExtendHandler{extSvc: extSvc}
}

// ===== 分类 =====

// CreateCategory 创建分类（M 端）
// POST /api/v1/material/categories
func (h *ExtendHandler) CreateCategory(ctx plugin.Context) {
	var req dto.CreateCategoryRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.extSvc.CreateCategory(getRegionID(ctx), &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMaterialError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("分类创建成功", info))
}

// UpdateCategory 更新分类（M 端）
// POST /api/v1/material/categories/:id
func (h *ExtendHandler) UpdateCategory(ctx plugin.Context) {
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if id == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("分类ID无效"))
		return
	}
	var req dto.UpdateCategoryRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.extSvc.UpdateCategory(uint(id), &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMaterialCategoryNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// DeleteCategory 删除分类（M 端）
// DELETE /api/v1/material/categories/:id
func (h *ExtendHandler) DeleteCategory(ctx plugin.Context) {
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if id == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("分类ID无效"))
		return
	}
	if err := h.extSvc.DeleteCategory(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMaterialCategoryNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// ListCategories 分类列表
// GET /api/v1/material/categories
func (h *ExtendHandler) ListCategories(ctx plugin.Context) {
	parentID, _ := strconv.ParseUint(ctx.Query("parent_id"), 10, 32)
	status, _ := strconv.Atoi(ctx.DefaultQuery("status", "1"))
	page, pageSize := parsePagination(ctx)
	list, total, err := h.extSvc.ListCategories(uint(parentID), status, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMaterialError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// ===== 标签 =====

// CreateTag 创建标签（M 端）
// POST /api/v1/material/tags
func (h *ExtendHandler) CreateTag(ctx plugin.Context) {
	var req dto.CreateTagRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.extSvc.CreateTag(getRegionID(ctx), &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMaterialError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("标签创建成功", info))
}

// UpdateTag 更新标签（M 端）
// POST /api/v1/material/tags/:id
func (h *ExtendHandler) UpdateTag(ctx plugin.Context) {
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if id == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("标签ID无效"))
		return
	}
	var fields map[string]interface{}
	if err := ctx.Bind(&fields); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.extSvc.UpdateTag(uint(id), fields); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMaterialTagNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// DeleteTag 删除标签（M 端）
// DELETE /api/v1/material/tags/:id
func (h *ExtendHandler) DeleteTag(ctx plugin.Context) {
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if id == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("标签ID无效"))
		return
	}
	if err := h.extSvc.DeleteTag(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMaterialTagNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// ListTags 标签列表
// GET /api/v1/material/tags
func (h *ExtendHandler) ListTags(ctx plugin.Context) {
	tagType := ctx.Query("tag_type")
	page, pageSize := parsePagination(ctx)
	list, total, err := h.extSvc.ListTags(tagType, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMaterialError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// ===== 图片标签 =====

// AddImageTags 给图片打标签
// POST /api/v1/material/images/tags
func (h *ExtendHandler) AddImageTags(ctx plugin.Context) {
	var req dto.AddImageTagsRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.extSvc.AddImageTags(getRegionID(ctx), &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMaterialError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("标签添加成功", nil))
}

// RemoveImageTag 移除图片标签
// POST /api/v1/material/images/tags/remove
func (h *ExtendHandler) RemoveImageTag(ctx plugin.Context) {
	var req dto.RemoveImageTagRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.extSvc.RemoveImageTag(&req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMaterialError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("标签已移除", nil))
}

// ListImageTags 查询图片的标签
// GET /api/v1/material/images/:image_id/tags
func (h *ExtendHandler) ListImageTags(ctx plugin.Context) {
	imageID, _ := strconv.ParseUint(ctx.Param("image_id"), 10, 32)
	if imageID == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("图片ID无效"))
		return
	}
	list, err := h.extSvc.ListImageTags(uint(imageID))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMaterialError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ListImagesByTag 通过标签查图片
// GET /api/v1/material/tags/:id/images
func (h *ExtendHandler) ListImagesByTag(ctx plugin.Context) {
	tagID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if tagID == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("标签ID无效"))
		return
	}
	page, pageSize := parsePagination(ctx)
	list, total, err := h.extSvc.ListImagesByTag(uint(tagID), page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMaterialError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// ===== 搜索历史 =====

// ListMySearchHistory 我的搜索历史
// GET /api/v1/material/search/history
func (h *ExtendHandler) ListMySearchHistory(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	list, total, err := h.extSvc.ListMySearchHistory(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMaterialError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// ===== 相似结果 =====

// ListSimilarResults 相似图结果
// GET /api/v1/material/similar/:source_image_id
func (h *ExtendHandler) ListSimilarResults(ctx plugin.Context) {
	sourceImageID, _ := strconv.ParseUint(ctx.Param("source_image_id"), 10, 32)
	if sourceImageID == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("图片ID无效"))
		return
	}
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	list, err := h.extSvc.ListSimilarResults(uint(sourceImageID), limit)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMaterialError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ===== OCR =====

// RecognizeOCR 触发 OCR 识别
// POST /api/v1/material/ocr
func (h *ExtendHandler) RecognizeOCR(ctx plugin.Context) {
	var req dto.OCRRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.extSvc.RecognizeOCR(getRegionID(ctx), &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMaterialOCRFail, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("OCR 识别完成", info))
}

// GetOCRByImageID 查询图片 OCR 结果
// GET /api/v1/material/ocr/:image_id
func (h *ExtendHandler) GetOCRByImageID(ctx plugin.Context) {
	imageID, _ := strconv.ParseUint(ctx.Param("image_id"), 10, 32)
	if imageID == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("图片ID无效"))
		return
	}
	info, err := h.extSvc.GetOCRByImageID(uint(imageID))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMaterialOCRFail, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListOCRResults OCR 结果列表（M 端）
// GET /api/v1/material/admin/ocr
func (h *ExtendHandler) ListOCRResults(ctx plugin.Context) {
	page, pageSize := parsePagination(ctx)
	list, total, err := h.extSvc.ListOCRResults(page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMaterialError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// ===== 统计 =====

// Statistics 素材统计（M 端）
// GET /api/v1/material/admin/statistics
func (h *ExtendHandler) Statistics(ctx plugin.Context) {
	resp, err := h.extSvc.Statistics()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeMaterialError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}
