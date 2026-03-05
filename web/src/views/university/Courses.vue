<template>
  <div class="space-y-8">
    <header>
      <h1 class="text-3xl font-bold tracking-tight text-zinc-900">就业指导课程与资源共建</h1>
      <p class="text-zinc-500 mt-2">结合企业需求与AI分析，共建就业指导课程体系</p>
    </header>

    <!-- Course Cards -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div v-for="(course, idx) in courses" :key="idx" class="bg-white rounded-3xl border border-zinc-100 shadow-sm overflow-hidden hover:shadow-md transition-shadow">
        <div class="h-40 bg-gradient-to-br flex items-center justify-center" :class="course.gradient">
          <component :is="course.icon" class="h-16 w-16 text-white/40" />
        </div>
        <div class="p-6">
          <div class="flex items-center gap-2 mb-2">
            <span class="px-2 py-0.5 bg-indigo-50 text-indigo-600 rounded text-xs font-medium">{{ course.category }}</span>
            <span class="text-xs text-zinc-400">{{ course.duration }}</span>
          </div>
          <h3 class="font-bold text-zinc-900 mb-2">{{ course.title }}</h3>
          <p class="text-sm text-zinc-500 mb-4">{{ course.desc }}</p>
          <div class="flex items-center justify-between">
            <span class="text-xs text-zinc-400">{{ course.students }} 人已学习</span>
            <button class="px-4 py-2 bg-indigo-600 text-white rounded-xl text-xs font-medium hover:bg-indigo-700 transition-colors">查看课程</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Resource Library -->
    <div class="bg-white rounded-3xl p-8 border border-zinc-100 shadow-sm">
      <h2 class="text-lg font-bold text-zinc-900 mb-6">资源库</h2>
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div v-for="(res, idx) in resources" :key="idx" class="p-4 rounded-2xl border border-zinc-100 hover:border-indigo-200 transition-colors cursor-pointer text-center">
          <div class="text-2xl font-bold text-indigo-600 mb-1">{{ res.count }}</div>
          <div class="text-sm text-zinc-600">{{ res.label }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { BookOpen, Video, FileText, Users, Code, MessageSquare } from 'lucide-vue-next'

const courses = ref([
  { title: '技术面试通关指南', desc: '系统化的技术面试准备方法论', category: '面试技巧', duration: '12课时', students: 1240, icon: Code, gradient: 'from-indigo-500 to-violet-500' },
  { title: 'HR面试高频问题', desc: '行为面试STAR法则与案例', category: '面试技巧', duration: '8课时', students: 980, icon: MessageSquare, gradient: 'from-emerald-500 to-teal-500' },
  { title: '简历优化工坊', desc: '针对不同岗位的简历优化策略', category: '求职准备', duration: '6课时', students: 1560, icon: FileText, gradient: 'from-amber-500 to-orange-500' },
  { title: '行业认知系列', desc: '各技术领域行业趋势与岗位分析', category: '行业认知', duration: '10课时', students: 750, icon: BookOpen, gradient: 'from-rose-500 to-pink-500' },
  { title: '群面模拟实战', desc: '无领导小组讨论技巧与实战', category: '面试技巧', duration: '4课时', students: 620, icon: Users, gradient: 'from-cyan-500 to-blue-500' },
  { title: '模拟面试回放赏析', desc: '优秀面试案例解析与学习', category: '面试示范', duration: '15课时', students: 890, icon: Video, gradient: 'from-violet-500 to-purple-500' },
])

const resources = ref([
  { label: '视频教程', count: 128 },
  { label: '面经文档', count: 456 },
  { label: '题库资源', count: 2340 },
  { label: '行业报告', count: 67 },
])
</script>
