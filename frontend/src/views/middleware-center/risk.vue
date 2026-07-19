<template>
  <PlaceholderView
    title="风控中台"
    description="内容审核、行为风控、反作弊"
    module="middleware-center/risk"
    icon="Warning"
    :features="features"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
const features = [
  '内容审核（文本/图片/视频）',
  '用户行为风控规则',
  '反作弊与黑名单',
  '风控事件中心'
]
</script>
