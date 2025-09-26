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
	Auth     AuthConfig     `yaml:"auth"`
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

// AuthConfig 身份验证配置
type AuthConfig struct {
	Admin AdminConfig `yaml:"admin"`
}

// AdminConfig 管理员账号配置
type AdminConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
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

	// 验证必要的配置
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %v", err)
	}

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

// validateConfig 验证配置的完整性和有效性
func validateConfig(config *Config) error {
	// 验证身份验证配置
	if config.Auth.Admin.Username == "" {
		return fmt.Errorf("身份验证配置错误: 管理员用户名不能为空 (auth.admin.username)")
	}
	if config.Auth.Admin.Password == "" {
		return fmt.Errorf("身份验证配置错误: 管理员密码不能为空 (auth.admin.password)")
	}

	// 验证密码强度（可选，但建议）
	if len(config.Auth.Admin.Password) < 3 {
		return fmt.Errorf("身份验证配置错误: 管理员密码长度至少为3个字符")
	}

	// 验证数据库配置
	if config.Database.Name == "" {
		return fmt.Errorf("数据库配置错误: 数据库名称不能为空 (database.name)")
	}
	if config.Database.Username == "" {
		return fmt.Errorf("数据库配置错误: 数据库用户名不能为空 (database.username)")
	}

	return nil
}
