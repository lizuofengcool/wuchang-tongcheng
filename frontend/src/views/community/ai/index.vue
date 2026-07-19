<template>
  <PlaceholderView
    title="AI助手"
    description="五常同城AI助手模块，提供智能问答、个性化推荐、写作助手等功能"
    module="ai"
    icon="MagicStick"
    :features="['智能问答', '个性化推荐', '写作助手', '内容审核', '标签生成', '数据分析']"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
</script>
