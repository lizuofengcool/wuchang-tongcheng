<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="公司名/行业" clearable style="width: 200px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="认证状态">
          <el-select v-model="filters.cert_status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="未认证" :value="0" />
            <el-option label="待审核" :value="1" />
            <el-option label="已认证" :value="2" />
            <el-option label="已拒绝" :value="3" />
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
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column label="公司名称" min-width="180">
          <template #default="{ row }">
            <div>{{ row.name }}</div>
            <div class="text-muted text-xs">{{ row.short_name || '' }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="industry" label="行业" width="120" />
        <el-table-column prop="scale" label="规模" width="100" />
        <el-table-column label="所在地区" width="120">
          <template #default="{ row }">{{ row.city || row.region || '-' }}</template>
        </el-table-column>
        <el-table-column label="认证状态" width="100">
          <template #default="{ row }">
            <el-tag :type="certStatusType(row.cert_status)" size="small">{{ certStatusText(row.cert_status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="职位数" width="80" prop="job_count" />
        <el-table-column label="创建时间" width="160" prop="created_at" sortable>
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.cert_status === 1" type="success" link size="small" @click="handleAudit(row, 2)">通过</el-button>
            <el-button v-if="row.cert_status === 1" type="danger" link size="small" @click="handleAudit(row, 3)">拒绝</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <el-dialog v-model="detailVisible" title="公司详情" width="700px">
      <div v-loading="detailLoading">
        <el-descriptions v-if="detail" :column="2" border>
          <el-descriptions-item label="公司名称">{{ detail.name }}</el-descriptions-item>
          <el-descriptions-item label="简称">{{ detail.short_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="行业">{{ detail.industry || '-' }}</el-descriptions-item>
          <el-descriptions-item label="规模">{{ detail.scale || '-' }}</el-descriptions-item>
          <el-descriptions-item label="所在地区">{{ detail.city || detail.region || '-' }}</el-descriptions-item>
          <el-descriptions-item label="地址">{{ detail.address || '-' }}</el-descriptions-item>
          <el-descriptions-item label="联系电话">{{ detail.contact_phone || '-' }}</el-descriptions-item>
          <el-descriptions-item label="邮箱">{{ detail.email || '-' }}</el-descriptions-item>
          <el-descriptions-item label="认证状态">
            <el-tag :type="certStatusType(detail.cert_status)" size="small">{{ certStatusText(detail.cert_status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="职位数">{{ detail.job_count || 0 }}</el-descriptions-item>
          <el-descriptions-item v-if="detail.description" label="公司简介" :span="2">{{ detail.description }}</el-descriptions-item>
          <el-descriptions-item label="创建时间" :span="2">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        </el-descriptions>
      </div>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search } from '@element-plus/icons-vue'
import { listCompanies, getCompany, auditCompany } from '@/api/job'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const filters = reactive({ keyword: '', cert_status: null })

const certStatusText = (s) => ({ 0: '未认证', 1: '待审核', 2: '已认证', 3: '已拒绝' }[s] ?? '-')
const certStatusType = (s) => ({ 0: 'info', 1: 'warning', 2: 'success', 3: 'danger' }[s] || 'info')

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => { Object.assign(filters, { keyword: '', cert_status: null }); page.value = 1; loadList() }

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value, page_size: pageSize.value,
      keyword: filters.keyword || undefined,
      cert_status: filters.cert_status === null || filters.cert_status === '' ? undefined : filters.cert_status
    }
    const res = await listCompanies(params)
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
  } catch (e) { list.value = []; total.value = 0 } finally { loading.value = false }
}

const detailVisible = ref(false)
const detailLoading = ref(false)
const detail = ref(null)
const openDetail = async (row) => {
  detailVisible.value = true; detailLoading.value = true
  try { const res = await getCompany(row.id); detail.value = res.data || row } catch (e) { detail.value = row } finally { detailLoading.value = false }
}

const handleAudit = async (row, status) => {
  try {
    const action = status === 2 ? '通过' : '拒绝'
    let note = ''
    if (status === 3) {
      const { value } = await ElMessageBox.prompt('请输入拒绝原因（可选）', '拒绝认证', { inputType: 'textarea' })
      note = value || ''
    } else {
      await ElMessageBox.confirm(`确定${action}公司 "${row.name}" 的认证吗？`, '提示', { type: 'warning' })
    }
    await auditCompany(row.id, { cert_status: status, note })
    ElMessage.success(`已${action}`); await loadList()
  } catch (e) { /* */ }
}

onMounted(() => loadList())
</script>

<style scoped>
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; margin-bottom: 12px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.text-muted { color: #999; }
.text-xs { font-size: 12px; }
</style>
