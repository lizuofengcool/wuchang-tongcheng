<template>
  <div class="app-container" v-loading="loading">
    <!-- 总览卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ overview.total_items }}</div><div class="stat-label">总商品</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ overview.today_new_items }}</div><div class="stat-label">今日新增</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ overview.active_sellers }}</div><div class="stat-label">活跃卖家</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">¥{{ formatAmount(overview.total_amount) }}</div><div class="stat-label">总成交额</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ (overview.refund_rate * 100).toFixed(1) }}%</div><div class="stat-label">退款率</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-danger">{{ (overview.completed_rate * 100).toFixed(1) }}%</div><div class="stat-label">完成率</div></div></el-card>
      </el-col>
    </el-row>

    <!-- 趋势图 + 热门商品 -->
    <el-row :gutter="16">
      <el-col :xs="24" :md="16">
        <el-card class="chart-card">
          <template #header>
            <div class="card-header-flex">
              <span>30 天价格趋势</span>
              <el-radio-group v-model="trendBrand" size="small" @change="loadPriceTrend">
                <el-radio-button label="">全部</el-radio-button>
                <el-radio-button v-for="b in trendBrands" :key="b" :label="b">{{ b }}</el-radio-button>
              </el-radio-group>
            </div>
          </template>
          <div ref="trendChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="8">
        <el-card class="chart-card">
          <template #header><span>热门商品 TOP 20</span></template>
          <el-table :data="hotItems" border size="small" max-height="400">
            <el-table-column type="index" label="#" width="50" />
            <el-table-column prop="title" label="商品" min-width="160" show-overflow-tooltip />
            <el-table-column label="价格" width="100">
              <template #default="{ row }">¥{{ Number(row.price || 0).toFixed(2) }}</template>
            </el-table-column>
            <el-table-column prop="view_count" label="浏览" width="70" />
            <el-table-column prop="fav_count" label="收藏" width="70" />
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

    <!-- 地区分布 + 推广 ROI -->
    <el-row :gutter="16">
      <el-col :xs="24" :md="12">
        <el-card class="chart-card">
          <template #header><span>地区分布</span></template>
          <div ref="regionChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="12">
        <el-card class="chart-card">
          <template #header><span>推广 ROI 报表</span></template>
          <el-table :data="promotionRoi" border size="small">
            <el-table-column prop="type" label="推广类型" width="120" />
            <el-table-column prop="amount" label="花费" width="100">
              <template #default="{ row }">¥{{ Number(row.amount || 0).toFixed(2) }}</template>
            </el-table-column>
            <el-table-column prop="impressions" label="曝光" width="80" />
            <el-table-column prop="clicks" label="点击" width="80" />
            <el-table-column prop="orders" label="下单" width="80" />
            <el-table-column label="ROI" width="80">
              <template #default="{ row }">{{ Number(row.roi || 0).toFixed(2) }}</template>
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
import {
  getErshouOverviewStats, getErshouPriceTrend, getErshouHotItems
} from '@/api/ershou'

const loading = ref(false)

const overview = reactive({
  total_items: 0, today_new_items: 0, active_sellers: 0,
  total_amount: 0, refund_rate: 0, completed_rate: 0
})

const hotItems = ref([])
const promotionRoi = ref([
  { type: '首页Banner', amount: 0, impressions: 0, clicks: 0, orders: 0, roi: 0 },
  { type: '频道置顶', amount: 0, impressions: 0, clicks: 0, orders: 0, roi: 0 },
  { type: '搜索置顶', amount: 0, impressions: 0, clicks: 0, orders: 0, roi: 0 },
  { type: '精选推荐', amount: 0, impressions: 0, clicks: 0, orders: 0, roi: 0 }
])

const trendBrand = ref('')
const trendBrands = ref(['苹果', '华为', '小米'])

// ===== 图表 refs =====
const trendChartRef = ref(null)
const priceChartRef = ref(null)
const funnelChartRef = ref(null)
const regionChartRef = ref(null)

let trendChart = null
let priceChart = null
let funnelChart = null
let regionChart = null

const formatAmount = (n) => Number(n || 0).toFixed(2)

// ===== 加载数据 =====
const loadOverview = async () => {
  try {
    const res = await getErshouOverviewStats()
    Object.assign(overview, res.data || {})
  } catch (e) { /* ignore */ }
}

const loadPriceTrend = async () => {
  try {
    const res = await getErshouPriceTrend({ brand: trendBrand.value || undefined })
    const d = res.data || {}
    renderTrendChart(d.dates || [], d.prices || [])
  } catch (e) {
    // 渲染空数据
    renderTrendChart([], [])
  }
}

const loadHotItems = async () => {
  try {
    const res = await getErshouHotItems({ page: 1, page_size: 20 })
    const d = res.data
    hotItems.value = d?.list || d || []
  } catch (e) {
    hotItems.value = []
  }
}

// ===== 渲染图表 =====
const renderTrendChart = (dates, prices) => {
  if (!trendChartRef.value) return
  if (!trendChart) {
    trendChart = echarts.init(trendChartRef.value)
  }
  trendChart.setOption({
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: dates },
    yAxis: { type: 'value', axisLabel: { formatter: '¥{value}' } },
    series: [{
      name: '均价', type: 'line', data: prices, smooth: true,
      areaStyle: { opacity: 0.3 }, itemStyle: { color: '#409eff' }
    }],
    grid: { left: 50, right: 20, top: 30, bottom: 30 }
  })
}

const renderPriceChart = () => {
  if (!priceChartRef.value) return
  if (!priceChart) {
    priceChart = echarts.init(priceChartRef.value)
  }
  priceChart.setOption({
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: ['0-100', '100-500', '500-1000', '1000-5000', '5000-1万', '1万+'] },
    yAxis: { type: 'value' },
    series: [{
      name: '商品数', type: 'bar', data: [120, 200, 150, 80, 30, 20],
      itemStyle: { color: '#67c23a' }
    }],
    grid: { left: 50, right: 20, top: 30, bottom: 30 }
  })
}

const renderFunnelChart = () => {
  if (!funnelChartRef.value) return
  if (!funnelChart) {
    funnelChart = echarts.init(funnelChartRef.value)
  }
  funnelChart.setOption({
    tooltip: { trigger: 'item', formatter: '{a} <br/>{b} : {c}%' },
    series: [{
      name: '转化漏斗', type: 'funnel',
      data: [
        { value: 100, name: '曝光' },
        { value: 35, name: '点击' },
        { value: 12, name: '咨询' },
        { value: 5, name: '下单' },
        { value: 3, name: '成交' }
      ],
      label: { show: true, position: 'inside' }
    }]
  })
}

const renderRegionChart = () => {
  if (!regionChartRef.value) return
  if (!regionChart) {
    regionChart = echarts.init(regionChartRef.value)
  }
  regionChart.setOption({
    tooltip: { trigger: 'item' },
    legend: { orient: 'vertical', left: 'left' },
    series: [{
      name: '地区分布', type: 'pie', radius: '60%',
      data: [
        { value: 1048, name: '北京' },
        { value: 735, name: '上海' },
        { value: 580, name: '广州' },
        { value: 484, name: '深圳' },
        { value: 300, name: '其他' }
      ],
      emphasis: { itemStyle: { shadowBlur: 10, shadowOffsetX: 0, shadowColor: 'rgba(0, 0, 0, 0.5)' } }
    }]
  })
}

const handleResize = () => {
  trendChart?.resize()
  priceChart?.resize()
  funnelChart?.resize()
  regionChart?.resize()
}

onMounted(async () => {
  loading.value = true
  try {
    await Promise.all([loadOverview(), loadHotItems()])
  } finally {
    loading.value = false
  }
  await nextTick()
  await loadPriceTrend()
  renderPriceChart()
  renderFunnelChart()
  renderRegionChart()
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  trendChart?.dispose()
  priceChart?.dispose()
  funnelChart?.dispose()
  regionChart?.dispose()
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
.card-header-flex {
  display: flex; justify-content: space-between; align-items: center;
}
</style>
