-- ============================================================
-- 012_pay_full.sql 支付/担保交易中台扩展表
-- 对标支付宝/微信支付/银联/Stripe
-- 创建 pay_transactions / pay_channels / pay_merchants / pay_callbacks
-- 现有的 pay_orders / pay_escrows / pay_refunds / pay_settlements / pay_accounts / pay_methods
-- 已在 005_p1_middlewares.sql 创建，此处仅创建扩展表
-- 全部幂等：CREATE TABLE IF NOT EXISTS
-- ============================================================

-- ============================================================
-- 1. pay_transactions 交易流水记录
-- 对标 Stripe charge / 支付宝 trade_record
-- ============================================================
CREATE TABLE IF NOT EXISTS pay_transactions (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    transaction_no VARCHAR(64) NOT NULL UNIQUE,                  -- 平台流水号
    order_id BIGINT NOT NULL DEFAULT 0,                           -- 关联订单ID
    order_no VARCHAR(64) NOT NULL DEFAULT '',                     -- 关联订单号
    user_id BIGINT NOT NULL,                                       -- 用户ID
    channel VARCHAR(32) NOT NULL DEFAULT '',                      -- 渠道：wechat/alipay/unionpay/stripe
    third_party_no VARCHAR(128) NOT NULL DEFAULT '',              -- 三方交易号
    amount DECIMAL(12,2) NOT NULL DEFAULT 0.00,                    -- 金额
    fee DECIMAL(12,2) NOT NULL DEFAULT 0.00,                      -- 手续费
    status SMALLINT NOT NULL DEFAULT 0,                           -- 0处理中 1成功 2失败 3撤销
    channel_resp JSONB NOT NULL DEFAULT '{}'::jsonb,              -- 渠道响应原文
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_pay_transactions_order_id ON pay_transactions(order_id);
CREATE INDEX IF NOT EXISTS idx_pay_transactions_user_id ON pay_transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_pay_transactions_channel ON pay_transactions(channel);
CREATE INDEX IF NOT EXISTS idx_pay_transactions_status ON pay_transactions(status);
CREATE INDEX IF NOT EXISTS idx_pay_transactions_third_party_no ON pay_transactions(third_party_no);
CREATE INDEX IF NOT EXISTS idx_pay_transactions_region_id ON pay_transactions(region_id);
COMMENT ON TABLE pay_transactions IS '支付交易流水记录';
COMMENT ON COLUMN pay_transactions.status IS '0处理中 1成功 2失败 3撤销';

-- ============================================================
-- 2. pay_channels 支付渠道配置
-- 对标 Stripe payment_method_configs
-- ============================================================
CREATE TABLE IF NOT EXISTS pay_channels (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    channel_code VARCHAR(32) NOT NULL,                            -- 渠道码 wechat/alipay/unionpay/stripe
    channel_name VARCHAR(64) NOT NULL DEFAULT '',                  -- 渠道名
    merchant_no VARCHAR(64) NOT NULL DEFAULT '',                   -- 商户号
    app_id VARCHAR(128) NOT NULL DEFAULT '',                        -- 应用ID
    secret_key VARCHAR(512) NOT NULL DEFAULT '',                   -- 密钥（加密存储）
    public_key TEXT NOT NULL DEFAULT '',                            -- 公钥
    private_key TEXT NOT NULL DEFAULT '',                            -- 私钥
    callback_url VARCHAR(256) NOT NULL DEFAULT '',                  -- 回调URL
    notify_url VARCHAR(256) NOT NULL DEFAULT '',                     -- 异步通知URL
    fee_rate DECIMAL(6,4) NOT NULL DEFAULT 0.0000,                  -- 费率
    fee_fixed DECIMAL(12,2) NOT NULL DEFAULT 0.00,                  -- 固定手续费
    config JSONB NOT NULL DEFAULT '{}'::jsonb,                       -- 额外配置
    sort INTEGER NOT NULL DEFAULT 0,
    status SMALLINT NOT NULL DEFAULT 1,                             -- 1启用 0禁用
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_pay_channels_code ON pay_channels(channel_code);
CREATE INDEX IF NOT EXISTS idx_pay_channels_status ON pay_channels(status);
CREATE INDEX IF NOT EXISTS idx_pay_channels_region_id ON pay_channels(region_id);
COMMENT ON TABLE pay_channels IS '支付渠道配置';

-- ============================================================
-- 3. pay_merchants 商户配置
-- ============================================================
CREATE TABLE IF NOT EXISTS pay_merchants (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    merchant_no VARCHAR(64) NOT NULL UNIQUE,                        -- 商户号
    merchant_name VARCHAR(128) NOT NULL DEFAULT '',                  -- 商户名
    user_id BIGINT NOT NULL DEFAULT 0,                               -- 关联用户ID
    contact_name VARCHAR(64) NOT NULL DEFAULT '',                    -- 联系人
    contact_phone VARCHAR(32) NOT NULL DEFAULT '',                    -- 联系电话
    fee_rate DECIMAL(6,4) NOT NULL DEFAULT 0.0000,                  -- 商户费率
    settlement_cycle VARCHAR(16) NOT NULL DEFAULT 'T1',             -- 结算周期 T1/T7/monthly
    business_license VARCHAR(64) NOT NULL DEFAULT '',               -- 营业执照号
    business_scope TEXT NOT NULL DEFAULT '',                         -- 经营范围
    bank_account JSONB NOT NULL DEFAULT '{}'::jsonb,                -- 银行账户信息
    status SMALLINT NOT NULL DEFAULT 0,                             -- 0待审 1通过 2拒绝 3冻结
    audit_remark VARCHAR(256) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_pay_merchants_user_id ON pay_merchants(user_id);
CREATE INDEX IF NOT EXISTS idx_pay_merchants_status ON pay_merchants(status);
CREATE INDEX IF NOT EXISTS idx_pay_merchants_region_id ON pay_merchants(region_id);
COMMENT ON TABLE pay_merchants IS '商户配置';

-- ============================================================
-- 4. pay_callbacks 三方回调通知记录
-- ============================================================
CREATE TABLE IF NOT EXISTS pay_callbacks (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    order_no VARCHAR(64) NOT NULL DEFAULT '',                       -- 订单号
    channel VARCHAR(32) NOT NULL DEFAULT '',                         -- 渠道
    third_party_no VARCHAR(128) NOT NULL DEFAULT '',                 -- 三方交易号
    notify_type VARCHAR(32) NOT NULL DEFAULT '',                    -- 通知类型 pay/refund/escrow
    raw_data TEXT NOT NULL DEFAULT '',                                -- 原始数据
    parsed_data JSONB NOT NULL DEFAULT '{}'::jsonb,                  -- 解析后数据
    signature VARCHAR(512) NOT NULL DEFAULT '',                       -- 签名
    status SMALLINT NOT NULL DEFAULT 0,                              -- 0待处理 1已处理 2失败
    process_count INTEGER NOT NULL DEFAULT 0,                        -- 处理次数
    error_msg TEXT NOT NULL DEFAULT '',                               -- 错误信息
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_pay_callbacks_order_no ON pay_callbacks(order_no);
CREATE INDEX IF NOT EXISTS idx_pay_callbacks_channel ON pay_callbacks(channel);
CREATE INDEX IF NOT EXISTS idx_pay_callbacks_status ON pay_callbacks(status);
CREATE INDEX IF NOT EXISTS idx_pay_callbacks_region_id ON pay_callbacks(region_id);
COMMENT ON TABLE pay_callbacks IS '三方支付回调通知记录';

-- ============================================================
-- 5. 扩展 pay_escrows 担保交易：新增仲裁相关列
-- 注：原有列已由 005 创建，此处仅添加扩展列（IF NOT EXISTS 不可用列，使用 DO 块）
-- ============================================================
DO $$
BEGIN
    -- 仲裁人ID
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'pay_escrows' AND column_name = 'arbitrator_id') THEN
        ALTER TABLE pay_escrows ADD COLUMN arbitrator_id BIGINT NOT NULL DEFAULT 0;
    END IF;
    -- 仲裁备注
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'pay_escrows' AND column_name = 'arbitration_remark') THEN
        ALTER TABLE pay_escrows ADD COLUMN arbitration_remark TEXT NOT NULL DEFAULT '';
    END IF;
    -- 仲裁时间
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'pay_escrows' AND column_name = 'arbitrated_at') THEN
        ALTER TABLE pay_escrows ADD COLUMN arbitrated_at TIMESTAMPTZ;
    END IF;
    -- 争议状态：0无 1争议中 2已仲裁
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'pay_escrows' AND column_name = 'dispute_status') THEN
        ALTER TABLE pay_escrows ADD COLUMN dispute_status SMALLINT NOT NULL DEFAULT 0;
    END IF;
    -- 争议原因
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'pay_escrows' AND column_name = 'dispute_reason') THEN
        ALTER TABLE pay_escrows ADD COLUMN dispute_reason VARCHAR(512) NOT NULL DEFAULT '';
    END IF;
END$$;

CREATE INDEX IF NOT EXISTS idx_pay_escrows_dispute_status ON pay_escrows(dispute_status);
COMMENT ON COLUMN pay_escrows.dispute_status IS '0无争议 1争议中 2已仲裁';
