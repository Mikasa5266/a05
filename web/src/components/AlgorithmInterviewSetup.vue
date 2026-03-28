<script setup>
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Binary, Tags, Code2, ChevronRight, Loader2 } from 'lucide-vue-next'

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

const difficultyOptions = [
  { key: 'easy', label: '简单', description: '基础题为主，重思路清晰' },
  { key: 'medium', label: '中等', description: '兼顾速度与边界处理' },
  { key: 'hard', label: '困难', description: '高复杂度与优化能力' }
]

const focusTagOptions = ['动态规划', '二叉树', '图论', '双指针', '贪心', '回溯', '并查集', '堆与优先队列']
const languageOptions = ['C++', 'Go', 'Java', 'Python', 'TypeScript']

const selectedDifficulty = ref('medium')
const selectedFocusTags = ref(['动态规划'])
const selectedLanguage = ref('Go')

const settingsProxy = computed(() => props.settings || {})

const difficultyToInterviewLevel = {
  easy: 'campus_intern',
  medium: 'campus_graduate',
  hard: 'social_junior'
}

const syncAlgorithmDefaults = () => {
  const current = props.settings || {}
  const nextDifficulty = difficultyToInterviewLevel[selectedDifficulty.value] || 'campus_graduate'
  if (
    current.interviewMode === 'ai' &&
    current.mode === 'technical' &&
    current.style === 'algorithm' &&
    current.difficulty === nextDifficulty
  ) {
    return
  }
  emit('update:settings', {
    ...current,
    interviewMode: 'ai',
    mode: 'technical',
    style: 'algorithm',
    difficulty: nextDifficulty
  })
}

watch(selectedDifficulty, () => {
  syncAlgorithmDefaults()
}, { immediate: true })

const toggleFocusTag = (tag) => {
  const current = new Set(selectedFocusTags.value)
  if (current.has(tag)) {
    current.delete(tag)
  } else {
    current.add(tag)
  }
  selectedFocusTags.value = Array.from(current)
}

const isValid = computed(() => {
  return Boolean(selectedDifficulty.value) && selectedFocusTags.value.length > 0 && Boolean(selectedLanguage.value)
})

const startInterview = () => {
  if (!isValid.value) {
    ElMessage.warning('请完整选择难度、侧重点与语言偏好')
    return
  }
  emit('update:settings', {
    ...(settingsProxy.value || {}),
    interviewMode: 'ai',
    mode: 'technical',
    style: 'algorithm',
    difficulty: difficultyToInterviewLevel[selectedDifficulty.value] || 'campus_graduate'
  })
  emit('start-interview', {
    initType: 'algorithm',
    algorithmConfig: {
      difficulty: selectedDifficulty.value,
      focusTags: selectedFocusTags.value,
      language: selectedLanguage.value
    }
  })
}
</script>

<template>
  <div class="space-y-4 bg-white p-6 rounded-2xl border border-zinc-100 shadow-sm">
    <section class="rounded-xl border border-zinc-100 p-4 space-y-3">
      <p class="text-xs font-bold text-zinc-400 uppercase tracking-wider">Step 1 · 算法挑战强度</p>
      <div class="grid grid-cols-3 gap-2">
        <button
          v-for="item in difficultyOptions"
          :key="item.key"
          type="button"
          class="rounded-xl border px-3 py-2.5 text-left transition-colors"
          :class="selectedDifficulty === item.key ? 'border-indigo-300 bg-indigo-50 text-indigo-700' : 'border-zinc-200 bg-white text-zinc-600 hover:bg-zinc-50'"
          @click="selectedDifficulty = item.key"
        >
          <div class="flex items-center gap-1.5 text-sm font-semibold">
            <Binary class="w-3.5 h-3.5" />
            {{ item.label }}
          </div>
          <p class="text-[11px] mt-1 opacity-80">{{ item.description }}</p>
        </button>
      </div>
    </section>

    <section class="rounded-xl border border-zinc-100 p-4 space-y-4">
      <p class="text-xs font-bold text-zinc-400 uppercase tracking-wider">Step 2 · 题型与语言偏好</p>

      <div class="space-y-2">
        <label class="text-xs font-semibold text-zinc-500 uppercase tracking-wider flex items-center gap-1.5">
          <Tags class="w-3.5 h-3.5" />
          考察侧重点（多选）
        </label>
        <div class="grid grid-cols-2 gap-2">
          <button
            v-for="tag in focusTagOptions"
            :key="tag"
            type="button"
            class="rounded-xl border px-3 py-2 text-sm text-left transition-colors"
            :class="selectedFocusTags.includes(tag) ? 'border-emerald-300 bg-emerald-50 text-emerald-700' : 'border-zinc-200 bg-white text-zinc-600 hover:bg-zinc-50'"
            @click="toggleFocusTag(tag)"
          >
            {{ tag }}
          </button>
        </div>
      </div>

      <div class="space-y-2">
        <label class="text-xs font-semibold text-zinc-500 uppercase tracking-wider flex items-center gap-1.5">
          <Code2 class="w-3.5 h-3.5" />
          编程语言偏好
        </label>
        <div class="grid grid-cols-5 gap-2">
          <button
            v-for="language in languageOptions"
            :key="language"
            type="button"
            class="rounded-xl border px-2 py-2 text-xs font-semibold transition-colors"
            :class="selectedLanguage === language ? 'border-violet-300 bg-violet-50 text-violet-700' : 'border-zinc-200 bg-white text-zinc-600 hover:bg-zinc-50'"
            @click="selectedLanguage = language"
          >
            {{ language }}
          </button>
        </div>
      </div>
    </section>

    <button
      type="button"
      class="w-full py-3 rounded-xl bg-violet-600 text-white font-bold hover:bg-violet-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center justify-center gap-2"
      :disabled="isProcessing || !isValid"
      @click="startInterview"
    >
      <Loader2 v-if="isProcessing" class="w-4 h-4 animate-spin" />
      <span>{{ isProcessing ? '启动中...' : '开始算法对决' }}</span>
      <ChevronRight v-if="!isProcessing" class="w-4 h-4" />
    </button>
  </div>
</template>
