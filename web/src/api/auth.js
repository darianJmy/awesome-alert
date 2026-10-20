import request from './request'

export function login(data) {
  return request({ url: '/login', method: 'post', data })
}

export function getUserInfo() {
  return request({ url: '/user/info', method: 'get' })
}

export function changePassword(data) {
  return request({ url: '/user/password', method: 'put', data })
}
