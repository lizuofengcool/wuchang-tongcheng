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
            <div class="stat-label">总岗位数</div>
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
            <el-icon :size="22"><User /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.applied }}</div>
            <div class="stat-label">已报名</div>
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
            placeholder="标题/公司/联系人"
            clearable
            style="width: 220px"
            :prefix-icon="Search"
            @keyup.enter="onSearch"
          />
        </el-form-item>
        <el-form-item label="岗位类型">
          <el-select v-model="filters.linggong_type" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option v-for="(label, val) in typeMap" :key="val" :label="label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item label="发布者">
          <el-select v-model="filters.publisher_type" placeholder="全部" clearable style="width: 130px" @change="onSearch">
            <el-option v-for="(label, val) in publisherMap" :key="val" :label="label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item label="计费">
          <el-select v-model="filters.billing_type" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option v-for="(label, val) in billingMap" :key="val" :label="label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item label="结算">
          <el-select v-model="filters.settlement" placeholder="全部" clearable style="width: 110px" @change="onSearch">
            <el-option v-for="(label, val) in settlementMap" :key="val" :label="label" :value="val" />
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
        <el-form-item label="岗位状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option v-for="(label, val) in statusMap" :key="val" :label="label" :value="Number(val)" />
          </el-select>
        </el-form-item>
        <el-form-item label="工作地">
          <el-input v-model="filters.city" placeholder="城市" clearable style="width: 120px" @keyup.enter="onSearch" />
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
          <el-button type="warning" :icon="Bottom" :disabled="!selection.length" @click="onBatchStatus(2)">批量下架</el-button>
          <el-button type="danger" :icon="Delete" :disabled="!selection.length" @click="onBatchDelete">批量删除</el-button>
        </div>
        <div class="toolbar-right">
          <el-button type="primary" :icon="Plus" @click="onCreate">新建岗位</el-button>
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
        <el-table-column label="封面" width="70">
          <template #default="{ row }">
            <el-image
              v-if="row.cover_image"
              :src="row.cover_image"
              fit="cover"
              class="cover-thumb"
              :preview-src-list="[row.cover_image]"
              preview-teleported
            />
            <div v-else class="cover-thumb cover-empty">无图</div>
          </template>
        </el-table-column>
        <el-table-column label="标题/标签" min-width="220">
          <template #default="{ row }">
            <div class="title-cell">
              <div class="title-text">
                <el-link type="primary" :underline="'never'" @click="openDetail(row)">{{ row.title }}</el-link>
                <el-tag v-if="row.featured" type="warning" size="small" effect="dark">精</el-tag>
                <el-tag v-if="row.verified" type="success" size="small">验</el-tag>
              </div>
              <div class="title-desc">
                <span class="price">¥{{ formatSalary(row) }}</span>
                <span class="price-unit">{{ row.salary_unit }}</span>
                <el-tag size="small" effect="plain" :type="billingTagType(row.billing_type)">{{ billingMap[row.billing_type] || row.billing_type }}</el-tag>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="typeTagType(row.linggong_type)">{{ typeMap[row.linggong_type] || row.linggong_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="发布者" width="150">
          <template #default="{ row }">
            <div class="publisher">
              <div class="publisher-name">{{ row.user_name || row.company_name || `用户#${row.user_id}` }}</div>
              <div class="publisher-phone">{{ maskPhone(row.contact_phone || row.user_phone) }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="招募" width="100">
          <template #default="{ row }">
            <span>{{ row.applied_count || 0 }} / {{ row.recruit_count || 0 }}</span>
          </template>
        </el-table-column>
        <el-table-column label="地区" width="120">
          <template #default="{ row }">
            <span v-if="row.city">{{ row.city }}{{ row.district ? '/' + row.district : '' }}</span>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="浏览" width="70" prop="view_count" sortable="custom" />
        <el-table-column label="收藏" width="70" prop="fav_count" sortable="custom" />
        <el-table-column label="报名" width="70" prop="application_count" />
        <el-table-column label="审核" width="90">
          <template #default="{ row }">
            <el-tag :type="auditTagType(row.audit_status)" size="small">{{ auditText(row.audit_status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small" effect="plain">{{ statusMap[row.status] || '-' }}</el-tag>
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
                  <el-dropdown-item :command="1" :disabled="row.status === 1">发布</el-dropdown-item>
                  <el-dropdown-item :command="2" :disabled="row.status === 2">下架</el-dropdown-item>
                  <el-dropdown-item :command="3" :disabled="row.status === 3">设为过期</el-dropdown-item>
                  <el-dropdown-item :command="5" :disabled="row.status === 5">设为满员</el-dropdown-item>
                  <el-dropdown-item :command="7" :disabled="row.status === 7">设为完成</el-dropdown-item>
                  <el-dropdown-item :command="'delete'" divided>删除</el-dropdown-item>
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
    <el-dialog v-model="detailVisible" title="零工岗位详情" width="900px" @close="onDetailClose" destroy-on-close>
      <div v-loading="detailLoading">
        <el-tabs v-if="detail" v-model="detailTab">
          <el-tab-pane label="基本信息" name="basic">
            <el-descriptions :column="3" border>
              <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
              <el-descriptions-item label="标题">{{ detail.title }}</el-descriptions-item>
              <el-descriptions-item label="岗位类型">
                <el-tag size="small" :type="typeTagType(detail.linggong_type)">{{ typeMap[detail.linggong_type] || detail.linggong_type }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="发布者类型">{{ publisherMap[detail.publisher_type] || detail.publisher_type }}</el-descriptions-item>
              <el-descriptions-item label="公司名称">{{ detail.company_name || '-' }}</el-descriptions-item>
              <el-descriptions-item label="联系人">{{ detail.contact_name || detail.user_name || '-' }}</el-descriptions-item>
              <el-descriptions-item label="联系电话">{{ maskPhone(detail.contact_phone || detail.user_phone) }}</el-descriptions-item>
              <el-descriptions-item label="微信">{{ detail.contact_wechat || '-' }}</el-descriptions-item>
              <el-descriptions-item label="计费">
                <el-tag size="small">{{ billingMap[detail.billing_type] || detail.billing_type }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="薪资">
                <span class="price">¥{{ formatSalary(detail) }}</span>
                <span class="price-unit">{{ detail.salary_unit }}</span>
              </el-descriptions-item>
              <el-descriptions-item label="结算周期">{{ settlementMap[detail.settlement] || detail.settlement }}</el-descriptions-item>
              <el-descriptions-item label="招募人数">{{ detail.recruit_count }}</el-descriptions-item>
              <el-descriptions-item label="已报名">{{ detail.applied_count }}</el-descriptions-item>
              <el-descriptions-item label="已确认">{{ detail.confirmed_count }}</el-descriptions-item>
              <el-descriptions-item label="工作日">{{ detail.work_days }} 天 / {{ detail.work_hours }} 小时</el-descriptions-item>
              <el-descriptions-item label="工作时间">{{ detail.work_time_start }} - {{ detail.work_time_end }}</el-descriptions-item>
              <el-descriptions-item label="性别要求">{{ genderText(detail.need_gender) }}</el-descriptions-item>
              <el-descriptions-item label="年龄要求">{{ detail.min_age || 0 }} - {{ detail.max_age || 0 }}</el-descriptions-item>
              <el-descriptions-item label="学历要求">{{ detail.education || '-' }}</el-descriptions-item>
              <el-descriptions-item label="经验要求">{{ detail.experience || '-' }}</el-descriptions-item>
              <el-descriptions-item label="需健康证">{{ detail.need_health_cert ? '是' : '否' }}</el-descriptions-item>
              <el-descriptions-item label="需身份证">{{ detail.need_id_card ? '是' : '否' }}</el-descriptions-item>
              <el-descriptions-item label="工作地点" :span="3">{{ formatAddress(detail) }}</el-descriptions-item>
              <el-descriptions-item label="工作方式">{{ workLocationText(detail.work_location_type) }}</el-descriptions-item>
              <el-descriptions-item label="工作强度">{{ intensityText(detail.work_intensity) }}</el-descriptions-item>
              <el-descriptions-item label="审核状态">
                <el-tag :type="auditTagType(detail.audit_status)" size="small">{{ auditText(detail.audit_status) }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="岗位状态">
                <el-tag :type="statusTagType(detail.status)" size="small" effect="plain">{{ statusMap[detail.status] || '-' }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item v-if="detail.audit_reason" label="审核原因" :span="3">{{ detail.audit_reason }}</el-descriptions-item>
              <el-descriptions-item v-if="detail.content" label="详细描述" :span="3">
                <div class="content-box">{{ detail.content }}</div>
              </el-descriptions-item>
              <el-descriptions-item label="浏览量">{{ detail.view_count }}</el-descriptions-item>
              <el-descriptions-item label="收藏量">{{ detail.fav_count }}</el-descriptions-item>
              <el-descriptions-item label="联系数">{{ detail.contact_count }}</el-descriptions-item>
              <el-descriptions-item label="分享数">{{ detail.share_count }}</el-descriptions-item>
              <el-descriptions-item label="报名时间">{{ formatTime(detail.last_applied_at) }}</el-descriptions-item>
              <el-descriptions-item label="发布时间">{{ formatTime(detail.published_at) }}</el-descriptions-item>
              <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
            </el-descriptions>
          </el-tab-pane>

          <el-tab-pane :label="`图集 (${images.length})`" name="images">
            <div v-if="!images.length" class="empty-text">暂无图片</div>
            <div v-else class="images-grid">
              <el-image
                v-for="(img, idx) in images"
                :key="idx"
                :src="img"
                fit="cover"
                class="image-item"
                :preview-src-list="images"
                :initial-index="idx"
                preview-teleported
              />
            </div>
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
  Briefcase, CirclePlus, Clock, Promotion, User, Warning
} from '@element-plus/icons-vue'
import {
  adminListLinggongs, adminGetLinggong, auditLinggong, adminUpdateLinggongStatus,
  batchUpdateLinggongStatus, deleteLinggong, getLinggongOverviewStats
} from '@/api/linggong'
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
  total: 0, todayNew: 0, pendingAudit: 0, published: 0, applied: 0, violation: 0
})

const loadStats = async () => {
  try {
    const res = await getLinggongOverviewStats()
    const d = res.data || {}
    stats.total = d.total || 0
    stats.todayNew = d.today_new || 0
    stats.pendingAudit = d.pending_audit || 0
    stats.published = d.published || 0
    stats.applied = d.applied || 0
    stats.violation = d.violation || 0
  } catch (e) {
    // 接口失败时统计保持默认值 0
  }
}

// ===== 筛选 =====
const filters = reactive({
  keyword: '', linggong_type: '', publisher_type: '', billing_type: '', settlement: '',
  min_salary: undefined, max_salary: undefined,
  audit_status: null, status: null, city: '', dateRange: null
})

const onSearch = () => {
  page.value = 1
  loadList()
}

const onReset = () => {
  Object.assign(filters, {
    keyword: '', linggong_type: '', publisher_type: '', billing_type: '', settlement: '',
    min_salary: undefined, max_salary: undefined,
    audit_status: null, status: null, city: '', dateRange: null
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

// ===== 状态格式化 =====
const typeMap = {
  short_term: '短期', long_term: '长期', task: '任务制',
  hourly: '小时工', daily: '日结', temp: '临时工'
}
const typeTagType = (t) => ({
  short_term: 'primary', long_term: 'success', task: 'warning',
  hourly: 'info', daily: 'danger', temp: 'info'
}[t] || 'info')

const publisherMap = {
  personal: '个人', company: '企业', agent: '中介', headhunter: '猎头'
}

const billingMap = {
  by_piece: '按件', by_hour: '按时', by_day: '按日', by_week: '按周',
  by_month: '按月', fixed: '固定', negotiable: '面议'
}
const billingTagType = (b) => ({
  by_piece: 'warning', by_hour: 'info', by_day: 'primary',
  by_week: 'info', by_month: 'success', fixed: 'success', negotiable: 'info'
}[b] || 'info')

const settlementMap = {
  'T+0': '当日结', 'T+1': '次日结', 'T+3': '三日结',
  'T+7': '周结', 'M+1': '月结', project: '项目结'
}

const statusMap = {
  0: '草稿', 1: '已发布', 2: '已下架', 3: '已过期',
  4: '已删除', 5: '已满员', 6: '已关闭', 7: '已完成'
}
const statusTagType = (s) => ({
  0: 'info', 1: 'success', 2: 'warning', 3: 'danger',
  4: 'info', 5: 'primary', 6: 'info', 7: 'success'
}[s] || 'info')

const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')
const genderText = (g) => ({ any: '不限', male: '男', female: '女' }[g] || '不限')
const workLocationText = (w) => ({ onsite: '现场', remote: '远程', hybrid: '混合' }[w] || '-')
const intensityText = (i) => ({ light: '轻松', medium: '中等', heavy: '繁重', extreme: '极重' }[i] || '-')

const formatSalary = (row) => {
  if (!row) return '0'
  if (row.salary_negotiable) return '面议'
  if (row.salary_min && row.salary_max) return `${Number(row.salary_min).toFixed(0)}-${Number(row.salary_max).toFixed(0)}`
  if (row.salary_min) return Number(row.salary_min).toFixed(0)
  if (row.salary_max) return Number(row.salary_max).toFixed(0)
  return '0'
}

const formatAddress = (row) => {
  if (!row) return '-'
  return [row.province, row.city, row.district, row.business_district, row.address].filter(Boolean).join(' ') || '-'
}

// 手机号脱敏
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
      linggong_type: filters.linggong_type || undefined,
      publisher_type: filters.publisher_type || undefined,
      billing_type: filters.billing_type || undefined,
      settlement: filters.settlement || undefined,
      min_salary: filters.min_salary || undefined,
      max_salary: filters.max_salary || undefined,
      audit_status: filters.audit_status === null || filters.audit_status === '' ? undefined : filters.audit_status,
      status: filters.status === null || filters.status === '' ? undefined : filters.status,
      city: filters.city || undefined,
      sort: sortField.value,
      order: sortOrder.value
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await adminListLinggongs(params)
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
const images = ref([])
const detailTab = ref('basic')

const openDetail = async (row) => {
  detail.value = row
  images.value = row.images || []
  detailTab.value = 'basic'
  detailVisible.value = true
  detailLoading.value = true
  try {
    const dRes = await adminGetLinggong(row.id)
    if (dRes.data) {
      detail.value = dRes.data
      images.value = dRes.data.images || []
    }
  } catch (e) {
    // 接口失败保持现状
  } finally {
    detailLoading.value = false
  }
}

const onDetailClose = () => {
  detail.value = null
  images.value = []
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
      await auditLinggong(row.id, { audit_status: auditStatus, audit_reason: value || '' })
    } else {
      await ElMessageBox.confirm(`确定${action}岗位 "${row.title}" 的审核吗？`, '提示', { type: 'warning' })
      await auditLinggong(row.id, { audit_status: auditStatus })
    }
    ElMessage.success(`已${action}`)
    await loadList()
    if (detailVisible.value && detail.value?.id === row.id) {
      const res = await adminGetLinggong(row.id)
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
      await ElMessageBox.confirm(`确定删除岗位 "${row.title}" 吗？删除后不可恢复！`, '危险操作', {
        type: 'error',
        confirmButtonText: '确认删除',
        cancelButtonText: '取消'
      })
      await deleteLinggong(row.id)
      ElMessage.success('已删除')
      await loadList()
      return
    }
    const label = statusMap[cmd]
    await ElMessageBox.confirm(`确定将岗位 "${row.title}" 设为「${label}」吗？`, '提示', { type: 'warning' })
    await adminUpdateLinggongStatus(row.id, cmd)
    ElMessage.success('状态更新成功')
    await loadList()
    if (detailVisible.value && detail.value?.id === row.id) {
      const res = await adminGetLinggong(row.id)
      if (res.data) detail.value = res.data
    }
  } catch (e) {
    // 取消
  }
}

// ===== 批量操作 =====
const onBatchAudit = async (auditStatus) => {
  try {
    await ElMessageBox.confirm(`确认批量审核通过 ${selection.value.length} 个岗位？`, '批量审核', { type: 'warning' })
    await Promise.all(selection.value.map((r) => auditLinggong(r.id, { audit_status: auditStatus })))
    ElMessage.success('批量审核完成')
    await loadList()
  } catch (e) {
    // 取消
  }
}

const onBatchStatus = async (status) => {
  try {
    const label = statusMap[status]
    await ElMessageBox.confirm(`确认批量将 ${selection.value.length} 个岗位设为「${label}」？`, '批量状态变更', { type: 'warning' })
    await batchUpdateLinggongStatus({
      ids: selection.value.map((r) => r.id),
      status
    })
    ElMessage.success('批量操作完成')
    await loadList()
  } catch (e) {
    // 取消
  }
}

const onBatchDelete = async () => {
  try {
    await ElMessageBox.confirm(`确认批量删除 ${selection.value.length} 个岗位？删除后不可恢复！`, '危险操作', {
      type: 'error',
      confirmButtonText: '确认删除',
      cancelButtonText: '取消'
    })
    await Promise.all(selection.value.map((r) => deleteLinggong(r.id)))
    ElMessage.success('批量删除完成')
    await loadList()
  } catch (e) {
    // 取消
  }
}

// ===== 新建 =====
const onCreate = () => {
  router.push('/business/linggong/list')
  ElMessage.info('请使用 C 端发布岗位')
}

onMounted(async () => {
  await loadStats()
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

.cover-thumb {
  width: 50px; height: 50px; border-radius: 4px; border: 1px solid #ebeef5;
}
.cover-empty {
  display: flex; align-items: center; justify-content: center;
  background: #fafafa; color: #909399; font-size: 12px;
}
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
.price { color: #f56c6c; font-weight: 600; }
.price-unit { color: #909399; font-size: 12px; margin-left: 2px; }
.publisher { font-size: 13px; }
.publisher-name { color: #303133; }
.publisher-phone { font-size: 12px; color: #909399; margin-top: 2px; }
.text-muted { color: #909399; }

.images-grid { display: flex; flex-wrap: wrap; gap: 8px; }
.image-item {
  width: 120px; height: 120px; border-radius: 4px; border: 1px solid #ebeef5;
}
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
