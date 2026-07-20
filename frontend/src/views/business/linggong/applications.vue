<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">总报名数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.pending }}</div><div class="stat-label">待审核</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.approved }}</div><div class="stat-label">已通过</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ stats.confirmed }}</div><div class="stat-label">已确认</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="报名编号/姓名" clearable style="width: 200px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="审核状态">
          <el-select v-model="filters.audit_status" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option label="待审核" :value="0" />
            <el-option label="已通过" :value="1" />
            <el-option label="已拒绝" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="报名状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option v-for="(label, val) in statusMap" :key="val" :label="label" :value="Number(val)" />
          </el-select>
        </el-form-item>
        <el-form-item label="岗位ID">
          <el-input v-model="filters.linggong_id" placeholder="岗位ID" clearable style="width: 140px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar">
        <el-button :icon="Refresh" @click="loadList">刷新</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" fixed="left" />
        <el-table-column prop="application_no" label="报名编号" width="180" />
        <el-table-column label="岗位" min-width="180">
          <template #default="{ row }">
            <div>{{ row.linggong_title || `岗位#${row.linggong_id}` }}</div>
            <div class="text-muted">¥{{ Number(row.salary || 0).toFixed(2) }}</div>
          </template>
        </el-table-column>
        <el-table-column label="报名人" width="140">
          <template #default="{ row }">
            <div>{{ row.worker_name || `用户#${row.worker_id}` }}</div>
            <div class="text-muted">{{ maskPhone(row.worker_phone) }}</div>
          </template>
        </el-table-column>
        <el-table-column label="期望薪资" width="120">
          <template #default="{ row }">¥{{ Number(row.expected_salary || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="可开工" width="120">
          <template #default="{ row }">{{ formatTime(row.available_date, 'YYYY-MM-DD') }}</template>
        </el-table-column>
        <el-table-column label="审核" width="100">
          <template #default="{ row }">
            <el-tag :type="auditTagType(row.audit_status)" size="small">{{ auditText(row.audit_status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusMap[row.status] || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="报名时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.audit_status === 0" type="success" link size="small" @click="onAudit(row, 1)">通过</el-button>
            <el-button v-if="row.audit_status === 0 || row.audit_status === 1" type="danger" link size="small" @click="onAudit(row, 2)">拒绝</el-button>
            <el-button v-if="row.audit_status === 1" type="warning" link size="small" @click="onCancel(row)">取消</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next, jumper"
          background
          @current-change="loadList"
          @size-change="loadList"
        />
      </div>
    </div>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="报名详情" width="700px">
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="报名编号">{{ detail.application_no }}</el-descriptions-item>
        <el-descriptions-item label="岗位" :span="2">{{ detail.linggong_title || `岗位#${detail.linggong_id}` }}</el-descriptions-item>
        <el-descriptions-item label="报名人">{{ detail.worker_name || `用户#${detail.worker_id}` }}</el-descriptions-item>
        <el-descriptions-item label="联系电话">{{ maskPhone(detail.worker_phone) }}</el-descriptions-item>
        <el-descriptions-item label="期望薪资">¥{{ Number(detail.expected_salary || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="可开工日期">{{ formatTime(detail.available_date, 'YYYY-MM-DD') }}</el-descriptions-item>
        <el-descriptions-item label="审核状态">
          <el-tag :type="auditTagType(detail.audit_status)" size="small">{{ auditText(detail.audit_status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="报名状态">
          <el-tag :type="statusTagType(detail.status)" size="small">{{ statusMap[detail.status] || '-' }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item v-if="detail.audit_reason" label="审核原因" :span="2">{{ detail.audit_reason }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.remark" label="报名备注" :span="2">{{ detail.remark }}</el-descriptions-item>
        <el-descriptions-item label="报名时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
      </el-descriptions>
      <template #footer><el-button @click="detailVisible = false">关闭</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search } from '@element-plus/icons-vue'
import { adminListLinggongApplications, auditLinggongApplication, cancelLinggongApplication } from '@/api/linggong'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const filters = reactive({ keyword: '', audit_status: null, status: null, linggong_id: '' })

const statusMap = {
  0: '待处理', 1: '已通过', 2: '已拒绝', 3: '已确认',
  4: '已取消', 5: '已过期', 6: '已分配', 7: '已完成'
}
const statusTagType = (s) => ({
  0: 'info', 1: 'success', 2: 'danger', 3: 'primary',
  4: 'warning', 5: 'info', 6: 'success', 7: 'success'
}[s] || 'info')

const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')

const stats = computed(() => {
  const total = list.value.length
  const pending = list.value.filter((r) => r.audit_status === 0).length
  const approved = list.value.filter((r) => r.audit_status === 1).length
  const confirmed = list.value.filter((r) => r.status === 3).length
  return { total, pending, approved, confirmed }
})

const maskPhone = (p) => {
  if (!p) return '-'
  const s = String(p)
  if (s.length < 7) return s
  return s.slice(0, 3) + '****' + s.slice(-4)
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.keyword = ''
  filters.audit_status = null
  filters.status = null
  filters.linggong_id = ''
  page.value = 1
  loadList()
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await adminListLinggongApplications({
      page: page.value,
      page_size: pageSize.value,
      keyword: filters.keyword || undefined,
      audit_status: filters.audit_status === null || filters.audit_status === '' ? undefined : filters.audit_status,
      status: filters.status === null || filters.status === '' ? undefined : filters.status,
      linggong_id: filters.linggong_id || undefined
    })
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
  } catch (e) {
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const detailVisible = ref(false)
const detail = ref(null)
const openDetail = (row) => {
  detail.value = row
  detailVisible.value = true
}

const onAudit = async (row, auditStatus) => {
  try {
    if (auditStatus === 2) {
      const { value } = await ElMessageBox.prompt('请输入拒绝原因（可选）', '拒绝报名', {
        inputType: 'textarea',
        inputPlaceholder: '拒绝原因（可不填）'
      })
      await auditLinggongApplication(row.id, { audit_status: auditStatus, audit_reason: value || '' })
    } else {
      await ElMessageBox.confirm('确定通过该报名吗？', '提示', { type: 'warning' })
      await auditLinggongApplication(row.id, { audit_status: auditStatus })
    }
    ElMessage.success('操作成功')
    await loadList()
  } catch (e) {
    // 取消
  }
}

const onCancel = async (row) => {
  try {
    await ElMessageBox.confirm('确定取消该报名吗？', '提示', { type: 'warning' })
    await cancelLinggongApplication(row.id)
    ElMessage.success('已取消')
    await loadList()
  } catch (e) {
    // 取消
  }
}

onMounted(() => loadList())
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-item { text-align: center; padding: 16px; }
.stat-value { font-size: 22px; font-weight: 600; color: #303133; }
.text-primary { color: #409eff; }
.text-warning { color: #e6a23c; }
.text-success { color: #67c23a; }
.text-muted { color: #909399; font-size: 12px; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: flex-start; margin-bottom: 12px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
