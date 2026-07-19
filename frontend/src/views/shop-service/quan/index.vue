<template>
  <PlaceholderView
    title="同城圈子"
    description="五常同城圈子模块，提供圈子创建、帖子发布、评论回复等功能"
    module="quan"
    icon="ChatLineRound"
    :features="['圈子创建', '帖子发布', '评论回复', '置顶加精', '成员管理', '圈子动态']"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
</script>
