import request from '../utils/request'

export function getReports(params) {
  return request({
    url: '/reports',
    method: 'get',
    params
  })
}

export function getReport(id) {
  return request({
    url: `/reports/${id}`,
    method: 'get'
  })
}

export function generateReport(data) {
  return request({
    url: '/reports/generate',
    method: 'post',
    data
  })
}
