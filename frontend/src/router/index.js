// 路由配置 + 路由守卫
import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { hasPermission } from '@/utils/auth'

const APP_TITLE = import.meta.env.VITE_APP_TITLE || '武昌同城管理后台'

// 静态路由
export const constantRoutes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login.vue'),
    meta: { title: '登录', hidden: true }
  },
  {
    path: '/403',
    name: 'Forbidden',
    component: () => import('@/views/error/403.vue'),
    meta: { title: '无权限', hidden: true }
  },
  {
    path: '/500',
    name: 'ServerError',
    component: () => import('@/views/error/500.vue'),
    meta: { title: '服务器错误', hidden: true }
  },
  {
    path: '/',
    component: () => import('@/layouts/MainLayout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/dashboard.vue'),
        meta: { title: '工作台', icon: 'HomeFilled' }
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('@/views/profile.vue'),
        meta: { title: '个人中心', hidden: true }
      },
      {
        path: 'news',
        name: 'News',
        component: () => import('@/views/news/index.vue'),
        meta: { title: '同城头条', icon: 'Document', permission: 'news:read' }
      },
      {
        path: 'news/detail/:id',
        name: 'NewsDetail',
        component: () => import('@/views/news/detail.vue'),
        meta: { title: '头条详情', hidden: true, permission: 'news:read' }
      },
      {
        path: 'category',
        name: 'Category',
        component: () => import('@/views/category/index.vue'),
        meta: { title: '分类管理', icon: 'Files', permission: 'category:read' }
      },
      {
        path: 'region',
        name: 'Region',
        component: () => import('@/views/region/index.vue'),
        meta: { title: '地区管理', icon: 'Location', permission: 'region:read' }
      },
      {
        path: 'system/user',
        name: 'SystemUser',
        component: () => import('@/views/user/index.vue'),
        meta: { title: '用户管理', icon: 'User', permission: 'user:read' }
      },
      {
        path: 'system/role',
        name: 'SystemRole',
        component: () => import('@/views/role/index.vue'),
        meta: { title: '角色管理', icon: 'UserFilled', permission: 'role:read' }
      },
      {
        path: 'system/permission',
        name: 'SystemPermission',
        component: () => import('@/views/permission/index.vue'),
        meta: { title: '权限管理', icon: 'Lock', permission: 'permission:read' }
      },
      {
        path: 'system/setting',
        name: 'SystemSetting',
        component: () => import('@/views/setting/index.vue'),
        meta: { title: '系统设置', icon: 'Setting', permission: 'setting:read' }
      },
      {
        path: 'system/file',
        name: 'SystemFile',
        component: () => import('@/views/file/index.vue'),
        meta: { title: '文件管理', icon: 'UploadFilled', permission: 'file:read' }
      }
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/404.vue'),
    meta: { title: '页面不存在', hidden: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes: constantRoutes,
  scrollBehavior: () => ({ left: 0, top: 0 })
})

// 全局前置守卫：未登录跳登录页；缺少路由权限跳 403
router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  document.title = to.meta?.title ? `${to.meta.title} - ${APP_TITLE}` : APP_TITLE

  // 错误页与登录页放行
  if (to.path === '/login' || to.path === '/403' || to.path === '/500') {
    // 已登录用户访问登录页，直接跳首页
    if (to.path === '/login' && userStore.isLoggedIn) {
      next('/')
      return
    }
    next()
    return
  }

  if (!userStore.isLoggedIn) {
    next({ path: '/login', query: { redirect: to.fullPath } })
    return
  }

  // 路由级权限校验：meta.permission 缺失则跳 403（admin 直通）
  if (to.meta?.permission && !hasPermission(to.meta.permission)) {
    next({ path: '/403' })
    return
  }

  next()
})

export default router
