<template>
  <view class="container">
    <view v-if="!userInfo" class="card">
      <view class="title-md mb-2">登录</view>
      <view class="mb-2">
        <input
          v-model="form.username"
          class="input"
          placeholder="用户名"
        />
      </view>
      <view class="mb-2">
        <input
          v-model="form.password"
          class="input"
          password
          placeholder="密码"
          @confirm="doLogin"
        />
      </view>
      <button
        class="btn-primary"
        :disabled="loginLoading"
        @tap="doLogin"
      >{{ loginLoading ? '登录中…' : '登录' }}</button>
      <view class="text-sm mt-2" style="text-align:center;">
        默认管理员：admin / admin123
      </view>
    </view>

    <view v-else>
      <view class="card flex-between">
        <view>
          <view class="title-md">{{ userInfo.username }}</view>
          <view class="text-sm mt-2">{{ userInfo.nickname || userInfo.email || '—' }}</view>
        </view>
        <view class="text-sm text-brand" @tap="doLogout">退出登录</view>
      </view>

      <view class="card">
        <view class="title-md mb-2">我的地区</view>
        <view class="text-sm">当前：{{ currentRegionName }}</view>
        <view class="flex mt-2" style="flex-wrap:wrap;gap:6px;">
          <view
            v-for="r in regions"
            :key="r.id"
            class="region-item"
            :class="{ active: r.id === currentRegionId }"
            @tap="switchRegion(r)"
          >{{ r.name }}</view>
        </view>
      </view>
    </view>
  </view>
</template>

<script>
import { login, getProfile, logout } from '@/api/user'
import { listRegions } from '@/api/news'

export default {
  data() {
    return {
      userInfo: null,
      form: { username: '', password: '' },
      loginLoading: false,
      regions: [],
      currentRegionId: 2,
      currentRegionName: '武汉市',
    }
  },
  onShow() {
    this.currentRegionId = uni.getStorageSync('regionId') || 2
    if (uni.getStorageSync('token')) {
      this.loadProfile()
    }
    this.loadRegions()
  },
  methods: {
    async doLogin() {
      if (!this.form.username || !this.form.password) {
        uni.showToast({ title: '请输入用户名和密码', icon: 'none' })
        return
      }
      this.loginLoading = true
      try {
        const data = await login(this.form.username, this.form.password)
        uni.setStorageSync('token', data.token)
        uni.showToast({ title: '登录成功', icon: 'success' })
        await this.loadProfile()
      } catch (e) {} finally {
        this.loginLoading = false
      }
    },
    async loadProfile() {
      try {
        this.userInfo = await getProfile()
      } catch (e) {
        this.userInfo = null
      }
    },
    doLogout() {
      logout()
      this.userInfo = null
      uni.showToast({ title: '已退出登录', icon: 'none' })
    },
    async loadRegions() {
      try {
        this.regions = (await listRegions()) || []
        const r = this.regions.find((x) => x.id === this.currentRegionId)
        if (r) this.currentRegionName = r.name
      } catch (e) {}
    },
    switchRegion(r) {
      this.currentRegionId = r.id
      this.currentRegionName = r.name
      uni.setStorageSync('regionId', r.id)
      uni.showToast({ title: `已切换到 ${r.name}`, icon: 'none' })
    },
  },
}
</script>

<style>
.input {
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  padding: 10px 12px;
  font-size: 14px;
  width: 100%;
  box-sizing: border-box;
}
.btn-primary {
  background: #dc2626;
  color: #fff;
  border-radius: 6px;
  font-size: 14px;
  margin-top: 8px;
}
.region-item {
  padding: 4px 10px;
  font-size: 12px;
  background: #f3f4f6;
  color: #374151;
  border-radius: 14px;
}
.region-item.active {
  background: #dc2626;
  color: #fff;
}
</style>
