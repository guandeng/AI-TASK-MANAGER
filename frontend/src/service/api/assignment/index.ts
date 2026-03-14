import { request } from '@/service/request';
import type {
  Assignment,
  AssignmentOverview,
  MemberAssignment,
  MemberWorkload,
  CreateAssignmentRequest,
  TaskTimeInfo
} from '@/typings/api/assignment';

const API_BASE = '/api';

/**
 * 获取任务分配列表
 */
export function fetchTaskAssignments(taskId: number) {
  return request<Assignment[]>({
    url: `${API_BASE}/tasks/${taskId}/assignments`,
    method: 'GET'
  });
}

/**
 * 获取任务分配概览
 */
export function fetchAssignmentOverview(taskId: number) {
  return request<AssignmentOverview>({
    url: `${API_BASE}/tasks/${taskId}/assignments/overview`,
    method: 'GET'
  });
}

/**
 * 分配任务给成员
 */
export function assignTaskToMember(taskId: number, data: CreateAssignmentRequest) {
  return request<Assignment[]>({
    url: `${API_BASE}/tasks/${taskId}/assignments`,
    method: 'POST',
    data
  });
}

/**
 * 移除任务分配
 */
export function unassignTaskFromMember(taskId: number, assignmentId: number) {
  return request<{ success: boolean; message: string }>({
    url: `${API_BASE}/tasks/${taskId}/assignments/${assignmentId}`,
    method: 'DELETE'
  });
}

/**
 * 获取子任务分配列表
 */
export function fetchSubtaskAssignments(taskId: number, subtaskId: number) {
  return request<Assignment[]>({
    url: `${API_BASE}/tasks/${taskId}/subtasks/${subtaskId}/assignments`,
    method: 'GET'
  });
}

/**
 * 分配子任务给成员
 */
export function assignSubtaskToMember(taskId: number, subtaskId: number, data: CreateAssignmentRequest) {
  return request<Assignment[]>({
    url: `${API_BASE}/tasks/${taskId}/subtasks/${subtaskId}/assignments`,
    method: 'POST',
    data
  });
}

/**
 * 移除子任务分配
 */
export function unassignSubtaskFromMember(taskId: number, subtaskId: number, assignmentId: number) {
  return request<{ success: boolean; message: string }>({
    url: `${API_BASE}/tasks/${taskId}/subtasks/${subtaskId}/assignments/${assignmentId}`,
    method: 'DELETE'
  });
}

/**
 * 获取成员任务分配列表
 */
export function fetchMemberAssignments(memberId: number, filters?: { role?: string; status?: string; limit?: number }) {
  return request<MemberAssignment[]>({
    url: `${API_BASE}/members/${memberId}/assignments`,
    method: 'GET',
    params: filters
  });
}

/**
 * 获取成员工作量
 */
export function fetchMemberWorkload(memberId: number) {
  return request<MemberWorkload>({
    url: `${API_BASE}/members/${memberId}/workload`,
    method: 'GET'
  });
}

/**
 * 更新任务时间信息
 */
export function updateTaskTime(taskId: number, data: TaskTimeInfo) {
  return request<{ success: boolean; message: string }>({
    url: `${API_BASE}/tasks/${taskId}/time`,
    method: 'PUT',
    data
  });
}
