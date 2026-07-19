<template>
  <div class="app-container">
    <div class="page-card">
      <el-tabs v-model="activeTab">
        <el-tab-pane label="品牌管理" name="brands">
          <el-form :inline="true" :model="brandFilters" class="filter-form" @submit.prevent>
            <el-form-item label="名称">
              <el-input v-model="brandFilters.keyword" placeholder="品牌名" clearable style="width: 180px" @keyup.enter="loadBrands" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :icon="Search" @click="loadBrands">搜索</el-button>
              <el-button type="primary" :icon="Plus" @click="openBrandCreate">新建品牌</el-button>
            </el-form-item>
          </el-form>

          <el-table v-loading="brandLoading" :data="brands" border stripe>
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column label="Logo" width="70">
              <template #default="{ row }">
                <el-image v-if="row.logo" :src="row.logo" fit="cover" class="logo-thumb" />
                <div v-else class="logo-thumb logo-empty">无</div>
              </template>
            </el-table-column>
            <el-table-column prop="name" label="品牌名" min-width="140" />
            <el-table-column prop="english_name" label="英文名" width="140" />
            <el-table-column prop="country" label="国家" width="100" />
            <el-table-column label="官方认证" width="100">
              <template #default="{ row }">
                <el-tag v-if="row.official_verified" type="success" size="small">已认证</el-tag>
                <span v-else class="text-muted">-</span>
              </template>
            </el-table-column>
            <el-table-column prop="use_count" label="使用次数" width="100" />
            <el-table-column prop="sort" label="排序" width="80" />
            <el-table-column label="操作" width="220" fixed="right">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="openBrandEdit(row)">编辑</el-button>
                <el-button type="success" link size="small" @click="openModels(row)">型号管理</el-button>
                <el-button type="danger" link size="small" @click="onDeleteBrand(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="型号管理" name="models">
          <el-form :inline="true" :model="modelFilters" class="filter-form" @submit.prevent>
            <el-form-item label="品牌">
              <el-select v-model="modelFilters.brand_id" placeholder="选择品牌" clearable filterable style="width: 200px" @change="loadModels">
                <el-option v-for="b in brands" :key="b.id" :label="b.name" :value="b.id" />
              </el-select>
            </el-form-item>
            <el-form-item label="名称">
              <el-input v-model="modelFilters.keyword" placeholder="型号名" clearable style="width: 180px" @keyup.enter="loadModels" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :icon="Search" @click="loadModels">搜索</el-button>
              <el-button type="primary" :icon="Plus" :disabled="!modelFilters.brand_id" @click="openModelCreate">新型号</el-button>
            </el-form-item>
          </el-form>

          <el-table v-loading="modelLoading" :data="models" border stripe>
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="name" label="型号名" min-width="140" />
            <el-table-column prop="full_name" label="全称" min-width="180" />
            <el-table-column label="图片" width="70">
              <template #default="{ row }">
                <el-image v-if="row.image" :src="row.image" fit="cover" class="logo-thumb" />
                <span v-else class="text-muted">-</span>
              </template>
            </el-table-column>
            <el-table-column prop="release_date" label="发布日期" width="120" />
            <el-table-column label="参考价" width="110">
              <template #default="{ row }">¥{{ Number(row.reference_price || 0).toFixed(2) }}</template>
            </el-table-column>
            <el-table-column prop="use_count" label="使用次数" width="100" />
            <el-table-column prop="sort" label="排序" width="80" />
            <el-table-column label="规格" min-width="200">
              <template #default="{ row }">
                <span v-if="row.specifications">{{ JSON.stringify(row.specifications) }}</span>
                <span v-else class="text-muted">-</span>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="140" fixed="right">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="openModelEdit(row)">编辑</el-button>
                <el-button type="danger" link size="small" @click="onDeleteModel(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </div>

    <!-- 品牌表单 -->
    <el-dialog v-model="brandFormVisible" :title="brandFormTitle" width="600px">
      <el-form ref="brandFormRef" :model="brandForm" :rules="brandRules" label-width="100px">
        <el-form-item label="品牌名" prop="name">
          <el-input v-model="brandForm.name" maxlength="128" show-word-limit />
        </el-form-item>
        <el-form-item label="英文名">
          <el-input v-model="brandForm.english_name" maxlength="128" />
        </el-form-item>
        <el-form-item label="Logo URL">
          <el-input v-model="brandForm.logo" placeholder="Logo图片URL" />
        </el-form-item>
        <el-form-item label="国家">
          <el-input v-model="brandForm.country" maxlength="32" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="brandForm.description" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="官网URL">
          <el-input v-model="brandForm.official_url" />
        </el-form-item>
        <el-form-item label="官方认证">
          <el-switch v-model="brandForm.official_verified" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="brandForm.sort" :min="0" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="brandForm.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="brandFormVisible = false">取消</el-button>
        <el-button type="primary" :loading="brandFormLoading" @click="onSubmitBrand">确认</el-button>
      </template>
    </el-dialog>

    <!-- 型号表单 -->
    <el-dialog v-model="modelFormVisible" :title="modelFormTitle" width="600px">
      <el-form ref="modelFormRef" :model="modelForm" :rules="modelRules" label-width="100px">
        <el-form-item label="所属品牌">
          <el-tag>{{ currentBrand?.name || '-' }}</el-tag>
        </el-form-item>
        <el-form-item label="型号名" prop="name">
          <el-input v-model="modelForm.name" maxlength="128" />
        </el-form-item>
        <el-form-item label="全称">
          <el-input v-model="modelForm.full_name" maxlength="255" />
        </el-form-item>
        <el-form-item label="图片">
          <el-input v-model="modelForm.image" placeholder="图片URL" />
        </el-form-item>
        <el-form-item label="发布日期">
          <el-input v-model="modelForm.release_date" placeholder="如 2024-01" />
        </el-form-item>
        <el-form-item label="参考价">
          <el-input-number v-model="modelForm.reference_price" :min="0" :precision="2" :controls="false" style="width: 100%" />
        </el-form-item>
        <el-form-item label="规格参数">
          <el-input v-model="modelForm.specificationsStr" type="textarea" :rows="4" placeholder='JSON 格式，如 {"屏幕":"6.1寸","内存":"8GB"}' />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="modelForm.sort" :min="0" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="modelForm.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="modelFormVisible = false">取消</el-button>
        <el-button type="primary" :loading="modelFormLoading" @click="onSubmitModel">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Plus } from '@element-plus/icons-vue'
import {
  listErshouBrands, createErshouBrand, updateErshouBrand, deleteErshouBrand,
  listErshouModels, createErshouModel, updateErshouModel, deleteErshouModel
} from '@/api/ershou'
import { formatTime } from '@/utils/format'

const activeTab = ref('brands')

// ===== 品牌 =====
const brandLoading = ref(false)
const brands = ref([])
const brandFilters = reactive({ keyword: '' })

const loadBrands = async () => {
  brandLoading.value = true
  try {
    const res = await listErshouBrands({ page: 1, page_size: 200, keyword: brandFilters.keyword || undefined })
    const data = res.data || {}
    brands.value = data.list || data || []
  } catch (e) {
    brands.value = []
  } finally {
    brandLoading.value = false
  }
}

const brandFormVisible = ref(false)
const brandFormLoading = ref(false)
const brandFormRef = ref(null)
const brandIsEdit = ref(false)
const brandFormTitle = computed(() => brandIsEdit.value ? '编辑品牌' : '新建品牌')
const brandForm = reactive({
  id: null, name: '', english_name: '', logo: '', country: '',
  description: '', official_url: '', official_verified: false,
  sort: 0, status: 1
})
const brandRules = { name: [{ required: true, message: '请输入品牌名', trigger: 'blur' }] }

const openBrandCreate = () => {
  brandIsEdit.value = false
  Object.assign(brandForm, {
    id: null, name: '', english_name: '', logo: '', country: '',
    description: '', official_url: '', official_verified: false, sort: 0, status: 1
  })
  brandFormVisible.value = true
}

const openBrandEdit = (row) => {
  brandIsEdit.value = true
  Object.assign(brandForm, { ...row })
  brandFormVisible.value = true
}

const onSubmitBrand = async () => {
  try {
    await brandFormRef.value.validate()
    brandFormLoading.value = true
    if (brandIsEdit.value) {
      await updateErshouBrand(brandForm.id, { ...brandForm })
    } else {
      await createErshouBrand({ ...brandForm })
    }
    ElMessage.success(brandIsEdit.value ? '更新成功' : '创建成功')
    brandFormVisible.value = false
    await loadBrands()
  } catch (e) {
    // ignore
  } finally {
    brandFormLoading.value = false
  }
}

const onDeleteBrand = async (row) => {
  try {
    await ElMessageBox.confirm(`确认删除品牌 "${row.name}"？删除后不可恢复`, '提示', { type: 'warning' })
    await deleteErshouBrand(row.id)
    ElMessage.success('删除成功')
    await loadBrands()
  } catch (e) { /* cancel */ }
}

// ===== 型号 =====
const modelLoading = ref(false)
const models = ref([])
const modelFilters = reactive({ brand_id: null, keyword: '' })
const currentBrand = ref(null)

const loadModels = async () => {
  if (!modelFilters.brand_id) {
    models.value = []
    return
  }
  modelLoading.value = true
  try {
    const res = await listErshouModels({
      page: 1, page_size: 200,
      brand_id: modelFilters.brand_id,
      keyword: modelFilters.keyword || undefined
    })
    const data = res.data || {}
    models.value = data.list || data || []
  } catch (e) {
    models.value = []
  } finally {
    modelLoading.value = false
  }
}

const openModels = (row) => {
  currentBrand.value = row
  modelFilters.brand_id = row.id
  activeTab.value = 'models'
  loadModels()
}

const modelFormVisible = ref(false)
const modelFormLoading = ref(false)
const modelFormRef = ref(null)
const modelIsEdit = ref(false)
const modelFormTitle = computed(() => modelIsEdit.value ? '编辑型号' : '新型号')
const modelForm = reactive({
  id: null, name: '', full_name: '', image: '', release_date: '',
  reference_price: 0, specificationsStr: '', sort: 0, status: 1
})
const modelRules = { name: [{ required: true, message: '请输入型号名', trigger: 'blur' }] }

const openModelCreate = () => {
  modelIsEdit.value = false
  Object.assign(modelForm, {
    id: null, name: '', full_name: '', image: '', release_date: '',
    reference_price: 0, specificationsStr: '', sort: 0, status: 1
  })
  modelFormVisible.value = true
}

const openModelEdit = (row) => {
  modelIsEdit.value = true
  Object.assign(modelForm, {
    ...row,
    specificationsStr: row.specifications ? JSON.stringify(row.specifications, null, 2) : ''
  })
  modelFormVisible.value = true
}

const onSubmitModel = async () => {
  try {
    await modelFormRef.value.validate()
    let specifications = null
    if (modelForm.specificationsStr) {
      try {
        specifications = JSON.parse(modelForm.specificationsStr)
      } catch (e) {
        ElMessage.error('规格参数 JSON 格式错误')
        return
      }
    }
    const payload = {
      name: modelForm.name,
      full_name: modelForm.full_name,
      image: modelForm.image,
      release_date: modelForm.release_date,
      reference_price: modelForm.reference_price,
      specifications,
      sort: modelForm.sort,
      status: modelForm.status
    }
    modelFormLoading.value = true
    if (modelIsEdit.value) {
      await updateErshouModel(modelForm.id, payload)
    } else {
      await createErshouModel(modelFilters.brand_id, payload)
    }
    ElMessage.success(modelIsEdit.value ? '更新成功' : '创建成功')
    modelFormVisible.value = false
    await loadModels()
  } catch (e) {
    // ignore
  } finally {
    modelFormLoading.value = false
  }
}

const onDeleteModel = async (row) => {
  try {
    await ElMessageBox.confirm(`确认删除型号 "${row.name}"？`, '提示', { type: 'warning' })
    await deleteErshouModel(row.id)
    ElMessage.success('删除成功')
    await loadModels()
  } catch (e) { /* cancel */ }
}

onMounted(() => loadBrands())
</script>

<style scoped>
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.logo-thumb { width: 40px; height: 40px; border-radius: 4px; border: 1px solid #ebeef5; }
.logo-empty { display: flex; align-items: center; justify-content: center; background: #fafafa; color: #909399; font-size: 12px; }
.text-muted { color: #909399; }
</style>
