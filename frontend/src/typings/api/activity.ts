/** 活动类型 */
export type ActivityAction =
  | 'task_created'
  | 'task_updated'
  | 'task_deleted'
  | 'task_status_changed'
  | 'task_priority_changed'
  | 'task_assigned'
  | 'task_unassigned'
  | 'subtask_created'
  | 'subtask_updated'
  | 'subtask_deleted'
  | 'subtask_status_changed'
  | 'subtask_completed'
  | 'subtask_assigned'
  | 'subtask_unassigned'
  | 'comment_added'
  | 'comment_updated'
  | 'comment_deleted'
  | 'time_estimated'
  | 'time_logged'
  | 'due_date_changed';

/** 活动日志 */
export interface Activity {
  id: number;
  taskId: number;
  subtaskId?: number;
  memberId?: number;
  member?: {
    id: number;
    name: string;
    email?: string;
    avatar?: string;
  };
  task?: {
    id: number;
    title: string;
  };
  action: ActivityAction;
  fieldName?: string;
  oldValue?: string;
  newValue?: string;
  description?: string;
  metadata?: Record<string, any>;
  createdAt: string;
}

/** 活动统计 */
export interface ActivityStatistics {
  byAction: Array<{
    action: string;
    count: number;
  }>;
  byDate: Array<{
    date: string;
    count: number;
  }>;
  byMember: Array<{
    memberId: number;
    memberName: string;
    count: number;
  }>;
  period: {
    startDate: string;
    endDate: string;
  };
  actionDescriptions?: Record<string, string>;
}

/** 活动查询参数 */
export interface ActivityQueryParams {
  memberId?: number;
  action?: string;
  startDate?: string;
  endDate?: string;
  limit?: number;
  offset?: number;
}

/** 活动类型标签 */
export const ACTIVITY_TYPE_LABELS: Record<ActivityAction, string> = {
  task_created: '创建任务',
  task_updated: '更新任务',
  task_deleted: '删除任务',
  task_status_changed: '更改状态',
  task_priority_changed: '更改优先级',
  task_assigned: '分配任务',
  task_unassigned: '取消分配',
  subtask_created: '创建子任务',
  subtask_updated: '更新子任务',
  subtask_deleted: '删除子任务',
  subtask_status_changed: '更改子任务状态',
  subtask_completed: '完成子任务',
  subtask_assigned: '分配子任务',
  subtask_unassigned: '取消子任务分配',
  comment_added: '添加评论',
  comment_updated: '更新评论',
  comment_deleted: '删除评论',
  time_estimated: '设置预估工时',
  time_logged: '记录工时',
  due_date_changed: '更改截止日期'
};

/** 活动类型图标 */
export const ACTIVITY_TYPE_ICONS: Record<ActivityAction, string> = {
  task_created: 'i-mdi:plus-circle',
  task_updated: 'i-mdi:pencil',
  task_deleted: 'i-mdi:delete',
  task_status_changed: 'i-mdi:swap-horizontal',
  task_priority_changed: 'i-mdi:flag',
  task_assigned: 'i-mdi:account-plus',
  task_unassigned: 'i-mdi:account-minus',
  subtask_created: 'i-mdi:plus-circle-outline',
  subtask_updated: 'i-mdi:pencil-outline',
  subtask_deleted: 'i-mdi:delete-outline',
  subtask_status_changed: 'i-mdi:swap-horizontal',
  subtask_completed: 'i-mdi:check-circle',
  subtask_assigned: 'i-mdi:account-plus-outline',
  subtask_unassigned: 'i-mdi:account-minus-outline',
  comment_added: 'i-mdi:comment-plus',
  comment_updated: 'i-mdi:comment-edit',
  comment_deleted: 'i-mdi:comment-remove',
  time_estimated: 'i-mdi:clock-outline',
  time_logged: 'i-mdi:clock-check',
  due_date_changed: 'i-mdi:calendar-clock'
};
