<template>
  <PlaceholderView
    title="分站中台"
    description="分站配置、独立域名、独立品牌"
    module="middleware-center/tenant"
    icon="OfficeBuilding"
    :features="features"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
const features = [
  '分站创建与配置',
  '独立品牌 / 域名 / Logo',
  '分站数据隔离',
  '分站运营数据对比'
]
</script>
