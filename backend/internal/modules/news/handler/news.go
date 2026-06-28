// Package handler 同城分类信息HTTP处理层
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

type Handler struct {
	service service.NewsService
}

func NewHandler(svc service.NewsService) *Handler {
	return &Handler{service: svc}
}

func getUserID(ctx plugin.Context) (uint, string) {
	userID, _ := ctx.Get(middleware.ContextUserID)
	username, _ := ctx.Get(middleware.ContextUsername)
	id, _ := userID.(uint)
	name, _ := username.(string)
	return id, name
}

func getRegionID(ctx plugin.Context) uint {
	if v, ok := ctx.Get(middleware.RegionIDKey); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return middleware.DefaultRegionID
}

func parseID(ctx plugin.Context) (uint, error) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// Create 创建分类信息
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

// Update 更新分类信息
func (h *Handler) Update(ctx plugin.Context) {
	userID, _ := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
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

// Delete 删除分类信息
func (h *Handler) Delete(ctx plugin.Context) {
	userID, _ := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	if err := h.service.Delete(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeNewsError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 获取详情
func (h *Handler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeNewsNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 分类信息列表
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

// Search 全文检索
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

// Like 点赞/取消点赞
func (h *Handler) Like(ctx plugin.Context) {
	userID, _ := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
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

// LikeStatus 查询点赞状态
func (h *Handler) LikeStatus(ctx plugin.Context) {
	userID, _ := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Success(dto.LikeResponse{Liked: false, LikeCount: 0}))
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	res, err := h.service.LikeStatus(userID, id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeNewsNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(res))
}

// ====== 收藏 ======

// Fav 收藏/取消收藏
func (h *Handler) Fav(ctx plugin.Context) {
	userID, _ := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	res, err := h.service.Fav(userID, id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeNewsError, err.Error()))
		return
	}
	if res.Faved {
		ctx.JSON(http.StatusOK, response.SuccessWithMessage("收藏成功", res))
	} else {
		ctx.JSON(http.StatusOK, response.SuccessWithMessage("已取消收藏", res))
	}
}

// FavStatus 查询收藏状态
func (h *Handler) FavStatus(ctx plugin.Context) {
	userID, _ := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Success(dto.FavResponse{Faved: false, FavCount: 0}))
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	res, err := h.service.FavStatus(userID, id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeNewsNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(res))
}

// ====== 评论 ======

// CreateComment 创建评论
func (h *Handler) CreateComment(ctx plugin.Context) {
	userID, username := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	newsID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	var req dto.CreateCommentRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}

	comment, err := h.service.CreateComment(newsID, userID, username, "", &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeNewsError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("评论成功", comment))
}

// ListComments 评论列表
func (h *Handler) ListComments(ctx plugin.Context) {
	newsID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	pageStr := ctx.Query("page")
	if pageStr == "" { pageStr = "1" }
	page, _ := strconv.Atoi(pageStr)
	pageSizeStr := ctx.Query("page_size")
	if pageSizeStr == "" { pageSizeStr = "20" }
	pageSize, _ := strconv.Atoi(pageSizeStr)

	list, total, err := h.service.ListComments(newsID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeNewsError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, &utils.Pagination{
		Page: page, PageSize: pageSize, Total: total,
	})))
}

// DeleteComment 删除评论
func (h *Handler) DeleteComment(ctx plugin.Context) {
	userID, _ := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的评论ID"))
		return
	}

	if err := h.service.DeleteComment(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeNewsError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// ====== 消息 ======

// ListMessages 消息列表
func (h *Handler) ListMessages(ctx plugin.Context) {
	userID, _ := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	pageStr := ctx.Query("page")
	if pageStr == "" { pageStr = "1" }
	page, _ := strconv.Atoi(pageStr)
	pageSizeStr := ctx.Query("page_size")
	if pageSizeStr == "" { pageSizeStr = "20" }
	pageSize, _ := strconv.Atoi(pageSizeStr)

	list, total, err := h.service.ListMessages(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeNewsError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, &utils.Pagination{
		Page: page, PageSize: pageSize, Total: total,
	})))
}

// UnreadCount 未读消息数
func (h *Handler) UnreadCount(ctx plugin.Context) {
	userID, _ := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Success(map[string]int64{"count": 0}))
		return
	}
	count, err := h.service.UnreadCount(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeNewsError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(map[string]int64{"count": count}))
}

// MarkRead 标记已读
func (h *Handler) MarkRead(ctx plugin.Context) {
	userID, _ := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req struct {
		IDs []uint `json:"ids"`
	}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}
	if err := h.service.MarkRead(userID, req.IDs); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeNewsError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已标记已读", nil))
}