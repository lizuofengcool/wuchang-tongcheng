<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="分类">
          <el-select v-model="filters.category" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option v-for="(label, val) in categoryMap" :key="val" :label="label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="名称/编码" clearable style="width: 200px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar">
        <el-button :icon="Refresh" @click="loadList">刷新</el-button>
        <el-button type="primary" :icon="Plus" @click="openCreate">新建组件</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="name" label="名称" min-width="140" show-overflow-tooltip />
        <el-table-column prop="code" label="编码" min-width="140" show-overflow-tooltip />
        <el-table-column label="分类" width="110">
          <template #default="{ row }">
            <el-tag size="small" :type="categoryTagType(row.category)">{{ row.category_text || categoryMap[row.category] || row.category }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="缩略图" width="100">
          <template #default="{ row }">
            <el-image v-if="row.thumbnail" :src="row.thumbnail" :preview-src-list="[row.thumbnail]" fit="cover" style="width: 60px; height: 40px" />
            <span v-else class="text-muted">无</span>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">{{ row.status_text || (row.status === 1 ? '启用' : '禁用') }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="更新时间" width="160">
          <template #default="{ row }">{{ formatTime(row.updated_at) }}</template>
        </el-table-column>
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
    </div>

    <!-- 表单弹窗 -->
    <el-dialog v-model="formVisible" :title="formTitle" width="720px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="名称" prop="name">
              <el-input v-model="form.name" maxlength="64" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="编码" prop="code">
              <el-input v-model="form.code" maxlength="64" :disabled="!!editingId" placeholder="唯一编码" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="分类" prop="category">
              <el-select v-model="form.category" style="width: 100%">
                <el-option v-for="(label, val) in categoryMap" :key="val" :label="label" :value="val" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="状态">
              <el-select v-model="form.status" style="width: 100%">
                <el-option label="启用" :value="1" />
                <el-option label="禁用" :value="0" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="缩略图">
          <el-input v-model="form.thumbnail" maxlength="500" placeholder="缩略图 URL" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="配置">
          <el-input v-model="form.configText" type="textarea" :rows="10" placeholder='JSON 格式，如 {"props":{}}' />
          <div class="form-tip">组件配置模板 JSONB（默认属性 schema，编辑器初始化使用）</div>
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
import { Refresh, RefreshLeft, Search, Plus } from '@element-plus/icons-vue'
import {
  getDiyComponentList,
  createDiyComponent,
  updateDiyComponent,
  deleteDiyComponent
} from '@/api/diy'

const loading = ref(false)
const submitting = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const filters = reactive({ category: '', status: null, keyword: '' })

const categoryMap = {
  basic: '基础组件',
  layout: '布局组件',
  business: '业务组件'
}
const categoryTagType = (c) => ({
  basic: '',
  layout: 'warning',
  business: 'success'
}[c] || 'info')

const formatTime = (t) => {
  if (!t) return '-'
  return String(t).replace('T', ' ').slice(0, 16)
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.category = ''
  filters.status = null
  filters.keyword = ''
  page.value = 1
  loadList()
}

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      category: filters.category || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status,
      keyword: filters.keyword || undefined
    }
    const res = await getDiyComponentList(params)
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

// ===== 表单 =====
const formRef = ref(null)
const formVisible = ref(false)
const formTitle = ref('')
const editingId = ref(null)
const form = reactive({
  name: '',
  code: '',
  category: 'basic',
  description: '',
  thumbnail: '',
  status: 1,
  configText: '{}'
})
const rules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入编码', trigger: 'blur' }],
  category: [{ required: true, message: '请选择分类', trigger: 'change' }]
}

const openCreate = () => {
  formTitle.value = '新建组件'
  editingId.value = null
  Object.assign(form, {
    name: '',
    code: '',
    category: 'basic',
    description: '',
    thumbnail: '',
    status: 1,
    configText: '{}'
  })
  formVisible.value = true
}

const openEdit = (row) => {
  formTitle.value = '编辑组件'
  editingId.value = row.id
  Object.assign(form, {
    name: row.name,
    code: row.code,
    category: row.category,
    description: row.description || '',
    thumbnail: row.thumbnail || '',
    status: row.status,
    configText: row.config ? JSON.stringify(row.config, null, 2) : '{}'
  })
  formVisible.value = true
}

const onSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    let configObj
    try {
      configObj = JSON.parse(form.configText || '{}')
    } catch (e) {
      ElMessage.error('配置 JSON 格式错误')
      return
    }
    submitting.value = true
    try {
      const payload = {
        name: form.name,
        category: form.category,
        description: form.description,
        thumbnail: form.thumbnail,
        status: form.status,
        config: configObj
      }
      if (editingId.value) {
        await updateDiyComponent(editingId.value, payload)
        ElMessage.success('更新成功')
      } else {
        payload.code = form.code
        await createDiyComponent(payload)
        ElMessage.success('创建成功')
      }
      formVisible.value = false
      loadList()
    } catch (e) {
    } finally {
      submitting.value = false
    }
  })
}

const onDelete = (row) => {
  ElMessageBox.confirm(`确认删除组件「${row.name}」？`, '提示', {
    type: 'warning',
    confirmButtonText: '确认',
    cancelButtonText: '取消'
  }).then(async () => {
    try {
      await deleteDiyComponent(row.id)
      ElMessage.success('删除成功')
      loadList()
    } catch (e) {}
  }).catch(() => {})
}

onMounted(() => {
  loadList()
})
</script>

<style scoped>
.page-card { background: #fff; padding: 16px; border-radius: 8px; box-shadow: 0 1px 4px rgba(0,0,0,0.04); }
.filter-form { margin-bottom: 12px; }
.toolbar { margin-bottom: 12px; display: flex; gap: 8px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.text-muted { color: #999; }
.form-tip { font-size: 12px; color: #909399; line-height: 1.4; margin-top: 4px; }
</style>
