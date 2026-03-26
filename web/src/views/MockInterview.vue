<script setup>
import { ref, onMounted, onUnmounted, nextTick, computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { startInterview as apiStartInterview, endInterview as apiEndInterview, getShadowCoachHint as apiGetShadowCoachHint, getInterviewConfig as apiGetInterviewConfig, revealRandomStyle as apiRevealRandomStyle, synthesizeInterviewSpeech as apiSynthesizeInterviewSpeech, getInviteCandidates as apiGetInviteCandidates, createHumanInvitation as apiCreateHumanInvitation, getHumanInvitations as apiGetHumanInvitations } from '../api/interview'
import { generateReport as apiGenerateReport } from '../api/report'
import InterviewContainer from '../components/InterviewContainer.vue'
import HumanInterviewModals from '../components/HumanInterviewModals.vue'
import { useMediaDevices } from '../composables/useMediaDevices'
import { useInterviewConfig } from '../composables/useInterviewConfig'
import { useInterviewChat } from '../composables/useInterviewChat'
import { useBlindBox } from '../composables/useBlindBox'
import { useInterviewRecording } from '../composables/useInterviewRecording'
import { useInterviewStore } from '../stores/useInterviewStore'

const route = useRoute()
const router = useRouter()
const {
  phase,
  settings,
  startInterview: enterInterviewPhase,
  endInterview: exitInterviewPhase
} = useInterviewConfig({ routeQuery: route.query })

let getChatAdditionalStreams = () => []
let stopChatSpeechAnalysis = () => {}

const interviewStore = useInterviewStore()
const {
  interviewId,
  questions,
  currentQuestion,
  shadowCoachEnabled,
  shadowCoachHints,
  shadowCoachBubbleText,
  shadowCoachBubbleVisible,
  shadowCoachHintPending,
  thinkingStreakSeconds
} = storeToRefs(interviewStore)
const {
  setQuietSeconds,
  hideShadowBubble,
  resetShadowHintProgress,
  preloadShadowHintPack,
  maybeDispatchTieredHint,
  startThinkingWatch,
  cleanupStoreTimers
} = interviewStore

const {
  isCameraOn,
  isMicOn,
  stream,
  toggleCamera,
  toggleMic,
  startCamera,
  stopCamera
} = useMediaDevices({
  getAdditionalStreams: () => getChatAdditionalStreams(),
  onCameraStopped: () => {
    stopChatSpeechAnalysis()
  }
})

// Interview State
const reportId = ref(null)
const isGeneratingReport = ref(false)
const isAvatarSpeaking = ref(false)
let currentSpeechAudio = null

const {
  blindBoxScenario,
  blindBoxRevealing,
  blindBoxRevealed,
  questionTimeLimit,
  questionTimer,
  pressureLevel,
  pressureColors,
  pressureLabels,
  drawBlindBox,
  reDrawBlindBox,
  startQuestionTimer,
  stopQuestionTimer
} = useBlindBox()

const {
  recordingStatus,
  startInterviewRecording,
  stopAndUploadInterviewRecording,
  cleanupInterviewRecording
} = useInterviewRecording({
  interviewId,
  settings,
  stream,
  isMicOn
})

const {
  messages,
  userInput,
  isProcessing,
  processingHint,
  currentQuestionIndex,
  pendingNextQuestion,
  pendingEnd,
  latestAIMessage,
  latestUserTranscript,
  canAnswerCurrentQuestion,
  answerVoiceStatus,
  answerVoiceSeconds,
  answerVoiceError,
  speechMetrics,
  energyLevel,
  speechAnalysisActive,
  appendMessage,
  replaceMessages,
  setCurrentQuestionIndex,
  incrementCurrentQuestionIndex,
  setPendingState,
  resetConversationState,
  sendMessage,
  startAnswerRecording,
  stopAnswerRecording,
  toggleAnswerRecording,
  getVoiceStatusLabel,
  stopSpeechAnalysis,
  getAdditionalStreams,
  cleanupInterviewChat
} = useInterviewChat({
  phase,
  settings,
  interviewId,
  currentQuestion,
  isAvatarSpeaking,
  isMicOn,
  stream,
  onAdvanceToNextQuestion: () => advanceToNextQuestion(),
  onCompleteInterview: () => completeInterview(),
  onScrollToBottom: () => scrollToBottom(),
  onResetQuietSeconds: () => {
    setQuietSeconds(0)
  }
})

getChatAdditionalStreams = getAdditionalStreams
stopChatSpeechAnalysis = stopSpeechAnalysis

const isAlgorithmStyle = computed(() => settings.value.style === 'algorithm')

// Interview Config from server
const interviewConfig = ref(null)

// Human Interview Invitation state
const inviteCandidates = ref([])
const inviteCandidatesLoading = ref(false)
const selectedInvitee = ref(null)
const showBookingDialog = ref(false)
const bookingForm = ref({ scheduledAt: '', notes: '' })
const userInvitations = ref([])
const showBookingsPanel = ref(false)
const activeInvitationId = ref(null)
const activeInvitation = ref(null)

// Random mode reveal state
const randomStyleRevealed = ref(false)
const revealedStyleInfo = ref(null)

// AI Shadow Coach
const modelViewerReady = ref(false)

const ensureModelViewerScript = () => {
  if (window.customElements && window.customElements.get('model-viewer')) {
    modelViewerReady.value = true
    return
  }

  const existing = document.getElementById('model-viewer-script')
  if (existing) {
    existing.addEventListener('load', () => {
      modelViewerReady.value = !!(window.customElements && window.customElements.get('model-viewer'))
    }, { once: true })
    return
  }

  const script = document.createElement('script')
  script.id = 'model-viewer-script'
  script.type = 'module'
  script.src = 'https://unpkg.com/@google/model-viewer/dist/model-viewer.min.js'
  script.onload = () => {
    modelViewerReady.value = !!(window.customElements && window.customElements.get('model-viewer'))
  }
  script.onerror = () => {
    modelViewerReady.value = false
    console.error('model-viewer script failed to load')
  }
  document.head.appendChild(script)
}

const stopAISpeech = () => {
  if (currentSpeechAudio) {
    currentSpeechAudio.pause()
    currentSpeechAudio = null
  }
  isAvatarSpeaking.value = false
}

const speakAIText = async (text) => {
  if (settings.value.interviewMode === 'human') return
  if (settings.value.presentationMode !== 'video_avatar') return
  if (!interviewId.value || !text) return

  try {
    const res = await apiSynthesizeInterviewSpeech(interviewId.value, { text })
    if (!res?.audio_base64) return
    stopAISpeech()
    const audio = new Audio(`data:audio/mpeg;base64,${res.audio_base64}`)
    currentSpeechAudio = audio
    isAvatarSpeaking.value = true
    audio.onended = () => {
      isAvatarSpeaking.value = false
      currentSpeechAudio = null
    }
    audio.onerror = () => {
      isAvatarSpeaking.value = false
      currentSpeechAudio = null
    }
    await audio.play()
  } catch (err) {
    isAvatarSpeaking.value = false
    console.warn('TTS playback failed:', err)
  }
}

const setPresentationMode = async (mode) => {
  if (mode === 'video_avatar' && settings.value.interviewMode === 'human') {
    settings.value.presentationMode = 'text_voice'
    stopAISpeech()
    stopCamera()
    ElMessage.warning('真人面试模式不展示 AI 虚拟面试官，请使用文字语音模式或进入实时面试间')
    return
  }

  settings.value.presentationMode = mode
  if (mode === 'video_avatar') {
    await startCamera()
    return
  }
  stopAISpeech()
  stopCamera()
}

const setInterviewMode = (mode) => {
  settings.value.interviewMode = mode
  if (mode === 'human') {
    if (settings.value.presentationMode === 'video_avatar') {
      settings.value.presentationMode = 'text_voice'
      stopAISpeech()
      stopCamera()
    }
    loadInviteCandidates()
    loadUserInvitations()
  }
}

// Interview Logic
const beginInterview = async () => {
  isProcessing.value = true
  processingHint.value = '正在初始化面试场景...'
  answerVoiceStatus.value = 'idle'
  answerVoiceError.value = ''
  answerVoiceSeconds.value = 0
  try {
    const startPayload = {
      position: settings.value.position,
      difficulty: settings.value.difficulty,
      mode: settings.value.mode,
      style: settings.value.style,
      company: settings.value.company,
      interview_mode: settings.value.interviewMode
    }
    if (settings.value.interviewMode === 'human') {
      if (!activeInvitationId.value) {
        throw new Error('请先选择一个已邀请对象')
      }
      startPayload.invitation_id = activeInvitationId.value
    }

    const res = await apiStartInterview({
      ...startPayload
    })
    
    // Backend returns { message: "...", interview: { ... } }
    // The interview object contains questions array if loaded correctly
    const interview = res.interview
    interviewId.value = interview.id
    if (settings.value.interviewMode === 'human') {
      activeInvitation.value = {
        ...activeInvitation.value,
        invitee: {
          ...(activeInvitation.value?.invitee || {}),
          username: interview.human_interviewer_name || activeInvitation.value?.invitee?.username
        },
        invitee_role: interview.human_interviewer_role || activeInvitation.value?.invitee_role
      }
    }
    setPendingState({ nextQuestion: null, isEnd: false })
    currentQuestion.value = null

    // Parse blindbox scenario if present
    if (interview.scenario) {
      try {
        blindBoxScenario.value = typeof interview.scenario === 'string'
          ? JSON.parse(interview.scenario)
          : interview.scenario
        blindBoxRevealed.value = true
        // Set time limit from scenario
        if (blindBoxScenario.value?.time_limit) {
          questionTimeLimit.value = blindBoxScenario.value.time_limit
        }
      } catch (_) { /* ignore parse errors */ }
    }

    questions.value = mapInterviewQuestions(interview.questions || [])
    
    if (questions.value.length === 0) {
      alert("未获取到面试题目，请重试")
      isProcessing.value = false
      return
    }

    // Switch to interview phase
    enterInterviewPhase()
    setCurrentQuestionIndex(0)
    currentQuestion.value = questions.value[0] || null
    
    // Initialize Chat — adapt greeting for different modes
    const isBlindBox = settings.value.mode === 'blindbox' && blindBoxScenario.value
    const isRandom = settings.value.interviewMode === 'random'
    const isHuman = settings.value.interviewMode === 'human'
    
    const modeLabels = { technical: '技术', hr: 'HR', comprehensive: '综合' }
    const styleLabels = { gentle: '温和型', stress: '压力型', deep: '技术深挖型', practical: '项目实战型', algorithm: '算法考察型' }
    const companyLabels = { ali: '阿里巴巴', bytedance: '字节跳动', tencent: '腾讯', meituan: '美团', baidu: '百度' }

    let scenarioGreeting
    if (isBlindBox) {
      scenarioGreeting = `🎲 盲盒场景已揭晓：${blindBoxScenario.value.icon} **${blindBoxScenario.value.name}**\n\n${blindBoxScenario.value.description}\n\n压力等级：${pressureLabels[blindBoxScenario.value.pressure] || '未知'}${blindBoxScenario.value.time_limit ? `\n每题限时：${blindBoxScenario.value.time_limit}秒` : ''}\n\n准备好了吗？让我们开始！`
    } else if (isHuman) {
      const interviewerName = interview.human_interviewer_name || activeInvitation.value?.invitee?.username || '真人面试官'
      const interviewerRole = interview.human_interviewer_role === 'university' ? '高校端' : interview.human_interviewer_role === 'enterprise' ? '企业端' : '协作方'
      scenarioGreeting = `已进入真人模拟面试（${interviewerRole}：${interviewerName}）。\n\n本场采用“破冰 -> 技术深挖 -> 场景追问 -> 反问总结”流程。\n系统将负责记录答题、语音监测与实时建议，不再使用虚拟AI面试官形象。`
    } else if (isRandom) {
      scenarioGreeting = `🎲 随机模式已启动！\n\n系统已为您随机分配了面试官风格，在面试过程中不会提前告知。\n这模拟了真实企业面试中的"突然切换风格"场景，请保持灵活应变！\n\n面试岗位：${settings.value.position}\n面试结束后将揭晓面试官风格，让我们开始吧！`
    } else {
      const modeLabel = modeLabels[settings.value.mode] || settings.value.mode
      const companyInfo = settings.value.company ? `（${companyLabels[settings.value.company] || settings.value.company}风格）` : ''
      scenarioGreeting = `你好！我是你的 AI 面试官${companyInfo}。我们将进行一场关于 ${settings.value.position} 的${modeLabel}面试，采用${styleLabels[settings.value.style] || settings.value.style}提问方式。准备好了吗？让我们开始吧。`
    }

    replaceMessages([
      {
        role: 'ai',
        content: scenarioGreeting,
        type: isBlindBox ? 'scenario' : undefined
      }
    ])
    
    // Push first question after a short delay. Algorithm style uses dedicated coding panel, so skip chat question push.
    if (!isAlgorithmStyle.value) {
      processingHint.value = isHuman ? '正在准备真人面试流程首题...' : '面试官正在组织首个话题...'
      setTimeout(() => {
        pushAIQuestion(currentQuestion.value)
        // Start question timer if scenario has time limit
        if (blindBoxScenario.value?.time_limit) {
          startQuestionTimer(blindBoxScenario.value.time_limit)
        }
        scrollToBottom()
      }, 1000)
    }

    // Handle video transition
    if (settings.value.presentationMode === 'video_avatar' && isCameraOn.value && settings.value.interviewMode !== 'human') {
      // Small delay to ensure DOM is ready
      setTimeout(async () => {
        if (!stream.value) await startCamera()
        startInterviewRecording()
      }, 500)
    }

  } catch (error) {
    console.error('Failed to start interview:', error)
    alert('启动面试失败: ' + (error.response?.data?.error || error.message))
  } finally {
    isProcessing.value = false
    processingHint.value = ''
  }
}

const onAlgorithmFinished = ({ total = 0, passed = 0, skipped = 0 } = {}) => {
  appendMessage({
    role: 'ai',
    type: 'system',
    content: `算法考察完成：共 ${total} 题，通过 ${passed} 题，跳过 ${skipped} 题。正在为你生成面试报告...`
  })
  completeInterview()
}

const pushAIQuestion = (question) => {
  const text = (question?.content || question?.title || '').trim()
  if (!text) return
  resetShadowHintProgress()
  appendMessage({
    role: 'ai',
    content: text,
    type: 'question'
  })
  if (settings.value.interviewMode !== 'human') {
    speakAIText(text)
  }
  preloadShadowHintPack({
    apiGetShadowCoachHint,
    latestUserTranscript
  })
}

const mapInterviewQuestions = (rawQuestions) => {
  return (rawQuestions || [])
    .map((item) => {
      const nested = item.question || {}
      return {
        mapId: item.id,
        questionId: item.question_id || nested.id,
        title: nested.title || item.title || '',
        content: nested.content || item.content || '',
        expectedAnswer: nested.expected_answer || item.expected_answer || ''
      }
    })
    .filter((q) => q.questionId && (q.content || q.title))
}

const advanceToNextQuestion = () => {
  if (answerVoiceStatus.value === 'recording') {
    stopAnswerRecording()
    return
  }
  if (pendingEnd.value) {
    stopQuestionTimer()
    stopAISpeech()
    appendMessage({
      role: 'ai',
      content: "面试结束！辛苦了。您可以点击下方按钮查看详细报告。",
      type: 'system'
    })
    if (settings.value.interviewMode === 'random') {
      revealStyle()
    }
    setPendingState({ nextQuestion: null, isEnd: false })
    completeInterview()
    scrollToBottom()
    return
  }

  if (pendingNextQuestion.value) {
    currentQuestion.value = pendingNextQuestion.value
    incrementCurrentQuestionIndex()
    thinkingStreakSeconds.value = 0
    pushAIQuestion(currentQuestion.value)
    if (blindBoxScenario.value?.time_limit) {
      startQuestionTimer(blindBoxScenario.value.time_limit)
    }
  }
  setPendingState({ nextQuestion: null, isEnd: false })
  scrollToBottom()
}

const completeInterview = async () => {
  if (isGeneratingReport.value || !interviewId.value) return
  isGeneratingReport.value = true
  stopAISpeech()
  try {
    const replayUploaded = await stopAndUploadInterviewRecording()
    await apiEndInterview(interviewId.value)
    if (settings.value.interviewMode === 'human') {
      await loadUserInvitations()
    }
    const reportRes = await apiGenerateReport({
      interview_id: interviewId.value
    })
    if (reportRes?.report?.id) {
      reportId.value = reportRes.report.id
    }
    if (!reportId.value) {
      appendMessage({
        role: 'system',
        content: '报告生成中，请稍后点击“查看面试报告”。',
        type: 'system'
      })
    }
    if (!replayUploaded) {
      ElMessage.warning('本次回放上传失败，报告可能不含回放视频')
    }
  } catch (error) {
    console.error('Failed to end interview:', error)
    const errMsg = error?.response?.data?.error || error?.message || '未知错误'
    appendMessage({
      role: 'system',
      content: `报告生成失败：${errMsg}`,
      type: 'system'
    })
  } finally {
    isGeneratingReport.value = false
    scrollToBottom()
  }
}

const viewReport = async () => {
  if (!reportId.value && interviewId.value) {
    try {
      const reportRes = await apiGenerateReport({
        interview_id: interviewId.value
      })
      if (reportRes?.report?.id) {
        reportId.value = reportRes.report.id
      }
    } catch (error) {
      const errMsg = error?.response?.data?.error || error?.message || '未知错误'
      appendMessage({
        role: 'system',
        content: `获取报告失败：${errMsg}`,
        type: 'system'
      })
      scrollToBottom()
      return
    }
  }
  if (reportId.value) {
    router.push(`/student/report/${reportId.value}`)
    return
  }
  appendMessage({
    role: 'system',
    content: '报告暂未生成完成，请稍后再试。',
    type: 'system'
  })
  scrollToBottom()
}

const scrollToBottom = () => {
  nextTick(() => {
    const container = document.getElementById('chat-container')
    if (container) container.scrollTop = container.scrollHeight
  })
}

const endInterviewEarly = async () => {
  if (confirm('确定要结束面试吗？进度将不会保存。')) {
    answerVoiceStatus.value = 'idle'
    answerVoiceError.value = ''
    answerVoiceSeconds.value = 0
    await stopAndUploadInterviewRecording()
    stopCamera()
    stopQuestionTimer()
    stopAISpeech()
    exitInterviewPhase()
    setCurrentQuestionIndex(0)
    currentQuestion.value = null
    resetConversationState()
    blindBoxScenario.value = null
    blindBoxRevealed.value = false
    randomStyleRevealed.value = false
    revealedStyleInfo.value = null
    if (settings.value.interviewMode === 'human') {
      loadUserInvitations()
    }
    resetShadowHintProgress()
    hideShadowBubble()
    if (interviewId.value) {
        try { await apiEndInterview(interviewId.value) } catch(e){}
    }
  }
}

// ===== Load Interview Config =====
const loadInterviewConfig = async () => {
  try {
    const res = await apiGetInterviewConfig()
    interviewConfig.value = res
  } catch (err) {
    console.warn('Failed to load interview config:', err)
  }
}

// ===== Human Interview Invitation Functions =====
const normalizeCandidateRole = (role) => {
  if (role === 'university') return '高校端'
  if (role === 'enterprise') return '企业端'
  return '协作方'
}

const loadInviteCandidates = async (roleFilter = '') => {
  inviteCandidatesLoading.value = true
  try {
    const role = roleFilter === 'campus' ? 'university' : roleFilter
    const res = await apiGetInviteCandidates({ role, page: 1, page_size: 50 })
    inviteCandidates.value = res.users || []
  } catch (err) {
    console.warn('Failed to load invite candidates:', err)
    inviteCandidates.value = []
  } finally {
    inviteCandidatesLoading.value = false
  }
}

const selectInviteCandidate = (candidate) => {
  selectedInvitee.value = candidate
  showBookingDialog.value = true
}

const submitBooking = async () => {
  if (!selectedInvitee.value) return
  try {
    const payload = {
      invitee_user_id: selectedInvitee.value.id,
      position: settings.value.position,
      difficulty: settings.value.difficulty,
      mode: settings.value.mode,
      style: settings.value.style,
      company: settings.value.company,
      notes: bookingForm.value.notes
    }
    if (bookingForm.value.scheduledAt) {
      payload.scheduled_at = new Date(bookingForm.value.scheduledAt).toISOString()
    }
    await apiCreateHumanInvitation(payload)
    alert('邀请已发送，已加入你的真人面试列表。')
    showBookingDialog.value = false
    bookingForm.value = { scheduledAt: '', notes: '' }
    selectedInvitee.value = null
    await loadUserInvitations()
  } catch (err) {
    alert('发送邀请失败：' + (err.response?.data?.error || err.message))
  }
}

const loadUserInvitations = async () => {
  try {
    const res = await apiGetHumanInvitations()
    userInvitations.value = res.invitations || []
    if (!activeInvitationId.value && userInvitations.value.length > 0) {
      const firstAvailable = userInvitations.value.find((i) => i.status === 'accepted' || i.status === 'in_progress')
      if (firstAvailable) {
        activeInvitationId.value = firstAvailable.id
        activeInvitation.value = firstAvailable
      }
    }
  } catch (err) {
    console.warn('Failed to load invitations:', err)
  }
}

const onOpenBookingsPanel = async () => {
  showBookingsPanel.value = true
  await loadUserInvitations()
}

const useInvitationForInterview = (invitation) => {
  activeInvitationId.value = invitation.id
  activeInvitation.value = invitation
  setInterviewMode('human')

  // Keep form selections aligned with the accepted invitation to avoid accidental AI-mode params.
  if (invitation.mode) settings.value.mode = invitation.mode
  if (invitation.style) settings.value.style = invitation.style
  if (invitation.position) settings.value.position = invitation.position
  if (invitation.difficulty) settings.value.difficulty = invitation.difficulty
  if (invitation.company) settings.value.company = invitation.company

  ElMessage.success('已切换到真人面试模式，请点击下方“开始面试”。')
}

const goLiveInterviewRoom = (invitation) => {
  if (!invitation?.id) {
    ElMessage.warning('邀请信息无效，请刷新后重试')
    return
  }
  useInvitationForInterview(invitation)
  router.push({
    path: '/student/live-interview',
    query: { invitation_id: String(invitation.id) }
  })
}

// ===== Random Mode Reveal =====
const revealStyle = async () => {
  if (!interviewId.value) return
  try {
    const res = await apiRevealRandomStyle(interviewId.value)
    revealedStyleInfo.value = res
    randomStyleRevealed.value = true
  } catch (err) {
    console.warn('Failed to reveal style:', err)
  }
}

const setupProps = computed(() => ({
  shadowCoachEnabled: shadowCoachEnabled.value,
  isCameraOn: isCameraOn.value,
  isMicOn: isMicOn.value,
  stream: stream.value,
  isProcessing: isProcessing.value,
  activeInvitationId: activeInvitationId.value,
  inviteCandidates: inviteCandidates.value,
  inviteCandidatesLoading: inviteCandidatesLoading.value,
  activeInvitation: activeInvitation.value,
  blindBoxRevealing: blindBoxRevealing.value,
  blindBoxRevealed: blindBoxRevealed.value,
  blindBoxScenario: blindBoxScenario.value,
  pressureColors,
  pressureLevel: pressureLevel.value,
  pressureLabels,
  normalizeCandidateRole
}))

const algorithmProps = computed(() => ({
  interviewId: interviewId.value,
  activeInvitation: activeInvitation.value,
  currentQuestionIndex: currentQuestionIndex.value,
  currentQuestion: currentQuestion.value,
  latestAiMessage: latestAIMessage.value,
  isProcessing: isProcessing.value,
  processingHint: processingHint.value,
  shadowCoachEnabled: shadowCoachEnabled.value,
  shadowCoachHints: shadowCoachHints.value,
  blindBoxScenario: blindBoxScenario.value,
  pressureColors,
  pressureLevel: pressureLevel.value,
  pressureLabels,
  questionTimer: questionTimer.value,
  normalizeCandidateRole
}))

const chatProps = computed(() => ({
  isCameraOn: isCameraOn.value,
  isMicOn: isMicOn.value,
  isAvatarSpeaking: isAvatarSpeaking.value,
  modelViewerReady: modelViewerReady.value,
  recordingStatus: recordingStatus.value,
  blindBoxScenario: blindBoxScenario.value,
  questionTimer: questionTimer.value,
  pressureLevel: pressureLevel.value,
  shadowCoachHintPending: shadowCoachHintPending.value,
  shadowCoachBubbleVisible: shadowCoachBubbleVisible.value,
  shadowCoachBubbleText: shadowCoachBubbleText.value,
  stream: stream.value,
  currentQuestionIndex: currentQuestionIndex.value,
  userInput: userInput.value,
  answerVoiceStatus: answerVoiceStatus.value,
  getVoiceStatusLabel,
  speechMetrics: speechMetrics.value,
  latestUserTranscript: latestUserTranscript.value,
  latestAiMessage: latestAIMessage.value,
  isProcessing: isProcessing.value,
  processingHint: processingHint.value,
  canAnswerCurrentQuestion: canAnswerCurrentQuestion.value,
  pendingEnd: pendingEnd.value,
  energyLevel: energyLevel.value,
  speechAnalysisActive: speechAnalysisActive.value,
  messages: messages.value,
  activeInvitation: activeInvitation.value,
  currentQuestion: currentQuestion.value,
  shadowCoachEnabled: shadowCoachEnabled.value,
  shadowCoachHints: shadowCoachHints.value,
  pressureColors,
  pressureLabels,
  normalizeCandidateRole,
  randomStyleRevealed: randomStyleRevealed.value,
  revealedStyleInfo: revealedStyleInfo.value
}))

const onUserInputUpdate = (nextValue) => {
  userInput.value = String(nextValue ?? '')
}

const onSettingsUpdate = (nextSettings) => {
  settings.value = nextSettings
}

const onShadowCoachEnabledUpdate = (enabled) => {
  shadowCoachEnabled.value = !!enabled
}

onMounted(() => {
  ensureModelViewerScript()
  if (settings.value.presentationMode === 'video_avatar') {
    startCamera()
  }
  loadInterviewConfig()
  startThinkingWatch({
    phase,
    isProcessing,
    latestAIMessage,
    canAnswerCurrentQuestion,
    answerVoiceStatus,
    energyLevel,
    onMaybeDispatch: () => maybeDispatchTieredHint({
      phase,
      isProcessing,
      latestAIMessage,
      apiGetShadowCoachHint,
      latestUserTranscript
    })
  })
})

onUnmounted(() => {
  document.body.style.overflow = ''
  cleanupInterviewRecording()
  cleanupInterviewChat()
  stopQuestionTimer()
  stopAISpeech()
  stopCamera()
  cleanupStoreTimers()
})
</script>

<template>
  <div class="min-h-[calc(100vh-8rem)] flex flex-col">
    <InterviewContainer
      :phase="phase"
      :settings="settings"
      :setup-props="setupProps"
      :algorithm-props="algorithmProps"
      :chat-props="chatProps"
      @update:settings="onSettingsUpdate"
      @update:shadow-coach-enabled="onShadowCoachEnabledUpdate"
      @toggle-mic="toggleMic"
      @toggle-camera="toggleCamera"
      @change-presentation-mode="setPresentationMode"
      @change-interview-mode="setInterviewMode"
      @load-invite-candidates="loadInviteCandidates"
      @select-invite-candidate="selectInviteCandidate"
      @open-bookings="onOpenBookingsPanel"
      @draw-blind-box="drawBlindBox"
      @redraw-blind-box="reDrawBlindBox"
      @start-interview="beginInterview"
      @view-report="viewReport"
      @complete-interview="onAlgorithmFinished"
      @update:user-input="onUserInputUpdate"
      @send-message="sendMessage"
      @toggle-answer-recording="toggleAnswerRecording"
      @close-random-reveal="randomStyleRevealed = false"
    />

    <HumanInterviewModals
      v-model:show-booking-dialog="showBookingDialog"
      v-model:show-bookings-panel="showBookingsPanel"
      v-model:booking-form="bookingForm"
      :selected-invitee="selectedInvitee"
      :user-invitations="userInvitations"
      :normalize-candidate-role="normalizeCandidateRole"
      @submit-booking="submitBooking"
      @go-live-room="goLiveInterviewRoom"
    />
  </div>
</template>

<style scoped>
/* Custom scrollbar for chat */
#chat-container::-webkit-scrollbar {
  width: 6px;
}
#chat-container::-webkit-scrollbar-track {
  background: transparent;
}
#chat-container::-webkit-scrollbar-thumb {
  background-color: #e4e4e7;
  border-radius: 20px;
}

.interview-room-scene {
  background-color: #0f172a;
}

.interview-room-scene::before {
  content: '';
  position: absolute;
  inset: 0;
  background:
    linear-gradient(180deg, rgba(2, 6, 23, 0.3) 0%, rgba(2, 6, 23, 0.62) 100%),
    linear-gradient(100deg, rgba(2, 132, 199, 0.14) 0%, rgba(2, 132, 199, 0) 52%),
    linear-gradient(180deg, rgba(15, 23, 42, 0) 64%, rgba(2, 6, 23, 0.55) 100%),
    url('/interview-room.jpg');
  background-size: cover;
  background-position: center;
  pointer-events: none;
}

.interview-room-scene::after {
  content: '';
  position: absolute;
  inset: 0;
  background:
    radial-gradient(56% 42% at 50% 82%, rgba(148, 163, 184, 0.18) 0%, rgba(148, 163, 184, 0) 78%),
    radial-gradient(38% 26% at 14% 14%, rgba(251, 191, 36, 0.12) 0%, rgba(251, 191, 36, 0) 70%);
  pointer-events: none;
}

.interview-room-scene--compact::after {
  opacity: 0.85;
}

.interview-room-desk {
  background: linear-gradient(90deg, rgba(120, 53, 15, 0.65) 0%, rgba(146, 64, 14, 0.62) 46%, rgba(120, 53, 15, 0.65) 100%);
  border: 1px solid rgba(253, 186, 116, 0.24);
  box-shadow: 0 0 34px rgba(15, 23, 42, 0.35);
}

.interview-room-wall {
  top: 18%;
  bottom: 14%;
  border: 1px solid rgba(148, 163, 184, 0.08);
  background: linear-gradient(180deg, rgba(15, 23, 42, 0.22) 0%, rgba(15, 23, 42, 0.08) 100%);
  opacity: 0.35;
}

.interview-room-wall--left {
  left: -8px;
  transform: perspective(420px) rotateY(18deg);
  transform-origin: left center;
  border-radius: 0 14px 14px 0;
}

.interview-room-wall--right {
  right: -10px;
  transform: perspective(420px) rotateY(-18deg);
  transform-origin: right center;
  border-radius: 14px 0 0 14px;
}

.interview-room-wall--compact {
  top: 24%;
  bottom: 20%;
}

.interview-room-floor {
  left: 8%;
  right: 8%;
  bottom: 6%;
  height: 24%;
  border-radius: 22px 22px 10px 10px;
  border: 1px solid rgba(56, 189, 248, 0.08);
  background:
    linear-gradient(180deg, rgba(15, 23, 42, 0) 0%, rgba(2, 132, 199, 0.05) 32%, rgba(15, 23, 42, 0.35) 100%),
    linear-gradient(100deg, rgba(30, 41, 59, 0.55) 0%, rgba(15, 23, 42, 0.7) 100%);
  transform: perspective(520px) rotateX(60deg);
  transform-origin: bottom center;
  box-shadow: 0 22px 40px rgba(2, 6, 23, 0.35);
  opacity: 0.45;
}

.interview-room-floor--compact {
  left: 12%;
  right: 12%;
  height: 22%;
}

/* Desktop: allow manual height drag on key interview panels. */
@media (min-width: 1024px) {
  .resizable-panel {
    resize: vertical;
    overflow: auto;
    min-height: 160px;
    max-height: 75vh;
  }

  .resizable-panel::-webkit-resizer {
    background: linear-gradient(135deg, #cbd5e1 0%, #94a3b8 100%);
    border-radius: 4px;
  }
}

  .interview-position-select {
    box-shadow: 0 1px 0 rgba(255, 255, 255, 0.8), 0 8px 16px rgba(15, 23, 42, 0.06);
  }

  .dropdown-fade-enter-active,
  .dropdown-fade-leave-active {
    transition: opacity 0.18s ease, transform 0.18s ease;
  }

  .dropdown-fade-enter-from,
  .dropdown-fade-leave-to {
    opacity: 0;
    transform: translateY(-6px);
  }

  .interviewer-stage {
    animation: interviewer-idle 5.2s ease-in-out infinite;
    transform-origin: center 78%;
  }

  .interviewer-static {
    pointer-events: none;
  }

  .interviewer-speaking {
    animation: interviewer-speaking 1.25s ease-in-out infinite;
  }

  .coach-bubble-enter-active,
  .coach-bubble-leave-active {
    transition: opacity 0.22s ease, transform 0.22s ease;
  }

  .coach-bubble-enter-from,
  .coach-bubble-leave-to {
    opacity: 0;
    transform: translateY(6px) scale(0.96);
  }

  @keyframes interviewer-idle {
    0% { transform: translateY(0px) rotate(0deg); }
    35% { transform: translateY(-1px) rotate(-0.2deg); }
    70% { transform: translateY(1px) rotate(0.2deg); }
    100% { transform: translateY(0px) rotate(0deg); }
  }

  @keyframes interviewer-speaking {
    0% { transform: translateY(0px) rotate(0deg) scale(1); }
    50% { transform: translateY(-1px) rotate(-0.2deg) scale(1.01); }
    100% { transform: translateY(0px) rotate(0deg) scale(1); }
  }
</style>
