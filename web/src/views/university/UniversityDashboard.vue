<template>
  <div class="space-y-8">
    <!-- Header -->
    <header class="flex items-end justify-between">
      <div>
        <h1 class="text-3xl font-bold tracking-tight text-zinc-900">高校就业指导智慧管理中台</h1>
        <p class="text-zinc-500 mt-2">全周期跟踪 · 精准帮扶 · 就业数据洞察</p>
      </div>
      <button class="px-5 py-2.5 bg-indigo-600 text-white rounded-xl text-sm font-medium hover:bg-indigo-700 transition-colors shadow-lg shadow-indigo-200">
        导出就业报告
      </button>
    </header>

    <!-- Key Metrics -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
      <div v-for="stat in stats" :key="stat.label" class="bg-white rounded-2xl p-6 border border-zinc-100 shadow-sm">
        <div class="flex items-center justify-between mb-4">
          <div class="h-10 w-10 rounded-xl flex items-center justify-center" :class="stat.bgColor">
            <component :is="stat.icon" class="h-5 w-5" :class="stat.iconColor" />
          </div>
          <span v-if="stat.trend" class="text-xs font-medium px-2 py-1 rounded-full" :class="stat.trend > 0 ? 'bg-emerald-50 text-emerald-600' : 'bg-rose-50 text-rose-600'">
            {{ stat.trend > 0 ? '+' : '' }}{{ stat.trend }}%
          </span>
        </div>
        <div class="text-2xl font-bold text-zinc-900">{{ stat.value }}</div>
        <div class="text-sm text-zinc-400 mt-1">{{ stat.label }}</div>
      </div>
    </div>

    <!-- Main Content -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Employment Risk Alert -->
      <div class="lg:col-span-2 bg-white rounded-3xl p-8 border border-zinc-100 shadow-sm">
        <h2 class="text-lg font-bold text-zinc-900 mb-6 flex items-center gap-2">
          <AlertTriangle class="h-5 w-5 text-amber-600" />
          就业预警 · 分级帮扶
        </h2>
        <div class="space-y-4">
          <div v-for="(student, idx) in atRiskStudents" :key="idx"
            class="flex items-center justify-between p-4 rounded-2xl border transition-colors"
            :class="student.risk === 'high' ? 'border-rose-100 bg-rose-50/30' : student.risk === 'medium' ? 'border-amber-100 bg-amber-50/30' : 'border-zinc-100'"
          >
            <div class="flex items-center gap-4">
              <div class="h-10 w-10 rounded-full flex items-center justify-center font-bold text-sm"
                :class="student.risk === 'high' ? 'bg-rose-100 text-rose-600' : student.risk === 'medium' ? 'bg-amber-100 text-amber-600' : 'bg-zinc-100 text-zinc-600'">
                {{ student.name.charAt(0) }}
              </div>
              <div>
                <div class="font-medium text-zinc-900">{{ student.name }}</div>
                <div class="text-xs text-zinc-400">{{ student.major }} · {{ student.grade }}</div>
              </div>
            </div>
            <div class="flex items-center gap-3">
              <span class="px-2 py-1 rounded-full text-xs font-bold"
                :class="student.risk === 'high' ? 'bg-rose-100 text-rose-600' : student.risk === 'medium' ? 'bg-amber-100 text-amber-600' : 'bg-emerald-100 text-emerald-600'">
                {{ student.risk === 'high' ? '高风险' : student.risk === 'medium' ? '中风险' : '良好' }}
              </span>
              <span class="text-sm text-zinc-500">面试 {{ student.interviews }} 次</span>
              <button class="px-3 py-1.5 bg-indigo-50 text-indigo-600 rounded-lg text-xs font-medium hover:bg-indigo-100 transition-colors">
                帮扶方案
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Right Cards -->
      <div class="space-y-6">
        <!-- Quick Links -->
        <div class="bg-gradient-to-br from-indigo-600 to-violet-600 rounded-3xl p-6 text-white">
          <h3 class="font-bold mb-4">管理功能</h3>
          <div class="space-y-3">
            <router-link to="/university/tracking" class="flex items-center gap-3 p-3 bg-white/10 rounded-xl hover:bg-white/20 transition-colors">
              <Target class="h-5 w-5" />
              <span class="text-sm font-medium">学生跟踪</span>
            </router-link>
            <router-link to="/university/courses" class="flex items-center gap-3 p-3 bg-white/10 rounded-xl hover:bg-white/20 transition-colors">
              <BookOpen class="h-5 w-5" />
              <span class="text-sm font-medium">课程资源共建</span>
            </router-link>
            <router-link to="/university/employment" class="flex items-center gap-3 p-3 bg-white/10 rounded-xl hover:bg-white/20 transition-colors">
              <BarChart3 class="h-5 w-5" />
              <span class="text-sm font-medium">就业数据分析</span>
            </router-link>
          </div>
        </div>

        <!-- Employment Rate Overview -->
        <div class="bg-white rounded-3xl p-6 border border-zinc-100 shadow-sm">
          <h3 class="font-bold text-zinc-900 mb-4">各专业就业率</h3>
          <div class="space-y-3">
            <div v-for="major in majorStats" :key="major.name">
              <div class="flex items-center justify-between text-sm mb-1">
                <span class="text-zinc-600">{{ major.name }}</span>
                <span class="font-bold" :class="major.rate >= 90 ? 'text-emerald-600' : major.rate >= 70 ? 'text-amber-600' : 'text-rose-600'">{{ major.rate }}%</span>
              </div>
              <div class="h-1.5 bg-zinc-100 rounded-full overflow-hidden">
                <div class="h-full rounded-full" :class="major.rate >= 90 ? 'bg-emerald-500' : major.rate >= 70 ? 'bg-amber-500' : 'bg-rose-500'" :style="{ width: major.rate + '%' }"></div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import {
  Users, GraduationCap, BarChart3, TrendingUp,
  AlertTriangle, Target, BookOpen, Briefcase
} from 'lucide-vue-next'

const stats = ref([
  { label: '在籍学生', value: '4,856', icon: Users, bgColor: 'bg-indigo-50', iconColor: 'text-indigo-600', trend: null },
  { label: '本届就业率', value: '92.3%', icon: Briefcase, bgColor: 'bg-emerald-50', iconColor: 'text-emerald-600', trend: 3 },
  { label: '面试总量', value: '12,480', icon: GraduationCap, bgColor: 'bg-amber-50', iconColor: 'text-amber-600', trend: 15 },
  { label: '平均面试分', value: '74.8', icon: TrendingUp, bgColor: 'bg-rose-50', iconColor: 'text-rose-600', trend: 5 },
])

const atRiskStudents = ref([
  { name: '小李', major: '计算机科学', grade: '大四', risk: 'high', interviews: 1 },
  { name: '小王', major: '软件工程', grade: '大四', risk: 'high', interviews: 0 },
  { name: '小张', major: '信息安全', grade: '大四', risk: 'medium', interviews: 3 },
  { name: '小刘', major: '数据科学', grade: '大四', risk: 'medium', interviews: 2 },
  { name: '小陈', major: '人工智能', grade: '大四', risk: 'low', interviews: 8 },
])

const majorStats = ref([
  { name: '计算机科学', rate: 95 },
  { name: '软件工程', rate: 92 },
  { name: '人工智能', rate: 88 },
  { name: '数据科学', rate: 85 },
  { name: '信息安全', rate: 78 },
])
</script>
