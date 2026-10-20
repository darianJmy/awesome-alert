import request from './request'

export function getRoles() {
  return request({ url: '/roles', method: 'get' })
}

export function createRole(data) {
  return request({ url: '/roles', method: 'post', data })
}

export function updateRole(id, data) {
  return request({ url: `/roles/${id}`, method: 'put', data })
}

export function deleteRole(id) {
  return request({ url: `/roles/${id}`, method: 'delete' })
}

export function assignPermissions(id, permission_ids) {
  return request({ url: `/roles/${id}/permissions`, method: 'put', data: { permission_ids } })
}

export function assignMenus(id, menu_ids) {
  return request({ url: `/roles/${id}/menus`, method: 'put', data: { menu_ids } })
}
