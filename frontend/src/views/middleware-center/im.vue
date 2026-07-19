<template>
  <PlaceholderView
    title="IM 中台"
    description="即时通讯服务，支持单聊/群聊/客服会话"
    module="middleware-center/im"
    icon="ChatDotRound"
    :features="features"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
const features = [
  '单聊 / 群聊 / 客服会话',
  '消息历史与未读数',
  'WebSocket 长连接管理',
  '消息推送（APP/微信/短信）'
]
</script>
