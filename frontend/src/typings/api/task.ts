/** 任务状态 */
export type TaskStatus = 'pending' | 'done' | 'deferred' | 'in-progress';

/** 任务优先级 */
export type TaskPriority = 'high' | 'medium' | 'low';

/** 子任务 */
export interface Subtask {
  id: number;
  title: string;
  titleTrans?: string;
  description?: string;
  descriptionTrans?: string;
  details?: string;
  detailsTrans?: string;
  status: TaskStatus;
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
  assignee?: string; // 负责人
}

/** 任务列表响应 */
export interface TaskListResponse {
  projectName: string;
  projectVersion: string;
  tasks: Task[];
}

export interface TaskListParams {
  requirementId?: number;
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
}

/** 子任务更新请求 */
export interface SubtaskUpdateRequest {
  status?: TaskStatus;
  title?: string;
  description?: string;
  details?: string;
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
