// 文件上传模块 API 封装
import request from '@/utils/request'

// 上传单个文件
export function uploadFile(file, onProgress) {
  const formData = new FormData()
  formData.append('file', file)
  return request.post('/file/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    onUploadProgress: (e) => {
      if (onProgress && e.total) {
        onProgress(Math.round((e.loaded * 100) / e.total))
      }
    }
  })
}

// 获取文件列表
export function listFiles(params) {
  return request.get('/file', { params })
}

// 删除文件
export function deleteFile(id) {
  return request.delete(`/file/${id}`)
}
