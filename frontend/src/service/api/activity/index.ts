import { request } from '@/service/request';
import type { Activity, ActivityStatistics, ActivityQueryParams } from '@/typings/api/activity';

const API_BASE = '/api';

/**
 * 获取任务活动日志
 */
export function fetchTaskActivities(taskId: number, options?: { subtaskId?: number; action?: string; limit?: number }) {
  return request<Activity[]>({
    url: `${API_BASE}/tasks/${taskId}/activities`,
    method: 'GET',
    params: options
  });
}

/**
 * 获取全局活动日志
 */
export function fetchGlobalActivities(params?: ActivityQueryParams) {
  return request<Activity[]>({
    url: `${API_BASE}/activities`,
    method: 'GET',
    params
  });
}

/**
 * 获取活动统计
 */
export function fetchActivityStatistics(params?: { startDate?: string; endDate?: string }) {
  return request<ActivityStatistics>({
    url: `${API_BASE}/activities/statistics`,
    method: 'GET',
    params
  });
}
