<template>
  <div class="space-y-8">
    <!-- Welcome Header -->
    <header>
      <h1 class="text-4xl font-bold tracking-tight text-zinc-900">欢迎回来，同学</h1>
      <p class="text-zinc-400 text-lg mt-2 italic">"机会总是留给有准备的人。"</p>
    </header>

    <!-- Main Content Grid -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Left: Carousel Feature Cards -->
      <div class="lg:col-span-2 relative">
        <div class="overflow-hidden rounded-3xl bg-white border border-zinc-100 shadow-sm">
          <!-- Carousel Slides -->
          <div class="relative min-h-[420px]">
            <!-- Slide: Resume Matching -->
            <transition
              enter-active-class="transition-all duration-500 ease-out"
              leave-active-class="transition-all duration-500 ease-in"
              enter-from-class="opacity-0 translate-x-8"
              enter-to-class="opacity-100 translate-x-0"
              leave-from-class="opacity-100 translate-x-0"
              leave-to-class="opacity-0 -translate-x-8"
            >
              <div v-if="activeSlide === 0" class="absolute inset-0 p-8">
                <div class="flex items-center gap-2 mb-6">
                  <Upload class="h-5 w-5 text-indigo-600" />
                  <h2 class="text-xl font-bold text-zinc-900">简历解析与岗位匹配</h2>
                </div>
                <div
                  class="border-2 border-dashed border-zinc-200 rounded-2xl p-12 flex flex-col items-center justify-center hover:border-indigo-300 transition-colors cursor-pointer"
                  @click="goToResume"
                  @dragover.prevent
                  @drop.prevent="handleDrop"
                >
                  <div class="h-16 w-16 bg-indigo-50 rounded-2xl text-indigo-600 mb-4 flex items-center justify-center">
                    <Upload class="h-8 w-8" />
                  </div>
                  <h3 class="text-lg font-medium text-zinc-700">点击或拖拽上传简历 (PDF/Word)</h3>
                  <p class="text-zinc-400 mt-2 text-sm">AI 将为您智能匹配最合适的岗位</p>
                </div>
              </div>
            </transition>

            <!-- Slide: AI Interview -->
            <transition
              enter-active-class="transition-all duration-500 ease-out"
              leave-active-class="transition-all duration-500 ease-in"
              enter-from-class="opacity-0 translate-x-8"
              enter-to-class="opacity-100 translate-x-0"
              leave-from-class="opacity-100 translate-x-0"
              leave-to-class="opacity-0 -translate-x-8"
            >
              <div v-if="activeSlide === 1" class="absolute inset-0 p-8">
                <div class="flex items-center gap-2 mb-6">
                  <BrainCircuit class="h-5 w-5 text-emerald-600" />
                  <h2 class="text-xl font-bold text-zinc-900">AI 多模态面试</h2>
                </div>
                <div class="grid grid-cols-2 gap-4">
                  <div v-for="mode in interviewModes" :key="mode.key"
                    class="p-5 rounded-2xl border border-zinc-100 hover:border-indigo-200 hover:shadow-sm transition-all cursor-pointer group"
                    @click="startMode(mode.key)"
                  >
                    <component :is="mode.icon" class="h-8 w-8 mb-3 text-zinc-400 group-hover:text-indigo-600 transition-colors" />
                    <h4 class="font-semibold text-zinc-800 text-sm">{{ mode.title }}</h4>
                    <p class="text-xs text-zinc-400 mt-1">{{ mode.desc }}</p>
                  </div>
                </div>
              </div>
            </transition>

            <!-- Slide: Growth Center -->
            <transition
              enter-active-class="transition-all duration-500 ease-out"
              leave-active-class="transition-all duration-500 ease-in"
              enter-from-class="opacity-0 translate-x-8"
              enter-to-class="opacity-100 translate-x-0"
              leave-from-class="opacity-100 translate-x-0"
              leave-to-class="opacity-0 -translate-x-8"
            >
              <div v-if="activeSlide === 2" class="absolute inset-0 p-8">
                <div class="flex items-center gap-2 mb-6">
                  <TrendingUp class="h-5 w-5 text-amber-600" />
                  <h2 class="text-xl font-bold text-zinc-900">复盘与成长</h2>
                </div>
                <div class="grid grid-cols-3 gap-4">
                  <div v-for="f in growthFeatures" :key="f.label"
                    class="text-center p-5 rounded-2xl bg-zinc-50 hover:bg-indigo-50 transition-colors cursor-pointer"
                    @click="f.action"
                  >
                    <component :is="f.icon" class="h-8 w-8 mx-auto mb-3 text-zinc-400" />
                    <h4 class="font-semibold text-zinc-700 text-sm">{{ f.label }}</h4>
                    <p class="text-xs text-zinc-400 mt-1">{{ f.desc }}</p>
                  </div>
                </div>
              </div>
            </transition>
          </div>

          <!-- Carousel Dots -->
          <div class="flex items-center justify-center gap-2 pb-6">
            <button v-for="(_, idx) in 3" :key="idx"
              @click="activeSlide = idx"
              class="w-2 h-2 rounded-full transition-all duration-300"
              :class="activeSlide === idx ? 'w-8 bg-indigo-600' : 'bg-zinc-300 hover:bg-zinc-400'"
            ></button>
          </div>
        </div>

        <!-- Carousel Arrow Buttons -->
        <button @click="prevSlide" class="absolute left-0 top-1/2 -translate-y-1/2 -translate-x-1/2 w-12 h-12 rounded-full bg-zinc-200/80 text-zinc-600 flex items-center justify-center hover:bg-zinc-300 transition-colors shadow-lg backdrop-blur-sm z-10">
          <ChevronLeft class="h-5 w-5" />
        </button>
        <button @click="nextSlide" class="absolute right-0 top-1/2 -translate-y-1/2 translate-x-1/2 w-12 h-12 rounded-full bg-zinc-200/80 text-zinc-600 flex items-center justify-center hover:bg-zinc-300 transition-colors shadow-lg backdrop-blur-sm z-10">
          <ChevronRight class="h-5 w-5" />
        </button>
      </div>

      <!-- Right: Quick Actions & Activity -->
      <div class="space-y-6">
        <!-- Interview Quick Start -->
        <div class="bg-gradient-to-br from-indigo-600 to-violet-600 rounded-3xl p-6 text-white relative overflow-hidden shadow-xl shadow-indigo-200">
          <div class="relative z-10">
            <h3 class="text-lg font-bold mb-1">准备好面试了吗？</h3>
            <p class="text-indigo-200 text-sm mb-5">开启 AI 模拟面试，获取实时反馈与专业评估</p>
            <router-link
              to="/student/interview"
              class="inline-flex items-center gap-2 px-6 py-3 bg-white text-indigo-600 rounded-xl font-bold text-sm hover:bg-indigo-50 transition-colors shadow-lg"
            >
              <Video class="h-4 w-4" />
              进入面试间
            </router-link>
          </div>
          <div class="absolute -bottom-8 -right-8 w-32 h-32 bg-white/10 rounded-full blur-xl"></div>
          <div class="absolute -top-4 -left-4 w-24 h-24 bg-white/5 rounded-full blur-lg"></div>
        </div>

        <!-- Recent Activity -->
        <div class="bg-white rounded-3xl p-6 border border-zinc-100 shadow-sm">
          <h3 class="text-base font-bold text-zinc-900 mb-4 flex items-center gap-2">
            <Clock class="h-4 w-4 text-zinc-400" />
            最近活动
          </h3>
          <div class="space-y-4">
            <div v-if="activities.length === 0" class="text-sm text-zinc-400 py-4 text-center">
              暂无活动记录
            </div>
            <div v-for="(activity, idx) in activities" :key="idx" class="flex items-start gap-3">
              <div class="w-2 h-2 rounded-full mt-2 shrink-0"
                :class="idx === 0 ? 'bg-indigo-600' : 'bg-zinc-300'"
              ></div>
              <div>
                <p class="text-sm font-medium text-zinc-700">{{ activity.title }}</p>
                <p class="text-xs text-zinc-400">{{ activity.time }}</p>
              </div>
            </div>
          </div>
        </div>

        <!-- Ability Overview Mini -->
        <div class="bg-white rounded-3xl p-6 border border-zinc-100 shadow-sm">
          <h3 class="text-base font-bold text-zinc-900 mb-4 flex items-center gap-2">
            <Target class="h-4 w-4 text-zinc-400" />
            能力概览
          </h3>
          <div class="space-y-3">
            <div v-for="skill in abilityOverview" :key="skill.name">
              <div class="flex items-center justify-between text-sm mb-1">
                <span class="text-zinc-600">{{ skill.name }}</span>
                <span class="font-bold text-zinc-900">{{ skill.value }}%</span>
              </div>
              <div class="h-1.5 bg-zinc-100 rounded-full overflow-hidden">
                <div class="h-full rounded-full transition-all duration-700" :class="skill.color" :style="{ width: skill.value + '%' }"></div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import {
  Upload, BrainCircuit, Video, TrendingUp,
  ChevronLeft, ChevronRight, Clock, Target,
  Monitor, Users, Shuffle, UserCheck,
  BarChart3, PlayCircle, BookOpen,
  FileText
} from 'lucide-vue-next'
import { getInterviews } from '../api/interview'
import dayjs from 'dayjs'

const router = useRouter()
const activeSlide = ref(0)
let slideInterval = null

const interviewModes = [
  { key: 'technical', title: 'AI 技术面', desc: '深度技术问答', icon: Monitor },
  { key: 'hr', title: 'AI HR面', desc: '行为与素质评估', icon: UserCheck },
  { key: 'group', title: '无领导小组', desc: '群面模拟', icon: Users },
  { key: 'random', title: '随机模式', desc: '不告知风格', icon: Shuffle },
]

const growthFeatures = [
  { label: '综合报告', desc: '可视化分析', icon: BarChart3, action: () => router.push('/student/history') },
  { label: '面试回放', desc: '对比优化', icon: PlayCircle, action: () => router.push('/student/history') },
  { label: '学习地图', desc: '技能提升', icon: BookOpen, action: () => router.push('/student/growth') },
]

const activities = ref([])
const abilityOverview = ref([
  { name: '技术深度', value: 0, color: 'bg-indigo-500' },
  { name: '表达能力', value: 0, color: 'bg-emerald-500' },
  { name: '逻辑严谨', value: 0, color: 'bg-amber-500' },
  { name: '岗位匹配', value: 0, color: 'bg-rose-500' },
])

const prevSlide = () => { activeSlide.value = (activeSlide.value - 1 + 3) % 3 }
const nextSlide = () => { activeSlide.value = (activeSlide.value + 1) % 3 }

const goToResume = () => router.push('/student/resume')
const startMode = (mode) => router.push({ path: '/student/interview', query: { mode } })

const fetchRecentActivity = async () => {
  try {
    const res = await getInterviews({ page: 1, page_size: 5 })
    const interviews = res?.interviews || []
    activities.value = interviews.map(item => ({
      title: `完成${item.position || ''}面试`,
      time: dayjs(item.created_at).fromNow ? dayjs(item.created_at).format('YYYY-MM-DD HH:mm') : '最近'
    }))
    // Update ability overview from latest interview (mock data for now)
    if (interviews.length > 0) {
      abilityOverview.value = [
        { name: '技术深度', value: 85, color: 'bg-indigo-500' },
        { name: '表达能力', value: 72, color: 'bg-emerald-500' },
        { name: '逻辑严谨', value: 90, color: 'bg-amber-500' },
        { name: '岗位匹配', value: 78, color: 'bg-rose-500' },
      ]
    }
  } catch (e) {
    // silent
  }
}

onMounted(() => {
  fetchRecentActivity()
  slideInterval = setInterval(nextSlide, 5000)
})

onUnmounted(() => {
  if (slideInterval) clearInterval(slideInterval)
})
</script>
