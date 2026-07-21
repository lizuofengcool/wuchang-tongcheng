<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="位置编码">
          <el-select v-model="filters.position_code" placeholder="全部" clearable style="width: 160px" @change="onSearch">
            <el-option v-for="(label, val) in positionMap" :key="val" :label="label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="标题" clearable style="width: 200px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
            <el-option label="待生效" :value="2" />
            <el-option label="已过期" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar">
        <el-button :icon="Refresh" @click="loadList">刷新</el-button>
        <el-button type="primary" :icon="Plus" @click="openCreate">新建广告位</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column label="位置" width="140">
          <template #default="{ row }">
            <el-tag size="small" :type="positionTagType(row.position_code)">{{ positionMap[row.position_code] || row.position_code }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="title" label="标题" min-width="180" show-overflow-tooltip />
        <el-table-column label="图片" width="120">
          <template #default="{ row }">
            <el-image v-if="row.image_url" :src="row.image_url" :preview-src-list="[row.image_url]" fit="cover" style="width: 80px; height: 40px" />
            <span v-else class="text-muted">无</span>
          </template>
        </el-table-column>
        <el-table-column prop="link_url" label="跳转链接" min-width="180" show-overflow-tooltip />
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column label="生效时间" width="200">
          <template #default="{ row }">
            <div class="time-cell">
              <div>{{ formatTime(row.start_at) }}</div>
              <div>至 {{ formatTime(row.end_at) }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusMap[row.status] || '未知' }}</el-tag>
          </template>
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
    <el-dialog v-model="formVisible" :title="formTitle" width="640px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="位置编码" prop="position_code">
          <el-select v-model="form.position_code" style="width: 100%">
            <el-option v-for="(label, val) in positionMap" :key="val" :label="label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item label="标题" prop="title">
          <el-input v-model="form.title" maxlength="100" />
        </el-form-item>
        <el-form-item label="图片URL" prop="image_url">
          <el-input v-model="form.image_url" maxlength="500" placeholder="广告图片 URL" />
        </el-form-item>
        <el-form-item label="跳转链接">
          <el-input v-model="form.link_url" maxlength="500" placeholder="点击跳转 URL（可选）" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="排序">
              <el-input-number v-model="form.sort" :min="0" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="状态">
              <el-select v-model="form.status" style="width: 100%">
                <el-option label="启用" :value="1" />
                <el-option label="禁用" :value="0" />
                <el-option label="待生效" :value="2" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="生效时间">
          <el-date-picker
            v-model="form.dateRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            value-format="YYYY-MM-DDTHH:mm:ssZ"
            style="width: 100%"
          />
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
  getMarketingAdList,
  createMarketingAd,
  updateMarketingAd,
  deleteMarketingAd
} from '@/api/marketing'

const loading = ref(false)
const submitting = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const filters = reactive({ position_code: '', keyword: '', status: null })

const positionMap = {
  home_banner: '首页Banner',
  list_top: '列表置顶',
  detail_banner: '详情页广告',
  category_top: '分类页置顶',
  search_top: '搜索页置顶',
  popup: '弹窗广告'
}
const positionTagType = (code) => ({
  home_banner: 'danger',
  list_top: 'warning',
  detail_banner: 'success',
  category_top: 'primary',
  search_top: 'info',
  popup: 'danger'
}[code] || 'info')

const statusMap = { 0: '禁用', 1: '启用', 2: '待生效', 3: '已过期' }
const statusTagType = (s) => ({ 0: 'info', 1: 'success', 2: 'warning', 3: 'danger' }[s] || 'info')

const formatTime = (t) => {
  if (!t) return '不限'
  return String(t).replace('T', ' ').slice(0, 16)
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.position_code = ''
  filters.keyword = ''
  filters.status = null
  page.value = 1
  loadList()
}

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      position_code: filters.position_code || undefined,
      keyword: filters.keyword || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status
    }
    const res = await getMarketingAdList(params)
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
  position_code: 'home_banner',
  title: '',
  image_url: '',
  link_url: '',
  sort: 0,
  status: 1,
  dateRange: []
})
const rules = {
  position_code: [{ required: true, message: '请选择位置编码', trigger: 'change' }],
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
  image_url: [{ required: true, message: '请输入图片URL', trigger: 'blur' }]
}

const openCreate = () => {
  formTitle.value = '新建广告位'
  editingId.value = null
  Object.assign(form, {
    position_code: 'home_banner',
    title: '',
    image_url: '',
    link_url: '',
    sort: 0,
    status: 1,
    dateRange: []
  })
  formVisible.value = true
}

const openEdit = (row) => {
  formTitle.value = '编辑广告位'
  editingId.value = row.id
  Object.assign(form, {
    position_code: row.position_code,
    title: row.title,
    image_url: row.image_url,
    link_url: row.link_url || '',
    sort: row.sort || 0,
    status: row.status,
    dateRange: row.start_at && row.end_at ? [row.start_at, row.end_at] : []
  })
  formVisible.value = true
}

const onSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      const payload = {
        position_code: form.position_code,
        title: form.title,
        image_url: form.image_url,
        link_url: form.link_url,
        sort: form.sort,
        status: form.status,
        start_at: form.dateRange && form.dateRange[0] ? form.dateRange[0] : null,
        end_at: form.dateRange && form.dateRange[1] ? form.dateRange[1] : null
      }
      if (editingId.value) {
        await updateMarketingAd(editingId.value, payload)
        ElMessage.success('更新成功')
      } else {
        await createMarketingAd(payload)
        ElMessage.success('创建成功')
      }
      formVisible.value = false
      loadList()
    } catch (e) {
      // 错误已由拦截器处理
    } finally {
      submitting.value = false
    }
  })
}

const onDelete = (row) => {
  ElMessageBox.confirm(`确认删除广告位「${row.title}」？`, '提示', {
    type: 'warning',
    confirmButtonText: '确认',
    cancelButtonText: '取消'
  }).then(async () => {
    try {
      await deleteMarketingAd(row.id)
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
.time-cell { font-size: 12px; color: #666; line-height: 1.6; }
.text-muted { color: #999; }
</style>
