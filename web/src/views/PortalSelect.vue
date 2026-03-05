<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-zinc-50 via-indigo-50/30 to-violet-50/30 relative overflow-hidden">
    <!-- Background decoration -->
    <div class="absolute top-0 left-0 w-[600px] h-[600px] bg-indigo-200/20 rounded-full -translate-x-1/2 -translate-y-1/2 blur-3xl"></div>
    <div class="absolute bottom-0 right-0 w-[500px] h-[500px] bg-violet-200/20 rounded-full translate-x-1/3 translate-y-1/3 blur-3xl"></div>

    <div class="w-full max-w-4xl mx-4 relative z-10 flex flex-col items-center">
      <!-- Logo & Title -->
      <div class="text-center mb-10">
        <div class="inline-flex items-center justify-center w-16 h-16 bg-gradient-to-br from-indigo-600 to-violet-600 rounded-2xl shadow-xl shadow-indigo-200 mb-4">
          <span class="text-2xl font-black text-white">AI</span>
        </div>
        <h1 class="text-3xl font-bold text-zinc-900">智聘 AI</h1>
        <p class="text-zinc-500 mt-2 text-sm">AI 驱动的智能面试与人才评估平台</p>
      </div>

      <!-- Portal Cards -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-5 w-full">
        <router-link
          v-for="portal in portals"
          :key="portal.key"
          :to="`/${portal.key}/login`"
          class="group bg-white/80 backdrop-blur-xl rounded-3xl shadow-lg shadow-zinc-200/50 border border-white/50 p-8 text-center hover:shadow-xl hover:-translate-y-1 transition-all duration-300 cursor-pointer flex flex-col items-center"
        >
          <div
            class="inline-flex items-center justify-center w-14 h-14 rounded-2xl mb-5 transition-colors duration-200"
            :class="portal.bgClass"
          >
            <component :is="portal.icon" class="h-7 w-7" :class="portal.iconClass" />
          </div>
          <h3 class="text-lg font-bold text-zinc-900 mb-2">{{ portal.label }}</h3>
          <p class="text-sm text-zinc-500 leading-relaxed">{{ portal.desc }}</p>
        </router-link>
      </div>

      <!-- Footer -->
      <p class="text-center text-xs text-zinc-400 mt-8">智聘 AI &copy; 2025 · AI 驱动面试评估平台</p>
    </div>
  </div>
</template>

<script setup>
import { User, Building2, GraduationCap } from 'lucide-vue-next'

const portals = [
  {
    key: 'student',
    label: '学生端',
    desc: '模拟面试、简历匹配、成长追踪',
    icon: User,
    bgClass: 'bg-indigo-50 group-hover:bg-indigo-100',
    iconClass: 'text-indigo-600'
  },
  {
    key: 'enterprise',
    label: '企业端',
    desc: '人才筛选、岗位管理、数据分析',
    icon: Building2,
    bgClass: 'bg-emerald-50 group-hover:bg-emerald-100',
    iconClass: 'text-emerald-600'
  },
  {
    key: 'university',
    label: '高校端',
    desc: '学生跟踪、就业统计、人才推送',
    icon: GraduationCap,
    bgClass: 'bg-amber-50 group-hover:bg-amber-100',
    iconClass: 'text-amber-600'
  }
]
</script>
