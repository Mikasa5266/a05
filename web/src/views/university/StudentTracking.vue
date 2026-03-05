<template>
  <div class="space-y-8">
    <header>
      <h1 class="text-3xl font-bold tracking-tight text-zinc-900">学生就业全周期跟踪</h1>
      <p class="text-zinc-500 mt-2">分级精准帮扶体系 · 实时监控每位学生就业状态</p>
    </header>

    <!-- Filter Bar -->
    <div class="flex items-center gap-3 flex-wrap">
      <input v-model="search" placeholder="搜索学生姓名..." class="px-4 py-2.5 bg-zinc-50 border border-zinc-200 rounded-xl text-sm w-64 focus:outline-none focus:ring-2 focus:ring-indigo-500" />
      <select v-model="filterRisk" class="px-4 py-2.5 bg-zinc-50 border border-zinc-200 rounded-xl text-sm">
        <option value="">全部风险等级</option>
        <option value="high">高风险</option>
        <option value="medium">中风险</option>
        <option value="low">低风险</option>
      </select>
      <select v-model="filterMajor" class="px-4 py-2.5 bg-zinc-50 border border-zinc-200 rounded-xl text-sm">
        <option value="">全部专业</option>
        <option value="cs">计算机科学</option>
        <option value="se">软件工程</option>
        <option value="ai">人工智能</option>
        <option value="ds">数据科学</option>
      </select>
    </div>

    <!-- Student Table -->
    <div class="bg-white rounded-3xl border border-zinc-100 shadow-sm overflow-hidden">
      <table class="w-full text-left text-sm">
        <thead class="bg-zinc-50 border-b border-zinc-100">
          <tr>
            <th class="px-6 py-4 font-medium text-zinc-500">学生</th>
            <th class="px-6 py-4 font-medium text-zinc-500">专业</th>
            <th class="px-6 py-4 font-medium text-zinc-500">面试次数</th>
            <th class="px-6 py-4 font-medium text-zinc-500">平均分</th>
            <th class="px-6 py-4 font-medium text-zinc-500">风险等级</th>
            <th class="px-6 py-4 font-medium text-zinc-500">就业状态</th>
            <th class="px-6 py-4 font-medium text-zinc-500 text-right">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-zinc-100">
          <tr v-for="(s, idx) in filteredStudents" :key="idx" class="hover:bg-zinc-50/50 transition-colors">
            <td class="px-6 py-4">
              <div class="flex items-center gap-3">
                <div class="h-8 w-8 rounded-full bg-indigo-100 text-indigo-600 flex items-center justify-center font-bold text-xs">{{ s.name.charAt(0) }}</div>
                <span class="font-medium text-zinc-900">{{ s.name }}</span>
              </div>
            </td>
            <td class="px-6 py-4 text-zinc-600">{{ s.major }}</td>
            <td class="px-6 py-4 text-zinc-600">{{ s.interviews }}</td>
            <td class="px-6 py-4">
              <span class="font-bold" :class="s.avgScore >= 80 ? 'text-emerald-600' : s.avgScore >= 60 ? 'text-amber-600' : 'text-rose-600'">{{ s.avgScore }}</span>
            </td>
            <td class="px-6 py-4">
              <span class="px-2 py-1 rounded-full text-xs font-bold"
                :class="s.risk === 'high' ? 'bg-rose-100 text-rose-600' : s.risk === 'medium' ? 'bg-amber-100 text-amber-600' : 'bg-emerald-100 text-emerald-600'">
                {{ s.risk === 'high' ? '高风险' : s.risk === 'medium' ? '中风险' : '低风险' }}
              </span>
            </td>
            <td class="px-6 py-4">
              <span class="text-xs" :class="s.employed ? 'text-emerald-600' : 'text-zinc-400'">{{ s.employed ? '已就业' : '求职中' }}</span>
            </td>
            <td class="px-6 py-4 text-right">
              <button class="text-indigo-600 hover:text-indigo-700 font-medium text-sm">详情</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const search = ref('')
const filterRisk = ref('')
const filterMajor = ref('')

const students = ref([
  { name: '小李', major: '计算机科学', interviews: 1, avgScore: 55, risk: 'high', employed: false },
  { name: '小王', major: '软件工程', interviews: 0, avgScore: 0, risk: 'high', employed: false },
  { name: '小张', major: '信息安全', interviews: 3, avgScore: 68, risk: 'medium', employed: false },
  { name: '小刘', major: '数据科学', interviews: 2, avgScore: 62, risk: 'medium', employed: false },
  { name: '小陈', major: '人工智能', interviews: 8, avgScore: 85, risk: 'low', employed: true },
  { name: '小赵', major: '计算机科学', interviews: 6, avgScore: 78, risk: 'low', employed: false },
  { name: '小孙', major: '软件工程', interviews: 10, avgScore: 92, risk: 'low', employed: true },
  { name: '小钱', major: '数据科学', interviews: 4, avgScore: 72, risk: 'medium', employed: false },
])

const filteredStudents = computed(() => {
  return students.value.filter(s => {
    if (search.value && !s.name.includes(search.value)) return false
    if (filterRisk.value && s.risk !== filterRisk.value) return false
    return true
  })
})
</script>
