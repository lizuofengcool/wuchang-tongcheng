<template>
  <PlaceholderView
    title="LBS 中台"
    description="地理位置服务，POI、行政区域、周边搜索"
    module="middleware-center/lbs"
    icon="Location"
    :features="features"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
const features = [
  'POI 兴趣点管理',
  '行政区域树与边界',
  '周边搜索 / 范围搜索',
  '高德 / 腾讯地图接入'
]
</script>
