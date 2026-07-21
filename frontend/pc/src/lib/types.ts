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

// ====== 相亲交友 Love ======
export interface Love {
  id: number
  user_id: number
  nickname?: string
  avatar?: string
  gender: number // 1男 2女
  age?: number
  age_range?: string
  height?: number
  region_id?: number
  region_name?: string
  city?: string
  occupation?: string
  education?: string
  income?: string
  marriage?: string // 未婚/离异/丧偶
  house?: string // 住房情况
  car?: string // 购车情况
  bio?: string // 个人简介
  voice_intro?: string
  photos?: string[]
  tags?: string[]
  status: number // 1发布 3下架
  audit_status: number
  verified?: boolean // 实名认证
  featured?: boolean // 推荐
  view_count: number
  like_count: number
  location?: string
  created_at: string
}

// 相亲动态广场
export interface LoveStory {
  id: number
  love_id?: number
  user_id: number
  nickname?: string
  avatar?: string
  content: string
  images?: string[]
  topic?: string
  view_count: number
  like_count: number
  comment_count: number
  share_count: number
  status: number
  featured?: boolean
  created_at: string
}

// ====== 拼车出行 Pinche ======
export interface Pinche {
  id: number
  user_id: number
  user_name?: string
  user_avatar?: string
  type: number // 1车主找乘客 2乘客找车主
  start_city?: string
  start_location: string
  end_city?: string
  end_location: string
  via?: string // 途经
  departure_time: string // 出发时间
  return_time?: string // 返程时间（往返）
  price: number
  seats_total: number
  seats_left: number
  vehicle_model?: string
  vehicle_color?: string
  vehicle_plate?: string
  region_id?: number
  region_name?: string
  remark?: string
  status: number // 1发布 2已满 3已结束
  audit_status: number
  view_count: number
  contact_count: number
  created_at: string
}

// 拼车路线
export interface PincheRoute {
  id: number
  user_id?: number
  name: string
  start_location: string
  end_location: string
  via?: string
  distance?: number
  duration?: number
  use_count: number
  favorite_count: number
  created_at: string
}

// ====== 零工兼职 Linggong ======
export interface Linggong {
  id: number
  user_id: number
  employer_id?: number
  employer_name?: string
  employer_avatar?: string
  title: string
  description?: string
  category_id?: number
  category_name?: string
  type: number // 1零工 2兼职 3全职
  salary: string // 薪资描述：如 200元/天
  salary_min?: number
  salary_max?: number
  salary_unit?: string // 时薪/日薪/月薪/件薪
  region_id?: number
  region_name?: string
  address?: string
  location?: string
  work_time?: string // 工作时间
  work_duration?: string // 工作周期
  headcount: number // 招聘人数
  applied_count: number // 已报名
  requirements?: string[]
  benefits?: string[]
  contact?: string
  phone?: string
  status: number // 1招聘中 2已满 3已结束
  audit_status: number
  view_count: number
  created_at: string
}

// 零工任务
export interface LinggongTask {
  id: number
  linggong_id?: number
  worker_id?: number
  worker_name?: string
  title: string
  description?: string
  status: number // 1进行中 2已完成 3已取消
  start_time?: string
  end_time?: string
  amount?: number
  created_at: string
}

// ====== 同城114 Dh114 ======
export interface Dh114 {
  id: number
  user_id: number
  name: string // 商户名称
  cover_image?: string
  logo?: string
  description?: string
  category_id?: number
  category_name?: string
  business_type?: string
  region_id?: number
  region_name?: string
  address?: string
  location?: string
  phone?: string
  mobile?: string
  business_hours?: string
  longitude?: number
  latitude?: number
  rating?: number // 评分
  review_count?: number
  avg_price?: number // 人均消费
  status: number // 1营业 2休息 3关闭
  audit_status: number
  featured?: boolean
  view_count: number
  fav_count: number
  created_at: string
}

// 114 分类
export interface Dh114Category {
  id: number
  name: string
  parent_id: number
  level: number
  sort: number
  icon?: string
  business_type?: string
  children?: Dh114Category[]
}

// 114 团购
export interface Dh114Groupbuy {
  id: number
  dh114_id: number
  dh114_name?: string
  title: string
  cover_image?: string
  description?: string
  price: number
  original_price?: number
  sales_count: number
  status: number
  start_time?: string
  end_time?: string
  created_at: string
}

// 114 优惠券
export interface Dh114Coupon {
  id: number
  dh114_id: number
  dh114_name?: string
  title: string
  type: number // 1满减 2折扣 3代金
  amount?: number
  threshold?: number // 满减门槛
  discount?: number // 折扣率
  total_count: number
  received_count: number
  start_time?: string
  end_time?: string
  status: number
  created_at: string
}

// ====== 同城商城 Mall ======
export interface MallProduct {
  id: number
  shop_id: number
  shop_name?: string
  title: string
  subtitle?: string
  cover_image?: string
  images?: string[]
  description?: string
  category_id?: number
  category_name?: string
  price: number
  original_price?: number
  stock: number
  sales_count: number
  view_count: number
  rating?: number
  review_count?: number
  status: number // 1上架 2下架
  featured?: boolean
  created_at: string
}

export interface MallShop {
  id: number
  user_id: number
  name: string
  logo?: string
  banner?: string
  description?: string
  category_id?: number
  category_name?: string
  region_id?: number
  region_name?: string
  address?: string
  phone?: string
  rating?: number
  review_count?: number
  product_count?: number
  status: number
  featured?: boolean
  created_at: string
}

export interface MallCategory {
  id: number
  name: string
  parent_id: number
  level: number
  sort: number
  icon?: string
  children?: MallCategory[]
}

// ====== 团购 GroupBuy ======
export interface GroupBuy {
  id: number
  shop_id?: number
  shop_name?: string
  title: string
  subtitle?: string
  cover_image?: string
  images?: string[]
  description?: string
  price: number
  original_price?: number
  stock: number
  sales_count: number
  limit_per_user?: number
  start_time?: string
  end_time?: string
  status: number // 1上架 2下架
  audit_status: number
  created_at: string
}

// 团购优惠券
export interface GroupBuyCoupon {
  id: number
  groupbuy_id?: number
  shop_id?: number
  shop_name?: string
  title: string
  type: number // 1满减 2折扣 3代金
  amount?: number
  threshold?: number
  discount?: number
  total_count: number
  received_count: number
  start_time?: string
  end_time?: string
  status: number
  created_at: string
}
