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

export function getQuestions(params) {
  return request({
    url: '/questions',
    method: 'get',
    params
  })
}