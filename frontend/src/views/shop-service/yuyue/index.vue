<template>
  <PlaceholderView
    title="预约服务"
    description="五常同城预约服务模块，提供服务预约、时段管理、到店预约等功能"
    module="yuyue"
    icon="Calendar"
    :features="['服务预约', '时段管理', '到店预约', '上门服务', '预约提醒', '取消改约']"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
</script>
