/** 任务状态 */
export type TaskStatus = 'pending' | 'done' | 'deferred' | 'in-progress';

/** 任务优先级 */
export type TaskPriority = 'high' | 'medium' | 'low';

/** 代码接口定义 */
export interface CodeInterface {
  name: string;
  inputs?: string;
  outputs?: string;
  example?: string;
}

/** 验收标准 */
export interface AcceptanceCriteria {
  id: number;
  description: string;
  completed: boolean;
}

/** 子任务 */
export interface Subtask {
  id: number;
  taskId: number;
  title: string;
  titleTrans?: string;
  description?: string;
  descriptionTrans?: string;
  details?: string;
  detailsTrans?: string;
  status: TaskStatus;
  priority: TaskPriority;
  sortOrder: number;
  estimatedHours?: number;
  actualHours?: number;
  codeInterface?: CodeInterface;
  acceptanceCriteria?: AcceptanceCriteria[];
  relatedFiles?: string[];
  codeHints?: string;
  createdAt?: string;
  updatedAt?: string;
  dependencies?: number[];
}

/** 任务 */
export interface Task {
  id: number;
  requirementTitle?: string;
  requirementId?: number;
  title: string;
  titleTrans?: string;
  description: string;
  descriptionTrans?: string;
  status: TaskStatus;
  priority: TaskPriority;
  dependencies: number[];
  details?: string;
  detailsTrans?: string;
  testStrategy?: string;
  testStrategyTrans?: string;
  subtasks?: Subtask[];
  /** 子任务总数（列表 API 返回） */
  subtaskCount?: number;
  /** 已完成子任务数（列表 API 返回） */
  subtaskDoneCount?: number;
  assignee?: string; // 负责人
  // 时间管理字段
  startDate?: string;
  dueDate?: string;
  completedAt?: string;
  estimatedHours?: number;
  actualHours?: number;
  // AI 扩展状态
  isExpanding?: boolean;
  expandMessageId?: number;
  // 扩展字段
  tags?: string[];
  createdAt?: string;
  updatedAt?: string;
}

/** 任务列表响应 */
export interface TaskListResponse {
  list: Task[];
  total: number;
  page: number;
  pageSize: number;
}

export interface TaskListParams {
  page?: number;
  pageSize?: number;
  requirementId?: number;
  status?: TaskStatus;
  priority?: TaskPriority;
  assignee?: string;
  keyword?: string;
}

/** 任务创建请求 */
export interface TaskCreateRequest {
  title: string;
  description?: string;
  details?: string;
  testStrategy?: string;
  priority?: TaskPriority;
  assignee?: string;
  requirementId?: number;
  dependencies?: number[];
  startDate?: string;
  dueDate?: string;
  estimatedHours?: number;
}

/** 任务更新请求 */
export interface TaskUpdateRequest {
  status?: TaskStatus;
  title?: string;
  description?: string;
  details?: string;
  testStrategy?: string;
  priority?: TaskPriority;
  assignee?: string;
  startDate?: string;
  dueDate?: string;
  estimatedHours?: number;
  actualHours?: number;
  completedAt?: string;
  tags?: string[];
}

/** 子任务更新请求 */
export interface SubtaskUpdateRequest {
  status?: TaskStatus;
  title?: string;
  description?: string;
  details?: string;
  priority?: TaskPriority;
  estimatedHours?: number;
  actualHours?: number;
  codeInterface?: CodeInterface;
  acceptanceCriteria?: AcceptanceCriteria[];
  relatedFiles?: string[];
  codeHints?: string;
}

/** 任务统计 */
export interface TaskStatistics {
  total: number;
  done: number;
  pending: number;
  deferred: number;
  inProgress: number;
  highPriority: number;
  mediumPriority: number;
  lowPriority: number;
}
