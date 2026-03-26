import { computed, onUnmounted, ref } from 'vue'
import { drawBlindBoxScenario as apiDrawBlindBox } from '../api/interview'

export function useBlindBox(options = {}) {
  const onDrawError = typeof options.onDrawError === 'function'
    ? options.onDrawError
    : (err) => {
      console.error('Failed to draw blindbox:', err)
      alert('抽取场景失败：' + (err.response?.data?.error || err.message))
    }

  const blindBoxScenario = ref(null)
  const blindBoxRevealing = ref(false)
  const blindBoxRevealed = ref(false)
  const questionTimeLimit = ref(0)
  const questionTimer = ref(0)

  let questionTimerInterval = null

  const pressureLevel = computed(() => blindBoxScenario.value?.pressure || 'low')
  const isHighPressure = computed(() => ['high', 'extreme'].includes(pressureLevel.value))

  const pressureColors = {
    low: { bg: 'bg-emerald-50', border: 'border-emerald-200', text: 'text-emerald-700', badge: 'bg-emerald-100 text-emerald-700' },
    medium: { bg: 'bg-amber-50', border: 'border-amber-200', text: 'text-amber-700', badge: 'bg-amber-100 text-amber-700' },
    high: { bg: 'bg-rose-50', border: 'border-rose-300', text: 'text-rose-700', badge: 'bg-rose-100 text-rose-700' },
    extreme: { bg: 'bg-red-50', border: 'border-red-400', text: 'text-red-800', badge: 'bg-red-200 text-red-800' }
  }

  const pressureLabels = {
    low: '轻松',
    medium: '中等',
    high: '高压',
    extreme: '极限'
  }

  const stopQuestionTimer = () => {
    if (questionTimerInterval) {
      clearInterval(questionTimerInterval)
      questionTimerInterval = null
    }
    questionTimer.value = 0
  }

  const startQuestionTimer = (limitSec) => {
    stopQuestionTimer()
    if (!limitSec || limitSec <= 0) return

    questionTimeLimit.value = limitSec
    questionTimer.value = limitSec
    questionTimerInterval = setInterval(() => {
      questionTimer.value -= 1
      if (questionTimer.value <= 0) {
        stopQuestionTimer()
      }
    }, 1000)
  }

  const drawBlindBox = async () => {
    blindBoxRevealing.value = true
    blindBoxRevealed.value = false
    blindBoxScenario.value = null

    try {
      const res = await apiDrawBlindBox()
      // Simulate slot-machine reveal delay.
      await new Promise((resolve) => setTimeout(resolve, 1500))
      blindBoxScenario.value = res.scenario
      blindBoxRevealed.value = true
    } catch (err) {
      onDrawError(err)
    } finally {
      blindBoxRevealing.value = false
    }
  }

  const reDrawBlindBox = () => {
    blindBoxRevealed.value = false
    drawBlindBox()
  }

  onUnmounted(() => {
    stopQuestionTimer()
  })

  return {
    blindBoxScenario,
    blindBoxRevealing,
    blindBoxRevealed,
    questionTimeLimit,
    questionTimer,
    pressureLevel,
    isHighPressure,
    pressureColors,
    pressureLabels,
    drawBlindBox,
    reDrawBlindBox,
    startQuestionTimer,
    stopQuestionTimer
  }
}
