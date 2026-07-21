// 团购优惠券模块 API 封装（管理后台）
// 对应后端 backend/internal/modules/groupbuy（路由前缀 /api/v1/groupbuy）
// 按模块分组：团购商品 / 团购订单 / 优惠券
// 管理后台调用 /admin 子路径，需 operator/super_admin/auditor 权限
import request from '@/utils/request'

// ====== 团购商品（admin） ======

// 管理后台 - 团购商品列表
export function getGroupbuyList(params) {
  return request.get('/groupbuy/admin/list', { params })
}

// 管理后台 - 团购商品详情
export function getGroupbuyDetail(id) {
  return request.get(`/groupbuy/admin/${id}`)
}

// 管理后台 - 创建团购商品
export function createGroupbuy(data) {
  return request.post('/groupbuy/admin', data)
}

// 管理后台 - 更新团购商品
export function updateGroupbuy(id, data) {
  return request.put(`/groupbuy/admin/${id}`, data)
}

// 管理后台 - 删除团购商品
export function deleteGroupbuy(id) {
  return request.delete(`/groupbuy/admin/${id}`)
}

// 管理后台 - 审核团购商品
export function auditGroupbuy(id, data) {
  return request.put(`/groupbuy/admin/${id}/audit`, data)
}

// 管理后台 - 更新团购商品状态（上下架）
export function updateGroupbuyStatus(id, status) {
  return request.put(`/groupbuy/admin/${id}/status`, { status })
}

// ====== 团购订单（admin） ======

// 管理后台 - 团购订单列表
export function getGroupbuyOrderList(params) {
  return request.get('/groupbuy/admin/orders', { params })
}

// 管理后台 - 团购订单详情
export function getGroupbuyOrderDetail(id) {
  return request.get(`/groupbuy/admin/orders/${id}`)
}

// 管理后台 - 关闭团购订单
export function closeGroupbuyOrder(id, data) {
  return request.put(`/groupbuy/admin/orders/${id}/close`, data)
}

// ====== 优惠券（admin） ======

// 管理后台 - 优惠券列表
export function getGroupbuyCouponList(params) {
  return request.get('/groupbuy/admin/coupons', { params })
}

// 管理后台 - 创建优惠券
export function createGroupbuyCoupon(data) {
  return request.post('/groupbuy/admin/coupons', data)
}

// 管理后台 - 更新优惠券
export function updateGroupbuyCoupon(id, data) {
  return request.put(`/groupbuy/admin/coupons/${id}`, data)
}

// 管理后台 - 删除优惠券
export function deleteGroupbuyCoupon(id) {
  return request.delete(`/groupbuy/admin/coupons/${id}`)
}
