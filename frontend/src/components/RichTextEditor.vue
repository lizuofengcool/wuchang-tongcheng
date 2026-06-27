<template>
  <div class="rich-editor" :class="{ disabled }">
    <div class="re-toolbar">
      <button type="button" class="re-btn" title="加粗" @click="exec('bold')"><b>B</b></button>
      <button type="button" class="re-btn" title="斜体" @click="exec('italic')"><i>I</i></button>
      <button type="button" class="re-btn" title="下划线" @click="exec('underline')"><u>U</u></button>
      <span class="re-divider" />
      <button type="button" class="re-btn" title="一级标题" @click="block('H1')">H1</button>
      <button type="button" class="re-btn" title="二级标题" @click="block('H2')">H2</button>
      <button type="button" class="re-btn" title="三级标题" @click="block('H3')">H3</button>
      <span class="re-divider" />
      <button type="button" class="re-btn" title="无序列表" @click="exec('insertUnorderedList')">• 列表</button>
      <button type="button" class="re-btn" title="有序列表" @click="exec('insertOrderedList')">1. 列表</button>
      <span class="re-divider" />
      <button type="button" class="re-btn" title="插入链接" @click="insertLink">链接</button>
      <button type="button" class="re-btn" title="插入图片" @click="triggerImage">图片</button>
      <span class="re-divider" />
      <button type="button" class="re-btn" title="清除格式" @click="exec('removeFormat')">清除</button>
      <label class="re-source-toggle">
        <input type="checkbox" v-model="sourceMode" /> HTML 源码
      </label>
    </div>

    <!-- 富文本编辑区 -->
    <div
      v-show="!sourceMode"
      ref="editorRef"
      class="re-content"
      contenteditable="true"
      :style="{ minHeight: minHeight + 'px' }"
      @input="onInput"
      @blur="onInput"
    />

    <!-- HTML 源码模式 -->
    <textarea
      v-show="sourceMode"
      class="re-source"
      :style="{ minHeight: minHeight + 'px' }"
      :value="modelValue"
      @input="onSourceInput"
    />

    <input ref="fileInputRef" type="file" accept="image/*" style="display:none" @change="onImageChange" />
    <div v-if="uploading" class="re-upload-tip">图片上传中... {{ progress }}%</div>
  </div>
</template>

<script setup>
import { ref, watch, onMounted, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { uploadFile } from '@/api/file'

const props = defineProps({
  modelValue: { type: String, default: '' },
  minHeight: { type: Number, default: 240 },
  disabled: { type: Boolean, default: false }
})
const emit = defineEmits(['update:modelValue'])

const editorRef = ref(null)
const fileInputRef = ref(null)
const sourceMode = ref(false)
const uploading = ref(false)
const progress = ref(0)

// 内部同步标志，避免回写触发循环
let syncing = false

const setHTML = (html) => {
  if (!editorRef.value) return
  editorRef.value.innerHTML = html || ''
}

onMounted(() => {
  setHTML(props.modelValue)
})

// 外部值变化时同步到编辑器（非编辑态）
watch(
  () => props.modelValue,
  (val) => {
    if (syncing) return
    if (editorRef.value && editorRef.value.innerHTML !== val) {
      setHTML(val)
    }
  }
)

const onInput = () => {
  if (sourceMode.value) return
  syncing = true
  emit('update:modelValue', editorRef.value.innerHTML)
  nextTick(() => (syncing = false))
}

const onSourceInput = (e) => {
  syncing = true
  emit('update:modelValue', e.target.value)
  nextTick(() => (syncing = false))
}

const exec = (cmd, val = null) => {
  if (sourceMode.value) return
  editorRef.value?.focus()
  document.execCommand(cmd, false, val)
  onInput()
}

const block = (tag) => {
  if (sourceMode.value) return
  editorRef.value?.focus()
  document.execCommand('formatBlock', false, tag)
  onInput()
}

const insertLink = () => {
  if (sourceMode.value) return
  const sel = window.getSelection()?.toString()
  const url = window.prompt('请输入链接地址（含 http/https）', 'https://')
  if (!url) return
  exec('createLink', url)
  if (!sel) {
    // 无选中文本时，补一个文本节点
    exec('insertText', url)
  }
}

const triggerImage = () => {
  if (sourceMode.value) return
  fileInputRef.value?.click()
}

const onImageChange = async (e) => {
  const file = e.target.files?.[0]
  if (!file) return
  if (file.size > 5 * 1024 * 1024) {
    ElMessage.error('图片不能超过 5MB')
    e.target.value = ''
    return
  }
  uploading.value = true
  progress.value = 0
  try {
    const res = await uploadFile(file, (p) => (progress.value = p))
    const url = res.data?.file_url
    if (url) {
      editorRef.value?.focus()
      document.execCommand('insertImage', false, url)
      onInput()
    }
  } catch (err) {
    // 错误已由 request 拦截器提示
  } finally {
    uploading.value = false
    e.target.value = ''
  }
}
</script>

<style scoped>
.rich-editor {
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  background: #fff;
}
.rich-editor.disabled {
  opacity: 0.6;
  pointer-events: none;
}
.re-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px;
  padding: 6px 8px;
  border-bottom: 1px solid #ebeef5;
  background: #fafafa;
}
.re-btn {
  min-width: 30px;
  height: 28px;
  padding: 0 8px;
  border: 1px solid #dcdfe6;
  background: #fff;
  border-radius: 3px;
  cursor: pointer;
  font-size: 13px;
  color: #303133;
}
.re-btn:hover {
  background: #ecf5ff;
  border-color: #409eff;
  color: #409eff;
}
.re-divider {
  display: inline-block;
  width: 1px;
  height: 18px;
  background: #dcdfe6;
  margin: 0 4px;
}
.re-source-toggle {
  margin-left: auto;
  font-size: 12px;
  color: #909399;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 4px;
}
.re-content {
  padding: 10px 12px;
  outline: none;
  font-size: 14px;
  line-height: 1.7;
  overflow-y: auto;
}
.re-content :deep(h1) { font-size: 1.6em; margin: 0.4em 0; }
.re-content :deep(h2) { font-size: 1.4em; margin: 0.4em 0; }
.re-content :deep(h3) { font-size: 1.2em; margin: 0.4em 0; }
.re-content :deep(ul),
.re-content :deep(ol) { padding-left: 24px; }
.re-content :deep(img) { max-width: 100%; height: auto; border-radius: 4px; }
.re-content :deep(a) { color: #409eff; }
.re-source {
  width: 100%;
  padding: 10px 12px;
  border: none;
  outline: none;
  font-family: Menlo, Consolas, monospace;
  font-size: 13px;
  resize: vertical;
  box-sizing: border-box;
}
.re-upload-tip {
  padding: 6px 12px;
  font-size: 12px;
  color: #909399;
  border-top: 1px solid #ebeef5;
}
</style>
