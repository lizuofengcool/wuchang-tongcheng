// 用户模块 API 封装
import request from '@/utils/request'

// 用户注册
export function register(data) {
  return request.post('/user/register', data)
}

// 用户登录
export function login(data) {
  return request.post('/user/login', data)
}

// 获取当前用户信息
export function getUserInfo() {
  return request.get('/user/info')
}

// 更新个人资料
export function updateProfile(data) {
  return request.put('/user/profile', data)
}

// 修改密码
export function changePassword(data) {
  return request.put('/user/password', data)
}

// ===== 管理后台 =====

// 用户列表（分页 + 关键词 + 状态筛选）
export function listUsers(params) {
  return request.get('/user/admin/users', { params })
}

// 获取指定用户
export function getUser(id) {
  return request.get(`/user/admin/users/${id}`)
}

// 管理员创建用户
export function adminCreateUser(data) {
  return request.post('/user/admin/users', data)
}

// 管理员更新用户资料
export function adminUpdateUser(id, data) {
  return request.put(`/user/admin/users/${id}`, data)
}

// 更新用户状态
export function updateUserStatus(id, status) {
  return request.put(`/user/admin/users/${id}/status`, { status })
}

// 重置用户密码
export function resetUserPassword(id, new_password) {
  return request.put(`/user/admin/users/${id}/password`, { new_password })
}

// 删除用户
export function deleteUser(id) {
  return request.delete(`/user/admin/users/${id}`)
}

