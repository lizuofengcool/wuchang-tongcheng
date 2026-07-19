<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="内容/用户" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option label="待审核" :value="0" />
            <el-option label="已通过" :value="1" />
            <el-option label="已拒绝" :value="2" />
            <el-option label="已隐藏" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="评分">
          <el-select v-model="filters.score" placeholder="全部" clearable style="width: 120px" @change="onSearch">
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

      <div class="toolbar"><el-button :icon="Refresh" @click="loadList">刷新</el-button></div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" fixed="left" />
        <el-table-column label="目标" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ row.target_name || `#${row.target_id}` }}</template>
        </el-table-column>
        <el-table-column label="用户" width="140">
          <template #default="{ row }">{{ row.user_name || `#${row.user_id}` }}</template>
        </el-table-column>
        <el-table-column label="评分" width="140">
          <template #default="{ row }">
            <el-rate v-model="row.score" disabled size="small" />
          </template>
        </el-table-column>
        <el-table-column prop="content" label="内容" min-width="200" show-overflow-tooltip />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status === 0" type="success" link size="small" @click="handleStatus(row, 1)">通过</el-button>
            <el-button v-if="row.status === 0 || row.status === 1" type="danger" link size="small" @click="handleStatus(row, 2)">拒绝</el-button>
            <el-button type="warning" link size="small" @click="openReply(row)">回复</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <el-dialog v-model="detailVisible" title="评价详情" width="700px" destroy-on-close>
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="目标">{{ detail.target_name || `#${detail.target_id}` }}</el-descriptions-item>
        <el-descriptions-item label="用户">{{ detail.user_name || `#${detail.user_id}` }}</el-descriptions-item>
        <el-descriptions-item label="评分">{{ detail.score }}星</el-descriptions-item>
        <el-descriptions-item label="状态">{{ statusText(detail.status) }}</el-descriptions-item>
        <el-descriptions-item label="点赞">{{ detail.like_count || 0 }}</el-descriptions-item>
        <el-descriptions-item label="内容" :span="2">{{ detail.content }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.reply" label="回复" :span="2">{{ detail.reply }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="回复时间">{{ formatTime(detail.replied_at) }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>

    <el-dialog v-model="replyVisible" title="回复评价" width="500px" destroy-on-close>
      <el-form :model="replyForm" label-width="100px">
        <el-form-item label="回复内容">
          <el-input v-model="replyForm.reply" type="textarea" :rows="4" placeholder="请输入回复内容" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="replyVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleReply">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, RefreshLeft, Refresh } from '@element-plus/icons-vue'
import { adminListReviews, updateReviewStatus, replyReview, adminDeleteReview, getReview } from '@/api/house'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filters = reactive({ keyword: '', status: '', score: '' })

const detailVisible = ref(false)
const detail = ref(null)
const replyVisible = ref(false)
const submitting = ref(false)
const replyForm = reactive({ id: null, reply: '' })

const statusText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝', 3: '已隐藏' }[s] || '-')
const statusTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger', 3: 'info' }[s] || 'info')

const buildParams = () => {
  const p = { page: page.value, page_size: pageSize.value }
  if (filters.keyword) p.keyword = filters.keyword
  if (filters.status !== '' && filters.status !== null) p.status = filters.status
  if (filters.score !== '' && filters.score !== null) p.score = filters.score
  return p
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await adminListReviews(buildParams())
    const d = res.data || {}
    list.value = d.list || []
    total.value = d.total || 0
  } catch (e) { list.value = [] } finally { loading.value = false }
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => { Object.assign(filters, { keyword: '', status: '', score: '' }); onSearch() }

const openDetail = async (row) => {
  try {
    const res = await getReview(row.id)
    detail.value = res.data || row
  } catch (e) { detail.value = row }
  detailVisible.value = true
}

const handleStatus = async (row, status) => {
  try {
    const label = statusText(status)
    await ElMessageBox.confirm(`确定设为「${label}」吗？`, '提示', { type: 'warning' })
    await updateReviewStatus(row.id, { status })
    ElMessage.success('操作成功')
    loadList()
  } catch (e) { /* cancel */ }
}

const openReply = (row) => {
  Object.assign(replyForm, { id: row.id, reply: row.reply || '' })
  replyVisible.value = true
}

const handleReply = async () => {
  try {
    submitting.value = true
    await replyReview(replyForm.id, { reply: replyForm.reply })
    ElMessage.success('回复成功')
    replyVisible.value = false
    loadList()
  } catch (e) {
    if (e?.message) ElMessage.error(e.message)
  } finally { submitting.value = false }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定删除该评价吗？', '危险操作', { type: 'error' })
    await adminDeleteReview(row.id)
    ElMessage.success('已删除')
    loadList()
  } catch (e) { /* cancel */ }
}

onMounted(() => loadList())
</script>
