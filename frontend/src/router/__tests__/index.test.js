// 单元测试：src/router/index.js 全局前置守卫 beforeEach
// 守卫是 SPA 的安全闸门：未登录跳登录页、已登录访问登录页跳首页、路由级权限校验缺失跳 403。
// 直接捕获 createRouter 调用时注册的 beforeEach 回调，构造 mock (to, from, next) 入参逐分支断言。
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'

// vue-router 在 node 环境下调 createWebHistory 会访问 window/history，因此 stub 掉：
// createWebHistory 返回空对象避免触发 DOM；createRouter 返回一个壳 router，
// 在调用 beforeEach 时把回调缓存到 capturedGuard，供测试用例直接驱动。
// vi.mock 工厂被 Vitest 提升到文件顶部，早于 const 声明执行，故用 vi.hoisted
// 把共享容器同样提升，保证工厂执行时 capturedGuard 已初始化。
const { capturedGuard } = vi.hoisted(() => ({ capturedGuard: { current: null } }))
vi.mock('vue-router', () => ({
  createRouter: ({ routes, scrollBehavior }) => {
    return {
      routes,
      scrollBehavior,
      beforeEach: (cb) => {
        capturedGuard.current = cb
      }
    }
  },
  createWebHistory: () => ({})
}))

// 与 stores/__tests__ 同风格：mock api 层切断 api→request→router 依赖链
// request.js 顶部 import router from '@/router'，若不切断会触发 vue-router 真实加载
vi.mock('@/api/user', () => ({
  login: vi.fn(),
  loginBySms: vi.fn(),
  getUserInfo: vi.fn()
}))
vi.mock('@/api/permission', () => ({
  myAuth: vi.fn()
}))

import { useUserStore } from '@/stores/user'
import router from '../index'

const APP_TITLE = '武昌同城管理后台'

// 构造 mock 路由对象 + next 断言助手
function makeRoute({ path, fullPath, meta }) {
  return {
    path,
    fullPath: fullPath ?? path,
    meta: meta ?? {}
  }
}

// 调用守卫并捕获 next 调用参数
function runGuard(to) {
  const calls = []
  const next = (arg) => {
    calls.push(arg)
  }
  capturedGuard.current(to, {}, next)
  return calls
}

describe('router 全局前置守卫 beforeEach', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    // router/index.js 守卫中直接写 document.title，node 环境无 document，提供最小桩
    globalThis.document = globalThis.document || { title: '' }
    globalThis.document.title = ''
  })

  describe('document.title 设置', () => {
    it('路由 meta.title 存在 → 标题拼接为 "title - APP_TITLE"', () => {
      const to = makeRoute({ path: '/news', meta: { title: '同城头条' } })
      runGuard(to)
      expect(globalThis.document.title).toBe(`同城头条 - ${APP_TITLE}`)
    })

    it('路由 meta.title 缺失 → 标题回退到 APP_TITLE', () => {
      const to = makeRoute({ path: '/unknown', meta: {} })
      runGuard(to)
      expect(globalThis.document.title).toBe(APP_TITLE)
    })

    it('路由 meta 整体缺失 → 标题回退到 APP_TITLE（不抛错）', () => {
      const to = { path: '/unknown', fullPath: '/unknown' /* meta undefined */ }
      runGuard(to)
      expect(globalThis.document.title).toBe(APP_TITLE)
    })
  })

  describe('登录页 / 错误页放行', () => {
    it('/login 未登录 → next() 放行', () => {
      const to = makeRoute({ path: '/login', meta: { title: '登录' } })
      const calls = runGuard(to)
      expect(calls).toHaveLength(1)
      expect(calls[0]).toBeUndefined()
    })

    it('/login 已登录 → next("/") 跳首页', () => {
      const store = useUserStore()
      store.$patch({ token: 't', userInfo: { id: 1, username: 'admin' } })
      const to = makeRoute({ path: '/login', meta: { title: '登录' } })
      const calls = runGuard(to)
      expect(calls).toEqual(['/'])
    })

    it('/403 → next() 放行（与登录态无关）', () => {
      const to = makeRoute({ path: '/403', meta: { title: '无权限' } })
      const calls = runGuard(to)
      expect(calls).toHaveLength(1)
      expect(calls[0]).toBeUndefined()
    })

    it('/500 → next() 放行（与登录态无关）', () => {
      const to = makeRoute({ path: '/500', meta: { title: '服务器错误' } })
      const calls = runGuard(to)
      expect(calls).toHaveLength(1)
      expect(calls[0]).toBeUndefined()
    })

    it('/403 已登录但缺少权限码仍放行（错误页本身不二次鉴权）', () => {
      const store = useUserStore()
      store.$patch({ token: 't', userInfo: { id: 2 }, permissions: [], roles: [] })
      const to = makeRoute({ path: '/403', meta: { title: '无权限' } })
      const calls = runGuard(to)
      expect(calls).toEqual([undefined])
    })
  })

  describe('未登录访问受保护路由 → 跳 /login 携带 redirect', () => {
    it('/news 未登录 → next({ path, query: { redirect: fullPath } })', () => {
      const to = makeRoute({
        path: '/news',
        fullPath: '/news?page=1',
        meta: { title: '同城头条', permission: 'news:read' }
      })
      const calls = runGuard(to)
      expect(calls).toEqual([
        { path: '/login', query: { redirect: '/news?page=1' } }
      ])
    })

    it('/dashboard 无 meta.permission 也未登录 → 仍跳 /login', () => {
      const to = makeRoute({ path: '/dashboard', meta: { title: '工作台' } })
      const calls = runGuard(to)
      expect(calls).toEqual([
        { path: '/login', query: { redirect: '/dashboard' } }
      ])
    })

    it('redirect 取 to.fullPath 而非 to.path（保留 query）', () => {
      const to = makeRoute({
        path: '/news',
        fullPath: '/news?category=2&keyword=foo',
        meta: { title: '同城头条', permission: 'news:read' }
      })
      const calls = runGuard(to)
      expect(calls[0].query.redirect).toBe('/news?category=2&keyword=foo')
    })
  })

  describe('已登录 + 路由权限校验', () => {
    function loginAs({ permissions = [], roles = [], token = 't' } = {}) {
      const store = useUserStore()
      store.$patch({
        token,
        userInfo: { id: 1, username: 'tester' },
        permissions,
        roles
      })
      return store
    }

    it('拥有 meta.permission 对应权限码 → next() 放行', () => {
      loginAs({ permissions: ['news:read'] })
      const to = makeRoute({
        path: '/news',
        meta: { title: '同城头条', permission: 'news:read' }
      })
      const calls = runGuard(to)
      expect(calls).toEqual([undefined])
    })

    it('缺少 meta.permission 权限码 → next({ path: "/403" })', () => {
      loginAs({ permissions: ['category:read'] })
      const to = makeRoute({
        path: '/system/user',
        meta: { title: '用户管理', permission: 'user:read' }
      })
      const calls = runGuard(to)
      expect(calls).toEqual([{ path: '/403' }])
    })

    it('admin 角色（超管）对任意 meta.permission 直通', () => {
      loginAs({ roles: ['admin'], permissions: [] })
      const to = makeRoute({
        path: '/system/permission',
        meta: { title: '权限管理', permission: 'permission:read' }
      })
      const calls = runGuard(to)
      expect(calls).toEqual([undefined])
    })

    it('meta.permission 缺失（如 dashboard）→ next() 放行，不进入 403 分支', () => {
      loginAs({ permissions: [] })
      const to = makeRoute({ path: '/dashboard', meta: { title: '工作台' } })
      const calls = runGuard(to)
      expect(calls).toEqual([undefined])
    })

    it('同时持有多个权限码，命中其中之一即放行', () => {
      loginAs({ permissions: ['region:read', 'category:read'] })
      const to = makeRoute({
        path: '/region',
        meta: { title: '地区管理', permission: 'region:read' }
      })
      const calls = runGuard(to)
      expect(calls).toEqual([undefined])
    })

    it('权限码数组形式的 meta.permission 任一命中即放行（hasPermission 数组语义）', () => {
      loginAs({ permissions: ['setting:read'] })
      const to = makeRoute({
        path: '/system/setting',
        meta: { title: '系统设置', permission: ['setting:read', 'setting:write'] }
      })
      const calls = runGuard(to)
      expect(calls).toEqual([undefined])
    })
  })

  describe('守卫注册', () => {
    it('createRouter 调用后已通过 router.beforeEach 注册守卫回调', () => {
      expect(capturedGuard.current).toBeTypeOf('function')
    })

    it('router 实例同时导出 routes 配置（供 MainLayout 菜单派生）', () => {
      expect(Array.isArray(router.routes)).toBe(true)
      // 至少包含根路径与登录页两条顶级路由
      const paths = router.routes.map((r) => r.path)
      expect(paths).toContain('/')
      expect(paths).toContain('/login')
    })
  })
})
