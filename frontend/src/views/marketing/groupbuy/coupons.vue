<template>
  <PlaceholderView
    title="团购优惠券"
    description="五常同城团购优惠券模块，提供券模板、领取记录、核销统计等功能"
    module="groupbuy-coupons"
    icon="Ticket"
    :features="['券模板', '领取记录', '核销统计', '使用规则', '有效期', '限量发放']"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
</script>
