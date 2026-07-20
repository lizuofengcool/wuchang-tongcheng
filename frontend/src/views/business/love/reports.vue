<template>
  <div class="app-container">
    <!-- SLA 监控卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">总举报数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.pending }}</div><div class="stat-label">待处理</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-danger">{{ stats.highRisk }}</div><div class="stat-label">高风险</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.handled }}</div><div class="stat-label">已处理</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="举报类型">
          <el-select v-model="filters.report_type" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option v-for="(label, val) in typeMap" :key="val" :label="label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option v-for="(label, val) in statusMap" :key="val" :label="label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item label="目标会员ID">
          <el-input v-model="filters.target_user_id" placeholder="目标会员ID" clearable style="width: 140px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="风险等级">
          <el-select v-model="filters.risk_level" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="低" value="low" />
            <el-option label="中" value="medium" />
            <el-option label="高" value="high" />
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
        <el-table-column prop="report_no" label="举报单号" width="180" fixed="left" />
        <el-table-column label="举报人" width="140">
          <template #default="{ row }">
            <div>{{ row.reporter_name || `#${row.reporter_id}` }}</div>
          </template>
        </el-table-column>
        <el-table-column label="被举报人" width="140">
          <template #default="{ row }">
            <div>{{ row.reported_name || `#${row.reported_user_id}` }}</div>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ typeMap[row.report_type] || row.report_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="原因" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ row.reason }}</template>
        </el-table-column>
        <el-table-column label="证据" width="80">
          <template #default="{ row }">
            <el-badge v-if="row.evidence_images && row.evidence_images.length" :value="row.evidence_images.length" type="primary">
              <el-icon><Picture /></el-icon>
            </el-badge>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="风险等级" width="100">
          <template #default="{ row }">
            <el-tag :type="riskTagType(row.risk_level)" size="small">{{ riskText(row.risk_level) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ row.status_text || statusMap[row.status] || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="举报时间" width="160" prop="created_at" sortable="custom">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status === 'pending'" type="warning" link size="small" @click="openProcess(row)">处理</el-button>
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
    <el-dialog v-model="detailVisible" title="举报详情" width="700px">
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="举报单号">{{ detail.report_no }}</el-descriptions-item>
        <el-descriptions-item label="举报人">{{ detail.reporter_name || `#${detail.reporter_id}` }}</el-descriptions-item>
        <el-descriptions-item label="被举报人">{{ detail.reported_name || `#${detail.reported_user_id}` }}</el-descriptions-item>
        <el-descriptions-item label="被举报ID">{{ detail.reported_user_id }}</el-descriptions-item>
        <el-descriptions-item label="举报类型">
          <el-tag size="small">{{ typeMap[detail.report_type] || detail.report_type }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTagType(detail.status)" size="small">{{ detail.status_text || statusMap[detail.status] }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="风险等级">
          <el-tag :type="riskTagType(detail.risk_level)" size="small">{{ riskText(detail.risk_level) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="处理人">{{ detail.handler_name || `#${detail.handler_id}` || '-' }}</el-descriptions-item>
        <el-descriptions-item label="原因" :span="2">{{ detail.reason }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.description" label="详细描述" :span="2">{{ detail.description }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.handle_result" label="处理结果" :span="2">{{ detail.handle_result }}</el-descriptions-item>
        <el-descriptions-item label="处罚类型">{{ penaltyMap[detail.penalty_type] || detail.penalty_type || '-' }}</el-descriptions-item>
        <el-descriptions-item label="处理时间">{{ formatTime(detail.handled_at) }}</el-descriptions-item>
        <el-descriptions-item label="举报时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
        <el-descriptions-item label="证据" :span="2">
          <div v-if="detail.evidence_images && detail.evidence_images.length" class="images-grid">
            <el-image v-for="(img, idx) in detail.evidence_images" :key="idx" :src="img" fit="cover" class="image-item" :preview-src-list="detail.evidence_images" :initial-index="idx" preview-teleported />
          </div>
          <span v-else class="text-muted">无</span>
        </el-descriptions-item>
      </el-descriptions>
      <template #footer><el-button @click="detailVisible = false">关闭</el-button></template>
    </el-dialog>

    <!-- 处理弹窗 -->
    <el-dialog v-model="processVisible" title="处理举报" width="500px">
      <el-form :model="processForm" label-width="100px">
        <el-form-item label="处理结果">
          <el-radio-group v-model="processForm.status">
            <el-radio value="verified">核实警告</el-radio>
            <el-radio value="banned">封号</el-radio>
            <el-radio value="rejected">驳回</el-radio>
            <el-radio value="transferred">转交</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="处罚类型">
          <el-select v-model="processForm.penalty_type" placeholder="请选择" clearable style="width: 100%">
            <el-option label="警告" value="warning" />
            <el-option label="限制" value="limit" />
            <el-option label="封禁1天" value="ban1d" />
            <el-option label="封禁7天" value="ban7d" />
            <el-option label="永久封禁" value="banForever" />
          </el-select>
        </el-form-item>
        <el-form-item label="处理说明">
          <el-input v-model="processForm.handle_result" type="textarea" :rows="3" placeholder="处理说明" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="processVisible = false">取消</el-button>
        <el-button type="primary" :loading="processLoading" @click="onProcess">确认处理</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, RefreshLeft, Search, Picture } from '@element-plus/icons-vue'
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
  report_type: '', status: '', target_user_id: '', risk_level: '', dateRange: null
})

const typeMap = {
  porn: '色情低俗', scam: '欺诈', fake: '虚假信息', harassment: '骚扰',
  spam: '垃圾信息', infringement: '侵权', prohibited: '违禁品', other: '其他'
}
const statusMap = {
  pending: '待处理', verified: '已核实警告', banned: '已封号',
  rejected: '已驳回', transferred: '已转交'
}
const penaltyMap = { warning: '警告', limit: '限制', ban1d: '封禁1天', ban7d: '封禁7天', banForever: '永久封禁' }

const statusTagType = (s) => ({
  pending: 'warning', verified: 'info', banned: 'danger',
  rejected: 'success', transferred: 'info'
}[s] || 'info')

const riskText = (r) => ({ low: '低', medium: '中', high: '高' }[r] || '-')
const riskTagType = (r) => ({ low: 'info', medium: 'warning', high: 'danger' }[r] || 'info')

const stats = computed(() => {
  const total = list.value.length
  const pending = list.value.filter((r) => r.status === 'pending').length
  const highRisk = list.value.filter((r) => r.risk_level === 'high').length
  const handled = list.value.filter((r) => r.status !== 'pending').length
  return { total, pending, highRisk, handled }
})

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.report_type = ''; filters.status = ''; filters.target_user_id = ''
  filters.risk_level = ''; filters.dateRange = null
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
      report_type: filters.report_type || undefined,
      status: filters.status || undefined,
      target_user_id: filters.target_user_id || undefined,
      risk_level: filters.risk_level || undefined,
      sort: sortField.value,
      order: sortOrder.value
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await request.get('/love/admin/reports', { params })
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
const openDetail = (row) => { detail.value = row; detailVisible.value = true }

const processVisible = ref(false)
const processLoading = ref(false)
const processForm = reactive({ id: null, status: 'verified', penalty_type: '', handle_result: '' })

const openProcess = (row) => {
  processForm.id = row.id
  processForm.status = 'verified'
  processForm.penalty_type = ''
  processForm.handle_result = ''
  processVisible.value = true
}

const onProcess = async () => {
  processLoading.value = true
  try {
    await request.put(`/love/admin/reports/${processForm.id}/handle`, {
      status: processForm.status,
      penalty_type: processForm.penalty_type || undefined,
      handle_result: processForm.handle_result
    })
    ElMessage.success('处理成功')
    processVisible.value = false
    await loadList()
  } catch (e) {
    // 失败已提示
  } finally {
    processLoading.value = false
  }
}

onMounted(() => loadList())
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-item { text-align: center; padding: 16px; }
.stat-value { font-size: 24px; font-weight: 600; color: #409eff; }
.text-warning { color: #e6a23c; }
.text-danger { color: #f56c6c; }
.text-success { color: #67c23a; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: flex-start; margin-bottom: 12px; }
.text-muted { color: #909399; }
.images-grid { display: flex; flex-wrap: wrap; gap: 8px; }
.image-item { width: 100px; height: 100px; border-radius: 4px; border: 1px solid #ebeef5; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
