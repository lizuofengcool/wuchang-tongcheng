<template>
  <div class="app-container" v-loading="loading">
    <!-- 总览卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #409eff">
            <el-icon :size="22"><Briefcase /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ overview.total_jobs }}</div>
            <div class="stat-label">总岗位数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #67c23a">
            <el-icon :size="22"><CirclePlus /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ overview.today_new }}</div>
            <div class="stat-label">今日新增</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #722ed1">
            <el-icon :size="22"><User /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ overview.total_employers }}</div>
            <div class="stat-label">雇主总数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #13c2c2">
            <el-icon :size="22"><Avatar /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ overview.total_workers }}</div>
            <div class="stat-label">工人总数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #e6a23c">
            <el-icon :size="22"><Money /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value text-primary">¥{{ formatAmount(overview.total_amount) }}</div>
            <div class="stat-label">成交总额</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #f56c6c">
            <el-icon :size="22"><TrendCharts /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value text-success">{{ overview.completion_rate }}%</div>
            <div class="stat-label">完成率</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 岗位类型分布 + 薪资区间 -->
    <el-row :gutter="16">
      <el-col :xs="24" :md="12">
        <el-card class="chart-card">
          <template #header><span>岗位类型分布</span></template>
          <div ref="typeChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="12">
        <el-card class="chart-card">
          <template #header><span>薪资区间分布</span></template>
          <div ref="salaryChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 热门技能 TOP10 + 评价分布 -->
    <el-row :gutter="16">
      <el-col :xs="24" :md="12">
        <el-card class="chart-card">
          <template #header>
            <div class="card-header-flex">
              <span>热门技能 TOP 10</span>
              <el-button text @click="loadHotSkills">刷新</el-button>
            </div>
          </template>
          <el-table :data="hotSkills" border size="small" max-height="320">
            <el-table-column type="index" label="#" width="50" />
            <el-table-column prop="name" label="技能名" min-width="120" show-overflow-tooltip />
            <el-table-column prop="category" label="分类" width="100" />
            <el-table-column prop="usage_count" label="使用次数" width="100" sortable />
            <el-table-column label="热度" width="160">
              <template #default="{ row }">
                <el-progress :percentage="getHotPercent(row)" :color="getHotColor(row)" />
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="12">
        <el-card class="chart-card">
          <template #header><span>评价分布</span></template>
          <div ref="ratingChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 地区分布 + 雇主排行 -->
    <el-row :gutter="16">
      <el-col :xs="24" :md="12">
        <el-card class="chart-card">
          <template #header><span>地区分布</span></template>
          <div ref="regionChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="12">
        <el-card class="chart-card">
          <template #header><span>雇主发布岗位排行 TOP 10</span></template>
          <el-table :data="topEmployers" border size="small" max-height="320">
            <el-table-column type="index" label="#" width="50" />
            <el-table-column prop="company_name" label="公司名" min-width="140" show-overflow-tooltip />
            <el-table-column prop="contact_name" label="联系人" width="90" />
            <el-table-column prop="job_count" label="岗位数" width="80" sortable />
            <el-table-column prop="applied_count" label="报名数" width="80" sortable />
            <el-table-column label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="row.verified ? 'success' : 'info'" size="small">
                  {{ row.verified ? '已认证' : '未认证' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, nextTick, onBeforeUnmount } from 'vue'
import * as echarts from 'echarts'
import {
  Briefcase, CirclePlus, User, Avatar, Money, TrendCharts
} from '@element-plus/icons-vue'
import {
  getLinggongOverviewStats, getLinggongHotSkills, adminListLinggongEmployers
} from '@/api/linggong'

const loading = ref(false)

const overview = reactive({
  total_jobs: 0,
  today_new: 0,
  total_employers: 0,
  total_workers: 0,
  total_amount: 0,
  completion_rate: 0
})

const hotSkills = ref([])
const topEmployers = ref([])

// 图表 refs
const typeChartRef = ref(null)
const salaryChartRef = ref(null)
const ratingChartRef = ref(null)
const regionChartRef = ref(null)

let typeChart = null
let salaryChart = null
let ratingChart = null
let regionChart = null

const formatAmount = (n) => Number(n || 0).toFixed(2)

const loadOverview = async () => {
  try {
    const res = await getLinggongOverviewStats()
    Object.assign(overview, res.data || {})
  } catch (e) { /* ignore */ }
}

const loadHotSkills = async () => {
  try {
    const res = await getLinggongHotSkills({ page: 1, page_size: 10 })
    const d = res.data
    hotSkills.value = d?.list || d || []
  } catch (e) {
    hotSkills.value = []
  }
}

const loadTopEmployers = async () => {
  try {
    const res = await adminListLinggongEmployers({ page: 1, page_size: 10, sort: 'job_count' })
    const d = res.data
    topEmployers.value = d?.list || d || []
  } catch (e) {
    topEmployers.value = []
  }
}

const getHotPercent = (row) => {
  if (!hotSkills.value.length) return 0
  const max = Math.max(...hotSkills.value.map((s) => Number(s.usage_count || 0)), 1)
  return Math.round((Number(row.usage_count || 0) / max) * 100)
}

const getHotColor = (row) => {
  const p = getHotPercent(row)
  if (p >= 70) return '#f56c6c'
  if (p >= 40) return '#e6a23c'
  return '#67c23a'
}

// 渲染岗位类型分布
const renderTypeChart = () => {
  if (!typeChartRef.value) return
  if (!typeChart) typeChart = echarts.init(typeChartRef.value)
  typeChart.setOption({
    tooltip: { trigger: 'item', formatter: '{a} <br/>{b}: {c} ({d}%)' },
    legend: { orient: 'vertical', left: 'left' },
    series: [{
      name: '岗位类型',
      type: 'pie',
      radius: ['40%', '70%'],
      avoidLabelOverlap: false,
      itemStyle: { borderRadius: 8, borderColor: '#fff', borderWidth: 2 },
      label: { show: false, position: 'center' },
      emphasis: { label: { show: true, fontSize: 16, fontWeight: 'bold' } },
      labelLine: { show: false },
      data: [
        { value: 1048, name: '短期兼职' },
        { value: 735, name: '长期兼职' },
        { value: 580, name: '任务制' },
        { value: 484, name: '小时工' },
        { value: 300, name: '日结工' },
        { value: 200, name: '临时工' }
      ]
    }]
  })
}

// 渲染薪资区间分布
const renderSalaryChart = () => {
  if (!salaryChartRef.value) return
  if (!salaryChart) salaryChart = echarts.init(salaryChartRef.value)
  salaryChart.setOption({
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: ['0-100', '100-200', '200-500', '500-1k', '1k-3k', '3k+'] },
    yAxis: { type: 'value' },
    series: [{
      name: '岗位数',
      type: 'bar',
      data: [120, 200, 150, 80, 60, 30],
      itemStyle: { color: '#67c23a' }
    }],
    grid: { left: 50, right: 20, top: 30, bottom: 30 }
  })
}

// 渲染评价分布
const renderRatingChart = () => {
  if (!ratingChartRef.value) return
  if (!ratingChart) ratingChart = echarts.init(ratingChartRef.value)
  ratingChart.setOption({
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: ['1星', '2星', '3星', '4星', '5星'] },
    yAxis: { type: 'value' },
    series: [{
      name: '评价数',
      type: 'bar',
      data: [
        { value: 20, itemStyle: { color: '#f56c6c' } },
        { value: 30, itemStyle: { color: '#e6a23c' } },
        { value: 80, itemStyle: { color: '#e6a23c' } },
        { value: 250, itemStyle: { color: '#67c23a' } },
        { value: 620, itemStyle: { color: '#67c23a' } }
      ]
    }],
    grid: { left: 50, right: 20, top: 30, bottom: 30 }
  })
}

// 渲染地区分布
const renderRegionChart = () => {
  if (!regionChartRef.value) return
  if (!regionChart) regionChart = echarts.init(regionChartRef.value)
  regionChart.setOption({
    tooltip: { trigger: 'item' },
    legend: { orient: 'vertical', left: 'left' },
    series: [{
      name: '地区分布',
      type: 'pie',
      radius: '60%',
      data: [
        { value: 1048, name: '北京' },
        { value: 735, name: '上海' },
        { value: 580, name: '广州' },
        { value: 484, name: '深圳' },
        { value: 300, name: '杭州' },
        { value: 200, name: '其他' }
      ],
      emphasis: {
        itemStyle: {
          shadowBlur: 10,
          shadowOffsetX: 0,
          shadowColor: 'rgba(0, 0, 0, 0.5)'
        }
      }
    }]
  })
}

const handleResize = () => {
  typeChart?.resize()
  salaryChart?.resize()
  ratingChart?.resize()
  regionChart?.resize()
}

onMounted(async () => {
  loading.value = true
  try {
    await Promise.all([loadOverview(), loadHotSkills(), loadTopEmployers()])
  } finally {
    loading.value = false
  }
  await nextTick()
  renderTypeChart()
  renderSalaryChart()
  renderRatingChart()
  renderRegionChart()
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  typeChart?.dispose()
  salaryChart?.dispose()
  ratingChart?.dispose()
  regionChart?.dispose()
})
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-card { display: flex; align-items: center; }
.stat-card :deep(.el-card__body) { display: flex; align-items: center; width: 100%; padding: 16px; }
.stat-icon { width: 48px; height: 48px; border-radius: 8px; color: #fff; display: flex; align-items: center; justify-content: center; margin-right: 12px; flex-shrink: 0; }
.stat-content { flex: 1; }
.stat-value { font-size: 22px; font-weight: 600; color: #303133; }
.stat-label { font-size: 12px; color: #909399; margin-top: 2px; }
.text-primary { color: #409eff; }
.text-success { color: #67c23a; font-weight: 600; }
.chart-card { margin-bottom: 16px; }
.chart-container { width: 100%; height: 320px; }
.card-header-flex { display: flex; justify-content: space-between; align-items: center; }
</style>
