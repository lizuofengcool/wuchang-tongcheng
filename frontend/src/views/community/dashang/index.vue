<template>
  <PlaceholderView
    title="打赏记录"
    description="五常同城打赏记录模块，提供打赏发起、答谢记录、金额统计等功能"
    module="dashang"
    icon="Coin"
    :features="['打赏发起', '答谢记录', '积分打赏', '金额统计', '打赏排行', '收支明细']"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
</script>
