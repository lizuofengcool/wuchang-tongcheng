// 商户中台模块 API 封装
// 对应后端 backend/internal/modules/merchant（路由前缀 /api/v1/merchant）
// 按模块分组：shop / staff / settle / category / verification
//
// 错误码范围：5701-5730（与 mall 5401-5430、dh114 5301-5340 等并列）
// 权限要求：admin 路由组需 merchant:audit 权限
import request from '@/utils/request'

// ====== 店铺（shops） ======

// 管理后台 - 店铺列表
export function getMerchantShopList(params) {
  return request.get('/merchant/admin/shops', { params })
}

// 管理后台 - 店铺详情
export function getMerchantShopDetail(id) {
  return request.get(`/merchant/admin/shops/${id}`)
}

// 管理后台 - 更新店铺状态
// data: { status: 0/1/2 }  0审核中 1正常 2停用
export function updateMerchantShopStatus(id, data) {
  return request.put(`/merchant/admin/shops/${id}/status`, data)
}

// 管理后台 - 信用分调整
// data: { delta: int, reason: string }
export function updateMerchantShopCredit(id, data) {
  return request.put(`/merchant/admin/shops/${id}/credit`, data)
}

// 管理后台 - 等级调整
// data: { level: 1-10 }
export function updateMerchantShopLevel(id, data) {
  return request.put(`/merchant/admin/shops/${id}/level`, data)
}

// C 端 - 店铺公开列表
export function getMerchantShopPublicList(params) {
  return request.get('/merchant/shops', { params })
}

// C 端 - 店铺详情
export function getMerchantShopPublicDetail(id) {
  return request.get(`/merchant/shops/${id}`)
}

// C 端 - 搜索店铺
export function searchMerchantShops(params) {
  return request.get('/merchant/shops/search', { params })
}

// C 端 - 我的店铺（需登录）
export function getMyMerchantShops(params) {
  return request.get('/merchant/shops/mine', { params })
}

// C 端 - 商户入驻
export function applyMerchantShop(data) {
  return request.post('/merchant/shops/apply', data)
}

// C 端 - 认领店铺
export function claimMerchantShop(data) {
  return request.post('/merchant/shops/claim', data)
}

// C 端 - 更新店铺
export function updateMerchantShop(id, data) {
  return request.put(`/merchant/shops/${id}`, data)
}

// ====== 员工（staff） ======

// 员工列表（公开只读 + 管理后台共用）
export function getMerchantStaffList(params) {
  return request.get('/merchant/staff', { params })
}

// 员工详情
export function getMerchantStaffDetail(id) {
  return request.get(`/merchant/staff/${id}`)
}

// 添加员工
export function createMerchantStaff(data) {
  return request.post('/merchant/staff', data)
}

// 更新员工
export function updateMerchantStaff(id, data) {
  return request.put(`/merchant/staff/${id}`, data)
}

// 删除员工
export function deleteMerchantStaff(id) {
  return request.delete(`/merchant/staff/${id}`)
}

// 分配员工权限
// data: { permissions: any }
export function assignMerchantStaffPermissions(id, data) {
  return request.put(`/merchant/staff/${id}/permissions`, data)
}

// 切换员工角色
// data: { role: 'owner'/'manager'/'clerk' }
export function switchMerchantStaffRole(id, data) {
  return request.put(`/merchant/staff/${id}/role`, data)
}

// ====== 结算（settles） ======

// 管理后台 - 结算列表
export function getMerchantSettleList(params) {
  return request.get('/merchant/admin/settles', { params })
}

// 管理后台 - 结算详情
export function getMerchantSettleDetail(id) {
  return request.get(`/merchant/admin/settles/${id}`)
}

// 管理后台 - 生成结算单
// data: { shop_id, period, total_amount, platform_rate }
export function generateMerchantSettle(data) {
  return request.post('/merchant/admin/settles', data)
}

// 管理后台 - 提现申请
export function withdrawMerchantSettle(id) {
  return request.put(`/merchant/admin/settles/${id}/withdraw`)
}

// 管理后台 - 提现审核
// data: { status: 1/2/3, reason: string }  1通过 2拒绝 3撤销
export function auditMerchantSettleWithdraw(id, data) {
  return request.put(`/merchant/admin/settles/${id}/audit`, data)
}

// 管理后台 - 按店铺汇总
export function getMerchantSettleSummaryByShop(params) {
  return request.get('/merchant/admin/settles/summary-by-shop', { params })
}

// 管理后台 - 按周期汇总
export function getMerchantSettleSummaryByPeriod(params) {
  return request.get('/merchant/admin/settles/summary-by-period', { params })
}

// C 端 - 按店铺查询结算
export function getMerchantSettlesByShop(shopId, params) {
  return request.get(`/merchant/shops/${shopId}/settles`, { params })
}

// ====== 类目（categories） ======

// 类目树（公开）
export function getMerchantCategoryTree() {
  return request.get('/merchant/categories/tree')
}

// 类目列表（公开 + 管理后台共用）
export function getMerchantCategoryList(params) {
  return request.get('/merchant/categories', { params })
}

// 类目详情（公开 + 管理后台共用）
export function getMerchantCategoryDetail(id) {
  return request.get(`/merchant/categories/${id}`)
}

// 管理后台 - 创建类目
export function createMerchantCategory(data) {
  return request.post('/merchant/admin/categories', data)
}

// 管理后台 - 更新类目
export function updateMerchantCategory(id, data) {
  return request.put(`/merchant/admin/categories/${id}`, data)
}

// 管理后台 - 删除类目
export function deleteMerchantCategory(id) {
  return request.delete(`/merchant/admin/categories/${id}`)
}

// 管理后台 - 更新类目状态
// data: { status: 0/1 }  0禁用 1启用
export function updateMerchantCategoryStatus(id, data) {
  return request.put(`/merchant/admin/categories/${id}/status`, data)
}

// ====== 认证（verifications） ======

// 管理后台 - 认证列表
export function getMerchantVerificationList(params) {
  return request.get('/merchant/admin/verifications', { params })
}

// 管理后台 - 审核认证
// data: { status: 1/2, audit_reason: string }  1通过 2拒绝
export function auditMerchantVerification(id, data) {
  return request.put(`/merchant/admin/verifications/${id}/audit`, data)
}

// C 端 - 认证列表
export function getMerchantVerificationPublicList(params) {
  return request.get('/merchant/verifications', { params })
}

// C 端 - 认证详情
export function getMerchantVerificationDetail(id) {
  return request.get(`/merchant/verifications/${id}`)
}

// C 端 - 提交认证
export function createMerchantVerification(data) {
  return request.post('/merchant/verifications', data)
}

// C 端 - 更新认证
export function updateMerchantVerification(id, data) {
  return request.put(`/merchant/verifications/${id}`, data)
}

// C 端 - 删除认证
export function deleteMerchantVerification(id) {
  return request.delete(`/merchant/verifications/${id}`)
}

// C 端 - 按店铺查询认证
export function getMerchantVerificationsByShop(shopId, params) {
  return request.get(`/merchant/shops/${shopId}/verifications`, { params })
}
