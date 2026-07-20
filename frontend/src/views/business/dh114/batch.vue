<template>
  <div class="app-container">
    <div class="page-card">
      <el-alert
        title="批量操作说明"
        type="info"
        description="支持批量审核、批量状态变更、批量删除、批量认证审核。先在商户列表选择多个商户，然后在此页执行批量操作；或直接输入商户 ID 列表进行批量处理。"
        :closable="false"
        show-icon
        style="margin-bottom: 16px"
      />

      <el-tabs v-model="activeTab">
        <!-- 批量审核 -->
        <el-tab-pane label="批量审核" name="audit">
          <el-form label-width="100px">
            <el-form-item label="商户ID列表">
              <el-input v-model="auditForm.idsStr" type="textarea" :rows="4" placeholder="多个ID用逗号或换行分隔，如 1,2,3" />
            </el-form-item>
            <el-form-item label="审核结果">
              <el-radio-group v-model="auditForm.audit_status">
                <el-radio :value="1">通过</el-radio>
                <el-radio :value="2">拒绝</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item v-if="auditForm.audit_status === 2" label="拒绝原因">
              <el-input v-model="auditForm.audit_reason" type="textarea" :rows="2" placeholder="拒绝原因（可选）" />
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
            <el-form-item label="商户ID列表">
              <el-input v-model="statusForm.idsStr" type="textarea" :rows="4" placeholder="多个ID用逗号或换行分隔" />
            </el-form-item>
            <el-form-item label="目标状态">
              <el-radio-group v-model="statusForm.status">
                <el-radio :value="1">发布</el-radio>
                <el-radio :value="3">下架</el-radio>
                <el-radio :value="4">关闭</el-radio>
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
            <el-form-item label="商户ID列表">
              <el-input v-model="deleteForm.idsStr" type="textarea" :rows="4" placeholder="多个ID用逗号或换行分隔" />
            </el-form-item>
            <el-form-item>
              <el-button type="danger" :loading="deleteLoading" @click="onBatchDelete">执行批量删除</el-button>
              <el-button @click="clearDelete">清空</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 批量认证审核 -->
        <el-tab-pane label="批量认证审核" name="verify">
          <el-form label-width="100px">
            <el-form-item label="认证ID列表">
              <el-input v-model="verifyForm.idsStr" type="textarea" :rows="4" placeholder="多个认证ID用逗号或换行分隔" />
            </el-form-item>
            <el-form-item label="审核结果">
              <el-radio-group v-model="verifyForm.status">
                <el-radio :value="2">通过</el-radio>
                <el-radio :value="3">拒绝</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item v-if="verifyForm.status === 3" label="拒绝原因">
              <el-input v-model="verifyForm.audit_remark" type="textarea" :rows="2" placeholder="拒绝原因（可选）" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="verifyLoading" @click="onBatchVerify">执行批量认证审核</el-button>
              <el-button @click="clearVerify">清空</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 批量推荐 -->
        <el-tab-pane label="批量推荐" name="recommend">
          <el-form label-width="100px">
            <el-form-item label="商户ID列表">
              <el-input v-model="recommendForm.idsStr" type="textarea" :rows="4" placeholder="多个商户ID用逗号或换行分隔" />
            </el-form-item>
            <el-form-item label="推荐类型">
              <el-select v-model="recommendForm.rec_type" style="width: 200px">
                <el-option label="首页推荐" value="home" />
                <el-option label="频道推荐" value="channel" />
                <el-option label="搜索推荐" value="search" />
                <el-option label="附近推荐" value="nearby" />
                <el-option label="分类推荐" value="category" />
              </el-select>
            </el-form-item>
            <el-form-item label="有效期">
              <el-date-picker
                v-model="recommendForm.dateRange"
                type="datetimerange"
                range-separator="至"
                start-placeholder="开始时间"
                end-placeholder="结束时间"
                value-format="YYYY-MM-DD HH:mm:ss"
                style="width: 100%"
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="recommendLoading" @click="onBatchRecommend">执行批量推荐</el-button>
              <el-button @click="clearRecommend">清空</el-button>
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
        </el-descriptions>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/utils/request'

const activeTab = ref('audit')
const result = ref(null)

// ===== 解析 ID 列表 =====
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
    ElMessage.warning('请输入有效的商户ID')
    return
  }
  try {
    await ElMessageBox.confirm(`确认对 ${ids.length} 个商户执行批量${auditForm.audit_status === 1 ? '通过' : '拒绝'}？`, '提示', { type: 'warning' })
    auditLoading.value = true
    const res = await request.post('/dh114/admin/dh114s/batch-audit', {
      ids,
      audit_status: auditForm.audit_status,
      audit_reason: auditForm.audit_reason
    })
    result.value = res.data || { total: ids.length, success: ids.length, failed: 0 }
    ElMessage.success('批量审核完成')
  } catch (e) {
    // 取消或失败
  } finally {
    auditLoading.value = false
  }
}

// ===== 批量状态变更 =====
const statusForm = reactive({ idsStr: '', status: 3 })
const statusLoading = ref(false)

const clearStatus = () => {
  statusForm.idsStr = ''
  statusForm.status = 3
  result.value = null
}

const onBatchStatus = async () => {
  const ids = parseIds(statusForm.idsStr)
  if (!ids.length) {
    ElMessage.warning('请输入有效的商户ID')
    return
  }
  const label = { 1: '发布', 3: '下架', 4: '关闭' }[statusForm.status]
  try {
    await ElMessageBox.confirm(`确认对 ${ids.length} 个商户批量设为「${label}」？`, '提示', { type: 'warning' })
    statusLoading.value = true
    let success = 0
    let failed = 0
    const failedIds = []
    for (const id of ids) {
      try {
        await request.put(`/dh114/admin/dh114s/${id}/status`, { status: statusForm.status })
        success++
      } catch (e) {
        failed++
        failedIds.push(id)
      }
    }
    result.value = { total: ids.length, success, failed, failed_ids: failedIds }
    ElMessage.success('批量状态变更完成')
  } catch (e) {
    // 取消
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
    ElMessage.warning('请输入有效的商户ID')
    return
  }
  try {
    await ElMessageBox.confirm(`确认批量删除 ${ids.length} 个商户？删除后不可恢复！`, '危险操作', {
      type: 'error', confirmButtonText: '确认删除', cancelButtonText: '取消'
    })
    deleteLoading.value = true
    let success = 0
    let failed = 0
    const failedIds = []
    for (const id of ids) {
      try {
        await request.delete(`/dh114/${id}`)
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

// ===== 批量认证审核 =====
const verifyForm = reactive({ idsStr: '', status: 2, audit_remark: '' })
const verifyLoading = ref(false)

const clearVerify = () => {
  verifyForm.idsStr = ''
  verifyForm.status = 2
  verifyForm.audit_remark = ''
  result.value = null
}

const onBatchVerify = async () => {
  const ids = parseIds(verifyForm.idsStr)
  if (!ids.length) {
    ElMessage.warning('请输入有效的认证ID')
    return
  }
  const label = verifyForm.status === 2 ? '通过' : '拒绝'
  try {
    await ElMessageBox.confirm(`确认对 ${ids.length} 个认证申请批量${label}？`, '提示', { type: 'warning' })
    verifyLoading.value = true
    let success = 0
    let failed = 0
    const failedIds = []
    for (const id of ids) {
      try {
        await request.put(`/dh114/admin/verifications/${id}/audit`, {
          status: verifyForm.status,
          audit_remark: verifyForm.audit_remark
        })
        success++
      } catch (e) {
        failed++
        failedIds.push(id)
      }
    }
    result.value = { total: ids.length, success, failed, failed_ids: failedIds }
    ElMessage.success('批量认证审核完成')
  } catch (e) {
    // 取消
  } finally {
    verifyLoading.value = false
  }
}

// ===== 批量推荐 =====
const recommendForm = reactive({ idsStr: '', rec_type: 'home', dateRange: [] })
const recommendLoading = ref(false)

const clearRecommend = () => {
  recommendForm.idsStr = ''
  recommendForm.rec_type = 'home'
  recommendForm.dateRange = []
  result.value = null
}

const onBatchRecommend = async () => {
  const ids = parseIds(recommendForm.idsStr)
  if (!ids.length) {
    ElMessage.warning('请输入有效的商户ID')
    return
  }
  try {
    await ElMessageBox.confirm(`确认为 ${ids.length} 个商户批量创建推荐？`, '提示', { type: 'warning' })
    recommendLoading.value = true
    let success = 0
    let failed = 0
    const failedIds = []
    for (const id of ids) {
      try {
        await request.post('/dh114/admin/recommendations', {
          dh114_id: id,
          rec_type: recommendForm.rec_type,
          start_time: recommendForm.dateRange?.[0] || undefined,
          end_time: recommendForm.dateRange?.[1] || undefined
        })
        success++
      } catch (e) {
        failed++
        failedIds.push(id)
      }
    }
    result.value = { total: ids.length, success, failed, failed_ids: failedIds }
    ElMessage.success('批量推荐完成')
  } catch (e) {
    // 取消
  } finally {
    recommendLoading.value = false
  }
}
</script>

<style scoped>
.result-card { margin-top: 24px; }
.text-success { color: #67c23a; font-weight: 600; }
.text-danger { color: #f56c6c; font-weight: 600; }
.failed-id-tag { margin-right: 4px; margin-bottom: 4px; }
</style>
