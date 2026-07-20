<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">总匹配数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.active }}</div><div class="stat-label">活跃匹配</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ stats.today }}</div><div class="stat-label">今日匹配</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.dissolved }}</div><div class="stat-label">已解除</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="会员ID/昵称" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="匹配类型">
          <el-select v-model="filters.match_type" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option label="相互喜欢" value="like_match" />
            <el-option label="系统推荐" value="system_match" />
            <el-option label="超级喜欢" value="super_like" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="匹配中" value="matched" />
            <el-option label="已解除" value="dissolved" />
            <el-option label="暂停" value="paused" />
          </el-select>
        </el-form-item>
        <el-form-item label="最低匹配分">
          <el-input-number v-model="filters.min_score" :min="0" :max="100" :controls="false" style="width: 100px" @change="onSearch" />
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

      <div class="toolbar">
        <el-button :icon="Refresh" @click="loadList">刷新</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe @sort-change="onSortChange">
        <el-table-column prop="id" label="ID" width="70" sortable="custom" />
        <el-table-column label="会员A" min-width="160">
          <template #default="{ row }">
            <div class="user-cell">
              <el-image v-if="row.user_a_avatar" :src="row.user_a_avatar" fit="cover" class="user-avatar" />
              <div class="user-info">
                <div class="user-name">{{ row.user_a_name || `#${row.user_a_id}` }}</div>
                <div class="user-meta">{{ row.user_a_gender === 'male' ? '男' : '女' }} · {{ row.user_a_age || '?' }}岁</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="会员B" min-width="160">
          <template #default="{ row }">
            <div class="user-cell">
              <el-image v-if="row.user_b_avatar" :src="row.user_b_avatar" fit="cover" class="user-avatar" />
              <div class="user-info">
                <div class="user-name">{{ row.user_b_name || `#${row.user_b_id}` }}</div>
                <div class="user-meta">{{ row.user_b_gender === 'male' ? '男' : '女' }} · {{ row.user_b_age || '?' }}岁</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="匹配分" width="120" prop="match_score" sortable="custom">
          <template #default="{ row }">
            <el-progress :percentage="Number(row.match_score || 0)" :stroke-width="14" :text-inside="true" />
          </template>
        </el-table-column>
        <el-table-column label="类型" width="110">
          <template #default="{ row }">
            <el-tag :type="typeTagType(row.match_type)" size="small">{{ typeText(row.match_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="会话" width="80">
          <template #default="{ row }">
            <el-tag v-if="row.chat_session_id" type="success" size="small">已建立</el-tag>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="匹配时间" width="160" prop="created_at" sortable="custom">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button
              v-if="row.status === 'matched'"
              type="warning"
              link
              size="small"
              @click="onDissolve(row)"
            >解除匹配</el-button>
            <el-button
              v-if="row.status === 'matched' && !row.chat_session_id"
              type="success"
              link
              size="small"
              @click="onCreateSession(row)"
            >创建会话</el-button>
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
    <el-dialog v-model="detailVisible" title="匹配详情" width="700px">
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="匹配分">{{ detail.match_score || 0 }}</el-descriptions-item>
        <el-descriptions-item label="会员A">{{ detail.user_a_name || `#${detail.user_a_id}` }}</el-descriptions-item>
        <el-descriptions-item label="会员A ID">{{ detail.user_a_id }}</el-descriptions-item>
        <el-descriptions-item label="会员B">{{ detail.user_b_name || `#${detail.user_b_id}` }}</el-descriptions-item>
        <el-descriptions-item label="会员B ID">{{ detail.user_b_id }}</el-descriptions-item>
        <el-descriptions-item label="匹配类型">
          <el-tag size="small">{{ typeText(detail.match_type) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTagType(detail.status)" size="small">{{ statusText(detail.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="会话ID">{{ detail.chat_session_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="解除人">{{ detail.dissolved_by_name || (detail.dissolved_by ? `#${detail.dissolved_by}` : '-') }}</el-descriptions-item>
        <el-descriptions-item label="解除时间">{{ formatTime(detail.dissolved_at) }}</el-descriptions-item>
        <el-descriptions-item label="匹配时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
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
import request from '@/utils/request'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const sortField = ref('created_at')
const sortOrder = ref('descending')

const filters = reactive({
  keyword: '', match_type: '', status: '', min_score: undefined, dateRange: null
})

const stats = reactive({ total: 0, active: 0, today: 0, dissolved: 0 })

const typeText = (t) => ({
  like_match: '相互喜欢', system_match: '系统推荐', super_like: '超级喜欢'
}[t] || t || '-')
const typeTagType = (t) => ({
  like_match: 'success', system_match: 'primary', super_like: 'danger'
}[t] || 'info')
const statusText = (s) => ({ matched: '匹配中', dissolved: '已解除', paused: '暂停' }[s] || '-')
const statusTagType = (s) => ({ matched: 'success', dissolved: 'info', paused: 'warning' }[s] || 'info')

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.keyword = ''; filters.match_type = ''; filters.status = ''; filters.min_score = undefined; filters.dateRange = null
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
      match_type: filters.match_type || undefined,
      status: filters.status || undefined,
      min_score: filters.min_score || undefined,
      sort: sortField.value,
      order: sortOrder.value
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await request.get('/love/matches', { params })
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
    calcStats()
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
  stats.active = list.value.filter((r) => r.status === 'matched').length
  stats.dissolved = list.value.filter((r) => r.status === 'dissolved').length
  stats.today = list.value.filter((r) => (r.created_at || '').startsWith(today)).length
}

// 详情
const detailVisible = ref(false)
const detail = ref(null)
const openDetail = (row) => { detail.value = row; detailVisible.value = true }

// 解除匹配
const onDissolve = async (row) => {
  try {
    await ElMessageBox.confirm(`确认解除匹配 #${row.id}？解除后双方将无法继续聊天。`, '解除匹配', { type: 'warning' })
    await request.post(`/love/matches/${row.id}/dissolve`)
    ElMessage.success('已解除匹配')
    await loadList()
  } catch (e) { /* cancel */ }
}

// 创建会话
const onCreateSession = async (row) => {
  try {
    await request.post(`/love/chat-sessions/${row.id}/action`, { action: 'create' })
    ElMessage.success('会话已创建')
    await loadList()
  } catch (e) { /* fail */ }
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
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: flex-start; margin-bottom: 12px; }
.text-muted { color: #909399; }
.user-cell { display: flex; align-items: center; gap: 8px; }
.user-avatar {
  width: 36px; height: 36px; border-radius: 50%; border: 1px solid #ebeef5;
}
.user-info { display: flex; flex-direction: column; }
.user-name { color: #303133; font-size: 13px; }
.user-meta { color: #909399; font-size: 12px; margin-top: 2px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
