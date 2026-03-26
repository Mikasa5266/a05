<script setup>
import { ref, watch, onMounted, onUnmounted, computed } from 'vue'
import { Timer, User, Mic, MicOff, Video, VideoOff } from 'lucide-vue-next'

const props = defineProps({
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
    required: true
  },
  modelViewerReady: {
    type: Boolean,
    required: true
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
  interviewStyle: {
    type: String,
    default: 'gentle'
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
  }
})

const emit = defineEmits(['toggle-mic', 'toggle-camera'])

const interviewVideo = ref(null)

const isHighPressure = computed(() => ['high', 'extreme'].includes(props.pressureLevel))

const recordingStatusClass = computed(() => {
  if (props.recordingStatus === 'uploaded') return 'bg-emerald-600 shadow-emerald-900/20'
  if (props.recordingStatus === 'failed') return 'bg-amber-600 shadow-amber-900/20'
  return 'bg-rose-600 shadow-rose-900/20'
})

const recordingStatusText = computed(() => {
  if (props.recordingStatus === 'uploaded') return 'REC OK'
  if (props.recordingStatus === 'failed') return 'REC WARN'
  return 'REC'
})

const timerClass = computed(() => {
  if (props.questionTimer <= 10) {
    return 'bg-rose-600/80 border-rose-400 text-white animate-pulse'
  }
  if (props.questionTimer <= 30) {
    return 'bg-amber-500/70 border-amber-300 text-white'
  }
  return 'bg-black/40 border-white/10 text-white/90'
})

const timerText = computed(() => {
  const mins = Math.floor((props.questionTimer || 0) / 60)
  const secs = (props.questionTimer || 0) % 60
  return `${mins}:${String(secs).padStart(2, '0')}`
})

const syncVideoStream = () => {
  if (!interviewVideo.value) return
  if (!props.isCameraOn) {
    interviewVideo.value.srcObject = null
    return
  }
  interviewVideo.value.srcObject = props.stream || null
  interviewVideo.value.play?.().catch(() => {})
}

watch(
  () => [props.isCameraOn, props.stream],
  syncVideoStream,
  { immediate: true }
)

watch(interviewVideo, () => {
  syncVideoStream()
})

onMounted(() => {
  syncVideoStream()
})

onUnmounted(() => {
  if (interviewVideo.value) {
    interviewVideo.value.srcObject = null
  }
})
</script>

<template>
  <div class="flex-1 min-h-[320px] lg:min-h-[420px] rounded-3xl relative overflow-hidden shadow-2xl group ring-1 ring-slate-900/10 bg-gradient-to-br from-slate-900 via-slate-800 to-cyan-900/70">
    <div class="absolute top-6 left-6 flex items-center gap-3 z-10 pointer-events-none">
      <div class="text-white text-xs font-bold px-3 py-1.5 rounded-full flex items-center gap-2 shadow-lg" :class="recordingStatusClass">
        <div class="w-2 h-2 bg-white rounded-full animate-pulse"></div>
        {{ recordingStatusText }}
      </div>
      <div v-if="blindBoxScenario" class="text-white text-xs font-bold px-3 py-1.5 rounded-full flex items-center gap-2 backdrop-blur-md border border-white/10 shadow-sm"
        :class="isHighPressure ? 'bg-rose-600/60' : 'bg-black/40'">
        <span>{{ blindBoxScenario.icon }}</span>
        {{ blindBoxScenario.name }}
      </div>
      <div v-else class="bg-black/40 text-white/90 text-xs px-3 py-1.5 rounded-full backdrop-blur-md border border-white/10 shadow-sm">
        多模态情绪监测中...
      </div>
    </div>

    <div v-if="questionTimer > 0" class="absolute top-6 right-6 z-10 pointer-events-none">
      <div class="flex items-center gap-2 px-4 py-2 rounded-full backdrop-blur-md shadow-lg border" :class="timerClass">
        <Timer class="w-4 h-4" />
        <span class="text-lg font-mono font-black tracking-wider">{{ timerText }}</span>
      </div>
    </div>

    <div v-if="isHighPressure" class="absolute inset-0 pointer-events-none z-[5]"
      :class="pressureLevel === 'extreme'
        ? 'shadow-[inset_0_0_80px_rgba(220,38,38,0.25)]'
        : 'shadow-[inset_0_0_60px_rgba(220,38,38,0.1)]'"
    ></div>

    <div class="absolute inset-0 interview-room-scene pointer-events-none">
      <div class="absolute inset-y-0 left-0 w-32 interview-room-wall interview-room-wall--left pointer-events-none"></div>
      <div class="absolute inset-y-0 right-0 w-36 interview-room-wall interview-room-wall--right pointer-events-none"></div>
      <div class="absolute left-0 right-0 bottom-0 h-28 interview-room-floor pointer-events-none"></div>
      <model-viewer
        v-if="modelViewerReady"
        src="/interview3.glb"
        autoplay
        environment-image="neutral"
        interaction-prompt="none"
        camera-target="auto auto auto"
        camera-orbit="0deg 78deg auto"
        field-of-view="28deg"
        shadow-intensity="0.65"
        exposure="1.35"
        class="w-full h-full interviewer-stage interviewer-static relative z-20 pointer-events-none"
        :class="isAvatarSpeaking ? 'interviewer-speaking' : ''"
      ></model-viewer>
      <div v-else class="absolute inset-0 z-20 flex items-center justify-center text-center px-6">
        <div class="rounded-2xl border border-white/20 bg-slate-900/45 text-slate-100 px-5 py-4 backdrop-blur-sm">
          <p class="text-sm font-semibold">AI 面试官模型加载中</p>
          <p class="text-xs text-slate-300 mt-1">当前网络较慢时，3D 模型可能延迟出现</p>
        </div>
      </div>

      <div class="absolute inset-x-0 bottom-0 h-24 bg-gradient-to-t from-slate-950/95 via-slate-900/45 to-transparent pointer-events-none"></div>
      <div class="absolute bottom-8 left-1/2 -translate-x-1/2 w-[56%] h-14 rounded-t-2xl interview-room-desk"></div>
      <div class="absolute top-6 right-6 text-xs px-3 py-1.5 rounded-full border backdrop-blur-sm"
        :class="interviewStyle === 'stress' ? 'bg-rose-500/20 text-rose-100 border-rose-300/40' : 'bg-emerald-500/15 text-emerald-100 border-emerald-300/40'">
        {{ interviewStyle === 'stress' ? '施压面试状态' : '安抚面试状态' }}
      </div>
    </div>

    <div class="absolute bottom-5 left-5 w-44 h-28 rounded-2xl overflow-hidden border border-white/20 bg-zinc-950/90 shadow-lg backdrop-blur-sm">
      <video ref="interviewVideo" class="w-full h-full object-cover transform scale-x-[-1]" autoplay muted v-if="isCameraOn"></video>
      <div v-else class="w-full h-full flex flex-col items-center justify-center text-zinc-600 bg-zinc-900/70">
        <User class="h-8 w-8 mb-2 opacity-30" />
        <p class="text-[11px] tracking-wide opacity-60">面试者画面关闭</p>
      </div>
      <div class="absolute bottom-1 left-1 text-[10px] px-2 py-0.5 rounded-full bg-black/55 text-white border border-white/10">面试者</div>
    </div>

    <div class="absolute bottom-5 right-5 w-36 h-40 rounded-2xl overflow-hidden border border-emerald-200/35 backdrop-blur-sm bg-zinc-950/90"
      :class="shadowCoachHintPending ? 'ring-2 ring-emerald-300/80 shadow-[0_0_24px_rgba(16,185,129,0.35)]' : ''">
      <model-viewer
        v-if="modelViewerReady"
        src="/cute_ghost.glb"
        autoplay
        auto-rotate
        camera-controls
        exposure="1.05"
        shadow-intensity="1"
        class="w-full h-full bg-gradient-to-b from-zinc-900 to-zinc-800 pointer-events-none"
      ></model-viewer>
      <div v-else class="w-full h-full flex items-center justify-center text-[11px] text-emerald-100 bg-zinc-900/70">影子教练加载中</div>
      <div class="absolute bottom-2 left-2 text-[10px] px-2 py-0.5 rounded-full bg-emerald-500/85 text-white border border-emerald-300/60">
        影子教练
      </div>
    </div>

    <transition name="coach-bubble">
      <div
        v-if="shadowCoachBubbleVisible && shadowCoachBubbleText"
        class="absolute z-[70] right-44 bottom-28 max-w-[250px] rounded-2xl border border-emerald-200 bg-white/95 px-4 py-3 text-xs leading-relaxed text-zinc-700 shadow-lg shadow-emerald-100"
      >
        <p class="font-semibold text-emerald-700 mb-1">小幽灵提示</p>
        <p class="whitespace-pre-wrap">{{ shadowCoachBubbleText }}</p>
      </div>
    </transition>

    <div class="absolute z-40 bottom-8 left-1/2 -translate-x-1/2 flex gap-4 transition-all duration-500 translate-y-4 opacity-0 group-hover:translate-y-0 group-hover:opacity-100 pointer-events-auto">
      <button
        @click="emit('toggle-mic')"
        class="h-12 w-12 rounded-full flex items-center justify-center backdrop-blur-md transition-all hover:scale-110 shadow-lg border border-white/10"
        :class="isMicOn ? 'bg-white/20 text-white hover:bg-white/30' : 'bg-rose-500 text-white'"
      >
        <Mic v-if="isMicOn" class="h-5 w-5" />
        <MicOff v-else class="h-5 w-5" />
      </button>
      <button
        @click="emit('toggle-camera')"
        class="h-12 w-12 rounded-full flex items-center justify-center backdrop-blur-md transition-all hover:scale-110 shadow-lg border border-white/10"
        :class="isCameraOn ? 'bg-white/20 text-white hover:bg-white/30' : 'bg-rose-500 text-white'"
      >
        <Video v-if="isCameraOn" class="h-5 w-5" />
        <VideoOff v-else class="h-5 w-5" />
      </button>
    </div>
  </div>
</template>

<style scoped>
.interview-room-scene {
  background-color: #0f172a;
}

.interview-room-scene::before {
  content: '';
  position: absolute;
  inset: 0;
  background:
    linear-gradient(180deg, rgba(2, 6, 23, 0.3) 0%, rgba(2, 6, 23, 0.62) 100%),
    linear-gradient(100deg, rgba(2, 132, 199, 0.14) 0%, rgba(2, 132, 199, 0) 52%),
    linear-gradient(180deg, rgba(15, 23, 42, 0) 64%, rgba(2, 6, 23, 0.55) 100%),
    url('/interview-room.jpg');
  background-size: cover;
  background-position: center;
  pointer-events: none;
}

.interview-room-scene::after {
  content: '';
  position: absolute;
  inset: 0;
  background:
    radial-gradient(56% 42% at 50% 82%, rgba(148, 163, 184, 0.18) 0%, rgba(148, 163, 184, 0) 78%),
    radial-gradient(38% 26% at 14% 14%, rgba(251, 191, 36, 0.12) 0%, rgba(251, 191, 36, 0) 70%);
  pointer-events: none;
}

.interview-room-desk {
  background: linear-gradient(90deg, rgba(120, 53, 15, 0.65) 0%, rgba(146, 64, 14, 0.62) 46%, rgba(120, 53, 15, 0.65) 100%);
  border: 1px solid rgba(253, 186, 116, 0.24);
  box-shadow: 0 0 34px rgba(15, 23, 42, 0.35);
}

.interview-room-wall {
  top: 18%;
  bottom: 14%;
  border: 1px solid rgba(148, 163, 184, 0.08);
  background: linear-gradient(180deg, rgba(15, 23, 42, 0.22) 0%, rgba(15, 23, 42, 0.08) 100%);
  opacity: 0.35;
}

.interview-room-wall--left {
  left: -8px;
  transform: perspective(420px) rotateY(18deg);
  transform-origin: left center;
  border-radius: 0 14px 14px 0;
}

.interview-room-wall--right {
  right: -10px;
  transform: perspective(420px) rotateY(-18deg);
  transform-origin: right center;
  border-radius: 14px 0 0 14px;
}

.interview-room-floor {
  left: 8%;
  right: 8%;
  bottom: 6%;
  height: 24%;
  border-radius: 22px 22px 10px 10px;
  border: 1px solid rgba(56, 189, 248, 0.08);
  background:
    linear-gradient(180deg, rgba(15, 23, 42, 0) 0%, rgba(2, 132, 199, 0.05) 32%, rgba(15, 23, 42, 0.35) 100%),
    linear-gradient(100deg, rgba(30, 41, 59, 0.55) 0%, rgba(15, 23, 42, 0.7) 100%);
  transform: perspective(520px) rotateX(60deg);
  transform-origin: bottom center;
  box-shadow: 0 22px 40px rgba(2, 6, 23, 0.35);
  opacity: 0.45;
}

.interviewer-stage {
  animation: interviewer-idle 5.2s ease-in-out infinite;
  transform-origin: center 78%;
}

.interviewer-static {
  pointer-events: none;
}

.interviewer-speaking {
  animation: interviewer-speaking 1.25s ease-in-out infinite;
}

.coach-bubble-enter-active,
.coach-bubble-leave-active {
  transition: opacity 0.22s ease, transform 0.22s ease;
}

.coach-bubble-enter-from,
.coach-bubble-leave-to {
  opacity: 0;
  transform: translateY(6px) scale(0.96);
}

@keyframes interviewer-idle {
  0% { transform: translateY(0px) rotate(0deg); }
  35% { transform: translateY(-1px) rotate(-0.2deg); }
  70% { transform: translateY(1px) rotate(0.2deg); }
  100% { transform: translateY(0px) rotate(0deg); }
}

@keyframes interviewer-speaking {
  0% { transform: translateY(0px) rotate(0deg) scale(1); }
  50% { transform: translateY(-1px) rotate(-0.2deg) scale(1.01); }
  100% { transform: translateY(0px) rotate(0deg) scale(1); }
}
</style>
