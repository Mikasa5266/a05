<script setup>
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { Video, VideoOff, Mic, MicOff } from 'lucide-vue-next'

const props = defineProps({
  presentationMode: {
    type: String,
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
  stream: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['toggle-mic', 'toggle-camera'])

const previewVideo = ref(null)

const syncPreviewStream = () => {
  if (!previewVideo.value) return
  if (props.presentationMode !== 'video_avatar' || !props.isCameraOn) {
    previewVideo.value.srcObject = null
    return
  }
  previewVideo.value.srcObject = props.stream || null
  previewVideo.value.play?.().catch(() => {})
}

watch(
  () => [props.presentationMode, props.isCameraOn, props.stream],
  syncPreviewStream,
  { immediate: true }
)

watch(previewVideo, () => {
  syncPreviewStream()
})

onMounted(() => {
  syncPreviewStream()
})

onUnmounted(() => {
  if (previewVideo.value) {
    previewVideo.value.srcObject = null
  }
})
</script>

<template>
  <div class="flex flex-col gap-4 min-h-[480px]">
    <div class="aspect-video rounded-2xl relative overflow-hidden flex items-center justify-center group shadow-xl bg-gradient-to-br from-slate-900 via-slate-800 to-cyan-900/80">
      <template v-if="presentationMode === 'video_avatar'">
        <div class="absolute inset-0 interview-room-scene interview-room-scene--compact pointer-events-none"></div>
        <div class="absolute inset-y-0 left-0 w-20 interview-room-wall interview-room-wall--left interview-room-wall--compact pointer-events-none"></div>
        <div class="absolute inset-y-0 right-0 w-24 interview-room-wall interview-room-wall--right interview-room-wall--compact pointer-events-none"></div>
        <div class="absolute left-0 right-0 bottom-0 h-20 interview-room-floor interview-room-floor--compact pointer-events-none"></div>
        <model-viewer
          src="/interview3.glb"
          autoplay
          environment-image="neutral"
          interaction-prompt="none"
          camera-target="auto auto auto"
          camera-orbit="0deg 78deg auto"
          field-of-view="28deg"
          exposure="1.35"
          shadow-intensity="0.65"
          class="w-full h-full interviewer-static relative z-20 pointer-events-none"
        ></model-viewer>

        <div class="absolute inset-x-0 bottom-0 h-20 bg-gradient-to-t from-slate-950/90 via-slate-900/45 to-transparent pointer-events-none"></div>
        <div class="absolute bottom-5 left-1/2 -translate-x-1/2 w-[58%] h-10 rounded-t-2xl interview-room-desk"></div>

        <div class="absolute bottom-4 left-4 w-36 h-24 rounded-2xl overflow-hidden border border-white/15 bg-zinc-950/90 shadow-lg">
          <video ref="previewVideo" class="w-full h-full object-cover transform scale-x-[-1]" autoplay muted v-if="isCameraOn"></video>
          <div v-else class="w-full h-full flex flex-col items-center justify-center text-zinc-500">
            <VideoOff class="h-6 w-6 mb-1" />
            <span class="text-[10px]">摄像头关闭</span>
          </div>
          <div class="absolute bottom-1 left-1 text-[10px] px-1.5 py-0.5 rounded-full bg-black/55 text-white border border-white/10">面试者</div>
        </div>

        <div class="absolute bottom-4 right-4 w-28 h-28 rounded-2xl overflow-hidden border border-emerald-200/30 bg-zinc-950/85 backdrop-blur-sm shadow-lg shadow-emerald-900/20">
          <model-viewer
            src="/cute_ghost.glb"
            autoplay
            auto-rotate
            camera-controls
            exposure="1.1"
            shadow-intensity="0.85"
            class="w-full h-full"
          ></model-viewer>
          <div class="absolute bottom-1 left-1 text-[10px] px-1.5 py-0.5 rounded-full bg-emerald-500/80 text-white border border-emerald-300/50">影子教练</div>
        </div>

        <div class="absolute z-40 bottom-4 left-1/2 -translate-x-1/2 flex items-center gap-3 pointer-events-auto">
          <button
            @click="emit('toggle-mic')"
            class="h-10 w-10 rounded-full flex items-center justify-center transition-all hover:scale-110 active:scale-95"
            :class="isMicOn ? 'bg-white/10 text-white backdrop-blur-md hover:bg-white/20' : 'bg-rose-500 text-white'"
          >
            <Mic v-if="isMicOn" class="h-4 w-4" />
            <MicOff v-else class="h-4 w-4" />
          </button>
          <button
            @click="emit('toggle-camera')"
            class="h-10 w-10 rounded-full flex items-center justify-center transition-all hover:scale-110 active:scale-95"
            :class="isCameraOn ? 'bg-white/10 text-white backdrop-blur-md hover:bg-white/20' : 'bg-rose-500 text-white'"
          >
            <Video v-if="isCameraOn" class="h-4 w-4" />
            <VideoOff v-else class="h-4 w-4" />
          </button>
        </div>
      </template>
      <template v-else>
        <div class="w-full h-full p-6 flex flex-col justify-between bg-gradient-to-br from-zinc-900 via-slate-900 to-zinc-800">
          <div>
            <p class="text-zinc-200 font-semibold">文字 + 语音模式</p>
            <p class="text-zinc-400 text-xs mt-1">专注内容质量与语言表达，适合长回答连续训练。</p>
          </div>
          <div class="grid grid-cols-3 gap-2">
            <div class="rounded-xl border border-white/10 bg-white/5 p-2">
              <p class="text-[10px] text-zinc-400">环节</p>
              <p class="text-xs text-zinc-100 mt-1">提问</p>
            </div>
            <div class="rounded-xl border border-white/10 bg-white/5 p-2">
              <p class="text-[10px] text-zinc-400">环节</p>
              <p class="text-xs text-zinc-100 mt-1">作答</p>
            </div>
            <div class="rounded-xl border border-white/10 bg-white/5 p-2">
              <p class="text-[10px] text-zinc-400">环节</p>
              <p class="text-xs text-zinc-100 mt-1">复盘</p>
            </div>
          </div>
          <div class="rounded-2xl bg-white/5 border border-white/10 p-3">
            <p class="text-zinc-200 text-xs">该模式下将保持当前高效流程：AI提问 -> 语音/文字回答 -> AI评估追问。</p>
            <div class="mt-3 h-1.5 w-full rounded-full bg-white/10 overflow-hidden">
              <div class="h-full w-2/3 bg-gradient-to-r from-cyan-400 to-emerald-400 rounded-full"></div>
            </div>
          </div>
        </div>
      </template>
    </div>

    <div class="flex-1 rounded-2xl border border-zinc-200 bg-gradient-to-br from-sky-50 via-white to-indigo-50 p-5 shadow-sm">
      <div class="grid grid-cols-2 gap-3 h-full">
        <div class="rounded-xl bg-white border border-sky-100 p-3">
          <p class="text-[11px] text-zinc-500">推荐模式</p>
          <p class="text-sm font-bold text-zinc-800 mt-1">视频模式（默认）</p>
          <p class="text-[11px] text-zinc-500 mt-2">更接近真实压力场景，结合表情与语音节奏反馈。</p>
        </div>
        <div class="rounded-xl bg-white border border-emerald-100 p-3">
          <p class="text-[11px] text-zinc-500">语音策略</p>
          <p class="text-sm font-bold text-zinc-800 mt-1">当前为无限制模式</p>
          <p class="text-[11px] text-zinc-500 mt-2">本地配置暂不限制单轮语音和TTS长度，后续可按需调整。</p>
        </div>
        <div class="rounded-xl bg-white border border-zinc-200 p-3 col-span-2">
          <div class="flex items-center justify-between">
            <p class="text-[11px] text-zinc-500">当前流程</p>
            <span class="text-[10px] px-2 py-0.5 rounded-full bg-indigo-100 text-indigo-700 font-medium">实时追问</span>
          </div>
          <p class="text-sm text-zinc-700 mt-1">AI 提问 -> 语音回答 -> 结束语音并分析 -> 下一题。追问实时生成，不读取题库缓存。</p>
        </div>
      </div>
    </div>
  </div>
</template>
