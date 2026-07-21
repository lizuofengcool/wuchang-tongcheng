// 营销中台模块 API 封装
// 对应后端 backend/internal/modules/marketing（路由前缀 /api/v1/marketing）
// 按子域分组：ad / coupon / signRule / activity
import request from '@/utils/request'

// ====== 广告位（ads） ======

// 管理后台 - 广告位列表
export function getMarketingAdList(params) {
  return request.get('/marketing/admin/ads', { params })
}

// 管理后台 - 广告位详情
export function getMarketingAdDetail(id) {
  return request.get(`/marketing/admin/ads/${id}`)
}

// 管理后台 - 创建广告位
export function createMarketingAd(data) {
  return request.post('/marketing/admin/ads', data)
}

// 管理后台 - 更新广告位
export function updateMarketingAd(id, data) {
  return request.put(`/marketing/admin/ads/${id}`, data)
}

// 管理后台 - 删除广告位
export function deleteMarketingAd(id) {
  return request.delete(`/marketing/admin/ads/${id}`)
}

// C 端 - 按位置编码获取广告
export function getMarketingAdsByPosition(code, params) {
  return request.get(`/marketing/positions/${code}/ads`, { params })
}

// ====== 优惠券（coupons） ======

// 管理后台 - 优惠券列表
export function getMarketingCouponList(params) {
  return request.get('/marketing/admin/coupons', { params })
}

// 管理后台 - 优惠券详情
export function getMarketingCouponDetail(id) {
  return request.get(`/marketing/admin/coupons/${id}`)
}

// 管理后台 - 创建优惠券
export function createMarketingCoupon(data) {
  return request.post('/marketing/admin/coupons', data)
}

// 管理后台 - 更新优惠券
export function updateMarketingCoupon(id, data) {
  return request.put(`/marketing/admin/coupons/${id}`, data)
}

// 管理后台 - 删除优惠券
export function deleteMarketingCoupon(id) {
  return request.delete(`/marketing/admin/coupons/${id}`)
}

// 管理后台 - 优惠券统计
export function getMarketingCouponStatistics() {
  return request.get('/marketing/admin/coupons/statistics')
}

// C 端 - 可领取优惠券列表
export function getMarketingAvailableCoupons(params) {
  return request.get('/marketing/coupons/available', { params })
}

// C 端 - 优惠券详情
export function getMarketingCouponPublicDetail(id) {
  return request.get(`/marketing/coupons/${id}`)
}

// C 端 - 领取优惠券
export function receiveMarketingCoupon(id, data) {
  return request.post(`/marketing/coupons/${id}/receive`, data)
}

// C 端 - 使用优惠券
export function useMarketingUserCoupon(id, data) {
  return request.post(`/marketing/user-coupons/${id}/use`, data)
}

// C 端 - 退还优惠券
export function refundMarketingUserCoupon(id) {
  return request.post(`/marketing/user-coupons/${id}/refund`)
}

// C 端 - 我的优惠券
export function getMyMarketingCoupons(params) {
  return request.get('/marketing/my-coupons', { params })
}

// ====== 签到规则（sign-rules） ======

// 管理后台 - 签到规则列表
export function getMarketingSignRuleList(params) {
  return request.get('/marketing/admin/sign-rules', { params })
}

// 管理后台 - 签到规则详情
export function getMarketingSignRuleDetail(id) {
  return request.get(`/marketing/admin/sign-rules/${id}`)
}

// 管理后台 - 创建签到规则
export function createMarketingSignRule(data) {
  return request.post('/marketing/admin/sign-rules', data)
}

// 管理后台 - 更新签到规则
export function updateMarketingSignRule(id, data) {
  return request.put(`/marketing/admin/sign-rules/${id}`, data)
}

// 管理后台 - 删除签到规则
export function deleteMarketingSignRule(id) {
  return request.delete(`/marketing/admin/sign-rules/${id}`)
}

// C 端 - 每日签到
export function marketingCheckIn() {
  return request.post('/marketing/sign/check-in')
}

// C 端 - 签到日历
export function getMarketingSignCalendar(params) {
  return request.get('/marketing/sign/calendar', { params })
}

// C 端 - 启用的签到规则
export function getMarketingEnabledSignRules() {
  return request.get('/marketing/sign/rules/enabled')
}

// ====== 营销活动（activities） ======

// 管理后台 - 活动列表
export function getMarketingActivityList(params) {
  return request.get('/marketing/admin/activities', { params })
}

// 管理后台 - 活动详情
export function getMarketingActivityDetail(id) {
  return request.get(`/marketing/admin/activities/${id}`)
}

// 管理后台 - 创建活动
export function createMarketingActivity(data) {
  return request.post('/marketing/admin/activities', data)
}

// 管理后台 - 更新活动
export function updateMarketingActivity(id, data) {
  return request.put(`/marketing/admin/activities/${id}`, data)
}

// 管理后台 - 删除活动
export function deleteMarketingActivity(id) {
  return request.delete(`/marketing/admin/activities/${id}`)
}

// 管理后台 - 更新活动状态
export function updateMarketingActivityStatus(id, data) {
  return request.put(`/marketing/admin/activities/${id}/status`, data)
}

// 管理后台 - 活动统计
export function getMarketingActivityStatistics() {
  return request.get('/marketing/admin/activities/statistics')
}

// C 端 - 进行中的活动
export function getMarketingOngoingActivities(params) {
  return request.get('/marketing/activities/ongoing', { params })
}

// C 端 - 即将开始的活动
export function getMarketingUpcomingActivities(params) {
  return request.get('/marketing/activities/upcoming', { params })
}

// C 端 - 已结束的活动
export function getMarketingEndedActivities(params) {
  return request.get('/marketing/activities/ended', { params })
}

// C 端 - 活动详情
export function getMarketingActivityPublicDetail(id) {
  return request.get(`/marketing/activities/${id}`)
}
