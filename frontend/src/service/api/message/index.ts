import { request } from '@/service/request';
import type { MessageListParams, Message } from '@/typings/api/message';

const API_BASE = '/api';

/** 获取消息列表 */
export function fetchMessages(params?: MessageListParams) {
  return request<{ messages: Message[]; total: number }>({
    url: `${API_BASE}/messages`,
    method: 'GET',
    params
  });
}

/** 获取单个消息 */
export function fetchMessage(id: number) {
  return request<Message>({
    url: `${API_BASE}/messages/${id}`,
    method: 'GET'
  });
}

/** 标记消息已读 */
export function markMessageRead(id: number) {
  return request<void>({
    url: `${API_BASE}/messages/${id}/read`,
    method: 'PUT'
  });
}

/** 删除消息 */
export function deleteMessage(id: number) {
  return request<void>({
    url: `${API_BASE}/messages/${id}`,
    method: 'DELETE'
  });
}

/** 获取未读消息数量 */
export function fetchUnreadCount() {
  return request<{ count: number }>({
    url: `${API_BASE}/messages/unread-count`,
    method: 'GET'
  });
}
