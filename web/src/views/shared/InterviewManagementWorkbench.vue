<template>
  <div class="space-y-6">
    <header class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-2xl md:text-3xl font-bold text-zinc-900">{{ titleText }}</h1>
        <p class="text-sm text-zinc-500 mt-1">{{ subtitleText }}</p>
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

      <div v-else-if="currentList.length === 0" class="h-56 flex items-center justify-center text-zinc-400 text-sm">
        当前状态暂无记录
      </div>

      <div v-else class="space-y-4">
        <article
          v-for="item in currentList"
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
            @delete-invitation="deleteInvitation"
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
  },
  scenario: {
    type: String,
    default: 'all'
  }
})

const router = useRouter()
const store = useInterviewWorkbenchStore()
const nowTick = ref(Date.now())
let countdownTimer = null
let refreshTimer = null

const routePrefix = computed(() => (props.portal === 'university' ? '/university' : '/enterprise'))
const normalizedScenario = computed(() => {
  const raw = String(props.scenario || '').trim().toLowerCase()
  return raw === 'group' || raw === 'single' ? raw : 'all'
})
const titleText = computed(() => {
  if (props.portal === 'university') {
    if (normalizedScenario.value === 'group') return '高校端群面工作台'
    if (normalizedScenario.value === 'single') return '高校端真人面试工作台'
    return '高校端面试管理工作台'
  }
  if (normalizedScenario.value === 'group') return '企业端群面工作台'
  if (normalizedScenario.value === 'single') return '企业端真人面试工作台'
  return '企业端面试管理工作台'
})
const subtitleText = computed(() => {
  if (normalizedScenario.value === 'group') {
    return '群面邀请处理、开考状态追踪与房间进入（AI 面试官 + 系统评分）'
  }
  if (normalizedScenario.value === 'single') {
    return '真人 1v1 邀请处理、进行中监控与历史评价（按评分模板人工给分）'
  }
  return '邀请处理、面试进行中监控与历史评价统一视图'
})

function normalizeScenarioType(rawScenario) {
  const normalized = String(rawScenario || '').trim().toLowerCase()
  return normalized === 'group' ? 'group' : normalized === 'single' ? 'single' : ''
}

function getItemTargetParticipants(item) {
  const raw = Number(item?.target_participants ?? 0)
  return Number.isFinite(raw) ? raw : 0
}

const isGroupInvitation = (item) => {
  const scenarioType = normalizeScenarioType(item?.scenario_type)
  const targetParticipants = getItemTargetParticipants(item)
  const startThreshold = Number(item?.start_threshold || 0)
  return scenarioType === 'group' || targetParticipants > 2 || startThreshold > 2
}

const matchesScenarioFilter = (item) => {
  if (normalizedScenario.value === 'group') return isGroupInvitation(item)
  if (normalizedScenario.value === 'single') return !isGroupInvitation(item)
  return true
}

const scopedInviteList = computed(() => store.inviteList.filter(matchesScenarioFilter))
const scopedPending = computed(() => store.pending.filter(matchesScenarioFilter))
const scopedProcessed = computed(() => store.processed.filter(matchesScenarioFilter))
const scopedInProgress = computed(() => store.inProgress.filter(matchesScenarioFilter))
const scopedHistory = computed(() => store.history.filter(matchesScenarioFilter))

const currentList = computed(() => {
  if (store.activeTab === 'pending') return scopedPending.value
  if (store.activeTab === 'processed') return scopedProcessed.value
  if (store.activeTab === 'in_progress') return scopedInProgress.value
  if (store.activeTab === 'history') return scopedHistory.value
  return scopedInviteList.value
})

const tabs = computed(() => ([
  { key: 'invite_list', label: '邀请列表', count: scopedInviteList.value.length },
  { key: 'pending', label: '待处理', count: scopedPending.value.length },
  { key: 'processed', label: '已处理', count: scopedProcessed.value.length },
  { key: 'in_progress', label: '正在面试', count: scopedInProgress.value.length },
  { key: 'history', label: '面试历史', count: scopedHistory.value.length }
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

const deleteInvitation = async (item) => {
  const invitationId = Number(item?.id || 0)
  if (!invitationId) return
  if (!window.confirm('删除后该邀请记录将从工作台移除，是否继续？')) {
    return
  }
  try {
    await store.deleteInvitation(invitationId)
    ElMessage.success('邀请记录已删除')
  } catch (error) {
    ElMessage.error(error?.response?.data?.error || '删除邀请失败')
  }
}

const enterLiveRoom = (item) => {
  const invitationId = String(item?.id || '').trim()
  if (!invitationId) {
    ElMessage.warning('邀请信息无效，无法进入房间')
    return
  }

  const isGroup = normalizedScenario.value === 'group' ? true : isGroupInvitation(item)
  const query = {}
  const invitationCode = String(item?.invitation_code || '').trim()
  if (invitationCode) {
    query.invitation_code = invitationCode
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
