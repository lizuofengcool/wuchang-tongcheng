<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">工人总数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.pending }}</div><div class="stat-label">待审核</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.verified }}</div><div class="stat-label">已认证</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ stats.active }}</div><div class="stat-label">活跃工人</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="姓名/手机号" clearable style="width: 200px" @keyup.enter="onSearch" />
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
        <el-table-column label="头像" width="70">
          <template #default="{ row }">
            <el-avatar :src="row.avatar" :size="40" v-if="row.avatar" />
            <div v-else class="cover-thumb cover-empty">{{ (row.name || '?').charAt(0) }}</div>
          </template>
        </el-table-column>
        <el-table-column label="姓名/标签" min-width="180">
          <template #default="{ row }">
            <div class="title-cell">
              <div class="title-text">
                <span>{{ row.name || `用户#${row.user_id}` }}</span>
                <el-tag v-if="row.verified" type="success" size="small">认证</el-tag>
                <el-tag v-if="row.gender === 'male'" type="primary" size="small">男</el-tag>
                <el-tag v-else-if="row.gender === 'female'" type="danger" size="small">女</el-tag>
              </div>
              <div class="text-muted">{{ maskPhone(row.phone) }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="年龄" width="70">
          <template #default="{ row }">{{ row.age || '-' }}</template>
        </el-table-column>
        <el-table-column label="信用分" width="90">
          <template #default="{ row }">
            <el-tag :type="creditTagType(row.credit_score)" size="small">{{ row.credit_score || 0 }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="completed_count" label="已完成" width="80" />
        <el-table-column prop="application_count" label="报名数" width="80" />
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
    <el-dialog v-model="detailVisible" title="工人详情" width="700px">
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="用户ID">{{ detail.user_id }}</el-descriptions-item>
        <el-descriptions-item label="姓名">{{ detail.name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="性别">{{ genderText(detail.gender) }}</el-descriptions-item>
        <el-descriptions-item label="年龄">{{ detail.age || '-' }}</el-descriptions-item>
        <el-descriptions-item label="手机号">{{ maskPhone(detail.phone) }}</el-descriptions-item>
        <el-descriptions-item label="学历">{{ detail.education || '-' }}</el-descriptions-item>
        <el-descriptions-item label="经验">{{ detail.experience || '-' }}</el-descriptions-item>
        <el-descriptions-item label="所在地" :span="2">{{ detail.address || '-' }}</el-descriptions-item>
        <el-descriptions-item label="信用分">{{ detail.credit_score || 0 }}</el-descriptions-item>
        <el-descriptions-item label="已完成数">{{ detail.completed_count || 0 }}</el-descriptions-item>
        <el-descriptions-item label="报名数">{{ detail.application_count || 0 }}</el-descriptions-item>
        <el-descriptions-item label="认证数">{{ detail.certification_count || 0 }}</el-descriptions-item>
        <el-descriptions-item label="审核状态">
          <el-tag :type="auditTagType(detail.audit_status)" size="small">{{ auditText(detail.audit_status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="状态">{{ detail.status === 1 ? '启用' : '禁用' }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.skills" label="技能" :span="2">{{ Array.isArray(detail.skills) ? detail.skills.join('、') : detail.skills }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.bio" label="简介" :span="2">{{ detail.bio }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.audit_reason" label="审核原因" :span="2">{{ detail.audit_reason }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
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
import {
  adminListLinggongWorkers, getLinggongWorker, auditLinggongWorker,
  adminUpdateLinggongWorkerStatus
} from '@/api/linggong'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const filters = reactive({ keyword: '', audit_status: null, status: null })

const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')
const genderText = (g) => ({ male: '男', female: '女', unknown: '未知' }[g] || '未知')
const creditTagType = (s) => {
  if (!s || s >= 80) return 'success'
  if (s >= 60) return 'primary'
  if (s >= 40) return 'warning'
  return 'danger'
}

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
  filters.audit_status = null
  filters.status = null
  page.value = 1
  loadList()
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await adminListLinggongWorkers({
      page: page.value,
      page_size: pageSize.value,
      keyword: filters.keyword || undefined,
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
    const res = await getLinggongWorker(row.id)
    detail.value = res.data || row
  } catch (e) {
    detail.value = row
  }
  detailVisible.value = true
}

const onAudit = async (row, auditStatus) => {
  try {
    if (auditStatus === 2) {
      const { value } = await ElMessageBox.prompt('请输入拒绝原因', '拒绝审核', {
        inputType: 'textarea',
        inputPlaceholder: '拒绝原因'
      })
      await auditLinggongWorker(row.id, { audit_status: auditStatus, audit_reason: value || '' })
    } else {
      await ElMessageBox.confirm('确定通过该工人审核吗？', '提示', { type: 'warning' })
      await auditLinggongWorker(row.id, { audit_status: auditStatus })
    }
    ElMessage.success('操作成功')
    await loadList()
  } catch (e) {
    // 取消
  }
}

const onToggle = async (row, val) => {
  try {
    await adminUpdateLinggongWorkerStatus(row.id, val ? 1 : 0)
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
.cover-thumb {
  width: 40px; height: 40px; border-radius: 50%; border: 1px solid #ebeef5;
}
.cover-empty {
  display: flex; align-items: center; justify-content: center;
  background: #fafafa; color: #909399; font-size: 14px;
}
.title-cell { display: flex; flex-direction: column; gap: 2px; }
.title-text { display: flex; align-items: center; gap: 6px; font-weight: 500; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
