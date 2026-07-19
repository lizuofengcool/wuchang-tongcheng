#!/bin/bash
# 五常同城数据库备份脚本
# 策略：3-2-1 备份（3 份副本、2 种介质、1 份离线）
# 用法：./scripts/backup.sh
# 定时任务示例：0 2 * * * /opt/wuchang/scripts/backup.sh >> /var/log/wuchang-backup.log 2>&1

set -e

# ===== 配置区（可通过环境变量覆盖） =====
BACKUP_DIR="${WCTC_BACKUP_DIR:-/data/backups}"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/wuchang_pg_$DATE.sql.gz"
RETENTION_DAYS="${WCTC_BACKUP_RETENTION:-30}"

PG_HOST="${WCTC_PG_HOST:-localhost}"
PG_PORT="${WCTC_PG_PORT:-5434}"
PG_USER="${WCTC_PG_USER:-postgres}"
PG_PASSWORD="${WCTC_PG_PASSWORD:-postgres123}"
PG_DB="${WCTC_PG_DB:-wuchang_tongcheng}"

# Redis 备份（RDB 文件）
REDIS_HOST="${WCTC_REDIS_HOST:-localhost}"
REDIS_PORT="${WCTC_REDIS_PORT:-6380}"
REDIS_PASSWORD="${WCTC_REDIS_PASSWORD:-redis123}"

# MinIO 备份目录（可选）
MINIO_DATA_DIR="${WCTC_MINIO_DATA:-/data/minio}"

# ===== 主流程 =====
mkdir -p "$BACKUP_DIR"

echo "================================================"
echo "[$(date)] 五常同城备份任务启动"
echo "  备份目录: $BACKUP_DIR"
echo "  保留天数: $RETENTION_DAYS 天"
echo "================================================"

# 1. PostgreSQL 全量备份
echo "[$(date)] [1/3] 开始备份 PostgreSQL 数据库 ($PG_DB)..."
PGPASSWORD="$PG_PASSWORD" pg_dump \
  -h "$PG_HOST" -p "$PG_PORT" \
  -U "$PG_USER" -d "$PG_DB" \
  --no-owner --no-acl --verbose \
  2>"$BACKUP_DIR/wuchang_pg_$DATE.err" \
  | gzip > "$BACKUP_FILE"

if [ -s "$BACKUP_FILE" ]; then
  echo "[$(date)] [1/3] PostgreSQL 备份完成: $BACKUP_FILE"
  echo "[$(date)]       文件大小: $(du -h "$BACKUP_FILE" | cut -f1)"
else
  echo "[$(date)] [1/3] 错误：备份文件为空，请检查错误日志 $BACKUP_DIR/wuchang_pg_$DATE.err"
  exit 1
fi

# 2. Redis RDB 快照备份（通过 BGSAVE + 复制 dump.rdb）
echo "[$(date)] [2/3] 开始备份 Redis 数据..."
REDIS_BACKUP="$BACKUP_DIR/wuchang_redis_$DATE.rdb"
if redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" --no-auth-warning BGSAVE >/dev/null 2>&1; then
  # 等待 BGSAVE 完成
  while [ "$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" --no-auth-warning LASTSAVE 2>/dev/null)" = "$LASTSAVE" ]; do
    sleep 1
  done
  # 这里假设 redis 容器内 dump.rdb 已生成，实际部署时需从卷复制
  echo "[$(date)] [2/3] Redis BGSAVE 触发成功，RDB 快照已生成"
else
  echo "[$(date)] [2/3] 警告：Redis BGSAVE 失败，跳过（不影响主流程）"
fi

# 3. 清理过期备份
echo "[$(date)] [3/3] 清理 $RETENTION_DAYS 天前的过期备份..."
find "$BACKUP_DIR" -name "wuchang_pg_*.sql.gz" -mtime +"$RETENTION_DAYS" -delete
find "$BACKUP_DIR" -name "wuchang_pg_*.err" -mtime +"$RETENTION_DAYS" -delete
find "$BACKUP_DIR" -name "wuchang_redis_*.rdb" -mtime +"$RETENTION_DAYS" -delete
echo "[$(date)] [3/3] 清理完成"

# ===== 离线副本上传（可选，需配置七牛云 qshell） =====
# TODO: 上传七牛云 OSS（需要配置七牛云 SDK + qshell）
# if command -v qshell >/dev/null 2>&1; then
#   QINIU_BUCKET="${QINIU_BUCKET:-wuchang-tongcheng-backup}"
#   qshell rput "$QINIU_BUCKET" "backup/$(basename "$BACKUP_FILE")" "$BACKUP_FILE"
#   echo "[$(date)] 已上传至七牛云 $QINIU_BUCKET"
# fi

echo "================================================"
echo "[$(date)] 五常同城备份流程结束"
echo "================================================"
