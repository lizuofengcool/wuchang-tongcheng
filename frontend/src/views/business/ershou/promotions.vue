<template>
  <div class="app-container">
    <!-- 推广统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">推广总数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.active }}</div><div class="stat-label">进行中</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">¥{{ stats.amount.toFixed(2) }}</div><div class="stat-label">总花费</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.roi.toFixed(2) }}</div><div class="stat-label">平均 ROI</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="商品ID">
          <el-input v-model="filters.ershou_id" placeholder="商品ID" clearable style="width: 140px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="推广类型">
          <el-select v-model="filters.promotion_type" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option v-for="(label, val) in typeMap" :key="val" :label="label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="进行中" :value="1" />
            <el-option label="已结束" :value="2" />
            <el-option label="已取消" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar">
        <el-button :icon="Refresh" @click="loadList">刷新</el-button>
        <el-button type="primary" :icon="Plus" @click="openCreate">创建推广</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="ershou_id" label="商品ID" width="90" />
        <el-table-column label="推广类型" width="120">
          <template #default="{ row }">
            <el-tag size="small">{{ typeMap[row.promotion_type] || row.promotion_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ row.status_text || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="开始时间" width="160">
          <template #default="{ row }">{{ formatTime(row.start_time) }}</template>
        </el-table-column>
        <el-table-column label="结束时间" width="160">
          <template #default="{ row }">{{ formatTime(row.end_time) }}</template>
        </el-table-column>
        <el-table-column label="花费" width="100">
          <template #default="{ row }">¥{{ Number(row.amount || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="impression_count" label="曝光" width="80" />
        <el-table-column prop="click_count" label="点击" width="80" />
        <el-table-column label="CTR" width="80">
          <template #default="{ row }">{{ row.impression_count ? ((row.click_count / row.impression_count) * 100).toFixed(1) + '%' : '-' }}</template>
        </el-table-column>
        <el-table-column prop="order_count" label="下单" width="80" />
        <el-table-column label="ROI" width="80">
          <template #default="{ row }">{{ Number(row.roi || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
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

    <!-- 创建推广弹窗 -->
    <el-dialog v-model="createVisible" title="创建推广" width="500px">
      <el-form ref="createFormRef" :model="createForm" :rules="createRules" label-width="100px">
        <el-form-item label="商品ID" prop="ershou_id">
          <el-input-number v-model="createForm.ershou_id" :min="1" :controls="false" style="width: 100%" />
        </el-form-item>
        <el-form-item label="推广类型" prop="promotion_type">
          <el-select v-model="createForm.promotion_type" placeholder="请选择" style="width: 100%">
            <el-option v-for="(label, val) in typeMap" :key="val" :label="label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item label="时长(天)" prop="duration_days">
          <el-input-number v-model="createForm.duration_days" :min="1" :max="365" />
        </el-form-item>
        <el-form-item label="金额" prop="amount">
          <el-input-number v-model="createForm.amount" :min="0" :precision="2" :controls="false" style="width: 100%" />
        </el-form-item>
        <el-form-item label="支付方式">
          <el-select v-model="createForm.pay_method" style="width: 100%">
            <el-option label="微信" value="wechat" />
            <el-option label="支付宝" value="alipay" />
            <el-option label="余额" value="balance" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="createLoading" @click="onCreate">确认创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, RefreshLeft, Search, Plus } from '@element-plus/icons-vue'
import { listErshouPromotions, createErshouPromotion, getErshouPromotionStats } from '@/api/ershou'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const filters = reactive({ ershou_id: '', promotion_type: '', status: null })

const typeMap = {
  home_banner: '首页Banner', channel_top: '频道置顶', search_top: '搜索置顶',
  featured: '精选推荐', urgent: '急转推广', refresh: '刷新排名'
}

const stats = reactive({ total: 0, active: 0, amount: 0, roi: 0 })

const statusTagType = (s) => ({ 1: 'success', 2: 'info', 3: 'danger' }[s] || 'info')

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => { filters.ershou_id = ''; filters.promotion_type = ''; filters.status = null; page.value = 1; loadList() }

const loadList = async () => {
  loading.value = true
  try {
    // 由于推广列表是按商品 ID 查询，若无 ershou_id 则展示空
    if (!filters.ershou_id) {
      list.value = []
      total.value = 0
      return
    }
    const res = await listErshouPromotions(filters.ershou_id, {
      page: page.value,
      page_size: pageSize.value,
      promotion_type: filters.promotion_type || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status
    })
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
    // 统计接口需指定商品 ID，此处展示全平台汇总
    if (!filters.ershou_id) return
    const res = await getErshouPromotionStats(filters.ershou_id)
    const d = res.data || {}
    stats.total = d.total_promotions || 0
    stats.active = d.active_promotions || 0
    stats.amount = d.total_amount || 0
    stats.roi = d.avg_roi || 0
  } catch (e) { /* ignore */ }
}

// ===== 创建推广 =====
const createVisible = ref(false)
const createLoading = ref(false)
const createFormRef = ref(null)
const createForm = reactive({
  ershou_id: undefined, promotion_type: '', duration_days: 7, amount: 0, pay_method: 'wechat'
})
const createRules = {
  ershou_id: [{ required: true, message: '请输入商品ID', trigger: 'blur' }],
  promotion_type: [{ required: true, message: '请选择推广类型', trigger: 'change' }],
  duration_days: [{ required: true, message: '请输入时长', trigger: 'blur' }]
}

const openCreate = () => {
  Object.assign(createForm, { ershou_id: undefined, promotion_type: '', duration_days: 7, amount: 0, pay_method: 'wechat' })
  createVisible.value = true
}

const onCreate = async () => {
  try {
    await createFormRef.value.validate()
    createLoading.value = true
    await createErshouPromotion(createForm.ershou_id, {
      promotion_type: createForm.promotion_type,
      duration_days: createForm.duration_days,
      amount: createForm.amount,
      pay_method: createForm.pay_method
    })
    ElMessage.success('创建成功')
    createVisible.value = false
    filters.ershou_id = createForm.ershou_id
    await Promise.all([loadList(), loadStats()])
  } catch (e) {
    // 校验失败或接口失败
  } finally {
    createLoading.value = false
  }
}

onMounted(() => loadList())
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-item { text-align: center; padding: 16px; }
.stat-value { font-size: 24px; font-weight: 600; color: #409eff; }
.stat-label { font-size: 12px; color: #909399; margin-top: 4px; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; margin-bottom: 12px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
