<template>
  <div class="error-page">
    <el-result icon="error" title="500" sub-title="抱歉，服务器开小差了，请稍后重试">
      <template #extra>
        <el-button type="primary" @click="goHome">返回首页</el-button>
        <el-button @click="reload">刷新重试</el-button>
      </template>
    </el-result>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()
const goHome = () => router.push('/')
const reload = () => window.location.reload()
</script>

<style scoped>
.error-page {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f0f2f5;
}
</style>
