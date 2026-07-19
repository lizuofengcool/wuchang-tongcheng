<template>
  <PlaceholderView
    title="数据概览"
    description="平台全局数据概览：用户/订单/营收/活跃"
    module="data-center/overview"
    icon="DataLine"
    :features="features"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
const features = [
  '用户数 / 活跃 / 留存',
  '订单数 / GMV / 客单价',
  '营收 / 成本 / 利润',
  '模块活跃度排行'
]
</script>
