<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Video, VideoOff, Mic, MicOff, PhoneOff, ArrowLeft, Users } from 'lucide-vue-next'
import { useUserStore } from '../stores/user'
import {
  endInterview,
  getHumanInvitations,
  getReceivedHumanInvitations,
  startInterview
} from '../api/interview'
import { generateReport } from '../api/report'
import { API_BASE_URL } from '../utils/backend'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const localVideoRef = ref(null)
const remoteVideoRef = ref(null)

const invitation = ref(null)
const loading = ref(true)
const statusText = ref('正在准备房间...')
const roomId = ref('')
const remoteUserId = ref('')
const interviewId = ref(0)

const cameraOn = ref(true)
const micOn = ref(true)
const finishing = ref(false)
const messageInput = ref('')
const questionInput = ref('')
const messages = ref([])

let localStream = null
let peer = null
let signalSocket = null
let pendingCandidates = []
let isMakingOffer = false

const role = computed(() => userStore.userInfo?.role || '')
const selfUserId = computed(() => String(userStore.userInfo?.id || ''))
const isStudent = computed(() => role.value === 'student')
const canPublishQuestion = computed(() => role.value === 'enterprise' || role.value === 'university')

const peerDisplayName = computed(() => {
  if (!invitation.value) return '对方'
  if (isStudent.value) return invitation.value?.invitee?.username || '面试官'
  return invitation.value?.student?.username || '学生'
})

const invitationStatusLabel = computed(() => {
  const map = {
    pending: '待处理',
    accepted: '已接受',
    in_progress: '进行中',
    completed: '已完成',
    rejected: '已拒绝',
    cancelled: '已取消'
  }
  return map[invitation.value?.status] || '未知'
})

const backPath = computed(() => {
  if (role.value === 'enterprise') return '/enterprise/dashboard'
  if (role.value === 'university') return '/university/dashboard'
  return '/student/interview'
})

function goBack() {
  router.push(backPath.value)
}

function getWsSignalUrl() {
  const url = new URL(API_BASE_URL, window.location.origin)
  url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
  url.pathname = `${url.pathname.replace(/\/$/, '')}/interview/live/ws`
  url.searchParams.set('room_id', roomId.value)
  url.searchParams.set('token', userStore.token || '')
  return url.toString()
}

function getSelfDisplayName() {
  return userStore.userInfo?.username || (isStudent.value ? '学生' : '面试官')
}

function appendMessage(kind, text, fromSelf, senderName) {
  const content = String(text || '').trim()
  if (!content) return
  messages.value.push({
    id: `${Date.now()}-${Math.random().toString(16).slice(2)}`,
    kind,
    text: content,
    fromSelf,
    senderName,
    createdAt: new Date().toLocaleTimeString()
  })
}

async function loadInvitation() {
  const invitationId = Number(route.query.invitation_id || 0)
  if (!invitationId) {
    throw new Error('缺少 invitation_id 参数')
  }

  let list = []
  if (isStudent.value) {
    const res = await getHumanInvitations()
    list = Array.isArray(res?.invitations) ? res.invitations : []
  } else {
    const res = await getReceivedHumanInvitations()
    list = Array.isArray(res?.invitations) ? res.invitations : []
  }

  const target = list.find((item) => Number(item.id) === invitationId)
  if (!target) {
    throw new Error('没有找到该邀请，或你无权限进入该房间')
  }
  if (target.status !== 'accepted' && target.status !== 'in_progress') {
    throw new Error(`当前邀请状态为 ${target.status}，暂不可进入视频面试`)
  }

  invitation.value = target
  roomId.value = `invitation-${target.id}`
  interviewId.value = Number(target.interview_id || 0)
}

async function ensureInterviewSession() {
  if (!isStudent.value) return
  if (interviewId.value > 0) return

  const payload = {
    position: invitation.value?.position || '真人模拟面试',
    difficulty: invitation.value?.difficulty || 'campus_intern',
    mode: invitation.value?.mode || 'comprehensive',
    style: invitation.value?.style || 'gentle',
    company: invitation.value?.company || '',
    interview_mode: 'human',
    invitation_id: Number(invitation.value?.id || 0)
  }

  const res = await startInterview(payload)
  const createdId = Number(res?.interview?.id || 0)
  if (!createdId) {
    throw new Error('真人面试会话创建失败，请稍后重试')
  }
  interviewId.value = createdId
  invitation.value = {
    ...invitation.value,
    interview_id: createdId,
    status: 'in_progress'
  }
}

async function initLocalMedia() {
  if (!window.isSecureContext) {
    throw new Error('当前浏览器环境不安全，无法访问摄像头/麦克风。请使用 HTTPS 或 localhost 访问。')
  }

  if (!navigator.mediaDevices?.getUserMedia) {
    throw new Error('当前浏览器不支持音视频采集，或权限被浏览器策略限制。请更换现代浏览器并检查权限设置。')
  }

  localStream = await navigator.mediaDevices.getUserMedia({ video: true, audio: true })
  bindLocalStreamToVideo()
}

function bindLocalStreamToVideo() {
  if (localVideoRef.value) {
    localVideoRef.value.srcObject = localStream
    localVideoRef.value.play?.().catch(() => {})
  }
}

function ensurePeer() {
  if (peer) return peer

  peer = new RTCPeerConnection({
    iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
  })

  localStream?.getTracks().forEach((track) => {
    peer.addTrack(track, localStream)
  })

  peer.onicecandidate = (event) => {
    if (event.candidate) {
      sendSignal('candidate', event.candidate)
    }
  }

  peer.ontrack = (event) => {
    const [remoteStream] = event.streams
    if (remoteVideoRef.value && remoteStream) {
      remoteVideoRef.value.srcObject = remoteStream
      statusText.value = `已连通 ${peerDisplayName.value}`
    }
  }

  peer.onconnectionstatechange = () => {
    if (!peer) return
    if (peer.connectionState === 'connected') {
      statusText.value = `已连通 ${peerDisplayName.value}`
    } else if (peer.connectionState === 'disconnected' || peer.connectionState === 'failed') {
      statusText.value = '连接中断，正在等待对方重连...'
    }
  }

  return peer
}

function sendSignal(type, data = {}) {
  if (!signalSocket || signalSocket.readyState !== WebSocket.OPEN) return
  signalSocket.send(
    JSON.stringify({
      type,
      interview_id: roomId.value,
      data
    })
  )
}

function sendChatMessage() {
  const content = messageInput.value.trim()
  if (!content) return
  appendMessage('chat', content, true, getSelfDisplayName())
  sendSignal('chat', {
    text: content,
    sender_name: getSelfDisplayName(),
    role: role.value,
    interview_id: interviewId.value
  })
  messageInput.value = ''
}

function publishQuestion() {
  if (!canPublishQuestion.value) return
  const content = questionInput.value.trim()
  if (!content) return
  appendMessage('question', content, true, getSelfDisplayName())
  sendSignal('question', {
    text: content,
    sender_name: getSelfDisplayName(),
    role: role.value,
    interview_id: interviewId.value
  })
  questionInput.value = ''
}

async function createAndSendOffer() {
  if (!isStudent.value || isMakingOffer) return
  isMakingOffer = true
  try {
    const pc = ensurePeer()
    const offer = await pc.createOffer()
    await pc.setLocalDescription(offer)
    sendSignal('offer', offer)
    statusText.value = '已发起通话邀请，等待对方接听...'
  } finally {
    isMakingOffer = false
  }
}

async function handleSignalMessage(raw) {
  const msg = JSON.parse(raw)
  if (!msg?.type) return
  if (String(msg.user_id || '') === selfUserId.value) return

  if (msg.user_id) {
    remoteUserId.value = String(msg.user_id)
  }

  const pc = ensurePeer()

  if (msg.type === 'join') {
    if (isStudent.value) {
      await createAndSendOffer()
    } else {
      sendSignal('ready', { ok: true })
      statusText.value = '对方已进入房间，等待建立连接...'
    }
    return
  }

  if (msg.type === 'ready') {
    if (isStudent.value) {
      await createAndSendOffer()
    }
    return
  }

  if (msg.type === 'offer') {
    await pc.setRemoteDescription(new RTCSessionDescription(msg.data))
    const answer = await pc.createAnswer()
    await pc.setLocalDescription(answer)
    sendSignal('answer', answer)
    while (pendingCandidates.length > 0) {
      const candidate = pendingCandidates.shift()
      await pc.addIceCandidate(new RTCIceCandidate(candidate))
    }
    statusText.value = '正在建立连接...'
    return
  }

  if (msg.type === 'answer') {
    await pc.setRemoteDescription(new RTCSessionDescription(msg.data))
    while (pendingCandidates.length > 0) {
      const candidate = pendingCandidates.shift()
      await pc.addIceCandidate(new RTCIceCandidate(candidate))
    }
    statusText.value = '连接协商完成，正在拉起音视频...'
    return
  }

  if (msg.type === 'candidate') {
    if (pc.remoteDescription) {
      await pc.addIceCandidate(new RTCIceCandidate(msg.data))
    } else {
      pendingCandidates.push(msg.data)
    }
    return
  }

  if (msg.type === 'leave') {
    statusText.value = '对方已离开房间'
    return
  }

  if (msg.type === 'chat' || msg.type === 'question') {
    appendMessage(
      msg.type,
      msg?.data?.text,
      false,
      msg?.data?.sender_name || (msg.type === 'question' ? '面试官' : '对方')
    )
    return
  }

  if (msg.type === 'session_sync') {
    const syncedInterviewId = Number(msg?.data?.interview_id || 0)
    if (syncedInterviewId > 0 && interviewId.value === 0) {
      interviewId.value = syncedInterviewId
      invitation.value = {
        ...invitation.value,
        interview_id: syncedInterviewId,
        status: invitation.value?.status === 'accepted' ? 'in_progress' : invitation.value?.status
      }
    }
  }
}

function connectSignalSocket() {
  signalSocket = new WebSocket(getWsSignalUrl())

  signalSocket.onopen = () => {
    statusText.value = '已进入房间，等待对方上线...'
    sendSignal('join', { role: role.value })
    if (interviewId.value > 0) {
      sendSignal('session_sync', { interview_id: interviewId.value })
    }
  }

  signalSocket.onmessage = async (event) => {
    try {
      await handleSignalMessage(event.data)
    } catch (err) {
      console.error('signal message handling failed', err)
    }
  }

  signalSocket.onerror = () => {
    statusText.value = '信令连接异常，请刷新重试'
  }

  signalSocket.onclose = () => {
    if (statusText.value !== '对方已离开房间') {
      statusText.value = '信令已断开'
    }
  }
}

function toggleMic() {
  if (!localStream) return
  micOn.value = !micOn.value
  localStream.getAudioTracks().forEach((track) => {
    track.enabled = micOn.value
  })
}

function toggleCamera() {
  if (!localStream) return
  cameraOn.value = !cameraOn.value
  localStream.getVideoTracks().forEach((track) => {
    track.enabled = cameraOn.value
  })
}

async function finalizeInterviewAndReport() {
  if (!isStudent.value || interviewId.value <= 0) return

  await endInterview(interviewId.value)
  const reportRes = await generateReport({ interview_id: interviewId.value })
  const reportId = Number(reportRes?.report?.id || 0)
  if (reportId > 0) {
    ElMessage.success('真人面试报告已生成并写入历史记录')
  } else {
    ElMessage.success('真人面试已结束，报告生成完成')
  }
}

function cleanup() {
  if (signalSocket && signalSocket.readyState === WebSocket.OPEN) {
    sendSignal('leave', { user_id: selfUserId.value })
    signalSocket.close()
  }
  signalSocket = null

  if (peer) {
    peer.close()
    peer = null
  }

  if (localStream) {
    localStream.getTracks().forEach((track) => track.stop())
    localStream = null
  }
}

async function leaveRoom() {
  if (finishing.value) return
  finishing.value = true
  try {
    await finalizeInterviewAndReport()
  } catch (err) {
    const message = err?.response?.data?.error || err.message || '结束面试失败'
    ElMessage.warning(message)
  } finally {
    cleanup()
    goBack()
    finishing.value = false
  }
}

onMounted(async () => {
  try {
    await loadInvitation()
    await ensureInterviewSession()
    await initLocalMedia()
    connectSignalSocket()
  } catch (err) {
    const message = err?.response?.data?.error || err.message || '进入房间失败'
    ElMessage.error(message)
    goBack()
  } finally {
    loading.value = false
    await nextTick()
    bindLocalStreamToVideo()
  }
})

watch(localVideoRef, () => {
  bindLocalStreamToVideo()
})

onUnmounted(() => {
  cleanup()
})
</script>

<template>
  <div class="space-y-6">
    <header class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-zinc-900">真人视频面试房间</h1>
        <p class="text-sm text-zinc-500 mt-1">
          {{ invitation?.position || '-' }} · {{ invitationStatusLabel }} · 房间 {{ roomId || '--' }}
        </p>
      </div>
      <button
        class="px-4 py-2 rounded-xl border border-zinc-200 text-zinc-700 hover:bg-zinc-50 transition-colors text-sm font-medium flex items-center gap-1.5"
        @click="goBack"
      >
        <ArrowLeft class="w-4 h-4" />
        返回
      </button>
    </header>

    <div class="rounded-2xl border border-zinc-100 bg-white p-4 flex items-center justify-between">
      <div class="flex items-center gap-2 text-sm text-zinc-600">
        <Users class="w-4 h-4 text-indigo-600" />
        <span>当前连线对象：{{ peerDisplayName }}</span>
      </div>
      <span class="text-xs px-2.5 py-1 rounded-full bg-indigo-50 text-indigo-700">{{ statusText }}</span>
    </div>

    <div v-if="loading" class="rounded-2xl border border-zinc-100 bg-white p-12 text-center text-zinc-500">
      正在加载视频设备与房间...
    </div>

    <div v-else class="grid grid-cols-1 xl:grid-cols-2 gap-6">
      <section class="rounded-3xl border border-zinc-100 bg-white p-5">
        <h2 class="font-semibold text-zinc-900 mb-3">我的画面</h2>
        <div class="aspect-video rounded-2xl overflow-hidden bg-zinc-900">
          <video ref="localVideoRef" autoplay playsinline muted class="w-full h-full object-cover"></video>
        </div>
      </section>

      <section class="rounded-3xl border border-zinc-100 bg-white p-5">
        <h2 class="font-semibold text-zinc-900 mb-3">对方画面</h2>
        <div class="aspect-video rounded-2xl overflow-hidden bg-zinc-900">
          <video ref="remoteVideoRef" autoplay playsinline class="w-full h-full object-cover"></video>
        </div>
      </section>

      <section class="rounded-3xl border border-zinc-100 bg-white p-5 xl:col-span-2">
        <div class="flex items-center justify-between mb-3">
          <h2 class="font-semibold text-zinc-900">文字聊天与问题板</h2>
          <span class="text-xs text-zinc-500">面试过程文字协同，结束后自动生成报告</span>
        </div>

        <div class="rounded-2xl border border-zinc-100 bg-zinc-50 p-3 h-64 overflow-y-auto space-y-2">
          <div v-if="messages.length === 0" class="text-sm text-zinc-400 text-center pt-20">
            暂无消息，面试官可以在下方发布题目。
          </div>
          <div
            v-for="item in messages"
            :key="item.id"
            class="rounded-xl px-3 py-2 text-sm"
            :class="[
              item.kind === 'question'
                ? 'border border-amber-200 bg-amber-50 text-amber-800'
                : item.fromSelf
                  ? 'border border-emerald-200 bg-emerald-50 text-emerald-800'
                  : 'border border-zinc-200 bg-white text-zinc-700'
            ]"
          >
            <div class="flex items-center justify-between mb-1">
              <span class="font-medium">{{ item.kind === 'question' ? '题目' : '聊天' }} · {{ item.senderName }}</span>
              <span class="text-xs opacity-70">{{ item.createdAt }}</span>
            </div>
            <p class="whitespace-pre-wrap wrap-break-word">{{ item.text }}</p>
          </div>
        </div>

        <div class="grid grid-cols-1 lg:grid-cols-2 gap-3 mt-3">
          <div class="flex gap-2">
            <input
              v-model="messageInput"
              type="text"
              class="flex-1 rounded-xl border border-zinc-200 px-3 py-2 text-sm outline-none focus:border-indigo-400"
              placeholder="输入聊天内容（例如追问、反馈）"
              @keyup.enter="sendChatMessage"
            />
            <button
              class="px-4 py-2 rounded-xl border border-indigo-200 bg-indigo-50 text-indigo-700 text-sm font-medium hover:bg-indigo-100"
              @click="sendChatMessage"
            >
              发送
            </button>
          </div>

          <div class="flex gap-2">
            <input
              v-model="questionInput"
              type="text"
              class="flex-1 rounded-xl border border-zinc-200 px-3 py-2 text-sm outline-none focus:border-amber-400"
              :placeholder="canPublishQuestion ? '输入题目内容，回车即可发布' : '仅面试官可发布题目'"
              :disabled="!canPublishQuestion"
              @keyup.enter="publishQuestion"
            />
            <button
              class="px-4 py-2 rounded-xl border text-sm font-medium"
              :class="canPublishQuestion ? 'border-amber-200 bg-amber-50 text-amber-700 hover:bg-amber-100' : 'border-zinc-200 bg-zinc-100 text-zinc-400 cursor-not-allowed'"
              :disabled="!canPublishQuestion"
              @click="publishQuestion"
            >
              发题
            </button>
          </div>
        </div>
      </section>
    </div>

    <div class="rounded-3xl border border-zinc-100 bg-white p-4 flex flex-wrap gap-3">
      <button
        class="px-4 py-2 rounded-xl border text-sm font-medium transition-colors flex items-center gap-1.5"
        :class="micOn ? 'border-emerald-200 bg-emerald-50 text-emerald-700 hover:bg-emerald-100' : 'border-rose-200 bg-rose-50 text-rose-700 hover:bg-rose-100'"
        @click="toggleMic"
      >
        <Mic v-if="micOn" class="w-4 h-4" />
        <MicOff v-else class="w-4 h-4" />
        {{ micOn ? '麦克风已开启' : '麦克风已关闭' }}
      </button>

      <button
        class="px-4 py-2 rounded-xl border text-sm font-medium transition-colors flex items-center gap-1.5"
        :class="cameraOn ? 'border-emerald-200 bg-emerald-50 text-emerald-700 hover:bg-emerald-100' : 'border-rose-200 bg-rose-50 text-rose-700 hover:bg-rose-100'"
        @click="toggleCamera"
      >
        <Video v-if="cameraOn" class="w-4 h-4" />
        <VideoOff v-else class="w-4 h-4" />
        {{ cameraOn ? '摄像头已开启' : '摄像头已关闭' }}
      </button>

      <button
        class="ml-auto px-4 py-2 rounded-xl border border-rose-200 bg-rose-50 text-rose-700 text-sm font-semibold hover:bg-rose-100 transition-colors flex items-center gap-1.5"
        :disabled="finishing"
        @click="leaveRoom"
      >
        <PhoneOff class="w-4 h-4" />
        {{ finishing ? '正在结束...' : '结束并离开' }}
      </button>
    </div>
  </div>
</template>
