-- ============================================================
-- job 招聘求职模块完整功能迁移脚本（v3.2.1）
-- 对标：BOSS直聘 / 拉勾 / 58招聘 / 智联 / 前程无忧
--
-- 内容：
--   1. ALTER TABLE jobs 主表新增 35+ 字段（薪资范围/学历/经验/工作地/福利/紧急/招聘类型等）
--   2. CREATE 19 张子表（job_ 前缀，依据数据库分表前缀规范 v1.0.0）
--   3. 索引、外键、触发器、注释
--   4. 全幂等：CREATE TABLE IF NOT EXISTS / ALTER TABLE ADD COLUMN IF NOT EXISTS
-- 依赖：001_p0_baseline.sql（update_updated_at_column 函数已创建）
-- ============================================================

-- ============================================================
-- 第一部分：扩展 jobs 主表（保持现有表名兼容已发布数据）
-- 注意：jobs 表由 GORM AutoMigrate 在应用启动时创建
--      本迁移对 jobs 的所有 ALTER 操作包装在 DO 块中，
--      若表不存在则跳过（待应用启动后再执行一次本迁移即可补齐字段）
-- ============================================================
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'jobs') THEN
        -- === 薪资相关 ===
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS salary_min DECIMAL(12,2) NOT NULL DEFAULT 0;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS salary_max DECIMAL(12,2) NOT NULL DEFAULT 0;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS salary_unit VARCHAR(16) NOT NULL DEFAULT 'month';
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS salary_monthly DECIMAL(12,2) NOT NULL DEFAULT 0;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS salary_negotiable BOOLEAN NOT NULL DEFAULT TRUE;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS salary_range_id BIGINT;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS show_salary BOOLEAN NOT NULL DEFAULT TRUE;
        CREATE INDEX IF NOT EXISTS idx_jobs_salary_min ON jobs(salary_min);
        CREATE INDEX IF NOT EXISTS idx_jobs_salary_max ON jobs(salary_max);
        CREATE INDEX IF NOT EXISTS idx_jobs_salary_range_id ON jobs(salary_range_id);

        -- === 学历/经验要求 ===
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS education VARCHAR(32) NOT NULL DEFAULT 'unlimited';
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS work_year_min INT NOT NULL DEFAULT 0;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS work_year_max INT NOT NULL DEFAULT 0;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS experience_text VARCHAR(64) NOT NULL DEFAULT '';
        CREATE INDEX IF NOT EXISTS idx_jobs_education ON jobs(education);
        CREATE INDEX IF NOT EXISTS idx_jobs_work_year_min ON jobs(work_year_min);

        -- === 工作地点 ===
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS work_address VARCHAR(255) NOT NULL DEFAULT '';
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS work_latitude DECIMAL(10,7) NOT NULL DEFAULT 0;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS work_longitude DECIMAL(10,7) NOT NULL DEFAULT 0;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS work_city VARCHAR(64) NOT NULL DEFAULT '';
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS work_district VARCHAR(64) NOT NULL DEFAULT '';
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS work_business_district VARCHAR(128) NOT NULL DEFAULT '';
        CREATE INDEX IF NOT EXISTS idx_jobs_work_city ON jobs(work_city);

        -- === 招聘类型与雇用方式 ===
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS recruitment_type VARCHAR(32) NOT NULL DEFAULT 'full_time';
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS employment_type VARCHAR(32) NOT NULL DEFAULT 'regular';
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS hiring_count INT NOT NULL DEFAULT 1;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS department VARCHAR(64) NOT NULL DEFAULT '';
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS position_template_id BIGINT;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS category_id BIGINT;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS company_id BIGINT;
        CREATE INDEX IF NOT EXISTS idx_jobs_recruitment_type ON jobs(recruitment_type);
        CREATE INDEX IF NOT EXISTS idx_jobs_employment_type ON jobs(employment_type);
        CREATE INDEX IF NOT EXISTS idx_jobs_category_id ON jobs(category_id);
        CREATE INDEX IF NOT EXISTS idx_jobs_company_id ON jobs(company_id);
        CREATE INDEX IF NOT EXISTS idx_jobs_position_template_id ON jobs(position_template_id);

        -- === 福利/技能/标签 ===
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS benefits JSONB;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS skills JSONB;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS tags JSONB;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS welfare_tags JSONB;
        CREATE INDEX IF NOT EXISTS idx_jobs_benefits ON jobs USING GIN(benefits);
        CREATE INDEX IF NOT EXISTS idx_jobs_tags ON jobs USING GIN(tags);

        -- === 招聘者/紧急/置顶 ===
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS recruiter_id BIGINT NOT NULL DEFAULT 0;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS recruiter_name VARCHAR(50) NOT NULL DEFAULT '';
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS recruiter_avatar VARCHAR(255) NOT NULL DEFAULT '';
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS recruiter_position VARCHAR(64) NOT NULL DEFAULT '';
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS is_urgent BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS urgent_expire TIMESTAMPTZ;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS is_top BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS top_expire TIMESTAMPTZ;
        CREATE INDEX IF NOT EXISTS idx_jobs_recruiter_id ON jobs(recruiter_id);
        CREATE INDEX IF NOT EXISTS idx_jobs_is_urgent ON jobs(is_urgent) WHERE is_urgent = TRUE;
        CREATE INDEX IF NOT EXISTS idx_jobs_is_top ON jobs(is_top) WHERE is_top = TRUE;
        CREATE INDEX IF NOT EXISTS idx_jobs_urgent_expire ON jobs(urgent_expire);

        -- === 应聘要求 ===
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS age_min INT NOT NULL DEFAULT 0;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS age_max INT NOT NULL DEFAULT 0;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS gender_requirement VARCHAR(16) NOT NULL DEFAULT 'unlimited';
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS major VARCHAR(128) NOT NULL DEFAULT '';
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS language_requirement VARCHAR(64) NOT NULL DEFAULT '';
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS certificate_requirement VARCHAR(255) NOT NULL DEFAULT '';
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS travel_frequency VARCHAR(16) NOT NULL DEFAULT 'none';

        -- === 试用期/社保/公积金 ===
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS probation_months INT NOT NULL DEFAULT 0;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS probation_salary_ratio DECIMAL(3,2) NOT NULL DEFAULT 1.00;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS has_social_insurance BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS has_housing_fund BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS allowances JSONB;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS promotion_channels JSONB;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS work_schedule VARCHAR(64) NOT NULL DEFAULT '';
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS overtime_status VARCHAR(16) NOT NULL DEFAULT 'unknown';
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS allow_remote BOOLEAN NOT NULL DEFAULT FALSE;

        -- === 联系方式/期限 ===
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS contact_name VARCHAR(50) NOT NULL DEFAULT '';
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS contact_phone VARCHAR(20) NOT NULL DEFAULT '';
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS contact_email VARCHAR(128) NOT NULL DEFAULT '';
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS contact_wechat VARCHAR(50) NOT NULL DEFAULT '';
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS application_deadline TIMESTAMPTZ;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS need_bg_check BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS need_health_check BOOLEAN NOT NULL DEFAULT FALSE;
        CREATE INDEX IF NOT EXISTS idx_jobs_application_deadline ON jobs(application_deadline);

        -- === 互动统计 ===
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS view_count INT NOT NULL DEFAULT 0;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS fav_count INT NOT NULL DEFAULT 0;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS deliver_count INT NOT NULL DEFAULT 0;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS interview_count INT NOT NULL DEFAULT 0;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS offer_count INT NOT NULL DEFAULT 0;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS message_count INT NOT NULL DEFAULT 0;

        -- === 风控 ===
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS content_hash VARCHAR(64);
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS risk_score INT NOT NULL DEFAULT 0;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS same_job_id VARCHAR(64);
        CREATE INDEX IF NOT EXISTS idx_jobs_content_hash ON jobs(content_hash);
        CREATE INDEX IF NOT EXISTS idx_jobs_risk_score ON jobs(risk_score);
        CREATE INDEX IF NOT EXISTS idx_jobs_same_job_id ON jobs(same_job_id);

        -- === 视频支持 ===
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS video_url VARCHAR(255);
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS video_cover VARCHAR(255);

        -- === 运营字段 ===
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS featured BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS picked BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS verified BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS promotion_level INT NOT NULL DEFAULT 0;
        ALTER TABLE jobs ADD COLUMN IF NOT EXISTS traffic_weight DECIMAL(3,2) NOT NULL DEFAULT 1.00;
        CREATE INDEX IF NOT EXISTS idx_jobs_featured ON jobs(featured) WHERE featured = TRUE;
        CREATE INDEX IF NOT EXISTS idx_jobs_picked ON jobs(picked) WHERE picked = TRUE;
        CREATE INDEX IF NOT EXISTS idx_jobs_verified ON jobs(verified) WHERE verified = TRUE;

        -- 字段注释
        COMMENT ON COLUMN jobs.salary_min IS '薪资下限（单位见 salary_unit）';
        COMMENT ON COLUMN jobs.salary_max IS '薪资上限（单位见 salary_unit）';
        COMMENT ON COLUMN jobs.salary_unit IS '薪资单位：month/year/hour/day';
        COMMENT ON COLUMN jobs.salary_monthly IS '月薪展示值（salary_min/max 折算到月）';
        COMMENT ON COLUMN jobs.salary_negotiable IS '是否面议';
        COMMENT ON COLUMN jobs.salary_range_id IS '关联薪资范围配置 ID';
        COMMENT ON COLUMN jobs.show_salary IS '是否公开薪资';
        COMMENT ON COLUMN jobs.education IS '学历要求：unlimited/junior_high/high_school/college/bachelor/master/phd';
        COMMENT ON COLUMN jobs.work_year_min IS '经验下限（年），0=不限';
        COMMENT ON COLUMN jobs.work_year_max IS '经验上限（年），0=不限';
        COMMENT ON COLUMN jobs.experience_text IS '经验要求展示文本';
        COMMENT ON COLUMN jobs.work_address IS '工作地点详细地址';
        COMMENT ON COLUMN jobs.work_latitude IS '工作地点纬度';
        COMMENT ON COLUMN jobs.work_longitude IS '工作地点经度';
        COMMENT ON COLUMN jobs.work_city IS '工作城市';
        COMMENT ON COLUMN jobs.work_district IS '工作行政区';
        COMMENT ON COLUMN jobs.work_business_district IS '工作商圈';
        COMMENT ON COLUMN jobs.recruitment_type IS '招聘类型：full_time/part_time/internship/temp/outsource/gig';
        COMMENT ON COLUMN jobs.employment_type IS '雇佣方式：regular/labor_dispatch/outsourcing/freelance';
        COMMENT ON COLUMN jobs.hiring_count IS '招聘人数';
        COMMENT ON COLUMN jobs.department IS '所属部门';
        COMMENT ON COLUMN jobs.position_template_id IS '职位模板 ID（关联 job_positions）';
        COMMENT ON COLUMN jobs.category_id IS '职位分类 ID（关联 job_categories）';
        COMMENT ON COLUMN jobs.company_id IS '所属公司 ID（关联 job_companies）';
        COMMENT ON COLUMN jobs.benefits IS '福利标签 ID 数组 JSON';
        COMMENT ON COLUMN jobs.skills IS '技能要求 ID 数组 JSON';
        COMMENT ON COLUMN jobs.tags IS '职位标签数组 JSON';
        COMMENT ON COLUMN jobs.welfare_tags IS '福利文案标签 JSON';
        COMMENT ON COLUMN jobs.recruiter_id IS '招聘者用户 ID';
        COMMENT ON COLUMN jobs.recruiter_name IS '招聘者昵称';
        COMMENT ON COLUMN jobs.recruiter_avatar IS '招聘者头像';
        COMMENT ON COLUMN jobs.recruiter_position IS '招聘者职位';
        COMMENT ON COLUMN jobs.is_urgent IS '是否紧急招聘';
        COMMENT ON COLUMN jobs.urgent_expire IS '紧急招聘到期时间';
        COMMENT ON COLUMN jobs.is_top IS '是否置顶';
        COMMENT ON COLUMN jobs.top_expire IS '置顶到期时间';
        COMMENT ON COLUMN jobs.age_min IS '年龄下限，0=不限';
        COMMENT ON COLUMN jobs.age_max IS '年龄上限，0=不限';
        COMMENT ON COLUMN jobs.gender_requirement IS '性别要求：unlimited/male/female';
        COMMENT ON COLUMN jobs.major IS '专业要求';
        COMMENT ON COLUMN jobs.language_requirement IS '语言要求';
        COMMENT ON COLUMN jobs.certificate_requirement IS '证书要求';
        COMMENT ON COLUMN jobs.travel_frequency IS '出差频率：none/occasional/frequent';
        COMMENT ON COLUMN jobs.probation_months IS '试用期月数';
        COMMENT ON COLUMN jobs.probation_salary_ratio IS '试用期薪资比例 0.00-1.00';
        COMMENT ON COLUMN jobs.has_social_insurance IS '是否五险';
        COMMENT ON COLUMN jobs.has_housing_fund IS '是否一金';
        COMMENT ON COLUMN jobs.allowances IS '补贴 JSON';
        COMMENT ON COLUMN jobs.promotion_channels IS '晋升通道 JSON';
        COMMENT ON COLUMN jobs.work_schedule IS '工作时间';
        COMMENT ON COLUMN jobs.overtime_status IS '加班情况：no/occasional/frequent/unknown';
        COMMENT ON COLUMN jobs.allow_remote IS '是否支持远程';
        COMMENT ON COLUMN jobs.contact_name IS '联系人姓名';
        COMMENT ON COLUMN jobs.contact_phone IS '联系电话';
        COMMENT ON COLUMN jobs.contact_email IS '联系邮箱';
        COMMENT ON COLUMN jobs.contact_wechat IS '联系微信';
        COMMENT ON COLUMN jobs.application_deadline IS '招聘截止时间';
        COMMENT ON COLUMN jobs.need_bg_check IS '是否需要背景调查';
        COMMENT ON COLUMN jobs.need_health_check IS '是否需要体检';
        COMMENT ON COLUMN jobs.view_count IS '浏览数';
        COMMENT ON COLUMN jobs.fav_count IS '收藏数';
        COMMENT ON COLUMN jobs.deliver_count IS '投递数';
        COMMENT ON COLUMN jobs.interview_count IS '面试数';
        COMMENT ON COLUMN jobs.offer_count IS 'Offer 数';
        COMMENT ON COLUMN jobs.message_count IS '消息数';
        COMMENT ON COLUMN jobs.content_hash IS '图文指纹（MD5/SHA256）';
        COMMENT ON COLUMN jobs.risk_score IS '风险评分 0-100，<30 限制发布';
        COMMENT ON COLUMN jobs.same_job_id IS '同职位识别 ID';
        COMMENT ON COLUMN jobs.video_url IS '视频 URL';
        COMMENT ON COLUMN jobs.video_cover IS '视频封面';
        COMMENT ON COLUMN jobs.featured IS '精选推荐';
        COMMENT ON COLUMN jobs.picked IS '运营甄选';
        COMMENT ON COLUMN jobs.verified IS '官方验真';
        COMMENT ON COLUMN jobs.promotion_level IS '推广等级 0-10';
        COMMENT ON COLUMN jobs.traffic_weight IS '流量权重 0.00-9.99';
    END IF;
END $$;

-- ============================================================
-- 第二部分：19 张子表
-- ============================================================

-- ------------------------------------------------------------
-- 1. job_companies 公司信息表
--    对标 BOSS直聘：公司主页/Logo/Banner/规模/行业/营业执照
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS job_companies (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id BIGINT NOT NULL,
    name VARCHAR(128) NOT NULL,
    short_name VARCHAR(64) NOT NULL DEFAULT '',
    logo VARCHAR(255) NOT NULL DEFAULT '',
    banner VARCHAR(255) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    industry VARCHAR(64) NOT NULL DEFAULT '',
    scale VARCHAR(32) NOT NULL DEFAULT '',
    level INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 0,
    contact_name VARCHAR(50) NOT NULL DEFAULT '',
    contact_phone VARCHAR(20) NOT NULL DEFAULT '',
    contact_email VARCHAR(128) NOT NULL DEFAULT '',
    contact_wechat VARCHAR(50) NOT NULL DEFAULT '',
    address VARCHAR(500) NOT NULL DEFAULT '',
    latitude DECIMAL(10,7) NOT NULL DEFAULT 0,
    longitude DECIMAL(10,7) NOT NULL DEFAULT 0,
    business_license VARCHAR(255) NOT NULL DEFAULT '',
    license_no VARCHAR(64) NOT NULL DEFAULT '',
    id_card_front VARCHAR(255) NOT NULL DEFAULT '',
    id_card_back VARCHAR(255) NOT NULL DEFAULT '',
    legal_person VARCHAR(50) NOT NULL DEFAULT '',
    legal_person_id_card VARCHAR(32) NOT NULL DEFAULT '',
    registered_capital DECIMAL(14,2) NOT NULL DEFAULT 0,
    founded_at DATE,
    website VARCHAR(255) NOT NULL DEFAULT '',
    verified_at TIMESTAMPTZ,
    approved_at TIMESTAMPTZ,
    rejected_reason VARCHAR(500) NOT NULL DEFAULT '',
    closed_at TIMESTAMPTZ,
    follower_count INT NOT NULL DEFAULT 0,
    job_count INT NOT NULL DEFAULT 0,
    employee_count INT NOT NULL DEFAULT 0,
    active_job_count INT NOT NULL DEFAULT 0,
    total_hired_count INT NOT NULL DEFAULT 0,
    good_rate DECIMAL(5,2) NOT NULL DEFAULT 100.00,
    deposit DECIMAL(12,2) NOT NULL DEFAULT 0,
    tags JSONB,
    CONSTRAINT uniq_job_companies_user_id UNIQUE (user_id)
);
CREATE INDEX IF NOT EXISTS idx_job_companies_region_id ON job_companies(region_id);
CREATE INDEX IF NOT EXISTS idx_job_companies_name ON job_companies(name);
CREATE INDEX IF NOT EXISTS idx_job_companies_industry ON job_companies(industry);
CREATE INDEX IF NOT EXISTS idx_job_companies_scale ON job_companies(scale);
CREATE INDEX IF NOT EXISTS idx_job_companies_level ON job_companies(level);
CREATE INDEX IF NOT EXISTS idx_job_companies_status ON job_companies(status);
CREATE INDEX IF NOT EXISTS idx_job_companies_license_no ON job_companies(license_no);
CREATE INDEX IF NOT EXISTS idx_job_companies_verified_at ON job_companies(verified_at);
CREATE INDEX IF NOT EXISTS idx_job_companies_approved_at ON job_companies(approved_at);
CREATE INDEX IF NOT EXISTS idx_job_companies_closed_at ON job_companies(closed_at);
CREATE INDEX IF NOT EXISTS idx_job_companies_deleted_at ON job_companies(deleted_at);
COMMENT ON TABLE job_companies IS '公司信息表（公司主页/Logo/Banner/规模/行业/营业执照/法人）';

-- ------------------------------------------------------------
-- 2. job_resumes 简历表
--    对标 BOSS直聘：教育/工作/项目/技能/期望
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS job_resumes (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id BIGINT NOT NULL,
    name VARCHAR(50) NOT NULL DEFAULT '',
    gender VARCHAR(16) NOT NULL DEFAULT 'unlimited',
    birth_date DATE,
    phone VARCHAR(20) NOT NULL DEFAULT '',
    email VARCHAR(128) NOT NULL DEFAULT '',
    avatar VARCHAR(255) NOT NULL DEFAULT '',
    education_level VARCHAR(32) NOT NULL DEFAULT 'unlimited',
    school VARCHAR(128) NOT NULL DEFAULT '',
    major VARCHAR(128) NOT NULL DEFAULT '',
    graduate_date DATE,
    work_years INT NOT NULL DEFAULT 0,
    current_company VARCHAR(128) NOT NULL DEFAULT '',
    current_position VARCHAR(64) NOT NULL DEFAULT '',
    current_salary DECIMAL(12,2) NOT NULL DEFAULT 0,
    expect_salary_min DECIMAL(12,2) NOT NULL DEFAULT 0,
    expect_salary_max DECIMAL(12,2) NOT NULL DEFAULT 0,
    expect_city VARCHAR(64) NOT NULL DEFAULT '',
    expect_position VARCHAR(128) NOT NULL DEFAULT '',
    expect_industry VARCHAR(64) NOT NULL DEFAULT '',
    expect_job_type VARCHAR(32) NOT NULL DEFAULT 'full_time',
    expect_employment_type VARCHAR(32) NOT NULL DEFAULT 'regular',
    status INT NOT NULL DEFAULT 1,
    completeness INT NOT NULL DEFAULT 0,
    is_public BOOLEAN NOT NULL DEFAULT TRUE,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    view_count INT NOT NULL DEFAULT 0,
    deliver_count INT NOT NULL DEFAULT 0,
    interview_count INT NOT NULL DEFAULT 0,
    offer_count INT NOT NULL DEFAULT 0,
    self_introduction TEXT NOT NULL DEFAULT '',
    advantage TEXT NOT NULL DEFAULT '',
    disadvantage TEXT NOT NULL DEFAULT '',
    attachments JSONB,
    educations JSONB,
    work_experiences JSONB,
    projects JSONB,
    skills JSONB,
    certificates JSONB,
    languages JSONB,
    tags JSONB,
    CONSTRAINT uniq_job_resumes_user_default UNIQUE (user_id, is_default)
);
CREATE INDEX IF NOT EXISTS idx_job_resumes_region_id ON job_resumes(region_id);
CREATE INDEX IF NOT EXISTS idx_job_resumes_user_id ON job_resumes(user_id);
CREATE INDEX IF NOT EXISTS idx_job_resumes_education_level ON job_resumes(education_level);
CREATE INDEX IF NOT EXISTS idx_job_resumes_work_years ON job_resumes(work_years);
CREATE INDEX IF NOT EXISTS idx_job_resumes_expect_city ON job_resumes(expect_city);
CREATE INDEX IF NOT EXISTS idx_job_resumes_expect_position ON job_resumes(expect_position);
CREATE INDEX IF NOT EXISTS idx_job_resumes_expect_job_type ON job_resumes(expect_job_type);
CREATE INDEX IF NOT EXISTS idx_job_resumes_status ON job_resumes(status);
CREATE INDEX IF NOT EXISTS idx_job_resumes_is_public ON job_resumes(is_public) WHERE is_public = TRUE;
CREATE INDEX IF NOT EXISTS idx_job_resumes_is_default ON job_resumes(is_default) WHERE is_default = TRUE;
CREATE INDEX IF NOT EXISTS idx_job_resumes_skills ON job_resumes USING GIN(skills);
CREATE INDEX IF NOT EXISTS idx_job_resumes_deleted_at ON job_resumes(deleted_at);
COMMENT ON TABLE job_resumes IS '简历表（教育/工作/项目/技能/期望 JSONB + 基础信息）';

-- ------------------------------------------------------------
-- 3. job_applications 投递记录表
--    对标 BOSS直聘：投递状态机 + 简历快照 + 附件
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS job_applications (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    application_no VARCHAR(64) NOT NULL,
    job_id BIGINT NOT NULL,
    resume_id BIGINT NOT NULL,
    applicant_id BIGINT NOT NULL,
    recruiter_id BIGINT NOT NULL,
    company_id BIGINT NOT NULL DEFAULT 0,
    position_name VARCHAR(128) NOT NULL DEFAULT '',
    position_snapshot JSONB,
    resume_snapshot JSONB,
    status INT NOT NULL DEFAULT 0,
    source VARCHAR(32) NOT NULL DEFAULT 'proactive',
    cover_letter TEXT NOT NULL DEFAULT '',
    attachments JSONB,
    read_at TIMESTAMPTZ,
    replied_at TIMESTAMPTZ,
    interview_count INT NOT NULL DEFAULT 0,
    offer_at TIMESTAMPTZ,
    offer_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    rejected_reason VARCHAR(500) NOT NULL DEFAULT '',
    rejected_at TIMESTAMPTZ,
    withdrawn_at TIMESTAMPTZ,
    withdrawn_reason VARCHAR(500) NOT NULL DEFAULT '',
    completed_at TIMESTAMPTZ,
    expired_at TIMESTAMPTZ,
    sla_deadline TIMESTAMPTZ,
    CONSTRAINT uniq_job_applications_no UNIQUE (application_no)
);
CREATE INDEX IF NOT EXISTS idx_job_applications_region_id ON job_applications(region_id);
CREATE INDEX IF NOT EXISTS idx_job_applications_job_id ON job_applications(job_id);
CREATE INDEX IF NOT EXISTS idx_job_applications_resume_id ON job_applications(resume_id);
CREATE INDEX IF NOT EXISTS idx_job_applications_applicant ON job_applications(applicant_id, status);
CREATE INDEX IF NOT EXISTS idx_job_applications_recruiter ON job_applications(recruiter_id, status);
CREATE INDEX IF NOT EXISTS idx_job_applications_company_id ON job_applications(company_id);
CREATE INDEX IF NOT EXISTS idx_job_applications_status ON job_applications(status);
CREATE INDEX IF NOT EXISTS idx_job_applications_source ON job_applications(source);
CREATE INDEX IF NOT EXISTS idx_job_applications_read_at ON job_applications(read_at);
CREATE INDEX IF NOT EXISTS idx_job_applications_replied_at ON job_applications(replied_at);
CREATE INDEX IF NOT EXISTS idx_job_applications_offer_at ON job_applications(offer_at);
CREATE INDEX IF NOT EXISTS idx_job_applications_sla_deadline ON job_applications(sla_deadline);
CREATE INDEX IF NOT EXISTS idx_job_applications_deleted_at ON job_applications(deleted_at);
COMMENT ON TABLE job_applications IS '投递记录表（职位/简历/状态机/快照/SLA）';

-- ------------------------------------------------------------
-- 4. job_interviews 面试邀约表
--    对标 BOSS直聘：在线/线下面试 + 多轮 + 面试反馈
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS job_interviews (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    interview_no VARCHAR(64) NOT NULL,
    application_id BIGINT NOT NULL,
    job_id BIGINT NOT NULL,
    applicant_id BIGINT NOT NULL,
    recruiter_id BIGINT NOT NULL,
    company_id BIGINT NOT NULL DEFAULT 0,
    round INT NOT NULL DEFAULT 1,
    interview_type VARCHAR(32) NOT NULL DEFAULT 'onsite',
    scheduled_at TIMESTAMPTZ,
    duration_minutes INT NOT NULL DEFAULT 60,
    location VARCHAR(255) NOT NULL DEFAULT '',
    online_url VARCHAR(500) NOT NULL DEFAULT '',
    online_password VARCHAR(64) NOT NULL DEFAULT '',
    interviewer_name VARCHAR(50) NOT NULL DEFAULT '',
    interviewer_position VARCHAR(64) NOT NULL DEFAULT '',
    contact_phone VARCHAR(20) NOT NULL DEFAULT '',
    status INT NOT NULL DEFAULT 0,
    result VARCHAR(32) NOT NULL DEFAULT 'pending',
    feedback TEXT NOT NULL DEFAULT '',
    rating INT NOT NULL DEFAULT 0,
    salary_offered DECIMAL(12,2) NOT NULL DEFAULT 0,
    position_offered VARCHAR(128) NOT NULL DEFAULT '',
    entry_date DATE,
    attachments JSONB,
    confirmed_at TIMESTAMPTZ,
    attended_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    canceled_at TIMESTAMPTZ,
    canceled_reason VARCHAR(500) NOT NULL DEFAULT '',
    CONSTRAINT uniq_job_interviews_no UNIQUE (interview_no)
);
CREATE INDEX IF NOT EXISTS idx_job_interviews_region_id ON job_interviews(region_id);
CREATE INDEX IF NOT EXISTS idx_job_interviews_application_id ON job_interviews(application_id);
CREATE INDEX IF NOT EXISTS idx_job_interviews_job_id ON job_interviews(job_id);
CREATE INDEX IF NOT EXISTS idx_job_interviews_applicant ON job_interviews(applicant_id, status);
CREATE INDEX IF NOT EXISTS idx_job_interviews_recruiter ON job_interviews(recruiter_id, status);
CREATE INDEX IF NOT EXISTS idx_job_interviews_company_id ON job_interviews(company_id);
CREATE INDEX IF NOT EXISTS idx_job_interviews_interview_type ON job_interviews(interview_type);
CREATE INDEX IF NOT EXISTS idx_job_interviews_status ON job_interviews(status);
CREATE INDEX IF NOT EXISTS idx_job_interviews_result ON job_interviews(result);
CREATE INDEX IF NOT EXISTS idx_job_interviews_scheduled_at ON job_interviews(scheduled_at);
CREATE INDEX IF NOT EXISTS idx_job_interviews_completed_at ON job_interviews(completed_at);
CREATE INDEX IF NOT EXISTS idx_job_interviews_deleted_at ON job_interviews(deleted_at);
-- 外键：依赖 job_applications 已创建
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_job_interviews_application') THEN
        ALTER TABLE job_interviews ADD CONSTRAINT fk_job_interviews_application
            FOREIGN KEY (application_id) REFERENCES job_applications(id) ON DELETE CASCADE;
    END IF;
END $$;
COMMENT ON TABLE job_interviews IS '面试邀约表（多轮/在线线下/反馈/Offer）';

-- ------------------------------------------------------------
-- 5. job_positions 职位模板表
--    对标 BOSS直聘：标准职位库 + 默认配置
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS job_positions (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    name VARCHAR(128) NOT NULL,
    code VARCHAR(64) NOT NULL DEFAULT '',
    category_id BIGINT NOT NULL DEFAULT 0,
    department VARCHAR(64) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    requirements TEXT NOT NULL DEFAULT '',
    responsibilities TEXT NOT NULL DEFAULT '',
    default_salary_min DECIMAL(12,2) NOT NULL DEFAULT 0,
    default_salary_max DECIMAL(12,2) NOT NULL DEFAULT 0,
    default_education VARCHAR(32) NOT NULL DEFAULT 'unlimited',
    default_work_year_min INT NOT NULL DEFAULT 0,
    default_work_year_max INT NOT NULL DEFAULT 0,
    default_recruitment_type VARCHAR(32) NOT NULL DEFAULT 'full_time',
    default_benefits JSONB,
    default_skills JSONB,
    status INT NOT NULL DEFAULT 1,
    sort INT NOT NULL DEFAULT 0,
    use_count INT NOT NULL DEFAULT 0,
    creator_id BIGINT NOT NULL DEFAULT 0,
    CONSTRAINT uniq_job_positions_name_code UNIQUE (name, code)
);
CREATE INDEX IF NOT EXISTS idx_job_positions_category_id ON job_positions(category_id);
CREATE INDEX IF NOT EXISTS idx_job_positions_status ON job_positions(status);
CREATE INDEX IF NOT EXISTS idx_job_positions_sort ON job_positions(sort);
CREATE INDEX IF NOT EXISTS idx_job_positions_creator_id ON job_positions(creator_id);
CREATE INDEX IF NOT EXISTS idx_job_positions_deleted_at ON job_positions(deleted_at);
COMMENT ON TABLE job_positions IS '职位模板表（标准职位库/默认配置）';

-- ------------------------------------------------------------
-- 6. job_categories 职位分类表
--    对标 BOSS直聘：互联网/金融/制造/教育/医疗等
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS job_categories (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    name VARCHAR(64) NOT NULL,
    code VARCHAR(64) NOT NULL,
    parent_id BIGINT NOT NULL DEFAULT 0,
    level INT NOT NULL DEFAULT 1,
    icon VARCHAR(64) NOT NULL DEFAULT '',
    color VARCHAR(16) NOT NULL DEFAULT '#409EFF',
    description VARCHAR(500) NOT NULL DEFAULT '',
    sort INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 1,
    job_count INT NOT NULL DEFAULT 0,
    CONSTRAINT uniq_job_categories_code UNIQUE (code)
);
CREATE INDEX IF NOT EXISTS idx_job_categories_parent_id ON job_categories(parent_id);
CREATE INDEX IF NOT EXISTS idx_job_categories_level ON job_categories(level);
CREATE INDEX IF NOT EXISTS idx_job_categories_status ON job_categories(status);
CREATE INDEX IF NOT EXISTS idx_job_categories_sort ON job_categories(sort);
CREATE INDEX IF NOT EXISTS idx_job_categories_deleted_at ON job_categories(deleted_at);
COMMENT ON TABLE job_categories IS '职位分类表（互联网/金融/制造/教育/医疗等）';

-- ------------------------------------------------------------
-- 7. job_skills 技能标签表
--    对标 BOSS直聘：Java/Python/PM/UI设计 + 技术层级
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS job_skills (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    name VARCHAR(64) NOT NULL,
    code VARCHAR(64) NOT NULL DEFAULT '',
    category VARCHAR(32) NOT NULL DEFAULT 'technical',
    description VARCHAR(500) NOT NULL DEFAULT '',
    icon VARCHAR(64) NOT NULL DEFAULT '',
    color VARCHAR(16) NOT NULL DEFAULT '#409EFF',
    status INT NOT NULL DEFAULT 1,
    sort INT NOT NULL DEFAULT 0,
    use_count INT NOT NULL DEFAULT 0,
    is_hot BOOLEAN NOT NULL DEFAULT FALSE,
    creator_id BIGINT NOT NULL DEFAULT 0,
    CONSTRAINT uniq_job_skills_name_category UNIQUE (name, category)
);
CREATE INDEX IF NOT EXISTS idx_job_skills_category ON job_skills(category);
CREATE INDEX IF NOT EXISTS idx_job_skills_status ON job_skills(status);
CREATE INDEX IF NOT EXISTS idx_job_skills_is_hot ON job_skills(is_hot) WHERE is_hot = TRUE;
CREATE INDEX IF NOT EXISTS idx_job_skills_sort ON job_skills(sort);
CREATE INDEX IF NOT EXISTS idx_job_skills_creator_id ON job_skills(creator_id);
CREATE INDEX IF NOT EXISTS idx_job_skills_deleted_at ON job_skills(deleted_at);
COMMENT ON TABLE job_skills IS '技能标签表（Java/Python/PM/UI 设计/软技能/语言等）';

-- ------------------------------------------------------------
-- 8. job_certifications 企业认证表
--    对标 BOSS直聘：营业执照/法人/认证状态
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS job_certifications (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    company_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    cert_type VARCHAR(32) NOT NULL DEFAULT 'business_license',
    cert_no VARCHAR(64) NOT NULL DEFAULT '',
    cert_name VARCHAR(128) NOT NULL DEFAULT '',
    cert_image VARCHAR(255) NOT NULL DEFAULT '',
    legal_person VARCHAR(50) NOT NULL DEFAULT '',
    legal_person_id_card VARCHAR(32) NOT NULL DEFAULT '',
    registered_capital DECIMAL(14,2) NOT NULL DEFAULT 0,
    business_scope TEXT NOT NULL DEFAULT '',
    valid_from DATE,
    valid_to DATE,
    status INT NOT NULL DEFAULT 0,
    verified_at TIMESTAMPTZ,
    verifier_id BIGINT NOT NULL DEFAULT 0,
    verifier_name VARCHAR(50) NOT NULL DEFAULT '',
    reject_reason VARCHAR(500) NOT NULL DEFAULT '',
    expired_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_job_certifications_company_id ON job_certifications(company_id);
CREATE INDEX IF NOT EXISTS idx_job_certifications_user_id ON job_certifications(user_id);
CREATE INDEX IF NOT EXISTS idx_job_certifications_cert_type ON job_certifications(cert_type);
CREATE INDEX IF NOT EXISTS idx_job_certifications_status ON job_certifications(status);
CREATE INDEX IF NOT EXISTS idx_job_certifications_verified_at ON job_certifications(verified_at);
CREATE INDEX IF NOT EXISTS idx_job_certifications_verifier_id ON job_certifications(verifier_id);
CREATE INDEX IF NOT EXISTS idx_job_certifications_expired_at ON job_certifications(expired_at);
CREATE INDEX IF NOT EXISTS idx_job_certifications_deleted_at ON job_certifications(deleted_at);
-- 外键：依赖 job_companies 已创建
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_job_certifications_company') THEN
        ALTER TABLE job_certifications ADD CONSTRAINT fk_job_certifications_company
            FOREIGN KEY (company_id) REFERENCES job_companies(id) ON DELETE CASCADE;
    END IF;
END $$;
COMMENT ON TABLE job_certifications IS '企业认证表（营业执照/法人/认证状态/有效期）';

-- ------------------------------------------------------------
-- 9. job_messages 沟通消息表
--    对标 BOSS直聘：在线聊天 + 卡片消息 + 系统消息
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS job_messages (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    conversation_id VARCHAR(64) NOT NULL,
    job_id BIGINT NOT NULL DEFAULT 0,
    application_id BIGINT NOT NULL DEFAULT 0,
    from_user_id BIGINT NOT NULL,
    to_user_id BIGINT NOT NULL,
    from_name VARCHAR(50) NOT NULL DEFAULT '',
    from_avatar VARCHAR(255) NOT NULL DEFAULT '',
    to_name VARCHAR(50) NOT NULL DEFAULT '',
    to_avatar VARCHAR(255) NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    message_type VARCHAR(32) NOT NULL DEFAULT 'text',
    attachments JSONB,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    read_at TIMESTAMPTZ,
    is_recruiter BOOLEAN NOT NULL DEFAULT FALSE,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    status INT NOT NULL DEFAULT 1,
    source VARCHAR(32) NOT NULL DEFAULT 'chat'
);
CREATE INDEX IF NOT EXISTS idx_job_messages_region_id ON job_messages(region_id);
CREATE INDEX IF NOT EXISTS idx_job_messages_conversation_id ON job_messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_job_messages_job_id ON job_messages(job_id);
CREATE INDEX IF NOT EXISTS idx_job_messages_application_id ON job_messages(application_id);
CREATE INDEX IF NOT EXISTS idx_job_messages_from_user ON job_messages(from_user_id, is_read);
CREATE INDEX IF NOT EXISTS idx_job_messages_to_user ON job_messages(to_user_id, is_read);
CREATE INDEX IF NOT EXISTS idx_job_messages_message_type ON job_messages(message_type);
CREATE INDEX IF NOT EXISTS idx_job_messages_is_read ON job_messages(is_read) WHERE is_read = FALSE;
CREATE INDEX IF NOT EXISTS idx_job_messages_is_system ON job_messages(is_system) WHERE is_system = TRUE;
CREATE INDEX IF NOT EXISTS idx_job_messages_status ON job_messages(status);
CREATE INDEX IF NOT EXISTS idx_job_messages_created_at ON job_messages(created_at);
CREATE INDEX IF NOT EXISTS idx_job_messages_deleted_at ON job_messages(deleted_at);
COMMENT ON TABLE job_messages IS '沟通消息表（在线聊天/卡片消息/系统消息）';

-- ------------------------------------------------------------
-- 10. job_favorites 职位收藏表
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS job_favorites (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id BIGINT NOT NULL,
    job_id BIGINT NOT NULL DEFAULT 0,
    company_id BIGINT NOT NULL DEFAULT 0,
    favorite_type VARCHAR(16) NOT NULL DEFAULT 'job',
    notify BOOLEAN NOT NULL DEFAULT TRUE,
    CONSTRAINT uniq_job_fav_user_type_target UNIQUE (user_id, favorite_type, job_id, company_id)
);
CREATE INDEX IF NOT EXISTS idx_job_favorites_user_id ON job_favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_job_favorites_job_id ON job_favorites(job_id);
CREATE INDEX IF NOT EXISTS idx_job_favorites_company_id ON job_favorites(company_id);
CREATE INDEX IF NOT EXISTS idx_job_favorites_favorite_type ON job_favorites(favorite_type);
CREATE INDEX IF NOT EXISTS idx_job_favorites_deleted_at ON job_favorites(deleted_at);
COMMENT ON TABLE job_favorites IS '职位/公司收藏表（用户/类型/目标 ID）';

-- ------------------------------------------------------------
-- 11. job_views 浏览记录表
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS job_views (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_id BIGINT NOT NULL,
    job_id BIGINT NOT NULL DEFAULT 0,
    company_id BIGINT NOT NULL DEFAULT 0,
    resume_id BIGINT NOT NULL DEFAULT 0,
    recruiter_id BIGINT NOT NULL DEFAULT 0,
    view_type VARCHAR(16) NOT NULL DEFAULT 'job',
    source VARCHAR(32) NOT NULL DEFAULT 'list',
    ip VARCHAR(64) NOT NULL DEFAULT '',
    user_agent VARCHAR(255) NOT NULL DEFAULT '',
    duration_seconds INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_job_views_region_id ON job_views(region_id);
CREATE INDEX IF NOT EXISTS idx_job_views_user_id ON job_views(user_id);
CREATE INDEX IF NOT EXISTS idx_job_views_job_id ON job_views(job_id);
CREATE INDEX IF NOT EXISTS idx_job_views_company_id ON job_views(company_id);
CREATE INDEX IF NOT EXISTS idx_job_views_resume_id ON job_views(resume_id);
CREATE INDEX IF NOT EXISTS idx_job_views_view_type ON job_views(view_type);
CREATE INDEX IF NOT EXISTS idx_job_views_source ON job_views(source);
CREATE INDEX IF NOT EXISTS idx_job_views_created_at ON job_views(created_at);
COMMENT ON TABLE job_views IS '浏览记录表（用户/职位/公司/简历/来源）';

-- ------------------------------------------------------------
-- 12. job_reports 举报工单表
--    对标 BOSS直聘：虚假招聘/诈骗/色情/侵权 + SLA
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS job_reports (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    report_no VARCHAR(64) NOT NULL,
    target_type VARCHAR(32) NOT NULL,
    target_id BIGINT NOT NULL,
    target_user_id BIGINT NOT NULL,
    reporter_id BIGINT NOT NULL,
    reporter_name VARCHAR(50) NOT NULL DEFAULT '',
    reported_user_id BIGINT NOT NULL,
    reported_user_name VARCHAR(50) NOT NULL DEFAULT '',
    report_type VARCHAR(32) NOT NULL,
    reason VARCHAR(500) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    evidence_images JSONB,
    status INT NOT NULL DEFAULT 0,
    handler_id BIGINT NOT NULL DEFAULT 0,
    handler_name VARCHAR(50) NOT NULL DEFAULT '',
    handle_result TEXT NOT NULL DEFAULT '',
    penalty_type VARCHAR(32) NOT NULL DEFAULT '',
    penalty_user_id BIGINT NOT NULL DEFAULT 0,
    sla_deadline TIMESTAMPTZ,
    handled_at TIMESTAMPTZ,
    appeal_reason TEXT NOT NULL DEFAULT '',
    appealed_at TIMESTAMPTZ,
    appeal_result TEXT NOT NULL DEFAULT '',
    appeal_handler_id BIGINT NOT NULL DEFAULT 0,
    appeal_handled_at TIMESTAMPTZ,
    CONSTRAINT uniq_job_reports_no UNIQUE (report_no)
);
CREATE INDEX IF NOT EXISTS idx_job_reports_target_type_target ON job_reports(target_type, target_id);
CREATE INDEX IF NOT EXISTS idx_job_reports_target_user_id ON job_reports(target_user_id);
CREATE INDEX IF NOT EXISTS idx_job_reports_reporter_id ON job_reports(reporter_id);
CREATE INDEX IF NOT EXISTS idx_job_reports_reported_user_id ON job_reports(reported_user_id);
CREATE INDEX IF NOT EXISTS idx_job_reports_report_type ON job_reports(report_type);
CREATE INDEX IF NOT EXISTS idx_job_reports_status ON job_reports(status);
CREATE INDEX IF NOT EXISTS idx_job_reports_handler_id ON job_reports(handler_id);
CREATE INDEX IF NOT EXISTS idx_job_reports_penalty_user_id ON job_reports(penalty_user_id);
CREATE INDEX IF NOT EXISTS idx_job_reports_sla_deadline ON job_reports(sla_deadline);
CREATE INDEX IF NOT EXISTS idx_job_reports_handled_at ON job_reports(handled_at);
CREATE INDEX IF NOT EXISTS idx_job_reports_appealed_at ON job_reports(appealed_at);
CREATE INDEX IF NOT EXISTS idx_job_reports_deleted_at ON job_reports(deleted_at);
COMMENT ON TABLE job_reports IS '举报工单表（目标/类型/原因/证据/状态/SLA/申诉）';

-- ------------------------------------------------------------
-- 13. job_reviews 公司评价表
--    对标 BOSS直聘/看准：5星+文字+图片+优缺点+追评
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS job_reviews (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    company_id BIGINT NOT NULL,
    reviewer_id BIGINT NOT NULL,
    reviewer_name VARCHAR(50) NOT NULL DEFAULT '',
    reviewer_avatar VARCHAR(255) NOT NULL DEFAULT '',
    review_type VARCHAR(32) NOT NULL DEFAULT 'employee',
    rating INT NOT NULL DEFAULT 5,
    content TEXT NOT NULL DEFAULT '',
    images JSONB,
    video_url VARCHAR(255) NOT NULL DEFAULT '',
    is_anonymous BOOLEAN NOT NULL DEFAULT FALSE,
    is_recommended BOOLEAN NOT NULL DEFAULT TRUE,
    tags JSONB,
    position VARCHAR(64) NOT NULL DEFAULT '',
    department VARCHAR(64) NOT NULL DEFAULT '',
    work_duration VARCHAR(32) NOT NULL DEFAULT '',
    salary_range VARCHAR(64) NOT NULL DEFAULT '',
    pros TEXT NOT NULL DEFAULT '',
    cons TEXT NOT NULL DEFAULT '',
    advice TEXT NOT NULL DEFAULT '',
    reply TEXT NOT NULL DEFAULT '',
    reply_at TIMESTAMPTZ,
    append_content TEXT NOT NULL DEFAULT '',
    append_images JSONB,
    append_at TIMESTAMPTZ,
    like_count INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 1,
    CONSTRAINT uniq_job_reviews_company_reviewer UNIQUE (company_id, reviewer_id)
);
CREATE INDEX IF NOT EXISTS idx_job_reviews_region_id ON job_reviews(region_id);
CREATE INDEX IF NOT EXISTS idx_job_reviews_company_id ON job_reviews(company_id);
CREATE INDEX IF NOT EXISTS idx_job_reviews_reviewer_id ON job_reviews(reviewer_id);
CREATE INDEX IF NOT EXISTS idx_job_reviews_review_type ON job_reviews(review_type);
CREATE INDEX IF NOT EXISTS idx_job_reviews_rating ON job_reviews(rating);
CREATE INDEX IF NOT EXISTS idx_job_reviews_is_recommended ON job_reviews(is_recommended) WHERE is_recommended = FALSE;
CREATE INDEX IF NOT EXISTS idx_job_reviews_status ON job_reviews(status);
CREATE INDEX IF NOT EXISTS idx_job_reviews_reply_at ON job_reviews(reply_at);
CREATE INDEX IF NOT EXISTS idx_job_reviews_append_at ON job_reviews(append_at);
CREATE INDEX IF NOT EXISTS idx_job_reviews_deleted_at ON job_reviews(deleted_at);
COMMENT ON TABLE job_reviews IS '公司评价表（5星/优缺点/追评/回复）';

-- ------------------------------------------------------------
-- 14. job_salary_ranges 薪资范围配置表
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS job_salary_ranges (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    name VARCHAR(64) NOT NULL,
    code VARCHAR(64) NOT NULL,
    min_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    max_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    currency VARCHAR(8) NOT NULL DEFAULT 'CNY',
    period VARCHAR(16) NOT NULL DEFAULT 'month',
    description VARCHAR(500) NOT NULL DEFAULT '',
    sort INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 1,
    is_hot BOOLEAN NOT NULL DEFAULT FALSE,
    use_count INT NOT NULL DEFAULT 0,
    CONSTRAINT uniq_job_salary_ranges_code UNIQUE (code)
);
CREATE INDEX IF NOT EXISTS idx_job_salary_ranges_status ON job_salary_ranges(status);
CREATE INDEX IF NOT EXISTS idx_job_salary_ranges_is_hot ON job_salary_ranges(is_hot) WHERE is_hot = TRUE;
CREATE INDEX IF NOT EXISTS idx_job_salary_ranges_sort ON job_salary_ranges(sort);
CREATE INDEX IF NOT EXISTS idx_job_salary_ranges_deleted_at ON job_salary_ranges(deleted_at);
COMMENT ON TABLE job_salary_ranges IS '薪资范围配置表（月/年/日/时）';

-- ------------------------------------------------------------
-- 15. job_benefits 福利标签表
--    对标 BOSS直聘：五险一金/年终奖/弹性工作/带薪年假
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS job_benefits (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    name VARCHAR(64) NOT NULL,
    code VARCHAR(64) NOT NULL,
    icon VARCHAR(64) NOT NULL DEFAULT '',
    color VARCHAR(16) NOT NULL DEFAULT '#67C23A',
    background VARCHAR(32) NOT NULL DEFAULT '',
    category VARCHAR(32) NOT NULL DEFAULT 'welfare',
    description VARCHAR(500) NOT NULL DEFAULT '',
    status INT NOT NULL DEFAULT 1,
    sort INT NOT NULL DEFAULT 0,
    use_count INT NOT NULL DEFAULT 0,
    is_hot BOOLEAN NOT NULL DEFAULT FALSE,
    creator_id BIGINT NOT NULL DEFAULT 0,
    CONSTRAINT uniq_job_benefits_code UNIQUE (code)
);
CREATE INDEX IF NOT EXISTS idx_job_benefits_category ON job_benefits(category);
CREATE INDEX IF NOT EXISTS idx_job_benefits_status ON job_benefits(status);
CREATE INDEX IF NOT EXISTS idx_job_benefits_is_hot ON job_benefits(is_hot) WHERE is_hot = TRUE;
CREATE INDEX IF NOT EXISTS idx_job_benefits_sort ON job_benefits(sort);
CREATE INDEX IF NOT EXISTS idx_job_benefits_creator_id ON job_benefits(creator_id);
CREATE INDEX IF NOT EXISTS idx_job_benefits_deleted_at ON job_benefits(deleted_at);
COMMENT ON TABLE job_benefits IS '福利标签表（五险一金/年终奖/弹性工作等）';

-- ------------------------------------------------------------
-- 16. job_audit_rules 审核规则表
--    对标 BOSS直聘：敏感词/薪资异常/频率限制/虚假招聘
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS job_audit_rules (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    rule_name VARCHAR(128) NOT NULL,
    rule_type VARCHAR(32) NOT NULL,
    rule_key VARCHAR(64) NOT NULL DEFAULT '',
    pattern TEXT NOT NULL DEFAULT '',
    threshold JSONB,
    action VARCHAR(32) NOT NULL DEFAULT 'reject',
    penalty_type VARCHAR(32) NOT NULL DEFAULT '',
    severity INT NOT NULL DEFAULT 1,
    status INT NOT NULL DEFAULT 1,
    description VARCHAR(500) NOT NULL DEFAULT '',
    sort INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_job_audit_rules_rule_type ON job_audit_rules(rule_type);
CREATE INDEX IF NOT EXISTS idx_job_audit_rules_rule_key ON job_audit_rules(rule_key);
CREATE INDEX IF NOT EXISTS idx_job_audit_rules_severity ON job_audit_rules(severity);
CREATE INDEX IF NOT EXISTS idx_job_audit_rules_status ON job_audit_rules(status);
CREATE INDEX IF NOT EXISTS idx_job_audit_rules_deleted_at ON job_audit_rules(deleted_at);
COMMENT ON TABLE job_audit_rules IS '审核规则表（敏感词/薪资异常/频率限制/虚假招聘）';

-- ------------------------------------------------------------
-- 17. job_statistics 数据统计表
--    对标 BOSS直聘：曝光/点击/投递/面试/Offer/入职转化
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS job_statistics (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    stat_date DATE NOT NULL,
    stat_type VARCHAR(32) NOT NULL,
    target_id BIGINT NOT NULL DEFAULT 0,
    target_name VARCHAR(128) NOT NULL DEFAULT '',
    impression_count INT NOT NULL DEFAULT 0,
    click_count INT NOT NULL DEFAULT 0,
    fav_count INT NOT NULL DEFAULT 0,
    deliver_count INT NOT NULL DEFAULT 0,
    interview_count INT NOT NULL DEFAULT 0,
    offer_count INT NOT NULL DEFAULT 0,
    onboarding_count INT NOT NULL DEFAULT 0,
    conversion_rate DECIMAL(5,2) NOT NULL DEFAULT 0,
    avg_salary DECIMAL(12,2) NOT NULL DEFAULT 0,
    median_salary DECIMAL(12,2) NOT NULL DEFAULT 0,
    retention_30d INT NOT NULL DEFAULT 0,
    retention_90d INT NOT NULL DEFAULT 0,
    CONSTRAINT uniq_job_statistics_date_type_target UNIQUE (stat_date, stat_type, target_id)
);
CREATE INDEX IF NOT EXISTS idx_job_statistics_region_id ON job_statistics(region_id);
CREATE INDEX IF NOT EXISTS idx_job_statistics_stat_date ON job_statistics(stat_date);
CREATE INDEX IF NOT EXISTS idx_job_statistics_stat_type ON job_statistics(stat_type);
CREATE INDEX IF NOT EXISTS idx_job_statistics_target_id ON job_statistics(target_id);
CREATE INDEX IF NOT EXISTS idx_job_statistics_deleted_at ON job_statistics(deleted_at);
COMMENT ON TABLE job_statistics IS '数据统计表（曝光/点击/投递/面试/Offer/入职/转化率）';

-- ------------------------------------------------------------
-- 18. job_escrows 担保交易表
--    对标 BOSS直聘：招聘保证金/中介费托管
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS job_escrows (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    escrow_no VARCHAR(64) NOT NULL,
    escrow_type VARCHAR(32) NOT NULL DEFAULT 'recruitment_deposit',
    job_id BIGINT NOT NULL DEFAULT 0,
    application_id BIGINT NOT NULL DEFAULT 0,
    company_id BIGINT NOT NULL DEFAULT 0,
    payer_id BIGINT NOT NULL,
    payee_id BIGINT NOT NULL,
    amount DECIMAL(12,2) NOT NULL,
    platform_fee DECIMAL(12,2) NOT NULL DEFAULT 0,
    payee_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 1,
    pay_method VARCHAR(32) NOT NULL DEFAULT 'wechat',
    pay_trade_no VARCHAR(128) NOT NULL DEFAULT '',
    paid_at TIMESTAMPTZ,
    frozen_at TIMESTAMPTZ,
    release_at TIMESTAMPTZ,
    refunded_at TIMESTAMPTZ,
    auto_release_at TIMESTAMPTZ,
    dispute_reason VARCHAR(500) NOT NULL DEFAULT '',
    arbitration_result TEXT NOT NULL DEFAULT '',
    completed_at TIMESTAMPTZ,
    CONSTRAINT uniq_job_escrows_no UNIQUE (escrow_no)
);
CREATE INDEX IF NOT EXISTS idx_job_escrows_region_id ON job_escrows(region_id);
CREATE INDEX IF NOT EXISTS idx_job_escrows_escrow_type ON job_escrows(escrow_type);
CREATE INDEX IF NOT EXISTS idx_job_escrows_job_id ON job_escrows(job_id);
CREATE INDEX IF NOT EXISTS idx_job_escrows_application_id ON job_escrows(application_id);
CREATE INDEX IF NOT EXISTS idx_job_escrows_company_id ON job_escrows(company_id);
CREATE INDEX IF NOT EXISTS idx_job_escrows_payer_id ON job_escrows(payer_id);
CREATE INDEX IF NOT EXISTS idx_job_escrows_payee_id ON job_escrows(payee_id);
CREATE INDEX IF NOT EXISTS idx_job_escrows_status ON job_escrows(status);
CREATE INDEX IF NOT EXISTS idx_job_escrows_pay_trade_no ON job_escrows(pay_trade_no);
CREATE INDEX IF NOT EXISTS idx_job_escrows_paid_at ON job_escrows(paid_at);
CREATE INDEX IF NOT EXISTS idx_job_escrows_frozen_at ON job_escrows(frozen_at);
CREATE INDEX IF NOT EXISTS idx_job_escrows_release_at ON job_escrows(release_at);
CREATE INDEX IF NOT EXISTS idx_job_escrows_refunded_at ON job_escrows(refunded_at);
CREATE INDEX IF NOT EXISTS idx_job_escrows_auto_release_at ON job_escrows(auto_release_at);
CREATE INDEX IF NOT EXISTS idx_job_escrows_deleted_at ON job_escrows(deleted_at);
COMMENT ON TABLE job_escrows IS '担保交易表（招聘保证金/中介费托管/资金冻结/解冻/放款）';

-- ------------------------------------------------------------
-- 19. job_recommendations 推荐记录表
--    对标 BOSS直聘：AI 智能推荐（人岗匹配）
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS job_recommendations (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id BIGINT NOT NULL,
    job_id BIGINT NOT NULL,
    rec_type VARCHAR(32) NOT NULL DEFAULT 'job_to_user',
    source VARCHAR(32) NOT NULL DEFAULT 'ai',
    score DECIMAL(5,2) NOT NULL DEFAULT 0,
    reason VARCHAR(500) NOT NULL DEFAULT '',
    position_match DECIMAL(5,2) NOT NULL DEFAULT 0,
    salary_match DECIMAL(5,2) NOT NULL DEFAULT 0,
    location_match DECIMAL(5,2) NOT NULL DEFAULT 0,
    skill_match DECIMAL(5,2) NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 0,
    clicked_at TIMESTAMPTZ,
    applied_at TIMESTAMPTZ,
    dismissed_at TIMESTAMPTZ,
    expired_at TIMESTAMPTZ,
    CONSTRAINT uniq_job_recs_user_job_type UNIQUE (user_id, job_id, rec_type)
);
CREATE INDEX IF NOT EXISTS idx_job_recommendations_region_id ON job_recommendations(region_id);
CREATE INDEX IF NOT EXISTS idx_job_recommendations_user_id ON job_recommendations(user_id);
CREATE INDEX IF NOT EXISTS idx_job_recommendations_job_id ON job_recommendations(job_id);
CREATE INDEX IF NOT EXISTS idx_job_recommendations_rec_type ON job_recommendations(rec_type);
CREATE INDEX IF NOT EXISTS idx_job_recommendations_source ON job_recommendations(source);
CREATE INDEX IF NOT EXISTS idx_job_recommendations_score ON job_recommendations(score);
CREATE INDEX IF NOT EXISTS idx_job_recommendations_status ON job_recommendations(status);
CREATE INDEX IF NOT EXISTS idx_job_recommendations_clicked_at ON job_recommendations(clicked_at);
CREATE INDEX IF NOT EXISTS idx_job_recommendations_applied_at ON job_recommendations(applied_at);
CREATE INDEX IF NOT EXISTS idx_job_recommendations_dismissed_at ON job_recommendations(dismissed_at);
CREATE INDEX IF NOT EXISTS idx_job_recommendations_expired_at ON job_recommendations(expired_at);
CREATE INDEX IF NOT EXISTS idx_job_recommendations_deleted_at ON job_recommendations(deleted_at);
COMMENT ON TABLE job_recommendations IS '推荐记录表（AI 智能推荐/人岗匹配/多维评分）';

-- ============================================================
-- 第三部分：为 19 张表挂载 updated_at 触发器
--   参考 001_p0_baseline.sql 中的 update_updated_at_column 函数
--   幂等：先 DROP IF EXISTS 再 CREATE
-- ============================================================
DO $$
DECLARE t TEXT;
BEGIN
    FOR t IN SELECT tablename FROM pg_tables WHERE schemaname='public' AND tablename IN (
        'job_companies','job_resumes','job_applications','job_interviews','job_positions',
        'job_categories','job_skills','job_certifications','job_messages','job_favorites',
        'job_views','job_reports','job_reviews','job_salary_ranges','job_benefits',
        'job_audit_rules','job_statistics','job_escrows','job_recommendations'
    )
    LOOP
        EXECUTE format('DROP TRIGGER IF EXISTS trg_%s_updated ON %s', t, t);
        EXECUTE format('CREATE TRIGGER trg_%s_updated BEFORE UPDATE ON %s FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()', t, t);
    END LOOP;
END $$;
