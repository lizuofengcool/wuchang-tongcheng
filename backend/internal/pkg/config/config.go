// Package config 配置管理
// 基于viper的配置管理，支持yaml配置文件和环境变量
package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 全局配置结构体
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Logger   LoggerConfig   `mapstructure:"logger"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
	ES       ESConfig       `mapstructure:"elasticsearch"`
	Storage  StorageConfig  `mapstructure:"storage"`
	Map      MapConfig      `mapstructure:"map"`
	SMS      SMSConfig      `mapstructure:"sms"`
	STS      STSConfig      `mapstructure:"sts"`
	OAuth    OAuthConfig    `mapstructure:"oauth"`
}

// OAuthConfig 第三方登录配置
type OAuthConfig struct {
	WeChat WeChatConfig `mapstructure:"wechat"`
}

// WeChatConfig 微信开放平台网站应用 OAuth 配置
//
// provider 取值：
//   - ""      → 不启用微信登录（/login/oauth/wechat 返回未启用）
//   - "mock"  → 联调模式：code 形如 "mock:<openid>[:<nickname>]" 直接构造身份，不访问微信
//   - "wechat"→ 真实微信 OAuth，AppID/AppSecret 齐全且非占位（your-）才激活，否则降级不启用
type WeChatConfig struct {
	Provider  string `mapstructure:"provider"`   // ""/mock/wechat
	AppID     string `mapstructure:"app_id"`     // 微信开放平台 AppID
	AppSecret string `mapstructure:"app_secret"` // 微信开放平台 AppSecret
}

// STSConfig 阿里云 OSS STS 临时凭据直传配置
//
// 与 storage 配置分离：STS 需要独立的 RAM 角色 ARN（扮演一个授予 OSS 写权限的角色），
// 凭据由后端用 AccessKey/SecretKey 调用 AssumeRole 换取后下发给前端直传。
// provider 非 aliyun 或任一必填项缺失/占位时自动降级 NoopProvider（AssumeRole 返回 ErrNotConfigured）。
type STSConfig struct {
	Provider        string `mapstructure:"provider"`         // ""/aliyun（aliyun 且 AK/SK/RoleArn 齐全才激活）
	AccessKey       string `mapstructure:"access_key"`       // 调用 AssumeRole 的 RAM 用户 AK
	SecretKey       string `mapstructure:"secret_key"`       // 调用 AssumeRole 的 RAM 用户 SK
	RoleArn         string `mapstructure:"role_arn"`         // 被扮演的 RAM 角色 ARN（授予 OSS 写权限）
	RoleSessionName string `mapstructure:"role_session_name"` // 会话名，默认 wuchang-upload
	DurationSeconds int    `mapstructure:"duration_seconds"`  // 凭据有效期（秒），范围 900~3600，默认 3600
	// OSS 落地信息（透传给前端，构造 OSS 客户端用）
	Bucket       string `mapstructure:"bucket"`        // OSS 桶名
	Region       string `mapstructure:"region"`        // OSS 区域，如 oss-cn-hangzhou
	Endpoint     string `mapstructure:"endpoint"`      // OSS 端点，如 https://oss-cn-hangzhou.aliyuncs.com
	ObjectPrefix string `mapstructure:"object_prefix"` // 对象 key 前缀，如 uploads/
}

// SMSConfig 短信验证码服务配置
type SMSConfig struct {
	Provider     string `mapstructure:"provider"`        // ""/mock（不发短信，dev 可返回验证码）/aliyun（dysmsapi RPC API，AK/SK/SignName/TemplateCode 未配置降级 mock）
	SignName     string `mapstructure:"sign_name"`       // 短信签名
	TemplateCode string `mapstructure:"template_code"`   // 短信模板
	AccessKey    string `mapstructure:"access_key"`      // 短信服务 AK
	SecretKey    string `mapstructure:"secret_key"`      // 短信服务 SK
	CodeTTL      int    `mapstructure:"code_ttl"`        // 验证码有效期（秒），默认 300
	CodeLength   int    `mapstructure:"code_length"`     // 验证码位数，默认 6
	MaxAttempts  int    `mapstructure:"max_attempts"`    // 最大尝试次数，默认 5
	DevReturnCode bool  `mapstructure:"dev_return_code"` // 开发模式：发送接口返回验证码明文（仅 mock provider 生效）
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret string `mapstructure:"secret"` // 签名密钥
	Expire int    `mapstructure:"expire"` // 过期时间（小时）
	Issuer string `mapstructure:"issuer"` // 签发者
}

// ServerConfig 服务配置
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug, release, test
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
	TimeZone string `mapstructure:"timezone"`
	// 连接池配置
	MaxOpenConns int `mapstructure:"max_open_conns"`
	MaxIdleConns int `mapstructure:"max_idle_conns"`
	MaxLifetime  int `mapstructure:"max_lifetime"` // 秒
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	// 连接池配置
	PoolSize int `mapstructure:"pool_size"`
	// 过期时间（秒）
	DefaultExpiration int `mapstructure:"default_expiration"`
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level      string `mapstructure:"level"`      // debug, info, warn, error
	Filename   string `mapstructure:"filename"`   // 日志文件路径
	MaxSize    int    `mapstructure:"max_size"`   // 每个日志文件最大大小（MB）
	MaxBackups int    `mapstructure:"max_backups"` // 保留的旧日志文件最大数量
	MaxAge     int    `mapstructure:"max_age"`    // 保留旧日志文件的最大天数
	Compress   bool   `mapstructure:"compress"`   // 是否压缩旧日志文件
	Console    bool   `mapstructure:"console"`    // 是否输出到控制台
}

// RabbitMQConfig RabbitMQ配置
type RabbitMQConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	VHost    string `mapstructure:"vhost"`
}

// ESConfig Elasticsearch配置
type ESConfig struct {
	Addresses []string `mapstructure:"addresses"`
	Username  string   `mapstructure:"username"`
	Password  string   `mapstructure:"password"`
}

// StorageConfig 对象存储配置
type StorageConfig struct {
	Type      string `mapstructure:"type"` // qiniu, local
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Bucket    string `mapstructure:"bucket"`
	Domain    string `mapstructure:"domain"`
	Region    string `mapstructure:"region"`
}

// MapConfig 地图服务配置
type MapConfig struct {
	Type string `mapstructure:"type"` // amap, baidu
	Key  string `mapstructure:"key"`
}

var (
	globalConfig *Config
)

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// 设置配置文件
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// 环境变量配置
	v.SetEnvPrefix("WCTC") // WUCHANG TONGCHENG
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config file failed: %w", err)
	}

	// 解析配置
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config failed: %w", err)
	}

	// 设置默认值
	setDefaults(&cfg)

	globalConfig = &cfg
	return &cfg, nil
}

// setDefaults 设置默认配置值
func setDefaults(cfg *Config) {
	// Server默认值
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "debug"
	}

	// Database默认值
	if cfg.Database.Host == "" {
		cfg.Database.Host = "localhost"
	}
	if cfg.Database.Port == 0 {
		cfg.Database.Port = 5432
	}
	if cfg.Database.User == "" {
		cfg.Database.User = "postgres"
	}
	if cfg.Database.DBName == "" {
		cfg.Database.DBName = "wuchang_tongcheng"
	}
	if cfg.Database.SSLMode == "" {
		cfg.Database.SSLMode = "disable"
	}
	if cfg.Database.TimeZone == "" {
		cfg.Database.TimeZone = "Asia/Shanghai"
	}
	if cfg.Database.MaxOpenConns == 0 {
		cfg.Database.MaxOpenConns = 100
	}
	if cfg.Database.MaxIdleConns == 0 {
		cfg.Database.MaxIdleConns = 10
	}
	if cfg.Database.MaxLifetime == 0 {
		cfg.Database.MaxLifetime = 3600
	}

	// Redis默认值
	if cfg.Redis.Host == "" {
		cfg.Redis.Host = "localhost"
	}
	if cfg.Redis.Port == 0 {
		cfg.Redis.Port = 6379
	}
	if cfg.Redis.PoolSize == 0 {
		cfg.Redis.PoolSize = 10
	}
	if cfg.Redis.DefaultExpiration == 0 {
		cfg.Redis.DefaultExpiration = 3600
	}

	// Logger默认值
	if cfg.Logger.Level == "" {
		cfg.Logger.Level = "info"
	}
	if cfg.Logger.MaxSize == 0 {
		cfg.Logger.MaxSize = 100
	}
	if cfg.Logger.MaxBackups == 0 {
		cfg.Logger.MaxBackups = 3
	}
	if cfg.Logger.MaxAge == 0 {
		cfg.Logger.MaxAge = 30
	}

	// SMS 默认值
	if cfg.SMS.CodeTTL == 0 {
		cfg.SMS.CodeTTL = 300
	}
	if cfg.SMS.CodeLength == 0 {
		cfg.SMS.CodeLength = 6
	}
	if cfg.SMS.MaxAttempts == 0 {
		cfg.SMS.MaxAttempts = 5
	}

	// STS 默认值
	if cfg.STS.DurationSeconds == 0 {
		cfg.STS.DurationSeconds = 3600
	}
}

// Get 获取全局配置
func Get() *Config {
	if globalConfig == nil {
		panic("config not loaded, call Load() first")
	}
	return globalConfig
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode, c.TimeZone)
}

// GetAddr 获取服务监听地址
func (c *ServerConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetRedisAddr 获取Redis地址
func (c *RedisConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetRabbitMQURL 获取RabbitMQ连接URL
func (c *RabbitMQConfig) GetURL() string {
	if c.VHost == "" {
		c.VHost = "/"
	}
	return fmt.Sprintf("amqp://%s:%s@%s:%d%s",
		c.User, c.Password, c.Host, c.Port, c.VHost)
}
