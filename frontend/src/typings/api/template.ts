/** 字段定义 */
export interface FieldDefinition {
  name: string;
  label: string;
  type: 'text' | 'textarea' | 'number' | 'select' | 'array' | 'boolean';
  options?: string[];
  required?: boolean;
}

/** 字段模式 */
export interface FieldSchema {
  fields: FieldDefinition[];
}

/** 项目模板 */
export interface ProjectTemplate {
  id: number;
  name: string;
  description?: string;
  category?: string;
  isPublic: boolean;
  createdBy?: number;
  creator?: {
    id: number;
    name: string;
  };
  usageCount: number;
  tags?: string[];
  fieldSchema?: FieldSchema;
  createdAt: string;
  updatedAt: string;
  tasks?: TemplateTask[];
}

/** 模板任务 */
export interface TemplateTask {
  id: number;
  templateId: number;
  title: string;
  description?: string;
  priority: 'low' | 'medium' | 'high' | 'urgent';
  order: number;
  estimatedHours?: number;
  dependencies?: number[];
  subtasks?: TemplateSubtask[];
  createdAt: string;
}

/** 模板子任务 */
export interface TemplateSubtask {
  id: number;
  templateTaskId: number;
  title: string;
  description?: string;
  order: number;
  estimatedHours?: number;
  createdAt: string;
}

/** 独立任务模板 */
export interface TaskTemplate {
  id: number;
  name: string;
  description?: string;
  title: string;
  taskDescription?: string;
  priority: 'low' | 'medium' | 'high' | 'urgent';
  estimatedHours?: number;
  subtasks?: string[];
  tags?: string[];
  isPublic: boolean;
  createdBy?: number;
  creator?: {
    id: number;
    name: string;
  };
  usageCount: number;
  createdAt: string;
  updatedAt: string;
}

/** 创建项目模板请求 */
export interface CreateProjectTemplateRequest {
  name: string;
  description?: string;
  category?: string;
  isPublic?: boolean;
  tags?: string[];
  fieldSchema?: FieldSchema;
  tasks?: CreateTemplateTaskRequest[];
}

/** 创建模板任务请求 */
export interface CreateTemplateTaskRequest {
  title: string;
  description?: string;
  priority?: 'low' | 'medium' | 'high' | 'urgent';
  order?: number;
  estimatedHours?: number;
  dependencies?: number[];
  subtasks?: CreateTemplateSubtaskRequest[];
}

/** 创建模板子任务请求 */
export interface CreateTemplateSubtaskRequest {
  title: string;
  description?: string;
  order?: number;
  estimatedHours?: number;
}

/** 创建独立任务模板请求 */
export interface CreateTaskTemplateRequest {
  name: string;
  description?: string;
  title: string;
  taskDescription?: string;
  priority?: 'low' | 'medium' | 'high' | 'urgent';
  estimatedHours?: number;
  subtasks?: string[];
  tags?: string[];
  isPublic?: boolean;
}

/** 实例化模板请求 */
export interface InstantiateTemplateRequest {
  name?: string;
  description?: string;
  startDate?: string;
  dueDate?: string;
}

/** 模板评分结果 */
export interface TemplateScoreResult {
  scores: {
    clarity: number;
    completeness: number;
    structure: number;
    actionability: number;
    consistency: number;
  };
  totalScore: number;
  strengths: string[];
  weaknesses: string[];
  suggestions: Array<{
    issue: string;
    suggestion: string;
  }>;
  analysis: string;
}

/** 模板分类选项 */
export const TEMPLATE_CATEGORY_OPTIONS = [
  { label: '产品开发', value: 'product' },
  { label: '市场推广', value: 'marketing' },
  { label: '运营活动', value: 'operation' },
  { label: '技术项目', value: 'technical' },
  { label: '其他', value: 'other' }
];

/** 优先级选项 */
export const TEMPLATE_PRIORITY_OPTIONS = [
  { label: '低', value: 'low' },
  { label: '中', value: 'medium' },
  { label: '高', value: 'high' },
  { label: '紧急', value: 'urgent' }
];

/** 优先级标签 */
export const TEMPLATE_PRIORITY_LABELS: Record<string, string> = {
  low: '低',
  medium: '中',
  high: '高',
  urgent: '紧急'
};

/** 优先级颜色 */
export const TEMPLATE_PRIORITY_COLORS: Record<string, 'default' | 'success' | 'warning' | 'error' | 'info'> = {
  low: 'default',
  medium: 'info',
  high: 'warning',
  urgent: 'error'
};
