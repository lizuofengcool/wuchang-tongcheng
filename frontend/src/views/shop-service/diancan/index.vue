<template>
  <PlaceholderView
    title="同城点餐"
    description="五常同城点餐模块，提供菜单管理、堂食点餐、外卖配送等功能"
    module="diancan"
    icon="Food"
    :features="['菜单管理', '堂食点餐', '外卖配送', '桌台管理', '在线支付', '订单排队']"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
</script>
