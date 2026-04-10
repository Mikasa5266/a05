<template>
  <div class="space-y-6">
    <header class="flex flex-col gap-4 xl:flex-row xl:items-end xl:justify-between">
      <div class="space-y-2">
        <p class="text-xs font-semibold uppercase tracking-[0.28em] text-slate-400">Resume Intelligence</p>
        <h1 class="text-3xl font-semibold tracking-tight text-slate-950 md:text-4xl">
          结构化简历分析中心
        </h1>
        <p class="max-w-3xl text-sm leading-7 text-slate-500 md:text-[15px]">
          把上传的简历转换成岗位匹配、技能图谱、项目深剖和面试准备建议，不再停留在“只给一段点评”的粗糙展示。
        </p>
      </div>

      <div class="flex flex-wrap items-center gap-3">
        <button
          class="inline-flex items-center gap-2 rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm font-medium text-slate-700 transition hover:border-slate-300 hover:bg-slate-50"
          :disabled="refreshing || uploading"
          @click="refreshLatest"
        >
          <RefreshCw class="h-4 w-4" :class="{ 'animate-spin': refreshing }" />
          刷新结果
        </button>
        <button
          class="inline-flex items-center gap-2 rounded-2xl bg-slate-950 px-4 py-3 text-sm font-semibold text-white transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60"
          :disabled="uploading"
          @click="openFilePicker"
        >
          <Upload class="h-4 w-4" />
          {{ uploading ? '正在解析...' : '上传并分析简历' }}
        </button>
      </div>
    </header>

    <section class="upload-shell">
      <div class="grid gap-5 xl:grid-cols-[minmax(0,1.12fr)_360px]">
        <div
          class="upload-dropzone"
          :class="{ 'upload-dropzone-active': dragActive }"
          @dragenter.prevent="dragActive = true"
          @dragover.prevent="dragActive = true"
          @dragleave.prevent="dragActive = false"
          @drop.prevent="handleDrop"
        >
          <input
            ref="fileInputRef"
            type="file"
            class="hidden"
            accept=".pdf,.doc,.docx,.txt,.png,.jpg,.jpeg"
            @change="handleFileChange"
          />

          <div class="flex h-full flex-col justify-between gap-6">
            <div class="space-y-4">
              <div class="inline-flex h-12 w-12 items-center justify-center rounded-2xl bg-white text-slate-900 shadow-sm">
                <FileText class="h-6 w-6" />
              </div>
              <div class="space-y-2">
                <h2 class="text-2xl font-semibold tracking-tight text-slate-950">
                  上传最新简历，生成高质量结构化提取
                </h2>
                <p class="max-w-2xl text-sm leading-7 text-slate-600">
                  支持 PDF、DOC、DOCX、TXT 以及常见图片简历。上传后会自动解析并刷新下方分析工作台。
                </p>
              </div>
            </div>

            <div class="flex flex-wrap items-center gap-3">
              <button
                class="inline-flex items-center gap-2 rounded-2xl bg-slate-950 px-4 py-3 text-sm font-semibold text-white transition hover:bg-slate-800 disabled:opacity-60"
                :disabled="uploading"
                @click="openFilePicker"
              >
                <Upload class="h-4 w-4" />
                {{ selectedFileName ? '重新选择文件' : '选择文件' }}
              </button>
              <div class="rounded-2xl border border-slate-200 bg-white/70 px-4 py-3 text-sm text-slate-600">
                {{ selectedFileName || '可直接拖拽文件到此区域' }}
              </div>
            </div>
          </div>
        </div>

        <aside class="status-card">
          <div class="space-y-2">
            <p class="text-xs font-semibold uppercase tracking-[0.24em] text-slate-400">Latest Snapshot</p>
            <h2 class="text-xl font-semibold text-slate-950">解析状态</h2>
          </div>

          <div class="space-y-3">
            <article
              v-for="item in statusCards"
              :key="item.label"
              class="rounded-2xl border border-slate-200 bg-white px-4 py-3"
            >
              <p class="text-[11px] uppercase tracking-[0.18em] text-slate-400">{{ item.label }}</p>
              <p class="mt-2 text-sm font-medium text-slate-900">{{ item.value }}</p>
            </article>
          </div>

          <div class="rounded-2xl border border-dashed border-slate-300 bg-slate-50 px-4 py-4">
            <div class="flex items-start gap-3">
              <Sparkles class="mt-0.5 h-5 w-5 text-sky-600" />
              <p class="text-sm leading-6 text-slate-600">
                {{ dashboardData ? '当前页面已切换到结构化分析视图，展示岗位匹配、技能图谱和项目深剖。' : '上传后会在这里生成结构化提取结果，并同步刷新完整分析工作台。' }}
              </p>
            </div>
          </div>
        </aside>
      </div>
    </section>

    <section v-if="initialLoading" class="loading-shell">
      <div class="loading-line w-40" />
      <div class="loading-line w-full" />
      <div class="loading-line w-10/12" />
      <div class="loading-grid">
        <div class="loading-block" />
        <div class="loading-block" />
        <div class="loading-block" />
      </div>
    </section>

    <section v-else-if="dashboardData">
      <ResumeAnalysisDashboard :view-model="dashboardData" />
    </section>

    <section v-else class="empty-shell">
      <div class="max-w-2xl space-y-4">
        <p class="text-xs font-semibold uppercase tracking-[0.24em] text-slate-400">No Analysis Yet</p>
        <h2 class="text-3xl font-semibold tracking-tight text-slate-950">
          先上传一份简历，我们再展示完整分析视图
        </h2>
        <p class="text-sm leading-7 text-slate-500">
          这套页面会在拿到结构化 JSON 后，自动生成岗位匹配结果、技能证据分组、项目经验拆解和 Drill Plan。
        </p>
      </div>

      <button
        class="inline-flex items-center gap-2 rounded-2xl bg-slate-950 px-5 py-3 text-sm font-semibold text-white transition hover:bg-slate-800"
        @click="openFilePicker"
      >
        <Upload class="h-4 w-4" />
        立即上传简历
      </button>

      <p v-if="fetchError" class="text-sm text-rose-600">{{ fetchError }}</p>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import dayjs from 'dayjs'
import { ElMessage } from 'element-plus'
import { FileText, RefreshCw, Sparkles, Upload } from 'lucide-vue-next'
import ResumeAnalysisDashboard from '../../components/resume/ResumeAnalysisDashboard.vue'
import { getLatestResumeAnalysis, uploadResumeForAnalysis } from '../../api/resume'

const fileInputRef = ref(null)
const dragActive = ref(false)
const uploading = ref(false)
const refreshing = ref(false)
const initialLoading = ref(true)
const selectedFileName = ref('')
const fetchError = ref('')
const analysisState = ref({
  analysis: null,
  record: null,
})

const dashboardData = computed(() =>
  normalizeResumeAnalysis(analysisState.value.analysis, analysisState.value.record)
)

const statusCards = computed(() => {
  if (dashboardData.value) {
    return [
      { label: '最近文件', value: dashboardData.value.recordMeta.fileName },
      { label: '更新时间', value: dashboardData.value.recordMeta.updatedAt },
      { label: '解析模式', value: dashboardData.value.recordMeta.parserMode },
      { label: '置信度', value: `${dashboardData.value.recordMeta.confidenceScore}%` },
    ]
  }

  return [
    { label: '最近文件', value: selectedFileName.value || '暂无上传记录' },
    { label: '更新时间', value: '等待第一次解析' },
    { label: '解析模式', value: uploading.value ? '正在解析' : '未开始' },
    { label: '置信度', value: '--' },
  ]
})

const openFilePicker = () => {
  fileInputRef.value?.click()
}

const handleFileChange = (event) => {
  const file = event.target?.files?.[0]
  if (!file) return
  analyzeFile(file)
  event.target.value = ''
}

const handleDrop = (event) => {
  dragActive.value = false
  const file = event.dataTransfer?.files?.[0]
  if (!file) return
  analyzeFile(file)
}

const refreshLatest = async () => {
  refreshing.value = true
  await fetchLatest({ silent: false })
  refreshing.value = false
}

const fetchLatest = async ({ silent = true } = {}) => {
  fetchError.value = ''
  try {
    const response = await getLatestResumeAnalysis()
    analysisState.value = {
      analysis: response?.analysis || null,
      record: response?.record || null,
    }
  } catch (error) {
    const status = Number(error?.response?.status || 0)
    if (status === 404) {
      analysisState.value = { analysis: null, record: null }
      return
    }
    fetchError.value = error?.message || '加载简历分析失败'
    if (!silent) {
      ElMessage.error(fetchError.value)
    }
  } finally {
    initialLoading.value = false
  }
}

const analyzeFile = async (file) => {
  const validationError = validateResumeFile(file)
  if (validationError) {
    ElMessage.error(validationError)
    return
  }

  uploading.value = true
  selectedFileName.value = file.name
  fetchError.value = ''

  const formData = new FormData()
  formData.append('file', file)
  formData.append('source', 'resume_center')

  try {
    const response = await uploadResumeForAnalysis(formData)
    analysisState.value = {
      analysis: response?.analysis || null,
      record: response?.record || null,
    }
    ElMessage.success('简历解析完成，已切换到结构化分析视图')
  } catch (error) {
    fetchError.value = error?.message || '简历解析失败，请稍后重试'
    ElMessage.error(fetchError.value)
  } finally {
    uploading.value = false
    initialLoading.value = false
  }
}

onMounted(() => {
  fetchLatest()
})

const SKILL_GROUP_LABELS = {
  programming_languages: '编程语言',
  frameworks: '框架与中间件',
  databases: '数据库',
  cloud_devops: '云原生与 DevOps',
  ai_data: 'AI / 数据能力',
  tooling: '工程工具',
  product_business: '产品与业务理解',
  others: '其他能力',
}

const BREAKDOWN_LABELS = [
  ['skill_depth', '技能深度'],
  ['project_relevance', '项目相关性'],
  ['domain_alignment', '领域贴合度'],
  ['delivery_impact', '交付影响'],
]

function normalizeResumeAnalysis(analysis, record) {
  if (!analysis?.structured_resume) {
    return null
  }

  const structured = analysis.structured_resume || {}
  const personalInfo = structured.personal_info || {}
  const matchResults = normalizeMatchResults(analysis.match_results, analysis.best_match)
  const bestMatch = matchResults[0] || createFallbackMatch(analysis.best_match)
  const skillGroups = normalizeSkillGroups(structured.skill_graph)
  const totalSkills = skillGroups.reduce((sum, group) => sum + group.count, 0)
  const projects = normalizeProjects(structured.project_experience)
  const workExperience = normalizeWorkExperience(structured.work_experience)
  const education = normalizeEducation(structured.education)
  const optimization = normalizeOptimization(analysis.optimization)
  const risks = normalizeRisks(analysis.risk_report)
  const integration = normalizeIntegration(analysis.integration, bestMatch)
  const interviewQuestions = normalizeInterviewQuestions(
    analysis.interview_questions,
    integration.questionRecommendations
  )
  const confidenceScore = toNumber(analysis.confidence_score || record?.confidence_score)

  const summary =
    firstText(
      structured.professional_summary,
      bestMatch.analysis,
      '已完成简历结构化提取，等待进一步补充摘要。'
    ) || '已完成简历结构化提取。'

  const focusChips = uniqueStrings([
    ...(safeArray(structured.career_intent?.target_roles)),
    structured.career_intent?.seniority,
    ...(safeArray(structured.career_intent?.target_cities)),
    ...(safeArray(structured.career_intent?.target_industries)),
    ...bestMatch.hitSkills.slice(0, 3),
  ]).slice(0, 8)

  const contactCards = normalizeContactCards(personalInfo)
  const metrics = [
    {
      label: '结构化置信度',
      value: `${confidenceScore}%`,
      detail: '当前解析结果的整体稳定性与字段完整度',
    },
    {
      label: '技能证据',
      value: `${totalSkills}`,
      detail: '已识别的技能与证据条目数量',
    },
    {
      label: '项目深剖',
      value: `${projects.length}`,
      detail: '可直接用于面试展开的项目经历数量',
    },
    {
      label: '优化动作',
      value: `${optimization.length}`,
      detail: '已生成的优先优化建议数量',
    },
  ]

  return {
    person: {
      name: firstText(personalInfo.name, '候选人'),
      email: firstText(personalInfo.email, '--'),
      phone: firstText(personalInfo.phone, '--'),
      location: firstText(personalInfo.location, '--'),
      github: firstText(personalInfo.github),
      portfolio: firstText(personalInfo.portfolio),
      linkedin: firstText(personalInfo.linkedin),
    },
    summary,
    focusChips,
    contactCards,
    metrics,
    bestMatch,
    matchResults,
    radarIndicators: bestMatch.breakdown,
    skillGroups,
    projects,
    workExperience,
    education,
    highlights: safeArray(structured.highlights).slice(0, 8),
    concerns: safeArray(structured.concerns).slice(0, 8),
    optimization,
    risks,
    interviewQuestions,
    drillPlan: integration.drillPlan,
    weakPoints: integration.weakPoints,
    integration,
    rawPreview: firstText(structured.raw_preview),
    recordMeta: {
      fileName: firstText(record?.file_name, selectedFileName.value, '最近上传的简历'),
      parserMode: firstText(analysis.parser_mode, record?.parser_mode, 'structured'),
      confidenceScore,
      updatedAt: formatDateTime(record?.updated_at || record?.created_at),
    },
  }
}

function normalizeContactCards(personalInfo) {
  const cards = [
    buildContactCard('邮箱', personalInfo?.email, (value) => `mailto:${value}`),
    buildContactCard('电话', personalInfo?.phone),
    buildContactCard('所在城市', personalInfo?.location),
    buildContactCard('作品入口', personalInfo?.portfolio || personalInfo?.github || personalInfo?.linkedin, normalizeExternalLink),
  ].filter(Boolean)

  return cards.length
    ? cards
    : [
        { label: '邮箱', value: '未提供' },
        { label: '电话', value: '未提供' },
      ]
}

function buildContactCard(label, value, hrefBuilder) {
  const text = firstText(value)
  if (!text) return null
  const href = typeof hrefBuilder === 'function' ? hrefBuilder(text) : ''
  return { label, value: text, href }
}

function normalizeMatchResults(matchResults, bestMatch) {
  const merged = [...safeArray(matchResults)]
  if (bestMatch && !merged.some((item) => item?.position_code === bestMatch.position_code)) {
    merged.unshift(bestMatch)
  }

  const normalized = merged
    .map((item) => {
      const breakdownSource = item?.score_breakdown || {}
      return {
        positionCode: firstText(item?.position_code, 'unknown'),
        positionName: firstText(item?.position_name, '待确认岗位'),
        roleKey: firstText(item?.role_key, '--'),
        score: toNumber(item?.score),
        analysis: firstText(item?.analysis, '暂无岗位分析说明'),
        hitSkills: safeArray(item?.hit_skills).slice(0, 6),
        gapSkills: safeArray(item?.gap_skills).slice(0, 6),
        requirements: safeArray(item?.requirements).slice(0, 6),
        breakdown: BREAKDOWN_LABELS.map(([key, label]) => ({
          key,
          label,
          value: toNumber(breakdownSource?.[key]),
        })),
      }
    })
    .sort((a, b) => b.score - a.score)

  return normalized.length ? normalized : [createFallbackMatch(bestMatch)]
}

function createFallbackMatch(bestMatch) {
  const breakdownSource = bestMatch?.score_breakdown || {}
  return {
    positionCode: firstText(bestMatch?.position_code, 'unknown'),
    positionName: firstText(bestMatch?.position_name, '待确认岗位'),
    roleKey: firstText(bestMatch?.role_key, '--'),
    score: toNumber(bestMatch?.score),
    analysis: firstText(bestMatch?.analysis, '暂无岗位分析说明'),
    hitSkills: safeArray(bestMatch?.hit_skills).slice(0, 6),
    gapSkills: safeArray(bestMatch?.gap_skills).slice(0, 6),
    requirements: safeArray(bestMatch?.requirements).slice(0, 6),
    breakdown: BREAKDOWN_LABELS.map(([key, label]) => ({
      key,
      label,
      value: toNumber(breakdownSource?.[key]),
    })),
  }
}

function normalizeSkillGroups(skillGraph) {
  return Object.entries(SKILL_GROUP_LABELS)
    .map(([key, label]) => {
      const items = safeArray(skillGraph?.[key]).map((item) => ({
        name: firstText(item?.name, '未命名能力'),
        level: normalizeLevel(item?.level),
        evidence: firstText(item?.evidence),
        lastUsed: firstText(item?.last_used),
      }))
      return {
        key,
        label,
        count: items.length,
        summary: items.slice(0, 2).map((item) => item.name).join(' · ') || '暂无技能锚点',
        items: items.slice(0, 8),
      }
    })
    .filter((group) => group.count > 0)
}

function normalizeProjects(items) {
  return safeArray(items).map((item) => ({
    name: firstText(item?.name, '未命名项目'),
    role: firstText(item?.role, '项目角色待补充'),
    period: formatPeriod(item?.start_date, item?.end_date),
    background: firstText(item?.background),
    summary: firstText(item?.summary, item?.background, '暂无项目摘要'),
    techStack: safeArray(item?.tech_stack).slice(0, 8),
    highlights: safeArray(item?.highlights).slice(0, 6),
    impact: safeArray(item?.impact).slice(0, 6),
  }))
}

function normalizeWorkExperience(items) {
  return safeArray(items).map((item) => ({
    company: firstText(item?.company, '未标注公司'),
    role: firstText(item?.role, '未标注岗位'),
    period: formatPeriod(item?.start_date, item?.end_date, item?.duration),
    summary: firstText(item?.summary, '暂无经历摘要'),
    responsibilities: safeArray(item?.responsibilities).slice(0, 5),
    achievements: safeArray(item?.achievements).slice(0, 5),
  }))
}

function normalizeEducation(items) {
  return safeArray(items).map((item) => ({
    school: firstText(item?.school, '未标注院校'),
    degree: firstText(item?.degree, '学历待补充'),
    major: firstText(item?.major, '专业待补充'),
    period: formatPeriod(item?.start_date, item?.end_date),
    gpa: firstText(item?.gpa),
    ranking: firstText(item?.ranking),
    highlights: safeArray(item?.highlights).slice(0, 5),
  }))
}

function normalizeOptimization(items) {
  return safeArray(items).map((item) => ({
    title: firstText(item?.title, '建议项'),
    action: firstText(item?.action, '请补充具体动作'),
    rationale: firstText(item?.rationale, '暂无原因说明'),
    priority: firstText(item?.priority, 'medium'),
  }))
}

function normalizeRisks(items) {
  return safeArray(items).map((item) => ({
    level: firstText(item?.level, 'medium'),
    item: firstText(item?.item, '待关注问题'),
    detail: firstText(item?.detail, '暂无详细说明'),
    evidence: safeArray(item?.evidence).slice(0, 4),
  }))
}

function normalizeInterviewQuestions(items, fallbackQuestions) {
  const normalized = safeArray(items).map((item) => ({
    question: firstText(item?.question, '请进一步补充简历细节'),
    intent: firstText(item?.intent, '用于深入确认项目与能力边界'),
    focusSkills: safeArray(item?.focus_skills).slice(0, 4),
  }))

  if (normalized.length) {
    return normalized
  }

  return safeArray(fallbackQuestions).map((question) => ({
    question,
    intent: '由岗位匹配模块自动生成的补充追问',
    focusSkills: [],
  }))
}

function normalizeIntegration(integration, bestMatch) {
  const drillPlan = [
    { phase: 'Phase 1', content: firstText(integration?.drill_plan?.phase_1, '梳理项目背景与个人职责表达。') },
    { phase: 'Phase 2', content: firstText(integration?.drill_plan?.phase_2, '补齐关键知识点与高频追问。') },
    { phase: 'Phase 3', content: firstText(integration?.drill_plan?.phase_3, '做模拟问答并沉淀最终表达模版。') },
  ]

  return {
    targetRole: firstText(integration?.target_role, bestMatch?.roleKey, '--'),
    targetPosition: firstText(integration?.target_position, bestMatch?.positionName, '待确认岗位'),
    weakPoints: safeArray(integration?.weak_points).slice(0, 6),
    questionRecommendations: safeArray(integration?.question_recommendations).slice(0, 4),
    drillPlan,
  }
}

function validateResumeFile(file) {
  const allowed = ['pdf', 'doc', 'docx', 'txt', 'png', 'jpg', 'jpeg']
  const ext = String(file?.name || '').split('.').pop()?.toLowerCase()
  if (!allowed.includes(ext)) {
    return '请上传 PDF、DOC、DOCX、TXT 或图片格式的简历文件'
  }
  if ((file?.size || 0) > 20 * 1024 * 1024) {
    return '文件大小不能超过 20MB'
  }
  return ''
}

function normalizeLevel(level) {
  const text = firstText(level, 'basic').toLowerCase()
  if (text.includes('advanced') || text.includes('expert') || text.includes('senior')) return 'Advanced'
  if (text.includes('intermediate') || text.includes('medium')) return 'Intermediate'
  if (text.includes('熟练')) return 'Advanced'
  if (text.includes('掌握') || text.includes('中级')) return 'Intermediate'
  return 'Basic'
}

function formatPeriod(start, end, duration) {
  const parts = [firstText(start), firstText(end)].filter(Boolean)
  const range = parts.length ? parts.join(' — ') : firstText(duration, '时间待补充')
  return firstText(range, '时间待补充')
}

function formatDateTime(value) {
  if (!value) return '暂无记录'
  const parsed = dayjs(value)
  return parsed.isValid() ? parsed.format('YYYY.MM.DD HH:mm') : String(value)
}

function normalizeExternalLink(value) {
  const text = firstText(value)
  if (!text) return ''
  if (/^https?:\/\//i.test(text)) return text
  return `https://${text}`
}

function uniqueStrings(items) {
  return [...new Set(safeArray(items).map((item) => String(item).trim()).filter(Boolean))]
}

function safeArray(value) {
  return Array.isArray(value) ? value.filter((item) => item !== null && item !== undefined && item !== '') : []
}

function firstText(...values) {
  for (const value of values) {
    if (typeof value === 'string' && value.trim()) {
      return value.trim()
    }
  }
  return ''
}

function toNumber(value) {
  const num = Number(value || 0)
  if (!Number.isFinite(num)) return 0
  return Math.max(0, Math.min(100, Math.round(num)))
}
</script>

<style scoped>
.upload-shell {
  overflow: hidden;
  border-radius: 32px;
  border: 1px solid rgba(226, 232, 240, 0.92);
  background:
    radial-gradient(circle at top left, rgba(14, 165, 233, 0.08), transparent 28%),
    radial-gradient(circle at right, rgba(34, 197, 94, 0.08), transparent 22%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(248, 250, 252, 0.96));
  padding: 24px;
  box-shadow: 0 22px 54px rgba(15, 23, 42, 0.05);
}

.upload-dropzone {
  min-height: 280px;
  border-radius: 28px;
  border: 1px dashed rgba(148, 163, 184, 0.5);
  background: rgba(255, 255, 255, 0.68);
  padding: 28px;
  transition:
    border-color 180ms ease,
    transform 180ms ease,
    box-shadow 180ms ease;
}

.upload-dropzone-active {
  border-color: rgba(37, 99, 235, 0.6);
  box-shadow: inset 0 0 0 1px rgba(37, 99, 235, 0.2);
  transform: translateY(-2px);
}

.status-card {
  display: grid;
  gap: 18px;
  align-content: start;
  border-radius: 28px;
  border: 1px solid rgba(226, 232, 240, 0.9);
  background: rgba(248, 250, 252, 0.88);
  padding: 24px;
}

.loading-shell {
  display: grid;
  gap: 14px;
  border-radius: 32px;
  border: 1px solid rgba(226, 232, 240, 0.92);
  background: rgba(255, 255, 255, 0.9);
  padding: 28px;
}

.loading-line,
.loading-block {
  border-radius: 999px;
  background: linear-gradient(90deg, rgba(226, 232, 240, 0.8), rgba(241, 245, 249, 0.96), rgba(226, 232, 240, 0.8));
  background-size: 200% 100%;
  animation: shimmer 1.2s linear infinite;
}

.loading-line {
  height: 14px;
}

.loading-grid {
  display: grid;
  gap: 16px;
  margin-top: 10px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.loading-block {
  height: 180px;
  border-radius: 28px;
}

.empty-shell {
  display: grid;
  gap: 18px;
  border-radius: 32px;
  border: 1px dashed rgba(148, 163, 184, 0.42);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.95), rgba(248, 250, 252, 0.9));
  padding: 36px;
}

@keyframes shimmer {
  from {
    background-position: 200% 0;
  }
  to {
    background-position: -200% 0;
  }
}

@media (max-width: 767px) {
  .upload-shell,
  .loading-shell,
  .empty-shell {
    border-radius: 28px;
    padding: 20px;
  }

  .upload-dropzone,
  .status-card {
    border-radius: 24px;
    padding: 20px;
  }

  .loading-grid {
    grid-template-columns: 1fr;
  }
}
</style>
