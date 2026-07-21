<template>
  <div class="app-container">
    <!-- 举报统计 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">举报总数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.pending }}</div><div class="stat-label">待处理</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.processed }}</div><div class="stat-label">已处理</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ stats.distinct_targets }}</div><div class="stat-label">被举报对象数</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="举报人ID">
          <el-input-number v-model="filters.reporter_id" :controls="false" :min="1" placeholder="举报人ID" style="width: 140px" />
        </el-form-item>
        <el-form-item label="目标类型">
          <el-select v-model="filters.target_type" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option v-for="(label, val) in targetTypeMap" :key="val" :label="label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item label="目标ID">
          <el-input-number v-model="filters.target_id" :controls="false" :min="1" placeholder="目标ID" style="width: 140px" />
        </el-form-item>
        <el-form-item label="举报类型">
          <el-select v-model="filters.report_type" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option v-for="(label, val) in reportTypeMap" :key="val" :label="label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item label="处理状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 130px" @change="onSearch">
            <el-option label="待处理" :value="0" />
            <el-option label="处理中" :value="1" />
            <el-option label="已处理" :value="2" />
            <el-option label="已驳回" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar">
        <el-button :icon="Refresh" @click="loadList">刷新</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe @sort-change="onSortChange">
        <el-table-column prop="id" label="ID" width="70" sortable="custom" fixed="left" />
        <el-table-column label="目标" min-width="200">
          <template #default="{ row }">
            <el-link type="primary" :underline="false" @click="openDetail(row)">
              [{{ targetTypeText(row.target_type) }}] {{ row.target_title || `#${row.target_id}` }}
            </el-link>
          </template>
        </el-table-column>
        <el-table-column label="举报类型" width="120">
          <template #default="{ row }">
            <el-tag size="small" :type="reportTypeTagType(row.report_type)">{{ reportTypeText(row.report_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="举报人" width="120">
          <template #default="{ row }">{{ row.reporter_name || `#${row.reporter_id}` }}</template>
        </el-table-column>
        <el-table-column label="店铺ID" width="90" prop="shop_id" />
        <el-table-column label="处理状态" width="110">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="处理人" width="120">
          <template #default="{ row }">{{ row.handler_name || '-' }}</template>
        </el-table-column>
        <el-table-column label="举报时间" width="160" prop="created_at" sortable="custom">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status === 0 || row.status === 1" type="success" link size="small" @click="openProcess(row, 2)">处理完成</el-button>
            <el-button v-if="row.status === 0 || row.status === 1" type="warning" link size="small" @click="openProcess(row, 3)">驳回</el-button>
            <el-button type="danger" link size="small" @click="onDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50, 100]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="举报详情" width="720px" destroy-on-close>
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="举报类型">
          <el-tag size="small" :type="reportTypeTagType(detail.report_type)">{{ reportTypeText(detail.report_type) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="目标类型">{{ targetTypeText(detail.target_type) }}</el-descriptions-item>
        <el-descriptions-item label="目标ID">{{ detail.target_id }}</el-descriptions-item>
        <el-descriptions-item label="目标标题" :span="2">{{ detail.target_title || '-' }}</el-descriptions-item>
        <el-descriptions-item label="店铺ID">{{ detail.shop_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="举报人ID">{{ detail.reporter_id }}</el-descriptions-item>
        <el-descriptions-item label="举报人名">{{ detail.reporter_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="举报人联系方式">{{ detail.reporter_contact || '-' }}</el-descriptions-item>
        <el-descriptions-item label="举报原因" :span="2">{{ detail.reason || '-' }}</el-descriptions-item>
        <el-descriptions-item label="举报描述" :span="2">{{ detail.description || '-' }}</el-descriptions-item>
        <el-descriptions-item label="举报图片" :span="2">
          <div v-if="detail.images && detail.images.length" class="image-list">
            <el-image v-for="(img, i) in detail.images" :key="i" :src="img" :preview-src-list="detail.images" fit="cover" class="report-thumb" />
          </div>
          <span v-else>-</span>
        </el-descriptions-item>
        <el-descriptions-item label="处理状态">
          <el-tag :type="statusTagType(detail.status)" size="small">{{ statusText(detail.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="处理人">{{ detail.handler_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="处理意见" :span="2">{{ detail.handle_remark || detail.process_remark || '-' }}</el-descriptions-item>
        <el-descriptions-item label="处理时间">{{ formatTime(detail.handled_at || detail.processed_at) }}</el-descriptions-item>
        <el-descriptions-item label="举报时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>

    <!-- 处理举报弹窗 -->
    <el-dialog v-model="processVisible" :title="processForm.status === 2 ? '处理举报' : '驳回举报'" width="480px" destroy-on-close>
      <el-form :model="processForm" label-width="100px">
        <el-form-item label="举报ID">{{ processForm.id }}</el-form-item>
        <el-form-item label="目标">[{{ targetTypeText(processForm.target_type) }}] {{ processForm.target_title || `#${processForm.target_id}` }}</el-form-item>
        <el-form-item label="处理结果">
          <el-tag :type="processForm.status === 2 ? 'success' : 'warning'">{{ processForm.status === 2 ? '处理完成' : '驳回' }}</el-tag>
        </el-form-item>
        <el-form-item label="处理意见">
          <el-input v-model="processForm.handle_remark" type="textarea" :rows="4" placeholder="处理意见/备注" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="processVisible = false">取消</el-button>
        <el-button type="primary" :loading="processLoading" @click="onProcessSubmit">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search } from '@element-plus/icons-vue'
import { formatTime } from '@/utils/format'
import { getMallReportList, getMallReportDetail, processMallReport, deleteMallReport, getMallReportStats } from '@/api/mall'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const sortField = ref('created_at')
const sortOrder = ref('descending')

const stats = reactive({ total: 0, pending: 0, processed: 0, distinct_targets: 0 })

const filters = reactive({
  reporter_id: null, target_type: '', target_id: null,
  report_type: '', status: null
})

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.reporter_id = null
  filters.target_type = ''
  filters.target_id = null
  filters.report_type = ''
  filters.status = null
  page.value = 1
  loadList()
}
const onSortChange = ({ prop, order }) => {
  sortField.value = prop || 'created_at'
  sortOrder.value = order || 'descending'
  loadList()
}

const targetTypeMap = {
  product: '商品', shop: '店铺', review: '评价', order: '订单', user: '用户', other: '其他'
}
const targetTypeText = (t) => targetTypeMap[t] || t || '-'

const reportTypeMap = {
  fake: '假冒伪劣', illegal: '违法违规', spam: '垃圾广告',
  infringement: '侵权', harassment: '骚扰', other: '其他'
}
const reportTypeText = (t) => reportTypeMap[t] || t || '-'
const reportTypeTagType = (t) => ({
  fake: 'danger', illegal: 'danger', spam: 'warning',
  infringement: 'warning', harassment: 'warning', other: 'info'
}[t] || 'info')

const statusText = (s) => ({ 0: '待处理', 1: '处理中', 2: '已处理', 3: '已驳回' }[s] ?? '-')
const statusTagType = (s) => ({ 0: 'warning', 1: 'primary', 2: 'success', 3: 'info' }[s] || 'info')

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      reporter_id: filters.reporter_id || undefined,
      target_type: filters.target_type || undefined,
      target_id: filters.target_id || undefined,
      report_type: filters.report_type || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status,
      sort: sortField.value,
      order: sortOrder.value
    }
    // 后端 List 返回 { list, pagination, stats } 特殊结构
    const res = await getMallReportList(params)
    const data = res.data || {}
    list.value = data.list || []
    const pagination = data.pagination || {}
    total.value = pagination.total || data.total || list.value.length
    // 若后端返回了 stats 字段则直接使用
    if (data.stats) {
      Object.assign(stats, {
        total: stats.total || pagination.total || list.value.length,
        ...data.stats
      })
    } else {
      computeStats()
    }
  } catch (e) {
    ElMessage.error('加载举报列表失败')
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const computeStats = () => {
  const total = list.value.length
  const pending = list.value.filter((r) => r.status === 0 || r.status === 1).length
  const processed = list.value.filter((r) => r.status === 2 || r.status === 3).length
  const distinctTargets = new Set(list.value.map((r) => `${r.target_type}_${r.target_id}`)).size
  Object.assign(stats, { total, pending, processed, distinct_targets: distinctTargets })
}

const detailVisible = ref(false)
const detail = ref(null)

const openDetail = async (row) => {
  try {
    const res = await getMallReportDetail(row.id)
    detail.value = res.data || null
    detailVisible.value = true
  } catch (e) {
    ElMessage.error('加载详情失败')
  }
}

// ===== 处理举报 =====
const processVisible = ref(false)
const processLoading = ref(false)
const processForm = reactive({
  id: null, target_type: '', target_id: null, target_title: '',
  status: 2, handle_remark: ''
})

const openProcess = (row, status) => {
  Object.assign(processForm, {
    id: row.id,
    target_type: row.target_type,
    target_id: row.target_id,
    target_title: row.target_title,
    status,
    handle_remark: ''
  })
  processVisible.value = true
}

const onProcessSubmit = async () => {
  try {
    processLoading.value = true
    await processMallReport(processForm.id, {
      status: processForm.status,
      handle_remark: processForm.handle_remark
    })
    ElMessage.success('处理成功')
    processVisible.value = false
    await loadList()
  } catch (e) {
    ElMessage.error('处理失败')
  } finally {
    processLoading.value = false
  }
}

const onDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确认删除举报 #${row.id}？`, '提示', { type: 'warning' })
    await deleteMallReport(row.id)
    ElMessage.success('删除成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

onMounted(() => loadList())
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-item { text-align: center; padding: 16px; }
.stat-value { font-size: 20px; font-weight: 600; color: #303133; }
.text-success { color: #67c23a; }
.text-warning { color: #e6a23c; }
.text-primary { color: #409eff; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; margin-bottom: 12px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.image-list { display: flex; gap: 8px; flex-wrap: wrap; }
.report-thumb { width: 80px; height: 80px; border-radius: 4px; }
</style>
