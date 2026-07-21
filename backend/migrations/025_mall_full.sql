-- =====================================================
-- mall 同城商城模块完整迁移脚本
-- 包含 15 张表 + 索引 + 触发器
-- 依据需求文档：对标淘宝/京东/拼多多/美团
-- 数据库：wuchang_tongcheng
-- =====================================================
--
-- 内容：
--   1. CREATE 15 张子表（mall_ 前缀；mall_malls 主表由 GORM AutoMigrate 创建）
--   2. 索引、触发器、注释
--   3. 全幂等：CREATE TABLE IF NOT EXISTS
-- 依赖：001_p0_baseline.sql（update_updated_at_column 函数已创建）
--
-- 表清单（15 张子表，主表 mall_malls 由 GORM AutoMigrate 创建）：
--   mall_shops / mall_categories / mall_products / mall_skus / mall_cart /
--   mall_orders / mall_order_items / mall_addresses / mall_payments / mall_refunds /
--   mall_logistics / mall_reviews / mall_audit_rules / mall_reports / mall_statistics
--
-- 说明：表名严格遵循 model.TableName() 定义：
--      - Shop → mall_shops（复数）
--      - Category → mall_categories（复数）
--      - Product → mall_products（复数）
--      - Sku → mall_skus（复数）
--      - Cart → mall_cart（单数，model 定义）
--      - Order → mall_orders（复数）
--      - OrderItem → mall_order_items（复数）
--      - Address → mall_addresses（复数）
--      - Payment → mall_payments（复数）
--      - Refund → mall_refunds（复数）
--      - Logistics → mall_logistics（复数）
--      - Review → mall_reviews（复数）
--      - AuditRule → mall_audit_rules（复数，BaseModel 无 region_id）
--      - Report → mall_reports（复数）
--      - Statistic → mall_statistics（复数）
-- =====================================================

-- ============================================================
-- 1. mall_shops 店铺主表
-- ============================================================
CREATE TABLE IF NOT EXISTS mall_shops (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    -- 关联
    user_id BIGINT NOT NULL,

    -- 基本信息
    shop_name VARCHAR(200) NOT NULL DEFAULT '',
    logo VARCHAR(255) NOT NULL DEFAULT '',
    description TEXT,
    shop_type VARCHAR(32) NOT NULL DEFAULT 'personal',  -- personal/enterprise/flagship

    -- 状态
    status INT NOT NULL DEFAULT 0,                -- 0草稿 1已开业 2已关闭 3已冻结 4已过期
    audit_status INT NOT NULL DEFAULT 0,          -- 0待审 1通过 2拒绝
    opened_at TIMESTAMPTZ,

    -- 联系方式
    contact_name VARCHAR(50) NOT NULL DEFAULT '',
    contact_phone VARCHAR(32) NOT NULL DEFAULT '',
    contact_email VARCHAR(128) NOT NULL DEFAULT '',
    wechat VARCHAR(64) NOT NULL DEFAULT '',
    qq VARCHAR(32) NOT NULL DEFAULT '',

    -- 地址
    province VARCHAR(64) NOT NULL DEFAULT '',
    city VARCHAR(64) NOT NULL DEFAULT '',
    district VARCHAR(64) NOT NULL DEFAULT '',
    address VARCHAR(500) NOT NULL DEFAULT '',
    latitude DECIMAL(10,7) NOT NULL DEFAULT 0,
    longitude DECIMAL(10,7) NOT NULL DEFAULT 0,

    -- 资质
    license_no VARCHAR(64) NOT NULL DEFAULT '',
    license_image VARCHAR(255) NOT NULL DEFAULT '',
    legal_person VARCHAR(64) NOT NULL DEFAULT '',
    legal_person_id VARCHAR(32) NOT NULL DEFAULT '',

    -- 统计
    product_count INT NOT NULL DEFAULT 0,
    order_count BIGINT NOT NULL DEFAULT 0,
    sale_amount DECIMAL(14,2) NOT NULL DEFAULT 0,
    rating DECIMAL(3,2) NOT NULL DEFAULT 0,
    review_count INT NOT NULL DEFAULT 0,
    favorite_count INT NOT NULL DEFAULT 0,
    view_count INT NOT NULL DEFAULT 0,

    -- 推广
    featured BOOLEAN NOT NULL DEFAULT FALSE,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    promotion_level INT NOT NULL DEFAULT 0,
    traffic_weight DECIMAL(4,2) NOT NULL DEFAULT 1.00,
    verified_at TIMESTAMPTZ,

    -- JSONB 扩展
    banners JSONB,
    tags JSONB,
    business_hours JSONB,
    facilities JSONB,

    -- 风控
    content_hash VARCHAR(64) NOT NULL DEFAULT '',
    risk_score INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_mall_shops_region_id ON mall_shops(region_id);
CREATE INDEX IF NOT EXISTS idx_mall_shops_user_id ON mall_shops(user_id);
CREATE INDEX IF NOT EXISTS idx_mall_shops_shop_name ON mall_shops(shop_name);
CREATE INDEX IF NOT EXISTS idx_mall_shops_shop_type ON mall_shops(shop_type);
CREATE INDEX IF NOT EXISTS idx_mall_shops_status ON mall_shops(status);
CREATE INDEX IF NOT EXISTS idx_mall_shops_audit_status ON mall_shops(audit_status);
CREATE INDEX IF NOT EXISTS idx_mall_shops_opened_at ON mall_shops(opened_at);
CREATE INDEX IF NOT EXISTS idx_mall_shops_province ON mall_shops(province);
CREATE INDEX IF NOT EXISTS idx_mall_shops_city ON mall_shops(city);
CREATE INDEX IF NOT EXISTS idx_mall_shops_featured ON mall_shops(featured) WHERE featured = TRUE;
CREATE INDEX IF NOT EXISTS idx_mall_shops_verified ON mall_shops(verified) WHERE verified = TRUE;
CREATE INDEX IF NOT EXISTS idx_mall_shops_verified_at ON mall_shops(verified_at);
CREATE INDEX IF NOT EXISTS idx_mall_shops_deleted_at ON mall_shops(deleted_at);

COMMENT ON TABLE mall_shops IS '同城商城店铺主表（商户入驻）';
COMMENT ON COLUMN mall_shops.shop_type IS '店铺类型：personal个人/enterprise企业/flagship旗舰店';
COMMENT ON COLUMN mall_shops.status IS '状态：0草稿 1已开业 2已关闭 3已冻结 4已过期';
COMMENT ON COLUMN mall_shops.audit_status IS '审核状态：0待审 1通过 2拒绝';
COMMENT ON COLUMN mall_shops.banners IS '店铺轮播图 JSONB';
COMMENT ON COLUMN mall_shops.tags IS '店铺标签 JSONB';
COMMENT ON COLUMN mall_shops.business_hours IS '营业时间 JSONB';
COMMENT ON COLUMN mall_shops.facilities IS '店铺设施 JSONB';

-- ============================================================
-- 2. mall_categories 商品分类表（树形结构 parent_id 自引用）
-- ============================================================
CREATE TABLE IF NOT EXISTS mall_categories (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    -- 树形结构
    parent_id BIGINT NOT NULL DEFAULT 0,           -- 父分类 ID（0=根）
    name VARCHAR(64) NOT NULL DEFAULT '',
    icon VARCHAR(255) NOT NULL DEFAULT '',
    cover VARCHAR(255) NOT NULL DEFAULT '',
    level INT NOT NULL DEFAULT 1,                  -- 层级 1/2/3
    path VARCHAR(500) NOT NULL DEFAULT '',         -- 路径如 1,5,12

    -- 显示与排序
    sort INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 1,                 -- 0禁用 1启用
    is_show BOOLEAN NOT NULL DEFAULT TRUE,

    -- 统计
    product_count INT NOT NULL DEFAULT 0,

    -- SEO
    keywords VARCHAR(255) NOT NULL DEFAULT '',
    description VARCHAR(500) NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_mall_categories_region_id ON mall_categories(region_id);
CREATE INDEX IF NOT EXISTS idx_mall_categories_parent_id ON mall_categories(parent_id);
CREATE INDEX IF NOT EXISTS idx_mall_categories_name ON mall_categories(name);
CREATE INDEX IF NOT EXISTS idx_mall_categories_level ON mall_categories(level);
CREATE INDEX IF NOT EXISTS idx_mall_categories_path ON mall_categories(path);
CREATE INDEX IF NOT EXISTS idx_mall_categories_sort ON mall_categories(sort);
CREATE INDEX IF NOT EXISTS idx_mall_categories_status ON mall_categories(status);
CREATE INDEX IF NOT EXISTS idx_mall_categories_is_show ON mall_categories(is_show) WHERE is_show = TRUE;
CREATE INDEX IF NOT EXISTS idx_mall_categories_deleted_at ON mall_categories(deleted_at);

COMMENT ON TABLE mall_categories IS '同城商城商品分类表（树形结构 parent_id 自引用）';
COMMENT ON COLUMN mall_categories.parent_id IS '父分类 ID（0=根分类）';
COMMENT ON COLUMN mall_categories.level IS '层级 1/2/3';
COMMENT ON COLUMN mall_categories.path IS '分类路径，如 1,5,12';
COMMENT ON COLUMN mall_categories.status IS '状态：0禁用 1启用';
COMMENT ON COLUMN mall_categories.is_show IS '是否前台显示';

-- ============================================================
-- 3. mall_products 商品 SPU 主表
-- ============================================================
CREATE TABLE IF NOT EXISTS mall_products (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    -- 基础信息
    shop_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    category_id BIGINT NOT NULL DEFAULT 0,
    brand_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(200) NOT NULL DEFAULT '',
    subtitle VARCHAR(500) NOT NULL DEFAULT '',
    main_image VARCHAR(255) NOT NULL DEFAULT '',
    detail TEXT,
    product_type VARCHAR(32) NOT NULL DEFAULT 'physical',  -- physical/virtual/service

    -- 价格
    price DECIMAL(12,2) NOT NULL DEFAULT 0,
    original_price DECIMAL(12,2) NOT NULL DEFAULT 0,
    cost_price DECIMAL(12,2) NOT NULL DEFAULT 0,
    min_price DECIMAL(12,2) NOT NULL DEFAULT 0,
    max_price DECIMAL(12,2) NOT NULL DEFAULT 0,

    -- 库存与销量
    stock INT NOT NULL DEFAULT 0,
    sales BIGINT NOT NULL DEFAULT 0,
    virtual_sales BIGINT NOT NULL DEFAULT 0,
    stock_warn INT NOT NULL DEFAULT 0,

    -- 状态
    status INT NOT NULL DEFAULT 0,                -- 0草稿 1在售 2下架 3售罄 4回收站
    audit_status INT NOT NULL DEFAULT 0,          -- 0待审 1通过 2拒绝
    audit_reason VARCHAR(500) NOT NULL DEFAULT '',
    published_at TIMESTAMPTZ,

    -- 统计
    view_count INT NOT NULL DEFAULT 0,
    favorite_count INT NOT NULL DEFAULT 0,
    review_count INT NOT NULL DEFAULT 0,
    rating DECIMAL(3,2) NOT NULL DEFAULT 0,
    good_rate DECIMAL(5,2) NOT NULL DEFAULT 0,

    -- 运营字段
    featured BOOLEAN NOT NULL DEFAULT FALSE,
    recommended BOOLEAN NOT NULL DEFAULT FALSE,
    new_arrival BOOLEAN NOT NULL DEFAULT FALSE,
    hot_sale BOOLEAN NOT NULL DEFAULT FALSE,
    promotion_level INT NOT NULL DEFAULT 0,
    traffic_weight DECIMAL(3,2) NOT NULL DEFAULT 1.00,
    sort INT NOT NULL DEFAULT 0,

    -- 物流参数
    free_shipping BOOLEAN NOT NULL DEFAULT FALSE,
    shipping_fee DECIMAL(10,2) NOT NULL DEFAULT 0,
    shipping_template_id BIGINT NOT NULL DEFAULT 0,
    weight DECIMAL(10,3) NOT NULL DEFAULT 0,
    volume DECIMAL(10,3) NOT NULL DEFAULT 0,

    -- JSONB 字段
    images JSONB,
    specs JSONB,
    attributes JSONB,
    tags JSONB,
    sku_specs JSONB,

    -- 风控
    content_hash VARCHAR(64) NOT NULL DEFAULT '',
    risk_score INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_mall_products_region_id ON mall_products(region_id);
CREATE INDEX IF NOT EXISTS idx_mall_products_shop_id ON mall_products(shop_id);
CREATE INDEX IF NOT EXISTS idx_mall_products_user_id ON mall_products(user_id);
CREATE INDEX IF NOT EXISTS idx_mall_products_category_id ON mall_products(category_id);
CREATE INDEX IF NOT EXISTS idx_mall_products_brand_id ON mall_products(brand_id);
CREATE INDEX IF NOT EXISTS idx_mall_products_name ON mall_products(name);
CREATE INDEX IF NOT EXISTS idx_mall_products_product_type ON mall_products(product_type);
CREATE INDEX IF NOT EXISTS idx_mall_products_price ON mall_products(price);
CREATE INDEX IF NOT EXISTS idx_mall_products_sales ON mall_products(sales);
CREATE INDEX IF NOT EXISTS idx_mall_products_status ON mall_products(status);
CREATE INDEX IF NOT EXISTS idx_mall_products_audit_status ON mall_products(audit_status);
CREATE INDEX IF NOT EXISTS idx_mall_products_published_at ON mall_products(published_at);
CREATE INDEX IF NOT EXISTS idx_mall_products_rating ON mall_products(rating);
CREATE INDEX IF NOT EXISTS idx_mall_products_featured ON mall_products(featured) WHERE featured = TRUE;
CREATE INDEX IF NOT EXISTS idx_mall_products_recommended ON mall_products(recommended) WHERE recommended = TRUE;
CREATE INDEX IF NOT EXISTS idx_mall_products_new_arrival ON mall_products(new_arrival) WHERE new_arrival = TRUE;
CREATE INDEX IF NOT EXISTS idx_mall_products_hot_sale ON mall_products(hot_sale) WHERE hot_sale = TRUE;
CREATE INDEX IF NOT EXISTS idx_mall_products_sort ON mall_products(sort);
CREATE INDEX IF NOT EXISTS idx_mall_products_content_hash ON mall_products(content_hash);
CREATE INDEX IF NOT EXISTS idx_mall_products_risk_score ON mall_products(risk_score);
CREATE INDEX IF NOT EXISTS idx_mall_products_deleted_at ON mall_products(deleted_at);

COMMENT ON TABLE mall_products IS '同城商城商品 SPU 主表';
COMMENT ON COLUMN mall_products.product_type IS '商品类型：physical实物/virtual虚拟/service服务';
COMMENT ON COLUMN mall_products.status IS '状态：0草稿 1在售 2下架 3售罄 4回收站';
COMMENT ON COLUMN mall_products.audit_status IS '审核状态：0待审 1通过 2拒绝';
COMMENT ON COLUMN mall_products.images IS '商品图片列表 JSONB';
COMMENT ON COLUMN mall_products.specs IS '规格定义 JSONB（颜色/尺寸等）';
COMMENT ON COLUMN mall_products.attributes IS '商品属性 JSONB（品牌/产地等）';
COMMENT ON COLUMN mall_products.tags IS '商品标签 JSONB';
COMMENT ON COLUMN mall_products.sku_specs IS 'SKU 规格组合预览 JSONB';

-- ============================================================
-- 4. mall_skus 商品 SKU 规格表
-- ============================================================
CREATE TABLE IF NOT EXISTS mall_skus (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    -- 关联
    product_id BIGINT NOT NULL,
    shop_id BIGINT NOT NULL DEFAULT 0,
    user_id BIGINT NOT NULL DEFAULT 0,

    -- 基础信息
    name VARCHAR(200) NOT NULL DEFAULT '',
    sku_code VARCHAR(64) NOT NULL DEFAULT '',
    barcode VARCHAR(64) NOT NULL DEFAULT '',

    -- 规格
    specs JSONB,

    -- 价格
    price DECIMAL(12,2) NOT NULL DEFAULT 0,
    original_price DECIMAL(12,2) NOT NULL DEFAULT 0,
    cost_price DECIMAL(12,2) NOT NULL DEFAULT 0,

    -- 库存与销量
    stock INT NOT NULL DEFAULT 0,
    sales BIGINT NOT NULL DEFAULT 0,
    warn_stock INT NOT NULL DEFAULT 0,

    -- 显示
    image VARCHAR(255) NOT NULL DEFAULT '',
    sort INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 1              -- 0禁用 1启用
);
CREATE INDEX IF NOT EXISTS idx_mall_skus_region_id ON mall_skus(region_id);
CREATE INDEX IF NOT EXISTS idx_mall_skus_product_id ON mall_skus(product_id);
CREATE INDEX IF NOT EXISTS idx_mall_skus_shop_id ON mall_skus(shop_id);
CREATE INDEX IF NOT EXISTS idx_mall_skus_user_id ON mall_skus(user_id);
CREATE INDEX IF NOT EXISTS idx_mall_skus_sku_code ON mall_skus(sku_code);
CREATE INDEX IF NOT EXISTS idx_mall_skus_barcode ON mall_skus(barcode);
CREATE INDEX IF NOT EXISTS idx_mall_skus_price ON mall_skus(price);
CREATE INDEX IF NOT EXISTS idx_mall_skus_stock ON mall_skus(stock);
CREATE INDEX IF NOT EXISTS idx_mall_skus_sales ON mall_skus(sales);
CREATE INDEX IF NOT EXISTS idx_mall_skus_sort ON mall_skus(sort);
CREATE INDEX IF NOT EXISTS idx_mall_skus_status ON mall_skus(status);
CREATE INDEX IF NOT EXISTS idx_mall_skus_deleted_at ON mall_skus(deleted_at);

COMMENT ON TABLE mall_skus IS '同城商城商品 SKU 规格表';
COMMENT ON COLUMN mall_skus.specs IS '规格键值 JSONB（如 [{name:颜色,value:红}]）';
COMMENT ON COLUMN mall_skus.status IS '状态：0禁用 1启用';

-- ============================================================
-- 5. mall_cart 购物车表（注意：model.TableName() 返回 mall_cart 单数）
-- ============================================================
CREATE TABLE IF NOT EXISTS mall_cart (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    -- 关联
    user_id BIGINT NOT NULL,
    shop_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    sku_id BIGINT NOT NULL DEFAULT 0,

    -- 冗余信息
    product_name VARCHAR(200) NOT NULL DEFAULT '',
    main_image VARCHAR(255) NOT NULL DEFAULT '',
    sku_name VARCHAR(200) NOT NULL DEFAULT '',
    sku_specs VARCHAR(500) NOT NULL DEFAULT '',

    -- 价格与数量
    price DECIMAL(12,2) NOT NULL DEFAULT 0,
    quantity INT NOT NULL DEFAULT 1,
    selected INT NOT NULL DEFAULT 1,           -- 0未选中 1已选中

    -- 状态
    status INT NOT NULL DEFAULT 1              -- 0失效 1有效
);
CREATE INDEX IF NOT EXISTS idx_mall_cart_region_id ON mall_cart(region_id);
CREATE INDEX IF NOT EXISTS idx_mall_cart_user_id ON mall_cart(user_id);
CREATE INDEX IF NOT EXISTS idx_mall_cart_shop_id ON mall_cart(shop_id);
CREATE INDEX IF NOT EXISTS idx_mall_cart_product_id ON mall_cart(product_id);
CREATE INDEX IF NOT EXISTS idx_mall_cart_sku_id ON mall_cart(sku_id);
CREATE INDEX IF NOT EXISTS idx_mall_cart_selected ON mall_cart(selected);
CREATE INDEX IF NOT EXISTS idx_mall_cart_status ON mall_cart(status);
CREATE INDEX IF NOT EXISTS idx_mall_cart_deleted_at ON mall_cart(deleted_at);

COMMENT ON TABLE mall_cart IS '同城商城购物车表';
COMMENT ON COLUMN mall_cart.selected IS '选中状态：0未选中 1已选中';
COMMENT ON COLUMN mall_cart.status IS '状态：0失效（商品下架/删除） 1有效';

-- ============================================================
-- 6. mall_orders 订单主表
-- ============================================================
CREATE TABLE IF NOT EXISTS mall_orders (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    -- 订单基础
    order_no VARCHAR(32) NOT NULL DEFAULT '',
    user_id BIGINT NOT NULL,
    shop_id BIGINT NOT NULL,
    shop_name VARCHAR(200) NOT NULL DEFAULT '',

    -- 买家信息
    buyer_name VARCHAR(50) NOT NULL DEFAULT '',
    buyer_phone VARCHAR(32) NOT NULL DEFAULT '',
    buyer_avatar VARCHAR(255) NOT NULL DEFAULT '',

    -- 收货地址快照
    address_id BIGINT NOT NULL DEFAULT 0,
    receiver_name VARCHAR(50) NOT NULL DEFAULT '',
    receiver_phone VARCHAR(32) NOT NULL DEFAULT '',
    province VARCHAR(64) NOT NULL DEFAULT '',
    city VARCHAR(64) NOT NULL DEFAULT '',
    district VARCHAR(64) NOT NULL DEFAULT '',
    address VARCHAR(500) NOT NULL DEFAULT '',
    zip_code VARCHAR(20) NOT NULL DEFAULT '',

    -- 金额
    total_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    shipping_fee DECIMAL(12,2) NOT NULL DEFAULT 0,
    discount_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    tax_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    pay_amount DECIMAL(12,2) NOT NULL DEFAULT 0,

    -- 状态
    status INT NOT NULL DEFAULT 0,                -- 0待付款 1已付款 2已发货 3已收货 4已完成 5已取消 6已退款 7已关闭
    status_text VARCHAR(32) NOT NULL DEFAULT '',
    remark VARCHAR(500) NOT NULL DEFAULT '',
    seller_remark VARCHAR(500) NOT NULL DEFAULT '',
    admin_remark VARCHAR(500) NOT NULL DEFAULT '',

    -- 时间节点
    paid_at TIMESTAMPTZ,
    shipped_at TIMESTAMPTZ,
    received_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    auto_close_at TIMESTAMPTZ,
    auto_confirm_at TIMESTAMPTZ,
    auto_review_at TIMESTAMPTZ,

    -- 支付信息
    payment_method VARCHAR(32) NOT NULL DEFAULT '',
    payment_no VARCHAR(64) NOT NULL DEFAULT '',

    -- 来源
    source VARCHAR(32) NOT NULL DEFAULT 'app',
    client_ip VARCHAR(64) NOT NULL DEFAULT '',
    user_agent VARCHAR(255) NOT NULL DEFAULT '',

    -- 优惠券
    coupon_id BIGINT NOT NULL DEFAULT 0,
    coupon_name VARCHAR(128) NOT NULL DEFAULT '',

    -- 评价标记
    has_review BOOLEAN NOT NULL DEFAULT FALSE,
    has_seller_review BOOLEAN NOT NULL DEFAULT FALSE,

    -- 物流
    logistics_company VARCHAR(64) NOT NULL DEFAULT '',
    logistics_no VARCHAR(64) NOT NULL DEFAULT '',

    -- 风控
    risk_score INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_mall_orders_region_id ON mall_orders(region_id);
CREATE INDEX IF NOT EXISTS idx_mall_orders_order_no ON mall_orders(order_no);
CREATE INDEX IF NOT EXISTS idx_mall_orders_user_id ON mall_orders(user_id);
CREATE INDEX IF NOT EXISTS idx_mall_orders_shop_id ON mall_orders(shop_id);
CREATE INDEX IF NOT EXISTS idx_mall_orders_address_id ON mall_orders(address_id);
CREATE INDEX IF NOT EXISTS idx_mall_orders_pay_amount ON mall_orders(pay_amount);
CREATE INDEX IF NOT EXISTS idx_mall_orders_status ON mall_orders(status);
CREATE INDEX IF NOT EXISTS idx_mall_orders_paid_at ON mall_orders(paid_at);
CREATE INDEX IF NOT EXISTS idx_mall_orders_shipped_at ON mall_orders(shipped_at);
CREATE INDEX IF NOT EXISTS idx_mall_orders_received_at ON mall_orders(received_at);
CREATE INDEX IF NOT EXISTS idx_mall_orders_completed_at ON mall_orders(completed_at);
CREATE INDEX IF NOT EXISTS idx_mall_orders_cancelled_at ON mall_orders(cancelled_at);
CREATE INDEX IF NOT EXISTS idx_mall_orders_auto_close_at ON mall_orders(auto_close_at);
CREATE INDEX IF NOT EXISTS idx_mall_orders_auto_confirm_at ON mall_orders(auto_confirm_at);
CREATE INDEX IF NOT EXISTS idx_mall_orders_auto_review_at ON mall_orders(auto_review_at);
CREATE INDEX IF NOT EXISTS idx_mall_orders_payment_no ON mall_orders(payment_no);
CREATE INDEX IF NOT EXISTS idx_mall_orders_source ON mall_orders(source);
CREATE INDEX IF NOT EXISTS idx_mall_orders_coupon_id ON mall_orders(coupon_id);
CREATE INDEX IF NOT EXISTS idx_mall_orders_has_review ON mall_orders(has_review) WHERE has_review = TRUE;
CREATE INDEX IF NOT EXISTS idx_mall_orders_logistics_no ON mall_orders(logistics_no);
CREATE INDEX IF NOT EXISTS idx_mall_orders_risk_score ON mall_orders(risk_score);
CREATE INDEX IF NOT EXISTS idx_mall_orders_deleted_at ON mall_orders(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_mall_orders_order_no ON mall_orders(order_no);

COMMENT ON TABLE mall_orders IS '同城商城订单主表';
COMMENT ON COLUMN mall_orders.status IS '状态：0待付款 1已付款 2已发货 3已收货 4已完成 5已取消 6已退款 7已关闭';
COMMENT ON COLUMN mall_orders.payment_method IS '支付方式：wechat/alipay/balance/cod/bankcard';
COMMENT ON COLUMN mall_orders.source IS '来源：app/web/miniapp/admin';

-- ============================================================
-- 7. mall_order_items 订单明细表
-- ============================================================
CREATE TABLE IF NOT EXISTS mall_order_items (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    -- 关联
    order_id BIGINT NOT NULL,
    order_no VARCHAR(32) NOT NULL DEFAULT '',
    user_id BIGINT NOT NULL,
    shop_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    sku_id BIGINT NOT NULL DEFAULT 0,

    -- 商品快照
    product_name VARCHAR(200) NOT NULL DEFAULT '',
    main_image VARCHAR(255) NOT NULL DEFAULT '',
    sku_name VARCHAR(200) NOT NULL DEFAULT '',
    sku_specs VARCHAR(500) NOT NULL DEFAULT '',
    sku_code VARCHAR(64) NOT NULL DEFAULT '',

    -- 金额
    price DECIMAL(12,2) NOT NULL DEFAULT 0,
    quantity INT NOT NULL DEFAULT 1,
    total_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    discount_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    shipping_fee DECIMAL(12,2) NOT NULL DEFAULT 0,
    pay_amount DECIMAL(12,2) NOT NULL DEFAULT 0,

    -- 评价
    has_review BOOLEAN NOT NULL DEFAULT FALSE,
    review_id BIGINT NOT NULL DEFAULT 0,

    -- 退款
    refund_status INT NOT NULL DEFAULT 0,    -- 0未退款 1待审核 2已同意 3已拒绝 4已退款 5已关闭
    refund_id BIGINT NOT NULL DEFAULT 0,

    -- 状态
    status INT NOT NULL DEFAULT 0             -- 0待付款 1已付款 2已发货 3已收货 4已完成 5已取消 6已退款 7已关闭
);
CREATE INDEX IF NOT EXISTS idx_mall_order_items_region_id ON mall_order_items(region_id);
CREATE INDEX IF NOT EXISTS idx_mall_order_items_order_id ON mall_order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_mall_order_items_order_no ON mall_order_items(order_no);
CREATE INDEX IF NOT EXISTS idx_mall_order_items_user_id ON mall_order_items(user_id);
CREATE INDEX IF NOT EXISTS idx_mall_order_items_shop_id ON mall_order_items(shop_id);
CREATE INDEX IF NOT EXISTS idx_mall_order_items_product_id ON mall_order_items(product_id);
CREATE INDEX IF NOT EXISTS idx_mall_order_items_sku_id ON mall_order_items(sku_id);
CREATE INDEX IF NOT EXISTS idx_mall_order_items_review_id ON mall_order_items(review_id);
CREATE INDEX IF NOT EXISTS idx_mall_order_items_refund_status ON mall_order_items(refund_status);
CREATE INDEX IF NOT EXISTS idx_mall_order_items_refund_id ON mall_order_items(refund_id);
CREATE INDEX IF NOT EXISTS idx_mall_order_items_status ON mall_order_items(status);
CREATE INDEX IF NOT EXISTS idx_mall_order_items_deleted_at ON mall_order_items(deleted_at);

COMMENT ON TABLE mall_order_items IS '同城商城订单明细表（商品快照）';
COMMENT ON COLUMN mall_order_items.refund_status IS '退款状态：0未退款 1待审核 2已同意 3已拒绝 4已退款 5已关闭';
COMMENT ON COLUMN mall_order_items.status IS '状态：0待付款 1已付款 2已发货 3已收货 4已完成 5已取消 6已退款 7已关闭';

-- ============================================================
-- 8. mall_addresses 收货地址表
-- ============================================================
CREATE TABLE IF NOT EXISTS mall_addresses (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    -- 关联
    user_id BIGINT NOT NULL,

    -- 收货人信息
    name VARCHAR(50) NOT NULL DEFAULT '',
    phone VARCHAR(32) NOT NULL DEFAULT '',
    zip_code VARCHAR(20) NOT NULL DEFAULT '',

    -- 行政区划
    province VARCHAR(64) NOT NULL DEFAULT '',
    city VARCHAR(64) NOT NULL DEFAULT '',
    district VARCHAR(64) NOT NULL DEFAULT '',
    province_code VARCHAR(16) NOT NULL DEFAULT '',
    city_code VARCHAR(16) NOT NULL DEFAULT '',
    district_code VARCHAR(16) NOT NULL DEFAULT '',

    -- 详细地址
    detail VARCHAR(500) NOT NULL DEFAULT '',
    latitude DECIMAL(10,7) NOT NULL DEFAULT 0,
    longitude DECIMAL(10,7) NOT NULL DEFAULT 0,

    -- 标签与默认
    tag VARCHAR(32) NOT NULL DEFAULT '',
    is_default INT NOT NULL DEFAULT 0,         -- 0非默认 1默认

    -- 状态
    status INT NOT NULL DEFAULT 1             -- 0禁用 1启用
);
CREATE INDEX IF NOT EXISTS idx_mall_addresses_region_id ON mall_addresses(region_id);
CREATE INDEX IF NOT EXISTS idx_mall_addresses_user_id ON mall_addresses(user_id);
CREATE INDEX IF NOT EXISTS idx_mall_addresses_phone ON mall_addresses(phone);
CREATE INDEX IF NOT EXISTS idx_mall_addresses_province ON mall_addresses(province);
CREATE INDEX IF NOT EXISTS idx_mall_addresses_city ON mall_addresses(city);
CREATE INDEX IF NOT EXISTS idx_mall_addresses_district ON mall_addresses(district);
CREATE INDEX IF NOT EXISTS idx_mall_addresses_is_default ON mall_addresses(is_default) WHERE is_default = 1;
CREATE INDEX IF NOT EXISTS idx_mall_addresses_status ON mall_addresses(status);
CREATE INDEX IF NOT EXISTS idx_mall_addresses_deleted_at ON mall_addresses(deleted_at);

COMMENT ON TABLE mall_addresses IS '同城商城收货地址表';
COMMENT ON COLUMN mall_addresses.is_default IS '默认标记：0非默认 1默认';
COMMENT ON COLUMN mall_addresses.status IS '状态：0禁用 1启用';

-- ============================================================
-- 9. mall_payments 支付记录表
-- ============================================================
CREATE TABLE IF NOT EXISTS mall_payments (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    -- 关联
    order_id BIGINT NOT NULL,
    order_no VARCHAR(32) NOT NULL DEFAULT '',
    user_id BIGINT NOT NULL,
    shop_id BIGINT NOT NULL,

    -- 支付单信息
    payment_no VARCHAR(64) NOT NULL DEFAULT '',
    trade_no VARCHAR(64) NOT NULL DEFAULT '',
    out_trade_no VARCHAR(64) NOT NULL DEFAULT '',

    -- 金额
    amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    refund_amount DECIMAL(12,2) NOT NULL DEFAULT 0,

    -- 支付方式
    method VARCHAR(32) NOT NULL DEFAULT 'wechat',  -- wechat/alipay/balance/cod/bankcard
    channel VARCHAR(32) NOT NULL DEFAULT '',

    -- 状态
    status INT NOT NULL DEFAULT 0,                -- 0待支付 1成功 2失败 3已退款 4已关闭
    error_code VARCHAR(32) NOT NULL DEFAULT '',
    error_msg VARCHAR(500) NOT NULL DEFAULT '',

    -- 时间
    paid_at TIMESTAMPTZ,
    expired_at TIMESTAMPTZ,
    closed_at TIMESTAMPTZ,
    refunded_at TIMESTAMPTZ,

    -- 第三方原始响应
    raw_response JSONB,

    -- 客户端
    client_ip VARCHAR(64) NOT NULL DEFAULT '',
    user_agent VARCHAR(255) NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_mall_payments_region_id ON mall_payments(region_id);
CREATE INDEX IF NOT EXISTS idx_mall_payments_order_id ON mall_payments(order_id);
CREATE INDEX IF NOT EXISTS idx_mall_payments_order_no ON mall_payments(order_no);
CREATE INDEX IF NOT EXISTS idx_mall_payments_user_id ON mall_payments(user_id);
CREATE INDEX IF NOT EXISTS idx_mall_payments_shop_id ON mall_payments(shop_id);
CREATE INDEX IF NOT EXISTS idx_mall_payments_payment_no ON mall_payments(payment_no);
CREATE INDEX IF NOT EXISTS idx_mall_payments_trade_no ON mall_payments(trade_no);
CREATE INDEX IF NOT EXISTS idx_mall_payments_method ON mall_payments(method);
CREATE INDEX IF NOT EXISTS idx_mall_payments_status ON mall_payments(status);
CREATE INDEX IF NOT EXISTS idx_mall_payments_paid_at ON mall_payments(paid_at);
CREATE INDEX IF NOT EXISTS idx_mall_payments_expired_at ON mall_payments(expired_at);
CREATE INDEX IF NOT EXISTS idx_mall_payments_closed_at ON mall_payments(closed_at);
CREATE INDEX IF NOT EXISTS idx_mall_payments_refunded_at ON mall_payments(refunded_at);
CREATE INDEX IF NOT EXISTS idx_mall_payments_deleted_at ON mall_payments(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_mall_payments_payment_no ON mall_payments(payment_no);

COMMENT ON TABLE mall_payments IS '同城商城支付记录表';
COMMENT ON COLUMN mall_payments.method IS '支付方式：wechat/alipay/balance/cod/bankcard';
COMMENT ON COLUMN mall_payments.status IS '状态：0待支付 1成功 2失败 3已退款 4已关闭';
COMMENT ON COLUMN mall_payments.raw_response IS '第三方回调原始数据 JSONB';

-- ============================================================
-- 10. mall_refunds 退款表
-- ============================================================
CREATE TABLE IF NOT EXISTS mall_refunds (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    -- 关联
    order_id BIGINT NOT NULL,
    order_no VARCHAR(32) NOT NULL DEFAULT '',
    order_item_id BIGINT NOT NULL DEFAULT 0,
    user_id BIGINT NOT NULL,
    shop_id BIGINT NOT NULL,
    payment_id BIGINT NOT NULL DEFAULT 0,

    -- 退款单信息
    refund_no VARCHAR(64) NOT NULL DEFAULT '',
    trade_no VARCHAR(64) NOT NULL DEFAULT '',

    -- 金额
    amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    refund_amount DECIMAL(12,2) NOT NULL DEFAULT 0,

    -- 退款类型与原因
    refund_type VARCHAR(32) NOT NULL DEFAULT 'only_refund',  -- only_refund/return
    reason VARCHAR(500) NOT NULL DEFAULT '',
    reason_code VARCHAR(32) NOT NULL DEFAULT '',
    description TEXT,
    evidence_images JSONB,
    express_company VARCHAR(64) NOT NULL DEFAULT '',
    express_no VARCHAR(64) NOT NULL DEFAULT '',

    -- 状态
    status INT NOT NULL DEFAULT 0,                -- 0待审核 1已同意 2已拒绝 3已退款 4已关闭
    seller_remark VARCHAR(500) NOT NULL DEFAULT '',
    admin_remark VARCHAR(500) NOT NULL DEFAULT '',

    -- 处理人
    handler_id BIGINT NOT NULL DEFAULT 0,
    handler_name VARCHAR(50) NOT NULL DEFAULT '',

    -- 时间
    approved_at TIMESTAMPTZ,
    rejected_at TIMESTAMPTZ,
    refunded_at TIMESTAMPTZ,
    closed_at TIMESTAMPTZ,
    shipped_at TIMESTAMPTZ,
    received_at TIMESTAMPTZ,

    -- 第三方原始响应
    raw_response JSONB
);
CREATE INDEX IF NOT EXISTS idx_mall_refunds_region_id ON mall_refunds(region_id);
CREATE INDEX IF NOT EXISTS idx_mall_refunds_order_id ON mall_refunds(order_id);
CREATE INDEX IF NOT EXISTS idx_mall_refunds_order_no ON mall_refunds(order_no);
CREATE INDEX IF NOT EXISTS idx_mall_refunds_order_item_id ON mall_refunds(order_item_id);
CREATE INDEX IF NOT EXISTS idx_mall_refunds_user_id ON mall_refunds(user_id);
CREATE INDEX IF NOT EXISTS idx_mall_refunds_shop_id ON mall_refunds(shop_id);
CREATE INDEX IF NOT EXISTS idx_mall_refunds_payment_id ON mall_refunds(payment_id);
CREATE INDEX IF NOT EXISTS idx_mall_refunds_refund_no ON mall_refunds(refund_no);
CREATE INDEX IF NOT EXISTS idx_mall_refunds_trade_no ON mall_refunds(trade_no);
CREATE INDEX IF NOT EXISTS idx_mall_refunds_refund_type ON mall_refunds(refund_type);
CREATE INDEX IF NOT EXISTS idx_mall_refunds_status ON mall_refunds(status);
CREATE INDEX IF NOT EXISTS idx_mall_refunds_handler_id ON mall_refunds(handler_id);
CREATE INDEX IF NOT EXISTS idx_mall_refunds_approved_at ON mall_refunds(approved_at);
CREATE INDEX IF NOT EXISTS idx_mall_refunds_rejected_at ON mall_refunds(rejected_at);
CREATE INDEX IF NOT EXISTS idx_mall_refunds_refunded_at ON mall_refunds(refunded_at);
CREATE INDEX IF NOT EXISTS idx_mall_refunds_closed_at ON mall_refunds(closed_at);
CREATE INDEX IF NOT EXISTS idx_mall_refunds_shipped_at ON mall_refunds(shipped_at);
CREATE INDEX IF NOT EXISTS idx_mall_refunds_received_at ON mall_refunds(received_at);
CREATE INDEX IF NOT EXISTS idx_mall_refunds_deleted_at ON mall_refunds(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_mall_refunds_refund_no ON mall_refunds(refund_no);

COMMENT ON TABLE mall_refunds IS '同城商城退款记录表';
COMMENT ON COLUMN mall_refunds.refund_type IS '退款类型：only_refund仅退款/return退货退款';
COMMENT ON COLUMN mall_refunds.status IS '状态：0待审核 1已同意 2已拒绝 3已退款 4已关闭';
COMMENT ON COLUMN mall_refunds.evidence_images IS '证据图片 JSONB';
COMMENT ON COLUMN mall_refunds.raw_response IS '第三方退款响应原始数据 JSONB';

-- ============================================================
-- 11. mall_logistics 物流记录表
-- ============================================================
CREATE TABLE IF NOT EXISTS mall_logistics (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    -- 关联
    order_id BIGINT NOT NULL,
    order_no VARCHAR(32) NOT NULL DEFAULT '',
    user_id BIGINT NOT NULL,
    shop_id BIGINT NOT NULL,

    -- 物流信息
    company VARCHAR(64) NOT NULL DEFAULT '',
    company_code VARCHAR(32) NOT NULL DEFAULT '',
    tracking_no VARCHAR(64) NOT NULL DEFAULT '',
    courier_name VARCHAR(50) NOT NULL DEFAULT '',
    courier_phone VARCHAR(32) NOT NULL DEFAULT '',

    -- 发货/收货地址
    sender_name VARCHAR(50) NOT NULL DEFAULT '',
    sender_phone VARCHAR(32) NOT NULL DEFAULT '',
    sender_address VARCHAR(500) NOT NULL DEFAULT '',
    receiver_name VARCHAR(50) NOT NULL DEFAULT '',
    receiver_phone VARCHAR(32) NOT NULL DEFAULT '',
    receiver_address VARCHAR(500) NOT NULL DEFAULT '',

    -- 状态
    status INT NOT NULL DEFAULT 0,                -- 0待发货 1已发货 2运输中 3已派送 4已签收 5已退回
    shipped_at TIMESTAMPTZ,
    in_transit_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    received_at TIMESTAMPTZ,
    returned_at TIMESTAMPTZ,

    -- 物流轨迹
    traces JSONB,

    -- 重量与体积
    weight DECIMAL(10,3) NOT NULL DEFAULT 0,
    volume DECIMAL(10,3) NOT NULL DEFAULT 0,
    pieces INT NOT NULL DEFAULT 1,

    -- 运费
    freight DECIMAL(12,2) NOT NULL DEFAULT 0,
    insured_fee DECIMAL(12,2) NOT NULL DEFAULT 0,
    cod_fee DECIMAL(12,2) NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_mall_logistics_region_id ON mall_logistics(region_id);
CREATE INDEX IF NOT EXISTS idx_mall_logistics_order_id ON mall_logistics(order_id);
CREATE INDEX IF NOT EXISTS idx_mall_logistics_order_no ON mall_logistics(order_no);
CREATE INDEX IF NOT EXISTS idx_mall_logistics_user_id ON mall_logistics(user_id);
CREATE INDEX IF NOT EXISTS idx_mall_logistics_shop_id ON mall_logistics(shop_id);
CREATE INDEX IF NOT EXISTS idx_mall_logistics_company ON mall_logistics(company);
CREATE INDEX IF NOT EXISTS idx_mall_logistics_tracking_no ON mall_logistics(tracking_no);
CREATE INDEX IF NOT EXISTS idx_mall_logistics_status ON mall_logistics(status);
CREATE INDEX IF NOT EXISTS idx_mall_logistics_shipped_at ON mall_logistics(shipped_at);
CREATE INDEX IF NOT EXISTS idx_mall_logistics_in_transit_at ON mall_logistics(in_transit_at);
CREATE INDEX IF NOT EXISTS idx_mall_logistics_delivered_at ON mall_logistics(delivered_at);
CREATE INDEX IF NOT EXISTS idx_mall_logistics_received_at ON mall_logistics(received_at);
CREATE INDEX IF NOT EXISTS idx_mall_logistics_returned_at ON mall_logistics(returned_at);
CREATE INDEX IF NOT EXISTS idx_mall_logistics_deleted_at ON mall_logistics(deleted_at);

COMMENT ON TABLE mall_logistics IS '同城商城物流记录表';
COMMENT ON COLUMN mall_logistics.status IS '状态：0待发货 1已发货 2运输中 3已派送 4已签收 5已退回';
COMMENT ON COLUMN mall_logistics.traces IS '物流轨迹 JSONB（[{time,desc,status}]）';

-- ============================================================
-- 12. mall_reviews 商品评价表
-- ============================================================
CREATE TABLE IF NOT EXISTS mall_reviews (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    -- 关联
    product_id BIGINT NOT NULL,
    sku_id BIGINT NOT NULL DEFAULT 0,
    order_id BIGINT NOT NULL,
    order_no VARCHAR(32) NOT NULL DEFAULT '',
    order_item_id BIGINT NOT NULL DEFAULT 0,
    user_id BIGINT NOT NULL,
    shop_id BIGINT NOT NULL,

    -- 评价人信息
    user_name VARCHAR(50) NOT NULL DEFAULT '',
    user_avatar VARCHAR(255) NOT NULL DEFAULT '',
    is_anonymous BOOLEAN NOT NULL DEFAULT FALSE,

    -- 评价内容
    rating INT NOT NULL DEFAULT 5,                -- 1-5
    content TEXT,
    images JSONB,
    video VARCHAR(255) NOT NULL DEFAULT '',

    -- 规格
    sku_name VARCHAR(200) NOT NULL DEFAULT '',
    sku_specs VARCHAR(500) NOT NULL DEFAULT '',

    -- 卖家回复
    reply TEXT,
    reply_at TIMESTAMPTZ,
    reply_user_id BIGINT NOT NULL DEFAULT 0,

    -- 评价标签
    tags JSONB,

    -- 状态
    status INT NOT NULL DEFAULT 0,                -- 0待审 1通过 2拒绝 3隐藏
    audit_reason VARCHAR(500) NOT NULL DEFAULT '',
    has_seller_reply BOOLEAN NOT NULL DEFAULT FALSE,

    -- 互动
    like_count INT NOT NULL DEFAULT 0,
    dislike_count INT NOT NULL DEFAULT 0,
    reply_count INT NOT NULL DEFAULT 0,

    -- 追评
    append_content TEXT,
    append_images JSONB,
    append_at TIMESTAMPTZ,

    -- 类型
    type VARCHAR(32) NOT NULL DEFAULT 'product',  -- product/logistics/service

    -- 风控
    content_hash VARCHAR(64) NOT NULL DEFAULT '',
    risk_score INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_mall_reviews_region_id ON mall_reviews(region_id);
CREATE INDEX IF NOT EXISTS idx_mall_reviews_product_id ON mall_reviews(product_id);
CREATE INDEX IF NOT EXISTS idx_mall_reviews_sku_id ON mall_reviews(sku_id);
CREATE INDEX IF NOT EXISTS idx_mall_reviews_order_id ON mall_reviews(order_id);
CREATE INDEX IF NOT EXISTS idx_mall_reviews_order_no ON mall_reviews(order_no);
CREATE INDEX IF NOT EXISTS idx_mall_reviews_order_item_id ON mall_reviews(order_item_id);
CREATE INDEX IF NOT EXISTS idx_mall_reviews_user_id ON mall_reviews(user_id);
CREATE INDEX IF NOT EXISTS idx_mall_reviews_shop_id ON mall_reviews(shop_id);
CREATE INDEX IF NOT EXISTS idx_mall_reviews_rating ON mall_reviews(rating);
CREATE INDEX IF NOT EXISTS idx_mall_reviews_reply_at ON mall_reviews(reply_at);
CREATE INDEX IF NOT EXISTS idx_mall_reviews_reply_user_id ON mall_reviews(reply_user_id);
CREATE INDEX IF NOT EXISTS idx_mall_reviews_status ON mall_reviews(status);
CREATE INDEX IF NOT EXISTS idx_mall_reviews_has_seller_reply ON mall_reviews(has_seller_reply) WHERE has_seller_reply = TRUE;
CREATE INDEX IF NOT EXISTS idx_mall_reviews_append_at ON mall_reviews(append_at);
CREATE INDEX IF NOT EXISTS idx_mall_reviews_type ON mall_reviews(type);
CREATE INDEX IF NOT EXISTS idx_mall_reviews_content_hash ON mall_reviews(content_hash);
CREATE INDEX IF NOT EXISTS idx_mall_reviews_risk_score ON mall_reviews(risk_score);
CREATE INDEX IF NOT EXISTS idx_mall_reviews_deleted_at ON mall_reviews(deleted_at);

COMMENT ON TABLE mall_reviews IS '同城商城商品评价表';
COMMENT ON COLUMN mall_reviews.rating IS '评分 1-5';
COMMENT ON COLUMN mall_reviews.status IS '状态：0待审 1通过 2拒绝 3隐藏';
COMMENT ON COLUMN mall_reviews.type IS '类型：product商品评价/logistics物流评价/service服务评价';
COMMENT ON COLUMN mall_reviews.images IS '评价图片 JSONB';
COMMENT ON COLUMN mall_reviews.tags IS '评价标签 JSONB';
COMMENT ON COLUMN mall_reviews.append_images IS '追评图片 JSONB';

-- ============================================================
-- 13. mall_audit_rules 审核规则表（BaseModel 无 region_id，全局表）
-- ============================================================
CREATE TABLE IF NOT EXISTS mall_audit_rules (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    -- 规则
    rule_name VARCHAR(128) NOT NULL,
    rule_type VARCHAR(32) NOT NULL,               -- sensitive_word/prohibited/contact/price_check/frequency
    rule_key VARCHAR(64) NOT NULL DEFAULT '',
    pattern TEXT,
    threshold JSONB,
    action VARCHAR(32) NOT NULL DEFAULT 'reject', -- reject/approval/filter/limit
    penalty_type VARCHAR(32) NOT NULL DEFAULT '', -- warning/ban24h/ban7d/ban30d/ban_forever/delete/limit
    severity INT NOT NULL DEFAULT 1,              -- 1-5
    status INT NOT NULL DEFAULT 1,               -- 0禁用 1启用
    description VARCHAR(500) NOT NULL DEFAULT '',
    sort INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_mall_audit_rules_rule_type ON mall_audit_rules(rule_type);
CREATE INDEX IF NOT EXISTS idx_mall_audit_rules_rule_key ON mall_audit_rules(rule_key);
CREATE INDEX IF NOT EXISTS idx_mall_audit_rules_action ON mall_audit_rules(action);
CREATE INDEX IF NOT EXISTS idx_mall_audit_rules_severity ON mall_audit_rules(severity);
CREATE INDEX IF NOT EXISTS idx_mall_audit_rules_status ON mall_audit_rules(status);
CREATE INDEX IF NOT EXISTS idx_mall_audit_rules_sort ON mall_audit_rules(sort);
CREATE INDEX IF NOT EXISTS idx_mall_audit_rules_deleted_at ON mall_audit_rules(deleted_at);

COMMENT ON TABLE mall_audit_rules IS '同城商城审核规则表（全局，BaseModel 无 region_id）';
COMMENT ON COLUMN mall_audit_rules.rule_type IS '规则类型：sensitive_word/prohibited/contact/price_check/frequency';
COMMENT ON COLUMN mall_audit_rules.action IS '动作：reject/approval/filter/limit';
COMMENT ON COLUMN mall_audit_rules.status IS '状态：0禁用 1启用';

-- ============================================================
-- 14. mall_reports 举报表
-- ============================================================
CREATE TABLE IF NOT EXISTS mall_reports (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    -- 举报人
    report_no VARCHAR(32) NOT NULL DEFAULT '',
    reporter_id BIGINT NOT NULL DEFAULT 0,
    reporter_name VARCHAR(50) NOT NULL DEFAULT '',

    -- 被举报对象
    target_type VARCHAR(32) NOT NULL DEFAULT '',  -- shop/product/review/order/user
    target_id BIGINT NOT NULL DEFAULT 0,
    target_name VARCHAR(200) NOT NULL DEFAULT '',

    -- 举报内容
    report_type VARCHAR(32) NOT NULL DEFAULT '',  -- porn/scam/fake/prohibited/infringement/spam/other
    report_reason VARCHAR(500) NOT NULL DEFAULT '',
    description TEXT,
    evidence_images JSONB,
    contact_info VARCHAR(100) NOT NULL DEFAULT '',

    -- 处理
    status INT NOT NULL DEFAULT 0,                -- 0待处理 1已核实警告 2已下架 3已封号 4已驳回 5已转交
    handler_id BIGINT NOT NULL DEFAULT 0,
    handler_name VARCHAR(50) NOT NULL DEFAULT '',
    handle_result VARCHAR(500) NOT NULL DEFAULT '',
    handled_at TIMESTAMPTZ,

    -- 处罚
    penalty_type VARCHAR(32) NOT NULL DEFAULT '', -- warning/limit/ban1d/ban7d/banForever
    penalty_target_id BIGINT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_mall_reports_region_id ON mall_reports(region_id);
CREATE INDEX IF NOT EXISTS idx_mall_reports_report_no ON mall_reports(report_no);
CREATE INDEX IF NOT EXISTS idx_mall_reports_reporter_id ON mall_reports(reporter_id);
CREATE INDEX IF NOT EXISTS idx_mall_reports_target_type ON mall_reports(target_type);
CREATE INDEX IF NOT EXISTS idx_mall_reports_target_id ON mall_reports(target_id);
CREATE INDEX IF NOT EXISTS idx_mall_reports_report_type ON mall_reports(report_type);
CREATE INDEX IF NOT EXISTS idx_mall_reports_status ON mall_reports(status);
CREATE INDEX IF NOT EXISTS idx_mall_reports_handler_id ON mall_reports(handler_id);
CREATE INDEX IF NOT EXISTS idx_mall_reports_handled_at ON mall_reports(handled_at);
CREATE INDEX IF NOT EXISTS idx_mall_reports_penalty_target_id ON mall_reports(penalty_target_id);
CREATE INDEX IF NOT EXISTS idx_mall_reports_deleted_at ON mall_reports(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_mall_reports_report_no ON mall_reports(report_no);

COMMENT ON TABLE mall_reports IS '同城商城举报表';
COMMENT ON COLUMN mall_reports.target_type IS '被举报对象类型：shop/product/review/order/user';
COMMENT ON COLUMN mall_reports.report_type IS '举报类型：porn/scam/fake/prohibited/infringement/spam/other';
COMMENT ON COLUMN mall_reports.status IS '状态：0待处理 1已核实警告 2已下架 3已封号 4已驳回 5已转交';
COMMENT ON COLUMN mall_reports.evidence_images IS '证据图片 JSONB';

-- ============================================================
-- 15. mall_statistics 统计表（日/店铺/商品/分类统计）
-- ============================================================
CREATE TABLE IF NOT EXISTS mall_statistics (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    -- 统计维度
    stat_date DATE NOT NULL,
    stat_type VARCHAR(32) NOT NULL DEFAULT 'daily',  -- daily/shop/product/category
    shop_id BIGINT NOT NULL DEFAULT 0,
    product_id BIGINT NOT NULL DEFAULT 0,
    category_id BIGINT,

    -- 订单统计
    order_count BIGINT NOT NULL DEFAULT 0,
    order_amount DECIMAL(14,2) NOT NULL DEFAULT 0,
    paid_order_count BIGINT NOT NULL DEFAULT 0,
    paid_order_amount DECIMAL(14,2) NOT NULL DEFAULT 0,
    cancelled_order_count BIGINT NOT NULL DEFAULT 0,
    refund_count BIGINT NOT NULL DEFAULT 0,
    refund_amount DECIMAL(14,2) NOT NULL DEFAULT 0,

    -- 商品统计
    view_count BIGINT NOT NULL DEFAULT 0,
    favorite_count BIGINT NOT NULL DEFAULT 0,
    cart_count BIGINT NOT NULL DEFAULT 0,
    sales_count BIGINT NOT NULL DEFAULT 0,

    -- 评价统计
    review_count BIGINT NOT NULL DEFAULT 0,
    new_review_count BIGINT NOT NULL DEFAULT 0,
    avg_rating DECIMAL(3,2) NOT NULL DEFAULT 0,
    good_rate DECIMAL(5,2) NOT NULL DEFAULT 0,

    -- 用户统计
    new_buyer_count BIGINT NOT NULL DEFAULT 0,
    active_buyer_count BIGINT NOT NULL DEFAULT 0,
    repurchase_count BIGINT NOT NULL DEFAULT 0,

    -- 转化率
    conversion_rate DECIMAL(5,2) NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_mall_statistics_region_id ON mall_statistics(region_id);
CREATE INDEX IF NOT EXISTS idx_mall_statistics_stat_date ON mall_statistics(stat_date);
CREATE INDEX IF NOT EXISTS idx_mall_statistics_stat_type ON mall_statistics(stat_type);
CREATE INDEX IF NOT EXISTS idx_mall_statistics_shop_id ON mall_statistics(shop_id);
CREATE INDEX IF NOT EXISTS idx_mall_statistics_product_id ON mall_statistics(product_id);
CREATE INDEX IF NOT EXISTS idx_mall_statistics_category_id ON mall_statistics(category_id);
CREATE INDEX IF NOT EXISTS idx_mall_statistics_deleted_at ON mall_statistics(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_mall_statistics_date_type_target ON mall_statistics(stat_date, stat_type, shop_id, product_id, category_id);

COMMENT ON TABLE mall_statistics IS '同城商城统计表（日统计/店铺统计/商品统计/分类统计）';
COMMENT ON COLUMN mall_statistics.stat_type IS '统计类型：daily/shop/product/category';
COMMENT ON COLUMN mall_statistics.category_id IS '分类 ID（category 类型时使用，可空）';

-- ============================================================
-- updated_at 触发器（依赖 001_p0_baseline.sql 中的 update_updated_at_column 函数）
-- PostgreSQL 不支持 CREATE TRIGGER IF NOT EXISTS，用 DROP TRIGGER IF EXISTS + CREATE TRIGGER
-- ============================================================
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'update_updated_at_column') THEN
        DROP TRIGGER IF EXISTS trg_mall_shops_updated_at ON mall_shops; CREATE TRIGGER trg_mall_shops_updated_at BEFORE UPDATE ON mall_shops FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_mall_categories_updated_at ON mall_categories; CREATE TRIGGER trg_mall_categories_updated_at BEFORE UPDATE ON mall_categories FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_mall_products_updated_at ON mall_products; CREATE TRIGGER trg_mall_products_updated_at BEFORE UPDATE ON mall_products FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_mall_skus_updated_at ON mall_skus; CREATE TRIGGER trg_mall_skus_updated_at BEFORE UPDATE ON mall_skus FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_mall_cart_updated_at ON mall_cart; CREATE TRIGGER trg_mall_cart_updated_at BEFORE UPDATE ON mall_cart FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_mall_orders_updated_at ON mall_orders; CREATE TRIGGER trg_mall_orders_updated_at BEFORE UPDATE ON mall_orders FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_mall_order_items_updated_at ON mall_order_items; CREATE TRIGGER trg_mall_order_items_updated_at BEFORE UPDATE ON mall_order_items FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_mall_addresses_updated_at ON mall_addresses; CREATE TRIGGER trg_mall_addresses_updated_at BEFORE UPDATE ON mall_addresses FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_mall_payments_updated_at ON mall_payments; CREATE TRIGGER trg_mall_payments_updated_at BEFORE UPDATE ON mall_payments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_mall_refunds_updated_at ON mall_refunds; CREATE TRIGGER trg_mall_refunds_updated_at BEFORE UPDATE ON mall_refunds FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_mall_logistics_updated_at ON mall_logistics; CREATE TRIGGER trg_mall_logistics_updated_at BEFORE UPDATE ON mall_logistics FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_mall_reviews_updated_at ON mall_reviews; CREATE TRIGGER trg_mall_reviews_updated_at BEFORE UPDATE ON mall_reviews FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_mall_audit_rules_updated_at ON mall_audit_rules; CREATE TRIGGER trg_mall_audit_rules_updated_at BEFORE UPDATE ON mall_audit_rules FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_mall_reports_updated_at ON mall_reports; CREATE TRIGGER trg_mall_reports_updated_at BEFORE UPDATE ON mall_reports FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_mall_statistics_updated_at ON mall_statistics; CREATE TRIGGER trg_mall_statistics_updated_at BEFORE UPDATE ON mall_statistics FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;
