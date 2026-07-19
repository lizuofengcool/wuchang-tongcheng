<template>
  <PlaceholderView
    title="团购订单"
    description="五常同城团购订单模块，提供订单列表、核销管理、退款处理等功能"
    module="groupbuy-orders"
    icon="Tickets"
    :features="['订单列表', '核销管理', '退款处理', '订单状态', '消费统计', '导出报表']"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
</script>
