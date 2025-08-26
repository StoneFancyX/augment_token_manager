package main

import (
	"augment_token_manager/internal/config"
	"augment_token_manager/internal/database"
	"augment_token_manager/internal/handlers"
	"augment_token_manager/internal/middleware"
	"augment_token_manager/internal/utils"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

// 版本信息变量（通过编译时的 -ldflags 设置）
var (
	Version    = "dev"
	BuildTime  = "unknown"
	CommitHash = "unknown"
)

func main() {
	// 加载配置文件
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 初始化日志系统
	utils.InitLogger(cfg)

	// 设置 Gin 运行模式
	gin.SetMode(cfg.Server.Mode)

	// 连接数据库
	if err := database.Connect(cfg.Database); err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer database.Close()

	// 初始化数据库表
	if err := database.InitTables(); err != nil {
		log.Fatalf("数据库表初始化失败: %v", err)
	}

	// 创建 Gin 路由器
	router := gin.Default()

	// 配置session中间件
	store := cookie.NewStore([]byte("augment-token-manager-secret-key-2025"))
	store.Options(sessions.Options{
		MaxAge:   24 * 60 * 60, // 24小时
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // 开发环境设为false，生产环境应设为true
	})
	router.Use(sessions.Sessions("atm_session", store))

	// 加载 HTML 模板
	router.LoadHTMLGlob("web/templates/*")

	// 静态文件服务
	router.Static("/static", "./web/static")

	// 创建处理器
	tokenHandler := handlers.NewTokenHandler()
	authHandler := handlers.NewAuthHandler(cfg)

	// 公开路由（不需要认证）
	router.GET("/login", authHandler.GetLoginPage)
	router.POST("/api/auth/login", authHandler.LoginAPI)

	// 受保护的路由（需要认证）
	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		// 主页面
		protected.GET("/", tokenHandler.GetTokensPage)

		// Token管理API
		protected.GET("/api/tokens", tokenHandler.GetTokensAPI)
		protected.POST("/api/tokens", tokenHandler.CreateTokenAPI)
		protected.POST("/api/tokens/batch-import", tokenHandler.BatchImportTokensAPI)
		protected.GET("/api/tokens/:id", tokenHandler.GetTokenByIDAPI)
		protected.PUT("/api/tokens/:id", tokenHandler.UpdateTokenAPI)
		protected.DELETE("/api/tokens/:id", tokenHandler.DeleteTokenAPI)
		protected.POST("/api/tokens/:id/refresh", tokenHandler.RefreshTokenAPI)
		protected.POST("/api/tokens/batch-refresh", tokenHandler.BatchRefreshTokensAPI)

		// OAuth相关API
		protected.GET("/api/auth/generate-url", authHandler.GenerateAuthURLAPI)
		protected.POST("/api/auth/exchange-token", authHandler.ExchangeTokenAPI)
		protected.POST("/api/auth/logout", authHandler.LogoutAPI)
	}

	// 健康检查端点
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "ok",
			"message":     "Augment Token Manager is running",
			"version":     Version,
			"build_time":  BuildTime,
			"commit_hash": CommitHash,
		})
	})

	// 启动服务器
	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Println("启动 Augment Token Manager 服务器...")
	log.Printf("版本: %s (构建时间: %s, 提交: %s)", Version, BuildTime, CommitHash)
	log.Printf("访问地址: http://localhost:%d", cfg.Server.Port)
	log.Printf("服务器监听地址: %s", serverAddr)

	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
