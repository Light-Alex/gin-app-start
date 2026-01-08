package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type PostgresConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	DBName       string
	SSLMode      string
	MaxIdleConns int
	MaxOpenConns int
	MaxLifetime  int
	LogLevel     string
}

type Repo interface {
	i()
	GetDB() *gorm.DB
	DbClose() error
}

type dbRepo struct {
	DB *gorm.DB
}

// var DB *gorm.DB
var DBRepo Repo

func NewPostgresDB(config *PostgresConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Asia/Shanghai",
		config.Host, config.User, config.Password, config.DBName, config.Port, config.SSLMode)

	var logLevel logger.LogLevel
	switch config.LogLevel {
	case "silent":
		logLevel = logger.Silent
	case "error":
		logLevel = logger.Error
	case "warn":
		logLevel = logger.Warn
	case "info":
		logLevel = logger.Info
	default:
		logLevel = logger.Info
	}

	// 初始化数据库连接
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取SQL数据库连接实例
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// 设置数据库连接池参数
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	if config.MaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(config.MaxLifetime) * time.Second)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	DBRepo = &dbRepo{DB: db}

	// 使用插件
	db.Use(&TracePlugin{})

	return db, nil
}

func (d *dbRepo) i() {}

func (d *dbRepo) GetDB() *gorm.DB {
	return d.DB
}

func (d *dbRepo) DbClose() error {
	if d.DB != nil {
		sqlDB, err := d.DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
