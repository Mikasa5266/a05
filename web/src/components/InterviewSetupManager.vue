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
  <div class="flex-1 flex flex-col items-center justify-center max-w-4xl mx-auto w-full space-y-8 animate-in fade-in duration-500">
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
  </div>
</template>
