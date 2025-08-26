package middleware

import (
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const (
	SessionUserKey = "user_id"
	SessionTimeKey = "login_time"
)

// AuthMiddleware 身份验证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		
		// 检查会话中是否有用户信息
		userID := session.Get(SessionUserKey)
		if userID == nil {
			// 如果是API请求，返回JSON错误
			if isAPIRequest(c) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"error":   "未授权访问，请先登录",
				})
				c.Abort()
				return
			}
			
			// 如果是页面请求，重定向到登录页
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		
		// 检查会话是否过期
		loginTime := session.Get(SessionTimeKey)
		if loginTime != nil {
			if loginTimeVal, ok := loginTime.(int64); ok {
				// 检查会话是否超过24小时（可配置）
				if time.Now().Unix()-loginTimeVal > 24*60*60 {
					// 清除过期会话
					session.Clear()
					session.Save()
					
					if isAPIRequest(c) {
						c.JSON(http.StatusUnauthorized, gin.H{
							"success": false,
							"error":   "会话已过期，请重新登录",
						})
						c.Abort()
						return
					}
					
					c.Redirect(http.StatusFound, "/login")
					c.Abort()
					return
				}
			}
		}
		
		// 更新会话活动时间
		session.Set(SessionTimeKey, time.Now().Unix())
		session.Save()
		
		// 将用户信息添加到上下文中
		c.Set("user_id", userID)
		c.Next()
	}
}

// isAPIRequest 判断是否为API请求
func isAPIRequest(c *gin.Context) bool {
	path := c.Request.URL.Path
	return len(path) >= 4 && path[:4] == "/api"
}

// RequireAuth 需要认证的路由组中间件
func RequireAuth() gin.HandlerFunc {
	return AuthMiddleware()
}

// GetCurrentUser 获取当前登录用户
func GetCurrentUser(c *gin.Context) (string, bool) {
	if userID, exists := c.Get("user_id"); exists {
		if userIDStr, ok := userID.(string); ok {
			return userIDStr, true
		}
	}
	return "", false
}

// SetUserSession 设置用户会话
func SetUserSession(c *gin.Context, userID string) error {
	session := sessions.Default(c)
	session.Set(SessionUserKey, userID)
	session.Set(SessionTimeKey, time.Now().Unix())
	return session.Save()
}

// ClearUserSession 清除用户会话
func ClearUserSession(c *gin.Context) error {
	session := sessions.Default(c)
	session.Clear()
	return session.Save()
}
