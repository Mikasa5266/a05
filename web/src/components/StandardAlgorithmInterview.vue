<script setup>
import { ref } from 'vue'
import AlgorithmInterviewPanel from './AlgorithmInterviewPanel.vue'
import InterviewFeedbackSidebar from './InterviewFeedbackSidebar.vue'

const props = defineProps({
  interviewId: {
    type: [Number, String],
    default: null
  },
  settings: {
    type: Object,
    required: true
  },
  activeInvitation: {
    type: Object,
    default: null
  },
  currentQuestionIndex: {
    type: Number,
    default: 0
  },
  currentQuestion: {
    type: Object,
    default: null
  },
  latestAiMessage: {
    type: Object,
    default: null
  },
  isProcessing: {
    type: Boolean,
    default: false
  },
  processingHint: {
    type: String,
    default: ''
  },
  shadowCoachEnabled: {
    type: Boolean,
    default: true
  },
  shadowCoachHints: {
    type: Array,
    default: () => []
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
  questionTimer: {
    type: Number,
    default: 0
  },
  normalizeCandidateRole: {
    type: Function,
    required: true
  }
})

const emit = defineEmits(['view-report', 'complete-interview'])

const algorithmBriefText = ref('请用算法思维，在满足复杂度约束的情况下，实现如下算法题目。')
const algorithmProgress = ref({ current: 1, total: 0, finished: 0, passed: 0, skipped: 0, failed: 0 })

const onAlgorithmBriefUpdated = (text) => {
  algorithmBriefText.value = String(text || '').trim() || '请用算法思维，在满足复杂度约束的情况下，实现如下算法题目。'
}

const onAlgorithmProgressUpdated = (progress) => {
  algorithmProgress.value = {
    current: progress?.current || 1,
    total: progress?.total || 0,
    finished: progress?.finished || 0,
    passed: progress?.passed || 0,
    skipped: progress?.skipped || 0,
    failed: progress?.failed || 0
  }
}

const onAlgorithmFinished = (payload) => {
  emit('complete-interview', payload)
}
</script>

<template>
  <div class="min-h-[calc(100vh-8rem)] flex flex-col lg:flex-row gap-6 p-6 bg-gradient-to-br from-slate-50 via-white to-cyan-50 overflow-y-auto">
    <div class="flex-1 flex flex-col gap-6 min-w-0">
      <AlgorithmInterviewPanel
        v-if="interviewId"
        :interview-id="interviewId"
        :difficulty="settings.difficulty"
        @brief-updated="onAlgorithmBriefUpdated"
        @progress-updated="onAlgorithmProgressUpdated"
        @finished="onAlgorithmFinished"
      />
    </div>

    <div class="w-full lg:w-[400px] flex flex-col gap-4 shrink-0 lg:h-full lg:min-h-0 lg:overflow-y-auto custom-scrollbar">
      <InterviewFeedbackSidebar
        :settings="settings"
        :active-invitation="activeInvitation"
        :is-algorithm-style="true"
        :current-question-index="currentQuestionIndex"
        :current-question="currentQuestion"
        :latest-ai-message="latestAiMessage"
        :is-processing="isProcessing"
        :processing-hint="processingHint"
        :algorithm-brief-text="algorithmBriefText"
        :algorithm-progress="algorithmProgress"
        :shadow-coach-enabled="shadowCoachEnabled"
        :shadow-coach-hints="shadowCoachHints"
        :blind-box-scenario="blindBoxScenario"
        :pressure-colors="pressureColors"
        :pressure-level="pressureLevel"
        :pressure-labels="pressureLabels"
        :question-timer="questionTimer"
        :normalize-candidate-role="normalizeCandidateRole"
        @view-report="emit('view-report')"
      />
    </div>
  </div>
</template>
