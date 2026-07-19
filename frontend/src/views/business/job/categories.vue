<template>
  <div class="app-container">
    <div class="page-card">
      <div class="toolbar">
        <el-button :icon="Refresh" @click="loadList">刷新</el-button>
      </div>
      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column type="index" label="#" width="60" />
        <el-table-column label="分类名称" min-width="160">
          <template #default="{ row }">{{ row.name || row.category || row.label || '-' }}</template>
        </el-table-column>
        <el-table-column label="职位数" width="120">
          <template #default="{ row }">{{ row.count || row.value || row.job_count || 0 }}</template>
        </el-table-column>
        <el-table-column label="占比" width="120">
          <template #default="{ row }">{{ ((row.percentage || row.ratio || 0) * 100).toFixed(1) }}%</template>
        </el-table-column>
        <el-table-column label="说明" min-width="200">
          <template #default="{ row }">{{ row.description || '-' }}</template>
        </el-table-column>
      </el-table>
      <div v-if="!loading && !list.length" class="empty-text">暂无分类数据</div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import { getCategoryStats } from '@/api/job'

const loading = ref(false)
const list = ref([])

const loadList = async () => {
  loading.value = true
  try {
    const res = await getCategoryStats()
    const d = res.data || {}
    list.value = d.list || (Array.isArray(d) ? d : [])
  } catch (e) { list.value = [] } finally { loading.value = false }
}

onMounted(() => loadList())
</script>

<style scoped>
.toolbar { display: flex; justify-content: space-between; margin-bottom: 12px; }
.empty-text { text-align: center; color: #999; padding: 40px; }
</style>
