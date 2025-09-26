<template>
  <a-modal
    v-model:open="visible"
    title="选择编辑器"
    :footer="null"
    width="700px"
    @cancel="handleClose"
  >
    <div class="ide-modal-content">
      <!-- VSCode系列编辑器 -->
      <div class="editor-section">
        <h4 class="section-title">VSCode 系列</h4>
        <div class="editor-grid">
          <div
            v-for="editor in vscodeEditors"
            :key="editor.id"
            class="editor-option"
            @click="handleEditorClick(editor.id)"
          >
            <div class="editor-icon">
              <img :src="editor.icon" :alt="editor.name" />
            </div>
            <div class="editor-info">
              <span class="editor-name">{{ editor.name }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- JetBrains系列编辑器 -->
      <div class="editor-section">
        <h4 class="section-title">JetBrains 系列</h4>
        <div class="editor-grid">
          <div
            v-for="editor in jetbrainsEditors"
            :key="editor.id"
            class="editor-option"
            @click="handleEditorClick(editor.id)"
          >
            <div class="editor-icon">
              <img :src="editor.icon" :alt="editor.name" />
            </div>
            <div class="editor-info">
              <span class="editor-name">{{ editor.name }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </a-modal>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { openEditor, getSupportedEditors, type Editor } from '../api/ide'
import type { Token } from '../types/token'

interface Props {
  token: Token | null
}

const props = defineProps<Props>()

const emit = defineEmits<{
  close: []
}>()

const visible = ref(false)
const loading = ref(false)
const vscodeEditors = ref<Editor[]>([])
const jetbrainsEditors = ref<Editor[]>([])

// 打开模态框
const open = () => {
  visible.value = true
  loadSupportedEditors()
}

// 关闭模态框
const handleClose = () => {
  visible.value = false
  emit('close')
}

// 加载支持的编辑器列表
const loadSupportedEditors = async () => {
  try {
    loading.value = true
    const editors = await getSupportedEditors()
    vscodeEditors.value = editors.vscode_editors
    jetbrainsEditors.value = editors.jetbrains_editors
  } catch (error) {
    message.error('加载编辑器列表失败')
    console.error('Failed to load supported editors:', error)
  } finally {
    loading.value = false
  }
}

// 处理编辑器点击
const handleEditorClick = async (editorType: string) => {
  if (!props.token) {
    message.error('Token信息不完整')
    return
  }

  try {
    loading.value = true
    
    const result = await openEditor({
      editor_type: editorType,
      token: props.token.access_token,
      tenant_url: props.token.tenant_url,
      portal_url: props.token.portal_url
    })

    if (result.success) {
      message.success(result.message)

      // 在浏览器中打开协议URL
      if (result.protocol_url) {
        // 创建一个隐藏的链接并点击它来触发协议
        const link = document.createElement('a')
        link.href = result.protocol_url
        link.style.display = 'none'
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
      }

      handleClose()
    } else {
      message.error('打开编辑器失败')
    }
  } catch (error) {
    message.error('打开编辑器失败')
    console.error('Failed to open editor:', error)
  } finally {
    loading.value = false
  }
}

// 暴露方法给父组件
defineExpose({
  open
})

onMounted(() => {
  loadSupportedEditors()
})
</script>

<style scoped>
.ide-modal-content {
  max-height: 500px;
  overflow-y: auto;
}

.editor-section {
  margin-bottom: 24px;
}

.editor-section:last-child {
  margin-bottom: 0;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: #1d1d1f;
  margin: 0 0 16px 0;
  padding-bottom: 8px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
}

.editor-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.editor-option {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border: 1px solid rgba(0, 0, 0, 0.06);
  border-radius: 8px;
  background: #ffffff;
  cursor: pointer;
  transition: all 0.2s ease;
  user-select: none;
}

.editor-option:hover {
  border-color: #007aff;
  background: rgba(0, 122, 255, 0.02);
  box-shadow: 0 2px 8px rgba(0, 122, 255, 0.12);
  transform: translateY(-1px);
}

.editor-option:active {
  transform: translateY(0);
  box-shadow: 0 1px 4px rgba(0, 122, 255, 0.08);
}

.editor-icon {
  flex-shrink: 0;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  background: rgba(0, 0, 0, 0.02);
  border: 1px solid rgba(0, 0, 0, 0.04);
}

.editor-icon img {
  width: 28px;
  height: 28px;
  object-fit: contain;
}

.editor-info {
  flex: 1;
}

.editor-name {
  font-size: 14px;
  font-weight: 500;
  color: #1d1d1f;
  letter-spacing: -0.01em;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .editor-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .editor-option {
    padding: 10px;
    gap: 10px;
  }

  .editor-icon {
    width: 36px;
    height: 36px;
  }

  .editor-icon img {
    width: 24px;
    height: 24px;
  }

  .editor-name {
    font-size: 13px;
  }
}

@media (max-width: 480px) {
  .editor-grid {
    grid-template-columns: 1fr;
  }
}
</style>
