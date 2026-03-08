<template>
  <div class="space-y-4 rounded-[28px] bg-[#f3f4f7] p-4 sm:p-5">
    <header class="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <h1 class="text-4xl sm:text-[38px] leading-none font-black tracking-tight text-zinc-900">面试复盘报告</h1>
        <p class="mt-2 text-zinc-500 text-sm sm:text-base leading-none">
          {{ report.position || '前端开发工程师' }} | {{ reportDate }}
        </p>
      </div>
      <div class="flex flex-wrap items-center gap-2">
        <button
          @click="goReplay"
          class="px-4 py-2.5 bg-zinc-100 text-zinc-600 rounded-2xl text-sm font-semibold hover:bg-zinc-200 transition-colors inline-flex items-center gap-2"
        >
          <span class="inline-flex h-4 w-4 items-center justify-center rounded-full border border-zinc-400 text-[10px]">▶</span>
          查看回放
        </button>
        <button
          @click="handleDownloadReport"
          class="px-4 py-2 bg-indigo-600 text-white rounded-xl text-sm font-semibold hover:bg-indigo-700 transition-colors shadow-md shadow-indigo-200"
        >
          下载 PDF
        </button>
      </div>
    </header>

    <section class="grid grid-cols-1 xl:grid-cols-[minmax(0,1fr)_360px] gap-4">
      <div class="space-y-4">
        <article class="bg-white rounded-[26px] border border-zinc-200 shadow-sm p-4 sm:p-5">
          <h2 class="text-2xl sm:text-[28px] leading-none font-black tracking-tight text-zinc-900 mb-3">综合能力评估</h2>
          <div class="grid grid-cols-1 lg:grid-cols-[360px_minmax(0,1fr)] gap-4 items-center">
            <div class="h-72 rounded-2xl border border-zinc-100 p-1.5">
              <RadarChart :data="abilityScores" />
            </div>
            <div class="space-y-3">
              <div class="rounded-2xl bg-[#eef0ff] border border-[#e5e8ff] p-4 sm:p-5">
                <p class="text-indigo-500 text-base sm:text-lg font-bold">综合评分</p>
                <p class="mt-1 text-indigo-700 font-black text-5xl sm:text-6xl leading-none">
                  {{ displayScore }}
                  <span class="text-xl sm:text-2xl font-bold">/ 100</span>
                </p>
              </div>
              <p class="text-zinc-600 text-sm sm:text-base leading-7">
                {{ reportSummaryText }}
              </p>
            </div>
          </div>
        </article>

        <article class="bg-white rounded-[26px] border border-zinc-200 shadow-sm p-4 sm:p-5">
          <h2 class="text-2xl sm:text-[28px] leading-none font-black tracking-tight text-zinc-900 mb-3">成长曲线</h2>
          <div class="h-75 rounded-2xl border border-zinc-100 bg-white p-2">
            <GrowthCurve :data="historyData" />
          </div>
        </article>
      </div>

      <aside class="space-y-4">
        <article class="bg-white rounded-[26px] border border-zinc-200 shadow-sm p-5">
          <h3 class="text-xl sm:text-2xl leading-none font-black tracking-tight text-zinc-900 mb-5 inline-flex items-center gap-2">
            <span class="h-6 w-6 rounded-full border-2 border-rose-300 inline-flex items-center justify-center">
              <span class="h-2 w-2 bg-rose-500 rounded-full"></span>
            </span>
            技能缺口分析
          </h3>
          <div class="space-y-3">
            <div v-for="item in skillGapList" :key="item.label">
              <div class="flex items-center justify-between text-sm mb-1.5">
                <span class="text-zinc-600">{{ item.label }}</span>
                <span class="text-zinc-400 font-semibold">{{ item.gap }}%</span>
              </div>
              <div class="h-2.5 bg-zinc-100 rounded-full overflow-hidden">
                <div class="h-full bg-indigo-500 rounded-full" :style="{ width: item.gap + '%' }"></div>
              </div>
            </div>
          </div>
          <button
            @click="openLearningMap"
            class="mt-5 w-full border border-indigo-200 text-indigo-600 rounded-2xl py-2.5 text-sm font-bold hover:bg-indigo-50 transition-colors"
          >
            生成学习地图
          </button>
        </article>

        <article class="bg-white rounded-[26px] border border-zinc-200 shadow-sm p-5">
          <h3 class="text-xl sm:text-2xl leading-none font-black tracking-tight text-zinc-900 mb-4 inline-flex items-center gap-2">
            <span class="inline-flex h-6 w-6 items-center justify-center rounded-md border-2 border-indigo-300 text-indigo-500 text-sm">▣</span>
            推荐学习资源
          </h3>
          <div class="space-y-2.5">
            <a
              v-for="(item, idx) in learningResources"
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

    <section class="grid grid-cols-1 lg:grid-cols-3 gap-4">
      <article class="bg-white rounded-2xl p-4 border border-zinc-200 shadow-sm">
        <h3 class="text-lg font-bold text-zinc-900 mb-3">优势分析</h3>
        <div class="space-y-2">
          <div
            v-for="(strength, idx) in (report.strengths || []).slice(0, 3)"
            :key="idx"
            class="p-2.5 bg-emerald-50 rounded-lg text-sm text-zinc-700"
          >
            {{ strength }}
          </div>
          <div v-if="!report.strengths?.length" class="text-sm text-zinc-400">暂无数据</div>
        </div>
      </article>

      <article class="bg-white rounded-2xl p-4 border border-zinc-200 shadow-sm">
        <h3 class="text-lg font-bold text-zinc-900 mb-3">待改进项</h3>
        <div class="space-y-2">
          <div
            v-for="(weakness, idx) in (report.weaknesses || []).slice(0, 3)"
            :key="idx"
            class="p-2.5 bg-amber-50 rounded-lg text-sm text-zinc-700"
          >
            {{ weakness }}
          </div>
          <div v-if="!report.weaknesses?.length" class="text-sm text-zinc-400">暂无数据</div>
        </div>
      </article>

      <article class="bg-white rounded-2xl p-4 border border-zinc-200 shadow-sm">
        <div class="flex items-start justify-between gap-3 mb-3">
          <h3 class="text-lg font-bold text-zinc-900">优化建议</h3>
          <button
            @click="showFeedbackModal = true"
            class="px-2.5 py-1.5 border border-zinc-200 text-zinc-500 rounded-lg text-xs font-semibold hover:bg-zinc-50 transition-colors"
          >
            提交反馈
          </button>
        </div>
        <div class="space-y-2">
          <div
            v-for="(suggestion, idx) in (report.suggestions || []).slice(0, 3)"
            :key="idx"
            class="p-2.5 bg-indigo-50 rounded-lg text-sm text-zinc-700"
          >
            {{ suggestion }}
          </div>
          <div v-if="!report.suggestions?.length" class="text-sm text-zinc-400">暂无数据</div>
        </div>
      </article>
    </section>

    <section class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <article class="bg-white rounded-2xl p-5 border border-zinc-200 shadow-sm">
        <h2 class="text-xl font-bold text-zinc-900 mb-2">答案对比优化</h2>
        <p class="text-sm text-zinc-500 mb-4">对比您的回答与标准答案之间的表达差异</p>
        <div class="space-y-3">
          <div class="p-3 rounded-xl border border-zinc-100 bg-zinc-50/70">
            <div class="text-xs font-bold text-zinc-400 uppercase mb-1.5">您的回答</div>
            <p class="text-sm text-zinc-600">{{ report.overall_analysis || '加载中...' }}</p>
          </div>
          <div class="p-3 rounded-xl border border-emerald-100 bg-emerald-50/70">
            <div class="text-xs font-bold text-emerald-600 uppercase mb-1.5">标准优秀回答</div>
            <p class="text-sm text-zinc-600">系统将根据岗位能力图谱生成标准回答范例，包括文字、语音和视频版本。</p>
          </div>
        </div>
      </article>

      <article id="report-replay" class="bg-white rounded-2xl p-5 border border-zinc-200 shadow-sm">
        <h2 class="text-xl font-bold text-zinc-900 mb-2">面试回放</h2>
        <p class="text-sm text-zinc-500 mb-4">查看完整面试过程回放，包含 AI 实时评估标注</p>
        <video
          v-if="report.replay_url"
          :src="report.replay_url"
          controls
          class="aspect-video w-full bg-zinc-900 rounded-xl object-contain"
        ></video>
        <div v-else class="aspect-video bg-zinc-900 rounded-xl flex items-center justify-center text-zinc-500">
          <div class="text-center">
            <div class="text-3xl mb-1">▶</div>
            <p class="text-sm">暂无可回放视频</p>
          </div>
        </div>
      </article>
    </section>

    <div v-if="showFeedbackModal" class="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50" @click.self="showFeedbackModal = false">
      <div class="bg-white rounded-3xl p-8 w-full max-w-lg shadow-2xl">
        <h2 class="text-xl font-bold text-zinc-900 mb-4">收集意见优化算法</h2>
        <p class="text-sm text-zinc-500 mb-6">如果您已经参加了真实面试，请分享您的反馈帮助我们持续优化模型</p>
        <div class="space-y-4">
          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase mb-1">AI评估准确度 (1-10)</label>
            <input type="number" min="1" max="10" v-model.number="feedbackForm.accuracy" class="w-full px-4 py-3 bg-zinc-50 border border-zinc-200 rounded-xl text-sm" />
          </div>
          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase mb-1">真实面试结果</label>
            <select v-model="feedbackForm.realResult" class="w-full px-4 py-3 bg-zinc-50 border border-zinc-200 rounded-xl text-sm">
              <option value="">请选择</option>
              <option value="passed">通过</option>
              <option value="failed">未通过</option>
              <option value="pending">等待结果</option>
            </select>
          </div>
          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase mb-1">详细反馈</label>
            <textarea v-model="feedbackForm.comments" class="w-full px-4 py-3 bg-zinc-50 border border-zinc-200 rounded-xl text-sm h-24 resize-none" placeholder="请描述AI评估与真实面试的差异..."></textarea>
          </div>
          <div class="flex items-center gap-3 justify-end">
            <button @click="showFeedbackModal = false" class="px-4 py-2 text-zinc-500 hover:bg-zinc-100 rounded-lg transition-colors">取消</button>
            <button @click="submitFeedback" class="px-6 py-2 bg-indigo-600 text-white rounded-xl font-medium hover:bg-indigo-700 transition-colors">提交反馈</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, reactive } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getReport, getReports, generateReport, downloadReport } from '../api/report'
import RadarChart from '../components/RadarChart.vue'
import GrowthCurve from '../components/GrowthCurve.vue'

const route = useRoute()
const router = useRouter()
const report = ref({})
const historyData = ref([])
const showFeedbackModal = ref(false)

const feedbackForm = reactive({
  accuracy: 7,
  realResult: '',
  comments: ''
})

const displayScore = computed(() => {
  const v = Number(report.value.average_score)
  return Number.isFinite(v) ? Math.round(v) : '--'
})

const scoreLevel = computed(() => {
  const s = Number(report.value.average_score)
  if (s >= 90) return '优秀'
  if (s >= 80) return '良好'
  if (s >= 60) return '及格'
  return '需提升'
})

const dimensions = computed(() => {
  const safe = (v) => { const n = Number(v); return Number.isFinite(n) ? n : 0 }
  return [
    { label: '技术深度', value: safe(report.value.technical_score), color: 'text-indigo-600', barColor: 'bg-indigo-500' },
    { label: '表达沟通', value: safe(report.value.expression_score), color: 'text-emerald-600', barColor: 'bg-emerald-500' },
    { label: '逻辑思维', value: safe(report.value.logic_score), color: 'text-amber-600', barColor: 'bg-amber-500' },
    { label: '岗位匹配', value: safe(report.value.matching_score), color: 'text-rose-600', barColor: 'bg-rose-500' },
  ]
})

const reportDate = computed(() => {
  const raw = report.value.created_at || report.value.end_time
  if (!raw) return '暂无日期'
  return raw.toString().slice(0, 10)
})

const reportSummaryText = computed(() => {
  if (report.value.overall_analysis && String(report.value.overall_analysis).trim()) {
    const text = String(report.value.overall_analysis).replace(/\s+/g, ' ').trim()
    if (text.length > 56) {
      return `${text.slice(0, 56)}...`
    }
    return text
  }
  return '您的技术底子非常扎实，尤其在 React 原理和工程化方面表现突出。建议在表达时更多结合实际业务场景，增强说服力。'
})

const skillGapList = computed(() => {
  const safe = (v) => {
    const n = Number(v)
    if (!Number.isFinite(n)) return 50
    return Math.max(20, Math.min(90, Math.round(100 - n)))
  }

  return [
    { label: '系统架构设计', gap: safe(report.value.technical_score) },
    { label: '性能优化实战', gap: safe(report.value.logic_score) },
    { label: '跨端开发经验', gap: safe(report.value.matching_score) }
  ]
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
  if (text.includes('真题') || text.includes('面试题') || text.includes('题库')) {
    return `https://leetcode.cn/problemset/all/?search=${encodeURIComponent(text.replace(/\s+/g, ''))}`
  }
  if (text.includes('复盘') || text.includes('模板')) {
    return `https://www.bing.com/search?q=${encodeURIComponent(`"${text}" 模板`)}`
  }

  return `https://www.bing.com/search?q=${encodeURIComponent(`"${text}" 教程`)}`
}

const learningResources = computed(() => {
  const titles = [
    '《前端架构设计》- 深度解析',
    'React 18 核心原理解析视频',
    '大厂面试真题库 - 2024版'
  ]

  return titles.map((title) => ({
    title,
    url: buildResourceLink(title)
  }))
})

const abilityScores = computed(() => {
  const safe = (value) => Number.isFinite(Number(value)) ? Number(value) : 0
  return {
    '逻辑思维': safe(report.value.logic_score),
    '语言表达': safe(report.value.expression_score),
    '专业知识': safe(report.value.technical_score),
    '抗压能力': safe(report.value.behavior_score),
    '岗位匹配': safe(report.value.matching_score)
  }
})

const submitFeedback = () => {
  // TODO: submit feedback to backend for algorithm optimization
  showFeedbackModal.value = false
  alert('感谢您的反馈！您的意见将帮助我们持续优化AI面试评估算法。')
}

const goReplay = () => {
  const replaySection = document.getElementById('report-replay')
  if (replaySection) {
    replaySection.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }
}

const openLearningMap = () => {
  router.push({ path: '/student/growth', query: { openMap: '1' } })
}

const handleDownloadReport = async () => {
  const id = report.value?.id || route.params.id
  if (!id) {
    alert('报告尚未生成，暂时无法下载')
    return
  }
  try {
    const blob = await downloadReport(id)
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `report_${id}.md`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
  } catch (error) {
    console.error('下载报告失败', error)
    alert('下载报告失败，请稍后重试')
  }
}

const buildHistoryData = (reports = []) => {
  const items = reports
    .map((item) => {
      const rawDate = item.created_at || item.end_time || ''
      const date = rawDate ? rawDate.toString().slice(0, 10) : ''
      const score = Number(item.average_score)
      return {
        date,
        score: Number.isFinite(score) ? score : 0,
        time: rawDate ? new Date(rawDate).getTime() : 0
      }
    })
    .filter((item) => item.date)
    .sort((a, b) => a.time - b.time)
    .slice(-12)
    .map(({ date, score }) => ({ date, score }))
  return items
}

onMounted(async () => {
  const id = route.params.id
  if (id) {
    try {
      let res
      try {
        res = await getReport(id)
      } catch (_) {
        const generated = await generateReport({ interview_id: Number(id) })
        if (generated?.report?.id) {
          res = await getReport(generated.report.id)
        }
      }
      report.value = res?.report || {}
      const listRes = await getReports({ page: 1, page_size: 50 })
      const reportList = listRes?.reports || []
      const trend = buildHistoryData(reportList)
      if (trend.length > 0) {
        historyData.value = trend
      } else {
        const reportDate = (report.value.created_at || report.value.end_time || '').toString().slice(0, 10) || '当前'
        historyData.value = [{ date: reportDate, score: Number(report.value.average_score) || 0 }]
      }
    } catch (error) {
      console.error('获取报告失败', error)
    }
  }
})
</script>
