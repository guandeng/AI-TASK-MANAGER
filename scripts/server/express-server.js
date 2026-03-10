/**
 * web-server.js
 * Express server for task visualization web interface
 */

import express from 'express';
import path from 'path';
import cors from 'cors';
import { fileURLToPath } from 'url';
import fs from 'fs';
import { log, CONFIG, findTaskById } from '../modules/utils.js';
import { readTaskData, writeTaskData, getTaskStorageMode } from '../modules/task-storage.js';
import { expandTask } from '../modules/task-manager.js';
import { generateSubtasks, parseSubtasksFromText, callGemini } from '../modules/ai-services.js';
import { loadConfig, saveConfig, updateConfig, getConfigPath, resetConfig } from '../modules/config-loader.js';
import {
  listRequirements,
  getRequirement,
  createNewRequirement,
  updateRequirementById,
  deleteRequirementById,
  uploadDocument,
  removeDocument,
  getStatistics,
  UPLOAD_DIR
} from '../modules/requirement-manager.js';
import { startLoadingIndicator, stopLoadingIndicator } from '../modules/ui.js';
import multer from 'multer';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

function removeDeletedTaskReferences(tasks, deletedTaskId) {
  const deletedTaskIdStr = String(deletedTaskId);
  const deletedTaskPrefix = `${deletedTaskIdStr}.`;

  tasks.forEach(task => {
    if (Array.isArray(task.dependencies)) {
      task.dependencies = task.dependencies.filter(dep => String(dep) !== deletedTaskIdStr);
    }

    if (!Array.isArray(task.subtasks)) {
      return;
    }

    task.subtasks = task.subtasks.map(subtask => {
      if (!Array.isArray(subtask.dependencies)) {
        return subtask;
      }

      return {
        ...subtask,
        dependencies: subtask.dependencies.filter(dep => {
          const depStr = String(dep);
          return depStr !== deletedTaskIdStr && !depStr.startsWith(deletedTaskPrefix);
        })
      };
    });
  });
}

/**
 * Start the web server
 * @param {Object} options Server configuration options
 * @param {number} options.port Port to run the server on
 * @param {string} options.tasksPath Path to the tasks.json file
 * @returns {Promise<Object>} Server instance
 */
export async function startWebServer(options = {}) {
  const port = options.port || 3002;
  const tasksPath = options.tasksPath || 'tasks/tasks.json';

  const app = express();

  // Enable CORS
  app.use(cors());

  // JSON parsing middleware
  app.use(express.json());

  // API endpoints
  app.get('/api/tasks', async (req, res) => {
    try {
      const requirementId = parseInt(req.query.requirementId, 10);
      const data = await readTaskData(tasksPath);
      if (!data) {
        return res.status(404).json({ error: 'Tasks file not found' });
      }

      const requirements = await listRequirements();
      const requirementTitleMap = new Map(
        requirements.map(requirement => [Number(requirement.id), requirement.title])
      );

      // 采用中文优先策略：克隆数据以避免修改原始数据
      const localizedData = JSON.parse(JSON.stringify(data));

      // 处理每个任务，优先使用中文翻译字段（如果存在）
      localizedData.tasks = localizedData.tasks.map(task => {
        if (task.requirementId && requirementTitleMap.has(Number(task.requirementId))) {
          task.requirementTitle = requirementTitleMap.get(Number(task.requirementId));
        }

        // 优先使用中文标题
        if (task.titleTrans) {
          task.title = task.titleTrans;
        }

        // 优先使用中文描述
        if (task.descriptionTrans) {
          task.description = task.descriptionTrans;
        }

        // 优先使用中文详情
        if (task.detailsTrans) {
          task.details = task.detailsTrans;
        }

        // 优先使用中文测试策略
        if (task.testStrategyTrans) {
          task.testStrategy = task.testStrategyTrans;
        }

        // 处理子任务的翻译字段
        if (task.subtasks && task.subtasks.length > 0) {
          task.subtasks = task.subtasks.map(subtask => {
            // 优先使用中文标题
            if (subtask.titleTrans) {
              subtask.title = subtask.titleTrans;
            }

            // 优先使用中文描述
            if (subtask.descriptionTrans) {
              subtask.description = subtask.descriptionTrans;
            }

            // 优先使用中文详情
            if (subtask.detailsTrans) {
              subtask.details = subtask.detailsTrans;
            }

            // 优先使用中文测试策略
            if (subtask.testStrategyTrans) {
              subtask.testStrategy = subtask.testStrategyTrans;
            }

            return subtask;
          });
        }

        return task;
      });

      if (!Number.isNaN(requirementId)) {
        localizedData.tasks = localizedData.tasks.filter(task => task.requirementId === requirementId);
      }

      res.json(localizedData);
    } catch (error) {
      log('error', `Error reading tasks: ${error.message}`);
      res.status(500).json({ error: 'Failed to read tasks data' });
    }
  });

  // 获取单个任务的API
  app.get('/api/tasks/:taskId', async (req, res) => {
    try {
      const taskId = parseInt(req.params.taskId, 10);
      const locale = req.query.locale || 'zh'; // 默认使用中文

      const data = await readTaskData(tasksPath);
      if (!data || !data.tasks) {
        return res.status(404).json({ error: 'Tasks file not found' });
      }

      // 查找任务
      const task = data.tasks.find(t => t.id === taskId);
      if (!task) {
        return res.status(404).json({ error: 'Task not found' });
      }

      if (task.requirementId) {
        const requirement = await getRequirement(task.requirementId);
        if (requirement?.title) {
          task.requirementTitle = requirement.title;
        }
      }

      // 克隆任务以避免修改原始数据
      const taskCopy = JSON.parse(JSON.stringify(task));

      // 如果请求英文内容，则返回原始字段
      if (locale === 'en') {
        return res.json(taskCopy);
      }

      // 如果请求中文内容，则进行本地化处理
      const localizedTask = { ...taskCopy };

      // 优先使用中文标题
      if (localizedTask.titleTrans) {
        localizedTask.title = localizedTask.titleTrans;
      }

      // 优先使用中文描述
      if (localizedTask.descriptionTrans) {
        localizedTask.description = localizedTask.descriptionTrans;
      }

      // 优先使用中文详情
      if (localizedTask.detailsTrans) {
        localizedTask.details = localizedTask.detailsTrans;
      }

      // 优先使用中文测试策略
      if (localizedTask.testStrategyTrans) {
        localizedTask.testStrategy = localizedTask.testStrategyTrans;
      }

      // 处理子任务的翻译字段
      if (localizedTask.subtasks && localizedTask.subtasks.length > 0) {
        localizedTask.subtasks = localizedTask.subtasks.map(subtask => {
          const subtaskCopy = { ...subtask };

          // 优先使用中文标题
          if (subtaskCopy.titleTrans) {
            subtaskCopy.title = subtaskCopy.titleTrans;
          }

          // 优先使用中文描述
          if (subtaskCopy.descriptionTrans) {
            subtaskCopy.description = subtaskCopy.descriptionTrans;
          }

          // 优先使用中文详情
          if (subtaskCopy.detailsTrans) {
            subtaskCopy.details = subtaskCopy.detailsTrans;
          }

          // 优先使用中文测试策略
          if (subtaskCopy.testStrategyTrans) {
            subtaskCopy.testStrategy = subtaskCopy.testStrategyTrans;
          }

          return subtaskCopy;
        });
      }

      res.json(localizedTask);
    } catch (error) {
      log('error', `Error reading task: ${error.message}`);
      res.status(500).json({ error: 'Failed to read task data' });
    }
  });

  // Task update endpoint
  app.put('/api/tasks/:taskId', async (req, res) => {
    try {
      const taskId = parseInt(req.params.taskId, 10);
      const updates = req.body;

      const data = await readTaskData(tasksPath);
      if (!data || !data.tasks) {
        return res.status(404).json({ error: 'Tasks file not found' });
      }

      // Find and update the task
      const taskIndex = data.tasks.findIndex(t => t.id === taskId);
      if (taskIndex === -1) {
        return res.status(404).json({ error: 'Task not found' });
      }

      // 获取原始任务
      const originalTask = data.tasks[taskIndex];

      // 根据翻译字段存在情况，处理更新
      // 注意：如果前端直接修改了title/description等字段，我们需要判断是修改翻译还是原始内容

      // 标题：如果任务有titleTrans字段，则更新titleTrans；否则更新title
      if (updates.title) {
        if (originalTask.titleTrans) {
          // 已有中文翻译，更新中文翻译
          updates.titleTrans = updates.title;
          // 保持原始英文不变
          updates.title = originalTask.title;
        }
        // 无中文翻译时，直接更新英文title
      }

      // 描述：如果任务有descriptionTrans字段，则更新descriptionTrans；否则更新description
      if (updates.description) {
        if (originalTask.descriptionTrans) {
          // 已有中文翻译，更新中文翻译
          updates.descriptionTrans = updates.description;
          // 保持原始英文不变
          updates.description = originalTask.description;
        }
        // 无中文翻译时，直接更新英文description
      }

      // 详情：如果任务有detailsTrans字段，则更新detailsTrans；否则更新details
      if (updates.details) {
        if (originalTask.detailsTrans) {
          // 已有中文翻译，更新中文翻译
          updates.detailsTrans = updates.details;
          // 保持原始英文不变
          updates.details = originalTask.details;
        }
        // 无中文翻译时，直接更新英文details
      }

      // 测试策略：如果任务有testStrategyTrans字段，则更新testStrategyTrans；否则更新testStrategy
      if (updates.testStrategy) {
        if (originalTask.testStrategyTrans) {
          // 已有中文翻译，更新中文翻译
          updates.testStrategyTrans = updates.testStrategy;
          // 保持原始英文不变
          updates.testStrategy = originalTask.testStrategy;
        }
        // 无中文翻译时，直接更新英文testStrategy
      }

      // 更新任务字段
      data.tasks[taskIndex] = { ...originalTask, ...updates };

      // 写入文件
      await writeTaskData(tasksPath, data);

      // 返回更新后的任务，自动处理返回值（优先使用中文翻译）
      const updatedTask = data.tasks[taskIndex];
      const localizedTask = JSON.parse(JSON.stringify(updatedTask));

      // 优先使用中文字段
      if (localizedTask.titleTrans) {
        localizedTask.title = localizedTask.titleTrans;
      }

      if (localizedTask.descriptionTrans) {
        localizedTask.description = localizedTask.descriptionTrans;
      }

      if (localizedTask.detailsTrans) {
        localizedTask.details = localizedTask.detailsTrans;
      }

      if (localizedTask.testStrategyTrans) {
        localizedTask.testStrategy = localizedTask.testStrategyTrans;
      }

      res.json(localizedTask);
    } catch (error) {
      log('error', `Error updating task: ${error.message}`);
      res.status(500).json({ error: 'Failed to update task' });
    }
  });

  app.delete('/api/tasks/:taskId', async (req, res) => {
    try {
      const taskId = parseInt(req.params.taskId, 10);

      const data = await readTaskData(tasksPath);
      if (!data || !data.tasks) {
        return res.status(404).json({ error: 'Tasks file not found' });
      }

      const taskIndex = data.tasks.findIndex(t => t.id === taskId);
      if (taskIndex === -1) {
        return res.status(404).json({ error: 'Task not found' });
      }

      data.tasks.splice(taskIndex, 1);
      removeDeletedTaskReferences(data.tasks, taskId);

      await writeTaskData(tasksPath, data);

      res.json({ success: true, taskId });
    } catch (error) {
      log('error', `Error deleting task: ${error.message}`);
      res.status(500).json({ error: 'Failed to delete task' });
    }
  });

  app.post('/api/tasks/batch-delete', async (req, res) => {
    try {
      const taskIds = Array.isArray(req.body?.taskIds)
        ? req.body.taskIds.map(id => Number(id)).filter(id => !Number.isNaN(id))
        : [];

      if (!taskIds.length) {
        return res.status(400).json({ error: 'taskIds is required' });
      }

      const data = await readTaskData(tasksPath);
      if (!data || !data.tasks) {
        return res.status(404).json({ error: 'Tasks file not found' });
      }

      const existingIds = new Set(data.tasks.map(task => Number(task.id)));
      const successIds = taskIds.filter(id => existingIds.has(id));

      data.tasks = data.tasks.filter(task => !successIds.includes(Number(task.id)));

      successIds.forEach(taskId => {
        removeDeletedTaskReferences(data.tasks, taskId);
      });

      await writeTaskData(tasksPath, data);

      res.json({
        success: true,
        successIds,
        failedIds: taskIds.filter(id => !successIds.includes(id))
      });
    } catch (error) {
      log('error', `Error batch deleting tasks: ${error.message}`);
      res.status(500).json({ error: 'Failed to batch delete tasks' });
    }
  });

  // Subtask update endpoint
  app.put('/api/tasks/:taskId/subtasks/:subtaskId', async (req, res) => {
    try {
      const taskId = parseInt(req.params.taskId, 10);
      const subtaskId = parseInt(req.params.subtaskId, 10);
      const updates = req.body;

      const data = await readTaskData(tasksPath);
      if (!data || !data.tasks) {
        return res.status(404).json({ error: 'Tasks file not found' });
      }

      // 查找父任务
      const taskIndex = data.tasks.findIndex(t => t.id === taskId);
      if (taskIndex === -1) {
        return res.status(404).json({ error: 'Parent task not found' });
      }

      const task = data.tasks[taskIndex];

      // 确保任务有子任务数组
      if (!task.subtasks || !Array.isArray(task.subtasks)) {
        return res.status(404).json({ error: 'Task has no subtasks' });
      }

      // 查找子任务
      const subtaskIndex = task.subtasks.findIndex(st => st.id === subtaskId);
      if (subtaskIndex === -1) {
        return res.status(404).json({ error: 'Subtask not found' });
      }

      // 获取原始子任务
      const originalSubtask = task.subtasks[subtaskIndex];

      // 根据翻译字段存在情况，处理更新

      // 标题：如果子任务有titleTrans字段，则更新titleTrans；否则更新title
      if (updates.title) {
        if (originalSubtask.titleTrans) {
          // 已有中文翻译，更新中文翻译
          updates.titleTrans = updates.title;
          // 保持原始英文不变
          updates.title = originalSubtask.title;
        }
        // 无中文翻译时，直接更新英文title
      }

      // 描述：如果子任务有descriptionTrans字段，则更新descriptionTrans；否则更新description
      if (updates.description) {
        if (originalSubtask.descriptionTrans) {
          // 已有中文翻译，更新中文翻译
          updates.descriptionTrans = updates.description;
          // 保持原始英文不变
          updates.description = originalSubtask.description;
        }
        // 无中文翻译时，直接更新英文description
      }

      // 详情：如果子任务有detailsTrans字段，则更新detailsTrans；否则更新details
      if (updates.details) {
        if (originalSubtask.detailsTrans) {
          // 已有中文翻译，更新中文翻译
          updates.detailsTrans = updates.details;
          // 保持原始英文不变
          updates.details = originalSubtask.details;
        }
        // 无中文翻译时，直接更新英文details
      }

      // 测试策略：如果子任务有testStrategyTrans字段，则更新testStrategyTrans；否则更新testStrategy
      if (updates.testStrategy) {
        if (originalSubtask.testStrategyTrans) {
          // 已有中文翻译，更新中文翻译
          updates.testStrategyTrans = updates.testStrategy;
          // 保持原始英文不变
          updates.testStrategy = originalSubtask.testStrategy;
        }
        // 无中文翻译时，直接更新英文testStrategy
      }

      // 更新子任务
      task.subtasks[subtaskIndex] = { ...originalSubtask, ...updates };

      // 写入文件
      await writeTaskData(tasksPath, data);

      // 返回更新后的子任务，处理返回值（优先使用中文翻译）
      const updatedSubtask = task.subtasks[subtaskIndex];
      const localizedSubtask = JSON.parse(JSON.stringify(updatedSubtask));

      // 优先使用中文字段
      if (localizedSubtask.titleTrans) {
        localizedSubtask.title = localizedSubtask.titleTrans;
      }

      if (localizedSubtask.descriptionTrans) {
        localizedSubtask.description = localizedSubtask.descriptionTrans;
      }

      if (localizedSubtask.detailsTrans) {
        localizedSubtask.details = localizedSubtask.detailsTrans;
      }

      if (localizedSubtask.testStrategyTrans) {
        localizedSubtask.testStrategy = localizedSubtask.testStrategyTrans;
      }

      res.json(localizedSubtask);
    } catch (error) {
      log('error', `Error updating subtask: ${error.message}`);
      res.status(500).json({ error: 'Failed to update subtask' });
    }
  });

  app.delete('/api/tasks/:taskId/subtasks', async (req, res) => {
    try {
      const taskId = parseInt(req.params.taskId, 10);

      const data = await readTaskData(tasksPath);
      if (!data || !data.tasks) {
        return res.status(404).json({ error: 'Tasks file not found' });
      }

      const task = data.tasks.find(t => t.id === taskId);
      if (!task) {
        return res.status(404).json({ error: 'Task not found' });
      }

      delete task.subtasks;

      await writeTaskData(tasksPath, data);

      const localizedTask = JSON.parse(JSON.stringify(task));
      if (localizedTask.titleTrans) localizedTask.title = localizedTask.titleTrans;
      if (localizedTask.descriptionTrans) localizedTask.description = localizedTask.descriptionTrans;
      if (localizedTask.detailsTrans) localizedTask.details = localizedTask.detailsTrans;
      if (localizedTask.testStrategyTrans) localizedTask.testStrategy = localizedTask.testStrategyTrans;

      res.json(localizedTask);
    } catch (error) {
      log('error', `Error clearing subtasks: ${error.message}`);
      res.status(500).json({ error: 'Failed to clear subtasks' });
    }
  });

  app.delete('/api/tasks/:taskId/subtasks/:subtaskId', async (req, res) => {
    try {
      const taskId = parseInt(req.params.taskId, 10);
      const subtaskId = parseInt(req.params.subtaskId, 10);

      const data = await readTaskData(tasksPath);
      if (!data || !data.tasks) {
        return res.status(404).json({ error: 'Tasks file not found' });
      }

      const task = data.tasks.find(t => t.id === taskId);
      if (!task) {
        return res.status(404).json({ error: 'Task not found' });
      }

      if (!task.subtasks || !Array.isArray(task.subtasks)) {
        return res.status(404).json({ error: 'Task has no subtasks' });
      }

      const subtaskIndex = task.subtasks.findIndex(st => st.id === subtaskId);
      if (subtaskIndex === -1) {
        return res.status(404).json({ error: 'Subtask not found' });
      }

      task.subtasks.splice(subtaskIndex, 1);

      if (task.subtasks.length > 0) {
        task.subtasks = task.subtasks.map(subtask => ({
          ...subtask,
          dependencies: Array.isArray(subtask.dependencies)
            ? subtask.dependencies.filter(dep => Number(dep) !== subtaskId)
            : subtask.dependencies
        }));
      } else {
        delete task.subtasks;
      }

      await writeTaskData(tasksPath, data);

      const localizedTask = JSON.parse(JSON.stringify(task));
      if (localizedTask.titleTrans) localizedTask.title = localizedTask.titleTrans;
      if (localizedTask.descriptionTrans) localizedTask.description = localizedTask.descriptionTrans;
      if (localizedTask.detailsTrans) localizedTask.details = localizedTask.detailsTrans;
      if (localizedTask.testStrategyTrans) localizedTask.testStrategy = localizedTask.testStrategyTrans;

      if (localizedTask.subtasks && localizedTask.subtasks.length > 0) {
        localizedTask.subtasks = localizedTask.subtasks.map(subtask => {
          const subtaskCopy = { ...subtask };
          if (subtaskCopy.titleTrans) subtaskCopy.title = subtaskCopy.titleTrans;
          if (subtaskCopy.descriptionTrans) subtaskCopy.description = subtaskCopy.descriptionTrans;
          if (subtaskCopy.detailsTrans) subtaskCopy.details = subtaskCopy.detailsTrans;
          if (subtaskCopy.testStrategyTrans) subtaskCopy.testStrategy = subtaskCopy.testStrategyTrans;
          return subtaskCopy;
        });
      }

      res.json(localizedTask);
    } catch (error) {
      log('error', `Error deleting subtask: ${error.message}`);
      res.status(500).json({ error: 'Failed to delete subtask' });
    }
  });

  // Regenerate a single subtask
  app.post('/api/tasks/:taskId/subtasks/:subtaskId/regenerate', async (req, res) => {
    try {
      const taskId = parseInt(req.params.taskId, 10);
      const subtaskId = parseInt(req.params.subtaskId, 10);
      const { prompt: additionalContext } = req.body || {};

      const data = await readTaskData(tasksPath);
      if (!data || !data.tasks) {
        return res.status(404).json({ error: 'Tasks file not found' });
      }

      const task = data.tasks.find(t => t.id === taskId);
      if (!task) {
        return res.status(404).json({ error: 'Task not found' });
      }

      if (!task.subtasks || !Array.isArray(task.subtasks)) {
        return res.status(404).json({ error: 'Task has no subtasks' });
      }

      const subtaskIndex = task.subtasks.findIndex(st => st.id === subtaskId);
      if (subtaskIndex === -1) {
        return res.status(404).json({ error: 'Subtask not found' });
      }

      const oldSubtask = task.subtasks[subtaskIndex];

      // 构造提示词，让 AI 重新生成这一个子任务
      const regeneratePrompt = additionalContext
        ? `Please regenerate this subtask with the following guidance: ${additionalContext}

Original subtask to regenerate:
- ID: ${subtaskId}
- Title: ${oldSubtask.titleTrans || oldSubtask.title}
- Description: ${oldSubtask.descriptionTrans || oldSubtask.description || 'N/A'}

Keep the same subtask ID (${subtaskId}) and maintain any existing dependencies.`
        : `Please regenerate this subtask with a fresh approach:

Original subtask to regenerate:
- ID: ${subtaskId}
- Title: ${oldSubtask.titleTrans || oldSubtask.title}
- Description: ${oldSubtask.descriptionTrans || oldSubtask.description || 'N/A'}

Keep the same subtask ID (${subtaskId}) and maintain any existing dependencies.`;

      // 调用 AI 生成单个子任务
      const newSubtasks = await generateSubtasks(
        task,
        1,  // 只生成一个子任务
        subtaskId,  // 使用相同的 ID
        regeneratePrompt,
        null,  // knowledgeBase
        0  // retryCount
      );

      if (!newSubtasks || newSubtasks.length === 0) {
        return res.status(500).json({ error: 'Failed to regenerate subtask' });
      }

      // 替换旧子任务
      const newSubtask = newSubtasks[0];
      // 保持原有依赖关系
      if (oldSubtask.dependencies) {
        newSubtask.dependencies = oldSubtask.dependencies;
      }
      task.subtasks[subtaskIndex] = newSubtask;

      // 写入文件
      await writeTaskData(tasksPath, data);

      // 返回更新后的任务
      const updatedData = await readTaskData(tasksPath);
      const updatedTask = updatedData?.tasks?.find(t => t.id === taskId);
      if (!updatedTask) {
        return res.status(404).json({ error: 'Task not found after regeneration' });
      }

      const localizedTask = JSON.parse(JSON.stringify(updatedTask));
      if (localizedTask.titleTrans) localizedTask.title = localizedTask.titleTrans;
      if (localizedTask.descriptionTrans) localizedTask.description = localizedTask.descriptionTrans;
      if (localizedTask.detailsTrans) localizedTask.details = localizedTask.detailsTrans;
      if (localizedTask.testStrategyTrans) localizedTask.testStrategy = localizedTask.testStrategyTrans;

      if (localizedTask.subtasks && localizedTask.subtasks.length > 0) {
        localizedTask.subtasks = localizedTask.subtasks.map(subtask => {
          const subtaskCopy = { ...subtask };
          if (subtaskCopy.titleTrans) subtaskCopy.title = subtaskCopy.titleTrans;
          if (subtaskCopy.descriptionTrans) subtaskCopy.description = subtaskCopy.descriptionTrans;
          if (subtaskCopy.detailsTrans) subtaskCopy.details = subtaskCopy.detailsTrans;
          if (subtaskCopy.testStrategyTrans) subtaskCopy.testStrategy = subtaskCopy.testStrategyTrans;
          return subtaskCopy;
        });
      }

      res.json(localizedTask);
    } catch (error) {
      log('error', `Error regenerating subtask: ${error.message}`);
      res.status(500).json({ error: 'Failed to regenerate subtask' });
    }
  });

  app.post('/api/tasks/:taskId/expand', async (req, res) => {
    try {
      const taskId = parseInt(req.params.taskId, 10);
      const numSubtasks = req.body?.num ? parseInt(req.body.num, 10) : null;
      const useResearch = req.body?.research === true;
      const additionalContext = req.body?.prompt || '';
      const knowledgeBasePath = req.body?.knowledgeBasePath || null;

      const data = await readTaskData(tasksPath);
      if (!data || !data.tasks) {
        return res.status(404).json({ error: 'Tasks file not found' });
      }

      const task = data.tasks.find(t => t.id === taskId);
      if (!task) {
        return res.status(404).json({ error: 'Task not found' });
      }

      await expandTask(taskId, numSubtasks, useResearch, additionalContext, knowledgeBasePath, tasksPath, false);

      const updatedData = await readTaskData(tasksPath);
      const updatedTask = updatedData?.tasks?.find(t => t.id === taskId);
      if (!updatedTask) {
        return res.status(404).json({ error: 'Task not found after expansion' });
      }

      const localizedTask = JSON.parse(JSON.stringify(updatedTask));
      if (localizedTask.titleTrans) localizedTask.title = localizedTask.titleTrans;
      if (localizedTask.descriptionTrans) localizedTask.description = localizedTask.descriptionTrans;
      if (localizedTask.detailsTrans) localizedTask.details = localizedTask.detailsTrans;
      if (localizedTask.testStrategyTrans) localizedTask.testStrategy = localizedTask.testStrategyTrans;

      if (localizedTask.subtasks && localizedTask.subtasks.length > 0) {
        localizedTask.subtasks = localizedTask.subtasks.map(subtask => {
          const subtaskCopy = { ...subtask };
          if (subtaskCopy.titleTrans) subtaskCopy.title = subtaskCopy.titleTrans;
          if (subtaskCopy.descriptionTrans) subtaskCopy.description = subtaskCopy.descriptionTrans;
          if (subtaskCopy.detailsTrans) subtaskCopy.details = subtaskCopy.detailsTrans;
          if (subtaskCopy.testStrategyTrans) subtaskCopy.testStrategy = subtaskCopy.testStrategyTrans;
          return subtaskCopy;
        });
      }

      res.json(localizedTask);
    } catch (error) {
      log('error', `Error expanding task: ${error.message}`);
      res.status(500).json({ error: 'Failed to expand task' });
    }
  });

  // ========================================
  // 需求管理 API 端点
  // ========================================

  // 配置 multer 用于文件上传
  const storage = multer.diskStorage({
    destination: (req, file, cb) => {
      const uploadDir = UPLOAD_DIR;
      if (!fs.existsSync(uploadDir)) {
        fs.mkdirSync(uploadDir, { recursive: true });
      }
      cb(null, uploadDir);
    },
    filename: (req, file, cb) => {
      // 临时文件名，后续会重命名
      cb(null, `temp_${Date.now()}_${file.originalname}`);
    }
  });

  const upload = multer({
    storage,
    limits: {
      fileSize: 50 * 1024 * 1024 // 50MB
    }
  });

  // 获取需求列表
  app.get('/api/requirements', async (req, res) => {
    try {
      const filters = {
        status: req.query.status,
        priority: req.query.priority,
        assignee: req.query.assignee,
        keyword: req.query.keyword,
        limit: req.query.limit
      };

      const requirements = await listRequirements(filters);
      res.json(requirements);
    } catch (error) {
      log('error', `Error fetching requirements: ${error.message}`);
      res.status(500).json({ error: 'Failed to fetch requirements' });
    }
  });

  // 获取需求统计
  app.get('/api/requirements/statistics', async (req, res) => {
    try {
      const stats = await getStatistics();
      res.json(stats);
    } catch (error) {
      log('error', `Error fetching requirement statistics: ${error.message}`);
      res.status(500).json({ error: 'Failed to fetch statistics' });
    }
  });

  // 获取单个需求详情
  app.get('/api/requirements/:id', async (req, res) => {
    try {
      const id = parseInt(req.params.id, 10);
      const requirement = await getRequirement(id);

      if (!requirement) {
        return res.status(404).json({ error: 'Requirement not found' });
      }

      res.json(requirement);
    } catch (error) {
      log('error', `Error fetching requirement: ${error.message}`);
      res.status(500).json({ error: 'Failed to fetch requirement' });
    }
  });

  // 创建需求
  app.post('/api/requirements', async (req, res) => {
    try {
      const requirement = await createNewRequirement(req.body);
      log('info', `Created requirement: ${requirement.title}`);
      res.status(201).json(requirement);
    } catch (error) {
      log('error', `Error creating requirement: ${error.message}`);
      res.status(400).json({ error: error.message });
    }
  });

  // 更新需求
  app.put('/api/requirements/:id', async (req, res) => {
    try {
      const id = parseInt(req.params.id, 10);
      const requirement = await updateRequirementById(id, req.body);

      if (!requirement) {
        return res.status(404).json({ error: 'Requirement not found' });
      }

      log('info', `Updated requirement: ${requirement.title}`);
      res.json(requirement);
    } catch (error) {
      log('error', `Error updating requirement: ${error.message}`);
      res.status(400).json({ error: error.message });
    }
  });

  // 删除需求
  app.delete('/api/requirements/:id', async (req, res) => {
    try {
      const id = parseInt(req.params.id, 10);
      const success = await deleteRequirementById(id);

      if (!success) {
        return res.status(404).json({ error: 'Requirement not found' });
      }

      log('info', `Deleted requirement: ${id}`);
      res.json({ success: true, message: 'Requirement deleted' });
    } catch (error) {
      log('error', `Error deleting requirement: ${error.message}`);
      res.status(400).json({ error: error.message });
    }
  });

  // 上传文档
  app.post('/api/requirements/:id/documents', upload.single('file'), async (req, res) => {
    try {
      const id = parseInt(req.params.id, 10);

      if (!req.file) {
        return res.status(400).json({ error: 'No file uploaded' });
      }

      const document = await uploadDocument(id, {
        originalname: req.file.originalname,
        name: req.file.originalname,
        path: req.file.path,
        size: req.file.size,
        mimetype: req.file.mimetype,
        uploadedBy: req.body.uploadedBy
      });

      log('info', `Uploaded document: ${document.name} for requirement ${id}`);
      res.status(201).json(document);
    } catch (error) {
      log('error', `Error uploading document: ${error.message}`);
      res.status(400).json({ error: error.message });
    }
  });

  // 删除文档
  app.delete('/api/requirements/:id/documents/:docId', async (req, res) => {
    try {
      const requirementId = parseInt(req.params.id, 10);
      const documentId = parseInt(req.params.docId, 10);

      const success = await removeDocument(requirementId, documentId);

      if (!success) {
        return res.status(404).json({ error: 'Document not found' });
      }

      log('info', `Deleted document ${documentId} from requirement ${requirementId}`);
      res.json({ success: true, message: 'Document deleted' });
    } catch (error) {
      log('error', `Error deleting document: ${error.message}`);
      res.status(400).json({ error: error.message });
    }
  });

  // 文档下载
  app.get('/api/requirements/:id/documents/:docId/download', async (req, res) => {
    try {
      const requirementId = parseInt(req.params.id, 10);
      const documentId = parseInt(req.params.docId, 10);

      const requirement = await getRequirement(requirementId);
      if (!requirement) {
        return res.status(404).json({ error: 'Requirement not found' });
      }

      const document = requirement.documents?.find(doc => doc.id === documentId);
      if (!document) {
        return res.status(404).json({ error: 'Document not found' });
      }

      if (!fs.existsSync(document.path)) {
        return res.status(404).json({ error: 'File not found' });
      }

      res.download(document.path, document.name);
    } catch (error) {
      log('error', `Error downloading document: ${error.message}`);
      res.status(500).json({ error: 'Failed to download document' });
    }
  });

  // ========================================
  // 需求拆分任务 API 端点
  // ========================================

  // 将需求拆分为任务
  app.post('/api/requirements/:id/split-tasks', async (req, res) => {
    try {
      const requirementId = parseInt(req.params.id, 10);
      const requirement = await getRequirement(requirementId);

      if (!requirement) {
        return res.status(404).json({ error: 'Requirement not found' });
      }

      log('info', `Splitting requirement ${requirementId} to tasks: ${requirement.title}`);

      // 读取现有任务数据
      const data = await readTaskData(tasksPath);
      if (!data || !data.tasks) {
        return res.status(500).json({ error: 'Tasks data not found' });
      }

      // 获取下一个任务 ID
      const maxTaskId = data.tasks.reduce((max, task) => Math.max(max, task.id), 0);
      const nextTaskId = maxTaskId + 1;

      // 构造任务对象用于 AI 生成
      const mockTask = {
        id: nextTaskId,
        title: requirement.title,
        description: requirement.content || requirement.title,
        details: requirement.content || ''
      };

      // 调用 AI 生成任务列表
      const subtasks = await generateSubtasks(
        mockTask,
        null, // 让 AI 决定最优数量
        1,    // 子任务起始 ID
        '',   // 额外上下文
        null  // 知识库路径
      );

      if (!subtasks || subtasks.length === 0) {
        return res.status(500).json({ error: 'Failed to generate tasks from requirement' });
      }

      // 将生成的子任务转换为独立任务
      const newTasks = subtasks.map((subtask, index) => {
        const taskId = nextTaskId + index;
        const taskTitle = subtask.titleTrans || subtask.title;
        const taskDescription = subtask.descriptionTrans || subtask.description || '';
        const taskDetails = subtask.detailsTrans || subtask.details || subtask.descriptionTrans || subtask.description || '';
        const taskTestStrategy = subtask.testStrategyTrans || subtask.testStrategy || '';

        return {
          id: taskId,
          requirementId: requirementId,
          title: taskTitle,
          titleTrans: taskTitle,
          description: taskDescription,
          descriptionTrans: taskDescription,
          details: taskDetails,
          detailsTrans: taskDetails,
          testStrategy: taskTestStrategy,
          testStrategyTrans: taskTestStrategy,
          status: 'pending',
          priority: requirement.priority === 'high' ? 'high' : requirement.priority === 'low' ? 'low' : 'medium',
          dependencies: subtask.dependencies || [],
          subtasks: [],
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString()
        };
      });

      // 添加到任务列表
      data.tasks.push(...newTasks);

      // 保存任务数据
      await writeTaskData(tasksPath, data);

      log('info', `Created ${newTasks.length} tasks from requirement ${requirementId}`);

      res.json({
        success: true,
        message: `成功拆分为 ${newTasks.length} 个任务`,
        tasks: newTasks
      });
    } catch (error) {
      log('error', `Error splitting requirement to tasks: ${error.message}`);
      res.status(500).json({ error: 'Failed to split requirement to tasks: ' + error.message });
    }
  });

  // ========================================
  // 配置管理 API 端点 (参考 OpenClaw 配置方式)
  // ========================================

  // 获取当前配置
  app.get('/api/config', (req, res) => {
    try {
      const config = loadConfig(true); // 强制重新加载
      // 隐藏敏感信息
      const safeConfig = JSON.parse(JSON.stringify(config));

      // 隐藏 API 密钥
      if (safeConfig.ai?.providers) {
        for (const provider of Object.keys(safeConfig.ai.providers)) {
          if (safeConfig.ai.providers[provider].apiKey) {
            safeConfig.ai.providers[provider].apiKey = safeConfig.ai.providers[provider].apiKey.slice(0, 8) + '...';
          }
        }
      }

      // 隐藏数据库密码
      if (safeConfig.storage?.database?.password) {
        safeConfig.storage.database.password = '********';
      }

      res.json({
        config: safeConfig,
        configPath: getConfigPath()
      });
    } catch (error) {
      log('error', `Error loading config: ${error.message}`);
      res.status(500).json({ error: 'Failed to load configuration' });
    }
  });

  // 更新配置
  app.put('/api/config', (req, res) => {
    try {
      const updates = req.body;

      // 验证更新对象
      if (!updates || typeof updates !== 'object') {
        return res.status(400).json({ error: 'Invalid update object' });
      }

      // 获取当前配置
      const currentConfig = loadConfig();

      // 处理 API 密钥更新 - 如果传入的是掩码值，则保留原值
      if (updates.ai?.providers) {
        for (const provider of Object.keys(updates.ai.providers)) {
          const newApiKey = updates.ai.providers[provider]?.apiKey;
          const currentApiKey = currentConfig.ai?.providers?.[provider]?.apiKey;

          if (newApiKey && newApiKey.endsWith('...')) {
            // 保留原值
            updates.ai.providers[provider].apiKey = currentApiKey;
          }
        }
      }

      // 处理密码更新
      if (updates.storage?.database?.password === '********') {
        updates.storage.database.password = currentConfig.storage?.database?.password;
      }

      // 合并配置
      const success = updateConfig(updates);

      if (success) {
        const updatedConfig = loadConfig(true);
        log('info', 'Configuration updated successfully');
        res.json({
          success: true,
          message: 'Configuration updated successfully',
          config: updatedConfig
        });
      } else {
        log('error', 'Failed to save configuration - updateConfig returned false');
        res.status(500).json({ error: 'Failed to save configuration. Check server logs for details.' });
      }
    } catch (error) {
      log('error', `Error updating config: ${error.message}`);
      console.error('Config update error:', error);
      res.status(500).json({ error: `Failed to update configuration: ${error.message}` });
    }
  });

  // 更新 AI 提供商
  app.put('/api/config/ai-provider', (req, res) => {
    try {
      const { provider } = req.body;

      if (!provider || !['gemini', 'qwen', 'perplexity'].includes(provider)) {
        return res.status(400).json({ error: 'Invalid provider. Must be gemini, qwen, or perplexity' });
      }

      const success = updateConfig({ ai: { provider } });

      if (success) {
        log('info', `AI provider changed to: ${provider}`);
        res.json({
          success: true,
          message: `AI provider changed to ${provider}`,
          provider
        });
      } else {
        res.status(500).json({ error: 'Failed to update AI provider' });
      }
    } catch (error) {
      log('error', `Error updating AI provider: ${error.message}`);
      res.status(500).json({ error: 'Failed to update AI provider' });
    }
  });

  // 更新提供商配置
  app.put('/api/config/ai-provider/:provider', (req, res) => {
    try {
      const { provider } = req.params;
      const updates = req.body;

      if (!['gemini', 'qwen', 'perplexity'].includes(provider)) {
        return res.status(400).json({ error: 'Invalid provider name' });
      }

      const currentConfig = loadConfig();

      // 处理 API 密钥掩码
      if (updates.apiKey && updates.apiKey.endsWith('...')) {
        updates.apiKey = currentConfig.ai?.providers?.[provider]?.apiKey;
      }

      const providerConfig = {
        ...currentConfig.ai?.providers?.[provider],
        ...updates
      };

      const success = updateConfig({
        ai: {
          providers: {
            [provider]: providerConfig
          }
        }
      });

      if (success) {
        log('info', `${provider} configuration updated`);
        res.json({
          success: true,
          message: `${provider} configuration updated`
        });
      } else {
        res.status(500).json({ error: 'Failed to update provider configuration' });
      }
    } catch (error) {
      log('error', `Error updating provider config: ${error.message}`);
      res.status(500).json({ error: 'Failed to update provider configuration' });
    }
  });

  // 重置配置为默认值
  app.post('/api/config/reset', (req, res) => {
    try {
      const success = resetConfig();

      if (success) {
        log('warn', 'Configuration reset to defaults');
        res.json({
          success: true,
          message: 'Configuration reset to defaults'
        });
      } else {
        res.status(500).json({ error: 'Failed to reset configuration' });
      }
    } catch (error) {
      log('error', `Error resetting config: ${error.message}`);
      res.status(500).json({ error: 'Failed to reset configuration' });
    }
  });

  // 非 API 路由不再提供旧静态页面
  app.get('*', (req, res) => {
    if (req.path.startsWith('/api/')) {
      return res.status(404).json({ error: 'API endpoint not found' });
    }

    res.status(404).json({ error: 'Route not found' });
  });

  // Start the server
  return new Promise((resolve, reject) => {
    try {
      const server = app.listen(port, () => {
        log('success', `Task Management API running at http://localhost:${port}`);
        console.log(`Task Management API running at http://localhost:${port}`);

        resolve(server);
      });
    } catch (error) {
      log('error', `Failed to start web server: ${error.message}`);
      reject(error);
    }
  });
}

/**
 * Stop the web server
 * @param {Object} server Server instance to stop
 * @returns {Promise<void>}
 */
export async function stopWebServer(server) {
  return new Promise((resolve, reject) => {
    if (!server) {
      log('warn', 'No server instance to stop');
      return resolve();
    }

    server.close((err) => {
      if (err) {
        log('error', `Error stopping server: ${err.message}`);
        return reject(err);
      }
      log('info', 'Web server stopped');
      resolve();
    });
  });
}