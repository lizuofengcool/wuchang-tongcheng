# 五常同城本地生活服务平台

五常同城是一个面向五常市的本地生活服务平台，提供分类信息、同城头条、商家服务等功能。

## 技术栈

### 后端
- **语言**: Go 1.22+
- **Web框架**: Gin
- **ORM**: GORM
- **架构模式**: 插件化架构 + Repository模式
- **数据库**: PostgreSQL 16（PostGIS 扩展已部署但代码未使用空间查询）
- **缓存**: Redis 7（已封装，业务模块接入待补齐）
- **搜索引擎**: Elasticsearch 8（基础设施已部署，代码集成待补齐）
- **消息队列**: RabbitMQ（基础设施已部署，代码集成待补齐）
- **实时通信**: WebSocket（规划中，待开发）
- **对象存储**: 已实现 LocalStorage；MinIO/七牛云Kodo 待补齐
- **地图服务**: 高德地图API（规划中，待开发）
- **鉴权**: JWT + RBAC（用户-角色-权限，超级管理员直通）
- **API文档**: Swagger（待补齐）

### 前端
- **管理后台**: Vue 3 + Vite + Element Plus + Pinia（当前已实现）
- **PC门户**: Next.js（规划中，待开发）
- **小程序**: Uni-app（规划中，待开发）

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
├── frontend/                   # 前端管理后台（Vue 3 + Vite + Element Plus）
│   ├── src/
│   │   ├── api/                # 接口封装
│   │   ├── components/         # 公共组件（RichTextEditor）
│   │   ├── directives/         # 自定义指令（v-permission/v-role）
│   │   ├── layouts/            # 布局（MainLayout）
│   │   ├── router/             # 路由 + 守卫
│   │   ├── stores/             # Pinia（user/region）
│   │   ├── utils/              # 工具（request/auth/format）
│   │   └── views/              # 页面（login/dashboard/profile/error/news/...）
│   ├── deploy/nginx.conf       # Nginx 配置
│   ├── Dockerfile              # 前端镜像构建
│   └── .env.development / .env.production
└── deploy/                     # 整体部署
    └── docker-compose.yml      # Docker Compose 配置（含 PG/Redis/RabbitMQ/ES/MinIO）
```

> 注：README 中提到的 `pc/`（Next.js 门户）与 `miniapp/`（Uni-app 小程序）尚未实现，仅为规划。`backend/docs/`、`backend/scripts/` 目录尚未建立。

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

## 功能完成度（对照规划）

> 本节用于诚实标注当前实际进度，避免与规划混淆。

### 已完成
- ✅ 插件化后端骨架（7 个业务模块均含 model/dto/repository/service/handler/plugin.go）
- ✅ RBAC 权限模型（用户-角色-权限，超级管理员直通，路由级权限校验）
- ✅ JWT 鉴权 + AuthRequired/RequirePermission 中间件全链路打通
- ✅ 地区数据隔离（中间件 + RegionBaseModel，news/category/setting 读取链路已生效）
- ✅ 种子数据（31 个权限码、5 个地区、admin 超管账号）
- ✅ Vue3 管理后台（login/dashboard/profile + 7 个业务管理页）
- ✅ 前端权限指令 v-permission/v-role、路由守卫 meta.permission
- ✅ 富文本编辑器组件（contenteditable + 图片上传）
- ✅ 前后端 Docker 多阶段构建 + Nginx 反代 + .env 配置

### 部分实现
- ⚠️ 对象存储：仅 LocalStorage，MinIO/七牛云为 stub
- ⚠️ Redis：封装完整但无业务模块调用
- ⚠️ setting 值类型：4 种类型仅作元数据标记，读取未反序列化
- ⚠️ news：状态流转 + 浏览量累加已实现，但无评论/点赞 API、无前端详情页
- ⚠️ region 隔离：news/category/setting 已生效，file.List 和 user 读写未隔离
- ⚠️ category/region：有树形结构但无层级深度限制

### 未实现（待开发）
- ❌ PC 门户站（Next.js）、小程序端（Uni-app）
- ❌ RabbitMQ 集成、Elasticsearch 集成、WebSocket、高德地图
- ❌ Swagger API 文档、数据库迁移 scripts
- ❌ PostGIS 空间查询
- ❌ 第三方登录、手机验证码登录
- ❌ 单元测试、CI/CD 流水线

## 许可证

MIT License
