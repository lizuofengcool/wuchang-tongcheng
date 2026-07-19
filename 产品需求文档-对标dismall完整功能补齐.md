# 产品需求文档 - 对标 dismall 完整功能补齐

> **文档性质**：项目唯一开发依据（依据用户规则第1条）
> **创建日期**：2026-07-19
> **对标来源**：
> - 西瓜先生 https://addon.dismall.com/developer-7402.html （59个应用）
> - 点微科技 https://addon.dismall.com/developer-26633.html （58个应用）
> **目标**：打造五常同城完整本地生活服务平台，对标 dismall 同城系列插件全部核心功能模块

---

## 一、项目架构（4端技术端 + 5端业务端）

> **重要更新（2026-07-19）**：经多端架构分析，原"5端架构"中 H5小程序和原生App合并为 uniappx 一套代码统一开发。
> 原 `frontend/miniapp/` (uni-app 3) 废弃，所有移动端能力统一由 `frontend/app/` (uniappx + UTS) 承载。

### 1.1 端架构总览

#### 1.1.1 技术端（部署维度，4端）

| 序号 | 端名称 | 目录 | 技术栈 | 说明 |
|------|--------|------|--------|------|
| 1 | 后端 API | `backend/` | Go 1.22+ (Gin + GORM + 插件化) | 统一 API 服务，所有端共用 |
| 2 | 管理后台 | `frontend/` | Vue 3 + Vite + Element Plus + Pinia | 平台管理端（M端），三级菜单 |
| 3 | PC门户 | `frontend/pc/` | Next.js 16.2 + React 19 + TypeScript + Tailwind | 桌面端门户站点（SEO引流） |
| 4 | 统一移动端 | `frontend/app/` | **uniappx + UTS + Vue 3** | 一套代码出 Android/iOS/鸿蒙/Web/微信小程序 5端 |

#### 1.1.2 业务端（角色维度，5端）

| 序号 | 业务端 | 用户群体 | 技术载体 | 入口路径 |
|------|--------|----------|----------|----------|
| 1 | **C端-用户端** | 普通用户/VIP | uniappx App（角色=用户） | App默认TabBar |
| 2 | **B端-商家端** | 商家/店员 | uniappx App（角色切换=商家） | App登录后角色切换 |
| 3 | **D端-骑手端** | 配送员/跑腿员 | uniappx App（角色切换=骑手） | App登录后角色切换 |
| 4 | **M端-平台管理端** | 平台运营/审核员/财务 | Web管理后台 | admin.wctc.wang |
| 5 | **A端-代理端**（可选） | 城市代理/合伙人 | Web管理后台子站点 | agent.wctc.wang |

#### 1.1.3 一套App多角色切换机制

uniappx App 登录后，根据用户角色动态切换 TabBar 与功能模块：

- **普通用户**：默认 TabBar（首页/分类/发布/消息/我的）
- **VIP用户**：普通用户 TabBar + 会员特权入口
- **商家**：切换至商家 TabBar（店铺/商品/订单/营销/数据）
- **店员**：商家 TabBar 的子集（仅订单/核销）
- **骑手**：切换至骑手 TabBar（接单/配送/签到/收益）
- **代理商**：H5 进入管理后台子站点（限地区）
- **平台运营/超管**：Web 管理后台（不在 App 内）

### 1.2 技术选型说明

#### uniappx 选型理由（统一移动端）
- DCloud 下一代跨平台框架，UTS 编译为纯原生代码（性能是 uni-app 3 的 3-5 倍）
- 一套代码同时输出 Android/iOS/鸿蒙/Web/微信小程序 5端
- 演示项目 `五常同城uniappx/` 已包含 30+ uni-ui-x 组件 + uni-pay-x（微信/支付宝/Apple IAP）+ uni-map-common（高德/腾讯地图）+ uni-stat（统计）+ uni-upgrade-center-app（升级中心）
- 性能接近原生，适合同城App高频交互场景（IM/地图/支付/推送）

#### 为何废弃 uni-app 3（frontend/miniapp/）
- uni-app 3 性能弱于 uniappx（解释执行 vs UTS 编译为原生）
- miniapp 项目无任何 uni-ui 组件，仅基础 @dcloudio/uni-app
- 维护两套移动端代码（uni-app 3 + uniappx）成本高且无收益
- uniappx 已支持微信小程序编译，uni-app 3 完全可被替代
- 已开发的首页/头条/搜索/我的 5 页需在 uniappx 中重新实现（成本可控，且性能更优）

#### 为何保留 Next.js 16（PC门户）
- SSR 服务端渲染，对 SEO 极其重要（同城分类信息 80% 流量来自搜索引擎）
- uniappx 编译到 Web 不支持 SSR，无法替代

#### 为何保留 Vue 3 + Element Plus（管理后台）
- 后台管理复杂表单/表格/权限交互，Element Plus 生态最成熟
- uniappx 不适合做后台管理界面

### 1.3 数据库与基础设施

- **PostgreSQL 16 + PostGIS**：业务表统一继承 `RegionBaseModel`（含 RegionID 地区隔离）
- **Redis 7**：cache-aside 缓存（region/category 树 30min、列表 60s）
- **Elasticsearch 8**：全文检索（已接入 news，扩展到二手/招聘等）
- **RabbitMQ**：异步索引解耦
- **对象存储**：LocalStorage + MinIO + 七牛云 Kodo + 阿里云 OSS STS 直传
- **容器化**：Docker + Docker Compose
- **CI/CD**：GitHub Actions（backend go vet/build/test、frontend npm test/build、tag 触发 docker publish）
- **监控告警**：Prometheus + Grafana（规划中）
- **日志**：ELK（规划中）

### 1.4 核心架构原则（"一次选型，终身不重构"）

> 依据用户与 AI 早期技术选型对话，确立 4 条不可违背的架构原则：

1. **模块化单体起步，预留微服务拆分能力**：初期单二进制部署，模块间通过接口通信；后期有性能瓶颈再拆微服务，业务代码无需改动。模块化单体足够支撑 10 万日活。
2. **一套业务逻辑，多端复用**：后端 API 统一，所有端（PC/H5/小程序/App）共用 `/api/v1/*`，禁止端独立写后端逻辑。
3. **插件化架构**：每个业务模块是独立 Go 插件，实现 `Plugin` 接口（Name/Version/Init/Install/Uninstall/Enable/Disable），像 Discuz 插件一样按需安装/卸载/启用/禁用，不安装的模块代码根本不加载。这是"以后不重构"的关键。
4. **多端覆盖范围**：本项目最终覆盖 **6 端**：
   - PC 门户站（Next.js 16.2 + React 19，SSR 利于 SEO）
   - 管理后台（Vue 3 + Element Plus）
   - H5 移动端（uni-app 3）
   - 微信小程序（uni-app 3 编译）
   - **抖音小程序**（uni-app 3 编译，规划中）
   - **快手小程序**（uni-app 3 编译，规划中）
   - 原生 App（uniappx + UTS）

### 1.5 关键技术建议（必须遵守）

1. **PostGIS 必选**：同城项目 90% 查询与地理位置相关（附近的人/店/房/工作），PostGIS 是地理查询工业标准，比 MySQL 空间索引快 5-10 倍。**禁止图省事用 MySQL 替换 PostgreSQL**。
2. **Uni-app 是多端最优解**：一套代码编译到微信/抖音/快手小程序 + H5 + App，小团队不可能维护 5 套端代码。**禁止听"原生才好"的建议改回原生开发**。
3. **内容审核必须做**：分类信息平台极易出现违规内容，初期接第三方 API（百度内容安全/阿里云绿网），后期数据量大了再自研。每个 UGC 模块（ershou/job/fang/love/quan 等）必须接入审核流程。
4. **插件化架构是"不重构"的关键**：前期多花时间把插件框架搭好，后期加模块速度会快 10 倍。**禁止把所有业务逻辑堆在单一 controller/service 包里**。
5. **不要一上来就微服务**：模块化单体足够支撑 10 万日活。微服务是有了明确的性能瓶颈和团队分工之后才拆的。**禁止过度设计**。

### 1.6 排除方案的理由（避免后期走回头路）

| 方案 | 不选原因 |
|------|----------|
| PHP + Laravel | 同步阻塞、WebSocket/长连接弱、部署复杂（php-fpm+Nginx）、后期数据量大了必须用 Go 重写核心 |
| Java + Spring Boot | 太重（Spring 全家桶）、开发效率低、服务器成本高、Java 开发者贵、小团队养不起 |
| Node.js + Express/NestJS | 单线程瓶颈（CPU 密集型任务差）、回调地狱、npm 生态混乱依赖地狱 |
| Python 做核心业务 | 解释型语言性能差，高并发场景不如 Go；只适合做 AI/数据分析辅助微服务 |

### 1.7 分阶段落地建议（避免一口吃成胖子）

| 阶段 | 周期 | 模块范围 | 端 | 目标 |
|------|------|----------|-----|------|
| 阶段1 MVP 核心 | 1-2 月 | 用户/地区/分类/头条 | 微信小程序+H5+管理后台 | 验证核心需求，跑通发布-浏览-联系闭环 |
| 阶段2 商业化 | 3-4 月 | 房产/招聘/二手/商家/会员/支付 | +PC门户+抖音小程序 | 跑通第一个变现模式，商家付费入驻 |
| 阶段3 生态完善 | 5-8 月 | 商城/拼团/活动/圈子/婚恋/零工 | +快手小程序+APP | 形成完整生态，用户粘性上来 |
| 阶段4 平台化 | 9-12 月 | 合伙人/分销/分站/开放API | 全端 | 多城市复制，成为平台 |

**当前进度**：阶段1 已完成（含部分阶段2 商家管理）；本次开发进入阶段2-3 之间，补齐对标 dismall 全部核心模块。

### 1.8 多端业务架构详细设计（4+1 业务端）

> 依据 2026-07-19 多端架构分析结果，定义各业务端的职责边界与技术载体。

#### 1.8.1 C端-用户端（消费者）

| 维度 | 说明 |
|------|------|
| 用户群体 | 普通用户、VIP用户、游客（未登录） |
| 核心场景 | 浏览同城信息、发布二手/招聘/房屋等信息、下单购物、预约服务、参加活动、IM聊天、支付、评价 |
| 技术载体 | uniappx App（Android/iOS/鸿蒙）+ 微信小程序 + H5（m.wctc.wang）+ PC门户（www.wctc.wang） |
| TabBar 设计 | 首页 / 分类 / 发布 / 消息 / 我的 |
| 关键功能 | LBS附近、全文搜索、内容发布、订单支付、消息通知、个人中心、收藏、关注、打赏 |

#### 1.8.2 B端-商家端（店铺运营）

| 维度 | 说明 |
|------|------|
| 用户群体 | 商家、店员、连锁店管理员 |
| 核心场景 | 店铺管理、商品上下架、订单处理、营销活动、数据统计、客户管理 |
| 技术载体 | uniappx App（角色切换为商家）+ H5独立入口（shop.wctc.wang） |
| TabBar 设计 | 店铺 / 商品 / 订单 / 营销 / 数据 |
| 关键功能 | 店铺装修、商品管理、订单核销、优惠券/拼团/砍价发布、营业数据、客户标签 |
| 数据隔离 | 商家只能管理自己的 ShopID 数据，店员仅限分配的店铺 |

#### 1.8.3 D端-骑手端（配送）

| 维度 | 说明 |
|------|------|
| 用户群体 | 配送员、跑腿员、全职/兼职骑手 |
| 核心场景 | 接单、配送路径规划、签到打卡、收益统计、订单状态同步 |
| 技术载体 | uniappx App（角色切换为骑手）+ H5独立入口（rider.wctc.wang） |
| TabBar 设计 | 接单 / 配送中 / 历史 / 收益 / 我的 |
| 关键功能 | 抢单/派单、地图导航、订单状态更新、配送费计算、在线/离线、收益提现 |
| 数据隔离 | 骑手只能查看分配给自己的订单，无法查看其他骑手数据 |

#### 1.8.4 M端-平台管理端（运营管理）

| 维度 | 说明 |
|------|------|
| 用户群体 | 超级管理员、平台运营、内容审核员、财务 |
| 核心场景 | 内容审核、用户管理、商家审核、订单监控、数据看板、营销配置、系统设置 |
| 技术载体 | Web管理后台（Vue 3 + Element Plus） |
| 入口路径 | admin.wctc.wang |
| 关键功能 | 顶部Tab+侧边栏布局、内容审核流、商家审核、举报处理、财务对账、数据可视化 |
| 数据范围 | 全平台数据可见（按角色权限过滤） |

#### 1.8.5 A端-代理端（城市代理，可选）

| 维度 | 说明 |
|------|------|
| 用户群体 | 城市代理商、同城合伙人 |
| 核心场景 | 分站管理、本地商家拓展、分润统计、本地数据看板 |
| 技术载体 | Web管理后台子站点（agent.wctc.wang） |
| 关键功能 | 分站配置、商家招募、订单分润、本地数据看板 |
| 数据隔离 | 代理商只能查看自己 RegionID 下的数据，无法跨地区查看 |

### 1.9 角色权限矩阵（10 角色 × 4 端权限）

> 依据用户确认（2026-07-19），采纳 10 角色完整矩阵。

#### 1.9.1 角色定义

| 角色 Code | 角色名称 | 所属业务端 | 说明 |
|-----------|----------|------------|------|
| `guest` | 游客 | C端 | 未登录用户，仅可浏览公开内容 |
| `user` | 普通用户 | C端 | 注册用户，可发布信息/下单/支付 |
| `vip` | VIP用户 | C端 | 付费会员，享会员特权/折扣 |
| `merchant` | 商家 | B端 | 入驻商家，可管理店铺/商品/订单/营销 |
| `clerk` | 店员 | B端 | 商家员工，仅限订单核销/查看 |
| `rider` | 骑手 | D端 | 配送员，接单/配送/收益 |
| `agent` | 代理商 | A端 | 城市代理，分站管理/分润 |
| `auditor` | 内容审核员 | M端 | 内容审核/举报处理 |
| `operator` | 平台运营 | M端 | 除超管外全部管理权限 |
| `super_admin` | 超级管理员 | M端 | 全平台最高权限 |

#### 1.9.2 权限矩阵（功能权限 × 角色）

| 功能权限 | guest | user | vip | merchant | clerk | rider | agent | auditor | operator | super_admin |
|----------|:-----:|:----:|:---:|:--------:|:-----:|:-----:|:-----:|:-------:|:--------:|:-----------:|
| **C端-浏览信息** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **C端-发布信息** | ❌ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ |
| **C端-下单支付** | ❌ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ |
| **C端-VIP特权** | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ |
| **B端-店铺管理** | ❌ | ❌ | ❌ | ✅(自有) | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ |
| **B端-商品管理** | ❌ | ❌ | ❌ | ✅(自有) | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ |
| **B端-订单核销** | ❌ | ❌ | ❌ | ✅(自有) | ✅(限店) | ❌ | ❌ | ❌ | ✅ | ✅ |
| **B端-营销活动** | ❌ | ❌ | ❌ | ✅(自有) | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ |
| **D端-接单配送** | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ |
| **D端-收益提现** | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ |
| **A端-分站管理** | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅(限地区) | ❌ | ✅ | ✅ |
| **A端-分润统计** | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅(限地区) | ❌ | ✅ | ✅ |
| **M端-内容审核** | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ | ✅ |
| **M端-用户管理** | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ |
| **M端-商家审核** | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅(限地区) | ❌ | ✅ | ✅ |
| **M端-系统设置** | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| **M端-财务对账** | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ |

#### 1.9.3 权限码命名规范

权限码采用 `模块:动作` 格式，例如：
- `ershou:read` - 浏览二手
- `ershou:publish` - 发布二手
- `shop:manage` - 管理店铺（B端）
- `shop:audit` - 审核商家（M端）
- `order:verify` - 核销订单（B端店员）
- `rider:deliver` - 配送订单（D端）
- `agent:station` - 分站管理（A端）
- `content:audit` - 内容审核（M端）
- `system:config` - 系统设置（M端超管）

### 1.10 数据隔离规则

> 所有业务表必须遵守以下数据隔离规则，确保多角色多端数据安全。

#### 1.10.1 四维数据隔离模型

| 隔离维度 | 字段 | 适用表 | 隔离规则 |
|----------|------|--------|----------|
| **地区隔离** | `region_id` | 所有业务表（继承 RegionBaseModel） | 代理商只能查看自己 RegionID 下的数据 |
| **商家隔离** | `shop_id` | 商品/订单/营销活动表 | 商家只能管理自己的 ShopID 数据 |
| **用户隔离** | `user_id` | 发布/订单/收藏/消息表 | 用户只能查看/操作自己的 UserID 数据 |
| **骑手隔离** | `rider_id` | 配送订单/骑手收益表 | 骑手只能查看分配给自己的订单 |

#### 1.10.2 后端实现规范

```go
// Repository 层统一注入隔离条件
type ErshouRepo struct {
    db *gorm.DB
}

func (r *ErshouRepo) List(ctx context.Context, query ErshouQuery) ([]*Ershou, int64, error) {
    db := r.db.WithContext(ctx).Model(&Ershou{})

    // 地区隔离（代理商/平台运营/超管可跨地区）
    if query.RegionID > 0 && !ctx.HasPermission("region:cross") {
        db = db.Where("region_id = ?", query.RegionID)
    }

    // 用户隔离（只能看自己的发布）
    if query.UserID > 0 && query.OnlyMine {
        db = db.Where("user_id = ?", query.UserID)
    }

    // 商家隔离（只能管理自己店铺的商品）
    if query.ShopID > 0 && query.OnlyMyShop {
        db = db.Where("shop_id = ?", query.ShopID)
    }

    // 骑手隔离（只能看分配给自己的订单）
    if query.RiderID > 0 && query.OnlyMyRider {
        db = db.Where("rider_id = ?", query.RiderID)
    }

    return r.find(db, query.Page, query.PageSize)
}
```

#### 1.10.3 前端路由守卫规范

- **管理后台（Vue3）**：路由 `meta.permission` 校验，无权限跳转 403
- **uniappx App**：登录后根据角色动态生成 TabBar 与路由表，未授权页面不渲染
- **PC门户（Next.js）**：仅 C 端浏览权限，无需复杂权限控制

---

## 二、模块清单（共 48 个模块）

### 2.1 已开发模块（9 个）✅

| 编号 | 模块名 | 中文名 | 路由前缀 | 状态 |
|------|--------|--------|----------|------|
| 1 | user | 用户中心 | /api/v1/user | ✅ 完成 |
| 2 | region | 地区管理 | /api/v1/region | ✅ 完成 |
| 3 | permission | 权限管理 | /api/v1/permission | ✅ 完成 |
| 4 | file | 文件存储 | /api/v1/file | ✅ 完成 |
| 5 | setting | 系统设置 | /api/v1/setting | ✅ 完成 |
| 6 | category | 分类管理 | /api/v1/category | ✅ 完成 |
| 7 | news | 同城头条 | /api/v1/news | ✅ 完成（含评论/点赞/收藏/消息） |
| 8 | shop | 商家管理 | /api/v1/shop | ✅ 后端完成，前端管理后台完成 |
| 9 | groupbuy | 团购 | /api/v1/groupbuy | ✅ 后端完成，前端待做 |

### 2.2 待开发模块（39 个）❌

#### A. 同城基础业务（8 个）- **第一批开发**

| 编号 | 模块名 | 中文名 | 路由前缀 | 核心功能 |
|------|--------|--------|----------|----------|
| 10 | ershou | 同城二手 | /api/v1/ershou | 商品发布/分类/搜索/留言/交易 |
| 11 | job | 同城招聘 | /api/v1/job | 职位发布/简历投递/职位搜索/企业认证 |
| 12 | fang | 房屋租售 | /api/v1/fang | 房源发布/小区/户型/地图找房/中介 |
| 13 | pinche | 同城拼车 | /api/v1/pinche | 行程发布/路线匹配/顺风车/包车 |
| 14 | love | 相亲交友 | /api/v1/love | 个人资料/相册/心动/匹配/聊天 |
| 15 | car | 车辆买卖 | /api/v1/car | 车辆发布/品牌/车系/年份/估价 |
| 16 | linggong | 零工兼职 | /api/v1/linggong | 任务发布/接单/结算/评价 |
| 17 | dh114 | 同城114 | /api/v1/dh114 | 黄页录入/分类/搜索/电话直达 |

#### B. 商家服务扩展（6 个）- **第二批开发**

| 编号 | 模块名 | 中文名 | 路由前缀 | 核心功能 |
|------|--------|--------|----------|----------|
| 18 | mall | 同城商城 | /api/v1/mall | 商品/SKU/购物车/订单/支付 |
| 19 | yuyue | 同城预约 | /api/v1/yuyue | 服务预约/时段/到店/上门 |
| 20 | daojia | 同城到家 | /api/v1/daojia | 服务下单/师傅接单/上门/结算 |
| 21 | diancan | 同城点餐 | /api/v1/diancan | 菜单/堂食/外卖/桌台/结算 |
| 22 | huodong | 同城活动 | /api/v1/huodong | 活动发布/报名/签到/抽奖 |
| 23 | quan | 同城圈子 | /api/v1/quan | 圈子/帖子/回复/置顶/加精 |

#### C. 营销活动（5 个）- **第三批开发**

| 编号 | 模块名 | 中文名 | 路由前缀 | 核心功能 |
|------|--------|--------|----------|----------|
| 24 | kanjia | 同城砍价 | /api/v1/kanjia | 砍价商品/帮砍/底价/下单 |
| 25 | pintuan | 同城拼团 | /api/v1/pintuan | 拼团商品/开团/参团/成团 |
| 26 | choujiang | 同城抽奖 | /api/v1/choujiang | 抽奖活动/奖品/中奖/兑换 |
| 27 | coupon | 优惠券系统 | /api/v1/coupon | 券模板/领取/核销/统计 |
| 28 | sign | 同城签到 | /api/v1/sign | 每日签到/积分/连续奖励/兑换 |

#### D. 用户运营（5 个）- **第四批开发**

| 编号 | 模块名 | 中文名 | 路由前缀 | 核心功能 |
|------|--------|--------|----------|----------|
| 29 | renzheng | 用户认证 | /api/v1/renzheng | 实名/企业/个人/资质审核 |
| 30 | vipcard | 会员卡 | /api/v1/vipcard | 会员等级/权益/黑卡/折扣 |
| 31 | partner | 同城合伙人 | /api/v1/partner | 分销/佣金/等级/提现 |
| 32 | majia | 同城马甲 | /api/v1/majia | 多账号/匿名发帖/管理 |
| 33 | jubao | 举报中心 | /api/v1/jubao | 举报/处理/封禁/申诉 |

#### E. 社区互动（6 个）- **第五批开发**

| 编号 | 模块名 | 中文名 | 路由前缀 | 核心功能 |
|------|--------|--------|----------|----------|
| 34 | dashang | 同城打赏 | /api/v1/dashang | 打赏/答谢/积分/记录 |
| 35 | mingpian | 同城名片 | /api/v1/mingpian | 名片/交换/收藏/分享 |
| 36 | zhibo | 同城直播 | /api/v1/zhibo | 开播/观看/弹幕/礼物 |
| 37 | ai | 同城AI助手 | /api/v1/ai | 智能问答/推荐/写作助手 |
| 38 | tuiwen | 推文助手 | /api/v1/tuiwen | 推文模板/一键生成/发布 |
| 39 | share | 一键分享 | /api/v1/share | 多端分享/海报生成/统计 |

#### F. 垂直行业（5 个）- **第六批开发（暂不在首批30个内）**

| 编号 | 模块名 | 中文名 | 路由前缀 | 核心功能 |
|------|--------|--------|----------|----------|
| 40 | edu | 教育培训 | /api/v1/edu | 机构/课程/报名/排课 |
| 41 | zhuangxiu | 同城装修 | /api/v1/zhuangxiu | 案例/设计师/施工队/报价 |
| 42 | qiche | 同城汽车 | /api/v1/qiche | 4S店/维修/保养/救援 |
| 43 | wxqun | 同城微信群 | /api/v1/wxqun | 群发布/分类/搜索/二维码 |
| 44 | dm | 同城DM | /api/v1/dm | 广告/DM单/投放/统计 |

#### G. 基础服务（4 个）- **第七批开发（暂不在首批30个内）**

| 编号 | 模块名 | 中文名 | 路由前缀 | 核心功能 |
|------|--------|--------|----------|----------|
| 45 | pay | 支付中心 | /api/v1/pay | 微信/支付宝/余额/退款 |
| 46 | subsite | 同城子站点 | /api/v1/subsite | 子站/分站/独立域名 |
| 47 | qrcode | 渠道二维码 | /api/v1/qrcode | 渠道码/统计/带参二维码 |
| 48 | tcmessage | 同城消息（扩展） | /api/v1/tcmessage | 站内信/系统通知/模板消息 |

---

## 三、首批开发范围（30 个模块）

> 用户确认首批开发：A+B+C+D+E = 8+6+5+5+6 = 30 个模块

### 3.1 首批模块清单

**A. 同城基础业务（8 个）**：ershou、job、fang、pinche、love、car、linggong、dh114
**B. 商家服务扩展（6 个）**：mall、yuyue、daojia、diancan、huodong、quan
**C. 营销活动（5 个）**：kanjia、pintuan、choujiang、coupon、sign
**D. 用户运营（5 个）**：renzheng、vipcard、partner、majia、jubao
**E. 社区互动（6 个）**：dashang、mingpian、zhibo、ai、tuiwen、share

### 3.2 第二批预留（9 个，暂不开发）

**F. 垂直行业（5 个）**：edu、zhuangxiu、qiche、wxqun、dm
**G. 基础服务（4 个）**：pay、subsite、qrcode、tcmessage

---

## 四、三级菜单结构设计

### 4.1 后台菜单重构方案

> 当前 `frontend/src/router/index.js` 是扁平二级菜单，需重构为**三级嵌套**。

```
管理后台
├── 仪表盘 (dashboard)
├── 内容管理
│   ├── 同城头条 (news)
│   │   ├── 头条列表
│   │   ├── 我的收藏
│   │   └── 消息中心
│   └── 分类管理 (category)
├── 同城业务
│   ├── 二手交易 (ershou)
│   ├── 招聘求职 (job)
│   ├── 房屋租售 (fang)
│   ├── 拼车出行 (pinche)
│   ├── 相亲交友 (love)
│   ├── 车辆买卖 (car)
│   ├── 零工兼职 (linggong)
│   └── 同城114 (dh114)
├── 商家服务
│   ├── 商家管理 (shop)
│   │   ├── 商家列表
│   │   └── 评价审核
│   ├── 同城商城 (mall)
│   ├── 预约服务 (yuyue)
│   ├── 同城到家 (daojia)
│   ├── 同城点餐 (diancan)
│   ├── 同城活动 (huodong)
│   └── 同城圈子 (quan)
├── 营销活动
│   ├── 团购管理 (groupbuy)
│   │   ├── 团购商品
│   │   ├── 订单管理
│   │   └── 优惠券
│   ├── 砍价活动 (kanjia)
│   ├── 拼团活动 (pintuan)
│   ├── 抽奖活动 (choujiang)
│   ├── 优惠券系统 (coupon)
│   └── 签到积分 (sign)
├── 用户运营
│   ├── 用户管理 (user)
│   ├── 用户认证 (renzheng)
│   ├── 会员卡 (vipcard)
│   ├── 同城合伙人 (partner)
│   ├── 马甲管理 (majia)
│   └── 举报中心 (jubao)
├── 社区互动
│   ├── 打赏记录 (dashang)
│   ├── 名片管理 (mingpian)
│   ├── 直播管理 (zhibo)
│   ├── AI助手 (ai)
│   ├── 推文助手 (tuiwen)
│   └── 分享统计 (share)
└── 系统设置
    ├── 地区管理 (region)
    ├── 权限管理 (permission)
    │   ├── 角色管理
    │   └── 权限列表
    ├── 系统设置 (setting)
    └── 文件管理 (file)
```

### 4.2 菜单技术实现

- 使用 `el-menu` 的 `el-sub-menu` 嵌套（支持无限级）
- 路由 `meta` 增加 `menuGroup` 字段标识所属分组
- 路由 `meta` 增加 `menuLevel` 字段标识层级
- 保留现有 `meta.permission` 权限码校验

---

## 五、开发顺序与里程碑

### 5.1 阶段0：基础重构（必须先做）

| 任务 | 说明 |
|------|------|
| 0.1 后台菜单三级嵌套重构 | 重构 `frontend/src/router/index.js` + `MainLayout.vue` |
| 0.2 创建 uniappx App 工程 | `frontend/app/` 初始化 uniappx 项目 |
| 0.3 公共组件库抽取 | 表单/表格/筛选器/详情弹窗等通用组件 |

### 5.2 阶段1：同城基础业务（A 组 8 个模块）

每个模块按**四端同步**开发：
- 后端：model + dto + repository + service + handler + plugin.go
- 管理后台：api/xx.js + views/xx/index.vue + 路由
- PC门户：app/xx/page.tsx + lib/api.ts 补充
- H5小程序：pages/xx/ + api/xx.js
- App（uniappx）：pages/xx/ + api/xx.uts

**模块开发顺序**：ershou → job → fang → pinche → love → car → linggong → dh114

### 5.3 阶段2：商家服务扩展（B 组 6 个模块）

mall → yuyue → daojia → diancan → huodong → quan

### 5.4 阶段3：营销活动（C 组 5 个模块）

kanjia → pintuan → choujiang → coupon → sign

### 5.5 阶段4：用户运营（D 组 5 个模块）

renzheng → vipcard → partner → majia → jubao

### 5.6 阶段5：社区互动（E 组 6 个模块）

dashang → mingpian → zhibo → ai → tuiwen → share

---

## 六、技术规范

### 6.1 后端开发规范

1. **插件化架构**：每个模块独立目录 `backend/internal/modules/{module}/`
2. **目录结构**：`model/` + `dto/` + `repository/` + `service/` + `handler/` + `plugin.go`
3. **数据模型**：所有业务表继承 `RegionBaseModel`（含 RegionID + CreatedAt + UpdatedAt + DeletedAt）
4. **API规范**：统一返回 `{code, message, data}`，code=0 成功
5. **路由前缀**：`/api/v1/{module}/`
6. **权限码**：`{module}:read` / `{module}:create` / `{module}:update` / `{module}:delete` / `{module}:audit`
7. **缓存**：cache-aside 模式，Redis 不可用降级走 DB
8. **测试**：每个模块配套单元测试（service + handler）

### 6.2 管理后台开发规范

1. **API封装**：`src/api/{module}.js`
2. **页面**：`src/views/{module}/index.vue`（多页面时拆分多个 .vue）
3. **路由**：在 `src/router/index.js` 按三级菜单结构注册
4. **权限**：`meta.permission` 控制访问，`v-permission` 控制按钮
5. **UI组件**：Element Plus，遵循现有风格（page-card + toolbar + el-table + el-pagination + el-dialog）
6. **格式化**：复用 `@/utils/format.js` 的 formatTime/statusText 等

### 6.3 PC门户开发规范

1. **页面**：`src/app/{module}/page.tsx` + `[id]/page.tsx`
2. **API**：`src/lib/api.ts` 集中封装
3. **SSR**：默认 SSR + try/catch 容错降级
4. **样式**：Tailwind CSS
5. **组件**：复用 `src/components/` 现有组件

### 6.4 H5小程序开发规范

1. **页面**：`src/pages/{module}/index.vue` + `detail.vue`
2. **API**：`src/api/{module}.js`
3. **路由**：在 `src/pages.json` 注册
4. **样式**：uni-app 内置样式 + rpx

### 6.5 App（uniappx）开发规范

1. **页面**：`pages/{module}/index.uvue` + `detail.uvue`
2. **API**：`api/{module}.uts`
3. **路由**：在 `pages.json` 注册
4. **样式**：uvue 内置样式 + rpx
5. **语言**：UTS（TypeScript 风格）

---

## 七、数据库设计原则

### 7.1 通用原则

1. 所有业务表必须包含：`id` (主键) + `region_id` (地区隔离) + `created_at` + `updated_at` + `deleted_at`
2. 用户相关表必须包含 `user_id` 外键
3. 状态字段统一：`status` (0禁用/1正常) + `audit_status` (0待审/1通过/2拒绝)
4. 软删除：GORM 默认软删除
5. 索引：`region_id` + `user_id` + `category_id` + `status` + `created_at` 组合索引

### 7.2 模块表命名规范

- 主表：`{module}s`（如 `ershous`、`jobs`、`fangs`）
- 子表：`{module}_{sub}`（如 `ershou_images`、`job_resumes`）
- 关联表：`{module}_{rel}`（如 `ershou_favs`、`job_applies`）

---

## 八、验收标准

### 8.1 每个模块的验收标准

1. **后端**：CRUD + 列表分页 + 筛选 + 权限码 + 单元测试通过
2. **管理后台**：列表页 + 筛选 + 新建/编辑/删除 + 详情 + 审核（如需）+ 路由菜单
3. **PC门户**：列表页 + 详情页 + API 对接
4. **H5小程序**：列表页 + 详情页 + API 对接
5. **App**：列表页 + 详情页 + API 对接

### 8.2 整体验收标准

1. 5 端全部能正常访问对应模块
2. 后端 API 文档（Swagger）完整
3. 三级菜单结构正常显示
4. 权限控制生效（无权限看不到菜单和按钮）
5. 数据库表结构自动迁移
6. 种子数据初始化（权限码 + 默认配置）

---

## 九、风险与依赖

### 9.1 技术风险

1. **uniappx 学习成本**：UTS 语言与 Vue 有差异，需要适配期
2. **5端同步开发工作量**：30 模块 × 5 端 = 150 页面，需分批迭代
3. **支付中心缺失**：营销活动模块依赖支付，需提前规划（暂用 mock）

### 9.2 外部依赖

1. **微信支付**：mall/coupon/groupbuy 等需要，暂未接入
2. **微信登录**：已接入（v1.7.0）
3. **短信服务**：已接入阿里云（v1.3.0）
4. **对象存储**：已接入 MinIO/七牛云/阿里云 OSS STS

---

## 十、文档变更记录

| 日期 | 变更内容 | 操作人 |
|------|----------|--------|
| 2026-07-19 | 初始版本，对标 dismall 完整功能清单 | AI 助手 |

---

## 附录 A：对标清单详尽映射

### 西瓜先生核心模块映射

| 西瓜模块 | 本项目模块 | 状态 |
|----------|-----------|------|
| 【西瓜】分类信息 | news + category | ✅ 已有 |
| 【西瓜】同城头条 | news | ✅ 已有 |
| 【西瓜】同城后台 | frontend 管理后台 | ✅ 已有 |
| 【西瓜】同城商圈 | shop | ✅ 已有 |
| 【西瓜】同城优惠 | groupbuy + coupon | ✅ 部分已有 |
| 【西瓜】同城商城 | mall | ❌ 待开发 |
| 【西瓜】同城合伙人 | partner | ❌ 待开发 |
| 【西瓜】同城招聘 | job | ❌ 待开发 |
| 【西瓜】同城二手 | ershou | ❌ 待开发 |
| 【西瓜】同城预约 | yuyue | ❌ 待开发 |
| 【西瓜】同城点餐 | diancan | ❌ 待开发 |
| 【西瓜】同城活动 | huodong | ❌ 待开发 |
| 【西瓜】同城砍价 | kanjia | ❌ 待开发 |
| 【西瓜】同城黑卡 | vipcard | ❌ 待开发 |
| 【西瓜】同城圈子 | quan | ❌ 待开发 |
| 【西瓜】同城到家 | daojia | ❌ 待开发 |
| 【西瓜】同城签到 | sign | ❌ 待开发 |
| 【西瓜】相亲交友 | love | ❌ 待开发 |
| 【西瓜】举报中心 | jubao | ❌ 待开发 |
| 【西瓜】同城认证 | renzheng | ❌ 待开发 |
| 【西瓜】同城马甲 | majia | ❌ 待开发 |
| 【西瓜】同城名片 | mingpian | ❌ 待开发 |
| 【西瓜】同城打赏 | dashang | ❌ 待开发 |
| 【西瓜】同城AI | ai | ❌ 待开发 |
| 【西瓜】同城分站 | subsite | ❌ 第二批 |
| 【西瓜】同城114 | dh114 | ❌ 待开发 |
| 【西瓜】零工驿站 | linggong | ❌ 待开发 |
| 【西瓜】同城DM | dm | ❌ 第二批 |
| 【西瓜】同城直播 | zhibo | ❌ 待开发 |
| 【西瓜】微信登录 | user (oauth) | ✅ 已有 |

### 点微科技核心模块映射

| 点微模块 | 本项目模块 | 状态 |
|----------|-----------|------|
| [点微]同城分类信息 | news + category | ✅ 已有 |
| [点微]同城头条 | news | ✅ 已有 |
| [点微]同城好店 | shop | ✅ 已有 |
| [点微]同城拼团 | pintuan | ❌ 待开发 |
| [点微]同城优惠抢购 | groupbuy | ✅ 已有 |
| [点微]同城商城 | mall | ❌ 待开发 |
| [点微]同城合伙人 | partner | ❌ 待开发 |
| [点微]同城CRM | partner 扩展 | ❌ 待开发 |
| [点微]同城预约 | yuyue | ❌ 待开发 |
| [点微]同城砍价 | kanjia | ❌ 待开发 |
| [点微]同城零工 | linggong | ❌ 待开发 |
| [点微]同城招聘 | job | ❌ 待开发 |
| [点微]同城二手交易 | ershou | ❌ 待开发 |
| [点微]同城抽奖 | choujiang | ❌ 待开发 |
| [点微]同城签到 | sign | ❌ 待开发 |
| [点微]同城认证 | renzheng | ❌ 待开发 |
| [点微]同城会员卡 | vipcard | ❌ 待开发 |
| [点微]用户中心 | user | ✅ 已有 |
| [点微]同城直播 | zhibo | ❌ 待开发 |
| [点微]同城任务 | linggong 扩展 | ❌ 待开发 |
| [点微]同城教育培训 | edu | ❌ 第二批 |
| [点微]同城装修 | zhuangxiu | ❌ 第二批 |
| [点微]支付中心 | pay | ❌ 第二批 |
| [点微]同城消息 | tcmessage | ❌ 第二批 |
| [点微]同城拼车 | pinche | ❌ 待开发 |
| [点微]同城汽车 | qiche | ❌ 第二批 |
| [点微]同城婚恋 | love | ❌ 待开发 |
| [点微]同城圈子 | quan | ❌ 待开发 |
| [点微]同城AI助手 | ai | ❌ 待开发 |
| [点微]同城子站点 | subsite | ❌ 第二批 |
| [点微]渠道二维码 | qrcode | ❌ 第二批 |
| [点微]同城推文助手 | tuiwen | ❌ 待开发 |

