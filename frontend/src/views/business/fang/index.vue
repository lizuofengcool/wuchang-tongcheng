<template>
  <div class="app-container">
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #409eff"><el-icon :size="22"><House /></el-icon></div>
          <div class="stat-content"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">总房源</div></div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #67c23a"><el-icon :size="22"><CirclePlus /></el-icon></div>
          <div class="stat-content"><div class="stat-value">{{ stats.todayNew }}</div><div class="stat-label">今日新增</div></div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #e6a23c"><el-icon :size="22"><Clock /></el-icon></div>
          <div class="stat-content"><div class="stat-value">{{ stats.pendingAudit }}</div><div class="stat-label">待审核</div></div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #13c2c2"><el-icon :size="22"><Promotion /></el-icon></div>
          <div class="stat-content"><div class="stat-value">{{ stats.published }}</div><div class="stat-label">已发布</div></div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #722ed1"><el-icon :size="22"><TrendCharts /></el-icon></div>
          <div class="stat-content"><div class="stat-value">{{ stats.totalDeals }}</div><div class="stat-label">成交量</div></div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #f56c6c"><el-icon :size="22"><Warning /></el-icon></div>
          <div class="stat-content"><div class="stat-value">{{ stats.violation }}</div><div class="stat-label">违规数</div></div>
        </el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="标题/小区/发布者" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="filters.house_type" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="出售" value="sale" />
            <el-option label="出租" value="rent" />
          </el-select>
        </el-form-item>
        <el-form-item label="户型">
          <el-select v-model="filters.layout" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="一室" value="1" />
            <el-option label="二室" value="2" />
            <el-option label="三室" value="3" />
            <el-option label="四室+" value="4+" />
          </el-select>
        </el-form-item>
        <el-form-item label="价格区间">
          <el-input-number v-model="filters.min_price" :min="0" :controls="false" placeholder="最低" style="width: 100px" />
          <span style="margin: 0 4px">-</span>
          <el-input-number v-model="filters.max_price" :min="0" :controls="false" placeholder="最高" style="width: 100px" />
        </el-form-item>
        <el-form-item label="审核状态">
          <el-select v-model="filters.audit_status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="待审核" :value="0" />
            <el-option label="已通过" :value="1" />
            <el-option label="已拒绝" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="草稿" :value="0" />
            <el-option label="已发布" :value="1" />
            <el-option label="已下架" :value="2" />
            <el-option label="已成交" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar">
        <div class="toolbar-left">
          <el-button :icon="Refresh" @click="loadList">刷新</el-button>
          <el-button type="success" :icon="Check" :disabled="!selection.length" @click="onBatchAudit(1)">批量通过</el-button>
          <el-button type="warning" :icon="Bottom" :disabled="!selection.length" @click="onBatchStatus(2)">批量下架</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="list" border stripe @selection-change="onSelectionChange">
        <el-table-column type="selection" width="44" fixed="left" />
        <el-table-column prop="id" label="ID" width="70" fixed="left" />
        <el-table-column label="房源/价格" min-width="220">
          <template #default="{ row }">
            <div class="title-cell">
              <div class="title-text">
                <el-link type="primary" :underline="false" @click="openDetail(row)">{{ row.title }}</el-link>
                <el-tag v-if="row.house_type === 'rent'" type="primary" size="small">租</el-tag>
                <el-tag v-else type="success" size="small">售</el-tag>
              </div>
              <div class="title-desc">
                <span class="price">{{ formatPrice(row) }}</span>
                <span class="text-muted">·{{ row.community_name || row.area_name || '-' }}</span>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="layout" label="户型" width="90">
          <template #default="{ row }">{{ row.layout || '-' }}</template>
        </el-table-column>
        <el-table-column label="面积" width="80">
          <template #default="{ row }">{{ row.area ? row.area + '㎡' : '-' }}</template>
        </el-table-column>
        <el-table-column label="发布者" width="140">
          <template #default="{ row }">
            <div>{{ row.user_name || `用户#${row.user_id}` }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="view_count" label="浏览" width="70" />
        <el-table-column prop="fav_count" label="收藏" width="70" />
        <el-table-column label="审核" width="90">
          <template #default="{ row }">
            <el-tag :type="auditTagType(row.audit_status)" size="small">{{ auditText(row.audit_status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small" effect="plain">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="发布时间" width="160" prop="created_at" sortable>
          <template #default="{ row }">{{ formatTime(row.published_at || row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.audit_status === 0 || row.audit_status === 2" type="success" link size="small" @click="handleAudit(row, 1)">通过</el-button>
            <el-button v-if="row.audit_status === 0 || row.audit_status === 1" type="danger" link size="small" @click="handleAudit(row, 2)">拒绝</el-button>
            <el-dropdown trigger="click" @command="(cmd) => handleStatusCommand(row, cmd)">
              <el-button type="warning" link size="small">更多<el-icon class="el-icon--right"><ArrowDown /></el-icon></el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item :command="1" :disabled="row.status === 1">发布</el-dropdown-item>
                  <el-dropdown-item :command="2" :disabled="row.status === 2">下架</el-dropdown-item>
                  <el-dropdown-item :command="'delete'" divided>删除</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50, 100]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <el-dialog v-model="detailVisible" title="房源详情" width="900px" @close="onDetailClose" destroy-on-close>
      <div v-loading="detailLoading">
        <el-tabs v-if="detail" v-model="detailTab">
          <el-tab-pane label="基本信息" name="basic">
            <el-descriptions :column="3" border>
              <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
              <el-descriptions-item label="标题">{{ detail.title }}</el-descriptions-item>
              <el-descriptions-item label="类型">{{ detail.house_type === 'rent' ? '出租' : '出售' }}</el-descriptions-item>
              <el-descriptions-item label="价格">{{ formatPrice(detail) }}</el-descriptions-item>
              <el-descriptions-item label="户型">{{ detail.layout || '-' }}</el-descriptions-item>
              <el-descriptions-item label="面积">{{ detail.area ? detail.area + '㎡' : '-' }}</el-descriptions-item>
              <el-descriptions-item label="小区">{{ detail.community_name || '-' }}</el-descriptions-item>
              <el-descriptions-item label="楼层">{{ detail.floor || '-' }}</el-descriptions-item>
              <el-descriptions-item label="朝向">{{ detail.orientation || '-' }}</el-descriptions-item>
              <el-descriptions-item label="审核状态"><el-tag :type="auditTagType(detail.audit_status)" size="small">{{ auditText(detail.audit_status) }}</el-tag></el-descriptions-item>
              <el-descriptions-item label="状态"><el-tag :type="statusTagType(detail.status)" size="small" effect="plain">{{ statusText(detail.status) }}</el-tag></el-descriptions-item>
              <el-descriptions-item label="发布者">{{ detail.user_name || `用户#${detail.user_id}` }}</el-descriptions-item>
              <el-descriptions-item label="地址" :span="3">{{ detail.address || '-' }}</el-descriptions-item>
              <el-descriptions-item v-if="detail.description" label="描述" :span="3"><div class="content-box">{{ detail.description }}</div></el-descriptions-item>
              <el-descriptions-item label="浏览量">{{ detail.view_count || 0 }}</el-descriptions-item>
              <el-descriptions-item label="收藏量">{{ detail.fav_count || 0 }}</el-descriptions-item>
              <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
            </el-descriptions>
          </el-tab-pane>
          <el-tab-pane :label="`举报历史 (${reports.length})`" name="reports">
            <el-table :data="reports" border size="small">
              <el-table-column prop="report_no" label="举报单号" width="160" />
              <el-table-column prop="reason" label="原因" min-width="180" show-overflow-tooltip />
              <el-table-column label="状态" width="100"><template #default="{ row }"><el-tag :type="row.status === 0 ? 'warning' : 'info'" size="small">{{ row.status === 0 ? '待处理' : '已处理' }}</el-tag></template></el-table-column>
              <el-table-column label="时间" width="160"><template #default="{ row }">{{ formatTime(row.created_at) }}</template></el-table-column>
            </el-table>
            <div v-if="!reports.length" class="empty-text">暂无举报记录</div>
          </el-tab-pane>
        </el-tabs>
      </div>
      <template #footer><el-button @click="detailVisible = false">关闭</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search, ArrowDown, Check, Bottom, House, CirclePlus, Clock, Promotion, Warning, TrendCharts } from '@element-plus/icons-vue'
import { adminListHouses, adminGetHouse, auditHouse, adminUpdateHouseStatus, deleteHouse, getOverviewStats, adminListReports } from '@/api/house'
import { formatTime } from '@/utils/format'

const router = useRouter()
const loading = ref(false)
const page = ref(1); const pageSize = ref(20); const total = ref(0); const list = ref([])
const selection = ref([])
const filters = reactive({ keyword: '', house_type: '', layout: '', min_price: undefined, max_price: undefined, audit_status: null, status: null, dateRange: null })
const stats = reactive({ total: 0, todayNew: 0, pendingAudit: 0, published: 0, totalDeals: 0, violation: 0 })

const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')
const statusText = (s) => ({ 0: '草稿', 1: '已发布', 2: '已下架', 3: '已成交' }[s] || '-')
const statusTagType = (s) => ({ 0: 'info', 1: 'success', 2: 'warning', 3: 'primary' }[s] || 'info')

const formatPrice = (row) => {
  const p = Number(row.price || 0)
  if (!p) return '面议'
  if (row.house_type === 'rent') return `${p}元/月`
  return `${(p / 10000).toFixed(1)}万`
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => { Object.assign(filters, { keyword: '', house_type: '', layout: '', min_price: undefined, max_price: undefined, audit_status: null, status: null, dateRange: null }); page.value = 1; loadList() }
const onSelectionChange = (rows) => { selection.value = rows }

const loadStats = async () => {
  try { const res = await getOverviewStats(); const d = res.data || {}; Object.assign(stats, { total: d.total_houses || 0, todayNew: d.today_new || 0, pendingAudit: d.pending_audit || 0, published: d.published || 0, totalDeals: d.total_deals || 0, violation: d.violation || 0 }) } catch (e) { /* */ }
}

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value, page_size: pageSize.value,
      keyword: filters.keyword || undefined, house_type: filters.house_type || undefined,
      layout: filters.layout || undefined, min_price: filters.min_price || undefined, max_price: filters.max_price || undefined,
      audit_status: filters.audit_status === null || filters.audit_status === '' ? undefined : filters.audit_status,
      status: filters.status === null || filters.status === '' ? undefined : filters.status
    }
    const res = await adminListHouses(params)
    const data = res.data || {}
    list.value = data.list || []; total.value = data.total || 0
    if (data.stats) Object.assign(stats, data.stats)
  } catch (e) { list.value = []; total.value = 0 } finally { loading.value = false }
}

const detailVisible = ref(false); const detailLoading = ref(false); const detail = ref(null)
const reports = ref([]); const detailTab = ref('basic')
const openDetail = async (row) => {
  detail.value = row; reports.value = []; detailTab.value = 'basic'; detailVisible.value = true; detailLoading.value = true
  try {
    const [dRes, rRes] = await Promise.all([
      adminGetHouse(row.id),
      adminListReports({ target_id: row.id, page: 1, page_size: 20 }).catch(() => ({ data: { list: [] } }))
    ])
    if (dRes.data) detail.value = dRes.data
    reports.value = rRes.data?.list || []
  } catch (e) { /* */ } finally { detailLoading.value = false }
}
const onDetailClose = () => { detail.value = null; reports.value = [] }

const handleAudit = async (row, auditStatus) => {
  try {
    if (auditStatus === 2) {
      const { value } = await ElMessageBox.prompt('请输入拒绝原因（可选）', '拒绝审核', { confirmButtonText: '确定拒绝', cancelButtonText: '取消', inputType: 'textarea' })
      await auditHouse(row.id, { audit_status: auditStatus, audit_reason: value || '' })
    } else {
      await ElMessageBox.confirm(`确定通过房源 "${row.title}" 的审核吗？`, '提示', { type: 'warning' })
      await auditHouse(row.id, { audit_status: auditStatus })
    }
    ElMessage.success('操作成功'); await loadList()
  } catch (e) { /* */ }
}

const handleStatusCommand = async (row, cmd) => {
  try {
    if (cmd === 'delete') {
      await ElMessageBox.confirm(`确定删除房源 "${row.title}" 吗？`, '危险操作', { type: 'error' })
      await deleteHouse(row.id)
      ElMessage.success('已删除'); await loadList(); return
    }
    const label = statusText(cmd)
    await ElMessageBox.confirm(`确定将房源设为「${label}」吗？`, '提示', { type: 'warning' })
    await adminUpdateHouseStatus(row.id, cmd)
    ElMessage.success('状态更新成功'); await loadList()
  } catch (e) { /* */ }
}

const onBatchAudit = async (auditStatus) => {
  try {
    await ElMessageBox.confirm(`确认批量审核通过 ${selection.value.length} 个房源？`, '批量审核', { type: 'warning' })
    await Promise.all(selection.value.map((r) => auditHouse(r.id, { audit_status: auditStatus })))
    ElMessage.success('批量审核完成'); await loadList()
  } catch (e) { /* */ }
}

const onBatchStatus = async (status) => {
  try {
    const label = statusText(status)
    await ElMessageBox.confirm(`确认批量将 ${selection.value.length} 个房源设为「${label}」？`, '批量状态变更', { type: 'warning' })
    await Promise.all(selection.value.map((r) => adminUpdateHouseStatus(r.id, status)))
    ElMessage.success('批量操作完成'); await loadList()
  } catch (e) { /* */ }
}

onMounted(async () => { await Promise.all([loadStats(), loadList()]) })
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-card { display: flex; align-items: center; }
.stat-card :deep(.el-card__body) { display: flex; align-items: center; gap: 14px; padding: 16px; width: 100%; }
.stat-icon { width: 44px; height: 44px; border-radius: 8px; display: flex; align-items: center; justify-content: center; color: #fff; flex-shrink: 0; }
.stat-value { font-size: 22px; font-weight: 600; color: #303133; }
.stat-label { font-size: 13px; color: #909399; margin-top: 2px; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; margin-bottom: 12px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.title-cell .title-text { display: flex; align-items: center; gap: 6px; }
.title-cell .price { color: #f56c6c; font-weight: 600; }
.text-muted { color: #909399; }
.title-desc { font-size: 12px; margin-top: 2px; }
.content-box { white-space: pre-wrap; line-height: 1.6; }
.empty-text { text-align: center; color: #999; padding: 20px; }
</style>
