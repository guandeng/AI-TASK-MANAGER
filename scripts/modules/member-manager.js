/**
 * member-manager.js
 * 团队成员业务逻辑模块
 */

import {
  getMemberList,
  getMemberCount,
  getMemberById,
  createMember,
  updateMember,
  deleteMember,
  getMemberStatistics,
  getDepartmentList,
  getMembersByIds
} from './member-storage.js';

/**
 * 获取成员列表
 * @param {Object} filters - 筛选条件
 */
export async function listMembers(filters = {}) {
  return getMemberList(filters);
}

/**
 * 获取分页成员列表
 * @param {Object} options - 分页选项
 * @param {number} options.page - 页码
 * @param {number} options.pageSize - 每页数量
 * @param {Object} options.filters - 筛选条件
 */
export async function listMembersWithPaging(options = {}) {
  const { page = 1, pageSize = 20, filters = {} } = options;
  const offset = (page - 1) * pageSize;

  const [members, total] = await Promise.all([
    getMemberList({ ...filters, limit: pageSize, offset }),
    getMemberCount(filters)
  ]);

  return {
    members,
    total,
    page,
    pageSize,
    totalPages: Math.ceil(total / pageSize)
  };
}

/**
 * 获取成员详情
 * @param {number} id - 成员ID
 */
export async function getMember(id) {
  return getMemberById(id);
}

/**
 * 创建新成员
 * @param {Object} data - 成员数据
 */
export async function createNewMember(data) {
  // 验证角色
  const validRoles = ['admin', 'leader', 'member'];
  if (data.role && !validRoles.includes(data.role)) {
    throw new Error(`无效的角色: ${data.role}。有效角色: ${validRoles.join(', ')}`);
  }

  // 验证状态
  const validStatuses = ['active', 'inactive'];
  if (data.status && !validStatuses.includes(data.status)) {
    throw new Error(`无效的状态: ${data.status}。有效状态: ${validStatuses.join(', ')}`);
  }

  return createMember({
    name: data.name,
    email: data.email,
    avatar: data.avatar,
    role: data.role || 'member',
    department: data.department,
    skills: data.skills || [],
    status: data.status || 'active'
  });
}

/**
 * 更新成员信息
 * @param {number} id - 成员ID
 * @param {Object} data - 更新数据
 */
export async function updateMemberById(id, data) {
  // 验证角色
  const validRoles = ['admin', 'leader', 'member'];
  if (data.role && !validRoles.includes(data.role)) {
    throw new Error(`无效的角色: ${data.role}`);
  }

  // 验证状态
  const validStatuses = ['active', 'inactive'];
  if (data.status && !validStatuses.includes(data.status)) {
    throw new Error(`无效的状态: ${data.status}`);
  }

  return updateMember(id, data);
}

/**
 * 删除成员
 * @param {number} id - 成员ID
 */
export async function deleteMemberById(id) {
  const existing = await getMemberById(id);
  if (!existing) {
    throw new Error('成员不存在');
  }

  return deleteMember(id);
}

/**
 * 获取成员统计信息
 */
export async function getMemberStats() {
  return getMemberStatistics();
}

/**
 * 获取所有部门
 */
export async function getDepartments() {
  return getDepartmentList();
}

/**
 * 批量获取成员
 * @param {number[]} ids - 成员ID列表
 */
export async function getMembers(ids) {
  return getMembersByIds(ids);
}

/**
 * 搜索成员
 * @param {string} keyword - 搜索关键词
 * @param {number} limit - 返回数量限制
 */
export async function searchMembers(keyword, limit = 10) {
  if (!keyword || !keyword.trim()) {
    return [];
  }

  return getMemberList({ keyword: keyword.trim(), limit, status: 'active' });
}

/**
 * 停用成员
 * @param {number} id - 成员ID
 */
export async function deactivateMember(id) {
  return updateMember(id, { status: 'inactive' });
}

/**
 * 激活成员
 * @param {number} id - 成员ID
 */
export async function activateMember(id) {
  return updateMember(id, { status: 'active' });
}

/**
 * 更新成员头像
 * @param {number} id - 成员ID
 * @param {string} avatarUrl - 头像URL
 */
export async function updateMemberAvatar(id, avatarUrl) {
  return updateMember(id, { avatar: avatarUrl });
}

export default {
  listMembers,
  listMembersWithPaging,
  getMember,
  createNewMember,
  updateMemberById,
  deleteMemberById,
  getMemberStats,
  getDepartments,
  getMembers,
  searchMembers,
  deactivateMember,
  activateMember,
  updateMemberAvatar
};
