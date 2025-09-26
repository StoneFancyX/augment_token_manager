import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { tokensApi } from '../api/tokens'
import type { Token, TokenCreate, TokenUpdate, TokenState } from '../types/token'

export const useTokensStore = defineStore('tokens', () => {
  // State
  const tokens = ref<Token[]>([])
  const currentToken = ref<Token | null>(null)
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // Getters
  const tokenCount = computed(() => tokens.value.length)
  const validTokens = computed(() => {
    // 定义封禁状态列表（按照0.6.0版本的is_banned=true状态）
    const bannedStatuses = [
      'SUSPENDED', 'UNAUTHORIZED', 'FORBIDDEN', 'UNKNOWN_ERROR', 'INVALID',
      'INVALID_TOKEN', 'EXPIRED'
    ]

    return tokens.value.filter(token => {
      // 检查ban_status - 封禁状态的token不计入有效token
      if (token.ban_status && bannedStatuses.includes(token.ban_status)) {
        return false
      }

      // 检查使用次数是否耗尽
      if (token.is_exhausted) {
        return false
      }

      return true
    })
  })
  const invalidTokens = computed(() => {
    // 定义封禁状态列表（按照0.6.0版本的is_banned=true状态）
    const bannedStatuses = [
      'SUSPENDED', 'UNAUTHORIZED', 'FORBIDDEN', 'UNKNOWN_ERROR', 'INVALID',
      'INVALID_TOKEN', 'EXPIRED', 'EXHAUSTED', 'USAGE_LIMIT'
    ]

    return tokens.value.filter(token => {
      // 检查ban_status - 封禁状态的都认为无效
      if (token.ban_status && bannedStatuses.includes(token.ban_status)) {
        return true
      }

      // 检查使用次数是否耗尽
      if (token.is_exhausted) {
        return true
      }

      return false
    })
  })

  // Actions
  const fetchTokens = async (skip = 0, limit = 100) => {
    try {
      isLoading.value = true
      error.value = null
      const data = await tokensApi.getTokens(skip, limit)
      tokens.value = data
      return data
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch tokens'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const fetchToken = async (id: string) => {
    try {
      isLoading.value = true
      error.value = null
      const token = await tokensApi.getToken(id)
      currentToken.value = token
      return token
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch token'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const createToken = async (data: TokenCreate) => {
    try {
      isLoading.value = true
      error.value = null
      const newToken = await tokensApi.createToken(data)
      tokens.value.unshift(newToken)
      return newToken
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to create token'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const updateToken = async (id: string, data: TokenUpdate) => {
    try {
      isLoading.value = true
      error.value = null
      const updatedToken = await tokensApi.updateToken(id, data)
      
      // 更新列表中的token
      const index = tokens.value.findIndex(token => token.id === id)
      if (index !== -1) {
        tokens.value[index] = updatedToken
      }
      
      // 更新当前token
      if (currentToken.value?.id === id) {
        currentToken.value = updatedToken
      }
      
      return updatedToken
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to update token'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const deleteToken = async (id: string) => {
    try {
      isLoading.value = true
      error.value = null
      await tokensApi.deleteToken(id)
      
      // 从列表中移除token
      tokens.value = tokens.value.filter(token => token.id !== id)
      
      // 清除当前token
      if (currentToken.value?.id === id) {
        currentToken.value = null
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to delete token'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const validateToken = async (id: string) => {
    try {
      isLoading.value = true
      error.value = null
      const result = await tokensApi.validateToken(id)
      
      // 更新token状态
      const index = tokens.value.findIndex(token => token.id === id)
      if (index !== -1) {
        tokens.value[index] = result.token
      }
      
      return result
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to validate token'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const refreshToken = async (id: string) => {
    try {
      isLoading.value = true
      error.value = null
      const updatedToken = await tokensApi.refreshToken(id)
      
      // 更新列表中的token
      const index = tokens.value.findIndex(token => token.id === id)
      if (index !== -1) {
        tokens.value[index] = updatedToken
      }
      
      return updatedToken
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to refresh token'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const batchDeleteTokens = async (tokenIds: string[]) => {
    try {
      isLoading.value = true
      error.value = null
      const result = await tokensApi.batchDeleteTokens(tokenIds)
      
      // 从列表中移除已删除的tokens
      tokens.value = tokens.value.filter(token => !tokenIds.includes(token.id))
      
      return result
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to batch delete tokens'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const batchValidateTokens = async (tokenIds: string[]) => {
    try {
      isLoading.value = true
      error.value = null
      const result = await tokensApi.batchValidateTokens(tokenIds)
      
      // 更新tokens状态
      result.results.forEach(({ token_id, is_valid }) => {
        const index = tokens.value.findIndex(token => token.id === token_id)
        if (index !== -1) {
          tokens.value[index].status = is_valid ? undefined : 'INVALID'
        }
      })
      
      return result
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to batch validate tokens'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const batchRefreshTokens = async () => {
    try {
      isLoading.value = true
      error.value = null
      const result = await tokensApi.batchRefreshTokens()

      // 重新获取所有tokens以确保数据同步
      await fetchTokens()

      return result
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to batch refresh tokens'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const clearError = () => {
    error.value = null
  }

  // 导出Token数据
  const exportTokens = async (tokenIds: number[]) => {
    try {
      isLoading.value = true
      error.value = null

      return await tokensApi.exportTokens(tokenIds)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to export tokens'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  // 导入Token数据
  const importTokens = async (file: File) => {
    try {
      isLoading.value = true
      error.value = null

      return await tokensApi.importTokens(file)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to import tokens'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  return {
    // State
    tokens,
    currentToken,
    isLoading,
    error,
    // Getters
    tokenCount,
    validTokens,
    invalidTokens,
    // Actions
    fetchTokens,
    fetchToken,
    createToken,
    updateToken,
    deleteToken,
    validateToken,
    refreshToken,
    batchDeleteTokens,
    batchValidateTokens,
    batchRefreshTokens,
    exportTokens,
    importTokens,
    clearError,
  }
})
