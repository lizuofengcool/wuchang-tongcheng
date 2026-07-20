# 【阶段3 第2批模块开发：love/pinche/linggong/dh114】开发进度文档

- 开发人员：AI助手（5 Agent 并行）
- 创建时间：2026-07-20
- 当前状态：🔄 启动阶段3第2批4模块全栈开发（对标行业第一）

## 一、开发目标

依据用户规则"必须对标同类行业第一的功能，不要做一般的模块"，按 ershou 样板标准开发第2批4个垂直业务模块：love 相亲交友 / pinche 拼车出行 / linggong 零工兼职 / dh114 同城114。

### 对标平台与差异化能力

| 模块 | 行业第一对标 | 关键差异化能力 |
|------|------------|--------------|
| love 相亲交友 | Soul/陌陌/探探/百合网 | 灵魂匹配算法/语音匹配/实名认证/会员等级/心动信号/隐私保护/视频认证 |
| pinche 拼车出行 | 哈啰出行/嘀嗒出行/滴滴顺风车 | 行程智能匹配/顺风车保险/车主认证/行程分享/紧急联系人/ETC支付 |
| linggong 零工兼职 | 斗米/青团兼职/兼职猫/猪八戒 | 任务制/薪资日结/雇主双向评价/技能标签/信用分/兼职合同电子化 |
| dh114 同城114 | 大众点评/美团/58同城 | 商家黄页/评价体系/团购/电话一键拨号/地图导航/收藏夹/商户认证 |

### 核心交付（每个模块）

- 数据库：18 张表（主表+17张子表）
- 后端：65+ API（CRUD+交易+审核+统计+批量+推荐）
- 前端管理后台：16 页面
- 前端C端：12 页面 + 10 组件
- 中台对接：pay 担保交易 / im 聊天 / material 以图搜图 / risk 举报 / ai 审核

### 总工作量

- 数据库：4模块 × 18表 = 72 张表
- 后端：4模块 × 65API = 260+ API
- 前端管理后台：4模块 × 16页面 = 64 页面
- 前端C端：4模块 × 12页面 + 40组件 = 48 页面 + 40 组件

## 二、表前缀规范

| 模块 | 前缀 | 主表 | 典型子表 |
|------|------|------|---------|
| love | `love_` | `loves` | `love_profiles`, `love_matches`, `love_stories`, `love_member_levels`, `love_verifications`, `love_impressions`, `love_visits`, `love_blocks`, `love_gifts`, `love_memberships` |
| pinche | `pinche_` | `pinches` | `pinche_routes`, `pinche_bookings`, `pinche_drivers`, `pinche_vehicles`, `pinche_insurances`, `pinche_payments`, `pinche_ratings`, `pinche_emergencies`, `pinche_trips` |
| linggong | `linggong_` | `linggongs` | `linggong_tasks`, `linggong_applications`, `linggong_employers`, `linggong_workers`, `linggong_contracts`, `linggong_payments`, `linggong_ratings`, `linggong_skills`, `linggong_credits` |
| dh114 | `dh114_` | `dh114s` | `dh114_businesses`, `dh114_categories`, `dh114_reviews`, `dh114_groupbuys`, `dh114_coupons`, `dh114_phone_calls`, `dh114_favorites`, `dh114_verifications`, `dh114_statistics` |

## 三、5 Agent 并行策略

| Agent | 任务范围 | 依赖 | 步骤预算 |
|-------|---------|------|---------|
| Agent A 数据库 | 4模块共72张表+迁移脚本+model | 无 | 立即开始 |
| Agent B 后端 | 4模块共260+ API（model/dto/repository/service/handler） | Agent A | Agent A 完成后启动 |
| Agent C 管理后台 | 4模块共64页面（frontend/src/views/business/{love,pinche,linggong,dh114}/） | Agent B | Agent B 完成后启动 |
| Agent D C端 | 4模块共48页面+40组件（frontend/app/pages/{love,pinche,linggong,dh114}/） | Agent B | Agent B 完成后启动 |
| Agent E 中台对接 | 4模块共用 pay/im/material/risk/ai（已就绪） | 无 | 与 Agent A 并行 |

**执行顺序**：
1. Agent A 数据库先行（无依赖）
2. Agent A 完成后 → Agent B 后端 + Agent E 中台对接 并行
3. Agent B 完成后 → Agent C 管理后台 + Agent D C端 并行
4. 全部完成后 → 集成验证 + commit push

## 四、love 相亲交友模块设计（对标 Soul/陌陌/探探）

### 4.1 数据库表（18张）

| 表名 | 用途 | 关键字段 |
|------|------|---------|
| `loves`（主表） | 用户交友档案 | nickname/gender/age/city/bio/avatar/tags |
| `love_profiles` | 详细资料 | height/weight/education/income/job/zodiac/blood_type/marital_status |
| `love_matches` | 匹配记录 | user_id/target_id/match_score/match_type/status |
| `love_stories` | 动态广场 | content/images/videos/likes/comments/location |
| `love_member_levels` | 会员等级 | level_name/price/privileges/discount |
| `love_verifications` | 实名/视频/学历认证 | type/status/real_name/id_card/video_url |
| `love_impressions` | 印象标签 | tag_name/tag_count/target_user_id |
| `love_visits` | 访客记录 | visitor_id/visited_id/visit_count/last_visit_at |
| `love_blocks` | 拉黑名单 | user_id/blocked_id/reason |
| `love_gifts` | 虚拟礼物 | name/icon/price/description/category |
| `love_memberships` | 会员订阅 | user_id/level_id/start_at/end_at/payment_id |
| `love_likes` | 喜欢/不喜欢 | user_id/target_id/action(like/pass)/super_like |
| `love_chat_sessions` | 匹配后聊天 | user_a/user_b/matched_at/last_message_at |
| `love_notifications` | 通知 | type(content/like/visit/gift/match)/recipient/read_at |
| `love_recommendations` | 推荐池 | user_id/recommended_id/score/reason |
| `love_privacy_settings` | 隐私设置 | hide_online/hide_location/hide_age/allow_stranger_chat |
| `love_reports` | 举报 | reporter/reported/reason/evidence/status |
| `love_audit_rules` | 审核规则 | field/rule/severity/action |

### 4.2 关键差异化能力

1. **灵魂匹配算法**（对标 Soul）：基于兴趣标签/性格测试/价值观/生活模式计算匹配度
2. **语音匹配**（对标 Soul）：语音简介 + 算法匹配
3. **视频认证**（对标 陌陌）：人脸识别 + 视频自拍认证
4. **会员等级**（对标 探探/百合网）：基础/高级/VIP/Premium 四级会员
5. **心动信号**（对标 探探）：Super Like 每天限量
6. **隐私保护**（对标 Soul）：可隐藏在线状态/位置/年龄
7. **印象标签**（对标 Soul）：他人评价，自我认知补充

### 4.3 API 数量：65+

## 五、pinche 拼车出行模块设计（对标 哈啰/嘀嗒/滴滴）

### 5.1 数据库表（18张）

| 表名 | 用途 |
|------|------|
| `pinches`（主表） | 行程信息 |
| `pinche_routes` | 路线（起点/终点/途经点/距离/时长） |
| `pinche_bookings` | 预订记录 |
| `pinche_drivers` | 车主认证 |
| `pinche_vehicles` | 车辆信息 |
| `pinche_insurances` | 顺风车保险 |
| `pinche_payments` | 支付记录 |
| `pinche_ratings` | 评价 |
| `pinche_emergencies` | 紧急联系人/一键报警 |
| `pinche_trips` | 完成行程 |
| `pinche_route_favorites` | 常用路线收藏 |
| `pinche_driver_locations` | 实时位置 |
| `pinche_messages` | 行程内消息 |
| `pinche_cancels` | 取消记录 |
| `pinche_refunds` | 退款记录 |
| `pinche_complaints` | 投诉 |
| `pinche_audit_rules` | 审核规则 |
| `pinche_statistics` | 统计 |

### 5.2 关键差异化能力

1. **行程智能匹配**（对标 哈啰）：路线+时间+座位+车主评分综合匹配
2. **顺风车保险**（对标 嘀嗒）：每次行程自动投保
3. **车主认证**（对标 滴滴）：身份证+驾驶证+行驶证+车辆照片
4. **行程分享**（对标 滴滴）：分享给紧急联系人，实时位置
5. **紧急联系人**（对标 滴滴）：一键报警 + 行程分享
6. **ETC支付**（对标 哈啰）：ETC 自动扣费

## 六、linggong 零工兼职模块设计（对标 斗米/青团/兼职猫）

### 6.1 数据库表（18张）

| 表名 | 用途 |
|------|------|
| `linggongs`（主表） | 兼职岗位 |
| `linggong_tasks` | 任务包 |
| `linggong_applications` | 报名记录 |
| `linggong_employers` | 雇主认证 |
| `linggong_workers` | 求职者档案 |
| `linggong_contracts` | 电子合同 |
| `linggong_payments` | 薪资支付 |
| `linggong_ratings` | 双向评价 |
| `linggong_skills` | 技能标签 |
| `linggong_credits` | 信用分 |
| `linggong_certifications` | 资质证书 |
| `linggong_attendances` | 考勤打卡 |
| `linggong_disputes` | 纠纷 |
| `linggong_withdrawals` | 提现 |
| `linggong_recommendations` | 推荐岗位 |
| `linggong_favorites` | 收藏 |
| `linggong_audit_rules` | 审核规则 |
| `linggong_statistics` | 统计 |

### 6.2 关键差异化能力

1. **任务制**（对标 斗米）：长短期任务分类，按件/按时/按日计费
2. **薪资日结**（对标 兼职猫）：T+0/T+1/T+7 多种结算方式
3. **雇主双向评价**（对标 美团）：工人评价雇主 + 雇主评价工人
4. **技能标签**（对标 猪八戒）：技能认证 + 标签匹配
5. **信用分**（对标 芝麻信用）：履约/违约影响后续接单
6. **兼职合同电子化**（对标 法大大）：在线签署电子合同

## 七、dh114 同城114模块设计（对标 大众点评/美团/58同城）

### 7.1 数据库表（18张）

| 表名 | 用途 |
|------|------|
| `dh114s`（主表） | 商家信息 |
| `dh114_categories` | 商家分类 |
| `dh114_businesses` | 商家详细 |
| `dh114_reviews` | 评价 |
| `dh114_groupbuys` | 团购 |
| `dh114_coupons` | 优惠券 |
| `dh114_phone_calls` | 电话拨打记录 |
| `dh114_favorites` | 收藏 |
| `dh114_verifications` | 商户认证 |
| `dh114_business_hours` | 营业时间 |
| `dh114_images` | 商家图片 |
| `dh114_menus` | 菜单/服务项目 |
| `dh114_tags` | 标签 |
| `dh114_reviews_replies` | 商家回复 |
| `dh114_recommendations` | 推荐商家 |
| `dh114_visits` | 浏览记录 |
| `dh114_audit_rules` | 审核规则 |
| `dh114_statistics` | 统计 |

### 7.2 关键差异化能力

1. **商家黄页**（对标 58同城）：全行业分类
2. **评价体系**（对标 大众点评）：5星评分 + 文字 + 图片 + 视频
3. **团购**（对标 美团）：限时抢购 + 数量限制
4. **电话一键拨号**（对标 大众点评）：点击直接拨打
5. **地图导航**（对标 美团）：高德地图集成
6. **收藏夹**（对标 美团）：个人收藏夹 + 分组管理
7. **商户认证**（对标 美团）：营业执照 + 实地认证

## 八、约束（公共要求）

1. 所有表用 RegionBaseModel（含 region_id 地区隔离）
2. 金额字段 decimal(12,2)，不用 float
3. JSON 字段用 jsonb
4. 索引名遵循 PostgreSQL 默认命名
5. 字段类型严格对标头部平台
6. 注释完整
7. 每个模块的 service 层暴露接口给其他模块调用（不通过 HTTP，直接 import）
8. Plugin 接口实现 Name/Version/Meta/Init/RegisterRoutes/Close
9. 迁移脚本全幂等（IF NOT EXISTS）
10. updated_at 触发器
11. 主表通过 GORM AutoMigrate 创建，子表通过 SQL 迁移脚本创建
12. JSONB 字段不能用 default:'[]'::jsonb（避免 GORM AutoMigrate 生成错误 SQL）
13. Gin 路由参数名一致性（同一路径前缀下参数名必须一致）

## 九、待完成任务

- [ ] 任务1：创建迁移脚本 009-012（4模块）+ 4个rollback
- [ ] 任务2：love 模块全栈开发（model/dto/repository/service/handler/plugin）
- [ ] 任务3：pinche 模块全栈开发
- [ ] 任务4：linggong 模块全栈开发
- [ ] 任务5：dh114 模块全栈开发
- [ ] 任务6：错误码体系（love/pinche/linggong/dh114 4段）
- [ ] 任务7：编译验证 + 迁移应用 + backend 启动
- [ ] 任务8：API 测试 4模块
- [ ] 任务9：管理后台 64 页面
- [ ] 任务10：C端 48 页面 + 40 组件
- [ ] 任务11：Playwright 集成验证
- [ ] 任务12：commit push 到 Gitee

## 十、下一步

立即启动 Agent A 数据库开发，4模块共 72 张表 + 4 迁移脚本 + 4 rollback 脚本。
