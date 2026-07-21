-- =====================================================
-- mall 骑手端后端扩展完整迁移脚本
-- 包含 3 张表 + 索引 + 触发器 + 注释
-- 依据需求文档：对标美团/达达/顺丰同城骑手
-- 数据库：wuchang_tongcheng
-- =====================================================
--
-- 内容：
--   1. CREATE 3 张子表（mall_riders / mall_deliveries / mall_rider_settlements）
--   2. 索引、触发器、注释
--   3. 全幂等：CREATE TABLE IF NOT EXISTS
-- 依赖：001_p0_baseline.sql（update_updated_at_column 函数已创建）
--
-- 表清单：
--   mall_riders              骑手认证表
--   mall_deliveries          配送单表
--   mall_rider_settlements   骑手结算表
-- =====================================================

-- ============================================================
-- 1. mall_riders 骑手认证表
-- ============================================================
CREATE TABLE IF NOT EXISTS mall_riders (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    -- 关联
    user_id BIGINT NOT NULL,                       -- 关联 users
    shop_id BIGINT,                                -- 可选，绑定店铺

    -- 基本信息
    real_name VARCHAR(50) NOT NULL DEFAULT '',
    phone VARCHAR(20) NOT NULL DEFAULT '',
    id_card VARCHAR(18) NOT NULL DEFAULT '',
    avatar VARCHAR(255) NOT NULL DEFAULT '',

    -- 车辆信息
    vehicle_type VARCHAR(20) NOT NULL DEFAULT '',  -- 电动车/摩托车/自行车/汽车
    vehicle_plate VARCHAR(20) NOT NULL DEFAULT '',
    license_url VARCHAR(255) NOT NULL DEFAULT '',  -- 驾驶证

    -- 状态
    status SMALLINT NOT NULL DEFAULT 0,             -- 0待审核 1通过 2拒绝 3冻结
    credit_score INT NOT NULL DEFAULT 100,          -- 信用分
    level INT NOT NULL DEFAULT 1,                   -- 等级

    -- 统计
    total_orders INT NOT NULL DEFAULT 0,            -- 累计订单
    total_earnings DECIMAL(12,2) NOT NULL DEFAULT 0, -- 累计收入

    -- 在线状态
    online_status SMALLINT NOT NULL DEFAULT 0,      -- 0下线 1在线 2配送中
    audit_reason VARCHAR(500) NOT NULL DEFAULT ''   -- 审核理由
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_mall_riders_user_id ON mall_riders(user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_mall_riders_shop_id ON mall_riders(shop_id);
CREATE INDEX IF NOT EXISTS idx_mall_riders_status ON mall_riders(status);
CREATE INDEX IF NOT EXISTS idx_mall_riders_region_id ON mall_riders(region_id);
CREATE INDEX IF NOT EXISTS idx_mall_riders_online_status ON mall_riders(online_status);
CREATE INDEX IF NOT EXISTS idx_mall_riders_deleted_at ON mall_riders(deleted_at);

COMMENT ON TABLE mall_riders IS '同城商城骑手认证表';
COMMENT ON COLUMN mall_riders.user_id IS '关联 users 表用户 ID';
COMMENT ON COLUMN mall_riders.shop_id IS '可选，绑定的店铺 ID';
COMMENT ON COLUMN mall_riders.real_name IS '骑手真实姓名';
COMMENT ON COLUMN mall_riders.phone IS '联系电话';
COMMENT ON COLUMN mall_riders.id_card IS '身份证号';
COMMENT ON COLUMN mall_riders.avatar IS '头像 URL';
COMMENT ON COLUMN mall_riders.vehicle_type IS '车辆类型：电动车/摩托车/自行车/汽车';
COMMENT ON COLUMN mall_riders.vehicle_plate IS '车牌号';
COMMENT ON COLUMN mall_riders.license_url IS '驾驶证/资质照片 URL';
COMMENT ON COLUMN mall_riders.status IS '状态：0待审核 1通过 2拒绝 3冻结';
COMMENT ON COLUMN mall_riders.credit_score IS '信用分（默认 100）';
COMMENT ON COLUMN mall_riders.level IS '骑手等级';
COMMENT ON COLUMN mall_riders.total_orders IS '累计订单数';
COMMENT ON COLUMN mall_riders.total_earnings IS '累计收入';
COMMENT ON COLUMN mall_riders.online_status IS '在线状态：0下线 1在线 2配送中';
COMMENT ON COLUMN mall_riders.audit_reason IS '审核理由';

-- ============================================================
-- 2. mall_deliveries 配送单表
-- ============================================================
CREATE TABLE IF NOT EXISTS mall_deliveries (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    -- 关联
    order_id BIGINT NOT NULL,                       -- 关联 mall_orders
    rider_id BIGINT,                                -- 认领骑手
    shop_id BIGINT NOT NULL,                        -- 店铺 ID
    user_id BIGINT NOT NULL,                        -- 收货人用户 ID

    -- 配送单号
    delivery_no VARCHAR(50) NOT NULL DEFAULT '',    -- 配送单号

    -- 状态
    status SMALLINT NOT NULL DEFAULT 0,             -- 0待接单 1已接单 2到店 3取货 4配送中 5已送达 6已取消

    -- 取货地址
    pickup_address VARCHAR(500) NOT NULL DEFAULT '',
    pickup_lat DECIMAL(10,7) NOT NULL DEFAULT 0,
    pickup_lng DECIMAL(10,7) NOT NULL DEFAULT 0,

    -- 送达地址
    delivery_address VARCHAR(500) NOT NULL DEFAULT '',
    delivery_lat DECIMAL(10,7) NOT NULL DEFAULT 0,
    delivery_lng DECIMAL(10,7) NOT NULL DEFAULT 0,

    -- 距离与金额
    distance DECIMAL(10,2) NOT NULL DEFAULT 0,      -- 公里
    delivery_fee DECIMAL(12,2) NOT NULL DEFAULT 0,  -- 配送费
    tip DECIMAL(12,2) NOT NULL DEFAULT 0,           -- 小费

    -- 时间节点
    accepted_at TIMESTAMPTZ,                        -- 接单时间
    picked_at TIMESTAMPTZ,                          -- 取货时间
    delivered_at TIMESTAMPTZ,                       -- 送达时间

    -- 备注
    cancel_reason VARCHAR(500) NOT NULL DEFAULT ''
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_mall_deliveries_order_id ON mall_deliveries(order_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uniq_mall_deliveries_delivery_no ON mall_deliveries(delivery_no);
CREATE INDEX IF NOT EXISTS idx_mall_deliveries_rider_id ON mall_deliveries(rider_id);
CREATE INDEX IF NOT EXISTS idx_mall_deliveries_status ON mall_deliveries(status);
CREATE INDEX IF NOT EXISTS idx_mall_deliveries_shop_id ON mall_deliveries(shop_id);
CREATE INDEX IF NOT EXISTS idx_mall_deliveries_region_id ON mall_deliveries(region_id);
CREATE INDEX IF NOT EXISTS idx_mall_deliveries_user_id ON mall_deliveries(user_id);
CREATE INDEX IF NOT EXISTS idx_mall_deliveries_deleted_at ON mall_deliveries(deleted_at);

COMMENT ON TABLE mall_deliveries IS '同城商城配送单表';
COMMENT ON COLUMN mall_deliveries.order_id IS '关联 mall_orders 表订单 ID';
COMMENT ON COLUMN mall_deliveries.rider_id IS '认领骑手 ID';
COMMENT ON COLUMN mall_deliveries.shop_id IS '店铺 ID';
COMMENT ON COLUMN mall_deliveries.user_id IS '收货人用户 ID';
COMMENT ON COLUMN mall_deliveries.delivery_no IS '配送单号（业务唯一）';
COMMENT ON COLUMN mall_deliveries.status IS '状态：0待接单 1已接单 2到店 3取货 4配送中 5已送达 6已取消';
COMMENT ON COLUMN mall_deliveries.pickup_address IS '取货地址';
COMMENT ON COLUMN mall_deliveries.pickup_lat IS '取货纬度';
COMMENT ON COLUMN mall_deliveries.pickup_lng IS '取货经度';
COMMENT ON COLUMN mall_deliveries.delivery_address IS '送达地址';
COMMENT ON COLUMN mall_deliveries.delivery_lat IS '送达纬度';
COMMENT ON COLUMN mall_deliveries.delivery_lng IS '送达经度';
COMMENT ON COLUMN mall_deliveries.distance IS '配送距离（公里）';
COMMENT ON COLUMN mall_deliveries.delivery_fee IS '配送费';
COMMENT ON COLUMN mall_deliveries.tip IS '小费';
COMMENT ON COLUMN mall_deliveries.accepted_at IS '接单时间';
COMMENT ON COLUMN mall_deliveries.picked_at IS '取货时间';
COMMENT ON COLUMN mall_deliveries.delivered_at IS '送达时间';
COMMENT ON COLUMN mall_deliveries.cancel_reason IS '取消原因';

-- ============================================================
-- 3. mall_rider_settlements 骑手结算表
-- ============================================================
CREATE TABLE IF NOT EXISTS mall_rider_settlements (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    -- 关联
    rider_id BIGINT NOT NULL,                       -- 骑手 ID

    -- 结算周期
    period VARCHAR(20) NOT NULL DEFAULT '',         -- 结算周期 2026-07

    -- 统计
    total_orders INT NOT NULL DEFAULT 0,            -- 订单总数
    total_amount DECIMAL(12,2) NOT NULL DEFAULT 0,  -- 订单总额
    total_fee DECIMAL(12,2) NOT NULL DEFAULT 0,     -- 配送费总额
    total_tip DECIMAL(12,2) NOT NULL DEFAULT 0,     -- 小费总额
    platform_fee DECIMAL(12,2) NOT NULL DEFAULT 0,  -- 平台抽成
    net_amount DECIMAL(12,2) NOT NULL DEFAULT 0,    -- 实发金额

    -- 状态
    status SMALLINT NOT NULL DEFAULT 0,             -- 0待结算 1已结算 2已提现
    settled_at TIMESTAMPTZ,                         -- 结算时间
    withdrawn_at TIMESTAMPTZ,                       -- 提现时间

    -- 备注
    audit_reason VARCHAR(500) NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_mall_rider_settlements_rider_id ON mall_rider_settlements(rider_id);
CREATE INDEX IF NOT EXISTS idx_mall_rider_settlements_period ON mall_rider_settlements(period);
CREATE INDEX IF NOT EXISTS idx_mall_rider_settlements_status ON mall_rider_settlements(status);
CREATE INDEX IF NOT EXISTS idx_mall_rider_settlements_region_id ON mall_rider_settlements(region_id);
CREATE INDEX IF NOT EXISTS idx_mall_rider_settlements_deleted_at ON mall_rider_settlements(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_mall_rider_settlements_rider_period ON mall_rider_settlements(rider_id, period) WHERE deleted_at IS NULL;

COMMENT ON TABLE mall_rider_settlements IS '同城商城骑手结算表';
COMMENT ON COLUMN mall_rider_settlements.rider_id IS '骑手 ID';
COMMENT ON COLUMN mall_rider_settlements.period IS '结算周期（YYYY-MM）';
COMMENT ON COLUMN mall_rider_settlements.total_orders IS '周期内订单总数';
COMMENT ON COLUMN mall_rider_settlements.total_amount IS '周期内订单总额';
COMMENT ON COLUMN mall_rider_settlements.total_fee IS '周期内配送费总额';
COMMENT ON COLUMN mall_rider_settlements.total_tip IS '周期内小费总额';
COMMENT ON COLUMN mall_rider_settlements.platform_fee IS '平台抽成';
COMMENT ON COLUMN mall_rider_settlements.net_amount IS '实发金额 = total_fee + total_tip - platform_fee';
COMMENT ON COLUMN mall_rider_settlements.status IS '状态：0待结算 1已结算 2已提现';
COMMENT ON COLUMN mall_rider_settlements.settled_at IS '结算时间';
COMMENT ON COLUMN mall_rider_settlements.withdrawn_at IS '提现时间';
COMMENT ON COLUMN mall_rider_settlements.audit_reason IS '审核备注';

-- ============================================================
-- updated_at 触发器（依赖 001_p0_baseline.sql 中的 update_updated_at_column 函数）
-- PostgreSQL 不支持 CREATE TRIGGER IF NOT EXISTS，用 DROP TRIGGER IF EXISTS + CREATE TRIGGER
-- ============================================================
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'update_updated_at_column') THEN
        DROP TRIGGER IF EXISTS trg_mall_riders_updated_at ON mall_riders; CREATE TRIGGER trg_mall_riders_updated_at BEFORE UPDATE ON mall_riders FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_mall_deliveries_updated_at ON mall_deliveries; CREATE TRIGGER trg_mall_deliveries_updated_at BEFORE UPDATE ON mall_deliveries FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_mall_rider_settlements_updated_at ON mall_rider_settlements; CREATE TRIGGER trg_mall_rider_settlements_updated_at BEFORE UPDATE ON mall_rider_settlements FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;
