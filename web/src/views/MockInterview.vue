<script setup>
import { ref, onMounted, onUnmounted, nextTick, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { 
  Mic, MicOff, ChevronRight, ChevronDown,
  BrainCircuit, User, LogOut, Send, Loader2,
  History, MessageSquare, Lightbulb,
  Monitor, Users, Shuffle, UserCheck, Shield, Headphones,
  Heart, Eye, Brain, Volume2, BarChart3, CheckCircle, AlertTriangle, BookOpen,
  Package, Timer, Zap, Building2, Star, Calendar, Clock, X,
  Flame, Search, Code, Briefcase, GraduationCap
} from 'lucide-vue-next'
import { startInterview as apiStartInterview, submitAnswer as apiSubmitAnswer, endInterview as apiEndInterview, uploadInterviewRecording as apiUploadInterviewRecording, analyzeSpeechChunk as apiAnalyzeSpeechChunk, getShadowCoachHint as apiGetShadowCoachHint, drawBlindBoxScenario as apiDrawBlindBox, getInterviewConfig as apiGetInterviewConfig, revealRandomStyle as apiRevealRandomStyle, synthesizeInterviewSpeech as apiSynthesizeInterviewSpeech, getInviteCandidates as apiGetInviteCandidates, createHumanInvitation as apiCreateHumanInvitation, getHumanInvitations as apiGetHumanInvitations } from '../api/interview'
import { generateReport as apiGenerateReport } from '../api/report'
import SpeechDashboard from '../components/SpeechDashboard.vue'
import AlgorithmInterviewPanel from '../components/AlgorithmInterviewPanel.vue'
import InterviewSetupPreview from '../components/InterviewSetupPreview.vue'
import InterviewSetupForm from '../components/InterviewSetupForm.vue'
import HumanInterviewModals from '../components/HumanInterviewModals.vue'
import InterviewVideoStage from '../components/InterviewVideoStage.vue'
import InterviewFeedbackSidebar from '../components/InterviewFeedbackSidebar.vue'

const route = useRoute()
const router = useRouter()
const phase = ref('setup') // setup, interview, summary
const isCameraOn = ref(true)
const isMicOn = ref(true)
const stream = ref(null)
const recordingStatus = ref('idle')
const recordingUrl = ref('')
const answerVoiceStatus = ref('idle') // idle, requesting, recording, transcribing, submitting, success, error
const answerVoiceSeconds = ref(0)
const answerVoiceError = ref('')
let interviewMediaRecorder = null
let interviewRecordedChunks = []
let interviewRecordingStream = null
let answerMediaRecorder = null
let answerAudioChunks = []
let answerVoiceTimer = null
let answerRecorderStream = null
let answerRecorderMimeType = ''

// Interview State
const interviewId = ref(null)
const questions = ref([])
const currentQuestionIndex = ref(0)
const currentQuestion = ref(null)
const messages = ref([])
const userInput = ref('')
const isProcessing = ref(false)
const processingHint = ref('')
const reportId = ref(null)
const isGeneratingReport = ref(false)
const showHistory = ref(false)
const pendingNextQuestion = ref(null)
const pendingEnd = ref(false)
const isAvatarSpeaking = ref(false)
let currentSpeechAudio = null

const latestAIMessage = computed(() => {
  const aiMsgs = messages.value.filter(m => m.role === 'ai' || m.type === 'system')
  return aiMsgs.length > 0 ? aiMsgs[aiMsgs.length - 1] : null
})

const canAnswerCurrentQuestion = computed(() => {
  if (phase.value !== 'interview') return false
  if (!currentQuestion.value?.questionId) return false
  if (isProcessing.value) return false
  if (isAvatarSpeaking.value) return false
  if (latestAIMessage.value?.type === 'feedback') return false
  // Need a visible question first; prevents answering during intro/system messages.
  return latestAIMessage.value?.type === 'question'
})

const latestUserTranscript = computed(() => {
  const userMsgs = messages.value.filter(m => m.role === 'user' && typeof m.rawTranscript === 'string' && m.rawTranscript.trim())
  return userMsgs.length > 0 ? userMsgs[userMsgs.length - 1].rawTranscript : ''
})

const isHumanInterviewMode = computed(() => settings.value.interviewMode === 'human')
const isVideoInterviewMode = computed(() => settings.value.presentationMode === 'video_avatar' && !isHumanInterviewMode.value)
const isAlgorithmStyle = computed(() => settings.value.style === 'algorithm')

const algorithmBriefText = ref('请用算法思维，在满足复杂度约束的情况下，实现如下算法题目。')
const algorithmProgress = ref({ current: 1, total: 0, finished: 0, passed: 0, skipped: 0, failed: 0 })

watch(showHistory, (visible) => {
  document.body.style.overflow = visible ? 'hidden' : ''
})

const settings = ref({
  position: route.query.position || 'Java后端工程师',
  difficulty: 'campus_intern',
  mode: route.query.mode || 'technical',
  style: 'gentle',
  company: '',
  interviewMode: 'ai',  // ai, human, random
  presentationMode: route.query.presentationMode || 'video_avatar' // text_voice, video_avatar
})

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
const shadowCoachEnabled = ref(true)
const shadowCoachHints = ref([])
const shadowCoachBubbleText = ref('')
const shadowCoachBubbleVisible = ref(false)
const shadowCoachHintPending = ref(false)
const modelViewerReady = ref(false)
const silenceStreakSeconds = ref(0)
const thinkingStreakSeconds = ref(0)
const quietSeconds = ref(0)
const shadowHintPack = ref([])
const shadowHintDelivered = ref([false, false, false])
const emotionFeedback = ref({ sentiment: '正常', confidence: 0, heartRate: 72 })
let shadowBubbleTimer = null
let thinkingWatchTimer = null

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

const hideShadowBubble = () => {
  shadowCoachBubbleVisible.value = false
  if (shadowBubbleTimer) {
    clearTimeout(shadowBubbleTimer)
    shadowBubbleTimer = null
  }
}

const showShadowBubble = (text) => {
  shadowCoachBubbleText.value = text
  shadowCoachBubbleVisible.value = true
  if (shadowBubbleTimer) clearTimeout(shadowBubbleTimer)
  shadowBubbleTimer = setTimeout(() => {
    shadowCoachBubbleVisible.value = false
  }, 9000)
}

const buildLocalHintDefaults = (questionText = '') => {
  const cleaned = String(questionText || '').replace(/[？?]/g, '').trim()
  const focus = cleaned ? cleaned.slice(0, 16) : '这道题'
  return [
    `先围绕“${focus}”抛一个判断，不用一次讲完。`,
    `把“${focus}”拆成两段：先原理，再补一个真实场景。`,
    `直接开口按“观点 -> 机制 -> 结果”讲，把“${focus}”串起来。`
  ]
}

const normalizeHintPack = (rawHints = [], questionText = '') => {
  const defaults = buildLocalHintDefaults(questionText)
  const cleaned = Array.isArray(rawHints)
    ? rawHints.map((item) => String(item || '').trim()).filter(Boolean)
    : []

  return [
    cleaned[0] || defaults[0],
    cleaned[1] || cleaned[0] || defaults[1],
    cleaned[2] || cleaned[1] || defaults[2]
  ]
}

const resetShadowHintProgress = () => {
  silenceStreakSeconds.value = 0
  thinkingStreakSeconds.value = 0
  quietSeconds.value = 0
  shadowHintPack.value = []
  shadowHintDelivered.value = [false, false, false]
}

const preloadShadowHintPack = async () => {
  if (!shadowCoachEnabled.value || !interviewId.value || !currentQuestion.value) return
  if (shadowHintPack.value.length >= 3 || shadowCoachHintPending.value) return

  shadowCoachHintPending.value = true
  try {
    const res = await apiGetShadowCoachHint(interviewId.value, {
      question: currentQuestion.value.content || currentQuestion.value.title || '',
      transcript: latestUserTranscript.value || '',
      expected_answer: currentQuestion.value.expectedAnswer || '',
      silence_seconds: 0
    })
    shadowHintPack.value = normalizeHintPack(
      res?.hints || (res?.hint ? [res.hint] : []),
      currentQuestion.value.content || currentQuestion.value.title || ''
    )
  } catch (err) {
    console.warn('preload shadow hint pack failed:', err)
    shadowHintPack.value = normalizeHintPack([], currentQuestion.value.content || currentQuestion.value.title || '')
  } finally {
    shadowCoachHintPending.value = false
  }
}

const maybeDispatchTieredHint = async () => {
  if (!shadowCoachEnabled.value || !interviewId.value || !currentQuestion.value) return
  if (phase.value !== 'interview' || isProcessing.value) return
  if (latestAIMessage.value?.type !== 'question') return

  if (shadowHintPack.value.length < 3) {
    await preloadShadowHintPack()
  }
  if (shadowHintPack.value.length < 3) return

  const milestones = [20, 40, 60]
  for (let i = 0; i < milestones.length; i += 1) {
    if (quietSeconds.value >= milestones[i] && !shadowHintDelivered.value[i]) {
      const hint = shadowHintPack.value[i]
      showShadowBubble(hint)
      shadowCoachHints.value.unshift({ text: hint, at: Date.now(), level: i + 1 })
      if (shadowCoachHints.value.length > 8) {
        shadowCoachHints.value = shadowCoachHints.value.slice(0, 8)
      }
      shadowHintDelivered.value = shadowHintDelivered.value.map((flag, idx) => (idx === i ? true : flag))
      break
    }
  }
}

const startThinkingWatch = () => {
  if (thinkingWatchTimer) return
  thinkingWatchTimer = setInterval(async () => {
    if (!shadowCoachEnabled.value || phase.value !== 'interview') {
      thinkingStreakSeconds.value = 0
      silenceStreakSeconds.value = 0
      quietSeconds.value = 0
      return
    }

    if (isProcessing.value || latestAIMessage.value?.type !== 'question' || !canAnswerCurrentQuestion.value) {
      thinkingStreakSeconds.value = 0
      silenceStreakSeconds.value = 0
      quietSeconds.value = 0
      return
    }

    if (answerVoiceStatus.value === 'recording') {
      thinkingStreakSeconds.value = 0
      if (energyLevel.value <= 0.035) {
        silenceStreakSeconds.value += 1
        quietSeconds.value += 1
      } else {
        silenceStreakSeconds.value = 0
        quietSeconds.value = 0
      }
    } else {
      silenceStreakSeconds.value = 0
      thinkingStreakSeconds.value += 1
      quietSeconds.value += 1
    }

    await maybeDispatchTieredHint()

    if (quietSeconds.value >= 65 || shadowHintDelivered.value.every(Boolean)) {
      quietSeconds.value = shadowHintDelivered.value.every(Boolean) ? quietSeconds.value : 60
    }
  }, 1000)
}

const stopThinkingWatch = () => {
  if (!thinkingWatchTimer) return
  clearInterval(thinkingWatchTimer)
  thinkingWatchTimer = null
  thinkingStreakSeconds.value = 0
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

// ===== Real-time Speech Metrics =====
const speechMetrics = ref({
  speechRate: 0,
  speechRateLevel: 'normal',
  fillerWordCount: 0,
  fluencyAlert: false,
  totalFillerWords: 0,
  transcribedText: ''
})
const energyLevel = ref(0)
const speechAnalysisActive = ref(false)
const speechRateSmoother = ref(0)
const answerRecordingPeakEnergy = ref(0)
const chunkTranscriptHistory = ref([])

const classifySpeechRateLevelClient = (rate) => {
  if (rate < 120) return 'slow'
  if (rate <= 240) return 'normal'
  return 'fast'
}

const normalizeChunkTranscript = (text = '') => {
  return String(text || '').replace(/\s+/g, ' ').trim()
}

const mergeChunkTranscript = (existing, incoming) => {
  const base = normalizeChunkTranscript(existing)
  const next = normalizeChunkTranscript(incoming)
  if (!next) return base
  if (!base) return next
  if (base.includes(next)) return base
  if (next.includes(base)) return next

  const maxOverlap = Math.min(base.length, next.length, 24)
  for (let overlap = maxOverlap; overlap >= 4; overlap -= 1) {
    if (base.slice(-overlap) === next.slice(0, overlap)) {
      return `${base}${next.slice(overlap)}`.trim()
    }
  }

  return `${base} ${next}`.trim()
}

// Audio chunk recording for speech analysis
let audioContext = null
let analyserNode = null
let chunkMediaRecorder = null
let chunkRecordingStream = null
let chunkInterval = null
let energyAnimFrame = null
let analysisSourceStream = null
const speechChunkSeconds = 6
let chunkRecorderMimeType = ''

const pickSupportedAudioMime = () => {
  const candidates = [
    'audio/webm;codecs=opus',
    'audio/webm',
    'audio/mp4',
    'audio/ogg;codecs=opus',
    'audio/ogg'
  ]
  if (typeof MediaRecorder === 'undefined' || typeof MediaRecorder.isTypeSupported !== 'function') {
    return ''
  }
  for (const mime of candidates) {
    if (MediaRecorder.isTypeSupported(mime)) return mime
  }
  return ''
}

const normalizeAudioMime = (mime) => {
  const raw = String(mime || '').trim().toLowerCase()
  if (!raw) return ''
  const semi = raw.indexOf(';')
  return semi > 0 ? raw.slice(0, semi) : raw
}

const startSpeechAnalysis = (sourceStream = null) => {
  const activeStream = sourceStream || answerRecorderStream || stream.value
  if (speechAnalysisActive.value || !activeStream) return
  const activeAudioTracks = activeStream.getAudioTracks()
  if (!activeAudioTracks.length) return

  analysisSourceStream = activeStream
  speechAnalysisActive.value = true

  // Set up Web Audio API for real-time energy
  audioContext = new (window.AudioContext || window.webkitAudioContext)()
  if (audioContext.state === 'suspended') {
    audioContext.resume().catch((err) => {
      console.warn('AudioContext resume failed:', err)
    })
  }
  const source = audioContext.createMediaStreamSource(activeStream)
  analyserNode = audioContext.createAnalyser()
  analyserNode.fftSize = 1024
  analyserNode.smoothingTimeConstant = 0.82
  source.connect(analyserNode)

  // Animate energy level
  const dataArray = new Uint8Array(analyserNode.fftSize)
  let smoothedEnergy = 0
  let noiseFloor = 0.003
  const updateEnergy = () => {
    if (!speechAnalysisActive.value) return
    analyserNode.getByteTimeDomainData(dataArray)
    let sumSquares = 0
    for (let i = 0; i < dataArray.length; i += 1) {
      const sample = (dataArray[i] - 128) / 128
      sumSquares += sample * sample
    }
    const rms = Math.sqrt(sumSquares / dataArray.length)
    noiseFloor = (noiseFloor * 0.992) + (rms * 0.008)
    const gated = Math.max(0, rms - (noiseFloor * 1.15))
    const normalized = Math.min(1, gated * 28)
    smoothedEnergy = (smoothedEnergy * 0.68) + (normalized * 0.32)
    energyLevel.value = Math.min(1, smoothedEnergy)
    if (answerVoiceStatus.value === 'recording') {
      answerRecordingPeakEnergy.value = Math.max(answerRecordingPeakEnergy.value, energyLevel.value)
    }
    energyAnimFrame = requestAnimationFrame(updateEnergy)
  }
  updateEnergy()

  // Start chunked recording: every 4 seconds, capture a chunk and send for analysis
  startChunkRecording(activeStream)
}

const startChunkRecording = (sourceStream) => {
  if (!sourceStream) return

  const startNewChunk = () => {
    if (!speechAnalysisActive.value || !sourceStream) return

    // Clone audio tracks for chunk recording
    const audioTracks = sourceStream.getAudioTracks()
    if (audioTracks.length === 0) return
    const clonedTracks = audioTracks.map((track) => track.clone())
    chunkRecordingStream = new MediaStream(clonedTracks)
    const preferredMime = pickSupportedAudioMime()

    try {
      chunkMediaRecorder = preferredMime
        ? new MediaRecorder(chunkRecordingStream, { mimeType: preferredMime })
        : new MediaRecorder(chunkRecordingStream)
    } catch {
      chunkMediaRecorder = new MediaRecorder(chunkRecordingStream)
    }
    chunkRecorderMimeType = normalizeAudioMime(chunkMediaRecorder.mimeType || preferredMime)

    const chunks = []
    let chunkPeakEnergy = 0
    const chunkEnergySampler = setInterval(() => {
      chunkPeakEnergy = Math.max(chunkPeakEnergy, Number(energyLevel.value) || 0)
    }, 120)
    chunkMediaRecorder.ondataavailable = (e) => { if (e.data.size > 0) chunks.push(e.data) }
    chunkMediaRecorder.onstop = () => {
      clearInterval(chunkEnergySampler)
      if (chunkRecordingStream) {
        chunkRecordingStream.getTracks().forEach((track) => track.stop())
        chunkRecordingStream = null
      }
      if (chunks.length === 0 || !interviewId.value) return
      const blob = new Blob(chunks, { type: chunkRecorderMimeType || 'audio/webm' })
      const reader = new FileReader()
      reader.onloadend = () => {
        const raw = String(reader.result || '')
        const parts = raw.split(',')
        if (parts.length < 2 || !parts[1]) return
        sendSpeechChunk(parts[1], speechChunkSeconds, chunkRecorderMimeType || '', chunkPeakEnergy)
      }
      reader.readAsDataURL(blob)
    }

    chunkMediaRecorder.start()

    // Use short chunks to make speech-rate feedback feel real-time.
    chunkInterval = setTimeout(() => {
      if (chunkMediaRecorder && chunkMediaRecorder.state === 'recording') {
        chunkMediaRecorder.stop()
      }
      // Start next chunk
      if (speechAnalysisActive.value) startNewChunk()
    }, speechChunkSeconds * 1000)
  }

  startNewChunk()
}

const sendSpeechChunk = async (audioBase64, duration, audioMime = '', chunkEnergy = 0) => {
  if (!interviewId.value) return
  try {
    const res = await apiAnalyzeSpeechChunk(interviewId.value, {
      audio_data: audioBase64,
      audio_mime: audioMime || undefined,
      duration: duration,
      energy_level: chunkEnergy
    })
    if (res.metrics) {
      const m = res.metrics
      const transcribed = String(m.transcribed_text || '').trim()
      const charCount = Number(m.char_count) || 0
      const rawRate = Number(m.speech_rate) || 0
      const audioDetected = (typeof m.audio_detected === 'boolean')
        ? m.audio_detected
        : ((Number(chunkEnergy) || 0) >= 0.02)
      let boundedRate = Math.max(0, Math.min(rawRate, 280))

      // Empty / near-empty chunks should not push the gauge to high speed.
      if (!audioDetected || !transcribed || charCount <= 1) {
        boundedRate = 0
      }

      const alpha = (audioDetected && transcribed) ? 0.35 : 0.2
      if (!speechRateSmoother.value || !Number.isFinite(speechRateSmoother.value)) {
        speechRateSmoother.value = boundedRate
      } else {
        speechRateSmoother.value = (speechRateSmoother.value * (1 - alpha)) + (boundedRate * alpha)
      }

      speechMetrics.value.speechRate = Math.round(speechRateSmoother.value * 10) / 10
      speechMetrics.value.speechRateLevel = classifySpeechRateLevelClient(speechMetrics.value.speechRate)
      speechMetrics.value.fillerWordCount = m.filler_word_count
      speechMetrics.value.fluencyAlert = m.fluency_alert
      speechMetrics.value.totalFillerWords += m.filler_word_count
      if (audioDetected && transcribed) {
        const merged = mergeChunkTranscript(speechMetrics.value.transcribedText, transcribed)
        speechMetrics.value.transcribedText = merged
        chunkTranscriptHistory.value.push(transcribed)
        if (chunkTranscriptHistory.value.length > 120) {
          chunkTranscriptHistory.value.shift()
        }
      }
    }
  } catch (err) {
    console.warn('Speech analysis chunk failed:', err)
  }
}

const stopSpeechAnalysis = () => {
  speechAnalysisActive.value = false
  if (chunkInterval) { clearTimeout(chunkInterval); chunkInterval = null }
  if (chunkMediaRecorder && chunkMediaRecorder.state === 'recording') {
    chunkMediaRecorder.stop()
  }
  if (chunkRecordingStream) {
    chunkRecordingStream.getTracks().forEach((track) => track.stop())
    chunkRecordingStream = null
  }
  if (energyAnimFrame) { cancelAnimationFrame(energyAnimFrame); energyAnimFrame = null }
  if (audioContext) {
    audioContext.close().catch(() => {})
    audioContext = null
  }
  energyLevel.value = 0
  analyserNode = null
  chunkMediaRecorder = null
  analysisSourceStream = null
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
  if (answerRecorderStream) {
    answerRecorderStream.getAudioTracks().forEach(track => { track.enabled = isMicOn.value })
  }
  if (analysisSourceStream) {
    analysisSourceStream.getAudioTracks().forEach(track => { track.enabled = isMicOn.value })
  }
}

const startCamera = async () => {
  try {
    stream.value = await navigator.mediaDevices.getUserMedia({ video: true, audio: true })
    stream.value.getAudioTracks().forEach(track => { track.enabled = isMicOn.value })
    isCameraOn.value = true
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

const stopInterviewRecordingStream = () => {
  if (!interviewRecordingStream) return
  if (stream.value && interviewRecordingStream === stream.value) return
  interviewRecordingStream.getTracks().forEach(track => track.stop())
  interviewRecordingStream = null
}

const startInterviewRecording = async () => {
  if (!interviewId.value || interviewMediaRecorder) return

  let targetStream = null
  if (settings.value.presentationMode === 'video_avatar' && stream.value) {
    targetStream = stream.value
  } else {
    try {
      // Fallback to audio-only recording so replay can still be generated.
      targetStream = await navigator.mediaDevices.getUserMedia({ audio: true, video: false })
      targetStream.getAudioTracks().forEach(track => { track.enabled = isMicOn.value })
    } catch (err) {
      console.warn('无法获取回放录制流:', err)
      recordingStatus.value = 'failed'
      return
    }
  }

  if (!targetStream) {
    recordingStatus.value = 'failed'
    return
  }

  interviewRecordingStream = targetStream

  try {
    interviewRecordedChunks = []
    recordingStatus.value = 'recording'
    interviewMediaRecorder = new MediaRecorder(targetStream, { mimeType: 'video/webm;codecs=vp8,opus' })
  } catch (_) {
    try {
      interviewMediaRecorder = new MediaRecorder(targetStream)
      recordingStatus.value = 'recording'
      interviewRecordedChunks = []
    } catch (err) {
      console.warn('无法创建视频录制器:', err)
      recordingStatus.value = 'failed'
      stopInterviewRecordingStream()
      return
    }
  }

  interviewMediaRecorder.ondataavailable = (e) => {
    if (e.data && e.data.size > 0) interviewRecordedChunks.push(e.data)
  }
  interviewMediaRecorder.onerror = (err) => {
    console.warn('回放录制异常:', err)
    recordingStatus.value = 'failed'
  }
  interviewMediaRecorder.start(1000)
}

const stopAndUploadInterviewRecording = async () => {
  if (!interviewId.value) return false
  if (!interviewMediaRecorder) {
    recordingStatus.value = recordingStatus.value === 'uploaded' ? 'uploaded' : 'failed'
    return false
  }

  if (interviewMediaRecorder.state === 'recording') {
    await new Promise((resolve) => {
      interviewMediaRecorder.onstop = resolve
      try {
        interviewMediaRecorder.stop()
      } catch (_) {
        resolve()
      }
    })
  }

  if (!interviewRecordedChunks.length) {
    recordingStatus.value = 'failed'
    interviewMediaRecorder = null
    stopInterviewRecordingStream()
    return false
  }

  const blob = new Blob(interviewRecordedChunks, { type: 'video/webm' })
  const formData = new FormData()
  formData.append('recording', blob, `interview_${interviewId.value}.webm`)

  try {
    const res = await apiUploadInterviewRecording(interviewId.value, formData)
    recordingUrl.value = res.recording_url || ''
    recordingStatus.value = 'uploaded'
    return true
  } catch (err) {
    console.warn('视频上传失败:', err)
    recordingStatus.value = 'failed'
    return false
  } finally {
    interviewMediaRecorder = null
    interviewRecordedChunks = []
    stopInterviewRecordingStream()
  }
}

// Interview Logic
const startInterview = async () => {
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
    pendingNextQuestion.value = null
    currentQuestion.value = null
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
    currentQuestion.value = questions.value[0] || null
    algorithmBriefText.value = '请用算法思维，在满足复杂度约束的情况下，实现如下算法题目。'
    algorithmProgress.value = { current: 1, total: 0, finished: 0, passed: 0, skipped: 0, failed: 0 }
    
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

    messages.value = [
      {
        role: 'ai',
        content: scenarioGreeting,
        type: isBlindBox ? 'scenario' : undefined
      }
    ]
    
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

const onAlgorithmBriefUpdated = (text) => {
  algorithmBriefText.value = String(text || '').trim() || '请用算法思维，在满足复杂度约束的情况下，实现如下算法题目。'
}

const onAlgorithmProgressUpdated = (progress) => {
  algorithmProgress.value = {
    current: progress?.current || 1,
    total: progress?.total || 0,
    finished: progress?.finished || 0,
    passed: progress?.passed || 0,
    skipped: progress?.skipped || 0,
    failed: progress?.failed || 0
  }
}

const onAlgorithmFinished = ({ total = 0, passed = 0, skipped = 0 }) => {
  messages.value.push({
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
  messages.value.push({
    role: 'ai',
    content: text,
    type: 'question'
  })
  if (settings.value.interviewMode !== 'human') {
    speakAIText(text)
  }
  preloadShadowHintPack()
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

const buildFeedbackPlainText = (sections) => {
  const lines = []
  const evaluation = (sections?.evaluation || '').trim()
  if (evaluation) {
    lines.push(`评价：${evaluation}`)
  }

  const suggestions = Array.isArray(sections?.suggestions) ? sections.suggestions.filter(Boolean) : []
  if (suggestions.length > 0) {
    lines.push('建议：')
    suggestions.forEach((item) => lines.push(`- ${item}`))
  }

  const followUp = (sections?.followUp || '').trim()
  if (followUp) {
    lines.push(`追问方向：${followUp}`)
  }

  return lines.join('\n').trim() || '回答已提交，建议补充更具体的技术细节。'
}

const getHistoryMessageContent = (msg) => {
  if (!msg) return ''

  if (msg.type === 'feedback') {
    const sections = {
      evaluation: msg.feedbackEvaluation,
      suggestions: msg.feedbackSuggestions,
      followUp: msg.feedbackFollowUp
    }
    const hasStructured = sections.evaluation || (sections.suggestions && sections.suggestions.length) || sections.followUp
    if (hasStructured) {
      return buildFeedbackPlainText(sections)
    }
    const fallback = splitFeedbackSections(formatFeedback(msg.content || ''))
    return buildFeedbackPlainText(fallback)
  }

  return typeof msg.content === 'string' ? msg.content : String(msg.content || '')
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
  if (/network\s*error|err_network|econnreset|wsarecv|forcibly\s+closed/i.test(text)) {
    return '网络连接中断（语音上传/转写链路异常）。请重试；若使用 ngrok，请确认隧道、前端 5173 与后端 8080 均在线'
  }
  if (/field\s+validation.*answer.*required/i.test(text) || /key:\s*'answer'/i.test(text)) {
    return '您似乎没有做出任何回答'
  }
  if (/audio\s+too\s+large|413/i.test(text)) {
    return '语音文件过大，请缩短录音后重试'
  }
  if (/status:\s*401|invalid\s+api\s*key|unauthorized|authentication/i.test(text)) {
    return '语音服务鉴权失败，请检查 ASR 的 API Key 是否有效'
  }
  if (/status:\s*429|quota|rate\s*limit|too\s+many\s+requests/i.test(text)) {
    return '语音服务额度或频率受限，请稍后重试'
  }
  if (/instruction\s+text|prompt\s+echo|possible\s+model\/provider\s+mismatch/i.test(text)) {
    return '语音转写服务返回了提示词回显，当前模型可能不兼容音频转写'
  }
  if (/model|unsupported\s+asr\s+provider|not\s+found/i.test(text) && /transcrib|audio/i.test(text)) {
    return '当前语音模型不可用，请检查 asr.model 和服务商兼容性'
  }
  if (/failed\s+to\s+transcribe\s+audio/i.test(text) || /empty\s+transcription\s+result/i.test(text)) {
    return '未识别到有效语音，请靠近麦克风并清晰作答后重试'
  }
  return text
}

const submitCurrentAnswer = async (answerText = '', audioData = '', audioMime = '') => {
  const currentQ = currentQuestion.value
  if (!currentQ || !currentQ.questionId) {
    throw new Error('当前题目ID无效，请重新开始面试')
  }

  const payload = {
    question_id: currentQ.questionId,
    question_title: currentQ.title || '',
    question_content: currentQ.content || '',
    answer: answerText,
    audio_data: audioData
  }
  if (audioMime) {
    payload.audio_mime = audioMime
  }
  const res = await apiSubmitAnswer(interviewId.value, payload)

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

  if (result.next_question) {
    pendingNextQuestion.value = {
      mapId: null,
      questionId: result.next_question.id || currentQ.questionId,
      title: result.next_question.title || '',
      content: result.next_question.content || '',
      expectedAnswer: result.next_question.expected_answer || '',
      source: result.next_question.source || 'standard'
    }
    pendingEnd.value = false
  } else {
    pendingNextQuestion.value = null
    pendingEnd.value = !!result.interview_completed
  }

  return result
}

const submitAudioAnswer = async (audioData, audioMime = '') => {
  if (!audioData) return
  if (isProcessing.value) return

  const userMsg = {
    role: 'user',
    content: '【语音回答转写中...】',
    rawTranscript: ''
  }
  const userMsgIndex = messages.value.length
  messages.value.push({
    ...userMsg
  })

  isProcessing.value = true
  processingHint.value = '面试官正在转写并评估你的语音回答...'
  answerVoiceStatus.value = 'submitting'
  answerVoiceError.value = ''

  try {
    const result = await submitCurrentAnswer('', audioData, audioMime)
    const transcript = String(result?.answer || '').trim()
    const plainText = transcript || '（未识别到有效语音文本）'
    const rendered = `【语音回答】\n${plainText}`
    userMsg.content = rendered
    userMsg.rawTranscript = plainText
    if (messages.value[userMsgIndex]) {
      messages.value[userMsgIndex] = { ...userMsg }
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
    isProcessing.value = false
    processingHint.value = ''
    scrollToBottom()
  }
}

const startAnswerRecording = async () => {
  if (isProcessing.value || !interviewId.value) return
  if (!canAnswerCurrentQuestion.value) {
    answerVoiceError.value = '请等待题目描述完成后再开始语音回答'
    answerVoiceStatus.value = 'error'
    return
  }
  if (!isMicOn.value) {
    answerVoiceError.value = '麦克风已关闭，请先开启麦克风'
    answerVoiceStatus.value = 'error'
    return
  }

  answerVoiceStatus.value = 'requesting'
  answerVoiceError.value = ''
  answerVoiceSeconds.value = 0
  quietSeconds.value = 0
  answerAudioChunks = []
  answerRecorderMimeType = ''
  answerRecordingPeakEnergy.value = 0
  speechRateSmoother.value = 0
  speechMetrics.value.transcribedText = ''
  chunkTranscriptHistory.value = []

  try {
    answerRecorderStream = await navigator.mediaDevices.getUserMedia({ audio: true })
    const preferredMime = pickSupportedAudioMime()

    try {
      answerMediaRecorder = preferredMime
        ? new MediaRecorder(answerRecorderStream, { mimeType: preferredMime })
        : new MediaRecorder(answerRecorderStream)
    } catch (_) {
      answerMediaRecorder = new MediaRecorder(answerRecorderStream)
    }
    answerRecorderMimeType = normalizeAudioMime(answerMediaRecorder.mimeType || preferredMime)

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
      if (isVideoInterviewMode.value && answerRecordingPeakEnergy.value < 0.06) {
        answerVoiceError.value = '未检测到有效语音输入，请检查麦克风并靠近后重试'
        answerVoiceStatus.value = 'error'
        return
      }

      answerVoiceStatus.value = 'transcribing'
      const audioBlob = new Blob(answerAudioChunks, { type: answerRecorderMimeType || 'audio/webm' })
      const reader = new FileReader()
      reader.onloadend = async () => {
        const raw = String(reader.result || '')
        const parts = raw.split(',')
        if (parts.length < 2 || !parts[1]) {
          answerVoiceError.value = '音频编码失败，请重试'
          answerVoiceStatus.value = 'error'
          return
        }
        await submitAudioAnswer(parts[1], answerRecorderMimeType || '')
      }
      reader.readAsDataURL(audioBlob)
    }

    answerMediaRecorder.start()
    if (isVideoInterviewMode.value) {
      startSpeechAnalysis(answerRecorderStream)
    }
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
  stopSpeechAnalysis()
  quietSeconds.value = 0
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
  if (isVideoInterviewMode.value) return
  if (!userInput.value.trim()) return
  
  const answer = userInput.value
  userInput.value = ''
  
  // 1. Add User Message
  messages.value.push({
    role: 'user',
    content: answer
  })
  
  isProcessing.value = true
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
    isProcessing.value = false
    processingHint.value = ''
    scrollToBottom()
  }
}

const advanceToNextQuestion = () => {
  if (answerVoiceStatus.value === 'recording') {
    stopAnswerRecording()
    return
  }
  if (pendingEnd.value) {
    stopQuestionTimer()
    stopAISpeech()
    messages.value.push({
      role: 'ai',
      content: "面试结束！辛苦了。您可以点击下方按钮查看详细报告。",
      type: 'system'
    })
    if (settings.value.interviewMode === 'random') {
      revealStyle()
    }
    pendingEnd.value = false
    pendingNextQuestion.value = null
    completeInterview()
    scrollToBottom()
    return
  }

  if (pendingNextQuestion.value) {
    currentQuestion.value = pendingNextQuestion.value
    currentQuestionIndex.value += 1
    thinkingStreakSeconds.value = 0
    pushAIQuestion(currentQuestion.value)
    if (blindBoxScenario.value?.time_limit) {
      startQuestionTimer(blindBoxScenario.value.time_limit)
    }
  }
  pendingNextQuestion.value = null
  pendingEnd.value = false
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
      messages.value.push({
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
    answerVoiceStatus.value = 'idle'
    answerVoiceError.value = ''
    answerVoiceSeconds.value = 0
    await stopAndUploadInterviewRecording()
    stopCamera()
    stopQuestionTimer()
    stopAISpeech()
    phase.value = 'setup'
    currentQuestionIndex.value = 0
    currentQuestion.value = null
    messages.value = []
    pendingNextQuestion.value = null
    pendingEnd.value = false
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

onMounted(() => {
  ensureModelViewerScript()
  if (settings.value.presentationMode === 'video_avatar') {
    startCamera()
  }
  loadInterviewConfig()
  startThinkingWatch()
})

onUnmounted(() => {
  document.body.style.overflow = ''
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
  stopInterviewRecordingStream()
  stopAISpeech()
  stopCamera()
  stopSpeechAnalysis()
  stopThinkingWatch()
  hideShadowBubble()
})
</script>

<template>
  <div class="min-h-[calc(100vh-8rem)] flex flex-col">
    <!-- Setup Phase -->
    <div v-if="phase === 'setup'" class="flex-1 flex flex-col items-center justify-center max-w-4xl mx-auto w-full space-y-8 animate-in fade-in duration-500">
      <header class="text-center">
        <h1 class="text-3xl font-bold text-zinc-900">AI 模拟面试</h1>
        <p class="text-zinc-500 mt-2">配置您的面试环境与偏好，开启真实对话体验</p>
      </header>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-8 w-full items-stretch">
        <!-- Preview Area -->
        <InterviewSetupPreview
          :presentation-mode="settings.presentationMode"
          :is-camera-on="isCameraOn"
          :is-mic-on="isMicOn"
          :stream="stream"
          @toggle-mic="toggleMic"
          @toggle-camera="toggleCamera"
        />

        <!-- Settings Area -->
        <InterviewSetupForm
          v-model:settings="settings"
          v-model:shadow-coach-enabled="shadowCoachEnabled"
          :is-processing="isProcessing"
          :active-invitation-id="activeInvitationId"
          :invite-candidates="inviteCandidates"
          :invite-candidates-loading="inviteCandidatesLoading"
          :active-invitation="activeInvitation"
          :blind-box-revealing="blindBoxRevealing"
          :blind-box-revealed="blindBoxRevealed"
          :blind-box-scenario="blindBoxScenario"
          :pressure-colors="pressureColors"
          :pressure-level="pressureLevel"
          :pressure-labels="pressureLabels"
          :normalize-candidate-role="normalizeCandidateRole"
          @change-presentation-mode="setPresentationMode"
          @change-interview-mode="setInterviewMode"
          @load-invite-candidates="loadInviteCandidates"
          @select-invite-candidate="selectInviteCandidate"
          @open-bookings="onOpenBookingsPanel"
          @draw-blind-box="drawBlindBox"
          @redraw-blind-box="reDrawBlindBox"
          @start-interview="startInterview"
        />
      </div>
    </div>

    <!-- Interview Phase (New Layout) -->
    <div v-else-if="phase === 'interview'" class="min-h-[calc(100vh-8rem)] flex flex-col lg:flex-row gap-6 p-6 bg-gradient-to-br from-slate-50 via-white to-cyan-50 overflow-y-auto">
      
      <!-- Left Main Column (Video + Input) -->
      <div class="flex-1 flex flex-col gap-6 min-w-0">
        <!-- Video Section (Top) -->
        <InterviewVideoStage
          v-if="settings.presentationMode === 'video_avatar' && settings.interviewMode !== 'human'"
          :is-camera-on="isCameraOn"
          :is-mic-on="isMicOn"
          :is-avatar-speaking="isAvatarSpeaking"
          :model-viewer-ready="modelViewerReady"
          :recording-status="recordingStatus"
          :blind-box-scenario="blindBoxScenario"
          :question-timer="questionTimer"
          :pressure-level="pressureLevel"
          :interview-style="settings.style"
          :shadow-coach-hint-pending="shadowCoachHintPending"
          :shadow-coach-bubble-visible="shadowCoachBubbleVisible"
          :shadow-coach-bubble-text="shadowCoachBubbleText"
          :stream="stream"
          @toggle-mic="toggleMic"
          @toggle-camera="toggleCamera"
        />
        <div v-else class="flex-1 rounded-3xl p-6 border border-zinc-200 bg-gradient-to-br from-emerald-50 via-white to-cyan-50 shadow-xl flex flex-col justify-between">
          <div>
            <p class="text-sm font-bold text-emerald-700">文字 + 语音模式进行中</p>
            <p class="text-xs text-zinc-500 mt-2">当前聚焦回答内容与逻辑质量，系统会持续展示实时语音转写与表达指标。</p>
          </div>
          <div class="grid grid-cols-2 gap-3 text-xs">
            <div class="rounded-xl bg-white border border-zinc-200 p-3">
              <p class="text-zinc-400">当前岗位</p>
              <p class="font-bold text-zinc-800 mt-1">{{ settings.position }}</p>
            </div>
            <div class="rounded-xl bg-white border border-zinc-200 p-3">
              <p class="text-zinc-400">面试进度</p>
              <p class="font-bold text-zinc-800 mt-1">第 {{ currentQuestionIndex + 1 }} 题</p>
            </div>
          </div>
        </div>

        <!-- Transcript / Input Section (Bottom) -->
        <div v-if="!isAlgorithmStyle" class="bg-white rounded-3xl p-6 shadow-xl shadow-zinc-200/50 border border-white flex flex-col relative transition-all duration-300 focus-within:ring-2 focus-within:ring-indigo-500/20 focus-within:shadow-indigo-500/10 group lg:resizable-panel lg:flex-none"
          :class="isVideoInterviewMode ? 'min-h-[320px] lg:min-h-[380px] lg:h-auto' : 'min-h-[260px] lg:min-h-[320px] lg:h-auto'">
           <div class="flex flex-col sm:flex-row sm:justify-between sm:items-center items-start gap-2 mb-4 shrink-0">
             <h3 class="font-bold text-zinc-900 flex items-center gap-2 group-focus-within:text-indigo-600 transition-colors">
               <div class="w-1.5 h-4 bg-zinc-300 rounded-full group-focus-within:bg-indigo-600 transition-colors"></div>
               {{ isVideoInterviewMode ? '语音回答控制台' : '实时回答转写' }}
               <span v-if="userInput.length > 0" class="text-xs font-normal text-emerald-600 flex items-center gap-1 bg-emerald-50 px-2 py-0.5 rounded-full animate-in fade-in zoom-in duration-300">
                 <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
                 正在输入...
               </span>
             </h3>
             <div class="flex items-center gap-2 w-full sm:w-auto justify-end">
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

           <template v-if="isVideoInterviewMode">
             <div class="flex-1 min-h-0 rounded-2xl border border-zinc-100 bg-zinc-50/70 p-4 flex flex-col gap-3 overflow-auto custom-scrollbar">
               <p class="text-sm text-zinc-600 leading-relaxed shrink-0">
                 视频模式已禁用文字提交。请点击下方按钮开始作答，回答完成后点击“结束语音并开始分析”。
               </p>
               <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 shrink-0">
                 <div class="rounded-xl bg-white border border-zinc-200 p-3">
                   <p class="text-[11px] text-zinc-400">语速</p>
                   <p class="text-sm font-bold text-zinc-800 mt-1">{{ speechMetrics.speechRate || 0 }} 字/分</p>
                 </div>
                 <div class="rounded-xl bg-white border border-zinc-200 p-3">
                   <p class="text-[11px] text-zinc-400">本轮填充词</p>
                   <p class="text-sm font-bold text-zinc-800 mt-1">{{ speechMetrics.fillerWordCount || 0 }}</p>
                 </div>
               </div>
                <div class="rounded-xl bg-white border border-zinc-200 p-3 min-h-0 flex-1 overflow-hidden">
                  <p class="text-[11px] text-zinc-400">最近语音转写</p>
                  <p class="text-sm text-zinc-700 mt-1 whitespace-pre-wrap break-words max-h-full overflow-y-auto custom-scrollbar">{{ speechMetrics.transcribedText || latestUserTranscript || '暂无转写内容' }}</p>
                </div>
             </div>
           </template>
           <template v-else>
             <textarea 
                v-model="userInput" 
                @keydown.ctrl.enter="sendMessage"
                placeholder="在此处输入您的回答..."
               class="flex-1 min-h-[220px] max-h-[55vh] w-full resize-y border-none focus:ring-0 p-4 text-lg text-zinc-700 placeholder:text-zinc-300 bg-zinc-50/50 rounded-xl leading-relaxed transition-all focus:bg-white focus:shadow-inner overflow-y-auto custom-scrollbar"
             ></textarea>

             <div class="absolute bottom-8 right-8 text-[10px] font-medium text-zinc-300 pointer-events-none bg-white/80 backdrop-blur px-2 py-1 rounded-md border border-zinc-100">
               Ctrl + Enter 发送
             </div>
           </template>
        </div>

        <AlgorithmInterviewPanel
          v-else-if="interviewId"
          :interview-id="interviewId"
          :difficulty="settings.difficulty"
          @brief-updated="onAlgorithmBriefUpdated"
          @progress-updated="onAlgorithmProgressUpdated"
          @finished="onAlgorithmFinished"
        />
      </div>

      <!-- Right Sidebar -->
      <div class="w-full lg:w-[400px] flex flex-col gap-4 shrink-0 lg:h-full lg:min-h-0 lg:overflow-y-auto custom-scrollbar">
        <InterviewFeedbackSidebar
          :settings="settings"
          :active-invitation="activeInvitation"
          :is-algorithm-style="isAlgorithmStyle"
          :current-question-index="currentQuestionIndex"
          :current-question="currentQuestion"
          :latest-ai-message="latestAIMessage"
          :is-processing="isProcessing"
          :processing-hint="processingHint"
          :algorithm-brief-text="algorithmBriefText"
          :algorithm-progress="algorithmProgress"
          :shadow-coach-enabled="shadowCoachEnabled"
          :shadow-coach-hints="shadowCoachHints"
          :blind-box-scenario="blindBoxScenario"
          :pressure-colors="pressureColors"
          :pressure-level="pressureLevel"
          :pressure-labels="pressureLabels"
          :question-timer="questionTimer"
          :normalize-candidate-role="normalizeCandidateRole"
          @view-report="viewReport"
        />

        <!-- Real-time Speech Dashboard -->
        <div v-if="!isAlgorithmStyle" class="bg-white rounded-3xl p-4 border border-zinc-100 shadow-sm shrink-0 lg:resizable-panel lg:flex-none lg:h-[250px]">
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
          v-if="!isAlgorithmStyle"
          @click="toggleAnswerRecording"
          :disabled="!canAnswerCurrentQuestion || answerVoiceStatus === 'requesting' || answerVoiceStatus === 'transcribing' || answerVoiceStatus === 'submitting'"
          class="w-full py-3 rounded-2xl font-bold text-base transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2 border shrink-0"
          :class="answerVoiceStatus === 'recording'
            ? 'bg-rose-600 text-white border-rose-600 hover:bg-rose-700'
            : 'bg-white text-zinc-700 border-zinc-200 hover:bg-zinc-50'"
        >
          <Mic v-if="answerVoiceStatus !== 'recording'" class="h-4 w-4" />
          <MicOff v-else class="h-4 w-4" />
          <span v-if="answerVoiceStatus === 'recording'">结束语音并开始分析</span>
          <span v-else>{{ canAnswerCurrentQuestion ? (isVideoInterviewMode ? '开始语音回答' : '语音回答') : '等待题目描述完成' }}</span>
        </button>

        <!-- Action Button -->
        <button 
          @click="sendMessage"
          v-if="!isAlgorithmStyle && (!isVideoInterviewMode || latestAIMessage?.type === 'feedback')"
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
      <div v-if="showHistory" class="fixed inset-0 z-70 bg-black/20 backdrop-blur-sm flex justify-end" @click.self="showHistory = false">
        <div class="w-96 max-w-[92vw] bg-white h-dvh shadow-2xl animate-in slide-in-from-right duration-300 flex flex-col border-l border-zinc-100">
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
                {{ msg.role === 'ai' ? (settings.interviewMode === 'human' ? '真人面试流程' : 'AI 面试官') : '你' }}
              </div>
              <div class="leading-relaxed whitespace-pre-wrap">{{ getHistoryMessageContent(msg) }}</div>
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
