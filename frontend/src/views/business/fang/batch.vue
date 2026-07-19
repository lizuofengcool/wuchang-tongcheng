<template>
  <div class="app-container">
    <el-tabs v-model="activeTab" type="card">
      <el-tab-pane label="批量审核" name="audit">
        <div class="page-card">
          <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
            <el-form-item label="关键词">
              <el-input v-model="filters.keyword" placeholder="标题/小区" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="loadList" />
            </el-form-item>
            <el-form-item label="审核状态">
              <el-select v-model="filters.audit_status" placeholder="全部" clearable style="width: 120px" @change="loadList">
                <el-option label="待审核" :value="0" />
                <el-option label="已通过" :value="1" />
                <el-option label="已拒绝" :value="2" />
              </el-select>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :icon="Search" @click="loadList">搜索</el-button>
            </el-form-item>
          </el-form>

          <div class="toolbar">
            <el-button type="success" :icon="Check" :disabled="!selection.length" @click="onBatchAudit(1)">批量通过</el-button>
            <el-button type="danger" :icon="Close" :disabled="!selection.length" @click="onBatchAudit(2)">批量拒绝</el-button>
          </div>

          <el-table v-loading="loading" :data="list" border stripe @selection-change="onSelectionChange">
            <el-table-column type="selection" width="44" fixed="left" />
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column label="标题" min-width="200" show-overflow-tooltip>
              <template #default="{ row }">{{ row.title || `房源#${row.id}` }}</template>
            </el-table-column>
            <el-table-column label="类型" width="80">
              <template #default="{ row }">{{ row.house_type === 'rent' ? '出租' : '出售' }}</template>
            </el-table-column>
            <el-table-column label="价格" width="120">
              <template #default="{ row }">{{ formatPrice(row) }}</template>
            </el-table-column>
            <el-table-column label="审核" width="100">
              <template #default="{ row }">
                <el-tag :type="auditTagType(row.audit_status)" size="small">{{ auditText(row.audit_status) }}</el-tag>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination-wrap">
            <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="批量状态变更" name="status">
        <div class="page-card">
          <div class="toolbar">
            <el-button type="primary" :disabled="!selection.length" @click="onBatchStatus(1)">批量发布</el-button>
            <el-button type="warning" :disabled="!selection.length" @click="onBatchStatus(2)">批量下架</el-button>
          </div>
          <el-table v-loading="loading" :data="list" border stripe @selection-change="onSelectionChange">
            <el-table-column type="selection" width="44" fixed="left" />
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column label="标题" min-width="200" show-overflow-tooltip>
              <template #default="{ row }">{{ row.title || `房源#${row.id}` }}</template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="statusTagType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
              </template>
            </el-table-column>
          </el-table>
          <div class="pagination-wrap">
            <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="批量删除" name="delete">
        <div class="page-card">
          <div class="toolbar">
            <el-button type="danger" :icon="Delete" :disabled="!selection.length" @click="onBatchDelete">批量删除</el-button>
          </div>
          <el-table v-loading="loading" :data="list" border stripe @selection-change="onSelectionChange">
            <el-table-column type="selection" width="44" fixed="left" />
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column label="标题" min-width="200" show-overflow-tooltip>
              <template #default="{ row }">{{ row.title || `房源#${row.id}` }}</template>
            </el-table-column>
            <el-table-column label="创建时间" width="160">
              <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
            </el-table-column>
          </el-table>
          <div class="pagination-wrap">
            <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Check, Close, Delete } from '@element-plus/icons-vue'
import { adminListHouses, auditHouse, adminUpdateHouseStatus, deleteHouse } from '@/api/house'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filters = reactive({ keyword: '', audit_status: '' })
const selection = ref([])
const activeTab = ref('audit')

const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')
const statusText = (s) => ({ 0: '草稿', 1: '已发布', 2: '已下架', 3: '已成交' }[s] || '-')
const statusTagType = (s) => ({ 0: 'info', 1: 'success', 2: 'warning', 3: 'primary' }[s] || 'info')

const formatPrice = (row) => {
  const p = Number(row.price || 0)
  if (!p) return '面议'
  if (row.house_type === 'rent') return `${p}元/月`
  return `${(p / 10000).toFixed(1)}万`
}

const buildParams = () => {
  const p = { page: page.value, page_size: pageSize.value }
  if (filters.keyword) p.keyword = filters.keyword
  if (filters.audit_status !== '' && filters.audit_status !== null) p.audit_status = filters.audit_status
  return p
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await adminListHouses(buildParams())
    const d = res.data || {}
    list.value = d.list || []
    total.value = d.total || 0
  } catch (e) { list.value = [] } finally { loading.value = false }
}

const onSelectionChange = (val) => { selection.value = val }

const onBatchAudit = async (auditStatus) => {
  if (!selection.value.length) return
  try {
    if (auditStatus === 2) {
      const { value } = await ElMessageBox.prompt('请输入拒绝原因（可选）', '批量拒绝', { confirmButtonText: '确定', cancelButtonText: '取消' })
      await Promise.all(selection.value.map(item => auditHouse(item.id, { audit_status: auditStatus, audit_reason: value || '' })))
    } else {
      await ElMessageBox.confirm(`确定批量通过 ${selection.value.length} 条吗？`, '提示', { type: 'warning' })
      await Promise.all(selection.value.map(item => auditHouse(item.id, { audit_status: auditStatus })))
    }
    ElMessage.success('批量操作成功')
    loadList()
  } catch (e) { /* cancel */ }
}

const onBatchStatus = async (status) => {
  if (!selection.value.length) return
  try {
    const label = statusText(status)
    await ElMessageBox.confirm(`确定批量设为「${label}」${selection.value.length} 条吗？`, '提示', { type: 'warning' })
    await Promise.all(selection.value.map(item => adminUpdateHouseStatus(item.id, status)))
    ElMessage.success('批量操作成功')
    loadList()
  } catch (e) { /* cancel */ }
}

const onBatchDelete = async () => {
  if (!selection.value.length) return
  try {
    await ElMessageBox.confirm(`确定批量删除 ${selection.value.length} 条房源吗？删除后不可恢复！`, '危险操作', { type: 'error' })
    await Promise.all(selection.value.map(item => deleteHouse(item.id)))
    ElMessage.success('批量删除成功')
    loadList()
  } catch (e) { /* cancel */ }
}

onMounted(() => loadList())
</script>
