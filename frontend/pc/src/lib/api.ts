// 后端 API 封装
// 所有请求走 /api/v1/<module>（开发环境由 next.config.mjs rewrites 代理到后端）
// 公共门户站：无 JWT，仅浏览用户调用的接口（list/get/search/like-status）

import type { ApiResponse, PageResult, News, Category, Region, LikeResponse, Ershou, Job, Fang, Love, LoveStory, Pinche, PincheRoute, Linggong, LinggongTask, Dh114, Dh114Category, Dh114Groupbuy, Dh114Coupon, MallProduct, MallShop, MallCategory, GroupBuy, GroupBuyCoupon } from './types'

const BASE = '/api/v1'

// 服务端 fetch（带绝对 URL，用于 SSR）vs 客户端 fetch（相对路径）
function buildUrl(path: string): string {
  // 在服务端渲染时需要绝对 URL
  if (typeof window === 'undefined') {
    const backend = process.env.BACKEND_URL || 'http://localhost:8088'
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

// ====== Ershou 二手交易 ======

export async function listErshous(params: {
  regionId?: number
  page?: number
  pageSize?: number
  categoryId?: number
  keyword?: string
}): Promise<PageResult<Ershou>> {
  const q = new URLSearchParams()
  if (params.regionId) q.set('region_id', String(params.regionId))
  q.set('page', String(params.page || 1))
  q.set('page_size', String(params.pageSize || 12))
  if (params.categoryId) q.set('category_id', String(params.categoryId))
  if (params.keyword) q.set('keyword', params.keyword)
  return get<PageResult<Ershou>>(`/ershou?${q.toString()}`)
}

export async function getErshou(id: number): Promise<Ershou> {
  return get<Ershou>(`/ershou/${id}`, 0)
}

// ====== Job 招聘求职（后端模块待开发，调用失败时返回空列表） ======

export async function listJobs(params: {
  regionId?: number
  page?: number
  pageSize?: number
  keyword?: string
}): Promise<PageResult<Job>> {
  const q = new URLSearchParams()
  if (params.regionId) q.set('region_id', String(params.regionId))
  q.set('page', String(params.page || 1))
  q.set('page_size', String(params.pageSize || 12))
  if (params.keyword) q.set('keyword', params.keyword)
  return get<PageResult<Job>>(`/job?${q.toString()}`)
}

// ====== Fang 房屋租售（后端模块待开发，调用失败时返回空列表） ======

export async function listFangs(params: {
  regionId?: number
  page?: number
  pageSize?: number
  keyword?: string
}): Promise<PageResult<Fang>> {
  const q = new URLSearchParams()
  if (params.regionId) q.set('region_id', String(params.regionId))
  q.set('page', String(params.page || 1))
  q.set('page_size', String(params.pageSize || 12))
  if (params.keyword) q.set('keyword', params.keyword)
  return get<PageResult<Fang>>(`/fang?${q.toString()}`)
}

// ====== Love 相亲交友 ======
// 后端路由：GET /love（列表）/ /love/:id（详情）/ /love/stories（动态广场）
// 注意：固定路径 /stories 必须在 /:id 之前调用，后端已处理顺序

export async function listLoves(params: {
  regionId?: number
  page?: number
  pageSize?: number
  gender?: number // 1男 2女
  ageRange?: string
  keyword?: string
}): Promise<PageResult<Love>> {
  const q = new URLSearchParams()
  if (params.regionId) q.set('region_id', String(params.regionId))
  q.set('page', String(params.page || 1))
  q.set('page_size', String(params.pageSize || 12))
  if (params.gender) q.set('gender', String(params.gender))
  if (params.ageRange) q.set('age_range', params.ageRange)
  if (params.keyword) q.set('keyword', params.keyword)
  return get<PageResult<Love>>(`/love?${q.toString()}`)
}

export async function getLoveDetail(id: number): Promise<Love> {
  return get<Love>(`/love/${id}`, 0) // 详情不缓存
}

export async function listLoveStories(params: {
  page?: number
  pageSize?: number
  topic?: string
}): Promise<PageResult<LoveStory>> {
  const q = new URLSearchParams()
  q.set('page', String(params.page || 1))
  q.set('page_size', String(params.pageSize || 10))
  if (params.topic) q.set('topic', params.topic)
  return get<PageResult<LoveStory>>(`/love/stories?${q.toString()}`)
}

// ====== Pinche 拼车出行 ======
// 后端路由：GET /pinche（列表）/ /pinche/:id（详情）/ /pinche/routes（路线）

export async function listPinches(params: {
  regionId?: number
  page?: number
  pageSize?: number
  startCity?: string
  endCity?: string
  date?: string // 出发日期 YYYY-MM-DD
  keyword?: string
}): Promise<PageResult<Pinche>> {
  const q = new URLSearchParams()
  if (params.regionId) q.set('region_id', String(params.regionId))
  q.set('page', String(params.page || 1))
  q.set('page_size', String(params.pageSize || 12))
  if (params.startCity) q.set('start_city', params.startCity)
  if (params.endCity) q.set('end_city', params.endCity)
  if (params.date) q.set('date', params.date)
  if (params.keyword) q.set('keyword', params.keyword)
  return get<PageResult<Pinche>>(`/pinche?${q.toString()}`)
}

export async function getPincheDetail(id: number): Promise<Pinche> {
  return get<Pinche>(`/pinche/${id}`, 0)
}

export async function listPincheRoutes(params: {
  page?: number
  pageSize?: number
}): Promise<PageResult<PincheRoute>> {
  const q = new URLSearchParams()
  q.set('page', String(params.page || 1))
  q.set('page_size', String(params.pageSize || 20))
  return get<PageResult<PincheRoute>>(`/pinche/routes?${q.toString()}`)
}

// ====== Linggong 零工兼职 ======
// 后端路由：GET /linggong（列表）/ /linggong/:id（详情）/ /linggong/tasks（任务列表）

export async function listLinggongs(params: {
  regionId?: number
  page?: number
  pageSize?: number
  categoryId?: number
  type?: number // 1零工 2兼职 3全职
  keyword?: string
}): Promise<PageResult<Linggong>> {
  const q = new URLSearchParams()
  if (params.regionId) q.set('region_id', String(params.regionId))
  q.set('page', String(params.page || 1))
  q.set('page_size', String(params.pageSize || 12))
  if (params.categoryId) q.set('category_id', String(params.categoryId))
  if (params.type) q.set('type', String(params.type))
  if (params.keyword) q.set('keyword', params.keyword)
  return get<PageResult<Linggong>>(`/linggong?${q.toString()}`)
}

export async function getLinggongDetail(id: number): Promise<Linggong> {
  return get<Linggong>(`/linggong/${id}`, 0)
}

export async function listLinggongTasks(params: {
  page?: number
  pageSize?: number
  linggongId?: number
}): Promise<PageResult<LinggongTask>> {
  const q = new URLSearchParams()
  q.set('page', String(params.page || 1))
  q.set('page_size', String(params.pageSize || 20))
  if (params.linggongId) q.set('linggong_id', String(params.linggongId))
  return get<PageResult<LinggongTask>>(`/linggong/tasks?${q.toString()}`)
}

// ====== Dh114 同城114 ======
// 后端路由：GET /dh114（商户列表）/ /dh114/:id（详情）/ /dh114/categories / /dh114/groupbuys / /dh114/coupons

export async function listDh114s(params: {
  regionId?: number
  page?: number
  pageSize?: number
  categoryId?: number
  keyword?: string
}): Promise<PageResult<Dh114>> {
  const q = new URLSearchParams()
  if (params.regionId) q.set('region_id', String(params.regionId))
  q.set('page', String(params.page || 1))
  q.set('page_size', String(params.pageSize || 12))
  if (params.categoryId) q.set('category_id', String(params.categoryId))
  if (params.keyword) q.set('keyword', params.keyword)
  return get<PageResult<Dh114>>(`/dh114?${q.toString()}`)
}

export async function getDh114Detail(id: number): Promise<Dh114> {
  return get<Dh114>(`/dh114/${id}`, 0)
}

export async function listDh114Categories(): Promise<Dh114Category[]> {
  return get<Dh114Category[]>(`/dh114/categories`, 600)
}

export async function listDh114Groupbuys(params?: {
  dh114Id?: number
  page?: number
  pageSize?: number
}): Promise<PageResult<Dh114Groupbuy>> {
  const q = new URLSearchParams()
  q.set('page', String(params?.page || 1))
  q.set('page_size', String(params?.pageSize || 12))
  if (params?.dh114Id) q.set('dh114_id', String(params.dh114Id))
  return get<PageResult<Dh114Groupbuy>>(`/dh114/groupbuys?${q.toString()}`)
}

export async function listDh114Coupons(params?: {
  dh114Id?: number
  page?: number
  pageSize?: number
}): Promise<PageResult<Dh114Coupon>> {
  const q = new URLSearchParams()
  q.set('page', String(params?.page || 1))
  q.set('page_size', String(params?.pageSize || 12))
  if (params?.dh114Id) q.set('dh114_id', String(params.dh114Id))
  return get<PageResult<Dh114Coupon>>(`/dh114/coupons?${q.toString()}`)
}

// ====== Mall 同城商城 ======
// 后端路由：GET /mall/products[/search|/:id] / /mall/shops[/:id] / /mall/categories

export async function listMallProducts(params: {
  regionId?: number
  page?: number
  pageSize?: number
  categoryId?: number
  shopId?: number
  keyword?: string
}): Promise<PageResult<MallProduct>> {
  const q = new URLSearchParams()
  if (params.regionId) q.set('region_id', String(params.regionId))
  q.set('page', String(params.page || 1))
  q.set('page_size', String(params.pageSize || 12))
  if (params.categoryId) q.set('category_id', String(params.categoryId))
  if (params.shopId) q.set('shop_id', String(params.shopId))
  if (params.keyword) q.set('keyword', params.keyword)
  return get<PageResult<MallProduct>>(`/mall/products?${q.toString()}`)
}

export async function searchMallProducts(params: {
  page?: number
  pageSize?: number
  keyword: string
  categoryId?: number
}): Promise<PageResult<MallProduct>> {
  const q = new URLSearchParams()
  q.set('page', String(params.page || 1))
  q.set('page_size', String(params.pageSize || 12))
  q.set('keyword', params.keyword)
  if (params.categoryId) q.set('category_id', String(params.categoryId))
  return get<PageResult<MallProduct>>(`/mall/products/search?${q.toString()}`)
}

export async function getMallProductDetail(id: number): Promise<MallProduct> {
  return get<MallProduct>(`/mall/products/${id}`, 0)
}

export async function listMallShops(params: {
  regionId?: number
  page?: number
  pageSize?: number
  categoryId?: number
  keyword?: string
}): Promise<PageResult<MallShop>> {
  const q = new URLSearchParams()
  if (params.regionId) q.set('region_id', String(params.regionId))
  q.set('page', String(params.page || 1))
  q.set('page_size', String(params.pageSize || 12))
  if (params.categoryId) q.set('category_id', String(params.categoryId))
  if (params.keyword) q.set('keyword', params.keyword)
  return get<PageResult<MallShop>>(`/mall/shops?${q.toString()}`)
}

export async function getMallShopDetail(id: number): Promise<MallShop> {
  return get<MallShop>(`/mall/shops/${id}`, 0)
}

export async function listMallCategories(): Promise<MallCategory[]> {
  return get<MallCategory[]>(`/mall/categories`, 600)
}

// ====== GroupBuy 团购 ======
// 后端路由：GET /groupbuy/list（列表）/ /groupbuy/:id（详情）/ /groupbuy/coupons（优惠券）

export async function listGroupBuys(params: {
  page?: number
  pageSize?: number
  shopId?: number
  keyword?: string
}): Promise<PageResult<GroupBuy>> {
  const q = new URLSearchParams()
  q.set('page', String(params.page || 1))
  q.set('page_size', String(params.pageSize || 12))
  if (params.shopId) q.set('shop_id', String(params.shopId))
  if (params.keyword) q.set('keyword', params.keyword)
  return get<PageResult<GroupBuy>>(`/groupbuy/list?${q.toString()}`)
}

export async function getGroupBuyDetail(id: number): Promise<GroupBuy> {
  return get<GroupBuy>(`/groupbuy/${id}`, 0)
}

export async function listGroupBuyCoupons(params?: {
  groupbuyId?: number
  page?: number
  pageSize?: number
}): Promise<PageResult<GroupBuyCoupon>> {
  const q = new URLSearchParams()
  q.set('page', String(params?.page || 1))
  q.set('page_size', String(params?.pageSize || 12))
  if (params?.groupbuyId) q.set('groupbuy_id', String(params.groupbuyId))
  return get<PageResult<GroupBuyCoupon>>(`/groupbuy/coupons?${q.toString()}`)
}
