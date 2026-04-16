<template>
  <div class="space-y-6">
    <header class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-2xl md:text-3xl font-bold text-zinc-900">学生端群面工作台</h1>
        <p class="text-sm text-zinc-500 mt-1">邀请同伴、追踪群面状态并进入独立群面房间</p>
        <p class="text-xs text-indigo-600 mt-2">测试阶段支持 2 人开考，目标容量默认 4 人（预留扩展到 5 人）。</p>
      </div>
      <div class="flex items-center gap-2">
        <button
          class="px-4 py-2 rounded-xl border border-zinc-200 bg-white text-zinc-700 text-sm font-medium hover:bg-zinc-50 transition-colors disabled:opacity-60"
          :disabled="loading"
          @click="fetchInvitations"
        >
          刷新
        </button>
        <button
          class="px-4 py-2 rounded-xl bg-indigo-600 text-white text-sm font-semibold hover:bg-indigo-700 transition-colors"
          @click="openCreateDialog"
        >
          发起群面邀请
        </button>
      </div>
    </header>

    <section class="bg-white rounded-2xl border border-zinc-100 p-3 shadow-sm">
      <div class="flex flex-wrap gap-2">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          class="px-3.5 py-2 rounded-xl text-sm font-medium border transition-colors"
          :class="activeTab === tab.key ? 'bg-indigo-50 text-indigo-700 border-indigo-100' : 'bg-white text-zinc-600 border-transparent hover:bg-zinc-50'"
          @click="activeTab = tab.key"
        >
          {{ tab.label }}
          <span class="ml-1.5 text-xs px-1.5 py-0.5 rounded-full" :class="activeTab === tab.key ? 'bg-indigo-100 text-indigo-700' : 'bg-zinc-100 text-zinc-600'">
            {{ tab.count }}
          </span>
        </button>
      </div>
    </section>

    <section class="bg-white rounded-3xl border border-zinc-100 p-5 md:p-6 shadow-sm min-h-90">
      <div v-if="loading" class="h-56 flex items-center justify-center text-zinc-500 text-sm">
        正在加载群面邀请...
      </div>

      <div v-else-if="filteredInvitations.length === 0" class="h-56 flex items-center justify-center text-zinc-400 text-sm">
        当前没有群面邀请记录
      </div>

      <div v-else class="space-y-4">
        <article
          v-for="inv in filteredInvitations"
          :key="inv.id"
          class="rounded-2xl border border-zinc-100 p-4 md:p-5"
        >
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div>
              <div class="flex items-center gap-2 flex-wrap">
                <h3 class="text-base font-semibold text-zinc-900">{{ invitationCounterpartName(inv) }}</h3>
                <span class="px-2 py-1 rounded-full text-xs font-medium" :class="statusClass(inv.status)">
                  {{ statusLabel(inv.status) }}
                </span>
                <span class="px-2 py-1 rounded-full text-xs font-medium bg-emerald-50 text-emerald-700">
                  群面
                </span>
                <span class="px-2 py-1 rounded-full text-xs font-medium" :class="inv.direction === 'incoming' ? 'bg-emerald-50 text-emerald-700' : 'bg-sky-50 text-sky-700'">
                  {{ inv.direction === 'incoming' ? '收到邀请' : '发起邀请' }}
                </span>
              </div>
              <p class="text-sm text-zinc-600 mt-1">
                {{ inv.position || '群面模拟场景' }} · {{ difficultyLabel(inv.difficulty) }} · {{ modeLabel(inv.mode) }}
              </p>
              <p class="text-xs text-zinc-500 mt-1">
                目标人数：{{ invitationTargetParticipants(inv) }} 人 · 开考阈值：{{ invitationStartThreshold(inv) }} 人
              </p>
              <p class="text-xs text-zinc-500 mt-1">邀请码：{{ inv.invitation_code || '-' }}</p>
              <p class="text-xs text-zinc-400 mt-1">拟定时间：{{ formatDateTime(inv.scheduled_at) }}</p>
              <p v-if="inv.notes" class="text-xs text-zinc-500 mt-2">备注：{{ inv.notes }}</p>
            </div>

            <div class="text-right min-w-45">
              <p class="text-xs text-zinc-500">更新时间</p>
              <p class="text-sm text-zinc-700 mt-1">{{ formatDateTime(inv.updated_at) }}</p>
            </div>
          </div>

          <div class="mt-4 flex flex-wrap gap-2">
            <button
              v-if="canRespondInvitation(inv)"
              class="px-3 py-1.5 rounded-lg bg-emerald-50 text-emerald-700 text-xs font-semibold hover:bg-emerald-100"
              @click="respondInvitation(inv, 'accept')"
            >
              接受邀请
            </button>
            <button
              v-if="canRespondInvitation(inv)"
              class="px-3 py-1.5 rounded-lg bg-rose-50 text-rose-700 text-xs font-semibold hover:bg-rose-100"
              @click="respondInvitation(inv, 'reject')"
            >
              拒绝邀请
            </button>
            <button
              v-if="canDeleteInvitation(inv)"
              class="px-3 py-1.5 rounded-lg bg-zinc-100 text-zinc-700 text-xs font-semibold hover:bg-zinc-200"
              @click="deleteInvitation(inv)"
            >
              删除记录
            </button>
            <button
              v-if="canEnterRoom(inv)"
              class="ml-auto px-3.5 py-1.5 rounded-lg border border-indigo-200 bg-indigo-50 text-indigo-700 text-xs font-semibold hover:bg-indigo-100"
              @click="enterLiveRoom(inv)"
            >
              进入群面房间
            </button>
          </div>
        </article>
      </div>
    </section>

    <div
      v-if="showCreateDialog"
      class="fixed inset-0 z-50 bg-black/45 backdrop-blur-sm flex items-center justify-center p-4"
      @click.self="showCreateDialog = false"
    >
      <div class="w-full max-w-3xl bg-white rounded-3xl border border-zinc-100 shadow-2xl p-6">
        <h2 class="text-xl font-bold text-zinc-900 mb-4">发起群面邀请</h2>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase">同伴角色范围</label>
            <select
              v-model="candidateRoleFilter"
              class="mt-1 w-full px-3 py-2.5 rounded-xl border border-zinc-200 bg-zinc-50 text-sm"
              @change="fetchCandidates"
            >
              <option value="student">学生端</option>
              <option value="">全部</option>
            </select>
          </div>

          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase">目标人数</label>
            <select v-model.number="createForm.target_participants" class="mt-1 w-full px-3 py-2.5 rounded-xl border border-zinc-200 bg-zinc-50 text-sm">
              <option :value="2">2 人（测试）</option>
              <option :value="3">3 人</option>
              <option :value="4">4 人</option>
              <option :value="5">5 人</option>
            </select>
          </div>

          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase">开考阈值</label>
            <select v-model.number="createForm.start_threshold" class="mt-1 w-full px-3 py-2.5 rounded-xl border border-zinc-200 bg-zinc-50 text-sm">
              <option :value="2">2 人</option>
              <option :value="3" :disabled="createForm.target_participants < 3">3 人</option>
              <option :value="4" :disabled="createForm.target_participants < 4">4 人</option>
              <option :value="5" :disabled="createForm.target_participants < 5">5 人</option>
            </select>
          </div>

          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase">预约时间（可选）</label>
            <input v-model="createForm.scheduled_at" type="datetime-local" class="mt-1 w-full px-3 py-2.5 rounded-xl border border-zinc-200 bg-zinc-50 text-sm" />
          </div>

          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase">目标岗位</label>
            <input v-model.trim="createForm.position" class="mt-1 w-full px-3 py-2.5 rounded-xl border border-zinc-200 bg-zinc-50 text-sm" placeholder="例如：产品经理" />
          </div>

          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase">面试类型</label>
            <select v-model="createForm.mode" class="mt-1 w-full px-3 py-2.5 rounded-xl border border-zinc-200 bg-zinc-50 text-sm">
              <option value="comprehensive">综合面</option>
              <option value="technical">技术面</option>
              <option value="hr">HR 面</option>
            </select>
          </div>

          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase">难度等级</label>
            <select v-model="createForm.difficulty" class="mt-1 w-full px-3 py-2.5 rounded-xl border border-zinc-200 bg-zinc-50 text-sm">
              <option value="campus_intern">校招实习</option>
              <option value="campus_graduate">校招全职</option>
              <option value="social_junior">社招初级</option>
            </select>
          </div>

          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase">面试风格</label>
            <select v-model="createForm.style" class="mt-1 w-full px-3 py-2.5 rounded-xl border border-zinc-200 bg-zinc-50 text-sm">
              <option value="gentle">引导型</option>
              <option value="stress">压力型</option>
              <option value="deep">深挖型</option>
              <option value="practical">实战型</option>
            </select>
          </div>
        </div>

        <div class="mt-4">
          <label class="text-xs font-bold text-zinc-400 uppercase">邀请同伴（最多 4 位）</label>
          <div class="mt-2 rounded-2xl border border-zinc-100 bg-zinc-50 p-3 max-h-55 overflow-y-auto">
            <p v-if="candidatesLoading" class="text-xs text-zinc-400">正在加载可邀请用户...</p>
            <p v-else-if="candidates.length === 0" class="text-xs text-zinc-400">暂无可邀请用户</p>
            <div v-else class="space-y-2">
              <label
                v-for="candidate in candidates"
                :key="candidate.id"
                class="flex items-center gap-3 rounded-xl border px-3 py-2 text-sm"
                :class="selectedInviteeIds.includes(candidate.id) ? 'border-indigo-200 bg-indigo-50 text-indigo-700' : 'border-zinc-200 bg-white text-zinc-700'"
              >
                <input
                  type="checkbox"
                  class="accent-indigo-600"
                  :checked="selectedInviteeIds.includes(candidate.id)"
                  :disabled="!selectedInviteeIds.includes(candidate.id) && selectedInviteeIds.length >= 4"
                  @change="toggleInvitee(candidate.id)"
                />
                <span class="font-medium">{{ candidate.username }}</span>
                <span class="text-xs text-zinc-500">{{ roleLabel(candidate.role) }}</span>
              </label>
            </div>
          </div>
          <p class="mt-2 text-xs text-zinc-500">
            已选 {{ selectedInviteeIds.length }} 位。当前配置：总人数 {{ createForm.target_participants }}，开考阈值 {{ normalizedStartThreshold }}。
          </p>
        </div>

        <div class="mt-4">
          <label class="text-xs font-bold text-zinc-400 uppercase">备注（可选）</label>
          <textarea v-model.trim="createForm.notes" rows="3" class="mt-1 w-full px-3 py-2.5 rounded-xl border border-zinc-200 bg-zinc-50 text-sm resize-none" placeholder="可填写讨论目标或观察重点"></textarea>
        </div>

        <div class="mt-5 flex items-center justify-end gap-2">
          <button class="px-4 py-2 rounded-lg text-zinc-600 hover:bg-zinc-100" @click="showCreateDialog = false">取消</button>
          <button
            class="px-5 py-2 rounded-xl bg-indigo-600 text-white text-sm font-semibold hover:bg-indigo-700 disabled:opacity-60"
            :disabled="createSubmitting"
            @click="submitCreateInvitation"
          >
            {{ createSubmitting ? '提交中...' : '发送群面邀请' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  createHumanInvitation,
  deleteHumanInvitation,
  getHumanInvitations,
  getInviteCandidates,
  getReceivedHumanInvitations,
  respondHumanInvitation,
} from '../../api/interview'

const router = useRouter()

const loading = ref(false)
const invitations = ref([])
const activeTab = ref('all')

const showCreateDialog = ref(false)
const candidates = ref([])
const candidatesLoading = ref(false)
const candidateRoleFilter = ref('student')
const createSubmitting = ref(false)

const createForm = reactive({
  invitee_user_ids: [],
  target_participants: 4,
  start_threshold: 2,
  position: '群面模拟场景',
  difficulty: 'campus_intern',
  mode: 'comprehensive',
  style: 'gentle',
  notes: '',
  scheduled_at: '',
})

const normalizedStartThreshold = computed(() => {
  const target = Number(createForm.target_participants || 2)
  let threshold = Number(createForm.start_threshold || 2)
  if (!Number.isFinite(threshold) || threshold < 2) threshold = 2
  if (threshold > target) threshold = target
  return threshold
})

const groupInvitations = computed(() => invitations.value.filter(isGroupInvitation))

const tabs = computed(() => {
  const list = groupInvitations.value
  const countByStatus = (status) => list.filter((item) => item.status === status).length
  return [
    { key: 'all', label: '邀请列表', count: list.length },
    { key: 'pending', label: '待处理', count: countByStatus('pending') },
    { key: 'processed', label: '已处理', count: list.filter((i) => i.status !== 'pending').length },
    { key: 'in_progress', label: '进行中', count: countByStatus('in_progress') },
    { key: 'history', label: '历史记录', count: countByStatus('completed') },
  ]
})

const filteredInvitations = computed(() => {
  const list = groupInvitations.value
  if (activeTab.value === 'pending') return list.filter((item) => item.status === 'pending')
  if (activeTab.value === 'processed') return list.filter((item) => item.status !== 'pending')
  if (activeTab.value === 'in_progress') return list.filter((item) => item.status === 'in_progress')
  if (activeTab.value === 'history') return list.filter((item) => item.status === 'completed')
  return list
})

const selectedInviteeIds = computed(() => createForm.invitee_user_ids)

function isGroupInvitation(invitation) {
  const scenarioType = String(invitation?.scenario_type || '').trim().toLowerCase()
  const targetParticipants = Number(invitation?.target_participants || 0)
  return scenarioType === 'group' || targetParticipants > 2
}

function invitationTargetParticipants(invitation) {
  const parsed = Number(invitation?.target_participants || 0)
  if (!Number.isFinite(parsed) || parsed < 2) return 2
  return parsed
}

function invitationStartThreshold(invitation) {
  const target = invitationTargetParticipants(invitation)
  const parsed = Number(invitation?.start_threshold || 0)
  let threshold = Number.isFinite(parsed) ? parsed : 2
  if (threshold < 2) threshold = 2
  if (threshold > target) threshold = target
  return threshold
}

function statusLabel(status) {
  const map = {
    pending: '待接受',
    accepted: '已接受',
    rejected: '已拒绝',
    in_progress: '进行中',
    completed: '已完成',
    cancelled: '已取消',
  }
  return map[status] || status || '未知'
}

function statusClass(status) {
  if (status === 'pending') return 'bg-amber-50 text-amber-700'
  if (status === 'accepted') return 'bg-emerald-50 text-emerald-700'
  if (status === 'rejected') return 'bg-rose-50 text-rose-700'
  if (status === 'in_progress') return 'bg-indigo-50 text-indigo-700'
  if (status === 'completed') return 'bg-zinc-100 text-zinc-700'
  return 'bg-zinc-100 text-zinc-600'
}

function difficultyLabel(difficulty) {
  const map = {
    campus_intern: '校招实习',
    campus_graduate: '校招全职',
    social_junior: '社招初级',
  }
  return map[difficulty] || difficulty || '未设置'
}

function modeLabel(mode) {
  const map = {
    technical: '技术面',
    hr: 'HR 面',
    comprehensive: '综合面',
    blindbox: '盲盒面',
  }
  return map[mode] || mode || '未设置'
}

function roleLabel(role) {
  if (role === 'enterprise') return '企业端'
  if (role === 'university') return '高校端'
  if (role === 'student') return '学生端'
  return '未知角色'
}

function invitationDirection(inv) {
  return String(inv?.direction || 'outgoing')
}

function invitationCounterpartName(inv) {
  if (invitationDirection(inv) === 'incoming') {
    return inv?.initiator?.username || inv?.student?.username || inv?.counterpart_name || `发起人#${inv?.initiator_user_id || inv?.student_id || '-'}`
  }
  return inv?.target?.username || inv?.invitee?.username || inv?.counterpart_name || `受邀者#${inv?.invitee_user_id || inv?.target_user_id || '-'}`
}

function canRespondInvitation(inv) {
  return invitationDirection(inv) === 'incoming' && inv?.status === 'pending'
}

function canDeleteInvitation(inv) {
  if (!inv) return false
  if (inv.status === 'in_progress') return false
  if (String(inv.interview_status || '').trim() === 'in_progress') return false
  return true
}

function canEnterRoom(inv) {
  if (!inv) return false
  if (inv.status === 'accepted' || inv.status === 'in_progress') return true
  if (invitationDirection(inv) === 'incoming' && inv.status === 'pending') return true
  return false
}

function formatDateTime(value) {
  if (!value) return '待确认'
  const dt = new Date(value)
  if (Number.isNaN(dt.getTime())) return '待确认'
  return dt.toLocaleString('zh-CN', { hour12: false })
}

async function fetchInvitations() {
  loading.value = true
  try {
    const [sentRes, receivedRes] = await Promise.all([
      getHumanInvitations(),
      getReceivedHumanInvitations(),
    ])
    const sent = Array.isArray(sentRes?.invitations) ? sentRes.invitations : []
    const received = Array.isArray(receivedRes?.invitations) ? receivedRes.invitations : []

    const merged = new Map()
    sent.forEach((item) => {
      merged.set(Number(item.id), { ...item, direction: 'outgoing' })
    })
    received.forEach((item) => {
      const key = Number(item.id)
      if (merged.has(key)) {
        merged.set(key, { ...merged.get(key), ...item, direction: 'incoming' })
      } else {
        merged.set(key, { ...item, direction: 'incoming' })
      }
    })

    invitations.value = Array.from(merged.values()).sort((a, b) => {
      const ta = new Date(a?.updated_at || a?.created_at || 0).getTime()
      const tb = new Date(b?.updated_at || b?.created_at || 0).getTime()
      return tb - ta
    })
  } catch (error) {
    ElMessage.error(error?.response?.data?.error || '加载邀请失败')
  } finally {
    loading.value = false
  }
}

async function respondInvitation(invitation, action) {
  const invitationId = Number(invitation?.id || 0)
  if (!invitationId) return
  try {
    await respondHumanInvitation(invitationId, action)
    ElMessage.success(action === 'accept' ? '已接受邀请' : '已拒绝邀请')
    await fetchInvitations()
  } catch (error) {
    ElMessage.error(error?.response?.data?.error || '处理邀请失败')
  }
}

async function deleteInvitation(invitation) {
  const invitationId = Number(invitation?.id || 0)
  if (!invitationId) return
  if (!window.confirm('删除后该邀请记录将从工作台移除，是否继续？')) {
    return
  }
  try {
    await deleteHumanInvitation(invitationId)
    ElMessage.success('邀请记录已删除')
    await fetchInvitations()
  } catch (error) {
    ElMessage.error(error?.response?.data?.error || '删除邀请失败')
  }
}

async function fetchCandidates() {
  candidatesLoading.value = true
  try {
    const res = await getInviteCandidates({
      role: candidateRoleFilter.value || undefined,
      page: 1,
      page_size: 80,
    })
    candidates.value = Array.isArray(res?.users) ? res.users : []
  } catch (error) {
    ElMessage.error(error?.response?.data?.error || '加载可邀请用户失败')
  } finally {
    candidatesLoading.value = false
  }
}

async function openCreateDialog() {
  showCreateDialog.value = true
  await fetchCandidates()
}

function toggleInvitee(userID) {
  const id = Number(userID || 0)
  if (!id) return
  const exists = createForm.invitee_user_ids.includes(id)
  if (exists) {
    createForm.invitee_user_ids = createForm.invitee_user_ids.filter((value) => value !== id)
    return
  }
  if (createForm.invitee_user_ids.length >= 4) {
    ElMessage.warning('最多可选择 4 位同伴')
    return
  }
  createForm.invitee_user_ids = [...createForm.invitee_user_ids, id]
}

async function submitCreateInvitation() {
  if (!createForm.position.trim()) {
    ElMessage.warning('请填写目标岗位')
    return
  }
  if (createForm.invitee_user_ids.length === 0) {
    ElMessage.warning('请至少选择 1 位同伴')
    return
  }

  const targetParticipants = Number(createForm.target_participants || 4)
  if (targetParticipants < 2 || targetParticipants > 5) {
    ElMessage.warning('目标人数需在 2 到 5 人之间')
    return
  }
  if (createForm.invitee_user_ids.length > targetParticipants - 1) {
    ElMessage.warning('已选同伴人数超过目标人数配置')
    return
  }

  createSubmitting.value = true
  try {
    const payload = {
      invitee_user_ids: [...createForm.invitee_user_ids],
      position: createForm.position.trim(),
      difficulty: createForm.difficulty,
      mode: createForm.mode,
      style: createForm.style,
      notes: createForm.notes.trim(),
      scenario_type: 'group',
      target_participants: targetParticipants,
      start_threshold: normalizedStartThreshold.value,
    }

    if (createForm.scheduled_at) {
      const dt = new Date(createForm.scheduled_at)
      if (!Number.isNaN(dt.getTime())) {
        payload.scheduled_at = dt.toISOString()
      }
    }

    await createHumanInvitation(payload)
    ElMessage.success('群面邀请已发送')
    showCreateDialog.value = false
    createForm.invitee_user_ids = []
    createForm.notes = ''
    createForm.scheduled_at = ''
    await fetchInvitations()
  } catch (error) {
    ElMessage.error(error?.response?.data?.error || '发送邀请失败')
  } finally {
    createSubmitting.value = false
  }
}

function enterLiveRoom(invitation) {
  const invitationId = String(invitation?.id || '').trim()
  if (!invitationId) {
    ElMessage.warning('邀请信息无效，无法进入房间')
    return
  }

  const invitationCode = String(invitation?.invitation_code || '').trim()
  const query = invitationCode ? { invitation_code: invitationCode } : {}

  router.push({
    path: `/interview/live/group/${encodeURIComponent(invitationId)}`,
    query,
  })
}

onMounted(() => {
  fetchInvitations()
})
</script>
