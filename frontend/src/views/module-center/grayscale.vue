<template>
  <PlaceholderView
    title="灰度发布"
    description="模块灰度发布管理，支持按地区/用户/比例灰度"
    module="module-center/grayscale"
    icon="Connection"
    :features="features"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
const features = [
  '创建灰度发布任务（按地区/用户/比例）',
  '灰度版本管理、回滚',
  '实时监控灰度覆盖率',
  '一键全量发布或终止'
]
</script>
