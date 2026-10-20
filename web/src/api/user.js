import request from './request'

export function getUsers(params) {
  return request({ url: '/users', method: 'get', params })
}

export function createUser(data) {
  return request({ url: '/users', method: 'post', data })
}

export function updateUser(id, data) {
  return request({ url: `/users/${id}`, method: 'put', data })
}

export function deleteUser(id) {
  return request({ url: `/users/${id}`, method: 'delete' })
}

export function assignUserRoles(id, role_ids) {
  return request({ url: `/users/${id}/roles`, method: 'put', data: { role_ids } })
}

export function resetPassword(id) {
  return request({ url: `/users/${id}/reset-password`, method: 'put' })
}
