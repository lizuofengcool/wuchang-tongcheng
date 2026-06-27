// Package handler 文件HTTP处理层
package handler

import (
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/file/dto"
	"wuchang-tongcheng/internal/modules/file/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// Handler 文件HTTP处理器
type Handler struct {
	service service.FileService
}

// NewHandler 创建文件处理器
func NewHandler(svc service.FileService) *Handler {
	return &Handler{service: svc}
}

// Upload 上传文件
func (h *Handler) Upload(ctx plugin.Context) {
	userID, _ := ctx.Get(middleware.ContextUserID)
	uid, _ := userID.(uint)
	if uid == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	// 获取上传文件
	fh, err := ctx.FormFile()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeFileNotFound, "请上传文件"))
		return
	}

	// 打开文件流
	src, err := fh.Open()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeFileError, "打开文件失败"))
		return
	}
	defer src.Close()

	// 推断MIME类型（简单实现，生产可用 http.DetectContentType 读首部字节）
	mimeType := guessMIME(fh.Filename())

	// 获取地区ID
	regionID := uint(0)
	if v, ok := ctx.Get(middleware.RegionIDKey); ok {
		if id, ok := v.(uint); ok {
			regionID = id
		}
	}

	record, err := h.service.Upload(regionID, uid, fh.Filename(), mimeType, fh.Size(), src)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeFileUploadError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.SuccessWithMessage("上传成功", record))
}

// List 文件列表
func (h *Handler) List(ctx plugin.Context) {
	var req dto.ListFilesRequest
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeFileError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Delete 删除文件
func (h *Handler) Delete(ctx plugin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的文件ID"))
		return
	}
	if err := h.service.Delete(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeFileError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// guessMIME 根据扩展名简单推断MIME类型
func guessMIME(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".mp4":
		return "video/mp4"
	case ".pdf":
		return "application/pdf"
	case ".doc", ".docx":
		return "application/msword"
	case ".xls", ".xlsx":
		return "application/vnd.ms-excel"
	case ".zip":
		return "application/zip"
	case ".mp3":
		return "audio/mpeg"
	default:
		return "application/octet-stream"
	}
}
