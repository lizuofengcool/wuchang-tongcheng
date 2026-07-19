<template>
  <PlaceholderView
    title="名片管理"
    description="五常同城名片管理模块，提供名片创建、名片交换、收藏名片等功能"
    module="mingpian"
    icon="Postcard"
    :features="['名片创建', '名片交换', '收藏名片', '分享名片', '名片夹', '二维码']"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
</script>
