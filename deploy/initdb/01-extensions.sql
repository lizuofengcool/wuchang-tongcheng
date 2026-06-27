-- 五常同城数据库初始化脚本
-- 由 docker-compose 挂载至 /docker-entrypoint-initdb.d/，仅在首次创建数据库时执行
-- 注意：业务表结构由后端 GORM AutoMigrate 自动创建，种子数据由 seed.Run() 写入，
--       此脚本只负责 AutoMigrate 无法完成的扩展安装。

-- PostGIS 空间扩展（docker-compose 使用 postgis/postgis 镜像，需显式启用）
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS postgis_topology;

-- 验证安装
-- SELECT postgis_full_version();

-- 未来空间查询示例（需在业务表增加 geom geometry 列后使用）：
-- 1. 附近商家（按经纬度半径查询）：
--    SELECT id, name, ST_Distance(geom::geography, ST_MakePoint(lon, lat)::geography) AS dist
--    FROM merchants WHERE ST_DWithin(geom::geography, ST_MakePoint(lon, lat)::geography, 5000)
--    ORDER BY dist LIMIT 20;
-- 2. 区域多边形包含判断：
--    SELECT * FROM regions WHERE ST_Contains(boundary, ST_MakePoint(lon, lat));
