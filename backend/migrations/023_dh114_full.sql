-- ============================================================
-- dh114 同城114模块完整功能迁移脚本（v3.2.1）
-- 对标：大众点评 / 美团 / 58同城
--
-- 内容：
--   1. CREATE 17 张子表（dh114_ 前缀；dh114s 主表由 GORM AutoMigrate 创建，此处亦创建以保持幂等）
--   2. 索引、触发器、注释
--   3. 全幂等：CREATE TABLE IF NOT EXISTS
-- 依赖：001_p0_baseline.sql（update_updated_at_column 函数已创建）
--
-- 表清单（18 张，含主表）：
--   主表：dh114s
--   子表：dh114_business / dh114_business_hours / dh114_categories / dh114_images / dh114_tags /
--         dh114_menus / dh114_coupons / dh114_groupbuys / dh114_reviews / dh114_review_replies /
--         dh114_favorites / dh114_phone_calls / dh114_visits / dh114_recommendations /
--         dh114_statistics / dh114_audit_rules / dh114_verifications
--
-- 注意：dh114s 主表也会在 plugin.go Init 中由 GORM AutoMigrate 创建
--      此处 CREATE TABLE IF NOT EXISTS 保证幂等，二者不冲突
--
-- 说明：表名严格遵循 model.TableName() 定义：
--      - Dh114 → dh114s（主表，复数）
--      - Dh114Business → dh114_business（单数，1:1 关联）
--      - 其他子表均采用复数形式
-- ============================================================

-- ============================================================
-- 1. dh114s 商户主表（也由 GORM AutoMigrate 创建）
-- ============================================================
CREATE TABLE IF NOT EXISTS dh114s (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    -- 基础信息
    title VARCHAR(200) NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    cover_image VARCHAR(255) NOT NULL DEFAULT '',
    user_id BIGINT NOT NULL,
    user_name VARCHAR(50) NOT NULL DEFAULT '',
    user_phone VARCHAR(20) NOT NULL DEFAULT '',
    user_avatar VARCHAR(255) NOT NULL DEFAULT '',
    status INT NOT NULL DEFAULT 0,                -- 0草稿 1已发布 2下架 3过期 4删除
    audit_status INT NOT NULL DEFAULT 0,           -- 0待审 1通过 2拒绝
    audit_reason VARCHAR(500) NOT NULL DEFAULT '',
    published_at TIMESTAMPTZ,

    -- 分类关联
    category_id BIGINT,
    category_name VARCHAR(64) NOT NULL DEFAULT '',
    business_type VARCHAR(32) NOT NULL DEFAULT 'other',  -- restaurant/retail/service/entertain/hotel/medical/education/life/other

    -- 来源类型
    source_type VARCHAR(16) NOT NULL DEFAULT 'personal',  -- personal/merchant/chain

    -- 联系方式
    phone VARCHAR(32) NOT NULL DEFAULT '',
    alt_phone VARCHAR(32) NOT NULL DEFAULT '',
    website VARCHAR(255) NOT NULL DEFAULT '',
    wechat VARCHAR(64) NOT NULL DEFAULT '',

    -- 地理位置
    city VARCHAR(64) NOT NULL DEFAULT '',
    district VARCHAR(64) NOT NULL DEFAULT '',
    business_district VARCHAR(128) NOT NULL DEFAULT '',
    address VARCHAR(500) NOT NULL DEFAULT '',
    latitude DECIMAL(10,7) NOT NULL DEFAULT 0,
    longitude DECIMAL(10,7) NOT NULL DEFAULT 0,

    -- 评分统计
    rating DECIMAL(3,2) NOT NULL DEFAULT 0,
    review_count INT NOT NULL DEFAULT 0,
    price_avg DECIMAL(10,2) NOT NULL DEFAULT 0,

    -- 互动统计
    view_count INT NOT NULL DEFAULT 0,
    fav_count INT NOT NULL DEFAULT 0,
    contact_count INT NOT NULL DEFAULT 0,
    share_count INT NOT NULL DEFAULT 0,
    call_count INT NOT NULL DEFAULT 0,
    last_call_at TIMESTAMPTZ,

    -- 风控
    content_hash VARCHAR(64) NOT NULL DEFAULT '',
    risk_score INT NOT NULL DEFAULT 0,

    -- 视频/VR
    video_url VARCHAR(255) NOT NULL DEFAULT '',
    video_cover VARCHAR(255) NOT NULL DEFAULT '',
    vr_url VARCHAR(255) NOT NULL DEFAULT '',

    -- JSONB 字段
    images JSONB,
    tags JSONB,
    business_hours JSONB,
    features JSONB,

    -- 运营字段
    featured BOOLEAN NOT NULL DEFAULT FALSE,
    picked BOOLEAN NOT NULL DEFAULT FALSE,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    promotion_level INT NOT NULL DEFAULT 0,
    traffic_weight DECIMAL(3,2) NOT NULL DEFAULT 1.00,

    -- 认证信息
    verified_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_dh114s_region_id ON dh114s(region_id);
CREATE INDEX IF NOT EXISTS idx_dh114s_user_id ON dh114s(user_id);
CREATE INDEX IF NOT EXISTS idx_dh114s_status ON dh114s(status);
CREATE INDEX IF NOT EXISTS idx_dh114s_audit_status ON dh114s(audit_status);
CREATE INDEX IF NOT EXISTS idx_dh114s_category_id ON dh114s(category_id);
CREATE INDEX IF NOT EXISTS idx_dh114s_category_name ON dh114s(category_name);
CREATE INDEX IF NOT EXISTS idx_dh114s_business_type ON dh114s(business_type);
CREATE INDEX IF NOT EXISTS idx_dh114s_source_type ON dh114s(source_type);
CREATE INDEX IF NOT EXISTS idx_dh114s_phone ON dh114s(phone);
CREATE INDEX IF NOT EXISTS idx_dh114s_city ON dh114s(city);
CREATE INDEX IF NOT EXISTS idx_dh114s_district ON dh114s(district);
CREATE INDEX IF NOT EXISTS idx_dh114s_latitude ON dh114s(latitude);
CREATE INDEX IF NOT EXISTS idx_dh114s_longitude ON dh114s(longitude);
CREATE INDEX IF NOT EXISTS idx_dh114s_rating ON dh114s(rating);
CREATE INDEX IF NOT EXISTS idx_dh114s_review_count ON dh114s(review_count);
CREATE INDEX IF NOT EXISTS idx_dh114s_last_call_at ON dh114s(last_call_at);
CREATE INDEX IF NOT EXISTS idx_dh114s_content_hash ON dh114s(content_hash);
CREATE INDEX IF NOT EXISTS idx_dh114s_risk_score ON dh114s(risk_score);
CREATE INDEX IF NOT EXISTS idx_dh114s_verified_at ON dh114s(verified_at);
CREATE INDEX IF NOT EXISTS idx_dh114s_published_at ON dh114s(published_at);
CREATE INDEX IF NOT EXISTS idx_dh114s_deleted_at ON dh114s(deleted_at);
CREATE INDEX IF NOT EXISTS idx_dh114s_featured ON dh114s(featured) WHERE featured = TRUE;
CREATE INDEX IF NOT EXISTS idx_dh114s_picked ON dh114s(picked) WHERE picked = TRUE;
CREATE INDEX IF NOT EXISTS idx_dh114s_verified ON dh114s(verified) WHERE verified = TRUE;

COMMENT ON TABLE dh114s IS '同城114商户主表';
COMMENT ON COLUMN dh114s.status IS '状态：0草稿 1已发布 2下架 3过期 4删除';
COMMENT ON COLUMN dh114s.audit_status IS '审核状态：0待审 1通过 2拒绝';
COMMENT ON COLUMN dh114s.business_type IS '商户类型：restaurant/retail/service/entertain/hotel/medical/education/life/other';
COMMENT ON COLUMN dh114s.source_type IS '来源类型：personal/merchant/chain';

-- ============================================================
-- 2. dh114_business 商户详情表（1:1 关联 dh114s）
-- ============================================================
CREATE TABLE IF NOT EXISTS dh114_business (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,
    dh114_id BIGINT NOT NULL,

    -- 营业执照信息
    business_name VARCHAR(200) NOT NULL DEFAULT '',
    license_no VARCHAR(64) NOT NULL DEFAULT '',
    license_image VARCHAR(255) NOT NULL DEFAULT '',
    legal_person VARCHAR(64) NOT NULL DEFAULT '',
    legal_person_id_card VARCHAR(32) NOT NULL DEFAULT '',
    business_scope TEXT,
    registered_capital DECIMAL(14,2) NOT NULL DEFAULT 0,
    established_date DATE,
    registered_address VARCHAR(500) NOT NULL DEFAULT '',

    -- 营业时间
    opening_hours VARCHAR(32) NOT NULL DEFAULT '',
    closing_hours VARCHAR(32) NOT NULL DEFAULT '',
    open_all_day BOOLEAN NOT NULL DEFAULT FALSE,
    closed_days JSONB,

    -- 价格信息
    price_avg DECIMAL(10,2) NOT NULL DEFAULT 0,
    price_range_min DECIMAL(10,2) NOT NULL DEFAULT 0,
    price_range_max DECIMAL(10,2) NOT NULL DEFAULT 0,

    -- 联系方式扩展
    website VARCHAR(255) NOT NULL DEFAULT '',
    wechat VARCHAR(64) NOT NULL DEFAULT '',
    wechat_qr VARCHAR(255) NOT NULL DEFAULT '',
    email VARCHAR(128) NOT NULL DEFAULT '',

    -- 设施服务（JSONB）
    facilities JSONB,

    -- 认证状态
    verification_status INT NOT NULL DEFAULT 0,  -- 0待审 1通过 2拒绝 3过期
    verified_at TIMESTAMPTZ,
    valid_until DATE
);
CREATE INDEX IF NOT EXISTS idx_dh114_business_region_id ON dh114_business(region_id);
CREATE INDEX IF NOT EXISTS idx_dh114_business_dh114_id ON dh114_business(dh114_id);
CREATE INDEX IF NOT EXISTS idx_dh114_business_license_no ON dh114_business(license_no);
CREATE INDEX IF NOT EXISTS idx_dh114_business_verification_status ON dh114_business(verification_status);
CREATE INDEX IF NOT EXISTS idx_dh114_business_verified_at ON dh114_business(verified_at);
CREATE INDEX IF NOT EXISTS idx_dh114_business_valid_until ON dh114_business(valid_until);
CREATE INDEX IF NOT EXISTS idx_dh114_business_deleted_at ON dh114_business(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_dh114_business_dh114_id ON dh114_business(dh114_id);

COMMENT ON TABLE dh114_business IS '同城114商户详情表（1:1 关联 dh114s）';
COMMENT ON COLUMN dh114_business.verification_status IS '认证状态：0待审 1通过 2拒绝 3过期';

-- ============================================================
-- 3. dh114_business_hours 营业时间表（按周设置）
-- ============================================================
CREATE TABLE IF NOT EXISTS dh114_business_hours (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    dh114_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    weekday INT NOT NULL DEFAULT 1,                -- 1-7（周一到周日）
    open_time VARCHAR(8) NOT NULL DEFAULT '09:00',
    close_time VARCHAR(8) NOT NULL DEFAULT '22:00',
    is_open BOOLEAN NOT NULL DEFAULT TRUE,
    is_24h BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_dh114_business_hours_dh114_id ON dh114_business_hours(dh114_id);
CREATE INDEX IF NOT EXISTS idx_dh114_business_hours_business_id ON dh114_business_hours(business_id);
CREATE INDEX IF NOT EXISTS idx_dh114_business_hours_is_open ON dh114_business_hours(is_open);
CREATE INDEX IF NOT EXISTS idx_dh114_business_hours_deleted_at ON dh114_business_hours(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_dh114_business_hours_dh114_weekday ON dh114_business_hours(dh114_id, weekday);

COMMENT ON TABLE dh114_business_hours IS '同城114商户营业时间表（按周设置）';
COMMENT ON COLUMN dh114_business_hours.weekday IS '星期几 1-7（周一到周日）';

-- ============================================================
-- 4. dh114_categories 商家分类表（全局，无 region_id）
-- ============================================================
CREATE TABLE IF NOT EXISTS dh114_categories (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    name VARCHAR(64) NOT NULL,
    code VARCHAR(64) NOT NULL,
    parent_id BIGINT NOT NULL DEFAULT 0,
    level INT NOT NULL DEFAULT 1,
    business_type VARCHAR(32) NOT NULL DEFAULT 'other',
    icon VARCHAR(64) NOT NULL DEFAULT '',
    color VARCHAR(32) NOT NULL DEFAULT '',
    description VARCHAR(500) NOT NULL DEFAULT '',
    sort INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 1,                -- 0草稿 1已发布 2下架
    business_count INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_dh114_categories_parent_id ON dh114_categories(parent_id);
CREATE INDEX IF NOT EXISTS idx_dh114_categories_level ON dh114_categories(level);
CREATE INDEX IF NOT EXISTS idx_dh114_categories_business_type ON dh114_categories(business_type);
CREATE INDEX IF NOT EXISTS idx_dh114_categories_sort ON dh114_categories(sort);
CREATE INDEX IF NOT EXISTS idx_dh114_categories_status ON dh114_categories(status);
CREATE INDEX IF NOT EXISTS idx_dh114_categories_deleted_at ON dh114_categories(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_dh114_categories_code ON dh114_categories(code);

COMMENT ON TABLE dh114_categories IS '同城114商家分类表（全局）';
COMMENT ON COLUMN dh114_categories.status IS '状态：0草稿 1已发布 2下架';

-- ============================================================
-- 5. dh114_images 商户图片表
-- ============================================================
CREATE TABLE IF NOT EXISTS dh114_images (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    dh114_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    image_type VARCHAR(32) NOT NULL DEFAULT 'other',  -- cover/environment/dish/license/other
    url VARCHAR(500) NOT NULL,
    thumbnail VARCHAR(500) NOT NULL DEFAULT '',
    title VARCHAR(128) NOT NULL DEFAULT '',
    description VARCHAR(500) NOT NULL DEFAULT '',
    sort INT NOT NULL DEFAULT 0,
    is_cover BOOLEAN NOT NULL DEFAULT FALSE,
    width INT NOT NULL DEFAULT 0,
    height INT NOT NULL DEFAULT 0,
    size BIGINT NOT NULL DEFAULT 0,
    tag VARCHAR(64) NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_dh114_images_dh114_id ON dh114_images(dh114_id);
CREATE INDEX IF NOT EXISTS idx_dh114_images_business_id ON dh114_images(business_id);
CREATE INDEX IF NOT EXISTS idx_dh114_images_image_type ON dh114_images(image_type);
CREATE INDEX IF NOT EXISTS idx_dh114_images_sort ON dh114_images(sort);
CREATE INDEX IF NOT EXISTS idx_dh114_images_is_cover ON dh114_images(is_cover);
CREATE INDEX IF NOT EXISTS idx_dh114_images_deleted_at ON dh114_images(deleted_at);

COMMENT ON TABLE dh114_images IS '同城114商户图片表';
COMMENT ON COLUMN dh114_images.image_type IS '图片类型：cover/environment/dish/license/other';

-- ============================================================
-- 6. dh114_tags 标签表（全局，无 region_id）
-- ============================================================
CREATE TABLE IF NOT EXISTS dh114_tags (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    name VARCHAR(64) NOT NULL,
    code VARCHAR(64) NOT NULL DEFAULT '',
    tag_type VARCHAR(32) NOT NULL DEFAULT 'business',  -- business/review/food/service
    parent_id BIGINT NOT NULL DEFAULT 0,
    level INT NOT NULL DEFAULT 1,
    icon VARCHAR(64) NOT NULL DEFAULT '',
    color VARCHAR(32) NOT NULL DEFAULT '',
    sort INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 1,                  -- 0禁用 1启用
    use_count INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_dh114_tags_code ON dh114_tags(code);
CREATE INDEX IF NOT EXISTS idx_dh114_tags_tag_type ON dh114_tags(tag_type);
CREATE INDEX IF NOT EXISTS idx_dh114_tags_parent_id ON dh114_tags(parent_id);
CREATE INDEX IF NOT EXISTS idx_dh114_tags_sort ON dh114_tags(sort);
CREATE INDEX IF NOT EXISTS idx_dh114_tags_status ON dh114_tags(status);
CREATE INDEX IF NOT EXISTS idx_dh114_tags_deleted_at ON dh114_tags(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_dh114_tags_name_type ON dh114_tags(name, tag_type);

COMMENT ON TABLE dh114_tags IS '同城114标签表（商户/评价/美食/服务）';
COMMENT ON COLUMN dh114_tags.tag_type IS '标签类型：business/review/food/service';
COMMENT ON COLUMN dh114_tags.status IS '状态：0禁用 1启用';

-- ============================================================
-- 7. dh114_menus 菜单/服务项目表
-- ============================================================
CREATE TABLE IF NOT EXISTS dh114_menus (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,
    dh114_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    menu_type VARCHAR(32) NOT NULL DEFAULT 'dish',  -- dish/service
    name VARCHAR(128) NOT NULL,
    description VARCHAR(500) NOT NULL DEFAULT '',
    price DECIMAL(12,2) NOT NULL DEFAULT 0,
    original_price DECIMAL(12,2) NOT NULL DEFAULT 0,
    image VARCHAR(255) NOT NULL DEFAULT '',
    unit VARCHAR(32) NOT NULL DEFAULT '',
    sort INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 1,                  -- 0下架 1上架
    order_count INT NOT NULL DEFAULT 0,
    tags JSONB,
    is_signature BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_dh114_menus_region_id ON dh114_menus(region_id);
CREATE INDEX IF NOT EXISTS idx_dh114_menus_dh114_id ON dh114_menus(dh114_id);
CREATE INDEX IF NOT EXISTS idx_dh114_menus_business_id ON dh114_menus(business_id);
CREATE INDEX IF NOT EXISTS idx_dh114_menus_menu_type ON dh114_menus(menu_type);
CREATE INDEX IF NOT EXISTS idx_dh114_menus_price ON dh114_menus(price);
CREATE INDEX IF NOT EXISTS idx_dh114_menus_sort ON dh114_menus(sort);
CREATE INDEX IF NOT EXISTS idx_dh114_menus_status ON dh114_menus(status);
CREATE INDEX IF NOT EXISTS idx_dh114_menus_is_signature ON dh114_menus(is_signature);
CREATE INDEX IF NOT EXISTS idx_dh114_menus_deleted_at ON dh114_menus(deleted_at);

COMMENT ON TABLE dh114_menus IS '同城114菜单/服务项目表';
COMMENT ON COLUMN dh114_menus.menu_type IS '类型：dish菜品/service服务';
COMMENT ON COLUMN dh114_menus.status IS '状态：0下架 1上架';

-- ============================================================
-- 8. dh114_coupons 优惠券表
-- ============================================================
CREATE TABLE IF NOT EXISTS dh114_coupons (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,
    coupon_no VARCHAR(64) NOT NULL,
    dh114_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,

    -- 基本信息
    title VARCHAR(200) NOT NULL,
    description TEXT,
    cover_image VARCHAR(255) NOT NULL DEFAULT '',

    -- 类型与面值
    coupon_type VARCHAR(32) NOT NULL DEFAULT 'discount',  -- discount/full_reduction/cash/gift
    face_value DECIMAL(12,2) NOT NULL DEFAULT 0,
    threshold DECIMAL(12,2) NOT NULL DEFAULT 0,
    discount DECIMAL(3,2) NOT NULL DEFAULT 0,
    max_discount DECIMAL(12,2) NOT NULL DEFAULT 0,

    -- 库存
    total_count INT NOT NULL DEFAULT 0,
    issued_count INT NOT NULL DEFAULT 0,
    used_count INT NOT NULL DEFAULT 0,
    per_user_limit INT NOT NULL DEFAULT 1,

    -- 时间
    start_time TIMESTAMPTZ,
    end_time TIMESTAMPTZ,
    valid_start DATE,
    valid_end DATE,
    valid_days INT NOT NULL DEFAULT 0,

    -- 使用规则
    use_instructions JSONB,
    use_threshold DECIMAL(12,2) NOT NULL DEFAULT 0,

    -- 状态
    status INT NOT NULL DEFAULT 0,                   -- 0草稿 1已发布 2已抢完 3已下架 4已过期
    audit_status INT NOT NULL DEFAULT 0,
    audit_reason VARCHAR(500) NOT NULL DEFAULT '',
    published_at TIMESTAMPTZ,

    -- 运营
    featured BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_dh114_coupons_region_id ON dh114_coupons(region_id);
CREATE INDEX IF NOT EXISTS idx_dh114_coupons_dh114_id ON dh114_coupons(dh114_id);
CREATE INDEX IF NOT EXISTS idx_dh114_coupons_business_id ON dh114_coupons(business_id);
CREATE INDEX IF NOT EXISTS idx_dh114_coupons_coupon_type ON dh114_coupons(coupon_type);
CREATE INDEX IF NOT EXISTS idx_dh114_coupons_issued_count ON dh114_coupons(issued_count);
CREATE INDEX IF NOT EXISTS idx_dh114_coupons_used_count ON dh114_coupons(used_count);
CREATE INDEX IF NOT EXISTS idx_dh114_coupons_start_time ON dh114_coupons(start_time);
CREATE INDEX IF NOT EXISTS idx_dh114_coupons_end_time ON dh114_coupons(end_time);
CREATE INDEX IF NOT EXISTS idx_dh114_coupons_valid_start ON dh114_coupons(valid_start);
CREATE INDEX IF NOT EXISTS idx_dh114_coupons_status ON dh114_coupons(status);
CREATE INDEX IF NOT EXISTS idx_dh114_coupons_audit_status ON dh114_coupons(audit_status);
CREATE INDEX IF NOT EXISTS idx_dh114_coupons_published_at ON dh114_coupons(published_at);
CREATE INDEX IF NOT EXISTS idx_dh114_coupons_featured ON dh114_coupons(featured) WHERE featured = TRUE;
CREATE INDEX IF NOT EXISTS idx_dh114_coupons_deleted_at ON dh114_coupons(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_dh114_coupons_no ON dh114_coupons(coupon_no);

COMMENT ON TABLE dh114_coupons IS '同城114优惠券表';
COMMENT ON COLUMN dh114_coupons.coupon_type IS '类型：discount折扣/full_reduction满减/cash代金券/gift礼品券';
COMMENT ON COLUMN dh114_coupons.status IS '状态：0草稿 1已发布 2已抢完 3已下架 4已过期';

-- ============================================================
-- 9. dh114_groupbuys 团购表
-- ============================================================
CREATE TABLE IF NOT EXISTS dh114_groupbuys (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,
    groupbuy_no VARCHAR(64) NOT NULL,
    dh114_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,

    -- 基本信息
    title VARCHAR(200) NOT NULL,
    description TEXT,
    cover_image VARCHAR(255) NOT NULL DEFAULT '',
    images JSONB,

    -- 价格
    original_price DECIMAL(12,2) NOT NULL DEFAULT 0,
    groupbuy_price DECIMAL(12,2) NOT NULL DEFAULT 0,
    discount DECIMAL(3,2) NOT NULL DEFAULT 0,

    -- 库存
    total_count INT NOT NULL DEFAULT 0,
    sold_count INT NOT NULL DEFAULT 0,
    per_user_limit INT NOT NULL DEFAULT 0,

    -- 时间
    start_time TIMESTAMPTZ,
    end_time TIMESTAMPTZ,
    valid_start DATE,
    valid_end DATE,
    valid_weekdays JSONB,

    -- 使用规则
    use_instructions JSONB,
    use_time_ranges JSONB,
    need_reservation BOOLEAN NOT NULL DEFAULT FALSE,

    -- 互动
    view_count INT NOT NULL DEFAULT 0,
    fav_count INT NOT NULL DEFAULT 0,

    -- 状态
    status INT NOT NULL DEFAULT 0,                   -- 0草稿 1已发布 2已售罄 3已下架 4已过期
    audit_status INT NOT NULL DEFAULT 0,
    audit_reason VARCHAR(500) NOT NULL DEFAULT '',
    published_at TIMESTAMPTZ,

    -- 运营
    featured BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_dh114_groupbuys_region_id ON dh114_groupbuys(region_id);
CREATE INDEX IF NOT EXISTS idx_dh114_groupbuys_dh114_id ON dh114_groupbuys(dh114_id);
CREATE INDEX IF NOT EXISTS idx_dh114_groupbuys_business_id ON dh114_groupbuys(business_id);
CREATE INDEX IF NOT EXISTS idx_dh114_groupbuys_original_price ON dh114_groupbuys(original_price);
CREATE INDEX IF NOT EXISTS idx_dh114_groupbuys_groupbuy_price ON dh114_groupbuys(groupbuy_price);
CREATE INDEX IF NOT EXISTS idx_dh114_groupbuys_sold_count ON dh114_groupbuys(sold_count);
CREATE INDEX IF NOT EXISTS idx_dh114_groupbuys_start_time ON dh114_groupbuys(start_time);
CREATE INDEX IF NOT EXISTS idx_dh114_groupbuys_end_time ON dh114_groupbuys(end_time);
CREATE INDEX IF NOT EXISTS idx_dh114_groupbuys_valid_start ON dh114_groupbuys(valid_start);
CREATE INDEX IF NOT EXISTS idx_dh114_groupbuys_status ON dh114_groupbuys(status);
CREATE INDEX IF NOT EXISTS idx_dh114_groupbuys_audit_status ON dh114_groupbuys(audit_status);
CREATE INDEX IF NOT EXISTS idx_dh114_groupbuys_published_at ON dh114_groupbuys(published_at);
CREATE INDEX IF NOT EXISTS idx_dh114_groupbuys_featured ON dh114_groupbuys(featured) WHERE featured = TRUE;
CREATE INDEX IF NOT EXISTS idx_dh114_groupbuys_deleted_at ON dh114_groupbuys(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_dh114_groupbuys_no ON dh114_groupbuys(groupbuy_no);

COMMENT ON TABLE dh114_groupbuys IS '同城114团购表（限时抢购）';
COMMENT ON COLUMN dh114_groupbuys.status IS '状态：0草稿 1已发布 2已售罄 3已下架 4已过期';

-- ============================================================
-- 10. dh114_reviews 评价表
-- ============================================================
CREATE TABLE IF NOT EXISTS dh114_reviews (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,
    review_no VARCHAR(64) NOT NULL,
    dh114_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,

    -- 评价人
    reviewer_id BIGINT NOT NULL,
    reviewer_name VARCHAR(50) NOT NULL DEFAULT '',
    reviewer_avatar VARCHAR(255) NOT NULL DEFAULT '',
    reviewer_phone VARCHAR(20) NOT NULL DEFAULT '',

    -- 评分
    rating INT NOT NULL DEFAULT 5,                   -- 综合评分 1-5
    taste_rating INT NOT NULL DEFAULT 5,
    service_rating INT NOT NULL DEFAULT 5,
    environment_rating INT NOT NULL DEFAULT 5,

    -- 内容
    content TEXT,
    images JSONB,
    video_url VARCHAR(255) NOT NULL DEFAULT '',
    video_cover VARCHAR(255) NOT NULL DEFAULT '',
    tags JSONB,

    -- 商家回复（冗余字段，子表 dh114_review_replies 存多条）
    reply TEXT,
    replied_at TIMESTAMPTZ,
    has_reply BOOLEAN NOT NULL DEFAULT FALSE,

    -- 互动
    like_count INT NOT NULL DEFAULT 0,
    liked_by JSONB,

    -- 状态
    status INT NOT NULL DEFAULT 0,                    -- 0待审 1通过 2拒绝 3隐藏
    audit_status INT NOT NULL DEFAULT 0,
    audit_reason VARCHAR(500) NOT NULL DEFAULT '',

    -- 关联订单
    order_id BIGINT NOT NULL DEFAULT 0,
    consumed_at DATE,

    -- 评价类型
    review_type VARCHAR(32) NOT NULL DEFAULT 'general'  -- general/order/visit
);
CREATE INDEX IF NOT EXISTS idx_dh114_reviews_region_id ON dh114_reviews(region_id);
CREATE INDEX IF NOT EXISTS idx_dh114_reviews_dh114_id ON dh114_reviews(dh114_id);
CREATE INDEX IF NOT EXISTS idx_dh114_reviews_business_id ON dh114_reviews(business_id);
CREATE INDEX IF NOT EXISTS idx_dh114_reviews_reviewer_id ON dh114_reviews(reviewer_id);
CREATE INDEX IF NOT EXISTS idx_dh114_reviews_rating ON dh114_reviews(rating);
CREATE INDEX IF NOT EXISTS idx_dh114_reviews_replied_at ON dh114_reviews(replied_at);
CREATE INDEX IF NOT EXISTS idx_dh114_reviews_has_reply ON dh114_reviews(has_reply);
CREATE INDEX IF NOT EXISTS idx_dh114_reviews_status ON dh114_reviews(status);
CREATE INDEX IF NOT EXISTS idx_dh114_reviews_audit_status ON dh114_reviews(audit_status);
CREATE INDEX IF NOT EXISTS idx_dh114_reviews_order_id ON dh114_reviews(order_id);
CREATE INDEX IF NOT EXISTS idx_dh114_reviews_review_type ON dh114_reviews(review_type);
CREATE INDEX IF NOT EXISTS idx_dh114_reviews_deleted_at ON dh114_reviews(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_dh114_reviews_no ON dh114_reviews(review_no);

COMMENT ON TABLE dh114_reviews IS '同城114商户评价表';
COMMENT ON COLUMN dh114_reviews.rating IS '综合评分 1-5';
COMMENT ON COLUMN dh114_reviews.status IS '状态：0待审 1通过 2拒绝 3隐藏';
COMMENT ON COLUMN dh114_reviews.review_type IS '类型：general/order/visit';

-- ============================================================
-- 11. dh114_review_replies 商家回复评价表
-- ============================================================
CREATE TABLE IF NOT EXISTS dh114_review_replies (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    review_id BIGINT NOT NULL,
    dh114_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    replier_id BIGINT NOT NULL,
    replier_name VARCHAR(50) NOT NULL DEFAULT '',
    replier_avatar VARCHAR(255) NOT NULL DEFAULT '',
    replier_type VARCHAR(16) NOT NULL DEFAULT 'merchant',  -- merchant/user/admin
    content TEXT,
    images JSONB,
    parent_id BIGINT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 1,                    -- 0隐藏 1显示
    like_count INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_dh114_review_replies_review_id ON dh114_review_replies(review_id);
CREATE INDEX IF NOT EXISTS idx_dh114_review_replies_dh114_id ON dh114_review_replies(dh114_id);
CREATE INDEX IF NOT EXISTS idx_dh114_review_replies_business_id ON dh114_review_replies(business_id);
CREATE INDEX IF NOT EXISTS idx_dh114_review_replies_replier_id ON dh114_review_replies(replier_id);
CREATE INDEX IF NOT EXISTS idx_dh114_review_replies_replier_type ON dh114_review_replies(replier_type);
CREATE INDEX IF NOT EXISTS idx_dh114_review_replies_parent_id ON dh114_review_replies(parent_id);
CREATE INDEX IF NOT EXISTS idx_dh114_review_replies_status ON dh114_review_replies(status);
CREATE INDEX IF NOT EXISTS idx_dh114_review_replies_deleted_at ON dh114_review_replies(deleted_at);

COMMENT ON TABLE dh114_review_replies IS '同城114商家回复评价表（支持追问/追答）';
COMMENT ON COLUMN dh114_review_replies.replier_type IS '回复人类型：merchant/user/admin';
COMMENT ON COLUMN dh114_review_replies.status IS '状态：0隐藏 1显示';

-- ============================================================
-- 12. dh114_favorites 收藏表
-- ============================================================
CREATE TABLE IF NOT EXISTS dh114_favorites (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id BIGINT NOT NULL,
    dh114_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    favorite_type VARCHAR(32) NOT NULL DEFAULT 'business',  -- business/groupbuy/coupon
    group_id BIGINT NOT NULL DEFAULT 0,
    remark VARCHAR(200) NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_dh114_favorites_user_id ON dh114_favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_dh114_favorites_dh114_id ON dh114_favorites(dh114_id);
CREATE INDEX IF NOT EXISTS idx_dh114_favorites_business_id ON dh114_favorites(business_id);
CREATE INDEX IF NOT EXISTS idx_dh114_favorites_favorite_type ON dh114_favorites(favorite_type);
CREATE INDEX IF NOT EXISTS idx_dh114_favorites_group_id ON dh114_favorites(group_id);
CREATE INDEX IF NOT EXISTS idx_dh114_favorites_deleted_at ON dh114_favorites(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_dh114_favorites_user_target_type ON dh114_favorites(user_id, dh114_id, favorite_type);

COMMENT ON TABLE dh114_favorites IS '同城114收藏表';
COMMENT ON COLUMN dh114_favorites.favorite_type IS '收藏类型：business/groupbuy/coupon';

-- ============================================================
-- 13. dh114_phone_calls 电话拨打记录表
-- ============================================================
CREATE TABLE IF NOT EXISTS dh114_phone_calls (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,
    call_no VARCHAR(64) NOT NULL,
    dh114_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    phone VARCHAR(32) NOT NULL DEFAULT '',

    -- 主叫
    caller_id BIGINT NOT NULL DEFAULT 0,
    caller_phone VARCHAR(20) NOT NULL DEFAULT '',
    caller_name VARCHAR(50) NOT NULL DEFAULT '',

    -- 拨打信息
    call_type VARCHAR(16) NOT NULL DEFAULT 'click',  -- click/call
    device VARCHAR(32) NOT NULL DEFAULT '',
    ip VARCHAR(64) NOT NULL DEFAULT '',
    user_agent VARCHAR(500) NOT NULL DEFAULT '',

    -- 结果
    status VARCHAR(16) NOT NULL DEFAULT 'success',   -- success/failed
    duration INT NOT NULL DEFAULT 0,
    called_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_dh114_phone_calls_region_id ON dh114_phone_calls(region_id);
CREATE INDEX IF NOT EXISTS idx_dh114_phone_calls_dh114_id ON dh114_phone_calls(dh114_id);
CREATE INDEX IF NOT EXISTS idx_dh114_phone_calls_business_id ON dh114_phone_calls(business_id);
CREATE INDEX IF NOT EXISTS idx_dh114_phone_calls_phone ON dh114_phone_calls(phone);
CREATE INDEX IF NOT EXISTS idx_dh114_phone_calls_caller_id ON dh114_phone_calls(caller_id);
CREATE INDEX IF NOT EXISTS idx_dh114_phone_calls_call_type ON dh114_phone_calls(call_type);
CREATE INDEX IF NOT EXISTS idx_dh114_phone_calls_status ON dh114_phone_calls(status);
CREATE INDEX IF NOT EXISTS idx_dh114_phone_calls_called_at ON dh114_phone_calls(called_at);
CREATE INDEX IF NOT EXISTS idx_dh114_phone_calls_deleted_at ON dh114_phone_calls(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_dh114_phone_calls_no ON dh114_phone_calls(call_no);

COMMENT ON TABLE dh114_phone_calls IS '同城114电话拨打记录表';
COMMENT ON COLUMN dh114_phone_calls.call_type IS '拨打类型：click点击/call直接拨打';
COMMENT ON COLUMN dh114_phone_calls.status IS '结果：success/failed';

-- ============================================================
-- 14. dh114_visits 浏览记录表
-- ============================================================
CREATE TABLE IF NOT EXISTS dh114_visits (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,
    user_id BIGINT NOT NULL DEFAULT 0,
    dh114_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    visit_type VARCHAR(32) NOT NULL DEFAULT 'business',  -- business/groupbuy/coupon
    ip VARCHAR(64) NOT NULL DEFAULT '',
    user_agent VARCHAR(500) NOT NULL DEFAULT '',
    referer VARCHAR(500) NOT NULL DEFAULT '',
    device VARCHAR(32) NOT NULL DEFAULT '',
    source VARCHAR(32) NOT NULL DEFAULT '',          -- search/category/recommend/direct/share
    duration INT NOT NULL DEFAULT 0,
    longitude DECIMAL(10,7) NOT NULL DEFAULT 0,
    latitude DECIMAL(10,7) NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_dh114_visits_region_id ON dh114_visits(region_id);
CREATE INDEX IF NOT EXISTS idx_dh114_visits_user_id ON dh114_visits(user_id);
CREATE INDEX IF NOT EXISTS idx_dh114_visits_dh114_id ON dh114_visits(dh114_id);
CREATE INDEX IF NOT EXISTS idx_dh114_visits_business_id ON dh114_visits(business_id);
CREATE INDEX IF NOT EXISTS idx_dh114_visits_visit_type ON dh114_visits(visit_type);
CREATE INDEX IF NOT EXISTS idx_dh114_visits_device ON dh114_visits(device);
CREATE INDEX IF NOT EXISTS idx_dh114_visits_source ON dh114_visits(source);
CREATE INDEX IF NOT EXISTS idx_dh114_visits_created_at ON dh114_visits(created_at);
CREATE INDEX IF NOT EXISTS idx_dh114_visits_deleted_at ON dh114_visits(deleted_at);

COMMENT ON TABLE dh114_visits IS '同城114浏览记录表';
COMMENT ON COLUMN dh114_visits.visit_type IS '浏览类型：business/groupbuy/coupon';
COMMENT ON COLUMN dh114_visits.source IS '来源：search/category/recommend/direct/share';

-- ============================================================
-- 15. dh114_recommendations 推荐商家表
-- ============================================================
CREATE TABLE IF NOT EXISTS dh114_recommendations (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,
    user_id BIGINT NOT NULL DEFAULT 0,                -- 推荐给的用户 ID（0 表示全员）
    dh114_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    recommend_type VARCHAR(32) NOT NULL DEFAULT 'home',  -- home/category/nearby/personalized
    position INT NOT NULL DEFAULT 0,
    score DECIMAL(5,2) NOT NULL DEFAULT 0,
    reason VARCHAR(500) NOT NULL DEFAULT '',
    category_id BIGINT,
    expire_at TIMESTAMPTZ,

    -- 状态
    status INT NOT NULL DEFAULT 0,                    -- 0已展示 1已点击 2已联系 3已忽略
    clicked_at TIMESTAMPTZ,
    contacted_at TIMESTAMPTZ,
    dismissed_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_dh114_recommendations_region_id ON dh114_recommendations(region_id);
CREATE INDEX IF NOT EXISTS idx_dh114_recommendations_user_id ON dh114_recommendations(user_id);
CREATE INDEX IF NOT EXISTS idx_dh114_recommendations_dh114_id ON dh114_recommendations(dh114_id);
CREATE INDEX IF NOT EXISTS idx_dh114_recommendations_business_id ON dh114_recommendations(business_id);
CREATE INDEX IF NOT EXISTS idx_dh114_recommendations_recommend_type ON dh114_recommendations(recommend_type);
CREATE INDEX IF NOT EXISTS idx_dh114_recommendations_position ON dh114_recommendations(position);
CREATE INDEX IF NOT EXISTS idx_dh114_recommendations_category_id ON dh114_recommendations(category_id);
CREATE INDEX IF NOT EXISTS idx_dh114_recommendations_expire_at ON dh114_recommendations(expire_at);
CREATE INDEX IF NOT EXISTS idx_dh114_recommendations_status ON dh114_recommendations(status);
CREATE INDEX IF NOT EXISTS idx_dh114_recommendations_clicked_at ON dh114_recommendations(clicked_at);
CREATE INDEX IF NOT EXISTS idx_dh114_recommendations_contacted_at ON dh114_recommendations(contacted_at);
CREATE INDEX IF NOT EXISTS idx_dh114_recommendations_dismissed_at ON dh114_recommendations(dismissed_at);
CREATE INDEX IF NOT EXISTS idx_dh114_recommendations_deleted_at ON dh114_recommendations(deleted_at);

COMMENT ON TABLE dh114_recommendations IS '同城114推荐商家表';
COMMENT ON COLUMN dh114_recommendations.recommend_type IS '推荐类型：home/category/nearby/personalized';
COMMENT ON COLUMN dh114_recommendations.status IS '状态：0已展示 1已点击 2已联系 3已忽略';

-- ============================================================
-- 16. dh114_statistics 统计表
-- ============================================================
CREATE TABLE IF NOT EXISTS dh114_statistics (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,
    stat_date DATE NOT NULL,
    stat_type VARCHAR(32) NOT NULL DEFAULT 'daily',  -- daily/business/category
    dh114_id BIGINT NOT NULL DEFAULT 0,
    business_id BIGINT NOT NULL DEFAULT 0,
    category_id BIGINT,

    -- 互动统计
    view_count BIGINT NOT NULL DEFAULT 0,
    fav_count BIGINT NOT NULL DEFAULT 0,
    call_count BIGINT NOT NULL DEFAULT 0,
    share_count BIGINT NOT NULL DEFAULT 0,
    contact_count BIGINT NOT NULL DEFAULT 0,
    visit_count BIGINT NOT NULL DEFAULT 0,

    -- 评价统计
    review_count BIGINT NOT NULL DEFAULT 0,
    new_review_count BIGINT NOT NULL DEFAULT 0,
    avg_rating DECIMAL(3,2) NOT NULL DEFAULT 0,
    good_rate DECIMAL(5,2) NOT NULL DEFAULT 0,

    -- 交易统计
    groupbuy_sold BIGINT NOT NULL DEFAULT 0,
    groupbuy_amount DECIMAL(14,2) NOT NULL DEFAULT 0,
    coupon_issued BIGINT NOT NULL DEFAULT 0,
    coupon_used BIGINT NOT NULL DEFAULT 0,
    order_count BIGINT NOT NULL DEFAULT 0,
    order_amount DECIMAL(14,2) NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_dh114_statistics_region_id ON dh114_statistics(region_id);
CREATE INDEX IF NOT EXISTS idx_dh114_statistics_stat_date ON dh114_statistics(stat_date);
CREATE INDEX IF NOT EXISTS idx_dh114_statistics_stat_type ON dh114_statistics(stat_type);
CREATE INDEX IF NOT EXISTS idx_dh114_statistics_dh114_id ON dh114_statistics(dh114_id);
CREATE INDEX IF NOT EXISTS idx_dh114_statistics_business_id ON dh114_statistics(business_id);
CREATE INDEX IF NOT EXISTS idx_dh114_statistics_category_id ON dh114_statistics(category_id);
CREATE INDEX IF NOT EXISTS idx_dh114_statistics_deleted_at ON dh114_statistics(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_dh114_statistics_date_type_target ON dh114_statistics(stat_date, stat_type, dh114_id);

COMMENT ON TABLE dh114_statistics IS '同城114统计表（日统计/商户统计/分类统计）';
COMMENT ON COLUMN dh114_statistics.stat_type IS '统计类型：daily/business/category';

-- ============================================================
-- 17. dh114_audit_rules 审核规则表（全局，无 region_id）
-- ============================================================
CREATE TABLE IF NOT EXISTS dh114_audit_rules (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    rule_name VARCHAR(128) NOT NULL,
    rule_type VARCHAR(32) NOT NULL,                  -- sensitive_word/prohibited/contact/price_check/frequency
    rule_key VARCHAR(64) NOT NULL DEFAULT '',
    pattern TEXT,
    threshold JSONB,
    action VARCHAR(32) NOT NULL DEFAULT 'reject',    -- reject/approval/filter/limit
    penalty_type VARCHAR(32) NOT NULL DEFAULT '',    -- warning/ban24h/ban7d/ban30d/ban_forever/delete/limit
    severity INT NOT NULL DEFAULT 1,                 -- 严重程度 1-5
    status INT NOT NULL DEFAULT 1,                   -- 0禁用 1启用
    description VARCHAR(500) NOT NULL DEFAULT '',
    sort INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_dh114_audit_rules_rule_type ON dh114_audit_rules(rule_type);
CREATE INDEX IF NOT EXISTS idx_dh114_audit_rules_rule_key ON dh114_audit_rules(rule_key);
CREATE INDEX IF NOT EXISTS idx_dh114_audit_rules_action ON dh114_audit_rules(action);
CREATE INDEX IF NOT EXISTS idx_dh114_audit_rules_severity ON dh114_audit_rules(severity);
CREATE INDEX IF NOT EXISTS idx_dh114_audit_rules_status ON dh114_audit_rules(status);
CREATE INDEX IF NOT EXISTS idx_dh114_audit_rules_sort ON dh114_audit_rules(sort);
CREATE INDEX IF NOT EXISTS idx_dh114_audit_rules_deleted_at ON dh114_audit_rules(deleted_at);

COMMENT ON TABLE dh114_audit_rules IS '同城114审核规则表（全局）';
COMMENT ON COLUMN dh114_audit_rules.rule_type IS '规则类型：sensitive_word/prohibited/contact/price_check/frequency';
COMMENT ON COLUMN dh114_audit_rules.action IS '动作：reject/approval/filter/limit';
COMMENT ON COLUMN dh114_audit_rules.status IS '状态：0禁用 1启用';

-- ============================================================
-- 18. dh114_verifications 商户认证表
-- ============================================================
CREATE TABLE IF NOT EXISTS dh114_verifications (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,
    verification_no VARCHAR(64) NOT NULL,
    dh114_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    user_id BIGINT NOT NULL,

    -- 认证类型
    verification_type VARCHAR(32) NOT NULL DEFAULT 'business_license',  -- business_license/field/brand/license_field

    -- 营业执照信息
    license_no VARCHAR(64) NOT NULL DEFAULT '',
    license_image VARCHAR(255) NOT NULL DEFAULT '',
    business_name VARCHAR(200) NOT NULL DEFAULT '',
    legal_person VARCHAR(64) NOT NULL DEFAULT '',
    legal_person_id_card VARCHAR(32) NOT NULL DEFAULT '',
    business_scope TEXT,
    registered_address VARCHAR(500) NOT NULL DEFAULT '',

    -- 实地认证信息
    field_photos JSONB,
    field_address VARCHAR(500) NOT NULL DEFAULT '',
    field_longitude DECIMAL(10,7) NOT NULL DEFAULT 0,
    field_latitude DECIMAL(10,7) NOT NULL DEFAULT 0,
    field_date DATE,
    inspector_id BIGINT NOT NULL DEFAULT 0,
    inspector_name VARCHAR(50) NOT NULL DEFAULT '',

    -- 品牌授权信息
    brand_name VARCHAR(128) NOT NULL DEFAULT '',
    brand_auth_image VARCHAR(255) NOT NULL DEFAULT '',

    -- 状态
    status INT NOT NULL DEFAULT 0,                    -- 0待审 1通过 2拒绝 3过期
    audit_remark VARCHAR(500) NOT NULL DEFAULT '',
    audited_by BIGINT NOT NULL DEFAULT 0,
    audited_at TIMESTAMPTZ,
    valid_until DATE
);
CREATE INDEX IF NOT EXISTS idx_dh114_verifications_region_id ON dh114_verifications(region_id);
CREATE INDEX IF NOT EXISTS idx_dh114_verifications_dh114_id ON dh114_verifications(dh114_id);
CREATE INDEX IF NOT EXISTS idx_dh114_verifications_business_id ON dh114_verifications(business_id);
CREATE INDEX IF NOT EXISTS idx_dh114_verifications_user_id ON dh114_verifications(user_id);
CREATE INDEX IF NOT EXISTS idx_dh114_verifications_verification_type ON dh114_verifications(verification_type);
CREATE INDEX IF NOT EXISTS idx_dh114_verifications_license_no ON dh114_verifications(license_no);
CREATE INDEX IF NOT EXISTS idx_dh114_verifications_status ON dh114_verifications(status);
CREATE INDEX IF NOT EXISTS idx_dh114_verifications_audited_by ON dh114_verifications(audited_by);
CREATE INDEX IF NOT EXISTS idx_dh114_verifications_audited_at ON dh114_verifications(audited_at);
CREATE INDEX IF NOT EXISTS idx_dh114_verifications_valid_until ON dh114_verifications(valid_until);
CREATE INDEX IF NOT EXISTS idx_dh114_verifications_deleted_at ON dh114_verifications(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_dh114_verifications_no ON dh114_verifications(verification_no);

COMMENT ON TABLE dh114_verifications IS '同城114商户认证表（营业执照+实地认证+品牌授权）';
COMMENT ON COLUMN dh114_verifications.verification_type IS '认证类型：business_license/field/brand/license_field';
COMMENT ON COLUMN dh114_verifications.status IS '状态：0待审 1通过 2拒绝 3过期';

-- ============================================================
-- updated_at 触发器（依赖 001_p0_baseline.sql 中的 update_updated_at_column 函数）
-- ============================================================
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'update_updated_at_column') THEN
        DROP TRIGGER IF EXISTS trg_dh114s_updated_at ON dh114s; CREATE TRIGGER trg_dh114s_updated_at BEFORE UPDATE ON dh114s FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_dh114_business_updated_at ON dh114_business; CREATE TRIGGER trg_dh114_business_updated_at BEFORE UPDATE ON dh114_business FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_dh114_business_hours_updated_at ON dh114_business_hours; CREATE TRIGGER trg_dh114_business_hours_updated_at BEFORE UPDATE ON dh114_business_hours FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_dh114_categories_updated_at ON dh114_categories; CREATE TRIGGER trg_dh114_categories_updated_at BEFORE UPDATE ON dh114_categories FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_dh114_images_updated_at ON dh114_images; CREATE TRIGGER trg_dh114_images_updated_at BEFORE UPDATE ON dh114_images FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_dh114_tags_updated_at ON dh114_tags; CREATE TRIGGER trg_dh114_tags_updated_at BEFORE UPDATE ON dh114_tags FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_dh114_menus_updated_at ON dh114_menus; CREATE TRIGGER trg_dh114_menus_updated_at BEFORE UPDATE ON dh114_menus FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_dh114_coupons_updated_at ON dh114_coupons; CREATE TRIGGER trg_dh114_coupons_updated_at BEFORE UPDATE ON dh114_coupons FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_dh114_groupbuys_updated_at ON dh114_groupbuys; CREATE TRIGGER trg_dh114_groupbuys_updated_at BEFORE UPDATE ON dh114_groupbuys FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_dh114_reviews_updated_at ON dh114_reviews; CREATE TRIGGER trg_dh114_reviews_updated_at BEFORE UPDATE ON dh114_reviews FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_dh114_review_replies_updated_at ON dh114_review_replies; CREATE TRIGGER trg_dh114_review_replies_updated_at BEFORE UPDATE ON dh114_review_replies FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_dh114_favorites_updated_at ON dh114_favorites; CREATE TRIGGER trg_dh114_favorites_updated_at BEFORE UPDATE ON dh114_favorites FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_dh114_phone_calls_updated_at ON dh114_phone_calls; CREATE TRIGGER trg_dh114_phone_calls_updated_at BEFORE UPDATE ON dh114_phone_calls FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_dh114_visits_updated_at ON dh114_visits; CREATE TRIGGER trg_dh114_visits_updated_at BEFORE UPDATE ON dh114_visits FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_dh114_recommendations_updated_at ON dh114_recommendations; CREATE TRIGGER trg_dh114_recommendations_updated_at BEFORE UPDATE ON dh114_recommendations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_dh114_statistics_updated_at ON dh114_statistics; CREATE TRIGGER trg_dh114_statistics_updated_at BEFORE UPDATE ON dh114_statistics FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_dh114_audit_rules_updated_at ON dh114_audit_rules; CREATE TRIGGER trg_dh114_audit_rules_updated_at BEFORE UPDATE ON dh114_audit_rules FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_dh114_verifications_updated_at ON dh114_verifications; CREATE TRIGGER trg_dh114_verifications_updated_at BEFORE UPDATE ON dh114_verifications FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;
