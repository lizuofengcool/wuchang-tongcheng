// Package handler 用户HTTP处理层
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/user/dto"
	"wuchang-tongcheng/internal/modules/user/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// Handler 用户HTTP处理器
type Handler struct {
	service service.UserService
}

// NewHandler 创建用户处理器
func NewHandler(svc service.UserService) *Handler {
	return &Handler{service: svc}
}

// getUserID 从上下文中获取登录用户ID（由Auth中间件注入）
func getUserID(ctx plugin.Context) uint {
	if v, ok := ctx.Get(middleware.ContextUserID); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}

// Register 用户注册
func (h *Handler) Register(ctx plugin.Context) {
	var req dto.RegisterRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}

	info, err := h.service.Register(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeUserAlreadyExists, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.SuccessWithMessage("注册成功", info))
}

// Login 用户登录
func (h *Handler) Login(ctx plugin.Context) {
	var req dto.LoginRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}

	result, err := h.service.Login(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeUserPasswordError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.SuccessWithMessage("登录成功", result))
}

// GetUserInfo 获取当前用户信息
func (h *Handler) GetUserInfo(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	info, err := h.service.GetUserInfo(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeUserNotFound, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(info))
}

// UpdateProfile 更新个人资料
func (h *Handler) UpdateProfile(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.UpdateProfileRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}

	if err := h.service.UpdateProfile(userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeUserError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// ChangePassword 修改密码
func (h *Handler) ChangePassword(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.ChangePasswordRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}

	if err := h.service.ChangePassword(userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeUserPasswordError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.SuccessWithMessage("密码修改成功", nil))
}
