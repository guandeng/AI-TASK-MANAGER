/**
 * assignment-manager.js
 * 任务分配业务逻辑模块
 */

import {
  getTaskAssignments,
  getSubtaskAssignments,
  assignTaskToMember,
  assignSubtaskToMember,
  removeTaskAssignment,
  removeSubtaskAssignment,
  getMemberAssignments,
  getMemberWorkload,
  updateTaskTimeFields
} from './assignment-storage.js';
import { getMemberById } from './member-storage.js';

/**
 * 获取任务的所有分配
 */
export async function listTaskAssignments(taskId) {
  return getTaskAssignments(taskId);
}

/**
 * 获取子任务的所有分配
 */
export async function listSubtaskAssignments(subtaskId) {
  return getSubtaskAssignments(subtaskId);
}

/**
 * 分配任务给成员
 */
export async function assignTask(taskId, memberId, data = {}) {
  // 验证成员存在
  const member = await getMemberById(memberId);
  if (!member) {
    throw new Error('成员不存在');
  }

  if (member.status !== 'active') {
    throw new Error('该成员已停用，无法分配任务');
  }

  return assignTaskToMember(taskId, memberId, data);
}

/**
 * 分配子任务给成员
 */
export async function assignSubtask(subtaskId, memberId, data = {}) {
  // 验证成员存在
  const member = await getMemberById(memberId);
  if (!member) {
    throw new Error('成员不存在');
  }

  if (member.status !== 'active') {
    throw new Error('该成员已停用，无法分配任务');
  }

  return assignSubtaskToMember(subtaskId, memberId, data);
}

/**
 * 移除任务配
 */
export async function unassignTask(taskId, assignmentId) {
  return removeTaskAssignment(taskId, assignmentId);
}

/**
 * 移除子任务分配
 */
export async function unassignSubtask(subtaskId, assignmentId) {
  return removeSubtaskAssignment(subtaskId, assignmentId);
}

/**
 * 获取成员的任务分配列表
 */
export async function listMemberAssignments(memberId, filters = {}) {
  return getMemberAssignments(memberId, filters);
}

/**
 * 获取成员工作量统计
 */
export async function getWorkload(memberId) {
  return getMemberWorkload(memberId);
}

/**
 * 更新任务时间信息
 */
export async function updateTaskTime(taskId, data) {
  return updateTaskTimeFields(taskId, data);
}

/**
 * 批量分配任务给成员
 */
export async function batchAssignTasks(taskIds, memberId, data = {}) {
  const results = {
    success: [],
    failed: []
  };

  for (const taskId of taskIds) {
    try {
      await assignTask(taskId, memberId, data);
      results.success.push(taskId);
    } catch (error) {
      results.failed.push({ taskId, error: error.message });
    }
  }

  return results;
}

/**
 * 获取任务分配概览
 */
export async function getAssignmentOverview(taskId) {
  const assignments = await getTaskAssignments(taskId);

  const overview = {
    assignees: [],
    reviewers: [],
    collaborators: [],
    totalEstimatedHours: 0,
    totalActualHours: 0
  };

  for (const assignment of assignments) {
    const memberInfo = assignment.member ? {
      id: assignment.member.id,
      name: assignment.member.name,
      avatar: assignment.member.avatar
    } : null;

    const assignmentInfo = {
      id: assignment.id,
      member: memberInfo,
      estimatedHours: assignment.estimatedHours,
      actualHours: assignment.actualHours
    };

    if (assignment.role === 'assignee') {
      overview.assignees.push(assignmentInfo);
    } else if (assignment.role === 'reviewer') {
      overview.reviewers.push(assignmentInfo);
    } else {
      overview.collaborators.push(assignmentInfo);
    }

    if (assignment.estimatedHours) {
      overview.totalEstimatedHours += assignment.estimatedHours;
    }
    if (assignment.actualHours) {
      overview.totalActualHours += assignment.actualHours;
    }
  }

  return overview;
}

export default {
  listTaskAssignments,
  listSubtaskAssignments,
  assignTask,
  assignSubtask,
  unassignTask,
  unassignSubtask,
  listMemberAssignments,
  getWorkload,
  updateTaskTime,
  batchAssignTasks,
  getAssignmentOverview
};
