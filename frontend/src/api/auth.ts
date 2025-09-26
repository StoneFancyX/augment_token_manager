import { apiClient } from './client'
import type { LoginRequest, LoginResponse, User } from '../types/auth'

export const authApi = {
  // 用户登录
  async login(credentials: LoginRequest): Promise<LoginResponse> {
    const formData = new FormData()
    formData.append('username', credentials.username)
    formData.append('password', credentials.password)
    
    return apiClient.post<LoginResponse>('/api/auth/login', formData, {
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
    })
  },

  // 获取当前用户信息
  async getCurrentUser(): Promise<User> {
    return apiClient.get<User>('/api/auth/me')
  },

  // 用户登出
  async logout(): Promise<void> {
    return apiClient.post<void>('/api/auth/logout')
  },
}
