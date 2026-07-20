<template>
  <div class="app-container" v-loading="loading">
    <!-- 总览卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ overview.total_business }}</div><div class="stat-label">总商户</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ overview.today_new }}</div><div class="stat-label">今日新增</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ overview.total_views }}</div><div class="stat-label">总浏览量</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ overview.total_calls }}</div><div class="stat-label">总电话量</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-danger">{{ overview.total_favs }}</div><div class="stat-label">总收藏</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ overview.avg_rating }}</div><div class="stat-label">平均评分</div></div></el-card>
      </el-col>
    </el-row>

    <!-- 日期范围选择 + 趋势图 -->
    <el-card class="chart-card">
      <template #header>
        <div class="card-header-flex">
          <span>商户数据趋势</span>
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            style="width: 260px"
            @change="loadTrend"
          />
        </div>
      </template>
      <div ref="trendChartRef" class="chart-container"></div>
    </el-card>

    <!-- 热门商户 + 分类分布 -->
    <el-row :gutter="16">
      <el-col :xs="24" :md="12">
        <el-card class="chart-card">
          <template #header><span>热门商户 TOP 20（按浏览量）</span></template>
          <el-table :data="hotBusiness" border size="small" max-height="400">
            <el-table-column type="index" label="#" width="50" />
            <el-table-column prop="name" label="商户" min-width="160" show-overflow-tooltip />
            <el-table-column prop="view_count" label="浏览" width="80" />
            <el-table-column prop="call_count" label="电话" width="80" />
            <el-table-column prop="fav_count" label="收藏" width="80" />
            <el-table-column label="评分" width="80">
              <template #default="{ row }">{{ Number(row.rating || 0).toFixed(1) }}</template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="12">
        <el-card class="chart-card">
          <template #header><span>热门分类 TOP 10</span></template>
          <div ref="categoryChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 分类统计表 -->
    <el-card class="chart-card">
      <template #header><span>分类商户数分布</span></template>
      <el-table :data="categoryStats" border size="small">
        <el-table-column type="index" label="#" width="50" />
        <el-table-column prop="category_name" label="分类" min-width="160" />
        <el-table-column prop="business_count" label="商户数" width="120" sortable />
        <el-table-column prop="total_views" label="总浏览" width="120" sortable />
        <el-table-column prop="total_calls" label="总电话" width="120" sortable />
        <el-table-column prop="total_favs" label="总收藏" width="120" sortable />
        <el-table-column label="平均评分" width="120">
          <template #default="{ row }">{{ Number(row.avg_rating || 0).toFixed(2) }}</template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, nextTick, onBeforeUnmount } from 'vue'
import * as echarts from 'echarts'
import request from '@/utils/request'

const loading = ref(false)

const overview = reactive({
  total_business: 0, today_new: 0, total_views: 0,
  total_calls: 0, total_favs: 0, avg_rating: '0.0'
})

const hotBusiness = ref([])
const categoryStats = ref([])
const dateRange = ref([])

// ===== 图表 refs =====
const trendChartRef = ref(null)
const categoryChartRef = ref(null)
let trendChart = null
let categoryChart = null

// ===== 加载数据 =====
const loadOverview = async () => {
  try {
    const res = await request.get('/dh114/admin/statistics/overview')
    Object.assign(overview, res.data || {})
  } catch (e) { /* ignore */ }
}

const loadHotBusiness = async () => {
  try {
    const res = await request.get('/dh114/statistics/hot-business', { params: { limit: 20 } })
    const d = res.data
    hotBusiness.value = d?.list || d || []
  } catch (e) {
    hotBusiness.value = []
  }
}

const loadCategoryStats = async () => {
  try {
    const res = await request.get('/dh114/statistics/by-category')
    const d = res.data
    categoryStats.value = d?.list || d || []
    renderCategoryChart()
  } catch (e) {
    categoryStats.value = []
  }
}

const loadTrend = async () => {
  try {
    const params = {}
    if (dateRange.value && dateRange.value.length === 2) {
      params.start_date = dateRange.value[0]
      params.end_date = dateRange.value[1]
    } else {
      const end = new Date()
      const start = new Date()
      start.setDate(start.getDate() - 30)
      params.start_date = start.toISOString().slice(0, 10)
      params.end_date = end.toISOString().slice(0, 10)
    }
    const res = await request.get('/dh114/statistics/date-range', { params })
    const d = res.data || {}
    renderTrendChart(d.dates || [], d.business_counts || [], d.view_counts || [])
  } catch (e) {
    renderTrendChart([], [], [])
  }
}

// ===== 渲染图表 =====
const renderTrendChart = (dates, businessCounts, viewCounts) => {
  if (!trendChartRef.value) return
  if (!trendChart) {
    trendChart = echarts.init(trendChartRef.value)
  }
  trendChart.setOption({
    tooltip: { trigger: 'axis' },
    legend: { data: ['新增商户', '浏览量'] },
    xAxis: { type: 'category', data: dates },
    yAxis: [
      { type: 'value', name: '商户数' },
      { type: 'value', name: '浏览量' }
    ],
    series: [
      {
        name: '新增商户', type: 'bar', data: businessCounts,
        itemStyle: { color: '#409eff' }
      },
      {
        name: '浏览量', type: 'line', yAxisIndex: 1, data: viewCounts, smooth: true,
        itemStyle: { color: '#67c23a' }, areaStyle: { opacity: 0.3 }
      }
    ],
    grid: { left: 50, right: 50, top: 40, bottom: 30 }
  })
}

const renderCategoryChart = () => {
  if (!categoryChartRef.value) return
  if (!categoryChart) {
    categoryChart = echarts.init(categoryChartRef.value)
  }
  const data = categoryStats.value.slice(0, 10).map((c) => ({
    name: c.category_name || `分类#${c.category_id}`,
    value: c.business_count || 0
  }))
  categoryChart.setOption({
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
    legend: { orient: 'vertical', left: 'left' },
    series: [{
      name: '商户数', type: 'pie', radius: ['40%', '70%'],
      data,
      emphasis: { itemStyle: { shadowBlur: 10, shadowOffsetX: 0, shadowColor: 'rgba(0, 0, 0, 0.5)' } }
    }]
  })
}

const handleResize = () => {
  trendChart?.resize()
  categoryChart?.resize()
}

onMounted(async () => {
  loading.value = true
  try {
    await Promise.all([loadOverview(), loadHotBusiness(), loadCategoryStats()])
  } finally {
    loading.value = false
  }
  await nextTick()
  await loadTrend()
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  trendChart?.dispose()
  categoryChart?.dispose()
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
