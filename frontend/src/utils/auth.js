// 权限/角色校验工具（供模板表达式或 JS 调用）
import { useUserStore } from '@/stores/user'

/**
 * 是否拥有指定权限码（admin 角色直通）
 * @param {string|string[]} code 权限码或权限码数组（任一满足即通过）
 */
export function hasPermission(code) {
  const userStore = useUserStore()
  if (userStore.isSuperAdmin) return true
  if (!code) return true
  const codes = Array.isArray(code) ? code : [code]
  return codes.some((c) => userStore.permissions.includes(c))
}

/**
 * 是否拥有指定角色码
 * @param {string|string[]} code 角色码或角色码数组（任一满足即通过）
 */
export function hasRole(code) {
  const userStore = useUserStore()
  if (!code) return true
  const codes = Array.isArray(code) ? code : [code]
  return codes.some((c) => userStore.roles.includes(c))
}

/**
 * 是否拥有全部指定权限码
 * @param {string[]} codes 权限码数组（全部满足才通过）
 */
export function hasAllPermissions(codes) {
  const userStore = useUserStore()
  if (userStore.isSuperAdmin) return true
  if (!codes || !codes.length) return true
  return codes.every((c) => userStore.permissions.includes(c))
}

export default { hasPermission, hasRole, hasAllPermissions }
