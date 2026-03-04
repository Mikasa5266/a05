<template>
  <div class="space-y-8">
    <!-- Header Section -->
    <header>
      <h1 class="text-4xl font-bold tracking-tight text-zinc-900">欢迎回来，面试官已就绪</h1>
      <p class="text-zinc-500 text-lg mt-2">开启你的 AI 驱动面试成长之旅</p>
    </header>

    <!-- Feature Cards Section -->
    <section class="grid grid-cols-1 md:grid-cols-3 gap-6">
      <!-- Resume Matching Card -->
      <router-link 
        to="/resume" 
        class="group relative overflow-hidden rounded-3xl bg-white p-8 shadow-sm border border-zinc-100 hover:shadow-md transition-all"
      >
        <div class="h-12 w-12 rounded-2xl bg-indigo-50 text-indigo-600 flex items-center justify-center group-hover:scale-110 transition-transform duration-300">
          <FileText class="h-6 w-6" />
        </div>
        <h3 class="mt-6 text-xl font-semibold text-zinc-900">简历匹配</h3>
        <p class="mt-2 text-zinc-500">上传简历，AI 自动解析并匹配最适合的岗位</p>
        <ChevronRight class="absolute bottom-8 right-8 h-6 w-6 text-zinc-400 group-hover:text-indigo-600 group-hover:translate-x-1 transition-all duration-300" />
      </router-link>

      <!-- Mock Interview Card -->
      <router-link 
        to="/interview" 
        class="group relative overflow-hidden rounded-3xl bg-white p-8 shadow-sm border border-zinc-100 hover:shadow-md transition-all"
      >
        <div class="h-12 w-12 rounded-2xl bg-emerald-50 text-emerald-600 flex items-center justify-center group-hover:scale-110 transition-transform duration-300">
          <Video class="h-6 w-6" />
        </div>
        <h3 class="mt-6 text-xl font-semibold text-zinc-900">模拟面试</h3>
        <p class="mt-2 text-zinc-500">沉浸式 AI 模拟面试，实时反馈与评分</p>
        <ChevronRight class="absolute bottom-8 right-8 h-6 w-6 text-zinc-400 group-hover:text-emerald-600 group-hover:translate-x-1 transition-all duration-300" />
      </router-link>

      <!-- Growth Center Card -->
      <router-link 
        to="/growth" 
        class="group relative overflow-hidden rounded-3xl bg-white p-8 shadow-sm border border-zinc-100 hover:shadow-md transition-all"
      >
        <div class="h-12 w-12 rounded-2xl bg-amber-50 text-amber-600 flex items-center justify-center group-hover:scale-110 transition-transform duration-300">
          <TrendingUp class="h-6 w-6" />
        </div>
        <h3 class="mt-6 text-xl font-semibold text-zinc-900">成长中心</h3>
        <p class="mt-2 text-zinc-500">查看能力雷达图与成长曲线，获取学习建议</p>
        <ChevronRight class="absolute bottom-8 right-8 h-6 w-6 text-zinc-400 group-hover:text-amber-600 group-hover:translate-x-1 transition-all duration-300" />
      </router-link>
    </section>

    <!-- Performance Dashboard Section -->
    <section class="rounded-3xl bg-zinc-900 p-8 text-white">
      <div class="flex items-center justify-between mb-6">
        <h2 class="text-2xl font-bold">最近的面试表现</h2>
        <button class="text-sm text-zinc-400 hover:text-white transition-colors">
          查看全部
        </button>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <!-- Technical Depth -->
        <div class="bg-zinc-800 rounded-2xl p-4">
          <div class="text-zinc-400 text-sm mb-2">技术深度</div>
          <div class="text-2xl font-bold mb-3">85%</div>
          <div class="h-1 w-full bg-zinc-700 rounded-full overflow-hidden">
            <div class="h-full bg-indigo-500 rounded-full" style="width: 85%"></div>
          </div>
        </div>

        <!-- Expression -->
        <div class="bg-zinc-800 rounded-2xl p-4">
          <div class="text-zinc-400 text-sm mb-2">表达能力</div>
          <div class="text-2xl font-bold mb-3">72%</div>
          <div class="h-1 w-full bg-zinc-700 rounded-full overflow-hidden">
            <div class="h-full bg-emerald-500 rounded-full" style="width: 72%"></div>
          </div>
        </div>

        <!-- Logic -->
        <div class="bg-zinc-800 rounded-2xl p-4">
          <div class="text-zinc-400 text-sm mb-2">逻辑严谨</div>
          <div class="text-2xl font-bold mb-3">90%</div>
          <div class="h-1 w-full bg-zinc-700 rounded-full overflow-hidden">
            <div class="h-full bg-amber-500 rounded-full" style="width: 90%"></div>
          </div>
        </div>

        <!-- Job Match -->
        <div class="bg-zinc-800 rounded-2xl p-4">
          <div class="text-zinc-400 text-sm mb-2">岗位匹配</div>
          <div class="text-2xl font-bold mb-3">78%</div>
          <div class="h-1 w-full bg-zinc-700 rounded-full overflow-hidden">
            <div class="h-full bg-rose-500 rounded-full" style="width: 78%"></div>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { FileText, Video, TrendingUp, ChevronRight } from 'lucide-vue-next'
</script>
