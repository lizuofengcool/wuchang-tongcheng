-- ============================================================
-- house 房屋租售模块完整功能回滚脚本（与 007_house_full.sql 配对）
-- 按反向顺序 DROP 19 张子表 + 触发器 + 主表新增字段
-- 幂等：使用 IF EXISTS，可重复执行
-- 警告：执行后 house 模块全部扩展数据将丢失（主表基础字段保留）
-- ============================================================

-- ------------------------------------------------------------
-- 1. 先移除 19 张子表的 updated_at 触发器
-- ------------------------------------------------------------
DROP TRIGGER IF EXISTS trg_house_recommendations_updated ON house_recommendations;
DROP TRIGGER IF EXISTS trg_house_deals_updated ON house_deals;
DROP TRIGGER IF EXISTS trg_house_escrows_updated ON house_escrows;
DROP TRIGGER IF EXISTS trg_house_statistics_updated ON house_statistics;
DROP TRIGGER IF EXISTS trg_house_audit_rules_updated ON house_audit_rules;
DROP TRIGGER IF EXISTS trg_house_mortgages_updated ON house_mortgages;
DROP TRIGGER IF EXISTS trg_house_reviews_updated ON house_reviews;
DROP TRIGGER IF EXISTS trg_house_reports_updated ON house_reports;
DROP TRIGGER IF EXISTS trg_house_views_updated ON house_views;
DROP TRIGGER IF EXISTS trg_house_favorites_updated ON house_favorites;
DROP TRIGGER IF EXISTS trg_house_categories_updated ON house_categories;
DROP TRIGGER IF EXISTS trg_house_vr_tours_updated ON house_vr_tours;
DROP TRIGGER IF EXISTS trg_house_images_updated ON house_images;
DROP TRIGGER IF EXISTS trg_house_facilities_updated ON house_facilities;
DROP TRIGGER IF EXISTS trg_house_viewings_updated ON house_viewings;
DROP TRIGGER IF EXISTS trg_house_contracts_updated ON house_contracts;
DROP TRIGGER IF EXISTS trg_house_listings_updated ON house_listings;
DROP TRIGGER IF EXISTS trg_house_agents_updated ON house_agents;
DROP TRIGGER IF EXISTS trg_house_communities_updated ON house_communities;

-- ------------------------------------------------------------
-- 2. 按反向顺序 DROP 19 张子表
--    依赖顺序：house_deals → house_contracts / house_escrows；
--             house_viewings → house_listings；house_listings → houses
-- ------------------------------------------------------------
DROP TABLE IF EXISTS house_recommendations;
DROP TABLE IF EXISTS house_deals;
DROP TABLE IF EXISTS house_escrows;
DROP TABLE IF EXISTS house_statistics;
DROP TABLE IF EXISTS house_audit_rules;
DROP TABLE IF EXISTS house_mortgages;
DROP TABLE IF EXISTS house_reviews;
DROP TABLE IF EXISTS house_reports;
DROP TABLE IF EXISTS house_views;
DROP TABLE IF EXISTS house_favorites;
DROP TABLE IF EXISTS house_categories;
DROP TABLE IF EXISTS house_vr_tours;
DROP TABLE IF EXISTS house_images;
DROP TABLE IF EXISTS house_facilities;
DROP TABLE IF EXISTS house_viewings;
DROP TABLE IF EXISTS house_contracts;
DROP TABLE IF EXISTS house_listings;
DROP TABLE IF EXISTS house_agents;
DROP TABLE IF EXISTS house_communities;

-- ------------------------------------------------------------
-- 3. 移除 houses 主表新增字段（按反向顺序）
--    包装在 DO 块中，表不存在时跳过（幂等）
-- ------------------------------------------------------------
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'houses') THEN
        -- 真房源认证
        ALTER TABLE houses DROP COLUMN IF EXISTS real_house_verified_at;
        ALTER TABLE houses DROP COLUMN IF EXISTS real_house_verified;

        -- 运营字段
        ALTER TABLE houses DROP COLUMN IF EXISTS traffic_weight;
        ALTER TABLE houses DROP COLUMN IF EXISTS promotion_level;
        ALTER TABLE houses DROP COLUMN IF EXISTS verified;
        ALTER TABLE houses DROP COLUMN IF EXISTS picked;
        ALTER TABLE houses DROP COLUMN IF EXISTS featured;

        -- 配套设施/标签（JSONB）
        ALTER TABLE houses DROP COLUMN IF EXISTS nearby_pois;
        ALTER TABLE houses DROP COLUMN IF EXISTS tags;
        ALTER TABLE houses DROP COLUMN IF EXISTS facilities;

        -- 视频/VR/全景
        ALTER TABLE houses DROP COLUMN IF EXISTS panorama_url;
        ALTER TABLE houses DROP COLUMN IF EXISTS vr_url;
        ALTER TABLE houses DROP COLUMN IF EXISTS video_cover;
        ALTER TABLE houses DROP COLUMN IF EXISTS video_url;

        -- 风控
        ALTER TABLE houses DROP COLUMN IF EXISTS same_house_id;
        ALTER TABLE houses DROP COLUMN IF EXISTS risk_score;
        ALTER TABLE houses DROP COLUMN IF EXISTS content_hash;

        -- 互动统计
        ALTER TABLE houses DROP COLUMN IF EXISTS last_viewing_at;
        ALTER TABLE houses DROP COLUMN IF EXISTS viewing_count;
        ALTER TABLE houses DROP COLUMN IF EXISTS share_count;
        ALTER TABLE houses DROP COLUMN IF EXISTS contact_count;
        ALTER TABLE houses DROP COLUMN IF EXISTS fav_count;
        ALTER TABLE houses DROP COLUMN IF EXISTS view_count;

        -- 地理位置冗余
        ALTER TABLE houses DROP COLUMN IF EXISTS longitude;
        ALTER TABLE houses DROP COLUMN IF EXISTS latitude;
        ALTER TABLE houses DROP COLUMN IF EXISTS address;
        ALTER TABLE houses DROP COLUMN IF EXISTS business_district;
        ALTER TABLE houses DROP COLUMN IF EXISTS district;
        ALTER TABLE houses DROP COLUMN IF EXISTS city;

        -- 关联 ID
        ALTER TABLE houses DROP COLUMN IF EXISTS category_id;
        ALTER TABLE houses DROP COLUMN IF EXISTS agent_id;
        ALTER TABLE houses DROP COLUMN IF EXISTS community_id;

        -- 装修/产权/年限
        ALTER TABLE houses DROP COLUMN IF EXISTS building_age;
        ALTER TABLE houses DROP COLUMN IF EXISTS building_year;
        ALTER TABLE houses DROP COLUMN IF EXISTS property_years;
        ALTER TABLE houses DROP COLUMN IF EXISTS property_ownership;
        ALTER TABLE houses DROP COLUMN IF EXISTS decoration;

        -- 楼层/朝向
        ALTER TABLE houses DROP COLUMN IF EXISTS has_elevator;
        ALTER TABLE houses DROP COLUMN IF EXISTS orientation;
        ALTER TABLE houses DROP COLUMN IF EXISTS floor_type;
        ALTER TABLE houses DROP COLUMN IF EXISTS total_floor;
        ALTER TABLE houses DROP COLUMN IF EXISTS floor;

        -- 面积
        ALTER TABLE houses DROP COLUMN IF EXISTS usable_area;
        ALTER TABLE houses DROP COLUMN IF EXISTS pool_ratio;
        ALTER TABLE houses DROP COLUMN IF EXISTS inner_area;
        ALTER TABLE houses DROP COLUMN IF EXISTS building_area;

        -- 户型
        ALTER TABLE houses DROP COLUMN IF EXISTS layout;
        ALTER TABLE houses DROP COLUMN IF EXISTS balconies;
        ALTER TABLE houses DROP COLUMN IF EXISTS kitchens;
        ALTER TABLE houses DROP COLUMN IF EXISTS bathrooms;
        ALTER TABLE houses DROP COLUMN IF EXISTS halls;
        ALTER TABLE houses DROP COLUMN IF EXISTS rooms;

        -- 售价相关
        ALTER TABLE houses DROP COLUMN IF EXISTS original_price;
        ALTER TABLE houses DROP COLUMN IF EXISTS average_price;
        ALTER TABLE houses DROP COLUMN IF EXISTS sale_negotiable;
        ALTER TABLE houses DROP COLUMN IF EXISTS sale_price;

        -- 租金相关
        ALTER TABLE houses DROP COLUMN IF EXISTS rent_max_months;
        ALTER TABLE houses DROP COLUMN IF EXISTS rent_min_months;
        ALTER TABLE houses DROP COLUMN IF EXISTS rent_negotiable;
        ALTER TABLE houses DROP COLUMN IF EXISTS payment_method;
        ALTER TABLE houses DROP COLUMN IF EXISTS deposit_type;
        ALTER TABLE houses DROP COLUMN IF EXISTS rent_type;
        ALTER TABLE houses DROP COLUMN IF EXISTS rent_unit;
        ALTER TABLE houses DROP COLUMN IF EXISTS rent_price;

        -- 交易类型/发布类型
        ALTER TABLE houses DROP COLUMN IF EXISTS source_type;
        ALTER TABLE houses DROP COLUMN IF EXISTS property_type;
        ALTER TABLE houses DROP COLUMN IF EXISTS listing_type;
    END IF;
END $$;
