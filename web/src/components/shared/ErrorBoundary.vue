<template>
  <div>
    <slot v-if="!hasError" />
    <div v-else class="rounded-2xl border border-rose-100 bg-rose-50 p-6 text-rose-700">
      <div class="font-bold mb-2">页面加载失败</div>
      <p class="text-sm opacity-80 mb-4">可能是网络、权限或脚本运行异常导致。</p>
      <div class="flex gap-2">
        <button @click="reload" class="px-4 py-2 rounded-lg bg-rose-600 text-white text-sm hover:bg-rose-700">刷新页面</button>
        <button @click="goBack" class="px-4 py-2 rounded-lg border border-rose-200 text-rose-700 text-sm hover:bg-white">返回上一页</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onErrorCaptured } from 'vue'
import { useRouter } from 'vue-router'

const hasError = ref(false)
const router = useRouter()

function reload() {
  // 强制整页刷新，避免白屏卡死
  window.location.reload()
}
function goBack() {
  router.back()
}

onErrorCaptured((err) => {
  // eslint-disable-next-line no-console
  console.warn('ErrorBoundary captured:', err)
  hasError.value = true
  // 阻止向上传播，避免全局报错变白屏
  return false
})
</script>

<style scoped>
</style>
