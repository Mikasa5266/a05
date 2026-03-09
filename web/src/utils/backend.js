const DEFAULT_API_BASE_URL = '/api/v1'

export const API_BASE_URL = import.meta.env.VITE_API_URL || DEFAULT_API_BASE_URL

export const BACKEND_ORIGIN = (() => {
  try {
    const base = typeof window !== 'undefined' ? window.location.origin : undefined
    const url = new URL(API_BASE_URL, base)
    return `${url.protocol}//${url.host}`
  } catch {
    return window.location.origin
  }
})()

export const getBackendAssetUrl = (path = '') => {
  const value = String(path || '').trim()
  if (!value) return ''
  if (/^https?:\/\//i.test(value)) return value
  return `${BACKEND_ORIGIN}${value.startsWith('/') ? '' : '/'}${value}`
}
