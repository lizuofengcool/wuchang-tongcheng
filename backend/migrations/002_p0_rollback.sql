-- ============================================================
-- P0 基线回滚脚本（与 001_p0_baseline.sql 配对）
-- 按反向顺序 DROP 6 张核心表及其触发器
-- 幂等：使用 IF EXISTS，可重复执行
-- 警告：执行后 modules / cron_jobs 等表数据将全部丢失
-- ============================================================

-- ------------------------------------------------------------
-- 1. 先移除 updated_at 触发器（表存在时才有触发器）
-- ------------------------------------------------------------
DROP TRIGGER IF EXISTS trg_module_metrics_updated ON module_metrics;
DROP TRIGGER IF EXISTS trg_module_station_configs_updated ON module_station_configs;
DROP TRIGGER IF EXISTS trg_message_queue_updated ON message_queue;
DROP TRIGGER IF EXISTS trg_module_grayscales_updated ON module_grayscales;
DROP TRIGGER IF EXISTS trg_cron_jobs_updated ON cron_jobs;
DROP TRIGGER IF EXISTS trg_modules_updated ON modules;

-- ------------------------------------------------------------
-- 2. 按反向顺序 DROP 表（先 DROP 有外键依赖风险的表，本组无外键，按声明逆序）
-- ------------------------------------------------------------
DROP TABLE IF EXISTS module_metrics;
DROP TABLE IF EXISTS module_station_configs;
DROP TABLE IF EXISTS message_queue;
DROP TABLE IF EXISTS module_grayscales;
DROP TABLE IF EXISTS cron_jobs;
DROP TABLE IF EXISTS modules;

-- ------------------------------------------------------------
-- 3. update_updated_at_column 函数保留（可能被其他业务表复用）
--    如需彻底清理，取消下面注释：
-- DROP FUNCTION IF EXISTS update_updated_at_column();
-- ------------------------------------------------------------
