package utils

import (
	"augment_token_manager/internal/config"
	"log"
	"sync"
)

// Logger 日志工具结构
type Logger struct {
	config *config.Config
	mu     sync.RWMutex
}

// 全局日志实例
var (
	globalLogger *Logger
	once         sync.Once
)

// InitLogger 初始化全局日志实例
func InitLogger(cfg *config.Config) {
	once.Do(func() {
		globalLogger = &Logger{
			config: cfg,
		}
	})
}

// GetLogger 获取全局日志实例
func GetLogger() *Logger {
	if globalLogger == nil {
		// 如果没有初始化，创建一个默认的logger
		defaultConfig := &config.Config{
			Server: config.ServerConfig{
				Mode: "release", // 默认为release模式
			},
		}
		globalLogger = &Logger{
			config: defaultConfig,
		}
	}
	return globalLogger
}

// UpdateConfig 更新日志配置
func (l *Logger) UpdateConfig(cfg *config.Config) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.config = cfg
}

// isDebugMode 检查是否为debug模式
func (l *Logger) isDebugMode() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.config.Server.Mode == "debug"
}

// Debug 输出DEBUG级别日志（只在debug模式下输出）
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.isDebugMode() {
		log.Printf("[DEBUG] "+format, args...)
	}
}

// Info 输出INFO级别日志（始终输出）
func (l *Logger) Info(format string, args ...interface{}) {
	log.Printf("[INFO] "+format, args...)
}

// Warn 输出WARN级别日志（始终输出）
func (l *Logger) Warn(format string, args ...interface{}) {
	log.Printf("[WARN] "+format, args...)
}

// Error 输出ERROR级别日志（始终输出）
func (l *Logger) Error(format string, args ...interface{}) {
	log.Printf("[ERROR] "+format, args...)
}

// Fatal 输出FATAL级别日志并退出程序（始终输出）
func (l *Logger) Fatal(format string, args ...interface{}) {
	log.Fatalf("[FATAL] "+format, args...)
}

// 全局便捷函数

// Debug 全局DEBUG日志函数
func Debug(format string, args ...interface{}) {
	GetLogger().Debug(format, args...)
}

// Info 全局INFO日志函数
func Info(format string, args ...interface{}) {
	GetLogger().Info(format, args...)
}

// Warn 全局WARN日志函数
func Warn(format string, args ...interface{}) {
	GetLogger().Warn(format, args...)
}

// Error 全局ERROR日志函数
func Error(format string, args ...interface{}) {
	GetLogger().Error(format, args...)
}

// Fatal 全局FATAL日志函数
func Fatal(format string, args ...interface{}) {
	GetLogger().Fatal(format, args...)
}
