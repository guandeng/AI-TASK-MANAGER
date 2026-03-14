/** 分配角色 */
export type AssignmentRole = 'assignee' | 'reviewer' | 'collaborator';

/** 分配信息 */
export interface Assignment {
  id: number;
  taskId?: number;
  subtaskId?: number;
  memberId: number;
  member?: {
    id: number;
    name: string;
    email?: string;
    avatar?: string;
    role?: string;
  };
  role: AssignmentRole;
  assignedBy?: number;
  estimatedHours?: number;
  actualHours?: number;
  createdAt: string;
  updatedAt: string;
}

/** 任务分配概览 */
export interface AssignmentOverview {
  assignees: Array<{
    id: number;
    member: {
      id: number;
      name: string;
      avatar?: string;
    } | null;
    estimatedHours?: number;
    actualHours?: number;
  }>;
  reviewers: Array<{
    id: number;
    member: {
      id: number;
      name: string;
      avatar?: string;
    } | null;
    estimatedHours?: number;
    actualHours?: number;
  }>;
  collaborators: Array<{
    id: number;
    member: {
      id: number;
      name: string;
      avatar?: string;
    } | null;
    estimatedHours?: number;
    actualHours?: number;
  }>;
  totalEstimatedHours: number;
  totalActualHours: number;
}

/** 成员任务分配 */
export interface MemberAssignment {
  id: number;
  taskId: number;
  memberId: number;
  role: AssignmentRole;
  taskTitle: string;
  taskStatus: string;
  taskPriority: string;
  estimatedHours?: number;
  actualHours?: number;
  createdAt: string;
}

/** 成员工作量统计 */
export interface MemberWorkload {
  totalTasks: number;
  completedTasks: number;
  inProgressTasks: number;
  pendingTasks: number;
  totalSubtasks: number;
  totalEstimatedHours: number;
  totalActualHours: number;
}

/** 任务时间信息 */
export interface TaskTimeInfo {
  startDate?: string;
  dueDate?: string;
  completedAt?: string;
  estimatedHours?: number;
  actualHours?: number;
}

/** 创建分配请求 */
export interface CreateAssignmentRequest {
  memberId: number;
  role?: AssignmentRole;
  assignedBy?: number;
  estimatedHours?: number;
  actualHours?: number;
}

/** 分配角色选项 */
export const ASSIGNMENT_ROLE_OPTIONS = [
  { label: '负责人', value: 'assignee' },
  { label: '审核人', value: 'reviewer' },
  { label: '协作者', value: 'collaborator' }
] as const;

/** 角色显示名称映射 */
export const ASSIGNMENT_ROLE_LABELS: Record<AssignmentRole, string> = {
  assignee: '负责人',
  reviewer: '审核人',
  collaborator: '协作者'
};

/** 角色颜色映射 */
export const ASSIGNMENT_ROLE_COLORS: Record<AssignmentRole, string> = {
  assignee: 'success',
  reviewer: 'warning',
  collaborator: 'info'
};
