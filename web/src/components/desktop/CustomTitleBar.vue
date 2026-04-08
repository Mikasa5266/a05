<script setup>
import { ref } from 'vue'
import native from '../../utils/native'

const isMaximized = ref(false)

const handleMinimize = async () => {
  await native.minimizeWindow()
}

const handleToggleMaximize = async () => {
  isMaximized.value = await native.toggleMaximizeWindow()
}

const handleClose = async () => {
  await native.closeWindow()
}
</script>

<template>
  <header class="custom-titlebar">
    <div class="titlebar-drag-region">
      <span class="titlebar-app-name">Interview AI</span>
    </div>

    <div class="titlebar-controls">
      <button type="button" class="titlebar-btn" @click="handleMinimize" aria-label="最小化">
        <span class="titlebar-icon">-</span>
      </button>
      <button
        type="button"
        class="titlebar-btn"
        @click="handleToggleMaximize"
        :aria-label="isMaximized ? '还原窗口' : '最大化窗口'"
      >
        <span class="titlebar-icon">{{ isMaximized ? '❐' : '□' }}</span>
      </button>
      <button type="button" class="titlebar-btn titlebar-btn-close" @click="handleClose" aria-label="关闭窗口">
        <span class="titlebar-icon">×</span>
      </button>
    </div>
  </header>
</template>

<style scoped>
.custom-titlebar {
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--app-border-color);
  background: var(--app-surface);
  user-select: none;
}

.titlebar-drag-region {
  flex: 1;
  height: 100%;
  display: flex;
  align-items: center;
  padding: 0 14px;
  -webkit-app-region: drag;
}

.titlebar-app-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--app-text-primary);
  letter-spacing: 0.02em;
}

.titlebar-controls {
  display: flex;
  align-items: stretch;
  height: 100%;
  -webkit-app-region: no-drag;
}

.titlebar-btn {
  width: 44px;
  border: 0;
  background: transparent;
  color: var(--app-text-regular);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.titlebar-btn:hover {
  background: var(--el-fill-color-light);
}

.titlebar-btn-close:hover {
  background: #e5484d;
  color: #fff;
}

.titlebar-icon {
  font-size: 14px;
  line-height: 1;
}
</style>
