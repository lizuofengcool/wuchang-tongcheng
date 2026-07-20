<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #409eff">
            <el-icon :size="22"><Document /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.total }}</div>
            <div class="stat-label">资质总数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #e6a23c">
            <el-icon :size="22"><Clock /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.pending }}</div>
            <div class="stat-label">待审核</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #67c23a">
            <el-icon :size="22"><CircleCheck /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.approved }}</div>
            <div class="stat-label">已通过</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #f56c6c">
            <el-icon :size="22"><Warning /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.expired }}</div>
            <div class="stat-label">已过期</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="证书名/用户" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="filters.cert_type" placeholder="全部" clearable style="width: 160px" @change="onSearch">
            <el-option v-for="(label, val) in typeMap" :key="val" :label="label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item label="审核状态">
          <el-select v-model="filters.audit_status" placeholder="全部" clearable style="width: 130px" @change="onSearch">
            <el-option label="待审核" :value="0" />
            <el-option label="已通过" :value="1" />
            <el-option label="已拒绝" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="用户ID">
          <el-input v-model="filters.user_id" placeholder="用户ID" clearable style="width: 120px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar">
        <div class="toolbar-left">
          <el-button :icon="Refresh" @click="loadList">刷新</el-button>
        </div>
        <div class="toolbar-right">
          <el-button type="primary" :icon="Plus" @click="openCreate">新建资质</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" fixed="left" />
        <el-table-column label="证书编号" width="170">
          <template #default="{ row }">
            <span class="text-primary">{{ row.cert_no || `#${row.id}` }}</span>
          </template>
        </el-table-column>
        <el-table-column label="用户" width="140">
          <template #default="{ row }">
            <div>{{ row.user_name || `#${row.user_id}` }}</div>
            <div class="text-muted text-xs">{{ maskPhone(row.user_phone) }}</div>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="120">
          <template #default="{ row }">
            <el-tag size="small" :type="typeTagType(row.cert_type)">{{ typeMap[row.cert_type] || row.cert_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="cert_name" label="证书名" min-width="140" show-overflow-tooltip />
        <el-table-column prop="issuer" label="发证机构" min-width="130" show-overflow-tooltip />
        <el-table-column label="有效期" width="200">
          <template #default="{ row }">
            <div>{{ formatDate(row.start_date) }} 至</div>
            <div :class="{ 'text-danger': isExpiringSoon(row.end_date) }">{{ formatDate(row.end_date) }}</div>
          </template>
        </el-table-column>
        <el-table-column label="证书图" width="80">
          <template #default="{ row }">
            <el-image v-if="row.cert_image" :src="row.cert_image" fit="cover" class="cover-thumb" :preview-src-list="[row.cert_image]" preview-teleported />
            <div v-else class="cover-thumb cover-empty">无</div>
          </template>
        </el-table-column>
        <el-table-column label="审核" width="100">
          <template #default="{ row }">
            <el-tag :type="auditTagType(row.audit_status)" size="small">{{ auditText(row.audit_status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="validTagType(row)" size="small" effect="plain">{{ validText(row) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="提交时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.audit_status === 0 || row.audit_status === 2" type="success" link size="small" @click="onVerify(row, 1)">通过</el-button>
            <el-button v-if="row.audit_status === 0 || row.audit_status === 1" type="danger" link size="small" @click="onVerify(row, 2)">拒绝</el-button>
            <el-button type="danger" link size="small" @click="onDelete(row)">删除</el-button>
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
    <el-dialog v-model="detailVisible" title="资质详情" width="700px">
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="证书编号">{{ detail.cert_no || `#${detail.id}` }}</el-descriptions-item>
        <el-descriptions-item label="用户">{{ detail.user_name || `#${detail.user_id}` }}</el-descriptions-item>
        <el-descriptions-item label="类型">
          <el-tag size="small" :type="typeTagType(detail.cert_type)">{{ typeMap[detail.cert_type] || detail.cert_type }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="证书名">{{ detail.cert_name }}</el-descriptions-item>
        <el-descriptions-item label="发证机构">{{ detail.issuer || '-' }}</el-descriptions-item>
        <el-descriptions-item label="证书编号">{{ detail.cert_number || '-' }}</el-descriptions-item>
        <el-descriptions-item label="开始日期">{{ formatDate(detail.start_date) }}</el-descriptions-item>
        <el-descriptions-item label="结束日期">{{ formatDate(detail.end_date) }}</el-descriptions-item>
        <el-descriptions-item label="审核状态">
          <el-tag :type="auditTagType(detail.audit_status)" size="small">{{ auditText(detail.audit_status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="审核原因">{{ detail.audit_reason || '-' }}</el-descriptions-item>
        <el-descriptions-item label="审核人">{{ detail.auditor_name || `#${detail.auditor_id}` }}</el-descriptions-item>
        <el-descriptions-item label="审核时间">{{ formatTime(detail.audited_at) }}</el-descriptions-item>
        <el-descriptions-item label="提交时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
        <el-descriptions-item label="证书图片" :span="2">
          <el-image v-if="detail.cert_image" :src="detail.cert_image" fit="cover" class="cert-image" :preview-src-list="[detail.cert_image]" preview-teleported />
          <span v-else class="text-muted">无</span>
        </el-descriptions-item>
        <el-descriptions-item v-if="detail.description" label="描述" :span="2">{{ detail.description }}</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 拒绝原因弹窗 -->
    <el-dialog v-model="rejectVisible" title="拒绝原因" width="500px">
      <el-form :model="rejectForm" label-width="100px">
        <el-form-item label="拒绝原因">
          <el-input v-model="rejectForm.audit_reason" type="textarea" :rows="4" placeholder="请输入拒绝原因" maxlength="500" show-word-limit />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rejectVisible = false">取消</el-button>
        <el-button type="danger" :loading="verifyLoading" @click="confirmReject">确认拒绝</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Document, Clock, CircleCheck, Warning, Refresh, RefreshLeft, Search, Plus } from '@element-plus/icons-vue'
import {
  listLinggongCertifications, getLinggongCertification, verifyLinggongCertification, deleteLinggongCertification, createLinggongCertification
} from '@/api/linggong'
import { formatTime, formatDate } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const filters = reactive({ keyword: '', cert_type: '', audit_status: null, user_id: '' })

const typeMap = {
  id_card: '身份证', health_cert: '健康证', skill_cert: '技能证',
  education: '学历证', professional: '职业资格', work_experience: '工作经历',
  other: '其他'
}
const typeTagType = (t) => ({
  id_card: 'primary', health_cert: 'success', skill_cert: 'warning',
  education: 'info', professional: 'danger', work_experience: '', other: 'info'
}[t] || 'info')

const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')

const isExpired = (end) => end && new Date(end).getTime() < Date.now()
const isExpiringSoon = (end) => {
  if (!end) return false
  const diff = new Date(end).getTime() - Date.now()
  return diff >= 0 && diff < 30 * 24 * 3600 * 1000
}
const validText = (row) => {
  if (row.audit_status !== 1) return '未生效'
  if (isExpired(row.end_date)) return '已过期'
  if (isExpiringSoon(row.end_date)) return '即将过期'
  return '有效'
}
const validTagType = (row) => {
  if (row.audit_status !== 1) return 'info'
  if (isExpired(row.end_date)) return 'danger'
  if (isExpiringSoon(row.end_date)) return 'warning'
  return 'success'
}

const maskPhone = (p) => {
  if (!p) return '-'
  const s = String(p)
  return s.length >= 11 ? s.slice(0, 3) + '****' + s.slice(-4) : s
}

const stats = computed(() => {
  const total = list.value.length
  const pending = list.value.filter((r) => r.audit_status === 0).length
  const approved = list.value.filter((r) => r.audit_status === 1 && !isExpired(r.end_date)).length
  const expired = list.value.filter((r) => isExpired(r.end_date)).length
  return { total, pending, approved, expired }
})

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.keyword = ''
  filters.cert_type = ''
  filters.audit_status = null
  filters.user_id = ''
  page.value = 1
  loadList()
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await listLinggongCertifications({
      page: page.value,
      page_size: pageSize.value,
      keyword: filters.keyword || undefined,
      cert_type: filters.cert_type || undefined,
      audit_status: filters.audit_status === null || filters.audit_status === '' ? undefined : filters.audit_status,
      user_id: filters.user_id || undefined
    })
    const data = res.data || {}
    list.value = data.list || data || []
    total.value = data.total || list.value.length
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
    const res = await getLinggongCertification(row.id)
    detail.value = res.data || row
    detailVisible.value = true
  } catch (e) {
    detail.value = row
    detailVisible.value = true
  }
}

const verifyLoading = ref(false)
const onVerify = async (row, status) => {
  if (status === 2) {
    rejectForm.id = row.id
    rejectForm.audit_reason = ''
    rejectVisible.value = true
    return
  }
  try {
    await ElMessageBox.confirm('确认通过该资质审核？', '提示', { type: 'warning' })
    verifyLoading.value = true
    await verifyLinggongCertification(row.id, { audit_status: 1 })
    ElMessage.success('已通过')
    await loadList()
  } catch (e) {
    // 取消或失败
  } finally {
    verifyLoading.value = false
  }
}

const rejectVisible = ref(false)
const rejectForm = reactive({ id: null, audit_reason: '' })
const confirmReject = async () => {
  verifyLoading.value = true
  try {
    await verifyLinggongCertification(rejectForm.id, {
      audit_status: 2,
      audit_reason: rejectForm.audit_reason
    })
    ElMessage.success('已拒绝')
    rejectVisible.value = false
    await loadList()
  } catch (e) {
    // 失败已提示
  } finally {
    verifyLoading.value = false
  }
}

const onDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确认删除资质「${row.cert_name || row.cert_no || row.id}」？`, '提示', { type: 'warning' })
    await deleteLinggongCertification(row.id)
    ElMessage.success('删除成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

// 新建资质弹窗（仅占位，主要资质由 C 端提交）
const openCreate = () => {
  ElMessage.info('资质证书通常由工人在 C 端提交，管理端仅审核与维护')
}

onMounted(() => loadList())
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-card { display: flex; align-items: center; }
.stat-card :deep(.el-card__body) { display: flex; align-items: center; width: 100%; padding: 16px; }
.stat-icon { width: 48px; height: 48px; border-radius: 8px; color: #fff; display: flex; align-items: center; justify-content: center; margin-right: 12px; flex-shrink: 0; }
.stat-content { flex: 1; }
.stat-value { font-size: 22px; font-weight: 600; color: #303133; }
.stat-label { font-size: 12px; color: #909399; margin-top: 2px; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; margin-bottom: 12px; }
.toolbar-left, .toolbar-right { display: flex; gap: 8px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.cover-thumb { width: 50px; height: 50px; border-radius: 4px; border: 1px solid #ebeef5; }
.cover-empty { display: flex; align-items: center; justify-content: center; color: #c0c4cc; font-size: 12px; background: #f5f7fa; }
.text-muted { color: #909399; }
.text-xs { font-size: 12px; }
.text-primary { color: #409eff; }
.text-danger { color: #f56c6c; }
.cert-image { width: 200px; height: 140px; border-radius: 4px; border: 1px solid #ebeef5; }
</style>
