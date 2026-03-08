<template>
  <div class="space-y-4 rounded-3xl bg-[#f3f4f7] p-4 sm:p-5">
    <header class="flex items-end justify-between">
      <div>
        <h1 class="text-3xl font-black tracking-tight text-zinc-900">成长中心</h1>
        <p class="text-zinc-500 mt-1.5 text-sm">查看您的能力评估与成长曲线，获取个性化学习建议</p>
      </div>
    </header>

    <div v-if="loading" class="text-center py-16 text-zinc-500">加载数据中...</div>

    <div v-else class="space-y-4">
      <section class="grid grid-cols-1 xl:grid-cols-[minmax(0,1fr)_340px] gap-4">
        <div class="space-y-4">
          <article class="bg-white rounded-3xl p-4 sm:p-5 border border-zinc-200 shadow-sm">
            <h2 class="text-2xl font-black tracking-tight text-zinc-900 mb-3">综合能力评估</h2>
            <div class="grid grid-cols-1 lg:grid-cols-[320px_minmax(0,1fr)] gap-4 items-center">
              <div class="h-64 rounded-2xl border border-zinc-100 p-2 relative">
                <Radar v-if="radarData" :data="radarData" :options="radarOptions" />
                <div v-else class="absolute inset-0 flex items-center justify-center text-zinc-400">暂无足够数据</div>
              </div>
              <div class="space-y-3">
                <div class="rounded-2xl bg-[#eef0ff] border border-[#e5e8ff] p-4">
                  <p class="text-indigo-500 text-base font-bold">综合评分</p>
                  <p class="mt-1 text-indigo-700 text-5xl font-black leading-none">
                    {{ displayScore }}
                    <span class="text-2xl font-bold">/ 100</span>
                  </p>
                </div>
                <p class="text-sm text-zinc-600 leading-7">
                  {{ growthSummary }}
                </p>
              </div>
            </div>
          </article>

          <article class="bg-white rounded-3xl p-4 sm:p-5 border border-zinc-200 shadow-sm">
            <h2 class="text-2xl font-black tracking-tight text-zinc-900 mb-3">成长曲线</h2>
            <div class="h-65 rounded-2xl border border-zinc-100 bg-white p-2 relative">
              <Line v-if="lineData" :data="lineData" :options="lineOptions" />
              <div v-else class="absolute inset-0 flex items-center justify-center text-zinc-400">完成更多面试以查看趋势</div>
            </div>
          </article>
        </div>

        <aside class="space-y-4">
          <article class="bg-white rounded-3xl border border-zinc-200 shadow-sm p-5">
            <h3 class="text-xl font-black tracking-tight text-zinc-900 mb-4 inline-flex items-center gap-2">
              <span class="h-5 w-5 rounded-full border-2 border-rose-300 inline-flex items-center justify-center">
                <span class="h-1.5 w-1.5 bg-rose-500 rounded-full"></span>
              </span>
              技能缺口分析
            </h3>
            <div class="space-y-3">
              <div v-for="item in normalizedSkillGaps" :key="item.name">
                <div class="flex items-center justify-between text-sm mb-1.5">
                  <span class="text-zinc-600">{{ item.name }}</span>
                  <span class="text-zinc-400 font-semibold">{{ item.percent }}%</span>
                </div>
                <div class="h-2 bg-zinc-100 rounded-full overflow-hidden">
                  <div class="h-full bg-indigo-500 rounded-full" :style="{ width: item.percent + '%' }"></div>
                </div>
              </div>
            </div>
            <button
              @click="showLearningMap = true"
              class="mt-5 w-full border border-indigo-200 text-indigo-600 rounded-2xl py-2.5 text-sm font-bold hover:bg-indigo-50 transition-colors"
            >
              生成学习地图
            </button>
          </article>

          <article class="bg-white rounded-3xl border border-zinc-200 shadow-sm p-5">
            <h3 class="text-xl font-black tracking-tight text-zinc-900 mb-4 inline-flex items-center gap-2">
              <span class="inline-flex h-5 w-5 items-center justify-center rounded-md border-2 border-indigo-300 text-indigo-500 text-xs">▣</span>
              推荐学习资源
            </h3>
            <div class="space-y-2.5">
                <a
                  v-for="(item, idx) in recommendedResources"
                  :key="idx"
                  :href="item.url"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="block rounded-xl bg-zinc-100/80 px-3.5 py-3 text-zinc-700 text-sm font-semibold hover:bg-zinc-200/90 transition-colors"
                >
                  {{ item.title }}
                </a>
            </div>
          </article>
        </aside>
      </section>

      <section class="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <article
          class="bg-zinc-900 text-white rounded-3xl p-5 relative overflow-hidden group cursor-pointer"
          @click="showLearningMap = true"
        >
          <div class="relative z-10">
            <div class="text-xs font-bold text-zinc-400 uppercase tracking-wider mb-2">个性化学习地图</div>
            <h3 class="text-xl font-bold mb-6 leading-tight">基于您的短板生成学习计划</h3>
            <button class="text-sm font-bold text-indigo-400 flex items-center gap-1 group-hover:gap-2 transition-all">
              开始学习 <ChevronRight class="h-4 w-4" />
            </button>
          </div>
          <BookOpen class="absolute -bottom-4 -right-4 h-28 w-28 text-white/5 rotate-12 transition-all duration-500" />
        </article>

        <article class="bg-white rounded-3xl p-5 border border-zinc-200 shadow-sm">
          <h3 class="text-lg font-bold text-zinc-900 mb-4 flex items-center gap-2">
            <FileText class="h-5 w-5 text-violet-600" />
            简历优化建议
          </h3>
          <p class="text-sm text-zinc-500 mb-4">基于面试表现与岗位能力图谱，为您生成简历优化建议</p>
          <div class="space-y-3">
            <div v-for="(tip, i) in resumeTips" :key="i" class="flex gap-3 p-3 rounded-xl bg-violet-50/50 border border-violet-100/50">
              <div class="w-6 h-6 rounded-full bg-violet-100 text-violet-600 text-xs flex items-center justify-center font-bold shrink-0 mt-0.5">{{ i + 1 }}</div>
              <p class="text-sm text-zinc-700">{{ tip }}</p>
            </div>
          </div>
          <button @click="generateResumeTips" class="mt-4 w-full py-2.5 bg-violet-600 text-white rounded-xl text-sm font-medium hover:bg-violet-700 transition-colors">
            {{ resumeTipsLoading ? '生成中...' : '重新生成建议' }}
          </button>
        </article>
      </section>
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
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { getGrowthStats } from '../api/growth'
import { 
  ChevronRight, BookOpen, FileText 
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

const loading = ref(true)
const route = useRoute()
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
      angleLines: { color: '#d9dce3' },
      grid: { color: '#d9dce3' },
      pointLabels: { color: '#8b8fa3', font: { size: 12, weight: 700 } },
      ticks: { display: false, stepSize: 25 },
      suggestedMin: 0,
      suggestedMax: 100
    }
  },
  plugins: { legend: { display: false } }
}

const lineOptions = computed(() => {
  const scores = (lineData.value?.datasets?.[0]?.data || [])
    .map((item) => Number(item))
    .filter((v) => Number.isFinite(v))

  const minScore = scores.length ? Math.min(...scores) : 60
  const maxScore = scores.length ? Math.max(...scores) : 80

  let min = Math.max(0, Math.floor((minScore - 10) / 5) * 5)
  let max = Math.min(100, Math.ceil((maxScore + 10) / 5) * 5)

  if (max - min < 20) {
    min = Math.max(0, min - 10)
    max = Math.min(100, max + 10)
  }

  return {
    responsive: true,
    maintainAspectRatio: false,
    scales: {
      x: { grid: { display: false }, ticks: { display: false } },
      y: {
        min,
        max,
        ticks: { color: '#8f93a5', stepSize: 10 },
        grid: { color: '#d9dce3', borderDash: [4, 4] }
      }
    },
    plugins: { legend: { display: false } }
  }
})

const displayScore = computed(() => {
  const scoreList = lineData.value?.datasets?.[0]?.data || []
  const lastScore = Number(scoreList[scoreList.length - 1])
  if (Number.isFinite(lastScore)) return Math.round(lastScore)

  const radarScore = radarData.value?.datasets?.[0]?.data || []
  if (!radarScore.length) return '--'
  const avg = radarScore.reduce((sum, v) => sum + Number(v || 0), 0) / radarScore.length
  return Math.round(avg)
})

const growthSummary = computed(() => {
  const score = Number(displayScore.value)
  if (!Number.isFinite(score)) {
    return '当前数据量较少，建议继续完成面试练习，系统会自动生成更准确的能力画像。'
  }
  if (score >= 85) {
    return '整体能力表现稳定，技术基础扎实。建议继续提升复杂场景的表达深度与系统化拆解能力。'
  }
  if (score >= 70) {
    return '能力处于持续上升阶段。建议围绕高频考点进行专项训练，强化关键技术点输出。'
  }
  return '当前仍有较大提升空间。建议先从核心短板切入，逐步建立稳定的作答结构。'
})

const normalizedSkillGaps = computed(() => {
  const fallbackFromRadar = () => {
    const labels = radarData.value?.labels || []
    const values = radarData.value?.datasets?.[0]?.data || []
    return labels
      .map((name, idx) => {
        const score = Number(values[idx])
        return {
          name,
          percent: Number.isFinite(score) ? Math.max(15, Math.min(90, Math.round(100 - score))) : 45
        }
      })
      .sort((a, b) => b.percent - a.percent)
      .slice(0, 3)
  }

  if (!skillGaps.value || skillGaps.value.length === 0) {
    return fallbackFromRadar()
  }

  const levelMap = {
    急需提升: 72,
    中等差距: 52,
    良好: 30
  }

  return skillGaps.value.slice(0, 3).map((item) => {
    const raw = Number(item.gap_percent)
    const level = (item.level || '').trim()
    const percent = Number.isFinite(raw)
      ? Math.max(15, Math.min(90, Math.round(raw)))
      : (levelMap[level] || 45)
    return {
      name: item.name || '待提升项',
      percent
    }
  })
})

const buildResourceLink = (title) => {
  const text = String(title || '')
  const lower = text.toLowerCase()

  const exactMap = {
    'React 18 核心原理解析视频': 'https://search.bilibili.com/all?keyword=React%2018%20%E6%A0%B8%E5%BF%83%E5%8E%9F%E7%90%86%20%E8%A7%A3%E6%9E%90',
    '大厂面试真题库 - 2024版': 'https://leetcode.cn/studyplan/top-interview-150/',
    '项目复盘与亮点提炼模板': 'https://www.notion.so/templates/search?query=retrospective',
    '《前端架构设计》- 深度解析': 'https://juejin.cn/search?query=%E5%89%8D%E7%AB%AF%E6%9E%B6%E6%9E%84%E8%AE%BE%E8%AE%A1',
    '高并发系统设计训练营': 'https://juejin.cn/search?query=%E9%AB%98%E5%B9%B6%E5%8F%91%20%E7%B3%BB%E7%BB%9F%E8%AE%BE%E8%AE%A1',
    '工程化性能优化实战': 'https://web.dev/learn/performance',
    'STAR 高分表达实战': 'https://www.bing.com/search?q=STAR+%E9%9D%A2%E8%AF%95%E8%A1%A8%E8%BE%BE+%E5%AE%9E%E6%88%98',
    '技术面试表达结构模板': 'https://www.bing.com/search?q=%E6%8A%80%E6%9C%AF%E9%9D%A2%E8%AF%95+STAR+%E7%BB%93%E6%9E%84%E5%8C%96%E8%A1%A8%E8%BE%BE',
    '大厂行为面试题库': 'https://www.nowcoder.com/interview/center'
  }

  if (exactMap[text]) {
    return exactMap[text]
  }

  if (lower.includes('react 18')) {
    return 'https://search.bilibili.com/all?keyword=React%2018%20%E5%8E%9F%E7%90%86'
  }
  if (text.includes('前端架构') || text.includes('系统设计')) {
    return `https://juejin.cn/search?query=${encodeURIComponent(text)}`
  }
  if (text.includes('视频')) {
    return `https://search.bilibili.com/all?keyword=${encodeURIComponent(text)}`
  }
  if (text.includes('面试') || text.includes('题库') || text.includes('真题')) {
    return `https://leetcode.cn/problemset/all/?search=${encodeURIComponent(text.replace(/\s+/g, ''))}`
  }
  if (text.includes('复盘') || text.includes('模板')) {
    return `https://www.bing.com/search?q=${encodeURIComponent(`"${text}" 模板`)}`
  }

  return `https://www.bing.com/search?q=${encodeURIComponent(`"${text}" 教程`)}`
}

const recommendedResources = computed(() => {
  const first = normalizedSkillGaps.value[0]?.name || ''
  if (first.includes('表达') || first.includes('沟通')) {
    return [
      { title: 'STAR 高分表达实战', url: buildResourceLink('STAR 高分表达实战') },
      { title: '技术面试表达结构模板', url: buildResourceLink('技术面试表达结构模板') },
      { title: '大厂行为面试题库', url: buildResourceLink('大厂行为面试题库') }
    ]
  }
  if (first.includes('系统') || first.includes('架构')) {
    return [
      { title: '《前端架构设计》- 深度解析', url: buildResourceLink('《前端架构设计》- 深度解析') },
      { title: '高并发系统设计训练营', url: buildResourceLink('高并发系统设计训练营') },
      { title: '工程化性能优化实战', url: buildResourceLink('工程化性能优化实战') }
    ]
  }
  return [
    { title: 'React 18 核心原理解析视频', url: buildResourceLink('React 18 核心原理解析视频') },
    { title: '大厂面试真题库 - 2024版', url: buildResourceLink('大厂面试真题库 - 2024版') },
    { title: '项目复盘与亮点提炼模板', url: buildResourceLink('项目复盘与亮点提炼模板') }
  ]
})

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
          backgroundColor: 'rgba(98, 96, 244, 0.5)',
          borderColor: '#5d63f2',
          borderWidth: 2,
          pointRadius: 0,
          pointHoverRadius: 0,
          pointBackgroundColor: '#5d63f2',
          pointBorderColor: '#fff',
          pointHoverBackgroundColor: '#fff',
          pointHoverBorderColor: '#5d63f2',
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
          backgroundColor: 'rgba(93, 99, 242, 0.15)',
          borderColor: '#5d63f2',
          borderWidth: 3,
          tension: 0.28,
          pointRadius: 4,
          pointHoverRadius: 5,
          pointBackgroundColor: '#5d63f2',
          pointBorderColor: '#fff',
          pointHoverBackgroundColor: '#fff',
          pointHoverBorderColor: '#5d63f2',
          fill: false,
          data: stats.growth_data.map(item => item.score),
          spanGaps: true
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
  if (route.query.openMap === '1') {
    showLearningMap.value = true
  }
})
</script>
