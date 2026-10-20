import request from './request'

export const getGitCredentials = () => request({ url: '/git-credentials', method: 'get' })
export const createGitCredential = data => request({ url: '/git-credentials', method: 'post', data })
export const updateGitCredential = (id, data) => request({ url: `/git-credentials/${id}`, method: 'put', data })
export const deleteGitCredential = id => request({ url: `/git-credentials/${id}`, method: 'delete' })
