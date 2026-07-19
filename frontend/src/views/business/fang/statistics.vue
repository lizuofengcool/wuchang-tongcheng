<template>
  <div class="app-container" v-loading="loading">
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ overview.total_houses || 0 }}</div><div class="stat-label">总房源</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ overview.today_new || 0 }}</div><div class="stat-label">今日新增</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">¥{{ formatAmount(overview.total_amount) }}</div><div class="stat-label">总成交额</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ overview.total_deals || 0 }}</div><div class="stat-label">成交量</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ overview.pending_audit || 0 }}</div><div class="stat-label">待审核</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-danger">{{ overview.violation || 0 }}</div><div class="stat-label">违规数</div></div></el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16">
      <el-col :xs="24" :md="16">
        <el-card class="chart-card">
          <template #header><span>30 天价格趋势</span></template>
          <div ref="trendChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="8">
        <el-card class="chart-card">
          <template #header><span>房源类型分布</span></template>
          <div ref="typeChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16">
      <el-col :xs="24" :md="12">
        <el-card class="chart-card">
          <template #header><span>城市分布 TOP 10</span></template>
          <div ref="cityChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="12">
        <el-card class="chart-card">
          <template #header><span>户型分布</span></template>
          <div ref="layoutChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, nextTick, onBeforeUnmount } from 'vue'
import * as echarts from 'echarts'
import { getOverviewStats, getPriceTrend } from '@/api/house'

const loading = ref(false)
const overview = reactive({ total_houses: 0, today_new: 0, total_amount: 0, total_deals: 0, pending_audit: 0, violation: 0 })

const trendChartRef = ref(null)
const typeChartRef = ref(null)
const cityChartRef = ref(null)
const layoutChartRef = ref(null)
let trendChart = null
let typeChart = null
let cityChart = null
let layoutChart = null

const formatAmount = (n) => Number(n || 0).toFixed(2)

const loadOverview = async () => {
  try {
    const res = await getOverviewStats()
    Object.assign(overview, res.data || {})
  } catch (e) { /* ignore */ }
}

const loadPriceTrend = async () => {
  try {
    const res = await getPriceTrend({ days: 30 })
    const d = res.data || {}
    if (!trendChartRef.value) return
    if (!trendChart) trendChart = echarts.init(trendChartRef.value)
    trendChart.setOption({
      tooltip: { trigger: 'axis' },
      xAxis: { type: 'category', data: d.dates || [] },
      yAxis: { type: 'value', axisLabel: { formatter: '¥{value}' } },
      series: [
        { name: '售价', type: 'line', data: d.sale_prices || [], smooth: true, itemStyle: { color: '#409eff' } },
        { name: '租金', type: 'line', data: d.rent_prices || [], smooth: true, itemStyle: { color: '#e6a23c' } }
      ],
      legend: { data: ['售价', '租金'] },
      grid: { left: 50, right: 20, top: 40, bottom: 30 }
    })
  } catch (e) { /* ignore */ }
}

const renderTypeChart = () => {
  if (!typeChartRef.value) return
  if (!typeChart) typeChart = echarts.init(typeChartRef.value)
  typeChart.setOption({
    tooltip: { trigger: 'item' },
    legend: { orient: 'vertical', left: 'left' },
    series: [{
      name: '类型', type: 'pie', radius: '60%',
      data: [
        { value: 0, name: '出售' },
        { value: 0, name: '出租' }
      ]
    }]
  })
}

const renderCityChart = () => {
  if (!cityChartRef.value) return
  if (!cityChart) cityChart = echarts.init(cityChartRef.value)
  cityChart.setOption({
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: ['北京', '上海', '广州', '深圳', '杭州', '成都', '武汉', '南京', '苏州', '西安'] },
    yAxis: { type: 'value' },
    series: [{ name: '房源数', type: 'bar', data: [0, 0, 0, 0, 0, 0, 0, 0, 0, 0], itemStyle: { color: '#409eff' } }],
    grid: { left: 50, right: 20, top: 30, bottom: 30 }
  })
}

const renderLayoutChart = () => {
  if (!layoutChartRef.value) return
  if (!layoutChart) layoutChart = echarts.init(layoutChartRef.value)
  layoutChart.setOption({
    tooltip: { trigger: 'item' },
    legend: { orient: 'vertical', left: 'left' },
    series: [{
      name: '户型', type: 'pie', radius: ['40%', '70%'],
      data: [
        { value: 0, name: '一室' },
        { value: 0, name: '二室' },
        { value: 0, name: '三室' },
        { value: 0, name: '四室+' }
      ]
    }]
  })
}

const handleResize = () => {
  trendChart?.resize()
  typeChart?.resize()
  cityChart?.resize()
  layoutChart?.resize()
}

onMounted(async () => {
  loading.value = true
  try {
    await loadOverview()
  } finally {
    loading.value = false
  }
  await nextTick()
  await loadPriceTrend()
  renderTypeChart()
  renderCityChart()
  renderLayoutChart()
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  trendChart?.dispose()
  typeChart?.dispose()
  cityChart?.dispose()
  layoutChart?.dispose()
})
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-item { text-align: center; padding: 16px; }
.stat-value { font-size: 22px; font-weight: 600; color: #303133; }
.text-success { color: #67c23a; }
.text-warning { color: #e6a23c; }
.text-danger { color: #f56c6c; }
.text-primary { color: #409eff; }
.chart-card { margin-bottom: 16px; }
.chart-container { width: 100%; height: 320px; }
</style>
