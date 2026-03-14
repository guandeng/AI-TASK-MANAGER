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

  // Reorder subtasks
  app.post('/api/tasks/:taskId/subtasks/reorder', async (req, res) => {
    try {
      const taskId = parseInt(req.params.taskId, 10);
      const { subtaskIds } = req.body;

      if (!Array.isArray(subtaskIds)) {
        return res.status(400).json({ error: 'subtaskIds must be an array' });
      }

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

      // Create a map of subtasks by id
      const subtaskMap = new Map(task.subtasks.map(st => [st.id, st]));

      // Reorder subtasks based on the provided order
      const reorderedSubtasks = subtaskIds
        .map(id => subtaskMap.get(id))
        .filter(st => st !== undefined);

      // If some subtasks were not included, append them at the end
      const reorderedIds = new Set(subtaskIds);
      const remainingSubtasks = task.subtasks.filter(st => !reorderedIds.has(st.id));
      task.subtasks = [...reorderedSubtasks, ...remainingSubtasks];

      await writeTaskData(tasksPath, data);

      // Return localized task
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
          return subtaskCopy;
        });
      }

      res.json(localizedTask);
    } catch (error) {
      log('error', `Error reordering subtasks: ${error.message}`);
      res.status(500).json({ error: 'Failed to reorder subtasks' });
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

  // 异步拆分子任务 - 立即返回消息ID，后台处理
  app.post('/api/tasks/:taskId/expand-async', async (req, res) => {
    try {
      const taskId = parseInt(req.params.taskId, 10);
      const additionalContext = req.body?.prompt || req.body?.additionalContext || '';
      const knowledgeBasePath = req.body?.knowledgeBasePath || null;

      const data = await readTaskData(tasksPath);
      if (!data || !data.tasks) {
        return res.status(404).json({ error: 'Tasks file not found' });
      }

      const task = data.tasks.find(t => t.id === taskId);
      if (!task) {
        return res.status(404).json({ error: 'Task not found' });
      }

      // 创建消息记录
      const { createMessage, updateMessageStatus } = await import('../modules/message-storage.js');
      const message = await createMessage(taskId, 'expand_task', `拆分任务: ${task.titleTrans || task.title}`);

      // 立即返回消息ID
      res.json({ messageId: message.id, status: 'pending' });

      // 后台异步处理
      (async () => {
        try {
          // 更新消息状态为处理中
          await updateMessageStatus(message.id, 'processing', {
            content: '正在拆分任务...'
          });

          // 执行拆分
          await expandTask(taskId, null, false, additionalContext, knowledgeBasePath, tasksPath, false);

          // 更新消息状态为成功
          await updateMessageStatus(message.id, 'success', {
            resultSummary: '任务拆分完成'
          });

          log('info', `Async task expansion completed: task ${taskId}, message ${message.id}`);
        } catch (error) {
          log('error', `Async task expansion failed: ${error.message}`);
          // 更新消息状态为失败
          await updateMessageStatus(message.id, 'failed', {
            errorMessage: error.message || '拆分任务失败'
          });
        }
      })();

    } catch (error) {
      log('error', `Error starting async task expansion: ${error.message}`);
      res.status(500).json({ error: 'Failed to start async task expansion' });
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

  // ========================================
  // 成员管理 API 端点
  // ========================================

  // 获取成员列表
  app.get('/api/members', async (req, res) => {
    try {
      const { status, department, role, search, page, pageSize } = req.query;
      const filters = {};
      if (status) filters.status = status;
      if (department) filters.department = department;
      if (role) filters.role = role;
      if (search) filters.search = search;

      const { listMembersWithPaging, listMembers } = await import('../modules/member-manager.js');

      if (page || pageSize) {
        const result = await listMembersWithPaging({
          page: page ? parseInt(page, 10) : 1,
          pageSize: pageSize ? parseInt(pageSize, 10) : 20,
          filters
        });
        res.json(result);
      } else {
        const members = await listMembers(filters);
        res.json(members);
      }
    } catch (error) {
      log('error', `Error fetching members: ${error.message}`);
      res.status(500).json({ error: 'Failed to fetch members' });
    }
  });

  // 获取单个成员
  app.get('/api/members/:id', async (req, res) => {
    try {
      const { getMember } = await import('../modules/member-manager.js');
      const member = await getMember(parseInt(req.params.id, 10));
      if (!member) {
        return res.status(404).json({ error: 'Member not found' });
      }
      res.json(member);
    } catch (error) {
      log('error', `Error fetching member: ${error.message}`);
      res.status(500).json({ error: 'Failed to fetch member' });
    }
  });

  // 创建成员
  app.post('/api/members', async (req, res) => {
    try {
      const { createNewMember } = await import('../modules/member-manager.js');
      const member = await createNewMember(req.body);
      res.status(201).json(member);
    } catch (error) {
      log('error', `Error creating member: ${error.message}`);
      res.status(400).json({ error: error.message });
    }
  });

  // 更新成员
  app.put('/api/members/:id', async (req, res) => {
    try {
      const { updateMemberById } = await import('../modules/member-manager.js');
      const member = await updateMemberById(parseInt(req.params.id, 10), req.body);
      if (!member) {
        return res.status(404).json({ error: 'Member not found' });
      }
      res.json(member);
    } catch (error) {
      log('error', `Error updating member: ${error.message}`);
      res.status(400).json({ error: error.message });
    }
  });

  // 删除成员
  app.delete('/api/members/:id', async (req, res) => {
    try {
      const { deleteMemberById } = await import('../modules/member-manager.js');
      const success = await deleteMemberById(parseInt(req.params.id, 10));
      if (!success) {
        return res.status(404).json({ error: 'Member not found' });
      }
      res.json({ success: true });
    } catch (error) {
      log('error', `Error deleting member: ${error.message}`);
      res.status(500).json({ error: 'Failed to delete member' });
    }
  });

  // ============ 活动日志 API ============

  // 获取任务活动日志
  app.get('/api/tasks/:taskId/activities', async (req, res) => {
    try {
      const { getTaskActivities } = await import('../modules/activity-storage.js');
      const taskId = parseInt(req.params.taskId, 10);
      const options = {
        subtaskId: req.query.subtaskId ? parseInt(req.query.subtaskId, 10) : undefined,
        action: req.query.action,
        limit: req.query.limit ? parseInt(req.query.limit, 10) : 50
      };
      const activities = await getTaskActivities(taskId, options);
      res.json(activities);
    } catch (error) {
      log('error', `Error fetching task activities: ${error.message}`);
      res.status(500).json({ error: 'Failed to fetch task activities' });
    }
  });

  // 获取全局活动日志
  app.get('/api/activities', async (req, res) => {
    try {
      const { getGlobalActivities } = await import('../modules/activity-storage.js');
      const options = {
        memberId: req.query.memberId ? parseInt(req.query.memberId, 10) : undefined,
        action: req.query.action,
        startDate: req.query.startDate,
        endDate: req.query.endDate,
        limit: req.query.limit ? parseInt(req.query.limit, 10) : 50,
        offset: req.query.offset ? parseInt(req.query.offset, 10) : 0
      };
      const activities = await getGlobalActivities(options);
      res.json(activities);
    } catch (error) {
      log('error', `Error fetching global activities: ${error.message}`);
      res.status(500).json({ error: 'Failed to fetch activities' });
    }
  });

  // 获取活动统计
  app.get('/api/activities/statistics', async (req, res) => {
    try {
      const { getActivityStatistics } = await import('../modules/activity-storage.js');
      const options = {
        startDate: req.query.startDate,
        endDate: req.query.endDate
      };
      const statistics = await getActivityStatistics(options);
      res.json(statistics);
    } catch (error) {
      log('error', `Error fetching activity statistics: ${error.message}`);
      res.status(500).json({ error: 'Failed to fetch activity statistics' });
    }
  });

  // ============ 菜单管理 API ============

  // 获取菜单列表
  app.get('/api/menus', async (req, res) => {
    try {
      const { getMenuList } = await import('../modules/menu-storage.js');
      const menus = await getMenuList();
      res.json({ menus });
    } catch (error) {
      log('error', `Error fetching menus: ${error.message}`);
      res.status(500).json({ error: 'Failed to fetch menus' });
    }
  });

  // 获取菜单树
  app.get('/api/menus/tree', async (req, res) => {
    try {
      const { getMenuTree } = await import('../modules/menu-storage.js');
      const menus = await getMenuTree();
      res.json({ menus });
    } catch (error) {
      log('error', `Error fetching menu tree: ${error.message}`);
      res.status(500).json({ error: 'Failed to fetch menu tree' });
    }
  });

  // 获取单个菜单
  app.get('/api/menus/:key', async (req, res) => {
    try {
      const { getMenuByKey } = await import('../modules/menu-storage.js');
      const menu = await getMenuByKey(req.params.key);
      if (!menu) {
        return res.status(404).json({ error: 'Menu not found' });
      }
      res.json({ menu });
    } catch (error) {
      log('error', `Error fetching menu: ${error.message}`);
      res.status(500).json({ error: 'Failed to fetch menu' });
    }
  });

  // 创建菜单
  app.post('/api/menus', async (req, res) => {
    try {
      const { createMenu } = await import('../modules/menu-storage.js');
      const menu = await createMenu(req.body);
      res.status(201).json({ success: true, message: 'Menu created successfully', menu });
    } catch (error) {
      log('error', `Error creating menu: ${error.message}`);
      res.status(400).json({ error: error.message });
    }
  });

  // 更新菜单
  app.put('/api/menus/:key', async (req, res) => {
    try {
      const { updateMenuByKey } = await import('../modules/menu-storage.js');
      const menu = await updateMenuByKey(req.params.key, req.body);
      if (!menu) {
        return res.status(404).json({ error: 'Menu not found' });
      }
      res.json({ success: true, message: 'Menu updated successfully', menu });
    } catch (error) {
      log('error', `Error updating menu: ${error.message}`);
      res.status(400).json({ error: error.message });
    }
  });

  // 删除菜单
  app.delete('/api/menus/:key', async (req, res) => {
    try {
      const { deleteMenuByKey } = await import('../modules/menu-storage.js');
      const success = await deleteMenuByKey(req.params.key);
      if (!success) {
        return res.status(404).json({ error: 'Menu not found' });
      }
      res.json({ success: true, message: 'Menu deleted successfully' });
    } catch (error) {
      log('error', `Error deleting menu: ${error.message}`);
      res.status(500).json({ error: 'Failed to delete menu' });
    }
  });

  // 批量删除菜单
  app.post('/api/menus/batch-delete', async (req, res) => {
    try {
      const { batchDeleteMenus } = await import('../modules/menu-storage.js');
      const result = await batchDeleteMenus(req.body.keys || []);
      res.json(result);
    } catch (error) {
      log('error', `Error batch deleting menus: ${error.message}`);
      res.status(500).json({ error: 'Failed to batch delete menus' });
    }
  });

  // 更新菜单排序
  app.put('/api/menus/reorder', async (req, res) => {
    try {
      const { reorderMenus } = await import('../modules/menu-storage.js');
      await reorderMenus(req.body);
      res.json({ success: true });
    } catch (error) {
      log('error', `Error reordering menus: ${error.message}`);
      res.status(500).json({ error: 'Failed to reorder menus' });
    }
  });

  // 移动菜单
  app.put('/api/menus/:key/move', async (req, res) => {
    try {
      const { moveMenu } = await import('../modules/menu-storage.js');
      const menu = await moveMenu(req.params.key, req.body.targetParentKey);
      if (!menu) {
        return res.status(404).json({ error: 'Menu not found' });
      }
      res.json({ success: true, message: 'Menu moved successfully', menu });
    } catch (error) {
      log('error', `Error moving menu: ${error.message}`);
      res.status(500).json({ error: 'Failed to move menu' });
    }
  });

  // 切换菜单启用状态
  app.put('/api/menus/:key/toggle', async (req, res) => {
    try {
      const { toggleMenuEnabled } = await import('../modules/menu-storage.js');
      const menu = await toggleMenuEnabled(req.params.key, req.body.enabled);
      if (!menu) {
        return res.status(404).json({ error: 'Menu not found' });
      }
      res.json({ success: true, message: 'Menu toggled successfully', menu });
    } catch (error) {
      log('error', `Error toggling menu: ${error.message}`);
      res.status(500).json({ error: 'Failed to toggle menu' });
    }
  });

  // ============ 消息管理 API ============

  // 获取消息列表
  app.get('/api/messages', async (req, res) => {
    try {
      const { getMessages } = await import('../modules/message-storage.js');
      const options = {
        taskId: req.query.taskId ? parseInt(req.query.taskId, 10) : undefined,
        type: req.query.type,
        status: req.query.status,
        isRead: req.query.isRead === 'true' ? true : req.query.isRead === 'false' ? false : undefined,
        limit: req.query.limit ? parseInt(req.query.limit, 10) : 50,
        offset: req.query.offset ? parseInt(req.query.offset, 10) : 0
      };
      const messages = await getMessages(options);
      res.json(messages);
    } catch (error) {
      log('error', `Error fetching messages: ${error.message}`);
      res.status(500).json({ error: 'Failed to fetch messages' });
    }
  });

  // 获取未读消息数量
  app.get('/api/messages/unread-count', async (req, res) => {
    try {
      const { getUnreadCount } = await import('../modules/message-storage.js');
      const count = await getUnreadCount();
      res.json({ count });
    } catch (error) {
      log('error', `Error fetching unread count: ${error.message}`);
      res.status(500).json({ error: 'Failed to fetch unread count' });
    }
  });

  // 标记消息为已读
  app.put('/api/messages/:id/read', async (req, res) => {
    try {
      const { markAsRead } = await import('../modules/message-storage.js');
      const success = await markAsRead(parseInt(req.params.id, 10));
      if (!success) {
        return res.status(404).json({ error: 'Message not found' });
      }
      res.json({ success: true });
    } catch (error) {
      log('error', `Error marking message as read: ${error.message}`);
      res.status(500).json({ error: 'Failed to mark message as read' });
    }
  });

  // 标记所有消息为已读
  app.put('/api/messages/read-all', async (req, res) => {
    try {
      const { markAllAsRead } = await import('../modules/message-storage.js');
      await markAllAsRead();
      res.json({ success: true });
    } catch (error) {
      log('error', `Error marking all messages as read: ${error.message}`);
      res.status(500).json({ error: 'Failed to mark all messages as read' });
    }
  });

  // 删除消息
  app.delete('/api/messages/:id', async (req, res) => {
    try {
      const { deleteMessage } = await import('../modules/message-storage.js');
      const success = await deleteMessage(parseInt(req.params.id, 10));
      if (!success) {
        return res.status(404).json({ error: 'Message not found' });
      }
      res.json({ success: true });
    } catch (error) {
      log('error', `Error deleting message: ${error.message}`);
      res.status(500).json({ error: 'Failed to delete message' });
    }
  });

  // ========================================
  // Assignment API Routes
  // ========================================
  const {
    getTaskAssignments,
    getSubtaskAssignments,
    assignTaskToMember,
    assignSubtaskToMember,
    removeTaskAssignment,
    removeSubtaskAssignment,
    getMemberAssignments,
    getMemberWorkload,
    updateTaskTimeFields
  } = await import('../modules/assignment-storage.js');

  // 获取任务分配列表
  app.get('/api/tasks/:id/assignments', async (req, res) => {
    try {
      const taskId = parseInt(req.params.id, 10);
      const assignments = await getTaskAssignments(taskId);
      res.json(assignments);
    } catch (error) {
      log('error', `Error getting task assignments: ${error.message}`);
      res.status(500).json({ error: 'Failed to get task assignments' });
    }
  });

  // 获取任务分配概览
  app.get('/api/tasks/:id/assignments/overview', async (req, res) => {
    try {
      const taskId = parseInt(req.params.id, 10);
      const assignments = await getTaskAssignments(taskId);

      const overview = {
        totalAssignees: assignments.length,
        assignees: assignments.filter(a => a.role === 'assignee').map(a => a.member),
        reviewers: assignments.filter(a => a.role === 'reviewer').map(a => a.member),
        collaborators: assignments.filter(a => a.role === 'collaborator').map(a => a.member),
        totalEstimatedHours: assignments.reduce((sum, a) => sum + (a.estimatedHours || 0), 0),
        totalActualHours: assignments.reduce((sum, a) => sum + (a.actualHours || 0), 0)
      };

      res.json(overview);
    } catch (error) {
      log('error', `Error getting assignment overview: ${error.message}`);
      res.status(500).json({ error: 'Failed to get assignment overview' });
    }
  });

  // 分配任务给成员
  app.post('/api/tasks/:id/assignments', async (req, res) => {
    try {
      const taskId = parseInt(req.params.id, 10);
      const { memberId, role, assignedBy, estimatedHours, actualHours } = req.body;

      if (!memberId) {
        return res.status(400).json({ error: 'memberId is required' });
      }

      const assignments = await assignTaskToMember(taskId, memberId, {
        role,
        assignedBy,
        estimatedHours,
        actualHours
      });

      res.json(assignments);
    } catch (error) {
      log('error', `Error assigning task: ${error.message}`);
      res.status(500).json({ error: error.message || 'Failed to assign task' });
    }
  });

  // 移除任务分配
  app.delete('/api/tasks/:taskId/assignments/:assignmentId', async (req, res) => {
    try {
      const taskId = parseInt(req.params.taskId, 10);
      const assignmentId = parseInt(req.params.assignmentId, 10);

      const success = await removeTaskAssignment(taskId, assignmentId);

      if (success) {
        res.json({ success: true, message: 'Assignment removed successfully' });
      } else {
        res.status(404).json({ error: 'Assignment not found' });
      }
    } catch (error) {
      log('error', `Error removing task assignment: ${error.message}`);
      res.status(500).json({ error: 'Failed to remove task assignment' });
    }
  });

  // 获取子任务分配列表
  app.get('/api/tasks/:taskId/subtasks/:subtaskId/assignments', async (req, res) => {
    try {
      const subtaskId = parseInt(req.params.subtaskId, 10);
      const assignments = await getSubtaskAssignments(subtaskId);
      res.json(assignments);
    } catch (error) {
      log('error', `Error getting subtask assignments: ${error.message}`);
      res.status(500).json({ error: 'Failed to get subtask assignments' });
    }
  });

  // 分配子任务给成员
  app.post('/api/tasks/:taskId/subtasks/:subtaskId/assignments', async (req, res) => {
    try {
      const subtaskId = parseInt(req.params.subtaskId, 10);
      const { memberId, role, assignedBy, estimatedHours, actualHours } = req.body;

      if (!memberId) {
        return res.status(400).json({ error: 'memberId is required' });
      }

      const assignments = await assignSubtaskToMember(subtaskId, memberId, {
        role,
        assignedBy,
        estimatedHours,
        actualHours
      });

      res.json(assignments);
    } catch (error) {
      log('error', `Error assigning subtask: ${error.message}`);
      res.status(500).json({ error: error.message || 'Failed to assign subtask' });
    }
  });

  // 移除子任务分配
  app.delete('/api/tasks/:taskId/subtasks/:subtaskId/assignments/:assignmentId', async (req, res) => {
    try {
      const subtaskId = parseInt(req.params.subtaskId, 10);
      const assignmentId = parseInt(req.params.assignmentId, 10);

      const success = await removeSubtaskAssignment(subtaskId, assignmentId);

      if (success) {
        res.json({ success: true, message: 'Assignment removed successfully' });
      } else {
        res.status(404).json({ error: 'Assignment not found' });
      }
    } catch (error) {
      log('error', `Error removing subtask assignment: ${error.message}`);
      res.status(500).json({ error: 'Failed to remove subtask assignment' });
    }
  });

  // 获取成员任务分配列表
  app.get('/api/members/:id/assignments', async (req, res) => {
    try {
      const memberId = parseInt(req.params.id, 10);
      const { role, status, limit } = req.query;

      const filters = {};
      if (role) filters.role = role;
      if (status) filters.status = status;
      if (limit) filters.limit = parseInt(limit, 10);

      const assignments = await getMemberAssignments(memberId, filters);
      res.json(assignments);
    } catch (error) {
      log('error', `Error getting member assignments: ${error.message}`);
      res.status(500).json({ error: 'Failed to get member assignments' });
    }
  });

  // 获取成员工作量
  app.get('/api/members/:id/workload', async (req, res) => {
    try {
      const memberId = parseInt(req.params.id, 10);
      const workload = await getMemberWorkload(memberId);
      res.json(workload);
    } catch (error) {
      log('error', `Error getting member workload: ${error.message}`);
      res.status(500).json({ error: 'Failed to get member workload' });
    }
  });

  // 更新任务时间信息
  app.put('/api/tasks/:id/time', async (req, res) => {
    try {
      const taskId = parseInt(req.params.id, 10);
      const { startDate, dueDate, completedAt, estimatedHours, actualHours } = req.body;

      const success = await updateTaskTimeFields(taskId, {
        startDate,
        dueDate,
        completedAt,
        estimatedHours,
        actualHours
      });

      if (success) {
        res.json({ success: true, message: 'Task time info updated successfully' });
      } else {
        res.json({ success: true, message: 'No fields to update' });
      }
    } catch (error) {
      log('error', `Error updating task time: ${error.message}`);
      res.status(500).json({ error: 'Failed to update task time' });
    }
  });

  // 提供前端静态文件
  const frontendDistPath = path.join(__dirname, '../../frontend/dist');

  // 检查前端构建目录是否存在
  if (fs.existsSync(frontendDistPath)) {
    // 静态资源文件（JS、CSS、图片等）
    app.use(express.static(frontendDistPath, { index: false }));

    // 所有非 API 路由都返回 index.html（支持前端路由）
    app.get('*', (req, res, next) => {
      if (req.path.startsWith('/api/')) {
        return res.status(404).json({ error: 'API endpoint not found' });
      }

      const indexPath = path.join(frontendDistPath, 'index.html');
      if (fs.existsSync(indexPath)) {
        return res.sendFile(indexPath);
      }

      res.status(404).json({ error: 'Frontend not built. Please run: cd frontend && pnpm build' });
    });

    log('success', `Serving frontend from: ${frontendDistPath}`);
  } else {
    log('warn', `Frontend dist not found at ${frontendDistPath}`);
    log('warn', 'Please build the frontend first: cd frontend && pnpm build');

    // 如果前端没有构建，返回提示信息
    app.get('*', (req, res) => {
      if (req.path.startsWith('/api/')) {
        return res.status(404).json({ error: 'API endpoint not found' });
      }

      res.status(404).json({
        error: 'Frontend not built',
        message: 'Please build the frontend first',
        command: 'cd frontend && pnpm build'
      });
    });
  }

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