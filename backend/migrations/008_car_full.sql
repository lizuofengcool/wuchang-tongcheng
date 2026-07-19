-- ============================================================
-- car 车辆买卖模块完整功能迁移脚本（v3.2.1）
-- 对标：瓜子 / 人人车 / 懂车帝 / 毛豆新车 / 易鑫车贷
--
-- 内容：
--   1. ALTER TABLE cars 主表新增 35+ 字段（品牌/型号/年份/里程/排量/变速/燃油/颜色/车况/过户/年检/保险/视频/VR/风控/运营）
--   2. CREATE 19 张子表（car_ 前缀，依据数据库分表前缀规范 v1.0.0）
--   3. 索引、外键、触发器、注释
--   4. 全幂等：CREATE TABLE IF NOT EXISTS / ALTER TABLE ADD COLUMN IF NOT EXISTS
-- 依赖：001_p0_baseline.sql（update_updated_at_column 函数已创建）
-- ============================================================

-- ============================================================
-- 第一部分：扩展 cars 主表（保持现有表名兼容已发布数据）
-- 注意：cars 表由 GORM AutoMigrate 在应用启动时创建
--      本迁移对 cars 的所有 ALTER 操作包装在 DO 块中，
--      若表不存在则跳过（待应用启动后再执行一次本迁移即可补齐字段）
-- ============================================================
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'cars') THEN
        -- === 发布类型/来源类型 ===
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS listing_type VARCHAR(16) NOT NULL DEFAULT 'used';
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS source_type VARCHAR(16) NOT NULL DEFAULT 'personal';
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS car_type VARCHAR(32) NOT NULL DEFAULT 'sedan';
        CREATE INDEX IF NOT EXISTS idx_cars_listing_type ON cars(listing_type);
        CREATE INDEX IF NOT EXISTS idx_cars_source_type ON cars(source_type);
        CREATE INDEX IF NOT EXISTS idx_cars_car_type ON cars(car_type);

        -- === 品牌型号关联 ===
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS brand_id BIGINT;
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS brand_name VARCHAR(64) NOT NULL DEFAULT '';
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS model_id BIGINT;
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS model_name VARCHAR(128) NOT NULL DEFAULT '';
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS series VARCHAR(64) NOT NULL DEFAULT '';
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS category_id BIGINT;
        CREATE INDEX IF NOT EXISTS idx_cars_brand_id ON cars(brand_id);
        CREATE INDEX IF NOT EXISTS idx_cars_brand_name ON cars(brand_name);
        CREATE INDEX IF NOT EXISTS idx_cars_model_id ON cars(model_id);
        CREATE INDEX IF NOT EXISTS idx_cars_category_id ON cars(category_id);

        -- === 价格 ===
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS price DECIMAL(14,2) NOT NULL DEFAULT 0;
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS original_price DECIMAL(14,2) NOT NULL DEFAULT 0;
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS average_price DECIMAL(10,2) NOT NULL DEFAULT 0;
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS price_negotiable BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS dealer_price DECIMAL(14,2) NOT NULL DEFAULT 0;
        CREATE INDEX IF NOT EXISTS idx_cars_price ON cars(price);
        CREATE INDEX IF NOT EXISTS idx_cars_average_price ON cars(average_price);

        -- === 上牌时间/里程 ===
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS registration_year INT NOT NULL DEFAULT 0;
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS registration_month INT NOT NULL DEFAULT 0;
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS first_registration_date DATE;
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS mileage DECIMAL(10,1) NOT NULL DEFAULT 0;
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS mileage_unit VARCHAR(16) NOT NULL DEFAULT 'km';
        CREATE INDEX IF NOT EXISTS idx_cars_registration_year ON cars(registration_year);
        CREATE INDEX IF NOT EXISTS idx_cars_mileage ON cars(mileage);

        -- === 排量/变速/燃油 ===
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS displacement DECIMAL(4,2) NOT NULL DEFAULT 0;
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS transmission VARCHAR(32) NOT NULL DEFAULT '';
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS fuel_type VARCHAR(32) NOT NULL DEFAULT 'gasoline';
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS emission_standard VARCHAR(32) NOT NULL DEFAULT '';
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS engine_type VARCHAR(64) NOT NULL DEFAULT '';
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS horsepower INT NOT NULL DEFAULT 0;
        CREATE INDEX IF NOT EXISTS idx_cars_displacement ON cars(displacement);
        CREATE INDEX IF NOT EXISTS idx_cars_transmission ON cars(transmission);
        CREATE INDEX IF NOT EXISTS idx_cars_fuel_type ON cars(fuel_type);
        CREATE INDEX IF NOT EXISTS idx_cars_emission_standard ON cars(emission_standard);

        -- === 颜色/座位/车门 ===
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS exterior_color VARCHAR(32) NOT NULL DEFAULT '';
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS interior_color VARCHAR(32) NOT NULL DEFAULT '';
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS seat_count INT NOT NULL DEFAULT 5;
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS door_count INT NOT NULL DEFAULT 4;

        -- === 车况 ===
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS condition_level VARCHAR(16) NOT NULL DEFAULT 'A';
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS condition_score INT NOT NULL DEFAULT 0;
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS accident_count INT NOT NULL DEFAULT 0;
        CREATE INDEX IF NOT EXISTS idx_cars_condition_level ON cars(condition_level);

        -- === 过户/年检/保险 ===
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS transfer_count INT NOT NULL DEFAULT 0;
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS last_transfer_date DATE;
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS annual_inspection_due DATE;
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS annual_inspection_status VARCHAR(16) NOT NULL DEFAULT 'valid';
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS insurance_due DATE;
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS insurance_status VARCHAR(16) NOT NULL DEFAULT 'valid';
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS commercial_insurance_due DATE;
        CREATE INDEX IF NOT EXISTS idx_cars_transfer_count ON cars(transfer_count);
        CREATE INDEX IF NOT EXISTS idx_cars_annual_inspection_due ON cars(annual_inspection_due);
        CREATE INDEX IF NOT EXISTS idx_cars_insurance_due ON cars(insurance_due);

        -- === 车架号/车牌 ===
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS vin VARCHAR(32) NOT NULL DEFAULT '';
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS license_plate VARCHAR(32) NOT NULL DEFAULT '';
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS license_location VARCHAR(64) NOT NULL DEFAULT '';
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS engine_no VARCHAR(64) NOT NULL DEFAULT '';
        CREATE INDEX IF NOT EXISTS idx_cars_vin ON cars(vin);
        CREATE INDEX IF NOT EXISTS idx_cars_license_plate ON cars(license_plate);

        -- === 使用性质 ===
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS use_type VARCHAR(32) NOT NULL DEFAULT 'non_operational';
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS maintenance_count INT NOT NULL DEFAULT 0;
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS last_maintenance_mileage DECIMAL(10,1) NOT NULL DEFAULT 0;

        -- === 地理位置 ===
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS city VARCHAR(64) NOT NULL DEFAULT '';
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS district VARCHAR(64) NOT NULL DEFAULT '';
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS business_district VARCHAR(128) NOT NULL DEFAULT '';
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS address VARCHAR(500) NOT NULL DEFAULT '';
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS latitude DECIMAL(10,7) NOT NULL DEFAULT 0;
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS longitude DECIMAL(10,7) NOT NULL DEFAULT 0;
        CREATE INDEX IF NOT EXISTS idx_cars_city ON cars(city);
        CREATE INDEX IF NOT EXISTS idx_cars_district ON cars(district);

        -- === 互动统计 ===
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS view_count INT NOT NULL DEFAULT 0;
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS fav_count INT NOT NULL DEFAULT 0;
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS contact_count INT NOT NULL DEFAULT 0;
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS share_count INT NOT NULL DEFAULT 0;
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS test_drive_count INT NOT NULL DEFAULT 0;
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS last_test_drive_at TIMESTAMPTZ;

        -- === 风控 ===
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS content_hash VARCHAR(64);
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS risk_score INT NOT NULL DEFAULT 0;
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS same_car_id VARCHAR(64);
        CREATE INDEX IF NOT EXISTS idx_cars_content_hash ON cars(content_hash);
        CREATE INDEX IF NOT EXISTS idx_cars_risk_score ON cars(risk_score);
        CREATE INDEX IF NOT EXISTS idx_cars_same_car_id ON cars(same_car_id);

        -- === 视频/360° ===
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS video_url VARCHAR(255);
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS video_cover VARCHAR(255);
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS panorama_360_url VARCHAR(255);

        -- === 配置/特征（JSONB） ===
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS features JSONB;
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS tags JSONB;
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS inspection_items JSONB;
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS accident_history JSONB;
        CREATE INDEX IF NOT EXISTS idx_cars_features ON cars USING GIN(features);
        CREATE INDEX IF NOT EXISTS idx_cars_tags ON cars USING GIN(tags);

        -- === 运营字段 ===
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS featured BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS picked BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS verified BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS promotion_level INT NOT NULL DEFAULT 0;
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS traffic_weight DECIMAL(3,2) NOT NULL DEFAULT 1.00;
        CREATE INDEX IF NOT EXISTS idx_cars_featured ON cars(featured) WHERE featured = TRUE;
        CREATE INDEX IF NOT EXISTS idx_cars_picked ON cars(picked) WHERE picked = TRUE;
        CREATE INDEX IF NOT EXISTS idx_cars_verified ON cars(verified) WHERE verified = TRUE;

        -- === 真车认证 ===
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS real_car_verified BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE cars ADD COLUMN IF NOT EXISTS real_car_verified_at TIMESTAMPTZ;
        CREATE INDEX IF NOT EXISTS idx_cars_real_car_verified ON cars(real_car_verified) WHERE real_car_verified = TRUE;

        -- 字段注释
        COMMENT ON COLUMN cars.listing_type IS '发布类型：new新车/used二手/replace置换/rental租车';
        COMMENT ON COLUMN cars.source_type IS '来源：personal个人/dealer车商/manufacturer厂商';
        COMMENT ON COLUMN cars.car_type IS '车型：sedan轿车/suv/mpv/new_energy/sports/truck';
        COMMENT ON COLUMN cars.mileage IS '里程（公里）';
        COMMENT ON COLUMN cars.displacement IS '排量（L，如1.5/2.0）';
        COMMENT ON COLUMN cars.transmission IS '变速箱：manual手动/auto自动/cvt/dct双离合';
        COMMENT ON COLUMN cars.fuel_type IS '燃油：gasoline汽油/diesel柴油/hybrid混动/pure_electric纯电/range_extender增程';
        COMMENT ON COLUMN cars.condition_level IS '车况等级：A极佳/B良好/C一般/D较差';
        COMMENT ON COLUMN cars.vin IS '车架号（17位）';
        COMMENT ON COLUMN cars.features IS '配置特征 JSON：天窗/导航/真皮/倒车雷达等';
        COMMENT ON COLUMN cars.inspection_items IS '254项检测报告 JSON';
        COMMENT ON COLUMN cars.accident_history IS '事故历史 JSON';
    END IF;
END $$;

-- ============================================================
-- 第二部分：19 张子表 CREATE TABLE IF NOT EXISTS
-- 表前缀 car_ 依据 docs/架构设计/数据库分表前缀规范.md
-- ============================================================

-- ------------------------------------------------------------
-- 1. car_models 车型库表
--    对标懂车帝/汽车之家：品牌/系列/年款/配置/指导价
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS car_models (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    brand VARCHAR(64) NOT NULL DEFAULT '',
    brand_logo VARCHAR(255) NOT NULL DEFAULT '',
    series VARCHAR(64) NOT NULL DEFAULT '',
    model_name VARCHAR(128) NOT NULL,
    year INT NOT NULL DEFAULT 0,
    trim VARCHAR(64) NOT NULL DEFAULT '',
    car_type VARCHAR(32) NOT NULL DEFAULT 'sedan',
    displacement DECIMAL(4,2) NOT NULL DEFAULT 0,
    transmission VARCHAR(32) NOT NULL DEFAULT '',
    fuel_type VARCHAR(32) NOT NULL DEFAULT 'gasoline',
    emission_standard VARCHAR(32) NOT NULL DEFAULT '',
    seat_count INT NOT NULL DEFAULT 5,
    door_count INT NOT NULL DEFAULT 4,
    exterior_color VARCHAR(32) NOT NULL DEFAULT '',
    interior_color VARCHAR(32) NOT NULL DEFAULT '',
    guide_price DECIMAL(14,2) NOT NULL DEFAULT 0,
    depreciation_rate DECIMAL(5,2) NOT NULL DEFAULT 0,
    engine_type VARCHAR(64) NOT NULL DEFAULT '',
    horsepower INT NOT NULL DEFAULT 0,
    description TEXT NOT NULL DEFAULT '',
    cover_image VARCHAR(255) NOT NULL DEFAULT '',
    status INT NOT NULL DEFAULT 1,
    sort INT NOT NULL DEFAULT 0,
    use_count INT NOT NULL DEFAULT 0,
    CONSTRAINT uniq_car_models_name_year_trim UNIQUE (model_name, year, trim)
);
CREATE INDEX IF NOT EXISTS idx_car_models_brand ON car_models(brand);
CREATE INDEX IF NOT EXISTS idx_car_models_series ON car_models(series);
CREATE INDEX IF NOT EXISTS idx_car_models_year ON car_models(year);
CREATE INDEX IF NOT EXISTS idx_car_models_car_type ON car_models(car_type);
CREATE INDEX IF NOT EXISTS idx_car_models_fuel_type ON car_models(fuel_type);
CREATE INDEX IF NOT EXISTS idx_car_models_status ON car_models(status);
CREATE INDEX IF NOT EXISTS idx_car_models_sort ON car_models(sort);
CREATE INDEX IF NOT EXISTS idx_car_models_deleted_at ON car_models(deleted_at);
COMMENT ON TABLE car_models IS '车型库表（品牌/系列/年款/配置/指导价/折旧率）';

-- ------------------------------------------------------------
-- 2. car_inspections 车况检测表
--    对标瓜子：254项检测/检测师/检测报告/三级复核
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS car_inspections (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    inspection_no VARCHAR(64) NOT NULL,
    car_id BIGINT NOT NULL,
    listing_id BIGINT NOT NULL DEFAULT 0,
    inspector_id BIGINT NOT NULL,
    inspector_name VARCHAR(50) NOT NULL DEFAULT '',
    inspector_level VARCHAR(16) NOT NULL DEFAULT 'junior',
    inspection_type VARCHAR(32) NOT NULL DEFAULT 'standard',
    total_items INT NOT NULL DEFAULT 254,
    passed_items INT NOT NULL DEFAULT 0,
    failed_items INT NOT NULL DEFAULT 0,
    warning_items INT NOT NULL DEFAULT 0,
    overall_score DECIMAL(5,2) NOT NULL DEFAULT 0,
    condition_level VARCHAR(16) NOT NULL DEFAULT 'A',
    exterior_score DECIMAL(5,2) NOT NULL DEFAULT 0,
    interior_score DECIMAL(5,2) NOT NULL DEFAULT 0,
    engine_score DECIMAL(5,2) NOT NULL DEFAULT 0,
    chassis_score DECIMAL(5,2) NOT NULL DEFAULT 0,
    electronics_score DECIMAL(5,2) NOT NULL DEFAULT 0,
    safety_score DECIMAL(5,2) NOT NULL DEFAULT 0,
    items JSONB,
    accident_history JSONB,
    has_accident BOOLEAN NOT NULL DEFAULT FALSE,
    has_flood BOOLEAN NOT NULL DEFAULT FALSE,
    has_fire BOOLEAN NOT NULL DEFAULT FALSE,
    has_overhaul BOOLEAN NOT NULL DEFAULT FALSE,
    report_url VARCHAR(255) NOT NULL DEFAULT '',
    report_images JSONB,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    reviewed_by BIGINT NOT NULL DEFAULT 0,
    reviewed_at TIMESTAMPTZ,
    status INT NOT NULL DEFAULT 0,
    remark TEXT NOT NULL DEFAULT '',
    CONSTRAINT uniq_car_inspections_no UNIQUE (inspection_no)
);
CREATE INDEX IF NOT EXISTS idx_car_inspections_region_id ON car_inspections(region_id);
CREATE INDEX IF NOT EXISTS idx_car_inspections_car_id ON car_inspections(car_id);
CREATE INDEX IF NOT EXISTS idx_car_inspections_listing_id ON car_inspections(listing_id);
CREATE INDEX IF NOT EXISTS idx_car_inspections_inspector_id ON car_inspections(inspector_id);
CREATE INDEX IF NOT EXISTS idx_car_inspections_inspection_type ON car_inspections(inspection_type);
CREATE INDEX IF NOT EXISTS idx_car_inspections_condition_level ON car_inspections(condition_level);
CREATE INDEX IF NOT EXISTS idx_car_inspections_has_accident ON car_inspections(has_accident) WHERE has_accident = TRUE;
CREATE INDEX IF NOT EXISTS idx_car_inspections_status ON car_inspections(status);
CREATE INDEX IF NOT EXISTS idx_car_inspections_completed_at ON car_inspections(completed_at);
CREATE INDEX IF NOT EXISTS idx_car_inspections_deleted_at ON car_inspections(deleted_at);
COMMENT ON TABLE car_inspections IS '车况检测表（254项检测/检测师/三级评分/事故记录）';

-- ------------------------------------------------------------
-- 3. car_listings 车源发布表
--    对标瓜子：新车/二手/置换/租车 + 车商/个人
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS car_listings (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    listing_no VARCHAR(64) NOT NULL,
    car_id BIGINT NOT NULL,
    model_id BIGINT NOT NULL DEFAULT 0,
    publisher_id BIGINT NOT NULL,
    publisher_name VARCHAR(50) NOT NULL DEFAULT '',
    publisher_avatar VARCHAR(255) NOT NULL DEFAULT '',
    publisher_type VARCHAR(16) NOT NULL DEFAULT 'personal',
    dealer_id BIGINT NOT NULL DEFAULT 0,
    dealer_name VARCHAR(128) NOT NULL DEFAULT '',
    listing_type VARCHAR(16) NOT NULL DEFAULT 'used',
    title VARCHAR(200) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    price DECIMAL(14,2) NOT NULL DEFAULT 0,
    original_price DECIMAL(14,2) NOT NULL DEFAULT 0,
    price_negotiable BOOLEAN NOT NULL DEFAULT FALSE,
    status INT NOT NULL DEFAULT 0,
    audit_status INT NOT NULL DEFAULT 0,
    audit_reason VARCHAR(500) NOT NULL DEFAULT '',
    published_at TIMESTAMPTZ,
    offline_at TIMESTAMPTZ,
    expired_at TIMESTAMPTZ,
    sold_at TIMESTAMPTZ,
    view_count INT NOT NULL DEFAULT 0,
    fav_count INT NOT NULL DEFAULT 0,
    contact_count INT NOT NULL DEFAULT 0,
    test_drive_count INT NOT NULL DEFAULT 0,
    inspection_status INT NOT NULL DEFAULT 0,
    inspection_id BIGINT NOT NULL DEFAULT 0,
    real_car_verified BOOLEAN NOT NULL DEFAULT FALSE,
    featured BOOLEAN NOT NULL DEFAULT FALSE,
    promotion_level INT NOT NULL DEFAULT 0,
    CONSTRAINT uniq_car_listings_no UNIQUE (listing_no)
);
CREATE INDEX IF NOT EXISTS idx_car_listings_region_id ON car_listings(region_id);
CREATE INDEX IF NOT EXISTS idx_car_listings_car_id ON car_listings(car_id);
CREATE INDEX IF NOT EXISTS idx_car_listings_model_id ON car_listings(model_id);
CREATE INDEX IF NOT EXISTS idx_car_listings_publisher_id ON car_listings(publisher_id);
CREATE INDEX IF NOT EXISTS idx_car_listings_publisher_type ON car_listings(publisher_type);
CREATE INDEX IF NOT EXISTS idx_car_listings_dealer_id ON car_listings(dealer_id);
CREATE INDEX IF NOT EXISTS idx_car_listings_listing_type ON car_listings(listing_type);
CREATE INDEX IF NOT EXISTS idx_car_listings_price ON car_listings(price);
CREATE INDEX IF NOT EXISTS idx_car_listings_status ON car_listings(status);
CREATE INDEX IF NOT EXISTS idx_car_listings_audit_status ON car_listings(audit_status);
CREATE INDEX IF NOT EXISTS idx_car_listings_published_at ON car_listings(published_at);
CREATE INDEX IF NOT EXISTS idx_car_listings_inspection_status ON car_listings(inspection_status);
CREATE INDEX IF NOT EXISTS idx_car_listings_real_car_verified ON car_listings(real_car_verified) WHERE real_car_verified = TRUE;
CREATE INDEX IF NOT EXISTS idx_car_listings_featured ON car_listings(featured) WHERE featured = TRUE;
CREATE INDEX IF NOT EXISTS idx_car_listings_deleted_at ON car_listings(deleted_at);
COMMENT ON TABLE car_listings IS '车源发布表（新车/二手/置换/租车 + 车商/个人 + 审核/检测状态）';

-- ------------------------------------------------------------
-- 4. car_test_drives 试驾预约表
--    对标懂车帝/汽车之家：时间/地点/用户/结果
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS car_test_drives (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    drive_no VARCHAR(64) NOT NULL,
    car_id BIGINT NOT NULL,
    listing_id BIGINT NOT NULL DEFAULT 0,
    user_id BIGINT NOT NULL,
    user_name VARCHAR(50) NOT NULL DEFAULT '',
    user_phone VARCHAR(20) NOT NULL DEFAULT '',
    user_avatar VARCHAR(255) NOT NULL DEFAULT '',
    dealer_id BIGINT NOT NULL DEFAULT 0,
    dealer_name VARCHAR(128) NOT NULL DEFAULT '',
    sales_id BIGINT NOT NULL DEFAULT 0,
    sales_name VARCHAR(50) NOT NULL DEFAULT '',
    appointment_date DATE NOT NULL,
    appointment_time VARCHAR(32) NOT NULL DEFAULT '',
    address VARCHAR(500) NOT NULL DEFAULT '',
    latitude DECIMAL(10,7) NOT NULL DEFAULT 0,
    longitude DECIMAL(10,7) NOT NULL DEFAULT 0,
    drive_type VARCHAR(32) NOT NULL DEFAULT 'test_drive',
    license_status VARCHAR(16) NOT NULL DEFAULT 'unsubmitted',
    license_images JSONB,
    remark VARCHAR(500) NOT NULL DEFAULT '',
    status INT NOT NULL DEFAULT 0,
    cancel_reason VARCHAR(500) NOT NULL DEFAULT '',
    result VARCHAR(32) NOT NULL DEFAULT '',
    result_remark TEXT NOT NULL DEFAULT '',
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    canceled_at TIMESTAMPTZ,
    CONSTRAINT uniq_car_test_drives_no UNIQUE (drive_no)
);
CREATE INDEX IF NOT EXISTS idx_car_test_drives_region_id ON car_test_drives(region_id);
CREATE INDEX IF NOT EXISTS idx_car_test_drives_car_id ON car_test_drives(car_id);
CREATE INDEX IF NOT EXISTS idx_car_test_drives_listing_id ON car_test_drives(listing_id);
CREATE INDEX IF NOT EXISTS idx_car_test_drives_user_id ON car_test_drives(user_id);
CREATE INDEX IF NOT EXISTS idx_car_test_drives_dealer_id ON car_test_drives(dealer_id);
CREATE INDEX IF NOT EXISTS idx_car_test_drives_sales_id ON car_test_drives(sales_id);
CREATE INDEX IF NOT EXISTS idx_car_test_drives_appointment_date ON car_test_drives(appointment_date);
CREATE INDEX IF NOT EXISTS idx_car_test_drives_drive_type ON car_test_drives(drive_type);
CREATE INDEX IF NOT EXISTS idx_car_test_drives_status ON car_test_drives(status);
CREATE INDEX IF NOT EXISTS idx_car_test_drives_completed_at ON car_test_drives(completed_at);
CREATE INDEX IF NOT EXISTS idx_car_test_drives_deleted_at ON car_test_drives(deleted_at);
COMMENT ON TABLE car_test_drives IS '试驾预约表（时间/地点/销售/驾照/结果）';

-- ------------------------------------------------------------
-- 5. car_contracts 合同电子化表
--    对标瓜子：买卖合同/置换协议/租车协议 + 电子签
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS car_contracts (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    contract_no VARCHAR(64) NOT NULL,
    contract_type VARCHAR(32) NOT NULL DEFAULT 'sale',
    car_id BIGINT NOT NULL,
    listing_id BIGINT NOT NULL DEFAULT 0,
    seller_id BIGINT NOT NULL,
    seller_name VARCHAR(50) NOT NULL DEFAULT '',
    seller_phone VARCHAR(20) NOT NULL DEFAULT '',
    seller_id_card VARCHAR(32) NOT NULL DEFAULT '',
    buyer_id BIGINT NOT NULL,
    buyer_name VARCHAR(50) NOT NULL DEFAULT '',
    buyer_phone VARCHAR(20) NOT NULL DEFAULT '',
    buyer_id_card VARCHAR(32) NOT NULL DEFAULT '',
    agent_id BIGINT NOT NULL DEFAULT 0,
    agent_name VARCHAR(50) NOT NULL DEFAULT '',
    deal_price DECIMAL(14,2) NOT NULL DEFAULT 0,
    deposit DECIMAL(14,2) NOT NULL DEFAULT 0,
    payment_method VARCHAR(32) NOT NULL DEFAULT 'full',
    loan_amount DECIMAL(14,2) NOT NULL DEFAULT 0,
    loan_periods INT NOT NULL DEFAULT 0,
    transfer_fee DECIMAL(14,2) NOT NULL DEFAULT 0,
    service_fee DECIMAL(14,2) NOT NULL DEFAULT 0,
    contract_url VARCHAR(255) NOT NULL DEFAULT '',
    attachments JSONB,
    signed_at TIMESTAMPTZ,
    effective_at TIMESTAMPTZ,
    expired_at TIMESTAMPTZ,
    terminated_at TIMESTAMPTZ,
    status INT NOT NULL DEFAULT 0,
    remark TEXT NOT NULL DEFAULT '',
    CONSTRAINT uniq_car_contracts_no UNIQUE (contract_no)
);
CREATE INDEX IF NOT EXISTS idx_car_contracts_region_id ON car_contracts(region_id);
CREATE INDEX IF NOT EXISTS idx_car_contracts_contract_type ON car_contracts(contract_type);
CREATE INDEX IF NOT EXISTS idx_car_contracts_car_id ON car_contracts(car_id);
CREATE INDEX IF NOT EXISTS idx_car_contracts_listing_id ON car_contracts(listing_id);
CREATE INDEX IF NOT EXISTS idx_car_contracts_seller_id ON car_contracts(seller_id);
CREATE INDEX IF NOT EXISTS idx_car_contracts_buyer_id ON car_contracts(buyer_id);
CREATE INDEX IF NOT EXISTS idx_car_contracts_agent_id ON car_contracts(agent_id);
CREATE INDEX IF NOT EXISTS idx_car_contracts_deal_price ON car_contracts(deal_price);
CREATE INDEX IF NOT EXISTS idx_car_contracts_status ON car_contracts(status);
CREATE INDEX IF NOT EXISTS idx_car_contracts_signed_at ON car_contracts(signed_at);
CREATE INDEX IF NOT EXISTS idx_car_contracts_deleted_at ON car_contracts(deleted_at);
COMMENT ON TABLE car_contracts IS '合同电子化表（买卖/置换/租车 + 电子签 + 贷款/服务费）';

-- ------------------------------------------------------------
-- 6. car_evaluations 车辆评估表
--    对标瓜子/人人车：估值/折旧/市场行情
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS car_evaluations (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    evaluation_no VARCHAR(64) NOT NULL,
    car_id BIGINT NOT NULL,
    model_id BIGINT NOT NULL DEFAULT 0,
    evaluator_id BIGINT NOT NULL,
    evaluator_name VARCHAR(50) NOT NULL DEFAULT '',
    evaluator_level VARCHAR(16) NOT NULL DEFAULT 'junior',
    evaluation_type VARCHAR(32) NOT NULL DEFAULT 'online',
    market_price DECIMAL(14,2) NOT NULL DEFAULT 0,
    trade_in_price DECIMAL(14,2) NOT NULL DEFAULT 0,
    retail_price DECIMAL(14,2) NOT NULL DEFAULT 0,
    wholesale_price DECIMAL(14,2) NOT NULL DEFAULT 0,
    depreciation_amount DECIMAL(14,2) NOT NULL DEFAULT 0,
    depreciation_rate DECIMAL(5,2) NOT NULL DEFAULT 0,
    final_price DECIMAL(14,2) NOT NULL DEFAULT 0,
    price_range_min DECIMAL(14,2) NOT NULL DEFAULT 0,
    price_range_max DECIMAL(14,2) NOT NULL DEFAULT 0,
    factors JSONB,
    similar_deals JSONB,
    description TEXT NOT NULL DEFAULT '',
    report_url VARCHAR(255) NOT NULL DEFAULT '',
    valid_until DATE,
    status INT NOT NULL DEFAULT 0,
    CONSTRAINT uniq_car_evaluations_no UNIQUE (evaluation_no)
);
CREATE INDEX IF NOT EXISTS idx_car_evaluations_region_id ON car_evaluations(region_id);
CREATE INDEX IF NOT EXISTS idx_car_evaluations_car_id ON car_evaluations(car_id);
CREATE INDEX IF NOT EXISTS idx_car_evaluations_model_id ON car_evaluations(model_id);
CREATE INDEX IF NOT EXISTS idx_car_evaluations_evaluator_id ON car_evaluations(evaluator_id);
CREATE INDEX IF NOT EXISTS idx_car_evaluations_evaluation_type ON car_evaluations(evaluation_type);
CREATE INDEX IF NOT EXISTS idx_car_evaluations_final_price ON car_evaluations(final_price);
CREATE INDEX IF NOT EXISTS idx_car_evaluations_status ON car_evaluations(status);
CREATE INDEX IF NOT EXISTS idx_car_evaluations_valid_until ON car_evaluations(valid_until);
CREATE INDEX IF NOT EXISTS idx_car_evaluations_deleted_at ON car_evaluations(deleted_at);
COMMENT ON TABLE car_evaluations IS '车辆评估表（市场价/收购价/零售价/批发价/折旧/相似成交）';

-- ------------------------------------------------------------
-- 7. car_financing 分期付款表
--    对标易鑫车贷/毛豆新车：首付/月供/利率/期数
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS car_financing (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    name VARCHAR(64) NOT NULL,
    code VARCHAR(64) NOT NULL,
    financing_type VARCHAR(32) NOT NULL DEFAULT 'loan',
    min_down_payment DECIMAL(5,2) NOT NULL DEFAULT 20.00,
    max_down_payment DECIMAL(5,2) NOT NULL DEFAULT 80.00,
    interest_rate DECIMAL(6,4) NOT NULL DEFAULT 0,
    annual_rate DECIMAL(6,4) NOT NULL DEFAULT 0,
    min_periods INT NOT NULL DEFAULT 12,
    max_periods INT NOT NULL DEFAULT 60,
    max_amount DECIMAL(14,2) NOT NULL DEFAULT 0,
    provider VARCHAR(128) NOT NULL DEFAULT '',
    description VARCHAR(500) NOT NULL DEFAULT '',
    sort INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 1,
    is_hot BOOLEAN NOT NULL DEFAULT FALSE,
    use_count INT NOT NULL DEFAULT 0,
    CONSTRAINT uniq_car_financing_code UNIQUE (code)
);
CREATE INDEX IF NOT EXISTS idx_car_financing_financing_type ON car_financing(financing_type);
CREATE INDEX IF NOT EXISTS idx_car_financing_provider ON car_financing(provider);
CREATE INDEX IF NOT EXISTS idx_car_financing_status ON car_financing(status);
CREATE INDEX IF NOT EXISTS idx_car_financing_is_hot ON car_financing(is_hot) WHERE is_hot = TRUE;
CREATE INDEX IF NOT EXISTS idx_car_financing_sort ON car_financing(sort);
CREATE INDEX IF NOT EXISTS idx_car_financing_deleted_at ON car_financing(deleted_at);
COMMENT ON TABLE car_financing IS '分期付款方案表（首付/月供/利率/期数/资方）';

-- ------------------------------------------------------------
-- 8. car_insurance 车险配置表
--    对标平安/太平洋/人保：交强/商业/第三方
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS car_insurance (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    name VARCHAR(64) NOT NULL,
    code VARCHAR(64) NOT NULL,
    insurance_type VARCHAR(32) NOT NULL DEFAULT 'compulsory',
    provider VARCHAR(128) NOT NULL DEFAULT '',
    coverage VARCHAR(64) NOT NULL DEFAULT '',
    coverage_amount DECIMAL(14,2) NOT NULL DEFAULT 0,
    premium DECIMAL(14,2) NOT NULL DEFAULT 0,
    deductible INT NOT NULL DEFAULT 0,
    description VARCHAR(500) NOT NULL DEFAULT '',
    sort INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 1,
    is_hot BOOLEAN NOT NULL DEFAULT FALSE,
    use_count INT NOT NULL DEFAULT 0,
    CONSTRAINT uniq_car_insurance_code UNIQUE (code)
);
CREATE INDEX IF NOT EXISTS idx_car_insurance_insurance_type ON car_insurance(insurance_type);
CREATE INDEX IF NOT EXISTS idx_car_insurance_provider ON car_insurance(provider);
CREATE INDEX IF NOT EXISTS idx_car_insurance_coverage ON car_insurance(coverage);
CREATE INDEX IF NOT EXISTS idx_car_insurance_status ON car_insurance(status);
CREATE INDEX IF NOT EXISTS idx_car_insurance_is_hot ON car_insurance(is_hot) WHERE is_hot = TRUE;
CREATE INDEX IF NOT EXISTS idx_car_insurance_sort ON car_insurance(sort);
CREATE INDEX IF NOT EXISTS idx_car_insurance_deleted_at ON car_insurance(deleted_at);
COMMENT ON TABLE car_insurance IS '车险配置表（交强/商业/第三方/赔付额/免赔额）';

-- ------------------------------------------------------------
-- 9. car_transfer 过户办理表
--    对标瓜子：流程/材料/状态/办理进度
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS car_transfer (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    transfer_no VARCHAR(64) NOT NULL,
    car_id BIGINT NOT NULL,
    contract_id BIGINT NOT NULL DEFAULT 0,
    listing_id BIGINT NOT NULL DEFAULT 0,
    seller_id BIGINT NOT NULL,
    seller_name VARCHAR(50) NOT NULL DEFAULT '',
    buyer_id BIGINT NOT NULL,
    buyer_name VARCHAR(50) NOT NULL DEFAULT '',
    agent_id BIGINT NOT NULL DEFAULT 0,
    agent_name VARCHAR(50) NOT NULL DEFAULT '',
    transfer_type VARCHAR(32) NOT NULL DEFAULT 'sale',
    vehicle_registration JSONB,
    documents JSONB,
    transfer_fee DECIMAL(14,2) NOT NULL DEFAULT 0,
    tax_fee DECIMAL(14,2) NOT NULL DEFAULT 0,
    other_fee DECIMAL(14,2) NOT NULL DEFAULT 0,
    location VARCHAR(255) NOT NULL DEFAULT '',
    appointment_date DATE,
    appointment_time VARCHAR(32) NOT NULL DEFAULT '',
    submitted_at TIMESTAMPTZ,
    reviewed_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    canceled_at TIMESTAMPTZ,
    new_license_plate VARCHAR(32) NOT NULL DEFAULT '',
    new_registration_cert VARCHAR(255) NOT NULL DEFAULT '',
    status INT NOT NULL DEFAULT 0,
    remark TEXT NOT NULL DEFAULT '',
    CONSTRAINT uniq_car_transfer_no UNIQUE (transfer_no)
);
CREATE INDEX IF NOT EXISTS idx_car_transfer_region_id ON car_transfer(region_id);
CREATE INDEX IF NOT EXISTS idx_car_transfer_car_id ON car_transfer(car_id);
CREATE INDEX IF NOT EXISTS idx_car_transfer_contract_id ON car_transfer(contract_id);
CREATE INDEX IF NOT EXISTS idx_car_transfer_listing_id ON car_transfer(listing_id);
CREATE INDEX IF NOT EXISTS idx_car_transfer_seller_id ON car_transfer(seller_id);
CREATE INDEX IF NOT EXISTS idx_car_transfer_buyer_id ON car_transfer(buyer_id);
CREATE INDEX IF NOT EXISTS idx_car_transfer_agent_id ON car_transfer(agent_id);
CREATE INDEX IF NOT EXISTS idx_car_transfer_transfer_type ON car_transfer(transfer_type);
CREATE INDEX IF NOT EXISTS idx_car_transfer_appointment_date ON car_transfer(appointment_date);
CREATE INDEX IF NOT EXISTS idx_car_transfer_status ON car_transfer(status);
CREATE INDEX IF NOT EXISTS idx_car_transfer_completed_at ON car_transfer(completed_at);
CREATE INDEX IF NOT EXISTS idx_car_transfer_deleted_at ON car_transfer(deleted_at);
COMMENT ON TABLE car_transfer IS '过户办理表（材料/流程/状态/新牌照/新登记证）';

-- ------------------------------------------------------------
-- 10. car_categories 车型分类表
--     对标懂车帝：轿车/SUV/MPV/新能源/跑车/皮卡
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS car_categories (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    name VARCHAR(64) NOT NULL,
    code VARCHAR(64) NOT NULL,
    parent_id BIGINT NOT NULL DEFAULT 0,
    level INT NOT NULL DEFAULT 1,
    car_type VARCHAR(32) NOT NULL DEFAULT '',
    icon VARCHAR(64) NOT NULL DEFAULT '',
    color VARCHAR(32) NOT NULL DEFAULT '',
    description VARCHAR(500) NOT NULL DEFAULT '',
    sort INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 1,
    car_count INT NOT NULL DEFAULT 0,
    CONSTRAINT uniq_car_categories_code UNIQUE (code)
);
CREATE INDEX IF NOT EXISTS idx_car_categories_parent_id ON car_categories(parent_id);
CREATE INDEX IF NOT EXISTS idx_car_categories_level ON car_categories(level);
CREATE INDEX IF NOT EXISTS idx_car_categories_car_type ON car_categories(car_type);
CREATE INDEX IF NOT EXISTS idx_car_categories_status ON car_categories(status);
CREATE INDEX IF NOT EXISTS idx_car_categories_sort ON car_categories(sort);
CREATE INDEX IF NOT EXISTS idx_car_categories_deleted_at ON car_categories(deleted_at);
COMMENT ON TABLE car_categories IS '车型分类表（轿车/SUV/MPV/新能源/跑车/皮卡）';

-- ------------------------------------------------------------
-- 11. car_favorites 车源收藏表
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS car_favorites (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id BIGINT NOT NULL,
    car_id BIGINT NOT NULL,
    listing_id BIGINT NOT NULL DEFAULT 0,
    favorite_type VARCHAR(32) NOT NULL DEFAULT 'car',
    remark VARCHAR(200) NOT NULL DEFAULT '',
    CONSTRAINT uniq_car_favorites_user_car_type UNIQUE (user_id, car_id, favorite_type)
);
CREATE INDEX IF NOT EXISTS idx_car_favorites_user_id ON car_favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_car_favorites_car_id ON car_favorites(car_id);
CREATE INDEX IF NOT EXISTS idx_car_favorites_listing_id ON car_favorites(listing_id);
CREATE INDEX IF NOT EXISTS idx_car_favorites_favorite_type ON car_favorites(favorite_type);
CREATE INDEX IF NOT EXISTS idx_car_favorites_deleted_at ON car_favorites(deleted_at);
COMMENT ON TABLE car_favorites IS '车源收藏表';

-- ------------------------------------------------------------
-- 12. car_views 浏览记录表
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS car_views (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id BIGINT NOT NULL DEFAULT 0,
    car_id BIGINT NOT NULL,
    listing_id BIGINT NOT NULL DEFAULT 0,
    ip VARCHAR(64) NOT NULL DEFAULT '',
    user_agent VARCHAR(500) NOT NULL DEFAULT '',
    referer VARCHAR(500) NOT NULL DEFAULT '',
    device VARCHAR(32) NOT NULL DEFAULT '',
    source VARCHAR(32) NOT NULL DEFAULT '',
    duration INT NOT NULL DEFAULT 0,
    longitude DECIMAL(10,7) NOT NULL DEFAULT 0,
    latitude DECIMAL(10,7) NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_car_views_region_id ON car_views(region_id);
CREATE INDEX IF NOT EXISTS idx_car_views_user_id ON car_views(user_id);
CREATE INDEX IF NOT EXISTS idx_car_views_car_id ON car_views(car_id);
CREATE INDEX IF NOT EXISTS idx_car_views_listing_id ON car_views(listing_id);
CREATE INDEX IF NOT EXISTS idx_car_views_created_at ON car_views(created_at);
CREATE INDEX IF NOT EXISTS idx_car_views_source ON car_views(source);
COMMENT ON TABLE car_views IS '浏览记录表';

-- ------------------------------------------------------------
-- 13. car_reports 举报工单表
--     对标瓜子/58：虚假车源/诈骗/色情/侵权 + SLA 24h/72h
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS car_reports (
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
    CONSTRAINT uniq_car_reports_no UNIQUE (report_no)
);
CREATE INDEX IF NOT EXISTS idx_car_reports_target_type_target ON car_reports(target_type, target_id);
CREATE INDEX IF NOT EXISTS idx_car_reports_target_user_id ON car_reports(target_user_id);
CREATE INDEX IF NOT EXISTS idx_car_reports_reporter_id ON car_reports(reporter_id);
CREATE INDEX IF NOT EXISTS idx_car_reports_reported_user_id ON car_reports(reported_user_id);
CREATE INDEX IF NOT EXISTS idx_car_reports_report_type ON car_reports(report_type);
CREATE INDEX IF NOT EXISTS idx_car_reports_status ON car_reports(status);
CREATE INDEX IF NOT EXISTS idx_car_reports_handler_id ON car_reports(handler_id);
CREATE INDEX IF NOT EXISTS idx_car_reports_penalty_user_id ON car_reports(penalty_user_id);
CREATE INDEX IF NOT EXISTS idx_car_reports_sla_deadline ON car_reports(sla_deadline);
CREATE INDEX IF NOT EXISTS idx_car_reports_handled_at ON car_reports(handled_at);
CREATE INDEX IF NOT EXISTS idx_car_reports_appealed_at ON car_reports(appealed_at);
CREATE INDEX IF NOT EXISTS idx_car_reports_deleted_at ON car_reports(deleted_at);
COMMENT ON TABLE car_reports IS '举报工单表（虚假车源/诈骗/色情/侵权 + SLA + 申诉）';

-- ------------------------------------------------------------
-- 14. car_reviews 评价表
--     对标懂车帝：车商/车源 5星+文字+图片+追评+回复
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS car_reviews (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    target_type VARCHAR(32) NOT NULL DEFAULT 'dealer',
    target_id BIGINT NOT NULL,
    reviewer_id BIGINT NOT NULL,
    reviewer_name VARCHAR(50) NOT NULL DEFAULT '',
    reviewer_avatar VARCHAR(255) NOT NULL DEFAULT '',
    review_type VARCHAR(32) NOT NULL DEFAULT 'buyer',
    rating INT NOT NULL DEFAULT 5,
    content TEXT NOT NULL DEFAULT '',
    images JSONB,
    video_url VARCHAR(255) NOT NULL DEFAULT '',
    is_anonymous BOOLEAN NOT NULL DEFAULT FALSE,
    is_recommended BOOLEAN NOT NULL DEFAULT TRUE,
    tags JSONB,
    deal_amount DECIMAL(14,2) NOT NULL DEFAULT 0,
    exterior_rating INT NOT NULL DEFAULT 5,
    interior_rating INT NOT NULL DEFAULT 5,
    engine_rating INT NOT NULL DEFAULT 5,
    paperwork_rating INT NOT NULL DEFAULT 5,
    service_attitude INT NOT NULL DEFAULT 5,
    professional_skill INT NOT NULL DEFAULT 5,
    reply TEXT NOT NULL DEFAULT '',
    reply_at TIMESTAMPTZ,
    append_content TEXT NOT NULL DEFAULT '',
    append_images JSONB,
    append_at TIMESTAMPTZ,
    like_count INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 1,
    CONSTRAINT uniq_car_reviews_target_reviewer UNIQUE (target_type, target_id, reviewer_id)
);
CREATE INDEX IF NOT EXISTS idx_car_reviews_region_id ON car_reviews(region_id);
CREATE INDEX IF NOT EXISTS idx_car_reviews_target ON car_reviews(target_type, target_id);
CREATE INDEX IF NOT EXISTS idx_car_reviews_reviewer_id ON car_reviews(reviewer_id);
CREATE INDEX IF NOT EXISTS idx_car_reviews_review_type ON car_reviews(review_type);
CREATE INDEX IF NOT EXISTS idx_car_reviews_rating ON car_reviews(rating);
CREATE INDEX IF NOT EXISTS idx_car_reviews_is_recommended ON car_reviews(is_recommended) WHERE is_recommended = FALSE;
CREATE INDEX IF NOT EXISTS idx_car_reviews_status ON car_reviews(status);
CREATE INDEX IF NOT EXISTS idx_car_reviews_reply_at ON car_reviews(reply_at);
CREATE INDEX IF NOT EXISTS idx_car_reviews_append_at ON car_reviews(append_at);
CREATE INDEX IF NOT EXISTS idx_car_reviews_deleted_at ON car_reviews(deleted_at);
COMMENT ON TABLE car_reviews IS '评价表（车商/车源 5星+外观+内饰+发动机+手续+服务态度+专业能力+追评+回复）';

-- ------------------------------------------------------------
-- 15. car_audit_rules 审核规则表
--     对标瓜子：车况/价格异常/频率限制/敏感词
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS car_audit_rules (
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
CREATE INDEX IF NOT EXISTS idx_car_audit_rules_rule_type ON car_audit_rules(rule_type);
CREATE INDEX IF NOT EXISTS idx_car_audit_rules_rule_key ON car_audit_rules(rule_key);
CREATE INDEX IF NOT EXISTS idx_car_audit_rules_severity ON car_audit_rules(severity);
CREATE INDEX IF NOT EXISTS idx_car_audit_rules_status ON car_audit_rules(status);
CREATE INDEX IF NOT EXISTS idx_car_audit_rules_deleted_at ON car_audit_rules(deleted_at);
COMMENT ON TABLE car_audit_rules IS '审核规则表（车况/价格异常/频率限制/敏感词）';

-- ------------------------------------------------------------
-- 16. car_statistics 数据统计表
--     对标懂车帝：曝光/点击/收藏/试驾/成交转化
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS car_statistics (
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
    test_drive_count INT NOT NULL DEFAULT 0,
    deal_count INT NOT NULL DEFAULT 0,
    conversion_rate DECIMAL(5,2) NOT NULL DEFAULT 0,
    avg_price DECIMAL(10,2) NOT NULL DEFAULT 0,
    avg_deal_days INT NOT NULL DEFAULT 0,
    CONSTRAINT uniq_car_stats_date_type_target UNIQUE (stat_date, stat_type, target_id)
);
CREATE INDEX IF NOT EXISTS idx_car_statistics_region_id ON car_statistics(region_id);
CREATE INDEX IF NOT EXISTS idx_car_statistics_stat_date ON car_statistics(stat_date);
CREATE INDEX IF NOT EXISTS idx_car_statistics_stat_type ON car_statistics(stat_type);
CREATE INDEX IF NOT EXISTS idx_car_statistics_target_id ON car_statistics(target_id);
CREATE INDEX IF NOT EXISTS idx_car_statistics_deleted_at ON car_statistics(deleted_at);
COMMENT ON TABLE car_statistics IS '数据统计表（曝光/点击/收藏/试驾/成交/转化率/均价/平均成交周期）';

-- ------------------------------------------------------------
-- 17. car_escrows 担保交易表
--     对标瓜子：定金/全款/资金托管/解冻/放款
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS car_escrows (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    escrow_no VARCHAR(64) NOT NULL,
    escrow_type VARCHAR(32) NOT NULL DEFAULT 'deposit',
    car_id BIGINT NOT NULL DEFAULT 0,
    listing_id BIGINT NOT NULL DEFAULT 0,
    contract_id BIGINT NOT NULL DEFAULT 0,
    payer_id BIGINT NOT NULL,
    payee_id BIGINT NOT NULL,
    dealer_id BIGINT NOT NULL DEFAULT 0,
    amount DECIMAL(14,2) NOT NULL,
    platform_fee DECIMAL(14,2) NOT NULL DEFAULT 0,
    dealer_fee DECIMAL(14,2) NOT NULL DEFAULT 0,
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
    CONSTRAINT uniq_car_escrows_no UNIQUE (escrow_no)
);
CREATE INDEX IF NOT EXISTS idx_car_escrows_region_id ON car_escrows(region_id);
CREATE INDEX IF NOT EXISTS idx_car_escrows_escrow_type ON car_escrows(escrow_type);
CREATE INDEX IF NOT EXISTS idx_car_escrows_car_id ON car_escrows(car_id);
CREATE INDEX IF NOT EXISTS idx_car_escrows_listing_id ON car_escrows(listing_id);
CREATE INDEX IF NOT EXISTS idx_car_escrows_contract_id ON car_escrows(contract_id);
CREATE INDEX IF NOT EXISTS idx_car_escrows_payer_id ON car_escrows(payer_id);
CREATE INDEX IF NOT EXISTS idx_car_escrows_payee_id ON car_escrows(payee_id);
CREATE INDEX IF NOT EXISTS idx_car_escrows_dealer_id ON car_escrows(dealer_id);
CREATE INDEX IF NOT EXISTS idx_car_escrows_status ON car_escrows(status);
CREATE INDEX IF NOT EXISTS idx_car_escrows_pay_trade_no ON car_escrows(pay_trade_no);
CREATE INDEX IF NOT EXISTS idx_car_escrows_paid_at ON car_escrows(paid_at);
CREATE INDEX IF NOT EXISTS idx_car_escrows_frozen_at ON car_escrows(frozen_at);
CREATE INDEX IF NOT EXISTS idx_car_escrows_release_at ON car_escrows(release_at);
CREATE INDEX IF NOT EXISTS idx_car_escrows_refunded_at ON car_escrows(refunded_at);
CREATE INDEX IF NOT EXISTS idx_car_escrows_auto_release_at ON car_escrows(auto_release_at);
CREATE INDEX IF NOT EXISTS idx_car_escrows_deleted_at ON car_escrows(deleted_at);
COMMENT ON TABLE car_escrows IS '担保交易表（定金/全款/资金托管/解冻/放款/仲裁）';

-- ------------------------------------------------------------
-- 18. car_images 车源图片表
--     对标瓜子：外观/内饰/发动机/底盘/事故
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS car_images (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    car_id BIGINT NOT NULL,
    listing_id BIGINT NOT NULL DEFAULT 0,
    image_type VARCHAR(32) NOT NULL DEFAULT 'exterior',
    url VARCHAR(500) NOT NULL,
    thumbnail VARCHAR(500) NOT NULL DEFAULT '',
    title VARCHAR(128) NOT NULL DEFAULT '',
    description VARCHAR(500) NOT NULL DEFAULT '',
    sort INT NOT NULL DEFAULT 0,
    is_cover BOOLEAN NOT NULL DEFAULT FALSE,
    width INT NOT NULL DEFAULT 0,
    height INT NOT NULL DEFAULT 0,
    size INT NOT NULL DEFAULT 0,
    tag VARCHAR(64) NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_car_images_region_id ON car_images(region_id);
CREATE INDEX IF NOT EXISTS idx_car_images_car_id ON car_images(car_id);
CREATE INDEX IF NOT EXISTS idx_car_images_listing_id ON car_images(listing_id);
CREATE INDEX IF NOT EXISTS idx_car_images_image_type ON car_images(image_type);
CREATE INDEX IF NOT EXISTS idx_car_images_is_cover ON car_images(is_cover) WHERE is_cover = TRUE;
CREATE INDEX IF NOT EXISTS idx_car_images_sort ON car_images(sort);
CREATE INDEX IF NOT EXISTS idx_car_images_deleted_at ON car_images(deleted_at);
COMMENT ON TABLE car_images IS '车源图片表（外观/内饰/发动机/底盘/事故 + 封面 + 排序）';

-- ------------------------------------------------------------
-- 19. car_recommendations 推荐记录表
--     对标懂车帝：AI 智能推荐（人车匹配）
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS car_recommendations (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id BIGINT NOT NULL,
    car_id BIGINT NOT NULL,
    rec_type VARCHAR(32) NOT NULL DEFAULT 'car_to_user',
    source VARCHAR(32) NOT NULL DEFAULT 'ai',
    score DECIMAL(5,2) NOT NULL DEFAULT 0,
    reason VARCHAR(500) NOT NULL DEFAULT '',
    price_match DECIMAL(5,2) NOT NULL DEFAULT 0,
    brand_match DECIMAL(5,2) NOT NULL DEFAULT 0,
    type_match DECIMAL(5,2) NOT NULL DEFAULT 0,
    condition_match DECIMAL(5,2) NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 0,
    clicked_at TIMESTAMPTZ,
    contacted_at TIMESTAMPTZ,
    viewed_at TIMESTAMPTZ,
    dismissed_at TIMESTAMPTZ,
    expired_at TIMESTAMPTZ,
    CONSTRAINT uniq_car_recs_user_car_type UNIQUE (user_id, car_id, rec_type)
);
CREATE INDEX IF NOT EXISTS idx_car_recommendations_region_id ON car_recommendations(region_id);
CREATE INDEX IF NOT EXISTS idx_car_recommendations_user_id ON car_recommendations(user_id);
CREATE INDEX IF NOT EXISTS idx_car_recommendations_car_id ON car_recommendations(car_id);
CREATE INDEX IF NOT EXISTS idx_car_recommendations_rec_type ON car_recommendations(rec_type);
CREATE INDEX IF NOT EXISTS idx_car_recommendations_source ON car_recommendations(source);
CREATE INDEX IF NOT EXISTS idx_car_recommendations_score ON car_recommendations(score);
CREATE INDEX IF NOT EXISTS idx_car_recommendations_status ON car_recommendations(status);
CREATE INDEX IF NOT EXISTS idx_car_recommendations_clicked_at ON car_recommendations(clicked_at);
CREATE INDEX IF NOT EXISTS idx_car_recommendations_contacted_at ON car_recommendations(contacted_at);
CREATE INDEX IF NOT EXISTS idx_car_recommendations_viewed_at ON car_recommendations(viewed_at);
CREATE INDEX IF NOT EXISTS idx_car_recommendations_dismissed_at ON car_recommendations(dismissed_at);
CREATE INDEX IF NOT EXISTS idx_car_recommendations_expired_at ON car_recommendations(expired_at);
CREATE INDEX IF NOT EXISTS idx_car_recommendations_deleted_at ON car_recommendations(deleted_at);
COMMENT ON TABLE car_recommendations IS '推荐记录表（AI 智能推荐/人车匹配/多维评分）';

-- ============================================================
-- 第三部分：为 19 张表挂载 updated_at 触发器
--   参考 001_p0_baseline.sql 中的 update_updated_at_column 函数
--   幂等：先 DROP IF EXISTS 再 CREATE
-- ============================================================
DO $$
DECLARE t TEXT;
BEGIN
    FOR t IN SELECT tablename FROM pg_tables WHERE schemaname='public' AND tablename IN (
        'car_models','car_inspections','car_listings','car_test_drives','car_contracts',
        'car_evaluations','car_financing','car_insurance','car_transfer','car_categories',
        'car_favorites','car_views','car_reports','car_reviews','car_audit_rules',
        'car_statistics','car_escrows','car_images','car_recommendations'
    )
    LOOP
        EXECUTE format('DROP TRIGGER IF EXISTS trg_%s_updated ON %s', t, t);
        EXECUTE format('CREATE TRIGGER trg_%s_updated BEFORE UPDATE ON %s FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()', t, t);
    END LOOP;
END $$;
