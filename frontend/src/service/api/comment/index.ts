import { request } from '@/service/request';
import type { Comment, CommentStatistics, CreateCommentRequest, UpdateCommentRequest } from '@/typings/api/comment';

const API_BASE = '/api';

/**
 * 获取任务评论列表
 */
export function fetchTaskComments(taskId: number, options?: { subtaskId?: number; limit?: number }) {
  return request<Comment[]>({
    url: `${API_BASE}/tasks/${taskId}/comments`,
    method: 'GET',
    params: options
  });
}

/**
 * 获取评论树形结构
 */
export function fetchCommentTree(taskId: number) {
  return request<Comment[]>({
    url: `${API_BASE}/tasks/${taskId}/comments/tree`,
    method: 'GET'
  });
}

/**
 * 获取评论统计
 */
export function fetchCommentStatistics(taskId: number) {
  return request<CommentStatistics>({
    url: `${API_BASE}/tasks/${taskId}/comments/statistics`,
    method: 'GET'
  });
}

/**
 * 获取单个评论
 */
export function fetchComment(taskId: number, commentId: number) {
  return request<Comment>({
    url: `${API_BASE}/tasks/${taskId}/comments/${commentId}`,
    method: 'GET'
  });
}

/**
 * 添加评论
 */
export function createComment(taskId: number, data: CreateCommentRequest) {
  return request<Comment>({
    url: `${API_BASE}/tasks/${taskId}/comments`,
    method: 'POST',
    data
  });
}

/**
 * 更新评论
 */
export function updateComment(taskId: number, commentId: number, data: UpdateCommentRequest) {
  return request<Comment>({
    url: `${API_BASE}/tasks/${taskId}/comments/${commentId}/update`,
    method: 'POST',
    data
  });
}

/**
 * 删除评论
 */
export function deleteComment(taskId: number, commentId: number, memberId: number) {
  return request<{ success: boolean; message: string }>({
    url: `${API_BASE}/tasks/${taskId}/comments/${commentId}/delete`,
    method: 'POST',
    data: { memberId }
  });
}

/**
 * 获取评论回复列表
 */
export function fetchCommentReplies(taskId: number, commentId: number) {
  return request<Comment[]>({
    url: `${API_BASE}/tasks/${taskId}/comments/${commentId}/replies`,
    method: 'GET'
  });
}
