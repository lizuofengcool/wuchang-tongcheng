<template>
  <div class="app-container" v-loading="loading">
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="4"><el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ overview.total_jobs || 0 }}</div><div class="stat-label">总职位</div></div></el-card></el-col>
      <el-col :xs="12" :sm="8" :md="4"><el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ overview.today_new || 0 }}</div><div class="stat-label">今日新增</div></div></el-card></el-col>
      <el-col :xs="12" :sm="8" :md="4"><el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ overview.pending_audit || 0 }}</div><div class="stat-label">待审核</div></div></el-card></el-col>
      <el-col :xs="12" :sm="8" :md="4"><el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ overview.published || 0 }}</div><div class="stat-label">已发布</div></div></el-card></el-col>
      <el-col :xs="12" :sm="8" :md="4"><el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ overview.total_applications || 0 }}</div><div class="stat-label">总投递</div></div></el-card></el-col>
      <el-col :xs="12" :sm="8" :md="4"><el-card shadow="hover"><div class="stat-item"><div class="stat-value text-danger">{{ overview.violation || 0 }}</div><div class="stat-label">违规</div></div></el-card></el-col>
    </el-row>

    <el-row :gutter="16">
      <el-col :xs="24" :md="16">
        <el-card class="chart-card">
          <template #header><span>薪资趋势</span></template>
          <div ref="salaryChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="8">
        <el-card class="chart-card">
          <template #header><span>热门职位 TOP 15</span></template>
          <el-table :data="hotJobs" border size="small" max-height="380">
            <el-table-column type="index" label="#" width="50" />
            <el-table-column prop="title" label="职位" min-width="140" show-overflow-tooltip />
            <el-table-column prop="application_count" label="投递" width="70" />
            <el-table-column prop="view_count" label="浏览" width="70" />
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16">
      <el-col :xs="24" :md="12">
        <el-card class="chart-card">
          <template #header><span>分类分布</span></template>
          <div ref="categoryChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="12">
        <el-card class="chart-card">
          <template #header><span>地区分布</span></template>
          <div ref="regionChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16">
      <el-col :span="24">
        <el-card class="chart-card">
          <template #header><span>转化漏斗</span></template>
          <div ref="funnelChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, nextTick, onBeforeUnmount } from 'vue'
import * as echarts from 'echarts'
import { getOverviewStats, getConversionFunnel, getHotJobs, getSalaryTrend, getCategoryStats, getRegionStats } from '@/api/job'

const loading = ref(false)
const overview = reactive({})
const hotJobs = ref([])

const salaryChartRef = ref(null)
const categoryChartRef = ref(null)
const regionChartRef = ref(null)
const funnelChartRef = ref(null)
let salaryChart, categoryChart, regionChart, funnelChart

const loadOverview = async () => {
  try { const res = await getOverviewStats(); Object.assign(overview, res.data || {}) } catch (e) { /* */ }
}

const loadHotJobs = async () => {
  try { const res = await getHotJobs({ page: 1, page_size: 15 }); const d = res.data || {}; hotJobs.value = d.list || d || [] } catch (e) { hotJobs.value = [] }
}

const renderSalaryChart = async () => {
  try {
    const res = await getSalaryTrend({ days: 30 })
    const d = res.data || {}
    const dates = d.dates || d.x_axis || []
    const values = d.values || d.avg_salaries || d.data || []
    await nextTick()
    if (!salaryChartRef.value) return
    salaryChart = echarts.init(salaryChartRef.value)
    salaryChart.setOption({
      tooltip: { trigger: 'axis' },
      xAxis: { type: 'category', data: dates },
      yAxis: { type: 'value', name: '薪资(K)' },
      series: [{ name: '平均薪资', type: 'line', smooth: true, data: values, areaStyle: {}, itemStyle: { color: '#409eff' } }]
    })
  } catch (e) { /* */ }
}

const renderCategoryChart = async () => {
  try {
    const res = await getCategoryStats()
    const d = res.data || {}
    const list = d.list || d || []
    await nextTick()
    if (!categoryChartRef.value) return
    categoryChart = echarts.init(categoryChartRef.value)
    categoryChart.setOption({
      tooltip: { trigger: 'item' },
      legend: { bottom: 0 },
      series: [{ type: 'pie', radius: ['40%', '70%'], data: list.map((i) => ({ name: i.name || i.category || i.label, value: i.count || i.value || 0 })) }]
    })
  } catch (e) { /* */ }
}

const renderRegionChart = async () => {
  try {
    const res = await getRegionStats()
    const d = res.data || {}
    const list = d.list || d || []
    await nextTick()
    if (!regionChartRef.value) return
    regionChart = echarts.init(regionChartRef.value)
    regionChart.setOption({
      tooltip: { trigger: 'axis' },
      xAxis: { type: 'category', data: list.map((i) => i.name || i.region || i.label || '-'), axisLabel: { rotate: 30 } },
      yAxis: { type: 'value' },
      series: [{ type: 'bar', data: list.map((i) => i.count || i.value || 0), itemStyle: { color: '#67c23a' } }]
    })
  } catch (e) { /* */ }
}

const renderFunnelChart = async () => {
  try {
    const res = await getConversionFunnel()
    const d = res.data || {}
    const stages = d.stages || d.list || [{ name: '浏览', value: d.view_count || 0 }, { name: '投递', value: d.application_count || 0 }, { name: '面试', value: d.interview_count || 0 }, { name: '录用', value: d.hired_count || 0 }]
    await nextTick()
    if (!funnelChartRef.value) return
    funnelChart = echarts.init(funnelChartRef.value)
    funnelChart.setOption({
      tooltip: { trigger: 'item', formatter: '{b}: {c}' },
      series: [{ type: 'funnel', left: '10%', width: '80%', data: stages.map((s) => ({ name: s.name || s.label, value: s.value || s.count || 0 })), label: { show: true, position: 'inside' } }]
    })
  } catch (e) { /* */ }
}

const handleResize = () => {
  salaryChart?.resize(); categoryChart?.resize(); regionChart?.resize(); funnelChart?.resize()
}

onMounted(async () => {
  loading.value = true
  try {
    await Promise.all([loadOverview(), loadHotJobs()])
    await Promise.all([renderSalaryChart(), renderCategoryChart(), renderRegionChart(), renderFunnelChart()])
  } finally { loading.value = false }
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  salaryChart?.dispose(); categoryChart?.dispose(); regionChart?.dispose(); funnelChart?.dispose()
})
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-item { padding: 12px 16px; }
.stat-value { font-size: 24px; font-weight: 600; color: #303133; }
.stat-label { font-size: 13px; color: #909399; margin-top: 4px; }
.text-success { color: #67c23a; }
.text-warning { color: #e6a23c; }
.text-primary { color: #409eff; }
.text-danger { color: #f56c6c; }
.chart-card { margin-bottom: 16px; }
.chart-container { height: 320px; }
</style>
