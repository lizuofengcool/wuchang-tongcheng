-- ============================================================
-- love 相亲交友模块完整功能迁移脚本（v3.2.1）
-- 对标：Soul / 陌陌 / 探探 / 百合网
--
-- 内容：
--   1. ALTER TABLE loves 主表新增字段（地区/状态/审核/认证/会员/匹配统计等）
--   2. CREATE 17 张子表（love_ 前缀，依据数据库分表前缀规范 v1.0.0）
--   3. 索引、外键、触发器、注释
--   4. 全幂等：CREATE TABLE IF NOT EXISTS / ALTER TABLE ADD COLUMN IF NOT EXISTS
-- 依赖：001_p0_baseline.sql（update_updated_at_column 函数已创建）
-- 主表 loves 由 GORM AutoMigrate 创建（见 plugin.go Init）
-- ============================================================

-- ============================================================
-- 第一部分：扩展 loves 主表
-- 注意：loves 表由 GORM AutoMigrate 在应用启动时创建
--      本迁移对 loves 的所有 ALTER 操作包装在 DO 块中，
--      若表不存在则跳过（待应用启动后再执行一次本迁移即可补齐字段）
-- ============================================================
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'loves') THEN
        -- === 用户基础信息 ===
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS user_id BIGINT NOT NULL DEFAULT 0;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS nickname VARCHAR(64) NOT NULL DEFAULT '';
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS avatar VARCHAR(255) NOT NULL DEFAULT '';
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS gender SMALLINT NOT NULL DEFAULT 0;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS age INT NOT NULL DEFAULT 0;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS birthday DATE;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS height INT NOT NULL DEFAULT 0;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS weight INT NOT NULL DEFAULT 0;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS constellation VARCHAR(16) NOT NULL DEFAULT '';
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS zodiac VARCHAR(16) NOT NULL DEFAULT '';
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS hometown VARCHAR(128) NOT NULL DEFAULT '';
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS residence VARCHAR(128) NOT NULL DEFAULT '';
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS education VARCHAR(32) NOT NULL DEFAULT '';
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS occupation VARCHAR(64) NOT NULL DEFAULT '';
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS income VARCHAR(32) NOT NULL DEFAULT '';
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS marriage VARCHAR(32) NOT NULL DEFAULT '';
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS house VARCHAR(32) NOT NULL DEFAULT '';
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS car VARCHAR(32) NOT NULL DEFAULT '';
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS drinking VARCHAR(32) NOT NULL DEFAULT '';
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS smoking VARCHAR(32) NOT NULL DEFAULT '';
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS want_kids VARCHAR(32) NOT NULL DEFAULT '';
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS bio TEXT NOT NULL DEFAULT '';
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS voice_intro_url VARCHAR(255) NOT NULL DEFAULT '';
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS cover_image VARCHAR(255) NOT NULL DEFAULT '';
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS photo_verified BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS video_verified BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS education_verified BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS real_name_verified BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS status INT NOT NULL DEFAULT 1;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS audit_status INT NOT NULL DEFAULT 1;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS audit_reason VARCHAR(500) NOT NULL DEFAULT '';
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS member_level INT NOT NULL DEFAULT 0;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS member_expired_at TIMESTAMPTZ;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS credits DECIMAL(12,2) NOT NULL DEFAULT 0;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS super_likes_today INT NOT NULL DEFAULT 0;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS super_likes_reset_at TIMESTAMPTZ;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS last_active_at TIMESTAMPTZ;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS last_active_ip VARCHAR(64) NOT NULL DEFAULT '';
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS longitude DECIMAL(10,7) NOT NULL DEFAULT 0;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS latitude DECIMAL(10,7) NOT NULL DEFAULT 0;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS location_updated_at TIMESTAMPTZ;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS hide_online BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS hide_location BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS hide_age BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS hide_distance BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS only_verified_match BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS contact_price DECIMAL(12,2) NOT NULL DEFAULT 0;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS view_count INT NOT NULL DEFAULT 0;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS like_count INT NOT NULL DEFAULT 0;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS liked_count INT NOT NULL DEFAULT 0;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS match_count INT NOT NULL DEFAULT 0;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS visitor_count INT NOT NULL DEFAULT 0;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS story_count INT NOT NULL DEFAULT 0;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS gift_count INT NOT NULL DEFAULT 0;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS impression_count INT NOT NULL DEFAULT 0;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS popularity_score DECIMAL(5,2) NOT NULL DEFAULT 0;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS featured BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS picked BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS risk_score INT NOT NULL DEFAULT 0;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS content_hash VARCHAR(64) NOT NULL DEFAULT '';
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS tags JSONB;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS interests JSONB;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS personality JSONB;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS values JSONB;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS photo_urls JSONB;
        ALTER TABLE loves ADD COLUMN IF NOT EXISTS match_preferences JSONB;
        CREATE INDEX IF NOT EXISTS idx_loves_user_id ON loves(user_id);
        CREATE INDEX IF NOT EXISTS idx_loves_gender ON loves(gender);
        CREATE INDEX IF NOT EXISTS idx_loves_age ON loves(age);
        CREATE INDEX IF NOT EXISTS idx_loves_status ON loves(status);
        CREATE INDEX IF NOT EXISTS idx_loves_audit_status ON loves(audit_status);
        CREATE INDEX IF NOT EXISTS idx_loves_member_level ON loves(member_level);
        CREATE INDEX IF NOT EXISTS idx_loves_real_name_verified ON loves(real_name_verified) WHERE real_name_verified = TRUE;
        CREATE INDEX IF NOT EXISTS idx_loves_photo_verified ON loves(photo_verified) WHERE photo_verified = TRUE;
        CREATE INDEX IF NOT EXISTS idx_loves_video_verified ON loves(video_verified) WHERE video_verified = TRUE;
        CREATE INDEX IF NOT EXISTS idx_loves_last_active_at ON loves(last_active_at);
        CREATE INDEX IF NOT EXISTS idx_loves_featured ON loves(featured) WHERE featured = TRUE;
        CREATE INDEX IF NOT EXISTS idx_loves_picked ON loves(picked) WHERE picked = TRUE;
        CREATE INDEX IF NOT EXISTS idx_loves_popularity_score ON loves(popularity_score);
        CREATE INDEX IF NOT EXISTS idx_loves_residence ON loves(residence);
        CREATE INDEX IF NOT EXISTS idx_loves_hometown ON loves(hometown);
        CREATE INDEX IF NOT EXISTS idx_loves_tags ON loves USING GIN(tags);
        CREATE INDEX IF NOT EXISTS idx_loves_interests ON loves USING GIN(interests);
        COMMENT ON COLUMN loves.gender IS '性别：0未知 1男 2女';
        COMMENT ON COLUMN loves.status IS '状态：0禁用 1正常 2冻结 3注销';
        COMMENT ON COLUMN loves.audit_status IS '审核：0待审 1通过 2拒绝';
        COMMENT ON COLUMN loves.member_level IS '会员等级：0普通 1基础 2高级 3VIP 4Premium';
        COMMENT ON COLUMN loves.credits IS '金币余额（用于送礼物/解锁功能）';
        COMMENT ON COLUMN loves.super_likes_today IS '今日心动信号已使用次数（每天限量1次）';
        COMMENT ON COLUMN loves.voice_intro_url IS '语音自我介绍 URL（对标 Soul）';
        COMMENT ON COLUMN loves.tags IS '个人标签 JSONB 数组（最多 8 个）';
        COMMENT ON COLUMN loves.interests IS '兴趣标签 JSONB：{"music":["pop"],"sports":["basketball"]}';
        COMMENT ON COLUMN loves.personality IS '性格测试结果 JSONB：MBTI/九型人格等';
        COMMENT ON COLUMN loves.values IS '价值观 JSONB：婚姻观/家庭观/事业观等';
        COMMENT ON COLUMN loves.match_preferences IS '匹配偏好 JSONB：年龄范围/身高范围/地区等';
    END IF;
END $$;

-- ============================================================
-- 第二部分：17 张子表 CREATE TABLE IF NOT EXISTS
-- 表前缀 love_ 依据 docs/架构设计/数据库分表前缀规范.md
-- ============================================================

-- ------------------------------------------------------------
-- 1. love_profiles 详细资料表
--    对标 Soul/陌陌：补充扩展信息（详细自我介绍/择偶要求/语音相册）
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS love_profiles (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    love_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    nickname VARCHAR(64) NOT NULL DEFAULT '',
    avatar VARCHAR(255) NOT NULL DEFAULT '',
    gender SMALLINT NOT NULL DEFAULT 0,
    age INT NOT NULL DEFAULT 0,
    height INT NOT NULL DEFAULT 0,
    weight INT NOT NULL DEFAULT 0,
    city VARCHAR(64) NOT NULL DEFAULT '',
    district VARCHAR(64) NOT NULL DEFAULT '',
    occupation VARCHAR(64) NOT NULL DEFAULT '',
    company VARCHAR(128) NOT NULL DEFAULT '',
    industry VARCHAR(64) NOT NULL DEFAULT '',
    education VARCHAR(32) NOT NULL DEFAULT '',
    school VARCHAR(128) NOT NULL DEFAULT '',
    income VARCHAR(32) NOT NULL DEFAULT '',
    marriage VARCHAR(32) NOT NULL DEFAULT '',
    children_status VARCHAR(32) NOT NULL DEFAULT '',
    house_status VARCHAR(32) NOT NULL DEFAULT '',
    car_status VARCHAR(32) NOT NULL DEFAULT '',
    drinking VARCHAR(32) NOT NULL DEFAULT '',
    smoking VARCHAR(32) NOT NULL DEFAULT '',
    exercise VARCHAR(32) NOT NULL DEFAULT '',
    diet VARCHAR(32) NOT NULL DEFAULT '',
    sleep VARCHAR(32) NOT NULL DEFAULT '',
    pets VARCHAR(32) NOT NULL DEFAULT '',
    languages JSONB,
    interests JSONB,
    skills JSONB,
    self_intro TEXT NOT NULL DEFAULT '',
    ideal_partner TEXT NOT NULL DEFAULT '',
    ideal_age_min INT NOT NULL DEFAULT 0,
    ideal_age_max INT NOT NULL DEFAULT 0,
    ideal_height_min INT NOT NULL DEFAULT 0,
    ideal_height_max INT NOT NULL DEFAULT 0,
    ideal_cities JSONB,
    ideal_education JSONB,
    ideal_income VARCHAR(32) NOT NULL DEFAULT '',
    ideal_marriage VARCHAR(32) NOT NULL DEFAULT '',
    ideal_house VARCHAR(32) NOT NULL DEFAULT '',
    ideal_car VARCHAR(32) NOT NULL DEFAULT '',
    ideal_smoking VARCHAR(32) NOT NULL DEFAULT '',
    ideal_drinking VARCHAR(32) NOT NULL DEFAULT '',
    voice_intro_url VARCHAR(255) NOT NULL DEFAULT '',
    voice_intro_duration INT NOT NULL DEFAULT 0,
    video_intro_url VARCHAR(255) NOT NULL DEFAULT '',
    video_cover VARCHAR(255) NOT NULL DEFAULT '',
    photo_urls JSONB,
    photo_count INT NOT NULL DEFAULT 0,
    profile_score INT NOT NULL DEFAULT 0,
    completed_step INT NOT NULL DEFAULT 0,
    completed_at TIMESTAMPTZ,
    status INT NOT NULL DEFAULT 1,
    CONSTRAINT uniq_love_profiles_love_id UNIQUE (love_id)
);
CREATE INDEX IF NOT EXISTS idx_love_profiles_region_id ON love_profiles(region_id);
CREATE INDEX IF NOT EXISTS idx_love_profiles_user_id ON love_profiles(user_id);
CREATE INDEX IF NOT EXISTS idx_love_profiles_gender ON love_profiles(gender);
CREATE INDEX IF NOT EXISTS idx_love_profiles_age ON love_profiles(age);
CREATE INDEX IF NOT EXISTS idx_love_profiles_city ON love_profiles(city);
CREATE INDEX IF NOT EXISTS idx_love_profiles_status ON love_profiles(status);
CREATE INDEX IF NOT EXISTS idx_love_profiles_deleted_at ON love_profiles(deleted_at);
COMMENT ON TABLE love_profiles IS '用户详细资料表（对标 Soul/陌陌：自我介绍/择偶要求/语音视频）';

-- ------------------------------------------------------------
-- 2. love_matches 匹配记录表
--    对标探探/Soul：双向喜欢则匹配成功，开始聊天
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS love_matches (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    match_no VARCHAR(64) NOT NULL,
    user_id_a BIGINT NOT NULL,
    user_id_b BIGINT NOT NULL,
    love_id_a BIGINT NOT NULL,
    love_id_b BIGINT NOT NULL,
    match_score DECIMAL(5,2) NOT NULL DEFAULT 0,
    match_type VARCHAR(32) NOT NULL DEFAULT 'both_like',
    interest_match DECIMAL(5,2) NOT NULL DEFAULT 0,
    personality_match DECIMAL(5,2) NOT NULL DEFAULT 0,
    value_match DECIMAL(5,2) NOT NULL DEFAULT 0,
    location_match DECIMAL(5,2) NOT NULL DEFAULT 0,
    age_match DECIMAL(5,2) NOT NULL DEFAULT 0,
    matched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    chat_session_id BIGINT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 1,
    last_message_at TIMESTAMPTZ,
    last_message_content TEXT NOT NULL DEFAULT '',
    last_message_type VARCHAR(16) NOT NULL DEFAULT '',
    unread_count_a INT NOT NULL DEFAULT 0,
    unread_count_b INT NOT NULL DEFAULT 0,
    unmuted_by_a BOOLEAN NOT NULL DEFAULT FALSE,
    unmuted_by_b BOOLEAN NOT NULL DEFAULT FALSE,
    dissolved_at TIMESTAMPTZ,
    dissolve_reason VARCHAR(255) NOT NULL DEFAULT '',
    dissolve_by BIGINT NOT NULL DEFAULT 0,
    CONSTRAINT uniq_love_matches_no UNIQUE (match_no)
);
CREATE INDEX IF NOT EXISTS idx_love_matches_region_id ON love_matches(region_id);
CREATE INDEX IF NOT EXISTS idx_love_matches_user_id_a ON love_matches(user_id_a);
CREATE INDEX IF NOT EXISTS idx_love_matches_user_id_b ON love_matches(user_id_b);
CREATE INDEX IF NOT EXISTS idx_love_matches_love_id_a ON love_matches(love_id_a);
CREATE INDEX IF NOT EXISTS idx_love_matches_love_id_b ON love_matches(love_id_b);
CREATE INDEX IF NOT EXISTS idx_love_matches_status ON love_matches(status);
CREATE INDEX IF NOT EXISTS idx_love_matches_matched_at ON love_matches(matched_at);
CREATE INDEX IF NOT EXISTS idx_love_matches_last_message_at ON love_matches(last_message_at);
CREATE INDEX IF NOT EXISTS idx_love_matches_deleted_at ON love_matches(deleted_at);
COMMENT ON TABLE love_matches IS '匹配记录表（双向喜欢则匹配，含匹配维度评分）';

-- ------------------------------------------------------------
-- 3. love_stories 动态广场表
--    对标陌陌/探探：发布动态（图文/视频/语音）
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS love_stories (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    story_no VARCHAR(64) NOT NULL,
    love_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    user_nickname VARCHAR(64) NOT NULL DEFAULT '',
    user_avatar VARCHAR(255) NOT NULL DEFAULT '',
    user_gender SMALLINT NOT NULL DEFAULT 0,
    title VARCHAR(200) NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    media_type VARCHAR(16) NOT NULL DEFAULT 'image',
    image_urls JSONB,
    video_url VARCHAR(255) NOT NULL DEFAULT '',
    video_cover VARCHAR(255) NOT NULL DEFAULT '',
    video_duration INT NOT NULL DEFAULT 0,
    voice_url VARCHAR(255) NOT NULL DEFAULT '',
    voice_duration INT NOT NULL DEFAULT 0,
    location VARCHAR(128) NOT NULL DEFAULT '',
    latitude DECIMAL(10,7) NOT NULL DEFAULT 0,
    longitude DECIMAL(10,7) NOT NULL DEFAULT 0,
    tags JSONB,
    topic VARCHAR(64) NOT NULL DEFAULT '',
    view_count INT NOT NULL DEFAULT 0,
    like_count INT NOT NULL DEFAULT 0,
    comment_count INT NOT NULL DEFAULT 0,
    share_count INT NOT NULL DEFAULT 0,
    report_count INT NOT NULL DEFAULT 0,
    featured BOOLEAN NOT NULL DEFAULT FALSE,
    status INT NOT NULL DEFAULT 1,
    audit_status INT NOT NULL DEFAULT 1,
    audit_reason VARCHAR(500) NOT NULL DEFAULT '',
    published_at TIMESTAMPTZ,
    hot_score DECIMAL(8,2) NOT NULL DEFAULT 0,
    content_hash VARCHAR(64) NOT NULL DEFAULT '',
    risk_score INT NOT NULL DEFAULT 0,
    CONSTRAINT uniq_love_stories_no UNIQUE (story_no)
);
CREATE INDEX IF NOT EXISTS idx_love_stories_region_id ON love_stories(region_id);
CREATE INDEX IF NOT EXISTS idx_love_stories_love_id ON love_stories(love_id);
CREATE INDEX IF NOT EXISTS idx_love_stories_user_id ON love_stories(user_id);
CREATE INDEX IF NOT EXISTS idx_love_stories_media_type ON love_stories(media_type);
CREATE INDEX IF NOT EXISTS idx_love_stories_status ON love_stories(status);
CREATE INDEX IF NOT EXISTS idx_love_stories_audit_status ON love_stories(audit_status);
CREATE INDEX IF NOT EXISTS idx_love_stories_published_at ON love_stories(published_at);
CREATE INDEX IF NOT EXISTS idx_love_stories_hot_score ON love_stories(hot_score);
CREATE INDEX IF NOT EXISTS idx_love_stories_topic ON love_stories(topic);
CREATE INDEX IF NOT EXISTS idx_love_stories_featured ON love_stories(featured) WHERE featured = TRUE;
CREATE INDEX IF NOT EXISTS idx_love_stories_tags ON love_stories USING GIN(tags);
CREATE INDEX IF NOT EXISTS idx_love_stories_deleted_at ON love_stories(deleted_at);
COMMENT ON TABLE love_stories IS '动态广场表（图文/视频/语音动态 + 话题）';

-- ------------------------------------------------------------
-- 4. love_member_levels 会员等级定义表
--    对标陌陌/Soul：基础/高级/VIP/Premium 四级
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS love_member_levels (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    level_code VARCHAR(32) NOT NULL,
    level_name VARCHAR(64) NOT NULL,
    level INT NOT NULL DEFAULT 0,
    description TEXT NOT NULL DEFAULT '',
    icon VARCHAR(255) NOT NULL DEFAULT '',
    color VARCHAR(32) NOT NULL DEFAULT '',
    monthly_price DECIMAL(12,2) NOT NULL DEFAULT 0,
    quarterly_price DECIMAL(12,2) NOT NULL DEFAULT 0,
    yearly_price DECIMAL(12,2) NOT NULL DEFAULT 0,
    daily_super_likes INT NOT NULL DEFAULT 0,
    daily_likes INT NOT NULL DEFAULT 0,
    daily_visits INT NOT NULL DEFAULT 0,
    daily_recommendations INT NOT NULL DEFAULT 0,
    can_see_visitors BOOLEAN NOT NULL DEFAULT FALSE,
    can_see_likes BOOLEAN NOT NULL DEFAULT FALSE,
    can_hide_online BOOLEAN NOT NULL DEFAULT FALSE,
    can_hide_location BOOLEAN NOT NULL DEFAULT FALSE,
    can_filter_verified BOOLEAN NOT NULL DEFAULT FALSE,
    can_advanced_filter BOOLEAN NOT NULL DEFAULT FALSE,
    can_super_like BOOLEAN NOT NULL DEFAULT FALSE,
    can_undo_swipe BOOLEAN NOT NULL DEFAULT FALSE,
    can_boost_profile BOOLEAN NOT NULL DEFAULT FALSE,
    can_see_match_score BOOLEAN NOT NULL DEFAULT FALSE,
    perks JSONB,
    sort INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 1,
    CONSTRAINT uniq_love_member_levels_level_code UNIQUE (level_code),
    CONSTRAINT uniq_love_member_levels_level UNIQUE (level)
);
CREATE INDEX IF NOT EXISTS idx_love_member_levels_status ON love_member_levels(status);
CREATE INDEX IF NOT EXISTS idx_love_member_levels_sort ON love_member_levels(sort);
CREATE INDEX IF NOT EXISTS idx_love_member_levels_deleted_at ON love_member_levels(deleted_at);
COMMENT ON TABLE love_member_levels IS '会员等级定义表（基础/高级/VIP/Premium 四级）';

-- ------------------------------------------------------------
-- 5. love_verifications 实名/视频/学历认证表
--    对标百合网/陌陌：三重认证
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS love_verifications (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    verify_no VARCHAR(64) NOT NULL,
    love_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    type VARCHAR(32) NOT NULL DEFAULT 'real_name',
    real_name VARCHAR(64) NOT NULL DEFAULT '',
    id_card_no VARCHAR(32) NOT NULL DEFAULT '',
    id_card_front VARCHAR(255) NOT NULL DEFAULT '',
    id_card_back VARCHAR(255) NOT NULL DEFAULT '',
    id_card_hold VARCHAR(255) NOT NULL DEFAULT '',
    face_image VARCHAR(255) NOT NULL DEFAULT '',
    video_url VARCHAR(255) NOT NULL DEFAULT '',
    video_cover VARCHAR(255) NOT NULL DEFAULT '',
    video_duration INT NOT NULL DEFAULT 0,
    school_name VARCHAR(128) NOT NULL DEFAULT '',
    diploma_image VARCHAR(255) NOT NULL DEFAULT '',
    diploma_no VARCHAR(64) NOT NULL DEFAULT '',
    education VARCHAR(32) NOT NULL DEFAULT '',
    graduation_year INT NOT NULL DEFAULT 0,
    property_image VARCHAR(255) NOT NULL DEFAULT '',
    property_no VARCHAR(64) NOT NULL DEFAULT '',
    third_party_token VARCHAR(255) NOT NULL DEFAULT '',
    third_party_score DECIMAL(5,2) NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 0,
    reject_reason VARCHAR(500) NOT NULL DEFAULT '',
    verified_at TIMESTAMPTZ,
    verified_by BIGINT NOT NULL DEFAULT 0,
    verified_name VARCHAR(64) NOT NULL DEFAULT '',
    expired_at TIMESTAMPTZ,
    CONSTRAINT uniq_love_verifications_no UNIQUE (verify_no)
);
CREATE INDEX IF NOT EXISTS idx_love_verifications_region_id ON love_verifications(region_id);
CREATE INDEX IF NOT EXISTS idx_love_verifications_love_id ON love_verifications(love_id);
CREATE INDEX IF NOT EXISTS idx_love_verifications_user_id ON love_verifications(user_id);
CREATE INDEX IF NOT EXISTS idx_love_verifications_type ON love_verifications(type);
CREATE INDEX IF NOT EXISTS idx_love_verifications_status ON love_verifications(status);
CREATE INDEX IF NOT EXISTS idx_love_verifications_verified_at ON love_verifications(verified_at);
CREATE INDEX IF NOT EXISTS idx_love_verifications_deleted_at ON love_verifications(deleted_at);
COMMENT ON TABLE love_verifications IS '认证表（实名/视频/学历/房产/车辆多重认证）';

-- ------------------------------------------------------------
-- 6. love_impressions 印象标签表
--    对标陌陌/探探：他人评价（如"温柔"、"有趣"）
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS love_impressions (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    love_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    from_user_id BIGINT NOT NULL,
    from_user_nickname VARCHAR(64) NOT NULL DEFAULT '',
    from_user_avatar VARCHAR(255) NOT NULL DEFAULT '',
    tag VARCHAR(32) NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    anonymous BOOLEAN NOT NULL DEFAULT FALSE,
    match_id BIGINT NOT NULL DEFAULT 0,
    is_anonymous BOOLEAN NOT NULL DEFAULT FALSE,
    status INT NOT NULL DEFAULT 1
);
CREATE INDEX IF NOT EXISTS idx_love_impressions_region_id ON love_impressions(region_id);
CREATE INDEX IF NOT EXISTS idx_love_impressions_love_id ON love_impressions(love_id);
CREATE INDEX IF NOT EXISTS idx_love_impressions_user_id ON love_impressions(user_id);
CREATE INDEX IF NOT EXISTS idx_love_impressions_from_user_id ON love_impressions(from_user_id);
CREATE INDEX IF NOT EXISTS idx_love_impressions_tag ON love_impressions(tag);
CREATE INDEX IF NOT EXISTS idx_love_impressions_status ON love_impressions(status);
CREATE INDEX IF NOT EXISTS idx_love_impressions_deleted_at ON love_impressions(deleted_at);
COMMENT ON TABLE love_impressions IS '印象标签表（他人评价：温柔/有趣/有才华等）';

-- ------------------------------------------------------------
-- 7. love_visits 访客记录表
--    对标陌陌/探探：谁看过我
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS love_visits (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    love_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    visitor_user_id BIGINT NOT NULL,
    visitor_love_id BIGINT NOT NULL,
    visitor_nickname VARCHAR(64) NOT NULL DEFAULT '',
    visitor_avatar VARCHAR(255) NOT NULL DEFAULT '',
    visitor_gender SMALLINT NOT NULL DEFAULT 0,
    visit_type VARCHAR(32) NOT NULL DEFAULT 'profile',
    source VARCHAR(32) NOT NULL DEFAULT 'recommend',
    ip VARCHAR(64) NOT NULL DEFAULT '',
    user_agent VARCHAR(255) NOT NULL DEFAULT '',
    duration INT NOT NULL DEFAULT 0,
    is_hidden BOOLEAN NOT NULL DEFAULT FALSE,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    status INT NOT NULL DEFAULT 1,
    CONSTRAINT uniq_love_visits_user_visitor UNIQUE (user_id, visitor_user_id)
);
CREATE INDEX IF NOT EXISTS idx_love_visits_region_id ON love_visits(region_id);
CREATE INDEX IF NOT EXISTS idx_love_visits_love_id ON love_visits(love_id);
CREATE INDEX IF NOT EXISTS idx_love_visits_user_id ON love_visits(user_id);
CREATE INDEX IF NOT EXISTS idx_love_visits_visitor_user_id ON love_visits(visitor_user_id);
CREATE INDEX IF NOT EXISTS idx_love_visits_visit_type ON love_visits(visit_type);
CREATE INDEX IF NOT EXISTS idx_love_visits_created_at ON love_visits(created_at);
CREATE INDEX IF NOT EXISTS idx_love_visits_is_read ON love_visits(is_read) WHERE is_read = FALSE;
CREATE INDEX IF NOT EXISTS idx_love_visits_status ON love_visits(status);
CREATE INDEX IF NOT EXISTS idx_love_visits_deleted_at ON love_visits(deleted_at);
COMMENT ON TABLE love_visits IS '访客记录表（谁看过我 + 来源/时长/未读状态）';

-- ------------------------------------------------------------
-- 8. love_blocks 拉黑名单表
--    对标陌陌/探探：拉黑后无法匹配/聊天/查看
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS love_blocks (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id BIGINT NOT NULL,
    love_id BIGINT NOT NULL,
    blocked_user_id BIGINT NOT NULL,
    blocked_love_id BIGINT NOT NULL,
    blocked_nickname VARCHAR(64) NOT NULL DEFAULT '',
    blocked_avatar VARCHAR(255) NOT NULL DEFAULT '',
    reason VARCHAR(255) NOT NULL DEFAULT '',
    report_id BIGINT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 1,
    CONSTRAINT uniq_love_blocks_user_blocked UNIQUE (user_id, blocked_user_id)
);
CREATE INDEX IF NOT EXISTS idx_love_blocks_region_id ON love_blocks(region_id);
CREATE INDEX IF NOT EXISTS idx_love_blocks_user_id ON love_blocks(user_id);
CREATE INDEX IF NOT EXISTS idx_love_blocks_blocked_user_id ON love_blocks(blocked_user_id);
CREATE INDEX IF NOT EXISTS idx_love_blocks_love_id ON love_blocks(love_id);
CREATE INDEX IF NOT EXISTS idx_love_blocks_blocked_love_id ON love_blocks(blocked_love_id);
CREATE INDEX IF NOT EXISTS idx_love_blocks_status ON love_blocks(status);
CREATE INDEX IF NOT EXISTS idx_love_blocks_deleted_at ON love_blocks(deleted_at);
COMMENT ON TABLE love_blocks IS '拉黑名单表（拉黑后无法匹配/聊天/查看）';

-- ------------------------------------------------------------
-- 9. love_gifts 虚拟礼物表
--    对标陌陌/Soul：礼物定义（玫瑰/礼物/动画礼物）
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS love_gifts (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    gift_code VARCHAR(32) NOT NULL,
    gift_name VARCHAR(64) NOT NULL,
    category VARCHAR(32) NOT NULL DEFAULT 'common',
    description TEXT NOT NULL DEFAULT '',
    icon VARCHAR(255) NOT NULL DEFAULT '',
    animation_url VARCHAR(255) NOT NULL DEFAULT '',
    animation_type VARCHAR(32) NOT NULL DEFAULT '',
    animation_duration INT NOT NULL DEFAULT 0,
    price DECIMAL(12,2) NOT NULL DEFAULT 0,
    original_price DECIMAL(12,2) NOT NULL DEFAULT 0,
    discount_price DECIMAL(12,2) NOT NULL DEFAULT 0,
    member_level INT NOT NULL DEFAULT 0,
    charm_value INT NOT NULL DEFAULT 0,
    is_limited BOOLEAN NOT NULL DEFAULT FALSE,
    is_animated BOOLEAN NOT NULL DEFAULT FALSE,
    is_combo BOOLEAN NOT NULL DEFAULT FALSE,
    combo_min INT NOT NULL DEFAULT 0,
    combo_max INT NOT NULL DEFAULT 0,
    daily_limit INT NOT NULL DEFAULT 0,
    sort INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 1,
    start_at TIMESTAMPTZ,
    end_at TIMESTAMPTZ,
    CONSTRAINT uniq_love_gifts_code UNIQUE (gift_code)
);
CREATE INDEX IF NOT EXISTS idx_love_gifts_category ON love_gifts(category);
CREATE INDEX IF NOT EXISTS idx_love_gifts_price ON love_gifts(price);
CREATE INDEX IF NOT EXISTS idx_love_gifts_member_level ON love_gifts(member_level);
CREATE INDEX IF NOT EXISTS idx_love_gifts_status ON love_gifts(status);
CREATE INDEX IF NOT EXISTS idx_love_gifts_sort ON love_gifts(sort);
CREATE INDEX IF NOT EXISTS idx_love_gifts_deleted_at ON love_gifts(deleted_at);
COMMENT ON TABLE love_gifts IS '虚拟礼物定义表（玫瑰/动画礼物/连击礼物）';

-- ------------------------------------------------------------
-- 10. love_memberships 会员订阅表
--     对标陌陌/Soul：订阅记录（按月/季/年）
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS love_memberships (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    sub_no VARCHAR(64) NOT NULL,
    user_id BIGINT NOT NULL,
    love_id BIGINT NOT NULL,
    level_code VARCHAR(32) NOT NULL,
    level_name VARCHAR(64) NOT NULL,
    level INT NOT NULL DEFAULT 0,
    plan VARCHAR(32) NOT NULL DEFAULT 'monthly',
    period INT NOT NULL DEFAULT 1,
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    price DECIMAL(12,2) NOT NULL DEFAULT 0,
    original_price DECIMAL(12,2) NOT NULL DEFAULT 0,
    discount DECIMAL(5,2) NOT NULL DEFAULT 0,
    pay_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    pay_method VARCHAR(32) NOT NULL DEFAULT '',
    pay_order_no VARCHAR(64) NOT NULL DEFAULT '',
    pay_at TIMESTAMPTZ,
    auto_renew BOOLEAN NOT NULL DEFAULT FALSE,
    renew_count INT NOT NULL DEFAULT 0,
    perks_snapshot JSONB,
    status INT NOT NULL DEFAULT 0,
    cancel_at TIMESTAMPTZ,
    cancel_reason VARCHAR(255) NOT NULL DEFAULT '',
    refund_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    refund_at TIMESTAMPTZ,
    refund_reason VARCHAR(255) NOT NULL DEFAULT '',
    source VARCHAR(32) NOT NULL DEFAULT 'self',
    remark VARCHAR(255) NOT NULL DEFAULT '',
    CONSTRAINT uniq_love_memberships_no UNIQUE (sub_no)
);
CREATE INDEX IF NOT EXISTS idx_love_memberships_region_id ON love_memberships(region_id);
CREATE INDEX IF NOT EXISTS idx_love_memberships_user_id ON love_memberships(user_id);
CREATE INDEX IF NOT EXISTS idx_love_memberships_love_id ON love_memberships(love_id);
CREATE INDEX IF NOT EXISTS idx_love_memberships_level_code ON love_memberships(level_code);
CREATE INDEX IF NOT EXISTS idx_love_memberships_plan ON love_memberships(plan);
CREATE INDEX IF NOT EXISTS idx_love_memberships_status ON love_memberships(status);
CREATE INDEX IF NOT EXISTS idx_love_memberships_start_at ON love_memberships(start_at);
CREATE INDEX IF NOT EXISTS idx_love_memberships_end_at ON love_memberships(end_at);
CREATE INDEX IF NOT EXISTS idx_love_memberships_auto_renew ON love_memberships(auto_renew) WHERE auto_renew = TRUE;
CREATE INDEX IF NOT EXISTS idx_love_memberships_deleted_at ON love_memberships(deleted_at);
COMMENT ON TABLE love_memberships IS '会员订阅表（按月/季/年订阅记录 + 自动续费 + 退款）';

-- ------------------------------------------------------------
-- 11. love_likes 喜欢/不喜欢/心动信号表
--     对标探探/Soul：滑动卡片，双向喜欢则匹配
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS love_likes (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id BIGINT NOT NULL,
    love_id BIGINT NOT NULL,
    target_user_id BIGINT NOT NULL,
    target_love_id BIGINT NOT NULL,
    target_nickname VARCHAR(64) NOT NULL DEFAULT '',
    target_avatar VARCHAR(255) NOT NULL DEFAULT '',
    target_gender SMALLINT NOT NULL DEFAULT 0,
    action VARCHAR(16) NOT NULL DEFAULT 'like',
    super_like BOOLEAN NOT NULL DEFAULT FALSE,
    match_score DECIMAL(5,2) NOT NULL DEFAULT 0,
    source VARCHAR(32) NOT NULL DEFAULT 'recommend',
    ip VARCHAR(64) NOT NULL DEFAULT '',
    is_matched BOOLEAN NOT NULL DEFAULT FALSE,
    match_id BIGINT NOT NULL DEFAULT 0,
    matched_at TIMESTAMPTZ,
    undone_at TIMESTAMPTZ,
    undo_reason VARCHAR(255) NOT NULL DEFAULT '',
    status INT NOT NULL DEFAULT 1,
    CONSTRAINT uniq_love_likes_user_target UNIQUE (user_id, target_user_id)
);
CREATE INDEX IF NOT EXISTS idx_love_likes_region_id ON love_likes(region_id);
CREATE INDEX IF NOT EXISTS idx_love_likes_user_id ON love_likes(user_id);
CREATE INDEX IF NOT EXISTS idx_love_likes_love_id ON love_likes(love_id);
CREATE INDEX IF NOT EXISTS idx_love_likes_target_user_id ON love_likes(target_user_id);
CREATE INDEX IF NOT EXISTS idx_love_likes_target_love_id ON love_likes(target_love_id);
CREATE INDEX IF NOT EXISTS idx_love_likes_action ON love_likes(action);
CREATE INDEX IF NOT EXISTS idx_love_likes_super_like ON love_likes(super_like) WHERE super_like = TRUE;
CREATE INDEX IF NOT EXISTS idx_love_likes_is_matched ON love_likes(is_matched) WHERE is_matched = TRUE;
CREATE INDEX IF NOT EXISTS idx_love_likes_created_at ON love_likes(created_at);
CREATE INDEX IF NOT EXISTS idx_love_likes_deleted_at ON love_likes(deleted_at);
COMMENT ON TABLE love_likes IS '喜欢记录表（like/dislike/super_like + 是否匹配 + 撤销）';

-- ------------------------------------------------------------
-- 12. love_chat_sessions 匹配后聊天会话表
--     对标陌陌/Soul：匹配后开聊，未读/最后消息
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS love_chat_sessions (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    session_no VARCHAR(64) NOT NULL,
    match_id BIGINT NOT NULL,
    user_id_a BIGINT NOT NULL,
    user_id_b BIGINT NOT NULL,
    love_id_a BIGINT NOT NULL,
    love_id_b BIGINT NOT NULL,
    nickname_a VARCHAR(64) NOT NULL DEFAULT '',
    nickname_b VARCHAR(64) NOT NULL DEFAULT '',
    avatar_a VARCHAR(255) NOT NULL DEFAULT '',
    avatar_b VARCHAR(255) NOT NULL DEFAULT '',
    last_message_id BIGINT NOT NULL DEFAULT 0,
    last_message_content TEXT NOT NULL DEFAULT '',
    last_message_type VARCHAR(16) NOT NULL DEFAULT '',
    last_message_at TIMESTAMPTZ,
    last_sender_id BIGINT NOT NULL DEFAULT 0,
    unread_count_a INT NOT NULL DEFAULT 0,
    unread_count_b INT NOT NULL DEFAULT 0,
    muted_a BOOLEAN NOT NULL DEFAULT FALSE,
    muted_b BOOLEAN NOT NULL DEFAULT FALSE,
    pinned_a BOOLEAN NOT NULL DEFAULT FALSE,
    pinned_b BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_a BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_b BOOLEAN NOT NULL DEFAULT FALSE,
    status INT NOT NULL DEFAULT 1,
    dissolved_at TIMESTAMPTZ,
    dissolve_reason VARCHAR(255) NOT NULL DEFAULT '',
    dissolve_by BIGINT NOT NULL DEFAULT 0,
    message_count INT NOT NULL DEFAULT 0,
    gift_count INT NOT NULL DEFAULT 0,
    CONSTRAINT uniq_love_chat_sessions_no UNIQUE (session_no),
    CONSTRAINT uniq_love_chat_sessions_match UNIQUE (match_id)
);
CREATE INDEX IF NOT EXISTS idx_love_chat_sessions_region_id ON love_chat_sessions(region_id);
CREATE INDEX IF NOT EXISTS idx_love_chat_sessions_match_id ON love_chat_sessions(match_id);
CREATE INDEX IF NOT EXISTS idx_love_chat_sessions_user_id_a ON love_chat_sessions(user_id_a);
CREATE INDEX IF NOT EXISTS idx_love_chat_sessions_user_id_b ON love_chat_sessions(user_id_b);
CREATE INDEX IF NOT EXISTS idx_love_chat_sessions_last_message_at ON love_chat_sessions(last_message_at);
CREATE INDEX IF NOT EXISTS idx_love_chat_sessions_status ON love_chat_sessions(status);
CREATE INDEX IF NOT EXISTS idx_love_chat_sessions_deleted_at ON love_chat_sessions(deleted_at);
COMMENT ON TABLE love_chat_sessions IS '聊天会话表（匹配后会话 + 未读数 + 最后消息）';

-- ------------------------------------------------------------
-- 13. love_notifications 通知表
--     对标陌陌/Soul：喜欢/匹配/访客/礼物通知
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS love_notifications (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id BIGINT NOT NULL,
    love_id BIGINT NOT NULL,
    type VARCHAR(32) NOT NULL DEFAULT 'system',
    title VARCHAR(200) NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    from_user_id BIGINT NOT NULL DEFAULT 0,
    from_user_nickname VARCHAR(64) NOT NULL DEFAULT '',
    from_user_avatar VARCHAR(255) NOT NULL DEFAULT '',
    target_type VARCHAR(32) NOT NULL DEFAULT '',
    target_id BIGINT NOT NULL DEFAULT 0,
    action_url VARCHAR(255) NOT NULL DEFAULT '',
    extra JSONB,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    read_at TIMESTAMPTZ,
    is_pushed BOOLEAN NOT NULL DEFAULT FALSE,
    pushed_at TIMESTAMPTZ,
    push_status VARCHAR(16) NOT NULL DEFAULT '',
    push_error VARCHAR(500) NOT NULL DEFAULT '',
    status INT NOT NULL DEFAULT 1,
    expired_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_love_notifications_region_id ON love_notifications(region_id);
CREATE INDEX IF NOT EXISTS idx_love_notifications_user_id ON love_notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_love_notifications_love_id ON love_notifications(love_id);
CREATE INDEX IF NOT EXISTS idx_love_notifications_type ON love_notifications(type);
CREATE INDEX IF NOT EXISTS idx_love_notifications_is_read ON love_notifications(is_read) WHERE is_read = FALSE;
CREATE INDEX IF NOT EXISTS idx_love_notifications_created_at ON love_notifications(created_at);
CREATE INDEX IF NOT EXISTS idx_love_notifications_status ON love_notifications(status);
CREATE INDEX IF NOT EXISTS idx_love_notifications_deleted_at ON love_notifications(deleted_at);
COMMENT ON TABLE love_notifications IS '通知表（喜欢/匹配/访客/礼物/系统通知）';

-- ------------------------------------------------------------
-- 14. love_recommendations 推荐池表
--     对标陌陌/探探：每日推荐/附近推荐/同城推荐
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS love_recommendations (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id BIGINT NOT NULL,
    love_id BIGINT NOT NULL,
    target_user_id BIGINT NOT NULL,
    target_love_id BIGINT NOT NULL,
    target_nickname VARCHAR(64) NOT NULL DEFAULT '',
    target_avatar VARCHAR(255) NOT NULL DEFAULT '',
    target_gender SMALLINT NOT NULL DEFAULT 0,
    target_age INT NOT NULL DEFAULT 0,
    target_distance DECIMAL(10,2) NOT NULL DEFAULT 0,
    rec_type VARCHAR(32) NOT NULL DEFAULT 'daily',
    source VARCHAR(32) NOT NULL DEFAULT 'algorithm',
    score DECIMAL(5,2) NOT NULL DEFAULT 0,
    interest_match DECIMAL(5,2) NOT NULL DEFAULT 0,
    personality_match DECIMAL(5,2) NOT NULL DEFAULT 0,
    value_match DECIMAL(5,2) NOT NULL DEFAULT 0,
    location_match DECIMAL(5,2) NOT NULL DEFAULT 0,
    age_match DECIMAL(5,2) NOT NULL DEFAULT 0,
    reason VARCHAR(500) NOT NULL DEFAULT '',
    is_viewed BOOLEAN NOT NULL DEFAULT FALSE,
    is_liked BOOLEAN NOT NULL DEFAULT FALSE,
    is_disliked BOOLEAN NOT NULL DEFAULT FALSE,
    is_super_liked BOOLEAN NOT NULL DEFAULT FALSE,
    is_skipped BOOLEAN NOT NULL DEFAULT FALSE,
    is_dismissed BOOLEAN NOT NULL DEFAULT FALSE,
    viewed_at TIMESTAMPTZ,
    liked_at TIMESTAMPTZ,
    disliked_at TIMESTAMPTZ,
    super_liked_at TIMESTAMPTZ,
    skipped_at TIMESTAMPTZ,
    dismissed_at TIMESTAMPTZ,
    expired_at TIMESTAMPTZ,
    status INT NOT NULL DEFAULT 0,
    CONSTRAINT uniq_love_recs_user_target_type UNIQUE (user_id, target_user_id, rec_type)
);
CREATE INDEX IF NOT EXISTS idx_love_recommendations_region_id ON love_recommendations(region_id);
CREATE INDEX IF NOT EXISTS idx_love_recommendations_user_id ON love_recommendations(user_id);
CREATE INDEX IF NOT EXISTS idx_love_recommendations_love_id ON love_recommendations(love_id);
CREATE INDEX IF NOT EXISTS idx_love_recommendations_target_user_id ON love_recommendations(target_user_id);
CREATE INDEX IF NOT EXISTS idx_love_recommendations_target_love_id ON love_recommendations(target_love_id);
CREATE INDEX IF NOT EXISTS idx_love_recommendations_rec_type ON love_recommendations(rec_type);
CREATE INDEX IF NOT EXISTS idx_love_recommendations_source ON love_recommendations(source);
CREATE INDEX IF NOT EXISTS idx_love_recommendations_score ON love_recommendations(score);
CREATE INDEX IF NOT EXISTS idx_love_recommendations_status ON love_recommendations(status);
CREATE INDEX IF NOT EXISTS idx_love_recommendations_expired_at ON love_recommendations(expired_at);
CREATE INDEX IF NOT EXISTS idx_love_recommendations_deleted_at ON love_recommendations(deleted_at);
COMMENT ON TABLE love_recommendations IS '推荐池表（每日推荐/附近推荐/同城推荐）';

-- ------------------------------------------------------------
-- 15. love_privacy_settings 隐私设置表
--     对标陌陌/探探：在线/位置/年龄/距离/通讯录
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS love_privacy_settings (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id BIGINT NOT NULL,
    love_id BIGINT NOT NULL,
    hide_online BOOLEAN NOT NULL DEFAULT FALSE,
    hide_location BOOLEAN NOT NULL DEFAULT FALSE,
    hide_age BOOLEAN NOT NULL DEFAULT FALSE,
    hide_distance BOOLEAN NOT NULL DEFAULT FALSE,
    hide_constellation BOOLEAN NOT NULL DEFAULT FALSE,
    hide_hometown BOOLEAN NOT NULL DEFAULT FALSE,
    hide_occupation BOOLEAN NOT NULL DEFAULT FALSE,
    hide_income BOOLEAN NOT NULL DEFAULT FALSE,
    hide_last_active BOOLEAN NOT NULL DEFAULT TRUE,
    hide_visitors BOOLEAN NOT NULL DEFAULT FALSE,
    only_verified_can_see BOOLEAN NOT NULL DEFAULT FALSE,
    only_verified_can_match BOOLEAN NOT NULL DEFAULT FALSE,
    only_member_can_chat BOOLEAN NOT NULL DEFAULT FALSE,
    block_strangers BOOLEAN NOT NULL DEFAULT FALSE,
    block_same_city BOOLEAN NOT NULL DEFAULT FALSE,
    allow_phone_lookup BOOLEAN NOT NULL DEFAULT FALSE,
    allow_contact_import BOOLEAN NOT NULL DEFAULT FALSE,
    allow_recommendation BOOLEAN NOT NULL DEFAULT TRUE,
    allow_story BOOLEAN NOT NULL DEFAULT TRUE,
    allow_impression BOOLEAN NOT NULL DEFAULT TRUE,
    distance_visibility INT NOT NULL DEFAULT 0,
    age_visibility INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 1,
    CONSTRAINT uniq_love_privacy_user UNIQUE (user_id),
    CONSTRAINT uniq_love_privacy_love UNIQUE (love_id)
);
CREATE INDEX IF NOT EXISTS idx_love_privacy_settings_region_id ON love_privacy_settings(region_id);
CREATE INDEX IF NOT EXISTS idx_love_privacy_settings_user_id ON love_privacy_settings(user_id);
CREATE INDEX IF NOT EXISTS idx_love_privacy_settings_love_id ON love_privacy_settings(love_id);
CREATE INDEX IF NOT EXISTS idx_love_privacy_settings_status ON love_privacy_settings(status);
CREATE INDEX IF NOT EXISTS idx_love_privacy_settings_deleted_at ON love_privacy_settings(deleted_at);
COMMENT ON TABLE love_privacy_settings IS '隐私设置表（隐藏在线/位置/年龄/距离/通讯录匹配）';

-- ------------------------------------------------------------
-- 16. love_reports 举报表
--     对标陌陌/Soul：举报用户/动态/消息
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS love_reports (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    report_no VARCHAR(64) NOT NULL,
    reporter_user_id BIGINT NOT NULL,
    reporter_love_id BIGINT NOT NULL,
    reporter_nickname VARCHAR(64) NOT NULL DEFAULT '',
    reporter_avatar VARCHAR(255) NOT NULL DEFAULT '',
    target_type VARCHAR(32) NOT NULL DEFAULT 'user',
    target_user_id BIGINT NOT NULL,
    target_love_id BIGINT NOT NULL,
    target_nickname VARCHAR(64) NOT NULL DEFAULT '',
    target_avatar VARCHAR(255) NOT NULL DEFAULT '',
    target_id BIGINT NOT NULL DEFAULT 0,
    reason_type VARCHAR(32) NOT NULL DEFAULT '',
    reason_detail TEXT NOT NULL DEFAULT '',
    evidence_images JSONB,
    evidence_videos JSONB,
    evidence_text TEXT NOT NULL DEFAULT '',
    chat_snapshot JSONB,
    ip VARCHAR(64) NOT NULL DEFAULT '',
    status INT NOT NULL DEFAULT 0,
    handled_by BIGINT NOT NULL DEFAULT 0,
    handled_at TIMESTAMPTZ,
    handle_result VARCHAR(32) NOT NULL DEFAULT '',
    handle_remark TEXT NOT NULL DEFAULT '',
    penalty_type VARCHAR(32) NOT NULL DEFAULT '',
    penalty_duration INT NOT NULL DEFAULT 0,
    penalty_expired_at TIMESTAMPTZ,
    appeal_status INT NOT NULL DEFAULT 0,
    appeal_reason TEXT NOT NULL DEFAULT '',
    appealed_at TIMESTAMPTZ,
    appeal_handled_by BIGINT NOT NULL DEFAULT 0,
    appeal_handled_at TIMESTAMPTZ,
    appeal_result VARCHAR(32) NOT NULL DEFAULT '',
    appeal_remark TEXT NOT NULL DEFAULT '',
    risk_score INT NOT NULL DEFAULT 0,
    CONSTRAINT uniq_love_reports_no UNIQUE (report_no)
);
CREATE INDEX IF NOT EXISTS idx_love_reports_region_id ON love_reports(region_id);
CREATE INDEX IF NOT EXISTS idx_love_reports_reporter_user_id ON love_reports(reporter_user_id);
CREATE INDEX IF NOT EXISTS idx_love_reports_target_user_id ON love_reports(target_user_id);
CREATE INDEX IF NOT EXISTS idx_love_reports_target_love_id ON love_reports(target_love_id);
CREATE INDEX IF NOT EXISTS idx_love_reports_target_type ON love_reports(target_type);
CREATE INDEX IF NOT EXISTS idx_love_reports_reason_type ON love_reports(reason_type);
CREATE INDEX IF NOT EXISTS idx_love_reports_status ON love_reports(status);
CREATE INDEX IF NOT EXISTS idx_love_reports_appeal_status ON love_reports(appeal_status);
CREATE INDEX IF NOT EXISTS idx_love_reports_handled_at ON love_reports(handled_at);
CREATE INDEX IF NOT EXISTS idx_love_reports_created_at ON love_reports(created_at);
CREATE INDEX IF NOT EXISTS idx_love_reports_deleted_at ON love_reports(deleted_at);
COMMENT ON TABLE love_reports IS '举报表（用户/动态/消息举报 + 处理 + 申诉）';

-- ------------------------------------------------------------
-- 17. love_audit_rules 审核规则表
--     对标陌陌/探探：敏感词/违规内容/频率限制
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS love_audit_rules (
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
    scope VARCHAR(32) NOT NULL DEFAULT 'all',
    target_type VARCHAR(32) NOT NULL DEFAULT 'all',
    description VARCHAR(500) NOT NULL DEFAULT '',
    sort INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 1,
    CONSTRAINT uniq_love_audit_rules_key UNIQUE (rule_key)
);
CREATE INDEX IF NOT EXISTS idx_love_audit_rules_rule_type ON love_audit_rules(rule_type);
CREATE INDEX IF NOT EXISTS idx_love_audit_rules_rule_key ON love_audit_rules(rule_key);
CREATE INDEX IF NOT EXISTS idx_love_audit_rules_action ON love_audit_rules(action);
CREATE INDEX IF NOT EXISTS idx_love_audit_rules_severity ON love_audit_rules(severity);
CREATE INDEX IF NOT EXISTS idx_love_audit_rules_status ON love_audit_rules(status);
CREATE INDEX IF NOT EXISTS idx_love_audit_rules_sort ON love_audit_rules(sort);
CREATE INDEX IF NOT EXISTS idx_love_audit_rules_deleted_at ON love_audit_rules(deleted_at);
COMMENT ON TABLE love_audit_rules IS '审核规则表（敏感词/违禁内容/频率限制/虚假资料）';

-- ============================================================
-- 第三部分：为 17 张子表挂载 updated_at 触发器
--   参考 001_p0_baseline.sql 中的 update_updated_at_column 函数
--   幂等：先 DROP IF EXISTS 再 CREATE
-- ============================================================
DO $$
DECLARE t TEXT;
BEGIN
    FOR t IN SELECT tablename FROM pg_tables WHERE schemaname='public' AND tablename IN (
        'love_profiles','love_matches','love_stories','love_member_levels','love_verifications',
        'love_impressions','love_visits','love_blocks','love_gifts','love_memberships',
        'love_likes','love_chat_sessions','love_notifications','love_recommendations',
        'love_privacy_settings','love_reports','love_audit_rules'
    )
    LOOP
        EXECUTE format('DROP TRIGGER IF EXISTS trg_%s_updated ON %s', t, t);
        EXECUTE format('CREATE TRIGGER trg_%s_updated BEFORE UPDATE ON %s FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()', t, t);
    END LOOP;
END $$;

-- ============================================================
-- 第四部分：初始化会员等级默认数据（幂等：仅当表为空时插入）
-- ============================================================
INSERT INTO love_member_levels (level_code, level_name, level, description, monthly_price, quarterly_price, yearly_price, daily_super_likes, daily_likes, daily_visits, daily_recommendations, can_see_visitors, can_see_likes, can_hide_online, can_hide_location, can_filter_verified, can_advanced_filter, can_super_like, can_undo_swipe, can_boost_profile, can_see_match_score, sort, status)
SELECT 'basic', '基础会员', 1, '基础会员：每天5次心动信号 + 查看访客', 18.00, 48.00, 168.00, 1, 30, 30, 10, TRUE, FALSE, FALSE, FALSE, FALSE, FALSE, TRUE, FALSE, FALSE, FALSE, 1, 1
WHERE NOT EXISTS (SELECT 1 FROM love_member_levels WHERE level_code = 'basic');

INSERT INTO love_member_levels (level_code, level_name, level, description, monthly_price, quarterly_price, yearly_price, daily_super_likes, daily_likes, daily_visits, daily_recommendations, can_see_visitors, can_see_likes, can_hide_online, can_hide_location, can_filter_verified, can_advanced_filter, can_super_like, can_undo_swipe, can_boost_profile, can_see_match_score, sort, status)
SELECT 'advanced', '高级会员', 2, '高级会员：5次心动信号 + 看谁喜欢我 + 隐藏在线', 38.00, 98.00, 348.00, 3, 100, 100, 30, TRUE, TRUE, TRUE, FALSE, TRUE, TRUE, TRUE, TRUE, FALSE, TRUE, 2, 1
WHERE NOT EXISTS (SELECT 1 FROM love_member_levels WHERE level_code = 'advanced');

INSERT INTO love_member_levels (level_code, level_name, level, description, monthly_price, quarterly_price, yearly_price, daily_super_likes, daily_likes, daily_visits, daily_recommendations, can_see_visitors, can_see_likes, can_hide_online, can_hide_location, can_filter_verified, can_advanced_filter, can_super_like, can_undo_swipe, can_boost_profile, can_see_match_score, sort, status)
SELECT 'vip', 'VIP 会员', 3, 'VIP 会员：无限心动信号 + 全功能 + 流量加成', 68.00, 168.00, 598.00, 5, 999, 999, 100, TRUE, TRUE, TRUE, TRUE, TRUE, TRUE, TRUE, TRUE, TRUE, TRUE, 3, 1
WHERE NOT EXISTS (SELECT 1 FROM love_member_levels WHERE level_code = 'vip');

INSERT INTO love_member_levels (level_code, level_name, level, description, monthly_price, quarterly_price, yearly_price, daily_super_likes, daily_likes, daily_visits, daily_recommendations, can_see_visitors, can_see_likes, can_hide_online, can_hide_location, can_filter_verified, can_advanced_filter, can_super_like, can_undo_swipe, can_boost_profile, can_see_match_score, sort, status)
SELECT 'premium', 'Premium 会员', 4, 'Premium 会员：尊贵身份 + 专属客服 + 红娘服务', 128.00, 328.00, 1188.00, 10, 9999, 9999, 999, TRUE, TRUE, TRUE, TRUE, TRUE, TRUE, TRUE, TRUE, TRUE, TRUE, 4, 1
WHERE NOT EXISTS (SELECT 1 FROM love_member_levels WHERE level_code = 'premium');

-- ============================================================
-- 第五部分：初始化默认虚拟礼物数据（幂等：仅当表为空时插入）
-- ============================================================
INSERT INTO love_gifts (gift_code, gift_name, category, description, icon, price, charm_value, sort, status)
SELECT 'rose', '玫瑰', 'common', '一朵红玫瑰，表达心动', '🌹', 1.00, 1, 1, 1
WHERE NOT EXISTS (SELECT 1 FROM love_gifts WHERE gift_code = 'rose');

INSERT INTO love_gifts (gift_code, gift_name, category, description, icon, price, charm_value, sort, status)
SELECT 'bouquet', '花束', 'common', '一束鲜花，浪漫表白', '💐', 18.00, 18, 2, 1
WHERE NOT EXISTS (SELECT 1 FROM love_gifts WHERE gift_code = 'bouquet');

INSERT INTO love_gifts (gift_code, gift_name, category, description, icon, price, charm_value, sort, status)
SELECT 'chocolate', '巧克力', 'common', '甜蜜巧克力，温暖关怀', '🍫', 9.90, 9, 3, 1
WHERE NOT EXISTS (SELECT 1 FROM love_gifts WHERE gift_code = 'chocolate');

INSERT INTO love_gifts (gift_code, gift_name, category, description, icon, price, charm_value, sort, status)
SELECT 'ring', '戒指', 'luxury', '珍贵钻戒，永恒承诺', '💍', 520.00, 520, 10, 1
WHERE NOT EXISTS (SELECT 1 FROM love_gifts WHERE gift_code = 'ring');

INSERT INTO love_gifts (gift_code, gift_name, category, description, icon, price, charm_value, sort, status)
SELECT 'car', '豪车', 'luxury', '限量跑车，霸气表白', '🚗', 1314.00, 1314, 20, 1
WHERE NOT EXISTS (SELECT 1 FROM love_gifts WHERE gift_code = 'car');
