<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #409eff">
            <el-icon :size="22"><Warning /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.total }}</div>
            <div class="stat-label">总举报数</div>
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
            <div class="stat-label">待处理</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #f56c6c">
            <el-icon :size="22"><WarnTriangleFilled /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.highPriority }}</div>
            <div class="stat-label">高优先级</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #67c23a">
            <el-icon :size="22"><CircleCheckFilled /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.processed }}</div>
            <div class="stat-label">已处理</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <!-- 筛选区 -->
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input
            v-model="filters.keyword"
            placeholder="举报单号/举报人/被举报人"
            clearable
            style="width: 220px"
            :prefix-icon="Search"
            @keyup.enter="onSearch"
          />
        </el-form-item>
        <el-form-item label="举报类型">
          <el-select v-model="filters.report_type" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option label="虚假信息" value="fake_info" />
            <el-option label="诈骗" value="fraud" />
            <el-option label="骚扰" value="harassment" />
            <el-option label="违规拼车" value="illegal_pinche" />
            <el-option label="价格欺诈" value="price_fraud" />
            <el-option label="爽约" value="no_show" />
            <el-option label="车辆问题" value="vehicle_issue" />
            <el-option label="安全问题" value="safety_issue" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="目标对象">
          <el-select v-model="filters.target_type" placeholder="全部" clearable style="width: 130px" @change="onSearch">
            <el-option label="拼车信息" value="pinche" />
            <el-option label="用户" value="user" />
            <el-option label="车主" value="driver" />
            <el-option label="车辆" value="vehicle" />
            <el-option label="行程" value="trip" />
            <el-option label="评价" value="rating" />
          </el-select>
        </el-form-item>
        <el-form-item label="处理状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="待处理" :value="0" />
            <el-option label="处理中" :value="1" />
            <el-option label="已处理" :value="2" />
            <el-option label="已驳回" :value="3" />
            <el-option label="已撤销" :value="4" />
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

      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button :icon="Refresh" @click="loadList">刷新</el-button>
          <el-button type="success" :icon="Check" :disabled="!selection.length" @click="onBatchProcess">批量处理</el-button>
          <el-button type="info" :icon="Close" :disabled="!selection.length" @click="onBatchReject">批量驳回</el-button>
        </div>
        <div class="toolbar-right">
          <el-button type="primary" :icon="Plus" @click="onCreate">新建举报</el-button>
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
        <el-table-column prop="report_no" label="举报单号" width="170" />
        <el-table-column label="举报类型" width="120">
          <template #default="{ row }">
            <el-tag :type="typeTagType(row.report_type)" size="small">{{ typeText(row.report_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="举报人" width="140">
          <template #default="{ row }">
            <div>{{ row.reporter_name || `用户#${row.reporter_id}` }}</div>
            <div class="text-muted">{{ maskPhone(row.reporter_phone) }}</div>
          </template>
        </el-table-column>
        <el-table-column label="被举报对象" width="180">
          <template #default="{ row }">
            <el-tag size="small" effect="plain">{{ targetTypeText(row.target_type) }}</el-tag>
            <span style="margin-left: 4px">#{{ row.target_id }}</span>
          </template>
        </el-table-column>
        <el-table-column label="举报原因" min-width="220">
          <template #default="{ row }">
            <div class="reason-text">{{ row.reason }}</div>
          </template>
        </el-table-column>
        <el-table-column label="证据" width="80">
          <template #default="{ row }">
            <el-badge v-if="row.evidence_count" :value="row.evidence_count" type="primary">
              <el-icon><Picture /></el-icon>
            </el-badge>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="举报时间" width="160" prop="created_at" sortable>
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button
              v-if="row.status === 0 || row.status === 1"
              type="success"
              link
              size="small"
              @click="onProcess(row)"
            >处理</el-button>
            <el-button
              v-if="row.status === 0 || row.status === 1"
              type="warning"
              link
              size="small"
              @click="onReject(row)"
            >驳回</el-button>
            <el-button
              v-if="row.status === 0 || row.status === 1"
              type="danger"
              link
              size="small"
              @click="onPenalty(row)"
            >处罚</el-button>
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
    </div>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="举报详情" width="800px" destroy-on-close>
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="举报单号">{{ detail.report_no }}</el-descriptions-item>
        <el-descriptions-item label="举报类型">
          <el-tag :type="typeTagType(detail.report_type)" size="small">{{ typeText(detail.report_type) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTagType(detail.status)" size="small">{{ statusText(detail.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="举报人">{{ detail.reporter_name || `用户#${detail.reporter_id}` }}</el-descriptions-item>
        <el-descriptions-item label="举报人电话">{{ detail.reporter_phone || '-' }}</el-descriptions-item>
        <el-descriptions-item label="被举报对象类型">{{ targetTypeText(detail.target_type) }}</el-descriptions-item>
        <el-descriptions-item label="被举报对象 ID">#{{ detail.target_id }}</el-descriptions-item>
        <el-descriptions-item label="举报原因" :span="2">{{ detail.reason }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.description" label="详细描述" :span="2">{{ detail.description }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.evidence_urls" label="证据图片" :span="2">
          <div class="evidence-list">
            <el-image
              v-for="(url, idx) in parseEvidence(detail.evidence_urls)"
              :key="idx"
              :src="url"
              :preview-src-list="parseEvidence(detail.evidence_urls)"
              fit="cover"
              class="evidence-img"
            />
          </div>
        </el-descriptions-item>
        <el-descriptions-item label="举报时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.handled_at" label="处理时间">{{ formatTime(detail.handled_at) }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.handler_id" label="处理人">管理员 #{{ detail.handler_id }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.handle_result" label="处理结果">{{ detail.handle_result }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.penalty_type" label="处罚类型">
          <el-tag type="danger" size="small">{{ penaltyText(detail.penalty_type) }}</el-tag>
        </el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
        <el-button v-if="detail && (detail.status === 0 || detail.status === 1)" type="success" @click="onProcess(detail)">立即处理</el-button>
      </template>
    </el-dialog>

    <!-- 处理弹窗 -->
    <el-dialog v-model="processVisible" title="处理举报" width="600px" destroy-on-close>
      <el-form :model="processForm" label-width="110px">
        <el-form-item label="举报单号">
          <span>{{ processForm.report_no }}</span>
        </el-form-item>
        <el-form-item label="处理结果">
          <el-radio-group v-model="processForm.result">
            <el-radio value="valid">举报成立</el-radio>
            <el-radio value="invalid">举报不成立</el-radio>
            <el-radio value="partial">部分成立</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="processForm.result !== 'invalid'" label="处罚类型">
          <el-select v-model="processForm.penalty_type" placeholder="请选择处罚类型" style="width: 100%">
            <el-option label="警告" value="warning" />
            <el-option label="限制发布 3 天" value="limit3" />
            <el-option label="限制发布 7 天" value="limit7" />
            <el-option label="封禁 1 天" value="ban1d" />
            <el-option label="封禁 7 天" value="ban7d" />
            <el-option label="永久封禁" value="banForever" />
          </el-select>
        </el-form-item>
        <el-form-item label="处理说明" required>
          <el-input
            v-model="processForm.handle_note"
            type="textarea"
            :rows="5"
            placeholder="请描述处理过程、依据、结果等"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="processVisible = false">取消</el-button>
        <el-button type="primary" @click="submitProcess">提交</el-button>
      </template>
    </el-dialog>

    <!-- 新建举报弹窗 -->
    <el-dialog v-model="createVisible" title="新建举报" width="600px" destroy-on-close>
      <el-form :model="createForm" label-width="110px">
        <el-form-item label="被举报对象类型">
          <el-select v-model="createForm.target_type" placeholder="请选择" style="width: 100%">
            <el-option label="拼车信息" value="pinche" />
            <el-option label="用户" value="user" />
            <el-option label="车主" value="driver" />
            <el-option label="车辆" value="vehicle" />
            <el-option label="行程" value="trip" />
            <el-option label="评价" value="rating" />
          </el-select>
        </el-form-item>
        <el-form-item label="被举报对象 ID">
          <el-input-number v-model="createForm.target_id" :min="1" controls-position="right" style="width: 100%" />
        </el-form-item>
        <el-form-item label="举报类型">
          <el-select v-model="createForm.report_type" placeholder="请选择" style="width: 100%">
            <el-option label="虚假信息" value="fake_info" />
            <el-option label="诈骗" value="fraud" />
            <el-option label="骚扰" value="harassment" />
            <el-option label="违规拼车" value="illegal_pinche" />
            <el-option label="价格欺诈" value="price_fraud" />
            <el-option label="爽约" value="no_show" />
            <el-option label="车辆问题" value="vehicle_issue" />
            <el-option label="安全问题" value="safety_issue" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="举报原因">
          <el-input
            v-model="createForm.reason"
            type="textarea"
            :rows="4"
            placeholder="请描述举报原因"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" @click="submitCreate">提交</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/utils/request'
import {
  Refresh, RefreshLeft, Search, Plus, Check, Close,
  Warning, Clock, WarnTriangleFilled, CircleCheckFilled, Picture
} from '@element-plus/icons-vue'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const selection = ref([])

const stats = reactive({ total: 0, pending: 0, highPriority: 0, processed: 0 })

const filters = reactive({
  keyword: '', report_type: '', target_type: '', status: null, dateRange: null
})

const statusText = (s) => ({ 0: '待处理', 1: '处理中', 2: '已处理', 3: '已驳回', 4: '已撤销' }[s] || '-')
const statusTagType = (s) => ({ 0: 'warning', 1: '', 2: 'success', 3: 'info', 4: 'info' }[s] || 'info')
const typeText = (t) => ({
  fake_info: '虚假信息', fraud: '诈骗', harassment: '骚扰',
  illegal_pinche: '违规拼车', price_fraud: '价格欺诈', no_show: '爽约',
  vehicle_issue: '车辆问题', safety_issue: '安全问题', other: '其他'
}[t] || '-')
const typeTagType = (t) => ({
  fake_info: 'warning', fraud: 'danger', harassment: 'danger',
  illegal_pinche: 'warning', price_fraud: 'warning', no_show: 'info',
  vehicle_issue: 'info', safety_issue: 'danger', other: 'info'
}[t] || 'info')
const targetTypeText = (t) => ({
  pinche: '拼车信息', user: '用户', driver: '车主',
  vehicle: '车辆', trip: '行程', rating: '评价'
}[t] || t || '-')
const penaltyText = (p) => ({
  warning: '警告', limit3: '限制发布 3 天', limit7: '限制发布 7 天',
  ban1d: '封禁 1 天', ban7d: '封禁 7 天', banForever: '永久封禁'
}[p] || p || '-')

const maskPhone = (p) => p ? p.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2') : '-'
const parseEvidence = (urls) => {
  if (!urls) return []
  if (Array.isArray(urls)) return urls
  try { return JSON.parse(urls) } catch (e) { return String(urls).split(',') }
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.keyword = ''; filters.report_type = ''
  filters.target_type = ''; filters.status = null; filters.dateRange = null
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
      report_type: filters.report_type || undefined,
      target_type: filters.target_type || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await request.get('/pinche/admin/reports', { params })
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
    if (data.stats) {
      stats.total = data.stats.total || 0
      stats.pending = data.stats.pending || 0
      stats.highPriority = data.stats.high_priority || 0
      stats.processed = data.stats.processed || 0
    } else {
      stats.total = total.value
      stats.pending = list.value.filter(i => i.status === 0 || i.status === 1).length
      stats.highPriority = list.value.filter(i => ['fraud', 'safety_issue'].includes(i.report_type)).length
      stats.processed = list.value.filter(i => i.status === 2).length
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
}

const processVisible = ref(false)
const processForm = reactive({
  id: 0, report_no: '', result: 'valid', penalty_type: 'warning', handle_note: ''
})

const onProcess = (row) => {
  processForm.id = row.id
  processForm.report_no = row.report_no || `#${row.id}`
  processForm.result = 'valid'
  processForm.penalty_type = 'warning'
  processForm.handle_note = ''
  processVisible.value = true
}

const submitProcess = async () => {
  if (!processForm.handle_note?.trim()) {
    ElMessage.warning('请填写处理说明')
    return
  }
  try {
    await request.put(`/pinche/admin/reports/${processForm.id}/process`, {
      handle_note: processForm.handle_note,
      result: processForm.result,
      penalty_type: processForm.result === 'invalid' ? '' : processForm.penalty_type,
      status: processForm.result === 'invalid' ? 3 : 2
    })
    ElMessage.success('处理成功')
    processVisible.value = false
    await loadList()
  } catch (e) { /* ignore */ }
}

const onReject = async (row) => {
  try {
    const { value } = await ElMessageBox.prompt('请输入驳回原因', '驳回举报', {
      inputType: 'textarea',
      inputPlaceholder: '驳回原因',
      inputValidator: (v) => !!v || '请输入驳回原因'
    })
    await request.put(`/pinche/admin/reports/${row.id}/process`, {
      handle_note: value,
      result: 'invalid',
      status: 3
    })
    ElMessage.success('举报已驳回')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onPenalty = async (row) => {
  try {
    const { value } = await ElMessageBox.prompt(
      `确认对被举报对象进行处罚？请选择处罚类型并说明原因`,
      '处罚',
      {
        inputType: 'textarea',
        inputPlaceholder: '处罚原因说明',
        inputValidator: (v) => !!v || '请输入处罚原因'
      }
    )
    // 拆分处罚类型和原因
    await request.put(`/pinche/admin/reports/${row.id}/process`, {
      handle_note: value,
      result: 'valid',
      penalty_type: 'ban1d',
      status: 2
    })
    ElMessage.success('处罚已执行')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onBatchProcess = async () => {
  if (!selection.value.length) return
  try {
    const { value } = await ElMessageBox.prompt(
      `确认批量处理 ${selection.value.length} 条举报？请填写统一处理说明`,
      '批量处理',
      { inputType: 'textarea', inputPlaceholder: '处理说明', inputValidator: (v) => !!v || '请输入处理说明' }
    )
    await Promise.all(selection.value.map(i =>
      request.put(`/pinche/admin/reports/${i.id}/process`, {
        handle_note: value, result: 'valid', penalty_type: 'warning', status: 2
      })
    ))
    ElMessage.success(`已批量处理 ${selection.value.length} 条举报`)
    await loadList()
  } catch (e) { /* cancel */ }
}

const onBatchReject = async () => {
  if (!selection.value.length) return
  try {
    await ElMessageBox.confirm(`确认批量驳回 ${selection.value.length} 条举报？`, '批量驳回', { type: 'warning' })
    await Promise.all(selection.value.map(i =>
      request.put(`/pinche/admin/reports/${i.id}/process`, {
        handle_note: '批量驳回', result: 'invalid', status: 3
      })
    ))
    ElMessage.success(`已批量驳回 ${selection.value.length} 条举报`)
    await loadList()
  } catch (e) { /* cancel */ }
}

const createVisible = ref(false)
const createForm = reactive({
  target_type: 'pinche', target_id: 0, report_type: 'other', reason: ''
})

const onCreate = () => {
  createForm.target_type = 'pinche'
  createForm.target_id = 0
  createForm.report_type = 'other'
  createForm.reason = ''
  createVisible.value = true
}

const submitCreate = async () => {
  if (!createForm.target_id || !createForm.reason?.trim()) {
    ElMessage.warning('请填写完整信息')
    return
  }
  try {
    await request.post('/pinche/admin/reports', { ...createForm })
    ElMessage.success('举报已创建')
    createVisible.value = false
    await loadList()
  } catch (e) { /* ignore */ }
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

.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }

.toolbar { display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 8px; margin-bottom: 12px; }
.toolbar-left { display: flex; gap: 8px; flex-wrap: wrap; }
.toolbar-right { display: flex; gap: 8px; }

.text-muted { color: #909399; font-size: 12px; }
.reason-text {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  word-break: break-all;
}
.evidence-list { display: flex; gap: 8px; flex-wrap: wrap; }
.evidence-img { width: 80px; height: 80px; border-radius: 4px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
