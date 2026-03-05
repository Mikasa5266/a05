<template>
  <div class="space-y-8">
    <!-- Header -->
    <header class="flex items-end justify-between">
      <div>
        <h1 class="text-3xl font-bold tracking-tight text-zinc-900">成长中心</h1>
        <p class="text-zinc-500 mt-2">查看您的能力雷达图与成长曲线，获取个性化学习建议</p>
      </div>
    </header>

    <div v-if="loading" class="text-center py-20 text-zinc-500">
      加载数据中...
    </div>

    <!-- Main Content -->
    <div v-else class="grid grid-cols-1 lg:grid-cols-3 gap-8">
      <!-- Left Column: Radar Chart -->
      <div class="lg:col-span-1 bg-white rounded-3xl p-8 shadow-sm border border-zinc-100 flex flex-col">
        <h3 class="text-lg font-bold text-zinc-900 mb-6 flex items-center gap-2">
          <Target class="h-5 w-5 text-indigo-600" />
          能力模型
        </h3>
        
        <div class="flex-1 min-h-[300px] relative">
          <Radar v-if="radarData" :data="radarData" :options="radarOptions" />
          <div v-else class="absolute inset-0 flex items-center justify-center text-zinc-400">
            暂无足够数据
          </div>
        </div>
      </div>

      <!-- Right Column: Growth & Skills -->
      <div class="lg:col-span-2 space-y-8">
        <!-- Growth Curve -->
        <div class="bg-white rounded-3xl p-8 shadow-sm border border-zinc-100">
          <h3 class="text-lg font-bold text-zinc-900 mb-6 flex items-center gap-2">
            <TrendingUp class="h-5 w-5 text-emerald-600" />
            成长趋势
          </h3>
          <div class="h-[250px] w-full relative">
             <Line v-if="lineData" :data="lineData" :options="lineOptions" />
             <div v-else class="absolute inset-0 flex items-center justify-center text-zinc-400">
               完成更多面试以查看趋势
             </div>
          </div>
        </div>

        <!-- Bottom Grid -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
          <!-- Skill Gaps -->
          <div class="bg-white rounded-3xl p-6 shadow-sm border border-zinc-100">
            <h3 class="text-lg font-bold text-zinc-900 mb-4">技能缺口分析</h3>
            <div class="space-y-3" v-if="skillGaps.length > 0">
              <div 
                v-for="gap in skillGaps" 
                :key="gap.name"
                class="flex items-center justify-between p-3 rounded-xl border border-zinc-50 bg-zinc-50/50"
              >
                <span class="text-sm font-medium text-zinc-700">{{ gap.name }}</span>
                <span 
                  class="text-[10px] font-bold px-2 py-1 rounded-lg uppercase tracking-wider"
                  :class="{
                    'bg-rose-100 text-rose-700': gap.level === '急需提升',
                    'bg-amber-100 text-amber-700': gap.level === '中等差距',
                    'bg-emerald-100 text-emerald-700': gap.level === '良好'
                  }"
                >
                  {{ gap.level }}
                </span>
              </div>
            </div>
            <div v-else class="text-sm text-zinc-500 py-4">
              暂无明显技能缺口数据
            </div>
          </div>

          <!-- Learning Map -->
          <div 
            class="bg-zinc-900 text-white rounded-3xl p-6 relative overflow-hidden group cursor-pointer hover:shadow-lg transition-shadow"
            @click="showLearningMap = true"
          >
            <div class="relative z-10">
              <div class="text-xs font-bold text-zinc-400 uppercase tracking-wider mb-2">个性化学习地图</div>
              <h3 class="text-xl font-bold mb-8 leading-tight">基于您的短板生成学习计划</h3>
              <button class="text-sm font-bold text-indigo-400 flex items-center gap-1 group-hover:gap-2 transition-all">
                开始学习 <ChevronRight class="h-4 w-4" />
              </button>
            </div>
            
            <BookOpen class="absolute -bottom-4 -right-4 h-32 w-32 text-white/5 rotate-12 group-hover:rotate-0 group-hover:scale-110 transition-all duration-500" />
          </div>
        </div>

        <!-- Resume Optimization Suggestions -->
        <div class="bg-white rounded-3xl p-6 shadow-sm border border-zinc-100">
          <h3 class="text-lg font-bold text-zinc-900 mb-4 flex items-center gap-2">
            <FileText class="h-5 w-5 text-violet-600" />
            简历优化建议
          </h3>
          <p class="text-sm text-zinc-500 mb-4">基于面试表现与岗位能力图谱，为您生成简历优化建议</p>
          <div class="space-y-3">
            <div v-for="(tip, i) in resumeTips" :key="i" class="flex gap-3 p-3 rounded-xl bg-violet-50/50 border border-violet-100/50">
              <div class="w-6 h-6 rounded-full bg-violet-100 text-violet-600 text-xs flex items-center justify-center font-bold flex-shrink-0 mt-0.5">{{ i + 1 }}</div>
              <p class="text-sm text-zinc-700">{{ tip }}</p>
            </div>
          </div>
          <button @click="generateResumeTips" class="mt-4 w-full py-2.5 bg-violet-600 text-white rounded-xl text-sm font-medium hover:bg-violet-700 transition-colors">
            {{ resumeTipsLoading ? '生成中...' : '重新生成建议' }}
          </button>
        </div>
      </div>
    </div>
  </div>

  <!-- Learning Map Modal -->
  <div v-if="showLearningMap" class="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50" @click.self="showLearningMap = false">
    <div class="bg-white rounded-3xl p-8 w-full max-w-3xl max-h-[80vh] overflow-y-auto shadow-2xl">
      <div class="flex items-center justify-between mb-6">
        <h2 class="text-xl font-bold text-zinc-900">个性化学习地图</h2>
        <button @click="showLearningMap = false" class="p-1 hover:bg-zinc-100 rounded-lg transition-colors">✕</button>
      </div>
      <div class="space-y-4">
        <div v-for="(phase, pi) in learningPhases" :key="pi" class="p-5 rounded-2xl border border-zinc-100">
          <div class="flex items-center gap-3 mb-3">
            <div class="w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold text-white" :class="phase.color">{{ pi + 1 }}</div>
            <div>
              <div class="font-bold text-zinc-900">{{ phase.title }}</div>
              <div class="text-xs text-zinc-400">{{ phase.duration }}</div>
            </div>
          </div>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-2 ml-11">
            <div v-for="(task, ti) in phase.tasks" :key="ti" class="flex items-center gap-2 text-sm text-zinc-600">
              <div class="w-1.5 h-1.5 rounded-full" :class="phase.dotColor"></div>
              {{ task }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getGrowthStats } from '../api/growth'
import { 
  Award, Target, TrendingUp, ChevronRight, BookOpen, FileText 
} from 'lucide-vue-next'
import {
  Chart as ChartJS,
  RadialLinearScale,
  PointElement,
  LineElement,
  Filler,
  Tooltip,
  Legend,
  CategoryScale,
  LinearScale
} from 'chart.js'
import { Radar, Line } from 'vue-chartjs'

ChartJS.register(
  RadialLinearScale,
  PointElement,
  LineElement,
  Filler,
  Tooltip,
  Legend,
  CategoryScale,
  LinearScale
)

const router = useRouter()
const loading = ref(true)
const radarData = ref(null)
const lineData = ref(null)
const skillGaps = ref([])
const showLearningMap = ref(false)
const resumeTips = ref([
  '在项目经验中量化成果，例如"优化接口响应时间降低40%"',
  '增加与目标岗位匹配的技术关键词，提升简历ATS通过率',
  '补充面试中表现突出的沟通协作经历',
])
const resumeTipsLoading = ref(false)

const learningPhases = ref([
  { title: '基础巩固期', duration: '第1-2周', color: 'bg-indigo-500', dotColor: 'bg-indigo-400', tasks: ['数据结构与算法复习', '编程语言核心特性回顾', '常见设计模式学习', '代码规范与最佳实践'] },
  { title: '专项突破期', duration: '第3-4周', color: 'bg-emerald-500', dotColor: 'bg-emerald-400', tasks: ['系统设计与架构思维', '项目经验提炼与表达', '行为面试STAR法则练习', '技术深度问题准备'] },
  { title: '模拟实战期', duration: '第5-6周', color: 'bg-amber-500', dotColor: 'bg-amber-400', tasks: ['每日模拟面试练习', '表达流畅度与逻辑训练', '压力面试应对策略', '综合能力提升冲刺'] },
])

const radarOptions = {
  responsive: true,
  maintainAspectRatio: false,
  scales: {
    r: {
      angleLines: { color: '#f4f4f5' },
      grid: { color: '#f4f4f5' },
      pointLabels: { color: '#71717a', font: { size: 12 } },
      ticks: { display: false, stepSize: 20 },
      suggestedMin: 0,
      suggestedMax: 100
    }
  },
  plugins: { legend: { display: false } }
}

const lineOptions = {
  responsive: true,
  maintainAspectRatio: false,
  scales: {
    x: { grid: { display: false }, ticks: { color: '#a1a1aa' } },
    y: { display: false, suggestedMin: 50, suggestedMax: 100 }
  },
  plugins: { legend: { display: false } }
}

const generateResumeTips = async () => {
  resumeTipsLoading.value = true
  // Simulate API call for resume optimization suggestions
  setTimeout(() => {
    resumeTips.value = [
      '突出面试中展现的技术亮点，例如分布式系统设计经验',
      '将项目描述与目标岗位JD高度对齐，减少无关经历',
      '增加数据驱动的成果描述，如"负责的模块覆盖率从60%提升至95%"',
      '补充开源贡献或技术博客链接，增强技术影响力背书',
    ]
    resumeTipsLoading.value = false
  }, 1500)
}

const fetchGrowthData = async () => {
  try {
    const res = await getGrowthStats()
    const stats = res.stats || res
    
    // Process Radar Data
    if (stats.radar_data && stats.radar_data.length > 0) {
      radarData.value = {
        labels: stats.radar_data.map(item => item.subject),
        datasets: [{
          label: '当前能力',
          backgroundColor: 'rgba(79, 70, 229, 0.2)',
          borderColor: '#4f46e5',
          pointBackgroundColor: '#4f46e5',
          pointBorderColor: '#fff',
          pointHoverBackgroundColor: '#fff',
          pointHoverBorderColor: '#4f46e5',
          data: stats.radar_data.map(item => item.A)
        }]
      }
    }

    // Process Growth Data
    if (stats.growth_data && stats.growth_data.length > 0) {
      lineData.value = {
        labels: stats.growth_data.map(item => item.name),
        datasets: [{
          label: '综合评分',
          backgroundColor: '#10b981',
          borderColor: '#10b981',
          borderWidth: 3,
          pointBackgroundColor: '#10b981',
          pointBorderColor: '#fff',
          pointHoverBackgroundColor: '#fff',
          pointHoverBorderColor: '#10b981',
          data: stats.growth_data.map(item => item.score),
          tension: 0.4
        }]
      }
    }

    // Process Skill Gaps
    skillGaps.value = stats.skill_gaps || []

  } catch (error) {
    console.error('Failed to fetch growth stats:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchGrowthData()
})
</script>
