/**
 * activity-manager.js
 * 活动日志业务逻辑模块
 */

import {
  logActivity,
  logActivities,
  getTaskActivities,
  getGlobalActivities,
  getActivityStatistics,
  cleanOldActivities
} from './activity-storage.js';

/**
 * 活动类型常量
 */
export const ACTIVITY_TYPES = {
  // 任务相关
  TASK_CREATED: 'task_created',
  TASK_UPDATED: 'task_updated',
  TASK_DELETED: 'task_deleted',
  TASK_STATUS_CHANGED: 'task_status_changed',
  TASK_PRIORITY_CHANGED: 'task_priority_changed',
  TASK_ASSIGNED: 'task_assigned',
  TASK_UNASSIGNED: 'task_unassigned',

  // 子任务相关
  SUBTASK_CREATED: 'subtask_created',
  SUBTASK_UPDATED: 'subtask_updated',
  SUBTASK_DELETED: 'subtask_deleted',
  SUBTASK_STATUS_CHANGED: 'subtask_status_changed',
  SUBTASK_ASSIGNED: 'subtask_assigned',
  SUBTASK_UNASSIGNED: 'subtask_unassigned',

  // 评论相关
  COMMENT_ADDED: 'comment_added',
  COMMENT_UPDATED: 'comment_updated',
  COMMENT_DELETED: 'comment_deleted',

  // 时间相关
  TIME_ESTIMATED: 'time_estimated',
  TIME_LOGGED: 'time_logged',
  DUE_DATE_CHANGED: 'due_date_changed',

  // 其他
  TASK_EXPANDED: 'task_expanded',
  SUBTASK_REGENERATED: 'subtask_regenerated'
};

/**
 * 活动类型描述映射
 */
export const ACTIVITY_DESCRIPTIONS = {
  [ACTIVITY_TYPES.TASK_CREATED]: '创建任务',
  [ACTIVITY_TYPES.TASK_UPDATED]: '更新任务',
  [ACTIVITY_TYPES.TASK_DELETED]: '删除任务',
  [ACTIVITY_TYPES.TASK_STATUS_CHANGED]: '更改状态',
  [ACTIVITY_TYPES.TASK_PRIORITY_CHANGED]: '更改优先级',
  [ACTIVITY_TYPES.TASK_ASSIGNED]: '分配任务',
  [ACTIVITY_TYPES.TASK_UNASSIGNED]: '取消分配',
  [ACTIVITY_TYPES.SUBTASK_CREATED]: '创建子任务',
  [ACTIVITY_TYPES.SUBTASK_UPDATED]: '更新子任务',
  [ACTIVITY_TYPES.SUBTASK_DELETED]: '删除子任务',
  [ACTIVITY_TYPES.SUBTASK_STATUS_CHANGED]: '更改子任务状态',
  [ACTIVITY_TYPES.SUBTASK_ASSIGNED]: '分配子任务',
  [ACTIVITY_TYPES.SUBTASK_UNASSIGNED]: '取消子任务分配',
  [ACTIVITY_TYPES.COMMENT_ADDED]: '添加评论',
  [ACTIVITY_TYPES.COMMENT_UPDATED]: '更新评论',
  [ACTIVITY_TYPES.COMMENT_DELETED]: '删除评论',
  [ACTIVITY_TYPES.TIME_ESTIMATED]: '设置预估工时',
  [ACTIVITY_TYPES.TIME_LOGGED]: '记录工时',
  [ACTIVITY_TYPES.DUE_DATE_CHANGED]: '更改截止日期',
  [ACTIVITY_TYPES.TASK_EXPANDED]: '展开任务为子任务',
  [ACTIVITY_TYPES.SUBTASK_REGENERATED]: '重新生成子任务'
};

/**
 * 记录任务创建
 */
export async function logTaskCreated(taskId, memberId, taskTitle) {
  await logActivity({
    taskId,
    memberId,
    action: ACTIVITY_TYPES.TASK_CREATED,
    description: `创建了任务: ${taskTitle}`
  });
}

/**
 * 记录状态变更
 */
export async function logStatusChange(taskId, memberId, oldStatus, newStatus, subtaskId = null) {
  await logActivity({
    taskId,
    subtaskId,
    memberId,
    action: subtaskId ? ACTIVITY_TYPES.SUBTASK_STATUS_CHANGED : ACTIVITY_TYPES.TASK_STATUS_CHANGED,
    fieldName: 'status',
    oldValue: oldStatus,
    newValue: newStatus,
    description: `状态从 "${oldStatus}" 变更为 "${newStatus}"`
  });
}

/**
 * 记录任务分配
 */
export async function logTaskAssigned(taskId, memberId, assignedMemberName, subtaskId = null) {
  await logActivity({
    taskId,
    subtaskId,
    memberId,
    action: subtaskId ? ACTIVITY_TYPES.SUBTASK_ASSIGNED : ACTIVITY_TYPES.TASK_ASSIGNED,
    description: `分配给 ${assignedMemberName}`
  });
}

/**
 * 记录评论添加
 */
export async function logCommentAdded(taskId, memberId, commentPreview, subtaskId = null) {
  const preview = commentPreview.length > 50 ? `${commentPreview.substring(0, 50)}...` : commentPreview;
  await logActivity({
    taskId,
    subtaskId,
    memberId,
    action: ACTIVITY_TYPES.COMMENT_ADDED,
    description: `添加评论: "${preview}"`
  });
}

/**
 * 记录工时
 */
export async function logTimeLogged(taskId, memberId, hours, subtaskId = null) {
  await logActivity({
    taskId,
    subtaskId,
    memberId,
    action: ACTIVITY_TYPES.TIME_LOGGED,
    fieldName: 'actualHours',
    newValue: String(hours),
    description: `记录工时: ${hours} 小时`
  });
}

/**
 * 获取任务活动日志
 */
export async function listTaskActivities(taskId, options = {}) {
  return getTaskActivities(taskId, options);
}

/**
 * 获取全局活动日志
 */
export async function listGlobalActivities(options = {}) {
  return getGlobalActivities(options);
}

/**
 * 获取活动统计
 */
export async function getActivityStats(options = {}) {
  const stats = await getActivityStatistics(options);

  return {
    ...stats,
    actionDescriptions: ACTIVITY_DESCRIPTIONS
  };
}

/**
 * 清理旧日志
 */
export async function cleanupOldLogs(daysToKeep = 90) {
  return cleanOldActivities(daysToKeep);
}

export default {
  ACTIVITY_TYPES,
  ACTIVITY_DESCRIPTIONS,
  logTaskCreated,
  logStatusChange,
  logTaskAssigned,
  logCommentAdded,
  logTimeLogged,
  logActivity,
  listTaskActivities,
  listGlobalActivities,
  getActivityStats,
  cleanupOldLogs
};
