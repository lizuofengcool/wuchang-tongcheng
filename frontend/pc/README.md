# 五常同城 PC门户站

Next.js 14 App Router + TypeScript + Tailwind CSS 实现的 C 端门户站。

## 技术栈

- Next.js 14（App Router，SSR + ISR）
- React 18 + TypeScript 5
- Tailwind CSS 3
- 后端 API：`/api/v1`（开发环境由 `next.config.mjs` rewrites 代理）

## 目录结构

```
src/
├── app/
│   ├── layout.tsx              # 根布局（Header + Footer + RegionProvider）
│   ├── page.tsx                # 首页（最新头条 + 分类导航）
│   ├── globals.css             # 全局样式（Tailwind）
│   ├── news/
│   │   ├── page.tsx            # 头条列表（分页 + 分类筛选）
│   │   ├── [id]/
│   │   │   ├── page.tsx        # 头条详情
│   │   │   └── LikeButton.tsx  # 点赞按钮（客户端组件）
│   │   └── search/
│   │       └── page.tsx        # 全文搜索（走 /news/search，ES 优先）
│   └── category/
│       └── [id]/
│           └── page.tsx        # 分类详情页
├── components/
│   ├── Header.tsx              # 顶栏（logo + 导航 + 搜索框）
│   ├── Footer.tsx
│   ├── NewsCard.tsx            # 头条卡片
│   └── RegionSelector.tsx      # 地区切换器（客户端组件）
└── lib/
    ├── types.ts                # TS 类型定义
    ├── api.ts                  # 后端 API 封装（fetch + ISR 缓存）
    └── region.ts               # 地区上下文（React Context）
```

## 快速开始

```bash
# 1. 安装依赖
npm install

# 2. 配置环境变量
cp .env.example .env.local
# 编辑 .env.local，设置 BACKEND_URL（默认指向本地后端）

# 3. 启动开发服务器
npm run dev
# 访问 http://localhost:3000

# 4. 生产构建
npm run build && npm start
```

## 后端 API 依赖

PC门户调用以下后端接口（需后端启动）：

- `GET /api/v1/news` - 头条列表（支持 region_id/category_id/keyword/status 过滤）
- `GET /api/v1/news/:id` - 头条详情
- `GET /api/v1/news/search` - 全文搜索（ES 优先，DB 降级）
- `GET /api/v1/news/:id/like` - 点赞状态查询
- `POST /api/v1/news/:id/like` - 点赞/取消（需 JWT）
- `GET /api/v1/category` - 分类列表
- `GET /api/v1/region` - 地区列表

## 默认配置

- 端口：3000
- 默认地区：武汉市（id=2，见后端 seed 数据）
- ISR 缓存：列表页 60s，详情页 0s（实时浏览量），搜索页 30s，分类页 60s

## 登录说明

PC门户默认未登录浏览。点赞需登录，当前实现：
- 若 `localStorage.token` 存在则带上 Authorization 调用点赞 API
- 否则提示用户去管理后台 `/login` 登录

> 后续可在 PC门户实现独立的登录/注册页（对接 `/api/v1/user/login`）。

## Docker 部署

```bash
docker build -t wuchang-pc .
docker run -p 3000:3000 -e BACKEND_URL=http://backend:8080 wuchang-pc
```
