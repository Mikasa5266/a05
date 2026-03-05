<template>
  <div class="space-y-8">
    <header>
      <h1 class="text-3xl font-bold tracking-tight text-zinc-900">HR 面试台</h1>
      <p class="text-zinc-500 mt-2">定制化面试场景 · 实时面试监控与评估</p>
    </header>

    <!-- Interview Sessions -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div v-for="(session, idx) in sessions" :key="idx"
        class="bg-white rounded-3xl border border-zinc-100 shadow-sm overflow-hidden hover:shadow-md transition-shadow"
      >
        <!-- Video Preview Area -->
        <div class="aspect-video bg-zinc-900 relative flex items-center justify-center">
          <div class="text-zinc-600 text-sm">候选人视频区域</div>
          <div class="absolute top-3 left-3 px-2 py-1 rounded-full text-xs font-bold"
            :class="session.status === 'live' ? 'bg-rose-500 text-white' : 'bg-zinc-700 text-zinc-300'">
            {{ session.status === 'live' ? '● 进行中' : '待开始' }}
          </div>
        </div>
        <div class="p-5">
          <div class="flex items-center justify-between mb-3">
            <div>
              <div class="font-bold text-zinc-900">{{ session.candidate }}</div>
              <div class="text-xs text-zinc-400">{{ session.position }} · {{ session.time }}</div>
            </div>
            <span class="text-sm font-bold" :class="session.score ? 'text-indigo-600' : 'text-zinc-400'">
              {{ session.score ? session.score + '分' : '--' }}
            </span>
          </div>
          <div class="flex items-center gap-2">
            <button v-if="session.status === 'live'" class="flex-1 py-2 bg-rose-600 text-white rounded-xl text-sm font-medium hover:bg-rose-700 transition-colors">
              进入面试
            </button>
            <button v-else class="flex-1 py-2 bg-indigo-600 text-white rounded-xl text-sm font-medium hover:bg-indigo-700 transition-colors">
              开始面试
            </button>
            <button class="px-3 py-2 border border-zinc-200 text-zinc-600 rounded-xl text-sm hover:bg-zinc-50 transition-colors">
              查看
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Custom Scenario -->
    <div class="bg-white rounded-3xl p-8 border border-zinc-100 shadow-sm">
      <h2 class="text-lg font-bold text-zinc-900 mb-4">定制化面试场景</h2>
      <p class="text-sm text-zinc-500 mb-6">根据企业需求定制面试流程、题库和评分标准</p>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div class="p-5 rounded-2xl border border-zinc-100 hover:border-indigo-200 transition-colors cursor-pointer">
          <h4 class="font-semibold text-zinc-800 mb-2">技术笔试</h4>
          <p class="text-xs text-zinc-400">自定义编程题库与评测标准</p>
        </div>
        <div class="p-5 rounded-2xl border border-zinc-100 hover:border-indigo-200 transition-colors cursor-pointer">
          <h4 class="font-semibold text-zinc-800 mb-2">行为面试</h4>
          <p class="text-xs text-zinc-400">企业文化与价值观匹配评估</p>
        </div>
        <div class="p-5 rounded-2xl border border-zinc-100 hover:border-indigo-200 transition-colors cursor-pointer">
          <h4 class="font-semibold text-zinc-800 mb-2">综合评估</h4>
          <p class="text-xs text-zinc-400">全链路笔试+面试联动评估</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const sessions = ref([
  { candidate: '张明', position: 'Java后端', time: '今天 14:00', status: 'live', score: null },
  { candidate: '李华', position: '前端开发', time: '今天 15:30', status: 'pending', score: null },
  { candidate: '王芳', position: 'AI工程师', time: '昨天 10:00', status: 'completed', score: 88 },
  { candidate: '刘洋', position: 'Go后端', time: '昨天 14:00', status: 'completed', score: 82 },
])
</script>
