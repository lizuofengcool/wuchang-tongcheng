<template>
  <div class="app-container">
    <!-- 统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="6" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total_coupons }}</div><div class="stat-label">优惠券总数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="6" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.active_coupons }}</div><div class="stat-label">进行中</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="6" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.total_received }}</div><div class="stat-label">累计领取</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="6" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-muted">{{ stats.total_used }}</div><div class="stat-label">累计使用</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="优惠券标题" clearable style="width: 200px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="filters.type" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option label="满减券" value="reduce" />
            <el-option label="折扣券" value="discount" />
            <el-option label="兑换券" value="exchange" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="禁用" :value="0" />
            <el-option label="进行中" :value="1" />
            <el-option label="草稿" :value="2" />
            <el-option label="已下架" :value="3" />
            <el-option label="已过期" :value="4" />
            <el-option label="已抢完" :value="5" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar">
        <el-button :icon="Refresh" @click="loadList">刷新</el-button>
        <el-button type="primary" :icon="Plus" @click="openCreate">新建优惠券</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="title" label="标题" min-width="180" show-overflow-tooltip />
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="typeTagType(row.type)" size="small">{{ typeText(row) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="面值/折扣" width="120">
          <template #default="{ row }">
            <span class="amount">{{ amountText(row) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="使用门槛" width="110">
          <template #default="{ row }">
            <span v-if="row.threshold > 0">满 {{ row.threshold }} 元</span>
            <span v-else class="text-muted">无门槛</span>
          </template>
        </el-table-column>
        <el-table-column label="领取进度" width="150">
          <template #default="{ row }">
            <span v-if="row.total_count > 0">{{ row.received_count }} / {{ row.total_count }}</span>
            <span v-else>{{ row.received_count }} / 不限</span>
          </template>
        </el-table-column>
        <el-table-column label="领取时间" width="220">
          <template #default="{ row }">
            <div class="time-cell">
              <div>{{ formatTime(row.start_at) }}</div>
              <div>至 {{ formatTime(row.end_at) }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusText(row) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openEdit(row)">编辑</el-button>
            <el-button v-if="row.status !== 1" type="success" link size="small" @click="onStatus(row, 1)">上架</el-button>
            <el-button v-if="row.status === 1" type="warning" link size="small" @click="onStatus(row, 3)">下架</el-button>
            <el-button type="danger" link size="small" @click="onDelete(row)">删除</el-button>
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

    <!-- 表单弹窗 -->
    <el-dialog v-model="formVisible" :title="formTitle" width="720px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="标题" prop="title">
          <el-input v-model="form.title" maxlength="100" placeholder="优惠券标题" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="类型" prop="type">
              <el-select v-model="form.type" style="width: 100%">
                <el-option label="满减券" value="reduce" />
                <el-option label="折扣券" value="discount" />
                <el-option label="兑换券" value="exchange" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="状态">
              <el-select v-model="form.status" style="width: 100%">
                <el-option label="禁用" :value="0" />
                <el-option label="进行中" :value="1" />
                <el-option label="草稿" :value="2" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item :label="amountLabel" prop="amount">
              <el-input-number v-model="form.amount" :min="0" :precision="2" :controls="false" style="width: 100%" />
              <div class="form-tip">{{ amountTip }}</div>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="使用门槛">
              <el-input-number v-model="form.threshold" :min="0" :precision="2" :controls="false" style="width: 100%" />
              <div class="form-tip">满此金额可用，0 表示无门槛</div>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="发放总量">
          <el-input-number v-model="form.total_count" :min="0" :controls="false" style="width: 100%" />
          <div class="form-tip">0 表示不限领取数量</div>
        </el-form-item>
        <el-form-item label="领取时间">
          <el-date-picker
            v-model="form.dateRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            value-format="YYYY-MM-DDTHH:mm:ssZ"
            style="width: 100%"
          />
          <div class="form-tip">不选表示不限领取时间</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onSubmit">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search, Plus } from '@element-plus/icons-vue'
import {
  getMarketingCouponList,
  createMarketingCoupon,
  updateMarketingCoupon,
  deleteMarketingCoupon,
  getMarketingCouponStatistics
} from '@/api/marketing'

const loading = ref(false)
const submitting = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const stats = reactive({
  total_coupons: 0,
  active_coupons: 0,
  total_received: 0,
  total_used: 0,
  receive_rate: 0,
  usage_rate: 0
})

const filters = reactive({ keyword: '', type: '', status: null })

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.keyword = ''
  filters.type = ''
  filters.status = null
  page.value = 1
  loadList()
}

const typeText = (row) => {
  if (row.type_text) return row.type_text
  return { reduce: '满减券', discount: '折扣券', exchange: '兑换券' }[row.type] || '-'
}
const typeTagType = (t) => ({ reduce: 'danger', discount: 'warning', exchange: 'success' }[t] || 'info')

const amountText = (row) => {
  if (row.type === 'discount') {
    // 折扣券：amount 为折扣率 0.01-0.99
    return `${row.amount} 折`
  }
  if (row.type === 'exchange') {
    return '兑换券'
  }
  // 满减券：面值
  return `¥${row.amount}`
}

const statusMap = { 0: '禁用', 1: '进行中', 2: '草稿', 3: '已下架', 4: '已过期', 5: '已抢完' }
const statusTagType = (s) => ({ 0: 'info', 1: 'success', 2: 'warning', 3: 'info', 4: 'danger', 5: 'danger' }[s] || 'info')
const statusText = (row) => {
  if (row.status_text) return row.status_text
  return statusMap[row.status] || '未知'
}

const formatTime = (t) => {
  if (!t) return '不限'
  return String(t).replace('T', ' ').slice(0, 16)
}

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      keyword: filters.keyword || undefined,
      type: filters.type || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status
    }
    const res = await getMarketingCouponList(params)
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

const loadStats = async () => {
  try {
    const res = await getMarketingCouponStatistics()
    Object.assign(stats, res.data || {})
  } catch (e) {}
}

// ===== 表单 =====
const formVisible = ref(false)
const formRef = ref(null)
const isEdit = ref(false)
const formTitle = computed(() => isEdit.value ? '编辑优惠券' : '新建优惠券')
const form = reactive({
  id: null,
  title: '',
  type: 'reduce',
  amount: 0,
  threshold: 0,
  total_count: 0,
  dateRange: [],
  status: 1
})
const rules = {
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
  type: [{ required: true, message: '请选择类型', trigger: 'change' }],
  amount: [{ required: true, message: '请输入面值/折扣', trigger: 'blur' }]
}

const amountLabel = computed(() => {
  if (form.type === 'discount') return '折扣率'
  if (form.type === 'exchange') return '面值'
  return '面值(元)'
})
const amountTip = computed(() => {
  if (form.type === 'discount') return '折扣券：0.01-0.99（如 0.8 表示 8 折）'
  if (form.type === 'exchange') return '兑换券：可填面值用于展示'
  return '满减券：减免金额（元）'
})

const resetForm = () => {
  Object.assign(form, {
    id: null,
    title: '',
    type: 'reduce',
    amount: 0,
    threshold: 0,
    total_count: 0,
    dateRange: [],
    status: 1
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
    id: row.id,
    title: row.title || '',
    type: row.type || 'reduce',
    amount: row.amount ?? 0,
    threshold: row.threshold ?? 0,
    total_count: row.total_count ?? 0,
    dateRange: row.start_at && row.end_at ? [row.start_at, row.end_at] : [],
    status: row.status ?? 1
  })
  formVisible.value = true
}

const onSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    // 折扣券额外校验
    if (form.type === 'discount' && (form.amount <= 0 || form.amount >= 1)) {
      ElMessage.error('折扣券的折扣率需在 0.01-0.99 之间')
      return
    }
    submitting.value = true
    try {
      const payload = {
        title: form.title,
        type: form.type,
        amount: form.amount,
        threshold: form.threshold,
        total_count: form.total_count,
        start_at: form.dateRange && form.dateRange[0] ? form.dateRange[0] : null,
        end_at: form.dateRange && form.dateRange[1] ? form.dateRange[1] : null,
        status: form.status
      }
      if (isEdit.value) {
        await updateMarketingCoupon(form.id, payload)
        ElMessage.success('更新成功')
      } else {
        await createMarketingCoupon(payload)
        ElMessage.success('创建成功')
      }
      formVisible.value = false
      loadList()
      loadStats()
    } catch (e) {
      // 错误已由拦截器处理
    } finally {
      submitting.value = false
    }
  })
}

const onStatus = (row, status) => {
  const action = status === 1 ? '上架' : '下架'
  ElMessageBox.confirm(`确认${action}优惠券「${row.title}」？`, '提示', {
    type: 'warning',
    confirmButtonText: '确认',
    cancelButtonText: '取消'
  }).then(async () => {
    try {
      await updateMarketingCoupon(row.id, { status })
      ElMessage.success(`${action}成功`)
      loadList()
      loadStats()
    } catch (e) {}
  }).catch(() => {})
}

const onDelete = (row) => {
  ElMessageBox.confirm(`确认删除优惠券「${row.title}」？`, '提示', {
    type: 'warning',
    confirmButtonText: '确认',
    cancelButtonText: '取消'
  }).then(async () => {
    try {
      await deleteMarketingCoupon(row.id)
      ElMessage.success('删除成功')
      loadList()
      loadStats()
    } catch (e) {}
  }).catch(() => {})
}

onMounted(() => {
  loadList()
  loadStats()
})
</script>

<style scoped>
.page-card { background: #fff; padding: 16px; border-radius: 8px; box-shadow: 0 1px 4px rgba(0,0,0,0.04); }
.filter-form { margin-bottom: 12px; }
.toolbar { margin-bottom: 12px; display: flex; gap: 8px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.time-cell { font-size: 12px; color: #666; line-height: 1.6; }
.stat-row { margin-bottom: 16px; }
.stat-item { text-align: center; padding: 8px 0; }
.stat-value { font-size: 22px; font-weight: 600; color: #303133; }
.stat-label { font-size: 12px; color: #909399; margin-top: 4px; }
.amount { color: #f56c6c; font-weight: 600; }
.text-success { color: #67c23a; }
.text-warning { color: #e6a23c; }
.text-muted { color: #909399; }
.form-tip { font-size: 12px; color: #909399; line-height: 1.4; margin-top: 4px; }
</style>
