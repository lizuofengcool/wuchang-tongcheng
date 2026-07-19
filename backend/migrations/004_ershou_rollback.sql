-- ============================================================
-- ershou 二手交易模块完整功能回滚脚本（与 003_ershou_full.sql 配对）
-- 按反向顺序 DROP 19 张子表 + 触发器 + 主表新增字段
-- 幂等：使用 IF EXISTS，可重复执行
-- 警告：执行后 ershou 模块全部扩展数据将丢失（主表基础字段保留）
-- ============================================================

-- ------------------------------------------------------------
-- 1. 先移除 19 张子表的 updated_at 触发器
-- ------------------------------------------------------------
DROP TRIGGER IF EXISTS trg_ers_user_credit_updated ON ers_user_credit;
DROP TRIGGER IF EXISTS trg_ers_audit_rules_updated ON ers_audit_rules;
DROP TRIGGER IF EXISTS trg_ers_category_attrs_updated ON ers_category_attrs;
DROP TRIGGER IF EXISTS trg_ers_models_updated ON ers_models;
DROP TRIGGER IF EXISTS trg_ers_brands_updated ON ers_brands;
DROP TRIGGER IF EXISTS trg_ers_tags_updated ON ers_tags;
DROP TRIGGER IF EXISTS trg_ers_refunds_updated ON ers_refunds;
DROP TRIGGER IF EXISTS trg_ers_escrows_updated ON ers_escrows;
DROP TRIGGER IF EXISTS trg_ers_logistics_updated ON ers_logistics;
DROP TRIGGER IF EXISTS trg_ers_shop_followers_updated ON ers_shop_followers;
DROP TRIGGER IF EXISTS trg_ers_shops_updated ON ers_shops;
DROP TRIGGER IF EXISTS trg_ers_reviews_updated ON ers_reviews;
DROP TRIGGER IF EXISTS trg_ers_reports_updated ON ers_reports;
DROP TRIGGER IF EXISTS trg_ers_promotions_updated ON ers_promotions;
DROP TRIGGER IF EXISTS trg_ers_auction_bids_updated ON ers_auction_bids;
DROP TRIGGER IF EXISTS trg_ers_auctions_updated ON ers_auctions;
DROP TRIGGER IF EXISTS trg_ers_order_items_updated ON ers_order_items;
DROP TRIGGER IF EXISTS trg_ers_orders_updated ON ers_orders;
DROP TRIGGER IF EXISTS trg_ers_skus_updated ON ers_skus;

-- ------------------------------------------------------------
-- 2. 按反向顺序 DROP 19 张子表（先 DROP 有外键依赖的子表）
--    依赖顺序：ers_models → ers_brands；ers_order_items/ers_logistics/ers_escrows/ers_refunds/ers_reviews → ers_orders；ers_auction_bids → ers_auctions；ers_shop_followers → ers_shops
-- ------------------------------------------------------------
DROP TABLE IF EXISTS ers_user_credit;
DROP TABLE IF EXISTS ers_audit_rules;
DROP TABLE IF EXISTS ers_category_attrs;
DROP TABLE IF EXISTS ers_models;
DROP TABLE IF EXISTS ers_brands;
DROP TABLE IF EXISTS ers_tags;
DROP TABLE IF EXISTS ers_refunds;
DROP TABLE IF EXISTS ers_escrows;
DROP TABLE IF EXISTS ers_logistics;
DROP TABLE IF EXISTS ers_shop_followers;
DROP TABLE IF EXISTS ers_shops;
DROP TABLE IF EXISTS ers_reviews;
DROP TABLE IF EXISTS ers_reports;
DROP TABLE IF EXISTS ers_promotions;
DROP TABLE IF EXISTS ers_auction_bids;
DROP TABLE IF EXISTS ers_auctions;
DROP TABLE IF EXISTS ers_order_items;
DROP TABLE IF EXISTS ers_orders;
DROP TABLE IF EXISTS ers_skus;

-- ------------------------------------------------------------
-- 3. 移除 erhous 主表新增字段（按反向顺序）
--    包装在 DO 块中，表不存在时跳过（幂等）
-- ------------------------------------------------------------
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'erhous') THEN
        -- 运营字段
        ALTER TABLE erhous DROP COLUMN IF EXISTS verified;
        ALTER TABLE erhous DROP COLUMN IF EXISTS picked;
        ALTER TABLE erhous DROP COLUMN IF EXISTS featured;

        -- 标签冗余
        ALTER TABLE erhous DROP COLUMN IF EXISTS tags;

        -- 多地交易
        ALTER TABLE erhous DROP COLUMN IF EXISTS pickup_time_slots;
        ALTER TABLE erhous DROP COLUMN IF EXISTS trade_locations;

        -- 360 展示
        ALTER TABLE erhous DROP COLUMN IF EXISTS panorama_url;

        -- 视频支持
        ALTER TABLE erhous DROP COLUMN IF EXISTS video_cover;
        ALTER TABLE erhous DROP COLUMN IF EXISTS video_url;

        -- 关联
        ALTER TABLE erhous DROP COLUMN IF EXISTS model_id;
        ALTER TABLE erhous DROP COLUMN IF EXISTS brand_id;
        ALTER TABLE erhous DROP COLUMN IF EXISTS shop_id;

        -- 风控
        ALTER TABLE erhous DROP COLUMN IF EXISTS same_item_id;
        ALTER TABLE erhous DROP COLUMN IF EXISTS risk_score;
        ALTER TABLE erhous DROP COLUMN IF EXISTS content_hash;

        -- 推广加权
        ALTER TABLE erhous DROP COLUMN IF EXISTS traffic_weight;
        ALTER TABLE erhous DROP COLUMN IF EXISTS promotion_level;

        -- 估值参考
        ALTER TABLE erhous DROP COLUMN IF EXISTS estimated_value;

        -- 使用情况
        ALTER TABLE erhous DROP COLUMN IF EXISTS warranty_expire;
        ALTER TABLE erhous DROP COLUMN IF EXISTS repair_history;
        ALTER TABLE erhous DROP COLUMN IF EXISTS use_duration;

        -- 包装附件
        ALTER TABLE erhous DROP COLUMN IF EXISTS has_invoice;
        ALTER TABLE erhous DROP COLUMN IF EXISTS has_warranty_card;
        ALTER TABLE erhous DROP COLUMN IF EXISTS has_accessories;
        ALTER TABLE erhous DROP COLUMN IF EXISTS has_original_box;

        -- 鉴定认证
        ALTER TABLE erhous DROP COLUMN IF EXISTS certification_no;
        ALTER TABLE erhous DROP COLUMN IF EXISTS certification_type;

        -- 物品来源 + 出手原因
        ALTER TABLE erhous DROP COLUMN IF EXISTS item_source;
        ALTER TABLE erhous DROP COLUMN IF EXISTS sell_reason;

        -- 互动开关
        ALTER TABLE erhous DROP COLUMN IF EXISTS allow_share;
        ALTER TABLE erhous DROP COLUMN IF EXISTS allow_comment;

        -- 隐私设置
        ALTER TABLE erhous DROP COLUMN IF EXISTS visibility;

        -- 分期相关
        ALTER TABLE erhous DROP COLUMN IF EXISTS installment_enabled;

        -- 担保相关
        ALTER TABLE erhous DROP COLUMN IF EXISTS escrow_enabled;

        -- 物流相关
        ALTER TABLE erhous DROP COLUMN IF EXISTS logistics_template_id;
        ALTER TABLE erhous DROP COLUMN IF EXISTS free_shipping;
        ALTER TABLE erhous DROP COLUMN IF EXISTS delivery_fee;

        -- 议价相关
        ALTER TABLE erhous DROP COLUMN IF EXISTS bargain_type;

        -- 拍卖相关
        ALTER TABLE erhous DROP COLUMN IF EXISTS auction_end_time;
        ALTER TABLE erhous DROP COLUMN IF EXISTS auction_start_time;
        ALTER TABLE erhous DROP COLUMN IF EXISTS is_auction;

        -- SKU 相关
        ALTER TABLE erhous DROP COLUMN IF EXISTS sku_enabled;
    END IF;
END $$;

-- ------------------------------------------------------------
-- 4. update_updated_at_column 函数保留（可能被其他业务表复用）
--    如需彻底清理，取消下面注释：
-- DROP FUNCTION IF EXISTS update_updated_at_column();
-- ------------------------------------------------------------
