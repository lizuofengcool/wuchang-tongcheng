// 单元测试：src/stores/user.js
// 覆盖 state 初始化（localStorage 回填）、getters（isLoggedIn/nickname/avatar/isSuperAdmin）、
// login / loginBySms / fetchProfile / fetchAuth / logout 全 action 行为
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'

// mock api/user：login / loginBySms / getUserInfo
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

describe('userStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    loginBySmsMock.mockReset()
    loginMock.mockReset()
    getUserInfoMock.mockReset()
    myAuthMock.mockReset()
  })

  describe('state 初始化', () => {
    it('localStorage 无值时回退空 token / null userInfo / 空权限数组', () => {
      const store = useUserStore()
      expect(store.token).toBe('')
      expect(store.userInfo).toBeNull()
      expect(store.permissions).toEqual([])
      expect(store.roles).toEqual([])
    })

    it('localStorage 存在 token / userInfo / 权限码时回填 state', () => {
      localStorage.setItem('token', 'persisted-token')
      localStorage.setItem('userInfo', JSON.stringify({ id: 9, username: 'persist' }))
      localStorage.setItem('permissions', JSON.stringify(['news:read', 'news:write']))
      localStorage.setItem('roles', JSON.stringify(['editor']))
      const store = useUserStore()
      expect(store.token).toBe('persisted-token')
      expect(store.userInfo).toEqual({ id: 9, username: 'persist' })
      expect(store.permissions).toEqual(['news:read', 'news:write'])
      expect(store.roles).toEqual(['editor'])
    })

    it('localStorage 中 userInfo 为空字符串时回退 null（|| "null" fallback）', () => {
      // 实现使用 JSON.parse(localStorage.getItem('userInfo') || 'null')
      // 空串走 || 短路 → JSON.parse('null') → null
      localStorage.setItem('userInfo', '')
      const store = useUserStore()
      expect(store.userInfo).toBeNull()
    })

    it('localStorage 中 permissions 为空字符串时回退空数组', () => {
      localStorage.setItem('permissions', '')
      const store = useUserStore()
      expect(store.permissions).toEqual([])
    })

    it('localStorage 中 roles 为空字符串时回退空数组', () => {
      localStorage.setItem('roles', '')
      const store = useUserStore()
      expect(store.roles).toEqual([])
    })
  })

  describe('getters', () => {
    it('isLoggedIn 仅在 token 非空时为真', () => {
      const store = useUserStore()
      expect(store.isLoggedIn).toBe(false)
      store.token = 'x'
      expect(store.isLoggedIn).toBe(true)
    })

    it('nickname 优先取 userInfo.nickname，缺失则回退 username，再缺失回退“管理员”', () => {
      const store = useUserStore()
      expect(store.nickname).toBe('管理员')
      store.userInfo = { username: 'alice' }
      expect(store.nickname).toBe('alice')
      store.userInfo = { username: 'alice', nickname: '小红' }
      expect(store.nickname).toBe('小红')
    })

    it('avatar 取 userInfo.avatar，缺失回退空串', () => {
      const store = useUserStore()
      expect(store.avatar).toBe('')
      store.userInfo = { avatar: 'https://cdn/x.png' }
      expect(store.avatar).toBe('https://cdn/x.png')
    })

    it('isSuperAdmin 仅当 roles 含 “admin” 时为真', () => {
      const store = useUserStore()
      expect(store.isSuperAdmin).toBe(false)
      store.roles = ['editor']
      expect(store.isSuperAdmin).toBe(false)
      store.roles = ['editor', 'admin']
      expect(store.isSuperAdmin).toBe(true)
    })
  })

  describe('login action（密码登录）', () => {
    it('成功 → token/userInfo 入 store 与 localStorage，并触发权限拉取', async () => {
      const token = 'jwt-pwd-token'
      const userInfo = { id: 1, username: 'admin', nickname: '超管' }
      loginMock.mockResolvedValueOnce({ code: 0, data: { token, expires: 3600, user_info: userInfo } })
      myAuthMock.mockResolvedValueOnce({ code: 0, data: { permissions: ['*'], roles: ['admin'] } })

      const store = useUserStore()
      const res = await store.login({ username: 'admin', password: '123456' })

      expect(loginMock).toHaveBeenCalledWith({ username: 'admin', password: '123456' })
      // 返回原始响应
      expect(res.data.token).toBe(token)
      // store 状态
      expect(store.token).toBe(token)
      expect(store.userInfo).toEqual(userInfo)
      expect(store.isLoggedIn).toBe(true)
      // localStorage 持久化
      expect(localStorage.getItem('token')).toBe(token)
      expect(JSON.parse(localStorage.getItem('userInfo'))).toEqual(userInfo)
      // 权限拉取
      expect(myAuthMock).toHaveBeenCalledTimes(1)
      expect(store.permissions).toEqual(['*'])
      expect(store.roles).toEqual(['admin'])
      expect(store.isSuperAdmin).toBe(true)
    })

    it('API 失败 → 抛出错误且不写入 token', async () => {
      loginMock.mockRejectedValueOnce(new Error('用户名或密码错误'))
      const store = useUserStore()
      await expect(store.login({ username: 'x', password: 'y' })).rejects.toThrow('用户名或密码错误')
      expect(store.token).toBe('')
      expect(store.isLoggedIn).toBe(false)
      expect(localStorage.getItem('token')).toBeNull()
      // 失败不应触发权限拉取
      expect(myAuthMock).not.toHaveBeenCalled()
    })

    it('登录成功但权限拉取失败 → 不阻断登录，权限为空', async () => {
      const token = 'jwt-pwd-token-2'
      const userInfo = { id: 2, username: 'bob' }
      loginMock.mockResolvedValueOnce({ code: 0, data: { token, expires: 3600, user_info: userInfo } })
      myAuthMock.mockRejectedValueOnce(new Error('network'))

      const store = useUserStore()
      await store.login({ username: 'bob', password: 'pwd' })
      expect(store.token).toBe(token)
      expect(store.isLoggedIn).toBe(true)
      expect(store.permissions).toEqual([])
      expect(store.roles).toEqual([])
    })
  })

  describe('loginBySms action', () => {
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

  describe('fetchProfile action', () => {
    it('成功 → 更新 userInfo 并持久化到 localStorage', async () => {
      const fresh = { id: 1, username: 'admin', nickname: '新昵称', avatar: 'https://cdn/a.png' }
      getUserInfoMock.mockResolvedValueOnce({ code: 0, data: fresh })
      const store = useUserStore()
      store.userInfo = { id: 1, username: 'admin', nickname: '旧昵称' }

      await store.fetchProfile()

      expect(getUserInfoMock).toHaveBeenCalledTimes(1)
      expect(store.userInfo).toEqual(fresh)
      expect(JSON.parse(localStorage.getItem('userInfo'))).toEqual(fresh)
      // 不应触碰 token / 权限
      expect(localStorage.getItem('token')).toBeNull()
    })

    it('失败 → 抛错且不修改既有 userInfo', async () => {
      getUserInfoMock.mockRejectedValueOnce(new Error('unauthorized'))
      const store = useUserStore()
      const existing = { id: 1, username: 'admin' }
      store.userInfo = existing
      await expect(store.fetchProfile()).rejects.toThrow('unauthorized')
      expect(store.userInfo).toEqual(existing)
    })
  })

  describe('fetchAuth action', () => {
    it('成功 → permissions/roles 入 store 并持久化', async () => {
      myAuthMock.mockResolvedValueOnce({
        code: 0,
        data: { permissions: ['news:read', 'news:write'], roles: ['editor', 'admin'] }
      })
      const store = useUserStore()
      await store.fetchAuth()
      expect(store.permissions).toEqual(['news:read', 'news:write'])
      expect(store.roles).toEqual(['editor', 'admin'])
      expect(JSON.parse(localStorage.getItem('permissions'))).toEqual(['news:read', 'news:write'])
      expect(JSON.parse(localStorage.getItem('roles'))).toEqual(['editor', 'admin'])
      expect(store.isSuperAdmin).toBe(true)
    })

    it('data.permissions 缺失 → 兜底为空数组', async () => {
      myAuthMock.mockResolvedValueOnce({ code: 0, data: { roles: ['editor'] } })
      const store = useUserStore()
      await store.fetchAuth()
      expect(store.permissions).toEqual([])
      expect(store.roles).toEqual(['editor'])
    })

    it('data.roles 缺失 → 兜底为空数组', async () => {
      myAuthMock.mockResolvedValueOnce({ code: 0, data: { permissions: ['news:read'] } })
      const store = useUserStore()
      await store.fetchAuth()
      expect(store.roles).toEqual([])
      expect(store.permissions).toEqual(['news:read'])
    })

    it('data 整体缺失 → permissions/roles 均兜底为空数组', async () => {
      myAuthMock.mockResolvedValueOnce({ code: 0 })
      const store = useUserStore()
      await store.fetchAuth()
      expect(store.permissions).toEqual([])
      expect(store.roles).toEqual([])
    })

    it('请求异常 → 静默失败不抛错，权限保持原值', async () => {
      myAuthMock.mockRejectedValueOnce(new Error('network'))
      const store = useUserStore()
      store.permissions = ['existing:perm']
      store.roles = ['existing:role']
      await expect(store.fetchAuth()).resolves.toBeUndefined()
      expect(store.permissions).toEqual(['existing:perm'])
      expect(store.roles).toEqual(['existing:role'])
    })
  })

  describe('logout action', () => {
    it('清空 store 状态并移除 localStorage 全部用户相关键', () => {
      // 预置状态
      localStorage.setItem('token', 't')
      localStorage.setItem('userInfo', JSON.stringify({ id: 1 }))
      localStorage.setItem('permissions', JSON.stringify(['x']))
      localStorage.setItem('roles', JSON.stringify(['admin']))
      // 同时塞一个无关键，验证不会被误删
      localStorage.setItem('currentRegionId', '5')

      const store = useUserStore()
      // 触发 state 初始化回填
      expect(store.token).toBe('t')
      store.logout()

      expect(store.token).toBe('')
      expect(store.userInfo).toBeNull()
      expect(store.permissions).toEqual([])
      expect(store.roles).toEqual([])
      expect(store.isLoggedIn).toBe(false)
      expect(store.isSuperAdmin).toBe(false)

      expect(localStorage.getItem('token')).toBeNull()
      expect(localStorage.getItem('userInfo')).toBeNull()
      expect(localStorage.getItem('permissions')).toBeNull()
      expect(localStorage.getItem('roles')).toBeNull()
      // 无关键应保留
      expect(localStorage.getItem('currentRegionId')).toBe('5')
    })

    it('重复调用 logout 不抛错', () => {
      const store = useUserStore()
      store.logout()
      expect(() => store.logout()).not.toThrow()
    })
  })
})
