// 后端 API 封装（uni-app 版本）
// 统一处理：baseURL、token 注入、地区头注入、业务码处理

const BASE_URL = (() => {
  // 微信小程序：必须用绝对地址
  // H5：dev 时 vite proxy 已代理 /api，prod 时由部署服务器反代
  // #ifdef MP-WEIXIN || APP-PLUS
  return 'http://localhost:8080' // 小程序发布前需改为线上域名
  // #endif
  // #ifdef H5
  return '' // H5 走相对路径，由 vite proxy / nginx 反代
  // #endif
})()

// 统一请求封装
export function request(options) {
  const {
    url,
    method = 'GET',
    data = {},
    header = {},
    requireAuth = false,
  } = options

  // 注入 JWT
  if (requireAuth) {
    const token = uni.getStorageSync('token')
    if (!token) {
      uni.showToast({ title: '请先登录', icon: 'none' })
      return Promise.reject(new Error('未登录'))
    }
    header.Authorization = `Bearer ${token}`
  }

  // 注入地区 ID
  const regionId = uni.getStorageSync('regionId') || 2
  header['X-Region-ID'] = String(regionId)
  header['Content-Type'] = header['Content-Type'] || 'application/json'

  return new Promise((resolve, reject) => {
    uni.request({
      url: `${BASE_URL}${url}`,
      method,
      data,
      header,
      success: (res) => {
        if (res.statusCode === 401) {
          uni.showToast({ title: '登录已过期', icon: 'none' })
          uni.removeStorageSync('token')
          reject(new Error('未授权'))
          return
        }
        if (res.statusCode >= 400) {
          uni.showToast({ title: `请求失败 (${res.statusCode})`, icon: 'none' })
          reject(new Error(`HTTP ${res.statusCode}`))
          return
        }
        const body = res.data
        if (body.code !== 0) {
          uni.showToast({ title: body.message || '请求失败', icon: 'none' })
          reject(new Error(body.message || '业务错误'))
          return
        }
        resolve(body.data)
      },
      fail: (err) => {
        uni.showToast({ title: '网络错误', icon: 'none' })
        reject(err)
      },
    })
  })
}
