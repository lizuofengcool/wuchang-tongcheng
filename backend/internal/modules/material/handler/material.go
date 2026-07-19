// Package handler 素材存储中台精简版HTTP处理层
package handler

import (
	"net/http"
	"strconv"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/material/dto"
	"wuchang-tongcheng/internal/modules/material/service"
)

// Handler 素材中台 HTTP 处理器
type Handler struct {
	svc service.MaterialService
}

// NewHandler 创建 Handler 实例
func NewHandler(svc service.MaterialService) *Handler {
	return &Handler{svc: svc}
}

// getUserID 从上下文获取登录用户ID
func getUserID(ctx plugin.Context) uint {
	if v, ok := ctx.Get(middleware.ContextUserID); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
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

// parsePagination 解析分页
func parsePagination(ctx plugin.Context) (page, pageSize int) {
	page, _ = strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ = strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return
}

// Upload 文件上传
// POST /api/v1/material/files （需登录）
func (h *Handler) Upload(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	file, err := ctx.FormFile()
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("文件不能为空: "+err.Error()))
		return
	}

	var req dto.UploadRequest
	_ = ctx.Bind(&req)

	src, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2901, "文件打开失败: "+err.Error()))
		return
	}
	defer src.Close()

	// MIME 通过 ctx.GetHeader 推断，缺失时由 service 根据 filename 推断
	mimeType := ctx.GetHeader("Content-Type")
	resp, err := h.svc.Upload(getRegionID(ctx), userID, &req, file.Filename(), file.Size(), mimeType, src)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2901, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("上传成功", resp))
}

// GetFile 查询文件
// GET /api/v1/material/files/:file_id
func (h *Handler) GetFile(ctx plugin.Context) {
	fileID := ctx.Param("file_id")
	if fileID == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("file_id 不能为空"))
		return
	}
	info, err := h.svc.GetFile(fileID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2902, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListFiles 文件列表
// GET /api/v1/material/files
func (h *Handler) ListFiles(ctx plugin.Context) {
	userID := getUserID(ctx)
	page, pageSize := parsePagination(ctx)

	req := &dto.FileInfoListRequest{
		UserID:   userID,
		FileType: ctx.Query("file_type"),
		Category: ctx.Query("category"),
		Page:     page,
		PageSize: pageSize,
	}
	list, total, err := h.svc.ListFiles(req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2901, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// DeleteFile 删除文件
// DELETE /api/v1/material/files/:file_id
func (h *Handler) DeleteFile(ctx plugin.Context) {
	fileID := ctx.Param("file_id")
	if fileID == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("file_id 不能为空"))
		return
	}
	if err := h.svc.DeleteFile(fileID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2901, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// SearchByImage 以图搜图
// POST /api/v1/material/search-by-image
func (h *Handler) SearchByImage(ctx plugin.Context) {
	var req dto.SearchByImageRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if req.RegionID == 0 {
		req.RegionID = getRegionID(ctx)
	}
	list, err := h.svc.SearchByImage(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2901, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// AddWatermark 添加水印
// POST /api/v1/material/watermark
func (h *Handler) AddWatermark(ctx plugin.Context) {
	var req dto.WatermarkRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.AddWatermark(&req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2901, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("水印添加成功", nil))
}

// GenerateThumbnail 生成缩略图
// POST /api/v1/material/thumbnails
func (h *Handler) GenerateThumbnail(ctx plugin.Context) {
	var req dto.ThumbnailRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	json, err := h.svc.GenerateThumbnail(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(2901, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("缩略图生成成功", json))
}
