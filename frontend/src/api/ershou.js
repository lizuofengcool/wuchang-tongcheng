// 同城二手物品模块 API 封装
// 对应后端路由前缀：/api/v1/ershou
import request from '@/utils/request'

// ====== 公开接口（无需登录） ======

// 二手物品列表（C端，已发布+已审核通过）
export function listErshous(params) {
  return request.get('/ershou', { params })
}

// 搜索
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

// ====== 用户接口（需登录） ======

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

// ====== 管理接口（需登录 + content:audit 权限） ======

// 管理端列表（分页 + 筛选：region_id/user_id/category_id/status/audit_status/keyword）
// status/audit_status 为 null/空时后端返回全部；为 0 时筛选草稿/待审
export function adminListErshous(params) {
  return request.get('/ershou/admin/list', { params })
}

// 管理端详情
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
