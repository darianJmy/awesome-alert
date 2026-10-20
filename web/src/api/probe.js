import request from './request'

export function instantProbe(data) {
  const timeoutMs = Number(data?.timeout_ms) || 10000
  return request({ url: '/instant-probe', method: 'post', data, timeout: timeoutMs + 15000 })
}
