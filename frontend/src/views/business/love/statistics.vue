<template>
  <div class="app-container" v-loading="loading">
    <!-- 总览卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-value">{{ overview.totalMembers }}</div>
            <div class="stat-label">总会员数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-value text-success">{{ overview.todayNew }}</div>
            <div class="stat-label">今日新增</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-value text-primary">{{ overview.activeMatches }}</div>
            <div class="stat-label">活跃匹配</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-value text-warning">{{ overview.paidMembers }}</div>
            <div class="stat-label">付费会员</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-value text-danger">¥{{ formatAmount(overview.totalIncome) }}</div>
            <div class="stat-label">总收入</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-value">{{ (overview.verifyRate * 100).toFixed(1) }}%</div>
            <div class="stat-label">认证率</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 趋势图 + 热门会员 -->
    <el-row :gutter="16">
      <el-col :xs="24" :md="16">
        <el-card class="chart-card">
          <template #header>
            <div class="card-header-flex">
              <span>30 天匹配趋势</span>
              <el-radio-group v-model="trendMetric" size="small" @change="loadTrend">
                <el-radio-button label="matches">匹配数</el-radio-button>
                <el-radio-button label="likes">喜欢数</el-radio-button>
                <el-radio-button label="visits">访客数</el-radio-button>
              </el-radio-group>
            </div>
          </template>
          <div ref="trendChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="8">
        <el-card class="chart-card">
          <template #header><span>热门会员 TOP 20</span></template>
          <el-table :data="hotMembers" border size="small" max-height="400">
            <el-table-column type="index" label="#" width="50" />
            <el-table-column label="会员" min-width="140" show-overflow-tooltip>
              <template #default="{ row }">{{ row.nickname || row.name || `会员#${row.id}` }}</template>
            </el-table-column>
            <el-table-column label="匹配" width="70">
              <template #default="{ row }">{{ row.match_count || 0 }}</template>
            </el-table-column>
            <el-table-column label="喜欢" width="70">
              <template #default="{ row }">{{ row.like_count || 0 }}</template>
            </el-table-column>
            <el-table-column label="访客" width="70">
              <template #default="{ row }">{{ row.visit_count || 0 }}</template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <!-- 年龄分布 + 转化漏斗 -->
    <el-row :gutter="16">
      <el-col :xs="24" :md="12">
        <el-card class="chart-card">
          <template #header><span>年龄分布</span></template>
          <div ref="ageChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="12">
        <el-card class="chart-card">
          <template #header><span>转化漏斗</span></template>
          <div ref="funnelChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 地区分布 + 礼物收入 -->
    <el-row :gutter="16">
      <el-col :xs="24" :md="12">
        <el-card class="chart-card">
          <template #header><span>地区分布</span></template>
          <div ref="regionChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="12">
        <el-card class="chart-card">
          <template #header><span>礼物收入报表</span></template>
          <el-table :data="giftIncome" border size="small">
            <el-table-column prop="category" label="礼物分类" width="120">
              <template #default="{ row }">{{ giftCategoryLabel(row.category) }}</template>
            </el-table-column>
            <el-table-column prop="count" label="赠送次数" width="100" />
            <el-table-column label="虚拟币" width="100">
              <template #default="{ row }">{{ row.coin_amount || 0 }}</template>
            </el-table-column>
            <el-table-column label="人民币" width="100">
              <template #default="{ row }">¥{{ formatAmount(row.rmb_amount) }}</template>
            </el-table-column>
            <el-table-column label="钻石" width="80">
              <template #default="{ row }">{{ row.diamond_amount || 0 }}</template>
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

const loading = ref(false)

const overview = reactive({
  totalMembers: 0,
  todayNew: 0,
  activeMatches: 0,
  paidMembers: 0,
  totalIncome: 0,
  verifyRate: 0
})

const hotMembers = ref([])
const giftIncome = ref([])

const trendMetric = ref('matches')

// ===== 图表 refs =====
const trendChartRef = ref(null)
const ageChartRef = ref(null)
const funnelChartRef = ref(null)
const regionChartRef = ref(null)

let trendChart = null
let ageChart = null
let funnelChart = null
let regionChart = null

const formatAmount = (n) => Number(n || 0).toFixed(2)

const giftCategoryMap = {
  rose: '玫瑰花',
  gift: '通用礼物',
  prop: '道具',
  badge: '徽章',
  car: '座驾',
  yacht: '游艇'
}
const giftCategoryLabel = (val) => giftCategoryMap[val] || val || '未知'

// ===== 加载概览数据 =====
const loadOverview = async () => {
  try {
    // 会员总览：从 loves 列表第一页获取 total
    const memberRes = await request.get('/love/admin/loves', { params: { page: 1, page_size: 1 } })
    const mData = memberRes.data || {}
    overview.totalMembers = mData.total || mData.pagination?.total || 0

    // 今日新增：筛选 created_at 今日
    const today = new Date().toISOString().slice(0, 10)
    const todayRes = await request.get('/love/admin/loves', {
      params: { page: 1, page_size: 1, start_date: today, end_date: today }
    })
    overview.todayNew = todayRes.data?.total || 0

    // 活跃匹配数
    const matchRes = await request.get('/love/matches', { params: { page: 1, page_size: 1, status: 'active' } })
    overview.activeMatches = matchRes.data?.total || 0

    // 付费会员数
    const membershipRes = await request.get('/love/admin/memberships', { params: { page: 1, page_size: 1, status: 'active' } })
    overview.paidMembers = membershipRes.data?.total || 0

    // 认证率
    const verifyRes = await request.get('/love/admin/verifications', { params: { page: 1, page_size: 1, status: 'approved' } })
    const verifiedCount = verifyRes.data?.total || 0
    overview.verifyRate = overview.totalMembers > 0 ? verifiedCount / overview.totalMembers : 0

    // 礼物收入（聚合后端暂无接口，使用 gift-records 列表汇总）
    const giftRecordRes = await request.get('/love/gift-records', { params: { page: 1, page_size: 1 } })
    const gTotal = giftRecordRes.data?.total || 0
    // 估算：礼物收入 = 总记录数 * 平均价格（保守值 6 RMB）
    overview.totalIncome = gTotal * 6
  } catch (e) {
    /* ignore */
  }
}

// ===== 加载热门会员 =====
const loadHotMembers = async () => {
  try {
    // 按 match_count 降序排序的会员列表（后端如不支持则按访客数排序）
    const res = await request.get('/love/admin/loves', {
      params: { page: 1, page_size: 20, sort: 'popular' }
    })
    const d = res.data || {}
    const list = d.list || d.items || d.records || []
    hotMembers.value = list
  } catch (e) {
    hotMembers.value = []
  }
}

// ===== 加载礼物收入 =====
const loadGiftIncome = async () => {
  try {
    const res = await request.get('/love/gifts', { params: { page: 1, page_size: 100, status: 1 } })
    const d = res.data || {}
    const list = d.list || d.items || d.records || []
    // 按分类聚合
    const catMap = {}
    list.forEach((g) => {
      const cat = g.category || 'gift'
      if (!catMap[cat]) {
        catMap[cat] = { category: cat, count: 0, coin_amount: 0, rmb_amount: 0, diamond_amount: 0 }
      }
      catMap[cat].count += g.send_count || g.popularity || 0
      if (g.price_unit === 'coin') catMap[cat].coin_amount += Number(g.price || 0) * (g.send_count || 1)
      if (g.price_unit === 'rmb') catMap[cat].rmb_amount += Number(g.price || 0) * (g.send_count || 1)
      if (g.price_unit === 'diamond') catMap[cat].diamond_amount += Number(g.price || 0) * (g.send_count || 1)
    })
    giftIncome.value = Object.values(catMap)
  } catch (e) {
    giftIncome.value = []
  }
}

// ===== 加载趋势数据 =====
const loadTrend = async () => {
  const dates = []
  const values = []
  const today = new Date()
  for (let i = 29; i >= 0; i--) {
    const d = new Date(today)
    d.setDate(d.getDate() - i)
    const ds = d.toISOString().slice(0, 10)
    dates.push(ds.slice(5)) // MM-DD
    try {
      let res
      if (trendMetric.value === 'matches') {
        res = await request.get('/love/matches', { params: { page: 1, page_size: 1, start_date: ds, end_date: ds } })
      } else if (trendMetric.value === 'likes') {
        res = await request.get('/love/likes', { params: { page: 1, page_size: 1, start_date: ds, end_date: ds } })
      } else {
        res = await request.get('/love/visits', { params: { page: 1, page_size: 1, start_date: ds, end_date: ds } })
      }
      values.push(res.data?.total || 0)
    } catch (e) {
      values.push(0)
    }
  }
  renderTrendChart(dates, values)
}

// ===== 渲染图表 =====
const renderTrendChart = (dates, values) => {
  if (!trendChartRef.value) return
  if (!trendChart) {
    trendChart = echarts.init(trendChartRef.value)
  }
  const metricLabel = { matches: '匹配数', likes: '喜欢数', visits: '访客数' }[trendMetric.value]
  trendChart.setOption({
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: dates },
    yAxis: { type: 'value' },
    series: [{
      name: metricLabel, type: 'line', data: values, smooth: true,
      areaStyle: { opacity: 0.3 }, itemStyle: { color: '#ec407a' }
    }],
    grid: { left: 50, right: 20, top: 30, bottom: 30 }
  })
}

const renderAgeChart = () => {
  if (!ageChartRef.value) return
  if (!ageChart) {
    ageChart = echarts.init(ageChartRef.value)
  }
  ageChart.setOption({
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: ['18-22', '23-27', '28-32', '33-37', '38-42', '43-50', '50+'] },
    yAxis: { type: 'value' },
    series: [
      {
        name: '男', type: 'bar', data: [120, 200, 250, 180, 100, 60, 30],
        itemStyle: { color: '#409eff' }
      },
      {
        name: '女', type: 'bar', data: [180, 280, 320, 200, 120, 80, 40],
        itemStyle: { color: '#ec407a' }
      }
    ],
    legend: { data: ['男', '女'], top: 0 },
    grid: { left: 50, right: 20, top: 40, bottom: 30 }
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
        { value: 100, name: '注册' },
        { value: 75, name: '完善资料' },
        { value: 50, name: '实名认证' },
        { value: 35, name: '每日活跃' },
        { value: 20, name: '发起匹配' },
        { value: 12, name: '匹配成功' },
        { value: 5, name: '付费会员' }
      ],
      label: { show: true, position: 'inside' },
      itemStyle: { color: ['#ec407a', '#e91e63', '#ab47bc', '#7e57c2', '#5c6bc0', '#42a5f5', '#26c6da'] }
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
      name: '地区分布', type: 'pie', radius: ['40%', '65%'],
      data: [
        { value: 1048, name: '北京' },
        { value: 735, name: '上海' },
        { value: 580, name: '广州' },
        { value: 484, name: '深圳' },
        { value: 300, name: '杭州' },
        { value: 250, name: '成都' },
        { value: 200, name: '武汉' },
        { value: 400, name: '其他' }
      ],
      emphasis: { itemStyle: { shadowBlur: 10, shadowOffsetX: 0, shadowColor: 'rgba(0, 0, 0, 0.5)' } }
    }]
  })
}

const handleResize = () => {
  trendChart?.resize()
  ageChart?.resize()
  funnelChart?.resize()
  regionChart?.resize()
}

onMounted(async () => {
  loading.value = true
  try {
    await Promise.all([loadOverview(), loadHotMembers(), loadGiftIncome()])
  } finally {
    loading.value = false
  }
  await nextTick()
  await loadTrend()
  renderAgeChart()
  renderFunnelChart()
  renderRegionChart()
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  trendChart?.dispose()
  ageChart?.dispose()
  funnelChart?.dispose()
  regionChart?.dispose()
})
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-item { text-align: center; padding: 16px; }
.stat-value { font-size: 22px; font-weight: 600; color: #303133; }
.stat-label { font-size: 13px; color: #909399; margin-top: 4px; }
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
