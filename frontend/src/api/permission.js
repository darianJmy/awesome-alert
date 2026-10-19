import request from './request'

export function getPermissions() {
  return request({ url: '/permissions', method: 'get' })
}
