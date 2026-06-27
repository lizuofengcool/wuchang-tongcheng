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
