<template>
  <div class="app-container">
    <div class="page-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button :icon="Refresh" @click="loadList">刷新</el-button>
          <el-button type="warning" :icon="Bell" @click="loadPending">查看待审核</el-button>
        </div>
        <div class="toolbar-right">
          <el-input-number
            v-model="partnerFilter"
            :min="0"
            placeholder="合伙人ID"
            style="width: 140px"
            @change="onSearch"
          />
          <el-select
            v-model="statusFilter"
            placeholder="状态"
            clearable
            style="width: 140px; margin-left: 8px"
            @change="onSearch"
          >
            <el-option label="申请中" :value="0" />
            <el-option label="已审核" :value="1" />
            <el-option label="已打款" :value="2" />
            <el-option label="已拒绝" :value="3" />
          </el-select>
          <el-button type="primary" :icon="Search" style="margin-left: 8px" @click="onSearch">搜索</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="partner_id" label="合伙人ID" width="100" />
        <el-table-column label="提现金额" width="130" align="right">
          <template #default="{ row }">
            <span class="amount">¥{{ formatAmount(row.amount) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">
              {{ row.status_text || statusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="收款信息" min-width="240">
          <template #default="{ row }">
            <span v-if="parseBank(row.bank_info)">
              {{ parseBank(row.bank_info).bank_name }} / {{ parseBank(row.bank_info).account_name }} / {{ maskAccount(parseBank(row.bank_info).account_no) }}
            </span>
            <span v-else class="muted">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="audit_reason" label="审核备注" min-width="160">
          <template #default="{ row }">{{ row.audit_reason || '-' }}</template>
        </el-table-column>
        <el-table-column prop="audited_at" label="审核时间" width="170">
          <template #default="{ row }">{{ row.audited_at || '-' }}</template>
        </el-table-column>
        <el-table-column prop="paid_at" label="打款时间" width="170">
          <template #default="{ row }">{{ row.paid_at || '-' }}</template>
        </el-table-column>
        <el-table-column prop="created_at" label="申请时间" width="170" />
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 0"
              link
              type="success"
              size="small"
              @click="openAudit(row)"
            >审核</el-button>
            <el-button
              v-if="row.status === 1"
              link
              type="primary"
              size="small"
              @click="onPay(row)"
            >打款确认</el-button>
            <el-button
              v-if="row.status === 0 || row.status === 1"
              link
              type="danger"
              size="small"
              @click="openReject(row)"
            >拒绝</el-button>
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

    <!-- 审核弹窗 -->
    <el-dialog v-model="auditVisible" title="提现审核" width="480px">
      <el-form :model="auditForm" label-width="100px">
        <el-form-item label="提现单号">
          <span>#{{ auditForm.id }}</span>
        </el-form-item>
        <el-form-item label="合伙人ID">
          <span>{{ auditForm.partner_id }}</span>
        </el-form-item>
        <el-form-item label="提现金额">
          <span class="amount">¥{{ formatAmount(auditForm.amount) }}</span>
        </el-form-item>
        <el-form-item label="收款信息">
          <div v-if="parseBank(auditForm.bank_info)">
            <div>银行：{{ parseBank(auditForm.bank_info).bank_name }}</div>
            <div>户名：{{ parseBank(auditForm.bank_info).account_name }}</div>
            <div>账号：{{ parseBank(auditForm.bank_info).account_no }}</div>
            <div>支行：{{ parseBank(auditForm.bank_info).branch || '-' }}</div>
            <div>类型：{{ bankTypeText(parseBank(auditForm.bank_info).type) }}</div>
          </div>
          <span v-else class="muted">无</span>
        </el-form-item>
        <el-form-item label="审核结果">
          <el-radio-group v-model="auditForm.status">
            <el-radio :value="1">通过</el-radio>
            <el-radio :value="3">拒绝</el-radio>
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

    <!-- 拒绝弹窗 -->
    <el-dialog v-model="rejectVisible" title="拒绝提现" width="460px">
      <el-form :model="rejectForm" label-width="100px">
        <el-form-item label="提现单号">
          <span>#{{ rejectForm.id }}</span>
        </el-form-item>
        <el-form-item label="拒绝原因">
          <el-input v-model="rejectForm.reason" type="textarea" :rows="3" maxlength="500" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rejectVisible = false">取消</el-button>
        <el-button type="danger" :loading="submitting" @click="onSubmitReject">确定拒绝</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Search, Bell } from '@element-plus/icons-vue'
import {
  getDistributionWithdrawalList,
  getDistributionWithdrawalPendingList,
  auditDistributionWithdrawal,
  payDistributionWithdrawal,
  rejectDistributionWithdrawal
} from '@/api/distribution'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const partnerFilter = ref(0)
const statusFilter = ref(null)
const showPendingOnly = ref(false)

const auditVisible = ref(false)
const rejectVisible = ref(false)
const submitting = ref(false)

const auditForm = reactive({
  id: 0, partner_id: 0, amount: 0, bank_info: null, status: 1, reason: ''
})
const rejectForm = reactive({ id: 0, reason: '' })

function statusText(s) {
  return ['申请中', '已审核', '已打款', '已拒绝'][s] || ''
}
function statusTagType(s) {
  return s === 2 ? 'success' : s === 3 ? 'danger' : s === 1 ? 'primary' : 'warning'
}
function formatAmount(n) {
  if (n === undefined || n === null) return '0.00'
  return Number(n).toFixed(2)
}
function parseBank(b) {
  if (!b) return null
  if (typeof b === 'object') return b
  try { return JSON.parse(b) } catch { return null }
}
function maskAccount(no) {
  if (!no) return ''
  const s = String(no)
  if (s.length <= 4) return s
  return s.slice(0, 2) + '****' + s.slice(-4)
}
function bankTypeText(t) {
  if (t === 'alipay') return '支付宝'
  if (t === 'wechat') return '微信'
  return '银行卡'
}

async function loadList() {
  loading.value = true
  showPendingOnly.value = false
  try {
    const params = { page: page.value, page_size: pageSize.value }
    if (partnerFilter.value > 0) params.partner_id = partnerFilter.value
    if (statusFilter.value !== null && statusFilter.value !== '') params.status = statusFilter.value
    const res = await getDistributionWithdrawalList(params)
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (e) {
    ElMessage.error('加载提现列表失败')
  } finally {
    loading.value = false
  }
}

async function loadPending() {
  loading.value = true
  showPendingOnly.value = true
  page.value = 1
  try {
    const res = await getDistributionWithdrawalPendingList({ page: page.value, page_size: pageSize.value })
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (e) {
    ElMessage.error('加载待审核列表失败')
  } finally {
    loading.value = false
  }
}

function onSearch() {
  page.value = 1
  if (showPendingOnly.value) {
    loadList()
  } else {
    loadList()
  }
}

function openAudit(row) {
  auditForm.id = row.id
  auditForm.partner_id = row.partner_id
  auditForm.amount = row.amount
  auditForm.bank_info = row.bank_info
  auditForm.status = 1
  auditForm.reason = ''
  auditVisible.value = true
}

async function onSubmitAudit() {
  submitting.value = true
  try {
    await auditDistributionWithdrawal(auditForm.id, {
      status: auditForm.status,
      reason: auditForm.reason
    })
    ElMessage.success('审核完成')
    auditVisible.value = false
    reloadCurrent()
  } catch (e) {
    ElMessage.error(e.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

async function onPay(row) {
  try {
    await ElMessageBox.confirm(`确定对提现单 #${row.id} 进行打款确认吗？`, '提示', { type: 'warning' })
    await payDistributionWithdrawal(row.id, {})
    ElMessage.success('已确认打款')
    reloadCurrent()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error(e.message || '操作失败')
  }
}

function openReject(row) {
  rejectForm.id = row.id
  rejectForm.reason = ''
  rejectVisible.value = true
}

async function onSubmitReject() {
  submitting.value = true
  try {
    await rejectDistributionWithdrawal(rejectForm.id, { reason: rejectForm.reason })
    ElMessage.success('已拒绝')
    rejectVisible.value = false
    reloadCurrent()
  } catch (e) {
    ElMessage.error(e.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

function reloadCurrent() {
  if (showPendingOnly.value) loadPending()
  else loadList()
}

onMounted(() => loadList())
</script>

<style scoped>
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.amount { font-weight: 600; color: #e6a23c; }
.muted { color: #c0c4cc; }
</style>
