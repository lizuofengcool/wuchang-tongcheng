<template>
  <PlaceholderView
    title="用户中台"
    description="统一用户账号、登录、认证、资料管理"
    module="middleware-center/user"
    icon="User"
    :features="features"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
const features = [
  '统一账号体系（手机号/邮箱/第三方登录）',
  '用户资料、实名认证、收货地址',
  'Token 颁发与续签',
  '用户标签与分群'
]
</script>
