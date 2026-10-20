import request from './request'

export const getMCPService = () => request({ url: '/mcp-service', method: 'get' })
export const getMCPClients = params => request({ url: '/mcp-clients', method: 'get', params })
export const createMCPClient = data => request({ url: '/mcp-clients', method: 'post', data })
export const updateMCPClient = (id, data) => request({ url: `/mcp-clients/${id}`, method: 'put', data })
export const updateMCPClientStatus = (id, status) => request({ url: `/mcp-clients/${id}/status`, method: 'put', data: { status } })
export const deleteMCPClient = id => request({ url: `/mcp-clients/${id}`, method: 'delete' })
