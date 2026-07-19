<template>
  <PlaceholderView
    title="会员卡"
    description="五常同城会员卡模块，提供会员等级、权益管理、续费管理等功能"
    module="vipcard"
    icon="CreditCard"
    :features="['会员等级', '权益管理', '黑卡会员', '折扣设置', '会员统计', '续费管理']"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
</script>
