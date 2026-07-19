-- ============================================================
-- 013_im_full.sql 即时通讯中台扩展表
-- 对标微信/QQ/钉钉/飞书/Telegram
-- 创建 im_session_users / im_user_settings / im_groups / im_group_members / im_message_reads
-- 现有的 im_sessions / im_messages / im_system_notifications / im_privacy_numbers
-- 已在 005_p1_middlewares.sql 创建
-- 全部幂等：CREATE TABLE IF NOT EXISTS
-- ============================================================

-- ============================================================
-- 1. im_message_reads 已读记录（每条消息每用户已读）
-- ============================================================
CREATE TABLE IF NOT EXISTS im_message_reads (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    message_id BIGINT NOT NULL,
    session_id VARCHAR(64) NOT NULL DEFAULT '',
    user_id BIGINT NOT NULL,
    read_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_im_msg_reads UNIQUE (message_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_im_msg_reads_session_user ON im_message_reads(session_id, user_id);
CREATE INDEX IF NOT EXISTS idx_im_msg_reads_user_id ON im_message_reads(user_id);
CREATE INDEX IF NOT EXISTS idx_im_msg_reads_region_id ON im_message_reads(region_id);
COMMENT ON TABLE im_message_reads IS 'IM 消息已读记录';

-- ============================================================
-- 2. im_session_users 会话用户关联表
-- ============================================================
CREATE TABLE IF NOT EXISTS im_session_users (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    session_id VARCHAR(64) NOT NULL,
    user_id BIGINT NOT NULL,
    role VARCHAR(16) NOT NULL DEFAULT 'member',                -- owner/admin/member
    nickname VARCHAR(64) NOT NULL DEFAULT '',                  -- 群昵称
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_read_at TIMESTAMPTZ,                                   -- 最后已读时间
    mute_until TIMESTAMPTZ,                                     -- 禁言截止
    is_pinned SMALLINT NOT NULL DEFAULT 0,                      -- 是否置顶
    is_muted SMALLINT NOT NULL DEFAULT 0,                       -- 是否免打扰
    status SMALLINT NOT NULL DEFAULT 1,                        -- 1正常 0退出
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_im_session_users UNIQUE (session_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_im_session_users_user_id ON im_session_users(user_id);
CREATE INDEX IF NOT EXISTS idx_im_session_users_status ON im_session_users(status);
CREATE INDEX IF NOT EXISTS idx_im_session_users_region_id ON im_session_users(region_id);
COMMENT ON TABLE im_session_users IS 'IM 会话用户关联表';

-- ============================================================
-- 3. im_user_settings 用户IM设置
-- ============================================================
CREATE TABLE IF NOT EXISTS im_user_settings (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    user_id BIGINT NOT NULL UNIQUE,
    online_status VARCHAR(16) NOT NULL DEFAULT 'online',         -- online/away/busy/offline
    auto_reply TEXT NOT NULL DEFAULT '',                          -- 自动回复内容
    auto_reply_enabled SMALLINT NOT NULL DEFAULT 0,
    do_not_disturb SMALLINT NOT NULL DEFAULT 0,                   -- 全局免打扰
    notification_sound SMALLINT NOT NULL DEFAULT 1,
    notification_vibrate SMALLINT NOT NULL DEFAULT 1,
    save_to_album SMALLINT NOT NULL DEFAULT 0,                    -- 图片自动保存
    last_active_at TIMESTAMPTZ,
    extra JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_im_user_settings_region_id ON im_user_settings(region_id);
COMMENT ON TABLE im_user_settings IS '用户IM设置';

-- ============================================================
-- 4. im_groups 群组
-- ============================================================
CREATE TABLE IF NOT EXISTS im_groups (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    group_id VARCHAR(64) NOT NULL UNIQUE,                       -- 群ID（业务层）
    group_name VARCHAR(128) NOT NULL DEFAULT '',
    avatar VARCHAR(256) NOT NULL DEFAULT '',
    announcement TEXT NOT NULL DEFAULT '',                       -- 群公告
    owner_id BIGINT NOT NULL,                                    -- 群主
    member_count INTEGER NOT NULL DEFAULT 0,
    max_members INTEGER NOT NULL DEFAULT 500,
    join_type SMALLINT NOT NULL DEFAULT 0,                      -- 0任意加入 1需审核 2仅邀请
    mute_all SMALLINT NOT NULL DEFAULT 0,                        -- 全员禁言
    status SMALLINT NOT NULL DEFAULT 1,                         -- 1正常 0解散
    extra JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_im_groups_owner_id ON im_groups(owner_id);
CREATE INDEX IF NOT EXISTS idx_im_groups_status ON im_groups(status);
CREATE INDEX IF NOT EXISTS idx_im_groups_region_id ON im_groups(region_id);
COMMENT ON TABLE im_groups IS 'IM 群组';

-- ============================================================
-- 5. im_group_members 群成员
-- ============================================================
CREATE TABLE IF NOT EXISTS im_group_members (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    group_id VARCHAR(64) NOT NULL,
    user_id BIGINT NOT NULL,
    role VARCHAR(16) NOT NULL DEFAULT 'member',                  -- owner/admin/member
    nickname VARCHAR(64) NOT NULL DEFAULT '',
    inviter_id BIGINT NOT NULL DEFAULT 0,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    mute_until TIMESTAMPTZ,
    status SMALLINT NOT NULL DEFAULT 1,                         -- 1在群 0退出
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_im_group_members UNIQUE (group_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_im_group_members_user_id ON im_group_members(user_id);
CREATE INDEX IF NOT EXISTS idx_im_group_members_status ON im_group_members(status);
CREATE INDEX IF NOT EXISTS idx_im_group_members_region_id ON im_group_members(region_id);
COMMENT ON TABLE im_group_members IS 'IM 群成员';
