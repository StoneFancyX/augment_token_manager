<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { message } from 'ant-design-vue'
import { UserOutlined, LockOutlined } from '@ant-design/icons-vue'
import LoginLayout from '../layouts/LoginLayout.vue'

const router = useRouter()
const authStore = useAuthStore()

const loading = ref(false)

const formState = reactive({
  username: '',
  password: ''
})

const onFinish = async (values: any) => {
  loading.value = true

  try {
    await authStore.login({
      username: values.username,
      password: values.password
    })

    message.success('登录成功')
    router.push('/tokens')
  } catch (err: any) {
    message.error(err.response?.data?.detail || '登录失败，请检查用户名和密码')
  } finally {
    loading.value = false
  }
}

const onFinishFailed = (errorInfo: any) => {
  console.log('Failed:', errorInfo)
}
</script>

<template>
  <LoginLayout>
    <div class="login-container">
      <div class="login-card">
        <div class="login-header">
          <div class="login-icon">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none">
              <rect x="3" y="4" width="18" height="16" rx="2" stroke="currentColor" stroke-width="2"/>
              <path d="M7 8h10M7 12h6" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
            </svg>
          </div>
          <h1 class="login-title">
            ATM
          </h1>
          <p class="login-subtitle">
            请使用您的管理员账户登录
          </p>
        </div>

      <a-form
        :model="formState"
        name="login"
        autocomplete="off"
        @finish="onFinish"
        @finishFailed="onFinishFailed"
      >
        <a-form-item
          name="username"
          :rules="[{ required: true, message: '请输入用户名!' }]"
        >
          <a-input
            v-model:value="formState.username"
            placeholder="用户名"
            size="large"
          >
            <template #prefix>
              <UserOutlined class="input-icon" />
            </template>
          </a-input>
        </a-form-item>

        <a-form-item
          name="password"
          :rules="[{ required: true, message: '请输入密码!' }]"
        >
          <a-input-password
            v-model:value="formState.password"
            placeholder="密码"
            size="large"
          >
            <template #prefix>
              <LockOutlined class="input-icon" />
            </template>
          </a-input-password>
        </a-form-item>

        <a-form-item>
          <a-button
            type="primary"
            html-type="submit"
            size="large"
            :loading="loading"
            class="login-button"
          >
            {{ loading ? '登录中...' : '登录' }}
          </a-button>
        </a-form-item>
      </a-form>
      </div>
    </div>
  </LoginLayout>
</template>
<style scoped>
/* 简约苹果风格登录界面 */
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  padding: 20px;
}

.login-card {
  width: 380px;
  max-width: 100%;
  background: var(--login-card-bg);
  border-radius: 16px;
  padding: 48px 40px;
  transition: all 0.3s;
}

/* 浅色主题下的卡片 */
:root .login-card {
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.08);
  border: 1px solid rgba(0, 0, 0, 0.05);
}

/* 深色主题下的卡片 */
:root.dark .login-card {
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.8), 0 0 0 1px rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.15);
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.login-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  background: #007aff;
  border-radius: 12px;
  color: white;
  margin: 0 auto 20px;
  transition: all 0.3s;
}

/* 深色主题下的图标增强 */
:root.dark .login-icon {
  background: #0a84ff;
  box-shadow: 0 0 20px rgba(10, 132, 255, 0.3);
}

.login-title {
  font-size: 24px;
  font-weight: 600;
  color: var(--login-text-primary);
  margin: 0 0 6px 0;
  font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', sans-serif;
  transition: color 0.3s;
}

.login-subtitle {
  font-size: 15px;
  color: var(--login-text-secondary);
  margin: 0;
  font-weight: 400;
  transition: color 0.3s;
}

/* 表单样式 */
.login-card :deep(.ant-form-item) {
  margin-bottom: 20px;
}

.login-card :deep(.ant-input-affix-wrapper),
.login-card :deep(.ant-input) {
  border-radius: 8px !important;
  border: 1px solid var(--login-input-border) !important;
  background: var(--login-input-bg) !important;
  color: var(--login-text-primary) !important;
  padding: 12px 16px !important;
  font-size: 16px !important;
  transition: all 0.15s ease !important;
  box-shadow: none !important;
}

.login-card :deep(.ant-input-affix-wrapper:hover),
.login-card :deep(.ant-input:hover) {
  border-color: #007aff !important;
  background: var(--login-card-bg) !important;
  box-shadow: none !important;
}

.login-card :deep(.ant-input-affix-wrapper:focus),
.login-card :deep(.ant-input-affix-wrapper-focused),
.login-card :deep(.ant-input:focus) {
  border-color: #007aff !important;
  background: var(--login-card-bg) !important;
  box-shadow: 0 0 0 2px rgba(0, 122, 255, 0.15) !important;
  outline: none !important;
}

.login-card :deep(.ant-input-affix-wrapper .ant-input) {
  border: none !important;
  box-shadow: none !important;
  background: transparent !important;
  padding: 0 !important;
}

.login-card :deep(.ant-input-prefix) {
  margin-right: 12px;
  color: var(--login-icon-color);
  transition: color 0.3s;
}

.input-icon {
  color: var(--login-icon-color);
  transition: color 0.3s;
}

/* 登录按钮 */
.login-button {
  width: 100%;
  height: 44px;
  border-radius: 8px;
  background: #007aff;
  border: none;
  font-size: 16px;
  font-weight: 500;
  transition: all 0.15s ease;
}

.login-button:hover {
  background: #0056cc;
}

.login-button:active {
  background: #004bb8;
}

/* 响应式设计 */
@media (max-width: 480px) {
  .login-container {
    padding: 16px;
  }

  .login-card {
    padding: 32px 24px;
    border-radius: 12px;
  }

  .login-icon {
    width: 48px;
    height: 48px;
  }

  .login-title {
    font-size: 22px;
  }

  .login-subtitle {
    font-size: 14px;
  }
}
</style>
