<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="品牌/车系/车型" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="品牌">
          <el-select v-model="filters.brand_id" placeholder="全部品牌" clearable filterable style="width: 160px" @change="onSearch">
            <el-option v-for="b in brands" :key="b.id" :label="b.name" :value="b.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="filters.category_id" placeholder="全部分类" clearable style="width: 160px" @change="onSearch">
            <el-option v-for="c in categories" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar"><el-button :icon="Refresh" @click="loadList">刷新</el-button></div>

      <el-tabs v-model="activeTab" @tab-change="onTabChange">
        <el-tab-pane label="车型库" name="models">
          <el-table v-loading="loading" :data="list" border stripe>
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="name" label="车型名称" min-width="160" />
            <el-table-column prop="brand_name" label="品牌" width="120" />
            <el-table-column prop="series" label="车系" width="120" />
            <el-table-column prop="year" label="年份" width="80" />
            <el-table-column label="指导价" width="120">
              <template #default="{ row }">¥{{ Number(row.guide_price || 0).toFixed(2) }}万</template>
            </el-table-column>
            <el-table-column prop="category_name" label="分类" width="100" />
            <el-table-column prop="level" label="级别" width="100" />
            <el-table-column label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
        <el-tab-pane label="品牌库" name="brands">
          <el-table v-loading="loading" :data="brands" border stripe>
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="name" label="品牌名称" min-width="160" />
            <el-table-column prop="logo" label="Logo" width="80">
              <template #default="{ row }">
                <el-image v-if="row.logo" :src="row.logo" fit="cover" style="width: 32px; height: 32px" />
                <span v-else class="text-muted">-</span>
              </template>
            </el-table-column>
            <el-table-column prop="country" label="国家" width="100" />
            <el-table-column prop="initial" label="首字母" width="80" />
            <el-table-column prop="model_count" label="车型数" width="80" />
            <el-table-column label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
        <el-tab-pane label="分类库" name="categories">
          <el-table v-loading="loading" :data="categories" border stripe row-key="id" default-expand-all>
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="name" label="分类名称" min-width="160" />
            <el-table-column prop="level" label="级别" width="80" />
            <el-table-column prop="parent_name" label="父分类" width="120" />
            <el-table-column prop="sort" label="排序" width="80" />
            <el-table-column label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>

      <div class="pagination-wrap" v-if="activeTab === 'models'">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, RefreshLeft, Refresh } from '@element-plus/icons-vue'
import { listModels, listAllBrands, listCategories } from '@/api/car'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filters = reactive({ keyword: '', brand_id: '', category_id: '' })

const activeTab = ref('models')
const brands = ref([])
const categories = ref([])

const buildParams = () => {
  const p = { page: page.value, page_size: pageSize.value }
  if (filters.keyword) p.keyword = filters.keyword
  if (filters.brand_id) p.brand_id = filters.brand_id
  if (filters.category_id) p.category_id = filters.category_id
  return p
}

const loadList = async () => {
  if (activeTab.value === 'models') {
    loading.value = true
    try {
      const res = await listModels(buildParams())
      const d = res.data || {}
      list.value = d.list || []
      total.value = d.total || 0
    } catch (e) { list.value = [] } finally { loading.value = false }
  } else if (activeTab.value === 'brands') {
    loading.value = true
    try {
      const res = await listAllBrands()
      brands.value = res.data || []
    } catch (e) { brands.value = [] } finally { loading.value = false }
  } else if (activeTab.value === 'categories') {
    loading.value = true
    try {
      const res = await listCategories({ page: 1, page_size: 200 })
      categories.value = res.data?.list || res.data || []
    } catch (e) { categories.value = [] } finally { loading.value = false }
  }
}

const onSearch = () => { if (activeTab.value === 'models') { page.value = 1; loadList() } }
const onReset = () => { Object.assign(filters, { keyword: '', brand_id: '', category_id: '' }); onSearch() }

const onTabChange = () => { loadList() }

const loadBrandsAndCategories = async () => {
  try {
    const [b, c] = await Promise.all([listAllBrands(), listCategories({ page: 1, page_size: 200 })])
    brands.value = b.data || []
    categories.value = c.data?.list || c.data || []
  } catch (e) {
    ElMessage.error('加载品牌和分类失败')
  }
}

onMounted(async () => {
  await loadBrandsAndCategories()
  loadList()
})
</script>

<style scoped>
.page-card { background: #fff; padding: 16px; border-radius: 4px; box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04); }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: flex-start; margin-bottom: 12px; }
.text-muted { color: #909399; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
