// Package handler 同城头条HTTP处理层
package handler

import (
	"net/http"
	"strconv"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/news/dto"
	"wuchang-tongcheng/internal/modules/news/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// Handler 头条HTTP处理器
type Handler struct {
	service service.NewsService
}

// NewHandler 创建头条处理器
func NewHandler(svc service.NewsService) *Handler {
	return &Handler{service: svc}
}

// getUserID 从上下文获取用户ID
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

// Create 创建头条
func (h *Handler) Create(ctx plugin.Context) {
	userID, username := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.CreateNewsRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}

	regionID := getRegionID(ctx)
	info, err := h.service.Create(regionID, userID, username, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeNewsPublishError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("发布成功", info))
}

// Update 更新头条
func (h *Handler) Update(ctx plugin.Context) {
	userID, _ := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的头条ID"))
		return
	}

	var req dto.UpdateNewsRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}

	if err := h.service.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeNewsError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除头条
func (h *Handler) Delete(ctx plugin.Context) {
	userID, _ := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的头条ID"))
		return
	}

	if err := h.service.Delete(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeNewsError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 获取头条详情
func (h *Handler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的头条ID"))
		return
	}

	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeNewsNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 头条列表
func (h *Handler) List(ctx plugin.Context) {
	var req dto.NewsListRequest
	_ = ctx.Bind(&req)

	regionID := getRegionID(ctx)
	pagination, list, err := h.service.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeNewsError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Search 头条全文检索（ES 优先，ES 不可用时降级到 DB LIKE）
func (h *Handler) Search(ctx plugin.Context) {
	var req dto.NewsSearchRequest
	_ = ctx.Bind(&req)
	if req.Keyword == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("关键词不能为空"))
		return
	}
	regionID := getRegionID(ctx)
	pagination, list, err := h.service.Search(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeNewsError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Like 点赞/取消点赞（toggle，幂等）
func (h *Handler) Like(ctx plugin.Context) {
	userID, _ := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的头条ID"))
		return
	}
	res, err := h.service.Like(userID, id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeNewsError, err.Error()))
		return
	}
	if res.Liked {
		ctx.JSON(http.StatusOK, response.SuccessWithMessage("点赞成功", res))
	} else {
		ctx.JSON(http.StatusOK, response.SuccessWithMessage("已取消点赞", res))
	}
}

// LikeStatus 查询当前用户对该头条的点赞状态
func (h *Handler) LikeStatus(ctx plugin.Context) {
	userID, _ := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Success(dto.LikeResponse{Liked: false, LikeCount: 0}))
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的头条ID"))
		return
	}
	res, err := h.service.LikeStatus(userID, id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeNewsNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(res))
}
