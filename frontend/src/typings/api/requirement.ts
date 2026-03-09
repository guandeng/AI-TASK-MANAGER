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
