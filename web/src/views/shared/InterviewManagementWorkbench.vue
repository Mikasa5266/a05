<template>
  <div class="space-y-6">
    <header class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-2xl md:text-3xl font-bold text-zinc-900">{{ titleText }}</h1>
        <p class="text-sm text-zinc-500 mt-1">邀请处理、面试进行中监控与历史评价统一视图</p>
      </div>
      <button
        class="px-4 py-2 rounded-xl border border-zinc-200 bg-white text-zinc-700 text-sm font-medium hover:bg-zinc-50 transition-colors disabled:opacity-60"
        :disabled="store.loading"
        @click="refreshWorkbench"
      >
        刷新数据
      </button>
    </header>

    <section class="bg-white border border-zinc-100 rounded-2xl p-2 shadow-sm">
      <div class="flex flex-wrap gap-2">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          class="px-3.5 py-2 rounded-xl text-sm font-medium transition-colors"
          :class="store.activeTab === tab.key
            ? 'bg-indigo-50 text-indigo-700 border border-indigo-100'
            : 'text-zinc-600 hover:bg-zinc-50 border border-transparent'"
          @click="store.setActiveTab(tab.key)"
        >
          {{ tab.label }}
          <span class="ml-1.5 text-xs rounded-full px-1.5 py-0.5"
            :class="store.activeTab === tab.key ? 'bg-indigo-100 text-indigo-700' : 'bg-zinc-100 text-zinc-600'"
          >
            {{ tab.count }}
          </span>
        </button>
      </div>
    </section>

    <section class="bg-white border border-zinc-100 rounded-3xl p-5 md:p-6 shadow-sm min-h-[360px]">
      <div v-if="store.loading" class="h-56 flex items-center justify-center text-zinc-500 text-sm">
        正在加载面试工作台数据...
      </div>

      <div v-else-if="store.currentList.length === 0" class="h-56 flex items-center justify-center text-zinc-400 text-sm">
        当前状态暂无记录
      </div>

      <div v-else class="space-y-4">
        <article
          v-for="item in store.currentList"
          :key="item.id"
          class="rounded-2xl border border-zinc-100 p-4 md:p-5 hover:shadow-sm transition-shadow"
        >
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div>
              <div class="flex items-center flex-wrap gap-2">
                <h3 class="text-base font-semibold text-zinc-900">
                  {{ item.student_name || `学生#${item.student_id}` }}
                </h3>
                <span class="px-2 py-1 rounded-full text-xs font-medium" :class="statusClass(item.status)">
                  {{ statusLabel(item.status) }}
                </span>
              </div>
              <p class="text-sm text-zinc-600 mt-1">
                {{ item.position || '未设置岗位' }} · {{ difficultyLabel(item.difficulty) }} · {{ modeLabel(item.mode) }}
              </p>
              <p class="text-xs text-zinc-500 mt-1">
                邀请码：{{ item.invitation_code || '-' }}
              </p>
              <p class="text-xs text-zinc-400 mt-1">
                创建时间：{{ formatDateTime(item.created_at) }}
              </p>
            </div>

            <div class="text-right min-w-[180px]">
              <p v-if="item.status === 'in_progress'" class="text-xs text-zinc-500">剩余倒计时</p>
              <p v-if="item.status === 'in_progress'" class="text-xl font-semibold text-indigo-700 mt-1">
                {{ formatCountdown(item) }}
              </p>
              <p v-else class="text-xs text-zinc-500">
                更新时间：{{ formatDateTime(item.updated_at) }}
              </p>
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-3 mt-4">
            <div class="rounded-xl bg-zinc-50 px-3 py-2">
              <p class="text-xs text-zinc-400">拟定时间</p>
              <p class="text-sm text-zinc-700 mt-1">{{ formatDateTime(item.scheduled_at) }}</p>
            </div>
            <div class="rounded-xl bg-zinc-50 px-3 py-2">
              <p class="text-xs text-zinc-400">面试状态</p>
              <p class="text-sm text-zinc-700 mt-1">{{ interviewStatusLabel(item.interview_status) }}</p>
            </div>
          </div>

          <div v-if="item.status === 'completed'" class="mt-4 rounded-xl border border-emerald-100 bg-emerald-50/50 p-3">
            <p class="text-xs text-emerald-700">评价记录</p>
            <p class="text-sm text-zinc-700 mt-1">评分：{{ item.human_score ?? '--' }}</p>
            <p class="text-sm text-zinc-600 mt-1">{{ item.human_feedback || '暂无评价内容' }}</p>
          </div>

          <p v-if="item.notes" class="text-xs text-zinc-500 mt-3">备注：{{ item.notes }}</p>

          <div class="mt-4 flex flex-wrap gap-2">
            <button
              v-if="item.status === 'pending'"
              class="px-3 py-1.5 rounded-lg bg-emerald-50 text-emerald-700 text-xs font-semibold hover:bg-emerald-100 disabled:opacity-60"
              :disabled="store.actionLoadingId === item.id"
              @click="respondInvitation(item.id, 'accept')"
            >
              接受邀请
            </button>
            <button
              v-if="item.status === 'pending'"
              class="px-3 py-1.5 rounded-lg bg-rose-50 text-rose-700 text-xs font-semibold hover:bg-rose-100 disabled:opacity-60"
              :disabled="store.actionLoadingId === item.id"
              @click="respondInvitation(item.id, 'reject')"
            >
              拒绝邀请
            </button>

            <button
              v-if="item.can_join"
              class="ml-auto px-3.5 py-1.5 rounded-lg border border-indigo-200 bg-indigo-50 text-indigo-700 text-xs font-semibold hover:bg-indigo-100"
              @click="enterLiveRoom(item)"
            >
              进入面试房间
            </button>
          </div>
        </article>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useInterviewWorkbenchStore } from '../../stores/interviewWorkbench'

const props = defineProps({
  portal: {
    type: String,
    required: true
  }
})

const router = useRouter()
const store = useInterviewWorkbenchStore()
const nowTick = ref(Date.now())
let countdownTimer = null
let refreshTimer = null

const routePrefix = computed(() => (props.portal === 'university' ? '/university' : '/enterprise'))
const titleText = computed(() => (props.portal === 'university' ? '高校端面试管理工作台' : '企业端面试管理工作台'))

const tabs = computed(() => ([
  { key: 'invite_list', label: '邀请列表', count: store.summary.invite_count },
  { key: 'pending', label: '待处理', count: store.summary.pending_count },
  { key: 'processed', label: '已处理', count: store.summary.processed_count },
  { key: 'in_progress', label: '正在面试', count: store.summary.in_progress_count },
  { key: 'history', label: '面试历史', count: store.summary.history_count }
]))

const statusLabel = (status) => {
  const map = {
    pending: '待处理',
    accepted: '已接受',
    rejected: '已拒绝',
    in_progress: '进行中',
    completed: '已完成',
    cancelled: '已取消'
  }
  return map[status] || status || '未知'
}

const statusClass = (status) => {
  if (status === 'pending') return 'bg-amber-50 text-amber-700'
  if (status === 'accepted') return 'bg-emerald-50 text-emerald-700'
  if (status === 'rejected') return 'bg-rose-50 text-rose-700'
  if (status === 'in_progress') return 'bg-indigo-50 text-indigo-700'
  if (status === 'completed') return 'bg-zinc-100 text-zinc-700'
  return 'bg-zinc-100 text-zinc-600'
}

const modeLabel = (mode) => {
  const map = {
    technical: '技术面',
    hr: 'HR 面',
    comprehensive: '综合面',
    blindbox: '盲盒面'
  }
  return map[mode] || mode || '未设置'
}

const difficultyLabel = (difficulty) => {
  const map = {
    campus_intern: '校招实习',
    campus_graduate: '校招全职',
    social_junior: '社招初级'
  }
  return map[difficulty] || difficulty || '未设置'
}

const interviewStatusLabel = (status) => {
  const map = {
    pending: '待进入',
    in_progress: '进行中',
    completed: '已完成'
  }
  return map[status] || status || '--'
}

const formatDateTime = (value) => {
  if (!value) return '待确认'
  const dt = new Date(value)
  if (Number.isNaN(dt.getTime())) return '待确认'
  return dt.toLocaleString('zh-CN', { hour12: false })
}

const formatCountdown = (item) => {
  const base = Math.max(0, Number(item?.remaining_seconds || 0))
  const elapsed = Math.max(0, Math.floor((nowTick.value - store.fetchedAt) / 1000))
  const remain = Math.max(0, base - elapsed)
  const mm = String(Math.floor(remain / 60)).padStart(2, '0')
  const ss = String(remain % 60).padStart(2, '0')
  return `${mm}:${ss}`
}

const refreshWorkbench = async () => {
  try {
    await store.fetchWorkbench()
  } catch (error) {
    ElMessage.error(error?.response?.data?.error || '工作台数据加载失败')
  }
}

const respondInvitation = async (invitationId, action) => {
  try {
    await store.respondInvitation(invitationId, action)
    ElMessage.success(action === 'accept' ? '已接受邀请' : '已拒绝邀请')
  } catch (error) {
    ElMessage.error(error?.response?.data?.error || '处理邀请失败')
  }
}

const enterLiveRoom = (item) => {
  router.push({
    path: `${routePrefix.value}/live-interview`,
    query: {
      invitation_id: String(item.id),
      invitation_code: String(item.invitation_code || '')
    }
  })
}

onMounted(async () => {
  await refreshWorkbench()

  countdownTimer = window.setInterval(() => {
    nowTick.value = Date.now()
  }, 1000)

  refreshTimer = window.setInterval(() => {
    refreshWorkbench()
  }, 15000)
})

onBeforeUnmount(() => {
  if (countdownTimer) {
    window.clearInterval(countdownTimer)
    countdownTimer = null
  }
  if (refreshTimer) {
    window.clearInterval(refreshTimer)
    refreshTimer = null
  }
})
</script>
