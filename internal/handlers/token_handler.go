package handlers

import (
	"augment_token_manager/internal/repository"
	"augment_token_manager/internal/services"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
)

// TokenHandler Token 处理器
type TokenHandler struct {
	tokenRepo      *repository.TokenRepository
	refreshService *services.TokenRefreshService
}

// NewTokenHandler 创建新的 TokenHandler 实例
func NewTokenHandler() *TokenHandler {
	return &TokenHandler{
		tokenRepo:      repository.NewTokenRepository(),
		refreshService: services.NewTokenRefreshService(),
	}
}

// GetTokensPage 获取 Token 管理页面
func (h *TokenHandler) GetTokensPage(c *gin.Context) {
	tokens, err := h.tokenRepo.GetAllTokens()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "获取 Token 列表失败: " + err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"tokens": tokens,
		"title":  "Augment Token Manager",
	})
}

// GetTokensAPI 获取 Token 列表 API（支持分页）
func (h *TokenHandler) GetTokensAPI(c *gin.Context) {
	// 解析分页参数
	var params repository.PaginationParams

	// 获取页码参数
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			params.Page = page
		}
	}

	// 获取每页记录数参数
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			params.Limit = limit
		}
	}

	// 设置默认分页参数
	if params.Page == 0 {
		params.Page = 1
	}
	if params.Limit == 0 {
		params.Limit = 10
	}

	// 使用分页查询
	result, err := h.tokenRepo.GetTokensWithPagination(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取分页 Token 列表失败: " + err.Error(),
		})
		return
	}

	// 转换为响应格式
	var tokenResponses []interface{}
	for _, token := range result.Data {
		tokenResponses = append(tokenResponses, token.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tokenResponses,
		"pagination": gin.H{
			"total":       result.Total,
			"page":        result.Page,
			"limit":       result.Limit,
			"total_pages": result.TotalPages,
			"has_next":    result.HasNext,
			"has_prev":    result.HasPrev,
		},
	})
}

// GetTokenByIDAPI 根据 ID 获取单个 Token API
func (h *TokenHandler) GetTokenByIDAPI(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Token ID 不能为空",
		})
		return
	}

	token, err := h.tokenRepo.GetTokenByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Token 不存在: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    token.ToResponse(),
	})
}

// RefreshTokenAPI 刷新单个 Token 信息 API
func (h *TokenHandler) RefreshTokenAPI(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Token ID 不能为空",
		})
		return
	}

	// 调用刷新服务
	token, err := h.refreshService.RefreshTokenInfo(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "刷新 Token 失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    token.ToResponse(),
		"message": "Token 信息已刷新",
	})
}

// BatchRefreshTokensAPI 批量刷新所有 Token 信息 API
func (h *TokenHandler) BatchRefreshTokensAPI(c *gin.Context) {
	// 获取所有 Token
	tokens, err := h.tokenRepo.GetAllTokens()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取 Token 列表失败: " + err.Error(),
		})
		return
	}

	if len(tokens) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "没有 Token 需要刷新",
			"data": gin.H{
				"total":   0,
				"success": 0,
				"failed":  0,
			},
		})
		return
	}

	// 批量刷新
	var successCount, failedCount int
	var refreshedTokens []interface{}
	var errors []string

	for _, token := range tokens {
		refreshedToken, err := h.refreshService.RefreshTokenInfo(token.ID)
		if err != nil {
			failedCount++
			errors = append(errors, fmt.Sprintf("Token %s: %s", token.ID, err.Error()))
		} else {
			successCount++
			refreshedTokens = append(refreshedTokens, refreshedToken.ToResponse())
		}
	}

	// 返回结果
	result := gin.H{
		"success": true,
		"data": gin.H{
			"total":   len(tokens),
			"success": successCount,
			"failed":  failedCount,
			"tokens":  refreshedTokens,
		},
	}

	if failedCount > 0 {
		result["errors"] = errors
		if successCount == 0 {
			result["message"] = fmt.Sprintf("批量刷新失败：所有 %d 个 Token 都刷新失败", failedCount)
		} else {
			result["message"] = fmt.Sprintf("批量刷新完成：%d 个成功，%d 个失败", successCount, failedCount)
		}
	} else {
		result["message"] = fmt.Sprintf("批量刷新成功：所有 %d 个 Token 都已刷新", successCount)
	}

	c.JSON(http.StatusOK, result)
}

// CreateTokenAPI 创建新Token API
func (h *TokenHandler) CreateTokenAPI(c *gin.Context) {
	var req repository.CreateTokenRequest

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

	// 验证URL格式
	if err := validateURL(req.TenantURL); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Tenant URL " + err.Error(),
		})
		return
	}

	if err := validateURL(req.PortalURL); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Portal URL " + err.Error(),
		})
		return
	}

	// 创建Token
	token, err := h.tokenRepo.CreateToken(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "创建 Token 失败: " + err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    token.ToResponse(),
		"message": "Token 创建成功",
	})
}

// DeleteTokenAPI 删除Token API
func (h *TokenHandler) DeleteTokenAPI(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Token ID 不能为空",
		})
		return
	}

	// 删除Token
	err := h.tokenRepo.DeleteToken(id)
	if err != nil {
		// 根据错误类型返回不同的状态码
		if err.Error() == "Token 不存在" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "删除 Token 失败: " + err.Error(),
			})
		}
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Token 删除成功",
	})
}

// BatchImportTokensAPI 批量导入 Token API
func (h *TokenHandler) BatchImportTokensAPI(c *gin.Context) {
	// 定义批量导入请求结构
	type BatchImportRequest struct {
		Tokens []repository.CreateTokenRequest `json:"tokens"`
	}

	var req BatchImportRequest

	// 绑定JSON数据
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求数据格式错误: " + err.Error(),
		})
		return
	}

	if len(req.Tokens) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "没有提供要导入的Token数据",
		})
		return
	}

	// 批量创建Token
	successful := 0
	failed := 0
	importErrors := []string{}

	for i, tokenReq := range req.Tokens {
		// 验证必填字段
		if tokenReq.TenantURL == "" {
			failed++
			importErrors = append(importErrors, fmt.Sprintf("第 %d 条: Tenant URL 不能为空", i+1))
			continue
		}

		if tokenReq.AccessToken == "" {
			failed++
			importErrors = append(importErrors, fmt.Sprintf("第 %d 条: Access Token 不能为空", i+1))
			continue
		}

		// 验证URL格式
		if err := validateURL(tokenReq.TenantURL); err != nil {
			failed++
			importErrors = append(importErrors, fmt.Sprintf("第 %d 条: Tenant URL %s", i+1, err.Error()))
			continue
		}

		if tokenReq.PortalURL != "" {
			if err := validateURL(tokenReq.PortalURL); err != nil {
				failed++
				importErrors = append(importErrors, fmt.Sprintf("第 %d 条: Portal URL %s", i+1, err.Error()))
				continue
			}
		}

		// 创建Token
		_, err := h.tokenRepo.CreateToken(tokenReq)
		if err != nil {
			failed++
			importErrors = append(importErrors, fmt.Sprintf("第 %d 条: %s", i+1, err.Error()))
		} else {
			successful++
		}
	}

	// 返回导入结果
	result := gin.H{
		"success":    true,
		"total":      len(req.Tokens),
		"successful": successful,
		"failed":     failed,
		"message":    fmt.Sprintf("批量导入完成，成功 %d 条，失败 %d 条", successful, failed),
	}

	if len(importErrors) > 0 {
		result["errors"] = importErrors
	}

	c.JSON(http.StatusOK, result)
}



// UpdateTokenAPI 更新Token API
func (h *TokenHandler) UpdateTokenAPI(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Token ID 不能为空",
		})
		return
	}

	var req repository.UpdateTokenRequest

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

	// 验证URL格式
	if err := validateURL(req.TenantURL); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Tenant URL " + err.Error(),
		})
		return
	}

	if err := validateURL(req.PortalURL); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Portal URL " + err.Error(),
		})
		return
	}

	// 更新Token
	token, err := h.tokenRepo.UpdateToken(id, req)
	if err != nil {
		// 根据错误类型返回不同的状态码
		if err.Error() == "Token 不存在" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "更新 Token 失败: " + err.Error(),
			})
		}
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    token.ToResponse(),
		"message": "Token 更新成功",
	})
}



// validateURL 验证URL格式
func validateURL(urlStr string) error {
	if urlStr == "" {
		return nil // 空URL是允许的（对于可选字段）
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("URL格式不正确")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL必须使用http或https协议")
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("URL必须包含有效的主机名")
	}

	return nil
}
