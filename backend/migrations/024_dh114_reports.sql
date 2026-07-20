-- ============================================================
-- dh114 举报表迁移脚本
-- 提供用户对商户/评价/团购等内容的举报功能
--
-- 内容：
--   1. CREATE dh114_reports 举报表
--   2. 索引、注释
--   3. 全幂等：CREATE TABLE IF NOT EXISTS / CREATE INDEX IF NOT EXISTS
-- 依赖：001_p0_baseline.sql（update_updated_at_column 函数已创建）
--
-- 表名遵循 model.TableName() 定义：Dh114Report → dh114_reports
-- ============================================================

-- dh114 举报表
CREATE TABLE IF NOT EXISTS dh114_reports (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    report_no VARCHAR(32) NOT NULL DEFAULT '',
    reporter_id BIGINT NOT NULL DEFAULT 0,
    reporter_name VARCHAR(50) NOT NULL DEFAULT '',

    target_type VARCHAR(32) NOT NULL DEFAULT '',
    target_id BIGINT NOT NULL DEFAULT 0,
    target_name VARCHAR(200) NOT NULL DEFAULT '',

    report_type VARCHAR(32) NOT NULL DEFAULT '',
    report_reason VARCHAR(500) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    evidence_images JSONB,

    status INT NOT NULL DEFAULT 0, -- 0待处理 1处理中 2已处理 3已驳回
    handler_id BIGINT,
    handler_name VARCHAR(50) NOT NULL DEFAULT '',
    handle_result VARCHAR(500) NOT NULL DEFAULT '',
    handled_at TIMESTAMPTZ,

    penalty_type VARCHAR(32) NOT NULL DEFAULT '',
    penalty_target_id BIGINT,

    contact_info VARCHAR(100) NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_dh114_reports_region_id ON dh114_reports(region_id);
CREATE INDEX IF NOT EXISTS idx_dh114_reports_report_no ON dh114_reports(report_no);
CREATE INDEX IF NOT EXISTS idx_dh114_reports_reporter_id ON dh114_reports(reporter_id);
CREATE INDEX IF NOT EXISTS idx_dh114_reports_target_type ON dh114_reports(target_type);
CREATE INDEX IF NOT EXISTS idx_dh114_reports_target_id ON dh114_reports(target_id);
CREATE INDEX IF NOT EXISTS idx_dh114_reports_report_type ON dh114_reports(report_type);
CREATE INDEX IF NOT EXISTS idx_dh114_reports_status ON dh114_reports(status);
CREATE INDEX IF NOT EXISTS idx_dh114_reports_deleted_at ON dh114_reports(deleted_at);

COMMENT ON TABLE dh114_reports IS '同城114举报表';

-- updated_at 触发器（依赖 001_p0_baseline.sql 中的 update_updated_at_column 函数）
-- 使用 DROP IF EXISTS + CREATE 保证幂等（PostgreSQL 不支持 CREATE TRIGGER IF NOT EXISTS）
DROP TRIGGER IF EXISTS trg_dh114_reports_updated_at ON dh114_reports;
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'update_updated_at_column') THEN
        CREATE TRIGGER trg_dh114_reports_updated_at
            BEFORE UPDATE ON dh114_reports
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;
