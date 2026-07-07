// 单元测试：src/utils/format.js 纯函数格式化工具
import { describe, it, expect } from 'vitest'
import {
  formatTime,
  formatDate,
  formatSize,
  statusText,
  statusTagType,
  newsStatusText,
  newsStatusTagType
} from '../format'

describe('formatTime', () => {
  it('空值返回占位符 "-"', () => {
    expect(formatTime(0)).toBe('-')
    expect(formatTime('')).toBe('-')
    expect(formatTime(null)).toBe('-')
    expect(formatTime(undefined)).toBe('-')
  })

  it('非法日期返回 "-"', () => {
    expect(formatTime('not-a-date')).toBe('-')
    expect(formatTime('2024-13-99')).toBe('-')
  })

  it('Date 对象按默认格式 YYYY-MM-DD HH:mm:ss 输出', () => {
    const d = new Date(2024, 0, 15, 10, 30, 45) // 本地时区 2024-01-15 10:30:45
    expect(formatTime(d)).toBe('2024-01-15 10:30:45')
  })

  it('时间戳（毫秒）正确格式化', () => {
    const d = new Date(2024, 5, 1, 8, 5, 9) // 2024-06-01 08:05:09
    expect(formatTime(d.getTime())).toBe('2024-06-01 08:05:09')
  })

  it('支持自定义格式 YYYY-MM-DD', () => {
    const d = new Date(2024, 11, 31, 23, 59, 59)
    expect(formatTime(d, 'YYYY-MM-DD')).toBe('2024-12-31')
  })

  it('月份/日期/时/分/秒均补零', () => {
    const d = new Date(2024, 0, 5, 3, 7, 1) // 2024-01-05 03:07:01
    expect(formatTime(d)).toBe('2024-01-05 03:07:01')
  })
})

describe('formatDate', () => {
  it('仅输出日期 YYYY-MM-DD', () => {
    const d = new Date(2024, 2, 9, 12, 0, 0)
    expect(formatDate(d)).toBe('2024-03-09')
  })

  it('空值返回 "-"', () => {
    expect(formatDate(null)).toBe('-')
  })
})

describe('formatSize', () => {
  it('0 字节返回 "0 B"', () => {
    expect(formatSize(0)).toBe('0 B')
  })

  it('小于 1024 输出 B', () => {
    expect(formatSize(1)).toBe('1 B')
    expect(formatSize(512)).toBe('512 B')
    expect(formatSize(1023)).toBe('1023 B')
  })

  it('1024 边界输出 KB（1.0 KB）', () => {
    expect(formatSize(1024)).toBe('1.0 KB')
  })

  it('KB 量级保留一位小数', () => {
    expect(formatSize(1536)).toBe('1.5 KB') // 1.5 KB
  })

  it('1 MiB 输出 MB', () => {
    expect(formatSize(1024 * 1024)).toBe('1.0 MB')
  })

  it('1 GiB 输出 GB（两位小数）', () => {
    expect(formatSize(1024 * 1024 * 1024)).toBe('1.00 GB')
  })
})

describe('statusText / statusTagType', () => {
  it('status=1 启用', () => {
    expect(statusText(1)).toBe('启用')
    expect(statusTagType(1)).toBe('success')
  })

  it('status=0 禁用', () => {
    expect(statusText(0)).toBe('禁用')
    expect(statusTagType(0)).toBe('info')
  })

  it('非 1 状态统一视为禁用', () => {
    expect(statusText(2)).toBe('禁用')
    expect(statusTagType(99)).toBe('info')
  })
})

describe('newsStatusText / newsStatusTagType', () => {
  it('0=草稿 info', () => {
    expect(newsStatusText(0)).toBe('草稿')
    expect(newsStatusTagType(0)).toBe('info')
  })

  it('1=已发布 success', () => {
    expect(newsStatusText(1)).toBe('已发布')
    expect(newsStatusTagType(1)).toBe('success')
  })

  it('2=已下架 warning', () => {
    expect(newsStatusText(2)).toBe('已下架')
    expect(newsStatusTagType(2)).toBe('warning')
  })

  it('未知状态返回 "-" / "info"', () => {
    expect(newsStatusText(99)).toBe('-')
    expect(newsStatusTagType(99)).toBe('info')
  })
})
