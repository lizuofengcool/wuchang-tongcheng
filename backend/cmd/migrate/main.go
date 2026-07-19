// Package main 数据库迁移工具
//
// 提供简单的 SQL 文件迁移执行能力，不引入第三方迁移库。
// 迁移脚本存放于 backend/migrations/*.sql：
//   - 正向迁移：{序号}_{名称}.sql（不含 _rollback 后缀）
//   - 回滚迁移：{序号}_{名称}_rollback.sql
//
// 状态记录表 schema_migrations 在首次执行时自动创建。
//
// 用法：
//
//	go run ./cmd/migrate up                    # 执行所有未应用的正向迁移
//	go run ./cmd/migrate down                  # 回滚所有已应用的迁移（执行回滚脚本）
//	go run ./cmd/migrate status                # 查看迁移状态
//	go run ./cmd/migrate up --module=ershou    # 按模块独立迁移（预留，暂未实现）
//	go run ./cmd/migrate up --config=./configs/config.yaml  # 指定配置文件
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"wuchang-tongcheng/internal/pkg/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ANSI 颜色转义码（Windows Terminal / PowerShell 5+ / 现代 cmd 均支持）
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
)

// 默认路径（相对于 backend/ 目录执行）
const (
	defaultMigrationsDir = "./migrations"
	defaultConfigPath    = "./configs/config.yaml"
)

// Migration 表示一个迁移文件
type Migration struct {
	Filename   string // 文件名（含扩展名）
	BaseName   string // 文件名（不含扩展名）
	Path       string // 完整路径
	IsRollback bool   // 是否为回滚脚本
}

// schemaMigrationRecord schema_migrations 表记录
type schemaMigrationRecord struct {
	Filename  string    `gorm:"primaryKey;size:256" json:"filename"`
	AppliedAt time.Time `gorm:"not null;default:now()" json:"applied_at"`
}

// TableName 指定表名
func (schemaMigrationRecord) TableName() string {
	return "schema_migrations"
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// 解析子命令参数
	fs := flag.NewFlagSet("migrate", flag.ExitOnError)
	configPath := fs.String("config", defaultConfigPath, "配置文件路径")
	migrationsDir := fs.String("dir", defaultMigrationsDir, "迁移脚本目录")
	module := fs.String("module", "", "按模块独立迁移（预留，暂未实现）")
	_ = fs.Parse(os.Args[2:])

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		printError("加载配置失败: %v", err)
		os.Exit(1)
	}
	printInfo("配置加载成功: %s", *configPath)

	// 连接数据库
	printInfo("正在连接数据库 %s:%d/%s ...", cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)
	db, err := gorm.Open(postgres.Open(cfg.Database.GetDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		printError("数据库连接失败: %v", err)
		os.Exit(1)
	}
	defer func() {
		if sqlDB, err := db.DB(); err == nil && sqlDB != nil {
			_ = sqlDB.Close()
		}
	}()
	printSuccess("数据库连接成功")

	// 确保 schema_migrations 表存在
	if err := ensureMigrationsTable(db); err != nil {
		printError("创建 schema_migrations 表失败: %v", err)
		os.Exit(1)
	}

	// 收集迁移文件
	migrations, err := collectMigrations(*migrationsDir)
	if err != nil {
		printError("收集迁移文件失败: %v", err)
		os.Exit(1)
	}
	printInfo("发现 %d 个迁移文件（%d 正向 / %d 回滚）",
		len(migrations), countForward(migrations), countRollback(migrations))

	// 执行命令
	switch command {
	case "up":
		if *module != "" {
			printWarn("--module=%s 参数已记录，按模块独立迁移功能暂未实现，将执行全量迁移", *module)
		}
		err = cmdUp(db, migrations)
	case "down":
		err = cmdDown(db, migrations)
	case "status":
		err = cmdStatus(db, migrations)
	case "help", "-h", "--help":
		printUsage()
		return
	default:
		printError("未知命令: %s", command)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		printError("%v", err)
		os.Exit(1)
	}
	printSuccess("迁移命令 '%s' 执行完成", command)
}

// ensureMigrationsTable 创建 schema_migrations 表（幂等）
func ensureMigrationsTable(db *gorm.DB) error {
	sql := `
CREATE TABLE IF NOT EXISTS schema_migrations (
    filename  VARCHAR(256) PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_schema_migrations_applied_at ON schema_migrations(applied_at);
`
	return db.Exec(sql).Error
}

// collectMigrations 收集迁移目录下所有 .sql 文件
func collectMigrations(dir string) ([]Migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("读取迁移目录 %s 失败: %w", dir, err)
	}

	var migrations []Migration
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".sql") {
			continue
		}
		base := strings.TrimSuffix(name, filepath.Ext(name))
		migrations = append(migrations, Migration{
			Filename:   name,
			BaseName:   base,
			Path:       filepath.Join(dir, name),
			IsRollback: strings.Contains(base, "_rollback"),
		})
	}

	// 按文件名升序排序
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Filename < migrations[j].Filename
	})
	return migrations, nil
}

// cmdUp 执行所有未应用的正向迁移
func cmdUp(db *gorm.DB, migrations []Migration) error {
	forward := filterForward(migrations)
	if len(forward) == 0 {
		printWarn("未发现正向迁移文件")
		return nil
	}

	applied, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("查询已应用迁移失败: %w", err)
	}

	pendingCount := 0
	for _, m := range forward {
		if _, ok := applied[m.Filename]; ok {
			printInfo("  [已应用] %s", m.Filename)
			continue
		}
		pendingCount++
		if err := applyMigration(db, m); err != nil {
			return fmt.Errorf("执行迁移 %s 失败: %w", m.Filename, err)
		}
	}

	if pendingCount == 0 {
		printInfo("所有正向迁移均已应用，无需操作")
	} else {
		printSuccess("已应用 %d 个正向迁移", pendingCount)
	}
	return nil
}

// cmdDown 执行所有回滚脚本（按文件名降序）
func cmdDown(db *gorm.DB, migrations []Migration) error {
	rollbacks := filterRollback(migrations)
	if len(rollbacks) == 0 {
		printWarn("未发现回滚脚本")
		return nil
	}

	// 查询已应用的正向迁移
	applied, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("查询已应用迁移失败: %w", err)
	}
	if len(applied) == 0 {
		printInfo("无已应用的迁移，无需回滚")
		return nil
	}

	// 按文件名降序执行回滚脚本
	sort.Slice(rollbacks, func(i, j int) bool {
		return rollbacks[i].Filename > rollbacks[j].Filename
	})

	rolledCount := 0
	for _, m := range rollbacks {
		if err := applyRollback(db, m); err != nil {
			return fmt.Errorf("执行回滚 %s 失败: %w", m.Filename, err)
		}
		rolledCount++
	}

	// 清空 schema_migrations 表（所有正向迁移均被回滚）
	if err := db.Where("1 = 1").Delete(&schemaMigrationRecord{}).Error; err != nil {
		printWarn("清空 schema_migrations 表失败: %v（回滚已执行，但状态记录未清理）", err)
	} else {
		printInfo("已清空 schema_migrations 表")
	}

	printSuccess("已执行 %d 个回滚脚本", rolledCount)
	return nil
}

// cmdStatus 查看迁移状态
func cmdStatus(db *gorm.DB, migrations []Migration) error {
	forward := filterForward(migrations)
	if len(forward) == 0 {
		printWarn("未发现正向迁移文件")
		return nil
	}

	applied, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("查询已应用迁移失败: %w", err)
	}

	fmt.Printf("\n%s迁移状态%s\n", colorCyan, colorReset)
	fmt.Printf("%s─────────────────────────────────────────────────────────────────%s\n", colorGray, colorReset)
	fmt.Printf("%-40s %-10s %-20s\n", "文件名", "状态", "应用时间")
	fmt.Printf("%s─────────────────────────────────────────────────────────────────%s\n", colorGray, colorReset)

	appliedCount := 0
	for _, m := range forward {
		status := colorYellow + "待应用" + colorReset
		appliedAt := "-"
		if at, ok := applied[m.Filename]; ok {
			status = colorGreen + "已应用" + colorReset
			appliedAt = at.Format("2006-01-02 15:04:05")
			appliedCount++
		}
		fmt.Printf("%-40s %-22s %-20s\n", m.Filename, status, appliedAt)
	}
	fmt.Printf("%s─────────────────────────────────────────────────────────────────%s\n", colorGray, colorReset)
	fmt.Printf("总计：%d 个正向迁移，%s%d 已应用%s / %s%d 待应用%s\n\n",
		len(forward),
		colorGreen, appliedCount, colorReset,
		colorYellow, len(forward)-appliedCount, colorReset)
	return nil
}

// applyMigration 执行单个正向迁移
func applyMigration(db *gorm.DB, m Migration) error {
	content, err := os.ReadFile(m.Path)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	printInfo("  [执行中] %s (%d 字节)", m.Filename, len(content))

	// 在事务中执行 SQL + 记录 schema_migrations
	return db.Transaction(func(tx *gorm.DB) error {
		// pgx 驱动支持多语句执行（简单查询协议）
		if err := tx.Exec(string(content)).Error; err != nil {
			return fmt.Errorf("SQL 执行失败: %w", err)
		}
		// 记录已应用
		record := schemaMigrationRecord{Filename: m.Filename, AppliedAt: time.Now()}
		if err := tx.Create(&record).Error; err != nil {
			return fmt.Errorf("记录 schema_migrations 失败: %w", err)
		}
		return nil
	})
}

// applyRollback 执行单个回滚脚本
func applyRollback(db *gorm.DB, m Migration) error {
	content, err := os.ReadFile(m.Path)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	printInfo("  [回滚中] %s (%d 字节)", m.Filename, len(content))

	// 回滚脚本在事务中执行（不记录 schema_migrations，由调用方统一清理）
	if err := db.Exec(string(content)).Error; err != nil {
		return fmt.Errorf("SQL 执行失败: %w", err)
	}
	printSuccess("  [已回滚] %s", m.Filename)
	return nil
}

// getAppliedMigrations 查询已应用的迁移文件名集合
func getAppliedMigrations(db *gorm.DB) (map[string]time.Time, error) {
	var records []schemaMigrationRecord
	if err := db.Find(&records).Error; err != nil {
		return nil, err
	}
	result := make(map[string]time.Time, len(records))
	for _, r := range records {
		result[r.Filename] = r.AppliedAt
	}
	return result, nil
}

// filterForward 过滤正向迁移文件（按文件名升序）
func filterForward(migrations []Migration) []Migration {
	var result []Migration
	for _, m := range migrations {
		if !m.IsRollback {
			result = append(result, m)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Filename < result[j].Filename
	})
	return result
}

// filterRollback 过滤回滚迁移文件
func filterRollback(migrations []Migration) []Migration {
	var result []Migration
	for _, m := range migrations {
		if m.IsRollback {
			result = append(result, m)
		}
	}
	return result
}

func countForward(migrations []Migration) int {
	count := 0
	for _, m := range migrations {
		if !m.IsRollback {
			count++
		}
	}
	return count
}

func countRollback(migrations []Migration) int {
	count := 0
	for _, m := range migrations {
		if m.IsRollback {
			count++
		}
	}
	return count
}

// printUsage 打印用法
func printUsage() {
	fmt.Printf(`
%s五常同城 - 数据库迁移工具%s

%s用法:%s
  go run ./cmd/migrate <命令> [参数]

%s命令:%s
  %sup%s       执行所有未应用的正向迁移
  %sdown%s     回滚所有已应用的迁移（执行回滚脚本）
  %sstatus%s   查看迁移状态

%s参数:%s
  --config=<路径>      配置文件路径（默认 ./configs/config.yaml）
  --dir=<路径>         迁移脚本目录（默认 ./migrations）
  --module=<模块名>    按模块独立迁移（预留，暂未实现）

%s示例:%s
  go run ./cmd/migrate up
  go run ./cmd/migrate down
  go run ./cmd/migrate status
  go run ./cmd/migrate up --config=./configs/config.yaml --dir=./migrations

`,
		colorCyan, colorReset,
		colorYellow, colorReset,
		colorYellow, colorReset,
		colorGreen, colorReset,
		colorGreen, colorReset,
		colorGreen, colorReset,
		colorYellow, colorReset,
		colorYellow, colorReset)
}

// 彩色输出辅助函数
func printSuccess(format string, args ...interface{}) {
	fmt.Printf("%s[成功]%s %s\n", colorGreen, colorReset, fmt.Sprintf(format, args...))
}

func printError(format string, args ...interface{}) {
	fmt.Printf("%s[错误]%s %s\n", colorRed, colorReset, fmt.Sprintf(format, args...))
}

func printWarn(format string, args ...interface{}) {
	fmt.Printf("%s[警告]%s %s\n", colorYellow, colorReset, fmt.Sprintf(format, args...))
}

func printInfo(format string, args ...interface{}) {
	fmt.Printf("%s[信息]%s %s\n", colorBlue, colorReset, fmt.Sprintf(format, args...))
}
