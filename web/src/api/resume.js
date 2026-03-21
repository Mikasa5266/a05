import request from '../utils/request'

export function parseResume(formData) {
  return request({
    url: '/resume/parse',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

export function analyzeResumeAuthenticity(payload) {
  return request({
    url: '/resume/authenticity',
    method: 'post',
    data: payload
  })
}

export function getResumeOptimizationSuggestions(payload) {
  return request({
    url: '/resume/optimize',
    method: 'post',
    data: payload
  })
}

export function generateResumeTemplate(payload) {
  return request({
    url: '/resume/template',
    method: 'post',
    data: payload
  })
}

export function aiChatFallback(payload) {
  return request({
    url: '/ai/chat',
    method: 'post',
    data: payload
  })
}
