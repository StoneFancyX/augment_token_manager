<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useTokensStore } from '../stores/tokens'
import { message } from 'ant-design-vue'
import type { TokenCreate } from '../types/token'

interface Props {
  showModal?: boolean
  embedded?: boolean
}

interface Emits {
  (e: 'close'): void
  (e: 'created'): void
}

const props = withDefaults(defineProps<Props>(), {
  showModal: true,
  embedded: false
})

const emit = defineEmits<Emits>()
const tokensStore = useTokensStore()

const visible = ref(props.showModal)
const loading = ref(false)

const formState = reactive<TokenCreate>({
  email_note: '',
  access_token: '',
  tenant_url: '',
  portal_url: ''
})

const rules = {
  access_token: [
    { required: true, message: '请输入 Access Token', trigger: 'blur' }
  ],
  tenant_url: [
    { required: true, message: '请输入 Tenant URL', trigger: 'blur' },
    { type: 'url', message: '请输入正确的 URL 格式', trigger: 'blur' }
  ],
  portal_url: [
    { required: true, message: '请输入 Portal URL', trigger: 'blur' },
    { type: 'url', message: '请输入正确的 URL 格式', trigger: 'blur' }
  ]
}

const onFinish = async (values: TokenCreate) => {
  try {
    loading.value = true

    await tokensStore.createToken({
      email_note: values.email_note.trim(),
      access_token: values.access_token.trim(),
      tenant_url: values.tenant_url.trim(),
      portal_url: values.portal_url.trim()
    })

    message.success('Token 创建成功')
    emit('created')
  } catch (err: any) {
    message.error(err.message || '创建失败')
  } finally {
    loading.value = false
  }
}

const onFinishFailed = (errorInfo: any) => {
  console.log('Failed:', errorInfo)
}

const handleCancel = () => {
  visible.value = false
  emit('close')
}
</script>

<template>
  <a-modal
    v-if="!embedded"
    v-model:open="visible"
    title="添加新 Token"
    :confirm-loading="loading"
    @cancel="handleCancel"
    :footer="null"
    class="token-modal"
  >
    <a-form
      :model="formState"
      :rules="rules"
      layout="vertical"
      @finish="onFinish"
      @finishFailed="onFinishFailed"
    >
      <a-form-item
        label="邮箱备注（可选）"
        name="email_note"
      >
        <a-input
          v-model:value="formState.email_note"
          placeholder="请输入邮箱备注（可选）"
        />
      </a-form-item>

      <a-form-item
        label="Access Token"
        name="access_token"
      >
        <a-textarea
          v-model:value="formState.access_token"
          placeholder="请输入 Access Token"
          :rows="3"
        />
      </a-form-item>

      <a-form-item
        label="Tenant URL"
        name="tenant_url"
      >
        <a-input
          v-model:value="formState.tenant_url"
          placeholder="https://example.com"
        />
      </a-form-item>

      <a-form-item
        label="Portal URL"
        name="portal_url"
      >
        <a-input
          v-model:value="formState.portal_url"
          placeholder="https://portal.example.com"
        />
      </a-form-item>

      <a-form-item v-if="!embedded" class="form-actions">
        <a-space>
          <a-button @click="handleCancel">
            取消
          </a-button>
          <a-button type="primary" html-type="submit" :loading="loading">
            创建
          </a-button>
        </a-space>
      </a-form-item>
    </a-form>
  </a-modal>

  <!-- 嵌入模式 -->
  <div v-if="embedded" class="embedded-form">
    <a-form
      :model="formState"
      :rules="rules"
      layout="vertical"
      @finish="onFinish"
      @finishFailed="onFinishFailed"
    >
      <a-form-item
        label="tenant_url"
        name="tenant_url"
      >
        <a-input
          v-model:value="formState.tenant_url"
          placeholder="请输入 tenant_url"
        />
      </a-form-item>

      <a-form-item
        label="access_token"
        name="access_token"
      >
        <a-textarea
          v-model:value="formState.access_token"
          placeholder="请输入 access_token"
          :rows="3"
        />
      </a-form-item>

      <a-form-item
        label="portal_url"
        name="portal_url"
      >
        <a-input
          v-model:value="formState.portal_url"
          placeholder="请输入 portal_url"
        />
      </a-form-item>

      <a-form-item
        label="email_note（可选）"
        name="email_note"
      >
        <a-input
          v-model:value="formState.email_note"
          placeholder="请输入 email_note（可选）"
        />
      </a-form-item>
    </a-form>
  </div>
</template>
<style scoped>
.token-modal {
  width: 100%;
  max-width: 672px;
}

.form-actions {
  margin-bottom: 0;
  text-align: right;
}

.embedded-form {
  padding: 16px 0;
}
</style>
