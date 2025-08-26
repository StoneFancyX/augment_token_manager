package models

import (
	"database/sql"
	"strings"
	"time"
)

// Token 表示 Augment Token 的数据结构
// 适配现有数据库表结构
type Token struct {
	ID          string         `json:"id"`
	TenantURL   sql.NullString `json:"tenant_url"`
	AccessToken sql.NullString `json:"access_token"`
	PortalURL   sql.NullString `json:"portal_url"`
	EmailNote   sql.NullString `json:"email_note"`
	BanStatus   sql.NullString `json:"ban_status"`
	PortalInfo  sql.NullString `json:"portal_info"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// GetTenantURL 获取 TenantURL 的字符串值
func (t *Token) GetTenantURL() string {
	if t.TenantURL.Valid {
		return t.TenantURL.String
	}
	return ""
}

// GetAccessToken 获取 AccessToken 的字符串值
func (t *Token) GetAccessToken() string {
	if t.AccessToken.Valid {
		return t.AccessToken.String
	}
	return ""
}

// GetPortalURL 获取 PortalURL 的字符串值
func (t *Token) GetPortalURL() string {
	if t.PortalURL.Valid {
		return t.PortalURL.String
	}
	return ""
}

// GetEmailNote 获取 EmailNote 的字符串值
func (t *Token) GetEmailNote() string {
	if t.EmailNote.Valid {
		return t.EmailNote.String
	}
	return ""
}

// GetBanStatus 获取 BanStatus 的字符串值
func (t *Token) GetBanStatus() string {
	if t.BanStatus.Valid {
		return t.BanStatus.String
	}
	return "{}"
}

// GetPortalInfo 获取 PortalInfo 的字符串值
func (t *Token) GetPortalInfo() string {
	if t.PortalInfo.Valid {
		return t.PortalInfo.String
	}
	return "{}"
}



// TokenResponse 用于 API 响应的简化结构
type TokenResponse struct {
	ID          string `json:"id"`
	TenantURL   string `json:"tenant_url"`
	AccessToken string `json:"access_token"`
	PortalURL   string `json:"portal_url"`
	EmailNote   string `json:"email_note"`
	BanStatus   string `json:"ban_status"`
	PortalInfo  string `json:"portal_info"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ToResponse 将 Token 转换为 TokenResponse
func (t *Token) ToResponse() TokenResponse {
	return TokenResponse{
		ID:          t.ID,
		TenantURL:   t.GetTenantURL(),
		AccessToken: t.GetAccessToken(),
		PortalURL:   t.GetPortalURL(),
		EmailNote:   t.GetEmailNote(),
		BanStatus:   t.GetBanStatus(),
		PortalInfo:  t.GetPortalInfo(),
		CreatedAt:   t.CreatedAt.Local().Format("2006-01-02 15:04:05"),
		UpdatedAt:   t.UpdatedAt.Local().Format("2006-01-02 15:04:05"),
	}
}

// GetExpiryDate 从 portal_info JSON 中解析过期时间
func (t *Token) GetExpiryDate() string {
	portalInfo := t.GetPortalInfo()
	if portalInfo == "{}" {
		return "未知"
	}

	// 简单的 JSON 解析，提取 expiry_date
	// 这里使用字符串匹配而不是完整的 JSON 解析以保持简单
	start := strings.Index(portalInfo, "\"expiry_date\": \"")
	if start == -1 {
		return "未知"
	}
	start += len("\"expiry_date\": \"")
	end := strings.Index(portalInfo[start:], "\"")
	if end == -1 {
		return "未知"
	}

	expiryStr := portalInfo[start : start+end]
	// 解析 ISO 8601 时间格式并转换为本地时间显示
	if expiryTime, err := time.Parse("2006-01-02T15:04:05Z07:00", expiryStr); err == nil {
		// 转换为本地时区
		localTime := expiryTime.Local()
		return localTime.Format("2006-01-02 15:04")
	}

	return expiryStr
}

// GetCreditsBalance 从 portal_info JSON 中解析剩余次数
func (t *Token) GetCreditsBalance() string {
	portalInfo := t.GetPortalInfo()
	if portalInfo == "{}" {
		return "0"
	}

	// 简单的 JSON 解析，提取 credits_balance
	start := strings.Index(portalInfo, "\"credits_balance\": ")
	if start == -1 {
		return "0"
	}
	start += len("\"credits_balance\": ")
	end := strings.IndexAny(portalInfo[start:], ",}")
	if end == -1 {
		return "0"
	}

	return portalInfo[start : start+end]
}

// GetFormattedCreatedAt 获取格式化的创建时间（本地时区）
func (t *Token) GetFormattedCreatedAt() string {
	// 转换为本地时区并格式化
	localTime := t.CreatedAt.Local()
	return localTime.Format("2006-01-02 15:04")
}
