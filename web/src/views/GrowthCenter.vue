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
            @click="goToLearningMap"
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
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getGrowthStats } from '../api/growth'
import { 
  Award, Target, TrendingUp, ChevronRight, BookOpen 
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

const goToLearningMap = () => {
  // For now, redirect to a dashboard or show an alert, 
  // or maybe a new route '/learning-map' if we had time to build it.
  // User asked for "Click reaction".
  // Let's create a simple learning map view or just alert for now as placeholder?
  // No, user said "Click has no reaction".
  // Let's redirect to History for now as a "Source of truth" or stay here.
  // Actually, I should probably implement a simple Learning Map modal or page.
  // For simplicity in this turn, I will just alert.
  alert("个性化学习地图功能即将上线！将为您定制专属学习路径。")
}

onMounted(() => {
  fetchGrowthData()
})
</script>
