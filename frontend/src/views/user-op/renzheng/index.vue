<template>
  <PlaceholderView
    title="用户认证"
    description="五常同城用户认证模块，提供实名认证、企业认证、资质审核等功能"
    module="renzheng"
    icon="Postcard"
    :features="['实名认证', '企业认证', '个人认证', '资质审核', '认证状态', '审核记录']"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
</script>
