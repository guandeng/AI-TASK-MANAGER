/** 评论信息 */
export interface Comment {
  id: number;
  taskId: number;
  subtaskId?: number;
  memberId: number;
  member?: {
    id: number;
    name: string;
    email?: string;
    avatar?: string;
  };
  parentId?: number;
  content: string;
  mentions: number[];
  replies?: Comment[];
  replyCount?: number;
  createdAt: string;
  updatedAt: string;
}

/** 评论统计 */
export interface CommentStatistics {
  total: number;
  uniqueAuthors: number;
}

/** 创建评论请求 */
export interface CreateCommentRequest {
  memberId: number;
  subtaskId?: number;
  parentId?: number;
  content: string;
  mentions?: number[];
}

/** 更新评论请求 */
export interface UpdateCommentRequest {
  memberId?: number;
  content?: string;
  mentions?: number[];
}
