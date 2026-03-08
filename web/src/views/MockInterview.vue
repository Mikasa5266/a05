<script setup>
import { ref, onMounted, onUnmounted, nextTick, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { 
  Video, VideoOff, Mic, MicOff, ChevronRight, ChevronDown,
  BrainCircuit, User, LogOut, Send, Loader2,
  History, MessageSquare, Lightbulb,
  Monitor, Users, Shuffle, UserCheck, Shield, Headphones,
  Heart, Eye, Brain, Volume2, BarChart3, CheckCircle, AlertTriangle, BookOpen,
  Package, Timer, Zap, Building2, Star, Calendar, Clock, X,
  Flame, Search, Code, Briefcase, GraduationCap
} from 'lucide-vue-next'
import { startInterview as apiStartInterview, getInterview as apiGetInterview, getInterviewSession as apiGetInterviewSession, submitAnswer as apiSubmitAnswer, endInterview as apiEndInterview, uploadInterviewRecording as apiUploadInterviewRecording, analyzeSpeechChunk as apiAnalyzeSpeechChunk, drawBlindBoxScenario as apiDrawBlindBox, getInterviewConfig as apiGetInterviewConfig, getHumanInterviewers as apiGetHumanInterviewers, bookHumanInterview as apiBookHumanInterview, getUserBookings as apiGetUserBookings, revealRandomStyle as apiRevealRandomStyle, generateTTS as apiGenerateTTS } from '../api/interview'
import { generateReport as apiGenerateReport } from '../api/report'
import WebSocketClient from '../utils/websocket'
import { useUserStore } from '../stores/user'
import SpeechDashboard from '../components/SpeechDashboard.vue'

const userStore = useUserStore()
let wsClient = null

const connectWebSocket = () => {
  if (wsClient) return
  const userId = userStore.userInfo?.id || 1
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = window.location.host
  // Assuming dev server proxies /ws to backend, or direct backend URL
  // If using Vite proxy, ws://localhost:5173/ws -> ws://localhost:8080/ws
  const wsUrl = `${protocol}//${host}/ws?user_id=${userId}&interview_id=${interviewId.value}`
  
  wsClient = new WebSocketClient(wsUrl)
  wsClient.connect((msg) => {
    handleWebSocketMessage(msg)
  })
}

const disconnectWebSocket = () => {
  if (wsClient) {
    wsClient.close()
    wsClient = null
  }
}

const handleWebSocketMessage = (msg) => {
  // console.log('WS Message:', msg)
  if (msg.type === 'interviewer.transcribing') {
     processingHint.value = '正在识别您的语音...'
     interviewerMotion.value = 'listening'
  } else if (msg.type === 'interviewer.thinking') {
     transitionSessionState('THINKING', 'interviewer.thinking')
     interviewerMotion.value = 'thinking'
  }
}


const route = useRoute()
const router = useRouter()
const phase = ref('setup') // setup, interview, summary
const isCameraOn = ref(true)
const isMicOn = ref(true)
const previewVideo = ref(null)
const interviewVideo = ref(null)
const stream = ref(null)
const recordingStatus = ref('idle')
const recordingUrl = ref('')
const answerVoiceStatus = ref('idle') // idle, requesting, recording, transcribing, submitting, success, error
const answerVoiceSeconds = ref(0)
const answerVoiceError = ref('')
let interviewMediaRecorder = null
let interviewRecordedChunks = []
let answerMediaRecorder = null
let answerAudioChunks = []
let answerVoiceTimer = null
let answerRecorderStream = null

// Interview State
const interviewId = ref(null)
const questions = ref([])
const currentQuestionIndex = ref(0)
const messages = ref([])
const userInput = ref('')
const isProcessing = ref(false)
const processingHint = ref('')
const reportId = ref(null)
const isGeneratingReport = ref(false)
const showHistory = ref(false)
const showModelAnswer = ref(false)
const pendingNextIndex = ref(null)
const pendingEnd = ref(false)
const sessionState = ref('INIT')
const sessionEvents = ref([])
const isOffline = ref(false)
let sessionHeartbeatTimer = null
let processingTimeoutTimer = null
const interviewerSpeaking = ref(false)
const interviewerMotion = ref('idle')
let interviewerMotionTimer = null
let currentAudio = null

const latestAIMessage = computed(() => {
  const aiMsgs = messages.value.filter(m => m.role === 'ai' || m.type === 'system')
  return aiMsgs.length > 0 ? aiMsgs[aiMsgs.length - 1] : null
})

const settings = ref({
  position: route.query.position || 'Java后端开发',
  difficulty: 'campus_intern',
  mode: route.query.mode || 'technical',
  style: 'gentle',
  company: '',
  interviewMode: 'ai',  // ai, human, random
  sceneMode: route.query.scene_mode === 'classic' ? 'classic' : 'video'
})

const isVideoSceneMode = computed(() => settings.value.sceneMode === 'video')

// Interview Config from server
const interviewConfig = ref(null)

// Human Interviewer state
const humanInterviewers = ref([])
const humanInterviewersLoading = ref(false)
const selectedInterviewer = ref(null)
const showBookingDialog = ref(false)
const bookingForm = ref({ scheduledAt: '', notes: '' })
const userBookings = ref([])
const showBookingsPanel = ref(false)

// Random mode reveal state
const randomStyleRevealed = ref(false)
const revealedStyleInfo = ref(null)

// AI Shadow Coach
const shadowCoachEnabled = ref(true)
const shadowCoachHints = ref([])
const emotionFeedback = ref({ sentiment: '正常', confidence: 0, heartRate: 72 })

// ===== Blind Box Mode =====
const blindBoxScenario = ref(null)       // The drawn scenario object
const blindBoxRevealing = ref(false)     // Whether the reveal animation is playing
const blindBoxRevealed = ref(false)      // Whether scenario has been revealed
const questionTimeLimit = ref(0)         // Per-question time limit in seconds
const questionTimer = ref(0)             // Current countdown
let questionTimerInterval = null

const pressureLevel = computed(() => blindBoxScenario.value?.pressure || 'low')
const isHighPressure = computed(() => ['high', 'extreme'].includes(pressureLevel.value))

const pressureColors = {
  low: { bg: 'bg-emerald-50', border: 'border-emerald-200', text: 'text-emerald-700', badge: 'bg-emerald-100 text-emerald-700' },
  medium: { bg: 'bg-amber-50', border: 'border-amber-200', text: 'text-amber-700', badge: 'bg-amber-100 text-amber-700' },
  high: { bg: 'bg-rose-50', border: 'border-rose-300', text: 'text-rose-700', badge: 'bg-rose-100 text-rose-700' },
  extreme: { bg: 'bg-red-50', border: 'border-red-400', text: 'text-red-800', badge: 'bg-red-200 text-red-800' },
}
const pressureLabels = { low: '轻松', medium: '中等', high: '高压', extreme: '极限' }

// Draw a blindbox scenario (preview before starting)
const drawBlindBox = async () => {
  blindBoxRevealing.value = true
  blindBoxRevealed.value = false
  blindBoxScenario.value = null

  try {
    const res = await apiDrawBlindBox()
    // Simulate slot-machine reveal delay
    await new Promise(resolve => setTimeout(resolve, 1500))
    blindBoxScenario.value = res.scenario
    blindBoxRevealed.value = true
  } catch (err) {
    console.error('Failed to draw blindbox:', err)
    alert('抽取场景失败：' + (err.response?.data?.error || err.message))
  } finally {
    blindBoxRevealing.value = false
  }
}

// Redraw a different scenario
const reDrawBlindBox = () => {
  blindBoxRevealed.value = false
  drawBlindBox()
}

// Start per-question timer (for timed scenarios)
const startQuestionTimer = (limitSec) => {
  stopQuestionTimer()
  if (!limitSec || limitSec <= 0) return
  questionTimeLimit.value = limitSec
  questionTimer.value = limitSec
  questionTimerInterval = setInterval(() => {
    questionTimer.value--
    if (questionTimer.value <= 0) {
      stopQuestionTimer()
    }
  }, 1000)
}

const stopQuestionTimer = () => {
  if (questionTimerInterval) {
    clearInterval(questionTimerInterval)
    questionTimerInterval = null
  }
  questionTimer.value = 0
}

// ===== Real-time Speech Metrics =====
const speechMetrics = ref({
  speechRate: 0,
  speechRateLevel: 'normal',
  fillerWordCount: 0,
  fluencyAlert: false,
  totalFillerWords: 0
})
const energyLevel = ref(0)
const speechAnalysisActive = ref(false)
const liveTranscript = ref('')
const latestChunkTranscript = ref('')
const speechLastUpdatedAt = ref(0)

const SPEECH_CHUNK_SECONDS = 2.0
const SPEECH_RATE_SMOOTH_ALPHA = 0.45
const SPEECH_RATE_IDLE_ALPHA = 0.2
const SPEECH_TRANSCRIPT_MAX_LEN = 240

// Audio chunk recording for speech analysis
let audioContext = null
let analyserNode = null
let chunkMediaRecorder = null
let chunkRecordingStream = null
let chunkInterval = null
let energyAnimFrame = null
let speechRequestSeq = 0
let speechAppliedSeq = 0

const startSpeechAnalysis = () => {
  if (speechAnalysisActive.value || !stream.value) return
  speechAnalysisActive.value = true
  liveTranscript.value = ''
  latestChunkTranscript.value = ''
  speechLastUpdatedAt.value = Date.now()
  speechRequestSeq = 0
  speechAppliedSeq = 0
  speechMetrics.value.speechRate = 0
  speechMetrics.value.speechRateLevel = 'normal'
  speechMetrics.value.fillerWordCount = 0
  speechMetrics.value.fluencyAlert = false
  speechMetrics.value.totalFillerWords = 0

  // Set up Web Audio API for real-time energy
  audioContext = new (window.AudioContext || window.webkitAudioContext)()
  const source = audioContext.createMediaStreamSource(stream.value)
  analyserNode = audioContext.createAnalyser()
  analyserNode.fftSize = 256
  source.connect(analyserNode)

  // Animate energy level
  const dataArray = new Uint8Array(analyserNode.frequencyBinCount)
  let silenceStartTime = null
  const VAD_THRESHOLD = 0.04
  const VAD_SILENCE_LIMIT = 2500 // 2.5s silence to trigger submit

  const updateEnergy = () => {
    if (!speechAnalysisActive.value) return
    analyserNode.getByteFrequencyData(dataArray)
    let sum = 0
    for (let i = 0; i < dataArray.length; i++) sum += dataArray[i]
    const avg = sum / dataArray.length / 255
    energyLevel.value = avg

    // VAD Logic for Video Mode
    if (isVideoSceneMode.value && answerVoiceStatus.value === 'recording' && answerVoiceSeconds.value > 2.5) {
       if (avg < VAD_THRESHOLD) {
          if (!silenceStartTime) silenceStartTime = Date.now()
          else if (Date.now() - silenceStartTime > VAD_SILENCE_LIMIT) {
             // Silence detected, auto submitting...
             stopAnswerRecording() 
             silenceStartTime = null
          }
       } else {
          silenceStartTime = null
       }
    } else {
       silenceStartTime = null
    }

    energyAnimFrame = requestAnimationFrame(updateEnergy)
  }
  updateEnergy()

  // Start chunked recording: every 4 seconds, capture a chunk and send for analysis
  startChunkRecording()
}

const startChunkRecording = () => {
  if (!stream.value) return

  const startNewChunk = () => {
    if (!speechAnalysisActive.value || !stream.value) return

    // Clone audio tracks for chunk recording
    const audioTracks = stream.value.getAudioTracks()
    if (audioTracks.length === 0) return
    chunkRecordingStream = new MediaStream(audioTracks)

    try {
      chunkMediaRecorder = new MediaRecorder(chunkRecordingStream, { mimeType: 'audio/webm' })
    } catch {
      chunkMediaRecorder = new MediaRecorder(chunkRecordingStream)
    }

    const chunks = []
    chunkMediaRecorder.ondataavailable = (e) => { if (e.data.size > 0) chunks.push(e.data) }
    chunkMediaRecorder.onstop = () => {
      if (chunks.length === 0 || !interviewId.value) return
      const blob = new Blob(chunks, { type: 'audio/webm' })
      const reader = new FileReader()
      reader.onloadend = () => {
        const base64 = reader.result.split(',')[1]
        sendSpeechChunk(base64, SPEECH_CHUNK_SECONDS)
      }
      reader.readAsDataURL(blob)
    }

    chunkMediaRecorder.start()

    // Stop after a short chunk and restart for near-real-time metrics.
    chunkInterval = setTimeout(() => {
      if (chunkMediaRecorder && chunkMediaRecorder.state === 'recording') {
        chunkMediaRecorder.stop()
      }
      // Start next chunk
      if (speechAnalysisActive.value) startNewChunk()
    }, Math.floor(SPEECH_CHUNK_SECONDS * 1000))
  }

  startNewChunk()
}

const sendSpeechChunk = async (audioBase64, duration) => {
  if (!interviewId.value) return
  const reqSeq = ++speechRequestSeq
  try {
    const res = await apiAnalyzeSpeechChunk(interviewId.value, {
      audio_data: audioBase64,
      duration: duration
    })
    if (reqSeq < speechAppliedSeq) {
      return
    }
    speechAppliedSeq = reqSeq

    if (res.metrics) {
      const m = res.metrics
      const rawRate = Number(m.speech_rate || 0)
      const chunkText = String(m.transcribed_text || '').trim()
      const alpha = chunkText ? SPEECH_RATE_SMOOTH_ALPHA : SPEECH_RATE_IDLE_ALPHA
      const prevRate = Number(speechMetrics.value.speechRate || 0)
      const smoothedRate = prevRate <= 0
        ? rawRate
        : (prevRate * (1 - alpha)) + (rawRate * alpha)

      speechMetrics.value.speechRate = Number(smoothedRate.toFixed(1))
      speechMetrics.value.speechRateLevel = classifySpeechRateLevel(speechMetrics.value.speechRate)
      speechMetrics.value.fillerWordCount = Number(m.filler_word_count || 0)
      speechMetrics.value.fluencyAlert = Boolean(m.fluency_alert)
      speechMetrics.value.totalFillerWords += Number(m.filler_word_count || 0)
      speechLastUpdatedAt.value = Date.now()

      if (chunkText) {
        latestChunkTranscript.value = chunkText
        const merged = `${liveTranscript.value} ${chunkText}`.trim()
        liveTranscript.value = merged.length > SPEECH_TRANSCRIPT_MAX_LEN
          ? merged.slice(merged.length - SPEECH_TRANSCRIPT_MAX_LEN)
          : merged
      }
    }
  } catch (err) {
    console.warn('Speech analysis chunk failed:', err)
  }
}

const classifySpeechRateLevel = (rate) => {
  if (rate < 120) return 'slow'
  if (rate <= 240) return 'normal'
  return 'fast'
}

const stopSpeechAnalysis = () => {
  speechAnalysisActive.value = false
  if (chunkInterval) { clearTimeout(chunkInterval); chunkInterval = null }
  if (chunkMediaRecorder && chunkMediaRecorder.state === 'recording') {
    chunkMediaRecorder.stop()
  }
  if (energyAnimFrame) { cancelAnimationFrame(energyAnimFrame); energyAnimFrame = null }
  if (audioContext) { audioContext.close(); audioContext = null }
  analyserNode = null
  chunkMediaRecorder = null
  chunkRecordingStream = null
}

// Camera Logic
const toggleCamera = async () => {
  if (isCameraOn.value) {
    stopCamera()
  } else {
    await startCamera()
  }
}

const toggleMic = () => {
  isMicOn.value = !isMicOn.value
  if (stream.value) {
    stream.value.getAudioTracks().forEach(track => { track.enabled = isMicOn.value })
  }
}

const startCamera = async () => {
  try {
    stream.value = await navigator.mediaDevices.getUserMedia({ video: true, audio: true })
    isCameraOn.value = true
    nextTick(() => {
      if (previewVideo.value) previewVideo.value.srcObject = stream.value
      if (interviewVideo.value) interviewVideo.value.srcObject = stream.value
    })
  } catch (err) {
    console.error("Camera access denied:", err)
    isCameraOn.value = false
  }
}

const stopCamera = () => {
  if (stream.value) {
    stream.value.getTracks().forEach(track => track.stop())
    stream.value = null
  }
  isCameraOn.value = false
  stopSpeechAnalysis()
}

const pushSessionEvent = (type, from, to, meta = {}) => {
  sessionEvents.value.push({
    type,
    from,
    to,
    timestamp: Date.now(),
    ...meta
  })
  if (sessionEvents.value.length > 30) {
    sessionEvents.value = sessionEvents.value.slice(sessionEvents.value.length - 30)
  }
}

const transitionSessionState = (to, type, meta = {}) => {
  const from = sessionState.value
  if (!to || from === to) return
  sessionState.value = to
  pushSessionEvent(type || `session.${String(to).toLowerCase()}`, from, to, meta)
}

const markProcessingTimeout = () => {
  if (processingTimeoutTimer) clearTimeout(processingTimeoutTimer)
  processingTimeoutTimer = setTimeout(() => {
    if (isProcessing.value) {
      messages.value.push({
        role: 'system',
        content: '当前网络较慢，系统正在重试同步面试状态...',
        type: 'system'
      })
      recoverSessionState()
    }
  }, 12000)
}

const clearProcessingTimeout = () => {
  if (processingTimeoutTimer) {
    clearTimeout(processingTimeoutTimer)
    processingTimeoutTimer = null
  }
}

const sanitizeSpeechText = (text) => {
  return String(text || '')
    .replace(/[*`#>_\-\[\]\(\)]/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
    .slice(0, 220)
}

const setInterviewerMotion = (motion, duration = 1200) => {
  interviewerMotion.value = motion
  if (interviewerMotionTimer) clearTimeout(interviewerMotionTimer)
  interviewerMotionTimer = setTimeout(() => {
    interviewerMotion.value = 'idle'
  }, duration)
}

const speakAsInterviewer = async (text) => {
  if (!isVideoSceneMode.value) return
  const spokenText = sanitizeSpeechText(text)
  if (!spokenText) return

  if (currentAudio) {
    currentAudio.pause()
    currentAudio = null
  }
  
  interviewerSpeaking.value = true
  setInterviewerMotion('thinking', 3000)

  try {
    let voice = 'xiaoxiao'
    if (settings.value.style === 'stress') voice = 'yunxi'
    else if (settings.value.style === 'deep') voice = 'yunyang'
    else if (settings.value.style === 'algorithm') voice = 'xiaoyi'
    else if (settings.value.style === 'practical') voice = 'panpan'
    
    const blob = await apiGenerateTTS({
      text: spokenText,
      voice: voice,
      rate: '+10%'
    })
    
    if (!blob || blob.size === 0) throw new Error('Empty audio')
    
    const url = URL.createObjectURL(blob)
    const audio = new Audio(url)
    currentAudio = audio
    
    audio.onplay = () => {
      interviewerSpeaking.value = true
      setInterviewerMotion('talking', 60000)
    }
    
    audio.onended = () => {
      interviewerSpeaking.value = false
      setInterviewerMotion('nod', 900)
      URL.revokeObjectURL(url)
      if (currentAudio === audio) currentAudio = null
      
      if (isVideoSceneMode.value && !pendingEnd.value && phase.value === 'interview') {
        setTimeout(() => {
          if (answerVoiceStatus.value === 'idle' || answerVoiceStatus.value === 'error') {
             startAnswerRecording()
          }
        }, 500)
      }
    }
    
    audio.onerror = (e) => {
      console.error('Audio playback error', e)
      interviewerSpeaking.value = false
      setInterviewerMotion('idle', 600)
      URL.revokeObjectURL(url)
      if (currentAudio === audio) currentAudio = null
    }
    
    await audio.play()
    
  } catch (err) {
    console.error('TTS failed:', err)
    interviewerSpeaking.value = false
    setInterviewerMotion('idle', 600)
  }
}

const recoverSessionState = async () => {
  if (!interviewId.value) return
  try {
    const stateRes = await apiGetInterviewSession(interviewId.value)
    const serverState = stateRes.session_state || 'READY'
    transitionSessionState(serverState, stateRes.recovery_event?.type || 'session.synced', { source: 'server' })
    const interviewRes = await apiGetInterview(interviewId.value)
    if (interviewRes?.interview?.status === 'completed') {
      pendingEnd.value = true
      transitionSessionState('END', 'session.ended', { source: 'recover' })
    }
  } catch (error) {
    console.warn('recoverSessionState failed:', error)
  }
}

const startSessionHeartbeat = () => {
  if (sessionHeartbeatTimer) clearInterval(sessionHeartbeatTimer)
  sessionHeartbeatTimer = setInterval(async () => {
    if (!interviewId.value || phase.value !== 'interview' || isOffline.value) return
    try {
      const stateRes = await apiGetInterviewSession(interviewId.value)
      const serverState = stateRes.session_state || ''
      if (serverState && serverState !== sessionState.value) {
        transitionSessionState(serverState, stateRes.recovery_event?.type || 'session.synced', { source: 'heartbeat' })
      }
    } catch (_) {
    }
  }, 7000)
}

const stopSessionHeartbeat = () => {
  if (sessionHeartbeatTimer) {
    clearInterval(sessionHeartbeatTimer)
    sessionHeartbeatTimer = null
  }
}

const startInterviewRecording = () => {
  if (!stream.value || !interviewId.value) return
  try {
    interviewRecordedChunks = []
    recordingStatus.value = 'recording'
    interviewMediaRecorder = new MediaRecorder(stream.value, { mimeType: 'video/webm;codecs=vp8,opus' })
  } catch (_) {
    try {
      interviewMediaRecorder = new MediaRecorder(stream.value)
      recordingStatus.value = 'recording'
      interviewRecordedChunks = []
    } catch (err) {
      console.warn('无法创建视频录制器:', err)
      recordingStatus.value = 'failed'
      return
    }
  }

  interviewMediaRecorder.ondataavailable = (e) => {
    if (e.data && e.data.size > 0) interviewRecordedChunks.push(e.data)
  }
  interviewMediaRecorder.start(1000)
}

const stopAndUploadInterviewRecording = async () => {
  if (!interviewMediaRecorder || recordingStatus.value !== 'recording' || !interviewId.value) return

  await new Promise((resolve) => {
    interviewMediaRecorder.onstop = resolve
    interviewMediaRecorder.stop()
  })

  if (!interviewRecordedChunks.length) {
    recordingStatus.value = 'failed'
    return
  }

  const blob = new Blob(interviewRecordedChunks, { type: 'video/webm' })
  const formData = new FormData()
  formData.append('recording', blob, `interview_${interviewId.value}.webm`)

  try {
    const res = await apiUploadInterviewRecording(interviewId.value, formData)
    recordingUrl.value = res.recording_url || ''
    recordingStatus.value = 'uploaded'
  } catch (err) {
    console.warn('视频上传失败:', err)
    recordingStatus.value = 'failed'
  } finally {
    interviewMediaRecorder = null
    interviewRecordedChunks = []
  }
}

// Interview Logic
const startInterview = async () => {
  transitionSessionState('INIT', 'session.init')
  isProcessing.value = true
  markProcessingTimeout()
  processingHint.value = '正在初始化面试场景...'
  answerVoiceStatus.value = 'idle'
  answerVoiceError.value = ''
  answerVoiceSeconds.value = 0
  try {
    if (isVideoSceneMode.value && (!isCameraOn.value || !isMicOn.value)) {
      await startCamera()
      if (!isCameraOn.value || !isMicOn.value) {
        throw new Error('视频面试模式要求开启摄像头和麦克风')
      }
    }

    const res = await apiStartInterview({
      position: settings.value.position,
      difficulty: settings.value.difficulty,
      mode: settings.value.mode,
      style: settings.value.style,
      company: settings.value.company,
      interview_mode: settings.value.interviewMode,
      session_state: sessionState.value
    })
    
    // Backend returns { message: "...", interview: { ... } }
    // The interview object contains questions array if loaded correctly
    const interview = res.interview
    interviewId.value = interview.id
    transitionSessionState(res.session_state || 'READY', res.event?.type || 'session.ready')
    pendingNextIndex.value = null
    pendingEnd.value = false

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
    phase.value = 'interview'
    currentQuestionIndex.value = 0
    
    // Initialize Chat — adapt greeting for different modes
    const isBlindBox = settings.value.mode === 'blindbox' && blindBoxScenario.value
    const isRandom = settings.value.interviewMode === 'random'
    
    const modeLabels = { technical: '技术', hr: 'HR', comprehensive: '综合' }
    const styleLabels = { gentle: '温和型', stress: '压力型', deep: '技术深挖型', practical: '项目实战型', algorithm: '算法考察型' }
    const companyLabels = { ali: '阿里巴巴', bytedance: '字节跳动', tencent: '腾讯', meituan: '美团', baidu: '百度' }

    let scenarioGreeting
    if (isBlindBox) {
      scenarioGreeting = `🎲 盲盒场景已揭晓：${blindBoxScenario.value.icon} **${blindBoxScenario.value.name}**\n\n${blindBoxScenario.value.description}\n\n压力等级：${pressureLabels[blindBoxScenario.value.pressure] || '未知'}${blindBoxScenario.value.time_limit ? `\n每题限时：${blindBoxScenario.value.time_limit}秒` : ''}\n\n准备好了吗？让我们开始！`
    } else if (isRandom) {
      scenarioGreeting = `🎲 随机模式已启动！\n\n系统已为您随机分配了面试官风格，在面试过程中不会提前告知。\n这模拟了真实企业面试中的"突然切换风格"场景，请保持灵活应变！\n\n面试岗位：${settings.value.position}\n面试结束后将揭晓面试官风格，让我们开始吧！`
    } else {
      const modeLabel = modeLabels[settings.value.mode] || settings.value.mode
      const companyInfo = settings.value.company ? `（${companyLabels[settings.value.company] || settings.value.company}风格）` : ''
      scenarioGreeting = `你好！我是你的 AI 面试官${companyInfo}。我们将进行一场关于 ${settings.value.position} 的${modeLabel}面试，采用${styleLabels[settings.value.style] || settings.value.style}提问方式。准备好了吗？让我们开始吧。`
    }

    messages.value = [
      {
        role: 'ai',
        content: scenarioGreeting,
        type: isBlindBox ? 'scenario' : undefined
      }
    ]
    speakAsInterviewer(scenarioGreeting)
    
    // Push first question after a short delay
    processingHint.value = '面试官正在组织首个话题...'
    setTimeout(() => {
      transitionSessionState('ASKING', 'interviewer.asking')
      pushAIQuestion(questions.value[0])
      // Start question timer if scenario has time limit
      if (blindBoxScenario.value?.time_limit) {
        startQuestionTimer(blindBoxScenario.value.time_limit)
      }
      scrollToBottom()
    }, 1000)

    // Handle video transition
    if (isCameraOn.value) {
      // Small delay to ensure DOM is ready
      setTimeout(async () => {
        if (!stream.value) await startCamera()
        else if (interviewVideo.value) interviewVideo.value.srcObject = stream.value
        // Start real-time speech analysis
        startSpeechAnalysis()
        startInterviewRecording()
        startSessionHeartbeat()
        connectWebSocket()
      }, 500)
    }
  } catch (error) {
    console.error('Failed to start interview:', error)
    alert('启动面试失败: ' + (error.response?.data?.error || error.message))
  } finally {
    clearProcessingTimeout()
    isProcessing.value = false
    processingHint.value = ''
  }
}

const pushAIQuestion = (question) => {
  const text = (question?.content || question?.title || '').trim()
  if (!text) return
  messages.value.push({
    role: 'ai',
    content: text,
    type: 'question'
  })
  speakAsInterviewer(text)
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

const refreshInterviewQuestions = async () => {
  if (!interviewId.value) return null
  const res = await apiGetInterview(interviewId.value)
  const interview = res.interview
  if (!interview) return null
  questions.value = mapInterviewQuestions(interview.questions || [])
  const nextIndex = Number.isInteger(interview.current_index)
    ? interview.current_index
    : currentQuestionIndex.value + 1
  return { interview, nextIndex }
}

const formatFeedback = (feedback) => {
  if (feedback == null) return '回答已提交，建议补充更具体的技术细节。'

  // 尝试解析为 JSON（新版多维度格式）
  if (typeof feedback === 'string') {
    const trimmed = feedback.trim()
    if (trimmed.startsWith('{')) {
      try {
        const parsed = JSON.parse(trimmed)
        if (parsed.evaluation) {
          // 这是新版结构化 JSON，直接返回原始 JSON 让 splitFeedbackSections 处理
          return trimmed
        }
      } catch (_) {
        // 不是合法 JSON，走旧逻辑
      }
    }
  }

  const extractText = (val) => {
    if (!val) return []
    if (typeof val === 'string') {
      const text = val.trim()
      if (!text) return []
      if (text.startsWith('{') || text.startsWith('[')) {
        try {
          return extractText(JSON.parse(text))
        } catch (_) {
          return [text]
        }
      }
      return [text]
    }
    if (Array.isArray(val)) {
      return val.flatMap((item) => extractText(item))
    }
    if (typeof val === 'object') {
      const blocks = []
      if (typeof val.content === 'string' && val.content.trim()) blocks.push(val.content.trim())
      if (Array.isArray(val.suggestions)) {
        val.suggestions.forEach((s) => {
          if (typeof s === 'string' && s.trim()) blocks.push(`建议：${s.trim()}`)
        })
      }
      const keys = ['feedback', 'analysis', 'comment', 'summary', 'advice', 'suggestion', 'message']
      keys.forEach((k) => {
        if (val[k] !== undefined) blocks.push(...extractText(val[k]))
      })
      return blocks
    }
    return []
  }

  const texts = extractText(feedback).filter(Boolean)
  return texts.length > 0 ? texts.join('\n') : '回答已提交，建议补充更具体的技术细节。'
}

const splitFeedbackSections = (text) => {
  const source = (text || '').trim()
  if (!source) {
    return {
      evaluation: '回答已提交，建议补充更具体的技术细节。',
      suggestions: [],
      dimensions: null,
      highlights: [],
      gaps: [],
      modelAnswerOutline: '',
      followUp: ''
    }
  }

  // 新版 JSON 格式解析
  if (source.startsWith('{')) {
    try {
      const parsed = JSON.parse(source)
      if (parsed.evaluation) {
        return {
          evaluation: parsed.evaluation || '',
          suggestions: Array.isArray(parsed.suggestions) ? parsed.suggestions : (parsed.suggestions ? [parsed.suggestions] : []),
          dimensions: parsed.dimensions || null,
          highlights: Array.isArray(parsed.highlights) ? parsed.highlights.filter(Boolean) : [],
          gaps: Array.isArray(parsed.gaps) ? parsed.gaps.filter(Boolean) : [],
          modelAnswerOutline: parsed.model_answer_outline || '',
          followUp: parsed.follow_up || ''
        }
      }
    } catch (_) {
      // fallthrough to legacy parsing
    }
  }

  // 旧版 【评价】【建议】 格式兼容
  const evalMatch = source.match(/【评价】([\s\S]*?)(?:【建议】|$)/)
  const suggestBlockMatch = source.match(/【建议】([\s\S]*)$/)
  if (evalMatch || suggestBlockMatch) {
    const evaluationText = (evalMatch?.[1] || '').trim() || source
    const suggestionLines = (suggestBlockMatch?.[1] || '')
      .split('\n')
      .map((line) => line.replace(/^[-•\d.)、\s]+/, '').trim())
      .filter(Boolean)
    return {
      evaluation: evaluationText,
      suggestions: suggestionLines,
      dimensions: null,
      highlights: [],
      gaps: [],
      modelAnswerOutline: '',
      followUp: ''
    }
  }

  const lines = source
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean)
  const evaluationParts = []
  const suggestions = []
  lines.forEach((line) => {
    const normalized = line.replace(/^[-•\d.)\s]+/, '').trim()
    if (/^(建议|改进建议|可优化|下一步|你可以)/.test(normalized)) {
      suggestions.push(normalized.replace(/^建议[:：]?\s*/, ''))
      return
    }
    if (/^(1|2|3|4|5)[.)、]\s*/.test(line) && /建议|改进|优化/.test(normalized)) {
      suggestions.push(normalized)
      return
    }
    if (/^(建议：|建议:)/.test(line)) {
      suggestions.push(line.replace(/^(建议：|建议:)\s*/, '').trim())
      return
    }
    evaluationParts.push(line)
  })
  return {
    evaluation: evaluationParts.join('\n') || source,
    suggestions,
    dimensions: null,
    highlights: [],
    gaps: [],
    modelAnswerOutline: '',
    followUp: ''
  }
}

const formatVoiceSeconds = (seconds) => {
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${mins}:${String(secs).padStart(2, '0')}`
}

const getVoiceStatusLabel = () => {
  const labels = {
    idle: '待命',
    requesting: '请求麦克风权限',
    recording: `录音中 ${formatVoiceSeconds(answerVoiceSeconds.value)}`,
    transcribing: '语音转写中',
    submitting: '提交语音答案中',
    success: '语音答案已提交',
    error: answerVoiceError.value || '语音失败'
  }
  return labels[answerVoiceStatus.value] || '待命'
}

const normalizeAnswerSubmitError = (msg = '') => {
  const text = String(msg || '')
  if (!text) return '未知错误'
  if (/audio\s+data\s+is\s+empty/i.test(text) || /empty\s+transcription\s+result/i.test(text)) {
    return '未识别到有效语音，请靠近麦克风并清晰作答后重试'
  }
  if (/status:\s*401|authentication_error|authorization|unauthorized/i.test(text)) {
    return '语音转写鉴权失败（ASR Key 无效或无权限），请联系管理员检查 server/config.yaml 的 asr.api_key'
  }
  if (/maximum\s+content\s+size\s+limit|status:\s*413|payload\s+too\s+large/i.test(text)) {
    return '语音文件过大，请缩短单次语音回答时长后重试'
  }
  if (/invalid\s+file\s+format|unsupported\s+audio|invalid\s+audio|decode\s+audio/i.test(text)) {
    return '语音格式暂不受支持，请使用 Chrome/Edge 重试并确保允许麦克风权限'
  }
  if (/model\s+field|model\s+is\s+required|unknown\s+model|invalid\s+model/i.test(text)) {
    return '语音转写服务配置异常（模型参数无效），请联系管理员检查 ASR 配置'
  }
  if (/field\s+validation.*answer.*required/i.test(text) || /key:\s*'answer'/i.test(text)) {
    return '您似乎没有做出任何回答'
  }
  if (/failed\s+to\s+transcribe\s+audio/i.test(text)) {
    return '语音转写失败，请稍后重试；若持续失败，请检查后端 ASR 配置与日志'
  }
  return text
}

const submitCurrentAnswer = async (answerText = '', audioData = '') => {
  const currentQ = questions.value[currentQuestionIndex.value]
  if (!currentQ || !currentQ.questionId) {
    throw new Error('当前题目ID无效，请重新开始面试')
  }

  const res = await apiSubmitAnswer(interviewId.value, {
    question_id: currentQ.questionId,
    answer: answerText,
    audio_data: audioData,
    session_state: sessionState.value
  })
  if (res?.session_state) {
    transitionSessionState(res.session_state, res.event?.type || 'session.synced', { source: 'submit' })
  }

  const result = res.result
  const formatted = formatFeedback(result.feedback)
  const feedbackSections = splitFeedbackSections(formatted)
  messages.value.push({
    role: 'ai',
    content: formatted,
    type: 'feedback',
    score: result.score,
    feedbackEvaluation: feedbackSections.evaluation,
    feedbackSuggestions: feedbackSections.suggestions,
    feedbackDimensions: feedbackSections.dimensions,
    feedbackHighlights: feedbackSections.highlights,
    feedbackGaps: feedbackSections.gaps,
    feedbackModelAnswer: feedbackSections.modelAnswerOutline,
    feedbackFollowUp: feedbackSections.followUp
  })

  try {
    const refreshed = await refreshInterviewQuestions()
    const nextIndex = refreshed?.nextIndex ?? (currentQuestionIndex.value + 1)
    if (nextIndex < questions.value.length) {
      pendingNextIndex.value = nextIndex
      pendingEnd.value = false
    } else {
      pendingNextIndex.value = null
      pendingEnd.value = true
    }
  } catch (e) {
    const fallbackIndex = currentQuestionIndex.value + 1
    if (fallbackIndex < questions.value.length) {
      pendingNextIndex.value = fallbackIndex
      pendingEnd.value = false
    } else {
      pendingNextIndex.value = null
      pendingEnd.value = true
    }
  }

  return result
}

const submitAudioAnswer = async (audioData) => {
  if (!audioData) return
  if (isProcessing.value) return

  const userMessageIndex = messages.value.push({
    role: 'user',
    content: '【语音回答】'
  }) - 1

  isProcessing.value = true
  transitionSessionState('THINKING', 'candidate.submitted_audio')
  markProcessingTimeout()
  processingHint.value = '面试官正在转写并评估你的语音回答...'
  answerVoiceStatus.value = 'submitting'
  answerVoiceError.value = ''

  try {
    const result = await submitCurrentAnswer('', audioData)
    const transcribedAnswer = String(result?.answer || '').trim()
    if (transcribedAnswer && messages.value[userMessageIndex]) {
      messages.value[userMessageIndex].content = `【语音转写】${transcribedAnswer}`
    }

    answerVoiceStatus.value = 'success'
    setTimeout(() => {
      if (answerVoiceStatus.value === 'success') {
        answerVoiceStatus.value = 'idle'
      }
    }, 1600)
  } catch (error) {
    const rawErrMsg = error?.response?.data?.error || error?.message || '未知错误'
    const errMsg = normalizeAnswerSubmitError(rawErrMsg)
    answerVoiceError.value = errMsg
    answerVoiceStatus.value = 'error'

    if (errMsg.includes('not in progress') || errMsg.includes('已结束')) {
      messages.value.push({
        role: 'ai',
        content: '面试结束！辛苦了。您可以点击下方按钮查看详细报告。',
        type: 'system'
      })
      completeInterview()
    } else {
      messages.value.push({
        role: 'system',
        content: `提交语音答案失败：${errMsg}`,
        type: 'system'
      })
    }
  } finally {
    clearProcessingTimeout()
    isProcessing.value = false
    processingHint.value = ''
    scrollToBottom()
  }
}

const startAnswerRecording = async () => {
  if (isProcessing.value || !interviewId.value) return
  if (!isMicOn.value) {
    answerVoiceError.value = '麦克风已关闭，请先开启麦克风'
    answerVoiceStatus.value = 'error'
    return
  }

  answerVoiceStatus.value = 'requesting'
  answerVoiceError.value = ''
  answerVoiceSeconds.value = 0
  answerAudioChunks = []

  try {
    answerRecorderStream = await navigator.mediaDevices.getUserMedia({ audio: true })

    try {
      answerMediaRecorder = new MediaRecorder(answerRecorderStream, { mimeType: 'audio/webm' })
    } catch (_) {
      answerMediaRecorder = new MediaRecorder(answerRecorderStream)
    }

    answerMediaRecorder.ondataavailable = (event) => {
      if (event.data && event.data.size > 0) {
        answerAudioChunks.push(event.data)
      }
    }

    answerMediaRecorder.onstop = async () => {
      if (answerVoiceTimer) {
        clearInterval(answerVoiceTimer)
        answerVoiceTimer = null
      }

      if (!answerAudioChunks.length) {
        answerVoiceError.value = '未检测到有效语音，请重试'
        answerVoiceStatus.value = 'error'
        return
      }

      answerVoiceStatus.value = 'transcribing'
      const audioBlob = new Blob(answerAudioChunks, { type: 'audio/webm' })
      const reader = new FileReader()
      reader.onloadend = async () => {
        const raw = reader.result || ''
        const parts = String(raw).split(',')
        if (parts.length < 2) {
          answerVoiceError.value = '音频编码失败，请重试'
          answerVoiceStatus.value = 'error'
          return
        }
        await submitAudioAnswer(parts[1])
      }
      reader.readAsDataURL(audioBlob)
    }

    answerMediaRecorder.start()
    transitionSessionState('LISTENING', 'candidate.listening')
    answerVoiceStatus.value = 'recording'
    answerVoiceTimer = setInterval(() => {
      answerVoiceSeconds.value += 1
    }, 1000)
  } catch (err) {
    console.warn('startAnswerRecording failed:', err)
    answerVoiceError.value = '无法访问麦克风权限'
    answerVoiceStatus.value = 'error'
  }
}

const stopAnswerRecording = () => {
  if (!answerMediaRecorder || answerVoiceStatus.value !== 'recording') return
  transitionSessionState('THINKING', 'candidate.stop_recording')
  answerMediaRecorder.stop()
  if (answerRecorderStream) {
    answerRecorderStream.getTracks().forEach(track => track.stop())
    answerRecorderStream = null
  }
}

const toggleAnswerRecording = async () => {
  if (answerVoiceStatus.value === 'recording') {
    stopAnswerRecording()
    return
  }
  await startAnswerRecording()
}

const sendMessage = async () => {
  if (isProcessing.value) return
  if (latestAIMessage.value?.type === 'feedback') {
    advanceToNextQuestion()
    return
  }
  if (!userInput.value.trim()) return
  
  const answer = userInput.value
  userInput.value = ''
  
  // 1. Add User Message
  messages.value.push({
    role: 'user',
    content: answer
  })
  
  isProcessing.value = true
  transitionSessionState('THINKING', 'candidate.submitted_text')
  markProcessingTimeout()
  processingHint.value = '面试官正在评估你的回答...'
  
  try {
    // 2. Submit to Backend
    await submitCurrentAnswer(answer, '')
    processingHint.value = '面试官正在生成下一轮追问...'
    
  } catch (error) {
    console.error('Failed to submit answer:', error)
    const rawErrMsg = error?.response?.data?.error || error?.message || '未知错误'
    const errMsg = normalizeAnswerSubmitError(rawErrMsg)
    
    // If the interview was already completed (e.g. backend marked it done), handle gracefully
    if (errMsg.includes('not in progress') || errMsg.includes('已结束')) {
      messages.value.push({
        role: 'ai',
        content: '面试结束！辛苦了。您可以点击下方按钮查看详细报告。',
        type: 'system'
      })
      completeInterview()
    } else {
      messages.value.push({
        role: 'system',
        content: `提交答案失败：${errMsg}`,
        type: 'system'
      })
    }
  } finally {
    clearProcessingTimeout()
    isProcessing.value = false
    processingHint.value = ''
    scrollToBottom()
  }
}

const advanceToNextQuestion = () => {
  if (pendingEnd.value) {
    stopQuestionTimer()
    messages.value.push({
      role: 'ai',
      content: "面试结束！辛苦了。您可以点击下方按钮查看详细报告。",
      type: 'system'
    })
    if (settings.value.interviewMode === 'random') {
      revealStyle()
    }
    pendingEnd.value = false
    pendingNextIndex.value = null
    completeInterview()
    scrollToBottom()
    return
  }

  const nextIndex = pendingNextIndex.value != null
    ? pendingNextIndex.value
    : currentQuestionIndex.value + 1
  if (nextIndex < questions.value.length) {
    currentQuestionIndex.value = nextIndex
    transitionSessionState('ASKING', 'interviewer.asking')
    pushAIQuestion(questions.value[currentQuestionIndex.value])
    if (blindBoxScenario.value?.time_limit) {
      startQuestionTimer(blindBoxScenario.value.time_limit)
    }
  }
  pendingNextIndex.value = null
  pendingEnd.value = false
  scrollToBottom()
}

const completeInterview = async () => {
  if (isGeneratingReport.value || !interviewId.value) return
  isGeneratingReport.value = true
  transitionSessionState('END', 'session.ending')
  try {
    await stopAndUploadInterviewRecording()
    await apiEndInterview(interviewId.value)
    const reportRes = await apiGenerateReport({
      interview_id: interviewId.value
    })
    if (reportRes?.report?.id) {
      reportId.value = reportRes.report.id
    }
    if (!reportId.value) {
      messages.value.push({
        role: 'system',
        content: '报告生成中，请稍后点击“查看面试报告”。',
        type: 'system'
      })
    }
  } catch (error) {
    console.error('Failed to end interview:', error)
    const errMsg = error?.response?.data?.error || error?.message || '未知错误'
    messages.value.push({
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
      messages.value.push({
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
  messages.value.push({
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
    transitionSessionState('END', 'session.cancelled')
    answerVoiceStatus.value = 'idle'
    answerVoiceError.value = ''
    answerVoiceSeconds.value = 0
    await stopAndUploadInterviewRecording()
    stopCamera()
    stopQuestionTimer()
    phase.value = 'setup'
    currentQuestionIndex.value = 0
    messages.value = []
    pendingNextIndex.value = null
    pendingEnd.value = false
    blindBoxScenario.value = null
    blindBoxRevealed.value = false
    randomStyleRevealed.value = false
    revealedStyleInfo.value = null
    stopSessionHeartbeat()
    sessionEvents.value = []
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

// ===== Human Interviewer Functions =====
const loadHumanInterviewers = async (type_filter = '') => {
  humanInterviewersLoading.value = true
  try {
    const res = await apiGetHumanInterviewers({ type: type_filter, page: 1, page_size: 50 })
    humanInterviewers.value = res.interviewers || []
  } catch (err) {
    console.warn('Failed to load human interviewers:', err)
    humanInterviewers.value = []
  } finally {
    humanInterviewersLoading.value = false
  }
}

const selectInterviewer = (interviewer) => {
  selectedInterviewer.value = interviewer
  showBookingDialog.value = true
}

const submitBooking = async () => {
  if (!selectedInterviewer.value || !bookingForm.value.scheduledAt) return
  try {
    const res = await apiBookHumanInterview({
      interviewer_id: selectedInterviewer.value.id,
      scheduled_at: new Date(bookingForm.value.scheduledAt).toISOString(),
      position: settings.value.position,
      difficulty: settings.value.difficulty,
      notes: bookingForm.value.notes
    })
    alert('预约成功！面试官确认后将收到通知。')
    showBookingDialog.value = false
    bookingForm.value = { scheduledAt: '', notes: '' }
    selectedInterviewer.value = null
    loadUserBookings()
  } catch (err) {
    alert('预约失败：' + (err.response?.data?.error || err.message))
  }
}

const loadUserBookings = async () => {
  try {
    const res = await apiGetUserBookings()
    userBookings.value = res.bookings || []
  } catch (err) {
    console.warn('Failed to load bookings:', err)
  }
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

const handleNetworkOnline = () => {
  isOffline.value = false
  recoverSessionState()
}

const handleNetworkOffline = () => {
  isOffline.value = true
  pushSessionEvent('session.connection_lost', sessionState.value, sessionState.value, { source: 'browser' })
}

onMounted(() => {
  transitionSessionState('INIT', 'session.boot')
  isOffline.value = !navigator.onLine
  startCamera()
  loadInterviewConfig()
  window.addEventListener('online', handleNetworkOnline)
  window.addEventListener('offline', handleNetworkOffline)
})

onUnmounted(() => {
  if (interviewMediaRecorder && interviewMediaRecorder.state === 'recording') {
    interviewMediaRecorder.stop()
  }
  if (answerMediaRecorder && answerMediaRecorder.state === 'recording') {
    answerMediaRecorder.stop()
  }
  if (answerRecorderStream) {
    answerRecorderStream.getTracks().forEach(track => track.stop())
    answerRecorderStream = null
  }
  if (answerVoiceTimer) {
    clearInterval(answerVoiceTimer)
    answerVoiceTimer = null
  }
  stopCamera()
  stopSpeechAnalysis()
  stopSessionHeartbeat()
  clearProcessingTimeout()
  if (interviewerMotionTimer) {
    clearTimeout(interviewerMotionTimer)
    interviewerMotionTimer = null
  }
  if (currentAudio) {
    currentAudio.pause()
    currentAudio = null
  }
  disconnectWebSocket()
  window.removeEventListener('online', handleNetworkOnline)
  window.removeEventListener('offline', handleNetworkOffline)
})
</script>

<template>
  <div class="h-[calc(100vh-8rem)] flex flex-col">
    <!-- Setup Phase -->
    <div v-if="phase === 'setup'" class="flex-1 flex flex-col items-center justify-center max-w-4xl mx-auto w-full space-y-8 animate-in fade-in duration-500">
      <header class="text-center">
        <h1 class="text-3xl font-bold text-zinc-900">AI 模拟面试</h1>
        <p class="text-zinc-500 mt-2">配置您的面试环境与偏好，开启真实对话体验</p>
      </header>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-8 w-full">
        <!-- Preview Area -->
        <div class="aspect-video bg-zinc-900 rounded-2xl relative overflow-hidden flex items-center justify-center group shadow-xl">
          <video ref="previewVideo" class="w-full h-full object-cover" autoplay muted v-if="isCameraOn"></video>
          <div v-else class="flex flex-col items-center text-zinc-500">
            <VideoOff class="h-12 w-12 mb-2" />
            <span>摄像头已关闭</span>
          </div>
          
          <div class="absolute bottom-4 left-1/2 -translate-x-1/2 flex items-center gap-3">
            <button 
              @click="toggleMic"
              class="h-10 w-10 rounded-full flex items-center justify-center transition-all hover:scale-110 active:scale-95"
              :class="isMicOn ? 'bg-white/10 text-white backdrop-blur-md hover:bg-white/20' : 'bg-rose-500 text-white'"
            >
              <Mic v-if="isMicOn" class="h-4 w-4" />
              <MicOff v-else class="h-4 w-4" />
            </button>
            <button 
              @click="toggleCamera"
              class="h-10 w-10 rounded-full flex items-center justify-center transition-all hover:scale-110 active:scale-95"
              :class="isCameraOn ? 'bg-white/10 text-white backdrop-blur-md hover:bg-white/20' : 'bg-rose-500 text-white'"
            >
              <Video v-if="isCameraOn" class="h-4 w-4" />
              <VideoOff v-else class="h-4 w-4" />
            </button>
          </div>
        </div>

        <!-- Settings Area -->
        <div class="space-y-5 bg-white p-6 rounded-2xl border border-zinc-100 shadow-sm overflow-y-auto max-h-[480px]">
          <div class="space-y-2">
            <label class="text-xs font-bold text-zinc-400 uppercase tracking-wider">目标岗位</label>
            <input 
              v-model="settings.position" 
              class="w-full bg-zinc-50 border border-zinc-200 rounded-xl px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all"
              placeholder="例如: Java 开发工程师"
            />
          </div>

          <!-- ===== 1. Interview Type (面试类型) ===== -->
          <div class="space-y-2">
            <label class="text-xs font-bold text-zinc-400 uppercase tracking-wider">面试类型</label>
            <div class="grid grid-cols-3 gap-2">
              <button 
                v-for="m in [
                  { key: 'technical', label: '技术面', icon: Monitor, desc: '编程/算法/系统设计' },
                  { key: 'hr', label: 'HR面', icon: UserCheck, desc: '沟通/职业规划/软技能' },
                  { key: 'comprehensive', label: '综合面', icon: BrainCircuit, desc: '技术+HR联合面试' }
                ]"
                :key="m.key"
                @click="settings.mode = m.key"
                class="flex flex-col items-center gap-1 px-3 py-3 rounded-xl text-sm font-medium border transition-all text-center relative group"
                :class="settings.mode === m.key 
                  ? 'bg-indigo-50 border-indigo-200 text-indigo-600 ring-1 ring-indigo-200'
                  : 'bg-white border-zinc-200 text-zinc-600 hover:bg-zinc-50'"
              >
                <component :is="m.icon" class="h-5 w-5 shrink-0" />
                <span class="font-bold text-xs">{{ m.label }}</span>
                <span class="text-[10px] text-zinc-400 leading-tight">{{ m.desc }}</span>
              </button>
            </div>
          </div>

          <!-- ===== 2. Interviewer Style (面试官风格) ===== -->
          <div class="space-y-2">
            <label class="text-xs font-bold text-zinc-400 uppercase tracking-wider">面试官风格</label>
            <div class="grid grid-cols-3 gap-2">
              <button 
                v-for="s in [
                  { key: 'gentle', label: '温和型', icon: Heart, color: 'emerald' },
                  { key: 'stress', label: '压力型', icon: Flame, color: 'rose' },
                  { key: 'deep', label: '技术深挖', icon: Search, color: 'violet' },
                  { key: 'practical', label: '项目实战', icon: Briefcase, color: 'amber' },
                  { key: 'algorithm', label: '算法考察', icon: Code, color: 'cyan' }
                ]" 
                :key="s.key"
                @click="settings.style = s.key"
                class="flex items-center gap-2 px-3 py-2.5 rounded-xl text-xs font-medium border transition-all"
                :class="settings.style === s.key ? 'bg-indigo-50 border-indigo-200 text-indigo-600 ring-1 ring-indigo-200' : 'bg-white border-zinc-200 text-zinc-600 hover:bg-zinc-50'"
              >
                <component :is="s.icon" class="h-3.5 w-3.5 shrink-0" />
                {{ s.label }}
              </button>
            </div>
          </div>

          <!-- ===== Company Style (大厂面试风格复刻) ===== -->
          <div class="space-y-2">
            <label class="text-xs font-bold text-zinc-400 uppercase tracking-wider flex items-center gap-2">
              <Building2 class="w-3.5 h-3.5" /> 大厂面试风格（可选）
            </label>
            <div class="grid grid-cols-3 gap-2">
              <button 
                v-for="c in [
                  { key: '', label: '不限', emoji: '🌐' },
                  { key: 'ali', label: '阿里', emoji: '🟠' },
                  { key: 'bytedance', label: '字节', emoji: '⚡' },
                  { key: 'tencent', label: '腾讯', emoji: '🐧' },
                  { key: 'meituan', label: '美团', emoji: '🟡' },
                  { key: 'baidu', label: '百度', emoji: '🔵' }
                ]" 
                :key="c.key"
                @click="settings.company = c.key"
                class="flex items-center gap-1.5 px-3 py-2 rounded-xl text-xs font-medium border transition-all"
                :class="settings.company === c.key ? 'bg-orange-50 border-orange-200 text-orange-700 ring-1 ring-orange-200' : 'bg-white border-zinc-200 text-zinc-600 hover:bg-zinc-50'"
              >
                <span>{{ c.emoji }}</span>
                {{ c.label }}
              </button>
            </div>
          </div>
          
          <!-- ===== 3. Difficulty Level (难度等级) ===== -->
          <div class="space-y-2">
            <label class="text-xs font-bold text-zinc-400 uppercase tracking-wider flex items-center gap-2">
              <GraduationCap class="w-3.5 h-3.5" /> 难度等级
            </label>
            <div class="grid grid-cols-3 gap-2">
              <button 
                v-for="d in [
                  { key: 'campus_intern', label: '校招实习', desc: '在校实习生' },
                  { key: 'campus_graduate', label: '校招应届', desc: '应届毕业生' },
                  { key: 'social_junior', label: '社招初级', desc: '1-3年经验' }
                ]" 
                :key="d.key"
                @click="settings.difficulty = d.key"
                class="flex flex-col items-center gap-0.5 px-3 py-2.5 rounded-xl text-xs font-medium border transition-all"
                :class="settings.difficulty === d.key ? 'bg-indigo-50 border-indigo-200 text-indigo-600 ring-1 ring-indigo-200' : 'bg-white border-zinc-200 text-zinc-600 hover:bg-zinc-50'"
              >
                <span class="font-bold">{{ d.label }}</span>
                <span class="text-[10px] text-zinc-400">{{ d.desc }}</span>
              </button>
            </div>
          </div>

          <!-- ===== 4. Interview Mode (面试模式: AI/真人/随机) ===== -->
          <div class="space-y-2">
            <label class="text-xs font-bold text-zinc-400 uppercase tracking-wider">面试模式</label>
            <div class="grid grid-cols-3 gap-2">
              <button 
                v-for="im in [
                  { key: 'ai', label: 'AI仿真面试官', icon: '🤖', desc: 'AI模拟真实面试' },
                  { key: 'human', label: '真人面试', icon: '👤', desc: '预约真人面试官' },
                  { key: 'random', label: '随机模式', icon: '🎲', desc: '风格随机不提前告知' }
                ]" 
                :key="im.key"
                @click="settings.interviewMode = im.key; if(im.key === 'human') loadHumanInterviewers()"
                class="flex flex-col items-center gap-1 px-3 py-3 rounded-xl text-xs font-medium border transition-all text-center"
                :class="settings.interviewMode === im.key 
                  ? (im.key === 'random' ? 'bg-violet-50 border-violet-300 text-violet-700 ring-1 ring-violet-200' : im.key === 'human' ? 'bg-emerald-50 border-emerald-200 text-emerald-700 ring-1 ring-emerald-200' : 'bg-indigo-50 border-indigo-200 text-indigo-600 ring-1 ring-indigo-200')
                  : 'bg-white border-zinc-200 text-zinc-600 hover:bg-zinc-50'"
              >
                <span class="text-xl">{{ im.icon }}</span>
                <span class="font-bold">{{ im.label }}</span>
                <span class="text-[10px] text-zinc-400 leading-tight">{{ im.desc }}</span>
              </button>
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-xs font-bold text-zinc-400 uppercase tracking-wider">对话呈现模式</label>
            <div class="grid grid-cols-2 gap-2">
              <button
                @click="settings.sceneMode = 'video'"
                class="flex items-center justify-center gap-2 px-3 py-2.5 rounded-xl text-xs font-medium border transition-all"
                :class="settings.sceneMode === 'video' ? 'bg-indigo-50 border-indigo-200 text-indigo-700 ring-1 ring-indigo-200' : 'bg-white border-zinc-200 text-zinc-600 hover:bg-zinc-50'"
              >
                <Video class="w-3.5 h-3.5" />
                视频化面试
              </button>
              <button
                @click="settings.sceneMode = 'classic'"
                class="flex items-center justify-center gap-2 px-3 py-2.5 rounded-xl text-xs font-medium border transition-all"
                :class="settings.sceneMode === 'classic' ? 'bg-zinc-100 border-zinc-300 text-zinc-800 ring-1 ring-zinc-200' : 'bg-white border-zinc-200 text-zinc-600 hover:bg-zinc-50'"
              >
                <MessageSquare class="w-3.5 h-3.5" />
                经典对话
              </button>
            </div>
            <p class="text-[11px] text-zinc-400">
              视频化模式将启用面试官形象语音播报，且必须开启摄像头与麦克风。
            </p>
          </div>

          <!-- Random Mode Notice -->
          <div v-if="settings.interviewMode === 'random'" class="p-3 bg-violet-50 rounded-xl border border-violet-200 animate-in fade-in duration-300">
            <div class="flex items-start gap-2">
              <Shuffle class="w-4 h-4 text-violet-600 mt-0.5 shrink-0" />
              <div>
                <p class="text-xs font-bold text-violet-700">随机模式说明</p>
                <p class="text-[11px] text-violet-600 leading-relaxed mt-1">
                  系统将随机分配面试官风格（温和/压力/深挖等），可能随机匹配大厂面试风格。
                  面试过程中不会提前告知风格类型，模拟真实企业的"突然压力追问"场景。
                  面试结束后将揭晓本次的面试官风格。
                </p>
              </div>
            </div>
          </div>

          <!-- ===== Human Interview Panel ===== -->
          <div v-if="settings.interviewMode === 'human'" class="space-y-3 animate-in fade-in slide-in-from-top-2 duration-300">
            <!-- Interviewer Type Tabs -->
            <div class="flex gap-2">
              <button 
                v-for="t in [{key: '', label: '全部'}, {key: 'campus', label: '🏫 校内老师'}, {key: 'enterprise', label: '🏢 企业专家'}]"
                :key="t.key"
                @click="loadHumanInterviewers(t.key)"
                class="px-3 py-1.5 rounded-lg text-xs font-medium border transition-all"
                :class="'bg-white border-zinc-200 text-zinc-600 hover:bg-zinc-50'"
              >
                {{ t.label }}
              </button>
            </div>

            <!-- Interviewers List -->
            <div v-if="humanInterviewersLoading" class="text-center py-4">
              <Loader2 class="w-6 h-6 text-zinc-400 animate-spin mx-auto" />
              <p class="text-xs text-zinc-400 mt-2">加载面试官列表...</p>
            </div>
            <div v-else-if="humanInterviewers.length > 0" class="space-y-2 max-h-[180px] overflow-y-auto custom-scrollbar">
              <div 
                v-for="interviewer in humanInterviewers" 
                :key="interviewer.id"
                @click="selectInterviewer(interviewer)"
                class="flex items-center gap-3 p-3 rounded-xl border border-zinc-100 hover:border-indigo-200 hover:bg-indigo-50/30 cursor-pointer transition-all group"
              >
                <div class="h-10 w-10 rounded-full bg-gradient-to-br from-indigo-100 to-purple-100 flex items-center justify-center text-indigo-700 font-bold text-sm shrink-0">
                  {{ interviewer.name?.[0] || '?' }}
                </div>
                <div class="flex-1 min-w-0">
                  <div class="flex items-center gap-2">
                    <span class="text-sm font-bold text-zinc-800">{{ interviewer.name }}</span>
                    <span class="text-[10px] px-1.5 py-0.5 rounded-full" 
                      :class="interviewer.type === 'campus' ? 'bg-blue-100 text-blue-700' : 'bg-orange-100 text-orange-700'">
                      {{ interviewer.type === 'campus' ? '校内' : '企业' }}
                    </span>
                  </div>
                  <p class="text-[11px] text-zinc-500 truncate">{{ interviewer.title }}{{ interviewer.company ? ` · ${interviewer.company}` : '' }}</p>
                  <div class="flex items-center gap-2 mt-0.5">
                    <span class="text-[10px] text-amber-600 flex items-center gap-0.5">
                      <Star class="w-2.5 h-2.5 fill-amber-400 text-amber-400" /> {{ interviewer.rating?.toFixed(1) }}
                    </span>
                    <span class="text-[10px] text-zinc-400">{{ interviewer.total_sessions }}次面试</span>
                  </div>
                </div>
                <Calendar class="w-4 h-4 text-zinc-300 group-hover:text-indigo-500 transition-colors shrink-0" />
              </div>
            </div>
            <div v-else class="p-4 bg-zinc-50 rounded-xl text-center">
              <p class="text-xs text-zinc-400">暂无可用面试官。校内老师和企业专家将陆续上线。</p>
            </div>

            <!-- Bookings Button -->
            <button @click="showBookingsPanel = true; loadUserBookings()" class="w-full py-2 rounded-xl text-xs font-medium border border-zinc-200 bg-white hover:bg-zinc-50 text-zinc-600 transition-all flex items-center justify-center gap-1.5">
              <Clock class="w-3 h-3" /> 查看我的预约
            </button>
          </div>

          <!-- Blind Box Mode (unchanged) -->
          <div v-if="settings.mode === 'blindbox'" class="space-y-3 animate-in fade-in slide-in-from-top-2 duration-300">
            <div v-if="blindBoxRevealing" class="p-6 bg-gradient-to-br from-violet-100 to-purple-50 rounded-2xl border border-violet-200 flex flex-col items-center justify-center gap-3">
              <div class="relative">
                <Package class="h-12 w-12 text-violet-600 animate-bounce" />
                <div class="absolute -top-1 -right-1 w-4 h-4 bg-yellow-400 rounded-full animate-ping"></div>
              </div>
              <p class="text-sm font-bold text-violet-700 animate-pulse">正在抽取面试场景...</p>
            </div>
            <div v-else-if="blindBoxRevealed && blindBoxScenario" 
              class="p-4 rounded-2xl border-2 shadow-md animate-in fade-in zoom-in-95 duration-500 relative overflow-hidden"
              :class="[pressureColors[pressureLevel]?.bg, pressureColors[pressureLevel]?.border]"
            >
              <div class="absolute top-0 right-0 w-20 h-20 bg-gradient-to-bl from-white/40 to-transparent rounded-bl-full pointer-events-none"></div>
              <div class="flex items-start gap-3">
                <span class="text-3xl">{{ blindBoxScenario.icon }}</span>
                <div class="flex-1 min-w-0">
                  <div class="flex items-center gap-2 mb-1">
                    <h4 class="font-bold text-base" :class="pressureColors[pressureLevel]?.text">{{ blindBoxScenario.name }}</h4>
                    <span class="text-[10px] font-bold px-2 py-0.5 rounded-full" :class="pressureColors[pressureLevel]?.badge">
                      {{ pressureLabels[pressureLevel] }}
                    </span>
                  </div>
                  <p class="text-xs text-zinc-600 leading-relaxed mb-2">{{ blindBoxScenario.description }}</p>
                  <div class="flex flex-wrap gap-1.5">
                    <span v-for="tag in blindBoxScenario.tags" :key="tag" class="text-[10px] px-2 py-0.5 rounded-full bg-white/60 text-zinc-500 border border-zinc-200">{{ tag }}</span>
                    <span v-if="blindBoxScenario.time_limit" class="text-[10px] px-2 py-0.5 rounded-full bg-white/60 text-zinc-500 border border-zinc-200 flex items-center gap-1">
                      <Timer class="w-2.5 h-2.5" /> {{ blindBoxScenario.time_limit }}s/题
                    </span>
                  </div>
                </div>
              </div>
              <button @click="reDrawBlindBox" class="mt-3 w-full py-2 rounded-xl text-xs font-medium border border-zinc-200 bg-white/80 hover:bg-white text-zinc-600 transition-all flex items-center justify-center gap-1.5">
                <Shuffle class="w-3 h-3" /> 换一个场景
              </button>
            </div>
            <div v-else class="p-4 bg-zinc-50 rounded-2xl border border-dashed border-zinc-300 text-center">
              <button @click="drawBlindBox" class="px-4 py-2 rounded-xl bg-violet-600 text-white text-sm font-medium hover:bg-violet-700 transition-all flex items-center gap-2 mx-auto">
                <Package class="w-4 h-4" /> 抽取盲盒场景
              </button>
            </div>
          </div>

          <!-- AI Shadow Coach Toggle -->
          <div class="flex items-center justify-between p-3 bg-zinc-50 rounded-xl">
            <div class="flex items-center gap-2">
              <Headphones class="h-4 w-4 text-indigo-600" />
              <span class="text-sm font-medium text-zinc-700">AI 影子教练 (实时耳返)</span>
            </div>
            <button
              @click="shadowCoachEnabled = !shadowCoachEnabled"
              class="w-10 h-5 rounded-full transition-colors relative"
              :class="shadowCoachEnabled ? 'bg-indigo-600' : 'bg-zinc-300'"
            >
              <div class="absolute top-0.5 left-0.5 w-4 h-4 bg-white rounded-full transition-transform" :class="shadowCoachEnabled ? 'translate-x-5' : ''"></div>
            </button>
          </div>

          <button 
            @click="startInterview"
            :disabled="isProcessing || (settings.interviewMode === 'human')"
            class="w-full mt-2 py-4 bg-indigo-600 text-white rounded-xl font-bold text-lg hover:bg-indigo-700 transition-all flex items-center justify-center gap-2 shadow-lg shadow-indigo-200 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Loader2 v-if="isProcessing" class="h-5 w-5 animate-spin" />
            <span v-else-if="settings.interviewMode === 'human'">请先预约真人面试官</span>
            <span v-else>开始面试</span>
            <ChevronRight v-if="!isProcessing && settings.interviewMode !== 'human'" class="h-5 w-5" />
          </button>
        </div>
      </div>
    </div>

    <!-- Interview Phase (New Layout) -->
    <div v-else-if="phase === 'interview'" class="h-full flex flex-col lg:flex-row gap-6 p-6 bg-zinc-50 overflow-y-auto">
      
      <!-- Left Main Column (Video + Input) -->
      <div class="flex-1 flex flex-col gap-6 min-w-0 h-full">
        <!-- Video Section (Top) -->
        <div class="flex-1 bg-black rounded-3xl relative overflow-hidden shadow-2xl group ring-1 ring-zinc-900/5">
          <!-- Status Badge -->
          <div class="absolute top-6 left-6 flex items-center gap-3 z-10 pointer-events-none">
            <div class="text-white text-xs font-bold px-3 py-1.5 rounded-full flex items-center gap-2 shadow-lg"
              :class="recordingStatus === 'uploaded' ? 'bg-emerald-600 shadow-emerald-900/20' : recordingStatus === 'failed' ? 'bg-amber-600 shadow-amber-900/20' : 'bg-rose-600 shadow-rose-900/20'">
              <div class="w-2 h-2 bg-white rounded-full animate-pulse"></div>
              {{ recordingStatus === 'uploaded' ? 'REC OK' : recordingStatus === 'failed' ? 'REC WARN' : 'REC' }}
            </div>
            <!-- Blind Box scenario badge -->
            <div v-if="blindBoxScenario" class="text-white text-xs font-bold px-3 py-1.5 rounded-full flex items-center gap-2 backdrop-blur-md border border-white/10 shadow-sm"
              :class="isHighPressure ? 'bg-rose-600/60' : 'bg-black/40'">
              <span>{{ blindBoxScenario.icon }}</span>
              {{ blindBoxScenario.name }}
            </div>
            <div v-else class="bg-black/40 text-white/90 text-xs px-3 py-1.5 rounded-full backdrop-blur-md border border-white/10 shadow-sm">
              多模态情绪监测中...
            </div>
          </div>

          <!-- Question Timer (top-right, for timed scenarios) -->
          <div v-if="questionTimer > 0" class="absolute top-6 right-6 z-10 pointer-events-none">
            <div class="flex items-center gap-2 px-4 py-2 rounded-full backdrop-blur-md shadow-lg border"
              :class="questionTimer <= 10 
                ? 'bg-rose-600/80 border-rose-400 text-white animate-pulse' 
                : questionTimer <= 30 
                  ? 'bg-amber-500/70 border-amber-300 text-white' 
                  : 'bg-black/40 border-white/10 text-white/90'">
              <Timer class="w-4 h-4" />
              <span class="text-lg font-mono font-black tracking-wider">
                {{ Math.floor(questionTimer / 60) }}:{{ (questionTimer % 60).toString().padStart(2, '0') }}
              </span>
            </div>
          </div>

          <!-- High-pressure overlay vignette -->
          <div v-if="isHighPressure" class="absolute inset-0 pointer-events-none z-[5]"
            :class="pressureLevel === 'extreme' 
              ? 'shadow-[inset_0_0_80px_rgba(220,38,38,0.25)]' 
              : 'shadow-[inset_0_0_60px_rgba(220,38,38,0.1)]'"
          ></div>
          
          <div class="absolute top-20 right-6 z-10 px-3 py-1.5 rounded-full text-[11px] font-semibold border backdrop-blur-md"
            :class="isOffline ? 'bg-amber-500/70 text-white border-amber-300' : 'bg-black/40 text-white/90 border-white/10'">
            {{ isOffline ? '网络重连中' : `状态 ${sessionState}` }}
          </div>

          <template v-if="isVideoSceneMode">
            <div class="w-full h-full bg-gradient-to-br from-indigo-900 via-violet-900 to-zinc-900 flex items-center justify-center">
              <div class="relative w-[320px] h-[360px] rounded-[40px] bg-white/10 border border-white/20 shadow-2xl backdrop-blur-md flex flex-col items-center justify-center">
                <div class="w-32 h-32 rounded-full bg-white/15 border border-white/20 flex items-center justify-center">
                  <BrainCircuit class="w-16 h-16 text-white/90" />
                </div>
                <div class="mt-6 px-6 text-center">
                  <p class="text-white text-lg font-bold">AI 面试官</p>
                  <p class="text-white/70 text-xs mt-1">语音播报 + 动作反馈</p>
                </div>
                <div class="mt-5 w-20 h-3 rounded-full bg-white/20 overflow-hidden">
                  <div class="h-full bg-emerald-300 transition-all duration-150"
                    :style="{ width: interviewerSpeaking ? '100%' : '20%' }"></div>
                </div>
                <div class="mt-4 text-[11px] px-3 py-1 rounded-full border transition-all duration-300"
                  :class="answerVoiceStatus === 'recording' 
                    ? 'bg-emerald-500/20 text-emerald-100 border-emerald-500/50 animate-pulse ring-1 ring-emerald-500/30' 
                    : 'text-white/80 border-white/15 bg-black/20'">
                  {{ answerVoiceStatus === 'recording' 
                      ? '🎤 正在倾听...' 
                      : (interviewerMotion === 'talking' ? '口型同步中' : interviewerMotion === 'nod' ? '点头反馈' : interviewerMotion === 'thinking' ? '思考中...' : '待机') 
                  }}
                </div>
              </div>
            </div>

            <div class="absolute bottom-6 right-6 w-56 h-36 rounded-2xl overflow-hidden border border-white/20 bg-zinc-900 shadow-2xl">
              <video ref="interviewVideo" class="w-full h-full object-cover transform scale-x-[-1]" autoplay muted v-if="isCameraOn"></video>
              <div v-else class="w-full h-full flex flex-col items-center justify-center text-zinc-400 bg-zinc-900/80">
                <VideoOff class="h-8 w-8 mb-1" />
                <span class="text-xs">候选人画面不可用</span>
              </div>
              <div class="absolute top-2 left-2 text-[10px] px-2 py-0.5 rounded-full bg-black/50 text-white">候选人</div>
            </div>
          </template>
          <template v-else>
            <video ref="interviewVideo" class="w-full h-full object-cover transform scale-x-[-1]" autoplay muted v-if="isCameraOn"></video>
            <div v-else class="w-full h-full flex flex-col items-center justify-center text-zinc-600 bg-zinc-900/50">
              <User class="h-24 w-24 mb-6 opacity-20" />
              <p class="font-medium tracking-wide opacity-50">摄像头已关闭</p>
            </div>
          </template>

          <!-- Controls (Bottom Center - Auto hide) -->
          <div class="absolute bottom-8 left-1/2 -translate-x-1/2 flex gap-4 transition-all duration-500 translate-y-4 opacity-0 group-hover:translate-y-0 group-hover:opacity-100">
            <button 
              @click="toggleMic"
              class="h-12 w-12 rounded-full flex items-center justify-center backdrop-blur-md transition-all hover:scale-110 shadow-lg border border-white/10"
              :class="isMicOn ? 'bg-white/20 text-white hover:bg-white/30' : 'bg-rose-500 text-white'"
            >
              <Mic v-if="isMicOn" class="h-5 w-5" />
              <MicOff v-else class="h-5 w-5" />
            </button>
            <button 
              @click="toggleCamera"
              class="h-12 w-12 rounded-full flex items-center justify-center backdrop-blur-md transition-all hover:scale-110 shadow-lg border border-white/10"
              :class="isCameraOn ? 'bg-white/20 text-white hover:bg-white/30' : 'bg-rose-500 text-white'"
            >
              <Video v-if="isCameraOn" class="h-5 w-5" />
              <VideoOff v-else class="h-5 w-5" />
            </button>
          </div>
        </div>

        <!-- Transcript / Input Section (Bottom) -->
        <div class="bg-white rounded-3xl p-6 shadow-xl shadow-zinc-200/50 border border-white flex flex-col relative transition-all duration-300 focus-within:ring-2 focus-within:ring-indigo-500/20 focus-within:shadow-indigo-500/10 group lg:resizable-panel lg:flex-none"
             :class="isVideoSceneMode ? 'h-auto min-h-0' : 'h-1/3 min-h-[200px] lg:h-[32vh]'">
           <div class="flex justify-between items-center mb-4">
             <h3 class="font-bold text-zinc-900 flex items-center gap-2 group-focus-within:text-indigo-600 transition-colors">
               <div class="w-1.5 h-4 bg-zinc-300 rounded-full group-focus-within:bg-indigo-600 transition-colors"></div>
               实时回答转写
               <span v-if="userInput.length > 0 && !isVideoSceneMode" class="text-xs font-normal text-emerald-600 flex items-center gap-1 bg-emerald-50 px-2 py-0.5 rounded-full animate-in fade-in zoom-in duration-300">
                 <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
                 正在输入...
               </span>
             </h3>
             <div class="flex items-center gap-2">
               <span class="text-[11px] font-medium px-2 py-1 rounded-full border"
                 :class="answerVoiceStatus === 'recording'
                   ? 'bg-rose-50 text-rose-600 border-rose-200'
                   : answerVoiceStatus === 'success'
                     ? 'bg-emerald-50 text-emerald-600 border-emerald-200'
                     : answerVoiceStatus === 'error'
                       ? 'bg-amber-50 text-amber-600 border-amber-200'
                       : 'bg-zinc-50 text-zinc-500 border-zinc-200'">
                 {{ getVoiceStatusLabel() }}
               </span>
               <button 
                  @click.stop="showHistory = true" 
                  class="text-xs text-zinc-400 hover:text-indigo-600 transition-colors flex items-center gap-1 px-2 py-1 hover:bg-zinc-50 rounded-lg"
               >
                 <History class="w-3 h-3" /> 历史记录
               </button>
             </div>
           </div>
           
           <div class="rounded-xl border border-emerald-100 bg-emerald-50/40 px-3 py-2" :class="isVideoSceneMode ? '' : 'mb-3'">
             <div class="flex items-center justify-between gap-2 mb-1">
               <span class="text-[11px] font-semibold text-emerald-700">实时ASR转写（最近片段）</span>
               <span class="text-[10px] text-emerald-600" v-if="speechLastUpdatedAt">
                 {{ new Date(speechLastUpdatedAt).toLocaleTimeString() }}
               </span>
             </div>
             <p class="text-xs text-emerald-900/90 leading-relaxed min-h-[18px]">
               {{ latestChunkTranscript || '等待语音输入...' }}
             </p>
             <p class="text-[10px] text-emerald-700/80 mt-1 truncate" v-if="liveTranscript">
               累计片段：{{ liveTranscript }}
             </p>
           </div>

           <textarea 
              v-if="!isVideoSceneMode"
              v-model="userInput" 
              @keydown.ctrl.enter="sendMessage"
              placeholder="在此处输入您的回答..."
              class="flex-1 w-full resize-none border-none focus:ring-0 p-4 text-lg text-zinc-700 placeholder:text-zinc-300 bg-zinc-50/50 rounded-xl leading-relaxed transition-all focus:bg-white focus:shadow-inner custom-scrollbar"
           ></textarea>

           <div v-if="!isVideoSceneMode" class="absolute bottom-8 right-8 text-[10px] font-medium text-zinc-300 pointer-events-none bg-white/80 backdrop-blur px-2 py-1 rounded-md border border-zinc-100">
             Ctrl + Enter 发送
           </div>
        </div>
      </div>

      <!-- Right Sidebar -->
      <div class="w-full lg:w-[400px] flex flex-col gap-4 shrink-0 lg:h-full lg:min-h-0 lg:overflow-y-auto custom-scrollbar">
        <!-- AI Profile Card -->
        <div class="bg-white p-4 rounded-3xl border border-white shadow-lg shadow-zinc-200/50 flex items-center gap-4 hover:shadow-xl transition-shadow duration-300 shrink-0">
          <div class="h-14 w-14 rounded-2xl bg-gradient-to-br from-indigo-600 to-violet-600 flex items-center justify-center text-white shadow-lg shadow-indigo-500/30 ring-4 ring-indigo-50">
            <BrainCircuit class="h-7 w-7" />
          </div>
          <div>
            <h3 class="font-bold text-zinc-900 text-lg">智聘智能引擎</h3>
            <p class="text-xs text-zinc-500 font-medium flex items-center gap-1">
              <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]"></span>
              <span v-if="settings.interviewMode === 'random'">🎲 随机面试模式</span>
              <span v-else>{{ settings.mode === 'hr' ? 'HR面试官' : settings.mode === 'comprehensive' ? '综合面试官' : 'AI 技术面试官' }} · {{ settings.style === 'gentle' ? '温和型' : settings.style === 'stress' ? '压力型' : settings.style === 'deep' ? '深挖型' : settings.style === 'practical' ? '实战型' : settings.style === 'algorithm' ? '算法型' : '标准' }}</span>
            </p>
          </div>
        </div>

        <!-- Blind Box Scenario Banner (during interview) -->
        <div v-if="blindBoxScenario" 
          class="p-3 rounded-2xl border shadow-sm shrink-0 flex items-center gap-3 animate-in fade-in duration-500"
          :class="[pressureColors[pressureLevel]?.bg, pressureColors[pressureLevel]?.border]">
          <span class="text-2xl">{{ blindBoxScenario.icon }}</span>
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2">
              <span class="text-sm font-bold" :class="pressureColors[pressureLevel]?.text">{{ blindBoxScenario.name }}</span>
              <span class="text-[10px] font-bold px-1.5 py-0.5 rounded-full" :class="pressureColors[pressureLevel]?.badge">
                {{ pressureLabels[pressureLevel] }}
              </span>
            </div>
            <p class="text-[10px] text-zinc-500 truncate mt-0.5">{{ blindBoxScenario.description }}</p>
          </div>
          <div v-if="questionTimer > 0" class="flex items-center gap-1 px-2 py-1 rounded-lg shrink-0"
            :class="questionTimer <= 10 ? 'bg-rose-200 text-rose-800 animate-pulse' : 'bg-white/60 text-zinc-600'">
            <Timer class="w-3 h-3" />
            <span class="text-xs font-mono font-bold">{{ Math.floor(questionTimer / 60) }}:{{ (questionTimer % 60).toString().padStart(2, '0') }}</span>
          </div>
        </div>

        <!-- Question / Feedback Card -->
        <div class="bg-white rounded-3xl border shadow-xl shadow-zinc-200/50 flex-[1.6] min-h-64 flex flex-col relative overflow-hidden group transition-all duration-300 hover:shadow-2xl hover:shadow-zinc-200/60 lg:resizable-panel lg:flex-none lg:h-[46vh]"
          :class="isHighPressure ? 'border-rose-100' : 'border-white'">
           <!-- Card Header -->
           <div class="px-6 py-5 border-b flex justify-between items-center backdrop-blur-sm z-10"
             :class="isHighPressure ? 'border-rose-50 bg-rose-50/30' : 'border-zinc-50 bg-zinc-50/50'">
            <div class="inline-flex items-center gap-1.5 px-3 py-1 text-xs font-bold rounded-full border"
               :class="isHighPressure ? 'bg-rose-50 text-rose-700 border-rose-100/50' : 'bg-indigo-50 text-indigo-700 border-indigo-100/50'">
               <span class="w-1.5 h-1.5 rounded-full" :class="isHighPressure ? 'bg-rose-600' : 'bg-indigo-600'"></span>
              当前题目 · 第 {{ currentQuestionIndex + 1 }} 题
             </div>
            <div v-if="isProcessing" class="flex items-center gap-2 text-indigo-600 animate-pulse">
               <Loader2 class="w-4 h-4 animate-spin" />
               <span class="text-xs font-medium">{{ processingHint || '面试官正在评估...' }}</span>
            </div>
            <div v-else-if="latestAIMessage?.type === 'feedback'" class="flex items-center gap-2 animate-in fade-in slide-in-from-right duration-500">
                <span class="text-xs text-zinc-400 font-medium">评分</span>
                <span class="text-xl font-black text-indigo-600 tracking-tight">{{ latestAIMessage.score }}</span>
             </div>
           </div>
           
           <!-- Content Area -->
           <div class="flex-1 min-h-0 overflow-y-auto p-6 custom-scrollbar relative">
             <!-- Loading State -->
             <div v-if="isProcessing" class="absolute inset-0 flex flex-col items-center justify-center text-zinc-400 gap-3 bg-white/80 backdrop-blur-sm z-20">
                <div class="relative">
                  <div class="absolute inset-0 bg-indigo-500/20 blur-xl rounded-full"></div>
                  <Loader2 class="h-10 w-10 animate-spin text-indigo-600 relative z-10" />
                </div>
                <p class="text-sm font-medium animate-pulse">{{ processingHint || '正在生成评估...' }}</p>
             </div>

             <!-- Content -->
             <div v-else class="space-y-4 animate-in fade-in slide-in-from-bottom-4 duration-500">
               <!-- If it's a Question -->
               <template v-if="latestAIMessage?.type === 'question' || (latestAIMessage?.role === 'ai' && !latestAIMessage?.type)">
                 <h2 class="text-xl font-bold text-zinc-900 leading-relaxed tracking-wide whitespace-pre-wrap wrap-break-word">
                   {{ latestAIMessage?.content }}
                 </h2>
               </template>

               <!-- If it's Feedback -->
               <template v-else-if="latestAIMessage?.type === 'feedback'">
                 <div class="space-y-3">
                   <!-- 综合评价 -->
                   <div class="p-4 bg-gradient-to-br from-amber-50 to-orange-50/30 rounded-2xl border border-amber-100/60 shadow-sm">
                      <h4 class="text-xs font-bold text-amber-600 uppercase mb-2 flex items-center gap-2">
                        <div class="p-1 bg-amber-100 rounded-md">
                          <MessageSquare class="w-3.5 h-3.5" />
                        </div>
                        综合评价
                      </h4>
                      <p class="text-sm text-zinc-800 leading-relaxed text-justify whitespace-pre-wrap wrap-break-word">{{ latestAIMessage.feedbackEvaluation }}</p>
                    </div>

                    <!-- 维度评分 -->
                    <div v-if="latestAIMessage.feedbackDimensions" class="p-4 bg-gradient-to-br from-indigo-50/80 to-violet-50/30 rounded-2xl border border-indigo-100/50 shadow-sm">
                      <h4 class="text-xs font-bold text-indigo-600 uppercase mb-3 flex items-center gap-2">
                        <div class="p-1 bg-indigo-100 rounded-md">
                          <BarChart3 class="w-3.5 h-3.5" />
                        </div>
                        维度评分
                      </h4>
                      <div class="space-y-2.5">
                        <div v-for="dim in [
                          { key: 'technical_depth', label: '技术深度', color: 'bg-violet-500' },
                          { key: 'expression', label: '表达清晰', color: 'bg-blue-500' },
                          { key: 'logic', label: '逻辑严谨', color: 'bg-cyan-500' },
                          { key: 'completeness', label: '覆盖完整', color: 'bg-emerald-500' }
                        ]" :key="dim.key" class="flex items-center gap-3">
                          <span class="text-xs text-zinc-500 w-14 shrink-0 text-right font-medium">{{ dim.label }}</span>
                          <div class="flex-1 h-2 bg-zinc-100 rounded-full overflow-hidden">
                            <div :class="dim.color" class="h-full rounded-full transition-all duration-700 ease-out" :style="{ width: (latestAIMessage.feedbackDimensions[dim.key] || 0) + '%' }"></div>
                          </div>
                          <span class="text-xs font-bold text-zinc-700 w-8 shrink-0">{{ latestAIMessage.feedbackDimensions[dim.key] || 0 }}</span>
                        </div>
                      </div>
                    </div>

                    <!-- 亮点 & 差距并排 -->
                    <div v-if="(latestAIMessage.feedbackHighlights?.length || latestAIMessage.feedbackGaps?.length)" class="grid grid-cols-2 gap-2">
                      <!-- 亮点 -->
                      <div v-if="latestAIMessage.feedbackHighlights?.length" class="p-3 bg-emerald-50/80 rounded-xl border border-emerald-100/50">
                        <h4 class="text-[10px] font-bold text-emerald-600 uppercase mb-2 flex items-center gap-1">
                          <CheckCircle class="w-3 h-3" /> 亮点
                        </h4>
                        <ul class="space-y-1">
                          <li v-for="(h, i) in latestAIMessage.feedbackHighlights" :key="i" class="text-xs text-emerald-800 leading-relaxed flex gap-1.5">
                            <span class="text-emerald-400 mt-0.5 shrink-0">✦</span>
                            <span>{{ h }}</span>
                          </li>
                        </ul>
                      </div>
                      <!-- 差距 -->
                      <div v-if="latestAIMessage.feedbackGaps?.length" class="p-3 bg-rose-50/80 rounded-xl border border-rose-100/50">
                        <h4 class="text-[10px] font-bold text-rose-600 uppercase mb-2 flex items-center gap-1">
                          <AlertTriangle class="w-3 h-3" /> 待补强
                        </h4>
                        <ul class="space-y-1">
                          <li v-for="(g, i) in latestAIMessage.feedbackGaps" :key="i" class="text-xs text-rose-800 leading-relaxed flex gap-1.5">
                            <span class="text-rose-400 mt-0.5 shrink-0">△</span>
                            <span>{{ g }}</span>
                          </li>
                        </ul>
                      </div>
                    </div>
                    
                    <!-- 改进建议 -->
                    <div class="p-4 bg-gradient-to-br from-emerald-50 to-teal-50/30 rounded-2xl border border-emerald-100/60 shadow-sm">
                      <h4 class="text-xs font-bold text-emerald-600 uppercase mb-2 flex items-center gap-2">
                        <div class="p-1 bg-emerald-100 rounded-md">
                          <Lightbulb class="w-3.5 h-3.5" />
                        </div>
                        改进建议
                      </h4>
                     <ul class="space-y-2">
                       <li v-for="(s, i) in latestAIMessage.feedbackSuggestions" :key="i" class="text-xs text-emerald-900 flex gap-2.5 leading-relaxed group/item wrap-break-word">
                         <span class="font-bold text-emerald-600/40 font-mono text-[10px] mt-0.5 group-hover/item:text-emerald-600 transition-colors shrink-0">0{{ i + 1 }}</span>
                         {{ s }}
                       </li>
                     </ul>
                   </div>

                   <!-- 参考答案思路（可折叠） -->
                   <div v-if="latestAIMessage.feedbackModelAnswer" class="p-4 bg-gradient-to-br from-sky-50/80 to-blue-50/30 rounded-2xl border border-sky-100/50 shadow-sm">
                      <h4 class="text-xs font-bold text-sky-600 uppercase mb-2 flex items-center gap-2 cursor-pointer select-none" @click="showModelAnswer = !showModelAnswer">
                        <div class="p-1 bg-sky-100 rounded-md">
                          <BookOpen class="w-3.5 h-3.5" />
                        </div>
                        参考答案思路
                        <ChevronDown class="w-3 h-3 ml-auto transition-transform duration-200" :class="showModelAnswer ? 'rotate-180' : ''" />
                      </h4>
                      <p v-show="showModelAnswer" class="text-xs text-zinc-700 leading-relaxed whitespace-pre-wrap wrap-break-word mt-1 animate-in fade-in slide-in-from-top-2 duration-300">{{ latestAIMessage.feedbackModelAnswer }}</p>
                    </div>

                    <!-- 追问方向 -->
                    <div v-if="latestAIMessage.feedbackFollowUp" class="p-3 bg-zinc-50 rounded-xl border border-zinc-100">
                      <p class="text-xs text-zinc-500 flex items-start gap-2">
                        <span class="text-indigo-400 font-bold shrink-0 mt-0.5">💬</span>
                        <span><span class="font-medium text-zinc-600">面试官可能追问：</span>{{ latestAIMessage.feedbackFollowUp }}</span>
                      </p>
                    </div>
                 </div>
               </template>
               
               <!-- System Message -->
               <template v-else-if="latestAIMessage?.type === 'system'">
                  <div class="p-6 bg-zinc-50 rounded-2xl text-center text-zinc-600 text-sm border border-zinc-100">
                    <p class="mb-4">{{ latestAIMessage.content }}</p>
                    <div v-if="messages[messages.length-1].content.includes('面试结束')" class="flex justify-center">
                       <button @click="viewReport" class="px-8 py-3 bg-indigo-600 text-white rounded-xl text-sm font-bold hover:bg-indigo-700 transition-all shadow-lg shadow-indigo-200 hover:shadow-indigo-300 hover:-translate-y-0.5 active:translate-y-0">
                         查看详细报告
                       </button>
                    </div>
                  </div>
               </template>
             </div>
           </div>
        </div>

        <!-- Hint Card / Shadow Coach -->
        <div v-if="shadowCoachEnabled" class="bg-gradient-to-br from-emerald-50/80 to-white p-4 rounded-3xl border border-white shadow-lg shadow-zinc-200/30 backdrop-blur-sm shrink-0 lg:resizable-panel lg:flex-none lg:h-[170px]">
          <h4 class="text-xs font-bold text-emerald-600 uppercase mb-3 flex items-center gap-2">
            <Headphones class="w-3.5 h-3.5" />
            AI 影子教练 · 实时耳返
          </h4>
          <p class="text-sm text-zinc-600 italic leading-relaxed opacity-80">
            "建议从 STAR 原则出发，重点描述你在项目中遇到的挑战以及你是如何克服它的。"
          </p>
        </div>

        <!-- Real-time Speech Dashboard -->
        <div class="bg-white rounded-3xl p-4 border border-zinc-100 shadow-sm shrink-0 lg:resizable-panel lg:flex-none lg:h-[250px]">
          <SpeechDashboard
            :speechRate="speechMetrics.speechRate"
            :speechRateLevel="speechMetrics.speechRateLevel"
            :energyLevel="energyLevel"
            :fillerWordCount="speechMetrics.fillerWordCount"
            :fluencyAlert="speechMetrics.fluencyAlert"
            :totalFillerWords="speechMetrics.totalFillerWords"
            :isActive="speechAnalysisActive"
          />
        </div>

        <button
          @click="toggleAnswerRecording"
          :disabled="isProcessing || answerVoiceStatus === 'requesting' || answerVoiceStatus === 'transcribing' || answerVoiceStatus === 'submitting'"
          class="w-full py-3 rounded-2xl font-bold text-base transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2 border shrink-0"
          :class="answerVoiceStatus === 'recording'
            ? 'bg-rose-600 text-white border-rose-600 hover:bg-rose-700'
            : 'bg-white text-zinc-700 border-zinc-200 hover:bg-zinc-50'"
        >
          <Mic v-if="answerVoiceStatus !== 'recording'" class="h-4 w-4" />
          <MicOff v-else class="h-4 w-4" />
          <span v-if="answerVoiceStatus === 'recording'">停止录音并提交</span>
          <span v-else>语音回答</span>
        </button>

        <!-- Action Button -->
        <button 
          @click="sendMessage"
          v-if="!isVideoSceneMode || latestAIMessage?.type === 'feedback' || pendingEnd"
          :disabled="isProcessing || (!userInput.trim() && latestAIMessage?.type !== 'feedback')"
          class="w-full py-3 bg-zinc-900 text-white rounded-2xl font-bold text-base hover:bg-zinc-800 hover:shadow-xl hover:shadow-zinc-900/20 active:scale-[0.98] transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-3 group relative overflow-hidden shrink-0"
        >
          <div class="absolute inset-0 bg-white/10 translate-y-full group-hover:translate-y-0 transition-transform duration-300"></div>
          <span v-if="isProcessing" class="flex items-center gap-2 relative z-10">
            <Loader2 class="w-5 h-5 animate-spin" />
            {{ processingHint || '正在思考...' }}
          </span>
          <span v-else-if="latestAIMessage?.type === 'feedback'" class="flex items-center gap-2 relative z-10">
            {{ pendingEnd ? '结束面试' : '下一题' }} <ChevronRight class="w-5 h-5 group-hover:translate-x-1 transition-transform" />
          </span>
          <span v-else class="flex items-center gap-2 relative z-10">
            发送回答 
            <Send class="h-4 w-4 group-hover:translate-x-1 transition-transform" />
          </span>
        </button>
      </div>

      <!-- History Drawer (Overlay) -->
      <div v-if="showHistory" class="absolute inset-0 z-50 bg-black/20 backdrop-blur-sm flex justify-end" @click.self="showHistory = false">
        <div class="w-96 bg-white h-full shadow-2xl animate-in slide-in-from-right duration-300 flex flex-col border-l border-zinc-100">
          <div class="p-5 border-b border-zinc-100 flex justify-between items-center bg-zinc-50/50">
            <h3 class="font-bold text-zinc-900 flex items-center gap-2">
              <History class="w-4 h-4 text-zinc-400" />
              对话历史
            </h3>
            <button @click="showHistory = false" class="p-2 hover:bg-zinc-200/50 rounded-full transition-colors text-zinc-400 hover:text-zinc-600">
              <ChevronRight class="h-5 w-5" />
            </button>
          </div>
          <div class="flex-1 overflow-y-auto p-4 space-y-4 custom-scrollbar bg-zinc-50/30">
            <div v-for="(msg, i) in messages" :key="i" class="text-sm p-4 rounded-2xl border shadow-sm transition-all hover:shadow-md" 
              :class="msg.role === 'user' ? 'bg-white border-zinc-100 text-zinc-800 ml-4' : 'bg-indigo-50/50 border-indigo-100 text-zinc-800 mr-4'">
              <div class="text-[10px] uppercase tracking-wider font-bold mb-2 flex items-center gap-1" 
                :class="msg.role === 'user' ? 'text-zinc-400 justify-end' : 'text-indigo-400'">
                <User v-if="msg.role === 'user'" class="w-3 h-3" />
                <BrainCircuit v-else class="w-3 h-3" />
                {{ msg.role === 'ai' ? 'AI 面试官' : '你' }}
              </div>
              <div class="leading-relaxed whitespace-pre-wrap">{{ msg.content }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- Random Style Reveal Banner (shown after interview ends in random mode) -->
      <div v-if="randomStyleRevealed && revealedStyleInfo" class="fixed bottom-6 left-1/2 -translate-x-1/2 z-50 animate-in fade-in slide-in-from-bottom-4 duration-500">
        <div class="bg-white rounded-2xl shadow-2xl border border-violet-200 p-5 min-w-[360px] max-w-md">
          <div class="flex items-center gap-3 mb-3">
            <div class="p-2 bg-violet-100 rounded-xl">
              <Shuffle class="w-5 h-5 text-violet-600" />
            </div>
            <div>
              <h4 class="font-bold text-zinc-900">🎲 随机面试风格揭晓！</h4>
              <p class="text-xs text-zinc-500">本次面试采用的隐藏风格</p>
            </div>
            <button @click="randomStyleRevealed = false" class="ml-auto p-1 hover:bg-zinc-100 rounded-lg transition-colors">
              <X class="w-4 h-4 text-zinc-400" />
            </button>
          </div>
          <div class="flex gap-3">
            <div class="flex-1 p-3 bg-violet-50 rounded-xl border border-violet-100">
              <p class="text-[10px] text-violet-500 font-bold uppercase mb-1">面试官风格</p>
              <p class="text-sm font-bold text-violet-700">{{ revealedStyleInfo.style_label }}</p>
            </div>
            <div v-if="revealedStyleInfo.company_label" class="flex-1 p-3 bg-orange-50 rounded-xl border border-orange-100">
              <p class="text-[10px] text-orange-500 font-bold uppercase mb-1">匹配公司</p>
              <p class="text-sm font-bold text-orange-700">{{ revealedStyleInfo.company_label }}</p>
            </div>
          </div>
        </div>
      </div>

    </div>

    <!-- ===== Booking Dialog (Overlay) ===== -->
    <div v-if="showBookingDialog && selectedInterviewer" class="fixed inset-0 z-[60] bg-black/30 backdrop-blur-sm flex items-center justify-center" @click.self="showBookingDialog = false">
      <div class="bg-white rounded-2xl shadow-2xl border border-zinc-100 p-6 w-[420px] max-w-[90vw] animate-in fade-in zoom-in-95 duration-300">
        <div class="flex items-center justify-between mb-5">
          <h3 class="font-bold text-lg text-zinc-900">预约面试</h3>
          <button @click="showBookingDialog = false" class="p-2 hover:bg-zinc-100 rounded-lg transition-colors">
            <X class="w-4 h-4 text-zinc-400" />
          </button>
        </div>

        <!-- Interviewer Info -->
        <div class="flex items-center gap-3 p-3 bg-zinc-50 rounded-xl mb-4">
          <div class="h-12 w-12 rounded-full bg-gradient-to-br from-indigo-100 to-purple-100 flex items-center justify-center text-indigo-700 font-bold shrink-0">
            {{ selectedInterviewer.name?.[0] || '?' }}
          </div>
          <div>
            <p class="font-bold text-zinc-800">{{ selectedInterviewer.name }}</p>
            <p class="text-xs text-zinc-500">{{ selectedInterviewer.title }}{{ selectedInterviewer.company ? ` · ${selectedInterviewer.company}` : '' }}</p>
          </div>
        </div>

        <!-- Booking Form -->
        <div class="space-y-3">
          <div>
            <label class="text-xs font-bold text-zinc-500 mb-1 block">预约时间</label>
            <input 
              type="datetime-local" 
              v-model="bookingForm.scheduledAt" 
              class="w-full bg-zinc-50 border border-zinc-200 rounded-xl px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            />
          </div>
          <div>
            <label class="text-xs font-bold text-zinc-500 mb-1 block">备注（可选）</label>
            <textarea 
              v-model="bookingForm.notes" 
              placeholder="如：希望重点考察微服务架构设计能力"
              class="w-full bg-zinc-50 border border-zinc-200 rounded-xl px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none h-20"
            ></textarea>
          </div>
        </div>

        <button 
          @click="submitBooking"
          :disabled="!bookingForm.scheduledAt"
          class="w-full mt-4 py-3 bg-indigo-600 text-white rounded-xl font-bold text-sm hover:bg-indigo-700 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
        >
          确认预约
        </button>
      </div>
    </div>

    <!-- ===== Bookings Panel (Overlay) ===== -->
    <div v-if="showBookingsPanel" class="fixed inset-0 z-[60] bg-black/30 backdrop-blur-sm flex items-center justify-center" @click.self="showBookingsPanel = false">
      <div class="bg-white rounded-2xl shadow-2xl border border-zinc-100 p-6 w-[480px] max-w-[90vw] max-h-[70vh] flex flex-col animate-in fade-in zoom-in-95 duration-300">
        <div class="flex items-center justify-between mb-4">
          <h3 class="font-bold text-lg text-zinc-900 flex items-center gap-2">
            <Calendar class="w-5 h-5 text-indigo-600" />
            我的面试预约
          </h3>
          <button @click="showBookingsPanel = false" class="p-2 hover:bg-zinc-100 rounded-lg transition-colors">
            <X class="w-4 h-4 text-zinc-400" />
          </button>
        </div>

        <div class="flex-1 overflow-y-auto space-y-3 custom-scrollbar">
          <div v-if="userBookings.length === 0" class="text-center py-8">
            <Calendar class="w-10 h-10 text-zinc-200 mx-auto mb-3" />
            <p class="text-sm text-zinc-400">暂无预约记录</p>
          </div>
          <div v-for="booking in userBookings" :key="booking.id" class="p-4 rounded-xl border border-zinc-100 hover:border-zinc-200 transition-all">
            <div class="flex items-center justify-between mb-2">
              <span class="text-sm font-bold text-zinc-800">{{ booking.interviewer?.name || '面试官' }}</span>
              <span class="text-[10px] px-2 py-0.5 rounded-full font-bold"
                :class="{
                  'bg-amber-100 text-amber-700': booking.status === 'pending',
                  'bg-emerald-100 text-emerald-700': booking.status === 'confirmed',
                  'bg-blue-100 text-blue-700': booking.status === 'completed',
                  'bg-zinc-100 text-zinc-500': booking.status === 'cancelled'
                }">
                {{ booking.status === 'pending' ? '待确认' : booking.status === 'confirmed' ? '已确认' : booking.status === 'completed' ? '已完成' : '已取消' }}
              </span>
            </div>
            <div class="text-xs text-zinc-500 space-y-1">
              <p class="flex items-center gap-1.5"><Clock class="w-3 h-3" /> {{ new Date(booking.scheduled_at).toLocaleString('zh-CN') }}</p>
              <p class="flex items-center gap-1.5"><Briefcase class="w-3 h-3" /> {{ booking.position }} · {{ booking.difficulty }}</p>
              <p v-if="booking.notes" class="flex items-start gap-1.5"><MessageSquare class="w-3 h-3 mt-0.5 shrink-0" /> {{ booking.notes }}</p>
            </div>
          </div>
        </div>
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
</style>
