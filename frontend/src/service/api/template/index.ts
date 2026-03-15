import { request } from '@/service/request';
import type {
  ProjectTemplate,
  TaskTemplate,
  CreateProjectTemplateRequest,
  CreateTaskTemplateRequest,
  InstantiateTemplateRequest
} from '@/typings/api/template';

const API_BASE = '/api';

// ============ 项目模板 ============

/**
 * 获取项目模板列表
 */
export function fetchProjectTemplates(params?: { category?: string; keyword?: string }) {
  return request<ProjectTemplate[]>({
    url: `${API_BASE}/templates/projects`,
    method: 'GET',
    params
  });
}

/**
 * 获取项目模板详情
 */
export function fetchProjectTemplate(id: number) {
  return request<ProjectTemplate>({
    url: `${API_BASE}/templates/projects/${id}`,
    method: 'GET'
  });
}

/**
 * 创建项目模板
 */
export function createProjectTemplate(data: CreateProjectTemplateRequest) {
  return request<ProjectTemplate>({
    url: `${API_BASE}/templates/projects`,
    method: 'POST',
    data
  });
}

/**
 * 更新项目模板
 */
export function updateProjectTemplate(id: number, data: Partial<CreateProjectTemplateRequest>) {
  return request<ProjectTemplate>({
    url: `${API_BASE}/templates/projects/${id}/update`,
    method: 'POST',
    data
  });
}

/**
 * 删除项目模板
 */
export function deleteProjectTemplate(id: number) {
  return request<{ success: boolean; message: string }>({
    url: `${API_BASE}/templates/projects/${id}/delete`,
    method: 'POST'
  });
}

/**
 * 从项目创建模板
 */
export function createTemplateFromProject(requirementId: number, data: { name: string; description?: string }) {
  return request<ProjectTemplate>({
    url: `${API_BASE}/requirements/${requirementId}/create-template`,
    method: 'POST',
    data
  });
}

/**
 * 实例化项目模板
 */
export function instantiateProjectTemplate(id: number, data?: InstantiateTemplateRequest) {
  return request<{ requirementId: number; taskIds: number[] }>({
    url: `${API_BASE}/templates/projects/${id}/instantiate`,
    method: 'POST',
    data
  });
}

/**
 * 对项目模板进行评分（同步）
 */
export function scoreProjectTemplate(data: { id: number }) {
  return request<any>({
    url: `${API_BASE}/templates/projects/score`,
    method: 'POST',
    data
  });
}

/**
 * 对项目模板进行评分（异步）
 */
export function scoreProjectTemplateAsync(data: { id: number }) {
  return request<{ messageId: number; message: string }>({
    url: `${API_BASE}/templates/projects/score-async`,
    method: 'POST',
    data
  });
}

// ============ 独立任务模板 ============

/**
 * 获取任务模板列表
 */
export function fetchTaskTemplates(params?: { keyword?: string }) {
  return request<TaskTemplate[]>({
    url: `${API_BASE}/templates/tasks`,
    method: 'GET',
    params
  });
}

/**
 * 获取任务模板详情
 */
export function fetchTaskTemplate(id: number) {
  return request<TaskTemplate>({
    url: `${API_BASE}/templates/tasks/${id}`,
    method: 'GET'
  });
}

/**
 * 创建任务模板
 */
export function createTaskTemplate(data: CreateTaskTemplateRequest) {
  return request<TaskTemplate>({
    url: `${API_BASE}/templates/tasks`,
    method: 'POST',
    data
  });
}

/**
 * 更新任务模板
 */
export function updateTaskTemplate(id: number, data: Partial<CreateTaskTemplateRequest>) {
  return request<TaskTemplate>({
    url: `${API_BASE}/templates/tasks/${id}/update`,
    method: 'POST',
    data
  });
}

/**
 * 删除任务模板
 */
export function deleteTaskTemplate(id: number) {
  return request<{ success: boolean; message: string }>({
    url: `${API_BASE}/templates/tasks/${id}/delete`,
    method: 'POST'
  });
}

/**
 * 从任务创建模板
 */
export function createTemplateFromTask(taskId: number, data: { name: string; description?: string }) {
  return request<TaskTemplate>({
    url: `${API_BASE}/tasks/${taskId}/create-template`,
    method: 'POST',
    data
  });
}

/**
 * 实例化任务模板
 */
export function instantiateTaskTemplate(id: number, data?: { requirementId?: number; title?: string }) {
  return request<{ taskId: number }>({
    url: `${API_BASE}/templates/tasks/${id}/instantiate`,
    method: 'POST',
    data
  });
}
