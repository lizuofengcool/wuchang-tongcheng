-- ============================================================
-- job 招聘求职模块完整功能回滚脚本（与 006_job_full.sql 配对）
-- 按反向顺序 DROP 19 张子表 + 触发器 + 主表新增字段
-- 幂等：使用 IF EXISTS，可重复执行
-- 警告：执行后 job 模块全部扩展数据将丢失（主表基础字段保留）
-- ============================================================

-- ------------------------------------------------------------
-- 1. 先移除 19 张子表的 updated_at 触发器
-- ------------------------------------------------------------
DROP TRIGGER IF EXISTS trg_job_recommendations_updated ON job_recommendations;
DROP TRIGGER IF EXISTS trg_job_escrows_updated ON job_escrows;
DROP TRIGGER IF EXISTS trg_job_statistics_updated ON job_statistics;
DROP TRIGGER IF EXISTS trg_job_audit_rules_updated ON job_audit_rules;
DROP TRIGGER IF EXISTS trg_job_benefits_updated ON job_benefits;
DROP TRIGGER IF EXISTS trg_job_salary_ranges_updated ON job_salary_ranges;
DROP TRIGGER IF EXISTS trg_job_reviews_updated ON job_reviews;
DROP TRIGGER IF EXISTS trg_job_reports_updated ON job_reports;
DROP TRIGGER IF EXISTS trg_job_views_updated ON job_views;
DROP TRIGGER IF EXISTS trg_job_favorites_updated ON job_favorites;
DROP TRIGGER IF EXISTS trg_job_messages_updated ON job_messages;
DROP TRIGGER IF EXISTS trg_job_certifications_updated ON job_certifications;
DROP TRIGGER IF EXISTS trg_job_skills_updated ON job_skills;
DROP TRIGGER IF EXISTS trg_job_categories_updated ON job_categories;
DROP TRIGGER IF EXISTS trg_job_positions_updated ON job_positions;
DROP TRIGGER IF EXISTS trg_job_interviews_updated ON job_interviews;
DROP TRIGGER IF EXISTS trg_job_applications_updated ON job_applications;
DROP TRIGGER IF EXISTS trg_job_resumes_updated ON job_resumes;
DROP TRIGGER IF EXISTS trg_job_companies_updated ON job_companies;

-- ------------------------------------------------------------
-- 2. 按反向顺序 DROP 19 张子表（先 DROP 有外键依赖的子表）
--    依赖顺序：job_certifications → job_companies；job_interviews → job_applications
-- ------------------------------------------------------------
DROP TABLE IF EXISTS job_recommendations;
DROP TABLE IF EXISTS job_escrows;
DROP TABLE IF EXISTS job_statistics;
DROP TABLE IF EXISTS job_audit_rules;
DROP TABLE IF EXISTS job_benefits;
DROP TABLE IF EXISTS job_salary_ranges;
DROP TABLE IF EXISTS job_reviews;
DROP TABLE IF EXISTS job_reports;
DROP TABLE IF EXISTS job_views;
DROP TABLE IF EXISTS job_favorites;
DROP TABLE IF EXISTS job_messages;
DROP TABLE IF EXISTS job_certifications;
DROP TABLE IF EXISTS job_skills;
DROP TABLE IF EXISTS job_categories;
DROP TABLE IF EXISTS job_positions;
DROP TABLE IF EXISTS job_interviews;
DROP TABLE IF EXISTS job_applications;
DROP TABLE IF EXISTS job_resumes;
DROP TABLE IF EXISTS job_companies;

-- ------------------------------------------------------------
-- 3. 移除 jobs 主表新增字段（按反向顺序）
--    包装在 DO 块中，表不存在时跳过（幂等）
-- ------------------------------------------------------------
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'jobs') THEN
        -- 运营字段
        ALTER TABLE jobs DROP COLUMN IF EXISTS traffic_weight;
        ALTER TABLE jobs DROP COLUMN IF EXISTS promotion_level;
        ALTER TABLE jobs DROP COLUMN IF EXISTS verified;
        ALTER TABLE jobs DROP COLUMN IF EXISTS picked;
        ALTER TABLE jobs DROP COLUMN IF EXISTS featured;

        -- 视频支持
        ALTER TABLE jobs DROP COLUMN IF EXISTS video_cover;
        ALTER TABLE jobs DROP COLUMN IF EXISTS video_url;

        -- 风控
        ALTER TABLE jobs DROP COLUMN IF EXISTS same_job_id;
        ALTER TABLE jobs DROP COLUMN IF EXISTS risk_score;
        ALTER TABLE jobs DROP COLUMN IF EXISTS content_hash;

        -- 互动统计
        ALTER TABLE jobs DROP COLUMN IF EXISTS message_count;
        ALTER TABLE jobs DROP COLUMN IF EXISTS offer_count;
        ALTER TABLE jobs DROP COLUMN IF EXISTS interview_count;
        ALTER TABLE jobs DROP COLUMN IF EXISTS deliver_count;
        ALTER TABLE jobs DROP COLUMN IF EXISTS fav_count;
        ALTER TABLE jobs DROP COLUMN IF EXISTS view_count;

        -- 联系方式/期限
        ALTER TABLE jobs DROP COLUMN IF EXISTS need_health_check;
        ALTER TABLE jobs DROP COLUMN IF EXISTS need_bg_check;
        ALTER TABLE jobs DROP COLUMN IF EXISTS application_deadline;
        ALTER TABLE jobs DROP COLUMN IF EXISTS contact_wechat;
        ALTER TABLE jobs DROP COLUMN IF EXISTS contact_email;
        ALTER TABLE jobs DROP COLUMN IF EXISTS contact_phone;
        ALTER TABLE jobs DROP COLUMN IF EXISTS contact_name;

        -- 试用期/社保/公积金
        ALTER TABLE jobs DROP COLUMN IF EXISTS allow_remote;
        ALTER TABLE jobs DROP COLUMN IF EXISTS overtime_status;
        ALTER TABLE jobs DROP COLUMN IF EXISTS work_schedule;
        ALTER TABLE jobs DROP COLUMN IF EXISTS promotion_channels;
        ALTER TABLE jobs DROP COLUMN IF EXISTS allowances;
        ALTER TABLE jobs DROP COLUMN IF EXISTS has_housing_fund;
        ALTER TABLE jobs DROP COLUMN IF EXISTS has_social_insurance;
        ALTER TABLE jobs DROP COLUMN IF EXISTS probation_salary_ratio;
        ALTER TABLE jobs DROP COLUMN IF EXISTS probation_months;

        -- 应聘要求
        ALTER TABLE jobs DROP COLUMN IF EXISTS travel_frequency;
        ALTER TABLE jobs DROP COLUMN IF EXISTS certificate_requirement;
        ALTER TABLE jobs DROP COLUMN IF EXISTS language_requirement;
        ALTER TABLE jobs DROP COLUMN IF EXISTS major;
        ALTER TABLE jobs DROP COLUMN IF EXISTS gender_requirement;
        ALTER TABLE jobs DROP COLUMN IF EXISTS age_max;
        ALTER TABLE jobs DROP COLUMN IF EXISTS age_min;

        -- 招聘者/紧急/置顶
        ALTER TABLE jobs DROP COLUMN IF EXISTS top_expire;
        ALTER TABLE jobs DROP COLUMN IF EXISTS is_top;
        ALTER TABLE jobs DROP COLUMN IF EXISTS urgent_expire;
        ALTER TABLE jobs DROP COLUMN IF EXISTS is_urgent;
        ALTER TABLE jobs DROP COLUMN IF EXISTS recruiter_position;
        ALTER TABLE jobs DROP COLUMN IF EXISTS recruiter_avatar;
        ALTER TABLE jobs DROP COLUMN IF EXISTS recruiter_name;
        ALTER TABLE jobs DROP COLUMN IF EXISTS recruiter_id;

        -- 福利/技能/标签
        ALTER TABLE jobs DROP COLUMN IF EXISTS welfare_tags;
        ALTER TABLE jobs DROP COLUMN IF EXISTS tags;
        ALTER TABLE jobs DROP COLUMN IF EXISTS skills;
        ALTER TABLE jobs DROP COLUMN IF EXISTS benefits;

        -- 招聘类型与雇用方式
        ALTER TABLE jobs DROP COLUMN IF EXISTS company_id;
        ALTER TABLE jobs DROP COLUMN IF EXISTS category_id;
        ALTER TABLE jobs DROP COLUMN IF EXISTS position_template_id;
        ALTER TABLE jobs DROP COLUMN IF EXISTS department;
        ALTER TABLE jobs DROP COLUMN IF EXISTS hiring_count;
        ALTER TABLE jobs DROP COLUMN IF EXISTS employment_type;
        ALTER TABLE jobs DROP COLUMN IF EXISTS recruitment_type;

        -- 工作地点
        ALTER TABLE jobs DROP COLUMN IF EXISTS work_business_district;
        ALTER TABLE jobs DROP COLUMN IF EXISTS work_district;
        ALTER TABLE jobs DROP COLUMN IF EXISTS work_city;
        ALTER TABLE jobs DROP COLUMN IF EXISTS work_longitude;
        ALTER TABLE jobs DROP COLUMN IF EXISTS work_latitude;
        ALTER TABLE jobs DROP COLUMN IF EXISTS work_address;

        -- 学历/经验要求
        ALTER TABLE jobs DROP COLUMN IF EXISTS experience_text;
        ALTER TABLE jobs DROP COLUMN IF EXISTS work_year_max;
        ALTER TABLE jobs DROP COLUMN IF EXISTS work_year_min;
        ALTER TABLE jobs DROP COLUMN IF EXISTS education;

        -- 薪资相关
        ALTER TABLE jobs DROP COLUMN IF EXISTS show_salary;
        ALTER TABLE jobs DROP COLUMN IF EXISTS salary_range_id;
        ALTER TABLE jobs DROP COLUMN IF EXISTS salary_negotiable;
        ALTER TABLE jobs DROP COLUMN IF EXISTS salary_monthly;
        ALTER TABLE jobs DROP COLUMN IF EXISTS salary_unit;
        ALTER TABLE jobs DROP COLUMN IF EXISTS salary_max;
        ALTER TABLE jobs DROP COLUMN IF EXISTS salary_min;
    END IF;
END $$;
