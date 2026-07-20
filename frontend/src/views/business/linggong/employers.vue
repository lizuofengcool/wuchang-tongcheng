<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">雇主总数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.pending }}</div><div class="stat-label">待审核</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.verified }}</div><div class="stat-label">已认证</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ stats.active }}</div><div class="stat-label">活跃雇主</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="公司名/联系人" clearable style="width: 200px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="filters.employer_type" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option v-for="(label, val) in typeMap" :key="val" :label="label" :value="val" />
          </el-select>
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
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
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

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" fixed="left" />
        <el-table-column label="公司/认证" min-width="200">
          <template #default="{ row }">
            <div class="title-cell">
              <div class="title-text">
                <span>{{ row.company_name || row.contact_name || `雇主#${row.id}` }}</span>
                <el-tag v-if="row.verified" type="success" size="small">认证</el-tag>
              </div>
              <div class="text-muted">{{ typeMap[row.employer_type] || row.employer_type }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="联系人" width="120">
          <template #default="{ row }">{{ row.contact_name || '-' }}</template>
        </el-table-column>
        <el-table-column label="电话" width="140">
          <template #default="{ row }">{{ maskPhone(row.contact_phone) }}</template>
        </el-table-column>
        <el-table-column prop="linggong_count" label="岗位数" width="90" />
        <el-table-column prop="task_count" label="任务数" width="90" />
        <el-table-column label="信用分" width="90">
          <template #default="{ row }">{{ row.credit_score || 0 }}</template>
        </el-table-column>
        <el-table-column label="审核" width="100">
          <template #default="{ row }">
            <el-tag :type="auditTagType(row.audit_status)" size="small">{{ auditText(row.audit_status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-switch :model-value="row.status === 1" @change="(val) => onToggle(row, val)" />
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.audit_status === 0" type="success" link size="small" @click="onAudit(row, 1)">通过</el-button>
            <el-button v-if="row.audit_status === 0 || row.audit_status === 1" type="danger" link size="small" @click="onAudit(row, 2)">拒绝</el-button>
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
    <el-dialog v-model="detailVisible" title="雇主详情" width="700px">
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="用户ID">{{ detail.user_id }}</el-descriptions-item>
        <el-descriptions-item label="公司名称" :span="2">{{ detail.company_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="雇主类型">{{ typeMap[detail.employer_type] || detail.employer_type }}</el-descriptions-item>
        <el-descriptions-item label="联系人">{{ detail.contact_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="联系电话">{{ maskPhone(detail.contact_phone) }}</el-descriptions-item>
        <el-descriptions-item label="邮箱">{{ detail.email || '-' }}</el-descriptions-item>
        <el-descriptions-item label="统一信用代码">{{ detail.credit_code || '-' }}</el-descriptions-item>
        <el-descriptions-item label="营业执照">{{ detail.business_license || '-' }}</el-descriptions-item>
        <el-descriptions-item label="地址" :span="2">{{ detail.address || '-' }}</el-descriptions-item>
        <el-descriptions-item label="岗位数">{{ detail.linggong_count || 0 }}</el-descriptions-item>
        <el-descriptions-item label="任务数">{{ detail.task_count || 0 }}</el-descriptions-item>
        <el-descriptions-item label="信用分">{{ detail.credit_score || 0 }}</el-descriptions-item>
        <el-descriptions-item label="完成订单">{{ detail.completed_count || 0 }}</el-descriptions-item>
        <el-descriptions-item label="审核状态">
          <el-tag :type="auditTagType(detail.audit_status)" size="small">{{ auditText(detail.audit_status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="状态">{{ detail.status === 1 ? '启用' : '禁用' }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.audit_reason" label="审核原因" :span="2">{{ detail.audit_reason }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.description" label="简介" :span="2">{{ detail.description }}</el-descriptions-item>
        <el-descriptions-item label="认证时间">{{ formatTime(detail.verified_at) }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
      </el-descriptions>
      <template #footer><el-button @click="detailVisible = false">关闭</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search } from '@element-plus/icons-vue'
import {
  adminListLinggongEmployers, getLinggongEmployer, auditLinggongEmployer,
  adminUpdateLinggongEmployerStatus
} from '@/api/linggong'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const filters = reactive({ keyword: '', employer_type: '', audit_status: null, status: null })

const typeMap = {
  personal: '个人雇主', company: '企业雇主', agent: '中介', headhunter: '猎头'
}
const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')

const stats = computed(() => {
  const total = list.value.length
  const pending = list.value.filter((r) => r.audit_status === 0).length
  const verified = list.value.filter((r) => r.verified || r.audit_status === 1).length
  const active = list.value.filter((r) => r.status === 1).length
  return { total, pending, verified, active }
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
  filters.employer_type = ''
  filters.audit_status = null
  filters.status = null
  page.value = 1
  loadList()
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await adminListLinggongEmployers({
      page: page.value,
      page_size: pageSize.value,
      keyword: filters.keyword || undefined,
      employer_type: filters.employer_type || undefined,
      audit_status: filters.audit_status === null || filters.audit_status === '' ? undefined : filters.audit_status,
      status: filters.status === null || filters.status === '' ? undefined : filters.status
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

const openDetail = async (row) => {
  try {
    const res = await getLinggongEmployer(row.id)
    detail.value = res.data || row
  } catch (e) {
    detail.value = row
  }
  detailVisible.value = true
}

const onAudit = async (row, auditStatus) => {
  try {
    if (auditStatus === 2) {
      const { value } = await ElMessageBox.prompt('请输入拒绝原因', '拒绝认证', {
        inputType: 'textarea',
        inputPlaceholder: '拒绝原因'
      })
      await auditLinggongEmployer(row.id, { audit_status: auditStatus, audit_reason: value || '' })
    } else {
      await ElMessageBox.confirm('确定通过该雇主认证吗？', '提示', { type: 'warning' })
      await auditLinggongEmployer(row.id, { audit_status: auditStatus })
    }
    ElMessage.success('操作成功')
    await loadList()
  } catch (e) {
    // 取消
  }
}

const onToggle = async (row, val) => {
  try {
    await adminUpdateLinggongEmployerStatus(row.id, val ? 1 : 0)
    row.status = val ? 1 : 0
    ElMessage.success('状态已更新')
  } catch (e) {
    // ignore
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
.title-cell { display: flex; flex-direction: column; gap: 2px; }
.title-text { display: flex; align-items: center; gap: 6px; font-weight: 500; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
