// 同城商城模块 API 封装
// 对应后端 backend/internal/modules/mall（路由前缀 /api/v1/mall）
// 按模块分组：shop / product / sku / order / orderItem / payment / refund / cart / address / category / logistics / review / auditRule / report / statistic
import request from '@/utils/request'

// ====== 店铺（shops） ======

// 管理后台 - 店铺列表
export function getMallShopList(params) {
  return request.get('/mall/admin/shops', { params })
}

// 店铺详情
export function getMallShopDetail(id) {
  return request.get(`/mall/shops/${id}`)
}

// 管理后台 - 审核店铺
export function auditMallShop(id, data) {
  return request.put(`/mall/admin/shops/${id}/audit`, data)
}

// 管理后台 - 更新店铺推广配置
export function updateMallShopPromotion(id, data) {
  return request.put(`/mall/admin/shops/${id}/promotion`, data)
}

// 按用户查询店铺
export function getMallShopsByUser(userId, params) {
  return request.get(`/mall/shops/by-user/${userId}`, { params })
}

// 按分类查询店铺
export function getMallShopsByCategory(categoryId, params) {
  return request.get(`/mall/shops/by-category/${categoryId}`, { params })
}

// 搜索店铺
export function searchMallShops(params) {
  return request.get('/mall/shops/search', { params })
}

// 店铺列表（公开）
export function getMallShopPublicList(params) {
  return request.get('/mall/shops', { params })
}

// ====== 商品（products） ======

// 管理后台 - 商品列表
export function getMallProductList(params) {
  return request.get('/mall/admin/products', { params })
}

// 商品详情
export function getMallProductDetail(id) {
  return request.get(`/mall/products/${id}`)
}

// 管理后台 - 审核商品
export function auditMallProduct(id, data) {
  return request.put(`/mall/admin/products/${id}/audit`, data)
}

// 管理后台 - 更新商品推广配置
export function updateMallProductPromotion(id, data) {
  return request.put(`/mall/admin/products/${id}/promotion`, data)
}

// 按店铺查询商品
export function getMallProductsByShop(shopId, params) {
  return request.get(`/mall/products/by-shop/${shopId}`, { params })
}

// 按分类查询商品
export function getMallProductsByCategory(categoryId, params) {
  return request.get(`/mall/products/by-category/${categoryId}`, { params })
}

// 搜索商品
export function searchMallProducts(params) {
  return request.get('/mall/products/search', { params })
}

// 精选/热销/新品
export function getMallProductsFeatured(params) {
  return request.get('/mall/products/featured', { params })
}

export function getMallProductsHot(params) {
  return request.get('/mall/products/hot', { params })
}

export function getMallProductsNew(params) {
  return request.get('/mall/products/new', { params })
}

// ====== SKU（skus） ======

// 按商品查询 SKU
export function getMallSkusByProduct(productId) {
  return request.get(`/mall/skus/by-product/${productId}`)
}

// 按店铺查询 SKU
export function getMallSkusByShop(shopId, params) {
  return request.get(`/mall/skus/by-shop/${shopId}`, { params })
}

// SKU 详情
export function getMallSkuDetail(id) {
  return request.get(`/mall/skus/${id}`)
}

// 更新 SKU 库存
export function updateMallSkuStock(id, data) {
  return request.put(`/mall/skus/${id}/stock`, data)
}

// 批量更新库存
export function batchUpdateMallSkuStock(data) {
  return request.put('/mall/skus/batch-stock', data)
}

// ====== 订单（orders） ======

// 管理后台 - 订单列表
export function getMallOrderList(params) {
  return request.get('/mall/admin/orders', { params })
}

// 订单详情
export function getMallOrderDetail(id) {
  return request.get(`/mall/orders/${id}`)
}

// 按订单号查询
export function getMallOrderByNo(orderNo) {
  return request.get(`/mall/orders/by-no/${orderNo}`)
}

// 按店铺查询订单
export function getMallOrdersByShop(shopId, params) {
  return request.get(`/mall/orders/by-shop/${shopId}`, { params })
}

// 管理后台 - 关闭订单
export function closeMallOrder(id, data) {
  return request.put(`/mall/admin/orders/${id}/close`, data)
}

// 批量更新订单状态
export function batchUpdateMallOrderStatus(data) {
  return request.put('/mall/admin/orders/batch-status', data)
}

// 自动关闭超时订单
export function autoCloseMallOrders() {
  return request.post('/mall/admin/orders/auto-close')
}

// 自动确认收货
export function autoConfirmMallOrders() {
  return request.post('/mall/admin/orders/auto-confirm')
}

// 自动评价
export function autoReviewMallOrders() {
  return request.post('/mall/admin/orders/auto-review')
}

// ====== 订单明细（order-items） ======

// 管理后台 - 订单明细列表
export function getMallOrderItemList(params) {
  return request.get('/mall/admin/order-items', { params })
}

// 按订单查询明细
export function getMallOrderItemsByOrder(orderId) {
  return request.get(`/mall/order-items/by-order/${orderId}`)
}

// 订单明细详情
export function getMallOrderItemDetail(id) {
  return request.get(`/mall/order-items/${id}`)
}

// 更新评价状态
export function updateMallOrderItemReviewStatus(id, data) {
  return request.put(`/mall/admin/order-items/${id}/review-status`, data)
}

// 更新退款状态
export function updateMallOrderItemRefundStatus(id, data) {
  return request.put(`/mall/admin/order-items/${id}/refund-status`, data)
}

// ====== 支付（payments） ======

// 管理后台 - 支付列表
export function getMallPaymentList(params) {
  return request.get('/mall/admin/payments', { params })
}

// 支付详情
export function getMallPaymentDetail(id) {
  return request.get(`/mall/payments/${id}`)
}

// 按支付单号查询
export function getMallPaymentByNo(paymentNo) {
  return request.get(`/mall/payments/by-no/${paymentNo}`)
}

// 按订单查询支付
export function getMallPaymentByOrder(orderId) {
  return request.get(`/mall/payments/by-order/${orderId}`)
}

// 按店铺查询支付
export function getMallPaymentsByShop(shopId, params) {
  return request.get(`/mall/payments/by-shop/${shopId}`, { params })
}

// 支付统计
export function getMallPaymentStats(params) {
  return request.get('/mall/admin/payments/stats', { params })
}

// ====== 退款（refunds） ======

// 退款列表（公开）
export function getMallRefundList(params) {
  return request.get('/mall/refunds', { params })
}

// 退款详情
export function getMallRefundDetail(id) {
  return request.get(`/mall/refunds/${id}`)
}

// 按退款单号查询
export function getMallRefundByNo(refundNo) {
  return request.get(`/mall/refunds/by-refund-no/${refundNo}`)
}

// 按订单查询退款
export function getMallRefundsByOrder(orderId) {
  return request.get(`/mall/refunds/by-order/${orderId}`)
}

// 按店铺查询退款
export function getMallRefundsByShop(shopId, params) {
  return request.get(`/mall/refunds/by-shop/${shopId}`, { params })
}

// 退款统计
export function getMallRefundStats(params) {
  return request.get('/mall/refunds/stats', { params })
}

// 管理后台 - 处理退款
export function adminProcessMallRefund(id, data) {
  return request.put(`/mall/admin/refunds/${id}/admin-process`, data)
}

// ====== 购物车（cart） ======

// 当前用户购物车列表
export function getMallCartList() {
  return request.get('/mall/cart')
}

// 按店铺查询购物车
export function getMallCartByShop(shopId) {
  return request.get(`/mall/cart/by-shop/${shopId}`)
}

// 购物车汇总
export function getMallCartSummary(params) {
  return request.get('/mall/cart/summary', { params })
}

// 购物车数量
export function getMallCartCount() {
  return request.get('/mall/cart/count')
}

// 已选购物车数量
export function getMallCartSelectedCount() {
  return request.get('/mall/cart/count-selected')
}

// 已选购物车项
export function getMallCartSelected() {
  return request.get('/mall/cart/selected')
}

// 按店铺分组
export function getMallCartGroupByShop() {
  return request.get('/mall/cart/group-by-shop')
}

// 购物车项详情
export function getMallCartDetail(id) {
  return request.get(`/mall/cart/${id}`)
}

// 删除购物车项
export function deleteMallCartItem(id) {
  return request.delete(`/mall/cart/${id}`)
}

// ====== 收货地址（addresses） ======

// 管理后台 - 地址列表
export function getMallAddressList(params) {
  return request.get('/mall/admin/addresses', { params })
}

// 地址详情
export function getMallAddressDetail(id) {
  return request.get(`/mall/addresses/${id}`)
}

// ====== 商品分类（categories） ======

// 分类列表
export function getMallCategoryList(params) {
  return request.get('/mall/categories', { params })
}

// 分类树
export function getMallCategoryTree() {
  return request.get('/mall/categories/tree')
}

// 按父级查询子分类
export function getMallCategoriesByParent(parentId) {
  return request.get('/mall/categories/by-parent', { params: { parent_id: parentId } })
}

// 启用中的分类
export function getMallCategoriesEnabled() {
  return request.get('/mall/categories/enabled')
}

// 分类详情
export function getMallCategoryDetail(id) {
  return request.get(`/mall/categories/${id}`)
}

// 创建分类
export function createMallCategory(data) {
  return request.post('/mall/admin/categories', data)
}

// 更新分类
export function updateMallCategory(id, data) {
  return request.put(`/mall/admin/categories/${id}`, data)
}

// 删除分类
export function deleteMallCategory(id) {
  return request.delete(`/mall/admin/categories/${id}`)
}

// 更新分类状态
export function updateMallCategoryStatus(id, status) {
  return request.put(`/mall/admin/categories/${id}/status`, { status })
}

// ====== 物流（logistics） ======

// 物流列表
export function getMallLogisticsList(params) {
  return request.get('/mall/logistics', { params })
}

// 物流详情
export function getMallLogisticsDetail(id) {
  return request.get(`/mall/logistics/${id}`)
}

// 按订单查询物流
export function getMallLogisticsByOrder(orderId) {
  return request.get(`/mall/logistics/by-order/${orderId}`)
}

// 按物流单号查询
export function getMallLogisticsByTrackingNo(trackingNo) {
  return request.get('/mall/logistics/by-tracking-no', { params: { tracking_no: trackingNo } })
}

// 按店铺查询物流
export function getMallLogisticsByShop(shopId, params) {
  return request.get(`/mall/logistics/by-shop/${shopId}`, { params })
}

// 更新物流状态
export function updateMallLogisticsStatus(id, status) {
  return request.put(`/mall/admin/logistics/${id}/status`, { status })
}

// ====== 评价（reviews） ======

// 管理后台 - 评价列表
export function getMallReviewList(params) {
  return request.get('/mall/admin/reviews', { params })
}

// 评价详情
export function getMallReviewDetail(id) {
  return request.get(`/mall/reviews/${id}`)
}

// 按商品查询评价
export function getMallReviewsByProduct(productId, params) {
  return request.get(`/mall/reviews/by-product/${productId}`, { params })
}

// 按店铺查询评价
export function getMallReviewsByShop(shopId, params) {
  return request.get(`/mall/reviews/by-shop/${shopId}`, { params })
}

// 按订单查询评价
export function getMallReviewsByOrder(orderId) {
  return request.get(`/mall/reviews/by-order/${orderId}`)
}

// 评价统计
export function getMallReviewStats(params) {
  return request.get('/mall/reviews/stats', { params })
}

// 管理后台 - 更新评价状态
export function updateMallReviewStatus(id, data) {
  return request.put(`/mall/admin/reviews/${id}/status`, data)
}

// 商家回复评价
export function replyMallReview(id, data) {
  return request.put(`/mall/reviews/${id}/reply`, data)
}

// ====== 审核规则（audit-rules） ======

// 审核规则列表
export function getMallAuditRuleList(params) {
  return request.get('/mall/audit-rules', { params })
}

// 已启用规则
export function getMallAuditRulesEnabled() {
  return request.get('/mall/audit-rules/enabled')
}

// 按类型查询规则
export function getMallAuditRulesByType(ruleType) {
  return request.get(`/mall/audit-rules/type/${ruleType}`)
}

// 规则详情
export function getMallAuditRuleDetail(id) {
  return request.get(`/mall/audit-rules/${id}`)
}

// 创建规则
export function createMallAuditRule(data) {
  return request.post('/mall/admin/audit-rules', data)
}

// 更新规则
export function updateMallAuditRule(id, data) {
  return request.put(`/mall/admin/audit-rules/${id}`, data)
}

// 删除规则
export function deleteMallAuditRule(id) {
  return request.delete(`/mall/admin/audit-rules/${id}`)
}

// 更新规则状态
export function updateMallAuditRuleStatus(id, status) {
  return request.put(`/mall/admin/audit-rules/${id}/status`, { status })
}

// 内容审核检测
export function checkMallAuditRule(data) {
  return request.post('/mall/admin/audit-rules/check', data)
}

// ====== 举报（reports） ======

// 举报列表
export function getMallReportList(params) {
  return request.get('/mall/reports', { params })
}

// 举报详情
export function getMallReportDetail(id) {
  return request.get(`/mall/reports/${id}`)
}

// 按用户查询举报
export function getMallReportsByUser(userId, params) {
  return request.get(`/mall/reports/by-user/${userId}`, { params })
}

// 按被举报对象查询
export function getMallReportsByTarget(params) {
  return request.get('/mall/reports/by-target', { params })
}

// 举报统计
export function getMallReportStats() {
  return request.get('/mall/reports/stats')
}

// 管理后台 - 处理举报
export function processMallReport(id, data) {
  return request.put(`/mall/admin/reports/${id}/process`, data)
}

// 管理后台 - 删除举报
export function deleteMallReport(id) {
  return request.delete(`/mall/admin/reports/${id}`)
}

// ====== 数据统计（statistics） ======

// 统计列表
export function getMallStatisticList(params) {
  return request.get('/mall/admin/statistics', { params })
}

// 写入/更新统计
export function upsertMallStatistic(data) {
  return request.post('/mall/admin/statistics', data)
}

// 平台总览
export function getMallStatisticOverview(params) {
  return request.get('/mall/admin/statistics/overview', { params })
}

// 统计汇总
export function getMallStatisticSummary(params) {
  return request.get('/mall/statistics/summary', { params })
}

// 热销商品榜
export function getMallHotProducts(params) {
  return request.get('/mall/statistics/hot-products', { params })
}

// 热门店铺榜
export function getMallHotShops(params) {
  return request.get('/mall/statistics/hot-shops', { params })
}

// 热门分类榜
export function getMallHotCategories(params) {
  return request.get('/mall/statistics/hot-categories', { params })
}
