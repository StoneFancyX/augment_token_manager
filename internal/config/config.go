package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 应用程序配置结构
type Config struct {
	Database DatabaseConfig `yaml:"database"`
	Server   ServerConfig   `yaml:"server"`
	Logging  LoggingConfig  `yaml:"logging"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string         `yaml:"host"`
	Port     int            `yaml:"port"`
	Name     string         `yaml:"name"`
	Username string         `yaml:"username"`
	Password string         `yaml:"password"`
	SSLMode  string         `yaml:"sslmode"`
	Pool     PoolConfig     `yaml:"pool"`
}

// PoolConfig 连接池配置
type PoolConfig struct {
	MaxIdleConns    int `yaml:"max_idle_conns"`
	MaxOpenConns    int `yaml:"max_open_conns"`
	ConnMaxLifetime int `yaml:"conn_max_lifetime"` // 分钟
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
	Mode string `yaml:"mode"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// LoadConfig 从配置文件加载配置
func LoadConfig(configPath string) (*Config, error) {
	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析 YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 设置默认值
	setDefaults(&config)

	return &config, nil
}

// setDefaults 设置默认配置值
func setDefaults(config *Config) {
	// 数据库默认值
	if config.Database.Host == "" {
		config.Database.Host = "localhost"
	}
	if config.Database.Port == 0 {
		config.Database.Port = 5432
	}
	if config.Database.SSLMode == "" {
		config.Database.SSLMode = "disable"
	}
	if config.Database.Pool.MaxIdleConns == 0 {
		config.Database.Pool.MaxIdleConns = 10
	}
	if config.Database.Pool.MaxOpenConns == 0 {
		config.Database.Pool.MaxOpenConns = 100
	}
	if config.Database.Pool.ConnMaxLifetime == 0 {
		config.Database.Pool.ConnMaxLifetime = 60
	}

	// 服务器默认值
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}
	if config.Server.Mode == "" {
		config.Server.Mode = "debug"
	}

	// 日志默认值
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
	if config.Logging.Format == "" {
		config.Logging.Format = "text"
	}
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.Username, c.Password, c.Name, c.SSLMode)
}

// GetConnMaxLifetime 获取连接最大生存时间
func (c *PoolConfig) GetConnMaxLifetime() time.Duration {
	return time.Duration(c.ConnMaxLifetime) * time.Minute
}
