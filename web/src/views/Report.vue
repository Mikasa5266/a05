<template>
  <div class="space-y-8">
    <!-- Header -->
    <header class="flex items-end justify-between">
      <div>
        <h1 class="text-3xl font-bold tracking-tight text-zinc-900">面试复盘报告</h1>
        <p class="text-zinc-500 mt-2">{{ report.position }} · {{ report.difficulty }} · 可视化综合评估</p>
      </div>
      <div class="flex items-center gap-3">
        <button @click="showFeedbackModal = true" class="px-4 py-2.5 border border-zinc-200 text-zinc-600 rounded-xl text-sm font-medium hover:bg-zinc-50 transition-colors">
          提交反馈优化算法
        </button>
        <button @click="handleDownloadReport" class="px-4 py-2.5 bg-indigo-600 text-white rounded-xl text-sm font-medium hover:bg-indigo-700 transition-colors shadow-lg shadow-indigo-200">
          下载报告
        </button>
      </div>
    </header>

    <!-- Score Overview -->
    <div class="grid grid-cols-1 md:grid-cols-5 gap-4">
      <div class="md:col-span-1 bg-gradient-to-br from-indigo-600 to-violet-600 rounded-3xl p-6 text-white text-center flex flex-col items-center justify-center">
        <div class="text-5xl font-black mb-2">{{ displayScore }}</div>
        <div class="text-indigo-200 text-sm font-medium">综合得分</div>
        <div class="mt-3 px-3 py-1 bg-white/20 rounded-full text-xs">{{ scoreLevel }}</div>
      </div>
      <div class="md:col-span-4 grid grid-cols-2 md:grid-cols-4 gap-4">
        <div v-for="dim in dimensions" :key="dim.label" class="bg-white rounded-2xl p-5 border border-zinc-100 shadow-sm">
          <div class="text-sm text-zinc-400 mb-2">{{ dim.label }}</div>
          <div class="text-2xl font-bold mb-2" :class="dim.color">{{ dim.value }}</div>
          <div class="h-1.5 bg-zinc-100 rounded-full overflow-hidden">
            <div class="h-full rounded-full transition-all duration-700" :class="dim.barColor" :style="{ width: dim.value + '%' }"></div>
          </div>
        </div>
      </div>
    </div>

    <!-- Charts Row -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Radar Chart -->
      <div class="bg-white rounded-3xl p-8 border border-zinc-100 shadow-sm">
        <h2 class="text-lg font-bold text-zinc-900 mb-6">能力雷达图</h2>
        <div class="h-[300px]">
          <RadarChart :data="abilityScores" />
        </div>
      </div>
      <!-- Growth Curve -->
      <div class="bg-white rounded-3xl p-8 border border-zinc-100 shadow-sm">
        <h2 class="text-lg font-bold text-zinc-900 mb-6">成长曲线</h2>
        <div class="h-[300px]">
          <GrowthCurve :data="historyData" />
        </div>
      </div>
    </div>

    <!-- Detailed Analysis -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Strengths -->
      <div class="bg-white rounded-3xl p-6 border border-zinc-100 shadow-sm">
        <h3 class="text-lg font-bold text-zinc-900 mb-4 flex items-center gap-2">
          <div class="w-2 h-2 rounded-full bg-emerald-500"></div>
          优势分析
        </h3>
        <div class="space-y-3">
          <div v-for="(strength, idx) in (report.strengths || [])" :key="idx" class="p-3 bg-emerald-50/50 rounded-xl text-sm text-zinc-700">
            {{ strength }}
          </div>
          <div v-if="!report.strengths?.length" class="text-sm text-zinc-400">暂无数据</div>
        </div>
      </div>
      <!-- Weaknesses -->
      <div class="bg-white rounded-3xl p-6 border border-zinc-100 shadow-sm">
        <h3 class="text-lg font-bold text-zinc-900 mb-4 flex items-center gap-2">
          <div class="w-2 h-2 rounded-full bg-amber-500"></div>
          待改进项
        </h3>
        <div class="space-y-3">
          <div v-for="(weakness, idx) in (report.weaknesses || [])" :key="idx" class="p-3 bg-amber-50/50 rounded-xl text-sm text-zinc-700">
            {{ weakness }}
          </div>
          <div v-if="!report.weaknesses?.length" class="text-sm text-zinc-400">暂无数据</div>
        </div>
      </div>
      <!-- Suggestions -->
      <div class="bg-white rounded-3xl p-6 border border-zinc-100 shadow-sm">
        <h3 class="text-lg font-bold text-zinc-900 mb-4 flex items-center gap-2">
          <div class="w-2 h-2 rounded-full bg-indigo-500"></div>
          优化建议
        </h3>
        <div class="space-y-3">
          <div v-for="(suggestion, idx) in (report.suggestions || [])" :key="idx" class="p-3 bg-indigo-50/50 rounded-xl text-sm text-zinc-700">
            {{ suggestion }}
          </div>
          <div v-if="!report.suggestions?.length" class="text-sm text-zinc-400">暂无数据</div>
        </div>
      </div>
    </div>

    <!-- Answer Comparison & Replay -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Answer Comparison -->
      <div class="bg-white rounded-3xl p-8 border border-zinc-100 shadow-sm">
        <h2 class="text-lg font-bold text-zinc-900 mb-4">答案对比优化</h2>
        <p class="text-sm text-zinc-500 mb-6">对比您的回答与优质标准答案的差异</p>
        <div class="space-y-4">
          <div class="p-4 rounded-2xl border border-zinc-100">
            <div class="text-xs font-bold text-zinc-400 uppercase mb-2">您的回答</div>
            <p class="text-sm text-zinc-600">{{ report.overall_analysis || '加载中...' }}</p>
          </div>
          <div class="p-4 rounded-2xl border border-emerald-100 bg-emerald-50/30">
            <div class="text-xs font-bold text-emerald-600 uppercase mb-2">标准优秀回答</div>
            <p class="text-sm text-zinc-600">系统将根据岗位能力图谱生成标准回答范例，包括文字、语音和视频版本。</p>
          </div>
        </div>
      </div>

      <!-- Interview Replay -->
      <div class="bg-white rounded-3xl p-8 border border-zinc-100 shadow-sm">
        <h2 class="text-lg font-bold text-zinc-900 mb-4">面试回放</h2>
        <p class="text-sm text-zinc-500 mb-6">查看完整面试过程回放，包含 AI 实时评估标注</p>
        <video v-if="report.replay_url" :src="report.replay_url" controls class="aspect-video w-full bg-zinc-900 rounded-2xl object-contain"></video>
        <div v-else class="aspect-video bg-zinc-900 rounded-2xl flex items-center justify-center text-zinc-500">
          <div class="text-center">
            <div class="text-4xl mb-2">▶</div>
            <p class="text-sm">暂无可回放视频</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Demo Library Banner -->
    <div class="bg-zinc-900 text-white rounded-3xl p-8 flex items-center justify-between">
      <div>
        <h2 class="text-xl font-bold mb-2">示范库</h2>
        <p class="text-zinc-400 text-sm">浏览 AI 面试视频示范和优秀面试语音范例，对照学习提升</p>
      </div>
      <button class="px-6 py-3 bg-white text-zinc-900 rounded-xl font-bold text-sm hover:bg-zinc-100 transition-colors">
        进入示范库
      </button>
    </div>

    <!-- Feedback Modal -->
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
import { useRoute } from 'vue-router'
import { getReport, getReports, generateReport, downloadReport } from '../api/report'
import RadarChart from '../components/RadarChart.vue'
import GrowthCurve from '../components/GrowthCurve.vue'

const route = useRoute()
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

const abilityScores = computed(() => {
  const safe = (value) => Number.isFinite(Number(value)) ? Number(value) : 0
  return {
    '技术能力': safe(report.value.technical_score),
    '表达沟通': safe(report.value.expression_score),
    '逻辑思维': safe(report.value.logic_score),
    '岗位匹配': safe(report.value.matching_score),
    '职业素养': safe(report.value.behavior_score)
  }
})

const submitFeedback = () => {
  // TODO: submit feedback to backend for algorithm optimization
  showFeedbackModal.value = false
  alert('感谢您的反馈！您的意见将帮助我们持续优化AI面试评估算法。')
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
