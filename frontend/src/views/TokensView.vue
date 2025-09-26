<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useTokensStore } from '../stores/tokens'
import { PlusOutlined, ReloadOutlined, UploadOutlined, DownloadOutlined, TableOutlined, AppstoreOutlined, EditOutlined, DeleteOutlined, FileTextOutlined, CodeOutlined, LinkOutlined } from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import TokenTable from '../components/TokenTable.vue'
import TokenModal from '../components/TokenModal.vue'
import IDEModal from '../components/IDEModal.vue'
import { tokensApi } from '../api/tokens'

const tokensStore = useTokensStore()
const showAddModal = ref(false)
const addType = ref('manual')
const importText = ref('')

// 视图模式
const viewMode = ref<'table' | 'card'>('table')

// 统计数据
const stats = ref({
  total_tokens: 0,
  total_credits: 0,
  available_credits: 0,
  unlimited_tokens: 0,
  expired_tokens: 0,
  valid_tokens: 0
})

// IDE相关
const ideModalRef = ref()
const selectedToken = ref(null)

// 编辑相关
const editModalVisible = ref(false)
const editingToken = ref(null)
const loading = ref(false)

// 获取状态标签（按照0.6.0版本逻辑）
const getStatusTag = (token: any) => {
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

// 获取额度显示文本（与TokenTable完全一致的逻辑）
const getCreditsText = (token: any) => {
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

// 获取额度显示样式类（与TokenTable完全一致的逻辑）
const getCreditsClass = (token: any) => {
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

// 获取过期时间显示文本（与TokenTable完全一致的逻辑）
const getExpiryText = (token: any) => {
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

// 获取过期时间样式类（与TokenTable完全一致的逻辑）
const getExpiryClass = (token: any) => {
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

// 格式化邮箱显示（与TokenTable完全一致的逻辑）
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


const handleAddModalClose = () => {
  showAddModal.value = false
  addType.value = 'manual'
  importText.value = ''
}

const handleTokenCreated = () => {
  showAddModal.value = false
  addType.value = 'manual'
  // 刷新token列表
  tokensStore.fetchTokens()
}

const handleBatchRefresh = async () => {
  try {
    const result = await tokensStore.batchRefreshTokens()
    message.success(result.message)
  } catch (error) {
    message.error('批量刷新失败')
  }
}

// 下载模板
const downloadTemplate = () => {
  const template = [
    {
      "tenant_url": "https://example.com",
      "access_token": "your_access_token_here",
      "portal_url": "https://portal.example.com",
      "email_note": "user@example.com"
    }
  ]

  const blob = new Blob([JSON.stringify(template, null, 2)], { type: 'application/json' })
  const url = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = 'token_template.json'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  window.URL.revokeObjectURL(url)

  message.success('模板文件已下载')
}

// 处理文件导入
const handleFileImport = async (file: File) => {
  try {
    const text = await file.text()
    importText.value = text
    message.success('文件内容已加载，请点击确定导入')
  } catch (error) {
    message.error('文件读取失败')
  }

  // 阻止默认上传行为
  return false
}

// 确认添加操作
const handleAddConfirm = async () => {
  if (addType.value === 'manual') {
    // 手动添加模式，触发嵌入表单的提交
    const embeddedForm = document.querySelector('.embedded-form form')
    if (embeddedForm) {
      const submitEvent = new Event('submit', { bubbles: true, cancelable: true })
      embeddedForm.dispatchEvent(submitEvent)
    }
  } else if (addType.value === 'file' || addType.value === 'text') {
    if (!importText.value.trim()) {
      message.warning('请输入或选择要导入的数据')
      return
    }

    try {
      // 验证JSON格式
      const data = JSON.parse(importText.value)
      if (!Array.isArray(data)) {
        throw new Error('数据格式错误，应为数组格式')
      }

      // 创建临时文件进行导入
      const blob = new Blob([importText.value], { type: 'application/json' })
      const file = new File([blob], 'import.json', { type: 'application/json' })

      const result = await tokensStore.importTokens(file)

      message.success(result.message)

      // 刷新列表
      await tokensStore.fetchTokens()

      // 关闭模态框
      handleAddModalClose()

    } catch (error: any) {
      if (error instanceof SyntaxError) {
        message.error('JSON格式错误，请检查数据格式')
      } else {
        message.error(error.message || '导入失败')
      }
    }
  } else if (addType.value === 'template') {
    // 模板下载，直接关闭模态框
    handleAddModalClose()
  }
}

// 卡片操作函数
const handleEdit = (token: any) => {
  editingToken.value = { ...token }
  editModalVisible.value = true
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

const handleCopyJSON = async (token: any) => {
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
    message.error('复制失败')
  }
}

const handleOpenPortal = (token: any) => {
  if (token.portal_url) {
    window.open(token.portal_url, '_blank')
  } else {
    message.warning('该Token没有Portal URL')
  }
}

const handleOpenIDE = (token: any) => {
  selectedToken.value = token
  ideModalRef.value?.open()
}

// 关闭IDE模态框
const handleIDEModalClose = () => {
  selectedToken.value = null
}

const handleRefresh = async (token: any) => {
  try {
    // 按照0.6.0版本逻辑：先验证状态，然后刷新Portal信息
    const result = await tokensStore.validateToken(token.id)

    // 如果有portal_url，同时刷新Portal信息
    if (token.portal_url) {
      try {
        await tokensStore.refreshToken(token.id)
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
  }
}

const handleDelete = (token: any) => {
  // TODO: 实现删除功能
  message.info('删除功能开发中...')
}

// 获取统计数据
const fetchStats = async () => {
  try {
    const data = await tokensApi.getTokenStats()
    console.log('统计数据:', data)
    stats.value = data
  } catch (error) {
    console.error('获取统计数据失败:', error)
  }
}

onMounted(() => {
  tokensStore.fetchTokens()
  fetchStats()
})
</script>

<template>
  <div>
    <!-- 页面标题和操作 -->
    <div class="page-header">
      <div>
        <h1 class="page-title">Token 管理</h1>
        <p class="page-description">
          管理您的 Augment Token，包括创建、编辑、验证和删除操作。
        </p>
      </div>
      <a-space>
        <!-- 视图切换 -->
        <a-button-group>
          <a-button
            :type="viewMode === 'table' ? 'primary' : 'default'"
            @click="viewMode = 'table'"
          >
            <template #icon>
              <TableOutlined />
            </template>
            表格
          </a-button>
          <a-button
            :type="viewMode === 'card' ? 'primary' : 'default'"
            @click="viewMode = 'card'"
          >
            <template #icon>
              <AppstoreOutlined />
            </template>
            卡片
          </a-button>
        </a-button-group>

        <a-button @click="handleBatchRefresh" :loading="tokensStore.isLoading">
          <template #icon>
            <ReloadOutlined />
          </template>
          批量刷新
        </a-button>
        <a-button type="primary" @click="showAddModal = true">
          <template #icon>
            <PlusOutlined />
          </template>
          添加
        </a-button>
      </a-space>
    </div>



    <!-- 统计信息 -->
    <a-row :gutter="20" class="stats-section">
      <a-col :span="6">
        <a-card class="stats-card stats-card-credits">
          <a-statistic
            title="总额度"
            :value="stats.total_credits"
            suffix="Credits"
            :value-style="{ color: '#ff6b35', fontWeight: '600', fontSize: '24px' }"
            :title-style="{ color: '#8e8e93', fontSize: '13px', fontWeight: '500' }"
          />
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card class="stats-card stats-card-total">
          <a-statistic
            title="总计"
            :value="stats.total_tokens"
            suffix="个 Token"
            :value-style="{ color: '#007aff', fontWeight: '600', fontSize: '24px' }"
            :title-style="{ color: '#8e8e93', fontSize: '13px', fontWeight: '500' }"
          />
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card class="stats-card stats-card-valid">
          <a-statistic
            title="有效"
            :value="stats.valid_tokens"
            suffix="个有效"
            :value-style="{ color: '#30d158', fontWeight: '600', fontSize: '24px' }"
            :title-style="{ color: '#8e8e93', fontSize: '13px', fontWeight: '500' }"
          />
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card class="stats-card stats-card-invalid">
          <a-statistic
            title="无效"
            :value="stats.total_tokens - stats.valid_tokens"
            suffix="个无效"
            :value-style="{ color: '#ff453a', fontWeight: '600', fontSize: '24px' }"
            :title-style="{ color: '#8e8e93', fontSize: '13px', fontWeight: '500' }"
          />
        </a-card>
      </a-col>
    </a-row>

    <!-- Token 列表 -->
    <div v-if="viewMode === 'table'">
      <TokenTable :tokens="tokensStore.tokens" />
    </div>

    <div v-else-if="viewMode === 'card'" class="token-cards">
      <a-row :gutter="[16, 16]">
        <a-col
          v-for="token in tokensStore.tokens"
          :key="token.id"
          :xs="24"
          :sm="12"
          :md="8"
          :lg="6"
        >
          <div
            class="token-card"
            :class="{
              'token-card-valid': getStatusTag(token).color === 'green',
              'token-card-invalid': getStatusTag(token).color === 'red',
              'token-card-expired': getStatusTag(token).color === 'orange'
            }"
          >
            <!-- 头部信息 -->
            <div class="token-card-header">
              <div class="token-title">
                <h4 class="token-email" :title="token.email_note || '未设置邮箱'">{{ formatEmailNote(token.email_note) }}</h4>
                <a-tag
                  :color="getStatusTag(token).color"
                  class="token-status-tag"
                >
                  {{ getStatusTag(token).text }}
                </a-tag>
              </div>
            </div>

            <!-- 主要信息 -->
            <div class="token-main-info">
              <div class="info-row">
                <span class="info-value tenant-url">{{ token.tenant_url }}</span>
              </div>

              <div class="info-row">
                <span class="info-label">剩余时间</span>
                <span class="info-value expiry-display" :class="getExpiryClass(token)">
                  {{ getExpiryText(token) }}
                </span>
              </div>

              <div class="info-row">
                <span class="info-label">剩余额度</span>
                <span class="info-value credits-display" :class="getCreditsClass(token)">
                  {{ getCreditsText(token) }}
                </span>
              </div>
            </div>

            <!-- 底部操作 -->
            <div class="token-card-actions">
              <div class="action-buttons">
                <a-tooltip title="复制JSON">
                  <a-button size="small" type="text" class="action-btn" @click="handleCopyJSON(token)">
                    <FileTextOutlined />
                  </a-button>
                </a-tooltip>

                <a-tooltip title="打开Portal">
                  <a-button size="small" type="text" class="action-btn" @click="handleOpenPortal(token)" v-if="token.portal_url">
                    <LinkOutlined />
                  </a-button>
                </a-tooltip>

                <a-tooltip title="在IDE中打开">
                  <a-button size="small" type="text" class="action-btn" @click="handleOpenIDE(token)">
                    <CodeOutlined />
                  </a-button>
                </a-tooltip>

                <a-tooltip title="验证并刷新">
                  <a-button size="small" type="text" class="action-btn" @click="handleRefresh(token)">
                    <ReloadOutlined />
                  </a-button>
                </a-tooltip>

                <a-tooltip title="编辑">
                  <a-button size="small" type="text" class="action-btn" @click="handleEdit(token)">
                    <EditOutlined />
                  </a-button>
                </a-tooltip>

                <a-tooltip title="删除">
                  <a-button size="small" type="text" danger class="action-btn" @click="handleDelete(token)">
                    <DeleteOutlined />
                  </a-button>
                </a-tooltip>
              </div>
              <span class="create-time">{{ new Date(token.created_at).toLocaleDateString() }}</span>
            </div>
          </div>
        </a-col>
      </a-row>
    </div>

    <!-- 添加Token模态框 -->
    <a-modal
      v-model:open="showAddModal"
      title="添加Token"
      width="600px"
      @ok="handleAddConfirm"
      @cancel="handleAddModalClose"
      :confirm-loading="tokensStore.isLoading"
      :ok-text="addType === 'manual' ? '创建' : addType === 'template' ? '关闭' : '导入'"
      :footer="addType === 'template' ? null : undefined"
    >
      <a-tabs v-model:activeKey="addType">
        <a-tab-pane key="manual" tab="手动添加">
          <TokenModal
            :show-modal="false"
            @created="handleTokenCreated"
            :embedded="true"
          />
        </a-tab-pane>

        <a-tab-pane key="file" tab="文件导入">
          <a-upload-dragger
            :before-upload="handleFileImport"
            :show-upload-list="false"
            accept=".json"
            class="import-dragger"
          >
            <p class="ant-upload-drag-icon">
              <UploadOutlined />
            </p>
            <p class="ant-upload-text">点击或拖拽JSON文件到此区域</p>
            <p class="ant-upload-hint">支持单个JSON文件上传</p>
          </a-upload-dragger>
        </a-tab-pane>

        <a-tab-pane key="text" tab="文本导入">
          <div class="text-import">
            <a-textarea
              v-model:value="importText"
              placeholder="请粘贴JSON格式的Token数据..."
              :rows="10"
              class="import-textarea"
            />
            <div class="import-hint">
              <p>请确保JSON格式正确，包含以下字段：</p>
              <ul>
                <li>tenant_url - 租户URL（必需）</li>
                <li>access_token - 访问令牌（必需）</li>
                <li>portal_url - 门户URL（必需）</li>
                <li>email_note - 邮箱备注（可选）</li>
              </ul>
            </div>
          </div>
        </a-tab-pane>

        <a-tab-pane key="template" tab="下载模板">
          <div class="template-download">
            <div class="template-info">
              <h4>JSON模板说明</h4>
              <p>下载标准的Token数据模板，包含所有必需字段的示例格式。</p>
            </div>
            <a-button type="primary" @click="downloadTemplate" block>
              <template #icon>
                <DownloadOutlined />
              </template>
              下载JSON模板
            </a-button>
          </div>
        </a-tab-pane>
      </a-tabs>
    </a-modal>

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

    <!-- IDE模态框 -->
    <IDEModal
      ref="ideModalRef"
      :token="selectedToken"
      @close="handleIDEModalClose"
    />
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

  /* 卡片主题变量 */
  --card-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  --card-shadow-hover: 0 8px 24px rgba(0, 0, 0, 0.15);
  --card-border: #e1e1e1;
}

:root.dark {
  --text-primary: #e8e8ed;
  --text-secondary: #98989d;
  --bg-card: #2c2c2e;
  --bg-secondary: #3a3a3c;
  --border-color: #48484a;
  --table-header-bg: #1c1c1e;

  /* 卡片深色主题变量 */
  --card-shadow: 0 4px 16px rgba(0, 0, 0, 0.3);
  --card-shadow-hover: 0 12px 32px rgba(0, 0, 0, 0.5);
  --card-border: rgba(255, 255, 255, 0.15);
}

/* 页面布局 */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
}

.page-title {
  font-size: 28px;
  font-weight: 700;
  margin: 0 0 8px 0;
  color: var(--text-primary);
  letter-spacing: -0.02em;
  font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', 'Helvetica Neue', sans-serif;
}

.page-description {
  color: var(--text-secondary);
  margin: 0;
  font-size: 16px;
  font-weight: 400;
  letter-spacing: -0.01em;
}

/* 统计卡片区域 */
.stats-section {
  margin-bottom: 32px;
}

/* 统计卡片 - 苹果风格 */
.stats-card {
  border-radius: 16px;
  box-shadow: var(--card-shadow);
  background: var(--bg-card);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
  user-select: none;
  -webkit-user-select: none;
  -moz-user-select: none;
  -ms-user-select: none;
  position: relative;
}

.stats-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--card-shadow-hover);
}

.stats-card :deep(.ant-card-body) {
  padding: 20px;
}

.stats-card :deep(.ant-statistic-title) {
  font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Text', 'Helvetica Neue', sans-serif;
  letter-spacing: -0.01em;
  margin-bottom: 8px;
}

.stats-card :deep(.ant-statistic-content) {
  font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', 'Helvetica Neue', sans-serif;
  letter-spacing: -0.02em;
}

/* 卡片特定样式 - 简约设计 */
.stats-card-credits {
  background: var(--bg-card);
  border: 2px solid #ff6b35;
}

.stats-card-credits:hover {
  border-color: #e55a2b;
  background: var(--bg-secondary);
}

.stats-card-total {
  background: var(--bg-card);
  border: 2px solid #007aff;
}

.stats-card-total:hover {
  border-color: #0056cc;
  background: var(--bg-secondary);
}

.stats-card-valid {
  background: var(--bg-card);
  border: 2px solid #30d158;
}

.stats-card-valid:hover {
  border-color: #28a745;
  background: var(--bg-secondary);
}

.stats-card-invalid {
  background: var(--bg-card);
  border: 2px solid #ff453a;
}

.stats-card-invalid:hover {
  border-color: #dc3545;
  background: var(--bg-secondary);
}

/* 深色主题下的卡片 - 简约设计 */
:root.dark .stats-card-credits {
  background: var(--bg-card);
  border: 2px solid #ff7b47;
}

:root.dark .stats-card-credits:hover {
  border-color: #ff9466;
  background: var(--bg-secondary);
}

:root.dark .stats-card-total {
  background: var(--bg-card);
  border: 2px solid #0a84ff;
}

:root.dark .stats-card-total:hover {
  border-color: #409cff;
  background: var(--bg-secondary);
}

:root.dark .stats-card-valid {
  background: var(--bg-card);
  border: 2px solid #32d74b;
}

:root.dark .stats-card-valid:hover {
  border-color: #64d86b;
  background: var(--bg-secondary);
}

:root.dark .stats-card-invalid {
  background: var(--bg-card);
  border: 2px solid #ff6961;
}

:root.dark .stats-card-invalid:hover {
  border-color: #ff8a80;
  background: var(--bg-secondary);
}

/* Token 卡片样式 */
.token-cards {
  margin-top: 24px;
}

.token-card {
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  background: #ffffff;
  transition: all 0.2s ease;
  border: 1px solid #e5e7eb;
  height: 100%;
  overflow: hidden;
  padding: 20px;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
}

.token-card:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  border-color: #d1d5db;
}

/* 深色主题下的卡片 */
:root.dark .token-card {
  background: #2c2c2e;
  border-color: #48484a;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
}

:root.dark .token-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.5);
  border-color: #6b7280;
}

.token-card-header {
  margin-bottom: 16px;
}

.token-title {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}

.token-email {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
  flex: 1;
  line-height: 1.4;
  word-break: break-word;
}

.token-status-tag {
  font-size: 11px;
  padding: 4px 8px;
  border-radius: 6px;
  font-weight: 500;
  flex-shrink: 0;
}

.token-main-info {
  flex: 1;
  margin-bottom: 16px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  gap: 12px;
}

.info-row:last-child {
  margin-bottom: 0;
}

.info-label {
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 500;
  min-width: 60px;
  flex-shrink: 0;
}

.info-value {
  color: var(--text-primary);
  font-size: 13px;
  text-align: right;
  word-break: break-all;
  font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Text', sans-serif;
}

.tenant-url {
  font-family: 'SF Mono', Monaco, 'Cascadia Code', 'Roboto Mono', Consolas, 'Courier New', monospace;
  font-size: 12px;
  color: var(--text-secondary);
  text-align: left;
  background: #f8f9fa;
  padding: 10px 12px;
  border-radius: 8px;
  margin-bottom: 12px;
  border: 1px solid #e9ecef;
  word-break: break-all;
}

/* 深色主题下的tenant-url */
:root.dark .tenant-url {
  background: #3a3a3c;
  border-color: #48484a;
}

.token-card-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 12px;
  border-top: 1px solid #e5e7eb;
  margin-top: auto;
}

.action-buttons {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

:root.dark .token-card-actions {
  border-top-color: #48484a;
}

.action-btn {
  padding: 6px;
  height: 28px;
  width: 28px;
  font-size: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 28px;
  border-radius: 6px;
  transition: all 0.2s ease;
}

.action-btn:hover {
  background: #f3f4f6;
}

:root.dark .action-btn:hover {
  background: #48484a;
}

.create-time {
  color: var(--text-secondary);
  font-size: 11px;
  margin-left: auto;
}

/* 额度显示样式 */
.credits-display {
  font-weight: 500;
}

.credits-display.available {
  color: #52c41a;
}

.credits-display.unlimited {
  color: #1890ff;
}

.credits-display.exhausted {
  color: #ff4d4f;
}

.credits-display.unknown {
  color: var(--text-secondary);
}

/* 剩余时间显示样式 */
.expiry-display {
  font-weight: 500;
}

.expiry-display.expiry-normal {
  color: #52c41a;
}

.expiry-display.expiry-caution {
  color: #faad14;
}

.expiry-display.expiry-warning {
  color: #ff7a45;
}

.expiry-display.expiry-expired {
  color: #ff4d4f;
}

.expiry-display.expiry-unknown {
  color: var(--text-secondary);
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
  background: var(--bg-secondary);
  border-radius: 6px;
  padding: 12px;
  border-left: 3px solid #007aff;
}

.import-hint p {
  margin: 0 0 8px 0;
  font-weight: 500;
  color: var(--text-primary);
}

.import-hint ul {
  margin: 0;
  padding-left: 20px;
}

.import-hint li {
  margin: 4px 0;
  color: var(--text-secondary);
  font-size: 13px;
}

/* 模板下载区域样式 */
.template-download {
  text-align: center;
  padding: 32px 16px;
}

.template-info {
  margin-bottom: 24px;
}

.template-info h4 {
  margin: 0 0 8px 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.template-info p {
  margin: 0;
  color: var(--text-secondary);
  font-size: 14px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }

  .stats-section {
    margin-bottom: 24px;
  }

  .stats-card {
    margin-bottom: 12px;
  }
}
</style>
