<template>
  <div class="app-container">
    <div class="page-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button type="primary" :icon="Plus" @click="openGenerate">生成结算单</el-button>
          <el-button :icon="Refresh" @click="loadList">刷新</el-button>
          <el-button :icon="DataAnalysis" @click="openSummary">查看汇总</el-button>
        </div>
        <div class="toolbar-right">
          <el-input-number
            v-model="shopFilter"
            :min="0"
            placeholder="店铺ID"
            style="width: 130px"
            @change="onSearch"
          />
          <el-input
            v-model="periodFilter"
            placeholder="周期 YYYY-MM"
            clearable
            style="width: 160px; margin-left: 8px"
            @keyup.enter="onSearch"
            @clear="onSearch"
          />
          <el-select
            v-model="statusFilter"
            placeholder="状态"
            clearable
            style="width: 140px; margin-left: 8px"
            @change="onSearch"
          >
            <el-option label="待结算" :value="0" />
            <el-option label="已结算" :value="1" />
            <el-option label="已提现" :value="2" />
            <el-option label="已撤销" :value="3" />
          </el-select>
          <el-button type="primary" :icon="Search" style="margin-left: 8px" @click="onSearch">搜索</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="shop_id" label="店铺ID" width="100" />
        <el-table-column prop="period" label="结算周期" width="120" />
        <el-table-column label="总金额" width="120" align="right">
          <template #default="{ row }">
            <span class="amount">¥{{ formatAmount(row.total_amount) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="平台佣金" width="120" align="right">
          <template #default="{ row }">
            <span class="amount-platform">¥{{ formatAmount(row.platform_fee) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="商户应得" width="120" align="right">
          <template #default="{ row }">
            <span class="amount-shop">¥{{ formatAmount(row.shop_amount) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">
              {{ row.status_text || statusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="settled_at" label="结算时间" width="170">
          <template #default="{ row }">{{ row.settled_at || '-' }}</template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="170" />
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 1"
              link
              type="warning"
              size="small"
              @click="onWithdraw(row)"
            >提现</el-button>
            <el-button
              v-if="row.status === 0 || row.status === 2"
              link
              type="primary"
              size="small"
              @click="openAudit(row)"
            >审核</el-button>
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
          @size-change="loadList"
          @current-change="loadList"
        />
      </div>
    </div>

    <!-- 生成结算单弹窗 -->
    <el-dialog v-model="generateVisible" title="生成结算单" width="520px">
      <el-form ref="genFormRef" :model="generateForm" :rules="generateRules" label-width="100px">
        <el-form-item label="店铺ID" prop="shop_id">
          <el-input-number v-model="generateForm.shop_id" :min="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="结算周期" prop="period">
          <el-input v-model="generateForm.period" placeholder="YYYY-MM，如 2026-07" />
        </el-form-item>
        <el-form-item label="总金额" prop="total_amount">
          <el-input-number v-model="generateForm.total_amount" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="平台佣金比例" prop="platform_rate">
          <el-input-number v-model="generateForm.platform_rate" :min="0" :max="1" :step="0.05" :precision="2" />
          <span class="tip">0-1 之间，如 0.05 表示 5%</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="generateVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onSubmitGenerate">生成</el-button>
      </template>
    </el-dialog>

    <!-- 审核弹窗 -->
    <el-dialog v-model="auditVisible" title="结算单审核" width="480px">
      <el-form :model="auditForm" label-width="100px">
        <el-form-item label="结算单">
          <span>#{{ auditForm.id }} (店铺 #{{ auditForm.shop_id }})</span>
        </el-form-item>
        <el-form-item label="周期">
          <span>{{ auditForm.period }}</span>
        </el-form-item>
        <el-form-item label="商户应得">
          <span class="amount-shop">¥{{ formatAmount(auditForm.shop_amount) }}</span>
        </el-form-item>
        <el-form-item label="当前状态">
          <el-tag :type="statusTagType(auditForm.current)" size="small">
            {{ statusText(auditForm.current) }}
          </el-tag>
        </el-form-item>
        <el-form-item label="审核结果">
          <el-radio-group v-model="auditForm.status">
            <el-radio :value="1">通过</el-radio>
            <el-radio :value="2">拒绝</el-radio>
            <el-radio :value="3">撤销</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="审核备注">
          <el-input v-model="auditForm.reason" type="textarea" :rows="2" maxlength="500" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="auditVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onSubmitAudit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 汇总弹窗 -->
    <el-dialog v-model="summaryVisible" title="结算汇总" width="640px">
      <el-form :model="summaryForm" label-width="100px" inline>
        <el-form-item label="店铺ID">
          <el-input-number v-model="summaryForm.shop_id" :min="0" placeholder="0=全部" />
        </el-form-item>
        <el-form-item label="周期">
          <el-input v-model="summaryForm.period" placeholder="YYYY-MM" style="width: 140px" />
        </el-form-item>
        <el-form-item label="开始时间">
          <el-input v-model="summaryForm.start_time" placeholder="2026-01-01" style="width: 140px" />
        </el-form-item>
        <el-form-item label="结束时间">
          <el-input v-model="summaryForm.end_time" placeholder="2026-12-31" style="width: 140px" />
        </el-form-item>
      </el-form>
      <div style="margin-bottom: 16px;">
        <el-button type="primary" :loading="summaryLoading" @click="loadSummaryByShop">按店铺汇总</el-button>
        <el-button type="success" :loading="summaryLoading" @click="loadSummaryByPeriod" style="margin-left: 8px">按周期汇总</el-button>
      </div>
      <el-table :data="summaryList" border stripe v-loading="summaryLoading">
        <el-table-column prop="shop_id" label="店铺ID" width="100" />
        <el-table-column label="总金额" width="120" align="right">
          <template #default="{ row }">¥{{ formatAmount(row.total_amount) }}</template>
        </el-table-column>
        <el-table-column label="平台佣金" width="120" align="right">
          <template #default="{ row }">¥{{ formatAmount(row.platform_fee) }}</template>
        </el-table-column>
        <el-table-column label="商户应得" width="120" align="right">
          <template #default="{ row }">¥{{ formatAmount(row.shop_amount) }}</template>
        </el-table-column>
        <el-table-column prop="count" label="结算单数" width="100" />
      </el-table>
      <div v-if="!summaryList.length && !summaryLoading" class="empty-tip">暂无数据</div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Search, DataAnalysis } from '@element-plus/icons-vue'
import {
  getMerchantSettleList,
  generateMerchantSettle,
  withdrawMerchantSettle,
  auditMerchantSettleWithdraw,
  getMerchantSettleSummaryByShop,
  getMerchantSettleSummaryByPeriod
} from '@/api/merchant'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const shopFilter = ref(0)
const periodFilter = ref('')
const statusFilter = ref(null)

const generateVisible = ref(false)
const auditVisible = ref(false)
const summaryVisible = ref(false)
const submitting = ref(false)
const summaryLoading = ref(false)
const genFormRef = ref(null)

const generateForm = reactive({
  shop_id: 1,
  period: '',
  total_amount: 0,
  platform_rate: 0.05
})
const generateRules = {
  shop_id: [{ required: true, message: '请输入店铺ID', trigger: 'blur' }],
  period: [
    { required: true, message: '请输入结算周期', trigger: 'blur' },
    { pattern: /^\d{4}-\d{2}$/, message: '格式应为 YYYY-MM', trigger: 'blur' }
  ],
  total_amount: [{ required: true, message: '请输入总金额', trigger: 'blur' }]
}

const auditForm = reactive({
  id: 0,
  shop_id: 0,
  period: '',
  shop_amount: 0,
  current: 0,
  status: 1,
  reason: ''
})

const summaryForm = reactive({
  shop_id: 0,
  period: '',
  start_time: '',
  end_time: ''
})
const summaryList = ref([])

function statusText(s) {
  return s === 1 ? '已结算' : s === 2 ? '已提现' : s === 3 ? '已撤销' : '待结算'
}

function statusTagType(s) {
  return s === 1 ? 'success' : s === 2 ? 'primary' : s === 3 ? 'info' : 'warning'
}

function formatAmount(n) {
  if (n === undefined || n === null) return '0.00'
  return Number(n).toFixed(2)
}

async function loadList() {
  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value }
    if (shopFilter.value > 0) params.shop_id = shopFilter.value
    if (periodFilter.value) params.period = periodFilter.value
    if (statusFilter.value !== null && statusFilter.value !== '') params.status = statusFilter.value
    const res = await getMerchantSettleList(params)
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (e) {
    ElMessage.error('加载结算列表失败')
  } finally {
    loading.value = false
  }
}

function onSearch() {
  page.value = 1
  loadList()
}

function openGenerate() {
  const now = new Date()
  const period = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
  Object.assign(generateForm, {
    shop_id: 1,
    period,
    total_amount: 0,
    platform_rate: 0.05
  })
  generateVisible.value = true
}

async function onSubmitGenerate() {
  if (!genFormRef.value) return
  await genFormRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      await generateMerchantSettle({
        shop_id: generateForm.shop_id,
        period: generateForm.period,
        total_amount: generateForm.total_amount,
        platform_rate: generateForm.platform_rate
      })
      ElMessage.success('结算单已生成')
      generateVisible.value = false
      loadList()
    } catch (e) {
      ElMessage.error(e.message || '操作失败')
    } finally {
      submitting.value = false
    }
  })
}

async function onWithdraw(row) {
  try {
    await ElMessageBox.confirm(`确定要对结算单 #${row.id} 发起提现申请吗？`, '提示', { type: 'warning' })
    await withdrawMerchantSettle(row.id)
    ElMessage.success('提现申请已提交')
    loadList()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error(e.message || '操作失败')
  }
}

function openAudit(row) {
  auditForm.id = row.id
  auditForm.shop_id = row.shop_id
  auditForm.period = row.period
  auditForm.shop_amount = row.shop_amount
  auditForm.current = row.status
  auditForm.status = 1
  auditForm.reason = ''
  auditVisible.value = true
}

async function onSubmitAudit() {
  submitting.value = true
  try {
    await auditMerchantSettleWithdraw(auditForm.id, {
      status: auditForm.status,
      reason: auditForm.reason
    })
    ElMessage.success('审核完成')
    auditVisible.value = false
    loadList()
  } catch (e) {
    ElMessage.error(e.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

function openSummary() {
  summaryVisible.value = true
  summaryList.value = []
}

async function loadSummaryByShop() {
  summaryLoading.value = true
  try {
    const params = {}
    if (summaryForm.shop_id > 0) params.shop_id = summaryForm.shop_id
    if (summaryForm.period) params.period = summaryForm.period
    if (summaryForm.start_time) params.start_time = summaryForm.start_time
    if (summaryForm.end_time) params.end_time = summaryForm.end_time
    const res = await getMerchantSettleSummaryByShop(params)
    const data = res.data
    if (Array.isArray(data)) {
      summaryList.value = data
    } else if (data && typeof data === 'object') {
      summaryList.value = [data]
    } else {
      summaryList.value = []
    }
  } catch (e) {
    ElMessage.error('加载汇总失败')
  } finally {
    summaryLoading.value = false
  }
}

async function loadSummaryByPeriod() {
  summaryLoading.value = true
  try {
    const params = {}
    if (summaryForm.shop_id > 0) params.shop_id = summaryForm.shop_id
    if (summaryForm.period) params.period = summaryForm.period
    if (summaryForm.start_time) params.start_time = summaryForm.start_time
    if (summaryForm.end_time) params.end_time = summaryForm.end_time
    const res = await getMerchantSettleSummaryByPeriod(params)
    const data = res.data
    if (Array.isArray(data)) {
      summaryList.value = data
    } else if (data && typeof data === 'object') {
      summaryList.value = [data]
    } else {
      summaryList.value = []
    }
  } catch (e) {
    ElMessage.error('加载汇总失败')
  } finally {
    summaryLoading.value = false
  }
}

onMounted(() => loadList())
</script>

<style scoped>
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.amount { font-weight: 600; }
.amount-platform { color: #e6a23c; }
.amount-shop { color: #67c23a; font-weight: 600; }
.tip { margin-left: 8px; color: #909399; font-size: 12px; }
.empty-tip { padding: 16px; text-align: center; color: #c0c4cc; }
</style>
