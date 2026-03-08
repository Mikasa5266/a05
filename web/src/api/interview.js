import request from '../utils/request'

export function startInterview(data) {
  return request({
    url: '/interview/start',
    method: 'post',
    data
  })
}

export function getInterview(id) {
  return request({
    url: `/interview/${id}`,
    method: 'get'
  })
}

export function getInterviews(params) {
  return request({
    url: '/interview',
    method: 'get',
    params
  })
}

export function submitAnswer(id, data) {
  return request({
    url: `/interview/${id}/answer`,
    method: 'put',
    data
  })
}

export function endInterview(id) {
  return request({
    url: `/interview/${id}/end`,
    method: 'post'
  })
}

export function uploadInterviewRecording(id, formData) {
  return request({
    url: `/interview/${id}/recording`,
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

export function getQuestions(params) {
  return request({
    url: '/questions',
    method: 'get',
    params
  })
}

export function analyzeSpeechChunk(interviewId, data) {
  return request({
    url: `/interview/${interviewId}/speech-analyze`,
    method: 'post',
    data
  })
}

export function getShadowCoachHint(interviewId, data) {
  return request({
    url: `/interview/${interviewId}/shadow-hint`,
    method: 'post',
    data
  })
}

export function synthesizeInterviewSpeech(interviewId, data) {
  return request({
    url: `/interview/${interviewId}/tts`,
    method: 'post',
    data
  })
}

export function drawBlindBoxScenario(data = {}) {
  return request({
    url: '/interview/blindbox/draw',
    method: 'post',
    data
  })
}

export function getBlindBoxScenarios() {
  return request({
    url: '/interview/blindbox/scenarios',
    method: 'get'
  })
}

// ========== New Interview Features ==========

// Get interview configuration options (modes, styles, companies, difficulties)
export function getInterviewConfig() {
  return request({
    url: '/interview/config',
    method: 'get'
  })
}

// Get available human interviewers
export function getHumanInterviewers(params) {
  return request({
    url: '/interview/human-interviewers',
    method: 'get',
    params
  })
}

// Get a specific human interviewer
export function getHumanInterviewer(id) {
  return request({
    url: `/interview/human-interviewers/${id}`,
    method: 'get'
  })
}

// Book a human interview
export function bookHumanInterview(data) {
  return request({
    url: '/interview/booking',
    method: 'post',
    data
  })
}

// Get user's interview bookings
export function getUserBookings() {
  return request({
    url: '/interview/bookings',
    method: 'get'
  })
}

// Submit human interviewer feedback
export function submitHumanFeedback(interviewId, data) {
  return request({
    url: `/interview/${interviewId}/human-feedback`,
    method: 'post',
    data
  })
}

// Reveal the hidden style after a random-mode interview
export function revealRandomStyle(interviewId) {
  return request({
    url: `/interview/${interviewId}/reveal-style`,
    method: 'get'
  })
}
