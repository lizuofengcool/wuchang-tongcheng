<template>
  <div class="app-container">
    <div class="page-card upload-card">
      <el-upload
        ref="uploadRef"
        class="upload-dragger"
        drag
        multiple
        :auto-upload="true"
        :show-file-list="false"
        :http-request="customUpload"
        accept="image/*,video/*,.pdf,.doc,.docx,.xls,.xlsx,.zip,.mp3"
      >
        <el-icon class="upload-icon"><UploadFilled /></el-icon>
        <div class="upload-text">将文件拖到此处，或<em>点击上传</em></div>
        <template #tip>
          <div class="upload-tip">支持图片/视频/文档/压缩包，单文件不超过 20MB</div>
        </template>
      </el-upload>
    </div>

    <div class="page-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <span class="title">已上传文件</span>
          <el-tag size="small">{{ fileList.length }} 个</el-tag>
        </div>
        <div class="toolbar-right">
          <el-button :icon="Delete" type="danger" plain :disabled="!fileList.length" @click="clearAll">清空列表</el-button>
        </div>
      </div>

      <el-table v-loading="uploading" :data="fileList" border stripe>
        <el-table-column label="预览" width="90">
          <template #default="{ row }">
            <el-image
              v-if="isImage(row.file_type)"
              :src="row.file_url"
              :preview-src-list="[row.file_url]"
              fit="cover"
              style="width: 60px; height: 60px; border-radius: 4px"
              :preview-teleported="true"
            />
            <el-icon v-else :size="32" class="file-icon">
              <Document />
            </el-icon>
          </template>
        </el-table-column>
        <el-table-column prop="file_name" label="文件名" min-width="200" show-overflow-tooltip />
        <el-table-column label="大小" width="110">
          <template #default="{ row }">{{ formatSize(row.file_size) }}</template>
        </el-table-column>
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ row.file_type || row.mime_type || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="file_url" label="访问URL" min-width="220" show-overflow-tooltip />
        <el-table-column label="上传时间" width="170">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="copyUrl(row)">复制URL</el-button>
            <el-button type="warning" link size="small" @click="removeRow(row)">移除</el-button>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="暂无上传文件" />
        </template>
      </el-table>
    </div>

    <!-- 当前上传进度 -->
    <el-dialog v-model="progressVisible" title="上传中" width="420px" :close-on-click-modal="false" :show-close="false">
      <el-progress :percentage="progress" :status="progress === 100 ? 'success' : ''" />
      <div class="progress-text">{{ currentFileName }}</div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { UploadFilled, Document, Delete } from '@element-plus/icons-vue'
import { uploadFile } from '@/api/file'
import { formatSize, formatTime } from '@/utils/format'

const uploadRef = ref(null)
const fileList = ref([])
const uploading = ref(false)
const progressVisible = ref(false)
const progress = ref(0)
const currentFileName = ref('')

const isImage = (type) => type === 'image' || (type && type.startsWith('image'))

const customUpload = async (options) => {
  const { file } = options
  // 校验大小
  if (file.size > 20 * 1024 * 1024) {
    ElMessage.error(`${file.name} 超过 20MB 限制`)
    return
  }
  uploading.value = true
  progressVisible.value = true
  progress.value = 0
  currentFileName.value = file.name
  try {
    const res = await uploadFile(file, (p) => {
      progress.value = p
    })
    const data = res.data || {}
    fileList.value.unshift({
      id: data.id || Date.now(),
      file_name: data.file_name || file.name,
      file_url: data.file_url || '',
      file_size: data.file_size || file.size,
      file_type: data.file_type || '',
      mime_type: data.mime_type || '',
      created_at: data.created_at || new Date().toISOString()
    })
    progress.value = 100
    ElMessage.success('上传成功')
  } catch (e) {
    ElMessage.error('上传失败')
  } finally {
    setTimeout(() => {
      progressVisible.value = false
      uploading.value = false
    }, 400)
  }
}

const copyUrl = (row) => {
  const url = row.file_url || ''
  if (!url) {
    ElMessage.warning('URL为空')
    return
  }
  // 兼容旧API：navigator.clipboard 不可用时回退
  if (navigator.clipboard?.writeText) {
    navigator.clipboard.writeText(url).then(() => ElMessage.success('已复制'))
  } else {
    const ta = document.createElement('textarea')
    ta.value = url
    document.body.appendChild(ta)
    ta.select()
    try {
      document.execCommand('copy')
      ElMessage.success('已复制')
    } catch {
      ElMessage.warning('复制失败，请手动复制')
    }
    document.body.removeChild(ta)
  }
}

const removeRow = (row) => {
  fileList.value = fileList.value.filter((i) => i.id !== row.id)
}

const clearAll = async () => {
  try {
    await ElMessageBox.confirm('确定清空当前文件列表吗？（仅清空展示，不影响已上传文件）', '提示', { type: 'warning' })
    fileList.value = []
    ElMessage.success('已清空')
  } catch (e) {
    // 取消
  }
}

onMounted(() => {
  // 列表为空提示
})
</script>

<style scoped>
.upload-card {
  margin-bottom: 16px;
}
.upload-dragger {
  width: 100%;
}
.upload-dragger :deep(.el-upload-dragger) {
  width: 100%;
  padding: 40px 20px;
}
.upload-icon {
  font-size: 64px;
  color: #c0c4cc;
  margin-bottom: 12px;
}
.upload-text {
  color: #606266;
  font-size: 14px;
}
.upload-text em {
  color: #409eff;
  font-style: normal;
}
.upload-tip {
  color: #909399;
  font-size: 12px;
  margin-top: 8px;
}
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.toolbar-left {
  display: flex;
  align-items: center;
  gap: 8px;
}
.toolbar-left .title {
  font-size: 16px;
  font-weight: 600;
}
.file-icon {
  color: #909399;
}
.progress-text {
  margin-top: 12px;
  text-align: center;
  color: #606266;
  font-size: 13px;
  word-break: break-all;
}
</style>
