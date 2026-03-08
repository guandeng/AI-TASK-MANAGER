import { request } from '@/service/request';

const API_BASE = '/api';

export function fetchTaskList() {
  return request({ url: `${API_BASE}/tasks`, method: 'GET' });
}

export function fetchTaskDetail(id: number, locale: string = 'zh') {
  return request({ url: `${API_BASE}/tasks/${id}`, method: 'GET', params: { locale } });
}

export function updateTask(id: number, data: Record<string, any>) {
  return request({ url: `${API_BASE}/tasks/${id}`, method: 'PUT', data });
}

export function updateSubtask(taskId: number, subtaskId: number, data: Record<string, any>) {
  return request({ url: `${API_BASE}/tasks/${taskId}/subtasks/${subtaskId}`, method: 'PUT', data });
}
