// 系统设置模块 API 封装
import request from '@/utils/request'

// 获取全部配置
export function getAllSettings() {
  return request.get('/setting')
}

// 按 group 获取配置
export function getSettingsByGroup(group) {
  return request.get(`/setting/group/${group}`)
}

// 创建配置
export function createSetting(data) {
  return request.post('/setting', data)
}

// 更新配置
export function updateSetting(id, data) {
  return request.put(`/setting/${id}`, data)
}

// 删除配置
export function deleteSetting(id) {
  return request.delete(`/setting/${id}`)
}

// 批量更新配置
export function batchUpdateSettings(items) {
  return request.put('/setting/batch', { items })
}
