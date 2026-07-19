// 同城房屋租售模块 API 封装（完整版 v1.0）
// 对应后端路由前缀：/api/v1/house（公开/C端） + /api/v1/house/admin（管理后台）
// 涵盖：房源/发布/小区/经纪人/合同/看房/VR/房贷/分类/设施/审核规则/担保/成交/评价/举报/统计
import request from '@/utils/request'

// ====================================================================
// 一、房源管理（公开 + C 端登录 + 管理后台）
// ====================================================================

// --- 公开接口（无需登录） ---

// 房源列表（C 端，已发布+已审核通过）
export function listHouses(params) {
  return request.get('/house', { params })
}

// 搜索房源
export function searchHouses(params) {
  return request.get('/house/search', { params })
}

// 高级搜索
export function advancedSearchHouses(params) {
  return request.get('/house/advanced-search', { params })
}

// 附近房源
export function listNearbyHouses(params) {
  return request.get('/house/nearby', { params })
}

// 房源详情
export function getHouse(id) {
  return request.get(`/house/${id}`)
}

// 相似房源
export function listSimilarHouses(id, params) {
  return request.get(`/house/${id}/similar`, { params })
}

// 收藏状态（公开）
export function getHouseFavStatus(id) {
  return request.get(`/house/${id}/fav`)
}

// 增加联系次数（公开）
export function incrHouseContact(id) {
  return request.post(`/house/${id}/contact`)
}

// 增加分享次数（公开）
export function incrHouseShare(id) {
  return request.post(`/house/${id}/share`)
}

// 房源评价列表（公开）
export function listHouseReviews(id, params) {
  return request.get(`/house/${id}/reviews`, { params })
}

// 房源评价统计
export function getHouseReviewStats(id) {
  return request.get(`/house/${id}/reviews/stats`)
}

// 房源举报列表（公开）
export function listHouseReports(id, params) {
  return request.get(`/house/${id}/reports`, { params })
}

// 房源担保列表
export function listHouseEscrows(id, params) {
  return request.get(`/house/${id}/escrows`, { params })
}

// 房源成交列表
export function listHouseDeals(id, params) {
  return request.get(`/house/${id}/deals`, { params })
}

// 按目标查询评价
export function listReviewsByTarget(targetType, targetId, params) {
  return request.get(`/house/by-target/${targetType}/${targetId}/reviews`, { params })
}

// 按目标查询评价统计
export function getReviewStatsByTarget(targetType, targetId) {
  return request.get(`/house/by-target/${targetType}/${targetId}/reviews/stats`)
}

// 按目标查询举报
export function listReportsByTarget(targetType, targetId, params) {
  return request.get(`/house/by-target/${targetType}/${targetId}/reports`, { params })
}

// --- C 端登录接口 ---

// 我的房源
export function listMyHouses(params) {
  return request.get('/house/mine', { params })
}

// 我的收藏
export function listMyHouseFavs(params) {
  return request.get('/house/favorites', { params })
}

// 发布房源
export function createHouse(data) {
  return request.post('/house', data)
}

// 更新房源
export function updateHouse(id, data) {
  return request.put(`/house/${id}`, data)
}

// 删除房源
export function deleteHouse(id) {
  return request.delete(`/house/${id}`)
}

// 收藏房源
export function favHouse(id) {
  return request.post(`/house/${id}/fav`)
}

// 推广房源
export function updateHousePromotion(id, data) {
  return request.put(`/house/${id}/promotion`, data)
}

// --- 管理后台接口（需 house:audit 权限） ---

// 管理端列表
export function adminListHouses(params) {
  return request.get('/house/admin/list', { params })
}

// 管理端详情
export function adminGetHouse(id) {
  return request.get(`/house/admin/${id}`)
}

// 审核（audit_status: 0待审 1通过 2拒绝）
export function auditHouse(id, data) {
  return request.put(`/house/admin/${id}/audit`, data)
}

// 强制下架/恢复
export function adminUpdateHouseStatus(id, status) {
  return request.put(`/house/admin/${id}/status`, { status })
}

// 批量审核
export function batchAuditHouses(data) {
  return request.post('/house/admin/batch/audit', data)
}

// 批量状态变更
export function batchUpdateHouseStatus(data) {
  return request.post('/house/admin/batch/status', data)
}

// 批量删除
export function batchDeleteHouses(data) {
  return request.post('/house/admin/batch/delete', data)
}

// ====================================================================
// 二、房源发布管理（listing）
// ====================================================================

// 发布单列表（公开）
export function listListings(params) {
  return request.get('/house/listings', { params })
}

// 创建发布单
export function createListing(data) {
  return request.post('/house/listings', data)
}

// 更新发布单
export function updateListing(id, data) {
  return request.put(`/house/listings/${id}`, data)
}

// 删除发布单
export function deleteListing(id) {
  return request.delete(`/house/listings/${id}`)
}

// 发布单详情
export function getListing(id) {
  return request.get(`/house/listings/${id}`)
}

// 我的发布单
export function listMyListings(params) {
  return request.get('/house/listings/mine', { params })
}

// 刷新发布单
export function refreshListing(id) {
  return request.post(`/house/listings/${id}/refresh`)
}

// 管理端 - 发布单列表
export function adminListListings(params) {
  return request.get('/house/admin/listings', { params })
}

// 管理端 - 发布单审核
export function auditListing(id, data) {
  return request.put(`/house/admin/listings/${id}/audit`, data)
}

// 管理端 - 发布单状态变更
export function adminUpdateListingStatus(id, data) {
  return request.put(`/house/admin/listings/${id}/status`, data)
}

// ====================================================================
// 三、小区管理
// ====================================================================

// 小区列表（公开）
export function listCommunities(params) {
  return request.get('/house/communities', { params })
}

// 附近小区
export function listNearbyCommunities(params) {
  return request.get('/house/communities/nearby', { params })
}

// 小区详情
export function getCommunity(id) {
  return request.get(`/house/communities/${id}`)
}

// 创建小区
export function createCommunity(data) {
  return request.post('/house/communities', data)
}

// 更新小区
export function updateCommunity(id, data) {
  return request.put(`/house/communities/${id}`, data)
}

// 关注小区
export function followCommunity(id) {
  return request.post(`/house/communities/${id}/follow`)
}

// 关注状态
export function getCommunityFollowStatus(id) {
  return request.get(`/house/communities/${id}/follow`)
}

// 管理端 - 小区列表
export function adminListCommunities(params) {
  return request.get('/house/admin/communities', { params })
}

// 管理端 - 小区状态变更
export function adminUpdateCommunityStatus(id, data) {
  return request.put(`/house/admin/communities/${id}/status`, data)
}

// ====================================================================
// 四、经纪人管理
// ====================================================================

// 经纪人列表（公开）
export function listAgents(params) {
  return request.get('/house/agents', { params })
}

// 经纪人详情
export function getAgent(id) {
  return request.get(`/house/agents/${id}`)
}

// 创建经纪人
export function createAgent(data) {
  return request.post('/house/agents', data)
}

// 更新经纪人
export function updateAgent(id, data) {
  return request.put(`/house/agents/${id}`, data)
}

// 我的经纪人
export function getMyAgent() {
  return request.get('/house/agents/mine')
}

// 关注经纪人
export function followAgent(id) {
  return request.post(`/house/agents/${id}/follow`)
}

// 关注状态
export function getAgentFollowStatus(id) {
  return request.get(`/house/agents/${id}/follow`)
}

// 管理端 - 经纪人列表
export function adminListAgents(params) {
  return request.get('/house/admin/agents', { params })
}

// 管理端 - 经纪人审核
export function auditAgent(id, data) {
  return request.put(`/house/admin/agents/${id}/audit`, data)
}

// 管理端 - 经纪人上下线
export function updateAgentOnlineStatus(id, data) {
  return request.put(`/house/admin/agents/${id}/online-status`, data)
}

// ====================================================================
// 五、合同管理
// ====================================================================

// 合同列表
export function listContracts(params) {
  return request.get('/house/contracts', { params })
}

// 我的合同
export function listMyContracts(params) {
  return request.get('/house/contracts/mine', { params })
}

// 合同详情
export function getContract(id) {
  return request.get(`/house/contracts/${id}`)
}

// 创建合同
export function createContract(data) {
  return request.post('/house/contracts', data)
}

// 更新合同
export function updateContract(id, data) {
  return request.put(`/house/contracts/${id}`, data)
}

// 签署合同
export function signContract(id) {
  return request.post(`/house/contracts/${id}/sign`)
}

// 终止合同
export function terminateContract(id) {
  return request.post(`/house/contracts/${id}/terminate`)
}

// 管理端 - 合同列表
export function adminListContracts(params) {
  return request.get('/house/admin/contracts', { params })
}

// ====================================================================
// 六、看房预约
// ====================================================================

// 看房列表
export function listViewings(params) {
  return request.get('/house/viewings', { params })
}

// 我的看房
export function listMyViewings(params) {
  return request.get('/house/viewings/mine', { params })
}

// 看房详情
export function getViewing(id) {
  return request.get(`/house/viewings/${id}`)
}

// 创建看房预约
export function createViewing(data) {
  return request.post('/house/viewings', data)
}

// 更新看房
export function updateViewing(id, data) {
  return request.put(`/house/viewings/${id}`, data)
}

// 确认预约
export function confirmViewing(id) {
  return request.post(`/house/viewings/${id}/confirm`)
}

// 取消预约
export function cancelViewing(id) {
  return request.post(`/house/viewings/${id}/cancel`)
}

// 改期
export function rescheduleViewing(id) {
  return request.post(`/house/viewings/${id}/reschedule`)
}

// 完成
export function completeViewing(id) {
  return request.post(`/house/viewings/${id}/complete`)
}

// 管理端 - 看房列表
export function adminListViewings(params) {
  return request.get('/house/admin/viewings', { params })
}

// ====================================================================
// 七、VR 看房
// ====================================================================

// VR 列表（公开）
export function listVRTours(params) {
  return request.get('/house/vr-tours', { params })
}

// VR 详情
export function getVRTour(id) {
  return request.get(`/house/vr-tours/${id}`)
}

// 分享 VR
export function shareVRTour(id) {
  return request.post(`/house/vr-tours/${id}/share`)
}

// 创建 VR
export function createVRTour(data) {
  return request.post('/house/vr-tours', data)
}

// 更新 VR
export function updateVRTour(id, data) {
  return request.put(`/house/vr-tours/${id}`, data)
}

// 删除 VR
export function deleteVRTour(id) {
  return request.delete(`/house/vr-tours/${id}`)
}

// 发布 VR
export function publishVRTour(id) {
  return request.post(`/house/vr-tours/${id}/publish`)
}

// 下线 VR
export function offlineVRTour(id) {
  return request.post(`/house/vr-tours/${id}/offline`)
}

// ====================================================================
// 八、房贷方案
// ====================================================================

// 房贷列表（公开）
export function listMortgages(params) {
  return request.get('/house/mortgages', { params })
}

// 房贷详情
export function getMortgage(id) {
  return request.get(`/house/mortgages/${id}`)
}

// 房贷计算
export function calculateMortgage(data) {
  return request.post('/house/mortgages/calculate', data)
}

// 管理端 - 创建房贷方案
export function createMortgage(data) {
  return request.post('/house/admin/mortgages', data)
}

// 管理端 - 更新房贷方案
export function updateMortgage(id, data) {
  return request.put(`/house/admin/mortgages/${id}`, data)
}

// 管理端 - 删除房贷方案
export function deleteMortgage(id) {
  return request.delete(`/house/admin/mortgages/${id}`)
}

// 管理端 - 房贷状态变更
export function updateMortgageStatus(id, data) {
  return request.put(`/house/admin/mortgages/${id}/status`, data)
}

// ====================================================================
// 九、房源分类
// ====================================================================

// 分类列表（公开）
export function listCategories(params) {
  return request.get('/house/categories', { params })
}

// 全部分类
export function listAllCategories() {
  return request.get('/house/categories/all')
}

// 按父级查分类
export function listCategoriesByParent(parentId, params) {
  return request.get(`/house/categories/parent/${parentId}`, { params })
}

// 分类详情
export function getCategory(id) {
  return request.get(`/house/categories/${id}`)
}

// 管理端 - 创建分类
export function createCategory(data) {
  return request.post('/house/admin/categories', data)
}

// 管理端 - 更新分类
export function updateCategory(id, data) {
  return request.put(`/house/admin/categories/${id}`, data)
}

// 管理端 - 删除分类
export function deleteCategory(id) {
  return request.delete(`/house/admin/categories/${id}`)
}

// 管理端 - 分类状态变更
export function updateCategoryStatus(id, data) {
  return request.put(`/house/admin/categories/${id}/status`, data)
}

// ====================================================================
// 十、配套设施
// ====================================================================

// 设施列表（公开）
export function listFacilities(params) {
  return request.get('/house/facilities', { params })
}

// 全部设施
export function listAllFacilities() {
  return request.get('/house/facilities/all')
}

// 热门设施
export function listHotFacilities(params) {
  return request.get('/house/facilities/hot', { params })
}

// 设施详情
export function getFacility(id) {
  return request.get(`/house/facilities/${id}`)
}

// 管理端 - 创建设施
export function createFacility(data) {
  return request.post('/house/admin/facilities', data)
}

// 管理端 - 更新设施
export function updateFacility(id, data) {
  return request.put(`/house/admin/facilities/${id}`, data)
}

// 管理端 - 删除设施
export function deleteFacility(id) {
  return request.delete(`/house/admin/facilities/${id}`)
}

// 管理端 - 设施状态变更
export function updateFacilityStatus(id, data) {
  return request.put(`/house/admin/facilities/${id}/status`, data)
}

// ====================================================================
// 十一、审核规则
// ====================================================================

// 启用的审核规则（C 端发布时使用）
export function listEnabledAuditRules() {
  return request.get('/house/audit-rules/enabled')
}

// 管理端 - 规则列表
export function listAuditRules(params) {
  return request.get('/house/admin/audit-rules', { params })
}

// 管理端 - 规则详情
export function getAuditRule(id) {
  return request.get(`/house/admin/audit-rules/${id}`)
}

// 管理端 - 创建规则
export function createAuditRule(data) {
  return request.post('/house/admin/audit-rules', data)
}

// 管理端 - 更新规则
export function updateAuditRule(id, data) {
  return request.put(`/house/admin/audit-rules/${id}`, data)
}

// 管理端 - 删除规则
export function deleteAuditRule(id) {
  return request.delete(`/house/admin/audit-rules/${id}`)
}

// 管理端 - 规则状态变更
export function updateAuditRuleStatus(id, data) {
  return request.put(`/house/admin/audit-rules/${id}/status`, data)
}

// ====================================================================
// 十二、评价管理
// ====================================================================

// 评价列表（公开）
export function listReviews(params) {
  return request.get('/house/reviews', { params })
}

// 评价详情
export function getReview(id) {
  return request.get(`/house/reviews/${id}`)
}

// 我的评价
export function listMyReviews(params) {
  return request.get('/house/reviews/mine', { params })
}

// 创建评价
export function createReview(data) {
  return request.post('/house/reviews', data)
}

// 评价回复
export function replyReview(id, data) {
  return request.post(`/house/reviews/${id}/reply`, data)
}

// 评价追加
export function appendReview(id, data) {
  return request.post(`/house/reviews/${id}/append`, data)
}

// 评价点赞
export function likeReview(id) {
  return request.post(`/house/reviews/${id}/like`)
}

// 管理端 - 评价列表
export function adminListReviews(params) {
  return request.get('/house/admin/reviews', { params })
}

// 管理端 - 评价状态变更
export function updateReviewStatus(id, data) {
  return request.put(`/house/admin/reviews/${id}/status`, data)
}

// 管理端 - 批量评价
export function batchUpdateReviewStatus(data) {
  return request.post('/house/admin/reviews/batch', data)
}

// 管理端 - 删除评价
export function adminDeleteReview(id) {
  return request.delete(`/house/admin/reviews/${id}`)
}

// ====================================================================
// 十三、举报管理
// ====================================================================

// 举报列表（C 端）
export function listReports(params) {
  return request.get('/house/reports', { params })
}

// 我的举报
export function listMyReports(params) {
  return request.get('/house/reports/mine', { params })
}

// 举报详情
export function getReport(id) {
  return request.get(`/house/reports/${id}`)
}

// 按编号查举报
export function getReportByNo(no) {
  return request.get(`/house/reports/no/${no}`)
}

// 创建举报
export function createReport(data) {
  return request.post('/house/reports', data)
}

// 举报申诉
export function appealReport(id, data) {
  return request.post(`/house/reports/${id}/appeal`, data)
}

// 管理端 - 举报列表
export function adminListReports(params) {
  return request.get('/house/admin/reports', { params })
}

// 管理端 - 举报详情
export function adminGetReport(id) {
  return request.get(`/house/admin/reports/${id}`)
}

// 管理端 - 处理举报
export function processReport(id, data) {
  return request.put(`/house/admin/reports/${id}/process`, data)
}

// 管理端 - 申诉处理
export function processAppeal(id, data) {
  return request.put(`/house/admin/reports/${id}/appeal`, data)
}

// 管理端 - 批量状态变更
export function batchUpdateReportStatus(data) {
  return request.post('/house/admin/reports/batch', data)
}

// 管理端 - 待处理数量
export function countPendingReports() {
  return request.get('/house/admin/reports/pending-count')
}

// ====================================================================
// 十四、担保交易
// ====================================================================

// 创建担保
export function createEscrow(data) {
  return request.post('/house/escrows', data)
}

// 担保列表
export function listEscrows(params) {
  return request.get('/house/escrows', { params })
}

// 担保详情
export function getEscrow(id) {
  return request.get(`/house/escrows/${id}`)
}

// 我付的担保
export function listMyPayerEscrows(params) {
  return request.get('/house/escrows/mine-payer', { params })
}

// 我收的担保
export function listMyPayeeEscrows(params) {
  return request.get('/house/escrows/mine-payee', { params })
}

// 标记已付款
export function markEscrowPaid(id) {
  return request.post(`/house/escrows/${id}/pay`)
}

// 放款
export function releaseEscrow(id, data) {
  return request.post(`/house/escrows/${id}/release`, data)
}

// 退款
export function refundEscrow(id, data) {
  return request.post(`/house/escrows/${id}/refund`, data)
}

// 争议
export function disputeEscrow(id, data) {
  return request.post(`/house/escrows/${id}/dispute`, data)
}

// 取消
export function cancelEscrow(id, data) {
  return request.post(`/house/escrows/${id}/cancel`, data)
}

// 管理端 - 担保列表
export function adminListEscrows(params) {
  return request.get('/house/admin/escrows', { params })
}

// 管理端 - 争议担保
export function listDisputedEscrows(params) {
  return request.get('/house/admin/escrows/disputed', { params })
}

// 管理端 - 仲裁
export function arbitrateEscrow(id, data) {
  return request.put(`/house/admin/escrows/${id}/arbitrate`, data)
}

// ====================================================================
// 十五、成交记录
// ====================================================================

// 成交列表（公开）
export function listDeals(params) {
  return request.get('/house/deals', { params })
}

// 成交详情
export function getDeal(id) {
  return request.get(`/house/deals/${id}`)
}

// 创建成交
export function createDeal(data) {
  return request.post('/house/deals', data)
}

// 确认成交
export function confirmDeal(id) {
  return request.post(`/house/deals/${id}/confirm`)
}

// 取消成交
export function cancelDeal(id) {
  return request.post(`/house/deals/${id}/cancel`)
}

// 管理端 - 成交列表
export function adminListDeals(params) {
  return request.get('/house/admin/deals', { params })
}

// 管理端 - 完成成交
export function completeDeal(id) {
  return request.post(`/house/admin/deals/${id}/complete`)
}

// ====================================================================
// 十六、推荐 / 数据统计
// ====================================================================

// 我的推荐
export function listMyRecommendations(params) {
  return request.get('/house/recommendations/mine', { params })
}

// 推荐详情
export function getRecommendation(id) {
  return request.get(`/house/recommendations/${id}`)
}

// 标记推荐已点击
export function markRecClicked(id) {
  return request.post(`/house/recommendations/${id}/click`)
}

// 标记推荐已联系
export function markRecContacted(id) {
  return request.post(`/house/recommendations/${id}/contact`)
}

// 标记推荐已浏览
export function markRecViewed(id) {
  return request.post(`/house/recommendations/${id}/view`)
}

// 标记推荐已忽略
export function markRecDismissed(id) {
  return request.post(`/house/recommendations/${id}/dismiss`)
}

// 公开 - 价格趋势
export function getPriceTrend(params) {
  return request.get('/house/statistics/price-trend', { params })
}

// 平台总览（登录可看）
export function getOverviewStats() {
  return request.get('/house/statistics/overview')
}

// 管理端 - 统计列表
export function adminListStats(params) {
  return request.get('/house/admin/statistics', { params })
}

// 管理端 - 按类型查统计
export function adminListStatsByType(params) {
  return request.get('/house/admin/statistics/by-type', { params })
}

// 管理端 - 统计详情
export function adminGetStat(id) {
  return request.get(`/house/admin/statistics/${id}`)
}

// ====================================================================
// 兼容别名
// ====================================================================

export const adminListHouseReports = adminListReports
export const adminListHouseReviews = adminListReviews
export const getHouseDetail = adminGetHouse
export const getOverviewStatsHouse = getOverviewStats
