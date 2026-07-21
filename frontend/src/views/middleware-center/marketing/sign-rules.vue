<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar">
        <el-button :icon="Refresh" @click="loadList">刷新</el-button>
        <el-button type="primary" :icon="Plus" @click="openCreate">新建签到规则</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column label="连续天数" width="120">
          <template #default="{ row }">
            <el-tag type="warning" size="small">第 {{ row.day }} 天</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="奖励积分" width="120">
          <template #default="{ row }">
            <span class="points">{{ row.points }}</span>
          </template>
        </el-table-column>
        <el-table-column label="额外奖励" min-width="240">
          <template #default="{ row }">
            <span v-if="!row.extra_reward" class="text-muted">无</span>
            <pre v-else class="extra-reward">{{ formatExtra(row.extra_reward) }}</pre>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-switch :model-value="row.status === 1" @change="(val) => onToggle(row, val)" />
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="170">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
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
      <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
        <el-form-item label="连续天数" prop="day">
          <el-input-number v-model="form.day" :min="1" :max="365" :controls="false" style="width: 100%" />
          <div class="form-tip">第 N 天连续签到触发该规则</div>
        </el-form-item>
        <el-form-item label="奖励积分" prop="points">
          <el-input-number v-model="form.points" :min="0" :controls="false" style="width: 100%" />
          <div class="form-tip">该天签到奖励的积分数量</div>
        </el-form-item>
        <el-form-item label="额外奖励">
          <el-input
            v-model="form.extra_reward_json"
            type="textarea"
            :rows="5"
            placeholder='JSON 格式，例如 {"coupon_id": 12, "coupon_count": 1}'
          />
          <div class="form-tip">可选，JSON 格式（如优惠券奖励、积分加成等），留空表示无额外奖励</div>
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" active-text="启用" inactive-text="禁用" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onSubmit">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search, Plus } from '@element-plus/icons-vue'
import {
  getMarketingSignRuleList,
  createMarketingSignRule,
  updateMarketingSignRule,
  deleteMarketingSignRule
} from '@/api/marketing'

const loading = ref(false)
const submitting = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const filters = reactive({ status: null })

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.status = null
  page.value = 1
  loadList()
}

const formatTime = (t) => {
  if (!t) return '-'
  return String(t).replace('T', ' ').slice(0, 16)
}

const formatExtra = (extra) => {
  if (!extra) return ''
  if (typeof extra === 'string') return extra
  try {
    return JSON.stringify(extra, null, 2)
  } catch (e) {
    return String(extra)
  }
}

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      status: filters.status === null || filters.status === '' ? undefined : filters.status
    }
    const res = await getMarketingSignRuleList(params)
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
const formVisible = ref(false)
const formRef = ref(null)
const isEdit = ref(false)
const formTitle = computed(() => isEdit.value ? '编辑签到规则' : '新建签到规则')
const form = reactive({
  id: null,
  day: 1,
  points: 0,
  extra_reward_json: '',
  status: 1
})
const rules = {
  day: [{ required: true, message: '请输入连续天数', trigger: 'blur' }],
  points: [{ required: true, message: '请输入奖励积分', trigger: 'blur' }]
}

const resetForm = () => {
  Object.assign(form, {
    id: null,
    day: 1,
    points: 0,
    extra_reward_json: '',
    status: 1
  })
}

const openCreate = () => {
  isEdit.value = false
  resetForm()
  formVisible.value = true
}

const openEdit = (row) => {
  isEdit.value = true
  Object.assign(form, {
    id: row.id,
    day: row.day || 1,
    points: row.points || 0,
    extra_reward_json: row.extra_reward ? formatExtra(row.extra_reward) : '',
    status: row.status ?? 1
  })
  formVisible.value = true
}

const onSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    // 校验额外奖励 JSON
    let extraReward = null
    if (form.extra_reward_json && form.extra_reward_json.trim()) {
      try {
        extraReward = JSON.parse(form.extra_reward_json)
      } catch (e) {
        ElMessage.error('额外奖励 JSON 格式错误')
        return
      }
    }
    submitting.value = true
    try {
      const payload = {
        day: form.day,
        points: form.points,
        extra_reward: extraReward,
        status: form.status
      }
      if (isEdit.value) {
        await updateMarketingSignRule(form.id, payload)
        ElMessage.success('更新成功')
      } else {
        await createMarketingSignRule(payload)
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

const onToggle = async (row, val) => {
  try {
    await updateMarketingSignRule(row.id, { status: val ? 1 : 0 })
    ElMessage.success(val ? '已启用' : '已禁用')
    row.status = val ? 1 : 0
  } catch (e) {}
}

const onDelete = (row) => {
  ElMessageBox.confirm(`确认删除第 ${row.day} 天签到规则？`, '提示', {
    type: 'warning',
    confirmButtonText: '确认',
    cancelButtonText: '取消'
  }).then(async () => {
    try {
      await deleteMarketingSignRule(row.id)
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
.points { color: #f56c6c; font-weight: 600; }
.text-muted { color: #999; }
.extra-reward { margin: 0; font-size: 12px; color: #555; white-space: pre-wrap; word-break: break-all; background: #f5f7fa; padding: 4px 8px; border-radius: 4px; }
.form-tip { font-size: 12px; color: #909399; line-height: 1.4; margin-top: 4px; }
</style>
