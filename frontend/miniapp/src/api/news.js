// 头条相关 API
import { request } from './request'

// 头条列表（带分页 + 分类 + 关键词过滤）
export function listNews(params = {}) {
  return request({
    url: '/api/v1/news',
    method: 'GET',
    data: {
      page: params.page || 1,
      page_size: params.pageSize || 10,
      category_id: params.categoryId,
      keyword: params.keyword,
      status: 1, // 仅展示已发布
    },
  })
}

// 头条全文搜索（走后端 /news/search，ES 优先 DB 降级）
export function searchNews(params) {
  return request({
    url: '/api/v1/news/search',
    method: 'GET',
    data: {
      keyword: params.keyword,
      page: params.page || 1,
      page_size: params.pageSize || 10,
      category_id: params.categoryId,
    },
  })
}

// 头条详情
export function getNewsDetail(id) {
  return request({
    url: `/api/v1/news/${id}`,
    method: 'GET',
  })
}

// 点赞状态查询（需登录）
export function getLikeStatus(newsId) {
  return request({
    url: `/api/v1/news/${newsId}/like`,
    method: 'GET',
    requireAuth: true,
  })
}

// 点赞/取消点赞（需登录）
export function toggleLike(newsId) {
  return request({
    url: `/api/v1/news/${newsId}/like`,
    method: 'POST',
    requireAuth: true,
  })
}

// 分类列表
export function listCategories() {
  return request({
    url: '/api/v1/category',
    method: 'GET',
  })
}

// 地区列表
export function listRegions() {
  return request({
    url: '/api/v1/region',
    method: 'GET',
  })
}
