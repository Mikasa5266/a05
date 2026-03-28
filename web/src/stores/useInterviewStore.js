import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useInterviewStore = defineStore('mockInterviewStore', () => {
  const interviewId = ref(null)
  const questions = ref([])
  const currentQuestion = ref(null)

  const shadowCoachEnabled = ref(true)
  const shadowCoachHints = ref([])
  const shadowCoachBubbleText = ref('')
  const shadowCoachBubbleVisible = ref(false)
  const shadowCoachHintPending = ref(false)
  const silenceStreakSeconds = ref(0)
  const thinkingStreakSeconds = ref(0)
  const quietSeconds = ref(0)
  const shadowHintPack = ref([])
  const shadowHintDelivered = ref([false, false, false])

  const emotionFeedback = ref({ sentiment: '正常', confidence: 0, heartRate: 72 })

  let shadowBubbleTimer = null
  let thinkingWatchTimer = null
  let hintDispatchInFlight = false

  const assignNumberRef = (target, nextValue) => {
    const normalized = Number(nextValue) || 0
    if (target.value !== normalized) {
      target.value = normalized
    }
  }

  const resetWatchCounters = () => {
    assignNumberRef(thinkingStreakSeconds, 0)
    assignNumberRef(silenceStreakSeconds, 0)
    assignNumberRef(quietSeconds, 0)
  }

  const setInterviewId = (id) => {
    interviewId.value = id
  }

  const setQuestions = (list) => {
    questions.value = Array.isArray(list) ? list : []
  }

  const setCurrentQuestion = (question) => {
    currentQuestion.value = question || null
  }

  const setQuietSeconds = (seconds) => {
    assignNumberRef(quietSeconds, seconds)
  }

  const hideShadowBubble = () => {
    shadowCoachBubbleVisible.value = false
    if (shadowBubbleTimer) {
      clearTimeout(shadowBubbleTimer)
      shadowBubbleTimer = null
    }
  }

  const showShadowBubble = (text, durationMs = 9000) => {
    shadowCoachBubbleText.value = String(text || '')
    shadowCoachBubbleVisible.value = true

    if (shadowBubbleTimer) {
      clearTimeout(shadowBubbleTimer)
    }
    shadowBubbleTimer = setTimeout(() => {
      shadowCoachBubbleVisible.value = false
      shadowBubbleTimer = null
    }, durationMs)
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
    resetWatchCounters()
    shadowHintPack.value = []
    shadowHintDelivered.value = [false, false, false]
  }

  const preloadShadowHintPack = async ({ apiGetShadowCoachHint, latestUserTranscript }) => {
    if (!shadowCoachEnabled.value || !interviewId.value || !currentQuestion.value) return
    if (shadowHintPack.value.length >= 3 || shadowCoachHintPending.value) return
    if (typeof apiGetShadowCoachHint !== 'function') return

    shadowCoachHintPending.value = true
    try {
      const questionText = currentQuestion.value.content || currentQuestion.value.title || ''
      const res = await apiGetShadowCoachHint(interviewId.value, {
        question: questionText,
        transcript: latestUserTranscript?.value || '',
        expected_answer: currentQuestion.value.expectedAnswer || '',
        silence_seconds: 0
      })
      shadowHintPack.value = normalizeHintPack(
        res?.hints || (res?.hint ? [res.hint] : []),
        questionText
      )
    } catch (err) {
      console.warn('preload shadow hint pack failed:', err)
      const questionText = currentQuestion.value.content || currentQuestion.value.title || ''
      shadowHintPack.value = normalizeHintPack([], questionText)
    } finally {
      shadowCoachHintPending.value = false
    }
  }

  const maybeDispatchTieredHint = async ({
    phase,
    isProcessing,
    latestAIMessage,
    apiGetShadowCoachHint,
    latestUserTranscript
  }) => {
    if (!shadowCoachEnabled.value || !interviewId.value || !currentQuestion.value) return
    if (phase?.value !== 'interview' || isProcessing?.value) return
    if (latestAIMessage?.value?.type !== 'question') return

    if (shadowHintPack.value.length < 3) {
      await preloadShadowHintPack({ apiGetShadowCoachHint, latestUserTranscript })
    }
    if (shadowHintPack.value.length < 3) return

    const milestones = [20, 40, 60]
    for (let i = 0; i < milestones.length; i += 1) {
      if (quietSeconds.value >= milestones[i] && !shadowHintDelivered.value[i]) {
        const hint = shadowHintPack.value[i]
        showShadowBubble(hint)
        shadowCoachHints.value.unshift({ text: hint, at: Date.now(), level: i + 1 })
        if (shadowCoachHints.value.length > 8) {
          shadowCoachHints.value.splice(8)
        }
        const delivered = [...shadowHintDelivered.value]
        delivered[i] = true
        shadowHintDelivered.value = delivered
        break
      }
    }
  }

  const startThinkingWatch = ({
    phase,
    isProcessing,
    latestAIMessage,
    canAnswerCurrentQuestion,
    answerVoiceStatus,
    energyLevel,
    onMaybeDispatch
  }) => {
    if (thinkingWatchTimer) return

    thinkingWatchTimer = setInterval(() => {
      if (!shadowCoachEnabled.value || phase?.value !== 'interview') {
        resetWatchCounters()
        return
      }

      if (isProcessing?.value || latestAIMessage?.value?.type !== 'question' || !canAnswerCurrentQuestion?.value) {
        resetWatchCounters()
        return
      }

      let nextThinking = thinkingStreakSeconds.value
      let nextSilence = silenceStreakSeconds.value
      let nextQuiet = quietSeconds.value

      if (answerVoiceStatus?.value === 'recording') {
        nextThinking = 0
        if ((energyLevel?.value || 0) <= 0.035) {
          nextSilence += 1
          nextQuiet += 1
        } else {
          nextSilence = 0
          nextQuiet = 0
        }
      } else {
        nextSilence = 0
        nextThinking += 1
        nextQuiet += 1
      }

      const allHintsDelivered = shadowHintDelivered.value.every(Boolean)
      if (!allHintsDelivered && nextQuiet >= 65) {
        nextQuiet = 60
      }

      const quietChanged = nextQuiet !== quietSeconds.value

      assignNumberRef(thinkingStreakSeconds, nextThinking)
      assignNumberRef(silenceStreakSeconds, nextSilence)
      assignNumberRef(quietSeconds, nextQuiet)

      if (typeof onMaybeDispatch === 'function' && quietChanged && nextQuiet >= 20 && !hintDispatchInFlight) {
        hintDispatchInFlight = true
        Promise.resolve(onMaybeDispatch())
          .catch((err) => {
            console.warn('dispatch tiered hint failed:', err)
          })
          .finally(() => {
            hintDispatchInFlight = false
          })
      }
    }, 1000)
  }

  const stopThinkingWatch = () => {
    if (!thinkingWatchTimer) return
    clearInterval(thinkingWatchTimer)
    thinkingWatchTimer = null
    hintDispatchInFlight = false
    resetWatchCounters()
  }

  const cleanupStoreTimers = () => {
    hideShadowBubble()
    stopThinkingWatch()
  }

  return {
    interviewId,
    questions,
    currentQuestion,
    shadowCoachEnabled,
    shadowCoachHints,
    shadowCoachBubbleText,
    shadowCoachBubbleVisible,
    shadowCoachHintPending,
    silenceStreakSeconds,
    thinkingStreakSeconds,
    quietSeconds,
    shadowHintPack,
    shadowHintDelivered,
    emotionFeedback,
    setInterviewId,
    setQuestions,
    setCurrentQuestion,
    setQuietSeconds,
    hideShadowBubble,
    showShadowBubble,
    buildLocalHintDefaults,
    normalizeHintPack,
    resetShadowHintProgress,
    preloadShadowHintPack,
    maybeDispatchTieredHint,
    startThinkingWatch,
    stopThinkingWatch,
    cleanupStoreTimers
  }
})
