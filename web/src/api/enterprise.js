import request from '../utils/request'

// ===== 企业仪表盘 =====
export function getEnterpriseDashboard() {
  return request({ url: '/enterprise/dashboard', method: 'get' })
}

// ===== 人才库 =====
export function getTalentPool(params) {
  return request({ url: '/enterprise/talent-pool', method: 'get', params })
}

export function inviteTalent(candidateId) {
  return request({ url: `/enterprise/talent-pool/${candidateId}/invite`, method: 'post' })
}

export function saveTalent(candidateId) {
  return request({ url: `/enterprise/talent-pool/${candidateId}/save`, method: 'post' })
}

// ===== 岗位管理 =====
export function getJobs(params) {
  return request({ url: '/enterprise/jobs', method: 'get', params })
}

export function createJob(data) {
  return request({ url: '/enterprise/jobs', method: 'post', data })
}

export function updateJob(id, data) {
  return request({ url: `/enterprise/jobs/${id}`, method: 'put', data })
}

export function deleteJob(id) {
  return request({ url: `/enterprise/jobs/${id}`, method: 'delete' })
}

export function getAbilityAtlas(jobId) {
  return request({ url: `/enterprise/jobs/${jobId}/ability-atlas`, method: 'get' })
}

// ===== HR 面试官面板 =====
export function getInterviewSessions(params) {
  return request({ url: '/enterprise/interview-sessions', method: 'get', params })
}

export function createCustomScenario(data) {
  return request({ url: '/enterprise/scenarios', method: 'post', data })
}

export function getCustomScenarios() {
  return request({ url: '/enterprise/scenarios', method: 'get' })
}

// ===== 数据分析 =====
export function getRecruitmentAnalytics(params) {
  return request({ url: '/enterprise/analytics', method: 'get', params })
}

export function getRecruitmentFunnel() {
  return request({ url: '/enterprise/analytics/funnel', method: 'get' })
}

export function getCandidateQualityDistribution() {
  return request({ url: '/enterprise/analytics/quality', method: 'get' })
}

// ===== 能力标准共建 =====
export function getCapabilityStandards(params) {
  return request({ url: '/enterprise/standards', method: 'get', params })
}

export function createCapabilityStandard(data) {
  return request({ url: '/enterprise/standards', method: 'post', data })
}

export function updateCapabilityStandard(id, data) {
  return request({ url: `/enterprise/standards/${id}`, method: 'put', data })
}

// ===== 人才认证 =====
export function getCertifiedCandidates(params) {
  return request({ url: '/enterprise/certified', method: 'get', params })
}

// ===== 内推渠道 =====
export function getReferralChannels() {
  return request({ url: '/enterprise/referrals', method: 'get' })
}

export function createReferral(data) {
  return request({ url: '/enterprise/referrals', method: 'post', data })
}

// ===== 真人面试邀请 =====
export function getReceivedInterviewInvitations() {
  return request({ url: '/interview/invitations/received', method: 'get' })
}

export function respondInterviewInvitation(invitationId, action) {
  return request({
    url: `/interview/invitations/${invitationId}/respond`,
    method: 'post',
    data: { action }
  })
}
