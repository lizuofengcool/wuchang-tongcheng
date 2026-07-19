<template>
  <PlaceholderView
    title="拼车出行"
    description="五常同城拼车出行模块，提供行程发布、路线匹配、顺风车等功能"
    module="pinche"
    icon="Van"
    :features="['行程发布', '路线匹配', '顺风车', '包车服务', '费用分摊', '安全认证']"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
</script>
