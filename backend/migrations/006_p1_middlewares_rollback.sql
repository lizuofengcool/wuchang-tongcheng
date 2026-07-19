-- ============================================================
-- P1 中台精简版回滚脚本（与 005_p1_middlewares.sql 配对）
-- 按反向顺序 DROP 5 个中台的 26 张表及其触发器
-- 幂等：使用 IF EXISTS，可重复执行
-- 警告：执行后 pay/im/material/risk/ai 表数据将全部丢失
-- ============================================================

-- ------------------------------------------------------------
-- 1. 先移除 updated_at 触发器
-- ------------------------------------------------------------
DROP TRIGGER IF EXISTS trg_ai_generations_updated ON ai_generations;
DROP TRIGGER IF EXISTS trg_ai_prompts_updated ON ai_prompts;
DROP TRIGGER IF EXISTS trg_ai_models_updated ON ai_models;
DROP TRIGGER IF EXISTS trg_ai_tasks_updated ON ai_tasks;

DROP TRIGGER IF EXISTS trg_risk_violations_updated ON risk_violations;
DROP TRIGGER IF EXISTS trg_risk_user_scores_updated ON risk_user_scores;
DROP TRIGGER IF EXISTS trg_risk_blacklist_updated ON risk_blacklist;
DROP TRIGGER IF EXISTS trg_risk_audit_rules_updated ON risk_audit_rules;
DROP TRIGGER IF EXISTS trg_risk_sensitive_words_updated ON risk_sensitive_words;
DROP TRIGGER IF EXISTS trg_risk_reports_updated ON risk_reports;

DROP TRIGGER IF EXISTS trg_mat_image_features_updated ON mat_image_features;
DROP TRIGGER IF EXISTS trg_mat_videos_updated ON mat_videos;
DROP TRIGGER IF EXISTS trg_mat_images_updated ON mat_images;
DROP TRIGGER IF EXISTS trg_mat_files_updated ON mat_files;

DROP TRIGGER IF EXISTS trg_im_templates_updated ON im_templates;
DROP TRIGGER IF EXISTS trg_im_privacy_numbers_updated ON im_privacy_numbers;
DROP TRIGGER IF EXISTS trg_im_system_notifications_updated ON im_system_notifications;
DROP TRIGGER IF EXISTS trg_im_messages_updated ON im_messages;
DROP TRIGGER IF EXISTS trg_im_sessions_updated ON im_sessions;

DROP TRIGGER IF EXISTS trg_pay_accounts_updated ON pay_accounts;
DROP TRIGGER IF EXISTS trg_pay_settlements_updated ON pay_settlements;
DROP TRIGGER IF EXISTS trg_pay_withdrawals_updated ON pay_withdrawals;
DROP TRIGGER IF EXISTS trg_pay_refunds_updated ON pay_refunds;
DROP TRIGGER IF EXISTS trg_pay_escrows_updated ON pay_escrows;
DROP TRIGGER IF EXISTS trg_pay_orders_updated ON pay_orders;
DROP TRIGGER IF EXISTS trg_pay_methods_updated ON pay_methods;

-- ------------------------------------------------------------
-- 2. 按反向顺序 DROP 表（AI → Risk → Material → IM → Pay）
-- ------------------------------------------------------------
-- AI 中台
DROP TABLE IF EXISTS ai_generations;
DROP TABLE IF EXISTS ai_prompts;
DROP TABLE IF EXISTS ai_models;
DROP TABLE IF EXISTS ai_tasks;

-- Risk 中台
DROP TABLE IF EXISTS risk_violations;
DROP TABLE IF EXISTS risk_user_scores;
DROP TABLE IF EXISTS risk_blacklist;
DROP TABLE IF EXISTS risk_audit_rules;
DROP TABLE IF EXISTS risk_sensitive_words;
DROP TABLE IF EXISTS risk_reports;

-- Material 中台
DROP TABLE IF EXISTS mat_image_features;
DROP TABLE IF EXISTS mat_videos;
DROP TABLE IF EXISTS mat_images;
DROP TABLE IF EXISTS mat_files;

-- IM 中台
DROP TABLE IF EXISTS im_templates;
DROP TABLE IF EXISTS im_privacy_numbers;
DROP TABLE IF EXISTS im_system_notifications;
DROP TABLE IF EXISTS im_messages;
DROP TABLE IF EXISTS im_sessions;

-- Pay 中台
DROP TABLE IF EXISTS pay_accounts;
DROP TABLE IF EXISTS pay_settlements;
DROP TABLE IF EXISTS pay_withdrawals;
DROP TABLE IF EXISTS pay_refunds;
DROP TABLE IF EXISTS pay_escrows;
DROP TABLE IF EXISTS pay_orders;
DROP TABLE IF EXISTS pay_methods;

-- ------------------------------------------------------------
-- 3. update_updated_at_column 函数保留（由 002_p0_rollback.sql 统一清理）
-- ------------------------------------------------------------
