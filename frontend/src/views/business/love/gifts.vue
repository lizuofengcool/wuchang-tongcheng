<template>
  <div class="app-container">
    <el-tabs v-model="activeTab" class="page-card" @tab-change="onTabChange">
      <el-tab-pane label="礼物列表" name="gifts">
        <!-- 顶部统计卡片 -->
        <el-row :gutter="16" class="stat-row">
          <el-col :xs="12" :sm="8" :md="6">
            <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ giftStats.total }}</div><div class="stat-label">礼物总数</div></div></el-card>
          </el-col>
          <el-col :xs="12" :sm="8" :md="6">
            <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ giftStats.enabled }}</div><div class="stat-label">启用中</div></div></el-card>
          </el-col>
          <el-col :xs="12" :sm="8" :md="6">
            <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ giftStats.totalSent }}</div><div class="stat-label">总送出数</div></div></el-card>
          </el-col>
          <el-col :xs="12" :sm="8" :md="6">
            <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">¥{{ giftStats.totalRevenue }}</div><div class="stat-label">总价值</div></div></el-card>
          </el-col>
        </el-row>

        <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
          <el-form-item label="关键词">
            <el-input v-model="filters.keyword" placeholder="礼物名/编码" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
          </el-form-item>
          <el-form-item label="分类">
            <el-select v-model="filters.category" placeholder="全部" clearable style="width: 140px" @change="onSearch">
              <el-option label="玫瑰" value="rose" />
              <el-option label="礼物" value="gift" />
              <el-option label="道具" value="prop" />
              <el-option label="徽章" value="badge" />
              <el-option label="车辆" value="car" />
              <el-option label="游艇" value="yacht" />
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

        <div class="toolbar">
          <el-button :icon="Refresh" @click="loadList">刷新</el-button>
          <el-button type="primary" :icon="Plus" @click="openCreate">新建礼物</el-button>
        </div>

        <el-table v-loading="loading" :data="list" border stripe @sort-change="onSortChange">
          <el-table-column prop="id" label="ID" width="70" sortable="custom" />
          <el-table-column prop="gift_code" label="编码" width="120" />
          <el-table-column label="图标" width="80">
            <template #default="{ row }">
              <el-image v-if="row.icon" :src="row.icon" fit="cover" class="cover-thumb" :preview-src-list="[row.icon]" preview-teleported />
              <div v-else class="cover-thumb cover-empty">无图</div>
            </template>
          </el-table-column>
          <el-table-column prop="name" label="名称" min-width="140" />
          <el-table-column label="分类" width="100">
            <template #default="{ row }">
              <el-tag size="small">{{ categoryText(row.category) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="价格" width="120">
            <template #default="{ row }">
              <span class="price">{{ row.price }} {{ priceUnitText(row.price_unit) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="折扣价" width="100">
            <template #default="{ row }">
              <span v-if="row.discount_price > 0" class="text-danger">{{ row.discount_price }} {{ priceUnitText(row.price_unit) }}</span>
              <span v-else class="text-muted">-</span>
            </template>
          </el-table-column>
          <el-table-column prop="sent_count" label="送出数" width="90" />
          <el-table-column label="排序" width="80" prop="sort" />
          <el-table-column label="状态" width="90">
            <template #default="{ row }">
              <el-switch :model-value="row.status === 1" @change="(val) => onToggle(row, val)" />
            </template>
          </el-table-column>
          <el-table-column prop="description" label="描述" min-width="160" show-overflow-tooltip />
          <el-table-column label="操作" width="140" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" link size="small" @click="openEdit(row)">编辑</el-button>
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
      </el-tab-pane>

      <el-tab-pane label="礼物记录" name="records">
        <!-- 记录统计 -->
        <el-row :gutter="16" class="stat-row">
          <el-col :xs="12" :sm="8" :md="6">
            <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ recordStats.total }}</div><div class="stat-label">总送出数</div></div></el-card>
          </el-col>
          <el-col :xs="12" :sm="8" :md="6">
            <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ recordStats.today }}</div><div class="stat-label">今日送出</div></div></el-card>
          </el-col>
          <el-col :xs="12" :sm="8" :md="6">
            <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ recordStats.uniqueSenders }}</div><div class="stat-label">送礼人数</div></div></el-card>
          </el-col>
          <el-col :xs="12" :sm="8" :md="6">
            <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">¥{{ recordStats.totalValue }}</div><div class="stat-label">总价值</div></div></el-card>
          </el-col>
        </el-row>

        <el-form :inline="true" :model="recordFilters" class="filter-form" @submit.prevent>
          <el-form-item label="关键词">
            <el-input v-model="recordFilters.keyword" placeholder="会员ID/昵称" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onRecordSearch" />
          </el-form-item>
          <el-form-item label="礼物">
            <el-select v-model="recordFilters.gift_id" placeholder="全部" clearable style="width: 160px" @change="onRecordSearch">
              <el-option v-for="g in gifts" :key="g.id" :label="g.name" :value="g.id" />
            </el-select>
          </el-form-item>
          <el-form-item label="时间范围">
            <el-date-picker
              v-model="recordFilters.dateRange"
              type="daterange"
              range-separator="至"
              start-placeholder="开始日期"
              end-placeholder="结束日期"
              value-format="YYYY-MM-DD"
              style="width: 240px"
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :icon="Search" @click="onRecordSearch">搜索</el-button>
            <el-button :icon="RefreshLeft" @click="onRecordReset">重置</el-button>
          </el-form-item>
        </el-form>

        <div class="toolbar"><el-button :icon="Refresh" @click="loadRecords">刷新</el-button></div>

        <el-table v-loading="recordLoading" :data="records" border stripe @sort-change="onRecordSortChange">
          <el-table-column prop="id" label="ID" width="70" sortable="custom" />
          <el-table-column label="送出者" min-width="160">
            <template #default="{ row }">
              <div class="user-cell">
                <el-image v-if="row.from_avatar" :src="row.from_avatar" fit="cover" class="user-avatar" />
                <div class="user-info">
                  <div class="user-name">{{ row.from_name || `#${row.from_user_id}` }}</div>
                </div>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="接收者" min-width="160">
            <template #default="{ row }">
              <div class="user-cell">
                <el-image v-if="row.to_avatar" :src="row.to_avatar" fit="cover" class="user-avatar" />
                <div class="user-info">
                  <div class="user-name">{{ row.to_name || `#${row.to_user_id}` }}</div>
                </div>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="礼物" width="160">
            <template #default="{ row }">
              <div class="gift-cell">
                <el-image v-if="row.gift_icon" :src="row.gift_icon" fit="cover" class="gift-icon" />
                <span>{{ row.gift_name }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="数量" width="80" prop="count" />
          <el-table-column label="总价值" width="120">
            <template #default="{ row }">
              <span class="price">{{ row.total_price }} {{ priceUnitText(row.price_unit) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="留言" min-width="180" show-overflow-tooltip>
            <template #default="{ row }">{{ row.message || '-' }}</template>
          </el-table-column>
          <el-table-column label="时间" width="160" prop="created_at" sortable="custom">
            <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
          </el-table-column>
        </el-table>

        <div class="pagination-wrap">
          <el-pagination
            v-model:current-page="recordPage"
            v-model:page-size="recordPageSize"
            :total="recordTotal"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next, jumper"
            background
            @current-change="loadRecords"
            @size-change="loadRecords"
          />
        </div>
      </el-tab-pane>
    </el-tabs>

    <!-- 礼物表单弹窗 -->
    <el-dialog v-model="formVisible" :title="formTitle" width="640px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="编码" prop="gift_code">
          <el-input v-model="form.gift_code" maxlength="32" placeholder="如：rose99" />
        </el-form-item>
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" maxlength="64" />
        </el-form-item>
        <el-form-item label="分类" prop="category">
          <el-select v-model="form.category" style="width: 100%">
            <el-option v-for="(label, val) in categoryMap" :key="val" :label="label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item label="价格" prop="price">
          <el-input-number v-model="form.price" :min="0" :precision="2" :step="1" style="width: 180px" />
          <el-select v-model="form.price_unit" style="width: 120px; margin-left: 8px">
            <el-option label="金币" value="coin" />
            <el-option label="人民币" value="rmb" />
            <el-option label="钻石" value="diamond" />
          </el-select>
        </el-form-item>
        <el-form-item label="折扣价">
          <el-input-number v-model="form.discount_price" :min="0" :precision="2" :step="1" style="width: 180px" />
        </el-form-item>
        <el-form-item label="图标URL">
          <el-input v-model="form.icon" placeholder="图标URL" />
        </el-form-item>
        <el-form-item label="动画URL">
          <el-input v-model="form.animation" placeholder="动画效果URL（可选）" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" maxlength="500" />
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

const activeTab = ref('gifts')

// ===== 礼物列表 =====
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const gifts = ref([])
const sortField = ref('sort')
const sortOrder = ref('ascending')

const filters = reactive({ keyword: '', category: '', status: null })

const giftStats = reactive({ total: 0, enabled: 0, totalSent: 0, totalRevenue: '0.00' })

const categoryMap = {
  rose: '玫瑰', gift: '礼物', prop: '道具', badge: '徽章', car: '车辆', yacht: '游艇'
}
const categoryText = (c) => categoryMap[c] || c || '-'
const priceUnitText = (u) => ({ coin: '金币', rmb: '元', diamond: '钻石' }[u] || u || '')

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.keyword = ''; filters.category = ''; filters.status = null
  page.value = 1; loadList()
}
const onSortChange = ({ prop, order }) => {
  sortField.value = prop || 'sort'
  sortOrder.value = order || 'ascending'
  loadList()
}

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      keyword: filters.keyword.trim() || undefined,
      category: filters.category || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status,
      sort: sortField.value,
      order: sortOrder.value
    }
    const res = await request.get('/love/admin/gifts', { params })
    const data = res.data || {}
    list.value = data.list || data || []
    total.value = data.total || list.value.length
    gifts.value = list.value
    calcGiftStats()
  } catch (e) {
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const calcGiftStats = () => {
  giftStats.total = list.value.length
  giftStats.enabled = list.value.filter((r) => r.status === 1).length
  giftStats.totalSent = list.value.reduce((s, r) => s + (r.sent_count || 0), 0)
  const totalRevenue = list.value.reduce((s, r) => s + (r.sent_count || 0) * Number(r.price || 0), 0)
  giftStats.totalRevenue = totalRevenue.toFixed(2)
}

// ===== 礼物记录 =====
const recordLoading = ref(false)
const recordPage = ref(1)
const recordPageSize = ref(20)
const recordTotal = ref(0)
const records = ref([])
const recordSortField = ref('created_at')
const recordSortOrder = ref('descending')

const recordFilters = reactive({ keyword: '', gift_id: null, dateRange: null })

const recordStats = reactive({ total: 0, today: 0, uniqueSenders: 0, totalValue: '0.00' })

const onRecordSearch = () => { recordPage.value = 1; loadRecords() }
const onRecordReset = () => {
  recordFilters.keyword = ''; recordFilters.gift_id = null; recordFilters.dateRange = null
  recordPage.value = 1; loadRecords()
}
const onRecordSortChange = ({ prop, order }) => {
  recordSortField.value = prop || 'created_at'
  recordSortOrder.value = order || 'descending'
  loadRecords()
}

const loadRecords = async () => {
  recordLoading.value = true
  try {
    const params = {
      page: recordPage.value,
      page_size: recordPageSize.value,
      keyword: recordFilters.keyword.trim() || undefined,
      gift_id: recordFilters.gift_id || undefined,
      sort: recordSortField.value,
      order: recordSortOrder.value
    }
    if (recordFilters.dateRange && recordFilters.dateRange.length === 2) {
      params.start_date = recordFilters.dateRange[0]
      params.end_date = recordFilters.dateRange[1]
    }
    // 通过 gift-records 接口（如无独立接口，复用 gifts/available）
    const res = await request.get('/love/gifts', { params })
    const data = res.data || {}
    records.value = data.list || data.records || []
    recordTotal.value = data.total || records.value.length
    calcRecordStats()
  } catch (e) {
    records.value = []
    recordTotal.value = 0
  } finally {
    recordLoading.value = false
  }
}

const calcRecordStats = () => {
  const today = new Date().toISOString().slice(0, 10)
  recordStats.total = records.value.length
  recordStats.today = records.value.filter((r) => (r.created_at || '').startsWith(today)).length
  const senders = new Set(records.value.map((r) => r.from_user_id))
  recordStats.uniqueSenders = senders.size
  const totalValue = records.value.reduce((s, r) => s + Number(r.total_price || 0), 0)
  recordStats.totalValue = totalValue.toFixed(2)
}

const onTabChange = (tab) => {
  if (tab === 'records' && !records.value.length) {
    loadRecords()
  }
}

// ===== 礼物表单 =====
const formVisible = ref(false)
const formLoading = ref(false)
const formRef = ref(null)
const isEdit = ref(false)
const formTitle = computed(() => isEdit.value ? '编辑礼物' : '新建礼物')
const form = reactive({
  id: null, gift_code: '', name: '', category: 'gift',
  price: 0, discount_price: 0, price_unit: 'coin',
  icon: '', animation: '', sort: 0, status: 1, description: ''
})
const rules = {
  gift_code: [{ required: true, message: '请输入编码', trigger: 'blur' }],
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  category: [{ required: true, message: '请选择分类', trigger: 'change' }],
  price: [{ required: true, message: '请输入价格', trigger: 'blur' }]
}

const openCreate = () => {
  isEdit.value = false
  Object.assign(form, {
    id: null, gift_code: '', name: '', category: 'gift',
    price: 0, discount_price: 0, price_unit: 'coin',
    icon: '', animation: '', sort: 0, status: 1, description: ''
  })
  formVisible.value = true
}

const openEdit = (row) => {
  isEdit.value = true
  Object.assign(form, {
    id: row.id, gift_code: row.gift_code || '', name: row.name || '',
    category: row.category || 'gift', price: row.price || 0,
    discount_price: row.discount_price || 0, price_unit: row.price_unit || 'coin',
    icon: row.icon || '', animation: row.animation || '',
    sort: row.sort || 0, status: row.status ?? 1, description: row.description || ''
  })
  formVisible.value = true
}

const onSubmit = async () => {
  try {
    await formRef.value.validate()
    formLoading.value = true
    if (isEdit.value) {
      await request.put(`/love/admin/gifts/${form.id}`, form)
    } else {
      await request.post('/love/admin/gifts', form)
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
    await ElMessageBox.confirm(`确认删除礼物 "${row.name}"？`, '提示', { type: 'warning' })
    await request.delete(`/love/admin/gifts/${row.id}`)
    ElMessage.success('删除成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onToggle = async (row, val) => {
  try {
    await request.put(`/love/admin/gifts/${row.id}/status`, { status: val ? 1 : 0 })
    row.status = val ? 1 : 0
    ElMessage.success('状态已更新')
  } catch (e) { /* ignore */ }
}

onMounted(() => loadList())
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-item { text-align: center; padding: 16px; }
.stat-value { font-size: 22px; font-weight: 600; color: #303133; }
.text-success { color: #67c23a; }
.text-warning { color: #e6a23c; }
.text-primary { color: #409eff; }
.text-danger { color: #f56c6c; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; margin-bottom: 12px; }
.text-muted { color: #909399; }
.cover-thumb { width: 50px; height: 50px; border-radius: 4px; border: 1px solid #ebeef5; }
.cover-empty { display: flex; align-items: center; justify-content: center; background: #fafafa; color: #909399; font-size: 12px; }
.price { color: #f56c6c; font-weight: 600; }
.user-cell { display: flex; align-items: center; gap: 8px; }
.user-avatar { width: 36px; height: 36px; border-radius: 50%; border: 1px solid #ebeef5; }
.user-info { display: flex; flex-direction: column; }
.user-name { color: #303133; font-size: 13px; }
.gift-cell { display: flex; align-items: center; gap: 8px; }
.gift-icon { width: 32px; height: 32px; border-radius: 4px; border: 1px solid #ebeef5; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
