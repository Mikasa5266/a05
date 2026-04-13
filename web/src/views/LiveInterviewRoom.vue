<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter, onBeforeRouteLeave } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Copy, Link2, Loader2, Mic, MicOff, PhoneOff, SendHorizontal, Users } from 'lucide-vue-next'
import { useUserStore } from '../stores/user'
import {
  analyzeSpeechChunk,
  endInterview,
  getHumanInvitations,
  getReceivedHumanInvitations,
  joinLiveInterview,
  startLiveInterview,
  startInterview
} from '../api/interview'
import { generateReport } from '../api/report'
import { API_BASE_URL, WEBRTC_ICE_SERVERS } from '../utils/backend'

const TEST_START_THRESHOLD = 2
const TARGET_PARTICIPANTS = 4

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const localVideoRef = ref(null)
const remoteVideoRefA = ref(null)
const remoteVideoRefB = ref(null)
const remoteVideoRefC = ref(null)

const invitation = ref(null)
const loading = ref(false)
const statusText = ref('待进入房间')
const roomId = ref('')
const invitationCode = ref('')
const interviewId = ref(0)

const micOn = ref(true)
const finishing = ref(false)
const joining = ref(false)
const startingInterview = ref(false)
const isRouteLeaving = ref(false)

const groupStarted = ref(false)
const groupReadyCount = ref(0)
const groupStartThreshold = ref(TEST_START_THRESHOLD)
const groupTargetParticipants = ref(TARGET_PARTICIPANTS)

const messageInput = ref('')
const messages = ref([])
const members = ref([])

let localStream = null
let peer = null
let signalSocket = null
let pendingCandidates = []
let isMakingOffer = false
let remoteStream = null

let chunkRecorder = null
let chunkRecorderStream = null

const role = computed(() => userStore.userInfo?.role || '')
const selfUserId = computed(() => String(userStore.userInfo?.id || ''))
const isStudent = computed(() => role.value === 'student')

const backPath = computed(() => {
  if (role.value === 'enterprise') return '/enterprise/interview-workbench'
  if (role.value === 'university') return '/university/interview-workbench'
  return '/interview/live/workbench'
})

const hasRoom = computed(() => Boolean(roomId.value))

const waitingStatus = computed(() => {
  if (!hasRoom.value) return '未加入房间'
  if (groupStarted.value) return '进行中'
  if (signalSocket?.readyState === WebSocket.OPEN) return '等待开考'
  return '连接中'
})

const roomMembers = computed(() => {
  return [...members.value].sort((a, b) => (a.isSelf === b.isSelf ? 0 : a.isSelf ? -1 : 1))
})

const remoteMembers = computed(() => roomMembers.value.filter((item) => !item.isSelf).slice(0, 3))

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

const canVoteStart = computed(() => {
  if (!hasRoom.value || !signalSocket || signalSocket.readyState !== WebSocket.OPEN) return false
  if (groupStarted.value) return false
  return true
})

const normalizedInvitationStatus = computed(() => String(invitation.value?.status || '').trim().toLowerCase())

const isInvitationInitiator = computed(() => {
  const invitationId = Number(invitation.value?.initiator_user_id || invitation.value?.student_id || 0)
  const currentUserId = Number(selfUserId.value || 0)
  return invitationId > 0 && currentUserId > 0 && invitationId === currentUserId
})

const canStartInterview = computed(() => {
  if (!hasRoom.value || !isInvitationInitiator.value || startingInterview.value) return false
  if (groupStarted.value) return false
  return normalizedInvitationStatus.value === 'pending' || normalizedInvitationStatus.value === 'accepted'
})

function goBack() {
  router.push(backPath.value)
}

function resolveInvitationIdFromRoute() {
  const fromQuery = Number(route.query?.invitation_id || 0)
  return fromQuery > 0 ? fromQuery : 0
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
  return userStore.userInfo?.username || (isStudent.value ? '候选人' : '面试官')
}

function upsertMember(payload = {}) {
  const id = String(payload.userId || payload.user_id || '')
  if (!id) return
  const displayName = String(payload.senderName || payload.sender_name || payload.username || '').trim() || `成员 ${id}`
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
    const [sentRes, receivedRes] = await Promise.all([
      getHumanInvitations(),
      getReceivedHumanInvitations()
    ])
    const sent = Array.isArray(sentRes?.invitations) ? sentRes.invitations : []
    const received = Array.isArray(receivedRes?.invitations) ? receivedRes.invitations : []
    list = [...sent, ...received]
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

  const currentUserID = Number(selfUserId.value || 0)
  const initiatorID = Number(invitation.value?.initiator_user_id || invitation.value?.student_id || 0)
  if (currentUserID <= 0 || initiatorID <= 0 || currentUserID !== initiatorID) {
    return
  }

  const payload = {
    position: invitation.value?.position || '群面模拟',
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

  localStream = await navigator.mediaDevices.getUserMedia({
    video: true,
    audio: true
  })

  // 群面阶段强制挂载视频流，不允许关闭视频轨。
  localStream.getVideoTracks().forEach((track) => {
    track.enabled = true
  })
  localStream.getAudioTracks().forEach((track) => {
    track.enabled = micOn.value
  })

  bindLocalStreamToVideo()
}

function bindLocalStreamToVideo() {
  if (!localVideoRef.value || !localStream) return
  localVideoRef.value.srcObject = localStream
  localVideoRef.value.play?.().catch(() => {})
}

function bindRemoteStreamToSlots() {
  const refs = [remoteVideoRefA.value, remoteVideoRefB.value, remoteVideoRefC.value]
  refs.forEach((target) => {
    if (!target) return
    target.srcObject = remoteStream || null
    target.play?.().catch(() => {})
  })
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
    const [stream] = event.streams
    if (!stream) return
    remoteStream = stream
    bindRemoteStreamToSlots()
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
  if (!signalSocket || signalSocket.readyState !== WebSocket.OPEN) {
    ElMessage.warning('信令未连接，消息发送失败')
    return
  }

  sendSignal('chat', {
    text: content,
    sender_name: getSelfDisplayName(),
    role: role.value,
    interview_id: interviewId.value
  })
  messageInput.value = ''
}

function sendGroupInvite() {
  if (!signalSocket || signalSocket.readyState !== WebSocket.OPEN) {
    ElMessage.warning('信令未连接，无法发起邀请')
    return
  }

  sendSignal('group_invite', {
    sender_name: getSelfDisplayName(),
    role: role.value,
    target_participants: TARGET_PARTICIPANTS,
    start_threshold: TEST_START_THRESHOLD,
    interview_id: interviewId.value
  })
}

function voteGroupStart() {
  if (!canVoteStart.value) return
  sendSignal('group_start_vote', {
    sender_name: getSelfDisplayName(),
    role: role.value,
    target_participants: TARGET_PARTICIPANTS,
    start_threshold: TEST_START_THRESHOLD,
    interview_id: interviewId.value
  })
}

async function triggerStartInterview() {
  if (!invitation.value?.id || !canStartInterview.value) return

  startingInterview.value = true
  try {
    const res = await startLiveInterview({ invitation_id: Number(invitation.value.id) })
    const session = res?.session || {}

    const nextStatus = String(session?.status || '').trim().toLowerCase()
    if (nextStatus) {
      invitation.value = {
        ...invitation.value,
        status: nextStatus
      }
    }

    const syncedInterviewId = Number(session?.interview_id || 0)
    if (syncedInterviewId > 0) {
      interviewId.value = syncedInterviewId
      invitation.value = {
        ...invitation.value,
        interview_id: syncedInterviewId
      }
    }

    groupStarted.value = true
    statusText.value = '主控已开始面试'

    sendSignal('group_start', {
      message: '主控方已开始面试',
      interview_id: interviewId.value,
      started_at: session?.started_at || undefined
    })

    ElMessage.success('已开始面试')
  } catch (err) {
    const message = err?.response?.data?.error || err.message || '开始面试失败'
    ElMessage.error(message)
  } finally {
    startingInterview.value = false
  }
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

function parseIncomingData(data) {
  if (!data || typeof data !== 'object') return {}
  return data
}

async function handleSignalMessage(raw) {
  const msg = JSON.parse(raw)
  if (!msg?.type) return

  const senderID = String(msg.user_id || '')
  const data = parseIncomingData(msg.data)
  const isSelfSignal = senderID && senderID === selfUserId.value

  if (senderID) {
    upsertMember({
      userId: senderID,
      senderName: data.sender_name,
      role: data.role
    })
  }

  if (msg.type === 'chat') {
    appendMessage(
      'chat',
      data.display_text || data.text,
      isSelfSignal,
      data.sender_name || (isSelfSignal ? getSelfDisplayName() : '成员')
    )
    return
  }

  if (msg.type === 'group_invite') {
    groupStartThreshold.value = Number(data.start_threshold || TEST_START_THRESHOLD)
    groupTargetParticipants.value = Number(data.target_participants || TARGET_PARTICIPANTS)
    appendMessage(
      'system',
      `${data.sender_name || '系统'} 发起了群面邀请，目标 ${groupTargetParticipants.value} 人，测试开考阈值 ${groupStartThreshold.value} 人。`,
      false,
      '系统'
    )
    return
  }

  if (msg.type === 'group_start_status') {
    groupReadyCount.value = Number(data.ready_count || 0)
    groupStartThreshold.value = Number(data.start_threshold || TEST_START_THRESHOLD)
    groupTargetParticipants.value = Number(data.target_participants || TARGET_PARTICIPANTS)
    if (data.started) {
      groupStarted.value = true
      statusText.value = '群面已开始'
    }
    return
  }

  if (msg.type === 'group_start') {
    groupStarted.value = true
    statusText.value = String(data.message || '群面流程已开始')
    appendMessage('system', statusText.value, false, '系统')
    return
  }

  if (isSelfSignal && msg.type !== 'session_sync') {
    return
  }

  const pc = ensurePeer()

  if (msg.type === 'join') {
    statusText.value = '成员已进入房间'
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
    statusText.value = '成员已离开房间'
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
    statusText.value = '已进入房间，等待群面开始'

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

    sendGroupInvite()
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
    statusText.value = '信令已断开'
  }
}

function pickSupportedAudioMime() {
  const mimeCandidates = [
    'audio/webm;codecs=opus',
    'audio/webm',
    'audio/mp4',
    'audio/ogg;codecs=opus',
    'audio/ogg'
  ]
  for (const candidate of mimeCandidates) {
    if (window.MediaRecorder?.isTypeSupported?.(candidate)) {
      return candidate
    }
  }
  return ''
}

function blobToBase64Payload(blob) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => {
      const result = String(reader.result || '')
      const commaIndex = result.indexOf(',')
      if (commaIndex < 0) {
        reject(new Error('音频编码失败'))
        return
      }
      resolve(result.slice(commaIndex + 1))
    }
    reader.onerror = () => reject(new Error('音频读取失败'))
    reader.readAsDataURL(blob)
  })
}

async function pushRealtimeASRChunk(blob, durationSec = 3) {
  if (!micOn.value || !blob || blob.size === 0) return
  if (!interviewId.value || interviewId.value <= 0) return

  try {
    const payload = await blobToBase64Payload(blob)
    await analyzeSpeechChunk(interviewId.value, {
      audio_data: payload,
      audio_mime: blob.type || undefined,
      duration: durationSec,
      energy_level: 0.12,
      room_id: roomId.value,
      audio_enabled: micOn.value
    })
  } catch (err) {
    console.warn('realtime asr push failed', err)
  }
}

function startRealtimeASR() {
  if (!micOn.value || !localStream) return
  if (chunkRecorder) return

  const audioTracks = localStream.getAudioTracks()
  if (!audioTracks.length) return

  if (!window.MediaRecorder) {
    return
  }

  const selectedMime = pickSupportedAudioMime()

  try {
    chunkRecorderStream = new MediaStream([audioTracks[0].clone()])
    chunkRecorder = selectedMime
      ? new MediaRecorder(chunkRecorderStream, { mimeType: selectedMime })
      : new MediaRecorder(chunkRecorderStream)

    chunkRecorder.ondataavailable = async (event) => {
      if (!event.data || event.data.size === 0) return
      await pushRealtimeASRChunk(event.data, 3)
    }

    chunkRecorder.onstop = () => {
      chunkRecorderStream?.getTracks?.().forEach((track) => track.stop())
      chunkRecorderStream = null
      chunkRecorder = null
    }

    chunkRecorder.start(3000)
  } catch (err) {
    console.warn('failed to start realtime asr recorder', err)
    chunkRecorder = null
    if (chunkRecorderStream) {
      chunkRecorderStream.getTracks().forEach((track) => track.stop())
      chunkRecorderStream = null
    }
  }
}

function stopRealtimeASR() {
  if (chunkRecorder) {
    try {
      chunkRecorder.stop()
    } catch {
      // noop
    }
    chunkRecorder = null
  }
  if (chunkRecorderStream) {
    chunkRecorderStream.getTracks().forEach((track) => track.stop())
    chunkRecorderStream = null
  }
}

function toggleMic() {
  if (!localStream) return
  micOn.value = !micOn.value
}

async function finalizeInterviewAndReport() {
  if (!isStudent.value || interviewId.value <= 0) return
  await endInterview(interviewId.value)
  const reportRes = await generateReport({ interview_id: interviewId.value })
  const reportId = Number(reportRes?.report?.id || 0)
  if (reportId > 0) {
    ElMessage.success('群面报告已生成')
  }
}

function cleanup() {
  stopRealtimeASR()

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

  remoteStream = null
  if (localVideoRef.value) localVideoRef.value.srcObject = null
  if (remoteVideoRefA.value) remoteVideoRefA.value.srcObject = null
  if (remoteVideoRefB.value) remoteVideoRefB.value.srcObject = null
  if (remoteVideoRefC.value) remoteVideoRefC.value.srcObject = null

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
    groupStarted.value = false
    groupReadyCount.value = 0
    groupStartThreshold.value = TEST_START_THRESHOLD
    groupTargetParticipants.value = TARGET_PARTICIPANTS

    if (invitationID <= 0) {
      throw new Error('缺少 invitation_id，无法进入群面房间')
    }

    await loadInvitationByID(invitationID)
    await ensureInterviewSession()
    await authorizeJoin(invitationID)
    await initLocalMedia()
    connectSignalSocket()

    await nextTick()
    bindLocalStreamToVideo()
    bindRemoteStreamToSlots()

    startRealtimeASR()
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
    statusText.value = '仅支持受邀用户进入群面房间'
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

watch([remoteVideoRefA, remoteVideoRefB, remoteVideoRefC], () => {
  bindRemoteStreamToSlots()
})

watch(micOn, (enabled) => {
  if (!localStream) return
  localStream.getAudioTracks().forEach((track) => {
    track.enabled = enabled
  })

  if (enabled) {
    startRealtimeASR()
  } else {
    stopRealtimeASR()
  }
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
    <div class="px-4 py-4 md:px-8 md:py-6">
      <div class="max-w-425 mx-auto">
        <header class="flex items-start justify-between gap-4 mb-5">
          <div>
            <h1 class="text-2xl md:text-3xl font-semibold tracking-wide">群面实战房间</h1>
            <p class="text-sm text-slate-200/90 mt-1">{{ invitation?.position || '多人在线 AI / 真人混合群面' }}</p>
          </div>
          <button class="btn-ghost" @click="goBack">
            <ArrowLeft class="w-4 h-4" />
            返回
          </button>
        </header>

        <div v-if="!hasRoom && !loading" class="h-[calc(100vh-12rem)] flex items-center justify-center">
          <div class="glass-panel w-full max-w-xl p-8">
            <h2 class="text-2xl font-semibold">等待房间授权</h2>
            <p class="text-sm text-slate-100/80 mt-3">群面房间仅允许受邀用户通过 invitation_id + invitation_code 进入。</p>
            <button class="btn-primary w-full mt-6" @click="goBack">返回工作台</button>
          </div>
        </div>

        <div v-else class="grid grid-cols-1 xl:grid-cols-[minmax(0,1fr)_380px] gap-5 h-[calc(100vh-12rem)]">
          <section class="glass-panel p-4 md:p-5 flex flex-col min-h-0">
            <div v-if="loading" class="h-full flex items-center justify-center gap-2 text-slate-100/80">
              <Loader2 class="w-5 h-5 animate-spin" />
              <span>正在初始化设备与连接...</span>
            </div>

            <template v-else>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
                <div class="meta-card">
                  <p class="meta-title">房间号</p>
                  <p class="meta-content break-all">{{ roomId || '--' }}</p>
                  <button class="btn-subtle mt-3" @click="copyInviteLink">
                    <Copy class="w-4 h-4" />
                    复制邀请链接
                  </button>
                  <div class="mt-2 text-xs text-slate-100/70 flex items-start gap-1.5 break-all">
                    <Link2 class="w-3.5 h-3.5 mt-0.5" />
                    <span>{{ inviteLink }}</span>
                  </div>
                </div>

                <div class="meta-card">
                  <p class="meta-title">开考状态</p>
                  <p class="meta-content">{{ waitingStatus }}</p>
                  <p class="text-sm text-slate-100/80 mt-2">{{ statusText }}</p>
                  <div class="mt-3 text-xs text-slate-100/80">
                    开考投票：{{ groupReadyCount }} / {{ groupStartThreshold }}（目标容量 {{ groupTargetParticipants }} 人）
                  </div>
                </div>
              </div>

              <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mt-4 flex-1 min-h-0">
                <article class="video-tile">
                  <div class="video-head">我的画面 · {{ getSelfDisplayName() }}</div>
                  <div class="video-wrap">
                    <video v-if="true" ref="localVideoRef" autoplay playsinline muted class="video-el"></video>
                  </div>
                </article>

                <article class="video-tile">
                  <div class="video-head">{{ remoteMembers[0]?.displayName || '席位 2（待接入）' }}</div>
                  <div class="video-wrap">
                    <video v-if="true" ref="remoteVideoRefA" autoplay playsinline class="video-el"></video>
                    <div v-if="remoteMembers.length < 1" class="slot-mask">等待成员加入</div>
                  </div>
                </article>

                <article class="video-tile">
                  <div class="video-head">{{ remoteMembers[1]?.displayName || '席位 3（待接入）' }}</div>
                  <div class="video-wrap">
                    <video v-if="true" ref="remoteVideoRefB" autoplay playsinline class="video-el"></video>
                    <div v-if="remoteMembers.length < 2" class="slot-mask">等待成员加入</div>
                  </div>
                </article>

                <article class="video-tile">
                  <div class="video-head">{{ remoteMembers[2]?.displayName || '席位 4（待接入）' }}</div>
                  <div class="video-wrap">
                    <video v-if="true" ref="remoteVideoRefC" autoplay playsinline class="video-el"></video>
                    <div v-if="remoteMembers.length < 3" class="slot-mask">等待成员加入</div>
                  </div>
                </article>
              </div>

              <div class="mt-4 flex flex-wrap gap-3">
                <button class="btn-subtle" @click="sendGroupInvite">
                  <Users class="w-4 h-4" />
                  发起群面邀请
                </button>
                <button
                  v-if="isInvitationInitiator"
                  class="btn-primary"
                  :disabled="!canStartInterview"
                  @click="triggerStartInterview"
                >
                  {{ startingInterview ? '开始中...' : '主控开始面试' }}
                </button>
                <button class="btn-subtle" :disabled="!canVoteStart" @click="voteGroupStart">
                  发送准备信号（阈值 {{ groupStartThreshold }}）
                </button>
                <button class="btn-subtle" @click="toggleMic">
                  <Mic v-if="micOn" class="w-4 h-4" />
                  <MicOff v-else class="w-4 h-4" />
                  {{ micOn ? '语音开启中' : '语音关闭' }}
                </button>
                <button class="btn-danger ml-auto" :disabled="finishing" @click="leaveRoom">
                  <PhoneOff class="w-4 h-4" />
                  {{ finishing ? '正在结束...' : '结束并离开' }}
                </button>
              </div>
            </template>
          </section>

          <aside class="chat-sidebar glass-panel p-4 md:p-5 flex flex-col min-h-0">
            <div class="flex items-center justify-between mb-3">
              <h2 class="text-base font-semibold tracking-wide">公屏聊天</h2>
              <span class="text-xs text-slate-100/75">{{ roomMembers.length }} 人在线</span>
            </div>

            <div class="member-strip mb-3">
              <div v-for="member in roomMembers" :key="member.userId" class="member-chip">
                <Users class="w-3.5 h-3.5" />
                <span>{{ member.displayName }}</span>
              </div>
              <p v-if="roomMembers.length === 0" class="text-xs text-slate-100/70">暂无成员</p>
            </div>

            <div class="chat-scroll flex-1 min-h-0">
              <p v-if="messages.length === 0" class="text-sm text-slate-100/70 text-center py-10">等待消息中...</p>
              <div
                v-for="item in messages"
                :key="item.id"
                class="bubble"
                :class="item.fromSelf ? 'bubble-self' : (item.kind === 'system' ? 'bubble-system' : 'bubble-other')"
              >
                <div class="bubble-head">
                  <span>{{ item.senderName }}</span>
                  <span>{{ item.createdAt }}</span>
                </div>
                <p class="bubble-text">{{ item.text }}</p>
              </div>
            </div>

            <div class="mt-3 flex gap-2">
              <input
                v-model="messageInput"
                type="text"
                class="chat-input"
                placeholder="输入群面公屏消息"
                @keyup.enter="sendChatMessage"
              />
              <button class="btn-primary px-3" @click="sendChatMessage">
                <SendHorizontal class="w-4 h-4" />
              </button>
            </div>
          </aside>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.live-room-page {
  background:
    radial-gradient(circle at 20% 12%, rgba(56, 189, 248, 0.26), transparent 42%),
    radial-gradient(circle at 82% 18%, rgba(59, 130, 246, 0.3), transparent 38%),
    linear-gradient(150deg, #0b1220 0%, #12213d 52%, #0f1a2f 100%);
}

.glass-panel {
  background: linear-gradient(135deg, rgba(19, 35, 68, 0.56), rgba(25, 55, 92, 0.42));
  border: 1px solid rgba(148, 201, 255, 0.28);
  border-radius: 24px;
  backdrop-filter: blur(14px);
  box-shadow: 0 18px 40px rgba(8, 24, 54, 0.35);
}

.meta-card {
  border: 1px solid rgba(164, 211, 255, 0.24);
  background: rgba(13, 28, 53, 0.46);
  border-radius: 18px;
  padding: 14px;
}

.meta-title {
  font-size: 11px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: rgba(229, 242, 255, 0.7);
}

.meta-content {
  margin-top: 6px;
  font-size: 18px;
  font-weight: 600;
  color: #f3f9ff;
}

.video-tile {
  border: 1px solid rgba(166, 216, 255, 0.24);
  background: rgba(9, 19, 40, 0.55);
  border-radius: 18px;
  padding: 10px;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.video-head {
  font-size: 12px;
  color: rgba(228, 243, 255, 0.86);
  margin-bottom: 8px;
}

.video-wrap {
  position: relative;
  border-radius: 14px;
  overflow: hidden;
  border: 1px solid rgba(140, 183, 230, 0.24);
  background: rgba(7, 13, 26, 0.8);
  aspect-ratio: 16 / 9;
}

.video-el {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.slot-mask {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: rgba(235, 247, 255, 0.75);
  font-size: 13px;
  background: linear-gradient(135deg, rgba(8, 19, 37, 0.25), rgba(20, 45, 76, 0.45));
}

.chat-sidebar {
  background: linear-gradient(160deg, rgba(12, 26, 48, 0.64), rgba(20, 44, 77, 0.52));
}

.member-strip {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.member-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: rgba(237, 247, 255, 0.9);
  border: 1px solid rgba(167, 214, 255, 0.25);
  border-radius: 999px;
  padding: 4px 10px;
  background: rgba(25, 56, 96, 0.35);
}

.chat-scroll {
  overflow-y: auto;
  padding-right: 4px;
}

.bubble {
  border-radius: 16px;
  border: 1px solid transparent;
  padding: 10px 12px;
  margin-bottom: 10px;
  backdrop-filter: blur(8px);
}

.bubble-self {
  background: linear-gradient(135deg, rgba(76, 159, 255, 0.28), rgba(98, 174, 255, 0.18));
  border-color: rgba(156, 214, 255, 0.45);
}

.bubble-other {
  background: linear-gradient(135deg, rgba(34, 76, 128, 0.35), rgba(37, 86, 145, 0.24));
  border-color: rgba(142, 194, 247, 0.35);
}

.bubble-system {
  background: linear-gradient(135deg, rgba(62, 143, 228, 0.22), rgba(31, 82, 142, 0.25));
  border-color: rgba(146, 210, 255, 0.35);
}

.bubble-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 11px;
  color: rgba(230, 244, 255, 0.85);
  margin-bottom: 6px;
}

.bubble-text {
  font-size: 14px;
  line-height: 1.5;
  color: rgba(244, 250, 255, 0.96);
  word-break: break-word;
  white-space: pre-wrap;
}

.chat-input {
  flex: 1;
  border-radius: 14px;
  border: 1px solid rgba(159, 214, 255, 0.35);
  background: rgba(13, 30, 56, 0.66);
  color: #f4fbff;
  padding: 10px 12px;
  outline: none;
}

.chat-input:focus {
  border-color: rgba(184, 225, 255, 0.7);
}

.btn-primary,
.btn-subtle,
.btn-ghost,
.btn-danger {
  border-radius: 12px;
  padding: 9px 14px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-size: 13px;
  transition: all 0.2s ease;
}

.btn-primary {
  background: linear-gradient(135deg, rgba(95, 178, 255, 0.75), rgba(62, 144, 236, 0.8));
  border: 1px solid rgba(197, 232, 255, 0.5);
  color: #f8fcff;
}

.btn-primary:hover {
  filter: brightness(1.06);
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-subtle {
  background: rgba(31, 71, 118, 0.4);
  border: 1px solid rgba(162, 210, 255, 0.35);
  color: rgba(241, 249, 255, 0.95);
}

.btn-subtle:hover {
  background: rgba(39, 87, 143, 0.48);
}

.btn-ghost {
  background: rgba(19, 40, 73, 0.45);
  border: 1px solid rgba(158, 209, 255, 0.3);
  color: rgba(239, 248, 255, 0.95);
}

.btn-danger {
  background: rgba(198, 74, 95, 0.25);
  border: 1px solid rgba(255, 182, 198, 0.5);
  color: #ffeef2;
}

.btn-danger:hover {
  background: rgba(216, 84, 106, 0.35);
}

@media (max-width: 1280px) {
  .chat-sidebar {
    min-height: 360px;
  }
}
</style>
