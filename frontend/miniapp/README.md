# 五常同城小程序端

uni-app + Vue 3 实现的小程序端（支持 H5/微信小程序/支付宝小程序等多端编译）。

## 技术栈

- uni-app 3（Vue 3 + Vite）
- 多端编译：H5、微信小程序、支付宝小程序、App
- 后端 API：`/api/v1`（H5 走 vite proxy，小程序走绝对地址）

## 目录结构

```
src/
├── App.vue              # 全局入口（初始化默认地区）
├── main.js              # createSSRApp 入口
├── manifest.json        # uni-app 配置（appid、各平台配置）
├── pages.json           # 页面路由 + tabBar + 全局样式
├── api/
│   ├── request.js       # uni.request 封装（token + 地区头注入 + 错误处理）
│   ├── news.js          # 头条 API（列表/详情/搜索/点赞/分类/地区）
│   └── user.js          # 用户 API（登录/profile/退出）
└── pages/
    ├── index/
    │   └── index.vue    # 首页（地区切换 + 分类导航 + 最新头条）
    ├── news/
    │   ├── list.vue     # 头条列表（横向分类筛选 + 上拉加载 + 下拉刷新）
    │   └── detail.vue   # 头条详情（rich-text 富文本 + 点赞）
    ├── search/
    │   └── index.vue    # 全文搜索（走后端 /news/search）
    └── user/
        └── index.vue    # 我的（登录/退出 + 地区切换）
```

## 快速开始

```bash
# 1. 安装依赖
npm install

# 2. H5 开发模式（推荐先用 H5 调试）
npm run dev:h5
# 访问 http://localhost:5173

# 3. 微信小程序开发模式
npm run dev:mp-weixin
# 然后用微信开发者工具导入 dist/dev/mp-weixin 目录

# 4. 生产构建
npm run build:h5          # H5
npm run build:mp-weixin   # 微信小程序
```

## 后端 API 依赖

- `POST /api/v1/user/login` - 用户登录
- `GET /api/v1/user/profile` - 个人信息（需 JWT）
- `GET /api/v1/news` - 头条列表
- `GET /api/v1/news/:id` - 头条详情
- `GET /api/v1/news/search` - 全文搜索（ES 优先）
- `GET /api/v1/news/:id/like` + `POST /api/v1/news/:id/like` - 点赞
- `GET /api/v1/category` - 分类列表
- `GET /api/v1/region` - 地区列表

## 配置说明

- **小程序端**：需在 `src/api/request.js` 中把 `BASE_URL` 改为线上 HTTPS 域名，并在微信公众平台配置 request 合法域名
- **微信 appid**：在 `src/manifest.json` 的 `mp-weixin.appid` 填入自己的小程序 appid
- **默认地区**：武汉市（id=2，见后端 seed 数据）
- **token 存储**：`uni.setStorageSync('token', ...)`，登录后自动注入到请求头

## tabBar

底部导航 4 个：首页 / 头条 / 搜索 / 我的
