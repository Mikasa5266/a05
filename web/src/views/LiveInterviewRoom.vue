<script setup>
import { computed, nextTick, onMounted, onBeforeUnmount, ref, watch } from 'vue'
import { useRoute, useRouter, onBeforeRouteLeave } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Video, VideoOff, Mic, MicOff, PhoneOff, ArrowLeft, Users, Copy, Link2, Loader2 } from 'lucide-vue-next'
import { useUserStore } from '../stores/user'
import {
  endInterview,
  getHumanInvitations,
  getReceivedHumanInvitations,
  joinLiveInterview,
  startInterview
} from '../api/interview'
import { generateReport } from '../api/report'
import { API_BASE_URL, WEBRTC_ICE_SERVERS } from '../utils/backend'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const localVideoRef = ref(null)
const remoteVideoRef = ref(null)

const invitation = ref(null)
const loading = ref(false)
const statusText = ref('待进入房间')
const roomId = ref('')
const invitationCode = ref('')
const remoteUserId = ref('')
const interviewId = ref(0)

const cameraOn = ref(true)
const micOn = ref(true)
const finishing = ref(false)
const joining = ref(false)
const isRouteLeaving = ref(false)
const messageInput = ref('')
const questionInput = ref('')
const sharedCode = ref('')
const messages = ref([])
const members = ref([])

let localStream = null
let peer = null
let signalSocket = null
let pendingCandidates = []
let isMakingOffer = false

const role = computed(() => userStore.userInfo?.role || '')
const selfUserId = computed(() => String(userStore.userInfo?.id || ''))
const isStudent = computed(() => role.value === 'student')
const canPublishQuestion = computed(() => role.value === 'enterprise' || role.value === 'university')

const backPath = computed(() => {
  if (role.value === 'enterprise') return '/enterprise/interview-workbench'
  if (role.value === 'university') return '/university/interview-workbench'
  return '/interview/live/workbench'
})

const hasRoom = computed(() => Boolean(roomId.value))
const waitingStatus = computed(() => {
  if (!hasRoom.value) return '未加入房间'
  if (remoteUserId.value) return '连线中'
  if (signalSocket?.readyState === WebSocket.OPEN) return '等候中'
  return '连接中'
})

const roomMembers = computed(() => {
  return members.value.sort((a, b) => (a.isSelf === b.isSelf ? 0 : a.isSelf ? -1 : 1))
})

const inviteLink = computed(() => {
  if (!invitation.value?.id) return ''
  const url = new URL(window.location.href)
  url.searchParams.delete('roomId')
  url.searchParams.set('invitation_id', String(invitation.value.id))
  if (invitationCode.value) {
    url.searchParams.set('invitation_code', invitationCode.value)
  }
  return url.toString()
})

function goBack() {
  router.push(backPath.value)
}

function resolveInvitationIdFromRoute() {
  const fromQuery = Number(route.query?.invitation_id || 0)
  if (fromQuery > 0) return fromQuery
  return 0
}

function resolveInvitationCodeFromRoute() {
  return String(route.query?.invitation_code || '').trim()
}

function getWsSignalUrl() {
  const url = new URL(API_BASE_URL, window.location.origin)
  url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
  url.pathname = `${url.pathname.replace(/\/$/, '')}/interview/live/ws`
  url.searchParams.set('room_id', roomId.value)
  url.searchParams.set('invitation_code', invitationCode.value || resolveInvitationCodeFromRoute())
  url.searchParams.set('token', userStore.token || '')
  return url.toString()
}

function getSelfDisplayName() {
  return userStore.userInfo?.username || (isStudent.value ? '学生' : '面试官')
}

function upsertMember(payload = {}) {
  const id = String(payload.userId || payload.user_id || '')
  if (!id) return
  const displayName = String(payload.senderName || payload.sender_name || payload.username || '').trim() || '参会者'
  const next = {
    userId: id,
    displayName,
    role: payload.role || '',
    isSelf: id === selfUserId.value
  }
  const index = members.value.findIndex((item) => item.userId === id)
  if (index >= 0) {
    members.value[index] = {
      ...members.value[index],
      ...next
    }
  } else {
    members.value.push(next)
  }
}

function removeMember(userId) {
  const id = String(userId || '')
  if (!id) return
  members.value = members.value.filter((item) => item.userId !== id)
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

async function loadInvitationByID(invitationID) {
  if (invitationID <= 0) return
  let list = []
  if (isStudent.value) {
    const res = await getHumanInvitations()
    list = Array.isArray(res?.invitations) ? res.invitations : []
  } else {
    const res = await getReceivedHumanInvitations()
    list = Array.isArray(res?.invitations) ? res.invitations : []
  }
  const target = list.find((item) => Number(item.id) === invitationID)
  if (!target) {
    invitation.value = null
    return
  }
  invitation.value = target
  invitationCode.value = String(target?.invitation_code || resolveInvitationCodeFromRoute())
  interviewId.value = Number(target.interview_id || 0)
}

async function ensureInterviewSession() {
  if (!isStudent.value) return
  if (!invitation.value) return
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
  if (createdId <= 0) return
  interviewId.value = createdId
  invitation.value = {
    ...invitation.value,
    interview_id: createdId
  }
}

async function authorizeJoin(invitationID) {
  const res = await joinLiveInterview({
    invitation_id: Number(invitationID || 0),
    invitation_code: invitationCode.value || resolveInvitationCodeFromRoute()
  })

  const session = res?.session || {}
  const authorizedRoomId = String(session?.room_id || '').trim()
  if (!authorizedRoomId) {
    throw new Error('房间授权信息无效')
  }

  roomId.value = authorizedRoomId
  invitationCode.value = String(session?.invitation_code || invitationCode.value || '')

  const syncedInterviewId = Number(session?.interview_id || 0)
  if (syncedInterviewId > 0) {
    interviewId.value = syncedInterviewId
  }
}

async function initLocalMedia() {
  if (!window.isSecureContext) {
    throw new Error('当前环境不安全，请使用 HTTPS 或 localhost 访问')
  }
  if (!navigator.mediaDevices?.getUserMedia) {
    throw new Error('当前浏览器无法访问摄像头/麦克风')
  }
  localStream = await navigator.mediaDevices.getUserMedia({ video: true, audio: true })
  bindLocalStreamToVideo()
}

function bindLocalStreamToVideo() {
  if (!localVideoRef.value || !localStream) return
  localVideoRef.value.srcObject = localStream
  localVideoRef.value.play?.().catch(() => {})
}

function ensurePeer() {
  if (peer) return peer
  peer = new RTCPeerConnection({ iceServers: WEBRTC_ICE_SERVERS })
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
    if (!remoteStream || !remoteVideoRef.value) return
    remoteVideoRef.value.srcObject = remoteStream
    statusText.value = '音视频已连通'
  }
  peer.onconnectionstatechange = () => {
    if (!peer) return
    if (peer.connectionState === 'connected') {
      statusText.value = '连接稳定'
    } else if (peer.connectionState === 'disconnected' || peer.connectionState === 'failed') {
      statusText.value = '连接中断，等待重连'
    }
  }
  return peer
}

function sendSignal(type, data = {}) {
  if (!signalSocket || signalSocket.readyState !== WebSocket.OPEN) return
  signalSocket.send(JSON.stringify({
    type,
    interview_id: roomId.value,
    data
  }))
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

function syncCode() {
  sendSignal('code_sync', {
    code: sharedCode.value,
    sender_name: getSelfDisplayName(),
    role: role.value
  })
}

async function createAndSendOffer() {
  if (!isStudent.value || isMakingOffer) return
  isMakingOffer = true
  try {
    const pc = ensurePeer()
    const offer = await pc.createOffer()
    await pc.setLocalDescription(offer)
    sendSignal('offer', offer)
    statusText.value = '已发起通话邀请，等待接听'
  } finally {
    isMakingOffer = false
  }
}

async function handleSignalMessage(raw) {
  const msg = JSON.parse(raw)
  if (!msg?.type) return
  const senderID = String(msg.user_id || '')
  const isSelfMessage = senderID && senderID === selfUserId.value
  if (senderID) {
    upsertMember({
      userId: senderID,
      senderName: msg?.data?.sender_name,
      role: msg?.data?.role
    })
  }
  if (isSelfMessage) return
  if (senderID) {
    remoteUserId.value = senderID
  }
  const pc = ensurePeer()

  if (msg.type === 'join') {
    statusText.value = '对方已进入房间'
    if (isStudent.value) {
      await createAndSendOffer()
    } else {
      sendSignal('ready', { ok: true, sender_name: getSelfDisplayName(), role: role.value })
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
    statusText.value = '正在建立连接'
    return
  }
  if (msg.type === 'answer') {
    await pc.setRemoteDescription(new RTCSessionDescription(msg.data))
    while (pendingCandidates.length > 0) {
      const candidate = pendingCandidates.shift()
      await pc.addIceCandidate(new RTCIceCandidate(candidate))
    }
    statusText.value = '连接协商完成'
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
    removeMember(senderID)
    if (senderID === remoteUserId.value) {
      remoteUserId.value = ''
      statusText.value = '对方已离开房间'
    }
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
  if (msg.type === 'code_sync') {
    sharedCode.value = String(msg?.data?.code || '')
    return
  }
  if (msg.type === 'session_sync') {
    const syncedInterviewId = Number(msg?.data?.interview_id || 0)
    if (syncedInterviewId > 0 && interviewId.value === 0) {
      interviewId.value = syncedInterviewId
    }
  }
}

function connectSignalSocket() {
  signalSocket = new WebSocket(getWsSignalUrl())
  signalSocket.onopen = () => {
    statusText.value = '已进入房间，等待成员加入'
    upsertMember({
      userId: selfUserId.value,
      senderName: getSelfDisplayName(),
      role: role.value
    })
    sendSignal('join', {
      role: role.value,
      sender_name: getSelfDisplayName()
    })
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
    statusText.value = '信令连接异常'
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
    ElMessage.success('真人面试报告已生成')
  }
}

function cleanup() {
  if (signalSocket) {
    if (signalSocket.readyState === WebSocket.OPEN) {
      sendSignal('leave', {
        user_id: selfUserId.value,
        sender_name: getSelfDisplayName(),
        role: role.value
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
  if (peer) {
    peer.onicecandidate = null
    peer.ontrack = null
    peer.onconnectionstatechange = null
    peer.close()
    peer = null
  }
  if (localStream) {
    localStream.getTracks().forEach((track) => track.stop())
    localStream = null
  }
  if (localVideoRef.value) {
    localVideoRef.value.srcObject = null
  }
  if (remoteVideoRef.value) {
    remoteVideoRef.value.srcObject = null
  }
  remoteUserId.value = ''
  pendingCandidates = []
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

async function initAndJoinRoom(invitationID = 0) {
  if (joining.value || isRouteLeaving.value) return
  joining.value = true
  loading.value = true
  statusText.value = '正在准备房间...'
  try {
    cleanup()
    members.value = []
    messages.value = []
    sharedCode.value = ''

    if (invitationID <= 0) {
      throw new Error('缺少 invitation_id，无法进入真人面试房间')
    }

    await loadInvitationByID(invitationID)
    await ensureInterviewSession()
    await authorizeJoin(invitationID)
    await initLocalMedia()
    connectSignalSocket()
    await nextTick()
    bindLocalStreamToVideo()
  } catch (err) {
    const message = err?.response?.data?.error || err.message || '进入房间失败'
    ElMessage.error(message)
    cleanup()
    roomId.value = ''
  } finally {
    loading.value = false
    joining.value = false
  }
}

async function copyInviteLink() {
  if (!inviteLink.value) return
  try {
    await navigator.clipboard.writeText(inviteLink.value)
    ElMessage.success('邀请链接已复制')
  } catch {
    ElMessage.warning('复制失败，请手动复制地址栏链接')
  }
}

onMounted(async () => {
  const invitationID = resolveInvitationIdFromRoute()
  if (invitationID <= 0) {
    statusText.value = '仅支持受邀用户进入真人面试房间'
    ElMessage.warning('请从面试管理工作台进入指定房间')
    loading.value = false
    goBack()
    return
  }

  await initAndJoinRoom(invitationID)
})

watch(() => route.query?.invitation_id, async () => {
  const invitationID = resolveInvitationIdFromRoute()
  if (invitationID <= 0 || joining.value || isRouteLeaving.value) return
  if (Number(invitation.value?.id || 0) === invitationID && hasRoom.value) return
  await initAndJoinRoom(invitationID)
})

watch(localVideoRef, () => {
  bindLocalStreamToVideo()
})

onBeforeRouteLeave((_to, _from, next) => {
  isRouteLeaving.value = true
  cleanup()
  next()
})

onBeforeUnmount(() => {
  isRouteLeaving.value = true
  cleanup()
})
</script>

<template>
  <div class="live-room-page min-h-screen text-slate-100">
    <div class="h-full p-5 md:p-8">
      <div class="max-w-[1600px] mx-auto h-full">
        <header class="flex items-center justify-between mb-5">
          <div>
            <h1 class="text-2xl md:text-3xl font-bold">真人协同面试会议室</h1>
            <p class="text-sm text-slate-300 mt-1">{{ invitation?.position || '实时语音/视频/代码协同面试' }}</p>
          </div>
          <button class="px-4 py-2 rounded-xl border border-slate-600 bg-slate-800/70 hover:bg-slate-700 transition-colors text-sm font-medium flex items-center gap-1.5" @click="goBack">
            <ArrowLeft class="w-4 h-4" />
            返回
          </button>
        </header>

        <div v-if="!hasRoom && !loading" class="h-[calc(100vh-12rem)] flex items-center justify-center">
          <div class="w-full max-w-2xl rounded-3xl border border-slate-700 bg-slate-900/70 backdrop-blur p-8 space-y-6">
            <h2 class="text-2xl font-bold">等待房间授权</h2>
            <p class="text-sm text-slate-300">真人面试仅允许受邀用户通过 invitation_id + invitation_code 进入。</p>
            <p class="text-xs text-slate-400">请从企业端/高校端面试管理工作台点击“进入面试房间”。</p>
            <button class="w-full py-3 rounded-xl border border-slate-500/70 bg-slate-800/70 text-slate-100 font-semibold hover:bg-slate-700 transition-colors" @click="goBack">
              返回工作台
            </button>
          </div>
        </div>

        <div v-else class="grid grid-cols-1 xl:grid-cols-[360px_1fr] gap-5 h-[calc(100vh-12rem)]">
          <aside class="rounded-3xl border border-slate-700 bg-slate-900/70 backdrop-blur p-5 space-y-4 overflow-y-auto">
            <div class="rounded-2xl border border-slate-700 bg-slate-950/60 p-4">
              <p class="text-xs text-slate-400 uppercase tracking-wider mb-2">房间号</p>
              <p class="text-base font-semibold break-all">{{ roomId || '--' }}</p>
              <button class="mt-3 w-full px-3 py-2 rounded-xl border border-indigo-300 bg-indigo-500/20 text-indigo-200 hover:bg-indigo-500/35 transition-colors text-sm inline-flex items-center justify-center gap-1.5" @click="copyInviteLink">
                <Copy class="w-4 h-4" />
                复制邀请链接
              </button>
              <div class="mt-2 text-xs text-slate-400 break-all inline-flex items-center gap-1.5">
                <Link2 class="w-3.5 h-3.5" />
                <span>{{ inviteLink }}</span>
              </div>
            </div>

            <div class="rounded-2xl border border-slate-700 bg-slate-950/60 p-4">
              <div class="flex items-center justify-between">
                <p class="text-xs text-slate-400 uppercase tracking-wider">等候室状态</p>
                <span class="text-xs px-2 py-0.5 rounded-full" :class="waitingStatus === '连线中' ? 'bg-emerald-500/25 text-emerald-300' : 'bg-amber-500/20 text-amber-300'">{{ waitingStatus }}</span>
              </div>
              <p class="text-sm mt-2 text-slate-200">{{ statusText }}</p>
            </div>

            <div class="rounded-2xl border border-slate-700 bg-slate-950/60 p-4">
              <p class="text-xs text-slate-400 uppercase tracking-wider mb-3">已加入成员</p>
              <div class="space-y-2">
                <div v-for="member in roomMembers" :key="member.userId" class="rounded-xl border border-slate-700 bg-slate-900 px-3 py-2 flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <Users class="w-4 h-4 text-indigo-300" />
                    <div>
                      <p class="text-sm font-medium">{{ member.displayName }}</p>
                      <p class="text-[11px] text-slate-400">{{ member.role || 'member' }}</p>
                    </div>
                  </div>
                  <span v-if="member.isSelf" class="text-[10px] px-1.5 py-0.5 rounded-full bg-indigo-500/30 text-indigo-200">我</span>
                </div>
                <p v-if="roomMembers.length === 0" class="text-xs text-slate-500">暂无成员</p>
              </div>
            </div>
          </aside>

          <main class="rounded-3xl border border-slate-700 bg-slate-900/70 backdrop-blur p-4 md:p-5 flex flex-col gap-4 overflow-hidden">
            <div v-if="loading" class="h-full flex items-center justify-center text-slate-300 gap-2">
              <Loader2 class="w-5 h-5 animate-spin" />
              <span>正在初始化设备与连接...</span>
            </div>

            <template v-else>
              <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
                <section class="rounded-2xl border border-slate-700 bg-slate-950/70 p-3">
                  <p class="text-xs text-slate-400 mb-2">我的画面</p>
                  <div class="aspect-video rounded-xl overflow-hidden bg-slate-950 border border-slate-800">
                    <video ref="localVideoRef" autoplay playsinline muted class="w-full h-full object-cover"></video>
                  </div>
                </section>
                <section class="rounded-2xl border border-slate-700 bg-slate-950/70 p-3">
                  <p class="text-xs text-slate-400 mb-2">对方画面</p>
                  <div class="aspect-video rounded-xl overflow-hidden bg-slate-950 border border-slate-800">
                    <video ref="remoteVideoRef" autoplay playsinline class="w-full h-full object-cover"></video>
                  </div>
                </section>
              </div>

              <div class="grid grid-cols-1 lg:grid-cols-2 gap-4 min-h-0 flex-1">
                <section class="rounded-2xl border border-slate-700 bg-slate-950/70 p-3 flex flex-col min-h-0">
                  <p class="text-xs text-slate-400 mb-2">聊天与发题协同</p>
                  <div class="flex-1 min-h-[180px] rounded-xl border border-slate-800 bg-slate-950 p-2.5 overflow-y-auto space-y-2">
                    <p v-if="messages.length === 0" class="text-xs text-slate-500 text-center py-12">等待消息中...</p>
                    <div
                      v-for="item in messages"
                      :key="item.id"
                      class="rounded-xl px-3 py-2 text-sm"
                      :class="[
                        item.kind === 'question'
                          ? 'border border-amber-500/40 bg-amber-500/15 text-amber-100'
                          : item.fromSelf
                            ? 'border border-emerald-500/40 bg-emerald-500/15 text-emerald-100'
                            : 'border border-slate-700 bg-slate-900 text-slate-100'
                      ]"
                    >
                      <div class="flex items-center justify-between mb-1">
                        <span class="text-xs opacity-80">{{ item.kind === 'question' ? '题目' : '聊天' }} · {{ item.senderName }}</span>
                        <span class="text-[10px] opacity-60">{{ item.createdAt }}</span>
                      </div>
                      <p class="whitespace-pre-wrap break-words">{{ item.text }}</p>
                    </div>
                  </div>
                  <div class="mt-2 space-y-2">
                    <div class="flex gap-2">
                      <input v-model="messageInput" type="text" class="flex-1 rounded-xl border border-slate-700 bg-slate-950 px-3 py-2 text-sm outline-none focus:border-indigo-400" placeholder="输入聊天内容" @keyup.enter="sendChatMessage" />
                      <button class="px-4 py-2 rounded-xl border border-indigo-300 bg-indigo-500/20 text-indigo-200 text-sm font-medium hover:bg-indigo-500/35" @click="sendChatMessage">发送</button>
                    </div>
                    <div class="flex gap-2">
                      <input v-model="questionInput" type="text" class="flex-1 rounded-xl border border-slate-700 bg-slate-950 px-3 py-2 text-sm outline-none focus:border-amber-400" :disabled="!canPublishQuestion" :placeholder="canPublishQuestion ? '输入面试题并发布' : '仅面试官可发布题目'" @keyup.enter="publishQuestion" />
                      <button class="px-4 py-2 rounded-xl border text-sm font-medium" :class="canPublishQuestion ? 'border-amber-300 bg-amber-500/20 text-amber-200 hover:bg-amber-500/35' : 'border-slate-700 bg-slate-900 text-slate-500 cursor-not-allowed'" :disabled="!canPublishQuestion" @click="publishQuestion">发题</button>
                    </div>
                  </div>
                </section>

                <section class="rounded-2xl border border-slate-700 bg-slate-950/70 p-3 flex flex-col min-h-0">
                  <p class="text-xs text-slate-400 mb-2">代码协同区</p>
                  <textarea
                    v-model="sharedCode"
                    class="flex-1 min-h-[260px] rounded-xl border border-slate-800 bg-slate-950 px-3 py-3 text-sm text-slate-100 outline-none focus:border-violet-400 font-mono resize-none"
                    placeholder="在此协同编写代码，内容会实时同步给房间成员"
                    @input="syncCode"
                  ></textarea>
                </section>
              </div>
            </template>
          </main>
        </div>

        <div v-if="hasRoom" class="mt-4 rounded-2xl border border-slate-700 bg-slate-900/60 backdrop-blur p-3 flex flex-wrap gap-3">
          <button class="px-4 py-2 rounded-xl border text-sm font-medium transition-colors flex items-center gap-1.5" :class="micOn ? 'border-emerald-300 bg-emerald-500/20 text-emerald-200 hover:bg-emerald-500/35' : 'border-rose-300 bg-rose-500/20 text-rose-200 hover:bg-rose-500/35'" @click="toggleMic">
            <Mic v-if="micOn" class="w-4 h-4" />
            <MicOff v-else class="w-4 h-4" />
            {{ micOn ? '麦克风开启' : '麦克风关闭' }}
          </button>
          <button class="px-4 py-2 rounded-xl border text-sm font-medium transition-colors flex items-center gap-1.5" :class="cameraOn ? 'border-emerald-300 bg-emerald-500/20 text-emerald-200 hover:bg-emerald-500/35' : 'border-rose-300 bg-rose-500/20 text-rose-200 hover:bg-rose-500/35'" @click="toggleCamera">
            <Video v-if="cameraOn" class="w-4 h-4" />
            <VideoOff v-else class="w-4 h-4" />
            {{ cameraOn ? '摄像头开启' : '摄像头关闭' }}
          </button>
          <button class="ml-auto px-4 py-2 rounded-xl border border-rose-300 bg-rose-500/20 text-rose-200 text-sm font-semibold hover:bg-rose-500/35 transition-colors flex items-center gap-1.5" :disabled="finishing" @click="leaveRoom">
            <PhoneOff class="w-4 h-4" />
            {{ finishing ? '正在结束...' : '结束并离开' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.live-room-page {
  background:
    radial-gradient(circle at 15% 15%, rgba(99, 102, 241, 0.25), transparent 38%),
    radial-gradient(circle at 85% 12%, rgba(14, 165, 233, 0.2), transparent 34%),
    linear-gradient(160deg, #020617 0%, #0f172a 48%, #111827 100%);
}
</style>
