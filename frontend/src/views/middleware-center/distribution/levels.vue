<template>
  <div class="app-container">
    <div class="page-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button type="primary" :icon="Plus" @click="openCreate">新建等级</el-button>
          <el-button :icon="Refresh" @click="loadList">刷新</el-button>
        </div>
        <div class="toolbar-right">
          <el-select
            v-model="statusFilter"
            placeholder="状态"
            clearable
            style="width: 140px"
            @change="onSearch"
          >
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
          </el-select>
          <el-button type="primary" :icon="Search" style="margin-left: 8px" @click="onSearch">搜索</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column label="等级值" width="100">
          <template #default="{ row }">
            <el-tag :type="levelTagType(row.level)" size="small">Lv.{{ row.level }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="等级名称" min-width="160" />
        <el-table-column label="升级所需佣金" width="140" align="right">
          <template #default="{ row }">¥{{ formatAmount(row.required_amount) }}</template>
        </el-table-column>
        <el-table-column label="默认佣金率" width="120" align="right">
          <template #default="{ row }">{{ formatRate(row.commission_rate) }}</template>
        </el-table-column>
        <el-table-column label="额外权益" min-width="200">
          <template #default="{ row }">
            <template v-if="row.extra_benefits && parseBenefits(row.extra_benefits).length">
              <el-tag
                v-for="(b, i) in parseBenefits(row.extra_benefits)"
                :key="i"
                size="small"
                style="margin-right: 4px"
              >{{ b.name || b.code }}</el-tag>
            </template>
            <span v-else class="muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="170" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="openEdit(row)">编辑</el-button>
            <el-button link type="danger" size="small" @click="onDelete(row)">删除</el-button>
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

    <!-- 新建/编辑弹窗 -->
    <el-dialog v-model="formVisible" :title="formMode === 'create' ? '新建等级' : '编辑等级'" width="560px">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="120px">
        <el-form-item label="等级值" prop="level">
          <el-input-number v-model="form.level" :min="1" :max="99" :disabled="formMode === 'edit'" />
          <span class="tip">如 1/2/3，等级值越大级别越高</span>
        </el-form-item>
        <el-form-item label="等级名称" prop="name">
          <el-input v-model="form.name" placeholder="如 普通合伙人" maxlength="64" />
        </el-form-item>
        <el-form-item label="升级所需佣金" prop="required_amount">
          <el-input-number v-model="form.required_amount" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="默认佣金率" prop="commission_rate">
          <el-input-number v-model="form.commission_rate" :min="0" :max="1" :step="0.01" :precision="4" />
          <span class="tip">0-1 之间，如 0.10 表示 10%</span>
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="额外权益">
          <div v-for="(b, i) in form.extraBenefits" :key="i" class="benefit-row">
            <el-input v-model="b.code" placeholder="权益代码" style="width: 120px" />
            <el-input v-model="b.name" placeholder="权益名称" style="width: 140px; margin-left: 4px" />
            <el-input v-model="b.desc" placeholder="描述" style="width: 180px; margin-left: 4px" />
            <el-button link type="danger" :icon="Delete" @click="removeBenefit(i)" style="margin-left: 4px" />
          </div>
          <el-button size="small" :icon="Plus" @click="addBenefit">添加权益</el-button>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onSubmit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Search, Delete } from '@element-plus/icons-vue'
import {
  getDistributionLevelList,
  createDistributionLevel,
  updateDistributionLevel,
  deleteDistributionLevel
} from '@/api/distribution'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const statusFilter = ref(null)

const formVisible = ref(false)
const formMode = ref('create')
const submitting = ref(false)
const formRef = ref(null)

const form = reactive({
  id: 0,
  level: 1,
  name: '',
  required_amount: 0,
  commission_rate: 0.1,
  status: 1,
  extraBenefits: []
})
const formRules = {
  level: [{ required: true, message: '请输入等级值', trigger: 'blur' }],
  name: [{ required: true, message: '请输入等级名称', trigger: 'blur' }]
}

function levelTagType(l) {
  return l >= 3 ? 'danger' : l === 2 ? 'warning' : 'info'
}
function formatAmount(n) {
  if (n === undefined || n === null) return '0.00'
  return Number(n).toFixed(2)
}
function formatRate(r) {
  if (r === undefined || r === null) return '0.00%'
  return (Number(r) * 100).toFixed(2) + '%'
}
function parseBenefits(b) {
  if (!b) return []
  if (Array.isArray(b)) return b
  try { return JSON.parse(typeof b === 'string' ? b : JSON.stringify(b)) } catch { return [] }
}

async function loadList() {
  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value }
    if (statusFilter.value !== null && statusFilter.value !== '') params.status = statusFilter.value
    const res = await getDistributionLevelList(params)
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (e) {
    ElMessage.error('加载等级列表失败')
  } finally {
    loading.value = false
  }
}

function onSearch() {
  page.value = 1
  loadList()
}

function openCreate() {
  formMode.value = 'create'
  Object.assign(form, {
    id: 0, level: 1, name: '', required_amount: 0, commission_rate: 0.1, status: 1, extraBenefits: []
  })
  formVisible.value = true
}

function openEdit(row) {
  formMode.value = 'edit'
  Object.assign(form, {
    id: row.id,
    level: row.level,
    name: row.name,
    required_amount: row.required_amount,
    commission_rate: row.commission_rate,
    status: row.status,
    extraBenefits: parseBenefits(row.extra_benefits).map(b => ({ ...b }))
  })
  formVisible.value = true
}

function addBenefit() {
  form.extraBenefits.push({ code: '', name: '', desc: '' })
}
function removeBenefit(i) {
  form.extraBenefits.splice(i, 1)
}

async function onSubmit() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      const payload = {
        level: form.level,
        name: form.name,
        required_amount: form.required_amount,
        commission_rate: form.commission_rate,
        status: form.status
      }
      if (form.extraBenefits.length) {
        payload.extra_benefits = form.extraBenefits.filter(b => b.code || b.name)
      }
      if (formMode.value === 'create') {
        await createDistributionLevel(payload)
        ElMessage.success('等级已创建')
      } else {
        await updateDistributionLevel(form.id, payload)
        ElMessage.success('等级已更新')
      }
      formVisible.value = false
      loadList()
    } catch (e) {
      ElMessage.error(e.message || '操作失败')
    } finally {
      submitting.value = false
    }
  })
}

async function onDelete(row) {
  try {
    await ElMessageBox.confirm(`确定删除等级「${row.name}」吗？`, '提示', { type: 'warning' })
    await deleteDistributionLevel(row.id)
    ElMessage.success('已删除')
    loadList()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error(e.message || '操作失败')
  }
}

onMounted(() => loadList())
</script>

<style scoped>
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.tip { margin-left: 8px; color: #909399; font-size: 12px; }
.muted { color: #c0c4cc; }
.benefit-row { margin-bottom: 6px; }
</style>
