import { apiClient } from './client'
import type {
  Token,
  TokenCreate,
  TokenUpdate,
  TokenValidationResult,
  TokenImportRequest,
  TokenImportResponse,
} from '../types/token'

export const tokensApi = {
  // 获取Token列表
  async getTokens(skip = 0, limit = 100): Promise<Token[]> {
    return apiClient.get<Token[]>(`/api/tokens?skip=${skip}&limit=${limit}`)
  },

  // 获取单个Token
  async getToken(id: string): Promise<Token> {
    return apiClient.get<Token>(`/api/tokens/${id}`)
  },

  // 创建Token
  async createToken(data: TokenCreate): Promise<Token> {
    return apiClient.post<Token>('/api/tokens', data)
  },

  // 更新Token
  async updateToken(id: string, data: TokenUpdate): Promise<Token> {
    return apiClient.put<Token>(`/api/tokens/${id}`, data)
  },

  // 删除Token
  async deleteToken(id: string): Promise<void> {
    return apiClient.delete<void>(`/api/tokens/${id}`)
  },

  // 验证Token
  async validateToken(id: string): Promise<TokenValidationResult> {
    return apiClient.post<TokenValidationResult>(`/api/tokens/${id}/validate`)
  },

  // 刷新Token信息
  async refreshToken(id: string): Promise<Token> {
    return apiClient.post<Token>(`/api/tokens/${id}/refresh`)
  },

  // 批量导入Token
  async importTokens(data: TokenImportRequest): Promise<TokenImportResponse> {
    return apiClient.post<TokenImportResponse>('/api/tokens/import', data)
  },

  // 批量删除Token
  async batchDeleteTokens(tokenIds: string[]): Promise<{
    success_count: number
    failed_count: number
    errors: string[]
    message: string
  }> {
    return apiClient.post('/api/tokens/batch-delete', tokenIds)
  },

  // 批量验证Token
  async batchValidateTokens(tokenIds: string[]): Promise<{
    results: Array<{
      token_id: string
      is_valid: boolean
      message: string
      error: boolean
    }>
    success_count: number
    failed_count: number
    message: string
  }> {
    return apiClient.post('/api/tokens/batch-validate', tokenIds)
  },

  // 批量刷新Token
  async batchRefreshTokens(): Promise<{
    results: Array<{
      token_id: string
      email_note: string
      status: string
      is_valid?: boolean
      message?: string
      error?: string
    }>
    success_count: number
    failed_count: number
    message: string
  }> {
    return apiClient.post('/api/tokens/batch-refresh')
  },

  // 导出Token数据
  async exportTokens(tokenIds: number[]): Promise<Blob> {
    const response = await fetch('/api/tokens/export', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('access_token')}`
      },
      body: JSON.stringify(tokenIds)
    })

    if (!response.ok) {
      throw new Error('导出失败')
    }

    return response.blob()
  },

  // 导入Token文件
  async importTokensFromFile(file: File): Promise<{
    message: string
    success_count: number
    failed_count: number
    results: Array<{
      success: boolean
      token_id?: number
      email_note?: string
      error?: string
      data?: any
    }>
  }> {
    const formData = new FormData()
    formData.append('file', file)

    const response = await fetch('/api/tokens/import', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('access_token')}`
      },
      body: formData
    })

    if (!response.ok) {
      throw new Error('导入失败')
    }

    return response.json()
  },

  // 获取Token统计信息
  async getTokenStats(): Promise<{
    total_tokens: number
    total_credits: number
    available_credits: number
    unlimited_tokens: number
    expired_tokens: number
    valid_tokens: number
  }> {
    return apiClient.get('/api/tokens/statistics')
  },
}
