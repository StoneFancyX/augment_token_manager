import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '../api/auth'
import type { User, LoginRequest, AuthState } from '../types/auth'

export const useAuthStore = defineStore('auth', () => {
  // State
  const user = ref<User | null>(null)
  const token = ref<string | null>(null)
  const isLoading = ref(false)

  // Getters
  const isAuthenticated = computed(() => !!token.value && !!user.value)

  // Actions
  const login = async (credentials: LoginRequest) => {
    try {
      isLoading.value = true
      const response = await authApi.login(credentials)
      
      token.value = response.access_token
      user.value = response.user
      
      // 保存到本地存储
      localStorage.setItem('access_token', response.access_token)
      localStorage.setItem('user', JSON.stringify(response.user))
      
      return response
    } catch (error) {
      console.error('Login failed:', error)
      throw error
    } finally {
      isLoading.value = false
    }
  }

  const logout = async () => {
    try {
      await authApi.logout()
    } catch (error) {
      console.error('Logout failed:', error)
    } finally {
      // 清除状态和本地存储
      user.value = null
      token.value = null
      localStorage.removeItem('access_token')
      localStorage.removeItem('user')
    }
  }

  const getCurrentUser = async () => {
    try {
      isLoading.value = true
      const userData = await authApi.getCurrentUser()
      user.value = userData
      return userData
    } catch (error) {
      console.error('Get current user failed:', error)
      // 如果获取用户信息失败，清除认证状态
      await logout()
      throw error
    } finally {
      isLoading.value = false
    }
  }

  const initializeAuth = () => {
    // 从本地存储恢复认证状态
    const savedToken = localStorage.getItem('access_token')
    const savedUser = localStorage.getItem('user')
    
    if (savedToken && savedUser) {
      try {
        token.value = savedToken
        user.value = JSON.parse(savedUser)
        // 验证token是否仍然有效
        getCurrentUser().catch(() => {
          // Token无效，清除状态
          logout()
        })
      } catch (error) {
        console.error('Failed to parse saved user data:', error)
        logout()
      }
    }
  }

  return {
    // State
    user,
    token,
    isLoading,
    // Getters
    isAuthenticated,
    // Actions
    login,
    logout,
    getCurrentUser,
    initializeAuth,
  }
})
