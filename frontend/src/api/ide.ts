import { apiClient } from './client'

export interface IDEOpenRequest {
  editor_type: string
  token: string
  tenant_url: string
  portal_url?: string
}

export interface Editor {
  id: string
  name: string
  icon: string
}

export interface SupportedEditorsResponse {
  vscode_editors: Editor[]
  jetbrains_editors: Editor[]
}

export interface IDEOpenResponse {
  success: boolean
  message: string
  protocol_url?: string
}

// 打开IDE
export const openEditor = async (request: IDEOpenRequest): Promise<IDEOpenResponse> => {
  return await apiClient.post<IDEOpenResponse>('/api/ide/open-editor', request)
}

// 获取支持的编辑器列表
export const getSupportedEditors = async (): Promise<SupportedEditorsResponse> => {
  return await apiClient.get<SupportedEditorsResponse>('/api/ide/supported-editors')
}
