<template>
  <PlaceholderView
    title="素材中台"
    description="统一图片、视频、文档素材库"
    module="middleware-center/material"
    icon="Picture"
    :features="features"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
const features = [
  '图片 / 视频 / 文档统一管理',
  '素材分类与标签',
  '素材引用统计',
  'CDN 加速与压缩'
]
</script>
