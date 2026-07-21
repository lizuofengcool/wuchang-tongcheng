-- =====================================================
-- diy DIY 前端页面中台模块完整迁移脚本
-- 包含 4 张表 + 索引 + 触发器 + 注释
-- 依据架构设计 4.12：拖拽生成首页/专题页/店铺页/活动页
-- 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
-- 数据库：wuchang_tongcheng
-- =====================================================
--
-- 内容：
--   1. CREATE 4 张表（diy 模块，diy_ 前缀，表名严格遵循 model.TableName()）
--   2. 索引、触发器、注释
--   3. 全幂等：CREATE TABLE IF NOT EXISTS + DROP TRIGGER IF EXISTS + CREATE TRIGGER
-- 依赖：001_p0_baseline.sql（update_updated_at_column 函数已创建）
--
-- 表清单（4 张）：
--   diy_pages       DIY 页面（RegionBaseModel，地区隔离 + user_id 用户隔离）
--   diy_components  DIY 组件库（BaseModel，全局共享）
--   diy_templates   DIY 页面模板（BaseModel，全局共享）
--   diy_page_stats  DIY 页面统计（BaseModel，按日期+page_id 汇总）
--
-- 说明：表名严格遵循 model.TableName() 定义：
--      - Page      → diy_pages（复数）
--      - Component → diy_components（复数）
--      - Template  → diy_templates（复数）
--      - PageStat  → diy_page_stats（复数）
-- =====================================================

-- ============================================================
-- 1. diy_pages DIY 页面表（地区隔离 + 用户隔离）
-- ============================================================
CREATE TABLE IF NOT EXISTS diy_pages (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    -- 基础信息
    title VARCHAR(100) NOT NULL DEFAULT '',                  -- 页面标题
    type VARCHAR(32) NOT NULL DEFAULT 'home',                -- 类型：home首页/topic专题页/shop店铺页/activity活动页
    slug VARCHAR(100) NOT NULL DEFAULT '',                   -- URL Slug（按 slug 获取已发布页面）
    status INT NOT NULL DEFAULT 0,                           -- 状态：0草稿 1已发布 2已下线
    user_id BIGINT NOT NULL DEFAULT 0,                       -- 创建者 ID（用户隔离）
    biz_id BIGINT NOT NULL DEFAULT 0,                        -- 业务 ID（如店铺 ID/活动 ID，0 表示无关联）

    -- 时间
    published_at TIMESTAMPTZ,                                -- 发布时间

    -- JSONB 配置
    components JSONB,                                        -- 组件配置（拖拽组件数组）
    settings JSONB                                           -- 页面设置（如 SEO/背景/全局样式）
);
CREATE INDEX IF NOT EXISTS idx_diy_pages_region_id ON diy_pages(region_id);
CREATE INDEX IF NOT EXISTS idx_diy_pages_user_id ON diy_pages(user_id);
CREATE INDEX IF NOT EXISTS idx_diy_pages_title ON diy_pages(title);
CREATE INDEX IF NOT EXISTS idx_diy_pages_type ON diy_pages(type);
CREATE INDEX IF NOT EXISTS idx_diy_pages_slug ON diy_pages(slug);
CREATE INDEX IF NOT EXISTS idx_diy_pages_status ON diy_pages(status);
CREATE INDEX IF NOT EXISTS idx_diy_pages_biz_id ON diy_pages(biz_id);
CREATE INDEX IF NOT EXISTS idx_diy_pages_published_at ON diy_pages(published_at);
CREATE INDEX IF NOT EXISTS idx_diy_pages_deleted_at ON diy_pages(deleted_at);

COMMENT ON TABLE diy_pages IS 'DIY 页面表（可视化装修：首页/专题页/店铺页/活动页）';
COMMENT ON COLUMN diy_pages.type IS '类型：home首页/topic专题页/shop店铺页/activity活动页';
COMMENT ON COLUMN diy_pages.slug IS 'URL Slug（按 slug 获取已发布页面，同 region 内已发布页面 slug 唯一）';
COMMENT ON COLUMN diy_pages.status IS '状态：0草稿 1已发布 2已下线';
COMMENT ON COLUMN diy_pages.user_id IS '创建者 ID（用户隔离，0 表示系统创建）';
COMMENT ON COLUMN diy_pages.biz_id IS '业务 ID（如店铺 ID/活动 ID，0 表示无关联）';
COMMENT ON COLUMN diy_pages.components IS '组件配置 JSONB（拖拽组件数组，含组件 code/props/排序等）';
COMMENT ON COLUMN diy_pages.settings IS '页面设置 JSONB（如 SEO/背景/全局样式）';

-- ============================================================
-- 2. diy_components DIY 组件库表（全局共享）
-- ============================================================
CREATE TABLE IF NOT EXISTS diy_components (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    -- 基础信息
    name VARCHAR(64) NOT NULL DEFAULT '',                    -- 组件名称
    code VARCHAR(64) NOT NULL DEFAULT '',                    -- 组件编码（唯一）
    category VARCHAR(32) NOT NULL DEFAULT 'basic',           -- 分类：basic基础/layout布局/business业务
    description TEXT,                                        -- 组件描述
    thumbnail VARCHAR(500) NOT NULL DEFAULT '',              -- 缩略图 URL
    status INT NOT NULL DEFAULT 1,                           -- 状态：0禁用 1启用

    -- JSONB 配置
    config JSONB                                             -- 组件配置模板（默认属性 schema）
);
-- 唯一索引：组件编码唯一（仅未删除记录）
CREATE UNIQUE INDEX IF NOT EXISTS uniq_diy_components_code ON diy_components(code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_diy_components_name ON diy_components(name);
CREATE INDEX IF NOT EXISTS idx_diy_components_category ON diy_components(category);
CREATE INDEX IF NOT EXISTS idx_diy_components_status ON diy_components(status);
CREATE INDEX IF NOT EXISTS idx_diy_components_deleted_at ON diy_components(deleted_at);

COMMENT ON TABLE diy_components IS 'DIY 组件库表（基础组件/布局组件/业务组件）';
COMMENT ON COLUMN diy_components.code IS '组件编码（唯一，未删除记录范围内唯一）';
COMMENT ON COLUMN diy_components.category IS '分类：basic基础组件/layout布局组件/business业务组件';
COMMENT ON COLUMN diy_components.status IS '状态：0禁用 1启用';
COMMENT ON COLUMN diy_components.config IS '组件配置模板 JSONB（默认属性 schema，编辑器初始化使用）';

-- ============================================================
-- 3. diy_templates DIY 页面模板表（全局共享）
-- ============================================================
CREATE TABLE IF NOT EXISTS diy_templates (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    -- 基础信息
    name VARCHAR(100) NOT NULL DEFAULT '',                   -- 模板名称
    thumbnail VARCHAR(500) NOT NULL DEFAULT '',              -- 缩略图 URL
    description TEXT,                                        -- 模板描述
    category VARCHAR(32) NOT NULL DEFAULT 'home',            -- 分类：home首页/topic专题页/shop店铺页/activity活动页
    status INT NOT NULL DEFAULT 1,                           -- 状态：0禁用 1启用

    -- JSONB 配置
    pages JSONB                                              -- 模板页面配置（包含 components/settings）
);
CREATE INDEX IF NOT EXISTS idx_diy_templates_name ON diy_templates(name);
CREATE INDEX IF NOT EXISTS idx_diy_templates_category ON diy_templates(category);
CREATE INDEX IF NOT EXISTS idx_diy_templates_status ON diy_templates(status);
CREATE INDEX IF NOT EXISTS idx_diy_templates_deleted_at ON diy_templates(deleted_at);

COMMENT ON TABLE diy_templates IS 'DIY 页面模板表（可应用于新页面，也可将现有页面保存为模板）';
COMMENT ON COLUMN diy_templates.category IS '分类：home首页模板/topic专题页模板/shop店铺页模板/activity活动页模板';
COMMENT ON COLUMN diy_templates.status IS '状态：0禁用 1启用';
COMMENT ON COLUMN diy_templates.pages IS '模板页面配置 JSONB（包含 components/settings，应用模板时复制到新页面）';

-- ============================================================
-- 4. diy_page_stats DIY 页面统计表（按日期+page_id 汇总）
-- ============================================================
CREATE TABLE IF NOT EXISTS diy_page_stats (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    -- 关联
    page_id BIGINT NOT NULL DEFAULT 0,                       -- 页面 ID

    -- 统计数据
    view_count INT NOT NULL DEFAULT 0,                       -- 浏览数
    click_count INT NOT NULL DEFAULT 0,                      -- 点击数
    conversion_count INT NOT NULL DEFAULT 0,                 -- 转化数
    stat_date DATE NOT NULL,                                 -- 统计日期

    -- 唯一约束：同页面同日期仅一条记录（upsert 使用）
    CONSTRAINT uniq_diy_page_stats_page_date UNIQUE (page_id, stat_date)
);
CREATE INDEX IF NOT EXISTS idx_diy_page_stats_page_id ON diy_page_stats(page_id);
CREATE INDEX IF NOT EXISTS idx_diy_page_stats_stat_date ON diy_page_stats(stat_date);
CREATE INDEX IF NOT EXISTS idx_diy_page_stats_deleted_at ON diy_page_stats(deleted_at);

COMMENT ON TABLE diy_page_stats IS 'DIY 页面统计表（按日期+page_id 汇总浏览/点击/转化数据）';
COMMENT ON COLUMN diy_page_stats.page_id IS '页面 ID';
COMMENT ON COLUMN diy_page_stats.view_count IS '浏览数（C 端访问埋点累计）';
COMMENT ON COLUMN diy_page_stats.click_count IS '点击数（C 端点击埋点累计）';
COMMENT ON COLUMN diy_page_stats.conversion_count IS '转化数（C 端转化埋点累计）';
COMMENT ON COLUMN diy_page_stats.stat_date IS '统计日期（DATE 类型，同页面同日期仅一条记录）';

-- ============================================================
-- updated_at 触发器（依赖 001_p0_baseline.sql 中的 update_updated_at_column 函数）
-- PostgreSQL 不支持 CREATE TRIGGER IF NOT EXISTS，用 DROP TRIGGER IF EXISTS + CREATE TRIGGER
-- ============================================================
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'update_updated_at_column') THEN
        DROP TRIGGER IF EXISTS trg_diy_pages_updated_at ON diy_pages; CREATE TRIGGER trg_diy_pages_updated_at BEFORE UPDATE ON diy_pages FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_diy_components_updated_at ON diy_components; CREATE TRIGGER trg_diy_components_updated_at BEFORE UPDATE ON diy_components FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_diy_templates_updated_at ON diy_templates; CREATE TRIGGER trg_diy_templates_updated_at BEFORE UPDATE ON diy_templates FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_diy_page_stats_updated_at ON diy_page_stats; CREATE TRIGGER trg_diy_page_stats_updated_at BEFORE UPDATE ON diy_page_stats FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;

-- ============================================================
-- 完成
-- ============================================================
-- 表清单（4 张）：
--   diy_pages / diy_components / diy_templates / diy_page_stats
-- 索引：20+ 个（含唯一索引 uniq_diy_components_code 与复合唯一约束 uniq_diy_page_stats_page_date）
-- 触发器：4 个 updated_at 触发器
-- 注释：表注释 + 关键字段注释
-- ============================================================
