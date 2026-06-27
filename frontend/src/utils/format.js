// 共享格式化工具：消除各 view 内联重复实现
// 使用：import { formatTime, formatSize, statusText, statusTagType } from '@/utils/format'

const pad = (n) => String(n).padStart(2, '0')

/**
 * 格式化时间戳/日期字符串为 YYYY-MM-DD HH:mm:ss
 * @param {number|string|Date} t
 * @param {string} fmt 默认 'YYYY-MM-DD HH:mm:ss'
 */
export function formatTime(t, fmt = 'YYYY-MM-DD HH:mm:ss') {
  if (!t) return '-'
  const d = t instanceof Date ? t : new Date(t)
  if (isNaN(d.getTime())) return '-'
  return fmt
    .replace('YYYY', d.getFullYear())
    .replace('MM', pad(d.getMonth() + 1))
    .replace('DD', pad(d.getDate()))
    .replace('HH', pad(d.getHours()))
    .replace('mm', pad(d.getMinutes()))
    .replace('ss', pad(d.getSeconds()))
}

/** 仅日期 YYYY-MM-DD */
export function formatDate(t) {
  return formatTime(t, 'YYYY-MM-DD')
}

/**
 * 格式化文件大小
 * @param {number} b 字节数
 */
export function formatSize(b) {
  if (!b) return '0 B'
  if (b < 1024) return b + ' B'
  if (b < 1024 * 1024) return (b / 1024).toFixed(1) + ' KB'
  if (b < 1024 * 1024 * 1024) return (b / 1024 / 1024).toFixed(1) + ' MB'
  return (b / 1024 / 1024 / 1024).toFixed(2) + ' GB'
}

/**
 * 通用启用/禁用状态文本
 * @param {number} status 1=启用 0=禁用
 */
export function statusText(status) {
  return status === 1 ? '启用' : '禁用'
}

/**
 * 通用启用/禁用状态 el-tag 类型
 */
export function statusTagType(status) {
  return status === 1 ? 'success' : 'info'
}

/**
 * 资讯状态文本（草稿/已发布/已下架）
 */
export function newsStatusText(status) {
  return { 0: '草稿', 1: '已发布', 2: '已下架' }[status] || '-'
}

/**
 * 资讯状态 el-tag 类型
 */
export function newsStatusTagType(status) {
  return { 0: 'info', 1: 'success', 2: 'warning' }[status] || 'info'
}

export default {
  formatTime,
  formatDate,
  formatSize,
  statusText,
  statusTagType,
  newsStatusText,
  newsStatusTagType
}
