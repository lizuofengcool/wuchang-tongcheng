<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="公司名/统一信用码" clearable style="width: 220px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="待审核" :value="0" />
            <el-option label="已通过" :value="1" />
            <el-option label="已拒绝" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar"><el-button :icon="Refresh" @click="loadList">刷新</el-button></div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column label="公司" min-width="160">
          <template #default="{ row }">{{ row.company_name || `公司#${row.company_id}` }}</template>
        </el-table-column>
        <el-table-column prop="credit_code" label="统一信用码" width="180" />
        <el-table-column prop="legal_person" label="法人" width="100" />
        <el-table-column prop="contact_phone" label="联系电话" width="130" />
        <el-table-column label="营业执照" width="100">
          <template #default="{ row }">
            <el-link v-if="row.license_url" type="primary" :href="row.license_url" target="_blank">查看</el-link>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="提交时间" width="160" prop="created_at" sortable>
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status === 0" type="success" link size="small" @click="handleProcess(row, 1)">通过</el-button>
            <el-button v-if="row.status === 0" type="danger" link size="small" @click="handleProcess(row, 2)">拒绝</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <el-dialog v-model="detailVisible" title="认证详情" width="700px">
      <div v-loading="detailLoading">
        <el-descriptions v-if="detail" :column="2" border>
          <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
          <el-descriptions-item label="状态"><el-tag :type="statusType(detail.status)" size="small">{{ statusText(detail.status) }}</el-tag></el-descriptions-item>
          <el-descriptions-item label="公司">{{ detail.company_name || `公司#${detail.company_id}` }}</el-descriptions-item>
          <el-descriptions-item label="统一信用码">{{ detail.credit_code || '-' }}</el-descriptions-item>
          <el-descriptions-item label="法人">{{ detail.legal_person || '-' }}</el-descriptions-item>
          <el-descriptions-item label="联系电话">{{ detail.contact_phone || '-' }}</el-descriptions-item>
          <el-descriptions-item label="注册地址" :span="2">{{ detail.registered_address || '-' }}</el-descriptions-item>
          <el-descriptions-item label="营业执照">
            <el-link v-if="detail.license_url" type="primary" :href="detail.license_url" target="_blank">查看营业执照</el-link>
            <span v-else>-</span>
          </el-descriptions-item>
          <el-descriptions-item label="提交时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
          <el-descriptions-item v-if="detail.reject_reason" label="拒绝原因" :span="2">{{ detail.reject_reason }}</el-descriptions-item>
        </el-descriptions>
      </div>
      <template #footer><el-button @click="detailVisible = false">关闭</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search } from '@element-plus/icons-vue'
import { listCertifications, getCertification, processCertification } from '@/api/job'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1); const pageSize = ref(20); const total = ref(0); const list = ref([])
const filters = reactive({ keyword: '', status: null })

const statusText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] ?? '-')
const statusType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => { Object.assign(filters, { keyword: '', status: null }); page.value = 1; loadList() }

const loadList = async () => {
  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value, keyword: filters.keyword || undefined, status: filters.status === null || filters.status === '' ? undefined : filters.status }
    const res = await listCertifications(params)
    const data = res.data || {}
    list.value = data.list || []; total.value = data.total || 0
  } catch (e) { list.value = []; total.value = 0 } finally { loading.value = false }
}

const detailVisible = ref(false); const detailLoading = ref(false); const detail = ref(null)
const openDetail = async (row) => {
  detailVisible.value = true; detailLoading.value = true
  try { const res = await getCertification(row.id); detail.value = res.data || row } catch (e) { detail.value = row } finally { detailLoading.value = false }
}

const handleProcess = async (row, status) => {
  try {
    let note = ''
    if (status === 2) {
      const { value } = await ElMessageBox.prompt('请输入拒绝原因', '拒绝认证', { inputType: 'textarea' })
      note = value || ''
    } else {
      await ElMessageBox.confirm('确定通过该企业认证吗？', '提示', { type: 'warning' })
    }
    await processCertification(row.id, { status, note })
    ElMessage.success('处理成功'); await loadList()
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
</style>
