// 用户状态管理（Pinia）
import { defineStore } from 'pinia'
import { login as loginApi, getUserInfo as fetchUserInfo } from '@/api/user'

export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    userInfo: JSON.parse(localStorage.getItem('userInfo') || 'null')
  }),
  getters: {
    isLoggedIn: (state) => !!state.token,
    nickname: (state) => state.userInfo?.nickname || state.userInfo?.username || '管理员',
    avatar: (state) => state.userInfo?.avatar || ''
  },
  actions: {
    // 登录
    async login(payload) {
      const res = await loginApi(payload)
      const { token, user_info } = res.data
      this.token = token
      this.userInfo = user_info
      localStorage.setItem('token', token)
      localStorage.setItem('userInfo', JSON.stringify(user_info))
      return res
    },
    // 拉取最新用户信息
    async fetchProfile() {
      const res = await fetchUserInfo()
      this.userInfo = res.data
      localStorage.setItem('userInfo', JSON.stringify(res.data))
      return res
    },
    // 退出登录
    logout() {
      this.token = ''
      this.userInfo = null
      localStorage.removeItem('token')
      localStorage.removeItem('userInfo')
    }
  }
})
