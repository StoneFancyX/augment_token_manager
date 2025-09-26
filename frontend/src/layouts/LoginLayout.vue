<template>
  <div class="login-layout">
    <!-- 主题切换按钮 -->
    <div class="theme-toggle-container">
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
    <slot/>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useThemeStore } from '../stores/theme'
import { BulbOutlined, BulbFilled } from '@ant-design/icons-vue'

const themeStore = useThemeStore()

onMounted(() => {
  themeStore.initTheme()
})
</script>

<style scoped>
.login-layout {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  transition: background-color 0.3s;
}

.theme-toggle-container {
  position: absolute;
  top: 24px;
  right: 24px;
  z-index: 10;
}

.theme-toggle-btn {
  width: 40px !important;
  height: 40px !important;
  display: flex !important;
  align-items: center !important;
  justify-content: center !important;
  border-radius: 6px !important;
  font-size: 18px !important;
  border: 1px solid !important;
  box-sizing: border-box !important;
  transition: all 0.3s !important;
  padding: 0 !important;
  min-width: 40px !important;
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

/* 浅色主题 */
:root .login-layout {
  background: #f5f5f7;
}

:root .theme-toggle-btn {
  background: #ffffff !important;
  color: #1d1d1f !important;
  border-color: #d9d9d9 !important;
}

:root .theme-toggle-btn:hover {
  background: rgba(24, 144, 255, 0.1) !important;
  border-color: #1890ff !important;
}

/* 深色主题 */
:root.dark .login-layout {
  background: #000000;
}

:root.dark .theme-toggle-btn {
  background: #2c2c2e !important;
  color: #e8e8ed !important;
  border-color: #48484a !important;
}

:root.dark .theme-toggle-btn:hover {
  background: rgba(24, 144, 255, 0.2) !important;
  border-color: #1890ff !important;
}
</style>
