<template>
  <div class="app-container">
    <div class="page-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button type="primary" :icon="Plus" @click="openCreate">新建分站</el-button>
          <el-button :icon="Refresh" @click="loadList">刷新</el-button>
          <el-button :icon="CopyDocument" @click="copyDialogVisible = true">配置复制</el-button>
        </div>
        <div class="toolbar-right">
          <el-input
            v-model="search"
            placeholder="名称/域名/描述"
            clearable
            style="width: 200px"
            :prefix-icon="Search"
            @keyup.enter="onSearch"
            @clear="onSearch"
          />
          <el-select
            v-model="statusFilter"
            placeholder="状态"
            clearable
            style="width: 120px; margin-left: 8px"
            @change="onSearch"
          >
            <el-option label="已启用" :value="1" />
            <el-option label="已停用" :value="0" />
          </el-select>
          <el-button type="primary" :icon="Search" style="margin-left: 8px" @click="onSearch">搜索</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="region_id" label="地区ID" width="90" />
        <el-table-column label="分站名称" min-width="160">
          <template #default="{ row }">
            <div class="station-name">
              <el-image v-if="row.logo" :src="row.logo" fit="cover" class="station-logo" />
              <span>{{ row.name || '-' }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="domain" label="主域名" min-width="160" show-overflow-tooltip />
        <el-table-column prop="description" label="描述" min-width="180" show-overflow-tooltip />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">{{ row.status_text }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="170" />
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="openEdit(row)">编辑</el-button>
            <el-button link :type="row.status === 1 ? 'warning' : 'success'" size="small" @click="toggleStatus(row)">
              {{ row.status === 1 ? '停用' : '启用' }}
            </el-button>
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
    <el-dialog v-model="formVisible" :title="isEdit ? '编辑分站' : '新建分站'" width="560px" @closed="resetForm">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="地区ID" prop="region_id">
          <el-input-number v-model="form.region_id" :min="1" :disabled="isEdit" style="width: 100%" />
        </el-form-item>
        <el-form-item label="分站名称" prop="name">
          <el-input v-model="form.name" maxlength="100" placeholder="请输入分站名称" />
        </el-form-item>
        <el-form-item label="主域名" prop="domain">
          <el-input v-model="form.domain" maxlength="200" placeholder="如 bj.wuchang.com" />
        </el-form-item>
        <el-form-item label="Logo" prop="logo">
          <el-input v-model="form.logo" maxlength="255" placeholder="Logo URL" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="2" maxlength="1000" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">已启用</el-radio>
            <el-radio :value="0">已停用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="运营配置" prop="config_text">
          <el-input v-model="form.config_text" type="textarea" :rows="4" placeholder='JSON 格式，如 {"theme":"blue"}' />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 配置复制弹窗 -->
    <el-dialog v-model="copyDialogVisible" title="配置复制" width="480px">
      <el-form :model="copyForm" label-width="110px">
        <el-form-item label="源分站ID">
          <el-input-number v-model="copyForm.source_station_id" :min="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="目标分站ID">
          <el-input-number v-model="copyForm.target_station_id" :min="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="业务模块">
          <el-input v-model="copyForm.biz_module" maxlength="50" placeholder="留空则复制全部模块" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="copyDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="copySubmitting" @click="onCopy">开始复制</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Search, CopyDocument } from '@element-plus/icons-vue'
import {
  getTenantStationAdminList,
  createTenantStation,
  updateTenantStation,
  deleteTenantStation,
  updateTenantStationStatus,
  copyTenantStationConfig
} from '@/api/tenant'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const search = ref('')
const statusFilter = ref(null)

const formVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref(null)
const form = reactive({
  id: 0,
  region_id: 1,
  name: '',
  domain: '',
  logo: '',
  description: '',
  status: 1,
  config_text: ''
})
const rules = {
  region_id: [{ required: true, message: '请输入地区ID', trigger: 'blur' }],
  name: [{ required: true, message: '请输入分站名称', trigger: 'blur' }]
}

const copyDialogVisible = ref(false)
const copySubmitting = ref(false)
const copyForm = reactive({ source_station_id: 1, target_station_id: 1, biz_module: '' })

async function loadList() {
  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value, keyword: search.value }
    if (statusFilter.value !== null && statusFilter.value !== '') params.status = statusFilter.value
    const res = await getTenantStationAdminList(params)
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (e) {
    ElMessage.error('加载分站列表失败')
  } finally {
    loading.value = false
  }
}

function onSearch() {
  page.value = 1
  loadList()
}

function openCreate() {
  isEdit.value = false
  resetForm()
  formVisible.value = true
}

function openEdit(row) {
  isEdit.value = true
  Object.assign(form, {
    id: row.id,
    region_id: row.region_id,
    name: row.name,
    domain: row.domain,
    logo: row.logo,
    description: row.description,
    status: row.status,
    config_text: row.config ? JSON.stringify(row.config, null, 2) : ''
  })
  formVisible.value = true
}

function resetForm() {
  Object.assign(form, {
    id: 0, region_id: 1, name: '', domain: '', logo: '', description: '', status: 1, config_text: ''
  })
}

function parseConfigText(text) {
  if (!text || !text.trim()) return null
  try {
    return JSON.parse(text)
  } catch (e) {
    return null
  }
}

async function onSubmit() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      const payload = {
        region_id: form.region_id,
        name: form.name,
        domain: form.domain,
        logo: form.logo,
        description: form.description,
        status: form.status,
        config: parseConfigText(form.config_text)
      }
      if (isEdit.value) {
        await updateTenantStation(form.id, payload)
        ElMessage.success('更新成功')
      } else {
        await createTenantStation(payload)
        ElMessage.success('创建成功')
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

async function toggleStatus(row) {
  const next = row.status === 1 ? 0 : 1
  try {
    await ElMessageBox.confirm(`确定要${next === 1 ? '启用' : '停用'}分站「${row.name}」吗？`, '提示', { type: 'warning' })
    await updateTenantStationStatus(row.id, next)
    ElMessage.success('状态已更新')
    loadList()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('操作失败')
  }
}

async function onDelete(row) {
  try {
    await ElMessageBox.confirm(`确定要删除分站「${row.name}」吗？`, '警告', { type: 'warning' })
    await deleteTenantStation(row.id)
    ElMessage.success('删除成功')
    loadList()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

async function onCopy() {
  copySubmitting.value = true
  try {
    const res = await copyTenantStationConfig({
      source_station_id: copyForm.source_station_id,
      target_station_id: copyForm.target_station_id,
      biz_module: copyForm.biz_module
    })
    ElMessage.success(`配置复制完成，共复制 ${res.data?.copied_count || 0} 项`)
    copyDialogVisible.value = false
  } catch (e) {
    ElMessage.error(e.message || '配置复制失败')
  } finally {
    copySubmitting.value = false
  }
}

onMounted(() => loadList())
</script>

<style scoped>
.station-name { display: flex; align-items: center; gap: 8px; }
.station-logo { width: 32px; height: 32px; border-radius: 4px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
