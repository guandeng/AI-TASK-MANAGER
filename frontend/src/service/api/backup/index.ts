import { request } from '@/service/request';

export interface BackupRecord {
  id: number;
  requirementId: number;
  type: string;
  dataSnapshot: string;
  taskCount: number;
  status: string;
  errorMessage?: string;
  createdAt: string;
}

export interface BackupSchedule {
  id?: number;
  requirementId: number;
  enabled: boolean;
  intervalType: 'minute' | 'hour';
  intervalValue: number;
  retentionCount: number;
  lastBackupAt?: string;
  createdAt?: string;
  updatedAt?: string;
}

/**
 * 获取备份列表
 */
export function fetchBackupList(requirementId: number, params?: { page?: number; pageSize?: number }) {
  return request({ url: `/api/requirements/${requirementId}/backups`, method: 'GET', params });
}

/**
 * 立即创建备份
 */
export function createBackup(requirementId: number) {
  return request({ url: `/api/requirements/${requirementId}/backups/create`, method: 'POST' });
}

/**
 * 恢复备份
 */
export function restoreBackup(requirementId: number, backupId: number) {
  return request({ url: `/api/requirements/${requirementId}/backups/${backupId}/restore`, method: 'POST' });
}

/**
 * 删除备份
 */
export function deleteBackup(requirementId: number, backupId: number) {
  return request({ url: `/api/requirements/${requirementId}/backups/${backupId}/delete`, method: 'POST' });
}

/**
 * 获取备份计划
 */
export function fetchBackupSchedule(requirementId: number) {
  return request<BackupSchedule>({ url: `/api/requirements/${requirementId}/backups/schedule`, method: 'GET' });
}

/**
 * 更新备份计划
 */
export function updateBackupSchedule(requirementId: number, data: Partial<BackupSchedule>) {
  return request({ url: `/api/requirements/${requirementId}/backups/schedule/update`, method: 'POST', data });
}

/**
 * 禁用备份计划
 */
export function disableBackupSchedule(requirementId: number) {
  return request({ url: `/api/requirements/${requirementId}/backups/schedule/disable`, method: 'POST' });
}
