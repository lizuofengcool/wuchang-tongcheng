<template>
  <div class="app-container">
    <div class="page-card">
      <el-alert
        title="批量操作说明"
        type="info"
        description="支持批量审核、批量状态变更、批量删除、批量导出。先在岗位列表选择多个，然后在此页执行批量操作；或直接输入岗位 ID 列表进行批量处理。"
        :closable="false"
        show-icon
        style="margin-bottom: 16px"
      />

      <el-tabs v-model="activeTab">
        <!-- 批量审核 -->
        <el-tab-pane label="批量审核" name="audit">
          <el-form label-width="100px">
            <el-form-item label="岗位ID列表">
              <el-input v-model="auditForm.idsStr" type="textarea" :rows="4" placeholder="多个ID用逗号或换行分隔，如 1,2,3" />
            </el-form-item>
            <el-form-item label="审核结果">
              <el-radio-group v-model="auditForm.audit_status">
                <el-radio :value="1">通过</el-radio>
                <el-radio :value="2">拒绝</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item v-if="auditForm.audit_status === 2" label="拒绝原因">
              <el-input v-model="auditForm.audit_reason" type="textarea" :rows="2" placeholder="拒绝原因（可选）" maxlength="500" show-word-limit />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="auditLoading" @click="onBatchAudit">执行批量审核</el-button>
              <el-button @click="clearAudit">清空</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 批量状态变更 -->
        <el-tab-pane label="批量状态变更" name="status">
          <el-form label-width="100px">
            <el-form-item label="岗位ID列表">
              <el-input v-model="statusForm.idsStr" type="textarea" :rows="4" placeholder="多个ID用逗号或换行分隔" />
            </el-form-item>
            <el-form-item label="目标状态">
              <el-radio-group v-model="statusForm.status">
                <el-radio :value="1">发布</el-radio>
                <el-radio :value="2">下架</el-radio>
                <el-radio :value="3">设为过期</el-radio>
                <el-radio :value="5">设为满员</el-radio>
                <el-radio :value="6">关闭</el-radio>
                <el-radio :value="7">完成</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="statusLoading" @click="onBatchStatus">执行批量状态变更</el-button>
              <el-button @click="clearStatus">清空</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 批量删除 -->
        <el-tab-pane label="批量删除" name="delete">
          <el-form label-width="100px">
            <el-form-item label="岗位ID列表">
              <el-input v-model="deleteForm.idsStr" type="textarea" :rows="4" placeholder="多个ID用逗号或换行分隔" />
            </el-form-item>
            <el-form-item>
              <el-button type="danger" :loading="deleteLoading" @click="onBatchDelete">执行批量删除</el-button>
              <el-button @click="clearDelete">清空</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 批量导出 -->
        <el-tab-pane label="批量导出" name="export">
          <el-form label-width="100px">
            <el-form-item label="导出格式">
              <el-radio-group v-model="exportForm.format">
                <el-radio value="excel">Excel</el-radio>
                <el-radio value="csv">CSV</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item label="岗位类型">
              <el-select v-model="exportForm.linggong_type" placeholder="全部" clearable style="width: 200px">
                <el-option v-for="(label, val) in typeMap" :key="val" :label="label" :value="val" />
              </el-select>
            </el-form-item>
            <el-form-item label="发布者">
              <el-select v-model="exportForm.publisher_type" placeholder="全部" clearable style="width: 200px">
                <el-option v-for="(label, val) in publisherMap" :key="val" :label="label" :value="val" />
              </el-select>
            </el-form-item>
            <el-form-item label="关键词">
              <el-input v-model="exportForm.keyword" placeholder="关键词（可选）" style="width: 240px" />
            </el-form-item>
            <el-form-item label="状态">
              <el-select v-model="exportForm.status" placeholder="全部" clearable style="width: 200px">
                <el-option v-for="(label, val) in statusMap" :key="val" :label="label" :value="Number(val)" />
              </el-select>
            </el-form-item>
            <el-form-item label="审核状态">
              <el-select v-model="exportForm.audit_status" placeholder="全部" clearable style="width: 200px">
                <el-option label="待审核" :value="0" />
                <el-option label="已通过" :value="1" />
                <el-option label="已拒绝" :value="2" />
              </el-select>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="exportLoading" @click="onExport">导出文件</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 批量审核规则删除 -->
        <el-tab-pane label="批量规则删除" name="rules">
          <el-form label-width="100px">
            <el-form-item label="规则ID列表">
              <el-input v-model="rulesForm.idsStr" type="textarea" :rows="4" placeholder="多个审核规则ID用逗号或换行分隔" />
            </el-form-item>
            <el-form-item>
              <el-button type="danger" :loading="rulesLoading" @click="onBatchDeleteRules">执行批量删除规则</el-button>
              <el-button @click="clearRules">清空</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>

      <!-- 操作结果 -->
      <el-card v-if="result" class="result-card">
        <template #header><span>操作结果</span></template>
        <el-descriptions :column="3" border>
          <el-descriptions-item label="总数">{{ result.total }}</el-descriptions-item>
          <el-descriptions-item label="成功">
            <span class="text-success">{{ result.success }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="失败">
            <span class="text-danger">{{ result.failed }}</span>
          </el-descriptions-item>
          <el-descriptions-item v-if="result.failed_ids && result.failed_ids.length" label="失败ID列表" :span="3">
            <el-tag v-for="id in result.failed_ids" :key="id" type="danger" size="small" class="failed-id-tag">{{ id }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item v-if="result.message" label="说明" :span="3">
            {{ result.message }}
          </el-descriptions-item>
        </el-descriptions>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  auditLinggong, batchUpdateLinggongStatus, batchDeleteLinggongAuditRules
} from '@/api/linggong'
import request from '@/utils/request'

const typeMap = {
  short_term: '短期兼职', long_term: '长期兼职', task: '任务制',
  hourly: '小时工', daily: '日结工', temp: '临时工'
}
const publisherMap = {
  personal: '个人', company: '企业', agent: '中介', headhunter: '猎头'
}
const statusMap = {
  0: '草稿', 1: '已发布', 2: '已下架', 3: '已过期',
  4: '已删除', 5: '已满员', 6: '已关闭', 7: '已完成'
}

const activeTab = ref('audit')
const result = ref(null)

const parseIds = (str) => {
  return String(str || '')
    .split(/[,，\n\s]+/)
    .map((s) => s.trim())
    .filter(Boolean)
    .map((s) => Number(s))
    .filter((n) => !isNaN(n) && n > 0)
}

// ===== 批量审核 =====
const auditForm = reactive({ idsStr: '', audit_status: 1, audit_reason: '' })
const auditLoading = ref(false)

const clearAudit = () => {
  auditForm.idsStr = ''
  auditForm.audit_status = 1
  auditForm.audit_reason = ''
  result.value = null
}

const onBatchAudit = async () => {
  const ids = parseIds(auditForm.idsStr)
  if (!ids.length) {
    ElMessage.warning('请输入有效的岗位ID')
    return
  }
  try {
    await ElMessageBox.confirm(`确认对 ${ids.length} 个岗位执行批量${auditForm.audit_status === 1 ? '通过' : '拒绝'}？`, '提示', { type: 'warning' })
    auditLoading.value = true
    let success = 0
    let failed = 0
    const failedIds = []
    for (const id of ids) {
      try {
        await auditLinggong(id, {
          audit_status: auditForm.audit_status,
          audit_reason: auditForm.audit_reason
        })
        success++
      } catch (e) {
        failed++
        failedIds.push(id)
      }
    }
    result.value = { total: ids.length, success, failed, failed_ids: failedIds }
    ElMessage.success('批量审核完成')
  } catch (e) {
    // 取消
  } finally {
    auditLoading.value = false
  }
}

// ===== 批量状态变更 =====
const statusForm = reactive({ idsStr: '', status: 2 })
const statusLoading = ref(false)

const clearStatus = () => {
  statusForm.idsStr = ''
  statusForm.status = 2
  result.value = null
}

const onBatchStatus = async () => {
  const ids = parseIds(statusForm.idsStr)
  if (!ids.length) {
    ElMessage.warning('请输入有效的岗位ID')
    return
  }
  const label = statusMap[statusForm.status]
  try {
    await ElMessageBox.confirm(`确认对 ${ids.length} 个岗位批量设为「${label}」？`, '提示', { type: 'warning' })
    statusLoading.value = true
    const res = await batchUpdateLinggongStatus({ ids, status: statusForm.status })
    const d = res.data || {}
    result.value = {
      total: d.total || ids.length,
      success: d.success || ids.length,
      failed: d.failed || 0,
      failed_ids: d.failed_ids || []
    }
    ElMessage.success('批量状态变更完成')
  } catch (e) {
    // 取消或失败
  } finally {
    statusLoading.value = false
  }
}

// ===== 批量删除 =====
const deleteForm = reactive({ idsStr: '' })
const deleteLoading = ref(false)

const clearDelete = () => {
  deleteForm.idsStr = ''
  result.value = null
}

const onBatchDelete = async () => {
  const ids = parseIds(deleteForm.idsStr)
  if (!ids.length) {
    ElMessage.warning('请输入有效的岗位ID')
    return
  }
  try {
    await ElMessageBox.confirm(`确认批量删除 ${ids.length} 个岗位？删除后不可恢复！`, '危险操作', {
      type: 'error', confirmButtonText: '确认删除', cancelButtonText: '取消'
    })
    deleteLoading.value = true
    let success = 0
    let failed = 0
    const failedIds = []
    for (const id of ids) {
      try {
        await request.delete(`/linggong/${id}`)
        success++
      } catch (e) {
        failed++
        failedIds.push(id)
      }
    }
    result.value = { total: ids.length, success, failed, failed_ids: failedIds }
    ElMessage.success('批量删除完成')
  } catch (e) {
    // 取消
  } finally {
    deleteLoading.value = false
  }
}

// ===== 批量导出 =====
const exportForm = reactive({
  format: 'excel',
  linggong_type: '',
  publisher_type: '',
  keyword: '',
  status: null,
  audit_status: null
})
const exportLoading = ref(false)

const onExport = async () => {
  exportLoading.value = true
  try {
    ElMessage.info('正在导出，请稍候...')
    const blob = await request.get('/linggong/admin/export', {
      params: {
        format: exportForm.format,
        linggong_type: exportForm.linggong_type || undefined,
        publisher_type: exportForm.publisher_type || undefined,
        keyword: exportForm.keyword || undefined,
        status: exportForm.status === null || exportForm.status === '' ? undefined : exportForm.status,
        audit_status: exportForm.audit_status === null || exportForm.audit_status === '' ? undefined : exportForm.audit_status
      },
      responseType: 'blob'
    })
    const url = window.URL.createObjectURL(new Blob([blob]))
    const link = document.createElement('a')
    link.href = url
    link.download = `linggong-${Date.now()}.${exportForm.format === 'excel' ? 'xlsx' : 'csv'}`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
    ElMessage.success('导出成功')
  } catch (e) {
    ElMessage.error('导出失败，请稍后重试')
  } finally {
    exportLoading.value = false
  }
}

// ===== 批量审核规则删除 =====
const rulesForm = reactive({ idsStr: '' })
const rulesLoading = ref(false)

const clearRules = () => {
  rulesForm.idsStr = ''
  result.value = null
}

const onBatchDeleteRules = async () => {
  const ids = parseIds(rulesForm.idsStr)
  if (!ids.length) {
    ElMessage.warning('请输入有效的规则ID')
    return
  }
  try {
    await ElMessageBox.confirm(`确认批量删除 ${ids.length} 条审核规则？`, '危险操作', {
      type: 'error', confirmButtonText: '确认删除', cancelButtonText: '取消'
    })
    rulesLoading.value = true
    const res = await batchDeleteLinggongAuditRules({ ids })
    const d = res.data || {}
    result.value = {
      total: d.total || ids.length,
      success: d.success || ids.length,
      failed: d.failed || 0,
      failed_ids: d.failed_ids || []
    }
    ElMessage.success('批量删除规则完成')
  } catch (e) {
    // 取消或失败
  } finally {
    rulesLoading.value = false
  }
}
</script>

<style scoped>
.result-card { margin-top: 24px; }
.text-success { color: #67c23a; font-weight: 600; }
.text-danger { color: #f56c6c; font-weight: 600; }
.failed-id-tag { margin-right: 4px; margin-bottom: 4px; }
</style>
