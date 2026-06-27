// 地区模块 API 封装
import request from '@/utils/request'

// 地区树形结构
export function getRegionTree() {
  return request.get('/region/tree')
}

// 根据父级ID获取子地区
export function getRegionChildren(parentId) {
  return request.get('/region/children', { params: { parent_id: parentId } })
}

// 地区详情
export function getRegion(id) {
  return request.get(`/region/${id}`)
}

// 创建地区
export function createRegion(data) {
  return request.post('/region', data)
}

// 更新地区
export function updateRegion(id, data) {
  return request.put(`/region/${id}`, data)
}

// 删除地区
export function deleteRegion(id) {
  return request.delete(`/region/${id}`)
}
