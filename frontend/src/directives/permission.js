// v-permission / v-role 自定义指令
// 用法：
//   <el-button v-permission="'user:create'">新增</el-button>
//   <el-button v-permission="['user:update','user:delete']">操作</el-button>（任一满足即显示）
//   <el-button v-role="'admin'">超管操作</el-button>
//   <div v-permission.hide="'user:create'">无权限时隐藏（默认移除）</div>
import { hasPermission, hasRole } from '@/utils/auth'

function evaluate(binding, checker) {
  // 修饰符 .hide 表示无权限时仅隐藏（display:none），默认移除元素
  const value = binding.value
  return checker(value)
}

const permissionDirective = {
  mounted(el, binding) {
    if (evaluate(binding, hasPermission)) return
    if (binding.modifiers.hide) {
      el.style.display = 'none'
    } else {
      el.parentNode && el.parentNode.removeChild(el)
    }
  },
  updated(el, binding) {
    // 仅对 .hide 模式有意义（DOM 已移除的无法恢复）
    if (binding.modifiers.hide) {
      el.style.display = evaluate(binding, hasPermission) ? '' : 'none'
    }
  }
}

const roleDirective = {
  mounted(el, binding) {
    if (evaluate(binding, hasRole)) return
    if (binding.modifiers.hide) {
      el.style.display = 'none'
    } else {
      el.parentNode && el.parentNode.removeChild(el)
    }
  },
  updated(el, binding) {
    if (binding.modifiers.hide) {
      el.style.display = evaluate(binding, hasRole) ? '' : 'none'
    }
  }
}

export { permissionDirective as permission, roleDirective as role }
