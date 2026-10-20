import request from './request'

export function getAlerts(params) {
  return request({ url: '/alerts', method: 'get', params })
}

export function getAlertEvents(id) {
  return request({ url: `/alerts/${id}/events`, method: 'get' })
}

export function getFaultCenters() {
  return request({ url: '/fault-centers', method: 'get' })
}

export function createFaultCenter(data) {
  return request({ url: '/fault-centers', method: 'post', data })
}

export function updateFaultCenter(id, data) {
  return request({ url: `/fault-centers/${id}`, method: 'put', data })
}

export function deleteFaultCenter(id) {
  return request({ url: `/fault-centers/${id}`, method: 'delete' })
}

export function getGroupingRules(centerId) {
  return request({ url: `/fault-centers/${centerId}/grouping-rules`, method: 'get' })
}

export function createGroupingRule(centerId, data) {
  return request({ url: `/fault-centers/${centerId}/grouping-rules`, method: 'post', data })
}

export function updateGroupingRule(centerId, ruleId, data) {
  return request({ url: `/fault-centers/${centerId}/grouping-rules/${ruleId}`, method: 'put', data })
}

export function deleteGroupingRule(centerId, ruleId) {
  return request({ url: `/fault-centers/${centerId}/grouping-rules/${ruleId}`, method: 'delete' })
}

export function getInhibitionRules(centerId) {
  return request({ url: `/fault-centers/${centerId}/inhibition-rules`, method: 'get' })
}

export function createInhibitionRule(centerId, data) {
  return request({ url: `/fault-centers/${centerId}/inhibition-rules`, method: 'post', data })
}

export function updateInhibitionRule(centerId, ruleId, data) {
  return request({ url: `/fault-centers/${centerId}/inhibition-rules/${ruleId}`, method: 'put', data })
}

export function deleteInhibitionRule(centerId, ruleId) {
  return request({ url: `/fault-centers/${centerId}/inhibition-rules/${ruleId}`, method: 'delete' })
}

export function getSilenceRules(centerId) {
  return request({ url: `/fault-centers/${centerId}/silence-rules`, method: 'get' })
}

export function createSilenceRule(centerId, data) {
  return request({ url: `/fault-centers/${centerId}/silence-rules`, method: 'post', data })
}

export function updateSilenceRule(centerId, ruleId, data) {
  return request({ url: `/fault-centers/${centerId}/silence-rules/${ruleId}`, method: 'put', data })
}

export function deleteSilenceRule(centerId, ruleId) {
  return request({ url: `/fault-centers/${centerId}/silence-rules/${ruleId}`, method: 'delete' })
}

export function getFaultCenterNotificationConfig(centerId) {
  return request({ url: `/fault-centers/${centerId}/notification-config`, method: 'get' })
}

export function updateFaultCenterNotificationConfig(centerId, data) {
  return request({ url: `/fault-centers/${centerId}/notification-config`, method: 'put', data })
}

export function silenceAlert(id, data) {
  return request({ url: `/alerts/${id}/silence`, method: 'put', data })
}

export function unsilenceAlert(id) {
  return request({ url: `/alerts/${id}/silence`, method: 'delete' })
}

export function getAlertActionRules(centerId) {
  return request({ url: `/fault-centers/${centerId}/action-rules`, method: 'get' })
}

export function createAlertActionRule(centerId, data) {
  return request({ url: `/fault-centers/${centerId}/action-rules`, method: 'post', data })
}

export function updateAlertActionRule(centerId, ruleId, data) {
  return request({ url: `/fault-centers/${centerId}/action-rules/${ruleId}`, method: 'put', data })
}

export function deleteAlertActionRule(centerId, ruleId) {
  return request({ url: `/fault-centers/${centerId}/action-rules/${ruleId}`, method: 'delete' })
}
