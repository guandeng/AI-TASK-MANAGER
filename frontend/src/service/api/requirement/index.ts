import { request } from '@/service/request';
import type {
  Requirement,
  RequirementDocument,
  RequirementListParams,
  RequirementFormData,
  RequirementStatistics
} from '@/typings/api/requirement';

/** 获取需求列表 */
export function fetchRequirementList(params?: RequirementListParams) {
  return request<Requirement[]>({
    url: '/api/requirements',
    method: 'get',
    params
  });
}

/** 获取需求统计 */
export function fetchRequirementStatistics() {
  return request<RequirementStatistics>({
    url: '/api/requirements/statistics',
    method: 'get'
  });
}

/** 获取需求详情 */
export function fetchRequirementDetail(id: number) {
  return request<Requirement>({
    url: `/api/requirements/${id}`,
    method: 'get'
  });
}

/** 创建需求 */
export function createRequirement(data: RequirementFormData) {
  return request<Requirement>({
    url: '/api/requirements',
    method: 'post',
    data
  });
}

/** 更新需求 */
export function updateRequirement(id: number, data: Partial<RequirementFormData>) {
  return request<Requirement>({
    url: `/api/requirements/${id}/update`,
    method: 'post',
    data
  });
}

/** 删除需求 */
export function deleteRequirement(id: number) {
  return request<{ success: boolean; message: string }>({
    url: `/api/requirements/${id}/delete`,
    method: 'post'
  });
}

/** 上传文档 */
export function uploadDocument(requirementId: number, file: File, uploadedBy?: string) {
  const formData = new FormData();
  formData.append('file', file);
  if (uploadedBy) {
    formData.append('uploadedBy', uploadedBy);
  }

  return request<RequirementDocument>({
    url: `/api/requirements/${requirementId}/documents`,
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  });
}

/** 删除文档 */
export function deleteDocument(requirementId: number, documentId: number) {
  return request<{ success: boolean; message: string }>({
    url: `/api/requirements/${requirementId}/documents/${documentId}/delete`,
    method: 'post'
  });
}

/** 获取文档下载链接 */
export function getDocumentDownloadUrl(requirementId: number, documentId: number) {
  return `/api/requirements/${requirementId}/documents/${documentId}/download`;
}

/** 任务类型 */
export type TaskType = 'frontend' | 'backend' | 'fullstack';

/** 将需求拆分为任务 */
export function splitRequirementToTasks(requirementId: number, taskType: TaskType = 'backend') {
  return request<{
    success: boolean;
    message: string;
    tasks: Array<{
      id: number;
      title: string;
      description: string;
      details: string;
      status: string;
      priority: string;
      dependencies: number[];
    }>;
  }>({
    url: `/api/requirements/${requirementId}/split-tasks`,
    method: 'post',
    data: { taskType }
  });
}

/** 将需求拆分为任务（异步） */
export function splitRequirementToTasksAsync(requirementId: number, taskType: TaskType = 'backend') {
  return request<{
    messageId: number;
    message: string;
  }>({
    url: `/api/requirements/${requirementId}/split-tasks-async`,
    method: 'post',
    data: { taskType }
  });
}
