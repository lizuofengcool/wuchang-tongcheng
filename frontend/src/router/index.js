// 路由配置 + 路由守卫
// v3.3.0 八大业务领域 Tab 架构（按大厂中后台业务领域组织，不按技术分层）
// 顶部 8 个 Tab：工作台 / 同城业务 / 商家服务 / 营销活动 / 用户运营 / 社区互动 / 数据中心 / 设置中心
// 中台服务（pay/im/material/risk/ai 等）归入设置中心，前端只暴露配置入口
import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { hasPermission } from '@/utils/auth'

const APP_TITLE = import.meta.env.VITE_APP_TITLE || '近享同城管理后台'

// 静态路由（八大业务领域 Tab 架构）
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
    redirect: '/workspace',
    children: [
      // ===== 个人中心（隐藏路由，不在菜单显示） =====
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('@/views/profile.vue'),
        meta: { title: '个人中心', hidden: true }
      },

      // ===== 1. 工作台 =====
      {
        path: 'workspace',
        name: 'Workspace',
        redirect: '/workspace/dashboard',
        meta: { title: '工作台', icon: 'HomeFilled', menuGroup: 'workspace', menuLevel: 1 },
        children: [
          {
            path: 'dashboard',
            name: 'Dashboard',
            component: () => import('@/views/dashboard.vue'),
            meta: { title: '仪表盘', icon: 'HomeFilled', menuLevel: 2 }
          },
          {
            path: 'message',
            name: 'Message',
            component: () => import('@/views/message/index.vue'),
            meta: { title: '消息中心', icon: 'Bell', menuLevel: 2 }
          }
        ]
      },

      // ===== 2. 同城业务 =====
      // 8 个垂直业务模块（ershou 已完成完整功能，其他 7 个为开发中空壳）
      {
        path: 'business',
        name: 'Business',
        redirect: '/business/ershou',
        meta: { title: '同城业务', icon: 'ShoppingBag', menuGroup: 'business', menuLevel: 1 },
        children: [
          // 二手交易（完整功能：16 个管理页面）
          {
            path: 'ershou',
            name: 'Ershou',
            redirect: '/business/ershou/list',
            meta: { title: '二手交易', icon: 'ShoppingBag', permission: 'ershou:read', menuLevel: 2 },
            children: [
              { path: 'list', name: 'ErshouList', component: () => import('@/views/business/ershou/index.vue'), meta: { title: '商品列表', icon: 'List', permission: 'ershou:read', menuLevel: 3 } },
              { path: 'detail/:id', name: 'ErshouDetail', component: () => import('@/views/business/ershou/detail.vue'), meta: { title: '商品详情', hidden: true, permission: 'ershou:read' } },
              { path: 'orders', name: 'ErshouOrders', component: () => import('@/views/business/ershou/orders.vue'), meta: { title: '订单管理', icon: 'Tickets', permission: 'ershou:read', menuLevel: 3 } },
              { path: 'auctions', name: 'ErshouAuctions', component: () => import('@/views/business/ershou/auctions.vue'), meta: { title: '拍卖管理', icon: 'Hammer', permission: 'ershou:read', menuLevel: 3 } },
              { path: 'promotions', name: 'ErshouPromotions', component: () => import('@/views/business/ershou/promotions.vue'), meta: { title: '付费推广', icon: 'Promotion', permission: 'ershou:read', menuLevel: 3 } },
              { path: 'refunds', name: 'ErshouRefunds', component: () => import('@/views/business/ershou/refunds.vue'), meta: { title: '退款管理', icon: 'RefreshLeft', permission: 'ershou:read', menuLevel: 3 } },
              { path: 'reports', name: 'ErshouReports', component: () => import('@/views/business/ershou/reports.vue'), meta: { title: '举报管理', icon: 'Warning', permission: 'ershou:read', menuLevel: 3 } },
              { path: 'reviews', name: 'ErshouReviews', component: () => import('@/views/business/ershou/reviews.vue'), meta: { title: '评价管理', icon: 'ChatDotRound', permission: 'ershou:read', menuLevel: 3 } },
              { path: 'shops', name: 'ErshouShops', component: () => import('@/views/business/ershou/shops.vue'), meta: { title: '商家店铺', icon: 'Shop', permission: 'ershou:read', menuLevel: 3 } },
              { path: 'tags', name: 'ErshouTags', component: () => import('@/views/business/ershou/tags.vue'), meta: { title: '标签管理', icon: 'PriceTag', permission: 'ershou:read', menuLevel: 3 } },
              { path: 'brands', name: 'ErshouBrands', component: () => import('@/views/business/ershou/brands.vue'), meta: { title: '品牌型号', icon: 'Collection', permission: 'ershou:read', menuLevel: 3 } },
              { path: 'category-attrs', name: 'ErshouCategoryAttrs', component: () => import('@/views/business/ershou/category-attrs.vue'), meta: { title: '分类属性', icon: 'Files', permission: 'ershou:read', menuLevel: 3 } },
              { path: 'audit-rules', name: 'ErshouAuditRules', component: () => import('@/views/business/ershou/audit-rules.vue'), meta: { title: '审核规则', icon: 'Filter', permission: 'content:audit', menuLevel: 3 } },
              { path: 'user-credit', name: 'ErshouUserCredit', component: () => import('@/views/business/ershou/user-credit.vue'), meta: { title: '用户信用', icon: 'CreditCard', permission: 'ershou:read', menuLevel: 3 } },
              { path: 'statistics', name: 'ErshouStatistics', component: () => import('@/views/business/ershou/statistics.vue'), meta: { title: '数据统计', icon: 'DataLine', permission: 'ershou:read', menuLevel: 3 } },
              { path: 'batch', name: 'ErshouBatch', component: () => import('@/views/business/ershou/batch.vue'), meta: { title: '批量操作', icon: 'Operation', permission: 'content:audit', menuLevel: 3 } }
            ]
          },
          // 其他 7 个垂直业务模块（空壳，开发中）
          { path: 'job', name: 'Job', component: () => import('@/views/business/job/index.vue'), meta: { title: '招聘求职', icon: 'Briefcase', permission: 'job:read', menuLevel: 2 } },
          { path: 'fang', name: 'Fang', component: () => import('@/views/business/fang/index.vue'), meta: { title: '房屋租售', icon: 'House', permission: 'fang:read', menuLevel: 2 } },
          { path: 'car', name: 'Car', component: () => import('@/views/business/car/index.vue'), meta: { title: '车辆买卖', icon: 'Van', permission: 'car:read', menuLevel: 2 } },
          { path: 'love', name: 'Love', component: () => import('@/views/business/love/index.vue'), meta: { title: '相亲交友', icon: 'User', permission: 'love:read', menuLevel: 2 } },
          { path: 'pinche', name: 'Pinche', component: () => import('@/views/business/pinche/index.vue'), meta: { title: '拼车出行', icon: 'Van', permission: 'pinche:read', menuLevel: 2 } },
          { path: 'linggong', name: 'Linggong', component: () => import('@/views/business/linggong/index.vue'), meta: { title: '零工兼职', icon: 'Briefcase', permission: 'linggong:read', menuLevel: 2 } },
          { path: 'dh114', name: 'Dh114', component: () => import('@/views/business/dh114/index.vue'), meta: { title: '同城114', icon: 'Phone', permission: 'dh114:read', menuLevel: 2 } }
        ]
      },

      // ===== 3. 商家服务 =====
      // 商家管理 + 6 个商家业务模块（mall/yuyue/daojia/diancan/huodong/quan 空壳）
      {
        path: 'merchant-service',
        name: 'MerchantService',
        redirect: '/merchant-service/shop',
        meta: { title: '商家服务', icon: 'Shop', menuGroup: 'merchant-service', menuLevel: 1 },
        children: [
          {
            path: 'shop',
            name: 'Shop',
            redirect: '/merchant-service/shop/list',
            meta: { title: '商家管理', icon: 'Shop', permission: 'shop:read', menuLevel: 2 },
            children: [
              { path: 'list', name: 'ShopList', component: () => import('@/views/shop/index.vue'), meta: { title: '商家列表', icon: 'List', permission: 'shop:read', menuLevel: 3 } },
              { path: 'reviews', name: 'ShopReviews', component: () => import('@/views/shop/reviews.vue'), meta: { title: '评价审核', icon: 'ChatDotRound', permission: 'shop:audit', menuLevel: 3 } }
            ]
          },
          { path: 'mall', name: 'Mall', component: () => import('@/views/shop-service/mall/index.vue'), meta: { title: '同城商城', icon: 'ShoppingCart', permission: 'mall:read', menuLevel: 2 } },
          { path: 'yuyue', name: 'Yuyue', component: () => import('@/views/shop-service/yuyue/index.vue'), meta: { title: '预约服务', icon: 'Calendar', permission: 'yuyue:read', menuLevel: 2 } },
          { path: 'daojia', name: 'Daojia', component: () => import('@/views/shop-service/daojia/index.vue'), meta: { title: '同城到家', icon: 'Service', permission: 'daojia:read', menuLevel: 2 } },
          { path: 'diancan', name: 'Diancan', component: () => import('@/views/shop-service/diancan/index.vue'), meta: { title: '同城点餐', icon: 'Food', permission: 'diancan:read', menuLevel: 2 } },
          { path: 'huodong', name: 'Huodong', component: () => import('@/views/shop-service/huodong/index.vue'), meta: { title: '同城活动', icon: 'Football', permission: 'huodong:read', menuLevel: 2 } },
          { path: 'quan', name: 'Quan', component: () => import('@/views/shop-service/quan/index.vue'), meta: { title: '同城圈子', icon: 'ChatLineRound', permission: 'quan:read', menuLevel: 2 } }
        ]
      },

      // ===== 4. 营销活动 =====
      {
        path: 'marketing',
        name: 'Marketing',
        redirect: '/marketing/groupbuy',
        meta: { title: '营销活动', icon: 'Present', menuGroup: 'marketing', menuLevel: 1 },
        children: [
          {
            path: 'groupbuy',
            name: 'Groupbuy',
            redirect: '/marketing/groupbuy/list',
            meta: { title: '团购管理', icon: 'Present', permission: 'groupbuy:read', menuLevel: 2 },
            children: [
              { path: 'list', name: 'GroupbuyList', component: () => import('@/views/marketing/groupbuy/index.vue'), meta: { title: '团购商品', icon: 'List', permission: 'groupbuy:read', menuLevel: 3 } },
              { path: 'orders', name: 'GroupbuyOrders', component: () => import('@/views/marketing/groupbuy/orders.vue'), meta: { title: '订单管理', icon: 'Tickets', permission: 'groupbuy:read', menuLevel: 3 } },
              { path: 'coupons', name: 'GroupbuyCoupons', component: () => import('@/views/marketing/groupbuy/coupons.vue'), meta: { title: '优惠券', icon: 'Ticket', permission: 'groupbuy:read', menuLevel: 3 } }
            ]
          },
          { path: 'coupon', name: 'Coupon', component: () => import('@/views/marketing/coupon/index.vue'), meta: { title: '优惠券系统', icon: 'Ticket', permission: 'coupon:read', menuLevel: 2 } },
          { path: 'kanjia', name: 'Kanjia', component: () => import('@/views/marketing/kanjia/index.vue'), meta: { title: '砍价活动', icon: 'Discount', permission: 'kanjia:read', menuLevel: 2 } },
          { path: 'pintuan', name: 'Pintuan', component: () => import('@/views/marketing/pintuan/index.vue'), meta: { title: '拼团活动', icon: 'Connection', permission: 'pintuan:read', menuLevel: 2 } },
          { path: 'choujiang', name: 'Choujiang', component: () => import('@/views/marketing/choujiang/index.vue'), meta: { title: '抽奖活动', icon: 'Trophy', permission: 'choujiang:read', menuLevel: 2 } },
          { path: 'sign', name: 'Sign', component: () => import('@/views/marketing/sign/index.vue'), meta: { title: '签到积分', icon: 'Calendar', permission: 'sign:read', menuLevel: 2 } }
        ]
      },

      // ===== 5. 用户运营 =====
      {
        path: 'user-op',
        name: 'UserOp',
        redirect: '/user-op/user',
        meta: { title: '用户运营', icon: 'User', menuGroup: 'user-op', menuLevel: 1 },
        children: [
          { path: 'user', name: 'SystemUser', component: () => import('@/views/user/index.vue'), meta: { title: '用户管理', icon: 'User', permission: 'user:read', menuLevel: 2 } },
          { path: 'renzheng', name: 'Renzheng', component: () => import('@/views/user-op/renzheng/index.vue'), meta: { title: '用户认证', icon: 'Postcard', permission: 'renzheng:read', menuLevel: 2 } },
          { path: 'vipcard', name: 'Vipcard', component: () => import('@/views/user-op/vipcard/index.vue'), meta: { title: '会员卡', icon: 'CreditCard', permission: 'vipcard:read', menuLevel: 2 } },
          { path: 'partner', name: 'Partner', component: () => import('@/views/user-op/partner/index.vue'), meta: { title: '同城合伙人', icon: 'UserFilled', permission: 'partner:read', menuLevel: 2 } },
          { path: 'majia', name: 'Majia', component: () => import('@/views/user-op/majia/index.vue'), meta: { title: '马甲管理', icon: 'Avatar', permission: 'majia:read', menuLevel: 2 } },
          { path: 'jubao', name: 'Jubao', component: () => import('@/views/user-op/jubao/index.vue'), meta: { title: '举报中心', icon: 'Warning', permission: 'jubao:read', menuLevel: 2 } }
        ]
      },

      // ===== 6. 社区互动 =====
      // 内容管理（同城头条/分类）+ 社区功能（打赏/名片/直播/AI/推文/分享）
      {
        path: 'community',
        name: 'Community',
        redirect: '/community/news',
        meta: { title: '社区互动', icon: 'ChatLineRound', menuGroup: 'community', menuLevel: 1 },
        children: [
          // 内容管理分组
          {
            path: 'news',
            name: 'News',
            redirect: '/community/news/list',
            meta: { title: '同城头条', icon: 'Document', permission: 'news:read', menuLevel: 2 },
            children: [
              { path: 'list', name: 'NewsList', component: () => import('@/views/news/index.vue'), meta: { title: '头条列表', icon: 'List', permission: 'news:read', menuLevel: 3 } },
              { path: 'detail/:id', name: 'NewsDetail', component: () => import('@/views/news/detail.vue'), meta: { title: '头条详情', hidden: true, permission: 'news:read' } },
              { path: 'favorites', name: 'NewsFavorites', component: () => import('@/views/news/favorites.vue'), meta: { title: '我的收藏', icon: 'Star', permission: 'news:read', menuLevel: 3 } }
            ]
          },
          { path: 'category', name: 'Category', component: () => import('@/views/category/index.vue'), meta: { title: '分类管理', icon: 'Files', permission: 'category:read', menuLevel: 2 } },
          // 社区功能
          { path: 'dashang', name: 'Dashang', component: () => import('@/views/community/dashang/index.vue'), meta: { title: '打赏记录', icon: 'Coin', permission: 'dashang:read', menuLevel: 2 } },
          { path: 'mingpian', name: 'Mingpian', component: () => import('@/views/community/mingpian/index.vue'), meta: { title: '名片管理', icon: 'Postcard', permission: 'mingpian:read', menuLevel: 2 } },
          { path: 'zhibo', name: 'Zhibo', component: () => import('@/views/community/zhibo/index.vue'), meta: { title: '直播管理', icon: 'VideoCamera', permission: 'zhibo:read', menuLevel: 2 } },
          { path: 'ai', name: 'Ai', component: () => import('@/views/community/ai/index.vue'), meta: { title: 'AI助手', icon: 'MagicStick', permission: 'ai:read', menuLevel: 2 } },
          { path: 'tuiwen', name: 'Tuiwen', component: () => import('@/views/community/tuiwen/index.vue'), meta: { title: '推文助手', icon: 'EditPen', permission: 'tuiwen:read', menuLevel: 2 } },
          { path: 'share', name: 'Share', component: () => import('@/views/community/share/index.vue'), meta: { title: '分享统计', icon: 'Share', permission: 'share:read', menuLevel: 2 } }
        ]
      },

      // ===== 7. 数据中心 =====
      // 数据概览 + 模块数据 + 财务报表 + 模块运营（总控/灰度/监控）
      {
        path: 'data-center',
        name: 'DataCenter',
        redirect: '/data-center/overview',
        meta: { title: '数据中心', icon: 'TrendCharts', menuGroup: 'data-center', menuLevel: 1 },
        children: [
          { path: 'overview', name: 'DataCenterOverview', component: () => import('@/views/data-center/overview.vue'), meta: { title: '数据概览', icon: 'DataLine', menuLevel: 2 } },
          { path: 'modules', name: 'DataCenterModules', component: () => import('@/views/data-center/modules.vue'), meta: { title: '模块数据', icon: 'Menu', menuLevel: 2 } },
          { path: 'finance', name: 'DataCenterFinance', component: () => import('@/views/data-center/finance.vue'), meta: { title: '财务报表', icon: 'Money', menuLevel: 2 } },
          // 模块运营（原模块中心的功能）
          { path: 'panel', name: 'ModulePanel', component: () => import('@/views/module-center/index.vue'), meta: { title: '模块总控面板', icon: 'Menu', menuLevel: 2 } },
          { path: 'grayscale', name: 'ModuleGrayscale', component: () => import('@/views/module-center/grayscale.vue'), meta: { title: '灰度发布', icon: 'Connection', menuLevel: 2 } },
          { path: 'metrics', name: 'ModuleMetrics', component: () => import('@/views/module-center/metrics.vue'), meta: { title: '模块监控', icon: 'DataLine', menuLevel: 2 } }
        ]
      },

      // ===== 8. 设置中心 =====
      // 系统设置 + 中台服务（pay/im/material/risk/ai 等 12 个中台配置入口）
      {
        path: 'settings-center',
        name: 'SettingsCenter',
        redirect: '/settings-center/setting',
        meta: { title: '设置中心', icon: 'Setting', menuGroup: 'settings-center', menuLevel: 1 },
        children: [
          // 系统基础设置
          { path: 'setting', name: 'SettingsCenterSetting', component: () => import('@/views/setting/index.vue'), meta: { title: '系统设置', icon: 'Setting', permission: 'setting:read', menuLevel: 2 } },
          { path: 'role', name: 'SettingsCenterRole', component: () => import('@/views/role/index.vue'), meta: { title: '角色管理', icon: 'UserFilled', permission: 'role:read', menuLevel: 2 } },
          { path: 'permission', name: 'SettingsCenterPermission', component: () => import('@/views/permission/index.vue'), meta: { title: '权限列表', icon: 'Lock', permission: 'permission:read', menuLevel: 2 } },
          { path: 'region', name: 'SettingsCenterRegion', component: () => import('@/views/region/index.vue'), meta: { title: '地区管理', icon: 'Location', permission: 'region:read', menuLevel: 2 } },
          { path: 'file', name: 'SettingsCenterFile', component: () => import('@/views/file/index.vue'), meta: { title: '文件管理', icon: 'UploadFilled', permission: 'file:read', menuLevel: 2 } },

          // 中台服务分组（12 个中台配置入口）
          {
            path: 'middleware',
            name: 'SettingsMiddleware',
            redirect: '/settings-center/middleware/user',
            meta: { title: '中台服务', icon: 'Coin', menuLevel: 2 },
            children: [
              { path: 'user', name: 'MiddlewareUser', component: () => import('@/views/middleware-center/user.vue'), meta: { title: '用户中台', icon: 'User', menuLevel: 3 } },
              { path: 'pay', name: 'MiddlewarePay', component: () => import('@/views/middleware-center/pay.vue'), meta: { title: '支付中台', icon: 'Wallet', menuLevel: 3 } },
              { path: 'im', name: 'MiddlewareIm', component: () => import('@/views/middleware-center/im.vue'), meta: { title: 'IM中台', icon: 'ChatDotRound', menuLevel: 3 } },
              { path: 'merchant', name: 'MiddlewareMerchant', component: () => import('@/views/middleware-center/merchant.vue'), meta: { title: '商家中台', icon: 'Shop', menuLevel: 3 } },
              { path: 'distribution', name: 'MiddlewareDistribution', component: () => import('@/views/middleware-center/distribution.vue'), meta: { title: '分销中台', icon: 'Share', menuLevel: 3 } },
              { path: 'marketing', name: 'MiddlewareMarketing', component: () => import('@/views/middleware-center/marketing.vue'), meta: { title: '营销中台', icon: 'Present', menuLevel: 3 } },
              { path: 'risk', name: 'MiddlewareRisk', component: () => import('@/views/middleware-center/risk.vue'), meta: { title: '风控中台', icon: 'Warning', menuLevel: 3 } },
              { path: 'lbs', name: 'MiddlewareLbs', component: () => import('@/views/middleware-center/lbs.vue'), meta: { title: 'LBS中台', icon: 'Location', menuLevel: 3 } },
              { path: 'ai', name: 'MiddlewareAi', component: () => import('@/views/middleware-center/ai.vue'), meta: { title: 'AI中台', icon: 'MagicStick', menuLevel: 3 } },
              { path: 'tenant', name: 'MiddlewareTenant', component: () => import('@/views/middleware-center/tenant.vue'), meta: { title: '分站中台', icon: 'OfficeBuilding', menuLevel: 3 } },
              { path: 'material', name: 'MiddlewareMaterial', component: () => import('@/views/middleware-center/material.vue'), meta: { title: '素材中台', icon: 'Picture', menuLevel: 3 } },
              { path: 'diy', name: 'MiddlewareDiy', component: () => import('@/views/middleware-center/diy.vue'), meta: { title: 'DIY中台', icon: 'Brush', menuLevel: 3 } }
            ]
          }
        ]
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

  // 路由级权限校验：meta.permission 缺失则跳 403（super_admin/admin 直通）
  if (to.meta?.permission && !hasPermission(to.meta.permission)) {
    next({ path: '/403' })
    return
  }

  next()
})

export default router
