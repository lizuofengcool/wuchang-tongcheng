// axios 封装：统一处理 JWT token、错误码
import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'
import router from '@/router'

// 创建 axios 实例
const service = axios.create({
  baseURL: '/api/v1',
  timeout: 15000
})

// 请求拦截器：自动注入 JWT token
service.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// 响应拦截器：统一处理业务码
service.interceptors.response.use(
  (response) => {
    const res = response.data
    // 后端统一响应：{ code, message, data }
    if (res.code === 0) {
      return res
    }

    // 401 / Token 失效或未登录
    if (res.code === 401 || res.code === 1004 || res.code === 2006 || res.code === 2007 || res.code === 2008) {
      handleUnauthorized(res.message)
      return Promise.reject(new Error(res.message || '登录已失效'))
    }

    ElMessage.error(res.message || '请求失败')
    return Promise.reject(new Error(res.message || '请求失败'))
  },
  (error) => {
    // HTTP 层错误
    if (error.response) {
      const status = error.response.status
      if (status === 401) {
        handleUnauthorized('登录已失效，请重新登录')
      } else if (status === 403) {
        ElMessage.error('禁止访问')
      } else if (status === 500) {
        ElMessage.error('服务器内部错误')
      } else {
        ElMessage.error(`请求错误 (${status})`)
      }
    } else if (error.message.includes('timeout')) {
      ElMessage.error('请求超时，请稍后重试')
    } else {
      ElMessage.error('网络异常，请检查网络连接')
    }
    return Promise.reject(error)
  }
)

// 处理未授权：清 token、跳登录页
let unauthorizedShown = false
function handleUnauthorized(message) {
  if (unauthorizedShown) return
  unauthorizedShown = true
  ElMessageBox.confirm(message || '请先登录', '提示', {
    confirmButtonText: '重新登录',
    cancelButtonText: '取消',
    type: 'warning'
  })
    .then(() => {
      localStorage.removeItem('token')
      localStorage.removeItem('userInfo')
      router.push('/login')
    })
    .catch(() => {})
    .finally(() => {
      unauthorizedShown = false
    })
}

export default service
