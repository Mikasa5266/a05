<template>
  <div class="space-y-8">
    <!-- Header -->
    <header class="flex items-end justify-between">
      <div>
        <h1 class="text-3xl font-bold tracking-tight text-zinc-900">企业端 · 校企直通平台</h1>
        <p class="text-zinc-500 mt-2">人才双选闭环 · 精准人才推送 · 岗位标准共建</p>
      </div>
      <button class="px-5 py-2.5 bg-indigo-600 text-white rounded-xl text-sm font-medium hover:bg-indigo-700 transition-colors shadow-lg shadow-indigo-200">
        发布新岗位
      </button>
    </header>

    <!-- Stats Overview -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
      <div v-for="stat in stats" :key="stat.label" class="bg-white rounded-2xl p-6 border border-zinc-100 shadow-sm">
        <div class="flex items-center justify-between mb-4">
          <div class="h-10 w-10 rounded-xl flex items-center justify-center" :class="stat.bgColor">
            <component :is="stat.icon" class="h-5 w-5" :class="stat.iconColor" />
          </div>
          <span class="text-xs font-medium px-2 py-1 rounded-full" :class="stat.trend > 0 ? 'bg-emerald-50 text-emerald-600' : 'bg-rose-50 text-rose-600'">
            {{ stat.trend > 0 ? '+' : '' }}{{ stat.trend }}%
          </span>
        </div>
        <div class="text-2xl font-bold text-zinc-900">{{ stat.value }}</div>
        <div class="text-sm text-zinc-400 mt-1">{{ stat.label }}</div>
      </div>
    </div>

    <!-- Main Content -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Talent Certification -->
      <div class="lg:col-span-2 bg-white rounded-3xl p-8 border border-zinc-100 shadow-sm">
        <h2 class="text-lg font-bold text-zinc-900 mb-6 flex items-center gap-2">
          <Award class="h-5 w-5 text-indigo-600" />
          能力达标认证 · 人才背书
        </h2>
        <div class="space-y-4">
          <div v-for="(candidate, idx) in certifiedCandidates" :key="idx" 
            class="flex items-center justify-between p-4 rounded-2xl border border-zinc-100 hover:bg-zinc-50 transition-colors"
          >
            <div class="flex items-center gap-4">
              <div class="h-10 w-10 rounded-full bg-indigo-100 flex items-center justify-center text-indigo-600 font-bold text-sm">
                {{ candidate.name.charAt(0) }}
              </div>
              <div>
                <div class="font-medium text-zinc-900">{{ candidate.name }}</div>
                <div class="text-xs text-zinc-400">{{ candidate.school }} · {{ candidate.major }}</div>
              </div>
            </div>
            <div class="flex items-center gap-3">
              <div class="text-right">
                <div class="text-sm font-bold" :class="candidate.score >= 85 ? 'text-emerald-600' : 'text-amber-600'">{{ candidate.score }}分</div>
                <div class="text-xs text-zinc-400">综合评分</div>
              </div>
              <span class="px-3 py-1 rounded-full text-xs font-bold" :class="candidate.certified ? 'bg-emerald-50 text-emerald-600' : 'bg-zinc-100 text-zinc-500'">
                {{ candidate.certified ? '已认证' : '评估中' }}
              </span>
              <button class="px-3 py-1.5 bg-indigo-50 text-indigo-600 rounded-lg text-xs font-medium hover:bg-indigo-100 transition-colors">
                查看详情
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Right Sidebar -->
      <div class="space-y-6">
        <!-- Quick Actions -->
        <div class="bg-gradient-to-br from-indigo-600 to-violet-600 rounded-3xl p-6 text-white">
          <h3 class="font-bold mb-4">快捷操作</h3>
          <div class="space-y-3">
            <router-link to="/enterprise/talent" class="flex items-center gap-3 p-3 bg-white/10 rounded-xl hover:bg-white/20 transition-colors">
              <UserCheck class="h-5 w-5" />
              <span class="text-sm font-medium">查看人才池</span>
            </router-link>
            <router-link to="/enterprise/hr-panel" class="flex items-center gap-3 p-3 bg-white/10 rounded-xl hover:bg-white/20 transition-colors">
              <Video class="h-5 w-5" />
              <span class="text-sm font-medium">HR面试台</span>
            </router-link>
            <router-link to="/enterprise/standards" class="flex items-center gap-3 p-3 bg-white/10 rounded-xl hover:bg-white/20 transition-colors">
              <Database class="h-5 w-5" />
              <span class="text-sm font-medium">岗位标准共建</span>
            </router-link>
          </div>
        </div>

        <!-- Internal Referrals -->
        <div class="bg-white rounded-3xl p-6 border border-zinc-100 shadow-sm">
          <h3 class="font-bold text-zinc-900 mb-4 flex items-center gap-2">
            <Send class="h-4 w-4 text-emerald-600" />
            内推绿色通道
          </h3>
          <p class="text-sm text-zinc-500 mb-4">已有 <span class="font-bold text-emerald-600">{{ referralCount }}</span> 位候选人通过内推进入面试</p>
          <button class="w-full py-2.5 border border-indigo-200 text-indigo-600 rounded-xl text-sm font-medium hover:bg-indigo-50 transition-colors">
            管理内推通道
          </button>
        </div>

        <!-- Talent Cultivation Feedback -->
        <div class="bg-white rounded-3xl p-6 border border-zinc-100 shadow-sm">
          <h3 class="font-bold text-zinc-900 mb-4 flex items-center gap-2">
            <RefreshCw class="h-4 w-4 text-amber-600" />
            人才培养反向赋能
          </h3>
          <p class="text-sm text-zinc-500 mb-3">基于企业面试数据反馈，帮助高校优化人才培养方案</p>
          <div class="space-y-2">
            <div class="flex items-center justify-between text-sm">
              <span class="text-zinc-600">已反馈技能缺口</span>
              <span class="font-bold text-zinc-900">12 项</span>
            </div>
            <div class="flex items-center justify-between text-sm">
              <span class="text-zinc-600">课程优化建议</span>
              <span class="font-bold text-zinc-900">8 条</span>
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
  Award, UserCheck, Video, Database, Send, RefreshCw,
  Users, Briefcase, BarChart3, TrendingUp
} from 'lucide-vue-next'

const stats = ref([
  { label: '人才池总量', value: '2,847', icon: Users, bgColor: 'bg-indigo-50', iconColor: 'text-indigo-600', trend: 12 },
  { label: '活跃岗位', value: '34', icon: Briefcase, bgColor: 'bg-emerald-50', iconColor: 'text-emerald-600', trend: 5 },
  { label: '本月面试', value: '156', icon: Video, bgColor: 'bg-amber-50', iconColor: 'text-amber-600', trend: 18 },
  { label: '匹配率', value: '76%', icon: TrendingUp, bgColor: 'bg-rose-50', iconColor: 'text-rose-600', trend: 3 },
])

const certifiedCandidates = ref([
  { name: '张三', school: '北京大学', major: '计算机科学', score: 92, certified: true },
  { name: '李四', school: '清华大学', major: '软件工程', score: 88, certified: true },
  { name: '王五', school: '浙江大学', major: '人工智能', score: 79, certified: false },
  { name: '赵六', school: '复旦大学', major: '数据科学', score: 85, certified: true },
])

const referralCount = ref(23)
</script>
