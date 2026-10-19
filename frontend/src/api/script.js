import request from './request'

export function getScripts(params) {
  return request({ url: '/scripts', method: 'get', params })
}

export function getScript(id) {
  return request({ url: `/scripts/${id}`, method: 'get' })
}

export function createScript(data) {
  return request({ url: '/scripts', method: 'post', data })
}

export function updateScript(id, data) {
  return request({ url: `/scripts/${id}`, method: 'put', data })
}

export function deleteScript(id) {
  return request({ url: `/scripts/${id}`, method: 'delete' })
}
