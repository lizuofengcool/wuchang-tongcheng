// 用户相关 API
import { request } from './request'

// 登录（用户名/密码）
export function login(username, password) {
  return request({
    url: '/api/v1/user/login',
    method: 'POST',
    data: { username, password },
  })
}

// 获取个人信息（需登录）
export function getProfile() {
  return request({
    url: '/api/v1/user/info',
    method: 'GET',
    requireAuth: true,
  })
}

// 退出登录（仅清除本地 token）
export function logout() {
  uni.removeStorageSync('token')
  uni.removeStorageSync('userInfo')
}
