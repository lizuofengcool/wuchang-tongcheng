<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #409eff">
            <el-icon :size="22"><Briefcase /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.total }}</div>
            <div class="stat-label">总职位数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #67c23a">
            <el-icon :size="22"><CirclePlus /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.todayNew }}</div>
            <div class="stat-label">今日新增</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #e6a23c">
            <el-icon :size="22"><Clock /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.pendingAudit }}</div>
            <div class="stat-label">待审核</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #13c2c2">
            <el-icon :size="22"><Promotion /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.published }}</div>
            <div class="stat-label">已发布</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #722ed1">
            <el-icon :size="22"><TrendCharts /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.totalApplications }}</div>
            <div class="stat-label">总投递数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #f56c6c">
            <el-icon :size="22"><Warning /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.violation }}</div>
            <div class="stat-label">违规数</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 高级筛选区 -->
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input
            v-model="filters.keyword"
            placeholder="职位/公司/发布者"
            clearable
            style="width: 200px"
            :prefix-icon="Search"
            @keyup.enter="onSearch"
          />
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="filters.category_id" placeholder="全部" clearable style="width: 160px" filterable @change="onSearch">
            <el-option v-for="c in categoryOptions" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="学历">
          <el-select v-model="filters.education" placeholder="全部" clearable style="width: 130px" @change="onSearch">
            <el-option label="不限" value="unlimited" />
            <el-option label="大专" value="college" />
            <el-option label="本科" value="bachelor" />
            <el-option label="硕士" value="master" />
            <el-option label="博士" value="doctor" />
          </el-select>
        </el-form-item>
        <el-form-item label="经验">
          <el-select v-model="filters.experience" placeholder="全部" clearable style="width: 130px" @change="onSearch">
            <el-option label="应届" value="fresh" />
            <el-option label="1-3年" value="1-3" />
            <el-option label="3-5年" value="3-5" />
            <el-option label="5-10年" value="5-10" />
            <el-option label="10年以上" value="10+" />
          </el-select>
        </el-form-item>
        <el-form-item label="薪资区间">
          <el-input-number v-model="filters.min_salary" :min="0" :controls="false" placeholder="最低" style="width: 100px" />
          <span style="margin: 0 4px">-</span>
          <el-input-number v-model="filters.max_salary" :min="0" :controls="false" placeholder="最高" style="width: 100px" />
        </el-form-item>
        <el-form-item label="审核状态">
          <el-select v-model="filters.audit_status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="待审核" :value="0" />
            <el-option label="已通过" :value="1" />
            <el-option label="已拒绝" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="职位状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="草稿" :value="0" />
            <el-option label="招聘中" :value="1" />
            <el-option label="已停招" :value="2" />
            <el-option label="已下架" :value="3" />
            <el-option label="已过期" :value="4" />
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
          <el-button type="success" :icon="Check" :disabled="!selection.length" @click="onBatchAudit(1)">批量通过</el-button>
          <el-button type="warning" :icon="Bottom" :disabled="!selection.length" @click="onBatchStatus(3)">批量下架</el-button>
          <el-button type="danger" :icon="Delete" :disabled="!selection.length" @click="onBatchDelete">批量删除</el-button>
        </div>
        <div class="toolbar-right">
          <el-button type="primary" :icon="Plus" @click="onCreate">新建职位</el-button>
        </div>
      </div>

      <!-- 表格 -->
      <el-table
        v-loading="loading"
        :data="list"
        border
        stripe
        @selection-change="onSelectionChange"
        @sort-change="onSortChange"
      >
        <el-table-column type="selection" width="44" fixed="left" />
        <el-table-column prop="id" label="ID" width="70" sortable="custom" fixed="left" />
        <el-table-column label="职位/薪资" min-width="220">
          <template #default="{ row }">
            <div class="title-cell">
              <div class="title-text">
                <el-link type="primary" :underline="'never'" @click="openDetail(row)">{{ row.title }}</el-link>
                <el-tag v-if="row.is_urgent" type="danger" size="small" effect="dark">急</el-tag>
              </div>
              <div class="title-desc">
                <span class="salary">{{ formatSalary(row) }}</span>
                <span class="text-muted">·{{ row.city || '不限' }}</span>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="公司" width="160">
          <template #default="{ row }">
            <div>{{ row.company_name || `公司#${row.company_id}` }}</div>
            <div class="text-muted text-xs">{{ row.industry || '-' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="学历" width="80">
          <template #default="{ row }">{{ educationText(row.education) }}</template>
        </el-table-column>
        <el-table-column label="经验" width="90">
          <template #default="{ row }">{{ row.experience || '-' }}</template>
        </el-table-column>
        <el-table-column label="发布者" width="150">
          <template #default="{ row }">
            <div class="publisher">
              <div class="publisher-name">{{ row.user_name || `用户#${row.user_id}` }}</div>
              <div class="publisher-phone">{{ maskPhone(row.user_phone) }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="浏览" width="70" prop="view_count" sortable="custom" />
        <el-table-column label="投递" width="70" prop="application_count" sortable="custom" />
        <el-table-column label="收藏" width="70" prop="fav_count" sortable="custom" />
        <el-table-column label="审核" width="90">
          <template #default="{ row }">
            <el-tag :type="auditTagType(row.audit_status)" size="small">{{ auditText(row.audit_status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small" effect="plain">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="发布时间" width="160" prop="created_at" sortable="custom">
          <template #default="{ row }">{{ formatTime(row.published_at || row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button
              v-if="row.audit_status === 0 || row.audit_status === 2"
              type="success"
              link
              size="small"
              @click="handleAudit(row, 1)"
            >通过</el-button>
            <el-button
              v-if="row.audit_status === 0 || row.audit_status === 1"
              type="danger"
              link
              size="small"
              @click="handleAudit(row, 2)"
            >拒绝</el-button>
            <el-dropdown trigger="click" @command="(cmd) => handleStatusCommand(row, cmd)">
              <el-button type="warning" link size="small">
                更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item :command="1" :disabled="row.status === 1">招聘中</el-dropdown-item>
                  <el-dropdown-item :command="2" :disabled="row.status === 2">停招</el-dropdown-item>
                  <el-dropdown-item :command="3" :disabled="row.status === 3">下架</el-dropdown-item>
                  <el-dropdown-item :command="4" :disabled="row.status === 4">设为过期</el-dropdown-item>
                  <el-dropdown-item :command="'delete'" divided>删除</el-dropdown-item>
                  <el-dropdown-item :command="'report'">举报处理</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
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
    <el-dialog v-model="detailVisible" title="职位详情" width="900px" @close="onDetailClose" destroy-on-close>
      <div v-loading="detailLoading">
        <el-tabs v-if="detail" v-model="detailTab">
          <el-tab-pane label="基本信息" name="basic">
            <el-descriptions :column="3" border>
              <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
              <el-descriptions-item label="职位">{{ detail.title }}</el-descriptions-item>
              <el-descriptions-item label="公司">{{ detail.company_name || `公司#${detail.company_id}` }}</el-descriptions-item>
              <el-descriptions-item label="薪资">{{ formatSalary(detail) }}</el-descriptions-item>
              <el-descriptions-item label="学历">{{ educationText(detail.education) }}</el-descriptions-item>
              <el-descriptions-item label="经验">{{ detail.experience || '-' }}</el-descriptions-item>
              <el-descriptions-item label="城市">{{ detail.city || '不限' }}</el-descriptions-item>
              <el-descriptions-item label="地址">{{ detail.address || '-' }}</el-descriptions-item>
              <el-descriptions-item label="招聘人数">{{ detail.headcount || '-' }}</el-descriptions-item>
              <el-descriptions-item label="审核状态">
                <el-tag :type="auditTagType(detail.audit_status)" size="small">{{ auditText(detail.audit_status) }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="职位状态">
                <el-tag :type="statusTagType(detail.status)" size="small" effect="plain">{{ statusText(detail.status) }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="加急">
                <el-tag v-if="detail.is_urgent" type="danger" size="small">急招</el-tag>
                <span v-else>否</span>
              </el-descriptions-item>
              <el-descriptions-item label="福利" :span="3">{{ Array.isArray(detail.benefits) ? detail.benefits.join('、') : (detail.benefits || '-') }}</el-descriptions-item>
              <el-descriptions-item label="技能要求" :span="3">{{ Array.isArray(detail.skills) ? detail.skills.join('、') : (detail.skills || '-') }}</el-descriptions-item>
              <el-descriptions-item v-if="detail.description" label="职位描述" :span="3">
                <div class="content-box">{{ detail.description }}</div>
              </el-descriptions-item>
              <el-descriptions-item v-if="detail.audit_reason" label="审核原因" :span="3">{{ detail.audit_reason }}</el-descriptions-item>
              <el-descriptions-item label="发布者">{{ detail.user_name || `用户#${detail.user_id}` }}</el-descriptions-item>
              <el-descriptions-item label="发布者ID">{{ detail.user_id }}</el-descriptions-item>
              <el-descriptions-item label="地区ID">{{ detail.region_id || '-' }}</el-descriptions-item>
              <el-descriptions-item label="浏览量">{{ detail.view_count || 0 }}</el-descriptions-item>
              <el-descriptions-item label="投递数">{{ detail.application_count || 0 }}</el-descriptions-item>
              <el-descriptions-item label="收藏量">{{ detail.fav_count || 0 }}</el-descriptions-item>
              <el-descriptions-item label="发布时间">{{ formatTime(detail.published_at) }}</el-descriptions-item>
              <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
              <el-descriptions-item label="过期时间">{{ formatTime(detail.expiry_time) }}</el-descriptions-item>
            </el-descriptions>
          </el-tab-pane>

          <el-tab-pane :label="`举报历史 (${reports.length})`" name="reports">
            <el-table :data="reports" border size="small">
              <el-table-column prop="report_no" label="举报单号" width="160" />
              <el-table-column prop="reason" label="原因" min-width="180" show-overflow-tooltip />
              <el-table-column label="状态" width="100">
                <template #default="{ row }">
                  <el-tag :type="row.status === 0 ? 'warning' : 'info'" size="small">{{ row.status === 0 ? '待处理' : '已处理' }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column label="时间" width="160">
                <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
              </el-table-column>
            </el-table>
            <div v-if="!reports.length" class="empty-text">暂无举报记录</div>
          </el-tab-pane>

          <el-tab-pane label="操作日志" name="logs">
            <div class="empty-text">操作日志功能开发中</div>
          </el-tab-pane>
        </el-tabs>
      </div>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Refresh, RefreshLeft, Search, ArrowDown, Check, Bottom, Delete, Plus,
  Briefcase, CirclePlus, Clock, Promotion, Warning, TrendCharts
} from '@element-plus/icons-vue'
import {
  adminListJobs, adminGetJob, auditJob, adminUpdateJobStatus,
  deleteJob, getOverviewStats
} from '@/api/job'
import { listCategories } from '@/api/job'
import { adminListReports } from '@/api/job'
import { formatTime } from '@/utils/format'

const router = useRouter()

// ===== 列表 =====
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const selection = ref([])
const sortField = ref('created_at')
const sortOrder = ref('descending')

// ===== 统计卡片 =====
const stats = reactive({
  total: 0, todayNew: 0, pendingAudit: 0, published: 0, totalApplications: 0, violation: 0
})

const loadStats = async () => {
  try {
    const res = await getOverviewStats()
    const d = res.data || {}
    stats.total = d.total_jobs || 0
    stats.todayNew = d.today_new || 0
    stats.pendingAudit = d.pending_audit || 0
    stats.published = d.published || 0
    stats.totalApplications = d.total_applications || 0
    stats.violation = d.violation || 0
  } catch (e) {
    // 接口失败时统计保持默认值 0
  }
}

// ===== 筛选 =====
const filters = reactive({
  keyword: '', category_id: null, education: '', experience: '',
  min_salary: undefined, max_salary: undefined,
  audit_status: null, status: null, dateRange: null
})

const onSearch = () => {
  page.value = 1
  loadList()
}

const onReset = () => {
  Object.assign(filters, {
    keyword: '', category_id: null, education: '', experience: '',
    min_salary: undefined, max_salary: undefined,
    audit_status: null, status: null, dateRange: null
  })
  page.value = 1
  loadList()
}

const onSortChange = ({ prop, order }) => {
  sortField.value = prop || 'created_at'
  sortOrder.value = order || 'descending'
  loadList()
}

const onSelectionChange = (rows) => {
  selection.value = rows
}

// ===== 分类 =====
const categoryOptions = ref([])
const loadCategories = async () => {
  try {
    const res = await listCategories({ page: 1, page_size: 100 })
    const data = res.data || {}
    categoryOptions.value = data.list || data || []
  } catch (e) {
    categoryOptions.value = []
  }
}

// ===== 状态格式化 =====
const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')
const statusText = (s) => ({ 0: '草稿', 1: '招聘中', 2: '已停招', 3: '已下架', 4: '已过期' }[s] || '-')
const statusTagType = (s) => ({ 0: 'info', 1: 'success', 2: 'warning', 3: 'warning', 4: 'danger' }[s] || 'info')
const educationText = (e) => ({ unlimited: '不限', college: '大专', bachelor: '本科', master: '硕士', doctor: '博士' }[e] || '-')

const formatSalary = (row) => {
  const min = Number(row.salary_min || 0)
  const max = Number(row.salary_max || 0)
  if (!min && !max) return '面议'
  if (min && max) return `${min}-${max}K`
  return min ? `${min}K起` : `${max}K以内`
}

const maskPhone = (p) => {
  if (!p) return '-'
  const s = String(p)
  if (s.length < 7) return s
  return s.slice(0, 3) + '****' + s.slice(-4)
}

// ===== 查询 =====
const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      keyword: filters.keyword.trim() || undefined,
      category_id: filters.category_id || undefined,
      education: filters.education || undefined,
      experience: filters.experience || undefined,
      min_salary: filters.min_salary || undefined,
      max_salary: filters.max_salary || undefined,
      audit_status: filters.audit_status === null || filters.audit_status === '' ? undefined : filters.audit_status,
      status: filters.status === null || filters.status === '' ? undefined : filters.status,
      sort: sortField.value,
      order: sortOrder.value
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await adminListJobs(params)
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
    if (data.stats) {
      Object.assign(stats, data.stats)
    }
  } catch (e) {
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

// ===== 详情弹窗 =====
const detailVisible = ref(false)
const detailLoading = ref(false)
const detail = ref(null)
const reports = ref([])
const detailTab = ref('basic')

const openDetail = async (row) => {
  detail.value = row
  reports.value = []
  detailTab.value = 'basic'
  detailVisible.value = true
  detailLoading.value = true
  try {
    const [dRes, rRes] = await Promise.all([
      adminGetJob(row.id),
      adminListReports({ target_id: row.id, page: 1, page_size: 20 }).catch(() => ({ data: { list: [] } }))
    ])
    if (dRes.data) detail.value = dRes.data
    reports.value = rRes.data?.list || []
  } catch (e) {
    // 接口失败保持现状
  } finally {
    detailLoading.value = false
  }
}

const onDetailClose = () => {
  detail.value = null
  reports.value = []
}

// ===== 审核 =====
const handleAudit = async (row, auditStatus) => {
  const action = auditStatus === 1 ? '通过' : '拒绝'
  try {
    if (auditStatus === 2) {
      const { value } = await ElMessageBox.prompt('请输入拒绝原因（可选）', '拒绝审核', {
        confirmButtonText: '确定拒绝',
        cancelButtonText: '取消',
        inputType: 'textarea',
        inputPlaceholder: '拒绝原因（可不填）'
      })
      await auditJob(row.id, { audit_status: auditStatus, audit_reason: value || '' })
    } else {
      await ElMessageBox.confirm(`确定${action}职位 "${row.title}" 的审核吗？`, '提示', { type: 'warning' })
      await auditJob(row.id, { audit_status: auditStatus })
    }
    ElMessage.success(`已${action}`)
    await loadList()
    if (detailVisible.value && detail.value?.id === row.id) {
      const res = await adminGetJob(row.id)
      if (res.data) detail.value = res.data
    }
  } catch (e) {
    // 取消
  }
}

// ===== 状态变更 =====
const handleStatusCommand = async (row, cmd) => {
  try {
    if (cmd === 'delete') {
      await ElMessageBox.confirm(`确定删除职位 "${row.title}" 吗？删除后不可恢复！`, '危险操作', {
        type: 'error',
        confirmButtonText: '确认删除',
        cancelButtonText: '取消'
      })
      await deleteJob(row.id)
      ElMessage.success('已删除')
      await loadList()
      return
    }
    if (cmd === 'report') {
      router.push('/business/job/reports?job_id=' + row.id)
      return
    }
    const label = statusText(cmd)
    await ElMessageBox.confirm(`确定将职位 "${row.title}" 设为「${label}」吗？`, '提示', { type: 'warning' })
    await adminUpdateJobStatus(row.id, cmd)
    ElMessage.success('状态更新成功')
    await loadList()
    if (detailVisible.value && detail.value?.id === row.id) {
      const res = await adminGetJob(row.id)
      if (res.data) detail.value = res.data
    }
  } catch (e) {
    // 取消
  }
}

// ===== 批量操作 =====
const onBatchAudit = async (auditStatus) => {
  try {
    await ElMessageBox.confirm(`确认批量审核通过 ${selection.value.length} 个职位？`, '批量审核', { type: 'warning' })
    // job 模块后端没有批量审核接口，循环调用单条
    await Promise.all(selection.value.map((r) => auditJob(r.id, { audit_status: auditStatus })))
    ElMessage.success('批量审核完成')
    await loadList()
  } catch (e) {
    // 取消
  }
}

const onBatchStatus = async (status) => {
  try {
    const label = statusText(status)
    await ElMessageBox.confirm(`确认批量将 ${selection.value.length} 个职位设为「${label}」？`, '批量状态变更', { type: 'warning' })
    await Promise.all(selection.value.map((r) => adminUpdateJobStatus(r.id, status)))
    ElMessage.success('批量操作完成')
    await loadList()
  } catch (e) {
    // 取消
  }
}

const onBatchDelete = async () => {
  try {
    await ElMessageBox.confirm(`确认批量删除 ${selection.value.length} 个职位？删除后不可恢复！`, '危险操作', {
      type: 'error',
      confirmButtonText: '确认删除',
      cancelButtonText: '取消'
    })
    await Promise.all(selection.value.map((r) => deleteJob(r.id)))
    ElMessage.success('批量删除完成')
    await loadList()
  } catch (e) {
    // 取消
  }
}

// ===== 新建 =====
const onCreate = () => {
  ElMessage.info('新建职位功能开发中，请使用 C 端发布')
}

onMounted(async () => {
  await Promise.all([loadCategories(), loadStats()])
  await loadList()
})
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

.filter-form {
  margin-bottom: 12px;
  padding: 12px 16px;
  background: #fafafa;
  border-radius: 4px;
}
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }

.toolbar {
  display: flex; justify-content: space-between; align-items: center;
  flex-wrap: wrap; gap: 8px; margin-bottom: 12px;
}
.toolbar-left { display: flex; gap: 8px; flex-wrap: wrap; }
.toolbar-right { display: flex; gap: 8px; }

.title-cell { display: flex; flex-direction: column; gap: 4px; min-width: 0; }
.title-text {
  font-weight: 500; color: #303133;
  display: flex; align-items: center; gap: 6px;
}
.title-text .el-link { max-width: 100%; }
.title-desc {
  font-size: 12px; color: #909399;
  display: flex; align-items: center; gap: 6px;
}
.salary { color: #f56c6c; font-weight: 600; }
.text-muted { color: #909399; }
.text-xs { font-size: 12px; }
.publisher { font-size: 13px; }
.publisher-name { color: #303133; }
.publisher-phone { font-size: 12px; color: #909399; margin-top: 2px; }

.content-box {
  white-space: pre-wrap; word-break: break-all;
  max-height: 200px; overflow-y: auto;
}
.empty-text { color: #909399; text-align: center; padding: 32px 0; }
.pagination-wrap {
  margin-top: 16px; display: flex; justify-content: flex-end;
}

@media (max-width: 1200px) {
  .filter-form :deep(.el-form-item) { margin-right: 8px; }
}
</style>
