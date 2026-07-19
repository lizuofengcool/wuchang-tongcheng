# 全平台大厂级架构方案 v3（行业第一标准）

> **文档性质**：全平台架构设计文档（待用户审核）
> **版本**：v3（已融合 `docs/架构设计/建议.md` 全平台去重提纯方案）
> **创建日期**：2026-07-19
> **审核状态**：待审核（审核通过后方可进入 P0 编码）
> **依据**：
> - 项目根目录《产品需求文档-对标 dismall 完整功能补齐.md》
> - 用户原话："你给我做成分类里面的一个信息展示了吗？模块设置，收益，售价，各种功能都没有吗？"
> - 用户原话："你要把所有模块都考虑一遍，不要开发重复，不要有冲突"
> - 用户原话："先写文档，我审核"
> - 用户原话："你要做到全国最牛"
> - 用户原话："D:\kaifa\wuchang-tongcheng\docs\架构设计\建议.md 你再看看"（用户最新梳理）
> - 用户原话："看看是否是最佳方案，你要好好梳理，看看模块多了，后台怎么设计才是最佳"
> - 用户审核反馈（v3 方向已确认）：
>   - 按新版 11 中台完全重构 v3（+DIY 独立为前端能力中台 = 12 中台）
>   - 16 业务领域为顶层，30 功能模块为子模块（DIY 移出 = 15 垂直业务）
>   - ershou 在 P2 第一批次第一个
>   - 17 类去重提纯表完整纳入 v3 作为模块设计权威依据
>   - 后台采用五中心分离架构
>   - 11 中台 + 内部子域分层
>   - DIY 独立为前端能力中台
>   - P2 分 3 批次内部并行开发

---

## 一、设计目标与对标参考

### 1.1 设计目标（v3 战略升级）

把整个五常同城平台从"30 个零散功能模块"升级为**对标行业头部的完整同城生活大厂平台**，实现"单模块完全独立可管控、全平台中台复用、一次开发终身不重构"三大战略目标：

**业务能力**（覆盖点微+西瓜全部业务）：
1. 15 个垂直业务模块（涵盖分类信息/二手/招聘/房产/黄页114/到家/商城/头条/圈子/活动/婚恋/汽车/教育/装修/直播）
2. 每个垂直模块独立可开关、独立配置、独立权限、独立数据表、独立定时任务
3. 完整商业化能力：5 种收费模式（佣金/付费发布/付费置顶/求购/付费刷新）+ 担保交易 + 分账 + 分销
4. 多端覆盖：UniAppX（APP/微信/抖音/快手小程序/H5）+ Next16 PC门户 + Vue3 管理后台 + 商家独立后台

**大厂级能力**（v3 核心壁垒）：
5. **12 大通用中台**（含 DIY 前端能力中台），全平台共享，杜绝重复开发
6. **中台内部子域分层**（支付财务分 5 子域、风控审核分 3 子域、营销活动分 4 子域），避免单中台过重
7. **模块独立管控机制**：单模块开关、独立配置、独立权限、独立定时任务、独立数据表前缀、独立灰度
8. **接口抽象层**：业务模块仅依赖 Plugin 抽象接口，不依赖实现，后期可独立打包微服务
9. **分站独立运营**：每个城市每个模块拥有一套独立配置（佣金/价格/风控/发布限制/UI）
10. **金融级资金防护**：钱包/订单全链路事务、自动对账、多级分账、资金风控
11. **LBS 推荐系统**：用户行为埋点、画像计算、多路召回、实时热点加权
12. **运维灰度体系**：Prometheus+Grafana 监控、模块独立监控面板、CI/CD 流水线、定时任务调度中心
13. **后台五中心分离架构**：工作台/模块中心/中台中心/设置中心/数据中心
14. **三层后台体系**：平台总后台 + 城市分站后台 + 商家独立后台

### 1.2 点微+西瓜两大插件底层通病（v3 完全规避）

点微、西瓜全部基于 **Discuz 论坛耦合 PHP 插件**，存在 6 大致命短板，v3 架构完全规避：

| # | 通病 | v3 规避方案 |
|---|------|------------|
| 1 | 强耦合：所有插件依赖「分类信息主插件」，不能独立启停、独立部署 | 12 中台 + 15 垂直业务完全独立，Plugin 抽象接口解耦 |
| 2 | 重复造轮子：用户、支付、消息、IM、分站、地图、审核每个插件单独写一套 | 12 大通用中台全局复用，杜绝重复 |
| 3 | 无金融级资金体系：无分布式事务、自动对账、分账、风控 | 本地消息表+定时对账+多级分账+资金风控 |
| 4 | LBS 性能差：MySQL 空间索引，多城市百万数据查询卡顿 | PostgreSQL16 + PostGIS 地理索引，毫秒级查询 |
| 5 | 无推荐/数据中台：纯静态列表，无个性化流量分发 | 用户行为埋点+画像+多路召回+排序 |
| 6 | 多端割裂：小程序、APP、PC、H5 代码完全分离 | UniAppX 一套代码输出 5 端 + Next16 PC |

### 1.3 对标参考与取舍

| 对标来源 | 借鉴能力 | 是否采纳 | 取舍理由 |
|---------|----------|----------|----------|
| dismall 点微全部插件 | 17 类业务模块全部能力 | ✅ 全部采纳 | 与本项目商业模式完全一致 |
| dismall 西瓜全部插件 | 17 类业务模块全部能力 | ✅ 全部采纳 | 与本项目技术栈一致 |
| 闲鱼 | 鱼塘社区 + 芝麻信用 + 短视频商品 | ❌ 鱼塘 / ⚠️ 信用走自研 / ✅ 短视频 | 鱼塘与 quan 重复；信用应独立通用模块 |
| 转转 | 367 道验机 + 质检报告 | ⚠️ 预留字段不实做 | 验机重运营，本项目先做 C2C |
| 58 同城 | 付费置顶 + 多位置顶 + 多规格筛选 | ✅ 全部采纳 | 与 dismall 一致 |
| 瓜子二手车 | 检测报告 + 过户 | ❌ 不采纳 | 车辆能力归 car 模块 |
| 找靓机 | 392 项质检 + 365 天质保 | ❌ 不采纳 | 重运营模式，本项目先做平台撮合 |
| 闲鱼/转转交易中台 | 金融级资金防护 + 自动对账 + 分账 | ✅ 采纳设计，MVP 简化 | 资金安全必须，但分布式事务 MVP 用本地消息表 |
| 大厂推荐系统 | LBS 多路召回 + 画像 | ✅ 采纳 MVP 版本 | Go 自建轻量级，Flink 后期再考虑 |
| 有赞/微盟/Shopify | 后台架构 + 三层后台体系 | ✅ 五中心分离 + 三层后台 | 大型 SaaS 标准架构 |

### 1.4 v3 关键技术决策（避免过度设计）

| 决策点 | v3 方案 | 理由 |
|--------|---------|------|
| 分布式事务 | MVP 用本地消息表+定时对账；后期接入 DTM（Go 原生） | Seata 是 Java 生态不适用；本地消息表足够支撑 MVP |
| 推荐系统 | Go 自建轻量级批处理（cron + goroutine）+ Redis 缓存 | Flink 运维成本高，初期数据量不需要 |
| K8s 部署 | 仅做接口抽象层，部署用 Docker Compose | 用户规则"模块化单体足够支撑10万日活，禁止过度设计" |
| WebSocket 集群 | MVP 单实例，接口预留集群扩展 | 单实例支持 1 万连接，初期足够 |
| 消息分片 | MVP 单表，索引优化；后期按 region_id+月份分表 | 单表百万级 PostgreSQL 无压力 |
| 监控告警 | Prometheus + Grafana 接入（轻量） | 不上 ELK 全套，按需扩展 |
| 后台架构 | 五中心分离 + 三层后台体系 | 27 模块不分离会迷失 |
| 中台分层 | 11 中台 + 内部子域分层 | 避免单中台过重，1 万行代码上限 |
| DIY 定位 | 独立为前端能力中台（第 12 个中台） | DIY 是能力型模块不是业务模块 |

### 1.5 通用能力抽取原则

依据用户原话"不要开发重复，不要有冲突"，以下能力**必须独立成通用中台模块**，跨 15 个垂直业务模块共享：

| 通用能力 | 归属中台 | 内部子域 | 服务业务 |
|---------|---------|---------|---------|
| 用户账号/实名/信用/会员 | 用户账号中台 | account/profile/credit/vip | 全部 |
| 支付/钱包/订单/退款/分账 | 支付财务中台 | pay/wallet/order/refund/settle | 全部付费场景 |
| 私聊/群聊/客服/通知 | IM消息中台 | chat/group/service/notify | 全部 |
| 商家入驻/店铺/CRM/结算 | 商家商户中台 | shop/crm/staff/settle | 全部 B 端 |
| 二级分销/分站分成/合伙人 | 分销合伙人中台 | distribution/station/partner | 全部付费场景 |
| 广告位/优惠券/签到/活动 | 营销活动中台 | ad/coupon/sign/activity | 全部 |
| 信用/敏感词/举报/审核 | 风控审核中台 | credit/audit/report | 全部 UGC |
| LBS 定位/附近/路线/围栏 | LBS地图中台 | location/nearby/route/fence | 全部 LBS 业务 |
| AI 发帖/润色/审核/客服 | AI智能中台 | generate/polish/audit/chat | 全部 |
| 多租户/分站/域名/权限 | 多租户分站中台 | tenant/domain/staff/config | 全部 |
| 图片/视频/转码/CDN | 素材存储中台 | upload/transcode/cdn/search | 全部 |
| DIY 拖拽/页面/专题/店铺 | 前端能力中台 | page/section/shop/topic | 全部前端 |

### 1.6 对比点微/西瓜核心优势（8 项行业第一壁垒）

依据用户原话"对比点微/西瓜，本方案核心优势（做到行业第一的核心壁垒）"，v3.2 在 1.1 节大厂级能力 14 项 + 1.2 节通病规避 6 项的基础上，集中对比 8 项核心壁垒：

| # | 核心优势 | 点微/西瓜短板 | v3.2 实现 |
|---|---------|--------------|-----------|
| 1 | **彻底解耦，无插件依赖** | Discuz 插件必须依赖「分类信息主插件」，单点故障，一个模块崩全站挂 | 12 中台 + 15 垂直业务完全独立，Plugin 抽象接口解耦，单模块可独立启停/单独部署/单独打包 |
| 2 | **金融级资金安全** | 无分布式事务、自动对账、分账、风控，资损风险极高 | 本地消息表+定时对账+多级分账+资金风控+RedLock 5实例+2FA 二次验证+JWT 刷新+行为分析防刷（v3.2 补强 8 项金融安全） |
| 3 | **LBS 性能碾压** | MySQL 空间索引，多城市百万数据查询卡顿 | PostgreSQL16 + PostGIS 地理索引，毫秒级附近查询，百万数据无压力 |
| 4 | **多城市精细化运营** | 分站能力薄弱，仅简单数据隔离 | 每城市每模块独立定价/佣金/风控/活动/展示规则，分站一键复制配置，独立域名/权限/数据隔离 |
| 5 | **统一多端一套代码** | 小程序/H5/APP 代码完全割裂，维护成本极高 | UniAppX 一套代码输出 5 端（APP/微信/抖音/快手/H5）+ Next16 PC，统一组件全业务复用（详见 3.4.1） |
| 6 | **可无限扩展、永不重构** | PHP Discuz 无法拆分微服务，用户量上涨必须全部重写 | 插件化 Go 架构，后期可将任意垂直模块拆为独立微服务，业务代码零修改，Istio 服务网格预留（详见第二十三章） |
| 7 | **全模块数据打通，统一推荐** | 各插件数据隔离，无法跨业务流量分发 | 用户行为全域采集，同城头条/二手/招聘/商家个性化推荐，多路召回+实时热点加权（详见第十章） |
| 8 | **完整独立数据看板** | 仅简单数据列表，无可视化分析 | 每模块独立运营大屏，分城市统计 GMV/付费/曝光/咨询，跨城市运营报表（详见 7.6 数据中心） |

**核心壁垒总结**：v3.2 通过"中台复用 + 垂直解耦 + 金融防护 + 多端统一 + 数据中台 + 灰度运维"六大维度，构建点微/西瓜无法逾越的行业第一壁垒，且 25 项 v3.2 补强细节（第二十四章）进一步在金融安全、运维完善度、大厂标准完整性三个维度显著超越。

---

## 二、点微+西瓜去重提纯（17 类择优整合方案）

> **权威性声明**：本章 17 类去重提纯表是 v3 各垂直业务模块字段/功能设计的**权威依据**。每个模块设计时必须对照本章"最终整合方案"列，确保覆盖点微+西瓜全部精华能力。

### 2.1 重复模块合并+择优取舍表（17 类）

| # | 重复业务模块 | 点微独有优势 | 西瓜独有优势 | 最终整合方案（合并精华） |
|---|------------|------------|------------|----------------------|
| 1 | 分类信息（核心底座） | 自定义分类模型、子站点独立分成、金币转发、PC独立页面、渠道二维码、批量导出信息 | AI发布辅助、AI润色、自定义发布表单、一键生成DM报纸、求购专区、朋友圈分享裂变 | 统一「分类信息核心模块」，融合全部能力；内置自定义类目、付费发布、AI辅助发布、DM生成、求购、分站独立定价、金币激励、PC/小程序双端独立页面 |
| 2 | 同城二手交易 | 担保交易完整流程、商品权重排序、视频加权、商家二手店铺体系 | 定金/全款双模式、以图搜图、成色多级筛选、隐私号码保护买家卖家 | 二手独立垂直模块：担保交易+定金模式、以图搜图、隐私虚拟号、视频加权、商家店铺、多维度筛选、信用分管控 |
| 3 | 同城招聘 | PC独立招聘站、企业认证、简历投递、岗位置顶套餐、零工专区 | 零工驿站、临时工/日结/包工、招工海报、工人信用评级、招工签到激励 | 招聘+零工合并模块：企业端+零工双体系、简历投递、日结包工、工人信用、招聘海报、独立PC门户 |
| 4 | 同城房产 | 新房/二手房/租房/商铺、房源核验、经纪人店铺、看房预约、房产PC站 | 租房朋友圈推广、房源DIY详情、房源自动下架、合租筛选 | 房产垂直模块：新房/二手/商铺/合租、经纪人店铺、看房预约、DIY详情、房源自动过期、独立SEO PC页面 |
| 5 | 同城商圈/114黄页 | 商家CRM、商家后台、小票打印、商家会员、店铺优惠券、团购 | 商家认领、电话批量导入、店铺动态、商家红包、堂食点餐、餐桌码、商家DIY主页 | 统一「商家黄页中台」：商家入驻/认领、CRM、点餐、团购、优惠券、红包、小票打印、自定义店铺装修、批量导入商家电话 |
| 6 | 同城商城/拼团/砍价 | 商城完整订单、拼团、砍价、抢购、礼品卡、会员黑卡、核销体系 | 积分兑换商城、礼品卡、阶梯拼团、到店自提、分销叠加优惠 | 电商统一中台：商城+拼团+砍价+抢购+礼品卡+积分商城，核销、多优惠叠加、商家独立结算 |
| 7 | 同城头条/资讯/圈子 | 同城头条、专题、活动、直播、图文投稿、本地爆料、微信群 | 同城DM报纸、同城微信群、投票、影音视频、DIY专题页面、传单生成 | 内容资讯中台：头条爆料、专题、活动、直播、社群、投票、DM传单、视频发布、拖拽DIY页面 |
| 8 | 同城合伙人分销 | 二级分销、分站分成、推广渠道统计、佣金结算、等级合伙人 | 合伙人加盟付费、任务推广、朋友圈推广佣金、线下地推物料 | 统一分销中台：二级分销+城市分站分成、付费合伙人、推广任务、渠道统计、自动佣金结算 |
| 9 | 同城预约/到家服务 | 上门家政、维修、预约服务、预约订单状态机、商家预约管理 | 到家需求发布、师傅接单、认养预定、定金预约、服务点评 | 本地服务模块：预约+到家+认养预定、师傅接单、定金全款、服务点评、上门履约流程 |
| 10 | 同城婚恋相亲 | 相亲会员、资料认证、红娘管理、相亲活动、匹配推荐 | 相亲付费解锁、线下相亲会、相亲签到、私信打赏、资料隐私保护 | 婚恋独立模块：实名认证、红娘、线上匹配、线下活动、付费解锁、隐私号码、打赏互动 |
| 11 | 同城营销工具 | 签到、抽奖、红包、会员卡、任务中心、渠道推广、优惠券 | 积分充值、积分兑换、投票、引导关注、广告位DIY、一键分享、手机广告 | 统一营销中台：签到、积分、抽奖、红包、会员卡、任务、优惠券、投票、引导关注、全端广告位可视化配置 |
| 12 | 多端小程序 | 微信/抖音/快手三端小程序、小程序订阅消息、模版消息 | 马甲APP、千帆APP、小程序一键分享、订阅消息批量推送 | 多端统一底层：UniAppX一套代码输出微信/抖音/快手小程序+H5+APP，统一消息推送、分享组件 |
| 13 | IM聊天/消息通知 | 订单消息、站内信、私聊、客服聊天、模版消息 | 同城聊天、社群聊天、消息批量群发、消息模板自定义 | 统一IM消息中台：私聊/群聊/客服、订单通知、批量群发、自定义消息模板、多渠道推送（APP/小程序/短信） |
| 14 | 分站/多城市管理 | 子站点独立后台、分站独立佣金、分站独立广告、分站数据隔离 | 分站一键复制配置、分站独立域名、分站独立运营人员权限 | 多租户分站中台：无限城市分站、独立配置、独立域名、独立权限、配置一键复制、数据隔离 |
| 15 | AI辅助工具 | AI助手、AI自动生成内容 | AI发布润色、AI图文生成、AI审核、多模型兼容（文心/通义/DeepSeek） | AI中台统一封装：AI发帖、AI润色、AI内容审核、AI图文生成，全业务模块可调用 |
| 16 | 支付/财务 | 微信/支付宝/余额支付、订单支付、退款、支付中心、提现 | 积分充值、礼品卡支付、批量提现、财务导出、支付模版 | 金融支付中台：多渠道支付、余额、积分、礼品卡、退款、提现、自动对账、财务报表、分账 |
| 17 | 举报/风控/认证 | 用户/商家实名认证、举报中心、敏感词过滤、违规下架 | 多层级处罚规则、防灌水、重复发布拦截、AI图片视频审核 | 风控审核中台：实名认证、AI图文审核、敏感词DFA、举报工单、分级处罚、防刷风控规则 |

### 2.2 完全剔除的冗余功能（4 类，无业务价值，不开发）

1. **Discuz 论坛底层**：帖子灌水、论坛板块、老旧触屏模板（完全抛弃，不用 PHP 论坛架构）
2. **多套马甲 APP 底层**：老旧第三方 APP 适配（统一 UniAppX 原生 APP，不再兼容千帆/小云）
3. **重复图片转换、单一广告小插件、英文防灌水**等边角工具（统一素材中台处理图片、统一风控中台防刷）
4. **分散独立短信、二维码、打印小插件**（统一封装到基础中台，全局复用）

### 2.3 去重提纯表的权威性

v3 各垂直业务模块设计时必须：
1. 对照本章"最终整合方案"列，确保覆盖所有能力
2. 对照点微独有优势和西瓜独有优势列，确保不遗漏任一开发商的能力
3. 模块字段设计、接口设计、管理后台页面设计都需对照本章

---

## 三、大厂四层架构

### 3.1 四层架构总览

```
┌─────────────────────────────────────────────────────────────────┐
│  接入层（无业务）                                                │
│  Nginx 网关 + Gin API 网关（模块路由动态注册、限流熔断、灰度）   │
└────────────────────────────┬────────────────────────────────────┘
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│  多端交付层（5 端统一）                                          │
│  UniAppX(APP/微信/抖音/快手小程序/H5) | Next16 PC门户 |          │
│  Vue3 平台总后台 | Vue3 城市分站后台 | Vue3 商家独立后台          │
└────────────────────────────┬────────────────────────────────────┘
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│  业务层（核心解耦，中台+垂直业务双拆分）                         │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  12 大通用中台（全局唯一，全业务复用，杜绝重复开发）       │  │
│  │  用户账号 / 支付财务 / IM消息 / 商家商户 / 分销合伙人 /    │  │
│  │  营销活动 / 风控审核 / LBS地图 / AI智能 / 多租户分站 /     │  │
│  │  素材存储 / 前端能力(DIY)                                  │  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  15 个独立垂直业务模块（每个完全独立可配置、可单独关停）   │  │
│  │  分类信息 / 二手 / 招聘 / 房产 / 黄页114 /                │  │
│  │  到家预约 / 商城电商 / 头条资讯 / 圈子社群 / 活动 /       │  │
│  │  婚恋 / 汽车 / 教育 / 装修 / 直播                         │  │
│  └──────────────────────────────────────────────────────────┘  │
└────────────────────────────┬────────────────────────────────────┘
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│  数据基础设施层（全局共享无业务）                                │
│  PG16+PostGIS / Redis7 集群 / Elasticsearch8 / RabbitMQ /        │
│  MinIO(七牛云OSS+CDN) / Prometheus+Grafana / 定时任务调度中心    │
└─────────────────────────────────────────────────────────────────┘
```

### 3.2 模块依赖关系图

```
垂直业务模块（15个）
   │
   │ 仅依赖抽象接口，不直接依赖实现
   ▼
12 大通用中台（接口抽象层）
   │
   │ 共享基础设施
   ▼
数据基础设施层
```

**关键原则**：
1. 垂直业务模块之间**互不依赖**（ershou 不依赖 job，job 不依赖 fang）
2. 垂直业务模块**仅依赖**通用中台的抽象接口（plugin.Plugin 接口）
3. 通用中台之间**最小依赖**（如支付财务中台依赖用户账号中台的用户信息）
4. 所有模块共享数据基础设施层

### 3.3 接入层设计

1. **Nginx 网关**：SSL 终止、静态资源、Gzip、限流
2. **Gin API 网关**：
   - 模块路由动态注册（基于 plugin.Plugin 接口）
   - 限流熔断（按模块独立配置）
   - 鉴权（JWT + 角色权限）
   - 灰度路由（基于城市/用户ID/模块版本）
3. **模块开关拦截**：模块关闭后，API 网关自动拦截该模块所有 API，返回"未开通"提示

### 3.4 多端交付层设计

#### 3.4.1 UniAppX 用户端（5 端统一，核心）

一套代码编译输出 **5 端**：微信小程序、抖音小程序、快手小程序、H5、安卓/iOS 原生 APP。

| 能力 | 设计 |
|------|------|
| **分包隔离** | 页面按业务模块分包：`ershou/job/house/shop/toutiao`，删除模块直接删除对应分包，不影响其他页面 |
| **动态路由控制** | 后台关闭模块后，前端自动隐藏底部导航 Tab、分类入口、推荐卡片（启动时拉取 `modules` 开关表） |
| **统一全局组件** | 封装地图 LBS、图片上传、IM 聊天、支付弹窗、筛选组件、地址选择器，全业务复用 |
| **UTS 原生语法** | APP 端性能对标原生开发，无 uni-app 旧版卡顿问题；小程序端走 Vue3 编译 |
| **角色切换** | 一套 App 内支持 10 角色切换（用户/商家/骑手/代理/平台/审核/运营/VIP/ clerk/访客），TabBar 动态切换 |
| **离线缓存** | 首页信息流 + 我的页面关键数据本地缓存，弱网可用 |
| **多端分发** | manifest.json 多端分发配置，按端差异化配置权限、SDK、UI |

#### 3.4.2 Next.js 16 PC 门户（SSR+PPR SEO 优化）

| 能力 | 设计 |
|------|------|
| **SSR+PPR** | 服务端渲染 + Partial Prerendering，首屏毫秒级，搜索引擎收录远超 Discuz 静态页 |
| **模块独立 PC 站** | 同城分类、头条、房产、招聘独立 PC 站点，每模块独立路由分组（`/ershou` `/news` `/fang` `/job`） |
| **可单独屏蔽** | 后台关闭某模块后，PC 端对应路由分组同步屏蔽，返回 404 或未开通提示 |
| **ES 全文检索** | 统一调用 Elasticsearch，PC 与 APP 数据互通，搜索结果一致 |
| **SEO 元信息** | 每模块独立 title/description/keywords 模板，自动生成 sitemap.xml、robots.txt |
| **结构化数据** | 商品/文章/招聘信息输出 Schema.org JSON-LD，提升搜索结果富文本展示 |
| **服务端组件** | React19 Server Components，数据获取在服务端，减少客户端 JS 体积 |

#### 3.4.3 Vue3 管理后台（三层后台体系）

三层后台体系（详见第七章 7.7 节）：

| 后台类型 | 端 | 核心职责 |
|---------|-----|---------|
| **平台总后台** | Vue3 + Element Plus | 全局中台配置、模块开关、分站管理、全平台数据大盘（五中心分离架构） |
| **城市分站后台** | Vue3 + Element Plus | 当前城市独立配置、本地数据、本地活动运营（数据隔离到 region_id） |
| **商家独立后台** | Vue3 + Element Plus（适配手机/PC） | 店铺管理、订单、商品、营销、结算（仅自身店铺数据权限） |

三后台共用同一套 Vue3 组件库（Element Plus + 自研业务组件），通过角色权限区分入口和数据范围。

### 3.5 业务层设计

#### 3.5.1 12 大通用中台

每个中台：
- 独立 Go package（`backend/internal/modules/<中台名>/`）
- 独立数据表前缀（如 `user_*` / `pay_*` / `im_*`）
- 独立配置项（每城市每中台独立配置）
- 独立监控面板（Prometheus 指标独立）
- 独立定时任务队列
- 内部子域分层（重负载中台拆分子域）

#### 3.5.2 15 个垂直业务模块

每个垂直模块：
- 独立 Go package（`backend/internal/modules/<业务名>/`）
- 独立数据表前缀（如 `ers_*` / `job_*` / `house_*`）
- 独立配置项（每城市每模块独立配置）
- 独立权限域（模块专员仅管理单一模块）
- 独立定时任务队列
- 独立灰度发布
- 完整复用 12 大通用中台能力

### 3.6 数据基础设施层设计

| 基础设施 | 用途 | 部署 |
|---------|------|------|
| PostgreSQL 16 + PostGIS | 主数据库 + LBS 地理查询 | Docker |
| Redis 7 集群 | 缓存 + 地理位置热点 + 队列 + 会话 | Docker |
| Elasticsearch 8 | 全文检索 + 分类/头条/商家搜索 | Docker |
| RabbitMQ | 异步消息 + 订单/通知/统计 | Docker |
| MinIO / 七牛云 OSS+CDN | 图片视频存储 + 转码 + CDN | 云服务 |
| Prometheus + Grafana | 全链路监控 + 模块独立面板 | Docker |
| 定时任务调度中心 | 模块过期/结算/关闭订单统一调度 | Docker |

---

## 四、12 大通用中台详细设计

### 4.1 用户账号中台（user）

**职责**：注册/登录/第三方授权/手机号/实名认证/信用分/会员等级/隐私设置

**内部子域**：
- account：注册/登录/密码/第三方授权
- profile：用户资料/头像/昵称/个人简介
- credit：信用分（与风控审核中台 credit 子域联动）
- vip：会员等级/特权/积分

**核心表**：
```sql
-- user_accounts 账号表（已存在，扩展）
-- user_profiles 资料表
-- user_credits 信用分流水分表
-- user_vip_levels 会员等级表
-- user_vip_orders 会员订单表
-- user_oauths 第三方授权表（微信/抖音/支付宝）
-- user_realnames 实名认证表
```

**核心接口**：
- POST /api/v1/user/register - 注册
- POST /api/v1/user/login - 登录
- GET /api/v1/user/profile - 获取资料
- PUT /api/v1/user/profile - 更新资料
- POST /api/v1/user/realname - 实名认证
- GET /api/v1/user/credit - 信用分查询
- POST /api/v1/user/vip/order - 购买会员

### 4.2 支付财务中台（pay）- 内部 5 子域

**职责**：微信/支付宝/余额/积分/礼品卡支付、订单、退款、提现、分账、自动对账、资金风控、财务报表

**内部 5 子域**：
- **pay**：支付渠道抽象（微信/支付宝/余额/积分/礼品卡）
- **wallet**：钱包余额（充值/提现/冻结/解冻）
- **order**：订单中心（状态机/幂等/超时关单）
- **refund**：退款中心（退款流程/仲裁）
- **settle**：结算分账（多级分账/自动对账/资金风控）

**核心表**：
```sql
-- pay 子域
CREATE TABLE pay_orders (
  id BIGSERIAL PRIMARY KEY,
  region_id BIGINT NOT NULL,
  out_trade_no VARCHAR(64) UNIQUE NOT NULL,      -- 业务幂等号
  biz_module VARCHAR(50) NOT NULL,                -- ershou/mall/yuyue/...
  biz_id BIGINT NOT NULL,                          -- 业务订单ID
  user_id BIGINT NOT NULL,
  amount DECIMAL(12,2) NOT NULL,
  channel VARCHAR(20) NOT NULL,                    -- wechat/alipay/balance/point/giftcard
  status INT NOT NULL DEFAULT 0,                   -- 0待支付 1已支付 2已关闭 3已退款
  paid_at TIMESTAMP,
  expired_at TIMESTAMP,                            -- 30分钟超时
  created_at TIMESTAMP
);
CREATE INDEX idx_pay_orders_user ON pay_orders(user_id, status);
CREATE INDEX idx_pay_orders_biz ON pay_orders(biz_module, biz_id);

-- wallet 子域
CREATE TABLE wallet_accounts (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT UNIQUE NOT NULL,
  balance DECIMAL(12,2) DEFAULT 0,                 -- 可用余额
  frozen DECIMAL(12,2) DEFAULT 0,                  -- 下单冻结
  frozen_after_sale DECIMAL(12,2) DEFAULT 0,       -- 售后冻结
  frozen_violation DECIMAL(12,2) DEFAULT 0,        -- 违规冻结
  total_recharge DECIMAL(12,2) DEFAULT 0,
  total_withdraw DECIMAL(12,2) DEFAULT 0,
  updated_at TIMESTAMP
);

CREATE TABLE wallet_transactions (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  type VARCHAR(20) NOT NULL,                       -- recharge/withdraw/freeze/unfreeze/transfer
  amount DECIMAL(12,2) NOT NULL,
  balance_after DECIMAL(12,2) NOT NULL,
  biz_module VARCHAR(50) DEFAULT '',
  biz_id BIGINT DEFAULT 0,
  remark VARCHAR(200) DEFAULT '',
  created_at TIMESTAMP
);
CREATE INDEX idx_wallet_tx_user ON wallet_transactions(user_id, created_at);

-- order 子域
CREATE TABLE order_orders (
  id BIGSERIAL PRIMARY KEY,
  region_id BIGINT NOT NULL,
  order_no VARCHAR(64) UNIQUE NOT NULL,
  biz_module VARCHAR(50) NOT NULL,                 -- ershou/mall/yuyue
  buyer_id BIGINT NOT NULL,
  seller_id BIGINT NOT NULL,
  amount DECIMAL(12,2) NOT NULL,
  status INT NOT NULL DEFAULT 0,                   -- 11状态机
  paid_at TIMESTAMP, shipped_at TIMESTAMP, received_at TIMESTAMP,
  settled_at TIMESTAMP, closed_at TIMESTAMP,
  created_at TIMESTAMP
);
CREATE INDEX idx_order_buyer ON order_orders(buyer_id, status);
CREATE INDEX idx_order_seller ON order_orders(seller_id, status);

-- refund 子域
CREATE TABLE refund_orders (
  id BIGSERIAL PRIMARY KEY,
  order_id BIGINT NOT NULL,
  refund_no VARCHAR(64) UNIQUE NOT NULL,
  amount DECIMAL(12,2) NOT NULL,
  reason VARCHAR(500),
  status INT NOT NULL DEFAULT 0,                   -- 0申请中 1卖家同意 2退款中 3已退款 4拒绝 5仲裁中
  evidence_urls TEXT,                              -- 买家举证
  seller_response TEXT,                            -- 卖家答辩
  admin_decision TEXT,                             -- 平台判决
  created_at TIMESTAMP
);

-- settle 子域
CREATE TABLE settle_rules (
  id BIGSERIAL PRIMARY KEY,
  region_id BIGINT NOT NULL,
  biz_module VARCHAR(50) NOT NULL,
  platform_rate DECIMAL(5,4) NOT NULL,             -- 平台佣金
  station_rate DECIMAL(5,4) DEFAULT 0,             -- 分站分成
  partner_rate DECIMAL(5,4) DEFAULT 0,             -- 合伙人分销
  seller_rate DECIMAL(5,4) NOT NULL,               -- 卖家所得
  updated_at TIMESTAMP
);

CREATE TABLE settle_records (
  id BIGSERIAL PRIMARY KEY,
  order_id BIGINT NOT NULL,
  platform_amount DECIMAL(12,2) NOT NULL,
  station_amount DECIMAL(12,2) NOT NULL,
  partner_amount DECIMAL(12,2) NOT NULL,
  seller_amount DECIMAL(12,2) NOT NULL,
  settled_at TIMESTAMP
);
```

**金融级资金防护**（详见第九章）：
- 本地消息表+定时对账（替代 Seata）
- 多级分账规则（平台/分站/合伙人/卖家）
- 资金风控规则（单日大额/高频/提现分级审核）
- 三种冻结池（frozen/frozen_after_sale/frozen_violation）

### 4.3 IM消息中台（im）

**职责**：私聊/群聊/客服/系统通知/小程序订阅消息/APP推送/短信兜底/消息模板

**内部子域**：
- chat：私聊（买卖双方议价、图片消息、商品卡片）
- group：群聊（同城社群）
- service：客服（商家客服）
- notify：系统通知（订单通知、支付通知）

**核心表**：
```sql
CREATE TABLE im_conversations (
  id BIGSERIAL PRIMARY KEY,
  type VARCHAR(20) NOT NULL,                       -- chat/group/service
  biz_module VARCHAR(50) DEFAULT '',               -- ershou/mall
  biz_id BIGINT DEFAULT 0,
  created_at TIMESTAMP
);

CREATE TABLE im_messages (
  id BIGSERIAL PRIMARY KEY,
  conversation_id BIGINT NOT NULL,
  sender_id BIGINT NOT NULL,
  msg_type VARCHAR(20) NOT NULL,                   -- text/image/card/video/voice
  content TEXT,
  extra JSON,                                      -- 商品卡片/订单卡片
  read_at TIMESTAMP,
  recalled_at TIMESTAMP,
  created_at TIMESTAMP
);
CREATE INDEX idx_im_msg_conv ON im_messages(conversation_id, created_at);

CREATE TABLE im_notify_templates (
  id BIGSERIAL PRIMARY KEY,
  code VARCHAR(50) UNIQUE NOT NULL,
  title VARCHAR(100),
  content TEXT,
  channels VARCHAR(100),                           -- app/mp/sms
  created_at TIMESTAMP
);
```

**架构**：基于自研 WebSocket（复用 `backend/internal/pkg/ws/hub.go`），MVP 单实例，接口预留集群扩展。

### 4.4 商家商户中台（merchant）

**职责**：入驻/认领/店铺管理/CRM/商家权限/商家结算/商家营销工具

**核心表**：
```sql
CREATE TABLE merchant_shops (
  id BIGSERIAL PRIMARY KEY,
  region_id BIGINT NOT NULL,
  owner_id BIGINT NOT NULL,
  name VARCHAR(100),
  logo VARCHAR(500),
  intro TEXT,
  category_id BIGINT,                              -- 主营类目
  status INT NOT NULL DEFAULT 0,                   -- 0审核中 1正常 2停用
  credit_score INT DEFAULT 100,
  level INT DEFAULT 1,
  settled_at TIMESTAMP,
  created_at TIMESTAMP
);

CREATE TABLE merchant_staff (
  id BIGSERIAL PRIMARY KEY,
  shop_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  role VARCHAR(20),                                -- owner/manager/clerk
  permissions JSON,
  created_at TIMESTAMP
);

CREATE TABLE merchant_settles (
  id BIGSERIAL PRIMARY KEY,
  shop_id BIGINT NOT NULL,
  period VARCHAR(20),                              -- 2026-07
  total_amount DECIMAL(12,2),
  platform_fee DECIMAL(12,2),
  shop_amount DECIMAL(12,2),
  settled_at TIMESTAMP
);
```

**说明**：所有"经纪人/企业/师傅/商家"角色统一复用此中台（房产经纪人、招聘企业、二手商家、到家师傅都是 merchant）。

### 4.5 分销合伙人中台（distribution）

**职责**：二级分销/城市分站分成/推广渠道统计/佣金自动结算/付费合伙人等级

**核心表**：
```sql
CREATE TABLE distribution_partners (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  level INT DEFAULT 1,                             -- 1普通 2高级 3城市合伙人
  parent_id BIGINT DEFAULT 0,                      -- 上级
  region_id BIGINT,
  commission_rate DECIMAL(5,4) DEFAULT 0,
  joined_at TIMESTAMP
);

CREATE TABLE distribution_channels (
  id BIGSERIAL PRIMARY KEY,
  partner_id BIGINT NOT NULL,
  code VARCHAR(50) UNIQUE,
  name VARCHAR(100),
  click_count INT DEFAULT 0,
  register_count INT DEFAULT 0,
  order_count INT DEFAULT 0,
  created_at TIMESTAMP
);

CREATE TABLE distribution_commissions (
  id BIGSERIAL PRIMARY KEY,
  partner_id BIGINT NOT NULL,
  order_id BIGINT NOT NULL,
  amount DECIMAL(12,2),
  level INT,                                       -- 1一级 2二级
  status INT DEFAULT 0,                            -- 0待结算 1已结算
  settled_at TIMESTAMP,
  created_at TIMESTAMP
);
```

### 4.6 营销活动中台（marketing）- 内部 4 子域

**职责**：广告位/优惠券/签到/活动/积分/任务中心

**内部 4 子域**：
- **ad**：广告位（首页Banner/列表置顶/详情广告）
- **coupon**：优惠券（满减/折扣/兑换）
- **sign**：签到（每日签到/连续签到奖励/积分）
- **activity**：活动（营销活动配置）

**核心表**：
```sql
-- ad 子域
CREATE TABLE ad_positions (
  id BIGSERIAL PRIMARY KEY,
  region_id BIGINT,
  position_code VARCHAR(50),                       -- home_banner/list_top/detail_banner
  title VARCHAR(100),
  image_url VARCHAR(500),
  link_url VARCHAR(500),
  start_at TIMESTAMP,
  end_at TIMESTAMP,
  status INT DEFAULT 1,
  created_at TIMESTAMP
);

-- coupon 子域
CREATE TABLE coupons (
  id BIGSERIAL PRIMARY KEY,
  region_id BIGINT,
  title VARCHAR(100),
  type VARCHAR(20),                                -- discount/reduce/exchange
  amount DECIMAL(12,2),
  threshold DECIMAL(12,2),
  total_count INT,
  received_count INT,
  start_at TIMESTAMP,
  end_at TIMESTAMP,
  created_at TIMESTAMP
);

CREATE TABLE user_coupons (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  coupon_id BIGINT NOT NULL,
  status INT DEFAULT 0,                            -- 0未使用 1已使用 2已过期
  used_at TIMESTAMP,
  created_at TIMESTAMP
);

-- sign 子域
CREATE TABLE sign_records (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  sign_date DATE,
  continuous_days INT,
  reward_points INT,
  created_at TIMESTAMP
);

-- activity 子域
CREATE TABLE activities (
  id BIGSERIAL PRIMARY KEY,
  region_id BIGINT,
  title VARCHAR(100),
  type VARCHAR(20),                                -- kanjia/pintuan/choujiang
  start_at TIMESTAMP,
  end_at TIMESTAMP,
  config JSON,
  status INT DEFAULT 1,
  created_at TIMESTAMP
);
```

### 4.7 风控审核中台（risk）- 内部 3 子域

**职责**：实名认证/AI图文视频审核/DFA敏感词/举报工单/违规分级处罚/防刷

**内部 3 子域**：
- **credit**：信用分（用户/商家信用分流水）
- **audit**：内容审核（AI图文视频审核 + DFA敏感词 + 人工审核）
- **report**：举报工单（用户举报 + 平台处罚 + 申诉）

**核心表**：
```sql
-- credit 子域
CREATE TABLE credit_scores (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  score INT DEFAULT 100,
  level INT DEFAULT 3,                             -- 1极差 2差 3正常 4优 5极优
  updated_at TIMESTAMP
);

CREATE TABLE credit_logs (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  delta INT NOT NULL,                              -- 正加分负减分
  reason VARCHAR(200),
  biz_module VARCHAR(50),
  biz_id BIGINT,
  created_at TIMESTAMP
);

-- audit 子域
CREATE TABLE audit_tasks (
  id BIGSERIAL PRIMARY KEY,
  biz_module VARCHAR(50) NOT NULL,
  biz_id BIGINT NOT NULL,
  content TEXT,
  media_urls TEXT,
  status INT DEFAULT 0,                            -- 0待审 1通过 2拒绝 3人工审核中
  ai_result JSON,
  reviewer_id BIGINT,
  reviewed_at TIMESTAMP,
  created_at TIMESTAMP
);

CREATE TABLE sensitive_words (
  id BIGSERIAL PRIMARY KEY,
  word VARCHAR(100),
  category VARCHAR(20),                            -- politics/porn/ad/political
  level INT,                                       -- 1替换 2拒绝
  created_at TIMESTAMP
);

-- report 子域
CREATE TABLE reports (
  id BIGSERIAL PRIMARY KEY,
  reporter_id BIGINT NOT NULL,
  biz_module VARCHAR(50) NOT NULL,
  biz_id BIGINT NOT NULL,
  reason VARCHAR(500),
  evidence_urls TEXT,
  status INT DEFAULT 0,                            -- 0待处理 1已处理 2已申诉
  handler_id BIGINT,
  decision VARCHAR(200),
  created_at TIMESTAMP
);

CREATE TABLE violations (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  type VARCHAR(50),                                -- 警告/下架/限时禁发/永久封禁
  reason VARCHAR(500),
  start_at TIMESTAMP,
  end_at TIMESTAMP,
  created_at TIMESTAMP
);
```

### 4.8 LBS地图中台（lbs）

**职责**：高德定位/附近检索/距离排序/POI/路线规划/地理围栏/分站区域隔离

**核心表**：
```sql
CREATE TABLE lbs_pois (
  id BIGSERIAL PRIMARY KEY,
  region_id BIGINT,
  name VARCHAR(100),
  address VARCHAR(200),
  location GEOGRAPHY(Point, 4326),                 -- PostGIS 地理字段
  category VARCHAR(50),
  created_at TIMESTAMP
);
CREATE INDEX idx_lbs_pois_location ON lbs_pois USING GIST(location);

CREATE TABLE lbs_regions (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(100),
  city_code VARCHAR(20),
  center GEOGRAPHY(Point, 4326),
  boundary GEOGRAPHY(Polygon, 4326),
  status INT DEFAULT 1,
  created_at TIMESTAMP
);
CREATE INDEX idx_lbs_regions_boundary ON lbs_regions USING GIST(boundary);
```

**核心接口**：
- GET /api/v1/lbs/nearby?lat=&lng=&radius= - 附近检索
- GET /api/v1/lbs/distance?from=&to= - 距离计算
- GET /api/v1/lbs/route?from=&to= - 路线规划（调用高德API）
- GET /api/v1/lbs/region?lat=&lng= - 根据经纬度判断分站

### 4.9 AI智能中台（ai）

**职责**：AI发帖/AI润色/AI图文生成/AI内容摘要/智能推荐/AI客服

**核心表**：
```sql
CREATE TABLE ai_models (
  id BIGSERIAL PRIMARY KEY,
  provider VARCHAR(20),                            -- wenxin/qwen/deepseek/openai
  model_name VARCHAR(50),
  api_key VARCHAR(500),
  endpoint VARCHAR(500),
  status INT DEFAULT 1,
  created_at TIMESTAMP
);

CREATE TABLE ai_usage_logs (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT,
  biz_module VARCHAR(50),
  action VARCHAR(50),                              -- generate/polish/audit/chat
  input_tokens INT,
  output_tokens INT,
  cost DECIMAL(10,4),
  created_at TIMESTAMP
);
```

**多模型兼容**：文心/通义/DeepSeek/OpenAI 统一抽象，按成本和场景动态选择。

### 4.10 多租户分站中台（tenant）

**职责**：无限城市分站/独立配置/独立域名/独立运营权限/配置一键复制/数据隔离

**核心表**：
```sql
CREATE TABLE tenant_stations (
  id BIGSERIAL PRIMARY KEY,
  region_id BIGINT UNIQUE NOT NULL,
  name VARCHAR(100),
  domain VARCHAR(200),                             -- 独立域名
  status INT DEFAULT 1,
  config JSON,                                     -- 独立运营配置
  created_at TIMESTAMP
);

CREATE TABLE tenant_staff (
  id BIGSERIAL PRIMARY KEY,
  station_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  role VARCHAR(20),                                -- operator/manager
  permissions JSON,
  created_at TIMESTAMP
);

CREATE TABLE tenant_configs (
  id BIGSERIAL PRIMARY KEY,
  station_id BIGINT NOT NULL,
  biz_module VARCHAR(50) NOT NULL,
  config_key VARCHAR(100),
  config_value TEXT,
  updated_at TIMESTAMP
);
```

### 4.11 素材存储中台（material）

**职责**：图片/视频上传/转码/压缩/CDN/以图搜图/图片水印/素材管理

**核心表**：
```sql
CREATE TABLE material_files (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT,
  type VARCHAR(20),                                -- image/video/audio/doc
  original_name VARCHAR(500),
  storage_path VARCHAR(500),
  cdn_url VARCHAR(500),
  size INT,
  mime_type VARCHAR(100),
  width INT, height INT,
  duration INT,                                    -- 视频时长
  status INT DEFAULT 1,
  created_at TIMESTAMP
);

CREATE TABLE material_video_tasks (
  id BIGSERIAL PRIMARY KEY,
  file_id BIGINT NOT NULL,
  task_type VARCHAR(20),                           -- transcode/thumbnail/compress
  status INT DEFAULT 0,                            -- 0待处理 1处理中 2完成 3失败
  result JSON,
  created_at TIMESTAMP
);
```

### 4.12 前端能力中台（diy）

**职责**：拖拽生成首页/专题页/店铺页/活动页

**核心表**：
```sql
CREATE TABLE diy_pages (
  id BIGSERIAL PRIMARY KEY,
  region_id BIGINT,
  biz_module VARCHAR(50),                          -- home/topic/shop/activity
  biz_id BIGINT,
  title VARCHAR(100),
  components JSON,                                 -- 拖拽组件配置
  status INT DEFAULT 1,
  updated_at TIMESTAMP
);

CREATE TABLE diy_templates (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(100),
  category VARCHAR(50),
  preview VARCHAR(500),
  components JSON,
  created_at TIMESTAMP
);
```

**说明**：DIY 是能力型模块，不是业务模块，因此独立为中台，所有业务模块（首页/分类页/专题页/店铺页/活动页）都可通过 DIY 中台拖拽生成。

---

## 五、15 个垂直业务模块详细设计

### 5.1 模块清单与需求文档映射

| # | 垂直业务模块 | 表前缀 | 需求文档对应子模块 | P2 批次 |
|---|------------|--------|-------------------|---------|
| 1 | 分类信息核心 | info_ | dh114 + 通用信息发布 | 第1批 |
| 2 | 同城二手交易（ershou） | ers_ | ershou | 第1批（首个） |
| 3 | 同城招聘+零工驿站 | job_ | job + linggong | 第1批 |
| 4 | 同城房产 | house_ | fang | 第1批 |
| 5 | 同城黄页114商圈 | shop_ | dh114 商家 + quan 商家 | 第1批 |
| 6 | 同城到家预约服务 | daojia_ | daojia + yuyue | 第2批 |
| 7 | 同城商城电商 | mall_ | mall + kanjia + pintuan + choujiang + coupon | 第2批 |
| 8 | 同城头条资讯 | toutiao_ | toutiao + tuiwen | 第2批 |
| 9 | 同城圈子社群 | quan_ | quan + share | 第2批 |
| 10 | 同城活动 | huodong_ | huodong | 第2批 |
| 11 | 同城婚恋相亲 | love_ | love | 第3批 |
| 12 | 同城汽车 | car_ | car + pinche | 第3批 |
| 13 | 同城教育培训 | edu_ | (需求文档未列，新增) | 第3批 |
| 14 | 同城装修 | zhuangxiu_ | (需求文档未列，新增) | 第3批 |
| 15 | 同城直播 | zhibo_ | zhibo | 第3批 |

**说明**：
- 需求文档的 30 个功能模块按"业务领域"归属到 15 个垂直业务模块
- 例如：商城电商领域包含 mall/kanjia/pintuan/choujiang/coupon 5 个功能模块
- 营销5件套（kanjia/pintuan/choujiang/coupon/sign）中：coupon/sign 归营销活动中台，kanjia/pintuan/choujiang 归商城电商领域
- 用户运营5件套（renzheng/vipcard/partner/majia/jubao）：renzheng 归用户账号中台，vipcard 归用户账号中台，partner 归分销合伙人中台，majia 归用户账号中台，jubao 归风控审核中台
- 社区互动6件套（dashang/mingpian/zhibo/ai/tuiwen/share）：dashang 归营销活动中台，mingpian 归商家商户中台，zhibo 独立，ai 归 AI 智能中台，tuiwen 归头条资讯，share 归圈子社群

### 5.2 各模块核心能力（对照第二章去重提纯表）

#### 5.2.1 分类信息核心模块（info）
- 自定义分类模型（多级类目树）
- 付费发布（按分类/分站独立定价）
- AI 辅助发布（AI 生成标题/内容/润色）
- 自定义发布表单（不同分类不同字段）
- 一键生成 DM 报纸
- 求购专区
- 分站独立定价
- 金币激励（发布/浏览/分享得金币）
- PC/小程序双端独立页面
- 渠道二维码追踪
- 批量导出信息
- 朋友圈分享裂变

#### 5.2.2 同城二手交易模块（ershou）- P2 第一批次首个
- 担保交易完整流程（11状态机）
- 定金/全款双模式
- 以图搜图（接入素材存储中台）
- 隐私虚拟号（保护买卖双方）
- 视频加权（带视频商品优先曝光）
- 商家二手店铺体系（接入商家商户中台）
- 多维度筛选（成色/价格/距离/视频优先/同城自提）
- 信用分管控（接入风控审核中台）
- 商品权重排序（置顶/刷新/信用/浏览/点赞加权）
- 求购信息反向撮合
- 短视频商品（接入素材存储中台）
- 海报推广（接入前端能力中台）
- 13 张专属表（详见第八章）

#### 5.2.3 同城招聘+零工驿站模块（job）
- 企业端：企业认证/简历投递/岗位置顶套餐
- 零工驿站：临时工/日结/包工
- 工人信用评级（接入风控审核中台）
- 招工海报
- 招工签到激励
- 独立 PC 门户

#### 5.2.4 同城房产模块（house）
- 新房/二手房/租房/商铺
- 房源核验
- 经纪人店铺（接入商家商户中台）
- 看房预约
- 房源 DIY 详情（接入前端能力中台）
- 房源自动下架
- 合租筛选
- 租房朋友圈推广
- 独立 SEO PC 页面

#### 5.2.5 同城黄页114商圈模块（shop）
- 商家入驻/认领（接入商家商户中台）
- CRM
- 堂食点餐/餐桌码
- 团购
- 优惠券/红包
- 小票打印
- 自定义店铺装修（接入前端能力中台）
- 批量导入商家电话
- 店铺动态

#### 5.2.6 同城到家预约服务模块（daojia）
- 上门家政/维修
- 预约订单状态机
- 师傅接单（接入商家商户中台）
- 认养预定
- 定金/全款预约
- 服务点评
- 上门履约流程

#### 5.2.7 同城商城电商模块（mall）
- 商城完整订单（接入支付财务中台）
- 拼团（阶梯拼团）
- 砍价
- 抢购
- 礼品卡
- 积分兑换商城
- 会员黑卡
- 核销体系
- 到店自提
- 分销叠加优惠（接入分销合伙人中台）
- 多优惠叠加
- 商家独立结算

#### 5.2.8 同城头条资讯模块（toutiao）
- 同城头条
- 专题
- 本地爆料
- 图文投稿
- DM 报纸
- 传单生成
- 投票
- 影音视频
- DIY 专题页面（接入前端能力中台）

#### 5.2.9 同城圈子社群模块（quan）
- 同城话题
- 微信群
- 投票
- 短视频发布
- 一键分享

#### 5.2.10 同城活动模块（huodong）
- 线下市集
- 招聘会
- 相亲会
- 商家活动
- 活动报名
- 活动签到

#### 5.2.11 同城婚恋相亲模块（love）
- 实名认证（接入用户账号中台）
- 红娘管理
- 线上匹配推荐
- 线下相亲会
- 付费解锁
- 隐私号码
- 打赏互动

#### 5.2.12 同城汽车模块（car）
- 二手车
- 驾校
- 拼车（pinche）
- 车辆服务

#### 5.2.13 同城教育培训模块（edu）- 新增
- 培训机构
- 家教
- 课程预约
- 课程评价

#### 5.2.14 同城装修模块（zhuangxiu）- 新增
- 装修公司
- 工地案例
- 装修报价预约
- 装修日记

#### 5.2.15 同城直播模块（zhibo）
- 本地商家直播
- 带货
- 同城短视频

---

## 六、模块独立管控机制

### 6.1 模块全局开关体系

总后台可视化面板，任意垂直模块或通用中台一键启用/关闭：

**关闭后效果**：
1. APP/小程序/H5/PC 全部入口隐藏
2. 接口层自动拦截该模块所有 API，返回"未开通"提示
3. 模块专属定时任务全部暂停执行
4. 后台菜单、运营权限同步隐藏，运营人员无法操作该模块数据

**实现**：
```go
// plugin.Plugin 接口扩展（v3）
type Plugin interface {
    Name() string
    Version() string
    Init(ctx context.Context) error
    RegisterRoutes(router RouterGroup)
    Close() error
    Meta() PluginMeta  // v3 新增
}

type PluginMeta struct {
    Name         string
    DisplayName  string
    Category     string   // system/common/business
    Description  string
    Dependencies []string
    Icon         string
    Author       string
    Homepage     string
}

// 数据库表：modules（模块注册表）
CREATE TABLE modules (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(50) UNIQUE NOT NULL,
  display_name VARCHAR(100),
  category VARCHAR(30),                            -- system/common/business
  version VARCHAR(20),
  description VARCHAR(500),
  dependencies VARCHAR(500),                       -- JSON 数组
  enabled BOOLEAN DEFAULT true,
  installed_at TIMESTAMP,
  updated_at TIMESTAMP
);
```

### 6.2 分站独立精细化配置

每个分站可单独配置每个垂直模块/通用中台的全部运营参数：

**配置维度**：
- 付费价格：置顶、刷新、会员、商家入驻费分站自定义
- 佣金比例：二手、招聘、商城、分销分账比例分站独立设置
- 风控阈值：每日发布条数、信用分限制、审核开关分站可调
- 营销规则：优惠券力度、抽奖活动、积分获取比例分站自定义
- 展示规则：列表排序权重、视频加权、距离筛选范围分站可改

**实现**：
```sql
-- 扩展 setting 表（已存在）
CREATE TABLE module_station_configs (
  id BIGSERIAL PRIMARY KEY,
  station_id BIGINT NOT NULL,                      -- 分站ID
  module_name VARCHAR(50) NOT NULL,                -- 模块名
  config_key VARCHAR(100) NOT NULL,
  config_value TEXT,
  updated_at TIMESTAMP,
  UNIQUE(station_id, module_name, config_key)
);
```

### 6.3 数据完全隔离

数据库分层规范，后期支持单模块分库、独立 Schema：

1. **全局中台共享表**：user/pay/msg/region/setting 全平台共用
2. **每个垂直模块独立数据表前缀**：
   - 二手：ers_xxx
   - 招聘：job_xxx
   - 房产：house_xxx
   - 商圈：shop_xxx
   - 到家：daojia_xxx
   - 商城：mall_xxx
   - 头条：toutiao_xxx
   - 圈子：quan_xxx
   - 活动：huodong_xxx
   - 婚恋：love_xxx
   - 汽车：car_xxx
   - 教育：edu_xxx
   - 装修：zhuangxiu_xxx
   - 直播：zhibo_xxx
3. **按 region_id 城市维度数据隔离**：A 城市运营看不到 B 城市数据
4. **删除单个模块不会删除其他业务数据**：无脏数据污染

### 6.4 独立权限域

每个垂直模块独立权限分组：
- **超级管理员**：全平台全部模块权限
- **分站运营**：仅管理当前城市所有模块
- **模块专员**：仅分配单一模块管理权限（如只管理二手举报、看不到招聘数据）
- **商家账号**：仅自身店铺所属模块数据权限

**实现**：基于现有 RBAC（role_permissions 表），新增"模块权限域"概念：
```sql
-- 权限码规范：<module>:<action>
-- 示例：ershou:read / ershou:publish / ershou:audit / job:read / job:publish
-- 模块专员角色仅分配单一模块的权限码
```

### 6.5 独立监控/灰度/故障隔离

1. **每个模块独立监控面板**：接口 QPS、错误率、慢 SQL、定时任务执行状态，模块故障单独告警，不混淆全站问题
2. **灰度发布**：新功能仅对指定城市/指定用户开放，单模块一键回滚，无需重启整个服务
3. **熔断降级**：某模块并发过高自动限流，不拖垮其他业务模块

**实现**：
```sql
CREATE TABLE module_grayscales (
  id BIGSERIAL PRIMARY KEY,
  module_name VARCHAR(50) NOT NULL,
  version VARCHAR(20),
  gray_type VARCHAR(20),                           -- city/user/percentage
  gray_config JSON,                                -- {"cities":[1,2,3]} 或 {"user_ids":[1,2,3]} 或 {"percentage":10}
  status INT DEFAULT 0,                            -- 0未启用 1灰度中 2全量 3已回滚
  created_at TIMESTAMP
);

CREATE TABLE module_metrics (
  id BIGSERIAL PRIMARY KEY,
  module_name VARCHAR(50) NOT NULL,
  metric_name VARCHAR(100),
  metric_value DECIMAL(12,4),
  labels JSON,
  recorded_at TIMESTAMP
);
```

---

## 七、后台五中心分离架构

### 7.1 整体架构

```
管理后台 (Vue3 + Element Plus)
├─ 工作台（首页）
│  ├─ 待办事项（审核/退款/举报/提现）
│  ├─ 数据概览（今日GMV/新增用户/订单量/付费收入）
│  └─ 快捷入口（常用模块入口）
│
├─ 模块中心（15个垂直业务模块）
│  ├─ 分类信息 / 二手 / 招聘 / 房产 / 黄页114
│  ├─ 到家预约 / 商城电商 / 头条资讯 / 圈子社群
│  ├─ 活动 / 婚恋 / 汽车 / 教育 / 装修 / 直播
│  └─ 每个模块统一包含：内容管理/审核/数据/配置 4 个子页
│
├─ 中台中心（12个通用中台）
│  ├─ 用户账号 / 支付财务 / IM消息 / 商家商户
│  ├─ 分销合伙人 / 营销活动 / 风控审核 / LBS地图
│  ├─ AI智能 / 多租户分站 / 素材存储 / 前端能力
│  └─ 每个中台统一包含：配置/数据/监控/日志 4 个子页
│
├─ 设置中心
│  ├─ 模块开关（一键启用/关闭任意模块）
│  ├─ 分站管理（城市分站独立配置）
│  ├─ 权限管理（10角色+模块独立权限域）
│  ├─ 灰度发布（模块版本灰度）
│  └─ 系统设置（基础配置）
│
└─ 数据中心
   ├─ 全平台数据大屏
   ├─ 分模块数据看板
   └─ 跨城市运营报表
```

### 7.2 工作台（首页）

**功能**：
1. **待办事项**：审核任务/退款申请/举报工单/提现申请，红点提醒
2. **数据概览**：今日 GMV/新增用户/订单量/付费收入/活跃用户，对比昨日
3. **快捷入口**：常用模块入口（可自定义）
4. **系统通知**：平台公告/模块故障告警/灰度发布通知

### 7.3 模块中心（15个垂直业务模块）

**统一布局**：每个模块 4 个子页
- **内容管理**：列表/详情/批量操作
- **审核**：待审核/已审核/拒绝
- **数据**：发布量/浏览量/咨询量/成交量/GMV
- **配置**：模块设置（价格/佣金/风控/UI，分站独立）

**示例 - 二手模块**：
```
模块中心
└─ 二手交易
   ├─ 内容管理（商品列表/详情/批量上下架）
   ├─ 审核（待审核/已审核/拒绝/违规）
   ├─ 数据（发布量/浏览UV/咨询量/成交GMV/付费收入/求购量/举报量/违规下架数/城市热度排行）
   └─ 配置（发布规则/置顶价格/佣金比例/求购规则/风控阈值/UI 配置）
```

### 7.4 中台中心（12个通用中台）

**统一布局**：每个中台 4 个子页
- **配置**：中台配置（如支付渠道/IM消息模板/风控规则）
- **数据**：中台数据（如订单量/钱包余额/IM消息量）
- **监控**：中台监控（接口QPS/错误率/慢SQL/定时任务状态）
- **日志**：中台日志（操作日志/资金流水/AI调用日志）

**示例 - 支付财务中台**：
```
中台中心
└─ 支付财务
   ├─ 配置（支付渠道/分账规则/风控阈值/对账配置）
   ├─ 数据（订单量/钱包余额/退款量/分账记录/对账报表）
   ├─ 监控（支付接口QPS/支付错误率/慢SQL/定时任务）
   └─ 日志（资金流水/提现日志/对账日志/风控日志）
```

### 7.5 设置中心

1. **模块开关**：可视化面板，一键启用/关闭任意模块
2. **分站管理**：城市分站列表/新增/编辑/独立配置
3. **权限管理**：10角色+模块独立权限域
4. **灰度发布**：模块版本灰度配置/回滚
5. **系统设置**：基础配置（站点信息/邮件/短信/CDN等）

### 7.6 数据中心

1. **全平台数据大屏**：总 GMV/总用户/总订单/总付费/活跃用户/在线用户
2. **分模块数据看板**：每个模块独立数据看板（GMV/付费/举报/违规/城市排行）
3. **跨城市运营报表**：城市对比/趋势分析/导出 Excel

### 7.7 三层后台体系

1. **平台总后台**：全局中台配置、模块开关、分站管理、全平台数据大盘
2. **城市分站后台**：当前城市独立配置、本地数据、本地活动运营
3. **商家独立后台**：店铺管理、订单、商品、营销、结算，适配手机/PC访问

**实现**：
- 平台总后台：`frontend/admin/`（现有）
- 城市分站后台：`frontend/admin-station/`（新建，复用平台总后台组件库）
- 商家独立后台：`frontend/admin-shop/`（新建，复用平台总后台组件库）

### 7.8 后台核心设计原则

1. **模块化菜单**：按"模块域"分组，根据用户权限和模块开关动态加载
2. **统一组件库**：表单/表格/筛选器/详情弹窗/图表 统一封装，各模块复用
3. **统一权限**：10角色 + 模块独立权限域
4. **统一数据**：全平台大屏 + 分模块看板 + 跨城市报表
5. **响应式布局**：自适应大屏/中屏/小屏，避免大屏两侧空白过多（用户规则）
6. **配色统一**：边框色/元素高度/hover 状态颜色统一（用户规则）

---

## 八、ershou 专属设计（P2 第一批次首个）

### 8.1 13 张专属表设计

> 已存在 4 张（erhous/ershou_images/ershou_favorites/ershou_messages），新增 9 张

```sql
-- 1. erhous 主表（已存在，扩展字段）
ALTER TABLE erhous ADD COLUMN IF NOT EXISTS video_url VARCHAR(500) DEFAULT '';
ALTER TABLE erhous ADD COLUMN IF NOT EXISTS video_cover VARCHAR(500) DEFAULT '';
ALTER TABLE erhous ADD COLUMN IF NOT EXISTS weight INT DEFAULT 0;  -- 商品权重
ALTER TABLE erhous ADD COLUMN IF NOT EXISTS shop_id BIGINT DEFAULT 0;  -- 商家店铺ID
ALTER TABLE erhous ADD COLUMN IF NOT EXISTS is_b2c BOOLEAN DEFAULT false;  -- 是否B2C商品
ALTER TABLE erhous ADD COLUMN IF NOT EXISTS virtual_phone VARCHAR(20) DEFAULT '';  -- 隐私虚拟号
ALTER TABLE erhous ADD COLUMN IF NOT EXISTS sold_price DECIMAL(12,2) DEFAULT 0;
ALTER TABLE erhous ADD COLUMN IF NOT EXISTS sold_at TIMESTAMP;

-- 2. ershou_images（已存在）
-- 3. ershou_favorites（已存在）
-- 4. ershou_messages（已存在）

-- 5. ershou_wanteds 求购信息表（新增）
CREATE TABLE ershou_wanteds (
  id BIGSERIAL PRIMARY KEY,
  region_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  title VARCHAR(200) NOT NULL,
  description TEXT,
  category_id BIGINT,
  budget_min DECIMAL(12,2) DEFAULT 0,
  budget_max DECIMAL(12,2) DEFAULT 0,
  status INT DEFAULT 1,                            -- 0草稿 1发布 2已成交 3已关闭
  expire_at TIMESTAMP,
  created_at TIMESTAMP
);
CREATE INDEX idx_ershou_wanteds_region ON ershou_wanteds(region_id, status);

-- 6. ershou_wanted_images 求购图片表（新增）
CREATE TABLE ershou_wanted_images (
  id BIGSERIAL PRIMARY KEY,
  wanted_id BIGINT NOT NULL,
  image_url VARCHAR(500),
  sort INT DEFAULT 0
);

-- 7. ershou_footprints 足迹表（新增）
CREATE TABLE ershou_footprints (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  ershou_id BIGINT NOT NULL,
  created_at TIMESTAMP
);
CREATE INDEX idx_ershou_footprints_user ON ershou_footprints(user_id, created_at);

-- 8. ershou_attributes 属性表（新增，对标58多规格筛选）
CREATE TABLE ershou_attributes (
  id BIGSERIAL PRIMARY KEY,
  category_id BIGINT NOT NULL,
  name VARCHAR(50) NOT NULL,                       -- 内存/成色/型号
  type VARCHAR(20),                                -- select/checkbox/input
  options JSON,                                    -- 选项列表
  sort INT DEFAULT 0
);

CREATE TABLE ershou_attribute_values (
  id BIGSERIAL PRIMARY KEY,
  ershou_id BIGINT NOT NULL,
  attribute_id BIGINT NOT NULL,
  value VARCHAR(200)
);

-- 9. ershou_user_behaviors 行为埋点表（新增，LBS推荐系统）
CREATE TABLE ershou_user_behaviors (
  id BIGSERIAL PRIMARY KEY,
  region_id BIGINT NOT NULL,
  user_id BIGINT DEFAULT 0,                        -- 0=未登录游客
  ershou_id BIGINT NOT NULL,
  action VARCHAR(20) NOT NULL,                     -- view/click/fav/message/order/share
  source VARCHAR(30) DEFAULT '',                   -- list/detail/search/recommend/nearby
  latitude DECIMAL(10,7) DEFAULT 0,
  longitude DECIMAL(10,7) DEFAULT 0,
  device_id VARCHAR(100) DEFAULT '',
  platform VARCHAR(20) DEFAULT '',                 -- android/ios/web/weapp
  created_at TIMESTAMP
);
CREATE INDEX idx_behavior_user ON ershou_user_behaviors(user_id, created_at);
CREATE INDEX idx_behavior_ershou ON ershou_user_behaviors(ershou_id, action);
CREATE INDEX idx_behavior_region_time ON ershou_user_behaviors(region_id, created_at);

-- 10. ershou_user_profiles 用户画像表（新增，LBS推荐系统）
CREATE TABLE ershou_user_profiles (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT UNIQUE NOT NULL,
  prefer_categories JSON,                          -- 偏好品类
  price_range JSON,                                -- 价格区间
  trade_radius INT DEFAULT 5000,                   -- 交易半径（米）
  active_hours JSON,                               -- 活跃时段
  updated_at TIMESTAMP
);

-- 11. ershou_shop_secondhand 商家二手店铺表（新增，B2C体系）
CREATE TABLE ershou_shop_secondhand (
  id BIGSERIAL PRIMARY KEY,
  shop_id BIGINT NOT NULL,                         -- 关联 merchant_shops
  name VARCHAR(100),
  level INT DEFAULT 1,
  credit_score INT DEFAULT 100,
  total_goods INT DEFAULT 0,
  sold_goods INT DEFAULT 0,
  settle_cycle INT DEFAULT 1,                      -- 1=T+1 7=T+7
  created_at TIMESTAMP
);

-- 12. ershou_video_tasks 视频处理任务表（新增）
CREATE TABLE ershou_video_tasks (
  id BIGSERIAL PRIMARY KEY,
  ershou_id BIGINT NOT NULL,
  file_id BIGINT NOT NULL,                         -- 关联 material_files
  task_type VARCHAR(20),                           -- transcode/thumbnail/compress
  status INT DEFAULT 0,                            -- 0待处理 1处理中 2完成 3失败
  result JSON,
  created_at TIMESTAMP
);

-- 13. ershou_data_stats 数据看板表（新增）
CREATE TABLE ershou_data_stats (
  id BIGSERIAL PRIMARY KEY,
  region_id BIGINT NOT NULL,
  stat_date DATE NOT NULL,
  publish_count INT DEFAULT 0,
  view_count INT DEFAULT 0,
  consult_count INT DEFAULT 0,
  order_count INT DEFAULT 0,
  gmv DECIMAL(12,2) DEFAULT 0,
  paid_income DECIMAL(12,2) DEFAULT 0,
  wanted_count INT DEFAULT 0,
  report_count INT DEFAULT 0,
  violation_count INT DEFAULT 0,
  UNIQUE(region_id, stat_date)
);
```

### 8.2 模块设置配置项（6 大类，分站独立）

1. **发布规则**：每日发布上限/草稿保留天数/重复发布拦截/低信用拦截
2. **置顶价格**：24h置顶/7天置顶/30天置顶/多位置顶（首页/分类页/列表页）
3. **佣金比例**：平台佣金/分站分成/合伙人分销/卖家所得
4. **求购规则**：每日求购上限/求购过期天数/求购是否付费
5. **风控阈值**：敏感词等级/AI审核开关/信用分限制/违规处罚规则
6. **UI 配置**：列表排序权重/视频加权倍数/距离筛选范围/同城自提开关

### 8.3 担保交易订单状态机（11 状态）

| # | 状态 | 触发动作 | 资金状态 |
|---|------|---------|---------|
| 0 | 待支付 | 下单 | 无 |
| 1 | 已支付待发货 | 支付成功 | 买家扣款，平台冻结 |
| 2 | 已发货 | 卖家发货 | 平台冻结 |
| 3 | 已收货待评价 | 买家确认收货 | 平台冻结 |
| 4 | 已结算完成 | 7天自动结算/主动评价 | 平台分账给卖家 |
| 5 | 已取消 | 买家取消/超时未支付 | 无 |
| 6 | 退款中 | 买家申请退款 | 平台冻结 |
| 7 | 退款待退货 | 卖家同意退款 | 平台冻结 |
| 8 | 已退款 | 买家退货确认 | 退款给买家 |
| 9 | 纠纷中 | 平台介入 | 平台冻结 |
| 10 | 已关闭 | 仲裁完成/超时 | 按判决分配 |

**超时自动任务**：
- 待支付 30 分钟自动关闭
- 已收货 7 天自动结算
- 退款申请 3 天未处理自动平台介入

### 8.4 金融级资金防护（详见第九章）

### 8.5 LBS 推荐系统（详见第十章）

---

## 九、金融级资金防护

### 9.1 本地消息表+定时对账（替代 Seata）

**架构**：
```
业务操作 + 写消息表（同一事务）
       ↓
定时任务扫描消息表
       ↓
执行跨模块调用（wallet.freeze / wallet.transfer / ...）
       ↓
标记消息已处理 / 失败重试
```

**表结构**：
```sql
CREATE TABLE message_queue (
  id BIGSERIAL PRIMARY KEY,
  biz_module VARCHAR(50) NOT NULL,                 -- 来源模块
  biz_id BIGINT NOT NULL,                          -- 业务ID
  message_type VARCHAR(50) NOT NULL,               -- wallet.freeze/wallet.unfreeze/wallet.transfer/...
  payload JSON NOT NULL,
  status INT NOT NULL DEFAULT 0,                   -- 0待处理 1已处理 2失败
  retry_count INT DEFAULT 0,
  max_retry INT DEFAULT 3,
  next_retry_at TIMESTAMP,
  processed_at TIMESTAMP,
  fail_reason VARCHAR(500),
  created_at TIMESTAMP
);
CREATE INDEX idx_mq_status ON message_queue(status, next_retry_at);
```

### 9.2 多级分账规则

```sql
CREATE TABLE settle_rules (
  id BIGSERIAL PRIMARY KEY,
  region_id BIGINT NOT NULL,
  biz_module VARCHAR(50) NOT NULL,                 -- ershou/mall/...
  platform_rate DECIMAL(5,4) NOT NULL,             -- 平台佣金
  station_rate DECIMAL(5,4) DEFAULT 0,             -- 分站分成
  partner_rate DECIMAL(5,4) DEFAULT 0,             -- 合伙人分销
  seller_rate DECIMAL(5,4) NOT NULL,               -- 卖家所得
  updated_at TIMESTAMP
);
-- 四方分账比例之和必须 = 1
```

### 9.3 资金风控规则

```sql
CREATE TABLE risk_rules (
  id BIGSERIAL PRIMARY KEY,
  region_id BIGINT,
  rule_type VARCHAR(50),                           -- large_amount/high_frequency/withdraw_audit
  threshold DECIMAL(12,2),                         -- 阈值
  action VARCHAR(50),                              -- block/manual_audit/alert
  enabled BOOLEAN DEFAULT true,
  created_at TIMESTAMP
);
```

**规则示例**：
- 单日大额交易拦截：单笔 > 10000 元自动人工审核
- 高频交易风控：1 小时内 > 10 笔自动拦截
- 提现分级审核：< 1000 自动通过 / 1000-10000 人工审核 / > 10000 主管审核

### 9.4 三种冻结池

```sql
-- wallet_accounts 表的三个冻结字段
frozen DECIMAL(12,2) DEFAULT 0,                    -- 下单冻结（待结算）
frozen_after_sale DECIMAL(12,2) DEFAULT 0,         -- 售后冻结（退款中）
frozen_violation DECIMAL(12,2) DEFAULT 0,          -- 违规冻结（平台处罚）
```

**说明**：三种冻结池互不干扰，避免资金混淆。

### 9.5 每日对账任务

```sql
CREATE TABLE reconcile_reports (
  id BIGSERIAL PRIMARY KEY,
  region_id BIGINT,
  report_date DATE,
  platform_total DECIMAL(12,2),                    -- 平台资金
  merchant_total DECIMAL(12,2),                    -- 商家资金
  user_balance_total DECIMAL(12,2),                -- 用户余额
  diff DECIMAL(12,2) DEFAULT 0,                    -- 差异
  status INT DEFAULT 0,                            -- 0正常 1异常
  created_at TIMESTAMP
);
```

每日凌晨 2:00 自动执行：
1. 平台资金 + 商家资金 + 用户余额 = 总资金
2. 差异 > 0.01 元自动生成工单预警
3. 异常自动通知管理员（短信+站内信）

---

## 十、数据中台与 LBS 推荐系统

### 10.1 用户行为埋点

**埋点表**：`ershou_user_behaviors`（已在第八章设计）

**埋点字段**：
- user_id / ershou_id / action（view/click/fav/message/order/share）
- source（list/detail/search/recommend/nearby）
- latitude / longitude（LBS 维度）
- device_id / platform（设备维度）

### 10.2 用户画像离线计算

**画像表**：`ershou_user_profiles`（已在第八章设计）

**画像字段**：
- prefer_categories：偏好品类（基于浏览/收藏/购买行为计算）
- price_range：价格区间（基于成交订单计算）
- trade_radius：交易半径（基于 LBS 行为计算）
- active_hours：活跃时段（基于登录时间计算）

**计算任务**：每日凌晨 3:00 执行（cron + goroutine）

### 10.3 多路召回

1. **附近召回**：PostGIS 查询用户活跃半径内的商品
2. **兴趣召回**：基于用户偏好品类召回
3. **热门召回**：基于全平台浏览/咨询量召回热门商品
4. **相似召回**：基于用户最近浏览商品的相似商品

### 10.4 排序公式

```
final_score = 
    weight_score * 0.4 +                           -- 商品权重（置顶/刷新/信用/浏览/点赞）
    distance_score * 0.3 +                         -- 距离得分（越近越高）
    behavior_score * 0.2 +                         -- 行为得分（浏览/收藏/咨询量）
    video_score * 0.1                              -- 视频加权（有视频加成）
```

**分站可调**：每个城市可独立配置 4 个权重系数。

### 10.5 实时热点加权

```sql
CREATE TABLE hot_items (
  id BIGSERIAL PRIMARY KEY,
  region_id BIGINT NOT NULL,
  biz_module VARCHAR(50),
  biz_id BIGINT NOT NULL,
  hot_score DECIMAL(12,4),
  expire_at TIMESTAMP,
  created_at TIMESTAMP
);
```

**实时热点规则**：1 小时内浏览量 > 100 自动加入热点表，加权倍数 1.5x（分站可调）。

---

## 十一、运维监控与灰度发布

### 11.1 Prometheus + Grafana 监控

**监控指标**：
- 接口响应时间（按模块独立）
- 错误率（按模块独立）
- 并发量（按模块独立）
- 数据库慢查询
- Redis 命中率
- ES 查询性能
- 定时任务执行状态

**告警规则**：
- 接口响应时间 > 1s 告警
- 错误率 > 1% 告警
- 数据库慢查询 > 1s 告警
- 定时任务失败告警
- 模块故障告警（钉钉/短信）

### 11.2 模块独立监控面板

Grafana 为每个模块创建独立 Dashboard：
- 二手模块面板：发布量/浏览量/咨询量/成交量/GMV/付费收入
- 招聘模块面板：发布量/投递量/企业入驻量
- 支付财务面板：订单量/支付量/退款量/对账差异/资金流水
- IM消息面板：消息量/在线用户/会话量/推送成功率
- 风控审核面板：审核量/违规量/举报量/处罚量

### 11.3 CI/CD 流水线

Gitee 代码提交自动触发：
1. 单元测试
2. 分模块打包
3. 灰度部署（按城市/用户ID）
4. 全量部署

**单模块独立发布**：支持单模块单独发布，不重启整个服务。

### 11.4 数据备份分级

| 数据级别 | 备份策略 | 恢复目标 |
|---------|---------|---------|
| 金融数据（订单/钱包/分账） | 3-2-1 异地备份，每日全量+实时增量 | RPO<1min, RTO<1h |
| 业务数据（商品/用户/消息） | 每日全量备份 | RPO<24h, RTO<4h |
| 日志数据 | 滚动保留 30 天 | 不恢复 |
| 监控数据 | 滚动保留 90 天 | 不恢复 |

**单模块数据单独恢复**：支持按模块独立恢复（如只恢复 ershou 模块数据）。

### 11.5 定时任务调度中心

```sql
CREATE TABLE cron_jobs (
  id BIGSERIAL PRIMARY KEY,
  module_name VARCHAR(50) NOT NULL,                -- 所属模块
  job_name VARCHAR(100) NOT NULL,
  cron_expr VARCHAR(50),                           -- cron 表达式
  handler VARCHAR(200),                            -- 处理函数
  params JSON,
  enabled BOOLEAN DEFAULT true,
  last_run_at TIMESTAMP,
  last_status INT,                                 -- 0未执行 1成功 2失败
  last_error TEXT,
  created_at TIMESTAMP
);
```

**功能**：
- 所有模块的定时任务统一调度
- 可单独启停某模块的定时任务
- 任务执行历史记录
- 失败自动重试 + 告警

---

## 十二、5 阶段递进开发计划（14.5 周）

### P0 底层基座（1.5 周，4 Agent 并行）

**交付**：
1. Go 模块化插件底层框架（统一 Plugin 接口 + Meta() 方法 + 模块开关 + 动态注册 + 独立路由）
2. modules 模块注册表 + cron_jobs 调度中心 + module_grayscales 灰度发布表
3. message_queue 本地消息表 + 对账任务基础框架
4. module_station_configs 分站配置中心扩展
5. Prometheus + Grafana 监控基座接入
6. 数据库分表前缀规范 + 自动迁移工具 + PostGIS 地理索引
7. 基础用户/地区/文件/权限通用底座
8. 4 端管理后台「模块总控面板」基础页面

**4 Agent 并行**：
- 后端架构 Agent：Go 插件底层框架、统一 Plugin 接口、路由动态注册
- 数据库 Agent：PG 表结构、分表隔离、迁移脚本、PostGIS 地理字段
- 前端基座 Agent：UniAppX 初始化、Next16 PC 基座、管理后台五中心基础框架
- 运维 Agent：Docker Compose、监控、Gitee CI/CD 流水线

### P1 12 大通用中台完整开发（4 周，多中台 Agent 并行）

**交付**：user/pay/im/merchant/distribution/marketing/risk/lbs/ai/tenant/material/diy 全套中台，配套完整数据库、接口、管理后台、定时任务、资金对账、分布式事务能力。

**内部子域分层**：
- pay 中台：pay/wallet/order/refund/settle 5 子域
- marketing 中台：ad/coupon/sign/activity 4 子域
- risk 中台：credit/audit/report 3 子域

### P2 垂直业务模块批量开发（6 周，3 批次内部并行）

**第一批次（2 周，5 模块并行）**：
1. **ershou 同城二手**（首个，最复杂，含担保交易+短视频+LBS推荐+商家B2C）
2. 分类信息核心
3. 招聘+零工驿站
4. 房产
5. 黄页114商圈

**第二批次（2 周，5 模块并行）**：
6. 到家预约服务
7. 商城电商（含拼团/砍价/抢购/礼品卡）
8. 头条资讯
9. 圈子社群
10. 活动

**第三批次（2 周，5 模块并行）**：
11. 婚恋相亲
12. 汽车（含拼车）
13. 教育培训
14. 装修
15. 直播

**每个模块 4 端同步**：后端 + 管理后台 + PC门户 + UniAppX

### P3 数据中台+推荐系统（2 周）

**交付**：
1. 用户行为埋点全端接入
2. 用户画像离线计算（cron + goroutine）
3. 多路召回（附近/兴趣/热门/相似）
4. 个性化推荐排序
5. 全平台数据大屏
6. 跨城市运营报表
7. 实时热点加权

### P4 灰度运维商业化完善（1 周）

**交付**：
1. 模块灰度发布后台
2. 分账体系完善
3. 财务对账导出
4. 批量运营工具
5. 多端营销组件（红包/抽奖/签到）
6. 全平台风控规则
7. 分站批量配置工具

### 总周期：1.5 + 4 + 6 + 2 + 1 = 14.5 周

---

## 十三、与现有模块的关系

### 13.1 已有模块映射

| 现有模块 | v3 归属 | 备注 |
|---------|---------|------|
| backend/internal/modules/ershou | 垂直业务 - 二手 | 已有 4 张表，扩展为 13 张 |
| backend/internal/modules/news | 垂直业务 - 头条资讯 | 现有 news 模块即 toutiao |
| backend/internal/modules/region | 多租户分站中台 | 现有 region 表扩展为 tenant_stations |
| backend/internal/modules/setting | 多租户分站中台 | 现有 setting 表扩展为 module_station_configs |
| backend/internal/modules/file | 素材存储中台 | 现有 file 模块扩展为 material |
| backend/internal/modules/user | 用户账号中台 | 现有 user 模块扩展为 user 中台 |
| backend/internal/modules/role | 用户账号中台 | 现有 role 模块归入 user 中台权限域 |
| backend/internal/modules/dashboard | 后台工作台 | 现有 dashboard 即工作台首页 |
| backend/internal/pkg/ws | IM消息中台 | 现有 ws 包扩展为 im 中台 |
| backend/internal/core/plugin | 接入层 | 扩展 Plugin 接口 + Meta() 方法 |

### 13.2 待迁移模块

现有部分模块需要按 v3 架构重新组织：
- 现有 shop 模块 → 拆分为 merchant 商家商户中台 + shop 垂直业务（黄页114商圈）
- 现有 groupbuy 模块 → 归入 shop 垂直业务（黄页114商圈的团购能力）

---

## 十四、风险与约束

### 14.1 技术风险

1. **支付/提现**：MVP 用 mock 支付，后续接入微信/支付宝真实支付
2. **实名认证**：MVP 用人工审核，后续接入第三方实名 API
3. **物流跟踪**：MVP 用快递公司公开 API，后续接入物流中台
4. **AI 审核**：MVP 用 DFA 敏感词过滤，后续接入 AI 图文审核
5. **数据迁移**：现有 4 张 ershou 表扩展字段时需兼容历史数据

### 14.2 性能风险

1. **PostGIS 查询**：百万级数据需优化索引，避免全表扫描
2. **WebSocket 单实例**：1 万连接上限，超过需集群扩展
3. **ES 索引**：分类/头条/商家全文检索需优化分词器
4. **Redis 缓存**：热点数据需合理设置过期时间，避免雪崩

### 14.3 数据风险

1. **金融数据**：必须 3-2-1 备份，RPO<1min
2. **用户数据**：每日全量备份，RPO<24h
3. **分站隔离**：A 城市运营不能看到 B 城市数据
4. **模块删除**：删除模块不删除其他业务数据

### 14.4 并发风险

1. **模块故障**：单模块故障不影响全站（熔断降级）
2. **定时任务**：模块独立队列，互不阻塞
3. **数据库连接**：按模块独立连接池配置

---

## 十五、审核清单

### 15.1 架构层面
- [ ] 大厂四层架构（接入/多端交付/业务/基础设施）是否合理
- [ ] 12 大通用中台 + 15 垂直业务 是否覆盖全部业务场景
- [ ] 中台内部子域分层（pay 5子域/marketing 4子域/risk 3子域）是否合理
- [ ] 模块依赖关系是否清晰（业务模块仅依赖中台抽象接口）

### 15.2 去重提纯
- [ ] 17 类去重提纯表是否完整覆盖点微+西瓜全部能力
- [ ] 4 类剔除功能是否合理（Discuz/马甲APP/边角工具/分散插件）
- [ ] 去重提纯表是否作为模块设计权威依据

### 15.3 模块独立管控
- [ ] 模块全局开关 / 独立配置 / 独立权限 / 独立定时任务 / 独立数据表前缀 / 灰度发布 是否完整
- [ ] 分站独立精细化配置（5 维度：价格/佣金/风控/营销/展示）是否完整
- [ ] 数据完全隔离（全局共享表 + 模块独立前缀 + region_id 隔离）是否到位

### 15.4 通用中台
- [ ] 12 个中台边界是否清晰，是否有遗漏或过度抽取
- [ ] 重负载中台内部子域分层是否合理
- [ ] DIY 独立为前端能力中台是否合理
- [ ] 金融级能力（本地消息表+对账+分账+风控+三种冻结池）是否到位

### 15.5 垂直业务
- [ ] 15 个垂直业务模块是否覆盖需求文档 30 个功能模块
- [ ] ershou 13 张专属表是否覆盖 dismall 5.1 + 闲鱼/转转/58 全部能力
- [ ] 字段是否完整
- [ ] ershou 作为 P2 第一批次首个是否合理

### 15.6 后台架构
- [ ] 五中心分离（工作台/模块中心/中台中心/设置中心/数据中心）是否合理
- [ ] 三层后台体系（平台总后台/城市分站后台/商家独立后台）是否合理
- [ ] 统一组件库 + 统一权限 + 统一数据 是否满足运营需求

### 15.7 开发计划
- [ ] 5 阶段 14.5 周（1.5+4+6+2+1）节奏是否合理
- [ ] P2 分 3 批次内部并行（每批次 5 模块并行）是否可行
- [ ] 依赖关系是否正确（P0 → P1 → P2 → P3 → P4）

### 15.8 关键技术决策
- [ ] 本地消息表替代 Seata 是否可接受
- [ ] Go 替代 Flink 自建推荐系统是否可行
- [ ] Docker Compose 替代 K8s 是否可接受
- [ ] WebSocket 单实例 MVP 是否足够

### 15.9 风险约束
- [ ] 支付/提现/实名/物流 mock 策略是否可接受
- [ ] 数据迁移风险是否可控
- [ ] 并发风险是否可控

### 15.10 大厂级能力补强（v3.1 新增）
- [ ] 链路追踪（OpenTelemetry + Jaeger）是否覆盖全链路排障场景
- [ ] 结构化日志体系（Loki + Promtail）是否满足日志检索+告警需求
- [ ] 监控告警规则（4 级告警 + 5 渠道通知）是否完整
- [ ] 安全防护（WAF/SQL注入/XSS/CSRF/DDoS/接口防重放）是否到位
- [ ] 分布式锁（Redis RedLock）是否覆盖金融级场景
- [ ] 缓存防护（穿透/击穿/雪崩）三件套是否到位
- [ ] 消息幂等（全链路）是否覆盖金融场景
- [ ] 配置中心（独立中台）是否支持热更新
- [ ] API 版本管理 + Swagger 文档生成是否到位
- [ ] 测试策略（单元/集成/E2E/压测）是否完整
- [ ] 数据库读写分离 + 分库分表策略是否预留
- [ ] 灾备方案（同城双活 + 异地灾备）是否到位
- [ ] 微服务演进路径（服务注册发现 + 拆分策略 + 服务网格）是否清晰

### 15.11 v3.2 补强细节审核（25 项，用户原话："全部补齐"）

用户在 v3.1 审核后选择"再次深度审计 v3.1"，发现 25 项可补强细节，并确认"全部补齐(推荐)"，文档升级为 v3.2。审核要点如下：

#### 高优先级 8 项（金融安全核心）
- [ ] RedLock 完整算法（5 实例 quorum + 耗时校验 + 分阶段引入 MVP→双实例→5 实例）是否覆盖金融级锁场景
- [ ] 退款资金来源（按订单状态自动选择：冻结释放 / 卖家扣减 / 平台垫付+欠款）是否明确无歧义
- [ ] 二次验证 2FA（7 类场景 + 短信/人脸/TOTP 三通道 + 5 次防爆破）是否覆盖全部敏感操作
- [ ] 分库分表分布式事务（MVP 本地消息表 → 中期 Saga → 远期 DTM 分阶段引入）是否可行
- [ ] 灾备 DNS 切换（GSLB + 智能DNS + 5s 健康检查 + 30s TTL + 30s 切换）是否满足 RTO<1min
- [ ] 微服务双写一致性校验（5 分钟定时对账 + 差异率告警 >0.01% + 自动修复以新库为准）是否到位
- [ ] JWT 刷新机制（Access 30min + Refresh 7d + Redis 黑名单 + token_version 密码修改失效）是否安全
- [ ] 行为分析防刷（设备指纹 + IP 画像 + 行为频率 + 风险评分 0-100 + pass/verify/block 三级决策）是否有效

#### 中优先级 12 项（大厂标准完整性）
- [ ] WebSocket 集群（Redis Pub/Sub 广播 + ClusterHub + 本地 Hub + 实例 ID）是否支持多实例横向扩展
- [ ] 布隆过滤器持久化（Redis Bitmap + 双 key 原子切换 + 每日重建）是否避免重启全量回灌
- [ ] 消息幂等表清理（PostgreSQL 月度分区 + 30 天 DROP + 死信保留 90 天）是否避免无限膨胀
- [ ] 配置中心高可用降级（本地缓存 → Redis → DB → 编译时默认值 四级 + 30s 定时拉取降级）是否到位
- [ ] 日志脱敏（zerolog Hook + 11 类敏感字段自动脱敏）是否覆盖 phone/idcard/bankcard/password/token 等
- [ ] 统一审计日志表（audit_logs + 月度分区 + 1 年保留）是否覆盖 create/update/delete/login/logout/export
- [ ] 文件上传安全（5 层防护：大小/扩展名/MIME/图片防马/ClamAV 病毒扫描）是否到位
- [ ] 第三方 API 容错（3s 连接/5s 响应头/10s 总 + 3 次指数退避 + 熔断器 5 次/30s + 缓存降级）是否到位
- [ ] 资金对账差异处理（<0.01 自动修复 / 0.01-1 人工审核 / >1 告警+暂停提现）分级是否合理
- [ ] 商家结算发票管理（merchant_invoices + 电子/纸质 + 金蝶/百望 API）是否完整
- [ ] 多端数据同步（WebSocket 实时推送 + 增量同步 last_sync_at + 服务端为准冲突解决）是否一致
- [ ] 蓝绿部署/金丝雀发布（蓝绿 30s DNS 切换 + 金丝雀 5%→20%→50%→100% 四阶段 + Nginx split_clients）是否安全

#### 低优先级 5 项（运维完善度）
- [ ] 监控长期存储（Thanos Sidecar+Query+Compactor+Store + OSS + 5min/1h 降采样）是否满足长期查询
- [ ] 容器资源限制（Docker Compose deploy.resources + OOM 处理 + cAdvisor 监控）是否避免单容器拖垮宿主
- [ ] 连接池配置（PG MaxOpen=100/MaxIdle=10 + Redis PoolSize=50/MinIdle=10）是否匹配并发预估
- [ ] 慢查询优化（EXPLAIN ANALYZE + 索引建议 + SQL 重写避免 SELECT */N+1/大 OFFSET + 缓存策略）是否形成 SOP
- [ ] CDN 缓存策略（三级缓存 CDN→Nginx→Redis + URL/目录/预热刷新 + 命中率监控 >95%）是否到位

---

## 十六、链路追踪与可观测性

> **目标**：实现全链路请求追踪、结构化日志采集、4 级告警体系，达到大厂级可观测性标准，任何一次线上故障可在 5 分钟内定位根因。

### 16.1 链路追踪（OpenTelemetry + Jaeger）

#### 16.1.1 技术选型
| 组件 | 选型 | 理由 |
|------|------|------|
| 采集 SDK | OpenTelemetry-Go | CNCF 标准，厂商无关，后期可切换 backend |
| 后端存储 | Jaeger + Elasticsearch | 开源主流，支持千万级 trace 检索 |
| 采样策略 | Tail-based Sampling | 按错误/慢请求 100% 采样，正常请求 1% 采样 |
| 可视化 | Jaeger UI + Grafana Tempo | 链路查询 + 服务依赖图 |

#### 16.1.2 接入范围（全链路）
```
用户端请求
  → Nginx（生成 trace_id，注入 header X-Trace-Id）
    → Gin API 网关（继承 trace_id）
      → 业务模块 Handler（继承 trace_id）
        → Service 层（继承 trace_id）
          → Repository 层（继承 trace_id，SQL 自动埋点）
          → Redis 调用（继承 trace_id）
          → RabbitMQ 投递（trace_id 注入 message.header）
            → 消费者（从 header 提取 trace_id）
          → 第三方 API（微信支付/高德/AI）
```

#### 16.1.3 Span 设计规范
```go
// 每个 Span 必须包含的标签
type SpanTags struct {
    TraceID      string // 链路 ID
    SpanID       string // 当前 Span ID
    ParentSpanID string // 父 Span ID
    Service      string // 服务名（如 ershou）
    Module       string // 模块名（如 ershou）
    Method       string // 方法名（如 CreateErshou）
    RegionID     int64  // 城市分站
    UserID       int64  // 用户 ID
    ShopID       int64  // 商家 ID
    Status       string // ok/error
    Duration     int64  // 耗时 ms
    Error        string // 错误信息（如有）
}

// 关键节点埋点
1. HTTP 入口（Gin middleware）
2. DB 查询（GORM callback）
3. Redis 操作（hook）
4. MQ 投递/消费（producer/consumer wrapper）
5. 第三方 HTTP 调用（http.Client transport）
6. 分布式锁获取/释放
7. 定时任务执行
```

#### 16.1.4 表设计
```sql
-- trace 采样表（仅采样落库的 trace，原始 trace 存 Jaeger）
CREATE TABLE trace_samples (
  id BIGSERIAL PRIMARY KEY,
  trace_id VARCHAR(64) NOT NULL,
  root_service VARCHAR(50) NOT NULL,
  root_method VARCHAR(100) NOT NULL,
  region_id BIGINT DEFAULT 0,
  user_id BIGINT DEFAULT 0,
  status VARCHAR(10) NOT NULL,           -- ok/error/slow
  duration_ms INT NOT NULL,
  span_count INT DEFAULT 0,
  error_msg VARCHAR(500) DEFAULT '',
  sampled_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP
);
CREATE INDEX idx_trace_status ON trace_samples(status, sampled_at);
CREATE INDEX idx_trace_service ON trace_samples(root_service, status, sampled_at);
```

---

### 16.2 结构化日志体系（Loki + Promtail）

#### 16.2.1 技术选型
| 组件 | 选型 | 理由 |
|------|------|------|
| 日志采集 | Promtail | 与 Loki 原生集成，轻量级 |
| 日志存储 | Loki | 比 ES 成本低 10 倍，按标签索引非全文 |
| 日志查询 | LogQL | 类 PromQL 语法，与 Grafana 统一 |
| 备选方案 | ELK（Elasticsearch + Logstash + Kibana） | 全文检索更强，但成本高 |

> **决策**：MVP 用 Loki，后期数据量上来后可平行接入 ELK 做全文检索。Loki 适合"按标签快速检索"，ELK 适合"按内容全文检索"。

#### 16.2.2 日志规范
```go
// 结构化日志格式（zerolog + JSON）
type LogEntry struct {
    Timestamp   string `json:"timestamp"`    // ISO8601
    Level       string `json:"level"`        // debug/info/warn/error/fatal
    Service     string `json:"service"`      // 服务名
    Module      string `json:"module"`       // 模块名
    TraceID     string `json:"trace_id"`     // 链路 ID（与 Jaeger 联动）
    SpanID      string `json:"span_id"`      // Span ID
    RegionID    int64  `json:"region_id"`    // 城市分站
    UserID      int64  `json:"user_id"`      // 用户 ID
    Msg         string `json:"msg"`          // 日志消息
    Fields      map[string]interface{} `json:"fields"` // 业务字段
    Caller      string `json:"caller"`       // 调用方（file:line）
}

// 日志级别使用规范
// DEBUG: 开发调试用，生产不输出
// INFO:  正常业务流程（用户登录/订单创建/支付成功）
// WARN:  异常但可恢复（重试成功/降级触发/限流触发）
// ERROR: 业务异常（订单失败/支付失败/第三方调用失败）
// FATAL: 系统级异常（DB 连接失败/配置加载失败，触发告警+重启）
```

#### 16.2.3 日志采集架构
```
应用容器（stdout/stderr JSON 日志）
  → Promtail（DaemonSet 采集）
    → Loki（按 service/module/level/region_id 标签索引）
      → Grafana（LogQL 查询 + 日志告警）
        → Alertmanager（告警通知）
```

#### 16.2.4 日志分级存储策略
| 日志类型 | 保留期 | 存储位置 | 查询频率 |
|---------|--------|---------|---------|
| ERROR/FATAL | 90 天 | Loki 热存储 | 高 |
| WARN | 30 天 | Loki 热存储 | 中 |
| INFO | 7 天 | Loki 热存储 | 低 |
| DEBUG | 1 天（仅开发环境） | 本地文件 | 开发期 |
| 访问日志 | 30 天 | Nginx + Loki | 中 |
| 审计日志 | 1 年（合规） | PG audit_logs 表 | 低 |

---

### 16.3 监控告警规则（4 级告警 + 5 渠道通知）

#### 16.3.1 告警级别
| 级别 | 含义 | 响应时效 | 通知渠道 | 示例 |
|------|------|---------|---------|------|
| P0 致命 | 核心服务不可用 | 5 分钟 | 电话+短信+IM+邮件+钉钉 | DB 宕机 / API 全 502 / 支付通道全断 |
| P1 严重 | 核心功能受损 | 15 分钟 | 短信+IM+邮件 | 单模块故障 / 支付成功率<95% / 慢查询>5s |
| P2 警告 | 潜在风险 | 1 小时 | IM+邮件 | CPU>80% / 慢查询>1s / 错误率>1% |
| P3 提示 | 需关注 | 4 小时 | IM | 磁盘>70% / 连接数>50% / 任务延迟 |

#### 16.3.2 告警规则矩阵
```yaml
# Prometheus 告警规则示例
groups:
- name: critical.rules
  rules:
  - alert: ServiceDown
    expr: up == 0
    for: 1m
    labels: { severity: P0 }
    annotations:
      summary: "服务 {{ $labels.service }} 宕机"
      description: "{{ $labels.instance }} 已离线超过 1 分钟"

  - alert: APIErrorRateHigh
    expr: |
      sum(rate(http_requests_total{status=~"5.."}[5m])) by (service, module)
      / sum(rate(http_requests_total[5m])) by (service, module) > 0.05
    for: 2m
    labels: { severity: P1 }
    annotations:
      summary: "{{ $labels.module }} 5xx 错误率 > 5%"

  - alert: DBConnectionsHigh
    expr: pg_stat_database_numbackends > 80
    for: 5m
    labels: { severity: P2 }
    annotations:
      summary: "PG 连接数 > 80"

  - alert: DiskSpaceLow
    expr: (1 - node_filesystem_avail_bytes / node_filesystem_size_bytes) > 0.7
    for: 10m
    labels: { severity: P3 }
    annotations:
      summary: "磁盘使用率 > 70%"

  - alert: PaymentSuccessRateLow
    expr: |
      sum(rate(pay_orders_total{status="success"}[5m]))
      / sum(rate(pay_orders_total[5m])) < 0.95
    for: 5m
    labels: { severity: P1 }
    annotations:
      summary: "支付成功率 < 95%"

  - alert: SlowQueryDetected
    expr: pg_stat_statements_mean_time_seconds > 5
    for: 3m
    labels: { severity: P1 }
    annotations:
      summary: "慢查询 > 5s: {{ $labels.query }}"
```

#### 16.3.3 告警通知渠道
```yaml
# Alertmanager 路由配置
route:
  receiver: default
  group_by: [alertname, service, module]
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 1h
  routes:
  - match: { severity: P0 }
    receiver: critical-p0
    group_wait: 0s
    repeat_interval: 5m
  - match: { severity: P1 }
    receiver: serious-p1
    group_wait: 30s
    repeat_interval: 30m
  - match: { severity: P2 }
    receiver: warning-p2
  - match: { severity: P3 }
    receiver: info-p3

receivers:
- name: critical-p0
  webhook_configs:
  - url: 'http://im-service/api/notify?level=P0'
  - url: 'http://sms-service/api/call?level=P0'  # 电话告警
  - url: 'http://sms-service/api/sms?level=P0'   # 短信告警
  email_configs:
  - to: 'ops@wuchang-tongcheng.com'
- name: serious-p1
  webhook_configs:
  - url: 'http://im-service/api/notify?level=P1'
  email_configs:
  - to: 'dev@wuchang-tongcheng.com'
```

#### 16.3.4 告警收敛与降噪
1. **分组**：同一服务/模块的告警合并为一条
2. **抑制**：P0 触发时抑制同模块 P1/P2 告警
3. **静默**：维护窗口期手动静默
4. **去重**：相同告警 1 小时内不重复通知
5. **聚合**：5 分钟内的关联告警聚合（如 DB 慢导致 API 慢）

---

## 十七、安全防护体系

> **目标**：达到金融级安全标准，防范 OWASP Top 10 全部威胁，通过等保三级测评。

### 17.1 WAF 防护（Web Application Firewall）

#### 17.1.1 部署架构
```
用户请求
  → 云 WAF（阿里云/腾讯云，前置防护）
    → Nginx（ModSecurity 规则，本地防护）
      → Gin 中间件（业务层防护）
        → 业务 Handler
```

#### 17.1.2 防护规则
1. **SQL 注入**：拦截 `UNION SELECT` / `OR 1=1` / `'; DROP TABLE` 等特征
2. **XSS 攻击**：拦截 `<script>` / `javascript:` / `onerror=` 等特征
3. **命令注入**：拦截 `;` / `|` / `&&` / `$()` 等组合
4. **路径穿越**：拦截 `../` / `..\\` / `%2e%2e` 等
5. **文件包含**：拦截 `php://` / `file://` / `data://` 等
6. **CC 攻击**：单 IP 60 秒内请求 > 1000 次自动封禁
7. **爬虫识别**：User-Agent 特征 + 行为分析

### 17.2 SQL 注入防护

#### 17.2.1 ORM 层防护
```go
// ✅ 正确：使用 GORM 参数化查询
db.Where("id = ? AND status = ?", id, status).Find(&result)

// ❌ 错误：字符串拼接（SQL 注入风险）
db.Where(fmt.Sprintf("id = %d AND status = '%s'", id, status)).Find(&result)

// GORM 安全配置
db.Use(
  tracer.New(tracer.Config{
    Enable:      true,
    LogLevel:    logger.Warn,
    SlowThreshold: 200 * time.Millisecond,
  }),
)

// 生产环境强制开启 PrepareStmt（防 SQL 注入 + 提升性能）
db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{
  PrepareStmt: true,
  ConnPool:    connPool,
})
```

#### 17.2.2 输入校验层
```go
// 所有用户输入必须经过 validator 校验
type CreateErshouReq struct {
    Title       string `json:"title" validate:"required,min=2,max=100"`
    Description string `json:"description" validate:"required,max=5000"`
    Price       int64  `json:"price" validate:"required,min=0,max=100000000"`
    CategoryID  int64  `json:"category_id" validate:"required,min=1"`
}

// 业务字段白名单
var allowedSortFields = map[string]bool{
    "created_at": true,
    "price":      true,
    "views":      true,
}
if !allowedSortFields[sortField] {
    return errors.New("invalid sort field")
}
```

### 17.3 XSS / CSRF 防护

#### 17.3.1 XSS 防护
```go
// 1. 输出转义（bluemonday 库）
import "github.com/microcosm-cc/bluemonday"
p := bluemonday.UGCPolicy()
safeHTML := p.Sanitize(userInput)

// 2. Content Security Policy
app.Use(func(c *gin.Context) {
    c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'")
    c.Header("X-Content-Type-Options", "nosniff")
    c.Header("X-Frame-Options", "DENY")
    c.Header("X-XSS-Protection", "1; mode=block")
})

// 3. Cookie 安全
c.SetCookie("token", token, 3600, "/", "", true, true) // httpOnly + secure
```

#### 17.3.2 CSRF 防护
```go
// 1. SameSite Cookie
c.SetSameSite(http.SameSiteStrictMode)

// 2. CSRF Token 双重提交
app.Use(csrf.Middleware(csrf.Options{
    Secret: config.CSRFSecret,
    TokenLookup: "header:X-CSRF-Token",
}))

// 3. Referer 校验
app.Use(func(c *gin.Context) {
    referer := c.Request.Header.Get("Referer")
    if !isAllowedReferer(referer) {
        c.AbortWithStatus(403)
        return
    }
})
```

### 17.4 DDoS 防护

#### 17.4.1 多层防护
```
第 1 层：云防护（阿里云 DDoS 高防 / 腾讯云大禹）
  - SYN Flood / UDP Flood / ICMP Flood 防护
  - 流量清洗（100G+ 攻击自动牵引）

第 2 层：Nginx 限流
  - limit_req_zone：单 IP 100 req/s
  - limit_conn_zone：单 IP 并发 50
  - limit_req_status 429

第 3 层：Gin 中间件限流
  - 令牌桶限流（按用户/IP/接口维度）
  - 滑动窗口限流（按模块维度）

第 4 层：Redis 分布式限流
  - 跨实例统一限流
  - 用户级：1000 req/min
  - 接口级：100 req/min
```

### 17.5 接口签名与防重放

#### 17.5.1 API 签名规范
```go
// 移动端 / 小程序 API 必须签名
// 签名算法：HMAC-SHA256
// 签名要素：method + path + timestamp + nonce + body_md5 + secret

type APIRequest struct {
    Method    string // GET/POST/PUT/DELETE
    Path      string // /api/v1/ershou/create
    Timestamp int64  // Unix 时间戳（秒）
    Nonce     string // 随机字符串（32 位）
    BodyMD5   string // 请求体 MD5
    Signature string // HMAC-SHA256 签名
}

// 服务端校验
func VerifySignature(req APIRequest, secret string) error {
    // 1. 时间戳校验（±5 分钟）
    if math.Abs(float64(time.Now().Unix() - req.Timestamp)) > 300 {
        return errors.New("timestamp expired")
    }

    // 2. Nonce 防重放（Redis 缓存 10 分钟）
    if exists, _ := redis.Exists(ctx, "nonce:"+req.Nonce).Result(); exists {
        return errors.New("nonce replay detected")
    }
    redis.Set(ctx, "nonce:"+req.Nonce, 1, 10*time.Minute)

    // 3. 签名校验
    signString := fmt.Sprintf("%s\n%s\n%d\n%s\n%s",
        req.Method, req.Path, req.Timestamp, req.Nonce, req.BodyMD5)
    expectedSig := hmacSHA256(signString, secret)
    if !hmac.Equal([]byte(req.Signature), []byte(expectedSig)) {
        return errors.New("invalid signature")
    }
    return nil
}
```

### 17.6 敏感数据加密

#### 17.6.1 数据加密规范
| 数据类型 | 加密方式 | 礘钥管理 |
|---------|---------|---------|
| 用户密码 | bcrypt（cost=12） | 无需密钥 |
| 手机号 | AES-256-GCM（存储）+ 虚拟号（展示） | KMS 管理 |
| 身份证号 | AES-256-GCM（存储）+ 脱敏（展示） | KMS 管理 |
| 银行卡号 | AES-256-GCM（不存储，仅支付通道） | - |
| API 密钥 | AES-256-GCM（配置文件加密） | KMS 管理 |
| JWT Token | RS256（非对称签名） | RSA 密钥对 |

---

## 十八、分布式锁与缓存防护

> **目标**：金融级防重场景全覆盖，高并发缓存三件套防护到位，杜绝超卖/重复下单/缓存雪崩。

### 18.1 分布式锁（Redis RedLock）

#### 18.1.1 锁的实现
```go
// 基于 Redis 的分布式锁（RedLock 算法简化版）
type DistributedLock struct {
    client *redis.Client
    key    string
    value  string // UUID，防误删
    ttl    time.Duration
}

func (l *DistributedLock) Acquire() (bool, error) {
    l.value = uuid.New().String()
    result, err := l.client.SetNX(ctx, l.key, l.value, l.ttl).Result()
    return result, err
}

func (l *DistributedLock) Release() error {
    // Lua 脚本保证原子性（防止误删他人锁）
    script := `
    if redis.call("GET", KEYS[1]) == ARGV[1] then
        return redis.call("DEL", KEYS[1])
    else
        return 0
    end
    `
    _, err := l.client.Eval(ctx, script, []string{l.key}, l.value).Result()
    return err
}

func (l *DistributedLock) Renew() (bool, error) {
    // 续期（业务执行时间 > TTL 时）
    script := `
    if redis.call("GET", KEYS[1]) == ARGV[1] then
        return redis.call("EXPIRE", KEYS[1], ARGV[2])
    else
        return 0
    end
    `
    result, err := l.client.Eval(ctx, script, []string{l.key}, l.value, l.ttl.Seconds()).Result()
    return result.(int64) == 1, err
}
```

#### 18.1.2 金融场景应用
```go
// 场景1：下单锁（防重复下单）
func CreateOrder(userID, ershouID int64) error {
    lock := &DistributedLock{client: redis, key: fmt.Sprintf("order:create:%d:%d", userID, ershouID), ttl: 10 * time.Second}
    acquired, _ := lock.Acquire()
    if !acquired {
        return errors.New("正在处理中，请勿重复提交")
    }
    defer lock.Release()
    // ... 下单业务逻辑
}

// 场景2：提现锁（防重复提现）
func Withdraw(userID int64, amount decimal.Decimal) error {
    lock := &DistributedLock{client: redis, key: fmt.Sprintf("withdraw:%d", userID), ttl: 30 * time.Second}
    acquired, _ := lock.Acquire()
    if !acquired {
        return errors.New("提现处理中，请稍后")
    }
    defer lock.Release()
    // ... 提现业务逻辑
}

// 场景3：库存扣减锁（防超卖）
func DeductStock(ershouID int64) error {
    lock := &DistributedLock{client: redis, key: fmt.Sprintf("stock:%d", ershouID), ttl: 5 * time.Second}
    acquired, _ := lock.Acquire()
    if !acquired {
        return errors.New("商品火爆，请重试")
    }
    defer lock.Release()
    // ... 库存校验 + 扣减
}

// 场景4：定时任务锁（防多实例重复执行）
func (s *CronService) RunJob(jobName string) {
    lock := &DistributedLock{client: redis, key: "cron:" + jobName, ttl: 5 * time.Minute}
    acquired, _ := lock.Acquire()
    if !acquired {
        return // 其他实例正在执行
    }
    defer lock.Release()
    // ... 任务执行
}
```

### 18.2 缓存防护三件套

#### 18.2.1 缓存穿透防护（布隆过滤器）
```go
// 场景：查询不存在的 ershou_id，每次都打到 DB
// 方案：布隆过滤器前置拦截

import "github.com/bits-and-blooms/bloom/v3"

type BloomFilter struct {
    filter *bloom.BloomFilter
}

// 启动时加载所有 ershou_id
func (bf *BloomFilter) Init(db *gorm.DB) error {
    var ids []int64
    db.Model(&Ershou{}).Pluck("id", &ids)
    bf.filter = bloom.NewWithEstimates(uint(len(ids)), 0.001) // 0.1% 误判率
    for _, id := range ids {
        bf.filter.AddInt64(id)
    }
    return nil
}

// 查询时先过布隆过滤器
func GetErshou(id int64) (*Ershou, error) {
    if !bloomFilter.TestInt64(id) {
        return nil, errors.New("商品不存在") // 不查 DB
    }
    // ... 查 Redis + DB
}

// 新增时同步更新布隆过滤器
func CreateErshou(ershou *Ershou) error {
    db.Create(ershou)
    bloomFilter.AddInt64(ershou.ID)
    return nil
}
```

#### 18.2.2 缓存击穿防护（互斥锁 + 永不过期）
```go
// 场景：热点 key 失效瞬间，大量请求打到 DB
// 方案1：互斥锁（singleflight）
import "golang.org/x/sync/singleflight"

var group singleflight.Group

func GetHotErshou(id int64) (*Ershou, error) {
    key := fmt.Sprintf("ershou:%d", id)
    // singleflight 保证同一 key 同时只有一个请求查 DB
    result, err, _ := group.Do(key, func() (interface{}, error) {
        // 查 Redis
        if cached, _ := redis.Get(ctx, key).Result(); cached != "" {
            return json.Unmarshal(cached)
        }
        // 查 DB
        ershou, err := db.Find(id)
        if err != nil {
            return nil, err
        }
        // 写 Redis（随机过期时间，防雪崩）
        ttl := 300 + rand.Intn(60) // 300-360 秒
        redis.Set(ctx, key, ershou, time.Duration(ttl)*time.Second)
        return ershou, nil
    })
    return result.(*Ershou), err
}

// 方案2：永不过期 + 异步刷新（适合超热点）
func GetSuperHotErshou(id int64) (*Ershou, error) {
    key := fmt.Sprintf("ershou:hot:%d", id)
    // Redis 永不过期，后台 goroutine 定时刷新
    if cached, _ := redis.Get(ctx, key).Result(); cached != "" {
        // 检查是否需要刷新（last_refresh_at > 5 分钟前）
        if needsRefresh(key) {
            go refreshCache(key, id) // 异步刷新
        }
        return json.Unmarshal(cached)
    }
    // 冷启动：查 DB 并写入
    return loadAndCache(key, id)
}
```

#### 18.2.3 缓存雪崩防护（随机过期 + 多级缓存）
```go
// 场景：大量 key 同时失效，DB 瞬间压力暴增
// 方案1：随机过期时间
func SetCache(key string, value interface{}, baseTTL time.Duration) {
    // 基础 TTL + 随机偏移（避免同时失效）
    ttl := baseTTL + time.Duration(rand.Intn(120))*time.Second
    redis.Set(ctx, key, value, ttl)
}

// 方案2：多级缓存（本地 + 分布式）
type MultiLevelCache struct {
    localCache *ristretto.Cache  // 本地 LRU 缓存（进程内）
    redis      *redis.Client     // 分布式缓存
}

func (c *MultiLevelCache) Get(key string) (interface{}, error) {
    // 1. 查本地缓存（1ms）
    if val, ok := c.localCache.Get(key); ok {
        return val, nil
    }
    // 2. 查 Redis（5ms）
    if val, err := c.redis.Get(ctx, key).Result(); err == nil {
        c.localCache.Set(key, val, 1) // 回填本地
        return val, nil
    }
    // 3. 查 DB（50ms）
    return nil, errors.New("not found")
}

// 方案3：熔断降级（DB 压力大时返回降级数据）
func GetErshouList(regionID int64) ([]*Ershou, error) {
    cacheKey := fmt.Sprintf("ershou:list:%d", regionID)
    if cached, _ := redis.Get(ctx, cacheKey).Result(); cached != "" {
        return json.Unmarshal(cached)
    }
    // DB 熔断检查
    if circuitBreaker.IsOpen("db") {
        return getDefaultErshouList(), nil // 返回降级数据
    }
    // 查 DB
    list, err := db.FindByRegion(regionID)
    if err != nil {
        return nil, err
    }
    SetCache(cacheKey, list, 5*time.Minute)
    return list, nil
}
```

---

## 十九、消息幂等与全链路一致性

> **目标**：金融级全链路幂等，消息消费不重复、接口调用不重复、业务操作不重复，杜绝资损。

### 19.1 接口幂等

#### 19.1.1 幂等设计规范
```go
// 方案1：客户端幂等号（前端生成 UUID）
type CreateOrderReq struct {
    IdempotentKey string `json:"idempotent_key" header:"X-Idempotent-Key"` // UUID
    ErshouID      int64  `json:"ershou_id"`
    // ...
}

func CreateOrder(c *gin.Context) {
    var req CreateOrderReq
    c.ShouldBindJSON(&req)
    if req.IdempotentKey == "" {
        c.JSON(400, gin.H{"msg": "missing idempotent key"})
        return
    }

    // 幂等校验（Redis SETNX）
    key := fmt.Sprintf("idem:%s", req.IdempotentKey)
    acquired, _ := redis.SetNX(ctx, key, "1", 24*time.Hour).Result()
    if !acquired {
        // 已处理过，返回上次结果
        result, _ := redis.Get(ctx, key+":result").Result()
        c.JSON(200, json.Unmarshal(result))
        return
    }

    // 执行业务
    order, err := orderService.Create(req)
    if err != nil {
        redis.Set(ctx, key+":result", err.Error(), 24*time.Hour)
        c.JSON(500, gin.H{"msg": err.Error()})
        return
    }
    redis.Set(ctx, key+":result", json.Marshal(order), 24*time.Hour)
    c.JSON(200, order)
}

// 方案2：业务唯一键（DB 唯一索引）
type PayOrder struct {
    OutTradeNo string `gorm:"uniqueIndex"` // 业务幂等号（业务规则生成）
    UserID     int64
    Amount     decimal.Decimal
}

// 方案3：状态机幂等（订单只能从待支付→已支付，重复请求无效）
func PayOrder(outTradeNo string) error {
    result := db.Model(&Order{}).
        Where("out_trade_no = ? AND status = ?", outTradeNo, StatusPending).
        Update("status", StatusPaid)
    if result.RowsAffected == 0 {
        return errors.New("order already paid or not found")
    }
    return nil
}
```

### 19.2 消息消费幂等

#### 19.2.1 消息幂等表
```sql
-- 消息消费记录表（防重复消费）
CREATE TABLE message_idempotent (
  id BIGSERIAL PRIMARY KEY,
  consumer_group VARCHAR(50) NOT NULL,    -- 消费组
  message_id VARCHAR(64) NOT NULL,        -- 消息 ID
  status VARCHAR(20) NOT NULL,            -- processing/done/failed
  result TEXT,                            -- 处理结果（JSON）
  retry_count INT DEFAULT 0,
  consumed_at TIMESTAMP,
  created_at TIMESTAMP,
  UNIQUE(consumer_group, message_id)      -- 唯一约束保证幂等
);
CREATE INDEX idx_idem_status ON message_idempotent(status, consumed_at);
```

#### 19.2.2 消费者幂等实现
```go
type MessageConsumer struct {
    db    *gorm.DB
    group string
}

func (c *MessageConsumer) Consume(msg Message) error {
    // 1. 幂等校验
    record := &MessageIdempotent{
        ConsumerGroup: c.group,
        MessageID:     msg.ID,
        Status:        "processing",
    }
    if err := c.db.Create(record).Error; err != nil {
        if isDuplicateKeyError(err) {
            // 已消费过，返回上次结果
            c.db.Where("consumer_group = ? AND message_id = ?", c.group, msg.ID).First(record)
            if record.Status == "done" {
                return nil // 幂等成功
            }
            return errors.New("previous consumption failed, needs retry")
        }
        return err
    }

    // 2. 执行业务
    result, err := c.handleBusiness(msg)

    // 3. 更新状态
    status := "done"
    if err != nil {
        status = "failed"
        if record.RetryCount >= 3 {
            status = "dead" // 死信
        }
    }
    c.db.Model(record).Updates(map[string]interface{}{
        "status":       status,
        "result":       json.Marshal(result),
        "consumed_at":  time.Now(),
        "retry_count":  record.RetryCount + 1,
    })
    return err
}
```

### 19.3 业务幂等场景清单

| 场景 | 幂等方案 | TTL |
|------|---------|-----|
| 创建订单 | 客户端幂等号 + Redis SETNX | 24h |
| 支付回调 | out_trade_no 唯一索引 + 状态机 | 永久 |
| 提现申请 | 用户锁 + 业务流水号唯一索引 | 永久 |
| 退款申请 | 退款单号唯一索引 | 永久 |
| 分账执行 | 分账记录唯一索引 | 永久 |
| 发红包 | 红包单号唯一索引 | 永久 |
| 签到 | (user_id + date) 联合唯一索引 | 永久 |
| 领券 | (user_id + coupon_id) 联合唯一索引 | 永久 |
| 发帖 | 客户端幂等号 + Redis SETNX | 1h |
| 评论 | 客户端幂等号 + Redis SETNX | 1h |
| 消息消费 | 消息幂等表 | 永久 |
| 定时任务 | 分布式锁 | 任务执行期 |

---

## 二十、配置中心与 API 治理

> **目标**：实现配置热更新、API 版本演进管理、API 文档自动化，达到大厂级 API 治理标准。

### 20.1 配置中心（第 13 个中台 - config）

#### 20.1.1 配置中心定位
作为 v3 的第 13 个中台（独立于 tenant 分站配置），专注于：
- 应用级配置（DB/Redis/MQ 连接信息）
- 业务级配置（开关/阈值/规则）
- 中台级配置（每个中台的参数）
- 模块级配置（每个垂直业务模块的参数）
- 灰度级配置（按城市/用户/版本的差异化配置）

> **与 tenant 中台的区别**：tenant 中台管"分站运营配置"（佣金/价格/风控阈值），config 中台管"系统配置"（DB 连接/开关/灰度规则）。

#### 20.1.2 配置中心架构
```
配置中心（config 中台）
  ├─ config_apps         # 应用级配置（DB/Redis/MQ 连接）
  ├─ config_business     # 业务级配置（全局开关/阈值）
  ├─ config_modules      # 模块级配置（每个模块独立配置）
  ├─ config_stations     # 分站级配置（继承自 tenant，覆盖模块配置）
  ├─ config_grayscales   # 灰度配置（按城市/用户/版本差异化）
  └─ config_audit_logs   # 配置变更审计日志
```

#### 20.1.3 表设计
```sql
-- 配置项定义表
CREATE TABLE config_items (
  id BIGSERIAL PRIMARY KEY,
  config_key VARCHAR(100) UNIQUE NOT NULL,
  config_name VARCHAR(100) NOT NULL,
  config_value TEXT NOT NULL,
  value_type VARCHAR(20) NOT NULL,          -- string/int/bool/json
  default_value TEXT,
  category VARCHAR(50) NOT NULL,            -- app/business/module/station/gray
  module VARCHAR(50) DEFAULT '',            -- 所属模块（business 时为空）
  description VARCHAR(500) DEFAULT '',
  is_editable BOOLEAN DEFAULT TRUE,
  is_sensitive BOOLEAN DEFAULT FALSE,       -- 敏感配置（显示 ***）
  validation_rule VARCHAR(200) DEFAULT '',
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);

-- 配置灰度规则表
CREATE TABLE config_grayscales (
  id BIGSERIAL PRIMARY KEY,
  config_key VARCHAR(100) NOT NULL,
  config_value TEXT NOT NULL,
  rule_type VARCHAR(20) NOT NULL,           -- city/user/version/percent
  rule_value JSON NOT NULL,                 -- {"cities": [1,2,3]} 或 {"percent": 10}
  priority INT DEFAULT 0,
  start_at TIMESTAMP,
  end_at TIMESTAMP,
  status INT DEFAULT 0,                     -- 0未启用 1启用 2已结束
  created_at TIMESTAMP
);

-- 配置变更审计表
CREATE TABLE config_audit_logs (
  id BIGSERIAL PRIMARY KEY,
  config_key VARCHAR(100) NOT NULL,
  old_value TEXT,
  new_value TEXT,
  operator_id BIGINT NOT NULL,
  operator_name VARCHAR(50) NOT NULL,
  operation VARCHAR(20) NOT NULL,           -- create/update/delete
  reason VARCHAR(500) DEFAULT '',
  created_at TIMESTAMP
);
CREATE INDEX idx_config_audit_key ON config_audit_logs(config_key, created_at);
```

#### 20.1.4 配置热更新机制
```go
// 配置监听（基于 Redis Pub/Sub）
type ConfigWatcher struct {
    redis *redis.Client
    cache *ristretto.Cache
}

func (w *ConfigWatcher) Subscribe() {
    pubsub := w.redis.Subscribe(ctx, "config:change")
    for {
        msg, err := pubsub.ReceiveMessage(ctx)
        if err != nil {
            continue
        }
        // 收到变更通知，刷新本地缓存
        configKey := msg.Payload
        w.cache.Del(configKey)
        // 重新加载配置
        w.loadConfig(configKey)
        // 通知业务模块（观察者模式）
        eventBus.Publish("config:"+configKey, newValue)
    }
}

// 配置变更广播
func (s *ConfigService) Update(key, value string) error {
    db.Model(&ConfigItem{}).Where("config_key = ?", key).Update("config_value", value)
    // 广播变更通知
    s.redis.Publish(ctx, "config:change", key)
    // 记录审计
    s.recordAudit(key, oldValue, value)
    return nil
}
```

### 20.2 API 版本管理

#### 20.2.1 版本策略
```
/api/v1/ershou/create      # 当前版本
/api/v2/ershou/create      # 新版本（与 v1 共存 3 个月）
/api/v1/ershou/list        # v1 保留但标记 deprecated
```

#### 20.2.2 版本管理实现
```go
// 路由分组
v1 := router.Group("/api/v1")
{
    v1.POST("/ershou/create", ershouHandler.Create)
    v1.GET("/ershou/list", ershouHandler.List)
}

v2 := router.Group("/api/v2")
{
    v2.POST("/ershou/create", ershouHandlerV2.Create) // 新版本
    v2.GET("/ershou/list", ershouHandlerV2.List)      // 新版本
}

// 版本兼容策略
// 1. 新增字段：v2 直接新增，v1 不返回该字段
// 2. 删除字段：v2 删除，v1 保留 3 个月（标记 deprecated）
// 3. 修改字段类型：v2 新增字段，v1 保留旧字段，3 个月后下线 v1
// 4. 破坏性变更：必须升大版本（v1 → v2），保留 v1 6 个月
```

#### 20.2.3 版本生命周期
| 阶段 | 状态 | 通知方式 | 时长 |
|------|------|---------|------|
| 发布新版本 | active | API 文档标注 | - |
| 旧版本标记废弃 | deprecated | 响应头 `Deprecation: true` + 日志告警 | 3 个月 |
| 旧版本下线 | sunset | 返回 410 Gone + 邮件通知 | 永久 |

### 20.3 API 文档生成（Swagger/OpenAPI）

#### 20.3.1 Swagger 集成
```go
import swaggerFiles "github.com/swaggo/files"
import ginSwagger "github.com/swaggo/gin-swagger"

// 1. 注解生成文档
// @Summary 创建二手商品
// @Description 创建一个新的二手商品
// @Tags 二手交易
// @Accept json
// @Produce json
// @Param body body CreateErshouReq true "请求体"
// @Success 200 {object} ErshouResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/ershou/create [post]
func CreateErshou(c *gin.Context) {
    // ...
}

// 2. 启动时生成文档
//go:generate swag init -g cmd/main.go -o docs --parseDependency --parseInternal

// 3. 暴露文档接口
router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

// 4. 访问地址
// 开发环境：http://localhost:8088/swagger/index.html
// 生产环境：https://api.wuchang-tongcheng.com/swagger/index.html（需鉴权）
```

#### 20.3.2 文档规范
1. **每个接口必须包含**：Summary / Description / Tags / Param / Success / Failure / Router
2. **Tags 按模块分组**：二手交易 / 招聘 / 房产 / 商城 / 支付 / IM / ...
3. **请求/响应必须定义 Model**：使用 `@Param body body XXXReq true` 和 `@Success 200 {object} XXXResponse`
4. **枚举值必须列出**：在 Description 中列出所有可选值
5. **错误码必须定义**：统一错误码表，文档中引用

---

## 二十一、测试与质量保障

> **目标**：建立四层测试体系，核心业务覆盖率 > 80%，上线前通过性能压测，达到大厂级质量标准。

### 21.1 测试策略四层体系

```
第 1 层：单元测试（unit）
  - 覆盖率目标：核心业务 > 80%，工具类 > 90%
  - 工具：Go testing + testify + gomock
  - 范围：Service 层 / Repository 层 / 工具函数

第 2 层：集成测试（integration）
  - 覆盖率目标：核心流程 100% 覆盖
  - 工具：testcontainers-go（启动真实 PG/Redis）
  - 范围：跨层调用（Handler → Service → Repository → DB）

第 3 层：E2E 测试（end-to-end）
  - 覆盖率目标：核心业务流程 100% 覆盖
  - 工具：Playwright（前端）+ API 自动化（后端）
  - 范围：用户完整流程（注册 → 发布 → 下单 → 支付 → 评价）

第 4 层：性能压测（performance）
  - 时机：上线前 + 每月例行
  - 工具：k6 / wrk / locust
  - 范围：核心接口 QPS / 并发 / 响应时间 / 资源占用
```

### 21.2 单元测试规范

```go
// Service 层单元测试（mock Repository）
func TestErshouService_Create(t *testing.T) {
    // Given
    mockRepo := new(MockErshouRepo)
    mockRepo.On("Create", mock.Anything).Return(&Ershou{ID: 1}, nil)
    service := NewErshouService(mockRepo)

    // When
    req := &CreateErshouReq{Title: "iPhone 15", Price: 5000}
    result, err := service.Create(req)

    // Then
    assert.NoError(t, err)
    assert.Equal(t, int64(1), result.ID)
    mockRepo.AssertExpectations(t)
}

// Repository 层单元测试（testcontainers）
func TestErshouRepo_FindByID(t *testing.T) {
    // 启动真实 PG 容器
    pgContainer, _ := postgres.RunContainer(ctx)
    defer pgContainer.Terminate(ctx)

    db, _ := gorm.Open(postgres.Open(pgContainer.ConnectionString()))
    db.AutoMigrate(&Ershou{})

    // Given
    db.Create(&Ershou{ID: 1, Title: "iPhone 15"})

    // When
    repo := NewErshouRepo(db)
    result, err := repo.FindByID(1)

    // Then
    assert.NoError(t, err)
    assert.Equal(t, "iPhone 15", result.Title)
}
```

### 21.3 集成测试规范

```go
// 完整下单流程集成测试
func TestOrderFlow_Integration(t *testing.T) {
    // 启动测试环境
    app := setupTestApp()
    defer app.Teardown()

    // 1. 用户注册登录
    token := app.RegisterAndLogin("13800138000", "123456")

    // 2. 发布二手商品
    ershouID := app.PublishErshou(token, &CreateErshouReq{
        Title: "iPhone 15", Price: 5000,
    })

    // 3. 创建订单
    orderID := app.CreateOrder(token, ershouID)

    // 4. 模拟支付回调
    app.MockPayCallback(orderID, "success")

    // 5. 验证订单状态
    order := app.GetOrder(token, orderID)
    assert.Equal(t, StatusPaid, order.Status)

    // 6. 验证钱包余额变动
    wallet := app.GetWallet(token)
    assert.Equal(t, decimal.NewFromFloat(5000), wallet.Balance)
}
```

### 21.4 E2E 测试规范

```javascript
// Playwright E2E 测试示例
const { test, expect } = require('@playwright/test')

test('用户完整下单流程', async ({ page }) => {
  // 1. 登录
  await page.goto('http://localhost:5178/login')
  await page.fill('[placeholder="手机号"]', '13800138000')
  await page.fill('[placeholder="密码"]', '123456')
  await page.click('button:has-text("登录")')

  // 2. 浏览二手商品
  await page.click('text=二手')
  await page.click('text=iPhone 15')

  // 3. 下单
  await page.click('button:has-text("立即购买")')
  await page.click('button:has-text("确认下单")')

  // 4. 支付
  await page.click('text=微信支付')
  await page.click('button:has-text("确认支付")')

  // 5. 验证订单
  await page.click('text=我的订单')
  await expect(page.locator('text=已支付')).toBeVisible()
})
```

### 21.5 性能压测方案

#### 21.5.1 压测工具选型
| 工具 | 适用场景 | 优势 |
|------|---------|------|
| k6 | HTTP API 压测 | 脚本化（JS）+ 云端分布式 |
| wrk | 简单接口压测 | 轻量级，单机高并发 |
| locust | 复杂场景压测 | Python 脚本，UI 可视化 |

#### 21.5.2 压测指标标准
| 接口类型 | 目标 QPS | P99 响应时间 | 错误率 |
|---------|---------|-------------|--------|
| 列表查询 | 2000 | < 200ms | < 0.1% |
| 详情查询 | 3000 | < 100ms | < 0.1% |
| 创建订单 | 500 | < 500ms | < 0.01% |
| 支付回调 | 1000 | < 300ms | < 0.01% |
| 搜索接口 | 1000 | < 500ms | < 0.5% |

#### 21.5.3 压测脚本示例
```javascript
// k6 压测脚本
import http from 'k6/http'
import { check, sleep } from 'k6'

export let options = {
  stages: [
    { duration: '2m', target: 100 },   // 2 分钟内升到 100 VU
    { duration: '5m', target: 100 },   // 保持 100 VU 5 分钟
    { duration: '2m', target: 500 },   // 升到 500 VU
    { duration: '5m', target: 500 },   // 保持 500 VU 5 分钟
    { duration: '2m', target: 0 },     // 降到 0
  ],
  thresholds: {
    http_req_duration: ['p(99)<500'],   // 99% 请求 < 500ms
    http_req_failed: ['rate<0.01'],     // 错误率 < 1%
  },
}

export default function () {
  let res = http.get('http://localhost:8088/api/v1/ershou/list?region_id=1')
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 200ms': (r) => r.timings.duration < 200,
  })
  sleep(1)
}
```

---

## 二十二、数据库高可用与扩展

> **目标**：数据库支持读写分离、分库分表、同城双活+异地灾备，满足金融级数据安全与高可用要求。

### 22.1 读写分离

#### 22.1.1 读写分离架构
```
应用层
  ↓
GORM（多数据源路由）
  ├─ Master（写）  → PG 主库（读写）
  └─ Slaves（读）  → PG 从库 1/2/3（只读）
```

#### 22.1.2 实现方案
```go
// GORM 读写分离
import "gorm.io/plugin/dbresolver"

db.Use(
    dbresolver.Register(dbresolver.Config{
        Sources:      []*gorm.DB{masterDB},           // 写库
        Replicas:     []*gorm.DB{slave1DB, slave2DB}, // 读库
        Policy:       dbresolver.RandomPolicy{},       // 负载均衡策略
        TraceResolverMode: true,                       // 日志记录使用的主从
    }).
    Register(dbresolver.Config{
        Sources:  []*gorm.DB{payMasterDB},            // pay 中台专用主库
        Replicas: []*gorm.DB{paySlave1DB, paySlave2DB},
    }, &PayOrder{}, &WalletAccount{}, &SettleRecord{}), // 指定表使用专用数据源
)
```

#### 22.1.3 强制走主库场景
```go
// 场景：写后立即读（主从延迟可能导致读不到）
func CreateAndQueryErshou(req *CreateErshouReq) (*Ershou, error) {
    ershou, err := repo.Create(req) // 写主库
    // 强制走主库读
    return repo.FindByIDMaster(ershou.ID) // 使用 db.Clauses(clause.ReadWrites{Read: "primary"})
}
```

### 22.2 分库分表策略

#### 22.2.1 分库分表时机
| 指标 | 阈值 | 动作 |
|------|------|------|
| 单表数据量 | > 1000 万 | 分表 |
| 单库 QPS | > 5000 | 分库 |
| 单库磁盘 | > 500GB | 分库 |
| 单表索引大小 | > 5GB | 分表 |

#### 22.2.2 分片策略
```go
// 分片键选择
// 1. 用户维度：user_id（用户相关表）
// 2. 城市维度：region_id（业务数据表）
// 3. 时间维度：created_at（日志/行为表）

// 分片规则
// - user_id % 16 → 16 个分表（user_0 ~ user_15）
// - region_id % 8 → 8 个分库（db_0 ~ db_7）
// - created_at 按月分表 → ershou_202601 / ershou_202602

// GORM 分表插件
import "github.com/go-gorm/sharding"

db.Use(sharding.Register(sharding.Config{
    ShardingKey:         "user_id",
    NumberOfShards:      16,
    PrimaryKeyGenerator: sharding.PKSnowflake,
}, &Ershou{}, &Order{}, &WalletTransaction{}))

// 查询时自动路由到正确分表
db.Create(&Ershou{UserID: 1001}) // 写入 ershou_9 (1001 % 16 = 9)
db.Where("user_id = ?", 1001).Find(&ershou) // 查 ershou_9
```

#### 22.2.3 跨分片查询处理
```go
// 场景：按城市查询所有二手商品（跨 user_id 分片）
// 方案1：ES 异步索引（推荐）
// - 写入时同步到 ES，按 region_id 索引
// - 查询时走 ES，不走分片 DB
db.Create(&Ershou{}) // 同时写入 DB + ES

// 方案2：广播查询（仅小数据量）
var allErshou []*Ershou
for i := 0; i < 16; i++ {
    var list []*Ershou
    db.Table(fmt.Sprintf("ershou_%d", i)).Where("region_id = ?", regionID).Find(&list)
    allErshou = append(allErshou, list...)
}
```

### 22.3 灾备方案

#### 22.3.1 灾备架构
```
【同城双活】（RPO=0, RTO<1min）
  北京机房 A（主）
    ├─ PG 主库
    ├─ Redis 主
    └─ 应用集群
       ↓ 同步复制（PG 流复制 + Redis 主从）
  北京机房 B（备）
    ├─ PG 从库（热备，可切换为主）
    ├─ Redis 从
    └─ 应用集群

【异地灾备】（RPO<1min, RTO<1h）
  北京机房（主）
    ↓ 异步复制（PG 逻辑复制 + OSS 跨区复制）
  上海机房（灾备）
    ├─ PG 灾备库（只读）
    ├─ Redis 灾备
    └─ 应用集群（冷启动）
```

#### 22.3.2 故障切换流程
```yaml
# 故障切换自动化（Patroni + etcd）
1. 主库故障检测（3 秒内）
2. etcd 选主（5 秒内）
3. 从库提升为主（10 秒内）
4. 应用连接切换（30 秒内）
5. 告警通知（1 分钟内）

# 切换条件
- 主库无响应 > 10 秒
- 主库磁盘满
- 主库 CPU > 95% 持续 1 分钟
- 主库连接数 > 90%
```

#### 22.3.3 数据备份策略
```yaml
# 3-2-1 备份策略
3 份数据：
  - 生产库（在线）
  - 同城备库（热备）
  - 异地备库（温备）

2 种介质：
  - PG 流复制（实时）
  - OSS 快照（每日）

1 份离线：
  - 每周全量备份到 OSS 归档（不可篡改）

# 备份保留期
- 实时备份：永久（主从复制）
- 每日全量：30 天
- 每周全量：90 天
- 每月全量：1 年
- 金融数据：永久（合规要求）
```

---

## 二十三、微服务演进路径

> **目标**：当前插件化单体足够支撑 10 万日活，但当流量增长时可平滑演进为微服务架构，业务代码零修改。

### 23.1 演进时机与路径

#### 23.1.1 演进触发条件
| 指标 | 阈值 | 触发动作 |
|------|------|---------|
| 日活用户 | > 10 万 | 评估拆分 |
| 单服务 QPS | > 5000 | 拆分高负载模块 |
| 团队规模 | > 20 人 | 按业务域拆分 |
| 部署冲突 | 频繁 | 拆分独立部署 |
| 故障影响 | 单模块故障影响全站 | 拆分故障隔离 |

#### 23.1.2 演进三阶段
```
阶段 1：插件化单体（当前）
  - 所有模块在一个进程
  - 共享 DB / Redis / MQ
  - Plugin 接口解耦
  ↓

阶段 2：核心模块微服务化（中期）
  - 拆出 pay 中台（独立部署）
  - 拆出 im 中台（独立部署）
  - 拆出高负载垂直业务（ershou / toutiao）
  - 服务间走 gRPC
  ↓

阶段 3：全微服务化（远期）
  - 所有中台 + 垂直业务独立部署
  - 服务网格（Istio）治理
  - K8s 编排
  - 多机房多活
```

### 23.2 服务注册发现（预留）

#### 23.2.1 技术选型
| 方案 | 适用场景 | 选型理由 |
|------|---------|---------|
| Consul | 中小规模 | Go 原生，集成简单 |
| Nacos | 中大规模 | 阿里开源，配置+注册一体化 |
| K8s Service | K8s 环境 | 原生集成，无需额外组件 |
| etcd | 自建场景 | 强一致性，性能好 |

> **决策**：MVP 不引入，预留接口。后期拆分时优先用 Consul（与 Go 生态契合）。

#### 23.2.2 接口预留
```go
// 服务发现接口（当前为 mock，后期接入 Consul）
type ServiceDiscovery interface {
    Register(serviceName, instanceID, addr string, port int, metadata map[string]string) error
    Deregister(instanceID string) error
    Discover(serviceName string) ([]*ServiceInstance, error)
    Watch(serviceName string) (<-chan []*ServiceInstance, error)
}

// 当前实现（单体返回自身）
type LocalDiscovery struct{}

func (d *LocalDiscovery) Discover(serviceName string) ([]*ServiceInstance, error) {
    return []*ServiceInstance{
        {ServiceName: serviceName, Addr: "127.0.0.1", Port: 8088},
    }, nil
}

// 后期实现（接入 Consul）
type ConsulDiscovery struct {
    client *consul.Client
}

func (d *ConsulDiscovery) Discover(serviceName string) ([]*ServiceInstance, error) {
    entries, _ := d.client.Health().Service(serviceName, "", true, nil)
    var instances []*ServiceInstance
    for _, entry := range entries {
        instances = append(instances, &ServiceInstance{
            ServiceName: serviceName,
            Addr:        entry.Service.Address,
            Port:        entry.Service.Port,
        })
    }
    return instances, nil
}
```

### 23.3 微服务拆分策略

#### 23.3.1 拆分原则
1. **按业务域拆分**：每个中台 / 垂直业务独立服务
2. **数据独占**：每个服务独占自己的数据库
3. **接口契约**：gRPC + Protobuf 定义服务契约
4. **渐进式拆分**：先拆高负载模块，再拆其他
5. **双写过渡**：拆分期间新旧服务双写，数据一致后切流

#### 23.3.2 拆分顺序
```
第 1 批拆分（高负载）：
  - pay 中台 → 独立服务（金融级隔离）
  - im 中台 → 独立服务（长连接）
  - ershou 垂直业务 → 独立服务（流量大）

第 2 批拆分（中负载）：
  - user 中台 → 独立服务
  - merchant 中台 → 独立服务
  - marketing 中台 → 独立服务
  - toutiao / quan 垂直业务 → 独立服务

第 3 批拆分（全拆）：
  - 其余所有中台 + 垂直业务
  - 引入服务网格
```

#### 23.3.3 代码零修改保证
```go
// 当前：Plugin 接口调用（同进程）
type ErshouService struct {
    payService pay.PayService // 接口，不依赖实现
}
func (s *ErshouService) CreateOrder() error {
    s.payService.Create() // 同进程调用
}

// 后期：替换为 gRPC 客户端（业务代码不变）
type ErshouService struct {
    payService pay.PayService // 同样的接口
}
func (s *ErshouService) CreateOrder() error {
    s.payService.Create() // 远程调用，业务代码无感知
}

// 实现切换（依赖注入）
func InitErshouService() *ErshouService {
    var payService pay.PayService
    if config.IsMonolith {
        payService = pay.NewLocalService() // 本地实现
    } else {
        payService = pay.NewGRPCClient() // 远程实现
    }
    return &ErshouService{payService: payService}
}
```

### 23.4 服务网格（Istio）预留

#### 23.4.1 Istio 能力
- 流量管理（灰度发布 / 流量分割 / 熔断）
- 安全（mTLS / 授权策略）
- 可观测性（链路追踪 / 指标 / 日志）

#### 23.4.2 预留策略
```
当前：不引入 Istio，用 Gin 中间件实现熔断/限流
后期：K8s 部署后引入 Istio，替换部分中间件能力

预留点：
1. 所有服务间调用走 HTTP/gRPC（可被 Istio 拦截）
2. 链路追踪用 OpenTelemetry（Istio 兼容）
3. 熔断限流用接口抽象（可切换为 Istio 规则）
```

---

## 二十四、补强细节（v3.2 新增 25 项）

> **目标**：补齐 v3.1 深度审计发现的 25 项技术细节，达到金融级大厂标准的最后一块拼图。
> **范围**：8 项高优先级（金融安全）+ 12 项中优先级（大厂标准）+ 5 项低优先级（运维完善）
> **新增表**：6 张（user_2fa_codes / user_device_fingerprints / ip_profiles / behavior_analyses / audit_logs / merchant_invoices）

### 24.1 高优先级补强（8 项，金融安全核心）

#### 24.1.1 分布式锁 RedLock 完整算法

**问题**：v3.1 第 18.1.1 节标注"RedLock 算法简化版"，实际是单 Redis 实例 SetNX，存在单点故障风险。

**完整方案**：
```go
// 真正的 RedLock 算法（5 实例 quorum）
type RedLock struct {
    clients []*redis.Client // 5 个独立 Redis 实例
    key     string
    value   string // UUID
    ttl     time.Duration
}

func (rl *RedLock) Acquire() (bool, error) {
    rl.value = uuid.New().String()
    start := time.Now()
    successCount := 0

    for _, client := range rl.clients {
        result, err := client.SetNX(ctx, rl.key, rl.value, rl.ttl).Result()
        if err == nil && result {
            successCount++
        }
    }

    // 计算耗时（必须 < ttl/2，否则锁已无意义）
    elapsed := time.Since(start)
    if successCount >= 3 && elapsed < rl.ttl/2 {
        // 续期（防止获取锁的耗时吃掉 TTL）
        rl.ttl -= elapsed
        return true, nil
    }

    // 未达到 quorum，释放已获取的锁
    rl.Release()
    return false, nil
}

func (rl *RedLock) Release() {
    script := `
    if redis.call("GET", KEYS[1]) == ARGV[1] then
        return redis.call("DEL", KEYS[1])
    else
        return 0
    end
    `
    for _, client := range rl.clients {
        client.Eval(ctx, script, []string{rl.key}, rl.value)
    }
}
```

**决策**：
| 阶段 | 方案 | 理由 |
|------|------|------|
| MVP | 单 Redis 实例锁 + 主从切换 | 初期流量小，单 Redis 足够；主从切换保证可用性 |
| 中期 | 双 Redis 实例锁（2/2 quorum） | 流量上来后提升安全性 |
| 远期 | RedLock 5 实例 quorum | 金融级高可用，跨机房部署 |

---

#### 24.1.2 退款资金来源明确

**问题**：v3.1 第 9 章退款流程未说明退款资金从哪里来。

**三种方案对比**：
| 方案 | 资金来源 | 优点 | 缺点 |
|------|---------|------|------|
| A. 平台垫付 | 平台自有资金 | 用户体验好 | 平台资金压力，坏账风险 |
| B. 卖家余额扣减 | 卖家钱包余额 | 资金安全 | 卖家可能余额不足 |
| C. 冻结资金释放 | 担保交易期间冻结的资金 | 资金闭环，无风险 | 仅适用担保交易场景 |

**决策**：组合方案（按订单状态自动选择）
```go
func ProcessRefund(order *Order) error {
    switch order.Status {
    case StatusPaid: // 已支付待发货
        // 资金在 frozen 池，直接释放给买家
        wallet.Unfreeze(order.BuyerID, order.Amount, "refund_paid")
        wallet.Refund(order.BuyerID, order.Amount)

    case StatusShipped: // 已发货，买家退货
        // 资金在 frozen_after_sale 池
        // 1. 优先从卖家余额扣减（卖家已发货，承担运费风险）
        if wallet.HasBalance(order.SellerID, order.Amount) {
            wallet.Deduct(order.SellerID, order.Amount, "refund_shipped")
            wallet.Refund(order.BuyerID, order.Amount)
        } else {
            // 2. 卖家余额不足，从冻结资金释放
            wallet.UnfreezeAfterSale(order.BuyerID, order.Amount, "refund_insufficient")
            wallet.Refund(order.BuyerID, order.Amount)
            // 3. 卖家欠款记录（后续追讨）
            wallet.RecordDebt(order.SellerID, order.Amount, "refund_debt")
        }

    case StatusReceived: // 已收货待评价
        // 资金即将结算给卖家，从卖家余额扣减
        if wallet.HasBalance(order.SellerID, order.Amount) {
            wallet.Deduct(order.SellerID, order.Amount, "refund_received")
            wallet.Refund(order.BuyerID, order.Amount)
        } else {
            // 卖家余额不足，平台垫付
            wallet.PlatformAdvance(order.BuyerID, order.Amount, "refund_platform_advance")
            wallet.RecordDebt(order.SellerID, order.Amount, "platform_advance_debt")
        }

    case StatusDispute: // 纠纷中，平台仲裁
        // 按仲裁结果分配
        applyArbitrationResult(order)
    }
    return nil
}
```

---

#### 24.1.3 二次验证（2FA）

**问题**：v3.1 第 17 章安全防护未含 2FA，金融场景缺失二次验证。

**适用场景**：
| 场景 | 阈值 | 验证方式 |
|------|------|---------|
| 提现 | 所有金额 | 短信验证码 |
| 大额支付 | > 1000 元 | 短信验证码 + 支付密码 |
| 超大额支付 | > 10000 元 | 短信验证码 + 人脸识别 |
| 修改密码 | - | 短信验证码 |
| 修改手机号 | - | 短信验证码 + 人脸识别 |
| 商家入驻 | - | 短信验证码 + 营业执照 OCR |
| 管理员操作 | - | 短信验证码 + TOTP（Google Authenticator） |

**表设计**：
```sql
CREATE TABLE user_2fa_codes (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  scene VARCHAR(30) NOT NULL,              -- withdraw/pay_large/change_pwd/change_phone/merchant_settle/admin
  code VARCHAR(10) NOT NULL,               -- 6 位验证码
  channel VARCHAR(20) NOT NULL,            -- sms/face/totp
  target VARCHAR(50) DEFAULT '',           -- 手机号 / 设备 ID
  status INT DEFAULT 0,                    -- 0未验证 1已验证 2已过期 3已失败
  expire_at TIMESTAMP NOT NULL,            -- 5 分钟过期
  verify_count INT DEFAULT 0,              -- 验证次数（防爆破，最多 5 次）
  verified_at TIMESTAMP,
  created_at TIMESTAMP
);
CREATE INDEX idx_2fa_user_scene ON user_2fa_codes(user_id, scene, status);
CREATE INDEX idx_2fa_expire ON user_2fa_codes(expire_at, status);
```

**实现**：
```go
// 1. 发送验证码
func Send2FACode(userID int64, scene string) error {
    code := randomCode(6)
    user := getUser(userID)
    // 存储
    db.Create(&User2FACode{
        UserID: userID, Scene: scene, Code: code,
        Channel: "sms", Target: user.Phone,
        ExpireAt: time.Now().Add(5 * time.Minute),
    })
    // 发送短信
    sms.Send(user.Phone, fmt.Sprintf("您的验证码：%s，5分钟内有效", code))
    return nil
}

// 2. 验证
func Verify2FA(userID int64, scene, code string) error {
    record := &User2FACode{}
    if err := db.Where("user_id = ? AND scene = ? AND status = 0 AND expire_at > ?",
        userID, scene, time.Now()).First(record).Error; err != nil {
        return errors.New("验证码已过期，请重新获取")
    }
    // 防爆破
    if record.VerifyCount >= 5 {
        db.Model(record).Update("status", 2)
        return errors.New("验证次数过多，请重新获取")
    }
    db.Model(record).UpdateColumn("verify_count", gorm.Expr("verify_count + 1"))

    if record.Code != code {
        return errors.New("验证码错误")
    }
    db.Model(record).Updates(map[string]interface{}{"status": 1, "verified_at": time.Now()})
    return nil
}
```

---

#### 24.1.4 分库分表后分布式事务

**问题**：v3.1 第 22.2 节分库分表，未说明跨库事务如何处理。

**方案选型**：
| 方案 | 适用场景 | 优点 | 缺点 |
|------|---------|------|------|
| 本地消息表 | MVP | 实现简单，与现有架构一致 | 仅支持最终一致性 |
| Saga | 中期 | 支持长事务，每步可回滚 | 实现复杂，需要补偿操作 |
| TCC | 金融级 | 强一致性，性能好 | 业务侵入大，每个操作需写 Try/Confirm/Cancel |
| DTM | 后期 | Go 原生分布式事务框架 | 引入新组件 |

**决策**：分阶段引入
```go
// MVP：本地消息表（已有）
// 适用：订单创建 + 钱包扣款 + 库存扣减
func CreateOrder(req *CreateOrderReq) error {
    return db.Transaction(func(tx *gorm.DB) error {
        // 1. 创建订单（主库）
        tx.Create(&Order{...})
        // 2. 写本地消息表（主库，同一事务）
        tx.Create(&MessageQueue{
            MessageType: "wallet.freeze",
            Payload:     json.Marshal(freezeReq),
        })
        tx.Create(&MessageQueue{
            MessageType: "stock.deduct",
            Payload:     json.Marshal(stockReq),
        })
        return nil
    })
    // 定时任务扫描消息表，跨库调用 wallet / stock 服务
}

// 中期：Saga 模式（补偿事务）
type SagaStep struct {
    Action      func() error
    Compensation func() error
}

func ExecuteSaga(steps []SagaStep) error {
    completed := []int{}
    for i, step := range steps {
        if err := step.Action(); err != nil {
            // 回滚已完成的步骤
            for j := len(completed) - 1; j >= 0; j-- {
                steps[completed[j]].Compensation()
            }
            return err
        }
        completed = append(completed, i)
    }
    return nil
}

// 远期：DTM 框架
// https://github.com/dtm-labs/dtm
// 支持 SAGA / TCC / 两阶段提交 / XA
```

---

#### 24.1.5 灾备 DNS 切换

**问题**：v3.1 第 22.3 节灾备架构完整，但 DNS 切换方案未提。

**方案**：GSLB（Global Server Load Balancing）+ 智能 DNS

**实现**：
```
【正常流量】
  用户 → 智能 DNS（阿里云 GTM / 腾讯云 DNSPod / Cloudflare）
    → 解析到北京机房 A（主）
      → Nginx LB → 应用集群

【故障切换】
  1. 健康检查（每 5 秒）
     - HTTP 健康检查：GET /health
     - TCP 健康检查：3306/6379/8088
     - 自定义脚本：检查 PG 主从状态

  2. 故障检测（10 秒内）
     - 连续 3 次健康检查失败
     - 触发 DNS 切换

  3. DNS 切换（30 秒内）
     - 智能 DNS 将 A 记录从北京 A → 北京 B
     - TTL 设置为 30 秒（平衡 DNS 缓存与切换速度）
     - 通知运维团队（短信 + IM）

  4. 流量引流（1 分钟内）
     - 北京 B 机房应用接管流量
     - PG 从库提升为主（Patroni + etcd）
     - Redis 从提升为主

  5. 故障恢复（人工确认）
     - 北京 A 修复后，作为从库加入集群
     - 数据同步完成后，可选择切回（人工操作）
```

**配置示例**（阿里云 GTM）：
```yaml
# 主地址池
primary_pool:
  - 1.1.1.1  # 北京 A 机房
  - 1.1.1.2
  health_check:
    type: http
    path: /health
    interval: 5s
    timeout: 3s
    healthy_threshold: 3
    unhealthy_threshold: 3

# 备用地址池
backup_pool:
  - 2.2.2.2  # 北京 B 机房
  - 2.2.2.3

# 切换策略
failover_policy:
  type: master_slave
  ttl: 30s
  notification:
    - sms: 13800138000
    - webhook: https://im-service/api/notify
```

---

#### 24.1.6 微服务双写一致性校验

**问题**：v3.1 第 23.3.3 节"双写过渡"，未说明一致性校验方案。

**方案**：定时对账 + 不一致告警 + 自动修复

```go
// 双写期间数据对账任务（每 5 分钟执行）
type DataReconcileJob struct {
    oldDB *gorm.DB // 旧服务数据库
    newDB *gorm.DB // 新服务数据库
}

func (j *DataReconcileJob) Run() {
    // 1. 对账订单数据
    j.reconcileOrders()

    // 2. 对账钱包数据
    j.reconcileWallets()

    // 3. 对账用户数据
    j.reconcileUsers()
}

func (j *DataReconcileJob) reconcileOrders() {
    // 获取最近 10 分钟的订单（双写期间）
    since := time.Now().Add(-10 * time.Minute)

    var oldOrders []Order
    j.oldDB.Where("updated_at > ?", since).Find(&oldOrders)

    var newOrders []Order
    j.newDB.Where("updated_at > ?", since).Find(&newOrders)

    // 对比
    oldMap := make(map[int64]*Order)
    for i := range oldOrders {
        oldMap[oldOrders[i].ID] = &oldOrders[i]
    }

    for _, newOrder := range newOrders {
        oldOrder, exists := oldMap[newOrder.ID]
        if !exists {
            // 新库有，旧库无 → 告警
            alertService.Send("data_inconsistent",
                fmt.Sprintf("订单 %d 新库有旧库无", newOrder.ID))
            continue
        }

        if !orderEqual(oldOrder, &newOrder) {
            // 数据不一致 → 告警 + 自动修复（以新库为准）
            alertService.Send("data_inconsistent",
                fmt.Sprintf("订单 %d 数据不一致，自动修复", newOrder.ID))
            j.oldDB.Model(&Order{}).Where("id = ?", newOrder.ID).Updates(newOrder)
        }
    }

    // 差异率告警
    diffRate := float64(len(inconsistentList)) / float64(len(newOrders))
    if diffRate > 0.0001 { // 0.01%
        alertService.Send("diff_rate_high",
            fmt.Sprintf("数据差异率 %.4f%% > 0.01%%", diffRate*100))
    }
}
```

---

#### 24.1.7 JWT 刷新机制

**问题**：v3.1 第 17.6 节提到 JWT，未说明刷新机制。

**方案**：Access Token + Refresh Token + 黑名单

```go
const (
    AccessTokenTTL  = 30 * time.Minute  // Access Token 30 分钟过期
    RefreshTokenTTL = 7 * 24 * time.Hour // Refresh Token 7 天过期
)

type TokenPair struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int64  `json:"expires_in"` // Access Token 过期时间戳
}

// 1. 登录时颁发 Token 对
func Login(userID int64) (*TokenPair, error) {
    accessToken, _ := generateToken(userID, "access", AccessTokenTTL)
    refreshToken, _ := generateToken(userID, "refresh", RefreshTokenTTL)

    // Refresh Token 存 Redis（用于校验有效性 + 滑动续期）
    redis.Set(ctx, fmt.Sprintf("refresh:%d", userID), refreshToken, RefreshTokenTTL)

    return &TokenPair{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresIn:    time.Now().Add(AccessTokenTTL).Unix(),
    }, nil
}

// 2. 刷新 Access Token
func RefreshToken(refreshToken string) (*TokenPair, error) {
    claims, err := parseToken(refreshToken)
    if err != nil {
        return nil, errors.New("refresh token invalid")
    }

    // 校验 Refresh Token 是否在 Redis 中（防止被盗用后无法吊销）
    stored, _ := redis.Get(ctx, fmt.Sprintf("refresh:%d", claims.UserID)).Result()
    if stored != refreshToken {
        return nil, errors.New("refresh token revoked")
    }

    // 检查是否在黑名单中（用户登出 / 修改密码时加入）
    if isBlacklisted(refreshToken) {
        return nil, errors.New("token blacklisted")
    }

    // 颁发新的 Token 对（滑动过期）
    return Login(claims.UserID)
}

// 3. 登出时加入黑名单
func Logout(userID int64, accessToken, refreshToken string) error {
    // Access Token 加入黑名单（剩余 TTL）
    claims, _ := parseToken(accessToken)
    ttl := time.Until(claims.ExpiresAt.Time)
    redis.Set(ctx, "blacklist:"+accessToken, 1, ttl)

    // 删除 Refresh Token
    redis.Del(ctx, fmt.Sprintf("refresh:%d", userID))

    return nil
}

// 4. 修改密码时吊销所有 Token
func ChangePassword(userID int64) error {
    // 删除 Refresh Token
    redis.Del(ctx, fmt.Sprintf("refresh:%d", userID))
    // 标记用户所有 Access Token 失效（通过 user_token_version）
    redis.Incr(ctx, fmt.Sprintf("token_version:%d", userID))
    return nil
}

// 5. 中间件校验
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractToken(c)
        // 1. 校验签名 + 过期时间
        claims, err := parseToken(token)
        if err != nil {
            c.AbortWithStatusJSON(401, gin.H{"msg": "token invalid"})
            return
        }
        // 2. 检查黑名单
        if isBlacklisted(token) {
            c.AbortWithStatusJSON(401, gin.H{"msg": "token revoked"})
            return
        }
        // 3. 检查 token_version（修改密码后旧 Token 失效）
        currentVersion, _ := redis.Get(ctx, fmt.Sprintf("token_version:%d", claims.UserID)).Int64()
        if claims.Version < currentVersion {
            c.AbortWithStatusJSON(401, gin.H{"msg": "token outdated"})
            return
        }
        c.Set("user_id", claims.UserID)
        c.Next()
    }
}
```

---

#### 24.1.8 行为分析防刷

**问题**：v3.1 第 17 章防刷主要是限流，缺少行为分析。

**方案**：设备指纹 + IP 画像 + 行为分析

**表设计**：
```sql
-- 设备指纹表
CREATE TABLE user_device_fingerprints (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  device_id VARCHAR(100) NOT NULL,          -- fingerprintjs 生成
  device_info JSON,                         -- {os, browser, screen, ...}
  first_seen TIMESTAMP,
  last_seen TIMESTAMP,
  risk_score INT DEFAULT 0,                 -- 风险评分 0-100
  status INT DEFAULT 0,                     -- 0正常 1可疑 2封禁
  UNIQUE(user_id, device_id)
);
CREATE INDEX idx_device_fp ON user_device_fingerprints(device_id, status);

-- IP 画像表
CREATE TABLE ip_profiles (
  id BIGSERIAL PRIMARY KEY,
  ip VARCHAR(50) UNIQUE NOT NULL,
  ip_type VARCHAR(20),                      -- residential/datacenter/mobile/vpn
  country VARCHAR(50),
  province VARCHAR(50),
  city VARCHAR(50),
  isp VARCHAR(100),
  risk_score INT DEFAULT 0,                 -- 风险评分 0-100
  request_count INT DEFAULT 0,              -- 总请求数
  violate_count INT DEFAULT 0,              -- 违规次数
  first_seen TIMESTAMP,
  last_seen TIMESTAMP,
  status INT DEFAULT 0                      -- 0正常 1可疑 2封禁
);
CREATE INDEX idx_ip_profile ON ip_profiles(ip, status);

-- 行为分析表
CREATE TABLE behavior_analyses (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT DEFAULT 0,
  device_id VARCHAR(100),
  ip VARCHAR(50),
  action VARCHAR(50),                       -- login/publish/pay/withdraw
  biz_module VARCHAR(50),
  risk_score INT DEFAULT 0,
  risk_reasons JSON,                        -- ["高频登录","异常IP","新设备"]
  created_at TIMESTAMP
);
CREATE INDEX idx_behavior_risk ON behavior_analyses(risk_score, created_at);
```

**实现**：
```go
type RiskEngine struct {
    redis *redis.Client
    db    *gorm.DB
}

// 风险评估
func (e *RiskEngine) Evaluate(ctx context.Context, req *RiskRequest) (*RiskResult, error) {
    score := 0
    var reasons []string

    // 1. IP 风险评估
    ipProfile := e.getIPProfile(req.IP)
    if ipProfile.IPType == "datacenter" {
        score += 30
        reasons = append(reasons, "机房IP")
    }
    if ipProfile.IPType == "vpn" {
        score += 40
        reasons = append(reasons, "VPN代理")
    }
    if ipProfile.ViolateCount > 5 {
        score += 20
        reasons = append(reasons, "IP违规历史")
    }

    // 2. 设备风险
    device := e.getDeviceFingerprint(req.UserID, req.DeviceID)
    if device.RiskScore > 50 {
        score += 30
        reasons = append(reasons, "设备风险")
    }
    if device.FirstSeen.After(time.Now().Add(-24 * time.Hour)) {
        score += 20
        reasons = append(reasons, "新设备")
    }

    // 3. 行为频率
    loginCount := e.countAction(req.UserID, "login", 1*time.Hour)
    if loginCount > 10 {
        score += 30
        reasons = append(reasons, "高频登录")
    }
    publishCount := e.countAction(req.UserID, "publish", 1*time.Hour)
    if publishCount > 20 {
        score += 40
        reasons = append(reasons, "高频发布")
    }

    // 4. 异常时间
    if isLateNight(time.Now()) {
        score += 10
        reasons = append(reasons, "凌晨操作")
    }

    // 5. 机器人特征
    if hasBotUA(req.UserAgent) {
        score += 50
        reasons = append(reasons, "爬虫UA")
    }

    return &RiskResult{
        Score:   score,
        Reasons: reasons,
        Action:  e.decideAction(score), // pass / verify / block
    }, nil
}

func (e *RiskEngine) decideAction(score int) string {
    if score >= 80 {
        return "block" // 直接拦截
    }
    if score >= 50 {
        return "verify" // 需要二次验证
    }
    return "pass" // 放行
}
```

---

### 24.2 中优先级补强（12 项，大厂标准完整性）

#### 24.2.1 WebSocket 集群方案

**问题**：v3.1 第 4.3 节 IM 消息中台"MVP 单实例，接口预留集群扩展"，未说明集群方案。

**三方案对比**：
| 方案 | 实现 | 优点 | 缺点 |
|------|------|------|------|
| A. Redis Pub/Sub 广播 | 消息发到 Redis 频道，所有实例订阅 | 简单，延迟低（< 10ms） | 消息冗余（所有实例都收到） |
| B. Sticky Session | Nginx ip_hash，同一用户固定到同一实例 | 简单，无状态 | 故障转移差，负载不均 |
| C. RabbitMQ 主题转发 | 消息发到 MQ，按 userID 路由 | 可靠，精确投递 | 延迟稍高（20-50ms） |

**决策**：方案 A（Redis Pub/Sub 广播）
```go
type ClusterHub struct {
    localHub    *Hub               // 本地连接
    redis       *redis.Client
    instanceID  string             // 当前实例 ID
}

func (h *ClusterHub) Start() {
    // 1. 订阅 Redis 频道
    pubsub := h.redis.Subscribe(ctx, "im:broadcast")
    go func() {
        for {
            msg, _ := pubsub.ReceiveMessage(ctx)
            var payload MessagePayload
            json.Unmarshal([]byte(msg.Payload), &payload)

            // 2. 如果目标用户在当前实例，推送
            if h.localHub.HasUser(payload.TargetUserID) {
                h.localHub.SendToUser(payload.TargetUserID, payload.Message)
            }
        }
    }()
}

func (h *ClusterHub) SendToUser(userID int64, msg *Message) error {
    // 1. 先查本地
    if h.localHub.HasUser(userID) {
        return h.localHub.SendToUser(userID, msg)
    }

    // 2. 本地无，广播到其他实例
    payload, _ := json.Marshal(MessagePayload{
        TargetUserID: userID,
        Message:      msg,
        FromInstance: h.instanceID,
    })
    h.redis.Publish(ctx, "im:broadcast", payload)
    return nil
}
```

---

#### 24.2.2 布隆过滤器持久化

**问题**：v3.1 第 18.2.1 节布隆过滤器启动时从 DB 加载，重启期间新增数据会丢失。

**方案**：Redis Bitmap 持久化
```go
type RedisBloomFilter struct {
    redis  *redis.Client
    key    string // "bloom:ershou_id"
    bits   uint   // 位数
    hashes uint   // 哈希函数数量
}

// 启动时从 Redis 加载（毫秒级）
func (bf *RedisBloomFilter) Init() error {
    // Redis Bitmap 已持久化，无需从 DB 重新加载
    // 仅校验是否存在
    exists, _ := bf.redis.Exists(ctx, bf.key).Result()
    if !exists {
        // 首次启动，从 DB 初始化
        return bf.rebuildFromDB()
    }
    return nil
}

// 新增数据时同步写入 Redis Bitmap
func (bf *RedisBloomFilter) Add(id int64) error {
    for i := uint(0); i < bf.hashes; i++ {
        pos := bf.hash(id, i) % bf.bits
        bf.redis.SetBit(ctx, bf.key, int64(pos), 1)
    }
    return nil
}

// 查询
func (bf *RedisBloomFilter) Test(id int64) bool {
    for i := uint(0); i < bf.hashes; i++ {
        pos := bf.hash(id, i) % bf.bits
        bit, _ := bf.redis.GetBit(ctx, bf.key, int64(pos)).Result()
        if bit == 0 {
            return false
        }
    }
    return true
}

// 定期重建（每日凌晨，防止误判率累积）
func (bf *RedisBloomFilter) rebuildFromDB() error {
    var ids []int64
    db.Model(&Ershou{}).Pluck("id", &ids)

    // 双布隆过滤器策略：先写新 key，再切换
    newKey := bf.key + ":new"
    for _, id := range ids {
        for i := uint(0); i < bf.hashes; i++ {
            pos := bf.hash(id, i) % bf.bits
            bf.redis.SetBit(ctx, newKey, int64(pos), 1)
        }
    }
    // 原子切换
    bf.redis.Rename(ctx, newKey, bf.key)
    return nil
}
```

---

#### 24.2.3 消息幂等表清理策略

**问题**：v3.1 第 19.2 节 message_idempotent 表会持续增长。

**方案**：分区表 + 定期清理
```sql
-- 按 consumed_at 月度分区
CREATE TABLE message_idempotent (
  id BIGSERIAL,
  consumer_group VARCHAR(50) NOT NULL,
  message_id VARCHAR(64) NOT NULL,
  status VARCHAR(20) NOT NULL,
  result TEXT,
  retry_count INT DEFAULT 0,
  consumed_at TIMESTAMP,
  created_at TIMESTAMP
) PARTITION BY RANGE (consumed_at);

-- 创建月度分区
CREATE TABLE message_idempotent_202607 PARTITION OF message_idempotent
  FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');
CREATE TABLE message_idempotent_202608 PARTITION OF message_idempotent
  FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');

-- 唯一约束（每个分区独立）
CREATE UNIQUE INDEX idx_idem_202607 ON message_idempotent_202607(consumer_group, message_id);

-- 定时清理（每月 1 号执行，DROP 30 天前的分区）
-- cron: 0 0 1 * *
DROP TABLE IF EXISTS message_idempotent_202606;

-- 归档（超过 30 天的死信消息转 OSS）
-- 保留 dead 状态消息 90 天，用于故障排查
```

---

#### 24.2.4 配置中心高可用降级

**问题**：v3.1 第 20.1.4 节配置热更新依赖 Redis Pub/Sub，Redis 故障时配置无法更新。

**方案**：三级降级
```go
type ConfigService struct {
    redis     *redis.Client
    db        *gorm.DB
    localCache *ristretto.Cache  // 进程内缓存
    fallback  map[string]string  // 默认配置（编译时内置）
}

func (s *ConfigService) Get(key string) string {
    // 1. 查本地缓存（1ms）
    if val, ok := s.localCache.Get(key); ok {
        return val.(string)
    }

    // 2. 查 Redis（5ms）
    if val, err := s.redis.Get(ctx, "config:"+key).Result(); err == nil {
        s.localCache.Set(key, val, 30*time.Second)
        return val
    }

    // 3. Redis 故障，查 DB（50ms）
    var item ConfigItem
    if err := s.db.Where("config_key = ?", key).First(&item).Error; err == nil {
        s.localCache.Set(key, item.ConfigValue, 30*time.Second)
        return item.ConfigValue
    }

    // 4. DB 故障，返回默认配置
    if val, ok := s.fallback[key]; ok {
        return val
    }

    return ""
}

// 定时拉取（30 秒，作为 Pub/Sub 的降级方案）
func (s *ConfigService) StartPuller() {
    ticker := time.NewTicker(30 * time.Second)
    go func() {
        for range ticker.C {
            s.pullConfigs()
        }
    }()
}

// Redis Pub/Sub 失效后的自动降级
func (s *ConfigService) Subscribe() {
    for {
        pubsub := s.redis.Subscribe(ctx, "config:change")
        _, err := pubsub.Receive(ctx)
        if err != nil {
            // Redis 故障，切换到定时拉取模式
            log.Warn("Redis Pub/Sub failed, switch to polling mode")
            time.Sleep(30 * time.Second)
            continue
        }
        // 正常模式
        s.handlePubSub(pubsub)
    }
}
```

---

#### 24.2.5 日志脱敏方案

**问题**：v3.1 第 16.2 节日志规范很详细，但敏感字段如何脱敏未说明。

**方案**：zerolog Hook 自动脱敏
```go
type SensitiveDataHook struct{}

var sensitiveKeys = map[string]string{
    "phone":          "phone",
    "mobile":         "phone",
    "id_card":        "idcard",
    "idcard":         "idcard",
    "bank_card":      "bankcard",
    "bankcard":       "bankcard",
    "password":       "password",
    "pwd":            "password",
    "token":          "token",
    "secret":         "secret",
    "api_key":        "apikey",
    "apikey":         "apikey",
}

func (h *SensitiveDataHook) Run(e *zerolog.Event, level zerolog.Level, message string) {
    // zerolog 的 Hook 在日志输出前调用
    // 这里通过自定义 Formatter 实现字段脱敏
}

// 脱敏函数
func MaskSensitive(key, value string) string {
    switch sensitiveKeys[strings.ToLower(key)] {
    case "phone":
        // 13800138000 → 138****8000
        if len(value) == 11 {
            return value[:3] + "****" + value[7:]
        }
    case "idcard":
        // 110101199001010001 → 110101********0001
        if len(value) >= 15 {
            return value[:6] + "********" + value[len(value)-4:]
        }
    case "bankcard":
        // 6222020200012345678 → 6222****5678
        if len(value) >= 8 {
            return value[:4] + "****" + value[len(value)-4:]
        }
    case "password", "token", "secret", "apikey":
        return "******" // 完全隐藏
    }
    return value
}

// 使用：业务代码无需关心脱敏
log.Info().
    Str("phone", user.Phone).     // 自动脱敏为 138****8000
    Str("id_card", user.IDCard).  // 自动脱敏为 110101********0001
    Str("action", "login").
    Msg("用户登录")
```

---

#### 24.2.6 统一审计日志表

**问题**：v3.1 多处提到审计日志，但没有统一的表设计。

**方案**：统一 audit_logs 表
```sql
CREATE TABLE audit_logs (
  id BIGSERIAL PRIMARY KEY,
  operator_id BIGINT NOT NULL,              -- 操作人 ID
  operator_name VARCHAR(50) NOT NULL,       -- 操作人姓名
  operator_role VARCHAR(30),                -- 操作人角色
  operation VARCHAR(30) NOT NULL,           -- create/update/delete/login/logout/export
  target_type VARCHAR(50) NOT NULL,         -- user/order/wallet/config/permission/refund
  target_id VARCHAR(50),                    -- 目标对象 ID
  target_name VARCHAR(200),                 -- 目标对象名称（便于查询）
  old_value JSON,                           -- 修改前的值
  new_value JSON,                           -- 修改后的值
  ip VARCHAR(50),                           -- 操作 IP
  user_agent VARCHAR(500),                  -- UA
  region_id BIGINT,                         -- 分站
  module VARCHAR(50),                       -- 所属模块
  result VARCHAR(20),                       -- success/failed
  error_msg VARCHAR(500),
  created_at TIMESTAMP
);
CREATE INDEX idx_audit_operator ON audit_logs(operator_id, created_at);
CREATE INDEX idx_audit_target ON audit_logs(target_type, target_id, created_at);
CREATE INDEX idx_audit_operation ON audit_logs(operation, created_at);
CREATE INDEX idx_audit_module ON audit_logs(module, created_at);
```

**适用场景**：
- 用户管理：创建/修改/删除用户
- 订单管理：退款/仲裁/状态变更
- 钱包管理：充值/提现/冻结/解冻
- 配置管理：配置项变更
- 权限管理：角色/权限分配
- 商家管理：入驻/认领/停用

**保留期**：1 年（合规要求），按月度分区。

---

#### 24.2.7 文件上传安全防护

**问题**：v3.1 素材存储中台负责文件上传，但安全防护未说明。

**方案**：5 层防护
```go
type UploadSecurityChecker struct{}

// 1. 文件类型白名单
var allowedMimeTypes = map[string]bool{
    "image/jpeg": true, "image/png": true, "image/gif": true, "image/webp": true,
    "video/mp4": true, "video/webm": true,
    "application/pdf": true,
}

// 2. 文件大小限制
var sizeLimits = map[string]int64{
    "image": 10 * 1024 * 1024,   // 10MB
    "video": 100 * 1024 * 1024,  // 100MB
    "doc":   20 * 1024 * 1024,   // 20MB
}

func (c *UploadSecurityChecker) Check(file *multipart.FileHeader) error {
    // 1. 文件大小
    if file.Size > 100*1024*1024 {
        return errors.New("文件超过 100MB 限制")
    }

    // 2. 扩展名校验
    ext := strings.ToLower(filepath.Ext(file.Filename))
    if !isAllowedExt(ext) {
        return errors.New("文件类型不允许")
    }

    // 3. MIME 类型校验（防止伪造扩展名）
    mime := detectMimeType(file)
    if !allowedMimeTypes[mime] {
        return errors.New("文件 MIME 类型不允许")
    }

    // 4. 图片防马（检查图片二次渲染）
    if strings.HasPrefix(mime, "image/") {
        if isMaliciousImage(file) {
            return errors.New("图片包含恶意代码")
        }
    }

    // 5. 病毒扫描（可选，ClamAV）
    if config.EnableVirusScan {
        if isInfected(file) {
            return errors.New("文件包含病毒")
        }
    }

    return nil
}

// 图片防马：尝试用 image.Decode 解码，失败说明图片有问题
func isMaliciousImage(file *multipart.FileHeader) bool {
    f, _ := file.Open()
    defer f.Close()
    _, err := image.DecodeConfig(f)
    return err != nil
}
```

---

#### 24.2.8 第三方 API 容错

**问题**：v3.1 多处调用第三方 API（微信支付/高德/AI），容错方案未说明。

**方案**：超时 + 重试 + 熔断 + 降级
```go
type ThirdPartyClient struct {
    httpClient   *http.Client
    circuitBreaker *CircuitBreaker
    cache        Cache
}

func NewThirdPartyClient() *ThirdPartyClient {
    return &ThirdPartyClient{
        httpClient: &http.Client{
            Timeout: 10 * time.Second, // 总超时
            Transport: &http.Transport{
                DialContext: (&net.Dialer{
                    Timeout:   3 * time.Second,  // 连接超时
                    KeepAlive: 30 * time.Second,
                }).DialContext,
                ResponseHeaderTimeout: 5 * time.Second, // 读响应头超时
            },
        },
        circuitBreaker: NewCircuitBreaker(),
    }
}

func (c *ThirdPartyClient) Call(req *http.Request) (*http.Response, error) {
    // 1. 熔断检查
    if !c.circuitBreaker.Allow() {
        return nil, errors.New("circuit breaker open")
    }

    // 2. 重试（3 次，指数退避）
    var lastErr error
    for i := 0; i < 3; i++ {
        resp, err := c.httpClient.Do(req)
        if err == nil && resp.StatusCode < 500 {
            c.circuitBreaker.RecordSuccess()
            return resp, nil
        }
        lastErr = err
        time.Sleep(time.Duration(math.Pow(2, float64(i))) * time.Second) // 1s, 2s, 4s
    }

    // 3. 熔断记录失败
    c.circuitBreaker.RecordFailure()

    // 4. 降级：返回缓存数据
    if cached, ok := c.cache.Get(req.URL.String()); ok {
        log.Warn("third party failed, return cached", "url", req.URL.String())
        return &http.Response{
            StatusCode: 200,
            Body:       ioutil.NopCloser(bytes.NewReader(cached)),
        }, nil
    }

    return nil, fmt.Errorf("third party call failed after 3 retries: %w", lastErr)
}

// 熔断器
type CircuitBreaker struct {
    failureCount int
    threshold    int // 5 次失败触发熔断
    openUntil    time.Time
    mu           sync.Mutex
}

func (cb *CircuitBreaker) Allow() bool {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    return time.Now().After(cb.openUntil)
}

func (cb *CircuitBreaker) RecordFailure() {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    cb.failureCount++
    if cb.failureCount >= cb.threshold {
        cb.openUntil = time.Now().Add(30 * time.Second) // 熔断 30 秒
    }
}
```

---

#### 24.2.9 资金对账差异处理流程

**问题**：v3.1 第 9.5 节每日对账，差异处理流程未说明。

**方案**：分级处理
```go
func ProcessReconcileDiff(diff decimal.Decimal) {
    absDiff := diff.Abs()

    switch {
    case absDiff.LessThan(decimal.NewFromFloat(0.01)):
        // 1. 差异 < 0.01 元：自动修复（浮点精度问题）
        log.Info("auto fix small diff", "diff", diff)
        autoFixDiff(diff)

    case absDiff.LessThan(decimal.NewFromInt(1)):
        // 2. 差异 0.01-1 元：记录日志，人工审核
        log.Warn("diff needs review", "diff", diff)
        createReconcileTicket(diff, "review")

    default:
        // 3. 差异 > 1 元：立即告警，财务介入
        log.Error("diff exceeds threshold", "diff", diff)
        createReconcileTicket(diff, "urgent")
        // 短信 + 电话告警
        alertService.SendUrgent(fmt.Sprintf("资金对账差异 %.2f 元，请立即核查", diff))
        // 暂停提现（防止资损扩大）
        walletService.PauseWithdraw()
    }
}

// 自动修复小差异
func autoFixDiff(diff decimal.Decimal) {
    // 记录修复日志
    db.Create(&ReconcileFixLog{
        Diff:      diff,
        Action:    "auto_fix",
        Reason:    "floating point precision",
        FixedAt:   time.Now(),
    })
    // 调整 wallet_accounts 的 total_recharge（资金平衡项）
    if diff.GreaterThan(decimal.Zero) {
        db.Exec("UPDATE wallet_accounts SET balance = balance + ? WHERE user_id = 0", diff)
    } else {
        db.Exec("UPDATE wallet_accounts SET balance = balance - ? WHERE user_id = 0", diff.Abs())
    }
}
```

---

#### 24.2.10 商家结算发票管理

**问题**：v3.1 商家结算涉及发票，但发票管理未说明。

**表设计**：
```sql
CREATE TABLE merchant_invoices (
  id BIGSERIAL PRIMARY KEY,
  shop_id BIGINT NOT NULL,
  settle_id BIGINT,                         -- 关联 merchant_settles
  invoice_type VARCHAR(20),                 -- electronic/paper
  invoice_status INT DEFAULT 0,             -- 0待开具 1已开具 2已邮寄 3已作废
  amount DECIMAL(12,2) NOT NULL,
  tax_amount DECIMAL(12,2) NOT NULL,
  invoice_title VARCHAR(200),               -- 发票抬头
  tax_no VARCHAR(50),                       -- 税号
  invoice_no VARCHAR(50),                   -- 发票号码
  invoice_url VARCHAR(500),                 -- 电子发票 PDF URL
  mail_address JSON,                        -- 邮寄地址
  mailed_at TIMESTAMP,
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);
CREATE INDEX idx_invoice_shop ON merchant_invoices(shop_id, invoice_status);
```

**开票流程**：
1. 商家申请开票 / 系统月度自动开票
2. 财务审核发票信息（抬头/税号）
3. 调用电子发票 API（如金蝶/百望）开具
4. 推送 PDF 给商家
5. 纸质发票邮寄 + 录入快递单号

---

#### 24.2.11 多端数据同步

**问题**：v3.1 UniAppX 多端，多端数据同步方案未说明。

**方案**：WebSocket 实时推送 + 增量同步
```go
// 1. 实时同步（WebSocket 推送）
// 订单状态变更、消息、通知等实时数据
func PushToAllDevices(userID int64, msg *Message) error {
    // ClusterHub 自动推送到用户所有设备
    return clusterHub.SendToUser(userID, msg)
}

// 2. 增量同步（客户端记录 last_sync_at）
// GET /api/v1/sync?last_sync_at=2026-07-19T10:00:00Z
type SyncResponse struct {
    Orders       []Order       `json:"orders"`        // 订单增量
    Messages     []Message     `json:"messages"`       // 消息增量
    Notifications []Notification `json:"notifications"` // 通知增量
    WalletTx     []WalletTx    `json:"wallet_tx"`      // 钱包流水增量
    ServerTime   time.Time     `json:"server_time"`    // 服务器当前时间
}

func Sync(c *gin.Context) {
    lastSyncAt, _ := time.Parse(time.RFC3339, c.Query("last_sync_at"))
    userID := c.GetInt64("user_id")

    resp := &SyncResponse{
        Orders:        orderRepo.FindUpdatedAfter(userID, lastSyncAt),
        Messages:      messageRepo.FindUpdatedAfter(userID, lastSyncAt),
        Notifications: notifyRepo.FindUpdatedAfter(userID, lastSyncAt),
        WalletTx:      walletRepo.FindUpdatedAfter(userID, lastSyncAt),
        ServerTime:    time.Now(),
    }
    c.JSON(200, resp)
}

// 3. 冲突解决：以服务端为准
// 客户端检测到冲突时，覆盖本地数据
// 示例：客户端修改了订单备注，但服务端订单状态已变更
// 解决：保留客户端的备注修改，但接受服务端的状态变更
```

---

#### 24.2.12 蓝绿部署/金丝雀发布

**问题**：v3.1 第 11.3 节 CI/CD，发布策略未说明。

**方案**：
```yaml
# 蓝绿部署（用于大版本发布）
blue_green:
  blue_env:
    replicas: 2
    version: v1.0.0  # 当前线上版本
  green_env:
    replicas: 2
    version: v1.1.0  # 新版本
  switch_strategy:
    - 部署 green 环境
    - 健康检查 green 环境
    - DNS 切换 blue → green（30 秒）
    - 观察 30 分钟
    - 无异常：下线 blue
    - 有异常：DNS 切回 blue（30 秒回滚）

# 金丝雀发布（用于小版本迭代）
canary:
  stages:
    - stage1:
        percentage: 5%  # 5% 流量到新版本
        duration: 1h
        success_criteria:
          error_rate: < 0.1%
          latency_p99: < 500ms
    - stage2:
        percentage: 20%
        duration: 2h
    - stage3:
        percentage: 50%
        duration: 4h
    - stage4:
        percentage: 100%  # 全量
  rollback_criteria:
    error_rate: > 1%
    latency_p99: > 1s
  traffic_split:
    method: header_based  # 按 header 分流（用户ID/城市）
    header: X-Canary
```

**实现**（Nginx 配置）：
```nginx
# 金丝雀发布：5% 流量到新版本
upstream backend_v1 {
    server 127.0.0.1:8088 weight=95;
}
upstream backend_v2 {
    server 127.0.0.1:8089 weight=5;
}

# 按 header 分流
split_clients "${remote_addr}${http_user_agent}" $backend {
    5%  backend_v2;
    *   backend_v1;
}

server {
    location / {
        proxy_pass http://$backend;
    }
}
```

---

### 24.3 低优先级补强（5 项，运维完善度）

#### 24.3.1 监控数据长期存储

**问题**：v3.1 Prometheus 默认存储 15 天，长期存储未说明。

**方案**：Thanos
```yaml
# Thanos 架构
thanos:
  sidecar:
    # 每个 Prometheus 实例部署 sidecar
    - 上传 Block 到 OSS（长期存储）
    - 暴露 StoreAPI（查询历史数据）
  query:
    # 统一查询入口
    - 聚合多个 Prometheus + Sidecar
    - 支持 PrometheusQL
  compactor:
    # 降采样
    - 5 分钟降采样（保留 1 年）
    - 1 小时降采样（保留 5 年）
  store:
    # OSS 存储
    - 原始数据保留 90 天
    - 降采样数据保留 5 年
```

---

#### 24.3.2 容器资源限制

**问题**：v3.1 Docker Compose 部署，容器资源限制未说明。

**方案**：
```yaml
# docker-compose.yml
services:
  backend:
    image: wuchang-tongcheng/backend:latest
    deploy:
      resources:
        limits:
          cpus: '4'           # 最大 4 核
          memory: 4G          # 最大 4GB
        reservations:
          cpus: '1'           # 保留 1 核
          memory: 1G          # 保留 1GB
    # OOM 处理
    oom_kill_disable: false   # 允许 OOM Killer
    oom_score_adj: 500        # OOM 优先级（越高越先被杀）

  postgres:
    deploy:
      resources:
        limits:
          cpus: '8'
          memory: 16G
        reservations:
          cpus: '2'
          memory: 4G

  redis:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 4G
        reservations:
          cpus: '1'
          memory: 1G

  # 资源监控
  cadvisor:
    image: gcr.io/cadvisor/cadvisor:latest
    ports:
      - "8080:8080"
    volumes:
      - /:/rootfs:ro
      - /var/run:/var/run:ro
      - /sys:/sys:ro
      - /var/lib/docker/:/var/lib/docker:ro
```

---

#### 24.3.3 数据库/Redis 连接池配置

**问题**：v3.1 连接池配置未说明。

**方案**：
```go
// PostgreSQL 连接池
func NewPGDB(dsn string) (*gorm.DB, error) {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        PrepareStmt: true,
    })
    if err != nil {
        return nil, err
    }
    sqlDB, _ := db.DB()
    sqlDB.SetMaxOpenConns(100)          // 最大连接数
    sqlDB.SetMaxIdleConns(10)           // 最大空闲连接
    sqlDB.SetConnMaxLifetime(30 * time.Minute)  // 连接最大存活时间
    sqlDB.SetConnMaxIdleTime(5 * time.Minute)   // 空闲连接最大存活时间
    return db, nil
}

// Redis 连接池
func NewRedisClient(addr string) *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr: addr,
        PoolSize:     50,            // 连接池大小
        MinIdleConns: 10,            // 最小空闲连接
        DialTimeout:  3 * time.Second,
        ReadTimeout:  3 * time.Second,
        WriteTimeout: 3 * time.Second,
        PoolTimeout:  4 * time.Second,
        IdleTimeout:  5 * time.Minute,
    })
}
```

---

#### 24.3.4 慢查询优化方案

**问题**：v3.1 提到慢查询告警，优化方案未说明。

**方案**：4 步优化
```go
// 1. EXPLAIN ANALYZE 分析
// 慢查询自动记录 + 定期分析
type SlowQueryAnalyzer struct{}

func (a *SlowQueryAnalyzer) Analyze(query string) {
    // 执行 EXPLAIN ANALYZE
    result := db.Raw("EXPLAIN ANALYZE " + query).Scan(&plan)
    // 检查是否 Seq Scan（全表扫描）
    if strings.Contains(plan, "Seq Scan") {
        suggestIndex(query)
    }
    // 检查是否高成本
    if extractCost(plan) > 1000 {
        alertService.Send("slow_query_cost", query)
    }
}

// 2. 索引优化建议
func suggestIndex(query string) {
    // 解析 WHERE / JOIN / ORDER BY 字段
    // 建议创建复合索引
    // 示例：WHERE region_id = ? AND status = ? → CREATE INDEX idx_region_status
}

// 3. SQL 重写
// 避免 SELECT *，改为 SELECT 具体字段
// 避免 N+1 查询，改为 Preload / Joins
// 避免大偏移量分页（OFFSET 1000000），改为游标分页

// 4. 缓存策略
// 热点数据 Redis 缓存（5 分钟）
// 列表数据分页缓存（首页缓存，其他页不缓存）
// 详情数据缓存（详情页 + 评论分开缓存）
```

---

#### 24.3.5 CDN 缓存策略

**问题**：v3.1 素材存储中台提到 CDN，缓存策略未说明。

**方案**：
```yaml
# CDN 缓存策略
cdn:
  # 缓存层级：CDN → Nginx → Redis → DB
  layers:
    - cdn:       # 阿里云 CDN / 七牛云 CDN
        ttl:
          static: 30 days       # 静态资源（图片/视频/CSS/JS）
          dynamic: 5 minutes    # 动态数据（API 响应）
        cache_key:
          - url
          - query_params
        rules:
          - pattern: "*.jpg"
            ttl: 30 days
          - pattern: "*.mp4"
            ttl: 30 days
          - pattern: "/api/v1/ershou/list"
            ttl: 1 minute        # 列表接口短缓存

    - nginx:
        proxy_cache: true
        cache_path: /var/cache/nginx levels=1:2 keys_zone=api_cache:100m
        ttl:
          static: 7 days
          dynamic: 30 seconds

    - redis:
        ttl:
          hot_data: 5 minutes    # 热点数据
          user_data: 30 seconds  # 用户数据

  # 刷新机制
  refresh:
    url_refresh:
      - 单个 URL 刷新（适用于修改单条数据）
    dir_refresh:
      - 目录刷新（适用于批量更新）
    preload:
      - 发布时主动预热（首页/Banner/热门商品）

  # 预热机制
  warmup:
    - 每日 8:00 预热首页 Banner
    - 每日 9:00 预热热门商品 Top 100
    - 发布新功能时预热新页面

  # 缓存命中率监控
  metrics:
    - hit_rate: > 95% (目标)
    - miss_rate: < 5%
    - 告警: hit_rate < 80% 持续 5 分钟
```

---

## 附录 A：表清单汇总

### A.1 P0 底层基座（6 张）
1. modules - 模块注册表
2. cron_jobs - 定时任务调度中心
3. module_grayscales - 灰度发布表
4. message_queue - 本地消息表
5. module_station_configs - 分站配置中心扩展
6. module_metrics - 监控指标表

### A.2 P1 通用中台（约 30 张）
- **user 中台（5张）**：user_accounts(已存在) / user_profiles / user_credits / user_vip_levels / user_vip_orders / user_oauths / user_realnames
- **pay 中台（7张）**：pay_orders / wallet_accounts / wallet_transactions / order_orders / refund_orders / settle_rules / settle_records
- **im 中台（3张）**：im_conversations / im_messages / im_notify_templates
- **merchant 中台（3张）**：merchant_shops / merchant_staff / merchant_settles
- **distribution 中台（3张）**：distribution_partners / distribution_channels / distribution_commissions
- **marketing 中台（6张）**：ad_positions / coupons / user_coupons / sign_records / activities
- **risk 中台（5张）**：credit_scores / credit_logs / audit_tasks / sensitive_words / reports / violations
- **lbs 中台（2张）**：lbs_pois / lbs_regions
- **ai 中台（2张）**：ai_models / ai_usage_logs
- **tenant 中台（3张）**：tenant_stations / tenant_staff / tenant_configs
- **material 中台（2张）**：material_files / material_video_tasks
- **diy 中台（2张）**：diy_pages / diy_templates

### A.3 P2 ershou 专属（13 张）
- 已存在 4 张：erhous / ershou_images / ershou_favorites / ershou_messages
- 新增 9 张：ershou_wanteds / ershou_wanted_images / ershou_footprints / ershou_attributes / ershou_attribute_values / ershou_user_behaviors / ershou_user_profiles / ershou_shop_secondhand / ershou_video_tasks / ershou_data_stats

### A.4 P1 大厂级能力补强（v3.1 新增 5 张）
- **可观测性（1张）**：trace_samples（链路追踪采样，第十六章）
- **幂等性（1张）**：message_idempotent（消息消费幂等，第十九章）
- **配置中心（3张）**：config_items（配置项） / config_grayscales（灰度配置） / config_audit_logs（配置审计，第二十章）

### A.5 v3.2 补强细节新增（6 张，第二十四章）
- **风控安全（4张）**：
  - user_2fa_codes（二次验证码表，24.1 高优先级 - 二次验证 2FA）
  - user_device_fingerprints（用户设备指纹表，24.1 高优先级 - 行为分析防刷）
  - ip_profiles（IP 画像表，24.1 高优先级 - 行为分析防刷）
  - behavior_analyses（行为分析记录表，24.1 高优先级 - 行为分析防刷）
- **审计合规（1张）**：audit_logs（统一审计日志表，24.2 中优先级 - 统一审计日志表，月度分区 + 1 年保留）
- **财务结算（1张）**：merchant_invoices（商家发票表，24.2 中优先级 - 商家结算发票管理，电子/纸质 + 金蝶/百望 API）

### A.6 总计
- P0：6 张
- P1：约 35 张（含子域表）
- P1 大厂级能力补强（v3.1）：5 张
- v3.2 补强细节新增：6 张
- P2 ershou：13 张（4 已存在 + 9 新增）
- **新增合计：约 65 张表**

---

## 附录 B：参考链接

- dismall 点微二手：https://addon.dismall.com/plugins/tom_tcershou.html
- dismall 西瓜二手：https://addon.dismall.com/plugins/xigua_es.html
- 点微科技开发者主页：https://addon.dismall.com/developer-26633.html
- 西瓜先生开发者主页：https://addon.dismall.com/developer-26633.html
- 闲鱼：https://www.goofish.com/
- 转转：https://www.zhuanzhuan.com/
- 58 同城：https://www.58.com/
- 有赞后台架构：https://www.youzan.com/
- 微盟后台架构：https://www.weimob.com/
- Shopify 后台架构：https://www.shopify.com/

---

## 附录 C：v2 → v3 变更说明

### C.1 战略范围升级
- v2：仅聚焦 ershou 单模块大厂级方案
- v3：扩展为全平台 12 中台 + 15 垂直业务的大厂级方案

### C.2 通用中台重构
- v2：11 中台（im/wallet/order/refund/payservice/credit/notification/express/ad/sensitiveword/jubao）
- v3：12 中台（user/pay/im/merchant/distribution/marketing/risk/lbs/ai/tenant/material/diy）
- **变化**：
  - order+refund → 合并到 pay 中台（内部分 order/refund 子域）
  - payservice → 拆为 marketing + distribution 两个独立中台
  - credit+sensitiveword+jubao → 合并到 risk 中台（内部分 credit/audit/report 子域）
  - express → 合并到 material 或 lbs 中台
  - ad → 合并到 marketing 中台
  - notification → 合并到 im 中台
  - 新增：user / merchant / ai / tenant / material / diy 6 个独立中台

### C.3 垂直业务扩展
- v2：8 个同城基础业务
- v3：15 个垂直业务（新增分类信息核心/头条资讯/圈子社群/活动/教育培训/装修/直播，DIY 移出为前端能力中台）

### C.4 后台架构升级
- v2：未明确后台架构
- v3：五中心分离架构（工作台/模块中心/中台中心/设置中心/数据中心）+ 三层后台体系（平台/分站/商家）

### C.5 开发计划调整
- v2：6 阶段 22 周（1.5+4+3+10.5+2+1）
- v3：5 阶段 14.5 周（1.5+4+6+2+1）
- **变化**：
  - ershou 完整 + 其余垂直业务合并为 P2（6 周，3 批次内部并行）
  - 多端适配跟随 P2 每个模块同步做（不单独 P3）
  - P3 改为数据中台+推荐系统

### C.6 新增内容
- 新增第二章「点微+西瓜去重提纯」（17 类择优整合表）
- 新增第七章「后台五中心分离架构」
- 新增"中台内部子域分层"设计
- 新增"三层后台体系"设计
- 新增"DIY 独立为前端能力中台"

### C.7 关键技术决策保持
- 本地消息表替代 Seata
- Go 替代 Flink 自建推荐系统
- Docker Compose 替代 K8s
- WebSocket 单实例 MVP
- Prometheus+Grafana 轻量监控

### C.8 v3.1 大厂级能力补强（用户原话："是最佳方案吗？不要怕麻烦，要最佳实现"）

用户审核反馈："是最佳方案吗？不要怕麻烦，要最佳实现" → 触发 v3.1 大厂级能力补强，新增 8 章节 16 项能力：

**新增章节**：
- 第十六章 链路追踪与可观测性（OpenTelemetry + Jaeger + Loki + 4级告警）
- 第十七章 安全防护体系（WAF / SQL注入 / XSS / CSRF / DDoS / 接口签名防重放 / 敏感数据加密）
- 第十八章 分布式锁与缓存防护（RedLock + 布隆过滤器 + singleflight + 多级缓存 + 熔断降级）
- 第十九章 消息幂等与全链路一致性（接口幂等 + 消息消费幂等 + 12 类业务幂等场景清单）
- 第二十章 配置中心与 API 治理（第 13 个中台 config + 热更新 + API 版本管理 + Swagger）
- 第二十一章 测试与质量保障（单元/集成/E2E/压测 四层 + k6 压测脚本 + 压测指标标准）
- 第二十二章 数据库高可用与扩展（读写分离 + 分库分表 + 同城双活 + 异地灾备 + 3-2-1 备份）
- 第二十三章 微服务演进路径（服务注册发现预留 + 拆分三阶段 + 代码零修改 + Istio 预留）

**关键技术选型**：
| 能力 | 选型 | 理由 |
|------|------|------|
| 链路追踪 | OpenTelemetry + Jaeger | CNCF 标准，厂商无关 |
| 日志体系 | Loki + Promtail（MVP）/ ELK（后期） | Loki 成本低，ELK 全文检索强 |
| 配置中心 | 独立 config 中台（第 13 个中台） | 应用/业务/模块/分站/灰度 5 级配置 |
| 分布式锁 | Redis RedLock + Lua 脚本 | 金融级防重，原子性保证 |
| 缓存防护 | 布隆过滤器 + singleflight + 随机过期 | 穿透/击穿/雪崩三件套 |
| 消息幂等 | message_idempotent 表 + 唯一约束 | 全链路幂等，杜绝资损 |
| API 文档 | swaggo/gin-swagger | 注解自动生成，与 Gin 原生集成 |
| 压测工具 | k6 | JS 脚本化，云端分布式 |
| 灾备 | 同城双活（Patroni+etcd）+ 异地灾备（PG 逻辑复制） | RPO=0, RTO<1min |
| 微服务预留 | ServiceDiscovery 接口抽象 | 当前 mock，后期接入 Consul |

**新增表清单（5 张）**：
1. trace_samples - 链路追踪采样表
2. message_idempotent - 消息消费幂等表
3. config_items - 配置项定义表
4. config_grayscales - 配置灰度规则表
5. config_audit_logs - 配置变更审计表

**审核清单新增**（第十五章 15.10）：
- 13 项大厂级能力审核要点（链路追踪/日志/告警/安全/锁/缓存/幂等/配置/API/测试/数据库/灾备/微服务）

### C.9 v3.2 补强细节（用户原话："再次深度审计 v3.1" → "全部补齐(推荐)"）

用户在 v3.1 审核通过后，主动要求"再次深度审计 v3.1"，对全 23 章节 + 3 附录进行逐章深度审计，识别出 25 项可补强的技术细节（8 高优先级金融安全 + 12 中优先级大厂标准 + 5 低优先级运维完善），用户确认"全部补齐(推荐)"，文档升级为 v3.2。

**新增章节**：
- 第二十四章 补强细节（v3.2 新增 25 项）
  - 24.1 高优先级补强（8 项，金融安全核心）
  - 24.2 中优先级补强（12 项，大厂标准完整性）
  - 24.3 低优先级补强（5 项，运维完善度）

**25 项补强清单**：

| 序号 | 优先级 | 能力 | 关键设计 |
|------|--------|------|----------|
| 1 | 高 | RedLock 完整算法 | 5 实例 quorum + 耗时校验 + 分阶段引入（MVP 单实例 → 中期双实例 → 远期 5 实例） |
| 2 | 高 | 退款资金来源明确 | 组合方案（按订单状态自动选择：冻结资金释放 / 卖家余额扣减 / 平台垫付+欠款记录） |
| 3 | 高 | 二次验证 2FA | 7 类场景（提现/大额支付/超大额/改密码/改手机号/商家入驻/管理员）+ 短信/人脸/TOTP 三通道 + 5 次防爆破 |
| 4 | 高 | 分库分表分布式事务 | 分阶段引入（MVP 本地消息表 → 中期 Saga 补偿 → 远期 DTM 框架） |
| 5 | 高 | 灾备 DNS 切换 | GSLB + 智能DNS + 5s 健康检查 + 30s TTL + 30s 切换 |
| 6 | 高 | 微服务双写一致性校验 | 5 分钟定时对账 + 差异率告警(>0.01%) + 自动修复（以新库为准） |
| 7 | 高 | JWT 刷新机制 | Access Token(30min) + Refresh Token(7天) + Redis 黑名单 + token_version 密码修改失效 |
| 8 | 高 | 行为分析防刷 | 设备指纹 + IP 画像(住宅/机房/VPN) + 行为频率 + 风险评分(0-100) + pass/verify/block 三级决策 |
| 9 | 中 | WebSocket 集群 | Redis Pub/Sub 广播方案（ClusterHub + 本地 Hub + 实例 ID） |
| 10 | 中 | 布隆过滤器持久化 | Redis Bitmap + 双 key 原子切换 + 每日重建 |
| 11 | 中 | 消息幂等表清理 | PostgreSQL 分区表（按 consumed_at 月度） + 30 天 DROP + 死信保留 90 天 |
| 12 | 中 | 配置中心高可用降级 | 三级降级（本地缓存→Redis→DB→编译时默认值） + 30s 定时拉取降级 |
| 13 | 中 | 日志脱敏 | zerolog Hook + 11 类敏感字段（phone/idcard/bankcard/password/token 等）自动脱敏 |
| 14 | 中 | 统一审计日志表 | audit_logs（操作人/类型/目标/前后值/IP/UA/模块） + 月度分区 + 1 年保留 |
| 15 | 中 | 文件上传安全 | 5 层防护（大小/扩展名/MIME/图片防马/病毒扫描 ClamAV） |
| 16 | 中 | 第三方 API 容错 | 超时(3s 连接/5s 响应头/10s 总) + 3 次指数退避重试 + 熔断器(5 次失败/30s 熔断) + 缓存降级 |
| 17 | 中 | 资金对账差异处理 | 分级处理（<0.01 元自动修复 / 0.01-1 元人工审核 / >1 元告警+暂停提现） |
| 18 | 中 | 商家结算发票管理 | merchant_invoices 表 + 电子/纸质发票 + 金蝶/百望 API |
| 19 | 中 | 多端数据同步 | WebSocket 实时推送 + 增量同步(last_sync_at) + 服务端为准冲突解决 |
| 20 | 中 | 蓝绿部署/金丝雀发布 | 蓝绿(30s DNS 切换) + 金丝雀(5%→20%→50%→100% 四阶段) + Nginx split_clients |
| 21 | 低 | 监控长期存储 | Thanos（Sidecar+Query+Compactor+Store） + OSS + 降采样(5min/1h) |
| 22 | 低 | 容器资源限制 | Docker Compose deploy.resources + OOM 处理 + cAdvisor 监控 |
| 23 | 低 | 连接池配置 | PG(MaxOpenConns=100/MaxIdleConns=10) + Redis(PoolSize=50/MinIdleConns=10) |
| 24 | 低 | 慢查询优化 | EXPLAIN ANALYZE + 索引建议 + SQL 重写(避免 SELECT */N+1/大 OFFSET) + 缓存策略 |
| 25 | 低 | CDN 缓存策略 | 三级缓存(CDN→Nginx→Redis) + 刷新机制(URL/目录/预热) + 命中率监控(>95% 目标) |

**新增表清单（6 张）**：
1. user_2fa_codes - 二次验证码表（24.1 高优先级 - 2FA）
2. user_device_fingerprints - 用户设备指纹表（24.1 高优先级 - 行为分析防刷）
3. ip_profiles - IP 画像表（24.1 高优先级 - 行为分析防刷）
4. behavior_analyses - 行为分析记录表（24.1 高优先级 - 行为分析防刷）
5. audit_logs - 统一审计日志表（24.2 中优先级 - 统一审计日志表）
6. merchant_invoices - 商家发票表（24.2 中优先级 - 商家结算发票管理）

**关键技术选型补强**：
| 能力 | 选型 | 理由 |
|------|------|------|
| 分布式锁完整版 | RedLock 5 实例 quorum | 金融级防重，耗时校验避免 GC 停顿导致锁失效 |
| 2FA 三通道 | 短信 + 人脸 + TOTP(Google Authenticator) | 多通道容灾，单通道故障可降级 |
| 分布式事务演进 | 本地消息表 → Saga → DTM | 分阶段引入，MVP 不引入复杂度，远期支持大厂级 |
| 灾备 DNS | GSLB + 智能DNS（DNSPod/阿里云DNS） | 30s 内全球切换，RTO<1min |
| JWT 双 Token | Access(30min) + Refresh(7d) + token_version | 短 Access 限风险，长 Refresh 体验，密码修改全失效 |
| 风险引擎 | 自研评分模型（0-100 分 + 三级决策） | 灵活可调，按场景配置阈值 |
| WebSocket 集群 | Redis Pub/Sub 广播 | MVP 单实例平滑扩展到多实例，无状态 |
| 布隆过滤器持久化 | Redis Bitmap + 双 key 切换 | 避免重启全量回灌，原子切换零停机 |
| 幂等表清理 | PostgreSQL 月度分区 + 自动 DROP | 自动清理，避免无限膨胀 |
| 配置中心降级 | 四级降级（本地→Redis→DB→默认值） | 任何一级故障都有兜底，永不阻断业务 |
| 日志脱敏 | zerolog Hook | 全局拦截，零侵入业务代码 |
| 审计日志 | audit_logs 月度分区 | 1 年保留满足合规要求，分区查询高效 |
| 文件上传安全 | 5 层防护 + ClamAV | 杜绝图片马、病毒文件上传 |
| 第三方 API 容错 | 超时 + 重试 + 熔断 + 降级四件套 | 任何第三方故障不拖垮主流程 |
| 对账差异分级 | 三级处理（自动/审核/告警+暂停） | 平衡自动化与资金安全 |
| 发票管理 | merchant_invoices + 金蝶/百望 | 电子发票自动化，纸质发票流程化 |
| 多端同步 | WebSocket + 增量同步 + 服务端为准 | 实时性强，冲突解决明确 |
| 蓝绿金丝雀 | Nginx split_clients + DNS 切换 | 单模块独立灰度，无需 K8s |
| 监控长期存储 | Thanos + OSS | PromQL 兼容，长期存储成本低 |
| 容器资源限制 | Docker Compose deploy.resources | 单容器 OOM 不拖垮宿主 |
| 连接池 | PG 100/10 + Redis 50/10 | 匹配 MVP 并发预估，避免连接风暴 |
| 慢查询 SOP | EXPLAIN ANALYZE + 索引建议 | 形成标准化优化流程 |
| CDN 三级缓存 | CDN→Nginx→Redis | 命中率 >95%，回源率 <5% |

**审核清单新增**（第十五章 15.11）：
- 25 项 v3.2 补强审核要点（高 8 + 中 12 + 低 5）

**v3.2 文档总规模**：24 章 + 3 附录，约 5200 行，约 65 张新表

**与建议.md 对照结论**：v3.2 在 v3.1 完整覆盖建议.md 全部内容的基础上，补强 25 项大厂级技术细节，建议.md 全部要求（11 中台 + 16 垂直业务 + 模块独立管控 + 多端统一 + 5 阶段计划）均已落地，且在金融安全、运维完善度、大厂标准完整性三个维度有显著超越。
