import { request } from '@/service/request';

/**
 * 任务复杂度分析结果
 */
export interface ComplexityAnalysis {
  taskId: number;
  taskTitle: string;
  complexityScore: number;
  complexityLevel: 'low' | 'medium' | 'high';
  reasoning: string;
  subtaskCount: number;
  timeEstimate: string;
  dependencies: number[];
  riskFactors: string[];
}

/**
 * 复杂度汇总
 */
export interface ComplexitySummary {
  totalTasks: number;
  lowComplexity: number;
  mediumComplexity: number;
  highComplexity: number;
  averageScore: number;
}

/**
 * 复杂度报告数据
 */
export interface ComplexityReportData {
  analyses: ComplexityAnalysis[];
  summary: ComplexitySummary;
  generatedAt: string;
}

/**
 * 复杂度报告
 */
export interface ComplexityReport {
  id: number;
  requirementId?: number;
  status: string;
  reportData: string;
  errorMessage?: string;
  createdAt: string;
  updatedAt: string;
}

/**
 * 分析单个任务复杂度
 */
export function analyzeTask(taskId: number) {
  return request<ComplexityAnalysis>({
    url: `/api/tasks/${taskId}/analyze`,
    method: 'POST'
  });
}

/**
 * 分析需求复杂度
 */
export function analyzeRequirement(
  requirementId: number,
  data?: { knowledgePaths?: string[]; useKnowledge?: boolean }
) {
  return request<{ reportId: number; analysis: ComplexityReportData }>({
    url: `/api/requirements/${requirementId}/analyze`,
    method: 'POST',
    data
  });
}

/**
 * 异步分析需求复杂度
 */
export function analyzeRequirementAsync(
  requirementId: number,
  data?: { knowledgePaths?: string[]; useKnowledge?: boolean }
) {
  return request<{ messageId: number; message: string }>({
    url: `/api/requirements/${requirementId}/analyze-async`,
    method: 'POST',
    data
  });
}

/**
 * 获取复杂度报告
 */
export function getComplexityReport(reportId: number) {
  return request<ComplexityReport>({
    url: `/api/complexity/reports/${reportId}`,
    method: 'GET'
  });
}

/**
 * 获取需求的所有复杂度报告
 */
export function getRequirementReports(requirementId: number) {
  return request<ComplexityReport[]>({
    url: `/api/requirements/${requirementId}/complexity-reports`,
    method: 'GET'
  });
}

/**
 * 依赖关系图节点
 */
export interface DependencyNode {
  id: number;
  label: string;
  status: string;
}

/**
 * 依赖关系图边
 */
export interface DependencyEdge {
  source: number;
  target: number;
}

/**
 * 依赖关系图
 */
export interface DependencyGraph {
  nodes: DependencyNode[];
  edges: DependencyEdge[];
}

/**
 * 获取依赖关系图
 */
export function getDependencyGraph(requirementId?: number) {
  const params = requirementId ? { requirementId } : {};
  return request<DependencyGraph>({
    url: '/api/tasks/dependencies/graph',
    method: 'GET',
    params
  });
}

/**
 * 修复无效的依赖关系
 */
export function fixDependencies() {
  return request<{ fixed: number; removed: number; removedIds: number[] }>({
    url: '/api/tasks/dependencies/fix',
    method: 'POST'
  });
}

/**
 * 获取接下来可执行的任务
 */
export function getNextTasks(requirementId?: number, limit?: number) {
  const params: Record<string, unknown> = {};
  if (requirementId) params.requirementId = requirementId;
  if (limit) params.limit = limit;
  return request<any[]>({
    url: '/api/tasks/next',
    method: 'GET',
    params
  });
}

/**
 * 带知识库展开任务
 */
export function expandTaskWithKnowledge(
  taskId: number,
  data: { knowledgePaths?: string[]; additionalContext?: string }
) {
  return request<any>({
    url: `/api/tasks/${taskId}/expand-with-knowledge`,
    method: 'POST',
    data
  });
}

/**
 * 带研究功能展开任务
 */
export function expandTaskWithResearch(taskId: number) {
  return request<any>({
    url: `/api/tasks/${taskId}/expand-with-research`,
    method: 'POST'
  });
}

/**
 * 知识库摘要
 */
export interface KnowledgeSummary {
  enabled: boolean;
  paths: string[];
  maxSize: number;
  maxFiles: number;
  fileTypes: string[];
  customPaths?: string[];
}

/**
 * 获取知识库摘要
 */
export function getKnowledgeSummary(paths?: string[]) {
  const params = paths ? { paths } : {};
  return request<KnowledgeSummary>({
    url: '/api/knowledge/summary',
    method: 'GET',
    params
  });
}

/**
 * 加载知识库
 */
export function loadKnowledge(data: { paths?: string[]; additionalContext?: string }) {
  return request<{ status: string; paths: string[]; hasContext: boolean }>({
    url: '/api/knowledge/load',
    method: 'POST',
    data
  });
}
