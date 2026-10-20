import request from './request'

export function getNotificationChannels() {
  return request({ url: '/notification-channels', method: 'get' })
}

export function createNotificationChannel(data) {
  return request({ url: '/notification-channels', method: 'post', data })
}

export function updateNotificationChannel(id, data) {
  return request({ url: `/notification-channels/${id}`, method: 'put', data })
}

export function deleteNotificationChannel(id) {
  return request({ url: `/notification-channels/${id}`, method: 'delete' })
}

export function testNotificationChannel(id) {
  return request({ url: `/notification-channels/${id}/test`, method: 'post' })
}

export function getNotificationRecords(params) {
  return request({ url: '/notification-records', method: 'get', params })
}

export function getNotificationTemplates() {
  return request({ url: '/notification-templates', method: 'get' })
}

export function createNotificationTemplate(data) {
  return request({ url: '/notification-templates', method: 'post', data })
}

export function updateNotificationTemplate(id, data) {
  return request({ url: `/notification-templates/${id}`, method: 'put', data })
}

export function deleteNotificationTemplate(id) {
  return request({ url: `/notification-templates/${id}`, method: 'delete' })
}
