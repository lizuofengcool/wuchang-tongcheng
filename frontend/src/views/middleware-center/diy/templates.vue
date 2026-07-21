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
          <el-input v-model="filters.keyword" placeholder="名称" clearable style="width: 200px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar">
        <el-button :icon="Refresh" @click="loadList">刷新</el-button>
        <el-button type="primary" :icon="Plus" @click="openCreate">新建模板</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="name" label="名称" min-width="180" show-overflow-tooltip />
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
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openEdit(row)">编辑</el-button>
            <el-button type="success" link size="small" @click="openApply(row)">应用</el-button>
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
    <el-dialog v-model="formVisible" :title="formTitle" width="780px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="名称" prop="name">
              <el-input v-model="form.name" maxlength="100" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="分类" prop="category">
              <el-select v-model="form.category" style="width: 100%">
                <el-option v-for="(label, val) in categoryMap" :key="val" :label="label" :value="val" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="缩略图">
              <el-input v-model="form.thumbnail" maxlength="500" placeholder="缩略图 URL" />
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
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="页面配置">
          <el-input v-model="form.pagesText" type="textarea" :rows="12" placeholder='JSON 格式，含 components/settings' />
          <div class="form-tip">模板页面配置 JSONB（应用模板时复制到新页面的 components/settings）</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 应用模板弹窗 -->
    <el-dialog v-model="applyVisible" :title="`应用模板 - ${applyTemplate.name}`" width="560px" destroy-on-close>
      <el-form ref="applyFormRef" :model="applyForm" :rules="applyRules" label-width="100px">
        <el-form-item label="新页面标题" prop="title">
          <el-input v-model="applyForm.title" maxlength="100" />
        </el-form-item>
        <el-form-item label="页面类型">
          <el-select v-model="applyForm.type" style="width: 100%">
            <el-option v-for="(label, val) in categoryMap" :key="val" :label="label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item label="Slug">
          <el-input v-model="applyForm.slug" maxlength="100" placeholder="可选，URL Slug" />
        </el-form-item>
        <el-form-item label="业务ID">
          <el-input-number v-model="applyForm.biz_id" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="applyForm.status" style="width: 100%">
            <el-option label="草稿" :value="0" />
            <el-option label="直接发布" :value="1" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="applyVisible = false">取消</el-button>
        <el-button type="primary" :loading="applying" @click="onApplySubmit">应用模板</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search, Plus } from '@element-plus/icons-vue'
import {
  getDiyTemplateList,
  createDiyTemplate,
  updateDiyTemplate,
  deleteDiyTemplate,
  applyDiyTemplate
} from '@/api/diy'

const loading = ref(false)
const submitting = ref(false)
const applying = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const filters = reactive({ category: '', status: null, keyword: '' })

const categoryMap = {
  home: '首页模板',
  topic: '专题页模板',
  shop: '店铺页模板',
  activity: '活动页模板'
}
const categoryTagType = (c) => ({
  home: 'danger',
  topic: 'warning',
  shop: 'success',
  activity: 'primary'
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
    const res = await getDiyTemplateList(params)
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
  category: 'home',
  thumbnail: '',
  description: '',
  status: 1,
  pagesText: '{"components":[],"settings":{}}'
})
const rules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  category: [{ required: true, message: '请选择分类', trigger: 'change' }]
}

const openCreate = () => {
  formTitle.value = '新建模板'
  editingId.value = null
  Object.assign(form, {
    name: '',
    category: 'home',
    thumbnail: '',
    description: '',
    status: 1,
    pagesText: '{"components":[],"settings":{}}'
  })
  formVisible.value = true
}

const openEdit = (row) => {
  formTitle.value = '编辑模板'
  editingId.value = row.id
  Object.assign(form, {
    name: row.name,
    category: row.category,
    thumbnail: row.thumbnail || '',
    description: row.description || '',
    status: row.status,
    pagesText: row.pages ? JSON.stringify(row.pages, null, 2) : '{"components":[],"settings":{}}'
  })
  formVisible.value = true
}

const onSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    let pagesObj
    try {
      pagesObj = JSON.parse(form.pagesText || '{}')
    } catch (e) {
      ElMessage.error('页面配置 JSON 格式错误')
      return
    }
    submitting.value = true
    try {
      const payload = {
        name: form.name,
        category: form.category,
        thumbnail: form.thumbnail,
        description: form.description,
        status: form.status,
        pages: pagesObj
      }
      if (editingId.value) {
        await updateDiyTemplate(editingId.value, payload)
        ElMessage.success('更新成功')
      } else {
        await createDiyTemplate(payload)
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
  ElMessageBox.confirm(`确认删除模板「${row.name}」？`, '提示', {
    type: 'warning',
    confirmButtonText: '确认',
    cancelButtonText: '取消'
  }).then(async () => {
    try {
      await deleteDiyTemplate(row.id)
      ElMessage.success('删除成功')
      loadList()
    } catch (e) {}
  }).catch(() => {})
}

// ===== 应用模板 =====
const applyVisible = ref(false)
const applyTemplate = ref({})
const applyFormRef = ref(null)
const applyForm = reactive({
  title: '',
  type: 'home',
  slug: '',
  biz_id: 0,
  status: 0
})
const applyRules = {
  title: [{ required: true, message: '请输入新页面标题', trigger: 'blur' }]
}

const openApply = (row) => {
  applyTemplate.value = row
  applyForm.title = row.name + ' - 副本'
  applyForm.type = row.category || 'home'
  applyForm.slug = ''
  applyForm.biz_id = 0
  applyForm.status = 0
  applyVisible.value = true
}

const onApplySubmit = async () => {
  if (!applyFormRef.value) return
  await applyFormRef.value.validate(async (valid) => {
    if (!valid) return
    applying.value = true
    try {
      await applyDiyTemplate(applyTemplate.value.id, {
        title: applyForm.title,
        type: applyForm.type,
        slug: applyForm.slug,
        biz_id: applyForm.biz_id,
        status: applyForm.status
      })
      ElMessage.success('应用成功，已创建新页面')
      applyVisible.value = false
    } catch (e) {
    } finally {
      applying.value = false
    }
  })
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
