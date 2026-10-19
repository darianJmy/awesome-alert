import request from './request'

export function getConversations() { return request({ url: '/conversations', method: 'get' }) }
export function createConversation(data) { return request({ url: '/conversations', method: 'post', data }) }
export function deleteConversation(id) { return request({ url: `/conversations/${id}`, method: 'delete' }) }
export function getMessages(id) { return request({ url: `/conversations/${id}/messages`, method: 'get' }) }
export function sendMessage(data) { return request({ url: '/chat', method: 'post', data, timeout: 180000 }) }
