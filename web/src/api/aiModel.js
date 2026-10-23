import request from './request'
export const getAIModels = () => request({ url: '/ai-models', method: 'get' })
export const saveAIModel = (id, data) => request({ url: id ? `/ai-models/${id}` : '/ai-models', method: id ? 'put' : 'post', data })
export const deleteAIModel = id => request({ url: `/ai-models/${id}`, method: 'delete' })
export const getChatModels = () => request({ url: '/chat/models', method: 'get' })
