package services

import (
	"augment_token_manager/internal/database"
	"augment_token_manager/internal/models"
	"augment_token_manager/internal/utils"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TokenRefreshService 处理 Token 刷新逻辑
type TokenRefreshService struct {
	httpClient *http.Client
}

// NewTokenRefreshService 创建新的 TokenRefreshService 实例
func NewTokenRefreshService() *TokenRefreshService {
	return &TokenRefreshService{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CustomerFromLinkResponse 第一步 API 响应结构
type CustomerFromLinkResponse struct {
	Customer struct {
		ID                 string `json:"id"`
		LedgerPricingUnits []struct {
			ID string `json:"id"`
		} `json:"ledger_pricing_units"`
	} `json:"customer"`
}

// LedgerSummaryResponse 第二步 API 响应结构
type LedgerSummaryResponse struct {
	CreditsBalance string `json:"credits_balance"`
	CreditBlocks   []struct {
		MaximumInitialBalance string    `json:"maximum_initial_balance"`
		ExpiryDate           string    `json:"expiry_date"`
		ID                   string    `json:"id"`
		PerUnitCostBasis     string    `json:"per_unit_cost_basis"`
		AllocationID         string    `json:"allocation_id"`
		EffectiveDate        string    `json:"effective_date"`
		Balance              string    `json:"balance"`
		IsActive             bool      `json:"is_active"`
	} `json:"credit_blocks"`
}

// RefreshTokenInfo 刷新单个 Token 的信息
func (s *TokenRefreshService) RefreshTokenInfo(tokenID string) (*models.Token, error) {
	utils.Debug("========== 开始刷新 Token: %s ==========", tokenID)

	// 数据准备阶段：从数据库获取 Token 信息
	utils.Debug("数据准备阶段：从数据库获取 Token 信息")
	token, err := s.getTokenFromDB(tokenID)
	if err != nil {
		utils.Error("获取 Token 信息失败: %v", err)
		return nil, fmt.Errorf("获取 Token 信息失败: %v", err)
	}
	utils.Debug("成功获取 Token 信息，ID: %s", token.ID)

	// 从 portal_url 中解析 token 参数
	portalURL := token.GetPortalURL()
	utils.Debug("获取到 portal_url: %s", portalURL)
	if portalURL == "" {
		utils.Error("Token 没有 portal_url 信息")
		return nil, fmt.Errorf("Token 没有 portal_url 信息")
	}

	tokenParam, err := s.extractTokenFromURL(portalURL)
	if err != nil {
		utils.Error("从 portal_url 解析 token 参数失败: %v", err)
		return nil, fmt.Errorf("从 portal_url 解析 token 参数失败: %v", err)
	}
	utils.Debug("成功解析 token 参数: %s", tokenParam)

	// 第一步：获取客户信息
	utils.Debug("========== 第一步：获取客户信息 ==========")
	customerInfo, err := s.getCustomerFromLink(tokenParam)
	if err != nil {
		utils.Error("获取客户信息失败: %v", err)
		return nil, fmt.Errorf("获取客户信息失败: %v", err)
	}
	utils.Debug("成功获取客户信息，客户ID: %s", customerInfo.Customer.ID)

	// 第二步：获取账户余额信息
	utils.Debug("========== 第二步：获取账户余额信息 ==========")
	ledgerInfo, err := s.getLedgerSummary(customerInfo, tokenParam)
	if err != nil {
		utils.Error("获取账户余额信息失败: %v", err)
		return nil, fmt.Errorf("获取账户余额信息失败: %v", err)
	}
	utils.Debug("成功获取账户余额信息，余额: %s", ledgerInfo.CreditsBalance)

	// 第三步：更新数据库
	utils.Debug("========== 第三步：更新数据库 ==========")
	updatedToken, err := s.updateTokenInDB(tokenID, ledgerInfo)
	if err != nil {
		utils.Error("更新数据库失败: %v", err)
		return nil, fmt.Errorf("更新数据库失败: %v", err)
	}
	utils.Debug("========== Token 刷新完成 ==========")

	return updatedToken, nil
}

// getTokenFromDB 从数据库获取 Token 信息
func (s *TokenRefreshService) getTokenFromDB(tokenID string) (*models.Token, error) {
	query := `
		SELECT id, tenant_url, access_token, portal_url, email_note, 
		       ban_status::text as ban_status,
		       portal_info::text as portal_info,
		       created_at, updated_at 
		FROM tokens 
		WHERE id = $1
	`

	var token models.Token
	err := database.DB.QueryRow(query, tokenID).Scan(
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
		return nil, err
	}

	return &token, nil
}

// extractTokenFromURL 从 portal_url 中提取 token 参数
func (s *TokenRefreshService) extractTokenFromURL(portalURL string) (string, error) {
	utils.Debug("解析 portal_url: %s", portalURL)

	u, err := url.Parse(portalURL)
	if err != nil {
		log.Printf("[ERROR] 解析 URL 失败: %v", err)
		return "", fmt.Errorf("解析 URL 失败: %v", err)
	}

	tokenParam := u.Query().Get("token")
	if tokenParam == "" {
		log.Printf("[ERROR] portal_url 中没有找到 token 参数")
		return "", fmt.Errorf("portal_url 中没有找到 token 参数")
	}

	utils.Debug("成功提取 token 参数: %s", tokenParam)
	return tokenParam, nil
}

// getCustomerFromLink 第一步：获取客户信息
func (s *TokenRefreshService) getCustomerFromLink(tokenParam string) (*CustomerFromLinkResponse, error) {
	// 构建客户信息 API URL
	apiURL := fmt.Sprintf("https://portal.withorb.com/api/v1/customer_from_link?token=%s", tokenParam)
	utils.Debug("构建客户信息 API URL: %s", apiURL)

	// 创建 HTTP 请求
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		utils.Error("创建 HTTP 请求失败: %v", err)
		return nil, fmt.Errorf("创建 HTTP 请求失败: %v", err)
	}

	// 添加必要的 HTTP 头部，模拟浏览器请求
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Referer", fmt.Sprintf("https://portal.withorb.com/view?token=%s", tokenParam))
	req.Header.Set("Origin", "https://portal.withorb.com")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")

	utils.Debug("发送 GET 请求到客户信息 API，包含 HTTP 头部")

	// 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Printf("[ERROR] 请求客户信息 API 失败: %v", err)
		return nil, fmt.Errorf("请求客户信息 API 失败: %v", err)
	}
	defer resp.Body.Close()

	utils.Debug("客户信息 API 响应状态码: %d", resp.StatusCode)

	// 读取响应体，处理可能的 gzip 压缩
	var reader io.Reader = resp.Body

	// 检查是否是 gzip 压缩
	if resp.Header.Get("Content-Encoding") == "gzip" {
		utils.Debug("检测到 gzip 压缩，进行解压")
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			log.Printf("[ERROR] 创建 gzip 读取器失败: %v", err)
			return nil, fmt.Errorf("创建 gzip 读取器失败: %v", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		log.Printf("[ERROR] 读取响应体失败: %v", err)
		return nil, fmt.Errorf("读取响应体失败: %v", err)
	}

	// 如果响应体看起来像是压缩的但没有正确的头部，尝试 gzip 解压
	if len(body) > 2 && body[0] == 0x1f && body[1] == 0x8b {
		utils.Debug("检测到 gzip 魔数，尝试解压缩")
		gzipReader, err := gzip.NewReader(strings.NewReader(string(body)))
		if err != nil {
			log.Printf("[ERROR] 创建 gzip 读取器失败: %v", err)
			return nil, fmt.Errorf("创建 gzip 读取器失败: %v", err)
		}
		defer gzipReader.Close()

		decompressed, err := io.ReadAll(gzipReader)
		if err != nil {
			log.Printf("[ERROR] gzip 解压失败: %v", err)
			return nil, fmt.Errorf("gzip 解压失败: %v", err)
		}
		body = decompressed
		utils.Debug("gzip 解压成功")
	}

	utils.Debug("客户信息 API 响应体: %s", string(body))

	if resp.StatusCode != http.StatusOK {
		utils.Error("客户信息 API 返回错误，状态码: %d, 响应体: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("客户信息 API 返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析 JSON 响应
	var customerResp CustomerFromLinkResponse
	if err := json.Unmarshal(body, &customerResp); err != nil {
		utils.Error("解析客户信息响应失败: %v, 响应体: %s", err, string(body))
		return nil, fmt.Errorf("解析客户信息响应失败: %v", err)
	}

	utils.Debug("成功解析客户信息:")
	utils.Debug("- 客户ID: %s", customerResp.Customer.ID)
	utils.Debug("- pricing units 数量: %d", len(customerResp.Customer.LedgerPricingUnits))
	if len(customerResp.Customer.LedgerPricingUnits) > 0 {
		utils.Debug("- 第一个 pricing unit ID: %s", customerResp.Customer.LedgerPricingUnits[0].ID)
	}

	return &customerResp, nil
}



// getLedgerSummary 第二步：获取账户余额信息
func (s *TokenRefreshService) getLedgerSummary(customerInfo *CustomerFromLinkResponse, tokenParam string) (*LedgerSummaryResponse, error) {
	if len(customerInfo.Customer.LedgerPricingUnits) == 0 {
		utils.Error("客户信息中没有 pricing unit")
		return nil, fmt.Errorf("客户信息中没有 pricing unit")
	}

	customerID := customerInfo.Customer.ID
	pricingUnitID := customerInfo.Customer.LedgerPricingUnits[0].ID
	utils.Debug("使用参数:")
	utils.Debug("- 客户ID: %s", customerID)
	utils.Debug("- pricing unit ID: %s", pricingUnitID)
	utils.Debug("- token: %s", tokenParam)

	// 构建账户余额 API URL
	apiURL := fmt.Sprintf("https://portal.withorb.com/api/v1/customers/%s/ledger_summary?pricing_unit_id=%s&token=%s",
		customerID, pricingUnitID, tokenParam)
	utils.Debug("构建账户余额 API URL: %s", apiURL)

	// 创建 HTTP 请求
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		utils.Error("创建账户余额 HTTP 请求失败: %v", err)
		return nil, fmt.Errorf("创建账户余额 HTTP 请求失败: %v", err)
	}

	// 添加必要的 HTTP 头部，避免压缩以防止乱码
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "identity") // 避免压缩
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Referer", fmt.Sprintf("https://portal.withorb.com/view?token=%s", tokenParam))
	req.Header.Set("Origin", "https://portal.withorb.com")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	utils.Debug("发送 GET 请求到账户余额 API，包含 HTTP 头部")

	// 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Printf("[ERROR] 请求账户余额 API 失败: %v", err)
		return nil, fmt.Errorf("请求账户余额 API 失败: %v", err)
	}
	defer resp.Body.Close()

	utils.Debug("账户余额 API 响应状态码: %d", resp.StatusCode)

	// 读取响应体，处理可能的 gzip 压缩
	var reader io.Reader = resp.Body

	// 检查是否是 gzip 压缩
	if resp.Header.Get("Content-Encoding") == "gzip" {
		utils.Debug("检测到 gzip 压缩，进行解压")
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			log.Printf("[ERROR] 创建 gzip 读取器失败: %v", err)
			return nil, fmt.Errorf("创建 gzip 读取器失败: %v", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		log.Printf("[ERROR] 读取账户余额响应体失败: %v", err)
		return nil, fmt.Errorf("读取响应体失败: %v", err)
	}

	// 如果响应体看起来像是压缩的但没有正确的头部，尝试 gzip 解压
	if len(body) > 2 && body[0] == 0x1f && body[1] == 0x8b {
		utils.Debug("检测到 gzip 魔数，尝试解压缩")
		gzipReader, err := gzip.NewReader(strings.NewReader(string(body)))
		if err != nil {
			log.Printf("[ERROR] 创建 gzip 读取器失败: %v", err)
			return nil, fmt.Errorf("创建 gzip 读取器失败: %v", err)
		}
		defer gzipReader.Close()

		decompressed, err := io.ReadAll(gzipReader)
		if err != nil {
			log.Printf("[ERROR] gzip 解压失败: %v", err)
			return nil, fmt.Errorf("gzip 解压失败: %v", err)
		}
		body = decompressed
		utils.Debug("gzip 解压成功")
	}

	utils.Debug("账户余额 API 响应体: %s", string(body))

	if resp.StatusCode != http.StatusOK {
		log.Printf("[ERROR] 账户余额 API 返回错误，状态码: %d, 响应体: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("账户余额 API 返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析 JSON 响应
	var ledgerResp LedgerSummaryResponse
	if err := json.Unmarshal(body, &ledgerResp); err != nil {
		log.Printf("[ERROR] 解析账户余额响应失败: %v, 响应体: %s", err, string(body))
		return nil, fmt.Errorf("解析账户余额响应失败: %v", err)
	}

	utils.Debug("成功解析账户余额信息:")
	utils.Debug("- credits_balance: %s", ledgerResp.CreditsBalance)
	utils.Debug("- credit_blocks 数量: %d", len(ledgerResp.CreditBlocks))
	if len(ledgerResp.CreditBlocks) > 0 {
		utils.Debug("- 第一个 credit_block:")
		utils.Debug("  - expiry_date: %s", ledgerResp.CreditBlocks[0].ExpiryDate)
		utils.Debug("  - is_active: %t", ledgerResp.CreditBlocks[0].IsActive)
		utils.Debug("  - balance: %s", ledgerResp.CreditBlocks[0].Balance)
	}

	return &ledgerResp, nil
}

// updateTokenInDB 更新数据库中的 Token 信息
func (s *TokenRefreshService) updateTokenInDB(tokenID string, ledgerInfo *LedgerSummaryResponse) (*models.Token, error) {
	utils.Debug("开始构建新的 portal_info JSON")

	// 构建新的 portal_info JSON 对象
	portalInfo := map[string]interface{}{}

	// 设置 credits_balance
	creditsBalance := parseCreditsBalance(ledgerInfo.CreditsBalance)
	portalInfo["credits_balance"] = creditsBalance
	utils.Debug("设置 credits_balance: %d", creditsBalance)

	// 设置 is_active 和 expiry_date
	if len(ledgerInfo.CreditBlocks) > 0 {
		firstBlock := ledgerInfo.CreditBlocks[0]
		portalInfo["is_active"] = firstBlock.IsActive
		portalInfo["expiry_date"] = firstBlock.ExpiryDate
		utils.Debug("设置 is_active: %t", firstBlock.IsActive)
		utils.Debug("设置 expiry_date: %s", firstBlock.ExpiryDate)
	} else {
		portalInfo["is_active"] = false
		portalInfo["expiry_date"] = ""
		utils.Debug("没有 credit_blocks，设置默认值")
	}

	// 序列化为 JSON
	portalInfoJSON, err := json.Marshal(portalInfo)
	if err != nil {
		utils.Error("序列化 portal_info 失败: %v", err)
		return nil, fmt.Errorf("序列化 portal_info 失败: %v", err)
	}
	utils.Debug("构建的 portal_info JSON: %s", string(portalInfoJSON))

	// 更新数据库
	updateQuery := `
		UPDATE tokens
		SET portal_info = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	utils.Debug("执行数据库更新查询，Token ID: %s", tokenID)
	result, err := database.DB.Exec(updateQuery, string(portalInfoJSON), tokenID)
	if err != nil {
		utils.Error("更新数据库失败: %v", err)
		return nil, fmt.Errorf("更新数据库失败: %v", err)
	}

	// 检查更新是否成功
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		utils.Error("获取受影响行数失败: %v", err)
	} else {
		utils.Debug("数据库更新成功，受影响行数: %d", rowsAffected)
	}

	// 返回更新后的 Token 信息
	utils.Debug("获取更新后的 Token 信息")
	updatedToken, err := s.getTokenFromDB(tokenID)
	if err != nil {
		utils.Error("获取更新后的 Token 信息失败: %v", err)
		return nil, fmt.Errorf("获取更新后的 Token 信息失败: %v", err)
	}

	utils.Debug("成功获取更新后的 Token 信息")
	return updatedToken, nil
}

// parseCreditsBalance 解析 credits_balance 字符串为数字
func parseCreditsBalance(creditsStr string) int {
	// 移除小数点，将 "16.00" 转换为 16
	if dotIndex := strings.Index(creditsStr, "."); dotIndex != -1 {
		creditsStr = creditsStr[:dotIndex]
	}
	
	// 简单的字符串到数字转换
	var credits int
	fmt.Sscanf(creditsStr, "%d", &credits)
	return credits
}
