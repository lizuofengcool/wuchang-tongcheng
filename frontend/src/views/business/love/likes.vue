<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">总喜欢数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.matched }}</div><div class="stat-label">相互喜欢</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ stats.today }}</div><div class="stat-label">今日喜欢</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-danger">{{ stats.superLike }}</div><div class="stat-label">超级喜欢</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.todayQuota }}</div><div class="stat-label">今日额度</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.unmatched }}</div><div class="stat-label">单方面喜欢</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="会员ID/昵称" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="filters.action_type" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="喜欢" value="like" />
            <el-option label="不喜欢" value="dislike" />
            <el-option label="超级喜欢" value="super_like" />
          </el-select>
        </el-form-item>
        <el-form-item label="是否匹配">
          <el-select v-model="filters.is_matched" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="相互喜欢" :value="1" />
            <el-option label="单方面" :value="0" />
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
        <el-table-column label="发起人" min-width="160">
          <template #default="{ row }">
            <div class="user-cell">
              <el-image v-if="row.from_avatar" :src="row.from_avatar" fit="cover" class="user-avatar" />
              <div class="user-info">
                <div class="user-name">{{ row.from_name || `#${row.from_user_id}` }}</div>
                <div class="user-meta">{{ row.from_gender === 'male' ? '男' : '女' }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="对象" min-width="160">
          <template #default="{ row }">
            <div class="user-cell">
              <el-image v-if="row.to_avatar" :src="row.to_avatar" fit="cover" class="user-avatar" />
              <div class="user-info">
                <div class="user-name">{{ row.to_name || `#${row.to_user_id}` }}</div>
                <div class="user-meta">{{ row.to_gender === 'male' ? '男' : '女' }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="110">
          <template #default="{ row }">
            <el-tag :type="actionTagType(row.action_type)" size="small">{{ actionText(row.action_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="匹配" width="90">
          <template #default="{ row }">
            <el-tag v-if="row.is_matched" type="success" size="small">已匹配</el-tag>
            <span v-else class="text-muted">未匹配</span>
          </template>
        </el-table-column>
        <el-table-column label="留言" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ row.message || '-' }}</template>
        </el-table-column>
        <el-table-column label="时间" width="160" prop="created_at" sortable="custom">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button type="danger" link size="small" @click="onUndo(row)">撤销</el-button>
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
    <el-dialog v-model="detailVisible" title="喜欢详情" width="600px">
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="类型">
          <el-tag size="small">{{ actionText(detail.action_type) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="发起人">{{ detail.from_name || `#${detail.from_user_id}` }}</el-descriptions-item>
        <el-descriptions-item label="发起人ID">{{ detail.from_user_id }}</el-descriptions-item>
        <el-descriptions-item label="对象">{{ detail.to_name || `#${detail.to_user_id}` }}</el-descriptions-item>
        <el-descriptions-item label="对象ID">{{ detail.to_user_id }}</el-descriptions-item>
        <el-descriptions-item label="是否匹配">
          <el-tag v-if="detail.is_matched" type="success" size="small">已匹配</el-tag>
          <el-tag v-else type="info" size="small">未匹配</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="匹配ID">{{ detail.match_id || '-' }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.message" label="留言" :span="2">{{ detail.message }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
      </el-descriptions>
      <template #footer><el-button @click="detailVisible = false">关闭</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search } from '@element-plus/icons-vue'
import request from '@/utils/request'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const sortField = ref('created_at')
const sortOrder = ref('descending')

const filters = reactive({ keyword: '', action_type: '', is_matched: null, dateRange: null })

const stats = reactive({ total: 0, matched: 0, today: 0, superLike: 0, todayQuota: 0, unmatched: 0 })

const actionText = (a) => ({ like: '喜欢', dislike: '不喜欢', super_like: '超级喜欢' }[a] || a || '-')
const actionTagType = (a) => ({ like: 'success', dislike: 'info', super_like: 'danger' }[a] || 'info')

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.keyword = ''; filters.action_type = ''; filters.is_matched = null; filters.dateRange = null
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
      action_type: filters.action_type || undefined,
      is_matched: filters.is_matched === null || filters.is_matched === '' ? undefined : filters.is_matched,
      sort: sortField.value,
      order: sortOrder.value
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await request.get('/love/likes', { params })
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
    calcStats()
    if (data.today_stats) {
      stats.today = data.today_stats.total || 0
      stats.todayQuota = data.today_stats.remaining || 0
    }
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
  stats.matched = list.value.filter((r) => r.is_matched).length
  stats.unmatched = stats.total - stats.matched
  stats.superLike = list.value.filter((r) => r.action_type === 'super_like').length
  if (!stats.today) stats.today = list.value.filter((r) => (r.created_at || '').startsWith(today)).length
}

const detailVisible = ref(false)
const detail = ref(null)
const openDetail = (row) => { detail.value = row; detailVisible.value = true }

const onUndo = async (row) => {
  try {
    await ElMessageBox.confirm(`确认撤销喜欢记录 #${row.id}？`, '提示', { type: 'warning' })
    await request.delete(`/love/likes/${row.id}`)
    ElMessage.success('已撤销')
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
.user-cell { display: flex; align-items: center; gap: 8px; }
.user-avatar { width: 36px; height: 36px; border-radius: 50%; border: 1px solid #ebeef5; }
.user-info { display: flex; flex-direction: column; }
.user-name { color: #303133; font-size: 13px; }
.user-meta { color: #909399; font-size: 12px; margin-top: 2px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
