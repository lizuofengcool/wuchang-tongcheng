-- ============================================================
-- linggong 零工兼职模块完整功能回滚脚本（与 021_linggong_full.sql 配对）
-- 按反向顺序 DROP 17 张子表 + 触发器 + 主表新增字段
-- 幂等：使用 IF EXISTS，可重复执行
-- 警告：执行后 linggong 模块全部扩展数据将丢失（主表基础字段保留）
-- ============================================================

-- ------------------------------------------------------------
-- 1. 先移除 17 张子表的 updated_at 触发器
-- ------------------------------------------------------------
DROP TRIGGER IF EXISTS trg_linggong_statistics_updated ON linggong_statistics;
DROP TRIGGER IF EXISTS trg_linggong_audit_rules_updated ON linggong_audit_rules;
DROP TRIGGER IF EXISTS trg_linggong_favorites_updated ON linggong_favorites;
DROP TRIGGER IF EXISTS trg_linggong_recommendations_updated ON linggong_recommendations;
DROP TRIGGER IF EXISTS trg_linggong_withdrawals_updated ON linggong_withdrawals;
DROP TRIGGER IF EXISTS trg_linggong_disputes_updated ON linggong_disputes;
DROP TRIGGER IF EXISTS trg_linggong_attendances_updated ON linggong_attendances;
DROP TRIGGER IF EXISTS trg_linggong_certifications_updated ON linggong_certifications;
DROP TRIGGER IF EXISTS trg_linggong_credits_updated ON linggong_credits;
DROP TRIGGER IF EXISTS trg_linggong_skills_updated ON linggong_skills;
DROP TRIGGER IF EXISTS trg_linggong_ratings_updated ON linggong_ratings;
DROP TRIGGER IF EXISTS trg_linggong_payments_updated ON linggong_payments;
DROP TRIGGER IF EXISTS trg_linggong_contracts_updated ON linggong_contracts;
DROP TRIGGER IF EXISTS trg_linggong_workers_updated ON linggong_workers;
DROP TRIGGER IF EXISTS trg_linggong_employers_updated ON linggong_employers;
DROP TRIGGER IF EXISTS trg_linggong_applications_updated ON linggong_applications;
DROP TRIGGER IF EXISTS trg_linggong_tasks_updated ON linggong_tasks;

-- ------------------------------------------------------------
-- 2. 按反向顺序 DROP 17 张子表
--    依赖顺序：linggong_disputes → linggong_payments → linggong_contracts；
--             linggong_attendances → linggong_applications → linggong_tasks → linggongs
-- ------------------------------------------------------------
DROP TABLE IF EXISTS linggong_statistics;
DROP TABLE IF EXISTS linggong_audit_rules;
DROP TABLE IF EXISTS linggong_favorites;
DROP TABLE IF EXISTS linggong_recommendations;
DROP TABLE IF EXISTS linggong_withdrawals;
DROP TABLE IF EXISTS linggong_disputes;
DROP TABLE IF EXISTS linggong_attendances;
DROP TABLE IF EXISTS linggong_certifications;
DROP TABLE IF EXISTS linggong_credits;
DROP TABLE IF EXISTS linggong_skills;
DROP TABLE IF EXISTS linggong_ratings;
DROP TABLE IF EXISTS linggong_payments;
DROP TABLE IF EXISTS linggong_contracts;
DROP TABLE IF EXISTS linggong_workers;
DROP TABLE IF EXISTS linggong_employers;
DROP TABLE IF EXISTS linggong_applications;
DROP TABLE IF EXISTS linggong_tasks;

-- ------------------------------------------------------------
-- 3. 移除 linggongs 主表新增字段（按反向顺序）
--    包装在 DO 块中，表不存在时跳过（幂等）
-- ------------------------------------------------------------
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'linggongs') THEN
        -- === 雇主认证状态 ===
        ALTER TABLE linggongs DROP COLUMN IF EXISTS employer_verified_at;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS employer_verified;

        -- === 运营字段 ===
        ALTER TABLE linggongs DROP COLUMN IF EXISTS traffic_weight;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS promotion_level;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS verified;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS picked;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS featured;

        -- === 配置/特征（JSONB） ===
        ALTER TABLE linggongs DROP COLUMN IF EXISTS requirements;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS images;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS welfare_tags;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS skill_tags;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS tags;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS features;

        -- === 视频/图片 ===
        ALTER TABLE linggongs DROP COLUMN IF EXISTS video_cover;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS video_url;

        -- === 风控 ===
        ALTER TABLE linggongs DROP COLUMN IF EXISTS risk_score;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS content_hash;

        -- === 互动统计 ===
        ALTER TABLE linggongs DROP COLUMN IF EXISTS last_applied_at;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS application_count;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS share_count;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS contact_count;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS fav_count;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS view_count;

        -- === 任务制相关 ===
        ALTER TABLE linggongs DROP COLUMN IF EXISTS completed_task_count;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS claimed_count;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS total_task_count;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS task_id;

        -- === 工作地点 ===
        ALTER TABLE linggongs DROP COLUMN IF EXISTS work_location_type;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS longitude;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS latitude;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS address;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS business_district;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS district;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS city;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS province;

        -- === 招募要求 ===
        ALTER TABLE linggongs DROP COLUMN IF EXISTS min_credit_score;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS need_id_card;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS need_health_cert;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS experience;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS education;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS max_age;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS min_age;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS need_gender;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS confirmed_count;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS applied_count;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS recruit_count;

        -- === 工作时间 ===
        ALTER TABLE linggongs DROP COLUMN IF EXISTS work_intensity;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS work_weekdays;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS work_time_end;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS work_time_start;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS work_hours;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS work_days;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS work_end_date;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS work_start_date;

        -- === 计费方式/薪资 ===
        ALTER TABLE linggongs DROP COLUMN IF EXISTS currency;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS settlement;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS salary_negotiable;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS salary_unit;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS salary_max;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS salary_min;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS billing_type;

        -- === 雇主关联 ===
        ALTER TABLE linggongs DROP COLUMN IF EXISTS contact_wechat;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS contact_phone;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS contact_name;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS company_name;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS employer_id;

        -- === 岗位类型/发布者类型 ===
        ALTER TABLE linggongs DROP COLUMN IF EXISTS publisher_type;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS linggong_type;

        -- === 基础信息扩展 ===
        ALTER TABLE linggongs DROP COLUMN IF EXISTS published_at;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS audit_reason;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS user_avatar;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS user_phone;
        ALTER TABLE linggongs DROP COLUMN IF EXISTS cover_image;
    END IF;
END $$;

-- ============================================================
-- 第四部分：移除表注释
-- ============================================================
COMMENT ON SCHEMA public IS 'public';
