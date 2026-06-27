// 同城头条模块 API 封装
import request from '@/utils/request'

// 头条列表（分页 + 分类/状态/关键词筛选）
export function listNews(params) {
  return request.get('/news', { params })
}

// 头条详情
export function getNews(id) {
  return request.get(`/news/${id}`)
}

// 发布头条
export function createNews(data) {
  return request.post('/news', data)
}

// 更新头条
export function updateNews(id, data) {
  return request.put(`/news/${id}`, data)
}

// 删除头条
export function deleteNews(id) {
  return request.delete(`/news/${id}`)
}
