<template>
  <div class="app-container">
    <div class="page-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button type="primary" :icon="Plus" @click="openCreate">手动创建</el-button>
          <el-button :icon="Refresh" @click="loadList">刷新</el-button>
          <el-button :icon="DataAnalysis" @click="openSummary">查看汇总</el-button>
        </div>
        <div class="toolbar-right">
          <el-input-number
            v-model="partnerFilter"
            :min="0"
            placeholder="合伙人ID"
            style="width: 130px"
            @change="onSearch"
          />
          <el-input-number
            v-model="orderFilter"
            :min="0"
            placeholder="订单ID"
            style="width: 130px; margin-left: 8px"
            @change="onSearch"
          />
          <el-select
            v-model="levelFilter"
            placeholder="级别"
            clearable
            style="width: 120px; margin-left: 8px"
            @change="onSearch"
          >
            <el-option label="一级分销" :value="1" />
            <el-option label="二级分销" :value="2" />
          </el-select>
          <el-select
            v-model="statusFilter"
            placeholder="状态"
            clearable
            style="width: 120px; margin-left: 8px"
            @change="onSearch"
          >
            <el-option label="待结算" :value="0" />
            <el-option label="已结算" :value="1" />
            <el-option label="已取消" :value="2" />
          </el-select>
          <el-button type="primary" :icon="Search" style="margin-left: 8px" @click="onSearch">搜索</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="list" border stripe @selection-change="onSelectionChange">
        <el-table-column type="selection" width="50" :selectable="(row) => row.status === 0" />
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="partner_id" label="合伙人ID" width="100" />
        <el-table-column prop="order_id" label="订单ID" width="100" />
        <el-table-column prop="channel_id" label="渠道ID" width="90">
          <template #default="{ row }">{{ row.channel_id || '-' }}</template>
        </el-table-column>
        <el-table-column label="级别" width="100">
          <template #default="{ row }">
            <el-tag :type="row.level === 2 ? 'warning' : 'primary'" size="small">
              {{ row.level === 2 ? '二级' : '一级' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="订单金额" width="120" align="right">
          <template #default="{ row }">¥{{ formatAmount(row.order_amount) }}</template>
        </el-table-column>
        <el-table-column label="佣金金额" width="120" align="right">
          <template #default="{ row }">
            <span class="amount">¥{{ formatAmount(row.commission_amount) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="佣金率" width="100" align="right">
          <template #default="{ row }">{{ formatRate(row.commission_rate) }}</template>
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
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 0"
              link
              type="success"
              size="small"
              @click="onSettle(row)"
            >结算</el-button>
            <el-button
              v-if="row.status === 0 || row.status === 1"
              link
              type="danger"
              size="small"
              @click="onCancel(row)"
            >取消</el-button>
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

      <div v-if="selectedIds.length" class="batch-bar">
        <span>已选 {{ selectedIds.length }} 条</span>
        <el-button type="success" size="small" style="margin-left: 12px" @click="onBatchSettle">批量结算</el-button>
      </div>
    </div>

    <!-- 创建弹窗 -->
    <el-dialog v-model="createVisible" title="手动创建佣金记录" width="560px">
      <el-form ref="createFormRef" :model="createForm" :rules="createRules" label-width="110px">
        <el-form-item label="合伙人ID" prop="partner_id">
          <el-input-number v-model="createForm.partner_id" :min="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="订单ID" prop="order_id">
          <el-input-number v-model="createForm.order_id" :min="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="渠道ID">
          <el-input-number v-model="createForm.channel_id" :min="0" />
          <span class="tip">0=无渠道</span>
        </el-form-item>
        <el-form-item label="级别" prop="level">
          <el-radio-group v-model="createForm.level">
            <el-radio :value="1">一级分销</el-radio>
            <el-radio :value="2">二级分销</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="订单金额" prop="order_amount">
          <el-input-number v-model="createForm.order_amount" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="佣金金额" prop="commission_amount">
          <el-input-number v-model="createForm.commission_amount" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="佣金比例" prop="commission_rate">
          <el-input-number v-model="createForm.commission_rate" :min="0" :max="1" :step="0.01" :precision="4" />
          <span class="tip">0-1 之间</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onSubmitCreate">创建</el-button>
      </template>
    </el-dialog>

    <!-- 汇总弹窗 -->
    <el-dialog v-model="summaryVisible" title="佣金汇总" width="520px">
      <el-form :model="summaryForm" label-width="100px" inline>
        <el-form-item label="合伙人ID">
          <el-input-number v-model="summaryForm.partner_id" :min="0" placeholder="0=全部" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="summaryLoading" @click="loadSummary">查询</el-button>
        </el-form-item>
      </el-form>
      <el-descriptions v-if="summaryData" :column="1" border>
        <el-descriptions-item label="总佣金">¥{{ formatAmount(summaryData.total) }}</el-descriptions-item>
        <el-descriptions-item label="待结算">¥{{ formatAmount(summaryData.pending) }}</el-descriptions-item>
        <el-descriptions-item label="已结算">¥{{ formatAmount(summaryData.settled) }}</el-descriptions-item>
        <el-descriptions-item label="已取消">¥{{ formatAmount(summaryData.canceled) }}</el-descriptions-item>
      </el-descriptions>
      <div v-else class="empty-tip">请输入查询条件</div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Search, DataAnalysis } from '@element-plus/icons-vue'
import {
  getDistributionCommissionList,
  createDistributionCommission,
  settleDistributionCommission,
  batchSettleDistributionCommission,
  cancelDistributionCommission,
  getDistributionCommissionSummary
} from '@/api/distribution'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const partnerFilter = ref(0)
const orderFilter = ref(0)
const levelFilter = ref(null)
const statusFilter = ref(null)
const selectedIds = ref([])

const createVisible = ref(false)
const summaryVisible = ref(false)
const submitting = ref(false)
const summaryLoading = ref(false)
const createFormRef = ref(null)

const createForm = reactive({
  partner_id: 1,
  order_id: 1,
  channel_id: 0,
  level: 1,
  order_amount: 0,
  commission_amount: 0,
  commission_rate: 0.1
})
const createRules = {
  partner_id: [{ required: true, message: '请输入合伙人ID', trigger: 'blur' }],
  order_id: [{ required: true, message: '请输入订单ID', trigger: 'blur' }],
  level: [{ required: true, message: '请选择级别', trigger: 'change' }],
  order_amount: [{ required: true, message: '请输入订单金额', trigger: 'blur' }],
  commission_amount: [{ required: true, message: '请输入佣金金额', trigger: 'blur' }],
  commission_rate: [{ required: true, message: '请输入佣金比例', trigger: 'blur' }]
}

const summaryForm = reactive({ partner_id: 0 })
const summaryData = ref(null)

function statusText(s) {
  return s === 1 ? '已结算' : s === 2 ? '已取消' : '待结算'
}
function statusTagType(s) {
  return s === 1 ? 'success' : s === 2 ? 'info' : 'warning'
}
function formatAmount(n) {
  if (n === undefined || n === null) return '0.00'
  return Number(n).toFixed(2)
}
function formatRate(r) {
  if (r === undefined || r === null) return '0.00%'
  return (Number(r) * 100).toFixed(2) + '%'
}

async function loadList() {
  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value }
    if (partnerFilter.value > 0) params.partner_id = partnerFilter.value
    if (orderFilter.value > 0) params.order_id = orderFilter.value
    if (levelFilter.value !== null && levelFilter.value !== '') params.level = levelFilter.value
    if (statusFilter.value !== null && statusFilter.value !== '') params.status = statusFilter.value
    const res = await getDistributionCommissionList(params)
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (e) {
    ElMessage.error('加载佣金列表失败')
  } finally {
    loading.value = false
  }
}

function onSearch() {
  page.value = 1
  loadList()
}

function onSelectionChange(rows) {
  selectedIds.value = rows.map(r => r.id)
}

function openCreate() {
  Object.assign(createForm, {
    partner_id: 1, order_id: 1, channel_id: 0, level: 1,
    order_amount: 0, commission_amount: 0, commission_rate: 0.1
  })
  createVisible.value = true
}

async function onSubmitCreate() {
  if (!createFormRef.value) return
  await createFormRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      await createDistributionCommission({
        partner_id: createForm.partner_id,
        order_id: createForm.order_id,
        channel_id: createForm.channel_id,
        level: createForm.level,
        order_amount: createForm.order_amount,
        commission_amount: createForm.commission_amount,
        commission_rate: createForm.commission_rate
      })
      ElMessage.success('佣金记录已创建')
      createVisible.value = false
      loadList()
    } catch (e) {
      ElMessage.error(e.message || '操作失败')
    } finally {
      submitting.value = false
    }
  })
}

async function onSettle(row) {
  try {
    await ElMessageBox.confirm(`确定结算佣金记录 #${row.id} 吗？`, '提示', { type: 'warning' })
    await settleDistributionCommission(row.id)
    ElMessage.success('结算成功')
    loadList()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error(e.message || '操作失败')
  }
}

async function onBatchSettle() {
  if (!selectedIds.value.length) return
  try {
    await ElMessageBox.confirm(`确定批量结算 ${selectedIds.value.length} 条佣金吗？`, '提示', { type: 'warning' })
    await batchSettleDistributionCommission({ ids: selectedIds.value })
    ElMessage.success('批量结算成功')
    loadList()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error(e.message || '操作失败')
  }
}

async function onCancel(row) {
  try {
    await ElMessageBox.confirm(`确定取消佣金记录 #${row.id} 吗？`, '提示', { type: 'warning' })
    await cancelDistributionCommission(row.id)
    ElMessage.success('已取消')
    loadList()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error(e.message || '操作失败')
  }
}

function openSummary() {
  summaryVisible.value = true
  summaryData.value = null
}

async function loadSummary() {
  summaryLoading.value = true
  try {
    const params = {}
    if (summaryForm.partner_id > 0) params.partner_id = summaryForm.partner_id
    const res = await getDistributionCommissionSummary(params)
    summaryData.value = res.data || null
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
.amount { font-weight: 600; color: #e6a23c; }
.tip { margin-left: 8px; color: #909399; font-size: 12px; }
.empty-tip { padding: 16px; text-align: center; color: #c0c4cc; }
.batch-bar { margin-top: 12px; padding: 8px 12px; background: #f4f4f5; border-radius: 4px; }
</style>
