<template>
  <PlaceholderView
    title="优惠券系统"
    description="五常同城优惠券系统模块，提供券模板管理、领取记录、核销统计等功能"
    module="coupon"
    icon="Ticket"
    :features="['券模板管理', '领取记录', '核销统计', '使用规则', '优惠券类型', '限量发放']"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
</script>
