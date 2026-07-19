#!/bin/bash
# 五常同城数据库恢复脚本
# 用法：
#   恢复最新备份：./scripts/restore.sh
#   恢复指定备份：./scripts/restore.sh /data/backups/wuchang_pg_20260719_020000.sql.gz
# 策略：3-2-1 备份的恢复演练，定期验证备份可用性

set -e

# ===== 配置区（可通过环境变量覆盖） =====
BACKUP_DIR="${WCTC_BACKUP_DIR:-/data/backups}"

PG_HOST="${WCTC_PG_HOST:-localhost}"
PG_PORT="${WCTC_PG_PORT:-5434}"
PG_USER="${WCTC_PG_USER:-postgres}"
PG_PASSWORD="${WCTC_PG_PASSWORD:-postgres123}"
PG_DB="${WCTC_PG_DB:-wuchang_tongcheng}"

# ===== 参数处理 =====
BACKUP_FILE="$1"

if [ -z "$BACKUP_FILE" ]; then
  # 未指定文件，恢复最新备份
  BACKUP_FILE=$(ls -t "$BACKUP_DIR"/wuchang_pg_*.sql.gz 2>/dev/null | head -n 1)
  if [ -z "$BACKUP_FILE" ]; then
    echo "错误：在 $BACKUP_DIR 下未找到备份文件"
    echo "用法：$0 [备份文件路径]"
    exit 1
  fi
fi

if [ ! -f "$BACKUP_FILE" ]; then
  echo "错误：备份文件不存在: $BACKUP_FILE"
  exit 1
fi

echo "================================================"
echo "[$(date)] 五常同城恢复任务启动"
echo "  备份文件: $BACKUP_FILE"
echo "  文件大小: $(du -h "$BACKUP_FILE" | cut -f1)"
echo "  目标数据库: $PG_DB @ $PG_HOST:$PG_PORT"
echo "================================================"
echo ""
echo "⚠️  警告：此操作将覆盖目标数据库的现有数据！"
echo "    建议：先在测试环境验证，再恢复到生产环境。"
echo ""
read -r -p "确认恢复？输入 yes 继续: " CONFIRM
if [ "$CONFIRM" != "yes" ]; then
  echo "已取消恢复。"
  exit 0
fi

# ===== 恢复流程 =====
echo "[$(date)] [1/2] 开始恢复 PostgreSQL 数据库..."

# 解压并导入
gunzip -c "$BACKUP_FILE" \
  | PGPASSWORD="$PG_PASSWORD" psql \
    -h "$PG_HOST" -p "$PG_PORT" \
    -U "$PG_USER" -d "$PG_DB" \
    --set ON_ERROR_STOP=on \
    -v ON_ERROR_STOP=1

echo "[$(date)] [1/2] PostgreSQL 恢复完成"

# ===== 验证 =====
echo "[$(date)] [2/2] 验证恢复结果..."
TABLE_COUNT=$(PGPASSWORD="$PG_PASSWORD" psql -h "$PG_HOST" -p "$PG_PORT" -U "$PG_USER" -d "$PG_DB" -t -c "SELECT count(*) FROM information_schema.tables WHERE table_schema='public';")
echo "[$(date)] [2/2] public schema 表数量: $TABLE_COUNT"

echo "================================================"
echo "[$(date)] 五常同城恢复流程结束"
echo "  建议执行："
echo "    1. 检查关键表数据是否完整"
echo "    2. 启动后端服务验证业务功能"
echo "    3. 抽样验证索引、约束、PostGIS 扩展"
echo "================================================"
