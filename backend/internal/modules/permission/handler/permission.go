// Package handler 权限HTTP处理层
package handler

import (
	"net/http"
	"strconv"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/permission/dto"
	"wuchang-tongcheng/internal/modules/permission/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// Handler 权限HTTP处理器
type Handler struct {
	service service.PermissionService
}

// NewHandler 创建权限处理器
func NewHandler(svc service.PermissionService) *Handler {
	return &Handler{service: svc}
}

func getUserID(ctx plugin.Context) uint {
	if v, ok := ctx.Get(middleware.ContextUserID); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}

func parseID(ctx plugin.Context) (uint, error) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// CreateRole 创建角色
func (h *Handler) CreateRole(ctx plugin.Context) {
	if getUserID(ctx) == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.CreateRoleRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}
	info, err := h.service.CreateRole(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeRoleAlreadyExists, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// UpdateRole 更新角色
func (h *Handler) UpdateRole(ctx plugin.Context) {
	if getUserID(ctx) == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的角色ID"))
		return
	}
	var req dto.UpdateRoleRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}
	if err := h.service.UpdateRole(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePermissionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// DeleteRole 删除角色
func (h *Handler) DeleteRole(ctx plugin.Context) {
	if getUserID(ctx) == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的角色ID"))
		return
	}
	if err := h.service.DeleteRole(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePermissionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetRoleByID 获取角色
func (h *Handler) GetRoleByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的角色ID"))
		return
	}
	info, err := h.service.GetRoleByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeRoleNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListRoles 角色列表
func (h *Handler) ListRoles(ctx plugin.Context) {
	list, err := h.service.ListRoles()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePermissionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// CreatePermission 创建权限
func (h *Handler) CreatePermission(ctx plugin.Context) {
	if getUserID(ctx) == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.CreatePermissionRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}
	info, err := h.service.CreatePermission(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePermissionAlreadyExists, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// DeletePermission 删除权限
func (h *Handler) DeletePermission(ctx plugin.Context) {
	if getUserID(ctx) == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的权限ID"))
		return
	}
	if err := h.service.DeletePermission(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePermissionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// ListPermissions 权限列表
func (h *Handler) ListPermissions(ctx plugin.Context) {
	list, err := h.service.ListPermissions()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePermissionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// AssignRoles 给用户分配角色
func (h *Handler) AssignRoles(ctx plugin.Context) {
	if getUserID(ctx) == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.AssignRolesRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}
	if err := h.service.AssignRolesToUser(&req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePermissionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("分配成功", nil))
}

// AssignPermissions 给角色分配权限
func (h *Handler) AssignPermissions(ctx plugin.Context) {
	if getUserID(ctx) == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.AssignPermissionsRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误"))
		return
	}
	if err := h.service.AssignPermissionsToRole(&req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePermissionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("分配成功", nil))
}

// MyPermissions 查询当前用户的权限
func (h *Handler) MyPermissions(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	perms, err := h.service.GetPermissionsByUserID(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePermissionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(perms))
}

// UserRoles 查询用户的角色
func (h *Handler) UserRoles(ctx plugin.Context) {
	if getUserID(ctx) == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的用户ID"))
		return
	}
	roles, err := h.service.GetRolesByUserID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePermissionError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(roles))
}
