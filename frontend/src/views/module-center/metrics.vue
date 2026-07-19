<template>
  <PlaceholderView
    title="模块监控"
    description="模块运行状态监控、QPS、错误率、延迟等指标"
    module="module-center/metrics"
    icon="DataLine"
    :features="features"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
const features = [
  '模块 QPS / 错误率 / 延迟实时监控',
  '模块依赖关系拓扑图',
  '异常告警与历史趋势',
  '与 Prometheus / Grafana 联动'
]
</script>
