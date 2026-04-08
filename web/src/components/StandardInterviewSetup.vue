<script setup>
import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Loader2, Briefcase, Sparkles, ChevronRight, FileText } from 'lucide-vue-next'

const props = defineProps({
  settings: {
    type: Object,
    required: true
  },
  isProcessing: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:settings', 'start-interview'])

const positionOptions = ['Java后端工程师', '前端工程师', '算法工程师', 'AI工程师']
const styleOptions = [
  { key: 'gentle', label: '引导型', description: '循序渐进，注重表达舒适度' },
  { key: 'stress', label: '压力型', description: '高压追问，模拟真实面试现场' },
  { key: 'deep', label: '深挖型', description: '聚焦原理与工程细节深度' },
  { key: 'practical', label: '实战型', description: '强调项目落地与问题解决' }
]

const settingsProxy = computed(() => props.settings || {})

const updateSettings = (patch) => {
  emit('update:settings', {
    ...(props.settings || {}),
    interviewMode: 'ai',
    mode: 'technical',
    ...patch
  })
}

const isValid = computed(() => {
  return Boolean(settingsProxy.value.position) && Boolean(settingsProxy.value.style)
})

const startInterview = () => {
  if (!isValid.value) {
    ElMessage.warning('请先完成岗位与面试官风格配置')
    return
  }
  emit('start-interview', {
    initType: 'standard'
  })
}
</script>

<template>
  <div class="space-y-4 bg-white p-6 rounded-2xl border border-zinc-100 shadow-sm">
    <section class="rounded-xl border border-zinc-100 p-4 space-y-4">
      <p class="text-xs font-bold text-zinc-400 uppercase tracking-wider">Step 1 · 面试配置</p>

      <div class="space-y-2">
        <label class="text-xs font-semibold text-zinc-500 uppercase tracking-wider flex items-center gap-1.5">
          <Briefcase class="w-3.5 h-3.5" />
          目标岗位
        </label>
        <el-select
          :model-value="settingsProxy.position"
          placeholder="请选择目标岗位"
          class="w-full standard-position-select"
          popper-class="custom-light-select-popper"
          @update:model-value="(value) => updateSettings({ position: value })"
        >
          <el-option v-for="position in positionOptions" :key="position" :label="position" :value="position" />
        </el-select>
      </div>

      <div class="space-y-2">
        <label class="text-xs font-semibold text-zinc-500 uppercase tracking-wider flex items-center gap-1.5">
          <Sparkles class="w-3.5 h-3.5" />
          面试官风格
        </label>
        <div class="grid grid-cols-2 gap-2">
          <button
            v-for="style in styleOptions"
            :key="style.key"
            type="button"
            class="rounded-xl border px-3 py-2.5 text-left transition-colors"
            :class="settingsProxy.style === style.key ? 'border-indigo-300 bg-indigo-50 text-indigo-700' : 'border-zinc-200 bg-white text-zinc-600 hover:bg-zinc-50'"
            @click="updateSettings({ style: style.key })"
          >
            <p class="text-sm font-semibold">{{ style.label }}</p>
            <p class="text-[11px] mt-1 opacity-80">{{ style.description }}</p>
          </button>
        </div>
      </div>
    </section>

    <button
      type="button"
      class="w-full py-3 rounded-xl bg-indigo-600 text-white font-bold hover:bg-indigo-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center justify-center gap-2"
      :disabled="isProcessing || !isValid"
      @click="startInterview"
    >
      <Loader2 v-if="isProcessing" class="w-4 h-4 animate-spin" />
      <FileText v-else class="w-4 h-4" />
      <span>{{ isProcessing ? '启动中...' : '开始常规面试' }}</span>
      <ChevronRight v-if="!isProcessing" class="w-4 h-4" />
    </button>
  </div>
</template>
