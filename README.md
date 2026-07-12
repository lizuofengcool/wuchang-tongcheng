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
- ✅ 七牛云 Kodo 对象存储（基于 github.com/qiniu/go-sdk/v7 经典 storage 包，Save 走 FormUploader + Delete 走 BucketManager，AK/SK 占位值自动降级到 local）
- ✅ 手机验证码登录（pkg/sms：Provider 接口 + NoopProvider + CodeStore Redis/内存双实现 + crypto/rand 生成 + 一次性消费 + 尝试次数限制；user 模块新增 /sms/code、/login/sms 路由，新手机号自动注册；mock provider + dev_return_code=true 联调返回验证码明文）
- ✅ 阿里云短信 SDK 真实接入（AliyunProvider：dysmsapi.aliyuncs.com RPC API + HMAC-SHA1 签名，标准库 net/http 无新依赖，与 pkg/amap 风格一致；AK/SK/SignName/TemplateCode 任一缺失或占位自动降级 NoopProvider；单元测试 11 用例覆盖 percentEncode/签名一致性/httptest 成功失败/网络错误/配置校验）
- ✅ PostGIS 空间查询业务接入（pkg/geo：HaversineKm + BoundingBox + PostGISAvailable 探测；news 模块 GET /api/v1/news/nearby 附近信息查询，PostGIS ST_DWithin 精确球面距离，扩展不可用降级纯 SQL Haversine + 边界框预筛；半径默认 5km/上限 100km 钳制、距离升序加急置顶、distance 字段回填；单元测试 14 用例 + 集成测试）
- ✅ 预签名直传（Storage 接口新增 PresignPut/AccessURL：MinIO/S3 走本地 SigV4 签名；file 模块 POST /file/presign 换 PUT URL + POST /file/commit 直传后按 object_name 拼装访问 URL 落库；LocalStorage/Qiniu 返回 ErrPresignNotSupported 降级普通上传；错误码 1306；单元测试 16 用例）
- ✅ 阿里云 OSS STS 临时凭据直传（pkg/sts：Provider 接口 + NoopProvider 降级 + AliyunProvider 走 sts.aliyuncs.com AssumeRole RPC API + HMAC-SHA1 签名，标准库 net/http 无新依赖；file 模块 POST /file/sts 下发临时 AK/SK/Token + OSS 落地信息供前端直传 OSS，一组凭据可复用上传多对象，与 /file/presign 单对象 URL 互补；配置缺失/占位降级 NoopProvider 回 1307；错误码 1307；单元测试 28 用例）
- ✅ 前端单元测试（Vitest 接入：独立 vitest.config.js 复用 @ 别名 + node 环境 + 内存 localStorage setup；src/utils/format.js 21 用例覆盖 formatTime/formatDate/formatSize 边界与状态文本、src/utils/auth.js 13 用例覆盖 hasPermission/hasRole/hasAllPermissions 含超管直通/数组任一/空码直通、src/utils/request.js 18 用例覆盖 axios 拦截器 JWT/X-Region-ID 头注入 + 业务码路由（0/非0/401/2006）+ HTTP 错误码（401/403/500 去重/502/timeout/网络异常）+ 未授权弹窗去重（unauthorizedShown 闭包标志，vi.resetModules 重置模块级状态）；mock api 层切断 router→createWebHistory 的 DOM 依赖链；frontend CI 新增 npm run test 步骤；共 68 用例）
- ✅ 微信 OAuth 第三方登录（pkg/oauth：Provider 接口 + NoopProvider 降级 + MockProvider 联调 + WeChatProvider 开放平台网站应用 OAuth2 两段式换取；user 模块 UserOAuth 绑定模型 + service.OAuthLogin 命中绑定登录/未命中自动注册+绑定 + 路由 POST /login/oauth/:provider；配置缺失降级，错误码 4006；单元测试 25 个）
- ✅ 管理后台短信验证码登录（前端登录页 Tab 切换"密码登录/短信登录"；api/user 新增 sendSmsCode/loginBySms/loginByOAuth 封装；stores/user 新增 loginBySms action 复用 token+权限拉取流程；手机号格式校验 + 60s 倒计时 + dev_code 联调自动回填；Vitest 单元测试 3 用例覆盖成功/失败/权限拉取降级）

### 未实现（待开发）
- 暂无待开发项，规划功能已全部落地（第三方登录已在 v1.7.0 完成）

## 许可证

MIT License
