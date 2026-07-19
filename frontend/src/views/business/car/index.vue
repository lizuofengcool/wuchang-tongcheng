<template>
  <PlaceholderView
    title="车辆买卖"
    description="五常同城车辆买卖模块，提供车辆发布、品牌车系、估价评估等功能"
    module="car"
    icon="Van"
    :features="['车辆发布', '品牌车系', '年份里程', '估价评估', '图片展示', '联系卖家']"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
</script>
