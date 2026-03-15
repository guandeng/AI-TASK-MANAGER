import { defineStore } from 'pinia';
import { computed, ref } from 'vue';
import {
  batchDeleteTasks as batchDeleteTasksApi,
  clearTaskSubtasks as clearTaskSubtasksApi,
  createTask as createTaskApi,
  deleteTask as deleteTaskApi,
  deleteSubtask as deleteSubtaskApi,
  expandTask as expandTaskApi,
  expandTaskAsync as expandTaskAsyncApi,
  fetchTaskDetail,
  fetchTaskList,
  regenerateSubtask as regenerateSubtaskApi,
  reorderSubtasks as reorderSubtasksApi,
  updateTask,
  updateSubtask as updateSubtaskApi,
  copyTask as copyTaskApi,
  getTaskDependencies as getTaskDependenciesApi,
  addTaskDependency as addTaskDependencyApi,
  removeTaskDependency as removeTaskDependencyApi,
  validateDependencies as validateDependenciesApi,
  getReadyTasks as getReadyTasksApi
} from '@/service/api/task';
import type { Task, Subtask, TaskStatus, TaskStatistics, TaskListParams, TaskCreateRequest, TaskDependency } from '@/typings/api/task';
import { createLoadingService, LoadingService, LOADING_PRESETS } from '@/utils/loading-service';

// 辅助函数：提取后端返回的 data 字段
// 后端返回格式: { code: 0, message: "success", data: {...} }
function extractData(responseData: any): any {
  if (!responseData) return null;
  // 如果有 data 字段，返回 data 字段内容
  if (responseData.data !== undefined) {
    return responseData.data;
  }
  // 否则返回原始数据
  return responseData;
}

export const useTaskStore = defineStore('task-store', () => {
  // 状态
  const tasks = ref<Task[]>([]);
  const currentTask = ref<Task | null>(null);
  const loading = ref(false);
  const projectName = ref('');
  const projectVersion = ref('');
  const total = ref(0);
  const page = ref(1);
  const pageSize = ref(20);
  const taskDependencies = ref<TaskDependency[]>([]);

  // 计算属性：任务统计
  const statistics = computed<TaskStatistics>(() => {
    const stats: TaskStatistics = {
      total: tasks.value.length,
      done: 0,
      pending: 0,
      deferred: 0,
      inProgress: 0,
      highPriority: 0,
      mediumPriority: 0,
      lowPriority: 0
    };

    tasks.value.forEach(task => {
      // 状态统计
      if (task.status === 'done') stats.done++;
      else if (task.status === 'pending') stats.pending++;
      else if (task.status === 'deferred') stats.deferred++;
      else if (task.status === 'in-progress') stats.inProgress++;

      // 优先级统计
      if (task.priority === 'high') stats.highPriority++;
      else if (task.priority === 'medium') stats.mediumPriority++;
      else if (task.priority === 'low') stats.lowPriority++;
    });

    return stats;
  });

  // 计算属性：按状态分组的任务
  const tasksByStatus = computed(() => {
    return {
      pending: tasks.value.filter(t => t.status === 'pending'),
      done: tasks.value.filter(t => t.status === 'done'),
      deferred: tasks.value.filter(t => t.status === 'deferred'),
      inProgress: tasks.value.filter(t => t.status === 'in-progress')
    };
  });

  // Actions
  async function loadTasks(params?: TaskListParams) {
    loading.value = true;
    try {
      const queryParams = {
        page: params?.page || 1,
        pageSize: params?.pageSize || 20,
        ...params
      };

      const { data, error } = await fetchTaskList(queryParams);
      if (!error && data) {
        const responseData = extractData(data);

        if (responseData && 'list' in responseData && 'total' in responseData) {
          // 分页响应格式
          tasks.value = responseData.list || [];
          total.value = responseData.total || 0;
          page.value = responseData.page || 1;
          pageSize.value = responseData.pageSize || 20;
        } else if (Array.isArray(responseData)) {
          // 数组格式（兼容旧接口）
          tasks.value = responseData;
          total.value = responseData.length;
        } else if (responseData && 'tasks' in responseData) {
          // 兼容旧格式
          tasks.value = responseData.tasks || [];
          projectName.value = responseData.projectName || '';
          projectVersion.value = responseData.projectVersion || '';
          total.value = responseData.tasks?.length || 0;
        }
      }
    } catch (error) {
      window.$message?.error('加载任务列表失败');
      console.error('Failed to load tasks:', error);
    } finally {
      loading.value = false;
    }
  }

  async function loadTaskDetail(id: number, locale: string = 'zh') {
    loading.value = true;
    try {
      const { data, error } = await fetchTaskDetail(id, locale);
      if (!error && data) {
        currentTask.value = extractData(data);
      }
      return extractData(data);
    } catch (error) {
      window.$message?.error('加载任务详情失败');
      console.error('Failed to load task detail:', error);
      return null;
    } finally {
      loading.value = false;
    }
  }

  async function setTaskStatus(id: number, status: TaskStatus) {
    try {
      const { data: responseData, error } = await updateTask(id, { status });
      if (!error && responseData) {
        const updatedTask = extractData(responseData);
        // 更新列表中的任务
        const index = tasks.value.findIndex(t => t.id === id);
        if (index !== -1) {
          tasks.value[index] = { ...tasks.value[index], ...updatedTask };
        }
        // 如果是当前任务，也更新
        if (currentTask.value?.id === id) {
          currentTask.value = { ...currentTask.value, ...updatedTask };
        }
        window.$message?.success('状态更新成功');
        return true;
      }
      return false;
    } catch (error) {
      window.$message?.error('状态更新失败');
      console.error('Failed to update task status:', error);
      return false;
    }
  }

  async function setTaskAssignee(id: number, assignee: string) {
    try {
      const { data: responseData, error } = await updateTask(id, { assignee });
      if (!error && responseData) {
        const updatedTask = extractData(responseData);
        // 更新列表中的任务
        const index = tasks.value.findIndex(t => t.id === id);
        if (index !== -1) {
          tasks.value[index] = { ...tasks.value[index], ...updatedTask };
        }
        // 如果是当前任务，也更新
        if (currentTask.value?.id === id) {
          currentTask.value = { ...currentTask.value, ...updatedTask };
        }
        window.$message?.success('负责人更新成功');
        return true;
      }
      return false;
    } catch (error) {
      window.$message?.error('负责人更新失败');
      console.error('Failed to update task assignee:', error);
      return false;
    }
  }

  // 通用更新任务方法
  async function updateTaskById(id: number, updateData: Record<string, any>) {
    try {
      const { data: responseData, error } = await updateTask(id, updateData);
      if (!error && responseData) {
        const updatedTask = extractData(responseData);
        // 更新列表中的任务
        const index = tasks.value.findIndex(t => t.id === id);
        if (index !== -1) {
          tasks.value[index] = { ...tasks.value[index], ...updatedTask };
        }
        // 如果是当前任务，完全替换对象以触发响应式更新
        if (currentTask.value?.id === id) {
          currentTask.value = null;
          currentTask.value = updatedTask;
        }
        return true;
      }
      return false;
    } catch (error) {
      console.error('Failed to update task:', error);
      return false;
    }
  }

  async function setSubtaskStatus(taskId: number, subtaskId: number, status: TaskStatus) {
    try {
      const { data: responseData, error } = await updateSubtaskApi(taskId, subtaskId, { status });
      if (!error && responseData) {
        const updatedTask = extractData(responseData);
        // 后端返回的是整个 task 对象（包含更新后的 subtasks）
        if (updatedTask && updatedTask.subtasks) {
          // 更新列表中的任务
          const taskIndex = tasks.value.findIndex(t => t.id === taskId);
          if (taskIndex !== -1) {
            tasks.value[taskIndex] = { ...tasks.value[taskIndex], ...updatedTask };
          }
          // 如果是当前任务，直接替换
          if (currentTask.value?.id === taskId) {
            currentTask.value = { ...currentTask.value, ...updatedTask };
          }
        }
        window.$message?.success('子任务状态更新成功');
        return true;
      }
      return false;
    } catch (error) {
      window.$message?.error('子任务状态更新失败');
      console.error('Failed to update subtask status:', error);
      return false;
    }
  }

  // 通用子任务更新���法
  async function updateSubtask(taskId: number, subtaskId: number, updateData: Record<string, any>) {
    console.log('[store.updateSubtask] 开始调用 API', { taskId, subtaskId, updateData });
    try {
      const { data: responseData, error } = await updateSubtaskApi(taskId, subtaskId, updateData);
      console.log('[store.updateSubtask] API 返回', { responseData, error });
      if (!error && responseData) {
        const updatedTask = extractData(responseData);
        // 后端返回的是整个 task 对象（包含更新后的 subtasks）
        if (updatedTask && updatedTask.subtasks) {
          // 更新列表中的任务
          const taskIndex = tasks.value.findIndex(t => t.id === taskId);
          if (taskIndex !== -1) {
            tasks.value[taskIndex] = { ...tasks.value[taskIndex], ...updatedTask };
          }
          // 如果是当前任务，创建新对象确保响应式更新
          if (currentTask.value?.id === taskId) {
            currentTask.value = {
              ...currentTask.value,
              ...updatedTask,
              subtasks: [...updatedTask.subtasks] // 确保是新的数组引用
            };
          }
        }
        return true;
      }
      return false;
    } catch (error) {
      console.error('Failed to update subtask:', error);
      return false;
    }
  }

  async function expandTask(id: number, loadingService?: LoadingService) {
    // 使用传入的 loadingService 或创建新的
    const ls = loadingService || createLoadingService();
    const shouldManageLoading = !loadingService;

    if (shouldManageLoading) {
      ls.start(LOADING_PRESETS.expandTask);
    }

    loading.value = true;
    try {
      // 步骤1: 分析任务内容
      if (shouldManageLoading) ls.nextStep();

      const { data: responseData, error } = await expandTaskApi(id);

      // 步骤2: 生成完成
      if (shouldManageLoading) ls.nextStep();

      if (!error && responseData) {
        const updatedTask = extractData(responseData);
        const index = tasks.value.findIndex(t => t.id === id);
        if (index !== -1) {
          tasks.value[index] = { ...tasks.value[index], ...updatedTask };
        }
        currentTask.value = updatedTask;
        await loadTasks();

        if (shouldManageLoading) {
          ls.success('子任务拆分成功');
        } else {
          window.$message?.success('子任务拆分成功');
        }
        return true;
      }

      if (shouldManageLoading) ls.error('子任务拆分失败');
      return false;
    } catch (error) {
      if (shouldManageLoading) {
        ls.error('子任务拆分失败');
      } else {
        window.$message?.error('子任务拆分失败');
      }
      console.error('Failed to expand task:', error);
      return false;
    } finally {
      loading.value = false;
    }
  }

  // 异步拆分子任务 - 返回消息ID用于轮询
  async function expandTaskAsync(id: number) {
    try {
      const { data, error } = await expandTaskAsyncApi(id);
      // 后端返回格式: { code: 0, message: "success", data: { message: "..." } }
      // 当 code === 0 时，error 为 null，data 为响应体
      if (!error && data) {
        // data 已经是后端返回的完整响应体 { code, message, data }
        const innerData = extractData(data);
        // 异步展开成功，返回 true 表示已开始
        if (innerData?.message || (data as any).code === 0) {
          return true;
        }
      }
      return null;
    } catch (error) {
      console.error('Failed to start async task expansion:', error);
      window.$message?.error('启动异步拆分失败');
      return null;
    }
  }

  async function clearTaskSubtasks(taskId: number) {
    loading.value = true;
    try {
      const { data: responseData, error } = await clearTaskSubtasksApi(taskId);
      if (!error && responseData) {
        const updatedTask = extractData(responseData);
        const index = tasks.value.findIndex(t => t.id === taskId);
        if (index !== -1) {
          tasks.value[index] = { ...tasks.value[index], ...updatedTask };
        }
        currentTask.value = updatedTask;
        await loadTasks();
        window.$message?.success('子任务已清空');
        return true;
      }
      return false;
    } catch (error) {
      window.$message?.error('清空子任务失败');
      console.error('Failed to clear subtasks:', error);
      return false;
    } finally {
      loading.value = false;
    }
  }

  async function deleteTask(taskId: number) {
    loading.value = true;
    try {
      const { data: result, error } = await deleteTaskApi(taskId);
      if (!error && result) {
        tasks.value = tasks.value.filter(task => task.id !== taskId);
        if (currentTask.value?.id === taskId) {
          currentTask.value = null;
        }
        await loadTasks();
        window.$message?.success('任务已删除');
        return true;
      }
      return false;
    } catch (error) {
      window.$message?.error('删除任务失败');
      console.error('Failed to delete task:', error);
      return false;
    } finally {
      loading.value = false;
    }
  }

  async function batchDeleteTasks(taskIds: number[]) {
    if (!taskIds.length) {
      return { successIds: [], failedIds: [] };
    }

    loading.value = true;

    try {
      const { data, error } = await batchDeleteTasksApi(taskIds);

      if (error) {
        return { successIds: [], failedIds: taskIds };
      }

      // 后端返回 { deleted: number } 格式
      const responseData = extractData(data);
      const deletedCount = responseData?.deleted || 0;

      // 如果删除数量大于0，认为全部成功
      const successIds = deletedCount > 0 ? taskIds : [];
      const failedIds = deletedCount > 0 ? [] : taskIds;

      if (successIds.length) {
        tasks.value = tasks.value.filter(task => !successIds.includes(task.id));
        if (currentTask.value?.id && successIds.includes(currentTask.value.id)) {
          currentTask.value = null;
        }
        await loadTasks();
      }

      return { successIds, failedIds };
    } finally {
      loading.value = false;
    }
  }

  async function deleteSubtask(taskId: number, subtaskId: number) {
    loading.value = true;
    try {
      const { data: responseData, error } = await deleteSubtaskApi(taskId, subtaskId);
      if (!error && responseData) {
        const updatedTask = extractData(responseData);
        // 后端返回的是整个 task 对象
        if (updatedTask) {
          const index = tasks.value.findIndex(t => t.id === taskId);
          if (index !== -1) {
            tasks.value[index] = { ...tasks.value[index], ...updatedTask };
          }
          // 更新当前任务
          if (currentTask.value?.id === taskId) {
            currentTask.value = { ...currentTask.value, ...updatedTask };
          }
        }
        window.$message?.success('子任务已删除');
        return true;
      }
      return false;
    } catch (error) {
      window.$message?.error('删除子任务失败');
      console.error('Failed to delete subtask:', error);
      return false;
    } finally {
      loading.value = false;
    }
  }

  async function regenerateSubtask(taskId: number, subtaskId: number, prompt?: string, loadingService?: LoadingService) {
    // 使用传入的 loadingService 或创建新的
    const ls = loadingService || createLoadingService();
    const shouldManageLoading = !loadingService;

    if (shouldManageLoading) {
      ls.start(LOADING_PRESETS.regenerateSubtask);
    }

    loading.value = true;
    try {
      // 步骤1: 分析原子任务
      if (shouldManageLoading) ls.nextStep();

      const { data: responseData, error } = await regenerateSubtaskApi(taskId, subtaskId, { prompt });

      // 步骤2: 重新生成完成
      if (shouldManageLoading) ls.nextStep();

      if (!error && responseData) {
        const updatedTask = extractData(responseData);
        // 后端返回的是整个 task 对象
        if (updatedTask) {
          const index = tasks.value.findIndex(t => t.id === taskId);
          if (index !== -1) {
            tasks.value[index] = { ...tasks.value[index], ...updatedTask };
          }
          // 更新当前任务
          if (currentTask.value?.id === taskId) {
            currentTask.value = { ...currentTask.value, ...updatedTask };
          }
        }

        if (shouldManageLoading) {
          ls.success('子任务重写成功');
        } else {
          window.$message?.success('子任务已重新生成');
        }
        return true;
      }

      if (shouldManageLoading) ls.error('子任务重写失败');
      return false;
    } catch (error) {
      if (shouldManageLoading) {
        ls.error('子任务重写失败');
      } else {
        window.$message?.error('重新生成子任务失败');
      }
      console.error('Failed to regenerate subtask:', error);
      return false;
    } finally {
      loading.value = false;
    }
  }

  // 子任务排序
  async function reorderSubtasks(taskId: number, subtaskIds: number[]) {
    try {
      const { data: responseData, error } = await reorderSubtasksApi(taskId, subtaskIds);
      if (!error && responseData) {
        const updatedTask = extractData(responseData);
        // 后端返回的是整个 task 对象
        if (updatedTask) {
          // 更新列表中的任务
          const index = tasks.value.findIndex(t => t.id === taskId);
          if (index !== -1) {
            tasks.value[index] = { ...tasks.value[index], ...updatedTask };
          }
          // 更新当前任务
          if (currentTask.value?.id === taskId) {
            currentTask.value = { ...currentTask.value, ...updatedTask };
          }
        }
        return true;
      }
      return false;
    } catch (error) {
      window.$message?.error('子任务排序失败');
      console.error('Failed to reorder subtasks:', error);
      return false;
    }
  }

  async function copyTask(taskId: number, loadingService?: LoadingService) {
    // 使用传入的 loadingService 或创建新的
    const ls = loadingService || createLoadingService();
    const shouldManageLoading = !loadingService;

    if (shouldManageLoading) {
      ls.start(LOADING_PRESETS.copyTask);
    }

    loading.value = true;
    try {
      // 步骤1: 复制任务
      if (shouldManageLoading) ls.nextStep();

      const { data, error } = await copyTaskApi(taskId);

      // 步骤2: 复制子任务
      if (shouldManageLoading) ls.nextStep();

      if (!error && data) {
        await loadTasks();

        if (shouldManageLoading) {
          ls.success('任务复制成功');
        } else {
          window.$message?.success('任务复制成功');
        }
        return extractData(data);
      }

      if (shouldManageLoading) ls.error('任务复制失败');
      return null;
    } catch (error) {
      if (shouldManageLoading) {
        ls.error('任务复制失败');
      } else {
        window.$message?.error('复制任务失败');
      }
      console.error('Failed to copy task:', error);
      return null;
    } finally {
      loading.value = false;
    }
  }

  async function createTask(formData: TaskCreateRequest) {
    loading.value = true;
    try {
      const { data, error } = await createTaskApi(formData);
      if (!error && data) {
        const newTask = extractData(data);
        // 添加到列表开头
        tasks.value.unshift(newTask);
        total.value += 1;
        window.$message?.success('任务创建成功');
        return { success: true, data: newTask };
      }
      return { success: false, error };
    } catch (err) {
      window.$message?.error('创建任务失败');
      console.error('Failed to create task:', err);
      return { success: false, error: err };
    } finally {
      loading.value = false;
    }
  }

  // 依赖关系相关方法
  async function loadTaskDependencies() {
    try {
      const { data, error } = await getTaskDependenciesApi();
      if (!error && data) {
        taskDependencies.value = extractData(data) || [];
      }
    } catch (error) {
      console.error('Failed to load task dependencies:', error);
    }
  }

  async function addTaskDependency(taskId: number, dependsOnTaskId: number) {
    try {
      const { data, error } = await addTaskDependencyApi(taskId, dependsOnTaskId);
      if (!error && data) {
        await loadTaskDependencies();
        window.$message?.success('依赖关系添加成功');
        return true;
      }
      return false;
    } catch (error) {
      window.$message?.error('添加依赖关系失败');
      console.error('Failed to add task dependency:', error);
      return false;
    }
  }

  async function removeTaskDependency(taskId: number, dependsOnTaskId: number) {
    try {
      const { data, error } = await removeTaskDependencyApi(taskId, dependsOnTaskId);
      if (!error && data) {
        await loadTaskDependencies();
        window.$message?.success('依赖关系移除成功');
        return true;
      }
      return false;
    } catch (error) {
      window.$message?.error('移除依赖关系失败');
      console.error('Failed to remove task dependency:', error);
      return false;
    }
  }

  async function validateDependencies() {
    try {
      const { data, error } = await validateDependenciesApi();
      if (!error && data) {
        const result = extractData(data);
        // result 格式: { message: string, valid: boolean }
        if (result && typeof result === 'object') {
          return { valid: result.valid, message: result.message };
        }
        return { valid: true, message: '依赖关系有效' };
      }
      return { valid: true, message: '依赖关系有效' };
    } catch (error) {
      window.$message?.error('验证依赖关系失败');
      console.error('Failed to validate dependencies:', error);
      return { valid: false, message: '验证失败' };
    }
  }

  async function loadReadyTasks() {
    try {
      const { data, error } = await getReadyTasksApi();
      if (!error && data) {
        return extractData(data) || [];
      }
      return [];
    } catch (error) {
      console.error('Failed to load ready tasks:', error);
      return [];
    }
  }

  function clearCurrentTask() {
    currentTask.value = null;
  }

  return {
    // 状态
    tasks,
    currentTask,
    loading,
    projectName,
    projectVersion,
    total,
    page,
    pageSize,
    taskDependencies,
    // 计算属性
    statistics,
    tasksByStatus,
    // Actions
    loadTasks,
    loadTaskDetail,
    setTaskStatus,
    setTaskAssignee,
    updateTask: updateTaskById,
    setSubtaskStatus,
    updateSubtask,
    expandTask,
    expandTaskAsync,
    clearTaskSubtasks,
    deleteTask,
    batchDeleteTasks,
    deleteSubtask,
    regenerateSubtask,
    reorderSubtasks,
    clearCurrentTask,
    copyTask,
    createTask,
    loadTaskDependencies,
    addTaskDependency,
    removeTaskDependency,
    validateDependencies,
    loadReadyTasks
  };
});
