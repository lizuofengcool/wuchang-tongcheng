// 单元测试：src/utils/auth.js 权限/角色校验工具
// 依赖 pinia store，通过 setActivePinia + $patch 构造不同用户态
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'

// stores/user 会 import api/user & api/permission，二者经 utils/request 拉入 router，
// router 的 createWebHistory 依赖浏览器 window。这里 mock 掉 api 层，切断该依赖链，
// 使测试聚焦于 auth 工具的权限/角色判定逻辑，不引入 DOM/router。
vi.mock('@/api/user', () => ({
  login: vi.fn(),
  getUserInfo: vi.fn()
}))
vi.mock('@/api/permission', () => ({
  myAuth: vi.fn()
}))

import { useUserStore } from '@/stores/user'
import { hasPermission, hasRole, hasAllPermissions } from '../auth'

// 构造一个指定权限/角色的用户并使其成为当前激活 store
function setUser({ permissions = [], roles = [] }) {
  const store = useUserStore()
  store.$patch({ permissions, roles })
  return store
}

describe('hasPermission', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('超管（admin 角色）对任意权限码直通', () => {
    setUser({ roles: ['admin'], permissions: [] })
    expect(hasPermission('user:delete')).toBe(true)
    expect(hasPermission('any:thing')).toBe(true)
  })

  it('普通用户拥有该权限码 → true', () => {
    setUser({ roles: ['editor'], permissions: ['news:read', 'news:create'] })
    expect(hasPermission('news:read')).toBe(true)
  })

  it('普通用户缺少该权限码 → false', () => {
    setUser({ roles: ['editor'], permissions: ['news:read'] })
    expect(hasPermission('user:delete')).toBe(false)
  })

  it('权限码数组：任一满足即通过', () => {
    setUser({ roles: ['editor'], permissions: ['news:read'] })
    expect(hasPermission(['news:read', 'news:delete'])).toBe(true)
    expect(hasPermission(['user:update', 'user:delete'])).toBe(false)
  })

  it('空权限码直通（视为无需权限）', () => {
    setUser({ roles: ['editor'], permissions: [] })
    expect(hasPermission('')).toBe(true)
    expect(hasPermission(null)).toBe(true)
    expect(hasPermission(undefined)).toBe(true)
  })
})

describe('hasRole', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('拥有该角色 → true', () => {
    setUser({ roles: ['editor'] })
    expect(hasRole('editor')).toBe(true)
  })

  it('不拥有该角色 → false', () => {
    setUser({ roles: ['editor'] })
    expect(hasRole('admin')).toBe(false)
  })

  it('角色数组：任一满足即通过', () => {
    setUser({ roles: ['editor', 'auditor'] })
    expect(hasRole(['admin', 'auditor'])).toBe(true)
    expect(hasRole(['admin', 'super'])).toBe(false)
  })

  it('空角色码直通', () => {
    setUser({ roles: ['editor'] })
    expect(hasRole('')).toBe(true)
    expect(hasRole(null)).toBe(true)
  })
})

describe('hasAllPermissions', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('超管直通', () => {
    setUser({ roles: ['admin'], permissions: [] })
    expect(hasAllPermissions(['user:read', 'user:delete'])).toBe(true)
  })

  it('全部权限都满足 → true', () => {
    setUser({ roles: ['editor'], permissions: ['news:read', 'news:create', 'news:update'] })
    expect(hasAllPermissions(['news:read', 'news:create'])).toBe(true)
  })

  it('部分权限缺失 → false', () => {
    setUser({ roles: ['editor'], permissions: ['news:read'] })
    expect(hasAllPermissions(['news:read', 'news:delete'])).toBe(false)
  })

  it('空数组 / null 直通', () => {
    setUser({ roles: ['editor'], permissions: [] })
    expect(hasAllPermissions([])).toBe(true)
    expect(hasAllPermissions(null)).toBe(true)
    expect(hasAllPermissions(undefined)).toBe(true)
  })
})
