// 用户状态管理（Pinia）
import { defineStore } from 'pinia'
import { login as loginApi, getUserInfo as fetchUserInfo } from '@/api/user'
import { myAuth } from '@/api/permission'

export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    userInfo: JSON.parse(localStorage.getItem('userInfo') || 'null'),
    // 权限码与角色码（供 v-permission / v-role 指令使用）
    permissions: JSON.parse(localStorage.getItem('permissions') || '[]'),
    roles: JSON.parse(localStorage.getItem('roles') || '[]')
  }),
  getters: {
    isLoggedIn: (state) => !!state.token,
    nickname: (state) => state.userInfo?.nickname || state.userInfo?.username || '管理员',
    avatar: (state) => state.userInfo?.avatar || '',
    isSuperAdmin: (state) => state.roles.includes('admin')
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
      // 登录后立即拉取权限
      await this.fetchAuth()
      return res
    },
    // 拉取最新用户信息
    async fetchProfile() {
      const res = await fetchUserInfo()
      this.userInfo = res.data
      localStorage.setItem('userInfo', JSON.stringify(res.data))
      return res
    },
    // 拉取权限/角色码
    async fetchAuth() {
      try {
        const res = await myAuth()
        this.permissions = res.data.permissions || []
        this.roles = res.data.roles || []
        localStorage.setItem('permissions', JSON.stringify(this.permissions))
        localStorage.setItem('roles', JSON.stringify(this.roles))
      } catch (e) {
        // 静默失败，避免阻塞登录流程
      }
    },
    // 退出登录
    logout() {
      this.token = ''
      this.userInfo = null
      this.permissions = []
      this.roles = []
      localStorage.removeItem('token')
      localStorage.removeItem('userInfo')
      localStorage.removeItem('permissions')
      localStorage.removeItem('roles')
    }
  }
})
