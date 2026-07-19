<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="内容/用户" clearable style="width: 200px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="待审核" :value="0" />
            <el-option label="已通过" :value="1" />
            <el-option label="已拒绝" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="评分">
          <el-select v-model="filters.rating" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="1星" :value="1" />
            <el-option label="2星" :value="2" />
            <el-option label="3星" :value="3" />
            <el-option label="4星" :value="4" />
            <el-option label="5星" :value="5" />
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
        <el-table-column label="评价内容" min-width="240">
          <template #default="{ row }">
            <div class="review-content">{{ row.content || row.comment || '-' }}</div>
            <div v-if="row.reply" class="review-reply">回复：{{ row.reply }}</div>
          </template>
        </el-table-column>
        <el-table-column label="评分" width="120">
          <template #default="{ row }">
            <el-rate :model-value="row.rating || 5" disabled size="small" />
          </template>
        </el-table-column>
        <el-table-column label="评价人" width="120">
          <template #default="{ row }">{{ row.user_name || `用户#${row.user_id}` }}</template>
        </el-table-column>
        <el-table-column label="评价对象" width="160">
          <template #default="{ row }">{{ row.target_title || `#${row.target_id}` }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="reviewStatusType(row.status)" size="small">{{ reviewStatusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="时间" width="160" prop="created_at" sortable>
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button v-if="row.status === 0" type="success" link size="small" @click="handleAudit(row, 1)">通过</el-button>
            <el-button v-if="row.status === 0" type="danger" link size="small" @click="handleAudit(row, 2)">拒绝</el-button>
            <el-button type="primary" link size="small" @click="openReply(row)">回复</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <el-dialog v-model="replyVisible" title="回复评价" width="560px">
      <el-form :model="replyForm" label-width="80px">
        <el-form-item label="评价内容">
          <div class="review-content">{{ replyForm.content }}</div>
        </el-form-item>
        <el-form-item label="回复内容">
          <el-input v-model="replyForm.reply" type="textarea" :rows="4" placeholder="请输入回复内容" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="replyVisible = false">取消</el-button>
        <el-button type="primary" :loading="replyLoading" @click="onReply">确认回复</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search } from '@element-plus/icons-vue'
import { listReviews, auditReview, replyReview, deleteReview } from '@/api/job'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const filters = reactive({ keyword: '', status: null, rating: null })

const reviewStatusText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const reviewStatusType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => { Object.assign(filters, { keyword: '', status: null, rating: null }); page.value = 1; loadList() }

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value, page_size: pageSize.value,
      keyword: filters.keyword || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status,
      rating: filters.rating || undefined
    }
    const res = await listReviews(params)
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
  } catch (e) { list.value = []; total.value = 0 } finally { loading.value = false }
}

const handleAudit = async (row, status) => {
  try {
    const action = status === 1 ? '通过' : '拒绝'
    await ElMessageBox.confirm(`确定${action}该评价吗？`, '提示', { type: 'warning' })
    await auditReview(row.id, { status })
    ElMessage.success(`已${action}`); await loadList()
  } catch (e) { /* */ }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定删除该评价吗？', '危险操作', { type: 'error' })
    await deleteReview(row.id)
    ElMessage.success('已删除'); await loadList()
  } catch (e) { /* */ }
}

const replyVisible = ref(false)
const replyLoading = ref(false)
const replyForm = reactive({ id: null, content: '', reply: '' })
const openReply = (row) => { Object.assign(replyForm, { id: row.id, content: row.content || row.comment || '', reply: row.reply || '' }); replyVisible.value = true }
const onReply = async () => {
  try {
    replyLoading.value = true
    await replyReview(replyForm.id, { reply: replyForm.reply })
    ElMessage.success('回复成功'); replyVisible.value = false; await loadList()
  } catch (e) { /* */ } finally { replyLoading.value = false }
}

onMounted(() => loadList())
</script>

<style scoped>
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; margin-bottom: 12px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.review-content { line-height: 1.5; }
.review-reply { color: #999; font-size: 12px; margin-top: 4px; }
</style>
