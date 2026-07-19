// 模块总控 API 封装
// 对应后端路由前缀：/api/v1/modules（P0 阶段 Agent 1 提供）
import request from '@/utils/request'

// 获取模块列表
export function getModules(params) {
  return request({ url: '/modules', method: 'get', params })
}

// 获取模块详情
export function getModuleDetail(name) {
  return request({ url: `/modules/${name}`, method: 'get' })
}

// 启用模块
export function enableModule(name) {
  return request({ url: `/modules/${name}/enable`, method: 'post' })
}

// 禁用模块
export function disableModule(name) {
  return request({ url: `/modules/${name}/disable`, method: 'post' })
}

// 更新模块元信息
export function updateModule(name, data) {
  return request({ url: `/modules/${name}`, method: 'put', data })
}
