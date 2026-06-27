// 分类信息模块 API 封装
import request from '@/utils/request'

// 分类树形结构
export function getCategoryTree() {
  return request.get('/category/tree')
}

// 根据父级ID获取子分类
export function getCategoryChildren(parentId) {
  return request.get('/category/children', { params: { parent_id: parentId } })
}

// 分类详情
export function getCategory(id) {
  return request.get(`/category/${id}`)
}

// 创建分类
export function createCategory(data) {
  return request.post('/category', data)
}

// 更新分类
export function updateCategory(id, data) {
  return request.put(`/category/${id}`, data)
}

// 删除分类
export function deleteCategory(id) {
  return request.delete(`/category/${id}`)
}
