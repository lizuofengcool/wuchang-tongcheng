-- ============================================================
-- dh114 同城114模块回滚脚本
-- 对应：023_dh114_full.sql
--
-- 内容：
--   1. DROP TRIGGER IF EXISTS（依赖 update_updated_at_column 函数，函数不删除）
--   2. DROP TABLE IF EXISTS 18 张表（倒序删除以避免外键依赖）
--   3. 全幂等：DROP ... IF EXISTS
-- ============================================================

-- ============================================================
-- 1. 先删除触发器（依赖 update_updated_at_column 函数）
-- ============================================================
DROP TRIGGER IF EXISTS trg_dh114s_updated_at ON dh114s;
DROP TRIGGER IF EXISTS trg_dh114_business_updated_at ON dh114_business;
DROP TRIGGER IF EXISTS trg_dh114_business_hours_updated_at ON dh114_business_hours;
DROP TRIGGER IF EXISTS trg_dh114_categories_updated_at ON dh114_categories;
DROP TRIGGER IF EXISTS trg_dh114_images_updated_at ON dh114_images;
DROP TRIGGER IF EXISTS trg_dh114_tags_updated_at ON dh114_tags;
DROP TRIGGER IF EXISTS trg_dh114_menus_updated_at ON dh114_menus;
DROP TRIGGER IF EXISTS trg_dh114_coupons_updated_at ON dh114_coupons;
DROP TRIGGER IF EXISTS trg_dh114_groupbuys_updated_at ON dh114_groupbuys;
DROP TRIGGER IF EXISTS trg_dh114_reviews_updated_at ON dh114_reviews;
DROP TRIGGER IF EXISTS trg_dh114_review_replies_updated_at ON dh114_review_replies;
DROP TRIGGER IF EXISTS trg_dh114_favorites_updated_at ON dh114_favorites;
DROP TRIGGER IF EXISTS trg_dh114_phone_calls_updated_at ON dh114_phone_calls;
DROP TRIGGER IF EXISTS trg_dh114_visits_updated_at ON dh114_visits;
DROP TRIGGER IF EXISTS trg_dh114_recommendations_updated_at ON dh114_recommendations;
DROP TRIGGER IF EXISTS trg_dh114_statistics_updated_at ON dh114_statistics;
DROP TRIGGER IF EXISTS trg_dh114_audit_rules_updated_at ON dh114_audit_rules;
DROP TRIGGER IF EXISTS trg_dh114_verifications_updated_at ON dh114_verifications;

-- ============================================================
-- 2. 倒序删除 18 张表
-- 注意：dh114s 主表也会被删除；若需保留主表，请注释最后一行
-- ============================================================
DROP TABLE IF EXISTS dh114_verifications;
DROP TABLE IF EXISTS dh114_audit_rules;
DROP TABLE IF EXISTS dh114_statistics;
DROP TABLE IF EXISTS dh114_recommendations;
DROP TABLE IF EXISTS dh114_visits;
DROP TABLE IF EXISTS dh114_phone_calls;
DROP TABLE IF EXISTS dh114_favorites;
DROP TABLE IF EXISTS dh114_review_replies;
DROP TABLE IF EXISTS dh114_reviews;
DROP TABLE IF EXISTS dh114_groupbuys;
DROP TABLE IF EXISTS dh114_coupons;
DROP TABLE IF EXISTS dh114_menus;
DROP TABLE IF EXISTS dh114_tags;
DROP TABLE IF EXISTS dh114_images;
DROP TABLE IF EXISTS dh114_categories;
DROP TABLE IF EXISTS dh114_business_hours;
DROP TABLE IF EXISTS dh114_business;
DROP TABLE IF EXISTS dh114s;
