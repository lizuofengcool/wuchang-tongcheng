<template>
  <PlaceholderView
    title="商家中台"
    description="统一商家入驻、店铺、商品、订单管理"
    module="middleware-center/merchant"
    icon="Shop"
    :features="features"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
const features = [
  '商家入驻与资质审核',
  '店铺装修、商品中心、订单中心',
  '商家结算与对账',
  '商家评分与等级体系'
]
</script>
