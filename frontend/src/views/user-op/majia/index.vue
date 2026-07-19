<template>
  <PlaceholderView
    title="马甲管理"
    description="五常同城马甲管理模块，提供多账号管理、匿名发帖、切换账号等功能"
    module="majia"
    icon="Avatar"
    :features="['多账号管理', '匿名发帖', '切换账号', '马甲分组', '使用记录', '权限控制']"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
</script>
