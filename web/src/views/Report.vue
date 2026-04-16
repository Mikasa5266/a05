<template>
  <div class="report-page">
    <div v-if="loading" class="report-shell">
      <section class="glass-card state-card">
        <h1 class="hero-title">复盘报告加载中</h1>
        <p class="hero-subtitle">正在读取报告数据，请稍候...</p>
      </section>
    </div>

    <div v-else-if="reportUnavailable" class="report-shell">
      <section class="glass-card state-card">
        <h1 class="hero-title">暂无可查看的复盘报告</h1>
        <p class="hero-subtitle">{{ loadError || '你还没有可用的面试记录，先完成一次面试后再来查看复盘。' }}</p>
        <div class="hero-actions">
          <button class="btn btn-secondary" @click="router.push('/student/history')">查看历史记录</button>
          <button class="btn btn-primary" @click="router.push('/interview/mode-select')">去开始面试</button>
        </div>
      </section>
    </div>

    <div v-else class="report-shell">
      <section class="glass-card hero-card">
        <div class="hero-head">
          <div>
            <h1 class="hero-title">面试反馈报告</h1>
            <p class="hero-subtitle">{{ report.position || '岗位未设置' }} · {{ difficultyText }} · RAG 评估结果可视化</p>
          </div>
          <div class="hero-actions">
            <button class="btn btn-secondary" @click="showFeedbackModal = true">提交反馈</button>
            <button class="btn btn-primary" @click="handleDownloadReport">下载报告</button>
          </div>
        </div>

        <div class="hero-body">
          <div class="hero-gauge">
            <GlassEchart :option="scoreGaugeOption" height="240px" />
            <div class="score-tag">{{ scoreLevel }}</div>
          </div>

          <div class="hero-dimensions">
            <div v-for="dimension in dimensions" :key="dimension.key" class="dimension-block">
              <p class="dimension-label">{{ dimension.label }}</p>
              <p class="dimension-value">{{ dimension.value }}</p>
              <div class="dimension-track">
                <div class="dimension-fill" :style="{ width: `${dimension.value}%` }"></div>
              </div>
            </div>
          </div>
        </div>

        <div class="hero-chart">
          <GlassEchart :option="dimensionBarOption" height="240px" />
        </div>
      </section>

      <section class="content-grid">
        <article class="glass-card process-card">
          <div class="panel-head">
            <h2>面试过程记录</h2>
            <span>{{ qaDetails.length }} 个题目节点</span>
          </div>

          <div v-if="isGroupScenario" class="group-replay-layout">
            <section class="group-replay-main">
              <p class="replay-hint">群面报告默认展示多人回放主轨，便于复盘成员互动。</p>
              <div class="replay-stage">
                <video v-if="groupPrimaryPlaybackUrl" :src="groupPrimaryPlaybackUrl" controls class="replay-video"></video>
                <div v-else class="replay-placeholder">
                  <p>暂无多人回放视频</p>
                  <small>可稍后重试或上传群面录屏文件。</small>
                </div>
              </div>
            </section>

            <section class="group-replay-side">
              <div class="group-replay-side-card">
                <h3>单人视角回放（可选）</h3>
                <div class="replay-stage compact">
                  <video v-if="groupSecondaryPlaybackUrl" :src="groupSecondaryPlaybackUrl" controls class="replay-video"></video>
                  <div v-else class="replay-placeholder compact">
                    <p>暂无单人视角视频</p>
                    <small>当前报告仅提供群面主轨。</small>
                  </div>
                </div>
              </div>
            </section>
          </div>

          <div v-else>
            <p class="replay-hint">单人面试仅展示单人回放轨道。</p>
            <div class="replay-stage">
              <video v-if="singlePlaybackUrl" :src="singlePlaybackUrl" controls class="replay-video"></video>
              <div v-else class="replay-placeholder">
                <p>暂无单人回放视频</p>
                <small>可稍后重试或上传单人录屏文件。</small>
              </div>
            </div>
          </div>

          <div class="timeline-wrap">
            <div v-if="qaDetails.length === 0" class="empty-tip">暂无题目回顾数据，请稍后刷新。</div>
            <article v-for="(item, index) in qaDetails" :key="`${index}-${item.question}`" class="timeline-item">
              <div class="timeline-index">Q{{ index + 1 }}</div>
              <div class="timeline-content">
                <h4>{{ item.question }}</h4>
                <p class="timeline-answer"><strong>你的回答：</strong>{{ item.user_answer }}</p>
                <p class="timeline-answer"><strong>优化建议：</strong>{{ item.optimized_answer }}</p>
                <div class="timeline-tags">
                  <span v-for="(tag, tagIndex) in item.key_improvements" :key="`${index}-${tagIndex}`">{{ tag }}</span>
                </div>
              </div>
            </article>
          </div>

          <div v-if="isGroupScenario" class="conversation-grid">
            <section>
              <h3>语音转写片段</h3>
              <ul>
                <li v-for="(item, index) in audioTranscripts" :key="`audio-${index}`">
                  {{ item.speaker_id ? `成员${item.speaker_id}` : '成员' }}：{{ item.content }}
                </li>
                <li v-if="audioTranscripts.length === 0" class="faded">暂无语音转写记录</li>
              </ul>
            </section>
            <section>
              <h3>公屏聊天记录</h3>
              <ul>
                <li v-for="(item, index) in chatMessages" :key="`chat-${index}`">
                  {{ item.sender_id ? `成员${item.sender_id}` : '成员' }}：{{ item.content }}
                </li>
                <li v-if="chatMessages.length === 0" class="faded">暂无聊天记录</li>
              </ul>
            </section>
          </div>
        </article>

        <aside class="side-stack">
          <article class="glass-card insight-card">
            <div class="panel-head">
              <h2>AI 深度建议</h2>
              <span>基于本次答题过程</span>
            </div>

            <p class="overall-analysis">{{ overallAnalysis }}</p>

            <div class="insight-section">
              <h3>优势能力</h3>
              <ul>
                <li v-for="(item, idx) in strengths" :key="`s-${idx}`">{{ item }}</li>
                <li v-if="strengths.length === 0" class="faded">暂无优势标签</li>
              </ul>
            </div>

            <div class="insight-section">
              <h3>短板风险</h3>
              <ul>
                <li v-for="(item, idx) in weaknesses" :key="`w-${idx}`">{{ item }}</li>
                <li v-if="weaknesses.length === 0" class="faded">暂无短板标签</li>
              </ul>
            </div>

            <div class="insight-section">
              <h3>可执行建议</h3>
              <ul>
                <li v-for="(item, idx) in suggestions" :key="`g-${idx}`">{{ item }}</li>
                <li v-if="suggestions.length === 0" class="faded">暂无建议项</li>
              </ul>
            </div>
          </article>

          <article class="glass-card">
            <div class="panel-head">
              <h2>能力雷达图</h2>
              <span>实时得分映射</span>
            </div>
            <GlassEchart :option="radarOption" height="300px" />
          </article>

          <article class="glass-card">
            <div class="panel-head">
              <h2>心理波动曲线（预留）</h2>
              <span>占位数据，用于后续接入情绪识别模型</span>
            </div>
            <GlassEchart :option="moodCurveOption" height="250px" />
          </article>

          <article class="glass-card">
            <div class="panel-head">
              <h2>近期待分趋势</h2>
              <span>历史报告趋势</span>
            </div>
            <GlassEchart :option="trendOption" height="250px" />
          </article>
        </aside>
      </section>
    </div>

    <div v-if="showFeedbackModal" class="modal-mask" @click.self="showFeedbackModal = false">
      <div class="glass-card modal-panel">
        <h3>收集意见优化算法</h3>
        <p>如果你已参加真实面试，欢迎补充模型偏差与建议。</p>

        <div class="form-grid">
          <label>
            AI 评估准确度（1-10）
            <input v-model.number="feedbackForm.accuracy" type="number" min="1" max="10" />
          </label>
          <label>
            真实面试结果
            <select v-model="feedbackForm.realResult">
              <option value="">请选择</option>
              <option value="passed">通过</option>
              <option value="failed">未通过</option>
              <option value="pending">等待结果</option>
            </select>
          </label>
          <label class="full-width">
            详细反馈
            <textarea v-model="feedbackForm.comments" placeholder="请描述 AI 评估与真实反馈的偏差点"></textarea>
          </label>
        </div>

        <div class="modal-actions">
          <button class="btn btn-secondary" @click="showFeedbackModal = false">取消</button>
          <button class="btn btn-primary" @click="submitFeedback">提交反馈</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { downloadReport, getReport, getReports } from '../api/report'
import GlassEchart from '../components/report/GlassEchart.vue'

const route = useRoute()
const router = useRouter()

const report = ref({})
const historyData = ref([])
const showFeedbackModal = ref(false)
const loading = ref(true)
const reportUnavailable = ref(false)
const loadError = ref('')

const feedbackForm = reactive({
  accuracy: 7,
  realResult: '',
  comments: ''
})

const chartTextColor = '#334155'
const chartLineColor = '#d8e1ea'

const safeNumber = (value) => {
  const n = Number(value)
  if (!Number.isFinite(n)) return 0
  return Math.max(0, Math.min(100, Math.round(n)))
}

const parseMaybeJSON = (raw, fallback = []) => {
  if (Array.isArray(raw)) return raw
  if (raw == null) return fallback
  if (typeof raw !== 'string') return fallback

  const trimmed = raw.trim()
  if (!trimmed) return fallback
  try {
    const parsed = JSON.parse(trimmed)
    return Array.isArray(parsed) ? parsed : fallback
  } catch (_) {
    return fallback
  }
}

const normalizeStringList = (raw) => {
  if (Array.isArray(raw)) {
    return raw.map((item) => String(item || '').trim()).filter(Boolean)
  }
  if (typeof raw === 'string') {
    const parsed = parseMaybeJSON(raw)
    if (parsed.length > 0) {
      return parsed.map((item) => String(item || '').trim()).filter(Boolean)
    }
    return raw
      .split('|')
      .map((item) => item.trim())
      .filter(Boolean)
  }
  return []
}

const normalizeQADetails = (raw) => {
  const source = Array.isArray(raw) ? raw : parseMaybeJSON(raw)
  return source
    .map((item) => {
      const question = String(item?.question || '').trim()
      const userAnswer = String(item?.user_answer || '').trim()
      const optimizedAnswer = String(item?.optimized_answer || '').trim()
      if (!question) return null

      const keyImprovements = Array.isArray(item?.key_improvements)
        ? item.key_improvements.map((entry) => String(entry || '').trim()).filter(Boolean)
        : []

      return {
        question,
        user_answer: userAnswer || '候选人回答缺失。',
        optimized_answer: optimizedAnswer || '建议补充结构化论证与可执行落地步骤。',
        key_improvements: keyImprovements.length > 0
          ? keyImprovements.slice(0, 4)
          : ['补充场景细节', '说明关键取舍依据']
      }
    })
    .filter(Boolean)
    .slice(0, 12)
}

const normalizeMessageList = (raw, roleKey) => {
  const source = Array.isArray(raw) ? raw : parseMaybeJSON(raw)
  return source
    .map((item) => {
      const content = String(item?.content || '').trim()
      if (!content) return null
      return {
        content,
        [roleKey]: Number(item?.[roleKey]) || 0
      }
    })
    .filter(Boolean)
    .slice(-24)
}

const resolveReplayUrl = (raw) => {
  const value = String(raw || '').trim()
  if (!value) return ''
  if (value.startsWith('/')) return value
  try {
    const url = new URL(value)
    const host = url.hostname.toLowerCase()
    if (host === '127.0.0.1' || host === 'localhost') {
      return `${url.pathname || '/'}${url.search || ''}`
    }
  } catch (_) {
    // Keep value as-is for non-standard URLs.
  }
  return value
}

const buildHistoryData = (reports = []) => {
  const points = reports
    .map((item) => {
      const rawDate = item.created_at || item.end_time || ''
      const score = safeNumber(item.average_score)
      return {
        date: rawDate ? rawDate.toString().slice(0, 10) : '',
        time: rawDate ? new Date(rawDate).getTime() : 0,
        score
      }
    })
    .filter((item) => item.date)
    .sort((a, b) => a.time - b.time)
    .slice(-10)
    .map(({ date, score }) => ({ date, score }))
  return points
}

const displayScore = computed(() => {
  const score = Number(report.value.average_score)
  return Number.isFinite(score) ? Math.round(score) : '--'
})

const scoreValue = computed(() => safeNumber(report.value.average_score))

const scoreLevel = computed(() => {
  const score = scoreValue.value
  if (score >= 90) return '优秀'
  if (score >= 80) return '良好'
  if (score >= 60) return '及格'
  return '待提升'
})

const difficultyText = computed(() => {
  const map = {
    campus_intern: '校招实习',
    campus_graduate: '校招应届',
    social_junior: '社招初级'
  }
  const key = String(report.value?.difficulty || '').trim()
  return map[key] || key || '难度未设置'
})

const dimensions = computed(() => ([
  { key: 'technical', label: '内容评估模块', value: safeNumber(report.value.technical_score) },
  { key: 'expression', label: '表达评估模块', value: safeNumber(report.value.expression_score) },
  { key: 'logic', label: '逻辑评估模块', value: safeNumber(report.value.logic_score) },
  { key: 'behavior', label: '行为评估模块', value: safeNumber(report.value.behavior_score) },
  { key: 'matching', label: '岗位匹配模块', value: safeNumber(report.value.matching_score) }
]))

const strengths = computed(() => normalizeStringList(report.value.strengths).slice(0, 5))
const weaknesses = computed(() => normalizeStringList(report.value.weaknesses).slice(0, 5))
const suggestions = computed(() => normalizeStringList(report.value.suggestions).slice(0, 5))
const qaDetails = computed(() => normalizeQADetails(report.value.qa_details))
const overallAnalysis = computed(() => String(report.value.overall_analysis || '').trim() || '尚未生成总体分析，请继续完成面试流程。')

const audioTranscripts = computed(() => normalizeMessageList(report.value.audio_transcripts, 'speaker_id'))
const chatMessages = computed(() => normalizeMessageList(report.value.chat_messages, 'sender_id'))

const resolvedScenarioType = computed(() => {
  const scenarioRaw = String(report.value.scenario_type || '').trim().toLowerCase()
  if (scenarioRaw === 'group') return 'group'
  if (scenarioRaw === 'single') return 'single'
  if (typeof report.value.is_group === 'boolean') {
    return report.value.is_group ? 'group' : 'single'
  }
  const hasMultiPlayback = String(report.value.multi_playback || '').trim() !== ''
  return hasMultiPlayback ? 'group' : 'single'
})

const isGroupScenario = computed(() => resolvedScenarioType.value === 'group')
const replayFallbackUrl = computed(() => resolveReplayUrl(report.value.replay_url))

const singlePlaybackUrl = computed(() => {
  const single = resolveReplayUrl(report.value.single_playback)
  if (single) return single
  if (!isGroupScenario.value) return replayFallbackUrl.value
  return ''
})

const multiPlaybackUrl = computed(() => {
  const multi = resolveReplayUrl(report.value.multi_playback)
  if (multi) return multi
  if (isGroupScenario.value) return replayFallbackUrl.value
  return ''
})

const groupPrimaryPlaybackUrl = computed(() => multiPlaybackUrl.value || singlePlaybackUrl.value)
const groupSecondaryPlaybackUrl = computed(() => singlePlaybackUrl.value)

const scoreGaugeOption = computed(() => ({
  backgroundColor: 'transparent',
  series: [
    {
      type: 'gauge',
      radius: '92%',
      min: 0,
      max: 100,
      axisLine: {
        lineStyle: {
          width: 16,
          color: [[1, 'rgba(187, 214, 245, 0.24)']]
        }
      },
      progress: {
        show: true,
        width: 16,
        roundCap: true,
        itemStyle: {
          color: '#79b7ff'
        }
      },
      pointer: {
        show: false
      },
      axisTick: { show: false },
      splitLine: { show: false },
      axisLabel: { show: false },
      detail: {
        valueAnimation: true,
        formatter: () => `${displayScore.value}`,
        color: '#0f172a',
        fontSize: 42,
        fontWeight: 700,
        offsetCenter: [0, '8%']
      },
      title: {
        show: true,
        offsetCenter: [0, '45%'],
        color: '#475569',
        fontSize: 13,
        formatter: '综合得分'
      },
      data: [{ value: scoreValue.value }]
    }
  ]
}))

const dimensionBarOption = computed(() => ({
  grid: { left: 16, right: 16, top: 18, bottom: 24, containLabel: true },
  xAxis: {
    type: 'category',
    data: dimensions.value.map((item) => item.label.replace('模块', '')),
    axisLabel: { color: chartTextColor, fontSize: 12 },
    axisLine: { lineStyle: { color: chartLineColor } }
  },
  yAxis: {
    type: 'value',
    min: 0,
    max: 100,
    axisLabel: { color: chartTextColor },
    splitLine: { lineStyle: { color: chartLineColor } }
  },
  tooltip: { trigger: 'axis' },
  series: [
    {
      type: 'bar',
      data: dimensions.value.map((item) => item.value),
      barWidth: 24,
      itemStyle: {
        borderRadius: [8, 8, 0, 0],
        color: '#87c5ff'
      }
    }
  ]
}))

const radarOption = computed(() => ({
  tooltip: {},
  radar: {
    radius: '68%',
    splitNumber: 4,
    axisName: { color: chartTextColor },
    splitArea: {
      areaStyle: {
        color: ['rgba(59, 100, 153, 0.12)', 'rgba(59, 100, 153, 0.2)']
      }
    },
    splitLine: { lineStyle: { color: chartLineColor } },
    axisLine: { lineStyle: { color: chartLineColor } },
    indicator: dimensions.value.map((item) => ({ name: item.label.replace('模块', ''), max: 100 }))
  },
  series: [
    {
      type: 'radar',
      data: [
        {
          value: dimensions.value.map((item) => item.value),
          areaStyle: { color: 'rgba(122, 194, 255, 0.22)' },
          lineStyle: { color: '#7ec9ff', width: 2 },
          symbolSize: 6,
          itemStyle: { color: '#b0ddff' }
        }
      ]
    }
  ]
}))

const moodCurveOption = computed(() => ({
  // 预留空间整改: 心理波动曲线目前使用占位数据，后续替换为实时情绪识别结果。
  grid: { left: 16, right: 12, top: 16, bottom: 22, containLabel: true },
  tooltip: { trigger: 'axis' },
  xAxis: {
    type: 'category',
    data: ['0:00', '0:40', '1:20', '2:00', '2:40', '3:20', '4:00'],
    axisLabel: { color: chartTextColor, fontSize: 11 },
    axisLine: { lineStyle: { color: chartLineColor } }
  },
  yAxis: {
    type: 'value',
    min: 50,
    max: 100,
    axisLabel: { color: chartTextColor, fontSize: 11 },
    splitLine: { lineStyle: { color: chartLineColor } }
  },
  series: [
    {
      type: 'line',
      smooth: true,
      data: [82, 77, 86, 80, 88, 84, 90],
      lineStyle: { width: 2, color: '#9ec8ff' },
      areaStyle: { color: 'rgba(158, 200, 255, 0.15)' },
      symbol: 'circle',
      symbolSize: 6,
      itemStyle: { color: '#c7e3ff' }
    }
  ]
}))

const trendOption = computed(() => {
  const points = historyData.value.length > 0
    ? historyData.value
    : [{ date: '当前', score: scoreValue.value }]

  return {
    grid: { left: 16, right: 12, top: 16, bottom: 22, containLabel: true },
    tooltip: { trigger: 'axis' },
    xAxis: {
      type: 'category',
      data: points.map((item) => item.date),
      axisLabel: { color: chartTextColor, fontSize: 11 },
      axisLine: { lineStyle: { color: chartLineColor } }
    },
    yAxis: {
      type: 'value',
      min: 0,
      max: 100,
      axisLabel: { color: chartTextColor, fontSize: 11 },
      splitLine: { lineStyle: { color: chartLineColor } }
    },
    series: [
      {
        type: 'line',
        smooth: true,
        data: points.map((item) => item.score),
        lineStyle: { width: 2, color: '#7fc3ff' },
        itemStyle: { color: '#bde2ff' },
        areaStyle: { color: 'rgba(127, 195, 255, 0.16)' }
      }
    ]
  }
})

const submitFeedback = () => {
  showFeedbackModal.value = false
  window.alert('感谢你的反馈，我们会用于后续评估模型优化。')
}

const handleDownloadReport = async () => {
  const id = report.value?.id || route.params.id
  if (!id) {
    window.alert('报告尚未生成，暂时无法下载')
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
    window.alert('下载报告失败，请稍后重试')
  }
}

onMounted(async () => {
  const id = route.params.id
  if (!id) {
    loading.value = false
    reportUnavailable.value = true
    loadError.value = '缺少报告编号，请从面试历史页进入。'
    return
  }

  try {
    const result = await getReport(id)

    report.value = result?.report || {}
    if (!report.value?.id) {
      reportUnavailable.value = true
      loadError.value = '报告尚未生成完成，请稍后重试。'
      return
    }

    const listRes = await getReports({ page: 1, page_size: 50 })
    const trend = buildHistoryData(listRes?.reports || [])
    historyData.value = trend.length > 0
      ? trend
      : [{ date: (report.value.created_at || '').toString().slice(0, 10) || '当前', score: scoreValue.value }]
  } catch (error) {
    console.error('获取报告失败', error)
    reportUnavailable.value = true
    const errMsg = error?.response?.data?.error || error?.message || '报告不存在或已被删除。'
    loadError.value = typeof errMsg === 'string' ? errMsg : '报告不存在或已被删除。'
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.report-page {
  --report-bg-0: #f3f8ff;
  --report-bg-1: #e7f1fb;
  --report-bg-2: #f8fbff;
  --report-card-bg: rgba(255, 255, 255, 0.95);
  --report-card-border: #d5e3f0;
  --report-panel-bg: #f7fbff;
  --report-panel-border: #d3e3f1;
  --report-title: #0f172a;
  --report-text: #334155;
  --report-muted: #64748b;
  min-height: 100vh;
  padding: 22px;
  color: var(--report-text);
  background:
    radial-gradient(circle at 8% 4%, rgba(126, 183, 244, 0.3), transparent 36%),
    radial-gradient(circle at 88% 12%, rgba(174, 209, 244, 0.33), transparent 34%),
    linear-gradient(156deg, var(--report-bg-0) 0%, var(--report-bg-1) 52%, var(--report-bg-2) 100%);
}

.report-shell {
  max-width: 1680px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.glass-card {
  border: 1px solid var(--report-card-border);
  border-radius: 24px;
  background: linear-gradient(165deg, var(--report-card-bg), rgba(247, 251, 255, 0.96));
  backdrop-filter: blur(10px);
  box-shadow: 0 16px 36px rgba(52, 87, 126, 0.13);
}

.state-card {
  padding: 28px;
}

.hero-card {
  padding: 22px;
}

.hero-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}

.hero-title {
  font-size: 32px;
  font-weight: 700;
  letter-spacing: 0.01em;
  color: var(--report-title);
}

.hero-subtitle {
  margin-top: 8px;
  font-size: 14px;
  color: var(--report-muted);
}

.hero-actions {
  display: flex;
  gap: 10px;
}

.btn {
  border-radius: 12px;
  border: 1px solid #b8d2ea;
  padding: 9px 16px;
  font-size: 13px;
  color: #12253b;
  transition: all 0.2s ease;
}

.btn-primary {
  background: linear-gradient(135deg, #5baaf8, #3f88d8);
  color: #f7fbff;
}

.btn-secondary {
  background: #edf4fd;
  color: #27476b;
}

.btn:hover {
  transform: translateY(-1px);
}

.hero-body {
  margin-top: 16px;
  display: grid;
  grid-template-columns: 320px 1fr;
  gap: 16px;
}

.hero-gauge {
  border-radius: 18px;
  border: 1px solid var(--report-panel-border);
  background: var(--report-panel-bg);
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
}

.score-tag {
  position: absolute;
  bottom: 16px;
  padding: 4px 12px;
  border-radius: 999px;
  background: #e1f0ff;
  color: #1c436d;
  font-size: 12px;
  border: 1px solid #b9d5ee;
}

.hero-dimensions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.dimension-block {
  border-radius: 16px;
  border: 1px solid var(--report-panel-border);
  background: var(--report-panel-bg);
  padding: 12px;
}

.dimension-label {
  color: var(--report-muted);
  font-size: 12px;
}

.dimension-value {
  margin-top: 4px;
  color: var(--report-title);
  font-size: 28px;
  font-weight: 700;
}

.dimension-track {
  margin-top: 8px;
  height: 6px;
  border-radius: 999px;
  background: #dfeaf3;
}

.dimension-fill {
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, #78beff 0%, #95d4ff 100%);
}

.hero-chart {
  margin-top: 16px;
  border-radius: 18px;
  border: 1px solid var(--report-panel-border);
  background: var(--report-panel-bg);
  padding: 8px;
}

.content-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 440px;
  gap: 16px;
}

.process-card {
  padding: 18px;
}

.panel-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.panel-head h2 {
  color: var(--report-title);
  font-size: 18px;
  font-weight: 600;
}

.panel-head span {
  color: var(--report-muted);
  font-size: 12px;
}

.replay-hint {
  margin-top: 8px;
  font-size: 12px;
  color: var(--report-muted);
}

.group-replay-layout {
  display: grid;
  grid-template-columns: minmax(0, 2fr) minmax(280px, 1fr);
  gap: 12px;
  margin-top: 8px;
}

.group-replay-main,
.group-replay-side {
  min-width: 0;
}

.group-replay-side-card {
  border-radius: 14px;
  border: 1px solid var(--report-panel-border);
  background: var(--report-panel-bg);
  padding: 10px;
}

.group-replay-side-card h3 {
  color: var(--report-text);
  font-size: 12px;
  margin-bottom: 8px;
}

.replay-stage {
  margin-top: 12px;
  border-radius: 16px;
  border: 1px solid var(--report-panel-border);
  overflow: hidden;
  background: #e9f2fb;
}

.replay-stage.compact {
  margin-top: 0;
}

.replay-video {
  width: 100%;
  aspect-ratio: 16 / 9;
  object-fit: contain;
  background: #0f172a;
}

.replay-placeholder {
  min-height: 240px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: var(--report-muted);
  gap: 8px;
}

.replay-placeholder.compact {
  min-height: 180px;
}

.replay-placeholder small {
  color: #859ab2;
}

.timeline-wrap {
  margin-top: 16px;
  max-height: 420px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding-right: 4px;
}

.timeline-item {
  display: grid;
  grid-template-columns: 48px 1fr;
  gap: 10px;
  border-radius: 16px;
  border: 1px solid var(--report-panel-border);
  background: var(--report-panel-bg);
  padding: 12px;
}

.timeline-index {
  width: 36px;
  height: 36px;
  border-radius: 999px;
  border: 1px solid #bad5ec;
  color: #1e4770;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  background: #deeeff;
}

.timeline-content h4 {
  color: var(--report-title);
  font-size: 14px;
  line-height: 1.45;
}

.timeline-answer {
  margin-top: 6px;
  color: var(--report-text);
  font-size: 12px;
  line-height: 1.5;
}

.timeline-answer strong {
  color: var(--report-title);
  font-weight: 600;
}

.timeline-tags {
  margin-top: 8px;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.timeline-tags span {
  font-size: 11px;
  color: #274f79;
  border: 1px solid #bfd8ec;
  border-radius: 999px;
  padding: 3px 9px;
  background: #eaf3fd;
}

.conversation-grid {
  margin-top: 16px;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.conversation-grid section {
  border-radius: 14px;
  border: 1px solid var(--report-panel-border);
  background: var(--report-panel-bg);
  padding: 10px;
}

.conversation-grid h3 {
  color: var(--report-title);
  font-size: 13px;
  margin-bottom: 8px;
}

.conversation-grid ul {
  max-height: 160px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.conversation-grid li {
  font-size: 12px;
  line-height: 1.45;
  color: var(--report-text);
}

.side-stack {
  display: grid;
  grid-template-rows: auto auto auto auto;
  gap: 12px;
}

.side-stack .glass-card {
  padding: 14px;
}

.overall-analysis {
  color: var(--report-text);
  font-size: 13px;
  line-height: 1.6;
  margin-bottom: 10px;
}

.insight-section {
  margin-top: 10px;
}

.insight-section h3 {
  font-size: 13px;
  color: var(--report-title);
  margin-bottom: 6px;
}

.insight-section ul {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.insight-section li {
  font-size: 12px;
  line-height: 1.45;
  color: var(--report-text);
}

.faded,
.empty-tip {
  color: #7f96af;
  font-size: 12px;
}

.modal-mask {
  position: fixed;
  inset: 0;
  background: rgba(12, 24, 40, 0.44);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  z-index: 80;
}

.modal-panel {
  width: min(640px, 100%);
  padding: 18px;
}

.modal-panel h3 {
  color: var(--report-title);
  font-size: 21px;
  font-weight: 700;
}

.modal-panel p {
  margin-top: 6px;
  color: var(--report-muted);
  font-size: 13px;
}

.form-grid {
  margin-top: 14px;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.form-grid label {
  display: flex;
  flex-direction: column;
  gap: 6px;
  color: var(--report-text);
  font-size: 12px;
}

.form-grid input,
.form-grid select,
.form-grid textarea {
  border-radius: 10px;
  border: 1px solid #bfd6eb;
  background: #f5faff;
  color: var(--report-title);
  padding: 8px 10px;
}

.form-grid textarea {
  min-height: 96px;
  resize: vertical;
}

.full-width {
  grid-column: 1 / -1;
}

.modal-actions {
  margin-top: 14px;
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

@media (max-width: 1280px) {
  .content-grid {
    grid-template-columns: 1fr;
  }

  .group-replay-layout {
    grid-template-columns: 1fr;
  }

  .hero-body {
    grid-template-columns: 1fr;
  }

  .side-stack {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .report-page {
    padding: 12px;
  }

  .hero-title {
    font-size: 24px;
  }

  .hero-actions {
    width: 100%;
  }

  .btn {
    flex: 1;
  }

  .hero-dimensions {
    grid-template-columns: 1fr;
  }

  .conversation-grid {
    grid-template-columns: 1fr;
  }

  .form-grid {
    grid-template-columns: 1fr;
  }
}
</style>
