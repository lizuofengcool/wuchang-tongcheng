// Package handler 同城拼车出行 HTTP 处理层 - 退款
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/pinche/dto"
	"wuchang-tongcheng/internal/modules/pinche/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// RefundHandler 退款 HTTP 处理器
type RefundHandler struct {
	service service.RefundService
}

// NewRefundHandler 创建 RefundHandler 实例
func NewRefundHandler(svc service.RefundService) *RefundHandler {
	return &RefundHandler{service: svc}
}

// AdminList 退款列表
// GET /api/v1/pinche/admin/refunds  （需 pinche:audit 权限）
func (h *RefundHandler) AdminList(ctx plugin.Context) {
	var req dto.RefundListRequest
	_ = ctx.Bind(&req)

	result, err := h.service.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(result))
}

// Process 处理退款
// PUT /api/v1/pinche/admin/refunds/:id/process  （需 pinche:audit 权限）
func (h *RefundHandler) Process(ctx plugin.Context) {
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

	var req dto.ProcessRefundRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.service.Process(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodePincheError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("处理成功", nil))
}
