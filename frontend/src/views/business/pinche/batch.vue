<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #409eff">
            <el-icon :size="22"><Operation /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.totalTasks }}</div>
            <div class="stat-label">总任务数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #e6a23c">
            <el-icon :size="22"><Clock /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.pending }}</div>
            <div class="stat-label">待执行</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #722ed1">
            <el-icon :size="22"><Loading /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.running }}</div>
            <div class="stat-label">执行中</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #67c23a">
            <el-icon :size="22"><CircleCheckFilled /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.completed }}</div>
            <div class="stat-label">已完成</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <!-- Tabs 切换：历史任务 / 创建新任务 -->
      <el-tabs v-model="activeTab" class="batch-tabs">
        <el-tab-pane label="批量任务历史" name="history">
          <!-- 筛选区 -->
          <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
            <el-form-item label="关键词">
              <el-input
                v-model="filters.keyword"
                placeholder="任务名称/操作人"
                clearable
                style="width: 220px"
                :prefix-icon="Search"
                @keyup.enter="onSearch"
              />
            </el-form-item>
            <el-form-item label="操作类型">
              <el-select v-model="filters.action" placeholder="全部" clearable style="width: 160px" @change="onSearch">
                <el-option label="批量审核通过" value="batch_audit_pass" />
                <el-option label="批量审核拒绝" value="batch_audit_reject" />
                <el-option label="批量发布" value="batch_publish" />
                <el-option label="批量下架" value="batch_offline" />
                <el-option label="批量删除" value="batch_delete" />
                <el-option label="批量改状态" value="batch_status" />
              </el-select>
            </el-form-item>
            <el-form-item label="目标对象">
              <el-select v-model="filters.target_type" placeholder="全部" clearable style="width: 140px" @change="onSearch">
                <el-option label="拼车信息" value="pinche" />
                <el-option label="路线" value="route" />
                <el-option label="行程" value="trip" />
                <el-option label="预订" value="booking" />
                <el-option label="车主" value="driver" />
                <el-option label="车辆" value="vehicle" />
                <el-option label="评价" value="rating" />
              </el-select>
            </el-form-item>
            <el-form-item label="状态">
              <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
                <el-option label="待执行" :value="0" />
                <el-option label="执行中" :value="1" />
                <el-option label="已完成" :value="2" />
                <el-option label="失败" :value="3" />
                <el-option label="已取消" :value="4" />
              </el-select>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
              <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
            </el-form-item>
          </el-form>

          <!-- 工具栏 -->
          <div class="toolbar">
            <div class="toolbar-left">
              <el-button :icon="Refresh" @click="loadList">刷新</el-button>
              <el-button type="warning" :icon="VideoPause" :disabled="!selection.length" @click="onBatchCancel">批量取消</el-button>
              <el-button type="danger" :icon="Delete" :disabled="!selection.length" @click="onBatchDelete">批量删除</el-button>
            </div>
            <div class="toolbar-right">
              <el-button type="primary" :icon="Plus" @click="activeTab = 'create'">新建批量任务</el-button>
            </div>
          </div>

          <!-- 表格 -->
          <el-table
            v-loading="loading"
            :data="list"
            border
            stripe
            @selection-change="onSelectionChange"
          >
            <el-table-column type="selection" width="44" fixed="left" />
            <el-table-column prop="id" label="ID" width="70" fixed="left" />
            <el-table-column prop="name" label="任务名称" min-width="180">
              <template #default="{ row }">
                <el-link type="primary" :underline="'never'" @click="openDetail(row)">{{ row.name }}</el-link>
              </template>
            </el-table-column>
            <el-table-column label="操作类型" width="140">
              <template #default="{ row }">
                <el-tag :type="actionTagType(row.action)" size="small">{{ actionText(row.action) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="目标对象" width="100">
              <template #default="{ row }">
                <el-tag size="small" effect="plain">{{ targetTypeText(row.target_type) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="进度" width="200">
              <template #default="{ row }">
                <el-progress
                  :percentage="calcProgress(row)"
                  :status="progressStatus(row.status)"
                />
                <div class="text-muted">{{ row.success_count || 0 }}/{{ row.total_count || 0 }} 成功</div>
              </template>
            </el-table-column>
            <el-table-column label="操作人" width="120">
              <template #default="{ row }">{{ row.operator_name || `管理员#${row.operator_id}` }}</template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="statusTagType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="创建时间" width="160" prop="created_at" sortable>
              <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="160" fixed="right">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
                <el-button
                  v-if="row.status === 0 || row.status === 1"
                  type="warning"
                  link
                  size="small"
                  @click="onCancel(row)"
                >取消</el-button>
                <el-button
                  v-if="row.status === 2"
                  type="success"
                  link
                  size="small"
                  @click="onRetry(row)"
                >重试</el-button>
              </template>
            </el-table-column>
          </el-table>

          <!-- 分页 -->
          <div class="pagination-wrap">
            <el-pagination
              v-model:current-page="page"
              v-model:page-size="pageSize"
              :total="total"
              :page-sizes="[10, 20, 50, 100]"
              layout="total, sizes, prev, pager, next, jumper"
              background
              @current-change="loadList"
              @size-change="loadList"
            />
          </div>
        </el-tab-pane>

        <el-tab-pane label="新建批量任务" name="create">
          <el-card>
            <el-form ref="formRef" :model="form" :rules="rules" label-width="120px" style="max-width: 800px">
              <el-form-item label="任务名称" prop="name">
                <el-input v-model="form.name" placeholder="请输入批量任务名称" maxlength="50" show-word-limit />
              </el-form-item>
              <el-form-item label="目标对象类型" prop="target_type">
                <el-select v-model="form.target_type" placeholder="请选择" style="width: 100%" @change="onTargetTypeChange">
                  <el-option label="拼车信息" value="pinche" />
                  <el-option label="路线" value="route" />
                  <el-option label="行程" value="trip" />
                  <el-option label="预订" value="booking" />
                  <el-option label="车主" value="driver" />
                  <el-option label="车辆" value="vehicle" />
                  <el-option label="评价" value="rating" />
                </el-select>
              </el-form-item>
              <el-form-item label="操作类型" prop="action">
                <el-select v-model="form.action" placeholder="请选择" style="width: 100%">
                  <el-option label="批量审核通过" value="batch_audit_pass" />
                  <el-option label="批量审核拒绝" value="batch_audit_reject" />
                  <el-option label="批量发布" value="batch_publish" />
                  <el-option label="批量下架" value="batch_offline" />
                  <el-option label="批量删除" value="batch_delete" />
                  <el-option label="批量改状态" value="batch_status" />
                </el-select>
              </el-form-item>
              <el-form-item v-if="form.action === 'batch_status'" label="目标状态">
                <el-select v-model="form.target_status" placeholder="请选择" style="width: 100%">
                  <el-option label="草稿" :value="0" />
                  <el-option label="已发布" :value="1" />
                  <el-option label="已满员" :value="2" />
                  <el-option label="已下架" :value="3" />
                  <el-option label="已过期" :value="4" />
                  <el-option label="已取消" :value="5" />
                </el-select>
              </el-form-item>
              <el-form-item label="目标 ID 列表" prop="target_ids">
                <el-input
                  v-model="form.target_ids_text"
                  type="textarea"
                  :rows="6"
                  placeholder="请输入目标 ID 列表，每行一个或以英文逗号分隔，如：1,2,3 或 1&#10;2&#10;3"
                />
                <div class="form-tip">
                  <el-link type="primary" :underline="'never'" @click="onLoadFromFilter">从筛选条件加载</el-link>
                  <el-link type="success" :underline="'never'" style="margin-left: 12px" @click="onLoadAll">加载全部</el-link>
                </div>
              </el-form-item>
              <el-form-item label="执行方式">
                <el-radio-group v-model="form.run_mode">
                  <el-radio value="sync">同步执行</el-radio>
                  <el-radio value="async">异步执行（推荐）</el-radio>
                </el-radio-group>
              </el-form-item>
              <el-form-item label="失败处理">
                <el-radio-group v-model="form.on_failure">
                  <el-radio value="stop">遇到失败停止</el-radio>
                  <el-radio value="continue">继续执行剩余</el-radio>
                </el-radio-group>
              </el-form-item>
              <el-form-item label="备注">
                <el-input
                  v-model="form.description"
                  type="textarea"
                  :rows="3"
                  placeholder="任务说明（选填）"
                />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" :icon="Upload" @click="submitForm">提交任务</el-button>
                <el-button @click="onFormReset">重置</el-button>
                <el-button @click="activeTab = 'history'">返回历史</el-button>
              </el-form-item>
            </el-form>
          </el-card>
        </el-tab-pane>
      </el-tabs>
    </div>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="批量任务详情" width="900px" destroy-on-close>
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="任务名称">{{ detail.name }}</el-descriptions-item>
        <el-descriptions-item label="操作类型">
          <el-tag :type="actionTagType(detail.action)" size="small">{{ actionText(detail.action) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="目标对象">{{ targetTypeText(detail.target_type) }}</el-descriptions-item>
        <el-descriptions-item label="总数">{{ detail.total_count || 0 }}</el-descriptions-item>
        <el-descriptions-item label="成功数">
          <span class="text-success">{{ detail.success_count || 0 }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="失败数">
          <span class="text-danger">{{ detail.fail_count || 0 }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="跳过数">{{ detail.skip_count || 0 }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTagType(detail.status)" size="small">{{ statusText(detail.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="操作人">{{ detail.operator_name || `管理员#${detail.operator_id}` }}</el-descriptions-item>
        <el-descriptions-item label="执行方式">{{ detail.run_mode === 'async' ? '异步' : '同步' }}</el-descriptions-item>
        <el-descriptions-item label="失败处理">{{ detail.on_failure === 'continue' ? '继续执行' : '停止' }}</el-descriptions-item>
        <el-descriptions-item label="开始时间">{{ formatTime(detail.started_at) }}</el-descriptions-item>
        <el-descriptions-item label="完成时间">{{ formatTime(detail.completed_at) }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.description" label="备注" :span="2">{{ detail.description }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.error_message" label="错误信息" :span="2">
          <div class="error-msg">{{ detail.error_message }}</div>
        </el-descriptions-item>
      </el-descriptions>

      <el-divider content-position="left">执行进度</el-divider>
      <el-progress
        :percentage="calcProgress(detail)"
        :status="progressStatus(detail.status)"
        :stroke-width="20"
        :text-inside="true"
      />

      <el-divider v-if="detail?.failed_items?.length" content-position="left">失败项目</el-divider>
      <el-table v-if="detail?.failed_items?.length" :data="detail.failed_items" border size="small" max-height="240">
        <el-table-column prop="target_id" label="目标 ID" width="100" />
        <el-table-column prop="error" label="失败原因" min-width="280" />
      </el-table>

      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/utils/request'
import {
  Refresh, RefreshLeft, Search, Plus, Delete, Upload,
  Operation, Clock, Loading, CircleCheckFilled, VideoPause
} from '@element-plus/icons-vue'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const selection = ref([])

const stats = reactive({ totalTasks: 0, pending: 0, running: 0, completed: 0 })

const activeTab = ref('history')

const filters = reactive({
  keyword: '', action: '', target_type: '', status: null
})

const statusText = (s) => ({ 0: '待执行', 1: '执行中', 2: '已完成', 3: '失败', 4: '已取消' }[s] || '-')
const statusTagType = (s) => ({ 0: 'info', 1: 'warning', 2: 'success', 3: 'danger', 4: 'info' }[s] || 'info')
const actionText = (a) => ({
  batch_audit_pass: '批量审核通过', batch_audit_reject: '批量审核拒绝',
  batch_publish: '批量发布', batch_offline: '批量下架',
  batch_delete: '批量删除', batch_status: '批量改状态'
}[a] || a || '-')
const actionTagType = (a) => ({
  batch_audit_pass: 'success', batch_audit_reject: 'danger',
  batch_publish: 'success', batch_offline: 'warning',
  batch_delete: 'danger', batch_status: 'primary'
}[a] || 'info')
const targetTypeText = (t) => ({
  pinche: '拼车信息', route: '路线', trip: '行程', booking: '预订',
  driver: '车主', vehicle: '车辆', rating: '评价'
}[t] || t || '-')

const calcProgress = (row) => {
  if (!row) return 0
  if (row.status === 2) return 100
  if (row.status === 4) return 0
  if (!row.total_count) return 0
  return Math.round(((row.success_count || 0) + (row.fail_count || 0) + (row.skip_count || 0)) / row.total_count * 100)
}
const progressStatus = (s) => {
  if (s === 2) return 'success'
  if (s === 3) return 'exception'
  return undefined
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.keyword = ''; filters.action = ''
  filters.target_type = ''; filters.status = null
  page.value = 1; loadList()
}

const onSelectionChange = (rows) => { selection.value = rows }

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      keyword: filters.keyword || undefined,
      action: filters.action || undefined,
      target_type: filters.target_type || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status
    }
    const res = await request.get('/pinche/admin/batch-tasks', { params })
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
    if (data.stats) {
      stats.totalTasks = data.stats.total || 0
      stats.pending = data.stats.pending || 0
      stats.running = data.stats.running || 0
      stats.completed = data.stats.completed || 0
    } else {
      stats.totalTasks = total.value
      stats.pending = list.value.filter(i => i.status === 0).length
      stats.running = list.value.filter(i => i.status === 1).length
      stats.completed = list.value.filter(i => i.status === 2).length
    }
  } catch (e) {
    list.value = []; total.value = 0
  } finally {
    loading.value = false
  }
}

const detailVisible = ref(false)
const detail = ref(null)
const openDetail = async (row) => {
  detail.value = { ...row }
  detailVisible.value = true
  try {
    const res = await request.get(`/pinche/admin/batch-tasks/${row.id}`)
    if (res.data) detail.value = res.data
  } catch (e) { /* ignore */ }
}

const onCancel = async (row) => {
  try {
    await ElMessageBox.confirm(`确认取消任务 "${row.name}"？`, '提示', { type: 'warning' })
    await request.put(`/pinche/admin/batch-tasks/${row.id}/cancel`)
    ElMessage.success('任务已取消')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onRetry = async (row) => {
  try {
    await ElMessageBox.confirm(`确认重新执行任务 "${row.name}"？`, '重试', { type: 'warning' })
    await request.post(`/pinche/admin/batch-tasks/${row.id}/retry`)
    ElMessage.success('任务已重新加入队列')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onBatchCancel = async () => {
  if (!selection.value.length) return
  try {
    await ElMessageBox.confirm(`确认批量取消 ${selection.value.length} 个任务？`, '提示', { type: 'warning' })
    await Promise.all(selection.value.map(i =>
      request.put(`/pinche/admin/batch-tasks/${i.id}/cancel`)
    ))
    ElMessage.success('批量取消成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onBatchDelete = async () => {
  if (!selection.value.length) return
  try {
    await ElMessageBox.confirm(`确认批量删除 ${selection.value.length} 个任务记录？此操作不可撤销`, '提示', { type: 'warning' })
    await Promise.all(selection.value.map(i =>
      request.delete(`/pinche/admin/batch-tasks/${i.id}`)
    ))
    ElMessage.success('批量删除成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

// ===== 新建表单 =====
const formRef = ref(null)
const form = reactive({
  name: '', target_type: 'pinche', action: 'batch_audit_pass',
  target_status: null, target_ids_text: '', run_mode: 'async',
  on_failure: 'continue', description: ''
})
const rules = {
  name: [{ required: true, message: '请输入任务名称', trigger: 'blur' }],
  target_type: [{ required: true, message: '请选择目标对象类型', trigger: 'change' }],
  action: [{ required: true, message: '请选择操作类型', trigger: 'change' }]
}

const onTargetTypeChange = () => {
  form.target_ids_text = ''
}

const parseIds = (text) => {
  return String(text || '')
    .split(/[\s,，;；\n]+/)
    .map(s => s.trim())
    .filter(Boolean)
    .map(s => parseInt(s, 10))
    .filter(n => !isNaN(n) && n > 0)
}

const onLoadFromFilter = async () => {
  ElMessage.info('从筛选条件加载中...')
  try {
    // 按当前选中的目标类型查询前 100 条 ID
    const res = await request.get('/pinche/admin/batch-tasks/preview-ids', {
      params: { target_type: form.target_type, limit: 100 }
    })
    const ids = res.data?.ids || []
    form.target_ids_text = ids.join(',\n')
    ElMessage.success(`已加载 ${ids.length} 条 ID`)
  } catch (e) {
    ElMessage.warning('加载失败，请手动输入 ID')
  }
}

const onLoadAll = async () => {
  ElMessage.info('加载全部 ID 中...')
  try {
    const res = await request.get('/pinche/admin/batch-tasks/preview-ids', {
      params: { target_type: form.target_type, limit: 500 }
    })
    const ids = res.data?.ids || []
    form.target_ids_text = ids.join(',\n')
    ElMessage.success(`已加载 ${ids.length} 条 ID`)
  } catch (e) {
    ElMessage.warning('加载失败')
  }
}

const submitForm = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    const ids = parseIds(form.target_ids_text)
    if (!ids.length) {
      ElMessage.warning('请输入有效的目标 ID 列表')
      return
    }
    try {
      const payload = {
        name: form.name,
        target_type: form.target_type,
        action: form.action,
        target_ids: ids,
        run_mode: form.run_mode,
        on_failure: form.on_failure,
        description: form.description
      }
      if (form.action === 'batch_status') {
        if (form.target_status === null) {
          ElMessage.warning('请选择目标状态')
          return
        }
        payload.target_status = form.target_status
      }
      await request.post('/pinche/admin/batch-tasks', payload)
      ElMessage.success(`批量任务已提交，共 ${ids.length} 个目标`)
      activeTab.value = 'history'
      onFormReset()
      await loadList()
    } catch (e) { /* ignore */ }
  })
}

const onFormReset = () => {
  form.name = ''
  form.target_type = 'pinche'
  form.action = 'batch_audit_pass'
  form.target_status = null
  form.target_ids_text = ''
  form.run_mode = 'async'
  form.on_failure = 'continue'
  form.description = ''
}

onMounted(() => loadList())
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-card { display: flex; align-items: center; }
.stat-card :deep(.el-card__body) {
  display: flex; align-items: center; gap: 14px; padding: 16px; width: 100%;
}
.stat-icon {
  width: 44px; height: 44px; border-radius: 8px;
  display: flex; align-items: center; justify-content: center;
  color: #fff; flex-shrink: 0;
}
.stat-content { flex: 1; min-width: 0; }
.stat-value { font-size: 22px; font-weight: 600; color: #303133; }
.stat-label { font-size: 12px; color: #909399; margin-top: 2px; }

.batch-tabs { margin-bottom: 12px; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }

.toolbar { display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 8px; margin-bottom: 12px; }
.toolbar-left { display: flex; gap: 8px; flex-wrap: wrap; }
.toolbar-right { display: flex; gap: 8px; }

.text-success { color: #67c23a; }
.text-danger { color: #f56c6c; }
.text-muted { color: #909399; font-size: 12px; }
.form-tip { margin-top: 6px; font-size: 13px; }
.error-msg {
  white-space: pre-wrap;
  background: #fef0f0;
  color: #f56c6c;
  padding: 8px;
  border-radius: 4px;
  word-break: break-all;
}
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
