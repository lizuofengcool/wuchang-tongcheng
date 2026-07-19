<template>
  <PlaceholderView
    title="直播管理"
    description="五常同城直播管理模块，提供开播申请、直播列表、弹幕管理等功能"
    module="zhibo"
    icon="VideoCamera"
    :features="['开播申请', '直播列表', '弹幕管理', '礼物统计', '直播回放', '违规处理']"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
</script>
