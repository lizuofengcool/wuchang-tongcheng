-- ============================================================
-- 015_risk_full.sql 风控/举报中台扩展表
-- 对标字节风控/腾讯安全/阿里风控
-- 创建 risk_report_evidence / risk_appeals / risk_rules / risk_score_records
-- / risk_audit_logs
-- 现有的 risk_reports / risk_sensitive_words / risk_audit_rules / risk_blacklist
-- / risk_user_scores / risk_violations 已在 005 创建
-- 全部幂等：CREATE TABLE IF NOT EXISTS
-- ============================================================

-- ============================================================
-- 1. risk_report_evidence 举报证据
-- ============================================================
CREATE TABLE IF NOT EXISTS risk_report_evidence (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    report_id BIGINT NOT NULL,
    evidence_type VARCHAR(16) NOT NULL DEFAULT 'image',          -- image/video/text/audio
    url VARCHAR(512) NOT NULL DEFAULT '',
    description VARCHAR(256) NOT NULL DEFAULT '',
    uploader_id BIGINT NOT NULL DEFAULT 0,
    file_size BIGINT NOT NULL DEFAULT 0,
    file_hash VARCHAR(64) NOT NULL DEFAULT '',
    extra JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_risk_report_evidence_report_id ON risk_report_evidence(report_id);
CREATE INDEX IF NOT EXISTS idx_risk_report_evidence_type ON risk_report_evidence(evidence_type);
CREATE INDEX IF NOT EXISTS idx_risk_report_evidence_region_id ON risk_report_evidence(region_id);
COMMENT ON TABLE risk_report_evidence IS '风控举报证据表';

-- ============================================================
-- 2. risk_appeals 申诉记录
-- ============================================================
CREATE TABLE IF NOT EXISTS risk_appeals (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    appeal_no VARCHAR(64) NOT NULL UNIQUE,
    violation_id BIGINT NOT NULL,                                -- 关联违规ID
    user_id BIGINT NOT NULL,
    reason TEXT NOT NULL DEFAULT '',                              -- 申诉理由
    evidence_images JSONB NOT NULL DEFAULT '[]'::jsonb,            -- 证据图片
    status SMALLINT NOT NULL DEFAULT 0,                           -- 0待审 1通过 2拒绝
    handler_id BIGINT NOT NULL DEFAULT 0,
    handle_remark TEXT NOT NULL DEFAULT '',
    handled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_risk_appeals_violation_id ON risk_appeals(violation_id);
CREATE INDEX IF NOT EXISTS idx_risk_appeals_user_id ON risk_appeals(user_id);
CREATE INDEX IF NOT EXISTS idx_risk_appeals_status ON risk_appeals(status);
CREATE INDEX IF NOT EXISTS idx_risk_appeals_region_id ON risk_appeals(region_id);
COMMENT ON TABLE risk_appeals IS '风控违规申诉表';

-- ============================================================
-- 3. risk_rules 风控规则（独立于 risk_audit_rules，用于实时风控决策）
-- ============================================================
CREATE TABLE IF NOT EXISTS risk_rules (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    rule_name VARCHAR(64) NOT NULL UNIQUE,
    rule_type VARCHAR(32) NOT NULL DEFAULT '',                   -- frequency/amount/content/behavior
    description VARCHAR(256) NOT NULL DEFAULT '',
    config JSONB NOT NULL DEFAULT '{}'::jsonb,                     -- 规则配置（阈值、时间窗等）
    action VARCHAR(32) NOT NULL DEFAULT 'block',                  -- block/review/warn/log
    priority INTEGER NOT NULL DEFAULT 100,                        -- 优先级（越小越高）
    status SMALLINT NOT NULL DEFAULT 1,
    hit_count BIGINT NOT NULL DEFAULT 0,                          -- 命中次数
    last_hit_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_risk_rules_type ON risk_rules(rule_type);
CREATE INDEX IF NOT EXISTS idx_risk_rules_status ON risk_rules(status);
CREATE INDEX IF NOT EXISTS idx_risk_rules_priority ON risk_rules(priority);
CREATE INDEX IF NOT EXISTS idx_risk_rules_region_id ON risk_rules(region_id);
COMMENT ON TABLE risk_rules IS '风控规则表（实时决策）';

-- ============================================================
-- 4. risk_score_records 风险评分记录
-- ============================================================
CREATE TABLE IF NOT EXISTS risk_score_records (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    user_id BIGINT NOT NULL,
    target_type VARCHAR(32) NOT NULL DEFAULT 'user',             -- user/content/ip/device
    target_value VARCHAR(128) NOT NULL DEFAULT '',                -- 目标值
    content_type VARCHAR(32) NOT NULL DEFAULT '',                 -- 内容类型
    content_id VARCHAR(128) NOT NULL DEFAULT '',                  -- 内容ID
    score INTEGER NOT NULL DEFAULT 0,                              -- 风险评分 0-100
    level VARCHAR(16) NOT NULL DEFAULT 'safe',                    -- safe/warning/danger
    reasons JSONB NOT NULL DEFAULT '[]'::jsonb,                    -- 命中原因
    rule_ids JSONB NOT NULL DEFAULT '[]'::jsonb,                   -- 命中的规则ID列表
    action_taken VARCHAR(32) NOT NULL DEFAULT '',                 -- 实际处理动作
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_risk_score_records_user_id ON risk_score_records(user_id);
CREATE INDEX IF NOT EXISTS idx_risk_score_records_target ON risk_score_records(target_type, target_value);
CREATE INDEX IF NOT EXISTS idx_risk_score_records_level ON risk_score_records(level);
CREATE INDEX IF NOT EXISTS idx_risk_score_records_created_at ON risk_score_records(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_risk_score_records_region_id ON risk_score_records(region_id);
COMMENT ON TABLE risk_score_records IS '风控风险评分记录';

-- ============================================================
-- 5. risk_audit_logs 审核日志
-- ============================================================
CREATE TABLE IF NOT EXISTS risk_audit_logs (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    auditor_id BIGINT NOT NULL,                                  -- 审核人ID
    auditor_name VARCHAR(64) NOT NULL DEFAULT '',
    action VARCHAR(32) NOT NULL DEFAULT '',                      -- approve/reject/escalate/ban/unban
    target_type VARCHAR(32) NOT NULL DEFAULT '',                 -- report/appeal/blacklist/content
    target_id BIGINT NOT NULL DEFAULT 0,
    biz_module VARCHAR(32) NOT NULL DEFAULT '',
    biz_id VARCHAR(128) NOT NULL DEFAULT '',
    before_status VARCHAR(32) NOT NULL DEFAULT '',
    after_status VARCHAR(32) NOT NULL DEFAULT '',
    remark TEXT NOT NULL DEFAULT '',
    ip VARCHAR(64) NOT NULL DEFAULT '',
    user_agent VARCHAR(256) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_risk_audit_logs_auditor_id ON risk_audit_logs(auditor_id);
CREATE INDEX IF NOT EXISTS idx_risk_audit_logs_action ON risk_audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_risk_audit_logs_target ON risk_audit_logs(target_type, target_id);
CREATE INDEX IF NOT EXISTS idx_risk_audit_logs_created_at ON risk_audit_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_risk_audit_logs_region_id ON risk_audit_logs(region_id);
COMMENT ON TABLE risk_audit_logs IS '风控审核日志';
