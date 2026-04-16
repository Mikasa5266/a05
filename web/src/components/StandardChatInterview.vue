<script setup>
import { computed, nextTick, onBeforeUnmount, ref } from 'vue'
import { Code2, History } from 'lucide-vue-next'
import InterviewVideoStage from './InterviewVideoStage.vue'
import InterviewStatusCard from './InterviewStatusCard.vue'
import InterviewFeedbackSidebar from './InterviewFeedbackSidebar.vue'
import InterviewSpeechDashboard from './InterviewSpeechDashboard.vue'
import InterviewHistoryDrawer from './InterviewHistoryDrawer.vue'
import RandomStyleReveal from './RandomStyleReveal.vue'
import LiveCodeEditor from './interview/LiveCodeEditor.vue'

const props = defineProps({
  settings: {
    type: Object,
    required: true
  },
  isCameraOn: {
    type: Boolean,
    required: true
  },
  isMicOn: {
    type: Boolean,
    required: true
  },
  isAvatarSpeaking: {
    type: Boolean,
    default: false
  },
  modelViewerReady: {
    type: Boolean,
    default: false
  },
  recordingStatus: {
    type: String,
    default: 'idle'
  },
  blindBoxScenario: {
    type: Object,
    default: null
  },
  questionTimer: {
    type: Number,
    default: 0
  },
  pressureLevel: {
    type: String,
    default: 'low'
  },
  shadowCoachHintPending: {
    type: Boolean,
    default: false
  },
  shadowCoachBubbleVisible: {
    type: Boolean,
    default: false
  },
  shadowCoachBubbleText: {
    type: String,
    default: ''
  },
  stream: {
    type: Object,
    default: null
  },
  currentQuestionIndex: {
    type: Number,
    default: 0
  },
  userInput: {
    type: String,
    default: ''
  },
  answerVoiceStatus: {
    type: String,
    default: 'idle'
  },
  getVoiceStatusLabel: {
    type: Function,
    required: true
  },
  speechMetrics: {
    type: Object,
    default: () => ({})
  },
  latestUserTranscript: {
    type: String,
    default: ''
  },
  latestAiMessage: {
    type: Object,
    default: null
  },
  isProcessing: {
    type: Boolean,
    default: false
  },
  isSubmitting: {
    type: Boolean,
    default: false
  },
  isFinishing: {
    type: Boolean,
    default: false
  },
  processingHint: {
    type: String,
    default: ''
  },
  canAnswerCurrentQuestion: {
    type: Boolean,
    default: false
  },
  pendingEnd: {
    type: Boolean,
    default: false
  },
  canEarlySubmit: {
    type: Boolean,
    default: false
  },
  energyLevel: {
    type: Number,
    default: 0
  },
  speechAnalysisActive: {
    type: Boolean,
    default: false
  },
  messages: {
    type: Array,
    default: () => []
  },
  activeInvitation: {
    type: Object,
    default: null
  },
  currentQuestion: {
    type: Object,
    default: null
  },
  shadowCoachEnabled: {
    type: Boolean,
    default: true
  },
  shadowCoachHints: {
    type: Array,
    default: () => []
  },
  pressureColors: {
    type: Object,
    default: () => ({})
  },
  pressureLabels: {
    type: Object,
    default: () => ({})
  },
  normalizeCandidateRole: {
    type: Function,
    required: true
  },
  randomStyleRevealed: {
    type: Boolean,
    default: false
  },
  revealedStyleInfo: {
    type: Object,
    default: null
  }
})

const emit = defineEmits([
  'toggle-mic',
  'toggle-camera',
  'update:user-input',
  'send-message',
  'toggle-answer-recording',
  'early-submit',
  'view-report',
  'close-random-reveal'
])

const showHistory = ref(false)
const showLiveCoding = ref(false)
const liveCodingCode = ref('')
const liveCodingLanguage = ref('javascript')
const liveCodingSubmitting = ref(false)
const liveCodingQuestion = ref({
  title: '突击算法题（Demo）：两数之和',
  content: '给定一个整数数组 nums 和一个目标值 target，请返回两个数的下标，使得它们相加等于 target。可以假设每种输入只会对应一个答案，且同一个元素不能重复使用。'
})

let liveCodingSubmitTimer = null

const isHumanInterviewMode = computed(() => props.settings.interviewMode === 'human')
const isVideoInterviewMode = computed(() => props.settings.presentationMode === 'video_avatar' && !isHumanInterviewMode.value)
const userInputProxy = computed({
  get: () => props.userInput,
  set: (value) => emit('update:user-input', value)
})

const voiceStatusClass = computed(() => {
  if (props.answerVoiceStatus === 'recording') return 'bg-rose-50 text-rose-600 border-rose-200'
  if (props.answerVoiceStatus === 'success') return 'bg-emerald-50 text-emerald-600 border-emerald-200'
  if (props.answerVoiceStatus === 'error') return 'bg-amber-50 text-amber-600 border-amber-200'
  return 'bg-zinc-50 text-zinc-500 border-zinc-200'
})

const buildLiveCodingStarter = (language) => {
  const lang = String(language || 'javascript').trim().toLowerCase()
  if (lang === 'go') {
    return [
      'package main',
      '',
      'func twoSum(nums []int, target int) []int {',
      '\tlookup := make(map[int]int)',
      '\tfor i, n := range nums {',
      '\t\tif j, ok := lookup[target-n]; ok {',
      '\t\t\treturn []int{j, i}',
      '\t\t}',
      '\t\tlookup[n] = i',
      '\t}',
      '\treturn nil',
      '}',
    ].join('\n')
  }

  if (lang === 'python') {
    return [
      'def two_sum(nums: list[int], target: int) -> list[int]:',
      '    lookup = {}',
      '    for i, n in enumerate(nums):',
      '        if target - n in lookup:',
      '            return [lookup[target - n], i]',
      '        lookup[n] = i',
      '    return []',
    ].join('\n')
  }

  return [
    'function twoSum(nums, target) {',
    '  const lookup = new Map();',
    '  for (let i = 0; i < nums.length; i += 1) {',
    '    const n = nums[i];',
    '    const pair = target - n;',
    '    if (lookup.has(pair)) return [lookup.get(pair), i];',
    '    lookup.set(n, i);',
    '  }',
    '  return [];',
    '}',
  ].join('\n')
}

const openLiveCodingDemo = () => {
  liveCodingLanguage.value = 'javascript'
  liveCodingCode.value = buildLiveCodingStarter(liveCodingLanguage.value)
  liveCodingSubmitting.value = false
  showLiveCoding.value = true
}

const closeLiveCodingDemo = () => {
  if (liveCodingSubmitTimer) {
    clearTimeout(liveCodingSubmitTimer)
    liveCodingSubmitTimer = null
  }
  liveCodingSubmitting.value = false
  showLiveCoding.value = false
}

const handleLiveCodingLanguageChange = (nextLanguage) => {
  const prevLanguage = liveCodingLanguage.value
  const normalized = String(nextLanguage || 'javascript').trim().toLowerCase() || 'javascript'
  const previousStarter = buildLiveCodingStarter(prevLanguage).trim()
  const currentCode = String(liveCodingCode.value || '').trim()

  liveCodingLanguage.value = normalized
  if (!currentCode || currentCode === previousStarter) {
    liveCodingCode.value = buildLiveCodingStarter(normalized)
  }
}

const submitLiveCodingDemo = () => {
  if (liveCodingSubmitting.value) return

  const submittedCode = String(liveCodingCode.value || '').trim()
  liveCodingSubmitting.value = true

  liveCodingSubmitTimer = setTimeout(async () => {
    liveCodingSubmitTimer = null
    liveCodingSubmitting.value = false
    showLiveCoding.value = false

    const injectedMockMessage = [
      '[系统Mock判题] 考生已提交代码，系统判定结果为0分。',
      '',
      '候选人提交代码片段：',
      submittedCode || '(空代码)'
    ].join('\n')

    emit('update:user-input', injectedMockMessage)
    await nextTick()
    emit('send-message')
  }, 1500)
}

onBeforeUnmount(() => {
  if (liveCodingSubmitTimer) {
    clearTimeout(liveCodingSubmitTimer)
    liveCodingSubmitTimer = null
  }
})
</script>

<template>
  <div class="min-h-[calc(100vh-8rem)] flex flex-col lg:flex-row gap-6 p-6 bg-linear-to-br from-slate-50 via-white to-cyan-50 overflow-y-auto">
    <div class="flex-1 flex flex-col gap-6 min-w-0">
      <InterviewVideoStage
        v-if="settings.presentationMode === 'video_avatar' && settings.interviewMode !== 'human'"
        :is-camera-on="isCameraOn"
        :is-mic-on="isMicOn"
        :is-avatar-speaking="isAvatarSpeaking"
        :model-viewer-ready="modelViewerReady"
        :recording-status="recordingStatus"
        :blind-box-scenario="blindBoxScenario"
        :question-timer="questionTimer"
        :pressure-level="pressureLevel"
        :interview-style="settings.style"
        :shadow-coach-hint-pending="shadowCoachHintPending"
        :shadow-coach-bubble-visible="shadowCoachBubbleVisible"
        :shadow-coach-bubble-text="shadowCoachBubbleText"
        :stream="stream"
        @toggle-mic="emit('toggle-mic')"
        @toggle-camera="emit('toggle-camera')"
      />
      <InterviewStatusCard
        v-else
        :position="settings.position"
        :current-question-index="currentQuestionIndex"
      />

      <div class="bg-white rounded-3xl p-6 shadow-xl shadow-zinc-200/50 border border-white flex flex-col relative transition-all duration-300 focus-within:ring-2 focus-within:ring-indigo-500/20 focus-within:shadow-indigo-500/10 group lg:resizable-panel lg:flex-none"
        :class="isVideoInterviewMode ? 'min-h-80 lg:min-h-95 lg:h-auto' : 'min-h-65 lg:min-h-80 lg:h-auto'">
        <div class="flex flex-col sm:flex-row sm:justify-between sm:items-center items-start gap-2 mb-4 shrink-0">
          <h3 class="font-bold text-zinc-900 flex items-center gap-2 group-focus-within:text-indigo-600 transition-colors">
            <div class="w-1.5 h-4 bg-zinc-300 rounded-full group-focus-within:bg-indigo-600 transition-colors"></div>
            {{ isVideoInterviewMode ? '语音回答控制台' : '实时回答转写' }}
            <span v-if="userInput.length > 0" class="text-xs font-normal text-emerald-600 flex items-center gap-1 bg-emerald-50 px-2 py-0.5 rounded-full animate-in fade-in zoom-in duration-300">
              <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
              正在输入...
            </span>
          </h3>
          <div class="flex items-center gap-2 w-full sm:w-auto justify-end">
            <span class="text-[11px] font-medium px-2 py-1 rounded-full border" :class="voiceStatusClass">
              {{ getVoiceStatusLabel() }}
            </span>
            <button
              @click.stop="openLiveCodingDemo"
              class="text-xs text-zinc-400 hover:text-indigo-600 transition-colors flex items-center gap-1 px-2 py-1 hover:bg-zinc-50 rounded-lg"
            >
              <Code2 class="w-3 h-3" /> 突击代码(Demo)
            </button>
            <button @click.stop="showHistory = true" class="text-xs text-zinc-400 hover:text-indigo-600 transition-colors flex items-center gap-1 px-2 py-1 hover:bg-zinc-50 rounded-lg">
              <History class="w-3 h-3" /> 历史记录
            </button>
          </div>
        </div>

        <template v-if="isVideoInterviewMode">
          <div class="flex-1 min-h-0 rounded-2xl border border-zinc-100 bg-zinc-50/70 p-4 flex flex-col gap-3 overflow-auto custom-scrollbar">
            <p class="text-sm text-zinc-600 leading-relaxed shrink-0">
              视频模式已禁用文字提交。请点击下方按钮开始作答，回答完成后点击“结束语音并开始分析”。
            </p>
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 shrink-0">
              <div class="rounded-xl bg-white border border-zinc-200 p-3">
                <p class="text-[11px] text-zinc-400">语速</p>
                <p class="text-sm font-bold text-zinc-800 mt-1">{{ speechMetrics.speechRate || 0 }} 字/分</p>
              </div>
              <div class="rounded-xl bg-white border border-zinc-200 p-3">
                <p class="text-[11px] text-zinc-400">本轮填充词</p>
                <p class="text-sm font-bold text-zinc-800 mt-1">{{ speechMetrics.fillerWordCount || 0 }}</p>
              </div>
            </div>
            <div class="rounded-xl bg-white border border-zinc-200 p-3 min-h-0 flex-1 overflow-hidden">
              <p class="text-[11px] text-zinc-400">最近语音转写</p>
              <p class="text-sm text-zinc-700 mt-1 whitespace-pre-wrap wrap-break-word max-h-full overflow-y-auto custom-scrollbar">{{ speechMetrics.transcribedText || latestUserTranscript || '暂无转写内容' }}</p>
            </div>
          </div>
        </template>
        <template v-else>
          <textarea
            v-model="userInputProxy"
            @keydown.ctrl.enter="emit('send-message')"
            placeholder="在此处输入您的回答..."
            class="flex-1 min-h-55 max-h-[55vh] w-full resize-y border-none focus:ring-0 p-4 text-lg text-zinc-700 placeholder:text-zinc-300 bg-zinc-50/50 rounded-xl leading-relaxed transition-all focus:bg-white focus:shadow-inner overflow-y-auto custom-scrollbar"
          ></textarea>

          <div class="absolute bottom-8 right-8 text-[10px] font-medium text-zinc-300 pointer-events-none bg-white/80 backdrop-blur px-2 py-1 rounded-md border border-zinc-100">
            Ctrl + Enter 发送
          </div>
        </template>
      </div>
    </div>

    <div class="w-full lg:w-100 flex flex-col gap-4 shrink-0 lg:h-full lg:min-h-0 lg:overflow-y-auto custom-scrollbar">
      <InterviewFeedbackSidebar
        :settings="settings"
        :active-invitation="activeInvitation"
        :is-algorithm-style="false"
        :current-question-index="currentQuestionIndex"
        :current-question="currentQuestion"
        :latest-ai-message="latestAiMessage"
        :is-processing="isProcessing"
        :is-finishing="isFinishing"
        :processing-hint="processingHint"
        :algorithm-brief-text="''"
        :algorithm-progress="{ current: 1, total: 0, finished: 0, passed: 0, skipped: 0, failed: 0 }"
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

      <InterviewSpeechDashboard
        :is-algorithm-style="false"
        :is-processing="isProcessing"
        :is-submitting="isSubmitting"
        :is-finishing="isFinishing"
        :processing-hint="processingHint"
        :can-answer-current-question="canAnswerCurrentQuestion"
        :answer-voice-status="answerVoiceStatus"
        :is-video-interview-mode="isVideoInterviewMode"
        :latest-ai-message="latestAiMessage"
        :pending-end="pendingEnd"
        :can-early-submit="canEarlySubmit"
        :user-input="userInput"
        :speech-metrics="speechMetrics"
        :energy-level="energyLevel"
        :speech-analysis-active="speechAnalysisActive"
        :stream="stream"
        :mic-volume="energyLevel"
        @toggle-answer-recording="emit('toggle-answer-recording')"
        @send-message="emit('send-message')"
        @early-submit="emit('early-submit')"
      />
    </div>

    <InterviewHistoryDrawer
      v-model:show-history="showHistory"
      :messages="messages"
      :settings="settings"
    />

    <RandomStyleReveal
      :visible="randomStyleRevealed"
      :reveal-info="revealedStyleInfo"
      @close="emit('close-random-reveal')"
    />

    <div
      v-if="showLiveCoding"
      class="fixed inset-0 z-200 bg-slate-950/75 backdrop-blur-sm flex items-center justify-center p-4 sm:p-6 animate-in fade-in duration-200"
    >
      <div
        class="w-full max-w-4xl max-h-[88vh] overflow-hidden rounded-3xl border border-white/15 bg-linear-to-br from-slate-900/95 via-slate-900/92 to-indigo-950/85 shadow-2xl animate-in zoom-in-95 duration-200"
      >
        <div class="max-h-[88vh] overflow-y-auto custom-scrollbar p-4 sm:p-6">
          <LiveCodeEditor
            :visible="showLiveCoding"
            :model-value="liveCodingCode"
            :question="liveCodingQuestion"
            :language="liveCodingLanguage"
            :submitting="liveCodingSubmitting"
            :read-only="false"
            @update:modelValue="liveCodingCode = $event"
            @language-change="handleLiveCodingLanguageChange"
            @submit="submitLiveCodingDemo"
            @close="closeLiveCodingDemo"
          />
        </div>
      </div>
    </div>
  </div>
</template>
