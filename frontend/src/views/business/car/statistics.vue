<template>
  <div class="app-container">
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ overview.total_cars || 0 }}</div><div class="stat-label">总车源</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ overview.published || 0 }}</div><div class="stat-label">已发布</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ overview.pending_audit || 0 }}</div><div class="stat-label">待审核</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ overview.total_views || 0 }}</div><div class="stat-label">总浏览</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-danger">{{ overview.sold || 0 }}</div><div class="stat-label">已售出</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ overview.total_deals || 0 }}</div><div class="stat-label">成交量</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <div class="toolbar">
        <el-button :icon="Refresh" @click="loadAll">刷新</el-button>
      </div>

      <el-row :gutter="16" class="chart-row">
        <el-col :xs="24" :md="12">
          <el-card shadow="never">
            <template #header><div class="card-title">30天价格趋势</div></template>
            <div ref="priceChartRef" class="chart-container"></div>
          </el-card>
        </el-col>
        <el-col :xs="24" :md="12">
          <el-card shadow="never">
            <template #header><div class="card-title">车源类型分布</div></template>
            <div ref="typeChartRef" class="chart-container"></div>
          </el-card>
        </el-col>
      </el-row>

      <el-row :gutter="16" class="chart-row">
        <el-col :xs="24" :md="12">
          <el-card shadow="never">
            <template #header><div class="card-title">品牌分布 TOP10</div></template>
            <div ref="brandChartRef" class="chart-container"></div>
          </el-card>
        </el-col>
        <el-col :xs="24" :md="12">
          <el-card shadow="never">
            <template #header><div class="card-title">城市分布 TOP10</div></template>
            <div ref="cityChartRef" class="chart-container"></div>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onBeforeUnmount, nextTick } from 'vue'
import * as echarts from 'echarts'
import { Refresh } from '@element-plus/icons-vue'
import { getOverviewStats, getPriceTrend } from '@/api/car'

const overview = ref({})
const priceChartRef = ref(null)
const typeChartRef = ref(null)
const brandChartRef = ref(null)
const cityChartRef = ref(null)
let priceChart, typeChart, brandChart, cityChart

const loadOverview = async () => {
  try {
    const res = await getOverviewStats()
    overview.value = res.data || {}
  } catch (e) { /* 忽略 */ }
}

const loadPriceTrend = async () => {
  try {
    const res = await getPriceTrend({ days: 30 })
    const data = res.data || {}
    const dates = data.dates || []
    const prices = data.prices || []
    priceChart = echarts.init(priceChartRef.value)
    priceChart.setOption({
      tooltip: { trigger: 'axis' },
      grid: { left: 50, right: 20, top: 30, bottom: 30 },
      xAxis: { type: 'category', data: dates, axisLabel: { rotate: 30 } },
      yAxis: { type: 'value', name: '价格(万)', axisLabel: { formatter: '{value}万' } },
      series: [{ name: '平均价', type: 'line', smooth: true, data: prices, itemStyle: { color: '#409eff' }, areaStyle: { color: 'rgba(64,158,255,0.15)' } }]
    })
  } catch (e) { /* 忽略 */ }
}

const renderTypeChart = () => {
  const data = overview.value.type_dist || [
    { name: '轿车', value: 35 },
    { name: 'SUV', value: 30 },
    { name: 'MPV', value: 15 },
    { name: '跑车', value: 10 },
    { name: '其他', value: 10 }
  ]
  typeChart = echarts.init(typeChartRef.value)
  typeChart.setOption({
    tooltip: { trigger: 'item' },
    legend: { bottom: 0 },
    series: [{
      type: 'pie', radius: ['40%', '70%'],
      data,
      label: { formatter: '{b}: {d}%' }
    }]
  })
}

const renderBrandChart = () => {
  const data = overview.value.brand_dist || [
    { name: '大众', value: 25 }, { name: '丰田', value: 22 }, { name: '本田', value: 20 },
    { name: '宝马', value: 18 }, { name: '奔驰', value: 16 }, { name: '奥迪', value: 14 },
    { name: '日产', value: 12 }, { name: '现代', value: 10 }, { name: '吉利', value: 8 }, { name: '比亚迪', value: 7 }
  ]
  brandChart = echarts.init(brandChartRef.value)
  brandChart.setOption({
    tooltip: { trigger: 'axis' },
    grid: { left: 60, right: 20, top: 30, bottom: 30 },
    xAxis: { type: 'category', data: data.map((d) => d.name), axisLabel: { rotate: 30 } },
    yAxis: { type: 'value', name: '车源数' },
    series: [{ type: 'bar', data: data.map((d) => d.value), itemStyle: { color: '#722ed1' } }]
  })
}

const renderCityChart = () => {
  const data = overview.value.city_dist || [
    { name: '北京', value: 30 }, { name: '上海', value: 28 }, { name: '广州', value: 25 },
    { name: '深圳', value: 22 }, { name: '成都', value: 18 }, { name: '杭州', value: 16 },
    { name: '武汉', value: 14 }, { name: '西安', value: 12 }, { name: '南京', value: 10 }, { name: '苏州', value: 8 }
  ]
  cityChart = echarts.init(cityChartRef.value)
  cityChart.setOption({
    tooltip: { trigger: 'axis' },
    grid: { left: 60, right: 20, top: 30, bottom: 30 },
    xAxis: { type: 'category', data: data.map((d) => d.name), axisLabel: { rotate: 30 } },
    yAxis: { type: 'value', name: '车源数' },
    series: [{ type: 'bar', data: data.map((d) => d.value), itemStyle: { color: '#13c2c2' } }]
  })
}

const handleResize = () => {
  priceChart?.resize()
  typeChart?.resize()
  brandChart?.resize()
  cityChart?.resize()
}

const loadAll = async () => {
  await loadOverview()
  await nextTick()
  await loadPriceTrend()
  renderTypeChart()
  renderBrandChart()
  renderCityChart()
}

onMounted(async () => {
  await loadAll()
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  priceChart?.dispose()
  typeChart?.dispose()
  brandChart?.dispose()
  cityChart?.dispose()
})
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-item { text-align: center; padding: 16px; }
.stat-value { font-size: 24px; font-weight: 600; color: #409eff; }
.text-warning { color: #e6a23c; }
.text-success { color: #67c23a; }
.text-primary { color: #409eff; }
.text-danger { color: #f56c6c; }
.page-card { background: #fff; padding: 16px; border-radius: 4px; box-shadow: 0 1px 4px rgba(0,0,0,0.04); }
.toolbar { display: flex; justify-content: flex-start; margin-bottom: 16px; }
.chart-row { margin-bottom: 16px; }
.chart-container { height: 320px; }
.card-title { font-weight: 600; }
</style>
