// Package pgtest 提供基于 testcontainers 的 PostgreSQL 集成测试夹具
//
// 用法：
//
//	db := pgtest.SetupPostgres(t)             // 启动 PG 容器，返回 *gorm.DB（自动 t.Cleanup 终止容器）
//	pgtest.MigrateAll(t, db)                  // AutoMigrate 全部业务模型
//	repo := repository.NewUserRepository(db) // 注入真实 DB 构造 repository
//
// Docker 不可用时会 t.Skip，保证 `go test ./...` 在无 Docker 环境（如 CI 短测）下仍可通过。
package pgtest

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	// 各模块 model，用于 MigrateAll 全量建表
	categoryModel "wuchang-tongcheng/internal/modules/category/model"
	fileModel "wuchang-tongcheng/internal/modules/file/model"
	newsModel "wuchang-tongcheng/internal/modules/news/model"
	permModel "wuchang-tongcheng/internal/modules/permission/model"
	regionModel "wuchang-tongcheng/internal/modules/region/model"
	settingModel "wuchang-tongcheng/internal/modules/setting/model"
	userModel "wuchang-tongcheng/internal/modules/user/model"
)

// allModels 全部业务模型，MigrateAll 会对其执行 AutoMigrate。
// 顺序无关（GORM 通过反射建表），但保持稳定顺序便于排查。
var allModels = []interface{}{
	&userModel.User{},
	&regionModel.Region{},
	&categoryModel.Category{},
	&newsModel.News{},
	&newsModel.NewsLike{},
	&permModel.Role{},
	&permModel.Permission{},
	&permModel.UserRole{},
	&permModel.RolePermission{},
	&fileModel.FileUpload{},
	&settingModel.Setting{},
}

// allModels 所有 model 的 TableName 已在各 model 文件中显式定义。
// 以下 init 用于编译期断言：确保模型实现了 TableName() string（防误删）。
func init() {
	for _, m := range allModels {
		if _, ok := any(m).(interface{ TableName() string }); !ok {
			panic(fmt.Sprintf("pgtest: model %T 未实现 TableName()", m))
		}
	}
}

// dockerHostSkipReasons 检查 Docker 是否可能可用。
// 返回非空字符串表示不可用原因（调用方据此 t.Skip）。
// 这是一组启发式检查，避免在无 Docker 环境下等待 testcontainers 长重试。
func dockerUnavailableReason() string {
	// 1. 显式禁用开关：WCTC_SKIP_INTEGRATION=1 直接跳过
	if os.Getenv("WCTC_SKIP_INTEGRATION") == "1" {
		return "WCTC_SKIP_INTEGRATION=1 已设置"
	}
	// 2. DOCKER_HOST 指定了远程 docker（testcontainers 可连）
	if os.Getenv("DOCKER_HOST") != "" {
		return ""
	}
	// 3. docker CLI 存在
	if _, err := exec.LookPath("docker"); err == nil {
		return ""
	}
	// 4. 默认 unix socket 存在
	if fi, err := os.Stat("/var/run/docker.sock"); err == nil && fi.Mode()&os.ModeSocket != 0 {
		return ""
	}
	return "docker CLI/daemon 不可用（无 docker 命令、无 /var/run/docker.sock、未设 DOCKER_HOST）"
}

// SetupPostgres 启动一个独立的 PostgreSQL 容器并返回连接到它的 *gorm.DB。
// 失败（如 Docker 不可用）会 t.Skip 而非 t.Fatal，保证集成测试可被优雅跳过。
// 容器终止、DB 关闭均通过 t.Cleanup 自动处理。
func SetupPostgres(t testing.TB) *gorm.DB {
	t.Helper()

	if reason := dockerUnavailableReason(); reason != "" {
		t.Skipf("跳过集成测试：%s", reason)
	}

	ctx := context.Background()

	// 用 postgres:16-alpine，启用 basic 等待策略（等待 ready for connections）
	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("wctc_test"),
		tcpostgres.WithUsername("wctc"),
		tcpostgres.WithPassword("wctc_pwd"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Skipf("跳过集成测试：启动 PostgreSQL 容器失败（%v）；请确认 Docker daemon 可用", err)
		return nil
	}
	t.Cleanup(func() {
		_ = container.Terminate(context.Background())
	})

	dsn, err := container.ConnectionString(ctx, "sslmode=disable", "TimeZone=Asia/Shanghai")
	if err != nil {
		t.Fatalf("获取 PG 连接串失败：%v", err)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		t.Fatalf("连接 PG 容器失败：%v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("获取 *sql.DB 失败：%v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	return db
}

// MigrateAll 对 db 执行全部业务模型的 AutoMigrate，并断言每张表创建成功。
// 失败直接 t.Fatal（建表出错说明夹具本身有问题，不应跳过）。
func MigrateAll(t testing.TB, db *gorm.DB) {
	t.Helper()
	if err := db.AutoMigrate(allModels...); err != nil {
		t.Fatalf("AutoMigrate 全部模型失败：%v", err)
	}
}

// SnapshotName 生成基于测试名的唯一快照/数据库名，便于复用容器的场景。
func SnapshotName(t testing.TB) string {
	return filepath.Base(t.Name())
}
