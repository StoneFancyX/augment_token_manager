package repository

import (
	"augment_token_manager/internal/database"
	"augment_token_manager/internal/models"
	"database/sql"
	"fmt"
	"math/rand"
	"time"
)

// TokenRepository Token 数据访问层
type TokenRepository struct{}

// PaginationParams 分页参数
type PaginationParams struct {
	Page  int `json:"page" form:"page"`   // 当前页码，从1开始
	Limit int `json:"limit" form:"limit"` // 每页记录数
}

// PaginationResult 分页结果
type PaginationResult struct {
	Data       []models.Token `json:"data"`        // 数据列表
	Total      int64          `json:"total"`       // 总记录数
	Page       int            `json:"page"`        // 当前页码
	Limit      int            `json:"limit"`       // 每页记录数
	TotalPages int            `json:"total_pages"` // 总页数
	HasNext    bool           `json:"has_next"`    // 是否有下一页
	HasPrev    bool           `json:"has_prev"`    // 是否有上一页
}

// NewTokenRepository 创建新的 TokenRepository 实例
func NewTokenRepository() *TokenRepository {
	return &TokenRepository{}
}

// GetAllTokens 获取所有 Token
func (r *TokenRepository) GetAllTokens() ([]models.Token, error) {
	query := `
		SELECT id, tenant_url, access_token, portal_url, email_note,
		       ban_status::text as ban_status,
		       portal_info::text as portal_info,
		       created_at, updated_at
		FROM tokens
		ORDER BY created_at DESC
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询 tokens 失败: %v", err)
	}
	defer rows.Close()

	var tokens []models.Token
	for rows.Next() {
		var token models.Token
		err := rows.Scan(
			&token.ID,
			&token.TenantURL,
			&token.AccessToken,
			&token.PortalURL,
			&token.EmailNote,
			&token.BanStatus,
			&token.PortalInfo,
			&token.CreatedAt,
			&token.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描 token 数据失败: %v", err)
		}
		tokens = append(tokens, token)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历结果集失败: %v", err)
	}

	return tokens, nil
}

// GetTokensWithPagination 获取分页的 Token 列表
func (r *TokenRepository) GetTokensWithPagination(params PaginationParams) (*PaginationResult, error) {
	// 设置默认值
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 10
	}

	// 计算偏移量
	offset := (params.Page - 1) * params.Limit

	// 获取总记录数
	var total int64
	countQuery := `SELECT COUNT(*) FROM tokens`
	err := database.DB.QueryRow(countQuery).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("获取总记录数失败: %v", err)
	}

	// 获取分页数据
	query := `
		SELECT id, tenant_url, access_token, portal_url, email_note,
		       ban_status::text as ban_status,
		       portal_info::text as portal_info,
		       created_at, updated_at
		FROM tokens
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := database.DB.Query(query, params.Limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询分页 tokens 失败: %v", err)
	}
	defer rows.Close()

	var tokens []models.Token
	for rows.Next() {
		var token models.Token
		err := rows.Scan(
			&token.ID,
			&token.TenantURL,
			&token.AccessToken,
			&token.PortalURL,
			&token.EmailNote,
			&token.BanStatus,
			&token.PortalInfo,
			&token.CreatedAt,
			&token.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描 token 数据失败: %v", err)
		}
		tokens = append(tokens, token)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历结果集失败: %v", err)
	}

	// 计算总页数
	totalPages := int((total + int64(params.Limit) - 1) / int64(params.Limit))

	// 构建分页结果
	result := &PaginationResult{
		Data:       tokens,
		Total:      total,
		Page:       params.Page,
		Limit:      params.Limit,
		TotalPages: totalPages,
		HasNext:    params.Page < totalPages,
		HasPrev:    params.Page > 1,
	}

	return result, nil
}

// GetTokenByID 根据 ID 获取单个 Token
func (r *TokenRepository) GetTokenByID(id string) (*models.Token, error) {
	query := `
		SELECT id, tenant_url, access_token, portal_url, email_note,
		       ban_status::text as ban_status,
		       portal_info::text as portal_info,
		       created_at, updated_at
		FROM tokens
		WHERE id = $1
	`

	var token models.Token
	err := database.DB.QueryRow(query, id).Scan(
		&token.ID,
		&token.TenantURL,
		&token.AccessToken,
		&token.PortalURL,
		&token.EmailNote,
		&token.BanStatus,
		&token.PortalInfo,
		&token.CreatedAt,
		&token.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("获取 token 失败: %v", err)
	}

	return &token, nil
}

// CreateTokenRequest 创建Token的请求结构
type CreateTokenRequest struct {
	TenantURL   string `json:"tenant_url" binding:"required"`
	AccessToken string `json:"access_token" binding:"required"`
	PortalURL   string `json:"portal_url"`
	EmailNote   string `json:"email_note"`
}

// UpdateTokenRequest 更新Token的请求结构
type UpdateTokenRequest struct {
	TenantURL   string `json:"tenant_url" binding:"required"`
	AccessToken string `json:"access_token" binding:"required"`
	PortalURL   string `json:"portal_url"`
	EmailNote   string `json:"email_note"`
}

// CreateToken 创建新的Token
func (r *TokenRepository) CreateToken(req CreateTokenRequest) (*models.Token, error) {
	// 生成唯一的Token ID
	tokenID := r.generateTokenID()

	// 准备数据库字段
	var tenantURL, accessToken, portalURL, emailNote sql.NullString

	tenantURL = sql.NullString{String: req.TenantURL, Valid: true}
	accessToken = sql.NullString{String: req.AccessToken, Valid: true}

	if req.PortalURL != "" {
		portalURL = sql.NullString{String: req.PortalURL, Valid: true}
	}

	if req.EmailNote != "" {
		emailNote = sql.NullString{String: req.EmailNote, Valid: true}
	}

	// 插入数据库
	query := `
		INSERT INTO tokens (id, tenant_url, access_token, portal_url, email_note, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING created_at, updated_at
	`

	var createdAt, updatedAt time.Time
	err := database.DB.QueryRow(query, tokenID, tenantURL, accessToken, portalURL, emailNote).Scan(&createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("创建 Token 失败: %v", err)
	}

	// 构建返回的Token对象
	token := &models.Token{
		ID:          tokenID,
		TenantURL:   tenantURL,
		AccessToken: accessToken,
		PortalURL:   portalURL,
		EmailNote:   emailNote,
		BanStatus:   sql.NullString{String: "{}", Valid: true},
		PortalInfo:  sql.NullString{String: "{}", Valid: true},
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}

	return token, nil
}

// generateTokenID 生成唯一的Token ID
func (r *TokenRepository) generateTokenID() string {
	// 使用时间戳和随机字符串生成唯一ID
	timestamp := time.Now().UnixMilli()
	randomStr := r.generateRandomString(10)
	return fmt.Sprintf("token_%d_%s", timestamp, randomStr)
}

// generateRandomString 生成指定长度的随机字符串
func (r *TokenRepository) generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

	// 使用当前时间作为随机种子
	rand.Seed(time.Now().UnixNano())

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

// DeleteToken 删除指定ID的Token
func (r *TokenRepository) DeleteToken(tokenID string) error {
	// 首先检查Token是否存在
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM tokens WHERE id = $1)`
	err := database.DB.QueryRow(checkQuery, tokenID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("检查 Token 是否存在失败: %v", err)
	}

	if !exists {
		return fmt.Errorf("Token 不存在")
	}

	// 执行删除操作
	deleteQuery := `DELETE FROM tokens WHERE id = $1`
	result, err := database.DB.Exec(deleteQuery, tokenID)
	if err != nil {
		return fmt.Errorf("删除 Token 失败: %v", err)
	}

	// 检查是否真的删除了记录
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取删除结果失败: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("Token 删除失败，没有记录被删除")
	}

	return nil
}

// UpdateToken 更新指定ID的Token
func (r *TokenRepository) UpdateToken(tokenID string, req UpdateTokenRequest) (*models.Token, error) {
	// 首先检查Token是否存在
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM tokens WHERE id = $1)`
	err := database.DB.QueryRow(checkQuery, tokenID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("检查 Token 是否存在失败: %v", err)
	}

	if !exists {
		return nil, fmt.Errorf("Token 不存在")
	}

	// 准备更新的数据
	var tenantURL, accessToken, portalURL, emailNote sql.NullString

	tenantURL = sql.NullString{String: req.TenantURL, Valid: true}
	accessToken = sql.NullString{String: req.AccessToken, Valid: true}

	if req.PortalURL != "" {
		portalURL = sql.NullString{String: req.PortalURL, Valid: true}
	} else {
		portalURL = sql.NullString{Valid: false}
	}

	if req.EmailNote != "" {
		emailNote = sql.NullString{String: req.EmailNote, Valid: true}
	} else {
		emailNote = sql.NullString{Valid: false}
	}

	// 执行更新操作
	updateQuery := `
		UPDATE tokens
		SET tenant_url = $2, access_token = $3, portal_url = $4, email_note = $5, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING created_at, updated_at
	`

	var createdAt, updatedAt time.Time
	err = database.DB.QueryRow(updateQuery, tokenID, tenantURL, accessToken, portalURL, emailNote).Scan(&createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("更新 Token 失败: %v", err)
	}

	// 获取完整的Token信息（包括ban_status和portal_info）
	return r.GetTokenByID(tokenID)
}

// UpdateTokenBanStatus 更新Token的ban_status字段
func (r *TokenRepository) UpdateTokenBanStatus(tokenID, banStatus string) error {
	var updateQuery string
	var err error

	if banStatus == "" {
		// 清除ban_status，设置为null
		updateQuery = `
			UPDATE tokens
			SET ban_status = NULL, updated_at = CURRENT_TIMESTAMP
			WHERE id = $1
		`
		_, err = database.DB.Exec(updateQuery, tokenID)
	} else {
		// 设置具体的ban_status值
		updateQuery = `
			UPDATE tokens
			SET ban_status = $1, updated_at = CURRENT_TIMESTAMP
			WHERE id = $2
		`
		_, err = database.DB.Exec(updateQuery, banStatus, tokenID)
	}

	if err != nil {
		return fmt.Errorf("更新 Token ban_status 失败: %v", err)
	}

	return nil
}
