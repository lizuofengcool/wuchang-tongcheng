-- ============================================================
-- love 相亲交友模块完整功能回滚脚本（与 017_love_full.sql 配对）
-- 按反向顺序 DROP 17 张子表 + 触发器 + 主表新增字段
-- 幂等：使用 IF EXISTS，可重复执行
-- 警告：执行后 love 模块全部扩展数据将丢失（主表基础字段保留）
-- ============================================================

-- ------------------------------------------------------------
-- 1. 先移除 17 张子表的 updated_at 触发器
-- ------------------------------------------------------------
DROP TRIGGER IF EXISTS trg_love_audit_rules_updated ON love_audit_rules;
DROP TRIGGER IF EXISTS trg_love_reports_updated ON love_reports;
DROP TRIGGER IF EXISTS trg_love_privacy_settings_updated ON love_privacy_settings;
DROP TRIGGER IF EXISTS trg_love_recommendations_updated ON love_recommendations;
DROP TRIGGER IF EXISTS trg_love_notifications_updated ON love_notifications;
DROP TRIGGER IF EXISTS trg_love_chat_sessions_updated ON love_chat_sessions;
DROP TRIGGER IF EXISTS trg_love_likes_updated ON love_likes;
DROP TRIGGER IF EXISTS trg_love_memberships_updated ON love_memberships;
DROP TRIGGER IF EXISTS trg_love_gifts_updated ON love_gifts;
DROP TRIGGER IF EXISTS trg_love_blocks_updated ON love_blocks;
DROP TRIGGER IF EXISTS trg_love_visits_updated ON love_visits;
DROP TRIGGER IF EXISTS trg_love_impressions_updated ON love_impressions;
DROP TRIGGER IF EXISTS trg_love_verifications_updated ON love_verifications;
DROP TRIGGER IF EXISTS trg_love_member_levels_updated ON love_member_levels;
DROP TRIGGER IF EXISTS trg_love_stories_updated ON love_stories;
DROP TRIGGER IF EXISTS trg_love_matches_updated ON love_matches;
DROP TRIGGER IF EXISTS trg_love_profiles_updated ON love_profiles;

-- ------------------------------------------------------------
-- 2. 按反向顺序 DROP 17 张子表
--    依赖顺序：love_chat_sessions → love_matches → love_likes；
--             love_recommendations → love_likes → love_matches
-- ------------------------------------------------------------
DROP TABLE IF EXISTS love_audit_rules;
DROP TABLE IF EXISTS love_reports;
DROP TABLE IF EXISTS love_privacy_settings;
DROP TABLE IF EXISTS love_recommendations;
DROP TABLE IF EXISTS love_notifications;
DROP TABLE IF EXISTS love_chat_sessions;
DROP TABLE IF EXISTS love_likes;
DROP TABLE IF EXISTS love_memberships;
DROP TABLE IF EXISTS love_gifts;
DROP TABLE IF EXISTS love_blocks;
DROP TABLE IF EXISTS love_visits;
DROP TABLE IF EXISTS love_impressions;
DROP TABLE IF EXISTS love_verifications;
DROP TABLE IF EXISTS love_member_levels;
DROP TABLE IF EXISTS love_stories;
DROP TABLE IF EXISTS love_matches;
DROP TABLE IF EXISTS love_profiles;

-- ------------------------------------------------------------
-- 3. 移除 loves 主表新增字段（按反向顺序）
--    包装在 DO 块中，表不存在时跳过（幂等）
-- ------------------------------------------------------------
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'loves') THEN
        -- JSONB 字段
        ALTER TABLE loves DROP COLUMN IF EXISTS match_preferences;
        ALTER TABLE loves DROP COLUMN IF EXISTS photo_urls;
        ALTER TABLE loves DROP COLUMN IF EXISTS values;
        ALTER TABLE loves DROP COLUMN IF EXISTS personality;
        ALTER TABLE loves DROP COLUMN IF EXISTS interests;
        ALTER TABLE loves DROP COLUMN IF EXISTS tags;

        -- 风控
        ALTER TABLE loves DROP COLUMN IF EXISTS content_hash;
        ALTER TABLE loves DROP COLUMN IF EXISTS risk_score;

        -- 运营字段
        ALTER TABLE loves DROP COLUMN IF EXISTS picked;
        ALTER TABLE loves DROP COLUMN IF EXISTS featured;

        -- 统计
        ALTER TABLE loves DROP COLUMN IF EXISTS popularity_score;
        ALTER TABLE loves DROP COLUMN IF EXISTS impression_count;
        ALTER TABLE loves DROP COLUMN IF EXISTS gift_count;
        ALTER TABLE loves DROP COLUMN IF EXISTS story_count;
        ALTER TABLE loves DROP COLUMN IF EXISTS visitor_count;
        ALTER TABLE loves DROP COLUMN IF EXISTS match_count;
        ALTER TABLE loves DROP COLUMN IF EXISTS liked_count;
        ALTER TABLE loves DROP COLUMN IF EXISTS like_count;
        ALTER TABLE loves DROP COLUMN IF EXISTS view_count;

        -- 联系
        ALTER TABLE loves DROP COLUMN IF EXISTS contact_price;
        ALTER TABLE loves DROP COLUMN IF EXISTS only_verified_match;
        ALTER TABLE loves DROP COLUMN IF EXISTS hide_distance;
        ALTER TABLE loves DROP COLUMN IF EXISTS hide_age;
        ALTER TABLE loves DROP COLUMN IF EXISTS hide_location;
        ALTER TABLE loves DROP COLUMN IF EXISTS hide_online;

        -- 地理位置
        ALTER TABLE loves DROP COLUMN IF EXISTS location_updated_at;
        ALTER TABLE loves DROP COLUMN IF EXISTS latitude;
        ALTER TABLE loves DROP COLUMN IF EXISTS longitude;

        -- 活跃信息
        ALTER TABLE loves DROP COLUMN IF EXISTS last_active_ip;
        ALTER TABLE loves DROP COLUMN IF EXISTS last_active_at;

        -- 心动信号
        ALTER TABLE loves DROP COLUMN IF EXISTS super_likes_reset_at;
        ALTER TABLE loves DROP COLUMN IF EXISTS super_likes_today;

        -- 金币
        ALTER TABLE loves DROP COLUMN IF EXISTS credits;

        -- 会员
        ALTER TABLE loves DROP COLUMN IF EXISTS member_expired_at;
        ALTER TABLE loves DROP COLUMN IF EXISTS member_level;

        -- 审核
        ALTER TABLE loves DROP COLUMN IF EXISTS audit_reason;
        ALTER TABLE loves DROP COLUMN IF EXISTS audit_status;
        ALTER TABLE loves DROP COLUMN IF EXISTS status;

        -- 认证
        ALTER TABLE loves DROP COLUMN IF EXISTS real_name_verified;
        ALTER TABLE loves DROP COLUMN IF EXISTS education_verified;
        ALTER TABLE loves DROP COLUMN IF EXISTS video_verified;
        ALTER TABLE loves DROP COLUMN IF EXISTS photo_verified;

        -- 资料
        ALTER TABLE loves DROP COLUMN IF EXISTS cover_image;
        ALTER TABLE loves DROP COLUMN IF EXISTS voice_intro_url;
        ALTER TABLE loves DROP COLUMN IF EXISTS bio;
        ALTER TABLE loves DROP COLUMN IF EXISTS want_kids;
        ALTER TABLE loves DROP COLUMN IF EXISTS smoking;
        ALTER TABLE loves DROP COLUMN IF EXISTS drinking;
        ALTER TABLE loves DROP COLUMN IF EXISTS car;
        ALTER TABLE loves DROP COLUMN IF EXISTS house;
        ALTER TABLE loves DROP COLUMN IF EXISTS marriage;
        ALTER TABLE loves DROP COLUMN IF EXISTS income;
        ALTER TABLE loves DROP COLUMN IF EXISTS occupation;
        ALTER TABLE loves DROP COLUMN IF EXISTS education;
        ALTER TABLE loves DROP COLUMN IF EXISTS residence;
        ALTER TABLE loves DROP COLUMN IF EXISTS hometown;
        ALTER TABLE loves DROP COLUMN IF EXISTS zodiac;
        ALTER TABLE loves DROP COLUMN IF EXISTS constellation;
        ALTER TABLE loves DROP COLUMN IF EXISTS weight;
        ALTER TABLE loves DROP COLUMN IF EXISTS height;
        ALTER TABLE loves DROP COLUMN IF EXISTS birthday;
        ALTER TABLE loves DROP COLUMN IF EXISTS age;
        ALTER TABLE loves DROP COLUMN IF EXISTS gender;
        ALTER TABLE loves DROP COLUMN IF EXISTS avatar;
        ALTER TABLE loves DROP COLUMN IF EXISTS nickname;
        ALTER TABLE loves DROP COLUMN IF EXISTS user_id;
    END IF;
END $$;
