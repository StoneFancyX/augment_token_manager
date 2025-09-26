export enum TokenStatus {
  ACTIVE = "ACTIVE",
  SUSPENDED = "SUSPENDED",
  INVALID_TOKEN = "INVALID_TOKEN",
  USAGE_LIMIT = "USAGE_LIMIT",
  EXHAUSTED = "EXHAUSTED",
  EXPIRED = "EXPIRED",
  INVALID = "INVALID"
}

export interface Token {
  id: string
  email_note?: string
  access_token: string
  tenant_url: string
  portal_url: string
  portal_info?: Record<string, any>
  ban_status?: string
  usage_count: number
  max_usage?: number
  status_display: string
  is_exhausted: boolean
  created_at: string
  updated_at: string
}

export interface TokenCreate {
  email_note?: string
  access_token: string
  tenant_url: string
  portal_url: string
  max_usage?: number
}

export interface TokenUpdate {
  email_note?: string
  access_token?: string
  tenant_url?: string
  portal_url?: string
  max_usage?: number
}

export interface TokenValidationResult {
  is_valid: boolean
  message: string
  token: Token
}

export interface TokenImportRequest {
  tokens: TokenCreate[]
}

export interface TokenImportResponse {
  success_count: number
  failed_count: number
  tokens: Token[]
  errors: string[]
}

export interface TokenState {
  tokens: Token[]
  currentToken: Token | null
  isLoading: boolean
  error: string | null
}
