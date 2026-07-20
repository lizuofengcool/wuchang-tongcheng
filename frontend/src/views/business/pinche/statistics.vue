<template>
  <div class="app-container" v-loading="loading">
    <!-- 总览卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #409eff">
            <el-icon :size="22"><Van /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ overview.total_pinches }}</div>
            <div class="stat-label">总拼车数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #67c23a">
            <el-icon :size="22"><CirclePlus /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value text-success">{{ overview.today_new }}</div>
            <div class="stat-label">今日新增</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #722ed1">
            <el-icon :size="22"><Finished /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ overview.completed_trips }}</div>
            <div class="stat-label">已成行数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #13c2c2">
            <el-icon :size="22"><User /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value text-primary">¥{{ formatAmount(overview.total_amount) }}</div>
            <div class="stat-label">总成交额</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #e6a23c">
            <el-icon :size="22"><TrendCharts /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value text-warning">{{ (overview.completion_rate * 100).toFixed(1) }}%</div>
            <div class="stat-label">成行率</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #f56c6c">
            <el-icon :size="22"><Warning /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value text-danger">{{ (overview.refund_rate * 100).toFixed(1) }}%</div>
            <div class="stat-label">退款率</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 趋势图 + 热门路线 -->
    <el-row :gutter="16">
      <el-col :xs="24" :md="16">
        <el-card class="chart-card">
          <template #header>
            <div class="card-header-flex">
              <span>30 天发布趋势</span>
              <el-radio-group v-model="trendType" size="small" @change="loadTrend">
                <el-radio-button label="">全部</el-radio-button>
                <el-radio-button label="passenger">人找车</el-radio-button>
                <el-radio-button label="driver">车找人</el-radio-button>
                <el-radio-button label="cargo">车找货</el-radio-button>
                <el-radio-button label="cargo_need">货找车</el-radio-button>
              </el-radio-group>
            </div>
          </template>
          <div ref="trendChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="8">
        <el-card class="chart-card">
          <template #header><span>热门路线 TOP 20</span></template>
          <el-table :data="hotRoutes" border size="small" max-height="400">
            <el-table-column type="index" label="#" width="50" />
            <el-table-column label="路线" min-width="160">
              <template #default="{ row }">
                <div class="route-text">{{ row.origin }} → {{ row.destination }}</div>
              </template>
            </el-table-column>
            <el-table-column prop="pinch_count" label="拼车数" width="80" />
            <el-table-column prop="trip_count" label="成行" width="70" />
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <!-- 价格分布 + 转化漏斗 -->
    <el-row :gutter="16">
      <el-col :xs="24" :md="12">
        <el-card class="chart-card">
          <template #header><span>价格区间分布</span></template>
          <div ref="priceChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="12">
        <el-card class="chart-card">
          <template #header><span>转化漏斗</span></template>
          <div ref="funnelChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 类型分布 + 地区分布 -->
    <el-row :gutter="16">
      <el-col :xs="24" :md="12">
        <el-card class="chart-card">
          <template #header><span>拼车类型分布</span></template>
          <div ref="typeChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="12">
        <el-card class="chart-card">
          <template #header><span>地区分布</span></template>
          <div ref="regionChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 业务报表 -->
    <el-row :gutter="16">
      <el-col :span="24">
        <el-card class="chart-card">
          <template #header>
            <div class="card-header-flex">
              <span>业务概览报表</span>
              <el-button type="primary" size="small" @click="onExport">导出报表</el-button>
            </div>
          </template>
          <el-table :data="reportData" border size="small">
            <el-table-column prop="module" label="业务模块" width="160" />
            <el-table-column prop="total" label="总数" width="100" />
            <el-table-column prop="today" label="今日新增" width="100" />
            <el-table-column prop="week" label="本周新增" width="100" />
            <el-table-column prop="month" label="本月新增" width="100" />
            <el-table-column label="活跃率" width="100">
              <template #default="{ row }">{{ (row.active_rate * 100).toFixed(1) }}%</template>
            </el-table-column>
            <el-table-column label="总金额" width="120">
              <template #default="{ row }">¥{{ Number(row.amount || 0).toFixed(2) }}</template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, nextTick, onBeforeUnmount } from 'vue'
import { ElMessage } from 'element-plus'
import * as echarts from 'echarts'
import request from '@/utils/request'
import {
  Van, CirclePlus, Finished, User, TrendCharts, Warning
} from '@element-plus/icons-vue'

const loading = ref(false)

const overview = reactive({
  total_pinches: 0, today_new: 0, completed_trips: 0,
  total_amount: 0, completion_rate: 0, refund_rate: 0
})

const hotRoutes = ref([])
const reportData = ref([
  { module: '拼车发布', total: 0, today: 0, week: 0, month: 0, active_rate: 0, amount: 0 },
  { module: '路线管理', total: 0, today: 0, week: 0, month: 0, active_rate: 0, amount: 0 },
  { module: '行程管理', total: 0, today: 0, week: 0, month: 0, active_rate: 0, amount: 0 },
  { module: '预订管理', total: 0, today: 0, week: 0, month: 0, active_rate: 0, amount: 0 },
  { module: '车主认证', total: 0, today: 0, week: 0, month: 0, active_rate: 0, amount: 0 },
  { module: '车辆管理', total: 0, today: 0, week: 0, month: 0, active_rate: 0, amount: 0 },
  { module: '支付订单', total: 0, today: 0, week: 0, month: 0, active_rate: 0, amount: 0 },
  { module: '保险订单', total: 0, today: 0, week: 0, month: 0, active_rate: 0, amount: 0 }
])

const trendType = ref('')

// ===== 图表 refs =====
const trendChartRef = ref(null)
const priceChartRef = ref(null)
const funnelChartRef = ref(null)
const typeChartRef = ref(null)
const regionChartRef = ref(null)

let trendChart = null
let priceChart = null
let funnelChart = null
let typeChart = null
let regionChart = null

const formatAmount = (n) => Number(n || 0).toFixed(2)

// ===== 加载数据 =====
const loadOverview = async () => {
  try {
    const res = await request.get('/pinche/statistics/overview')
    Object.assign(overview, res.data || {})
  } catch (e) { /* ignore */ }
}

const loadTrend = async () => {
  try {
    const res = await request.get('/pinche/statistics/trend', {
      params: { type: trendType.value || undefined, days: 30 }
    })
    const d = res.data || {}
    renderTrendChart(d.dates || [], d.counts || [], d.amounts || [])
  } catch (e) {
    renderTrendChart([], [], [])
  }
}

const loadHotRoutes = async () => {
  try {
    const res = await request.get('/pinche/statistics/hot-routes', { params: { limit: 20 } })
    const d = res.data
    hotRoutes.value = d?.list || d || []
  } catch (e) {
    hotRoutes.value = []
  }
}

// ===== 渲染图表 =====
const renderTrendChart = (dates, counts, amounts) => {
  if (!trendChartRef.value) return
  if (!trendChart) {
    trendChart = echarts.init(trendChartRef.value)
  }
  trendChart.setOption({
    tooltip: { trigger: 'axis' },
    legend: { data: ['发布数', '成交额'] },
    xAxis: { type: 'category', data: dates },
    yAxis: [
      { type: 'value', name: '数量', position: 'left' },
      { type: 'value', name: '金额', position: 'right', axisLabel: { formatter: '¥{value}' } }
    ],
    series: [
      {
        name: '发布数', type: 'bar', data: counts,
        itemStyle: { color: '#409eff' }
      },
      {
        name: '成交额', type: 'line', yAxisIndex: 1, data: amounts, smooth: true,
        areaStyle: { opacity: 0.3 }, itemStyle: { color: '#67c23a' }
      }
    ],
    grid: { left: 50, right: 60, top: 40, bottom: 30 }
  })
}

const renderPriceChart = async () => {
  let data = []
  try {
    const res = await request.get('/pinche/statistics/price-distribution')
    data = res.data?.ranges || []
  } catch (e) { /* ignore */ }
  if (!data.length) {
    data = [
      { range: '0-20', count: 0 }, { range: '20-50', count: 0 },
      { range: '50-100', count: 0 }, { range: '100-200', count: 0 },
      { range: '200-500', count: 0 }, { range: '500+', count: 0 }
    ]
  }
  if (!priceChartRef.value) return
  if (!priceChart) {
    priceChart = echarts.init(priceChartRef.value)
  }
  priceChart.setOption({
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: data.map(i => i.range) },
    yAxis: { type: 'value' },
    series: [{
      name: '拼车数', type: 'bar', data: data.map(i => i.count),
      itemStyle: { color: '#67c23a' }
    }],
    grid: { left: 50, right: 20, top: 30, bottom: 30 }
  })
}

const renderFunnelChart = async () => {
  let data = []
  try {
    const res = await request.get('/pinche/statistics/funnel')
    data = res.data?.steps || []
  } catch (e) { /* ignore */ }
  if (!data.length) {
    data = [
      { value: 100, name: '发布' },
      { value: 60, name: '审核通过' },
      { value: 40, name: '预订' },
      { value: 30, name: '支付' },
      { value: 25, name: '成行' },
      { value: 23, name: '完成' }
    ]
  }
  if (!funnelChartRef.value) return
  if (!funnelChart) {
    funnelChart = echarts.init(funnelChartRef.value)
  }
  funnelChart.setOption({
    tooltip: { trigger: 'item', formatter: '{a} <br/>{b} : {c}%' },
    series: [{
      name: '转化漏斗', type: 'funnel', data,
      label: { show: true, position: 'inside' }
    }]
  })
}

const renderTypeChart = async () => {
  let data = []
  try {
    const res = await request.get('/pinche/statistics/type-distribution')
    data = res.data?.types || []
  } catch (e) { /* ignore */ }
  if (!data.length) {
    data = [
      { value: 0, name: '人找车' },
      { value: 0, name: '车找人' },
      { value: 0, name: '车找货' },
      { value: 0, name: '货找车' }
    ]
  }
  if (!typeChartRef.value) return
  if (!typeChart) {
    typeChart = echarts.init(typeChartRef.value)
  }
  typeChart.setOption({
    tooltip: { trigger: 'item', formatter: '{a} <br/>{b}: {c} ({d}%)' },
    legend: { orient: 'vertical', left: 'left' },
    series: [{
      name: '拼车类型', type: 'pie', radius: ['40%', '70%'],
      data,
      label: { show: true, formatter: '{b}: {d}%' },
      emphasis: { itemStyle: { shadowBlur: 10, shadowOffsetX: 0, shadowColor: 'rgba(0, 0, 0, 0.5)' } }
    }]
  })
}

const renderRegionChart = async () => {
  let data = []
  try {
    const res = await request.get('/pinche/statistics/region-distribution')
    data = res.data?.regions || []
  } catch (e) { /* ignore */ }
  if (!data.length) {
    data = [{ value: 0, name: '暂无数据' }]
  }
  if (!regionChartRef.value) return
  if (!regionChart) {
    regionChart = echarts.init(regionChartRef.value)
  }
  regionChart.setOption({
    tooltip: { trigger: 'item' },
    legend: { orient: 'vertical', left: 'left' },
    series: [{
      name: '地区分布', type: 'pie', radius: '60%',
      data,
      emphasis: { itemStyle: { shadowBlur: 10, shadowOffsetX: 0, shadowColor: 'rgba(0, 0, 0, 0.5)' } }
    }]
  })
}

const loadReport = async () => {
  try {
    const res = await request.get('/pinche/statistics/report')
    const d = res.data?.list || []
    if (d.length) reportData.value = d
  } catch (e) { /* ignore */ }
}

const onExport = () => {
  ElMessage.success('报表已导出（模拟）')
}

const handleResize = () => {
  trendChart?.resize()
  priceChart?.resize()
  funnelChart?.resize()
  typeChart?.resize()
  regionChart?.resize()
}

onMounted(async () => {
  loading.value = true
  try {
    await Promise.all([loadOverview(), loadHotRoutes(), loadReport()])
  } finally {
    loading.value = false
  }
  await nextTick()
  await loadTrend()
  renderPriceChart()
  renderFunnelChart()
  renderTypeChart()
  renderRegionChart()
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  trendChart?.dispose()
  priceChart?.dispose()
  funnelChart?.dispose()
  typeChart?.dispose()
  regionChart?.dispose()
})
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-card { display: flex; align-items: center; }
.stat-card :deep(.el-card__body) {
  display: flex; align-items: center; gap: 14px; padding: 16px; width: 100%;
}
.stat-icon {
  width: 44px; height: 44px; border-radius: 8px;
  display: flex; align-items: center; justify-content: center;
  color: #fff; flex-shrink: 0;
}
.stat-content { flex: 1; min-width: 0; }
.stat-value { font-size: 22px; font-weight: 600; color: #303133; }
.stat-label { font-size: 12px; color: #909399; margin-top: 2px; }

.text-success { color: #67c23a; }
.text-warning { color: #e6a23c; }
.text-danger { color: #f56c6c; }
.text-primary { color: #409eff; }

.chart-card { margin-bottom: 16px; }
.chart-container { width: 100%; height: 320px; }
.card-header-flex {
  display: flex; justify-content: space-between; align-items: center;
}

.route-text { font-size: 13px; }
</style>
