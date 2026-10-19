import request from './request'

export function getWorkflows() {
  return request({ url: '/workflows', method: 'get' })
}

export function createWorkflow(data) {
  return request({ url: '/workflows', method: 'post', data })
}

export function updateWorkflow(id, data) {
  return request({ url: `/workflows/${id}`, method: 'put', data })
}

export function deleteWorkflow(id) {
  return request({ url: `/workflows/${id}`, method: 'delete' })
}

export function runWorkflow(id) {
  return request({ url: `/workflows/${id}/run`, method: 'post' })
}
