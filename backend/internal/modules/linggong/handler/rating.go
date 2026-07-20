// Package handler 同城零工兼职 HTTP 处理层 - 双向评价
package handler

import (
	"net/http"
	"strconv"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/linggong/dto"
	"wuchang-tongcheng/internal/modules/linggong/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// RatingHandler 双向评价 HTTP 处理器
type RatingHandler struct {
	service service.RatingService
}

// NewRatingHandler 创建 RatingHandler 实例
func NewRatingHandler(svc service.RatingService) *RatingHandler {
	return &RatingHandler{service: svc}
}

// ===== C 端 =====

// Create 创建评价
// POST /api/v1/linggong/ratings  （需登录）
func (h *RatingHandler) Create(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.CreateRatingRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	regionID := getRegionID(ctx)
	info, err := h.service.Create(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongRatingDuplicate, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("评价成功", info))
}

// Update 更新评价（仅评价人本人）
// PUT /api/v1/linggong/ratings/:id  （需登录）
func (h *RatingHandler) Update(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	var req dto.UpdateRatingRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.service.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongRatingNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除评价（仅评价人本人）
// DELETE /api/v1/linggong/ratings/:id  （需登录）
func (h *RatingHandler) Delete(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
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
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongRatingNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetByID 评价详情
// GET /api/v1/linggong/ratings/:id  （公开）
func (h *RatingHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongRatingNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetByRatingNo 按评价编号查询
// GET /api/v1/linggong/ratings/no/:rating_no  （公开）
func (h *RatingHandler) GetByRatingNo(ctx plugin.Context) {
	ratingNo := ctx.Param("rating_no")
	if ratingNo == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的评价编号"))
		return
	}
	info, err := h.service.GetByRatingNo(ratingNo)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongRatingNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 评价列表
// GET /api/v1/linggong/ratings  （公开）
func (h *RatingHandler) List(ctx plugin.Context) {
	var req dto.RatingListRequest
	_ = ctx.Bind(&req)

	regionID := getRegionID(ctx)
	pagination, list, err := h.service.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByLinggong 按岗位查询评价
// GET /api/v1/linggong/:id/ratings  （公开）
func (h *RatingHandler) ListByLinggong(ctx plugin.Context) {
	linggongID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的岗位ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListByLinggong(linggongID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByRater 我发出的评价
// GET /api/v1/linggong/ratings/rater  （需登录）
func (h *RatingHandler) ListByRater(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListByRater(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByTarget 按被评价对象查询
// GET /api/v1/linggong/ratings/target/:target_type/:id  （公开）
func (h *RatingHandler) ListByTarget(ctx plugin.Context) {
	targetType := ctx.Param("target_type")
	targetID, err := parseSubID(ctx, "id")
	if err != nil || targetType == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的被评价对象"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListByTarget(targetType, targetID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Reply 回复评价
// POST /api/v1/linggong/ratings/:id/reply  （需登录）
func (h *RatingHandler) Reply(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.RatingReplyRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.Reply(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("回复成功", nil))
}

// Append 追加评价
// POST /api/v1/linggong/ratings/:id/append  （需登录）
func (h *RatingHandler) Append(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.RatingAppendRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.Append(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("追加成功", nil))
}

// Like 点赞评价
// POST /api/v1/linggong/ratings/:id/like  （公开）
func (h *RatingHandler) Like(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.service.Like(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("点赞成功", nil))
}

// Stats 评价统计
// GET /api/v1/linggong/ratings/stats  （公开）
func (h *RatingHandler) Stats(ctx plugin.Context) {
	targetType := ctx.Query("target_type")
	targetIDStr := ctx.Query("target_id")
	if targetType == "" || targetIDStr == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("target_type 和 target_id 必填"))
		return
	}
	idVal, err := strconv.ParseUint(targetIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的 target_id"))
		return
	}
	stats, err := h.service.GetStats(targetType, uint(idVal))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(stats))
}

// ===== M 端管理 =====

// AdminList 管理后台评价列表
// GET /api/v1/linggong/admin/ratings  （需 linggong:audit 权限）
func (h *RatingHandler) AdminList(ctx plugin.Context) {
	var req dto.RatingAdminListRequest
	_ = ctx.Bind(&req)

	pagination, list, err := h.service.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Audit 评价审核
// PUT /api/v1/linggong/admin/ratings/:id/audit  （需 linggong:audit 权限）
func (h *RatingHandler) Audit(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.RatingAuditRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.Audit(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeLinggongAuditError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核完成", nil))
}
