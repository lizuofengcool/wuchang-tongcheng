<template>
  <div class="app-container">
    <!-- 统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">总电话量</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.today }}</div><div class="stat-label">今日电话</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.uniqueCallers }}</div><div class="stat-label">独立呼叫人</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ stats.uniqueBusiness }}</div><div class="stat-label">被叫商户数</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="商户ID">
          <el-input v-model="filters.dh114_id" placeholder="商户ID" clearable style="width: 140px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="呼叫人">
          <el-input v-model="filters.caller" placeholder="呼叫人/手机号" clearable style="width: 180px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="呼叫类型">
          <el-select v-model="filters.call_type" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option label="电话" value="phone" />
            <el-option label="在线咨询" value="chat" />
            <el-option label="微信" value="wechat" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="成功" value="success" />
            <el-option label="失败" value="failed" />
            <el-option label="未接" value="noanswer" />
          </el-select>
        </el-form-item>
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="filters.dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始"
            end-placeholder="结束"
            value-format="YYYY-MM-DD"
            style="width: 240px"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar">
        <el-button :icon="Refresh" @click="loadList">刷新</el-button>
        <el-button type="primary" :icon="Plus" @click="openCreate">新建记录</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="dh114_id" label="商户ID" width="90" />
        <el-table-column label="商户名" min-width="160">
          <template #default="{ row }">{{ row.business_name || `商户#${row.dh114_id}` }}</template>
        </el-table-column>
        <el-table-column label="呼叫人" width="140">
          <template #default="{ row }">
            <div>{{ row.caller_name || `用户#${row.caller_id}` }}</div>
            <div class="text-muted">{{ maskPhone(row.caller_phone) }}</div>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="callTypeTag(row.call_type)" size="small">{{ callTypeText(row.call_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="时长" width="100">
          <template #default="{ row }">{{ row.duration ? row.duration + '秒' : '-' }}</template>
        </el-table-column>
        <el-table-column label="备注" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ row.remark || '-' }}</template>
        </el-table-column>
        <el-table-column label="呼叫时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
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

    <!-- 新建弹窗 -->
    <el-dialog v-model="formVisible" title="新建电话记录" width="500px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="商户ID" prop="dh114_id">
          <el-input-number v-model="form.dh114_id" :min="1" :controls="false" style="width: 100%" />
        </el-form-item>
        <el-form-item label="呼叫类型">
          <el-select v-model="form.call_type" style="width: 100%">
            <el-option label="电话" value="phone" />
            <el-option label="在线咨询" value="chat" />
            <el-option label="微信" value="wechat" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status" style="width: 100%">
            <el-option label="成功" value="success" />
            <el-option label="失败" value="failed" />
            <el-option label="未接" value="noanswer" />
          </el-select>
        </el-form-item>
        <el-form-item label="时长(秒)">
          <el-input-number v-model="form.duration" :min="0" :controls="false" style="width: 100%" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="2" maxlength="200" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="formLoading" @click="onSubmit">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search, Plus } from '@element-plus/icons-vue'
import request from '@/utils/request'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const stats = reactive({ total: 0, today: 0, uniqueCallers: 0, uniqueBusiness: 0 })

const filters = reactive({ dh114_id: '', caller: '', call_type: '', status: '', dateRange: null })

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.dh114_id = ''
  filters.caller = ''
  filters.call_type = ''
  filters.status = ''
  filters.dateRange = null
  page.value = 1
  loadList()
}

const callTypeText = (t) => ({ phone: '电话', chat: '咨询', wechat: '微信' }[t] || '-')
const callTypeTag = (t) => ({ phone: 'primary', chat: 'success', wechat: 'success' }[t] || 'info')
const statusText = (s) => ({ success: '成功', failed: '失败', noanswer: '未接' }[s] || '-')
const statusTag = (s) => ({ success: 'success', failed: 'danger', noanswer: 'warning' }[s] || 'info')

const maskPhone = (p) => {
  if (!p) return '-'
  const s = String(p)
  if (s.length < 7) return s
  return s.slice(0, 3) + '****' + s.slice(-4)
}

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      dh114_id: filters.dh114_id || undefined,
      caller: filters.caller || undefined,
      call_type: filters.call_type || undefined,
      status: filters.status || undefined
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await request.get('/dh114/admin/phone-calls', { params })
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
    calcStats()
  } catch (e) {
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const calcStats = () => {
  const today = new Date().toISOString().slice(0, 10)
  const total = list.value.length
  const todayCount = list.value.filter((r) => r.created_at && r.created_at.slice(0, 10) === today).length
  const callers = new Set(list.value.map((r) => r.caller_id || r.caller_phone).filter(Boolean)).size
  const business = new Set(list.value.map((r) => r.dh114_id).filter(Boolean)).size
  Object.assign(stats, { total, today: todayCount, uniqueCallers: callers, uniqueBusiness: business })
}

const onDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确认删除该电话记录？', '提示', { type: 'warning' })
    await request.delete(`/dh114/phone-calls/${row.id}`)
    ElMessage.success('删除成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

// ===== 新建 =====
const formVisible = ref(false)
const formLoading = ref(false)
const formRef = ref(null)
const form = reactive({
  dh114_id: undefined, call_type: 'phone', status: 'success', duration: 0, remark: ''
})
const rules = {
  dh114_id: [{ required: true, message: '请输入商户ID', trigger: 'blur' }]
}

const openCreate = () => {
  Object.assign(form, { dh114_id: undefined, call_type: 'phone', status: 'success', duration: 0, remark: '' })
  formVisible.value = true
}

const onSubmit = async () => {
  try {
    await formRef.value.validate()
    formLoading.value = true
    await request.post('/dh114/phone-calls', { ...form })
    ElMessage.success('创建成功')
    formVisible.value = false
    await loadList()
  } catch (e) {
    // 校验或接口失败
  } finally {
    formLoading.value = false
  }
}

onMounted(() => loadList())
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-item { text-align: center; padding: 16px; }
.stat-value { font-size: 22px; font-weight: 600; color: #409eff; }
.text-success { color: #67c23a; }
.text-warning { color: #e6a23c; }
.text-danger { color: #f56c6c; }
.text-primary { color: #409eff; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; margin-bottom: 12px; }
.text-muted { color: #909399; font-size: 12px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
