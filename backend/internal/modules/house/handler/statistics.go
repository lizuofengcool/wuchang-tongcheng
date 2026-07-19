// Package handler 数据统计 + 房贷 + 担保 + 成交 + 推荐 + VR + 分类 + 配套 + 审核规则 HTTP 处理层
// 依据 v3.2.1 架构方案第五章：对标贝壳/链家
// 聚合 StatisticService/MortgageService/EscrowService/DealService/RecommendationService/VRTourService/CategoryService/FacilityService/AuditRuleService 9 个 service
package handler

import (
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/house/dto"
	"wuchang-tongcheng/internal/modules/house/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// StatisticsHandler 统计 + 房贷 + 担保 + 成交 + 推荐 + VR + 分类 + 配套 + 审核规则 聚合 HTTP 处理器
type StatisticsHandler struct {
	statSvc  service.StatisticService
	mortgSvc service.MortgageService
	escrowSvc service.EscrowService
	dealSvc  service.DealService
	recSvc   service.RecommendationService
	vrSvc    service.VRTourService
	catSvc   service.CategoryService
	facSvc   service.FacilityService
	ruleSvc  service.AuditRuleService
}

// NewStatisticsHandler 创建 Statistics Handler 实例
func NewStatisticsHandler(
	statSvc service.StatisticService,
	mortgSvc service.MortgageService,
	escrowSvc service.EscrowService,
	dealSvc service.DealService,
	recSvc service.RecommendationService,
	vrSvc service.VRTourService,
	catSvc service.CategoryService,
	facSvc service.FacilityService,
	ruleSvc service.AuditRuleService,
) *StatisticsHandler {
	return &StatisticsHandler{
		statSvc:   statSvc,
		mortgSvc:  mortgSvc,
		escrowSvc: escrowSvc,
		dealSvc:   dealSvc,
		recSvc:    recSvc,
		vrSvc:     vrSvc,
		catSvc:    catSvc,
		facSvc:    facSvc,
		ruleSvc:   ruleSvc,
	}
}

// ===== 统计 =====

// Overview 平台总览
// GET /api/v1/house/statistics/overview  （需登录 + house:audit 权限）
func (h *StatisticsHandler) Overview(ctx plugin.Context) {
	regionID := getRegionID(ctx)
	overview, err := h.statSvc.GetOverview(regionID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(overview))
}

// ListStats 统计列表
// GET /api/v1/admin/house/statistics  （需 house:audit 权限）
func (h *StatisticsHandler) ListStats(ctx plugin.Context) {
	var req dto.StatListQuery
	_ = ctx.Bind(&req)
	pagination, list, err := h.statSvc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// PriceTrend 价格趋势
// GET /api/v1/house/statistics/price-trend  （公开）
func (h *StatisticsHandler) PriceTrend(ctx plugin.Context) {
	start := ctx.Query("start_date")
	end := ctx.Query("end_date")
	statType := ctx.Query("stat_type")
	targetIDStr := ctx.Query("target_id")
	targetID, _ := parseIDStr(targetIDStr)
	list, err := h.statSvc.ListByDateRange(start, end, statType, targetID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// GetStat 统计记录详情
// GET /api/v1/admin/house/statistics/:id  （需 house:audit 权限）
func (h *StatisticsHandler) GetStat(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.statSvc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListStatsByType 按统计类型查询
// GET /api/v1/admin/house/statistics/by-type  （需 house:audit 权限）
func (h *StatisticsHandler) ListStatsByType(ctx plugin.Context) {
	statType := ctx.Query("stat_type")
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.statSvc.ListByType(statType, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ===== 房贷方案 =====

// CreateMortgage 创建房贷方案
// POST /api/v1/admin/house/mortgages  （需 house:audit 权限）
func (h *StatisticsHandler) CreateMortgage(ctx plugin.Context) {
	var req dto.MortgageCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.mortgSvc.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// UpdateMortgage 更新房贷方案
// PUT /api/v1/admin/house/mortgages/:id  （需 house:audit 权限）
func (h *StatisticsHandler) UpdateMortgage(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.MortgageCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.mortgSvc.Update(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// DeleteMortgage 删除房贷方案
// DELETE /api/v1/admin/house/mortgages/:id  （需 house:audit 权限）
func (h *StatisticsHandler) DeleteMortgage(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.mortgSvc.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetMortgage 房贷方案详情
// GET /api/v1/house/mortgages/:id  （公开）
func (h *StatisticsHandler) GetMortgage(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.mortgSvc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListMortgages 房贷方案列表
// GET /api/v1/house/mortgages  （公开）
func (h *StatisticsHandler) ListMortgages(ctx plugin.Context) {
	var req dto.MortgageListQuery
	_ = ctx.Bind(&req)
	pagination, list, err := h.mortgSvc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// CalculateMortgage 房贷计算
// POST /api/v1/house/mortgages/calculate  （公开）
func (h *StatisticsHandler) CalculateMortgage(ctx plugin.Context) {
	var req dto.MortgageCalculateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.mortgSvc.Calculate(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// UpdateMortgageStatus 更新房贷方案状态
// PUT /api/v1/admin/house/mortgages/:id/status  （需 house:audit 权限）
func (h *StatisticsHandler) UpdateMortgageStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateHouseStatusRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.mortgSvc.UpdateStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态更新成功", nil))
}

// ===== 担保交易 =====

// CreateEscrow 创建担保交易
// POST /api/v1/house/escrows  （需登录）
func (h *StatisticsHandler) CreateEscrow(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.EscrowCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.escrowSvc.Create(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// GetEscrow 担保交易详情
// GET /api/v1/house/escrows/:id  （需登录）
func (h *StatisticsHandler) GetEscrow(ctx plugin.Context) {
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
	info, err := h.escrowSvc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListEscrows 担保交易列表
// GET /api/v1/house/escrows  （需登录）
func (h *StatisticsHandler) ListEscrows(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.EscrowListQuery
	_ = ctx.Bind(&req)
	pagination, list, err := h.escrowSvc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListMyPayerEscrows 我作为付款方的担保列表
// GET /api/v1/house/escrows/mine-payer  （需登录）
func (h *StatisticsHandler) ListMyPayerEscrows(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.escrowSvc.ListByPayer(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListMyPayeeEscrows 我作为收款方的担保列表
// GET /api/v1/house/escrows/mine-payee  （需登录）
func (h *StatisticsHandler) ListMyPayeeEscrows(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.escrowSvc.ListByPayee(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListEscrowsByHouse 按房源查询担保交易
// GET /api/v1/house/houses/:id/escrows  （公开）
func (h *StatisticsHandler) ListEscrowsByHouse(ctx plugin.Context) {
	houseID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的房源ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.escrowSvc.ListByHouse(houseID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// MarkEscrowPaid 标记担保已支付
// POST /api/v1/house/escrows/:id/pay  （需登录 + 付款方）
func (h *StatisticsHandler) MarkEscrowPaid(ctx plugin.Context) {
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
	var req struct {
		PayMethod  string `json:"pay_method" binding:"omitempty,oneof=wechat alipay bank balance"`
		PayTradeNo string `json:"pay_trade_no" binding:"max=64"`
	}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.escrowSvc.MarkPaid(id, req.PayMethod, req.PayTradeNo); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("支付成功", nil))
}

// ReleaseEscrow 放款
// POST /api/v1/house/escrows/:id/release  （需登录）
func (h *StatisticsHandler) ReleaseEscrow(ctx plugin.Context) {
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
	_ = userID
	if err := h.escrowSvc.Release(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("放款成功", nil))
}

// RefundEscrow 退款
// POST /api/v1/house/escrows/:id/refund  （需登录）
func (h *StatisticsHandler) RefundEscrow(ctx plugin.Context) {
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
	_ = userID
	if err := h.escrowSvc.Refund(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("退款成功", nil))
}

// DisputeEscrow 发起争议
// POST /api/v1/house/escrows/:id/dispute  （需登录 + 付款方/收款方）
func (h *StatisticsHandler) DisputeEscrow(ctx plugin.Context) {
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
	var req dto.EscrowDisputeRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.escrowSvc.Dispute(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("争议已发起", nil))
}

// ArbitrateEscrow 仲裁（M 端）
// PUT /api/v1/admin/house/escrows/:id/arbitrate  （需 house:audit 权限）
func (h *StatisticsHandler) ArbitrateEscrow(ctx plugin.Context) {
	handlerID, _, _, _ := getUserProfile(ctx)
	if handlerID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.EscrowArbitrateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.escrowSvc.Arbitrate(id, handlerID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("仲裁完成", nil))
}

// CancelEscrow 取消担保
// POST /api/v1/house/escrows/:id/cancel  （需登录 + 付款方）
func (h *StatisticsHandler) CancelEscrow(ctx plugin.Context) {
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
	if err := h.escrowSvc.Cancel(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("取消成功", nil))
}

// AdminListEscrows 管理后台担保列表
// GET /api/v1/admin/house/escrows  （需 house:audit 权限）
func (h *StatisticsHandler) AdminListEscrows(ctx plugin.Context) {
	var req dto.EscrowListQuery
	_ = ctx.Bind(&req)
	pagination, list, err := h.escrowSvc.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListDisputedEscrows 争议中担保列表（M 端）
// GET /api/v1/admin/house/escrows/disputed  （需 house:audit 权限）
func (h *StatisticsHandler) ListDisputedEscrows(ctx plugin.Context) {
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.escrowSvc.ListDisputed(page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ===== 成交记录 =====

// CreateDeal 创建成交记录
// POST /api/v1/house/deals  （需登录）
func (h *StatisticsHandler) CreateDeal(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.DealCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	regionID := getRegionID(ctx)
	info, err := h.dealSvc.Create(regionID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// GetDeal 成交记录详情
// GET /api/v1/house/deals/:id  （需登录）
func (h *StatisticsHandler) GetDeal(ctx plugin.Context) {
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
	info, err := h.dealSvc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListDeals 成交记录列表
// GET /api/v1/house/deals  （公开）
func (h *StatisticsHandler) ListDeals(ctx plugin.Context) {
	var req dto.DealListQuery
	_ = ctx.Bind(&req)
	pagination, list, err := h.dealSvc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListDealsByHouse 按房源查询成交记录
// GET /api/v1/house/houses/:id/deals  （公开）
func (h *StatisticsHandler) ListDealsByHouse(ctx plugin.Context) {
	houseID, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的房源ID"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.dealSvc.ListByHouse(houseID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ConfirmDeal 确认成交
// POST /api/v1/house/deals/:id/confirm  （需登录）
func (h *StatisticsHandler) ConfirmDeal(ctx plugin.Context) {
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
	var req dto.DealConfirmRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.dealSvc.Confirm(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("确认成功", nil))
}

// CancelDeal 取消成交
// POST /api/v1/house/deals/:id/cancel  （需登录）
func (h *StatisticsHandler) CancelDeal(ctx plugin.Context) {
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
	var req dto.DealCancelRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.dealSvc.Cancel(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("取消成功", nil))
}

// CompleteDeal 完成成交（M 端）
// POST /api/v1/admin/house/deals/:id/complete  （需 house:audit 权限）
func (h *StatisticsHandler) CompleteDeal(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.dealSvc.Complete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已完成", nil))
}

// AdminListDeals 管理后台成交记录列表
// GET /api/v1/admin/house/deals  （需 house:audit 权限）
func (h *StatisticsHandler) AdminListDeals(ctx plugin.Context) {
	var req dto.DealListQuery
	_ = ctx.Bind(&req)
	pagination, list, err := h.dealSvc.AdminList(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ===== 推荐 =====

// ListMyRecommendations 我的推荐列表
// GET /api/v1/house/recommendations/mine  （需登录）
func (h *StatisticsHandler) ListMyRecommendations(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	pagination, list, err := h.recSvc.ListByUser(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// GetRecommendation 推荐详情
// GET /api/v1/house/recommendations/:id  （需登录）
func (h *StatisticsHandler) GetRecommendation(ctx plugin.Context) {
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
	info, err := h.recSvc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	_ = userID
	ctx.JSON(http.StatusOK, response.Success(info))
}

// MarkRecClicked 标记推荐已点击
// POST /api/v1/house/recommendations/:id/click  （需登录）
func (h *StatisticsHandler) MarkRecClicked(ctx plugin.Context) {
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
	if err := h.recSvc.MarkClicked(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("操作成功", nil))
}

// MarkRecContacted 标记推荐已联系
// POST /api/v1/house/recommendations/:id/contact  （需登录）
func (h *StatisticsHandler) MarkRecContacted(ctx plugin.Context) {
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
	if err := h.recSvc.MarkContacted(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("操作成功", nil))
}

// MarkRecViewed 标记推荐已查看
// POST /api/v1/house/recommendations/:id/view  （需登录）
func (h *StatisticsHandler) MarkRecViewed(ctx plugin.Context) {
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
	if err := h.recSvc.MarkViewed(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("操作成功", nil))
}

// MarkRecDismissed 标记推荐已忽略
// POST /api/v1/house/recommendations/:id/dismiss  （需登录）
func (h *StatisticsHandler) MarkRecDismissed(ctx plugin.Context) {
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
	if err := h.recSvc.MarkDismissed(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("操作成功", nil))
}

// ===== VR 看房 =====

// CreateVRTour 创建 VR 看房
// POST /api/v1/house/vr-tours  （需登录）
func (h *StatisticsHandler) CreateVRTour(ctx plugin.Context) {
	userID, username, _, _ := getUserProfile(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.VRTourCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.vrSvc.Create(userID, username, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// UpdateVRTour 更新 VR 看房
// PUT /api/v1/house/vr-tours/:id  （需登录 + 录制人本人）
func (h *StatisticsHandler) UpdateVRTour(ctx plugin.Context) {
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
	var req dto.VRTourCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.vrSvc.Update(id, userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// DeleteVRTour 删除 VR 看房
// DELETE /api/v1/house/vr-tours/:id  （需登录 + 录制人本人）
func (h *StatisticsHandler) DeleteVRTour(ctx plugin.Context) {
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
	if err := h.vrSvc.Delete(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetVRTour VR 看房详情
// GET /api/v1/house/vr-tours/:id  （公开）
func (h *StatisticsHandler) GetVRTour(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	// 浏览数自增
	_ = h.vrSvc.IncrViewCount(id)
	info, err := h.vrSvc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListVRTours VR 看房列表
// GET /api/v1/house/vr-tours  （公开）
func (h *StatisticsHandler) ListVRTours(ctx plugin.Context) {
	var req dto.VRTourListQuery
	_ = ctx.Bind(&req)
	pagination, list, err := h.vrSvc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// PublishVRTour 发布 VR 看房
// POST /api/v1/house/vr-tours/:id/publish  （需登录 + 录制人本人）
func (h *StatisticsHandler) PublishVRTour(ctx plugin.Context) {
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
	if err := h.vrSvc.Publish(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("发布成功", nil))
}

// OfflineVRTour 下架 VR 看房
// POST /api/v1/house/vr-tours/:id/offline  （需登录 + 录制人本人）
func (h *StatisticsHandler) OfflineVRTour(ctx plugin.Context) {
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
	if err := h.vrSvc.Offline(id, userID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("下架成功", nil))
}

// ShareVRTour 分享 VR 看房（自增分享数）
// POST /api/v1/house/vr-tours/:id/share  （公开）
func (h *StatisticsHandler) ShareVRTour(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.vrSvc.IncrShareCount(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("分享成功", nil))
}

// ===== 房源分类 =====

// CreateCategory 创建分类
// POST /api/v1/admin/house/categories  （需 house:audit 权限）
func (h *StatisticsHandler) CreateCategory(ctx plugin.Context) {
	var req dto.CategoryCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.catSvc.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// UpdateCategory 更新分类
// PUT /api/v1/admin/house/categories/:id  （需 house:audit 权限）
func (h *StatisticsHandler) UpdateCategory(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.CategoryCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.catSvc.Update(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// DeleteCategory 删除分类
// DELETE /api/v1/admin/house/categories/:id  （需 house:audit 权限）
func (h *StatisticsHandler) DeleteCategory(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.catSvc.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetCategory 分类详情
// GET /api/v1/house/categories/:id  （公开）
func (h *StatisticsHandler) GetCategory(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.catSvc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListCategories 分类列表
// GET /api/v1/house/categories  （公开）
func (h *StatisticsHandler) ListCategories(ctx plugin.Context) {
	var req dto.CategoryListQuery
	_ = ctx.Bind(&req)
	pagination, list, err := h.catSvc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListAllCategories 全部分类（树形）
// GET /api/v1/house/categories/all  （公开）
func (h *StatisticsHandler) ListAllCategories(ctx plugin.Context) {
	list, err := h.catSvc.ListAll()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ListCategoriesByParent 按父级查询子分类
// GET /api/v1/house/categories/parent/:parent_id  （公开）
func (h *StatisticsHandler) ListCategoriesByParent(ctx plugin.Context) {
	parentID, err := parseSubID(ctx, "parent_id")
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的父级ID"))
		return
	}
	list, err := h.catSvc.ListByParent(parentID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// UpdateCategoryStatus 更新分类状态
// PUT /api/v1/admin/house/categories/:id/status  （需 house:audit 权限）
func (h *StatisticsHandler) UpdateCategoryStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateHouseStatusRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.catSvc.UpdateStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态更新成功", nil))
}

// ===== 配套设施 =====

// CreateFacility 创建配套设施
// POST /api/v1/admin/house/facilities  （需 house:audit 权限）
func (h *StatisticsHandler) CreateFacility(ctx plugin.Context) {
	userID, _, _, _ := getUserProfile(ctx)
	var req dto.FacilityCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.facSvc.Create(userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// UpdateFacility 更新配套设施
// PUT /api/v1/admin/house/facilities/:id  （需 house:audit 权限）
func (h *StatisticsHandler) UpdateFacility(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.FacilityCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.facSvc.Update(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// DeleteFacility 删除配套设施
// DELETE /api/v1/admin/house/facilities/:id  （需 house:audit 权限）
func (h *StatisticsHandler) DeleteFacility(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.facSvc.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetFacility 配套设施详情
// GET /api/v1/house/facilities/:id  （公开）
func (h *StatisticsHandler) GetFacility(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.facSvc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListFacilities 配套设施列表
// GET /api/v1/house/facilities  （公开）
func (h *StatisticsHandler) ListFacilities(ctx plugin.Context) {
	var req dto.FacilityListQuery
	_ = ctx.Bind(&req)
	pagination, list, err := h.facSvc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListAllFacilities 全部配套设施
// GET /api/v1/house/facilities/all  （公开）
func (h *StatisticsHandler) ListAllFacilities(ctx plugin.Context) {
	list, err := h.facSvc.ListAll()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// ListHotFacilities 热门配套设施
// GET /api/v1/house/facilities/hot  （公开）
func (h *StatisticsHandler) ListHotFacilities(ctx plugin.Context) {
	limit := 10
	list, err := h.facSvc.ListHot(limit)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// UpdateFacilityStatus 更新配套设施状态
// PUT /api/v1/admin/house/facilities/:id/status  （需 house:audit 权限）
func (h *StatisticsHandler) UpdateFacilityStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateHouseStatusRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.facSvc.UpdateStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态更新成功", nil))
}

// ===== 审核规则 =====

// CreateAuditRule 创建审核规则
// POST /api/v1/admin/house/audit-rules  （需 house:audit 权限）
func (h *StatisticsHandler) CreateAuditRule(ctx plugin.Context) {
	var req dto.AuditRuleCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.ruleSvc.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseAuditRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", info))
}

// UpdateAuditRule 更新审核规则
// PUT /api/v1/admin/house/audit-rules/:id  （需 house:audit 权限）
func (h *StatisticsHandler) UpdateAuditRule(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.AuditRuleCreateRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.ruleSvc.Update(id, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseAuditRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// DeleteAuditRule 删除审核规则
// DELETE /api/v1/admin/house/audit-rules/:id  （需 house:audit 权限）
func (h *StatisticsHandler) DeleteAuditRule(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	if err := h.ruleSvc.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseAuditRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// GetAuditRule 审核规则详情
// GET /api/v1/admin/house/audit-rules/:id  （需 house:audit 权限）
func (h *StatisticsHandler) GetAuditRule(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	info, err := h.ruleSvc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseAuditRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListAuditRules 审核规则列表
// GET /api/v1/admin/house/audit-rules  （需 house:audit 权限）
func (h *StatisticsHandler) ListAuditRules(ctx plugin.Context) {
	var req dto.AuditRuleListQuery
	_ = ctx.Bind(&req)
	pagination, list, err := h.ruleSvc.List(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseAuditRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(utils.PageResult(list, pagination)))
}

// ListEnabledAuditRules 启用中的审核规则（C 端发布时使用）
// GET /api/v1/house/audit-rules/enabled  （公开）
func (h *StatisticsHandler) ListEnabledAuditRules(ctx plugin.Context) {
	list, err := h.ruleSvc.ListEnabled()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseAuditRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// UpdateAuditRuleStatus 更新审核规则状态
// PUT /api/v1/admin/house/audit-rules/:id/status  （需 house:audit 权限）
func (h *StatisticsHandler) UpdateAuditRuleStatus(ctx plugin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("无效的ID"))
		return
	}
	var req dto.UpdateHouseStatusRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.ruleSvc.UpdateStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeHouseAuditRuleError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("状态更新成功", nil))
}
