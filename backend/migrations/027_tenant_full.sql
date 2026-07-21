-- =====================================================
-- tenant 多租户分站中台模块完整迁移脚本
-- 包含 4 张表 + 索引 + 触发器 + 注释
-- 依据架构设计 ershou-大模块架构方案.md 第 4.10 节：多租户分站中台
-- 数据库：wuchang_tongcheng
-- =====================================================
--
-- 内容：
--   1. CREATE 4 张表（tenant_stations / tenant_staff / tenant_configs / tenant_domains）
--   2. 索引、updated_at 触发器、COMMENT 注释
--   3. 全幂等：CREATE TABLE IF NOT EXISTS / CREATE INDEX IF NOT EXISTS
-- 依赖：001_p0_baseline.sql（update_updated_at_column 函数已创建）
--
-- 说明：4 张表亦由 GORM AutoMigrate 创建，本脚本保证索引/触发器/注释完整且幂等。
--      表名严格遵循 model.TableName() 定义：
--      - Station → tenant_stations
--      - Staff   → tenant_staff
--      - Config  → tenant_configs
--      - Domain  → tenant_domains
-- =====================================================

-- ============================================================
-- 1. tenant_stations 分站主表（region_id 唯一，一地区一分站）
-- ============================================================
CREATE TABLE IF NOT EXISTS tenant_stations (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    region_id BIGINT NOT NULL,                       -- 地区 ID（唯一，一地区一分站）
    name VARCHAR(100) NOT NULL DEFAULT '',           -- 分站名称
    domain VARCHAR(200) NOT NULL DEFAULT '',         -- 主域名（冗余，便于快速查询）
    logo VARCHAR(255) NOT NULL DEFAULT '',           -- 分站 Logo
    description TEXT,                                -- 分站描述
    status INT NOT NULL DEFAULT 1,                   -- 0已停用 1已启用
    config JSONB                                     -- 独立运营配置（JSONB）
);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_tenant_stations_region_id ON tenant_stations(region_id);
CREATE INDEX IF NOT EXISTS idx_tenant_stations_domain ON tenant_stations(domain);
CREATE INDEX IF NOT EXISTS idx_tenant_stations_status ON tenant_stations(status);
CREATE INDEX IF NOT EXISTS idx_tenant_stations_name ON tenant_stations(name);

COMMENT ON TABLE tenant_stations IS '多租户分站主表（一地区一分站，region_id 唯一）';
COMMENT ON COLUMN tenant_stations.region_id IS '地区 ID（唯一，一个地区只能对应一个分站）';
COMMENT ON COLUMN tenant_stations.name IS '分站名称';
COMMENT ON COLUMN tenant_stations.domain IS '主域名（冗余字段，便于快速查询，与 tenant_domains 主域名同步）';
COMMENT ON COLUMN tenant_stations.logo IS '分站 Logo';
COMMENT ON COLUMN tenant_stations.description IS '分站描述';
COMMENT ON COLUMN tenant_stations.status IS '状态：0已停用 1已启用';
COMMENT ON COLUMN tenant_stations.config IS '独立运营配置（JSONB）';

-- ============================================================
-- 2. tenant_staff 分站员工表（operator运营员 / manager管理员）
-- ============================================================
CREATE TABLE IF NOT EXISTS tenant_staff (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    station_id BIGINT NOT NULL,                      -- 分站 ID
    user_id BIGINT NOT NULL,                         -- 用户 ID
    role VARCHAR(20) NOT NULL DEFAULT 'operator',    -- operator运营员 / manager管理员
    permissions JSONB,                               -- 权限列表（JSONB）
    status INT NOT NULL DEFAULT 1                    -- 0已停用 1已启用
);
CREATE INDEX IF NOT EXISTS idx_tenant_staff_station_id ON tenant_staff(station_id);
CREATE INDEX IF NOT EXISTS idx_tenant_staff_user_id ON tenant_staff(user_id);
CREATE INDEX IF NOT EXISTS idx_tenant_staff_role ON tenant_staff(role);
CREATE INDEX IF NOT EXISTS idx_tenant_staff_status ON tenant_staff(status);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_tenant_staff_station_user ON tenant_staff(station_id, user_id);

COMMENT ON TABLE tenant_staff IS '分站员工表（独立运营权限）';
COMMENT ON COLUMN tenant_staff.station_id IS '分站 ID';
COMMENT ON COLUMN tenant_staff.user_id IS '用户 ID';
COMMENT ON COLUMN tenant_staff.role IS '角色：operator运营员 / manager管理员';
COMMENT ON COLUMN tenant_staff.permissions IS '权限列表（JSONB）';
COMMENT ON COLUMN tenant_staff.status IS '状态：0已停用 1已启用';

-- ============================================================
-- 3. tenant_configs 分站配置表（按 biz_module + config_key 维度）
-- ============================================================
CREATE TABLE IF NOT EXISTS tenant_configs (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    station_id BIGINT NOT NULL,                      -- 分站 ID
    biz_module VARCHAR(50) NOT NULL DEFAULT '',      -- 业务模块（如 dh114/mall/ershou）
    config_key VARCHAR(100) NOT NULL DEFAULT '',     -- 配置键
    config_value TEXT                                -- 配置值（TEXT）
);
CREATE INDEX IF NOT EXISTS idx_tenant_configs_station_id ON tenant_configs(station_id);
CREATE INDEX IF NOT EXISTS idx_tenant_configs_biz_module ON tenant_configs(biz_module);
CREATE INDEX IF NOT EXISTS idx_tenant_configs_config_key ON tenant_configs(config_key);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_tenant_configs_station_module_key ON tenant_configs(station_id, biz_module, config_key);

COMMENT ON TABLE tenant_configs IS '分站配置表（按 biz_module + config_key 维度，支持配置继承）';
COMMENT ON COLUMN tenant_configs.station_id IS '分站 ID';
COMMENT ON COLUMN tenant_configs.biz_module IS '业务模块（如 dh114/mall/ershou）';
COMMENT ON COLUMN tenant_configs.config_key IS '配置键';
COMMENT ON COLUMN tenant_configs.config_value IS '配置值（TEXT）';

-- ============================================================
-- 4. tenant_domains 域名绑定表（domain 唯一，主域名 + SSL 状态）
-- ============================================================
CREATE TABLE IF NOT EXISTS tenant_domains (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    station_id BIGINT NOT NULL,                      -- 分站 ID
    domain VARCHAR(200) NOT NULL,                    -- 域名（唯一）
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,       -- 是否主域名
    ssl_status VARCHAR(20) NOT NULL DEFAULT 'none'   -- none未配置 / pending申请中 / active已生效 / failed失败
);
CREATE INDEX IF NOT EXISTS idx_tenant_domains_station_id ON tenant_domains(station_id);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_tenant_domains_domain ON tenant_domains(domain);
CREATE INDEX IF NOT EXISTS idx_tenant_domains_is_primary ON tenant_domains(is_primary) WHERE is_primary = TRUE;
CREATE INDEX IF NOT EXISTS idx_tenant_domains_ssl_status ON tenant_domains(ssl_status);

COMMENT ON TABLE tenant_domains IS '域名绑定表（一个分站可绑定多个域名，其中一个为主域名）';
COMMENT ON COLUMN tenant_domains.station_id IS '分站 ID';
COMMENT ON COLUMN tenant_domains.domain IS '域名（唯一）';
COMMENT ON COLUMN tenant_domains.is_primary IS '是否主域名';
COMMENT ON COLUMN tenant_domains.ssl_status IS 'SSL 状态：none未配置 / pending申请中 / active已生效 / failed失败';

-- ============================================================
-- updated_at 触发器（依赖 001_p0_baseline.sql 中的 update_updated_at_column 函数）
-- PostgreSQL 不支持 CREATE TRIGGER IF NOT EXISTS，用 DROP TRIGGER IF EXISTS + CREATE TRIGGER
-- ============================================================
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'update_updated_at_column') THEN
        DROP TRIGGER IF EXISTS trg_tenant_stations_updated_at ON tenant_stations; CREATE TRIGGER trg_tenant_stations_updated_at BEFORE UPDATE ON tenant_stations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_tenant_staff_updated_at ON tenant_staff; CREATE TRIGGER trg_tenant_staff_updated_at BEFORE UPDATE ON tenant_staff FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_tenant_configs_updated_at ON tenant_configs; CREATE TRIGGER trg_tenant_configs_updated_at BEFORE UPDATE ON tenant_configs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_tenant_domains_updated_at ON tenant_domains; CREATE TRIGGER trg_tenant_domains_updated_at BEFORE UPDATE ON tenant_domains FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;
