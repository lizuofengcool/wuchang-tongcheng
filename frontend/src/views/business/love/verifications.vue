<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">总认证数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.pending }}</div><div class="stat-label">待审核</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.approved }}</div><div class="stat-label">已通过</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-danger">{{ stats.rejected }}</div><div class="stat-label">已拒绝</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ stats.today }}</div><div class="stat-label">今日提交</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.verifyRate }}%</div><div class="stat-label">认证率</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="会员ID/昵称/姓名" clearable style="width: 220px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="认证类型">
          <el-select v-model="filters.verify_type" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option label="身份证" value="id_card" />
            <el-option label="人脸识别" value="face" />
            <el-option label="视频认证" value="video" />
            <el-option label="学历认证" value="education" />
            <el-option label="房产认证" value="house" />
            <el-option label="车辆认证" value="car" />
          </el-select>
        </el-form-item>
        <el-form-item label="审核状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="待审核" value="pending" />
            <el-option label="已通过" value="approved" />
            <el-option label="已拒绝" value="rejected" />
          </el-select>
        </el-form-item>
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="filters.dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            style="width: 240px"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar"><el-button :icon="Refresh" @click="loadList">刷新</el-button></div>

      <el-table v-loading="loading" :data="list" border stripe @sort-change="onSortChange">
        <el-table-column prop="id" label="ID" width="70" sortable="custom" />
        <el-table-column prop="love_id" label="会员ID" width="90" />
        <el-table-column label="会员" min-width="140">
          <template #default="{ row }">
            <el-link type="primary" :underline="'never'" @click="openDetail(row)">{{ row.nickname || `#${row.love_id}` }}</el-link>
          </template>
        </el-table-column>
        <el-table-column label="真实姓名" width="120">
          <template #default="{ row }">{{ row.real_name || maskName(row.real_name) || '-' }}</template>
        </el-table-column>
        <el-table-column label="身份证号" width="180">
          <template #default="{ row }">{{ maskIdCard(row.id_card) }}</template>
        </el-table-column>
        <el-table-column label="类型" width="120">
          <template #default="{ row }">
            <el-tag :type="typeTagType(row.verify_type)" size="small">{{ typeText(row.verify_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="证件" width="80">
          <template #default="{ row }">
            <el-badge v-if="row.id_card_image" type="primary" value="图">
              <el-icon><Picture /></el-icon>
            </el-badge>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="审核人" width="120">
          <template #default="{ row }">{{ row.auditor_name || `#${row.auditor_id}` || '-' }}</template>
        </el-table-column>
        <el-table-column label="提交时间" width="160" prop="created_at" sortable="custom">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <template v-if="row.status === 'pending'">
              <el-button type="success" link size="small" @click="handleAudit(row, 'approved')">通过</el-button>
              <el-button type="danger" link size="small" @click="handleAudit(row, 'rejected')">拒绝</el-button>
            </template>
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
    <el-dialog v-model="detailVisible" title="认证详情" width="700px">
      <div v-loading="detailLoading">
        <el-descriptions v-if="detail" :column="2" border>
          <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
          <el-descriptions-item label="会员ID">{{ detail.love_id }}</el-descriptions-item>
          <el-descriptions-item label="真实姓名">{{ detail.real_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="身份证号">{{ maskIdCard(detail.id_card) }}</el-descriptions-item>
          <el-descriptions-item label="认证类型">
            <el-tag :type="typeTagType(detail.verify_type)" size="small">{{ typeText(detail.verify_type) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusTagType(detail.status)" size="small">{{ statusText(detail.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="审核人">{{ detail.auditor_name || `#${detail.auditor_id}` || '-' }}</el-descriptions-item>
          <el-descriptions-item label="审核时间">{{ formatTime(detail.audited_at) }}</el-descriptions-item>
          <el-descriptions-item v-if="detail.audit_reason" label="拒绝原因" :span="2">{{ detail.audit_reason }}</el-descriptions-item>
          <el-descriptions-item label="提交时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="更新时间">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
          <el-descriptions-item v-if="detail.id_card_image" label="证件照" :span="2">
            <el-image :src="detail.id_card_image" fit="cover" class="image-item" :preview-src-list="[detail.id_card_image]" preview-teleported />
          </el-descriptions-item>
          <el-descriptions-item v-if="detail.face_image" label="人脸照" :span="2">
            <el-image :src="detail.face_image" fit="cover" class="image-item" :preview-src-list="[detail.face_image]" preview-teleported />
          </el-descriptions-item>
        </el-descriptions>
      </div>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
        <template v-if="detail && detail.status === 'pending'">
          <el-button type="success" :loading="detailLoading" @click="handleAudit(detail, 'approved')">通过</el-button>
          <el-button type="danger" :loading="detailLoading" @click="handleAudit(detail, 'rejected')">拒绝</el-button>
        </template>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search, Picture } from '@element-plus/icons-vue'
import request from '@/utils/request'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const sortField = ref('created_at')
const sortOrder = ref('descending')

const filters = reactive({ keyword: '', verify_type: '', status: '', dateRange: null })

const stats = reactive({ total: 0, pending: 0, approved: 0, rejected: 0, today: 0, verifyRate: 0 })

const typeText = (t) => ({
  id_card: '身份证', face: '人脸识别', video: '视频认证',
  education: '学历认证', house: '房产认证', car: '车辆认证'
}[t] || t || '-')
const typeTagType = (t) => ({
  id_card: 'primary', face: 'success', video: 'warning',
  education: 'info', house: 'danger', car: 'info'
}[t] || 'info')
const statusText = (s) => ({ pending: '待审核', approved: '已通过', rejected: '已拒绝' }[s] || '-')
const statusTagType = (s) => ({ pending: 'warning', approved: 'success', rejected: 'danger' }[s] || 'info')

const maskIdCard = (id) => {
  if (!id) return '-'
  const s = String(id)
  if (s.length < 8) return s
  return s.slice(0, 4) + '********' + s.slice(-4)
}
const maskName = (name) => {
  if (!name) return ''
  if (name.length <= 1) return name
  return name[0] + '*'.repeat(name.length - 2) + name[name.length - 1]
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.keyword = ''; filters.verify_type = ''; filters.status = ''; filters.dateRange = null
  page.value = 1; loadList()
}
const onSortChange = ({ prop, order }) => {
  sortField.value = prop || 'created_at'
  sortOrder.value = order || 'descending'
  loadList()
}

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      keyword: filters.keyword.trim() || undefined,
      verify_type: filters.verify_type || undefined,
      status: filters.status || undefined,
      sort: sortField.value,
      order: sortOrder.value
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await request.get('/love/admin/verifications', { params })
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
    calcStats()
  } catch (e) {
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const calcStats = () => {
  const today = new Date().toISOString().slice(0, 10)
  stats.total = list.value.length
  stats.pending = list.value.filter((r) => r.status === 'pending').length
  stats.approved = list.value.filter((r) => r.status === 'approved').length
  stats.rejected = list.value.filter((r) => r.status === 'rejected').length
  stats.today = list.value.filter((r) => (r.created_at || '').startsWith(today)).length
  stats.verifyRate = stats.total ? Math.round((stats.approved / stats.total) * 100) : 0
}

const detailVisible = ref(false)
const detailLoading = ref(false)
const detail = ref(null)

const openDetail = async (row) => {
  detail.value = row
  detailVisible.value = true
  detailLoading.value = true
  try {
    const res = await request.get(`/love/verifications/${row.id}`)
    if (res.data) detail.value = res.data
  } catch (e) { /* keep */ }
  finally { detailLoading.value = false }
}

const handleAudit = async (row, status) => {
  try {
    let audit_reason = ''
    if (status === 'rejected') {
      const { value } = await ElMessageBox.prompt('请输入拒绝原因', '拒绝认证', {
        inputType: 'textarea',
        inputPlaceholder: '请输入拒绝原因'
      })
      audit_reason = value || ''
    } else {
      await ElMessageBox.confirm(`确认通过会员 #${row.love_id} 的${typeText(row.verify_type)}认证？`, '提示', { type: 'success' })
    }
    await request.put(`/love/admin/verifications/${row.id}/audit`, { status, audit_reason })
    ElMessage.success('操作成功')
    if (detailVisible.value) detailVisible.value = false
    await loadList()
  } catch (e) { /* cancel */ }
}

onMounted(() => loadList())
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-item { text-align: center; padding: 16px; }
.stat-value { font-size: 22px; font-weight: 600; color: #303133; }
.text-success { color: #67c23a; }
.text-warning { color: #e6a23c; }
.text-primary { color: #409eff; }
.text-danger { color: #f56c6c; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: flex-start; margin-bottom: 12px; }
.text-muted { color: #909399; }
.image-item { width: 160px; height: 160px; border-radius: 4px; border: 1px solid #ebeef5; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
