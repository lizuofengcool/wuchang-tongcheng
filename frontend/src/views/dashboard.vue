<template>
  <div class="dashboard app-container">
    <el-row :gutter="20">
      <el-col :span="24">
        <el-card shadow="never" class="welcome-card">
          <div class="welcome">
            <el-avatar :size="64" :src="userStore.avatar">
              {{ userStore.nickname.charAt(0) }}
            </el-avatar>
            <div class="welcome-text">
              <h2>欢迎回来，{{ userStore.nickname }}</h2>
              <p>今天是 {{ today }}，祝您工作愉快！</p>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="stat-row">
      <el-col :xs="12" :sm="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #409eff">
            <el-icon :size="28"><User /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.users }}</div>
            <div class="stat-label">用户总数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #67c23a">
            <el-icon :size="28"><UserFilled /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.roles }}</div>
            <div class="stat-label">角色数量</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #e6a23c">
            <el-icon :size="28"><Lock /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.permissions }}</div>
            <div class="stat-label">权限数量</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #f56c6c">
            <el-icon :size="28"><Bell /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.myPermissions }}</div>
            <div class="stat-label">我的权限数</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="stat-row">
      <el-col :span="24">
        <el-card shadow="never">
          <template #header>
            <div class="card-header">
              <span>系统信息</span>
            </div>
          </template>
          <el-descriptions :column="2" border>
            <el-descriptions-item label="系统名称">武昌同城本地生活服务平台</el-descriptions-item>
            <el-descriptions-item label="管理后台版本">v1.0.0</el-descriptions-item>
            <el-descriptions-item label="后端技术栈">Go + Gin + GORM + PostgreSQL</el-descriptions-item>
            <el-descriptions-item label="前端技术栈">Vue3 + Vite + Element Plus + Pinia</el-descriptions-item>
            <el-descriptions-item label="权限模型">RBAC（用户-角色-权限）</el-descriptions-item>
            <el-descriptions-item label="当前登录账号">{{ userStore.userInfo?.username || '-' }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useUserStore } from '@/stores/user'
import { listRoles, listPermissions, myPermissions } from '@/api/permission'

const userStore = useUserStore()
const today = computed(() => {
  const d = new Date()
  const week = ['日', '一', '二', '三', '四', '五', '六']
  return `${d.getFullYear()}年${d.getMonth() + 1}月${d.getDate()}日 星期${week[d.getDay()]}`
})

const stats = ref({ users: 0, roles: 0, permissions: 0, myPermissions: 0 })

const loadStats = async () => {
  try {
    const [rolesRes, permsRes, myRes] = await Promise.all([
      listRoles(),
      listPermissions(),
      myPermissions().catch(() => ({ data: [] }))
    ])
    stats.value.roles = rolesRes.data?.length || 0
    stats.value.permissions = permsRes.data?.length || 0
    stats.value.myPermissions = myRes.data?.length || 0
  } catch (e) {
    // 忽略统计错误
  }
}

onMounted(loadStats)
</script>

<style scoped>
.welcome-card {
  margin-bottom: 20px;
}
.welcome {
  display: flex;
  align-items: center;
  gap: 20px;
}
.welcome-text h2 {
  margin-bottom: 8px;
  color: #303133;
}
.welcome-text p {
  color: #909399;
}
.stat-row {
  margin-bottom: 20px;
}
.stat-card {
  display: flex;
  align-items: center;
}
.stat-card :deep(.el-card__body) {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
}
.stat-icon {
  width: 56px;
  height: 56px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
}
.stat-content {
  flex: 1;
}
.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
}
.stat-label {
  font-size: 13px;
  color: #909399;
  margin-top: 4px;
}
.card-header {
  font-weight: 600;
}
</style>
