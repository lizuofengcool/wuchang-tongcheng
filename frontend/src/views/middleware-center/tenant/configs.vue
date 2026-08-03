<template>
  <div class="app-container">
    <div class="page-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button type="primary" :icon="Plus" @click="openCreate">新建配置</el-button>
          <el-button :icon="Refresh" @click="loadList">刷新</el-button>
        </div>
        <div class="toolbar-right">
          <el-select
            v-model="stationFilter"
            placeholder="选择分站"
            clearable
            filterable
            style="width: 200px"
            @change="onSearch"
          >
            <el-option v-for="s in stations" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
          <el-input
            v-model="bizModuleFilter"
            placeholder="业务模块"
            clearable
            style="width: 160px; margin-left: 8px"
            @keyup.enter="onSearch"
            @clear="onSearch"
          />
          <el-input
            v-model="search"
            placeholder="配置键关键词"
            clearable
            style="width: 200px; margin-left: 8px"
            :prefix-icon="Search"
            @keyup.enter="onSearch"
            @clear="onSearch"
          />
          <el-button type="primary" :icon="Search" style="margin-left: 8px" @click="onSearch">搜索</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column label="所属分站" min-width="150">
          <template #default="{ row }">{{ stationName(row.station_id) }}</template>
        </el-table-column>
        <el-table-column prop="biz_module" label="业务模块" width="140">
          <template #default="{ row }">
            <el-tag size="small" type="primary">{{ row.biz_module }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="config_key" label="配置键" min-width="180" show-overflow-tooltip />
        <el-table-column prop="config_value" label="配置值" min-width="240" show-overflow-tooltip />
        <el-table-column prop="updated_at" label="更新时间" width="170" />
        <el-table-column label="操作" width="160" fixed="right">
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
    <el-dialog v-model="formVisible" :title="isEdit ? '编辑配置' : '新建配置'" width="560px" @closed="resetForm">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="所属分站" prop="station_id">
          <el-select v-model="form.station_id" filterable placeholder="请选择分站" style="width: 100%" :disabled="isEdit">
            <el-option v-for="s in stations" :key="s.id" :label="`${s.name}（ID:${s.id}）`" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="业务模块" prop="biz_module">
          <el-input v-model="form.biz_module" maxlength="50" placeholder="如 dh114/mall/ershou" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="配置键" prop="config_key">
          <el-input v-model="form.config_key" maxlength="100" placeholder="如 theme_color/contact_phone" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="配置值" prop="config_value">
          <el-input v-model="form.config_value" type="textarea" :rows="5" placeholder="配置值（纯文本或 JSON 字符串）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Search } from '@element-plus/icons-vue'
import {
  getTenantConfigList,
  upsertTenantConfig,
  updateTenantConfig,
  deleteTenantConfig,
  getTenantStationAdminList
} from '@/api/tenant'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const search = ref('')
const stationFilter = ref(null)
const bizModuleFilter = ref('')
const stations = ref([])

const formVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref(null)
const form = reactive({
  id: 0,
  station_id: null,
  biz_module: '',
  config_key: '',
  config_value: ''
})
const rules = {
  station_id: [{ required: true, message: '请选择分站', trigger: 'change' }],
  biz_module: [{ required: true, message: '请输入业务模块', trigger: 'blur' }],
  config_key: [{ required: true, message: '请输入配置键', trigger: 'blur' }]
}

function stationName(id) {
  const s = stations.value.find((x) => x.id === id)
  return s ? s.name : `#${id}`
}

async function loadStations() {
  try {
    const res = await getTenantStationAdminList({ page: 1, page_size: 200 })
    stations.value = res.data?.list || []
  } catch (e) {
    stations.value = []
  }
}

async function loadList() {
  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value, keyword: search.value }
    if (stationFilter.value) params.station_id = stationFilter.value
    if (bizModuleFilter.value) params.biz_module = bizModuleFilter.value
    const res = await getTenantConfigList(params)
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (e) {
    ElMessage.error('加载配置列表失败')
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
    station_id: row.station_id,
    biz_module: row.biz_module,
    config_key: row.config_key,
    config_value: row.config_value || ''
  })
  formVisible.value = true
}

function resetForm() {
  Object.assign(form, {
    id: 0, station_id: null, biz_module: '', config_key: '', config_value: ''
  })
}

async function onSubmit() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      if (isEdit.value) {
        // 编辑仅允许更新配置值（后端 UpdateConfigRequest 仅接收 config_value）
        await updateTenantConfig(form.id, { config_value: form.config_value })
        ElMessage.success('更新成功')
      } else {
        // 新建走 upsert（按 station_id + biz_module + config_key 唯一）
        await upsertTenantConfig({
          station_id: form.station_id,
          biz_module: form.biz_module,
          config_key: form.config_key,
          config_value: form.config_value
        })
        ElMessage.success('保存成功')
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
    await ElMessageBox.confirm(`确定要删除配置「${row.biz_module}/${row.config_key}」吗？`, '警告', { type: 'warning' })
    await deleteTenantConfig(row.id)
    ElMessage.success('删除成功')
    loadList()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

onMounted(async () => {
  await loadStations()
  loadList()
})
</script>

<style scoped>
.page-card { background: #fff; padding: 16px; border-radius: 8px; box-shadow: 0 1px 4px rgba(0,0,0,0.04); }
.toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; flex-wrap: wrap; gap: 8px; }
.toolbar-left { display: flex; gap: 8px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
