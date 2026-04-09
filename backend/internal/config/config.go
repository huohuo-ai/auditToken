package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	ClickHouse ClickHouseConfig `mapstructure:"clickhouse"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Audit    AuditConfig    `mapstructure:"audit"`
}

type ServerConfig struct {
	Port         string `mapstructure:"port"`
	Mode         string `mapstructure:"mode"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

type DatabaseConfig struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type ClickHouseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type JWTConfig struct {
	Secret    string `mapstructure:"secret"`
	ExpiresIn int    `mapstructure:"expires_in"`
}

type AuditConfig struct {
	HighRiskPromptPatterns []string `mapstructure:"high_risk_prompt_patterns"`
	OffHoursStart          int      `mapstructure:"off_hours_start"`
	OffHoursEnd            int      `mapstructure:"off_hours_end"`
	TokenThresholdHourly   int64    `mapstructure:"token_threshold_hourly"`
	SuspiciousIPList       []string `mapstructure:"suspicious_ip_list"`
}

var GlobalConfig *Config

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	// 设置默认值
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("database.driver", "mysql")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("clickhouse.host", "localhost")
	viper.SetDefault("clickhouse.port", 8123)
	viper.SetDefault("jwt.expires_in", 86400)
	viper.SetDefault("audit.off_hours_start", 22)
	viper.SetDefault("audit.off_hours_end", 6)
	viper.SetDefault("audit.token_threshold_hourly", 100000)

	var config Config
	configLoadErr := viper.ReadInConfig()
	if configLoadErr != nil {
		// 配置文件读取失败，使用默认值继续
		configLoadErr = fmt.Errorf("failed to read config file: %w", configLoadErr)
	}

	// Unmarshal 配置（包括默认值）
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	GlobalConfig = &config
	return &config, configLoadErr
}

func GetConfig() *Config {
	if GlobalConfig == nil {
		panic("config not loaded")
	}
	return GlobalConfig
}
