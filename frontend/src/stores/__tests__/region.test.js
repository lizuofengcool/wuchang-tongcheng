// 单元测试：src/stores/region.js
// 覆盖 state 初始化（localStorage 回退默认地区）、currentRegionName 树遍历 getter、
// loadTree 成功/失败、setCurrentRegion 持久化
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'

// mock api/region：getRegionTree 返回地区树
const getRegionTreeMock = vi.fn()
vi.mock('@/api/region', () => ({
  getRegionTree: (...args) => getRegionTreeMock(...args)
}))

import { useRegionStore } from '../region'

// 构造一棵三层地区树用于 getter 遍历验证
function buildTree() {
  return [
    { id: 1, name: '湖北省', children: [
      { id: 2, name: '武汉市', children: [
        { id: 5, name: '武昌区', children: [] }
      ] },
      { id: 3, name: '宜昌市', children: [] }
    ] },
    { id: 4, name: '湖南省', children: [
      { id: 6, name: '长沙市', children: [] }
    ] }
  ]
}

describe('regionStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    getRegionTreeMock.mockReset()
  })

  describe('state 初始化', () => {
    it('localStorage 无值时回退默认地区 ID=2（与后端 DefaultRegionID 一致）', () => {
      const store = useRegionStore()
      expect(store.currentRegionId).toBe(2)
      expect(store.regionTree).toEqual([])
      expect(store.loaded).toBe(false)
    })

    it('localStorage 存在有效值时回填 currentRegionId', () => {
      localStorage.setItem('currentRegionId', '5')
      const store = useRegionStore()
      expect(store.currentRegionId).toBe(5)
    })
  })

  describe('currentRegionName getter', () => {
    it('命中根节点返回其 name', () => {
      const store = useRegionStore()
      store.regionTree = buildTree()
      store.currentRegionId = 1
      expect(store.currentRegionName).toBe('湖北省')
    })

    it('命中深层子节点（递归遍历）返回其 name', () => {
      const store = useRegionStore()
      store.regionTree = buildTree()
      store.currentRegionId = 5
      expect(store.currentRegionName).toBe('武昌区')
    })

    it('命中兄弟子树节点', () => {
      const store = useRegionStore()
      store.regionTree = buildTree()
      store.currentRegionId = 6
      expect(store.currentRegionName).toBe('长沙市')
    })

    it('未命中任何节点时回退“默认地区”', () => {
      const store = useRegionStore()
      store.regionTree = buildTree()
      store.currentRegionId = 999
      expect(store.currentRegionName).toBe('默认地区')
    })

    it('空树时回退“默认地区”', () => {
      const store = useRegionStore()
      store.regionTree = []
      expect(store.currentRegionName).toBe('默认地区')
    })
  })

  describe('loadTree action', () => {
    it('成功拉取 → regionTree 落库、loaded 置真', async () => {
      const tree = buildTree()
      getRegionTreeMock.mockResolvedValueOnce({ code: 0, data: tree })
      const store = useRegionStore()
      await store.loadTree()
      expect(getRegionTreeMock).toHaveBeenCalledTimes(1)
      expect(store.regionTree).toEqual(tree)
      expect(store.loaded).toBe(true)
    })

    it('data 为空数组时 regionTree 置空且 loaded 置真', async () => {
      getRegionTreeMock.mockResolvedValueOnce({ code: 0, data: [] })
      const store = useRegionStore()
      await store.loadTree()
      expect(store.regionTree).toEqual([])
      expect(store.loaded).toBe(true)
    })

    it('data 缺失时 regionTree 兜底为空数组', async () => {
      getRegionTreeMock.mockResolvedValueOnce({ code: 0 })
      const store = useRegionStore()
      await store.loadTree()
      expect(store.regionTree).toEqual([])
      expect(store.loaded).toBe(true)
    })

    it('请求异常 → 不抛错、regionTree 清空、loaded 保持假', async () => {
      getRegionTreeMock.mockRejectedValueOnce(new Error('network'))
      const store = useRegionStore()
      // 即便之前有数据，异常也清空
      store.regionTree = buildTree()
      await expect(store.loadTree()).resolves.toBeUndefined()
      expect(store.regionTree).toEqual([])
      expect(store.loaded).toBe(false)
    })
  })

  describe('setCurrentRegion action', () => {
    it('设置当前地区并持久化到 localStorage', () => {
      const store = useRegionStore()
      store.setCurrentRegion(5)
      expect(store.currentRegionId).toBe(5)
      expect(localStorage.getItem('currentRegionId')).toBe('5')
    })

    it('连续切换覆盖旧值', () => {
      const store = useRegionStore()
      store.setCurrentRegion(3)
      store.setCurrentRegion(6)
      expect(store.currentRegionId).toBe(6)
      expect(localStorage.getItem('currentRegionId')).toBe('6')
    })
  })
})
