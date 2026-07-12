// 单元测试：src/stores/user.js 的 loginBySms action
// 验证短信验证码登录后 token/userInfo 落 store + localStorage，并触发权限拉取
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'

// mock api/user：loginBySms 返回与密码登录同构的 { token, expires, user_info }
const loginBySmsMock = vi.fn()
const loginMock = vi.fn()
const getUserInfoMock = vi.fn()
vi.mock('@/api/user', () => ({
  login: (...args) => loginMock(...args),
  loginBySms: (...args) => loginBySmsMock(...args),
  getUserInfo: (...args) => getUserInfoMock(...args)
}))

// mock api/permission：myAuth 返回权限码与角色码
const myAuthMock = vi.fn()
vi.mock('@/api/permission', () => ({
  myAuth: (...args) => myAuthMock(...args)
}))

import { useUserStore } from '../user'

describe('userStore.loginBySms', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    loginBySmsMock.mockReset()
    loginMock.mockReset()
    myAuthMock.mockReset()
  })

  it('短信登录成功 → token/userInfo 入 store 与 localStorage，并拉取权限', async () => {
    const token = 'sms-jwt-token-xxx'
    const userInfo = { id: 7, username: '13800138000', nickname: '短信用户' }
    loginBySmsMock.mockResolvedValueOnce({ code: 0, data: { token, expires: 3600, user_info: userInfo } })
    myAuthMock.mockResolvedValueOnce({ code: 0, data: { permissions: ['news:read'], roles: ['editor'] } })

    const store = useUserStore()
    await store.loginBySms({ phone: '13800138000', code: '123456' })

    // API 被以正确参数调用
    expect(loginBySmsMock).toHaveBeenCalledWith({ phone: '13800138000', code: '123456' })
    // store 状态
    expect(store.token).toBe(token)
    expect(store.userInfo).toEqual(userInfo)
    expect(store.isLoggedIn).toBe(true)
    // localStorage 持久化
    expect(localStorage.getItem('token')).toBe(token)
    expect(JSON.parse(localStorage.getItem('userInfo'))).toEqual(userInfo)
    // 权限拉取
    expect(myAuthMock).toHaveBeenCalledTimes(1)
    expect(store.permissions).toEqual(['news:read'])
    expect(store.roles).toEqual(['editor'])
  })

  it('短信登录 API 失败 → 抛出错误且不写入 token', async () => {
    loginBySmsMock.mockRejectedValueOnce(new Error('验证码错误或已过期'))
    const store = useUserStore()
    await expect(store.loginBySms({ phone: '13800138000', code: 'wrong' })).rejects.toThrow('验证码错误或已过期')
    expect(store.token).toBe('')
    expect(store.isLoggedIn).toBe(false)
    expect(localStorage.getItem('token')).toBeNull()
  })

  it('权限拉取失败不应阻断登录流程（静默降级）', async () => {
    const token = 'sms-jwt-token-yyy'
    const userInfo = { id: 8, username: '13900139000' }
    loginBySmsMock.mockResolvedValueOnce({ code: 0, data: { token, expires: 3600, user_info: userInfo } })
    myAuthMock.mockRejectedValueOnce(new Error('network'))

    const store = useUserStore()
    await store.loginBySms({ phone: '13900139000', code: '654321' })

    // 登录主流程仍成功
    expect(store.token).toBe(token)
    expect(store.isLoggedIn).toBe(true)
    // 权限为空但不抛错
    expect(store.permissions).toEqual([])
    expect(store.roles).toEqual([])
  })
})
