import { defineStore } from 'pinia';
import { computed, ref } from 'vue';
import { fetchTaskList, fetchTaskDetail, updateTask, updateSubtask } from '@/service/api/task';
import type { Task, Subtask, TaskStatus, TaskStatistics } from '@/typings/api/task';

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
  async function loadTasks() {
    loading.value = true;
    try {
      const data = await fetchTaskList();
      if (data) {
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
      const task = await fetchTaskDetail(id, locale);
      if (task) {
        currentTask.value = task;
      }
      return task;
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
      const updatedTask = await updateTask(id, { status });
      if (updatedTask) {
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

  async function setSubtaskStatus(taskId: number, subtaskId: number, status: TaskStatus) {
    try {
      const updatedSubtask = await updateSubtask(taskId, subtaskId, { status });
      if (updatedSubtask) {
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
    setSubtaskStatus,
    clearCurrentTask
  };
});
