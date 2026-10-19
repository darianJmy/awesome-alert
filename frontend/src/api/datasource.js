import request from './request'

export function getDatasources(params) {
  return request({ url: '/datasources', method: 'get', params })
}

export function createDatasource(data) {
  return request({ url: '/datasources', method: 'post', data })
}

export function updateDatasource(id, data) {
  return request({ url: `/datasources/${id}`, method: 'put', data })
}

export function deleteDatasource(id) {
  return request({ url: `/datasources/${id}`, method: 'delete' })
}

export function batchUpdateDatasourceStatus(ids, status) {
  return request({ url: '/datasources/batch/status', method: 'put', data: { ids, status } })
}

export function batchDeleteDatasources(ids) {
  return request({ url: '/datasources/batch', method: 'delete', data: { ids } })
}

export function testConnection(data) {
  return request({ url: '/datasources/test', method: 'post', data })
}

export function testSavedConnection(id, data) {
  return request({ url: `/datasources/${id}/test`, method: 'post', data })
}
