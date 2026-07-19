// 同城二手物品模块 API 封装（完整版 v2.0）
// 对应后端路由前缀：/api/v1/ershou
// 涵盖：商品/SKU/订单/拍卖/推广/物流/担保/退款/举报/评价/店铺/标签/品牌/型号/分类属性/审核规则/用户信用/统计/批量/导出/搜索推荐
import request from '@/utils/request'

// ====================================================================
// 一、商品管理（含 C 端公开/C 端登录/管理后台）
// ====================================================================

// --- 公开接口（无需登录） ---

// 二手物品列表（C端，已发布+已审核通过）
export function listErshous(params) {
  return request.get('/ershou', { params })
}

// 搜索（走 Elasticsearch）
export function searchErshous(params) {
  return request.get('/ershou/search', { params })
}

// 附近查询
export function listNearbyErshous(params) {
  return request.get('/ershou/nearby', { params })
}

// 详情
export function getErshou(id) {
  return request.get(`/ershou/${id}`)
}

// 留言列表（公开）
export function listErshouMessages(id, params) {
  return request.get(`/ershou/${id}/messages`, { params })
}

// 收藏状态查询（公开）
export function getErshouFavStatus(id) {
  return request.get(`/ershou/${id}/fav`)
}

// --- C 端登录接口 ---

// 发布二手物品
export function createErshou(data) {
  return request.post('/ershou', data)
}

// 更新二手物品
export function updateErshou(id, data) {
  return request.put(`/ershou/${id}`, data)
}

// 删除二手物品（仅发布者本人）
export function deleteErshou(id) {
  return request.delete(`/ershou/${id}`)
}

// 我的发布
export function listMyErshous(params) {
  return request.get('/ershou/mine', { params })
}

// 我的收藏
export function listMyErshouFavs(params) {
  return request.get('/ershou/favorites', { params })
}

// 收藏 / 取消收藏（toggle 语义）
export function toggleErshouFav(id) {
  return request.post(`/ershou/${id}/fav`)
}

// 发表留言
export function createErshouMessage(id, data) {
  return request.post(`/ershou/${id}/messages`, data)
}

// --- 管理后台接口（需 content:audit 权限） ---

// 管理端列表（分页 + 筛选）
// status/audit_status 为 null/空时后端返回全部；为 0 时筛选草稿/待审
export function adminListErshous(params) {
  return request.get('/ershou/admin/list', { params })
}

// 管理端详情（聚合详情：主信息+SKU+评价+店铺+推广+物流+担保+拍卖）
export function adminGetErshou(id) {
  return request.get(`/ershou/admin/${id}`)
}

// 审核（audit_status: 0待审 1通过 2拒绝；audit_reason 可选）
export function auditErshou(id, data) {
  return request.put(`/ershou/admin/${id}/audit`, data)
}

// 强制下架/恢复（status: 1发布 3下架 4过期）
export function adminUpdateErshouStatus(id, status) {
  return request.put(`/ershou/admin/${id}/status`, { status })
}

// ====================================================================
// 二、SKU 规格
// ====================================================================

// 商品 SKU 列表（公开）
export function listErshouSKUs(id) {
  return request.get(`/ershou/${id}/skus`)
}

// 创建 SKU（仅发布者本人）
export function createErshouSKU(id, data) {
  return request.post(`/ershou/${id}/skus`, data)
}

// 更新 SKU
export function updateErshouSKU(id, skuId, data) {
  return request.put(`/ershou/${id}/skus/${skuId}`, data)
}

// 删除 SKU
export function deleteErshouSKU(id, skuId) {
  return request.delete(`/ershou/${id}/skus/${skuId}`)
}

// ====================================================================
// 三、订单管理（需登录）
// ====================================================================

// 订单列表（role: buyer/seller/all）
export function listErshouOrders(params) {
  return request.get('/ershou/orders', { params })
}

// 订单详情
export function getErshouOrder(id) {
  return request.get(`/ershou/orders/${id}`)
}

// 创建订单
export function createErshouOrder(data) {
  return request.post('/ershou/orders', data)
}

// 订单状态变更（action: pay/ship/receive/cancel/complete）
export function updateErshouOrderStatus(id, data) {
  return request.put(`/ershou/orders/${id}/status`, data)
}

// 订单付款
export function payErshouOrder(id) {
  return request.post(`/ershou/orders/${id}/pay`)
}

// 订单发货
export function shipErshouOrder(id) {
  return request.post(`/ershou/orders/${id}/ship`)
}

// 订单确认收货
export function receiveErshouOrder(id) {
  return request.post(`/ershou/orders/${id}/receive`)
}

// 订单取消
export function cancelErshouOrder(id) {
  return request.post(`/ershou/orders/${id}/cancel`)
}

// ====================================================================
// 四、拍卖管理
// ====================================================================

// 拍卖列表（公开）
export function listErshouAuctions(params) {
  return request.get('/ershou/auctions', { params })
}

// 商品拍卖详情
export function getErshouAuctionByErshouId(ershouId) {
  return request.get(`/ershou/${ershouId}/auction`)
}

// 创建拍卖
export function createErshouAuction(ershouId, data) {
  return request.post(`/ershou/${ershouId}/auction`, data)
}

// 出价
export function bidErshouAuction(ershouId, data) {
  return request.post(`/ershou/${ershouId}/auction/bid`, data)
}

// 手动截拍
export function endErshouAuction(ershouId) {
  return request.post(`/ershou/${ershouId}/auction/end`)
}

// ====================================================================
// 五、付费推广
// ====================================================================

// 商品推广记录列表（公开）
export function listErshouPromotions(ershouId, params) {
  return request.get(`/ershou/${ershouId}/promotions`, { params })
}

// 推广效果统计
export function getErshouPromotionStats(ershouId) {
  return request.get(`/ershou/${ershouId}/promotions/stats`)
}

// 创建推广
export function createErshouPromotion(ershouId, data) {
  return request.post(`/ershou/${ershouId}/promotions`, data)
}

// ====================================================================
// 六、物流管理
// ====================================================================

// 创建物流
export function createErshouLogistics(orderId, data) {
  return request.post(`/ershou/orders/${orderId}/logistics`, data)
}

// 物流详情
export function getErshouLogistics(orderId) {
  return request.get(`/ershou/orders/${orderId}/logistics`)
}

// 更新物流
export function updateErshouLogistics(orderId, data) {
  return request.put(`/ershou/orders/${orderId}/logistics`, data)
}

// ====================================================================
// 七、担保交易
// ====================================================================

// 创建担保
export function createErshouEscrow(orderId, data) {
  return request.post(`/ershou/orders/${orderId}/escrow`, data)
}

// 担保详情
export function getErshouEscrow(orderId) {
  return request.get(`/ershou/orders/${orderId}/escrow`)
}

// 放款
export function releaseErshouEscrow(orderId, data) {
  return request.post(`/ershou/orders/${orderId}/escrow/release`, data)
}

// ====================================================================
// 八、退款管理
// ====================================================================

// 退款列表
export function listErshouRefunds(params) {
  return request.get('/ershou/refunds', { params })
}

// 订单退款详情
export function getErshouOrderRefund(orderId) {
  return request.get(`/ershou/orders/${orderId}/refund`)
}

// 申请退款
export function createErshouRefund(orderId, data) {
  return request.post(`/ershou/orders/${orderId}/refund`, data)
}

// 处理退款（action: approve/reject/arbitrate）
export function processErshouRefund(refundId, data) {
  return request.put(`/ershou/refunds/${refundId}/process`, data)
}

// ====================================================================
// 九、举报管理
// ====================================================================

// 举报列表（管理端）
export function listErshouReports(params) {
  return request.get('/ershou/reports', { params })
}

// 举报详情
export function getErshouReport(id) {
  return request.get(`/ershou/reports/${id}`)
}

// 创建举报
export function createErshouReport(data) {
  return request.post('/ershou/reports', data)
}

// 处理举报（status: 1-5；penalty_type: warning/limit/ban1d/ban7d/banForever）
export function processErshouReport(id, data) {
  return request.put(`/ershou/reports/${id}/process`, data)
}

// 商品被举报列表（公开）
export function listErshouReportsByErshouId(ershouId, params) {
  return request.get(`/ershou/${ershouId}/reports`, { params })
}

// ====================================================================
// 十、评价管理
// ====================================================================

// 评价列表
export function listErshouReviews(params) {
  return request.get('/ershou/reviews', { params })
}

// 评价详情
export function getErshouReview(id) {
  return request.get(`/ershou/reviews/${id}`)
}

// 商品评价列表（公开）
export function listErshouReviewsByErshouId(ershouId, params) {
  return request.get(`/ershou/${ershouId}/reviews`, { params })
}

// 商品评价统计
export function getErshouReviewStats(ershouId) {
  return request.get(`/ershou/${ershouId}/reviews/stats`)
}

// 创建评价
export function createErshouReview(orderId, data) {
  return request.post(`/ershou/orders/${orderId}/reviews`, data)
}

// 评价回复
export function replyErshouReview(id, data) {
  return request.post(`/ershou/reviews/${id}/reply`, data)
}

// ====================================================================
// 十一、店铺管理
// ====================================================================

// 店铺列表（公开）
export function listErshouShops(params) {
  return request.get('/ershou/shops', { params })
}

// 店铺详情
export function getErshouShop(id) {
  return request.get(`/ershou/shops/${id}`)
}

// 创建店铺
export function createErshouShop(data) {
  return request.post('/ershou/shops', data)
}

// 我的店铺
export function getMyErshouShop() {
  return request.get('/ershou/shops/mine')
}

// 更新店铺
export function updateErshouShop(id, data) {
  return request.put(`/ershou/shops/${id}`, data)
}

// 店铺审核（管理端）
export function auditErshouShop(id, data) {
  return request.put(`/ershou/shops/${id}/audit`, data)
}

// 店铺状态变更（管理端）
export function updateErshouShopStatus(id, data) {
  return request.put(`/ershou/shops/${id}/status`, data)
}

// 关注店铺
export function followErshouShop(id, data) {
  return request.post(`/ershou/shops/${id}/follow`, data || {})
}

// 取消关注
export function unfollowErshouShop(id) {
  return request.delete(`/ershou/shops/${id}/follow`)
}

// 店铺粉丝列表
export function listErshouShopFollowers(id, params) {
  return request.get(`/ershou/shops/${id}/followers`, { params })
}

// 我关注的店铺
export function listMyFollowingShops(params) {
  return request.get('/ershou/shops/following', { params })
}

// ====================================================================
// 十二、标签管理
// ====================================================================

// 标签列表（公开）
export function listErshouTags(params) {
  return request.get('/ershou/tags', { params })
}

// 热门标签
export function listHotErshouTags(params) {
  return request.get('/ershou/tags/hot', { params })
}

// 标签详情
export function getErshouTag(id) {
  return request.get(`/ershou/tags/${id}`)
}

// 创建标签（管理端）
export function createErshouTag(data) {
  return request.post('/ershou/tags', data)
}

// 更新标签
export function updateErshouTag(id, data) {
  return request.put(`/ershou/tags/${id}`, data)
}

// 删除标签
export function deleteErshouTag(id) {
  return request.delete(`/ershou/tags/${id}`)
}

// ====================================================================
// 十三、品牌型号管理
// ====================================================================

// 品牌列表
export function listErshouBrands(params) {
  return request.get('/ershou/brands', { params })
}

// 品牌详情
export function getErshouBrand(id) {
  return request.get(`/ershou/brands/${id}`)
}

// 品牌下型号列表
export function listErshouModelsByBrand(brandId, params) {
  return request.get(`/ershou/brands/${brandId}/models`, { params })
}

// 创建品牌
export function createErshouBrand(data) {
  return request.post('/ershou/brands', data)
}

// 更新品牌
export function updateErshouBrand(id, data) {
  return request.put(`/ershou/brands/${id}`, data)
}

// 删除品牌
export function deleteErshouBrand(id) {
  return request.delete(`/ershou/brands/${id}`)
}

// 型号列表
export function listErshouModels(params) {
  return request.get('/ershou/models', { params })
}

// 型号详情
export function getErshouModel(id) {
  return request.get(`/ershou/models/${id}`)
}

// 创建型号（指定品牌）
export function createErshouModel(brandId, data) {
  return request.post(`/ershou/brands/${brandId}/models`, data)
}

// 更新型号
export function updateErshouModel(id, data) {
  return request.put(`/ershou/models/${id}`, data)
}

// 删除型号
export function deleteErshouModel(id) {
  return request.delete(`/ershou/models/${id}`)
}

// ====================================================================
// 十四、分类属性管理
// ====================================================================

// 分类属性列表
export function listErshouCategoryAttrs(params) {
  return request.get('/ershou/category-attrs', { params })
}

// 分类属性详情
export function getErshouCategoryAttr(id) {
  return request.get(`/ershou/category-attrs/${id}`)
}

// 指定分类下属性列表
export function listErshouAttrsByCategory(categoryId, params) {
  return request.get(`/ershou/categories/${categoryId}/attrs`, { params })
}

// 创建分类属性
export function createErshouCategoryAttr(data) {
  return request.post('/ershou/category-attrs', data)
}

// 更新分类属性
export function updateErshouCategoryAttr(id, data) {
  return request.put(`/ershou/category-attrs/${id}`, data)
}

// 删除分类属性
export function deleteErshouCategoryAttr(id) {
  return request.delete(`/ershou/category-attrs/${id}`)
}

// ====================================================================
// 十五、审核规则管理
// ====================================================================

// 审核规则列表
export function listErshouAuditRules(params) {
  return request.get('/ershou/audit-rules', { params })
}

// 启用中的规则列表（C 端发布时使用）
export function listEnabledErshouAuditRules() {
  return request.get('/ershou/audit-rules/enabled')
}

// 审核规则详情
export function getErshouAuditRule(id) {
  return request.get(`/ershou/audit-rules/${id}`)
}

// 创建审核规则
export function createErshouAuditRule(data) {
  return request.post('/ershou/audit-rules', data)
}

// 更新审核规则
export function updateErshouAuditRule(id, data) {
  return request.put(`/ershou/audit-rules/${id}`, data)
}

// 删除审核规则
export function deleteErshouAuditRule(id) {
  return request.delete(`/ershou/audit-rules/${id}`)
}

// ====================================================================
// 十六、用户信用管理
// ====================================================================

// 当前用户信用（C 端）
export function getMyErshouCredit() {
  return request.get('/ershou/credit')
}

// 指定用户信用（管理端）
export function getErshouUserCredit(userId) {
  return request.get(`/ershou/credit/${userId}`)
}

// 更新用户信用（管理端）
export function updateErshouUserCredit(userId, data) {
  return request.put(`/ershou/credit/${userId}`, data)
}

// ====================================================================
// 十七、数据统计
// ====================================================================

// 平台总览统计
export function getErshouOverviewStats() {
  return request.get('/ershou/statistics/overview')
}

// 卖家数据统计
export function getErshouSellerStats() {
  return request.get('/ershou/statistics/seller')
}

// 热门商品 TOP N
export function getErshouHotItems(params) {
  return request.get('/ershou/statistics/hot-items', { params })
}

// 价格趋势
export function getErshouPriceTrend(params) {
  return request.get('/ershou/statistics/price-trend', { params })
}

// ====================================================================
// 十八、批量操作
// ====================================================================

// 批量审核
export function batchAuditErshou(data) {
  return request.post('/ershou/batch/audit', data)
}

// 批量状态变更
export function batchUpdateErshouStatus(data) {
  return request.post('/ershou/batch/status', data)
}

// 批量删除
export function batchDeleteErshou(data) {
  return request.post('/ershou/batch/delete', data)
}

// 批量导出
export function exportErshou(data) {
  return request.post('/ershou/batch/export', data, { responseType: 'blob' })
}

// ====================================================================
// 十九、搜索推荐（C 端，预留接口）
// ====================================================================

// 复用 searchErshous 即高级搜索入口
export function advancedSearchErshou(params) {
  return request.get('/ershou/search', { params })
}

// 相似推荐（基于商品详情内部推荐，预留）
export function getSimilarErshouItems(id, params) {
  return request.get(`/ershou/${id}/similar`, { params })
}

// 个性化推荐（基于用户行为，预留）
export function getPersonalRecommendErshou(params) {
  return request.get('/ershou/recommend', { params })
}

// 热搜词（预留）
export function getErshouHotSearch(params) {
  return request.get('/ershou/search/hot', { params })
}

// 搜索历史（C 端登录）
export function getErshouSearchHistory(params) {
  return request.get('/ershou/search/history', { params })
}

// ====================================================================
// 兼容旧导出名（保持向后兼容）
// ====================================================================

// 旧版别名：adminListErshous 已存在；以下别名便于其他模块统一引用
export const getErshouDetail = adminGetErshou
export const getPromotionStats = getErshouPromotionStats
export const getOverviewStats = getErshouOverviewStats
export const getSellerStats = getErshouSellerStats
export const getHotItems = getErshouHotItems
export const getPriceTrend = getErshouPriceTrend
export const batchAudit = batchAuditErshou
export const batchStatusUpdate = batchUpdateErshouStatus
export const batchDelete = batchDeleteErshou
export const exportErshouList = exportErshou
export const searchErshou = searchErshous
export const getSimilar = getSimilarErshouItems
export const getPersonalRecommend = getPersonalRecommendErshou
export const getHotSearch = getErshouHotSearch
export const getSearchHistory = getErshouSearchHistory
