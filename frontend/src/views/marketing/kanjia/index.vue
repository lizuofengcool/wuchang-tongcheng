<template>
  <PlaceholderView
    title="砍价活动"
    description="五常同城砍价活动模块，提供砍价商品、帮砍记录、底价设置等功能"
    module="kanjia"
    icon="Discount"
    :features="['砍价商品', '帮砍记录', '底价设置', '活动时间', '砍价日志', '下单转化']"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
</script>
