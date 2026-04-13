<template>
  <div class="history-page">
    <div class="history-shell">
      <header class="history-header">
        <div>
          <h1>面试历史</h1>
          <p>查看历史记录、快速预览报告详情，并跳转完整反馈页。</p>
        </div>
        <router-link to="/student/interview" class="action-btn">去面试</router-link>
      </header>

      <section class="glass-card table-card">
        <div class="filter-row">
          <button
            v-for="item in filterOptions"
            :key="item.key"
            class="filter-pill"
            :class="{ active: activeFilter === item.key }"
            @click="activeFilter = item.key"
          >
            {{ item.label }}
          </button>
        </div>

        <div v-if="loading" class="state-box">加载中...</div>

        <div v-else-if="filteredRecords.length === 0" class="state-box">
          <div class="empty-icon">
            <FileText class="h-8 w-8" />
          </div>
          <p>当前分类暂无记录</p>
        </div>

        <div v-else class="table-wrap">
          <table>
            <thead>
              <tr>
                <th>日期</th>
                <th>岗位</th>
                <th>难度</th>
                <th>综合得分</th>
                <th>状态</th>
                <th class="align-right">操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="record in filteredRecords" :key="record.interview_id">
                <td>{{ formatDate(record.created_at) }}</td>
                <td>{{ record.position || '未设置岗位' }}</td>
                <td>
                  <span class="difficulty-tag" :class="difficultyBadgeClass(record.difficulty)">
                    {{ formatDifficulty(record.difficulty) }}
                  </span>
                </td>
                <td>
                  <span v-if="record.average_score !== null" class="score-text" :class="getScoreColor(record.average_score)">
                    {{ record.average_score }}
                  </span>
                  <span v-else class="faded">--</span>
                </td>
                <td>
                  <span class="status-text" :class="record.is_successful ? (record.report_id ? 'status-ok' : 'status-warn') : 'status-off'">
                    {{ record.is_successful ? (record.report_id ? '报告已生成' : '待生成报告') : '面试中断' }}
                  </span>
                </td>
                <td class="align-right">
                  <div v-if="record.is_successful" class="row-actions">
                    <button class="text-btn" @click="openHistoryDetail(record)">详情弹窗</button>
                    <button class="text-btn" @click="viewReport(record)">
                      {{ record.report_id ? '查看报告' : '生成并查看' }}
                      <ChevronRight class="h-4 w-4" />
                    </button>
                  </div>
                  <span v-else class="faded">不可操作</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </div>

    <div v-if="detailVisible" class="modal-mask" @click.self="closeHistoryDetail">
      <div class="glass-card detail-modal">
        <div class="detail-head">
          <div>
            <h2>面试历史详情</h2>
            <p>{{ detailReport.position || detailRecord?.position || '岗位未设置' }} · {{ formatDifficulty(detailReport.difficulty || detailRecord?.difficulty) }}</p>
          </div>
          <button class="action-btn" @click="closeHistoryDetail">关闭</button>
        </div>

        <div v-if="detailLoading" class="state-box">正在加载详情...</div>

        <div v-else class="detail-content">
          <section class="summary-strip">
            <div class="summary-card">
              <p>综合得分</p>
              <h3>{{ detailScore }}</h3>
            </div>
            <div v-for="item in detailDimensions" :key="item.key" class="summary-card">
              <p>{{ item.label }}</p>
              <h3>{{ item.value }}</h3>
            </div>
          </section>

          <section class="detail-grid">
            <article class="glass-subcard">
              <h4>能力雷达图</h4>
              <GlassEchart :option="detailRadarOption" height="260px" />
            </article>

            <article class="glass-subcard">
              <h4>心理波动曲线（预留）</h4>
              <GlassEchart :option="detailMoodOption" height="260px" />
            </article>
          </section>

          <section class="detail-grid">
            <article class="glass-subcard">
              <h4>AI 总结</h4>
              <p class="plain-text">{{ detailOverall }}</p>
            </article>

            <article class="glass-subcard">
              <h4>建议摘要</h4>
              <ul class="plain-list">
                <li v-for="(item, idx) in detailSuggestions" :key="`suggestion-${idx}`">{{ item }}</li>
                <li v-if="detailSuggestions.length === 0" class="faded">暂无建议项</li>
              </ul>
            </article>
          </section>

          <div class="detail-footer">
            <button class="action-btn" @click="viewReport(detailRecord)">查看完整报告页</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ChevronRight, FileText } from 'lucide-vue-next'
import dayjs from 'dayjs'
import { generateReport, getReport, getReports } from '../api/report'
import { getInterviews } from '../api/interview'
import GlassEchart from '../components/report/GlassEchart.vue'

const router = useRouter()

const records = ref([])
const loading = ref(true)
const activeFilter = ref('completed')

const detailVisible = ref(false)
const detailLoading = ref(false)
const detailRecord = ref(null)
const detailReport = ref({})

const filterOptions = [
  { key: 'all', label: '全部' },
  { key: 'completed', label: '面试完成' },
  { key: 'interrupted', label: '面试中断' }
]

const safeNumber = (value) => {
  const n = Number(value)
  if (!Number.isFinite(n)) return 0
  return Math.max(0, Math.min(100, Math.round(n)))
}

const parseStringList = (raw) => {
  if (Array.isArray(raw)) return raw.map((item) => String(item || '').trim()).filter(Boolean)
  if (typeof raw === 'string') {
    const trimmed = raw.trim()
    if (!trimmed) return []
    try {
      const parsed = JSON.parse(trimmed)
      if (Array.isArray(parsed)) return parsed.map((item) => String(item || '').trim()).filter(Boolean)
    } catch (_) {
      return trimmed.split('|').map((item) => item.trim()).filter(Boolean)
    }
  }
  return []
}

const isSuccessfulInterview = (interview, report) => {
  if (report?.id || interview?.report?.id) {
    return true
  }

  if (!interview || interview.status !== 'completed') return false

  const plannedCount = Number(interview.total_question_target) || 0
  const arrangedCount = Array.isArray(interview.questions) ? interview.questions.length : 0
  const completedCount = Number(interview.current_index) || 0

  const target = Math.max(plannedCount, arrangedCount)
  if (target <= 0) {
    return true
  }

  return completedCount >= target
}

const filteredRecords = computed(() => {
  if (activeFilter.value === 'completed') {
    return records.value.filter((item) => item.is_successful)
  }
  if (activeFilter.value === 'interrupted') {
    return records.value.filter((item) => item.is_interrupted)
  }
  return records.value
})

const detailScore = computed(() => {
  const score = Number(detailReport.value.average_score ?? detailRecord.value?.average_score)
  return Number.isFinite(score) ? Math.round(score) : '--'
})

const detailDimensions = computed(() => ([
  { key: 'technical', label: '内容评估', value: safeNumber(detailReport.value.technical_score) },
  { key: 'expression', label: '表达评估', value: safeNumber(detailReport.value.expression_score) },
  { key: 'logic', label: '逻辑评估', value: safeNumber(detailReport.value.logic_score) },
  { key: 'behavior', label: '行为评估', value: safeNumber(detailReport.value.behavior_score) }
]))

const detailOverall = computed(() => {
  return String(detailReport.value.overall_analysis || '').trim() || '暂无总体分析。'
})

const detailSuggestions = computed(() => parseStringList(detailReport.value.suggestions).slice(0, 5))

const chartTextColor = 'rgba(222, 238, 255, 0.74)'
const chartLineColor = 'rgba(154, 190, 232, 0.24)'

const detailRadarOption = computed(() => ({
  tooltip: {},
  radar: {
    radius: '65%',
    splitNumber: 4,
    axisName: { color: chartTextColor },
    splitArea: {
      areaStyle: { color: ['rgba(70, 110, 160, 0.12)', 'rgba(70, 110, 160, 0.2)'] }
    },
    splitLine: { lineStyle: { color: chartLineColor } },
    axisLine: { lineStyle: { color: chartLineColor } },
    indicator: detailDimensions.value.map((item) => ({ name: item.label, max: 100 }))
  },
  series: [
    {
      type: 'radar',
      data: [
        {
          value: detailDimensions.value.map((item) => item.value),
          areaStyle: { color: 'rgba(122, 194, 255, 0.25)' },
          lineStyle: { color: '#7fc9ff', width: 2 },
          symbolSize: 6,
          itemStyle: { color: '#b0deff' }
        }
      ]
    }
  ]
}))

const detailMoodOption = computed(() => ({
  // 预留空间整改: 心理波动曲线目前使用占位数据，后续接入情绪/行为识别模型。
  grid: { left: 14, right: 10, top: 16, bottom: 22, containLabel: true },
  tooltip: { trigger: 'axis' },
  xAxis: {
    type: 'category',
    data: ['片段1', '片段2', '片段3', '片段4', '片段5'],
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
      data: [84, 76, 88, 79, 83],
      lineStyle: { width: 2, color: '#9acbff' },
      areaStyle: { color: 'rgba(154, 203, 255, 0.15)' },
      symbolSize: 6,
      itemStyle: { color: '#c9e5ff' }
    }
  ]
}))

const fetchReports = async () => {
  try {
    const [reportSettled, interviewSettled] = await Promise.allSettled([
      getReports({ page: 1, page_size: 100 }),
      getInterviews({ page: 1, page_size: 100 })
    ])

    const reportList = reportSettled.status === 'fulfilled' ? (reportSettled.value?.reports || []) : []
    const interviewList = interviewSettled.status === 'fulfilled' ? (interviewSettled.value?.interviews || []) : []
    const reportMap = new Map(reportList.map((item) => [item.interview_id, item]))

    records.value = interviewList.map((interview) => {
      const report = reportMap.get(interview.id)
      const isSuccessful = isSuccessfulInterview(interview, report)
      return {
        interview_id: interview.id,
        position: interview.position,
        difficulty: interview.difficulty,
        status: interview.status,
        current_index: interview.current_index,
        total_question_target: interview.total_question_target,
        questions: interview.questions,
        created_at: interview.created_at,
        average_score: report?.average_score ?? null,
        report_id: report?.id ?? interview.report?.id ?? null,
        is_successful: isSuccessful,
        is_interrupted: !isSuccessful
      }
    })

    if (records.value.length === 0 && reportList.length > 0) {
      records.value = reportList.map((report) => ({
        interview_id: report.interview_id,
        position: report.position,
        difficulty: report.difficulty,
        status: 'completed',
        created_at: report.created_at,
        average_score: report.average_score ?? null,
        report_id: report.id,
        is_successful: true,
        is_interrupted: false
      }))
    }
  } catch (error) {
    console.error('Failed to fetch reports:', error)
  } finally {
    loading.value = false
  }
}

const ensureReportForRecord = async (record) => {
  if (!record) return null
  if (record.report_id) return record.report_id
  if (record.status !== 'completed') return null

  const generated = await generateReport({ interview_id: record.interview_id })
  const reportId = generated?.report?.id
  if (reportId) {
    const target = records.value.find((item) => item.interview_id === record.interview_id)
    if (target) {
      target.report_id = reportId
      target.average_score = generated?.report?.average_score ?? target.average_score
    }
  }
  return reportId || null
}

const openHistoryDetail = async (record) => {
  if (!record?.is_successful) return
  detailVisible.value = true
  detailLoading.value = true
  detailRecord.value = record
  detailReport.value = {}

  try {
    const reportId = await ensureReportForRecord(record)
    if (!reportId) return
    const res = await getReport(reportId)
    detailReport.value = res?.report || {}
  } catch (error) {
    console.error('Failed to open history detail:', error)
  } finally {
    detailLoading.value = false
  }
}

const closeHistoryDetail = () => {
  detailVisible.value = false
  detailLoading.value = false
}

const viewReport = async (record) => {
  if (!record?.is_successful) return

  try {
    const reportId = await ensureReportForRecord(record)
    if (reportId) {
      router.push(`/student/report/${reportId}`)
    }
  } catch (error) {
    console.error('Failed to generate report:', error)
  }
}

const formatDate = (date) => dayjs(date).format('YYYY-MM-DD HH:mm')

const formatDifficulty = (difficulty) => {
  const map = {
    campus_intern: '校招实习',
    campus_graduate: '校招应届',
    social_junior: '社招初级',
    Junior: '初级',
    Middle: '中级',
    Mid: '中级',
    Senior: '高级'
  }
  return map[difficulty] || difficulty || '未知'
}

const difficultyBadgeClass = (difficulty) => {
  if (difficulty === 'social_junior' || difficulty === 'Senior') {
    return 'tag-rose'
  }
  if (difficulty === 'campus_graduate' || difficulty === 'Middle' || difficulty === 'Mid') {
    return 'tag-amber'
  }
  return 'tag-emerald'
}

const getScoreColor = (score) => {
  if (score >= 80) return 'score-good'
  if (score >= 60) return 'score-mid'
  return 'score-bad'
}

onMounted(() => {
  fetchReports()
})
</script>

<style scoped>
.history-page {
  min-height: 100vh;
  padding: 24px;
  background:
    radial-gradient(circle at 14% 8%, rgba(74, 143, 224, 0.34), transparent 36%),
    radial-gradient(circle at 86% 20%, rgba(128, 186, 255, 0.22), transparent 34%),
    linear-gradient(145deg, #071224 0%, #102342 55%, #0f1a35 100%);
}

.history-shell {
  max-width: 1480px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.history-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}

.history-header h1 {
  color: #f2f8ff;
  font-size: 32px;
  font-weight: 700;
}

.history-header p {
  margin-top: 8px;
  color: rgba(208, 228, 252, 0.86);
  font-size: 14px;
}

.action-btn {
  border-radius: 12px;
  border: 1px solid rgba(177, 213, 252, 0.48);
  padding: 8px 14px;
  font-size: 13px;
  color: #f4faff;
  background: rgba(66, 131, 210, 0.54);
  transition: all 0.2s ease;
}

.action-btn:hover {
  transform: translateY(-1px);
}

.glass-card {
  border: 1px solid rgba(184, 215, 255, 0.32);
  border-radius: 24px;
  background: linear-gradient(160deg, rgba(18, 39, 69, 0.64), rgba(25, 53, 88, 0.4));
  backdrop-filter: blur(14px);
  box-shadow: 0 20px 60px rgba(4, 18, 42, 0.36);
}

.table-card {
  padding: 16px;
}

.filter-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
}

.filter-pill {
  border-radius: 999px;
  border: 1px solid rgba(166, 204, 247, 0.26);
  background: rgba(29, 61, 102, 0.4);
  color: rgba(221, 238, 255, 0.85);
  font-size: 12px;
  padding: 6px 12px;
}

.filter-pill.active {
  border-color: rgba(188, 224, 255, 0.66);
  background: rgba(108, 178, 255, 0.24);
}

.state-box {
  min-height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  color: rgba(188, 211, 238, 0.74);
  gap: 10px;
}

.empty-icon {
  width: 64px;
  height: 64px;
  border-radius: 999px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(184, 216, 255, 0.4);
  color: rgba(200, 224, 250, 0.9);
}

.table-wrap {
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th,
td {
  padding: 14px 12px;
  border-bottom: 1px solid rgba(173, 206, 248, 0.2);
  text-align: left;
  white-space: nowrap;
  color: rgba(225, 239, 255, 0.92);
  font-size: 13px;
}

th {
  color: rgba(193, 215, 240, 0.75);
  font-size: 12px;
}

.align-right {
  text-align: right;
}

.difficulty-tag {
  display: inline-flex;
  border-radius: 999px;
  border: 1px solid;
  padding: 3px 10px;
  font-size: 11px;
}

.tag-emerald {
  color: #ccffe8;
  background: rgba(83, 173, 139, 0.25);
  border-color: rgba(163, 239, 204, 0.5);
}

.tag-amber {
  color: #ffe7c7;
  background: rgba(178, 132, 77, 0.24);
  border-color: rgba(255, 214, 156, 0.5);
}

.tag-rose {
  color: #ffd7e2;
  background: rgba(185, 86, 119, 0.23);
  border-color: rgba(255, 188, 208, 0.5);
}

.score-text {
  font-weight: 700;
}

.score-good {
  color: #8cf4c6;
}

.score-mid {
  color: #ffd49c;
}

.score-bad {
  color: #ffb2c8;
}

.status-text {
  font-size: 12px;
}

.status-ok {
  color: #88f2c2;
}

.status-warn {
  color: #ffd49c;
}

.status-off {
  color: rgba(190, 211, 236, 0.65);
}

.row-actions {
  display: inline-flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
}

.text-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: #b6deff;
  font-size: 12px;
  transition: color 0.2s ease;
}

.text-btn:hover {
  color: #d3ebff;
}

.faded {
  color: rgba(177, 199, 224, 0.68);
  font-size: 12px;
}

.modal-mask {
  position: fixed;
  inset: 0;
  background: rgba(4, 10, 20, 0.68);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  z-index: 90;
}

.detail-modal {
  width: min(1180px, 100%);
  max-height: 92vh;
  overflow-y: auto;
  padding: 18px;
}

.detail-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
}

.detail-head h2 {
  color: #f4faff;
  font-size: 24px;
  font-weight: 700;
}

.detail-head p {
  margin-top: 6px;
  color: rgba(205, 228, 252, 0.84);
  font-size: 13px;
}

.detail-content {
  margin-top: 12px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.summary-strip {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 10px;
}

.summary-card {
  border-radius: 14px;
  border: 1px solid rgba(172, 207, 248, 0.24);
  background: rgba(11, 27, 49, 0.56);
  padding: 10px;
}

.summary-card p {
  color: rgba(201, 225, 251, 0.78);
  font-size: 12px;
}

.summary-card h3 {
  margin-top: 6px;
  color: #f2f9ff;
  font-size: 28px;
  font-weight: 700;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.glass-subcard {
  border-radius: 16px;
  border: 1px solid rgba(171, 206, 247, 0.2);
  background: rgba(9, 24, 45, 0.58);
  padding: 12px;
}

.glass-subcard h4 {
  color: #f1f8ff;
  font-size: 14px;
  margin-bottom: 8px;
}

.plain-text {
  color: rgba(210, 230, 249, 0.92);
  font-size: 13px;
  line-height: 1.6;
}

.plain-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.plain-list li {
  color: rgba(210, 230, 249, 0.92);
  font-size: 13px;
  line-height: 1.5;
}

.detail-footer {
  display: flex;
  justify-content: flex-end;
}

@media (max-width: 960px) {
  .history-page {
    padding: 12px;
  }

  .summary-strip {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .detail-grid {
    grid-template-columns: 1fr;
  }

  .history-header {
    flex-direction: column;
  }

  .action-btn {
    width: 100%;
    text-align: center;
  }
}
</style>
