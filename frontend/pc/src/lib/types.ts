// 后端统一响应格式：{ code, message, data }
// code=0 表示成功

export interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  pageSize: number
}

// 业务实体
export interface News {
  id: number
  title: string
  content: string
  cover_image: string
  summary: string
  author_id: number
  author_name: string
  category_id: number
  tags: string
  view_count: number
  like_count: number
  status: number
  published_at: string | null
  created_at: string
}

export interface Category {
  id: number
  name: string
  parent_id: number
  level: number
  sort: number
  icon?: string
  children?: Category[]
}

export interface Region {
  id: number
  name: string
  parent_id: number
  level: number
  sort: number
}

export interface LikeResponse {
  liked: boolean
  like_count: number
}

// ====== 二手交易 Ershou ======
export interface Ershou {
  id: number
  title: string
  description: string
  price: number
  original_price?: number
  cover_image?: string
  images?: string[]
  category_id?: number
  category_name?: string
  region_id?: number
  region_name?: string
  user_id: number
  user_name?: string
  user_avatar?: string
  contact?: string
  status: number // 1发布 3下架 4过期
  audit_status: number // 0待审 1通过 2拒绝
  view_count: number
  like_count: number
  location?: string
  created_at: string
}

// ====== 招聘求职 Job（后端模块待开发，先定义占位结构） ======
export interface Job {
  id: number
  title: string
  company: string
  salary: string
  location?: string
  description?: string
  category?: string
  status?: number
  created_at: string
}

// ====== 房屋租售 Fang（后端模块待开发，先定义占位结构） ======
export interface Fang {
  id: number
  title: string
  type: string // 出租/出售
  price: string
  area?: string
  layout?: string
  location?: string
  description?: string
  status?: number
  created_at: string
}
