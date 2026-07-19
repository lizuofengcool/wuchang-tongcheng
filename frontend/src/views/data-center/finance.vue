<template>
  <PlaceholderView
    title="财务报表"
    description="平台财务收支、结算、对账"
    module="data-center/finance"
    icon="Money"
    :features="features"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
const features = [
  '收支明细报表',
  '商家结算对账',
  '提现与分账记录',
  '财务周期汇总'
]
</script>
