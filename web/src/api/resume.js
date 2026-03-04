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
