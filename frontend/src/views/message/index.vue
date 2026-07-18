<template>
  <div class="app-container">
    <div class="page-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button :icon="Refresh" @click="loadMessages(true)">刷新</el-button>
          <el-button
            type="primary"
            :icon="Check"
            :disabled="!hasUnread"
            @click="handleMarkAllRead"
          >全部已读</el-button>
          <el-button
            type="success"
            :icon="Check"
            :disabled="selectedIds.length === 0"
            @click="handleMarkSelectedRead"
          >标记选中已读</el-button>
        </div>
        <div class="toolbar-right">
          <el-radio-group v-model="filter" @change="onFilterChange">
            <el-radio-button label="all">全部 ({{ stats.total }})</el-radio-button>
            <el-radio-button label="unread">未读 ({{ stats.unread }})</el-radio-button>
            <el-radio-button label="read">已读 ({{ stats.read }})</el-radio-button>
          </el-radio-group>
        </div>
      </div>

      <el-table
        v-loading="loading"
        :data="filteredList"
        border
        stripe
        @selection-change="onSelectionChange"
      >
        <el-table-column type="selection" width="48" />
        <el-table-column label="消息" min-width="320">
          <template #default="{ row }">
            <div class="msg-cell">
              <el-tag :type="typeTag(row.type)" size="small" effect="plain">{{ typeText(row.type) }}</el-tag>
              <span class="msg-content">{{ row.content }}</span>
              <el-badge v-if="!row.is_read" is-dot type="danger" class="unread-dot" />
            </div>
          </template>
        </el-table-column>
        <el-table-column label="来源" width="120">
          <template #default="{ row }">
            <span v-if="row.news_id">
              <el-link type="primary" @click="goNews(row.news_id)">查看头条</el-link>
            </span>
            <span v-else class="muted">系统</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.is_read ? 'info' : 'danger'" size="small">
              {{ row.is_read ? '已读' : '未读' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="时间" width="170">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="!row.is_read"
              type="primary"
              link
              size="small"
              @click="handleMarkOneRead(row)"
            >标记已读</el-button>
            <span v-else class="muted">—</span>
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
          @current-change="loadMessages(false)"
          @size-change="loadMessages(true)"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Refresh, Check } from '@element-plus/icons-vue'
import { listMessages, markMessagesRead, getUnreadCount } from '@/api/news'
import { formatTime } from '@/utils/format'

const router = useRouter()

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filter = ref('all')
const unreadCount = ref(0)
const selectedIds = ref([])

const stats = reactive({
  total: 0,
  unread: 0,
  read: 0
})

const hasUnread = computed(() => stats.unread > 0)

const typeText = (t) => {
  switch (t) {
    case 'like': return '点赞'
    case 'comment': return '评论'
    case 'reply': return '回复'
    case 'system': return '系统'
    default: return '消息'
  }
}

const typeTag = (t) => {
  switch (t) {
    case 'like': return 'danger'
    case 'comment': return 'warning'
    case 'reply': return 'success'
    case 'system': return 'info'
    default: return ''
  }
}

const onFilterChange = () => {
  // 客户端筛选：后端 ListMessages 不区分已读/未读，前端按 filter 过滤展示
  // 切换 filter 不重新请求，仅前端过滤当前页
}

// 客户端筛选辅助：filter=all 展示当前页全部；filter=unread/read 仅展示对应项
const filteredList = computed(() => {
  if (filter.value === 'all') return list.value
  if (filter.value === 'unread') return list.value.filter((x) => !x.is_read)
  return list.value.filter((x) => x.is_read)
})

const onSelectionChange = (rows) => {
  selectedIds.value = rows.map((r) => r.id)
}

const loadUnread = async () => {
  try {
    const res = await getUnreadCount()
    unreadCount.value = res.data?.count || 0
    stats.unread = unreadCount.value
  } catch (e) {
    // 静默
  }
}

const loadMessages = async (reset = false) => {
  if (reset) {
    page.value = 1
  }
  loading.value = true
  try {
    const res = await listMessages({
      page: page.value,
      page_size: pageSize.value
    })
    const data = res.data || {}
    list.value = data.list || []
    stats.total = data.total || 0
    stats.unread = unreadCount.value
    stats.read = Math.max(0, stats.total - stats.unread)
    // 分页 total：filter=all 用后端 total；其他模式用当前页筛选后的条数
    total.value = filter.value === 'all' ? (data.total || 0) : filteredList.value.length
  } catch (e) {
    // 错误已由 request 拦截器提示
  } finally {
    loading.value = false
  }
}

const handleMarkOneRead = async (row) => {
  try {
    await markMessagesRead([row.id])
    row.is_read = true
    unreadCount.value = Math.max(0, unreadCount.value - 1)
    stats.unread = unreadCount.value
    stats.read = Math.max(0, stats.total - stats.unread)
    ElMessage.success('已标记已读')
  } catch (e) {
    // ignore
  }
}

const handleMarkSelectedRead = async () => {
  if (selectedIds.value.length === 0) return
  try {
    await markMessagesRead(selectedIds.value)
    let changed = 0
    list.value.forEach((x) => {
      if (selectedIds.value.includes(x.id) && !x.is_read) {
        x.is_read = true
        changed += 1
      }
    })
    unreadCount.value = Math.max(0, unreadCount.value - changed)
    stats.unread = unreadCount.value
    stats.read = Math.max(0, stats.total - stats.unread)
    ElMessage.success(`已标记 ${changed} 条已读`)
  } catch (e) {
    // ignore
  }
}

const handleMarkAllRead = async () => {
  try {
    // ids=[] 表示全部已读
    await markMessagesRead([])
    list.value.forEach((x) => { x.is_read = true })
    unreadCount.value = 0
    stats.unread = 0
    stats.read = stats.total
    ElMessage.success('全部已标记已读')
  } catch (e) {
    // ignore
  }
}

const goNews = (newsId) => {
  router.push({ name: 'NewsDetail', params: { id: newsId } })
}

onMounted(async () => {
  await Promise.all([loadUnread(), loadMessages(true)])
})
</script>

<style scoped>
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 16px;
}
.toolbar-left {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.toolbar-right {
  display: flex;
  align-items: center;
}
.msg-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}
.msg-content {
  flex: 1;
  color: #303133;
  word-break: break-word;
}
.unread-dot {
  margin-left: 4px;
}
.muted {
  color: #909399;
}
.pagination-wrap {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
