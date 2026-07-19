<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="姓名/求职意向" clearable style="width: 200px" @keyup.enter="onSearch" />
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

      <div class="toolbar"><el-button :icon="Refresh" @click="loadList">刷新</el-button></div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column label="求职者" width="130">
          <template #default="{ row }">{{ row.user_name || `用户#${row.user_id}` }}</template>
        </el-table-column>
        <el-table-column prop="title" label="简历标题" min-width="160" />
        <el-table-column prop="intention" label="求职意向" min-width="140" show-overflow-tooltip />
        <el-table-column prop="expected_salary" label="期望薪资" width="120" />
        <el-table-column prop="education" label="学历" width="80" />
        <el-table-column prop="experience" label="经验" width="80" />
        <el-table-column label="默认" width="70">
          <template #default="{ row }">
            <el-tag v-if="row.is_default" type="success" size="small">默认</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="更新时间" width="160" prop="updated_at" sortable>
          <template #default="{ row }">{{ formatTime(row.updated_at || row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status === 1" type="warning" link size="small" @click="handleStatus(row, 0)">禁用</el-button>
            <el-button v-else type="success" link size="small" @click="handleStatus(row, 1)">启用</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <el-dialog v-model="detailVisible" title="简历详情" width="800px">
      <div v-loading="detailLoading">
        <el-descriptions v-if="detail" :column="2" border>
          <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
          <el-descriptions-item label="求职者">{{ detail.user_name || `用户#${detail.user_id}` }}</el-descriptions-item>
          <el-descriptions-item label="简历标题">{{ detail.title }}</el-descriptions-item>
          <el-descriptions-item label="求职意向">{{ detail.intention || '-' }}</el-descriptions-item>
          <el-descriptions-item label="期望薪资">{{ detail.expected_salary || '-' }}</el-descriptions-item>
          <el-descriptions-item label="学历">{{ detail.education || '-' }}</el-descriptions-item>
          <el-descriptions-item label="经验">{{ detail.experience || '-' }}</el-descriptions-item>
          <el-descriptions-item label="期望城市">{{ detail.expected_city || '-' }}</el-descriptions-item>
          <el-descriptions-item v-if="detail.skills" label="技能" :span="2">{{ Array.isArray(detail.skills) ? detail.skills.join('、') : detail.skills }}</el-descriptions-item>
          <el-descriptions-item v-if="detail.self_intro" label="自我介绍" :span="2">{{ detail.self_intro }}</el-descriptions-item>
          <el-descriptions-item v-if="detail.work_exp" label="工作经历" :span="2">{{ detail.work_exp }}</el-descriptions-item>
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
import { listResumes, getResume, updateResumeStatus, deleteResume } from '@/api/job'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1); const pageSize = ref(20); const total = ref(0); const list = ref([])
const filters = reactive({ keyword: '', status: null })

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => { Object.assign(filters, { keyword: '', status: null }); page.value = 1; loadList() }

const loadList = async () => {
  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value, keyword: filters.keyword || undefined, status: filters.status === null || filters.status === '' ? undefined : filters.status }
    const res = await listResumes(params)
    const data = res.data || {}
    list.value = data.list || []; total.value = data.total || 0
  } catch (e) { list.value = []; total.value = 0 } finally { loading.value = false }
}

const detailVisible = ref(false); const detailLoading = ref(false); const detail = ref(null)
const openDetail = async (row) => {
  detailVisible.value = true; detailLoading.value = true
  try { const res = await getResume(row.id); detail.value = res.data || row } catch (e) { detail.value = row } finally { detailLoading.value = false }
}

const handleStatus = async (row, status) => {
  try {
    await ElMessageBox.confirm(`确定${status === 1 ? '启用' : '禁用'}该简历吗？`, '提示', { type: 'warning' })
    await updateResumeStatus(row.id, { status })
    ElMessage.success('操作成功'); await loadList()
  } catch (e) { /* */ }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定删除该简历吗？', '危险操作', { type: 'error' })
    await deleteResume(row.id)
    ElMessage.success('已删除'); await loadList()
  } catch (e) { /* */ }
}

onMounted(() => loadList())
</script>

<style scoped>
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; margin-bottom: 12px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
