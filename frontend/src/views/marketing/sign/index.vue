<template>
  <PlaceholderView
    title="签到积分"
    description="五常同城签到积分模块，提供每日签到、积分记录、积分兑换等功能"
    module="sign"
    icon="Calendar"
    :features="['每日签到', '积分记录', '连续奖励', '积分兑换', '签到日历', '积分商城']"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
</script>
