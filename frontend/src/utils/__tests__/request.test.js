// 单元测试：src/utils/request.js axios 封装
// 验证请求拦截器（JWT token / X-Region-ID 注入）、响应拦截器（业务码 / HTTP 错误码处理）、未授权去重
import { describe, it, expect, beforeEach, vi } from 'vitest'

// mock element-plus：ElMessage.error / ElMessageBox.confirm
// 用惰性 wrapper 包裹，避免 vi.mock 提升导致的 TDZ（与 stores/__tests__ 同风格）
const elMessageError = vi.fn()
const elMessageBoxConfirm = vi.fn()
vi.mock('element-plus', () => ({
  ElMessage: { error: (...a) => elMessageError(...a) },
  ElMessageBox: { confirm: (...a) => elMessageBoxConfirm(...a) }
}))

// mock @/router：currentRoute.value.path 可控（用于 /500 去重判断）+ push 可观测
const routerPush = vi.fn()
const currentRouteValue = { path: '/dashboard' }
vi.mock('@/router', () => ({
  default: {
    push: (...a) => routerPush(...a),
    currentRoute: { value: currentRouteValue }
  }
}))

// mock @/stores/user：useUserStore().logout() 可观测（未授权时清 store）
const userLogout = vi.fn()
vi.mock('@/stores/user', () => ({
  useUserStore: () => ({ logout: (...a) => userLogout(...a) })
}))

// 每次重新动态 import 获取全新的 request 模块（重置模块级 unauthorizedShown 闭包标志 + axios 实例）
// vi.resetModules 后 vi.mock 工厂仍生效，惰性 wrapper 仍指向同一组 vi.fn
async function loadRequest() {
  vi.resetModules()
  const mod = await import('../request')
  const service = mod.default
  const reqHandler = service.interceptors.request.handlers[0]
  const resHandler = service.interceptors.response.handlers[0]
  return { service, reqHandler, resHandler }
}

beforeEach(() => {
  vi.resetModules()
  localStorage.clear()
  elMessageError.mockReset()
  elMessageBoxConfirm.mockReset()
  // 默认返回永不 resolve 的 promise，保持 unauthorizedShown=true 以便测试去重逻辑
  elMessageBoxConfirm.mockImplementation(() => new Promise(() => {}))
  routerPush.mockReset()
  userLogout.mockReset()
  currentRouteValue.path = '/dashboard'
})

describe('request 拦截器 - 请求阶段（JWT token / 地区头注入）', () => {
  it('localStorage 有 token → 注入 Authorization: Bearer <token>', async () => {
    localStorage.setItem('token', 'jwt-abc-123')
    const { reqHandler } = await loadRequest()
    const config = { headers: {} }
    const result = await reqHandler.fulfilled(config)
    expect(result.headers['Authorization']).toBe('Bearer jwt-abc-123')
  })

  it('localStorage 无 token → 不注入 Authorization', async () => {
    const { reqHandler } = await loadRequest()
    const config = { headers: {} }
    const result = await reqHandler.fulfilled(config)
    expect(result.headers['Authorization']).toBeUndefined()
  })

  it('localStorage 有 currentRegionId → 注入 X-Region-ID', async () => {
    localStorage.setItem('currentRegionId', '5')
    const { reqHandler } = await loadRequest()
    const config = { headers: {} }
    const result = await reqHandler.fulfilled(config)
    expect(result.headers['X-Region-ID']).toBe('5')
  })

  it('localStorage 无 currentRegionId → 不注入 X-Region-ID', async () => {
    const { reqHandler } = await loadRequest()
    const config = { headers: {} }
    const result = await reqHandler.fulfilled(config)
    expect(result.headers['X-Region-ID']).toBeUndefined()
  })

  it('token 与 region 同时存在 → 两个头都注入', async () => {
    localStorage.setItem('token', 'T')
    localStorage.setItem('currentRegionId', '3')
    const { reqHandler } = await loadRequest()
    const result = await reqHandler.fulfilled({ headers: {} })
    expect(result.headers['Authorization']).toBe('Bearer T')
    expect(result.headers['X-Region-ID']).toBe('3')
  })

  it('请求拦截器 rejected 直接透传错误', async () => {
    const { reqHandler } = await loadRequest()
    const err = new Error('req setup fail')
    await expect(reqHandler.rejected(err)).rejects.toBe(err)
  })
})

describe('request 拦截器 - 响应阶段（业务码处理）', () => {
  it('code === 0 → 原样返回 res 业务体', async () => {
    const { resHandler } = await loadRequest()
    const res = { code: 0, message: 'ok', data: { id: 1 } }
    const result = await resHandler.fulfilled({ data: res })
    expect(result).toEqual(res)
  })

  it('非 0 且非鉴权码 → ElMessage.error 提示并 reject', async () => {
    const { resHandler } = await loadRequest()
    const res = { code: 1001, message: '参数错误' }
    await expect(resHandler.fulfilled({ data: res })).rejects.toThrow('参数错误')
    expect(elMessageError).toHaveBeenCalledWith('参数错误')
    // 不触发未授权弹窗与登出
    expect(elMessageBoxConfirm).not.toHaveBeenCalled()
    expect(userLogout).not.toHaveBeenCalled()
  })

  it('code === 401 → 触发未授权流程（弹窗 + 登出 + reject）', async () => {
    const { resHandler } = await loadRequest()
    const res = { code: 401, message: 'token失效' }
    await expect(resHandler.fulfilled({ data: res })).rejects.toThrow('token失效')
    expect(elMessageBoxConfirm).toHaveBeenCalledTimes(1)
    expect(userLogout).toHaveBeenCalledTimes(1)
  })

  it('code === 2006（token 失效业务码）→ 触发未授权流程', async () => {
    const { resHandler } = await loadRequest()
    await expect(
      resHandler.fulfilled({ data: { code: 2006, message: '登录已失效' } })
    ).rejects.toThrow('登录已失效')
    expect(elMessageBoxConfirm).toHaveBeenCalledTimes(1)
    expect(userLogout).toHaveBeenCalledTimes(1)
  })

  it('未授权去重：连续两次 401 业务码仅弹一次窗、登出一次', async () => {
    const { resHandler } = await loadRequest()
    const res = { code: 401, message: '失效' }
    await expect(resHandler.fulfilled({ data: res })).rejects.toThrow()
    // 第二次：unauthorizedShown 已为 true，handleUnauthorized 提前返回
    await expect(resHandler.fulfilled({ data: res })).rejects.toThrow()
    expect(elMessageBoxConfirm).toHaveBeenCalledTimes(1)
    expect(userLogout).toHaveBeenCalledTimes(1)
  })
})

describe('request 拦截器 - HTTP 错误码处理（rejected）', () => {
  it('HTTP 401 → 触发未授权流程并 reject', async () => {
    const { resHandler } = await loadRequest()
    const error = { response: { status: 401 }, message: 'Request failed with status code 401' }
    await expect(resHandler.rejected(error)).rejects.toBe(error)
    expect(elMessageBoxConfirm).toHaveBeenCalledTimes(1)
    expect(userLogout).toHaveBeenCalledTimes(1)
  })

  it('HTTP 403 → ElMessage.error("禁止访问")', async () => {
    const { resHandler } = await loadRequest()
    const error = { response: { status: 403 }, message: 'forbidden' }
    await expect(resHandler.rejected(error)).rejects.toBe(error)
    expect(elMessageError).toHaveBeenCalledWith('禁止访问')
    expect(elMessageBoxConfirm).not.toHaveBeenCalled()
  })

  it('HTTP 500（当前不在 /500）→ 提示并跳转 /500', async () => {
    const { resHandler } = await loadRequest()
    currentRouteValue.path = '/dashboard'
    const error = { response: { status: 500 }, message: 'server error' }
    await expect(resHandler.rejected(error)).rejects.toBe(error)
    expect(elMessageError).toHaveBeenCalledWith('服务器内部错误')
    expect(routerPush).toHaveBeenCalledWith('/500')
  })

  it('HTTP 500（已在 /500）→ 提示但不重复跳转', async () => {
    const { resHandler } = await loadRequest()
    currentRouteValue.path = '/500'
    const error = { response: { status: 500 }, message: 'server error' }
    await expect(resHandler.rejected(error)).rejects.toBe(error)
    expect(elMessageError).toHaveBeenCalledWith('服务器内部错误')
    expect(routerPush).not.toHaveBeenCalled()
  })

  it('其他 HTTP 状态码 → 提示 "请求错误 (status)"', async () => {
    const { resHandler } = await loadRequest()
    const error = { response: { status: 502 }, message: 'bad gateway' }
    await expect(resHandler.rejected(error)).rejects.toBe(error)
    expect(elMessageError).toHaveBeenCalledWith('请求错误 (502)')
  })

  it('超时错误（message 含 timeout）→ 提示请求超时', async () => {
    const { resHandler } = await loadRequest()
    const error = { message: 'timeout of 15000ms exceeded' }
    await expect(resHandler.rejected(error)).rejects.toBe(error)
    expect(elMessageError).toHaveBeenCalledWith('请求超时，请稍后重试')
  })

  it('网络错误（无 response）→ 提示网络异常', async () => {
    const { resHandler } = await loadRequest()
    const error = { message: 'Network Error' }
    await expect(resHandler.rejected(error)).rejects.toBe(error)
    expect(elMessageError).toHaveBeenCalledWith('网络异常，请检查网络连接')
  })
})
