<template>
  <div class="space-y-8">
    <header>
      <h1 class="text-3xl font-bold tracking-tight text-zinc-900">精准帮扶体系</h1>
      <p class="text-zinc-500 mt-2">针对不同风险等级学生提供差异化就业帮扶方案</p>
    </header>

    <!-- Risk Level Groups -->
    <div v-for="group in riskGroups" :key="group.level" class="bg-white rounded-3xl p-8 border border-zinc-100 shadow-sm">
      <div class="flex items-center justify-between mb-6">
        <h2 class="text-lg font-bold text-zinc-900 flex items-center gap-2">
          <div class="w-3 h-3 rounded-full" :class="group.dotColor"></div>
          {{ group.title }}
          <span class="text-sm font-normal text-zinc-400">（{{ group.students.length }}人）</span>
        </h2>
        <button class="px-4 py-2 bg-indigo-50 text-indigo-600 rounded-xl text-sm font-medium hover:bg-indigo-100 transition-colors">
          批量帮扶
        </button>
      </div>

      <div class="space-y-3">
        <div v-for="(s, idx) in group.students" :key="idx" class="flex items-center justify-between p-4 rounded-2xl border border-zinc-100 hover:bg-zinc-50 transition-colors">
          <div class="flex items-center gap-4">
            <div class="h-10 w-10 rounded-full bg-zinc-100 text-zinc-600 flex items-center justify-center font-bold text-sm">{{ s.name.charAt(0) }}</div>
            <div>
              <div class="font-medium text-zinc-900">{{ s.name }}</div>
              <div class="text-xs text-zinc-400">{{ s.major }}</div>
            </div>
          </div>
          <div class="flex items-center gap-4">
            <div class="text-right">
              <div class="text-sm text-zinc-500">{{ s.issue }}</div>
            </div>
            <div class="flex items-center gap-2">
              <button class="px-3 py-1.5 bg-indigo-600 text-white rounded-lg text-xs font-medium hover:bg-indigo-700 transition-colors">指派导师</button>
              <button class="px-3 py-1.5 border border-zinc-200 text-zinc-600 rounded-lg text-xs hover:bg-zinc-50 transition-colors">推荐课程</button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const riskGroups = ref([
  {
    level: 'high',
    title: '高风险学生',
    dotColor: 'bg-rose-500',
    students: [
      { name: '小李', major: '计算机科学', issue: '面试次数过少，技术能力薄弱' },
      { name: '小王', major: '软件工程', issue: '从未参加模拟面试' },
    ]
  },
  {
    level: 'medium',
    title: '中风险学生',
    dotColor: 'bg-amber-500',
    students: [
      { name: '小张', major: '信息安全', issue: '表达能力需提升' },
      { name: '小刘', major: '数据科学', issue: '技术深度不够' },
      { name: '小钱', major: '数据科学', issue: '面试经验不足' },
    ]
  },
  {
    level: 'low',
    title: '低风险学生',
    dotColor: 'bg-emerald-500',
    students: [
      { name: '小陈', major: '人工智能', issue: '状态良好' },
      { name: '小赵', major: '计算机科学', issue: '继续保持' },
      { name: '小孙', major: '软件工程', issue: '已获得Offer' },
    ]
  }
])
</script>
