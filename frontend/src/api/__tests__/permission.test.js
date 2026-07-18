// 单元测试：src/api/permission.js
// 验证每个 API 函数调用 request 的方法、URL、入参（含角色/权限/分配/当前用户授权）
import { describe, it, expect, beforeEach, vi } from 'vitest'

const requestMock = vi.fn()
vi.mock('@/utils/request', () => ({
  default: {
    get: (...a) => requestMock('get', ...a),
    post: (...a) => requestMock('post', ...a),
    put: (...a) => requestMock('put', ...a),
    delete: (...a) => requestMock('delete', ...a)
  }
}))

import * as permissionApi from '../permission'

beforeEach(() => {
  requestMock.mockReset()
  requestMock.mockImplementation(() => Promise.resolve({ code: 0, data: {} }))
})

describe('permission API - 角色', () => {
  it('listRoles → GET /permission/roles', async () => {
    await permissionApi.listRoles()
    expect(requestMock).toHaveBeenCalledWith('get', '/permission/roles')
  })

  it('getRole → GET /permission/roles/:id', async () => {
    await permissionApi.getRole(3)
    expect(requestMock).toHaveBeenCalledWith('get', '/permission/roles/3')
  })

  it('createRole → POST /permission/roles，body 透传', async () => {
    const data = { name: '编辑', code: 'editor', permission_ids: [1, 2] }
    await permissionApi.createRole(data)
    expect(requestMock).toHaveBeenCalledWith('post', '/permission/roles', data)
  })

  it('updateRole → PUT /permission/roles/:id，body 透传', async () => {
    const data = { name: '编辑2' }
    await permissionApi.updateRole(3, data)
    expect(requestMock).toHaveBeenCalledWith('put', '/permission/roles/3', data)
  })

  it('deleteRole → DELETE /permission/roles/:id', async () => {
    await permissionApi.deleteRole(3)
    expect(requestMock).toHaveBeenCalledWith('delete', '/permission/roles/3')
  })

  it('getRolePermissions → GET /permission/roles/:roleId/permissions（回显用）', async () => {
    await permissionApi.getRolePermissions(3)
    expect(requestMock).toHaveBeenCalledWith('get', '/permission/roles/3/permissions')
  })
})

describe('permission API - 权限', () => {
  it('listPermissions → GET /permission/permissions', async () => {
    await permissionApi.listPermissions()
    expect(requestMock).toHaveBeenCalledWith('get', '/permission/permissions')
  })

  it('createPermission → POST /permission/permissions，body 透传', async () => {
    const data = { name: '创建用户', code: 'user:create' }
    await permissionApi.createPermission(data)
    expect(requestMock).toHaveBeenCalledWith('post', '/permission/permissions', data)
  })

  it('updatePermission → PUT /permission/permissions/:id，body 透传', async () => {
    const data = { name: '改名' }
    await permissionApi.updatePermission(5, data)
    expect(requestMock).toHaveBeenCalledWith('put', '/permission/permissions/5', data)
  })

  it('deletePermission → DELETE /permission/permissions/:id', async () => {
    await permissionApi.deletePermission(5)
    expect(requestMock).toHaveBeenCalledWith('delete', '/permission/permissions/5')
  })
})

describe('permission API - 分配', () => {
  it('assignRoles → POST /permission/assign-roles，body 透传', async () => {
    const data = { user_id: 1, role_ids: [1, 2] }
    await permissionApi.assignRoles(data)
    expect(requestMock).toHaveBeenCalledWith('post', '/permission/assign-roles', data)
  })

  it('assignPermissions → POST /permission/assign-permissions，body 透传', async () => {
    const data = { role_id: 1, permission_ids: [1, 2, 3] }
    await permissionApi.assignPermissions(data)
    expect(requestMock).toHaveBeenCalledWith('post', '/permission/assign-permissions', data)
  })

  it('getUserRoles → GET /permission/users/:userId/roles', async () => {
    await permissionApi.getUserRoles(7)
    expect(requestMock).toHaveBeenCalledWith('get', '/permission/users/7/roles')
  })
})

describe('permission API - 当前用户授权', () => {
  it('myPermissions → GET /permission/my-permissions', async () => {
    await permissionApi.myPermissions()
    expect(requestMock).toHaveBeenCalledWith('get', '/permission/my-permissions')
  })

  it('myAuth → GET /permission/my-auth', async () => {
    await permissionApi.myAuth()
    expect(requestMock).toHaveBeenCalledWith('get', '/permission/my-auth')
  })
})
