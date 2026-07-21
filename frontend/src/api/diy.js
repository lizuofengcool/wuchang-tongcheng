// DIY 前端页面中台模块 API 封装
// 对应后端 backend/internal/modules/diy（路由前缀 /api/v1/diy）
// 按子域分组：page / component / template / stat
import request from '@/utils/request'

// ====== 页面（pages） ======

// 管理后台 - 页面列表（全部页面）
export function getDiyPageList(params) {
  return request.get('/diy/admin/pages', { params })
}

// 管理后台 - 页面详情
export function getDiyPageDetail(id) {
  return request.get(`/diy/admin/pages/${id}`)
}

// 管理后台 - 更新页面状态（强制下线/恢复）
export function updateDiyPageStatus(id, data) {
  return request.put(`/diy/admin/pages/${id}/status`, data)
}

// C 端 - 按 slug 获取已发布页面
export function getDiyPageBySlug(slug) {
  return request.get(`/diy/pages/by-slug/${slug}`)
}

// C 端 - 我的页面列表
export function getMyDiyPages(params) {
  return request.get('/diy/pages/mine', { params })
}

// C 端 - 创建页面
export function createDiyPage(data) {
  return request.post('/diy/pages', data)
}

// C 端 - 更新页面
export function updateDiyPage(id, data) {
  return request.put(`/diy/pages/${id}`, data)
}

// C 端 - 删除页面
export function deleteDiyPage(id) {
  return request.delete(`/diy/pages/${id}`)
}

// C 端 - 发布页面
export function publishDiyPage(id) {
  return request.post(`/diy/pages/${id}/publish`)
}

// C 端 - 下线页面
export function offlineDiyPage(id) {
  return request.post(`/diy/pages/${id}/offline`)
}

// C 端 - 复制页面
export function copyDiyPage(id, data) {
  return request.post(`/diy/pages/${id}/copy`, data)
}

// ====== 组件库（components） ======

// 管理后台 - 组件列表
export function getDiyComponentList(params) {
  return request.get('/diy/admin/components', { params })
}

// 管理后台 - 组件详情
export function getDiyComponentDetail(id) {
  return request.get(`/diy/admin/components/${id}`)
}

// 管理后台 - 创建组件
export function createDiyComponent(data) {
  return request.post('/diy/admin/components', data)
}

// 管理后台 - 更新组件
export function updateDiyComponent(id, data) {
  return request.put(`/diy/admin/components/${id}`, data)
}

// 管理后台 - 删除组件
export function deleteDiyComponent(id) {
  return request.delete(`/diy/admin/components/${id}`)
}

// C 端 - 按分类获取组件
export function getDiyComponentsByCategory(category) {
  return request.get(`/diy/components/by-category/${category}`)
}

// ====== 模板（templates） ======

// 管理后台 - 模板列表
export function getDiyTemplateList(params) {
  return request.get('/diy/admin/templates', { params })
}

// 管理后台 - 模板详情
export function getDiyTemplateDetail(id) {
  return request.get(`/diy/admin/templates/${id}`)
}

// 管理后台 - 创建模板
export function createDiyTemplate(data) {
  return request.post('/diy/admin/templates', data)
}

// 管理后台 - 更新模板
export function updateDiyTemplate(id, data) {
  return request.put(`/diy/admin/templates/${id}`, data)
}

// 管理后台 - 删除模板
export function deleteDiyTemplate(id) {
  return request.delete(`/diy/admin/templates/${id}`)
}

// C 端 - 应用模板创建新页面
export function applyDiyTemplate(id, data) {
  return request.post(`/diy/templates/${id}/apply`, data)
}

// ====== 统计（stats） ======

// 管理后台 - 按页面 ID 列出统计
export function getDiyStatListByPage(id, params) {
  return request.get(`/diy/admin/stats/page/${id}`, { params })
}

// 管理后台 - 按日期范围列出统计
export function getDiyStatListByDateRange(params) {
  return request.get('/diy/admin/stats/date-range', { params })
}

// 管理后台 - 按页面 ID 汇总统计
export function getDiyStatSummaryByPage(id) {
  return request.get(`/diy/admin/stats/summary/page/${id}`)
}

// 管理后台 - 按日期范围汇总统计
export function getDiyStatSummary(params) {
  return request.get('/diy/admin/stats/summary', { params })
}

// C 端 - 记录浏览
export function recordDiyView(data) {
  return request.post('/diy/stats/view', data)
}

// C 端 - 记录点击
export function recordDiyClick(data) {
  return request.post('/diy/stats/click', data)
}

// C 端 - 记录转化
export function recordDiyConversion(data) {
  return request.post('/diy/stats/conversion', data)
}
