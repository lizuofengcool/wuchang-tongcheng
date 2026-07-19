<template>
  <el-container class="main-layout">
    <!-- 顶部导航栏：Logo + 5 大中心 Tab + 用户区 -->
    <el-header class="top-header">
      <div class="header-left">
        <div class="logo" @click="router.push('/workspace')">
          <el-icon :size="22"><Promotion /></el-icon>
          <span class="logo-text">近享同城</span>
        </div>
        <div class="top-tabs">
          <div
            v-for="item in topMenus"
            :key="item.path"
            class="top-tab-item"
            :class="{ active: currentTopPath === item.path }"
            @click="onTopMenuSelect(item.path)"
          >
            <el-icon v-if="item.icon" :size="16"><component :is="item.icon" /></el-icon>
            <span>{{ item.title }}</span>
          </div>
        </div>
      </div>
      <div class="header-right">
        <!-- 地区选择器 -->
        <el-tree-select
          v-model="regionStore.currentRegionId"
          :data="regionStore.regionTree"
          :props="{ value: 'id', label: 'name', children: 'children' }"
          check-strictly
          node-key="id"
          placeholder="选择地区"
          size="small"
          class="region-select"
          @change="onRegionChange"
        >
          <template #prefix>
            <el-icon><Location /></el-icon>
          </template>
        </el-tree-select>
        <!-- 消息通知 -->
        <el-badge
          :value="unreadCount"
          :hidden="unreadCount === 0"
          :max="99"
          class="message-badge"
        >
          <el-icon
            class="message-icon"
            :size="18"
            @click="router.push('/workspace/message')"
          >
            <Bell />
          </el-icon>
        </el-badge>
        <el-dropdown @command="handleCommand">
          <span class="user-info">
            <el-avatar :size="28" :src="userStore.avatar">
              {{ userStore.nickname.charAt(0) }}
            </el-avatar>
            <span class="username">{{ userStore.nickname }}</span>
            <el-icon><ArrowDown /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="profile">
                <el-icon><UserFilled /></el-icon>个人中心
              </el-dropdown-item>
              <el-dropdown-item command="logout" divided>
                <el-icon><SwitchButton /></el-icon>退出登录
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </el-header>

    <!-- 下方：侧边栏 + 主内容区 -->
    <el-container class="bottom-container">
      <el-aside :width="isCollapse ? '64px' : '220px'" class="sidebar">
        <el-scrollbar class="menu-scrollbar">
          <el-menu
            :default-active="activeMenu"
            :default-openeds="defaultOpeneds"
            :collapse="isCollapse"
            :collapse-transition="false"
            :unique-opened="false"
            router
            background-color="#001529"
            text-color="#bfcbd9"
            active-text-color="#409eff"
          >
            <SidebarItem
              v-for="item in sidebarMenus"
              :key="item.path"
              :item="item"
              :base-path="currentTopPath"
            />
          </el-menu>
        </el-scrollbar>
      </el-aside>

      <el-container class="main-container">
        <!-- 二级头部：折叠按钮 + 面包屑 -->
        <el-header class="sub-header">
          <div class="sub-header-left">
            <el-icon class="collapse-btn" :size="18" @click="isCollapse = !isCollapse">
              <Fold v-if="!isCollapse" />
              <Expand v-else />
            </el-icon>
            <el-breadcrumb separator="/" class="breadcrumb">
              <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
              <el-breadcrumb-item v-for="b in breadcrumbItems" :key="b.path">
                {{ b.title }}
              </el-breadcrumb-item>
            </el-breadcrumb>
          </div>
        </el-header>

        <!-- 主内容区 -->
        <el-main class="main">
          <router-view v-slot="{ Component }">
            <transition name="fade-transform" mode="out-in">
              <keep-alive>
                <component :is="Component" />
              </keep-alive>
            </transition>
          </router-view>
        </el-main>
      </el-container>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, h, resolveComponent } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox, ElMessage, ElIcon, ElSubMenu, ElMenuItem } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { useRegionStore } from '@/stores/region'
import { getUnreadCount } from '@/api/news'
import { constantRoutes } from '@/router'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const regionStore = useRegionStore()

const isCollapse = ref(false)

// ====== 未读消息数轮询 ======
const unreadCount = ref(0)
let unreadTimer = null

const loadUnreadCount = async () => {
  if (!userStore.isLoggedIn) return
  try {
    const res = await getUnreadCount()
    unreadCount.value = res.data?.count || 0
  } catch (e) {
    // 静默：未登录或网络异常不打扰用户
  }
}

// 路由切换到消息中心时刷新未读数
const stopRouteWatch = router.afterEach((to) => {
  if (to.path === '/workspace/message') {
    setTimeout(loadUnreadCount, 800)
  }
})

// ====== 一级菜单（顶部 Tab：5 大中心） ======
// 从 constantRoutes 派生：工作台 / 模块中心 / 中台中心 / 设置中心 / 数据中心
const rootRoute = computed(() => constantRoutes.find((r) => r.path === '/'))

const topMenus = computed(() => {
  if (!rootRoute.value || !rootRoute.value.children) return []
  return rootRoute.value.children
    .filter((r) => !r.meta?.hidden && r.meta?.menuGroup)
    .map((r) => ({
      path: '/' + r.path,
      title: r.meta?.title,
      icon: r.meta?.icon,
      raw: r
    }))
})

// ====== 当前激活的一级 Tab 路径 ======
const currentTopPath = computed(() => {
  if (route.matched.length >= 2) {
    return route.matched[1].path
  }
  return '/workspace'
})

// ====== 侧边栏菜单（当前一级 Tab 下的所有子菜单，递归构建） ======
const sidebarMenus = computed(() => {
  const current = topMenus.value.find((m) => m.path === currentTopPath.value)
  if (!current || !current.raw?.children) return []
  return buildMenuTree(current.raw.children, current.path)
})

// 递归构建菜单树
function buildMenuTree(children, basePath) {
  if (!Array.isArray(children)) return []
  return children
    .filter((c) => !c.meta?.hidden)
    .map((c) => {
      const path = basePath + '/' + c.path
      const node = {
        path,
        title: c.meta?.title,
        icon: c.meta?.icon,
        children: []
      }
      if (c.children && c.children.length) {
        node.children = buildMenuTree(c.children, path)
      }
      return node
    })
}

// 默认展开第一层（让用户直接看到分组下的子菜单）
const defaultOpeneds = computed(() => sidebarMenus.value.map((m) => m.path))

// ====== 顶部 Tab 切换 ======
const onTopMenuSelect = (path) => {
  router.push(path)
}

// ====== 面包屑（按 matched 路由构造） ======
const breadcrumbItems = computed(() => {
  return route.matched
    .filter((r) => r.meta?.title && r.meta.title !== '工作台')
    .map((r) => ({ path: r.path, title: r.meta.title }))
})

const activeMenu = computed(() => route.path)

const onRegionChange = (val) => {
  regionStore.setCurrentRegion(val)
  ElMessage.success(`已切换地区：${regionStore.currentRegionName}`)
}

const handleCommand = async (cmd) => {
  if (cmd === 'profile') {
    router.push('/profile')
  } else if (cmd === 'logout') {
    try {
      await ElMessageBox.confirm('确定退出登录吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
      userStore.logout()
      ElMessage.success('已退出登录')
      router.push('/login')
    } catch (e) {
      // 取消
    }
  }
}

onMounted(() => {
  if (!regionStore.loaded) {
    regionStore.loadTree()
  }
  loadUnreadCount()
  unreadTimer = setInterval(loadUnreadCount, 60000)
})

onBeforeUnmount(() => {
  if (unreadTimer) {
    clearInterval(unreadTimer)
    unreadTimer = null
  }
  if (typeof stopRouteWatch === 'function') {
    stopRouteWatch()
  }
})

// ====== 递归菜单组件（内联定义） ======
const SidebarItem = {
  name: 'SidebarItem',
  props: {
    item: { type: Object, required: true },
    basePath: { type: String, default: '' }
  },
  setup(props) {
    return () => {
      const { item } = props
      const hasChildren = item.children && item.children.length > 0

      const iconNode = item.icon
        ? h(ElIcon, null, { default: () => h(resolveComponent(item.icon)) })
        : null

      if (hasChildren) {
        return h(ElSubMenu, { index: item.path }, {
          title: () => [iconNode, h('span', null, item.title)],
          default: () => item.children.map((child) =>
            h(SidebarItem, { item: child, key: child.path, 'base-path': item.path })
          )
        })
      }
      return h(ElMenuItem, { index: item.path }, {
        default: () => [iconNode, h('span', null, item.title)]
      })
    }
  }
}
</script>

<style scoped>
.main-layout {
  height: 100vh;
  flex-direction: column;
}

/* ====== 顶部导航栏 ====== */
.top-header {
  height: 60px;
  background-color: #001529;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 20px 0 0;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
  z-index: 10;
}

.header-left {
  display: flex;
  align-items: center;
  height: 100%;
  flex: 1;
  min-width: 0;
}

.logo {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #fff;
  padding: 0 24px 0 20px;
  height: 100%;
  cursor: pointer;
  background-color: #002140;
  flex-shrink: 0;
}

.logo-text {
  font-size: 18px;
  font-weight: 600;
  white-space: nowrap;
}

/* 顶部一级 Tab（5 大中心） */
.top-tabs {
  flex: 1;
  height: 60px;
  display: flex;
  align-items: stretch;
  min-width: 0;
  overflow-x: auto;
  overflow-y: hidden;
  scrollbar-width: thin;
}

.top-tabs::-webkit-scrollbar {
  height: 4px;
}

.top-tabs::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 2px;
}

.top-tab-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 0 24px;
  height: 60px;
  color: #bfcbd9;
  cursor: pointer;
  font-size: 14px;
  white-space: nowrap;
  flex-shrink: 0;
  border-bottom: 2px solid transparent;
  transition: all 0.2s;
}

.top-tab-item:hover {
  color: #fff;
  background-color: #002140;
}

.top-tab-item.active {
  color: #409eff;
  border-bottom-color: #409eff;
  background-color: #002140;
}

.top-tab-item .el-icon {
  margin-bottom: 2px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-shrink: 0;
}

.region-select {
  width: 160px;
}

.message-badge {
  display: inline-flex;
  align-items: center;
}

.message-icon {
  cursor: pointer;
  color: #bfcbd9;
  padding: 4px;
  border-radius: 4px;
  transition: all 0.2s;
}

.message-icon:hover {
  color: #409eff;
  background: rgba(64, 158, 255, 0.1);
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  outline: none;
  color: #bfcbd9;
}

.username {
  font-size: 14px;
  color: #bfcbd9;
}

/* ====== 下方容器 ====== */
.bottom-container {
  height: calc(100vh - 60px);
  overflow: hidden;
}

.sidebar {
  background-color: #001529;
  transition: width 0.28s;
  overflow: hidden;
}

.menu-scrollbar {
  height: 100%;
}

.sidebar :deep(.el-menu) {
  border-right: none;
}

.main-container {
  height: 100%;
  overflow: hidden;
  flex-direction: column;
}

/* ====== 二级头部（折叠按钮 + 面包屑） ====== */
.sub-header {
  background: #fff;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 20px;
  height: 50px;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.05);
  z-index: 5;
}

.sub-header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.collapse-btn {
  cursor: pointer;
  color: #5a5e66;
}

.collapse-btn:hover {
  color: #409eff;
}

.main {
  background-color: #f0f2f5;
  padding: 20px;
  overflow-y: auto;
  flex: 1;
}

/* 路由切换动画 */
.fade-transform-enter-active,
.fade-transform-leave-active {
  transition: all 0.3s;
}

.fade-transform-enter-from {
  opacity: 0;
  transform: translateX(-20px);
}

.fade-transform-leave-to {
  opacity: 0;
  transform: translateX(20px);
}
</style>
