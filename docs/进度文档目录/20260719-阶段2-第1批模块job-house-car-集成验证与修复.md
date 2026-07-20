# 【阶段2 第1批模块集成验证与 car 修复】进度文档

- 开发人员：AI助手
- 创建时间：2026-07-20
- 当前状态：✅ Playwright 49/49 页面全部 PASS + car API 路径修复完成

## 一、开发目标

依据用户规则"按大厂处理"和阶段2进度文档中的"待完成 - 集成验证：Playwright完整验证52个页面渲染"，对 job/house/car 三模块的 49 个管理后台页面（不含 3 个 detail/:id hidden 路由）进行 Playwright 渲染验证，发现问题立即修复。

## 二、Playwright 52 页面渲染验证结果

### 总体结果

| 模块 | 总页面 | PASS | FAIL | 状态 |
|------|--------|------|------|------|
| job 招聘求职 | 16 | 16 | 0 | ✅ 全部通过 |
| fang 房屋租售 | 17 | 17 | 0 | ✅ 全部通过 |
| car 车辆买卖 | 16 | 0→16 | 16→0 | ✅ 修复后全部通过 |
| **总计** | **49** | **49** | **0** | ✅ 100% 通过 |

> 注：3 个 detail/:id hidden 路由未纳入验证（需要参数）。

### 验证方法
- 工具：MCP Playwright（chromium, headless）
- 账号：admin / admin123（super_admin 角色）
- 每个页面访问后：
  1. `playwright_console_logs(type=error, clear=true)` 检查控制台错误
  2. `playwright_get_visible_text` 获取页面文字确认有内容
- 并行验证：3 个 browser_use subagent 同时执行（job 剩余/fang 全部/car 全部）

### 控制台错误检查
- 所有 49 个页面均无 JavaScript 错误
- fang/statistics 页面有 1 条 ECharts DOM 警告（不阻断渲染）
- 无 Vue 路由错误、无 API 401/403/404/500 错误、无组件加载失败

## 三、关键修复：car 模块 API 路径不匹配

### 问题诊断

**现象**：car 模块 16 个页面全部提示"请求的接口不存在"，列表数据全部为空

**根因**：
- 前端 `frontend/src/api/car.js` 调用 `/admin/car/*`（admin 在前）
- 后端 `backend/internal/modules/car/plugin.go` 实际路径是 `/api/v1/car/admin/*`（admin 在后，line 327: `admin := router.Group("/admin")`）
- 路径不匹配导致 404

**对比验证**：
| 模块 | 前端调用路径 | 后端实际路径 | 是否匹配 |
|------|------------|------------|---------|
| job | `/job/admin/*` | `/api/v1/job/admin/*` | ✅ |
| house(fang) | `/house/admin/*` | `/api/v1/house/admin/*` | ✅ |
| car | `/admin/car/*`（错误） | `/api/v1/car/admin/*` | ❌ |

### 修复方案

将 `frontend/src/api/car.js` 中所有 `/admin/car/` 替换为 `/car/admin/`：
- 共替换 **60 处** API 调用路径
- 同步修改文件头部注释（line 2）
- Vite 自动热更新，无需重新构建

### 修复验证

**后端 API 测试**（PowerShell Invoke-RestMethod）：
```
/api/v1/car/admin/cars => code:0 msg:success
/api/v1/car/admin/listings => code:0 msg:success
/api/v1/car/admin/inspections => code:0 msg:success
/api/v1/car/admin/evaluations => code:0 msg:success
/api/v1/car/admin/financings => code:0 msg:success
/api/v1/car/admin/insurances => code:0 msg:success
/api/v1/car/admin/test-drives => code:0 msg:success
/api/v1/car/admin/transfers => code:0 msg:success
/api/v1/car/admin/escrows => code:0 msg:success
/api/v1/car/admin/contracts => code:0 msg:success
/api/v1/car/admin/reviews => code:0 msg:success
/api/v1/car/admin/reports => code:0 msg:success
/api/v1/car/admin/recommendations => code:0 msg:success
/api/v1/car/admin/audit-rules => code:0 msg:success
/api/v1/car/admin/statistics/overview => code:0 msg:success
/api/v1/car/admin/statistics => code:0 msg:success
```
16/16 API 全部返回 success ✅

**前端 Playwright 重新验证**：
- 修复前：16/16 FAIL（"请求的接口不存在"）
- 修复后：16/16 PASS ✅

## 四、修改的文件

### 1. `frontend/src/api/car.js`
- 修改 60 处 API 调用路径：`/admin/car/*` → `/car/admin/*`
- 修改文件头部注释（line 2）
- 涵盖：车源/发布单/检测/评估/分期/车险/试驾/过户/合同/担保/评价/举报/统计/推荐/审核规则/车型库 全部 16 类资源

## 五、约束（公共要求）

1. 不修复任何非必要的代码（仅修复发现的 bug）
2. 修复后必须重新验证
3. 验证报告必须详细记录每个页面的状态
4. 修改文件最小化（仅 car.js 一个文件）

## 六、遇到的问题与解决方案

### 问题1：car 模块 16 页面全部"请求的接口不存在"
- **现象**：Playwright 验证 car 模块 16 个页面，全部显示"请求的接口不存在"
- **原因**：前端 car.js 把 admin 写在 car 前面（`/admin/car/*`），与后端路径 `/car/admin/*` 不匹配
- **解决方案**：批量替换 60 处路径，重启验证全部通过

### 问题2：car 模块部分页面首次访问跳转到 fang 模块
- **现象**：/business/car/models 和 /business/car/test-drives 首次访问跳转到 fang 模块
- **原因**：浏览器状态残留（subagent 验证过程中浏览器状态混乱）
- **解决方案**：路由定义本身没问题，修复 car.js 后重新验证未出现跳转错误

### 问题3：fang/statistics 页面 ECharts DOM 警告
- **现象**：控制台 1 条 ECharts 警告 "Can't get DOM width or height"
- **原因**：ECharts 在容器渲染前尝试获取尺寸
- **解决方案**：不阻断渲染，暂不修复（属于已有问题，不在本次修复范围）

## 七、下一步计划

1. ✅ Playwright 49/49 页面验证通过
2. ✅ car 模块 API 路径修复完成
3. ⏳ commit push 修复到 Gitee
4. ⏳ 启动阶段3第2批模块开发（love/pinche/linggong/dh114）
