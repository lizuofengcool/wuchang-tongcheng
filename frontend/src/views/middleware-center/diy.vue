<template>
  <PlaceholderView
    title="DIY 中台"
    description="可视化页面装修、组件库、模板管理"
    module="middleware-center/diy"
    icon="Brush"
    :features="features"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
const features = [
  '可视化拖拽页面装修',
  '组件库与模板',
  '首页 / 商家页 / 专题页',
  '页面发布与回滚'
]
</script>
