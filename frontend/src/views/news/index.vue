<template>
  <div class="app-container">
    <div class="page-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button type="primary" :icon="Plus" @click="openCreate">发布头条</el-button>
          <el-button :icon="Refresh" @click="loadNews">刷新</el-button>
        </div>
        <div class="toolbar-right">
          <el-input
            v-model="search"
            placeholder="标题关键词"
            clearable
            style="width: 180px"
            :prefix-icon="Search"
            @keyup.enter="onSearch"
            @clear="onSearch"
          />
          <el-select
            v-model="categoryFilter"
            placeholder="分类"
            clearable
            style="width: 140px; margin-left: 8px"
            @change="onSearch"
          >
            <el-option v-for="c in flatCategories" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
          <el-select
            v-model="statusFilter"
            placeholder="状态"
            clearable
            style="width: 110px; margin-left: 8px"
            @change="onSearch"
          >
            <el-option label="草稿" :value="0" />
            <el-option label="已发布" :value="1" />
            <el-option label="已下架" :value="2" />
          </el-select>
          <el-button type="primary" :icon="Search" style="margin-left: 8px" @click="onSearch">搜索</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="title" label="标题" min-width="220" show-overflow-tooltip />
        <el-table-column label="分类" width="120">
          <template #default="{ row }">{{ categoryName(row.category_id) }}</template>
        </el-table-column>
        <el-table-column prop="author_name" label="作者" width="120" />
        <el-table-column prop="view_count" label="浏览" width="80" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="发布时间" width="170">
          <template #default="{ row }">{{ formatTime(row.published_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="goDetail(row)">查看</el-button>
            <el-button type="warning" link size="small" @click="openEdit(row)">编辑</el-button>
            <el-button
              v-if="row.status !== 1"
              type="success"
              link
              size="small"
              @click="handlePublish(row)"
            >发布</el-button>
            <el-button
              v-if="row.status === 1"
              type="info"
              link
              size="small"
              @click="handleOffline(row)"
            >下架</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
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
          @current-change="loadNews"
          @size-change="loadNews"
        />
      </div>
    </div>

    <!-- 新建/编辑 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑头条' : '发布头条'" width="780px" @close="onDialogClose">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="80px">
        <el-form-item label="标题" prop="title">
          <el-input v-model="form.title" placeholder="请输入标题" maxlength="200" show-word-limit />
        </el-form-item>
        <el-form-item label="分类" prop="category_id">
          <el-select v-model="form.category_id" placeholder="请选择分类" filterable clearable style="width: 100%">
            <el-option v-for="c in flatCategories" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="封面图" prop="cover_image">
          <div class="cover-uploader">
            <el-upload
              class="cover-uploader-trigger"
              :show-file-list="false"
              :before-upload="beforeCoverUpload"
              :http-request="handleCoverUpload"
              accept="image/*"
            >
              <div v-if="form.cover_image" class="cover-preview">
                <img :src="form.cover_image" alt="封面图" />
                <div class="cover-mask">点击替换</div>
              </div>
              <div v-else class="cover-placeholder">
                <el-icon size="28"><Plus /></el-icon>
                <span>上传封面</span>
              </div>
            </el-upload>
            <el-input
              v-model="form.cover_image"
              placeholder="或直接填写图片URL"
              style="flex:1; margin-left: 12px"
              clearable
            />
          </div>
          <div class="form-tip">建议尺寸 16:9，单张不超过 5MB</div>
        </el-form-item>
        <el-form-item label="摘要" prop="summary">
          <el-input v-model="form.summary" type="textarea" :rows="2" maxlength="500" show-word-limit />
        </el-form-item>
        <el-form-item label="标签" prop="tags">
          <el-input v-model="form.tags" placeholder="多个标签用英文逗号分隔" />
        </el-form-item>
        <el-form-item label="内容" prop="content">
          <RichTextEditor v-model="form.content" :min-height="320" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio :value="0">草稿</el-radio>
            <el-radio :value="1">发布</el-radio>
          </el-radio-group>
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
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Search } from '@element-plus/icons-vue'
import { listNews, createNews, updateNews, deleteNews } from '@/api/news'
import { getCategoryTree } from '@/api/category'
import { uploadFile } from '@/api/file'
import { newsStatusText as statusText, newsStatusTagType as statusTag, formatTime } from '@/utils/format'
import RichTextEditor from '@/components/RichTextEditor.vue'

const router = useRouter()

const loading = ref(false)
const search = ref('')
const categoryFilter = ref(null)
const statusFilter = ref(null)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const list = ref([])

// 分类树 + 扁平化
const categoryTree = ref([])
const flatCategories = computed(() => {
  const result = []
  const walk = (nodes) => {
    if (!Array.isArray(nodes)) return
    nodes.forEach((n) => {
      result.push({ id: n.id, name: n.name })
      if (n.children?.length) walk(n.children)
    })
  }
  walk(categoryTree.value)
  return result
})
const categoryName = (id) => flatCategories.value.find((c) => c.id === id)?.name || '-'

const loadCategories = async () => {
  try {
    const res = await getCategoryTree()
    categoryTree.value = res.data || []
  } catch (e) {
    categoryTree.value = []
  }
}

const onSearch = () => {
  page.value = 1
  loadNews()
}

const loadNews = async () => {
  loading.value = true
  try {
    const res = await listNews({
      page: page.value,
      page_size: pageSize.value,
      keyword: search.value.trim(),
      category_id: categoryFilter.value || undefined,
      status: statusFilter.value === null || statusFilter.value === '' ? undefined : statusFilter.value
    })
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
  } finally {
    loading.value = false
  }
}

// ===== 新建/编辑 =====
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref(null)
const form = reactive({
  id: 0,
  title: '',
  content: '',
  cover_image: '',
  summary: '',
  category_id: null,
  tags: '',
  status: 1
})
const formRules = {
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }]
}

// ===== 封面上传 =====
const coverUploading = ref(false)
const beforeCoverUpload = (file) => {
  if (!file.type.startsWith('image/')) {
    ElMessage.error('仅支持图片格式')
    return false
  }
  if (file.size > 5 * 1024 * 1024) {
    ElMessage.error('封面图不能超过 5MB')
    return false
  }
  return true
}
const handleCoverUpload = async (options) => {
  const { file } = options
  coverUploading.value = true
  try {
    const res = await uploadFile(file)
    if (res.data?.file_url) {
      form.cover_image = res.data.file_url
      ElMessage.success('封面上传成功')
    }
  } catch (e) {
    // 错误已由 request 拦截器提示
  } finally {
    coverUploading.value = false
  }
}

const openCreate = () => {
  isEdit.value = false
  Object.assign(form, { id: 0, title: '', content: '', cover_image: '', summary: '', category_id: null, tags: '', status: 1 })
  dialogVisible.value = true
}

// 跳转详情页
const goDetail = (row) => {
  router.push({ name: 'NewsDetail', params: { id: row.id } })
}

const openEdit = (row) => {
  isEdit.value = true
  Object.assign(form, {
    id: row.id,
    title: row.title || '',
    content: row.content || '',
    cover_image: row.cover_image || '',
    summary: row.summary || '',
    category_id: row.category_id || null,
    tags: row.tags || '',
    status: row.status === 1 ? 1 : 0
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
        await updateNews(form.id, {
          title: form.title,
          content: form.content,
          cover_image: form.cover_image,
          summary: form.summary,
          category_id: form.category_id || 0,
          tags: form.tags,
          status: form.status
        })
        ElMessage.success('更新成功')
      } else {
        await createNews({
          title: form.title,
          content: form.content,
          cover_image: form.cover_image,
          summary: form.summary,
          category_id: form.category_id || 0,
          tags: form.tags,
          status: form.status
        })
        ElMessage.success('发布成功')
      }
      dialogVisible.value = false
      await loadNews()
    } catch (e) {
      // ignore
    } finally {
      submitting.value = false
    }
  })
}

const handlePublish = async (row) => {
  try {
    await updateNews(row.id, { status: 1 })
    ElMessage.success('已发布')
    await loadNews()
  } catch (e) {
    // ignore
  }
}

const handleOffline = async (row) => {
  try {
    await updateNews(row.id, { status: 2 })
    ElMessage.success('已下架')
    await loadNews()
  } catch (e) {
    // ignore
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除头条 "${row.title}" 吗？`, '提示', { type: 'warning' })
    await deleteNews(row.id)
    ElMessage.success('删除成功')
    await loadNews()
  } catch (e) {
    // 取消
  }
}

onMounted(async () => {
  await loadCategories()
  await loadNews()
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
.cover-uploader {
  display: flex;
  align-items: center;
  width: 100%;
}
.cover-uploader-trigger {
  flex-shrink: 0;
}
.cover-preview,
.cover-placeholder {
  width: 160px;
  height: 90px;
  border: 1px dashed #dcdfe6;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  position: relative;
  background: #fafafa;
  cursor: pointer;
}
.cover-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.cover-mask {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  opacity: 0;
  transition: opacity 0.2s;
}
.cover-preview:hover .cover-mask {
  opacity: 1;
}
.cover-placeholder {
  flex-direction: column;
  color: #909399;
  font-size: 12px;
  gap: 4px;
}
.cover-placeholder:hover {
  border-color: #409eff;
  color: #409eff;
}
.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}
</style>
