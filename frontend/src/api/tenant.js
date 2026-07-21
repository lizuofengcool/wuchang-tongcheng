// 多租户分站中台 API 封装
// 对应后端 backend/internal/modules/tenant（路由前缀 /api/v1/tenant）
// 按模块分组：station / staff / config / domain
import request from '@/utils/request'

// ====== 分站（stations） ======

// 根据域名获取当前分站（公开）
export function getCurrentTenant(domain) {
  return request.get('/tenant/stations/current', { params: { domain } })
}

// 分站列表（需登录）
export function getTenantStationList(params) {
  return request.get('/tenant/stations', { params })
}

// 管理后台 - 分站列表
export function getTenantStationAdminList(params) {
  return request.get('/tenant/admin/stations', { params })
}

// 管理后台 - 分站详情
export function getTenantStationDetail(id) {
  return request.get(`/tenant/admin/stations/${id}`)
}

// 管理后台 - 创建分站
export function createTenantStation(data) {
  return request.post('/tenant/admin/stations', data)
}

// 管理后台 - 更新分站
export function updateTenantStation(id, data) {
  return request.put(`/tenant/admin/stations/${id}`, data)
}

// 管理后台 - 删除分站
export function deleteTenantStation(id) {
  return request.delete(`/tenant/admin/stations/${id}`)
}

// 管理后台 - 启停分站
export function updateTenantStationStatus(id, status) {
  return request.put(`/tenant/admin/stations/${id}/status`, { status })
}

// 管理后台 - 配置复制（从源分站复制到目标分站）
export function copyTenantStationConfig(data) {
  return request.post('/tenant/admin/stations/copy-config', data)
}

// ====== 员工（staff） ======

// 管理后台 - 员工列表
export function getTenantStaffList(params) {
  return request.get('/tenant/admin/staff', { params })
}

// 管理后台 - 员工详情
export function getTenantStaffDetail(id) {
  return request.get(`/tenant/admin/staff/${id}`)
}

// 管理后台 - 添加员工
export function createTenantStaff(data) {
  return request.post('/tenant/admin/staff', data)
}

// 管理后台 - 更新员工（角色/权限/状态）
export function updateTenantStaff(id, data) {
  return request.put(`/tenant/admin/staff/${id}`, data)
}

// 管理后台 - 删除员工
export function deleteTenantStaff(id) {
  return request.delete(`/tenant/admin/staff/${id}`)
}

// 管理后台 - 按分站查询员工
export function getTenantStaffByStation(stationId) {
  return request.get(`/tenant/admin/staff/by-station/${stationId}`)
}

// ====== 配置（configs） ======

// 管理后台 - 配置列表
export function getTenantConfigList(params) {
  return request.get('/tenant/admin/configs', { params })
}

// 管理后台 - 配置详情
export function getTenantConfigDetail(id) {
  return request.get(`/tenant/admin/configs/${id}`)
}

// 管理后台 - 新增/更新配置（Upsert）
export function upsertTenantConfig(data) {
  return request.post('/tenant/admin/configs', data)
}

// 管理后台 - 更新配置值
export function updateTenantConfig(id, data) {
  return request.put(`/tenant/admin/configs/${id}`, data)
}

// 管理后台 - 删除配置
export function deleteTenantConfig(id) {
  return request.delete(`/tenant/admin/configs/${id}`)
}

// 管理后台 - 按分站+模块查询配置
export function getTenantConfigByStationModule(stationId, bizModule) {
  return request.get('/tenant/admin/configs/by-station-module', { params: { station_id: stationId, biz_module: bizModule } })
}

// 管理后台 - 批量获取配置
export function batchGetTenantConfig(data) {
  return request.post('/tenant/admin/configs/batch-get', data)
}

// ====== 域名（domains） ======

// 管理后台 - 域名列表
export function getTenantDomainList(params) {
  return request.get('/tenant/admin/domains', { params })
}

// 管理后台 - 域名详情
export function getTenantDomainDetail(id) {
  return request.get(`/tenant/admin/domains/${id}`)
}

// 管理后台 - 绑定域名
export function createTenantDomain(data) {
  return request.post('/tenant/admin/domains', data)
}

// 管理后台 - 更新域名（SSL 状态）
export function updateTenantDomain(id, data) {
  return request.put(`/tenant/admin/domains/${id}`, data)
}

// 管理后台 - 删除域名绑定
export function deleteTenantDomain(id) {
  return request.delete(`/tenant/admin/domains/${id}`)
}

// 管理后台 - 设置主域名
export function setTenantDomainPrimary(id) {
  return request.put(`/tenant/admin/domains/${id}/primary`)
}

// 管理后台 - 更新 SSL 状态
export function updateTenantDomainSSL(id, sslStatus) {
  return request.put(`/tenant/admin/domains/${id}/ssl`, { ssl_status: sslStatus })
}

// 管理后台 - 按分站查询域名
export function getTenantDomainByStation(stationId) {
  return request.get(`/tenant/admin/domains/by-station/${stationId}`)
}
