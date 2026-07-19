<template>
  <div class="app-container">
    <div class="page-card">
      <el-tabs v-model="activeTab">
        <el-tab-pane label="批量审核" name="audit">
          <el-form :inline="true" :model="auditFilters" class="filter-form" @submit.prevent>
            <el-form-item label="审核状态">
              <el-select v-model="auditFilters.audit_status" placeholder="全部" clearable style="width: 140px" @change="loadAuditList">
                <el-option label="待审核" :value="0" />
                <el-option label="已通过" :value="1" />
                <el-option label="已拒绝" :value="2" />
              </el-select>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :icon="Search" @click="loadAuditList">搜索</el-button>
            </el-form-item>
          </el-form>
          <el-table v-loading="auditLoading" :data="auditList" border stripe @selection-change="(r) => auditSelection = r">
            <el-table-column type="selection" width="44" />
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="title" label="职位" min-width="180" show-overflow-tooltip />
            <el-table-column label="审核状态" width="100">
              <template #default="{ row }">
                <el-tag :type="auditTagType(row.audit_status)" size="small">{{ auditText(row.audit_status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" width="160">
              <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
            </el-table-column>
          </el-table>
          <div class="batch-actions">
            <el-button type="success" :disabled="!auditSelection.length" :loading="batchLoading" @click="onBatchAudit(1)">批量通过</el-button>
            <el-button type="danger" :disabled="!auditSelection.length" :loading="batchLoading" @click="onBatchAudit(2)">批量拒绝</el-button>
          </div>
        </el-tab-pane>

        <el-tab-pane label="批量状态变更" name="status">
          <el-form :inline="true" :model="statusFilters" class="filter-form" @submit.prevent>
            <el-form-item label="职位状态">
              <el-select v-model="statusFilters.status" placeholder="全部" clearable style="width: 140px" @change="loadStatusList">
                <el-option label="草稿" :value="0" />
                <el-option label="招聘中" :value="1" />
                <el-option label="已停招" :value="2" />
                <el-option label="已下架" :value="3" />
                <el-option label="已过期" :value="4" />
              </el-select>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :icon="Search" @click="loadStatusList">搜索</el-button>
            </el-form-item>
          </el-form>
          <el-table v-loading="statusLoading" :data="statusList" border stripe @selection-change="(r) => statusSelection = r">
            <el-table-column type="selection" width="44" />
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="title" label="职位" min-width="180" show-overflow-tooltip />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="statusTagType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
              </template>
            </el-table-column>
          </el-table>
          <div class="batch-actions">
            <el-button type="primary" :disabled="!statusSelection.length" :loading="batchLoading" @click="onBatchStatus(1)">设为招聘中</el-button>
            <el-button type="warning" :disabled="!statusSelection.length" :loading="batchLoading" @click="onBatchStatus(3)">批量下架</el-button>
            <el-button :disabled="!statusSelection.length" :loading="batchLoading" @click="onBatchStatus(4)">设为过期</el-button>
          </div>
        </el-tab-pane>

        <el-tab-pane label="批量删除" name="delete">
          <el-form :inline="true" :model="deleteFilters" class="filter-form" @submit.prevent>
            <el-form-item label="关键词">
              <el-input v-model="deleteFilters.keyword" placeholder="职位名" clearable style="width: 200px" @keyup.enter="loadDeleteList" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :icon="Search" @click="loadDeleteList">搜索</el-button>
            </el-form-item>
          </el-form>
          <el-table v-loading="deleteLoading" :data="deleteList" border stripe @selection-change="(r) => deleteSelection = r">
            <el-table-column type="selection" width="44" />
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="title" label="职位" min-width="180" show-overflow-tooltip />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="statusTagType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" width="160">
              <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
            </el-table-column>
          </el-table>
          <div class="batch-actions">
            <el-button type="danger" :disabled="!deleteSelection.length" :loading="batchLoading" @click="onBatchDelete">批量删除</el-button>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { adminListJobs, auditJob, adminUpdateJobStatus, deleteJob } from '@/api/job'
import { formatTime } from '@/utils/format'

const activeTab = ref('audit')
const batchLoading = ref(false)

const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')
const statusText = (s) => ({ 0: '草稿', 1: '招聘中', 2: '已停招', 3: '已下架', 4: '已过期' }[s] || '-')
const statusTagType = (s) => ({ 0: 'info', 1: 'success', 2: 'warning', 3: 'warning', 4: 'danger' }[s] || 'info')

// 审核 Tab
const auditLoading = ref(false); const auditList = ref([]); const auditSelection = ref([])
const auditFilters = reactive({ audit_status: 0 })
const loadAuditList = async () => {
  auditLoading.value = true
  try { const res = await adminListJobs({ page: 1, page_size: 50, audit_status: auditFilters.audit_status }); const d = res.data || {}; auditList.value = d.list || [] } catch (e) { auditList.value = [] } finally { auditLoading.value = false }
}
const onBatchAudit = async (status) => {
  try {
    await ElMessageBox.confirm(`确认批量${status === 1 ? '通过' : '拒绝'} ${auditSelection.value.length} 个职位？`, '批量审核', { type: 'warning' })
    batchLoading.value = true
    await Promise.all(auditSelection.value.map((r) => auditJob(r.id, { audit_status: status })))
    ElMessage.success('批量审核完成'); await loadAuditList()
  } catch (e) { /* */ } finally { batchLoading.value = false }
}

// 状态 Tab
const statusLoading = ref(false); const statusList = ref([]); const statusSelection = ref([])
const statusFilters = reactive({ status: null })
const loadStatusList = async () => {
  statusLoading.value = true
  try { const res = await adminListJobs({ page: 1, page_size: 50, status: statusFilters.status === null ? undefined : statusFilters.status }); const d = res.data || {}; statusList.value = d.list || [] } catch (e) { statusList.value = [] } finally { statusLoading.value = false }
}
const onBatchStatus = async (status) => {
  try {
    const label = statusText(status)
    await ElMessageBox.confirm(`确认批量将 ${statusSelection.value.length} 个职位设为「${label}」？`, '批量状态变更', { type: 'warning' })
    batchLoading.value = true
    await Promise.all(statusSelection.value.map((r) => adminUpdateJobStatus(r.id, status)))
    ElMessage.success('批量操作完成'); await loadStatusList()
  } catch (e) { /* */ } finally { batchLoading.value = false }
}

// 删除 Tab
const deleteLoading = ref(false); const deleteList = ref([]); const deleteSelection = ref([])
const deleteFilters = reactive({ keyword: '' })
const loadDeleteList = async () => {
  deleteLoading.value = true
  try { const res = await adminListJobs({ page: 1, page_size: 50, keyword: deleteFilters.keyword || undefined }); const d = res.data || {}; deleteList.value = d.list || [] } catch (e) { deleteList.value = [] } finally { deleteLoading.value = false }
}
const onBatchDelete = async () => {
  try {
    await ElMessageBox.confirm(`确认批量删除 ${deleteSelection.value.length} 个职位？删除后不可恢复！`, '危险操作', { type: 'error' })
    batchLoading.value = true
    await Promise.all(deleteSelection.value.map((r) => deleteJob(r.id)))
    ElMessage.success('批量删除完成'); await loadDeleteList()
  } catch (e) { /* */ } finally { batchLoading.value = false }
}

onMounted(() => { loadAuditList(); loadStatusList(); loadDeleteList() })
</script>

<style scoped>
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.batch-actions { margin-top: 16px; display: flex; gap: 8px; }
</style>
