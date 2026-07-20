<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #409eff">
            <el-icon :size="22"><Document /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.total }}</div>
            <div class="stat-label">规则总数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #67c23a">
            <el-icon :size="22"><CircleCheckFilled /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.active }}</div>
            <div class="stat-label">启用中</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #e6a23c">
            <el-icon :size="22"><Warning /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.disabled }}</div>
            <div class="stat-label">已禁用</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #722ed1">
            <el-icon :size="22"><Filter /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.hitToday }}</div>
            <div class="stat-label">今日命中</div>
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
            placeholder="规则名称/描述"
            clearable
            style="width: 220px"
            :prefix-icon="Search"
            @keyup.enter="onSearch"
          />
        </el-form-item>
        <el-form-item label="业务模块">
          <el-select v-model="filters.module" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option label="拼车发布" value="pinche" />
            <el-option label="路线管理" value="route" />
            <el-option label="车主认证" value="driver" />
            <el-option label="车辆审核" value="vehicle" />
            <el-option label="行程管理" value="trip" />
            <el-option label="预订管理" value="booking" />
          </el-select>
        </el-form-item>
        <el-form-item label="规则类型">
          <el-select v-model="filters.rule_type" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option label="关键词过滤" value="keyword" />
            <el-option label="价格校验" value="price" />
            <el-option label="次数限制" value="count_limit" />
            <el-option label="信用分限制" value="credit_score" />
            <el-option label="黑名单" value="blacklist" />
            <el-option label="自定义" value="custom" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
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
          <el-button type="success" :icon="Check" :disabled="!selection.length" @click="onBatchEnable">批量启用</el-button>
          <el-button type="warning" :icon="Close" :disabled="!selection.length" @click="onBatchDisable">批量禁用</el-button>
          <el-button type="danger" :icon="Delete" :disabled="!selection.length" @click="onBatchDelete">批量删除</el-button>
        </div>
        <div class="toolbar-right">
          <el-button type="primary" :icon="Plus" @click="onCreate">新建规则</el-button>
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
        <el-table-column prop="name" label="规则名称" min-width="180">
          <template #default="{ row }">
            <el-link type="primary" :underline="'never'" @click="openEdit(row)">{{ row.name }}</el-link>
          </template>
        </el-table-column>
        <el-table-column label="业务模块" width="120">
          <template #default="{ row }">
            <el-tag size="small">{{ moduleText(row.module) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="规则类型" width="130">
          <template #default="{ row }">
            <el-tag :type="ruleTypeTagType(row.rule_type)" size="small" effect="plain">{{ ruleTypeText(row.rule_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="规则内容" min-width="280">
          <template #default="{ row }">
            <div class="rule-content">{{ row.content || row.pattern || '-' }}</div>
            <div v-if="row.description" class="text-muted">{{ row.description }}</div>
          </template>
        </el-table-column>
        <el-table-column label="优先级" width="90">
          <template #default="{ row }">
            <el-rate
              v-model="row.priority"
              disabled
              :max="5"
              :colors="['#909399', '#e6a23c', '#f56c6c']"
            />
          </template>
        </el-table-column>
        <el-table-column label="命中次数" width="100" prop="hit_count" sortable />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-switch
              v-model="row.status"
              :active-value="1"
              :inactive-value="0"
              @change="onToggleStatus(row)"
            />
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160" prop="created_at" sortable>
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openEdit(row)">编辑</el-button>
            <el-button type="warning" link size="small" @click="onTestRule(row)">测试</el-button>
            <el-button type="danger" link size="small" @click="onDelete(row)">删除</el-button>
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

    <!-- 新建/编辑弹窗 -->
    <el-dialog
      v-model="formVisible"
      :title="form.id ? '编辑审核规则' : '新建审核规则'"
      width="700px"
      destroy-on-close
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
        <el-form-item label="规则名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入规则名称" maxlength="50" show-word-limit />
        </el-form-item>
        <el-form-item label="业务模块" prop="module">
          <el-select v-model="form.module" placeholder="请选择业务模块" style="width: 100%">
            <el-option label="拼车发布" value="pinche" />
            <el-option label="路线管理" value="route" />
            <el-option label="车主认证" value="driver" />
            <el-option label="车辆审核" value="vehicle" />
            <el-option label="行程管理" value="trip" />
            <el-option label="预订管理" value="booking" />
          </el-select>
        </el-form-item>
        <el-form-item label="规则类型" prop="rule_type">
          <el-select v-model="form.rule_type" placeholder="请选择规则类型" style="width: 100%">
            <el-option label="关键词过滤" value="keyword" />
            <el-option label="价格校验" value="price" />
            <el-option label="次数限制" value="count_limit" />
            <el-option label="信用分限制" value="credit_score" />
            <el-option label="黑名单" value="blacklist" />
            <el-option label="自定义" value="custom" />
          </el-select>
        </el-form-item>
        <el-form-item label="规则内容" prop="content">
          <el-input
            v-model="form.content"
            type="textarea"
            :rows="4"
            placeholder="关键词以逗号分隔；正则/表达式可直接填写；价格/次数可填入具体值"
          />
        </el-form-item>
        <el-form-item label="优先级" prop="priority">
          <el-rate
            v-model="form.priority"
            :max="5"
            :colors="['#909399', '#e6a23c', '#f56c6c']"
            show-text
            :texts="['最低', '低', '中', '高', '最高']"
          />
        </el-form-item>
        <el-form-item label="动作" prop="action">
          <el-radio-group v-model="form.action">
            <el-radio value="block">拦截</el-radio>
            <el-radio value="review">转人工审核</el-radio>
            <el-radio value="warn">仅警告</el-radio>
            <el-radio value="log">仅记录</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="描述">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="2"
            placeholder="规则说明（选填）"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm">保存</el-button>
      </template>
    </el-dialog>

    <!-- 测试弹窗 -->
    <el-dialog v-model="testVisible" title="测试规则" width="600px" destroy-on-close>
      <el-form label-width="100px">
        <el-form-item label="规则名称">
          <span>{{ testRule.name }}</span>
        </el-form-item>
        <el-form-item label="测试输入">
          <el-input
            v-model="testInput"
            type="textarea"
            :rows="4"
            placeholder="请输入需要测试的内容（如发布内容、价格、电话等）"
          />
        </el-form-item>
        <el-form-item label="测试结果">
          <el-tag v-if="testResult === true" type="success" size="large">命中规则</el-tag>
          <el-tag v-else-if="testResult === false" type="info" size="large">未命中</el-tag>
          <span v-else class="text-muted">点击「执行测试」查看结果</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="testVisible = false">关闭</el-button>
        <el-button type="primary" @click="runTest">执行测试</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/utils/request'
import {
  Refresh, RefreshLeft, Search, Plus, Delete, Check, Close,
  Document, CircleCheckFilled, Warning, Filter
} from '@element-plus/icons-vue'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const selection = ref([])

const stats = reactive({ total: 0, active: 0, disabled: 0, hitToday: 0 })

const filters = reactive({
  keyword: '', module: '', rule_type: '', status: null
})

const moduleText = (m) => ({
  pinche: '拼车发布', route: '路线管理', driver: '车主认证',
  vehicle: '车辆审核', trip: '行程管理', booking: '预订管理'
}[m] || m || '-')
const ruleTypeText = (t) => ({
  keyword: '关键词过滤', price: '价格校验', count_limit: '次数限制',
  credit_score: '信用分限制', blacklist: '黑名单', custom: '自定义'
}[t] || '-')
const ruleTypeTagType = (t) => ({
  keyword: 'danger', price: 'warning', count_limit: 'info',
  credit_score: 'primary', blacklist: 'danger', custom: 'info'
}[t] || 'info')

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.keyword = ''; filters.module = ''
  filters.rule_type = ''; filters.status = null
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
      module: filters.module || undefined,
      rule_type: filters.rule_type || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status
    }
    const res = await request.get('/pinche/admin/audit-rules', { params })
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
    stats.total = total.value
    stats.active = list.value.filter(i => i.status === 1).length
    stats.disabled = list.value.filter(i => i.status === 0).length
    stats.hitToday = list.value.reduce((s, i) => s + (i.today_hit_count || 0), 0)
  } catch (e) {
    list.value = []; total.value = 0
  } finally {
    loading.value = false
  }
}

const formVisible = ref(false)
const formRef = ref(null)
const form = reactive({
  id: 0, name: '', module: 'pinche', rule_type: 'keyword',
  content: '', priority: 3, action: 'review', status: 1, description: ''
})
const rules = {
  name: [{ required: true, message: '请输入规则名称', trigger: 'blur' }],
  module: [{ required: true, message: '请选择业务模块', trigger: 'change' }],
  rule_type: [{ required: true, message: '请选择规则类型', trigger: 'change' }],
  content: [{ required: true, message: '请输入规则内容', trigger: 'blur' }]
}

const resetForm = () => {
  form.id = 0
  form.name = ''
  form.module = 'pinche'
  form.rule_type = 'keyword'
  form.content = ''
  form.priority = 3
  form.action = 'review'
  form.status = 1
  form.description = ''
}

const onCreate = () => {
  resetForm()
  formVisible.value = true
}

const openEdit = (row) => {
  resetForm()
  Object.assign(form, row)
  formVisible.value = true
}

const submitForm = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    try {
      if (form.id) {
        await request.put(`/pinche/admin/audit-rules/${form.id}`, { ...form })
        ElMessage.success('更新成功')
      } else {
        const { id, ...payload } = form
        await request.post('/pinche/admin/audit-rules', { ...payload })
        ElMessage.success('创建成功')
      }
      formVisible.value = false
      await loadList()
    } catch (e) { /* ignore */ }
  })
}

const onToggleStatus = async (row) => {
  try {
    await request.put(`/pinche/admin/audit-rules/${row.id}`, { status: row.status })
    ElMessage.success(row.status === 1 ? '已启用' : '已禁用')
  } catch (e) {
    row.status = row.status === 1 ? 0 : 1
  }
}

const onDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确认删除规则 "${row.name}"？`, '提示', { type: 'warning' })
    await request.delete(`/pinche/admin/audit-rules/${row.id}`)
    ElMessage.success('删除成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onBatchEnable = async () => {
  if (!selection.value.length) return
  try {
    await ElMessageBox.confirm(`确认批量启用 ${selection.value.length} 条规则？`, '提示')
    await Promise.all(selection.value.map(i =>
      request.put(`/pinche/admin/audit-rules/${i.id}`, { status: 1 })
    ))
    ElMessage.success('批量启用成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onBatchDisable = async () => {
  if (!selection.value.length) return
  try {
    await ElMessageBox.confirm(`确认批量禁用 ${selection.value.length} 条规则？`, '提示', { type: 'warning' })
    await Promise.all(selection.value.map(i =>
      request.put(`/pinche/admin/audit-rules/${i.id}`, { status: 0 })
    ))
    ElMessage.success('批量禁用成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onBatchDelete = async () => {
  if (!selection.value.length) return
  try {
    await ElMessageBox.confirm(`确认批量删除 ${selection.value.length} 条规则？此操作不可撤销`, '提示', { type: 'warning' })
    await Promise.all(selection.value.map(i =>
      request.delete(`/pinche/admin/audit-rules/${i.id}`)
    ))
    ElMessage.success('批量删除成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

const testVisible = ref(false)
const testRule = reactive({ id: 0, name: '', content: '', rule_type: '' })
const testInput = ref('')
const testResult = ref(null)

const onTestRule = (row) => {
  testRule.id = row.id
  testRule.name = row.name
  testRule.content = row.content || ''
  testRule.rule_type = row.rule_type
  testInput.value = ''
  testResult.value = null
  testVisible.value = true
}

const runTest = () => {
  if (!testInput.value) {
    ElMessage.warning('请输入测试内容')
    return
  }
  // 简单本地模拟：根据规则类型判断
  const content = testRule.content
  const input = testInput.value
  let hit = false
  if (testRule.rule_type === 'keyword') {
    const keywords = content.split(/[,，]/).map(k => k.trim()).filter(Boolean)
    hit = keywords.some(k => input.includes(k))
  } else if (testRule.rule_type === 'price') {
    const maxPrice = parseFloat(content)
    const inputPrice = parseFloat(input)
    if (!isNaN(maxPrice) && !isNaN(inputPrice)) hit = inputPrice < maxPrice || inputPrice > maxPrice * 10
  } else if (testRule.rule_type === 'blacklist') {
    const items = content.split(/[,，\n]/).map(k => k.trim()).filter(Boolean)
    hit = items.includes(input)
  } else {
    // 自定义规则尝试正则匹配
    try {
      const regex = new RegExp(content)
      hit = regex.test(input)
    } catch (e) {
      hit = input.includes(content)
    }
  }
  testResult.value = hit
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

.rule-content {
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 60px;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}
.text-muted { color: #909399; font-size: 12px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
