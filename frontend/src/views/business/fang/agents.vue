<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="姓名/手机/公司" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="认证状态">
          <el-select v-model="filters.audit_status" placeholder="全部" clearable style="width: 130px" @change="onSearch">
            <el-option label="待审核" :value="0" />
            <el-option label="已通过" :value="1" />
            <el-option label="已拒绝" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="在线状态">
          <el-select v-model="filters.online_status" placeholder="全部" clearable style="width: 130px" @change="onSearch">
            <el-option label="在线" :value="1" />
            <el-option label="离线" :value="0" />
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
        <el-table-column label="经纪人" min-width="180">
          <template #default="{ row }">
            <div>{{ row.name || `经纪人#${row.id}` }}</div>
            <div class="text-muted text-xs">{{ row.phone || '-' }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="company" label="公司" width="160" show-overflow-tooltip />
        <el-table-column prop="city" label="城市" width="100" />
        <el-table-column prop="deal_count" label="成交" width="80" />
        <el-table-column prop="avg_score" label="评分" width="80">
          <template #default="{ row }">{{ Number(row.avg_score || 0).toFixed(1) }}</template>
        </el-table-column>
        <el-table-column label="认证" width="100">
          <template #default="{ row }">
            <el-tag :type="auditTagType(row.audit_status)" size="small">{{ auditText(row.audit_status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="在线" width="80">
          <template #default="{ row }">
            <el-tag :type="row.online_status === 1 ? 'success' : 'info'" size="small">{{ row.online_status === 1 ? '在线' : '离线' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.audit_status === 0" type="success" link size="small" @click="handleAudit(row, 1)">通过</el-button>
            <el-button v-if="row.audit_status === 0 || row.audit_status === 1" type="danger" link size="small" @click="handleAudit(row, 2)">拒绝</el-button>
            <el-button v-if="row.online_status === 1" type="warning" link size="small" @click="handleOnline(row, 0)">下线</el-button>
            <el-button v-else type="success" link size="small" @click="handleOnline(row, 1)">上线</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <el-dialog v-model="detailVisible" title="经纪人详情" width="700px" destroy-on-close>
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="姓名">{{ detail.name }}</el-descriptions-item>
        <el-descriptions-item label="手机">{{ detail.phone }}</el-descriptions-item>
        <el-descriptions-item label="公司">{{ detail.company }}</el-descriptions-item>
        <el-descriptions-item label="城市">{{ detail.city }}</el-descriptions-item>
        <el-descriptions-item label="区域">{{ detail.district || '-' }}</el-descriptions-item>
        <el-descriptions-item label="成交数">{{ detail.deal_count || 0 }}</el-descriptions-item>
        <el-descriptions-item label="评分">{{ Number(detail.avg_score || 0).toFixed(1) }}</el-descriptions-item>
        <el-descriptions-item label="认证">{{ auditText(detail.audit_status) }}</el-descriptions-item>
        <el-descriptions-item label="在线">{{ detail.online_status === 1 ? '在线' : '离线' }}</el-descriptions-item>
        <el-descriptions-item label="简介" :span="2">{{ detail.description || '-' }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, RefreshLeft, Refresh } from '@element-plus/icons-vue'
import { adminListAgents, auditAgent, updateAgentOnlineStatus, getAgent } from '@/api/house'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filters = reactive({ keyword: '', audit_status: '', online_status: '' })

const detailVisible = ref(false)
const detail = ref(null)

const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')

const buildParams = () => {
  const p = { page: page.value, page_size: pageSize.value }
  if (filters.keyword) p.keyword = filters.keyword
  if (filters.audit_status !== '' && filters.audit_status !== null) p.audit_status = filters.audit_status
  if (filters.online_status !== '' && filters.online_status !== null) p.online_status = filters.online_status
  return p
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await adminListAgents(buildParams())
    const d = res.data || {}
    list.value = d.list || []
    total.value = d.total || 0
  } catch (e) {
    list.value = []
  } finally {
    loading.value = false
  }
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  Object.assign(filters, { keyword: '', audit_status: '', online_status: '' })
  onSearch()
}

const openDetail = async (row) => {
  try {
    const res = await getAgent(row.id)
    detail.value = res.data || row
  } catch (e) {
    detail.value = row
  }
  detailVisible.value = true
}

const handleAudit = async (row, auditStatus) => {
  try {
    if (auditStatus === 2) {
      const { value } = await ElMessageBox.prompt('请输入拒绝原因', '拒绝审核', { confirmButtonText: '确定', cancelButtonText: '取消' })
      await auditAgent(row.id, { audit_status: auditStatus, audit_reason: value || '' })
    } else {
      await ElMessageBox.confirm('确定审核通过吗？', '提示', { type: 'warning' })
      await auditAgent(row.id, { audit_status: auditStatus })
    }
    ElMessage.success('操作成功')
    loadList()
  } catch (e) { /* cancel */ }
}

const handleOnline = async (row, onlineStatus) => {
  try {
    await ElMessageBox.confirm(`确定${onlineStatus === 1 ? '上线' : '下线'}该经纪人吗？`, '提示', { type: 'warning' })
    await updateAgentOnlineStatus(row.id, { online_status: onlineStatus })
    ElMessage.success('操作成功')
    loadList()
  } catch (e) { /* cancel */ }
}

onMounted(() => loadList())
</script>

<style scoped>
.text-muted { color: #909399; }
.text-xs { font-size: 12px; }
</style>
