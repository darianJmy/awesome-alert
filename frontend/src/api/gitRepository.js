import request from './request'

export const getGitRepositories = () => request({ url: '/git-repositories', method: 'get' })
export const createGitRepository = data => request({ url: '/git-repositories', method: 'post', data })
export const updateGitRepository = (id, data) => request({ url: `/git-repositories/${id}`, method: 'put', data })
export const deleteGitRepository = id => request({ url: `/git-repositories/${id}`, method: 'delete' })
export const syncGitRepository = id => request({ url: `/git-repositories/${id}/sync`, method: 'post', timeout: 310000 })
