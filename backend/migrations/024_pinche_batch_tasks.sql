-- ============================================================
-- 024_pinche_batch_tasks.sql
-- 拼车模块 - 批量任务表
-- 复用已有 region_id 地区隔离设计；幂等（IF NOT EXISTS）
-- ============================================================

CREATE TABLE IF NOT EXISTS pinche_batch_tasks (
    id              BIGSERIAL    PRIMARY KEY,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,
    region_id       BIGINT       NOT NULL DEFAULT 1,
    task_no         VARCHAR(32)  NOT NULL,
    task_name       VARCHAR(100) NOT NULL DEFAULT '',
    task_type       VARCHAR(32)  NOT NULL DEFAULT '',
    target_ids      JSONB,
    target_count    INT          NOT NULL DEFAULT 0,
    filters         JSONB,
    action          VARCHAR(32)  NOT NULL DEFAULT '',
    action_params   JSONB,
    status          INT          NOT NULL DEFAULT 0, -- 0待执行 1执行中 2已完成 3失败 4已取消
    progress        INT          NOT NULL DEFAULT 0,
    success_count   INT          NOT NULL DEFAULT 0,
    fail_count      INT          NOT NULL DEFAULT 0,
    fail_reason     TEXT         NOT NULL DEFAULT '',
    operator_id     BIGINT,
    operator_name   VARCHAR(50)  NOT NULL DEFAULT '',
    started_at      TIMESTAMPTZ,
    finished_at     TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_pinche_batch_tasks_region_id ON pinche_batch_tasks(region_id);
CREATE INDEX IF NOT EXISTS idx_pinche_batch_tasks_task_no  ON pinche_batch_tasks(task_no);
CREATE INDEX IF NOT EXISTS idx_pinche_batch_tasks_task_type ON pinche_batch_tasks(task_type);
CREATE INDEX IF NOT EXISTS idx_pinche_batch_tasks_status   ON pinche_batch_tasks(status);
CREATE INDEX IF NOT EXISTS idx_pinche_batch_tasks_deleted_at ON pinche_batch_tasks(deleted_at);

COMMENT ON TABLE  pinche_batch_tasks IS '拼车批量任务表';
COMMENT ON COLUMN pinche_batch_tasks.status IS '0待执行 1执行中 2已完成 3失败 4已取消';
