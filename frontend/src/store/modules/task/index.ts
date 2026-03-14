import { defineStore } from 'pinia';
import { computed, ref } from 'vue';
import {
  batchDeleteTasks as batchDeleteTasksApi,
  clearTaskSubtasks as clearTaskSubtasksApi,
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
  copyTask as copyTaskApi
} from '@/service/api/task';
import type { Task, Subtask, TaskStatus, TaskStatistics, TaskListParams } from '@/typings/api/task';
import { createLoadingService, LoadingService, LOADING_PRESETS } from '@/utils/loading-service';

export const useTaskStore = defineStore('task-store', () => {
  // 状态
  const tasks = ref<Task[]>([]);
  const currentTask = ref<Task | null>(null);
  const loading = ref(false);
  const projectName = ref('');
  const projectVersion = ref('');

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
      const { data, error } = await fetchTaskList(params);
      if (!error && data) {
        tasks.value = data.tasks || [];
        projectName.value = data.projectName || '';
        projectVersion.value = data.projectVersion || '';
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
        currentTask.value = data;
      }
      return data;
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
      const { data: updatedTask, error } = await updateTask(id, { status });
      if (!error && updatedTask) {
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
      const { data: updatedTask, error } = await updateTask(id, { assignee });
      if (!error && updatedTask) {
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
  async function updateTaskById(id: number, data: Record<string, any>) {
    try {
      const { data: updatedTask, error } = await updateTask(id, data);
      if (!error && updatedTask) {
        // 更新列表中的任务
        const index = tasks.value.findIndex(t => t.id === id);
        if (index !== -1) {
          tasks.value[index] = { ...tasks.value[index], ...updatedTask };
        }
        // 如果是当前任务，也更新
        if (currentTask.value?.id === id) {
          currentTask.value = { ...currentTask.value, ...updatedTask };
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
      const { data: updatedSubtask, error } = await updateSubtask(taskId, subtaskId, { status });
      if (!error && updatedSubtask) {
        // 更新列表中任务的子任务
        const taskIndex = tasks.value.findIndex(t => t.id === taskId);
        if (taskIndex !== -1 && tasks.value[taskIndex].subtasks) {
          const subtaskIndex = tasks.value[taskIndex].subtasks!.findIndex(st => st.id === subtaskId);
          if (subtaskIndex !== -1) {
            tasks.value[taskIndex].subtasks![subtaskIndex] = {
              ...tasks.value[taskIndex].subtasks![subtaskIndex],
              ...updatedSubtask
            };
          }
        }
        // 如果是当前任务，也更新
        if (currentTask.value?.id === taskId && currentTask.value.subtasks) {
          const subtaskIndex = currentTask.value.subtasks.findIndex(st => st.id === subtaskId);
          if (subtaskIndex !== -1) {
            currentTask.value.subtasks[subtaskIndex] = {
              ...currentTask.value.subtasks[subtaskIndex],
              ...updatedSubtask
            };
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

  // 通用子任务更新方法
  async function updateSubtask(taskId: number, subtaskId: number, data: Record<string, any>) {
    try {
      const { data: updatedSubtask, error } = await updateSubtaskApi(taskId, subtaskId, data);
      if (!error && updatedSubtask) {
        // 更新列表中任务的子任务
        const taskIndex = tasks.value.findIndex(t => t.id === taskId);
        if (taskIndex !== -1 && tasks.value[taskIndex].subtasks) {
          const subtaskIndex = tasks.value[taskIndex].subtasks!.findIndex(st => st.id === subtaskId);
          if (subtaskIndex !== -1) {
            tasks.value[taskIndex].subtasks![subtaskIndex] = {
              ...tasks.value[taskIndex].subtasks![subtaskIndex],
              ...updatedSubtask
            };
          }
        }
        // 如果是当前任务，也更新
        if (currentTask.value?.id === taskId && currentTask.value.subtasks) {
          const subtaskIndex = currentTask.value.subtasks.findIndex(st => st.id === subtaskId);
          if (subtaskIndex !== -1) {
            currentTask.value.subtasks[subtaskIndex] = {
              ...currentTask.value.subtasks[subtaskIndex],
              ...updatedSubtask
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

      const { data: updatedTask, error } = await expandTaskApi(id);

      // 步骤2: 生成完成
      if (shouldManageLoading) ls.nextStep();

      if (!error && updatedTask) {
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
      if (!error && data) {
        return data.messageId;
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
      const { data: updatedTask, error } = await clearTaskSubtasksApi(taskId);
      if (!error && updatedTask) {
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

      const successIds = !error && data ? data.successIds || [] : [];
      const failedIds = !error && data ? data.failedIds || [] : taskIds;

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
      const { data: updatedTask, error } = await deleteSubtaskApi(taskId, subtaskId);
      if (!error && updatedTask) {
        const index = tasks.value.findIndex(t => t.id === taskId);
        if (index !== -1) {
          tasks.value[index] = { ...tasks.value[index], ...updatedTask };
        }
        currentTask.value = updatedTask;
        await loadTasks();
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

      const { data: updatedTask, error } = await regenerateSubtaskApi(taskId, subtaskId, { prompt });

      // 步骤2: 重新生成完成
      if (shouldManageLoading) ls.nextStep();

      if (!error && updatedTask) {
        const index = tasks.value.findIndex(t => t.id === taskId);
        if (index !== -1) {
          tasks.value[index] = { ...tasks.value[index], ...updatedTask };
        }
        currentTask.value = updatedTask;

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
      const { data: updatedTask, error } = await reorderSubtasksApi(taskId, subtaskIds);
      if (!error && updatedTask) {
        // 更新列表中的任务
        const index = tasks.value.findIndex(t => t.id === taskId);
        if (index !== -1) {
          tasks.value[index] = { ...tasks.value[index], ...updatedTask };
        }
        // 更新当前任务
        if (currentTask.value?.id === taskId) {
          currentTask.value = updatedTask;
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
        return data;
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
    // 计算属性
    statistics,
    tasksByStatus,
    // Actions
    loadTasks,
    loadTaskDetail,
    setTaskStatus,
    setTaskAssignee,
    updateTask,
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
    copyTask
  };
});
