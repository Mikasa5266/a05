<template>
  <div class="space-y-8">
    <header>
      <h1 class="text-3xl font-bold tracking-tight text-zinc-900">面试历史</h1>
      <p class="text-zinc-500 mt-2">查看您过去的面试记录与详细报告</p>
    </header>

    <div class="bg-white rounded-3xl shadow-sm border border-zinc-100 overflow-hidden">
      <div v-if="loading" class="p-12 text-center text-zinc-500">
        加载中...
      </div>
      
      <div v-else-if="records.length === 0" class="p-12 text-center flex flex-col items-center">
        <div class="h-16 w-16 bg-zinc-50 rounded-full flex items-center justify-center mb-4">
          <FileText class="h-8 w-8 text-zinc-400" />
        </div>
        <h3 class="text-lg font-medium text-zinc-900">暂无面试记录</h3>
        <p class="text-zinc-500 mt-1">开始您的第一次模拟面试吧！</p>
        <router-link 
          to="/interview"
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
          <tr v-for="record in records" :key="record.interview_id" class="hover:bg-zinc-50/50 transition-colors">
            <td class="px-6 py-4 text-zinc-600">{{ formatDate(record.created_at) }}</td>
            <td class="px-6 py-4 font-medium text-zinc-900">{{ record.position }}</td>
            <td class="px-6 py-4">
              <span class="px-2 py-1 rounded-lg text-xs font-medium" 
                :class="{
                  'bg-emerald-50 text-emerald-700': record.difficulty === 'Junior',
                  'bg-amber-50 text-amber-700': record.difficulty === 'Middle' || record.difficulty === 'Mid',
                  'bg-rose-50 text-rose-700': record.difficulty === 'Senior'
                }"
              >
                {{ record.difficulty }}
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
                :class="record.report_id ? 'text-emerald-600' : (record.status === 'completed' ? 'text-amber-600' : 'text-zinc-500')">
                {{ record.report_id ? '报告已生成' : (record.status === 'completed' ? '待生成报告' : '进行中') }}
              </span>
            </td>
            <td class="px-6 py-4 text-right">
              <button 
                @click="viewReport(record)"
                class="text-indigo-600 hover:text-indigo-700 font-medium flex items-center gap-1 justify-end ml-auto"
              >
                {{ record.report_id ? '查看报告' : '生成并查看' }}
                <ChevronRight class="h-4 w-4" />
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getReports, generateReport } from '../api/report'
import { getInterviews } from '../api/interview'
import { FileText, ChevronRight } from 'lucide-vue-next'
import dayjs from 'dayjs'

const router = useRouter()
const records = ref([])
const loading = ref(true)

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
      return {
        interview_id: interview.id,
        position: interview.position,
        difficulty: interview.difficulty,
        status: interview.status,
        created_at: interview.created_at,
        average_score: report?.average_score ?? null,
        report_id: report?.id ?? interview.report?.id ?? null
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
        report_id: report.id
      }))
    }
  } catch (error) {
    console.error('Failed to fetch reports:', error)
  } finally {
    loading.value = false
  }
}

const viewReport = async (record) => {
  if (record.report_id) {
    router.push(`/report/${record.report_id}`)
    return
  }

  if (record.status !== 'completed') {
    return
  }

  try {
    const res = await generateReport({ interview_id: record.interview_id })
    const reportId = res?.report?.id
    if (reportId) {
      router.push(`/report/${reportId}`)
    }
  } catch (error) {
    console.error('Failed to generate report:', error)
  }
}

const formatDate = (date) => {
  return dayjs(date).format('YYYY-MM-DD HH:mm')
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
