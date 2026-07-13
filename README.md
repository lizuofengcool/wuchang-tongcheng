# 五常同城本地生活服务平台

五常同城是一个面向五常市的本地生活服务平台，提供分类信息、同城头条、商家服务等功能。

## 技术栈

### 后端
- **语言**: Go 1.22+
- **Web框架**: Gin
- **ORM**: GORM
- **架构模式**: 插件化架构 + Repository模式
- **数据库**: PostgreSQL 16（PostGIS 扩展已部署并接入业务：news 附近信息查询走 ST_DWithin 球面距离，扩展不可用降级纯 SQL Haversine）
- **缓存**: Redis 7（已封装，限流 + 业务缓存接入：region/category 树 30min TTL、news 列表 60s TTL，写操作按前缀失效，Redis 不可用降级走 DB）
- **搜索引擎**: Elasticsearch 8（已集成：news 全文检索 multi_match + 异步索引，ES 不可用降级 DB LIKE）
- **消息队列**: RabbitMQ（已集成：news 写入异步索引解耦，topic 交换机发布订阅，手动 ack）
- **实时通信**: WebSocket（已实现：Hub 连接管理 + JWT 鉴权升级端点 /ws，单用户多连接定向推送 + 全局广播，点赞实时通知作者）
- **对象存储**: 已实现 LocalStorage + MinIO（S3 协议兼容，可适配 AWS S3/阿里云 OSS/腾讯云 COS）+ 七牛云 Kodo（含降级机制）+ 阿里云 OSS STS 临时凭据直传（pkg/sts，下发短期 AK/SK/Token 供前端直传）
- **地图服务**: 高德地图API（已实现：地理编码/逆地理编码/周边 POI 搜索，key 未配置降级返回 503）
- **鉴权**: JWT + RBAC（用户-角色-权限，超级管理员直通）
- **API文档**: Swagger（gin-swagger + swaggo/swag，已集成）
- **限流防刷**: 基于 Redis INCR 的固定窗口限流（登录 5/min、新闻读取 60/min、点赞 30/min），Redis 不可用时优雅降级
- **CI/CD**: GitHub Actions（backend go vet/build/test、frontend npm test/build、tag 触发 docker publish 推送 GHCR）
- **测试**: 后端单元测试（标准库 testing，无外部依赖）+ 后端集成测试（testcontainers-go + testify，自动拉起 PostgreSQL 容器，无 Docker 时优雅 SKIP）+ 前端单元测试（Vitest，纯函数 + Pinia 状态校验，node 环境免 DOM）

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
│   │   │   ├── storage/        # 文件存储（已实现 local/minio/qiniu）
  │   │   │   ├── seed/           # 种子数据（地区/权限/admin）
  │   │   │   ├── sms/            # 短信验证码服务（Provider + CodeStore Redis/内存 + 生成/校验）
  │   │   │   ├── sts/            # 阿里云 OSS STS 临时凭据（Provider + NoopProvider 降级 + AssumeRole）
  │   │   │   ├── oauth/          # 第三方 OAuth 登录（Provider + Mock/WeChat + code 换身份）
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
- v1.2.0 - 手机验证码登录（D21）
  - pkg/sms 包：Provider 接口（NoopProvider mock/dev）+ CodeStore（Redis 优先，不可用降级内存）+ crypto/rand 生成 + 一次性消费 + 尝试次数限制（防暴力枚举）
  - user 模块新增 POST /api/v1/user/sms/code（发送验证码，限流 5/min）、POST /api/v1/user/login/sms（验证码登录，限流 5/min）
  - 新手机号验证通过自动注册（用户名=手机号，随机占位密码无法走密码登录），老用户直接签发 JWT，禁用用户拒绝登录
  - 配置项 sms.provider（mock/aliyun 预留）+ dev_return_code（联调返回验证码明文）
  - 单元测试 14 用例（sms 包 8 + user service SMS 登录 6）
- v1.3.0 - 阿里云短信 SDK 真实接入（D22）
  - AliyunProvider：dysmsapi.aliyuncs.com RPC API（SendSms）+ HMAC-SHA1 签名（RPC v1 规则，RFC3986 percentEncode）
  - 标准库 net/http 实现，无新外部依赖（与 pkg/amap 风格一致），crypto/rand 生成 SignatureNonce
  - resolveProvider 升级：provider=aliyun 且 AK/SK/SignName/TemplateCode 齐全 → AliyunProvider；任一缺失或占位（your-）→ 降级 NoopProvider
  - 模板参数 {"code":"xxxxxx"} 自动 marshal，响应 Code!=OK 包装错误（含阿里云错误码）
  - 单元测试 11 用例：percentEncode（含中文/空格/特殊符号）、canonicalQueryString 排序、IsAvailable 四项校验、Send 成功（httptest 独立验证签名）/失败（业务限流）/网络错误/非 JSON 响应/空手机号/配置不全、resolveProvider 正向解析、签名确定性
- v1.4.0 - PostGIS 空间查询业务接入（D23）
  - pkg/geo 包：HaversineKm 球面距离 + BoundingBox 经纬度边界框 + PostGISAvailable 扩展探测（进程级缓存）
  - news 模块新增 GET /api/v1/news/nearby（限流 30/min）：以 (lat,lng) 为中心返回 radius_km 公里内已发布信息，按距离升序、加急置顶
  - 双路实现：PostGIS 可用走 ST_DWithin + ST_Distance(geography) 精确球面距离；不可用降级纯 SQL Haversine + 边界框预筛（走普通索引）
  - DTO 新增 NewsNearbyRequest + NewsInfo.distance（公里，仅 nearby 接口回填）；半径默认 5km、上限 100km 自动钳制；经纬度范围校验
  - 单元测试 14 用例（Haversine 对称性/北京-上海/赤道/对跖点/哈尔滨-五常、边界框赤道/高纬/零半径/负半径/极点退化、PostGIS nil 探测）+ 集成测试（半径过滤/距离排序/Distance 回填/草稿排除/无坐标排除/分类过滤/半径钳制，无 Docker 自动 SKIP）
- v1.5.0 - 预签名直传（D24）
  - Storage 接口新增 PresignPut / AccessURL：MinIO（S3 协议，兼容 AWS S3 / 阿里云 OSS / 腾讯云 COS）走 minio-go PresignedPutObject 本地 SigV4 签名（无网络请求），LocalStorage / QiniuStorage 返回 ErrPresignNotSupported 优雅降级
  - file 模块新增 POST /api/v1/file/presign（换取预签名 PUT URL）+ POST /api/v1/file/commit（直传后按 object_name 由后端重新拼装访问 URL 落库，避免前端伪造）；权限 file:upload，复用现有类型/大小校验（50MB 上限、扩展名白名单）
  - 前端先 presign 拿 upload_url → 直接 PUT 二进制到对象存储（绕过后端带宽，适合大文件）→ commit 提交记录；本地存储环境回退普通 POST /upload
  - 错误码新增 1306 CodeFilePresignError（当前存储不支持预签名直传）
  - 单元测试 16 用例：storage 离线签名 URL 拼装/签名参数/有效期/对象名唯一 + service 类型/大小校验 + 本地存储降级路径
- v1.6.0 - 阿里云 OSS STS 临时凭据直传（D25）
  - pkg/sts 包：Provider 接口 + NoopProvider（降级）+ AliyunProvider（sts.aliyuncs.com AssumeRole RPC API + HMAC-SHA1 签名，与 pkg/sms 同算法，标准库 net/http 无新依赖）
  - file 模块新增 POST /api/v1/file/sts：后端用 AK/SK 调用 AssumeRole 扮演 RAM 角色换取临时凭据（AccessKeyID/AccessKeySecret/SecurityToken/Expiration）+ OSS 落地信息（Bucket/Region/Endpoint/ObjectPrefix）下发给前端；前端用临时凭据 + x-amz-security-token 头直接 PUT 到 OSS，一组凭据可复用上传多对象（与 /file/presign 单对象一次性 URL 互补，适合批量/大文件）
  - 降级策略与项目其它第三方集成一致：provider 非 aliyun 或 AccessKey/SecretKey/RoleArn 任一缺失/占位（your-）→ NoopProvider → AssumeRole 返回 ErrNotConfigured → 接口回 1307，前端回退 /file/upload 或 /file/presign
  - 配置新增 sts 块（与 storage 分离：storage 是后端自用 AK/SK，sts 是下发前端的临时凭据，需独立 RoleArn）；DurationSeconds 范围 900~3600 自动钳制；错误码新增 1307 CodeFileSTSError
  - 前端 src/api/file.js 新增 getSTSCredentials() 封装
  - 单元测试 28 用例：sts 包（percentEncode/canonicalQueryString/签名一致性/IsAvailable 七场景/resolveProvider 四场景/NoopProvider/normalizeObjectPrefix/DurationSeconds 钳制/RoleSessionName 默认/AssumeRole 成功+业务错误+网络错误+非JSON+空凭据+未配置/Init 解析）+ file service 降级路径
- v1.7.0 - 微信 OAuth 第三方登录（D26）
  - pkg/oauth 包：Provider 接口 + NoopProvider（降级）+ MockProvider（联调：code="mock:<openid>[:<nickname>]" 直接构造身份不访问网络）+ WeChatProvider（开放平台网站应用 OAuth2，标准库 net/http 无新依赖，与 pkg/sms/sts 同风格）
  - WeChatProvider 两段式换取：code → sns/oauth2/access_token（access_token+openid+unionid）→ sns/userinfo（nickname+headimgurl），errcode 错误码识别、空凭据校验、5s 超时
  - 配置 oauth.wechat（provider=""/mock/wechat，AppID/AppSecret 缺失或占位 your- 降级不启用）；resolveWeChatProvider 与 sms/sts 同风格
  - user 模块新增 UserOAuth 绑定模型（(provider,open_id) 唯一索引 + union_id 索引）+ UserOAuthRepository + service.OAuthLogin：code 换身份 → 命中绑定登录既有用户，未命中自动注册（username=provider_openid，随机占位密码无法走密码登录）+ 写绑定 + 签发 JWT；禁用用户拒绝登录
  - 路由 POST /api/v1/user/login/oauth/:provider（公开，复用 login 限流 5/min）；错误码新增 4006 CodeOAuthError
  - 单元测试 25 个（oauth 包 16：MockProvider 解析/NoopProvider/resolveWeChatProvider 八场景/NewService nil-mock-wechat/Login 空码-未知 provider/WeChatProvider IsAvailable 六场景/未配置/空码/httptest 两段式成功-badcode-userinfoerr-网络错误；user service 9：未配置/provider 未注册/OAuth 错误透传/命中绑定登录/自动注册+绑定/空昵称兜底/禁用用户/写绑定失败/genOAuthUsername 截断）
- v1.8.0 - 前端 axios 封装单元测试
  - src/utils/__tests__/request.test.js 18 用例补齐 axios 请求拦截器（JWT Bearer token / X-Region-ID 地区头注入 + rejected 透传）+ 响应拦截器业务码路由（code=0 直通 / 非 0 ElMessage.error / 401·1004·2006·2007·2008 触发未授权流程）+ HTTP 错误码（401·403·500 跳转去重·502·timeout·网络异常）+ 未授权弹窗去重（unauthorizedShown 闭包标志，连续 401 仅弹一次）
  - mock 策略：惰性 wrapper 包裹 vi.fn 切断 element-plus / @/router / @/stores/user 依赖（与 stores/__tests__ 同风格）；vi.resetModules + 动态 import 重置模块级 unauthorizedShown 闭包，每个用例获取全新 axios 实例与拦截器 handler
  - 前端单测总数 34 → 68（format 21 + auth 13 + request 18 + region 13 + user 3）
- v1.9.0 - 前端 v-permission/v-role 自定义指令单元测试
  - src/directives/__tests__/permission.test.js 17 用例覆盖 v-permission / v-role 指令 mounted 钩子（有权限保留元素、无权限默认从父节点移除、无权限 .hide 修饰符仅设 display:none、数组权限码透传、parentNode 为空优雅降级不抛错）+ updated 钩子（非 .hide 模式空操作、.hide 模式按权限重评估切换 display 空串/none）+ 导出形态校验
  - mock 策略：vi.mock('@/utils/auth') 包裹 hasPermission/hasRole 切断 Pinia store 依赖链，伪元素对象（style + parentNode.removeChild）模拟 Vue 指令钩子入参 el；与 stores/__tests__ 同风格 vi.resetModules + 动态 import
  - 前端单测总数 68 → 85（format 21 + auth 13 + request 18 + region 13 + user 3 + permission 17）
- v2.0.0 - 前端 user store 单元测试补齐
  - src/stores/__tests__/user.test.js 由 3 用例（仅 loginBySms）扩展至 24 用例，覆盖 state 初始化（localStorage 回填 token/userInfo/permissions/roles + 空串 fallback `|| 'null'` / `|| '[]'` 路径）、getters（isLoggedIn/nickname/avatar/isSuperAdmin 四个计算属性边界）、login action（密码登录成功 + API 失败不写 token + 权限拉取失败不阻断）、loginBySms action（保留原 3 用例）、fetchProfile action（成功更新持久化 + 失败不修改既有值）、fetchAuth action（成功持久化 + permissions/roles 缺失兜底空数组 + data 整体缺失兜底 + 异常静默不抛错）、logout action（清空 state + 移除 localStorage 用户键 + 不误删无关键如 currentRegionId + 重复调用不抛错）
  - mock 策略沿用 region store 风格：vi.mock('@/api/user') + vi.mock('@/api/permission') 包裹 login/loginBySms/getUserInfo/myAuth 切断网络依赖，beforeEach 内 setActivePinia(createPinia()) + localStorage.clear() + mockReset 保证用例隔离
  - 前端单测总数 85 → 106（format 21 + auth 13 + request 18 + region 13 + user 24 + permission 17）
- v2.1.0 - 前端 API 层单元测试补齐
  - src/api/__tests__/ 新增 7 个测试文件覆盖全部 7 个 API 模块（user/category/news/region/file/permission/setting），共 72 用例
  - user API 17 用例：公开鉴权接口（register/login/sendSmsCode/loginBySms/loginByOAuth 路径插值 + 不同 provider 切换/getUserInfo/updateProfile/changePassword）+ 管理后台接口（listUsers params 透传 + 无参兜底/getUser/adminCreateUser/adminUpdateUser/updateUserStatus 仅含 status/resetUserPassword 仅含 new_password/deleteUser）
  - category API 7 用例：getCategoryTree 无参 / getCategoryChildren parent_id 走 query + 不同 ID 切换 / getCategory 路径插值 / createCategory body 透传 / updateCategory / deleteCategory
  - news API 10 用例：listNews params 透传 + 无参兜底 / getNews / createNews / updateNews / deleteNews / toggleNewsStatus 仅含 status / searchNews params 透传 / likeNews toggle / getNewsLikeStatus
  - region API 7 用例：getRegionTree / getRegionChildren parent_id 走 query + 0 顶层 / getRegion / createRegion / updateRegion / deleteRegion
  - file API 8 用例：uploadFile 构造 FormData 携带 file + multipart 头 + onUploadProgress 函数配置；onProgress 回调按 loaded/total 换算百分比 + 多次触发累计调用 + e.total=0 不调用（避免除零）+ 未传 onProgress 不抛错；listFiles params 透传 + 无参 / deleteFile / getSTSCredentials 无 body
  - permission API 15 用例：角色（listRoles/getRole/createRole/updateRole/deleteRole/getRolePermissions 回显）/ 权限（listPermissions/createPermission/updatePermission/deletePermission）/ 分配（assignRoles/assignPermissions body 透传/getUserRoles 路径插值）/ 当前用户授权（myPermissions/myAuth）
  - setting API 8 用例：getAllSettings / getSettingsByGroup 路径插值 + 不同 group 切换 / createSetting / updateSetting / deleteSetting / batchUpdateSettings items 包裹在 { items } 内 + 空数组兜底
  - mock 策略：vi.mock('@/utils/request') 包裹 get/post/put/delete 四方法，单个 requestMock vi.fn 以首参 method 区分，记录 url + body/params/config；与 stores/__tests__ 同风格 mockReset + Promise.resolve 固定返回值
  - 前端单测总数 106 → 178（format 21 + auth 13 + request 18 + region store 13 + user store 24 + permission directive 17 + user API 17 + category API 7 + news API 10 + region API 7 + file API 8 + permission API 15 + setting API 8）
- v2.2.0 - category service 单元测试补齐
  - backend/internal/modules/category/service/category_test.go 24 用例，覆盖纯函数（categoryCacheKeyTree/categoryCacheKeyByID/categoryCacheKeyByParent 缓存键生成 + toCategoryInfo DTO 转换含 NilSafe 路径）+ 构造函数（NewCategoryService 返回 *categoryService 类型断言）+ Create 业务逻辑（顶级 level=1 + Status=0 默认填充 1 + 子级 level=父级+1 + 第三层边界 + 超过 MaxCategoryLevel=3 拒绝 ErrCategoryMaxLevel + 父分类不存在 ErrCategoryParentInvalid + repo Create 错误透传 + 父查询非 NotFound 错误原样透传）+ GetTree（空列表 + 三层多级树递归构建：root→child→grandchild + 兄弟节点）+ GetAll（空列表 + 平铺列表含层级信息 + repo 错误透传）+ Delete（NotFound + HasChildren 阻止 + 叶子节点删除后二次 NotFound）+ Update（NotFound + 全字段更新回写校验 + 仅传 name 时空字符串字段不覆盖原值而 sort 始终被更新为零值）
  - mock 策略沿用 permission service 风格：内存 mockCategoryRepo 实现 CategoryRepository 全部 7 方法（Create/FindByID/FindByParentID/FindByRegionID/Update/UpdateFields/Delete），byID map + byParent map + byRegion slice 三索引维护；ID 在嵌入 BaseModel 中无法在结构体字面量直接赋值，构造后单独设置；Redis 不可用时 GetJSON 降级 miss、SetJSON no-op，cache-aside 链路自动走 mock repo
- v2.3.0 - region service 单元测试补齐
  - backend/internal/modules/region/service/region_test.go 38 用例，覆盖纯函数（regionCacheKeyTree/regionCacheKeyByID/regionCacheKeyByParent 缓存键生成 + toRegionInfo DTO 转换含 NilSafe 路径）+ 构造函数（NewRegionService 返回 *regionService 类型断言）+ Create 业务逻辑（顶级 level=1 + Status=0 默认填充 1 + 子级 level=父级+1 + 第三层边界 + 超过 MaxRegionLevel=3 拒绝 ErrRegionMaxLevel + 父地区不存在 ErrRegionParentInvalid + 编码重复 ErrRegionCodeExists + repo Create 错误透传 + 父查询非 NotFound 错误原样透传 + 编码查询非 NotFound 错误原样透传）+ GetTree（空列表 + 三层多级树递归构建：root→child→grandchild + 兄弟节点 + repo 错误透传）+ GetAll（空列表 + 平铺列表含层级信息 + repo 错误透传）+ GetByParentID（空列表 + 多子节点回写 + repo 错误透传）+ GetByID（NotFound 转 ErrRegionNotFound + 成功回写 + repo 错误透传）+ Delete（NotFound + HasChildren 阻止 + 叶子节点删除后二次 NotFound + 查询子节点错误透传 + repo Delete 错误透传）+ Update（NotFound + 全字段更新回写校验 + 仅传 name 时 sort 始终被更新为零值 + status 非 0/1 时不更新保留原值 + repo 错误透传）
  - mock 策略沿用 category service 风格：内存 mockRegionRepo 实现 RegionRepository 全部 8 方法（Create/FindByID/FindByCode/FindByParentID/FindAll/Update/UpdateFields/Delete），byID map + byParent map + byCode map + all slice 四索引维护；与 category 区别在于 region 是地区维度本身，无 regionID 入参，cache 键不含 regionID；Redis 不可用时 GetJSON 降级 miss、SetJSON/DelByPrefix no-op，cache-aside 链路自动走 mock repo
- v2.4.0 - news service 单元测试补齐
  - backend/internal/modules/news/service/news_test.go 64 用例，覆盖纯函数（newsCacheKeyList 缓存键生成含 nil *bool 的 %v 格式化 + toNewsInfo/toCommentInfo/toMessageInfo DTO 转换含 NilSafe 路径）+ 构造函数（NewNewsService 返回 *newsService 类型断言 + nil indexer 默认替换为 NoopIndexer）+ Create 业务逻辑（默认 ExpireDays=30 + 自定义 ExpireDays + 空 PriceUnit 默认"元" + 空 ListingType 默认 sell + 空 Condition 默认 used + Status=1 设置 PublishedAt + Status=0 草稿不设置 + RegionID 注入 + repo Create 错误透传 + indexer.OnIndex 调用验证）+ Update（NotFound 转 ErrNewsNotFound + AuthorID!=operatorID 返回 ErrNewsNoPermission + 草稿→发布状态切换设置 published_at + FindByID 非 NotFound 错误透传 + UpdateFields 错误透传）+ Delete（NotFound + 无权限 + 成功删除+indexer.OnDelete 调用 + repo Delete 错误透传）+ GetByID（NotFound + 成功 IncrViewCount 副作用 + 返回 ViewCount+1 + repo 错误透传）+ List（空列表 + 多结果 + keyword 走非缓存路径 + repo 错误透传）+ ListNearby（RadiusKm=0 默认 5 + 有结果 + repo 错误透传）+ Like（NotFound + toggle on 计数+1+IncrLikeCount + toggle off 计数-1+DecrLikeCount + 计数钳制 0 不为负 + 自赞不产生消息 + FindByID 错误透传 + LikeExists 错误透传）+ LikeStatus（NotFound + liked/notLiked）+ Fav（NotFound + toggle on + toggle off 钳制 + FavExists 错误透传）+ FavStatus（NotFound + faved/notFaved）+ CreateComment（NotFound + 成功+IncrCommentCount+作者消息通知 + 自评不产生消息 + repo CreateComment 错误透传）+ ListComments（按 newsID 过滤）+ DeleteComment + ListMessages（按 userID 过滤）+ UnreadCount（未读计数）+ MarkRead（按 ids 选择性标记）+ Search（ES 不可用 DB LIKE 降级 + repo 错误透传）+ buildESFilters（仅 status + 全过滤器 status/region/category/listingType + 部分过滤器组合）
  - mock 策略沿用 category/region service 风格：内存 mockNewsRepo 实现 NewsRepository 全部 27 方法（Create/FindByID/Update/UpdateFields/Delete/List/ListNearby/IncrViewCount/LikeExists/CreateLike/DeleteLike/IncrLikeCount/DecrLikeCount/FavExists/CreateFav/DeleteFav/IncrFavCount/DecrFavCount/CreateComment/ListComments/DeleteComment/IncrCommentCount/CreateMessage/ListMessages/UnreadCount/MarkRead/FindByIDs），byID map + likes/favs 嵌套 map + comments/messages 切片 + 各 Incr/Decr 计数 map 六索引维护；额外 mockIndexer 记录 OnIndex/OnDelete 调用用于断言索引同步副作用；Redis 不可用时 GetJSON 降级 miss、SetJSON/DelByPrefix no-op，cache-aside 链路自动走 mock repo；ws Hub 未初始化时 GetHub 返回 nil，Like 消息通知路径跳过（CreateMessage 不调用），CreateComment 消息通知不依赖 Hub 直连 repo
- v2.5.0 - file service 主流程单元测试补齐
  - backend/internal/modules/file/service/file_test.go 16 用例，与既有 file_presign_test.go / file_sts_test.go 互补，共同构成 file service 完整单元测试矩阵。覆盖 Upload 输入校验路径（size=0 ErrFileEmpty + 负 size 视为空 + 超 50MB 上限 ErrFileTooLarge + 不支持扩展名 .exe ErrFileTypeInvalid）+ List 主流程（空列表 total=0 + 多结果 total 计数 + regionID 隔离过滤 + 分页参数 page/pageSize 透传 + FileType 筛选透传 + Keyword 子串匹配透传 + 默认分页 1/10 兜底 + repo.List 错误原样透传）+ Delete 主流程（NotFound 转 ErrFileNotFound + GetByID 非 NotFound 错误透传 + storage.Delete 失败不阻塞 repo.Delete 成功 + repo.Delete 错误透传 + 合法 /uploads/ 前缀 LocalStorage 静默删除）+ 构造函数 NewFileService 返回 *fileService 类型断言
  - mock 策略沿用 category/region/news service 风格：内存 mockFileRepo 实现 FileRepository 全部 4 方法（Create/GetByID/List/Delete），byID map + lastXxx 字段记录最近一次调用参数用于断言透传 + xxxErr 注入错误用于断言透传；Upload 成功路径会真实落盘 ./uploads 故仅覆盖校验路径（size/扩展名校验在 storage.Save 之前返回，不依赖存储与 repo）；Delete 路径中 storage.GetStorage() 兜底返回 LocalStorage，对非 /uploads/ 前缀 URL 返回 "invalid file url" 错误但 service 层 _ 忽略，验证降级不阻塞；LocalStorage.Delete 对不存在的 /uploads/ 路径静默（os.IsNotExist 不报错），故 Delete 成功路径无副作用
- v2.6.0 - 前端路由守卫单元测试补齐
  - frontend/src/router/__tests__/index.test.js 19 用例，补齐此前 0 覆盖的 SPA 安全闸门 beforeEach 全局前置守卫全分支测试。覆盖 document.title 设置（meta.title 拼接 APP_TITLE + meta 缺失/整体缺失回退 APP_TITLE 不抛错）+ 登录页/错误页放行（/login 未登录放行 / /login 已登录跳首页 / /403 /500 与登录态无关放行 / 已登录但无权限码访问 403 仍放行不二次鉴权）+ 未登录访问受保护路由跳 /login 携带 redirect（取 to.fullPath 保留 query / 无 meta.permission 的 dashboard 也跳 /login / redirect 用 fullPath 而非 path 保留 query 参数）+ 已登录路由权限校验（持有 meta.permission 放行 / 缺权限码跳 /403 / admin 角色超管直通任意 permission / 无 meta.permission 放行不进 403 分支 / 多权限码命中其一放行 / permission 数组形式任一命中放行 hasPermission 数组语义）+ 守卫注册（router.beforeEach 已注册 function 类型回调 / router.routes 配置可读供 MainLayout 菜单派生）
  - mock 策略：vi.hoisted 提升共享 capturedGuard 容器（vi.mock 工厂被 Vitest 提升到文件顶部早于 const 执行，hoisted 同步提升保证工厂执行时容器已初始化）+ vi.mock('vue-router') 把 createRouter 改为返回壳 router 并在 beforeEach 调用时把回调缓存到 capturedGuard.current 供用例直接驱动（createWebHistory 改为返回空对象避免 node 环境访问 window/history 抛错）+ vi.mock('@/api/user')&@/api/permission 切断 api→request→router 依赖链（与 stores/__tests__ 同风格），用例通过 setActivePinia+createPinia + $patch 构造登录态/权限码/角色码；node 环境无 document，beforeEach 内 globalThis.document 兜底桩；next 改写为数组记录调用参数，单次断言 toEqual([undefined]) 或 ["/"] 或 [{ path: "/403" }] 或 [{ path: "/login", query: { redirect: fullPath } }]
- v2.7.0 - 前端头条评论 + 消息通知中心
  - 后端 news 模块早已实现评论与消息 API（POST/GET /news/:id/comments、DELETE /news/comments/:id、GET /news/messages、GET /news/messages/unread、PUT /news/messages/read），本次补齐前端缺口，打通全链路
  - frontend/src/api/news.js 新增 6 个 API 封装：listComments（公开评论列表，params 透传分页）/ createComment（发表评论，body 透传 content+parent_id+reply_to）/ deleteComment（按 commentId 删除）/ listMessages（我的消息分页）/ getUnreadCount（未读数）/ markMessagesRead（PUT /news/messages/read，ids 空数组兜底为 []）
  - frontend/src/views/news/detail.vue 新增评论区：评论数标题（同步 news.comment_count）+ 登录用户展示 el-input textarea 发表评论（500 字上限 + 空内容禁用提交按钮 + 提交后重载第一页保证顺序最新）+ 未登录展示 el-alert 引导登录（router-link 携带 redirect）+ 评论列表（头像/用户名/回复 @某人/时间/内容 + is_dot 未读小红点 + 加载更多分页）+ 删除按钮（canDeleteComment：超管 isSuperAdmin 直通或作者本人 user_id 匹配，ElMessageBox 二次确认 + 本地列表过滤 + 计数同步）
  - frontend/src/views/message/index.vue 新建消息通知中心页面：刷新/全部已读/标记选中已读工具栏 + 全部/未读/已读三态客户端筛选（filter 切换不重新请求，仅前端过滤当前页；filter=all 用后端 total，其他模式用 filteredList.length）+ el-table 多选（type=selection）+ 消息类型 el-tag（like/comment/reply/system 四色）+ 来源跳转头条详情 + 状态列 + 标记单条已读 + el-pagination 分页 + 本地 unreadCount/stats 同步（标记已读后递减未读数并推算已读数）
  - frontend/src/router/index.js 注册 GET /message 路由（name: Message，icon: Bell，无 meta.permission 故任何登录用户可访问自己的消息）
  - frontend/src/layouts/MainLayout.vue 头部新增未读消息徽标：el-badge（max=99，unreadCount=0 时 :hidden 隐藏）+ Bell 图标点击跳转 /message + onMounted 初始拉取一次 + setInterval 60s 轮询 getUnreadCount + router.afterEach 钩子进入 /message 后延迟 800ms 刷新（让页面先标记已读）+ onBeforeUnmount 清理 timer 与 afterEach 钩子避免内存泄漏
  - frontend/src/api/__tests__/news.test.js 新增 14 用例：评论 7（listComments params 透传/无参/不同 newsId 路径插值、createComment body 透传/仅 content、deleteComment 路径插值/不同 commentId）+ 消息 7（listMessages params 透传/无参、getUnreadCount 无参、markMessagesRead ids 透传/空数组/不传参兜底 []/null 兜底 []）
  - 前端单测总数 197 → 211（news API 由 10 增至 24）
- v2.8.0 - user service 主流程单元测试补齐
  - backend/internal/modules/user/service/user_crud_test.go 50 用例，与既有 user_test.go（HashPassword/CheckPassword 纯函数）、sms_login_test.go、oauth_login_test.go 互补，共同构成 user service 完整单元测试矩阵。覆盖纯函数（toUserInfo DTO 转换含 NilSafe 路径 + randomPassword 长度/n<=0 默认 32/两次生成不同）+ 构造函数（NewUserService 返回 *userService 类型断言）+ Register 业务逻辑（成功 + Nickname 为空回退 Username + 用户名已存在 ErrUserAlreadyExists + FindByUsername 非 NotFound 错误原样透传 + Create 错误透传 + 密码哈希而非明文落库）+ Login（成功签发 token + 用户不存在 ErrUserNotFound + FindByUsername 非 NotFound 错误透传 + 禁用用户 ErrUserDisabled + 密码错误 ErrPasswordInvalid）+ GetUserInfo（成功 + NotFound 转 ErrUserNotFound + repo 错误透传）+ UpdateProfile（全部零值 no-op 不调 UpdateFields + 部分字段 + Gender=0 伴随其他字段写入 + 仅 Gender!=0 触发更新 + UpdateFields 错误透传）+ ChangePassword（成功 + 新密码可登录 + NotFound + 旧密码错误 ErrOldPasswordWrong 不调 UpdateFields + Find 错误透传）+ ListUsers（空列表 + 多结果 total 计数 + 状态过滤 status=1/0/-1 三态 + regionID 隔离 regionID=0 跨区/regionID>0 本区 + 分页/keyword/status 参数透传 + 默认分页 1/10 兜底 + repo.List 错误透传）+ AdminCreateUser（成功 + Status=0 默认填充 1 + Nickname 为空回退 Username + 用户名已存在 + FindByUsername 非 NotFound 错误透传 + Create 错误透传 + region_id 写入）+ AdminUpdateUser（部分字段 + 全字段 + gender 总是写入 + UpdateFields 错误透传）+ UpdateUserStatus（成功 + 禁用后登录被拒 + 启用后恢复 + UpdateFields 错误透传）+ ResetPassword（成功 + 新密码可登录旧密码失效 + UpdateFields 错误透传）+ DeleteUser（成功 + 删除后 GetUserInfo NotFound + NotFound 转 ErrUserNotFound + Find 错误透传 + Delete 错误透传）
  - mock 策略沿用 category/region/news/file service 风格：内存 mockUserRepo 实现 UserRepository 全部 9 方法（Create/FindByID/FindByUsername/FindByPhone/Update/UpdateFields/List/Delete），byID map + nextID 自增 + 各路径错误注入字段（createErr/findByIDErr/findByUsernameErr/findByPhoneErr/updateErr/updateFieldsErr/listErr/deleteErr）+ 调用记录（lastListRegionID/lastListKeyword/lastListStatus/lastListPage 验证参数透传 + updatedFields map 记录 UpdateFields 入参 + deletedIDs 切片记录删除）；与 sms_login_test.go 的 fakeUserRepo 区别在于支持 Update/UpdateFields/List/Delete 真实语义以便覆盖管理后台路径；bcrypt 哈希真实执行（HashPassword/CheckPassword 已在 user_test.go 验证）；JWT GenerateToken 使用包级默认 secretKey 无需 Init；List mock 严格对齐真实仓库语义（status∈{0,1} 才过滤，status=-1 不过滤，对应 DTO 注释 -1全部/0禁用/1正常）
- v2.9.0 - category handler HTTP 处理层单元测试补齐
  - backend/internal/modules/category/handler/category_handler_test.go 25 用例，补齐此前 0 覆盖的 category handler 装配层（此前仅 user 模块有 handler 级测试，且为 testcontainers e2e 需 Docker）。使用 gin + httptest + 内存 mockCategoryService（实现 service.CategoryService 接口全 7 方法，记录 lastXxx 入参 + 预设 result/err），无 DB/Redis/Docker 依赖，纯内存验证 handler 装配逻辑。覆盖 Create（成功 + 未登录 userID=0 拦截 401 + Bind 失败非法 JSON 400 + service 错误业务码 2301 透传，验证 regionID 透传 + 请求体透传）+ Update（成功 + 未登录 + 非数字 :id 解析失败 400 + Bind 失败 + service 错误透传）+ Delete（成功 + 未登录 + 非数字 :id + service 错误如"存在子分类"透传）+ GetByID（成功 + 非数字 :id + service NotFound 业务码 2302 透传）+ GetByParentID（成功 + parent_id 缺失默认 0 + 非数字 parent_id 兜底 0 + service 错误透传，验证 ParentID/RegionID 透传）+ GetTree（成功含 Children 嵌套断言 + service 错误透传，验证 regionID 透传）+ GetAll（成功 + service 错误透传）+ regionID=0 上下文透传边界（ctx.Get 命中 uint(0) 不走 DefaultRegionID 兜底）
  - mock 策略：mockCategoryService 实现 CategoryService 接口记录每次调用入参（lastCreateRegionID/lastCreateReq/lastUpdateID/lastUpdateReq/lastDeleteID/lastGetByID/lastGetByParent/lastGetAllRegionID/lastGetTreeRegionID）+ 预设返回值（createResult/getByIDResult/getByParentResult/getAllResult/getTreeResult）+ 各路径错误注入字段（createErr/updateErr/deleteErr/getByIDErr/getByParentErr/getAllErr/getTreeErr）；newHandlerEnv 构造 gin 引擎 + 模拟 Region+Auth 中间件（c.Set RegionIDKey/ContextUserID）+ 经 coreRouter.RouterGroup 注册 7 条路由（路径与 category/plugin.go RegisterRoutes 一致，去掉权限中间件纯测 handler）+ doJSON/doRaw 发起请求解析 {code,message,data}；用例断言业务码（0 成功 / 400 参数错误 / 401 未登录 / 2301 CategoryError / 2302 CategoryNotFound）+ message 透传 + service 调用入参 + data 透传


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
- ✅ 后端单元测试（utils/setting/user 28 用例 + permission/category/region/news service 纯函数+业务逻辑 mock repo 用例 + file service Upload/List/Delete 主流程 + presign/sts 降级路径用例 + user service Register/Login/GetUserInfo/UpdateProfile/ChangePassword/ListUsers/AdminCreateUser/AdminUpdateUser/UpdateUserStatus/ResetPassword/DeleteUser 主流程 50 用例 + category handler 装配层 25 用例（gin+httptest+mock service，覆盖未登录拦截/参数解析/Bind 失败/业务码透传/regionID 注入，无 DB/Redis/Docker 依赖），覆盖无 DB/Redis 依赖）
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
- ✅ 七牛云 Kodo 对象存储（基于 github.com/qiniu/go-sdk/v7 经典 storage 包，Save 走 FormUploader + Delete 走 BucketManager，AK/SK 占位值自动降级到 local）
- ✅ 手机验证码登录（pkg/sms：Provider 接口 + NoopProvider + CodeStore Redis/内存双实现 + crypto/rand 生成 + 一次性消费 + 尝试次数限制；user 模块新增 /sms/code、/login/sms 路由，新手机号自动注册；mock provider + dev_return_code=true 联调返回验证码明文）
- ✅ 阿里云短信 SDK 真实接入（AliyunProvider：dysmsapi.aliyuncs.com RPC API + HMAC-SHA1 签名，标准库 net/http 无新依赖，与 pkg/amap 风格一致；AK/SK/SignName/TemplateCode 任一缺失或占位自动降级 NoopProvider；单元测试 11 用例覆盖 percentEncode/签名一致性/httptest 成功失败/网络错误/配置校验）
- ✅ PostGIS 空间查询业务接入（pkg/geo：HaversineKm + BoundingBox + PostGISAvailable 探测；news 模块 GET /api/v1/news/nearby 附近信息查询，PostGIS ST_DWithin 精确球面距离，扩展不可用降级纯 SQL Haversine + 边界框预筛；半径默认 5km/上限 100km 钳制、距离升序加急置顶、distance 字段回填；单元测试 14 用例 + 集成测试）
- ✅ 预签名直传（Storage 接口新增 PresignPut/AccessURL：MinIO/S3 走本地 SigV4 签名；file 模块 POST /file/presign 换 PUT URL + POST /file/commit 直传后按 object_name 拼装访问 URL 落库；LocalStorage/Qiniu 返回 ErrPresignNotSupported 降级普通上传；错误码 1306；单元测试 16 用例）
- ✅ 阿里云 OSS STS 临时凭据直传（pkg/sts：Provider 接口 + NoopProvider 降级 + AliyunProvider 走 sts.aliyuncs.com AssumeRole RPC API + HMAC-SHA1 签名，标准库 net/http 无新依赖；file 模块 POST /file/sts 下发临时 AK/SK/Token + OSS 落地信息供前端直传 OSS，一组凭据可复用上传多对象，与 /file/presign 单对象 URL 互补；配置缺失/占位降级 NoopProvider 回 1307；错误码 1307；单元测试 28 用例）
- ✅ 前端单元测试（Vitest 接入：独立 vitest.config.js 复用 @ 别名 + node 环境 + 内存 localStorage setup；src/utils/format.js 21 用例覆盖 formatTime/formatDate/formatSize 边界与状态文本、src/utils/auth.js 13 用例覆盖 hasPermission/hasRole/hasAllPermissions 含超管直通/数组任一/空码直通、src/utils/request.js 18 用例覆盖 axios 拦截器 JWT/X-Region-ID 头注入 + 业务码路由（0/非0/401/2006）+ HTTP 错误码（401/403/500 去重/502/timeout/网络异常）+ 未授权弹窗去重（unauthorizedShown 闭包标志，vi.resetModules 重置模块级状态）、src/directives/permission.js 17 用例覆盖 v-permission/v-role 指令 mounted/updated 钩子（保留/移除/.hide 隐藏/数组透传/空父节点降级）、src/stores/region.js 13 用例覆盖 state 初始化/currentRegionName 树遍历/loadTree/setCurrentRegion、src/stores/user.js 24 用例覆盖 state 初始化（localStorage 回填 + 空串 fallback）/getters（isLoggedIn/nickname/avatar/isSuperAdmin）/login（成功+失败+权限拉取降级）/loginBySms/fetchProfile/fetchAuth（兜底+静默降级）/logout（清 state+移除 localStorage 键+不误删无关键）、src/api/__tests__/ 7 个测试文件 72 用例覆盖全部 7 个 API 模块（user 17/category 7/news 10/region 7/file 8/permission 15/setting 8），mock @/utils/request 验证 URL/HTTP 方法/body/params/路径插值，file.uploadFile 额外验证 FormData 构造 + multipart 头 + onUploadProgress 进度回调（含 e.total=0 除零兜底）；mock api 层切断 router→createWebHistory 的 DOM 依赖链；src/router/index.js 19 用例覆盖 beforeEach 全局前置守卫全分支（document.title 拼接 / 登录页·错误页放行·已登录跳首页·未登录跳 /login 携带 redirect·路由权限校验放行/跳 403·admin 超管直通·hasPermission 数组语义）；frontend CI 新增 npm run test 步骤；共 197 用例）
- ✅ 微信 OAuth 第三方登录（pkg/oauth：Provider 接口 + NoopProvider 降级 + MockProvider 联调 + WeChatProvider 开放平台网站应用 OAuth2 两段式换取；user 模块 UserOAuth 绑定模型 + service.OAuthLogin 命中绑定登录/未命中自动注册+绑定 + 路由 POST /login/oauth/:provider；配置缺失降级，错误码 4006；单元测试 25 个）
- ✅ 管理后台短信验证码登录（前端登录页 Tab 切换"密码登录/短信登录"；api/user 新增 sendSmsCode/loginBySms/loginByOAuth 封装；stores/user 新增 loginBySms action 复用 token+权限拉取流程；手机号格式校验 + 60s 倒计时 + dev_code 联调自动回填；Vitest 单元测试 3 用例覆盖成功/失败/权限拉取降级）
- ✅ 前端头条评论 + 消息通知中心（v2.7.0）：news/detail.vue 新增评论区（发表/列表/删除自己的评论/加载更多，超管可删任意评论）+ 新建 views/message/index.vue 消息中心（全部/未读/已读三态筛选 + 批量标记已读 + 单条标记已读 + 来源跳转头条）+ MainLayout 头部未读消息 el-badge 徽标（60s 轮询 + 进入消息页延迟刷新）+ api/news.js 新增 6 个评论/消息 API 封装 + 14 用例单元测试（共 211 用例）
- ✅ 我的收藏列表（v2.7.1）：补齐收藏功能闭环（此前仅有 toggle 收藏/查询收藏状态，无法查看自己的收藏列表）。后端 news 模块新增 repository.ListFavs（按 created_at 倒序分页）+ service.ListFavorites（分页拉取收藏记录 → FindByIDs 批量回填 NewsInfo，保留收藏顺序，软删除 News 自动过滤）+ handler.ListFavorites + 路由 GET /news/favorites（auth + 60/min 限流，静态路由注册在 /:id 之前避免歧义）；前端 api/news.js 新增 listFavorites/favNews/getNewsFavStatus 封装 + 新建 views/news/favorites.vue 我的收藏页（表格展示标题/类型/价格/浏览/收藏数/状态/发布时间，支持分页 + 取消收藏 + 跳转详情，空态引导）+ 路由 /news/favorites 自动入侧边栏（Star 图标）；后端单元测试 5 用例（空列表/返回已收藏/分页/跳过软删除/非法分页参数兜底）+ 前端 API 单元测试 5 用例（共 216 用例）

### 未实现（待开发）
- 暂无待开发项，规划功能已全部落地（第三方登录已在 v1.7.0 完成）

## 许可证

MIT License
