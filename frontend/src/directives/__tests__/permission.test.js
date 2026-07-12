// 单元测试：src/directives/permission.js 的 v-permission / v-role 自定义指令
// 验证 mounted/updated 钩子：有权限保留元素、无权限 .hide 隐藏、无权限默认移除元素
import { describe, it, expect, beforeEach, vi } from 'vitest'

// mock @/utils/auth：hasPermission / hasRole 行为可控（指令内部通过 evaluate 调用）
const hasPermissionMock = vi.fn()
const hasRoleMock = vi.fn()
vi.mock('@/utils/auth', () => ({
  hasPermission: (...a) => hasPermissionMock(...a),
  hasRole: (...a) => hasRoleMock(...a)
}))

// 动态 import 获取指令模块（配合 vi.resetModules + mock 工厂惰性 wrapper 重置状态）
async function loadDirectives() {
  vi.resetModules()
  const mod = await import('../permission')
  return { permission: mod.permission, role: mod.role }
}

// 构造一个伪元素 + 伪父节点，模拟 Vue 指令钩子入参 el
function createElement() {
  const parent = { removeChild: vi.fn() }
  const el = {
    style: {},
    parentNode: parent
  }
  return { el, parent }
}

// 构造一个 binding 对象（value + modifiers）
function binding(value, modifiers = {}) {
  return { value, modifiers }
}

beforeEach(() => {
  vi.resetModules()
  hasPermissionMock.mockReset()
  hasRoleMock.mockReset()
})

describe('v-permission 指令 - mounted 钩子', () => {
  it('有权限 → 元素保留（不移除、不隐藏）', async () => {
    const { permission } = await loadDirectives()
    const { el, parent } = createElement()
    hasPermissionMock.mockReturnValue(true)

    permission.mounted(el, binding('user:create'))

    expect(hasPermissionMock).toHaveBeenCalledWith('user:create')
    expect(parent.removeChild).not.toHaveBeenCalled()
    expect(el.style.display).toBeUndefined()
  })

  it('无权限且无 .hide 修饰符 → 从父节点移除元素', async () => {
    const { permission } = await loadDirectives()
    const { el, parent } = createElement()
    hasPermissionMock.mockReturnValue(false)

    permission.mounted(el, binding('user:delete'))

    expect(parent.removeChild).toHaveBeenCalledWith(el)
    expect(el.style.display).toBeUndefined()
  })

  it('无权限且 .hide 修饰符 → 仅设置 display:none 不移除', async () => {
    const { permission } = await loadDirectives()
    const { el, parent } = createElement()
    hasPermissionMock.mockReturnValue(false)

    permission.mounted(el, binding('user:delete', { hide: true }))

    expect(parent.removeChild).not.toHaveBeenCalled()
    expect(el.style.display).toBe('none')
  })

  it('权限码数组形式 → 整个数组传给 hasPermission', async () => {
    const { permission } = await loadDirectives()
    const { el } = createElement()
    hasPermissionMock.mockReturnValue(true)

    const codes = ['user:update', 'user:delete']
    permission.mounted(el, binding(codes))

    expect(hasPermissionMock).toHaveBeenCalledWith(codes)
  })

  it('元素无 parentNode 时无权限不抛错（优雅降级）', async () => {
    const { permission } = await loadDirectives()
    const { el } = createElement()
    el.parentNode = null
    hasPermissionMock.mockReturnValue(false)

    expect(() => permission.mounted(el, binding('user:delete'))).not.toThrow()
  })
})

describe('v-permission 指令 - updated 钩子', () => {
  it('非 .hide 模式 → updated 为空操作（不修改 style）', async () => {
    const { permission } = await loadDirectives()
    const { el } = createElement()
    hasPermissionMock.mockReturnValue(true)

    permission.updated(el, binding('user:create'))

    expect(el.style.display).toBeUndefined()
  })

  it('.hide 模式 + 有权限 → display 恢复为空串', async () => {
    const { permission } = await loadDirectives()
    const { el } = createElement()
    el.style.display = 'none'
    hasPermissionMock.mockReturnValue(true)

    permission.updated(el, binding('user:create', { hide: true }))

    expect(el.style.display).toBe('')
  })

  it('.hide 模式 + 无权限 → display 保持 none', async () => {
    const { permission } = await loadDirectives()
    const { el } = createElement()
    el.style.display = 'none'
    hasPermissionMock.mockReturnValue(false)

    permission.updated(el, binding('user:delete', { hide: true }))

    expect(el.style.display).toBe('none')
  })

  it('.hide 模式 + 权限由有变无 → display 切为 none', async () => {
    const { permission } = await loadDirectives()
    const { el } = createElement()
    el.style.display = ''
    hasPermissionMock.mockReturnValue(false)

    permission.updated(el, binding('user:delete', { hide: true }))

    expect(el.style.display).toBe('none')
  })
})

describe('v-role 指令 - mounted 钩子', () => {
  it('有角色 → 元素保留', async () => {
    const { role } = await loadDirectives()
    const { el, parent } = createElement()
    hasRoleMock.mockReturnValue(true)

    role.mounted(el, binding('admin'))

    expect(hasRoleMock).toHaveBeenCalledWith('admin')
    expect(parent.removeChild).not.toHaveBeenCalled()
  })

  it('无角色且无 .hide → 从父节点移除', async () => {
    const { role } = await loadDirectives()
    const { el, parent } = createElement()
    hasRoleMock.mockReturnValue(false)

    role.mounted(el, binding('admin'))

    expect(parent.removeChild).toHaveBeenCalledWith(el)
  })

  it('无角色且 .hide → display:none 不移除', async () => {
    const { role } = await loadDirectives()
    const { el, parent } = createElement()
    hasRoleMock.mockReturnValue(false)

    role.mounted(el, binding('admin', { hide: true }))

    expect(parent.removeChild).not.toHaveBeenCalled()
    expect(el.style.display).toBe('none')
  })

  it('角色码数组形式 → 整个数组传给 hasRole', async () => {
    const { role } = await loadDirectives()
    const { el } = createElement()
    hasRoleMock.mockReturnValue(true)

    const codes = ['admin', 'editor']
    role.mounted(el, binding(codes))

    expect(hasRoleMock).toHaveBeenCalledWith(codes)
  })
})

describe('v-role 指令 - updated 钩子', () => {
  it('非 .hide 模式 → updated 为空操作', async () => {
    const { role } = await loadDirectives()
    const { el } = createElement()
    hasRoleMock.mockReturnValue(true)

    role.updated(el, binding('admin'))

    expect(el.style.display).toBeUndefined()
  })

  it('.hide 模式 + 有角色 → display 恢复为空串', async () => {
    const { role } = await loadDirectives()
    const { el } = createElement()
    el.style.display = 'none'
    hasRoleMock.mockReturnValue(true)

    role.updated(el, binding('admin', { hide: true }))

    expect(el.style.display).toBe('')
  })

  it('.hide 模式 + 无角色 → display 保持 none', async () => {
    const { role } = await loadDirectives()
    const { el } = createElement()
    el.style.display = 'none'
    hasRoleMock.mockReturnValue(false)

    role.updated(el, binding('admin', { hide: true }))

    expect(el.style.display).toBe('none')
  })
})

describe('指令导出形态', () => {
  it('permission 与 role 均导出含 mounted/updated 钩子的对象', async () => {
    const { permission, role } = await loadDirectives()
    expect(typeof permission.mounted).toBe('function')
    expect(typeof permission.updated).toBe('function')
    expect(typeof role.mounted).toBe('function')
    expect(typeof role.updated).toBe('function')
  })
})
