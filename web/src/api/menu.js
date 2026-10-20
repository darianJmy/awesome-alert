import request from './request'

export function getMenus() {
  return request({ url: '/menus', method: 'get' })
}

export function createMenu(data) {
  return request({ url: '/menus', method: 'post', data })
}

export function updateMenu(id, data) {
  return request({ url: `/menus/${id}`, method: 'put', data })
}

export function deleteMenu(id) {
  return request({ url: `/menus/${id}`, method: 'delete' })
}

export function getMenuOptions() {
  return request({ url: '/menu-options', method: 'get' })
}
