<template>
  <div class="app-container">
    <div class="page-card" v-loading="loading">
      <div class="editor-header">
        <div class="header-left">
          <el-button :icon="ArrowLeft" @click="goBack">返回</el-button>
          <span class="page-title">{{ pageTitle }}</span>
          <el-tag v-if="pageInfo.id" :type="statusTagType(pageInfo.status)" size="small">
            {{ pageInfo.status_text || statusMap[pageInfo.status] || '未知' }}
          </el-tag>
        </div>
        <div class="header-right">
          <el-button v-if="pageInfo.status === 0 || pageInfo.status === 2" type="success" :icon="Promotion" :loading="actioning" @click="onPublish">
            发布
          </el-button>
          <el-button v-if="pageInfo.status === 1" type="warning" :icon="SwitchButton" :loading="actioning" @click="onOffline">
            下线
          </el-button>
          <el-button :icon="DocumentCopy" @click="openCopy">复制</el-button>
          <el-button type="primary" :icon="Check" :loading="saving" @click="onSave">保存</el-button>
        </div>
      </div>

      <el-alert v-if="!pageInfo.id && !loading" title="未提供页面 ID 或页面不存在" type="warning" :closable="false" show-icon>
        <template #default>
          请从 <router-link :to="{ name: 'DiyPages' }">页面管理</router-link> 列表点击「编辑」进入编辑器。
        </template>
      </el-alert>

      <el-form v-if="pageInfo.id" ref="formRef" :model="form" :rules="rules" label-width="100px" class="editor-form">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="标题" prop="title">
              <el-input v-model="form.title" maxlength="100" />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="类型" prop="type">
              <el-select v-model="form.type" style="width: 100%">
                <el-option v-for="(label, val) in typeMap" :key="val" :label="label" :value="val" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="状态">
              <el-select v-model="form.status" style="width: 100%" :disabled="true">
                <el-option v-for="(label, val) in statusMap" :key="val" :label="label" :value="Number(val)" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="Slug">
              <el-input v-model="form.slug" maxlength="100" placeholder="URL Slug（发布后必填）" />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="业务ID">
              <el-input-number v-model="form.biz_id" :min="0" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="创建者">
              <el-input :model-value="String(pageInfo.user_id || 0)" disabled />
            </el-form-item>
          </el-col>
        </el-row>

        <el-tabs v-model="activeTab" class="editor-tabs">
          <el-tab-pane label="组件配置" name="components">
            <div class="tab-toolbar">
              <span class="tab-tip">组件配置 JSONB（拖拽组件数组，含组件 code/props/排序等）</span>
              <el-button size="small" :icon="MagicStick" @click="formatComponents">格式化</el-button>
            </div>
            <el-input
              v-model="form.componentsText"
              type="textarea"
              :rows="18"
              placeholder='JSON 格式，如 [{"code":"text","props":{"content":"Hello"}}]'
              class="json-textarea"
            />
          </el-tab-pane>
          <el-tab-pane label="页面设置" name="settings">
            <div class="tab-toolbar">
              <span class="tab-tip">页面设置 JSONB（如 SEO/背景/全局样式）</span>
              <el-button size="small" :icon="MagicStick" @click="formatSettings">格式化</el-button>
            </div>
            <el-input
              v-model="form.settingsText"
              type="textarea"
              :rows="18"
              placeholder='JSON 格式，如 {"seo":{"title":"首页"},"background":"#fff"}'
              class="json-textarea"
            />
          </el-tab-pane>
          <el-tab-pane label="元信息" name="meta">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="ID">{{ pageInfo.id }}</el-descriptions-item>
              <el-descriptions-item label="地区ID">{{ pageInfo.region_id }}</el-descriptions-item>
              <el-descriptions-item label="创建时间">{{ formatTime(pageInfo.created_at) }}</el-descriptions-item>
              <el-descriptions-item label="更新时间">{{ formatTime(pageInfo.updated_at) }}</el-descriptions-item>
              <el-descriptions-item label="发布时间">{{ formatTime(pageInfo.published_at) }}</el-descriptions-item>
              <el-descriptions-item label="类型">{{ pageInfo.type_text }}</el-descriptions-item>
            </el-descriptions>
          </el-tab-pane>
        </el-tabs>
      </el-form>
    </div>

    <!-- 复制页面弹窗 -->
    <el-dialog v-model="copyVisible" title="复制页面" width="480px" destroy-on-close>
      <el-form :model="copyForm" label-width="100px">
        <el-form-item label="新页面标题">
          <el-input v-model="copyForm.title" maxlength="100" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="copyVisible = false">取消</el-button>
        <el-button type="primary" :loading="copying" @click="onCopySubmit">确认复制</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ArrowLeft, Check, DocumentCopy, MagicStick, Promotion, SwitchButton
} from '@element-plus/icons-vue'
import {
  getDiyPageDetail,
  updateDiyPage,
  publishDiyPage,
  offlineDiyPage,
  copyDiyPage
} from '@/api/diy'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const saving = ref(false)
const actioning = ref(false)
const copying = ref(false)

const pageInfo = ref({})
const activeTab = ref('components')

const typeMap = {
  home: '首页',
  topic: '专题页',
  shop: '店铺页',
  activity: '活动页'
}
const statusMap = { 0: '草稿', 1: '已发布', 2: '已下线' }
const statusTagType = (s) => ({ 0: 'info', 1: 'success', 2: 'danger' }[s] || 'info')

const formatTime = (t) => {
  if (!t) return '-'
  return String(t).replace('T', ' ').slice(0, 19)
}

const pageTitle = computed(() => {
  if (!pageInfo.value.id) return '页面编辑器'
  return `页面编辑器 - #${pageInfo.value.id} ${pageInfo.value.title || ''}`
})

const form = reactive({
  title: '',
  type: 'home',
  slug: '',
  biz_id: 0,
  status: 0,
  componentsText: '[]',
  settingsText: '{}'
})
const rules = {
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
  type: [{ required: true, message: '请选择类型', trigger: 'change' }]
}

const goBack = () => {
  router.push({ name: 'DiyPages' })
}

const loadPage = async () => {
  const id = route.query.id
  if (!id) {
    return
  }
  loading.value = true
  try {
    const res = await getDiyPageDetail(id)
    const data = res.data || {}
    pageInfo.value = data
    form.title = data.title || ''
    form.type = data.type || 'home'
    form.slug = data.slug || ''
    form.biz_id = data.biz_id || 0
    form.status = data.status
    form.componentsText = data.components ? JSON.stringify(data.components, null, 2) : '[]'
    form.settingsText = data.settings ? JSON.stringify(data.settings, null, 2) : '{}'
  } catch (e) {
    pageInfo.value = {}
  } finally {
    loading.value = false
  }
}

const formatComponents = () => {
  try {
    const obj = JSON.parse(form.componentsText || '[]')
    form.componentsText = JSON.stringify(obj, null, 2)
    ElMessage.success('格式化成功')
  } catch (e) {
    ElMessage.error('JSON 格式错误')
  }
}

const formatSettings = () => {
  try {
    const obj = JSON.parse(form.settingsText || '{}')
    form.settingsText = JSON.stringify(obj, null, 2)
    ElMessage.success('格式化成功')
  } catch (e) {
    ElMessage.error('JSON 格式错误')
  }
}

const onSave = async () => {
  if (!pageInfo.value.id) return
  let componentsObj, settingsObj
  try {
    componentsObj = JSON.parse(form.componentsText || '[]')
  } catch (e) {
    ElMessage.error('组件配置 JSON 格式错误')
    activeTab.value = 'components'
    return
  }
  try {
    settingsObj = JSON.parse(form.settingsText || '{}')
  } catch (e) {
    ElMessage.error('页面设置 JSON 格式错误')
    activeTab.value = 'settings'
    return
  }
  saving.value = true
  try {
    const payload = {
      title: form.title,
      type: form.type,
      slug: form.slug,
      biz_id: form.biz_id,
      components: componentsObj,
      settings: settingsObj
    }
    await updateDiyPage(pageInfo.value.id, payload)
    ElMessage.success('保存成功')
    loadPage()
  } catch (e) {
  } finally {
    saving.value = false
  }
}

const onPublish = async () => {
  if (!pageInfo.value.id) return
  if (!form.slug) {
    ElMessage.warning('发布前请填写 Slug')
    return
  }
  ElMessageBox.confirm('确认发布该页面？发布后可通过 slug 访问。', '提示', {
    type: 'warning',
    confirmButtonText: '确认发布',
    cancelButtonText: '取消'
  }).then(async () => {
    actioning.value = true
    try {
      await publishDiyPage(pageInfo.value.id)
      ElMessage.success('发布成功')
      loadPage()
    } catch (e) {
    } finally {
      actioning.value = false
    }
  }).catch(() => {})
}

const onOffline = async () => {
  if (!pageInfo.value.id) return
  ElMessageBox.confirm('确认下线该页面？下线后 C 端将无法访问。', '提示', {
    type: 'warning',
    confirmButtonText: '确认下线',
    cancelButtonText: '取消'
  }).then(async () => {
    actioning.value = true
    try {
      await offlineDiyPage(pageInfo.value.id)
      ElMessage.success('已下线')
      loadPage()
    } catch (e) {
    } finally {
      actioning.value = false
    }
  }).catch(() => {})
}

// ===== 复制 =====
const copyVisible = ref(false)
const copyForm = reactive({ title: '' })

const openCopy = () => {
  copyForm.title = (pageInfo.value.title || '') + ' - 副本'
  copyVisible.value = true
}

const onCopySubmit = async () => {
  if (!pageInfo.value.id) return
  copying.value = true
  try {
    await copyDiyPage(pageInfo.value.id, { title: copyForm.title })
    ElMessage.success('复制成功')
    copyVisible.value = false
    router.push({ name: 'DiyPages' })
  } catch (e) {
  } finally {
    copying.value = false
  }
}

onMounted(() => {
  loadPage()
})
</script>

<style scoped>
.page-card { background: #fff; padding: 16px; border-radius: 8px; box-shadow: 0 1px 4px rgba(0,0,0,0.04); }
.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 12px;
  border-bottom: 1px solid #ebeef5;
  margin-bottom: 16px;
}
.header-left { display: flex; align-items: center; gap: 12px; }
.header-right { display: flex; gap: 8px; }
.page-title { font-size: 16px; font-weight: 600; color: #303133; }
.editor-form { margin-top: 8px; }
.editor-tabs { margin-top: 16px; }
.tab-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}
.tab-tip { font-size: 12px; color: #909399; }
.json-textarea :deep(.el-textarea__inner) {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
}
</style>
