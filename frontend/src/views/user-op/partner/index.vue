<template>
  <PlaceholderView
    title="同城合伙人"
    description="五常同城合伙人模块，提供合伙人等级、分销佣金、提现管理等功能"
    module="partner"
    icon="UserFilled"
    :features="['合伙人等级', '分销佣金', '提现管理', '业绩统计', '邀请奖励', '结算记录']"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
</script>
