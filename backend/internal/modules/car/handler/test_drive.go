// Package handler 同城车辆买卖 HTTP 处理层 - 试驾预约
// 依据 v3.2.1 架构方案：对标瓜子试驾/人人车预约/懂车帝看车
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 状态机：0待确认 1已确认 2已完成 3已取消 4已爽约 5已改期
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/car/dto"
	"wuchang-tongcheng/internal/modules/car/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// TestDriveHandler 试驾预约 HTTP 处理器
type TestDriveHandler struct {
	service service.TestDriveService
}

// NewTestDriveHandler 创建 TestDriveHandler 实例
func NewTestDriveHandler(svc service.TestDriveService) *TestDriveHandler {
	return &TestDriveHandler{service: svc}
}

// ===== C 端 =====

// Create 创建试驾预约
// POST /api/v1/car/test-drives  （需登录）
func (h *TestDriveHandler) Create(ctx plugin.Context) {
	userID, username, phone, avatar := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.CreateTestDriveRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	regionID := getRegionID(ctx)
	info, err := h.service.Create(regionID, userID, username, phone, avatar, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeTestDriveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("试驾预约成功", info))
}

// Update 更新试驾预约（仅预约人）
// PUT /api/v1/car/test-drives/:id  （需登录）
func (h *TestDriveHandler) Update(ctx plugin.Context) {
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

	var req dto.UpdateTestDriveRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.service.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeTestDriveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Cancel 取消试驾预约（仅预约人）
// POST /api/v1/car/test-drives/:id/cancel  （需登录）
func (h *TestDriveHandler) Cancel(ctx plugin.Context) {
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

	reason := ctx.Query("reason")
	if err := h.service.Cancel(id, userID, reason); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeTestDriveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("取消成功", nil))
}

// GetByID 获取试驾预约详情
// GET /api/v1/car/test-drives/:id  （需登录）
func (h *TestDriveHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeTestDriveNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 试驾预约列表（公开列表，按车源/经销商筛选）
// GET /api/v1/car/test-drives  （公开）
func (h *TestDriveHandler) List(ctx plugin.Context) {
	var req dto.TestDriveListRequest
	_ = ctx.Bind(&req)

	regionID := getRegionID(ctx)
	pagination, list, err := h.service.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeTestDriveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByUser 我的试驾预约
// GET /api/v1/car/test-drives/mine  （需登录）
func (h *TestDriveHandler) ListByUser(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListByUser(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeTestDriveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListBySales 销售顾问的试驾任务
// GET /api/v1/car/test-drives/sales  （需登录）
func (h *TestDriveHandler) ListBySales(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListBySales(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeTestDriveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByDealer 经销商的试驾预约
// GET /api/v1/car/dealers/:id/test-drives  （公开）
func (h *TestDriveHandler) ListByDealer(ctx plugin.Context) {
	dealerID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的经销商ID"))
		return
	}

	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListByDealer(dealerID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeTestDriveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// UploadLicense 上传驾照图片
// POST /api/v1/car/test-drives/:id/license  （需登录）
func (h *TestDriveHandler) UploadLicense(ctx plugin.Context) {
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

	var req dto.TestDriveLicenseUploadRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.service.UploadLicense(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeTestDriveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("驾照上传成功", nil))
}

// ===== M 端管理 =====

// AdminList 管理后台试驾预约列表
// GET /api/v1/car/admin/test-drives  （需 car:audit 权限）
func (h *TestDriveHandler) AdminList(ctx plugin.Context) {
	var req dto.TestDriveListRequest
	_ = ctx.Bind(&req)

	pagination, list, err := h.service.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeTestDriveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdminGetByID 管理后台试驾预约详情
// GET /api/v1/car/admin/test-drives/:id  （需 car:audit 权限）
func (h *TestDriveHandler) AdminGetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	info, err := h.service.AdminGetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeTestDriveNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// UpdateStatus 更新试驾预约状态
// PUT /api/v1/car/admin/test-drives/:id/status  （需 car:audit 权限）
func (h *TestDriveHandler) UpdateStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	var req dto.TestDriveStatusUpdateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.service.UpdateStatus(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeTestDriveError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态更新成功", nil))
}
