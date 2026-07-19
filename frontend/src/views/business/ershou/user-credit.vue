<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="用户ID">
          <el-input v-model="filters.user_id" placeholder="用户ID" clearable style="width: 140px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="信用等级">
          <el-select v-model="filters.credit_level" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option v-for="(label, val) in levelMap" :key="val" :label="label" :value="Number(val)" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.is_frozen" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="正常" :value="0" />
            <el-option label="已冻结" :value="1" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar"><el-button :icon="Refresh" @click="loadList">刷新</el-button></div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="user_id" label="用户ID" width="100" fixed="left" />
        <el-table-column label="信用分" width="120">
          <template #default="{ row }">
            <span :class="['credit-score', creditClass(row.credit_score)]">{{ row.credit_score }}</span>
          </template>
        </el-table-column>
        <el-table-column label="等级" width="100">
          <template #default="{ row }">
            <el-tag :type="levelTagType(row.credit_level)" size="small">{{ row.credit_level_text || levelMap[row.credit_level] }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="total_transactions" label="历史交易" width="100" />
        <el-table-column prop="success_transactions" label="成功交易" width="100" />
        <el-table-column prop="cancel_transactions" label="取消交易" width="100" />
        <el-table-column prop="good_reviews" label="好评数" width="80" />
        <el-table-column prop="bad_reviews" label="差评数" width="80" />
        <el-table-column label="好评率" width="100">
          <template #default="{ row }">{{ (row.good_rate * 100).toFixed(1) }}%</template>
        </el-table-column>
        <el-table-column prop="disputes" label="纠纷数" width="80" />
        <el-table-column prop="reports" label="被举报" width="80" />
        <el-table-column prop="penalties" label="处罚数" width="80" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.is_frozen" type="danger" size="small">已冻结</el-tag>
            <el-tag v-else type="success" size="small">正常</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="最后交易" width="160">
          <template #default="{ row }">{{ formatTime(row.last_transaction_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">明细</el-button>
            <el-button type="warning" link size="small" @click="openAdjust(row)">调整</el-button>
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

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="用户信用详情" width="700px">
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="用户ID">{{ detail.user_id }}</el-descriptions-item>
        <el-descriptions-item label="信用分">
          <span :class="['credit-score', creditClass(detail.credit_score)]">{{ detail.credit_score }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="等级">
          <el-tag :type="levelTagType(detail.credit_level)" size="small">{{ detail.credit_level_text || levelMap[detail.credit_level] }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag v-if="detail.is_frozen" type="danger" size="small">已冻结</el-tag>
          <el-tag v-else type="success" size="small">正常</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="总交易数">{{ detail.total_transactions }}</el-descriptions-item>
        <el-descriptions-item label="成功交易">{{ detail.success_transactions }}</el-descriptions-item>
        <el-descriptions-item label="取消交易">{{ detail.cancel_transactions }}</el-descriptions-item>
        <el-descriptions-item label="好评数">{{ detail.good_reviews }}</el-descriptions-item>
        <el-descriptions-item label="中评数">{{ detail.medium_reviews }}</el-descriptions-item>
        <el-descriptions-item label="差评数">{{ detail.bad_reviews }}</el-descriptions-item>
        <el-descriptions-item label="好评率">{{ (detail.good_rate * 100).toFixed(1) }}%</el-descriptions-item>
        <el-descriptions-item label="纠纷数">{{ detail.disputes }}</el-descriptions-item>
        <el-descriptions-item label="被举报">{{ detail.reports }}</el-descriptions-item>
        <el-descriptions-item label="处罚数">{{ detail.penalties }}</el-descriptions-item>
        <el-descriptions-item label="最后交易">{{ formatTime(detail.last_transaction_at) }}</el-descriptions-item>
        <el-descriptions-item label="冻结原因">{{ detail.frozen_reason || '-' }}</el-descriptions-item>
        <el-descriptions-item label="冻结至">{{ formatTime(detail.frozen_until) }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
      </el-descriptions>
      <template #footer><el-button @click="detailVisible = false">关闭</el-button></template>
    </el-dialog>

    <!-- 调整弹窗 -->
    <el-dialog v-model="adjustVisible" title="调整信用分" width="500px">
      <el-form :model="adjustForm" label-width="100px">
        <el-form-item label="用户ID">
          <el-input :model-value="adjustForm.user_id" disabled />
        </el-form-item>
        <el-form-item label="当前分数">
          <el-input :model-value="adjustForm.current_score" disabled />
        </el-form-item>
        <el-form-item label="调整后分数">
          <el-input-number v-model="adjustForm.credit_score" :min="0" :max="1000" :controls="false" style="width: 100%" />
        </el-form-item>
        <el-form-item label="调整等级">
          <el-select v-model="adjustForm.credit_level" style="width: 100%">
            <el-option v-for="(label, val) in levelMap" :key="val" :label="label" :value="Number(val)" />
          </el-select>
        </el-form-item>
        <el-form-item label="冻结">
          <el-switch v-model="adjustForm.is_frozen" />
        </el-form-item>
        <el-form-item v-if="adjustForm.is_frozen" label="冻结原因">
          <el-input v-model="adjustForm.frozen_reason" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item v-if="adjustForm.is_frozen" label="冻结至">
          <el-date-picker v-model="adjustForm.frozen_until" type="datetime" value-format="YYYY-MM-DDTHH:mm:ss" style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="adjustVisible = false">取消</el-button>
        <el-button type="primary" :loading="adjustLoading" @click="onAdjust">确认调整</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, RefreshLeft, Search } from '@element-plus/icons-vue'
import { getErshouUserCredit, updateErshouUserCredit } from '@/api/ershou'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const filters = reactive({ user_id: '', credit_level: null, is_frozen: null })

const levelMap = { 0: '未评级', 1: '一星', 2: '二星', 3: '三星', 4: '四星', 5: '五星' }
const levelTagType = (l) => ({ 0: 'info', 1: 'info', 2: 'info', 3: 'warning', 4: 'primary', 5: 'success' }[l] || 'info')

const creditClass = (score) => {
  if (score >= 800) return 'credit-excellent'
  if (score >= 600) return 'credit-good'
  if (score >= 400) return 'credit-medium'
  if (score >= 200) return 'credit-poor'
  return 'credit-bad'
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => { filters.user_id = ''; filters.credit_level = null; filters.is_frozen = null; page.value = 1; loadList() }

const loadList = async () => {
  // 用户信用列表接口若不提供列表端点，则只能按用户ID单查
  // 此处先支持按 user_id 查询，未传时展示空
  if (!filters.user_id) {
    list.value = []
    total.value = 0
    return
  }
  loading.value = true
  try {
    const res = await getErshouUserCredit(filters.user_id)
    const d = res.data
    if (d) {
      list.value = [d]
      total.value = 1
    } else {
      list.value = []
      total.value = 0
    }
  } catch (e) {
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const detailVisible = ref(false)
const detail = ref(null)
const openDetail = (row) => { detail.value = row; detailVisible.value = true }

const adjustVisible = ref(false)
const adjustLoading = ref(false)
const adjustForm = reactive({
  user_id: '', current_score: 0, credit_score: 0, credit_level: 0,
  is_frozen: false, frozen_reason: '', frozen_until: null
})

const openAdjust = (row) => {
  adjustForm.user_id = row.user_id
  adjustForm.current_score = row.credit_score
  adjustForm.credit_score = row.credit_score
  adjustForm.credit_level = row.credit_level
  adjustForm.is_frozen = row.is_frozen
  adjustForm.frozen_reason = row.frozen_reason || ''
  adjustForm.frozen_until = row.frozen_until || null
  adjustVisible.value = true
}

const onAdjust = async () => {
  adjustLoading.value = true
  try {
    await updateErshouUserCredit(adjustForm.user_id, {
      credit_score: adjustForm.credit_score,
      credit_level: adjustForm.credit_level,
      frozen_reason: adjustForm.frozen_reason,
      frozen_until: adjustForm.frozen_until
    })
    ElMessage.success('调整成功')
    adjustVisible.value = false
    await loadList()
  } catch (e) {
    // 失败已提示
  } finally {
    adjustLoading.value = false
  }
}

onMounted(() => loadList())
</script>

<style scoped>
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: flex-start; margin-bottom: 12px; }
.credit-score { font-weight: 600; font-size: 16px; }
.credit-excellent { color: #67c23a; }
.credit-good { color: #409eff; }
.credit-medium { color: #e6a23c; }
.credit-poor { color: #f56c6c; }
.credit-bad { color: #909399; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
