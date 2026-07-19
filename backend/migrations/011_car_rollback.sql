-- ============================================================
-- car 车辆买卖模块完整功能回滚脚本（与 008_car_full.sql 配对）
-- 按反向顺序 DROP 19 张子表 + 触发器 + 主表新增字段
-- 幂等：使用 IF EXISTS，可重复执行
-- 警告：执行后 car 模块全部扩展数据将丢失（主表基础字段保留）
-- ============================================================

-- ------------------------------------------------------------
-- 1. 先移除 19 张子表的 updated_at 触发器
-- ------------------------------------------------------------
DROP TRIGGER IF EXISTS trg_car_recommendations_updated ON car_recommendations;
DROP TRIGGER IF EXISTS trg_car_images_updated ON car_images;
DROP TRIGGER IF EXISTS trg_car_escrows_updated ON car_escrows;
DROP TRIGGER IF EXISTS trg_car_statistics_updated ON car_statistics;
DROP TRIGGER IF EXISTS trg_car_audit_rules_updated ON car_audit_rules;
DROP TRIGGER IF EXISTS trg_car_reviews_updated ON car_reviews;
DROP TRIGGER IF EXISTS trg_car_reports_updated ON car_reports;
DROP TRIGGER IF EXISTS trg_car_views_updated ON car_views;
DROP TRIGGER IF EXISTS trg_car_favorites_updated ON car_favorites;
DROP TRIGGER IF EXISTS trg_car_categories_updated ON car_categories;
DROP TRIGGER IF EXISTS trg_car_transfer_updated ON car_transfer;
DROP TRIGGER IF EXISTS trg_car_insurance_updated ON car_insurance;
DROP TRIGGER IF EXISTS trg_car_financing_updated ON car_financing;
DROP TRIGGER IF EXISTS trg_car_evaluations_updated ON car_evaluations;
DROP TRIGGER IF EXISTS trg_car_contracts_updated ON car_contracts;
DROP TRIGGER IF EXISTS trg_car_test_drives_updated ON car_test_drives;
DROP TRIGGER IF EXISTS trg_car_listings_updated ON car_listings;
DROP TRIGGER IF EXISTS trg_car_inspections_updated ON car_inspections;
DROP TRIGGER IF EXISTS trg_car_models_updated ON car_models;

-- ------------------------------------------------------------
-- 2. 按反向顺序 DROP 19 张子表
--    依赖顺序：car_transfer → car_contracts / car_escrows；
--             car_test_drives → car_listings；car_listings → cars
-- ------------------------------------------------------------
DROP TABLE IF EXISTS car_recommendations;
DROP TABLE IF EXISTS car_images;
DROP TABLE IF EXISTS car_escrows;
DROP TABLE IF EXISTS car_statistics;
DROP TABLE IF EXISTS car_audit_rules;
DROP TABLE IF EXISTS car_reviews;
DROP TABLE IF EXISTS car_reports;
DROP TABLE IF EXISTS car_views;
DROP TABLE IF EXISTS car_favorites;
DROP TABLE IF EXISTS car_categories;
DROP TABLE IF EXISTS car_transfer;
DROP TABLE IF EXISTS car_insurance;
DROP TABLE IF EXISTS car_financing;
DROP TABLE IF EXISTS car_evaluations;
DROP TABLE IF EXISTS car_contracts;
DROP TABLE IF EXISTS car_test_drives;
DROP TABLE IF EXISTS car_listings;
DROP TABLE IF EXISTS car_inspections;
DROP TABLE IF EXISTS car_models;

-- ------------------------------------------------------------
-- 3. 移除 cars 主表新增字段（按反向顺序）
--    包装在 DO 块中，表不存在时跳过（幂等）
-- ------------------------------------------------------------
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'cars') THEN
        -- 真车认证
        ALTER TABLE cars DROP COLUMN IF EXISTS real_car_verified_at;
        ALTER TABLE cars DROP COLUMN IF EXISTS real_car_verified;

        -- 运营字段
        ALTER TABLE cars DROP COLUMN IF EXISTS traffic_weight;
        ALTER TABLE cars DROP COLUMN IF EXISTS promotion_level;
        ALTER TABLE cars DROP COLUMN IF EXISTS verified;
        ALTER TABLE cars DROP COLUMN IF EXISTS picked;
        ALTER TABLE cars DROP COLUMN IF EXISTS featured;

        -- 配置/特征（JSONB）
        ALTER TABLE cars DROP COLUMN IF EXISTS accident_history;
        ALTER TABLE cars DROP COLUMN IF EXISTS inspection_items;
        ALTER TABLE cars DROP COLUMN IF EXISTS tags;
        ALTER TABLE cars DROP COLUMN IF EXISTS features;

        -- 视频/360°
        ALTER TABLE cars DROP COLUMN IF EXISTS panorama_360_url;
        ALTER TABLE cars DROP COLUMN IF EXISTS video_cover;
        ALTER TABLE cars DROP COLUMN IF EXISTS video_url;

        -- 风控
        ALTER TABLE cars DROP COLUMN IF EXISTS same_car_id;
        ALTER TABLE cars DROP COLUMN IF EXISTS risk_score;
        ALTER TABLE cars DROP COLUMN IF EXISTS content_hash;

        -- 互动统计
        ALTER TABLE cars DROP COLUMN IF EXISTS last_test_drive_at;
        ALTER TABLE cars DROP COLUMN IF EXISTS test_drive_count;
        ALTER TABLE cars DROP COLUMN IF EXISTS share_count;
        ALTER TABLE cars DROP COLUMN IF EXISTS contact_count;
        ALTER TABLE cars DROP COLUMN IF EXISTS fav_count;
        ALTER TABLE cars DROP COLUMN IF EXISTS view_count;

        -- 地理位置
        ALTER TABLE cars DROP COLUMN IF EXISTS longitude;
        ALTER TABLE cars DROP COLUMN IF EXISTS latitude;
        ALTER TABLE cars DROP COLUMN IF EXISTS address;
        ALTER TABLE cars DROP COLUMN IF EXISTS business_district;
        ALTER TABLE cars DROP COLUMN IF EXISTS district;
        ALTER TABLE cars DROP COLUMN IF EXISTS city;

        -- 使用性质
        ALTER TABLE cars DROP COLUMN IF EXISTS last_maintenance_mileage;
        ALTER TABLE cars DROP COLUMN IF EXISTS maintenance_count;
        ALTER TABLE cars DROP COLUMN IF EXISTS use_type;

        -- 车架号/车牌
        ALTER TABLE cars DROP COLUMN IF EXISTS engine_no;
        ALTER TABLE cars DROP COLUMN IF EXISTS license_location;
        ALTER TABLE cars DROP COLUMN IF EXISTS license_plate;
        ALTER TABLE cars DROP COLUMN IF EXISTS vin;

        -- 过户/年检/保险
        ALTER TABLE cars DROP COLUMN IF EXISTS commercial_insurance_due;
        ALTER TABLE cars DROP COLUMN IF EXISTS insurance_status;
        ALTER TABLE cars DROP COLUMN IF EXISTS insurance_due;
        ALTER TABLE cars DROP COLUMN IF EXISTS annual_inspection_status;
        ALTER TABLE cars DROP COLUMN IF EXISTS annual_inspection_due;
        ALTER TABLE cars DROP COLUMN IF EXISTS last_transfer_date;
        ALTER TABLE cars DROP COLUMN IF EXISTS transfer_count;

        -- 车况
        ALTER TABLE cars DROP COLUMN IF EXISTS accident_count;
        ALTER TABLE cars DROP COLUMN IF EXISTS condition_score;
        ALTER TABLE cars DROP COLUMN IF EXISTS condition_level;

        -- 颜色/座位/车门
        ALTER TABLE cars DROP COLUMN IF EXISTS door_count;
        ALTER TABLE cars DROP COLUMN IF EXISTS seat_count;
        ALTER TABLE cars DROP COLUMN IF EXISTS interior_color;
        ALTER TABLE cars DROP COLUMN IF EXISTS exterior_color;

        -- 排量/变速/燃油
        ALTER TABLE cars DROP COLUMN IF EXISTS horsepower;
        ALTER TABLE cars DROP COLUMN IF EXISTS engine_type;
        ALTER TABLE cars DROP COLUMN IF EXISTS emission_standard;
        ALTER TABLE cars DROP COLUMN IF EXISTS fuel_type;
        ALTER TABLE cars DROP COLUMN IF EXISTS transmission;
        ALTER TABLE cars DROP COLUMN IF EXISTS displacement;

        -- 上牌时间/里程
        ALTER TABLE cars DROP COLUMN IF EXISTS mileage_unit;
        ALTER TABLE cars DROP COLUMN IF EXISTS mileage;
        ALTER TABLE cars DROP COLUMN IF EXISTS first_registration_date;
        ALTER TABLE cars DROP COLUMN IF EXISTS registration_month;
        ALTER TABLE cars DROP COLUMN IF EXISTS registration_year;

        -- 价格
        ALTER TABLE cars DROP COLUMN IF EXISTS dealer_price;
        ALTER TABLE cars DROP COLUMN IF EXISTS price_negotiable;
        ALTER TABLE cars DROP COLUMN IF EXISTS average_price;
        ALTER TABLE cars DROP COLUMN IF EXISTS original_price;
        ALTER TABLE cars DROP COLUMN IF EXISTS price;

        -- 品牌型号关联
        ALTER TABLE cars DROP COLUMN IF EXISTS category_id;
        ALTER TABLE cars DROP COLUMN IF EXISTS series;
        ALTER TABLE cars DROP COLUMN IF EXISTS model_name;
        ALTER TABLE cars DROP COLUMN IF EXISTS model_id;
        ALTER TABLE cars DROP COLUMN IF EXISTS brand_name;
        ALTER TABLE cars DROP COLUMN IF EXISTS brand_id;

        -- 发布类型/来源类型
        ALTER TABLE cars DROP COLUMN IF EXISTS car_type;
        ALTER TABLE cars DROP COLUMN IF EXISTS source_type;
        ALTER TABLE cars DROP COLUMN IF EXISTS listing_type;
    END IF;
END $$;
