-- ============================================================
-- 016_ai_full.sql AI 审核/智能推荐中台扩展表
-- 对标字节智能推荐/百度AI/阿里云AI
-- 创建 ai_audit_results / ai_recommendations / ai_model_configs
-- / ai_chat_sessions / ai_chat_messages / ai_training_data
-- 现有的 ai_tasks / ai_models / ai_prompts / ai_generations 已在 005 创建
-- 全部幂等：CREATE TABLE IF NOT EXISTS
-- ============================================================

-- ============================================================
-- 1. ai_audit_results 审核结果
-- ============================================================
CREATE TABLE IF NOT EXISTS ai_audit_results (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    task_id VARCHAR(64) NOT NULL DEFAULT '',                       -- 关联 ai_tasks.task_id
    biz_module VARCHAR(32) NOT NULL DEFAULT '',
    biz_id VARCHAR(128) NOT NULL DEFAULT '',
    user_id BIGINT NOT NULL DEFAULT 0,
    content_type VARCHAR(32) NOT NULL DEFAULT 'text',             -- text/image/video
    content_hash VARCHAR(64) NOT NULL DEFAULT '',                 -- 内容哈希
    algorithm VARCHAR(32) NOT NULL DEFAULT 'local',               -- 算法 local/dfa/aliyun/tencent
    risk_score DECIMAL(5,2) NOT NULL DEFAULT 0.00,                -- 风险分数 0-100
    risk_level VARCHAR(16) NOT NULL DEFAULT 'safe',              -- safe/warning/danger
    labels JSONB NOT NULL DEFAULT '[]'::jsonb,                     -- 标签列表
    hit_rules JSONB NOT NULL DEFAULT '[]'::jsonb,                  -- 命中规则
    suggestion VARCHAR(256) NOT NULL DEFAULT '',                  -- 处理建议 pass/review/block
    passed SMALLINT NOT NULL DEFAULT 1,                            -- 1通过 0不通过
    cost_ms INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_ai_audit_results_task_id ON ai_audit_results(task_id);
CREATE INDEX IF NOT EXISTS idx_ai_audit_results_biz ON ai_audit_results(biz_module, biz_id);
CREATE INDEX IF NOT EXISTS idx_ai_audit_results_user_id ON ai_audit_results(user_id);
CREATE INDEX IF NOT EXISTS idx_ai_audit_results_risk_level ON ai_audit_results(risk_level);
CREATE INDEX IF NOT EXISTS idx_ai_audit_results_passed ON ai_audit_results(passed);
CREATE INDEX IF NOT EXISTS idx_ai_audit_results_region_id ON ai_audit_results(region_id);
COMMENT ON TABLE ai_audit_results IS 'AI 审核结果表';

-- ============================================================
-- 2. ai_recommendations 推荐记录
-- ============================================================
CREATE TABLE IF NOT EXISTS ai_recommendations (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    user_id BIGINT NOT NULL,
    biz_module VARCHAR(32) NOT NULL DEFAULT '',
    content_type VARCHAR(32) NOT NULL DEFAULT 'item',              -- item/news/video/ad
    content_id VARCHAR(128) NOT NULL DEFAULT '',
    rec_type VARCHAR(32) NOT NULL DEFAULT 'hot',                  -- hot/new/personalized/similar
    score DECIMAL(8,2) NOT NULL DEFAULT 0.00,                     -- 推荐分数
    reason VARCHAR(256) NOT NULL DEFAULT '',                       -- 推荐理由
    is_clicked SMALLINT NOT NULL DEFAULT 0,                       -- 是否被点击
    clicked_at TIMESTAMPTZ,
    is_liked SMALLINT NOT NULL DEFAULT 0,
    is_disliked SMALLINT NOT NULL DEFAULT 0,
    dwell_ms INTEGER NOT NULL DEFAULT 0,                          -- 停留时长
    extra JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_ai_recommendations_user_id ON ai_recommendations(user_id);
CREATE INDEX IF NOT EXISTS idx_ai_recommendations_biz ON ai_recommendations(biz_module, content_id);
CREATE INDEX IF NOT EXISTS idx_ai_recommendations_rec_type ON ai_recommendations(rec_type);
CREATE INDEX IF NOT EXISTS idx_ai_recommendations_clicked ON ai_recommendations(is_clicked);
CREATE INDEX IF NOT EXISTS idx_ai_recommendations_created_at ON ai_recommendations(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_ai_recommendations_region_id ON ai_recommendations(region_id);
COMMENT ON TABLE ai_recommendations IS 'AI 智能推荐记录表';

-- ============================================================
-- 3. ai_model_configs 模型参数配置
-- ============================================================
CREATE TABLE IF NOT EXISTS ai_model_configs (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    model_id BIGINT NOT NULL DEFAULT 0,                           -- 关联 ai_models.id
    config_key VARCHAR(64) NOT NULL,                              -- 参数名 temperature/top_p/max_tokens
    config_value VARCHAR(256) NOT NULL DEFAULT '',                -- 参数值
    config_type VARCHAR(16) NOT NULL DEFAULT 'string',             -- string/number/boolean/json
    description VARCHAR(256) NOT NULL DEFAULT '',
    status SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_ai_model_configs UNIQUE (model_id, config_key)
);
CREATE INDEX IF NOT EXISTS idx_ai_model_configs_model_id ON ai_model_configs(model_id);
CREATE INDEX IF NOT EXISTS idx_ai_model_configs_status ON ai_model_configs(status);
CREATE INDEX IF NOT EXISTS idx_ai_model_configs_region_id ON ai_model_configs(region_id);
COMMENT ON TABLE ai_model_configs IS 'AI 模型参数配置';

-- ============================================================
-- 4. ai_chat_sessions AI 对话会话
-- ============================================================
CREATE TABLE IF NOT EXISTS ai_chat_sessions (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    session_id VARCHAR(64) NOT NULL UNIQUE,
    user_id BIGINT NOT NULL,
    title VARCHAR(128) NOT NULL DEFAULT '',
    model_name VARCHAR(64) NOT NULL DEFAULT '',
    system_prompt TEXT NOT NULL DEFAULT '',                        -- 系统提示
    context_length INTEGER NOT NULL DEFAULT 10,                    -- 上下文长度
    total_messages INTEGER NOT NULL DEFAULT 0,
    total_tokens INTEGER NOT NULL DEFAULT 0,
    status SMALLINT NOT NULL DEFAULT 1,                            -- 1活跃 0归档
    extra JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_ai_chat_sessions_user_id ON ai_chat_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_ai_chat_sessions_status ON ai_chat_sessions(status);
CREATE INDEX IF NOT EXISTS idx_ai_chat_sessions_created_at ON ai_chat_sessions(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_ai_chat_sessions_region_id ON ai_chat_sessions(region_id);
COMMENT ON TABLE ai_chat_sessions IS 'AI 对话会话表';

-- ============================================================
-- 5. ai_chat_messages AI 对话消息
-- ============================================================
CREATE TABLE IF NOT EXISTS ai_chat_messages (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    session_id VARCHAR(64) NOT NULL,
    user_id BIGINT NOT NULL,
    role VARCHAR(16) NOT NULL DEFAULT 'user',                     -- user/assistant/system
    content TEXT NOT NULL DEFAULT '',
    tokens INTEGER NOT NULL DEFAULT 0,
    model_name VARCHAR(64) NOT NULL DEFAULT '',
    cost_ms INTEGER NOT NULL DEFAULT 0,
    images JSONB NOT NULL DEFAULT '[]'::jsonb,                     -- 附带图片
    feedback SMALLINT NOT NULL DEFAULT 0,                         -- -1负面 0无 1正面
    feedback_text TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_ai_chat_messages_session_id ON ai_chat_messages(session_id);
CREATE INDEX IF NOT EXISTS idx_ai_chat_messages_user_id ON ai_chat_messages(user_id);
CREATE INDEX IF NOT EXISTS idx_ai_chat_messages_role ON ai_chat_messages(role);
CREATE INDEX IF NOT EXISTS idx_ai_chat_messages_created_at ON ai_chat_messages(created_at);
CREATE INDEX IF NOT EXISTS idx_ai_chat_messages_region_id ON ai_chat_messages(region_id);
COMMENT ON TABLE ai_chat_messages IS 'AI 对话消息表';

-- ============================================================
-- 6. ai_training_data 训练数据
-- ============================================================
CREATE TABLE IF NOT EXISTS ai_training_data (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    data_type VARCHAR(32) NOT NULL DEFAULT 'text',                 -- text/image/conversation
    biz_module VARCHAR(32) NOT NULL DEFAULT '',
    biz_id VARCHAR(128) NOT NULL DEFAULT '',
    user_id BIGINT NOT NULL DEFAULT 0,
    input TEXT NOT NULL DEFAULT '',
    output TEXT NOT NULL DEFAULT '',
    label VARCHAR(64) NOT NULL DEFAULT '',                         -- 标注标签
    quality_score DECIMAL(3,2) NOT NULL DEFAULT 0.00,              -- 质量分
    is_used SMALLINT NOT NULL DEFAULT 0,                          -- 是否已用于训练
    used_model_id BIGINT NOT NULL DEFAULT 0,
    extra JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_ai_training_data_type ON ai_training_data(data_type);
CREATE INDEX IF NOT EXISTS idx_ai_training_data_biz ON ai_training_data(biz_module, biz_id);
CREATE INDEX IF NOT EXISTS idx_ai_training_data_label ON ai_training_data(label);
CREATE INDEX IF NOT EXISTS idx_ai_training_data_is_used ON ai_training_data(is_used);
CREATE INDEX IF NOT EXISTS idx_ai_training_data_region_id ON ai_training_data(region_id);
COMMENT ON TABLE ai_training_data IS 'AI 训练数据表';
