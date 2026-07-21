<template>
  <div class="app-container">
    <!-- 统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="6" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total_activities }}</div><div class="stat-label">总活动数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="6" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.ongoing_activities }}</div><div class="stat-label">进行中</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="6" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.pending_activities }}</div><div class="stat-label">未开始</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="6" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-muted">{{ stats.ended_activities }}</div><div class="stat-label">已结束</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="活动标题" clearable style="width: 200px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="filters.type" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option label="拼团" value="groupbuy" />
            <el-option label="砍价" value="bargain" />
            <el-option label="秒杀" value="seckill" />
            <el-option label="抽奖" value="lottery" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
            <el-option label="待生效" :value="2" />
            <el-option label="已结束" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar">
        <el-button :icon="Refresh" @click="loadList">刷新</el-button>
        <el-button type="primary" :icon="Plus" @click="openCreate">新建活动</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column label="封面" width="100">
          <template #default="{ row }">
            <el-image v-if="row.cover_image" :src="row.cover_image" :preview-src-list="[row.cover_image]" fit="cover" style="width: 60px; height: 40px" />
            <span v-else class="text-muted">无</span>
          </template>
        </el-table-column>
        <el-table-column prop="title" label="标题" min-width="180" show-overflow-tooltip />
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="typeTagType(row.type)" size="small">{{ typeText(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="活动时间" width="220">
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
            <el-button v-if="row.status !== 1" type="success" link size="small" @click="onStatus(row, 1)">启用</el-button>
            <el-button v-if="row.status === 1" type="warning" link size="small" @click="onStatus(row, 0)">禁用</el-button>
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
    <el-dialog v-model="formVisible" :title="formTitle" width="780px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="标题" prop="title">
          <el-input v-model="form.title" maxlength="100" placeholder="活动标题" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="类型" prop="type">
              <el-select v-model="form.type" style="width: 100%">
                <el-option label="拼团" value="groupbuy" />
                <el-option label="砍价" value="bargain" />
                <el-option label="秒杀" value="seckill" />
                <el-option label="抽奖" value="lottery" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="状态">
              <el-select v-model="form.status" style="width: 100%">
                <el-option label="启用" :value="1" />
                <el-option label="禁用" :value="0" />
                <el-option label="待生效" :value="2" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="封面图">
          <el-input v-model="form.cover_image" maxlength="500" placeholder="封面图 URL（可选）" />
        </el-form-item>
        <el-form-item label="活动时间">
          <el-date-picker
            v-model="form.dateRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            value-format="YYYY-MM-DDTHH:mm:ssZ"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" maxlength="500" show-word-limit placeholder="活动描述" />
        </el-form-item>
        <el-form-item label="活动配置">
          <el-input
            v-model="form.config_json"
            type="textarea"
            :rows="5"
            placeholder='JSON 格式，例如 {"group_size": 3, "discount_rate": 0.8}'
          />
          <div class="form-tip">可选，JSON 格式活动配置（拼团人数、砍价底价、秒杀库存、抽奖规则等）</div>
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
  getMarketingActivityList,
  createMarketingActivity,
  updateMarketingActivity,
  deleteMarketingActivity,
  updateMarketingActivityStatus,
  getMarketingActivityStatistics
} from '@/api/marketing'

const loading = ref(false)
const submitting = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const stats = reactive({
  total_activities: 0,
  ongoing_activities: 0,
  pending_activities: 0,
  ended_activities: 0
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

const typeText = (t) => ({ groupbuy: '拼团', bargain: '砍价', seckill: '秒杀', lottery: '抽奖' }[t] || '-')
const typeTagType = (t) => ({ groupbuy: 'warning', bargain: 'success', seckill: 'danger', lottery: 'primary' }[t] || 'info')

const statusMap = { 0: '禁用', 1: '进行中', 2: '待生效', 3: '已结束' }
const statusTagType = (s) => ({ 0: 'info', 1: 'success', 2: 'warning', 3: 'danger' }[s] || 'info')
const statusText = (row) => {
  // 优先使用后端返回的 status_text，回退到本地映射
  if (row.status_text) return row.status_text
  return statusMap[row.status] || '未知'
}

const formatTime = (t) => {
  if (!t) return '不限'
  return String(t).replace('T', ' ').slice(0, 16)
}

const formatConfig = (config) => {
  if (!config) return ''
  if (typeof config === 'string') return config
  try {
    return JSON.stringify(config, null, 2)
  } catch (e) {
    return String(config)
  }
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
    const res = await getMarketingActivityList(params)
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
    const res = await getMarketingActivityStatistics()
    Object.assign(stats, res.data || {})
  } catch (e) {}
}

// ===== 表单 =====
const formVisible = ref(false)
const formRef = ref(null)
const isEdit = ref(false)
const formTitle = computed(() => isEdit.value ? '编辑活动' : '新建活动')
const form = reactive({
  id: null,
  title: '',
  type: 'groupbuy',
  description: '',
  cover_image: '',
  dateRange: [],
  status: 1,
  config_json: ''
})
const rules = {
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
  type: [{ required: true, message: '请选择类型', trigger: 'change' }]
}

const resetForm = () => {
  Object.assign(form, {
    id: null,
    title: '',
    type: 'groupbuy',
    description: '',
    cover_image: '',
    dateRange: [],
    status: 1,
    config_json: ''
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
    type: row.type || 'groupbuy',
    description: row.description || '',
    cover_image: row.cover_image || '',
    dateRange: row.start_at && row.end_at ? [row.start_at, row.end_at] : [],
    status: row.status ?? 1,
    config_json: row.config ? formatConfig(row.config) : ''
  })
  formVisible.value = true
}

const onSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    // 校验活动配置 JSON
    let config = null
    if (form.config_json && form.config_json.trim()) {
      try {
        config = JSON.parse(form.config_json)
      } catch (e) {
        ElMessage.error('活动配置 JSON 格式错误')
        return
      }
    }
    submitting.value = true
    try {
      const payload = {
        title: form.title,
        type: form.type,
        description: form.description,
        cover_image: form.cover_image,
        start_at: form.dateRange && form.dateRange[0] ? form.dateRange[0] : null,
        end_at: form.dateRange && form.dateRange[1] ? form.dateRange[1] : null,
        status: form.status,
        config: config
      }
      if (isEdit.value) {
        await updateMarketingActivity(form.id, payload)
        ElMessage.success('更新成功')
      } else {
        await createMarketingActivity(payload)
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
  const action = status === 1 ? '启用' : '禁用'
  ElMessageBox.confirm(`确认${action}活动「${row.title}」？`, '提示', {
    type: 'warning',
    confirmButtonText: '确认',
    cancelButtonText: '取消'
  }).then(async () => {
    try {
      await updateMarketingActivityStatus(row.id, { status })
      ElMessage.success(`${action}成功`)
      loadList()
      loadStats()
    } catch (e) {}
  }).catch(() => {})
}

const onDelete = (row) => {
  ElMessageBox.confirm(`确认删除活动「${row.title}」？`, '提示', {
    type: 'warning',
    confirmButtonText: '确认',
    cancelButtonText: '取消'
  }).then(async () => {
    try {
      await deleteMarketingActivity(row.id)
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
.text-success { color: #67c23a; }
.text-warning { color: #e6a23c; }
.text-danger { color: #f56c6c; }
.text-muted { color: #909399; }
.form-tip { font-size: 12px; color: #909399; line-height: 1.4; margin-top: 4px; }
</style>
