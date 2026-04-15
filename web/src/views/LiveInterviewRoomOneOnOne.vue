<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { onBeforeRouteLeave, useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Loader2, Mic, MicOff, PhoneOff } from 'lucide-vue-next'
import { useUserStore } from '../stores/user'
import { useLiveHumanStore } from '../stores/useLiveHumanStore'
import { joinLiveInterview } from '../api/interview'
import { API_BASE_URL, WEBRTC_ICE_SERVERS } from '../utils/backend'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const localVideoRef = ref(null)
const remoteVideoRef = ref(null)

const loading = ref(false)
const joining = ref(false)
const finishing = ref(false)
const roomId = ref('')
const invitationCode = ref('')
const statusText = ref('待进入房间')
const isRouteLeaving = ref(false)

let signalSocket = null

const liveHumanStore = useLiveHumanStore()
const micOn = computed(() => liveHumanStore.micOn)

const role = computed(() => String(userStore.userInfo?.role || '').trim().toLowerCase())
const isStudent = computed(() => role.value === 'student')
const hasRoom = computed(() => Boolean(roomId.value))

const backPath = computed(() => {
  const currentPath = String(route.path || '')
  if (currentPath.startsWith('/enterprise/')) return '/enterprise/interview-workbench'
  if (currentPath.startsWith('/university/')) return '/university/interview-workbench'
  return '/interview/live/workbench'
})

function goBack() {
  router.push(backPath.value)
}

function resolveInvitationIdFromRoute() {
  const fromParams = Number(route.params?.id || 0)
  if (fromParams > 0) return fromParams
  const fromQuery = Number(route.query?.invitation_id || 0)
  return fromQuery > 0 ? fromQuery : 0
}

function resolveInvitationCodeFromRoute() {
  return String(route.query?.invitation_code || '').trim()
}

function bindLocalStream() {
  liveHumanStore.setLocalVideoElement(localVideoRef.value)
}

function bindRemoteStream() {
  liveHumanStore.setRemoteVideoElements([remoteVideoRef.value])
}

function getWsSignalUrl() {
  const url = new URL(API_BASE_URL, window.location.origin)
  url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
  url.pathname = `${url.pathname.replace(/\/$/, '')}/ws/interview/live`
  url.searchParams.set('room_id', roomId.value)
  url.searchParams.set('invitation_code', invitationCode.value || resolveInvitationCodeFromRoute())
  url.searchParams.set('token', userStore.token || '')
  return url.toString()
}

function sendSignal(type, data = {}) {
  if (!signalSocket || signalSocket.readyState !== WebSocket.OPEN) return
  signalSocket.send(JSON.stringify({
    type,
    interview_id: roomId.value,
    data
  }))
}

async function createAndSendOffer() {
  if (!isStudent.value) return
  await liveHumanStore.createAndSendOffer({
    iceServers: WEBRTC_ICE_SERVERS,
    sendSignal,
    onStatusChange: (next) => {
      statusText.value = next
    }
  })
}

async function handleSignalMessage(raw) {
  const msg = JSON.parse(raw)
  if (!msg?.type) return

  if (msg.type === 'join') {
    statusText.value = '对端已进入房间'
    if (isStudent.value) {
      await createAndSendOffer()
    } else {
      sendSignal('ready', { ok: true })
    }
    return
  }

  if (msg.type === 'ready') {
    if (isStudent.value) {
      await createAndSendOffer()
    }
    return
  }

  await liveHumanStore.handleSignalMessage(msg, {
    iceServers: WEBRTC_ICE_SERVERS,
    sendSignal,
    onStatusChange: (next) => {
      statusText.value = next
    }
  })
}

async function initLocalMedia() {
  await liveHumanStore.initLocalMedia({ video: true, audio: true })
  bindLocalStream()
}

function connectSignalSocket() {
  signalSocket = new WebSocket(getWsSignalUrl())

  signalSocket.onopen = () => {
    statusText.value = '已进入房间，等待对端连接'
    sendSignal('join', {
      role: role.value,
      sender_name: userStore.userInfo?.username || ''
    })
  }

  signalSocket.onmessage = async (event) => {
    try {
      await handleSignalMessage(event.data)
    } catch (error) {
      console.error('signal message handling failed', error)
    }
  }

  signalSocket.onerror = () => {
    statusText.value = '信令连接异常'
  }

  signalSocket.onclose = () => {
    statusText.value = '信令已断开'
  }
}

function toggleMic() {
  liveHumanStore.toggleMic()
}

function cleanup() {
  if (signalSocket) {
    if (signalSocket.readyState === WebSocket.OPEN) {
      sendSignal('leave', {
        role: role.value,
        sender_name: userStore.userInfo?.username || ''
      })
    }

    signalSocket.onopen = null
    signalSocket.onmessage = null
    signalSocket.onerror = null
    signalSocket.onclose = null

    if (signalSocket.readyState === WebSocket.OPEN || signalSocket.readyState === WebSocket.CONNECTING) {
      signalSocket.close()
    }
  }
  signalSocket = null

  liveHumanStore.cleanup()
}

async function leaveRoom() {
  if (finishing.value) return
  finishing.value = true
  cleanup()
  goBack()
  finishing.value = false
}

async function initAndJoinRoom(invitationID = 0) {
  if (joining.value || isRouteLeaving.value) return

  joining.value = true
  loading.value = true
  statusText.value = '正在准备房间...'

  try {
    cleanup()

    if (invitationID <= 0) {
      throw new Error('缺少 invitation_id，无法进入 1v1 房间')
    }

    const res = await joinLiveInterview({
      invitation_id: invitationID,
      invitation_code: resolveInvitationCodeFromRoute()
    })

    const session = res?.session || {}
    const authorizedRoomId = String(session?.room_id || '').trim()
    if (!authorizedRoomId) {
      throw new Error('房间授权信息无效')
    }

    roomId.value = authorizedRoomId
    invitationCode.value = String(session?.invitation_code || resolveInvitationCodeFromRoute())

    await initLocalMedia()
    connectSignalSocket()

    await nextTick()
    bindLocalStream()
    bindRemoteStream()
  } catch (error) {
    ElMessage.error(error?.response?.data?.error || error?.message || '进入房间失败')
    cleanup()
    roomId.value = ''
  } finally {
    loading.value = false
    joining.value = false
  }
}

onMounted(async () => {
  const invitationID = resolveInvitationIdFromRoute()
  if (invitationID <= 0) {
    ElMessage.warning('请从面试管理工作台进入指定房间')
    goBack()
    return
  }
  await initAndJoinRoom(invitationID)
})

watch(
  () => [route.params?.id, route.query?.invitation_id],
  async () => {
    const invitationID = resolveInvitationIdFromRoute()
    if (invitationID <= 0 || joining.value || isRouteLeaving.value) return
    await initAndJoinRoom(invitationID)
  }
)

onBeforeRouteLeave(() => {
  isRouteLeaving.value = true
  cleanup()
  return true
})

onBeforeUnmount(() => {
  isRouteLeaving.value = true
  cleanup()
})
</script>

<template>
  <div class="min-h-screen bg-slate-950 text-slate-100 px-4 py-5 md:px-8 md:py-6">
    <div class="max-w-350 mx-auto">
      <header class="flex items-start justify-between gap-4 mb-5">
        <div>
          <h1 class="text-2xl md:text-3xl font-semibold tracking-wide">1v1 实战房间</h1>
          <p class="text-sm text-slate-300 mt-1">一对一真人在线面试</p>
        </div>
        <button class="px-3 py-2 rounded-lg border border-slate-700 bg-slate-900 hover:bg-slate-800 inline-flex items-center gap-2" @click="goBack">
          <ArrowLeft class="w-4 h-4" />
          返回
        </button>
      </header>

      <div v-if="!hasRoom && !loading" class="h-[calc(100vh-12rem)] flex items-center justify-center">
        <div class="w-full max-w-xl rounded-2xl border border-slate-700 bg-slate-900 p-8">
          <h2 class="text-xl font-semibold">等待房间授权</h2>
          <p class="text-sm text-slate-300 mt-3">1v1 房间仅允许受邀用户进入。</p>
          <button class="mt-6 w-full rounded-lg bg-sky-600 hover:bg-sky-500 py-2" @click="goBack">返回工作台</button>
        </div>
      </div>

      <div v-else class="rounded-2xl border border-slate-700 bg-slate-900 p-4 md:p-5 h-[calc(100vh-12rem)] flex flex-col">
        <div v-if="loading" class="h-full flex items-center justify-center gap-2 text-slate-300">
          <Loader2 class="w-5 h-5 animate-spin" />
          <span>正在初始化设备与连接...</span>
        </div>

        <template v-else>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-3 mb-4">
            <div class="rounded-xl border border-slate-700 bg-slate-800/60 p-3">
              <p class="text-xs uppercase tracking-wide text-slate-400">房间号</p>
              <p class="mt-2 text-base font-semibold break-all">{{ roomId || '--' }}</p>
            </div>
            <div class="rounded-xl border border-slate-700 bg-slate-800/60 p-3">
              <p class="text-xs uppercase tracking-wide text-slate-400">连接状态</p>
              <p class="mt-2 text-base font-semibold">{{ hasRoom ? '通话房间已就绪' : '未连接' }}</p>
              <p class="text-sm text-slate-300 mt-1">{{ statusText }}</p>
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4 flex-1 min-h-0">
            <article class="rounded-xl border border-slate-700 bg-slate-800/60 p-3 flex flex-col min-h-0">
              <div class="text-xs text-slate-300 mb-2">我的画面</div>
              <div class="relative rounded-lg overflow-hidden border border-slate-700 bg-slate-950 aspect-video">
                <video ref="localVideoRef" autoplay playsinline muted class="w-full h-full object-cover"></video>
              </div>
            </article>

            <article class="rounded-xl border border-slate-700 bg-slate-800/60 p-3 flex flex-col min-h-0">
              <div class="text-xs text-slate-300 mb-2">对端画面</div>
              <div class="relative rounded-lg overflow-hidden border border-slate-700 bg-slate-950 aspect-video">
                <video ref="remoteVideoRef" autoplay playsinline class="w-full h-full object-cover"></video>
              </div>
            </article>
          </div>

          <div class="mt-4 flex flex-wrap gap-3">
            <button class="px-4 py-2 rounded-lg border border-slate-700 bg-slate-800 hover:bg-slate-700 inline-flex items-center gap-2" @click="toggleMic">
              <Mic v-if="micOn" class="w-4 h-4" />
              <MicOff v-else class="w-4 h-4" />
              {{ micOn ? '语音开启中' : '语音关闭' }}
            </button>
            <button
              class="ml-auto px-4 py-2 rounded-lg border border-rose-400/40 bg-rose-500/20 hover:bg-rose-500/30 inline-flex items-center gap-2 disabled:opacity-60"
              :disabled="finishing"
              @click="leaveRoom"
            >
              <PhoneOff class="w-4 h-4" />
              {{ finishing ? '正在结束...' : '结束并离开' }}
            </button>
          </div>
        </template>
      </div>
    </div>
  </div>
</template>
