<template>
  <div class="app-container">
    <div class="page-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button type="primary" :icon="Plus" @click="openCreate">新增配置</el-button>
          <el-button :icon="Refresh" @click="loadAll">刷新</el-button>
          <el-button type="success" :icon="Check" :loading="batchSaving" @click="handleBatchSave">批量保存</el-button>
        </div>
        <div class="toolbar-right">
          <el-input
            v-model="search"
            placeholder="key/描述"
            clearable
            style="width: 200px"
            :prefix-icon="Search"
          />
          <el-select v-model="groupFilter" placeholder="分组" clearable style="width: 160px; margin-left: 8px">
            <el-option v-for="g in groups" :key="g" :label="g" :value="g" />
          </el-select>
        </div>
      </div>

      <el-table v-loading="loading" :data="filteredList" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="group" label="分组" width="130">
          <template #default="{ row }">
            <el-tag size="small">{{ row.group }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="key" label="配置项" width="200" show-overflow-tooltip />
        <el-table-column label="值" min-width="220">
          <template #default="{ row }">
            <el-input v-if="row._editing" v-model="row.value" size="small" />
            <span v-else class="value-text">{{ row.value || '(空)' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="typeTag(row.value_type)" size="small">{{ row.value_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column prop="description" label="描述" min-width="180" show-overflow-tooltip />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button v-if="!row._editing" type="primary" link size="small" @click="row._editing = true">编辑</el-button>
            <el-button v-else type="success" link size="small" :loading="row._saving" @click="saveRow(row)">保存</el-button>
            <el-button v-if="row._editing" link size="small" @click="cancelEdit(row)">取消</el-button>
            <el-button type="warning" link size="small" @click="openEdit(row)">详细</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 新建/详细编辑 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑配置' : '新增配置'" width="560px" @close="onDialogClose">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="90px">
        <el-form-item label="分组" prop="group">
          <el-input v-model="form.group" placeholder="如 site / business / sms" maxlength="50" />
          <el-select
            v-if="groups.length"
            v-model="form.group"
            placeholder="或选择已有分组"
            filterable
            allow-create
            style="width: 100%; margin-top: 4px"
          >
            <el-option v-for="g in groups" :key="g" :label="g" :value="g" />
          </el-select>
        </el-form-item>
        <el-form-item label="配置项" prop="key">
          <el-input v-model="form.key" placeholder="如 site_name" :disabled="isEdit" maxlength="100" show-word-limit />
        </el-form-item>
        <el-form-item label="值" prop="value">
          <el-input v-model="form.value" type="textarea" :rows="3" placeholder="配置值" />
        </el-form-item>
        <el-form-item label="值类型" prop="value_type">
          <el-radio-group v-model="form.value_type">
            <el-radio value="string">字符串</el-radio>
            <el-radio value="number">数字</el-radio>
            <el-radio value="bool">布尔</el-radio>
            <el-radio value="json">JSON</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" placeholder="可选描述" maxlength="255" show-word-limit />
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="form.sort" :min="0" :max="9999" controls-position="right" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Search, Check } from '@element-plus/icons-vue'
import {
  getAllSettings,
  createSetting,
  updateSetting,
  deleteSetting,
  batchUpdateSettings
} from '@/api/setting'

const loading = ref(false)
const search = ref('')
const groupFilter = ref('')
const list = ref([])

const typeTag = (t) => ({ string: '', number: 'success', bool: 'warning', json: 'info' }[t] || '')

const groups = computed(() => {
  const set = new Set()
  list.value.forEach((i) => set.add(i.group))
  return Array.from(set).sort()
})

const filteredList = computed(() => {
  let arr = list.value
  if (groupFilter.value) arr = arr.filter((i) => i.group === groupFilter.value)
  const kw = search.value.trim().toLowerCase()
  if (kw) {
    arr = arr.filter(
      (i) => (i.key || '').toLowerCase().includes(kw) || (i.description || '').toLowerCase().includes(kw)
    )
  }
  return arr
})

const loadAll = async () => {
  loading.value = true
  try {
    const res = await getAllSettings()
    const data = res.data || []
    list.value = data.map((i) => ({ ...i, _editing: false, _saving: false, _origin: i.value }))
  } catch (e) {
    list.value = []
  } finally {
    loading.value = false
  }
}

// ===== 行内编辑 =====
const saveRow = async (row) => {
  row._saving = true
  try {
    await updateSetting(row.id, { value: row.value })
    row._origin = row.value
    row._editing = false
    ElMessage.success('已保存')
  } catch (e) {
    // ignore
  } finally {
    row._saving = false
  }
}

const cancelEdit = (row) => {
  row.value = row._origin
  row._editing = false
}

// 批量保存所有已编辑项
const batchSaving = ref(false)
const handleBatchSave = async () => {
  const edited = list.value.filter((r) => r._editing && r.value !== r._origin)
  if (!edited.length) {
    ElMessage.info('没有修改项')
    return
  }
  try {
    await ElMessageBox.confirm(`确定批量保存 ${edited.length} 项修改吗？`, '提示', { type: 'warning' })
    batchSaving.value = true
    await batchUpdateSettings(edited.map((i) => ({ key: i.key, value: i.value })))
    edited.forEach((r) => {
      r._origin = r.value
      r._editing = false
    })
    ElMessage.success('批量保存成功')
  } catch (e) {
    // 取消
  } finally {
    batchSaving.value = false
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除配置 "${row.key}" 吗？`, '提示', { type: 'warning' })
    await deleteSetting(row.id)
    ElMessage.success('删除成功')
    await loadAll()
  } catch (e) {
    // 取消
  }
}

// ===== 新建/详细编辑 =====
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref(null)
const form = reactive({
  id: 0,
  group: '',
  key: '',
  value: '',
  value_type: 'string',
  description: '',
  sort: 0
})
const formRules = {
  group: [{ required: true, message: '请输入分组', trigger: 'blur' }],
  key: [{ required: true, message: '请输入配置项', trigger: 'blur' }],
  value_type: [{ required: true, message: '请选择值类型', trigger: 'change' }]
}

const resetForm = () => {
  Object.assign(form, { id: 0, group: '', key: '', value: '', value_type: 'string', description: '', sort: 0 })
}

const openCreate = () => {
  isEdit.value = false
  resetForm()
  dialogVisible.value = true
}

const openEdit = (row) => {
  isEdit.value = true
  Object.assign(form, {
    id: row.id,
    group: row.group || '',
    key: row.key || '',
    value: row.value || '',
    value_type: row.value_type || 'string',
    description: row.description || '',
    sort: row.sort ?? 0
  })
  dialogVisible.value = true
}

const onDialogClose = () => {
  formRef.value?.clearValidate()
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      if (isEdit.value) {
        await updateSetting(form.id, {
          value: form.value,
          description: form.description,
          sort: form.sort
        })
        ElMessage.success('更新成功')
      } else {
        await createSetting({
          group: form.group,
          key: form.key,
          value: form.value,
          value_type: form.value_type,
          description: form.description,
          sort: form.sort
        })
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      await loadAll()
    } catch (e) {
      // ignore
    } finally {
      submitting.value = false
    }
  })
}

onMounted(() => {
  loadAll()
})
</script>

<style scoped>
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}
.toolbar-left {
  display: flex;
  gap: 8px;
}
.toolbar-right {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
}
.value-text {
  display: inline-block;
  max-width: 100%;
  word-break: break-all;
  color: #303133;
}
</style>
