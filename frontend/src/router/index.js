// 路由配置 + 路由守卫
import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'

// 静态路由
export const constantRoutes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login.vue'),
    meta: { title: '登录', hidden: true }
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
        meta: { title: '同城头条', icon: 'Document' }
      },
      {
        path: 'category',
        name: 'Category',
        component: () => import('@/views/category/index.vue'),
        meta: { title: '分类管理', icon: 'Files' }
      },
      {
        path: 'region',
        name: 'Region',
        component: () => import('@/views/region/index.vue'),
        meta: { title: '地区管理', icon: 'Location' }
      },
      {
        path: 'system/user',
        name: 'SystemUser',
        component: () => import('@/views/user/index.vue'),
        meta: { title: '用户管理', icon: 'User' }
      },
      {
        path: 'system/role',
        name: 'SystemRole',
        component: () => import('@/views/role/index.vue'),
        meta: { title: '角色管理', icon: 'UserFilled' }
      },
      {
        path: 'system/permission',
        name: 'SystemPermission',
        component: () => import('@/views/permission/index.vue'),
        meta: { title: '权限管理', icon: 'Lock' }
      },
      {
        path: 'system/setting',
        name: 'SystemSetting',
        component: () => import('@/views/setting/index.vue'),
        meta: { title: '系统设置', icon: 'Setting' }
      },
      {
        path: 'system/file',
        name: 'SystemFile',
        component: () => import('@/views/file/index.vue'),
        meta: { title: '文件管理', icon: 'UploadFilled' }
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

// 全局前置守卫：未登录跳登录页
router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  document.title = to.meta?.title ? `${to.meta.title} - 武昌同城管理后台` : '武昌同城管理后台'

  if (to.path === '/login') {
    // 已登录用户访问登录页，直接跳首页
    if (userStore.isLoggedIn) {
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

  next()
})

export default router
