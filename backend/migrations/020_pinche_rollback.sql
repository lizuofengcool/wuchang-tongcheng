-- ============================================================
-- pinche 拼车出行模块回滚脚本
-- 对应：019_pinche_full.sql
--
-- 内容：
--   1. DROP TABLE IF EXISTS 18 张表（倒序删除以避免外键依赖）
--   2. DROP TRIGGER IF EXISTS（依赖 update_updated_at_column 函数，函数不删除）
--   3. 全幂等：DROP ... IF EXISTS
-- ============================================================

-- ============================================================
-- 1. 先删除触发器
-- ============================================================
DROP TRIGGER IF EXISTS trg_pinches_updated_at ON pinches;
DROP TRIGGER IF EXISTS trg_pinche_routes_updated_at ON pinche_routes;
DROP TRIGGER IF EXISTS trg_pinche_bookings_updated_at ON pinche_bookings;
DROP TRIGGER IF EXISTS trg_pinche_drivers_updated_at ON pinche_drivers;
DROP TRIGGER IF EXISTS trg_pinche_vehicles_updated_at ON pinche_vehicles;
DROP TRIGGER IF EXISTS trg_pinche_insurances_updated_at ON pinche_insurances;
DROP TRIGGER IF EXISTS trg_pinche_payments_updated_at ON pinche_payments;
DROP TRIGGER IF EXISTS trg_pinche_ratings_updated_at ON pinche_ratings;
DROP TRIGGER IF EXISTS trg_pinche_emergencies_updated_at ON pinche_emergencies;
DROP TRIGGER IF EXISTS trg_pinche_trips_updated_at ON pinche_trips;
DROP TRIGGER IF EXISTS trg_pinche_route_favorites_updated_at ON pinche_route_favorites;
DROP TRIGGER IF EXISTS trg_pinche_messages_updated_at ON pinche_messages;
DROP TRIGGER IF EXISTS trg_pinche_cancels_updated_at ON pinche_cancels;
DROP TRIGGER IF EXISTS trg_pinche_refunds_updated_at ON pinche_refunds;
DROP TRIGGER IF EXISTS trg_pinche_complaints_updated_at ON pinche_complaints;
DROP TRIGGER IF EXISTS trg_pinche_audit_rules_updated_at ON pinche_audit_rules;
DROP TRIGGER IF EXISTS trg_pinche_statistics_updated_at ON pinche_statistics;

-- ============================================================
-- 2. 倒序删除 18 张表
-- ============================================================
DROP TABLE IF EXISTS pinche_statistics;
DROP TABLE IF EXISTS pinche_audit_rules;
DROP TABLE IF EXISTS pinche_complaints;
DROP TABLE IF EXISTS pinche_refunds;
DROP TABLE IF EXISTS pinche_cancels;
DROP TABLE IF EXISTS pinche_messages;
DROP TABLE IF EXISTS pinche_driver_locations;
DROP TABLE IF EXISTS pinche_route_favorites;
DROP TABLE IF EXISTS pinche_trips;
DROP TABLE IF EXISTS pinche_emergencies;
DROP TABLE IF EXISTS pinche_ratings;
DROP TABLE IF EXISTS pinche_payments;
DROP TABLE IF EXISTS pinche_insurances;
DROP TABLE IF EXISTS pinche_vehicles;
DROP TABLE IF EXISTS pinche_drivers;
DROP TABLE IF EXISTS pinche_bookings;
DROP TABLE IF EXISTS pinche_routes;
DROP TABLE IF EXISTS pinches;
