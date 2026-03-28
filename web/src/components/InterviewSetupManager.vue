<script setup>
import { computed } from 'vue'
import InterviewSetupPreview from './InterviewSetupPreview.vue'
import StandardInterviewSetup from './StandardInterviewSetup.vue'
import AlgorithmInterviewSetup from './AlgorithmInterviewSetup.vue'

const props = defineProps({
  settings: {
    type: Object,
    required: true
  },
  setupType: {
    type: String,
    default: 'standard'
  },
  isCameraOn: {
    type: Boolean,
    required: true
  },
  isMicOn: {
    type: Boolean,
    required: true
  },
  stream: {
    type: Object,
    default: null
  },
  isProcessing: {
    type: Boolean,
    default: false
  },
  initLoadingStageText: {
    type: String,
    default: ''
  },
  initLoadingStageIndex: {
    type: Number,
    default: 0
  },
  initLoadingStageTotal: {
    type: Number,
    default: 4
  },
  initLoadingElapsedSeconds: {
    type: Number,
    default: 0
  }
})

const emit = defineEmits([
  'update:settings',
  'toggle-mic',
  'toggle-camera',
  'start-interview'
])

const settingsProxy = computed({
  get: () => props.settings,
  set: (value) => emit('update:settings', value)
})

const currentSetupComponent = computed(() => {
  return props.setupType === 'algorithm' ? AlgorithmInterviewSetup : StandardInterviewSetup
})
</script>

<template>
  <div class="relative flex-1 flex flex-col items-center justify-center max-w-4xl mx-auto w-full space-y-8 animate-in fade-in duration-500">
    <header class="text-center">
      <h1 class="text-3xl font-bold text-zinc-900">AI 模拟面试</h1>
      <p class="text-zinc-500 mt-2">配置您的面试环境与偏好，开启真实对话体验</p>
    </header>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-8 w-full items-stretch">
      <InterviewSetupPreview
        :presentation-mode="settings.presentationMode"
        :is-camera-on="isCameraOn"
        :is-mic-on="isMicOn"
        :stream="stream"
        @toggle-mic="emit('toggle-mic')"
        @toggle-camera="emit('toggle-camera')"
      />

      <component
        :is="currentSetupComponent"
        v-model:settings="settingsProxy"
        :is-processing="isProcessing"
        @start-interview="emit('start-interview', $event)"
      />
    </div>

    <div
      v-if="isProcessing"
      class="fixed inset-0 z-2147483000 flex items-center justify-center bg-white/76 backdrop-blur-[2px]"
    >
      <div class="w-[min(92vw,460px)] rounded-3xl border border-indigo-100 bg-white/95 px-6 py-6 shadow-[0_20px_60px_rgba(30,64,175,0.16)] animate-in fade-in zoom-in-95 duration-300">
        <div class="flex items-center gap-3">
          <div class="h-10 w-10 rounded-2xl bg-linear-to-br from-indigo-500 via-sky-500 to-cyan-400 p-0.5 shadow-lg shadow-indigo-200">
            <div class="h-full w-full rounded-[14px] bg-white flex items-center justify-center">
              <span class="h-4 w-4 rounded-full border-2 border-indigo-500 border-t-transparent animate-spin"></span>
            </div>
          </div>
          <div>
            <p class="text-sm text-zinc-500">面试初始化进行中</p>
            <p class="text-base font-semibold text-zinc-900" aria-live="polite">
              {{ initLoadingStageText || '正在准备您的专属面试场景...' }}
            </p>
          </div>
        </div>

        <div class="mt-5 flex items-center gap-2">
          <span
            v-for="step in Math.max(initLoadingStageTotal, 1)"
            :key="step"
            class="h-1.5 flex-1 rounded-full transition-all duration-500"
            :class="step - 1 <= initLoadingStageIndex ? 'bg-linear-to-r from-indigo-500 to-sky-500' : 'bg-zinc-200'"
          ></span>
        </div>

        <p class="mt-3 text-xs text-zinc-500">
          已等待 {{ initLoadingElapsedSeconds }} 秒，系统正在保证题目质量，请稍候...
        </p>
      </div>
    </div>
  </div>
</template>
