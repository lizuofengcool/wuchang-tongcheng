<template>
  <PlaceholderView
    title="AI 中台"
    description="AI 能力中心，文生图、智能客服、内容生成"
    module="middleware-center/ai"
    icon="MagicStick"
    :features="features"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
const features = [
  '智能客服 / 问答',
  '内容生成（推文 / 文案 / 描述）',
  '图片生成与识别',
  'AI 调用日志与计费'
]
</script>
