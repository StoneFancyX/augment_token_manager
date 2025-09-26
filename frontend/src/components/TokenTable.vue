<script setup lang="ts">
import { ref, computed, h } from 'vue'
import { useTokensStore } from '../stores/tokens'
import { message, Modal } from 'ant-design-vue'
import {
  EditOutlined,
  DeleteOutlined,
  ReloadOutlined,
  ExclamationCircleOutlined,
  CodeOutlined,
  DownloadOutlined,
  FileTextOutlined,
  LinkOutlined,
  CopyOutlined
} from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import type { Token } from '../types/token'
import IDEModal from './IDEModal.vue'

interface Props {
  tokens: Token[]
}

const props = defineProps<Props>()
const tokensStore = useTokensStore()

const selectedRowKeys = ref<string[]>([])
const loading = ref(false)
const editModalVisible = ref(false)
const editingToken = ref<Token | null>(null)
const selectedToken = ref<Token | null>(null)
const ideModalRef = ref()

const formatDate = (dateString: string) => {
  return dayjs(dateString).format('YYYY-MM-DD HH:mm:ss')
}

const formatEmailNote = (emailNote?: string) => {
  if (!emailNote) {
    return '未设置'
  }

  // 如果长度小于等于15，直接显示
  if (emailNote.length <= 15) {
    return emailNote
  }

  // 如果包含@符号，特殊处理邮箱
  if (emailNote.includes('@')) {
    const [localPart, domain] = emailNote.split('@')

    // 如果用户名部分太长，省略中间
    if (localPart.length > 6) {
      const start = localPart.substring(0, 3)
      const end = localPart.substring(localPart.length - 2)
      return `${start}...${end}@${domain}`
    }

    return emailNote
  }

  // 如果不包含@符号，简单的首尾保留
  const start = emailNote.substring(0, 6)
  const end = emailNote.substring(emailNote.length - 4)
  return `${start}...${end}`
}

const getStatusTag = (token: Token) => {
  // 优先检查ban_status（按照0.6.0版本逻辑）
  if (token.ban_status) {
    const status = token.ban_status
    // 按照0.6.0版本的is_banned逻辑映射
    if (status === 'SUSPENDED') return { color: 'orange', text: '已封禁' }
    if (status === 'INVALID_TOKEN') return { color: 'red', text: 'Token失效' }
    if (status === 'ACTIVE') return { color: 'green', text: '正常' }
    // 以下状态在0.6.0版本中is_banned=true，都应显示为"已封禁"
    if (status === 'UNAUTHORIZED') return { color: 'orange', text: '已封禁' }
    if (status === 'FORBIDDEN') return { color: 'orange', text: '已封禁' }
    if (status === 'UNKNOWN_ERROR') return { color: 'orange', text: '已封禁' }
    if (status === 'INVALID') return { color: 'orange', text: '已封禁' }  // INVALID也应该显示为已封禁
    // 以下状态在0.6.0版本中is_banned=false，显示具体状态
    if (status === 'RATE_LIMITED') return { color: 'blue', text: '限流中' }
    if (status === 'SERVER_ERROR') return { color: 'gray', text: '服务器错误' }
    // 兼容旧状态
    if (status === 'EXHAUSTED' || status === 'USAGE_LIMIT') return { color: 'purple', text: '使用次数耗尽' }
    if (status === 'EXPIRED') return { color: 'gray', text: '已过期' }
    return { color: 'volcano', text: status }  // 未知状态用火山红
  }

  // 然后检查portal_info中的余额状态
  if (token.portal_info && typeof token.portal_info === 'object') {
    const creditsBalance = token.portal_info.credits_balance || 0
    const isActive = token.portal_info.is_active !== false

    // 如果余额为0，状态仍显示为正常（具体余额信息在额度列显示）
    if (creditsBalance === 0) {
      return { color: 'green', text: '正常' }
    }

    // 如果余额大于0，状态显示为正常（余额信息在额度列显示）
    if (creditsBalance > 0) {
      return { color: 'green', text: '正常' }
    }

    // 如果不活跃
    if (!isActive) {
      return { color: 'red', text: '不活跃' }
    }
  }

  // 检查使用次数是否耗尽
  if (token.is_exhausted) {
    return { color: 'purple', text: '次数耗尽' }
  }

  // 默认为正常
  return { color: 'green', text: '正常' }
}

// 获取额度显示文本（按照原版逻辑 - 只显示portal_info的额度，不管ban_status）
const getCreditsText = (token: Token) => {
  if (token.portal_info && typeof token.portal_info === 'object') {
    const creditsBalance = token.portal_info.credits_balance || 0

    // 检查是否有无限制权限
    const hasUnlimitedUsage = token.portal_info.subscription_info &&
                             token.portal_info.subscription_info.plan_type === 'unlimited'

    if (creditsBalance === 0) {
      if (hasUnlimitedUsage) {
        return '还能使用'
      } else {
        return '使用次数耗尽'
      }
    } else {
      return `剩余 ${creditsBalance}`
    }
  }

  return '-'
}

// 获取额度显示样式类
const getCreditsClass = (token: Token) => {
  if (token.portal_info && typeof token.portal_info === 'object') {
    const creditsBalance = token.portal_info.credits_balance || 0
    const hasUnlimitedUsage = token.portal_info.subscription_info &&
                             token.portal_info.subscription_info.plan_type === 'unlimited'

    if (creditsBalance === 0 && !hasUnlimitedUsage) {
      return 'exhausted'
    } else if (creditsBalance === 0 && hasUnlimitedUsage) {
      return 'unlimited'
    } else if (creditsBalance > 0) {
      return 'available'
    }
  }

  return 'unknown'
}

// 获取过期时间显示文本（显示剩余时间，精确到分钟）
const getExpiryText = (token: Token) => {
  if (!token.portal_info || !token.portal_info.expiry_date) {
    return '-'
  }

  const expiryDate = new Date(token.portal_info.expiry_date)
  const now = new Date()
  const diffMs = expiryDate.getTime() - now.getTime()

  if (diffMs < 0) {
    return '已过期'
  }

  const totalMinutes = Math.floor(diffMs / (1000 * 60))
  const days = Math.floor(totalMinutes / (24 * 60))
  const hours = Math.floor((totalMinutes % (24 * 60)) / 60)
  const minutes = totalMinutes % 60

  if (days > 0) {
    if (hours > 0) {
      return `${days}天${hours}小时${minutes}分钟`
    } else if (minutes > 0) {
      return `${days}天${minutes}分钟`
    } else {
      return `${days}天`
    }
  } else if (hours > 0) {
    if (minutes > 0) {
      return `${hours}小时${minutes}分钟`
    } else {
      return `${hours}小时`
    }
  } else {
    return minutes > 0 ? `${minutes}分钟` : '即将过期'
  }
}

// 获取过期时间样式类
const getExpiryClass = (token: Token) => {
  if (!token.portal_info || !token.portal_info.expiry_date) {
    return 'expiry-unknown'
  }

  const expiryDate = new Date(token.portal_info.expiry_date)
  const now = new Date()
  const diffDays = Math.ceil((expiryDate.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))

  if (diffDays < 0) {
    return 'expiry-expired'  // 已过期
  } else if (diffDays <= 1) {
    return 'expiry-warning'  // 1天内过期
  } else if (diffDays <= 3) {
    return 'expiry-caution'  // 3天内过期
  } else {
    return 'expiry-normal'   // 正常（4-7天）
  }
}

const columns = [
  {
    title: 'Email Note',
    dataIndex: 'email_note',
    key: 'email_note',
    width: 250,
    ellipsis: {
      showTitle: false,
    },
  },
  {
    title: '剩余时间',
    dataIndex: 'expiry_date',
    key: 'expiry_date',
    width: 180,
    sorter: (a: Token, b: Token) => {
      const getExpiryMinutes = (token: Token) => {
        if (!token.portal_info || !token.portal_info.expiry_date) {
          return -1 // 无过期时间的排在最后
        }
        const expiryDate = new Date(token.portal_info.expiry_date)
        const now = new Date()
        const diffMs = expiryDate.getTime() - now.getTime()
        return Math.floor(diffMs / (1000 * 60)) // 转换为分钟
      }

      return getExpiryMinutes(a) - getExpiryMinutes(b)
    },
    defaultSortOrder: 'ascend' as const, // 默认升序，即将过期的在前
  },
  {
    title: '额度',
    dataIndex: 'credits',
    key: 'credits',
    width: 120,
    filters: [
      { text: '有余额', value: 'has_credits' },
      { text: '使用次数耗尽', value: 'exhausted' },
      { text: '还能使用', value: 'unlimited' },
      { text: '未知', value: 'unknown' },
    ],
    onFilter: (value: string, record: Token) => {
      if (!record.portal_info) {
        return value === 'unknown'
      }

      const credits = record.portal_info.credits_balance || 0
      const hasUnlimited = record.portal_info.subscription_info?.plan_type === 'unlimited'

      if (value === 'has_credits') {
        return credits > 0
      } else if (value === 'exhausted') {
        return credits === 0 && !hasUnlimited
      } else if (value === 'unlimited') {
        return credits === 0 && hasUnlimited
      } else if (value === 'unknown') {
        return !record.portal_info
      }

      return false
    },
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 100,
    filters: [
      { text: '正常', value: 'ACTIVE' },
      { text: 'Token失效', value: 'INVALID' },
      { text: '已封禁', value: 'SUSPENDED' },
      { text: '使用次数耗尽', value: 'EXHAUSTED' },
      { text: '已过期', value: 'EXPIRED' },
    ],
    onFilter: (value: string, record: Token) => {
      // 处理多个状态值映射到同一显示文本的情况
      if (value === 'INVALID') {
        return record.status_display === 'INVALID_TOKEN' || record.status_display === 'INVALID'
      }
      if (value === 'EXHAUSTED') {
        return record.status_display === 'EXHAUSTED' || record.status_display === 'USAGE_LIMIT'
      }
      return record.status_display === value
    },
  },
  {
    title: '创建时间',
    dataIndex: 'created_at',
    key: 'created_at',
    width: 180,
  },
  {
    title: '操作',
    key: 'action',
    width: 240,
  },
]

const rowSelection = {
  selectedRowKeys: selectedRowKeys,
  onChange: (keys: string[]) => {
    selectedRowKeys.value = keys
  },
}



const editToken = (tokenId: string) => {
  const token = props.tokens.find(t => t.id === tokenId)
  if (token) {
    editingToken.value = { ...token }
    editModalVisible.value = true
  }
}

const saveEdit = async () => {
  if (!editingToken.value) return

  try {
    loading.value = true
    await tokensStore.updateToken(editingToken.value.id, {
      email_note: editingToken.value.email_note,
      tenant_url: editingToken.value.tenant_url,
      access_token: editingToken.value.access_token,
      portal_url: editingToken.value.portal_url,
      max_usage: editingToken.value.max_usage
    })
    editModalVisible.value = false
    editingToken.value = null
    message.success('更新成功')
  } catch (error) {
    message.error('更新失败')
  } finally {
    loading.value = false
  }
}

const cancelEdit = () => {
  editModalVisible.value = false
  editingToken.value = null
}

const copyEmailNote = async (emailNote?: string) => {
  if (!emailNote) {
    message.warning('邮箱备注为空，无法复制')
    return
  }

  try {
    await navigator.clipboard.writeText(emailNote)
    message.success('邮箱备注已复制到剪贴板')
  } catch (error) {
    // 降级方案：使用传统方法复制
    const textArea = document.createElement('textarea')
    textArea.value = emailNote
    document.body.appendChild(textArea)
    textArea.select()
    document.execCommand('copy')
    document.body.removeChild(textArea)
    message.success('邮箱备注已复制到剪贴板')
  }
}

const openPortal = (token: Token) => {
  if (token.portal_url) {
    window.open(token.portal_url, '_blank')
  } else {
    message.warning('该Token没有Portal URL')
  }
}

const copyTokenJSON = async (token: Token) => {
  const tokenData = {
    tenant_url: token.tenant_url,
    access_token: token.access_token,
    portal_url: token.portal_url,
    email_note: token.email_note || undefined
  }

  // 移除undefined值
  const cleanData = Object.fromEntries(
    Object.entries(tokenData).filter(([_, value]) => value !== undefined)
  )

  const jsonString = JSON.stringify(cleanData, null, 2)

  try {
    await navigator.clipboard.writeText(jsonString)
    message.success('Token JSON已复制到剪贴板')
  } catch (error) {
    // 降级方案：使用传统方法复制
    const textArea = document.createElement('textarea')
    textArea.value = jsonString
    document.body.appendChild(textArea)
    textArea.select()
    document.execCommand('copy')
    document.body.removeChild(textArea)
    message.success('Token JSON已复制到剪贴板')
  }
}

const deleteToken = async (tokenId: string) => {
  Modal.confirm({
    title: '确认删除',
    content: '确定要删除这个 Token 吗？',
    icon: h(ExclamationCircleOutlined),
    onOk: async () => {
      try {
        await tokensStore.deleteToken(tokenId)
        message.success('删除成功')
      } catch (error) {
        message.error('删除失败')
      }
    },
  })
}



const refreshToken = async (tokenId: string) => {
  try {
    loading.value = true

    // 按照0.6.0版本逻辑：先验证状态，然后刷新Portal信息
    const token = props.tokens.find(t => t.id === tokenId)

    // 1. 验证账号状态
    const result = await tokensStore.validateToken(tokenId)

    // 2. 如果有portal_url，刷新Portal信息
    if (token?.portal_url) {
      try {
        await tokensStore.refreshToken(tokenId)
      } catch (portalError) {
        console.warn('Portal信息刷新失败:', portalError)
        // Portal刷新失败不影响状态验证的成功提示
      }
    }

    // 根据验证结果显示不同消息
    if (result.is_valid) {
      message.success('Token状态正常')
    } else {
      message.warning('Token状态已更新')
    }
  } catch (error) {
    message.error('验证失败')
  } finally {
    loading.value = false
  }
}

const batchDelete = async () => {
  if (selectedRowKeys.value.length === 0) return

  Modal.confirm({
    title: '确认批量删除',
    content: `确定要删除选中的 ${selectedRowKeys.value.length} 个 Token 吗？`,
    icon: h(ExclamationCircleOutlined),
    onOk: async () => {
      try {
        loading.value = true
        await tokensStore.batchDeleteTokens(selectedRowKeys.value)
        selectedRowKeys.value = []
        message.success('批量删除成功')
      } catch (error) {
        message.error('批量删除失败')
      } finally {
        loading.value = false
      }
    },
  })
}

const batchValidate = async () => {
  if (selectedRowKeys.value.length === 0) return

  try {
    loading.value = true
    await tokensStore.batchValidateTokens(selectedRowKeys.value)
    message.success('批量验证完成')
  } catch (error) {
    message.error('批量验证失败')
  } finally {
    loading.value = false
  }
}

// 在IDE中打开
const openInIDE = (token: Token) => {
  selectedToken.value = token
  ideModalRef.value?.open()
}

// 关闭IDE模态框
const handleIDEModalClose = () => {
  selectedToken.value = null
}

// 批量导出
const batchExport = async () => {
  if (selectedRowKeys.value.length === 0) {
    message.warning('请选择要导出的Token')
    return
  }

  try {
    loading.value = true
    const blob = await tokensStore.exportTokens(selectedRowKeys.value.map(Number))

    // 创建下载链接
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `tokens_export_${new Date().toISOString().split('T')[0]}.json`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)

    message.success(`已导出 ${selectedRowKeys.value.length} 个Token`)
    selectedRowKeys.value = []
  } catch (error) {
    message.error('导出失败')
  } finally {
    loading.value = false
  }
}


</script>

<template>
  <div>
    <!-- 批量操作栏 -->
    <div v-if="selectedRowKeys.length > 0" class="batch-actions">
      <a-alert
        :message="`已选择 ${selectedRowKeys.length} 个 Token`"
        type="info"
        show-icon
      >
        <template #action>
          <a-space>
            <a-button size="small" @click="batchValidate" :loading="loading">
              批量验证
            </a-button>
            <a-button size="small" @click="batchExport" :loading="loading">
              <template #icon>
                <DownloadOutlined />
              </template>
              导出
            </a-button>
            <a-button size="small" danger @click="batchDelete" :loading="loading">
              批量删除
            </a-button>
          </a-space>
        </template>
      </a-alert>
    </div>





    <!-- 表格 -->
    <a-table
      :columns="columns"
      :data-source="tokens"
      :row-selection="rowSelection"
      :loading="loading"
      row-key="id"
      :pagination="{ pageSize: 10, showSizeChanger: true, showQuickJumper: true }"
      class="token-table"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'email_note'">
          <div class="email-note-container">
            <a-tooltip :title="record.email_note || '未设置'">
              <span class="email-note-text">
                {{ formatEmailNote(record.email_note) }}
              </span>
            </a-tooltip>
            <a-button
              v-if="record.email_note"
              type="text"
              size="small"
              class="copy-btn"
              @click="copyEmailNote(record.email_note)"
            >
              <template #icon>
                <CopyOutlined />
              </template>
            </a-button>
          </div>
        </template>

        <template v-else-if="column.key === 'credits'">
          <span :class="['credits-display', getCreditsClass(record)]">
            {{ getCreditsText(record) }}
          </span>
        </template>

        <template v-else-if="column.key === 'status'">
          <a-tag :color="getStatusTag(record).color">
            {{ getStatusTag(record).text }}
          </a-tag>
        </template>

        <template v-else-if="column.key === 'expiry_date'">
          <span :class="getExpiryClass(record)">
            {{ getExpiryText(record) }}
          </span>
        </template>

        <template v-else-if="column.key === 'created_at'">
          {{ formatDate(record.created_at) }}
        </template>

        <template v-else-if="column.key === 'action'">
          <a-space>
            <a-tooltip title="复制JSON">
              <a-button type="text" size="small" @click="copyTokenJSON(record)">
                <template #icon>
                  <FileTextOutlined />
                </template>
              </a-button>
            </a-tooltip>

            <a-tooltip title="打开Portal" v-if="record.portal_url">
              <a-button type="text" size="small" @click="openPortal(record)">
                <template #icon>
                  <LinkOutlined />
                </template>
              </a-button>
            </a-tooltip>

            <a-tooltip title="在IDE中打开">
              <a-button type="text" size="small" @click="openInIDE(record)">
                <template #icon>
                  <CodeOutlined />
                </template>
              </a-button>
            </a-tooltip>

            <a-tooltip title="验证并刷新">
              <a-button type="text" size="small" @click="refreshToken(record.id)" :loading="loading">
                <template #icon>
                  <ReloadOutlined />
                </template>
              </a-button>
            </a-tooltip>

            <a-tooltip title="编辑">
              <a-button type="text" size="small" @click="editToken(record.id)">
                <template #icon>
                  <EditOutlined />
                </template>
              </a-button>
            </a-tooltip>

            <a-tooltip title="删除">
              <a-button type="text" size="small" danger @click="deleteToken(record.id)">
                <template #icon>
                  <DeleteOutlined />
                </template>
              </a-button>
            </a-tooltip>
          </a-space>
        </template>
      </template>
    </a-table>

    <!-- 编辑模态框 -->
    <a-modal
      v-model:open="editModalVisible"
      title="编辑 Token"
      :confirm-loading="loading"
      @ok="saveEdit"
      @cancel="cancelEdit"
      width="600px"
    >
      <a-form
        v-if="editingToken"
        :model="editingToken"
        layout="vertical"
        class="edit-form"
      >
        <a-form-item label="tenant_url" required>
          <a-input
            v-model:value="editingToken.tenant_url"
            placeholder="请输入 tenant_url"
          />
        </a-form-item>

        <a-form-item label="access_token" required>
          <a-textarea
            v-model:value="editingToken.access_token"
            placeholder="请输入 access_token"
            :rows="3"
          />
        </a-form-item>

        <a-form-item label="portal_url">
          <a-input
            v-model:value="editingToken.portal_url"
            placeholder="请输入 portal_url"
          />
        </a-form-item>

        <a-form-item label="max_usage">
          <a-input-number
            v-model:value="editingToken.max_usage"
            placeholder="请输入 max_usage（可选）"
            :min="1"
            style="width: 100%"
          />
        </a-form-item>

        <a-form-item label="email_note（可选）">
          <a-input
            v-model:value="editingToken.email_note"
            placeholder="请输入 email_note（可选）"
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- IDE选择模态框 -->
    <IDEModal ref="ideModalRef" :token="selectedToken" @close="handleIDEModalClose" />
  </div>
</template>
<style scoped>
/* 主题变量 */
:root {
  --text-primary: #1d1d1f;
  --text-secondary: #86868b;
  --bg-card: #ffffff;
  --bg-secondary: #f6f8fa;
  --border-color: #e5e5e7;
  --table-header-bg: #fafafa;
}

:root.dark {
  --text-primary: #e8e8ed;
  --text-secondary: #98989d;
  --bg-card: #2c2c2e;
  --bg-secondary: #3a3a3c;
  --border-color: #48484a;
  --table-header-bg: #1c1c1e;
}

/* 苹果美学风格表格 */
.batch-actions {
  margin-bottom: 16px;
}

.token-table {
  background: var(--bg-card);
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  border: 1px solid var(--border-color);
  overflow: hidden;
}

/* 表格头部样式 */
.token-table :deep(.ant-table-thead > tr > th) {
  background: var(--table-header-bg) !important;
  border-bottom: 1px solid var(--border-color) !important;
  font-weight: 600;
  font-size: 13px;
  color: var(--text-primary) !important;
  padding: 16px 12px;
  letter-spacing: -0.01em;
}

/* 表格行样式 */
.token-table :deep(.ant-table-tbody > tr > td) {
  background: var(--bg-card) !important;
  border-bottom: 1px solid var(--border-color) !important;
  padding: 14px 12px;
  font-size: 14px;
  color: var(--text-primary) !important;
  line-height: 1.4;
}

/* 悬停效果 */
.token-table :deep(.ant-table-tbody > tr:hover > td) {
  background: var(--bg-secondary) !important;
}

/* 分页组件居中 */
.token-table :deep(.ant-pagination) {
  display: flex !important;
  justify-content: center !important;
  margin-top: 24px !important;
}

/* 选中行样式 */
.token-table :deep(.ant-table-tbody > tr.ant-table-row-selected > td) {
  background: rgba(0, 122, 255, 0.05);
}

/* 移除默认边框 */
.token-table :deep(.ant-table-container) {
  border: none;
}

.token-table :deep(.ant-table-content) {
  border-radius: 0;
}

/* 额度显示 - 苹果风格 */
.credits-display {
  font-weight: 600;
  padding: 4px 8px;
  border-radius: 6px;
  font-size: 12px;
  display: inline-block;
  letter-spacing: -0.01em;
}

.credits-display.exhausted {
  color: #ff453a;
  background-color: rgba(255, 69, 58, 0.1);
}

.credits-display.unlimited {
  color: #30d158;
  background-color: rgba(48, 209, 88, 0.1);
}

.credits-display.available {
  color: #007aff;
  background-color: rgba(0, 122, 255, 0.1);
}

.credits-display.unknown {
  color: #8e8e93;
  background-color: rgba(142, 142, 147, 0.1);
}

/* 状态标签 - 苹果风格 */
.status-active {
  color: #30d158;
  font-weight: 600;
  font-size: 12px;
  background: rgba(48, 209, 88, 0.1);
  padding: 4px 8px;
  border-radius: 6px;
  display: inline-block;
  letter-spacing: -0.01em;
}

.status-suspended {
  color: #ff9f0a;
  font-weight: 600;
  font-size: 12px;
  background: rgba(255, 159, 10, 0.1);
  padding: 4px 8px;
  border-radius: 6px;
  display: inline-block;
  letter-spacing: -0.01em;
}

.status-invalid {
  color: #ff453a;
  font-weight: 600;
  font-size: 12px;
  background: rgba(255, 69, 58, 0.1);
  padding: 4px 8px;
  border-radius: 6px;
  display: inline-block;
  letter-spacing: -0.01em;
}

.status-exhausted {
  color: #8e8e93;
  font-weight: 600;
  font-size: 12px;
  background: rgba(142, 142, 147, 0.1);
  padding: 4px 8px;
  border-radius: 6px;
  display: inline-block;
  letter-spacing: -0.01em;
}

.status-expired {
  color: #ff453a;
  font-weight: 600;
  font-size: 12px;
  background: rgba(255, 69, 58, 0.1);
  padding: 4px 8px;
  border-radius: 6px;
  display: inline-block;
  letter-spacing: -0.01em;
}

.status-unknown {
  color: #8e8e93;
  font-weight: 600;
  font-size: 12px;
  background: rgba(142, 142, 147, 0.1);
  padding: 4px 8px;
  border-radius: 6px;
  display: inline-block;
  letter-spacing: -0.01em;
}

.edit-form {
  margin-top: 16px;
}

.edit-form .ant-form-item {
  margin-bottom: 16px;
}

/* 邮箱备注 - 苹果风格 */
.email-note-container {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 2px 0;
}

.email-note-text {
  flex: 1;
  min-width: 0;
  color: var(--text-primary);
  font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Text', 'Helvetica Neue', sans-serif;
  font-size: 13px;
  font-weight: 400;
  letter-spacing: -0.01em;
}

.copy-btn {
  opacity: 0;
  transition: all 0.2s ease;
  padding: 4px;
  height: 24px;
  min-width: 24px;
  border-radius: 4px;
  background: rgba(0, 122, 255, 0.1);
  border: none;
  color: #007aff;
}

.copy-btn:hover {
  background: rgba(0, 122, 255, 0.15);
  transform: scale(1.05);
}

.email-note-container:hover .copy-btn {
  opacity: 1;
}

/* 过期时间样式 - 苹果风格 */
.expiry-normal {
  color: #30d158;
  font-weight: 600;
  font-size: 13px;
  letter-spacing: -0.01em;
}

.expiry-caution {
  color: #ff9f0a;
  font-weight: 600;
  font-size: 13px;
  letter-spacing: -0.01em;
}

.expiry-warning {
  color: #ff9f0a;
  font-weight: 600;
  font-size: 13px;
  letter-spacing: -0.01em;
}

.expiry-expired {
  color: #ff453a;
  font-weight: 700;
  font-size: 13px;
  letter-spacing: -0.01em;
}

.expiry-unknown {
  color: #8e8e93;
  font-weight: 500;
  font-size: 13px;
  letter-spacing: -0.01em;
}

/* 操作按钮 - 苹果风格 */
.token-table :deep(.ant-btn) {
  border-radius: 6px;
  font-weight: 500;
  font-size: 12px;
  height: 28px;
  padding: 0 8px;
  border: 1px solid rgba(0, 0, 0, 0.1);
  transition: all 0.2s ease;
  letter-spacing: -0.01em;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  line-height: 1;
}

.token-table :deep(.ant-btn .anticon) {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
}

.token-table :deep(.ant-btn-primary) {
  background: #007aff;
  border-color: #007aff;
  box-shadow: 0 1px 2px rgba(0, 122, 255, 0.2);
}

.token-table :deep(.ant-btn-primary:hover) {
  background: #0056cc;
  border-color: #0056cc;
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 122, 255, 0.3);
}

.token-table :deep(.ant-btn-default) {
  background: var(--bg-card);
  color: var(--text-primary);
  border-color: var(--border-color);
}

.token-table :deep(.ant-btn-default:hover) {
  background: var(--bg-secondary);
  border-color: var(--border-color);
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.token-table :deep(.ant-btn-dangerous) {
  color: #ff453a;
  border-color: rgba(255, 69, 58, 0.3);
}

.token-table :deep(.ant-btn-dangerous:hover) {
  background: rgba(255, 69, 58, 0.05);
  border-color: #ff453a;
  color: #ff453a;
  transform: translateY(-1px);
}

/* 按钮组间距 */
.token-table :deep(.ant-space-item) {
  margin-right: 4px !important;
}

/* 导入区域样式 */
.import-section {
  margin-bottom: 16px;
  padding: 16px;
  background: #fafafa;
  border-radius: 8px;
  border: 1px solid #f0f0f0;
}

.import-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.import-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #1d1d1f;
}

.import-section .ant-btn {
  border-radius: 6px;
  font-weight: 500;
}

/* 导入模态框样式 */
.import-dragger {
  margin: 16px 0;
}

.text-import {
  padding: 16px 0;
}

.import-textarea {
  margin-bottom: 16px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
}

.import-hint {
  background: #f6f8fa;
  border-radius: 6px;
  padding: 12px;
  border-left: 3px solid #007aff;
}

.import-hint p {
  margin: 0 0 8px 0;
  font-weight: 500;
  color: #1d1d1f;
}

.import-hint ul {
  margin: 0;
  padding-left: 20px;
}

.import-hint li {
  margin: 4px 0;
  color: #86868b;
  font-size: 13px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .import-section {
    padding: 12px;
    margin-bottom: 12px;
  }
}
</style>
