<template>
  <router-view />
</template>

<script setup>
import { onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import router from '@/router'

const userStore = useUserStore()

// 应用启动时校验 token 是否有效：如果本地有 token 但已过期，立即清除并跳登录页
onMounted(async () => {
  if (userStore.token) {
    try {
      await userStore.fetchProfile()
    } catch (e) {
      // token 无效，清除状态
      userStore.logout()
      router.replace('/login')
    }
  }
})
</script>
