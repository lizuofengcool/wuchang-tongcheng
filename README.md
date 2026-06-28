# 五常同城本地生活服务平台

五常同城是一个面向五常市的本地生活服务平台，提供分类信息、同城头条、商家服务等功能。

## 技术栈

### 后端
- **语言**: Go 1.22+
- **Web框架**: Gin
- **ORM**: GORM
- **架构模式**: 插件化架构 + Repository模式
- **数据库**: PostgreSQL 16（PostGIS 扩展已部署但代码未使用空间查询）
- **缓存**: Redis 7（已封装，限流 + 业务缓存接入：region/category 树 30min TTL、news 列表 60s TTL，写操作按前缀失效，Redis 不可用降级走 DB）
- **搜索引擎**: Elasticsearch 8（已集成：news 全文检索 multi_match + 异步索引，ES 不可用降级 DB LIKE）
- **消息队列**: RabbitMQ（已集成：news 写入异步索引解耦，topic 交换机发布订阅，手动 ack）
- **实时通信**: WebSocket（已实现：Hub 连接管理 + JWT 鉴权升级端点 /ws，单用户多连接定向推送 + 全局广播，点赞实时通知作者）
- **对象存储**: 已实现 LocalStorage + MinIO（S3 协议兼容，可适配 AWS S3/阿里云 OSS/腾讯云 COS）；七牛云Kodo 待补齐
- **地图服务**: 高德地图API（已实现：地理编码/逆地理编码/周边 POI 搜索，key 未配置降级返回 503）
- **鉴权**: JWT + RBAC（用户-角色-权限，超级管理员直通）
- **API文档**: Swagger（gin-swagger + swaggo/swag，已集成）
- **限流防刷**: 基于 Redis INCR 的固定窗口限流（登录 5/min、新闻读取 60/min、点赞 30/min），Redis 不可用时优雅降级
- **CI/CD**: GitHub Actions（backend go vet/build/test、frontend npm build、tag 触发 docker publish 推送 GHCR）
- **测试**: 单元测试（标准库 testing，无外部依赖）+ 集成测试（testcontainers-go + testify，自动拉起 PostgreSQL 容器，无 Docker 时优雅 SKIP）

### 前端
- **管理后台**: Vue 3 + Vite + Element Plus + Pinia（当前已实现）
- **PC门户**: Next.js 14 App Router + TypeScript + Tailwind CSS（已实现：首页 ISR、头条列表/详情、分类页、搜索、点赞组件，SSR try/catch 容错降级）
- **小程序**: Uni-app 3 + Vue 3 + Vite（已实现：首页/头条列表/详情/搜索/我的 5 页 + tabBar，H5/微信小程序多端编译）

## 项目结构

```
wuchang-tongcheng/
├── backend/                    # 后端Go项目
│   ├── cmd/                    # 应用入口
│   │   └── server/             # HTTP服务入口
│   ├── internal/               # 内部代码
│   │   ├── core/               # 核心框架
│   │   │   ├── plugin/         # 插件系统
│   │   │   ├── router/         # 路由封装
│   │   │   ├── middleware/     # 中间件（auth/cors/logger/permission/region/recovery）
│   │   │   └── response/       # 统一响应
│   │   ├── pkg/                # 公共包
│   │   │   ├── config/         # 配置管理
│   │   │   ├── database/       # 数据库封装
│   │   │   ├── redis/          # Redis封装
│   │   │   ├── jwt/            # JWT 鉴权
│   │   │   ├── logger/         # 日志封装
│   │   │   ├── storage/        # 文件存储（已实现 local，minio/qiniu 待补齐）
│   │   │   ├── seed/           # 种子数据（地区/权限/admin）
│   │   │   └── utils/          # 工具函数（分页/错误码/helper）
│   │   └── modules/            # 业务模块（插件，每个含 model/dto/repository/service/handler/plugin.go）
│   │       ├── user/           # 用户模块
│   │       ├── region/         # 地区模块
│   │       ├── permission/     # 权限模块（RBAC）
│   │       ├── file/           # 文件存储模块
│   │       ├── setting/        # 系统设置模块
│   │       ├── category/       # 分类信息模块
│   │       └── news/           # 同城头条模块
│   ├── configs/                # 配置文件
│   ├── Dockerfile               # 后端镜像构建
│   └── Makefile                # 构建脚本
├── frontend/                   # 前端工程（三端）
│   ├── src/                    # 管理后台（Vue 3 + Vite + Element Plus）
│   │   ├── api/                # 接口封装
│   │   ├── components/         # 公共组件（RichTextEditor）
│   │   ├── directives/         # 自定义指令（v-permission/v-role）
│   │   ├── layouts/            # 布局（MainLayout）
│   │   ├── router/             # 路由 + 守卫
│   │   ├── stores/             # Pinia（user/region）
│   │   ├── utils/              # 工具（request/auth/format）
│   │   └── views/              # 页面（login/dashboard/profile/error/news/...）
│   ├── deploy/nginx.conf       # Nginx 配置
│   ├── Dockerfile              # 管理后台镜像构建
│   ├── .env.development / .env.production
│   ├── pc/                     # PC门户站（Next.js 14 App Router + TS + Tailwind）
│   │   ├── src/app/            # 路由：首页/头条列表/详情/分类/搜索
│   │   ├── src/components/     # Header/Footer/NewsCard/RegionSelector
│   │   └── src/lib/            # api/region/types
│   └── miniapp/                # 小程序端（Uni-app 3 + Vue 3 + Vite）
│       └── src/                # pages（index/news/search/user）+ api + manifest
└── deploy/                     # 整体部署
    └── docker-compose.yml      # Docker Compose 配置（含 PG/Redis/RabbitMQ/ES/MinIO）
```

> 注：`backend/scripts/`（数据库迁移脚本）目录尚未建立，PostGIS 空间查询代码未接入。

## 快速开始

### 环境要求
- Go 1.22+
- Node.js 20+
- Docker & Docker Compose
- PostgreSQL 16+
- Redis 7+

### 1. 启动基础设施

```bash
cd deploy
docker-compose up -d
```

### 2. 配置文件

```bash
cd backend/configs
cp config.yaml.example config.yaml
# 修改 config.yaml 中的配置
```

### 3. 运行服务

```bash
cd backend
go run cmd/server/main.go
```

或者使用Makefile：

```bash
make run
```

### 4. 验证服务

访问 http://localhost:8080/health 检查服务状态。

## 开发规范

### 数据库规范
1. 所有业务表必须携带 `region_id` 字段实现地区数据隔离
2. 所有表必须包含 `created_at`、`updated_at`、`deleted_at` 字段（GORM软删除）
3. 使用 `RegionBaseModel` 作为业务表的基类

### API规范
1. 统一返回格式：`{code, message, data}`
2. code=0 表示成功，非0表示失败
3. 使用统一的响应封装 `response.Success()`、`response.Fail()`

### 插件开发规范
1. 每个业务模块都是独立的Go插件
2. 实现 `plugin.Plugin` 接口
3. 通过 `plugin.GetManager().Register()` 注册插件
4. 路由自动注册到 `/api/v1/{plugin_name}/` 路径下

## 核心模块

### 插件系统
- 统一的插件接口定义
- 插件生命周期管理（Init/RegisterRoutes/Close）
- 插件路由自动注册
- 插件依赖管理

### 地区数据隔离
- 所有业务数据按地区隔离
- 通过中间件自动注入 region_id
- 支持多地区部署

### 统一响应
- 标准的 API 响应格式
- 统一的错误码定义
- 分页结果封装

## 部署

### Docker Compose一键部署

```bash
cd deploy
docker-compose up -d
```

包含服务：
- PostgreSQL 16 + PostGIS
- Redis 7
- RabbitMQ 3.12（带管理界面）
- Elasticsearch 8
- Kibana（可选）
- MinIO（可选，开发环境替代七牛云）
- pgAdmin（可选）

## 版本历史

- v0.1.0 - 初始版本，核心框架搭建完成
- v0.2.0 - 补齐 RBAC 全链路、地区隔离前端落地、富文本编辑器、Docker 化部署
  - 后端：CORS 修复、WrapGin 中间件桥、seed 种子数据、file/permission 模块补齐、my-auth 端点
  - 前端：v-permission/v-role 指令、地区选择器、403/500 错误页、.env、news 富文本+封面上传、role 权限回显、permission 编辑
  - 工程：前后端多阶段 Dockerfile + Nginx 反代配置
- v0.3.0 - 基础设施与防护层补齐（D3-D9）
  - D3 MinIO 对象存储（S3 协议，自动建桶+公开读+按日期分目录）
  - D4 news 点赞 API（幂等 toggle，NewsLike 唯一索引）+ 详情页
  - D5 地区隔离全链路（file.List + user 读写 + X-Region-ID）
  - D6 setting 值类型反序列化 + category/region 层级深度限制（MaxLevel=3）
  - D7 Redis 限流中间件（登录/读取/点赞分级，降级容错）
  - D8 后端单元测试（28 用例，覆盖 utils/setting/user 纯函数）
  - D9 GitHub Actions CI/CD（backend/frontend CI + tag 触发 docker-publish GHCR）
- v0.4.0 - 异步索引与全文检索（D10）
  - RabbitMQ 封装（topic 交换机+手动 ack，连接关闭自愈）
  - Elasticsearch 封装（esapi 函数式，IndexDoc/DeleteDoc/SearchByQuery/CreateIndexIfNotExists）
  - indexer 三态工厂（NoopIndexer/MQIndexer/DirectESIndexer，按可用性自动选择）
  - news 写入异步索引（fire-and-forget）+ Search 全文检索（multi_match 4 字段加权，ES 不可用降级 DB LIKE）
- v0.5.0 - 三端前端落地（D11-D12）
  - D11 PC门户站 Next.js 14（ISR 首页、头条列表/详情、分类、搜索、点赞组件，SSR 容错降级，多阶段 Dockerfile）
  - D12 小程序 Uni-app 3（首页/头条列表/详情/搜索/我的 5 页 + tabBar，H5/微信小程序多端编译）
- v0.6.0 - Redis 业务缓存（D14）
  - cache-aside 助手（GetJSON/SetJSON/DelByPrefix，Redis 不可用降级 miss）
  - region/category 树缓存（30min TTL，写操作 SCAN+DEL 按前缀失效）
  - news 列表缓存（60s TTL，仅 keyword 为空的热点 feed）
- v0.7.0 - WebSocket 实时通知（D15）
  - ws Hub 连接管理（单用户多连接、注册/注销事件循环、定向推送/广播）
  - /ws 升级端点（JWT query token 鉴权 + 读写泵 + ping 保活）
  - 点赞实时通知作者（fire-and-forget，不在线丢弃，不通知自己）
- v0.8.0 - 数据库初始化与工程脚本（D16）
  - deploy/initdb/01-extensions.sql（PostGIS 扩展，docker-compose 挂载点补齐）
  - Makefile migrate 目标（AutoMigrate + seed 说明）、swagger 目标（swag init）
- v0.9.0 - 高德地图 API 集成（D17）
  - amap 客户端（Regeocode/Geocode/Around，标准库 net/http，无新依赖）
  - key 未配置降级（占位值/空值/非 amap 类型均不激活，返回 503）
  - /api/v1/map/{regeocode,geocode,around} 路由（需登录 + 限流 30/min）
- v1.0.0 - 后端集成测试（D18）
  - testcontainers-go + testify 测试栈，PG 16-alpine 容器自动拉起
  - pgtest 夹具包（SetupPostgres + MigrateAll 全量建表，无 Docker 自动 SKIP）
  - user/region/news 三模块 repository 集成测试（CRUD/唯一索引/分页/过滤/软删除/点赞幂等流程）
  - Makefile test-integration / test-unit 目标（WCTC_SKIP_INTEGRATION=1 跳过开关）
- v1.1.0 - HTTP 端到端集成测试（D19）
  - user 模块 HTTP e2e（gin + httptest，真实 PG 容器）
  - 覆盖全链路：全局中间件（Region+Auth）→ 路由 → handler → service → repository → DB
  - 15 用例：注册/登录/鉴权/资料更新/改密码/admin 权限（超管直通 + 普通用户 403）
  - seed.Run 复用 + 真实 permission service 注入权限校验器

## 功能完成度（对照规划）

> 本节用于诚实标注当前实际进度，避免与规划混淆。

### 已完成
- ✅ 插件化后端骨架（7 个业务模块均含 model/dto/repository/service/handler/plugin.go）
- ✅ RBAC 权限模型（用户-角色-权限，超级管理员直通，路由级权限校验）
- ✅ JWT 鉴权 + AuthRequired/RequirePermission 中间件全链路打通
- ✅ 地区数据隔离（中间件 + RegionBaseModel，全链路：news/category/setting 读取 + file.List + user 读写）
- ✅ 种子数据（31 个权限码、5 个地区、admin 超管账号）
- ✅ Vue3 管理后台（login/dashboard/profile + 7 个业务管理页 + news 详情页）
- ✅ 前端权限指令 v-permission/v-role、路由守卫 meta.permission
- ✅ 富文本编辑器组件（contenteditable + 图片上传）
- ✅ 前后端 Docker 多阶段构建 + Nginx 反代 + .env 配置
- ✅ Swagger API 文档（gin-swagger + swaggo/swag，路由 /swagger/index.html）
- ✅ MinIO 对象存储（S3 协议兼容，可适配 AWS S3/阿里云 OSS/腾讯云 COS；自动建桶 + 公开读策略 + 按日期分目录）
- ✅ news 点赞 API（幂等 toggle，NewsLike 表 user_id+news_id 唯一索引）+ 前端详情页
- ✅ setting 值类型反序列化（string/number/bool/json 四类型，写入校验 + 读取解析）
- ✅ category/region 层级深度限制（MaxLevel=3，Level 按 ParentID 自动计算）
- ✅ 限流防刷（基于 Redis INCR 固定窗口，登录 5/min、news 读取 60/min、点赞 30/min，Redis 不可用优雅降级）
- ✅ 后端单元测试（utils/setting/user 共 28 个用例，覆盖纯函数无 DB/Redis 依赖）
- ✅ 后端集成测试（testcontainers-go + testify，pgtest 夹具自动拉起 PG 容器，user/region/news 三模块 repository 端到端验证 CRUD/唯一索引/分页/过滤/软删除/点赞幂等；无 Docker 自动 SKIP，WCTC_SKIP_INTEGRATION=1 强制跳过）
- ✅ HTTP 端到端集成测试（gin + httptest，user 模块 15 用例覆盖注册/登录/JWT 鉴权/资料更新/改密码/admin 权限全链路，seed 复用 + 真实 permission service 注入）
- ✅ GitHub Actions CI/CD（backend CI、frontend CI、docker-publish 推送 GHCR）
- ✅ RabbitMQ 集成（topic 交换机发布订阅 + 手动 ack + 连接自愈，news 异步索引解耦）
- ✅ Elasticsearch 集成（esapi 函数式封装，IndexDoc/DeleteDoc/SearchByQuery/CreateIndexIfNotExists，news 全文检索 multi_match + 降级 DB LIKE）
- ✅ indexer 三态工厂（NoopIndexer/MQIndexer/DirectESIndexer，按 MQ/ES 可用性自动选择）
- ✅ PC门户站 Next.js 14（首页 ISR 60s、头条列表/详情、分类页、搜索、点赞组件，SSR try/catch 容错降级）
- ✅ 小程序 Uni-app 3（首页/头条列表/详情/搜索/我的 5 页 + tabBar，H5/微信小程序多端编译）
- ✅ Redis 业务缓存（cache-aside：region/category 树 30min + news 列表 60s，写操作 SCAN+DEL 按前缀失效，Redis 不可用全链路降级走 DB）
- ✅ WebSocket 实时通知（Hub 单用户多连接 + JWT 鉴权 /ws 端点 + 定向推送/广播，点赞实时通知作者，不在线 fire-and-forget 丢弃）
- ✅ 数据库初始化脚本（deploy/initdb PostGIS 扩展 + Makefile migrate/swagger 目标补齐）
- ✅ 高德地图 API 集成（地理编码/逆地理编码/周边搜索，key 未配置降级 503，限流 30/min）

### 未实现（待开发）
- ❌ PostGIS 空间查询业务接入
- ❌ 第三方登录、手机验证码登录
- ❌ 七牛云 Kodo 存储、阿里云 OSS 直传

## 许可证

MIT License
