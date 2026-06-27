// 地区状态管理（Pinia）
// 维护地区树与当前选中的地区ID，用于数据隔离（请求头 X-Region-ID）
import { defineStore } from 'pinia'
import { getRegionTree } from '@/api/region'

const STORAGE_KEY = 'currentRegionId'
// 默认地区：武汉市（与后端 DefaultRegionID 一致）
const DEFAULT_REGION_ID = 2

export const useRegionStore = defineStore('region', {
  state: () => ({
    regionTree: [],
    currentRegionId: Number(localStorage.getItem(STORAGE_KEY)) || DEFAULT_REGION_ID,
    loaded: false
  }),
  getters: {
    currentRegionName: (state) => {
      const walk = (nodes) => {
        for (const n of nodes || []) {
          if (n.id === state.currentRegionId) return n.name
          const f = walk(n.children)
          if (f) return f
        }
        return ''
      }
      return walk(state.regionTree) || '默认地区'
    }
  },
  actions: {
    // 拉取地区树
    async loadTree() {
      try {
        const res = await getRegionTree()
        this.regionTree = res.data || []
        this.loaded = true
      } catch (e) {
        this.regionTree = []
      }
    },
    // 切换当前地区
    setCurrentRegion(id) {
      this.currentRegionId = id
      localStorage.setItem(STORAGE_KEY, String(id))
    }
  }
})
