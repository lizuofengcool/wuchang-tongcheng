<template>
  <PlaceholderView
    title="举报中心"
    description="五常同城举报中心模块，提供举报提交、举报处理、封禁管理等功能"
    module="jubao"
    icon="Warning"
    :features="['举报提交', '举报处理', '封禁管理', '申诉流程', '举报统计', '违规记录']"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
</script>
