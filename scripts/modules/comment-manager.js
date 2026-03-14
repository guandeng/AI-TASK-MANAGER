/**
 * comment-manager.js
 * 评论业务逻辑模块
 */

import {
  getTaskComments,
  getCommentById,
  createComment,
  updateComment,
  deleteComment,
  getCommentReplies,
  getCommentStatistics
} from './comment-storage.js';
import { getMemberById } from './member-storage.js';

/**
 * 获取任务评论列表
 */
export async function listComments(taskId, options = {}) {
  return getTaskComments(taskId, options);
}

/**
 * 获取评论详情
 */
export async function getComment(id) {
  return getCommentById(id);
}

/**
 * 添加评论
 */
export async function addComment(data) {
  // 验证评论人
  const member = await getMemberById(data.memberId);
  if (!member) {
    throw new Error('评论人不存在');
  }

  // 如果有父评论，验证父评论存在
  if (data.parentId) {
    const parent = await getCommentById(data.parentId);
    if (!parent) {
      throw new Error('父评论不存在');
    }
  }

  return createComment(data);
}

/**
 * 更新评论
 */
export async function updateCommentById(id, data, memberId) {
  const existing = await getCommentById(id);
  if (!existing) {
    throw new Error('评论不存在');
  }

  // 验证是否是评论作者
  if (memberId && existing.memberId !== memberId) {
    throw new Error('只能修改自己的评论');
  }

  return updateComment(id, data);
}

/**
 * 删除评论
 */
export async function deleteCommentById(id, memberId) {
  const existing = await getCommentById(id);
  if (!existing) {
    throw new Error('评论不存在');
  }

  // 验证是否是评论作者
  if (memberId && existing.memberId !== memberId) {
    throw new Error('只能删除自己的评论');
  }

  return deleteComment(id);
}

/**
 * 获取评论回复
 */
export async function listReplies(parentId) {
  return getCommentReplies(parentId);
}

/**
 * 获取评论统计
 */
export async function getStats(taskId) {
  return getCommentStatistics(taskId);
}

/**
 * 获取评论树形结构
 */
export async function getCommentTree(taskId, options = {}) {
  const comments = await getTaskComments(taskId, options);

  // 构建树形结构
  const commentMap = new Map();
  const rootComments = [];

  // 先创建所有评论的映射
  for (const comment of comments) {
    comment.replies = [];
    commentMap.set(comment.id, comment);
  }

  // 构建树形结构
  for (const comment of comments) {
    if (comment.parentId) {
      const parent = commentMap.get(comment.parentId);
      if (parent) {
        parent.replies.push(comment);
      }
    } else {
      rootComments.push(comment);
    }
  }

  return rootComments;
}

/**
 * 提取评论中@的成员
 */
export function extractMentions(content) {
  const mentionRegex = /@(\S+)/g;
  const mentions = [];
  let match;

  while ((match = mentionRegex.exec(content)) !== null) {
    mentions.push(match[1]);
  }

  return [...new Set(mentions)];
}

export default {
  listComments,
  getComment,
  addComment,
  updateCommentById,
  deleteCommentById,
  listReplies,
  getStats,
  getCommentTree,
  extractMentions
};
