package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Language LanguageConfig `mapstructure:"language"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Log      LogConfig      `mapstructure:"log"`
	File     FileConfig     `mapstructure:"file"`
	Session  SessionConfig  `mapstructure:"session"`
}

type ServerConfig struct {
	Port         int    `mapstructure:"port"`
	Mode         string `mapstructure:"mode"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	LimitNum     int    `mapstructure:"limit_num"`
}

type LanguageConfig struct {
	Local string `mapstructure:"local"`
}

type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"dbname"`
	SSLMode      string `mapstructure:"sslmode"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxLifetime  int    `mapstructure:"max_lifetime"`
	LogLevel     string `mapstructure:"log_level"`
	AutoMigrate  bool   `mapstructure:"auto_migrate"`
}

type RedisConfig struct {
	Addr         string `mapstructure:"addr"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
	MaxRetries   int    `mapstructure:"max_retries"`
}

type LogConfig struct {
	Level    string `mapstructure:"level"`
	FilePath string `mapstructure:"file_path"`
	MaxSize  int    `mapstructure:"max_size"`
	MaxAge   int    `mapstructure:"max_age"`
}

type FileConfig struct {
	DirName   string `mapstructure:"dir_name"`
	UrlPrefix string `mapstructure:"url_prefix"`
	MaxSize   int64  `mapstructure:"max_size"`
}

type SessionConfig struct {
	UseRedis bool   `mapstructure:"use_redis"`
	Name     string `mapstructure:"name"`
	Size     int    `mapstructure:"size"`
	Key      string `mapstructure:"key"`
	MaxAge   int    `mapstructure:"max_age"`
	Path     string `mapstructure:"path"`
	Domain   string `mapstructure:"domain"`
	HttpOnly bool   `mapstructure:"http_only"`
	Secure   bool   `mapstructure:"secure"`
}

var GlobalConfig *Config

func Load() (*Config, error) {
	env := os.Getenv("SERVER_ENV")
	if env == "" {
		env = "local"
	}

	viper.SetConfigName(fmt.Sprintf("config.%s", env))
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	// 自动检查环境变量是否与现有的配置键匹配
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	GlobalConfig = &config
	return &config, nil
}

func GetConfig() *Config {
	return GlobalConfig
}
