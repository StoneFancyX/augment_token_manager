package handlers

import (
	"augment_token_manager/internal/config"
	"augment_token_manager/internal/middleware"
	"augment_token_manager/internal/repository"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	ClientID    = "v" // 实际使用时替换为你的 client_id
	AuthBaseURL = "https://auth.augmentcode.com"
)

// AuthHandler 授权处理器
type AuthHandler struct {
	tokenRepo     *repository.TokenRepository
	loginAttempts map[string]*LoginAttempt // 简单的内存存储，生产环境建议使用Redis
	config        *config.Config           // 配置对象
}

// NewAuthHandler 创建新的 AuthHandler 实例
func NewAuthHandler(cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		tokenRepo:     repository.NewTokenRepository(),
		loginAttempts: make(map[string]*LoginAttempt),
		config:        cfg,
	}
}

// AugmentOAuthState 保存 OAuth 所需的随机信息
type AugmentOAuthState struct {
	CodeVerifier  string `json:"code_verifier"`
	CodeChallenge string `json:"code_challenge"`
	State         string `json:"state"`
	CreationTime  int64  `json:"creation_time"`
}

// AuthResponse 授权响应结构
type AuthResponse struct {
	Code      string `json:"code" binding:"required"`
	State     string `json:"state" binding:"required"`
	TenantURL string `json:"tenant_url" binding:"required"`
}

// TokenApiResponse 对应服务端返回的 access_token
type TokenApiResponse struct {
	AccessToken string `json:"access_token"`
	TenantURL   string `json:"tenant_url,omitempty"`
	Email       string `json:"email,omitempty"`
	PortalURL   string `json:"portal_url,omitempty"`
}

// LoginRequest 登录请求结构
type LoginRequest struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	RememberMe bool   `json:"remember_me"`
}

// LoginAttempt 登录尝试记录
type LoginAttempt struct {
	Username    string
	Attempts    int
	LastAttempt time.Time
	LockedUntil time.Time
}

// generateRandomBytes 生成随机字节
func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return b, err
}

// base64URLEncode Base64 URL 安全编码（不带 padding）
func base64URLEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// createAugmentOAuthState 创建 OAuth 状态 (包含 code_verifier / code_challenge / state)
func createAugmentOAuthState() (*AugmentOAuthState, error) {
	// code_verifier
	verifierBytes, err := generateRandomBytes(32)
	if err != nil {
		return nil, err
	}
	codeVerifier := base64URLEncode(verifierBytes)

	// code_challenge = BASE64URL(SHA256(code_verifier))
	hash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64URLEncode(hash[:])

	// state
	stateBytes, err := generateRandomBytes(8)
	if err != nil {
		return nil, err
	}
	state := base64URLEncode(stateBytes)

	return &AugmentOAuthState{
		CodeVerifier:  codeVerifier,
		CodeChallenge: codeChallenge,
		State:         state,
		CreationTime:  time.Now().UnixMilli(),
	}, nil
}

// generateAugmentAuthorizeURL 生成 OAuth 授权 URL
func generateAugmentAuthorizeURL(oauthState *AugmentOAuthState) (string, error) {
	u, err := url.Parse(AuthBaseURL + "/authorize")
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("response_type", "code")
	q.Set("code_challenge", oauthState.CodeChallenge)
	q.Set("client_id", ClientID)
	q.Set("state", oauthState.State)
	q.Set("prompt", "login")
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// GenerateAuthURLAPI 生成授权URL API
func (h *AuthHandler) GenerateAuthURLAPI(c *gin.Context) {
	// 创建OAuth状态
	oauthState, err := createAugmentOAuthState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "生成OAuth状态失败: " + err.Error(),
		})
		return
	}

	// 生成授权URL
	authURL, err := generateAugmentAuthorizeURL(oauthState)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "生成授权URL失败: " + err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"auth_url":       authURL,
			"code_verifier":  oauthState.CodeVerifier,
			"code_challenge": oauthState.CodeChallenge,
			"state":          oauthState.State,
			"creation_time":  oauthState.CreationTime,
		},
		"message": "授权URL生成成功",
	})
}

// getAugmentAccessToken 使用授权码换取访问令牌
func (h *AuthHandler) getAugmentAccessToken(tenantURL, codeVerifier, code string) (*TokenApiResponse, error) {
	data := map[string]string{
		"grant_type":    "authorization_code",
		"client_id":     ClientID,
		"code_verifier": codeVerifier,
		"redirect_uri":  "", // 如果服务端要求 redirect_uri，这里要保持一致
		"code":          code,
	}

	body, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("序列化请求数据失败: %v", err)
	}

	tokenURL := tenantURL + "token"

	resp, err := http.Post(tokenURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("请求token端点失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token请求失败: %s", resp.Status)
	}

	var tokenResp TokenApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("解析token响应失败: %v", err)
	}

	return &tokenResp, nil
}


// ValidateAuthResponseRequest 验证授权响应请求结构
type ValidateAuthResponseRequest struct {
	AuthResponse AuthResponse      `json:"auth_response" binding:"required"`
	OAuthState   AugmentOAuthState `json:"oauth_state" binding:"required"`
}

// ValidateAuthResponseAPI 验证授权响应API（第2步）
func (h *AuthHandler) ValidateAuthResponseAPI(c *gin.Context) {
	var req ValidateAuthResponseRequest

	// 绑定JSON数据
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求数据格式错误: " + err.Error(),
		})
		return
	}

	// 验证state参数以防止CSRF攻击
	if req.AuthResponse.State != req.OAuthState.State {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "state参数不匹配，可能存在安全风险",
		})
		return
	}

	// 验证OAuth状态的时效性（例如：30分钟内有效）
	currentTime := time.Now().UnixMilli()
	if currentTime-req.OAuthState.CreationTime > 30*60*1000 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "OAuth状态已过期，请重新开始授权流程",
		})
		return
	}

	// 使用授权码换取access token（但不保存）
	tokenResp, err := h.getAugmentAccessToken(req.AuthResponse.TenantURL, req.OAuthState.CodeVerifier, req.AuthResponse.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取access token失败: " + err.Error(),
		})
		return
	}

	// 返回验证成功的响应，包含解析后的token信息（但不保存到数据库）
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"tenant_url":   req.AuthResponse.TenantURL,
			"access_token": tokenResp.AccessToken,
			"email":        tokenResp.Email,
			"portal_url":   tokenResp.PortalURL,
		},
		"message": "授权响应验证成功",
	})
}

// SaveTokenRequest 保存token请求结构（第3步）
type SaveTokenRequest struct {
	TenantURL   string `json:"tenant_url" binding:"required"`
	AccessToken string `json:"access_token" binding:"required"`
	Email       string `json:"email"`
	PortalURL   string `json:"portal_url"`
	EmailNote   string `json:"email_note"`
}

// SaveTokenAPI 保存token API（第3步）
func (h *AuthHandler) SaveTokenAPI(c *gin.Context) {
	var req SaveTokenRequest

	// 绑定JSON数据
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求数据格式错误: " + err.Error(),
		})
		return
	}

	// 验证必填字段
	if req.TenantURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Tenant URL 不能为空",
		})
		return
	}

	if req.AccessToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Access Token 不能为空",
		})
		return
	}

	// 准备保存到数据库的数据
	createReq := repository.CreateTokenRequest{
		TenantURL:   req.TenantURL,
		AccessToken: req.AccessToken,
		PortalURL:   req.PortalURL,
		EmailNote:   req.EmailNote,
	}

	// 如果用户没有输入邮箱备注，使用从token响应中获取的email
	if createReq.EmailNote == "" && req.Email != "" {
		createReq.EmailNote = req.Email
	}

	// 保存token到数据库
	token, err := h.tokenRepo.CreateToken(createReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "保存Token失败: " + err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    token.ToResponse(),
		"message": "Token保存成功",
	})
}

// GetLoginPage 获取登录页面
func (h *AuthHandler) GetLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "登录 - Augment Token Manager",
	})
}

// LoginAPI 登录API
func (h *AuthHandler) LoginAPI(c *gin.Context) {
	var req LoginRequest

	// 绑定JSON数据
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求数据格式错误: " + err.Error(),
		})
		return
	}

	// 检查账号是否被锁定
	if h.isAccountLocked(req.Username) {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"success": false,
			"error":   "账号已被锁定，请稍后再试",
		})
		return
	}

	// 验证用户凭据
	if !h.validateCredentials(req.Username, req.Password) {
		// 记录失败尝试
		h.recordFailedAttempt(req.Username)

		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "用户名或密码错误",
		})
		return
	}

	// 清除失败尝试记录
	h.clearFailedAttempts(req.Username)

	// 设置用户会话
	if err := middleware.SetUserSession(c, req.Username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "设置会话失败",
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "登录成功",
		"data": gin.H{
			"username": req.Username,
		},
	})
}

// LogoutAPI 登出API
func (h *AuthHandler) LogoutAPI(c *gin.Context) {
	// 清除用户会话
	if err := middleware.ClearUserSession(c); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "登出失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "登出成功",
	})
}

// validateCredentials 验证用户凭据
func (h *AuthHandler) validateCredentials(username, password string) bool {
	// 从配置文件读取管理员账号信息
	adminUsername := h.config.Auth.Admin.Username
	adminPassword := h.config.Auth.Admin.Password

	// 验证管理员账号
	if username == adminUsername && password == adminPassword {
		return true
	}

	// 可以扩展支持数据库中的其他用户
	return false
}

// isAccountLocked 检查账号是否被锁定
func (h *AuthHandler) isAccountLocked(username string) bool {
	attempt, exists := h.loginAttempts[username]
	if !exists {
		return false
	}

	maxAttempts := 5 // 简化处理：硬编码最大尝试次数

	// 检查是否超过最大尝试次数
	if attempt.Attempts >= maxAttempts {
		// 检查锁定时间是否已过
		if time.Now().Before(attempt.LockedUntil) {
			return true
		}
		// 锁定时间已过，重置尝试次数
		delete(h.loginAttempts, username)
	}

	return false
}

// recordFailedAttempt 记录失败尝试
func (h *AuthHandler) recordFailedAttempt(username string) {
	now := time.Now()

	attempt, exists := h.loginAttempts[username]
	if !exists {
		attempt = &LoginAttempt{
			Username: username,
		}
		h.loginAttempts[username] = attempt
	}

	attempt.Attempts++
	attempt.LastAttempt = now

	maxAttempts := 5 // 简化处理：硬编码最大尝试次数

	// 如果达到最大尝试次数，设置锁定时间
	if attempt.Attempts >= maxAttempts {
		lockoutDuration := 15 * time.Minute // 简化处理：硬编码锁定时间
		attempt.LockedUntil = now.Add(lockoutDuration)
	}
}

// clearFailedAttempts 清除失败尝试记录
func (h *AuthHandler) clearFailedAttempts(username string) {
	delete(h.loginAttempts, username)
}
