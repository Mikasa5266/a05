import request from '../utils/request'

export function getGrowthStats() {
  return request({
    url: '/growth/stats',
    method: 'get'
  })
}
