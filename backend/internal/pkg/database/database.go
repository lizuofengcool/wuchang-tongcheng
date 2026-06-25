// Package database 数据库封装
// 基于GORM的PostgreSQL数据库连接封装，支持连接池和自动迁移
package database

import (
	"fmt"
	"time"

	"wuchang-tongcheng/internal/pkg/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db *gorm.DB
)

// Init 初始化数据库连接
func Init(cfg *config.DatabaseConfig) error {
	dsn := cfg.GetDSN()

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	var err error
	db, err = gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("connect database failed: %w", err)
	}

	// 设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql.DB failed: %w", err)
	}

	// 最大打开连接数
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	// 最大空闲连接数
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	// 连接最大生命周期
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second)

	return nil
}

// GetDB 获取数据库连接
func GetDB() *gorm.DB {
	if db == nil {
		panic("database not initialized, call Init() first")
	}
	return db
}

// Close 关闭数据库连接
func Close() error {
	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// AutoMigrate 自动迁移数据库表
func AutoMigrate(models ...interface{}) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	return db.AutoMigrate(models...)
}

// BaseModel 基础模型，包含通用字段
// 所有业务表都需要继承此模型，并添加region_id字段
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// RegionBaseModel 带地区隔离的基础模型
// 所有业务表都应该使用此模型，实现地区数据隔离
type RegionBaseModel struct {
	BaseModel
	RegionID uint `gorm:"index;not null;default:1" json:"region_id"` // 地区ID，用于数据隔离
}

// TableName 表名接口
type TableNamer interface {
	TableName() string
}
