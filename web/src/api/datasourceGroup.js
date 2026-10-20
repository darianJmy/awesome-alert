import request from './request'

export function getDatasourceGroups(params) {
  return request({ url: '/datasource-groups', method: 'get', params })
}

export function createDatasourceGroup(data) {
  return request({ url: '/datasource-groups', method: 'post', data })
}

export function updateDatasourceGroup(id, data) {
  return request({ url: `/datasource-groups/${id}`, method: 'put', data })
}

export function deleteDatasourceGroup(id) {
  return request({ url: `/datasource-groups/${id}`, method: 'delete' })
}
