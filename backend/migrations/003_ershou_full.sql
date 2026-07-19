-- ============================================================
-- ershou 二手交易模块完整功能迁移脚本（v3.2.1）
-- 对标：闲鱼 / 转转 / 58同城 / 瓜子 / 贝壳 / 趣店
--
-- 内容：
--   1. ALTER TABLE erhous 主表新增 35+ 字段（SKU/拍卖/议价/物流/担保/分期/隐私/互动/物品信息/包装/使用情况/估值/推广/风控/关联/视频/360/多地交易/标签/运营）
--   2. CREATE 19 张子表（ers_ 前缀，依据数据库分表前缀规范 v1.0.0）
--   3. 索引、外键、触发器、注释
--   4. 全幂等：CREATE TABLE IF NOT EXISTS / ALTER TABLE ADD COLUMN IF NOT EXISTS
-- 依赖：001_p0_baseline.sql（update_updated_at_column 函数已创建）
-- ============================================================

-- ============================================================
-- 第一部分：扩展 erhous 主表（保持现有表名兼容已发布数据）
-- 注意：erhous 表由 GORM AutoMigrate 在应用启动时创建
--      本迁移对 erhous 的所有 ALTER 操作包装在 DO 块中，
--      若表不存在则跳过（待应用启动后再执行一次本迁移即可补齐字段）
-- ============================================================
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'erhous') THEN
        -- === SKU 相关 ===
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS sku_enabled BOOLEAN NOT NULL DEFAULT FALSE;
        CREATE INDEX IF NOT EXISTS idx_erhous_sku_enabled ON erhous(sku_enabled) WHERE sku_enabled = TRUE;

        -- === 拍卖相关 ===
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS is_auction BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS auction_start_time TIMESTAMPTZ;
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS auction_end_time TIMESTAMPTZ;
        CREATE INDEX IF NOT EXISTS idx_erhous_is_auction ON erhous(is_auction) WHERE is_auction = TRUE;
        CREATE INDEX IF NOT EXISTS idx_erhous_auction_start_time ON erhous(auction_start_time);
        CREATE INDEX IF NOT EXISTS idx_erhous_auction_end_time ON erhous(auction_end_time);

        -- === 议价相关 ===
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS bargain_type VARCHAR(20) NOT NULL DEFAULT 'small';

        -- === 物流相关 ===
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS delivery_fee DECIMAL(12,2) NOT NULL DEFAULT 0;
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS free_shipping BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS logistics_template_id BIGINT;
        CREATE INDEX IF NOT EXISTS idx_erhous_free_shipping ON erhous(free_shipping) WHERE free_shipping = TRUE;
        CREATE INDEX IF NOT EXISTS idx_erhous_logistics_template_id ON erhous(logistics_template_id);

        -- === 担保相关 ===
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS escrow_enabled BOOLEAN NOT NULL DEFAULT FALSE;
        CREATE INDEX IF NOT EXISTS idx_erhous_escrow_enabled ON erhous(escrow_enabled) WHERE escrow_enabled = TRUE;

        -- === 分期相关 ===
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS installment_enabled BOOLEAN NOT NULL DEFAULT FALSE;
        CREATE INDEX IF NOT EXISTS idx_erhous_installment_enabled ON erhous(installment_enabled) WHERE installment_enabled = TRUE;

        -- === 隐私设置 ===
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS visibility VARCHAR(20) NOT NULL DEFAULT 'public';
        CREATE INDEX IF NOT EXISTS idx_erhous_visibility ON erhous(visibility);

        -- === 互动开关 ===
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS allow_comment BOOLEAN NOT NULL DEFAULT TRUE;
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS allow_share BOOLEAN NOT NULL DEFAULT TRUE;

        -- === 出手原因 + 物品来源 ===
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS sell_reason VARCHAR(50);
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS item_source VARCHAR(50) NOT NULL DEFAULT 'personal';
        CREATE INDEX IF NOT EXISTS idx_erhous_item_source ON erhous(item_source);

        -- === 鉴定认证 ===
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS certification_type VARCHAR(50);
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS certification_no VARCHAR(100);

        -- === 包装附件 ===
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS has_original_box BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS has_accessories BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS has_warranty_card BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS has_invoice BOOLEAN NOT NULL DEFAULT FALSE;

        -- === 使用情况 ===
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS use_duration VARCHAR(50);
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS repair_history TEXT;
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS warranty_expire DATE;
        CREATE INDEX IF NOT EXISTS idx_erhous_warranty_expire ON erhous(warranty_expire);

        -- === 估值参考 ===
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS estimated_value DECIMAL(12,2) NOT NULL DEFAULT 0;

        -- === 推广加权 ===
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS promotion_level INT NOT NULL DEFAULT 0;
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS traffic_weight DECIMAL(3,2) NOT NULL DEFAULT 1.00;
        CREATE INDEX IF NOT EXISTS idx_erhous_promotion_level ON erhous(promotion_level);

        -- === 风控 ===
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS content_hash VARCHAR(64);
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS risk_score INT NOT NULL DEFAULT 0;
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS same_item_id VARCHAR(64);
        CREATE INDEX IF NOT EXISTS idx_erhous_content_hash ON erhous(content_hash);
        CREATE INDEX IF NOT EXISTS idx_erhous_risk_score ON erhous(risk_score);
        CREATE INDEX IF NOT EXISTS idx_erhous_same_item_id ON erhous(same_item_id);

        -- === 关联 ===
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS shop_id BIGINT;
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS brand_id BIGINT;
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS model_id BIGINT;
        CREATE INDEX IF NOT EXISTS idx_erhous_shop_id ON erhous(shop_id);
        CREATE INDEX IF NOT EXISTS idx_erhous_brand_id ON erhous(brand_id);
        CREATE INDEX IF NOT EXISTS idx_erhous_model_id ON erhous(model_id);

        -- === 视频支持 ===
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS video_url VARCHAR(255);
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS video_cover VARCHAR(255);

        -- === 360 展示 ===
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS panorama_url VARCHAR(255);

        -- === 多地交易 ===
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS trade_locations JSONB;
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS pickup_time_slots JSONB;

        -- === 标签冗余 ===
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS tags JSONB;
        CREATE INDEX IF NOT EXISTS idx_erhous_tags ON erhous USING GIN(tags);

        -- === 运营字段 ===
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS featured BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS picked BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE erhous ADD COLUMN IF NOT EXISTS verified BOOLEAN NOT NULL DEFAULT FALSE;
        CREATE INDEX IF NOT EXISTS idx_erhous_featured ON erhous(featured) WHERE featured = TRUE;
        CREATE INDEX IF NOT EXISTS idx_erhous_picked ON erhous(picked) WHERE picked = TRUE;
        CREATE INDEX IF NOT EXISTS idx_erhous_verified ON erhous(verified) WHERE verified = TRUE;

        -- 字段注释
        COMMENT ON COLUMN erhous.sku_enabled IS '是否启用多规格 SKU';
        COMMENT ON COLUMN erhous.is_auction IS '是否为拍卖商品';
        COMMENT ON COLUMN erhous.auction_start_time IS '拍卖开始时间';
        COMMENT ON COLUMN erhous.auction_end_time IS '拍卖截拍时间';
        COMMENT ON COLUMN erhous.bargain_type IS '议价类型：none/small/big/fixed';
        COMMENT ON COLUMN erhous.delivery_fee IS '运费';
        COMMENT ON COLUMN erhous.free_shipping IS '是否包邮';
        COMMENT ON COLUMN erhous.logistics_template_id IS '物流模板ID';
        COMMENT ON COLUMN erhous.escrow_enabled IS '是否启用担保交易';
        COMMENT ON COLUMN erhous.installment_enabled IS '是否支持分期付款';
        COMMENT ON COLUMN erhous.visibility IS '隐私可见性：public/friends/city_only/followers_only';
        COMMENT ON COLUMN erhous.allow_comment IS '是否允许留言';
        COMMENT ON COLUMN erhous.allow_share IS '是否允许转发';
        COMMENT ON COLUMN erhous.sell_reason IS '出手原因：换新/搬家/缺钱/不喜欢';
        COMMENT ON COLUMN erhous.item_source IS '物品来源：personal/merchant/overseas/taobao';
        COMMENT ON COLUMN erhous.certification_type IS '鉴定类型：official/third_party/none';
        COMMENT ON COLUMN erhous.certification_no IS '鉴定证书编号';
        COMMENT ON COLUMN erhous.has_original_box IS '是否有原包装';
        COMMENT ON COLUMN erhous.has_accessories IS '是否有配件';
        COMMENT ON COLUMN erhous.has_warranty_card IS '是否有保修卡';
        COMMENT ON COLUMN erhous.has_invoice IS '是否有发票';
        COMMENT ON COLUMN erhous.use_duration IS '使用时长：1个月/3个月/半年/1年/2年+';
        COMMENT ON COLUMN erhous.repair_history IS '维修历史';
        COMMENT ON COLUMN erhous.warranty_expire IS '保修到期日期';
        COMMENT ON COLUMN erhous.estimated_value IS '估值参考价';
        COMMENT ON COLUMN erhous.promotion_level IS '推广等级 0-10';
        COMMENT ON COLUMN erhous.traffic_weight IS '流量权重 0.00-9.99';
        COMMENT ON COLUMN erhous.content_hash IS '图文指纹（MD5/SHA256）';
        COMMENT ON COLUMN erhous.risk_score IS '风险评分 0-100，<30 限制交易';
        COMMENT ON COLUMN erhous.same_item_id IS '同款识别 ID';
        COMMENT ON COLUMN erhous.shop_id IS '关联店铺ID（个人闲置为 0）';
        COMMENT ON COLUMN erhous.brand_id IS '关联品牌库ID';
        COMMENT ON COLUMN erhous.model_id IS '关联型号库ID';
        COMMENT ON COLUMN erhous.video_url IS '视频 URL';
        COMMENT ON COLUMN erhous.video_cover IS '视频封面';
        COMMENT ON COLUMN erhous.panorama_url IS '360° 全景图 URL';
        COMMENT ON COLUMN erhous.trade_locations IS '多个 POI 交易地点 JSON';
        COMMENT ON COLUMN erhous.pickup_time_slots IS '自提时间段 JSON';
        COMMENT ON COLUMN erhous.tags IS '标签数组（最多5个）';
        COMMENT ON COLUMN erhous.featured IS '精选推荐';
        COMMENT ON COLUMN erhous.picked IS '运营甄选';
        COMMENT ON COLUMN erhous.verified IS '官方验真';
    END IF;
END $$;

-- ============================================================
-- 第二部分：19 张子表
-- ============================================================

-- ------------------------------------------------------------
-- 1. ers_skus 商品 SKU 规格子表
--    对标闲鱼/转转：颜色×尺寸×版本组合 + 独立库存 + 独立价格
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ers_skus (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    ershou_id BIGINT NOT NULL,
    sku_code VARCHAR(64) NOT NULL,
    name VARCHAR(200) NOT NULL DEFAULT '',
    color VARCHAR(50) NOT NULL DEFAULT '',
    size VARCHAR(50) NOT NULL DEFAULT '',
    version VARCHAR(100) NOT NULL DEFAULT '',
    price DECIMAL(12,2) NOT NULL DEFAULT 0,
    stock INT NOT NULL DEFAULT 0,
    sold_count INT NOT NULL DEFAULT 0,
    image VARCHAR(255) NOT NULL DEFAULT '',
    weight DECIMAL(8,2) NOT NULL DEFAULT 0,
    barcode VARCHAR(64) NOT NULL DEFAULT '',
    status INT NOT NULL DEFAULT 1,
    attributes JSONB,
    sort INT NOT NULL DEFAULT 0,
    CONSTRAINT uniq_sku_ershou_code UNIQUE (ershou_id, sku_code)
);
CREATE INDEX IF NOT EXISTS idx_ers_skus_ershou_id ON ers_skus(ershou_id);
CREATE INDEX IF NOT EXISTS idx_ers_skus_barcode ON ers_skus(barcode);
CREATE INDEX IF NOT EXISTS idx_ers_skus_status ON ers_skus(status);
CREATE INDEX IF NOT EXISTS idx_ers_skus_deleted_at ON ers_skus(deleted_at);
COMMENT ON TABLE ers_skus IS '商品 SKU 规格子表（颜色×尺寸×版本组合 + 独立库存 + 独立价格）';

-- ------------------------------------------------------------
-- 2. ers_orders 订单主表
--    对标闲鱼担保交易：11 状态机
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ers_orders (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    order_no VARCHAR(64) NOT NULL,
    buyer_id BIGINT NOT NULL,
    seller_id BIGINT NOT NULL,
    shop_id BIGINT NOT NULL DEFAULT 0,
    total_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    item_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    delivery_fee DECIMAL(12,2) NOT NULL DEFAULT 0,
    discount_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 0,
    pay_method VARCHAR(32) NOT NULL DEFAULT 'wechat',
    pay_trade_no VARCHAR(128) NOT NULL DEFAULT '',
    delivery_method VARCHAR(32) NOT NULL DEFAULT 'face',
    remark VARCHAR(500) NOT NULL DEFAULT '',
    contact_name VARCHAR(50) NOT NULL DEFAULT '',
    contact_phone VARCHAR(20) NOT NULL DEFAULT '',
    contact_address VARCHAR(500) NOT NULL DEFAULT '',
    escrow_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    installment_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    installment_periods INT NOT NULL DEFAULT 0,
    paid_at TIMESTAMPTZ,
    shipped_at TIMESTAMPTZ,
    received_at TIMESTAMPTZ,
    settled_at TIMESTAMPTZ,
    closed_at TIMESTAMPTZ,
    auto_close_at TIMESTAMPTZ,
    auto_receive_at TIMESTAMPTZ,
    CONSTRAINT uniq_orders_order_no UNIQUE (order_no)
);
CREATE INDEX IF NOT EXISTS idx_ers_orders_region_id ON ers_orders(region_id);
CREATE INDEX IF NOT EXISTS idx_ers_orders_buyer ON ers_orders(buyer_id, status);
CREATE INDEX IF NOT EXISTS idx_ers_orders_seller ON ers_orders(seller_id, status);
CREATE INDEX IF NOT EXISTS idx_ers_orders_shop_id ON ers_orders(shop_id);
CREATE INDEX IF NOT EXISTS idx_ers_orders_status ON ers_orders(status);
CREATE INDEX IF NOT EXISTS idx_ers_orders_pay_trade_no ON ers_orders(pay_trade_no);
CREATE INDEX IF NOT EXISTS idx_ers_orders_paid_at ON ers_orders(paid_at);
CREATE INDEX IF NOT EXISTS idx_ers_orders_shipped_at ON ers_orders(shipped_at);
CREATE INDEX IF NOT EXISTS idx_ers_orders_received_at ON ers_orders(received_at);
CREATE INDEX IF NOT EXISTS idx_ers_orders_settled_at ON ers_orders(settled_at);
CREATE INDEX IF NOT EXISTS idx_ers_orders_closed_at ON ers_orders(closed_at);
CREATE INDEX IF NOT EXISTS idx_ers_orders_auto_close_at ON ers_orders(auto_close_at);
CREATE INDEX IF NOT EXISTS idx_ers_orders_auto_receive_at ON ers_orders(auto_receive_at);
CREATE INDEX IF NOT EXISTS idx_ers_orders_deleted_at ON ers_orders(deleted_at);
COMMENT ON TABLE ers_orders IS '订单主表（11状态机：0待支付/1已支付待发货/2已发货/3待收货/4已完成/5已取消/6退款中/7退款完成/8申诉中/9申诉完成/10已关闭）';

-- ------------------------------------------------------------
-- 3. ers_order_items 订单子项表
--    一个订单可包含多个商品子项（多件合并下单）
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ers_order_items (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    order_id BIGINT NOT NULL,
    ershou_id BIGINT NOT NULL,
    sku_id BIGINT NOT NULL DEFAULT 0,
    sku_code VARCHAR(64) NOT NULL DEFAULT '',
    title VARCHAR(200) NOT NULL DEFAULT '',
    cover_image VARCHAR(255) NOT NULL DEFAULT '',
    quantity INT NOT NULL DEFAULT 1,
    unit_price DECIMAL(12,2) NOT NULL DEFAULT 0,
    subtotal DECIMAL(12,2) NOT NULL DEFAULT 0,
    remark VARCHAR(500) NOT NULL DEFAULT '',
    CONSTRAINT uniq_order_item UNIQUE (order_id, quantity)
);
CREATE INDEX IF NOT EXISTS idx_ers_order_items_order_id ON ers_order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_ers_order_items_ershou_id ON ers_order_items(ershou_id);
CREATE INDEX IF NOT EXISTS idx_ers_order_items_sku_id ON ers_order_items(sku_id);
CREATE INDEX IF NOT EXISTS idx_ers_order_items_deleted_at ON ers_order_items(deleted_at);
-- 外键：依赖 ers_orders 已创建
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_order_items_order') THEN
        ALTER TABLE ers_order_items ADD CONSTRAINT fk_order_items_order
            FOREIGN KEY (order_id) REFERENCES ers_orders(id) ON DELETE CASCADE;
    END IF;
END $$;
COMMENT ON TABLE ers_order_items IS '订单子项（订单ID/物品ID/SKU ID/数量/单价/小计）';

-- ------------------------------------------------------------
-- 4. ers_auctions 拍卖信息表
--    对标闲鱼拍卖：起拍价/加价幅度/保留价/截拍时间
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ers_auctions (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    ershou_id BIGINT NOT NULL,
    start_price DECIMAL(12,2) NOT NULL DEFAULT 0,
    step_price DECIMAL(12,2) NOT NULL DEFAULT 1,
    reserve_price DECIMAL(12,2) NOT NULL DEFAULT 0,
    bond_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    current_bid_price DECIMAL(12,2) NOT NULL DEFAULT 0,
    current_bid_user_id BIGINT NOT NULL DEFAULT 0,
    bid_count INT NOT NULL DEFAULT 0,
    watcher_count INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 0,
    start_time TIMESTAMPTZ,
    end_time TIMESTAMPTZ,
    auto_extend_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    winner_id BIGINT NOT NULL DEFAULT 0,
    winner_price DECIMAL(12,2) NOT NULL DEFAULT 0,
    closed_at TIMESTAMPTZ,
    CONSTRAINT uniq_auctions_ershou_id UNIQUE (ershou_id)
);
CREATE INDEX IF NOT EXISTS idx_ers_auctions_region_id ON ers_auctions(region_id);
CREATE INDEX IF NOT EXISTS idx_ers_auctions_current_bid_price ON ers_auctions(current_bid_price);
CREATE INDEX IF NOT EXISTS idx_ers_auctions_current_bid_user_id ON ers_auctions(current_bid_user_id);
CREATE INDEX IF NOT EXISTS idx_ers_auctions_status ON ers_auctions(status);
CREATE INDEX IF NOT EXISTS idx_ers_auctions_start_time ON ers_auctions(start_time);
CREATE INDEX IF NOT EXISTS idx_ers_auctions_end_time ON ers_auctions(end_time);
CREATE INDEX IF NOT EXISTS idx_ers_auctions_winner_id ON ers_auctions(winner_id);
CREATE INDEX IF NOT EXISTS idx_ers_auctions_closed_at ON ers_auctions(closed_at);
CREATE INDEX IF NOT EXISTS idx_ers_auctions_deleted_at ON ers_auctions(deleted_at);
COMMENT ON TABLE ers_auctions IS '拍卖信息表（起拍价/加价幅度/保留价/截拍时间/当前最高价/状态）';

-- ------------------------------------------------------------
-- 5. ers_auction_bids 拍卖出价记录表
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ers_auction_bids (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    auction_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    bid_price DECIMAL(12,2) NOT NULL,
    bid_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_winner BOOLEAN NOT NULL DEFAULT FALSE,
    ip VARCHAR(64) NOT NULL DEFAULT '',
    user_agent VARCHAR(255) NOT NULL DEFAULT '',
    is_invalid BOOLEAN NOT NULL DEFAULT FALSE,
    invalid_reason VARCHAR(500) NOT NULL DEFAULT '',
    CONSTRAINT uniq_auction_bid UNIQUE (auction_id, bid_price)
);
CREATE INDEX IF NOT EXISTS idx_ers_auction_bids_auction ON ers_auction_bids(auction_id);
CREATE INDEX IF NOT EXISTS idx_ers_auction_bids_user_id ON ers_auction_bids(user_id);
CREATE INDEX IF NOT EXISTS idx_ers_auction_bids_is_winner ON ers_auction_bids(is_winner) WHERE is_winner = TRUE;
CREATE INDEX IF NOT EXISTS idx_ers_auction_bids_is_invalid ON ers_auction_bids(is_invalid);
CREATE INDEX IF NOT EXISTS idx_ers_auction_bids_deleted_at ON ers_auction_bids(deleted_at);
-- 外键：依赖 ers_auctions 已创建
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_auction_bids_auction') THEN
        ALTER TABLE ers_auction_bids ADD CONSTRAINT fk_auction_bids_auction
            FOREIGN KEY (auction_id) REFERENCES ers_auctions(id) ON DELETE CASCADE;
    END IF;
END $$;
COMMENT ON TABLE ers_auction_bids IS '拍卖出价记录（拍卖ID/用户ID/出价金额/出价时间）';

-- ------------------------------------------------------------
-- 6. ers_promotions 付费推广记录表
--    对标闲鱼：首页轮播/频道置顶/搜索置顶/加急
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ers_promotions (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    ershou_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    promotion_type VARCHAR(32) NOT NULL,
    status INT NOT NULL DEFAULT 0,
    start_time TIMESTAMPTZ,
    end_time TIMESTAMPTZ,
    duration_days INT NOT NULL DEFAULT 1,
    amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    pay_method VARCHAR(32) NOT NULL DEFAULT '',
    pay_trade_no VARCHAR(128) NOT NULL DEFAULT '',
    paid_at TIMESTAMPTZ,
    impression_count INT NOT NULL DEFAULT 0,
    click_count INT NOT NULL DEFAULT 0,
    fav_count INT NOT NULL DEFAULT 0,
    consult_count INT NOT NULL DEFAULT 0,
    order_count INT NOT NULL DEFAULT 0,
    roi DECIMAL(5,2) NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_ers_promotions_region_id ON ers_promotions(region_id);
CREATE INDEX IF NOT EXISTS idx_ers_promotions_ershou_status ON ers_promotions(ershou_id, status);
CREATE INDEX IF NOT EXISTS idx_ers_promotions_user_id ON ers_promotions(user_id);
CREATE INDEX IF NOT EXISTS idx_ers_promotions_promotion_type ON ers_promotions(promotion_type);
CREATE INDEX IF NOT EXISTS idx_ers_promotions_status ON ers_promotions(status);
CREATE INDEX IF NOT EXISTS idx_ers_promotions_start_time ON ers_promotions(start_time);
CREATE INDEX IF NOT EXISTS idx_ers_promotions_end_time ON ers_promotions(end_time);
CREATE INDEX IF NOT EXISTS idx_ers_promotions_pay_trade_no ON ers_promotions(pay_trade_no);
CREATE INDEX IF NOT EXISTS idx_ers_promotions_paid_at ON ers_promotions(paid_at);
CREATE INDEX IF NOT EXISTS idx_ers_promotions_deleted_at ON ers_promotions(deleted_at);
COMMENT ON TABLE ers_promotions IS '付费推广记录（物品ID/推广类型/起止时间/花费/曝光/点击/ROI）';

-- ------------------------------------------------------------
-- 7. ers_reports 举报工单表
--    对标转转：色情/诈骗/虚假/违禁/侵权 + SLA 24h响应/72h处理
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ers_reports (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    report_no VARCHAR(64) NOT NULL,
    ershou_id BIGINT NOT NULL,
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
    CONSTRAINT uniq_reports_report_no UNIQUE (report_no)
);
CREATE INDEX IF NOT EXISTS idx_ers_reports_ershou_id ON ers_reports(ershou_id);
CREATE INDEX IF NOT EXISTS idx_ers_reports_reporter_id ON ers_reports(reporter_id);
CREATE INDEX IF NOT EXISTS idx_ers_reports_reported_user_id ON ers_reports(reported_user_id);
CREATE INDEX IF NOT EXISTS idx_ers_reports_report_type ON ers_reports(report_type);
CREATE INDEX IF NOT EXISTS idx_ers_reports_status ON ers_reports(status);
CREATE INDEX IF NOT EXISTS idx_ers_reports_handler_id ON ers_reports(handler_id);
CREATE INDEX IF NOT EXISTS idx_ers_reports_penalty_user_id ON ers_reports(penalty_user_id);
CREATE INDEX IF NOT EXISTS idx_ers_reports_sla_deadline ON ers_reports(sla_deadline);
CREATE INDEX IF NOT EXISTS idx_ers_reports_handled_at ON ers_reports(handled_at);
CREATE INDEX IF NOT EXISTS idx_ers_reports_appealed_at ON ers_reports(appealed_at);
CREATE INDEX IF NOT EXISTS idx_ers_reports_deleted_at ON ers_reports(deleted_at);
COMMENT ON TABLE ers_reports IS '举报工单（物品ID/举报人/被举报人/类型/原因/证据/状态/处理结果/SLA）';

-- ------------------------------------------------------------
-- 8. ers_reviews 交易评价表
--    对标转转：5 星 + 文字 + 图片 + 追评 + 回复
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ers_reviews (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    order_id BIGINT NOT NULL,
    ershou_id BIGINT NOT NULL,
    reviewer_id BIGINT NOT NULL,
    reviewer_name VARCHAR(50) NOT NULL DEFAULT '',
    reviewer_avatar VARCHAR(255) NOT NULL DEFAULT '',
    reviewee_id BIGINT NOT NULL,
    review_type VARCHAR(16) NOT NULL DEFAULT 'buyer_to_seller',
    rating INT NOT NULL DEFAULT 5,
    content TEXT NOT NULL DEFAULT '',
    images JSONB,
    video_url VARCHAR(255) NOT NULL DEFAULT '',
    is_anonymous BOOLEAN NOT NULL DEFAULT FALSE,
    is_recommended BOOLEAN NOT NULL DEFAULT TRUE,
    tags JSONB,
    reply TEXT NOT NULL DEFAULT '',
    reply_at TIMESTAMPTZ,
    append_content TEXT NOT NULL DEFAULT '',
    append_images JSONB,
    append_at TIMESTAMPTZ,
    like_count INT NOT NULL DEFAULT 0,
    CONSTRAINT uniq_review_order_user UNIQUE (order_id, reviewer_id)
);
CREATE INDEX IF NOT EXISTS idx_ers_reviews_ershou_id ON ers_reviews(ershou_id);
CREATE INDEX IF NOT EXISTS idx_ers_reviews_reviewer_id ON ers_reviews(reviewer_id);
CREATE INDEX IF NOT EXISTS idx_ers_reviews_reviewee_id ON ers_reviews(reviewee_id);
CREATE INDEX IF NOT EXISTS idx_ers_reviews_rating ON ers_reviews(rating);
CREATE INDEX IF NOT EXISTS idx_ers_reviews_is_recommended ON ers_reviews(is_recommended) WHERE is_recommended = FALSE;
CREATE INDEX IF NOT EXISTS idx_ers_reviews_reply_at ON ers_reviews(reply_at);
CREATE INDEX IF NOT EXISTS idx_ers_reviews_append_at ON ers_reviews(append_at);
CREATE INDEX IF NOT EXISTS idx_ers_reviews_deleted_at ON ers_reviews(deleted_at);
-- 外键：依赖 ers_orders 已创建
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_reviews_order') THEN
        ALTER TABLE ers_reviews ADD CONSTRAINT fk_reviews_order
            FOREIGN KEY (order_id) REFERENCES ers_orders(id) ON DELETE CASCADE;
    END IF;
END $$;
COMMENT ON TABLE ers_reviews IS '交易评价（订单ID/评价人/被评价人/评分/内容/图片/追评/回复）';

-- ------------------------------------------------------------
-- 9. ers_shops 商家店铺表
--    对标转转商家版：商家入驻/装修/Banner/粉丝
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ers_shops (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id BIGINT NOT NULL,
    shop_name VARCHAR(128) NOT NULL,
    logo VARCHAR(255) NOT NULL DEFAULT '',
    banner VARCHAR(255) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    level INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 0,
    contact_name VARCHAR(50) NOT NULL DEFAULT '',
    contact_phone VARCHAR(20) NOT NULL DEFAULT '',
    contact_wechat VARCHAR(50) NOT NULL DEFAULT '',
    address VARCHAR(500) NOT NULL DEFAULT '',
    latitude DECIMAL(10,7) NOT NULL DEFAULT 0,
    longitude DECIMAL(10,7) NOT NULL DEFAULT 0,
    business_license VARCHAR(255) NOT NULL DEFAULT '',
    license_no VARCHAR(64) NOT NULL DEFAULT '',
    id_card_front VARCHAR(255) NOT NULL DEFAULT '',
    id_card_back VARCHAR(255) NOT NULL DEFAULT '',
    verified_at TIMESTAMPTZ,
    follower_count INT NOT NULL DEFAULT 0,
    item_count INT NOT NULL DEFAULT 0,
    sold_count INT NOT NULL DEFAULT 0,
    total_amount DECIMAL(14,2) NOT NULL DEFAULT 0,
    good_rate DECIMAL(5,2) NOT NULL DEFAULT 100.00,
    deposit DECIMAL(12,2) NOT NULL DEFAULT 0,
    tags JSONB,
    approved_at TIMESTAMPTZ,
    rejected_reason VARCHAR(500) NOT NULL DEFAULT '',
    closed_at TIMESTAMPTZ,
    CONSTRAINT uniq_shops_user_id UNIQUE (user_id)
);
CREATE INDEX IF NOT EXISTS idx_ers_shops_region_id ON ers_shops(region_id);
CREATE INDEX IF NOT EXISTS idx_ers_shops_shop_name ON ers_shops(shop_name);
CREATE INDEX IF NOT EXISTS idx_ers_shops_level ON ers_shops(level);
CREATE INDEX IF NOT EXISTS idx_ers_shops_status ON ers_shops(status);
CREATE INDEX IF NOT EXISTS idx_ers_shops_license_no ON ers_shops(license_no);
CREATE INDEX IF NOT EXISTS idx_ers_shops_verified_at ON ers_shops(verified_at);
CREATE INDEX IF NOT EXISTS idx_ers_shops_approved_at ON ers_shops(approved_at);
CREATE INDEX IF NOT EXISTS idx_ers_shops_closed_at ON ers_shops(closed_at);
CREATE INDEX IF NOT EXISTS idx_ers_shops_deleted_at ON ers_shops(deleted_at);
COMMENT ON TABLE ers_shops IS '商家店铺（商家ID/店铺名/Logo/Banner/简介/等级/状态/粉丝数/商品数）';

-- ------------------------------------------------------------
-- 10. ers_shop_followers 店铺粉丝关联表
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ers_shop_followers (
    id BIGSERIAL PRIMARY KEY,
    shop_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    notify BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uniq_shop_follower UNIQUE (shop_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_ers_shop_followers_shop_id ON ers_shop_followers(shop_id);
CREATE INDEX IF NOT EXISTS idx_ers_shop_followers_user_id ON ers_shop_followers(user_id);
-- 外键：依赖 ers_shops 已创建
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_shop_followers_shop') THEN
        ALTER TABLE ers_shop_followers ADD CONSTRAINT fk_shop_followers_shop
            FOREIGN KEY (shop_id) REFERENCES ers_shops(id) ON DELETE CASCADE;
    END IF;
END $$;
COMMENT ON TABLE ers_shop_followers IS '店铺粉丝关联（店铺ID/用户ID/关注时间）';

-- ------------------------------------------------------------
-- 11. ers_logistics 物流记录表
--    对标转转：快递公司/运单号/状态/跟踪信息JSON
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ers_logistics (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    order_id BIGINT NOT NULL,
    ershou_id BIGINT NOT NULL,
    sku_id BIGINT NOT NULL DEFAULT 0,
    express_company VARCHAR(50) NOT NULL DEFAULT '',
    express_code VARCHAR(32) NOT NULL DEFAULT '',
    tracking_no VARCHAR(64) NOT NULL,
    status INT NOT NULL DEFAULT 0,
    shipper_name VARCHAR(50) NOT NULL DEFAULT '',
    shipper_phone VARCHAR(20) NOT NULL DEFAULT '',
    shipper_address VARCHAR(500) NOT NULL DEFAULT '',
    receiver_name VARCHAR(50) NOT NULL DEFAULT '',
    receiver_phone VARCHAR(20) NOT NULL DEFAULT '',
    receiver_address VARCHAR(500) NOT NULL DEFAULT '',
    weight DECIMAL(8,2) NOT NULL DEFAULT 0,
    freight DECIMAL(12,2) NOT NULL DEFAULT 0,
    tracking_info JSONB,
    shipped_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    remark VARCHAR(500) NOT NULL DEFAULT '',
    CONSTRAINT uniq_logistics_order_id UNIQUE (order_id)
);
CREATE INDEX IF NOT EXISTS idx_ers_logistics_ershou_id ON ers_logistics(ershou_id);
CREATE INDEX IF NOT EXISTS idx_ers_logistics_sku_id ON ers_logistics(sku_id);
CREATE INDEX IF NOT EXISTS idx_ers_logistics_tracking_no ON ers_logistics(tracking_no);
CREATE INDEX IF NOT EXISTS idx_ers_logistics_status ON ers_logistics(status);
CREATE INDEX IF NOT EXISTS idx_ers_logistics_shipped_at ON ers_logistics(shipped_at);
CREATE INDEX IF NOT EXISTS idx_ers_logistics_delivered_at ON ers_logistics(delivered_at);
CREATE INDEX IF NOT EXISTS idx_ers_logistics_deleted_at ON ers_logistics(deleted_at);
-- 外键：依赖 ers_orders 已创建
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_logistics_order') THEN
        ALTER TABLE ers_logistics ADD CONSTRAINT fk_logistics_order
            FOREIGN KEY (order_id) REFERENCES ers_orders(id) ON DELETE CASCADE;
    END IF;
END $$;
COMMENT ON TABLE ers_logistics IS '物流记录（订单ID/快递公司/运单号/状态/跟踪信息JSON）';

-- ------------------------------------------------------------
-- 12. ers_escrows 担保交易表
--    对标闲鱼/瓜子：资金托管/冻结/解冻/放款
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ers_escrows (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    order_id BIGINT NOT NULL,
    ershou_id BIGINT NOT NULL,
    buyer_id BIGINT NOT NULL,
    seller_id BIGINT NOT NULL,
    escrow_amount DECIMAL(12,2) NOT NULL,
    platform_fee DECIMAL(12,2) NOT NULL DEFAULT 0,
    seller_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 1,
    frozen_at TIMESTAMPTZ,
    release_at TIMESTAMPTZ,
    paid_at TIMESTAMPTZ,
    refunded_at TIMESTAMPTZ,
    auto_release_at TIMESTAMPTZ,
    dispute_reason VARCHAR(500) NOT NULL DEFAULT '',
    arbitration_result TEXT NOT NULL DEFAULT '',
    CONSTRAINT uniq_escrows_order_id UNIQUE (order_id)
);
CREATE INDEX IF NOT EXISTS idx_ers_escrows_ershou_id ON ers_escrows(ershou_id);
CREATE INDEX IF NOT EXISTS idx_ers_escrows_buyer_id ON ers_escrows(buyer_id);
CREATE INDEX IF NOT EXISTS idx_ers_escrows_seller_id ON ers_escrows(seller_id);
CREATE INDEX IF NOT EXISTS idx_ers_escrows_status ON ers_escrows(status);
CREATE INDEX IF NOT EXISTS idx_ers_escrows_frozen_at ON ers_escrows(frozen_at);
CREATE INDEX IF NOT EXISTS idx_ers_escrows_release_at ON ers_escrows(release_at);
CREATE INDEX IF NOT EXISTS idx_ers_escrows_paid_at ON ers_escrows(paid_at);
CREATE INDEX IF NOT EXISTS idx_ers_escrows_refunded_at ON ers_escrows(refunded_at);
CREATE INDEX IF NOT EXISTS idx_ers_escrows_auto_release_at ON ers_escrows(auto_release_at);
CREATE INDEX IF NOT EXISTS idx_ers_escrows_deleted_at ON ers_escrows(deleted_at);
-- 外键：依赖 ers_orders 已创建
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_escrows_order') THEN
        ALTER TABLE ers_escrows ADD CONSTRAINT fk_escrows_order
            FOREIGN KEY (order_id) REFERENCES ers_orders(id) ON DELETE CASCADE;
    END IF;
END $$;
COMMENT ON TABLE ers_escrows IS '担保交易（订单ID/托管金额/状态/冻结时间/解冻时间/放款时间）';

-- ------------------------------------------------------------
-- 13. ers_refunds 退款工单表
--    对标转转：退货/换货/维修/纠纷
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ers_refunds (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    refund_no VARCHAR(64) NOT NULL,
    order_id BIGINT NOT NULL,
    ershou_id BIGINT NOT NULL,
    buyer_id BIGINT NOT NULL,
    seller_id BIGINT NOT NULL,
    refund_type VARCHAR(32) NOT NULL,
    reason VARCHAR(500) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    evidence_images JSONB,
    refund_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 0,
    seller_reason VARCHAR(500) NOT NULL DEFAULT '',
    arbitration_result TEXT NOT NULL DEFAULT '',
    arbitrator_id BIGINT NOT NULL DEFAULT 0,
    arbitrated_at TIMESTAMPTZ,
    sla_deadline TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    CONSTRAINT uniq_refunds_refund_no UNIQUE (refund_no)
);
CREATE INDEX IF NOT EXISTS idx_ers_refunds_order_id ON ers_refunds(order_id);
CREATE INDEX IF NOT EXISTS idx_ers_refunds_ershou_id ON ers_refunds(ershou_id);
CREATE INDEX IF NOT EXISTS idx_ers_refunds_buyer_id ON ers_refunds(buyer_id);
CREATE INDEX IF NOT EXISTS idx_ers_refunds_seller_id ON ers_refunds(seller_id);
CREATE INDEX IF NOT EXISTS idx_ers_refunds_refund_type ON ers_refunds(refund_type);
CREATE INDEX IF NOT EXISTS idx_ers_refunds_status ON ers_refunds(status);
CREATE INDEX IF NOT EXISTS idx_ers_refunds_arbitrator_id ON ers_refunds(arbitrator_id);
CREATE INDEX IF NOT EXISTS idx_ers_refunds_arbitrated_at ON ers_refunds(arbitrated_at);
CREATE INDEX IF NOT EXISTS idx_ers_refunds_sla_deadline ON ers_refunds(sla_deadline);
CREATE INDEX IF NOT EXISTS idx_ers_refunds_completed_at ON ers_refunds(completed_at);
CREATE INDEX IF NOT EXISTS idx_ers_refunds_deleted_at ON ers_refunds(deleted_at);
-- 外键：依赖 ers_orders 已创建
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_refunds_order') THEN
        ALTER TABLE ers_refunds ADD CONSTRAINT fk_refunds_order
            FOREIGN KEY (order_id) REFERENCES ers_orders(id) ON DELETE CASCADE;
    END IF;
END $$;
COMMENT ON TABLE ers_refunds IS '退款工单（订单ID/类型/原因/金额/状态/凭证/仲裁结果）';

-- ------------------------------------------------------------
-- 14. ers_tags 标签库表
--    对标转转：智能/运营/自定义 + 颜色 + 排序 + 使用次数
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ers_tags (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    name VARCHAR(64) NOT NULL,
    type VARCHAR(16) NOT NULL,
    color VARCHAR(16) NOT NULL DEFAULT '#409EFF',
    icon VARCHAR(64) NOT NULL DEFAULT '',
    background VARCHAR(32) NOT NULL DEFAULT '',
    status INT NOT NULL DEFAULT 1,
    sort INT NOT NULL DEFAULT 0,
    use_count INT NOT NULL DEFAULT 0,
    is_hot BOOLEAN NOT NULL DEFAULT FALSE,
    creator_id BIGINT NOT NULL DEFAULT 0,
    CONSTRAINT uniq_tag_name_type UNIQUE (name, type)
);
CREATE INDEX IF NOT EXISTS idx_ers_tags_type ON ers_tags(type);
CREATE INDEX IF NOT EXISTS idx_ers_tags_status ON ers_tags(status);
CREATE INDEX IF NOT EXISTS idx_ers_tags_is_hot ON ers_tags(is_hot) WHERE is_hot = TRUE;
CREATE INDEX IF NOT EXISTS idx_ers_tags_creator_id ON ers_tags(creator_id);
CREATE INDEX IF NOT EXISTS idx_ers_tags_deleted_at ON ers_tags(deleted_at);
COMMENT ON TABLE ers_tags IS '标签库（名称/类型 smart/operation/custom/颜色/排序/使用次数）';

-- ------------------------------------------------------------
-- 15. ers_brands 品牌库表
--    对标转转：名/Logo/官方认证/排序
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ers_brands (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    name VARCHAR(128) NOT NULL,
    logo VARCHAR(255) NOT NULL DEFAULT '',
    english_name VARCHAR(128) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    country VARCHAR(32) NOT NULL DEFAULT '',
    official_verified BOOLEAN NOT NULL DEFAULT FALSE,
    official_url VARCHAR(255) NOT NULL DEFAULT '',
    category_ids JSONB,
    status INT NOT NULL DEFAULT 1,
    sort INT NOT NULL DEFAULT 0,
    use_count INT NOT NULL DEFAULT 0,
    CONSTRAINT uniq_brand_name UNIQUE (name)
);
CREATE INDEX IF NOT EXISTS idx_ers_brands_official_verified ON ers_brands(official_verified) WHERE official_verified = TRUE;
CREATE INDEX IF NOT EXISTS idx_ers_brands_status ON ers_brands(status);
CREATE INDEX IF NOT EXISTS idx_ers_brands_sort ON ers_brands(sort);
CREATE INDEX IF NOT EXISTS idx_ers_brands_deleted_at ON ers_brands(deleted_at);
COMMENT ON TABLE ers_brands IS '品牌库（名称/Logo/官方认证/排序）';

-- ------------------------------------------------------------
-- 16. ers_models 型号库表
--    对标转转：品牌ID/名称/规格参数JSON/排序
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ers_models (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    brand_id BIGINT NOT NULL,
    name VARCHAR(128) NOT NULL,
    full_name VARCHAR(255) NOT NULL DEFAULT '',
    specifications JSONB,
    image VARCHAR(255) NOT NULL DEFAULT '',
    release_date VARCHAR(16) NOT NULL DEFAULT '',
    status INT NOT NULL DEFAULT 1,
    sort INT NOT NULL DEFAULT 0,
    use_count INT NOT NULL DEFAULT 0,
    reference_price DECIMAL(12,2) NOT NULL DEFAULT 0,
    CONSTRAINT uniq_model_name_brand UNIQUE (brand_id, name)
);
CREATE INDEX IF NOT EXISTS idx_ers_models_brand_name ON ers_models(brand_id, name);
CREATE INDEX IF NOT EXISTS idx_ers_models_status ON ers_models(status);
CREATE INDEX IF NOT EXISTS idx_ers_models_sort ON ers_models(sort);
CREATE INDEX IF NOT EXISTS idx_ers_models_deleted_at ON ers_models(deleted_at);
-- 外键：依赖 ers_brands 已创建
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_models_brand') THEN
        ALTER TABLE ers_models ADD CONSTRAINT fk_models_brand
            FOREIGN KEY (brand_id) REFERENCES ers_brands(id) ON DELETE CASCADE;
    END IF;
END $$;
COMMENT ON TABLE ers_models IS '型号库（品牌ID/名称/规格参数JSON/排序）';

-- ------------------------------------------------------------
-- 17. ers_category_attrs 分类属性配置表
--    对标转转：分类ID/属性名/类型/可选值JSON/是否必填
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ers_category_attrs (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    category_id BIGINT NOT NULL,
    attr_name VARCHAR(64) NOT NULL,
    attr_key VARCHAR(64) NOT NULL DEFAULT '',
    attr_type VARCHAR(32) NOT NULL DEFAULT 'string',
    options JSONB,
    unit VARCHAR(32) NOT NULL DEFAULT '',
    is_required BOOLEAN NOT NULL DEFAULT FALSE,
    is_filterable BOOLEAN NOT NULL DEFAULT FALSE,
    is_searchable BOOLEAN NOT NULL DEFAULT FALSE,
    default_value VARCHAR(255) NOT NULL DEFAULT '',
    placeholder VARCHAR(255) NOT NULL DEFAULT '',
    description VARCHAR(500) NOT NULL DEFAULT '',
    status INT NOT NULL DEFAULT 1,
    sort INT NOT NULL DEFAULT 0,
    CONSTRAINT uniq_cat_attr_name UNIQUE (category_id, attr_name)
);
CREATE INDEX IF NOT EXISTS idx_ers_category_attrs_category ON ers_category_attrs(category_id);
CREATE INDEX IF NOT EXISTS idx_ers_category_attrs_is_filterable ON ers_category_attrs(is_filterable) WHERE is_filterable = TRUE;
CREATE INDEX IF NOT EXISTS idx_ers_category_attrs_status ON ers_category_attrs(status);
CREATE INDEX IF NOT EXISTS idx_ers_category_attrs_deleted_at ON ers_category_attrs(deleted_at);
COMMENT ON TABLE ers_category_attrs IS '分类属性配置（分类ID/属性名/类型/可选值JSON/是否必填）';

-- ------------------------------------------------------------
-- 18. ers_audit_rules 审核规则表
--    对标转转：敏感词/价格异常/频率限制/违禁品/内容
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ers_audit_rules (
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
CREATE INDEX IF NOT EXISTS idx_ers_audit_rules_rule_type ON ers_audit_rules(rule_type);
CREATE INDEX IF NOT EXISTS idx_ers_audit_rules_rule_key ON ers_audit_rules(rule_key);
CREATE INDEX IF NOT EXISTS idx_ers_audit_rules_severity ON ers_audit_rules(severity);
CREATE INDEX IF NOT EXISTS idx_ers_audit_rules_status ON ers_audit_rules(status);
CREATE INDEX IF NOT EXISTS idx_ers_audit_rules_deleted_at ON ers_audit_rules(deleted_at);
COMMENT ON TABLE ers_audit_rules IS '审核规则（规则名/类型 sensitive_word/price_check/frequency/违禁品/内容/状态）';

-- ------------------------------------------------------------
-- 19. ers_user_credit 用户信用表
--    对标转转：信用分/等级/历史交易数/好评数/差评数/纠纷数
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ers_user_credit (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id BIGINT NOT NULL,
    credit_score INT NOT NULL DEFAULT 100,
    credit_level INT NOT NULL DEFAULT 0,
    total_transactions INT NOT NULL DEFAULT 0,
    success_transactions INT NOT NULL DEFAULT 0,
    cancel_transactions INT NOT NULL DEFAULT 0,
    good_reviews INT NOT NULL DEFAULT 0,
    medium_reviews INT NOT NULL DEFAULT 0,
    bad_reviews INT NOT NULL DEFAULT 0,
    good_rate DECIMAL(5,2) NOT NULL DEFAULT 100.00,
    disputes INT NOT NULL DEFAULT 0,
    reports INT NOT NULL DEFAULT 0,
    penalties INT NOT NULL DEFAULT 0,
    last_transaction_at TIMESTAMPTZ,
    frozen_reason VARCHAR(500) NOT NULL DEFAULT '',
    frozen_until TIMESTAMPTZ,
    CONSTRAINT uniq_user_credit_user_id UNIQUE (user_id)
);
CREATE INDEX IF NOT EXISTS idx_ers_user_credit_credit_score ON ers_user_credit(credit_score);
CREATE INDEX IF NOT EXISTS idx_ers_user_credit_credit_level ON ers_user_credit(credit_level);
CREATE INDEX IF NOT EXISTS idx_ers_user_credit_last_transaction_at ON ers_user_credit(last_transaction_at);
CREATE INDEX IF NOT EXISTS idx_ers_user_credit_frozen_until ON ers_user_credit(frozen_until);
CREATE INDEX IF NOT EXISTS idx_ers_user_credit_deleted_at ON ers_user_credit(deleted_at);
COMMENT ON TABLE ers_user_credit IS '用户信用（用户ID/信用分/等级/历史交易数/好评数/差评数/纠纷数）';

-- ============================================================
-- 第三部分：为 19 张表挂载 updated_at 触发器
--   参考 001_p0_baseline.sql 中的 update_updated_at_column 函数
--   幂等：先 DROP IF EXISTS 再 CREATE
-- ============================================================
DO $$
DECLARE t TEXT;
BEGIN
    FOR t IN SELECT tablename FROM pg_tables WHERE schemaname='public' AND tablename IN (
        'ers_skus','ers_orders','ers_order_items','ers_auctions','ers_auction_bids',
        'ers_promotions','ers_reports','ers_reviews','ers_shops','ers_shop_followers',
        'ers_logistics','ers_escrows','ers_refunds','ers_tags','ers_brands',
        'ers_models','ers_category_attrs','ers_audit_rules','ers_user_credit'
    )
    LOOP
        EXECUTE format('DROP TRIGGER IF EXISTS trg_%s_updated ON %s', t, t);
        EXECUTE format('CREATE TRIGGER trg_%s_updated BEFORE UPDATE ON %s FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()', t, t);
    END LOOP;
END $$;
