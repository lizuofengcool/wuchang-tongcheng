<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #409eff"><el-icon :size="22"><Van /></el-icon></div>
          <div class="stat-content"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">总车源</div></div>
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
          <div class="stat-content"><div class="stat-value">{{ stats.totalViews }}</div><div class="stat-label">总浏览</div></div>
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
          <el-input v-model="filters.keyword" placeholder="标题/品牌/发布者" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="品牌">
          <el-input v-model="filters.brand" placeholder="品牌" clearable style="width: 120px" @keyup.enter="onSearch" />
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
            <el-option label="已售出" :value="3" />
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
        <el-table-column label="车源/价格" min-width="220">
          <template #default="{ row }">
            <div class="title-cell">
              <div class="title-text">
                <el-link type="primary" :underline="'never'" @click="openDetail(row)">{{ row.title }}</el-link>
                <el-tag v-if="row.is_real_car" type="success" size="small">真车</el-tag>
              </div>
              <div class="title-desc">
                <span class="price">¥{{ Number(row.price || 0).toFixed(2) }}万</span>
                <span class="text-muted">·{{ row.brand_name || row.brand }} {{ row.model_name || row.model }}</span>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="车系" width="100">
          <template #default="{ row }">{{ row.series || row.model_name || '-' }}</template>
        </el-table-column>
        <el-table-column label="年份" width="80">
          <template #default="{ row }">{{ row.year || '-' }}</template>
        </el-table-column>
        <el-table-column label="里程" width="90">
          <template #default="{ row }">{{ row.mileage ? row.mileage + '万km' : '-' }}</template>
        </el-table-column>
        <el-table-column label="发布者" width="140">
          <template #default="{ row }">{{ row.user_name || `用户#${row.user_id}` }}</template>
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
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.audit_status === 0 || row.audit_status === 2" type="success" link size="small" @click="handleAudit(row, 1)">通过</el-button>
            <el-button v-if="row.audit_status === 0 || row.audit_status === 1" type="danger" link size="small" @click="handleAudit(row, 2)">拒绝</el-button>
            <el-button v-if="row.status === 1" type="warning" link size="small" @click="handleStatus(row, 2)">下架</el-button>
            <el-button v-if="row.status === 2 || row.status === 0" type="primary" link size="small" @click="handleStatus(row, 1)">上架</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <el-dialog v-model="detailVisible" title="车源详情" width="800px" destroy-on-close>
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="标题">{{ detail.title }}</el-descriptions-item>
        <el-descriptions-item label="品牌">{{ detail.brand_name || detail.brand }}</el-descriptions-item>
        <el-descriptions-item label="车系">{{ detail.series || detail.model_name }}</el-descriptions-item>
        <el-descriptions-item label="车型">{{ detail.model }}</el-descriptions-item>
        <el-descriptions-item label="年份">{{ detail.year }}</el-descriptions-item>
        <el-descriptions-item label="里程">{{ detail.mileage }}万公里</el-descriptions-item>
        <el-descriptions-item label="价格">¥{{ Number(detail.price || 0).toFixed(2) }}万</el-descriptions-item>
        <el-descriptions-item label="颜色">{{ detail.color || '-' }}</el-descriptions-item>
        <el-descriptions-item label="排量">{{ detail.displacement || '-' }}</el-descriptions-item>
        <el-descriptions-item label="变速箱">{{ detail.gearbox || '-' }}</el-descriptions-item>
        <el-descriptions-item label="燃油类型">{{ detail.fuel_type || '-' }}</el-descriptions-item>
        <el-descriptions-item label="上牌时间">{{ formatTime(detail.first_plate_date, 'YYYY-MM-DD') }}</el-descriptions-item>
        <el-descriptions-item label="排放标准">{{ detail.emission_standard || '-' }}</el-descriptions-item>
        <el-descriptions-item label="城市">{{ detail.city || '-' }}</el-descriptions-item>
        <el-descriptions-item label="地址">{{ detail.address || '-' }}</el-descriptions-item>
        <el-descriptions-item label="发布者">{{ detail.user_name || `#${detail.user_id}` }}</el-descriptions-item>
        <el-descriptions-item label="联系电话">{{ detail.contact_phone || '-' }}</el-descriptions-item>
        <el-descriptions-item label="真车验证">{{ detail.is_real_car ? '已验证' : '未验证' }}</el-descriptions-item>
        <el-descriptions-item label="审核状态">{{ auditText(detail.audit_status) }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ statusText(detail.status) }}</el-descriptions-item>
        <el-descriptions-item label="浏览数">{{ detail.view_count }}</el-descriptions-item>
        <el-descriptions-item label="收藏数">{{ detail.fav_count }}</el-descriptions-item>
        <el-descriptions-item label="发布时间">{{ formatTime(detail.published_at) }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.description" label="描述" :span="2">{{ detail.description }}</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
        <el-button v-if="detail && detail.audit_status === 0" type="success" @click="handleAudit(detail, 1)">审核通过</el-button>
        <el-button v-if="detail && detail.audit_status === 0" type="danger" @click="handleAudit(detail, 2)">审核拒绝</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Van, CirclePlus, Clock, Promotion, TrendCharts, Warning, Search, RefreshLeft, Refresh, Check, Bottom } from '@element-plus/icons-vue'
import { adminListCars, adminGetCar, auditCar, adminUpdateCarStatus } from '@/api/car'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const selection = ref([])
const filters = reactive({ keyword: '', brand: '', min_price: null, max_price: null, audit_status: '', status: '' })

const detailVisible = ref(false)
const detail = ref(null)

const stats = ref({ total: 0, todayNew: 0, pendingAudit: 0, published: 0, totalViews: 0, violation: 0 })

const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')
const statusText = (s) => ({ 0: '草稿', 1: '已发布', 2: '已下架', 3: '已售出' }[s] || '-')
const statusTagType = (s) => ({ 0: 'info', 1: 'success', 2: 'warning', 3: 'primary' }[s] || 'info')

const buildParams = () => {
  const p = { page: page.value, page_size: pageSize.value }
  if (filters.keyword) p.keyword = filters.keyword
  if (filters.brand) p.brand = filters.brand
  if (filters.min_price != null) p.min_price = filters.min_price
  if (filters.max_price != null) p.max_price = filters.max_price
  if (filters.audit_status !== '' && filters.audit_status !== null) p.audit_status = filters.audit_status
  if (filters.status !== '' && filters.status !== null) p.status = filters.status
  return p
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await adminListCars(buildParams())
    const d = res.data || {}
    list.value = d.list || []
    total.value = d.total || 0
    stats.value = {
      total: d.total || 0,
      todayNew: d.today_new || 0,
      pendingAudit: d.pending_audit || 0,
      published: d.published || 0,
      totalViews: d.total_views || 0,
      violation: d.violation || 0
    }
  } catch (e) { list.value = [] } finally { loading.value = false }
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => { Object.assign(filters, { keyword: '', brand: '', min_price: null, max_price: null, audit_status: '', status: '' }); onSearch() }
const onSelectionChange = (rows) => { selection.value = rows }

const openDetail = async (row) => {
  try {
    const res = await adminGetCar(row.id)
    detail.value = res.data || row
  } catch (e) { detail.value = row }
  detailVisible.value = true
}

const handleAudit = async (row, status) => {
  try {
    if (status === 2) {
      const { value } = await ElMessageBox.prompt('请输入拒绝原因', '审核拒绝', { type: 'warning', inputType: 'textarea' })
      await auditCar(row.id, { audit_status: status, audit_reason: value })
    } else {
      await auditCar(row.id, { audit_status: status })
    }
    ElMessage.success('操作成功')
    if (detail.value && detail.value.id === row.id) detailVisible.value = false
    loadList()
  } catch (e) {
    if (e !== 'cancel' && e?.message) ElMessage.error(e.message)
  }
}

const handleStatus = async (row, status) => {
  try {
    await adminUpdateCarStatus(row.id, { status })
    ElMessage.success('操作成功')
    loadList()
  } catch (e) {
    if (e?.message) ElMessage.error(e.message)
  }
}

const onBatchAudit = async (status) => {
  try {
    await ElMessageBox.confirm(`确认批量审核通过 ${selection.value.length} 条车源？`, '提示', { type: 'warning' })
    await Promise.all(selection.value.map((r) => auditCar(r.id, { audit_status: status })))
    ElMessage.success('批量审核完成')
    loadList()
  } catch (e) {
    if (e !== 'cancel' && e?.message) ElMessage.error(e.message)
  }
}

const onBatchStatus = async (status) => {
  try {
    await ElMessageBox.confirm(`确认批量下架 ${selection.value.length} 条车源？`, '提示', { type: 'warning' })
    await Promise.all(selection.value.map((r) => adminUpdateCarStatus(r.id, { status })))
    ElMessage.success('批量操作完成')
    loadList()
  } catch (e) {
    if (e !== 'cancel' && e?.message) ElMessage.error(e.message)
  }
}

onMounted(() => loadList())
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-card { display: flex; align-items: center; }
.stat-card :deep(.el-card__body) { display: flex; align-items: center; gap: 12px; padding: 14px 16px; width: 100%; }
.stat-icon { width: 44px; height: 44px; border-radius: 8px; display: flex; align-items: center; justify-content: center; color: #fff; flex-shrink: 0; }
.stat-content { flex: 1; }
.stat-value { font-size: 22px; font-weight: 600; color: #303133; }
.stat-label { font-size: 12px; color: #909399; margin-top: 2px; }
.page-card { background: #fff; padding: 16px; border-radius: 4px; box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04); }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.text-muted { color: #909399; font-size: 12px; }
.title-cell { display: flex; flex-direction: column; gap: 4px; }
.title-text { display: flex; align-items: center; gap: 6px; }
.title-desc { font-size: 12px; color: #606266; }
.title-desc .price { color: #f56c6c; font-weight: 600; margin-right: 6px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
