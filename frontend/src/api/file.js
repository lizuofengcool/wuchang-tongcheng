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

// 申请 OSS STS 临时凭据（前端直传 OSS）
//
// 拿到临时 AK/SK/Token 后用 OSS 浏览器 SDK 或 SigV4 + x-amz-security-token 头
// 直接 PUT 到 OSS，一组凭据可在 expires_in 内上传任意多对象。
// 与 presign 互补；未配置 STS 时后端返回 1307，前端应回退 uploadFile / presignUpload。
export function getSTSCredentials() {
  return request.post('/file/sts')
}
