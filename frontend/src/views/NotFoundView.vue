<script setup lang="ts">
import { useRouter } from 'vue-router'
import { HomeOutlined } from '@ant-design/icons-vue'
import LoginLayout from '../layouts/LoginLayout.vue'

const router = useRouter()

const goHome = () => {
  router.push('/tokens')
}
</script>

<template>
  <LoginLayout>
    <div class="not-found-container">
      <a-result
        status="404"
        title="404"
        sub-title="抱歉，您访问的页面不存在或已被移除。"
      >
        <template #extra>
          <a-button type="primary" @click="goHome">
            <template #icon>
              <HomeOutlined />
            </template>
            返回首页
          </a-button>
        </template>
      </a-result>
    </div>
  </LoginLayout>
</template>

<style scoped>
@reference "tailwindcss";

.not-found-container {
  @apply w-full max-w-md;
}
</style>
