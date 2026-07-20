-- ============================================================
-- linggong 零工兼职模块完整功能迁移脚本（v3.2.1）
-- 对标：斗米 / 青团兼职 / 兼职猫 / 猪八戒
--
-- 内容：
--   1. ALTER TABLE linggongs 主表新增字段（类型/计费/结算/任务/统计/运营等）
--   2. CREATE 17 张子表（linggong_ 前缀，依据数据库分表前缀规范 v1.0.0）
--   3. 索引、外键、触发器、注释
--   4. 全幂等：CREATE TABLE IF NOT EXISTS / ALTER TABLE ADD COLUMN IF NOT EXISTS
-- 依赖：001_p0_baseline.sql（update_updated_at_column 函数已创建）
-- ============================================================

-- ============================================================
-- 第一部分：扩展 linggongs 主表
-- 注意：linggongs 表由 GORM AutoMigrate 在应用启动时创建
--      本迁移对 linggongs 的所有 ALTER 操作包装在 DO 块中，
--      若表不存在则跳过（待应用启动后再执行一次本迁移即可补齐字段）
-- ============================================================
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'linggongs') THEN
        -- === 基础信息 ===
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS cover_image VARCHAR(255);
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS user_phone VARCHAR(20);
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS user_avatar VARCHAR(255);
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS audit_reason VARCHAR(500);
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS published_at TIMESTAMPTZ;
        CREATE INDEX IF NOT EXISTS idx_linggongs_published_at ON linggongs(published_at);

        -- === 岗位类型/发布者类型 ===
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS linggong_type VARCHAR(32) NOT NULL DEFAULT 'short_term';
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS publisher_type VARCHAR(32) NOT NULL DEFAULT 'personal';
        CREATE INDEX IF NOT EXISTS idx_linggongs_linggong_type ON linggongs(linggong_type);
        CREATE INDEX IF NOT EXISTS idx_linggongs_publisher_type ON linggongs(publisher_type);

        -- === 雇主关联 ===
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS employer_id BIGINT NOT NULL DEFAULT 0;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS company_name VARCHAR(128) NOT NULL DEFAULT '';
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS contact_name VARCHAR(50) NOT NULL DEFAULT '';
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS contact_phone VARCHAR(20) NOT NULL DEFAULT '';
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS contact_wechat VARCHAR(64) NOT NULL DEFAULT '';
        CREATE INDEX IF NOT EXISTS idx_linggongs_employer_id ON linggongs(employer_id);
        CREATE INDEX IF NOT EXISTS idx_linggongs_company_name ON linggongs(company_name);

        -- === 计费方式/薪资 ===
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS billing_type VARCHAR(32) NOT NULL DEFAULT 'by_day';
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS salary_min DECIMAL(12,2) NOT NULL DEFAULT 0;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS salary_max DECIMAL(12,2) NOT NULL DEFAULT 0;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS salary_unit VARCHAR(32) NOT NULL DEFAULT '';
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS salary_negotiable BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS settlement VARCHAR(16) NOT NULL DEFAULT 'T+1';
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS currency VARCHAR(16) NOT NULL DEFAULT 'CNY';
        CREATE INDEX IF NOT EXISTS idx_linggongs_billing_type ON linggongs(billing_type);
        CREATE INDEX IF NOT EXISTS idx_linggongs_salary_min ON linggongs(salary_min);
        CREATE INDEX IF NOT EXISTS idx_linggongs_salary_max ON linggongs(salary_max);
        CREATE INDEX IF NOT EXISTS idx_linggongs_settlement ON linggongs(settlement);

        -- === 工作时间 ===
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS work_start_date DATE;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS work_end_date DATE;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS work_days INT NOT NULL DEFAULT 0;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS work_hours INT NOT NULL DEFAULT 0;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS work_time_start VARCHAR(16) NOT NULL DEFAULT '';
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS work_time_end VARCHAR(16) NOT NULL DEFAULT '';
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS work_weekdays VARCHAR(32) NOT NULL DEFAULT '1,2,3,4,5';
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS work_intensity VARCHAR(16) NOT NULL DEFAULT 'medium';
        CREATE INDEX IF NOT EXISTS idx_linggongs_work_start_date ON linggongs(work_start_date);
        CREATE INDEX IF NOT EXISTS idx_linggongs_work_end_date ON linggongs(work_end_date);

        -- === 招募要求 ===
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS recruit_count INT NOT NULL DEFAULT 1;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS applied_count INT NOT NULL DEFAULT 0;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS confirmed_count INT NOT NULL DEFAULT 0;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS need_gender VARCHAR(16) NOT NULL DEFAULT 'any';
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS min_age INT NOT NULL DEFAULT 0;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS max_age INT NOT NULL DEFAULT 0;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS education VARCHAR(32) NOT NULL DEFAULT '';
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS experience VARCHAR(64) NOT NULL DEFAULT '';
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS need_health_cert BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS need_id_card BOOLEAN NOT NULL DEFAULT TRUE;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS min_credit_score INT NOT NULL DEFAULT 0;

        -- === 工作地点 ===
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS province VARCHAR(64) NOT NULL DEFAULT '';
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS city VARCHAR(64) NOT NULL DEFAULT '';
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS district VARCHAR(64) NOT NULL DEFAULT '';
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS business_district VARCHAR(128) NOT NULL DEFAULT '';
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS address VARCHAR(500) NOT NULL DEFAULT '';
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS latitude DECIMAL(10,7) NOT NULL DEFAULT 0;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS longitude DECIMAL(10,7) NOT NULL DEFAULT 0;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS work_location_type VARCHAR(16) NOT NULL DEFAULT 'onsite';
        CREATE INDEX IF NOT EXISTS idx_linggongs_province ON linggongs(province);
        CREATE INDEX IF NOT EXISTS idx_linggongs_city ON linggongs(city);
        CREATE INDEX IF NOT EXISTS idx_linggongs_district ON linggongs(district);

        -- === 任务制相关 ===
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS task_id BIGINT NOT NULL DEFAULT 0;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS total_task_count INT NOT NULL DEFAULT 0;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS claimed_count INT NOT NULL DEFAULT 0;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS completed_task_count INT NOT NULL DEFAULT 0;
        CREATE INDEX IF NOT EXISTS idx_linggongs_task_id ON linggongs(task_id);

        -- === 互动统计 ===
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS view_count INT NOT NULL DEFAULT 0;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS fav_count INT NOT NULL DEFAULT 0;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS contact_count INT NOT NULL DEFAULT 0;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS share_count INT NOT NULL DEFAULT 0;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS application_count INT NOT NULL DEFAULT 0;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS last_applied_at TIMESTAMPTZ;
        CREATE INDEX IF NOT EXISTS idx_linggongs_last_applied_at ON linggongs(last_applied_at);

        -- === 风控 ===
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS content_hash VARCHAR(64);
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS risk_score INT NOT NULL DEFAULT 0;
        CREATE INDEX IF NOT EXISTS idx_linggongs_content_hash ON linggongs(content_hash);
        CREATE INDEX IF NOT EXISTS idx_linggongs_risk_score ON linggongs(risk_score);

        -- === 视频/图片 ===
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS video_url VARCHAR(255);
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS video_cover VARCHAR(255);

        -- === 配置/特征（JSONB） ===
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS features JSONB;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS tags JSONB;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS skill_tags JSONB;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS welfare_tags JSONB;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS images JSONB;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS requirements JSONB;
        CREATE INDEX IF NOT EXISTS idx_linggongs_features ON linggongs USING GIN(features);
        CREATE INDEX IF NOT EXISTS idx_linggongs_tags ON linggongs USING GIN(tags);
        CREATE INDEX IF NOT EXISTS idx_linggongs_skill_tags ON linggongs USING GIN(skill_tags);
        CREATE INDEX IF NOT EXISTS idx_linggongs_welfare_tags ON linggongs USING GIN(welfare_tags);

        -- === 运营字段 ===
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS featured BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS picked BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS verified BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS promotion_level INT NOT NULL DEFAULT 0;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS traffic_weight DECIMAL(3,2) NOT NULL DEFAULT 1.00;
        CREATE INDEX IF NOT EXISTS idx_linggongs_featured ON linggongs(featured) WHERE featured = TRUE;
        CREATE INDEX IF NOT EXISTS idx_linggongs_picked ON linggongs(picked) WHERE picked = TRUE;
        CREATE INDEX IF NOT EXISTS idx_linggongs_verified ON linggongs(verified) WHERE verified = TRUE;

        -- === 雇主认证状态 ===
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS employer_verified BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE linggongs ADD COLUMN IF NOT EXISTS employer_verified_at TIMESTAMPTZ;
        CREATE INDEX IF NOT EXISTS idx_linggongs_employer_verified ON linggongs(employer_verified) WHERE employer_verified = TRUE;

        -- 字段注释
        COMMENT ON COLUMN linggongs.linggong_type IS '岗位类型：short_term短期/long_term长期/task任务制/hourly小时工/daily日结工/temp临时';
        COMMENT ON COLUMN linggongs.publisher_type IS '发布者：personal个人/company企业/agent中介/headhunter猎头';
        COMMENT ON COLUMN linggongs.billing_type IS '计费：by_piece按件/by_hour按时/by_day按日/by_week按周/by_month按月/fixed固定/negotiable面议';
        COMMENT ON COLUMN linggongs.settlement IS '结算周期：T+0当日结/T+1次日结/T+3三日结/T+7周结/M+1月结/project项目结';
        COMMENT ON COLUMN linggongs.work_location_type IS '工作地点类型：onsite现场/remote远程/hybrid混合';
        COMMENT ON COLUMN linggongs.salary_min IS '薪资范围下限（元）';
        COMMENT ON COLUMN linggongs.salary_max IS '薪资范围上限（元）';
        COMMENT ON COLUMN linggongs.features IS '岗位特征 JSON：包含工作内容/福利/要求/设备';
        COMMENT ON COLUMN linggongs.skill_tags IS '技能标签 ID 数组';
        COMMENT ON COLUMN linggongs.welfare_tags IS '福利标签：包吃/包住/五险一金/年终奖';
    END IF;
END $$;

-- ============================================================
-- 第二部分：17 张子表 CREATE TABLE IF NOT EXISTS
-- 表前缀 linggong_ 依据 docs/架构设计/数据库分表前缀规范.md
-- ============================================================

-- ------------------------------------------------------------
-- 1. linggong_tasks 任务包表（对标斗米任务制 + 猪八戒威客）
--    长短期任务分类 + 按件/按时/按日计费 + 任务领取/交付
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS linggong_tasks (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    task_no VARCHAR(64) NOT NULL,
    linggong_id BIGINT NOT NULL,
    employer_id BIGINT NOT NULL,
    employer_name VARCHAR(128) NOT NULL DEFAULT '',
    title VARCHAR(200) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    task_type VARCHAR(32) NOT NULL DEFAULT 'single',
    difficulty VARCHAR(16) NOT NULL DEFAULT 'easy',
    delivery_method VARCHAR(16) NOT NULL DEFAULT 'online',
    billing_type VARCHAR(32) NOT NULL DEFAULT 'by_piece',
    unit_price DECIMAL(12,2) NOT NULL DEFAULT 0,
    total_count INT NOT NULL DEFAULT 1,
    claimed_count INT NOT NULL DEFAULT 0,
    completed_count INT NOT NULL DEFAULT 0,
    verified_count INT NOT NULL DEFAULT 0,
    max_claim_per_user INT NOT NULL DEFAULT 1,
    start_time TIMESTAMPTZ,
    end_time TIMESTAMPTZ,
    claim_deadline TIMESTAMPTZ,
    submit_deadline TIMESTAMPTZ,
    total_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    paid_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 0,
    attachment_url VARCHAR(255) NOT NULL DEFAULT '',
    tags JSONB,
    requirements JSONB,
    published_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    CONSTRAINT uniq_linggong_tasks_no UNIQUE (task_no)
);
CREATE INDEX IF NOT EXISTS idx_linggong_tasks_region_id ON linggong_tasks(region_id);
CREATE INDEX IF NOT EXISTS idx_linggong_tasks_linggong_id ON linggong_tasks(linggong_id);
CREATE INDEX IF NOT EXISTS idx_linggong_tasks_employer_id ON linggong_tasks(employer_id);
CREATE INDEX IF NOT EXISTS idx_linggong_tasks_task_type ON linggong_tasks(task_type);
CREATE INDEX IF NOT EXISTS idx_linggong_tasks_difficulty ON linggong_tasks(difficulty);
CREATE INDEX IF NOT EXISTS idx_linggong_tasks_billing_type ON linggong_tasks(billing_type);
CREATE INDEX IF NOT EXISTS idx_linggong_tasks_unit_price ON linggong_tasks(unit_price);
CREATE INDEX IF NOT EXISTS idx_linggong_tasks_status ON linggong_tasks(status);
CREATE INDEX IF NOT EXISTS idx_linggong_tasks_published_at ON linggong_tasks(published_at);
CREATE INDEX IF NOT EXISTS idx_linggong_tasks_end_time ON linggong_tasks(end_time);
CREATE INDEX IF NOT EXISTS idx_linggong_tasks_deleted_at ON linggong_tasks(deleted_at);
COMMENT ON TABLE linggong_tasks IS '任务包表（长短期任务 + 按件/按时/按日计费 + 任务领取/交付）';

-- ------------------------------------------------------------
-- 2. linggong_applications 报名记录表（对标斗米/兼职猫）
--    报名状态机 + 报名审核 + 报名取消/拒绝
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS linggong_applications (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    application_no VARCHAR(64) NOT NULL,
    linggong_id BIGINT NOT NULL,
    task_id BIGINT NOT NULL DEFAULT 0,
    employer_id BIGINT NOT NULL,
    employer_name VARCHAR(128) NOT NULL DEFAULT '',
    worker_id BIGINT NOT NULL,
    worker_name VARCHAR(50) NOT NULL DEFAULT '',
    worker_avatar VARCHAR(255) NOT NULL DEFAULT '',
    worker_phone VARCHAR(20) NOT NULL DEFAULT '',
    worker_age INT NOT NULL DEFAULT 0,
    worker_gender VARCHAR(16) NOT NULL DEFAULT 'unknown',
    worker_city VARCHAR(64) NOT NULL DEFAULT '',
    worker_credit_score INT NOT NULL DEFAULT 0,
    worker_profile_id BIGINT NOT NULL DEFAULT 0,
    source VARCHAR(32) NOT NULL DEFAULT 'direct',
    method VARCHAR(16) NOT NULL DEFAULT 'online',
    status INT NOT NULL DEFAULT 0,
    applied_count INT NOT NULL DEFAULT 1,
    cover_letter TEXT NOT NULL DEFAULT '',
    employer_remark TEXT NOT NULL DEFAULT '',
    reject_reason VARCHAR(500) NOT NULL DEFAULT '',
    cancel_reason VARCHAR(500) NOT NULL DEFAULT '',
    reviewed_at TIMESTAMPTZ,
    confirmed_at TIMESTAMPTZ,
    onboarded_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    canceled_at TIMESTAMPTZ,
    evaluated BOOLEAN NOT NULL DEFAULT FALSE,
    attachment_url VARCHAR(255) NOT NULL DEFAULT '',
    CONSTRAINT uniq_linggong_applications_no UNIQUE (application_no)
);
CREATE INDEX IF NOT EXISTS idx_linggong_applications_region_id ON linggong_applications(region_id);
CREATE INDEX IF NOT EXISTS idx_linggong_applications_linggong_id ON linggong_applications(linggong_id);
CREATE INDEX IF NOT EXISTS idx_linggong_applications_task_id ON linggong_applications(task_id);
CREATE INDEX IF NOT EXISTS idx_linggong_applications_employer_id ON linggong_applications(employer_id);
CREATE INDEX IF NOT EXISTS idx_linggong_applications_worker_id ON linggong_applications(worker_id);
CREATE INDEX IF NOT EXISTS idx_linggong_applications_source ON linggong_applications(source);
CREATE INDEX IF NOT EXISTS idx_linggong_applications_status ON linggong_applications(status);
CREATE INDEX IF NOT EXISTS idx_linggong_applications_evaluated ON linggong_applications(evaluated) WHERE evaluated = FALSE;
CREATE INDEX IF NOT EXISTS idx_linggong_applications_reviewed_at ON linggong_applications(reviewed_at);
CREATE INDEX IF NOT EXISTS idx_linggong_applications_onboarded_at ON linggong_applications(onboarded_at);
CREATE INDEX IF NOT EXISTS idx_linggong_applications_completed_at ON linggong_applications(completed_at);
CREATE INDEX IF NOT EXISTS idx_linggong_applications_deleted_at ON linggong_applications(deleted_at);
COMMENT ON TABLE linggong_applications IS '报名记录表（状态机 + 审核 + 取消/拒绝 + 评价关联）';

-- ------------------------------------------------------------
-- 3. linggong_employers 雇主认证表（对标斗米企业认证 + 猪八戒雇主）
--    企业/个人雇主 + 营业执照 + 法人认证 + 信用等级
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS linggong_employers (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id BIGINT NOT NULL,
    employer_type VARCHAR(32) NOT NULL DEFAULT 'personal',
    company_name VARCHAR(128) NOT NULL DEFAULT '',
    company_short_name VARCHAR(64) NOT NULL DEFAULT '',
    contact_name VARCHAR(50) NOT NULL DEFAULT '',
    contact_phone VARCHAR(20) NOT NULL DEFAULT '',
    contact_email VARCHAR(128) NOT NULL DEFAULT '',
    contact_wechat VARCHAR(64) NOT NULL DEFAULT '',
    license_no VARCHAR(64) NOT NULL DEFAULT '',
    license_url VARCHAR(255) NOT NULL DEFAULT '',
    legal_person VARCHAR(50) NOT NULL DEFAULT '',
    legal_person_id_card VARCHAR(32) NOT NULL DEFAULT '',
    legal_person_id_card_url VARCHAR(255) NOT NULL DEFAULT '',
    bank_account VARCHAR(64) NOT NULL DEFAULT '',
    bank_name VARCHAR(64) NOT NULL DEFAULT '',
    brand_auth_url VARCHAR(255) NOT NULL DEFAULT '',
    company_address VARCHAR(500) NOT NULL DEFAULT '',
    company_latitude DECIMAL(10,7) NOT NULL DEFAULT 0,
    company_longitude DECIMAL(10,7) NOT NULL DEFAULT 0,
    company_description TEXT NOT NULL DEFAULT '',
    company_logo VARCHAR(255) NOT NULL DEFAULT '',
    company_cover VARCHAR(255) NOT NULL DEFAULT '',
    industry VARCHAR(64) NOT NULL DEFAULT '',
    company_size VARCHAR(32) NOT NULL DEFAULT '',
    level INT NOT NULL DEFAULT 1,
    credit_score INT NOT NULL DEFAULT 100,
    status INT NOT NULL DEFAULT 0,
    reject_reason VARCHAR(500) NOT NULL DEFAULT '',
    verified_at TIMESTAMPTZ,
    verified_by BIGINT NOT NULL DEFAULT 0,
    verified_by_name VARCHAR(50) NOT NULL DEFAULT '',
    published_count INT NOT NULL DEFAULT 0,
    ongoing_count INT NOT NULL DEFAULT 0,
    completed_count INT NOT NULL DEFAULT 0,
    total_workers INT NOT NULL DEFAULT 0,
    total_paid DECIMAL(12,2) NOT NULL DEFAULT 0,
    avg_rating DECIMAL(3,2) NOT NULL DEFAULT 0,
    rating_count INT NOT NULL DEFAULT 0,
    CONSTRAINT uniq_linggong_employers_user UNIQUE (user_id)
);
CREATE INDEX IF NOT EXISTS idx_linggong_employers_region_id ON linggong_employers(region_id);
CREATE INDEX IF NOT EXISTS idx_linggong_employers_employer_type ON linggong_employers(employer_type);
CREATE INDEX IF NOT EXISTS idx_linggong_employers_company_name ON linggong_employers(company_name);
CREATE INDEX IF NOT EXISTS idx_linggong_employers_contact_phone ON linggong_employers(contact_phone);
CREATE INDEX IF NOT EXISTS idx_linggong_employers_level ON linggong_employers(level);
CREATE INDEX IF NOT EXISTS idx_linggong_employers_credit_score ON linggong_employers(credit_score);
CREATE INDEX IF NOT EXISTS idx_linggong_employers_status ON linggong_employers(status);
CREATE INDEX IF NOT EXISTS idx_linggong_employers_verified_at ON linggong_employers(verified_at);
CREATE INDEX IF NOT EXISTS idx_linggong_employers_deleted_at ON linggong_employers(deleted_at);
COMMENT ON TABLE linggong_employers IS '雇主认证表（企业/个人 + 营业执照 + 法人 + 信用等级）';

-- ------------------------------------------------------------
-- 4. linggong_workers 求职者档案表（对标斗米/兼职猫）
--    求职意向 + 工作经历 + 教育背景 + 技能认证
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS linggong_workers (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id BIGINT NOT NULL,
    real_name VARCHAR(50) NOT NULL DEFAULT '',
    real_name_verified BOOLEAN NOT NULL DEFAULT FALSE,
    gender VARCHAR(16) NOT NULL DEFAULT 'unknown',
    birthday DATE,
    age INT NOT NULL DEFAULT 0,
    id_card VARCHAR(32) NOT NULL DEFAULT '',
    id_card_verified BOOLEAN NOT NULL DEFAULT FALSE,
    id_card_front_url VARCHAR(255) NOT NULL DEFAULT '',
    id_card_back_url VARCHAR(255) NOT NULL DEFAULT '',
    id_card_hand_url VARCHAR(255) NOT NULL DEFAULT '',
    nickname VARCHAR(50) NOT NULL DEFAULT '',
    avatar VARCHAR(255) NOT NULL DEFAULT '',
    phone VARCHAR(20) NOT NULL DEFAULT '',
    email VARCHAR(128) NOT NULL DEFAULT '',
    wechat VARCHAR(64) NOT NULL DEFAULT '',
    province VARCHAR(64) NOT NULL DEFAULT '',
    city VARCHAR(64) NOT NULL DEFAULT '',
    district VARCHAR(64) NOT NULL DEFAULT '',
    address VARCHAR(500) NOT NULL DEFAULT '',
    latitude DECIMAL(10,7) NOT NULL DEFAULT 0,
    longitude DECIMAL(10,7) NOT NULL DEFAULT 0,
    education VARCHAR(32) NOT NULL DEFAULT '',
    school VARCHAR(128) NOT NULL DEFAULT '',
    major VARCHAR(128) NOT NULL DEFAULT '',
    graduation_year INT NOT NULL DEFAULT 0,
    job_intention VARCHAR(32) NOT NULL DEFAULT 'any',
    expected_salary DECIMAL(12,2) NOT NULL DEFAULT 0,
    available_time VARCHAR(64) NOT NULL DEFAULT '',
    available_now BOOLEAN NOT NULL DEFAULT FALSE,
    health_cert_url VARCHAR(255) NOT NULL DEFAULT '',
    health_cert_valid_until DATE,
    has_criminal_record BOOLEAN NOT NULL DEFAULT FALSE,
    criminal_record_url VARCHAR(255) NOT NULL DEFAULT '',
    bank_account VARCHAR(64) NOT NULL DEFAULT '',
    bank_name VARCHAR(64) NOT NULL DEFAULT '',
    alipay_account VARCHAR(128) NOT NULL DEFAULT '',
    wechat_pay_account VARCHAR(128) NOT NULL DEFAULT '',
    bio TEXT NOT NULL DEFAULT '',
    skill_tags JSONB,
    category_tags JSONB,
    work_experience JSONB,
    education_history JSONB,
    portfolio JSONB,
    credit_score INT NOT NULL DEFAULT 100,
    level INT NOT NULL DEFAULT 1,
    status INT NOT NULL DEFAULT 1,
    applied_count INT NOT NULL DEFAULT 0,
    completed_count INT NOT NULL DEFAULT 0,
    total_work_hours INT NOT NULL DEFAULT 0,
    total_earnings DECIMAL(12,2) NOT NULL DEFAULT 0,
    avg_rating DECIMAL(3,2) NOT NULL DEFAULT 0,
    rating_count INT NOT NULL DEFAULT 0,
    punctuality_rate DECIMAL(3,2) NOT NULL DEFAULT 0,
    completion_rate DECIMAL(3,2) NOT NULL DEFAULT 0,
    CONSTRAINT uniq_linggong_workers_user UNIQUE (user_id)
);
CREATE INDEX IF NOT EXISTS idx_linggong_workers_region_id ON linggong_workers(region_id);
CREATE INDEX IF NOT EXISTS idx_linggong_workers_real_name_verified ON linggong_workers(real_name_verified) WHERE real_name_verified = TRUE;
CREATE INDEX IF NOT EXISTS idx_linggong_workers_id_card_verified ON linggong_workers(id_card_verified) WHERE id_card_verified = TRUE;
CREATE INDEX IF NOT EXISTS idx_linggong_workers_phone ON linggong_workers(phone);
CREATE INDEX IF NOT EXISTS idx_linggong_workers_province ON linggong_workers(province);
CREATE INDEX IF NOT EXISTS idx_linggong_workers_city ON linggong_workers(city);
CREATE INDEX IF NOT EXISTS idx_linggong_workers_job_intention ON linggong_workers(job_intention);
CREATE INDEX IF NOT EXISTS idx_linggong_workers_available_now ON linggong_workers(available_now) WHERE available_now = TRUE;
CREATE INDEX IF NOT EXISTS idx_linggong_workers_credit_score ON linggong_workers(credit_score);
CREATE INDEX IF NOT EXISTS idx_linggong_workers_level ON linggong_workers(level);
CREATE INDEX IF NOT EXISTS idx_linggong_workers_status ON linggong_workers(status);
CREATE INDEX IF NOT EXISTS idx_linggong_workers_skill_tags ON linggong_workers USING GIN(skill_tags);
CREATE INDEX IF NOT EXISTS idx_linggong_workers_deleted_at ON linggong_workers(deleted_at);
COMMENT ON TABLE linggong_workers IS '求职者档案表（实名认证 + 求职意向 + 技能标签 + 信用）';

-- ------------------------------------------------------------
-- 5. linggong_contracts 电子合同表（对标法大大/e签宝）
--    兼职合同电子化 + 在线签署 + 多种签署方式
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS linggong_contracts (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    contract_no VARCHAR(64) NOT NULL,
    contract_type VARCHAR(32) NOT NULL DEFAULT 'part_time',
    linggong_id BIGINT NOT NULL,
    task_id BIGINT NOT NULL DEFAULT 0,
    application_id BIGINT NOT NULL DEFAULT 0,
    employer_id BIGINT NOT NULL,
    employer_name VARCHAR(128) NOT NULL DEFAULT '',
    employer_phone VARCHAR(20) NOT NULL DEFAULT '',
    employer_id_card VARCHAR(32) NOT NULL DEFAULT '',
    employer_sign_url VARCHAR(255) NOT NULL DEFAULT '',
    worker_id BIGINT NOT NULL,
    worker_name VARCHAR(50) NOT NULL DEFAULT '',
    worker_phone VARCHAR(20) NOT NULL DEFAULT '',
    worker_id_card VARCHAR(32) NOT NULL DEFAULT '',
    worker_sign_url VARCHAR(255) NOT NULL DEFAULT '',
    work_start_date DATE,
    work_end_date DATE,
    work_content TEXT NOT NULL DEFAULT '',
    work_place VARCHAR(500) NOT NULL DEFAULT '',
    billing_type VARCHAR(32) NOT NULL DEFAULT 'by_day',
    salary_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    salary_unit VARCHAR(32) NOT NULL DEFAULT '',
    settlement VARCHAR(16) NOT NULL DEFAULT 'T+1',
    total_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    paid_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    deposit DECIMAL(12,2) NOT NULL DEFAULT 0,
    penalty_breach DECIMAL(12,2) NOT NULL DEFAULT 0,
    confidential BOOLEAN NOT NULL DEFAULT FALSE,
    non_compete BOOLEAN NOT NULL DEFAULT FALSE,
    sign_method VARCHAR(32) NOT NULL DEFAULT 'handwritten',
    contract_url VARCHAR(255) NOT NULL DEFAULT '',
    attachments JSONB,
    employer_signed_at TIMESTAMPTZ,
    worker_signed_at TIMESTAMPTZ,
    signed_at TIMESTAMPTZ,
    effective_at TIMESTAMPTZ,
    expired_at TIMESTAMPTZ,
    terminated_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    status INT NOT NULL DEFAULT 0,
    template_id BIGINT NOT NULL DEFAULT 0,
    remark TEXT NOT NULL DEFAULT '',
    CONSTRAINT uniq_linggong_contracts_no UNIQUE (contract_no)
);
CREATE INDEX IF NOT EXISTS idx_linggong_contracts_region_id ON linggong_contracts(region_id);
CREATE INDEX IF NOT EXISTS idx_linggong_contracts_contract_type ON linggong_contracts(contract_type);
CREATE INDEX IF NOT EXISTS idx_linggong_contracts_linggong_id ON linggong_contracts(linggong_id);
CREATE INDEX IF NOT EXISTS idx_linggong_contracts_task_id ON linggong_contracts(task_id);
CREATE INDEX IF NOT EXISTS idx_linggong_contracts_application_id ON linggong_contracts(application_id);
CREATE INDEX IF NOT EXISTS idx_linggong_contracts_employer_id ON linggong_contracts(employer_id);
CREATE INDEX IF NOT EXISTS idx_linggong_contracts_worker_id ON linggong_contracts(worker_id);
CREATE INDEX IF NOT EXISTS idx_linggong_contracts_signed_at ON linggong_contracts(signed_at);
CREATE INDEX IF NOT EXISTS idx_linggong_contracts_status ON linggong_contracts(status);
CREATE INDEX IF NOT EXISTS idx_linggong_contracts_deleted_at ON linggong_contracts(deleted_at);
COMMENT ON TABLE linggong_contracts IS '电子合同表（兼职/临时/任务/服务/实习/项目/外包 + 电子签 + 多种签署）';

-- ------------------------------------------------------------
-- 6. linggong_payments 薪资支付表（对标兼职猫日结 + 支付中台）
--    薪资日结：T+0/T+1/T+7 多种结算方式 + 工资单/支付记录
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS linggong_payments (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    payment_no VARCHAR(64) NOT NULL,
    linggong_id BIGINT NOT NULL,
    task_id BIGINT NOT NULL DEFAULT 0,
    application_id BIGINT NOT NULL DEFAULT 0,
    contract_id BIGINT NOT NULL DEFAULT 0,
    employer_id BIGINT NOT NULL,
    employer_name VARCHAR(128) NOT NULL DEFAULT '',
    worker_id BIGINT NOT NULL,
    worker_name VARCHAR(50) NOT NULL DEFAULT '',
    worker_phone VARCHAR(20) NOT NULL DEFAULT '',
    worker_bank_account VARCHAR(64) NOT NULL DEFAULT '',
    worker_alipay VARCHAR(128) NOT NULL DEFAULT '',
    worker_wechat VARCHAR(128) NOT NULL DEFAULT '',
    payment_type VARCHAR(32) NOT NULL DEFAULT 'salary',
    amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    work_hours DECIMAL(8,2) NOT NULL DEFAULT 0,
    work_days INT NOT NULL DEFAULT 0,
    task_count INT NOT NULL DEFAULT 0,
    unit_price DECIMAL(12,2) NOT NULL DEFAULT 0,
    quantity DECIMAL(10,2) NOT NULL DEFAULT 0,
    settlement VARCHAR(16) NOT NULL DEFAULT 'T+1',
    settlement_status INT NOT NULL DEFAULT 0,
    settlement_at TIMESTAMPTZ,
    due_at TIMESTAMPTZ,
    pay_method VARCHAR(32) NOT NULL DEFAULT 'wechat',
    pay_trade_no VARCHAR(128) NOT NULL DEFAULT '',
    pay_channel VARCHAR(32) NOT NULL DEFAULT '',
    payee_name VARCHAR(50) NOT NULL DEFAULT '',
    platform_fee DECIMAL(12,2) NOT NULL DEFAULT 0,
    tax_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    actual_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 0,
    failed_reason VARCHAR(500) NOT NULL DEFAULT '',
    paid_at TIMESTAMPTZ,
    confirmed_at TIMESTAMPTZ,
    canceled_at TIMESTAMPTZ,
    work_start_date DATE,
    work_end_date DATE,
    evidence_images JSONB,
    remark TEXT NOT NULL DEFAULT '',
    invoice_url VARCHAR(255) NOT NULL DEFAULT '',
    CONSTRAINT uniq_linggong_payments_no UNIQUE (payment_no)
);
CREATE INDEX IF NOT EXISTS idx_linggong_payments_region_id ON linggong_payments(region_id);
CREATE INDEX IF NOT EXISTS idx_linggong_payments_linggong_id ON linggong_payments(linggong_id);
CREATE INDEX IF NOT EXISTS idx_linggong_payments_task_id ON linggong_payments(task_id);
CREATE INDEX IF NOT EXISTS idx_linggong_payments_application_id ON linggong_payments(application_id);
CREATE INDEX IF NOT EXISTS idx_linggong_payments_contract_id ON linggong_payments(contract_id);
CREATE INDEX IF NOT EXISTS idx_linggong_payments_employer_id ON linggong_payments(employer_id);
CREATE INDEX IF NOT EXISTS idx_linggong_payments_worker_id ON linggong_payments(worker_id);
CREATE INDEX IF NOT EXISTS idx_linggong_payments_payment_type ON linggong_payments(payment_type);
CREATE INDEX IF NOT EXISTS idx_linggong_payments_amount ON linggong_payments(amount);
CREATE INDEX IF NOT EXISTS idx_linggong_payments_settlement ON linggong_payments(settlement);
CREATE INDEX IF NOT EXISTS idx_linggong_payments_settlement_status ON linggong_payments(settlement_status);
CREATE INDEX IF NOT EXISTS idx_linggong_payments_pay_method ON linggong_payments(pay_method);
CREATE INDEX IF NOT EXISTS idx_linggong_payments_pay_trade_no ON linggong_payments(pay_trade_no);
CREATE INDEX IF NOT EXISTS idx_linggong_payments_status ON linggong_payments(status);
CREATE INDEX IF NOT EXISTS idx_linggong_payments_paid_at ON linggong_payments(paid_at);
CREATE INDEX IF NOT EXISTS idx_linggong_payments_due_at ON linggong_payments(due_at);
CREATE INDEX IF NOT EXISTS idx_linggong_payments_deleted_at ON linggong_payments(deleted_at);
COMMENT ON TABLE linggong_payments IS '薪资支付表（日结 T+0/T+1/T+7 + 工资单 + 支付记录）';

-- ------------------------------------------------------------
-- 7. linggong_ratings 双向评价表（对标美团/斗米）
--    工人评价雇主 + 雇主评价工人 + 多维度评分
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS linggong_ratings (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    rating_no VARCHAR(64) NOT NULL,
    linggong_id BIGINT NOT NULL,
    task_id BIGINT NOT NULL DEFAULT 0,
    application_id BIGINT NOT NULL DEFAULT 0,
    contract_id BIGINT NOT NULL DEFAULT 0,
    payment_id BIGINT NOT NULL DEFAULT 0,
    rater_type VARCHAR(16) NOT NULL DEFAULT 'employer',
    rater_id BIGINT NOT NULL,
    rater_name VARCHAR(50) NOT NULL DEFAULT '',
    rater_avatar VARCHAR(255) NOT NULL DEFAULT '',
    target_type VARCHAR(32) NOT NULL DEFAULT 'worker',
    target_id BIGINT NOT NULL,
    target_name VARCHAR(128) NOT NULL DEFAULT '',
    rating INT NOT NULL DEFAULT 5,
    content TEXT NOT NULL DEFAULT '',
    images JSONB,
    video_url VARCHAR(255) NOT NULL DEFAULT '',
    is_anonymous BOOLEAN NOT NULL DEFAULT FALSE,
    is_recommended VARCHAR(16) NOT NULL DEFAULT 'yes',
    tags JSONB,
    deal_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    work_quality INT NOT NULL DEFAULT 5,
    punctuality INT NOT NULL DEFAULT 5,
    communication INT NOT NULL DEFAULT 5,
    attitude INT NOT NULL DEFAULT 5,
    professionalism INT NOT NULL DEFAULT 5,
    payment_timeliness INT NOT NULL DEFAULT 5,
    work_environment INT NOT NULL DEFAULT 5,
    salary_match INT NOT NULL DEFAULT 5,
    reply TEXT NOT NULL DEFAULT '',
    reply_at TIMESTAMPTZ,
    append_content TEXT NOT NULL DEFAULT '',
    append_images JSONB,
    append_at TIMESTAMPTZ,
    like_count INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 1,
    rejected_reason VARCHAR(500) NOT NULL DEFAULT '',
    evaluated_at TIMESTAMPTZ,
    CONSTRAINT uniq_linggong_ratings_no UNIQUE (rating_no)
);
CREATE INDEX IF NOT EXISTS idx_linggong_ratings_region_id ON linggong_ratings(region_id);
CREATE INDEX IF NOT EXISTS idx_linggong_ratings_linggong_id ON linggong_ratings(linggong_id);
CREATE INDEX IF NOT EXISTS idx_linggong_ratings_task_id ON linggong_ratings(task_id);
CREATE INDEX IF NOT EXISTS idx_linggong_ratings_application_id ON linggong_ratings(application_id);
CREATE INDEX IF NOT EXISTS idx_linggong_ratings_rater_type ON linggong_ratings(rater_type);
CREATE INDEX IF NOT EXISTS idx_linggong_ratings_rater_id ON linggong_ratings(rater_id);
CREATE INDEX IF NOT EXISTS idx_linggong_ratings_target_type ON linggong_ratings(target_type);
CREATE INDEX IF NOT EXISTS idx_linggong_ratings_target_id ON linggong_ratings(target_id);
CREATE INDEX IF NOT EXISTS idx_linggong_ratings_rating ON linggong_ratings(rating);
CREATE INDEX IF NOT EXISTS idx_linggong_ratings_is_recommended ON linggong_ratings(is_recommended);
CREATE INDEX IF NOT EXISTS idx_linggong_ratings_status ON linggong_ratings(status);
CREATE INDEX IF NOT EXISTS idx_linggong_ratings_evaluated_at ON linggong_ratings(evaluated_at);
CREATE INDEX IF NOT EXISTS idx_linggong_ratings_deleted_at ON linggong_ratings(deleted_at);
COMMENT ON TABLE linggong_ratings IS '双向评价表（工人评雇主 + 雇主评工人 + 多维度）';

-- ------------------------------------------------------------
-- 8. linggong_skills 技能标签表（对标猪八戒威客）
--    技能分类 + 认证 + 评分 + 关联岗位
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS linggong_skills (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    name VARCHAR(64) NOT NULL,
    code VARCHAR(64) NOT NULL DEFAULT '',
    category VARCHAR(32) NOT NULL DEFAULT 'other',
    parent_id BIGINT NOT NULL DEFAULT 0,
    level INT NOT NULL DEFAULT 1,
    icon VARCHAR(255) NOT NULL DEFAULT '',
    color VARCHAR(32) NOT NULL DEFAULT '',
    description VARCHAR(500) NOT NULL DEFAULT '',
    worker_count INT NOT NULL DEFAULT 0,
    linggong_count INT NOT NULL DEFAULT 0,
    avg_salary DECIMAL(12,2) NOT NULL DEFAULT 0,
    hot_score INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 1,
    sort INT NOT NULL DEFAULT 0,
    CONSTRAINT uniq_linggong_skills_name UNIQUE (name)
);
CREATE INDEX IF NOT EXISTS idx_linggong_skills_name ON linggong_skills(name);
CREATE INDEX IF NOT EXISTS idx_linggong_skills_code ON linggong_skills(code);
CREATE INDEX IF NOT EXISTS idx_linggong_skills_category ON linggong_skills(category);
CREATE INDEX IF NOT EXISTS idx_linggong_skills_parent_id ON linggong_skills(parent_id);
CREATE INDEX IF NOT EXISTS idx_linggong_skills_hot_score ON linggong_skills(hot_score);
CREATE INDEX IF NOT EXISTS idx_linggong_skills_status ON linggong_skills(status);
CREATE INDEX IF NOT EXISTS idx_linggong_skills_sort ON linggong_skills(sort);
CREATE INDEX IF NOT EXISTS idx_linggong_skills_deleted_at ON linggong_skills(deleted_at);
COMMENT ON TABLE linggong_skills IS '技能标签表（分类 + 认证 + 评分）';

-- ------------------------------------------------------------
-- 9. linggong_credits 信用分表（对标芝麻信用/猪八戒）
--    履约 +10 / 违约 -20 / 影响接单 + 历史变更记录
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS linggong_credits (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id BIGINT NOT NULL,
    user_type VARCHAR(16) NOT NULL DEFAULT 'worker',
    reason VARCHAR(32) NOT NULL DEFAULT 'manual',
    change_type VARCHAR(16) NOT NULL DEFAULT 'add',
    change_score INT NOT NULL DEFAULT 0,
    before_score INT NOT NULL DEFAULT 0,
    after_score INT NOT NULL DEFAULT 0,
    linggong_id BIGINT NOT NULL DEFAULT 0,
    task_id BIGINT NOT NULL DEFAULT 0,
    application_id BIGINT NOT NULL DEFAULT 0,
    rating_id BIGINT NOT NULL DEFAULT 0,
    operator_id BIGINT NOT NULL DEFAULT 0,
    operator_name VARCHAR(50) NOT NULL DEFAULT '',
    description VARCHAR(500) NOT NULL DEFAULT '',
    evidence_url VARCHAR(255) NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_linggong_credits_region_id ON linggong_credits(region_id);
CREATE INDEX IF NOT EXISTS idx_linggong_credits_user_id ON linggong_credits(user_id);
CREATE INDEX IF NOT EXISTS idx_linggong_credits_user_type ON linggong_credits(user_type);
CREATE INDEX IF NOT EXISTS idx_linggong_credits_reason ON linggong_credits(reason);
CREATE INDEX IF NOT EXISTS idx_linggong_credits_after_score ON linggong_credits(after_score);
CREATE INDEX IF NOT EXISTS idx_linggong_credits_linggong_id ON linggong_credits(linggong_id);
CREATE INDEX IF NOT EXISTS idx_linggong_credits_created_at ON linggong_credits(created_at);
CREATE INDEX IF NOT EXISTS idx_linggong_credits_deleted_at ON linggong_credits(deleted_at);
COMMENT ON TABLE linggong_credits IS '信用分表（履约+10/违约-20 + 历史变更）';

-- ------------------------------------------------------------
-- 10. linggong_certifications 资质证书表（对标猪八戒/斗米认证）
--     求职者证书 + 雇主认证 + 平台认证
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS linggong_certifications (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    cert_no VARCHAR(64) NOT NULL,
    user_id BIGINT NOT NULL,
    user_type VARCHAR(16) NOT NULL DEFAULT 'worker',
    worker_id BIGINT NOT NULL DEFAULT 0,
    employer_id BIGINT NOT NULL DEFAULT 0,
    cert_type VARCHAR(32) NOT NULL DEFAULT 'id_card',
    cert_name VARCHAR(128) NOT NULL,
    cert_code VARCHAR(128) NOT NULL DEFAULT '',
    issuer_name VARCHAR(128) NOT NULL DEFAULT '',
    issuer_type VARCHAR(32) NOT NULL DEFAULT 'other',
    issue_date DATE,
    valid_from DATE,
    valid_until DATE,
    image_url VARCHAR(255) NOT NULL DEFAULT '',
    image_back_url VARCHAR(255) NOT NULL DEFAULT '',
    skill_id BIGINT NOT NULL DEFAULT 0,
    skill_name VARCHAR(64) NOT NULL DEFAULT '',
    level VARCHAR(32) NOT NULL DEFAULT '',
    score DECIMAL(5,2) NOT NULL DEFAULT 0,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    verified_at TIMESTAMPTZ,
    verified_by BIGINT NOT NULL DEFAULT 0,
    verified_by_name VARCHAR(50) NOT NULL DEFAULT '',
    status INT NOT NULL DEFAULT 0,
    reject_reason VARCHAR(500) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    CONSTRAINT uniq_linggong_certifications_no UNIQUE (cert_no)
);
CREATE INDEX IF NOT EXISTS idx_linggong_certifications_region_id ON linggong_certifications(region_id);
CREATE INDEX IF NOT EXISTS idx_linggong_certifications_user_id ON linggong_certifications(user_id);
CREATE INDEX IF NOT EXISTS idx_linggong_certifications_user_type ON linggong_certifications(user_type);
CREATE INDEX IF NOT EXISTS idx_linggong_certifications_worker_id ON linggong_certifications(worker_id);
CREATE INDEX IF NOT EXISTS idx_linggong_certifications_employer_id ON linggong_certifications(employer_id);
CREATE INDEX IF NOT EXISTS idx_linggong_certifications_cert_type ON linggong_certifications(cert_type);
CREATE INDEX IF NOT EXISTS idx_linggong_certifications_valid_until ON linggong_certifications(valid_until);
CREATE INDEX IF NOT EXISTS idx_linggong_certifications_verified ON linggong_certifications(verified) WHERE verified = TRUE;
CREATE INDEX IF NOT EXISTS idx_linggong_certifications_status ON linggong_certifications(status);
CREATE INDEX IF NOT EXISTS idx_linggong_certifications_deleted_at ON linggong_certifications(deleted_at);
COMMENT ON TABLE linggong_certifications IS '资质证书表（求职者证书 + 雇主认证 + 平台认证）';

-- ------------------------------------------------------------
-- 11. linggong_attendances 考勤打卡表（对标钉钉/企业微信考勤）
--     GPS 定位 + WiFi + 人脸 + 工时统计
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS linggong_attendances (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    attendance_no VARCHAR(64) NOT NULL,
    linggong_id BIGINT NOT NULL,
    task_id BIGINT NOT NULL DEFAULT 0,
    application_id BIGINT NOT NULL DEFAULT 0,
    contract_id BIGINT NOT NULL DEFAULT 0,
    employer_id BIGINT NOT NULL,
    worker_id BIGINT NOT NULL,
    worker_name VARCHAR(50) NOT NULL DEFAULT '',
    attendance_type VARCHAR(16) NOT NULL DEFAULT 'clock_in',
    clock_method VARCHAR(16) NOT NULL DEFAULT 'gps',
    clock_time TIMESTAMPTZ NOT NULL,
    clock_date DATE,
    latitude DECIMAL(10,7) NOT NULL DEFAULT 0,
    longitude DECIMAL(10,7) NOT NULL DEFAULT 0,
    address VARCHAR(500) NOT NULL DEFAULT '',
    wifi_name VARCHAR(128) NOT NULL DEFAULT '',
    wifi_mac VARCHAR(64) NOT NULL DEFAULT '',
    face_image_url VARCHAR(255) NOT NULL DEFAULT '',
    qr_code_content VARCHAR(255) NOT NULL DEFAULT '',
    status INT NOT NULL DEFAULT 0,
    late_minutes INT NOT NULL DEFAULT 0,
    early_minutes INT NOT NULL DEFAULT 0,
    work_hours DECIMAL(8,2) NOT NULL DEFAULT 0,
    overtime_hours DECIMAL(8,2) NOT NULL DEFAULT 0,
    break_hours DECIMAL(8,2) NOT NULL DEFAULT 0,
    task_count INT NOT NULL DEFAULT 0,
    remark TEXT NOT NULL DEFAULT '',
    approved BOOLEAN NOT NULL DEFAULT FALSE,
    approved_by BIGINT NOT NULL DEFAULT 0,
    approved_at TIMESTAMPTZ,
    CONSTRAINT uniq_linggong_attendances_no UNIQUE (attendance_no)
);
CREATE INDEX IF NOT EXISTS idx_linggong_attendances_region_id ON linggong_attendances(region_id);
CREATE INDEX IF NOT EXISTS idx_linggong_attendances_linggong_id ON linggong_attendances(linggong_id);
CREATE INDEX IF NOT EXISTS idx_linggong_attendances_task_id ON linggong_attendances(task_id);
CREATE INDEX IF NOT EXISTS idx_linggong_attendances_application_id ON linggong_attendances(application_id);
CREATE INDEX IF NOT EXISTS idx_linggong_attendances_contract_id ON linggong_attendances(contract_id);
CREATE INDEX IF NOT EXISTS idx_linggong_attendances_employer_id ON linggong_attendances(employer_id);
CREATE INDEX IF NOT EXISTS idx_linggong_attendances_worker_id ON linggong_attendances(worker_id);
CREATE INDEX IF NOT EXISTS idx_linggong_attendances_attendance_type ON linggong_attendances(attendance_type);
CREATE INDEX IF NOT EXISTS idx_linggong_attendances_clock_time ON linggong_attendances(clock_time);
CREATE INDEX IF NOT EXISTS idx_linggong_attendances_clock_date ON linggong_attendances(clock_date);
CREATE INDEX IF NOT EXISTS idx_linggong_attendances_status ON linggong_attendances(status);
CREATE INDEX IF NOT EXISTS idx_linggong_attendances_approved ON linggong_attendances(approved) WHERE approved = FALSE;
CREATE INDEX IF NOT EXISTS idx_linggong_attendances_deleted_at ON linggong_attendances(deleted_at);
COMMENT ON TABLE linggong_attendances IS '考勤打卡表（GPS + WiFi + 人脸 + 工时统计）';

-- ------------------------------------------------------------
-- 12. linggong_disputes 纠纷表（对标闲鱼/瓜子）
--     工单 + 证据 + 调解 + 仲裁
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS linggong_disputes (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    dispute_no VARCHAR(64) NOT NULL,
    linggong_id BIGINT NOT NULL,
    task_id BIGINT NOT NULL DEFAULT 0,
    application_id BIGINT NOT NULL DEFAULT 0,
    contract_id BIGINT NOT NULL DEFAULT 0,
    payment_id BIGINT NOT NULL DEFAULT 0,
    dispute_type VARCHAR(32) NOT NULL DEFAULT 'other',
    applicant_type VARCHAR(16) NOT NULL DEFAULT 'worker',
    applicant_id BIGINT NOT NULL,
    applicant_name VARCHAR(50) NOT NULL DEFAULT '',
    respondent_id BIGINT NOT NULL,
    respondent_name VARCHAR(50) NOT NULL DEFAULT '',
    title VARCHAR(200) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    evidence_images JSONB,
    evidence_videos JSONB,
    evidence_docs JSONB,
    claim_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 0,
    handler_id BIGINT NOT NULL DEFAULT 0,
    handler_name VARCHAR(50) NOT NULL DEFAULT '',
    mediation_result TEXT NOT NULL DEFAULT '',
    arbitration_result TEXT NOT NULL DEFAULT '',
    final_result VARCHAR(32) NOT NULL DEFAULT '',
    compensation_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    sla_deadline TIMESTAMPTZ,
    handled_at TIMESTAMPTZ,
    resolved_at TIMESTAMPTZ,
    closed_at TIMESTAMPTZ,
    appeal_reason TEXT NOT NULL DEFAULT '',
    appealed_at TIMESTAMPTZ,
    appeal_result TEXT NOT NULL DEFAULT '',
    appeal_handler_id BIGINT NOT NULL DEFAULT 0,
    appeal_handled_at TIMESTAMPTZ,
    CONSTRAINT uniq_linggong_disputes_no UNIQUE (dispute_no)
);
CREATE INDEX IF NOT EXISTS idx_linggong_disputes_region_id ON linggong_disputes(region_id);
CREATE INDEX IF NOT EXISTS idx_linggong_disputes_linggong_id ON linggong_disputes(linggong_id);
CREATE INDEX IF NOT EXISTS idx_linggong_disputes_task_id ON linggong_disputes(task_id);
CREATE INDEX IF NOT EXISTS idx_linggong_disputes_application_id ON linggong_disputes(application_id);
CREATE INDEX IF NOT EXISTS idx_linggong_disputes_contract_id ON linggong_disputes(contract_id);
CREATE INDEX IF NOT EXISTS idx_linggong_disputes_payment_id ON linggong_disputes(payment_id);
CREATE INDEX IF NOT EXISTS idx_linggong_disputes_dispute_type ON linggong_disputes(dispute_type);
CREATE INDEX IF NOT EXISTS idx_linggong_disputes_applicant_type ON linggong_disputes(applicant_type);
CREATE INDEX IF NOT EXISTS idx_linggong_disputes_applicant_id ON linggong_disputes(applicant_id);
CREATE INDEX IF NOT EXISTS idx_linggong_disputes_respondent_id ON linggong_disputes(respondent_id);
CREATE INDEX IF NOT EXISTS idx_linggong_disputes_status ON linggong_disputes(status);
CREATE INDEX IF NOT EXISTS idx_linggong_disputes_handler_id ON linggong_disputes(handler_id);
CREATE INDEX IF NOT EXISTS idx_linggong_disputes_sla_deadline ON linggong_disputes(sla_deadline);
CREATE INDEX IF NOT EXISTS idx_linggong_disputes_resolved_at ON linggong_disputes(resolved_at);
CREATE INDEX IF NOT EXISTS idx_linggong_disputes_deleted_at ON linggong_disputes(deleted_at);
COMMENT ON TABLE linggong_disputes IS '纠纷表（工单 + 证据 + 调解 + 仲裁 + 申诉）';

-- ------------------------------------------------------------
-- 13. linggong_withdrawals 提现表（对接支付中台）
--     求职者提现 + 雇主提现 + 提现状态机
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS linggong_withdrawals (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    withdrawal_no VARCHAR(64) NOT NULL,
    user_id BIGINT NOT NULL,
    user_type VARCHAR(16) NOT NULL DEFAULT 'worker',
    user_name VARCHAR(50) NOT NULL DEFAULT '',
    user_phone VARCHAR(20) NOT NULL DEFAULT '',
    amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    fee DECIMAL(12,2) NOT NULL DEFAULT 0,
    tax DECIMAL(12,2) NOT NULL DEFAULT 0,
    actual_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    balance_before DECIMAL(12,2) NOT NULL DEFAULT 0,
    balance_after DECIMAL(12,2) NOT NULL DEFAULT 0,
    method VARCHAR(16) NOT NULL DEFAULT 'wechat',
    payee_name VARCHAR(50) NOT NULL DEFAULT '',
    payee_account VARCHAR(128) NOT NULL DEFAULT '',
    payee_bank VARCHAR(64) NOT NULL DEFAULT '',
    payee_bank_branch VARCHAR(128) NOT NULL DEFAULT '',
    bank_card_no VARCHAR(64) NOT NULL DEFAULT '',
    alipay_account VARCHAR(128) NOT NULL DEFAULT '',
    wechat_account VARCHAR(128) NOT NULL DEFAULT '',
    status INT NOT NULL DEFAULT 0,
    failed_reason VARCHAR(500) NOT NULL DEFAULT '',
    reviewed_by BIGINT NOT NULL DEFAULT 0,
    reviewed_by_name VARCHAR(50) NOT NULL DEFAULT '',
    reviewed_at TIMESTAMPTZ,
    reviewed_remark VARCHAR(500) NOT NULL DEFAULT '',
    pay_trade_no VARCHAR(128) NOT NULL DEFAULT '',
    pay_channel VARCHAR(32) NOT NULL DEFAULT '',
    processed_at TIMESTAMPTZ,
    succeeded_at TIMESTAMPTZ,
    canceled_at TIMESTAMPTZ,
    estimated_arrival TIMESTAMPTZ,
    remark TEXT NOT NULL DEFAULT '',
    CONSTRAINT uniq_linggong_withdrawals_no UNIQUE (withdrawal_no)
);
CREATE INDEX IF NOT EXISTS idx_linggong_withdrawals_region_id ON linggong_withdrawals(region_id);
CREATE INDEX IF NOT EXISTS idx_linggong_withdrawals_user_id ON linggong_withdrawals(user_id);
CREATE INDEX IF NOT EXISTS idx_linggong_withdrawals_user_type ON linggong_withdrawals(user_type);
CREATE INDEX IF NOT EXISTS idx_linggong_withdrawals_amount ON linggong_withdrawals(amount);
CREATE INDEX IF NOT EXISTS idx_linggong_withdrawals_method ON linggong_withdrawals(method);
CREATE INDEX IF NOT EXISTS idx_linggong_withdrawals_pay_trade_no ON linggong_withdrawals(pay_trade_no);
CREATE INDEX IF NOT EXISTS idx_linggong_withdrawals_status ON linggong_withdrawals(status);
CREATE INDEX IF NOT EXISTS idx_linggong_withdrawals_reviewed_at ON linggong_withdrawals(reviewed_at);
CREATE INDEX IF NOT EXISTS idx_linggong_withdrawals_succeeded_at ON linggong_withdrawals(succeeded_at);
CREATE INDEX IF NOT EXISTS idx_linggong_withdrawals_deleted_at ON linggong_withdrawals(deleted_at);
COMMENT ON TABLE linggong_withdrawals IS '提现表（求职者 + 雇主 + 状态机）';

-- ------------------------------------------------------------
-- 14. linggong_recommendations 推荐岗位表（对标斗米智能推荐 + 猪八戒威客匹配）
--     岗位推荐 + 求职者推荐 + 热门推荐
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS linggong_recommendations (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id BIGINT NOT NULL,
    linggong_id BIGINT NOT NULL,
    rec_type VARCHAR(32) NOT NULL DEFAULT 'linggong_to_worker',
    source VARCHAR(32) NOT NULL DEFAULT 'ai',
    score DECIMAL(5,2) NOT NULL DEFAULT 0,
    reason VARCHAR(500) NOT NULL DEFAULT '',
    salary_match DECIMAL(5,2) NOT NULL DEFAULT 0,
    skill_match DECIMAL(5,2) NOT NULL DEFAULT 0,
    location_match DECIMAL(5,2) NOT NULL DEFAULT 0,
    time_match DECIMAL(5,2) NOT NULL DEFAULT 0,
    credit_match DECIMAL(5,2) NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 0,
    clicked_at TIMESTAMPTZ,
    applied_at TIMESTAMPTZ,
    viewed_at TIMESTAMPTZ,
    dismissed_at TIMESTAMPTZ,
    expired_at TIMESTAMPTZ,
    CONSTRAINT uniq_linggong_recs_user_target_type UNIQUE (user_id, linggong_id, rec_type)
);
CREATE INDEX IF NOT EXISTS idx_linggong_recommendations_region_id ON linggong_recommendations(region_id);
CREATE INDEX IF NOT EXISTS idx_linggong_recommendations_user_id ON linggong_recommendations(user_id);
CREATE INDEX IF NOT EXISTS idx_linggong_recommendations_linggong_id ON linggong_recommendations(linggong_id);
CREATE INDEX IF NOT EXISTS idx_linggong_recommendations_rec_type ON linggong_recommendations(rec_type);
CREATE INDEX IF NOT EXISTS idx_linggong_recommendations_source ON linggong_recommendations(source);
CREATE INDEX IF NOT EXISTS idx_linggong_recommendations_score ON linggong_recommendations(score);
CREATE INDEX IF NOT EXISTS idx_linggong_recommendations_status ON linggong_recommendations(status);
CREATE INDEX IF NOT EXISTS idx_linggong_recommendations_deleted_at ON linggong_recommendations(deleted_at);
COMMENT ON TABLE linggong_recommendations IS '推荐岗位表（智能推荐 + 热门推荐 + 匹配度）';

-- ------------------------------------------------------------
-- 15. linggong_favorites 收藏表
--     求职者收藏岗位 + 雇主收藏求职者
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS linggong_favorites (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id BIGINT NOT NULL,
    target_id BIGINT NOT NULL,
    favorite_type VARCHAR(32) NOT NULL DEFAULT 'linggong',
    remark VARCHAR(200) NOT NULL DEFAULT '',
    notify_on_update BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT uniq_linggong_favorites_user_target_type UNIQUE (user_id, target_id, favorite_type)
);
CREATE INDEX IF NOT EXISTS idx_linggong_favorites_user_id ON linggong_favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_linggong_favorites_target_id ON linggong_favorites(target_id);
CREATE INDEX IF NOT EXISTS idx_linggong_favorites_favorite_type ON linggong_favorites(favorite_type);
CREATE INDEX IF NOT EXISTS idx_linggong_favorites_deleted_at ON linggong_favorites(deleted_at);
COMMENT ON TABLE linggong_favorites IS '收藏表（岗位收藏 + 求职者收藏 + 搜索条件收藏）';

-- ------------------------------------------------------------
-- 16. linggong_audit_rules 审核规则表（全局，无 region_id）
--     M 端规则管理 + 内部审核检查能力
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS linggong_audit_rules (
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
CREATE INDEX IF NOT EXISTS idx_linggong_audit_rules_rule_type ON linggong_audit_rules(rule_type);
CREATE INDEX IF NOT EXISTS idx_linggong_audit_rules_rule_key ON linggong_audit_rules(rule_key);
CREATE INDEX IF NOT EXISTS idx_linggong_audit_rules_severity ON linggong_audit_rules(severity);
CREATE INDEX IF NOT EXISTS idx_linggong_audit_rules_status ON linggong_audit_rules(status);
CREATE INDEX IF NOT EXISTS idx_linggong_audit_rules_sort ON linggong_audit_rules(sort);
CREATE INDEX IF NOT EXISTS idx_linggong_audit_rules_deleted_at ON linggong_audit_rules(deleted_at);
COMMENT ON TABLE linggong_audit_rules IS '审核规则表（敏感词/薪资异常/频率限制/虚假岗位/黑名单）';

-- ------------------------------------------------------------
-- 17. linggong_statistics 数据统计表
--     岗位/雇主/求职者/技能/地区/平台多维度统计
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS linggong_statistics (
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
    contact_count INT NOT NULL DEFAULT 0,
    application_count INT NOT NULL DEFAULT 0,
    hired_count INT NOT NULL DEFAULT 0,
    completed_count INT NOT NULL DEFAULT 0,
    deal_count INT NOT NULL DEFAULT 0,
    conversion_rate DECIMAL(5,2) NOT NULL DEFAULT 0,
    total_salary DECIMAL(12,2) NOT NULL DEFAULT 0,
    avg_salary DECIMAL(12,2) NOT NULL DEFAULT 0,
    avg_deal_days INT NOT NULL DEFAULT 0,
    CONSTRAINT uniq_linggong_stats_date_type_target UNIQUE (stat_date, stat_type, target_id)
);
CREATE INDEX IF NOT EXISTS idx_linggong_statistics_region_id ON linggong_statistics(region_id);
CREATE INDEX IF NOT EXISTS idx_linggong_statistics_stat_date ON linggong_statistics(stat_date);
CREATE INDEX IF NOT EXISTS idx_linggong_statistics_stat_type ON linggong_statistics(stat_type);
CREATE INDEX IF NOT EXISTS idx_linggong_statistics_target_id ON linggong_statistics(target_id);
CREATE INDEX IF NOT EXISTS idx_linggong_statistics_deleted_at ON linggong_statistics(deleted_at);
COMMENT ON TABLE linggong_statistics IS '数据统计表（多维度：岗位/雇主/求职者/技能/地区/平台）';

-- ============================================================
-- 第三部分：updated_at 触发器
-- 依赖：001_p0_baseline.sql 已创建 update_updated_at_column() 函数
-- ============================================================

CREATE TRIGGER trg_linggong_tasks_updated
    BEFORE UPDATE ON linggong_tasks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_linggong_applications_updated
    BEFORE UPDATE ON linggong_applications
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_linggong_employers_updated
    BEFORE UPDATE ON linggong_employers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_linggong_workers_updated
    BEFORE UPDATE ON linggong_workers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_linggong_contracts_updated
    BEFORE UPDATE ON linggong_contracts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_linggong_payments_updated
    BEFORE UPDATE ON linggong_payments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_linggong_ratings_updated
    BEFORE UPDATE ON linggong_ratings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_linggong_skills_updated
    BEFORE UPDATE ON linggong_skills
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_linggong_credits_updated
    BEFORE UPDATE ON linggong_credits
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_linggong_certifications_updated
    BEFORE UPDATE ON linggong_certifications
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_linggong_attendances_updated
    BEFORE UPDATE ON linggong_attendances
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_linggong_disputes_updated
    BEFORE UPDATE ON linggong_disputes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_linggong_withdrawals_updated
    BEFORE UPDATE ON linggong_withdrawals
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_linggong_recommendations_updated
    BEFORE UPDATE ON linggong_recommendations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_linggong_favorites_updated
    BEFORE UPDATE ON linggong_favorites
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_linggong_audit_rules_updated
    BEFORE UPDATE ON linggong_audit_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_linggong_statistics_updated
    BEFORE UPDATE ON linggong_statistics
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- 第四部分：表注释
-- ============================================================
COMMENT ON SCHEMA public IS 'linggong 零工兼职模块（v3.2.1）：18 张表（主表 linggongs + 17 张子表）';
