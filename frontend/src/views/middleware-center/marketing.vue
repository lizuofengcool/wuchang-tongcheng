<template>
  <PlaceholderView
    title="营销中台"
    description="优惠券、拼团、砍价、抽奖等营销活动"
    module="middleware-center/marketing"
    icon="Present"
    :features="features"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
const features = [
  '优惠券 / 满减 / 折扣',
  '拼团 / 砍价 / 秒杀',
  '抽奖 / 红包 / 积分',
  '活动效果分析'
]
</script>
