import request from '../utils/request'

// ===== 帖子/经验分享 =====
export function getPosts(params) {
  return request({ url: '/community/posts', method: 'get', params })
}

export function getPost(id) {
  return request({ url: `/community/posts/${id}`, method: 'get' })
}

export function createPost(data) {
  return request({ url: '/community/posts', method: 'post', data })
}

export function likePost(id) {
  return request({ url: `/community/posts/${id}/like`, method: 'post' })
}

export function commentOnPost(id, data) {
  return request({ url: `/community/posts/${id}/comments`, method: 'post', data })
}

export function getPostComments(id, params) {
  return request({ url: `/community/posts/${id}/comments`, method: 'get', params })
}

// ===== 1v1 预约 =====
export function bookMentor(mentorId, data) {
  return request({ url: `/community/mentors/${mentorId}/book`, method: 'post', data })
}

export function getMentors(params) {
  return request({ url: '/community/mentors', method: 'get', params })
}

export function getBookings(params) {
  return request({ url: '/community/bookings', method: 'get', params })
}

// ===== AI 知识库 =====
export function queryKnowledgeBase(data) {
  return request({ url: '/community/knowledge/query', method: 'post', data })
}

// ===== 热门 =====
export function getTopAlumni() {
  return request({ url: '/community/top-alumni', method: 'get' })
}

export function getHotCompanies() {
  return request({ url: '/community/hot-companies', method: 'get' })
}
