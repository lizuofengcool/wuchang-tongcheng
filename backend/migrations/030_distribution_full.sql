-- =====================================================
-- distribution 分销合伙人中台模块完整迁移脚本
-- 包含 5 张表 + 索引 + 触发器 + 注释
-- 依据架构设计 4.5：分销合伙人中台（distribution）
-- 职责：二级分销/城市分站分成/推广渠道统计/佣金自动结算/付费合伙人等级
-- 数据库：wuchang_tongcheng
-- =====================================================
--
-- 内容：
--   1. CREATE 5 张表（distribution_ 前缀；distribution_partners 主表由 GORM AutoMigrate 创建）
--   2. 索引、触发器、注释
--   3. 全幂等：CREATE TABLE IF NOT EXISTS
-- 依赖：001_p0_baseline.sql（update_updated_at_column 函数已创建）
--
-- 表清单（5 张表，主表 distribution_partners 由 GORM AutoMigrate 创建）：
--   distribution_partners / distribution_channels / distribution_commissions /
--   distribution_levels / distribution_withdrawals
--
-- 说明：表名严格遵循 model.TableName() 定义：
--      - Partner → distribution_partners（复数）
--      - Channel → distribution_channels（复数）
--      - Commission → distribution_commissions（复数）
--      - Level → distribution_levels（复数）
--      - Withdrawal → distribution_withdrawals（复数）
-- =====================================================

-- ============================================================
-- 1. distribution_partners 分销合伙人主表
-- ============================================================
CREATE TABLE IF NOT EXISTS distribution_partners (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    -- 关联
    user_id BIGINT NOT NULL,                         -- 关联用户 ID
    parent_id BIGINT NOT NULL DEFAULT 0,             -- 上级合伙人 ID（0=无上级）

    -- 等级与佣金
    level INT NOT NULL DEFAULT 1,                    -- 1普通 2高级 3城市合伙人
    commission_rate DECIMAL(5,4) NOT NULL DEFAULT 0, -- 佣金比例 0-1
    total_commission DECIMAL(12,2) NOT NULL DEFAULT 0,     -- 累计佣金
    settled_commission DECIMAL(12,2) NOT NULL DEFAULT 0,   -- 已结算佣金
    pending_commission DECIMAL(12,2) NOT NULL DEFAULT 0,   -- 待结算佣金

    -- 状态
    status INT NOT NULL DEFAULT 0,                   -- 0待审核 1正常 2冻结 3拒绝 4退出
    joined_at TIMESTAMPTZ                            -- 加入时间
);
CREATE INDEX IF NOT EXISTS idx_distribution_partners_user_id ON distribution_partners(user_id);
CREATE INDEX IF NOT EXISTS idx_distribution_partners_parent_id ON distribution_partners(parent_id);
CREATE INDEX IF NOT EXISTS idx_distribution_partners_region_id ON distribution_partners(region_id);
CREATE INDEX IF NOT EXISTS idx_distribution_partners_level ON distribution_partners(level);
CREATE INDEX IF NOT EXISTS idx_distribution_partners_status ON distribution_partners(status);
CREATE INDEX IF NOT EXISTS idx_distribution_partners_joined_at ON distribution_partners(joined_at);
CREATE INDEX IF NOT EXISTS idx_distribution_partners_deleted_at ON distribution_partners(deleted_at);

COMMENT ON TABLE distribution_partners IS '分销合伙人主表';
COMMENT ON COLUMN distribution_partners.user_id IS '关联用户 ID';
COMMENT ON COLUMN distribution_partners.parent_id IS '上级合伙人 ID（0=无上级）';
COMMENT ON COLUMN distribution_partners.level IS '等级：1普通 2高级 3城市合伙人';
COMMENT ON COLUMN distribution_partners.commission_rate IS '佣金比例 0-1（DECIMAL(5,4)）';
COMMENT ON COLUMN distribution_partners.total_commission IS '累计佣金（DECIMAL(12,2)）';
COMMENT ON COLUMN distribution_partners.settled_commission IS '已结算佣金';
COMMENT ON COLUMN distribution_partners.pending_commission IS '待结算佣金';
COMMENT ON COLUMN distribution_partners.status IS '状态：0待审核 1正常 2冻结 3拒绝 4退出';
COMMENT ON COLUMN distribution_partners.joined_at IS '加入时间（审核通过时设置）';

-- ============================================================
-- 2. distribution_channels 推广渠道表
-- ============================================================
CREATE TABLE IF NOT EXISTS distribution_channels (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    -- 关联
    partner_id BIGINT NOT NULL,                      -- 所属合伙人 ID
    code VARCHAR(50) NOT NULL,                       -- 推广码（唯一）
    name VARCHAR(100) NOT NULL DEFAULT '',           -- 渠道名称

    -- 统计
    click_count INT NOT NULL DEFAULT 0,              -- 点击数
    register_count INT NOT NULL DEFAULT 0,           -- 注册数
    order_count INT NOT NULL DEFAULT 0,              -- 订单数
    commission_amount DECIMAL(12,2) NOT NULL DEFAULT 0  -- 累计佣金
);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_distribution_channels_code ON distribution_channels(code);
CREATE INDEX IF NOT EXISTS idx_distribution_channels_partner_id ON distribution_channels(partner_id);
CREATE INDEX IF NOT EXISTS idx_distribution_channels_code ON distribution_channels(code);
CREATE INDEX IF NOT EXISTS idx_distribution_channels_deleted_at ON distribution_channels(deleted_at);

COMMENT ON TABLE distribution_channels IS '分销推广渠道表';
COMMENT ON COLUMN distribution_channels.partner_id IS '所属合伙人 ID';
COMMENT ON COLUMN distribution_channels.code IS '推广码（唯一）';
COMMENT ON COLUMN distribution_channels.click_count IS '点击数';
COMMENT ON COLUMN distribution_channels.register_count IS '注册数';
COMMENT ON COLUMN distribution_channels.order_count IS '订单数';
COMMENT ON COLUMN distribution_channels.commission_amount IS '累计佣金（DECIMAL(12,2)）';

-- ============================================================
-- 3. distribution_commissions 佣金记录表
-- ============================================================
CREATE TABLE IF NOT EXISTS distribution_commissions (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    -- 关联
    partner_id BIGINT NOT NULL,                      -- 合伙人 ID
    order_id BIGINT NOT NULL,                        -- 订单 ID
    channel_id BIGINT NOT NULL DEFAULT 0,            -- 渠道 ID（0=无渠道）

    -- 金额
    order_amount DECIMAL(12,2) NOT NULL DEFAULT 0,   -- 订单金额
    commission_amount DECIMAL(12,2) NOT NULL DEFAULT 0,  -- 佣金金额
    commission_rate DECIMAL(5,4) NOT NULL DEFAULT 0,     -- 佣金比例（快照）

    -- 级别与状态
    level INT NOT NULL DEFAULT 1,                    -- 1一级 2二级
    status INT NOT NULL DEFAULT 0,                   -- 0待结算 1已结算 2已取消
    settled_at TIMESTAMPTZ                           -- 结算时间
);
CREATE INDEX IF NOT EXISTS idx_distribution_commissions_partner_id ON distribution_commissions(partner_id);
CREATE INDEX IF NOT EXISTS idx_distribution_commissions_order_id ON distribution_commissions(order_id);
CREATE INDEX IF NOT EXISTS idx_distribution_commissions_channel_id ON distribution_commissions(channel_id);
CREATE INDEX IF NOT EXISTS idx_distribution_commissions_level ON distribution_commissions(level);
CREATE INDEX IF NOT EXISTS idx_distribution_commissions_status ON distribution_commissions(status);
CREATE INDEX IF NOT EXISTS idx_distribution_commissions_settled_at ON distribution_commissions(settled_at);
CREATE INDEX IF NOT EXISTS idx_distribution_commissions_deleted_at ON distribution_commissions(deleted_at);

COMMENT ON TABLE distribution_commissions IS '分销佣金记录表';
COMMENT ON COLUMN distribution_commissions.partner_id IS '合伙人 ID';
COMMENT ON COLUMN distribution_commissions.order_id IS '订单 ID';
COMMENT ON COLUMN distribution_commissions.channel_id IS '渠道 ID（0=无渠道）';
COMMENT ON COLUMN distribution_commissions.order_amount IS '订单金额（DECIMAL(12,2)）';
COMMENT ON COLUMN distribution_commissions.commission_amount IS '佣金金额（DECIMAL(12,2)）';
COMMENT ON COLUMN distribution_commissions.commission_rate IS '佣金比例快照（DECIMAL(5,4)）';
COMMENT ON COLUMN distribution_commissions.level IS '级别：1一级分销 2二级分销';
COMMENT ON COLUMN distribution_commissions.status IS '状态：0待结算 1已结算 2已取消';
COMMENT ON COLUMN distribution_commissions.settled_at IS '结算时间';

-- ============================================================
-- 4. distribution_levels 合伙人等级表
-- ============================================================
CREATE TABLE IF NOT EXISTS distribution_levels (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    -- 等级定义
    level INT NOT NULL,                              -- 等级值 1/2/3
    name VARCHAR(64) NOT NULL DEFAULT '',            -- 等级名称
    required_amount DECIMAL(12,2) NOT NULL DEFAULT 0, -- 升级所需累计佣金
    commission_rate DECIMAL(5,4) NOT NULL DEFAULT 0,  -- 默认佣金比例

    -- 扩展
    extra_benefits JSONB,                            -- 额外权益 JSONB
    status INT NOT NULL DEFAULT 1                    -- 0禁用 1启用
);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_distribution_levels_level ON distribution_levels(level);
CREATE INDEX IF NOT EXISTS idx_distribution_levels_status ON distribution_levels(status);
CREATE INDEX IF NOT EXISTS idx_distribution_levels_deleted_at ON distribution_levels(deleted_at);

COMMENT ON TABLE distribution_levels IS '分销合伙人等级表';
COMMENT ON COLUMN distribution_levels.level IS '等级值 1/2/3（唯一）';
COMMENT ON COLUMN distribution_levels.required_amount IS '升级所需累计佣金（DECIMAL(12,2)）';
COMMENT ON COLUMN distribution_levels.commission_rate IS '默认佣金比例（DECIMAL(5,4)）';
COMMENT ON COLUMN distribution_levels.extra_benefits IS '额外权益 JSONB';
COMMENT ON COLUMN distribution_levels.status IS '状态：0禁用 1启用';

-- ============================================================
-- 5. distribution_withdrawals 提现记录表
-- ============================================================
CREATE TABLE IF NOT EXISTS distribution_withdrawals (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    -- 关联
    partner_id BIGINT NOT NULL,                      -- 合伙人 ID

    -- 金额
    amount DECIMAL(12,2) NOT NULL DEFAULT 0,         -- 提现金额

    -- 状态
    status INT NOT NULL DEFAULT 0,                   -- 0申请中 1已审核 2已打款 3已拒绝
    bank_info JSONB,                                 -- 银行/账户信息 JSONB
    audit_reason VARCHAR(500) NOT NULL DEFAULT '',   -- 审核备注
    audited_at TIMESTAMPTZ,                          -- 审核时间
    paid_at TIMESTAMPTZ                              -- 打款时间
);
CREATE INDEX IF NOT EXISTS idx_distribution_withdrawals_partner_id ON distribution_withdrawals(partner_id);
CREATE INDEX IF NOT EXISTS idx_distribution_withdrawals_status ON distribution_withdrawals(status);
CREATE INDEX IF NOT EXISTS idx_distribution_withdrawals_audited_at ON distribution_withdrawals(audited_at);
CREATE INDEX IF NOT EXISTS idx_distribution_withdrawals_paid_at ON distribution_withdrawals(paid_at);
CREATE INDEX IF NOT EXISTS idx_distribution_withdrawals_deleted_at ON distribution_withdrawals(deleted_at);

COMMENT ON TABLE distribution_withdrawals IS '分销提现记录表';
COMMENT ON COLUMN distribution_withdrawals.partner_id IS '合伙人 ID';
COMMENT ON COLUMN distribution_withdrawals.amount IS '提现金额（DECIMAL(12,2)）';
COMMENT ON COLUMN distribution_withdrawals.status IS '状态：0申请中 1已审核 2已打款 3已拒绝';
COMMENT ON COLUMN distribution_withdrawals.bank_info IS '银行/账户信息 JSONB';
COMMENT ON COLUMN distribution_withdrawals.audit_reason IS '审核备注';
COMMENT ON COLUMN distribution_withdrawals.audited_at IS '审核时间';
COMMENT ON COLUMN distribution_withdrawals.paid_at IS '打款时间';

-- ============================================================
-- updated_at 触发器（依赖 001_p0_baseline.sql 中的 update_updated_at_column 函数）
-- PostgreSQL 不支持 CREATE TRIGGER IF NOT EXISTS，用 DROP TRIGGER IF EXISTS + CREATE TRIGGER
-- ============================================================
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'update_updated_at_column') THEN
        DROP TRIGGER IF EXISTS trg_distribution_partners_updated_at ON distribution_partners; CREATE TRIGGER trg_distribution_partners_updated_at BEFORE UPDATE ON distribution_partners FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_distribution_channels_updated_at ON distribution_channels; CREATE TRIGGER trg_distribution_channels_updated_at BEFORE UPDATE ON distribution_channels FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_distribution_commissions_updated_at ON distribution_commissions; CREATE TRIGGER trg_distribution_commissions_updated_at BEFORE UPDATE ON distribution_commissions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_distribution_levels_updated_at ON distribution_levels; CREATE TRIGGER trg_distribution_levels_updated_at BEFORE UPDATE ON distribution_levels FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_distribution_withdrawals_updated_at ON distribution_withdrawals; CREATE TRIGGER trg_distribution_withdrawals_updated_at BEFORE UPDATE ON distribution_withdrawals FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;
