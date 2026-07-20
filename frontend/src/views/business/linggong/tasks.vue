<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">任务总数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ stats.claimed }}</div><div class="stat-label">已领取</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.pending }}</div><div class="stat-label">待验收</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.completed }}</div><div class="stat-label">已完成</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="任务名/编号" clearable style="width: 200px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option v-for="(label, val) in statusMap" :key="val" :label="label" :value="Number(val)" />
          </el-select>
        </el-form-item>
        <el-form-item label="雇主ID">
          <el-input v-model="filters.employer_id" placeholder="雇主ID" clearable style="width: 140px" @keyup.enter="onSearch" />
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
        <el-table-column prop="task_no" label="任务编号" width="180" />
        <el-table-column prop="title" label="任务名" min-width="180" show-overflow-tooltip />
        <el-table-column label="岗位ID" width="100">
          <template #default="{ row }">{{ row.linggong_id || '-' }}</template>
        </el-table-column>
        <el-table-column label="雇主" width="140">
          <template #default="{ row }">
            <span>{{ row.employer_name || `#${row.employer_id}` }}</span>
          </template>
        </el-table-column>
        <el-table-column label="单价" width="100">
          <template #default="{ row }">¥{{ Number(row.unit_price || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="总数" width="80" prop="total_count" />
        <el-table-column label="已领" width="80" prop="claimed_count" />
        <el-table-column label="已完" width="80" prop="completed_count" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusMap[row.status] || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="截止时间" width="160">
          <template #default="{ row }">{{ formatTime(row.deadline) }}</template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button type="warning" link size="small" @click="openStatus(row)">改状态</el-button>
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
    <el-dialog v-model="detailVisible" title="任务详情" width="700px">
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="任务编号">{{ detail.task_no }}</el-descriptions-item>
        <el-descriptions-item label="任务名" :span="2">{{ detail.title }}</el-descriptions-item>
        <el-descriptions-item label="岗位ID">{{ detail.linggong_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="雇主">{{ detail.employer_name || `#${detail.employer_id}` }}</el-descriptions-item>
        <el-descriptions-item label="单价">¥{{ Number(detail.unit_price || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="总数">{{ detail.total_count }}</el-descriptions-item>
        <el-descriptions-item label="已领取">{{ detail.claimed_count }}</el-descriptions-item>
        <el-descriptions-item label="已完成">{{ detail.completed_count }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTagType(detail.status)" size="small">{{ statusMap[detail.status] || '-' }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="截止时间">{{ formatTime(detail.deadline) }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.description" label="描述" :span="2">{{ detail.description }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
      </el-descriptions>
      <template #footer><el-button @click="detailVisible = false">关闭</el-button></template>
    </el-dialog>

    <!-- 状态变更弹窗 -->
    <el-dialog v-model="statusVisible" title="任务状态变更" width="500px">
      <el-form :model="statusForm" label-width="100px">
        <el-form-item label="任务">
          <span>{{ statusForm.title }} (#{{ statusForm.id }})</span>
        </el-form-item>
        <el-form-item label="目标状态">
          <el-select v-model="statusForm.status" style="width: 100%">
            <el-option v-for="(label, val) in statusMap" :key="val" :label="label" :value="Number(val)" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="statusVisible = false">取消</el-button>
        <el-button type="primary" :loading="statusLoading" @click="onStatusSubmit">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search } from '@element-plus/icons-vue'
import { adminListLinggongTasks, adminGetLinggongTask, adminUpdateLinggongTaskStatus } from '@/api/linggong'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const filters = reactive({ keyword: '', status: null, employer_id: '' })

const statusMap = {
  0: '草稿', 1: '已发布', 2: '进行中', 3: '待验收',
  4: '已完成', 5: '已取消', 6: '已过期'
}
const statusTagType = (s) => ({
  0: 'info', 1: 'success', 2: 'primary', 3: 'warning',
  4: 'success', 5: 'danger', 6: 'info'
}[s] || 'info')

const stats = computed(() => {
  const total = list.value.length
  const claimed = list.value.filter((r) => r.claimed_count > 0).length
  const pending = list.value.filter((r) => r.status === 3).length
  const completed = list.value.filter((r) => r.status === 4).length
  return { total, claimed, pending, completed }
})

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.keyword = ''
  filters.status = null
  filters.employer_id = ''
  page.value = 1
  loadList()
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await adminListLinggongTasks({
      page: page.value,
      page_size: pageSize.value,
      keyword: filters.keyword || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status,
      employer_id: filters.employer_id || undefined
    })
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
  } catch (e) {
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const detailVisible = ref(false)
const detail = ref(null)

const openDetail = async (row) => {
  try {
    const res = await adminGetLinggongTask(row.id)
    detail.value = res.data || row
  } catch (e) {
    detail.value = row
  }
  detailVisible.value = true
}

const statusVisible = ref(false)
const statusLoading = ref(false)
const statusForm = reactive({ id: null, title: '', status: 0 })

const openStatus = (row) => {
  statusForm.id = row.id
  statusForm.title = row.title
  statusForm.status = row.status
  statusVisible.value = true
}

const onStatusSubmit = async () => {
  statusLoading.value = true
  try {
    await adminUpdateLinggongTaskStatus(statusForm.id, statusForm.status)
    ElMessage.success('状态更新成功')
    statusVisible.value = false
    await loadList()
  } catch (e) {
    // 失败已提示
  } finally {
    statusLoading.value = false
  }
}

onMounted(() => loadList())
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-item { text-align: center; padding: 16px; }
.stat-value { font-size: 22px; font-weight: 600; color: #303133; }
.text-primary { color: #409eff; }
.text-warning { color: #e6a23c; }
.text-success { color: #67c23a; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: flex-start; margin-bottom: 12px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
