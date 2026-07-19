// Package handler 合同电子化 HTTP 处理层
// 依据 v3.2.1 架构方案第五章：对标贝壳/链家
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/house/dto"
	"wuchang-tongcheng/internal/modules/house/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ContractHandler 合同 HTTP 处理器
type ContractHandler struct {
	service service.ContractService
}

// NewContractHandler 创建 ContractHandler 实例
func NewContractHandler(svc service.ContractService) *ContractHandler {
	return &ContractHandler{service: svc}
}

// ===== C 端 =====

// Create 创建合同
// POST /api/v1/house/contracts  （需登录）
func (h *ContractHandler) Create(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.ContractCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.service.Create(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseContractError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("合同创建成功", info))
}

// Update 更新合同（仅甲方或经纪人，状态必须为草稿/待签）
// PUT /api/v1/house/contracts/:id  （需登录）
func (h *ContractHandler) Update(ctx plugin.Context) {
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
	var req dto.ContractUpdateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseContractError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// GetByID 获取合同详情（甲方/乙方/经纪人/管理员可看）
// GET /api/v1/house/contracts/:id  （需登录）
func (h *ContractHandler) GetByID(ctx plugin.Context) {
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
	info, err := h.service.GetByID(id, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseContractNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 合同列表
// GET /api/v1/house/contracts  （需登录）
func (h *ContractHandler) List(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.ContractListQuery
	_ = ctx.Bind(&req)
	regionID := getRegionID(ctx)
	pagination, list, err := h.service.List(regionID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseContractError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListMine 我的合同列表（作为甲方/乙方/经纪人）
// GET /api/v1/house/contracts/mine  （需登录）
func (h *ContractHandler) ListMine(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListMine(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseContractError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Sign 签署合同（甲方/乙方/经纪人）
// POST /api/v1/house/contracts/:id/sign  （需登录）
func (h *ContractHandler) Sign(ctx plugin.Context) {
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
	var req dto.ContractSignRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.Sign(id, userID, req.Party); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseContractError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("签署成功", nil))
}

// Terminate 终止合同
// POST /api/v1/house/contracts/:id/terminate  （需登录 + 合同当事人）
func (h *ContractHandler) Terminate(ctx plugin.Context) {
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
	var req dto.ContractTerminateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.service.Terminate(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseContractError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("合同已终止", nil))
}

// ===== M 端 =====

// AdminList 管理后台合同列表
// GET /api/v1/admin/house/contracts  （需 house:audit 权限）
func (h *ContractHandler) AdminList(ctx plugin.Context) {
	var req dto.ContractAdminListQuery
	_ = ctx.Bind(&req)
	pagination, list, err := h.service.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseContractError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}
