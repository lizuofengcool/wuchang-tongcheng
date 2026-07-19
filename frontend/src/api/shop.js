// 商家模块 API 封装
import request from '@/utils/request'

// ====== 公开接口（无需登录） ======

// 商家列表（公开）
export function listShops(params) {
  return request.get('/shop/list', { params })
}

// 商家详情
export function getShop(id) {
  return request.get(`/shop/${id}`)
}

// 商家相册
export function getShopImages(id) {
  return request.get(`/shop/${id}/images`)
}

// 商家评价列表（公开，仅返回已通过）
export function getShopReviews(id, params) {
  return request.get(`/shop/${id}/reviews`, { params })
}

// ====== 用户接口（需登录） ======

// 商家入驻申请
export function applyShop(data) {
  return request.post('/shop/apply', data)
}

// 我的店铺
export function getMyShop() {
  return request.get('/shop/my')
}

// 编辑我的店铺
export function updateMyShop(data) {
  return request.put('/shop/my', data)
}

// 上传店铺图片
export function addShopImage(data) {
  return request.post('/shop/my/images', data)
}

// 删除店铺图片
export function deleteShopImage(id) {
  return request.delete(`/shop/my/images/${id}`)
}

// 发表评价
export function createShopReview(shopId, data) {
  return request.post(`/shop/${shopId}/reviews`, data)
}

// ====== 管理接口 ======

// 管理端商家列表（分页 + 筛选：category_id/keyword/audit_status/status/is_recommend）
export function adminListShops(params) {
  return request.get('/shop/admin/list', { params })
}

// 审核店铺（audit_status: 1通过 2拒绝）
export function auditShop(id, data) {
  return request.put(`/shop/admin/${id}/audit`, data)
}

// 修改营业状态（status: 0歇业 1营业中 2休息中）
export function updateShopStatus(id, status) {
  return request.put(`/shop/admin/${id}/status`, { status })
}

// 设置推荐（is_recommend: 0否 1是）
export function setShopRecommend(id, isRecommend) {
  return request.put(`/shop/admin/${id}/recommend`, { is_recommend: isRecommend })
}

// 删除店铺
export function deleteShop(id) {
  return request.delete(`/shop/admin/${id}`)
}

// 管理端评价列表（分页 + 筛选：shop_id/status）
export function adminListShopReviews(params) {
  return request.get('/shop/admin/reviews', { params })
}

// 审核评价（status: 1通过 2拒绝）
export function auditShopReview(id, status) {
  return request.put(`/shop/admin/reviews/${id}/audit`, { status })
}
