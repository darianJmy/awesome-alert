import request from './request'

export function getAlertRuleGroups(params) {
  return request({ url: '/alert-rule-groups', method: 'get', params })
}

export function getAlertRuleGroup(id) {
  return request({ url: `/alert-rule-groups/${id}`, method: 'get' })
}

export function createAlertRuleGroup(data) {
  return request({ url: '/alert-rule-groups', method: 'post', data })
}

export function updateAlertRuleGroup(id, data) {
  return request({ url: `/alert-rule-groups/${id}`, method: 'put', data })
}

export function deleteAlertRuleGroup(id) {
  return request({ url: `/alert-rule-groups/${id}`, method: 'delete' })
}

export function getAlertRules(params) {
  return request({ url: '/alert-rules', method: 'get', params })
}

export function exportAlertRules(groupId) {
  return request({ url: '/alert-rules/export', method: 'get', params: { group_id: groupId }, responseType: 'blob', timeout: 30000 })
}

export function importAlertRules(groupId, content, overwrite, dryRun = false) {
  return request({ url: '/alert-rules/import', method: 'post', data: { group_id: groupId, content, overwrite, dry_run: dryRun }, timeout: 30000 })
}

export function getAlertRule(id) {
  return request({ url: `/alert-rules/${id}`, method: 'get' })
}

export function createAlertRule(data) {
  return request({ url: '/alert-rules', method: 'post', data })
}

export function updateAlertRule(id, data) {
  return request({ url: `/alert-rules/${id}`, method: 'put', data })
}

export function updateAlertRuleStatus(id, status) {
  return request({ url: `/alert-rules/${id}/status`, method: 'put', data: { status } })
}

export function batchUpdateAlertRuleStatus(ids, status) {
  return request({ url: '/alert-rules/batch/status', method: 'put', data: { ids, status } })
}

export function batchDeleteAlertRules(ids) {
  return request({ url: '/alert-rules/batch', method: 'delete', data: { ids } })
}

export function deleteAlertRule(id) {
  return request({ url: `/alert-rules/${id}`, method: 'delete' })
}

export function runAlertRule(id) {
  return request({ url: `/alert-rules/${id}/run`, method: 'post' })
}

export function getAlertEvaluations(params) {
  return request({ url: '/alert-evaluations', method: 'get', params })
}
