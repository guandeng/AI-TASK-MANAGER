/** 需求状态 */
export type RequirementStatus = 'draft' | 'active' | 'completed' | 'archived';

/** 需求优先级 */
export type RequirementPriority = 'high' | 'medium' | 'low';

/** 文档接口 */
export interface RequirementDocument {
  id: number;
  name: string;
  path: string;
  size: number;
  mimeType?: string;
  description?: string;
  uploadedBy?: string;
  createdAt: string;
}

/** 需求接口 */
export interface Requirement {
  id: number;
  title: string;
  content: string;
  status: RequirementStatus;
  priority: RequirementPriority;
  tags?: string[];
  assignee?: string;
  createdAt: string;
  updatedAt: string;
  documents?: RequirementDocument[];
}

/** 需求统计 */
export interface RequirementStatistics {
  total: number;
  draft: number;
  active: number;
  completed: number;
  archived: number;
  highPriority: number;
  mediumPriority: number;
  lowPriority: number;
}

/** 需求列表查询参数 */
export interface RequirementListParams {
  status?: RequirementStatus;
  priority?: RequirementPriority;
  assignee?: string;
  keyword?: string;
  limit?: number;
}

/** 创建/更新需求参数 */
export interface RequirementFormData {
  title: string;
  content?: string;
  status?: RequirementStatus;
  priority?: RequirementPriority;
  tags?: string[];
  assignee?: string;
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
  status: string;
  priority: string;
  sortOrder: number;
  estimatedHours?: number;
  actualHours?: number;
  codeInterface?: string;
  acceptanceCriteria?: string;
  relatedFiles?: string;
  codeHints?: string;
  createdAt?: string;
  updatedAt?: string;
}

/** 任务依赖 */
export interface TaskDependency {
  id: number;
  taskId: number;
  dependsOnTaskId: number;
  createdAt?: string;
}

/** 任务带子任务（结构化数据） */
export interface TaskWithSubtasks {
  id: number;
  title: string;
  titleTrans?: string;
  description: string;
  descriptionTrans?: string;
  status: string;
  priority: string;
  category: string;
  details: string;
  detailsTrans?: string;
  testStrategy: string;
  testStrategyTrans?: string;
  module?: string;
  input?: string;
  output?: string;
  risk?: string;
  acceptanceCriteria?: string;
  assignee?: string;
  customFields?: string;
  isExpanding: boolean;
  startDate?: string;
  dueDate?: string;
  completedAt?: string;
  estimatedHours?: number;
  actualHours?: number;
  createdAt: string;
  updatedAt: string;
  subtasks: Subtask[];
  dependencies: TaskDependency[];
}

/** 需求树形结构（需求 + 任务 + 子任务） */
export interface RequirementTree {
  id: number;
  title: string;
  content: string;
  status: string;
  priority: string;
  tags?: string;
  assignee?: string;
  createdAt: string;
  updatedAt: string;
  documents: RequirementDocument[];
  tasks: TaskWithSubtasks[];
}
