import request from './request'

export function getScheduledTasks() { return request({ url: '/scheduled-tasks', method: 'get' }) }
export function createScheduledTask(data) { return request({ url: '/scheduled-tasks', method: 'post', data }) }
export function updateScheduledTask(id, data) { return request({ url: `/scheduled-tasks/${id}`, method: 'put', data }) }
export function deleteScheduledTask(id) { return request({ url: `/scheduled-tasks/${id}`, method: 'delete' }) }
export function runScheduledTask(id) { return request({ url: `/scheduled-tasks/${id}/run`, method: 'post' }) }
export function getTaskExecutions(params) { return request({ url: '/scheduled-task-executions', method: 'get', params }) }
export function getCronHelp() { return request({ url: '/cron-help', method: 'get' }) }
