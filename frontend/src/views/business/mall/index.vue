<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #409eff"><el-icon :size="22"><Shop /></el-icon></div>
          <div class="stat-content"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">总店铺数</div></div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #e6a23c"><el-icon :size="22"><Clock /></el-icon></div>
          <div class="stat-content"><div class="stat-value">{{ stats.pendingAudit }}</div><div class="stat-label">待审核</div></div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #67c23a"><el-icon :size="22"><CircleCheck /></el-icon></div>
          <div class="stat-content"><div class="stat-value">{{ stats.audited }}</div><div class="stat-label">已通过</div></div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #f56c6c"><el-icon :size="22"><Warning /></el-icon></div>
          <div class="stat-content"><div class="stat-value">{{ stats.rejected }}</div><div class="stat-label">已拒绝</div></div>
        </el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <!-- 筛选区 -->
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="店铺名/联系电话" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="审核状态">
          <el-select v-model="filters.audit_status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="待审核" :value="0" />
            <el-option label="已通过" :value="1" />
            <el-option label="已拒绝" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="店铺状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="关闭" :value="0" />
            <el-option label="营业中" :value="1" />
            <el-option label="休息中" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="时间范围">
          <el-date-picker v-model="filters.dateRange" type="daterange" range-separator="至" start-placeholder="开始" end-placeholder="结束" value-format="YYYY-MM-DD" style="width: 240px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button :icon="Refresh" @click="loadList">刷新</el-button>
        </div>
      </div>

      <!-- 表格 -->
      <el-table v-loading="loading" :data="list" border stripe @sort-change="onSortChange">
        <el-table-column prop="id" label="ID" width="70" sortable="custom" fixed="left" />
        <el-table-column label="店铺名" min-width="200">
          <template #default="{ row }">
            <el-link type="primary" :underline="false" @click="openDetail(row)">{{ row.name }}</el-link>
            <div class="text-muted text-xs">{{ row.short_name || '-' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="店主" width="140">
          <template #default="{ row }">
            <div>{{ row.owner_name || '-' }}</div>
            <div class="text-muted text-xs">{{ row.contact_phone || '-' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="分类" width="120">
          <template #default="{ row }">{{ row.category_name || row.category_id || '-' }}</template>
        </el-table-column>
        <el-table-column label="地区" width="100">
          <template #default="{ row }">{{ row.region_id || '-' }}</template>
        </el-table-column>
        <el-table-column label="商品数" width="90" prop="product_count" />
        <el-table-column label="销量" width="80" prop="sale_count" />
        <el-table-column label="浏览" width="80" prop="view_count" />
        <el-table-column label="评分" width="100">
          <template #default="{ row }">
            <el-rate v-if="row.rating" :model-value="row.rating" disabled size="small" />
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
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
        <el-table-column label="创建时间" width="160" prop="created_at" sortable="custom">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.audit_status === 0 || row.audit_status === 2" type="success" link size="small" @click="handleAudit(row, 1)">通过</el-button>
            <el-button v-if="row.audit_status === 0 || row.audit_status === 1" type="danger" link size="small" @click="handleAudit(row, 2)">拒绝</el-button>
            <el-button type="warning" link size="small" @click="openPromotion(row)">推广</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50, 100]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <!-- 推广配置弹窗 -->
    <el-dialog v-model="promotionVisible" title="推广配置" width="520px" destroy-on-close>
      <el-form :model="promotionForm" label-width="110px">
        <el-form-item label="是否推荐">
          <el-switch v-model="promotionForm.is_recommended" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="是否置顶">
          <el-switch v-model="promotionForm.is_pinned" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="推广权重">
          <el-input-number v-model="promotionForm.promotion_weight" :min="0" :max="9999" />
        </el-form-item>
        <el-form-item label="推广开始">
          <el-date-picker v-model="promotionForm.promotion_start" type="datetime" value-format="YYYY-MM-DD HH:mm:ss" style="width: 100%" />
        </el-form-item>
        <el-form-item label="推广结束">
          <el-date-picker v-model="promotionForm.promotion_end" type="datetime" value-format="YYYY-MM-DD HH:mm:ss" style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="promotionVisible = false">取消</el-button>
        <el-button type="primary" :loading="promotionLoading" @click="onPromotionSubmit">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search, Shop, Clock, CircleCheck, Warning } from '@element-plus/icons-vue'
import { formatTime } from '@/utils/format'
import { getMallShopList, auditMallShop, updateMallShopPromotion } from '@/api/mall'

const router = useRouter()

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const sortField = ref('created_at')
const sortOrder = ref('descending')

const stats = reactive({ total: 0, pendingAudit: 0, audited: 0, rejected: 0 })

const filters = reactive({ keyword: '', audit_status: null, status: null, dateRange: null })

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.keyword = ''
  filters.audit_status = null
  filters.status = null
  filters.dateRange = null
  page.value = 1
  loadList()
}
const onSortChange = ({ prop, order }) => {
  sortField.value = prop || 'created_at'
  sortOrder.value = order || 'descending'
  loadList()
}

const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')
const statusText = (s) => ({ 0: '关闭', 1: '营业中', 2: '休息中' }[s] || '-')
const statusTagType = (s) => ({ 0: 'danger', 1: 'success', 2: 'warning' }[s] || 'info')

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      keyword: filters.keyword.trim() || undefined,
      audit_status: filters.audit_status === null || filters.audit_status === '' ? undefined : filters.audit_status,
      status: filters.status === null || filters.status === '' ? undefined : filters.status,
      sort: sortField.value,
      order: sortOrder.value
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await getMallShopList(params)
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
    calcStats()
  } catch (e) {
    ElMessage.error('加载店铺列表失败')
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const calcStats = () => {
  const all = list.value || []
  stats.total = total.value
  stats.pendingAudit = all.filter((r) => r.audit_status === 0).length
  stats.audited = all.filter((r) => r.audit_status === 1).length
  stats.rejected = all.filter((r) => r.audit_status === 2).length
}

const openDetail = (row) => {
  router.push(`/business/mall/shop/${row.id}`)
}

const handleAudit = async (row, auditStatus) => {
  try {
    if (auditStatus === 2) {
      const { value } = await ElMessageBox.prompt('请输入拒绝原因（可选）', '拒绝审核', {
        confirmButtonText: '确定', cancelButtonText: '取消', inputType: 'textarea', inputPlaceholder: '拒绝原因'
      })
      await auditMallShop(row.id, { audit_status: auditStatus, audit_reason: value || '' })
    } else {
      await ElMessageBox.confirm(`确定通过店铺 "${row.name}" 的审核吗？`, '提示', { type: 'warning' })
      await auditMallShop(row.id, { audit_status: auditStatus })
    }
    ElMessage.success('操作成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

const promotionVisible = ref(false)
const promotionLoading = ref(false)
const promotionForm = reactive({
  id: null, is_recommended: 0, is_pinned: 0, promotion_weight: 0,
  promotion_start: null, promotion_end: null
})

const openPromotion = (row) => {
  Object.assign(promotionForm, {
    id: row.id,
    is_recommended: row.is_recommended || 0,
    is_pinned: row.is_pinned || 0,
    promotion_weight: row.promotion_weight || 0,
    promotion_start: row.promotion_start || null,
    promotion_end: row.promotion_end || null
  })
  promotionVisible.value = true
}

const onPromotionSubmit = async () => {
  try {
    promotionLoading.value = true
    await updateMallShopPromotion(promotionForm.id, { ...promotionForm })
    ElMessage.success('推广配置已更新')
    promotionVisible.value = false
    await loadList()
  } catch (e) {
    ElMessage.error('更新失败')
  } finally {
    promotionLoading.value = false
  }
}

onMounted(() => loadList())
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-card { display: flex; align-items: center; }
.stat-card :deep(.el-card__body) { display: flex; align-items: center; gap: 14px; padding: 16px; width: 100%; }
.stat-icon { width: 44px; height: 44px; border-radius: 8px; display: flex; align-items: center; justify-content: center; color: #fff; flex-shrink: 0; }
.stat-content { flex: 1; min-width: 0; }
.stat-value { font-size: 22px; font-weight: 600; color: #303133; }
.stat-label { font-size: 12px; color: #909399; margin-top: 2px; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 8px; margin-bottom: 12px; }
.toolbar-left { display: flex; gap: 8px; flex-wrap: wrap; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.text-muted { color: #909399; }
.text-xs { font-size: 12px; }
</style>
