// 后端 API 封装
// 所有请求走 /api/v1/<module>（开发环境由 next.config.mjs rewrites 代理到后端）
// 公共门户站：无 JWT，仅浏览用户调用的接口（list/get/search/like-status）

import type { ApiResponse, PageResult, News, Category, Region, LikeResponse } from './types'

const BASE = '/api/v1'

// 服务端 fetch（带绝对 URL，用于 SSR）vs 客户端 fetch（相对路径）
function buildUrl(path: string): string {
  // 在服务端渲染时需要绝对 URL
  if (typeof window === 'undefined') {
    const backend = process.env.BACKEND_URL || 'http://localhost:8080'
    return `${backend}${BASE}${path}`
  }
  // 客户端走 rewrites 代理
  return `${BASE}${path}`
}

async function get<T>(path: string, revalidate: number = 60): Promise<T> {
  const res = await fetch(buildUrl(path), {
    next: { revalidate }, // ISR：默认 60 秒缓存
    headers: { 'Content-Type': 'application/json' },
  })
  if (!res.ok) {
    throw new Error(`API ${res.status}: ${res.statusText}`)
  }
  const json: ApiResponse<T> = await res.json()
  if (json.code !== 0) {
    throw new Error(`API code=${json.code}: ${json.message}`)
  }
  return json.data
}

// ====== News ======

export async function listNews(params: {
  regionId?: number
  page?: number
  pageSize?: number
  categoryId?: number
  keyword?: string
}): Promise<PageResult<News>> {
  const q = new URLSearchParams()
  if (params.regionId) q.set('region_id', String(params.regionId))
  q.set('page', String(params.page || 1))
  q.set('page_size', String(params.pageSize || 10))
  if (params.categoryId) q.set('category_id', String(params.categoryId))
  q.set('status', '1') // 仅展示已发布
  if (params.keyword) q.set('keyword', params.keyword)
  return get<PageResult<News>>(`/news?${q.toString()}`)
}

export async function searchNews(params: {
  regionId?: number
  page?: number
  pageSize?: number
  keyword: string
  categoryId?: number
}): Promise<PageResult<News>> {
  const q = new URLSearchParams()
  if (params.regionId) q.set('region_id', String(params.regionId))
  q.set('page', String(params.page || 1))
  q.set('page_size', String(params.pageSize || 10))
  q.set('keyword', params.keyword)
  if (params.categoryId) q.set('category_id', String(params.categoryId))
  return get<PageResult<News>>(`/news/search?${q.toString()}`)
}

export async function getNews(id: number): Promise<News> {
  return get<News>(`/news/${id}`, 0) // 详情不缓存
}

export async function getNewsLikeStatus(newsId: number, token?: string): Promise<LikeResponse> {
  // 点赞状态需登录，PC门户默认未登录，返回未点赞状态
  // 如已登录，传 Authorization: Bearer <token>
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  if (token) headers.Authorization = `Bearer ${token}`
  const res = await fetch(buildUrl(`/news/${newsId}/like`), { headers, cache: 'no-store' })
  if (!res.ok) return { liked: false, like_count: 0 }
  const json: ApiResponse<LikeResponse> = await res.json()
  return json.data || { liked: false, like_count: 0 }
}

export async function toggleNewsLike(newsId: number, token: string): Promise<LikeResponse> {
  const res = await fetch(buildUrl(`/news/${newsId}/like`), {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
  })
  if (!res.ok) throw new Error(`点赞失败: ${res.status}`)
  const json: ApiResponse<LikeResponse> = await res.json()
  if (json.code !== 0) throw new Error(json.message)
  return json.data
}

// ====== Category ======

export async function listCategories(): Promise<Category[]> {
  return get<Category[]>(`/category`, 600) // 分类变更少，缓存 10 分钟
}

// ====== Region ======

export async function listRegions(): Promise<Region[]> {
  return get<Region[]>(`/region`, 600)
}
