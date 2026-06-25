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
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
	ES       ESConfig       `mapstructure:"elasticsearch"`
	Storage  StorageConfig  `mapstructure:"storage"`
	Map      MapConfig      `mapstructure:"map"`
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
