<template>
  <div class="app-container">
    <!-- 统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">总申请数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.pending }}</div><div class="stat-label">待审核</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.verified }}</div><div class="stat-label">已认证</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-danger">{{ stats.failed }}</div><div class="stat-label">认证失败</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="商户ID">
          <el-input v-model="filters.dh114_id" placeholder="商户ID" clearable style="width: 140px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="认证类型">
          <el-select v-model="filters.verification_type" placeholder="全部" clearable style="width: 160px" @change="onSearch">
            <el-option label="个人认证" value="personal" />
            <el-option label="企业认证" value="enterprise" />
            <el-option label="官方认证" value="official" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="未认证" :value="0" />
            <el-option label="待审核" :value="1" />
            <el-option label="已认证" :value="2" />
            <el-option label="认证失败" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="商户名/法人" clearable style="width: 180px" @keyup.enter="onSearch" />
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
        <el-button type="primary" :icon="Plus" @click="openCreate">新建认证</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="dh114_id" label="商户ID" width="90" />
        <el-table-column label="商户名" min-width="160">
          <template #default="{ row }">{{ row.business_name || `商户#${row.dh114_id}` }}</template>
        </el-table-column>
        <el-table-column label="认证类型" width="120">
          <template #default="{ row }">
            <el-tag :type="typeTag(row.verification_type)" size="small">{{ typeText(row.verification_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="法人/联系人" width="120">
          <template #default="{ row }">{{ row.legal_person || row.contact_name || '-' }}</template>
        </el-table-column>
        <el-table-column label="执照号" width="160">
          <template #default="{ row }">{{ row.license_no || '-' }}</template>
        </el-table-column>
        <el-table-column label="执照图片" width="90">
          <template #default="{ row }">
            <el-image v-if="row.license_image" :src="row.license_image" fit="cover" class="thumb" preview-teleported :preview-src-list="[row.license_image]" />
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="提交时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status === 1" type="success" link size="small" @click="onAudit(row, 2)">通过</el-button>
            <el-button v-if="row.status === 1" type="danger" link size="small" @click="onAudit(row, 3)">拒绝</el-button>
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
    <el-dialog v-model="detailVisible" title="认证详情" width="720px" destroy-on-close>
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="商户ID">{{ detail.dh114_id }}</el-descriptions-item>
        <el-descriptions-item label="商户名">{{ detail.business_name || `商户#${detail.dh114_id}` }}</el-descriptions-item>
        <el-descriptions-item label="认证类型">
          <el-tag :type="typeTag(detail.verification_type)" size="small">{{ typeText(detail.verification_type) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="法人">{{ detail.legal_person || '-' }}</el-descriptions-item>
        <el-descriptions-item label="联系电话">{{ maskPhone(detail.contact_phone) }}</el-descriptions-item>
        <el-descriptions-item label="执照号">{{ detail.license_no || '-' }}</el-descriptions-item>
        <el-descriptions-item label="执照图片" :span="2">
          <el-image v-if="detail.license_image" :src="detail.license_image" fit="cover" class="license-img" preview-teleported :preview-src-list="[detail.license_image]" />
          <span v-else class="text-muted">无</span>
        </el-descriptions-item>
        <el-descriptions-item label="证件号">{{ detail.id_card || '-' }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTag(detail.status)" size="small">{{ statusText(detail.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item v-if="detail.reason" label="原因" :span="2">{{ detail.reason }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.audit_remark" label="审核备注" :span="2">{{ detail.audit_remark }}</el-descriptions-item>
        <el-descriptions-item label="审核人">{{ detail.auditor_name || `#${detail.auditor_id}` }}</el-descriptions-item>
        <el-descriptions-item label="审核时间">{{ formatTime(detail.audited_at) }}</el-descriptions-item>
        <el-descriptions-item label="提交时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
      </el-descriptions>
      <template #footer><el-button @click="detailVisible = false">关闭</el-button></template>
    </el-dialog>

    <!-- 新建认证弹窗 -->
    <el-dialog v-model="formVisible" title="新建认证" width="600px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="商户ID" prop="dh114_id">
          <el-input-number v-model="form.dh114_id" :min="1" :controls="false" style="width: 100%" />
        </el-form-item>
        <el-form-item label="认证类型" prop="verification_type">
          <el-select v-model="form.verification_type" style="width: 100%">
            <el-option label="个人认证" value="personal" />
            <el-option label="企业认证" value="enterprise" />
            <el-option label="官方认证" value="official" />
          </el-select>
        </el-form-item>
        <el-form-item label="法人">
          <el-input v-model="form.legal_person" maxlength="64" />
        </el-form-item>
        <el-form-item label="联系电话">
          <el-input v-model="form.contact_phone" maxlength="20" />
        </el-form-item>
        <el-form-item label="执照号">
          <el-input v-model="form.license_no" maxlength="64" />
        </el-form-item>
        <el-form-item label="执照图片">
          <el-input v-model="form.license_image" placeholder="图片 URL" />
        </el-form-item>
        <el-form-item label="证件号">
          <el-input v-model="form.id_card" maxlength="32" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.reason" type="textarea" :rows="2" maxlength="200" />
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

const stats = reactive({ total: 0, pending: 0, verified: 0, failed: 0 })

const filters = reactive({ dh114_id: '', verification_type: '', status: null, keyword: '' })

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.dh114_id = ''
  filters.verification_type = ''
  filters.status = null
  filters.keyword = ''
  page.value = 1
  loadList()
}

const typeText = (t) => ({ personal: '个人', enterprise: '企业', official: '官方' }[t] || '-')
const typeTag = (t) => ({ personal: 'info', enterprise: 'primary', official: 'danger' }[t] || 'info')
const statusText = (s) => ({ 0: '未认证', 1: '待审核', 2: '已认证', 3: '认证失败' }[s] || '-')
const statusTag = (s) => ({ 0: 'info', 1: 'warning', 2: 'success', 3: 'danger' }[s] || 'info')

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
      verification_type: filters.verification_type || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status,
      keyword: filters.keyword || undefined
    }
    const res = await request.get('/dh114/admin/verifications', { params })
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
  const pending = list.value.filter((r) => r.status === 1).length
  const verified = list.value.filter((r) => r.status === 2).length
  const failed = list.value.filter((r) => r.status === 3).length
  Object.assign(stats, { total, pending, verified, failed })
}

// ===== 详情 =====
const detailVisible = ref(false)
const detail = ref(null)
const openDetail = (row) => { detail.value = row; detailVisible.value = true }

// ===== 审核 =====
const onAudit = async (row, status) => {
  try {
    if (status === 3) {
      const { value } = await ElMessageBox.prompt('请输入拒绝原因', '拒绝认证', {
        inputType: 'textarea',
        inputPlaceholder: '拒绝原因'
      })
      await request.put(`/dh114/admin/verifications/${row.id}/audit`, {
        status, audit_remark: value || ''
      })
    } else {
      await ElMessageBox.confirm('确认通过该认证申请？', '提示', { type: 'warning' })
      await request.put(`/dh114/admin/verifications/${row.id}/audit`, { status })
    }
    ElMessage.success('操作成功')
    await loadList()
  } catch (e) {
    // 取消
  }
}

const onDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确认删除该认证记录？', '提示', { type: 'warning' })
    await request.delete(`/dh114/verifications/${row.id}`)
    ElMessage.success('删除成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

// ===== 新建 =====
const formVisible = ref(false)
const formLoading = ref(false)
const formRef = ref(null)
const form = reactive({
  dh114_id: undefined, verification_type: 'enterprise', legal_person: '',
  contact_phone: '', license_no: '', license_image: '', id_card: '', reason: ''
})
const rules = {
  dh114_id: [{ required: true, message: '请输入商户ID', trigger: 'blur' }],
  verification_type: [{ required: true, message: '请选择认证类型', trigger: 'change' }]
}

const openCreate = () => {
  Object.assign(form, {
    dh114_id: undefined, verification_type: 'enterprise', legal_person: '',
    contact_phone: '', license_no: '', license_image: '', id_card: '', reason: ''
  })
  formVisible.value = true
}

const onSubmit = async () => {
  try {
    await formRef.value.validate()
    formLoading.value = true
    await request.post('/dh114/verifications', { ...form, status: 1 })
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
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; margin-bottom: 12px; }
.toolbar-left { display: flex; gap: 8px; }
.text-muted { color: #909399; }
.thumb { width: 50px; height: 50px; border-radius: 4px; border: 1px solid #ebeef5; }
.license-img { width: 200px; height: 140px; border-radius: 4px; border: 1px solid #ebeef5; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
