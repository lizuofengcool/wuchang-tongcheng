<template>
  <el-container class="main-layout">
    <!-- 侧边栏 -->
    <el-aside :width="isCollapse ? '64px' : '210px'" class="sidebar">
      <div class="logo">
        <el-icon :size="24"><Promotion /></el-icon>
        <span v-show="!isCollapse" class="logo-text">武昌同城</span>
      </div>
      <el-scrollbar>
        <el-menu
          :default-active="activeMenu"
          :collapse="isCollapse"
          :collapse-transition="false"
          router
          background-color="#001529"
          text-color="#bfcbd9"
          active-text-color="#409eff"
        >
          <template v-for="route in menuRoutes" :key="route.path">
            <el-menu-item :index="route.path">
              <el-icon v-if="route.meta?.icon">
                <component :is="route.meta.icon" />
              </el-icon>
              <template #title>{{ route.meta?.title }}</template>
            </el-menu-item>
          </template>
        </el-menu>
      </el-scrollbar>
    </el-aside>

    <el-container class="main-container">
      <!-- 顶部头部 -->
      <el-header class="header">
        <div class="header-left">
          <el-icon class="collapse-btn" :size="20" @click="isCollapse = !isCollapse">
            <Fold v-if="!isCollapse" />
            <Expand v-else />
          </el-icon>
          <el-breadcrumb separator="/" class="breadcrumb">
            <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item v-if="currentTitle && currentTitle !== '工作台'">
              {{ currentTitle }}
            </el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="header-right">
          <el-dropdown @command="handleCommand">
            <span class="user-info">
              <el-avatar :size="32" :src="userStore.avatar">
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

      <!-- 主区域 -->
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
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox, ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { constantRoutes } from '@/router'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const isCollapse = ref(false)

// 从根布局的 children 中筛选出菜单项（hidden 除外）
const menuRoutes = computed(() => {
  const root = constantRoutes.find((r) => r.path === '/')
  if (!root || !root.children) return []
  return root.children
    .filter((r) => !r.meta?.hidden)
    .map((r) => ({
      path: '/' + r.path,
      meta: r.meta
    }))
})

const activeMenu = computed(() => route.path)
const currentTitle = computed(() => route.meta?.title || '')

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
</script>

<style scoped>
.main-layout {
  height: 100vh;
}

.sidebar {
  background-color: #001529;
  transition: width 0.28s;
  overflow: hidden;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: #fff;
  background-color: #002140;
  overflow: hidden;
  white-space: nowrap;
}

.logo-text {
  font-size: 18px;
  font-weight: 600;
}

.sidebar :deep(.el-menu) {
  border-right: none;
}

.main-container {
  height: 100vh;
  overflow: hidden;
}

.header {
  background: #fff;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 20px;
  height: 60px;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
  z-index: 10;
}

.header-left {
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

.header-right {
  display: flex;
  align-items: center;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  outline: none;
}

.username {
  font-size: 14px;
  color: #303133;
}

.main {
  background-color: #f0f2f5;
  padding: 20px;
  overflow-y: auto;
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
