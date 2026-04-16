<script setup>
import { computed } from 'vue'
import { Mic, MicOff, Send, Loader2, ChevronRight } from 'lucide-vue-next'
import SpeechDashboard from './SpeechDashboard.vue'

const props = defineProps({
  isAlgorithmStyle: {
    type: Boolean,
    default: false
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
  answerVoiceStatus: {
    type: String,
    default: 'idle'
  },
  isVideoInterviewMode: {
    type: Boolean,
    default: false
  },
  latestAiMessage: {
    type: Object,
    default: null
  },
  pendingEnd: {
    type: Boolean,
    default: false
  },
  canEarlySubmit: {
    type: Boolean,
    default: false
  },
  userInput: {
    type: String,
    default: ''
  },
  speechMetrics: {
    type: Object,
    default: () => ({
      speechRate: 0,
      speechRateLevel: 'normal',
      fillerWordCount: 0,
      fluencyAlert: false,
      totalFillerWords: 0
    })
  },
  energyLevel: {
    type: Number,
    default: 0
  },
  speechAnalysisActive: {
    type: Boolean,
    default: false
  },
  stream: {
    type: Object,
    default: null
  },
  micVolume: {
    type: Number,
    default: 0
  }
})

const emit = defineEmits(['toggle-answer-recording', 'send-message', 'toggle-mute', 'early-submit'])

const showSendButton = computed(() => {
  return !props.isAlgorithmStyle && (!props.isVideoInterviewMode || props.latestAiMessage?.type === 'feedback')
})

const sendDisabled = computed(() => {
  return props.isProcessing || props.isSubmitting || props.isFinishing || (!String(props.userInput || '').trim() && props.latestAiMessage?.type !== 'feedback')
})

const recordDisabled = computed(() => {
  return props.isSubmitting || props.isFinishing || !props.canAnswerCurrentQuestion || props.answerVoiceStatus === 'requesting' || props.answerVoiceStatus === 'transcribing' || props.answerVoiceStatus === 'submitting'
})

const earlySubmitDisabled = computed(() => {
  return !props.canEarlySubmit || props.isFinishing
})

const onToggleAnswerRecording = () => {
  if (props.isSubmitting || props.isFinishing) return
  emit('toggle-answer-recording')
}

const onSendMessage = () => {
  if (props.isSubmitting || props.isFinishing) return
  emit('send-message')
}

const onEarlySubmit = () => {
  if (props.isFinishing || !props.canEarlySubmit) return
  emit('early-submit')
}
</script>

<template>
  <template v-if="!isAlgorithmStyle">
    <div class="bg-white rounded-3xl p-4 border border-zinc-100 shadow-sm shrink-0 lg:resizable-panel lg:flex-none lg:h-62.5">
      <SpeechDashboard
        :speechRate="speechMetrics.speechRate"
        :speechRateLevel="speechMetrics.speechRateLevel"
        :energyLevel="energyLevel"
        :fillerWordCount="speechMetrics.fillerWordCount"
        :fluencyAlert="speechMetrics.fluencyAlert"
        :totalFillerWords="speechMetrics.totalFillerWords"
        :isActive="speechAnalysisActive"
      />
    </div>

    <button
      @click="onToggleAnswerRecording"
      :disabled="recordDisabled"
      class="w-full py-3 rounded-2xl font-bold text-base transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2 border shrink-0"
      :class="answerVoiceStatus === 'recording'
        ? 'bg-rose-600 text-white border-rose-600 hover:bg-rose-700'
        : 'bg-white text-zinc-700 border-zinc-200 hover:bg-zinc-50'"
    >
      <Mic v-if="answerVoiceStatus !== 'recording'" class="h-4 w-4" />
      <MicOff v-else class="h-4 w-4" />
      <span v-if="answerVoiceStatus === 'recording'">结束语音并开始分析</span>
      <span v-else>{{ canAnswerCurrentQuestion ? (isVideoInterviewMode ? '开始语音回答' : '语音回答') : '等待题目描述完成' }}</span>
    </button>

    <button
      @click="onEarlySubmit"
      :disabled="earlySubmitDisabled"
      class="w-full py-2.5 rounded-2xl font-semibold text-sm transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2 border shrink-0"
      :class="earlySubmitDisabled ? 'bg-zinc-100 text-zinc-400 border-zinc-200' : 'bg-amber-50 text-amber-700 border-amber-200 hover:bg-amber-100'"
    >
      提前交卷并生成报告
    </button>

    <button
      v-if="showSendButton"
      @click="onSendMessage"
      :disabled="sendDisabled"
      class="w-full py-3 bg-zinc-900 text-white rounded-2xl font-bold text-base hover:bg-zinc-800 hover:shadow-xl hover:shadow-zinc-900/20 active:scale-[0.98] transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-3 group relative overflow-hidden shrink-0"
    >
      <div class="absolute inset-0 bg-white/10 translate-y-full group-hover:translate-y-0 transition-transform duration-300"></div>
      <span v-if="isFinishing" class="flex items-center gap-2 relative z-10">
        <Loader2 class="w-5 h-5 animate-spin" />
        正在结束面试...
      </span>
      <span v-else-if="isProcessing || isSubmitting" class="flex items-center gap-2 relative z-10">
        <Loader2 class="w-5 h-5 animate-spin" />
        {{ isProcessing ? (processingHint || '正在思考...') : '正在提交...' }}
      </span>
      <span v-else-if="latestAiMessage?.type === 'feedback'" class="flex items-center gap-2 relative z-10">
        {{ pendingEnd ? '结束面试' : '下一题' }} <ChevronRight class="w-5 h-5 group-hover:translate-x-1 transition-transform" />
      </span>
      <span v-else class="flex items-center gap-2 relative z-10">
        发送回答
        <Send class="h-4 w-4 group-hover:translate-x-1 transition-transform" />
      </span>
    </button>
  </template>
</template>
