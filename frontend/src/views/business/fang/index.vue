<template>
  <PlaceholderView
    title="房屋租售"
    description="五常同城房屋租售模块，提供房源发布、地图找房、中介服务等功能"
    module="fang"
    icon="House"
    :features="['房源发布', '小区管理', '户型选择', '地图找房', '中介服务', '租售类型']"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
</script>
