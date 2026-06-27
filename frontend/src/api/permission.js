// 权限模块 API 封装
import request from '@/utils/request'

// ===== 角色 =====
export function listRoles() {
  return request.get('/permission/roles')
}

export function getRole(id) {
  return request.get(`/permission/roles/${id}`)
}

export function createRole(data) {
  return request.post('/permission/roles', data)
}

export function updateRole(id, data) {
  return request.put(`/permission/roles/${id}`, data)
}

export function deleteRole(id) {
  return request.delete(`/permission/roles/${id}`)
}

// ===== 权限 =====
export function listPermissions() {
  return request.get('/permission/permissions')
}

export function createPermission(data) {
  return request.post('/permission/permissions', data)
}

export function updatePermission(id, data) {
  return request.put(`/permission/permissions/${id}`, data)
}

export function deletePermission(id) {
  return request.delete(`/permission/permissions/${id}`)
}

// 查询角色已分配的权限（用于回显）
export function getRolePermissions(roleId) {
  return request.get(`/permission/roles/${roleId}/permissions`)
}

// ===== 关联分配 =====
export function assignRoles(data) {
  return request.post('/permission/assign-roles', data)
}

export function assignPermissions(data) {
  return request.post('/permission/assign-permissions', data)
}

// 查询用户拥有的角色
export function getUserRoles(userId) {
  return request.get(`/permission/users/${userId}/roles`)
}

// 当前用户拥有的权限
export function myPermissions() {
  return request.get('/permission/my-permissions')
}

// 当前用户授权概览（权限码 + 角色码）
export function myAuth() {
  return request.get('/permission/my-auth')
}
