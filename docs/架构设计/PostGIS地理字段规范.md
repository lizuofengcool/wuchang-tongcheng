# PostGIS 地理字段规范

> 版本：v1.0.0
> 适用范围：五常同城全平台所有涉及地理位置的业务表
> 依据：`docs/架构设计/ershou-大模块架构方案.md`（v3.2.1）第 6.4 节 LBS 中台、第 11.5 节 监控指标
> 维护：Agent 2 数据库
> 创建时间：2026-07-19

---

## 一、设计目标

1. **统一地理字段类型**：所有点 / 多边形 / 线段地理数据使用 PostGIS `GEOGRAPHY` 类型，避免坐标系歧义
2. **统一坐标系**：全平台使用 WGS84（SRID=4326），与高德 / 百度地图 API 对齐
3. **统一索引策略**：所有 `GEOGRAPHY` 字段必须建 GIST 索引，保证空间查询性能
4. **降级兼容**：未启用 PostGIS 的环境使用 `DECIMAL(10,7)` 经纬度字段，业务可降级运行

---

## 二、扩展启用

### 2.1 启用 PostGIS 扩展

```sql
-- 在目标数据库中执行（docker-compose 已通过 initdb/01-extensions.sql 自动启用）
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS postgis_topology;

-- 验证安装
SELECT postgis_full_version();
```

### 2.2 镜像要求

`deploy/docker-compose.yml` 使用 `postgis/postgis:16-3.4` 镜像，已内置 PostGIS 3.4，无需额外安装。

### 2.3 客户端连接要求

- GORM `gorm.io/driver/postgres` v1.6.0+：原生支持 PostGIS 字段读写
- 应用层读取 `GEOGRAPHY` 字段时使用 `ST_AsText()` 或 `ST_AsGeoJSON()` 转换为文本

---

## 三、字段类型规范

### 3.1 点位字段（Point）

适用场景：POI、用户位置、商家位置、二手商品发布位置

```sql
-- 字段定义
location GEOGRAPHY(Point, 4326) NOT NULL

-- GIST 索引（强制）
CREATE INDEX idx_{表名}_location ON {表名} USING GIST(location);
```

**典型表**：
- `lbs_pois.location` - POI 兴趣点位置
- `ers_listings.location` - 二手商品发布位置
- `merchant_shops.location` - 商家店铺位置
- `user_profiles.last_location` - 用户最后已知位置（可选）

### 3.2 多边形字段（Polygon）

适用场景：行政区域边界、商圈范围、配送范围

```sql
-- 字段定义
boundary GEOGRAPHY(Polygon, 4326) NOT NULL

-- GIST 索引（强制）
CREATE INDEX idx_{表名}_boundary ON {表名} USING GIST(boundary);
```

**典型表**：
- `lbs_regions.boundary` - 行政区域边界（省/市/县/乡镇）
- `distribution_stations.coverage` - 代理商覆盖范围
- `daojia_services.coverage_area` - 上门服务覆盖范围

### 3.3 经纬度数字字段（DECIMAL，降级方案）

适用场景：未启用 PostGIS 的环境、轻量级地理查询、行为日志表

```sql
-- 字段定义（精度 10,7：小数点前 3 位 + 小数点后 7 位，可精确到 1cm）
latitude  DECIMAL(10,7) NOT NULL DEFAULT 0,
longitude DECIMAL(10,7) NOT NULL DEFAULT 0

-- 普通索引（如需按经纬度范围查询）
CREATE INDEX idx_{表名}_lat_lng ON {表名}(latitude, longitude);
```

**典型表**：
- `ers_user_behaviors.latitude / longitude` - 用户行为日志（高频写入，降级用 DECIMAL）
- `module_metrics.labels.latitude / longitude` - 监控指标中的位置标签（JSONB 内）

---

## 四、查询规范

### 4.1 附近查询（ST_DWithin + ST_Distance）

`ST_DWithin` 比 `ST_Distance + ORDER BY` 性能更好，因为它可以利用 GIST 索引进行预过滤。

```sql
-- 附近 5km 的二手商品（location 字段为 GEOGRAPHY(Point, 4326)）
SELECT
    id,
    title,
    ST_AsText(location) AS location_wkt,
    ST_Distance(location, ST_MakePoint(126.6592, 44.9320)::geography) AS distance_meters
FROM ers_listings
WHERE deleted_at IS NULL
  AND status = 'published'
  AND ST_DWithin(location, ST_MakePoint(126.6592, 44.9320)::geography, 5000)  -- 5km = 5000m
ORDER BY distance_meters ASC
LIMIT 20 OFFSET 0;
```

**要点**：
1. `ST_MakePoint(lng, lat)` 注意参数顺序：**经度在前，纬度在后**
2. `::geography` 强制转换为 GEOGRAPHY 类型，距离单位为**米**
3. `ST_DWithin` 第三参数单位为**米**（GEOGRAPHY 模式下）
4. 配合 GIST 索引，可避免全表扫描

### 4.2 按距离排序（ORDER BY ST_Distance）

```sql
-- 按距离排序的附近商家（带分页）
SELECT
    s.id,
    s.name,
    s.address,
    ST_Distance(s.location, ST_MakePoint(126.6592, 44.9320)::geography) AS distance_meters
FROM merchant_shops s
WHERE s.deleted_at IS NULL
  AND s.status = 1
  AND ST_DWithin(s.location, ST_MakePoint(126.6592, 44.9320)::geography, 10000)  -- 先用 DWithin 过滤 10km 范围
ORDER BY distance_meters ASC
LIMIT 20 OFFSET 0;
```

### 4.3 区域包含查询（ST_Contains / ST_Within）

```sql
-- 查询某行政区域内的所有 POI
SELECT p.id, p.name, p.address
FROM lbs_pois p
JOIN lbs_regions r ON r.code = '230184'  -- 五常市
WHERE ST_Within(p.location::geometry, r.boundary::geometry);
-- 注意：ST_Within 需要 geometry 类型，需从 geography 转换

-- 查询某点所在的行政区域
SELECT id, name, code, level
FROM lbs_regions
WHERE ST_Contains(boundary::geometry, ST_MakePoint(126.6592, 44.9320)::geometry);
```

### 4.4 经纬度数字字段的降级查询

当表使用 `DECIMAL(10,7)` 经纬度字段时（如 `ers_user_behaviors`），无法使用 PostGIS 空间索引，采用矩形范围近似查询：

```sql
-- 降级方案：矩形范围查询（近似 5km 范围）
-- 1 度纬度 ≈ 111km，5km ≈ 0.045 度
-- 1 度经度 ≈ 111km × cos(lat)，五常市纬度 44.93°，cos(44.93°) ≈ 0.708
-- 5km 经度范围 ≈ 0.045 / 0.708 ≈ 0.064 度
SELECT id, user_id, latitude, longitude,
       SQRT(POWER(latitude - 44.9320, 2) + POWER(longitude - 126.6592, 2)) AS rough_distance
FROM ers_user_behaviors
WHERE latitude  BETWEEN 44.9320 - 0.045 AND 44.9320 + 0.045
  AND longitude BETWEEN 126.6592 - 0.064 AND 126.6592 + 0.064
ORDER BY rough_distance ASC
LIMIT 20;
```

> **注意**：降级方案为矩形过滤，结果含误差，仅用于行为日志等精度要求不高的场景。POI / 商家等核心场景必须使用 PostGIS `GEOGRAPHY` 字段。

---

## 五、索引规范

### 5.1 GIST 索引（强制）

所有 `GEOGRAPHY` / `GEOMETRY` 字段必须建 GIST 索引：

```sql
-- 点位字段
CREATE INDEX IF NOT EXISTS idx_lbs_pois_location ON lbs_pois USING GIST(location);
CREATE INDEX IF NOT EXISTS idx_ers_listings_location ON ers_listings USING GIST(location);
CREATE INDEX IF NOT EXISTS idx_merchant_shops_location ON merchant_shops USING GIST(location);

-- 多边形字段
CREATE INDEX IF NOT EXISTS idx_lbs_regions_boundary ON lbs_regions USING GIST(boundary);
```

### 5.2 索引性能对比

| 查询方式 | 无索引 | B-Tree 索引 | GIST 索引 |
|---------|-------|------------|----------|
| `ST_DWithin` 全表扫描 | 慢（O(N)） | 不支持 | 快（O(log N)） |
| `ST_Distance + ORDER BY` | 慢（O(N log N)） | 不支持 | 快（O(log N)） |
| `ST_Contains` | 慢（O(N)） | 不支持 | 快（O(log N)） |

### 5.3 索引维护

```sql
-- 重建 GIST 索引（数据量大时定期执行）
REINDEX INDEX idx_lbs_pois_location;

-- 分析表统计信息（重要查询前执行）
ANALYZE lbs_pois;
ANALYZE lbs_regions;
```

---

## 六、数据写入规范

### 6.1 写入 GEOGRAPHY 字段

```sql
-- 方式 1：ST_MakePoint + ::geography 转换
INSERT INTO lbs_pois (name, location) VALUES
('五常市政府', ST_MakePoint(126.6592, 44.9320)::geography);

-- 方式 2：ST_GeomFromText + ::geography 转换
INSERT INTO lbs_pois (name, location) VALUES
('五常市政府', ST_GeomFromText('POINT(126.6592 44.9320)', 4326)::geography);

-- 方式 3：GORM 应用层写入（推荐）
-- Go: poi.Location = geo.Point{Lng: 126.6592, Lat: 44.9320}
-- GORM 自动转换为 WKT 写入
```

### 6.2 写入多边形字段

```sql
-- 行政区域边界（五常市简化边界）
INSERT INTO lbs_regions (name, code, level, boundary) VALUES
('五常市', '230184', 3,
  ST_GeomFromText('POLYGON((126.5 44.8, 126.8 44.8, 126.8 45.1, 126.5 45.1, 126.5 44.8))', 4326)::geography
);
```

### 6.3 读取 GEOGRAPHY 字段

```sql
-- 方式 1：WKT 文本（推荐用于 API 返回）
SELECT id, name, ST_AsText(location) AS location_wkt FROM lbs_pois WHERE id = 1;
-- 返回：'POINT(126.6592 44.9320)'

-- 方式 2：GeoJSON（推荐用于前端地图渲染）
SELECT id, name, ST_AsGeoJSON(location) AS location_geojson FROM lbs_pois WHERE id = 1;
-- 返回：'{"type":"Point","coordinates":[126.6592,44.9320]}'

-- 方式 3：经纬度分离（推荐用于业务逻辑处理）
SELECT id, name,
       ST_Y(location::geometry) AS latitude,
       ST_X(location::geometry) AS longitude
FROM lbs_pois WHERE id = 1;
```

---

## 七、GORM 集成规范

### 7.1 Model 定义

```go
import "gorm.io/gorm/schema"

type LbsPoi struct {
    ID          int64          `gorm:"primarykey" json:"id"`
    Name        string         `gorm:"size:128;not null" json:"name"`
    Location    string         `gorm:"type:geography(Point,4326);not null" json:"location"` // WKT 文本
    // ... 其他字段
}

// 在 AutoMigrate 后手动创建 GIST 索引（GORM 不支持 GIST 索引自动创建）
// DB.Exec("CREATE INDEX IF NOT EXISTS idx_lbs_pois_location ON lbs_pois USING GIST(location)")
```

### 7.2 查询示例

```go
// 附近 5km 的 POI
var pois []LbsPoi
db.Raw(`
    SELECT id, name,
           ST_AsText(location) AS location,
           ST_Distance(location, ST_MakePoint(?, ?)::geography) AS distance
    FROM lbs_pois
    WHERE ST_DWithin(location, ST_MakePoint(?, ?)::geography, ?)
    ORDER BY distance
    LIMIT ?
`, lng, lat, lng, lat, radiusMeters, limit).Scan(&pois)
```

---

## 八、完整示例：附近 5km 二手商品查询

### 8.1 表结构

```sql
-- ers_listings 二手商品表（地理位置核心表）
CREATE TABLE IF NOT EXISTS ers_listings (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(128) NOT NULL,
    category_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    price DECIMAL(12,2) NOT NULL DEFAULT 0,
    description TEXT NOT NULL DEFAULT '',
    location GEOGRAPHY(Point, 4326),           -- PostGIS 字段
    latitude DECIMAL(10,7) NOT NULL DEFAULT 0,  -- 降级字段（与 location 二选一）
    longitude DECIMAL(10,7) NOT NULL DEFAULT 0, -- 降级字段
    status VARCHAR(32) NOT NULL DEFAULT 'draft',
    region_id BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_ers_listings_location ON ers_listings USING GIST(location);
CREATE INDEX IF NOT EXISTS idx_ers_listings_status ON ers_listings(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ers_listings_user ON ers_listings(user_id);
CREATE INDEX IF NOT EXISTS idx_ers_listings_region ON ers_listings(region_id);
```

### 8.2 完整查询 SQL

```sql
-- 五常市中心 (126.6592, 44.9320) 附近 5km 的二手商品，按距离排序，分页 20 条
SELECT
    l.id,
    l.title,
    l.price,
    l.description,
    ST_AsText(l.location) AS location_wkt,
    ST_X(l.location::geometry) AS longitude,
    ST_Y(l.location::geometry) AS latitude,
    ROUND(ST_Distance(l.location, ST_MakePoint(126.6592, 44.9320)::geography)::numeric, 2) AS distance_meters,
    u.nickname AS seller_nickname,
    c.name AS category_name
FROM ers_listings l
LEFT JOIN user_profiles u ON u.id = l.user_id
LEFT JOIN info_categories c ON c.id = l.category_id
WHERE l.deleted_at IS NULL
  AND l.status = 'published'
  AND ST_DWithin(l.location, ST_MakePoint(126.6592, 44.9320)::geography, 5000)
ORDER BY distance_meters ASC
LIMIT 20 OFFSET 0;
```

### 8.3 性能要点

1. `ST_DWithin` 利用 GIST 索引预过滤，避免计算所有行的距离
2. `ST_Distance` 仅在过滤后的少量行上计算
3. `ORDER BY distance_meters` 配合 LIMIT，PostgreSQL 会选择最优执行计划
4. 大数据量场景（百万级）建议增加 `region_id` 过滤先缩小范围

---

## 九、版本历史

| 版本 | 日期 | 变更说明 | 维护人 |
|------|------|---------|--------|
| v1.0.0 | 2026-07-19 | 首版发布：GEOGRAPHY 字段规范、GIST 索引、查询规范、降级方案、完整示例 | Agent 2 |
