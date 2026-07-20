<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">资料总数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.completeRate }}%</div><div class="stat-label">完整率</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.stepAvg }}</div><div class="stat-label">平均步骤</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.todayNew }}</div><div class="stat-label">今日新增</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ stats.verified }}</div><div class="stat-label">实名用户</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-danger">{{ stats.incomplete }}</div><div class="stat-label">未完成</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="会员ID/昵称" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="资料步骤">
          <el-select v-model="filters.step" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="步骤1" :value="1" />
            <el-option label="步骤2" :value="2" />
            <el-option label="步骤3" :value="3" />
            <el-option label="步骤4" :value="4" />
            <el-option label="步骤5" :value="5" />
          </el-select>
        </el-form-item>
        <el-form-item label="是否实名">
          <el-select v-model="filters.id_verified" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="已实名" :value="1" />
            <el-option label="未实名" :value="0" />
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

      <div class="toolbar">
        <el-button :icon="Refresh" @click="loadList">刷新</el-button>
        <el-button type="primary" :icon="Plus" @click="openCreate">新建资料</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe @sort-change="onSortChange">
        <el-table-column prop="id" label="ID" width="70" sortable="custom" />
        <el-table-column prop="love_id" label="会员ID" width="90" />
        <el-table-column label="昵称" min-width="140">
          <template #default="{ row }">
            <el-link type="primary" :underline="'never'" @click="openDetail(row)">{{ row.nickname || `会员#${row.love_id}` }}</el-link>
          </template>
        </el-table-column>
        <el-table-column label="籍贯" width="120">
          <template #default="{ row }">{{ row.hometown || '-' }}</template>
        </el-table-column>
        <el-table-column label="行业" width="120">
          <template #default="{ row }">{{ row.industry || '-' }}</template>
        </el-table-column>
        <el-table-column label="公司" width="140">
          <template #default="{ row }">{{ row.company || '-' }}</template>
        </el-table-column>
        <el-table-column label="MBTI" width="80">
          <template #default="{ row }">{{ row.mbti || '-' }}</template>
        </el-table-column>
        <el-table-column label="星座" width="90">
          <template #default="{ row }">{{ row.zodiac || '-' }}</template>
        </el-table-column>
        <el-table-column label="房产" width="90">
          <template #default="{ row }">
            <el-tag v-if="row.house_status === 'owned'" type="success" size="small">有</el-tag>
            <el-tag v-else-if="row.house_status === 'mortgage'" type="warning" size="small">按揭</el-tag>
            <el-tag v-else-if="row.house_status === 'rent'" type="info" size="small">租房</el-tag>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="车辆" width="90">
          <template #default="{ row }">
            <el-tag v-if="row.car_status === 'owned'" type="success" size="small">有</el-tag>
            <el-tag v-else-if="row.car_status === 'none'" type="info" size="small">无</el-tag>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="步骤" width="90">
          <template #default="{ row }">
            <el-tag :type="row.step >= 5 ? 'success' : 'warning'" size="small">第{{ row.step || 0 }}步</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="完成度" width="120">
          <template #default="{ row }">
            <el-progress :percentage="calcComplete(row)" :stroke-width="6" />
          </template>
        </el-table-column>
        <el-table-column label="更新时间" width="160" prop="updated_at" sortable="custom">
          <template #default="{ row }">{{ formatTime(row.updated_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button type="warning" link size="small" @click="openEdit(row)">编辑</el-button>
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
    <el-dialog v-model="detailVisible" title="资料详情" width="700px">
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="会员ID">{{ detail.love_id }}</el-descriptions-item>
        <el-descriptions-item label="籍贯">{{ detail.hometown || '-' }}</el-descriptions-item>
        <el-descriptions-item label="行业">{{ detail.industry || '-' }}</el-descriptions-item>
        <el-descriptions-item label="公司">{{ detail.company || '-' }}</el-descriptions-item>
        <el-descriptions-item label="职位">{{ detail.job_title || '-' }}</el-descriptions-item>
        <el-descriptions-item label="房产">{{ houseText(detail.house_status) }}</el-descriptions-item>
        <el-descriptions-item label="车辆">{{ carText(detail.car_status) }}</el-descriptions-item>
        <el-descriptions-item label="是否吸烟">{{ boolText(detail.smoking) }}</el-descriptions-item>
        <el-descriptions-item label="是否饮酒">{{ boolText(detail.drinking) }}</el-descriptions-item>
        <el-descriptions-item label="是否要小孩">{{ boolText(detail.want_children) }}</el-descriptions-item>
        <el-descriptions-item label="子女数">{{ detail.children_count || 0 }}</el-descriptions-item>
        <el-descriptions-item label="MBTI">{{ detail.mbti || '-' }}</el-descriptions-item>
        <el-descriptions-item label="星座">{{ detail.zodiac || '-' }}</el-descriptions-item>
        <el-descriptions-item label="血型">{{ detail.blood_type || '-' }}</el-descriptions-item>
        <el-descriptions-item label="资料步骤">第 {{ detail.step || 0 }} 步</el-descriptions-item>
        <el-descriptions-item v-if="detail.interests" label="兴趣爱好" :span="2">{{ detail.interests }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.hobbies" label="特长" :span="2">{{ detail.hobbies }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
      </el-descriptions>
      <template #footer><el-button @click="detailVisible = false">关闭</el-button></template>
    </el-dialog>

    <!-- 编辑弹窗 -->
    <el-dialog v-model="formVisible" :title="formTitle" width="640px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="会员ID" prop="love_id">
          <el-input-number v-model="form.love_id" :min="1" :disabled="isEdit" style="width: 100%" />
        </el-form-item>
        <el-form-item label="籍贯">
          <el-input v-model="form.hometown" maxlength="64" />
        </el-form-item>
        <el-form-item label="行业">
          <el-input v-model="form.industry" maxlength="64" />
        </el-form-item>
        <el-form-item label="公司">
          <el-input v-model="form.company" maxlength="128" />
        </el-form-item>
        <el-form-item label="职位">
          <el-input v-model="form.job_title" maxlength="64" />
        </el-form-item>
        <el-form-item label="房产">
          <el-select v-model="form.house_status" clearable style="width: 100%">
            <el-option label="有" value="owned" />
            <el-option label="按揭" value="mortgage" />
            <el-option label="租房" value="rent" />
          </el-select>
        </el-form-item>
        <el-form-item label="车辆">
          <el-select v-model="form.car_status" clearable style="width: 100%">
            <el-option label="有" value="owned" />
            <el-option label="无" value="none" />
          </el-select>
        </el-form-item>
        <el-form-item label="是否吸烟">
          <el-radio-group v-model="form.smoking">
            <el-radio :value="true">是</el-radio>
            <el-radio :value="false">否</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="是否饮酒">
          <el-radio-group v-model="form.drinking">
            <el-radio :value="true">是</el-radio>
            <el-radio :value="false">否</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="是否要小孩">
          <el-radio-group v-model="form.want_children">
            <el-radio :value="true">是</el-radio>
            <el-radio :value="false">否</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="MBTI">
          <el-select v-model="form.mbti" clearable filterable allow-create style="width: 100%">
            <el-option v-for="m in mbtiList" :key="m" :label="m" :value="m" />
          </el-select>
        </el-form-item>
        <el-form-item label="星座">
          <el-select v-model="form.zodiac" clearable style="width: 100%">
            <el-option v-for="z in zodiacList" :key="z" :label="z" :value="z" />
          </el-select>
        </el-form-item>
        <el-form-item label="兴趣爱好">
          <el-input v-model="form.interests" type="textarea" :rows="3" maxlength="500" />
        </el-form-item>
        <el-form-item label="资料步骤">
          <el-input-number v-model="form.step" :min="0" :max="5" />
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
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search, Plus } from '@element-plus/icons-vue'
import request from '@/utils/request'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const sortField = ref('updated_at')
const sortOrder = ref('descending')

const filters = reactive({ keyword: '', step: null, id_verified: null, dateRange: null })

const stats = reactive({ total: 0, completeRate: 0, stepAvg: 0, todayNew: 0, verified: 0, incomplete: 0 })

const mbtiList = ['INTJ', 'INTP', 'ENTJ', 'ENTP', 'INFJ', 'INFP', 'ENFJ', 'ENFP', 'ISTJ', 'ISFJ', 'ESTJ', 'ESFJ', 'ISTP', 'ISFP', 'ESTP', 'ESFP']
const zodiacList = ['白羊座', '金牛座', '双子座', '巨蟹座', '狮子座', '处女座', '天秤座', '天蝎座', '射手座', '摩羯座', '水瓶座', '双鱼座']

const houseText = (s) => ({ owned: '有', mortgage: '按揭', rent: '租房' }[s] || '-')
const carText = (s) => ({ owned: '有', none: '无' }[s] || '-')
const boolText = (v) => v === true ? '是' : (v === false ? '否' : '-')

const calcComplete = (row) => {
  let count = 0
  const fields = ['hometown', 'industry', 'company', 'job_title', 'house_status', 'car_status', 'mbti', 'zodiac', 'blood_type', 'interests']
  fields.forEach((f) => { if (row[f]) count++ })
  return Math.round((count / fields.length) * 100)
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.keyword = ''; filters.step = null; filters.id_verified = null; filters.dateRange = null
  page.value = 1; loadList()
}

const onSortChange = ({ prop, order }) => {
  sortField.value = prop || 'updated_at'
  sortOrder.value = order || 'descending'
  loadList()
}

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      keyword: filters.keyword.trim() || undefined,
      step: filters.step || undefined,
      id_verified: filters.id_verified === null || filters.id_verified === '' ? undefined : filters.id_verified,
      sort: sortField.value,
      order: sortOrder.value
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await request.get('/love/profiles', { params })
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
  const total = list.value.length
  const complete = list.value.filter((r) => calcComplete(r) >= 80).length
  const stepSum = list.value.reduce((s, r) => s + (r.step || 0), 0)
  const verified = list.value.filter((r) => r.id_verified).length
  const today = new Date().toISOString().slice(0, 10)
  const todayNew = list.value.filter((r) => (r.created_at || '').startsWith(today)).length
  stats.total = total
  stats.completeRate = total ? Math.round((complete / total) * 100) : 0
  stats.stepAvg = total ? (stepSum / total).toFixed(1) : '0'
  stats.verified = verified
  stats.todayNew = todayNew
  stats.incomplete = total - complete
}

// ===== 详情 =====
const detailVisible = ref(false)
const detail = ref(null)
const openDetail = (row) => { detail.value = row; detailVisible.value = true }

// ===== 新建/编辑 =====
const formVisible = ref(false)
const formLoading = ref(false)
const formRef = ref(null)
const isEdit = ref(false)
const formTitle = computed(() => isEdit.value ? '编辑资料' : '新建资料')
const form = reactive({
  id: null, love_id: null, hometown: '', industry: '', company: '', job_title: '',
  house_status: '', car_status: '', smoking: false, drinking: false,
  want_children: false, mbti: '', zodiac: '', blood_type: '', interests: '', step: 0
})
const rules = {
  love_id: [{ required: true, message: '请输入会员ID', trigger: 'blur' }]
}

const openCreate = () => {
  isEdit.value = false
  Object.assign(form, {
    id: null, love_id: null, hometown: '', industry: '', company: '', job_title: '',
    house_status: '', car_status: '', smoking: false, drinking: false,
    want_children: false, mbti: '', zodiac: '', blood_type: '', interests: '', step: 0
  })
  formVisible.value = true
}

const openEdit = (row) => {
  isEdit.value = true
  Object.assign(form, {
    id: row.id, love_id: row.love_id, hometown: row.hometown || '', industry: row.industry || '',
    company: row.company || '', job_title: row.job_title || '', house_status: row.house_status || '',
    car_status: row.car_status || '', smoking: !!row.smoking, drinking: !!row.drinking,
    want_children: !!row.want_children, mbti: row.mbti || '', zodiac: row.zodiac || '',
    blood_type: row.blood_type || '', interests: row.interests || '', step: row.step || 0
  })
  formVisible.value = true
}

const onSubmit = async () => {
  try {
    await formRef.value.validate()
    formLoading.value = true
    if (isEdit.value) {
      await request.put(`/love/profiles/${form.id}`, form)
    } else {
      await request.post('/love/profiles', form)
    }
    ElMessage.success(isEdit.value ? '更新成功' : '创建成功')
    formVisible.value = false
    await loadList()
  } catch (e) {
    // ignore
  } finally {
    formLoading.value = false
  }
}

const onDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确认删除资料（会员ID：${row.love_id}）？`, '提示', { type: 'warning' })
    await request.delete(`/love/profiles/${row.id}`)
    ElMessage.success('删除成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

onMounted(() => loadList())
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-item { text-align: center; padding: 16px; }
.stat-value { font-size: 22px; font-weight: 600; color: #303133; }
.text-success { color: #67c23a; }
.text-warning { color: #e6a23c; }
.text-danger { color: #f56c6c; }
.text-primary { color: #409eff; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; margin-bottom: 12px; }
.text-muted { color: #909399; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
