-- ============================================================
-- house 房屋租售模块完整功能迁移脚本（v3.2.1）
-- 对标：贝壳 / 链家 / 安居客 / 我爱我家 / 58房产
--
-- 内容：
--   1. ALTER TABLE houses 主表新增 35+ 字段（租金/售价/户型/面积/楼层/朝向/装修/产权/年限/视频/VR/风控/运营）
--   2. CREATE 19 张子表（house_ 前缀，依据数据库分表前缀规范 v1.0.0）
--   3. 索引、外键、触发器、注释
--   4. 全幂等：CREATE TABLE IF NOT EXISTS / ALTER TABLE ADD COLUMN IF NOT EXISTS
-- 依赖：001_p0_baseline.sql（update_updated_at_column 函数已创建）
-- ============================================================

-- ============================================================
-- 第一部分：扩展 houses 主表（保持现有表名兼容已发布数据）
-- 注意：houses 表由 GORM AutoMigrate 在应用启动时创建
--      本迁移对 houses 的所有 ALTER 操作包装在 DO 块中，
--      若表不存在则跳过（待应用启动后再执行一次本迁移即可补齐字段）
-- ============================================================
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'houses') THEN
        -- === 交易类型/发布类型 ===
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS listing_type VARCHAR(16) NOT NULL DEFAULT 'rent';
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS property_type VARCHAR(32) NOT NULL DEFAULT 'residential';
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS source_type VARCHAR(16) NOT NULL DEFAULT 'personal';
        CREATE INDEX IF NOT EXISTS idx_houses_listing_type ON houses(listing_type);
        CREATE INDEX IF NOT EXISTS idx_houses_property_type ON houses(property_type);
        CREATE INDEX IF NOT EXISTS idx_houses_source_type ON houses(source_type);

        -- === 租金相关 ===
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS rent_price DECIMAL(12,2) NOT NULL DEFAULT 0;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS rent_unit VARCHAR(16) NOT NULL DEFAULT 'month';
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS rent_type VARCHAR(16) NOT NULL DEFAULT 'entire';
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS deposit_type VARCHAR(16) NOT NULL DEFAULT 'one_month';
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS payment_method VARCHAR(16) NOT NULL DEFAULT 'monthly';
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS rent_negotiable BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS rent_min_months INT NOT NULL DEFAULT 0;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS rent_max_months INT NOT NULL DEFAULT 0;
        CREATE INDEX IF NOT EXISTS idx_houses_rent_price ON houses(rent_price);
        CREATE INDEX IF NOT EXISTS idx_houses_rent_type ON houses(rent_type);

        -- === 售价相关 ===
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS sale_price DECIMAL(14,2) NOT NULL DEFAULT 0;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS sale_negotiable BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS average_price DECIMAL(10,2) NOT NULL DEFAULT 0;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS original_price DECIMAL(14,2) NOT NULL DEFAULT 0;
        CREATE INDEX IF NOT EXISTS idx_houses_sale_price ON houses(sale_price);
        CREATE INDEX IF NOT EXISTS idx_houses_average_price ON houses(average_price);

        -- === 户型 ===
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS rooms INT NOT NULL DEFAULT 0;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS halls INT NOT NULL DEFAULT 0;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS bathrooms INT NOT NULL DEFAULT 0;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS kitchens INT NOT NULL DEFAULT 0;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS balconies INT NOT NULL DEFAULT 0;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS layout VARCHAR(32) NOT NULL DEFAULT '';
        CREATE INDEX IF NOT EXISTS idx_houses_rooms ON houses(rooms);

        -- === 面积 ===
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS building_area DECIMAL(10,2) NOT NULL DEFAULT 0;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS inner_area DECIMAL(10,2) NOT NULL DEFAULT 0;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS pool_ratio DECIMAL(5,2) NOT NULL DEFAULT 0;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS usable_area DECIMAL(10,2) NOT NULL DEFAULT 0;
        CREATE INDEX IF NOT EXISTS idx_houses_building_area ON houses(building_area);

        -- === 楼层/朝向 ===
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS floor INT NOT NULL DEFAULT 0;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS total_floor INT NOT NULL DEFAULT 0;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS floor_type VARCHAR(16) NOT NULL DEFAULT 'mid';
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS orientation VARCHAR(32) NOT NULL DEFAULT '';
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS has_elevator BOOLEAN NOT NULL DEFAULT FALSE;
        CREATE INDEX IF NOT EXISTS idx_houses_floor ON houses(floor);
        CREATE INDEX IF NOT EXISTS idx_houses_total_floor ON houses(total_floor);
        CREATE INDEX IF NOT EXISTS idx_houses_floor_type ON houses(floor_type);
        CREATE INDEX IF NOT EXISTS idx_houses_orientation ON houses(orientation);
        CREATE INDEX IF NOT EXISTS idx_houses_has_elevator ON houses(has_elevator) WHERE has_elevator = TRUE;

        -- === 装修/产权/年限 ===
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS decoration VARCHAR(16) NOT NULL DEFAULT 'rough';
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS property_ownership VARCHAR(32) NOT NULL DEFAULT 'commercial';
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS property_years INT NOT NULL DEFAULT 70;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS building_year INT NOT NULL DEFAULT 0;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS building_age INT NOT NULL DEFAULT 0;
        CREATE INDEX IF NOT EXISTS idx_houses_decoration ON houses(decoration);
        CREATE INDEX IF NOT EXISTS idx_houses_property_ownership ON houses(property_ownership);
        CREATE INDEX IF NOT EXISTS idx_houses_building_year ON houses(building_year);

        -- === 关联 ID ===
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS community_id BIGINT;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS agent_id BIGINT;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS category_id BIGINT;
        CREATE INDEX IF NOT EXISTS idx_houses_community_id ON houses(community_id);
        CREATE INDEX IF NOT EXISTS idx_houses_agent_id ON houses(agent_id);
        CREATE INDEX IF NOT EXISTS idx_houses_category_id ON houses(category_id);

        -- === 地理位置冗余（小区信息冗余存储以便 LBS 检索） ===
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS city VARCHAR(64) NOT NULL DEFAULT '';
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS district VARCHAR(64) NOT NULL DEFAULT '';
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS business_district VARCHAR(128) NOT NULL DEFAULT '';
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS address VARCHAR(500) NOT NULL DEFAULT '';
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS latitude DECIMAL(10,7) NOT NULL DEFAULT 0;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS longitude DECIMAL(10,7) NOT NULL DEFAULT 0;
        CREATE INDEX IF NOT EXISTS idx_houses_city ON houses(city);
        CREATE INDEX IF NOT EXISTS idx_houses_district ON houses(district);

        -- === 互动统计 ===
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS view_count INT NOT NULL DEFAULT 0;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS fav_count INT NOT NULL DEFAULT 0;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS contact_count INT NOT NULL DEFAULT 0;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS share_count INT NOT NULL DEFAULT 0;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS viewing_count INT NOT NULL DEFAULT 0;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS last_viewing_at TIMESTAMPTZ;

        -- === 风控 ===
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS content_hash VARCHAR(64);
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS risk_score INT NOT NULL DEFAULT 0;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS same_house_id VARCHAR(64);
        CREATE INDEX IF NOT EXISTS idx_houses_content_hash ON houses(content_hash);
        CREATE INDEX IF NOT EXISTS idx_houses_risk_score ON houses(risk_score);
        CREATE INDEX IF NOT EXISTS idx_houses_same_house_id ON houses(same_house_id);

        -- === 视频/VR/全景 ===
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS video_url VARCHAR(255);
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS video_cover VARCHAR(255);
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS vr_url VARCHAR(255);
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS panorama_url VARCHAR(255);

        -- === 配套设施/标签（JSONB） ===
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS facilities JSONB;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS tags JSONB;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS nearby_pois JSONB;
        CREATE INDEX IF NOT EXISTS idx_houses_facilities ON houses USING GIN(facilities);
        CREATE INDEX IF NOT EXISTS idx_houses_tags ON houses USING GIN(tags);

        -- === 运营字段 ===
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS featured BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS picked BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS verified BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS promotion_level INT NOT NULL DEFAULT 0;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS traffic_weight DECIMAL(3,2) NOT NULL DEFAULT 1.00;
        CREATE INDEX IF NOT EXISTS idx_houses_featured ON houses(featured) WHERE featured = TRUE;
        CREATE INDEX IF NOT EXISTS idx_houses_picked ON houses(picked) WHERE picked = TRUE;
        CREATE INDEX IF NOT EXISTS idx_houses_verified ON houses(verified) WHERE verified = TRUE;

        -- === 真房源认证 ===
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS real_house_verified BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE houses ADD COLUMN IF NOT EXISTS real_house_verified_at TIMESTAMPTZ;
        CREATE INDEX IF NOT EXISTS idx_houses_real_house_verified ON houses(real_house_verified) WHERE real_house_verified = TRUE;

        -- 字段注释
        COMMENT ON COLUMN houses.listing_type IS '发布类型：rent/sale/transfer（出租/出售/转让）';
        COMMENT ON COLUMN houses.property_type IS '物业类型：residential/apartment/villa/loft/office/shop（住宅/公寓/别墅/ loft/写字楼/商铺）';
        COMMENT ON COLUMN houses.source_type IS '来源：personal/agent/developer（个人/经纪人/开发商）';
        COMMENT ON COLUMN houses.rent_price IS '租金（单位见 rent_unit）';
        COMMENT ON COLUMN houses.rent_unit IS '租金单位：month/year/day';
        COMMENT ON COLUMN houses.rent_type IS '租赁类型：entire/shared（整租/合租）';
        COMMENT ON COLUMN houses.deposit_type IS '押金类型：none/one_month/two_month/three_month/pay_one_deposit_three';
        COMMENT ON COLUMN houses.payment_method IS '付款方式：monthly/quarterly/half_year/yearly/one_time';
        COMMENT ON COLUMN houses.rent_negotiable IS '租金是否可议';
        COMMENT ON COLUMN houses.rent_min_months IS '最短租期（月），0=不限';
        COMMENT ON COLUMN houses.rent_max_months IS '最长租期（月），0=不限';
        COMMENT ON COLUMN houses.sale_price IS '售价（元）';
        COMMENT ON COLUMN houses.sale_negotiable IS '售价是否可议';
        COMMENT ON COLUMN houses.average_price IS '每平米均价';
        COMMENT ON COLUMN houses.original_price IS '原价/挂牌价';
        COMMENT ON COLUMN houses.rooms IS '室';
        COMMENT ON COLUMN houses.halls IS '厅';
        COMMENT ON COLUMN houses.bathrooms IS '卫';
        COMMENT ON COLUMN houses.kitchens IS '厨';
        COMMENT ON COLUMN houses.balconies IS '阳台数';
        COMMENT ON COLUMN houses.layout IS '户型文本（如 3室2厅2卫）';
        COMMENT ON COLUMN houses.building_area IS '建筑面积（㎡）';
        COMMENT ON COLUMN houses.inner_area IS '套内面积（㎡）';
        COMMENT ON COLUMN houses.pool_ratio IS '公摊比例 0.00-1.00';
        COMMENT ON COLUMN houses.usable_area IS '使用面积（㎡）';
        COMMENT ON COLUMN houses.floor IS '所在楼层';
        COMMENT ON COLUMN houses.total_floor IS '总楼层';
        COMMENT ON COLUMN houses.floor_type IS '楼层段：low/mid/high';
        COMMENT ON COLUMN houses.orientation IS '朝向：east/south/west/north/south_north 等';
        COMMENT ON COLUMN houses.has_elevator IS '是否有电梯';
        COMMENT ON COLUMN houses.decoration IS '装修：rough/simple/fine/luxury（毛坯/简装/精装/豪装）';
        COMMENT ON COLUMN houses.property_ownership IS '产权：commercial/reformed/affordable/small_property（商品房/房改房/经济适用房/小产权）';
        COMMENT ON COLUMN houses.property_years IS '产权年限 70/50/40';
        COMMENT ON COLUMN houses.building_year IS '建造年代（年份）';
        COMMENT ON COLUMN houses.building_age IS '房龄（年）';
        COMMENT ON COLUMN houses.community_id IS '关联小区 ID（house_communities.id）';
        COMMENT ON COLUMN houses.agent_id IS '关联经纪人 ID（house_agents.id）';
        COMMENT ON COLUMN houses.category_id IS '关联房源分类 ID（house_categories.id）';
        COMMENT ON COLUMN houses.city IS '城市';
        COMMENT ON COLUMN houses.district IS '行政区';
        COMMENT ON COLUMN houses.business_district IS '商圈';
        COMMENT ON COLUMN houses.address IS '详细地址';
        COMMENT ON COLUMN houses.latitude IS '纬度';
        COMMENT ON COLUMN houses.longitude IS '经度';
        COMMENT ON COLUMN houses.view_count IS '浏览数';
        COMMENT ON COLUMN houses.fav_count IS '收藏数';
        COMMENT ON COLUMN houses.contact_count IS '联系数';
        COMMENT ON COLUMN houses.share_count IS '分享数';
        COMMENT ON COLUMN houses.viewing_count IS '看房预约数';
        COMMENT ON COLUMN houses.last_viewing_at IS '最近看房时间';
        COMMENT ON COLUMN houses.content_hash IS '图文指纹（MD5/SHA256）';
        COMMENT ON COLUMN houses.risk_score IS '风险评分 0-100，<30 限制发布';
        COMMENT ON COLUMN houses.same_house_id IS '同房源识别 ID';
        COMMENT ON COLUMN houses.video_url IS '视频 URL';
        COMMENT ON COLUMN houses.video_cover IS '视频封面';
        COMMENT ON COLUMN houses.vr_url IS 'VR 看房 URL（720°全景）';
        COMMENT ON COLUMN houses.panorama_url IS '360° 全景图 URL';
        COMMENT ON COLUMN houses.facilities IS '配套设施 ID 数组 JSON';
        COMMENT ON COLUMN houses.tags IS '标签数组（最多5个）';
        COMMENT ON COLUMN houses.nearby_pois IS '附近 POI JSON（地铁/学校/医院/商超等）';
        COMMENT ON COLUMN houses.featured IS '精选推荐';
        COMMENT ON COLUMN houses.picked IS '运营甄选';
        COMMENT ON COLUMN houses.verified IS '官方验真';
        COMMENT ON COLUMN houses.promotion_level IS '推广等级 0-10';
        COMMENT ON COLUMN houses.traffic_weight IS '流量权重 0.00-9.99';
        COMMENT ON COLUMN houses.real_house_verified IS '真房源认证';
        COMMENT ON COLUMN houses.real_house_verified_at IS '真房源认证时间';
    END IF;
END $$;

-- ============================================================
-- 第二部分：19 张子表
-- ============================================================

-- ------------------------------------------------------------
-- 1. house_communities 小区信息表
--    对标贝壳/链家：小区主页/位置/建筑年代/物业费/均价
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS house_communities (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    name VARCHAR(128) NOT NULL,
    alias VARCHAR(128) NOT NULL DEFAULT '',
    city VARCHAR(64) NOT NULL DEFAULT '',
    district VARCHAR(64) NOT NULL DEFAULT '',
    business_district VARCHAR(128) NOT NULL DEFAULT '',
    address VARCHAR(500) NOT NULL DEFAULT '',
    latitude DECIMAL(10,7) NOT NULL DEFAULT 0,
    longitude DECIMAL(10,7) NOT NULL DEFAULT 0,
    building_count INT NOT NULL DEFAULT 0,
    house_count INT NOT NULL DEFAULT 0,
    building_year INT NOT NULL DEFAULT 0,
    building_type VARCHAR(32) NOT NULL DEFAULT '',
    developer VARCHAR(128) NOT NULL DEFAULT '',
    property_company VARCHAR(128) NOT NULL DEFAULT '',
    property_fee DECIMAL(8,2) NOT NULL DEFAULT 0,
    parking_ratio VARCHAR(32) NOT NULL DEFAULT '',
    greening_rate DECIMAL(5,2) NOT NULL DEFAULT 0,
    plot_ratio DECIMAL(5,2) NOT NULL DEFAULT 0,
    avg_sale_price DECIMAL(10,2) NOT NULL DEFAULT 0,
    avg_rent_price DECIMAL(10,2) NOT NULL DEFAULT 0,
    description TEXT NOT NULL DEFAULT '',
    cover_image VARCHAR(255) NOT NULL DEFAULT '',
    images JSONB,
    nearby_pois JSONB,
    status INT NOT NULL DEFAULT 1,
    follower_count INT NOT NULL DEFAULT 0,
    on_sale_count INT NOT NULL DEFAULT 0,
    on_rent_count INT NOT NULL DEFAULT 0,
    CONSTRAINT uniq_house_communities_name_city UNIQUE (name, city)
);
CREATE INDEX IF NOT EXISTS idx_house_communities_region_id ON house_communities(region_id);
CREATE INDEX IF NOT EXISTS idx_house_communities_name ON house_communities(name);
CREATE INDEX IF NOT EXISTS idx_house_communities_city ON house_communities(city);
CREATE INDEX IF NOT EXISTS idx_house_communities_district ON house_communities(district);
CREATE INDEX IF NOT EXISTS idx_house_communities_business_district ON house_communities(business_district);
CREATE INDEX IF NOT EXISTS idx_house_communities_status ON house_communities(status);
CREATE INDEX IF NOT EXISTS idx_house_communities_deleted_at ON house_communities(deleted_at);
COMMENT ON TABLE house_communities IS '小区信息表（小区主页/位置/建筑年代/物业费/均价/开发商）';

-- ------------------------------------------------------------
-- 2. house_agents 经纪人表
--    对标贝壳/链家：姓名/手机/门店/评分/成交量/认证
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS house_agents (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(50) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    avatar VARCHAR(255) NOT NULL DEFAULT '',
    gender VARCHAR(16) NOT NULL DEFAULT 'unlimited',
    store_id BIGINT NOT NULL DEFAULT 0,
    store_name VARCHAR(128) NOT NULL DEFAULT '',
    company VARCHAR(128) NOT NULL DEFAULT '',
    title VARCHAR(64) NOT NULL DEFAULT '',
    level INT NOT NULL DEFAULT 0,
    license_no VARCHAR(64) NOT NULL DEFAULT '',
    license_image VARCHAR(255) NOT NULL DEFAULT '',
    id_card_front VARCHAR(255) NOT NULL DEFAULT '',
    id_card_back VARCHAR(255) NOT NULL DEFAULT '',
    business_card VARCHAR(255) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    good_at JSONB,
    service_area JSONB,
    rating DECIMAL(3,2) NOT NULL DEFAULT 5.00,
    rating_count INT NOT NULL DEFAULT 0,
    listing_count INT NOT NULL DEFAULT 0,
    deal_count INT NOT NULL DEFAULT 0,
    total_amount DECIMAL(14,2) NOT NULL DEFAULT 0,
    response_time INT NOT NULL DEFAULT 0,
    response_rate DECIMAL(5,2) NOT NULL DEFAULT 0,
    online_status INT NOT NULL DEFAULT 0,
    last_active_at TIMESTAMPTZ,
    verified_at TIMESTAMPTZ,
    approved_at TIMESTAMPTZ,
    rejected_reason VARCHAR(500) NOT NULL DEFAULT '',
    status INT NOT NULL DEFAULT 0,
    follower_count INT NOT NULL DEFAULT 0,
    tags JSONB,
    CONSTRAINT uniq_house_agents_phone UNIQUE (phone)
);
CREATE INDEX IF NOT EXISTS idx_house_agents_region_id ON house_agents(region_id);
CREATE INDEX IF NOT EXISTS idx_house_agents_user_id ON house_agents(user_id);
CREATE INDEX IF NOT EXISTS idx_house_agents_name ON house_agents(name);
CREATE INDEX IF NOT EXISTS idx_house_agents_store_id ON house_agents(store_id);
CREATE INDEX IF NOT EXISTS idx_house_agents_company ON house_agents(company);
CREATE INDEX IF NOT EXISTS idx_house_agents_level ON house_agents(level);
CREATE INDEX IF NOT EXISTS idx_house_agents_license_no ON house_agents(license_no);
CREATE INDEX IF NOT EXISTS idx_house_agents_rating ON house_agents(rating);
CREATE INDEX IF NOT EXISTS idx_house_agents_status ON house_agents(status);
CREATE INDEX IF NOT EXISTS idx_house_agents_verified_at ON house_agents(verified_at);
CREATE INDEX IF NOT EXISTS idx_house_agents_approved_at ON house_agents(approved_at);
CREATE INDEX IF NOT EXISTS idx_house_agents_deleted_at ON house_agents(deleted_at);
COMMENT ON TABLE house_agents IS '经纪人表（姓名/手机/门店/评分/成交量/认证状态）';

-- ------------------------------------------------------------
-- 3. house_listings 房源发布表（与 houses 主表 1:1 冗余发布信息）
--    对标贝壳/链家：发布人/发布类型/发布时间/有效期/发布状态
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS house_listings (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    listing_no VARCHAR(64) NOT NULL,
    house_id BIGINT NOT NULL,
    community_id BIGINT NOT NULL DEFAULT 0,
    agent_id BIGINT NOT NULL DEFAULT 0,
    publisher_id BIGINT NOT NULL,
    publisher_name VARCHAR(50) NOT NULL DEFAULT '',
    publisher_phone VARCHAR(20) NOT NULL DEFAULT '',
    publisher_avatar VARCHAR(255) NOT NULL DEFAULT '',
    publisher_type VARCHAR(16) NOT NULL DEFAULT 'personal',
    listing_type VARCHAR(16) NOT NULL DEFAULT 'rent',
    title VARCHAR(200) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    price DECIMAL(14,2) NOT NULL DEFAULT 0,
    price_unit VARCHAR(16) NOT NULL DEFAULT 'month',
    decoration VARCHAR(16) NOT NULL DEFAULT 'rough',
    orientation VARCHAR(32) NOT NULL DEFAULT '',
    layout VARCHAR(32) NOT NULL DEFAULT '',
    building_area DECIMAL(10,2) NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 0,
    audit_status INT NOT NULL DEFAULT 0,
    audit_reason VARCHAR(500) NOT NULL DEFAULT '',
    published_at TIMESTAMPTZ,
    expired_at TIMESTAMPTZ,
    refreshed_at TIMESTAMPTZ,
    offline_at TIMESTAMPTZ,
    refresh_count INT NOT NULL DEFAULT 0,
    view_count INT NOT NULL DEFAULT 0,
    fav_count INT NOT NULL DEFAULT 0,
    contact_count INT NOT NULL DEFAULT 0,
    CONSTRAINT uniq_house_listings_no UNIQUE (listing_no)
);
CREATE INDEX IF NOT EXISTS idx_house_listings_region_id ON house_listings(region_id);
CREATE INDEX IF NOT EXISTS idx_house_listings_house_id ON house_listings(house_id);
CREATE INDEX IF NOT EXISTS idx_house_listings_community_id ON house_listings(community_id);
CREATE INDEX IF NOT EXISTS idx_house_listings_agent_id ON house_listings(agent_id);
CREATE INDEX IF NOT EXISTS idx_house_listings_publisher ON house_listings(publisher_id, status);
CREATE INDEX IF NOT EXISTS idx_house_listings_listing_type ON house_listings(listing_type);
CREATE INDEX IF NOT EXISTS idx_house_listings_status ON house_listings(status);
CREATE INDEX IF NOT EXISTS idx_house_listings_audit_status ON house_listings(audit_status);
CREATE INDEX IF NOT EXISTS idx_house_listings_published_at ON house_listings(published_at);
CREATE INDEX IF NOT EXISTS idx_house_listings_expired_at ON house_listings(expired_at);
CREATE INDEX IF NOT EXISTS idx_house_listings_refreshed_at ON house_listings(refreshed_at);
CREATE INDEX IF NOT EXISTS idx_house_listings_deleted_at ON house_listings(deleted_at);
COMMENT ON TABLE house_listings IS '房源发布表（发布人/发布类型/标题/价格/状态机/有效期/刷新）';

-- ------------------------------------------------------------
-- 4. house_contracts 合同电子化表
--    对标贝壳/链家：租约/买卖合同/电子签/三方协议
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS house_contracts (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    contract_no VARCHAR(64) NOT NULL,
    contract_type VARCHAR(32) NOT NULL DEFAULT 'rent',
    house_id BIGINT NOT NULL,
    listing_id BIGINT NOT NULL DEFAULT 0,
    community_id BIGINT NOT NULL DEFAULT 0,
    party_a_id BIGINT NOT NULL,
    party_a_name VARCHAR(50) NOT NULL DEFAULT '',
    party_a_phone VARCHAR(20) NOT NULL DEFAULT '',
    party_a_id_card VARCHAR(32) NOT NULL DEFAULT '',
    party_b_id BIGINT NOT NULL,
    party_b_name VARCHAR(50) NOT NULL DEFAULT '',
    party_b_phone VARCHAR(20) NOT NULL DEFAULT '',
    party_b_id_card VARCHAR(32) NOT NULL DEFAULT '',
    agent_id BIGINT NOT NULL DEFAULT 0,
    agent_name VARCHAR(50) NOT NULL DEFAULT '',
    title VARCHAR(200) NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    attachments JSONB,
    amount DECIMAL(14,2) NOT NULL DEFAULT 0,
    deposit DECIMAL(14,2) NOT NULL DEFAULT 0,
    commission DECIMAL(14,2) NOT NULL DEFAULT 0,
    commission_payer VARCHAR(16) NOT NULL DEFAULT 'both',
    start_date DATE,
    end_date DATE,
    payment_method VARCHAR(32) NOT NULL DEFAULT 'monthly',
    sign_method VARCHAR(32) NOT NULL DEFAULT 'online',
    party_a_signed_at TIMESTAMPTZ,
    party_b_signed_at TIMESTAMPTZ,
    agent_signed_at TIMESTAMPTZ,
    status INT NOT NULL DEFAULT 0,
    effective_at TIMESTAMPTZ,
    terminated_at TIMESTAMPTZ,
    terminated_reason VARCHAR(500) NOT NULL DEFAULT '',
    archived_at TIMESTAMPTZ,
    CONSTRAINT uniq_house_contracts_no UNIQUE (contract_no)
);
CREATE INDEX IF NOT EXISTS idx_house_contracts_region_id ON house_contracts(region_id);
CREATE INDEX IF NOT EXISTS idx_house_contracts_house_id ON house_contracts(house_id);
CREATE INDEX IF NOT EXISTS idx_house_contracts_listing_id ON house_contracts(listing_id);
CREATE INDEX IF NOT EXISTS idx_house_contracts_community_id ON house_contracts(community_id);
CREATE INDEX IF NOT EXISTS idx_house_contracts_party_a ON house_contracts(party_a_id, status);
CREATE INDEX IF NOT EXISTS idx_house_contracts_party_b ON house_contracts(party_b_id, status);
CREATE INDEX IF NOT EXISTS idx_house_contracts_agent_id ON house_contracts(agent_id);
CREATE INDEX IF NOT EXISTS idx_house_contracts_contract_type ON house_contracts(contract_type);
CREATE INDEX IF NOT EXISTS idx_house_contracts_status ON house_contracts(status);
CREATE INDEX IF NOT EXISTS idx_house_contracts_start_date ON house_contracts(start_date);
CREATE INDEX IF NOT EXISTS idx_house_contracts_end_date ON house_contracts(end_date);
CREATE INDEX IF NOT EXISTS idx_house_contracts_effective_at ON house_contracts(effective_at);
CREATE INDEX IF NOT EXISTS idx_house_contracts_terminated_at ON house_contracts(terminated_at);
CREATE INDEX IF NOT EXISTS idx_house_contracts_deleted_at ON house_contracts(deleted_at);
COMMENT ON TABLE house_contracts IS '合同电子化表（租约/买卖合同/电子签/三方协议/状态机）';

-- ------------------------------------------------------------
-- 5. house_viewings 看房预约表
--    对标贝壳/链家：预约时间/经纪人/用户/看房结果
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS house_viewings (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    viewing_no VARCHAR(64) NOT NULL,
    house_id BIGINT NOT NULL,
    listing_id BIGINT NOT NULL DEFAULT 0,
    community_id BIGINT NOT NULL DEFAULT 0,
    user_id BIGINT NOT NULL,
    user_name VARCHAR(50) NOT NULL DEFAULT '',
    user_phone VARCHAR(20) NOT NULL DEFAULT '',
    user_avatar VARCHAR(255) NOT NULL DEFAULT '',
    agent_id BIGINT NOT NULL DEFAULT 0,
    agent_name VARCHAR(50) NOT NULL DEFAULT '',
    agent_phone VARCHAR(20) NOT NULL DEFAULT '',
    scheduled_at TIMESTAMPTZ,
    duration_minutes INT NOT NULL DEFAULT 30,
    viewing_type VARCHAR(32) NOT NULL DEFAULT 'offline',
    online_url VARCHAR(500) NOT NULL DEFAULT '',
    online_password VARCHAR(64) NOT NULL DEFAULT '',
    meet_location VARCHAR(255) NOT NULL DEFAULT '',
    remark VARCHAR(500) NOT NULL DEFAULT '',
    status INT NOT NULL DEFAULT 0,
    result VARCHAR(32) NOT NULL DEFAULT 'pending',
    feedback TEXT NOT NULL DEFAULT '',
    rating INT NOT NULL DEFAULT 0,
    attended_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    canceled_at TIMESTAMPTZ,
    canceled_reason VARCHAR(500) NOT NULL DEFAULT '',
    canceled_by BIGINT NOT NULL DEFAULT 0,
    reminder_sent BOOLEAN NOT NULL DEFAULT FALSE,
    reminder_sent_at TIMESTAMPTZ,
    CONSTRAINT uniq_house_viewings_no UNIQUE (viewing_no)
);
CREATE INDEX IF NOT EXISTS idx_house_viewings_region_id ON house_viewings(region_id);
CREATE INDEX IF NOT EXISTS idx_house_viewings_house_id ON house_viewings(house_id);
CREATE INDEX IF NOT EXISTS idx_house_viewings_listing_id ON house_viewings(listing_id);
CREATE INDEX IF NOT EXISTS idx_house_viewings_user ON house_viewings(user_id, status);
CREATE INDEX IF NOT EXISTS idx_house_viewings_agent ON house_viewings(agent_id, status);
CREATE INDEX IF NOT EXISTS idx_house_viewings_status ON house_viewings(status);
CREATE INDEX IF NOT EXISTS idx_house_viewings_scheduled_at ON house_viewings(scheduled_at);
CREATE INDEX IF NOT EXISTS idx_house_viewings_viewing_type ON house_viewings(viewing_type);
CREATE INDEX IF NOT EXISTS idx_house_viewings_result ON house_viewings(result);
CREATE INDEX IF NOT EXISTS idx_house_viewings_attended_at ON house_viewings(attended_at);
CREATE INDEX IF NOT EXISTS idx_house_viewings_completed_at ON house_viewings(completed_at);
CREATE INDEX IF NOT EXISTS idx_house_viewings_canceled_at ON house_viewings(canceled_at);
CREATE INDEX IF NOT EXISTS idx_house_viewings_deleted_at ON house_viewings(deleted_at);
COMMENT ON TABLE house_viewings IS '看房预约表（预约时间/经纪人/用户/在线线下/结果/评分）';

-- ------------------------------------------------------------
-- 6. house_facilities 配套设施表
--    对标贝壳：家具/家电/独立卫浴/阳台/车位等
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS house_facilities (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    name VARCHAR(64) NOT NULL,
    code VARCHAR(64) NOT NULL,
    category VARCHAR(32) NOT NULL DEFAULT 'indoor',
    icon VARCHAR(64) NOT NULL DEFAULT '',
    color VARCHAR(16) NOT NULL DEFAULT '#409EFF',
    background VARCHAR(32) NOT NULL DEFAULT '',
    description VARCHAR(500) NOT NULL DEFAULT '',
    status INT NOT NULL DEFAULT 1,
    sort INT NOT NULL DEFAULT 0,
    use_count INT NOT NULL DEFAULT 0,
    is_hot BOOLEAN NOT NULL DEFAULT FALSE,
    creator_id BIGINT NOT NULL DEFAULT 0,
    CONSTRAINT uniq_house_facilities_code UNIQUE (code)
);
CREATE INDEX IF NOT EXISTS idx_house_facilities_category ON house_facilities(category);
CREATE INDEX IF NOT EXISTS idx_house_facilities_status ON house_facilities(status);
CREATE INDEX IF NOT EXISTS idx_house_facilities_is_hot ON house_facilities(is_hot) WHERE is_hot = TRUE;
CREATE INDEX IF NOT EXISTS idx_house_facilities_sort ON house_facilities(sort);
CREATE INDEX IF NOT EXISTS idx_house_facilities_creator_id ON house_facilities(creator_id);
CREATE INDEX IF NOT EXISTS idx_house_facilities_deleted_at ON house_facilities(deleted_at);
COMMENT ON TABLE house_facilities IS '配套设施表（家具/家电/独立卫浴/阳台/车位/暖气等）';

-- ------------------------------------------------------------
-- 7. house_images 房源图片表
--    对标贝壳：户型图/实景图/客厅/卧室/厨房/卫生间/小区环境
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS house_images (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    house_id BIGINT NOT NULL,
    listing_id BIGINT NOT NULL DEFAULT 0,
    url VARCHAR(500) NOT NULL,
    thumbnail VARCHAR(500) NOT NULL DEFAULT '',
    image_type VARCHAR(32) NOT NULL DEFAULT 'real',
    title VARCHAR(128) NOT NULL DEFAULT '',
    description VARCHAR(500) NOT NULL DEFAULT '',
    width INT NOT NULL DEFAULT 0,
    height INT NOT NULL DEFAULT 0,
    size BIGINT NOT NULL DEFAULT 0,
    sort INT NOT NULL DEFAULT 0,
    is_cover BOOLEAN NOT NULL DEFAULT FALSE,
    status INT NOT NULL DEFAULT 1,
    uploader_id BIGINT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_house_images_house_id ON house_images(house_id);
CREATE INDEX IF NOT EXISTS idx_house_images_listing_id ON house_images(listing_id);
CREATE INDEX IF NOT EXISTS idx_house_images_image_type ON house_images(image_type);
CREATE INDEX IF NOT EXISTS idx_house_images_is_cover ON house_images(is_cover) WHERE is_cover = TRUE;
CREATE INDEX IF NOT EXISTS idx_house_images_sort ON house_images(sort);
CREATE INDEX IF NOT EXISTS idx_house_images_status ON house_images(status);
CREATE INDEX IF NOT EXISTS idx_house_images_deleted_at ON house_images(deleted_at);
COMMENT ON TABLE house_images IS '房源图片表（户型图/实景图/客厅/卧室/厨房/卫生间/小区环境）';

-- ------------------------------------------------------------
-- 8. house_vr_tours VR 看房记录表
--    对标贝壳：720°全景/虚拟看房/VR 设备/录制信息
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS house_vr_tours (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    house_id BIGINT NOT NULL,
    listing_id BIGINT NOT NULL DEFAULT 0,
    community_id BIGINT NOT NULL DEFAULT 0,
    vr_no VARCHAR(64) NOT NULL,
    title VARCHAR(200) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    vr_type VARCHAR(32) NOT NULL DEFAULT 'panorama',
    vr_url VARCHAR(500) NOT NULL,
    cover_image VARCHAR(500) NOT NULL DEFAULT '',
    scenes JSONB,
    duration_seconds INT NOT NULL DEFAULT 0,
    view_count INT NOT NULL DEFAULT 0,
    share_count INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 0,
    recorder_id BIGINT NOT NULL DEFAULT 0,
    recorder_name VARCHAR(50) NOT NULL DEFAULT '',
    recorded_at TIMESTAMPTZ,
    published_at TIMESTAMPTZ,
    offline_at TIMESTAMPTZ,
    equipment VARCHAR(64) NOT NULL DEFAULT '',
    resolution VARCHAR(32) NOT NULL DEFAULT '',
    file_size BIGINT NOT NULL DEFAULT 0,
    CONSTRAINT uniq_house_vr_tours_no UNIQUE (vr_no)
);
CREATE INDEX IF NOT EXISTS idx_house_vr_tours_house_id ON house_vr_tours(house_id);
CREATE INDEX IF NOT EXISTS idx_house_vr_tours_listing_id ON house_vr_tours(listing_id);
CREATE INDEX IF NOT EXISTS idx_house_vr_tours_community_id ON house_vr_tours(community_id);
CREATE INDEX IF NOT EXISTS idx_house_vr_tours_vr_type ON house_vr_tours(vr_type);
CREATE INDEX IF NOT EXISTS idx_house_vr_tours_status ON house_vr_tours(status);
CREATE INDEX IF NOT EXISTS idx_house_vr_tours_recorder_id ON house_vr_tours(recorder_id);
CREATE INDEX IF NOT EXISTS idx_house_vr_tours_published_at ON house_vr_tours(published_at);
CREATE INDEX IF NOT EXISTS idx_house_vr_tours_deleted_at ON house_vr_tours(deleted_at);
COMMENT ON TABLE house_vr_tours IS 'VR 看房记录表（720°全景/虚拟看房/录制信息/场景 JSON）';

-- ------------------------------------------------------------
-- 9. house_categories 房源分类表
--    对标贝壳：整租/合租/独栋/公寓/别墅/写字楼/商铺
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS house_categories (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    name VARCHAR(64) NOT NULL,
    code VARCHAR(64) NOT NULL,
    parent_id BIGINT NOT NULL DEFAULT 0,
    level INT NOT NULL DEFAULT 1,
    listing_type VARCHAR(16) NOT NULL DEFAULT 'rent',
    property_type VARCHAR(32) NOT NULL DEFAULT 'residential',
    icon VARCHAR(64) NOT NULL DEFAULT '',
    color VARCHAR(16) NOT NULL DEFAULT '#409EFF',
    description VARCHAR(500) NOT NULL DEFAULT '',
    sort INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 1,
    house_count INT NOT NULL DEFAULT 0,
    CONSTRAINT uniq_house_categories_code UNIQUE (code)
);
CREATE INDEX IF NOT EXISTS idx_house_categories_parent_id ON house_categories(parent_id);
CREATE INDEX IF NOT EXISTS idx_house_categories_level ON house_categories(level);
CREATE INDEX IF NOT EXISTS idx_house_categories_listing_type ON house_categories(listing_type);
CREATE INDEX IF NOT EXISTS idx_house_categories_property_type ON house_categories(property_type);
CREATE INDEX IF NOT EXISTS idx_house_categories_status ON house_categories(status);
CREATE INDEX IF NOT EXISTS idx_house_categories_sort ON house_categories(sort);
CREATE INDEX IF NOT EXISTS idx_house_categories_deleted_at ON house_categories(deleted_at);
COMMENT ON TABLE house_categories IS '房源分类表（整租/合租/独栋/公寓/别墅/写字楼/商铺）';

-- ------------------------------------------------------------
-- 10. house_favorites 房源收藏表
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS house_favorites (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id BIGINT NOT NULL,
    house_id BIGINT NOT NULL DEFAULT 0,
    listing_id BIGINT NOT NULL DEFAULT 0,
    community_id BIGINT NOT NULL DEFAULT 0,
    agent_id BIGINT NOT NULL DEFAULT 0,
    favorite_type VARCHAR(16) NOT NULL DEFAULT 'house',
    notify BOOLEAN NOT NULL DEFAULT TRUE,
    remark VARCHAR(500) NOT NULL DEFAULT '',
    CONSTRAINT uniq_house_fav_user_type_target UNIQUE (user_id, favorite_type, house_id, listing_id, community_id, agent_id)
);
CREATE INDEX IF NOT EXISTS idx_house_favorites_user_id ON house_favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_house_favorites_house_id ON house_favorites(house_id);
CREATE INDEX IF NOT EXISTS idx_house_favorites_listing_id ON house_favorites(listing_id);
CREATE INDEX IF NOT EXISTS idx_house_favorites_community_id ON house_favorites(community_id);
CREATE INDEX IF NOT EXISTS idx_house_favorites_agent_id ON house_favorites(agent_id);
CREATE INDEX IF NOT EXISTS idx_house_favorites_favorite_type ON house_favorites(favorite_type);
CREATE INDEX IF NOT EXISTS idx_house_favorites_deleted_at ON house_favorites(deleted_at);
COMMENT ON TABLE house_favorites IS '房源/小区/经纪人收藏表（用户/类型/目标 ID）';

-- ------------------------------------------------------------
-- 11. house_views 浏览记录表
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS house_views (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_id BIGINT NOT NULL,
    house_id BIGINT NOT NULL DEFAULT 0,
    listing_id BIGINT NOT NULL DEFAULT 0,
    community_id BIGINT NOT NULL DEFAULT 0,
    agent_id BIGINT NOT NULL DEFAULT 0,
    view_type VARCHAR(16) NOT NULL DEFAULT 'house',
    source VARCHAR(32) NOT NULL DEFAULT 'list',
    ip VARCHAR(64) NOT NULL DEFAULT '',
    user_agent VARCHAR(255) NOT NULL DEFAULT '',
    duration_seconds INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_house_views_region_id ON house_views(region_id);
CREATE INDEX IF NOT EXISTS idx_house_views_user_id ON house_views(user_id);
CREATE INDEX IF NOT EXISTS idx_house_views_house_id ON house_views(house_id);
CREATE INDEX IF NOT EXISTS idx_house_views_listing_id ON house_views(listing_id);
CREATE INDEX IF NOT EXISTS idx_house_views_community_id ON house_views(community_id);
CREATE INDEX IF NOT EXISTS idx_house_views_agent_id ON house_views(agent_id);
CREATE INDEX IF NOT EXISTS idx_house_views_view_type ON house_views(view_type);
CREATE INDEX IF NOT EXISTS idx_house_views_source ON house_views(source);
CREATE INDEX IF NOT EXISTS idx_house_views_created_at ON house_views(created_at);
COMMENT ON TABLE house_views IS '浏览记录表（用户/房源/小区/经纪人/来源）';

-- ------------------------------------------------------------
-- 12. house_reports 举报工单表
--    对标贝壳/58：虚假房源/诈骗/色情/侵权 + SLA 24h/72h
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS house_reports (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    report_no VARCHAR(64) NOT NULL,
    target_type VARCHAR(32) NOT NULL,
    target_id BIGINT NOT NULL,
    target_user_id BIGINT NOT NULL,
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
    CONSTRAINT uniq_house_reports_no UNIQUE (report_no)
);
CREATE INDEX IF NOT EXISTS idx_house_reports_target_type_target ON house_reports(target_type, target_id);
CREATE INDEX IF NOT EXISTS idx_house_reports_target_user_id ON house_reports(target_user_id);
CREATE INDEX IF NOT EXISTS idx_house_reports_reporter_id ON house_reports(reporter_id);
CREATE INDEX IF NOT EXISTS idx_house_reports_reported_user_id ON house_reports(reported_user_id);
CREATE INDEX IF NOT EXISTS idx_house_reports_report_type ON house_reports(report_type);
CREATE INDEX IF NOT EXISTS idx_house_reports_status ON house_reports(status);
CREATE INDEX IF NOT EXISTS idx_house_reports_handler_id ON house_reports(handler_id);
CREATE INDEX IF NOT EXISTS idx_house_reports_penalty_user_id ON house_reports(penalty_user_id);
CREATE INDEX IF NOT EXISTS idx_house_reports_sla_deadline ON house_reports(sla_deadline);
CREATE INDEX IF NOT EXISTS idx_house_reports_handled_at ON house_reports(handled_at);
CREATE INDEX IF NOT EXISTS idx_house_reports_appealed_at ON house_reports(appealed_at);
CREATE INDEX IF NOT EXISTS idx_house_reports_deleted_at ON house_reports(deleted_at);
COMMENT ON TABLE house_reports IS '举报工单表（目标/类型/原因/证据/状态/SLA/申诉）';

-- ------------------------------------------------------------
-- 13. house_reviews 评价表
--    对标贝壳/链家：经纪人/小区 5 星+文字+图片+追评+回复
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS house_reviews (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    target_type VARCHAR(32) NOT NULL DEFAULT 'agent',
    target_id BIGINT NOT NULL,
    reviewer_id BIGINT NOT NULL,
    reviewer_name VARCHAR(50) NOT NULL DEFAULT '',
    reviewer_avatar VARCHAR(255) NOT NULL DEFAULT '',
    review_type VARCHAR(32) NOT NULL DEFAULT 'tenant',
    rating INT NOT NULL DEFAULT 5,
    content TEXT NOT NULL DEFAULT '',
    images JSONB,
    video_url VARCHAR(255) NOT NULL DEFAULT '',
    is_anonymous BOOLEAN NOT NULL DEFAULT FALSE,
    is_recommended BOOLEAN NOT NULL DEFAULT TRUE,
    tags JSONB,
    deal_amount DECIMAL(14,2) NOT NULL DEFAULT 0,
    service_attitude INT NOT NULL DEFAULT 5,
    professional_skill INT NOT NULL DEFAULT 5,
    reply TEXT NOT NULL DEFAULT '',
    reply_at TIMESTAMPTZ,
    append_content TEXT NOT NULL DEFAULT '',
    append_images JSONB,
    append_at TIMESTAMPTZ,
    like_count INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 1,
    CONSTRAINT uniq_house_reviews_target_reviewer UNIQUE (target_type, target_id, reviewer_id)
);
CREATE INDEX IF NOT EXISTS idx_house_reviews_region_id ON house_reviews(region_id);
CREATE INDEX IF NOT EXISTS idx_house_reviews_target ON house_reviews(target_type, target_id);
CREATE INDEX IF NOT EXISTS idx_house_reviews_reviewer_id ON house_reviews(reviewer_id);
CREATE INDEX IF NOT EXISTS idx_house_reviews_review_type ON house_reviews(review_type);
CREATE INDEX IF NOT EXISTS idx_house_reviews_rating ON house_reviews(rating);
CREATE INDEX IF NOT EXISTS idx_house_reviews_is_recommended ON house_reviews(is_recommended) WHERE is_recommended = FALSE;
CREATE INDEX IF NOT EXISTS idx_house_reviews_status ON house_reviews(status);
CREATE INDEX IF NOT EXISTS idx_house_reviews_reply_at ON house_reviews(reply_at);
CREATE INDEX IF NOT EXISTS idx_house_reviews_append_at ON house_reviews(append_at);
CREATE INDEX IF NOT EXISTS idx_house_reviews_deleted_at ON house_reviews(deleted_at);
COMMENT ON TABLE house_reviews IS '评价表（经纪人/小区 5星+文字+图片+服务态度+专业能力+追评+回复）';

-- ------------------------------------------------------------
-- 14. house_mortgages 房贷计算配置表
--    对标贝壳/安居客：首付比例/利率/期数/月供
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS house_mortgages (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    name VARCHAR(64) NOT NULL,
    code VARCHAR(64) NOT NULL,
    loan_type VARCHAR(32) NOT NULL DEFAULT 'commercial',
    min_down_payment DECIMAL(5,2) NOT NULL DEFAULT 30.00,
    max_down_payment DECIMAL(5,2) NOT NULL DEFAULT 100.00,
    interest_rate DECIMAL(6,4) NOT NULL DEFAULT 0,
    max_periods INT NOT NULL DEFAULT 360,
    max_amount DECIMAL(14,2) NOT NULL DEFAULT 0,
    description VARCHAR(500) NOT NULL DEFAULT '',
    sort INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 1,
    is_hot BOOLEAN NOT NULL DEFAULT FALSE,
    use_count INT NOT NULL DEFAULT 0,
    CONSTRAINT uniq_house_mortgages_code UNIQUE (code)
);
CREATE INDEX IF NOT EXISTS idx_house_mortgages_loan_type ON house_mortgages(loan_type);
CREATE INDEX IF NOT EXISTS idx_house_mortgages_status ON house_mortgages(status);
CREATE INDEX IF NOT EXISTS idx_house_mortgages_is_hot ON house_mortgages(is_hot) WHERE is_hot = TRUE;
CREATE INDEX IF NOT EXISTS idx_house_mortgages_sort ON house_mortgages(sort);
CREATE INDEX IF NOT EXISTS idx_house_mortgages_deleted_at ON house_mortgages(deleted_at);
COMMENT ON TABLE house_mortgages IS '房贷计算配置表（首付比例/利率/期数/月供计算）';

-- ------------------------------------------------------------
-- 15. house_audit_rules 审核规则表
--    对标贝壳：真房源/价格异常/频率限制/敏感词
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS house_audit_rules (
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
CREATE INDEX IF NOT EXISTS idx_house_audit_rules_rule_type ON house_audit_rules(rule_type);
CREATE INDEX IF NOT EXISTS idx_house_audit_rules_rule_key ON house_audit_rules(rule_key);
CREATE INDEX IF NOT EXISTS idx_house_audit_rules_severity ON house_audit_rules(severity);
CREATE INDEX IF NOT EXISTS idx_house_audit_rules_status ON house_audit_rules(status);
CREATE INDEX IF NOT EXISTS idx_house_audit_rules_deleted_at ON house_audit_rules(deleted_at);
COMMENT ON TABLE house_audit_rules IS '审核规则表（真房源/价格异常/频率限制/敏感词）';

-- ------------------------------------------------------------
-- 16. house_statistics 数据统计表
--    对标贝壳：曝光/点击/收藏/看房/成交转化
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS house_statistics (
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
    viewing_count INT NOT NULL DEFAULT 0,
    deal_count INT NOT NULL DEFAULT 0,
    conversion_rate DECIMAL(5,2) NOT NULL DEFAULT 0,
    avg_sale_price DECIMAL(10,2) NOT NULL DEFAULT 0,
    avg_rent_price DECIMAL(10,2) NOT NULL DEFAULT 0,
    avg_deal_days INT NOT NULL DEFAULT 0,
    CONSTRAINT uniq_house_stats_date_type_target UNIQUE (stat_date, stat_type, target_id)
);
CREATE INDEX IF NOT EXISTS idx_house_statistics_region_id ON house_statistics(region_id);
CREATE INDEX IF NOT EXISTS idx_house_statistics_stat_date ON house_statistics(stat_date);
CREATE INDEX IF NOT EXISTS idx_house_statistics_stat_type ON house_statistics(stat_type);
CREATE INDEX IF NOT EXISTS idx_house_statistics_target_id ON house_statistics(target_id);
CREATE INDEX IF NOT EXISTS idx_house_statistics_deleted_at ON house_statistics(deleted_at);
COMMENT ON TABLE house_statistics IS '数据统计表（曝光/点击/收藏/看房/成交/转化率/均价/平均成交周期）';

-- ------------------------------------------------------------
-- 17. house_escrows 担保交易表
--    对标贝壳/链家：定金/中介费/资金托管/解冻/放款
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS house_escrows (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    escrow_no VARCHAR(64) NOT NULL,
    escrow_type VARCHAR(32) NOT NULL DEFAULT 'deposit',
    house_id BIGINT NOT NULL DEFAULT 0,
    listing_id BIGINT NOT NULL DEFAULT 0,
    contract_id BIGINT NOT NULL DEFAULT 0,
    community_id BIGINT NOT NULL DEFAULT 0,
    payer_id BIGINT NOT NULL,
    payee_id BIGINT NOT NULL,
    agent_id BIGINT NOT NULL DEFAULT 0,
    amount DECIMAL(14,2) NOT NULL,
    platform_fee DECIMAL(14,2) NOT NULL DEFAULT 0,
    agent_fee DECIMAL(14,2) NOT NULL DEFAULT 0,
    payee_amount DECIMAL(14,2) NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 1,
    pay_method VARCHAR(32) NOT NULL DEFAULT 'wechat',
    pay_trade_no VARCHAR(128) NOT NULL DEFAULT '',
    paid_at TIMESTAMPTZ,
    frozen_at TIMESTAMPTZ,
    release_at TIMESTAMPTZ,
    refunded_at TIMESTAMPTZ,
    auto_release_at TIMESTAMPTZ,
    dispute_reason VARCHAR(500) NOT NULL DEFAULT '',
    arbitration_result TEXT NOT NULL DEFAULT '',
    completed_at TIMESTAMPTZ,
    CONSTRAINT uniq_house_escrows_no UNIQUE (escrow_no)
);
CREATE INDEX IF NOT EXISTS idx_house_escrows_region_id ON house_escrows(region_id);
CREATE INDEX IF NOT EXISTS idx_house_escrows_escrow_type ON house_escrows(escrow_type);
CREATE INDEX IF NOT EXISTS idx_house_escrows_house_id ON house_escrows(house_id);
CREATE INDEX IF NOT EXISTS idx_house_escrows_listing_id ON house_escrows(listing_id);
CREATE INDEX IF NOT EXISTS idx_house_escrows_contract_id ON house_escrows(contract_id);
CREATE INDEX IF NOT EXISTS idx_house_escrows_community_id ON house_escrows(community_id);
CREATE INDEX IF NOT EXISTS idx_house_escrows_payer_id ON house_escrows(payer_id);
CREATE INDEX IF NOT EXISTS idx_house_escrows_payee_id ON house_escrows(payee_id);
CREATE INDEX IF NOT EXISTS idx_house_escrows_agent_id ON house_escrows(agent_id);
CREATE INDEX IF NOT EXISTS idx_house_escrows_status ON house_escrows(status);
CREATE INDEX IF NOT EXISTS idx_house_escrows_pay_trade_no ON house_escrows(pay_trade_no);
CREATE INDEX IF NOT EXISTS idx_house_escrows_paid_at ON house_escrows(paid_at);
CREATE INDEX IF NOT EXISTS idx_house_escrows_frozen_at ON house_escrows(frozen_at);
CREATE INDEX IF NOT EXISTS idx_house_escrows_release_at ON house_escrows(release_at);
CREATE INDEX IF NOT EXISTS idx_house_escrows_refunded_at ON house_escrows(refunded_at);
CREATE INDEX IF NOT EXISTS idx_house_escrows_auto_release_at ON house_escrows(auto_release_at);
CREATE INDEX IF NOT EXISTS idx_house_escrows_deleted_at ON house_escrows(deleted_at);
COMMENT ON TABLE house_escrows IS '担保交易表（定金/中介费/资金托管/解冻/放款/仲裁）';

-- ------------------------------------------------------------
-- 18. house_deals 成交记录表
--    对标贝壳/链家：成交价/周期/历史/买卖/租赁
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS house_deals (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    deal_no VARCHAR(64) NOT NULL,
    house_id BIGINT NOT NULL,
    listing_id BIGINT NOT NULL DEFAULT 0,
    community_id BIGINT NOT NULL DEFAULT 0,
    contract_id BIGINT NOT NULL DEFAULT 0,
    escrow_id BIGINT NOT NULL DEFAULT 0,
    deal_type VARCHAR(16) NOT NULL DEFAULT 'sale',
    seller_id BIGINT NOT NULL DEFAULT 0,
    seller_name VARCHAR(50) NOT NULL DEFAULT '',
    buyer_id BIGINT NOT NULL DEFAULT 0,
    buyer_name VARCHAR(50) NOT NULL DEFAULT '',
    agent_id BIGINT NOT NULL DEFAULT 0,
    agent_name VARCHAR(50) NOT NULL DEFAULT '',
    deal_price DECIMAL(14,2) NOT NULL DEFAULT 0,
    average_price DECIMAL(10,2) NOT NULL DEFAULT 0,
    commission DECIMAL(14,2) NOT NULL DEFAULT 0,
    deal_date DATE,
    listed_at TIMESTAMPTZ,
    deal_days INT NOT NULL DEFAULT 0,
    payment_method VARCHAR(32) NOT NULL DEFAULT '',
    loan_amount DECIMAL(14,2) NOT NULL DEFAULT 0,
    loan_periods INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 1,
    completed_at TIMESTAMPTZ,
    canceled_at TIMESTAMPTZ,
    canceled_reason VARCHAR(500) NOT NULL DEFAULT '',
    remark VARCHAR(500) NOT NULL DEFAULT '',
    CONSTRAINT uniq_house_deals_no UNIQUE (deal_no)
);
CREATE INDEX IF NOT EXISTS idx_house_deals_region_id ON house_deals(region_id);
CREATE INDEX IF NOT EXISTS idx_house_deals_house_id ON house_deals(house_id);
CREATE INDEX IF NOT EXISTS idx_house_deals_listing_id ON house_deals(listing_id);
CREATE INDEX IF NOT EXISTS idx_house_deals_community_id ON house_deals(community_id);
CREATE INDEX IF NOT EXISTS idx_house_deals_contract_id ON house_deals(contract_id);
CREATE INDEX IF NOT EXISTS idx_house_deals_deal_type ON house_deals(deal_type);
CREATE INDEX IF NOT EXISTS idx_house_deals_seller_id ON house_deals(seller_id);
CREATE INDEX IF NOT EXISTS idx_house_deals_buyer_id ON house_deals(buyer_id);
CREATE INDEX IF NOT EXISTS idx_house_deals_agent_id ON house_deals(agent_id);
CREATE INDEX IF NOT EXISTS idx_house_deals_deal_price ON house_deals(deal_price);
CREATE INDEX IF NOT EXISTS idx_house_deals_deal_date ON house_deals(deal_date);
CREATE INDEX IF NOT EXISTS idx_house_deals_status ON house_deals(status);
CREATE INDEX IF NOT EXISTS idx_house_deals_completed_at ON house_deals(completed_at);
CREATE INDEX IF NOT EXISTS idx_house_deals_deleted_at ON house_deals(deleted_at);
COMMENT ON TABLE house_deals IS '成交记录表（成交价/周期/历史/买卖/租赁/贷款）';

-- ------------------------------------------------------------
-- 19. house_recommendations 推荐记录表
--    对标贝壳：AI 智能推荐（人房匹配）
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS house_recommendations (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id BIGINT NOT NULL,
    house_id BIGINT NOT NULL,
    rec_type VARCHAR(32) NOT NULL DEFAULT 'house_to_user',
    source VARCHAR(32) NOT NULL DEFAULT 'ai',
    score DECIMAL(5,2) NOT NULL DEFAULT 0,
    reason VARCHAR(500) NOT NULL DEFAULT '',
    price_match DECIMAL(5,2) NOT NULL DEFAULT 0,
    location_match DECIMAL(5,2) NOT NULL DEFAULT 0,
    layout_match DECIMAL(5,2) NOT NULL DEFAULT 0,
    facility_match DECIMAL(5,2) NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 0,
    clicked_at TIMESTAMPTZ,
    contacted_at TIMESTAMPTZ,
    viewed_at TIMESTAMPTZ,
    dismissed_at TIMESTAMPTZ,
    expired_at TIMESTAMPTZ,
    CONSTRAINT uniq_house_recs_user_house_type UNIQUE (user_id, house_id, rec_type)
);
CREATE INDEX IF NOT EXISTS idx_house_recommendations_region_id ON house_recommendations(region_id);
CREATE INDEX IF NOT EXISTS idx_house_recommendations_user_id ON house_recommendations(user_id);
CREATE INDEX IF NOT EXISTS idx_house_recommendations_house_id ON house_recommendations(house_id);
CREATE INDEX IF NOT EXISTS idx_house_recommendations_rec_type ON house_recommendations(rec_type);
CREATE INDEX IF NOT EXISTS idx_house_recommendations_source ON house_recommendations(source);
CREATE INDEX IF NOT EXISTS idx_house_recommendations_score ON house_recommendations(score);
CREATE INDEX IF NOT EXISTS idx_house_recommendations_status ON house_recommendations(status);
CREATE INDEX IF NOT EXISTS idx_house_recommendations_clicked_at ON house_recommendations(clicked_at);
CREATE INDEX IF NOT EXISTS idx_house_recommendations_contacted_at ON house_recommendations(contacted_at);
CREATE INDEX IF NOT EXISTS idx_house_recommendations_viewed_at ON house_recommendations(viewed_at);
CREATE INDEX IF NOT EXISTS idx_house_recommendations_dismissed_at ON house_recommendations(dismissed_at);
CREATE INDEX IF NOT EXISTS idx_house_recommendations_expired_at ON house_recommendations(expired_at);
CREATE INDEX IF NOT EXISTS idx_house_recommendations_deleted_at ON house_recommendations(deleted_at);
COMMENT ON TABLE house_recommendations IS '推荐记录表（AI 智能推荐/人房匹配/多维评分）';

-- ============================================================
-- 第三部分：为 19 张表挂载 updated_at 触发器
--   参考 001_p0_baseline.sql 中的 update_updated_at_column 函数
--   幂等：先 DROP IF EXISTS 再 CREATE
-- ============================================================
DO $$
DECLARE t TEXT;
BEGIN
    FOR t IN SELECT tablename FROM pg_tables WHERE schemaname='public' AND tablename IN (
        'house_communities','house_agents','house_listings','house_contracts','house_viewings',
        'house_facilities','house_images','house_vr_tours','house_categories','house_favorites',
        'house_views','house_reports','house_reviews','house_mortgages','house_audit_rules',
        'house_statistics','house_escrows','house_deals','house_recommendations'
    )
    LOOP
        EXECUTE format('DROP TRIGGER IF EXISTS trg_%s_updated ON %s', t, t);
        EXECUTE format('CREATE TRIGGER trg_%s_updated BEFORE UPDATE ON %s FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()', t, t);
    END LOOP;
END $$;
