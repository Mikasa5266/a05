<script setup>
import { Shuffle, X } from 'lucide-vue-next'

defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  revealInfo: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['close'])
</script>

<template>
  <div v-if="visible && revealInfo" class="fixed bottom-6 left-1/2 -translate-x-1/2 z-50 animate-in fade-in slide-in-from-bottom-4 duration-500">
    <div class="bg-white rounded-2xl shadow-2xl border border-violet-200 p-5 min-w-[360px] max-w-md">
      <div class="flex items-center gap-3 mb-3">
        <div class="p-2 bg-violet-100 rounded-xl">
          <Shuffle class="w-5 h-5 text-violet-600" />
        </div>
        <div>
          <h4 class="font-bold text-zinc-900">🎲 随机面试风格揭晓！</h4>
          <p class="text-xs text-zinc-500">本次面试采用的隐藏风格</p>
        </div>
        <button @click="emit('close')" class="ml-auto p-1 hover:bg-zinc-100 rounded-lg transition-colors">
          <X class="w-4 h-4 text-zinc-400" />
        </button>
      </div>
      <div class="flex gap-3">
        <div class="flex-1 p-3 bg-violet-50 rounded-xl border border-violet-100">
          <p class="text-[10px] text-violet-500 font-bold uppercase mb-1">面试官风格</p>
          <p class="text-sm font-bold text-violet-700">{{ revealInfo.style_label }}</p>
        </div>
        <div v-if="revealInfo.company_label" class="flex-1 p-3 bg-orange-50 rounded-xl border border-orange-100">
          <p class="text-[10px] text-orange-500 font-bold uppercase mb-1">匹配公司</p>
          <p class="text-sm font-bold text-orange-700">{{ revealInfo.company_label }}</p>
        </div>
      </div>
    </div>
  </div>
</template>
