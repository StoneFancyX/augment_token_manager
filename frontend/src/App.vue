<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { RouterView } from 'vue-router'
import AppLayout from './layouts/AppLayout.vue'

const route = useRoute()

// 不需要布局的页面
const noLayoutPages = ['/login', '/404']
const needsLayout = computed(() => {
  return !noLayoutPages.includes(route.path) && !route.path.includes('/404')
})
</script>

<template>
  <div id="app">
    <AppLayout v-if="needsLayout">
      <RouterView />
    </AppLayout>
    <RouterView v-else />
  </div>
</template>


<style>
/* 全局主题变量 */
:root {
  /* 登录页面主题变量 */
  --login-bg: #f5f5f7;
  --login-card-bg: #ffffff;
  --login-text-primary: #1d1d1f;
  --login-text-secondary: #86868b;
  --login-input-bg: #fafafa;
  --login-input-border: #d2d2d7;
  --login-shadow: rgba(0, 0, 0, 0.08);
  --login-icon-color: rgba(0, 0, 0, 0.25);
}

:root.dark {
  /* 登录页面深色主题变量 */
  --login-bg: #000000;
  --login-card-bg: #1c1c1e;
  --login-text-primary: #ffffff;
  --login-text-secondary: #a1a1a6;
  --login-input-bg: #2c2c2e;
  --login-input-border: #3a3a3c;
  --login-shadow: rgba(0, 0, 0, 0.6);
  --login-icon-color: rgba(255, 255, 255, 0.6);
}

html, body, #app { height: 100%; }
body { margin: 0; overflow-x: hidden; }
</style>
