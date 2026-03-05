import request from '../utils/request'

// ===== 高校仪表盘 =====
export function getUniversityDashboard() {
  return request({ url: '/university/dashboard', method: 'get' })
}

// ===== 学生跟踪 =====
export function getStudentTracking(params) {
  return request({ url: '/university/students', method: 'get', params })
}

export function getStudentDetail(studentId) {
  return request({ url: `/university/students/${studentId}`, method: 'get' })
}

export function updateStudentRisk(studentId, data) {
  return request({ url: `/university/students/${studentId}/risk`, method: 'put', data })
}

// ===== 帮扶体系 =====
export function getRiskGroups() {
  return request({ url: '/university/risk-groups', method: 'get' })
}

export function assignMentor(data) {
  return request({ url: '/university/mentor/assign', method: 'post', data })
}

export function batchSupport(data) {
  return request({ url: '/university/support/batch', method: 'post', data })
}

export function recommendCourse(data) {
  return request({ url: '/university/support/recommend-course', method: 'post', data })
}

// ===== 课程与资源 =====
export function getCourses(params) {
  return request({ url: '/university/courses', method: 'get', params })
}

export function createCourse(data) {
  return request({ url: '/university/courses', method: 'post', data })
}

export function getResources(params) {
  return request({ url: '/university/resources', method: 'get', params })
}

// ===== 就业数据 =====
export function getEmploymentStats(params) {
  return request({ url: '/university/employment/stats', method: 'get', params })
}

export function getMajorEmployment() {
  return request({ url: '/university/employment/by-major', method: 'get' })
}

export function getSalaryDistribution() {
  return request({ url: '/university/employment/salary', method: 'get' })
}

export function getCityDistribution() {
  return request({ url: '/university/employment/city', method: 'get' })
}

export function getIndustryDistribution() {
  return request({ url: '/university/employment/industry', method: 'get' })
}

// ===== 人才推送 =====
export function getRecommendedStudents(params) {
  return request({ url: '/university/talent-push/recommended', method: 'get', params })
}

export function pushStudentsToEnterprise(data) {
  return request({ url: '/university/talent-push', method: 'post', data })
}

export function getPushHistory(params) {
  return request({ url: '/university/talent-push/history', method: 'get', params })
}
