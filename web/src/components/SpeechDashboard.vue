<script setup>
import { ref, computed, watch, onUnmounted } from 'vue'
import { Activity, Gauge, AlertTriangle, Volume2 } from 'lucide-vue-next'

const props = defineProps({
  // Speech rate in chars/min
  speechRate: { type: Number, default: 0 },
  // "slow" | "normal" | "fast"
  speechRateLevel: { type: String, default: 'normal' },
  // Real-time energy level 0-1 from Web Audio API
  energyLevel: { type: Number, default: 0 },
  // Filler word count in latest chunk
  fillerWordCount: { type: Number, default: 0 },
  // Whether fluency alert is active
  fluencyAlert: { type: Boolean, default: false },
  // Accumulated filler words total
  totalFillerWords: { type: Number, default: 0 },
  // Whether currently recording/active
  isActive: { type: Boolean, default: false }
})

// Fluency alert blink animation
const alertBlink = ref(false)
let blinkTimer = null

watch(() => props.fluencyAlert, (val) => {
  if (val) {
    alertBlink.value = true
    blinkTimer = setInterval(() => {
      alertBlink.value = !alertBlink.value
    }, 500)
  } else {
    alertBlink.value = false
    if (blinkTimer) clearInterval(blinkTimer)
  }
})

onUnmounted(() => {
  if (blinkTimer) clearInterval(blinkTimer)
})

// Speech rate gauge: mapped to 0-100 scale, where 200 chars/min = 50%
const speechRatePercent = computed(() => {
  return Math.min(100, Math.max(0, (props.speechRate / 400) * 100))
})

const speechRateColor = computed(() => {
  if (props.speechRate < 120) return 'text-blue-500'
  if (props.speechRate <= 240) return 'text-emerald-500'
  return 'text-rose-500'
})

const speechRateBarColor = computed(() => {
  if (props.speechRate < 120) return 'bg-blue-500'
  if (props.speechRate <= 240) return 'bg-emerald-500'
  return 'bg-rose-500'
})

const speechRateLabel = computed(() => {
  const map = { slow: '偏慢', normal: '正常', fast: '偏快' }
  return map[props.speechRateLevel] || '正常'
})

// Energy bar: 0-100%
const energyPercent = computed(() => {
  return Math.min(100, Math.max(0, props.energyLevel * 100))
})

const energyBarColor = computed(() => {
  if (props.energyLevel < 0.15) return 'bg-zinc-300'
  if (props.energyLevel < 0.5) return 'bg-indigo-500'
  if (props.energyLevel < 0.8) return 'bg-emerald-500'
  return 'bg-amber-500'
})

const energyLabel = computed(() => {
  if (props.energyLevel < 0.15) return '过低'
  if (props.energyLevel < 0.5) return '偏低'
  if (props.energyLevel < 0.8) return '适中'
  return '饱满'
})
</script>

<template>
  <div class="speech-dashboard space-y-3">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <h4 class="text-xs font-bold text-zinc-400 uppercase flex items-center gap-2">
        <Activity class="w-3.5 h-3.5 text-indigo-500" />
        实时语音仪表盘
      </h4>
      <div v-if="isActive" class="flex items-center gap-1.5">
        <span class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse shadow-[0_0_6px_rgba(16,185,129,0.5)]"></span>
        <span class="text-[10px] text-emerald-600 font-medium">监测中</span>
      </div>
    </div>

    <!-- Speech Rate Gauge -->
    <div class="p-3 bg-zinc-50/80 rounded-xl border border-zinc-100 space-y-2">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2 text-sm text-zinc-600">
          <Gauge class="w-3.5 h-3.5 text-indigo-500" />
          <span class="font-medium">语速</span>
        </div>
        <div class="flex items-center gap-2">
          <span class="text-lg font-black tracking-tight" :class="speechRateColor">
            {{ Math.round(speechRate) }}
          </span>
          <span class="text-[10px] text-zinc-400">字/分钟</span>
        </div>
      </div>
      <!-- Progress bar with zone markers -->
      <div class="relative">
        <div class="h-2 bg-zinc-200 rounded-full overflow-hidden">
          <div
            :class="speechRateBarColor"
            class="h-full rounded-full transition-all duration-500 ease-out"
            :style="{ width: speechRatePercent + '%' }"
          ></div>
        </div>
        <!-- Zone markers -->
        <div class="flex justify-between mt-1 text-[9px] text-zinc-300 font-medium">
          <span>0</span>
          <span class="text-blue-400">120</span>
          <span class="text-emerald-400">180-220</span>
          <span class="text-rose-400">300+</span>
        </div>
      </div>
      <div class="flex justify-end">
        <span
          class="text-[10px] font-medium px-2 py-0.5 rounded-full"
          :class="{
            'bg-blue-50 text-blue-600': speechRateLevel === 'slow',
            'bg-emerald-50 text-emerald-600': speechRateLevel === 'normal',
            'bg-rose-50 text-rose-600': speechRateLevel === 'fast'
          }"
        >
          {{ speechRateLabel }}
        </span>
      </div>
    </div>

    <!-- Energy Bar -->
    <div class="p-3 bg-zinc-50/80 rounded-xl border border-zinc-100 space-y-2">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2 text-sm text-zinc-600">
          <Volume2 class="w-3.5 h-3.5 text-violet-500" />
          <span class="font-medium">能量</span>
        </div>
        <span
          class="text-[10px] font-medium px-2 py-0.5 rounded-full"
          :class="{
            'bg-zinc-100 text-zinc-500': energyLevel < 0.15,
            'bg-indigo-50 text-indigo-600': energyLevel >= 0.15 && energyLevel < 0.5,
            'bg-emerald-50 text-emerald-600': energyLevel >= 0.5 && energyLevel < 0.8,
            'bg-amber-50 text-amber-600': energyLevel >= 0.8
          }"
        >
          {{ energyLabel }}
        </span>
      </div>
      <!-- Animated energy bars -->
      <div class="flex items-end gap-0.5 h-6">
        <div
          v-for="i in 20"
          :key="i"
          class="flex-1 rounded-sm transition-all duration-150"
          :class="i / 20 <= energyLevel ? energyBarColor : 'bg-zinc-200'"
          :style="{
            height: (i / 20 <= energyLevel)
              ? Math.max(20, Math.min(100, (energyLevel * 100) + Math.sin(i * 0.8) * 15)) + '%'
              : '15%'
          }"
        ></div>
      </div>
    </div>

    <!-- Fluency / Filler Words -->
    <div
      class="p-3 rounded-xl border space-y-1.5 transition-all duration-300"
      :class="fluencyAlert
        ? (alertBlink ? 'bg-rose-50 border-rose-200 shadow-sm shadow-rose-100' : 'bg-rose-50/50 border-rose-100')
        : 'bg-zinc-50/80 border-zinc-100'"
    >
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2 text-sm text-zinc-600">
          <AlertTriangle
            class="w-3.5 h-3.5 transition-colors"
            :class="fluencyAlert ? 'text-rose-500' : 'text-amber-500'"
          />
          <span class="font-medium">流畅度</span>
        </div>
        <div class="flex items-center gap-2">
          <span
            v-if="fluencyAlert"
            class="text-[10px] font-bold px-2 py-0.5 rounded-full bg-rose-100 text-rose-600 animate-pulse"
          >
            注意语气词
          </span>
          <span v-else class="text-[10px] font-medium px-2 py-0.5 rounded-full bg-emerald-50 text-emerald-600">
            良好
          </span>
        </div>
      </div>
      <div class="flex items-center justify-between text-xs text-zinc-500">
        <span>本段助词: <strong class="text-zinc-700">{{ fillerWordCount }}</strong> 个</span>
        <span>累计: <strong class="text-zinc-700">{{ totalFillerWords }}</strong> 个</span>
      </div>
    </div>
  </div>
</template>
