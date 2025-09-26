package database

import (
	"augment_token_manager/internal/config"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// DB 全局数据库连接
var DB *sql.DB

// Connect 连接到 PostgreSQL 数据库
func Connect(dbConfig config.DatabaseConfig) error {
	// 使用配置构建连接字符串
	connStr := dbConfig.GetDSN()

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("无法打开数据库连接: %v", err)
	}

	// 配置连接池
	DB.SetMaxIdleConns(dbConfig.Pool.MaxIdleConns)
	DB.SetMaxOpenConns(dbConfig.Pool.MaxOpenConns)
	DB.SetConnMaxLifetime(dbConfig.Pool.GetConnMaxLifetime())

	// 测试连接
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("无法连接到数据库: %v", err)
	}

	log.Printf("成功连接到 PostgreSQL 数据库 (%s:%d/%s)",
		dbConfig.Host, dbConfig.Port, dbConfig.Name)
	log.Printf("连接池配置: MaxIdle=%d, MaxOpen=%d, MaxLifetime=%v",
		dbConfig.Pool.MaxIdleConns, dbConfig.Pool.MaxOpenConns, dbConfig.Pool.GetConnMaxLifetime())
	return nil
}

// Close 关闭数据库连接
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// InitTables 初始化数据库表
func InitTables() error {
	// 首先检查表是否存在
	checkTableSQL := `
	SELECT EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_name = 'tokens'
	);`

	var exists bool
	err := DB.QueryRow(checkTableSQL).Scan(&exists)
	if err != nil {
		return fmt.Errorf("检查表是否存在失败: %v", err)
	}

	if exists {
		log.Println("tokens 表已存在，检查表结构...")

		// 检查表结构
		columnsSQL := `
		SELECT column_name, data_type
		FROM information_schema.columns
		WHERE table_name = 'tokens'
		ORDER BY ordinal_position;`

		rows, err := DB.Query(columnsSQL)
		if err != nil {
			return fmt.Errorf("查询表结构失败: %v", err)
		}
		defer rows.Close()

		log.Println("当前 tokens 表结构:")
		for rows.Next() {
			var columnName, dataType string
			if err := rows.Scan(&columnName, &dataType); err != nil {
				return fmt.Errorf("扫描列信息失败: %v", err)
			}
			log.Printf("  - %s: %s", columnName, dataType)
		}
	} else {
		log.Println("tokens 表不存在，正在创建...")

		// 创建 tokens 表
		createTableSQL := `
		CREATE TABLE tokens (
			id VARCHAR(255) PRIMARY KEY,
			tenant_url TEXT,
			access_token TEXT,
			portal_url TEXT,
			email_note TEXT,
			ban_status JSONB DEFAULT '{}',
			portal_info JSONB DEFAULT '{}',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`

		_, err = DB.Exec(createTableSQL)
		if err != nil {
			return fmt.Errorf("创建 tokens 表失败: %v", err)
		}

		log.Println("tokens 表创建成功")

		// 创建索引以提高查询性能
		createIndexSQL := `
		CREATE INDEX IF NOT EXISTS idx_tokens_created_at ON tokens(created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_tokens_updated_at ON tokens(updated_at DESC);`

		_, err = DB.Exec(createIndexSQL)
		if err != nil {
			log.Printf("创建索引时出现警告: %v", err)
			// 索引创建失败不是致命错误，继续执行
		} else {
			log.Println("数据库索引创建成功")
		}
	}

	log.Println("数据库表初始化完成")
	return nil
}
