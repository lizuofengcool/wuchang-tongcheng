<template>
  <PlaceholderView
    title="支付中台"
    description="统一支付网关，聚合微信/支付宝/银联"
    module="middleware-center/pay"
    icon="Wallet"
    :features="features"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
const features = [
  '聚合支付（微信/支付宝/银联）',
  '统一订单、退款、对账',
  '商户结算与分账',
  '支付风控规则配置'
]
</script>
