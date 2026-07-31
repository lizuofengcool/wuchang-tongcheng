package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// withCleanEnv 保存并清除所有 WCTC_ 前缀环境变量，返回恢复函数。
// Load 内部调用 viper.AutomaticEnv + SetEnvPrefix("WCTC")，会把 WCTC_SERVER_PORT
// 这类变量注入到解析结果并覆盖文件值，导致 Load 测试不可复现。
// 测试期间临时清空 WCTC_* 变量可保证仅以 yaml 文件内容为准，t.Cleanup 恢复原始环境。
func withCleanEnv(t *testing.T) {
	t.Helper()
	type kv struct{ k, v string }
	var saved []kv
	for _, e := range os.Environ() {
		// 环境变量条目形如 KEY=VALUE；仅处理 WCTC_ 前缀（大小写敏感，viper 前缀按原样匹配）
		if strings.HasPrefix(e, "WCTC_") {
			if i := strings.IndexByte(e, '='); i >= 0 {
				saved = append(saved, kv{e[:i], e[i+1:]})
				os.Unsetenv(e[:i])
			}
		}
	}
	t.Cleanup(func() {
		for _, p := range saved {
			os.Setenv(p.k, p.v)
		}
	})
}

// resetGlobal 清除 config 包级 globalConfig，保证 Get() 行为在用例间隔离。
func resetGlobal(t *testing.T) {
	t.Helper()
	globalConfig = nil
	t.Cleanup(func() { globalConfig = nil })
}

// writeYaml 写入临时 yaml 文件并返回其路径，t.Cleanup 自动清理。
func writeYaml(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write yaml: %v", err)
	}
	return path
}

// --- setDefaults 白盒测试 ---

func TestSetDefaults_EmptyConfigAppliesAllDefaults(t *testing.T) {
	var cfg Config
	setDefaults(&cfg)

	// Server 默认值
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Server.Host = %q, want 0.0.0.0", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want 8080", cfg.Server.Port)
	}
	if cfg.Server.Mode != "debug" {
		t.Errorf("Server.Mode = %q, want debug", cfg.Server.Mode)
	}

	// Database 默认值
	if cfg.Database.Host != "localhost" {
		t.Errorf("Database.Host = %q, want localhost", cfg.Database.Host)
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("Database.Port = %d, want 5432", cfg.Database.Port)
	}
	if cfg.Database.User != "postgres" {
		t.Errorf("Database.User = %q, want postgres", cfg.Database.User)
	}
	if cfg.Database.DBName != "wuchang_tongcheng" {
		t.Errorf("Database.DBName = %q, want wuchang_tongcheng", cfg.Database.DBName)
	}
	if cfg.Database.SSLMode != "disable" {
		t.Errorf("Database.SSLMode = %q, want disable", cfg.Database.SSLMode)
	}
	if cfg.Database.TimeZone != "Asia/Shanghai" {
		t.Errorf("Database.TimeZone = %q, want Asia/Shanghai", cfg.Database.TimeZone)
	}
	if cfg.Database.MaxOpenConns != 100 {
		t.Errorf("Database.MaxOpenConns = %d, want 100", cfg.Database.MaxOpenConns)
	}
	if cfg.Database.MaxIdleConns != 10 {
		t.Errorf("Database.MaxIdleConns = %d, want 10", cfg.Database.MaxIdleConns)
	}
	if cfg.Database.MaxLifetime != 3600 {
		t.Errorf("Database.MaxLifetime = %d, want 3600", cfg.Database.MaxLifetime)
	}

	// Redis 默认值
	if cfg.Redis.Host != "localhost" {
		t.Errorf("Redis.Host = %q, want localhost", cfg.Redis.Host)
	}
	if cfg.Redis.Port != 6379 {
		t.Errorf("Redis.Port = %d, want 6379", cfg.Redis.Port)
	}
	if cfg.Redis.PoolSize != 10 {
		t.Errorf("Redis.PoolSize = %d, want 10", cfg.Redis.PoolSize)
	}
	if cfg.Redis.DefaultExpiration != 3600 {
		t.Errorf("Redis.DefaultExpiration = %d, want 3600", cfg.Redis.DefaultExpiration)
	}

	// Logger 默认值
	if cfg.Logger.Level != "info" {
		t.Errorf("Logger.Level = %q, want info", cfg.Logger.Level)
	}
	if cfg.Logger.MaxSize != 100 {
		t.Errorf("Logger.MaxSize = %d, want 100", cfg.Logger.MaxSize)
	}
	if cfg.Logger.MaxBackups != 3 {
		t.Errorf("Logger.MaxBackups = %d, want 3", cfg.Logger.MaxBackups)
	}
	if cfg.Logger.MaxAge != 30 {
		t.Errorf("Logger.MaxAge = %d, want 30", cfg.Logger.MaxAge)
	}

	// SMS 默认值
	if cfg.SMS.CodeTTL != 300 {
		t.Errorf("SMS.CodeTTL = %d, want 300", cfg.SMS.CodeTTL)
	}
	if cfg.SMS.CodeLength != 6 {
		t.Errorf("SMS.CodeLength = %d, want 6", cfg.SMS.CodeLength)
	}
	if cfg.SMS.MaxAttempts != 5 {
		t.Errorf("SMS.MaxAttempts = %d, want 5", cfg.SMS.MaxAttempts)
	}

	// STS 默认值
	if cfg.STS.DurationSeconds != 3600 {
		t.Errorf("STS.DurationSeconds = %d, want 3600", cfg.STS.DurationSeconds)
	}
}

func TestSetDefaults_NonZeroValuesArePreserved(t *testing.T) {
	// 全部字段显式赋值，setDefaults 不应覆盖任何非零字段
	in := Config{
		Server:   ServerConfig{Host: "1.2.3.4", Port: 9090, Mode: "release"},
		Database: DatabaseConfig{Host: "db.host", Port: 6543, User: "u", Password: "p", DBName: "mydb", SSLMode: "require", TimeZone: "UTC", MaxOpenConns: 50, MaxIdleConns: 5, MaxLifetime: 600},
		Redis:    RedisConfig{Host: "r.host", Port: 6380, Password: "rp", DB: 2, PoolSize: 20, DefaultExpiration: 120},
		Logger:   LoggerConfig{Level: "warn", Filename: "/var/log/a.log", MaxSize: 50, MaxBackups: 7, MaxAge: 14, Compress: true, Console: true},
		SMS:      SMSConfig{Provider: "aliyun", SignName: "s", TemplateCode: "t", AccessKey: "ak", SecretKey: "sk", CodeTTL: 180, CodeLength: 4, MaxAttempts: 3, DevReturnCode: true},
		STS:      STSConfig{Provider: "aliyun", AccessKey: "a", SecretKey: "s", RoleArn: "arn", RoleSessionName: "sess", DurationSeconds: 900, Bucket: "b", Region: "r", Endpoint: "e", ObjectPrefix: "pre/"},
	}
	orig := in // 值拷贝快照
	setDefaults(&in)

	if !reflect.DeepEqual(in, orig) {
		t.Errorf("setDefaults 覆盖了已赋值字段:\n got  = %+v\n want = %+v", in, orig)
	}
}

func TestSetDefaults_ZeroishButMeaningfulFieldsUntouched(t *testing.T) {
	// setDefaults 不处理这些字段，确认它们保持零值不被误填默认
	var cfg Config
	setDefaults(&cfg)

	if cfg.Database.Password != "" {
		t.Errorf("Database.Password 被误填默认值: %q", cfg.Database.Password)
	}
	if cfg.Redis.Password != "" {
		t.Errorf("Redis.Password 被误填默认值: %q", cfg.Redis.Password)
	}
	if cfg.Redis.DB != 0 {
		t.Errorf("Redis.DB 被误填默认值: %d", cfg.Redis.DB)
	}
	if cfg.Logger.Filename != "" {
		t.Errorf("Logger.Filename 被误填默认值: %q", cfg.Logger.Filename)
	}
	if cfg.Logger.Compress {
		t.Errorf("Logger.Compress 被误填默认值: true")
	}
	if cfg.Logger.Console {
		t.Errorf("Logger.Console 被误填默认值: true")
	}
}

// --- 格式化方法测试 ---

func TestDatabaseConfig_GetDSN(t *testing.T) {
	c := DatabaseConfig{Host: "h", Port: 5432, User: "u", Password: "p", DBName: "db", SSLMode: "disable", TimeZone: "Asia/Shanghai"}
	want := "host=h port=5432 user=u password=p dbname=db sslmode=disable TimeZone=Asia/Shanghai"
	if got := c.GetDSN(); got != want {
		t.Errorf("GetDSN() = %q, want %q", got, want)
	}
}

func TestDatabaseConfig_GetDSN_EmptyPassword(t *testing.T) {
	// 空密码应在串中保留 "password=" 占位，避免拼出非法 DSN
	c := DatabaseConfig{Host: "h", Port: 5432, User: "u", DBName: "db", SSLMode: "disable", TimeZone: "UTC"}
	want := "host=h port=5432 user=u password= dbname=db sslmode=disable TimeZone=UTC"
	if got := c.GetDSN(); got != want {
		t.Errorf("GetDSN() = %q, want %q", got, want)
	}
}

func TestServerConfig_GetAddr(t *testing.T) {
	c := ServerConfig{Host: "0.0.0.0", Port: 8080}
	if got, want := c.GetAddr(), "0.0.0.0:8080"; got != want {
		t.Errorf("GetAddr() = %q, want %q", got, want)
	}
}

func TestRedisConfig_GetAddr(t *testing.T) {
	c := RedisConfig{Host: "localhost", Port: 6379}
	if got, want := c.GetAddr(), "localhost:6379"; got != want {
		t.Errorf("GetAddr() = %q, want %q", got, want)
	}
}

func TestRabbitMQConfig_GetURL_VHostAppendedVerbatim(t *testing.T) {
	// GetURL 将 VHost 原样拼接到 host:port 之后，不补前导 "/"；
	// 因此标准 amqp URI 的 /prod 需调用方在配置里写 "/prod"。
	c := RabbitMQConfig{Host: "mq", Port: 5672, User: "guest", Password: "guest", VHost: "/prod"}
	want := "amqp://guest:guest@mq:5672/prod"
	if got := c.GetURL(); got != want {
		t.Errorf("GetURL() = %q, want %q", got, want)
	}
	if c.VHost != "/prod" {
		t.Errorf("VHost 被误改: got %q, want /prod", c.VHost)
	}
}

func TestRabbitMQConfig_GetURL_VHostWithoutLeadingSlash(t *testing.T) {
	// 文档化：无前导 "/" 的 vhost 会原样拼接，得到 ...5672prod（非标准 amqp URI）
	c := RabbitMQConfig{Host: "mq", Port: 5672, User: "u", Password: "p", VHost: "prod"}
	want := "amqp://u:p@mq:5672prod"
	if got := c.GetURL(); got != want {
		t.Errorf("GetURL() = %q, want %q", got, want)
	}
}

func TestRabbitMQConfig_GetURL_EmptyVHostDefaultsToSlash(t *testing.T) {
	// 空 vhost 时 GetURL 会就地写入 "/"，URL 末尾带 "/"
	c := RabbitMQConfig{Host: "mq", Port: 5672, User: "u", Password: "p", VHost: ""}
	want := "amqp://u:p@mq:5672/"
	if got := c.GetURL(); got != want {
		t.Errorf("GetURL() = %q, want %q", got, want)
	}
	if c.VHost != "/" {
		t.Errorf("空 VHost 未被补默认 /: got %q", c.VHost)
	}
}

// --- Load 测试 ---

func TestLoad_HappyPath(t *testing.T) {
	withCleanEnv(t)
	resetGlobal(t)

	yaml := `
server:
  host: "127.0.0.1"
  port: 9000
  mode: "release"
database:
  host: "db.example.com"
  port: 6543
  user: "app"
  password: "secret"
  dbname: "prod_db"
  sslmode: "require"
  timezone: "UTC"
  max_open_conns: 25
  max_idle_conns: 5
  max_lifetime: 300
redis:
  host: "redis.example.com"
  port: 6380
  password: "rpw"
  db: 1
  pool_size: 15
  default_expiration: 7200
`
	path := writeYaml(t, "config.yaml", yaml)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load 返回错误: %v", err)
	}

	// 文件值应透传
	if cfg.Server.Host != "127.0.0.1" || cfg.Server.Port != 9000 || cfg.Server.Mode != "release" {
		t.Errorf("Server 透传失败: %+v", cfg.Server)
	}
	if cfg.Database.Host != "db.example.com" || cfg.Database.Port != 6543 || cfg.Database.DBName != "prod_db" || cfg.Database.SSLMode != "require" {
		t.Errorf("Database 透传失败: %+v", cfg.Database)
	}
	if cfg.Redis.Host != "redis.example.com" || cfg.Redis.Port != 6380 || cfg.Redis.Password != "rpw" || cfg.Redis.DB != 1 {
		t.Errorf("Redis 透传失败: %+v", cfg.Redis)
	}

	// 未在文件中给出的字段应被 setDefaults 补默认
	if cfg.Database.MaxOpenConns != 25 {
		t.Errorf("Database.MaxOpenConns 文件值 25 被覆盖: %d", cfg.Database.MaxOpenConns)
	}
	if cfg.Logger.Level != "info" {
		t.Errorf("Logger.Level 未补默认 info: %q", cfg.Logger.Level)
	}
	if cfg.SMS.CodeTTL != 300 {
		t.Errorf("SMS.CodeTTL 未补默认 300: %d", cfg.SMS.CodeTTL)
	}
	if cfg.STS.DurationSeconds != 3600 {
		t.Errorf("STS.DurationSeconds 未补默认 3600: %d", cfg.STS.DurationSeconds)
	}

	// Load 成功后 globalConfig 应被设置，Get() 返回同一指针
	if Get() != cfg {
		t.Errorf("Get() 返回的指针与 Load 返回的不一致（globalConfig 未设置）")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	withCleanEnv(t)
	resetGlobal(t)

	_, err := Load(filepath.Join(t.TempDir(), "no-such-file.yaml"))
	if err == nil {
		t.Fatal("缺少文件时 Load 应返回错误，得到 nil")
	}
	if !strings.Contains(err.Error(), "read config file failed") {
		t.Errorf("错误信息未包含 read config file failed: %v", err)
	}

	// 失败路径不应设置 globalConfig
	if globalConfig != nil {
		t.Errorf("失败路径不应设置 globalConfig")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	withCleanEnv(t)
	resetGlobal(t)

	// 故意写入语法错误的 yaml：键后缺少冒号
	path := writeYaml(t, "bad.yaml", "server\n  host bad\n  port 8080\n")
	_, err := Load(path)
	if err == nil {
		t.Fatal("yaml 语法错误时 Load 应返回错误，得到 nil")
	}
	if !strings.Contains(err.Error(), "read config file failed") {
		t.Errorf("错误信息未包含 read config file failed: %v", err)
	}
}

// --- Get 测试 ---

func TestGet_PanicsWhenNotLoaded(t *testing.T) {
	withCleanEnv(t)
	resetGlobal(t)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("globalConfig 为 nil 时 Get() 应 panic，未触发")
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "config not loaded") {
			t.Errorf("panic 信息不含 config not loaded: %v", r)
		}
	}()
	_ = Get()
}

func TestGet_ReturnsLoadedConfig(t *testing.T) {
	withCleanEnv(t)
	resetGlobal(t)

	path := writeYaml(t, "c.yaml", "server:\n  port: 7777\n")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load 失败: %v", err)
	}

	got := Get()
	if got != cfg {
		t.Fatalf("Get() 指针与 Load 返回不一致")
	}
	if got.Server.Port != 7777 {
		t.Errorf("Get().Server.Port = %d, want 7777", got.Server.Port)
	}
}
