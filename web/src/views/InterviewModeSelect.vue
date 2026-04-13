<template>
  <div class="mode-select-page w-full px-4 py-4 md:min-h-[calc(100vh-10rem)] md:flex md:items-center md:justify-center md:px-6 md:py-5 rounded-3xl border border-indigo-100/70">
    <div class="w-full max-w-7xl">
      <div class="text-center mb-8 md:mb-10">
        <p class="inline-flex items-center px-4 py-1.5 rounded-full text-sm font-medium bg-indigo-50 text-indigo-600 border border-indigo-100">
          面试系统大厅
        </p>
        <h1 class="mt-6 text-4xl md:text-5xl font-bold tracking-tight text-zinc-900">选择你的面试模式</h1>
        <p class="mt-4 text-base md:text-lg text-zinc-500">根据目标场景进入不同流程，开始本次训练与对弈</p>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6">
        <button
          v-for="item in cards"
          :key="item.title"
          type="button"
          class="group text-left rounded-3xl p-7 bg-white border border-zinc-100 shadow-sm transition-all duration-300"
          :class="item.enabled
            ? 'hover:-translate-y-1 hover:shadow-[0_14px_28px_rgba(15,23,42,0.08)] hover:border-zinc-200 cursor-pointer'
            : 'opacity-80 cursor-not-allowed'"
          @click="goTo(item)"
        >
          <div class="w-14 h-14 rounded-2xl flex items-center justify-center" :class="item.iconWrapClass">
            <component :is="item.icon" class="w-7 h-7" :class="item.iconClass" />
          </div>
          <h2 class="mt-6 text-2xl font-bold text-zinc-900">{{ item.title }}</h2>
          <p class="mt-3 text-sm leading-6 text-zinc-500">{{ item.description }}</p>
          <div class="mt-8 inline-flex items-center text-sm font-semibold" :class="item.enabled ? 'text-indigo-600 group-hover:text-indigo-500' : 'text-zinc-400'">
            {{ item.enabled ? '进入场景' : '即将上线' }}
            <ChevronRight v-if="item.enabled" class="w-4 h-4 ml-1.5 transition-transform duration-300 group-hover:translate-x-1" />
          </div>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ChevronRight, Bot, UserCheck, Users, Package } from 'lucide-vue-next'

const router = useRouter()

const cards = [
  {
    title: 'AI 模拟面试',
    description: '进入 Gemini 驱动的 AI 全真模拟面试，激活 3D 面试官，还原真实面试全流程。',
    icon: Bot,
    iconWrapClass: 'bg-indigo-100',
    iconClass: 'text-indigo-600',
    enabled: true,
    path: '/interview/video',
    query: {
      mode: 'technical',
      style: 'gentle',
      interviewMode: 'ai',
      presentationMode: 'video_avatar'
    }
  },
  {
    title: '真人面试',
    description: '进入真人面试工作台，发起面试邀请、实时音视频连线，同步面试状态与记录。',
    icon: UserCheck,
    iconWrapClass: 'bg-sky-100',
    iconClass: 'text-sky-700',
    enabled: true,
    path: '/interview/live/workbench',
    query: {}
  },
  {
    title: '群面模式',
    description: '进入 AI 群面模拟场景，还原无领导小组讨论流程，支持多角色 AI 队友/面试官。',
    icon: Users,
    iconWrapClass: 'bg-emerald-100',
    iconClass: 'text-emerald-600',
    enabled: true,
    path: '/interview/live/workbench',
    query: {
      group_mode: '1',
      source: 'group_mode'
    }
  },
  {
    title: '盲盒模式',
    description: '进入随机盲盒面试，随机匹配岗位、面试官与面试题型，沉浸式模拟未知场景。',
    icon: Package,
    iconWrapClass: 'bg-amber-100',
    iconClass: 'text-amber-700',
    enabled: true,
    path: '/interview/standard/setup',
    query: {
      mode: 'blindbox',
      style: 'gentle',
      interviewMode: 'ai',
      presentationMode: 'text_voice'
    }
  }
]

const goTo = (item) => {
  if (!item.enabled || !item.path) {
    ElMessage.info('该模式正在开发中，敬请期待')
    return
  }
  router.push({
    path: item.path,
    query: item.query
  })
}
</script>

<style scoped>
.mode-select-page {
  background:
    radial-gradient(circle at 14% 12%, rgba(99, 102, 241, 0.2), transparent 42%),
    radial-gradient(circle at 84% 16%, rgba(56, 189, 248, 0.18), transparent 40%),
    radial-gradient(circle at 54% 90%, rgba(14, 165, 233, 0.12), transparent 38%),
    linear-gradient(155deg, #f9fbff 0%, #f3f7ff 46%, #eef6ff 100%);
}
</style>
