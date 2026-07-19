// Package handler 同城车辆买卖 HTTP 处理层 - 统计/交易/推荐/审核规则/车型库
// 依据 v3.2.1 架构方案：对标瓜子数据中心/人人车推荐/懂车帝行情
//
// 本文件聚合 5 个 Handler：
//   - StatisticsHandler：核心统计（Overview/SellerOverview/HotCars/PriceTrend）
//   - TradeHandler：担保交易 + 合同（Escrow + Contract）
//   - RecommendationHandler：推荐记录
//   - AuditRuleHandler：审核规则管理
//   - CatalogHandler：车型库/分类/品牌（直连 Repository，无 Service 层）
//
// 设计说明：
//   - Category/Model/Brand 为全局配置数据（无 region_id），无需 Service 层封装业务逻辑
//   - 直连 Repository 模式可减少无谓的代码层级，符合现有架构现状
package handler

import (
	"net/http"
	"strconv"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/car/dto"
	"wuchang-tongcheng/internal/modules/car/repository"
	"wuchang-tongcheng/internal/modules/car/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// =====================================================================
// StatisticsHandler 核心统计
// =====================================================================

// StatisticsHandler 统计 HTTP 处理器
type StatisticsHandler struct {
	service service.StatisticService
}

// NewStatisticsHandler 创建 StatisticsHandler 实例
func NewStatisticsHandler(svc service.StatisticService) *StatisticsHandler {
	return &StatisticsHandler{service: svc}
}

// Overview 平台总览统计
// GET /api/v1/car/admin/statistics/overview  （需 car:audit 权限）
func (h *StatisticsHandler) Overview(ctx plugin.Context) {
	resp, err := h.service.Overview()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// SellerOverview 卖家统计
// GET /api/v1/car/statistics/seller  （需登录）
func (h *StatisticsHandler) SellerOverview(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	resp, err := h.service.SellerOverview(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// HotCars 热门车源
// GET /api/v1/car/statistics/hot-cars  （公开）
func (h *StatisticsHandler) HotCars(ctx plugin.Context) {
	limitStr := ctx.Query("limit")
	limit := 10
	if n, err := strconv.Atoi(limitStr); err == nil && n > 0 && n <= 100 {
		limit = n
	}

	list, err := h.service.HotCars(limit)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// PriceTrend 价格趋势
// GET /api/v1/car/statistics/price-trend  （公开）
func (h *StatisticsHandler) PriceTrend(ctx plugin.Context) {
	regionID := getRegionID(ctx)
	startDate := ctx.Query("start_date")
	endDate := ctx.Query("end_date")

	list, err := h.service.PriceTrend(regionID, startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// AdminList 管理后台统计记录列表
// GET /api/v1/car/admin/statistics  （需 car:audit 权限）
func (h *StatisticsHandler) AdminList(ctx plugin.Context) {
	var req dto.StatisticListRequest
	_ = ctx.Bind(&req)

	pagination, list, err := h.service.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// =====================================================================
// TradeHandler 担保交易 + 合同
// =====================================================================

// TradeHandler 交易聚合 HTTP 处理器（Escrow + Contract）
type TradeHandler struct {
	escrowSvc   service.EscrowService
	contractSvc service.ContractService
}

// NewTradeHandler 创建 TradeHandler 实例
func NewTradeHandler(escrowSvc service.EscrowService, contractSvc service.ContractService) *TradeHandler {
	return &TradeHandler{escrowSvc: escrowSvc, contractSvc: contractSvc}
}

// ===== 担保交易 =====

// GetEscrow 担保交易详情
// GET /api/v1/car/escrows/:id  （需登录）
func (h *TradeHandler) GetEscrow(ctx plugin.Context) {
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

	info, err := h.escrowSvc.GetByID(id, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetEscrowByNo 按担保单号查询
// GET /api/v1/car/escrows/no/:escrow_no  （需登录）
func (h *TradeHandler) GetEscrowByNo(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	escrowNo := ctx.Param("escrow_no")
	if escrowNo == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的担保单号"))
		return
	}

	info, err := h.escrowSvc.GetByEscrowNo(escrowNo, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListEscrows 我的担保交易
// GET /api/v1/car/escrows  （需登录）
func (h *TradeHandler) ListEscrows(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.EscrowListRequest
	_ = ctx.Bind(&req)

	pagination, list, err := h.escrowSvc.List(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListEscrowsByCarID 车源的担保交易
// GET /api/v1/car/cars/:id/escrows  （公开）
func (h *TradeHandler) ListEscrowsByCarID(ctx plugin.Context) {
	carID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的车源ID"))
		return
	}

	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.escrowSvc.ListByCarID(carID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// EscrowAction 担保操作（放款/退款/争议/仲裁）
// POST /api/v1/car/escrows/:id/action  （需登录）
func (h *TradeHandler) EscrowAction(ctx plugin.Context) {
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

	var req dto.EscrowActionRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.escrowSvc.Action(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("操作成功", nil))
}

// AdminListEscrows 管理后台担保交易列表
// GET /api/v1/car/admin/escrows  （需 car:audit 权限）
func (h *TradeHandler) AdminListEscrows(ctx plugin.Context) {
	var req dto.EscrowListRequest
	_ = ctx.Bind(&req)

	pagination, list, err := h.escrowSvc.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdminGetEscrow 管理后台担保交易详情
// GET /api/v1/car/admin/escrows/:id  （需 car:audit 权限）
func (h *TradeHandler) AdminGetEscrow(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	info, err := h.escrowSvc.AdminGetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// UpdateEscrowStatus 更新担保状态（M 端）
// PUT /api/v1/car/admin/escrows/:id/status  （需 car:audit 权限）
func (h *TradeHandler) UpdateEscrowStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	var req dto.AdminUpdateStatusRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.escrowSvc.UpdateStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态更新成功", nil))
}

// ===== 合同 =====

// GetContract 合同详情
// GET /api/v1/car/contracts/:id  （需登录）
func (h *TradeHandler) GetContract(ctx plugin.Context) {
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

	info, err := h.contractSvc.GetByID(id, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetContractByNo 按合同号查询
// GET /api/v1/car/contracts/no/:contract_no  （需登录）
func (h *TradeHandler) GetContractByNo(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	contractNo := ctx.Param("contract_no")
	if contractNo == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的合同号"))
		return
	}

	info, err := h.contractSvc.GetByContractNo(contractNo, userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListContracts 我的合同
// GET /api/v1/car/contracts  （需登录）
func (h *TradeHandler) ListContracts(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.ContractListRequest
	_ = ctx.Bind(&req)

	pagination, list, err := h.contractSvc.List(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListContractsByCarID 车源的合同
// GET /api/v1/car/cars/:id/contracts  （公开）
func (h *TradeHandler) ListContractsByCarID(ctx plugin.Context) {
	carID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的车源ID"))
		return
	}

	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.contractSvc.ListByCarID(carID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// SignContract 签署合同
// POST /api/v1/car/contracts/:id/sign  （需登录）
func (h *TradeHandler) SignContract(ctx plugin.Context) {
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

	if err := h.contractSvc.Sign(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("合同签署成功", nil))
}

// TerminateContract 终止合同
// POST /api/v1/car/contracts/:id/terminate  （需登录）
func (h *TradeHandler) TerminateContract(ctx plugin.Context) {
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
	if err := h.contractSvc.Terminate(id, userID, reason); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("合同已终止", nil))
}

// AdminListContracts 管理后台合同列表
// GET /api/v1/car/admin/contracts  （需 car:audit 权限）
func (h *TradeHandler) AdminListContracts(ctx plugin.Context) {
	var req dto.ContractListRequest
	_ = ctx.Bind(&req)

	pagination, list, err := h.contractSvc.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// AdminGetContract 管理后台合同详情
// GET /api/v1/car/admin/contracts/:id  （需 car:audit 权限）
func (h *TradeHandler) AdminGetContract(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	info, err := h.contractSvc.AdminGetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// UpdateContractStatus 更新合同状态（M 端）
// PUT /api/v1/car/admin/contracts/:id/status  （需 car:audit 权限）
func (h *TradeHandler) UpdateContractStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	var req dto.AdminUpdateStatusRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.contractSvc.UpdateStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态更新成功", nil))
}

// =====================================================================
// RecommendationHandler 推荐
// =====================================================================

// RecommendationHandler 推荐记录 HTTP 处理器
type RecommendationHandler struct {
	service service.RecommendationService
}

// NewRecommendationHandler 创建 RecommendationHandler 实例
func NewRecommendationHandler(svc service.RecommendationService) *RecommendationHandler {
	return &RecommendationHandler{service: svc}
}

// ListByUser 我的推荐
// GET /api/v1/car/recommendations  （需登录）
func (h *RecommendationHandler) ListByUser(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	recType := ctx.Query("rec_type")
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListByUser(userID, recType, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListByCarID 车源的推荐记录
// GET /api/v1/car/cars/:id/recommendations  （公开）
func (h *RecommendationHandler) ListByCarID(ctx plugin.Context) {
	carID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的车源ID"))
		return
	}

	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.service.ListByCarID(carID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// MarkClicked 标记推荐已点击
// POST /api/v1/car/recommendations/:id/click  （需登录）
func (h *RecommendationHandler) MarkClicked(ctx plugin.Context) {
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

	if err := h.service.MarkClicked(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已标记点击", nil))
}

// MarkContacted 标记推荐已联系
// POST /api/v1/car/recommendations/:id/contact  （需登录）
func (h *RecommendationHandler) MarkContacted(ctx plugin.Context) {
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

	if err := h.service.MarkContacted(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已标记联系", nil))
}

// MarkDismissed 标记推荐已忽略
// POST /api/v1/car/recommendations/:id/dismiss  （需登录）
func (h *RecommendationHandler) MarkDismissed(ctx plugin.Context) {
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

	if err := h.service.MarkDismissed(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已标记忽略", nil))
}

// AdminList 管理后台推荐列表
// GET /api/v1/car/admin/recommendations  （需 car:audit 权限）
func (h *RecommendationHandler) AdminList(ctx plugin.Context) {
	var req dto.RecommendationListRequest
	_ = ctx.Bind(&req)

	pagination, list, err := h.service.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Delete 删除推荐记录（M 端）
// DELETE /api/v1/car/admin/recommendations/:id  （需 car:audit 权限）
func (h *RecommendationHandler) Delete(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	if err := h.service.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// =====================================================================
// AuditRuleHandler 审核规则
// =====================================================================

// AuditRuleHandler 审核规则 HTTP 处理器
type AuditRuleHandler struct {
	service service.AuditRuleService
}

// NewAuditRuleHandler 创建 AuditRuleHandler 实例
func NewAuditRuleHandler(svc service.AuditRuleService) *AuditRuleHandler {
	return &AuditRuleHandler{service: svc}
}

// GetByID 审核规则详情
// GET /api/v1/car/admin/audit-rules/:id  （需 car:audit 权限）
func (h *AuditRuleHandler) GetByID(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的规则ID"))
		return
	}

	info, err := h.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeAuditRuleErrorCar, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// List 审核规则列表
// GET /api/v1/car/admin/audit-rules  （需 car:audit 权限）
func (h *AuditRuleHandler) List(ctx plugin.Context) {
	var req dto.AuditRuleListRequest
	_ = ctx.Bind(&req)

	pagination, list, err := h.service.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeAuditRuleErrorCar, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// Create 创建审核规则
// POST /api/v1/car/admin/audit-rules  （需 car:audit 权限）
func (h *AuditRuleHandler) Create(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	var req dto.CreateAuditRuleRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	info, err := h.service.Create(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeAuditRuleErrorCar, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核规则创建成功", info))
}

// Update 更新审核规则
// PUT /api/v1/car/admin/audit-rules/:id  （需 car:audit 权限）
func (h *AuditRuleHandler) Update(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的规则ID"))
		return
	}

	var req dto.UpdateAuditRuleRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.service.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeAuditRuleErrorCar, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// Delete 删除审核规则
// DELETE /api/v1/car/admin/audit-rules/:id  （需 car:audit 权限）
func (h *AuditRuleHandler) Delete(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的规则ID"))
		return
	}

	if err := h.service.Delete(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeAuditRuleErrorCar, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// UpdateStatus 更新审核规则状态
// PUT /api/v1/car/admin/audit-rules/:id/status  （需 car:audit 权限）
func (h *AuditRuleHandler) UpdateStatus(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}

	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的规则ID"))
		return
	}

	var req dto.AdminUpdateStatusRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if err := h.service.UpdateStatus(id, userID, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeAuditRuleErrorCar, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态更新成功", nil))
}

// Check 审核检查（内部调用，对外暴露用于调试）
// POST /api/v1/car/admin/audit-rules/check  （需 car:audit 权限）
func (h *AuditRuleHandler) Check(ctx plugin.Context) {
	var req dto.AuditCheckRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	resp, err := h.service.Check(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeAuditRuleErrorCar, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// =====================================================================
// CatalogHandler 车型库/分类/品牌（直连 Repository）
// =====================================================================

// CatalogHandler 车型库聚合 HTTP 处理器（直连 Model/Category/Brand Repository）
// 说明：Category/Model/Brand 为全局配置数据，无 Service 层
type CatalogHandler struct {
	modelRepo    repository.ModelRepository
	categoryRepo repository.CategoryRepository
	brandRepo    repository.BrandRepository
}

// NewCatalogHandler 创建 CatalogHandler 实例
func NewCatalogHandler(modelRepo repository.ModelRepository, categoryRepo repository.CategoryRepository, brandRepo repository.BrandRepository) *CatalogHandler {
	return &CatalogHandler{
		modelRepo:    modelRepo,
		categoryRepo: categoryRepo,
		brandRepo:    brandRepo,
	}
}

// ===== 车型库 =====

// ListModels 车型库列表
// GET /api/v1/car/models  （公开）
func (h *CatalogHandler) ListModels(ctx plugin.Context) {
	var query repository.ModelListQuery
	_ = ctx.Bind(&query)

	page, pageSize := parsePagination(ctx)
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := h.modelRepo.List(query, pagination)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	pagination.Total = total
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// GetModel 车型详情
// GET /api/v1/car/models/:id  （公开）
func (h *CatalogHandler) GetModel(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	info, err := h.modelRepo.FindByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListModelsByBrand 按品牌查询车型
// GET /api/v1/car/brands/:brand/models  （公开）
func (h *CatalogHandler) ListModelsByBrand(ctx plugin.Context) {
	brand := ctx.Param("brand")
	if brand == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的品牌标识"))
		return
	}

	page, pageSize := parsePagination(ctx)
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := h.modelRepo.ListByBrand(brand, pagination)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	pagination.Total = total
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ===== 分类 =====

// ListCategories 分类列表
// GET /api/v1/car/categories  （公开）
func (h *CatalogHandler) ListCategories(ctx plugin.Context) {
	var query repository.CategoryListQuery
	_ = ctx.Bind(&query)

	page, pageSize := parsePagination(ctx)
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := h.categoryRepo.List(query, pagination)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	pagination.Total = total
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// GetCategory 分类详情
// GET /api/v1/car/categories/:id  （公开）
func (h *CatalogHandler) GetCategory(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	info, err := h.categoryRepo.FindByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListCategoriesByLevel 按层级查询分类
// GET /api/v1/car/categories/level/:level  （公开）
func (h *CatalogHandler) ListCategoriesByLevel(ctx plugin.Context) {
	levelStr := ctx.Param("level")
	level, err := strconv.Atoi(levelStr)
	if err != nil || level < 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的层级"))
		return
	}

	list, err := h.categoryRepo.ListByLevel(level)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ListCategoriesByParent 按父级查询子分类
// GET /api/v1/car/categories/:id/children  （公开）
func (h *CatalogHandler) ListCategoriesByParent(ctx plugin.Context) {
	parentID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}

	list, err := h.categoryRepo.ListByParentID(parentID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ===== 品牌 =====

// ListBrands 品牌列表
// GET /api/v1/car/brands  （公开）
func (h *CatalogHandler) ListBrands(ctx plugin.Context) {
	var query repository.BrandListQuery
	_ = ctx.Bind(&query)

	page, pageSize := parsePagination(ctx)
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := h.brandRepo.ListBrands(query, pagination)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	pagination.Total = total
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListAllBrands 所有品牌（不分页）
// GET /api/v1/car/brands/all  （公开）
func (h *CatalogHandler) ListAllBrands(ctx plugin.Context) {
	list, err := h.brandRepo.ListAllBrands()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeCarError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}
