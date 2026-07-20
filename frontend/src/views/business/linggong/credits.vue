<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #409eff">
            <el-icon :size="22"><Trophy /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.total }}</div>
            <div class="stat-label">变动记录</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #67c23a">
            <el-icon :size="22"><Top /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.plusCount }}</div>
            <div class="stat-label">加分次数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #f56c6c">
            <el-icon :size="22"><Bottom /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.minusCount }}</div>
            <div class="stat-label">扣分次数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #722ed1">
            <el-icon :size="22"><TrendCharts /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.avgScore }}</div>
            <div class="stat-label">平均分</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="用户ID">
          <el-input v-model="filters.user_id" placeholder="用户ID" clearable style="width: 140px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="变动类型">
          <el-select v-model="filters.change_type" placeholder="全部" clearable style="width: 160px" @change="onSearch">
            <el-option v-for="(label, val) in changeTypeMap" :key="val" :label="label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item label="来源">
          <el-select v-model="filters.source" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option v-for="(label, val) in sourceMap" :key="val" :label="label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="原因/操作人" clearable style="width: 200px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar">
        <div class="toolbar-left">
          <el-button :icon="Refresh" @click="loadList">刷新</el-button>
        </div>
        <div class="toolbar-right">
          <el-button type="primary" :icon="Plus" @click="openAdjust">手动调整</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" fixed="left" />
        <el-table-column label="用户" width="160">
          <template #default="{ row }">
            <div>{{ row.user_name || `#${row.user_id}` }}</div>
            <div class="text-muted text-xs">{{ maskPhone(row.user_phone) }}</div>
          </template>
        </el-table-column>
        <el-table-column label="变动类型" width="120">
          <template #default="{ row }">
            <el-tag size="small" :type="changeTagType(row.change_type)">{{ changeTypeMap[row.change_type] || row.change_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="分值" width="100">
          <template #default="{ row }">
            <span :class="row.score_change >= 0 ? 'text-success' : 'text-danger'">
              {{ row.score_change >= 0 ? '+' : '' }}{{ row.score_change }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="变动后" width="100">
          <template #default="{ row }">
            <el-tag :type="creditTagType(row.score_after)" size="small" effect="plain">{{ row.score_after }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="来源" width="120">
          <template #default="{ row }">
            <el-tag size="small" effect="plain" :type="sourceTagType(row.source)">{{ sourceMap[row.source] || row.source }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="原因" min-width="220" show-overflow-tooltip>
          <template #default="{ row }">{{ row.reason || row.description || '-' }}</template>
        </el-table-column>
        <el-table-column label="操作人" width="120">
          <template #default="{ row }">{{ row.operator_name || `#${row.operator_id}` }}</template>
        </el-table-column>
        <el-table-column label="时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
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

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="信用变动详情" width="640px">
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="记录ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="用户">{{ detail.user_name || `#${detail.user_id}` }}</el-descriptions-item>
        <el-descriptions-item label="变动类型">
          <el-tag size="small" :type="changeTagType(detail.change_type)">{{ changeTypeMap[detail.change_type] || detail.change_type }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="来源">
          <el-tag size="small" effect="plain" :type="sourceTagType(detail.source)">{{ sourceMap[detail.source] || detail.source }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="变动分值">
          <span :class="detail.score_change >= 0 ? 'text-success' : 'text-danger'">
            {{ detail.score_change >= 0 ? '+' : '' }}{{ detail.score_change }}
          </span>
        </el-descriptions-item>
        <el-descriptions-item label="变动后分值">
          <el-tag :type="creditTagType(detail.score_after)" size="small">{{ detail.score_after }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="变动前分值">{{ detail.score_before }}</el-descriptions-item>
        <el-descriptions-item label="关联ID">{{ detail.ref_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="原因" :span="2">{{ detail.reason || detail.description || '-' }}</el-descriptions-item>
        <el-descriptions-item label="操作人">{{ detail.operator_name || `#${detail.operator_id}` }}</el-descriptions-item>
        <el-descriptions-item label="时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 手动调整弹窗 -->
    <el-dialog v-model="adjustVisible" title="手动调整信用分" width="540px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="用户ID" prop="user_id">
          <el-input-number v-model="form.user_id" :min="1" :controls="false" placeholder="用户ID" style="width: 100%" />
        </el-form-item>
        <el-form-item label="变动类型" prop="change_type">
          <el-select v-model="form.change_type" style="width: 100%">
            <el-option v-for="(label, val) in changeTypeMap" :key="val" :label="label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item label="分值变动" prop="score_change">
          <el-input-number v-model="form.score_change" :step="10" :min="-100" :max="100" style="width: 100%" />
          <div class="form-tip">正数加分，负数扣分</div>
        </el-form-item>
        <el-form-item label="来源" prop="source">
          <el-select v-model="form.source" style="width: 100%">
            <el-option v-for="(label, val) in sourceMap" :key="val" :label="label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item label="关联ID">
          <el-input-number v-model="form.ref_id" :min="0" :controls="false" placeholder="可选" style="width: 100%" />
        </el-form-item>
        <el-form-item label="原因" prop="reason">
          <el-input v-model="form.reason" type="textarea" :rows="3" placeholder="请输入调整原因" maxlength="500" show-word-limit />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="adjustVisible = false">取消</el-button>
        <el-button type="primary" :loading="formLoading" @click="onSubmit">确认调整</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Trophy, Top, Bottom, TrendCharts, Refresh, RefreshLeft, Search, Plus } from '@element-plus/icons-vue'
import {
  listLinggongCredits, getLinggongCredit, adjustLinggongCredit, deleteLinggongCredit
} from '@/api/linggong'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const filters = reactive({ user_id: '', change_type: '', source: '', keyword: '' })

const changeTypeMap = {
  complete_task: '完成任务', good_rating: '好评奖励', on_time: '准时结算',
  violation: '违规扣分', bad_rating: '差评扣分', complaint: '投诉扣分',
  cancel_contract: '违约扣分', manual: '手动调整', register: '注册赠送', verify: '认证奖励'
}
const changeTagType = (t) => ({
  complete_task: 'success', good_rating: 'success', on_time: 'success',
  violation: 'danger', bad_rating: 'danger', complaint: 'danger',
  cancel_contract: 'danger', manual: 'warning', register: 'info', verify: 'success'
}[t] || 'info')

const sourceMap = {
  system: '系统自动', admin: '管理员', employer: '雇主', worker: '工人', platform: '平台'
}
const sourceTagType = (s) => ({
  system: 'info', admin: 'warning', employer: '', worker: 'success', platform: 'danger'
}[s] || 'info')

const creditTagType = (score) => {
  const s = Number(score || 0)
  if (s >= 90) return 'success'
  if (s >= 70) return 'primary'
  if (s >= 50) return 'warning'
  return 'danger'
}

const maskPhone = (p) => {
  if (!p) return '-'
  const s = String(p)
  return s.length >= 11 ? s.slice(0, 3) + '****' + s.slice(-4) : s
}

const stats = computed(() => {
  const total = list.value.length
  const plusCount = list.value.filter((r) => Number(r.score_change) >= 0).length
  const minusCount = list.value.filter((r) => Number(r.score_change) < 0).length
  const scoreList = list.value.filter((r) => r.score_after !== undefined && r.score_after !== null)
  const avgScore = scoreList.length
    ? (scoreList.reduce((s, r) => s + Number(r.score_after || 0), 0) / scoreList.length).toFixed(1)
    : 0
  return { total, plusCount, minusCount, avgScore }
})

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.user_id = ''
  filters.change_type = ''
  filters.source = ''
  filters.keyword = ''
  page.value = 1
  loadList()
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await listLinggongCredits({
      page: page.value,
      page_size: pageSize.value,
      user_id: filters.user_id || undefined,
      change_type: filters.change_type || undefined,
      source: filters.source || undefined,
      keyword: filters.keyword || undefined
    })
    const data = res.data || {}
    list.value = data.list || data || []
    total.value = data.total || list.value.length
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
    const res = await getLinggongCredit(row.id)
    detail.value = res.data || row
    detailVisible.value = true
  } catch (e) {
    detail.value = row
    detailVisible.value = true
  }
}

const adjustVisible = ref(false)
const formLoading = ref(false)
const formRef = ref(null)
const form = reactive({
  user_id: null,
  change_type: 'manual',
  score_change: 10,
  source: 'admin',
  ref_id: 0,
  reason: ''
})
const rules = {
  user_id: [{ required: true, message: '请输入用户ID', trigger: 'blur' }],
  change_type: [{ required: true, message: '请选择变动类型', trigger: 'change' }],
  score_change: [{ required: true, message: '请输入分值变动', trigger: 'blur' }],
  source: [{ required: true, message: '请选择来源', trigger: 'change' }],
  reason: [{ required: true, message: '请输入调整原因', trigger: 'blur' }]
}

const openAdjust = () => {
  Object.assign(form, {
    user_id: null,
    change_type: 'manual',
    score_change: 10,
    source: 'admin',
    ref_id: 0,
    reason: ''
  })
  adjustVisible.value = true
}

const onSubmit = async () => {
  try {
    await formRef.value.validate()
    formLoading.value = true
    await adjustLinggongCredit({
      user_id: form.user_id,
      change_type: form.change_type,
      score_change: form.score_change,
      source: form.source,
      ref_id: form.ref_id || 0,
      reason: form.reason
    })
    ElMessage.success('调整成功')
    adjustVisible.value = false
    await loadList()
  } catch (e) {
    // 校验失败或接口失败
  } finally {
    formLoading.value = false
  }
}

const onDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确认删除该信用变动记录？删除后不可恢复', '提示', { type: 'warning' })
    await deleteLinggongCredit(row.id)
    ElMessage.success('删除成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

onMounted(() => loadList())
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-card { display: flex; align-items: center; }
.stat-card :deep(.el-card__body) { display: flex; align-items: center; width: 100%; padding: 16px; }
.stat-icon { width: 48px; height: 48px; border-radius: 8px; color: #fff; display: flex; align-items: center; justify-content: center; margin-right: 12px; flex-shrink: 0; }
.stat-content { flex: 1; }
.stat-value { font-size: 22px; font-weight: 600; color: #303133; }
.stat-label { font-size: 12px; color: #909399; margin-top: 2px; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; margin-bottom: 12px; }
.toolbar-left, .toolbar-right { display: flex; gap: 8px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.text-muted { color: #909399; }
.text-xs { font-size: 12px; }
.text-success { color: #67c23a; font-weight: 600; }
.text-danger { color: #f56c6c; font-weight: 600; }
.form-tip { color: #909399; font-size: 12px; line-height: 1.5; margin-top: 4px; }
</style>
