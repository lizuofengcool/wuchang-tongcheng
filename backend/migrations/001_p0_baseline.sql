-- ============================================================
-- P0 基线迁移脚本（v3.2.1 架构方案 6.1/6.2/6.5/9.1/11.5 节）
-- 创建 6 张核心表：modules / cron_jobs / module_grayscales /
--                  message_queue / module_station_configs / module_metrics
-- 幂等：所有对象使用 IF NOT EXISTS 创建，可重复执行
-- ============================================================

-- ------------------------------------------------------------
-- 1. modules 模块注册表
--    记录 12 中台 + 15 垂直业务的安装状态、版本、依赖关系
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS modules (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(64) NOT NULL UNIQUE,
    display_name VARCHAR(128) NOT NULL DEFAULT '',
    category VARCHAR(32) NOT NULL DEFAULT 'business',
    description VARCHAR(512) NOT NULL DEFAULT '',
    version VARCHAR(32) NOT NULL DEFAULT '1.0.0',
    dependencies JSONB NOT NULL DEFAULT '[]'::jsonb,
    icon VARCHAR(256) NOT NULL DEFAULT '',
    author VARCHAR(64) NOT NULL DEFAULT '',
    homepage VARCHAR(256) NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    installed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_modules_category ON modules(category);
CREATE INDEX IF NOT EXISTS idx_modules_enabled ON modules(enabled);

-- ------------------------------------------------------------
-- 2. cron_jobs 定时任务调度中心
--    module_name + job_name 唯一，handler 指向 Go 注册的处理函数
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS cron_jobs (
    id BIGSERIAL PRIMARY KEY,
    module_name VARCHAR(64) NOT NULL,
    job_name VARCHAR(128) NOT NULL,
    cron_expr VARCHAR(64) NOT NULL,
    handler VARCHAR(256) NOT NULL,
    params JSONB NOT NULL DEFAULT '{}'::jsonb,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    last_run_at TIMESTAMPTZ,
    next_run_at TIMESTAMPTZ,
    last_status VARCHAR(32) NOT NULL DEFAULT '',
    last_error TEXT NOT NULL DEFAULT '',
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retry INTEGER NOT NULL DEFAULT 3,
    timeout_seconds INTEGER NOT NULL DEFAULT 300,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(module_name, job_name)
);
CREATE INDEX IF NOT EXISTS idx_cron_jobs_enabled ON cron_jobs(enabled);
CREATE INDEX IF NOT EXISTS idx_cron_jobs_next_run ON cron_jobs(next_run_at) WHERE enabled = TRUE;

-- ------------------------------------------------------------
-- 3. module_grayscales 灰度发布表
--    gray_type: user/city/percentage
--    status: pending/releasing/rollback/completed
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS module_grayscales (
    id BIGSERIAL PRIMARY KEY,
    module_name VARCHAR(64) NOT NULL,
    version VARCHAR(32) NOT NULL,
    gray_type VARCHAR(32) NOT NULL,
    gray_config JSONB NOT NULL DEFAULT '{}'::jsonb,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    target_count INTEGER NOT NULL DEFAULT 0,
    actual_count INTEGER NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_by VARCHAR(64) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_grayscales_module ON module_grayscales(module_name);
CREATE INDEX IF NOT EXISTS idx_grayscales_status ON module_grayscales(status);

-- ------------------------------------------------------------
-- 4. message_queue 本地消息表（最终一致性）
--    status: pending/sent/failed/dead
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS message_queue (
    id BIGSERIAL PRIMARY KEY,
    biz_module VARCHAR(64) NOT NULL,
    biz_id VARCHAR(128) NOT NULL,
    message_type VARCHAR(64) NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retry INTEGER NOT NULL DEFAULT 5,
    next_retry_at TIMESTAMPTZ,
    last_error TEXT NOT NULL DEFAULT '',
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_mq_status ON message_queue(status);
CREATE INDEX IF NOT EXISTS idx_mq_next_retry ON message_queue(next_retry_at) WHERE status = 'pending' OR status = 'failed';
CREATE INDEX IF NOT EXISTS idx_mq_biz ON message_queue(biz_module, biz_id);

-- ------------------------------------------------------------
-- 5. module_station_configs 分站配置中心扩展
--    station_id 关联 regions 表，按 (station_id, module_name, config_key) 唯一
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS module_station_configs (
    id BIGSERIAL PRIMARY KEY,
    station_id BIGINT NOT NULL,
    module_name VARCHAR(64) NOT NULL,
    config_key VARCHAR(128) NOT NULL,
    config_value JSONB NOT NULL DEFAULT '{}'::jsonb,
    config_type VARCHAR(32) NOT NULL DEFAULT 'string',
    description VARCHAR(256) NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(station_id, module_name, config_key)
);
CREATE INDEX IF NOT EXISTS idx_station_configs_station ON module_station_configs(station_id);
CREATE INDEX IF NOT EXISTS idx_station_configs_module ON module_station_configs(module_name);

-- ------------------------------------------------------------
-- 6. module_metrics 监控指标表
--    配合 Prometheus + Grafana，存储模块级业务指标
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS module_metrics (
    id BIGSERIAL PRIMARY KEY,
    module_name VARCHAR(64) NOT NULL,
    metric_name VARCHAR(128) NOT NULL,
    metric_value DOUBLE PRECISION NOT NULL,
    labels JSONB NOT NULL DEFAULT '{}'::jsonb,
    recorded_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_metrics_module_name ON module_metrics(module_name, metric_name);
CREATE INDEX IF NOT EXISTS idx_metrics_recorded_at ON module_metrics(recorded_at);

-- ------------------------------------------------------------
-- updated_at 触发器函数（幂等：CREATE OR REPLACE）
-- ------------------------------------------------------------
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- ------------------------------------------------------------
-- 为 6 张表挂载 updated_at 触发器（幂等：先 DROP IF EXISTS 再 CREATE）
-- ------------------------------------------------------------
DO $$
DECLARE t TEXT;
BEGIN
    FOR t IN SELECT tablename FROM pg_tables WHERE schemaname='public' AND tablename IN ('modules','cron_jobs','module_grayscales','message_queue','module_station_configs','module_metrics')
    LOOP
        EXECUTE format('DROP TRIGGER IF EXISTS trg_%s_updated ON %s', t, t);
        EXECUTE format('CREATE TRIGGER trg_%s_updated BEFORE UPDATE ON %s FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()', t, t);
    END LOOP;
END $$;
