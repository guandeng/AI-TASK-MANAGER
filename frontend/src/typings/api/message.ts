/** 消息类型 */
export type MessageType = 'expand_task' | 'regenerate_subtask';

/** 消息状态 */
export type MessageStatus = 'pending' | 'processing' | 'success' | 'failed';

/** 消息 */
export interface Message {
  id: number;
  taskId: number;
  type: MessageType;
  status: MessageStatus;
  title: string;
  content?: string;
  errorMessage?: string;
  resultSummary?: string;
  isRead: boolean;
  createdAt: string;
  updatedAt: string;
}

/** 消息列表响应 */
export interface MessageListResponse {
  messages: Message[];
  total: number;
}

/** 消息列表参数 */
export interface MessageListParams {
  taskId?: number;
  status?: MessageStatus;
  unreadOnly?: boolean;
  limit?: number;
  offset?: number;
}
