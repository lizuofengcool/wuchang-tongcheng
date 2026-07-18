// 单元测试：src/api/user.js
// 验证每个 API 函数调用 request 的方法（get/post/put/delete）、URL、入参（body / params / 路径插值）
import { describe, it, expect, beforeEach, vi } from 'vitest'

// mock @/utils/request：记录 method/url/config，返回固定 promise 便于断言
const requestMock = vi.fn()
vi.mock('@/utils/request', () => ({
  default: {
    get: (...a) => requestMock('get', ...a),
    post: (...a) => requestMock('post', ...a),
    put: (...a) => requestMock('put', ...a),
    delete: (...a) => requestMock('delete', ...a)
  }
}))

import * as userApi from '../user'

beforeEach(() => {
  requestMock.mockReset()
  requestMock.mockImplementation(() => Promise.resolve({ code: 0, data: {} }))
})

describe('user API - 公开鉴权接口', () => {
  it('register → POST /user/register，body 透传', async () => {
    const data = { username: 'u', password: 'p', phone: '13800000000' }
    await userApi.register(data)
    expect(requestMock).toHaveBeenCalledWith('post', '/user/register', data)
  })

  it('login → POST /user/login，body 透传', async () => {
    const data = { username: 'u', password: 'p' }
    await userApi.login(data)
    expect(requestMock).toHaveBeenCalledWith('post', '/user/login', data)
  })

  it('sendSmsCode → POST /user/sms/code，body 仅含 phone', async () => {
    await userApi.sendSmsCode('13800000000')
    expect(requestMock).toHaveBeenCalledWith('post', '/user/sms/code', { phone: '13800000000' })
  })

  it('loginBySms → POST /user/login/sms，body 含 phone + code', async () => {
    const data = { phone: '13800000000', code: '123456' }
    await userApi.loginBySms(data)
    expect(requestMock).toHaveBeenCalledWith('post', '/user/login/sms', data)
  })

  it('loginByOAuth → POST /user/login/oauth/:provider，路径参数插值，body 含 code', async () => {
    await userApi.loginByOAuth('wechat', 'oauth-code-abc')
    expect(requestMock).toHaveBeenCalledWith('post', '/user/login/oauth/wechat', { code: 'oauth-code-abc' })
  })

  it('loginByOAuth 不同 provider → 路径正确切换', async () => {
    await userApi.loginByOAuth('github', 'c2')
    expect(requestMock).toHaveBeenCalledWith('post', '/user/login/oauth/github', { code: 'c2' })
  })

  it('getUserInfo → GET /user/info', async () => {
    await userApi.getUserInfo()
    expect(requestMock).toHaveBeenCalledWith('get', '/user/info')
  })

  it('updateProfile → PUT /user/profile，body 透传', async () => {
    const data = { nickname: 'n', avatar: 'a.png' }
    await userApi.updateProfile(data)
    expect(requestMock).toHaveBeenCalledWith('put', '/user/profile', data)
  })

  it('changePassword → PUT /user/password，body 透传', async () => {
    const data = { old_password: 'o', new_password: 'n' }
    await userApi.changePassword(data)
    expect(requestMock).toHaveBeenCalledWith('put', '/user/password', data)
  })
})

describe('user API - 管理后台接口', () => {
  it('listUsers → GET /user/admin/users，params 透传', async () => {
    const params = { page: 1, page_size: 10, keyword: 'foo', status: 1 }
    await userApi.listUsers(params)
    expect(requestMock).toHaveBeenCalledWith('get', '/user/admin/users', { params })
  })

  it('listUsers 无参 → params 为 undefined 仍能调用', async () => {
    await userApi.listUsers()
    expect(requestMock).toHaveBeenCalledWith('get', '/user/admin/users', { params: undefined })
  })

  it('getUser → GET /user/admin/users/:id，路径插值', async () => {
    await userApi.getUser(42)
    expect(requestMock).toHaveBeenCalledWith('get', '/user/admin/users/42')
  })

  it('adminCreateUser → POST /user/admin/users，body 透传', async () => {
    const data = { username: 'new', password: 'p', role_ids: [1, 2] }
    await userApi.adminCreateUser(data)
    expect(requestMock).toHaveBeenCalledWith('post', '/user/admin/users', data)
  })

  it('adminUpdateUser → PUT /user/admin/users/:id，body 透传', async () => {
    const data = { nickname: 'n2' }
    await userApi.adminUpdateUser(7, data)
    expect(requestMock).toHaveBeenCalledWith('put', '/user/admin/users/7', data)
  })

  it('updateUserStatus → PUT /user/admin/users/:id/status，body 仅含 status', async () => {
    await userApi.updateUserStatus(7, 0)
    expect(requestMock).toHaveBeenCalledWith('put', '/user/admin/users/7/status', { status: 0 })
  })

  it('resetUserPassword → PUT /user/admin/users/:id/password，body 仅含 new_password', async () => {
    await userApi.resetUserPassword(7, 'newpass123')
    expect(requestMock).toHaveBeenCalledWith('put', '/user/admin/users/7/password', { new_password: 'newpass123' })
  })

  it('deleteUser → DELETE /user/admin/users/:id', async () => {
    await userApi.deleteUser(9)
    expect(requestMock).toHaveBeenCalledWith('delete', '/user/admin/users/9')
  })
})
