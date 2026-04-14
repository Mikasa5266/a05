<template>
  <div class="space-y-6">
    <section class="resume-hero reveal-panel" style="animation-delay: 40ms">
      <div class="grid gap-6 xl:grid-cols-[minmax(0,1.35fr)_340px]">
        <div class="space-y-5">
          <div class="flex flex-wrap items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.26em] text-white/70">
            <span>Structured Extraction</span>
            <span class="h-1 w-1 rounded-full bg-white/35" />
            <span>{{ viewModel.recordMeta.parserMode }}</span>
            <span class="h-1 w-1 rounded-full bg-white/35" />
            <span>{{ viewModel.recordMeta.updatedAt }}</span>
          </div>

          <div class="space-y-3">
            <div class="flex flex-wrap items-center gap-3">
              <h1 class="text-3xl font-semibold tracking-tight text-white md:text-4xl">
                {{ viewModel.person.name }}
              </h1>
              <span class="inline-flex items-center rounded-full border border-white/18 bg-white/10 px-3 py-1 text-xs font-medium text-white/85">
                {{ viewModel.bestMatch.positionName }}
              </span>
            </div>

            <p class="max-w-3xl text-sm leading-7 text-slate-200 md:text-[15px]">
              {{ viewModel.summary }}
            </p>
          </div>

          <div class="flex flex-wrap gap-2">
            <span
              v-for="chip in viewModel.focusChips"
              :key="chip"
              class="inline-flex items-center rounded-full border border-white/14 bg-white/8 px-3 py-1.5 text-xs text-white/80"
            >
              {{ chip }}
            </span>
          </div>

          <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
            <div
              v-for="item in viewModel.contactCards"
              :key="item.label"
              class="rounded-2xl border border-white/12 bg-white/8 px-4 py-3"
            >
              <p class="text-[11px] uppercase tracking-[0.18em] text-white/55">{{ item.label }}</p>
              <p class="mt-2 text-sm font-medium text-white">
                <a
                  v-if="item.href"
                  :href="item.href"
                  target="_blank"
                  rel="noreferrer"
                  class="hover:text-cyan-200 transition-colors"
                >
                  {{ item.value }}
                </a>
                <span v-else>{{ item.value }}</span>
              </p>
            </div>
          </div>
        </div>

        <aside class="resume-hero-sidebar">
          <div class="flex items-start justify-between gap-4">
            <div>
              <p class="text-xs font-semibold uppercase tracking-[0.22em] text-sky-200/90">岗位匹配度</p>
              <h2 class="mt-2 text-2xl font-semibold text-white">{{ viewModel.bestMatch.positionName }}</h2>
              <p class="mt-2 text-sm leading-6 text-slate-200">
                {{ viewModel.bestMatch.analysis }}
              </p>
            </div>
            <div class="score-ring">
              <span class="text-3xl font-semibold text-white">{{ viewModel.bestMatch.score }}</span>
              <span class="text-[11px] uppercase tracking-[0.2em] text-white/60">Score</span>
            </div>
          </div>

          <div class="grid gap-3 sm:grid-cols-2">
            <article
              v-for="metric in viewModel.metrics"
              :key="metric.label"
              class="rounded-2xl border border-white/12 bg-black/14 px-4 py-4"
            >
              <p class="text-[11px] uppercase tracking-[0.18em] text-white/50">{{ metric.label }}</p>
              <p class="mt-2 text-2xl font-semibold text-white">{{ metric.value }}</p>
              <p class="mt-2 text-xs leading-5 text-white/65">{{ metric.detail }}</p>
            </article>
          </div>
        </aside>
      </div>
    </section>

    <div class="grid gap-7 2xl:grid-cols-[minmax(0,1.55fr)_minmax(360px,0.85fr)]">
      <div class="space-y-6">
        <section class="resume-panel panel-profile reveal-panel" style="animation-delay: 70ms">
          <div class="resume-section-heading">
            <div>
              <p class="resume-eyebrow">Quick Profile</p>
              <h2 class="resume-title">核心能力概览</h2>
            </div>
          </div>

          <div class="quick-skill-grid">
            <ResumeSkillCard
              v-for="card in quickOverviewCards"
              :key="card.title"
              :title="card.title"
              :subtitle="card.subtitle"
              :badge="card.badge"
              :tone="card.tone"
              :items="card.items"
              :empty-text="card.emptyText"
            />
          </div>
        </section>

        <section class="resume-panel panel-skill reveal-panel" style="animation-delay: 90ms">
          <div class="resume-section-heading">
            <div>
              <p class="resume-eyebrow">Skill Atlas</p>
              <h2 class="resume-title">技能图谱与岗位维度</h2>
            </div>
          </div>

          <div class="grid gap-6 xl:grid-cols-[minmax(0,1.08fr)_340px]">
            <div class="grid gap-5 lg:grid-cols-2">
              <article
                v-for="group in viewModel.skillGroups"
                :key="group.key"
                class="skill-group-card"
              >
                <div class="flex items-center justify-between gap-3">
                  <div>
                    <p class="text-sm font-semibold text-slate-900">{{ group.label }}</p>
                    <p class="mt-1 text-xs text-slate-500">{{ group.count }} 项明确证据</p>
                  </div>
                  <span
                    class="skill-group-summary rounded-full bg-white px-2.5 py-1 text-[11px] font-medium text-slate-500"
                    :title="group.summary"
                  >
                    {{ group.summary }}
                  </span>
                </div>

                <div class="mt-4 grid gap-2.5">
                  <div
                    v-for="skill in compactSkillItems(group.items)"
                    :key="`${group.key}-${skill.name}`"
                    class="skill-chip"
                    :class="skillToneClass(skill.level)"
                  >
                    <div class="min-w-0">
                      <p class="truncate text-sm font-medium" :title="skill.name">{{ skill.name }}</p>
                      <p
                        class="mt-1 line-clamp-2 text-[11px] leading-5 opacity-80"
                        :title="skill.evidence || skill.lastUsed || '未补充证据'"
                      >
                        {{ skill.evidence || skill.lastUsed || '未补充证据' }}
                      </p>
                    </div>
                    <span
                      class="shrink-0 rounded-full bg-white/70 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide"
                      :title="skill.level"
                    >
                      {{ skill.level }}
                    </span>
                  </div>
                </div>

                <p v-if="hiddenSkillCount(group.items) > 0" class="skill-group-more">
                  +{{ hiddenSkillCount(group.items) }} 项能力已折叠展示
                </p>
              </article>
            </div>

            <div class="rounded-[28px] border border-slate-200 bg-white p-5 shadow-[0_24px_60px_rgba(15,23,42,0.06)]">
              <div class="flex items-center justify-between gap-3">
                <div>
                  <p class="resume-eyebrow text-slate-500">Radar</p>
                  <h3 class="text-lg font-semibold text-slate-900">岗位能力雷达</h3>
                </div>
                <span class="rounded-full bg-sky-50 px-3 py-1 text-xs font-medium text-sky-700">
                  {{ viewModel.bestMatch.positionCode }}
                </span>
              </div>

              <div class="radar-shell mt-5">
                <Radar v-if="hasRadar" :data="radarData" :options="radarOptions" />
                <div v-else class="flex h-full items-center justify-center text-sm text-slate-400">
                  暂无足够的能力维度数据
                </div>
              </div>

              <div class="mt-5 space-y-3">
                <div
                  v-for="indicator in viewModel.radarIndicators"
                  :key="indicator.label"
                  class="space-y-1.5"
                >
                  <div class="flex items-center justify-between text-xs text-slate-500">
                    <span>{{ indicator.label }}</span>
                    <span>{{ indicator.value }}</span>
                  </div>
                  <div class="h-2 rounded-full bg-slate-100">
                    <div
                      class="h-full rounded-full bg-linear-to-r from-sky-500 to-indigo-500"
                      :style="{ width: `${indicator.value}%` }"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section class="resume-panel panel-project reveal-panel" style="animation-delay: 140ms">
          <div class="resume-section-heading">
            <div>
              <p class="resume-eyebrow">Project Review</p>
              <h2 class="resume-title">项目经验深度剖析</h2>
            </div>
          </div>

          <div class="space-y-4">
            <article
              v-for="(project, index) in viewModel.projects"
              :key="projectKey(project, index)"
              class="project-row"
            >
              <div class="project-meta">
                <p class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-400">{{ project.period }}</p>
                <h3 class="mt-2 text-xl font-semibold text-slate-900">{{ project.name }}</h3>
                <p class="mt-2 text-sm text-slate-500">{{ project.role }}</p>
              </div>

              <div class="space-y-5">
                <p
                  class="text-sm leading-7 text-slate-600"
                  :class="{ 'line-clamp-3': isProjectLong(project) && !isProjectExpanded(projectKey(project, index)) }"
                >
                  {{ project.summary || project.background || '暂无项目摘要' }}
                </p>

                <div class="grid gap-4 lg:grid-cols-2">
                  <div>
                    <p class="resume-subtitle">亮点拆解</p>
                    <ul class="space-y-2">
                      <li
                        v-for="item in visibleProjectItems(project.highlights, projectKey(project, index))"
                        :key="item"
                        class="list-row"
                      >
                        {{ item }}
                      </li>
                    </ul>
                  </div>

                  <div>
                    <p class="resume-subtitle">结果影响</p>
                    <ul class="space-y-2">
                      <li
                        v-for="item in visibleProjectItems(project.impact, projectKey(project, index))"
                        :key="item"
                        class="list-row"
                      >
                        {{ item }}
                      </li>
                    </ul>
                  </div>
                </div>

                <div class="flex flex-wrap gap-2">
                  <span
                    v-for="tech in project.techStack"
                    :key="tech"
                    class="rounded-full bg-slate-900 px-3 py-1 text-xs font-medium text-white"
                  >
                    {{ tech }}
                  </span>
                </div>

                <div v-if="isProjectLong(project)">
                  <button
                    type="button"
                    class="project-toggle"
                    @click="toggleProject(projectKey(project, index))"
                  >
                    {{ isProjectExpanded(projectKey(project, index)) ? '收起详情' : '展开详情' }}
                  </button>
                </div>
              </div>
            </article>
          </div>
        </section>

        <section class="resume-panel panel-timeline reveal-panel" style="animation-delay: 190ms">
          <div class="resume-section-heading">
            <div>
              <p class="resume-eyebrow">Timeline</p>
              <h2 class="resume-title">经历与教育概览</h2>
            </div>
          </div>

          <div class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_320px]">
            <div class="space-y-4">
              <article
                v-for="work in viewModel.workExperience"
                :key="`${work.company}-${work.role}-${work.period}`"
                class="timeline-row"
              >
                <div class="timeline-dot" />
                <div class="space-y-3">
                  <div class="flex flex-wrap items-center justify-between gap-3">
                    <div>
                      <h3 class="text-base font-semibold text-slate-900">{{ work.company }}</h3>
                      <p class="mt-1 text-sm text-slate-500">{{ work.role }}</p>
                    </div>
                    <span class="rounded-full bg-slate-100 px-3 py-1 text-xs font-medium text-slate-600">
                      {{ work.period }}
                    </span>
                  </div>
                  <p class="text-sm leading-7 text-slate-600">{{ work.summary }}</p>
                  <div class="grid gap-4 md:grid-cols-2">
                    <div>
                      <p class="resume-subtitle">职责</p>
                      <ul class="space-y-2">
                        <li v-for="item in work.responsibilities" :key="item" class="list-row">{{ item }}</li>
                      </ul>
                    </div>
                    <div>
                      <p class="resume-subtitle">成果</p>
                      <ul class="space-y-2">
                        <li v-for="item in work.achievements" :key="item" class="list-row">{{ item }}</li>
                      </ul>
                    </div>
                  </div>
                </div>
              </article>
            </div>

            <aside class="space-y-4">
              <template v-if="viewModel.education.length">
                <article
                  v-for="education in viewModel.education"
                  :key="`${education.school}-${education.degree}-${education.period}`"
                  class="rounded-3xl border border-slate-200 bg-slate-50 p-5"
                >
                  <p class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-400">{{ education.period }}</p>
                  <h3 class="mt-2 text-lg font-semibold text-slate-900">{{ education.school }}</h3>
                  <p class="mt-1 text-sm text-slate-500">{{ education.degree }} · {{ education.major }}</p>
                  <div class="mt-4 space-y-2 text-sm text-slate-600">
                    <p v-if="education.gpa">GPA：{{ education.gpa }}</p>
                    <p v-if="education.ranking">排名：{{ education.ranking }}</p>
                  </div>
                  <div class="mt-4 space-y-2">
                    <div
                      v-for="item in education.highlights"
                      :key="item"
                      class="list-row"
                    >
                      {{ item }}
                    </div>
                  </div>
                </article>
              </template>

              <article v-else class="education-empty-card">
                <p class="resume-subtitle">教育经历</p>
                <p class="mt-3 text-sm leading-6 text-amber-700">教育经历补充不充分，当前未识别到可展示内容。</p>
              </article>

              <article class="rounded-3xl border border-slate-200 bg-white p-5">
                <p class="resume-subtitle">核心亮点</p>
                <div class="mt-3 flex flex-wrap gap-2">
                  <span
                    v-for="item in viewModel.highlights"
                    :key="item"
                    class="rounded-full bg-emerald-50 px-3 py-1.5 text-xs font-medium text-emerald-700"
                  >
                    {{ item }}
                  </span>
                </div>
                <p class="resume-subtitle mt-6">关注点</p>
                <div class="mt-3 flex flex-wrap gap-2">
                  <span
                    v-for="item in viewModel.concerns"
                    :key="item"
                    class="rounded-full bg-amber-50 px-3 py-1.5 text-xs font-medium text-amber-700"
                  >
                    {{ item }}
                  </span>
                </div>
              </article>
            </aside>
          </div>
        </section>
      </div>

      <aside class="space-y-6">
        <section class="resume-panel panel-match reveal-panel" style="animation-delay: 110ms">
          <div class="resume-section-heading">
            <div>
              <p class="resume-eyebrow">Role Match</p>
              <h2 class="resume-title">岗位匹配建议</h2>
            </div>
          </div>

          <div class="space-y-4">
            <article
              v-for="match in viewModel.matchResults"
              :key="match.positionCode"
              class="rounded-3xl border border-slate-200 bg-slate-50 p-4"
            >
              <div class="flex items-center justify-between gap-3">
                <div>
                  <h3 class="text-base font-semibold text-slate-900">{{ match.positionName }}</h3>
                  <p class="mt-1 text-xs text-slate-500">{{ match.roleKey }}</p>
                </div>
                <span class="rounded-full bg-slate-900 px-3 py-1 text-xs font-semibold text-white">
                  {{ match.score }}
                </span>
              </div>

              <p class="mt-3 text-sm leading-6 text-slate-600">{{ match.analysis }}</p>

              <div class="mt-4 space-y-3">
                <div
                  v-for="item in match.breakdown"
                  :key="item.label"
                  class="space-y-1.5"
                >
                  <div class="flex items-center justify-between text-xs text-slate-500">
                    <span>{{ item.label }}</span>
                    <span>{{ item.value }}</span>
                  </div>
                  <div class="h-2 rounded-full bg-white">
                    <div class="h-full rounded-full bg-linear-to-r from-cyan-500 to-indigo-500" :style="{ width: `${item.value}%` }" />
                  </div>
                </div>
              </div>

              <div class="mt-4 space-y-3">
                <div>
                  <p class="resume-subtitle">命中技能</p>
                  <div class="mt-2 flex flex-wrap gap-2">
                    <span v-for="item in match.hitSkills" :key="item" class="chip chip-positive">{{ item }}</span>
                  </div>
                </div>
                <div>
                  <p class="resume-subtitle">待补能力</p>
                  <div class="mt-2 flex flex-wrap gap-2">
                    <span v-for="item in match.gapSkills" :key="item" class="chip chip-warning">{{ item }}</span>
                  </div>
                </div>
              </div>
            </article>
          </div>
        </section>

        <section class="resume-panel panel-action reveal-panel" style="animation-delay: 160ms">
          <div class="resume-section-heading">
            <div>
              <p class="resume-eyebrow">Action Items</p>
              <h2 class="resume-title">优化建议</h2>
            </div>
          </div>

          <div class="space-y-3">
            <article
              v-for="item in viewModel.optimization"
              :key="`${item.title}-${item.priority}`"
              class="rounded-3xl border border-slate-200 bg-white p-4"
            >
              <div class="flex items-start justify-between gap-3">
                <div>
                  <h3 class="text-sm font-semibold text-slate-900">{{ item.title }}</h3>
                  <p class="mt-2 text-sm leading-6 text-slate-600">{{ item.action }}</p>
                </div>
                <span class="rounded-full px-2.5 py-1 text-[11px] font-semibold" :class="priorityClass(item.priority)">
                  {{ item.priority }}
                </span>
              </div>
              <p class="mt-3 text-xs leading-5 text-slate-500">{{ item.rationale }}</p>
            </article>
          </div>
        </section>

        <section class="resume-panel panel-risk reveal-panel" style="animation-delay: 210ms">
          <div class="resume-section-heading">
            <div>
              <p class="resume-eyebrow">Risk Review</p>
              <h2 class="resume-title">风险提示</h2>
            </div>
          </div>

          <div class="space-y-3">
            <article
              v-for="risk in viewModel.risks"
              :key="`${risk.level}-${risk.item}`"
              class="rounded-3xl border p-4"
              :class="riskClass(risk.level)"
            >
              <div class="flex items-start justify-between gap-3">
                <div>
                  <h3 class="text-sm font-semibold">{{ risk.item }}</h3>
                  <p class="mt-2 text-sm leading-6 opacity-85">{{ risk.detail }}</p>
                </div>
                <span class="rounded-full bg-white/80 px-2.5 py-1 text-[11px] font-semibold uppercase tracking-wide">
                  {{ risk.level }}
                </span>
              </div>
              <div v-if="risk.evidence.length" class="mt-3 flex flex-wrap gap-2">
                <span v-for="item in risk.evidence" :key="item" class="chip chip-neutral">{{ item }}</span>
              </div>
            </article>
          </div>
        </section>

        <section class="resume-panel panel-interview reveal-panel" style="animation-delay: 260ms">
          <div class="resume-section-heading">
            <div>
              <p class="resume-eyebrow">Interview Prep</p>
              <h2 class="resume-title">面试准备与 Drill Plan</h2>
            </div>
          </div>

          <div class="space-y-5">
            <div>
              <p class="resume-subtitle">推荐追问</p>
              <div class="mt-3 space-y-3">
                <article
                  v-for="question in viewModel.interviewQuestions"
                  :key="question.question"
                  class="rounded-2xl border border-slate-200 bg-slate-50 p-4"
                >
                  <p class="text-sm font-medium leading-6 text-slate-900">{{ question.question }}</p>
                  <p class="mt-2 text-xs text-slate-500">{{ question.intent }}</p>
                  <div class="mt-3 flex flex-wrap gap-2">
                    <span v-for="item in question.focusSkills" :key="item" class="chip chip-neutral">{{ item }}</span>
                  </div>
                </article>
              </div>
            </div>

            <div>
              <p class="resume-subtitle">阶段计划</p>
              <div class="mt-3 space-y-3">
                <article
                  v-for="phase in viewModel.drillPlan"
                  :key="phase.phase"
                  class="rounded-2xl border border-slate-200 bg-white p-4"
                >
                  <p class="phase-label">{{ phase.phase }}</p>
                  <p class="mt-2 text-sm leading-6 text-slate-700">{{ phase.content }}</p>
                </article>
              </div>
            </div>

            <div class="rounded-3xl border border-dashed border-slate-300 bg-slate-50 p-4">
              <p class="resume-subtitle">同步到刷题计划</p>
              <div class="mt-3 flex flex-wrap gap-2">
                <span v-for="item in viewModel.weakPoints" :key="item" class="chip chip-warning">{{ item }}</span>
              </div>
              <div class="mt-4 space-y-2 text-sm text-slate-600">
                <p>目标岗位：{{ viewModel.integration.targetPosition }}</p>
                <p>题目建议：{{ viewModel.integration.questionRecommendations.join(' · ') }}</p>
              </div>
            </div>
          </div>
        </section>
      </aside>
    </div>

  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import {
  Chart as ChartJS,
  Filler,
  Legend,
  LineElement,
  PointElement,
  RadialLinearScale,
  Tooltip,
} from 'chart.js'
import { Radar } from 'vue-chartjs'
import ResumeSkillCard from './ResumeSkillCard.vue'

ChartJS.register(RadialLinearScale, PointElement, LineElement, Filler, Tooltip, Legend)

const props = defineProps({
  viewModel: {
    type: Object,
    required: true,
  },
})

const hasRadar = computed(() =>
  props.viewModel.radarIndicators.some((item) => Number(item.value || 0) > 0)
)

const radarData = computed(() => ({
  labels: props.viewModel.radarIndicators.map((item) => item.label),
  datasets: [
    {
      label: '岗位匹配维度',
      data: props.viewModel.radarIndicators.map((item) => item.value),
      backgroundColor: 'rgba(34, 197, 94, 0.14)',
      borderColor: '#2563eb',
      pointBackgroundColor: '#0f172a',
      pointBorderColor: '#ffffff',
      pointHoverBackgroundColor: '#ffffff',
      pointHoverBorderColor: '#2563eb',
      borderWidth: 2,
    },
  ],
}))

const radarOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  scales: {
    r: {
      suggestedMin: 0,
      suggestedMax: 100,
      grid: {
        color: '#e2e8f0',
      },
      angleLines: {
        color: '#e2e8f0',
      },
      pointLabels: {
        color: '#475569',
        font: {
          size: 11,
          weight: 600,
        },
      },
      ticks: {
        stepSize: 20,
        color: '#94a3b8',
        backdropColor: 'transparent',
      },
    },
  },
  plugins: {
    legend: {
      display: false,
    },
  },
}))

const quickOverviewCards = computed(() => {
  const skillGroups = Array.isArray(props.viewModel?.skillGroups)
    ? props.viewModel.skillGroups
    : []
  const programmingGroup = skillGroups.find((group) => group.key === 'programming_languages')
  const frameworkGroup = skillGroups.find((group) => group.key === 'frameworks')

  const mapSkillItems = (group) =>
    (Array.isArray(group?.items) ? group.items : []).slice(0, 5).map((skill) => ({
      name: skill.name || '未命名技能',
      meta: skill.level || 'Basic',
      description: skill.evidence || skill.lastUsed || '',
    }))

  return [
    {
      title: '基本信息',
      subtitle: '候选人与目标岗位快照',
      tone: 'profile',
      badge: props.viewModel?.bestMatch?.positionCode || '基础信息',
      items: [
        { name: '姓名', meta: props.viewModel?.person?.name || '--' },
        { name: '目标岗位', meta: props.viewModel?.bestMatch?.positionName || '--' },
        { name: '邮箱', meta: props.viewModel?.person?.email || '--' },
        { name: '手机号', meta: props.viewModel?.person?.phone || '--' },
      ],
      emptyText: '暂无基础信息',
    },
    {
      title: '编程语言',
      subtitle: '语言能力证据概览',
      tone: 'language',
      badge: `${programmingGroup?.count || 0} 项`,
      items: mapSkillItems(programmingGroup),
      emptyText: '暂无编程语言数据',
    },
    {
      title: '框架与中间件',
      subtitle: '工程框架与中间件栈',
      tone: 'framework',
      badge: `${frameworkGroup?.count || 0} 项`,
      items: mapSkillItems(frameworkGroup),
      emptyText: '暂无框架能力数据',
    },
  ]
})

const expandedProjectKeys = ref(new Set())

const projectKey = (project, index) => `${project?.name || 'project'}-${index}`

const isProjectLong = (project) => {
  const summaryLength = String(project?.summary || project?.background || '').length
  const highlightCount = Array.isArray(project?.highlights) ? project.highlights.length : 0
  const impactCount = Array.isArray(project?.impact) ? project.impact.length : 0
  return summaryLength > 160 || highlightCount > 3 || impactCount > 3
}

const isProjectExpanded = (key) => expandedProjectKeys.value.has(key)

const toggleProject = (key) => {
  const next = new Set(expandedProjectKeys.value)
  if (next.has(key)) {
    next.delete(key)
  } else {
    next.add(key)
  }
  expandedProjectKeys.value = next
}

const visibleProjectItems = (items, key) => {
  const list = Array.isArray(items) ? items.filter(Boolean) : []
  if (isProjectExpanded(key)) {
    return list
  }
  return list.slice(0, 3)
}

const compactSkillItems = (items) => {
  const list = Array.isArray(items) ? items.filter(Boolean) : []
  return list.slice(0, 4)
}

const hiddenSkillCount = (items) => {
  const list = Array.isArray(items) ? items.filter(Boolean) : []
  return Math.max(0, list.length - 4)
}

const priorityClass = (priority = '') => {
  const normalized = String(priority || '').toLowerCase()
  if (normalized.includes('high') || normalized.includes('高')) {
    return 'bg-rose-50 text-rose-700'
  }
  if (normalized.includes('medium') || normalized.includes('中')) {
    return 'bg-amber-50 text-amber-700'
  }
  return 'bg-emerald-50 text-emerald-700'
}

const riskClass = (level = '') => {
  const normalized = String(level || '').toLowerCase()
  if (normalized.includes('high') || normalized.includes('高')) {
    return 'border-rose-200 bg-rose-50 text-rose-800'
  }
  if (normalized.includes('medium') || normalized.includes('中')) {
    return 'border-amber-200 bg-amber-50 text-amber-800'
  }
  return 'border-sky-200 bg-sky-50 text-sky-800'
}

const skillToneClass = (level = '') => {
  const normalized = String(level || '').toLowerCase()
  if (normalized.includes('advanced') || normalized.includes('expert') || normalized.includes('熟练')) {
    return 'bg-emerald-50 text-emerald-900'
  }
  if (normalized.includes('intermediate') || normalized.includes('medium') || normalized.includes('掌握')) {
    return 'bg-sky-50 text-sky-900'
  }
  return 'bg-amber-50 text-amber-900'
}
</script>

<style scoped>
.resume-hero {
  position: relative;
  overflow: hidden;
  border-radius: 36px;
  padding: 32px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  background:
    radial-gradient(circle at top left, rgba(34, 197, 94, 0.2), transparent 28%),
    radial-gradient(circle at 80% 20%, rgba(56, 189, 248, 0.25), transparent 30%),
    linear-gradient(135deg, #0f172a 0%, #111827 54%, #172554 100%);
  box-shadow: 0 32px 72px rgba(15, 23, 42, 0.18);
}

.resume-hero-sidebar {
  display: grid;
  gap: 18px;
  align-content: space-between;
  border-radius: 28px;
  padding: 22px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.12), rgba(15, 23, 42, 0.12));
  backdrop-filter: blur(10px);
}

.score-ring {
  display: inline-flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-width: 108px;
  min-height: 108px;
  border-radius: 999px;
  border: 1px solid rgba(255, 255, 255, 0.16);
  background:
    radial-gradient(circle at top, rgba(125, 211, 252, 0.36), transparent 52%),
    rgba(255, 255, 255, 0.08);
}

.resume-panel {
  --panel-accent: #2563eb;
  --panel-tint: rgba(37, 99, 235, 0.14);
  --panel-heading-font: 'HarmonyOS Sans SC', 'Source Han Sans SC', 'PingFang SC', 'Microsoft YaHei', sans-serif;
  --panel-body-font: 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif;
  --panel-meta-font: 'JetBrains Mono', 'Cascadia Code', 'Consolas', monospace;
  border-radius: 32px;
  border: 1px solid rgba(226, 232, 240, 0.92);
  border-top: 3px solid var(--panel-accent);
  background:
    radial-gradient(circle at top right, var(--panel-tint), transparent 42%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.97), rgba(255, 255, 255, 0.93));
  padding: 28px;
  box-shadow: 0 22px 54px rgba(15, 23, 42, 0.06);
  font-family: var(--panel-body-font);
}

.panel-profile {
  --panel-accent: #0891b2;
  --panel-tint: rgba(8, 145, 178, 0.16);
  --panel-heading-font: 'HarmonyOS Sans SC', 'Source Han Sans SC', 'PingFang SC', 'Microsoft YaHei', sans-serif;
}

.panel-skill {
  --panel-accent: #2563eb;
  --panel-tint: rgba(37, 99, 235, 0.16);
  --panel-heading-font: 'DIN Alternate', 'HarmonyOS Sans SC', 'PingFang SC', 'Microsoft YaHei', sans-serif;
}

.panel-project {
  --panel-accent: #059669;
  --panel-tint: rgba(5, 150, 105, 0.15);
  --panel-heading-font: 'Source Han Serif SC', 'Noto Serif SC', 'Songti SC', serif;
}

.panel-timeline {
  --panel-accent: #0284c7;
  --panel-tint: rgba(2, 132, 199, 0.16);
  --panel-heading-font: 'Source Han Serif SC', 'Noto Serif SC', 'Songti SC', serif;
}

.panel-match {
  --panel-accent: #4f46e5;
  --panel-tint: rgba(79, 70, 229, 0.15);
}

.panel-action {
  --panel-accent: #ea580c;
  --panel-tint: rgba(234, 88, 12, 0.14);
}

.panel-risk {
  --panel-accent: #b45309;
  --panel-tint: rgba(180, 83, 9, 0.14);
}

.panel-interview {
  --panel-accent: #0f766e;
  --panel-tint: rgba(15, 118, 110, 0.14);
}

.resume-section-heading {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 20px;
  margin-bottom: 26px;
}

.resume-eyebrow {
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.22em;
  text-transform: uppercase;
  color: #64748b;
  font-family: var(--panel-meta-font);
}

.resume-title {
  margin-top: 8px;
  font-size: 24px;
  line-height: 1.15;
  font-weight: 600;
  color: #0f172a;
  font-family: var(--panel-heading-font);
}

.resume-caption {
  max-width: 420px;
  font-size: 13px;
  line-height: 1.7;
  color: #64748b;
}

.resume-subtitle {
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: #64748b;
  font-family: var(--panel-meta-font);
}

.radar-shell {
  height: 320px;
}

.quick-skill-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 18px;
}

.skill-group-card {
  min-width: 0;
  overflow: hidden;
  border-radius: 24px;
  border: 1px solid rgba(226, 232, 240, 0.86);
  background: linear-gradient(180deg, rgba(248, 250, 252, 0.9), rgba(255, 255, 255, 0.96));
  padding: 18px;
}

.skill-group-more {
  margin-top: 10px;
  font-size: 12px;
  color: #64748b;
}

.skill-group-summary {
  max-width: 56%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.skill-chip {
  box-sizing: border-box;
  display: flex;
  width: 100%;
  min-width: 0;
  max-width: 100%;
  justify-content: space-between;
  align-items: flex-start;
  gap: 14px;
  border-radius: 24px;
  border: 1px solid rgba(226, 232, 240, 0.72);
  padding: 12px 14px 12px 16px;
}

.skill-chip .min-w-0 {
  flex: 1 1 auto;
  min-width: 0;
}

.education-empty-card {
  border-radius: 24px;
  border: 1px dashed rgba(245, 158, 11, 0.5);
  background: #fffbeb;
  padding: 16px;
}

.project-row {
  display: grid;
  gap: 24px;
  border-radius: 28px;
  border: 1px solid rgba(226, 232, 240, 0.9);
  background: linear-gradient(180deg, #ffffff 0%, #f8fafc 100%);
  padding: 26px;
}

.project-meta {
  min-width: 0;
}

.project-toggle {
  border: 1px solid #cbd5e1;
  background: #ffffff;
  color: #334155;
  border-radius: 12px;
  padding: 6px 10px;
  font-size: 12px;
  font-weight: 600;
  transition: all 180ms ease;
}

.project-toggle:hover {
  border-color: #94a3b8;
  color: #0f172a;
}

.timeline-row {
  position: relative;
  display: grid;
  grid-template-columns: 18px minmax(0, 1fr);
  gap: 18px;
  border-radius: 28px;
  border: 1px solid rgba(226, 232, 240, 0.85);
  background: #ffffff;
  padding: 24px;
}

.phase-label {
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.1em;
  color: #64748b;
  font-family: var(--panel-meta-font);
}

.timeline-dot {
  width: 10px;
  height: 10px;
  margin-top: 8px;
  border-radius: 999px;
  background: linear-gradient(135deg, #0ea5e9, #2563eb);
  box-shadow: 0 0 0 6px rgba(14, 165, 233, 0.08);
}

.list-row {
  position: relative;
  padding-left: 14px;
  font-size: 13px;
  line-height: 1.65;
  color: #475569;
}

.list-row::before {
  content: '';
  position: absolute;
  left: 0;
  top: 9px;
  width: 6px;
  height: 6px;
  border-radius: 999px;
  background: #22c55e;
}

.chip {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  padding: 6px 10px;
  font-size: 11px;
  font-weight: 600;
}

.chip-positive {
  background: #ecfdf5;
  color: #047857;
}

.chip-warning {
  background: #fff7ed;
  color: #c2410c;
}

.chip-neutral {
  background: #eef2ff;
  color: #4338ca;
}

.reveal-panel {
  opacity: 0;
  animation: panelRise 0.62s ease forwards;
}

@keyframes panelRise {
  from {
    opacity: 0;
    transform: translateY(18px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (max-width: 1279px) {
  .resume-hero,
  .resume-panel {
    padding: 24px;
  }

  .quick-skill-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 767px) {
  .resume-hero,
  .resume-panel {
    border-radius: 28px;
    padding: 20px;
  }

  .resume-title {
    font-size: 20px;
  }

  .resume-section-heading {
    flex-direction: column;
    align-items: flex-start;
    margin-bottom: 18px;
  }

  .resume-caption {
    max-width: none;
  }

  .radar-shell {
    height: 280px;
  }

  .project-row {
    padding: 20px;
  }

  .quick-skill-grid {
    grid-template-columns: 1fr;
  }
}
</style>
