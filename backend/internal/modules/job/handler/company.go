// Package handler 公司 + 企业认证 HTTP 处理层
package handler

import (
	"net/http"
	"strconv"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/job/dto"
	"wuchang-tongcheng/internal/modules/job/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// CompanyHandler 公司 HTTP 处理器
type CompanyHandler struct {
	svc service.CompanyService
}

// NewCompanyHandler 创建公司 Handler 实例
func NewCompanyHandler(svc service.CompanyService) *CompanyHandler {
	return &CompanyHandler{svc: svc}
}

// getUserID 从上下文获取用户 ID（含审核员姓名）
func getUserIDAndName(ctx plugin.Context) (userID uint, username string) {
	if v, ok := ctx.Get(middleware.ContextUserID); ok {
		if id, ok := v.(uint); ok {
			userID = id
		}
	}
	if v, ok := ctx.Get(middleware.ContextUsername); ok {
		if name, ok := v.(string); ok {
			username = name
		}
	}
	return
}

// Create 创建公司
// POST /api/v1/job/companies
func (h *CompanyHandler) Create(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.CompanyCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.svc.Create(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCompanyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// Update 更新公司
// PUT /api/v1/job/companies/:id
func (h *CompanyHandler) Update(ctx plugin.Context) {
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
	var req dto.CompanyUpdateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCompanyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// GetByID 获取公司详情
// GET /api/v1/job/companies/:id
func (h *CompanyHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	userID, _, _, _ := getUserProfile(ctx)
	info, err := h.svc.GetByID(id, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCompanyNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetMyCompany 我的公司
// GET /api/v1/job/companies/mine
func (h *CompanyHandler) GetMyCompany(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	info, err := h.svc.GetMyCompany(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCompanyNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 公司列表
// GET /api/v1/job/companies
func (h *CompanyHandler) List(ctx plugin.Context) {
	var req dto.CompanyListQuery
	_ = ctx.Bind(&req)
	pagination, list, err := h.svc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCompanyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Follow 关注公司
// POST /api/v1/job/companies/:id/follow
func (h *CompanyHandler) Follow(ctx plugin.Context) {
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
	var req dto.CompanyFollowRequest
	_ = ctx.Bind(&req)
	if err := h.svc.Follow(userID, id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCompanyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("关注成功", nil))
}

// Unfollow 取消关注
// DELETE /api/v1/job/companies/:id/follow
func (h *CompanyHandler) Unfollow(ctx plugin.Context) {
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
	if err := h.svc.Unfollow(userID, id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCompanyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("取消关注成功", nil))
}

// ListFollowing 我关注的公司
// GET /api/v1/job/companies/following
func (h *CompanyHandler) ListFollowing(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.svc.ListFollowing(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCompanyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Audit 审核公司
// PUT /api/v1/job/admin/companies/:id/audit
func (h *CompanyHandler) Audit(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.CompanyAuditRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Audit(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCompanyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核成功", nil))
}

// ===== 企业认证 =====

// CreateCert 创建认证
// POST /api/v1/job/certifications
func (h *CompanyHandler) CreateCert(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.CertificationCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.svc.CreateCert(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCompanyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("提交认证成功", info))
}

// GetCert 认证详情
// GET /api/v1/job/certifications/:id
func (h *CompanyHandler) GetCert(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.svc.GetCert(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCompanyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListCerts 认证列表
// GET /api/v1/job/certifications
func (h *CompanyHandler) ListCerts(ctx plugin.Context) {
	var req dto.CertificationListQuery
	_ = ctx.Bind(&req)
	pagination, list, err := h.svc.ListCerts(req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCompanyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListCertsByCompany 公司的认证列表
// GET /api/v1/job/companies/:id/certifications
func (h *CompanyHandler) ListCertsByCompany(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	list, err := h.svc.ListCertsByCompany(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCompanyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ProcessCert 审核认证
// PUT /api/v1/job/admin/certifications/:id/process
func (h *CompanyHandler) ProcessCert(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	verifierID, verifierName := getUserIDAndName(ctx)
	var req dto.CertificationProcessRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.svc.ProcessCert(id, verifierID, verifierName, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCompanyError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核成功", info))
}

// parseIntQuery 解析整型查询参数（用于 limit 等）
func parseIntQuery(ctx plugin.Context, key string, defaultVal int) int {
	val := ctx.DefaultQuery(key, strconv.Itoa(defaultVal))
	n, _ := strconv.Atoi(val)
	return n
}
