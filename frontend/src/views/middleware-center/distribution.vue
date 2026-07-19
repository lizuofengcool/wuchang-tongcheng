<template>
  <PlaceholderView
    title="分销中台"
    description="分销商、推广员、佣金结算"
    module="middleware-center/distribution"
    icon="Share"
    :features="features"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
const features = [
  '分销商 / 推广员管理',
  '多级分销佣金规则',
  '推广链接与二维码',
  '佣金结算与提现'
]
</script>
