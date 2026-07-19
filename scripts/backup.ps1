# 五常同城数据库备份脚本（Windows PowerShell 版）
# 策略：3-2-1 备份（3 份副本、2 种介质、1 份离线）
# 用法：
#   .\scripts\backup.ps1
#   .\scripts\backup.ps1 -BackupDir D:\backups
# 定时任务示例（任务计划程序）：每天 02:00 执行
#   powershell.exe -ExecutionPolicy Bypass -File D:\kaifa\wuchang-tongcheng\scripts\backup.ps1

param(
    [string]$BackupDir = $env:WCTC_BACKUP_DIR,
    [int]$RetentionDays = 30,
    [string]$PgHost = "localhost",
    [int]$PgPort = 5434,
    [string]$PgUser = "postgres",
    [string]$PgPassword = "postgres123",
    [string]$PgDb = "wuchang_tongcheng"
)

$ErrorActionPreference = "Stop"

# ===== 默认备份目录（项目同级目录） =====
if (-not $BackupDir) {
    $ProjectRoot = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
    $BackupDir = Join-Path $ProjectRoot "backups"
}

# ===== 主流程 =====
if (-not (Test-Path $BackupDir)) {
    New-Item -ItemType Directory -Path $BackupDir -Force | Out-Null
}

$DateStr = Get-Date -Format "yyyyMMdd_HHmmss"
$BackupFile = Join-Path $BackupDir "wuchang_pg_$DateStr.sql.gz"

Write-Host "================================================" -ForegroundColor Cyan
Write-Host "[$(Get-Date)] 五常同城备份任务启动" -ForegroundColor Cyan
Write-Host "  备份目录: $BackupDir"
Write-Host "  保留天数: $RetentionDays 天"
Write-Host "================================================" -ForegroundColor Cyan

# 1. 检查 pg_dump 是否可用
$pgDump = Get-Command pg_dump -ErrorAction SilentlyContinue
if (-not $pgDump) {
    # 尝试常见安装路径
    $pgCandidates = @(
        "C:\Program Files\PostgreSQL\16\bin\pg_dump.exe",
        "C:\Program Files\PostgreSQL\15\bin\pg_dump.exe",
        "C:\Program Files\PostgreSQL\14\bin\pg_dump.exe"
    )
    foreach ($cand in $pgCandidates) {
        if (Test-Path $cand) {
            $pgDumpPath = $cand
            break
        }
    }
    if (-not $pgDumpPath) {
        Write-Host "[$(Get-Date)] 错误：未找到 pg_dump，请安装 PostgreSQL 客户端工具或将其加入 PATH" -ForegroundColor Red
        exit 1
    }
} else {
    $pgDumpPath = $pgDump.Source
}

# 2. PostgreSQL 全量备份
Write-Host "[$(Get-Date)] [1/3] 开始备份 PostgreSQL 数据库 ($PgDb)..." -ForegroundColor Yellow

$env:PGPASSWORD = $PgPassword
$dumpArgs = @(
    "-h", $PgHost,
    "-p", $PgPort,
    "-U", $PgUser,
    "-d", $PgDb,
    "--no-owner",
    "--no-acl",
    "--verbose"
)

# 使用 .NET GZipStream 直接压缩（避免依赖 gzip 命令）
$ErrorFile = Join-Path $BackupDir "wuchang_pg_$DateStr.err"
$tempSql = Join-Path $env:TEMP "wuchang_pg_$DateStr.sql"

try {
    # 调用 pg_dump 输出到临时文件
    & $pgDumpPath @dumpArgs 2>$ErrorFile | Out-File -FilePath $tempSql -Encoding utf8

    # 使用 GZip 压缩
    Add-Type -AssemblyName System.IO.Compression.FileSystem
    $sourceBytes = [System.IO.File]::ReadAllBytes($tempSql)
    $destStream = [System.IO.File]::Create($BackupFile)
    $gzipStream = New-Object System.IO.Compression.GZipStream($destStream, [System.IO.Compression.CompressionMode]::Compress)
    $gzipStream.Write($sourceBytes, 0, $sourceBytes.Length)
    $gzipStream.Close()
    $destStream.Close()

    # 验证备份文件
    $fileSize = (Get-Item $BackupFile).Length
    if ($fileSize -gt 0) {
        $sizeStr = if ($fileSize -gt 1MB) { "{0:N2} MB" -f ($fileSize / 1MB) } else { "{0:N2} KB" -f ($fileSize / 1KB) }
        Write-Host "[$(Get-Date)] [1/3] PostgreSQL 备份完成: $BackupFile" -ForegroundColor Green
        Write-Host "[$(Get-Date)]       文件大小: $sizeStr"
    } else {
        Write-Host "[$(Get-Date)] [1/3] 错误：备份文件为空，请检查错误日志 $ErrorFile" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "[$(Get-Date)] [1/3] 备份失败: $_" -ForegroundColor Red
    Write-Host "[$(Get-Date)]       错误日志: $ErrorFile" -ForegroundColor Red
    exit 1
} finally {
    if (Test-Path $tempSql) { Remove-Item $tempSql -Force }
    Remove-Item Env:\PGPASSWORD -ErrorAction SilentlyContinue
}

# 3. Redis 备份提示（Windows 下需通过 docker exec 触发）
Write-Host "[$(Get-Date)] [2/3] Redis 备份提示：" -ForegroundColor Yellow
Write-Host "  Windows 环境下请手动执行："
Write-Host "    docker exec wuchang-redis redis-cli -a redis123 BGSAVE"
Write-Host "  然后从 Redis 数据卷复制 dump.rdb 到 $BackupDir"

# 4. 清理过期备份
Write-Host "[$(Get-Date)] [3/3] 清理 $RetentionDays 天前的过期备份..." -ForegroundColor Yellow
$cutoff = (Get-Date).AddDays(-$RetentionDays)
Get-ChildItem -Path $BackupDir -Filter "wuchang_pg_*.sql.gz" | Where-Object { $_.LastWriteTime -lt $cutoff } | Remove-Item -Force
Get-ChildItem -Path $BackupDir -Filter "wuchang_pg_*.err" | Where-Object { $_.LastWriteTime -lt $cutoff } | Remove-Item -Force
Write-Host "[$(Get-Date)] [3/3] 清理完成" -ForegroundColor Green

# ===== 离线副本上传（可选，需配置七牛云 qshell） =====
# TODO: 上传七牛云 OSS（需要配置七牛云 SDK + qshell）
# $qshell = Get-Command qshell -ErrorAction SilentlyContinue
# if ($qshell) {
#     $bucket = if ($env:QINIU_BUCKET) { $env:QINIU_BUCKET } else { "wuchang-tongcheng-backup" }
#     & qshell rput $bucket "backup/$(Split-Path -Leaf $BackupFile)" $BackupFile
#     Write-Host "[$(Get-Date)] 已上传至七牛云 $bucket" -ForegroundColor Green
# }

Write-Host "================================================" -ForegroundColor Cyan
Write-Host "[$(Get-Date)] 五常同城备份流程结束" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
