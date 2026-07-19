<template>
  <PlaceholderView
    title="团购商品"
    description="五常同城团购商品模块，提供团购商品、限时优惠、库存管理等功能"
    module="groupbuy"
    icon="Present"
    :features="['团购商品', '限时优惠', '库存管理', '销量统计', '上下架', '价格策略']"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
</script>
