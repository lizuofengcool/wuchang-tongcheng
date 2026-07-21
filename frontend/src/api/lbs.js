// LBS地图中台模块 API 封装
// 对应后端 backend/internal/modules/lbs（路由前缀 /api/v1/lbs）
// 按模块分组：poi / region / geofence / 公共能力（distance/route）
import request from '@/utils/request'

// ====== POI 兴趣点（pois） ======

// POI 列表（公开）
export function getLbsPoiList(params) {
  return request.get('/lbs/pois', { params })
}

// 附近 POI 检索
export function getLbsPoiNearby(params) {
  return request.get('/lbs/pois/nearby', { params })
}

// POI 详情
export function getLbsPoiDetail(id) {
  return request.get(`/lbs/pois/${id}`)
}

// 我的 POI 列表
export function getMyLbsPois(params) {
  return request.get('/lbs/pois/mine', { params })
}

// 创建 POI
export function createLbsPoi(data) {
  return request.post('/lbs/pois', data)
}

// 更新 POI
export function updateLbsPoi(id, data) {
  return request.put(`/lbs/pois/${id}`, data)
}

// 删除 POI
export function deleteLbsPoi(id) {
  return request.delete(`/lbs/pois/${id}`)
}

// ====== 区域分站（regions） ======

// 区域列表
export function getLbsRegionList(params) {
  return request.get('/lbs/regions', { params })
}

// 按父级列出
export function getLbsRegionsByParent(parentId) {
  return request.get(`/lbs/regions/by-parent/${parentId}`)
}

// 按城市编码查找
export function getLbsRegionByCityCode(cityCode) {
  return request.get('/lbs/regions/by-city-code', { params: { city_code: cityCode } })
}

// 根据经纬度判断分站
export function getLbsRegionByLocation(latitude, longitude) {
  return request.get('/lbs/regions/by-location', { params: { latitude, longitude } })
}

// 区域详情
export function getLbsRegionDetail(id) {
  return request.get(`/lbs/regions/${id}`)
}

// 创建区域
export function createLbsRegion(data) {
  return request.post('/lbs/regions', data)
}

// 更新区域
export function updateLbsRegion(id, data) {
  return request.put(`/lbs/regions/${id}`, data)
}

// 删除区域
export function deleteLbsRegion(id) {
  return request.delete(`/lbs/regions/${id}`)
}

// ====== 地理围栏（geofences） ======

// 围栏列表
export function getLbsGeofenceList(params) {
  return request.get('/lbs/geofences', { params })
}

// 按区域列出
export function getLbsGeofencesByRegion(regionId) {
  return request.get(`/lbs/geofences/by-region/${regionId}`)
}

// 按所有者列出
export function getLbsGeofencesByOwner(ownerId, ownerType) {
  return request.get(`/lbs/geofences/by-owner/${ownerId}`, { params: { owner_type: ownerType } })
}

// 围栏详情
export function getLbsGeofenceDetail(id) {
  return request.get(`/lbs/geofences/${id}`)
}

// 检查点是否在指定围栏内
export function checkLbsPointInGeofence(id, data) {
  return request.post(`/lbs/geofences/${id}/check-point`, data)
}

// 检查点是否在区域内任一围栏内
export function checkLbsPointInRegion(regionId, data) {
  return request.post('/lbs/geofences/check-point-in-region', data, { params: { region_id: regionId } })
}

// 创建围栏
export function createLbsGeofence(data) {
  return request.post('/lbs/geofences', data)
}

// 更新围栏
export function updateLbsGeofence(id, data) {
  return request.put(`/lbs/geofences/${id}`, data)
}

// 删除围栏
export function deleteLbsGeofence(id) {
  return request.delete(`/lbs/geofences/${id}`)
}

// ====== 公共能力 ======

// 距离计算
export function calcLbsDistance(params) {
  return request.get('/lbs/distance', { params })
}

// 路线规划
export function planLbsRoute(params) {
  return request.get('/lbs/route', { params })
}

// ====== 管理后台 ======

// 管理后台 - POI 列表
export function getLbsAdminPoiList(params) {
  return request.get('/lbs/admin/pois', { params })
}

// 管理后台 - POI 分类列表
export function getLbsAdminPoiCategories() {
  return request.get('/lbs/admin/pois/categories')
}

// 管理后台 - POI 详情
export function getLbsAdminPoiDetail(id) {
  return request.get(`/lbs/admin/pois/${id}`)
}

// 管理后台 - 更新 POI 状态
export function updateLbsAdminPoiStatus(id, status) {
  return request.put(`/lbs/admin/pois/${id}/status`, { status })
}

// 管理后台 - 删除 POI
export function deleteLbsAdminPoi(id) {
  return request.delete(`/lbs/admin/pois/${id}`)
}

// 管理后台 - 区域列表
export function getLbsAdminRegionList(params) {
  return request.get('/lbs/admin/regions', { params })
}

// 管理后台 - 更新区域状态
export function updateLbsAdminRegionStatus(id, status) {
  return request.put(`/lbs/admin/regions/${id}/status`, { status })
}

// 管理后台 - 围栏列表
export function getLbsAdminGeofenceList(params) {
  return request.get('/lbs/admin/geofences', { params })
}

// 管理后台 - 更新围栏状态
export function updateLbsAdminGeofenceStatus(id, status) {
  return request.put(`/lbs/admin/geofences/${id}/status`, { status })
}
