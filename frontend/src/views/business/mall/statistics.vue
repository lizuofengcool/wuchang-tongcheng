<template>
  <div class="app-container" v-loading="loading">
    <!-- 平台总览 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ overview.shop_count }}</div><div class="stat-label">店铺总数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ overview.product_count }}</div><div class="stat-label">商品总数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ overview.order_count }}</div><div class="stat-label">订单总数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">¥{{ Number(overview.total_amount || 0).toFixed(2) }}</div><div class="stat-label">成交总额</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ overview.user_count }}</div><div class="stat-label">活跃用户数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-danger">{{ overview.refund_count }}</div><div class="stat-label">退款单数</div></div></el-card>
      </el-col>
    </el-row>

    <!-- 日期范围选择 -->
    <el-card class="chart-card">
      <template #header>
        <div class="card-header-flex">
          <span>数据总览（按日期范围）</span>
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            style="width: 260px"
            @change="loadOverview"
          />
        </div>
      </template>
      <el-row :gutter="16">
        <el-col :xs="24" :md="12">
          <div ref="categoryChartRef" class="chart-container"></div>
        </el-col>
        <el-col :xs="24" :md="12">
          <div ref="trendChartRef" class="chart-container"></div>
        </el-col>
      </el-row>
    </el-card>

    <!-- 热销商品 + 热门店铺 -->
    <el-row :gutter="16">
      <el-col :xs="24" :md="12">
        <el-card class="chart-card">
          <template #header><span>热销商品 TOP 20</span></template>
          <el-table :data="hotProducts" border size="small" max-height="420">
            <el-table-column type="index" label="#" width="50" />
            <el-table-column prop="product_name" label="商品" min-width="160" show-overflow-tooltip />
            <el-table-column prop="shop_name" label="店铺" width="120" show-overflow-tooltip />
            <el-table-column prop="sales_count" label="销量" width="80" sortable />
            <el-table-column label="销售额" width="110">
              <template #default="{ row }">¥{{ Number(row.sales_amount || 0).toFixed(2) }}</template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="12">
        <el-card class="chart-card">
          <template #header><span>热门店铺 TOP 20</span></template>
          <el-table :data="hotShops" border size="small" max-height="420">
            <el-table-column type="index" label="#" width="50" />
            <el-table-column prop="shop_name" label="店铺" min-width="160" show-overflow-tooltip />
            <el-table-column prop="order_count" label="订单数" width="90" sortable />
            <el-table-column prop="product_count" label="商品数" width="90" sortable />
            <el-table-column label="成交额" width="110">
              <template #default="{ row }">¥{{ Number(row.total_amount || 0).toFixed(2) }}</template>
            </el-table-column>
            <el-table-column label="评分" width="80">
              <template #default="{ row }">{{ Number(row.avg_rating || 0).toFixed(1) }}</template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <!-- 历史统计列表 -->
    <el-card class="chart-card">
      <template #header>
        <div class="card-header-flex">
          <span>历史统计列表</span>
          <el-button :icon="Refresh" @click="loadStatList">刷新</el-button>
        </div>
      </template>
      <el-table v-loading="listLoading" :data="statList" border size="small" @sort-change="onSortChange">
        <el-table-column prop="id" label="ID" width="70" sortable="custom" />
        <el-table-column prop="stat_date" label="统计日期" width="120" sortable="custom" />
        <el-table-column prop="shop_count" label="店铺数" width="90" sortable="custom" />
        <el-table-column prop="product_count" label="商品数" width="90" sortable="custom" />
        <el-table-column prop="order_count" label="订单数" width="90" sortable="custom" />
        <el-table-column label="成交额" width="120" sortable="custom" prop="total_amount">
          <template #default="{ row }">¥{{ Number(row.total_amount || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="user_count" label="用户数" width="90" sortable="custom" />
        <el-table-column prop="refund_count" label="退款数" width="90" sortable="custom" />
        <el-table-column label="退款金额" width="120">
          <template #default="{ row }">¥{{ Number(row.refund_amount || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="创建时间" width="160" prop="created_at" sortable="custom">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
      </el-table>
      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadStatList" @size-change="loadStatList" />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, nextTick, onBeforeUnmount } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import * as echarts from 'echarts'
import { formatTime } from '@/utils/format'
import {
  getMallStatisticOverview, getMallStatisticList,
  getMallHotProducts, getMallHotShops, getMallHotCategories
} from '@/api/mall'

const loading = ref(false)
const listLoading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const statList = ref([])
const sortField = ref('stat_date')
const sortOrder = ref('descending')

const overview = reactive({
  shop_count: 0, product_count: 0, order_count: 0,
  total_amount: 0, user_count: 0, refund_count: 0
})

const hotProducts = ref([])
const hotShops = ref([])
const hotCategories = ref([])
const dateRange = ref([])

// ===== 图表 refs =====
const categoryChartRef = ref(null)
const trendChartRef = ref(null)
let categoryChart = null
let trendChart = null

// ===== 加载数据 =====
const loadOverview = async () => {
  try {
    const params = {}
    if (dateRange.value && dateRange.value.length === 2) {
      params.start_date = dateRange.value[0]
      params.end_date = dateRange.value[1]
    }
    const res = await getMallStatisticOverview(params)
    Object.assign(overview, res.data || {})
    // 渲染趋势图（overview 中可能带 daily_trend 数据）
    const data = res.data || {}
    renderTrendChart(data.daily_trend || data.trend || [], data.daily_dates || data.dates || [])
  } catch (e) { /* ignore */ }
}

const loadHotProducts = async () => {
  try {
    const res = await getMallHotProducts({ limit: 20 })
    const d = res.data
    hotProducts.value = d?.list || d || []
  } catch (e) {
    hotProducts.value = []
  }
}

const loadHotShops = async () => {
  try {
    const res = await getMallHotShops({ limit: 20 })
    const d = res.data
    hotShops.value = d?.list || d || []
  } catch (e) {
    hotShops.value = []
  }
}

const loadHotCategories = async () => {
  try {
    const params = {}
    if (dateRange.value && dateRange.value.length === 2) {
      params.start_date = dateRange.value[0]
      params.end_date = dateRange.value[1]
    }
    const res = await getMallHotCategories(params)
    const d = res.data
    hotCategories.value = d?.list || d || []
    renderCategoryChart()
  } catch (e) {
    hotCategories.value = []
  }
}

const loadStatList = async () => {
  listLoading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      sort: sortField.value,
      order: sortOrder.value
    }
    const res = await getMallStatisticList(params)
    const data = res.data || {}
    statList.value = data.list || []
    total.value = data.total || 0
  } catch (e) {
    ElMessage.error('加载统计列表失败')
    statList.value = []
    total.value = 0
  } finally {
    listLoading.value = false
  }
}

const onSortChange = ({ prop, order }) => {
  sortField.value = prop || 'stat_date'
  sortOrder.value = order || 'descending'
  loadStatList()
}

// ===== 渲染图表 =====
const renderCategoryChart = () => {
  if (!categoryChartRef.value) return
  if (!categoryChart) {
    categoryChart = echarts.init(categoryChartRef.value)
  }
  const data = hotCategories.value.slice(0, 10).map((c) => ({
    name: c.category_name || c.name || `分类#${c.category_id}`,
    value: c.count || c.sales_count || c.order_count || 0
  }))
  categoryChart.setOption({
    title: { text: '热门分类 TOP 10', left: 'center', textStyle: { fontSize: 14 } },
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
    legend: { orient: 'vertical', left: 'left', type: 'scroll' },
    series: [{
      name: '分类占比', type: 'pie', radius: ['40%', '70%'],
      center: ['60%', '55%'],
      data,
      emphasis: { itemStyle: { shadowBlur: 10, shadowOffsetX: 0, shadowColor: 'rgba(0, 0, 0, 0.5)' } }
    }]
  })
}

const renderTrendChart = (trendArr, datesArr) => {
  if (!trendChartRef.value) return
  if (!trendChart) {
    trendChart = echarts.init(trendChartRef.value)
  }
  // 兼容两种数据格式：数组对象 / 平行数组
  let dates = datesArr
  let orderCounts = []
  let amountArr = []
  if (Array.isArray(trendArr) && trendArr.length && typeof trendArr[0] === 'object') {
    dates = trendArr.map((t) => t.date || t.stat_date)
    orderCounts = trendArr.map((t) => t.order_count || 0)
    amountArr = trendArr.map((t) => Number(t.total_amount || 0))
  }
  trendChart.setOption({
    title: { text: '订单/成交额趋势', left: 'center', textStyle: { fontSize: 14 } },
    tooltip: { trigger: 'axis' },
    legend: { data: ['订单数', '成交额'], top: 25 },
    xAxis: { type: 'category', data: dates },
    yAxis: [
      { type: 'value', name: '订单数' },
      { type: 'value', name: '成交额', axisLabel: { formatter: (v) => '¥' + v } }
    ],
    series: [
      { name: '订单数', type: 'bar', data: orderCounts, itemStyle: { color: '#409eff' } },
      { name: '成交额', type: 'line', yAxisIndex: 1, data: amountArr, smooth: true, itemStyle: { color: '#67c23a' }, areaStyle: { opacity: 0.3 } }
    ],
    grid: { left: 50, right: 60, top: 60, bottom: 30 }
  })
}

const handleResize = () => {
  categoryChart?.resize()
  trendChart?.resize()
}

onMounted(async () => {
  loading.value = true
  try {
    await Promise.all([loadOverview(), loadHotProducts(), loadHotShops(), loadHotCategories(), loadStatList()])
  } finally {
    loading.value = false
  }
  await nextTick()
  // 若 overview 未带 trend，重新渲染空图保证占位
  if (!trendChart && trendChartRef.value) {
    renderTrendChart([], [])
  }
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  categoryChart?.dispose()
  trendChart?.dispose()
})
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-item { text-align: center; padding: 16px; }
.stat-value { font-size: 20px; font-weight: 600; color: #303133; }
.text-success { color: #67c23a; }
.text-warning { color: #e6a23c; }
.text-danger { color: #f56c6c; }
.text-primary { color: #409eff; }
.chart-card { margin-bottom: 16px; }
.chart-container { width: 100%; height: 320px; }
.card-header-flex { display: flex; justify-content: space-between; align-items: center; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
