-- =====================================================
-- lbs LBS地图中台模块完整迁移脚本
-- 包含 3 张表 + 索引 + 触发器
-- 依据 v3.2.1 架构方案第 4.8 节：高德定位/附近检索/距离排序/POI/路线规划/地理围栏/分站区域隔离
-- 依据 docs/架构设计/PostGIS地理字段规范.md：未启用 PostGIS 时使用 DECIMAL(10,7) 降级方案
-- 数据库：wuchang_tongcheng
-- =====================================================
--
-- 内容：
--   1. CREATE 3 张表（lbs_ 前缀；lbs_pois 主表由 GORM AutoMigrate 创建）
--   2. 索引、触发器、注释
--   3. 全幂等：CREATE TABLE IF NOT EXISTS
-- 依赖：001_p0_baseline.sql（update_updated_at_column 函数已创建）
--
-- 表清单（3 张表，主表 lbs_pois 由 GORM AutoMigrate 创建）：
--   lbs_pois / lbs_regions / lbs_geofences
--
-- 说明：
--   1. 经纬度统一使用 DECIMAL(10,7)（精度 1cm），不依赖 PostGIS 扩展
--   2. 附近查询使用 Haversine 公式（6371km 地球半径）+ BoundingBox 索引优化
--   3. 多边形数据存储在 JSONB 字段中（boundary/points），由应用层算法判断
-- =====================================================

-- ============================================================
-- 1. lbs_pois POI 兴趣点表
-- ============================================================
CREATE TABLE IF NOT EXISTS lbs_pois (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    -- 基础信息
    name VARCHAR(200) NOT NULL DEFAULT '',
    address VARCHAR(500) NOT NULL DEFAULT '',
    category VARCHAR(64) NOT NULL DEFAULT '',
    phone VARCHAR(32) NOT NULL DEFAULT '',
    icon VARCHAR(255) NOT NULL DEFAULT '',
    status INT NOT NULL DEFAULT 1,                -- 0下架 1上线 2待审 3拒绝 4删除

    -- 地理位置（DECIMAL 降级方案，精度 1cm）
    latitude DECIMAL(10,7) NOT NULL DEFAULT 0,
    longitude DECIMAL(10,7) NOT NULL DEFAULT 0,

    -- 扩展信息
    user_id BIGINT NOT NULL DEFAULT 0,
    source VARCHAR(32) NOT NULL DEFAULT 'manual',  -- manual/amap/import
    external_id VARCHAR(64) NOT NULL DEFAULT '',
    tags JSONB,
    extra JSONB,
    published_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_lbs_pois_region_id ON lbs_pois(region_id);
CREATE INDEX IF NOT EXISTS idx_lbs_pois_user_id ON lbs_pois(user_id);
CREATE INDEX IF NOT EXISTS idx_lbs_pois_name ON lbs_pois(name);
CREATE INDEX IF NOT EXISTS idx_lbs_pois_category ON lbs_pois(category);
CREATE INDEX IF NOT EXISTS idx_lbs_pois_status ON lbs_pois(status);
CREATE INDEX IF NOT EXISTS idx_lbs_pois_source ON lbs_pois(source);
CREATE INDEX IF NOT EXISTS idx_lbs_pois_external_id ON lbs_pois(external_id);
CREATE INDEX IF NOT EXISTS idx_lbs_pois_published_at ON lbs_pois(published_at);
CREATE INDEX IF NOT EXISTS idx_lbs_pois_deleted_at ON lbs_pois(deleted_at);
-- 复合索引：附近检索优化（按 region_id + status + 经纬度范围）
CREATE INDEX IF NOT EXISTS idx_lbs_pois_lat_lng ON lbs_pois(latitude, longitude);
CREATE INDEX IF NOT EXISTS idx_lbs_pois_region_status_lat_lng ON lbs_pois(region_id, status, latitude, longitude);
-- JSONB 索引（可选，用于按标签查询）
CREATE INDEX IF NOT EXISTS idx_lbs_pois_tags_gin ON lbs_pois USING GIN(tags) WHERE tags IS NOT NULL;

COMMENT ON TABLE lbs_pois IS 'LBS POI 兴趣点表（地图标注点：商户/地标/兴趣点）';
COMMENT ON COLUMN lbs_pois.status IS '状态：0下架 1上线 2待审 3拒绝 4删除';
COMMENT ON COLUMN lbs_pois.source IS '来源：manual手动/amap高德/import导入';
COMMENT ON COLUMN lbs_pois.latitude IS '纬度 DECIMAL(10,7) 精度 1cm';
COMMENT ON COLUMN lbs_pois.longitude IS '经度 DECIMAL(10,7) 精度 1cm';
COMMENT ON COLUMN lbs_pois.tags IS '标签 JSONB 数组';
COMMENT ON COLUMN lbs_pois.extra IS '扩展字段 JSONB';
COMMENT ON COLUMN lbs_pois.external_id IS '外部 ID（如高德 POI ID）';

-- ============================================================
-- 2. lbs_regions 区域分站表
-- ============================================================
CREATE TABLE IF NOT EXISTS lbs_regions (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    -- 基础信息
    name VARCHAR(100) NOT NULL DEFAULT '',
    city_code VARCHAR(20) NOT NULL DEFAULT '',
    parent_id BIGINT NOT NULL DEFAULT 0,
    level INT NOT NULL DEFAULT 1,                -- 1省 2市 3区 4乡镇
    path VARCHAR(500) NOT NULL DEFAULT '',       -- 路径如 1,5,12
    sort INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 1,               -- 0禁用 1启用

    -- 中心点（DECIMAL 降级方案）
    center_lat DECIMAL(10,7) NOT NULL DEFAULT 0,
    center_lng DECIMAL(10,7) NOT NULL DEFAULT 0,

    -- 边界（多边形顶点 JSONB：[{lat,lng},...]）
    boundary JSONB,

    -- 扩展信息
    ad_code VARCHAR(20) NOT NULL DEFAULT '',     -- 行政区划代码
    zip_code VARCHAR(20) NOT NULL DEFAULT '',     -- 邮编
    description VARCHAR(500) NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_lbs_regions_name ON lbs_regions(name);
CREATE INDEX IF NOT EXISTS idx_lbs_regions_city_code ON lbs_regions(city_code);
CREATE INDEX IF NOT EXISTS idx_lbs_regions_parent_id ON lbs_regions(parent_id);
CREATE INDEX IF NOT EXISTS idx_lbs_regions_level ON lbs_regions(level);
CREATE INDEX IF NOT EXISTS idx_lbs_regions_status ON lbs_regions(status);
CREATE INDEX IF NOT EXISTS idx_lbs_regions_ad_code ON lbs_regions(ad_code);
CREATE INDEX IF NOT EXISTS idx_lbs_regions_center_lat_lng ON lbs_regions(center_lat, center_lng);
CREATE INDEX IF NOT EXISTS idx_lbs_regions_deleted_at ON lbs_regions(deleted_at);
CREATE INDEX IF NOT EXISTS idx_lbs_regions_boundary_gin ON lbs_regions USING GIN(boundary) WHERE boundary IS NOT NULL;

COMMENT ON TABLE lbs_regions IS 'LBS 区域分站表（行政区域/分站隔离/配送范围）';
COMMENT ON COLUMN lbs_regions.level IS '层级：1省 2市 3区 4乡镇';
COMMENT ON COLUMN lbs_regions.status IS '状态：0禁用 1启用';
COMMENT ON COLUMN lbs_regions.center_lat IS '中心纬度 DECIMAL(10,7)';
COMMENT ON COLUMN lbs_regions.center_lng IS '中心经度 DECIMAL(10,7)';
COMMENT ON COLUMN lbs_regions.boundary IS '边界多边形 JSONB [{lat,lng},...]';
COMMENT ON COLUMN lbs_regions.ad_code IS '行政区划代码';

-- ============================================================
-- 3. lbs_geofences 地理围栏表
-- ============================================================
CREATE TABLE IF NOT EXISTS lbs_geofences (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    -- 基础信息
    name VARCHAR(100) NOT NULL DEFAULT '',
    type VARCHAR(16) NOT NULL DEFAULT 'circle',  -- circle/polygon
    status INT NOT NULL DEFAULT 1,                -- 0禁用 1启用
    sort INT NOT NULL DEFAULT 0,
    description VARCHAR(500) NOT NULL DEFAULT '',

    -- 圆形围栏参数
    center_lat DECIMAL(10,7) NOT NULL DEFAULT 0,
    center_lng DECIMAL(10,7) NOT NULL DEFAULT 0,
    radius DECIMAL(10,2) NOT NULL DEFAULT 0,      -- 半径（米）

    -- 多边形围栏参数（顶点 JSONB：[{lat,lng},...]）
    points JSONB,

    -- 扩展信息
    owner_id BIGINT NOT NULL DEFAULT 0,
    owner_type VARCHAR(32) NOT NULL DEFAULT '',   -- shop/agent/daojia
    extra JSONB
);
CREATE INDEX IF NOT EXISTS idx_lbs_geofences_region_id ON lbs_geofences(region_id);
CREATE INDEX IF NOT EXISTS idx_lbs_geofences_name ON lbs_geofences(name);
CREATE INDEX IF NOT EXISTS idx_lbs_geofences_type ON lbs_geofences(type);
CREATE INDEX IF NOT EXISTS idx_lbs_geofences_status ON lbs_geofences(status);
CREATE INDEX IF NOT EXISTS idx_lbs_geofences_owner_id ON lbs_geofences(owner_id);
CREATE INDEX IF NOT EXISTS idx_lbs_geofences_owner_type ON lbs_geofences(owner_type);
CREATE INDEX IF NOT EXISTS idx_lbs_geofences_center_lat_lng ON lbs_geofences(center_lat, center_lng);
CREATE INDEX IF NOT EXISTS idx_lbs_geofences_deleted_at ON lbs_geofences(deleted_at);
CREATE INDEX IF NOT EXISTS idx_lbs_geofences_points_gin ON lbs_geofences USING GIN(points) WHERE points IS NOT NULL;

COMMENT ON TABLE lbs_geofences IS 'LBS 地理围栏表（配送范围/电子围栏/考勤打卡区域）';
COMMENT ON COLUMN lbs_geofences.type IS '围栏类型：circle圆形/polygon多边形';
COMMENT ON COLUMN lbs_geofences.status IS '状态：0禁用 1启用';
COMMENT ON COLUMN lbs_geofences.center_lat IS '圆心纬度 DECIMAL(10,7)';
COMMENT ON COLUMN lbs_geofences.center_lng IS '圆心经度 DECIMAL(10,7)';
COMMENT ON COLUMN lbs_geofences.radius IS '圆形围栏半径（米）';
COMMENT ON COLUMN lbs_geofences.points IS '多边形顶点 JSONB [{lat,lng},...]';
COMMENT ON COLUMN lbs_geofences.owner_type IS '所有者类型：shop/agent/daojia';

-- ============================================================
-- updated_at 触发器（依赖 001_p0_baseline.sql 中的 update_updated_at_column 函数）
-- PostgreSQL 不支持 CREATE TRIGGER IF NOT EXISTS，用 DROP TRIGGER IF EXISTS + CREATE TRIGGER
-- ============================================================
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'update_updated_at_column') THEN
        DROP TRIGGER IF EXISTS trg_lbs_pois_updated_at ON lbs_pois;
        CREATE TRIGGER trg_lbs_pois_updated_at BEFORE UPDATE ON lbs_pois FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

        DROP TRIGGER IF EXISTS trg_lbs_regions_updated_at ON lbs_regions;
        CREATE TRIGGER trg_lbs_regions_updated_at BEFORE UPDATE ON lbs_regions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

        DROP TRIGGER IF EXISTS trg_lbs_geofences_updated_at ON lbs_geofences;
        CREATE TRIGGER trg_lbs_geofences_updated_at BEFORE UPDATE ON lbs_geofences FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;
