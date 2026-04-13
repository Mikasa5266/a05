<template>
  <div class="space-y-6">
    <header class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-2xl md:text-3xl font-bold text-zinc-900">{{ pageTitle }}</h1>
        <p class="text-sm text-zinc-500 mt-1">{{ pageSubtitle }}</p>
        <p v-if="isGroupMode" class="text-xs text-indigo-600 mt-2">测试阶段：2 人可开始群面流程，目标容量预留 4 人。</p>
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
          {{ inviteButtonText }}
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
        正在加载邀请...
      </div>

      <div v-else-if="filteredInvitations.length === 0" class="h-56 flex items-center justify-center text-zinc-400 text-sm">
        当前没有对应状态的邀请记录
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
                <h3 class="text-base font-semibold text-zinc-900">{{ inv.invitee?.username || `面试官#${inv.invitee_user_id}` }}</h3>
                <span class="px-2 py-1 rounded-full text-xs font-medium" :class="statusClass(inv.status)">
                  {{ statusLabel(inv.status) }}
                </span>
                <span class="px-2 py-1 rounded-full text-xs font-medium bg-zinc-100 text-zinc-600">
                  {{ roleLabel(inv.invitee_role) }}
                </span>
              </div>
              <p class="text-sm text-zinc-600 mt-1">
                {{ inv.position || '未设置岗位' }} · {{ difficultyLabel(inv.difficulty) }} · {{ modeLabel(inv.mode) }}
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
              v-if="inv.status === 'accepted' || inv.status === 'in_progress'"
              class="ml-auto px-3.5 py-1.5 rounded-lg border border-indigo-200 bg-indigo-50 text-indigo-700 text-xs font-semibold hover:bg-indigo-100"
              @click="enterLiveRoom(inv)"
            >
              {{ enterRoomText }}
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
      <div class="w-full max-w-2xl bg-white rounded-3xl border border-zinc-100 shadow-2xl p-6">
        <h2 class="text-xl font-bold text-zinc-900 mb-4">{{ dialogTitle }}</h2>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase">邀请对象类型</label>
            <select
              v-model="candidateRoleFilter"
              class="mt-1 w-full px-3 py-2.5 rounded-xl border border-zinc-200 bg-zinc-50 text-sm"
              @change="fetchCandidates"
            >
              <option value="">全部</option>
              <option value="enterprise">企业端</option>
              <option value="university">高校端</option>
            </select>
          </div>

          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase">选择面试官</label>
            <select v-model.number="createForm.invitee_user_id" class="mt-1 w-full px-3 py-2.5 rounded-xl border border-zinc-200 bg-zinc-50 text-sm">
              <option :value="0">请选择</option>
              <option v-for="candidate in candidates" :key="candidate.id" :value="candidate.id">
                {{ candidate.username }}（{{ roleLabel(candidate.role) }}）
              </option>
            </select>
            <p v-if="candidatesLoading" class="text-xs text-zinc-400 mt-1">正在加载可邀请用户...</p>
          </div>

          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase">目标岗位</label>
            <input v-model.trim="createForm.position" class="mt-1 w-full px-3 py-2.5 rounded-xl border border-zinc-200 bg-zinc-50 text-sm" placeholder="例如：前端工程师" />
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
            <label class="text-xs font-bold text-zinc-400 uppercase">面试类型</label>
            <select v-model="createForm.mode" class="mt-1 w-full px-3 py-2.5 rounded-xl border border-zinc-200 bg-zinc-50 text-sm">
              <option value="technical">技术面</option>
              <option value="hr">HR面</option>
              <option value="comprehensive">综合面</option>
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

          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase">公司偏好（可选）</label>
            <input v-model.trim="createForm.company" class="mt-1 w-full px-3 py-2.5 rounded-xl border border-zinc-200 bg-zinc-50 text-sm" placeholder="例如：字节跳动" />
          </div>

          <div>
            <label class="text-xs font-bold text-zinc-400 uppercase">预约时间（可选）</label>
            <input v-model="createForm.scheduled_at" type="datetime-local" class="mt-1 w-full px-3 py-2.5 rounded-xl border border-zinc-200 bg-zinc-50 text-sm" />
          </div>
        </div>

        <div class="mt-4">
          <label class="text-xs font-bold text-zinc-400 uppercase">备注（可选）</label>
          <textarea v-model.trim="createForm.notes" rows="3" class="mt-1 w-full px-3 py-2.5 rounded-xl border border-zinc-200 bg-zinc-50 text-sm resize-none" placeholder="可填写希望重点考察的方向"></textarea>
        </div>

        <div class="mt-5 flex items-center justify-end gap-2">
          <button class="px-4 py-2 rounded-lg text-zinc-600 hover:bg-zinc-100" @click="showCreateDialog = false">取消</button>
          <button
            class="px-5 py-2 rounded-xl bg-indigo-600 text-white text-sm font-semibold hover:bg-indigo-700 disabled:opacity-60"
            :disabled="createSubmitting"
            @click="submitCreateInvitation"
          >
            {{ createSubmitting ? '提交中...' : '发送邀请' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { createHumanInvitation, getHumanInvitations, getInviteCandidates } from '../../api/interview'

const router = useRouter()
const route = useRoute()

const isGroupMode = computed(() => String(route.query?.group_mode || '') === '1')
const pageTitle = computed(() => (isGroupMode.value ? '学生端群面工作台' : '学生端真人面试工作台'))
const pageSubtitle = computed(() => (isGroupMode.value ? '发起群面邀请、追踪状态、进入群面房间' : '发起邀请、追踪状态、进入房间'))
const dialogTitle = computed(() => (isGroupMode.value ? '发起群面邀请' : '发起真人面试邀请'))
const inviteButtonText = computed(() => (isGroupMode.value ? '发起群面邀请' : '发起邀请'))
const enterRoomText = computed(() => (isGroupMode.value ? '进入群面房间' : '进入真人面试房间'))

const loading = ref(false)
const invitations = ref([])
const activeTab = ref('all')

const showCreateDialog = ref(false)
const candidates = ref([])
const candidatesLoading = ref(false)
const candidateRoleFilter = ref('')
const createSubmitting = ref(false)

const createForm = reactive({
  invitee_user_id: 0,
  position: '群面模拟场景',
  difficulty: 'campus_intern',
  mode: 'comprehensive',
  style: 'gentle',
  company: '',
  scheduled_at: '',
  notes: ''
})

const tabs = computed(() => {
  const list = invitations.value
  const countByStatus = (status) => list.filter((item) => item.status === status).length
  return [
    { key: 'all', label: '邀请列表', count: list.length },
    { key: 'pending', label: '待处理', count: countByStatus('pending') },
    { key: 'processed', label: '已处理', count: list.filter((i) => i.status !== 'pending').length },
    { key: 'in_progress', label: '正在面试', count: countByStatus('in_progress') },
    { key: 'history', label: '历史记录', count: countByStatus('completed') }
  ]
})

const filteredInvitations = computed(() => {
  const list = invitations.value
  if (activeTab.value === 'pending') return list.filter((item) => item.status === 'pending')
  if (activeTab.value === 'processed') return list.filter((item) => item.status !== 'pending')
  if (activeTab.value === 'in_progress') return list.filter((item) => item.status === 'in_progress')
  if (activeTab.value === 'history') return list.filter((item) => item.status === 'completed')
  return list
})

const statusLabel = (status) => {
  const map = {
    pending: '待接受',
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

const difficultyLabel = (difficulty) => {
  const map = {
    campus_intern: '校招实习',
    campus_graduate: '校招全职',
    social_junior: '社招初级'
  }
  return map[difficulty] || difficulty || '未设置'
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

const roleLabel = (role) => {
  if (role === 'enterprise') return '企业端'
  if (role === 'university') return '高校端'
  return '未知角色'
}

const formatDateTime = (value) => {
  if (!value) return '待确认'
  const dt = new Date(value)
  if (Number.isNaN(dt.getTime())) return '待确认'
  return dt.toLocaleString('zh-CN', { hour12: false })
}

const fetchInvitations = async () => {
  loading.value = true
  try {
    const res = await getHumanInvitations()
    invitations.value = Array.isArray(res?.invitations) ? res.invitations : []
  } catch (error) {
    ElMessage.error(error?.response?.data?.error || '加载邀请失败')
  } finally {
    loading.value = false
  }
}

const fetchCandidates = async () => {
  candidatesLoading.value = true
  try {
    const res = await getInviteCandidates({
      role: candidateRoleFilter.value || undefined,
      page: 1,
      page_size: 50
    })
    candidates.value = Array.isArray(res?.users) ? res.users : []
  } catch (error) {
    ElMessage.error(error?.response?.data?.error || '加载可邀请用户失败')
  } finally {
    candidatesLoading.value = false
  }
}

const openCreateDialog = async () => {
  showCreateDialog.value = true
  await fetchCandidates()
}

const resetCreateForm = () => {
  createForm.invitee_user_id = 0
  createForm.position = isGroupMode.value ? '群面模拟场景' : 'Java后端工程师'
  createForm.difficulty = 'campus_intern'
  createForm.mode = isGroupMode.value ? 'comprehensive' : 'technical'
  createForm.style = 'gentle'
  createForm.company = ''
  createForm.scheduled_at = ''
  createForm.notes = ''
}

const submitCreateInvitation = async () => {
  if (!createForm.invitee_user_id) {
    ElMessage.warning('请先选择面试官')
    return
  }
  if (!createForm.position.trim()) {
    ElMessage.warning('请填写目标岗位')
    return
  }

  createSubmitting.value = true
  try {
    const payload = {
      invitee_user_id: Number(createForm.invitee_user_id),
      position: createForm.position.trim(),
      difficulty: createForm.difficulty,
      mode: createForm.mode,
      style: createForm.style,
      company: createForm.company.trim(),
      notes: createForm.notes.trim()
    }

    if (createForm.scheduled_at) {
      const dt = new Date(createForm.scheduled_at)
      if (!Number.isNaN(dt.getTime())) {
        payload.scheduled_at = dt.toISOString()
      }
    }

    await createHumanInvitation(payload)
    ElMessage.success('邀请已发送')
    showCreateDialog.value = false
    resetCreateForm()
    await fetchInvitations()
  } catch (error) {
    ElMessage.error(error?.response?.data?.error || '发送邀请失败')
  } finally {
    createSubmitting.value = false
  }
}

const enterLiveRoom = (invitation) => {
  router.push({
    path: '/interview/live/room',
    query: {
      invitation_id: String(invitation.id),
      invitation_code: String(invitation.invitation_code || '')
    }
  })
}

onMounted(() => {
  resetCreateForm()
  fetchInvitations()
})
</script>
