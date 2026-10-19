import request from './request'

export const getPublicSettings = () => request({ url: '/public/settings', method: 'get' })
export const getSystemSettings = () => request({ url: '/system-settings', method: 'get' })
export const updateSystemSettings = data => request({ url: '/system-settings', method: 'put', data })
export const getAuthSettings = () => request({ url: '/auth-settings', method: 'get' })
export const updateAuthSettings = data => request({ url: '/auth-settings', method: 'put', data })
export const testLDAPSettings = data => request({ url: '/auth-settings/test-ldap', method: 'post', data })
export const testOIDCSettings = data => request({ url: '/auth-settings/test-oidc', method: 'post', data })
export const completeOIDCLogin = () => request({ url: '/auth/oidc/complete', method: 'post' })
