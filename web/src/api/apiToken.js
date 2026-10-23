import request from './request'

export const getAPITokens = () => request({ url: '/api-tokens', method: 'get' })
export const createAPIToken = data => request({ url: '/api-tokens', method: 'post', data })
export const updateAPIToken = (id, data) => request({ url: `/api-tokens/${id}`, method: 'put', data })
export const regenerateAPIToken = (id, expiresInDays) => request({ url: `/api-tokens/${id}/regenerate`, method: 'post', data: { expires_in_days: expiresInDays } })
export const updateAPITokenStatus = (id, status) => request({ url: `/api-tokens/${id}/status`, method: 'put', data: { status } })
export const deleteAPIToken = id => request({ url: `/api-tokens/${id}`, method: 'delete' })
