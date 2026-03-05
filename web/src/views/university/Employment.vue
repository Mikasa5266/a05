<template>
  <div class="space-y-8">
    <header>
      <h1 class="text-3xl font-bold tracking-tight text-zinc-900">就业数据分析</h1>
      <p class="text-zinc-500 mt-2">全维度就业数据可视化分析</p>
    </header>

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      <div v-for="metric in metrics" :key="metric.label" class="bg-white rounded-2xl p-5 border border-zinc-100 shadow-sm">
        <div class="text-sm text-zinc-400 mb-2">{{ metric.label }}</div>
        <div class="text-3xl font-bold text-zinc-900">{{ metric.value }}</div>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Employment by Major -->
      <div class="bg-white rounded-3xl p-8 border border-zinc-100 shadow-sm">
        <h2 class="text-lg font-bold text-zinc-900 mb-6">各专业就业情况</h2>
        <div class="space-y-4">
          <div v-for="m in majorData" :key="m.name">
            <div class="flex items-center justify-between text-sm mb-1">
              <span class="text-zinc-600">{{ m.name }}</span>
              <span class="font-bold text-zinc-900">{{ m.employed }}/{{ m.total }}</span>
            </div>
            <div class="h-2 bg-zinc-100 rounded-full overflow-hidden">
              <div class="h-full bg-indigo-500 rounded-full" :style="{ width: (m.employed / m.total * 100) + '%' }"></div>
            </div>
          </div>
        </div>
      </div>

      <!-- Salary Distribution -->
      <div class="bg-white rounded-3xl p-8 border border-zinc-100 shadow-sm">
        <h2 class="text-lg font-bold text-zinc-900 mb-6">薪资分布</h2>
        <div class="space-y-4">
          <div v-for="s in salaryData" :key="s.range">
            <div class="flex items-center justify-between text-sm mb-1">
              <span class="text-zinc-600">{{ s.range }}</span>
              <span class="font-bold text-zinc-900">{{ s.count }}人</span>
            </div>
            <div class="h-2 bg-zinc-100 rounded-full overflow-hidden">
              <div class="h-full bg-emerald-500 rounded-full" :style="{ width: s.percent + '%' }"></div>
            </div>
          </div>
        </div>
      </div>

      <!-- City Distribution -->
      <div class="bg-white rounded-3xl p-8 border border-zinc-100 shadow-sm">
        <h2 class="text-lg font-bold text-zinc-900 mb-6">就业城市分布</h2>
        <div class="grid grid-cols-2 gap-3">
          <div v-for="city in cityData" :key="city.name" class="flex items-center justify-between p-3 rounded-xl bg-zinc-50">
            <span class="text-sm text-zinc-600">{{ city.name }}</span>
            <span class="font-bold text-sm text-zinc-900">{{ city.count }}</span>
          </div>
        </div>
      </div>

      <!-- Industry Distribution -->
      <div class="bg-white rounded-3xl p-8 border border-zinc-100 shadow-sm">
        <h2 class="text-lg font-bold text-zinc-900 mb-6">行业就业分布</h2>
        <div class="space-y-3">
          <div v-for="ind in industryData" :key="ind.name" class="flex items-center justify-between p-3 rounded-xl bg-zinc-50">
            <span class="text-sm text-zinc-600">{{ ind.name }}</span>
            <span class="font-bold text-sm text-indigo-600">{{ ind.percent }}%</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const metrics = ref([
  { label: '总毕业人数', value: '1,247' },
  { label: '已就业人数', value: '1,152' },
  { label: '就业率', value: '92.3%' },
  { label: '平均薪资', value: '¥12.5k' },
])

const majorData = ref([
  { name: '计算机科学', total: 320, employed: 304 },
  { name: '软件工程', total: 280, employed: 258 },
  { name: '人工智能', total: 150, employed: 132 },
  { name: '数据科学', total: 200, employed: 172 },
  { name: '信息安全', total: 120, employed: 96 },
])

const salaryData = ref([
  { range: '5k以下', count: 45, percent: 10 },
  { range: '5k-10k', count: 312, percent: 30 },
  { range: '10k-15k', count: 420, percent: 40 },
  { range: '15k-20k', count: 180, percent: 15 },
  { range: '20k以上', count: 95, percent: 5 },
])

const cityData = ref([
  { name: '北京', count: 285 },
  { name: '上海', count: 240 },
  { name: '深圳', count: 195 },
  { name: '杭州', count: 168 },
  { name: '广州', count: 125 },
  { name: '其他', count: 139 },
])

const industryData = ref([
  { name: '互联网/IT', percent: 45 },
  { name: '金融科技', percent: 18 },
  { name: '人工智能', percent: 15 },
  { name: '通信/电子', percent: 12 },
  { name: '其他', percent: 10 },
])
</script>
