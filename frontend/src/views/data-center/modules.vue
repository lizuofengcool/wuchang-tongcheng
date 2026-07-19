<template>
  <PlaceholderView
    title="模块数据"
    description="按模块维度查看业务数据"
    module="data-center/modules"
    icon="Menu"
    :features="features"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
const features = [
  '模块维度数据对比',
  '模块营收 / 用户 / 订单',
  '模块间数据关联分析',
  '自定义指标看板'
]
</script>
