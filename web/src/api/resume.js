import request from '../utils/request'

export function uploadResumeForAnalysis(formData) {
  return request({
    url: '/resume/parse',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data',
      'X-Skip-Error-Toast': 'true',
    },
    timeout: 120000,
  })
}

export function getLatestResumeAnalysis() {
  return request({
    url: '/resume/latest',
    method: 'get',
    headers: {
      'X-Skip-Error-Toast': 'true',
    },
  })
}
