// Vitest 全局 setup：为 node 环境提供最小 localStorage 实现
// user store 的 state 初始化器会读取 localStorage（token/userInfo/permissions/roles），
// node 环境下无 localStorage 会抛 ReferenceError，此处用内存 Map 兜底。
const memStore = new Map()

globalThis.localStorage = {
  getItem: (key) => (memStore.has(key) ? memStore.get(key) : null),
  setItem: (key, value) => memStore.set(key, String(value)),
  removeItem: (key) => memStore.delete(key),
  clear: () => memStore.clear(),
  key: () => null,
  get length() {
    return memStore.size
  }
}
