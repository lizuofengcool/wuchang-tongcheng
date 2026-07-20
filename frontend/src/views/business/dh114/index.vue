<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #409eff">
            <el-icon :size="22"><OfficeBuilding /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.total }}</div>
            <div class="stat-label">总商户数</div>
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
            <el-icon :size="22"><Postcard /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.verified }}</div>
            <div class="stat-label">已认证</div>
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

    <!-- 筛选区 -->
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input
            v-model="filters.keyword"
            placeholder="商户名/联系电话"
            clearable
            style="width: 200px"
            :prefix-icon="Search"
            @keyup.enter="onSearch"
          />
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="filters.category_id" placeholder="全部" clearable filterable style="width: 160px" @change="onSearch">
            <el-option v-for="c in categoryOptions" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="商户类型">
          <el-select v-model="filters.business_type" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option label="餐饮美食" value="food" />
            <el-option label="酒店住宿" value="hotel" />
            <el-option label="生活服务" value="life" />
            <el-option label="休闲娱乐" value="leisure" />
            <el-option label="购物" value="shopping" />
            <el-option label="教育培训" value="education" />
            <el-option label="医疗健康" value="medical" />
          </el-select>
        </el-form-item>
        <el-form-item label="审核状态">
          <el-select v-model="filters.audit_status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="待审核" :value="0" />
            <el-option label="已通过" :value="1" />
            <el-option label="已拒绝" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="商户状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="草稿" :value="0" />
            <el-option label="已发布" :value="1" />
            <el-option label="已下架" :value="3" />
            <el-option label="已关闭" :value="4" />
          </el-select>
        </el-form-item>
        <el-form-item label="认证状态">
          <el-select v-model="filters.verification_status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="未认证" :value="0" />
            <el-option label="待审核" :value="1" />
            <el-option label="已认证" :value="2" />
            <el-option label="认证失败" :value="3" />
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
          <el-button type="primary" :icon="Plus" @click="openCreate">新建商户</el-button>
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
        <el-table-column label="商户名/标签" min-width="220">
          <template #default="{ row }">
            <div class="title-cell">
              <div class="title-text">
                <el-link type="primary" :underline="'never'" @click="openDetail(row)">{{ row.name }}</el-link>
                <el-tag v-if="row.is_recommended" type="warning" size="small" effect="dark">荐</el-tag>
                <el-tag v-if="row.verification_status === 2" type="success" size="small" effect="dark">认证</el-tag>
              </div>
              <div class="title-desc">
                <span class="text-muted">{{ row.business_type_text || row.business_type || '-' }}</span>
                <span v-if="row.address" class="text-muted">· {{ row.address }}</span>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="分类" width="110">
          <template #default="{ row }">{{ categoryName(row.category_id) }}</template>
        </el-table-column>
        <el-table-column label="联系电话" width="140">
          <template #default="{ row }">
            <span v-if="row.contact_phone">{{ maskPhone(row.contact_phone) }}</span>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="评分" width="100">
          <template #default="{ row }">
            <el-rate v-if="row.rating" :model-value="row.rating" disabled size="small" />
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="浏览" width="70" prop="view_count" sortable="custom" />
        <el-table-column label="收藏" width="70" prop="fav_count" sortable="custom" />
        <el-table-column label="电话量" width="80" prop="call_count" />
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
        <el-table-column label="创建时间" width="160" prop="created_at" sortable="custom">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
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
            <el-dropdown trigger="click" @command="(cmd) => handleCommand(row, cmd)">
              <el-button type="warning" link size="small">
                更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item :command="1" :disabled="row.status === 1">发布</el-dropdown-item>
                  <el-dropdown-item :command="3" :disabled="row.status === 3">下架</el-dropdown-item>
                  <el-dropdown-item :command="4" :disabled="row.status === 4">关闭</el-dropdown-item>
                  <el-dropdown-item :command="'edit'">编辑</el-dropdown-item>
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

    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="formVisible" :title="formTitle" width="720px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="商户名" prop="name">
              <el-input v-model="form.name" maxlength="128" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="商户类型" prop="business_type">
              <el-select v-model="form.business_type" style="width: 100%">
                <el-option label="餐饮美食" value="food" />
                <el-option label="酒店住宿" value="hotel" />
                <el-option label="生活服务" value="life" />
                <el-option label="休闲娱乐" value="leisure" />
                <el-option label="购物" value="shopping" />
                <el-option label="教育培训" value="education" />
                <el-option label="医疗健康" value="medical" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="分类">
              <el-select v-model="form.category_id" filterable clearable style="width: 100%">
                <el-option v-for="c in categoryOptions" :key="c.id" :label="c.name" :value="c.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="联系电话" prop="contact_phone">
              <el-input v-model="form.contact_phone" maxlength="20" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="封面图">
          <el-input v-model="form.cover_image" placeholder="图片 URL" />
        </el-form-item>
        <el-form-item label="地址">
          <el-input v-model="form.address" maxlength="256" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="经度">
              <el-input-number v-model="form.longitude" :controls="false" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="纬度">
              <el-input-number v-model="form.latitude" :controls="false" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="简介">
          <el-input v-model="form.summary" type="textarea" :rows="2" maxlength="500" show-word-limit />
        </el-form-item>
        <el-form-item label="详细描述">
          <el-input v-model="form.content" type="textarea" :rows="4" maxlength="2000" show-word-limit />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="8">
            <el-form-item label="推荐">
              <el-switch v-model="form.is_recommended" :active-value="1" :inactive-value="0" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="排序">
              <el-input-number v-model="form.sort" :min="0" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="状态">
              <el-select v-model="form.status" style="width: 100%">
                <el-option label="草稿" :value="0" />
                <el-option label="已发布" :value="1" />
                <el-option label="已下架" :value="3" />
                <el-option label="已关闭" :value="4" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="formLoading" @click="onSubmit">确认</el-button>
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
  OfficeBuilding, CirclePlus, Clock, Promotion, Postcard, Warning
} from '@element-plus/icons-vue'
import request from '@/utils/request'
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

// ===== 统计 =====
const stats = reactive({
  total: 0, todayNew: 0, pendingAudit: 0, published: 0, verified: 0, violation: 0
})

const loadStats = async () => {
  try {
    const res = await request.get('/dh114/admin/statistics/overview')
    const d = res.data || {}
    stats.total = d.total || 0
    stats.todayNew = d.today_new || 0
    stats.pendingAudit = d.pending_audit || 0
    stats.published = d.published || 0
    stats.verified = d.verified || 0
    stats.violation = d.violation || 0
  } catch (e) {
    // 接口失败保持默认 0
  }
}

// ===== 筛选 =====
const filters = reactive({
  keyword: '', category_id: null, business_type: '',
  audit_status: null, status: null, verification_status: null, dateRange: null
})

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.keyword = ''
  filters.category_id = null
  filters.business_type = ''
  filters.audit_status = null
  filters.status = null
  filters.verification_status = null
  filters.dateRange = null
  page.value = 1
  loadList()
}
const onSortChange = ({ prop, order }) => {
  sortField.value = prop || 'created_at'
  sortOrder.value = order || 'descending'
  loadList()
}
const onSelectionChange = (rows) => { selection.value = rows }

// ===== 分类 =====
const categoryOptions = ref([])
const loadCategories = async () => {
  try {
    const res = await request.get('/dh114/categories', { params: { page: 1, page_size: 200 } })
    const data = res.data || {}
    categoryOptions.value = data.list || data || []
  } catch (e) {
    categoryOptions.value = []
  }
}
const categoryName = (id) => categoryOptions.value.find((c) => c.id === id)?.name || '-'

// ===== 格式化 =====
const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')
const statusText = (s) => ({ 0: '草稿', 1: '已发布', 3: '已下架', 4: '已关闭' }[s] || '-')
const statusTagType = (s) => ({ 0: 'info', 1: 'success', 3: 'warning', 4: 'danger' }[s] || 'info')

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
      business_type: filters.business_type || undefined,
      audit_status: filters.audit_status === null || filters.audit_status === '' ? undefined : filters.audit_status,
      status: filters.status === null || filters.status === '' ? undefined : filters.status,
      verification_status: filters.verification_status === null || filters.verification_status === '' ? undefined : filters.verification_status,
      sort: sortField.value,
      order: sortOrder.value
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await request.get('/dh114/admin/dh114s', { params })
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
    if (data.stats) Object.assign(stats, data.stats)
  } catch (e) {
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

// ===== 详情 =====
const openDetail = (row) => {
  router.push(`/business/dh114/detail/${row.id}`)
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
      await request.put(`/dh114/admin/dh114s/${row.id}/audit`, { audit_status: auditStatus, audit_reason: value || '' })
    } else {
      await ElMessageBox.confirm(`确定${action}商户 "${row.name}" 的审核吗？`, '提示', { type: 'warning' })
      await request.put(`/dh114/admin/dh114s/${row.id}/audit`, { audit_status: auditStatus })
    }
    ElMessage.success(`已${action}`)
    await loadList()
  } catch (e) {
    // 取消
  }
}

// ===== 状态变更 =====
const handleCommand = async (row, cmd) => {
  try {
    if (cmd === 'delete') {
      await ElMessageBox.confirm(`确定删除商户 "${row.name}" 吗？删除后不可恢复！`, '危险操作', {
        type: 'error', confirmButtonText: '确认删除', cancelButtonText: '取消'
      })
      await request.delete(`/dh114/${row.id}`)
      ElMessage.success('已删除')
      await loadList()
      return
    }
    if (cmd === 'edit') {
      openEdit(row)
      return
    }
    const label = statusText(cmd)
    await ElMessageBox.confirm(`确定将商户 "${row.name}" 设为「${label}」吗？`, '提示', { type: 'warning' })
    await request.put(`/dh114/admin/dh114s/${row.id}/status`, { status: cmd })
    ElMessage.success('状态更新成功')
    await loadList()
  } catch (e) {
    // 取消
  }
}

// ===== 批量操作 =====
const onBatchAudit = async (auditStatus) => {
  try {
    await ElMessageBox.confirm(`确认批量审核通过 ${selection.value.length} 个商户？`, '批量审核', { type: 'warning' })
    await request.post('/dh114/admin/dh114s/batch-audit', {
      ids: selection.value.map((r) => r.id),
      audit_status: auditStatus
    })
    ElMessage.success('批量审核完成')
    await loadList()
  } catch (e) {
    // 取消
  }
}

const onBatchStatus = async (status) => {
  try {
    const label = statusText(status)
    await ElMessageBox.confirm(`确认批量将 ${selection.value.length} 个商户设为「${label}」？`, '批量状态变更', { type: 'warning' })
    await request.put(`/dh114/admin/dh114s/${selection.value[0]?.id}/status`, { status, ids: selection.value.map((r) => r.id) })
    ElMessage.success('批量操作完成')
    await loadList()
  } catch (e) {
    // 取消
  }
}

const onBatchDelete = async () => {
  try {
    await ElMessageBox.confirm(`确认批量删除 ${selection.value.length} 个商户？删除后不可恢复！`, '危险操作', {
      type: 'error', confirmButtonText: '确认删除', cancelButtonText: '取消'
    })
    for (const row of selection.value) {
      await request.delete(`/dh114/${row.id}`)
    }
    ElMessage.success('批量删除完成')
    await loadList()
  } catch (e) {
    // 取消
  }
}

// ===== 新增/编辑 =====
const formVisible = ref(false)
const formLoading = ref(false)
const formRef = ref(null)
const isEdit = ref(false)
const formTitle = computed(() => isEdit.value ? '编辑商户' : '新建商户')
const form = reactive({
  id: null, name: '', business_type: 'food', category_id: null,
  contact_phone: '', cover_image: '', address: '',
  longitude: undefined, latitude: undefined,
  summary: '', content: '', is_recommended: 0, sort: 0, status: 0
})
const rules = {
  name: [{ required: true, message: '请输入商户名', trigger: 'blur' }],
  business_type: [{ required: true, message: '请选择类型', trigger: 'change' }],
  contact_phone: [{ required: true, message: '请输入联系电话', trigger: 'blur' }]
}

const resetForm = () => {
  Object.assign(form, {
    id: null, name: '', business_type: 'food', category_id: null,
    contact_phone: '', cover_image: '', address: '',
    longitude: undefined, latitude: undefined,
    summary: '', content: '', is_recommended: 0, sort: 0, status: 0
  })
}

const openCreate = () => {
  isEdit.value = false
  resetForm()
  formVisible.value = true
}

const openEdit = (row) => {
  isEdit.value = true
  Object.assign(form, {
    id: row.id, name: row.name || '', business_type: row.business_type || 'food',
    category_id: row.category_id || null, contact_phone: row.contact_phone || '',
    cover_image: row.cover_image || '', address: row.address || '',
    longitude: row.longitude, latitude: row.latitude,
    summary: row.summary || '', content: row.content || '',
    is_recommended: row.is_recommended || 0, sort: row.sort || 0, status: row.status || 0
  })
  formVisible.value = true
}

const onSubmit = async () => {
  try {
    await formRef.value.validate()
    formLoading.value = true
    const payload = { ...form }
    if (isEdit.value) {
      await request.put(`/dh114/${form.id}`, payload)
    } else {
      await request.post('/dh114', payload)
    }
    ElMessage.success(isEdit.value ? '更新成功' : '创建成功')
    formVisible.value = false
    await loadList()
  } catch (e) {
    // 校验或接口失败
  } finally {
    formLoading.value = false
  }
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
  margin-bottom: 12px; padding: 12px 16px;
  background: #fafafa; border-radius: 4px;
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
.title-desc { font-size: 12px; color: #909399; }
.text-muted { color: #909399; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }

@media (max-width: 1200px) {
  .filter-form :deep(.el-form-item) { margin-right: 8px; }
}
</style>
