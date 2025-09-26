<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useThemeStore } from '../stores/theme'
import { theme } from 'ant-design-vue'
import {
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  FileTextOutlined,
  BulbOutlined,
  BulbFilled,
  UserOutlined,
  LogoutOutlined,
  MoreOutlined
} from '@ant-design/icons-vue'

const authStore = useAuthStore()
const themeStore = useThemeStore()
const router = useRouter()
const route = useRoute()
const collapsed = ref(false)

const selectedKeys = computed(() => {
  if (route.path.includes('/tokens')) return ['tokens']
  return ['tokens']
})

const pageTitle = computed(() => {
  if (route.path.includes('/tokens')) return 'Token 管理'
  return 'Token Manager'
})

const themeConfig = computed(() => ({
  algorithm: themeStore.isDark ? theme.darkAlgorithm : theme.defaultAlgorithm,
  token: {
    colorPrimary: '#1890ff',
  }
}))

const handleMenuClick = ({ key }: { key: string }) => {
  if (key === 'tokens') {
    router.push('/tokens')
  }
}

const handleLogout = async () => {
  await authStore.logout()
  router.push('/login')
}

onMounted(() => {
  themeStore.initTheme()
})
</script>

<template>
  <a-config-provider :theme="themeConfig">
    <a-layout style="min-height: 100vh">
      <!-- 侧边栏 -->
      <a-layout-sider
        v-model:collapsed="collapsed"
        :trigger="null"
        collapsible
        :theme="themeStore.isDark ? 'dark' : 'light'"
        class="custom-sider"
      >
        <div class="sider-content">
          <div class="logo">
            <h1 class="logo-text">
              ATM
            </h1>
          </div>

          <a-menu
            v-model:selectedKeys="selectedKeys"
            :theme="themeStore.isDark ? 'dark' : 'light'"
            mode="inline"
            @click="handleMenuClick"
            class="main-menu"
          >
            <a-menu-item key="tokens">
              <template #icon>
                <FileTextOutlined />
              </template>
              Token 管理
            </a-menu-item>
          </a-menu>

          <!-- 用户信息区域 -->
          <div class="user-section">
            <a-dropdown placement="topRight" :trigger="['click']">
              <div class="user-info-clickable">
                <div class="user-avatar">
                  <UserOutlined />
                </div>
                <div v-if="!collapsed" class="user-details">
                  <div class="username">{{ authStore.user?.username }}</div>
                  <div class="user-role">管理员</div>
                </div>
                <div class="user-menu-trigger">
                  <MoreOutlined v-if="!collapsed" />
                </div>
              </div>
              <template #overlay>
                <a-menu>
                  <a-menu-item key="logout" @click="handleLogout">
                    <template #icon>
                      <LogoutOutlined />
                    </template>
                    退出登录
                  </a-menu-item>
                </a-menu>
              </template>
            </a-dropdown>
          </div>
        </div>
      </a-layout-sider>

    <a-layout>
      <!-- 顶部导航栏 -->
      <a-layout-header class="main-header">
        <div class="header-content">
          <div class="header-left">
            <MenuUnfoldOutlined
              v-if="collapsed"
              class="trigger"
              @click="() => (collapsed = !collapsed)"
            />
            <MenuFoldOutlined
              v-else
              class="trigger"
              @click="() => (collapsed = !collapsed)"
            />
            <h2 class="page-title-header">{{ pageTitle }}</h2>
          </div>

          <div class="header-right">
            <a-button
              type="default"
              @click="themeStore.toggleTheme"
              class="theme-toggle-btn"
              :title="themeStore.isDark ? '切换到浅色模式' : '切换到深色模式'"
            >
              <template #icon>
                <BulbFilled v-if="themeStore.isDark" />
                <BulbOutlined v-else />
              </template>
            </a-button>
          </div>
        </div>
      </a-layout-header>

      <!-- 页面内容 -->
      <a-layout-content class="main-content">
        <slot />
      </a-layout-content>
    </a-layout>
  </a-layout>
  </a-config-provider>
</template>

<style scoped>
.trigger {
  font-size: 18px;
  line-height: 64px;
  padding: 0 24px;
  cursor: pointer;
  transition: color 0.3s;
}

.logo {
  height: 40px;
  margin: 12px 16px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s;
}

.logo-text {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  text-align: center;
  transition: color 0.3s;
}

/* 浅色主题下的logo */
:root .custom-sider .logo {
  background: rgba(0, 0, 0, 0.05);
}

:root .custom-sider .logo-text {
  color: #1d1d1f;
}

/* 深色主题下的logo */
:root.dark .custom-sider .logo {
  background: rgba(255, 255, 255, 0.1);
}

:root.dark .custom-sider .logo-text {
  color: #e8e8ed;
}

.custom-sider {
  position: relative;
}

.sider-content {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.main-menu {
  flex: 1;
  border-right: none;
}

.theme-toggle-btn {
  width: 40px !important;
  height: 40px !important;
  display: flex !important;
  align-items: center !important;
  justify-content: center !important;
  border-radius: 6px !important;
  font-size: 18px !important;
  background: transparent !important;
  border: 1px solid #d9d9d9 !important;
  box-sizing: border-box !important;
  transition: all 0.3s !important;
  padding: 0 !important;
  min-width: 40px !important;
}

.theme-toggle-btn:hover {
  background: rgba(24, 144, 255, 0.1) !important;
  border-color: #1890ff !important;
}

.theme-toggle-btn .anticon {
  font-size: 18px !important;
}

.theme-toggle-btn .anticon-bulb-filled {
  color: #fadb14 !important;
}

.theme-toggle-btn .anticon-bulb {
  color: #666 !important;
}

.user-section {
  padding: 16px;
  margin-top: auto;
  transition: all 0.3s;
}

.user-info-clickable {
  display: flex;
  align-items: center;
  cursor: pointer;
  padding: 8px;
  border-radius: 6px;
  transition: all 0.3s;
}

/* 折叠状态下的用户信息居中 */
.custom-sider.ant-layout-sider-collapsed .user-info-clickable {
  justify-content: center;
  padding: 8px 4px;
}

.custom-sider.ant-layout-sider-collapsed .user-avatar {
  margin-right: 0;
}

.user-avatar {
  width: 32px;
  height: 32px;
  min-width: 32px;
  min-height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 12px;
  font-size: 16px;
  transition: all 0.3s;
  flex-shrink: 0;
}

.user-details {
  flex: 1;
}

.username {
  font-size: 14px;
  font-weight: 500;
  line-height: 1.4;
  transition: color 0.3s;
}

.user-role {
  font-size: 12px;
  line-height: 1.4;
  transition: color 0.3s;
}

.user-menu-trigger {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  font-size: 12px;
  transition: color 0.3s;
}

/* 浅色主题下的用户区域 */
:root .custom-sider .user-section {
  border-top: 1px solid rgba(0, 0, 0, 0.1);
}

:root .custom-sider .user-info-clickable {
  color: #1d1d1f;
}

:root .custom-sider .user-info-clickable:hover {
  background: rgba(0, 0, 0, 0.05);
}

:root .custom-sider .user-avatar {
  background: rgba(0, 0, 0, 0.1);
  color: #1d1d1f;
}

:root .custom-sider .user-role {
  color: #86868b;
}

:root .custom-sider .user-menu-trigger {
  color: #86868b;
}

/* 深色主题下的用户区域 */
:root.dark .custom-sider .user-section {
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

:root.dark .custom-sider .user-info-clickable {
  color: rgba(255, 255, 255, 0.85);
}

:root.dark .custom-sider .user-info-clickable:hover {
  background: rgba(255, 255, 255, 0.1);
}

:root.dark .custom-sider .user-avatar {
  background: rgba(255, 255, 255, 0.1);
  color: rgba(255, 255, 255, 0.85);
}

:root.dark .custom-sider .user-role {
  color: rgba(255, 255, 255, 0.6);
}

:root.dark .custom-sider .user-menu-trigger {
  color: rgba(255, 255, 255, 0.6);
}

.main-content {
  margin: 24px;
  padding: 24px;
  border-radius: 6px;
  background: transparent;
}

/* 顶部导航栏样式 */
.main-header {
  padding: 0;
  transition: all 0.3s;
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  height: 100%;
}

.header-left {
  display: flex;
  align-items: center;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
  min-width: 60px;
}

.page-title-header {
  margin: 0 0 0 16px;
  font-size: 18px;
  font-weight: 500;
  transition: color 0.3s;
}

/* 浅色主题下的顶部导航栏 */
:root .main-header {
  background: #ffffff;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
  border-bottom: 1px solid #f0f0f0;
}

:root .page-title-header {
  color: #1d1d1f;
}

:root .trigger {
  color: #1d1d1f;
}

:root .trigger:hover {
  color: #1890ff;
}

/* 深色主题下的顶部导航栏 */
:root.dark .main-header {
  background: #2c2c2e;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.2);
  border-bottom: 1px solid #48484a;
}

:root.dark .page-title-header {
  color: #e8e8ed;
}

:root.dark .trigger {
  color: #e8e8ed;
}

:root.dark .trigger:hover {
  color: #1890ff;
}
</style>
