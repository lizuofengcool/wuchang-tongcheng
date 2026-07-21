// 分销合伙人中台模块 API 封装
// 对应后端 backend/internal/modules/distribution（路由前缀 /api/v1/distribution）
// 按模块分组：partner / channel / commission / level / withdrawal
//
// 错误码范围：5901-5930（与 merchant 5701-5730、mall 5401-5430、dh114 5301-5340 等并列）
// 权限要求：admin 路由组需 distribution:manage 权限
import request from '@/utils/request'

// ====== 合伙人（partners） ======

// 管理后台 - 合伙人列表
export function getDistributionPartnerList(params) {
  return request.get('/distribution/admin/partners', { params })
}

// 管理后台 - 合伙人详情
export function getDistributionPartnerDetail(id) {
  return request.get(`/distribution/admin/partners/${id}`)
}

// 管理后台 - 更新合伙人
// data: { level, commission_rate, parent_id, status }
export function updateDistributionPartner(id, data) {
  return request.put(`/distribution/admin/partners/${id}`, data)
}

// 管理后台 - 更新合伙人状态
// data: { status: 0/1/2/3/4 }  0待审核 1正常 2冻结 3拒绝 4退出
export function updateDistributionPartnerStatus(id, data) {
  return request.put(`/distribution/admin/partners/${id}/status`, data)
}

// 管理后台 - 升级合伙人等级
// data: { level: 1/2/3 }
export function upgradeDistributionPartner(id, data) {
  return request.put(`/distribution/admin/partners/${id}/upgrade`, data)
}

// 管理后台 - 调整佣金比例
// data: { commission_rate: 0-1 }
export function adjustDistributionPartnerRate(id, data) {
  return request.put(`/distribution/admin/partners/${id}/commission-rate`, data)
}

// C 端 - 合伙人公开列表
export function getDistributionPartnerPublicList(params) {
  return request.get('/distribution/partners', { params })
}

// C 端 - 合伙人详情
export function getDistributionPartnerPublicDetail(id) {
  return request.get(`/distribution/partners/${id}`)
}

// C 端 - 申请加入合伙人
// data: { parent_id, level, commission_rate }
export function applyDistributionPartner(data) {
  return request.post('/distribution/partners/apply', data)
}

// C 端 - 我的合伙人信息
export function getMyDistributionPartner() {
  return request.get('/distribution/partners/mine')
}

// C 端 - 上下级树
export function getDistributionPartnerTree(params) {
  return request.get('/distribution/partners/tree', { params })
}

// ====== 推广渠道（channels） ======

// 管理后台 - 渠道列表
export function getDistributionChannelList(params) {
  return request.get('/distribution/admin/channels', { params })
}

// C 端 - 渠道公开列表
export function getDistributionChannelPublicList(params) {
  return request.get('/distribution/channels', { params })
}

// C 端 - 渠道详情
export function getDistributionChannelPublicDetail(id) {
  return request.get(`/distribution/channels/${id}`)
}

// C 端 - 创建渠道
export function createDistributionChannel(data) {
  return request.post('/distribution/channels', data)
}

// C 端 - 我的渠道
export function getMyDistributionChannels(params) {
  return request.get('/distribution/channels/mine', { params })
}

// C 端 - 渠道统计
export function getDistributionChannelStats(params) {
  return request.get('/distribution/channels/stats', { params })
}

// C 端 - 更新渠道
export function updateDistributionChannel(id, data) {
  return request.put(`/distribution/channels/${id}`, data)
}

// C 端 - 删除渠道
export function deleteDistributionChannel(id) {
  return request.delete(`/distribution/channels/${id}`)
}

// C 端 - 渠道追踪（点击/注册/下单）
export function trackDistributionChannel(data) {
  return request.post('/distribution/channels/track', data)
}

// ====== 佣金记录（commissions） ======

// 管理后台 - 佣金列表
export function getDistributionCommissionList(params) {
  return request.get('/distribution/admin/commissions', { params })
}

// 管理后台 - 佣金汇总
export function getDistributionCommissionSummary(params) {
  return request.get('/distribution/admin/commissions/summary', { params })
}

// 管理后台 - 手动创建佣金记录
// data: { partner_id, order_id, channel_id?, order_amount, commission_amount, commission_rate, level }
export function createDistributionCommission(data) {
  return request.post('/distribution/admin/commissions', data)
}

// 管理后台 - 结算单条佣金
export function settleDistributionCommission(id) {
  return request.put(`/distribution/admin/commissions/${id}/settle`)
}

// 管理后台 - 批量结算佣金
// data: { ids: [1,2,3] }
export function batchSettleDistributionCommission(data) {
  return request.post('/distribution/admin/commissions/batch-settle', data)
}

// 管理后台 - 取消佣金
export function cancelDistributionCommission(id) {
  return request.put(`/distribution/admin/commissions/${id}/cancel`)
}

// C 端 - 佣金公开列表
export function getDistributionCommissionPublicList(params) {
  return request.get('/distribution/commissions', { params })
}

// C 端 - 佣金详情
export function getDistributionCommissionPublicDetail(id) {
  return request.get(`/distribution/commissions/${id}`)
}

// C 端 - 佣金汇总
export function getDistributionCommissionPublicSummary(params) {
  return request.get('/distribution/commissions/summary', { params })
}

// C 端 - 我的佣金
export function getMyDistributionCommissions(params) {
  return request.get('/distribution/commissions/mine', { params })
}

// ====== 合伙人等级（levels） ======

// 管理后台 - 等级列表
export function getDistributionLevelList(params) {
  return request.get('/distribution/admin/levels', { params })
}

// 管理后台 - 创建等级
// data: { level, name, required_amount, commission_rate, status, extra_benefits? }
export function createDistributionLevel(data) {
  return request.post('/distribution/admin/levels', data)
}

// 管理后台 - 更新等级
export function updateDistributionLevel(id, data) {
  return request.put(`/distribution/admin/levels/${id}`, data)
}

// 管理后台 - 删除等级
export function deleteDistributionLevel(id) {
  return request.delete(`/distribution/admin/levels/${id}`)
}

// 管理后台 - 检查自动升级
// data: { partner_id }
export function checkDistributionLevelUpgrade(data) {
  return request.post('/distribution/admin/levels/check-upgrade', data)
}

// C 端 - 等级列表（公开）
export function getDistributionLevelPublicList(params) {
  return request.get('/distribution/levels', { params })
}

// C 端 - 全部启用等级
export function getDistributionLevelAll() {
  return request.get('/distribution/levels/all')
}

// C 端 - 等级详情
export function getDistributionLevelPublicDetail(id) {
  return request.get(`/distribution/levels/${id}`)
}

// ====== 提现记录（withdrawals） ======

// 管理后台 - 提现列表
export function getDistributionWithdrawalList(params) {
  return request.get('/distribution/admin/withdrawals', { params })
}

// 管理后台 - 待审核提现列表
export function getDistributionWithdrawalPendingList(params) {
  return request.get('/distribution/admin/withdrawals/pending', { params })
}

// 管理后台 - 提现详情
export function getDistributionWithdrawalDetail(id) {
  return request.get(`/distribution/admin/withdrawals/${id}`)
}

// 管理后台 - 审核提现（通过/拒绝）
// data: { status: 1/3, reason }  1通过 3拒绝
export function auditDistributionWithdrawal(id, data) {
  return request.put(`/distribution/admin/withdrawals/${id}/audit`, data)
}

// 管理后台 - 打款确认
// data: { reason? }
export function payDistributionWithdrawal(id, data) {
  return request.put(`/distribution/admin/withdrawals/${id}/pay`, data)
}

// 管理后台 - 拒绝提现
// data: { reason }
export function rejectDistributionWithdrawal(id, data) {
  return request.put(`/distribution/admin/withdrawals/${id}/reject`, data)
}

// C 端 - 提现公开列表
export function getDistributionWithdrawalPublicList(params) {
  return request.get('/distribution/withdrawals', { params })
}

// C 端 - 提现详情
export function getDistributionWithdrawalPublicDetail(id) {
  return request.get(`/distribution/withdrawals/${id}`)
}

// C 端 - 申请提现
// data: { partner_id, amount, bank_info }
export function applyDistributionWithdrawal(data) {
  return request.post('/distribution/withdrawals/apply', data)
}

// C 端 - 我的提现
export function getMyDistributionWithdrawals(params) {
  return request.get('/distribution/withdrawals/mine', { params })
}
