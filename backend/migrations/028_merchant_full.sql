-- =====================================================
-- merchant 商户中台模块完整迁移脚本
-- 包含 5 张表 + 索引 + 触发器 + COMMENT 注释
-- 依据架构设计 4.4：商家商户中台（对标美团/大众点评/有赞商户中台）
-- 数据库：wuchang_tongcheng
-- =====================================================
--
-- 内容：
--   1. CREATE 5 张表（merchant_ 前缀；merchant_shops 主表与 GORM AutoMigrate 共存，IF NOT EXISTS 保证幂等）
--   2. 索引（region_id/owner_id/shop_id/user_id/status/category_id/period 等）
--   3. updated_at 触发器（依赖 001_p0_baseline.sql 中的 update_updated_at_column 函数）
--   4. COMMENT 注释
--   5. 全幂等：CREATE TABLE IF NOT EXISTS / CREATE INDEX IF NOT EXISTS / DROP TRIGGER IF EXISTS
-- 依赖：001_p0_baseline.sql（update_updated_at_column 函数已创建）
--
-- 表清单（5 张）：
--   merchant_shops        店铺主表（RegionBaseModel，地区隔离）
--   merchant_staff        商户员工表（BaseModel，全局数据）
--   merchant_settles      商户结算表（BaseModel，DECIMAL(12,2) 金额）
--   merchant_categories   商户类目表（BaseModel，树形结构 parent_id 自引用）
--   merchant_verifications 商户认证表（RegionBaseModel，地区隔离）
--
-- 说明：表名严格遵循 model.TableName() 定义（均为复数）：
--      - Shop → merchant_shops
--      - Staff → merchant_staff
--      - Settle → merchant_settles
--      - Category → merchant_categories
--      - Verification → merchant_verifications
-- =====================================================

-- ============================================================
-- 1. merchant_shops 店铺主表（RegionBaseModel 地区隔离）
-- ============================================================
CREATE TABLE IF NOT EXISTS merchant_shops (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    -- 关联
    owner_id BIGINT NOT NULL,                       -- 店主用户 ID
    category_id BIGINT,                             -- 主营类目 ID（可空）

    -- 基本信息
    name VARCHAR(100) NOT NULL DEFAULT '',          -- 商户名
    logo VARCHAR(500) NOT NULL DEFAULT '',           -- 商户 LOGO URL
    intro TEXT,                                     -- 商户简介

    -- 状态
    status INT NOT NULL DEFAULT 0,                   -- 0审核中 1正常 2停用
    credit_score INT NOT NULL DEFAULT 100,           -- 信用分（初始 100）
    level INT NOT NULL DEFAULT 1,                    -- 商户等级（1-10）
    settled_at TIMESTAMPTZ                           -- 入驻时间（认领后写入）
);
CREATE INDEX IF NOT EXISTS idx_merchant_shops_region_id ON merchant_shops(region_id);
CREATE INDEX IF NOT EXISTS idx_merchant_shops_owner_id ON merchant_shops(owner_id);
CREATE INDEX IF NOT EXISTS idx_merchant_shops_category_id ON merchant_shops(category_id);
CREATE INDEX IF NOT EXISTS idx_merchant_shops_name ON merchant_shops(name);
CREATE INDEX IF NOT EXISTS idx_merchant_shops_status ON merchant_shops(status);
CREATE INDEX IF NOT EXISTS idx_merchant_shops_credit_score ON merchant_shops(credit_score);
CREATE INDEX IF NOT EXISTS idx_merchant_shops_level ON merchant_shops(level);
CREATE INDEX IF NOT EXISTS idx_merchant_shops_settled_at ON merchant_shops(settled_at);
CREATE INDEX IF NOT EXISTS idx_merchant_shops_deleted_at ON merchant_shops(deleted_at);

COMMENT ON TABLE merchant_shops IS '商户中台店铺主表（入驻/认领/状态/信用分/等级）';
COMMENT ON COLUMN merchant_shops.region_id IS '地区 ID，用于 4 维数据隔离';
COMMENT ON COLUMN merchant_shops.owner_id IS '店主用户 ID（认领后写入）';
COMMENT ON COLUMN merchant_shops.category_id IS '主营类目 ID（关联 merchant_categories，可空）';
COMMENT ON COLUMN merchant_shops.name IS '商户名称（最长 100 字符）';
COMMENT ON COLUMN merchant_shops.logo IS '商户 LOGO URL（最长 500 字符）';
COMMENT ON COLUMN merchant_shops.intro IS '商户简介（TEXT）';
COMMENT ON COLUMN merchant_shops.status IS '状态：0审核中 1正常 2停用';
COMMENT ON COLUMN merchant_shops.credit_score IS '信用分（初始 100，可由 M 端调整）';
COMMENT ON COLUMN merchant_shops.level IS '商户等级（1-10）';
COMMENT ON COLUMN merchant_shops.settled_at IS '入驻时间（认领成功后写入）';

-- ============================================================
-- 2. merchant_staff 商户员工表（BaseModel，全局数据）
-- ============================================================
CREATE TABLE IF NOT EXISTS merchant_staff (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    -- 关联
    shop_id BIGINT NOT NULL,                         -- 所属商户 ID
    user_id BIGINT NOT NULL,                         -- 关联用户 ID

    -- 角色
    role VARCHAR(20) NOT NULL DEFAULT 'clerk',       -- owner店主 / manager管理员 / clerk店员

    -- 权限（JSONB）
    permissions JSONB,                               -- 权限配置 JSON

    -- 状态
    status INT NOT NULL DEFAULT 1                    -- 1在职 2停用
);
CREATE INDEX IF NOT EXISTS idx_merchant_staff_shop_id ON merchant_staff(shop_id);
CREATE INDEX IF NOT EXISTS idx_merchant_staff_user_id ON merchant_staff(user_id);
CREATE INDEX IF NOT EXISTS idx_merchant_staff_role ON merchant_staff(role);
CREATE INDEX IF NOT EXISTS idx_merchant_staff_status ON merchant_staff(status);
CREATE INDEX IF NOT EXISTS idx_merchant_staff_deleted_at ON merchant_staff(deleted_at);
-- 店铺 + 用户 唯一约束（一个用户在同一店铺只能有一条员工记录）
CREATE UNIQUE INDEX IF NOT EXISTS uniq_merchant_staff_shop_user ON merchant_staff(shop_id, user_id) WHERE deleted_at IS NULL;

COMMENT ON TABLE merchant_staff IS '商户中台员工表（owner/manager/clerk，权限 JSONB 配置）';
COMMENT ON COLUMN merchant_staff.shop_id IS '所属商户 ID（关联 merchant_shops）';
COMMENT ON COLUMN merchant_staff.user_id IS '关联用户 ID';
COMMENT ON COLUMN merchant_staff.role IS '角色：owner店主 / manager管理员 / clerk店员';
COMMENT ON COLUMN merchant_staff.permissions IS '权限配置 JSONB（如 [{code,name,scope}]）';
COMMENT ON COLUMN merchant_staff.status IS '状态：1在职 2停用';

-- ============================================================
-- 3. merchant_settles 商户结算表（BaseModel，DECIMAL(12,2) 金额）
-- ============================================================
CREATE TABLE IF NOT EXISTS merchant_settles (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    -- 关联
    shop_id BIGINT NOT NULL,                         -- 所属商户 ID
    period VARCHAR(20) NOT NULL DEFAULT '',          -- 结算周期 YYYY-MM

    -- 金额（DECIMAL(12,2)）
    total_amount DECIMAL(12,2) NOT NULL DEFAULT 0,  -- 总金额
    platform_fee DECIMAL(12,2) NOT NULL DEFAULT 0,   -- 平台佣金
    shop_amount DECIMAL(12,2) NOT NULL DEFAULT 0,    -- 商户应得

    -- 状态
    status INT NOT NULL DEFAULT 0,                   -- 0待结算 1已结算 2已提现 3已撤销
    settled_at TIMESTAMPTZ                           -- 结算时间
);
CREATE INDEX IF NOT EXISTS idx_merchant_settles_shop_id ON merchant_settles(shop_id);
CREATE INDEX IF NOT EXISTS idx_merchant_settles_period ON merchant_settles(period);
CREATE INDEX IF NOT EXISTS idx_merchant_settles_status ON merchant_settles(status);
CREATE INDEX IF NOT EXISTS idx_merchant_settles_settled_at ON merchant_settles(settled_at);
CREATE INDEX IF NOT EXISTS idx_merchant_settles_deleted_at ON merchant_settles(deleted_at);
-- 商户 + 周期 唯一约束（同一商户同一周期只能有一张结算单）
CREATE UNIQUE INDEX IF NOT EXISTS uniq_merchant_settles_shop_period ON merchant_settles(shop_id, period) WHERE deleted_at IS NULL;

COMMENT ON TABLE merchant_settles IS '商户中台结算表（按月生成：总金额/平台佣金/商户应得）';
COMMENT ON COLUMN merchant_settles.shop_id IS '所属商户 ID（关联 merchant_shops）';
COMMENT ON COLUMN merchant_settles.period IS '结算周期 YYYY-MM（如 2026-07）';
COMMENT ON COLUMN merchant_settles.total_amount IS '总金额 DECIMAL(12,2)';
COMMENT ON COLUMN merchant_settles.platform_fee IS '平台佣金 DECIMAL(12,2)';
COMMENT ON COLUMN merchant_settles.shop_amount IS '商户应得 DECIMAL(12,2)';
COMMENT ON COLUMN merchant_settles.status IS '状态：0待结算 1已结算 2已提现 3已撤销';
COMMENT ON COLUMN merchant_settles.settled_at IS '结算时间';

-- ============================================================
-- 4. merchant_categories 商户类目表（BaseModel，parent_id 自引用树形）
-- ============================================================
CREATE TABLE IF NOT EXISTS merchant_categories (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    -- 树形结构
    parent_id BIGINT NOT NULL DEFAULT 0,             -- 父类目 ID（0=根）

    -- 基本信息
    name VARCHAR(64) NOT NULL DEFAULT '',           -- 类目名
    icon VARCHAR(255) NOT NULL DEFAULT '',           -- 图标 URL

    -- 显示与排序
    sort INT NOT NULL DEFAULT 0,                     -- 排序（越小越靠前）
    status INT NOT NULL DEFAULT 1                    -- 0禁用 1启用
);
CREATE INDEX IF NOT EXISTS idx_merchant_categories_parent_id ON merchant_categories(parent_id);
CREATE INDEX IF NOT EXISTS idx_merchant_categories_name ON merchant_categories(name);
CREATE INDEX IF NOT EXISTS idx_merchant_categories_sort ON merchant_categories(sort);
CREATE INDEX IF NOT EXISTS idx_merchant_categories_status ON merchant_categories(status);
CREATE INDEX IF NOT EXISTS idx_merchant_categories_deleted_at ON merchant_categories(deleted_at);

COMMENT ON TABLE merchant_categories IS '商户中台类目表（树形结构 parent_id 自引用，全局数据）';
COMMENT ON COLUMN merchant_categories.parent_id IS '父类目 ID（0=根类目）';
COMMENT ON COLUMN merchant_categories.name IS '类目名（最长 64 字符）';
COMMENT ON COLUMN merchant_categories.icon IS '图标 URL（最长 255 字符）';
COMMENT ON COLUMN merchant_categories.sort IS '排序值（越小越靠前）';
COMMENT ON COLUMN merchant_categories.status IS '状态：0禁用 1启用';

-- ============================================================
-- 5. merchant_verifications 商户认证表（RegionBaseModel 地区隔离）
-- ============================================================
CREATE TABLE IF NOT EXISTS merchant_verifications (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    -- 关联
    shop_id BIGINT NOT NULL,                         -- 所属商户 ID

    -- 认证类型
    type VARCHAR(20) NOT NULL DEFAULT 'business',    -- business企业认证 / personal个人认证

    -- 营业执照信息
    license_no VARCHAR(64) NOT NULL DEFAULT '',      -- 营业执照号
    license_image VARCHAR(255) NOT NULL DEFAULT '',  -- 营业执照图片 URL
    legal_person VARCHAR(64) NOT NULL DEFAULT '',    -- 法人代表
    legal_person_id VARCHAR(32) NOT NULL DEFAULT '',  -- 法人身份证号

    -- 状态
    status INT NOT NULL DEFAULT 0,                   -- 0待审 1通过 2拒绝
    audit_reason VARCHAR(500) NOT NULL DEFAULT '',    -- 审核备注
    audited_at TIMESTAMPTZ                           -- 审核时间
);
CREATE INDEX IF NOT EXISTS idx_merchant_verifications_region_id ON merchant_verifications(region_id);
CREATE INDEX IF NOT EXISTS idx_merchant_verifications_shop_id ON merchant_verifications(shop_id);
CREATE INDEX IF NOT EXISTS idx_merchant_verifications_type ON merchant_verifications(type);
CREATE INDEX IF NOT EXISTS idx_merchant_verifications_license_no ON merchant_verifications(license_no);
CREATE INDEX IF NOT EXISTS idx_merchant_verifications_status ON merchant_verifications(status);
CREATE INDEX IF NOT EXISTS idx_merchant_verifications_audited_at ON merchant_verifications(audited_at);
CREATE INDEX IF NOT EXISTS idx_merchant_verifications_deleted_at ON merchant_verifications(deleted_at);

COMMENT ON TABLE merchant_verifications IS '商户中台认证表（企业认证/个人认证）';
COMMENT ON COLUMN merchant_verifications.region_id IS '地区 ID，用于 4 维数据隔离';
COMMENT ON COLUMN merchant_verifications.shop_id IS '所属商户 ID（关联 merchant_shops）';
COMMENT ON COLUMN merchant_verifications.type IS '认证类型：business企业 / personal个人';
COMMENT ON COLUMN merchant_verifications.license_no IS '营业执照号（最长 64 字符）';
COMMENT ON COLUMN merchant_verifications.license_image IS '营业执照图片 URL';
COMMENT ON COLUMN merchant_verifications.legal_person IS '法人代表（最长 64 字符）';
COMMENT ON COLUMN merchant_verifications.legal_person_id IS '法人身份证号（最长 32 字符）';
COMMENT ON COLUMN merchant_verifications.status IS '状态：0待审 1通过 2拒绝';
COMMENT ON COLUMN merchant_verifications.audit_reason IS '审核备注（最长 500 字符）';
COMMENT ON COLUMN merchant_verifications.audited_at IS '审核时间';

-- ============================================================
-- updated_at 触发器（依赖 001_p0_baseline.sql 中的 update_updated_at_column 函数）
-- PostgreSQL 不支持 CREATE TRIGGER IF NOT EXISTS，用 DROP TRIGGER IF EXISTS + CREATE TRIGGER
-- ============================================================
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'update_updated_at_column') THEN
        DROP TRIGGER IF EXISTS trg_merchant_shops_updated_at ON merchant_shops; CREATE TRIGGER trg_merchant_shops_updated_at BEFORE UPDATE ON merchant_shops FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_merchant_staff_updated_at ON merchant_staff; CREATE TRIGGER trg_merchant_staff_updated_at BEFORE UPDATE ON merchant_staff FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_merchant_settles_updated_at ON merchant_settles; CREATE TRIGGER trg_merchant_settles_updated_at BEFORE UPDATE ON merchant_settles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_merchant_categories_updated_at ON merchant_categories; CREATE TRIGGER trg_merchant_categories_updated_at BEFORE UPDATE ON merchant_categories FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_merchant_verifications_updated_at ON merchant_verifications; CREATE TRIGGER trg_merchant_verifications_updated_at BEFORE UPDATE ON merchant_verifications FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;

-- =====================================================
-- 迁移完成：merchant 商户中台 5 张表 + 索引 + 触发器 + COMMENT
-- =====================================================
