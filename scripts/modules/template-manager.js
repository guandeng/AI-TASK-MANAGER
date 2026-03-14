/**
 * template-manager.js
 * 模板业务逻辑模块
 */

import {
  getProjectTemplateList,
  getProjectTemplateById,
  createProjectTemplate,
  updateProjectTemplate,
  deleteProjectTemplate,
  incrementTemplateUsage,
  getTaskTemplateList,
  getTaskTemplateById,
  createTaskTemplate,
  updateTaskTemplate,
  deleteTaskTemplate,
  incrementTaskTemplateUsage,
  getTemplateCategories
} from './template-storage.js';

// ============================================
// 项目模板相关
// ============================================

/**
 * 获取项目模板列表
 */
export async function listProjectTemplates(filters = {}) {
  return getProjectTemplateList(filters);
}

/**
 * 获取项目模板详情
 */
export async function getProjectTemplate(id) {
  return getProjectTemplateById(id);
}

/**
 * 创建项目模板
 */
export async function createNewProjectTemplate(data) {
  // 验证分类
  const validCategories = ['web', 'mobile', 'api', 'cli', 'other'];
  if (data.category && !validCategories.includes(data.category)) {
    data.category = 'other';
  }

  return createProjectTemplate(data);
}

/**
 * 更新项目模板
 */
export async function updateProjectTemplateById(id, data) {
  return updateProjectTemplate(id, data);
}

/**
 * 删除项目模板
 */
export async function deleteProjectTemplateById(id) {
  return deleteProjectTemplate(id);
}

/**
 * 从模板创建项目任务
 */
export async function instantiateProjectTemplate(templateId, customizations = {}) {
  const template = await getProjectTemplateById(templateId);
  if (!template) {
    throw new Error('模板不存在');
  }

  // 增加使用次数
  await incrementTemplateUsage(templateId);

  // 返回实例化后的任务数据
  const tasks = template.tasks.map((task, index) => ({
    title: customizations.taskPrefix ? `${customizations.taskPrefix} ${task.title}` : task.title,
    description: task.description,
    details: task.details,
    testStrategy: task.testStrategy,
    priority: task.priority,
    estimatedHours: task.estimatedHours,
    dependencies: task.dependencies.map(d => d + (customizations.taskIdOffset || 0)),
    subtasks: task.subtasks.map(st => ({
      title: st.title,
      description: st.description,
      details: st.details,
      dependencies: st.dependencies
    }))
  }));

  return {
    templateName: template.name,
    tasks
  };
}

// ============================================
// 任务模板相关
// ============================================

/**
 * 获取任务模板列表
 */
export async function listTaskTemplates(filters = {}) {
  return getTaskTemplateList(filters);
}

/**
 * 获取任务模板详情
 */
export async function getTaskTemplate(id) {
  return getTaskTemplateById(id);
}

/**
 * 创建任务模板
 */
export async function createNewTaskTemplate(data) {
  const validCategories = ['frontend', 'backend', 'testing', 'devops', 'documentation', 'other'];
  if (data.category && !validCategories.includes(data.category)) {
    data.category = 'other';
  }

  return createTaskTemplate(data);
}

/**
 * 更新任务模板
 */
export async function updateTaskTemplateById(id, data) {
  return updateTaskTemplate(id, data);
}

/**
 * 删除任务模板
 */
export async function deleteTaskTemplateById(id) {
  return deleteTaskTemplate(id);
}

/**
 * 从模板创建任务
 */
export async function instantiateTaskTemplate(templateId, customizations = {}) {
  const template = await getTaskTemplateById(templateId);
  if (!template) {
    throw new Error('模板不存在');
  }

  // 增加使用次数
  await incrementTaskTemplateUsage(templateId);

  return {
    title: customizations.title || template.title,
    description: customizations.description || template.description,
    details: customizations.details || template.details,
    testStrategy: customizations.testStrategy || template.testStrategy,
    priority: customizations.priority || template.priority,
    templateId: Number(templateId),
    templateName: template.name
  };
}

// ============================================
// 通用模板相关
// ============================================

/**
 * 获取所有模板分类
 */
export async function listTemplateCategories() {
  return getTemplateCategories();
}

/**
 * 搜索模板
 */
export async function searchTemplates(keyword, type = 'all') {
  const filters = { keyword };

  const results = {
    projectTemplates: [],
    taskTemplates: []
  };

  if (type === 'all' || type === 'project') {
    results.projectTemplates = await getProjectTemplateList(filters);
  }

  if (type === 'all' || type === 'task') {
    results.taskTemplates = await getTaskTemplateList(filters);
  }

  return results;
}

export default {
  listProjectTemplates,
  getProjectTemplate,
  createNewProjectTemplate,
  updateProjectTemplateById,
  deleteProjectTemplateById,
  instantiateProjectTemplate,
  listTaskTemplates,
  getTaskTemplate,
  createNewTaskTemplate,
  updateTaskTemplateById,
  deleteTaskTemplateById,
  instantiateTaskTemplate,
  listTemplateCategories,
  searchTemplates
};
