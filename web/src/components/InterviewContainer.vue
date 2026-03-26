<script setup>
import { computed } from 'vue'
import InterviewSetupManager from './InterviewSetupManager.vue'
import StandardAlgorithmInterview from './StandardAlgorithmInterview.vue'
import StandardChatInterview from './StandardChatInterview.vue'

const props = defineProps({
  phase: {
    type: String,
    required: true
  },
  settings: {
    type: Object,
    required: true
  },
  setupProps: {
    type: Object,
    default: () => ({})
  },
  algorithmProps: {
    type: Object,
    default: () => ({})
  },
  chatProps: {
    type: Object,
    default: () => ({})
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
  'start-interview',
  'view-report',
  'complete-interview',
  'update:user-input',
  'send-message',
  'toggle-answer-recording',
  'close-random-reveal'
])

const currentInterviewComponent = computed(() => {
  return props.settings?.style === 'algorithm' ? StandardAlgorithmInterview : StandardChatInterview
})

const currentInterviewProps = computed(() => {
  return props.settings?.style === 'algorithm' ? props.algorithmProps : props.chatProps
})
</script>

<template>
  <div class="min-h-[calc(100vh-8rem)] flex flex-col">
    <InterviewSetupManager
      v-if="phase === 'setup'"
      v-bind="setupProps"
      :settings="settings"
      @update:settings="emit('update:settings', $event)"
      @update:shadow-coach-enabled="emit('update:shadow-coach-enabled', $event)"
      @toggle-mic="emit('toggle-mic')"
      @toggle-camera="emit('toggle-camera')"
      @change-presentation-mode="emit('change-presentation-mode', $event)"
      @change-interview-mode="emit('change-interview-mode', $event)"
      @load-invite-candidates="emit('load-invite-candidates', $event)"
      @select-invite-candidate="emit('select-invite-candidate', $event)"
      @open-bookings="emit('open-bookings')"
      @draw-blind-box="emit('draw-blind-box')"
      @redraw-blind-box="emit('redraw-blind-box')"
      @start-interview="emit('start-interview')"
    />

    <component
      :is="currentInterviewComponent"
      v-else-if="phase === 'interview'"
      v-bind="currentInterviewProps"
      :settings="settings"
      @view-report="emit('view-report')"
      @complete-interview="emit('complete-interview', $event)"
      @toggle-mic="emit('toggle-mic')"
      @toggle-camera="emit('toggle-camera')"
      @update:user-input="emit('update:user-input', $event)"
      @send-message="emit('send-message')"
      @toggle-answer-recording="emit('toggle-answer-recording')"
      @close-random-reveal="emit('close-random-reveal')"
    />
  </div>
</template>
