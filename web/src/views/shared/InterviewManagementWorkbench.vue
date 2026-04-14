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

    <WorkbenchSidebarTabs
      :tabs="tabs"
      :active-tab="store.activeTab"
      @select="store.setActiveTab"
    />

    <section class="bg-white border border-zinc-100 rounded-3xl p-5 md:p-6 shadow-sm min-h-90">
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
          <WorkbenchStatusDisplay
            :item="item"
            :status-label="statusLabel"
            :status-class="statusClass"
            :difficulty-label="difficultyLabel"
            :mode-label="modeLabel"
            :interview-status-label="interviewStatusLabel"
            :format-date-time="formatDateTime"
            :format-countdown="formatCountdown"
          />

          <WorkbenchActionPanel
            :item="item"
            :action-loading-id="store.actionLoadingId"
            @accept="respondInvitation($event, 'accept')"
            @reject="respondInvitation($event, 'reject')"
            @enter-room="enterLiveRoom"
          />
        </article>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import WorkbenchActionPanel from '../../components/workbench/WorkbenchActionPanel.vue'
import WorkbenchSidebarTabs from '../../components/workbench/WorkbenchSidebarTabs.vue'
import WorkbenchStatusDisplay from '../../components/workbench/WorkbenchStatusDisplay.vue'
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

const isGroupInvitation = (item) => {
  const scenarioType = String(item?.scenario_type || '').trim().toLowerCase()
  const targetParticipants = Number(item?.target_participants || 0)
  return scenarioType === 'group' || targetParticipants > 2
}

const enterLiveRoom = (item) => {
  const invitationId = String(item?.id || '').trim()
  if (!invitationId) {
    ElMessage.warning('邀请信息无效，无法进入房间')
    return
  }

  const isGroup = isGroupInvitation(item)
  const query = {}
  const invitationCode = String(item?.invitation_code || '').trim()
  if (invitationCode) {
    query.invitation_code = invitationCode
  }
  if (isGroup) {
    query.group_mode = '1'
  }

  router.push({
    path: isGroup
      ? `${routePrefix.value}/live-interview/group/${encodeURIComponent(invitationId)}`
      : `${routePrefix.value}/live-interview/1v1/${encodeURIComponent(invitationId)}`,
    query
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
