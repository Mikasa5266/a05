<script setup>
import { computed } from 'vue'
import InterviewSetupPreview from './InterviewSetupPreview.vue'
import InterviewSetupForm from './InterviewSetupForm.vue'

const props = defineProps({
  settings: {
    type: Object,
    required: true
  },
  shadowCoachEnabled: {
    type: Boolean,
    default: true
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
  activeInvitationId: {
    type: [Number, String],
    default: null
  },
  inviteCandidates: {
    type: Array,
    default: () => []
  },
  inviteCandidatesLoading: {
    type: Boolean,
    default: false
  },
  activeInvitation: {
    type: Object,
    default: null
  },
  blindBoxRevealing: {
    type: Boolean,
    default: false
  },
  blindBoxRevealed: {
    type: Boolean,
    default: false
  },
  blindBoxScenario: {
    type: Object,
    default: null
  },
  pressureColors: {
    type: Object,
    default: () => ({})
  },
  pressureLevel: {
    type: String,
    default: 'low'
  },
  pressureLabels: {
    type: Object,
    default: () => ({})
  },
  normalizeCandidateRole: {
    type: Function,
    required: true
  }
})

const emit = defineEmits([
  'update:settings',
  'update:shadow-coach-enabled',
  'toggle-mic',
  'toggle-camera',
  'change-presentation-mode',
  'change-interview-mode',
  'load-invite-candidates',
  'select-invite-candidate',
  'open-bookings',
  'draw-blind-box',
  'redraw-blind-box',
  'start-interview'
])

const settingsProxy = computed({
  get: () => props.settings,
  set: (value) => emit('update:settings', value)
})

const shadowCoachProxy = computed({
  get: () => props.shadowCoachEnabled,
  set: (value) => emit('update:shadow-coach-enabled', value)
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

      <InterviewSetupForm
        v-model:settings="settingsProxy"
        v-model:shadow-coach-enabled="shadowCoachProxy"
        :is-processing="isProcessing"
        :active-invitation-id="activeInvitationId"
        :invite-candidates="inviteCandidates"
        :invite-candidates-loading="inviteCandidatesLoading"
        :active-invitation="activeInvitation"
        :blind-box-revealing="blindBoxRevealing"
        :blind-box-revealed="blindBoxRevealed"
        :blind-box-scenario="blindBoxScenario"
        :pressure-colors="pressureColors"
        :pressure-level="pressureLevel"
        :pressure-labels="pressureLabels"
        :normalize-candidate-role="normalizeCandidateRole"
        @change-presentation-mode="emit('change-presentation-mode', $event)"
        @change-interview-mode="emit('change-interview-mode', $event)"
        @load-invite-candidates="emit('load-invite-candidates', $event)"
        @select-invite-candidate="emit('select-invite-candidate', $event)"
        @open-bookings="emit('open-bookings')"
        @draw-blind-box="emit('draw-blind-box')"
        @redraw-blind-box="emit('redraw-blind-box')"
        @start-interview="emit('start-interview')"
      />
    </div>
  </div>
</template>
