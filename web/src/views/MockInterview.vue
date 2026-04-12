<script setup>
import { ref, onMounted, onBeforeUnmount, nextTick, computed, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useRoute, useRouter, onBeforeRouteLeave } from 'vue-router'
import { ElMessage } from 'element-plus'
import { startStandardInterview as apiStartStandardInterview, startAlgorithmInterview as apiStartAlgorithmInterview, endInterview as apiEndInterview, getShadowCoachHint as apiGetShadowCoachHint, getInterviewConfig as apiGetInterviewConfig, revealRandomStyle as apiRevealRandomStyle } from '../api/interview'
import { generateReport as apiGenerateReport } from '../api/report'
import InterviewContainer from '../components/InterviewContainer.vue'
import { useInterviewConfig } from '../composables/useInterviewConfig'
import { useInterviewChat } from '../composables/useInterviewChat'
import { useInterviewCore } from '../composables/useInterviewCore'
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
  stopCamera,
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
  stopQuestionTimer,
  recordingStatus,
  startInterviewRecording,
  stopAndUploadInterviewRecording,
  isAvatarSpeaking,
  stopAISpeech,
  speakAIText,
  registerManagedTimeout,
  initLoadingStageText,
  initLoadingStageIndex,
  initLoadingElapsedSeconds,
  initLoadingStageTotal,
  startInterviewInitLoadingFlow,
  stopInterviewInitLoadingFlow,
  markInterviewInitReady,
  isInterviewInitTimeoutError,
  showInterviewInitTimeoutDialog,
  cleanupInterviewCore,
  bindProcessingState
} = useInterviewCore({
  settings,
  interviewId,
  getAdditionalStreams: () => getChatAdditionalStreams(),
  onCameraStopped: () => {
    stopChatSpeechAnalysis()
  }
})

// Interview State
const reportId = ref(null)
const isGeneratingReport = ref(false)
const isSubmitting = ref(false)
const isFinishing = ref(false)
const reportNavigationLocked = ref(false)
const finishLoadingText = '正在为您深度生成面试报告与多维评分，请耐心等待...'

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
bindProcessingState({
  isProcessing,
  processingHint
})

const isAlgorithmStyle = computed(() => settings.value.style === 'algorithm')
const setupType = computed(() => route.path.includes('/interview/algorithm/setup') ? 'algorithm' : 'standard')

const syncSetupDefaultsByRoute = () => {
  if (setupType.value === 'algorithm') {
    settings.value = {
      ...settings.value,
      interviewMode: 'ai',
      mode: 'technical',
      style: 'algorithm'
    }
    return
  }
  settings.value = {
    ...settings.value,
    interviewMode: 'ai',
    style: settings.value.style === 'algorithm' ? 'gentle' : settings.value.style
  }
}

watch(() => route.path, () => {
  syncSetupDefaultsByRoute()
}, { immediate: true })

// Interview Config from server
const interviewConfig = ref(null)

const activeInvitationId = ref(null)
const activeInvitation = ref(null)

// Random mode reveal state
const randomStyleRevealed = ref(false)
const revealedStyleInfo = ref(null)

// AI Shadow Coach
const modelViewerReady = ref(false)
const mockInterviewRootRef = ref(null)

let modelViewerScript = null
let modelViewerLoadListener = null

const detachModelViewerLoadListener = () => {
  if (modelViewerScript && modelViewerLoadListener) {
    modelViewerScript.removeEventListener('load', modelViewerLoadListener)
  }
  modelViewerScript = null
  modelViewerLoadListener = null
}

const destroyModelViewerInstances = () => {
  const root = mockInterviewRootRef.value
  if (!root) return
  const viewers = root.querySelectorAll('model-viewer')
  viewers.forEach((viewer) => {
    try {
      viewer.pause?.()
      viewer.removeAttribute('src')
      viewer.load?.()
    } catch {
      // ignore model-viewer cleanup errors
    }
  })
}

const ensureModelViewerScript = () => {
  if (window.customElements && window.customElements.get('model-viewer')) {
    modelViewerReady.value = true
    return
  }

  detachModelViewerLoadListener()

  const existing = document.getElementById('model-viewer-script')
  if (existing) {
    modelViewerScript = existing
    modelViewerLoadListener = () => {
      modelViewerReady.value = !!(window.customElements && window.customElements.get('model-viewer'))
      detachModelViewerLoadListener()
    }
    existing.addEventListener('load', modelViewerLoadListener)
    return
  }

  const script = document.createElement('script')
  script.id = 'model-viewer-script'
  script.type = 'module'
  script.src = 'https://unpkg.com/@google/model-viewer/dist/model-viewer.min.js'
  modelViewerScript = script
  script.onload = () => {
    modelViewerReady.value = !!(window.customElements && window.customElements.get('model-viewer'))
    detachModelViewerLoadListener()
  }
  script.onerror = () => {
    modelViewerReady.value = false
    console.error('model-viewer script failed to load')
    detachModelViewerLoadListener()
  }
  document.head.appendChild(script)
}

// Interview Logic
const beginInterview = async (setupContext = {}) => {
  let shouldRetry = false
  let shouldShowTimeoutDialog = false
  isProcessing.value = true
  startInterviewInitLoadingFlow()
  reportId.value = null
  isSubmitting.value = false
  isFinishing.value = false
  reportNavigationLocked.value = false
  answerVoiceStatus.value = 'idle'
  answerVoiceError.value = ''
  answerVoiceSeconds.value = 0
  try {
    const initType = setupContext?.initType || setupType.value
    const algorithmConfig = setupContext?.algorithmConfig || {}
    const algorithmDifficultyMap = {
      easy: 'campus_intern',
      medium: 'campus_graduate',
      hard: 'social_junior'
    }

    const startPayload = {
      position: settings.value.position,
      difficulty: settings.value.difficulty,
      mode: settings.value.mode,
      style: settings.value.style,
      company: settings.value.company,
      interview_mode: settings.value.interviewMode
    }
    if (initType === 'algorithm') {
      startPayload.difficulty = algorithmDifficultyMap[algorithmConfig.difficulty] || settings.value.difficulty
      startPayload.mode = 'technical'
      startPayload.style = 'algorithm'
      startPayload.interview_mode = 'ai'
      startPayload.algorithm_difficulty = algorithmConfig.difficulty || 'medium'
      startPayload.algorithm_focus_tags = Array.isArray(algorithmConfig.focusTags) ? algorithmConfig.focusTags : []
      startPayload.algorithm_language = algorithmConfig.language || ''
    }
    if (settings.value.interviewMode === 'human') {
      if (!activeInvitationId.value) {
        throw new Error('请先选择一个已邀请对象')
      }
      startPayload.invitation_id = activeInvitationId.value
    }

    const starter = initType === 'algorithm' ? apiStartAlgorithmInterview : apiStartStandardInterview
    const res = await starter({ ...startPayload })
    
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
      ElMessage.warning('未获取到面试题目，请稍后重试')
      return
    }

    await markInterviewInitReady()

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
      registerManagedTimeout(() => {
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
      registerManagedTimeout(async () => {
        if (!stream.value) await startCamera()
        startInterviewRecording()
      }, 500)
    }

  } catch (error) {
    console.error('Failed to start interview:', error)
    if (isInterviewInitTimeoutError(error)) {
      shouldShowTimeoutDialog = true
      return
    }
    const errMsg = error?.response?.data?.error || error?.message || '未知错误'
    ElMessage.error(`启动面试失败：${errMsg}`)
  } finally {
    isProcessing.value = false
    processingHint.value = ''
    stopInterviewInitLoadingFlow()
    if (shouldShowTimeoutDialog) {
      shouldRetry = await showInterviewInitTimeoutDialog()
    }
    if (shouldRetry) {
      registerManagedTimeout(() => {
        beginInterview(setupContext)
      }, 120)
    }
  }
}

const onAlgorithmFinished = ({ total = 0, passed = 0, skipped = 0 } = {}) => {
  if (isFinishing.value) return
  appendMessage({
    role: 'ai',
    type: 'system',
    content: `算法考察完成：共 ${total} 题，通过 ${passed} 题，跳过 ${skipped} 题。正在为你生成面试报告...`
  })
  completeInterview()
}

const normalizeReportId = (value) => {
  const parsed = Number(value)
  if (!Number.isInteger(parsed) || parsed <= 0) return 0
  return parsed
}

const navigateToReportOnce = async () => {
  const validReportId = normalizeReportId(reportId.value)
  if (!validReportId) return false
  if (reportNavigationLocked.value) return true

  reportNavigationLocked.value = true
  try {
    await router.replace({
      name: 'Report',
      params: { id: validReportId }
    })
    return true
  } catch (err) {
    reportNavigationLocked.value = false
    throw err
  }
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
  if (isFinishing.value || !interviewId.value) return
  isFinishing.value = true
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
    const nextReportId = normalizeReportId(reportRes?.report?.id)
    if (nextReportId) {
      reportId.value = nextReportId
    }
    if (!normalizeReportId(reportId.value)) {
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
    isFinishing.value = false
    scrollToBottom()
  }
}

const viewReport = async () => {
  if (isFinishing.value || reportNavigationLocked.value) return

  isFinishing.value = true
  try {
    if (!normalizeReportId(reportId.value) && interviewId.value) {
      const reportRes = await apiGenerateReport({
        interview_id: interviewId.value
      })
      const nextReportId = normalizeReportId(reportRes?.report?.id)
      if (nextReportId) {
        reportId.value = nextReportId
      }
    }

    const didNavigate = await navigateToReportOnce()
    if (didNavigate) return

    appendMessage({
      role: 'system',
      content: '报告暂未生成完成，请稍后再试。',
      type: 'system'
    })
    scrollToBottom()
  } catch (error) {
    const errMsg = error?.response?.data?.error || error?.message || '未知错误'
    appendMessage({
      role: 'system',
      content: `获取报告失败：${errMsg}`,
      type: 'system'
    })
    scrollToBottom()
  } finally {
    isFinishing.value = false
  }
}

const scrollToBottom = () => {
  nextTick(() => {
    const container = document.getElementById('chat-container')
    if (container) container.scrollTop = container.scrollHeight
  })
}

const endInterviewEarly = async () => {
  if (isFinishing.value) return
  if (confirm('确定要结束面试吗？进度将不会保存。')) {
    isFinishing.value = true
    answerVoiceStatus.value = 'idle'
    answerVoiceError.value = ''
    answerVoiceSeconds.value = 0
    try {
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
    } finally {
      isFinishing.value = false
    }
  }
}

const onToggleAnswerRecording = async () => {
  if (isSubmitting.value || isFinishing.value) return
  isSubmitting.value = true
  try {
    await toggleAnswerRecording()
  } finally {
    isSubmitting.value = false
  }
}

const onSendMessage = async () => {
  if (isSubmitting.value || isFinishing.value) return
  isSubmitting.value = true
  try {
    await sendMessage()
  } finally {
    isSubmitting.value = false
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
  setupType: setupType.value,
  isCameraOn: isCameraOn.value,
  isMicOn: isMicOn.value,
  stream: stream.value,
  isProcessing: isProcessing.value,
  initLoadingStageText: initLoadingStageText.value,
  initLoadingStageIndex: initLoadingStageIndex.value,
  initLoadingStageTotal: initLoadingStageTotal.value,
  initLoadingElapsedSeconds: initLoadingElapsedSeconds.value
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
  isSubmitting: isSubmitting.value,
  isFinishing: isFinishing.value,
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

const cleanupInterviewPageSideEffects = () => {
  isSubmitting.value = false
  isFinishing.value = false
  reportNavigationLocked.value = false
  cleanupInterviewCore()
  detachModelViewerLoadListener()
  destroyModelViewerInstances()

  if (typeof document !== 'undefined' && document.body) {
    document.body.style.overflow = ''
  }

  cleanupInterviewChat()
  cleanupStoreTimers()
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

onBeforeRouteLeave((_to, _from, next) => {
  cleanupInterviewPageSideEffects()
  next()
})

onBeforeUnmount(() => {
  cleanupInterviewPageSideEffects()
})
</script>

<template>
  <div ref="mockInterviewRootRef" class="min-h-[calc(100vh-8rem)] flex flex-col">
    <InterviewContainer
      :phase="phase"
      :settings="settings"
      :setup-props="setupProps"
      :algorithm-props="algorithmProps"
      :chat-props="chatProps"
      @update:settings="onSettingsUpdate"
      @toggle-mic="toggleMic"
      @toggle-camera="toggleCamera"
      @start-interview="beginInterview"
      @view-report="viewReport"
      @complete-interview="onAlgorithmFinished"
      @update:user-input="onUserInputUpdate"
      @send-message="onSendMessage"
      @toggle-answer-recording="onToggleAnswerRecording"
      @close-random-reveal="randomStyleRevealed = false"
    />

    <div
      v-if="isFinishing"
      class="fixed inset-0 z-1200 bg-slate-950/45 backdrop-blur-sm flex items-center justify-center px-6"
    >
      <div class="w-full max-w-md rounded-3xl border border-white/30 bg-white/96 shadow-2xl p-8 text-center">
        <div class="h-10 w-10 mx-auto rounded-full border-4 border-indigo-200 border-t-indigo-600 animate-spin"></div>
        <p class="mt-5 text-base font-semibold text-zinc-900 leading-relaxed">{{ finishLoadingText }}</p>
        <p class="mt-2 text-xs text-zinc-500">期间请勿重复点击按钮，我们会在报告可用后自动为您继续。</p>
      </div>
    </div>

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
