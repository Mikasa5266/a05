<script setup>
import { ref, computed } from 'vue'
import {
  BrainCircuit,
  User,
  Timer,
  Loader2,
  MessageSquare,
  Lightbulb,
  Headphones,
  BarChart3,
  CheckCircle,
  AlertTriangle,
  BookOpen,
  ChevronDown
} from 'lucide-vue-next'

const props = defineProps({
  settings: {
    type: Object,
    required: true
  },
  activeInvitation: {
    type: Object,
    default: null
  },
  isAlgorithmStyle: {
    type: Boolean,
    default: false
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
  algorithmBriefText: {
    type: String,
    default: ''
  },
  algorithmProgress: {
    type: Object,
    default: () => ({ current: 1, total: 0, finished: 0, passed: 0, skipped: 0, failed: 0 })
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

const emit = defineEmits(['view-report'])

const showModelAnswer = ref(false)
const isHighPressure = computed(() => ['high', 'extreme'].includes(props.pressureLevel))

const timerText = computed(() => {
  const mins = Math.floor((props.questionTimer || 0) / 60)
  const secs = (props.questionTimer || 0) % 60
  return `${mins}:${String(secs).padStart(2, '0')}`
})

const fallbackQuestionText = computed(() => {
  return (
    props.currentQuestion?.content ||
    props.currentQuestion?.title ||
    props.latestAiMessage?.content ||
    '题目加载中，请稍候...'
  )
})

const showReportButton = computed(() => {
  return props.latestAiMessage?.type === 'system' && String(props.latestAiMessage?.content || '').includes('面试结束')
})
</script>

<template>
  <div class="space-y-4">
    <div class="bg-white p-4 rounded-3xl border border-white shadow-lg shadow-zinc-200/50 flex items-center gap-4 hover:shadow-xl transition-shadow duration-300 shrink-0">
      <div class="h-14 w-14 rounded-2xl bg-gradient-to-br from-indigo-600 to-violet-600 flex items-center justify-center text-white shadow-lg shadow-indigo-500/30 ring-4 ring-indigo-50">
        <User v-if="settings.interviewMode === 'human'" class="h-7 w-7" />
        <BrainCircuit v-else class="h-7 w-7" />
      </div>
      <div>
        <h3 class="font-bold text-zinc-900 text-lg">{{ settings.interviewMode === 'human' ? '真人模拟面试' : '智聘智能引擎' }}</h3>
        <p class="text-xs text-zinc-500 font-medium flex items-center gap-1">
          <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]"></span>
          <span v-if="settings.interviewMode === 'random'">🎲 随机面试模式</span>
          <span v-else-if="settings.interviewMode === 'human'">{{ activeInvitation?.invitee?.username || '真人面试官' }} · {{ normalizeCandidateRole(activeInvitation?.invitee_role) }}</span>
          <span v-else>{{ settings.mode === 'hr' ? 'HR面试官' : settings.mode === 'comprehensive' ? '综合面试官' : 'AI 技术面试官' }} · {{ settings.style === 'gentle' ? '温和型' : settings.style === 'stress' ? '压力型' : settings.style === 'deep' ? '深挖型' : settings.style === 'practical' ? '实战型' : settings.style === 'algorithm' ? '算法型' : '标准' }}</span>
        </p>
      </div>
    </div>

    <div v-if="blindBoxScenario"
      class="p-3 rounded-2xl border shadow-sm shrink-0 flex items-center gap-3 animate-in fade-in duration-500"
      :class="[pressureColors[pressureLevel]?.bg, pressureColors[pressureLevel]?.border]">
      <span class="text-2xl">{{ blindBoxScenario.icon }}</span>
      <div class="flex-1 min-w-0">
        <div class="flex items-center gap-2">
          <span class="text-sm font-bold" :class="pressureColors[pressureLevel]?.text">{{ blindBoxScenario.name }}</span>
          <span class="text-[10px] font-bold px-1.5 py-0.5 rounded-full" :class="pressureColors[pressureLevel]?.badge">
            {{ pressureLabels[pressureLevel] }}
          </span>
        </div>
        <p class="text-[10px] text-zinc-500 truncate mt-0.5">{{ blindBoxScenario.description }}</p>
      </div>
      <div v-if="questionTimer > 0" class="flex items-center gap-1 px-2 py-1 rounded-lg shrink-0"
        :class="questionTimer <= 10 ? 'bg-rose-200 text-rose-800 animate-pulse' : 'bg-white/60 text-zinc-600'">
        <Timer class="w-3 h-3" />
        <span class="text-xs font-mono font-bold">{{ timerText }}</span>
      </div>
    </div>

    <div class="bg-white rounded-3xl border shadow-xl shadow-zinc-200/50 flex-[1.6] min-h-64 flex flex-col relative overflow-hidden group transition-all duration-300 hover:shadow-2xl hover:shadow-zinc-200/60 lg:resizable-panel lg:flex-none lg:h-[46vh]"
      :class="isHighPressure ? 'border-rose-100' : 'border-white'">
      <div class="px-6 py-5 border-b flex justify-between items-center backdrop-blur-sm z-10"
        :class="isHighPressure ? 'border-rose-50 bg-rose-50/30' : 'border-zinc-50 bg-zinc-50/50'">
        <div class="inline-flex items-center gap-1.5 px-3 py-1 text-xs font-bold rounded-full border"
          :class="isHighPressure ? 'bg-rose-50 text-rose-700 border-rose-100/50' : 'bg-indigo-50 text-indigo-700 border-indigo-100/50'">
          <span class="w-1.5 h-1.5 rounded-full" :class="isHighPressure ? 'bg-rose-600' : 'bg-indigo-600'"></span>
          {{ isAlgorithmStyle ? '语音播报引导' : ('当前题目 · 第 ' + (currentQuestionIndex + 1) + ' 题') }}
        </div>
        <div v-if="isProcessing" class="flex items-center gap-2 text-indigo-600 animate-pulse">
          <Loader2 class="w-4 h-4 animate-spin" />
          <span class="text-xs font-medium">{{ processingHint || '面试官正在评估...' }}</span>
        </div>
        <div v-else-if="!isAlgorithmStyle && latestAiMessage?.type === 'feedback'" class="flex items-center gap-2 animate-in fade-in slide-in-from-right duration-500">
          <span class="text-xs text-zinc-400 font-medium">评分</span>
          <span class="text-xl font-black text-indigo-600 tracking-tight">{{ latestAiMessage.score }}</span>
        </div>
      </div>

      <div class="flex-1 min-h-0 overflow-y-auto p-6 custom-scrollbar relative">
        <div v-if="isProcessing" class="absolute inset-0 flex flex-col items-center justify-center text-zinc-400 gap-3 bg-white/80 backdrop-blur-sm z-20">
          <div class="relative">
            <div class="absolute inset-0 bg-indigo-500/20 blur-xl rounded-full"></div>
            <Loader2 class="h-10 w-10 animate-spin text-indigo-600 relative z-10" />
          </div>
          <p class="text-sm font-medium animate-pulse">{{ processingHint || '正在生成评估...' }}</p>
        </div>

        <div v-else class="space-y-4 animate-in fade-in slide-in-from-bottom-4 duration-500">
          <template v-if="isAlgorithmStyle">
            <div class="rounded-2xl border border-cyan-100 bg-gradient-to-br from-cyan-50 to-white p-4">
              <p class="text-xs uppercase tracking-wider font-bold text-cyan-700">算法考察提示</p>
              <p class="text-base font-semibold text-zinc-900 mt-2 leading-relaxed">{{ algorithmBriefText }}</p>
              <div class="mt-4 grid grid-cols-2 gap-2 text-xs">
                <div class="rounded-xl border border-zinc-200 bg-white p-2.5">
                  <p class="text-zinc-400">当前题号</p>
                  <p class="font-bold text-zinc-800 mt-1">第 {{ algorithmProgress.current }} / {{ algorithmProgress.total || '-' }} 题</p>
                </div>
                <div class="rounded-xl border border-zinc-200 bg-white p-2.5">
                  <p class="text-zinc-400">完成进度</p>
                  <p class="font-bold text-zinc-800 mt-1">{{ algorithmProgress.finished }} 题</p>
                </div>
                <div class="rounded-xl border border-zinc-200 bg-white p-2.5">
                  <p class="text-zinc-400">通过</p>
                  <p class="font-bold text-emerald-700 mt-1">{{ algorithmProgress.passed }}</p>
                </div>
                <div class="rounded-xl border border-zinc-200 bg-white p-2.5">
                  <p class="text-zinc-400">跳过</p>
                  <p class="font-bold text-amber-700 mt-1">{{ algorithmProgress.skipped }}</p>
                </div>
              </div>
            </div>
          </template>

          <template v-else>
            <template v-if="latestAiMessage?.type === 'question' || (latestAiMessage?.role === 'ai' && !latestAiMessage?.type)">
              <h2 class="text-xl font-bold text-zinc-900 leading-relaxed tracking-wide whitespace-pre-wrap wrap-break-word">
                {{ latestAiMessage?.content }}
              </h2>
            </template>

            <template v-else-if="latestAiMessage?.type === 'feedback'">
              <div class="space-y-3">
                <div class="p-4 bg-gradient-to-br from-amber-50 to-orange-50/30 rounded-2xl border border-amber-100/60 shadow-sm">
                  <h4 class="text-xs font-bold text-amber-600 uppercase mb-2 flex items-center gap-2">
                    <div class="p-1 bg-amber-100 rounded-md">
                      <MessageSquare class="w-3.5 h-3.5" />
                    </div>
                    综合评价
                  </h4>
                  <p class="text-sm text-zinc-800 leading-relaxed text-justify whitespace-pre-wrap wrap-break-word">{{ latestAiMessage.feedbackEvaluation }}</p>
                </div>

                <div v-if="latestAiMessage.feedbackDimensions" class="p-4 bg-gradient-to-br from-indigo-50/80 to-violet-50/30 rounded-2xl border border-indigo-100/50 shadow-sm">
                  <h4 class="text-xs font-bold text-indigo-600 uppercase mb-3 flex items-center gap-2">
                    <div class="p-1 bg-indigo-100 rounded-md">
                      <BarChart3 class="w-3.5 h-3.5" />
                    </div>
                    维度评分
                  </h4>
                  <div class="space-y-2.5">
                    <div v-for="dim in [
                      { key: 'technical_depth', label: '技术深度', color: 'bg-violet-500' },
                      { key: 'expression', label: '表达清晰', color: 'bg-blue-500' },
                      { key: 'logic', label: '逻辑严谨', color: 'bg-cyan-500' },
                      { key: 'completeness', label: '覆盖完整', color: 'bg-emerald-500' }
                    ]" :key="dim.key" class="flex items-center gap-3">
                      <span class="text-xs text-zinc-500 w-14 shrink-0 text-right font-medium">{{ dim.label }}</span>
                      <div class="flex-1 h-2 bg-zinc-100 rounded-full overflow-hidden">
                        <div :class="dim.color" class="h-full rounded-full transition-all duration-700 ease-out" :style="{ width: (latestAiMessage.feedbackDimensions[dim.key] || 0) + '%' }"></div>
                      </div>
                      <span class="text-xs font-bold text-zinc-700 w-8 shrink-0">{{ latestAiMessage.feedbackDimensions[dim.key] || 0 }}</span>
                    </div>
                  </div>
                </div>

                <div v-if="(latestAiMessage.feedbackHighlights?.length || latestAiMessage.feedbackGaps?.length)" class="grid grid-cols-2 gap-2">
                  <div v-if="latestAiMessage.feedbackHighlights?.length" class="p-3 bg-emerald-50/80 rounded-xl border border-emerald-100/50">
                    <h4 class="text-[10px] font-bold text-emerald-600 uppercase mb-2 flex items-center gap-1">
                      <CheckCircle class="w-3 h-3" /> 亮点
                    </h4>
                    <ul class="space-y-1">
                      <li v-for="(h, i) in latestAiMessage.feedbackHighlights" :key="i" class="text-xs text-emerald-800 leading-relaxed flex gap-1.5">
                        <span class="text-emerald-400 mt-0.5 shrink-0">✦</span>
                        <span>{{ h }}</span>
                      </li>
                    </ul>
                  </div>
                  <div v-if="latestAiMessage.feedbackGaps?.length" class="p-3 bg-rose-50/80 rounded-xl border border-rose-100/50">
                    <h4 class="text-[10px] font-bold text-rose-600 uppercase mb-2 flex items-center gap-1">
                      <AlertTriangle class="w-3 h-3" /> 待补强
                    </h4>
                    <ul class="space-y-1">
                      <li v-for="(g, i) in latestAiMessage.feedbackGaps" :key="i" class="text-xs text-rose-800 leading-relaxed flex gap-1.5">
                        <span class="text-rose-400 mt-0.5 shrink-0">△</span>
                        <span>{{ g }}</span>
                      </li>
                    </ul>
                  </div>
                </div>

                <div class="p-4 bg-gradient-to-br from-emerald-50 to-teal-50/30 rounded-2xl border border-emerald-100/60 shadow-sm">
                  <h4 class="text-xs font-bold text-emerald-600 uppercase mb-2 flex items-center gap-2">
                    <div class="p-1 bg-emerald-100 rounded-md">
                      <Lightbulb class="w-3.5 h-3.5" />
                    </div>
                    改进建议
                  </h4>
                  <ul class="space-y-2">
                    <li v-for="(s, i) in latestAiMessage.feedbackSuggestions" :key="i" class="text-xs text-emerald-900 flex gap-2.5 leading-relaxed group/item wrap-break-word">
                      <span class="font-bold text-emerald-600/40 font-mono text-[10px] mt-0.5 group-hover/item:text-emerald-600 transition-colors shrink-0">0{{ i + 1 }}</span>
                      {{ s }}
                    </li>
                  </ul>
                </div>

                <div v-if="latestAiMessage.feedbackModelAnswer" class="p-4 bg-gradient-to-br from-sky-50/80 to-blue-50/30 rounded-2xl border border-sky-100/50 shadow-sm">
                  <h4 class="text-xs font-bold text-sky-600 uppercase mb-2 flex items-center gap-2 cursor-pointer select-none" @click="showModelAnswer = !showModelAnswer">
                    <div class="p-1 bg-sky-100 rounded-md">
                      <BookOpen class="w-3.5 h-3.5" />
                    </div>
                    参考答案思路
                    <ChevronDown class="w-3 h-3 ml-auto transition-transform duration-200" :class="showModelAnswer ? 'rotate-180' : ''" />
                  </h4>
                  <p v-show="showModelAnswer" class="text-xs text-zinc-700 leading-relaxed whitespace-pre-wrap wrap-break-word mt-1 animate-in fade-in slide-in-from-top-2 duration-300">{{ latestAiMessage.feedbackModelAnswer }}</p>
                </div>

                <div v-if="latestAiMessage.feedbackFollowUp" class="p-3 bg-zinc-50 rounded-xl border border-zinc-100">
                  <p class="text-xs text-zinc-500 flex items-start gap-2">
                    <span class="text-indigo-400 font-bold shrink-0 mt-0.5">💬</span>
                    <span><span class="font-medium text-zinc-600">面试官可能追问：</span>{{ latestAiMessage.feedbackFollowUp }}</span>
                  </p>
                </div>
              </div>
            </template>

            <template v-else-if="latestAiMessage?.type === 'system'">
              <div class="p-6 bg-zinc-50 rounded-2xl text-center text-zinc-600 text-sm border border-zinc-100">
                <p class="mb-4">{{ latestAiMessage.content }}</p>
                <div v-if="showReportButton" class="flex justify-center">
                  <button @click="emit('view-report')" class="px-8 py-3 bg-indigo-600 text-white rounded-xl text-sm font-bold hover:bg-indigo-700 transition-all shadow-lg shadow-indigo-200 hover:shadow-indigo-300 hover:-translate-y-0.5 active:translate-y-0">
                    查看详细报告
                  </button>
                </div>
              </div>
            </template>

            <template v-else>
              <h2 class="text-xl font-bold text-zinc-900 leading-relaxed tracking-wide whitespace-pre-wrap wrap-break-word">
                {{ fallbackQuestionText }}
              </h2>
            </template>
          </template>
        </div>
      </div>
    </div>

    <div v-if="shadowCoachEnabled" class="bg-gradient-to-br from-emerald-50/80 to-white p-4 rounded-3xl border border-white shadow-lg shadow-zinc-200/30 backdrop-blur-sm shrink-0 lg:resizable-panel lg:flex-none lg:h-[170px]">
      <h4 class="text-xs font-bold text-emerald-600 uppercase mb-3 flex items-center gap-2">
        <Headphones class="w-3.5 h-3.5" />
        AI 影子教练 · 实时耳返
      </h4>
      <div v-if="shadowCoachHints.length > 0" class="space-y-2">
        <p class="text-sm text-zinc-700 leading-relaxed">{{ shadowCoachHints[0].text }}</p>
        <p class="text-[11px] text-zinc-400">检测到你思考较久时，会自动给你一小点方向提示。</p>
      </div>
      <p v-else class="text-sm text-zinc-500 leading-relaxed">
        影子教练待命中。当你长时间停顿时，小幽灵会给你一句方向提醒，不会直接给答案。
      </p>
    </div>
  </div>
</template>
