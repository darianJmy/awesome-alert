import request from './request'

export function getCertificateExecutors() {
  return request({ url: '/certificate-executors', method: 'get' })
}

export function createCertificateExecutor(data) {
  return request({ url: '/certificate-executors', method: 'post', data })
}

export function updateCertificateExecutor(id, data) {
  return request({ url: `/certificate-executors/${id}`, method: 'put', data })
}

export function deleteCertificateExecutor(id) {
  return request({ url: `/certificate-executors/${id}`, method: 'delete' })
}

export function testCertificateExecutor(id) {
  return request({ url: `/certificate-executors/${id}/test`, method: 'post' })
}

export function getAgents() {
  return request({ url: '/agents', method: 'get' })
}

export function createAgent(data) {
  return request({ url: '/agents', method: 'post', data })
}

export function updateAgent(id, data) {
  return request({ url: `/agents/${id}`, method: 'put', data })
}

export function deleteAgent(id) {
  return request({ url: `/agents/${id}`, method: 'delete' })
}

export function rotateAgentToken(id) {
  return request({ url: `/agents/${id}/rotate-token`, method: 'post' })
}

export function runAgentCommand(id, data) {
  const timeoutSeconds = Number(data?.timeout_seconds) || 300
  return request({
    url: `/agents/${id}/run`,
    method: 'post',
    data,
    timeout: (timeoutSeconds + 10) * 1000,
  })
}
