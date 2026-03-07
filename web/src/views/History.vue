<template>
  <div class="space-y-8">
    <header>
      <h1 class="text-3xl font-bold tracking-tight text-zinc-900">面试历史</h1>
      <p class="text-zinc-500 mt-2">查看您过去的面试记录与详细报告</p>
    </header>

    <div class="bg-white rounded-3xl shadow-sm border border-zinc-100 overflow-hidden">
      <div class="px-6 pt-5 pb-3 border-b border-zinc-100 bg-zinc-50/50">
        <div class="inline-flex p-1 rounded-xl bg-zinc-100 gap-1">
          <button
            v-for="item in filterOptions"
            :key="item.key"
            @click="activeFilter = item.key"
            class="px-3 py-1.5 text-xs font-medium rounded-lg transition-colors"
            :class="activeFilter === item.key ? 'bg-white text-zinc-900 shadow-sm' : 'text-zinc-500 hover:text-zinc-700'"
          >
            {{ item.label }}
          </button>
        </div>
      </div>

      <div v-if="loading" class="p-12 text-center text-zinc-500">
        加载中...
      </div>
      
      <div v-else-if="filteredRecords.length === 0" class="p-12 text-center flex flex-col items-center">
        <div class="h-16 w-16 bg-zinc-50 rounded-full flex items-center justify-center mb-4">
          <FileText class="h-8 w-8 text-zinc-400" />
        </div>
        <h3 class="text-lg font-medium text-zinc-900">当前分类暂无记录</h3>
        <p class="text-zinc-500 mt-1">开始一次面试后，这里会显示对应历史。</p>
        <router-link 
          to="/student/interview"
          class="mt-6 px-6 py-2 bg-indigo-600 text-white rounded-xl text-sm font-medium hover:bg-indigo-700 transition-colors"
        >
          去面试
        </router-link>
      </div>

      <table v-else class="w-full text-left text-sm">
        <thead class="bg-zinc-50 border-b border-zinc-100">
          <tr>
            <th class="px-6 py-4 font-medium text-zinc-500">日期</th>
            <th class="px-6 py-4 font-medium text-zinc-500">岗位</th>
            <th class="px-6 py-4 font-medium text-zinc-500">难度</th>
            <th class="px-6 py-4 font-medium text-zinc-500">综合得分</th>
            <th class="px-6 py-4 font-medium text-zinc-500">状态</th>
            <th class="px-6 py-4 font-medium text-zinc-500 text-right">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-zinc-100">
          <tr v-for="record in filteredRecords" :key="record.interview_id" class="hover:bg-zinc-50/50 transition-colors">
            <td class="px-6 py-4 text-zinc-600">{{ formatDate(record.created_at) }}</td>
            <td class="px-6 py-4 font-medium text-zinc-900">{{ record.position }}</td>
            <td class="px-6 py-4">
              <span class="px-2 py-1 rounded-lg text-xs font-medium" 
                :class="difficultyBadgeClass(record.difficulty)"
              >
                {{ formatDifficulty(record.difficulty) }}
              </span>
            </td>
            <td class="px-6 py-4">
              <span v-if="record.average_score !== null" class="font-bold" :class="getScoreColor(record.average_score)">
                {{ record.average_score }}
              </span>
              <span v-else class="text-zinc-400">--</span>
            </td>
            <td class="px-6 py-4">
              <span class="text-xs"
                :class="record.is_successful ? (record.report_id ? 'text-emerald-600' : 'text-amber-600') : 'text-zinc-500'">
                {{ record.is_successful ? (record.report_id ? '报告已生成' : '待生成报告') : '面试中断' }}
              </span>
            </td>
            <td class="px-6 py-4 text-right">
              <button 
                v-if="record.is_successful"
                @click="viewReport(record)"
                class="text-indigo-600 hover:text-indigo-700 font-medium flex items-center gap-1 justify-end ml-auto"
              >
                {{ record.report_id ? '查看报告' : '生成并查看' }}
                <ChevronRight class="h-4 w-4" />
              </button>
              <span v-else class="text-zinc-400 text-xs">不可操作</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getReports, generateReport } from '../api/report'
import { getInterviews } from '../api/interview'
import { FileText, ChevronRight } from 'lucide-vue-next'
import dayjs from 'dayjs'

const router = useRouter()
const records = ref([])
const loading = ref(true)
const activeFilter = ref('all')

const filterOptions = [
  { key: 'all', label: '全部' },
  { key: 'completed', label: '面试完成' },
  { key: 'interrupted', label: '面试中断' }
]

const isSuccessfulInterview = (interview, report) => {
  // If a report already exists, this interview has been finalized in business terms.
  if (report?.id || interview?.report?.id) {
    return true
  }

  if (!interview || interview.status !== 'completed') return false

  const plannedCount = Number(interview.total_question_target) || 0
  const arrangedCount = Array.isArray(interview.questions) ? interview.questions.length : 0
  const completedCount = Number(interview.current_index) || 0

  const target = Math.max(plannedCount, arrangedCount)
  if (target <= 0) {
    // Fallback: completed interview without enough metadata is treated as completed.
    return true
  }

  return completedCount >= target
}

const isInterruptedInterview = (interview) => !isSuccessfulInterview(interview)

const filteredRecords = computed(() => {
  if (activeFilter.value === 'completed') {
    return records.value.filter((r) => r.is_successful)
  }
  if (activeFilter.value === 'interrupted') {
    return records.value.filter((r) => r.is_interrupted)
  }
  return records.value
})

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

const viewReport = async (record) => {
  if (!record.is_successful) return

  if (record.report_id) {
    router.push(`/student/report/${record.report_id}`)
    return
  }

  if (record.status !== 'completed') {
    return
  }

  try {
    const res = await generateReport({ interview_id: record.interview_id })
    const reportId = res?.report?.id
    if (reportId) {
      router.push(`/student/report/${reportId}`)
    }
  } catch (error) {
    console.error('Failed to generate report:', error)
  }
}

const formatDate = (date) => {
  return dayjs(date).format('YYYY-MM-DD HH:mm')
}

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
    return 'bg-rose-50 text-rose-700'
  }
  if (difficulty === 'campus_graduate' || difficulty === 'Middle' || difficulty === 'Mid') {
    return 'bg-amber-50 text-amber-700'
  }
  return 'bg-emerald-50 text-emerald-700'
}

const getScoreColor = (score) => {
  if (score >= 80) return 'text-emerald-600'
  if (score >= 60) return 'text-amber-600'
  return 'text-rose-600'
}

onMounted(() => {
  fetchReports()
})
</script>
