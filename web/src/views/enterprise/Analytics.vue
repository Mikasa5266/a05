<template>
  <div class="space-y-8">
    <header>
      <h1 class="text-3xl font-bold tracking-tight text-zinc-900">数据分析</h1>
      <p class="text-zinc-500 mt-2">招聘全流程数据洞察</p>
    </header>

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      <div v-for="metric in metrics" :key="metric.label" class="bg-white rounded-2xl p-5 border border-zinc-100 shadow-sm">
        <div class="text-sm text-zinc-400 mb-2">{{ metric.label }}</div>
        <div class="text-3xl font-bold text-zinc-900">{{ metric.value }}</div>
        <div class="text-xs mt-2" :class="metric.change > 0 ? 'text-emerald-600' : 'text-rose-600'">
          {{ metric.change > 0 ? '↑' : '↓' }} {{ Math.abs(metric.change) }}% 较上月
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Funnel -->
      <div class="bg-white rounded-3xl p-8 border border-zinc-100 shadow-sm">
        <h2 class="text-lg font-bold text-zinc-900 mb-6">招聘漏斗</h2>
        <div class="space-y-4">
          <div v-for="(stage, idx) in funnel" :key="idx" class="flex items-center gap-4">
            <div class="w-24 text-sm text-zinc-500 text-right shrink-0">{{ stage.name }}</div>
            <div class="flex-1 h-8 bg-zinc-100 rounded-lg overflow-hidden relative">
              <div class="h-full bg-indigo-500 rounded-lg transition-all duration-700" :style="{ width: stage.percent + '%' }"></div>
              <span class="absolute inset-0 flex items-center justify-center text-xs font-bold" :class="stage.percent > 50 ? 'text-white' : 'text-zinc-600'">{{ stage.count }}人</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Quality Overview -->
      <div class="bg-white rounded-3xl p-8 border border-zinc-100 shadow-sm">
        <h2 class="text-lg font-bold text-zinc-900 mb-6">候选人质量分布</h2>
        <div class="space-y-4">
          <div v-for="q in qualityDist" :key="q.label" class="flex items-center gap-4">
            <div class="w-20 text-sm text-zinc-500 text-right">{{ q.label }}</div>
            <div class="flex-1 h-6 bg-zinc-100 rounded-full overflow-hidden">
              <div class="h-full rounded-full" :class="q.color" :style="{ width: q.percent + '%' }"></div>
            </div>
            <span class="text-sm font-bold text-zinc-700 w-12">{{ q.percent }}%</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const metrics = ref([
  { label: '简历投递量', value: '1,247', change: 15 },
  { label: '面试完成率', value: '78%', change: 5 },
  { label: '平均面试分', value: '76.5', change: 3 },
  { label: '录用转化率', value: '23%', change: -2 },
])

const funnel = ref([
  { name: '简历投递', count: 1247, percent: 100 },
  { name: '初筛通过', count: 680, percent: 55 },
  { name: '笔试完成', count: 420, percent: 34 },
  { name: '面试完成', count: 280, percent: 22 },
  { name: '发出Offer', count: 85, percent: 7 },
])

const qualityDist = ref([
  { label: '优秀', percent: 25, color: 'bg-emerald-500' },
  { label: '良好', percent: 40, color: 'bg-indigo-500' },
  { label: '一般', percent: 25, color: 'bg-amber-500' },
  { label: '不足', percent: 10, color: 'bg-rose-500' },
])
</script>
