-- ============================================================
-- P1 中台精简版迁移脚本（pay / im / material / risk / ai）
-- 依据 v3.2.1 架构方案 + ershou 模块依赖
-- 创建 5 个中台精简版的全部数据表（共 26 张表）
-- 幂等：所有对象使用 IF NOT EXISTS 创建，可重复执行
-- ============================================================

-- ============================================================
-- 一、pay 支付财务中台精简版（7 张表）
-- 支付方式 + 担保交易 + 订单 + 退款 + 提现 + 结算 + 资金账户
-- ============================================================

-- 1.0 pay_methods 支付方式表
CREATE TABLE IF NOT EXISTS pay_methods (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    method_code VARCHAR(32) NOT NULL UNIQUE,
    method_name VARCHAR(64) NOT NULL DEFAULT '',
    icon VARCHAR(256) NOT NULL DEFAULT '',
    description VARCHAR(256) NOT NULL DEFAULT '',
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    fee_rate DECIMAL(6,4) NOT NULL DEFAULT 0.0000,
    fee_fixed DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    sort INTEGER NOT NULL DEFAULT 0,
    status SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_pay_methods_region_id ON pay_methods(region_id);
CREATE INDEX IF NOT EXISTS idx_pay_methods_status ON pay_methods(status);
COMMENT ON TABLE pay_methods IS '支付方式配置表';
COMMENT ON COLUMN pay_methods.method_code IS '支付方式码：wechat/alipay/balance/bank_card/point/giftcard';
COMMENT ON COLUMN pay_methods.config IS '支付配置 JSON（app_id/mch_id/回调地址等）';
COMMENT ON COLUMN pay_methods.fee_rate IS '费率（小数，0.006=千分之六）';
COMMENT ON COLUMN pay_methods.fee_fixed IS '固定手续费';
COMMENT ON COLUMN pay_methods.status IS '状态：1启用 0禁用';

-- 1.1 pay_orders 支付订单表
CREATE TABLE IF NOT EXISTS pay_orders (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    order_no VARCHAR(64) NOT NULL UNIQUE,
    user_id BIGINT NOT NULL,
    biz_module VARCHAR(32) NOT NULL DEFAULT '',
    biz_id VARCHAR(128) NOT NULL DEFAULT '',
    title VARCHAR(256) NOT NULL DEFAULT '',
    amount DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    pay_method VARCHAR(32) NOT NULL DEFAULT '',
    pay_status SMALLINT NOT NULL DEFAULT 0,
    third_party_no VARCHAR(128) NOT NULL DEFAULT '',
    paid_at TIMESTAMPTZ,
    expire_at TIMESTAMPTZ,
    extra JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_pay_orders_user_id ON pay_orders(user_id);
CREATE INDEX IF NOT EXISTS idx_pay_orders_biz ON pay_orders(biz_module, biz_id);
CREATE INDEX IF NOT EXISTS idx_pay_orders_pay_status ON pay_orders(pay_status);
CREATE INDEX IF NOT EXISTS idx_pay_orders_region_id ON pay_orders(region_id);
CREATE INDEX IF NOT EXISTS idx_pay_orders_deleted_at ON pay_orders(deleted_at);
COMMENT ON TABLE pay_orders IS '支付订单表';
COMMENT ON COLUMN pay_orders.order_no IS '订单号（业务唯一）';
COMMENT ON COLUMN pay_orders.biz_module IS '业务模块名（如 ershou/groupbuy）';
COMMENT ON COLUMN pay_orders.biz_id IS '业务订单ID';
COMMENT ON COLUMN pay_orders.pay_method IS '支付方式：wechat/alipay/balance/point/giftcard';
COMMENT ON COLUMN pay_orders.pay_status IS '支付状态：0待支付 1已支付 2已关闭 3已退款 4部分退款';
COMMENT ON COLUMN pay_orders.third_party_no IS '第三方流水号';

-- 1.2 pay_escrows 担保交易表
CREATE TABLE IF NOT EXISTS pay_escrows (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    order_id BIGINT NOT NULL,
    order_no VARCHAR(64) NOT NULL DEFAULT '',
    user_id BIGINT NOT NULL,
    merchant_id BIGINT NOT NULL DEFAULT 0,
    amount DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    status SMALLINT NOT NULL DEFAULT 0,
    frozen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    release_at TIMESTAMPTZ,
    auto_release_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_pay_escrows_order_id ON pay_escrows(order_id);
CREATE INDEX IF NOT EXISTS idx_pay_escrows_user_id ON pay_escrows(user_id);
CREATE INDEX IF NOT EXISTS idx_pay_escrows_status ON pay_escrows(status);
CREATE INDEX IF NOT EXISTS idx_pay_escrows_auto_release ON pay_escrows(auto_release_at) WHERE status = 0;
COMMENT ON TABLE pay_escrows IS '担保交易表（资金托管→确认收货→放款）';
COMMENT ON COLUMN pay_escrows.status IS '状态：0冻结中 1已放款 2已退款 3已部分退款';
COMMENT ON COLUMN pay_escrows.auto_release_at IS '自动放款时间（确认收货 N 天后）';

-- 1.3 pay_refunds 退款单表
CREATE TABLE IF NOT EXISTS pay_refunds (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    refund_no VARCHAR(64) NOT NULL UNIQUE,
    order_id BIGINT NOT NULL,
    order_no VARCHAR(64) NOT NULL DEFAULT '',
    user_id BIGINT NOT NULL,
    amount DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    reason VARCHAR(256) NOT NULL DEFAULT '',
    status SMALLINT NOT NULL DEFAULT 0,
    third_party_refund_no VARCHAR(128) NOT NULL DEFAULT '',
    refund_method VARCHAR(32) NOT NULL DEFAULT '',
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_pay_refunds_order_id ON pay_refunds(order_id);
CREATE INDEX IF NOT EXISTS idx_pay_refunds_user_id ON pay_refunds(user_id);
CREATE INDEX IF NOT EXISTS idx_pay_refunds_status ON pay_refunds(status);
COMMENT ON TABLE pay_refunds IS '退款单表（原路返回）';
COMMENT ON COLUMN pay_refunds.status IS '状态：0待处理 1已退款 2已拒绝 3处理中';

-- 1.4 pay_withdrawals 提现表
CREATE TABLE IF NOT EXISTS pay_withdrawals (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    withdrawal_no VARCHAR(64) NOT NULL UNIQUE,
    user_id BIGINT NOT NULL,
    amount DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    fee DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    actual_amount DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    bank_card JSONB NOT NULL DEFAULT '{}'::jsonb,
    status SMALLINT NOT NULL DEFAULT 0,
    reject_reason VARCHAR(256) NOT NULL DEFAULT '',
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_pay_withdrawals_user_id ON pay_withdrawals(user_id);
CREATE INDEX IF NOT EXISTS idx_pay_withdrawals_status ON pay_withdrawals(status);
COMMENT ON TABLE pay_withdrawals IS '提现申请表';
COMMENT ON COLUMN pay_withdrawals.bank_card IS '银行卡信息 JSON：{card_no,name,bank,phone}';
COMMENT ON COLUMN pay_withdrawals.status IS '状态：0待审核 1已通过 2已拒绝 3已打款 4打款失败';

-- 1.5 pay_settlements 结算单表
CREATE TABLE IF NOT EXISTS pay_settlements (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    settlement_no VARCHAR(64) NOT NULL UNIQUE,
    merchant_id BIGINT NOT NULL,
    period_type VARCHAR(16) NOT NULL DEFAULT 'T1',
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    order_count INTEGER NOT NULL DEFAULT 0,
    total_amount DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    commission DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    settlement_amount DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    status SMALLINT NOT NULL DEFAULT 0,
    settled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_pay_settlements_merchant_id ON pay_settlements(merchant_id);
CREATE INDEX IF NOT EXISTS idx_pay_settlements_status ON pay_settlements(status);
CREATE INDEX IF NOT EXISTS idx_pay_settlements_period ON pay_settlements(period_start, period_end);
COMMENT ON TABLE pay_settlements IS '商家结算单表';
COMMENT ON COLUMN pay_settlements.period_type IS '结算周期：T1/T7/monthly';
COMMENT ON COLUMN pay_settlements.commission IS '平台抽成';
COMMENT ON COLUMN pay_settlements.settlement_amount IS '实际到账金额';
COMMENT ON COLUMN pay_settlements.status IS '状态：0待结算 1已结算 2已失败';

-- 1.6 pay_accounts 资金账户表
CREATE TABLE IF NOT EXISTS pay_accounts (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    user_id BIGINT NOT NULL UNIQUE,
    balance DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    frozen_amount DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    total_income DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    total_expense DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    bank_cards JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_pay_accounts_user_id ON pay_accounts(user_id);
COMMENT ON TABLE pay_accounts IS '用户资金账户表（余额/冻结/累计收入/累计支出）';
COMMENT ON COLUMN pay_accounts.balance IS '可用余额';
COMMENT ON COLUMN pay_accounts.frozen_amount IS '冻结金额（担保交易占用）';
COMMENT ON COLUMN pay_accounts.bank_cards IS '银行卡列表 JSON 数组';

-- ============================================================
-- 二、im IM 消息中台精简版（5 张表）
-- 私聊 + 系统通知 + 隐私号码 + 消息模板
-- ============================================================

-- 2.1 im_sessions 会话表
CREATE TABLE IF NOT EXISTS im_sessions (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    session_id VARCHAR(64) NOT NULL UNIQUE,
    session_type VARCHAR(16) NOT NULL DEFAULT 'private',
    participants JSONB NOT NULL DEFAULT '[]'::jsonb,
    last_message JSONB NOT NULL DEFAULT '{}'::jsonb,
    last_message_at TIMESTAMPTZ,
    unread_count JSONB NOT NULL DEFAULT '{}'::jsonb,
    status SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_im_sessions_participants ON im_sessions USING GIN (participants);
CREATE INDEX IF NOT EXISTS idx_im_sessions_last_message_at ON im_sessions(last_message_at);
CREATE INDEX IF NOT EXISTS idx_im_sessions_status ON im_sessions(status);
COMMENT ON TABLE im_sessions IS 'IM 会话表';
COMMENT ON COLUMN im_sessions.session_type IS '会话类型：private/group';
COMMENT ON COLUMN im_sessions.participants IS '参与者 ID JSON 数组：[1,2,3]';
COMMENT ON COLUMN im_sessions.last_message IS '最后一条消息 JSON 快照';
COMMENT ON COLUMN im_sessions.unread_count IS '各用户未读数 JSON：{"1":5,"2":0}';

-- 2.2 im_messages 消息表
CREATE TABLE IF NOT EXISTS im_messages (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    session_id VARCHAR(64) NOT NULL,
    sender_id BIGINT NOT NULL,
    msg_type VARCHAR(16) NOT NULL DEFAULT 'text',
    content TEXT NOT NULL DEFAULT '',
    extra JSONB NOT NULL DEFAULT '{}'::jsonb,
    read_status JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_recalled BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_im_messages_session_id ON im_messages(session_id);
CREATE INDEX IF NOT EXISTS idx_im_messages_sender_id ON im_messages(sender_id);
CREATE INDEX IF NOT EXISTS idx_im_messages_created_at ON im_messages(created_at);
COMMENT ON TABLE im_messages IS 'IM 消息表';
COMMENT ON COLUMN im_messages.msg_type IS '消息类型：text/image/voice/video/card';
COMMENT ON COLUMN im_messages.extra IS '扩展信息 JSON（图片URL/语音时长/视频封面等）';
COMMENT ON COLUMN im_messages.read_status IS '各用户已读状态 JSON：{"1":true,"2":false}';

-- 2.3 im_system_notifications 系统通知表
CREATE TABLE IF NOT EXISTS im_system_notifications (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    user_id BIGINT NOT NULL,
    notify_type VARCHAR(32) NOT NULL DEFAULT '',
    title VARCHAR(128) NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    jump_url VARCHAR(256) NOT NULL DEFAULT '',
    extra JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_im_sys_notif_user_id ON im_system_notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_im_sys_notif_is_read ON im_system_notifications(is_read);
CREATE INDEX IF NOT EXISTS idx_im_sys_notif_type ON im_system_notifications(notify_type);
COMMENT ON TABLE im_system_notifications IS '系统通知表（交易状态变更/活动通知）';
COMMENT ON COLUMN im_system_notifications.notify_type IS '通知类型：order/refund/activity/system';

-- 2.4 im_privacy_numbers 隐私号码表
CREATE TABLE IF NOT EXISTS im_privacy_numbers (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    privacy_no VARCHAR(32) NOT NULL,
    real_no_a VARCHAR(32) NOT NULL,
    real_no_b VARCHAR(32) NOT NULL,
    user_id_a BIGINT NOT NULL,
    user_id_b BIGINT NOT NULL,
    biz_module VARCHAR(32) NOT NULL DEFAULT '',
    biz_id VARCHAR(128) NOT NULL DEFAULT '',
    call_records JSONB NOT NULL DEFAULT '[]'::jsonb,
    bound_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    unbound_at TIMESTAMPTZ,
    status SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_im_privacy_numbers_privacy_no ON im_privacy_numbers(privacy_no);
CREATE INDEX IF NOT EXISTS idx_im_privacy_numbers_user_a ON im_privacy_numbers(user_id_a);
CREATE INDEX IF NOT EXISTS idx_im_privacy_numbers_user_b ON im_privacy_numbers(user_id_b);
CREATE INDEX IF NOT EXISTS idx_im_privacy_numbers_status ON im_privacy_numbers(status);
COMMENT ON TABLE im_privacy_numbers IS '隐私号码表（虚拟号绑定/解绑/通话记录）';
COMMENT ON COLUMN im_privacy_numbers.call_records IS '通话记录 JSON 数组：[{start,end,duration}]';
COMMENT ON COLUMN im_privacy_numbers.status IS '状态：1绑定中 0已解绑';

-- 2.5 im_templates 消息模板表
CREATE TABLE IF NOT EXISTS im_templates (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    template_code VARCHAR(64) NOT NULL UNIQUE,
    template_name VARCHAR(128) NOT NULL DEFAULT '',
    template_type VARCHAR(32) NOT NULL DEFAULT 'system',
    title VARCHAR(128) NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    variables JSONB NOT NULL DEFAULT '[]'::jsonb,
    jump_url VARCHAR(256) NOT NULL DEFAULT '',
    description VARCHAR(256) NOT NULL DEFAULT '',
    status SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_im_templates_region_id ON im_templates(region_id);
CREATE INDEX IF NOT EXISTS idx_im_templates_template_type ON im_templates(template_type);
CREATE INDEX IF NOT EXISTS idx_im_templates_status ON im_templates(status);
COMMENT ON TABLE im_templates IS '消息模板表（系统通知/活动/订单/退款等）';
COMMENT ON COLUMN im_templates.template_code IS '模板码（业务唯一）';
COMMENT ON COLUMN im_templates.template_type IS '模板类型：system/order/refund/activity/welcome';
COMMENT ON COLUMN im_templates.content IS '模板内容，支持 {{var}} 变量替换';
COMMENT ON COLUMN im_templates.variables IS '变量列表 JSON：["username","order_no"]';

-- ============================================================
-- 三、material 素材存储中台精简版（4 张表）
-- 图片/视频/文件 + 以图搜图
-- ============================================================

-- 3.1 mat_files 文件表
CREATE TABLE IF NOT EXISTS mat_files (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    file_id VARCHAR(64) NOT NULL UNIQUE,
    user_id BIGINT NOT NULL DEFAULT 0,
    file_type VARCHAR(16) NOT NULL DEFAULT 'image',
    file_url VARCHAR(512) NOT NULL DEFAULT '',
    file_size BIGINT NOT NULL DEFAULT 0,
    mime_type VARCHAR(64) NOT NULL DEFAULT '',
    file_hash VARCHAR(64) NOT NULL DEFAULT '',
    original_name VARCHAR(256) NOT NULL DEFAULT '',
    category VARCHAR(32) NOT NULL DEFAULT 'user',
    storage_driver VARCHAR(32) NOT NULL DEFAULT 'local',
    extra JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_mat_files_user_id ON mat_files(user_id);
CREATE INDEX IF NOT EXISTS idx_mat_files_file_type ON mat_files(file_type);
CREATE INDEX IF NOT EXISTS idx_mat_files_file_hash ON mat_files(file_hash);
CREATE INDEX IF NOT EXISTS idx_mat_files_category ON mat_files(category);
COMMENT ON TABLE mat_files IS '素材文件表（图片/视频/文档）';
COMMENT ON COLUMN mat_files.file_type IS '文件类型：image/video/document';
COMMENT ON COLUMN mat_files.category IS '素材分类：user用户/merchant商家/operation运营';
COMMENT ON COLUMN mat_files.storage_driver IS '存储驱动：local/minio/qiniu';

-- 3.2 mat_images 图片表
CREATE TABLE IF NOT EXISTS mat_images (
    id BIGSERIAL PRIMARY KEY,
    file_id VARCHAR(64) NOT NULL UNIQUE,
    region_id BIGINT NOT NULL DEFAULT 1,
    width INTEGER NOT NULL DEFAULT 0,
    height INTEGER NOT NULL DEFAULT 0,
    thumbnails JSONB NOT NULL DEFAULT '{}'::jsonb,
    phash VARCHAR(64) NOT NULL DEFAULT '',
    color_histogram JSONB NOT NULL DEFAULT '[]'::jsonb,
    watermarked BOOLEAN NOT NULL DEFAULT FALSE,
    extra JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_mat_images_phash ON mat_images(phash);
COMMENT ON TABLE mat_images IS '图片元数据表';
COMMENT ON COLUMN mat_images.thumbnails IS '缩略图 URL JSON：{"100x100":"url","300x300":"url","800x800":"url"}';
COMMENT ON COLUMN mat_images.phash IS '感知哈希（用于以图搜图）';
COMMENT ON COLUMN mat_images.color_histogram IS '颜色直方图 JSON 数组';

-- 3.3 mat_videos 视频表
CREATE TABLE IF NOT EXISTS mat_videos (
    id BIGSERIAL PRIMARY KEY,
    file_id VARCHAR(64) NOT NULL UNIQUE,
    region_id BIGINT NOT NULL DEFAULT 1,
    duration INTEGER NOT NULL DEFAULT 0,
    resolution VARCHAR(16) NOT NULL DEFAULT '',
    codec VARCHAR(16) NOT NULL DEFAULT '',
    cover_url VARCHAR(512) NOT NULL DEFAULT '',
    transcode_status SMALLINT NOT NULL DEFAULT 0,
    transcode_jobs JSONB NOT NULL DEFAULT '[]'::jsonb,
    extra JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_mat_videos_transcode_status ON mat_videos(transcode_status);
COMMENT ON TABLE mat_videos IS '视频元数据表';
COMMENT ON COLUMN mat_videos.resolution IS '分辨率：480p/720p/1080p';
COMMENT ON COLUMN mat_videos.codec IS '编码：H.264/H.265';
COMMENT ON COLUMN mat_videos.transcode_status IS '转码状态：0待转码 1转码中 2已完成 3失败';
COMMENT ON COLUMN mat_videos.transcode_jobs IS '转码任务 JSON 数组';

-- 3.4 mat_image_features 图片特征向量表
CREATE TABLE IF NOT EXISTS mat_image_features (
    id BIGSERIAL PRIMARY KEY,
    image_id BIGINT NOT NULL UNIQUE,
    file_id VARCHAR(64) NOT NULL DEFAULT '',
    region_id BIGINT NOT NULL DEFAULT 1,
    phash VARCHAR(64) NOT NULL DEFAULT '',
    feature_vector JSONB NOT NULL DEFAULT '[]'::jsonb,
    color_histogram JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_mat_image_features_phash ON mat_image_features(phash);
CREATE INDEX IF NOT EXISTS idx_mat_image_features_image_id ON mat_image_features(image_id);
COMMENT ON TABLE mat_image_features IS '图片特征向量表（以图搜图用）';
COMMENT ON COLUMN mat_image_features.feature_vector IS '特征向量 JSON 数组';

-- ============================================================
-- 四、risk 风控审核中台精简版（6 张表）
-- 举报 + 敏感词 + 审核规则 + 黑名单 + 风险分 + 违规处罚
-- ============================================================

-- 4.1 risk_reports 举报工单表
CREATE TABLE IF NOT EXISTS risk_reports (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    report_no VARCHAR(64) NOT NULL UNIQUE,
    reporter_id BIGINT NOT NULL,
    reported_user_id BIGINT NOT NULL DEFAULT 0,
    reported_biz_module VARCHAR(32) NOT NULL DEFAULT '',
    reported_biz_id VARCHAR(128) NOT NULL DEFAULT '',
    report_type VARCHAR(32) NOT NULL DEFAULT '',
    reason VARCHAR(512) NOT NULL DEFAULT '',
    evidence_images JSONB NOT NULL DEFAULT '[]'::jsonb,
    status SMALLINT NOT NULL DEFAULT 0,
    handle_result VARCHAR(32) NOT NULL DEFAULT '',
    handle_remark TEXT NOT NULL DEFAULT '',
    handler_id BIGINT NOT NULL DEFAULT 0,
    handled_at TIMESTAMPTZ,
    sla_deadline TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_risk_reports_reporter_id ON risk_reports(reporter_id);
CREATE INDEX IF NOT EXISTS idx_risk_reports_reported_user_id ON risk_reports(reported_user_id);
CREATE INDEX IF NOT EXISTS idx_risk_reports_biz ON risk_reports(reported_biz_module, reported_biz_id);
CREATE INDEX IF NOT EXISTS idx_risk_reports_status ON risk_reports(status);
CREATE INDEX IF NOT EXISTS idx_risk_reports_sla ON risk_reports(sla_deadline) WHERE status = 0;
COMMENT ON TABLE risk_reports IS '举报工单表';
COMMENT ON COLUMN risk_reports.report_type IS '举报类型：spam/fraud/porn/violence/contraband/other';
COMMENT ON COLUMN risk_reports.status IS '状态：0待处理 1处理中 2已处理 3已仲裁 4已撤销';
COMMENT ON COLUMN risk_reports.handle_result IS '处理结果：warning/remove/ban/note';
COMMENT ON COLUMN risk_reports.sla_deadline IS 'SLA 到期时间';

-- 4.2 risk_sensitive_words 敏感词库表
CREATE TABLE IF NOT EXISTS risk_sensitive_words (
    id BIGSERIAL PRIMARY KEY,
    word VARCHAR(64) NOT NULL UNIQUE,
    word_type VARCHAR(32) NOT NULL DEFAULT 'politics',
    category VARCHAR(32) NOT NULL DEFAULT '',
    replacement VARCHAR(32) NOT NULL DEFAULT '***',
    status SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_risk_sensitive_words_type ON risk_sensitive_words(word_type);
CREATE INDEX IF NOT EXISTS idx_risk_sensitive_words_status ON risk_sensitive_words(status);
COMMENT ON TABLE risk_sensitive_words IS '敏感词库表（DFA 算法）';
COMMENT ON COLUMN risk_sensitive_words.word_type IS '类型：politics/porn/violence/contraband/ad';
COMMENT ON COLUMN risk_sensitive_words.replacement IS '替换字符';

-- 4.3 risk_audit_rules 审核规则表
CREATE TABLE IF NOT EXISTS risk_audit_rules (
    id BIGSERIAL PRIMARY KEY,
    rule_name VARCHAR(64) NOT NULL UNIQUE,
    rule_type VARCHAR(32) NOT NULL DEFAULT '',
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    description VARCHAR(256) NOT NULL DEFAULT '',
    status SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_risk_audit_rules_type ON risk_audit_rules(rule_type);
CREATE INDEX IF NOT EXISTS idx_risk_audit_rules_status ON risk_audit_rules(status);
COMMENT ON TABLE risk_audit_rules IS '审核规则表';
COMMENT ON COLUMN risk_audit_rules.rule_type IS '规则类型：sensitive_word/price/frequency/contraband';
COMMENT ON COLUMN risk_audit_rules.config IS '规则配置 JSON';

-- 4.4 risk_blacklist 黑名单表
CREATE TABLE IF NOT EXISTS risk_blacklist (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    target_type VARCHAR(16) NOT NULL DEFAULT 'user',
    target_value VARCHAR(128) NOT NULL,
    reason VARCHAR(256) NOT NULL DEFAULT '',
    operator_id BIGINT NOT NULL DEFAULT 0,
    expire_at TIMESTAMPTZ,
    status SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_risk_blacklist_target ON risk_blacklist(target_type, target_value);
CREATE INDEX IF NOT EXISTS idx_risk_blacklist_status ON risk_blacklist(status);
COMMENT ON TABLE risk_blacklist IS '黑名单表（用户/IP/设备）';
COMMENT ON COLUMN risk_blacklist.target_type IS '目标类型：user/ip/device';
COMMENT ON COLUMN risk_blacklist.target_value IS '目标值（用户ID/IP/设备指纹）';

-- 4.5 risk_user_scores 用户风险分表
CREATE TABLE IF NOT EXISTS risk_user_scores (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    user_id BIGINT NOT NULL UNIQUE,
    score INTEGER NOT NULL DEFAULT 100,
    level VARCHAR(16) NOT NULL DEFAULT 'safe',
    violation_count INTEGER NOT NULL DEFAULT 0,
    report_count INTEGER NOT NULL DEFAULT 0,
    last_violation_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_risk_user_scores_user_id ON risk_user_scores(user_id);
CREATE INDEX IF NOT EXISTS idx_risk_user_scores_score ON risk_user_scores(score);
COMMENT ON TABLE risk_user_scores IS '用户风险分表（0-100，低于30限制交易）';
COMMENT ON COLUMN risk_user_scores.score IS '风险分：0-100，越低越危险';
COMMENT ON COLUMN risk_user_scores.level IS '风险等级：safe/warning/danger';

-- 4.6 risk_violations 违规处罚表
CREATE TABLE IF NOT EXISTS risk_violations (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    user_id BIGINT NOT NULL,
    violation_type VARCHAR(32) NOT NULL DEFAULT '',
    level VARCHAR(16) NOT NULL DEFAULT 'warning',
    reason VARCHAR(512) NOT NULL DEFAULT '',
    biz_module VARCHAR(32) NOT NULL DEFAULT '',
    biz_id VARCHAR(128) NOT NULL DEFAULT '',
    report_id BIGINT NOT NULL DEFAULT 0,
    penalty_start TIMESTAMPTZ,
    penalty_end TIMESTAMPTZ,
    status SMALLINT NOT NULL DEFAULT 1,
    appeal_status SMALLINT NOT NULL DEFAULT 0,
    appeal_remark TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_risk_violations_user_id ON risk_violations(user_id);
CREATE INDEX IF NOT EXISTS idx_risk_violations_status ON risk_violations(status);
CREATE INDEX IF NOT EXISTS idx_risk_violations_penalty_end ON risk_violations(penalty_end) WHERE status = 1;
COMMENT ON TABLE risk_violations IS '违规处罚表';
COMMENT ON COLUMN risk_violations.violation_type IS '违规类型：spam/fraud/porn/violence/contraband/ad';
COMMENT ON COLUMN risk_violations.level IS '处罚级别：warning/limit/mute/ban_1d/ban_7d/ban_forever';
COMMENT ON COLUMN risk_violations.status IS '状态：1生效中 0已结束 2已申诉撤销';
COMMENT ON COLUMN risk_violations.appeal_status IS '申诉状态：0未申诉 1申诉中 2申诉成功 3申诉失败';

-- ============================================================
-- 五、ai AI 智能中台精简版（4 张表）
-- 图文审核 + 标题优化 + 价格建议 + 描述生成
-- ============================================================

-- 5.1 ai_tasks AI 任务表
CREATE TABLE IF NOT EXISTS ai_tasks (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    task_id VARCHAR(64) NOT NULL UNIQUE,
    task_type VARCHAR(32) NOT NULL DEFAULT '',
    user_id BIGINT NOT NULL DEFAULT 0,
    input JSONB NOT NULL DEFAULT '{}'::jsonb,
    output JSONB NOT NULL DEFAULT '{}'::jsonb,
    status SMALLINT NOT NULL DEFAULT 0,
    model_name VARCHAR(64) NOT NULL DEFAULT '',
    cost_ms INTEGER NOT NULL DEFAULT 0,
    error_msg TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_ai_tasks_task_type ON ai_tasks(task_type);
CREATE INDEX IF NOT EXISTS idx_ai_tasks_user_id ON ai_tasks(user_id);
CREATE INDEX IF NOT EXISTS idx_ai_tasks_status ON ai_tasks(status);
COMMENT ON TABLE ai_tasks IS 'AI 任务表';
COMMENT ON COLUMN ai_tasks.task_type IS '任务类型：audit_image/audit_text/optimize_title/generate_description/suggest_price/generate_summary';
COMMENT ON COLUMN ai_tasks.status IS '状态：0待处理 1处理中 2已完成 3失败';

-- 5.2 ai_models AI 模型配置表
CREATE TABLE IF NOT EXISTS ai_models (
    id BIGSERIAL PRIMARY KEY,
    model_name VARCHAR(64) NOT NULL UNIQUE,
    provider VARCHAR(32) NOT NULL DEFAULT '',
    model_type VARCHAR(32) NOT NULL DEFAULT '',
    api_key VARCHAR(256) NOT NULL DEFAULT '',
    endpoint VARCHAR(256) NOT NULL DEFAULT '',
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    status SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_ai_models_provider ON ai_models(provider);
CREATE INDEX IF NOT EXISTS idx_ai_models_status ON ai_models(status);
COMMENT ON TABLE ai_models IS 'AI 模型配置表';
COMMENT ON COLUMN ai_models.provider IS '提供商：aliyun/tencent/qwen/wenxin/xfyun';
COMMENT ON COLUMN ai_models.model_type IS '类型：audit_image/audit_text/llm';

-- 5.3 ai_prompts 提示词模板表
CREATE TABLE IF NOT EXISTS ai_prompts (
    id BIGSERIAL PRIMARY KEY,
    template_name VARCHAR(64) NOT NULL UNIQUE,
    template_type VARCHAR(32) NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    variables JSONB NOT NULL DEFAULT '[]'::jsonb,
    description VARCHAR(256) NOT NULL DEFAULT '',
    status SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_ai_prompts_template_type ON ai_prompts(template_type);
CREATE INDEX IF NOT EXISTS idx_ai_prompts_status ON ai_prompts(status);
COMMENT ON TABLE ai_prompts IS 'AI 提示词模板表';
COMMENT ON COLUMN ai_prompts.template_type IS '模板类型：optimize_title/generate_description/suggest_price/audit_text';
COMMENT ON COLUMN ai_prompts.variables IS '变量列表 JSON：["title","brand","condition"]';

-- 5.4 ai_generations AI 生成记录表
CREATE TABLE IF NOT EXISTS ai_generations (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    task_id VARCHAR(64) NOT NULL DEFAULT '',
    user_id BIGINT NOT NULL,
    generation_type VARCHAR(32) NOT NULL DEFAULT '',
    input JSONB NOT NULL DEFAULT '{}'::jsonb,
    output JSONB NOT NULL DEFAULT '{}'::jsonb,
    rating SMALLINT NOT NULL DEFAULT 0,
    feedback TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_ai_generations_user_id ON ai_generations(user_id);
CREATE INDEX IF NOT EXISTS idx_ai_generations_generation_type ON ai_generations(generation_type);
CREATE INDEX IF NOT EXISTS idx_ai_generations_task_id ON ai_generations(task_id);
COMMENT ON TABLE ai_generations IS 'AI 生成记录表';
COMMENT ON COLUMN ai_generations.generation_type IS '生成类型：title/description/price/summary';
COMMENT ON COLUMN ai_generations.rating IS '用户评分 1-5，0表示未评分';

-- ============================================================
-- updated_at 触发器（依赖 001_p0_baseline.sql 中的 update_updated_at_column 函数）
-- 幂等：先 DROP IF EXISTS 再 CREATE
-- ============================================================
DO $$
DECLARE t TEXT;
BEGIN
    FOR t IN SELECT tablename FROM pg_tables WHERE schemaname='public' AND tablename IN (
        'pay_methods','pay_orders','pay_escrows','pay_refunds','pay_withdrawals','pay_settlements','pay_accounts',
        'im_sessions','im_messages','im_system_notifications','im_privacy_numbers','im_templates',
        'mat_files','mat_images','mat_videos','mat_image_features',
        'risk_reports','risk_sensitive_words','risk_audit_rules','risk_blacklist','risk_user_scores','risk_violations',
        'ai_tasks','ai_models','ai_prompts','ai_generations'
    )
    LOOP
        EXECUTE format('DROP TRIGGER IF EXISTS trg_%s_updated ON %s', t, t);
        EXECUTE format('CREATE TRIGGER trg_%s_updated BEFORE UPDATE ON %s FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()', t, t);
    END LOOP;
END $$;
