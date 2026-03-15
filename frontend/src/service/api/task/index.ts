import { request } from '@/service/request';
import type { TaskCreateRequest, TaskListParams } from '@/typings/api/task';

const API_BASE = '/api';

export function fetchTaskList(params?: TaskListParams) {
  return request({ url: `${API_BASE}/tasks`, method: 'GET', params });
}

export function fetchTaskDetail(id: number, locale: string = 'zh') {
  return request({ url: `${API_BASE}/tasks/${id}`, method: 'GET', params: { locale } });
}

/** 创建任务 */
export function createTask(data: TaskCreateRequest) {
  return request({ url: `${API_BASE}/tasks`, method: 'POST', data });
}

export function updateTask(id: number, data: Record<string, any>) {
  return request({ url: `${API_BASE}/tasks/${id}/update`, method: 'POST', data });
}

export function deleteTask(id: number) {
  return request({ url: `${API_BASE}/tasks/${id}/delete`, method: 'POST' });
}

export function batchDeleteTasks(ids: number[]) {
  return request({ url: `${API_BASE}/tasks/batch-delete`, method: 'POST', data: { ids } });
}

export function updateSubtask(taskId: number, subtaskId: number, data: Record<string, any>) {
  return request({ url: `${API_BASE}/tasks/${taskId}/subtasks/${subtaskId}/update`, method: 'POST', data });
}

export function expandTask(id: number, data: Record<string, any> = {}) {
  return request({ url: `${API_BASE}/tasks/${id}/expand`, method: 'POST', data, timeout: 5 * 60 * 1000 }); // 5 分钟超时
}

export function clearTaskSubtasks(taskId: number) {
  return request({ url: `${API_BASE}/tasks/${taskId}/subtasks/delete`, method: 'POST' });
}

export function deleteSubtask(taskId: number, subtaskId: number) {
  return request({ url: `${API_BASE}/tasks/${taskId}/subtasks/${subtaskId}/delete`, method: 'POST' });
}

export function regenerateSubtask(taskId: number, subtaskId: number, data: { prompt?: string } = {}) {
  return request({
    url: `${API_BASE}/tasks/${taskId}/subtasks/${subtaskId}/regenerate`,
    method: 'POST',
    data,
    timeout: 5 * 60 * 1000 // 5 分钟超时
  });
}

export function reorderSubtasks(taskId: number, subtaskIds: number[]) {
  return request({
    url: `${API_BASE}/tasks/${taskId}/subtasks/reorder`,
    method: 'POST',
    data: { subtaskIds }
  });
}

export function copyTask(taskId: number) {
  return request({
    url: `${API_BASE}/tasks/${taskId}/copy`,
    method: 'POST',
    timeout: 60 * 1000 // 1 分钟超时
  });
}

// 异步拆分子任务 - 立即返回消息ID
export function expandTaskAsync(id: number, data?: { prompt?: string; additionalContext?: string }) {
  return request<{ messageId: number }>({
    url: `${API_BASE}/tasks/${id}/expand-async`,
    method: 'POST',
    data,
    timeout: 300 * 1000 // 5 分钟超时
  });
}

// 依赖关系相关 API
export function getTaskDependencies() {
  return request({ url: `${API_BASE}/tasks/dependencies`, method: 'GET' });
}

export function addTaskDependency(taskId: number, dependsOnTaskId: number) {
  return request({
    url: `${API_BASE}/tasks/${taskId}/dependencies`,
    method: 'POST',
    data: { dependsOnTaskId }
  });
}

export function removeTaskDependency(taskId: number, dependsOnTaskId: number) {
  return request({
    url: `${API_BASE}/tasks/${taskId}/dependencies/${dependsOnTaskId}`,
    method: 'DELETE'
  });
}

export function validateDependencies() {
  return request({ url: `${API_BASE}/tasks/dependencies/validate`, method: 'GET' });
}

export function getReadyTasks() {
  return request({ url: `${API_BASE}/tasks/ready`, method: 'GET' });
}

// 任务质量评分相关 API
export function scoreTask(id: number) {
  return request({ url: `${API_BASE}/tasks/${id}/score`, method: 'POST' });
}

export function fetchTaskScoreHistory(id: number, page = 1, pageSize = 20) {
  return request({
    url: `${API_BASE}/tasks/${id}/scores`,
    method: 'GET',
    params: { page, pageSize }
  });
}

export function fetchTaskScoreDetail(id: number, scoreId: number) {
  return request({ url: `${API_BASE}/tasks/${id}/scores/${scoreId}`, method: 'GET' });
}

export function restoreTaskScore(id: number, scoreId: number) {
  return request({ url: `${API_BASE}/tasks/${id}/scores/${scoreId}/restore`, method: 'POST' });
}

export function deleteTaskScore(id: number, scoreId: number) {
  return request({ url: `${API_BASE}/tasks/${id}/scores/${scoreId}`, method: 'DELETE' });
}
