<template>
  <div class="app-container">
    <div class="page-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button :icon="Refresh" @click="loadList">刷新</el-button>
        </div>
        <div class="toolbar-right">
          <el-input-number
            v-model="shopFilter"
            :min="0"
            placeholder="店铺ID"
            style="width: 130px"
            @change="onSearch"
          />
          <el-select
            v-model="typeFilter"
            placeholder="认证类型"
            clearable
            style="width: 140px; margin-left: 8px"
            @change="onSearch"
          >
            <el-option label="企业认证" value="business" />
            <el-option label="个人认证" value="personal" />
          </el-select>
          <el-select
            v-model="statusFilter"
            placeholder="状态"
            clearable
            style="width: 120px; margin-left: 8px"
            @change="onSearch"
          >
            <el-option label="待审" :value="0" />
            <el-option label="通过" :value="1" />
            <el-option label="拒绝" :value="2" />
          </el-select>
          <el-button type="primary" :icon="Search" style="margin-left: 8px" @click="onSearch">搜索</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="region_id" label="地区ID" width="90" />
        <el-table-column prop="shop_id" label="店铺ID" width="100" />
        <el-table-column label="认证类型" width="110">
          <template #default="{ row }">
            <el-tag :type="row.type === 'business' ? 'warning' : 'info'" size="small">
              {{ row.type_text || (row.type === 'business' ? '企业认证' : '个人认证') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="license_no" label="营业执照号" min-width="160" show-overflow-tooltip />
        <el-table-column prop="legal_person" label="法人代表" min-width="120" />
        <el-table-column label="营业执照" width="120">
          <template #default="{ row }">
            <el-image
              v-if="row.license_image"
              :src="row.license_image"
              :preview-src-list="[row.license_image]"
              fit="cover"
              class="license-thumb"
              :preview-teleported="true"
            />
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">
              {{ row.status_text || statusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="audited_at" label="审核时间" width="170">
          <template #default="{ row }">{{ row.audited_at || '-' }}</template>
        </el-table-column>
        <el-table-column prop="created_at" label="提交时间" width="170" />
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="openDetail(row)">详情</el-button>
            <el-button
              v-if="row.status === 0"
              link
              type="warning"
              size="small"
              @click="openAudit(row)"
            >审核</el-button>
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
          @size-change="loadList"
          @current-change="loadList"
        />
      </div>
    </div>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="认证详情" width="640px">
      <el-descriptions :column="2" border v-if="detail">
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="地区ID">{{ detail.region_id }}</el-descriptions-item>
        <el-descriptions-item label="店铺ID">{{ detail.shop_id }}</el-descriptions-item>
        <el-descriptions-item label="认证类型">
          {{ detail.type_text || (detail.type === 'business' ? '企业认证' : '个人认证') }}
        </el-descriptions-item>
        <el-descriptions-item label="营业执照号">{{ detail.license_no || '-' }}</el-descriptions-item>
        <el-descriptions-item label="法人代表">{{ detail.legal_person || '-' }}</el-descriptions-item>
        <el-descriptions-item label="法人身份证号" :span="2">{{ detail.legal_person_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTagType(detail.status)" size="small">
            {{ detail.status_text || statusText(detail.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="审核时间">{{ detail.audited_at || '-' }}</el-descriptions-item>
        <el-descriptions-item label="审核备注" :span="2">{{ detail.audit_reason || '-' }}</el-descriptions-item>
        <el-descriptions-item label="提交时间" :span="2">{{ detail.created_at }}</el-descriptions-item>
        <el-descriptions-item label="营业执照图片" :span="2">
          <el-image
            v-if="detail.license_image"
            :src="detail.license_image"
            :preview-src-list="[detail.license_image]"
            fit="cover"
            class="license-image"
            :preview-teleported="true"
          />
          <span v-else class="text-muted">未上传</span>
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>

    <!-- 审核弹窗 -->
    <el-dialog v-model="auditVisible" title="审核认证" width="480px">
      <el-form :model="auditForm" label-width="100px">
        <el-form-item label="认证ID">
          <span>#{{ auditForm.id }}</span>
        </el-form-item>
        <el-form-item label="店铺ID">
          <span>{{ auditForm.shop_id }}</span>
        </el-form-item>
        <el-form-item label="认证类型">
          <el-tag :type="auditForm.type === 'business' ? 'warning' : 'info'" size="small">
            {{ auditForm.type === 'business' ? '企业认证' : '个人认证' }}
          </el-tag>
        </el-form-item>
        <el-form-item label="营业执照号">
          <span>{{ auditForm.license_no || '-' }}</span>
        </el-form-item>
        <el-form-item label="法人代表">
          <span>{{ auditForm.legal_person || '-' }}</span>
        </el-form-item>
        <el-form-item label="审核结果">
          <el-radio-group v-model="auditForm.status">
            <el-radio :value="1">通过</el-radio>
            <el-radio :value="2">拒绝</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="审核备注">
          <el-input v-model="auditForm.audit_reason" type="textarea" :rows="3" maxlength="500" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="auditVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onSubmitAudit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Search } from '@element-plus/icons-vue'
import {
  getMerchantVerificationList,
  auditMerchantVerification
} from '@/api/merchant'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const shopFilter = ref(0)
const typeFilter = ref('')
const statusFilter = ref(null)

const detailVisible = ref(false)
const detail = ref(null)

const auditVisible = ref(false)
const submitting = ref(false)
const auditForm = reactive({
  id: 0,
  shop_id: 0,
  type: 'business',
  license_no: '',
  legal_person: '',
  status: 1,
  audit_reason: ''
})

function statusText(s) {
  return s === 1 ? '通过' : s === 2 ? '拒绝' : '待审'
}

function statusTagType(s) {
  return s === 1 ? 'success' : s === 2 ? 'danger' : 'warning'
}

async function loadList() {
  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value }
    if (shopFilter.value > 0) params.shop_id = shopFilter.value
    if (typeFilter.value) params.type = typeFilter.value
    if (statusFilter.value !== null && statusFilter.value !== '') params.status = statusFilter.value
    const res = await getMerchantVerificationList(params)
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (e) {
    ElMessage.error('加载认证列表失败')
  } finally {
    loading.value = false
  }
}

function onSearch() {
  page.value = 1
  loadList()
}

function openDetail(row) {
  detail.value = row
  detailVisible.value = true
}

function openAudit(row) {
  auditForm.id = row.id
  auditForm.shop_id = row.shop_id
  auditForm.type = row.type
  auditForm.license_no = row.license_no
  auditForm.legal_person = row.legal_person
  auditForm.status = 1
  auditForm.audit_reason = ''
  auditVisible.value = true
}

async function onSubmitAudit() {
  submitting.value = true
  try {
    await auditMerchantVerification(auditForm.id, {
      status: auditForm.status,
      audit_reason: auditForm.audit_reason
    })
    ElMessage.success('审核完成')
    auditVisible.value = false
    loadList()
  } catch (e) {
    ElMessage.error(e.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

onMounted(() => loadList())
</script>

<style scoped>
.license-thumb { width: 60px; height: 40px; border-radius: 4px; }
.license-image { width: 200px; height: 150px; border-radius: 4px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.text-muted { color: #c0c4cc; }
</style>
