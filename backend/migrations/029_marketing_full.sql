-- =====================================================
-- marketing 营销活动中台模块完整迁移脚本
-- 包含 6 张表 + 索引 + 触发器 + 注释
-- 依据需求文档：对标淘宝/京东/拼多多/美团营销中台
-- 数据库：wuchang_tongcheng
-- =====================================================
--
-- 内容：
--   1. CREATE 6 张表（marketing 模块，无 mall_ 前缀，表名严格遵循 model.TableName()）
--   2. 索引、触发器、注释
--   3. 全幂等：CREATE TABLE IF NOT EXISTS
-- 依赖：001_p0_baseline.sql（update_updated_at_column 函数已创建）
--
-- 表清单（6 张）：
--   ad_positions   广告位（RegionBaseModel，地区隔离）
--   coupons        优惠券（RegionBaseModel，地区隔离）
--   user_coupons   用户优惠券（BaseModel，用户隔离）
--   sign_records   签到记录（BaseModel，用户隔离）
--   sign_rules     签到规则（BaseModel）
--   activities     营销活动（RegionBaseModel，地区隔离）
--
-- 说明：表名严格遵循 model.TableName() 定义：
--      - AdPosition → ad_positions（复数）
--      - Coupon → coupons（复数）
--      - UserCoupon → user_coupons（复数）
--      - SignRecord → sign_records（复数）
--      - SignRule → sign_rules（复数）
--      - Activity → activities（复数）
-- =====================================================

-- ============================================================
-- 1. ad_positions 广告位表
-- ============================================================
CREATE TABLE IF NOT EXISTS ad_positions (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    -- 广告位基本信息
    position_code VARCHAR(50) NOT NULL DEFAULT '',           -- 位置编码：home_banner/list_top/detail_banner/category_top/search_top/popup
    title VARCHAR(100) NOT NULL DEFAULT '',                  -- 广告标题
    image_url VARCHAR(500) NOT NULL DEFAULT '',              -- 广告图片 URL
    link_url VARCHAR(500) NOT NULL DEFAULT '',               -- 跳转链接
    sort INT NOT NULL DEFAULT 0,                             -- 排序（升序）

    -- 生效时间
    start_at TIMESTAMPTZ,
    end_at TIMESTAMPTZ,

    -- 状态：0禁用 1启用 2待生效 3已过期
    status INT NOT NULL DEFAULT 1
);
CREATE INDEX IF NOT EXISTS idx_ad_positions_region_id ON ad_positions(region_id);
CREATE INDEX IF NOT EXISTS idx_ad_positions_code ON ad_positions(position_code);
CREATE INDEX IF NOT EXISTS idx_ad_positions_sort ON ad_positions(sort);
CREATE INDEX IF NOT EXISTS idx_ad_positions_start_at ON ad_positions(start_at);
CREATE INDEX IF NOT EXISTS idx_ad_positions_end_at ON ad_positions(end_at);
CREATE INDEX IF NOT EXISTS idx_ad_positions_status ON ad_positions(status);
CREATE INDEX IF NOT EXISTS idx_ad_positions_deleted_at ON ad_positions(deleted_at);

COMMENT ON TABLE ad_positions IS '广告位表（首页Banner/列表置顶/详情广告/分类置顶/搜索置顶/弹窗）';
COMMENT ON COLUMN ad_positions.position_code IS '位置编码：home_banner/list_top/detail_banner/category_top/search_top/popup';
COMMENT ON COLUMN ad_positions.status IS '状态：0禁用 1启用 2待生效 3已过期';

-- ============================================================
-- 2. coupons 优惠券表
-- ============================================================
CREATE TABLE IF NOT EXISTS coupons (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    -- 基本信息
    title VARCHAR(100) NOT NULL DEFAULT '',                  -- 优惠券标题
    type VARCHAR(20) NOT NULL DEFAULT 'reduce',              -- 类型：discount折扣/reduce满减/exchange兑换

    -- 金额（DECIMAL(12,2) 支持大额）
    amount DECIMAL(12,2) NOT NULL DEFAULT 0,                 -- 面值/折扣率（折扣券 0.01-0.99）
    threshold DECIMAL(12,2) NOT NULL DEFAULT 0,              -- 使用门槛（满 N 元可用）

    -- 库存
    total_count INT NOT NULL DEFAULT 0,                      -- 发放总量（0=不限）
    received_count INT NOT NULL DEFAULT 0,                   -- 已领取数

    -- 领取时间
    start_at TIMESTAMPTZ,
    end_at TIMESTAMPTZ,

    -- 状态：0禁用 1进行中 2草稿 3已下架 4已过期 5已抢完
    status INT NOT NULL DEFAULT 1
);
CREATE INDEX IF NOT EXISTS idx_coupons_region_id ON coupons(region_id);
CREATE INDEX IF NOT EXISTS idx_coupons_title ON coupons(title);
CREATE INDEX IF NOT EXISTS idx_coupons_type ON coupons(type);
CREATE INDEX IF NOT EXISTS idx_coupons_status ON coupons(status);
CREATE INDEX IF NOT EXISTS idx_coupons_start_at ON coupons(start_at);
CREATE INDEX IF NOT EXISTS idx_coupons_end_at ON coupons(end_at);
CREATE INDEX IF NOT EXISTS idx_coupons_deleted_at ON coupons(deleted_at);

COMMENT ON TABLE coupons IS '优惠券表（满减/折扣/兑换）';
COMMENT ON COLUMN coupons.type IS '类型：discount折扣券/reduce满减券/exchange兑换券';
COMMENT ON COLUMN coupons.amount IS '面值/折扣率（折扣券 0.01-0.99）';
COMMENT ON COLUMN coupons.threshold IS '使用门槛（满 N 元可用）';
COMMENT ON COLUMN coupons.total_count IS '发放总量（0=不限）';
COMMENT ON COLUMN coupons.status IS '状态：0禁用 1进行中 2草稿 3已下架 4已过期 5已抢完';

-- ============================================================
-- 3. user_coupons 用户优惠券表（用户隔离，无 region_id）
-- ============================================================
CREATE TABLE IF NOT EXISTS user_coupons (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    -- 关联
    user_id BIGINT NOT NULL,                                 -- 用户 ID
    coupon_id BIGINT NOT NULL,                               -- 优惠券 ID

    -- 状态与来源
    status VARCHAR(20) NOT NULL DEFAULT 'unused',            -- unused未使用/used已使用/expired已过期
    source VARCHAR(32) NOT NULL DEFAULT 'receive',           -- receive主动领取/gift系统赠送/activity活动奖励/new_user新人礼包

    -- 使用信息
    used_at TIMESTAMPTZ,                                     -- 使用时间
    order_id BIGINT NOT NULL DEFAULT 0                       -- 使用的订单 ID
);
CREATE INDEX IF NOT EXISTS idx_user_coupons_user ON user_coupons(user_id);
CREATE INDEX IF NOT EXISTS idx_user_coupons_coupon ON user_coupons(coupon_id);
CREATE INDEX IF NOT EXISTS idx_user_coupons_status ON user_coupons(status);
CREATE INDEX IF NOT EXISTS idx_user_coupons_source ON user_coupons(source);
CREATE INDEX IF NOT EXISTS idx_user_coupons_order_id ON user_coupons(order_id);
CREATE INDEX IF NOT EXISTS idx_user_coupons_deleted_at ON user_coupons(deleted_at);

COMMENT ON TABLE user_coupons IS '用户优惠券表（领取/使用/过期记录）';
COMMENT ON COLUMN user_coupons.status IS '状态：unused未使用/used已使用/expired已过期';
COMMENT ON COLUMN user_coupons.source IS '来源：receive主动领取/gift系统赠送/activity活动奖励/new_user新人礼包';

-- ============================================================
-- 4. sign_records 签到记录表（用户隔离，无 region_id）
-- ============================================================
CREATE TABLE IF NOT EXISTS sign_records (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    -- 关联
    user_id BIGINT NOT NULL,                                 -- 用户 ID

    -- 签到信息
    sign_date DATE NOT NULL,                                 -- 签到日期
    continuous_days INT NOT NULL DEFAULT 1,                  -- 连续签到天数
    points INT NOT NULL DEFAULT 0                            -- 本次签到获得积分
);
CREATE INDEX IF NOT EXISTS idx_sign_records_user ON sign_records(user_id);
CREATE INDEX IF NOT EXISTS idx_sign_records_date ON sign_records(sign_date);
CREATE INDEX IF NOT EXISTS idx_sign_records_deleted_at ON sign_records(deleted_at);

COMMENT ON TABLE sign_records IS '签到记录表（每日签到/连续签到/积分发放）';
COMMENT ON COLUMN sign_records.sign_date IS '签到日期（DATE 类型，每日仅一条）';
COMMENT ON COLUMN sign_records.continuous_days IS '连续签到天数（昨日有签到则 +1，否则重置为 1）';
COMMENT ON COLUMN sign_records.points IS '本次签到获得积分（命中规则取规则积分，否则基础 1 积分）';

-- ============================================================
-- 5. sign_rules 签到规则表
-- ============================================================
CREATE TABLE IF NOT EXISTS sign_rules (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    -- 规则配置
    day INT NOT NULL,                                        -- 连续签到第 N 天
    points INT NOT NULL DEFAULT 0,                           -- 奖励积分
    extra_reward JSONB,                                      -- 额外奖励（JSONB，如优惠券/积分倍数）

    -- 状态：0禁用 1启用
    status INT NOT NULL DEFAULT 1
);
-- 唯一索引：同一天数规则唯一
CREATE UNIQUE INDEX IF NOT EXISTS uniq_sign_rules_day ON sign_rules(day) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_sign_rules_status ON sign_rules(status);
CREATE INDEX IF NOT EXISTS idx_sign_rules_deleted_at ON sign_rules(deleted_at);

COMMENT ON TABLE sign_rules IS '签到规则表（按连续签到天数配置奖励积分与额外奖励）';
COMMENT ON COLUMN sign_rules.day IS '连续签到第 N 天（唯一）';
COMMENT ON COLUMN sign_rules.points IS '奖励积分';
COMMENT ON COLUMN sign_rules.extra_reward IS '额外奖励 JSONB（如优惠券/积分倍数）';
COMMENT ON COLUMN sign_rules.status IS '状态：0禁用 1启用';

-- ============================================================
-- 6. activities 营销活动表
-- ============================================================
CREATE TABLE IF NOT EXISTS activities (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    -- 基本信息
    title VARCHAR(100) NOT NULL DEFAULT '',                  -- 活动标题
    type VARCHAR(20) NOT NULL DEFAULT 'groupbuy',            -- 类型：groupbuy拼团/bargain砍价/seckill秒杀/lottery抽奖
    description TEXT,                                        -- 活动描述
    cover_image VARCHAR(500) NOT NULL DEFAULT '',            -- 封面图

    -- 活动时间
    start_at TIMESTAMPTZ,
    end_at TIMESTAMPTZ,

    -- 状态：0禁用 1待开始 2进行中 3已结束 4已取消
    status INT NOT NULL DEFAULT 1,

    -- 活动配置（JSONB）
    config JSONB                                             -- 活动配置（如拼团人数/砍价底价/秒杀库存/抽奖概率）
);
CREATE INDEX IF NOT EXISTS idx_activities_region_id ON activities(region_id);
CREATE INDEX IF NOT EXISTS idx_activities_title ON activities(title);
CREATE INDEX IF NOT EXISTS idx_activities_type ON activities(type);
CREATE INDEX IF NOT EXISTS idx_activities_status ON activities(status);
CREATE INDEX IF NOT EXISTS idx_activities_start_at ON activities(start_at);
CREATE INDEX IF NOT EXISTS idx_activities_end_at ON activities(end_at);
CREATE INDEX IF NOT EXISTS idx_activities_deleted_at ON activities(deleted_at);

COMMENT ON TABLE activities IS '营销活动表（拼团/砍价/秒杀/抽奖）';
COMMENT ON COLUMN activities.type IS '类型：groupbuy拼团/bargain砍价/seckill秒杀/lottery抽奖';
COMMENT ON COLUMN activities.status IS '状态：0禁用 1待开始 2进行中 3已结束 4已取消';
COMMENT ON COLUMN activities.config IS '活动配置 JSONB（如拼团人数/砍价底价/秒杀库存/抽奖概率）';

-- ============================================================
-- updated_at 触发器（依赖 001_p0_baseline.sql 中的 update_updated_at_column 函数）
-- PostgreSQL 不支持 CREATE TRIGGER IF NOT EXISTS，用 DROP TRIGGER IF EXISTS + CREATE TRIGGER
-- ============================================================
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'update_updated_at_column') THEN
        DROP TRIGGER IF EXISTS trg_ad_positions_updated_at ON ad_positions; CREATE TRIGGER trg_ad_positions_updated_at BEFORE UPDATE ON ad_positions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_coupons_updated_at ON coupons; CREATE TRIGGER trg_coupons_updated_at BEFORE UPDATE ON coupons FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_user_coupons_updated_at ON user_coupons; CREATE TRIGGER trg_user_coupons_updated_at BEFORE UPDATE ON user_coupons FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_sign_records_updated_at ON sign_records; CREATE TRIGGER trg_sign_records_updated_at BEFORE UPDATE ON sign_records FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_sign_rules_updated_at ON sign_rules; CREATE TRIGGER trg_sign_rules_updated_at BEFORE UPDATE ON sign_rules FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_activities_updated_at ON activities; CREATE TRIGGER trg_activities_updated_at BEFORE UPDATE ON activities FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;

-- ============================================================
-- 完成
-- ============================================================
-- 表清单（6 张）：
--   ad_positions / coupons / user_coupons / sign_records / sign_rules / activities
-- 索引：20+ 个（含唯一索引 uniq_sign_rules_day）
-- 触发器：6 个 updated_at 触发器
-- 注释：表注释 + 关键字段注释
-- ============================================================
