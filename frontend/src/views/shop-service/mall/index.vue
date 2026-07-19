<template>
  <PlaceholderView
    title="同城商城"
    description="五常同城商城模块，提供商品管理、购物车、订单管理等功能"
    module="mall"
    icon="ShoppingCart"
    :features="['商品管理', 'SKU规格', '购物车', '订单管理', '支付结算', '物流配送']"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
</script>
