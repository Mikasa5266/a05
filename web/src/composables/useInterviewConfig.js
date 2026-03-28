import { ref, watch } from 'vue'

const SETTINGS_STORAGE_KEY = 'mock_interview_settings_v1'

const buildDefaultSettings = (routeQuery = {}) => ({
  position: routeQuery.position || 'Java后端工程师',
  difficulty: 'campus_intern',
  mode: routeQuery.mode || 'technical',
  style: 'gentle',
  company: '',
  interviewMode: 'ai',
  presentationMode: routeQuery.presentationMode || 'video_avatar'
})

const normalizeSettings = (raw, defaults) => {
  const candidate = (raw && typeof raw === 'object') ? raw : {}
  const normalized = {
    ...defaults,
    ...candidate
  }

  const allowedInterviewModes = new Set(['ai', 'human', 'random'])
  const allowedPresentationModes = new Set(['text_voice', 'video_avatar'])

  if (!allowedInterviewModes.has(normalized.interviewMode)) {
    normalized.interviewMode = defaults.interviewMode
  }
  if (!allowedPresentationModes.has(normalized.presentationMode)) {
    normalized.presentationMode = defaults.presentationMode
  }

  return normalized
}

const loadSettingsFromStorage = (defaults) => {
  try {
    const raw = window.localStorage.getItem(SETTINGS_STORAGE_KEY)
    if (!raw) return defaults
    const parsed = JSON.parse(raw)
    return normalizeSettings(parsed, defaults)
  } catch {
    return defaults
  }
}

const applyRouteQueryOverrides = (currentSettings, routeQuery, defaults) => {
  const next = {
    ...(currentSettings || {})
  }

  if (routeQuery.position) next.position = String(routeQuery.position)
  if (routeQuery.mode) next.mode = String(routeQuery.mode)
  if (routeQuery.style) next.style = String(routeQuery.style)
  if (routeQuery.company) next.company = String(routeQuery.company)
  if (routeQuery.interviewMode) next.interviewMode = String(routeQuery.interviewMode)
  if (routeQuery.presentationMode) next.presentationMode = String(routeQuery.presentationMode)

  return normalizeSettings(next, defaults)
}

export function useInterviewConfig(options = {}) {
  const routeQuery = options.routeQuery || {}
  const defaults = buildDefaultSettings(routeQuery)

  const phase = ref('setup')
  const settings = ref(applyRouteQueryOverrides(loadSettingsFromStorage(defaults), routeQuery, defaults))

  watch(settings, (next) => {
    try {
      window.localStorage.setItem(SETTINGS_STORAGE_KEY, JSON.stringify(next))
    } catch {
      // Ignore storage errors in private mode or restricted environments.
    }
  }, { deep: true })

  const startInterview = () => {
    phase.value = 'interview'
  }

  const endInterview = () => {
    phase.value = 'setup'
  }

  const resetSettings = (overrides = {}) => {
    settings.value = normalizeSettings({ ...defaults, ...overrides }, defaults)
  }

  return {
    phase,
    settings,
    startInterview,
    endInterview,
    resetSettings
  }
}
